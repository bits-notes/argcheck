package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// disputationProseResponse is the single-node "Build B" contract: an inbound
// prose intake result that classifies how a skeptic's prose engages a formal
// numbered-argument graph, localizes it, and — only for a premise attack —
// carries a fully structured disputation_challenge that must itself pass the
// deterministic disputation_challenge contract.
//
// The challenge sub-object is validated by delegating to the existing
// disputationChallengeValidator, so the "false as written" localization
// guarantee is enforced exactly once, in one place.
type disputationProseResponseValidator struct {
	options disputationChallengeOptions
}

const (
	proseEngagementAttacksPremise        = "attacks_premise"
	proseEngagementAttacksInferenceStep  = "attacks_inference_step"
	proseEngagementContestsRendering     = "contests_rendering"
	proseEngagementIntroducesNewEvidence = "introduces_new_evidence"
	proseEngagementDoesNotEngage         = "does_not_engage"
)

var proseResponseEngagements = map[string]struct{}{
	proseEngagementAttacksPremise:        {},
	proseEngagementAttacksInferenceStep:  {},
	proseEngagementContestsRendering:     {},
	proseEngagementIntroducesNewEvidence: {},
	proseEngagementDoesNotEngage:         {},
}

type disputationProseResponseOutput struct {
	Engagement            string          `json:"engagement"`
	TargetArgumentLabel   string          `json:"target_argument_label"`
	LocalizationStatement string          `json:"localization_statement"`
	Justification         string          `json:"justification"`
	Challenge             json.RawMessage `json:"challenge"`
}

func parseDisputationProseResponseOptions(raw json.RawMessage) (disputationChallengeOptions, error) {
	if len(raw) == 0 {
		return disputationChallengeOptions{}, nil
	}
	var options disputationChallengeOptions
	if err := json.Unmarshal(raw, &options); err != nil {
		return disputationChallengeOptions{}, fmt.Errorf("invalid disputation_prose_response options: %w", err)
	}
	return options, nil
}

func (v disputationProseResponseValidator) Validate(ctx context.Context, output string) Result {
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
	var response disputationProseResponseOutput
	if err := decoder.Decode(&response); err != nil {
		return Result{Passed: false, Message: fmt.Sprintf("output is not a valid disputation prose response JSON object: %v", err)}
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return Result{Passed: false, Message: "output must contain exactly one JSON value"}
	} else if err != io.EOF {
		return Result{Passed: false, Message: "output must contain only the disputation prose response JSON object"}
	}

	engagement := strings.TrimSpace(response.Engagement)
	if _, ok := proseResponseEngagements[engagement]; !ok {
		return Result{Passed: false, Message: fmt.Sprintf("engagement %q is not a recognized category", response.Engagement)}
	}
	if strings.TrimSpace(response.LocalizationStatement) == "" {
		return Result{Passed: false, Message: "localization_statement is required"}
	}

	challengePresent := len(response.Challenge) > 0 && strings.TrimSpace(string(response.Challenge)) != "null"

	if engagement != proseEngagementAttacksPremise {
		if challengePresent {
			return Result{Passed: false, Message: fmt.Sprintf("challenge must be null when engagement is %q", engagement)}
		}
		return Result{
			Passed:  true,
			Message: "output matches disputation_prose_response",
			Details: map[string]any{
				"engagement":            engagement,
				"challenge_recommended": false,
			},
			ParsedOutput: json.RawMessage(unwrapped),
		}
	}

	// attacks_premise: the embedded challenge must pass the deterministic
	// disputation_challenge contract, correctly localized to the target premise.
	if !challengePresent {
		return Result{Passed: false, Message: "attacks_premise requires a non-null challenge object"}
	}
	challengeResult := disputationChallengeValidator{options: v.options}.Validate(ctx, string(response.Challenge))
	if !challengeResult.Passed {
		return Result{
			Passed:  false,
			Message: "embedded challenge does not pass disputation_challenge contract: " + challengeResult.Message,
			Details: challengeResult.Details,
		}
	}

	label := strings.TrimSpace(response.TargetArgumentLabel)
	if label != "" {
		if challengeLabel, ok := challengeResult.Details["target_argument_label"].(string); ok {
			if !strings.EqualFold(strings.TrimSpace(challengeLabel), label) {
				return Result{Passed: false, Message: fmt.Sprintf("target_argument_label %q does not match the embedded challenge label %q", label, challengeLabel)}
			}
		}
	}

	details := map[string]any{
		"engagement":            engagement,
		"challenge_recommended": true,
	}
	for key, value := range challengeResult.Details {
		details[key] = value
	}
	return Result{
		Passed:       true,
		Message:      "output matches disputation_prose_response",
		Details:      details,
		ParsedOutput: json.RawMessage(unwrapped),
	}
}
