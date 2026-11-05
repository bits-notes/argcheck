package validation

import "testing"

// Catalog v18 makes parentheses alone the biconditional-side scaffold. A
// leading "it is the case that" is the ordinary wrapper and may precede it;
// "it holds that" is ordinary text with no special meaning.

func TestCatalogV18Version(t *testing.T) {
	if EntailmentGateCatalogVersion != "v18" {
		t.Fatalf("catalog = %q, want v18", EntailmentGateCatalogVersion)
	}
}

func TestCatalogV18ParenthesizedBiconditionalSides(t *testing.T) {
	tests := []struct {
		name, argument, wantForm string
	}{
		{"bare parentheses", `1. (If it rains, then the street is wet) if and only if (the city is prepared).
2. If it rains, then the street is wet.
3. Therefore, the city is prepared. [1,2]`, "biconditional modus ponens"},
		{"wrapped in it is the case that", `1. It is the case that (if it rains, then the street is wet) if and only if (the city is prepared).
2. If it rains, then the street is wet.
3. Therefore, the city is prepared. [1,2]`, "biconditional modus ponens"},
		{"wrapper and bare spellings are the same clause", `1. It is the case that (if it rains, then the street is wet) if and only if (the city is prepared).
2. Therefore, (if it rains, then the street is wet) if and only if (the city is prepared). [1]`, "self-entailment"},
		{"unparenthesized reading unchanged", `1. If it rains, then the street is wet if and only if the city is prepared.
2. It rains.
3. Therefore, the street is wet if and only if the city is prepared. [1,2]`, "modus ponens"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if !report.Passed || !result.Passed {
				t.Fatalf("rejected: %s\n%s", result.Message, tc.argument)
			}
			if got := report.Steps[len(report.Steps)-1].Form; got != tc.wantForm {
				t.Fatalf("form = %q, want %q", got, tc.wantForm)
			}
		})
	}
}

func TestCatalogV18ItHoldsThatIsOrdinaryText(t *testing.T) {
	for _, argument := range []string{
		`1. It holds that it rains.
2. Therefore, it rains. [1]`,
		`1. It holds that (if it rains, then the street is wet) if and only if (the city is prepared).
2. If it rains, then the street is wet.
3. Therefore, the city is prepared. [1,2]`,
	} {
		if report, _ := CheckEntailmentGate(argument); report.Passed {
			t.Errorf("\"it holds that\" treated as special:\n%s", argument)
		}
	}
}
