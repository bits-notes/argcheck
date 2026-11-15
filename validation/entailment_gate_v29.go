package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// v29: "P unless Q" on its weak reading, "if it is not the case that Q, then P"
// (equivalently "P or Q"). Never the biconditional or exclusive reading, so
// "P unless Q" with Q does not yield "not P", and with P does not yield "not Q".
// The clause itself stays verbatim everywhere else: these are two new forms,
// not a normalization, so no earlier certificate changes.
// Appended after the v28 forms, like every catalog form since v20.
var v29UnlessForms = []string{"unless elimination", "unless conversion"}

func v29UnlessForm(name string) bool {
	return name == v29UnlessForms[0] || name == v29UnlessForms[1]
}

func init() {
	entailmentSchemas = append(entailmentSchemas,
		entailmentSchema{name: "unless elimination", arity: 2, match: matchUnlessElimination},
		entailmentSchema{name: "unless conversion", arity: 1, match: matchUnlessConversion},
	)
}

// unlessAnaphor marks an unless-clause whose pronoun may point back into a
// quantified main clause ("Every student passes unless he fails"); the
// propositional reading would then prove more than the sentence says.
var unlessAnaphor = regexp.MustCompile(`\b(?:he|she|it|they|him|her|them|his|hers|its|their|theirs|himself|herself|itself|themselves)\b`)

// unlessCandidates splits "P unless Q" and "Unless Q, P" at each "unless".
func unlessCandidates(statement string) []clausePair {
	normalized := normalizeClause(statement)
	result := appendSplitCandidates(nil, normalized, " unless ", false)
	if rest, ok := strings.CutPrefix(normalized, "unless "); ok {
		if q, p, found := strings.Cut(rest, ", "); found {
			result = append(result, clausePair{left: strings.TrimSpace(p), right: strings.TrimSpace(q)})
		}
	}
	kept := result[:0]
	for _, pair := range result {
		if pair.left != "" && pair.right != "" && !unlessAnaphor.MatchString(pair.right) {
			kept = append(kept, pair)
		}
	}
	return kept
}

// unless elimination: "P unless Q", "it is not the case that Q" ⊢ "P".
func matchUnlessElimination(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	negated := negationCandidates(premises[1].Statement)
	for _, pair := range unlessCandidates(premises[0].Statement) {
		if pair.left == "" || pair.right == "" {
			continue
		}
		for _, q := range negated {
			candidate := evaluateIdentities(
				clauseIdentity{fmt.Sprintf("premise %d unless-clause", premises[0].Number), pair.right, fmt.Sprintf("premise %d negated clause", premises[1].Number), q},
				clauseIdentity{fmt.Sprintf("premise %d main clause", premises[0].Number), pair.left, fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement},
			)
			best = betterOutcome(best, candidate)
			if best.matched {
				return best
			}
		}
	}
	return best
}

// unless conversion: "P unless Q" ⊣⊢ "if it is not the case that Q, then P".
func matchUnlessConversion(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, direction := range [][2]NumberedArgumentStep{{premises[0], conclusion}, {conclusion, premises[0]}} {
		unlessStep, conditionalStep := direction[0], direction[1]
		for _, pair := range unlessCandidates(unlessStep.Statement) {
			if pair.left == "" || pair.right == "" {
				continue
			}
			for _, conditional := range conditionalCandidates(conditionalStep.Statement) {
				for _, q := range negationCandidates(conditional.left) {
					candidate := evaluateIdentities(
						clauseIdentity{fmt.Sprintf("step %d unless-clause", unlessStep.Number), pair.right, fmt.Sprintf("step %d negated antecedent", conditionalStep.Number), q},
						clauseIdentity{fmt.Sprintf("step %d main clause", unlessStep.Number), pair.left, fmt.Sprintf("step %d consequent", conditionalStep.Number), conditional.right},
					)
					best = betterOutcome(best, candidate)
					if best.matched {
						return best
					}
				}
			}
		}
	}
	return best
}
