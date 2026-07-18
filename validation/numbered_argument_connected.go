package validation

import "sort"

// NumberedArgumentUnusedSteps returns the premise steps that no step cites — i.e.
// premises that participate in no entailment. Every premise must entail a
// conclusion (be cited by at least one step); conclusions themselves are always
// permitted, whether or not they are cited, since an argument may establish more
// than one conclusion. A premise cited by nothing is dead weight the argument
// asserts but never uses. This is a purely structural, deterministic check (no
// LLM).
func NumberedArgumentUnusedSteps(arg NumberedArgument) []int {
	cited := make(map[int]bool, len(arg.Steps))
	for _, step := range arg.Steps {
		for _, c := range step.Citations {
			cited[c] = true
		}
	}
	unused := make([]int, 0)
	for _, step := range arg.Steps {
		if step.Kind == numberedArgumentStepConclusion {
			continue
		}
		if !cited[step.Number] {
			unused = append(unused, step.Number)
		}
	}
	sort.Ints(unused)
	return unused
}
