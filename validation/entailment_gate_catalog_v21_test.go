package validation

import (
	"fmt"
	"strings"
	"testing"
)

const v21Trans = `1. Plato has every property that Socrates has.
2. Aristotle has every property that Plato has.
| 3. Let P be an arbitrary property that Socrates has.
| 4. Therefore, Plato has P. [1,3]
| 5. Therefore, Aristotle has P. [2,4]
6. Therefore, Aristotle has every property that Socrates has. [3-5]`

const v21Differ = `1. Socrates has the property of being wise.
2. It is not the case that Plato has the property of being wise.
3. Therefore, Socrates has some property that Plato lacks. [1,2]`

const v21ArbitraryDiffer = `| 1. Let P be an arbitrary property that Socrates has.
| | 2. Suppose It is not the case that Plato has P.
| | 3. Therefore, Socrates has some property that Plato lacks. [1,2]
| | 4. Therefore, It is not the case that Socrates is identical to Plato. [3]
| 5. Therefore, If It is not the case that Plato has P, then It is not the case that Socrates is identical to Plato. [2-4]
| 6. Therefore, Socrates has P. [1]
7. Therefore, Socrates has every property that Socrates has. [1-6]`

var v21AcceptedCases = []v20GateCase{
	{"property inclusion transitivity", v21Trans, "property generalization"},
	{"distinguishing property", v21Differ, "distinguishing property"},
	{"nonidentity", v21Differ + "\n4. Therefore, It is not the case that Socrates is identical to Plato. [3]", "distinguishing property nonidentity"},
	{"non-inclusion", v21Differ + "\n4. Therefore, It is not the case that Plato has every property that Socrates has. [3]", "distinguishing property non-inclusion"},
	{"nested arbitrary distinguishing", v21ArbitraryDiffer, "property generalization"},
	{"reversed citations", strings.Replace(v21Differ, "[1,2]", "[2,1]", 1), "distinguishing property"},
	{"long eigenname", strings.ReplaceAll(v21Trans, " P", " Prop_12"), "property generalization"},
	{"literal that-clause predicate", strings.ReplaceAll(v21Differ, "wise", "such that Lois believes that Socrates flies"), "distinguishing property"},
}

func TestCatalogV21(t *testing.T) {
	if EntailmentGateCatalogVersion != "v27" {
		t.Fatal(EntailmentGateCatalogVersion)
	}
	for _, tc := range v21AcceptedCases {
		t.Run(tc.name, func(t *testing.T) { checkV20Case(t, tc) })
	}
	for _, tc := range []struct {
		arg  string
		step int
		form string
	}{
		{v21Trans, 4, "property instantiation (arbitrary property)"},
		{v21Trans, 5, "property instantiation (arbitrary property)"},
		{v21ArbitraryDiffer, 3, "distinguishing property"},
	} {
		report, _ := CheckEntailmentGate(tc.arg)
		found := false
		for _, s := range report.Steps {
			if s.Step == tc.step {
				found = s.Passed && s.Form == tc.form
			}
		}
		if !found {
			t.Fatalf("step %d: %+v", tc.step, report)
		}
	}
}

func TestCatalogV21Restrictions(t *testing.T) {
	cases := []v20GateCase{
		{"wrong bearer", strings.Replace(v21Differ, "Therefore, Socrates", "Therefore, Aristotle", 1), ""},
		{"reversed direction", strings.Replace(v21Differ, "Socrates has some property that Plato lacks", "Plato has some property that Socrates lacks", 1), ""},
		{"mismatched predicate", strings.Replace(v21Differ, "Plato has the property of being wise", "Plato has the property of being just", 1), ""},
		{"predicate paraphrase", strings.Replace(v21Differ, "Plato has the property of being wise", "Plato has the property of being intelligent", 1), ""},
		{"positive second premise", strings.Replace(v21Differ, "It is not the case that ", "", 1), ""},
		{"alternate lacks wording", strings.Replace(v21Differ, "Plato lacks", "Plato does not have", 1), ""},
		{"nonidentity wrong bearer", "1. Socrates has some property that Plato lacks.\n2. Therefore, It is not the case that Socrates is identical to Aristotle. [1]", ""},
		{"nonidentity positive", "1. Socrates has some property that Plato lacks.\n2. Therefore, Socrates is identical to Plato. [1]", ""},
		{"non-inclusion wrong direction", "1. Socrates has some property that Plato lacks.\n2. Therefore, It is not the case that Socrates has every property that Plato has. [1]", ""},
		{"non-inclusion positive", "1. Socrates has some property that Plato lacks.\n2. Therefore, Plato has every property that Socrates has. [1]", ""},
	}
	for _, term := range []string{"P", "courage", "a dog"} {
		cases = append(cases, v20GateCase{"non-eigenname " + term, fmt.Sprintf("1. Plato has every property that Socrates has.\n2. Socrates has %s.\n3. Therefore, Plato has %s. [1,2]", term, term), ""})
		cases = append(cases, v20GateCase{"untyped distinguishing " + term, strings.ReplaceAll(v21Differ, "the property of being wise", term), ""})
	}
	for _, bearer := range []string{"Every man", "Some man", "No man", "The property of being wise"} {
		for _, name := range []string{"Socrates", "Plato"} {
			cases = append(cases, v20GateCase{"distinguishing bearer " + name + bearer, strings.ReplaceAll(v21Differ, name, bearer), ""})
			for _, conclusion := range []string{"It is not the case that Socrates is identical to Plato", "It is not the case that Plato has every property that Socrates has"} {
				arg := "1. Socrates has some property that Plato lacks.\n2. Therefore, " + conclusion + ". [1]"
				cases = append(cases, v20GateCase{"consequence bearer " + name + bearer + conclusion, strings.ReplaceAll(arg, name, bearer), ""})
			}
		}
	}
	cases = append(cases, v20GateCase{"no modal lifting", "1. Necessarily, Socrates has some property that Plato lacks.\n2. Therefore, Necessarily, It is not the case that Socrates is identical to Plato. [1]", ""})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkV20Case(t, tc) })
	}
}

