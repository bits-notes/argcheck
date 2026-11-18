package validation

import (
	"fmt"
	"strings"
	"testing"
)

func v28Check(t *testing.T, text string, valid bool) {
	t.Helper()
	_, r := CheckEntailmentGate(text)
	if r.Passed != valid {
		t.Fatalf("want valid=%v: %s\n%s", valid, r.Message, text)
	}
}

func TestCatalogV28Identity(t *testing.T) {
	for _, tc := range []struct{ name, identity, premise, conclusion, negative string }{
		{"predicate and", "Ada is identical to Ben", "Ada is necessary and often obligatory", "Ben is necessary and often obligatory", "Ben is necessary and never obligatory"},
		// v35.0: substitution is certified in clause subjects, so the gerund
		// and coordinated-name terms are tested as subjects.
		{"gerund subject", "Standing as an equal is identical to being recognized as a peer", "Standing as an equal is valuable", "Being recognized as a peer is valuable", "Being recognized as a peer is not valuable"},
		{"coordinated name", "Smith and Wesson is identical to the gunmaker", "Smith and Wesson is a company", "The gunmaker is a company", "The gunmaker is a person"},
		{"negative predicate", "Hesperus is identical to Phosphorus", "Hesperus is not a star", "Phosphorus is not a star", "Phosphorus is a star"},
		{"identical with", "Ada is identical with Ben", "Ada is identical with Cy", "Ben is identical with Cy", "Ben is identical with Dee"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v28Check(t, v26Argument([]string{tc.identity, tc.premise}, tc.conclusion), true)
			v28Check(t, v26Argument([]string{tc.identity, tc.premise}, tc.negative), false)
		})
	}
	for _, tc := range []struct{ name, p, c string }{
		{"partial replacement", "Ada admires Ada's teacher", "Ben admires Ada's teacher"},
		{"substring", "Adaline sings", "Benline sings"},
		{"attitude", "Lois believes Ada sings", "Lois believes Ben sings"},
		{"modal", "It is necessary that Ada sings", "It is necessary that Ben sings"},
	} {
		t.Run(tc.name, func(t *testing.T) { v28Check(t, v26Argument([]string{"Ada is identical to Ben", tc.p}, tc.c), false) })
	}
}

func v28LongMP(n int, letter string, bad bool) string {
	var ants, ps []string
	for i := 0; i < n; i++ {
		ants = append(ants, fmt.Sprintf("%s satisfies condition%d", letter, i))
		name := "Ada"
		if bad && i == n-1 {
			name = "Ben"
		}
		ps = append(ps, fmt.Sprintf("%s satisfies condition%d", name, i))
	}
	ps = append([]string{"If " + strings.Join(ants, " and ") + ", then " + letter + " celebrates"}, ps...)
	return v26Argument(ps, "Ada celebrates")
}
func TestCatalogV28Schematic(t *testing.T) {
	for n := 6; n <= 10; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			v28Check(t, v28LongMP(n, "X", false), true)
			v28Check(t, v28LongMP(n, "X", true), false)
		})
	}
	for _, letter := range []string{"U", "V", "W"} {
		t.Run(letter, func(t *testing.T) {
			v28Check(t, v28LongMP(3, letter, false), true)
			v28Check(t, v28LongMP(3, letter, true), false)
		})
	}
	ps := []string{"If X is a person and Y is a person and Z is a person and X receives help from Y and Y receives help from Z, then X thanks Z", "Ada is a person", "Ben is a person", "Cy is a person", "Ada receives help from Ben", "Ben receives help from Cy"}
	t.Run("same category and from Y", func(t *testing.T) {
		v28Check(t, v26Argument(ps, "Ada thanks Cy"), true)
		v28Check(t, v26Argument(ps, "Ada thanks Ben"), false)
	})
}

func TestCatalogV28PremiseOrder(t *testing.T) {
	v28Check(t, "1. It rains.\n2. Therefore, it rains. [1]\n3. If it rains, then it pours.\n4. Therefore, it pours. [2,3]", true)
	v28Check(t, "1. It rains.\n2. Therefore, it pours. [1,3]\n3. If it rains, then it pours.\n4. Therefore, it pours. [2]", false)
	v28Check(t, "| 1. Suppose it rains.\n| 2. Therefore, it rains. [1]\n3. Therefore, if it rains, then it rains. [1-2]\n4. It rains.\n5. Therefore, it rains. [3,4]", true)
}

