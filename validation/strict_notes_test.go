package validation

import (
	"reflect"
	"testing"
)

func TestStrictNotesExamples(t *testing.T) {
	fixes := map[string]string{
		"opener.frame":                "1. If X is registered in 2019, then X is eligible in 2019.\n2. The firm is registered in 2019.\n3. Therefore, the firm is eligible in 2019. [1,2]",
		"opener.adverb":               "1. If X is safe, then X is useful.\n2. Machinery that is highly reliable is safe.\n3. Therefore, machinery that is highly reliable is useful. [1,2]",
		"coordination.clause":         "1. The pump malfunctioned.\n2. The valve stuck.\n3. Therefore, both the pump malfunctioned and the valve stuck. [1,2]",
		"comma.exact":                 "1. Every device is small, light and portable.\n2. Alpha is a device.\n3. Therefore, Alpha is small, light and portable. [1,2]",
		"identity.safe_position":      "1. The robot kicks Beta.\n2. Therefore, the robot kicks Beta. [1]",
		"slot.indefinite_object":      "1. If the robot admires X, then X is admirable.\n2. The robot admires Alpha.\n3. Therefore, Alpha is admirable. [1,2]",
		"slot.modified_quantifier":    "1. If X is red, then X is warm.\n2. Alpha is red.\n3. Therefore, Alpha is warm. [1,2]",
		"slot.quantity":               "1. If X is registered, then X is eligible.\n2. The team is registered.\n3. Therefore, the team is eligible. [1,2]",
		"term.clause_inside":          "1. If X are useful, then X are valuable.\n2. The exploits are useful.\n3. Therefore, the exploits are valuable. [1,2]",
		"noun.pluralia_tantum":        "1. Every glass is fragile.\n2. The cup is a glass.\n3. Therefore, the cup is fragile. [1,2]",
		"conditional.postposed_if":    "1. If the motor runs, then the machine vibrates.\n2. The motor runs.\n3. Therefore, the machine vibrates. [1,2]",
		"relation.object_extensional": "1. The robot admires Alpha.\n2. Therefore, the robot admires Alpha. [1]",
		"unless.anaphor":              "1. The alarm sounds unless the battery is dead.\n2. It is not the case that the battery is dead.\n3. Therefore, the alarm sounds. [1,2]",
	}
	for _, note := range strictNotes {
		t.Run(note.ID, func(t *testing.T) {
			for _, example := range note.AlsoRefuses {
				report, result := CheckEntailmentGate(example, notesOn)
				if report.Passed || result.Passed {
					t.Fatalf("expected refusal: %+v", report)
				}
				found := false
				for _, step := range report.Steps {
					for _, got := range step.Notes {
						if got.ID == note.ID {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("missing %s: %+v", note.ID, report.Steps)
				}
				argument, parsed := ParseNumberedArgument(example, numberedArgumentOptions{})
				if !parsed.Passed {
					t.Fatal(parsed.Message)
				}
				baseline, baselineResult := allStrictRules().gateParsedArgument(protectArgumentNameLetters(argument), false)
				for i := range report.Steps {
					report.Steps[i].Notes = nil
				}
				if result.Passed != baselineResult.Passed || result.Message != baselineResult.Message || !reflect.DeepEqual(report, baseline) {
					t.Errorf("diagnostics changed baseline: %+v / %+v", report, baseline)
				}
			}
			report, result := CheckEntailmentGate(fixes[note.ID], notesOn)
			if !report.Passed {
				t.Errorf("fix fails: %+v / %s", report.Steps, result.Message)
			}
			for _, step := range report.Steps {
				if len(step.Notes) > 0 {
					t.Error("passing fix has notes")
				}
			}
		})
	}
}

var notesOn = EntailmentGateOptions{StrictNotes: true}

// Notes are opt-in: the default check never re-runs failing steps, so its cost
// and output are those of v42.0.
func TestStrictNotesOffByDefault(t *testing.T) {
	for _, note := range strictNotes {
		report, _ := CheckEntailmentGate(note.AlsoRefuses[0])
		for _, step := range report.Steps {
			if len(step.Notes) > 0 {
				t.Fatalf("%s: notes attached without StrictNotes", note.ID)
			}
		}
	}
}
