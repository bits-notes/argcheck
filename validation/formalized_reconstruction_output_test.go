package validation

import (
	"encoding/json"
	"strings"
	"testing"
)

const reconstructionTestArgument = "1. Every cat is an animal.\n2. Milo is a cat.\n3. Therefore, Milo is an animal. [1,2]"
const reconstructionTestPrompt = "<source_material>Every cat is an animal. Milo is a cat. Milo is an animal.</source_material>"

func TestFormalizedReconstructionOutput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*FormalizedReconstructionOutput)
		gate   bool
		want   string
	}{
		{name: "valid", gate: true},
		{name: "supplied universal", change: func(r *FormalizedReconstructionOutput) { r.Lines[0].Sentences = []int{} }, gate: true},
		{name: "syntax", change: func(r *FormalizedReconstructionOutput) { r.Argument = "Not numbered" }, want: "numbered_argument_syntax"},
		{name: "no conclusion", change: func(r *FormalizedReconstructionOutput) { r.Argument = "1. Milo is a cat.\n2. Every cat is an animal." }, want: "conclusion"},
		{name: "missing entry", change: func(r *FormalizedReconstructionOutput) { r.Lines = r.Lines[:2] }, want: "line 3 is missing"},
		{name: "duplicate entry", change: func(r *FormalizedReconstructionOutput) { r.Lines = append(r.Lines, r.Lines[0]) }, want: "line 1 is repeated"},
		{name: "extra entry", change: func(r *FormalizedReconstructionOutput) {
			r.Lines = append(r.Lines, ReconstructionLine{Line: 4, Sentences: []int{1}})
		}, want: "line 4 is outside"},
		{name: "missing sentences", change: func(r *FormalizedReconstructionOutput) { r.Lines[0].Sentences = nil }, want: "line 1 requires"},
		{name: "zero sentence", change: func(r *FormalizedReconstructionOutput) { r.Lines[0].Sentences = []int{0} }, want: "line 1 references sentence 0"},
		{name: "out of range", change: func(r *FormalizedReconstructionOutput) { r.Lines[0].Sentences = []int{4} }, want: "line 1 references sentence 4"},
		{name: "duplicate sentence", change: func(r *FormalizedReconstructionOutput) { r.Lines[0].Sentences = []int{1, 1} }, want: "line 1 repeats sentence 1"},
		{name: "all supplied", change: func(r *FormalizedReconstructionOutput) {
			r.Lines[0].Sentences = []int{}
			r.Lines[1].Sentences = []int{}
		}, want: "at least one premise"},
		{name: "final unsourced", change: func(r *FormalizedReconstructionOutput) { r.Lines[2].Sentences = []int{} }, want: "line 3 is the final conclusion"},
		{name: "associated conditional", change: func(r *FormalizedReconstructionOutput) {
			r.Argument = "1. If Milo is a cat, then Milo is an animal.\n2. Milo is a cat.\n3. Therefore, Milo is an animal. [1,2]"
			r.Lines[0].Sentences = []int{}
		}, want: "step 1 supplies an associated conditional"},
		{name: "gate off", change: func(r *FormalizedReconstructionOutput) {
			r.Argument = strings.Replace(r.Argument, "Therefore, Milo is an animal", "Therefore, Milo is a tree", 1)
		}},
		{name: "gate on", change: func(r *FormalizedReconstructionOutput) {
			r.Argument = strings.Replace(r.Argument, "Therefore, Milo is an animal", "Therefore, Milo is a tree", 1)
		}, gate: true, want: "entailment gate failed at step 3"},
		{name: "orphan", change: func(r *FormalizedReconstructionOutput) {
			r.Argument = "1. Unused premise.\n2. Milo is a cat.\n3. Therefore, Milo is a cat. [2]"
		}, want: "do not support"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := FormalizedReconstructionOutput{Argument: reconstructionTestArgument, Lines: []ReconstructionLine{{1, []int{1}}, {2, []int{2}}, {3, []int{3}}}}
			if tc.change != nil {
				tc.change(&r)
			}
			raw, _ := json.Marshal(r)
			options, _ := json.Marshal(formalizedRepairOutputOptions{EntailmentGate: tc.gate})
			result := ValidateFormalizedReconstructionOutputWithPrompt(t.Context(), string(raw), reconstructionTestPrompt, options)
			if tc.want == "" {
				if !result.Passed {
					t.Fatalf("%s", result.Message)
				}
			} else {
				if result.Passed || !strings.Contains(result.Message, tc.want) {
					t.Fatalf("got %v %q, want %q", result.Passed, result.Message, tc.want)
				}
				if result.Details["raw_reply_excerpt"] != string(raw) {
					t.Fatal("missing raw reply")
				}
			}
		})
	}
}

