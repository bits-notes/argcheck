package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// A parenthesized step number such as "(1)" or "(1, 2)" is ordinary text to
// the gate. When a failing step involves one that names earlier steps, the
// writer almost certainly meant a bracketed step reference, so the failure
// message says how to write it. This changes messages only, never verdicts.
var parenStepReferencePattern = regexp.MustCompile(`\((\s*[1-9][0-9]*(?:\s*,\s*[1-9][0-9]*)*\s*)\)`)

func stepReferenceHint(statement string, stepNumber int) string {
	for _, match := range parenStepReferencePattern.FindAllStringSubmatch(statement, -1) {
		parts := strings.Split(match[1], ",")
		numbers := make([]string, 0, len(parts))
		earlier := true
		for _, part := range parts {
			var n int
			if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &n); err != nil || n >= stepNumber {
				earlier = false
				break
			}
			numbers = append(numbers, fmt.Sprint(n))
		}
		if !earlier || len(numbers) == 0 {
			continue
		}
		bracket := "[" + strings.Join(numbers, ",") + "]"
		if len(numbers) == 1 {
			return fmt.Sprintf("step %d writes %q; to refer to step %s, write %s (a bridge premise such as \"If %s, then ...\")", stepNumber, strings.TrimSpace(match[0]), numbers[0], bracket, bracket)
		}
		return fmt.Sprintf("step %d writes %q; to refer to steps %s, write %s (a bridge premise such as \"If %s, then ...\")", stepNumber, strings.TrimSpace(match[0]), strings.Join(numbers, " and "), bracket, bracket)
	}
	return ""
}

// addStepReferenceHints appends a hint to each failing step whose own text or
// cited steps contain a parenthesized earlier-step number.
func addStepReferenceHints(report *EntailmentGateReport, argument NumberedArgument) {
	statements := make(map[int]string, len(argument.Steps))
	for _, step := range argument.Steps {
		statements[step.Number] = step.Statement
	}
	for i := range report.Steps {
		step := &report.Steps[i]
		if step.Passed {
			continue
		}
		candidates := append([]int{step.Step}, step.Citations...)
		for _, n := range candidates {
			statement, ok := statements[n]
			if !ok {
				continue
			}
			if hint := stepReferenceHint(statement, n); hint != "" {
				if step.Mismatch == "" {
					step.Mismatch = hint
				} else {
					step.Mismatch += "; hint: " + hint
				}
				break
			}
		}
	}
}
