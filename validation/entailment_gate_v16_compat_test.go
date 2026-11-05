package validation

import (
	"encoding/json"
	"os"
	"testing"
)

// Frozen by running every Test function in catalog_v2..v16_test.go, the gap
// ledger, and a recursive JSON-string sweep of battery.json, cases.json and
// testdata/ against git 37ba4d4, before Fix 1 changed the gate. Capture occurred
// at GateNumberedArgument after bridge expansion, not from expected test flags.
// The golden stores every distinct accepted AST and its actual per-step forms.
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
		got := GateNumberedArgument(tc.Argument)
		if !got.Passed || len(got.Steps) != len(tc.Forms) {
			t.Errorf("v16 case %d (%v) regressed: %+v", i, tc.Sources, got)
			continue
		}
		for j, form := range tc.Forms {
			if got.Steps[j].Form != form {
				t.Errorf("v16 case %d (%v), step %d form %q, want %q", i, tc.Sources, got.Steps[j].Step, got.Steps[j].Form, form)
			}
		}
	}
	t.Logf("all %d accepted v16 arguments retain their per-step form names", len(golden.Cases))
}
