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

// v43.0: a variable written just after an indefinite in the rule ("a X") is
// refused in every position: the instance need not be an article plus a noun.
func TestWrittenIndefiniteBeforeVariable(t *testing.T) {
	refused := []string{
		"1. If Bob greets a X, then Bob greets X.\n2. Bob greets a lot of people.\n3. Therefore, Bob greets lot of people. [1,2]",
		"1. If Bob greets a X, then Bob greets someone who is a X.\n2. Bob greets a lot of people.\n3. Therefore, Bob greets someone who is a lot of people. [1,2]",
		"1. If Lois is looking for a X, then a X exists.\n2. Lois is looking for a unicorn.\n3. Therefore, a unicorn exists. [1,2]",
		"1. If Lois is looking for an X, then an X exists.\n2. Lois is looking for an elf.\n3. Therefore, an elf exists. [1,2]",
		"1. If Mary imagines some X, then some X exists.\n2. Mary imagines some dragon.\n3. Therefore, some dragon exists. [1,2]",
		// Subject and predicate positions have the same defect.
		"1. If a X arrived, then someone who is a X arrived.\n2. A lot of people arrived.\n3. Therefore, someone who is a lot of people arrived. [1,2]",
		"1. If Rex is a X, then Fido is a X.\n2. Rex is a lot of fun.\n3. Therefore, Fido is a lot of fun. [1,2]",
		"1. If Rex is a X, then Fido is a X.\n2. Rex is a dog.\n3. Therefore, Fido is a dog. [1,2]",
	}
	for _, text := range refused {
		report, result := CheckEntailmentGate(text)
		if result.Passed {
			t.Fatalf("certified a variable written after an indefinite:\n%s", text)
		}
		message := ""
		for _, step := range report.Steps {
			if !step.Passed {
				message = step.Mismatch
			}
		}
		if !strings.Contains(message, "which the rule writes before the variable") || !strings.Contains(message, "from the rule and let the variable stand for the whole phrase") {
			t.Fatalf("message should name the written indefinite and the repair, got %q for\n%s", message, text)
		}
	}
	// Refused already before v43.0 by another check; must stay refused.
	if _, result := CheckEntailmentGate("1. If Bob greets a X, then Bob greets someone who is a X.\n2. Bob greets a couple of neighbours.\n3. Therefore, Bob greets someone who is a couple of neighbours. [1,2]"); result.Passed {
		t.Fatal("certified \"a couple of neighbours\" after a written indefinite")
	}
	certified := []string{
		// The repair: the variable stands for the whole phrase.
		"1. If Rex is X, then Fido is X.\n2. Rex is a dog.\n3. Therefore, Fido is a dog. [1,2]",
		"1. If X is hungry, then X barks.\n2. A puppy is hungry.\n3. Therefore, a puppy barks. [1,2]",
		// A definite article before the variable is unaffected.
		"1. If Rex is the X, then Fido is the X.\n2. Rex is the dog.\n3. Therefore, Fido is the dog. [1,2]",
	}
	for _, text := range certified {
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Fatalf("should stay certified (%s):\n%s", result.Message, text)
		}
	}
}

// v43.0: an indefinite value for a schematic variable must be a lone
// indefinite pronoun or a/an/some/one, modifiers and one head noun.
func TestIndefiniteValueShape(t *testing.T) {
	refused := []string{
		// Following the object-slot hint ("There is ... that Bob greets") does
		// not reopen it.
		"1. There is a lot of people that Bob greets.\n2. If there is X that Bob greets, then Bob greets someone who is X.\n3. Therefore, Bob greets someone who is a lot of people. [1,2]",
		"1. If X is in the room, then someone who is X is in the room.\n2. A lot of people is in the room.\n3. Therefore, someone who is a lot of people is in the room. [1,2]",
		"1. If X arrived, then X is someone who arrived.\n2. A lot of people arrived.\n3. Therefore, a lot of people is someone who arrived. [1,2]",
		"1. If X arrived, then X is someone who arrived.\n2. A friend of Bob arrived.\n3. Therefore, a friend of Bob is someone who arrived. [1,2]",
		// Closed-class determiners beginning with an article.
		"1. If X arrived, then X is someone who arrived.\n2. A few arrived.\n3. Therefore, a few is someone who arrived. [1,2]",
		"1. If X arrived, then X is someone who arrived.\n2. A few runners arrived.\n3. Therefore, a few runners is someone who arrived. [1,2]",
		"1. If X is in the tank, then X is cold.\n2. A little water is in the tank.\n3. Therefore, a little water is cold. [1,2]",
		"1. If X arrived, then X is someone who arrived.\n2. A great many runners arrived.\n3. Therefore, a great many runners is someone who arrived. [1,2]",
		// The existence restatement does not reopen it.
		"1. There is a lot of people that arrived.\n2. If there is X that arrived, then X is someone who arrived.\n3. Therefore, a lot of people is someone who arrived. [1,2]",
	}
	for _, text := range refused {
		report, result := CheckEntailmentGate(text)
		if result.Passed {
			t.Fatalf("certified a quantity or complement value:\n%s", text)
		}
		message := ""
		for _, step := range report.Steps {
			if !step.Passed {
				message = step.Mismatch
			}
		}
		if !strings.Contains(message, "need not be one thing") || !strings.Contains(message, "Write the phrase itself into the rule") {
			t.Fatalf("message should explain the value shape and the repair, got %q for\n%s", message, text)
		}
	}
	certified := []string{
		"1. If X arrived, then X is someone who arrived.\n2. A dog arrived.\n3. Therefore, a dog is someone who arrived. [1,2]",
		"1. If X arrived, then X is someone who arrived.\n2. A big red dog arrived.\n3. Therefore, a big red dog is someone who arrived. [1,2]",
		"1. If X arrived late, then X missed the briefing.\n2. Someone arrived late.\n3. Therefore, someone missed the briefing. [1,2]",
		// After a copula the value is a predicate; any predicate phrase may
		// be substituted.
		"1. If Rex is X, then Fido is X.\n2. Rex is a lot of fun.\n3. Therefore, Fido is a lot of fun. [1,2]",
		"1. If X is the best agent and X is Y, then the ideal agent is Y.\n2. AIXI is the best agent.\n3. AIXI is a Bayesian agent with the universal prior.\n4. Therefore, the ideal agent is a Bayesian agent with the universal prior. [1,2,3]",
		// Known limit (documented): a bare quantity-noun idiom has the shape
		// of an article and one noun, and the gate has no lexicon to tell
		// "a lot" from "a dog". If this starts failing, update docs/coverage.md.
		"1. If X arrived, then X is someone who arrived.\n2. A lot arrived.\n3. Therefore, a lot is someone who arrived. [1,2]",
		// The suggested repair: the phrase written into the rule.
		"1. If a lot of people arrived, then a lot of people is someone who arrived.\n2. A lot of people arrived.\n3. Therefore, a lot of people is someone who arrived. [1,2]",
	}
	for _, text := range certified {
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Fatalf("should stay certified (%s):\n%s", result.Message, text)
		}
	}
}
