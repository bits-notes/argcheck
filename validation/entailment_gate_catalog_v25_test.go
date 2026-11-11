package validation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
)

// One fifth of the generator's formulas require 30–40 simultaneous binders;
// every binder occurs below the entire prefix, not merely in a disjoint scope.
func v25Formula(r *rand.Rand, large bool) *folFormula {
	if !large {
		return v24Formula(r)
	}
	n := 30 + r.Intn(11)
	var f *folFormula
	for i := 0; i < n; i++ {
		atom := foAtom("admires", fmt.Sprintf("$v%d", i), fmt.Sprintf("$v%d", (i+1)%n))
		if f == nil {
			f = atom
		} else {
			f = fol("and", f, atom)
		}
	}
	for i := n - 1; i >= 0; i-- {
		kind := "every"
		if r.Intn(2) == 0 {
			kind = "some"
		}
		f = v24Q(kind, fmt.Sprintf("$v%d", i), f)
	}
	return f
}
func TestCatalogV25UnboundedVariables(t *testing.T) {
	r := rand.New(rand.NewSource(25001))
	for i := 0; i < 1000; i++ {
		f := v25Formula(r, i%5 == 0)
		got, err := parseFOL(f.Render())
		if err != nil || !folSame(f, got) {
			t.Fatalf("round trip %d: %v", i, err)
		}
		if i%5 == 0 {
			// Exercise the gate, not just the parser, on every large generated formula.
			v24Check(t, v24Flat([]*folFormula{f}, f))
			if f.op == "every" {
				v24Check(t, v24Flat([]*folFormula{f}, folSub(f.children[0], f.symbol, "witness")))
			}
		}
	}
	for _, n := range []int{26, 27, 30, 64} {
		f := foAtom("admires", "$v0", fmt.Sprintf("$v%d", n-1))
		for i := n - 1; i >= 0; i-- {
			f = v24Q("every", fmt.Sprintf("$v%d", i), f)
		}
		got, err := parseFOL(f.Render())
		if err != nil || !folSame(f, got) {
			t.Fatal(n, err)
		}
	}
	for _, s := range []string{
		"every x1 is such that x2 admires x1",
		"every x1 is such that some x1 is such that x1 is red",
		"every x1 is such that (x1 is red) and (y12 is blue)",
		"every x1 is such that x1 is identical to x2",
	} {
		_, r := CheckEntailmentGate("1. " + s + ".\n2. Therefore, " + s + ". [1]")
		if r.Passed || !strings.Contains(r.Message, "nearest form") {
			t.Fatal(s, r.Message)
		}
	}
	for _, s := range []string{
		"every x1 is such that some y12 is such that x1 admires y12",
		"(every x1 is such that x1 is red) and (some x1 is such that x1 is blue)",
	} {
		f, err := parseFOL(s)
		if err != nil {
			t.Fatal(err)
		}
		v24Check(t, v24Flat([]*folFormula{f}, f))
	}
	// The implicit book witness must also work after exhausting all bare letters.
	s := "x gave y a book"
	for _, v := range "abcdefghijklmnopqrstuvwxyz" {
		s = "every " + string(v) + " is such that " + s
	}
	f, err := parseFOL(s)
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseFOL(f.Render())
	if err != nil || !folSame(f, got) {
		t.Fatal("implicit witness", err)
	}
	t.Log("1000/1000 round trips; 200 with 30–40 simultaneous, nonvacuous binders; explicit boundary cases 26, 27, 30, 64; gate and scope guards passed")
}
func v25Conversion(t *testing.T, a, b string) {
	t.Helper()
	for _, pair := range [][2]string{{a, b}, {b, a}} {
		steps := map[int]NumberedArgumentStep{1: {Number: 1, Statement: pair[0]}}
		c := NumberedArgumentStep{Number: 2, Statement: pair[1], Citations: []int{1}}
		r := notationEquivalence(c, steps)
		if !r.Passed || r.Form != "notation equivalence" {
			t.Fatalf("%s -> %s: %+v", pair[0], pair[1], r)
		}
		v24Check(t, "1. "+pair[0]+".\n2. Therefore, "+pair[1]+". [1]")
	}
}
func TestCatalogV25NotationEquivalence(t *testing.T) {
	cases := []struct {
		s string
		f *folFormula
	}{
		{"Every student is red", v24Q("every", "$a", fol("implies", foAtom("is/student", "$a"), foAtom("is/red", "$a")))},
		{"Some student is red", v24Q("some", "$a", fol("and", foAtom("is/student", "$a"), foAtom("is/red", "$a")))},
		{"No student is red", v24Q("every", "$a", fol("implies", foAtom("is/student", "$a"), fol("not", foAtom("is/red", "$a"))))},
		{"Every student admires blair", v24Q("every", "$a", fol("implies", foAtom("is/student", "$a"), foAtom("admires", "$a", "blair")))},
		{"Alice admires some teacher", v24Q("some", "$a", fol("and", foAtom("is/teacher", "$a"), foAtom("admires", "alice", "$a")))},
		{"Some teacher is such that every student admires that teacher", v24Q("some", "$a", fol("and", foAtom("is/teacher", "$a"), v24Q("every", "$b", fol("implies", foAtom("is/student", "$b"), foAtom("admires", "$b", "$a")))))},
		{"Alice is a student", foAtom("is/student", "alice")},
		{"Alice is not red", fol("not", foAtom("is/red", "alice"))},
		{"Alice sends blair to casey", foAtom("sends/to", "alice", "blair", "casey")},
		{"Every student that alice admires is red", v24Q("every", "$a", fol("implies", fol("and", foAtom("is/student", "$a"), foAtom("admires", "alice", "$a")), foAtom("is/red", "$a")))},
		{"Alice gave blair a book", v24Q("some", "$a", fol("and", foAtom("is/book", "$a"), foAtom("gave/to", "alice", "$a", "blair")))},
	}
	for _, tc := range cases {
		t.Run(tc.s, func(t *testing.T) {
			fs, why := legacyFOLReadings(tc.s)
			if why != "" || len(fs) != 1 || !folSame(fs[0], tc.f) {
				t.Fatalf("translation: %s, %d readings", why, len(fs))
			}
			v25Conversion(t, tc.s, tc.f.Render())
			// Independently specified formula versus the parsed rendering, both ways.
			f, err := parseFOL(tc.f.Render())
			if err != nil {
				t.Fatal(err)
			}
			if finiteCountermodel([]*folFormula{fs[0]}, f, 3) || finiteCountermodel([]*folFormula{f}, fs[0], 3) {
				t.Fatal("countermodel")
			}
		})
	}
	for _, s := range []string{"Every student admires some teacher", "Every student admires every teacher"} {
		fs, why := legacyFOLReadings(s)
		if why != "" || len(fs) < 2 {
			t.Fatal(s, why, len(fs))
		}
		for _, f := range fs {
			for _, pair := range [][2]string{{s, f.Render()}, {f.Render(), s}} {
				_, r := CheckEntailmentGate("1. " + pair[0] + ".\n2. Therefore, " + pair[1] + ". [1]")
				if r.Passed || !strings.Contains(r.Message, "use is such that") {
					t.Fatal(s, r.Message)
				}
			}
		}
	}
	for _, tc := range []struct {
		p, c string
		refs []int
	}{
		{"Alice admires blair", "Blair admires alice", []int{1}},
		{"Alice is red", "Alice is blue", []int{1}},
		{"Alice is red", "Alice is red", []int{1, 2}},
		{"Alice is red", "Alice is red", nil},
	} {
		r := notationEquivalence(NumberedArgumentStep{Number: 3, Statement: tc.c, Citations: tc.refs}, map[int]NumberedArgumentStep{1: {Statement: tc.p}, 2: {Statement: tc.p}})
		if r.Passed {
			t.Fatal(tc)
		}
	}
}

