package validation

import "sort"

// NumberedArgumentUnusedSteps returns unsupported lines. Flat proofs preserve
// the existing contract: all premises must be cited, while conclusions may
// stand independently. In Fitch proofs, range citations use every covered line
// (including their assumptions and endpoints); other scoped lines must be used
// by a citation. This is a structural check, not an entailment check.
func NumberedArgumentUnusedSteps(arg NumberedArgument) []int {
	cited := make(map[int]bool, len(arg.Steps))
	for _, step := range arg.Steps {
		for _, r := range step.Ranges {
			for n := r[0]; n <= r[1] && n <= len(arg.Steps); n++ {
				cited[n] = true
			}
		}
		for _, c := range step.Citations {
			cited[c] = true
		}
	}
	unused := make([]int, 0)
	for _, step := range arg.Steps {
		if step.Kind == numberedArgumentStepConclusion && step.Depth == 0 {
			continue
		}
		if !cited[step.Number] {
			unused = append(unused, step.Number)
		}
	}
	sort.Ints(unused)
	return unused
}
