package validation

import "testing"

func TestRelationalScopeConversion(t *testing.T) {
	// Like universals commute on every reading, so the gate accepts the step
	// as this form. Pulling an existential wide is the quantifier-shift
	// fallacy: some readings entail, and the failure names this form.
	positive := relationArgument(
		[]string{"Every student admires every teacher"},
		"every teacher is such that every student admires that teacher",
	)
	negative := relationArgument(
		[]string{"Every student admires some teacher"},
		"some teacher is such that every student admires that teacher",
	)

	t.Run("positive", func(t *testing.T) {
		report, result := CheckEntailmentGate(positive)
		if !result.Passed || !report.Passed {
			t.Fatalf("gate failed: result=%#v report=%#v", result, report)
		}
		if len(report.Steps) != 1 {
			t.Fatalf("steps = %+v", report.Steps)
		}
		if got := report.Steps[0].Form; got != "relational scope conversion" {
			t.Fatalf("form = %q, want %q", got, "relational scope conversion")
		}
	})

	t.Run("negative", func(t *testing.T) {
		report, result := CheckEntailmentGate(negative)
		if result.Passed || report.Passed {
			t.Fatalf("gate unexpectedly passed: result=%#v report=%#v", result, report)
		}
		if len(report.Steps) != 1 {
			t.Fatalf("steps = %+v", report.Steps)
		}
		step := report.Steps[0]
		if step.NearestForm != "relational scope conversion" {
			t.Fatalf("nearest form = %q, want %q; mismatch=%q", step.NearestForm, "relational scope conversion", step.Mismatch)
		}
	})
}
