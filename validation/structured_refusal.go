package validation

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type RefusalSpan struct {
	Step  int    `json:"step"`
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}
type StepRefusal struct {
	Span       RefusalSpan `json:"span"`
	Rule       string      `json:"rule"`
	Hint       string      `json:"hint"`
	RuleSource string      `json:"rule_source"`
}

func wholeRefusalSpan(step NumberedArgumentStep) RefusalSpan {
	text := strings.ReplaceAll(step.Statement, string(nameLetterJoiner), "")
	return RefusalSpan{Step: step.Number, Text: text, End: utf8.RuneCountInString(text)}
}

// Select the smallest literal excerpt the existing diagnostic actually names.
// Rune slicing preserves offsets even when Unicode case folding changes bytes.
func refusalExcerpt(step NumberedArgumentStep, excerpt string) (RefusalSpan, bool) {
	whole := wholeRefusalSpan(step)
	rs, needle := []rune(whole.Text), []rune(excerpt)
	if len(needle) == 0 {
		return whole, false
	}
	for i := 0; i+len(needle) <= len(rs); i++ {
		if strings.EqualFold(string(rs[i:i+len(needle)]), excerpt) {
			return RefusalSpan{Step: step.Number, Text: string(rs[i : i+len(needle)]), Start: i, End: i + len(needle)}, true
		}
	}
	return whole, false
}

func refusalFor(failed EntailmentGateStepReport, steps map[int]NumberedArgumentStep) *StepRefusal {
	step, found := steps[failed.Step]
	if !found {
		// No location in the old diagnostic: deterministically show the first statement.
		for id, s := range steps {
			if step.Number == 0 || id < step.Number {
				step = s
			}
		}
	}
	r := &StepRefusal{Span: wholeRefusalSpan(step), Rule: failed.NearestForm, Hint: failed.Mismatch, RuleSource: "nearest_form"}
	var note *StrictNote
	if len(failed.Notes) > 0 {
		note = &failed.Notes[0]
		r.RuleSource = "note"
	} else if failed.fired != nil {
		// Prefer the main diagnostic's rule; otherwise use catalog order.
		for i := range strictNotes {
			if !failed.fired.ids[strictNotes[i].ID] {
				continue
			}
			if note == nil {
				note = &strictNotes[i]
			}
			if strings.HasPrefix(failed.Mismatch, strictNotes[i].ID+":") {
				note = &strictNotes[i]
				break
			}
		}
		if note != nil {
			r.RuleSource = "fired"
		}
	}
	if note != nil {
		r.Rule = note.ID
		r.Hint = strings.Join(note.Fix, " ")
		if note.ID == "scope.mixed_andor" && strings.HasPrefix(failed.Mismatch, "mixed and/or inside an explicit frame; ") {
			// New frame refusals carry a repair using the actual operands.
			// Keep the catalog hint unchanged for all pre-existing refusals.
			_, r.Hint, _ = strings.Cut(failed.Mismatch, "; ")
		}
	}
	candidates := []NumberedArgumentStep{step}
	for _, id := range failed.Citations {
		candidates = append(candidates, steps[id])
	}
	for _, span := range failed.Ranges {
		candidates = append(candidates, steps[span[0]], steps[span[1]])
	}
	if r.Rule == "term.pronoun_cross_statement" || r.Rule == "term.possessive_cross_statement" {
		for _, candidate := range candidates {
			text := wholeRefusalSpan(candidate).Text
			for _, loc := range crossStatementWord.FindAllStringIndex(text, -1) {
				w := text[loc[0]:loc[1]]
				selected := crossStatementPronoun(w)
				if r.Rule == "term.possessive_cross_statement" {
					selected = crossStatementPossessive(w)
				}
				if !selected || strings.EqualFold(w, "it") && crossStatementScaffold(text[loc[0]:]) {
					continue
				}
				r.Span = RefusalSpan{Step: candidate.Number, Text: w, Start: utf8.RuneCountInString(text[:loc[0]]), End: utf8.RuneCountInString(text[:loc[1]])}
				return r
			}
		}
	}
	best := -1
	// The mismatch uses Go-quoted excerpts; do not invent a diagnostic location.
	for i := 0; i < len(failed.Mismatch); i++ {
		if failed.Mismatch[i] != '"' {
			continue
		}
		for j := i + 1; j < len(failed.Mismatch); j++ {
			if failed.Mismatch[j] == '\\' {
				j++
				continue
			}
			if failed.Mismatch[j] != '"' {
				continue
			}
			excerpt, err := strconv.Unquote(failed.Mismatch[i : j+1])
			if err == nil {
				for _, candidate := range candidates {
					if span, ok := refusalExcerpt(candidate, excerpt); ok && (best < 0 || span.End-span.Start < best) {
						r.Span = span
						best = span.End - span.Start
					}
				}
			}
			i = j
			break
		}
	}
	return r
}

// Output enrichment happens after verdicts, messages and strict notes are final.
func attachRefusals(argument NumberedArgument, report *EntailmentGateReport) {
	defer func() {
		for i := range report.Steps {
			report.Steps[i].fired = nil
		}
	}()
	if report.Passed {
		return
	}
	steps := make(map[int]NumberedArgumentStep, len(argument.Steps))
	for _, s := range argument.Steps {
		steps[s.Number] = s
	}
	for i := range report.Steps {
		failed := &report.Steps[i]
		if failed.Passed {
			continue
		}
		failed.Paraphrase, failed.ParaphraseForm = "", ""
		failed.Refusal = refusalFor(*failed, steps)
	}
}
