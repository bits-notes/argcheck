package validation

import (
	"context"
	"fmt"
)

// EntailmentGateOptions are the parsing options shared with the
// numbered_argument_syntax builtin.
type EntailmentGateOptions struct {
	AllowSingleFencedBlock bool `json:"allow_single_fenced_block"`
	AllowBlankLines        bool `json:"allow_blank_lines"`
}

// EntailmentGateReport is the deterministic result for every conclusion in a
// numbered argument.
type EntailmentGateReport struct {
	Passed bool                       `json:"passed"`
	Steps  []EntailmentGateStepReport `json:"steps"`
}

// EntailmentGateStepReport records the matched schema or nearest failed form
// for one conclusion step.
type EntailmentGateStepReport struct {
	Step        int    `json:"step"`
	Citations   []int  `json:"citations"`
	Passed      bool   `json:"passed"`
	Form        string `json:"form,omitempty"`
	NearestForm string `json:"nearest_form,omitempty"`
	Mismatch    string `json:"mismatch,omitempty"`
}

type entailmentGateValidator struct {
	options numberedArgumentOptions
}

func (v entailmentGateValidator) Validate(ctx context.Context, output string) Result {
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Passed: false, Message: "validation timed out"}
	default:
	}

	_, result := checkEntailmentGate(output, v.options)
	return result
}

// CheckEntailmentGate parses, expands, and gates a numbered argument. At most
// one options value may be supplied.
func CheckEntailmentGate(output string, options ...EntailmentGateOptions) (EntailmentGateReport, Result) {
	if len(options) > 1 {
		return EntailmentGateReport{}, Result{Passed: false, Message: "entailment_gate accepts at most one options value"}
	}
	parseOptions := numberedArgumentOptions{}
	if len(options) == 1 {
		parseOptions.AllowSingleFencedBlock = options[0].AllowSingleFencedBlock
		parseOptions.AllowBlankLines = options[0].AllowBlankLines
	}
	return checkEntailmentGate(output, parseOptions)
}

func checkEntailmentGate(output string, options numberedArgumentOptions) (EntailmentGateReport, Result) {
	argument, parseResult := ParseNumberedArgument(output, options)
	if !parseResult.Passed {
		return EntailmentGateReport{Passed: false}, parseResult
	}

	expanded, expansions, err := ExpandNumberedArgumentBridges(argument)
	if err != nil {
		return EntailmentGateReport{Passed: false}, Result{
			Passed:  false,
			Message: fmt.Sprintf("entailment gate bridge expansion failed: %v", err),
		}
	}
	report := GateNumberedArgument(expanded)
	passedSteps := 0
	for _, step := range report.Steps {
		if step.Passed {
			passedSteps++
		}
	}
	details := map[string]any{
		"steps":        report.Steps,
		"gated_steps":  len(report.Steps),
		"passed_steps": passedSteps,
		"expansions":   expansions,
	}
	if report.Passed {
		return report, Result{
			Passed:  true,
			Message: "output passes entailment_gate",
			Details: details,
		}
	}

	message := "output fails entailment_gate"
	for _, step := range report.Steps {
		if step.Passed {
			continue
		}
		message = fmt.Sprintf("step %d failed entailment_gate: nearest form %s; %s", step.Step, step.NearestForm, step.Mismatch)
		break
	}
	return report, Result{Passed: false, Message: message, Details: details}
}

// GateNumberedArgument checks every conclusion step of an already-expanded
// numbered argument.
func GateNumberedArgument(arg NumberedArgument) EntailmentGateReport {
	stepsByNumber := make(map[int]NumberedArgumentStep, len(arg.Steps))
	for _, step := range arg.Steps {
		stepsByNumber[step.Number] = step
	}

	report := EntailmentGateReport{Passed: true, Steps: make([]EntailmentGateStepReport, 0)}
	for _, step := range arg.Steps {
		if step.Kind != numberedArgumentStepConclusion {
			continue
		}
		stepReport := gateConclusionStep(step, stepsByNumber)
		report.Steps = append(report.Steps, stepReport)
		if !stepReport.Passed {
			report.Passed = false
		}
	}
	return report
}

func gateConclusionStep(conclusion NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	report := EntailmentGateStepReport{
		Step:      conclusion.Number,
		Citations: append([]int(nil), conclusion.Citations...),
	}
	premises := make([]NumberedArgumentStep, 0, len(conclusion.Citations))
	for _, citation := range conclusion.Citations {
		premise, ok := steps[citation]
		if !ok {
			report.NearestForm = "none"
			report.Mismatch = fmt.Sprintf("cited step %d is unavailable for conclusion step %d", citation, conclusion.Number)
			return report
		}
		premises = append(premises, premise)
	}

	for _, schema := range entailmentSchemas {
		if !schemaAcceptsPremiseCount(schema, len(premises)) {
			continue
		}
		for _, ordered := range orderedPremiseSelections(premises, schema.arity) {
			outcome := schema.match(ordered, conclusion)
			if outcome.matched {
				report.Passed = true
				report.Form = schema.name
				return report
			}
		}
	}

	if len(premises) < 3 {
		if special, ok := conjunctiveAntecedentModusPonensFailure(premises, conclusion); ok {
			report.NearestForm = "modus ponens"
			report.Mismatch = special
			return report
		}
	}

	best := schemaMatchOutcome{}
	bestName := "none"
	for _, schema := range entailmentSchemas {
		selections := orderedPremiseSelections(premises, schema.arity)
		for _, ordered := range selections {
			outcome := schema.match(ordered, conclusion)
			if !schemaAcceptsPremiseCount(schema, len(premises)) {
				outcome.traction = false
			}
			if outcome.score > best.score || (outcome.score == best.score && outcome.traction && !best.traction) {
				best = outcome
				bestName = schema.name
			}
		}
		if schema.name == "universal instantiation" {
			for _, premise := range premises {
				outcome := scoreUniversalPredicateSwap(premise, conclusion)
				if outcome.score > best.score || (outcome.score == best.score && outcome.traction && !best.traction) {
					best = outcome
					bestName = schema.name
				}
			}
		}
	}
	if best.score > 0 || best.traction {
		report.NearestForm = bestName
		report.Mismatch = best.mismatch
		if report.Mismatch == "" {
			report.Mismatch = fmt.Sprintf("conclusion step %d does not instantiate %s", conclusion.Number, bestName)
		}
		return report
	}
	report.NearestForm = "none"
	report.Mismatch = fmt.Sprintf("no cited premise matches conclusion step %d under any entailment-gate schema", conclusion.Number)
	return report
}

func schemaAcceptsPremiseCount(schema entailmentSchema, premiseCount int) bool {
	if schema.arity == variableArity {
		return premiseCount >= 3
	}
	return schema.arity == premiseCount
}

func orderedPremiseSelections(premises []NumberedArgumentStep, arity int) [][]NumberedArgumentStep {
	if arity == variableArity {
		if len(premises) < 3 {
			return nil
		}
		return [][]NumberedArgumentStep{premises}
	}
	if arity == 1 {
		result := make([][]NumberedArgumentStep, 0, len(premises))
		for _, premise := range premises {
			result = append(result, []NumberedArgumentStep{premise})
		}
		return result
	}
	if arity == 2 {
		result := make([][]NumberedArgumentStep, 0, len(premises)*(len(premises)-1))
		for first := range premises {
			for second := range premises {
				if first == second {
					continue
				}
				result = append(result, []NumberedArgumentStep{premises[first], premises[second]})
			}
		}
		return result
	}
	return nil
}
