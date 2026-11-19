package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestV31Fixtures(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestV31Fixtures/schematic-might.txt": {"scope.clause_start", "move the operator into each clause"},
	}

	files, err := filepath.Glob("testdata/v31/*.txt")
	if err != nil || len(files) != 6 {
		t.Fatalf("fixtures: count=%d err=%v", len(files), err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if listedStrictRefusal(t, string(data), refused) {
				return
			}
			_, result := CheckEntailmentGate(string(data))
			if !result.Passed {
				t.Fatal(result.Message)
			}
		})
	}
}

func TestV31OpaqueModalAtoms(t *testing.T) {
	for _, modal := range []string{"might", "could", "may", "can"} {
		atom := "the system " + modal + " seek power"
		for _, tc := range []struct {
			name       string
			premises   []string
			conclusion string
			valid      bool
		}{
			{"MP", []string{"If the system is misaligned, then " + atom, "The system is misaligned"}, atom, true},
			{"MT", []string{"If the system is misaligned, then " + atom, "It is not the case that " + atom}, "It is not the case that the system is misaligned", true},
			{"antecedent", []string{"If " + atom + ", then the alarm sounds", atom}, "The alarm sounds", true},
			{"conjunctive MP", []string{"If the system is misaligned and the alarm sounds, then " + atom, "The system is misaligned", "The alarm sounds"}, atom, true},
			{"schematic", []string{"If X is misaligned, then X " + modal + " seek power", "The system is misaligned"}, atom, true},
			{"no converse", []string{"If the system is misaligned, then " + atom, atom}, "The system is misaligned", false},
			{"no elimination", []string{atom}, "The system seeks power", false},
			{"no introduction", []string{"The system seeks power"}, atom, false},
			{"no necessity", []string{atom}, "Necessarily, the system seeks power", false},
		} {
			t.Run(modal+"/"+tc.name, func(t *testing.T) {
				_, result := CheckEntailmentGate(v26Argument(tc.premises, tc.conclusion))
				if result.Passed != tc.valid {
					t.Fatal(result.Message)
				}
			})
		}
	}
	for _, tc := range []struct {
		name       string
		premises   []string
		conclusion string
		valid      bool
	}{
		{"literal no elimination", []string{"Q might R"}, "Q R", false},
		{"literal no introduction", []string{"Q R"}, "Q might R", false},
		{"literal no converse", []string{"If P, then Q might R", "Q might R"}, "P", false},
		{"no modal substitution", []string{"Q might R"}, "Q could R", false},
		{"perfect counterfactual", []string{"If the match had been struck, then the match might have lit", "The match had been struck"}, "The match might have lit", false},
		{"were counterfactual", []string{"If the system were misaligned, then the system might seek power", "The system were misaligned"}, "The system might seek power", false},
		{"necessity elimination", []string{"Necessarily, the system is misaligned"}, "The system is misaligned", true},
		{"possibility introduction", []string{"The system is misaligned"}, "Possibly, the system is misaligned", true},
		{"no possibility elimination", []string{"Possibly, the system is misaligned"}, "The system is misaligned", false},
		{"no necessity introduction", []string{"The system is misaligned"}, "Necessarily, the system is misaligned", false},
		{"explicit modal on opaque atom", []string{"Necessarily, the system might seek power"}, "The system might seek power", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, result := CheckEntailmentGate(v26Argument(tc.premises, tc.conclusion))
			if result.Passed != tc.valid {
				t.Fatal(result.Message)
			}
		})
	}
}

func TestV31V27LegacyNegationEmptyRestrictor(t *testing.T) {
	text := v26Argument([]string{"It is not the case that the present king of France is bald"}, "The present king of France is not bald")
	v27Parity(t, text)
	refused := []string{text}
	for _, argument := range refused {
		assertStrictRefusal(t, argument, "scope.description_negation", "keep the wide negation")
	}
	king, bald := foAtom("king", "$count"), foAtom("bald", "$count")
	// Force an empty king extension in a nonempty (one-object) domain.
	noKing := countQ("every", "$count", foNot(king))
	outerNegation := foNot(v27Russell(king, bald))
	predicateNegation := v27Russell(king, foNot(bald))
	if !finiteCountermodel([]*folFormula{noKing, outerNegation}, predicateNegation, 1) {
		t.Fatal("expected no-king countermodel to the Russellian reading")
	}
}

func TestV31ConditionalMoodBoundary(t *testing.T) {
	for _, tc := range []struct {
		antecedent, consequent string
		kind                   gateConditionalKind
	}{
		{"the system is misaligned", "the system might seek power", materialConditional},
		{"the system would be connected", "the system might seek power", materialConditional},
		{"the system WERE misaligned", "the system might seek power", mightConditional},
		{"the match had been struck", "the match might have lit", mightConditional},
		{"the match had been struck", "the match could have lit", mightConditional},
		{"the match is dry", "the match might not have lit", mightConditional},
		{"the match is dry", "the match could not have lit", materialConditional},
		{"the system is misaligned", "it is not the case that the system might seek power", materialConditional},
		{"the system is misaligned", "either the system might seek power or the alarm sounds", materialConditional},
		{"the system is misaligned", "if the match had been struck, then the match might have lit", materialConditional},
	} {
		t.Run(tc.antecedent+"/"+tc.consequent, func(t *testing.T) {
			if got := conditionalKind(tc.consequent, tc.antecedent); got != tc.kind {
				t.Fatalf("kind=%v, want %v", got, tc.kind)
			}
		})
	}
}
