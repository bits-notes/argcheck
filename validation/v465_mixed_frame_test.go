package validation

import (
	"strings"
	"testing"
)

func TestV465MixedFrames(t *testing.T) {
	for name, text := range map[string]string{
		"Fitch":    "| 1. Suppose both the lamp is lit and the door is open or the bell rings.\n| 2. Therefore, the lamp is lit. [1]\n3. Therefore, if both the lamp is lit and the door is open or the bell rings, then the lamp is lit. [1-2]",
		"negation": "1. It is not the case that both the lamp is lit and the door is open or the bell rings.\n2. Therefore, it is not the case that both the lamp is lit and the door is open or the bell rings. [1]",
		"modal":    "1. Necessarily, either the lamp is lit or the door is open and the bell rings.\n2. Therefore, necessarily, either the lamp is lit or the door is open and the bell rings. [1]",

		"033":                    "1. Both the lamp is lit and the door is open or the bell rings.\n2. The lamp is lit. [1]",
		"mirror syllogism":       "1. Either the lamp is lit or the door is open and the bell rings.\n2. It is not the case that the lamp is lit.\n3. Therefore, the door is open and the bell rings. [1,2]",
		"mirror commutation":     "1. Either the lamp is lit or the door is open and the bell rings.\n2. Therefore, either the door is open and the bell rings or the lamp is lit. [1]",
		"conclusion":             "1. The lamp is lit.\n2. The door is open or the bell rings.\n3. Therefore, both the lamp is lit and the door is open or the bell rings. [1,2]",
		"conditional antecedent": "1. If both the lamp is lit and the door is open or the bell rings, then the ground is wet.\n2. Both the lamp is lit and the door is open or the bell rings.\n3. Therefore, the ground is wet. [1,2]",
		"conditional consequent": "1. If the ground is wet, then both the lamp is lit and the door is open or the bell rings.\n2. The ground is wet.\n3. Therefore, both the lamp is lit and the door is open or the bell rings. [1,2]",
		"nested":                 "1. Both the ground is wet and either the lamp is lit or the door is open and the bell rings.\n2. Therefore, the ground is wet. [1]",
	} {
		t.Run(name, func(t *testing.T) {
			report, res := CheckEntailmentGate(text, notesOn)
			if report.Passed {
				t.Fatal("certified ambiguous frame")
			}
			found := false
			for _, step := range report.Steps {
				for _, note := range step.Notes {
					found = found || note.ID == "scope.mixed_andor"
				}
			}
			if !found {
				t.Fatalf("missing note: %s (%+v)", res.Message, report.Steps)
			}
			if name == "033" && (report.Steps[0].Refusal == nil || report.Steps[0].Refusal.Hint != "use 'Both the lamp is lit and either the door is open or the bell rings.'") {
				t.Fatalf("missing concrete repair: %+v", report.Steps[0].Refusal)
			}
			// Turning off precisely this strict rule restores the old certificate.
			if r, res := allStrictRules().without("scope.mixed_andor").checkEntailmentGate(text, numberedArgumentOptions{}); !r.Passed {
				t.Fatal(res.Message)
			}
		})
	}
}

func TestV465AllowedFrames(t *testing.T) {
	for _, text := range []string{
		"1. Both the lamp is lit and either the door is open or the bell rings.\n2. Therefore, the lamp is lit. [1]",
		"1. Either the lamp is lit or both the door is open and the bell rings.\n2. It is not the case that the lamp is lit.\n3. Therefore, both the door is open and the bell rings. [1,2]",
		"1. Both the lamp is lit and (the door is open or the bell rings).\n2. Therefore, the lamp is lit. [1]",
		"1. Both the lamp is lit and the door is open.\n2. Therefore, the lamp is lit. [1]",
		"1. Either the lamp is lit or the door is open.\n2. Therefore, either the door is open or the lamp is lit. [1]",
		"1. It is the case both that the lamp is lit, and that the door is open.\n2. Therefore, the lamp is lit. [1]",
	} {
		if r, res := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("%s: %s", text, res.Message)
		}
	}
}

