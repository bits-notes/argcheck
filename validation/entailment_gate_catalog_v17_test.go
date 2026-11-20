package validation

import (
	"fmt"
	"strings"
	"testing"
)

func relationArgument(p []string, c string) string {
	var lines, refs []string
	for i, s := range p {
		lines = append(lines, fmt.Sprintf("%d. %s.", i+1, s))
		refs = append(refs, fmt.Sprint(i+1))
	}
	return strings.Join(append(lines, fmt.Sprintf("%d. Therefore, %s. [%s]", len(p)+1, c, strings.Join(refs, ","))), "\n")
}

const relationFitchUG = `1. Every student greets every teacher.
| 2. Let b be an arbitrary student.
| 3. Therefore, b greets every teacher. [1,2]
4. Therefore, Every student greets every teacher. [2-3]`
const relationFitchEE = `1. Some student greets Bob.
| 2. Suppose c is a student and c greets Bob.
| 3. Therefore, c is a student. [2]
| 4. Therefore, c greets Bob. [2]
| 5. Therefore, Some student greets Bob. [4,3]
6. Therefore, Some student greets Bob. [1,2-5]`

func TestEntailmentGateCatalogV17(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// eg.subject_only: dedicated object witnesses remain conservatively refused.
		"TestEntailmentGateCatalogV17/object_EG": {"eg.subject_only", "state the existential claim as its own premise"},
		// v46.0: Alice is now evidenced (names allowlist); valid.
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV17/caused": {"eg.subject_only", "state the existential claim as its own premise"},
	}

	for _, tc := range []struct {
		name string
		p    []string
		c    string
	}{
		{"subject UI", []string{"Every student greets every teacher", "Alice is a student"}, "Alice greets every teacher"},
		{"object UI", []string{"Alice greets every teacher", "Bob is a teacher"}, "Alice greets Bob"},
		{"each UI", []string{"Alice greets each teacher", "Bob is a teacher"}, "Alice greets Bob"},
		{"both scope readings UI verbatim", []string{"Every student greets some teacher", "Alice is a student"}, "Alice greets some teacher"},
		{"object EG", []string{"Alice greets Bob", "Bob is a teacher"}, "Alice greets some teacher"},
		{"subject EG", []string{"Alice greets Bob", "Alice is a student"}, "Some student greets Bob"},
		{"object negative", []string{"Alice greets no tyrant", "Bob is a tyrant"}, "It is not the case that Alice greets Bob"},
		{"subject negative", []string{"No student greets Bob", "Alice is a student"}, "It is not the case that Alice greets Bob"},
		{"not some", []string{"It is not the case that some student greets Bob"}, "No student greets Bob"},
		{"no to not some", []string{"No student greets Bob"}, "It is not the case that some student greets Bob"},
		{"not every", []string{"It is not the case that every student greets Bob"}, "Some student is such that it is not the case that that student greets Bob"},
		{"not every object", []string{"It is not the case that Alice greets every teacher"}, "Some teacher is such that it is not the case that Alice greets that teacher"},
		{"negative witness reverse", []string{"Some teacher is such that it is not the case that Alice greets that teacher"}, "It is not the case that Alice greets every teacher"},
		{"scaffold intro", []string{"Every student greets Bob"}, "Every student is such that that student greets Bob"},
		{"scaffold elimination", []string{"Every student is such that that student greets Bob"}, "Every student greets Bob"},
		{"wide witness weakening", []string{"Some teacher is such that every student greets that teacher"}, "Every student is such that that student greets some teacher"},
		{"wide witness to ambiguous", []string{"Some teacher is such that every student greets that teacher"}, "Every student greets some teacher"},
		{"wide witness introduction", []string{"Every student greets Bob", "Bob is a teacher"}, "Some teacher is such that every student greets that teacher"},
		{"nested all commute", []string{"Every student greets every teacher"}, "Every teacher is such that every student greets that teacher"},
		{"exact copular relation", []string{"Alice is preferable to every teacher", "Bob is a teacher"}, "Alice is preferable to Bob"},
		{"caused", []string{"Alice caused Bob", "Bob is a teacher"}, "Alice caused some teacher"},
		{"outweighs", []string{"Alice outweighs every teacher", "Bob is a teacher"}, "Alice outweighs Bob"},
		{"passive stays passive", []string{"Alice is greeted by every teacher", "Bob is a teacher"}, "Alice is greeted by Bob"},
		{"multiword class", []string{"Every graduate student greets Bob", "Alice is a graduate student"}, "Alice greets Bob"},
		{"multiword mixed scope class", []string{"Every graduate student greets some senior teacher", "Alice is a graduate student"}, "Alice greets some senior teacher"},
		{"prepositional relation", []string{"Every student works with some teacher", "Alice is a student"}, "Alice works with some teacher"},
		{"negative shared inhabited class", []string{"No student greets every student", "Alice is a student"}, "It is not the case that Alice greets every student"},
		{"scope uses its enclosing inhabitant", []string{"Every student greets no student"}, "No student is such that every student greets that student"},
		{"multiword name", []string{"Mary Jane greets every teacher", "Bob is a teacher"}, "Mary Jane greets Bob"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg := relationArgument(tc.p, tc.c)
			if listedStrictRefusal(t, arg, refused) {
				return
			}
			report, result := CheckEntailmentGate(arg)
			if !result.Passed {
				t.Fatalf("%s\n%s\n%+v", arg, result.Message, report)
			}
		})
	}
	// scope.clause_start: v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit).
	assertStrictRefusal(t, relationFitchEE, "scope.clause_start", "move the operator into each clause")
	for _, arg := range []string{relationFitchUG} {
		if _, r := CheckEntailmentGate(arg); !r.Passed {
			t.Fatalf("%s\n%s", arg, r.Message)
		}
	}
}
func TestEntailmentGateV17NearMisses(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    []string
		c    string
	}{
		{"UI wrong class", []string{"Alice greets every teacher", "Bob is a student"}, "Alice greets Bob"},
		{"UI wrong position", []string{"Alice greets every teacher", "Bob is a teacher"}, "Bob greets Alice"},
		{"UI wrong verb", []string{"Alice greets every teacher", "Bob is a teacher"}, "Alice respects Bob"},
		{"EG wrong class", []string{"Alice greets Bob", "Bob is a student"}, "Alice greets some teacher"},
		{"EG wrong name", []string{"Alice greets Bob", "Alice is a teacher"}, "Alice greets some teacher"},
		{"EG wrong position", []string{"Alice greets Bob", "Bob is a teacher"}, "Some teacher greets Alice"},
		{"EG wrong verb", []string{"Alice greets Bob", "Bob is a teacher"}, "Alice respects some teacher"},
		{"negative wrong class", []string{"Alice greets no teacher", "Bob is a student"}, "It is not the case that Alice greets Bob"},
		{"negative wrong position", []string{"Alice greets no teacher", "Bob is a teacher"}, "It is not the case that Bob greets Alice"},
		{"negative wrong verb", []string{"Alice greets no teacher", "Bob is a teacher"}, "It is not the case that Alice respects Bob"},
		{"negation wrong position", []string{"It is not the case that some student greets Bob"}, "No student is greeted by Bob"},
		{"negation wrong class", []string{"It is not the case that some student greets Bob"}, "No teacher greets Bob"},
		{"scope flip", []string{"Every student is such that that student greets some teacher"}, "Some teacher is such that every student greets that teacher"},
		{"ambiguity verbatim", []string{"Every student greets some teacher"}, "Some teacher is such that every student greets that teacher"},
		{"ambiguous conclusion", []string{"Every student is such that that student greets some teacher"}, "Every student greets some teacher"},
		{"passive conversion", []string{"Alice greets every teacher", "Bob is a teacher"}, "Bob is greeted by Alice"},
		{"symmetry", []string{"Alice greets Bob"}, "Bob greets Alice"},
		{"any", []string{"Alice greets any teacher", "Bob is a teacher"}, "Alice greets Bob"},
		{"plural", []string{"All students greet every teacher", "Alice is a student"}, "Alice greets every teacher"},
		{"nested object noun", []string{"Alice greets every student of some school", "Bob is a student of some school"}, "Alice greets Bob"},
		{"scaffold wrong class", []string{"Every student greets Bob"}, "Every teacher is such that that teacher greets Bob"},
		{"scaffold wrong position", []string{"Every student greets Bob"}, "Every student is such that Bob greets that student"},
		{"scaffold wrong verb", []string{"Every student greets Bob"}, "Every student is such that that student respects Bob"},
		{"noun nesting", []string{"Every student of some school greets Bob", "Alice is a student of some school"}, "Alice greets Bob"},
		{"third quantified slot", []string{"Every student gives every teacher some book", "Alice is a student"}, "Alice gives every teacher some book"},
		{"extra citation", []string{"Alice greets Bob", "Bob is a teacher", "Alice is a student"}, "Alice greets some teacher"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg := relationArgument(tc.p, tc.c)
			if _, r := CheckEntailmentGate(arg); r.Passed {
				t.Fatalf("accepted %s", arg)
			}
		})
	}
	for _, arg := range []string{
		strings.Replace(relationFitchUG, "1. Every student greets every teacher.", "1. b greets every teacher.", 1),
		strings.Replace(relationFitchUG, "b greets every teacher", "B greets every teacher", 1),
		strings.Replace(relationFitchUG, "4. Therefore, Every student", "4. Therefore, Every teacher", 1),
		strings.Replace(relationFitchUG, "4. Therefore, Every student greets every teacher.", "4. Therefore, Some teacher is such that every student greets that teacher.", 1),
		strings.Replace(relationFitchEE, "6. Therefore, Some student greets Bob.", "6. Therefore, c greets Bob.", 1),
		strings.ReplaceAll(relationFitchEE, "c ", "Bob "),
	} {
		if _, r := CheckEntailmentGate(arg); r.Passed {
			t.Fatalf("accepted %s", arg)
		}
	}
}
func TestEntailmentGateV17AmbiguityNote(t *testing.T) {
	report, r := CheckEntailmentGate("1. Every student admires some teacher.\n2. Therefore, Some teacher is such that every student admires that teacher. [1]")
	want := `step 2: "every student admires some teacher" has two scope readings; this step follows only if some teacher is admired by every student. To mean that, write "Some teacher is such that every student admires that teacher."`
	if r.Passed || len(report.Steps) != 1 || report.Steps[0].AmbiguityNote != want || !strings.Contains(r.Message, want) {
		t.Fatalf("%+v %s", report, r.Message)
	}
	_, r = CheckEntailmentGate(relationArgument([]string{"Every student of some school admires Bob", "Alice is a student of some school"}, "Alice admires Bob"))
	if !strings.Contains(r.Message, "noun phrases containing nested quantifiers") {
		t.Fatal(r.Message)
	}
}

