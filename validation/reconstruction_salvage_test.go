package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func salvageReply(t *testing.T, argument string, supplied ...int) string {
	t.Helper()
	parsed, result := ParseNumberedArgumentAllowingNoConclusions(argument)
	if !result.Passed {
		t.Fatal(result.Message)
	}
	output := FormalizedReconstructionOutput{Argument: argument}
	for _, step := range parsed.Steps {
		sentences := []int{1}
		for _, n := range supplied {
			if n == step.Number {
				sentences = []int{}
			}
		}
		output.Lines = append(output.Lines, ReconstructionLine{Line: step.Number, Sentences: sentences})
	}
	raw, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestSalvageReconstructionBridgeAndRenumber(t *testing.T) {
	argument := "1. If A, then B.\n2. A.\n3. Therefore, C. [1,2]\n4. Therefore, C or D. [3]"
	got, err := SalvageReconstruction("Here it is:\n```json\n"+salvageReply(t, argument)+"\n```", "A source sentence.")
	if err != nil {
		t.Fatal(err)
	}
	want := "1. If A, then B.\n2. A.\n3. If If A, then B and a, then C.\n4. Therefore, C. [1,2,3]\n5. Therefore, C or D. [4]"
	if got.Reconstruction.Argument != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got.Reconstruction.Argument, want)
	}
	if !reflect.DeepEqual(got.AssumedLines, []int{3}) || !reflect.DeepEqual(got.FailingSteps, []int{3}) {
		t.Fatalf("provenance: %#v", got)
	}
	if !got.Reconstruction.Lines[2].Assumed || got.Reconstruction.Lines[2].Sentences == nil || len(got.Reconstruction.Lines[2].Sentences) != 0 {
		t.Fatal("missing assumption metadata")
	}
	if report, _ := CheckEntailmentGate(got.Reconstruction.Argument); !report.Passed {
		t.Fatal("gate failed")
	}
}

func TestSalvageReconstructionRemovesPadding(t *testing.T) {
	argument := "1. A.\n2. B.\n3. If A and B, then C.\n4. Therefore, C. [1,2,3]\n5. Therefore, C or D. [4]"
	got, err := SalvageReconstruction(salvageReply(t, argument, 3), "A source sentence.")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.AssumedLines, []int{3}) || !reflect.DeepEqual(got.FailingSteps, []int{4}) {
		t.Fatalf("provenance: %#v", got)
	}
	if strings.Count(got.Reconstruction.Argument, "If a and B, then C") != 1 || !got.Reconstruction.Lines[2].Assumed {
		t.Fatalf("padding not replaced: %#v", got)
	}
}

func TestSalvageReconstructionLimits(t *testing.T) {
	// Four failures out of eight conclusions: the absolute limit must reject it.
	lines := []string{"1. A."}
	for i := 0; i < 4; i++ {
		n := 2 + i*2
		lines = append(lines, fmt.Sprintf("%d. Therefore, B%d. [%d]", n, i, n-1), fmt.Sprintf("%d. Therefore, B%d or D. [%d]", n+1, i, n))
	}
	for _, argument := range []string{strings.Join(lines, "\n"), "1. A.\n2. Therefore, B. [1]"} {
		if _, err := SalvageReconstruction(salvageReply(t, argument), "A source sentence."); err == nil || !strings.Contains(err.Error(), "too many") {
			t.Fatalf("expected bridge limit, got %v", err)
		}
	}
}

func TestSalvageReconstructionRequiresStructure(t *testing.T) {
	good := salvageReply(t, "1. A.\n2. Therefore, A or B. [1]")
	for _, reply := range []string{
		strings.Replace(good, `"sentences":[1]`, `"sentences":[2]`, 1),
		strings.ReplaceAll(good, `"sentences":[1]`, `"sentences":[]`),
		strings.Replace(good, `"line":1`, `"line":2`, 1),
		strings.Replace(good, `"line":1`, `"line":1,"assumed":true`, 1),
	} {
		if _, err := SalvageReconstruction(reply, "A source sentence."); err == nil {
			t.Fatalf("accepted %s", reply)
		}
	}
}

func TestReconstructionWriterCannotSetAssumed(t *testing.T) {
	reply := salvageReply(t, "1. A.\n2. Therefore, A or B. [1]")
	for _, value := range []string{"true", "false", "null"} {
		raw := strings.Replace(reply, `"line":1`, `"line":1,"assumed":`+value, 1)
		result := ValidateFormalizedReconstructionSource(t.Context(), raw, "A source sentence.")
		if result.Passed || !strings.Contains(result.Message, "assumed") {
			t.Fatalf("accepted writer assumption: %#v", result)
		}
	}
}

func TestRawReplyDetailsRetainsWholeLongReply(t *testing.T) {
	for _, count := range []int{4001, 64000, 64001} {
		raw := strings.Repeat("界", count)
		details := RawReplyDetails(raw)
		if details["raw_reply_chars"] != count || details["raw_reply_excerpt"] != strings.Repeat("界", min(count, 64000)) {
			t.Fatalf("incorrect capture at %d", count)
		}
	}
}

func TestSalvageReconstructionMultipleInsertions(t *testing.T) {
	argument := "1. A.\n2. Therefore, B. [1]\n3. Therefore, B or C. [2]\n4. Therefore, D. [3]\n5. Therefore, D or E. [4]"
	got, err := SalvageReconstruction(salvageReply(t, argument), "A source sentence.")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.AssumedLines, []int{2, 5}) || !reflect.DeepEqual(got.FailingSteps, []int{2, 4}) {
		t.Fatalf("provenance: %#v", got)
	}
	if !strings.Contains(got.Reconstruction.Argument, "6. Therefore, D. [4,5]\n7. Therefore, D or E. [6]") {
		t.Fatalf("citations: %s", got.Reconstruction.Argument)
	}
}

func TestSalvageReconstructionRemovesEveryPaddingPremise(t *testing.T) {
	argument := "1. A.\n2. B.\n3. If A and B, then C.\n4. Therefore, C. [1,2,3]\n5. Therefore, C or D. [4]\n6. If C and C or D, then E.\n7. Therefore, E. [4,5,6]\n8. Therefore, E or F. [7]"
	got, err := SalvageReconstruction(salvageReply(t, argument, 3, 6), "A source sentence.")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.AssumedLines, []int{3, 6}) || !reflect.DeepEqual(got.FailingSteps, []int{4, 7}) {
		t.Fatalf("provenance: %#v", got)
	}
}
