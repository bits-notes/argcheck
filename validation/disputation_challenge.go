package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	challengeForceFalseAsWritten = "false_as_written"
	targetSideAffirmative        = "affirmative"
	targetSideNegative           = "negative"
)

type disputationChallengeOptions struct {
	AllowSingleFencedBlock bool `json:"allow_single_fenced_block"`
}

type disputationChallengeValidator struct {
	options disputationChallengeOptions
}

type disputationChallengeOutput struct {
	TargetSide          string `json:"target_side"`
	TargetArgumentLabel string `json:"target_argument_label"`
	TargetPremiseNumber int    `json:"target_premise_number"`
	TargetPremiseText   string `json:"target_premise_text"`
	ChallengeForce      string `json:"challenge_force"`
	Argument            string `json:"argument"`
}

var weakChallengeConclusionPatterns = []string{
	"not necessarily",
	"not false as written",
	"may be false as written",
	"might be false as written",
	"could be false as written",
	"possibly false as written",
}

var challengeConclusionWhitespacePattern = regexp.MustCompile(`\s+`)

func parseDisputationChallengeOptions(raw json.RawMessage) (disputationChallengeOptions, error) {
	if len(raw) == 0 {
		return disputationChallengeOptions{}, nil
	}
	var options disputationChallengeOptions
	if err := json.Unmarshal(raw, &options); err != nil {
		return disputationChallengeOptions{}, fmt.Errorf("invalid disputation_challenge options: %w", err)
	}
	return options, nil
}

func (v disputationChallengeValidator) Validate(ctx context.Context, output string) Result {
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

	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return Result{Passed: false, Message: "output is empty"}
	}

	unwrapped, unwrapResult := unwrapJSONOutput(normalized, jsonSchemaOptions{
		AllowSingleFencedBlock: v.options.AllowSingleFencedBlock,
	})
	if !unwrapResult.Passed {
		return unwrapResult
	}

	decoder := json.NewDecoder(strings.NewReader(unwrapped))
	decoder.DisallowUnknownFields()
	var challenge disputationChallengeOutput
	if err := decoder.Decode(&challenge); err != nil {
		return Result{Passed: false, Message: fmt.Sprintf("output is not a valid disputation challenge JSON object: %v", err)}
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return Result{Passed: false, Message: "output must contain exactly one JSON value"}
	} else if err != io.EOF {
		return Result{Passed: false, Message: "output must contain only the disputation challenge JSON object"}
	}

	if result := validateDisputationChallengeMetadata(challenge); !result.Passed {
		return result
	}

	argument, argumentResult := ParseNumberedArgument(challenge.Argument, numberedArgumentOptions{})
	if !argumentResult.Passed {
		return Result{
			Passed:  false,
			Message: "challenge argument is invalid: " + argumentResult.Message,
			Details: argumentResult.Details,
		}
	}
	finalStep := argument.Steps[len(argument.Steps)-1]
	if result := validateChallengeFinalConclusion(challenge, finalStep.Statement); !result.Passed {
		return result
	}

	raw, err := canonicalJSON(challenge)
	if err != nil {
		raw = json.RawMessage(unwrapped)
	}
	return Result{
		Passed:  true,
		Message: "output matches disputation_challenge",
		Details: map[string]any{
			"step_count":             len(argument.Steps),
			"target_side":            challenge.TargetSide,
			"target_argument_label":  challenge.TargetArgumentLabel,
			"target_premise_number":  challenge.TargetPremiseNumber,
			"challenge_force":        challenge.ChallengeForce,
			"final_conclusion_force": challengeForceFalseAsWritten,
		},
		ParsedOutput: raw,
	}
}

func validateDisputationChallengeMetadata(challenge disputationChallengeOutput) Result {
	side := strings.ToLower(strings.TrimSpace(challenge.TargetSide))
	switch side {
	case targetSideAffirmative, targetSideNegative:
	default:
		return Result{Passed: false, Message: `target_side must be "affirmative" or "negative"`}
	}
	label := strings.TrimSpace(challenge.TargetArgumentLabel)
	if label == "" {
		return Result{Passed: false, Message: "target_argument_label is required"}
	}
	upperLabel := strings.ToUpper(label)
	if side == targetSideAffirmative && !strings.HasPrefix(upperLabel, "A") {
		return Result{Passed: false, Message: "affirmative challenges must target an A* argument label"}
	}
	if side == targetSideNegative && !strings.HasPrefix(upperLabel, "N") {
		return Result{Passed: false, Message: "negative challenges must target an N* argument label"}
	}
	if challenge.TargetPremiseNumber < 1 {
		return Result{Passed: false, Message: "target_premise_number must be positive"}
	}
	if strings.TrimSpace(challenge.TargetPremiseText) == "" {
		return Result{Passed: false, Message: "target_premise_text is required"}
	}
	if strings.TrimSpace(challenge.ChallengeForce) != challengeForceFalseAsWritten {
		return Result{Passed: false, Message: `challenge_force must be "false_as_written"`}
	}
	if strings.TrimSpace(challenge.Argument) == "" {
		return Result{Passed: false, Message: "argument is required"}
	}
	return Result{Passed: true}
}

func validateChallengeFinalConclusion(challenge disputationChallengeOutput, statement string) Result {
	lower := strings.ToLower(statement)
	lower = challengeConclusionWhitespacePattern.ReplaceAllString(strings.TrimSpace(lower), " ")
	for _, weakPattern := range weakChallengeConclusionPatterns {
		if strings.Contains(lower, weakPattern) {
			return Result{Passed: false, Message: "challenge final conclusion is too weak; it must conclude that the target premise is false as written"}
		}
	}
	allowedConclusions := map[string]struct{}{
		"target premise is false as written":                                         {},
		"the target premise is false as written":                                     {},
		fmt.Sprintf("premise %d is false as written", challenge.TargetPremiseNumber): {},
	}
	if _, ok := allowedConclusions[lower]; !ok {
		if strings.Contains(lower, "false as written") {
			return Result{Passed: false, Message: "challenge final conclusion must assert exactly that the target premise is false as written"}
		}
		return Result{Passed: false, Message: "challenge final conclusion must state that the target premise is false as written"}
	}
	return Result{Passed: true}
}
