package validation

import (
	"strings"
	"testing"
)

const fitchCP = `1. If it rains, then the street is wet.
2. If the street is wet, then the cars slow.
| 3. Suppose it rains.
| 4. Therefore, the street is wet. [1,3]
| 5. Therefore, the cars slow. [2,4]
6. Therefore, if it rains, then the cars slow. [3-5]`
const fitchRAA = `1. It is not the case that it rains.
| 2. Suppose it rains.
| 3. Therefore, contradiction. [1,2]
4. Therefore, it is not the case that it rains. [2-3]`
const fitchClassical = `1. it rains.
| 2. Suppose it is not the case that it rains.
| 3. Therefore, contradiction. [2,1]
4. Therefore, it rains. [2-3]`
const fitchCases = `1. Either it rains or it snows.
2. If it rains, then the street is wet.
3. If it snows, then the street is wet.
| 4. Suppose it rains.
| 5. Therefore, the street is wet. [4,2]
| 6. Suppose it snows.
| 7. Therefore, the street is wet. [3,6]
8. Therefore, the street is wet. [6-7,1,4-5]`
const fitchUG = `1. Every man is mortal.
| 2. Let a be an arbitrary man.
| 3. Therefore, a is mortal. [1,2]
4. Therefore, Every man is mortal. [2-3]`
const fitchEE = `1. Some man is mortal.
| 2. Suppose a is a man and a is mortal.
| 3. Therefore, a is mortal. [2]
| 4. Therefore, something is mortal. [3]
5. Therefore, something is mortal. [1,2-4]`
const fitchExcludedMiddle = `| 1. Suppose it is not the case that either it rains or it is not the case that it rains.
| | 2. Suppose it rains.
| | 3. Therefore, either it rains or it is not the case that it rains. [2]
| | 4. Therefore, contradiction. [1,3]
| 5. Therefore, it is not the case that it rains. [2-4]
| 6. Therefore, either it rains or it is not the case that it rains. [5]
| 7. Therefore, contradiction. [1,6]
8. Therefore, either it rains or it is not the case that it rains. [1-7]`

func TestEntailmentGateCatalogV14(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestEntailmentGateCatalogV14/EE": {"eg.name_witness", "state the existential claim as its own premise"},
	}

	for _, tc := range []struct{ name, text, form string }{
		{"CP", fitchCP, "conditional proof"}, {"RAA", fitchRAA, "reductio"}, {"classical RAA", fitchClassical, "reductio"},
		{"cases", fitchCases, "proof by cases"}, {"UG", fitchUG, "universal generalization"}, {"EE", fitchEE, "existential elimination"}, {"excluded middle", fitchExcludedMiddle, "reductio"},
		{"identity CP", "| 1. Suppose it rains.\n2. Therefore, if it rains, then it rains. [1-1]", "conditional proof"},
		{"arbitrary reflexivity", "| 1. Let z be an arbitrary man.\n| 2. Therefore, z is a man. [1]\n3. Therefore, Every man is a man. [1-2]", "universal generalization"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg, syntax := ParseNumberedArgument(tc.text, numberedArgumentOptions{})
			if !syntax.Passed {
				t.Fatal(syntax.Message)
			}
			if unused := NumberedArgumentUnusedSteps(arg); len(unused) > 0 {
				t.Fatalf("orphans: %v", unused)
			}
			if listedStrictRefusal(t, tc.text, refused) {
				return
			}
			report, result := CheckEntailmentGate(tc.text)
			if !result.Passed {
				t.Fatalf("%s: %+v", result.Message, report)
			}
			if got := report.Steps[len(report.Steps)-1].Form; got != tc.form {
				t.Fatalf("%q != %q", got, tc.form)
			}
		})
	}
}