func TestCatalogV28IdentityOracle(t *testing.T) {
	// Manually assigned meanings; the production parser never supplies these ASTs.
	a := foAtom
	equal := func(x, y string) *folFormula { return foEqual(x, y) }
	cases := []struct {
		name, id, p, c, bad string
		pi, pp, pc, pbad    *folFormula
	}{
		{"predicate and", "Ada is identical to Ben", "Ada is necessary and often obligatory", "Ben is necessary and often obligatory", "Ben is necessary and never obligatory", equal("ada", "ben"), foAnd(a("necessary", "ada"), a("often obligatory", "ada")), foAnd(a("necessary", "ben"), a("often obligatory", "ben")), foAnd(a("necessary", "ben"), a("never obligatory", "ben"))},
		{"gerund object", "Standing as an equal is identical to being recognized as a peer", "Every citizen deserves standing as an equal", "Every citizen deserves being recognized as a peer", "No citizen deserves being recognized as a peer", equal("standing", "recognized"), countQ("every", "$x", fol("implies", a("citizen", "$x"), a("deserves", "$x", "standing"))), countQ("every", "$x", fol("implies", a("citizen", "$x"), a("deserves", "$x", "recognized"))), countQ("every", "$x", fol("implies", a("citizen", "$x"), foNot(a("deserves", "$x", "recognized"))))},
		{"coordinated name", "Smith and Wesson is identical to the gunmaker", "Ada admires Smith and Wesson", "Ada admires the gunmaker", "Ben admires the gunmaker", equal("sw", "gunmaker"), a("admires", "ada", "sw"), a("admires", "ada", "gunmaker"), a("admires", "ben", "gunmaker")},
		{"negative predicate", "Hesperus is identical to Phosphorus", "Hesperus is not a star", "Phosphorus is not a star", "Phosphorus is a star", equal("h", "p"), foNot(a("star", "h")), foNot(a("star", "p")), a("star", "p")},
		{"identical with", "Ada is identical with Ben", "Ada is identical with Cy", "Ben is identical with Cy", "Ben is identical with Dee", equal("ada", "ben"), equal("ada", "cy"), equal("ben", "cy"), equal("ben", "dee")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ps := []nOracleSurface{rcSurface(tc.id, tc.pi), rcSurface(tc.p, tc.pp)}
			if nOracleInvalid(ps, rcSurface(tc.c, tc.pc)) {
				t.Fatal("positive has a countermodel")
			}
			if !nOracleInvalid(ps, rcSurface(tc.bad, tc.pbad)) {
				t.Fatal("negative has no countermodel")
			}
			// v35.0: valid, but the replaced term is an object, which the
			// safe-position rule does not certify.
			objectPosition := tc.name == "gerund object" || tc.name == "coordinated name"
			v28Check(t, v26Argument([]string{tc.id, tc.p}, tc.c), !objectPosition)
			v28Check(t, v26Argument([]string{tc.id, tc.p}, tc.bad), false)
		})
	}
}

func TestCatalogV28SchematicOracle(t *testing.T) {
	// Each valuation checks a ground instance of the universal schema. Binding
	// is separately checked against every supplied conjunct by the gate fixtures.
	for n := 6; n <= 10; n++ {
		ant := gateOracleAtom(0)
		for i := 1; i < n; i++ {
			ant = gateOracleBinary("and", ant, gateOracleAtom(i))
		}
		rule := gateOracleBinary("if", ant, gateOracleAtom(n))
		negativeCountermodel := false
		for valuation := 0; valuation < 1<<(n+2); valuation++ {
			all, near := true, true
			for i := 0; i < n; i++ {
				all = all && gateOracleAtom(i).eval(valuation)
				if i < n-1 {
					near = near && gateOracleAtom(i).eval(valuation)
				}
			}
			near = near && gateOracleAtom(n+1).eval(valuation)
			if rule.eval(valuation) && all && !gateOracleAtom(n).eval(valuation) {
				t.Fatal("MP countermodel")
			}
			negativeCountermodel = negativeCountermodel || (rule.eval(valuation) && near && !gateOracleAtom(n).eval(valuation))
		}
		if !negativeCountermodel {
			t.Fatal("expected inconsistent-binding countermodel")
		}
	}
}

func TestCatalogV28Diagnostics(t *testing.T) {
	text := v28LongMP(6, "X", true)
	report, _ := CheckEntailmentGate(text)
	mismatch := report.Steps[len(report.Steps)-1].Mismatch
	for _, want := range []string{"X satisfies condition5", "X=\"ada\"", "no matching"} {
		if !strings.Contains(mismatch, want) {
			t.Errorf("missing %q in %q", want, mismatch)
		}
	}
	report, _ = CheckEntailmentGate(v28LongMP(11, "X", false))
	mismatch = report.Steps[len(report.Steps)-1].Mismatch
	if !strings.Contains(mismatch, "antecedent has 11 conjuncts; the limit is 10") {
		t.Fatal(mismatch)
	}
}

