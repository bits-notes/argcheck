package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type formalizedRepairOutput struct {
	Argument       *string                         `json:"argument"`
	CoverageAnchor *string                         `json:"coverage_anchor"`
	LineMapping    *[]formalizedRepairLineMapping  `json:"line_mapping"`
	Dropped        *[]formalizedRepairDroppedEntry `json:"dropped"`
}

type formalizedRepairLineMapping = FormalizationLineMapping

type formalizedRepairDroppedEntry struct {
	AsFoundLine int    `json:"as_found_line"`
	Kind        string `json:"kind"`
	Reason      string `json:"reason"`
}

type formalizedRepairOutputValidator struct{}

func (formalizedRepairOutputValidator) Validate(ctx context.Context, output string) (result Result) {
	defer func() { result = captureRepairFailure(result, output) }()
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Passed: false, Message: "validation timed out"}
	default:
	}

	extracted, normalization, err := extractRepairJSONObject(output)
	if err != nil {
		return Result{Message: "output must be one formalized repair JSON object: " + err.Error()}
	}
	decoder := json.NewDecoder(bytes.NewBufferString(extracted))
	decoder.DisallowUnknownFields()
	var repair formalizedRepairOutput
	if err := decoder.Decode(&repair); err != nil {
		return Result{Passed: false, Message: "output must be one formalized repair JSON object: " + err.Error()}
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Result{Passed: false, Message: "output must contain only one formalized repair JSON object"}
	}
	if repair.Argument == nil {
		return Result{Passed: false, Message: "formalized repair output requires argument"}
	}
	if repair.CoverageAnchor == nil {
		return Result{Passed: false, Message: "formalized repair output requires coverage_anchor"}
	}
	if *repair.CoverageAnchor != FormalizationCoverageAnchorSource && *repair.CoverageAnchor != FormalizationCoverageAnchorAsFound {
		return Result{Passed: false, Message: "formalized repair coverage_anchor must be source or as_found"}
	}
	if repair.LineMapping == nil {
		return Result{Passed: false, Message: "formalized repair output requires line_mapping"}
	}
	if repair.Dropped == nil {
		return Result{Passed: false, Message: "formalized repair output requires dropped"}
	}

	argumentResult := numberedArgumentSyntaxValidator{}.Validate(ctx, *repair.Argument)
	if !argumentResult.Passed {
		return Result{
			Passed:  false,
			Message: "formalized repair argument failed numbered_argument_syntax: " + argumentResult.Message,
			Details: map[string]any{"argument_validation": argumentResult.Details},
		}
	}
	if len(*repair.LineMapping) == 0 {
		return Result{Passed: false, Message: "formalized repair line_mapping must contain at least one entry"}
	}
	seenRepaired := make(map[int]bool, len(*repair.LineMapping))
	for index, mapping := range *repair.LineMapping {
		if mapping.RepairedLine < 1 {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].repaired_line must be a positive integer", index)}
		}
		if mapping.RepairedLine > argumentResult.Details["step_count"].(int) {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].repaired_line references line %d outside the argument", index, mapping.RepairedLine)}
		}
		if seenRepaired[mapping.RepairedLine] {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping has duplicate repaired_line %d", mapping.RepairedLine)}
		}
		seenRepaired[mapping.RepairedLine] = true
		if len(mapping.FromAsFoundLines) == 0 && mapping.FromSourceQuote == nil {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].from_as_found_lines must contain at least one line", index)}
		}
		if mapping.FromSourceQuote != nil {
			parsed, _ := ParseNumberedArgumentAllowingNoConclusions(*repair.Argument)
			// Without prompt context, still check length and lexical alignment. Source
			// membership is checked by the prompt-aware validator and by assembly.
			if _, _, err := ValidateSourceQuote(*mapping.FromSourceQuote, parsed.Steps[mapping.RepairedLine-1].Statement, *mapping.FromSourceQuote); err != nil {
				return Result{Passed: false, Message: fmt.Sprintf("repaired line %d: %v", mapping.RepairedLine, err)}
			}
		}
		seenSources := make(map[int]bool, len(mapping.FromAsFoundLines))
		for _, line := range mapping.FromAsFoundLines {
			if line < 1 {
				return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].from_as_found_lines must contain only positive integers", index)}
			}
			if seenSources[line] {
				return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].from_as_found_lines repeats line %d", index, line)}
			}
			seenSources[line] = true
		}
	}

	seenDropped := make(map[int]bool, len(*repair.Dropped))
	for index, dropped := range *repair.Dropped {
		if dropped.AsFoundLine < 1 {
			return Result{Passed: false, Message: fmt.Sprintf("dropped[%d].as_found_line must be a positive integer", index)}
		}
		if seenDropped[dropped.AsFoundLine] {
			return Result{Passed: false, Message: fmt.Sprintf("dropped has duplicate as_found_line %d", dropped.AsFoundLine)}
		}
		seenDropped[dropped.AsFoundLine] = true
		switch dropped.Kind {
		case DroppedContentKindPremise, DroppedContentKindConclusion, DroppedContentKindCitation:
		default:
			return Result{Passed: false, Message: fmt.Sprintf("dropped[%d].kind must be premise, conclusion, or citation", index)}
		}
		if strings.TrimSpace(dropped.Reason) == "" {
			return Result{Passed: false, Message: fmt.Sprintf("dropped[%d].reason must be non-empty", index)}
		}
	}

	details := map[string]any{"step_count": argumentResult.Details["step_count"], "mapping_count": len(*repair.LineMapping), "dropped_count": len(*repair.Dropped)}
	if normalization != "" {
		details["normalizations"] = []string{normalization}
	}
	return Result{
		Passed:       true,
		Message:      "output matches formalized_repair_output",
		Details:      details,
		ParsedOutput: json.RawMessage(extracted),
	}
}

