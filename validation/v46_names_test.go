package validation

import (
	"strings"
	"testing"
)

func TestV46NameSlots(t *testing.T) {
	for _, tc := range []struct {
		name string
		pass bool
	}{
		{"Leah", true}, {"Patel", true}, {"Lots", false}, {"Plenty", false},
		{"Most", false}, {"Banks", false}, {"Zorvan", false}, {"leah", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := "1. " + tc.name + " is a nurse.\n2. If X is a nurse, then X is trained.\n3. Therefore, " + tc.name + " is trained. [1,2]"
			r, res := CheckEntailmentGate(text)
			if r.Passed != tc.pass {
				t.Fatalf("pass=%v want=%v: %s", r.Passed, tc.pass, res.Message)
			}
		})
	}
}

func TestV46ContextNameEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, extra string
		pass        bool
	}{
		{"Lots", "If Lots arrived, then the bell rings.", false},
		{"Plenty", "If Plenty arrived, then the bell rings.", false},
		{"Dozens", "If Dozens arrived, then the bell rings.", false},
		{"Zorvan", "If Zorvan arrived, then the bell rings.", true},
		{"Zorvan", "If Zorvan arrived, then zorvan leaves.", false},
		{"Zorvan", "The heading is Zorvan.", false},
		{"Zorvan", "Bob said Zorvan, the bell rings.", false},
		{"Zorvan", "Bob admires Zorvan.", true},
	} {
		t.Run(tc.name+tc.extra, func(t *testing.T) {
			text := "1. If X arrived, then X was seen.\n2. " + tc.name + " arrived.\n3. " + tc.extra + "\n4. Therefore, " + tc.name + " was seen. [1,2]"
			r, res := CheckEntailmentGate(text)
			if r.Passed != tc.pass {
				t.Fatalf("pass=%v want=%v: %s", r.Passed, tc.pass, res.Message)
			}
			steps := map[int]NumberedArgumentStep{1: {Statement: tc.name + " arrived."}, 2: {Statement: tc.extra}}
			if got := (*schematicNameEvidence(steps))[strings.ToLower(tc.name)]; got != tc.pass {
				t.Errorf("evidence=%v want=%v", got, tc.pass)
			}
		})
	}
	// Veto every generated quantity/closed form even when planted in a term position.
	for _, name := range strings.Fields("Lots Plenty Most Many Few Several Dozens Some All Half None") {
		steps := map[int]NumberedArgumentStep{1: {Statement: "If " + name + " arrived, then the bell rings."}}
		if (*schematicNameEvidence(steps))[strings.ToLower(name)] {
			t.Errorf("quantity evidence: %s", name)
		}
	}
	for _, name := range []string{"AIXI", "McCarthy", "Leah's", "Leah’s"} {
		steps := map[int]NumberedArgumentStep{1: {Statement: name + " dog arrived."}}
		base := strings.TrimSuffix(strings.TrimSuffix(strings.ToLower(name), "'s"), "’s")
		if !(*schematicNameEvidence(steps))[base] {
			t.Errorf("missing evidence: %s", name)
		}
	}
}

func TestV46DegreeOpeners(t *testing.T) {
	for _, tc := range []struct {
		subject string
		pass    bool
	}{
		{"Purely deferential moral philosophy", true}, {"Entirely deferential moral philosophy", true},
		{"Purely, deferential moral philosophy", false}, {"Only deferential moral philosophy", false},
		{"Likely deferential moral philosophy", false}, {"Probably deferential moral philosophy", false},
		{"Arguably deferential moral philosophy", false}, {"Purely philosophy", false},
	} {
		t.Run(tc.subject, func(t *testing.T) {
			statement := tc.subject + " is unstable"
			if got := allStrictRules().sentenceFrameOpener(statement, nil) == ""; got != tc.pass {
				t.Errorf("opener pass=%v want=%v", got, tc.pass)
			}
			text := "1. Both " + statement + " and the bell rings.\n2. Therefore, " + statement + ". [1]"
			r, res := CheckEntailmentGate(text)
			if r.Passed != tc.pass {
				t.Fatalf("pass=%v want=%v: %s", r.Passed, tc.pass, res.Message)
			}
		})
	}
	for _, word := range strings.Fields("Only Likely Probably Arguably Possibly Presumably Supposedly Apparently Clearly Obviously Certainly Surely Really Actually Generally Usually Typically Ultimately Finally Similarly Consequently Accordingly Importantly") {
		if frameDegreeAdjectiveSubject([]string{word, "capable", "nurses"}) {
			t.Errorf("hedge accepted: %s", word)
		}
	}
}