// Target-step checks isolate name typing from discharge/freshness failures.
// Every fixture still goes through the full parser and scope ownership check.
func TestCatalogV21PropertyScope(t *testing.T) {
	cases := []struct {
		name, argument string
		target         int
		accept         bool
	}{
		{"open", v21Trans, 4, true},
		{"wrong eigenname case", strings.Replace(v21Trans, "Plato has P", "Plato has p", 1), 4, false},
		{"wrong source", strings.Replace(v21Trans, "property that Socrates has.\n| 4", "property that Aristotle has.\n| 4", 1), 4, false},
		{"non-eigenname in scope", strings.Replace(v21Trans, "Plato has P", "Plato has courage", 1), 4, false},
		{"closed scope", `1. Plato has every property that Socrates has.
2. Socrates has P.
| 3. Let P be an arbitrary property that Socrates has.
| 4. Therefore, Socrates has P. [3]
5. Therefore, Socrates has every property that Socrates has. [3-4]
6. Therefore, Plato has P. [1,2]`, 6, false},
		{"sibling scope", `1. Plato has every property that Socrates has.
2. Socrates has P.
| 3. Let P be an arbitrary property that Socrates has.
| 4. Therefore, Socrates has P. [3]
| 5. Suppose Socrates is wise.
| 6. Therefore, Plato has P. [1,2]
7. Therefore, If Socrates is wise, then Plato has P. [5-6]`, 6, false},
		{"individual arbitrary name", strings.Replace(v21Trans, "Let P be an arbitrary property that Socrates has", "Let P be an arbitrary man", 1), 4, false},
		{"quantifier assumption bearer", strings.Replace(v21Trans, "Let P be an arbitrary property that Socrates has", "Let P be an arbitrary property that every man has", 1), 4, false},
		{"nested instantiation", `1. Plato has every property that Socrates has.
| 2. Let P be an arbitrary property that Socrates has.
| | 3. Suppose Socrates is wise.
| | 4. Therefore, Plato has P. [1,2]
| 5. Therefore, If Socrates is wise, then Plato has P. [3-4]
| 6. Therefore, Socrates has P. [2]
7. Therefore, Socrates has every property that Socrates has. [2-6]`, 4, true},
		{"nested distinguishing", v21ArbitraryDiffer, 3, true},
		{"closed distinguishing", `1. Socrates has P.
2. It is not the case that Plato has P.
| 3. Let P be an arbitrary property that Socrates has.
| 4. Therefore, Socrates has P. [3]
5. Therefore, Socrates has every property that Socrates has. [3-4]
6. Therefore, Socrates has some property that Plato lacks. [1,2]`, 6, false},
		{"sibling distinguishing", `1. Socrates has P.
2. It is not the case that Plato has P.
| 3. Let P be an arbitrary property that Socrates has.
| 4. Therefore, Socrates has P. [3]
| 5. Suppose Socrates is wise.
| 6. Therefore, Socrates has some property that Plato lacks. [1,2]
7. Therefore, If Socrates is wise, then Socrates has some property that Plato lacks. [5-6]`, 6, false},
		{"negative eigenname case", strings.Replace(v21ArbitraryDiffer, "Plato has P", "Plato has p", 1), 3, false},
		{"distinct property kinds", strings.ReplaceAll(strings.Replace(v21ArbitraryDiffer, "Plato has P", "Plato has the property of being wise", 1), " P", " wise"), 3, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg, parsed := ParseNumberedArgument(tc.argument, numberedArgumentOptions{})
			if !parsed.Passed {
				t.Fatal(parsed.Message)
			}
			steps := map[int]NumberedArgumentStep{}
			for _, step := range arg.Steps {
				steps[step.Number] = step
			}
			report := gateConclusionStep(steps[tc.target], steps)
			if report.Passed != tc.accept {
				t.Fatalf("want accepted=%v: %+v", tc.accept, report)
			}
		})
	}
}

func TestCatalogV21FormsAppendOnly(t *testing.T) {
	names := EntailmentGateForms()
	if len(names) != 113 { // v26 appends six counting forms.
		t.Fatalf("form count %d", len(names))
	}
	want := "property instantiation (arbitrary property)|distinguishing property|distinguishing property nonidentity|distinguishing property non-inclusion"
	if got := strings.Join(names[81:85], "|"); got != want {
		t.Fatalf("v21 forms: %s", got)
	}
}
