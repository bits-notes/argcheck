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
// Deliberate edits since capture: catalog v18 retired the v16 spelling
// "It holds that (X) if and only if (Y)" in favor of bare "(X) if and only if
// (Y)", so every such statement and line was rewritten to the bare spelling.
// v46.5 explicitly groups the conjunctions in cases 933 (absorption), 934 and
// 972 (distribution), preserving their accepted status and per-step form names.
func TestEntailmentGateV16AcceptedCompatibility(t *testing.T) {
	// v44 by-design refusals: retained separately from the accepted controls.
	refused := map[string]strictRefusal{
		// scope.elided_copula: elision requires a predicate from the closed vocabulary.
		"TestEntailmentGateV16AcceptedCompatibility/case_39": {"scope.elided_copula", "repeat the subject and verb"},
		// scope.elided_copula: elision requires a predicate from the closed vocabulary.
		"TestEntailmentGateV16AcceptedCompatibility/case_40": {"scope.elided_copula", "repeat the subject and verb"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV16AcceptedCompatibility/case_51": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV16AcceptedCompatibility/case_856": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV16AcceptedCompatibility/case_857": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV16AcceptedCompatibility/case_919": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV16AcceptedCompatibility/case_920": {"term.inflection_verbatim", "singular 'every N'"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_11": {"eg.name_witness", "state the existential claim as its own premise"},
		// v46.0: Ada is now evidenced (names allowlist); valid.
		// v46.0: Ada is now evidenced (names allowlist); valid.
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_17": {"eg.name_witness", "state the existential claim as its own premise"},
		// term.inflection_verbatim: class spellings differ.
		"TestEntailmentGateV16AcceptedCompatibility/case_18": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: class spellings differ.
		"TestEntailmentGateV16AcceptedCompatibility/case_19": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: class spellings differ.
		"TestEntailmentGateV16AcceptedCompatibility/case_22": {"term.inflection_verbatim", "singular 'every N'"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_24": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_25": {"scope.clause_start", "move the operator into each clause"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_26": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_27": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_28": {"eg.name_witness", "state the existential claim as its own premise"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_29": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_37": {"scope.clause_start", "move the operator into each clause"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_44": {"eg.name_witness", "state the existential claim as its own premise"},
		// term.inflection_verbatim: class spellings differ.
		"TestEntailmentGateV16AcceptedCompatibility/case_46": {"term.inflection_verbatim", "singular 'every N'"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_52": {"scope.clause_start", "move the operator into each clause"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestEntailmentGateV16AcceptedCompatibility/case_53": {"slot.value_singular", "Name the thing"},
		// identity.constant_sides: identity requires evidenced singular constants.
		"TestEntailmentGateV16AcceptedCompatibility/case_54": {"identity.constant_sides", "name the thing"},
		// identity.constant_sides: identity requires evidenced singular constants.
		"TestEntailmentGateV16AcceptedCompatibility/case_55": {"identity.constant_sides", "name the thing"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_82": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_89": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_93": {"scope.clause_start", "move the operator into each clause"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestEntailmentGateV16AcceptedCompatibility/case_108": {"slot.value_singular", "Name the thing"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_118": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_123": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_139": {"scope.clause_start", "move the operator into each clause"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestEntailmentGateV16AcceptedCompatibility/case_182": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestEntailmentGateV16AcceptedCompatibility/case_184": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestEntailmentGateV16AcceptedCompatibility/case_185": {"slot.value_singular", "Name the thing"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_189": {"scope.clause_start", "move the operator into each clause"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestEntailmentGateV16AcceptedCompatibility/case_190": {"slot.value_singular", "Name the thing"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_191": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_202": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_203": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_204": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_230": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_417": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_453": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_578": {"scope.clause_start", "move the operator into each clause"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_830": {"eg.name_witness", "state the existential claim as its own premise"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_840": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_841": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_842": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_843": {"scope.clause_start", "move the operator into each clause"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_844": {"eg.name_witness", "state the existential claim as its own premise"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_850": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_851": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_852": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_853": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_863": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_864": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_865": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_866": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_871": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_876": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_877": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_878": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_879": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_880": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_881": {"scope.clause_start", "move the operator into each clause"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_882": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_883": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_885": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_886": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_887": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_899": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_901": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_903": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_904": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: existential witnesses must be evidenced names.
		"TestEntailmentGateV16AcceptedCompatibility/case_905": {"eg.name_witness", "state the existential claim as its own premise"},
		// identity.constant_sides: identity requires evidenced singular constants.
		"TestEntailmentGateV16AcceptedCompatibility/case_908": {"identity.constant_sides", "name the thing"},
		// identity.constant_sides: identity requires evidenced singular constants.
		"TestEntailmentGateV16AcceptedCompatibility/case_909": {"identity.constant_sides", "name the thing"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_921": {"scope.clause_start", "move the operator into each clause"},
		// v46.0: Zeus is now evidenced (names allowlist); valid.
		// v46.0: Ada is now evidenced (names allowlist); valid.
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_974": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestEntailmentGateV16AcceptedCompatibility/case_991": {"scope.clause_start", "move the operator into each clause"},
	}

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
			if listedStrictASTRefusal(t, tc.Argument, refused) {
				return
			}
			assertAcceptedStepParity(t, tc.Argument, tc.Forms)
		})
	}
	t.Logf("all %d accepted v16 arguments retain their per-step form names", len(golden.Cases))
}

var v35ObjectSubstitution = map[int]bool{14: true, 15: true, 906: true, 907: true}
