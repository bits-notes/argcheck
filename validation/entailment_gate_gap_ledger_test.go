package validation

import "testing"

// This ledger pins the boundary between supported one-step forms and valid
// inferences that the gate cannot yet certify. An empty wantForm asserts a
// known rejection; when support lands, deliberately move that row to the
// accepted section and record the actual reported form. These are surface-form
// examples, not claims that every English rendering of a rule is unsupported.
// Nested conjunctions use the gate's "both that ..., and that ..." scaffold
// to delimit scope. "Either" introduces the outer disjunction; the nested
// conditional in exportation/importation is in the consequent (P -> (Q -> R)).
// Necessity implies possibility assumes the reflexive modal setting already
// used by the gate's necessity elimination, not unrestricted modal K.
//
// Exception: relation transitivity without a premise is NOT logically valid
// for an arbitrary relation. Its rejection is arguably correct (and for
// "admires", correct): a transitivity premise is required. It is retained as
// the requested boundary case, not counted as a valid missing inference.
//
// Subproof-requiring rules: conditional proof, reductio, existential
// elimination, and universal generalization are not representable in the
// current format; they deliberately have no rows here.
func TestEntailmentGateGapLedger(t *testing.T) {
	tests := []struct {
		name, argument, wantForm string
	}{
		// Known gaps (including the explicitly qualified relation boundary).
		{"contraposition", `1. if it rains, then it snows.
2. Therefore, if it is not the case that it snows, then it is not the case that it rains. [1]`, ""},
		{"De Morgan and forward", `1. it is not the case that it is the case both that it rains, and that it snows.
2. Therefore, either it is not the case that it rains or it is not the case that it snows. [1]`, ""},
		{"De Morgan and reverse", `1. either it is not the case that it rains or it is not the case that it snows.
2. Therefore, it is not the case that it is the case both that it rains, and that it snows. [1]`, ""},
		{"De Morgan or forward", `1. it is not the case that either it rains or it snows.
2. Therefore, it is the case both that it is not the case that it rains, and that it is not the case that it snows. [1]`, ""},
		{"De Morgan or reverse", `1. it is the case both that it is not the case that it rains, and that it is not the case that it snows.
2. Therefore, it is not the case that either it rains or it snows. [1]`, ""},
		{"and commutation", `1. it rains and it snows.
2. Therefore, it snows and it rains. [1]`, ""},
		{"or commutation", `1. either it rains or it snows.
2. Therefore, either it snows or it rains. [1]`, ""},
		{"exportation", `1. if it rains and it snows, then the street is wet.
2. Therefore, if it rains, then if it snows, then the street is wet. [1]`, ""},
		{"importation", `1. if it rains, then if it snows, then the street is wet.
2. Therefore, if it rains and it snows, then the street is wet. [1]`, ""},
		{"material implication forward", `1. if it rains, then it snows.
2. Therefore, either it is not the case that it rains or it snows. [1]`, ""},
		{"material implication reverse", `1. either it is not the case that it rains or it snows.
2. Therefore, if it rains, then it snows. [1]`, ""},
		{"and distribution", `1. it is the case both that it rains, and that either it snows or the street is wet.
2. Therefore, either it rains and it snows or it rains and the street is wet. [1]`, ""},
		{"or distribution", `1. either it rains or it snows and the street is wet.
2. Therefore, it is the case both that either it rains or it snows, and that either it rains or the street is wet. [1]`, ""},
		{"absorption", `1. either it rains or it rains and it snows.
2. Therefore, it rains. [1]`, ""},
		{"conditional absorption", `1. if it rains, then it snows.
2. Therefore, if it rains, then it rains and it snows. [1]`, ""},
		{"double negation introduction", `1. it rains.
2. Therefore, it is not the case that it is not the case that it rains. [1]`, ""},
		{"biconditional modus tollens", `1. it rains if and only if it snows.
2. it is not the case that it snows.
3. Therefore, it is not the case that it rains. [1,2]`, ""},
		{"biconditional chaining", `1. it rains if and only if it snows.
2. it snows if and only if the street is wet.
3. Therefore, it rains if and only if the street is wet. [1,2]`, ""},
		{"biconditional commutation", `1. it rains if and only if it snows.
2. Therefore, it snows if and only if it rains. [1]`, ""},
		{"negated conditional antecedent", `1. it is not the case that if it rains, then it snows.
2. Therefore, it rains. [1]`, ""},
		{"negated conditional negated consequent", `1. it is not the case that if it rains, then it snows.
2. Therefore, it is not the case that it snows. [1]`, ""},
		{"identity symmetry", `1. Ada is identical to Ben.
2. Therefore, Ben is identical to Ada. [1]`, ""},
		{"quantifier negation universal", `1. it is not the case that every man is mortal.
2. Therefore, some man is not mortal. [1]`, ""},
		{"quantifier negation existential", `1. it is not the case that some man is immortal.
2. Therefore, no man is immortal. [1]`, ""},
		{"existential conversion", `1. some poet is a sailor.
2. Therefore, some sailor is a poet. [1]`, ""},
		{"universal negative conversion", `1. no poet is a sailor.
2. Therefore, no sailor is a poet. [1]`, ""},
		{"universal modus tollens", `1. every man is mortal.
2. it is not the case that Zeus is mortal.
3. Therefore, it is not the case that Zeus is a man. [1,2]`, ""},
		{"necessity implies possibility", `1. necessarily, it rains.
2. Therefore, possibly, it rains. [1]`, ""},
		{"explosion", `1. it rains.
2. it is not the case that it rains.
3. Therefore, it snows. [1,2]`, ""},
		{"excluded middle from anything", `1. the street is wet.
2. Therefore, either it rains or it is not the case that it rains. [1]`, ""},
		{"relation transitivity without premise", `1. Ada admires Ben.
2. Ben admires Cy.
3. Therefore, Ada admires Cy. [1,2]`, ""},
		// Currently accepted forms, across the major catalog families.
		{"modus ponens", `1. if it rains, then it snows.
2. it rains.
3. Therefore, it snows. [1,2]`, "modus ponens"},
		{"modus tollens", `1. if it rains, then it snows.
2. it is not the case that it snows.
3. Therefore, it is not the case that it rains. [1,2]`, "modus tollens"},
		{"hypothetical syllogism", `1. if it rains, then it snows.
2. if it snows, then the street is wet.
3. Therefore, if it rains, then the street is wet. [1,2]`, "hypothetical syllogism"},
		{"conjunctive modus ponens", `1. if it rains and it snows, then the street is wet.
2. it rains.
3. it snows.
4. Therefore, the street is wet. [1,2,3]`, "conjunctive modus ponens"},
		{"biconditional introduction", `1. if it rains, then it snows.
2. if it snows, then it rains.
3. Therefore, it rains if and only if it snows. [1,2]`, "biconditional introduction"},
		{"biconditional elimination", `1. it rains if and only if it snows.
2. Therefore, if it rains, then it snows. [1]`, "biconditional elimination"},
		{"biconditional modus ponens", `1. it rains if and only if it snows.
2. it rains.
3. Therefore, it snows. [1,2]`, "biconditional modus ponens"},
		{"conjunction introduction", `1. it rains.
2. it snows.
3. Therefore, it rains and it snows. [1,2]`, "conjunction introduction"},
		{"conjunction elimination", `1. it rains and it snows.
2. Therefore, it rains. [1]`, "conjunction elimination"},
		{"disjunction introduction", `1. it rains.
2. Therefore, either it rains or it snows. [1]`, "disjunction introduction"},
		{"disjunctive syllogism", `1. either it rains or it snows.
2. it is not the case that it rains.
3. Therefore, it snows. [1,2]`, "disjunctive syllogism"},
		{"double negation elimination", `1. it is not the case that it is not the case that it rains.
2. Therefore, it rains. [1]`, "double negation"},
		{"self entailment", `1. it rains.
2. Therefore, it rains. [1]`, "self-entailment"},
		{"identity transitivity", `1. Ada is identical to Ben.
2. Ben is identical to Cy.
3. Therefore, Ada is identical to Cy. [1,2]`, "identity transitivity"},
		{"universal instantiation", `1. every man is mortal.
2. Zeus is a man.
3. Therefore, Zeus is mortal. [1,2]`, "universal instantiation"},
		{"universal syllogism", `1. every man is mortal.
2. every mortal is vulnerable.
3. Therefore, every man is vulnerable. [1,2]`, "universal syllogism"},
		{"existential generalization", `1. Zeus is mortal.
2. Therefore, something is mortal. [1]`, "existential generalization"},
		{"necessity elimination", `1. necessarily, it rains.
2. Therefore, it rains. [1]`, "necessity elimination"},
		{"schematic instantiation", `1. if X sings, then X dances.
2. Ada sings.
3. Therefore, Ada dances. [1,2]`, "schematic instantiation"},
		{"simple dilemma", `1. either it rains or it snows.
2. if it rains, then the street is wet.
3. if it snows, then the street is wet.
4. Therefore, the street is wet. [1,2,3]`, "simple dilemma"},
		{"constructive dilemma", `1. either it rains or it snows.
2. if it rains, then the street is wet.
3. if it snows, then it rains.
4. Therefore, either the street is wet or it rains. [1,2,3]`, "constructive dilemma"},
		{"destructive dilemma", `1. if it rains, then it snows.
2. if it snows, then the street is wet.
3. either it is not the case that it snows or it is not the case that the street is wet.
4. Therefore, either it is not the case that it rains or it is not the case that it snows. [1,2,3]`, "destructive dilemma"},
	}
	gaps, accepted := 0, 0
	for _, tc := range tests {
		if tc.wantForm == "" {
			gaps++
		} else {
			accepted++
		}
		t.Run(tc.name, func(t *testing.T) {
			// A syntax failure must never masquerade as a known entailment gap.
			if _, parsed := ParseNumberedArgument(tc.argument, numberedArgumentOptions{}); !parsed.Passed {
				t.Fatalf("invalid ledger fixture: %s\n%s", parsed.Message, tc.argument)
			}
			report, result := CheckEntailmentGate(tc.argument)
			if len(report.Steps) != 1 {
				t.Fatalf("want exactly one gated step: %s %+v", result.Message, report)
			}
			step := report.Steps[0]
			t.Logf("accepted=%v form=%q", report.Passed, step.Form)
			if tc.wantForm == "" {
				if report.Passed || result.Passed || step.Passed {
					t.Fatalf("known gap %s now accepted — set wantForm to %q and move it out of the gap list\n%s", tc.name, step.Form, tc.argument)
				}
				return
			}
			if !report.Passed || !result.Passed || !step.Passed {
				t.Fatalf("expected form %q: %s\n%s", tc.wantForm, result.Message, tc.argument)
			}
			if step.Form != tc.wantForm {
				t.Fatalf("form = %q, want %q", step.Form, tc.wantForm)
			}
		})
	}
	t.Logf("ledger: %d rejection rows (%d valid gaps plus one relation boundary), %d accepted rows", gaps, gaps-1, accepted)
}
