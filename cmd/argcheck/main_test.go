package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runArgcheck(t *testing.T, input string, args ...string) (int, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(args, strings.NewReader(input), &stdout, &stderr)
	return code, stdout.String() + stderr.String()
}

func TestArgcheckValid(t *testing.T) {
	code, out := runArgcheck(t, "1. If it rains, then the street is wet.\n2. It rains.\n3. Therefore, the street is wet. [1,2]\n")
	if code != exitValid || !strings.Contains(out, "step 3 [1,2]: ok: modus ponens") {
		t.Fatalf("exit %d, output:\n%s", code, out)
	}
}

func TestArgcheckReportsFailingStep(t *testing.T) {
	code, out := runArgcheck(t, "1. If it rains, then the street is wet.\n2. The street is wet.\n3. Therefore, it rains. [1,2]")
	if code != exitInvalid || !strings.Contains(out, "step 3 [1,2]: FAIL: nearest form") {
		t.Fatalf("exit %d, output:\n%s", code, out)
	}
}

func TestArgcheckHintsInertCitation(t *testing.T) {
	input := "1. If it rains, then the street is wet.\n2. It rains.\n3. The sky is grey.\n4. Therefore, the street is wet. [1,2,3]"
	code, out := runArgcheck(t, input)
	if code != exitInvalid || !strings.Contains(out, "modus ponens passes on [1,2]; citation(s) [3] do no work") {
		t.Fatalf("exit %d, output:\n%s", code, out)
	}
}

func TestArgcheckSyntaxFailureJSON(t *testing.T) {
	code, out := runArgcheck(t, "1. It rains.\n2. The street is wet [1]", "--json")
	if code != exitInvalid {
		t.Fatalf("exit %d, output:\n%s", code, out)
	}
	var v verdict
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if v.Valid || v.Syntax.Passed || v.Entailment != nil || v.Backend != "local" || v.Catalog == "" {
		t.Fatalf("unexpected verdict: %+v", v)
	}
}

func TestArgcheckEmptyInput(t *testing.T) {
	if code, _ := runArgcheck(t, "  \n"); code != exitError {
		t.Fatalf("exit %d, want %d", code, exitError)
	}
}

func TestArgcheckForms(t *testing.T) {
	code, out := runArgcheck(t, "", "--forms")
	if code != exitValid || !strings.Contains(out, "modus tollens") || !strings.Contains(out, "destructive dilemma") {
		t.Fatalf("exit %d, output:\n%s", code, out)
	}
}
