package validation

import (
	"strings"
	"testing"
)

func TestIdentitySubstitutionSafePositions(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// identity.constant_sides: identity sides require evidenced names or bounded singular descriptions.
		"TestIdentitySubstitutionSafePositions/subject": {"identity.constant_sides", "name the thing"},
		// identity.constant_sides: identity sides require evidenced names or bounded singular descriptions.
		"TestIdentitySubstitutionSafePositions/subject_copula": {"identity.constant_sides", "name the thing"},
		// identity.constant_sides: identity sides require evidenced names or bounded singular descriptions.
		"TestIdentitySubstitutionSafePositions/subject_modal_must": {"identity.constant_sides", "name the thing"},
	}

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
			if listedStrictRefusal(t, text, refused) {
				return
			}
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
	// identity.constant_sides: Superman has no non-initial name evidence.
	refused := []string{chain}
	for _, text := range refused {
		assertStrictRefusal(t, text, "identity.constant_sides", "name the thing")
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
	// identity.constant_sides: v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit).
	assertStrictRefusal(t, quantified, "identity.constant_sides", "name the thing")
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
		"1. If X is a dog, then X barks.\n2. The puppy is a dog.\n3. Therefore, the puppy barks. [1,2]",
		"1. If X arrived late, then X missed the briefing.\n2. The visitor arrived late.\n3. Therefore, the visitor missed the briefing. [1,2]",
		// Predicate after a copula.
		"1. If Rex is X, then Fido is X.\n2. Rex is happy.\n3. Therefore, Fido is happy. [1,2]",
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
		"1. If Rex is X, then Fido is X.\n2. Rex is happy.\n3. Therefore, Fido is happy. [1,2]",
		"1. If X is hungry, then X barks.\n2. The puppy is hungry.\n3. Therefore, the puppy barks. [1,2]",
		// A whole singular definite description can fill the variable.
		"1. If Rex is X, then Fido is X.\n2. Rex is the dog.\n3. Therefore, Fido is the dog. [1,2]",
	}
	for _, text := range certified {
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Fatalf("should stay certified (%s):\n%s", result.Message, text)
		}
	}
}

// v44.0: term values must pass the singular-term allowlist.
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
	for _, value := range []string{"a lot", "lots", "lots of people", "hundreds of people", "tons of people", "most", "all", "half", "each", "both", "enough", "plenty", "dozens", "two people", "a dozen", "a bunch", "a handful", "a ton", "Bob and Carol", "Fido or Rex", "everyone", "somebody", "the few", "the most people", "the two", "the dogs", "the majority", "the minority", "the rest", "the bulk", "the remainder", "the lot", "the whole", "a dog", "a big red dog", "someone", "an apple", "something", "some dog", "one dog"} {
		refused = append(refused, "1. If X arrived, then X is someone who arrived.\n2. "+strings.ToUpper(value[:1])+value[1:]+" arrived.\n3. Therefore, "+strings.ToLower(value)+" is someone who arrived. [1,2]")
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
		// After a copula the value is a predicate; any predicate phrase may
		// be substituted.
		"1. If Rex is X, then Fido is X.\n2. Rex is a lot of fun.\n3. Therefore, Fido is a lot of fun. [1,2]",
		"1. If X is the best agent and X is Y, then the ideal agent is Y.\n2. AIXI is the best agent.\n3. AIXI is a Bayesian agent with the universal prior.\n4. Therefore, the ideal agent is a Bayesian agent with the universal prior. [1,2,3]",
		// The suggested repair: the phrase written into the rule.
		"1. If a lot of people arrived, then a lot of people is someone who arrived.\n2. A lot of people arrived.\n3. Therefore, a lot of people is someone who arrived. [1,2]",
	}
	for _, text := range certified {
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Fatalf("should stay certified (%s):\n%s", result.Message, text)
		}
	}
}

