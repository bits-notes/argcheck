package validation

// BridgeCompressionEntry records whether one bracket-bridge premise is
// removable: whether the conclusion that cites it still gate-certifies when
// cited directly on the remaining premises. Removable bridges are pure
// regimentation overhead; bridges that resist removal carry substantive
// linking content.
type BridgeCompressionEntry struct {
	BridgeStep int    `json:"bridge_step"`
	Conclusion int    `json:"conclusion_step"`
	Removable  bool   `json:"removable"`
	Form       string `json:"form,omitempty"`
	Mismatch   string `json:"mismatch,omitempty"`
}

// CheckBridgeCompression parses a numbered argument and tests every
// bracket-bridge premise for removability.
func CheckBridgeCompression(output string) ([]BridgeCompressionEntry, Result) {
	argument, parseResult := ParseNumberedArgument(output, numberedArgumentOptions{})
	if !parseResult.Passed {
		return nil, parseResult
	}
	entries, err := NumberedArgumentBridgeCompression(argument)
	if err != nil {
		return nil, Result{Passed: false, Message: err.Error()}
	}
	return entries, Result{Passed: true, Message: "bridge compression computed"}
}

// NumberedArgumentBridgeCompression tests each bracket-bridge premise of a
// parsed numbered argument for removability. Bridges are detected on the
// unexpanded argument; gating runs against the expanded statements, so a
// conclusion citing several bridges is tested one removal at a time.
func NumberedArgumentBridgeCompression(arg NumberedArgument) ([]BridgeCompressionEntry, error) {
	bridgeSteps := make(map[int]bool)
	for _, step := range arg.Steps {
		if step.Kind == numberedArgumentStepConclusion {
			continue
		}
		if inlineStepBridgeExpansionPattern.MatchString(step.Statement) {
			bridgeSteps[step.Number] = true
		}
	}

	expanded, _, err := ExpandNumberedArgumentBridges(arg)
	if err != nil {
		return nil, err
	}
	stepsByNumber := make(map[int]NumberedArgumentStep, len(expanded.Steps))
	for _, step := range expanded.Steps {
		stepsByNumber[step.Number] = step
	}

	entries := make([]BridgeCompressionEntry, 0)
	for _, conclusion := range expanded.Steps {
		if conclusion.Kind != numberedArgumentStepConclusion {
			continue
		}
		for _, citation := range conclusion.Citations {
			if !bridgeSteps[citation] {
				continue
			}
			trimmed := conclusion
			trimmed.Citations = make([]int, 0, len(conclusion.Citations)-1)
			for _, kept := range conclusion.Citations {
				if kept != citation {
					trimmed.Citations = append(trimmed.Citations, kept)
				}
			}
			entry := BridgeCompressionEntry{BridgeStep: citation, Conclusion: conclusion.Number}
			if len(trimmed.Citations) > 0 {
				report := gateConclusionStep(trimmed, stepsByNumber)
				entry.Removable = report.Passed
				entry.Form = report.Form
				if !report.Passed {
					entry.Form = report.NearestForm
					entry.Mismatch = report.Mismatch
				}
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}