type v25LegacySentence struct {
	Sentence  string   `json:"sentence"`
	Forms     []string `json:"forms"`
	Exception string   `json:"exception,omitempty"`
}

func TestCatalogV25MixedScopeRejection(t *testing.T) {
	readable := "Every student is red and alice is blue"
	narrow := fol("and", v24Q("every", "$x", fol("implies", foAtom("is/student", "$x"), foAtom("is/red", "$x"))), foAtom("is/blue", "alice"))
	wide := v24Q("every", "$x", fol("implies", foAtom("is/student", "$x"), fol("and", foAtom("is/red", "$x"), foAtom("is/blue", "alice"))))
	if !finiteCountermodel([]*folFormula{wide}, narrow, 3) {
		t.Fatal("expected empty-class scope countermodel")
	}
	for _, f := range []*folFormula{wide, narrow} {
		for _, pair := range [][2]string{{readable, f.Render()}, {f.Render(), readable}} {
			_, r := CheckEntailmentGate("1. " + pair[0] + ".\n2. Therefore, " + pair[1] + ". [1]")
			if r.Passed || !strings.Contains(r.Message, "use is such that") {
				t.Fatal(r.Message)
			}
		}
	}
}

func TestCatalogV25SymbolBoundaries(t *testing.T) {
	// Legacy opaque text must not spoof the serialization of a structured atom.
	a := &folFormula{op: "atom", symbol: "is/red/alice"}
	b := foAtom("is/red", "alice")
	if folSame(a, b) {
		t.Fatal("symbol/term boundary collision")
	}
	steps := map[int]NumberedArgumentStep{1: {Statement: "is/red/alice"}}
	if notationEquivalence(NumberedArgumentStep{Number: 2, Statement: "alice is red", Citations: []int{1}}, steps).Passed {
		t.Fatal("opaque key spoof")
	}
	fs, why := legacyFOLReadings("Every student admires $slot0")
	if why != "" || len(fs) != 1 {
		t.Fatal(why, len(fs))
	}
	expected := v24Q("every", "$x", fol("implies", foAtom("is/student", "$x"), foAtom("admires", "$x", legacyFOLTerm("$slot0"))))
	if !folSame(fs[0], expected) {
		t.Fatal("surface name captured as internal variable")
	}
	if _, why := legacyFOLRender(fs[0]); why == "" {
		t.Fatal("non-core name rendered as bound variable")
	}
}

