package validation

import (
	"strings"
	"testing"
)

func TestSentenceFrames(t *testing.T) {
	cases := []struct {
		name, text, message string
		valid               bool
	}{
		{"plain", "1. The server crashed and the network failed.\n2. Therefore, the network failed. [1]", "", true},
		{"both", "1. Both the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "", true},
		{"reckons", "1. Bob reckons the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "write 'Both", false},
		{"fancies", "1. Bob fancies the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "write 'Both", false},
		{"insinuated", "1. Bob insinuated the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "write 'Both", false},
		{"owner", "1. The owner of the robot believes alpha and beta.\n2. Therefore, The owner of the robot believes alpha. [1]", "write 'Both", false},
		{"introduction", "1. Bob reckons the server crashed.\n2. The network failed.\n3. Therefore, Bob reckons the server crashed and the network failed. [1,2]", "write 'Both", false},
		{"embedded-parentheses", "1. Bob reckons (the server crashed and the network failed).\n2. Therefore, the network failed. [1]", "", false},
		{"outer-parentheses", "1. (Bob reckons the server crashed) and (the network failed).\n2. Therefore, the network failed. [1]", "", true},
		{"restate-parentheses", "1. Bob reckons (the server crashed and the network failed).\n2. Therefore, Bob reckons (the server crashed and the network failed). [1]", "", true},
		{"unsafe-universal", "1. Every man is mortal.\n2. Ostensibly Socrates is a man.\n3. Therefore, Ostensibly Socrates is mortal. [1,2]", "may cover the whole sentence", false},
		{"unsafe-schematic", "1. If X is good, then X is valuable.\n2. Purportedly the server is good.\n3. Therefore, Purportedly the server is valuable. [1,2]", "may cover the whole sentence", false},
		{"opaque-reiteration", "1. Hypothetically the server crashed.\n2. Therefore, Hypothetically the server crashed. [1]", "", true},
		{"universal", "1. Every man is mortal.\n2. Socrates is a man.\n3. Therefore, Socrates is mortal. [1,2]", "", true},
		{"gerund", "1. If X is good, then X is valuable.\n2. Trying to remove all assumptions is good.\n3. Therefore, Trying to remove all assumptions is valuable. [1,2]", "", true},
		{"bare-noun", "1. Consciousness is real and the server crashed.\n2. Therefore, the server crashed. [1]", "", true},
		{"name", "1. Plan B is good and the server crashed.\n2. Therefore, the server crashed. [1]", "", true},
		{"panda", "1. Every panda eats shoots and leaves.\n2. Bao is a panda.\n3. Therefore, Bao eats, shoots, and leaves. [1,2]", "differs only in punctuation", false},
		{"relative", "1. Every philosopher who is wise is happy.\n2. Therefore, Every philosopher, who is wise, is happy. [1]", "differs only in punctuation", false},
	}
	for _, opener := range []string{"Arguably", "Purportedly", "Ostensibly", "Hypothetically", "Seemingly", "In the story,"} {
		cases = append(cases, struct {
			name, text, message string
			valid               bool
		}{opener, "1. " + opener + " the server crashed and the network failed.\n2. Therefore, the network failed. [1]", "may cover the whole sentence", false})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
			if len(r.Steps) == 0 || !strings.Contains(r.Steps[0].Mismatch, "write 'Both") {
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
	for _, text := range []string{
		"1. The server crashed and the network failed.\n2. Therefore, the network failed. [1]",
		"1. Almost every robot is red.\n2. Therefore, something is red. [1]",
		"1. Nearly all robots are red.\n2. Therefore, something is red. [1]",
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
	for _, text := range []string{
		"1. The man who sings is tall.\n2. Therefore, something is tall. [1]",
		"1. If X is useful, then X is valuable.\n2. An embedded agent is useful.\n3. Therefore, an embedded agent is valuable. [1,2]",
		"1. If X is useful, then X is valuable.\n2. The proportion of tested hypotheses that are true is useful.\n3. Therefore, the proportion of tested hypotheses that are true is valuable. [1,2]",
		"1. All premises are valuable.\n2. Alpha is a premise.\n3. Therefore, Alpha is valuable. [1,2]",
		"1. The ground is wet if it rains.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]",
		"1. Superman can fly.\n2. Superman is identical to Clark.\n3. Therefore, Clark can fly. [1,2]",
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
	}
	for name, text := range refused {
		if report, _ := CheckEntailmentGate(text); report.Passed {
			t.Errorf("%s: want NOT CERTIFIED", name)
		}
	}
	valid := map[string]string{
		"majority schematic": "1. If X is red, then X is warm.\n2. A majority of robots is red.\n3. Therefore, a majority of robots is warm. [1,2]\n",
		"value term":         "1. If X is a small number, then X is finite.\n2. Zero probability is a small number.\n3. Therefore, Zero probability is finite. [1,2]\n",
		"the proportion":     "1. If X is a prior, then X is subjective.\n2. The proportion of true hypotheses is a prior.\n3. Therefore, The proportion of true hypotheses is subjective. [1,2]\n",
		"unless plain":       "1. The alarm sounds unless the battery is dead.\n2. Therefore, if it is not the case that the battery is dead, then the alarm sounds. [1]\n",
		"subject UI":         "1. Every courier sends Parcel to Bob.\n2. Alice is a courier.\n3. Therefore, Alice sends Parcel to Bob. [1,2]\n",
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
