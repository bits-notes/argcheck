package validation

import (
	"strings"
	"testing"
)

func TestIdentitySubstitutionSafePositions(t *testing.T) {
	const identity = "\n2. Superman is identical to Clark.\n"
	reason := "substitution is certified only in identity sentences and in the subject of a plain sentence"
	for _, tc := range []struct {
		name, premise, conclusion string
		valid                     bool
	}{
		{"looking for", "Lois is looking for Superman", "Lois is looking for Clark", false},
		{"thinking about", "Lois is thinking about Superman", "Lois is thinking about Clark", false},
		{"admires", "Lois admires Superman", "Lois admires Clark", false},
		{"seeks", "Lois seeks Superman", "Lois seeks Clark", false},
		{"worships", "Lois worships Superman", "Lois worships Clark", false},
		{"imagines", "Lois imagines Superman", "Lois imagines Clark", false},
		{"needs", "Lois needs Superman", "Lois needs Clark", false},
		{"extensional object too", "Lois kissed Superman", "Lois kissed Clark", false},
		{"passive", "Superman is admired by Lois", "Clark is admired by Lois", false},
		{"raising", "Superman is believed to be a hero", "Clark is believed to be a hero", false},
		{"subject", "Superman can fly", "Clark can fly", true},
		{"subject copula", "Superman is a reporter", "Clark is a reporter", true},
		{"subject modal must", "Superman must be a hero", "Clark must be a hero", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := "1. " + tc.premise + "." + identity + "3. Therefore, " + tc.conclusion + ". [1,2]"
			report, result := CheckEntailmentGate(text)
			if result.Passed != tc.valid {
				t.Fatalf("got %v: %s", result.Passed, result.Message)
			}
			if !tc.valid && tc.name != "imagines" && tc.name != "worships" && tc.name != "needs" && tc.name != "seeks" && tc.name != "looking for" {
				if len(report.Steps) == 0 || !strings.Contains(report.Steps[len(report.Steps)-1].Mismatch, reason) {
					t.Fatalf("message lacks the position reason: %+v", report.Steps)
				}
			}
		})
	}
}

func TestIdentitySubstitutionChainsAndUnicorn(t *testing.T) {
	chain := "1. Superman is identical to Clark.\n2. Clark is identical to Kent.\n3. Therefore, Superman is identical to Kent. [1,2]"
	if _, result := CheckEntailmentGate(chain); !result.Passed {
		t.Fatalf("identity chain refused: %s", result.Message)
	}
	for _, text := range []string{
		"1. Lois is looking for a unicorn.\n2. Therefore, there is a unicorn that Lois is looking for. [1]",
		"1. Lois is looking for a unicorn.\n2. Therefore, something is a unicorn. [1]",
	} {
		if _, result := CheckEntailmentGate(text); result.Passed {
			t.Fatalf("unicorn existence certified: %s", text)
		}
	}
	quantified := "1. alice is identical to blair.\n2. every x is such that (alice admires x).\n3. Therefore, every x is such that (blair admires x). [1,2]"
	if _, result := CheckEntailmentGate(quantified); !result.Passed {
		t.Fatalf("subject inside a quantifier scope refused: %s", result.Message)
	}
}

// task_9b65b028: an indefinite ("a unicorn", "someone") binds a schematic
// variable only as a clause subject or after a copula. In an object slot it
// need not pick anything out, so the schema is not instantiated with it.
func TestIndefiniteSchematicBindingStrictSlots(t *testing.T) {
	refused := []string{
		"1. If Lois is looking for X, then X exists.\n2. Lois is looking for a unicorn.\n3. Therefore, a unicorn exists. [1,2]",
		"1. If Lois wants X, then X exists.\n2. Lois wants a sloop.\n3. Therefore, a sloop exists. [1,2]",
		"1. If Ponce de Leon seeks X, then X is real.\n2. Ponce de Leon seeks a fountain of youth.\n3. Therefore, a fountain of youth is real. [1,2]",
		"1. If the expedition needs X, then X is available.\n2. The expedition needs a guide.\n3. Therefore, a guide is available. [1,2]",
		"1. If Mary imagines X, then X exists.\n2. Mary imagines a golden mountain.\n3. Therefore, a golden mountain exists. [1,2]",
		"1. If Tom is thinking about X, then X is nearby.\n2. Tom is thinking about someone.\n3. Therefore, someone is nearby. [1,2]",
		// Extensional control: refused too (the rule is positional, not lexical).
		"1. If Lois admires X, then X is admirable.\n2. Lois admires a hero.\n3. Therefore, a hero is admirable. [1,2]",
	}
	for _, text := range refused {
		report, result := CheckEntailmentGate(text)
		if result.Passed {
			t.Fatalf("certified an object-slot indefinite binding:\n%s", text)
		}
		message := ""
		for _, step := range report.Steps {
			if !step.Passed {
				message = step.Mismatch
			}
		}
		if !strings.Contains(message, "in an object position") || !strings.Contains(message, "State the existence claim as its own premise, e.g. \"There is ") {
			t.Fatalf("message should name the slot and suggest a restatement, got %q for\n%s", message, text)
		}
	}
	certified := []string{
		// Subject position.
		"1. If X is a dog, then X barks.\n2. A puppy is a dog.\n3. Therefore, a puppy barks. [1,2]",
		"1. If X arrived late, then X missed the briefing.\n2. Someone arrived late.\n3. Therefore, someone missed the briefing. [1,2]",
		// Predicate after a copula.
		"1. If Rex is X, then Fido is X.\n2. Rex is a dog.\n3. Therefore, Fido is a dog. [1,2]",
		// Names and definite descriptions in object slots are unaffected.
		"1. If Lois admires X, then X is admirable.\n2. Lois admires Superman.\n3. Therefore, Superman is admirable. [1,2]",
		"1. If Lois admires X, then X is admirable.\n2. Lois admires the mayor.\n3. Therefore, the mayor is admirable. [1,2]",
		// The suggested rewrite route: the existence claim stated as a premise.
		"1. There is a unicorn that Lois is looking for.\n2. If there is a unicorn that Lois is looking for, then a unicorn exists.\n3. Therefore, a unicorn exists. [1,2]",
	}
	for _, text := range certified {
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Fatalf("should stay certified (%s):\n%s", result.Message, text)
		}
	}
	if got := indefiniteRestatement("Lois is looking for a unicorn.", "a unicorn"); got != "There is a unicorn that Lois is looking for." {
		t.Fatalf("restatement = %q", got)
	}
}
