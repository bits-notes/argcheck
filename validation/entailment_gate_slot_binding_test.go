package validation

import (
	"strings"
	"testing"
)

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

		// v30: a slot repeated in the antecedent binds one fixed thing.
		{"repeated some", twice("Some philosopher"), false},
		{"repeated a", twice("A philosopher"), false},
		{"repeated disjunction", twice("Socrates or Plato"), false},
		{"repeated someone", twice("Someone"), false},
		{"repeated number", twice("Two philosophers"), false},
		{"repeated slot inside one premise", "1. Some philosopher admires some philosopher.\n2. If X admires X, then X is vain.\n3. Therefore, some philosopher is vain. [1,2]", false},
		{"universal spread in the consequent", "1. Every philosopher is wise.\n2. If X is wise, then X admires X.\n3. Therefore, every philosopher admires every philosopher. [1,2]", false},
		{"repeated definite description", twice("The oldest philosopher"), true},
	}
	refused := []struct{ name, argument, hint string }{
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"name", one("Socrates is wise.", "Socrates is happy."), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"every", one("Every philosopher is wise.", "every philosopher is happy."), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"some", one("Some philosopher is wise.", "some philosopher is happy."), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"negation inside a relative clause", one("The philosopher who is not rich is wise.", "the philosopher who is not rich is happy."), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"conditional consequent", one("If Plato is right, Socrates is wise.", "if Plato is right, Socrates is happy."), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"complete clause inside a noun phrase", one("The claim that the store is nearby is wise.", "the claim that the store is nearby is happy."), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"repeated name", twice("Socrates"), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"repeated every", twice("Every philosopher"), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"definite description with or", twice("The philosopher who writes or teaches"), "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"single every", "1. Every philosopher is wise.\n2. If X is wise and Plato is rich, then X is happy.\n3. Plato is rich.\n4. Therefore, every philosopher is happy. [1,3,2]", "Name the thing"},
		// slot.value_singular: only evidenced names and bounded singular descriptions fill term slots.
		{"one premise holding both conjuncts", "1. Some philosopher is wise and rich.\n2. If X is wise and rich, then X is happy.\n3. Therefore, some philosopher is happy. [1,2]", "Name the thing"},
	}
	for _, tc := range refused {
		t.Run(tc.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if report.Passed || result.Passed {
				t.Fatal("strict refusal certified")
			}
			if !strings.Contains(result.Message, tc.hint) {
				t.Fatalf("missing %q: %s", tc.hint, result.Message)
			}
		})
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
