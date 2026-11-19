package validation

import (
	"strings"
	"testing"
)

func TestCatalogV13Forms(t *testing.T) {

	tests := []struct {
		form, name, argument string
		accepted             bool
	}{
		{"contraposition", "forward", `1. if it rains, then it snows.
2. Therefore, if it is not the case that it snows, then it is not the case that it rains. [1]`, true},
		{"contraposition", "reverse", `1. if it is not the case that it snows, then it is not the case that it rains.
2. Therefore, if it rains, then it snows. [1]`, true},
		{"contraposition", "not swapped", `1. if it rains, then it snows.
2. Therefore, if it is not the case that it rains, then it is not the case that it snows. [1]`, false},
		{"contraposition", "one negation missing", `1. if it rains, then it snows.
2. Therefore, if it snows, then it is not the case that it rains. [1]`, false},
		{"De Morgan", "and forward", `1. it is not the case that it is the case both that it rains, and that it snows.
2. Therefore, either it is not the case that it rains or it is not the case that it snows. [1]`, true},
		{"De Morgan", "and reverse", `1. either it is not the case that it rains or it is not the case that it snows.
2. Therefore, it is not the case that it is the case both that it rains, and that it snows. [1]`, true},
		{"De Morgan", "or forward", `1. it is not the case that either it rains or it snows.
2. Therefore, it is the case both that it is not the case that it rains, and that it is not the case that it snows. [1]`, true},
		{"De Morgan", "or reverse", `1. it is the case both that it is not the case that it rains, and that it is not the case that it snows.
2. Therefore, it is not the case that either it rains or it snows. [1]`, true},
		{"De Morgan", "wrong connective", `1. it is not the case that it is the case both that it rains, and that it snows.
2. Therefore, it is the case both that it is not the case that it rains, and that it is not the case that it snows. [1]`, false},
		{"De Morgan", "missing negation", `1. it is not the case that either it rains or it snows.
2. Therefore, it is the case both that it is not the case that it rains, and that it snows. [1]`, false},
		{"De Morgan", "wrong reverse connective", `1. either it is not the case that it rains or it is not the case that it snows.
2. Therefore, it is not the case that either it rains or it snows. [1]`, false},
		{"commutation", "and", `1. it is the case both that it rains, and that it snows.
2. Therefore, it is the case both that it snows, and that it rains. [1]`, true},
		{"commutation", "and different clause", `1. it is the case both that it rains, and that it snows.
2. Therefore, it is the case both that the street is wet, and that it rains. [1]`, false},
		{"commutation", "or", `1. either it rains or it snows.
2. Therefore, either it snows or it rains. [1]`, true},
		{"commutation", "or different clause", `1. either it rains or it snows.
2. Therefore, either the street is wet or it rains. [1]`, false},
		{"commutation", "negation scope", `1. it is not the case that it is the case both that it rains, and that it snows.
2. Therefore, it is the case both that it is not the case that it snows, and that it rains. [1]`, false},
		{"exportation", "forward", `1. if it is the case both that it rains, and that it snows, then the street is wet.
2. Therefore, if it rains, then if it snows, then the street is wet. [1]`, true},
		{"exportation", "disjunctive antecedent", `1. if either it rains or it snows, then the street is wet.
2. Therefore, if it rains, then if it snows, then the street is wet. [1]`, false},
		{"exportation", "wrong consequent", `1. if it is the case both that it rains, and that it snows, then the street is wet.
2. Therefore, if it rains, then if it snows, then it rains. [1]`, false},
		{"importation", "forward", `1. if it rains, then if it snows, then the street is wet.
2. Therefore, if it is the case both that it rains, and that it snows, then the street is wet. [1]`, true},
		{"importation", "disjunctive antecedent", `1. if it rains, then if it snows, then the street is wet.
2. Therefore, if either it rains or it snows, then the street is wet. [1]`, false},
		{"importation", "wrong consequent", `1. if it rains, then if it snows, then the street is wet.
2. Therefore, if it is the case both that it rains, and that it snows, then it rains. [1]`, false},
		{"material implication", "forward", `1. if it rains, then it snows.
2. Therefore, either it is not the case that it rains or it snows. [1]`, true},
		{"material implication", "reverse", `1. either it is not the case that it rains or it snows.
2. Therefore, if it rains, then it snows. [1]`, true},
		{"material implication", "positive antecedent", `1. if it rains, then it snows.
2. Therefore, either it rains or it snows. [1]`, false},
		{"material implication", "negative consequent", `1. either it is not the case that it rains or it snows.
2. Therefore, if it rains, then it is not the case that it snows. [1]`, false},
		{"distribution", "and forward", `1. it is the case both that it rains, and that either it snows or the street is wet.
2. Therefore, either it is the case both that it rains, and that it snows or it is the case both that it rains, and that the street is wet. [1]`, true},
		{"distribution", "and reverse", `1. either it is the case both that it rains, and that it snows or it is the case both that it rains, and that the street is wet.
2. Therefore, it is the case both that it rains, and that either it snows or the street is wet. [1]`, true},
		{"distribution", "or forward", `1. either it rains or it is the case both that it snows, and that the street is wet.
2. Therefore, it is the case both that either it rains or it snows, and that either it rains or the street is wet. [1]`, true},
		{"distribution", "or reverse", `1. it is the case both that either it rains or it snows, and that either it rains or the street is wet.
2. Therefore, either it rains or it is the case both that it snows, and that the street is wet. [1]`, true},
		{"distribution", "missing repeated term", `1. it is the case both that it rains, and that either it snows or the street is wet.
2. Therefore, either it is the case both that it rains, and that it snows or it is the case both that it snows, and that the street is wet. [1]`, false},
		{"distribution", "wrong outer connective", `1. either it rains or it is the case both that it snows, and that the street is wet.
2. Therefore, it is the case both that it rains, and that either it snows or the street is wet. [1]`, false},
		{"absorption", "disjunctive", `1. either it rains or it is the case both that it rains, and that it snows.
2. Therefore, it rains. [1]`, true},
		{"absorption", "conditional", `1. if it rains, then it snows.
2. Therefore, if it rains, then it is the case both that it rains, and that it snows. [1]`, true},
		{"absorption", "other conjunct", `1. either it rains or it is the case both that it rains, and that it snows.
2. Therefore, it snows. [1]`, false},
		{"absorption", "wrong repeated antecedent", `1. if it rains, then it snows.
2. Therefore, if it rains, then it is the case both that the street is wet, and that it snows. [1]`, false},
		{"double negation introduction", "positive", `1. it rains.
2. Therefore, it is not the case that it is not the case that it rains. [1]`, true},
		{"double negation introduction", "wrong clause", `1. it rains.
2. Therefore, it is not the case that it is not the case that it snows. [1]`, false},
		{"double negation introduction", "single negation", `1. it rains.
2. Therefore, it is not the case that it rains. [1]`, false},
		{"biconditional modus tollens", "right denied", `1. it rains if and only if it snows.
2. it is not the case that it snows.
3. Therefore, it is not the case that it rains. [1,2]`, true},
		{"biconditional modus tollens", "left denied reversed citations", `1. it is not the case that it rains.
2. it rains if and only if it snows.
3. Therefore, it is not the case that it snows. [1,2]`, true},
		{"biconditional modus tollens", "wrong denial", `1. it rains if and only if it snows.
2. it is not the case that the street is wet.
3. Therefore, it is not the case that it rains. [1,2]`, false},
		{"biconditional modus tollens", "affirmed result", `1. it rains if and only if it snows.
2. it is not the case that it snows.
3. Therefore, it rains. [1,2]`, false},
		{"biconditional chaining", "forward", `1. it rains if and only if it snows.
2. it snows if and only if the street is wet.
3. Therefore, it rains if and only if the street is wet. [1,2]`, true},
		{"biconditional chaining", "reversed citations", `1. it snows if and only if the street is wet.
2. it rains if and only if it snows.
3. Therefore, it rains if and only if the street is wet. [1,2]`, true},
		{"biconditional chaining", "middle mismatch", `1. it rains if and only if it snows.
2. Ada sings if and only if the street is wet.
3. Therefore, it rains if and only if the street is wet. [1,2]`, false},
		{"biconditional chaining", "wrong endpoint", `1. it rains if and only if it snows.
2. it snows if and only if the street is wet.
3. Therefore, it rains if and only if Ada sings. [1,2]`, false},
		{"biconditional commutation", "forward", `1. it rains if and only if it snows.
2. Therefore, it snows if and only if it rains. [1]`, true},
		{"biconditional commutation", "wrong left", `1. it rains if and only if it snows.
2. Therefore, the street is wet if and only if it rains. [1]`, false},
		{"biconditional commutation", "wrong right", `1. it rains if and only if it snows.
2. Therefore, it snows if and only if the street is wet. [1]`, false},
		{"negated conditional", "antecedent", `1. it is not the case that if it rains, then it snows.
2. Therefore, it rains. [1]`, true},
		{"negated conditional", "denied consequent", `1. it is not the case that if it rains, then it snows.
2. Therefore, it is not the case that it snows. [1]`, true},
		{"negated conditional", "affirmed consequent", `1. it is not the case that if it rains, then it snows.
2. Therefore, it snows. [1]`, false},
		{"negated conditional", "denied antecedent", `1. it is not the case that if it rains, then it snows.
2. Therefore, it is not the case that it rains. [1]`, false},
		{"identity symmetry", "forward", `1. Ada is identical to Ben.
2. Therefore, Ben is identical to Ada. [1]`, true},
		{"identity symmetry", "different subject", `1. Ada is identical to Ben.
2. Therefore, Cy is identical to Ada. [1]`, false},
		{"identity symmetry", "different object", `1. Ada is identical to Ben.
2. Therefore, Ben is identical to Cy. [1]`, false},
		{"identity symmetry", "opaque report", `1. Lois believes Ada is identical to Ben.
2. Therefore, Ben is identical to Lois believes Ada. [1]`, false},
		{"quantifier negation", "universal", `1. it is not the case that every man is mortal.
2. Therefore, some man is not mortal. [1]`, true},
		{"quantifier negation", "existential", `1. it is not the case that some man is immortal.
2. Therefore, no man is immortal. [1]`, true},
		{"quantifier negation", "universal wrong class", `1. it is not the case that every man is mortal.
2. Therefore, some poet is not mortal. [1]`, false},
		{"quantifier negation", "existential wrong predicate", `1. it is not the case that some man is immortal.
2. Therefore, no man is mortal. [1]`, false},
		{"quantifier negation", "universal affirmative", `1. it is not the case that every man is mortal.
2. Therefore, some man is mortal. [1]`, false},
		{"conversion", "existential", `1. some poet is a sailor.
2. Therefore, some sailor is a poet. [1]`, true},
		{"conversion", "negative", `1. no poet is a sailor.
2. Therefore, no sailor is a poet. [1]`, true},
		{"conversion", "an", `1. some artist is an engineer.
2. Therefore, some engineer is an artist. [1]`, true},
		{"conversion", "invalid universal", `1. every poet is a sailor.
2. Therefore, every sailor is a poet. [1]`, false},
		{"conversion", "wrong class", `1. some poet is a sailor.
2. Therefore, some sailor is an artist. [1]`, false},
		{"conversion", "invalid particular negative", `1. some poet is not a sailor.
2. Therefore, some sailor is not a poet. [1]`, false},
		{"conversion", "negative wrong class", `1. no poet is a sailor.
2. Therefore, no sailor is an artist. [1]`, false},
		{"universal modus tollens", "forward", `1. every man is mortal.
2. it is not the case that Zeus is mortal.
3. Therefore, it is not the case that Zeus is a man. [1,2]`, true},
		{"universal modus tollens", "an reversed citations", `1. it is not the case that Ada is mortal.
2. every artist is mortal.
3. Therefore, it is not the case that Ada is an artist. [1,2]`, true},
		{"universal modus tollens", "wrong class", `1. every man is mortal.
2. it is not the case that Zeus is mortal.
3. Therefore, it is not the case that Zeus is a poet. [1,2]`, false},
		{"universal modus tollens", "wrong instance", `1. every man is mortal.
2. it is not the case that Zeus is mortal.
3. Therefore, it is not the case that Ada is a man. [1,2]`, false},
		{"universal modus tollens", "wrong predicate", `1. every man is mortal.
2. it is not the case that Zeus is immortal.
3. Therefore, it is not the case that Zeus is a man. [1,2]`, false},
		{"possibility introduction", "actual", `1. it rains.
2. Therefore, possibly, it rains. [1]`, true},
		{"possibility introduction", "necessary", `1. necessarily, it rains.
2. Therefore, possibly, it rains. [1]`, true},
		{"possibility introduction", "wrong clause", `1. it rains.
2. Therefore, possibly, it snows. [1]`, false},
		{"possibility introduction", "possibility to necessity", `1. possibly, it rains.
2. Therefore, necessarily, it rains. [1]`, false},
		{"possibility introduction", "opaque identity substitution", `1. Ada is identical to Ben.
2. possibly, Ada sings.
3. Therefore, possibly, Ben sings. [1,2]`, false},
		{"possibility introduction", "opaque existential generalization", `1. possibly, Ada sings.
2. Therefore, there is an X such that possibly, X sings. [1]`, false},
		{"possibility introduction", "actual to necessity", `1. it rains.
2. Therefore, necessarily, it rains. [1]`, false},
	}
	for _, tc := range tests {
		t.Run(tc.form+"/"+tc.name, func(t *testing.T) {
			parsed, syntax := ParseNumberedArgument(tc.argument, numberedArgumentOptions{})
			if !syntax.Passed {
				t.Fatalf("bad fixture: %s", syntax.Message)
			}
			report, result := CheckEntailmentGate(tc.argument)
			if report.Passed != tc.accepted || result.Passed != tc.accepted {
				t.Fatalf("want accepted=%v: %s\n%s", tc.accepted, result.Message, tc.argument)
			}
			if len(report.Steps) != 1 {
				t.Fatalf("want one checked step: %+v", report)
			}
			if tc.accepted && report.Steps[0].Form != tc.form {
				t.Fatalf("form=%q want %q", report.Steps[0].Form, tc.form)
			}
			if !tc.accepted {
				for _, schema := range entailmentSchemas {
					if schema.name != tc.form || schema.arity != len(parsed.Steps)-1 {
						continue
					}
					outcome := schema.match(parsed.Steps[:len(parsed.Steps)-1], parsed.Steps[len(parsed.Steps)-1])
					if outcome.matched || !strings.Contains(outcome.mismatch, "\"") {
						t.Fatalf("matcher must reject with a quoted clause: %+v", outcome)
					}
				}
			}
		})
	}
}
