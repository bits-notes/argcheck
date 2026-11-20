package validation

import (
	"fmt"
	"testing"
	"time"
)

// Independent Russell semantics: existence with predication and a universal
// identity condition, rather than production's negated two-witness definition.
func v27Russell(g, p *folFormula) *folFormula {
	x, y := "$count", "$oracleOther"
	unique := countQ("every", y, fol("implies", folSub(g, x, y), foEqual(y, x)))
	return countQ("some", x, foAnd(g, foAnd(p, unique)))
}
func v27Exercises() []nplaceExercise {
	var out []nplaceExercise
	add := func(name string, ps []nOracleSurface, c nOracleSurface, valid bool) {
		var texts []string
		for _, p := range ps {
			texts = append(texts, p.text)
		}
		out = append(out, nplaceExercise{name, v26Argument(texts, c.text), ps, c, valid})
	}
	a := foAtom
	x := "$count"
	author := a("is/author/of", x, "waverley")
	writer := func(t string) *folFormula { return a("wrote", t, "ivanhoe") }
	desc := func(s string, g, p *folFormula) nOracleSurface { return rcSurface(s, v27Russell(g, p)) }
	named := desc("The author of Waverley is Scott", author, foEqual(x, "scott"))
	wrote := desc("The author of Waverley wrote Ivanhoe", author, writer(x))
	scott := rcSurface("Scott wrote Ivanhoe", writer("scott"))
	member := rcSurface("Blair is an author of Waverley", folSub(author, x, "blair"))
	add("b2", []nOracleSurface{named, wrote}, scott, true)
	add("uniqueness", []nOracleSurface{named, member}, rcSurface("Blair is Scott", foEqual("blair", "scott")), true)
	add("membership", []nOracleSurface{named}, rcSurface("Scott is an author of Waverley", folSub(author, x, "scott")), true)
	add("existence and predication", []nOracleSurface{wrote}, rcSurface("Some author of Waverley wrote Ivanhoe", countQ("some", x, foAnd(author, writer(x)))), true)
	add("uniqueness bound", []nOracleSurface{wrote}, rcSurface("At most one author of Waverley", v26OracleDefinition("at most", 1, author)), true)
	add("exact count", []nOracleSurface{wrote}, rcSurface("Exactly one author of Waverley", v26OracleDefinition("exactly", 1, author)), true)
	red := a("is/red", x)
	lab := a("is/lab", x)
	redLab := desc("The lab is red", lab, red)
	exists := rcSurface("Some lab is red", countQ("some", x, foAnd(lab, red)))
	atMost := rcSurface("At most one lab", v26OracleDefinition("at most", 1, lab))
	add("existential introduction", []nOracleSurface{exists, atMost}, redLab, true)
	labAda := rcSurface("Ada is a lab", folSub(lab, x, "ada"))
	redAda := rcSurface("Ada is red", folSub(red, x, "ada"))
	add("ground introduction", []nOracleSurface{labAda, redAda, atMost}, redLab, true)
	add("count introduction", []nOracleSurface{labAda, redAda, rcSurface("Exactly one lab", v26OracleDefinition("exactly", 1, lab))}, redLab, true)
	add("count author introduction", []nOracleSurface{rcSurface("Exactly one author of Waverley", v26OracleDefinition("exactly", 1, author)), rcSurface("Scott is an author of Waverley", folSub(author, x, "scott")), scott}, wrote, true)
	add("eliminate lab", []nOracleSurface{redLab}, exists, true)
	add("eliminate lab bound", []nOracleSurface{redLab}, atMost, true)
	courier := foAnd(a("is/courier", x), a("admires", x, "blair"))
	relName := desc("The courier who admires Blair is Ada", courier, foEqual(x, "ada"))
	relRed := desc("The courier who admires Blair is red", courier, red)
	add("relative substitution", []nOracleSurface{relName, relRed}, redAda, true)
	book := foAnd(a("is/book", x), a("gave/to", "ada", x, "blair"))
	add("relative three-place substitution", []nOracleSurface{desc("The book that Ada gave Blair is Atlas", book, foEqual(x, "atlas")), desc("The book that Ada gave Blair is red", book, red)}, rcSurface("Atlas is red", folSub(red, x, "atlas")), true)
	king := a("is/present king/of", x, "france")
	bald := a("is/bald", x)
	kingPos := desc("The present king of France is bald", king, bald)
	narrow := rcSurface("It is not the case that the present king of France is bald", foNot(v27Russell(king, bald)))
	ambiguous := nOracleSurface{"The present king of France is not bald", []*folFormula{v27Russell(king, foNot(bald)), foNot(v27Russell(king, bald))}}
	add("explicit narrow double negation", []nOracleSurface{rcSurface("It is not the case that it is not the case that the present king of France is bald", foNot(foNot(v27Russell(king, bald))))}, kingPos, true)
	add("narrow reiteration", []nOracleSurface{narrow}, narrow, true)
	add("both readings imply narrow", []nOracleSurface{ambiguous}, narrow, true)
	add("negative substitution all readings", []nOracleSurface{named, nOracleSurface{"The author of Waverley is not red", []*folFormula{v27Russell(author, foNot(red)), foNot(v27Russell(author, red))}}}, rcSurface("Scott is not red", foNot(folSub(red, x, "scott"))), true)
	add("no existence", []nOracleSurface{rcSurface("At most one author of Waverley", v26OracleDefinition("at most", 1, author)), scott}, rcSurface("Scott is an author of Waverley", folSub(author, x, "scott")), false)
	add("missing uniqueness", []nOracleSurface{rcSurface("Some author of Waverley is Scott", countQ("some", x, foAnd(author, foEqual(x, "scott")))), rcSurface("Some author of Waverley wrote Ivanhoe", countQ("some", x, foAnd(author, writer(x))))}, scott, false)
	for _, prefix := range []string{"Ann believes that ", "It is necessary that ", "It is obligatory that "} {
		p := prefix + "the author of Waverley wrote Ivanhoe"
		c := prefix + "Scott wrote Ivanhoe"
		add("opaque "+prefix, []nOracleSurface{named, rcSurface(p, a("opaque/"+normalizeClause(p)))}, rcSurface(c, a("opaque/"+normalizeClause(c))), false)
	}
	add("king existence from one reading", []nOracleSurface{ambiguous}, rcSurface("Some present king of France is not bald", countQ("some", x, foAnd(king, foNot(bald)))), false)
	add("narrow does not imply wide", []nOracleSurface{narrow, rcSurface("At most one present king of France", v26OracleDefinition("at most", 1, king))}, ambiguous, false)
	add("description does not name witness", []nOracleSurface{wrote}, scott, false)
	other := a("is/author/of", x, "ivanhoe")
	add("verbatim restrictor", []nOracleSurface{named, desc("The author of Ivanhoe wrote Ivanhoe", other, writer(x))}, scott, false)
	expanded := descriptionDefinition(lab, red).Render()
	add("definition expansion", []nOracleSurface{redLab}, rcSurface(expanded, v27Russell(lab, red)), true)
	add("definition contraction", []nOracleSurface{rcSurface(expanded, v27Russell(lab, red))}, redLab, true)
	return out
}
func TestCatalogV27Exercises(t *testing.T) {
	// Strict identity requires evidenced constants, bounded descriptions and closed copular predicates; description negation needs explicit scope.
	refused := map[string]strictRefusal{
		"TestCatalogV27Exercises/b2":                         {"identity.extensional_context", "state the substituted claim as a premise"},
		"TestCatalogV27Exercises/membership":                 {"identity.extensional_context", "state the substituted claim as a premise"},
		"TestCatalogV27Exercises/relative_substitution":      {"identity.constant_sides", "name the thing"},
		"TestCatalogV27Exercises/both_readings_imply_narrow": {"scope.description_negation", "keep the wide negation"},
		// identity.constant_sides: v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit); the expansion has explicit bound-variable identity.
		"TestCatalogV27Exercises/definition_expansion": {"identity.constant_sides", "name the thing"},
		// identity.constant_sides: v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit); the contraction has explicit bound-variable identity.
		"TestCatalogV27Exercises/definition_contraction": {"identity.constant_sides", "name the thing"},
	}
	// v46.0: Blair is now evidenced (names allowlist); valid.
	// v46.0: Ada is now evidenced (names allowlist); valid.
	// Uniqueness and relative three-place substitution retain their independent oracle checks.
	start := time.Now()
	valid, invalid := 0, 0
	for _, tc := range v27Exercises() {
		t.Run(tc.name, func(t *testing.T) {
			_, r := CheckEntailmentGate(tc.argument)
			if !listedStrictRefusal(t, tc.argument, refused) && r.Passed != tc.valid {
				t.Errorf("gate: %s\n%s", tc.argument, r.Message)
			}
			if nOracleInvalid(tc.p, tc.c) == tc.valid {
				t.Error("independent domains 1-3 oracle disagrees")
			}
		})
		if tc.valid {
			valid++
		} else {
			invalid++
		}
	}
	t.Logf("%d valid, %d invalid; exhaustive domains 1-3; runtime %s", valid, invalid, time.Since(start))
	if valid < 15 || invalid < 7 {
		t.Fatal("exercise floor")
	}
}
func TestCatalogV27ParserSemantics(t *testing.T) {
	// Check production expansions against independently supplied semantics in
	// both directions, including each negation reading, over domains 1-4.
	start := time.Now()
	seen := map[string]bool{}
	total := 0
	for _, tc := range v27Exercises() {
		for _, s := range append(tc.p, tc.c) {
			if seen[s.text] {
				continue
			}
			seen[s.text] = true
			fs, e := descriptionParse(s.text)
			if e != nil {
				t.Errorf("%s: %v", s.text, e)
				continue
			}
			if len(fs) != len(s.readings) {
				t.Errorf("reading count: %s", s.text)
				continue
			}
			for i, f := range fs {
				if finiteCountermodel([]*folFormula{f}, s.readings[i], 4) || finiteCountermodel([]*folFormula{s.readings[i]}, f, 4) {
					t.Errorf("semantics: %s\n%s", s.text, f.Render())
				}
				total++
			}
		}
	}
	t.Logf("%d reading equivalences; domains 1-4; runtime %s", total, time.Since(start))
}
func TestCatalogV27Fitch(t *testing.T) {
	proof := `1. The author of Waverley is Scott.
| 2. Suppose the author of Waverley wrote Ivanhoe.
| 3. Therefore, Scott wrote Ivanhoe. [1,2]
4. Therefore, if the author of Waverley wrote Ivanhoe then Scott wrote Ivanhoe. [2-3]`
	// identity.extensional_context: "wrote" is outside the closed substitution grammar.
	refused := []string{proof}
	for _, argument := range refused {
		assertStrictRefusal(t, argument, "identity.extensional_context", "state the substituted claim as a premise")
	}
	ex := v27Exercises()[0]
	if finiteCountermodel([]*folFormula{ex.p[0].readings[0]}, fol("implies", ex.p[1].readings[0], ex.c.readings[0]), 4) {
		t.Fatal("Fitch discharge countermodel")
	}
	t.Log(fmt.Sprint("local and conditional discharge checked over domains 1-4"))
}

