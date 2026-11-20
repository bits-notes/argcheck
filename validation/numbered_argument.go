package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxNumberedArgumentLines       = 2000
	maxCitationsPerConclusion      = 100
	numberedArgumentStepAssumption = "assumption"
	numberedArgumentStepArbitrary  = "arbitrary"
	numberedArgumentStepPremise    = "premise"
	numberedArgumentStepConclusion = "conclusion"
)

var (
	premiseLinePattern      = regexp.MustCompile(`^([1-9][0-9]*)\. (.+)\.$`)
	conclusionLinePattern   = regexp.MustCompile(`^([1-9][0-9]*)\. (.+)\. \[([1-9][0-9]*(?:-[1-9][0-9]*)?(?:,[1-9][0-9]*(?:-[1-9][0-9]*)?)*)\]$`)
	conclusionMarkerPattern = regexp.MustCompile(`(?i)^(therefore|so|thus|hence)(,\s*|\s+)`)
	numberedLineBodyPattern = regexp.MustCompile(`^[1-9][0-9]*\. (.*)$`)
	bracketReferencePattern = regexp.MustCompile(`\[([^\]]+)\]`)
	numericReferencePattern = regexp.MustCompile(`^[1-9][0-9]*(?:,[1-9][0-9]*)*$`)
	inlineStepBridgePattern = regexp.MustCompile(`^If ((?:\[[1-9][0-9]*(?:,[1-9][0-9]*)*\](?:,? and )?)+)(?:,? then |, ).+`)
	// v31.1: a bridge may also name earlier steps as its whole consequent.
	inlineStepConsequentPattern = regexp.MustCompile(`^If (.+?),? then ((?:\[[1-9][0-9]*(?:,[1-9][0-9]*)*\](?:,? and )?)+)\.?$`)
	fenceLinePattern            = regexp.MustCompile("^```[A-Za-z0-9_-]*$")
)

type NumberedArgument struct {
	Steps []NumberedArgumentStep `json:"steps"`
}

type NumberedArgumentStep struct {
	Number    int      `json:"number"`
	Kind      string   `json:"kind"`
	Statement string   `json:"statement"`
	Citations []int    `json:"citations,omitempty"`
	Line      string   `json:"line"`
	Depth     int      `json:"depth,omitempty"`
	Ranges    [][2]int `json:"ranges,omitempty"`
	// relaxIndefinite is set only on copies used to diagnose a refusal: it
	// lets schematic unification bind an indefinite in an object slot.
	relaxIndefinite bool
	// relaxClauseTerm likewise lets a term that reads as containing a clause
	// bind a variable, only to name it in the refusal message.
	relaxClauseTerm bool
}

type numberedArgumentOptions struct {
	AllowSingleFencedBlock bool `json:"allow_single_fenced_block"`
	AllowBlankLines        bool `json:"allow_blank_lines"`
	AllowDisconnectedSteps bool `json:"allow_disconnected_steps"`
	allowNoConclusions     bool
}

type numberedArgumentSyntaxValidator struct {
	options numberedArgumentOptions
}

func parseNumberedArgumentOptions(raw json.RawMessage) (numberedArgumentOptions, error) {
	if len(raw) == 0 {
		return numberedArgumentOptions{}, nil
	}
	var options numberedArgumentOptions
	if err := json.Unmarshal(raw, &options); err != nil {
		return numberedArgumentOptions{}, fmt.Errorf("invalid numbered_argument_syntax options: %w", err)
	}
	return options, nil
}

func (v numberedArgumentSyntaxValidator) Validate(ctx context.Context, output string) Result {
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

	argument, result := ParseNumberedArgument(output, v.options)
	if !result.Passed {
		return result
	}
	if unused := NumberedArgumentUnusedSteps(argument); len(unused) > 0 && !v.options.AllowDisconnectedSteps {
		return Result{
			Passed: false,
			Message: fmt.Sprintf(
				"steps %v do not support the final conclusion; every premise must feed a conclusion and every intermediate conclusion must be cited",
				unused,
			),
			Details: map[string]any{"unused_steps": unused},
		}
	}
	return Result{
		Passed:  true,
		Message: "output matches numbered_argument_syntax",
		Details: map[string]any{
			"step_count": len(argument.Steps),
		},
	}
}

