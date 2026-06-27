package validation

import (
	"context"
	"strings"
	"testing"
)

func TestDisputationChallengeValidatorPassesFalseAsWrittenChallenge(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "false_as_written",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. A conditional claim under TeV-scale gravity assumptions does not by itself establish that the issue is live.\n3. Therefore, premise 2 is false as written. [1,2]"
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if !result.Passed {
		t.Fatalf("Passed = false, message = %q", result.Message)
	}
	if len(result.ParsedOutput) == 0 {
		t.Fatal("ParsedOutput is empty")
	}
}

func TestDisputationChallengeValidatorRejectsNotNecessarilyConclusion(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "false_as_written",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. Therefore, the target premise is not necessarily true. [1]"
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "too weak") {
		t.Fatalf("message = %q, want weak-challenge failure", result.Message)
	}
}

func TestDisputationChallengeValidatorRejectsNegatedFalseAsWrittenConclusion(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "false_as_written",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. Therefore, the target premise is not false as written. [1]"
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "too weak") {
		t.Fatalf("message = %q, want weak-challenge failure", result.Message)
	}
}

func TestDisputationChallengeValidatorRejectsHedgedFalseAsWrittenConclusion(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "false_as_written",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. Therefore, premise 2 may be false as written. [1]"
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "too weak") {
		t.Fatalf("message = %q, want weak-challenge failure", result.Message)
	}
}

func TestDisputationChallengeValidatorRejectsQuestionAboutFalseAsWritten(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "false_as_written",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. Therefore, whether the target premise is false as written is unresolved. [1]"
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "assert exactly") {
		t.Fatalf("message = %q, want exact-assertion failure", result.Message)
	}
}

func TestDisputationChallengeValidatorRejectsConditionalFalseAsWritten(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "false_as_written",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. Therefore, the target premise is false as written if TeV-scale gravity assumptions fail. [1]"
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "assert exactly") {
		t.Fatalf("message = %q, want exact-assertion failure", result.Message)
	}
}

func TestDisputationChallengeValidatorRejectsWrongChallengeForce(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "unsupported",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. Therefore, the target premise is false as written. [1]"
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, `challenge_force must be "false_as_written"`) {
		t.Fatalf("message = %q, want challenge_force failure", result.Message)
	}
}

func TestDisputationChallengeValidatorRejectsMalformedArgument(t *testing.T) {
	output := `{
		"target_side": "negative",
		"target_argument_label": "N1",
		"target_premise_number": 4,
		"target_premise_text": "The safety analysis rules out every dangerous scenario.",
		"challenge_force": "false_as_written",
		"argument": "1. The safety analysis does not cover the accelerator-specific scenario.\n2. Therefore, premise 4 is false as written."
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "challenge argument is invalid") {
		t.Fatalf("message = %q, want malformed-argument failure", result.Message)
	}
}

func TestDisputationChallengeValidatorRejectsTrailingJunk(t *testing.T) {
	output := `{
		"target_side": "affirmative",
		"target_argument_label": "A1",
		"target_premise_number": 2,
		"target_premise_text": "Published production theory alone made the issue live.",
		"challenge_force": "false_as_written",
		"argument": "1. Published production theory supports only a conditional claim under TeV-scale gravity assumptions.\n2. Therefore, the target premise is false as written. [1]"
	}
	This is extra commentary.`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationChallenge), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "output must contain only") {
		t.Fatalf("message = %q, want trailing-junk failure", result.Message)
	}
}
