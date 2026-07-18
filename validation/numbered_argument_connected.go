package validation

import "sort"

// NumberedArgumentUnusedSteps returns the step numbers that do not support the
// argument's final conclusion — i.e. steps not reachable by walking citation
// edges backward from the last step. A well-formed numbered argument has none:
// every premise must feed a conclusion, and every intermediate conclusion must
// itself be cited, so the whole graph is connected to the final conclusion.
// This is a purely structural, deterministic check (no LLM). It returns nil when
// the argument has no conclusions (nothing to be connected to).
func NumberedArgumentUnusedSteps(arg NumberedArgument) []int {
	if len(arg.Steps) == 0 {
		return nil
	}
	final := arg.Steps[len(arg.Steps)-1]
	if final.Kind != numberedArgumentStepConclusion {
		return nil
	}

	citations := make(map[int][]int, len(arg.Steps))
	present := make(map[int]bool, len(arg.Steps))
	for _, step := range arg.Steps {
		citations[step.Number] = step.Citations
		present[step.Number] = true
	}

	reachable := make(map[int]bool, len(arg.Steps))
	stack := []int{final.Number}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable[n] {
			continue
		}
		reachable[n] = true
		for _, cited := range citations[n] {
			if present[cited] && !reachable[cited] {
				stack = append(stack, cited)
			}
		}
	}

	unused := make([]int, 0)
	for _, step := range arg.Steps {
		if !reachable[step.Number] {
			unused = append(unused, step.Number)
		}
	}
	sort.Ints(unused)
	return unused
}
