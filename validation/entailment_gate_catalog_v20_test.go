package validation

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

type v20GateCase struct{ name, argument, wantForm string }

const v20Descartes = `1. If Descartes is identical to his body, then Descartes has every property that his body has.
2. His body has the property of being doubtable by Descartes.
3. It is not the case that Descartes has the property of being doubtable by Descartes.
| 4. Suppose Descartes has every property that his body has.
| 5. Therefore, Descartes has the property of being doubtable by Descartes. [2,4]
| 6. Therefore, contradiction. [3,5]
7. Therefore, it is not the case that Descartes has every property that his body has. [4-6]
8. Therefore, it is not the case that Descartes is identical to his body. [1,7]`

const v20IdentityOfIndiscernibles = `1. Socrates has every property that Plato has.
2. Therefore, Plato is identical to Plato. [1]
3. Therefore, Plato has the property of being identical to Plato. [2]
4. Therefore, Socrates has the property of being identical to Plato. [1,3]
5. Therefore, Socrates is identical to Plato. [4]`

const v20UG = `1. Socrates is identical to Plato.
| 2. Let P be an arbitrary property that Plato has.
| 3. Therefore, Socrates has P. [1,2]
4. Therefore, Socrates has every property that Plato has. [2-3]`

var v20AcceptedCases = []v20GateCase{
	{"abstraction", "1. Socrates is wise.\n2. Therefore, Socrates has the property of being wise. [1]", "property abstraction"},
	{"concretion", "1. Socrates has the property of being wise.\n2. Therefore, Socrates is wise. [1]", "property concretion"},
	{"instantiation", "1. Socrates has every property that Plato has.\n2. Plato has the property of being wise.\n3. Therefore, Socrates has the property of being wise. [1,2]", "property instantiation"},
	{"existential", "1. Socrates has the property of being wise.\n2. Therefore, Socrates has some property. [1]", "property existential"},
	{"shared property", "1. Socrates has the property of being wise.\n2. Plato has the property of being wise.\n3. Therefore, Socrates shares some property with Plato. [1,2]", "shared property"},
	{"Leibniz", "1. Socrates is identical to Plato.\n2. Therefore, Socrates has every property that Plato has. [1]", "indiscernibility of identicals"},
	{"property UG", v20UG, "property generalization"},
	{"Descartes", v20Descartes, "modus tollens"},
	{"identity of indiscernibles", v20IdentityOfIndiscernibles, "property concretion"},
	{"negative abstraction", "1. It is not the case that Socrates is wise.\n2. Therefore, it is not the case that Socrates has the property of being wise. [1]", "property abstraction"},
	{"negative concretion", "1. It is not the case that Socrates has the property of being wise.\n2. Therefore, it is not the case that Socrates is wise. [1]", "property concretion"},
	{"negative predicate", "1. Socrates is not wise.\n2. Therefore, Socrates has the property of being not wise. [1]", "property abstraction"},
	{"opaque predicate named whole", "1. Socrates is such that Lois believes that Socrates flies.\n2. Therefore, Socrates has the property of being such that Lois believes that Socrates flies. [1]", "property abstraction"},
	{"opaque predicate concreted whole", "1. Socrates has the property of being such that Lois believes that Socrates flies.\n2. Therefore, Socrates is such that Lois believes that Socrates flies. [1]", "property concretion"},
	{"opaque property instantiation", "1. Socrates has every property that Plato has.\n2. Plato has the property of being such that Lois believes that Plato flies.\n3. Therefore, Socrates has the property of being such that Lois believes that Plato flies. [1,2]", "property instantiation"},
	{"embedded copula subject", "1. The claim that the store is nearby is false.\n2. Therefore, The claim that the store is nearby has the property of being false. [1]", "property abstraction"},
	{"embedded copula concretion", "1. The claim that the store is nearby has the property of being false.\n2. Therefore, The claim that the store is nearby is false. [1]", "property concretion"},
	{"possessive bearer", "1. His body is doubtable by Descartes.\n2. Therefore, His body has the property of being doubtable by Descartes. [1]", "property abstraction"},
	{"am copula", "1. I am wise.\n2. Therefore, I has the property of being wise. [1]", "property abstraction"},
	{"are copula", "1. You are wise.\n2. Therefore, You has the property of being wise. [1]", "property abstraction"},
	{"pronoun preserved", "1. Socrates is identical to himself.\n2. Therefore, Socrates has the property of being identical to himself. [1]", "property abstraction"},
	{"assumption citable", "| 1. Let P be an arbitrary property that Plato has.\n| 2. Therefore, Plato has P. [1]\n3. Therefore, Plato has every property that Plato has. [1-2]", "property generalization"},
	{"UG direct assumption range", "| 1. Let P be an arbitrary property that Plato has.\n2. Therefore, Plato has every property that Plato has. [1-1]", "property generalization"},
	{"nested UG", "| 1. Suppose Socrates is identical to Plato.\n| | 2. Let P be an arbitrary property that Plato has.\n| | 3. Therefore, Socrates has P. [1,2]\n| 4. Therefore, Socrates has every property that Plato has. [2-3]\n5. Therefore, If Socrates is identical to Plato, then Socrates has every property that Plato has. [1-4]", "conditional proof"},
}

