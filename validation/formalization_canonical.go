package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	BuiltinFormalizationCanonicalizeOutput    = "formalization_canonicalize_output"
	BuiltinFormalizationLineEquivalenceOutput = "formalization_line_equivalence_output"
)

type CanonicalLine struct {
	Line      int    `json:"line"`
	Canonical string `json:"canonical"`
}
type CanonicalizeOutput struct {
	Lines []CanonicalLine `json:"lines"`
}
type LineEquivalence struct {
	Line   int    `json:"line"`
	Same   *bool  `json:"same"`
	Reason string `json:"reason"`
}
type LineEquivalenceOutput struct {
	Pairs []LineEquivalence `json:"pairs"`
}
type CanonicalChange struct {
	Line      int    `json:"line"`
	Original  string `json:"original"`
	Canonical string `json:"canonical"`
	Accepted  bool   `json:"accepted"`
	Why       string `json:"why"`
}
type CanonicalApplication struct {
	Argument string            `json:"argument"`
	Original string            `json:"original"`
	Changes  []CanonicalChange `json:"changes"`
}

type canonicalOutputValidator struct{ name string }

func (v canonicalOutputValidator) Validate(ctx context.Context, output string) Result {
	return ValidateCanonicalOutputWithPrompt(ctx, v.name, output, "")
}

// As with the repair contract, bare validation checks structure; the queued
// runner supplies the rendered prompt to enforce the one-to-one contract.
func ValidateCanonicalOutputWithPrompt(ctx context.Context, name, output, prompt string) Result {
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Message: "validation timed out"}
	default:
	}
	raw, normalization, err := extractRepairJSONObject(output)
	if err != nil {
		return Result{Message: err.Error()}
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	fail := func(err error) Result { return Result{Message: err.Error()} }
	switch name {
	case BuiltinFormalizationCanonicalizeOutput:
		var value CanonicalizeOutput
		if err := decoder.Decode(&value); err != nil {
			return fail(err)
		}
		if len(value.Lines) == 0 {
			return Result{Message: "lines must be a non-empty array"}
		}
		for i, line := range value.Lines {
			if line.Line != i+1 {
				return Result{Message: "lines must enumerate every line in order starting at 1"}
			}
			if err := canonicalStatementShape(line.Canonical); err != nil {
				return fail(err)
			}
		}
		if prompt != "" {
			original, ok := formalizedRepairTaggedValue(prompt, "as_found_argument")
			if !ok {
				return Result{Message: "prompt requires <as_found_argument>"}
			}
			if err := ValidateCanonicalLines(original, value); err != nil {
				return fail(err)
			}
		}
	case BuiltinFormalizationLineEquivalenceOutput:
		var value LineEquivalenceOutput
		if err := decoder.Decode(&value); err != nil {
			return fail(err)
		}
		if len(value.Pairs) == 0 {
			return Result{Message: "pairs must be a non-empty array"}
		}
		last := 0
		for _, pair := range value.Pairs {
			if pair.Line <= last || pair.Same == nil || strings.TrimSpace(pair.Reason) == "" {
				return Result{Message: "pairs require ordered unique positive line numbers, boolean same, and non-empty reason"}
			}
			last = pair.Line
		}
		if prompt != "" {
			block, ok := formalizedRepairTaggedValue(prompt, "pairs")
			if !ok {
				return Result{Message: "prompt requires <pairs>"}
			}
			rows := strings.Split(block, "\n")
			if len(rows) != len(value.Pairs) {
				return Result{Message: "pair count differs from prompt"}
			}
			for i, row := range rows {
				fields := strings.Split(row, " | ")
				if len(fields) != 3 || strings.TrimSpace(fields[1]) == "" || strings.TrimSpace(fields[2]) == "" {
					return Result{Message: "pairs format must be N | ORIGINAL | CANONICAL"}
				}
				number, err := strconv.Atoi(strings.TrimSpace(fields[0]))
				if err != nil || number != value.Pairs[i].Line {
					return Result{Message: "pair line differs from prompt"}
				}
			}
		}
	default:
		return Result{Message: "unknown canonical output contract"}
	}
	if err := requireJSONEOF(decoder); err != nil {
		return fail(err)
	}
	details := map[string]any{}
	if normalization != "" {
		details["normalizations"] = []string{normalization}
	}
	return Result{Passed: true, Message: "output matches " + name, ParsedOutput: json.RawMessage(raw), Details: details}
}

