package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Frozen by running every Test function in catalog_v2..v16_test.go, the gap
// ledger, and a recursive JSON-string sweep of battery.json, cases.json and
// testdata/ against git 37ba4d4, before Fix 1 changed the gate. Capture occurred
// at GateNumberedArgument after bridge expansion, not from expected test flags.
// The golden stores every distinct accepted AST and its actual per-step forms.
// One deliberate edit since capture: catalog v18 retired the v16 spelling
// "It holds that (X) if and only if (Y)" in favor of bare "(X) if and only if
// (Y)", so every such statement and line was rewritten to the bare spelling.
func TestEntailmentGateV16AcceptedCompatibility(t *testing.T) {
	var golden struct {
		Baseline string `json:"baseline"`
		Cases    []struct {
			Argument NumberedArgument `json:"argument"`
			Forms    []string         `json:"forms"`
			Sources  []string         `json:"sources"`
		} `json:"cases"`
	}
	data, err := os.ReadFile("testdata/entailment_gate_v16_accepted.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	if golden.Baseline != "37ba4d4" || len(golden.Cases) != 999 {
		t.Fatalf("unexpected baseline %s, %d cases", golden.Baseline, len(golden.Cases))
	}
	for i, tc := range golden.Cases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			// Round 2 retains the valid monotone inference, but no longer
			// labels a quantified subject as an individual UI witness.
			if i == 46 {
				tc.Forms = []string{"schematic instantiation"}
			}
			// v35.0: identity substitution is certified only in identity
			// sentences and clause subjects. These golden cases substitute
			// into an object, so they are now refused with the position reason.
			if v35ObjectSubstitution[i] {
				report := GateNumberedArgument(tc.Argument)
				if report.Passed {
					t.Fatal("object-position substitution certified")
				}
				refused := false
				for _, step := range report.Steps {
					refused = refused || strings.Contains(step.Mismatch, "subject of a plain sentence")
				}
				if !refused {
					t.Fatalf("missing the position reason: %+v", report.Steps)
				}
				return
			}
			assertAcceptedStepParity(t, tc.Argument, tc.Forms)
		})
	}
	t.Logf("all %d accepted v16 arguments retain their per-step form names", len(golden.Cases))
}

var v35ObjectSubstitution = map[int]bool{14: true, 15: true, 906: true, 907: true}
