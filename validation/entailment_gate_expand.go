package validation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var inlineStepBridgeExpansionPattern = regexp.MustCompile(`^If ((?:\[[1-9][0-9]*(?:,[1-9][0-9]*)*\](?:,? and )?)+)(?:,? then,? |, )(.+)$`)

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
		original := step.Statement

		// v31.1: references filling the whole consequent expand the same way.
		if c := inlineStepConsequentPattern.FindStringSubmatch(step.Statement); c != nil && strings.Contains(c[2], "[") {
			text, err := expandStepReferenceList(c[2], step.Number, statements)
			if err != nil {
				return NumberedArgument{}, nil, err
			}
			if text != "" {
				step.Statement = "If " + c[1] + ", then " + text
			}
		}

		if matches := inlineStepBridgeExpansionPattern.FindStringSubmatch(step.Statement); matches != nil {
			text, err := expandStepReferenceList(matches[1], step.Number, statements)
			if err != nil {
				return NumberedArgument{}, nil, err
			}
			if text != "" {
				step.Statement = "If " + text + ", then " + matches[2]
			}
		}

		if step.Statement != original {
			step.Line = numberedArgumentStepLine(*step)
			expansions = append(expansions, BridgeExpansion{Step: step.Number, Original: original, Expanded: step.Statement})
		}
		statements[step.Number] = step.Statement
	}

	return expanded, expansions, nil
}

func numberedArgumentStepLine(step NumberedArgumentStep) string {
	prefix := strings.Repeat("| ", step.Depth)
	if step.Kind == numberedArgumentStepConclusion {
		citations := make([]string, len(step.Citations))
		for index, citation := range step.Citations {
			citations[index] = fmt.Sprintf("%d", citation)
		}
		for _, r := range step.Ranges {
			citations = append(citations, fmt.Sprintf("%d-%d", r[0], r[1]))
		}
		return prefix + fmt.Sprintf("%d. Therefore, %s. [%s]", step.Number, step.Statement, strings.Join(citations, ","))
	}
	statement := step.Statement
	if step.Kind == numberedArgumentStepAssumption {
		statement = "Suppose " + statement
	}
	return prefix + fmt.Sprintf("%d. %s.", step.Number, statement)
}

// expandStepReferenceList turns "[1] and [2]" or "[1,2]" into
// "it is the case that S1, and it is the case that S2", copying each earlier
// step verbatim.
func expandStepReferenceList(raw string, stepNumber int, statements map[int]string) (string, error) {
	parts := make([]string, 0)
	for _, referenceMatch := range bracketReferencePattern.FindAllStringSubmatch(raw, -1) {
		if len(referenceMatch) < 2 {
			continue
		}
		stepNumbers, err := parseCitations(referenceMatch[1])
		if err != nil {
			return "", fmt.Errorf("step %d has invalid bridge reference %q: %w", stepNumber, referenceMatch[0], err)
		}
		for _, referencedStep := range stepNumbers {
			statement, ok := statements[referencedStep]
			if !ok || referencedStep >= stepNumber {
				return "", fmt.Errorf("step %d bridge references unavailable prior step %d", stepNumber, referencedStep)
			}
			parts = append(parts, "it is the case that "+statement)
		}
	}
	return strings.Join(parts, ", and "), nil
}
