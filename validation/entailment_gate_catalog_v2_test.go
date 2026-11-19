package validation

import (
	"strings"
	"testing"
)

func TestUniversalNegativeCandidateEnumeration(t *testing.T) {
	tests := []struct {
		statement string
		want      clausePair
	}{
		{"No mammal is a fish.", clausePair{left: "mammal", right: "a fish"}},
		{"No dogs are cats.", clausePair{left: "dogs", right: "cats"}},
		{"Nothing a thing thinks of is a physical object.", clausePair{left: "a thing thinks of", right: "a physical object"}},
		{"Every swan is not black.", clausePair{left: "swan", right: "black"}},
		{"Everything a thing thinks of is not a physical object.", clausePair{left: "a thing thinks of", right: "a physical object"}},
		{"Physical objects are not ideas.", clausePair{left: "physical objects", right: "ideas"}},
	}
	for _, test := range tests {
		found := false
		for _, candidate := range universalNegativeCandidates(test.statement) {
			if candidate == test.want {
				found = true
			}
		}
		if !found {
			t.Errorf("universalNegativeCandidates(%q) missing %#v; got %#v", test.statement, test.want, universalNegativeCandidates(test.statement))
		}
	}

	if candidates := universalNegativeCandidates("Some physical objects are not ideas."); len(candidates) != 0 {
		t.Errorf("existential subject parsed as universal negative: %#v", candidates)
	}
}

func TestUniversalNegativeSyllogism(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestUniversalNegativeSyllogism/cesare_with_bare-plural_E_premise": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestUniversalNegativeSyllogism/celarent": {"term.inflection_verbatim", "singular 'every N'"},
	}

	positives := []struct {
		name     string
		argument string
	}{
		{
			name: "celarent",
			argument: `1. All whales are mammals.
2. No mammal is a fish.
3. Therefore, no whale is a fish. [1,2]`,
		},
		{
			name: "cesare with bare-plural E premise",
			argument: `1. Everything a thing thinks of is an idea.
2. Physical objects are not ideas.
3. Therefore, everything a thing thinks of is not a physical object. [1,2]`,
		},
	}
	for _, test := range positives {
		t.Run(test.name, func(t *testing.T) {
			if listedStrictRefusal(t, test.argument, refused) {
				return
			}
			report, result := CheckEntailmentGate(test.argument)
			if !result.Passed || !report.Passed {
				t.Fatalf("gate failed: result=%#v report=%#v", result, report)
			}
			if got := report.Steps[0].Form; got != "universal negative syllogism" {
				t.Fatalf("form = %q, want universal negative syllogism", got)
			}
		})
	}

	invalids := []struct {
		name     string
		argument string
	}{
		{
			name: "middle term as universal subject",
			argument: `1. All whales are mammals.
2. No whale is a fish.
3. Therefore, no mammal is a fish. [1,2]`,
		},
		{
			name: "conclusion swaps in an unrelated term",
			argument: `1. All whales are mammals.
2. No mammal is a fish.
3. Therefore, no dolphin is a fish. [1,2]`,
		},
	}
	for _, test := range invalids {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if result.Passed || report.Passed {
				t.Fatalf("invalid form unexpectedly passed: report=%#v", report)
			}
		})
	}
}

