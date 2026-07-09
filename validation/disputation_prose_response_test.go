package validation

import (
	"context"
	"strings"
	"testing"
)

// A premise attack whose embedded challenge concludes exactly that the target
// premise is false as written must pass.
func TestDisputationProseResponsePassesPremiseAttack(t *testing.T) {
	output := `{
		"engagement": "attacks_premise",
		"target_argument_label": "A1",
		"localization_statement": "You are attacking premise 19 of A1 ('In the specified use context, moderate egg intake up to about one whole egg per day does not substitute for clearly superior cardiometabolic alternatives.'). Conclusion 21 depends on it, so a challenge lands on premise 19.",
		"justification": "The prose denies exactly what premise 19 asserts.",
		"challenge": {
			"target_side": "affirmative",
			"target_argument_label": "A1",
			"target_premise_number": 19,
			"target_premise_text": "In the specified use context, moderate egg intake up to about one whole egg per day does not substitute for clearly superior cardiometabolic alternatives.",
			"challenge_force": "false_as_written",
			"argument": "1. In ordinary diets an egg eaten at breakfast replaces foods such as oats, fruit, or legumes.\n2. For most adults those replaced foods are clearly superior cardiometabolic alternatives.\n3. Therefore, premise 19 is false as written. [1,2]"
		}
	}`

	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationProseResponse), output)
	if !result.Passed {
		t.Fatalf("Passed = false, message = %q", result.Message)
	}
	if rec, _ := result.Details["challenge_recommended"].(bool); !rec {
		t.Fatalf("challenge_recommended = %v, want true", result.Details["challenge_recommended"])
	}
}

// Non-premise engagement categories must carry a null challenge and still pass.
func TestDisputationProseResponsePassesNonAttackWithNullChallenge(t *testing.T) {
	for _, engagement := range []string{
		"attacks_inference_step", "contests_rendering", "introduces_new_evidence", "does_not_engage",
	} {
		output := `{
			"engagement": "` + engagement + `",
			"target_argument_label": "",
			"localization_statement": "The prose does not assert any single premise is false as written; no challenge is produced.",
			"justification": "Classified by strongest substantive engagement.",
			"challenge": null
		}`
		result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationProseResponse), output)
		if !result.Passed {
			t.Fatalf("engagement %q: Passed = false, message = %q", engagement, result.Message)
		}
		if rec, _ := result.Details["challenge_recommended"].(bool); rec {
			t.Fatalf("engagement %q: challenge_recommended = true, want false", engagement)
		}
	}
}

// A premise attack with a weakened embedded conclusion must be rejected by the
// delegated disputation_challenge contract.
func TestDisputationProseResponseRejectsWeakEmbeddedChallenge(t *testing.T) {
	output := `{
		"engagement": "attacks_premise",
		"target_argument_label": "A1",
		"localization_statement": "You are attacking premise 19 of A1.",
		"justification": "x",
		"challenge": {
			"target_side": "affirmative",
			"target_argument_label": "A1",
			"target_premise_number": 19,
			"target_premise_text": "In the specified use context, moderate egg intake up to about one whole egg per day does not substitute for clearly superior cardiometabolic alternatives.",
			"challenge_force": "false_as_written",
			"argument": "1. Eggs often replace superior foods.\n2. Therefore, premise 19 may be false as written. [1]"
		}
	}`
	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationProseResponse), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "disputation_challenge contract") {
		t.Fatalf("message = %q, want delegated-contract failure", result.Message)
	}
}

// A premise attack with a null challenge must be rejected.
func TestDisputationProseResponseRejectsPremiseAttackWithoutChallenge(t *testing.T) {
	output := `{
		"engagement": "attacks_premise",
		"target_argument_label": "A1",
		"localization_statement": "You are attacking premise 19 of A1.",
		"justification": "x",
		"challenge": null
	}`
	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationProseResponse), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "requires a non-null challenge") {
		t.Fatalf("message = %q, want missing-challenge failure", result.Message)
	}
}

// A non-attack category that still carries a challenge must be rejected.
func TestDisputationProseResponseRejectsStrayChallenge(t *testing.T) {
	output := `{
		"engagement": "does_not_engage",
		"target_argument_label": "",
		"localization_statement": "Off-topic venting.",
		"justification": "x",
		"challenge": {
			"target_side": "affirmative",
			"target_argument_label": "A1",
			"target_premise_number": 19,
			"target_premise_text": "x",
			"challenge_force": "false_as_written",
			"argument": "1. p.\n2. Therefore, premise 19 is false as written. [1]"
		}
	}`
	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationProseResponse), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "challenge must be null") {
		t.Fatalf("message = %q, want stray-challenge failure", result.Message)
	}
}

// A mismatched wrapper label vs. embedded challenge label must be rejected,
// protecting localization trust.
func TestDisputationProseResponseRejectsLabelMismatch(t *testing.T) {
	output := `{
		"engagement": "attacks_premise",
		"target_argument_label": "A2",
		"localization_statement": "You are attacking premise 19 of A1.",
		"justification": "x",
		"challenge": {
			"target_side": "affirmative",
			"target_argument_label": "A1",
			"target_premise_number": 19,
			"target_premise_text": "In the specified use context, moderate egg intake up to about one whole egg per day does not substitute for clearly superior cardiometabolic alternatives.",
			"challenge_force": "false_as_written",
			"argument": "1. Eggs often replace superior foods.\n2. Therefore, premise 19 is false as written. [1]"
		}
	}`
	result := Validate(context.Background(), BuiltinSpec(BuiltinDisputationProseResponse), output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "does not match") {
		t.Fatalf("message = %q, want label-mismatch failure", result.Message)
	}
}