// ValidateFormalizedRepairOutputWithPrompt runs the same built-in validator as
// the atomic output contract, then enforces coverage using the source and
// as-found arguments embedded in the repair node's rendered prompt. This is
// called by the LLM runner, so failures enter the ordinary paid retry path.
// Re-validation without the original prompt remains structurally deterministic;
// the assembly transform performs the defensive dependency-backed check.
func ValidateFormalizedRepairOutputWithPrompt(ctx context.Context, output, prompt string) (result Result) {
	defer func() { result = captureRepairFailure(result, output) }()
	result = formalizedRepairOutputValidator{}.Validate(ctx, output)
	if !result.Passed {
		return result
	}
	source, sourceOK := formalizedRepairTaggedValue(prompt, "source_material")
	asFound, asFoundOK := formalizedRepairTaggedValue(prompt, "as_found_argument")

	var repair formalizedRepairOutput
	if err := json.Unmarshal(result.ParsedOutput, &repair); err != nil {
		return Result{Passed: false, Message: "decode formalized repair coverage context: " + err.Error()}
	}
	if err := ValidateRestoredLineMappings(source, *repair.Argument, *repair.LineMapping); err != nil {
		return Result{Passed: false, Message: err.Error()}
	}
	if !sourceOK || !asFoundOK {
		return result
	}
	expectedAnchor := FormalizationCoverageAnchorAsFound
	anchorArgument := asFound
	if _, parsed := ParseNumberedArgumentAllowingNoConclusions(source); parsed.Passed {
		expectedAnchor = FormalizationCoverageAnchorSource
		anchorArgument = source
	}
	if repair.CoverageAnchor == nil || *repair.CoverageAnchor != expectedAnchor {
		normalizations, _ := result.Details["normalizations"].([]string)
		result.Details["normalizations"] = append(normalizations, "coverage_anchor:"+stringValue(repair.CoverageAnchor)+"->"+expectedAnchor)
		repair.CoverageAnchor = &expectedAnchor
		result.ParsedOutput, _ = json.Marshal(repair)
	}

	mapping := make([]FormalizationLineMapping, len(*repair.LineMapping))
	mappedRepaired := make(map[int]bool, len(mapping))
	for index, entry := range *repair.LineMapping {
		mapping[index] = entry
		mappedRepaired[entry.RepairedLine] = true
	}
	dropped := make([]FormalizationDroppedContent, len(*repair.Dropped))
	for index, entry := range *repair.Dropped {
		dropped[index] = FormalizationDroppedContent{AsFoundLine: entry.AsFoundLine, Kind: entry.Kind, Reason: entry.Reason}
	}

	repairedArgument, parsed := ParseNumberedArgumentAllowingNoConclusions(*repair.Argument)
	if !parsed.Passed {
		return Result{Passed: false, Message: "parse formalized repair argument for coverage: " + parsed.Message}
	}
	suppliedLines := make([]int, 0)
	for _, step := range repairedArgument.Steps {
		if step.Kind != DroppedContentKindPremise || mappedRepaired[step.Number] {
			continue
		}
		if formalizedRepairStatementPresent(step.Statement, source) || formalizedRepairStatementPresent(step.Statement, asFound) {
			return Result{Passed: false, Message: fmt.Sprintf("repaired premise line %d is present in the source/as-found input and must be mapped", step.Number)}
		}
		suppliedLines = append(suppliedLines, step.Number)
	}
	if _, _, err := ValidateFormalizationCoverageAnchored(expectedAnchor, anchorArgument, *repair.Argument, mapping, dropped, suppliedLines, source); err != nil {
		return Result{Passed: false, Message: "formalized repair coverage validation failed: " + err.Error()}
	}
	result.Details["coverage_anchor"] = expectedAnchor
	result.Details["coverage_passed"] = true
	return result
}