func TestEntailmentGateV14NearMisses(t *testing.T) {
	for _, tc := range []struct{ name, text string }{
		{"contradiction without complement", "1. it rains.\n2. it is not the case that it snows.\n3. Therefore, contradiction. [1,2]"},
		{"CP endpoint", strings.Replace(fitchCP, "6. Therefore, if it rains, then the cars slow.", "6. Therefore, if it rains, then it snows.", 1)},
		{"CP antecedent", strings.Replace(fitchCP, "6. Therefore, if it rains,", "6. Therefore, if it snows,", 1)},
		{"RAA endpoint", strings.Replace(fitchRAA, "contradiction. [1,2]", "it rains. [2]", 1)},
		{"cases endpoint", strings.Replace(fitchCases, "| 7. Therefore, the street is wet. [3,6]", "| 7. Therefore, it snows. [6]", 1)},
		{"cases missing branch", strings.Replace(fitchCases, "[6-7,1,4-5]", "[1,4-5]", 1)},
		{"UG stale name", strings.Replace(fitchUG, "Every man is mortal.", "a is mortal.", 1)},
		{"UG verbatim name", strings.Replace(fitchUG, "| 3. Therefore, a is mortal.", "| 3. Therefore, A is mortal.", 1)},
		{"EE escaping name", strings.Replace(fitchEE, "5. Therefore, something is mortal.", "5. Therefore, a is mortal.", 1)},
		{"EE stale name", strings.ReplaceAll(fitchEE, "a is", "man is")},
		{"incomplete range", strings.Replace(fitchCP, "[3-5]", "[3-4]", 1)},
		{"closed citation", strings.Replace(fitchCP, "[3-5]", "[5]", 1)},
		{"inner invalid", strings.Replace(fitchCP, "| 4. Therefore, the street is wet. [1,3]", "| 4. Therefore, the street is wet. [3]", 1)},
		{"range cannot be ignored", fitchCP + "\n7. Therefore, if it rains, then the cars slow. [6,3-5]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if report, result := CheckEntailmentGate(tc.text); report.Passed || result.Passed {
				t.Fatalf("accepted: %s", tc.text)
			}
		})
	}
}

func TestNumberedArgumentFitchScope(t *testing.T) {
	for _, tc := range []struct{ name, text, want string }{
		{"jump", "| | 1. Suppose it rains.\n2. Therefore, it rains. [1]", "step 1 increases depth"},
		{"no assumption", "1. it rains.\n| 2. Therefore, it rains. [1]\n3. Therefore, it rains. [1]", "step 2 increases depth without"},
		{"open block", "1. it rains.\n| 2. Suppose it rains.\n| 3. Therefore, it rains. [2]", "final step 3 must be at depth 0"},
		// "late premise" moved: v28 accepts a root premise after an inference (TestCatalogV28PremiseOrder).
		{"sibling leak", strings.Replace(fitchCases, "[3,6]", "[5]", 1), "step 7 cites 5, which is inside the closed subproof 4-5"},
		{"range at wrong depth", strings.Replace(fitchExcludedMiddle, "[1-7]", "[2-4]", 1), "step 8 range 2-4"},
		{"range reversed", strings.Replace(fitchCP, "[3-5]", "[5-3]", 1), "range 5-3 is reversed"},
		{"root assumption", "1. Suppose it rains.\n2. Therefore, it rains. [1]", "step 1 assumption"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, r := ParseNumberedArgument(tc.text, numberedArgumentOptions{})
			if r.Passed || !strings.Contains(r.Message, tc.want) {
				t.Fatalf("%+v want %s", r, tc.want)
			}
		})
	}
}

func TestFitchFlatConsumersReject(t *testing.T) {
	if _, r := ParseFlatNumberedArgumentAllowingNoConclusions(fitchCP); r.Passed || !strings.Contains(r.Message, "Fitch") {
		t.Fatal(r)
	}
	if _, r := CheckBridgeCompression(fitchCP); r.Passed || !strings.Contains(r.Message, "Fitch") {
		t.Fatal(r)
	}
	arg, _ := ParseNumberedArgument(fitchCP, numberedArgumentOptions{})
	arg.Steps[5].Citations = []int{4}
	if r := GateNumberedArgument(arg); r.Passed || !strings.Contains(r.Steps[0].Mismatch, "closed subproof") {
		t.Fatal(r)
	}
}