var v20KeptGaps = []v20GateCase{
	{"verb predicates", "1. Superman flies.\n2. Therefore, Superman has the property of being flying. [1]", ""},
	{"property of a property", "1. The property of being wise is admirable.\n2. Therefore, The property of being wise has the property of being admirable. [1]", ""},
	{"general induction", "1. Zero has P.\n2. Every number is such that if that number has P, then its successor has P.\n3. Therefore, Every number has P. [1,2]", ""},
	{"alternate property wording", "1. Socrates is wise.\n2. Therefore, Being wise is a property that Socrates has. [1]", ""},
}

func checkV20Case(t *testing.T, tc v20GateCase) {
	t.Helper()
	if _, parsed := ParseNumberedArgument(tc.argument, numberedArgumentOptions{}); !parsed.Passed {
		t.Fatalf("bad fixture: %s", parsed.Message)
	}
	report, result := CheckEntailmentGate(tc.argument)
	if len(report.Steps) == 0 {
		t.Fatalf("no gated steps: %s", result.Message)
	}
	last := report.Steps[len(report.Steps)-1]
	if tc.wantForm == "" {
		if report.Passed || last.Passed {
			t.Fatalf("rejection accepted: %+v", report)
		}
	} else if !report.Passed || last.Form != tc.wantForm {
		t.Fatalf("want %s: %+v (%s)", tc.wantForm, report, result.Message)
	}
}

func TestCatalogV20(t *testing.T) {
	if EntailmentGateCatalogVersion != "v43.1" {
		t.Fatal(EntailmentGateCatalogVersion)
	}
	for _, tc := range append(append([]v20GateCase{}, v20AcceptedCases...), v20KeptGaps...) {
		t.Run(tc.name, func(t *testing.T) { checkV20Case(t, tc) })
	}
}

