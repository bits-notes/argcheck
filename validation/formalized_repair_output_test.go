package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const validFormalizedRepairOutput = `{
  "argument": "1. If A, then B.\n2. A.\n3. Therefore, B. [1,2]",
  "coverage_anchor": "source",
  "line_mapping": [
    {"repaired_line": 1, "from_as_found_lines": [1]},
    {"repaired_line": 2, "from_as_found_lines": [2]},
    {"repaired_line": 3, "from_as_found_lines": [3]}
  ],
  "dropped": []
}`

func TestFormalizedRepairOutputValidatorAcceptsContractAndParsesJSON(t *testing.T) {
	result := Validate(context.Background(), BuiltinSpec(BuiltinFormalizedRepairOutput), validFormalizedRepairOutput)
	if !result.Passed || len(result.ParsedOutput) == 0 || result.Details["step_count"] != 3 {
		t.Fatalf("validation result = %+v", result)
	}
}

func TestFormalizedRepairOutputValidatorRejectsMalformedContract(t *testing.T) {
	tests := []struct {
		name, output, want string
	}{
		{name: "missing dropped", output: `{"argument":"1. A.\n2. Therefore, A. [1]","coverage_anchor":"source","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]}]}`, want: "requires dropped"},
		{name: "missing coverage anchor", output: strings.Replace(validFormalizedRepairOutput, "  \"coverage_anchor\": \"source\",\n", "", 1), want: "requires coverage_anchor"},
		{name: "unknown top-level field", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [], "extra": true`, 1), want: "unknown field"},
		{name: "unknown dropped field", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [{"as_found_line":4,"kind":"premise","reason":"unused","text":"model-authored"}]`, 1), want: "unknown field"},
		{name: "invalid kind", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [{"as_found_line":4,"kind":"other","reason":"unused"}]`, 1), want: "kind must be"},
		{name: "empty reason", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [{"as_found_line":4,"kind":"premise","reason":" "}]`, 1), want: "reason must be non-empty"},
		{name: "empty mapping sources", output: strings.Replace(validFormalizedRepairOutput, `"from_as_found_lines": [1]`, `"from_as_found_lines": []`, 1), want: "at least one line"},
		{name: "mapping outside argument", output: strings.Replace(validFormalizedRepairOutput, `"repaired_line": 3`, `"repaired_line": 4`, 1), want: "outside the argument"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Validate(context.Background(), BuiltinSpec(BuiltinFormalizedRepairOutput), testCase.output)
			if result.Passed || !strings.Contains(result.Message, testCase.want) {
				t.Fatalf("validation result = %+v, want message containing %q", result, testCase.want)
			}
		})
	}
}

func TestFormalizedRepairOutputValidatorAppliesNumberedArgumentSyntax(t *testing.T) {
	invalid := strings.Replace(validFormalizedRepairOutput, "3. Therefore, B. [1,2]", "3. B. [1,2]", 1)
	result := Validate(context.Background(), BuiltinSpec(BuiltinFormalizedRepairOutput), invalid)
	if result.Passed || !strings.Contains(result.Message, "failed numbered_argument_syntax") {
		t.Fatalf("validation result = %+v", result)
	}
}

func TestFormalizedRepairOutputValidatorRejectsInvalidOptions(t *testing.T) {
	for _, raw := range []string{`{"allow_blank_lines":true}`, `{"entailment_gate":"true"}`, `{"entailment_gate":true} {}`} {
		spec := BuiltinSpec(BuiltinFormalizedRepairOutput)
		spec.Options = []byte(raw)
		if _, err := Compile(spec); err == nil {
			t.Fatalf("Compile accepted options %s", raw)
		}
	}
}

