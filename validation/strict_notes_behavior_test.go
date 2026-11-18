package validation

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func TestStrictNotesPassingAndCombinedRefusals(t *testing.T) {
	for _, example := range []string{
		"1. The lamp is on.\n2. Therefore, the lamp is on. [1]",
		// Both the fronted-frame and clause-coordination guards block this step.
		"1. In 2019 the firm cut prices and the firm lost money.\n2. Therefore, the firm lost money. [1]",
	} {
		report, _ := CheckEntailmentGate(example, notesOn)
		for _, step := range report.Steps {
			if len(step.Notes) != 0 {
				t.Fatalf("unexpected notes: %+v", step)
			}
		}
		data, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), `"notes"`) {
			t.Fatal("empty notes must be omitted")
		}
	}
}

func TestStrictNotesConcurrentAndOwned(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, note := range strictNotes {
				report, _ := CheckEntailmentGate(note.AlsoRefuses[0], notesOn)
				if report.Passed {
					t.Errorf("diagnostic leaked into verdict for %s", note.ID)
					continue
				}
				found := false
				for _, step := range report.Steps {
					for _, got := range step.Notes {
						if got.ID == note.ID {
							found = true
							got.Fix[0] = "client edit"
							got.AlsoRefuses[0] = "client edit"
						}
					}
				}
				if !found {
					t.Errorf("missing concurrent note %s", note.ID)
				}
			}
		}()
	}
	wg.Wait()
	for _, note := range strictNotes {
		if note.Fix[0] == "client edit" || note.AlsoRefuses[0] == "client edit" {
			t.Error("client mutated catalog")
		}
	}
}

func TestStrictNotesJSONAndLibrary(t *testing.T) {
	example := strictNotes[0].AlsoRefuses[0]
	report, result := CheckEntailmentGate(example, notesOn)
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded EntailmentGateReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Steps) != 1 || len(decoded.Steps[0].Notes) != 1 || decoded.Steps[0].Notes[0].ID != "opener.frame" {
		t.Fatalf("missing structured note: %s", data)
	}
	details, err := json.Marshal(result.Details)
	if err != nil || !strings.Contains(string(details), `"notes"`) {
		t.Fatalf("library result details omit notes: %s / %v", details, err)
	}
	argument, parsed := ParseNumberedArgument(example, numberedArgumentOptions{})
	if !parsed.Passed {
		t.Fatal(parsed.Message)
	}
	withNotes := allStrictRules()
	withNotes.notes = true
	direct := withNotes.GateNumberedArgument(argument)
	if len(direct.Steps[0].Notes) != 1 {
		t.Fatal("direct gate omits notes")
	}
	if plain := GateNumberedArgument(argument); len(plain.Steps[0].Notes) != 0 {
		t.Fatal("default direct gate attached notes")
	}
}
