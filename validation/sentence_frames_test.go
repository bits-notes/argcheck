package validation

import (
	"strings"
	"testing"
)

func TestSentenceFrames(t *testing.T) {
	// scope.clause_start: an unknown frame verb is refused even in an
	// explicitly parenthesised operand; retain the exact historical case.
	strictRefused := map[string]strictRefusal{
		"TestSentenceFrames/outer-parentheses": {"scope.clause_start", "move the operator into each clause"},
	}
	cases := []struct {
		name, text, message string
		valid               bool
	}{
		{"plain", "1. The server crashed and the network failed.\n2. Therefore, the network failed. [1]", "", true},
		{"both", "1. Both the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "", true},
		{"reckons", "1. Bob reckons the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "move the operator into each clause", false},
		{"fancies", "1. Bob fancies the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "move the operator into each clause", false},
		{"insinuated", "1. Bob insinuated the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "move the operator into each clause", false},
		{"owner", "1. The owner of the robot believes alpha and beta.\n2. Therefore, The owner of the robot believes alpha. [1]", "move the operator into each clause", false},
		{"introduction", "1. Bob reckons the server crashed.\n2. The network failed.\n3. Therefore, Bob reckons the server crashed and the network failed. [1,2]", "move the operator into each clause", false},
		{"embedded-parentheses", "1. Bob reckons (the server crashed and the network failed).\n2. Therefore, the network failed. [1]", "", false},
		{"outer-parentheses", "1. (Bob reckons the server crashed) and (the network failed).\n2. Therefore, the network failed. [1]", "", true},
		{"restate-parentheses", "1. Bob reckons (the server crashed and the network failed).\n2. Therefore, Bob reckons (the server crashed and the network failed). [1]", "", true},
		{"unsafe-universal", "1. Every man is mortal.\n2. Ostensibly Socrates is a man.\n3. Therefore, Ostensibly Socrates is mortal. [1,2]", "may cover the whole sentence", false},
		{"unsafe-schematic", "1. If X is good, then X is valuable.\n2. Purportedly the server is good.\n3. Therefore, Purportedly the server is valuable. [1,2]", "may cover the whole sentence", false},
		{"opaque-reiteration", "1. Hypothetically the server crashed.\n2. Therefore, Hypothetically the server crashed. [1]", "", true},
		{"universal", "1. Every man is mortal.\n2. Socrates is a man.\n3. Therefore, Socrates is mortal. [1,2]", "", true},
		{"bare-noun", "1. Consciousness is real and the server crashed.\n2. Therefore, the server crashed. [1]", "", true},
		{"name", "1. Plan B is good and the server crashed.\n2. Therefore, the server crashed. [1]", "", true},
		{"panda", "1. Every panda eats shoots and leaves.\n2. Bao is a panda.\n3. Therefore, Bao eats, shoots, and leaves. [1,2]", "differs only in punctuation", false},
		{"relative", "1. Every philosopher who is wise is happy.\n2. Therefore, Every philosopher, who is wise, is happy. [1]", "differs only in punctuation", false},
	}
	for _, opener := range []string{"Arguably", "Purportedly", "Ostensibly", "Hypothetically", "Seemingly", "In the story,"} {
		cases = append(cases, struct {
			name, text, message string
			valid               bool
		}{opener, "1. " + opener + " the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "move the operator into each clause", false})
	}
	// slot.value_singular: gerund phrases are not admitted singular term values.
	refused := []struct{ name, text, hint string }{
		{"gerund", "1. If X is good, then X is valuable.\n2. Trying to remove all assumptions is good.\n3. Therefore, Trying to remove all assumptions is valuable. [1,2]", "Name the thing"},
	}
	for _, tc := range refused {
		t.Run(tc.name, func(t *testing.T) { assertStrictRefusal(t, tc.text, "slot.value_singular", tc.hint) })
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if listedStrictRefusal(t, tc.text, strictRefused) {
				return
			}
			report, result := CheckEntailmentGate(tc.text)
			if report.Passed != tc.valid {
				t.Errorf("passed=%v want %v: %+v / %+v", report.Passed, tc.valid, report, result)
			}
			messages := result.Message
			for _, s := range report.Steps {
				messages += " " + s.Mismatch
			}
			if !strings.Contains(messages, tc.message) {
				t.Errorf("missing %q: %s", tc.message, messages)
			}
		})
	}
}

func TestSentenceFramesDisjunction(t *testing.T) {
	for _, tc := range []struct{ name, text string }{
		{"introduction", "1. The network failed.\n2. Therefore, Bob reckons the server crashed or the network failed. [1]"},
		{"elimination", "1. Bob reckons the server crashed or the network failed.\n2. It is not the case that Bob reckons the server crashed.\n3. Therefore, the network failed. [1,2]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := CheckEntailmentGate(tc.text)
			if r.Passed {
				t.Fatal("unsafe complement coordination certified")
			}
			if len(r.Steps) == 0 || !strings.Contains(r.Steps[0].Mismatch, "move the operator into each clause") {
				t.Fatalf("missing grouping hint: %+v", r)
			}
		})
	}
}

func TestSentenceFramesOpeners(t *testing.T) {
	for _, s := range []string{"arguably the server crashed", "Maybe the server crashed", "In the story, the server crashed", "Purportedly the server crashed"} {
		if sentenceFrameOpener(s, nil) == "" {
			t.Errorf("certified %q", s)
		}
	}
	for _, s := range []string{"Consciousness is real", "consciousness is real", "Human value consists of welfare", "Trying to remove all assumptions is difficult", "The owner of the robot is happy", "Plan B is good", "plan B is good", "Alice struck x"} {
		if why := sentenceFrameOpener(s, nil); why != "" {
			t.Errorf("refused %q: %s", s, why)
		}
	}
}

func TestSentenceFramePunctuationKeys(t *testing.T) {
	for _, pair := range [][2]string{
		{`atom{"p, and q"}`, `atom{"p and q"}`},
		{"Every panda eats shoots, and leaves", "Every panda eats shoots and leaves"},
		{"Bob ran, and jumped", "Bob ran and jumped"},
	} {
		if equalClauses(pair[0], pair[1]) {
			t.Errorf("lexical comma lost: %q / %q", pair[0], pair[1])
		}
	}
	if !equalClauses("Ann sings, and Bob dances", "Ann sings and Bob dances") {
		t.Fatal("structural clause comma was not recognized")
	}
}

func TestModifiedQuantifierSlots(t *testing.T) {
	for _, q := range []string{"Practically no person", "Virtually no person", "Next to no person", "Precisely one person", "Roughly one person"} {
		ui := "1. Every person is mortal.\n2. " + q + " is a philosopher.\n3. Therefore, " + q + " is mortal. [1,2]"
		schema := "1. If X is a philosopher, then X is wise.\n2. " + q + " is a philosopher.\n3. Therefore, " + q + " is wise. [1,2]"
		for _, text := range []string{ui, schema} {
			if _, r := CheckEntailmentGate(text); r.Passed {
				t.Errorf("modified quantifier bound as a term: %s", text)
			}
		}
	}
	for _, ok := range []string{"carefully designed systems", "the roughly built shed", "consciousness"} {
		if modifiedQuantifierPhrase(ok) {
			t.Errorf("not a quantifier phrase: %q", ok)
		}
	}
}

func TestClosedClassOpeners(t *testing.T) {
	refuse := []string{
		"According to Bob,", "According to Bob", "According to the report,", "As per the log,", "In view of the outage,",
		"With respect to the logs,", "On behalf of the team,", "In terms of uptime,", "Despite the alarm,", "Unlike yesterday,",
		"Regarding the incident,", "Given the logs,", "Of course", "No doubt", "In theory,", "In principle,", "In practice,",
		"So to speak,", "In effect,", "Still,", "Even so,", "Indeed,", "Perhaps", "Maybe", "Surely", "Certainly",
	}
	for _, opener := range refuse {
		text := "1. " + opener + " the server crashed and the network failed.\n2. Therefore, the network failed. [1]"
		if _, r := CheckEntailmentGate(text); r.Passed {
			t.Errorf("opener %q certified a sentence-internal step", opener)
		}
	}
	// eg.name_witness: quantified descriptions are not fixed named witnesses.
	strictRefused := []string{
		"1. Almost every robot is red.\n2. Therefore, something is red. [1]",
		"1. Nearly all robots are red.\n2. Therefore, something is red. [1]",
	}
	for _, text := range strictRefused {
		assertStrictRefusal(t, text, "eg.name_witness", "state the existential claim as its own premise")
	}
	for _, text := range []string{
		"1. The server crashed and the network failed.\n2. Therefore, the network failed. [1]",
	} {
		if _, r := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("control refused: %s: %s", text, r.Message)
		}
	}
}

func TestStructuralTermRules(t *testing.T) {
	refused := []string{
		// An unlisted attitude verb inside a term (schema, UI, description EG).
		"1. If X is red, then X is warm.\n2. The observer reckons Alpha is red.\n3. Therefore, The observer reckons Alpha is warm. [1,2]",
		"1. If X is red, then X is warm.\n2. The observer fancies that Alpha is red.\n3. Therefore, The observer fancies that Alpha is warm. [1,2]",
		"1. Every red thing is warm.\n2. The observer reckons Alpha is a red thing.\n3. Therefore, The observer reckons Alpha is warm. [1,2]",
		"1. The observer reckons Alpha is red.\n2. Therefore, something is red. [1]",
		// Raising verbs block identity substitution.
		"1. Superman seems to be strong.\n2. Superman is identical to Clark.\n3. Therefore, Clark seems to be strong. [1,2]",
		"1. Superman appears to fly.\n2. Superman is identical to Clark.\n3. Therefore, Clark appears to fly. [1,2]",
		// Pluralia tantum are not plurals of a singular.
		"1. All tongs are useful.\n2. Alpha is a tong.\n3. Therefore, Alpha is useful. [1,2]",
		"1. All scissors are sharp.\n2. Alpha is a scissor.\n3. Therefore, Alpha is sharp. [1,2]",
		// Postposed if with an unlisted question verb.
		"1. The observer ponders if Alpha is red.\n2. Alpha is red.\n3. Therefore, The observer ponders. [1,2]",
		"1. They ponder if Alpha is red.\n2. Alpha is red.\n3. Therefore, They ponder. [1,2]",
	}
	for _, text := range refused {
		if _, r := CheckEntailmentGate(text); r.Passed {
			t.Errorf("certified: %s", text)
		}
	}
	strictRefused := []struct{ text, rule, hint string }{
		// eg.name_witness: outside the strict term/witness/class grammar.
		{"1. The man who sings is tall.\n2. Therefore, something is tall. [1]", "eg.name_witness", "state the existential claim as its own premise"},
		// slot.value_singular: outside the strict term/witness/class grammar.
		{"1. If X is useful, then X is valuable.\n2. An embedded agent is useful.\n3. Therefore, an embedded agent is valuable. [1,2]", "slot.value_singular", "Name the thing"},
		// slot.value_singular: outside the strict term/witness/class grammar.
		{"1. If X is useful, then X is valuable.\n2. The proportion of tested hypotheses that are true is useful.\n3. Therefore, the proportion of tested hypotheses that are true is valuable. [1,2]", "slot.value_singular", "Name the thing"},
		// term.inflection_verbatim: outside the strict term/witness/class grammar.
		{"1. All premises are valuable.\n2. Alpha is a premise.\n3. Therefore, Alpha is valuable. [1,2]", "term.inflection_verbatim", "singular 'every N'"},
		// identity.extensional_context: outside the strict term/witness/class grammar.
		{"1. Superman can fly.\n2. Superman is identical to Clark.\n3. Therefore, Clark can fly. [1,2]", "identity.extensional_context", "state the substituted claim as a premise"},
	}
	for _, tc := range strictRefused {
		assertStrictRefusal(t, tc.text, tc.rule, tc.hint)
	}
	for _, text := range []string{
		"1. The ground is wet if it rains.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]",
	} {
		if _, r := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("control refused: %s: %s", text, r.Message)
		}
	}
}

// v38.0: the remaining blocklist-audit rows are closed classes or structural
// rules. Each refused case was VALID on v37.0.
func TestAuditPart2Rules(t *testing.T) {
	refused := map[string]string{
		"hesitant if":        "1. Bob is hesitant if Alpha is red.\n2. Alpha is red.\n3. Therefore, Bob is hesitant. [1,2]\n",
		"minority schematic": "1. If X is red, then X is warm.\n2. A minority of robots is red.\n3. Therefore, A minority of robots is warm. [1,2]\n",
		"minority UI":        "1. Every red thing is warm.\n2. A minority of robots is a red thing.\n3. Therefore, A minority of robots is warm. [1,2]\n",
		"zero schematic":     "1. If X is red, then X is warm.\n2. Zero robots is red.\n3. Therefore, Zero robots is warm. [1,2]\n",
		"unless one":         "1. A battery powers the alarm unless one is dead.\n2. Therefore, if it is not the case that one is dead, then a battery powers the alarm. [1]\n",
		"unless this":        "1. The alarm sounds unless this is dead.\n2. Therefore, if it is not the case that this is dead, then the alarm sounds. [1]\n",
		"object UI awaits":   "1. Lois awaits every guest.\n2. Bruce is a guest.\n3. Therefore, Lois awaits Bruce. [1,2]\n",
		"object UI craves":   "1. Lois craves every pastry.\n2. Bruce is a pastry.\n3. Therefore, Lois craves Bruce. [1,2]\n",
		"object EG seeks":    "1. Lois seeks Bruce.\n2. Bruce is a unicorn.\n3. Therefore, Lois seeks some unicorn. [1,2]\n",
		"cf strengthening":   "1. If Bob had struck the match, then it would have lit.\n2. Therefore, if Bob had struck the match and the match had been wet, then it would have lit. [1]\n",
		"cf contraposition":  "1. If Bob had been rich, then Bob would have been happy.\n2. Therefore, if Bob had not been happy, then Bob would not have been rich. [1]\n",
		"cf chaining":        "1. If Bob had been rich, then Bob would have bought a boat.\n2. If Bob had bought a boat, then Bob would have drowned.\n3. Therefore, if Bob had been rich, then Bob would have drowned. [1,2]\n",
		"outer necessarily":  "1. Necessarily the server crashed and the network failed.\n2. Therefore, the network failed. [1]\n",
		"outer possibly":     "1. Possibly the server crashed and the network failed.\n2. Therefore, the network failed. [1]\n",
		"grouping in belief": "1. Ann believes that Bob dances or Cal reads and Dan sleeps.\n2. Therefore, Dan sleeps. [1]\n",
		// v43.0: an article-led value with a complement need not be one
		// thing (slot.indefinite_value); certified before v43.0.
		"majority schematic": "1. If X is red, then X is warm.\n2. A majority of robots is red.\n3. Therefore, a majority of robots is warm. [1,2]\n",
	}
	for name, text := range refused {
		if report, _ := CheckEntailmentGate(text); report.Passed {
			t.Errorf("%s: want NOT CERTIFIED", name)
		}
	}
	// slot.value_singular: bare and partitive values do not name a singular term.
	strictRefused := map[string]string{
		"value term":     "1. If X is a small number, then X is finite.\n2. Zero probability is a small number.\n3. Therefore, Zero probability is finite. [1,2]\n",
		"the proportion": "1. If X is a prior, then X is subjective.\n2. The proportion of true hypotheses is a prior.\n3. Therefore, The proportion of true hypotheses is subjective. [1,2]\n",
	}
	for name, text := range strictRefused {
		t.Run(name, func(t *testing.T) { assertStrictRefusal(t, text, "slot.value_singular", "Name the thing") })
	}
	valid := map[string]string{
		"unless plain": "1. The alarm sounds unless the battery is dead.\n2. Therefore, if it is not the case that the battery is dead, then the alarm sounds. [1]\n",
		"subject UI":   "1. Every courier sends Parcel to Bob.\n2. Alice is a courier.\n3. Therefore, Alice sends Parcel to Bob. [1,2]\n",
	}
	for name, text := range valid {
		if report, _ := CheckEntailmentGate(text); !report.Passed {
			t.Errorf("%s: want VALID: %+v", name, report.Steps)
		}
	}
}

// v38.1: "admires" is not extensional in object position (Lois admires
// Superman, not Clark), so object instantiation through it is refused.
func TestAdmiresNotExtensional(t *testing.T) {
	for _, text := range []string{
		"1. Lois admires every hero.\n2. Bruce is a hero.\n3. Therefore, Lois admires Bruce. [1,2]\n",
		"1. Lois admires Bruce.\n2. Bruce is a hero.\n3. Therefore, Lois admires some hero. [1,2]\n",
	} {
		if report, _ := CheckEntailmentGate(text); report.Passed {
			t.Errorf("want NOT CERTIFIED: %q", text)
		}
	}
	if report, _ := CheckEntailmentGate("1. Lois greets every hero.\n2. Bruce is a hero.\n3. Therefore, Lois greets Bruce. [1,2]\n"); !report.Passed {
		t.Errorf("greets control: want VALID")
	}
}

// v41.0: "Both A and B", the form the clause-coordination hint recommends,
// gives conjunction introduction and elimination.
func TestBothConjunction(t *testing.T) {
	valid := []string{
		"1. Mary had all the physical information.\n2. Mary's previous knowledge was incomplete.\n3. Therefore, both Mary had all the physical information and Mary's previous knowledge was incomplete. [1,2]\n",
		"1. Both Ann sings and Bob dances.\n2. Therefore, Bob dances. [1]\n",
	}
	for _, text := range valid {
		if report, _ := CheckEntailmentGate(text); !report.Passed {
			t.Errorf("want VALID: %q %+v", text, report.Steps)
		}
	}
	refused := []string{
		"1. Both Ann sings and Bob dances or Cal reads.\n2. Therefore, Bob dances. [1]\n",
		"1. Ann believes both Bob sings and Cal dances.\n2. Therefore, Cal dances. [1]\n",
	}
	for _, text := range refused {
		if report, _ := CheckEntailmentGate(text); report.Passed {
			t.Errorf("want NOT CERTIFIED: %q", text)
		}
	}
}

// task_eb737ad9: four false positives found by the AI-safety map inventory.
// Each relaxation is limited to readings that are not ambiguous.
func TestInventoryFalsePositivesV42(t *testing.T) {
	strictRefused := []struct{ text, rule, hint string }{
		// slot.value_singular: outside the strict term/witness/class grammar.
		{"1. If X is natural for large networks, then X is common.\n2. Subassembly organization is natural for large networks.\n3. Therefore, subassembly organization is common. [1,2]", "slot.value_singular", "Name the thing"},
		// slot.value_singular: outside the strict term/witness/class grammar.
		{"1. If X depends on factual tests, then X needs data.\n2. Friendly action depends on factual tests.\n3. Therefore, Friendly action needs data. [1,2]", "slot.value_singular", "Name the thing"},
		// slot.value_singular: outside the strict term/witness/class grammar.
		{"1. Rex runs fast even when machines are noisy.\n2. If X runs fast even when U, then X wins even when U.\n3. Therefore, Rex wins even when machines are noisy. [1,2]", "slot.value_singular", "Name the thing"},
		// slot.value_singular: outside the strict term/witness/class grammar.
		{"1. If considering X invites caution, then X is risky.\n2. Considering brain emulation invites caution.\n3. Therefore, brain emulation is risky. [1,2]", "slot.value_singular", "Name the thing"},
	}
	for _, tc := range strictRefused {
		assertStrictRefusal(t, tc.text, tc.rule, tc.hint)
	}
	certified := []string{
		// -ly nouns and -ly adjectives opening a noun-phrase subject.
		// A whole clause as the value of a clause-slot variable.
		"1. The plan fails because costs are high.\n2. If the plan fails because U, then U.\n3. Therefore, costs are high. [1,2]",
		// A gerund subject: no comma, no other subject, a finite verb follows.
	}
	for _, text := range certified {
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Fatalf("should be certified (%s):\n%s", result.Message, text)
		}
	}
	refused := []string{
		// Adverb-capable -ly words and comma-separated -ly words stay openers.
		"1. If X arrives, then X wins.\n2. Likely the team arrives.\n3. Therefore, likely the team wins. [1,2]",
		"1. If X is designed, then X is safe.\n2. Friendly, the robot is designed.\n3. Therefore, the robot is safe. [1,2]",
		// A participial frame (comma, or its own subject) stays an opener.
		"1. If considering X, the team stops, then X is risky.\n2. Considering the cost, the team stops.\n3. Therefore, the cost is risky. [1,2]",
		"1. If considering X we stop, then X is risky.\n2. Considering the cost we stop.\n3. Therefore, the cost is risky. [1,2]",
	}
	for _, text := range refused {
		if _, result := CheckEntailmentGate(text); result.Passed {
			t.Fatalf("should stay refused:\n%s", text)
		}
	}
	// A term that reads as containing a clause stays refused (a noun/verb
	// homonym may be ambiguous), but the message names the binding.
	homonym := "1. If X are available threat channels, then X are dangerous.\n2. Physics exploits not known to humans are available threat channels.\n3. Therefore, physics exploits not known to humans are dangerous. [1,2]"
	report, result := CheckEntailmentGate(homonym)
	if result.Passed {
		t.Fatal("homonym term should stay refused")
	}
	message := ""
	for _, step := range report.Steps {
		if !step.Passed {
			message = step.Mismatch
		}
	}
	if !strings.Contains(message, `"physics exploits not known to humans"`) || !strings.Contains(message, "reads as containing a clause") {
		t.Fatalf("message should name the refused term, got %q", message)
	}
	// A that-clause slot is not treated as clause-only (that can open a term).
	if clauseSlotBefore([]rune("he believes that "), len([]rune("he believes that "))) {
		t.Fatal("that is not a clause-only subordinator")
	}
}
