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
	if code != exitValid {
		t.Fatalf("exit %d, output:\n%s", code, out)
	}
	for _, name := range []string{
		"entailment gate catalog v14", "modus tollens", "destructive dilemma",
		"contraposition", "De Morgan", "commutation", "exportation", "importation",
		"material implication", "distribution", "absorption", "double negation introduction",
		"biconditional modus tollens", "biconditional chaining", "biconditional commutation",
		"negated conditional", "identity symmetry", "quantifier negation", "conversion",
		"universal modus tollens", "possibility introduction",
	} {
		if !strings.Contains(out, name) {
			t.Errorf("--forms missing %q:\n%s", name, out)
		}
	}
}

func TestArgcheckFitch(t *testing.T) {
	const text = "1. If it rains, then it snows.\n| 2. Suppose it rains.\n| 3. Therefore, it snows. [1,2]\n4. Therefore, if it rains, then it snows. [2-3]"
	code, out := runArgcheck(t, text)
	if code != 0 || !strings.Contains(out, "| 2. Suppose it rains.") || !strings.Contains(out, "step 4 [2-3]: ok: conditional proof") {
		t.Fatalf("%d %s", code, out)
	}
	code, out = runArgcheck(t, text, "--json")
	if code != 0 || !strings.Contains(out, `"depth": 1`) || !strings.Contains(out, `"ranges":`) {
		t.Fatalf("%d %s", code, out)
	}
	code, out = runArgcheck(t, strings.Replace(text, "[2-3]", "[3]", 1))
	if code != 1 || !strings.Contains(out, "step 4 cites 3, which is inside the closed subproof 2-3") {
		t.Fatalf("%d %s", code, out)
	}
	code, out = runArgcheck(t, "", "--forms")
	for _, name := range []string{"conditional proof", "reductio", "contradiction introduction", "proof by cases", "universal generalization", "existential elimination"} {
		if code != 0 || !strings.Contains(out, name) {
			t.Fatalf("missing %s: %s", name, out)
		}
	}
}
