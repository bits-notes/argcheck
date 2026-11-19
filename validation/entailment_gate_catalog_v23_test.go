package validation

import (
	"testing"
	"time"
)

const v23A2 = `1. Every regulator that Anthropic lobbies is a regulator that Anthropic funds.
2. Anthropic lobbies the FTC.
3. The FTC is a regulator.
4. Therefore, the FTC is a regulator that Anthropic funds. [1,2,3]`

// Formulas below are supplied independently of the production English parser.
// In particular, a relative noun is a conjunction, never an atomic class label.
func rcSurface(text string, f *finiteFormula) nOracleSurface {
	return nOracleSurface{text, []*finiteFormula{f}}
}
func rcBind(kind string, guard, body *finiteFormula) *finiteFormula {
	if kind == "no" {
		return foNot(rcBind("some", guard, body))
	}
	if kind == "every" {
		body = foOr(foNot(guard), body)
	} else {
		body = foAnd(guard, body)
	}
	return &finiteFormula{op: kind, symbol: "$x", children: []*finiteFormula{body}}
}
func relativeExercises() []nplaceExercise {
	var out []nplaceExercise
	add := func(name string, p []nOracleSurface, c nOracleSurface, valid bool) {
		var texts []string
		for _, s := range p {
			texts = append(texts, s.text)
		}
		out = append(out, nplaceExercise{name, relationArgument(texts, c.text), p, c, valid})
	}
	atom := foAtom
	regulator := func(x string) *finiteFormula { return atom("regulator", x) }
	lobby := func(x string) *finiteFormula { return atom("lobbies", "firm", x) }
	fund := func(x string) *finiteFormula { return atom("funds", "firm", x) }
	rule := rcSurface("Every regulator that Anthropic lobbies is a regulator that Anthropic funds", rcBind("every", foAnd(regulator("$x"), lobby("$x")), foAnd(regulator("$x"), fund("$x"))))
	member := rcSurface("The FTC is a regulator", regulator("ftc"))
	relation := rcSurface("Anthropic lobbies the FTC", lobby("ftc"))
	conclusion := rcSurface("The FTC is a regulator that Anthropic funds", foAnd(regulator("ftc"), fund("ftc")))
	add("a2", []nOracleSurface{rule, relation, member}, conclusion, true)
	add("missing relation", []nOracleSurface{rule, member}, conclusion, false)
	// The changed relation premise must not itself supply the whole conclusion.
	target := rcSurface("Anthropic funds the FTC", fund("ftc"))
	add("unmatched relation", []nOracleSurface{rcSurface("Every regulator that Anthropic lobbies is active", rcBind("every", foAnd(regulator("$x"), lobby("$x")), atom("active", "$x"))), target, member}, rcSurface("The FTC is active", atom("active", "ftc")), false)
	add("swapped gap", []nOracleSurface{rule, rcSurface("The FTC lobbies Anthropic", atom("lobbies", "ftc", "firm")), member}, conclusion, false)
	add("missing class", []nOracleSurface{rule, relation}, conclusion, false)
	add("strengthened conclusion", []nOracleSurface{rule, relation, member}, rcSurface("The FTC is a regulator that Anthropic funds and the FTC is active", foAnd(regulator("ftc"), fund("ftc"), atom("active", "ftc"))), false)
	for _, pronoun := range []string{"that", "who", "which"} {
		guard := foAnd(atom("runner", "$x"), atom("runs", "$x"))
		add("unary "+pronoun, []nOracleSurface{rcSurface("Every runner "+pronoun+" runs is active", rcBind("every", guard, atom("active", "$x"))), rcSurface("Robin is a runner", atom("runner", "r")), rcSurface("Robin runs", atom("runs", "r"))}, rcSurface("Robin is active", atom("active", "r")), true)
	}
	guard := foAnd(atom("reader", "$x"), atom("admires", "$x", "p"))
	add("subject gap", []nOracleSurface{rcSurface("Every reader who admires Pat is active", rcBind("every", guard, atom("active", "$x"))), rcSurface("Robin is a reader", atom("reader", "r")), rcSurface("Robin admires Pat", atom("admires", "r", "p"))}, rcSurface("Robin is active", atom("active", "r")), true)
	hero := foAnd(atom("hero", "$x"), atom("admires", "s", "$x"))
	hmember := rcSurface("Robin is a hero", atom("hero", "r"))
	hrel := rcSurface("Sam admires Robin", atom("admires", "s", "r"))
	add("object universal", []nOracleSurface{rcSurface("Pat admires every hero that Sam admires", rcBind("every", hero, atom("admires", "p", "$x"))), hmember, hrel}, rcSurface("Pat admires Robin", atom("admires", "p", "r")), true)
	add("object existential", []nOracleSurface{rcSurface("Pat admires Robin", atom("admires", "p", "r")), hmember, hrel}, rcSurface("Pat admires some hero that Sam admires", rcBind("some", hero, atom("admires", "p", "$x"))), true)
	add("subject existential", []nOracleSurface{hmember, hrel, rcSurface("Robin is active", atom("active", "r"))}, rcSurface("Some hero that Sam admires is active", rcBind("some", hero, atom("active", "$x"))), true)
	add("object no", []nOracleSurface{rcSurface("Pat admires no hero that Sam admires", rcBind("no", hero, atom("admires", "p", "$x"))), hmember, hrel}, rcSurface("It is not the case that Pat admires Robin", foNot(atom("admires", "p", "r"))), true)
	add("subject no", []nOracleSurface{rcSurface("No hero that Sam admires is active", rcBind("no", hero, atom("active", "$x"))), hmember, hrel}, rcSurface("It is not the case that Robin is active", foNot(atom("active", "r"))), true)
	nominal := rcSurface("Robin is a hero that Sam admires", foAnd(atom("hero", "r"), atom("admires", "s", "r")))
	add("nominal class elimination", []nOracleSurface{nominal}, hmember, true)
	add("nominal relation elimination", []nOracleSurface{nominal}, hrel, true)
	add("nominal introduction", []nOracleSurface{hmember, hrel}, nominal, true)
	book := foAnd(atom("book", "$x"), atom("gave", "a", "b", "$x"))
	add("three-place object gap", []nOracleSurface{rcSurface("Every book that Alex gave Blair is rare", rcBind("every", book, atom("rare", "$x"))), rcSurface("Atlas is a book", atom("book", "atlas")), rcSurface("Alex gave Blair Atlas", atom("gave", "a", "b", "atlas"))}, rcSurface("Atlas is rare", atom("rare", "atlas")), true)
	add("three-place subject gap", []nOracleSurface{rcSurface("Alex is a courier", atom("courier", "a")), rcSurface("Alex gave Blair Atlas", atom("gave", "a", "b", "atlas"))}, rcSurface("Alex is a courier who gave Blair Atlas", foAnd(atom("courier", "a"), atom("gave", "a", "b", "atlas"))), true)
	add("existence does not name witness", []nOracleSurface{rcSurface("Some hero that Sam admires is active", rcBind("some", hero, atom("active", "$x")))}, rcSurface("Robin is active", atom("active", "r")), false)
	add("nominal role swap", []nOracleSurface{nominal}, rcSurface("Robin admires Sam", atom("admires", "r", "s")), false)
	add("nominal missing relation", []nOracleSurface{hmember}, nominal, false)
	add("nominal missing class", []nOracleSurface{hrel}, nominal, false)
	add("no loses negation", []nOracleSurface{rcSurface("No hero that Sam admires is active", rcBind("no", hero, atom("active", "$x"))), hmember, hrel}, rcSurface("Robin is active", atom("active", "r")), false)
	add("no existential import", []nOracleSurface{rcSurface("Every hero that Sam admires is active", rcBind("every", hero, atom("active", "$x")))}, rcSurface("Some hero that Sam admires is active", rcBind("some", hero, atom("active", "$x"))), false)
	add("ternary role swap", []nOracleSurface{rcSurface("Alex is a courier", atom("courier", "a")), rcSurface("Alex gave Atlas Blair", atom("gave", "a", "atlas", "b"))}, rcSurface("Alex is a courier who gave Blair Atlas", foAnd(atom("courier", "a"), atom("gave", "a", "b", "atlas"))), false)
	nestedBody := rcBind("every", foAnd(foAtom("reader", "$x"), foAtom("visits", "$x", "$h")), foAtom("admires", "p", "$x"))
	outerGuard := foAnd(foAtom("hero", "$h"), foAtom("admires", "s", "$h"))
	outer := &finiteFormula{op: "every", symbol: "$h", children: []*finiteFormula{foOr(foNot(outerGuard), nestedBody)}}
	inner := rcBind("every", foAnd(foAtom("reader", "$x"), foAtom("visits", "$x", "r")), foAtom("admires", "p", "$x"))
	add("explicit nested scaffold", []nOracleSurface{rcSurface("Every hero that Sam admires is such that Pat admires every reader who visits that hero", outer), hmember, hrel}, rcSurface("Pat admires every reader who visits Robin", inner), true)

	return out
}
func TestCatalogV23Exercises(t *testing.T) {
	if EntailmentGateCatalogVersion != "v44.0" {
		t.Fatal(EntailmentGateCatalogVersion)
	}
	for _, tc := range relativeExercises() {
		t.Run(tc.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if result.Passed != tc.valid {
				t.Fatalf("want %v: %+v\n%+v\n%s", tc.valid, result, report, tc.argument)
			}
		})
	}
}
func TestCatalogV23FiniteOracle(t *testing.T) {
	start := time.Now()
	valid, invalid := 0, 0
	for _, tc := range relativeExercises() {
		t.Run(tc.name, func(t *testing.T) {
			if nOracleInvalid(tc.p, tc.c) == tc.valid {
				t.Fatal("oracle disagrees")
			}
			if tc.valid {
				valid++
			} else {
				invalid++
			}
		})
	}
	t.Logf("domains 1–3; %d valid, %d invalid; exhaustive finite search; runtime %s", valid, invalid, time.Since(start))
}

