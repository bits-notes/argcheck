package validation

import (
	"context"
	"strings"
	"testing"
)

const roundFiveArgument = `1. Any standard for load-bearing argumentation that satisfies both condition (a) and condition (b) is the correct standard for load-bearing argumentation.
2. If the articulability thesis is true, then any valid argument can be brought to meet the transparent-validity standard.
3. If it is the case that, if the articulability thesis is true, then any valid argument can be brought to meet the transparent-validity standard, then it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a).
4. Therefore, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a). [2,3]
5. Every inferential link of an argument that meets the transparent-validity standard is an immediate entailment.
6. Every immediate entailment can be verified by direct inspection of the propositions involved.
7. If every inferential link of an argument that meets the transparent-validity standard is an immediate entailment, then, if every immediate entailment can be verified by direct inspection of the propositions involved, then every argument that meets the transparent-validity standard has every one of its inferential links verifiable by direct inspection.
8. Therefore, if every immediate entailment can be verified by direct inspection of the propositions involved, then every argument that meets the transparent-validity standard has every one of its inferential links verifiable by direct inspection. [5,7]
9. Therefore, every argument that meets the transparent-validity standard has every one of its inferential links verifiable by direct inspection. [6,8]
10. If every argument that meets the transparent-validity standard has every one of its inferential links verifiable by direct inspection, then the transparent-validity standard satisfies condition (b).
11. Therefore, the transparent-validity standard satisfies condition (b). [9,10]
12. If the transparent-validity standard satisfies condition (b), then, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a), then, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b).
13. Therefore, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a), then, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b). [11,12]
14. Therefore, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b). [4,13]
15. If any standard for load-bearing argumentation that satisfies both condition (a) and condition (b) is the correct standard for load-bearing argumentation, then, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b), then, if the articulability thesis is true, then the transparent-validity standard is the correct standard for load-bearing argumentation.
16. Therefore, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b), then, if the articulability thesis is true, then the transparent-validity standard is the correct standard for load-bearing argumentation. [1,15]
17. Therefore, if the articulability thesis is true, then the transparent-validity standard is the correct standard for load-bearing argumentation. [14,16]`

func TestRoundFiveArgumentPassesAllStepsAsModusPonens(t *testing.T) {
	report, result := CheckEntailmentGate(roundFiveArgument)
	if !result.Passed || !report.Passed {
		t.Fatalf("round-five gate failed: result=%#v report=%#v", result, report)
	}
	if len(report.Steps) != 8 {
		t.Fatalf("gated steps = %d, want 8", len(report.Steps))
	}
	for _, step := range report.Steps {
		if !step.Passed || step.Form != "modus ponens" {
			t.Fatalf("step %d = %#v, want passing modus ponens", step.Step, step)
		}
	}
}

func TestRoundFiveBracketBridgeVariantPasses(t *testing.T) {
	variant := roundFiveArgument
	replacements := map[string]string{
		"3. If it is the case that, if the articulability thesis is true, then any valid argument can be brought to meet the transparent-validity standard, then it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a).":                                                                                                                                                                                   "3. If [2], then it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a).",
		"7. If every inferential link of an argument that meets the transparent-validity standard is an immediate entailment, then, if every immediate entailment can be verified by direct inspection of the propositions involved, then every argument that meets the transparent-validity standard has every one of its inferential links verifiable by direct inspection.":                                                                                                 "7. If [5], then, if every immediate entailment can be verified by direct inspection of the propositions involved, then every argument that meets the transparent-validity standard has every one of its inferential links verifiable by direct inspection.",
		"12. If the transparent-validity standard satisfies condition (b), then, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a), then, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b).":                                                                                                                             "12. If [11], then, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a), then, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b).",
		"15. If any standard for load-bearing argumentation that satisfies both condition (a) and condition (b) is the correct standard for load-bearing argumentation, then, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b), then, if the articulability thesis is true, then the transparent-validity standard is the correct standard for load-bearing argumentation.": "15. If [1], then, if it is the case that, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b), then, if the articulability thesis is true, then the transparent-validity standard is the correct standard for load-bearing argumentation.",
	}
	for original, replacement := range replacements {
		variant = strings.Replace(variant, original, replacement, 1)
	}

	report, result := CheckEntailmentGate(variant)
	if !result.Passed || !report.Passed {
		t.Fatalf("bracket variant failed: result=%#v report=%#v", result, report)
	}
	if expansions, ok := result.Details["expansions"].([]BridgeExpansion); !ok || len(expansions) != 4 {
		t.Fatalf("expansions = %#v, want four records", result.Details["expansions"])
	}
}

func TestEntailmentGateRealDiagnosticFixtures(t *testing.T) {
	tests := []struct {
		name             string
		argument         string
		passed           bool
		form             string
		mismatchContains []string
	}{
		{
			name: "clean hypothetical syllogism",
			argument: `1. If the articulability thesis is true, then any valid argument can be brought to meet the transparent-validity standard.
2. If any valid argument can be brought to meet the transparent-validity standard, then the transparent-validity standard satisfies condition (a).
3. Therefore, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a). [1,2]`,
			passed: true,
			form:   "hypothetical syllogism",
		},
		{
			name: "round two conjunctive antecedent wobble",
			argument: `1. If the articulability thesis is true, then the transparent-validity standard satisfies condition (a).
2. The transparent-validity standard satisfies condition (b).
3. If it is the case both that, if the articulability thesis is true, then the transparent-validity standard satisfies condition (a), and that the transparent-validity standard satisfies condition (b), then, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b).
4. Therefore, if the articulability thesis is true, then the transparent-validity standard satisfies both condition (a) and condition (b). [1,2,3]`,
			form: "modus ponens",
			mismatchContains: []string{
				"antecedent", "is not matched verbatim by any cited premise",
			},
		},
		{
			name: "predicate smuggling",
			argument: `1. Every inferential link of an argument that meets the transparent-validity standard is an immediate entailment.
2. Therefore, every inferential link of an argument that meets the transparent-validity standard is verifiable by direct inspection. [1]`,
			form: "universal instantiation",
			mismatchContains: []string{
				"is an immediate entailment", "is verifiable by direct inspection",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if report.Passed != test.passed || result.Passed != test.passed {
				t.Fatalf("passed = (%v, %v), want %v; result=%#v report=%#v", report.Passed, result.Passed, test.passed, result, report)
			}
			step := report.Steps[0]
			gotForm := step.Form
			if !step.Passed {
				gotForm = step.NearestForm
			}
			if gotForm != test.form {
				t.Fatalf("form = %q, want %q; step=%#v", gotForm, test.form, step)
			}
			for _, expected := range test.mismatchContains {
				if !strings.Contains(step.Mismatch, expected) {
					t.Fatalf("mismatch %q does not contain %q", step.Mismatch, expected)
				}
			}
		})
	}
}

func TestEntailmentGateBuiltinOptionsAndDetails(t *testing.T) {
	validator, err := Compile(Spec{
		Kind:    KindBuiltin,
		Name:    BuiltinEntailmentGate,
		Options: []byte(`{"allow_single_fenced_block":true,"allow_blank_lines":true}`),
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	output := "```text\n1. Rain falls.\n\n2. Therefore, rain falls. [1]\n```"
	result := validator.Validate(context.Background(), output)
	if !result.Passed {
		t.Fatalf("validate: %#v", result)
	}
	if result.Details["gated_steps"] != 1 || result.Details["passed_steps"] != 1 {
		t.Fatalf("details = %#v", result.Details)
	}
}