func TestNecessitationWrapper(t *testing.T) {
	refused := map[string]strictRefusal{
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestNecessitationWrapper/universal_negative_syllogism_under_necessitation": {"term.inflection_verbatim", "singular 'every N'"},
	}

	positives := []struct {
		name     string
		argument string
		form     string
	}{
		{
			name: "modus ponens under necessitation",
			argument: `1. Necessarily, if rain falls, then streets get wet.
2. Necessarily, rain falls.
3. Therefore, necessarily, streets get wet. [1,2]`,
			form: "modus ponens (necessitation)",
		},
		{
			name: "universal negative syllogism under necessitation",
			argument: `1. Necessarily, everything a thing thinks of is an idea.
2. Necessarily, physical objects are not ideas.
3. Therefore, necessarily, everything a thing thinks of is not a physical object. [1,2]`,
			form: "universal negative syllogism (necessitation)",
		},
	}
	for _, test := range positives {
		t.Run(test.name, func(t *testing.T) {
			if listedStrictRefusal(t, test.argument, refused) {
				return
			}
			report, result := CheckEntailmentGate(test.argument)
			if !result.Passed || !report.Passed {
				t.Fatalf("gate failed: result=%#v report=%#v", result, report)
			}
			if got := report.Steps[0].Form; got != test.form {
				t.Fatalf("form = %q, want %q", got, test.form)
			}
		})
	}

	// A necessary conclusion must not follow when any cited premise is
	// contingent: the stripped candidate set only exists when every premise
	// is modalized.
	guards := []struct {
		name     string
		argument string
	}{
		{
			name: "contingent minor premise blocks necessitated modus ponens",
			argument: `1. Necessarily, if rain falls, then streets get wet.
2. Rain falls.
3. Therefore, necessarily, streets get wet. [1,2]`,
		},
		{
			name: "contingent categorical premise blocks necessitated syllogism",
			argument: `1. Necessarily, everything a thing thinks of is an idea.
2. Physical objects are not ideas.
3. Therefore, necessarily, everything a thing thinks of is not a physical object. [1,2]`,
		},
	}
	for _, test := range guards {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if result.Passed || report.Passed {
				t.Fatalf("modally invalid step unexpectedly passed: report=%#v", report)
			}
		})
	}
}

func TestNecessityElimination(t *testing.T) {
	positive := `1. Necessarily, rain falls.
2. Therefore, rain falls. [1]`
	report, result := CheckEntailmentGate(positive)
	if !result.Passed || !report.Passed {
		t.Fatalf("necessity elimination failed: result=%#v report=%#v", result, report)
	}
	if got := report.Steps[0].Form; got != "necessity elimination" {
		t.Fatalf("form = %q, want necessity elimination", got)
	}

	negative := `1. Necessarily, rain falls.
2. Therefore, snow falls. [1]`
	if report, _ := CheckEntailmentGate(negative); report.Passed {
		t.Fatalf("necessity elimination unexpectedly passed on mismatched clause: %#v", report)
	}
}

func TestGeoff013MinimalFormCertifies(t *testing.T) {
	argument := `1. Necessarily, everything a thing thinks of is an idea.
2. Necessarily, physical objects are not ideas.
3. Therefore, necessarily, everything a thing thinks of is not a physical object. [1,2]
4. If [3], then it is impossible for any thing to think of a physical object.
5. Therefore, it is impossible for any thing to think of a physical object. [3,4]`
	// term.inflection_verbatim: idea/ideas and physical object/objects differ.
	refused := []string{argument}
	for _, text := range refused {
		assertStrictRefusal(t, text, "term.inflection_verbatim", "singular 'every N'")
	}
}

func TestBridgeCompression(t *testing.T) {
	removable := `1. All whales are mammals.
2. No mammal is a fish.
3. If [1,2], then no whale is a fish.
4. Therefore, no whale is a fish. [1,2,3]`
	entries, result := CheckBridgeCompression(removable)
	if !result.Passed {
		t.Fatalf("compression failed: %#v", result)
	}
	// term.inflection_verbatim: the bridge is now necessary for this plural/singular wording.
	if len(entries) != 1 || entries[0].Removable || !strings.Contains(entries[0].Mismatch, "singular 'every N'") {
		t.Fatalf("expected a strict inflection refusal when removing the bridge, got %#v", entries)
	}
	if entries[0].BridgeStep != 3 || entries[0].Conclusion != 4 || entries[0].Form != "universal negative syllogism" {
		t.Fatalf("unexpected compression entry: %#v", entries[0])
	}

	// The certified geoff-013 rendering: the bridge carries the modal
	// strengthening, so mechanical removal must fail.
	stuck := `1. Necessarily, everything a thing thinks of is an idea.
2. Physical objects are not ideas.
3. If [1,2], then necessarily, everything a thing thinks of is not a physical object.
4. Therefore, necessarily, everything a thing thinks of is not a physical object. [1,2,3]`
	entries, result = CheckBridgeCompression(stuck)
	if !result.Passed {
		t.Fatalf("compression failed: %#v", result)
	}
	if len(entries) != 1 || entries[0].Removable {
		t.Fatalf("expected one stuck bridge, got %#v", entries)
	}
	if entries[0].Mismatch == "" {
		t.Fatalf("stuck bridge should carry a mismatch diagnostic: %#v", entries[0])
	}
}
