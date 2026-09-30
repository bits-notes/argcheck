package validation

import (
	"context"
	"fmt"
)

// EntailmentGateCatalogVersion identifies the schema catalog in force; bump it
// whenever a schema is added or changed so certificates stay comparable.
const EntailmentGateCatalogVersion = "v10"

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

// InertCitationReport identifies a conclusion whose full citation set fails
// the gate even though a proper subset instantiates a known gate schema. The
// omitted citations therefore do no work in that certified entailment.
type InertCitationReport struct {
	Step           int    `json:"step"`
	Form           string `json:"form"`
	UsedCitations  []int  `json:"used_citations"`
	InertCitations []int  `json:"inert_citations"`
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
	argument, parseResult := ParseNumberedArgument(normalizeGateMarkers(output), options)
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
		"catalog":      EntailmentGateCatalogVersion,
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

// NumberedArgumentInertCitations finds gate-failing conclusions for which a
// fixed-arity schema passes on a proper subset of the cited premises. It is a
// conservative structural check: it reports only a witnessed gate proof and
// never guesses whether semantically difficult residue contains an extra
// citation.
func NumberedArgumentInertCitations(arg NumberedArgument) []InertCitationReport {
	stepsByNumber := make(map[int]NumberedArgumentStep, len(arg.Steps))
	for _, step := range arg.Steps {
		stepsByNumber[step.Number] = step
	}

	reports := make([]InertCitationReport, 0)
	for _, conclusion := range arg.Steps {
		if conclusion.Kind != numberedArgumentStepConclusion || len(conclusion.Citations) < 2 {
			continue
		}
		if gateConclusionStep(conclusion, stepsByNumber).Passed {
			continue
		}

		premises := make([]NumberedArgumentStep, 0, len(conclusion.Citations))
		for _, citation := range conclusion.Citations {
			premises = append(premises, stepsByNumber[citation])
		}
		for _, schema := range entailmentSchemas {
			if schema.arity < 0 || schema.arity >= len(premises) {
				continue
			}
			for _, selection := range orderedPremiseSelections(premises, schema.arity) {
				if !schema.match(selection, conclusion).matched {
					continue
				}
				used := make(map[int]bool, len(selection))
				usedCitations := make([]int, 0, len(selection))
				for _, premise := range selection {
					used[premise.Number] = true
					usedCitations = append(usedCitations, premise.Number)
				}
				inert := make([]int, 0, len(premises)-len(selection))
				for _, citation := range conclusion.Citations {
					if !used[citation] {
						inert = append(inert, citation)
					}
				}
				reports = append(reports, InertCitationReport{
					Step:           conclusion.Number,
					Form:           schema.name,
					UsedCitations:  usedCitations,
					InertCitations: inert,
				})
				break
			}
			if len(reports) > 0 && reports[len(reports)-1].Step == conclusion.Number {
				break
			}
		}
	}
	return reports
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

	for _, set := range gateCandidateSets(premises, conclusion) {
		for _, schema := range entailmentSchemas {
			if !schemaAcceptsPremiseCount(schema, len(set.premises)) {
				continue
			}
			for _, ordered := range orderedPremiseSelections(set.premises, schema.arity) {
				outcome := schema.match(ordered, set.conclusion)
				if outcome.matched {
					report.Passed = true
					report.Form = schema.name + set.formSuffix
					return report
				}
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
	for _, set := range gateCandidateSets(premises, conclusion) {
		for _, schema := range entailmentSchemas {
			selections := orderedPremiseSelections(set.premises, schema.arity)
			for _, ordered := range selections {
				outcome := schema.match(ordered, set.conclusion)
				if !schemaAcceptsPremiseCount(schema, len(set.premises)) {
					outcome.traction = false
				}
				if outcome.score > best.score || (outcome.score == best.score && outcome.traction && !best.traction) {
					best = outcome
					bestName = schema.name + set.formSuffix
				}
			}
			if schema.name == "universal instantiation" {
				for _, premise := range set.premises {
					outcome := scoreUniversalPredicateSwap(premise, set.conclusion)
					if outcome.score > best.score || (outcome.score == best.score && outcome.traction && !best.traction) {
						best = outcome
						bestName = schema.name + set.formSuffix
					}
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

type gateCandidateSet struct {
	premises   []NumberedArgumentStep
	conclusion NumberedArgumentStep
	formSuffix string
}

// gateCandidateSets returns the step as written plus, when the conclusion and
// every cited premise carry a leading "Necessarily," operator, the
// de-modalized step. Certifying the stripped step under any catalog schema is
// sound in modal logic K: from boxed premises whose unboxed forms entail C,
// infer boxed C. Steps where only some premises are modalized get no stripped
// set — necessity in the conclusion must not outrun a contingent premise.
func gateCandidateSets(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) []gateCandidateSet {
	sets := []gateCandidateSet{{premises: premises, conclusion: conclusion}}
	if len(premises) == 0 {
		return sets
	}
	strippedConclusions := necessityCandidates(conclusion.Statement)
	if len(strippedConclusions) == 0 {
		return sets
	}
	strippedPremises := make([]NumberedArgumentStep, len(premises))
	for index, premise := range premises {
		stripped := necessityCandidates(premise.Statement)
		if len(stripped) == 0 {
			return sets
		}
		strippedPremises[index] = premise
		strippedPremises[index].Statement = stripped[0]
	}
	strippedConclusion := conclusion
	strippedConclusion.Statement = strippedConclusions[0]
	return append(sets, gateCandidateSet{
		premises:   strippedPremises,
		conclusion: strippedConclusion,
		formSuffix: " (necessitation)",
	})
}

func schemaAcceptsPremiseCount(schema entailmentSchema, premiseCount int) bool {
	if schema.arity == variableArityAtLeastTwo {
		return premiseCount >= 2
	}
	if schema.arity == variableArity {
		return premiseCount >= 3
	}
	return schema.arity == premiseCount
}

func orderedPremiseSelections(premises []NumberedArgumentStep, arity int) [][]NumberedArgumentStep {
	if arity == variableArityAtLeastTwo {
		if len(premises) < 2 {
			return nil
		}
		return [][]NumberedArgumentStep{premises}
	}
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