const v23FitchEE = `1. Some hero that Sam admires is active.
| 2. Suppose z is a hero that Sam admires and z is active.
| 3. Therefore, z is a hero that Sam admires. [2]
| 4. Therefore, z is a hero. [3]
| 5. Therefore, Sam admires z. [3]
| 6. Therefore, z is active. [2]
| 7. Therefore, Some hero that Sam admires is active. [4,5,6]
8. Therefore, Some hero that Sam admires is active. [1,2-7]`
const v23FitchUG = `1. Every hero that Sam admires is active.
| 2. Let z be an arbitrary hero that Sam admires.
| 3. Therefore, z is a hero. [2]
| 4. Therefore, Sam admires z. [2]
| 5. Therefore, z is active. [1,3,4]
6. Therefore, Every hero that Sam admires is active. [2-5]`

func TestCatalogV23Fitch(t *testing.T) {
	// scope.clause_start: v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit).
	assertStrictRefusal(t, v23FitchEE, "scope.clause_start", "move the operator into each clause")
	for _, argument := range []string{v23FitchUG} {
		report, result := CheckEntailmentGate(argument)
		if !result.Passed {
			t.Fatalf("%+v\n%+v", result, report)
		}
	}
}
func TestCatalogV23FitchOracle(t *testing.T) {
	start := time.Now()
	hero := foAtom("hero", "z")
	relation := foAtom("admires", "s", "z")
	active := foAtom("active", "z")
	nominal := foAnd(hero, relation)
	assumption := foAnd(nominal, active)
	guard := foAnd(foAtom("hero", "$x"), foAtom("admires", "s", "$x"))
	universal := rcBind("every", guard, foAtom("active", "$x"))
	existential := rcBind("some", guard, foAtom("active", "$x"))
	obligations := []struct {
		p []*finiteFormula
		c *finiteFormula
	}{
		{[]*finiteFormula{assumption}, nominal}, {[]*finiteFormula{nominal}, hero}, {[]*finiteFormula{nominal}, relation},
		{[]*finiteFormula{assumption}, active}, {[]*finiteFormula{hero, relation, active}, existential},
		{[]*finiteFormula{universal, hero, relation}, active},
	}
	for i, o := range obligations {
		if finiteCountermodel(o.p, o.c, 3) {
			t.Fatalf("local obligation %d", i)
		}
	}
	// Discharge: the branch is universally closed over the fresh eigenname.
	branch := rcBind("every", foAnd(guard, foAtom("active", "$x")), existential)
	if finiteCountermodel([]*finiteFormula{existential, branch}, existential, 3) {
		t.Fatal("EE discharge")
	}
	if finiteCountermodel([]*finiteFormula{universal}, universal, 3) {
		t.Fatal("UG discharge")
	}
	t.Logf("domains 1–3; 6 local obligations and 2 discharges; runtime %s", time.Since(start))
}

