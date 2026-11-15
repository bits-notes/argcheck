package validation

import "testing"

// Found by the FOL mutation fuzz (Nov 2026): a schematic slot could bind text
// that carries a scope-taking operator, so the rule's predicate landed under
// that operator. Each rejected case is an invalid inference that catalog v28
// accepted; each accepted case is a valid control that must keep passing.
func TestEntailmentGateSlotBindingScope(t *testing.T) {
	rule := "2. If X is wise, then X is happy.\n"
	one := func(premise, conclusion string) string {
		return "1. " + premise + "\n" + rule + "3. Therefore, " + conclusion + " [1,2]"
	}
	twice := func(subject string) string {
		return "1. " + subject + " is wise.\n2. " + subject + " is rich.\n" +
			"3. If X is wise and X is rich, then X is happy.\n4. Therefore, " + subject + " is happy. [1,2,3]"
	}
	tests := []struct {
		name     string
		argument string
		want     bool
	}{
		{"negation wrapper", one("It is not the case that Socrates is wise.", "it is not the case that Socrates is happy."), false},
		{"negation under clause wrapper", "1. It is not the case that Socrates is wise.\n2. If it is the case that X is wise, then X is happy.\n3. Therefore, it is not the case that Socrates is happy. [1,2]", false},
		{"falsity wrapper", one("It is false that Socrates is wise.", "it is false that Socrates is happy."), false},
		{"attitude complement", one("Plato denies that Socrates is wise.", "Plato denies that Socrates is happy."), false},
		{"no", one("No philosopher is wise.", "no philosopher is happy."), false},
		{"nobody", one("Nobody is wise.", "nobody is happy."), false},
		{"not every", one("Not every philosopher is wise.", "not every philosopher is happy."), false},
		{"at most", one("At most one philosopher is wise.", "at most one philosopher is happy."), false},
		{"conjunctive rule under negation", "1. It is not the case that Socrates is wise.\n2. Plato is tall.\n3. If it is the case that X is wise and Y is tall, then the city is safe.\n4. Therefore, the city is safe. [1,2,3]", false},

		{"name", one("Socrates is wise.", "Socrates is happy."), true},
		{"every", one("Every philosopher is wise.", "every philosopher is happy."), true},
		{"some", one("Some philosopher is wise.", "some philosopher is happy."), true},
		{"negation inside a relative clause", one("The philosopher who is not rich is wise.", "the philosopher who is not rich is happy."), true},
		{"conditional consequent", one("If Plato is right, Socrates is wise.", "if Plato is right, Socrates is happy."), true},
		{"complete clause inside a noun phrase", one("The claim that the store is nearby is wise.", "the claim that the store is nearby is happy."), true},
		{"repeated name", twice("Socrates"), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report, _ := CheckEntailmentGate(tc.argument)
			if report.Passed != tc.want {
				t.Fatalf("accepted=%v, want %v\n%s", report.Passed, tc.want, tc.argument)
			}
		})
	}
}

// Known gap, not fixed here: an indefinite or disjunctive term bound to a slot
// that occurs twice in the antecedent lets two premises with different
// witnesses combine ("Some philosopher is wise", "Some philosopher is rich").
// The gate still accepts these; the test records it so a fix shows up.
func TestEntailmentGateSlotBindingRepeatedIndefiniteGap(t *testing.T) {
	for _, subject := range []string{"Some philosopher", "A philosopher", "Socrates or Plato"} {
		argument := "1. " + subject + " is wise.\n2. " + subject + " is rich.\n" +
			"3. If X is wise and X is rich, then X is happy.\n4. Therefore, " + subject + " is happy. [1,2,3]"
		if report, _ := CheckEntailmentGate(argument); !report.Passed {
			t.Fatalf("repeated-indefinite gap for %q is now rejected; move it into TestEntailmentGateSlotBindingScope", subject)
		}
	}
}