func TestFormalizedRepairOutputValidatorEnforcesSourceCoverageFromRepairPrompt(t *testing.T) {
	prompt := `Repair.
<source_material>
1. If it rains, then the ground is wet.
2. It rains.
3. Therefore, the ground is wet. [1,2]
4. If the sun shines, then the sky is bright.
5. The sun shines.
6. Therefore, the sky is bright. [4,5]
</source_material>
<as_found_argument>
1. If it rains, then the ground is wet.
2. It rains.
3. Therefore, the ground is wet. [1,2]
</as_found_argument>`
	output := `{
  "argument":"1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]",
  "coverage_anchor":"source",
  "line_mapping":[
    {"repaired_line":1,"from_as_found_lines":[1]},
    {"repaired_line":2,"from_as_found_lines":[2]},
    {"repaired_line":3,"from_as_found_lines":[3]}
  ],
  "dropped":[
    {"as_found_line":4,"kind":"premise","reason":"separate sub-argument"},
    {"as_found_line":5,"kind":"premise","reason":"separate sub-argument"},
    {"as_found_line":6,"kind":"conclusion","reason":"separate sub-argument"}
  ]
}`
	result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), output, prompt)
	if !result.Passed || result.Details["coverage_anchor"] != FormalizationCoverageAnchorSource {
		t.Fatalf("source-anchored validation = %+v", result)
	}

	orphan := strings.Replace(output, `,
    {"as_found_line":6,"kind":"conclusion","reason":"separate sub-argument"}`, "", 1)
	result = ValidateFormalizedRepairOutputWithPrompt(context.Background(), orphan, prompt)
	if result.Passed || !strings.Contains(result.Message, "source line 6 is not covered") {
		t.Fatalf("orphan source validation = %+v", result)
	}
}

func TestFormalizedRepairOutputValidatorAcceptsDeadPremiseAndDerivedConclusionFixtures(t *testing.T) {
	deadPrompt := `<source_material>
1. If it rains, then the ground is wet.
2. It rains.
3. The moon orbits the Earth.
4. Therefore, the ground is wet. [1,2]
</source_material>
<as_found_argument>
1. If it rains, then the ground is wet.
2. It rains.
3. Therefore, the ground is wet. [1,2]
</as_found_argument>`
	deadOutput := `{"argument":"1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]","coverage_anchor":"source","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]},{"repaired_line":2,"from_as_found_lines":[2]},{"repaired_line":3,"from_as_found_lines":[4]}],"dropped":[{"as_found_line":3,"kind":"premise","reason":"dead premise"}]}`
	if result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), deadOutput, deadPrompt); !result.Passed {
		t.Fatalf("dead-premise repair rejected: %+v", result)
	}

	derivedPrompt := `<source_material>
1. A thermometer recorded 38.4 degrees.
2. The measurement conditions were valid.
3. Therefore, the patient had pyrexia. [1,2]
</source_material>
<as_found_argument>
1. A thermometer recorded 38.4 degrees.
2. The measurement conditions were valid.
3. Therefore, the patient had pyrexia. [1,2]
</as_found_argument>`
	derivedOutput := `{"argument":"1. A thermometer recorded 38.4 degrees.\n2. The measurement conditions were valid.\n3. If a thermometer recorded 38.4 degrees under valid conditions, then the temperature exceeded 38 degrees.\n4. Therefore, the temperature exceeded 38 degrees. [1,2,3]\n5. If the temperature exceeded 38 degrees, then the patient had pyrexia.\n6. Therefore, the patient had pyrexia. [4,5]","coverage_anchor":"source","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]},{"repaired_line":2,"from_as_found_lines":[2]},{"repaired_line":6,"from_as_found_lines":[3]}],"dropped":[]}`
	if result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), derivedOutput, derivedPrompt); !result.Passed {
		t.Fatalf("derived-conclusion repair rejected: %+v", result)
	}
}