func TestCatalogV20Restrictions(t *testing.T) {
	cases := []v20GateCase{
		{"predicate paraphrase", "1. Socrates is wise.\n2. Therefore, Socrates has the property of being intelligent. [1]", ""},
		{"concretion paraphrase", "1. Socrates has the property of being wise.\n2. Therefore, Socrates is intelligent. [1]", ""},
		{"predicate articles verbatim", "1. Socrates is a teacher.\n2. Therefore, Socrates has the property of being the teacher. [1]", ""},
		{"predicate plural verbatim", "1. Socrates is a teacher.\n2. Therefore, Socrates has the property of being a teachers. [1]", ""},
		{"concretion wrong bearer", "1. Socrates has the property of being wise.\n2. Therefore, Plato is wise. [1]", ""},
		{"abstraction polarity", "1. It is not the case that Socrates is wise.\n2. Therefore, Socrates has the property of being wise. [1]", ""},
		{"concretion polarity", "1. Socrates has the property of being wise.\n2. Therefore, It is not the case that Socrates is wise. [1]", ""},
		{"double negative rewrite", "1. It is not the case that it is not the case that Socrates is wise.\n2. Therefore, It is not the case that it is not the case that Socrates has the property of being wise. [1]", ""},
		{"instantiation wrong direction", "1. Socrates has every property that Plato has.\n2. Socrates has the property of being wise.\n3. Therefore, Plato has the property of being wise. [1,2]", ""},
		{"instantiation mismatch", "1. Socrates has every property that Plato has.\n2. Plato has the property of being wise.\n3. Therefore, Socrates has the property of being just. [1,2]", ""},
		{"instantiation needs canonical property", "1. Socrates has every property that Plato has.\n2. Plato is wise.\n3. Therefore, Socrates is wise. [1,2]", ""},
		{"existential wrong bearer", "1. Socrates has the property of being wise.\n2. Therefore, Plato has some property. [1]", ""},
		{"existential needs positive possession", "1. It is not the case that Socrates has the property of being wise.\n2. Therefore, Socrates has some property. [1]", ""},
		{"shared different properties", "1. Socrates has the property of being wise.\n2. Plato has the property of being just.\n3. Therefore, Socrates shares some property with Plato. [1,2]", ""},
		{"shared wrong bearer", "1. Socrates has the property of being wise.\n2. Plato has the property of being wise.\n3. Therefore, Socrates shares some property with Aristotle. [1,2]", ""},
		{"Leibniz negative identity", "1. It is not the case that Socrates is identical to Plato.\n2. Therefore, Socrates has every property that Plato has. [1]", ""},
		{"Leibniz different bearer", "1. Socrates is identical to Plato.\n2. Therefore, Socrates has every property that Aristotle has. [1]", ""},
		{"no pronoun interpretation", "1. Socrates is identical to himself.\n2. Therefore, Socrates has the property of being identical to Socrates. [1]", ""},
		{"no that-clause substitution", "1. Socrates has every property that Plato has.\n2. Plato has the property of being such that Lois believes that Plato flies.\n3. Therefore, Socrates has the property of being such that Lois believes that Socrates flies. [1,2]", ""},
		{"partial embedded copula", "1. The claim that the store is nearby is false.\n2. Therefore, The claim that the store has the property of being nearby is false. [1]", ""},
		{"UG fixed property", strings.Replace(v20UG, "Socrates has P.", "Socrates has the property of being wise.", 1), ""},
		{"UG wrong source", strings.Replace(v20UG, "every property that Plato has", "every property that Aristotle has", 1), ""},
		{"UG earlier eigenname", "1. P is wise.\n" + renumberV20UG(), ""},
		{"UG eigenname escapes", "| 1. Let P be an arbitrary property that Plato has.\n| 2. Therefore, Plato has P. [1]\n3. Therefore, P has every property that Plato has. [1-2]", ""},
		{"UG case changed", "| 1. Let P be an arbitrary property that Plato has.\n| 2. Therefore, Plato has p. [1]\n3. Therefore, Plato has every property that Plato has. [1-2]", ""},
		{"UG wrong target", strings.Replace(v20UG, "Socrates has every property", "Aristotle has every property", 1), ""},
	}
	// Each flat form independently enforces the individual type restriction.
	for _, bearer := range []string{"Every man", "Some man", "No man", "The property of being wise"} {
		for _, tc := range v20AcceptedCases[:6] {
			cases = append(cases, v20GateCase{tc.name + " bearer " + bearer, strings.ReplaceAll(tc.argument, "Socrates", bearer), ""})
			if strings.Contains(tc.argument, "Plato") {
				cases = append(cases, v20GateCase{tc.name + " second bearer " + bearer, strings.ReplaceAll(tc.argument, "Plato", bearer), ""})
			}
		}
		cases = append(cases, v20GateCase{"UG source " + bearer, strings.ReplaceAll(v20UG, "Plato", bearer), ""})
	}
	// A property rewrite must never run below these outer operators, in either direction.
	for _, wrap := range []struct{ name, prefix, suffix string }{
		{"necessity", "Necessarily, ", ""}, {"possibility", "Possibly, ", ""},
		{"conditional", "If the lamp is lit, then ", ""},
		{"counterfactual antecedent", "If ", ", then the lamp would be lit"},
		{"belief", "Lois believes that ", ""},
		{"conjunction", "The lamp is lit and ", ""},
		{"disjunction", "Either the lamp is lit or ", ""},
	} {
		a, b := wrap.prefix+"Socrates is wise"+wrap.suffix, wrap.prefix+"Socrates has the property of being wise"+wrap.suffix
		cases = append(cases, v20GateCase{wrap.name + " abstraction", fmt.Sprintf("1. %s.\n2. Therefore, %s. [1]", a, b), ""}, v20GateCase{wrap.name + " concretion", fmt.Sprintf("1. %s.\n2. Therefore, %s. [1]", b, a), ""})
	}
	// All new flat forms, not just the equivalence, stay outside modal lifting.
	for _, tc := range v20AcceptedCases[:6] {
		lines := strings.Split(tc.argument, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Therefore, ") {
				lines[i] = strings.Replace(line, "Therefore, ", "Therefore, Necessarily, ", 1)
			} else {
				number, body, _ := strings.Cut(line, ". ")
				lines[i] = number + ". Necessarily, " + body
			}
		}
		cases = append(cases, v20GateCase{tc.name + " no modal lifting", strings.Join(lines, "\n"), ""})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkV20Case(t, tc) })
	}
}

