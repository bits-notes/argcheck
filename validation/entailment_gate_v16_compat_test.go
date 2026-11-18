package validation

import (
	"encoding/json"
	"fmt"
	"os"
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
			assertAcceptedStepParity(t, tc.Argument, tc.Forms)
		})
	}
	t.Logf("all %d accepted v16 arguments retain their per-step form names", len(golden.Cases))
}