func TestEntailmentGateV17ConsistentScopeOccurrences(t *testing.T) {
	report, r := CheckEntailmentGate(relationArgument([]string{"Every student admires some teacher"}, "Every student admires some teacher"))
	if !r.Passed || len(report.Steps) != 1 || report.Steps[0].Form != "self-entailment" || report.Steps[0].AmbiguityNote != "" {
		t.Fatalf("%+v %s", report, r.Message)
	}
}

func TestEntailmentGateV17InertHintsRespectScope(t *testing.T) {
	arg, syntax := ParseNumberedArgument(relationArgument([]string{"Every student admires some teacher", "it rains"}, "Every student admires some teacher"), numberedArgumentOptions{})
	if !syntax.Passed {
		t.Fatal(syntax.Message)
	}
	hints := NumberedArgumentInertCitations(arg)
	if len(hints) != 1 || hints[0].Form != "self-entailment" || len(hints[0].InertCitations) != 1 || hints[0].InertCitations[0] != 2 {
		t.Fatalf("consistent reiteration must support the inert hint: %+v", hints)
	}
}

func TestEntailmentGateV17OpaqueClauseCertificates(t *testing.T) {
	const p = "Every student admires some teacher"
	const q = "the school thrives"
	const r = "the bell rings"
	for _, tc := range []struct {
		p       []string
		c, form string
	}{
		{[]string{p}, strings.ToLower(p), "self-entailment"},
		{[]string{"If " + strings.ToLower(p) + ", then " + q, p}, q, "modus ponens"},
		{[]string{"If " + q + ", then " + p, q}, p, "modus ponens"},
		{[]string{"If " + q + ", then " + p, "If " + p + ", then " + r}, "If " + q + ", then " + r, "hypothetical syllogism"},
		{[]string{q + " and " + p}, p, "conjunction elimination"},
		{[]string{q + " if and only if " + p, p}, q, "biconditional modus ponens"},
		{[]string{"If " + p + ", then " + q, "It is not the case that " + q}, "It is not the case that " + p, "modus tollens"},
		{[]string{"If " + p + ", then " + q, "If " + q + ", then " + r}, "If " + p + ", then " + r, "hypothetical syllogism"},
		{[]string{p, q}, p + " and " + q, "conjunction introduction"},
		{[]string{p + " and " + q}, p, "conjunction elimination"},
		{[]string{p}, "Either " + p + " or " + q, "disjunction introduction"},
		{[]string{"Either " + p + " or " + q, "It is not the case that " + q}, p, "disjunctive syllogism"},
		{[]string{"Either " + p + " or " + q, "It is not the case that " + p}, q, "disjunctive syllogism"},
		{[]string{p + " if and only if " + q, p}, q, "biconditional modus ponens"},
		{[]string{p + " if and only if " + q}, "If " + p + ", then " + q, "biconditional elimination"},
		{[]string{"If " + p + ", then " + q, "If " + q + ", then " + p}, p + " if and only if " + q, "biconditional introduction"},
		{[]string{"Either " + p + " or " + r, "If " + p + ", then " + q, "If " + r + ", then " + q}, q, "simple dilemma"},
	} {
		t.Run(tc.form+"/"+tc.c, func(t *testing.T) {
			argument := relationArgument(tc.p, tc.c)
			report, result := CheckEntailmentGate(argument)
			if !result.Passed || report.Steps[len(report.Steps)-1].Form != tc.form {
				t.Fatalf("%s\n%+v %s", argument, report, result.Message)
			}
		})
	}
	for _, tc := range []struct{ argument, form string }{
		{"| 1. Suppose every student admires some teacher.\n| 2. Therefore, Every student admires some teacher. [1]\n3. Therefore, if every student admires some teacher, then every student admires some teacher. [1-2]", "conditional proof"},
		{"1. It is not the case that every student admires some teacher.\n| 2. Suppose every student admires some teacher.\n| 3. Therefore, contradiction. [1,2]\n4. Therefore, it is not the case that every student admires some teacher. [2-3]", "reductio"},
		{"1. Either every student admires some teacher or the bell rings.\n2. If every student admires some teacher, then the school thrives.\n3. If the bell rings, then the school thrives.\n| 4. Suppose every student admires some teacher.\n| 5. Therefore, the school thrives. [2,4]\n| 6. Suppose the bell rings.\n| 7. Therefore, the school thrives. [3,6]\n8. Therefore, the school thrives. [1,4-5,6-7]", "proof by cases"},
	} {
		t.Run(tc.form, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if !result.Passed || report.Steps[len(report.Steps)-1].Form != tc.form {
				t.Fatalf("%s\n%+v %s", tc.argument, report, result.Message)
			}
		})
	}
}

func TestEntailmentGateV17DistinctClausesStayDistinct(t *testing.T) {
	argument := relationArgument([]string{
		"If every student greets some teacher, then the school thrives",
		"Every student greets some teacher of mathematics",
	}, "The school thrives")
	if _, result := CheckEntailmentGate(argument); result.Passed {
		t.Fatal("a longer distinct clause must not be replaced as an occurrence of its prefix")
	}
}

func TestEntailmentGateV17UGScopeWeakening(t *testing.T) {
	const argument = `1. Every student is such that that student greets no student.
| 2. Let b be an arbitrary student.
| 3. Therefore, b greets no student. [1,2]
4. Therefore, Every student greets no student. [2-3]`
	if report, result := CheckEntailmentGate(argument); !result.Passed || report.Steps[len(report.Steps)-1].Form != "relational universal generalization" {
		t.Fatalf("%+v %s", report, result.Message)
	}
	// With a different inner class, an empty student class blocks the
	// existential weakening in the object-wide reading of the conclusion.
	wrongClass := strings.ReplaceAll(argument, "no student", "no teacher")
	if _, result := CheckEntailmentGate(wrongClass); result.Passed {
		t.Fatal("universal classes must not import existence")
	}
}
