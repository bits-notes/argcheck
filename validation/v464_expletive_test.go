package validation

import (
	"fmt"
	"strings"
	"testing"
)

func v464MP(clause string) string {
	return fmt.Sprintf("1. If %s, then the ground is wet.\n2. %s.\n3. Therefore, the ground is wet. [1,2]", clause, clause)
}

func TestV464ExpletiveAllowlist(t *testing.T) {
	var frames []string
	for _, adjective := range []string{"true", "false", "the case", "possible", "necessary", "likely", "unlikely", "certain", "clear", "obvious"} {
		for _, neg := range []string{"", "not "} {
			frames = append(frames, "it is "+neg+adjective+" that (the pump is broken)")
		}
	}
	for _, verb := range []string{"seems", "appears", "follows"} {
		frames = append(frames, "it "+verb+" that (the pump is broken)")
	}
	frames = append(frames, "it's true that (the pump is broken)", "It’s true that (the pump is broken)", "it is the case both that Alice sings, and that Bob sings")
	for _, verb := range []string{"rains", "snows", "hails", "drizzles", "pours", "sleets", "thunders", "is raining", "is snowing", "is hailing", "is drizzling", "is pouring", "is sleeting", "is thundering"} {
		frames = append(frames, "it "+verb)
		for _, suffix := range []string{"", ".", ",", ", then the bell rings", " then the bell rings", " and the bell rings", " or the bell rings", ")"} {
			if !crossStatementScaffold("it " + verb + suffix) {
				t.Errorf("missing weather boundary: %q", verb+suffix)
			}
		}
		for _, suffix := range []string{" water", " the tea", "water", "é", "_water", ". water", " andré"} {
			if crossStatementScaffold("it " + verb + suffix) {
				t.Errorf("exempted object/partial token: %q", verb+suffix)
			}
		}
	}
	for _, frame := range frames {
		t.Run(frame, func(t *testing.T) {
			if _, changed := isolateStatementPronouns(frame, "Internal", 1, crossStatementPronoun); changed {
				t.Fatal("isolated dummy it")
			}
			if allStrictRules().crossStatementBindingRefused(frame) {
				t.Fatal("refused dummy-it binding")
			}
			if r, res := CheckEntailmentGate(v464MP(frame)); !r.Passed {
				t.Fatal(res.Message)
			}
		})
	}
	for _, frame := range []string{"it is true", "it is clear", "it is wet", "it is cold", "it rains water", "it pours the tea", "it is true thatcher", "it seems thatcher", "it is true that it is valid"} {
		if _, changed := isolateStatementPronouns(frame, "Internal", 1, crossStatementPronoun); !changed {
			t.Errorf("did not isolate %q", frame)
		}
		if !allStrictRules().crossStatementBindingRefused(frame) {
			t.Errorf("did not refuse binding %q", frame)
		}
	}
	got, _ := isolateStatementPronouns("It is true that it is valid", "Internal", 1, crossStatementPronoun)
	if got != "It is true that Internalbit is valid" {
		t.Fatalf("wrong token exempted: %s", got)
	}
}

func TestV464ExpletiveRefusals(t *testing.T) {
	nested := strings.ReplaceAll(v462PronounProbes["D2-4"].text, "it is valid", "it is true that it is valid")
	nested = strings.ReplaceAll(nested, "It is valid", "It is true that it is valid")
	for _, text := range []string{nested, v462PronounProbes["weather"].text, "1. The sprinkler is on.\n2. If it rains water, then the lawn is wet.\n3. It rains water.\n4. Therefore, the lawn is wet. [2,3]", v464MP("It is true"), v464MP("it pours the tea")} {
		r, _ := CheckEntailmentGate(text, notesOn)
		if r.Passed {
			t.Fatalf("certified: %s", text)
		}
		found := false
		for _, step := range r.Steps {
			for _, note := range step.Notes {
				found = found || note.ID == "term.pronoun_cross_statement"
			}
		}
		if !found {
			t.Fatalf("missing pronoun note: %+v", r.Steps)
		}
	}
}

func TestV464ExpletiveProbeFiles(t *testing.T) {
	probes := map[string]string{
		"rain":  "1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]",
		"true":  "1. If it is true that (the lamp is lit), then the door is open.\n2. It is true that (the lamp is lit).\n3. Therefore, the door is open. [1,2]",
		"mt":    "1. If the lamp is lit, then the door is open.\n2. It is not the case that the door is open.\n3. Therefore, it is not the case that the lamp is lit. [1,2]",
		"expl1": "1. If it is true that (the pump is broken), then the valve is open.\n2. It is true that (the pump is broken).\n3. Therefore, the valve is open. [1,2]",
	}
	for name, text := range probes {
		t.Run(name, func(t *testing.T) {
			if r, res := CheckEntailmentGate(text); !r.Passed {
				t.Fatal(res.Message)
			}
		})
	}
	// This original v46.2 modal fixture now contains only allowlisted dummy it.
	text := "1. It is not the case that possibly, it is not the case that it rains.\n2. Therefore, necessarily, it rains. [1]"
	if r, res := CheckEntailmentGate(text); !r.Passed {
		t.Fatal(res.Message)
	}
}