func ParseNumberedArgument(output string, options numberedArgumentOptions) (NumberedArgument, Result) {
	return allStrictRules().ParseNumberedArgument(output, options)
}

func (policy strictRules) ParseNumberedArgument(output string, options numberedArgumentOptions) (NumberedArgument, Result) {
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return NumberedArgument{}, Result{Passed: false, Message: "output is empty"}
	}

	var unwrapResult Result
	normalized, unwrapResult = unwrapOptionalFence(normalized, options)
	if !unwrapResult.Passed {
		return NumberedArgument{}, unwrapResult
	}

	lines := strings.Split(normalized, "\n")
	if len(lines) > maxNumberedArgumentLines {
		return NumberedArgument{}, Result{
			Passed:  false,
			Message: fmt.Sprintf("numbered argument exceeds %d line validation limit", maxNumberedArgumentLines),
		}
	}

	argument := NumberedArgument{Steps: make([]NumberedArgumentStep, 0, len(lines))}
	seenPremise := false
	seenConclusion := false
	expectedNumber := 1

	for index, line := range lines {
		lineNumber := index + 1
		if strings.TrimSpace(line) == "" {
			if options.AllowBlankLines {
				continue
			}
			return NumberedArgument{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d is blank; numbered arguments cannot contain blank lines", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		if line != strings.TrimSpace(line) {
			return NumberedArgument{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d has leading or trailing whitespace", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}

		step, result := policy.parseNumberedArgumentLine(line, lineNumber)
		if !result.Passed {
			return NumberedArgument{}, result
		}
		if step.Number != expectedNumber {
			return NumberedArgument{}, Result{
				Passed: false,
				Message: fmt.Sprintf(
					"numbering jumps from %d to %d",
					expectedNumber-1,
					step.Number,
				),
				Details: map[string]any{"line": lineNumber, "expected": expectedNumber, "actual": step.Number},
			}
		}
		if step.Kind == numberedArgumentStepPremise {
			seenPremise = true
		}
		if step.Kind == numberedArgumentStepConclusion {
			seenConclusion = true
			if len(step.Citations)+len(step.Ranges) == 0 {
				return NumberedArgument{}, Result{
					Passed:  false,
					Message: fmt.Sprintf("step %d conclusion must cite at least one prior step", step.Number),
					Details: map[string]any{"line": lineNumber, "step": step.Number},
				}
			}
			if len(step.Citations)+len(step.Ranges) > maxCitationsPerConclusion {
				return NumberedArgument{}, Result{
					Passed:  false,
					Message: fmt.Sprintf("step %d exceeds %d citation validation limit", step.Number, maxCitationsPerConclusion),
					Details: map[string]any{"line": lineNumber, "step": step.Number},
				}
			}
			for _, citation := range step.Citations {
				if citation >= step.Number {
					return NumberedArgument{}, Result{
						Passed:  false,
						Message: fmt.Sprintf("step %d cites current or future step %d", step.Number, citation),
						Details: map[string]any{"line": lineNumber, "step": step.Number, "citation": citation},
					}
				}
			}
		}

		argument.Steps = append(argument.Steps, step)
		expectedNumber++
	}

	if len(argument.Steps) < 2 {
		return NumberedArgument{}, Result{Passed: false, Message: "numbered argument must contain at least two steps"}
	}
	if !seenPremise && !argument.HasSubproofs() {
		return NumberedArgument{}, Result{Passed: false, Message: "numbered argument must contain at least one premise"}
	}
	if !seenConclusion && !options.allowNoConclusions {
		return NumberedArgument{}, Result{Passed: false, Message: "numbered argument must contain at least one conclusion"}
	}
	if seenConclusion && argument.Steps[len(argument.Steps)-1].Kind != numberedArgumentStepConclusion {
		return NumberedArgument{}, Result{Passed: false, Message: "final step must be a conclusion"}
	}
	if result := checkNumberedArgumentScope(argument); !result.Passed {
		return NumberedArgument{}, result
	}
	return argument, Result{Passed: true}
}

// ParseNumberedArgumentAllowingNoConclusions applies the numbered-argument
// parser while permitting a premise-only argument. All other syntax and
// structural checks remain unchanged. It accepts flat arguments only: its
// callers predate Fitch subproofs and drop scope and discharge edges. Use
// ParseScopedNumberedArgumentAllowingNoConclusions where subproofs are handled.
func ParseNumberedArgumentAllowingNoConclusions(output string) (NumberedArgument, Result) {
	return allStrictRules().ParseNumberedArgumentAllowingNoConclusions(output)
}

func (policy strictRules) ParseNumberedArgumentAllowingNoConclusions(output string) (NumberedArgument, Result) {
	arg, result := policy.ParseScopedNumberedArgumentAllowingNoConclusions(output)
	if result.Passed && arg.HasSubproofs() {
		return NumberedArgument{}, Result{Passed: false, Message: "Fitch subproofs are not supported by this flat-argument transformation"}
	}
	return arg, result
}

// ParseScopedNumberedArgumentAllowingNoConclusions is
// ParseNumberedArgumentAllowingNoConclusions with Fitch subproofs permitted.
func ParseScopedNumberedArgumentAllowingNoConclusions(output string) (NumberedArgument, Result) {
	return allStrictRules().ParseScopedNumberedArgumentAllowingNoConclusions(output)
}

func (policy strictRules) ParseScopedNumberedArgumentAllowingNoConclusions(output string) (NumberedArgument, Result) {
	return policy.ParseNumberedArgument(output, numberedArgumentOptions{allowNoConclusions: true})
}

func unwrapOptionalFence(output string, options numberedArgumentOptions) (string, Result) {
	if !strings.HasPrefix(output, "```") {
		return output, Result{Passed: true}
	}
	if !options.AllowSingleFencedBlock {
		return "", Result{Passed: false, Message: "fenced code blocks are not allowed for this validator"}
	}
	lines := strings.Split(output, "\n")
	if len(lines) < 3 || !fenceLinePattern.MatchString(lines[0]) || strings.TrimSpace(lines[len(lines)-1]) != "```" {
		return "", Result{Passed: false, Message: "output must contain exactly one fenced numbered argument block"}
	}
	for _, line := range lines[1 : len(lines)-1] {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			return "", Result{Passed: false, Message: "output contains an extra code fence inside the argument block"}
		}
	}
	content := strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	if content == "" {
		return "", Result{Passed: false, Message: "fenced numbered argument block is empty"}
	}
	return content, Result{Passed: true}
}

func parseNumberedArgumentLine(line string, lineNumber int) (NumberedArgumentStep, Result) {
	return allStrictRules().parseNumberedArgumentLine(line, lineNumber)
}

func (policy strictRules) parseNumberedArgumentLine(line string, lineNumber int) (NumberedArgumentStep, Result) {
	original := line
	depth := 0
	for strings.HasPrefix(line, "| ") {
		depth++
		line = strings.TrimPrefix(line, "| ")
	}
	step, result := policy.parseNumberedArgumentBareLine(line, lineNumber)
	if !result.Passed {
		return step, result
	}
	if grouping := policy.validateStatementGrouping(strings.TrimPrefix(step.Statement, "Suppose "), lineNumber); !grouping.Passed {
		rule := "syntax"
		if strings.Contains(grouping.Message, "and/or") || strings.Contains(grouping.Message, "top-level unless") {
			rule = "scope.mixed_andor"
		}
		refusal := &StepRefusal{Span: wholeRefusalSpan(step), Rule: rule, Hint: grouping.Message, RuleSource: "syntax"}
		if _, hint, ok := strings.Cut(grouping.Message, "; "); ok {
			refusal.Hint = hint
		}
		grouping.Details["refusal"] = refusal
		grouping.Details["steps"] = []EntailmentGateStepReport{{Step: step.Number, NearestForm: "syntax", Mismatch: grouping.Message, Refusal: refusal}}
		return NumberedArgumentStep{}, grouping
	}
	step.Depth, step.Line = depth, original
	if step.Kind == numberedArgumentStepPremise {
		if strings.HasPrefix(step.Statement, "Suppose ") {
			step.Kind = numberedArgumentStepAssumption
			step.Statement = strings.TrimPrefix(step.Statement, "Suppose ")
		} else if arbitraryNamePattern.MatchString(step.Statement) || arbitraryFOLPattern.MatchString(step.Statement) {
			step.Kind = numberedArgumentStepArbitrary
		}
	}
	return step, result
}

// stripConclusionMarker is shared by syntax and gate parsing. A comma or
// whitespace boundary prevents words such as "Sophie" from becoming markers.
func stripConclusionMarker(statement string) (string, string) {
	match := conclusionMarkerPattern.FindStringSubmatch(statement)
	if match == nil {
		return "", statement
	}
	body := statement[len(match[0]):]
	// Preserve the existing distinction between inferential "So" and degree
	// phrases such as "So many" when there is no separating comma.
	if strings.EqualFold(match[1], "so") && !strings.HasPrefix(match[2], ",") {
		for _, prefix := range []string{"many ", "much ", "few ", "little ", "far ", "long ", "that ", "called "} {
			if strings.HasPrefix(strings.ToLower(body), prefix) {
				return "", statement
			}
		}
	}
	return match[1], body
}

func (policy strictRules) parseNumberedArgumentBareLine(line string, lineNumber int) (NumberedArgumentStep, Result) {
	if matches := conclusionLinePattern.FindStringSubmatch(line); matches != nil {
		number, _ := strconv.Atoi(matches[1])
		citations, ranges, err := parseTerminalCitations(matches[3])
		if err != nil {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d has invalid citations: %v", lineNumber, err),
				Details: map[string]any{"line": lineNumber},
			}
		}
		_, body := stripConclusionMarker(matches[2])
		statement := strings.TrimSpace(body)
		if statement == "" {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d conclusion statement is empty", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		if result := policy.validateStatementBracketReferences(statement, lineNumber, number); !result.Passed {
			return NumberedArgumentStep{}, result
		}
		return NumberedArgumentStep{
			Number:    number,
			Kind:      numberedArgumentStepConclusion,
			Statement: statement,
			Citations: citations,
			Ranges:    ranges,
			Line:      line,
		}, Result{Passed: true}
	}

	// Diagnose markers even when malformed citations prevent a premise match.
	if body := numberedLineBodyPattern.FindStringSubmatch(line); body != nil {
		if marker, _ := stripConclusionMarker(body[1]); marker != "" {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d starts with %s, but does not have valid terminal citations: end the line with a period, a space and the cited step numbers in square brackets, e.g. \"... wet. [1,2]\"", lineNumber, marker),
				Details: map[string]any{"line": lineNumber},
			}
		}
	}

	if matches := premiseLinePattern.FindStringSubmatch(line); matches != nil {
		number, _ := strconv.Atoi(matches[1])
		statement := strings.TrimSpace(matches[2])
		if statement == "" {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d premise statement is empty", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		if result := policy.validateStatementBracketReferences(strings.TrimPrefix(statement, "Suppose "), lineNumber, number); !result.Passed {
			return NumberedArgumentStep{}, result
		}
		return NumberedArgumentStep{
			Number:    number,
			Kind:      numberedArgumentStepPremise,
			Statement: statement,
			Line:      line,
		}, Result{Passed: true}
	}

	return NumberedArgumentStep{}, Result{
		Passed:  false,
		Message: fmt.Sprintf("line %d is not a valid premise or conclusion%s", lineNumber, numberedLineHint(line)),
		Details: map[string]any{"line": lineNumber},
	}
}

