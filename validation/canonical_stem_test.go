package validation

import (
	"strings"
	"testing"
)

func TestCanonicalEnglishStem(t *testing.T) {
	for _, family := range [][]string{
		{"hold", "holds", "holder", "holders", "holding"},
		{"win", "wins", "winning", "winner", "winners"},
		{"eligible", "eligibility", "eligibilities"},
		{"adapt", "adapted", "adapting", "adaptable", "adaptability", "adaptabilities"},
		{"flex", "flexible", "flexibility"},
		{"civil", "civility", "civilities"},
		{"box", "boxes"}, {"wish", "wishes"}, {"study", "studies"},
		{"quick", "quickly"}, {"propose", "proposed"}, {"make", "making"},
	} {
		for _, word := range family[1:] {
			if got, want := canonicalEnglishStem(word), canonicalEnglishStem(family[0]); got != want {
				t.Errorf("stem(%q) = %q; stem(%q) = %q", word, got, family[0], want)
			}
		}
	}
	for _, word := range []string{"far", "red", "no", "not", "never", "123", "42nd", "café", "glass", "status", "analysis"} {
		if got := canonicalEnglishStem(word); got != word {
			t.Errorf("stem(%q) = %q", word, got)
		}
	}
}

func TestCanonicalStemContentChecks(t *testing.T) {
	for _, tc := range []struct{ name, original, canonical, errorText string }{
		{"badge holder", "Imani holds a badge.", "Imani is a badge holder.", ""},
		{"badge reverse retention", "Imani is a badge holder.", "Imani holds a badge.", ""},
		// "year's" contributes separate "year" and "s" tokens: losing rules/make retains 8 of 10 stems.
		{"fellowship accepted at exactly 80 percent retention despite losing rules and make", "The rules make winning last year's departmental prize sufficient for eligibility for the fellowship.", "If a person wins last year's departmental prize, then that person is eligible for the fellowship.", ""},
		{"new modifier", "The reservoir is below the red mark.", "The reservoir is far below the red mark.", "new content words: far"},
		{"proposed is new", "The response-dependent account holds.", "The proposed response-dependent account holds.", "new content words: proposed"},
		{"real loss", "Imani holds a bright red badge.", "Imani is a badge holder.", "less than 80%"},
		{"exact threshold", "Imani holds a bright red badge.", "Imani is a red badge holder.", ""},
		{"framing not globally ignored", "Imani holds a badge.", "Imani makes rules for a badge holder.", "new content words:"},
		{"negation still surface", "Imani holds no badge.", "Imani is a badge holder.", "negation count changed"},
		{"number still surface", "Imani holds 2 badges.", "Imani is a 3 badge holder.", "number tokens changed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := CanonicalContentCheck(tc.original, tc.canonical)
			if tc.errorText == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.errorText) {
				t.Fatalf("error = %v; want %q", err, tc.errorText)
			}
		})
	}
	for word := range canonicalFormWords {
		if word == "no" || word == "not" {
			continue
		}
		if err := CanonicalContentCheck("Imani holds a badge.", "Imani holds a badge "+word+"."); err != nil {
			t.Errorf("form word %q: %v", word, err)
		}
	}
}

func TestCanonicalStemmingDoesNotLoosenQuotes(t *testing.T) {
	for _, statement := range []string{"Imani is a badge holder.", "Imani holds a badge person."} {
		quote := "Imani holds a badge."
		if SourceContentAligned(statement, quote) {
			t.Errorf("surface overlap accepted %q", statement)
		}
		if _, _, err := ValidateSourceQuote(quote, statement, quote); err == nil {
			t.Errorf("quote accepted %q", statement)
		}
	}
	if !SourceContentAligned("Imani holds a badge.", "Imani holds a badge.") {
		t.Fatal("exact quote rejected")
	}
}
