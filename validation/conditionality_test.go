package validation

import (
	"context"
	"strings"
	"testing"
)

func TestConditionalityPreservationValidatorFlagsDroppedAntecedent(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	result := Validate(context.Background(), spec, "1. Microscopic black holes may be produced.")
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "conditionality drift detected") {
		t.Fatalf("message = %q, want drift failure", result.Message)
	}
}

func TestConditionalityPreservationValidatorFlagsNegatedAntecedentMention(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	output := strings.Join([]string{
		"1. It is not the case that TeV-scale gravity assumptions hold.",
		"2. Microscopic black holes may be produced.",
	}, "\n")
	result := Validate(context.Background(), spec, output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "conditionality drift detected") {
		t.Fatalf("message = %q, want drift failure", result.Message)
	}
}

func TestConditionalityPreservationValidatorFlagsMentionOnlyAntecedent(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	output := strings.Join([]string{
		"1. Whether TeV-scale gravity assumptions hold is unresolved.",
		"2. Microscopic black holes may be produced.",
	}, "\n")
	result := Validate(context.Background(), spec, output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "conditionality drift detected") {
		t.Fatalf("message = %q, want drift failure", result.Message)
	}
}

func TestConditionalityPreservationValidatorFlagsNegatedPreservedConditional(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	result := Validate(context.Background(), spec, "1. If it is not the case that TeV-scale gravity assumptions hold, microscopic black holes may be produced.")
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "conditionality drift detected") {
		t.Fatalf("message = %q, want drift failure", result.Message)
	}
}

func TestConditionalityPreservationValidatorFlagsHedgedPreservedConditional(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	result := Validate(context.Background(), spec, "1. If it is possible that TeV-scale gravity assumptions hold, microscopic black holes may be produced.")
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "conditionality drift detected") {
		t.Fatalf("message = %q, want drift failure", result.Message)
	}
}

func TestConditionalityPreservationValidatorFlagsFalseAntecedentSuffix(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	output := strings.Join([]string{
		"1. TeV-scale gravity assumptions hold is false.",
		"2. Microscopic black holes may be produced.",
	}, "\n")
	result := Validate(context.Background(), spec, output)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "conditionality drift detected") {
		t.Fatalf("message = %q, want drift failure", result.Message)
	}
}

func TestConditionalityPreservationValidatorPassesPreservedConditional(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	result := Validate(context.Background(), spec, "1. If TeV-scale gravity assumptions hold, microscopic black holes may be produced.")
	if !result.Passed {
		t.Fatalf("Passed = false, message = %q", result.Message)
	}
}

func TestConditionalityPreservationValidatorPassesEstablishedAntecedent(t *testing.T) {
	spec := BuiltinSpec(BuiltinConditionalityPreserve)
	spec.Options = []byte(`{
		"conditionals": [
			{
				"antecedent": "TeV-scale gravity assumptions hold",
				"consequent": "microscopic black holes may be produced"
			}
		]
	}`)

	output := strings.Join([]string{
		"1. TeV-scale gravity assumptions hold.",
		"2. If TeV-scale gravity assumptions hold, microscopic black holes may be produced.",
		"3. Therefore, microscopic black holes may be produced. [1,2]",
	}, "\n")
	result := Validate(context.Background(), spec, output)
	if !result.Passed {
		t.Fatalf("Passed = false, message = %q", result.Message)
	}
}
