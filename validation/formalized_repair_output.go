package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
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

type formalizedRepairOutputOptions struct {
	EntailmentGate bool `json:"entailment_gate"`
}

type formalizedRepairOutputValidator struct {
	options formalizedRepairOutputOptions
	source  string // Used only to enrich early lexical-failure diagnostics.
}

func parseFormalizedRepairOutputOptions(raw json.RawMessage) (formalizedRepairOutputOptions, error) {
	var options formalizedRepairOutputOptions
	if len(raw) == 0 {
		return options, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&options); err != nil {
		return options, fmt.Errorf("invalid formalized_repair_output options: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return options, fmt.Errorf("invalid formalized_repair_output options: %w", err)
	}
	return options, nil
}

// The repair contract has no verdict: assembly obtains it from repair-validator.
// Gate only opted-in IP versions; legacy versions can return uncertified repairs.
func applyRepairEntailmentGate(result Result, enabled bool) Result {
	if !result.Passed || !enabled {
		return result
	}
	var repair formalizedRepairOutput
	if err := json.Unmarshal(result.ParsedOutput, &repair); err != nil {
		return Result{Message: "decode formalized repair for entailment gate: " + err.Error()}
	}
	return applyArgumentEntailmentGate(result, *repair.Argument, *repair.LineMapping)
}

func applyArgumentEntailmentGate(result Result, argument string, mappings []FormalizationLineMapping) Result {
	report, gate := CheckEntailmentGate(argument)
	result.Details["entailment_gate"] = gate.Details
	if gate.Passed {
		if message := repairAssociatedConditional(argument, mappings); message != "" {
			result.Passed = false
			result.ParsedOutput = nil
			result.Message = message
		}
		return result
	}
	messages := make([]string, 0, 3)
	for _, step := range report.Steps {
		if !step.Passed {
			messages = append(messages, fmt.Sprintf("entailment gate failed at step %d (cites %v): %s", step.Step, step.Citations, step.Mismatch)+repairClauseDiffHint(step.Mismatch))
			if len(messages) == 3 {
				break
			}
		}
	}
	if len(messages) == 0 {
		messages = append(messages, "entailment gate failed: "+gate.Message)
	}
	result.Passed = false
	result.ParsedOutput = nil
	result.Message = strings.Join(messages, "; ") + ". Rephrase so each premise clause matches the clause it instantiates word-for-word, using the canonical forms (If A, then B. / Either A or B. / It is not the case that A. / Every S is P.). Do not change the argument's content."
	result.Details["entailment_gate_message"] = result.Message
	return result
}

// Shared rule with web/app/_components/argumentDisplay.ts: a supplied premise
// is padding when its antecedent is exactly the set of other cited statements
// and its consequent is the citing conclusion. For a single other citation,
// require a mapped premise, the FINAL conclusion, and only one supplied premise.
// This is a repair-content check after validity passes, not a validity rule.
func repairAssociatedConditional(text string, mappings []FormalizationLineMapping) string {
	for _, bridge := range repairAssociatedConditionalSteps(text, mappings) {
		return fmt.Sprintf("step %d supplies an associated conditional (%q): its antecedent just restates the steps it is cited with and its consequent is the conclusion, so it adds no content. Supply the general principle the author relies on (a universal 'Every S is P.' or an 'If X …, then … X …' rule, S8), or omit the bridge if there is none.", bridge.Number, bridge.Statement)
	}
	return ""
}

func repairAssociatedConditionalSteps(text string, mappings []FormalizationLineMapping) []NumberedArgumentStep {
	var found []NumberedArgumentStep
	argument, _ := ParseFlatNumberedArgumentAllowingNoConclusions(text)
	mapped := make(map[int]bool)
	for _, mapping := range mappings {
		if len(mapping.FromAsFoundLines) > 0 || mapping.FromSourceQuote != nil {
			mapped[mapping.RepairedLine] = true
		}
	}
	byNumber := make(map[int]NumberedArgumentStep)
	suppliedCount := 0
	for _, step := range argument.Steps {
		byNumber[step.Number] = step
		if step.Kind == "premise" && !mapped[step.Number] {
			suppliedCount++
		}
	}
bridges:
	for _, bridge := range argument.Steps {
		if bridge.Kind != "premise" || mapped[bridge.Number] {
			continue
		}
		for _, step := range argument.Steps {
			citesBridge := false
			others := make(map[string]bool)
			otherNumbers := make(map[int]bool)
			for _, number := range step.Citations {
				if number == bridge.Number {
					citesBridge = true
				} else {
					others[comparisonClause(byNumber[number].Statement)] = true
					otherNumbers[number] = true
				}
			}
			if !citesBridge || len(others) == 0 {
				continue
			}
			if len(otherNumbers) == 1 {
				if step.Number != argument.Steps[len(argument.Steps)-1].Number || suppliedCount != 1 {
					continue
				}
				for number := range otherNumbers {
					if !mapped[number] || byNumber[number].Kind != "premise" {
						citesBridge = false
					}
				}
				if !citesBridge {
					continue
				}
			}
			for _, conditional := range conditionalCandidates(bridge.Statement) {
				if !equalClauses(conditional.right, step.Statement) {
					continue
				}
				candidates := [][]string{{conditional.left}}
				if len(others) > 1 {
					candidates = conjunctiveAntecedentCandidates(conditional.left, len(others))
				}
				for _, parts := range candidates {
					matched := make(map[string]bool)
					for _, part := range parts {
						matched[comparisonClause(part)] = true
					}
					if len(matched) != len(others) {
						continue
					}
					equal := true
					for part := range matched {
						equal = equal && others[part]
					}
					if equal {
						found = append(found, bridge)
						continue bridges
					}
				}
			}
		}
	}
	return found
}

func (v formalizedRepairOutputValidator) Validate(ctx context.Context, output string) (result Result) {
	defer func() {
		result = captureRepairFailure(applyRepairEntailmentGate(result, v.options.EntailmentGate), output)
	}()
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
	if _, flat := ParseFlatNumberedArgumentAllowingNoConclusions(*repair.Argument); !flat.Passed {
		return flat
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
			parsed, _ := ParseFlatNumberedArgumentAllowingNoConclusions(*repair.Argument)
			// Without prompt context, still check length and lexical alignment. Source
			// membership is checked by the prompt-aware validator and by assembly.
			if _, _, err := ValidateSourceQuote(*mapping.FromSourceQuote, parsed.Steps[mapping.RepairedLine-1].Statement, *mapping.FromSourceQuote); err != nil {
				return Result{Passed: false, Message: fmt.Sprintf("repaired line %d: %v", mapping.RepairedLine, err) + sourceQuoteSentenceSuggestion(v.source, *mapping.FromSourceQuote)}
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
// the assembly transform performs the defensive dependency-backed check. An
// optional builtin options object enables the gate after coverage passes.
func ValidateFormalizedRepairOutputWithPrompt(ctx context.Context, output, prompt string, options ...json.RawMessage) (result Result) {
	var gateEnabled bool
	defer func() { result = captureRepairFailure(applyRepairEntailmentGate(result, gateEnabled), output) }()
	if len(options) > 1 {
		return Result{Message: "formalized_repair_output accepts at most one options value"}
	}
	if len(options) == 1 {
		parsed, err := parseFormalizedRepairOutputOptions(options[0])
		if err != nil {
			return Result{Message: err.Error()}
		}
		gateEnabled = parsed.EntailmentGate
	}
	// Defer the gate until all prompt-aware coverage checks have passed.
	source, sourceOK := formalizedRepairTaggedValue(prompt, "source_material")
	result = formalizedRepairOutputValidator{source: source}.Validate(ctx, output)
	if !result.Passed {
		return result
	}
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
	if _, parsed := ParseFlatNumberedArgumentAllowingNoConclusions(source); parsed.Passed {
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

	repairedArgument, parsed := ParseFlatNumberedArgumentAllowingNoConclusions(*repair.Argument)
	if !parsed.Passed {
		return Result{Passed: false, Message: "parse formalized repair argument for coverage: " + parsed.Message}
	}
	suppliedLines := make([]int, 0)
	for _, step := range repairedArgument.Steps {
		if step.Kind != DroppedContentKindPremise || mappedRepaired[step.Number] {
			continue
		}
		if formalizedRepairStatementPresent(step.Statement, source) || formalizedRepairStatementPresent(step.Statement, asFound) {
			return Result{Passed: false, Message: unmappedRepairMessage(step, source, asFound, expectedAnchor)}
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
		if count == 64000 {
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

func unmappedRepairMessage(step NumberedArgumentStep, source, asFound, anchor string) string {
	// A source-numbered anchor changes what from_as_found_lines indexes.
	anchorText, label := asFound, "as-found"
	if anchor == FormalizationCoverageAnchorSource {
		anchorText, label = source, "source"
	}
	if parsed, result := ParseFlatNumberedArgumentAllowingNoConclusions(anchorText); result.Passed {
		for _, original := range parsed.Steps {
			if formalizedRepairStatementPresent(step.Statement, original.Statement) {
				return fmt.Sprintf("repaired premise line %d restates part of %s line %d (%q); add %d to its from_as_found_lines (a split premise may map to several repaired lines)", step.Number, label, original.Number, original.Statement, original.Number)
			}
		}
	}
	for _, sentence := range sourceSentences(source) {
		if formalizedRepairStatementPresent(step.Statement, sentence) {
			return fmt.Sprintf("repaired premise line %d restates part of a source sentence; add from_source_quote: %q", step.Number, sentence)
		}
	}
	// Preserve rejection even for normalized matches spanning sentence boundaries.
	return fmt.Sprintf("repaired premise line %d is present in the source/as-found input and must be mapped", step.Number)
}

// Gate mismatches already carry Go-quoted clauses; enrich only that diagnostic
// shape, leaving the gate's selection and verdict completely untouched.
var repairClauseMismatch = regexp.MustCompile(`("(?:[^"\\]|\\.)*") does not match [^"\n]+ ("(?:[^"\\]|\\.)*") verbatim`)

var repairNearestClauseMismatch = regexp.MustCompile(`("(?:[^"\\]|\\.)*") is not matched verbatim by any cited premise; nearest is step [0-9]+ ("(?:[^"\\]|\\.)*")`)

func repairClauseDiffHint(mismatch string) string {
	match := repairClauseMismatch.FindStringSubmatch(mismatch)
	if match == nil {
		match = repairNearestClauseMismatch.FindStringSubmatch(mismatch)
	}
	if match == nil {
		return ""
	}
	left, errLeft := strconv.Unquote(match[1])
	right, errRight := strconv.Unquote(match[2])
	if errLeft != nil || errRight != nil {
		return ""
	}
	// Count words to retain repeated-word differences. Order-only mismatches get
	// an explicit hint rather than an empty diff.
	difference := func(a, b string) []string {
		counts := map[string]int{}
		for _, word := range strings.Fields(b) {
			counts[word]++
		}
		var extra []string
		for _, word := range strings.Fields(a) {
			if counts[word] > 0 {
				counts[word]--
			} else {
				extra = append(extra, word)
			}
		}
		return extra
	}
	hint := fmt.Sprintf("\nfirst clause: %q\nsecond clause: %q", left, right)
	extra, missing := difference(left, right), difference(right, left)
	if len(extra) > 0 {
		hint += "\nfirst clause extra: " + strings.Join(extra, ", ")
	}
	if len(missing) > 0 {
		hint += "\nfirst clause missing: " + strings.Join(missing, ", ")
	}
	if len(extra) == 0 && len(missing) == 0 {
		hint += "\nword order differs"
	}
	return hint
}