func formalizedRepairTaggedValue(prompt, tag string) (string, bool) {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(prompt, open)
	if start < 0 {
		return "", false
	}
	start += len(open)
	end := strings.Index(prompt[start:], close)
	if end < 0 {
		return "", false
	}
	return strings.TrimSpace(prompt[start : start+end]), true
}

func formalizedRepairStatementPresent(statement, text string) bool {
	normalizedStatement := normalizeFormalizedRepairText(statement)
	normalizedText := normalizeFormalizedRepairText(text)
	return normalizedStatement != "" && strings.Contains(normalizedText, normalizedStatement)
}

func normalizeFormalizedRepairText(value string) string {
	var normalized strings.Builder
	space := false
	for _, char := range strings.ToLower(value) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			if space && normalized.Len() > 0 {
				normalized.WriteByte(' ')
			}
			space = false
			normalized.WriteRune(char)
		} else {
			space = true
		}
	}
	return normalized.String()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("extra JSON value")
		}
		return err
	}
	return nil
}

// Capture only model output, never prompts, credentials, or request metadata.
func captureRepairFailure(result Result, output string) Result {
	if result.Passed {
		return result
	}
	if result.Details == nil {
		result.Details = map[string]any{}
	}
	for key, value := range RawReplyDetails(output) {
		result.Details[key] = value
	}
	return result
}

// RawReplyDetails bounds diagnostics by Unicode characters, retaining only the
// supplied model reply. Callers must not pass prompts or request metadata.
func RawReplyDetails(output string) map[string]any {
	chars := utf8.RuneCountInString(output)
	excerpt := output
	count := 0
	for i := range output {
		if count == 4000 {
			excerpt = output[:i]
			break
		}
		count++
	}
	return map[string]any{"raw_reply_excerpt": excerpt, "raw_reply_chars": chars}
}

// extractRepairJSONObject counts balanced objects across the entire reply, even
// outside a fence. Braces and escaped quotes inside JSON strings are ignored.
func extractRepairJSONObject(output string) (string, string, error) {
	trimmed := strings.TrimSpace(output)
	if json.Valid([]byte(trimmed)) {
		if strings.HasPrefix(trimmed, "{") {
			return trimmed, "", nil
		}
		return "", "", fmt.Errorf("expected a JSON object")
	}
	start, depth, count := -1, 0, 0
	quoted, escaped := false, false
	object := ""
	for i, c := range output {
		if depth == 0 {
			if c == '{' {
				start, depth = i, 1
			}
			continue
		}
		if quoted {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				quoted = false
			}
			continue
		}
		switch c {
		case '"':
			quoted = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				count++
				object = output[start : i+1]
			}
		}
	}
	if count != 1 || depth != 0 {
		return "", "", fmt.Errorf("expected exactly one balanced JSON object")
	}
	normalization := "extracted_embedded_json"
	if strings.Contains(output, "```") {
		parts := strings.Split(output, "```")
		if len(parts) != 3 {
			return "", "", fmt.Errorf("expected a single JSON fence")
		}
		block := strings.TrimSpace(parts[1])
		if strings.HasPrefix(block, "json\n") {
			block = strings.TrimSpace(strings.TrimPrefix(block, "json\n"))
		}
		if strings.HasPrefix(block, "json\r\n") {
			block = strings.TrimSpace(strings.TrimPrefix(block, "json\r\n"))
		}
		if block != object {
			return "", "", fmt.Errorf("fence must contain only the JSON object")
		}
		normalization = "extracted_fenced_json"
	}
	return object, normalization, nil
}