func TestCatalogV25LegacyCoverage(t *testing.T) {
	data, err := os.ReadFile("testdata/legacy_v24_sentences.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []v25LegacySentence
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 6386 {
		t.Fatalf("frozen v24 inventory size changed: %d", len(rows))
	}
	converted, ambiguous, exceptions, sampled := 0, 0, 0, 0
	forms := map[string]bool{}
	reasons := map[string]int{}
	for i, row := range rows {
		fs, why := legacyFOLReadings(row.Sentence)
		if len(fs) > 1 || strings.Contains(why, "ambiguous") {
			why = "multiple non-alpha-equivalent readings"
			ambiguous++
		} else if why == "" && len(fs) == 1 {
			_, why = legacyFOLRender(fs[0])
		}
		if os.Getenv("ARGCHECK_UPDATE_LEGACY") == "1" {
			rows[i].Exception = why
			row.Exception = why
		}
		if why != "" {
			if row.Exception != why {
				t.Errorf("exception changed for %q: %q -> %q", row.Sentence, row.Exception, why)
			}
			reasons[why]++
			exceptions++
			continue
		}
		if len(fs) != 1 {
			t.Fatalf("no reading: %s", row.Sentence)
		}
		if row.Exception != "" {
			t.Errorf("resolved exception must be removed: %s", row.Sentence)
		}
		rendered, _ := legacyFOLRender(fs[0])
		v25Conversion(t, row.Sentence, rendered)
		converted++
		for _, form := range row.Forms {
			forms[form] = true
		}
		// Bounded sample excludes very deep explicit formulas captured from v24's
		// inventory: grounding those is exponential in quantifier depth.
		if converted%20 == 0 && !folExplicit(row.Sentence) && strings.Count(rendered, "is such that") <= 3 {
			f, err := parseFOL(rendered)
			if err != nil {
				t.Fatal(err)
			}
			if finiteCountermodel([]*folFormula{fs[0]}, f, 3) || finiteCountermodel([]*folFormula{f}, fs[0], 3) {
				t.Fatal("conversion oracle", row.Sentence)
			}
			sampled++
		}
	}
	if os.Getenv("ARGCHECK_UPDATE_LEGACY") == "1" {
		data, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile("testdata/legacy_v24_sentences.json", append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var names []string
	for f := range forms {
		names = append(names, f)
	}
	sort.Strings(names)
	t.Logf("%d distinct sentences; %d single-reading conversions in both directions; %d ambiguous; %d named exceptions (including ambiguous); %d oracle samples; zero conversion failures", len(rows), converted, ambiguous, exceptions, sampled)
	t.Logf("source forms: %s", strings.Join(names, ", "))
	t.Logf("exception reasons: %v", reasons)
}
