package validation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var inlineStepBridgeExpansionPattern = regexp.MustCompile(`^If ((?:\[[1-9][0-9]*(?:,[1-9][0-9]*)*\](?: and )?)+)(?:,? then,? |, )(.+)$`)

// BridgeExpansion records one bracket-bridge statement replaced by
// ExpandNumberedArgumentBridges.
type BridgeExpansion struct {
	Step     int    `json:"step"`
	Original string `json:"original"`
	Expanded string `json:"expanded"`
}

// ExpandNumberedArgumentBridges replaces bracket-bridge statements with
// canonical text assembled verbatim from the referenced earlier steps.
func ExpandNumberedArgumentBridges(arg NumberedArgument) (NumberedArgument, []BridgeExpansion, error) {
	expanded := NumberedArgument{Steps: append([]NumberedArgumentStep(nil), arg.Steps...)}
	order := make([]int, len(expanded.Steps))
	for index := range expanded.Steps {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool {
		return expanded.Steps[order[i]].Number < expanded.Steps[order[j]].Number
	})

	statements := make(map[int]string, len(expanded.Steps))
	expansions := make([]BridgeExpansion, 0)
	for _, index := range order {
		step := &expanded.Steps[index]
		if _, exists := statements[step.Number]; exists {
			return NumberedArgument{}, nil, fmt.Errorf("duplicate numbered-argument step %d", step.Number)
		}

		matches := inlineStepBridgeExpansionPattern.FindStringSubmatch(step.Statement)
		if matches == nil {
			statements[step.Number] = step.Statement
			continue
		}

		references := bracketReferencePattern.FindAllStringSubmatch(matches[1], -1)
		parts := make([]string, 0, len(references))
		for _, referenceMatch := range references {
			if len(referenceMatch) < 2 {
				continue
			}
			stepNumbers, err := parseCitations(referenceMatch[1])
			if err != nil {
				return NumberedArgument{}, nil, fmt.Errorf("step %d has invalid bridge reference %q: %w", step.Number, referenceMatch[0], err)
			}
			for _, referencedStep := range stepNumbers {
				statement, ok := statements[referencedStep]
				if !ok || referencedStep >= step.Number {
					return NumberedArgument{}, nil, fmt.Errorf("step %d bridge references unavailable prior step %d", step.Number, referencedStep)
				}
				parts = append(parts, "it is the case that "+statement)
			}
		}
		if len(parts) == 0 {
			statements[step.Number] = step.Statement
			continue
		}

		original := step.Statement
		antecedent := parts[0]
		if len(parts) > 1 {
			antecedent = strings.Join(parts, ", and ")
		}
		step.Statement = "If " + antecedent + ", then " + matches[2]
		step.Line = numberedArgumentStepLine(*step)
		statements[step.Number] = step.Statement
		expansions = append(expansions, BridgeExpansion{
			Step:     step.Number,
			Original: original,
			Expanded: step.Statement,
		})
	}

	return expanded, expansions, nil
}

func numberedArgumentStepLine(step NumberedArgumentStep) string {
	if step.Kind == numberedArgumentStepConclusion {
		citations := make([]string, len(step.Citations))
		for index, citation := range step.Citations {
			citations[index] = fmt.Sprintf("%d", citation)
		}
		return fmt.Sprintf("%d. Therefore, %s. [%s]", step.Number, step.Statement, strings.Join(citations, ","))
	}
	return fmt.Sprintf("%d. %s.", step.Number, step.Statement)
}
