package validation

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSourceQuoteAcceptance(t *testing.T) {
	source := "Before.  When mortgages become more expensive, fewer people buy homes. After."
	statement := "Therefore, when mortgages become more expensive, fewer people buy homes. [1,4]"
	tests := []struct {
		name, source, statement, quote string
		valid                          bool
	}{
		{"verbatim", source, statement, "When mortgages become more expensive, fewer people buy homes.", true},
		{"normalization", "😀 He said “mortgages”—\n  become MORE expensive!", "He said mortgages become more expensive.", `he said "mortgages"- become more expensive.`, true},
		{"unicode case folding", "Straße became more expensive.", "STRASSE became more expensive.", "STRASSE became more expensive.", true},
		{"absent", source, statement, "When mortgages become cheaper, more people buy homes.", false},
		{"unrelated", source, "The moon is made of cheese.", "When mortgages become more expensive, fewer people buy homes.", false},
		{"short", "People buy.", "People buy.", "People buy.", false},
		{"empty", source, statement, "", false},
		{"negation", source, "When mortgages become more expensive, fewer people do not buy homes.", "When mortgages become more expensive, fewer people buy homes.", false},
		{"negated contraction", source, "When mortgages become more expensive, fewer people don’t buy homes.", "When mortgages become more expensive, fewer people buy homes.", false},
		{"number changed", "Traffic fell by 20 percent in 2003.", "Traffic fell by 50 percent in 2003.", "Traffic fell by 20 percent in 2003.", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start, end, err := ValidateSourceQuote(tc.source, tc.statement, tc.quote)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
			if err == nil && (start < 0 || end <= start || end > len([]rune(tc.source))) {
				t.Fatalf("bad offsets %d:%d", start, end)
			}
		})
	}
}

func TestSourceContentAlignedThresholds(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"alpha beta gamma delta epsilon", "alpha beta gamma delta zeta", true}, // 4/5 containment, 4/6 Jaccard
		{"alpha beta gamma delta epsilon", "alpha beta gamma zeta eta", false},
		{"alpha beta gamma", "alpha beta gamma delta epsilon", true}, // 3/5 Jaccard
		{"alpha beta gamma", "alpha beta gamma delta epsilon zeta", false},
		{"Therefore, so thus [1,2]", "Therefore, so thus [1,2]", false},
	} {
		if got := SourceContentAligned(tc.a, tc.b); got != tc.want {
			t.Errorf("%q / %q = %v", tc.a, tc.b, got)
		}
	}
}

func TestRestoredRepairPromptAndCoverage(t *testing.T) {
	raw, err := os.ReadFile("testdata/restored_causal_chain.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Source  string
		AsFound string `json:"as_found"`
		Repair  json.RawMessage
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	prompt := "<source_material>" + fixture.Source + "</source_material><as_found_argument>" + fixture.AsFound + "</as_found_argument>"
	result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), string(fixture.Repair), prompt)
	if !result.Passed {
		t.Fatal(result.Message)
	}
	var repair formalizedRepairOutput
	if err = json.Unmarshal(fixture.Repair, &repair); err != nil {
		t.Fatal(err)
	}
	mappings := *repair.LineMapping
	if _, _, err = ValidateFormalizationCoverageAnchored("as_found", fixture.AsFound, *repair.Argument, mappings, nil, nil, fixture.Source); err != nil {
		t.Fatal(err)
	}
	// Both origins may be provided together, with independent verification.
	firstQuote := "When interest rates rise, mortgages become more expensive."
	mappings[0].FromSourceQuote = &firstQuote
	if _, _, err = ValidateFormalizationCoverageAnchored("as_found", fixture.AsFound, *repair.Argument, mappings, nil, nil, fixture.Source); err != nil {
		t.Fatal(err)
	}
	mappings[0].FromSourceQuote = nil
	// Restorations do not discharge unmapped anchor content.
	if _, _, err = ValidateFormalizationCoverageAnchored("as_found", fixture.AsFound, *repair.Argument, mappings[1:], nil, nil, fixture.Source); err == nil {
		t.Fatal("accepted uncovered anchor")
	}
	for _, prompt := range []string{"", "<source_material>unrelated source text</source_material>", strings.Replace(prompt, fixture.Source, "unrelated source text", 1)} {
		result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), string(fixture.Repair), prompt)
		if result.Passed || !strings.Contains(result.Message, "line 2") {
			t.Fatalf("missing/false source accepted: %+v", result)
		}
	}
	for _, quote := range []string{"When fewer", "The central bank has just raised interest rates.", "invented quote not in source"} {
		var changed formalizedRepairOutput
		if err := json.Unmarshal(fixture.Repair, &changed); err != nil {
			t.Fatal(err)
		}
		(*changed.LineMapping)[1].FromSourceQuote = &quote
		output, err := json.Marshal(changed)
		if err != nil {
			t.Fatal(err)
		}
		result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), string(output), prompt)
		if result.Passed || !strings.Contains(result.Message, "line 2") {
			t.Fatalf("bad quote %q: %+v", quote, result)
		}
	}
	if _, _, err := ValidateFormalizationCoverageAnchored("as_found", fixture.AsFound, *repair.Argument, mappings, nil, []int{2}, fixture.Source); err == nil {
		t.Fatal("restored line accepted as supplied")
	}
	// A mapping with anchor indices must not bypass a bad quote.
	bad := "unrelated quote that is not in the passage"
	mappings[0].FromSourceQuote = &bad
	if _, _, err = ValidateFormalizationCoverageAnchored("as_found", fixture.AsFound, *repair.Argument, mappings, nil, nil, fixture.Source); err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("bad quote with indices: %v", err)
	}
}

func TestLegacyRepairJSONUnchanged(t *testing.T) {
	var repair formalizedRepairOutput
	if err := json.Unmarshal([]byte(validFormalizedRepairOutput), &repair); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(repair)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "from_source_quote") {
		t.Fatal(string(raw))
	}
	result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), validFormalizedRepairOutput, "")
	if !result.Passed || string(result.ParsedOutput) != validFormalizedRepairOutput {
		t.Fatal("legacy output changed")
	}
}