func TestFitchNaryCasesAndNestedScopes(t *testing.T) {
	text := `1. Either it rains or it snows or the sun shines.
2. If it rains, then the street is wet.
3. If it snows, then the street is wet.
4. If the sun shines, then the street is wet.
| 5. Suppose it rains.
| 6. Therefore, the street is wet. [2,5]
| 7. Suppose it snows.
| 8. Therefore, the street is wet. [3,7]
| 9. Suppose the sun shines.
| 10. Therefore, the street is wet. [4,9]
11. Therefore, the street is wet. [9-10,5-6,1,7-8]`
	if report, r := CheckEntailmentGate(text); !r.Passed || report.Steps[len(report.Steps)-1].Form != "proof by cases" {
		t.Fatal(r)
	}
	// A nested branch may cite outer assumptions and may discharge its sibling
	// predecessor, but a block belonging to another parent cannot be discharged.
	text = `| 1. Suppose it rains.
| | 2. Suppose it snows.
| | 3. Therefore, it rains. [1]
| 4. Therefore, if it snows, then it rains. [2-3]
| 5. Suppose the sun shines.
| 6. Therefore, if it snows, then it rains. [2-3]
7. Therefore, if the sun shines, then if it snows, then it rains. [5-6]`
	if _, r := ParseNumberedArgument(text, numberedArgumentOptions{}); r.Passed || !strings.Contains(r.Message, "step 6 range 2-3") {
		t.Fatal(r)
	}
}

func TestFitchFreshNamesAndScopeControls(t *testing.T) {
	for _, text := range []string{
		// Valid inner derivation and matching endpoint: freshness alone rejects it.
		`1. Every man is mortal.
2. a is mortal.
| 3. Let a be an arbitrary man.
| 4. Therefore, a is mortal. [1,3]
5. Therefore, Every man is mortal. [3-4]`,
		`1. Some man is mortal.
| 2. Suppose a is a man and a is mortal.
| 3. Therefore, a is mortal. [2]
4. Therefore, a is mortal. [1,2-3]`,
		`1. Some man is mortal.
2. a is mortal.
| 3. Suppose a is a man and a is mortal.
| 4. Therefore, a is mortal. [3]
| 5. Therefore, something is mortal. [4]
6. Therefore, something is mortal. [1,3-5]`,
	} {
		report, r := CheckEntailmentGate(text)
		if strings.Contains(text, "Suppose a is a man and a is mortal.") {
			// eg.name_witness: v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit).
			// Still refused; the refusal now precedes the freshness check, so the mismatch is the witness-shape rewrite.
			hint := "existential witness assumption"
			if strings.Contains(text, "something is mortal") {
				hint = "state the existential claim as its own premise"
			}
			assertStrictRefusal(t, text, "eg.name_witness", hint)
			continue
		}
		if r.Passed || !strings.Contains(report.Steps[len(report.Steps)-1].Mismatch, "fresh name") {
			t.Fatalf("%+v\n%s", report, text)
		}
	}
	// Inline bridge references have the same scope restrictions as citations.
	text := `| 1. Suppose it rains.
| 2. Therefore, it rains. [1]
3. Therefore, if it rains, then it rains. [1-2]
| 4. Suppose If [2], then it snows.
| 5. Therefore, If [2], then it snows. [4]
6. Therefore, if it rains, then it rains. [3]`
	if _, r := ParseNumberedArgument(text, numberedArgumentOptions{}); r.Passed || !strings.Contains(r.Message, "closed subproof 1-2") {
		t.Fatal(r)
	}
}