func TestCatalogV28LegacyLettersAndNegation(t *testing.T) {
	for _, letter := range []string{"U", "V", "W"} {
		v28Check(t, v26Argument([]string{"If X smiles, then X admires " + letter, "Ada smiles"}, "Ada admires "+letter), true)
		v28Check(t, v26Argument([]string{"If " + letter + " smiles, then " + letter + " dances", "Ada smiles"}, "Ada dances"), true)
		v28Check(t, v26Argument([]string{"If " + letter + " smiles, then " + letter + " dances", "Ada smiles"}, "Ben dances"), false)
	}
	TestCatalogV27InheritedNegation(t)
	v28Check(t, v26Argument([]string{"It is not the case that the present king of France is bald"}, "Some present king of France is not bald"), false)
}

func TestCatalogV28IdentityGuards(t *testing.T) {
	for _, tc := range []struct{ name, id, p, c string }{
		{"bound definite", "The citizen is identical with Ada", "Every citizen admires the citizen", "Every citizen admires Ada"},
		{"predicate slot", "Red is identical with crimson", "Ada is red", "Ada is crimson"},
		{"embedded attitude", "Ada is identical with Ben", "Lois believes Ada is ready", "Lois believes Ben is ready"},
		{"quoted name", "Ada is identical with Ben", "Ada is called Ada", "Ben is called Ben"},
		{"nonidentity connective", "Ada is identical with Ben or Cy", "Ada is ready", "Ben or Cy is ready"},
	} {
		t.Run(tc.name, func(t *testing.T) { v28Check(t, v26Argument([]string{tc.id, tc.p}, tc.c), false) })
	}
}

func TestCatalogV28RepeatedCategoryOracle(t *testing.T) {
	a := foAtom
	person := func(x string) *folFormula { return a("person", x) }
	help := func(x, y string) *folFormula { return a("help", x, y) }
	thanks := func(x, y string) *folFormula { return a("thanks", x, y) }
	body := fol("implies", countAnd(person("$x"), person("$y"), person("$z"), help("$x", "$y"), help("$y", "$z")), thanks("$x", "$z"))
	rule := countQ("every", "$x", countQ("every", "$y", countQ("every", "$z", body)))
	ps := []*folFormula{rule, person("ada"), person("ben"), person("cy"), help("ada", "ben"), help("ben", "cy")}
	if finiteCountermodel(ps, thanks("ada", "cy"), 3) {
		t.Fatal("uniform three-variable substitution has a countermodel")
	}
	if !finiteCountermodel(ps, thanks("ada", "ben"), 3) {
		t.Fatal("changed object must have a countermodel")
	}
	for _, letter := range []string{"U", "V", "W"} {
		rule := countQ("every", "$x", fol("implies", a("smiles", "$x"), a("dances", "$x")))
		ps := []nOracleSurface{rcSurface("If "+letter+" smiles, then "+letter+" dances", rule), rcSurface("Ada smiles", a("smiles", "ada"))}
		if nOracleInvalid(ps, rcSurface("Ada dances", a("dances", "ada"))) {
			t.Fatal(letter)
		}
		if !nOracleInvalid(ps, rcSurface("Ben dances", a("dances", "ben"))) {
			t.Fatal(letter)
		}
	}
}

func TestCatalogV28LaterPremiseEigenname(t *testing.T) {
	text := "| 1. Suppose a admires a.\n| 2. Therefore, some x is such that x admires x. [1]\n3. Some x is such that x admires a.\n4. Therefore, some x is such that x admires x. [3,1-2]"
	v28Check(t, text, false)
	p := countQ("some", "$x", foAtom("admires", "$x", "a"))
	c := countQ("some", "$x", foAtom("admires", "$x", "$x"))
	if !finiteCountermodel([]*folFormula{p}, c, 2) {
		t.Fatal("expected two-person countermodel without a self-loop")
	}
}

func TestCatalogV28PartitionLimit(t *testing.T) {
	text := strings.ReplaceAll(v28LongMP(10, "X", false), "condition9", "condition9 and alpha and beta and gamma and delta")
	report, result := CheckEntailmentGate(text)
	if result.Passed {
		t.Fatal("expected bounded partition rejection")
	}
	if got := report.Steps[len(report.Steps)-1].Mismatch; !strings.Contains(got, "partition limit is 12") {
		t.Fatal(got)
	}
}

func TestCatalogV28WrappedLengthCompatibility(t *testing.T) {
	text := v28LongMP(11, "X", false)
	text = strings.ReplaceAll(text, " and X", ", and it is the case that X")
	v28Check(t, text, true)
	v28Check(t, strings.Replace(text, "Ada satisfies condition10", "Ben satisfies condition10", 1), false)
}
