package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		"entailment gate catalog v29", "modus tollens", "destructive dilemma",
		"contraposition", "De Morgan", "commutation", "exportation", "importation",
		"material implication", "distribution", "absorption", "double negation introduction",
		"biconditional modus tollens", "biconditional chaining", "biconditional commutation",
		"negated conditional", "identity symmetry", "quantifier negation", "conversion",
		"universal modus tollens", "possibility introduction",
		"identity reflexivity", "modal duality", "modal distribution",
		"counterfactual modus ponens", "counterfactual modus tollens", "counterfactual to material",
		"counterfactual consequent conjunction", "counterfactual consequent conjunction elimination",
		"counterfactual consequent disjunction introduction", "counterfactual dilemma", "counterfactual reflexivity",
		"property abstraction", "property concretion", "property instantiation", "property existential",
		"shared property", "indiscernibility of identicals", "property generalization",
		"property instantiation (arbitrary property)", "distinguishing property",
		"distinguishing property nonidentity", "distinguishing property non-inclusion",
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

func TestArgcheckRelationalAmbiguity(t *testing.T) {
	const input = "1. Every student admires some teacher.\n2. Therefore, Some teacher is such that every student admires that teacher. [1]"
	for _, args := range [][]string{nil, {"--json"}} {
		code, out := runArgcheck(t, input, args...)
		if code != exitInvalid || !strings.Contains(out, "has two scope readings") || !strings.Contains(out, "some teacher is admired by every student") {
			t.Fatalf("%d %s", code, out)
		}
		if len(args) > 0 {
			var v verdict
			if err := json.Unmarshal([]byte(out), &v); err != nil {
				t.Fatal(err)
			}
			steps := v.Entailment.Details["steps"].([]any)
			if note, ok := steps[0].(map[string]any)["ambiguity_note"].(string); !ok || !strings.Contains(note, "Some teacher is such that") {
				t.Fatalf("missing note: %s", out)
			}
		}
	}
}

func TestArgcheckConsistentClauseScope(t *testing.T) {
	for _, input := range []string{
		"1. Every student admires some teacher.\n2. Therefore, every student admires some teacher. [1]",
		"1. If every student admires some teacher, then the school thrives.\n2. Every student admires some teacher.\n3. Therefore, the school thrives. [1,2]",
	} {
		for _, args := range [][]string{nil, {"--json"}} {
			code, out := runArgcheck(t, input, args...)
			if code != exitValid || strings.Contains(out, "ambiguity_note") {
				t.Fatalf("%d %s", code, out)
			}
		}
	}
}

func TestArgcheckCounterfactualNotes(t *testing.T) {
	for _, tc := range []struct{ argument, note string }{
		{"1. If the match had been struck, then the match would light.\n2. The match was struck.\n3. Therefore, the match lights. [1,2]", "visible bridging premise"},
		{"1. If the match had been struck, then the match would have lit.\n2. Therefore, if the match had not lit, then the match would not have been struck. [1]", "contraposition"},
	} {
		for _, jsonOutput := range []bool{false, true} {
			var args []string
			if jsonOutput {
				args = []string{"--json"}
			}
			code, out := runArgcheck(t, tc.argument, args...)
			if code != exitInvalid {
				t.Fatalf("exit %d: %s", code, out)
			}
			if jsonOutput {
				var v verdict
				if err := json.Unmarshal([]byte(out), &v); err != nil {
					t.Fatal(err)
				}
				steps := stepsOf(v.Entailment)
				if len(steps) != 1 || !strings.Contains(steps[0].Mismatch, tc.note) {
					t.Fatalf("missing JSON note %q: %s", tc.note, out)
				}
			} else if !strings.Contains(out, tc.note) {
				t.Fatalf("missing plain note %q: %s", tc.note, out)
			}
		}
	}
}

func TestArgcheckVersion(t *testing.T) {
	old := commit
	t.Cleanup(func() { commit = old })
	commit = "v28-test-commit"
	code, out := runArgcheck(t, "", "--version")
	if code != exitValid || !strings.Contains(out, "catalog v29") || !strings.Contains(out, "argcheck@v28-test-commit") {
		t.Fatalf("%d: %s", code, out)
	}
	if code, _ := runArgcheck(t, "", "--version-typo"); code != exitError {
		t.Fatalf("unknown flag returned %d", code)
	}
	commit = ""
	if got := localVersion(); !strings.HasPrefix(got, "argcheck@") {
		t.Fatal(got)
	}
}

