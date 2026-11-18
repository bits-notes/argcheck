package validation

import (
	"strings"
	"testing"
)

func TestIdentitySubstitutionSafePositions(t *testing.T) {
	const identity = "\n2. Superman is identical to Clark.\n"
	reason := "substitution is certified only in identity sentences and in the subject of a plain sentence"
	for _, tc := range []struct {
		name, premise, conclusion string
		valid                     bool
	}{
		{"looking for", "Lois is looking for Superman", "Lois is looking for Clark", false},
		{"thinking about", "Lois is thinking about Superman", "Lois is thinking about Clark", false},
		{"admires", "Lois admires Superman", "Lois admires Clark", false},
		{"seeks", "Lois seeks Superman", "Lois seeks Clark", false},
		{"worships", "Lois worships Superman", "Lois worships Clark", false},
		{"imagines", "Lois imagines Superman", "Lois imagines Clark", false},
		{"needs", "Lois needs Superman", "Lois needs Clark", false},
		{"extensional object too", "Lois kissed Superman", "Lois kissed Clark", false},
		{"passive", "Superman is admired by Lois", "Clark is admired by Lois", false},
		{"raising", "Superman is believed to be a hero", "Clark is believed to be a hero", false},
		{"subject", "Superman can fly", "Clark can fly", true},
		{"subject copula", "Superman is a reporter", "Clark is a reporter", true},
		{"subject modal must", "Superman must be a hero", "Clark must be a hero", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := "1. " + tc.premise + "." + identity + "3. Therefore, " + tc.conclusion + ". [1,2]"
			report, result := CheckEntailmentGate(text)
			if result.Passed != tc.valid {
				t.Fatalf("got %v: %s", result.Passed, result.Message)
			}
			if !tc.valid && tc.name != "imagines" && tc.name != "worships" && tc.name != "needs" && tc.name != "seeks" && tc.name != "looking for" {
				if len(report.Steps) == 0 || !strings.Contains(report.Steps[len(report.Steps)-1].Mismatch, reason) {
					t.Fatalf("message lacks the position reason: %+v", report.Steps)
				}
			}
		})
	}
}

func TestIdentitySubstitutionChainsAndUnicorn(t *testing.T) {
	chain := "1. Superman is identical to Clark.\n2. Clark is identical to Kent.\n3. Therefore, Superman is identical to Kent. [1,2]"
	if _, result := CheckEntailmentGate(chain); !result.Passed {
		t.Fatalf("identity chain refused: %s", result.Message)
	}
	for _, text := range []string{
		"1. Lois is looking for a unicorn.\n2. Therefore, there is a unicorn that Lois is looking for. [1]",
		"1. Lois is looking for a unicorn.\n2. Therefore, something is a unicorn. [1]",
	} {
		if _, result := CheckEntailmentGate(text); result.Passed {
			t.Fatalf("unicorn existence certified: %s", text)
		}
	}
	quantified := "1. alice is identical to blair.\n2. every x is such that (alice admires x).\n3. Therefore, every x is such that (blair admires x). [1,2]"
	if _, result := CheckEntailmentGate(quantified); !result.Passed {
		t.Fatalf("subject inside a quantifier scope refused: %s", result.Message)
	}
}
