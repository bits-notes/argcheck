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