// numberedLineHint says what to fix in a line that is neither a premise nor
// a conclusion, for the common slips.
func numberedLineHint(line string) string {
	body := numberedLineBodyPattern.FindStringSubmatch(line)
	if body == nil {
		return `: start the line with its step number, a period and a space, e.g. "1. "`
	}
	text := strings.TrimSpace(body[1])
	switch {
	case text == "":
		return ": the line has a step number but no sentence"
	case strings.HasSuffix(text, "]"):
		return `: end the sentence with a period before the citations, e.g. "... wet. [1,2]"`
	case !strings.HasSuffix(text, "."):
		return ": end the sentence with a period"
	}
	return ""
}

func (policy strictRules) validateStatementBracketReferences(statement string, lineNumber int, stepNumber int) Result {
	matches := bracketReferencePattern.FindAllStringSubmatch(statement, -1)
	if len(matches) == 0 {
		return Result{Passed: true}
	}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		if !numericReferencePattern.MatchString(match[1]) {
			return Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d contains non-numeric bracket reference %q; formal numbered arguments may only use bracketed step references like [1] inside statements", lineNumber, match[0]),
				Details: map[string]any{"line": lineNumber, "reference": match[0]},
			}
		}
	}
	if !policy.statementAllowsInlineStepReferences(statement) {
		return Result{
			Passed:  false,
			Message: fmt.Sprintf("line %d contains inline bracket references outside an explicit bridge premise", lineNumber),
			Details: map[string]any{"line": lineNumber},
		}
	}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		references, err := parseCitations(match[1])
		if err != nil {
			return Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d contains invalid bracketed step reference %q", lineNumber, match[0]),
				Details: map[string]any{"line": lineNumber, "reference": match[0]},
			}
		}
		for _, reference := range references {
			if reference >= stepNumber {
				return Result{
					Passed:  false,
					Message: fmt.Sprintf("line %d contains bracketed step reference %q that does not refer to a prior step", lineNumber, match[0]),
					Details: map[string]any{"line": lineNumber, "reference": match[0], "step": stepNumber},
				}
			}
		}
	}
	return Result{Passed: true}
}

