package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// FormalizedReconstructionOutput records one-based source sentence references.
type FormalizedReconstructionOutput struct {
	Argument string               `json:"argument"`
	Lines    []ReconstructionLine `json:"lines"`
}

type ReconstructionLine struct {
	Line      int   `json:"line"`
	Sentences []int `json:"sentences"`
}

type formalizedReconstructionOutputValidator struct{ options formalizedRepairOutputOptions }

func (v formalizedReconstructionOutputValidator) Validate(ctx context.Context, output string) Result {
	// Promptless validation checks structure; source bounds are checked by the
	// prompt-aware runner and again by assembly against its source envelope.
	return validateReconstruction(ctx, output, nil, v.options.EntailmentGate)
}

func ValidateFormalizedReconstructionOutputWithPrompt(ctx context.Context, output, prompt string, options ...json.RawMessage) (result Result) {
	defer func() { result = captureRepairFailure(result, output) }()
	if len(options) > 1 {
		return Result{Message: "formalized_reconstruction_output accepts at most one options value"}
	}
	var parsed formalizedRepairOutputOptions
	if len(options) == 1 {
		var err error
		parsed, err = parseFormalizedRepairOutputOptions(options[0])
		if err != nil {
			return Result{Message: strings.ReplaceAll(err.Error(), BuiltinFormalizedRepairOutput, BuiltinFormalizedReconstructionOutput)}
		}
	}
	source, ok := formalizedRepairTaggedValue(prompt, "source_material")
	if !ok {
		return Result{Message: "rendered prompt must contain <source_material>…</source_material> for sentence references"}
	}
	n := len(SegmentSourceSentences(source))
	return validateReconstruction(ctx, output, &n, parsed.EntailmentGate)
}

// ValidateFormalizedReconstructionSource checks a normalized workflow artifact
// directly against its source, without reparsing prompt tags.
func ValidateFormalizedReconstructionSource(ctx context.Context, output, source string) Result {
	n := len(SegmentSourceSentences(source))
	return validateReconstruction(ctx, output, &n, false)
}

func validateReconstruction(ctx context.Context, output string, sentenceCount *int, gate bool) (result Result) {
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
	if message := repairAssociatedConditional(reconstruction.Argument, mappings); message != "" {
		return Result{Message: message}
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