func TestFormalizedRepairExtraction(t *testing.T) {
	for _, tc := range []struct {
		name, output, normalization string
		passed                      bool
	}{
		{"bare", validFormalizedRepairOutput, "", true},
		{"prose", "I'll repair this.\n" + validFormalizedRepairOutput + "\nDone.", "extracted_embedded_json", true},
		{"json fence", "The repair:\n```json\n" + validFormalizedRepairOutput + "\n```", "extracted_fenced_json", true},
		{"plain fence", "```\n" + validFormalizedRepairOutput + "\n```", "extracted_fenced_json", true},
		{"two objects", validFormalizedRepairOutput + "\n" + validFormalizedRepairOutput, "", false},
		{"object outside fence", "```json\n" + validFormalizedRepairOutput + "\n```\n{}", "", false},
		{"two fences", "```json\n" + validFormalizedRepairOutput + "\n```\n```\n{}\n```", "", false},
		{"prose only", "I'll repair this argument.", "", false},
		{"array", "[" + validFormalizedRepairOutput + "]", "", false},
		{"null", "null", "", false},
		{"unknown field wrapped", "The repair: " + strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [], "extra": true`, 1), "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := Validate(t.Context(), BuiltinSpec(BuiltinFormalizedRepairOutput), tc.output)
			if result.Passed != tc.passed {
				t.Fatalf("result = %+v", result)
			}
			if !tc.passed {
				if result.Details["raw_reply_excerpt"] != tc.output || result.Details["raw_reply_chars"] != len([]rune(tc.output)) {
					t.Fatalf("missing raw reply: %+v", result)
				}
				return
			}
			if string(result.ParsedOutput) != validFormalizedRepairOutput {
				t.Fatalf("parsed output = %s", result.ParsedOutput)
			}
			notes, _ := result.Details["normalizations"].([]string)
			if tc.normalization == "" {
				if len(notes) != 0 {
					t.Fatal(notes)
				}
			} else if len(notes) != 1 || notes[0] != tc.normalization {
				t.Fatal(notes)
			}
		})
	}
}

func TestRepairExtractionStringAware(t *testing.T) {
	object := `{"argument":"braces { } and escaped quote \" and slash \\"}`
	got, note, err := extractRepairJSONObject("Text " + object + " trailing text")
	if err != nil || got != object || note != "extracted_embedded_json" {
		t.Fatalf("%q %q %v", got, note, err)
	}
}

func TestRepairFailureExcerptUsesCharacters(t *testing.T) {
	output := strings.Repeat("界", 4100)
	result := ValidateFormalizedRepairOutputWithPrompt(t.Context(), output, "")
	if result.Passed || result.Details["raw_reply_chars"] != 4100 || result.Details["raw_reply_excerpt"] != strings.Repeat("界", 4000) {
		t.Fatalf("excerpt length/count incorrect")
	}
}

func TestRepairAnchorNormalizationPreservesCoverageChecks(t *testing.T) {
	prompt := "<source_material>A prose argument.</source_material><as_found_argument>1. If A, then B.\n2. A.\n3. Therefore, B. [1,2]</as_found_argument>"
	result := ValidateFormalizedRepairOutputWithPrompt(t.Context(), "Here: "+validFormalizedRepairOutput, prompt)
	if !result.Passed {
		t.Fatalf("%+v", result)
	}
	notes := result.Details["normalizations"].([]string)
	if len(notes) != 2 || notes[0] != "extracted_embedded_json" || notes[1] != "coverage_anchor:source->as_found" {
		t.Fatal(notes)
	}
	if !strings.Contains(string(result.ParsedOutput), `"coverage_anchor":"as_found"`) {
		t.Fatalf("%s", result.ParsedOutput)
	}
	bad := strings.Replace(validFormalizedRepairOutput, `"from_as_found_lines": [3]`, `"from_as_found_lines": [2]`, 1)
	result = ValidateFormalizedRepairOutputWithPrompt(t.Context(), bad, prompt)
	if result.Passed || !strings.Contains(result.Message, "coverage validation failed") || result.Details["raw_reply_excerpt"] != bad {
		t.Fatalf("%+v", result)
	}
}

func TestFormalizedRepairEntailmentGateOptIn(t *testing.T) {
	failed := strings.Replace(validFormalizedRepairOutput, "2. A.", "2. Something like A.", 1)
	for _, tc := range []struct {
		name, options, output string
		passed                bool
	}{
		{"default off", "", failed, true},
		{"explicit off", `{"entailment_gate":false}`, failed, true},
		{"on mismatch", `{"entailment_gate":true}`, failed, false},
		{"on valid", `{"entailment_gate":true}`, validFormalizedRepairOutput, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec := BuiltinSpec(BuiltinFormalizedRepairOutput)
			spec.Options = json.RawMessage(tc.options)
			results := []Result{
				Validate(t.Context(), spec, tc.output),
				ValidateFormalizedRepairOutputWithPrompt(t.Context(), tc.output, "", spec.Options),
			}
			for _, result := range results {
				if result.Passed != tc.passed {
					t.Fatalf("result = %+v", result)
				}
				if tc.passed {
					continue
				}
				var repair formalizedRepairOutput
				if err := json.Unmarshal([]byte(tc.output), &repair); err != nil {
					t.Fatal(err)
				}
				report, _ := CheckEntailmentGate(*repair.Argument)
				for _, want := range []string{"entailment gate failed at step 3 (cites [1 2])", report.Steps[0].Mismatch, "If A, then B.", "Either A or B.", "It is not the case that A.", "Every S is P.", "Do not change the argument's content."} {
					if !strings.Contains(result.Message, want) {
						t.Fatalf("message %q missing %q", result.Message, want)
					}
				}
				if result.Details["raw_reply_excerpt"] != tc.output || result.Details["entailment_gate"] == nil || result.Details["entailment_gate_message"] != result.Message {
					t.Fatalf("details = %+v", result.Details)
				}
			}
		})
	}
}

func TestFormalizedRepairEntailmentGateLimitsFeedback(t *testing.T) {
	output := `{"argument":"1. A.\n2. Therefore, B. [1]\n3. Therefore, C. [2]\n4. Therefore, D. [3]\n5. Therefore, E. [4]","coverage_anchor":"source","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]}],"dropped":[]}`
	spec := BuiltinSpec(BuiltinFormalizedRepairOutput)
	spec.Options = json.RawMessage(`{"entailment_gate":true}`)
	result := Validate(t.Context(), spec, output)
	if result.Passed || strings.Count(result.Message, "entailment gate failed at step") != 3 || strings.Contains(result.Message, "step 5") {
		t.Fatalf("result = %+v", result)
	}
	steps := result.Details["entailment_gate"].(map[string]any)["steps"].([]EntailmentGateStepReport)
	if len(steps) != 4 {
		t.Fatalf("full diagnostics lost: %+v", steps)
	}
}

func TestFormalizedRepairEntailmentGatePreservesStructuralFailure(t *testing.T) {
	output := strings.Replace(validFormalizedRepairOutput, "3. Therefore, B.", "3. B.", 1)
	spec := BuiltinSpec(BuiltinFormalizedRepairOutput)
	before := Validate(t.Context(), spec, output)
	spec.Options = json.RawMessage(`{"entailment_gate":true}`)
	after := Validate(t.Context(), spec, output)
	if after.Passed || before.Message != after.Message {
		t.Fatalf("before=%+v after=%+v", before, after)
	}
}

func TestRepairQuoteOverlapActionableMessages(t *testing.T) {
	for _, tc := range []struct {
		name, statement, quote, source, missing string
		whole                                   bool
	}{
		{"homes", "If fewer people buy homes, then house prices will fall.", "When fewer people buy homes, house prices fall.", "When fewer people buy homes, house prices fall.", "will", false},
		{"homes fragment", "If fewer people buy homes, then house prices will fall.", "When fewer people buy homes, house prices fall.", "In this market, when fewer people buy homes, house prices fall.", "will", true},
		{"neighboring sentence", "If fewer people buy homes, then house prices will fall.", "When fewer people buy homes, house prices fall.", "When fewer people buy homes, house prices fall. This is another sentence.", "will", false},
		{"congestion", "The revenue from a congestion charge in Manhattan could fund the subway.", "the revenue could fund the subway", "So a congestion charge in Manhattan would reduce traffic there too, and the revenue could fund the subway.", "congestion, charge, manhattan", true},
		{"normalized fragment", "The revenue from a congestion charge in Manhattan could fund the subway.", "THE REVENUE could fund the subway", "So a congestion charge in Manhattan would reduce traffic there too, and the revenue\ncould fund the subway.", "congestion, charge, manhattan", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := json.Marshal(map[string]any{
				"argument":        "1. A.\n2. B.\n3. " + tc.statement + "\n4. Therefore, A. [1,2,3]",
				"coverage_anchor": "as_found", "dropped": []any{},
				"line_mapping": []any{map[string]any{"repaired_line": 3, "from_source_quote": tc.quote}},
			})
			if err != nil {
				t.Fatal(err)
			}
			result := ValidateFormalizedRepairOutputWithPrompt(t.Context(), string(output), "<source_material>"+tc.source+"</source_material>")
			if result.Passed || !strings.Contains(result.Message, "repaired line 3: from_source_quote does not substantially overlap") || !strings.Contains(result.Message, "content words missing from quote: "+tc.missing) {
				t.Fatalf("unexpected diagnostic: %s", result.Message)
			}
			suggestion := fmt.Sprintf("quote the whole source sentence: %q", tc.source)
			if tc.whole != strings.Contains(result.Message, suggestion) {
				t.Fatalf("whole sentence diagnostic: %s", result.Message)
			}
			if !tc.whole && strings.Contains(result.Message, "quote the whole") {
				t.Fatal(result.Message)
			}
		})
	}
}

func TestRepairUnmappedSplitPremiseMessage(t *testing.T) {
	sentence := "Most of the suffering caused by factory farming is unnecessary because people can eat well without meat."
	output := `{"argument":"1. A.\n2. People can eat well without meat.\n3. Therefore, A. [1,2]","coverage_anchor":"as_found","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]},{"repaired_line":3,"from_as_found_lines":[2]}],"dropped":[]}`
	for _, tc := range []struct{ name, source, asFound, want string }{
		{"as-found split", sentence, "1. A.\n2. B.\n3. " + sentence, fmt.Sprintf("repaired premise line 2 restates part of as-found line 3 (%q); add 3 to its from_as_found_lines (a split premise may map to several repaired lines)", strings.TrimSuffix(sentence, "."))},
		{"source only", sentence, "1. A.\n2. B.", fmt.Sprintf("repaired premise line 2 restates part of a source sentence; add from_source_quote: %q", sentence)},
		{"numbered source anchor", "1. A.\n2. B.\n3. " + sentence, "1. " + sentence, fmt.Sprintf("repaired premise line 2 restates part of source line 3 (%q); add 3 to its from_as_found_lines (a split premise may map to several repaired lines)", strings.TrimSuffix(sentence, "."))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prompt := "<source_material>" + tc.source + "</source_material><as_found_argument>" + tc.asFound + "</as_found_argument>"
			result := ValidateFormalizedRepairOutputWithPrompt(t.Context(), output, prompt)
			if result.Passed || result.Message != tc.want {
				t.Fatalf("got %s; want %s", result.Message, tc.want)
			}
		})
	}
}

func TestRepairGateVerbatimClauseDiff(t *testing.T) {
	argument := "1. If fewer people buy homes, then house prices will fall.\n2. Fewer people buy homes.\n3. Therefore, house prices fall. [1,2]"
	output, _ := json.Marshal(map[string]any{"argument": argument, "coverage_anchor": "as_found", "line_mapping": []any{map[string]any{"repaired_line": 1, "from_as_found_lines": []int{1}}}, "dropped": []any{}})
	spec := BuiltinSpec(BuiltinFormalizedRepairOutput)
	spec.Options = json.RawMessage(`{"entailment_gate":true}`)
	result := Validate(t.Context(), spec, string(output))
	if result.Passed {
		t.Fatal("mismatch accepted")
	}
	for _, want := range []string{"verbatim", "\nfirst clause: \"house prices will fall\"", "\nsecond clause: \"house prices fall\"", "\nfirst clause extra: will", "Do not change the argument's content."} {
		if !strings.Contains(result.Message, want) {
			t.Fatalf("missing %q: %s", want, result.Message)
		}
	}
	if hint := repairClauseDiffHint("no cited premise matches conclusion step 3 under any entailment-gate schema"); hint != "" {
		t.Fatal(hint)
	}
}

func TestRepairClauseDiffHints(t *testing.T) {
	for _, tc := range []struct{ mismatch, want string }{
		{`premise 1 antecedent "prices will fall" is not matched verbatim by any cited premise; nearest is step 2 "prices fall"`, "first clause extra: will"},
		{`premise 1 "prices fall" does not match conclusion 3 "prices will fall" verbatim`, "first clause missing: will"},
		{`premise 1 "a a b" does not match conclusion 3 "a b" verbatim`, "first clause extra: a"},
		{`premise 1 "a b" does not match conclusion 3 "b a" verbatim`, "word order differs"},
		{`premise 1 "say \"prices will fall\"" does not match conclusion 3 "say \"prices fall\"" verbatim`, "first clause extra: will"},
	} {
		if hint := repairClauseDiffHint(tc.mismatch); !strings.Contains(hint, tc.want) {
			t.Fatalf("%s: got %s", tc.mismatch, hint)
		}
	}
}

func TestRepairWholeSentenceSuggestionCanFixOverlap(t *testing.T) {
	source := "The revenue from a congestion charge in Manhattan could fund the subway."
	quote := "could fund the subway"
	_, _, err := ValidateSourceQuote(source, source, quote)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("quote the whole source sentence: %q", source)) {
		t.Fatalf("fragment diagnostic: %v", err)
	}
	if _, _, err := ValidateSourceQuote(source, source, source); err != nil {
		t.Fatalf("whole sentence rejected: %v", err)
	}
	// The suggested sentence is context, not an automatic relaxation of overlap.
	if _, _, err := ValidateSourceQuote("When fewer people buy homes, house prices fall.", "If fewer people buy homes, then house prices will fall.", "When fewer people buy homes, house prices fall."); err == nil {
		t.Fatal("unchanged overlap threshold accepted the homes mismatch")
	}
}
