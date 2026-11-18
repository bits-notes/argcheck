package validation

import (
	"fmt"
	"strings"
	"testing"
)

func TestAndOrGroupingSyntax(t *testing.T) {
	for _, tc := range []struct{ name, sentence, reason string }{
		{"symbolic mixture", "A and B or C", "mixes top-level and/or without explicit grouping"},
		{"empty parentheses", "() and Ann sings", "parentheses must contain a clause"},
		{"and then or", "Ann sings and Bob dances or Cal reads", "mixes top-level and/or without explicit grouping"},
		{"or then and", "Ann sings or Bob dances and Cal reads", "mixes top-level and/or without explicit grouping"},
		{"conditional antecedent", "If Ann sings and Bob dances or Cal reads, then Dan sleeps", "mixes top-level and/or without explicit grouping"},
		{"conditional consequent", "If Dan sleeps, then Ann sings and Bob dances or Cal reads", "mixes top-level and/or without explicit grouping"},
		{"quantified scope", "Every x is such that x sings and x dances or x reads", "mixes top-level and/or without explicit grouping"},
		{"inside parentheses", "(Ann sings and Bob dances or Cal reads)", "mixes top-level and/or without explicit grouping"},
		{"both commas", "Ann sings, and Bob dances, or Cal reads", "ambiguous comma grouping: multiple commas mix list punctuation with connective boundaries"},
		{"list comma", "Ann sings, Bob dances and Cal reads, or Dan sleeps", "ambiguous comma grouping: multiple commas mix list punctuation with connective boundaries"},
		{"missing close", "Ann sings and (Bob dances or Cal reads", "unbalanced parentheses: missing closing parenthesis"},
		{"extra close", "Ann sings) and Bob dances", "unbalanced parentheses: unexpected closing parenthesis"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, conclusion := range []bool{false, true} {
				text := "1. " + tc.sentence + ".\n2. Therefore, Ann sings. [1]"
				line := 1
				if tc.name == "or then and" {
					text = "1. " + tc.sentence + ".\n2. Therefore, Cal reads. [1]"
				}
				if conclusion {
					text = "1. Ann sings.\n2. Therefore, " + tc.sentence + ". [1]"
					line = 2
				}
				want := fmt.Sprintf("line %d sentence %q: %s; use either...or / both...and, a comma, or parentheses to group", line, tc.sentence+".", tc.reason)
				_, result := CheckEntailmentGate(text)
				if result.Passed || result.Message != want {
					t.Fatalf("got %v %q; want %q", result.Passed, result.Message, want)
				}
			}
		})
	}
}

func TestAndOrGroupingEntailment(t *testing.T) {
	for _, tc := range []struct {
		name, premise, conclusion string
		valid                     bool
	}{
		{"parenthesized disjunction no elimination", "(Ann sings and Bob dances) or Cal reads", "Ann sings", false},
		{"parenthesized conjunction elimination", "Ann sings and (Bob dances or Cal reads)", "Ann sings", true},
		{"comma disjunction no elimination", "Ann sings and Bob dances, or Cal reads", "Ann sings", false},
		{"comma conjunction elimination", "Ann sings, and Bob dances or Cal reads", "Ann sings", true},
		{"either disjunction", "Either Ann sings and Bob dances, or Cal reads", "Ann sings", false},
		{"both comma disjunction", "Both Ann sings and Bob dances, or Cal reads", "Ann sings", false},
		{"either comma right conjunct", "Either Ann sings, or Bob dances and Cal reads", "Cal reads", false},
		{"either scopes over the rest", "Either Ann sings or Bob dances and Cal reads", "Cal reads", false},
		{"both scopes over the rest", "Both Ann sings and Bob dances or Cal reads", "Ann sings", true},
		{"both comma left operand", "Both Ann sings and Bob dances, or Cal reads", "Both Ann sings and Bob dances, or Cal reads", true},
		{"either without comma", "Either Ann sings and Bob dances or Cal reads", "Ann sings", false},
		{"grouped right conjunct", "Ann sings and (Bob dances or Cal reads)", "Bob dances or Cal reads", true},
		{"nested parentheses", "((Ann sings and Bob dances)) and Cal reads", "Ann sings and Bob dances", true},
		{"single conjunction elimination", "Ann sings and Bob dances", "Ann sings", true},
		{"single disjunction no elimination", "Ann sings or Bob dances", "Ann sings", false},
		{"conjunction list elimination", "Ann sings, Bob dances, and Cal reads", "Cal reads", true},
		{"comma parentheses equivalence", "Ann sings and Bob dances, or Cal reads", "(Ann sings and Bob dances) or Cal reads", true},
		{"conjunction list", "Ann sings, Bob dances, and Cal reads", "Ann sings, Bob dances, and Cal reads", true},
		{"disjunction list", "Ann sings, Bob dances, or Cal reads", "Ann sings, Bob dances, or Cal reads", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, result := CheckEntailmentGate("1. " + tc.premise + ".\n2. Therefore, " + tc.conclusion + ". [1]")
			if result.Passed != tc.valid {
				t.Fatal(result.Message)
			}
		})
	}
}

func TestAndOrGroupingDisjunctionElimination(t *testing.T) {
	for _, premise := range []string{"(Ann sings and Bob dances) or Cal reads", "Ann sings and Bob dances, or Cal reads", "Either Ann sings and Bob dances, or Cal reads"} {
		text := "1. " + premise + ".\n2. It is not the case that Ann sings and Bob dances.\n3. Therefore, Cal reads. [1,2]"
		_, result := CheckEntailmentGate(text)
		if !result.Passed {
			t.Fatalf("%s: %s", premise, result.Message)
		}
	}
}