var canonicalSentenceBreak = regexp.MustCompile(`[.!?]\s+\S`)
var canonicalTail = regexp.MustCompile(` \[[0-9, ]+\]$`)

func canonicalStatementShape(text string) error {
	if strings.TrimSpace(text) == "" || strings.ContainsAny(text, "\r\n") {
		return fmt.Errorf("canonical must be a single non-empty statement")
	}
	body := canonicalTail.ReplaceAllString(text, "")
	if canonicalSentenceBreak.MatchString(body) {
		return fmt.Errorf("canonical must be a single statement")
	}
	// Reuse syntax checks (without connectivity or gate requirements).
	if _, result := parseNumberedArgumentLine("1. "+text, 1); !result.Passed {
		return fmt.Errorf("canonical statement syntax: %s", result.Message)
	}
	return nil
}

// Canonical text includes its terminal period, Therefore marker, and citations,
// but excludes the numbered-line prefix. Keep citation spelling byte-identical.
func CanonicalOriginalLines(original string) ([]string, error) {
	argument, result := ParseNumberedArgumentAllowingNoConclusions(original)
	if !result.Passed {
		return nil, fmt.Errorf("as-found argument: %s", result.Message)
	}
	lines := make([]string, len(argument.Steps))
	for i, step := range argument.Steps {
		lines[i] = strings.TrimPrefix(step.Line, fmt.Sprintf("%d. ", step.Number))
	}
	return lines, nil
}
func ValidateCanonicalLines(original string, value CanonicalizeOutput) error {
	lines, err := CanonicalOriginalLines(original)
	if err != nil {
		return err
	}
	if len(lines) != len(value.Lines) {
		return fmt.Errorf("canonical line count differs from as-found argument")
	}
	for i, line := range value.Lines {
		if line.Line != i+1 {
			return fmt.Errorf("canonical lines must match as-found order")
		}
		if err := canonicalStatementShape(line.Canonical); err != nil {
			return err
		}
		if err := canonicalMarkers(lines[i], line.Canonical); err != nil {
			return fmt.Errorf("line %d: %w", line.Line, err)
		}
	}
	return nil
}
func canonicalMarkers(original, canonical string) error {
	if strings.HasPrefix(original, "Therefore,") != strings.HasPrefix(canonical, "Therefore,") || canonicalTail.FindString(original) != canonicalTail.FindString(canonical) {
		return fmt.Errorf("conclusion marker or citations changed")
	}
	if !strings.HasPrefix(original, "Therefore,") && strings.ContainsAny(canonical, "[]") {
		return fmt.Errorf("premises cannot contain citations")
	}
	return nil
}

var canonicalFormWords = map[string]bool{"if": true, "then": true, "either": true, "or": true, "not": true, "case": true, "it": true, "is": true, "the": true, "that": true, "every": true, "all": true, "no": true}
var canonicalNumber = regexp.MustCompile(`[+-]?[0-9]+(?:[.,][0-9]+)*`)

func canonicalPolarity(text string) int {
	text = sourceQuoteNegatedContraction.ReplaceAllString(strings.ToLower(text), " not")
	count := 0
	for _, word := range sourceQuoteWords.FindAllString(text, -1) {
		// "no one" and "it is not the case that" each contribute just
		// their single "no" or "not" token. Contractions are expanded above.
		switch word {
		case "not", "no", "never", "nothing", "none", "nobody", "nowhere", "neither", "nor", "cannot":
			count++
		}
	}
	return count
}
func CanonicalContentCheck(original, canonical string) error {
	if err := canonicalStatementShape(canonical); err != nil {
		return err
	}
	if err := canonicalMarkers(original, canonical); err != nil {
		return err
	}
	aText, bText := canonicalTail.ReplaceAllString(original, ""), canonicalTail.ReplaceAllString(canonical, "")
	if canonicalPolarity(aText) != canonicalPolarity(bText) {
		return fmt.Errorf("negation count changed")
	}
	if strings.Join(canonicalNumber.FindAllString(aText, -1), "|") != strings.Join(canonicalNumber.FindAllString(bText, -1), "|") {
		return fmt.Errorf("number tokens changed")
	}
	a, b := sourceContentWords(aText), sourceContentWords(bText)
	// Normalize the equivalent polarity spellings without weakening the shared
	// source-quote checker. Counts above retain repeated and double negation.
	for _, words := range []map[string]bool{a, b} {
		if words["no"] {
			delete(words, "no")
			words["not"] = true
		}
		// Polarity is checked separately above; these share the same content
		// in rewrites such as "heard nothing" -> "not ... heard anything".
		for _, word := range []string{"anything", "nothing"} {
			if words[word] {
				delete(words, word)
				words["something"] = true
			}
		}
	}
	shared := 0
	var introduced []string
	for word := range b {
		if !a[word] && !canonicalFormWords[word] {
			introduced = append(introduced, word)
		}
	}
	if len(introduced) > 0 {
		sort.Strings(introduced)
		return fmt.Errorf("new content words: %s", strings.Join(introduced, ", "))
	}
	for word := range a {
		if b[word] {
			shared++
		}
	}
	if 5*shared < 4*len(a) {
		return fmt.Errorf("less than 80%% of original content words retained")
	}
	return nil
}