func TestCatalogV20FormsAppendOnly(t *testing.T) {
	names := EntailmentGateForms()
	const oldCount = 74
	const oldDigest = "e3189c4796b5b03ca3493b3a1a08b9ec3c429801f684a44741d64ad9781e7c61"
	if len(names) < oldCount+7 {
		t.Fatalf("form count %d", len(names))
	}
	// Frozen SHA-256 of the v19 public list, joined by newlines. It includes
	// the Fitch forms, so inserting new names before them fails this check.
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(names[:oldCount], "\n")))); got != oldDigest {
		t.Fatalf("v19 catalog prefix changed: %s", got)
	}
	want := "property abstraction|property concretion|property instantiation|property existential|shared property|indiscernibility of identicals|property generalization"
	if got := strings.Join(names[oldCount:oldCount+7], "|"); got != want {
		t.Fatalf("v20 forms: %s", got)
	}
}

func renumberV20UG() string {
	return `2. Socrates is identical to Plato.
| 3. Let P be an arbitrary property that Plato has.
| 4. Therefore, Socrates has P. [2,3]
5. Therefore, Socrates has every property that Plato has. [3-4]`
}

func TestCatalogV20ScopeFailures(t *testing.T) {
	for _, argument := range []string{
		strings.Replace(v20UG, "[2-3]", "[2,3]", 1),
		strings.Replace(v20UG, "[2-3]", "[3-3]", 1),
		"| 1. Let P be an arbitrary property that Plato has.\n| 2. Therefore, Plato has P. [1]\n| 3. Let Q be an arbitrary property that Socrates has.\n| 4. Therefore, Plato has P. [2]\n5. Therefore, Socrates has every property that Plato has. [3-4]",
	} {
		if report, _ := CheckEntailmentGate(argument); report.Passed {
			t.Fatalf("scope escape: %s", argument)
		}
	}
}

func TestCatalogV20WorkedProofSteps(t *testing.T) {
	for _, tc := range []struct {
		argument string
		forms    []string
	}{
		{v20Descartes, []string{"property instantiation", "contradiction introduction", "reductio", "modus tollens"}},
		{v20IdentityOfIndiscernibles, []string{"identity reflexivity", "property abstraction", "property instantiation", "property concretion"}},
	} {
		report, result := CheckEntailmentGate(tc.argument)
		if !report.Passed || len(report.Steps) != len(tc.forms) {
			t.Fatal(result.Message)
		}
		for i, form := range tc.forms {
			if report.Steps[i].Form != form {
				t.Fatalf("step %d: %+v", i, report.Steps[i])
			}
		}
	}
}