func TestCatalogV27FallbackBoundary(t *testing.T) {
	// Exercise the new path itself even when an old schema can certify a case.
	for _, tc := range v27Exercises() {
		arg, syntax := ParseNumberedArgument(tc.argument, numberedArgumentOptions{})
		if !syntax.Passed {
			t.Fatal(syntax.Message)
		}
		steps := map[int]NumberedArgumentStep{}
		for _, s := range arg.Steps {
			steps[s.Number] = s
		}
		c := arg.Steps[len(arg.Steps)-1]
		r := gateDescriptions(c, steps)
		if r.Passed != tc.valid {
			t.Errorf("fallback %s: got %v, want %v", tc.name, r.Passed, tc.valid)
		}
	}
	// Equivalent relative clauses still are not verbatim descriptions.
	text := v26Argument([]string{"The courier who admires Blair is Ada", "The courier that admires Blair is red"}, "Ada is red")
	_, r := CheckEntailmentGate(text)
	if r.Passed {
		t.Fatal("nonverbatim description substitution")
	}
	// No unrelated extra citation may be silently discarded.
	text = v26Argument([]string{"The author of Waverley is Scott", "The author of Waverley wrote Ivanhoe", "Ada is red"}, "Scott wrote Ivanhoe")
	_, r = CheckEntailmentGate(text)
	if r.Passed {
		t.Fatal("unrelated citation discarded")
	}
}
