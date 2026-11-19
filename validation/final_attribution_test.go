package validation

import (
	"strings"
	"testing"
)

func TestFinalSingularAttribution(t *testing.T) {
	for _, variable := range []string{"U", "V", "W", "X", "Y", "Z"} {
		t.Run(variable, func(t *testing.T) {
			text := "1. If " + variable + " are useful, then " + variable + " are valuable.\n2. Training signals are useful.\n3. Therefore, training signals are valuable. [1,2]"
			report, result := CheckEntailmentGate(text)
			if report.Passed || result.Passed {
				t.Fatal("unexpected certification")
			}
			for _, hint := range []string{"Training signals", "need not be one thing", "Write the phrase itself into the rule"} {
				if !strings.Contains(result.Message, hint) {
					t.Errorf("missing %q: %s", hint, result.Message)
				}
			}
			if strings.Contains(result.Message, "an indefinite (a/an/some/someone") {
				t.Error(result.Message)
			}
		})
	}
}

// The literal-rule suggestion keeps the value as the writer spelled it: a name
// seen only at sentence start must not come back lowercased.
func TestFinalSingularSuggestionKeepsCase(t *testing.T) {
	text := "1. If X is useful, then X is valuable.\n2. Alpha is useful.\n3. Therefore, Alpha is valuable. [1,2]"
	_, result := CheckEntailmentGate(text)
	if result.Passed {
		t.Fatal("unexpected certification")
	}
	if !strings.Contains(result.Message, `("Alpha is useful")`) || strings.Contains(result.Message, `"alpha`) {
		t.Fatal(result.Message)
	}
}

func TestFinalDescriptionAttribution(t *testing.T) {
	text := "1. If W arrived, then W was seen.\n2. The Bob dog arrived.\n3. Therefore, the Bob dog was seen. [1,2]"
	report, result := CheckEntailmentGate(text)
	if report.Passed || result.Passed {
		t.Fatal("unexpected certification")
	}
	for _, hint := range []string{"slot.description_shape", "Name it"} {
		if !strings.Contains(result.Message, hint) {
			t.Errorf("missing %q: %s", hint, result.Message)
		}
	}
}

func TestFinalCounterfactualAttribution(t *testing.T) {
	argument := "1. If X had struck x, then X would have lit.\n2. Alice struck x.\n3. Therefore, Alice lit. [1,2]"
	assertStrictRefusal(t, argument, "slot.value_singular", "Name the thing")
}

func TestFinalOpaqueLiteralScope(t *testing.T) {
	for _, payload := range []string{`either it rains`, `perhaps Alice is red and Bob is blue`, `daily Alice sings or Bob dances`} {
		atom := `atom{"` + payload + `"}`
		text := "1. (" + atom + ") and (it rains).\n2. Therefore, " + atom + ". [1]"
		if _, result := CheckEntailmentGate(text); !result.Passed {
			t.Fatal(result.Message)
		}
	}
	for _, malformed := range []string{`atom{perhaps Alice}`, `atom{"p", x}`, `atom{"p"} trailing`} {
		if scopeLiteralAtom(malformed) {
			t.Errorf("malformed/unbound literal admitted: %s", malformed)
		}
	}
}

func TestFinalCanonicalDistributionScope(t *testing.T) {
	text := "1. It is the case both that it rains, and that either it snows or the street is wet.\n2. Therefore, either it is the case both that it rains, and that it snows or it is the case both that it rains, and that the street is wet. [1]"
	if _, result := CheckEntailmentGate(text); !result.Passed {
		t.Fatal(result.Message)
	}
}

func TestFinalNegatedIdentityTerms(t *testing.T) {
	argument := "1. Socrates has some property that Plato lacks.\n2. Therefore, It is not the case that Socrates is identical to Plato. [1]"
	if _, result := CheckEntailmentGate(argument); !result.Passed {
		t.Fatal(result.Message)
	}
	policy := allStrictRules()
	names := map[string]bool{"socrates": true, "plato": true}
	policy.slotNames = &names
	for _, text := range []string{"It is not the case that all dogs is identical to Plato", "It is not the case that the majority is identical to Plato"} {
		pairs := policy.round2IdentityPairs(text)
		if len(pairs) != 1 || policy.singularValue(pairs[0].left) {
			t.Fatalf("unsafe identity term admitted: %s", text)
		}
	}
}