func TestAndOrGroupingOpaqueScopes(t *testing.T) {
	for _, sentence := range []string{
		"Ann believes that (Bob dances or (Cal reads and Dan sleeps))",
		"Ann might sing and dance or read",
		"Ann sings unless (Bob dances or (Cal reads and Dan sleeps))",
		"Unless Ann sings or Bob dances, Cal reads and Dan sleeps",
		"Everyone who sings or dances and reads is happy",
		"Ann and Bob are friends or colleagues",
		"Ann wonders whether or not Bob dances and Cal reads",
		"Ann believes that (Bob dances or Cal reads)",
		"Ann and Bob are friends",
		"Ann (sings) and Bob dances",
		"(Ann and Bob) are friends",
		"The result is significant (p < 0.01) and reliable",
	} {
		_, result := ParseNumberedArgument("1. "+sentence+".\n2. Therefore, "+sentence+". [1]", numberedArgumentOptions{})
		if !result.Passed {
			t.Fatalf("%s: %s", sentence, result.Message)
		}
	}
}

func TestAndOrGroupingEchoAndReferences(t *testing.T) {
	text := "1. Cal reads.\n2. Therefore, (Ann sings and Bob dances) or Dan sleeps. [1]"
	report, result := CheckEntailmentGate(text)
	if result.Passed || !strings.Contains(strings.ToLower(fmt.Sprintf("%+v", report)), "(ann sings and bob dances)") {
		t.Fatalf("lost grouping: %+v %s", report, result.Message)
	}
	_, result = CheckEntailmentGate("1. Ann sings.\n2. Bob dances.\n3. If [1] and [2], then Cal reads.\n4. Therefore, Cal reads. [1,2,3]")
	if !result.Passed {
		t.Fatal(result.Message)
	}
}

func TestAndOrGroupingConditionalAndQuantified(t *testing.T) {
	for _, text := range []string{
		"1. If (Ann sings and Bob dances) or Cal reads, then Dan sleeps.\n2. (Ann sings and Bob dances) or Cal reads.\n3. Therefore, Dan sleeps. [1,2]",
		"1. Every x is such that (x is red and (x is blue or x is green)).\n2. Therefore, Ann is red and (Ann is blue or Ann is green). [1]",
	} {
		_, result := CheckEntailmentGate(text)
		if !result.Passed {
			t.Fatal(result.Message)
		}
	}
}

func TestAndOrGroupingDilemmaBoundaries(t *testing.T) {
	s := "Either (Ann sings or Bob dances) and Cal reads or Dan sleeps"
	parts := eitherDisjunctPartitions(s, 2)
	if len(parts) != 1 || len(parts[0]) != 2 || parts[0][0] != "(ann sings or bob dances) and cal reads" || parts[0][1] != "dan sleeps" {
		t.Fatalf("wrong visible disjuncts: %v", parts)
	}
	if parts := eitherDisjunctPartitions(s, 3); len(parts) != 0 {
		t.Fatalf("split embedded or: %v", parts)
	}
}

func TestAndOrGroupingProofByCases(t *testing.T) {
	for _, premise := range []string{"(Ann sings and Bob dances) or Cal reads", "Ann sings and Bob dances, or Cal reads", "Either Ann sings and Bob dances, or Cal reads"} {
		text := "1. " + premise + ".\n2. If Ann sings and Bob dances, then Dan sleeps.\n3. If Cal reads, then Dan sleeps.\n4. Therefore, Dan sleeps. [1,2,3]"
		_, result := CheckEntailmentGate(text)
		if !result.Passed {
			t.Fatalf("%s: %s", premise, result.Message)
		}
	}
}

func TestAndOrGroupingSingleConnectiveLegacyParser(t *testing.T) {
	for _, sentence := range []string{"Both Ann sings and Bob dances", "Ann sings, Bob dances, and Cal reads", "Ann sings, Bob dances, or Cal reads"} {
		if _, _, _, managed, _ := groupingPlan(sentence); managed {
			t.Fatalf("new parser intercepted single-connective text: %s", sentence)
		}
	}
}

func TestAndOrGroupingPhraseLevelCoordination(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		valid      bool
	}{
		{"noun phrase or after determiner", "1. The theory leaves no space for degrees of evidence or a special probability logic and Ann sings.\n2. Therefore, Ann sings. [1]", true},
		{"adjective or", "1. If X results in U and U is a bad or catastrophic outcome, then X is risky.\n2. Sam results in Ron and Ron is a bad or catastrophic outcome.\n3. Therefore, Sam is risky. [1,2]", true},
		{"verb phrase or", "1. Ann is approaching or exceeding the record and Bob dances.\n2. Therefore, Bob dances. [1]", true},
		{"determiner opens a clause", "1. Ann sings and the dog barks or Cal reads.\n2. Therefore, Ann sings. [1]", false},
		{"associative comma conjunction", "1. Reform requires guidance for belief or action.\n2. Ann sings, and the bridge principles can recommend beliefs or actions.\n3. If reform requires guidance for belief or action and Ann sings, and the bridge principles can recommend beliefs or actions, then Bob dances.\n4. Therefore, Bob dances. [1,2,3]", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, result := CheckEntailmentGate(tc.text)
			if result.Passed != tc.valid {
				t.Fatalf("got %v: %s", result.Passed, result.Message)
			}
		})
	}
}
