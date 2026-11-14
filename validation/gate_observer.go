package validation

// gateStepObserver is for coverage tests only.
var gateStepObserver func(form string, passed bool)

// observeGateReport reports each step in the final gate report. A passed step
// contributes its Form; a failed step contributes its NearestForm. A nil
// observer returns immediately.
func observeGateReport(report EntailmentGateReport) EntailmentGateReport {
	observe := gateStepObserver
	if observe == nil {
		return report
	}
	for _, step := range report.Steps {
		form := step.NearestForm
		if step.Passed {
			form = step.Form
		}
		observe(form, step.Passed)
	}
	return report
}
