package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// FormalizedReconstructionOutput records one-based source sentence references.
type FormalizedReconstructionOutput struct {
	Argument string               `json:"argument"`
	Lines    []ReconstructionLine `json:"lines"`
}

type ReconstructionLine struct {
	Line      int    `json:"line"`
	Sentences []int  `json:"sentences"`
	Assumed   bool   `json:"assumed,omitempty"`
	Hedge     string `json:"hedge,omitempty"`
}

type formalizedReconstructionOutputOptions struct {
	EntailmentGate      bool `json:"entailment_gate"`
	SalvageFinalAttempt bool `json:"salvage_final_attempt"`
}

func parseFormalizedReconstructionOutputOptions(raw json.RawMessage) (formalizedReconstructionOutputOptions, error) {
	var options formalizedReconstructionOutputOptions
	if len(raw) == 0 {
		return options, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&options); err != nil {
		return options, fmt.Errorf("invalid formalized_reconstruction_output options: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return options, fmt.Errorf("invalid formalized_reconstruction_output options: %w", err)
	}
	return options, nil
}

type formalizedReconstructionOutputValidator struct {
	options formalizedReconstructionOutputOptions
}

func (v formalizedReconstructionOutputValidator) Validate(ctx context.Context, output string) Result {
	// Promptless validation checks structure; source bounds and hedge matches are
	// checked by the prompt-aware runner and assembly against the source envelope.
	return validateReconstruction(ctx, output, nil, v.options.EntailmentGate)
}

func ValidateFormalizedReconstructionOutputWithPrompt(ctx context.Context, output, prompt string, options ...json.RawMessage) Result {
	return ValidateFormalizedReconstructionOutputForAttempt(ctx, output, prompt, 0, 0, options...)
}

// ValidateFormalizedReconstructionOutputForAttempt validates the raw writer reply.
// Attempt numbers are one-based; absent attempt information never enables salvage.
func ValidateFormalizedReconstructionOutputForAttempt(ctx context.Context, output, prompt string, attemptNumber, maxAttempts int, options ...json.RawMessage) (result Result) {
	defer func() { result = captureRepairFailure(result, output) }()
	if len(options) > 1 {
		return Result{Message: "formalized_reconstruction_output accepts at most one options value"}
	}
	var parsed formalizedReconstructionOutputOptions
	if len(options) == 1 {
		var err error
		parsed, err = parseFormalizedReconstructionOutputOptions(options[0])
		if err != nil {
			return Result{Message: err.Error()}
		}
	}
	source, ok := formalizedRepairTaggedValue(prompt, "source_material")
	if !ok {
		return Result{Message: "rendered prompt must contain <source_material>…</source_material> for sentence references"}
	}
	result = validateReconstruction(ctx, output, &source, parsed.EntailmentGate)
	if result.Passed || !parsed.SalvageFinalAttempt || maxAttempts <= 0 || attemptNumber < maxAttempts || ctx.Err() != nil {
		return result
	}
	// Salvage independently checks all writer structure and provenance constraints,
	// skipping only associated-conditional padding and the entailment gate.
	salvaged, err := SalvageReconstruction(output, source)
	if err != nil {
		return result
	}
	normalized, _ := json.Marshal(salvaged.Reconstruction)
	checked := validateReconstructionOutput(ctx, string(normalized), &source, parsed.EntailmentGate, true, false)
	if !checked.Passed {
		return result
	}
	checked.Details["salvaged"] = map[string]any{"assumed_lines": salvaged.AssumedLines, "failing_steps": salvaged.FailingSteps}
	return checked
}

// ValidateFormalizedReconstructionSource checks a writer reply directly against
// its source, without reparsing prompt tags. Writer-set assumed remains forbidden.
func ValidateFormalizedReconstructionSource(ctx context.Context, output, source string) Result {
	return validateReconstruction(ctx, output, &source, false)
}

func validateReconstruction(ctx context.Context, output string, source *string, gate bool) Result {
	return validateReconstructionChecks(ctx, output, source, gate, true)
}

// validateReconstructionChecks shares all writer structure checks with salvage.
func validateReconstructionChecks(ctx context.Context, output string, source *string, gate, padding bool) Result {
	return validateReconstructionOutput(ctx, output, source, gate, padding, true)
}

// ValidateNormalizedReconstructionOutput accepts tool-owned assumed markers in
// downstream workflow artifacts. Never use this entry point for writer replies.
func ValidateNormalizedReconstructionOutput(ctx context.Context, output string, options ...json.RawMessage) Result {
	if len(options) > 1 {
		return Result{Message: "formalized_reconstruction_output accepts at most one options value"}
	}
	var parsed formalizedReconstructionOutputOptions
	if len(options) == 1 {
		var err error
		parsed, err = parseFormalizedReconstructionOutputOptions(options[0])
		if err != nil {
			return Result{Message: err.Error()}
		}
	}
	return validateReconstructionOutput(ctx, output, nil, parsed.EntailmentGate, true, false)
}

// ValidateNormalizedReconstructionSource also checks source sentence bounds and hedge matches.
func ValidateNormalizedReconstructionSource(ctx context.Context, output, source string) Result {
	return validateReconstructionOutput(ctx, output, &source, false, true, false)
}

func validateReconstructionOutput(ctx context.Context, output string, source *string, gate, padding, writer bool) (result Result) {
	defer func() { result = captureRepairFailure(result, output) }()
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Message: "validation timed out"}
	default:
	}
	var sentenceCount *int
	var sourceSentences []string
	if source != nil {
		runes := []rune(*source)
		for _, sentence := range SegmentSourceSentences(*source) {
			sourceSentences = append(sourceSentences, string(runes[sentence.Start:sentence.End]))
		}
		n := len(sourceSentences)
		sentenceCount = &n
	}
	extracted, normalization, err := extractRepairJSONObject(output)
	if err != nil {
		return Result{Message: "return one reconstruction JSON object with argument and lines: " + err.Error()}
	}
	var reconstruction FormalizedReconstructionOutput
	decoder := json.NewDecoder(bytes.NewBufferString(extracted))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reconstruction); err != nil {
		return Result{Message: "return argument and lines with integer line and sentence numbers: " + err.Error()}
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Result{Message: "return only one reconstruction JSON object"}
	}
	// assumed is reserved for deterministic recovery, even when explicitly false.
	var fields struct {
		Lines []map[string]json.RawMessage `json:"lines"`
	}
	_ = json.Unmarshal([]byte(extracted), &fields)
	for i, line := range fields.Lines {
		for field, raw := range line {
			if strings.EqualFold(field, "hedge") {
				var hedge string
				if string(raw) == "null" || json.Unmarshal(raw, &hedge) != nil || strings.TrimSpace(hedge) == "" {
					return Result{Message: fmt.Sprintf("line %d hedge must be a non-empty string; omit hedge if there is no author hedge", reconstruction.Lines[i].Line)}
				}
			}
			if writer && strings.EqualFold(field, "assumed") {
				return Result{Message: "writer reconstruction must not set assumed"}
			}
		}
	}
	syntax := numberedArgumentSyntaxValidator{}.Validate(ctx, reconstruction.Argument)
	if !syntax.Passed {
		return Result{Message: "reconstruction argument failed numbered_argument_syntax: " + syntax.Message, Details: map[string]any{"argument_validation": syntax.Details}}
	}
	argument, _ := ParseNumberedArgumentAllowingNoConclusions(reconstruction.Argument)
	final := argument.Steps[len(argument.Steps)-1]
	if final.Kind != "conclusion" {
		return Result{Message: fmt.Sprintf("line %d must be the overall conclusion; add a conclusion marker and citations", final.Number)}
	}
	seen := map[int]bool{}
	mappings := []FormalizationLineMapping{}
	grounded := false
	for _, line := range reconstruction.Lines {
		if line.Line < 1 || line.Line > len(argument.Steps) {
			return Result{Message: fmt.Sprintf("line %d is outside the argument; use an existing line number in 1..%d", line.Line, len(argument.Steps))}
		}
		if seen[line.Line] {
			return Result{Message: fmt.Sprintf("line %d is repeated; keep exactly one lines entry for it", line.Line)}
		}
		seen[line.Line] = true
		if line.Sentences == nil {
			return Result{Message: fmt.Sprintf("line %d requires a sentences array; use [] for an unsourced line", line.Line)}
		}
		sentenceSeen := map[int]bool{}
		for _, n := range line.Sentences {
			if n < 1 || (sentenceCount != nil && n > *sentenceCount) {
				bounds := "positive one-based source sentence numbers"
				if sentenceCount != nil {
					bounds = fmt.Sprintf("source sentence numbers in 1..%d", *sentenceCount)
				}
				return Result{Message: fmt.Sprintf("line %d references sentence %d; use %s", line.Line, n, bounds)}
			}
			if sentenceSeen[n] {
				return Result{Message: fmt.Sprintf("line %d repeats sentence %d; remove the duplicate", line.Line, n)}
			}
			sentenceSeen[n] = true
		}
		if line.Hedge != "" {
			if line.Assumed {
				return Result{Message: fmt.Sprintf("line %d is assumed; remove its hedge", line.Line)}
			}
			if utf8.RuneCountInString(line.Hedge) > 80 {
				return Result{Message: fmt.Sprintf("line %d hedge must be at most 80 characters; quote only the author's hedge words", line.Line)}
			}
			if len(line.Sentences) == 0 {
				return Result{Message: fmt.Sprintf("line %d has a hedge; cite at least one source sentence containing it", line.Line)}
			}
			if source != nil {
				found := false
				labels := []string{}
				for _, n := range line.Sentences {
					labels = append(labels, fmt.Sprintf("S%d", n))
					if strings.Contains(normalizeAuthorHedge(sourceSentences[n-1]), normalizeAuthorHedge(line.Hedge)) {
						found = true
					}
				}
				if !found {
					return Result{Message: fmt.Sprintf("line %d hedge %q does not appear in its cited sentences %s; quote hedge words from a cited sentence or correct the citations", line.Line, line.Hedge, strings.Join(labels, ","))}
				}
			}
		}
		if line.Assumed {
			if len(line.Sentences) > 0 || argument.Steps[line.Line-1].Kind != "premise" {
				return Result{Message: fmt.Sprintf("assumed line %d must be an unsourced premise", line.Line)}
			}
			// Tool-owned bridges are explicitly disclosed, not writer padding.
			mappings = append(mappings, FormalizationLineMapping{RepairedLine: line.Line, FromAsFoundLines: []int{line.Line}})
		}
		if len(line.Sentences) > 0 {
			// Use the legacy detector's mapped-line interface without changing its
			// interpretation of stored repair mappings.
			mappings = append(mappings, FormalizationLineMapping{RepairedLine: line.Line, FromAsFoundLines: []int{line.Line}})
			if argument.Steps[line.Line-1].Kind == "premise" {
				grounded = true
			}
		}
		if line.Line == final.Number && len(line.Sentences) == 0 {
			return Result{Message: fmt.Sprintf("line %d is the final conclusion; cite at least one source sentence containing the author's conclusion", line.Line)}
		}
	}
	for _, step := range argument.Steps {
		if !seen[step.Number] {
			return Result{Message: fmt.Sprintf("line %d is missing from lines; add exactly one entry with its sentences array", step.Number)}
		}
	}
	if !grounded {
		return Result{Message: "every premise line is supplied; add source sentence references to at least one premise line"}
	}
	if padding {
		if message := repairAssociatedConditional(reconstruction.Argument, mappings); message != "" {
			return Result{Message: message}
		}
	}
	sort.Slice(reconstruction.Lines, func(i, j int) bool { return reconstruction.Lines[i].Line < reconstruction.Lines[j].Line })
	normalized, _ := json.Marshal(reconstruction)
	result = Result{Passed: true, Message: "output matches formalized_reconstruction_output", ParsedOutput: normalized, Details: map[string]any{"step_count": len(argument.Steps)}}
	if normalization != "" {
		result.Details["normalizations"] = []string{normalization}
	}
	if gate {
		result = applyArgumentEntailmentGate(result, reconstruction.Argument, mappings)
	}
	return result
}

// Match only typography, whitespace and case differences; keep the quoted words.
func normalizeAuthorHedge(text string) string {
	text = strings.NewReplacer("‘", "'", "’", "'", "“", "\"", "”", "\"").Replace(text)
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}