func TestArgcheckVerdictLineFirstWithSubproof(t *testing.T) {
	const text = "1. If it rains, then it snows.\n| 2. Suppose it rains.\n| 3. Therefore, it snows. [1,2]\n4. Therefore, if it rains, then it snows. [2-3]"
	code, out := runArgcheck(t, text)
	if code != 0 || !strings.HasPrefix(out, "VALID  (") || !strings.Contains(out, "| 2. Suppose it rains.") {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestArgcheckQuiet(t *testing.T) {
	const valid = "1. If it rains, then it snows.\n| 2. Suppose it rains.\n| 3. Therefore, it snows. [1,2]\n4. Therefore, if it rains, then it snows. [2-3]"
	code, out := runArgcheck(t, valid, "--quiet")
	if code != 0 || out != "VALID\n" {
		t.Fatalf("valid: code=%d out=%q", code, out)
	}
	code, out = runArgcheck(t, "1. If it rains, then it snows.\n2. It snows.\n3. Therefore, it rains. [1,2]", "--quiet")
	if code != 1 || out != "INVALID\n" {
		t.Fatalf("invalid: code=%d out=%q", code, out)
	}
}

func TestArgcheckRemoteRequiresEnv(t *testing.T) {
	const valid = "1. If it rains, then the street is wet.\n2. It rains.\n3. Therefore, the street is wet. [1,2]"
	t.Setenv("ARGCHECK_REMOTE_URL", "")
	t.Setenv("ARGCHECK_REMOTE_TOOL", "")
	for _, flagName := range []string{"--prod", "--both"} {
		code, out := runArgcheck(t, valid, flagName)
		if code != exitError || !strings.Contains(out, "ARGCHECK_REMOTE_URL") {
			t.Fatalf("%s without URL: exit %d, output:\n%s", flagName, code, out)
		}
	}
	t.Setenv("ARGCHECK_REMOTE_URL", "http://127.0.0.1:1/mcp")
	for _, flagName := range []string{"--prod", "--both"} {
		code, out := runArgcheck(t, valid, flagName)
		if code != exitError || !strings.Contains(out, "ARGCHECK_REMOTE_TOOL") {
			t.Fatalf("%s without tool: exit %d, output:\n%s", flagName, code, out)
		}
	}
}

func TestArgcheckRemoteAndBoth(t *testing.T) {
	const valid = "1. If it rains, then the street is wet.\n2. It rains.\n3. Therefore, the street is wet. [1,2]"
	flip := false
	seenTool := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
			Params struct {
				Name      string            `json:"name"`
				Arguments map[string]string `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		seenTool = req.Params.Name
		v := checkLocal(req.Params.Arguments["argument"])
		if flip {
			v.Valid = !v.Valid
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]any{
				"structuredContent": v,
			},
		})
	}))
	t.Cleanup(server.Close)

	t.Setenv("ARGCHECK_REMOTE_URL", server.URL)
	t.Setenv("ARGCHECK_REMOTE_TOOL", "custom.validate")

	code, out := runArgcheck(t, valid, "--prod")
	if code != exitValid || seenTool != "custom.validate" || !strings.Contains(out, "step 3 [1,2]: ok: modus ponens") {
		t.Fatalf("--prod: code=%d tool=%q out:\n%s", code, seenTool, out)
	}
	code, out = runArgcheck(t, valid, "--both")
	if code != exitValid {
		t.Fatalf("--both agree: code=%d out:\n%s", code, out)
	}

	flip = true
	code, out = runArgcheck(t, valid, "--both")
	if code != exitDivergent || !strings.Contains(out, "DIVERGENT") {
		t.Fatalf("--both diverge: code=%d out:\n%s", code, out)
	}
	code, out = runArgcheck(t, valid, "--both", "--quiet")
	if code != exitDivergent || out != "DIVERGENT\n" {
		t.Fatalf("--both --quiet diverge: code=%d out=%q", code, out)
	}
}