func TestCatalogV23UnusedCitation(t *testing.T) {
	argument := `1. Every regulator that Anthropic lobbies is a regulator that Anthropic funds.
2. Anthropic funds the FTC.
3. The FTC is a regulator.
4. Therefore, the FTC is a regulator that Anthropic funds. [1,2,3]`
	if _, r := CheckEntailmentGate(argument); r.Passed {
		t.Fatal("unused universal citation accepted")
	}
	// This is logically VALID; only the citation discipline rejects it. There
	// cannot be a countermodel for the negative case as originally specified.
	c := foAnd(foAtom("regulator", "f"), foAtom("funds", "a", "f"))
	if finiteCountermodel([]*finiteFormula{foAtom("regulator", "f"), foAtom("funds", "a", "f")}, c, 3) {
		t.Fatal("direct conjunction")
	}
}

func TestCatalogV23NestingDiagnostic(t *testing.T) {
	argument := `1. Pat admires every reader who visits a hero that Sam admires.
2. Robin is a reader.
3. Therefore, Pat admires Robin. [1,2]`
	report, result := CheckEntailmentGate(argument)
	if result.Passed || len(report.Steps) != 1 || report.Steps[0].NearestForm != "relative clause inference" {
		t.Fatalf("%+v", report)
	}
	// Even a nested restrictor cannot be dropped: take no visiting readers.
	inner := rcBind("some", foAnd(foAtom("hero", "$x"), foAtom("admires", "s", "$x")), foAtom("visits", "$r", "$x"))
	premise := &finiteFormula{op: "every", symbol: "$r", children: []*finiteFormula{foOr(foNot(foAnd(foAtom("reader", "$r"), inner)), foAtom("admires", "p", "$r"))}}
	if !finiteCountermodel([]*finiteFormula{premise, foAtom("reader", "r")}, foAtom("admires", "p", "r"), 3) {
		t.Fatal("missing countermodel")
	}
}

