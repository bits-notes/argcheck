package validation

import (
	"strings"
	"testing"
)

func TestNormalizeClauseAndWrapperEquality(t *testing.T) {
	if got := normalizeClause("  ,  Alpha   BETA, .  "); got != "alpha beta, " {
		t.Fatalf("normalizeClause = %q, want %q", got, "alpha beta, ")
	}
	if !equalClauses("It is the case that, it is the case that Alpha.", "alpha") {
		t.Fatal("repeated wrapper stripping did not preserve clause equality")
	}
	if equalClauses("the red box", "the blue box") {
		t.Fatal("clause equality rewrote non-wrapper vocabulary")
	}
}

func TestClauseCandidateEnumeration(t *testing.T) {
	conditionals := conditionalCandidates("If P, then, if Q, then R.")
	if len(conditionals) != 2 {
		t.Fatalf("conditional candidates = %d, want 2", len(conditionals))
	}
	if conditionals[1] != (clausePair{left: "p, then, if q", right: "r"}) {
		t.Fatalf("outer conditional = %#v", conditionals[1])
	}

	conjunctions := conjunctionCandidates("It is the case both that, P, and that Q.")
	found := false
	for _, conjunction := range conjunctions {
		if equalClauses(conjunction.left, "P") && equalClauses(conjunction.right, "Q") {
			found = true
		}
	}
	if !found {
		t.Fatalf("scaffold conjunction candidate missing: %#v", conjunctions)
	}
}

func TestEntailmentSchemaCatalogPositiveAndNegative(t *testing.T) {
	tests := []struct {
		name             string
		form             string
		positive         string
		negative         string
		mismatchContains [2]string
	}{
		{
			name: "modus ponens", form: "modus ponens",
			positive:         "1. If rain falls, then streets get wet.\n2. Rain falls.\n3. Therefore, streets get wet. [1,2]",
			negative:         "1. If rain falls, then streets get wet.\n2. Rain falls.\n3. Therefore, streets get damp. [1,2]",
			mismatchContains: [2]string{"streets get wet", "streets get damp"},
		},
		{
			name: "modus tollens", form: "modus tollens",
			positive:         "1. If rain falls, then streets get wet.\n2. It is not the case that streets get wet.\n3. Therefore, it is not the case that rain falls. [1,2]",
			negative:         "1. If rain falls, then streets get wet.\n2. It is not the case that streets get wet.\n3. Therefore, it is not the case that snow falls. [1,2]",
			mismatchContains: [2]string{"rain falls", "snow falls"},
		},
		{
			name: "hypothetical syllogism", form: "hypothetical syllogism",
			positive:         "1. If rain falls, then streets get wet.\n2. If streets get wet, then traffic slows.\n3. Therefore, if rain falls, then traffic slows. [1,2]",
			negative:         "1. If rain falls, then streets get wet.\n2. If streets get wet, then traffic slows.\n3. Therefore, if rain falls, then traffic stops. [1,2]",
			mismatchContains: [2]string{"traffic slows", "traffic stops"},
		},
		{
			name: "conjunction introduction", form: "conjunction introduction",
			positive:         "1. Rain falls.\n2. Wind blows.\n3. Therefore, rain falls and wind blows. [1,2]",
			negative:         "1. Rain falls.\n2. Wind blows.\n3. Therefore, rain falls and snow falls. [1,2]",
			mismatchContains: [2]string{"wind blows", "snow falls"},
		},
		{
			name: "conjunction elimination", form: "conjunction elimination",
			positive:         "1. Rain falls and wind blows.\n2. Therefore, wind blows. [1]",
			negative:         "1. Rain falls and wind blows.\n2. Therefore, snow falls. [1]",
			mismatchContains: [2]string{"rain falls", "snow falls"},
		},
		{
			name: "disjunction introduction", form: "disjunction introduction",
			positive:         "1. Rain falls.\n2. Therefore, either rain falls or snow falls. [1]",
			negative:         "1. Rain falls.\n2. Therefore, either wind blows or snow falls. [1]",
			mismatchContains: [2]string{"rain falls", "wind blows"},
		},
		{
			name: "disjunctive syllogism", form: "disjunctive syllogism",
			positive:         "1. Rain falls or snow falls.\n2. It is not the case that rain falls.\n3. Therefore, snow falls. [1,2]",
			negative:         "1. Rain falls or snow falls.\n2. It is not the case that rain falls.\n3. Therefore, hail falls. [1,2]",
			mismatchContains: [2]string{"snow falls", "hail falls"},
		},
		{
			name: "double negation", form: "double negation",
			positive:         "1. It is not the case that it is not the case that rain falls.\n2. Therefore, rain falls. [1]",
			negative:         "1. It is not the case that it is not the case that rain falls.\n2. Therefore, snow falls. [1]",
			mismatchContains: [2]string{"rain falls", "snow falls"},
		},
		{
			name: "self-entailment", form: "self-entailment",
			positive:         "1. Rain falls.\n2. Therefore, rain falls. [1]",
			negative:         "1. Rain falls.\n2. Therefore, snow falls. [1]",
			mismatchContains: [2]string{"rain falls", "snow falls"},
		},
		{
			name: "identity transitivity", form: "identity transitivity",
			positive:         "1. Alpha is identical to beta.\n2. Beta is identical to gamma.\n3. Therefore, alpha is identical to gamma. [1,2]",
			negative:         "1. Alpha is identical to beta.\n2. Beta is identical to gamma.\n3. Therefore, alpha is identical to delta. [1,2]",
			mismatchContains: [2]string{"gamma", "delta"},
		},
		{
			name: "universal instantiation", form: "universal instantiation",
			positive:         "1. Every raven is black.\n2. Hugin is raven.\n3. Therefore, Hugin is black. [1,2]",
			negative:         "1. Every raven is black.\n2. Hugin is raven.\n3. Therefore, Hugin is dark. [1,2]",
			mismatchContains: [2]string{"black", "dark"},
		},
	}

	for _, test := range tests {
		t.Run(test.name+" positive", func(t *testing.T) {
			report, result := CheckEntailmentGate(test.positive)
			if !result.Passed || !report.Passed {
				t.Fatalf("gate failed: result=%#v report=%#v", result, report)
			}
			if got := report.Steps[0].Form; got != test.form {
				t.Fatalf("form = %q, want %q", got, test.form)
			}
		})
		t.Run(test.name+" negative", func(t *testing.T) {
			report, result := CheckEntailmentGate(test.negative)
			if result.Passed || report.Passed {
				t.Fatalf("gate unexpectedly passed: result=%#v report=%#v", result, report)
			}
			step := report.Steps[0]
			if step.NearestForm != test.form {
				t.Fatalf("nearest form = %q, want %q; mismatch=%q", step.NearestForm, test.form, step.Mismatch)
			}
			for _, expected := range test.mismatchContains {
				if !strings.Contains(step.Mismatch, expected) {
					t.Fatalf("mismatch %q does not quote %q", step.Mismatch, expected)
				}
			}
		})
	}
}
