package validation

import (
	"strings"
	"testing"
)

func assertStrictRefusal(t *testing.T, argument, rule, hint string) {
	t.Helper()
	report, result := CheckEntailmentGate(argument)
	if report.Passed || result.Passed {
		t.Fatalf("%s refusal unexpectedly certified:\n%s", rule, argument)
	}
	if !strings.Contains(strings.ToLower(result.Message), strings.ToLower(hint)) {
		t.Fatalf("%s missing rewrite %q: %s\n%s", rule, hint, result.Message, argument)
	}
}

type strictRefusal struct{ rule, hint string }

func listedStrictRefusal(t *testing.T, argument string, refused map[string]strictRefusal) bool {
	t.Helper()
	r, ok := refused[t.Name()]
	if !ok {
		return false
	}
	assertStrictRefusal(t, argument, r.rule, r.hint)
	return true
}

func listedStrictASTRefusal(t *testing.T, argument NumberedArgument, refused map[string]strictRefusal) bool {
	t.Helper()
	r, ok := refused[t.Name()]
	if !ok {
		return false
	}
	report := GateNumberedArgument(argument)
	if report.Passed {
		t.Fatalf("%s refusal unexpectedly certified", r.rule)
	}
	for _, step := range report.Steps {
		if !step.Passed && strings.Contains(strings.ToLower(step.Mismatch), strings.ToLower(r.hint)) {
			return true
		}
	}
	t.Fatalf("%s missing rewrite %q: %+v", r.rule, r.hint, report)
	return true
}

func runV11Strict(t *testing.T, cases []v11Case, refused map[string]strictRefusal) {
	t.Helper()
	var accepted []v11Case
	for _, tc := range cases {
		key := t.Name() + "/" + strings.ReplaceAll(tc.name, " ", "_")
		if r, ok := refused[key]; ok {
			t.Run(tc.name, func(t *testing.T) { assertStrictRefusal(t, tc.argument, r.rule, r.hint) })
		} else {
			accepted = append(accepted, tc)
		}
	}
	runV11Cases(t, accepted)
}
