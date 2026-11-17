package validation

import "testing"

// v31.2: "Y if X." reads exactly as "If X, then Y." wherever a conditional
// premise is accepted. Ambiguous or embedded uses of "if" are not rewritten.
func TestPostposedConditional(t *testing.T) {
	tests := []struct {
		name     string
		argument string
		want     bool
	}{
		{"modus ponens", "1. Steve will wear a raincoat tomorrow if it rains.\n2. It rains.\n3. Therefore, Steve will wear a raincoat tomorrow. [1,2]", true},
		{"comma variant", "1. Steve will wear a raincoat tomorrow, if it rains.\n2. It rains.\n3. Therefore, Steve will wear a raincoat tomorrow. [1,2]", true},
		{"modus tollens", "1. The match is cancelled if it rains.\n2. It is not the case that the match is cancelled.\n3. Therefore, it is not the case that it rains. [1,2]", true},
		{"hypothetical syllogism", "1. The street is wet if it rains.\n2. Cars slow down if the street is wet.\n3. Therefore, if it rains, then cars slow down. [1,2]", true},
		{"conjunctive antecedent", "1. The picnic is cancelled if it rains and the park is closed.\n2. It rains and the park is closed.\n3. Therefore, the picnic is cancelled. [1,2]", true},
		{"step reference", "1. The street is wet.\n2. It rained if [1].\n3. Therefore, it rained. [1,2]", true},
		{"mixed with a preposed conditional", "1. If it rains, then the street is wet.\n2. Cars slow down if the street is wet.\n3. Therefore, if it rains, then cars slow down. [1,2]", true},

		{"only if is not forward", "1. Steve wears a raincoat only if it rains.\n2. It rains.\n3. Therefore, Steve wears a raincoat. [1,2]", false},
		{"even if is not a conditional", "1. Steve wears a raincoat even if it rains.\n2. It rains.\n3. Therefore, Steve wears a raincoat. [1,2]", false},
		{"as if is not a conditional", "1. Steve acts as if it rains.\n2. It rains.\n3. Therefore, Steve acts. [1,2]", false},
		{"if and only if is not postposed", "1. Steve wears a raincoat if and only if it rains.\n2. It rains.\n3. Therefore, Steve wears a raincoat if and only if. [1,2]", false},
		{"unless is not rewritten", "1. Steve stays home unless it rains if the bus is late.\n2. The bus is late.\n3. Therefore, Steve stays home unless it rains. [1,2]", false},
		{"embedded question", "1. Ann asked if it rains.\n2. It rains.\n3. Therefore, Ann asked. [1,2]", false},
		{"wonder if", "1. Ann wonders if it rains.\n2. It rains.\n3. Therefore, Ann wonders. [1,2]", false},
		{"ambiguous consequent scope", "1. The street is wet and the sky is grey if it rains.\n2. It rains.\n3. Therefore, the street is wet and the sky is grey. [1,2]", false},
		{"disjunctive antecedent not rewritten", "1. The street is wet if it rains or it snows.\n2. It rains or it snows.\n3. Therefore, the street is wet. [1,2]", false},
		{"two ifs", "1. The street is wet if it rains if the sky is grey.\n2. The sky is grey.\n3. Therefore, the street is wet if it rains. [1,2]", false},
		{"affirming the consequent stays invalid", "1. The street is wet if it rains.\n2. The street is wet.\n3. Therefore, it rains. [1,2]", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, report, msg := checkRulesDocArgument(tc.argument)
			if ok != tc.want {
				t.Fatalf("valid=%v, want %v: %s\nsteps=%+v\n%s", ok, tc.want, msg, report.Steps, tc.argument)
			}
		})
	}
}
