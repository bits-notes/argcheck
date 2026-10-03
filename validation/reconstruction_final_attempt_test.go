package validation

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestReconstructionFinalAttemptSalvage(t *testing.T) {
	source := "A source sentence."
	prompt := "<source_material>" + source + "</source_material>"
	reply := salvageReply(t, "1. A.\n2. Therefore, B. [1]")
	options := json.RawMessage(`{"entailment_gate":true,"salvage_final_attempt":true}`)
	baseline := ValidateFormalizedReconstructionOutputWithPrompt(t.Context(), reply, prompt, options)
	for _, tc := range []struct {
		name         string
		attempt, max int
		options      json.RawMessage
		pass         bool
	}{
		{"final", 3, 3, options, true},
		{"lower budget", 1, 1, options, true},
		{"after provider retries", 4, 3, options, true},
		{"first", 1, 3, options, false},
		{"penultimate", 2, 3, options, false},
		{"unknown", 0, 0, options, false},
		{"off", 3, 3, json.RawMessage(`{"entailment_gate":true}`), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), reply, prompt, tc.attempt, tc.max, tc.options)
			if result.Passed != tc.pass {
				t.Fatalf("%+v", result)
			}
			if !tc.pass {
				if !reflect.DeepEqual(result, baseline) {
					t.Fatalf("failure changed: %+v vs %+v", result, baseline)
				}
				return
			}
			expected, err := SalvageReconstruction(reply, source)
			if err != nil {
				t.Fatal(err)
			}
			normalized, _ := json.Marshal(expected.Reconstruction)
			if string(result.ParsedOutput) != string(normalized) {
				t.Fatalf("normalized: %s", result.ParsedOutput)
			}
			want := map[string]any{"assumed_lines": []int{2}, "failing_steps": []int{2}}
			if !reflect.DeepEqual(result.Details["salvaged"], want) {
				t.Fatalf("details: %+v", result.Details)
			}
			if got := ValidateNormalizedReconstructionOutput(t.Context(), string(normalized), options); !got.Passed {
				t.Fatalf("downstream: %+v", got)
			}
			if got := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), string(normalized), prompt, 3, 3, options); got.Passed {
				t.Fatal("writer supplied assumptions accepted")
			}
		})
	}
}

func TestReconstructionFinalAttemptPreservesFailures(t *testing.T) {
	prompt := "<source_material>A source sentence.</source_material>"
	reply := salvageReply(t, "1. A.\n2. Therefore, B. [1]")
	options := json.RawMessage(`{"entailment_gate":true,"salvage_final_attempt":true}`)
	for _, raw := range []string{
		"not JSON",
		strings.Replace(reply, `"line":1`, `"line":2`, 1),
		strings.Replace(reply, `"sentences":[1]`, `"sentences":[2]`, 1),
		strings.Replace(reply, `"sentences":[1]`, `"sentences":null`, 1),
		strings.ReplaceAll(reply, `"sentences":[1]`, `"sentences":[]`),
		strings.Replace(reply, `"line":1`, `"line":1,"assumed":true`, 1),
		strings.Replace(reply, `"line":1`, `"line":1,"assumed":false`, 1),
		strings.Replace(reply, `"line":1`, `"line":1,"Assumed":null`, 1),
		salvageReply(t, "1. A.\n2. Therefore, B. [1]\n3. Therefore, C. [2]"), // too many bridges
	} {
		before := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), raw, prompt, 2, 3, options)
		final := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), raw, prompt, 3, 3, options)
		if before.Passed || final.Passed || !reflect.DeepEqual(before, final) {
			t.Fatalf("failure changed: %+v vs %+v", before, final)
		}
	}
}

func TestReconstructionFinalAttemptAssociatedConditional(t *testing.T) {
	reply := salvageReply(t, "1. A.\n2. B.\n3. If A and B, then C.\n4. Therefore, C. [1,2,3]\n5. Therefore, C or D. [4]", 3)
	options := json.RawMessage(`{"entailment_gate":true,"salvage_final_attempt":true}`)
	prompt := "<source_material>A source sentence.</source_material>"
	before := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), reply, prompt, 2, 3, options)
	if before.Passed || !strings.Contains(before.Message, "associated conditional") {
		t.Fatalf("%+v", before)
	}
	final := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), reply, prompt, 3, 3, options)
	if !final.Passed || final.Details["salvaged"] == nil {
		t.Fatalf("%+v", final)
	}
}

func TestReconstructionFinalAttemptValidReplyUnchanged(t *testing.T) {
	reply := salvageReply(t, "1. A.\n2. Therefore, A or B. [1]")
	prompt := "<source_material>A source sentence.</source_material>"
	options := json.RawMessage(`{"entailment_gate":true,"salvage_final_attempt":true}`)
	before := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), reply, prompt, 1, 3, options)
	final := ValidateFormalizedReconstructionOutputForAttempt(t.Context(), reply, prompt, 3, 3, options)
	if !final.Passed || !reflect.DeepEqual(before, final) || final.Details["salvaged"] != nil {
		t.Fatalf("%+v", final)
	}
}

func TestReconstructionSalvageOptionParsing(t *testing.T) {
	for _, raw := range []string{`{"salvage_final_attempt":"true"}`, `{"salvage_final_attempt":true,"unknown":true}`} {
		spec := BuiltinSpec(BuiltinFormalizedReconstructionOutput)
		spec.Options = json.RawMessage(raw)
		if _, err := Compile(spec); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestReconstructionSalvageOnNoProgress(t *testing.T) {
	source := "A source sentence."
	prompt := "<source_material>" + source + "</source_material>"
	reply := salvageReply(t, "1. A.\n2. Therefore, B. [1]")
	on := json.RawMessage(`{"entailment_gate":true,"salvage_final_attempt":true,"salvage_on_no_progress":true}`)
	first := ValidateFormalizedReconstructionOutputForRetry(t.Context(), reply, prompt, 1, 3, "", on)
	if first.Passed {
		t.Fatal("first attempt salvaged")
	}
	for _, tc := range []struct {
		name     string
		attempt  int
		previous string
		options  json.RawMessage
		pass     bool
	}{
		{"same failure on retry", 2, first.Message, on, true},
		{"different step", 2, strings.Replace(first.Message, "step 2", "step 3", 1), on, false},
		{"different stage", 2, "step 2 supplies an associated conditional (\"If A, then B\")", on, false},
		{"no previous", 2, "", on, false},
		{"first attempt", 1, first.Message, on, false},
		{"option off", 2, first.Message, json.RawMessage(`{"entailment_gate":true,"salvage_final_attempt":true}`), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateFormalizedReconstructionOutputForRetry(t.Context(), reply, prompt, tc.attempt, 3, tc.previous, tc.options)
			if got.Passed != tc.pass {
				t.Fatalf("passed=%v: %s", got.Passed, got.Message)
			}
			if tc.pass {
				salvaged, _ := got.Details["salvaged"].(map[string]any)
				if salvaged["reason"] != "no_progress" {
					t.Fatalf("missing no_progress provenance: %v", got.Details["salvaged"])
				}
			}
		})
	}
}