func TestSingularValueAllowlist(t *testing.T) {
	for _, value := range []string{"Rex", "The dog", "the dog", "the red dog", "Bob's dog", "The committee", "the committee"} {
		text := "1. If X arrived, then X was seen.\n2. " + value + " arrived.\n3. Carol saw Rex and Bob.\n4. Therefore, " + value + " was seen. [1,2]"
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Errorf("%s: %s", value, result.Message)
		}
	}
	for _, text := range []string{
		"1. If X arrived, then X was seen.\n2. Rex arrived.\n3. Therefore, Rex was seen. [1,2]",
		"1. If X arrived, then X was seen.\n2. The Bank of England arrived.\n3. Therefore, the Bank of England was seen. [1,2]",
		"1. If Rex is X, then X arrived.\n2. Rex is happy.\n3. Therefore, happy arrived. [1,2]",
	} {
		if _, result := CheckEntailmentGate(text); result.Passed {
			t.Errorf("certified: %s", text)
		}
	}
	for _, text := range []string{
		"1. If Rex is X, then Fido is X.\n2. Rex is happy.\n3. Therefore, Fido is happy. [1,2]",
		"1. If Rex is X, then Fido is X.\n2. Rex is a hero.\n3. Therefore, Fido is a hero. [1,2]",
		// Internal capitals are name evidence even at sentence start.
		"1. If X arrived, then X was seen.\n2. AIXI arrived.\n3. Therefore, AIXI was seen. [1,2]",
	} {
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Errorf("%s\n%s", result.Message, text)
		}
	}
	// Ordinary initial capitals are not.
	if _, result := CheckEntailmentGate("1. If X arrived, then X is someone who arrived.\n2. Lots arrived.\n3. Therefore, Lots is someone who arrived. [1,2]"); result.Passed {
		t.Error("certified 'Lots' as a name")
	}
}

func TestSingularValuePositions(t *testing.T) {
	for _, pattern := range []string{"X arrived", "Rex admires X", "Rex stands beside X", "Rex is identical to X", "Rex is X"} {
		for _, value := range []string{"the dog", "Bob", "a dog", "the dogs", "Bob and Carol", "everyone"} {
			text := "1. If " + pattern + ", then " + pattern + " again.\n2. " + strings.ReplaceAll(pattern, "X", value) + ".\n3. Carol saw Bob.\n4. Therefore, " + strings.ReplaceAll(pattern, "X", value) + " again. [1,2]"
			want := value == "the dog" || value == "Bob"
			// After a copula X is a predicate: any phrase may be substituted.
			// ("Rex is Bob and Carol again" is refused by scope.clause_start:
			// its "and" may join clauses.)
			if pattern == "Rex is X" {
				want = value != "Bob and Carol"
			}
			if _, result := CheckEntailmentGate(text); result.Passed != want {
				t.Errorf("pattern=%q value=%q want=%v: %s", pattern, value, want, result.Message)
			}
		}
	}
	// Each capitalized token needs evidence outside a sentence/conclusion start.
	for _, lead := range []string{"Therefore,", "So,", "Thus,", "Hence,"} {
		text := "1. If X arrived, then X was seen.\n2. Rex arrived.\n3. " + lead + " Rex was seen. [1,2]"
		if _, result := CheckEntailmentGate(text); result.Passed {
			t.Errorf("initial capitalization counted: %s", lead)
		}
	}
	for _, value := range []string{"Alice Smith", "Bob's dog", "Bob’s dog"} {
		text := "1. If X arrived, then X was seen.\n2. " + value + " arrived.\n3. Carol saw " + value + ".\n4. Therefore, " + value + " was seen. [1,2]"
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Errorf("%s: %s", value, result.Message)
		}
	}
	// A whole clause retains its previous propositional interpretation.
	text := "1. If the bell rings when X, then the lamp glows when X.\n2. The bell rings when the dogs are happy.\n3. Therefore, the lamp glows when the dogs are happy. [1,2]"
	if _, result := CheckEntailmentGate(text); !result.Passed {
		t.Fatal(result.Message)
	}
}

func TestSingularUnknownCopularPredicates(t *testing.T) {
	for _, value := range []string{"bob", "people", "water", "euphoric"} {
		text := "1. If Rex is X, then Fido is X.\n2. Rex is " + value + ".\n3. Therefore, Fido is " + value + ". [1,2]"
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Errorf("predicate-only copular value %s refused: %s", value, result.Message)
		}
	}
}

// v44.0: "the N of D" with a singular definite or name complement is a
// singular term; quantity, plural, indefinite or nested complements are not.
func TestSingularValueOfComplement(t *testing.T) {
	for _, value := range []string{"the death of the child", "the author of Waverley", "the capital of France"} {
		text := "1. If X is bad, then X should be prevented.\n2. " + strings.ToUpper(value[:1]) + value[1:] + " is bad.\n3. Bob saw Waverley and France.\n4. Therefore, " + value + " should be prevented. [1,2]"
		if _, r := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("%s: %s", value, r.Message)
		}
	}
	for _, value := range []string{"the majority of voters", "the majority of the voters", "the rest of them", "the friend of a friend", "the death of many children", "the death of the child of the king", "the fear of the observer dreamt Alpha"} {
		text := "1. If X is bad, then X should be prevented.\n2. " + strings.ToUpper(value[:1]) + value[1:] + " is bad.\n3. Therefore, " + value + " should be prevented. [1,2]"
		if _, r := CheckEntailmentGate(text); r.Passed {
			t.Errorf("certified %q", value)
		}
	}
}
