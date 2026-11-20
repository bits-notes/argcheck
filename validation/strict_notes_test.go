package validation

import (
	"reflect"
	"testing"
)

func TestStrictNotesExamples(t *testing.T) {
	fixes := map[string]string{
		"term.pronoun_cross_statement":    "1. If the card is valid, then the alarm sounds.\n2. The card is valid.\n3. Therefore, the alarm sounds. [1,2]",
		"term.possessive_cross_statement": "1. If John helps Mary, then the alarm sounds.\n2. John helps Mary.\n3. Therefore, the alarm sounds. [1,2]",
		"slot.plural_witness":             "1. If the water is in the cup, then the water is in both vessels.\n2. The water is in the cup.\n3. Therefore, the water is in both vessels. [1,2]",
		"slot.description_shape":          "1. If X is red, then X is warm.\n2. The dog is red.\n3. Therefore, the dog is warm. [1,2]",
		"slot.one_type":                   "1. The dog is red.\n2. Therefore, the dog is red. [1]",
		"eg.subject_only":                 "1. There is an X such that Lois greets X.\n2. Therefore, there is an X such that Lois greets X. [1]",
		"eg.name_witness":                 "1. Something is red.\n2. Therefore, something is red. [1]",
		"term.numeral":                    "1. Every rodent is warm.\n2. The mouse is a rodent.\n3. Therefore, the mouse is warm. [1,2]",
		"identity.constant_sides":         "1. The dog is identical to Alpha.\n2. The dog is identical to Beta.\n3. Therefore, Alpha is identical to Beta. [1,2]",
		"identity.extensional_context":    "1. The man is identical to Clark.\n2. The man is red.\n3. Therefore, Clark is red. [1,2]",
		"slot.letter_in_name":             "1. The dog is red.\n2. Therefore, the dog is red. [1]",
		"term.inflection_verbatim":        "1. Every dog is wet.\n2. The dog is a dog.\n3. Therefore, the dog is wet. [1,2]",
		"slot.value_singular":             "1. If X arrived, then X was seen.\n2. The dog arrived.\n3. Therefore, the dog was seen. [1,2]",
		"opener.frame":                    "1. If X is registered in 2019, then X is eligible in 2019.\n2. The firm is registered in 2019.\n3. Therefore, the firm is eligible in 2019. [1,2]",
		"opener.adverb":                   "1. If X is safe, then X is useful.\n2. The reliable machine is safe.\n3. Therefore, the reliable machine is useful. [1,2]",
		"coordination.clause":             "1. The pump malfunctioned.\n2. The valve stuck.\n3. Therefore, both the pump malfunctioned and the valve stuck. [1,2]",
		"comma.exact":                     "1. Every device is small, light and portable.\n2. Alpha is a device.\n3. Therefore, Alpha is small, light and portable. [1,2]",
		"identity.safe_position":          "1. The robot kicks Beta.\n2. Therefore, the robot kicks Beta. [1]",
		"slot.indefinite_object":          "1. If the robot admires X, then X is admirable.\n2. The robot admires Alpha.\n3. Therefore, Alpha is admirable. [1,2]",
		"slot.indefinite_written":         "1. If Rex is a dog, then Fido is a dog.\n2. Rex is a dog.\n3. Therefore, Fido is a dog. [1,2]",
		"slot.indefinite_value":           "1. If a lot of people arrived, then a lot of people is someone who arrived.\n2. A lot of people arrived.\n3. Therefore, a lot of people is someone who arrived. [1,2]",
		"slot.modified_quantifier":        "1. If X is red, then X is warm.\n2. The robot is red.\n3. Therefore, the robot is warm. [1,2]",
		"slot.quantity":                   "1. If X is registered, then X is eligible.\n2. The team is registered.\n3. Therefore, the team is eligible. [1,2]",
		"term.clause_inside":              "1. If X is useful, then X is valuable.\n2. The exploit is useful.\n3. Therefore, the exploit is valuable. [1,2]",
		"noun.pluralia_tantum":            "1. Every glass is fragile.\n2. The cup is a glass.\n3. Therefore, the cup is fragile. [1,2]",
		"conditional.postposed_if":        "1. If the motor runs, then the machine vibrates.\n2. The motor runs.\n3. Therefore, the machine vibrates. [1,2]",
		"relation.object_extensional":     "1. The robot admires Alpha.\n2. Therefore, the robot admires Alpha. [1]",
		"unless.anaphor":                  "1. The alarm sounds unless the battery is dead.\n2. It is not the case that the battery is dead.\n3. Therefore, the alarm sounds. [1,2]",
		"scope.mixed_andor":               "1. Both Ann sings and Bob dances.\n2. Therefore, Ann sings. [1]",
		"scope.clause_start":              "1. The lamp is lit and the door is open.\n2. Therefore, the lamp is lit. [1]",
		"opener.ing_frame":                "1. Both Alice is happy and Bob is sad.\n2. Therefore, Bob is sad. [1]",
		"scope.elided_copula":             "1. Either the men are quiet or the men are fish.\n2. Therefore, either the men are quiet or the men are fish. [1]",
		"scope.description_negation":      "1. It is not the case that Louis is bald.\n2. Therefore, Louis is not bald. [1]",
		"opener.ly_exception":             "1. The lamp is lit and the door is open.\n2. Therefore, the door is open and the lamp is lit. [1]",
	}
	for _, note := range strictNotes {
		t.Run(note.ID, func(t *testing.T) {
			// v46.1: bare plural/mass slot values refused (reading shift).
			// The singular-value guard now explains the plural-witness example.
			wantID := note.ID
			if wantID == "slot.plural_witness" {
				wantID = "slot.value_singular"
			}
			for _, example := range note.AlsoRefuses {
				report, result := CheckEntailmentGate(example, notesOn)
				if report.Passed || result.Passed {
					t.Fatalf("expected refusal: %+v", report)
				}
				found := false
				for _, step := range report.Steps {
					for _, got := range step.Notes {
						if got.ID == wantID {
							found = true
						}
					}
				}
				// v46.2: the pronoun example also fails the new cross-statement
				// guard, so disabling identity.constant_sides alone cannot certify it.
				if !found && wantID != "identity.constant_sides" {
					t.Errorf("missing %s: %+v", wantID, report.Steps)
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
		if len(note.AlsoRefuses) == 0 {
			continue
		}
		report, _ := CheckEntailmentGate(note.AlsoRefuses[0])
		for _, step := range report.Steps {
			if len(step.Notes) > 0 {
				t.Fatalf("%s: notes attached without StrictNotes", note.ID)
			}
		}
	}
}