func TestCatalogV23FitchInvalid(t *testing.T) {
	cases := []struct {
		argument string
		p, c     *finiteFormula
	}{
		{`1. Some hero that Sam admires is active.
| 2. Suppose z is a hero that Sam admires and z is active.
| 3. Therefore, z is a hero that Sam admires. [2]
| 4. Therefore, Sam admires z. [3]
5. Therefore, Sam admires z. [1,2-4]`,
			rcBind("some", foAnd(foAtom("hero", "$x"), foAtom("admires", "s", "$x")), foAtom("active", "$x")), foAtom("admires", "s", "z")},
		{`1. Some hero that Sam admires is active.
| 2. Suppose z is a hero that Sam admires and z is a reader who visits Pat.
| 3. Therefore, z is a reader who visits Pat. [2]
| 4. Therefore, z visits Pat. [3]
| 5. Therefore, z is a reader. [3]
| 6. Therefore, Some reader visits Pat. [4,5]
7. Therefore, Some reader visits Pat. [1,2-6]`,
			rcBind("some", foAnd(foAtom("hero", "$x"), foAtom("admires", "s", "$x")), foAtom("active", "$x")), rcBind("some", foAtom("reader", "$x"), foAtom("visits", "$x", "p"))},
	}
	for _, tc := range cases {
		if _, r := CheckEntailmentGate(tc.argument); r.Passed {
			t.Fatal("invalid witness accepted")
		}
		if !finiteCountermodel([]*finiteFormula{tc.p}, tc.c, 3) {
			t.Fatal("missing witness countermodel")
		}
	}
	// Test the reverse-introduction guard directly: additional witness facts
	// cannot be assumed even when every required fact is present.
	first := NumberedArgumentStep{Statement: "z is a hero that Sam admires and z is a reader who visits Pat"}
	existential := NumberedArgumentStep{Statement: "Some hero that Sam admires is a reader"}
	if relativeEEName(first, existential) != "" {
		t.Fatal("extra witness conjunct accepted")
	}
	p := rcBind("some", foAnd(foAtom("hero", "$x"), foAtom("admires", "s", "$x")), foAtom("reader", "$x"))
	c := rcBind("some", foAtom("reader", "$x"), foAtom("visits", "$x", "p"))
	if !finiteCountermodel([]*finiteFormula{p}, c, 3) {
		t.Fatal("extra conjunct countermodel")
	}
}

func TestCatalogV23GrammarBoundaries(t *testing.T) {
	// Truth operators are not predicate nominals, and an outer negation cannot
	// be mistaken for a relation's subject while instantiating its object.
	for _, s := range []string{"It is true that Sam runs", "It is false that Sam admires Robin"} {
		if _, ok := relativeGround(s); ok {
			t.Fatalf("non-nominal decomposed: %s", s)
		}
	}
	for _, s := range []string{
		"It is false that Pat admires every hero that Sam admires",
		"Every hero that Sam admires is such that some hero is such that Pat admires that hero",
		"Every hero that Sam admires is such that Pat admires zzrelativeholezz and that hero runs",
	} {
		if _, ok := relativeQuant(s); ok {
			t.Fatalf("unsafe binder: %s", s)
		}
	}
}