func (policy strictRules) statementAllowsInlineStepReferences(statement string) bool {
	statement = strings.TrimSpace(statement)
	// v31.2: "C if [1]." is the postposed form of "If [1], then C.".
	if rewritten, ok := policy.postposedConditional(strings.TrimSuffix(statement, ".")); ok {
		statement = rewritten
	}
	total := len(bracketReferencePattern.FindAllStringSubmatch(statement, -1))
	// References may fill the whole consequent; the antecedent is then plain
	// text or itself a whole-antecedent reference list.
	if c := inlineStepConsequentPattern.FindStringSubmatch(statement); c != nil {
		inConsequent := len(bracketReferencePattern.FindAllStringSubmatch(c[2], -1))
		inAntecedent := len(bracketReferencePattern.FindAllStringSubmatch(c[1], -1))
		if inAntecedent == 0 {
			return total == inConsequent
		}
		if a := inlineStepBridgePattern.FindStringSubmatch("If " + c[1] + ", then x"); a != nil {
			return total == inConsequent+len(bracketReferencePattern.FindAllStringSubmatch(a[1], -1)) &&
				len(bracketReferencePattern.FindAllStringSubmatch(a[1], -1)) == inAntecedent
		}
		return false
	}
	matches := inlineStepBridgePattern.FindStringSubmatch(statement)
	if len(matches) < 2 {
		return false
	}
	return total == len(bracketReferencePattern.FindAllStringSubmatch(matches[1], -1))
}

func parseCitations(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	citations := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		citations = append(citations, n)
	}
	return citations, nil
}