// ApplyCanonicalization fails malformed contracts, but rejects unsafe proposals
// per line so an invalid candidate can never replace the author's words.
func ApplyCanonicalization(original string, proposals CanonicalizeOutput, judge LineEquivalenceOutput) (CanonicalApplication, error) {
	lines, err := CanonicalOriginalLines(original)
	if err != nil {
		return CanonicalApplication{}, err
	}
	result := CanonicalApplication{Original: original, Changes: make([]CanonicalChange, 0, len(lines))}
	if len(proposals.Lines) != len(lines) || len(judge.Pairs) != len(lines) {
		return result, fmt.Errorf("canonicalize and equivalence outputs must cover every as-found line")
	}
	rendered := make([]string, len(lines))
	for i, text := range lines {
		proposal, pair := proposals.Lines[i], judge.Pairs[i]
		if proposal.Line != i+1 || pair.Line != i+1 || pair.Same == nil || strings.TrimSpace(pair.Reason) == "" {
			return result, fmt.Errorf("invalid canonicalize/equivalence line %d", i+1)
		}
		change := CanonicalChange{Line: i + 1, Original: text, Canonical: proposal.Canonical}
		switch {
		case text == proposal.Canonical:
			change.Accepted = true
			change.Why = "unchanged"
		case !*pair.Same:
			change.Why = "equivalence judge rejected: " + pair.Reason
		default:
			if err := CanonicalContentCheck(text, proposal.Canonical); err != nil {
				change.Why = err.Error()
			} else {
				change.Accepted = true
				change.Why = "deterministic checks and equivalence judge passed"
			}
		}
		if change.Accepted {
			text = proposal.Canonical
		}
		rendered[i] = fmt.Sprintf("%d. %s", i+1, text)
		result.Changes = append(result.Changes, change)
	}
	result.Argument = strings.Join(rendered, "\n")
	return result, nil
}

// DecodeCanonicalOutputs shares tolerant JSON extraction with both contracts.
// Per-line text checks run during apply, where rejection retains the original.
func DecodeCanonicalOutputs(proposalText, judgeText string) (CanonicalizeOutput, LineEquivalenceOutput, error) {
	var proposals CanonicalizeOutput
	var judge LineEquivalenceOutput
	for _, item := range []struct {
		text   string
		target any
	}{{proposalText, &proposals}, {judgeText, &judge}} {
		raw, _, err := extractRepairJSONObject(item.text)
		if err != nil {
			return proposals, judge, err
		}
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(item.target); err != nil {
			return proposals, judge, err
		}
		if err := requireJSONEOF(decoder); err != nil {
			return proposals, judge, err
		}
	}
	return proposals, judge, nil
}

func ValidateCanonicalApplication(value CanonicalApplication, original string) error {
	if value.Original != original {
		return fmt.Errorf("canonical original differs from as-found argument")
	}
	lines, err := CanonicalOriginalLines(original)
	if err != nil {
		return err
	}
	if len(lines) != len(value.Changes) {
		return fmt.Errorf("canonical changes must cover every original line")
	}
	rendered := make([]string, len(lines))
	for i, change := range value.Changes {
		if change.Line != i+1 || change.Original != lines[i] {
			return fmt.Errorf("canonical change does not match original line %d", i+1)
		}
		text := lines[i]
		if change.Accepted {
			if text != change.Canonical {
				if err := CanonicalContentCheck(text, change.Canonical); err != nil {
					return err
				}
			}
			text = change.Canonical
		}
		rendered[i] = fmt.Sprintf("%d. %s", i+1, text)
	}
	if value.Argument != strings.Join(rendered, "\n") {
		return fmt.Errorf("canonical argument differs from accepted changes")
	}
	return nil
}
