package validation

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRefusalDefaultDoesNotRetry(t *testing.T) {
	var retries atomic.Int64
	strictRetryCounter.Store(&retries)
	t.Cleanup(func() { strictRetryCounter.Store(nil) })
	text := "1. If he sings, then the bell rings.\n2. He sings.\n3. Therefore, the bell rings. [1,2]"
	report, result := CheckEntailmentGate(text)
	if result.Passed || report.Steps[0].Refusal.RuleSource != "fired" {
		t.Fatalf("expected fired refusal: %+v", report)
	}
	Validate(context.Background(), BuiltinSpec(BuiltinEntailmentGate), text)
	argument, parsed := ParseNumberedArgument(text, numberedArgumentOptions{})
	if !parsed.Passed {
		t.Fatal(parsed.Message)
	}
	GateNumberedArgument(argument)
	if got := retries.Load(); got != 0 {
		t.Fatalf("default reporting ran %d diagnostic retries", got)
	}
	report, _ = CheckEntailmentGate(text, notesOn)
	if retries.Load() == 0 {
		t.Fatal("counter did not observe opt-in retries")
	}
	if report.Steps[0].Refusal.RuleSource != "note" {
		t.Fatalf("expected proven note: %+v", report.Steps[0])
	}
}

func TestRefusalRuleSources(t *testing.T) {
	for _, tc := range []struct{ text, source string }{
		{"1. Alice sings.\n2. Therefore, Bob dances. [1]", "nearest_form"},
		{"1. Alice sings, Bob dances and Carol runs or Dave walks.\n2. Therefore, Alice sings. [1]", "syntax"},
	} {
		report, result := CheckEntailmentGate(tc.text)
		if result.Passed || len(report.Steps) == 0 || report.Steps[0].Refusal == nil || report.Steps[0].Refusal.RuleSource != tc.source {
			t.Fatalf("want %s: %+v / %+v", tc.source, report, result)
		}
	}
	// Selection is catalog-ordered, with the main mismatch taking precedence.
	fired := &firedRules{ids: map[string]bool{strictNotes[0].ID: true, strictNotes[1].ID: true}}
	steps := map[int]NumberedArgumentStep{1: {Number: 1, Statement: "Alice sings."}}
	failed := EntailmentGateStepReport{Step: 1, fired: fired}
	for i := 0; i < 10; i++ {
		r := refusalFor(failed, steps)
		if r.Rule != strictNotes[0].ID || r.Hint != strings.Join(strictNotes[0].Fix, " ") || r.RuleSource != "fired" {
			t.Fatalf("%+v", r)
		}
	}
	failed.Mismatch = strictNotes[1].ID + ": refused"
	if r := refusalFor(failed, steps); r.Rule != strictNotes[1].ID {
		t.Fatalf("%+v", r)
	}
	failed.Notes = []StrictNote{strictNotes[0]}
	if r := refusalFor(failed, steps); r.Rule != strictNotes[0].ID || r.RuleSource != "note" {
		t.Fatalf("%+v", r)
	}
}
