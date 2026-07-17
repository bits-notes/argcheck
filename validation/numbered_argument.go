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
	numberedArgumentStepPremise    = "premise"
	numberedArgumentStepConclusion = "conclusion"
)

var (
	premiseLinePattern      = regexp.MustCompile(`^([1-9][0-9]*)\. (.+)\.$`)
	conclusionLinePattern   = regexp.MustCompile(`^([1-9][0-9]*)\. Therefore, (.+)\. \[([1-9][0-9]*(?:,[1-9][0-9]*)*)\]$`)
	terminalCitationPattern = regexp.MustCompile(`\. \[[1-9][0-9]*(?:,[1-9][0-9]*)*\]$`)
	bracketReferencePattern = regexp.MustCompile(`\[([^\]]+)\]`)
	numericReferencePattern = regexp.MustCompile(`^[1-9][0-9]*(?:,[1-9][0-9]*)*$`)
	inlineStepBridgePattern = regexp.MustCompile(`^If ((?:\[[1-9][0-9]*(?:,[1-9][0-9]*)*\](?: and )?)+)(?:,? then |, ).+`)
	fenceLinePattern        = regexp.MustCompile("^```[A-Za-z0-9_-]*$")
)

type NumberedArgument struct {
	Steps []NumberedArgumentStep `json:"steps"`
}

type NumberedArgumentStep struct {
	Number    int    `json:"number"`
	Kind      string `json:"kind"`
	Statement string `json:"statement"`
	Citations []int  `json:"citations,omitempty"`
	Line      string `json:"line"`
}

type numberedArgumentOptions struct {
	AllowSingleFencedBlock bool `json:"allow_single_fenced_block"`
	AllowBlankLines        bool `json:"allow_blank_lines"`
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
	return Result{
		Passed:  true,
		Message: "output matches numbered_argument_syntax",
		Details: map[string]any{
			"step_count": len(argument.Steps),
		},
	}
}

func ParseNumberedArgument(output string, options numberedArgumentOptions) (NumberedArgument, Result) {
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

		step, result := parseNumberedArgumentLine(line, lineNumber)
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
			if len(step.Citations) == 0 {
				return NumberedArgument{}, Result{
					Passed:  false,
					Message: fmt.Sprintf("step %d conclusion must cite at least one prior step", step.Number),
					Details: map[string]any{"line": lineNumber, "step": step.Number},
				}
			}
			if len(step.Citations) > maxCitationsPerConclusion {
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
	if !seenPremise {
		return NumberedArgument{}, Result{Passed: false, Message: "numbered argument must contain at least one premise"}
	}
	if !seenConclusion && !options.allowNoConclusions {
		return NumberedArgument{}, Result{Passed: false, Message: "numbered argument must contain at least one conclusion"}
	}
	if seenConclusion && argument.Steps[len(argument.Steps)-1].Kind != numberedArgumentStepConclusion {
		return NumberedArgument{}, Result{Passed: false, Message: "final step must be a conclusion"}
	}
	return argument, Result{Passed: true}
}

// ParseNumberedArgumentAllowingNoConclusions applies the numbered-argument
// parser while permitting a premise-only argument. All other syntax and
// structural checks remain unchanged.
func ParseNumberedArgumentAllowingNoConclusions(output string) (NumberedArgument, Result) {
	return ParseNumberedArgument(output, numberedArgumentOptions{allowNoConclusions: true})
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
	if matches := conclusionLinePattern.FindStringSubmatch(line); matches != nil {
		number, _ := strconv.Atoi(matches[1])
		citations, err := parseCitations(matches[3])
		if err != nil {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d has invalid citations: %v", lineNumber, err),
				Details: map[string]any{"line": lineNumber},
			}
		}
		statement := strings.TrimSpace(matches[2])
		if statement == "" {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d conclusion statement is empty", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		if result := validateStatementBracketReferences(statement, lineNumber, number); !result.Passed {
			return NumberedArgumentStep{}, result
		}
		return NumberedArgumentStep{
			Number:    number,
			Kind:      numberedArgumentStepConclusion,
			Statement: statement,
			Citations: citations,
			Line:      line,
		}, Result{Passed: true}
	}

	if matches := premiseLinePattern.FindStringSubmatch(line); matches != nil {
		number, _ := strconv.Atoi(matches[1])
		statement := strings.TrimSpace(matches[2])
		if strings.HasPrefix(statement, "Therefore,") {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d starts with Therefore, but does not have valid terminal citations", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		if terminalCitationPattern.MatchString(line) {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d has terminal citations but does not start with Therefore,", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		if statement == "" {
			return NumberedArgumentStep{}, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d premise statement is empty", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		if result := validateStatementBracketReferences(statement, lineNumber, number); !result.Passed {
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
		Message: fmt.Sprintf("line %d is not a valid premise or conclusion", lineNumber),
		Details: map[string]any{"line": lineNumber},
	}
}

func validateStatementBracketReferences(statement string, lineNumber int, stepNumber int) Result {
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
	if !statementAllowsInlineStepReferences(statement) {
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

func statementAllowsInlineStepReferences(statement string) bool {
	statement = strings.TrimSpace(statement)
	matches := inlineStepBridgePattern.FindStringSubmatch(statement)
	if len(matches) < 2 {
		return false
	}
	return len(bracketReferencePattern.FindAllStringSubmatch(statement, -1)) == len(bracketReferencePattern.FindAllStringSubmatch(matches[1], -1))
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
