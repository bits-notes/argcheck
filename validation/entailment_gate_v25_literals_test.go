package validation

import (
	"strings"
	"testing"
)

func TestCatalogV25LiteralNames(t *testing.T) {
	cases := []struct {
		text string
		want *folFormula
	}{
		{"every x is such that x admires name{x}", v24Q("every", "$x", foAtom("admires", "$x", "x"))},
		{"every y2 is such that name{y2} is identical to y2", v24Q("every", "$y2", &folFormula{op: "equal", terms: []string{"y2", "$y2"}})},
		{"name{x} is very active", foAtom("is/very active", "x")},
		{"name{x} is identical to name{x}", &folFormula{op: "equal", terms: []string{"x", "x"}}},
		{`name{"claim that it rains"} is identical to name{"$slot0"}`, &folFormula{op: "equal", terms: []string{"claim that it rains", legacyFOLTerm("$slot0")}}},
	}
	for _, tc := range cases {
		f, err := parseFOL(tc.text)
		if err != nil || !folSame(f, tc.want) {
			t.Fatalf("%s: %v", tc.text, err)
		}
		v24Check(t, v24Flat([]*folFormula{tc.want}, tc.want))
	}
	v25Conversion(t, "X is red", "name{x} is red")
	v24Check(t, "1. Every x is such that x admires name{x}.\n2. Therefore, alice admires name{x}. [1]")
	v24Check(t, "1. name{x} is identical to alice.\n2. name{x} is red.\n3. Therefore, alice is red. [1,2]")
	f, _ := parseFOL("x is red")
	if !folSame(f, foAtom("is/red", "x")) {
		t.Fatal("bare legacy name changed")
	}
}
func TestCatalogV25LiteralAtoms(t *testing.T) {
	cases := []*folFormula{
		foAtom("works with", "alice", "$x"),
		foAtom("is/nearby is retracted", "claim that the store"),
		foAtom("either it rains"),
		foAtom("every hero is such that it rains"),
		foAtom(`literal ( and ) or "quoted" text`),
		foAtom("is/red", "x"),
		foAtom("p", "alice", "blair", "casey"),
		foAtom("is/red", legacyFOLTerm("$slot0")),
	}
	for _, f := range cases {
		text := folRenderLiteral(f)
		got, err := parseFOLScope(text, map[string]bool{"x": true}, true)
		if err != nil || !folSame(f, got) {
			t.Fatalf("%s: %v", text, err)
		}
		rendered := f.Render()
		round, roundErr := parseFOLScope(rendered, map[string]bool{"x": true}, true)
		if roundErr != nil || !folSame(f, round) {
			t.Fatalf("literal render %s: %v", rendered, roundErr)
		}
		if len(f.terms) == 0 {
			c := fol("and", f, foAtom("it rains"))
			text := "(" + folRenderLiteral(f) + ") and (it rains)"
			got, err = parseFOL(text)
			if err != nil || !folSame(c, got) {
				t.Fatalf("compound %s: %v", text, err)
			}
			v24Check(t, "1. "+text+".\n2. Therefore, "+folRenderLiteral(f)+". [1]")
		}
	}
	for _, text := range []string{`atom{"admires", x}`, `every x is such that atom{"admires", y}`, `atom{"p", "alice"} trailing`, `atom{"p",}`, `atom{"p"`, `name{alice} is red`, `name{x} is name{x}`} {
		if _, err := parseFOL(text); err == nil {
			t.Fatal("malformed literal accepted", text)
		}
	}
	if folSame(foAtom("it rains"), foAtom("it rains", "alice")) {
		t.Fatal("arity collapsed")
	}
}
func TestCatalogV25OpaqueFreshRenaming(t *testing.T) {
	if _, err := parseFOL("every x is such that x would admire alice"); err == nil {
		t.Fatal("opaque guard weakened")
	}
	f, err := parseFOL("every y is such that x would admire alice")
	if err != nil {
		t.Fatal(err)
	}
	want := v24Q("every", "$z", foAtom("x would admire alice"))
	if !folSame(f, want) {
		t.Fatal("renaming changed opaque atom")
	}
	rendered := want.Render()
	if strings.HasPrefix(rendered, "every x ") {
		t.Fatal("renderer failed to freshen binder", rendered)
	}
	got, err := parseFOL(rendered)
	if err != nil || !folSame(got, want) {
		t.Fatal(rendered, err)
	}
	v24Check(t, "1. Every y is such that x would admire alice.\n2. Therefore, x would admire alice. [1]")
}

func TestCatalogV25LiteralFreshness(t *testing.T) {
	for _, argument := range []string{
		`1. Some x is such that x is red.
| 2. Suppose name{"$slot0"} is red.
| 3. Therefore, atom{"is/red", "$slot0"}. [2]
4. Therefore, atom{"is/red", "$slot0"}. [1,2-3]`,
		`1. atom{"is/red", "\u0077itness"}.
| 2. Let witness be arbitrary.
| 3. Therefore, witness is red. [1]
4. Therefore, every x is such that x is red. [2-3]`,
		`1. Some x is such that x is red.
| 2. Suppose witness is red.
| 3. Therefore, atom{"is/red", "\u0077itness"}. [2]
4. Therefore, atom{"is/red", "\u0077itness"}. [1,2-3]`,
	} {
		report, r := CheckEntailmentGate(argument)
		if r.Passed {
			t.Fatal("escaped eigenname escaped freshness check", argument)
		}
		if len(report.Steps) != 2 || report.Steps[0].Step != 3 || !report.Steps[0].Passed || report.Steps[1].Step != 4 || report.Steps[1].Passed {
			t.Fatalf("expected only the discharge to fail: %+v", report)
		}
	}
	f, err := parseFOL(`name{"\u0058"} is red`)
	if err != nil || !folSame(f, foAtom("is/red", "x")) {
		t.Fatal("decoded case normalization", err)
	}
}

func TestCatalogV25LiteralSignatureRoundTrip(t *testing.T) {
	symbols := []string{"p", "is/very red", "works/with", "either it rains", `a ( b ) and c, "d"`, "every x is such that rain", "can't fly", "東京", "literal [1] and [x]", "a  b", "a\u2003b"}
	constants := []string{"x", "y2", "claim that the store is nearby", "either alice or blair", legacyFOLTerm("$slot0"), "alice", "can't", "東京"}
	for i, symbol := range symbols {
		for arity := 0; arity < 6; arity++ {
			terms := []string{}
			for j := 0; j < arity; j++ {
				terms = append(terms, constants[(i+j)%len(constants)])
			}
			if arity > 0 {
				terms[0] = "$v"
			}
			f := v24Q("every", "$v", foAtom(symbol, terms...))
			text := f.Render()
			got, err := parseFOL(text)
			if err != nil || !folSame(f, got) {
				t.Fatalf("%s: %v", text, err)
			}
			v24Check(t, v24Flat([]*folFormula{f}, f))
			v24Check(t, v24Flat([]*folFormula{f}, folSub(f.children[0], f.symbol, "witness")))
		}
	}
	for _, a := range constants {
		f := &folFormula{op: "equal", terms: []string{a, "alice"}}
		got, err := parseFOL(f.Render())
		if err != nil || !folSame(f, got) {
			t.Fatal(f.Render(), err)
		}
		v24Check(t, v24Flat([]*folFormula{f}, f))
	}
}