func TestReconstructionJSONNormalizationAndOptions(t *testing.T) {
	raw := `{"lines":[{"line":3,"sentences":[3]},{"line":1,"sentences":[1]},{"line":2,"sentences":[2]}],"argument":` + string(mustReconstructionJSON(t, reconstructionTestArgument)) + `}`
	for _, reply := range []string{raw, "Here is the reconstruction:\n```json\n" + raw + "\n```"} {
		result := ValidateFormalizedReconstructionOutputWithPrompt(t.Context(), reply, reconstructionTestPrompt)
		if !result.Passed || !strings.HasPrefix(string(result.ParsedOutput), `{"argument":`) {
			t.Fatalf("%+v", result)
		}
		var parsed FormalizedReconstructionOutput
		_ = json.Unmarshal(result.ParsedOutput, &parsed)
		if parsed.Lines[0].Line != 1 {
			t.Fatal("unstable line ordering")
		}
	}
	for _, reply := range []string{strings.Replace(raw, `"sentences":[3]`, `"sentences":[1.5]`, 1), strings.Replace(raw, `"sentences":[3]`, `"sentences":["3"]`, 1), raw + raw} {
		if result := ValidateFormalizedReconstructionOutputWithPrompt(t.Context(), reply, reconstructionTestPrompt); result.Passed {
			t.Fatal("accepted invalid JSON contract")
		}
	}
	if result := ValidateFormalizedReconstructionOutputWithPrompt(t.Context(), raw, ""); result.Passed {
		t.Fatal("accepted missing source")
	}
	spec := BuiltinSpec(BuiltinFormalizedReconstructionOutput)
	spec.Options = json.RawMessage(`{"unknown":true}`)
	if _, err := Compile(spec); err == nil {
		t.Fatal("accepted unknown option")
	}
	spec = LegacyKindSpec(BuiltinFormalizedReconstructionOutput)
	validator, err := Compile(spec)
	if err != nil {
		t.Fatal(err)
	}
	if result := validator.Validate(t.Context(), raw); !result.Passed {
		t.Fatal(result.Message)
	}
}
func mustReconstructionJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestLegacyMappingRoundTrip(t *testing.T) {
	raw := `{"repaired_line":1,"from_as_found_lines":[2],"from_source_quote":"A quotation."}`
	var mapping FormalizationLineMapping
	if err := json.Unmarshal([]byte(raw), &mapping); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(mapping)
	if string(encoded) != raw {
		t.Fatalf("legacy changed: %s", encoded)
	}
}

func TestSegmentSourceSentencesCodepoints(t *testing.T) {
	source := "Éva met Dr. Smith.\n2.\nCats sleep!"
	ranges := SegmentSourceSentences(source)
	if len(ranges) != 2 {
		t.Fatalf("ranges: %+v", ranges)
	}
	for i, r := range ranges {
		if r.Index != i {
			t.Fatal("indices must stay zero-based")
		}
	}
	if ranges[0].Start != 0 || ranges[0].End != 18 || string([]rune(source)[ranges[1].Start:ranges[1].End]) != "2.\nCats sleep!" {
		t.Fatalf("codepoint ranges: %+v", ranges)
	}
}

func TestLegacyCoverageIgnoresSentenceMappings(t *testing.T) {
	mappings := []FormalizationLineMapping{{RepairedLine: 1, FromAsFoundLines: []int{1}, FromSourceSentences: []int{-10}}, {RepairedLine: 2, FromAsFoundLines: []int{2}}, {RepairedLine: 3, FromAsFoundLines: []int{3}}}
	if _, _, err := ValidateFormalizationCoverage(reconstructionTestArgument, reconstructionTestArgument, mappings, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRestoredLineMappings("", reconstructionTestArgument, mappings); err != nil {
		t.Fatal(err)
	}
}

func TestReconstructionUnsourcedIntermediateConclusion(t *testing.T) {
	r := FormalizedReconstructionOutput{
		Argument: "1. If Milo is a cat, then Milo is an animal.\n2. Milo is a cat.\n3. Therefore, Milo is an animal. [1,2]\n4. If Milo is an animal, then Milo is alive.\n5. Therefore, Milo is alive. [3,4]",
		Lines:    []ReconstructionLine{{1, []int{}}, {2, []int{1}}, {3, []int{}}, {4, []int{2}}, {5, []int{3}}},
	}
	raw, _ := json.Marshal(r)
	result := ValidateFormalizedReconstructionOutputWithPrompt(t.Context(), string(raw), reconstructionTestPrompt, json.RawMessage(`{"entailment_gate":true}`))
	if !result.Passed {
		t.Fatal(result.Message)
	}
}