func TestV465MixedFrameScopes(t *testing.T) {
	for _, frame := range []string{
		"Both the lamp is lit and the door is open or the bell rings",
		"Either the lamp is lit or the door is open and the bell rings",
		"Both the door is open or the bell rings and the lamp is lit",
		"Either the door is open and the bell rings or the lamp is lit",
		"Both the ground is wet and either the lamp is lit or the door is open and the bell rings",
		"Either the ground is wet or both the lamp is lit and the door is open or the bell rings",
	} {
		for _, wrapper := range []string{"%s", "It is not the case that %s", "Necessarily, %s", "Possibly %s", "If %s, then the ground is wet", "If the ground is wet, then %s", "Every x is such that %s", "(%s)"} {
			s := strings.Replace(wrapper, "%s", frame, 1)
			if mixedFrameHint(s) == "" {
				t.Errorf("missed %s", s)
			}
		}
	}
	for _, s := range []string{
		"Both the lamp is lit and either the door is open or the bell rings",
		"Either both the door is open and the bell rings or the lamp is lit",
		"Both (the door is open or the bell rings) and the lamp is lit",
		"It is the case both that the lamp is lit, and that the door is open",
		"Both A and B", "Either A or B",
	} {
		if hint := mixedFrameHint(s); hint != "" {
			t.Errorf("%s: %s", s, hint)
		}
	}
	want := "Both the lamp is lit and either the door is open or the bell rings."
	if hint := mixedFrameHint("Both the lamp is lit and the door is open or the bell rings"); !strings.Contains(hint, want) {
		t.Fatal(hint)
	}
	want = "Either the lamp is lit or both the door is open and the bell rings."
	if hint := mixedFrameHint("Either the lamp is lit or the door is open and the bell rings"); !strings.Contains(hint, want) {
		t.Fatal(hint)
	}
}

func TestV465MixedFramePhrasePolicy(t *testing.T) {
	// The framed check must make the same coordination decision as the
	// pre-existing unframed check, including its conservative nominal cases.
	for _, pair := range []struct{ frame, body string }{
		{"Both the lamp is lit and the salt or pepper is dry", "the lamp is lit and the salt or pepper is dry"},
		{"Either the lamp is lit or the salt and pepper is dry", "the lamp is lit or the salt and pepper is dry"},
		{"Both the lamp is lit and every guard who sings or dances is happy", "the lamp is lit and every guard who sings or dances is happy"},
		{"Either the lamp is lit or every guard who sings and dances is happy", "the lamp is lit or every guard who sings and dances is happy"},
	} {
		want := groupingProblem(pair.body) != "" || allStrictRules().scopeExpressionFailure(pair.body) == mixedScopeHint
		if got := mixedFrameHint(pair.frame) != ""; got != want {
			t.Errorf("%s: refused=%v, unframed=%v", pair.frame, got, want)
		}
	}
}

func TestV465MixedFrameFOL(t *testing.T) {
	// The explicit-variable route already refuses this on v46.4. Keep its
	// original mismatch rather than replacing it with the new frame hint.
	text := "1. Every x is such that both x is red and x is blue or x is green.\n2. Therefore, both Ann is red and Ann is blue or Ann is green. [1]"
	r, res := CheckEntailmentGate(text)
	if r.Passed || res.Message != "step 2 failed entailment_gate: nearest form FOL universal instantiation; "+mixedScopeHint {
		t.Fatal(res.Message)
	}
}

func TestV465GroupedSchemas(t *testing.T) {
	const p = "the rain falls"
	const q = "the snow falls"
	const r = "the street is wet"
	and := func(a, b string) string { return "both " + a + " and " + b }
	or := func(a, b string) string { return "either " + a + " or " + b }
	for _, tc := range []struct{ name, form, premise, conclusion string }{
		{"and forward", "distribution", and(p, or(q, r)), or(and(p, q), and(p, r))},
		{"and reverse", "distribution", or(and(p, q), and(p, r)), and(p, or(q, r))},
		{"or forward", "distribution", or(p, and(q, r)), and(or(p, q), or(p, r))},
		{"or reverse", "distribution", and(or(p, q), or(p, r)), or(p, and(q, r))},
		{"disjunctive", "absorption", or(p, and(p, q)), p},
	} {
		t.Run(tc.form+"/"+tc.name, func(t *testing.T) {
			argument := "1. " + tc.premise + ".\n2. Therefore, " + tc.conclusion + ". [1]"
			report, res := CheckEntailmentGate(argument)
			if !report.Passed || report.Steps[0].Form != tc.form {
				t.Fatalf("grouped: %s; %+v", res.Message, report.Steps)
			}
			scaffold := strings.ReplaceAll(strings.ReplaceAll(argument, "both ", "it is the case both that "), " and ", ", and that ")
			if report, res := CheckEntailmentGate(scaffold); !report.Passed || report.Steps[0].Form != tc.form {
				t.Fatalf("scaffold: %s; %+v", res.Message, report.Steps)
			}
			unmarked := strings.ReplaceAll(strings.ReplaceAll(argument, "both ", ""), "either ", "")
			if report, _ := CheckEntailmentGate(unmarked); report.Passed {
				t.Fatal("certified unmarked mixture", unmarked)
			}
		})
	}
}
