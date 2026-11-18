package validation

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func v24Random(r *rand.Rand, depth int, scope []string, budget *int) *folFormula {
	if depth > 0 {
		switch r.Intn(9) {
		case 0, 1:
			if *budget > 0 {
				v := "$" + string("xyzuvwabcdefghijklmnopqrst"[len(scope)])
				*budget--
				kind := "every"
				if r.Intn(2) == 0 {
					kind = "some"
				}
				q := fol(kind, v24Random(r, depth-1, append(append([]string{}, scope...), v), budget))
				q.symbol = v
				return q
			}
		case 2:
			return fol("not", v24Random(r, depth-1, scope, budget))
		case 3, 4, 5, 6:
			op := []string{"and", "or", "implies", "iff"}[r.Intn(4)]
			return fol(op, v24Random(r, depth-1, scope, budget), v24Random(r, depth-1, scope, budget))
		}
	}
	term := func() string {
		if len(scope) > 0 && r.Intn(3) != 0 {
			return scope[r.Intn(len(scope))]
		}
		return []string{"alice", "blair"}[r.Intn(2)]
	}
	if r.Intn(4) == 0 {
		return &folFormula{op: "equal", terms: []string{term(), term()}}
	}
	n := 1 + r.Intn(4)
	symbol := []string{"is/red", "admires", "sends/to", "sends/to/from"}[n-1]
	ts := []string{}
	for i := 0; i < n; i++ {
		ts = append(ts, term())
	}
	return &folFormula{op: "atom", symbol: symbol, terms: ts}
}
func v24Formula(r *rand.Rand) *folFormula { budget := 6; return v24Random(r, 5, nil, &budget) }
func TestCatalogV24RoundTrip(t *testing.T) {
	start := time.Now()
	r := rand.New(rand.NewSource(24001))
	for i := 0; i < 2000; i++ {
		f := v24Formula(r)
		text := f.Render()
		got, e := parseFOL(text)
		if e != nil || !folSame(f, got) {
			t.Fatalf("%d: %s: %v\nwant %s\ngot %s", i, text, e, folKey(f, nil, 0), folKey(got, nil, 0))
		}
	}
	t.Logf("2000/2000 (100%%), depth <=5, quantifier budget 6; elapsed %s", time.Since(start))
}
func v24Check(t *testing.T, argument string) {
	t.Helper()
	_, result := CheckEntailmentGate(argument)
	if !result.Passed {
		t.Fatalf("%s\n%s", argument, result.Message)
	}
}
func v24Flat(ps []*folFormula, c *folFormula) string {
	var b strings.Builder
	var refs []string
	for i, p := range ps {
		fmt.Fprintf(&b, "%d. %s.\n", i+1, p.Render())
		refs = append(refs, fmt.Sprint(i+1))
	}
	fmt.Fprintf(&b, "%d. Therefore, %s. [%s]", len(ps)+1, c.Render(), strings.Join(refs, ","))
	return b.String()
}
func v24Q(kind, v string, body *folFormula) *folFormula { q := fol(kind, body); q.symbol = v; return q }
func TestCatalogV24RuleInventory(t *testing.T) {
	start := time.Now()
	oracleTime := time.Duration(0)
	r := rand.New(rand.NewSource(24002))
	names := []string{"and-I", "and-E", "or-I", "or-E", "implies-I", "implies-E", "not-I", "not-E", "bottom-E", "double-negation", "iff-I", "iff-E", "forall-I", "forall-E", "exists-I", "exists-E", "identity-I", "identity-E"}
	for _, rule := range names {
		t.Run(rule, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				a, b := v24Formula(r), v24Formula(r)
				var ps []*folFormula
				var c *folFormula
				argument := ""
				// For Fitch obligations the oracle checks the discharged sequent, rather
				// than treating an eigenname as a fixed constant in universal introduction.
				switch rule {
				case "and-I":
					ps = []*folFormula{a, b}
					c = fol("and", a, b)
				case "and-E":
					ps = []*folFormula{fol("and", a, b)}
					c = a
				case "or-I":
					ps = []*folFormula{a}
					c = fol("or", a, b)
				case "implies-E":
					ps = []*folFormula{fol("implies", a, b), a}
					c = b
				case "not-E":
					ps = []*folFormula{a, fol("not", a)}
					c = fol("bottom")
				case "bottom-E":
					ps = []*folFormula{fol("bottom")}
					c = a
				case "double-negation":
					ps = []*folFormula{fol("not", fol("not", a))}
					c = a
				case "iff-I":
					ps = []*folFormula{fol("implies", a, b), fol("implies", b, a)}
					c = fol("iff", a, b)
				case "iff-E":
					ps = []*folFormula{fol("iff", a, b)}
					c = fol("implies", a, b)
				case "forall-E", "exists-I", "exists-E", "forall-I":
					budget := 4
					body := v24Random(r, 3, []string{"$x"}, &budget)
					// Ensure a nonvacuous binder and exercise repeated variable occurrences.
					body = fol("and", body, &folFormula{op: "atom", symbol: "admires", terms: []string{"$x", "$x"}})
					instance := folSub(body, "$x", "witness")
					ex := v24Q("some", "$x", body)
					switch rule {
					case "forall-E":
						ps = []*folFormula{v24Q("every", "$x", body)}
						c = instance
					case "exists-I":
						ps = []*folFormula{instance}
						c = ex
					case "exists-E":
						bridge := v24Q("every", "$x", fol("implies", body, b))
						ps = []*folFormula{ex, bridge}
						c = b
						argument = fmt.Sprintf("1. %s.\n2. %s.\n| 3. Suppose %s.\n| 4. Therefore, %s. [2]\n| 5. Therefore, %s. [3,4]\n6. Therefore, %s. [1,3-5]", ex.Render(), bridge.Render(), instance.Render(), fol("implies", instance, b).Render(), b.Render(), b.Render())
					case "forall-I":
						c = v24Q("every", "$x", body)
						ps = []*folFormula{c}
						argument = fmt.Sprintf("1. %s.\n| 2. Let witness be arbitrary.\n| 3. Therefore, %s. [1]\n4. Therefore, %s. [2-3]", c.Render(), instance.Render(), c.Render())

					}
				case "identity-I":
					a = fol("and", a, foAtom("is/red", "alice"))
					ps = []*folFormula{a}
					c = &folFormula{op: "equal", terms: []string{"alice", "alice"}}
				case "identity-E":
					// v35.0: substitution is certified in clause subjects, so
					// the generated instance replaces a subject.
					// Fresh names keep every replaced occurrence a subject.
					ps = []*folFormula{&folFormula{op: "equal", terms: []string{"casey", "dana"}}, fol("and", a, foAtom("is/red", "casey"))}
					c = fol("and", a, foAtom("is/red", "dana"))
				case "implies-I":
					c = fol("implies", a, a)
					argument = fmt.Sprintf("| 1. Suppose %s.\n| 2. Therefore, %s. [1]\n3. Therefore, %s. [1-2]", a.Render(), a.Render(), c.Render())
				case "not-I":
					ps = []*folFormula{fol("not", a)}
					c = fol("not", a)
					argument = fmt.Sprintf("1. %s.\n| 2. Suppose %s.\n| 3. Therefore, contradiction. [1,2]\n4. Therefore, %s. [2-3]", c.Render(), a.Render(), c.Render())
				case "or-E":
					disj := fol("or", a, b)
					c = v24Formula(r)
					left, right := fol("implies", a, c), fol("implies", b, c)
					ps = []*folFormula{disj, left, right}
					argument = fmt.Sprintf("1. %s.\n2. %s.\n3. %s.\n| 4. Suppose %s.\n| 5. Therefore, %s. [2,4]\n| 6. Suppose %s.\n| 7. Therefore, %s. [3,6]\n8. Therefore, %s. [1,4-5,6-7]", disj.Render(), left.Render(), right.Render(), a.Render(), c.Render(), b.Render(), c.Render(), c.Render())
				}
				if argument == "" {
					argument = v24Flat(ps, c)
				}
				v24Check(t, argument)
				// Deterministic 10/50 sample per rule, all models of domains 1–3.
				if i%5 == 0 {
					at := time.Now()
					if finiteCountermodel(ps, c, 3) {
						t.Fatalf("oracle countermodel: %s", argument)
					}
					v24LocalOracle(t, argument)
					oracleTime += time.Since(at)
				}
			}
			t.Log("50/50 accepted; 10/50 oracle checked over domains 1–3")
		})
	}
	t.Logf("900 rule instances; oracle elapsed %s; total elapsed %s", oracleTime, time.Since(start))
}

type v24Exercise struct {
	name  string
	p     []string
	c     string
	valid bool
}

func v24Exercises() []v24Exercise {
	return []v24Exercise{
		{"transitivity instantiation", []string{"every x is such that every y is such that every z is such that if x outranks y and y outranks z then x outranks z"}, "every y is such that every z is such that if alice outranks y and y outranks z then alice outranks z", true},
		{"forall exists forall", []string{"every x is such that some y is such that every z is such that x sends y to z"}, "some y is such that every z is such that alice sends y to z", true},
		{"exists forall exists", []string{"every y is such that some z is such that alice sends y to z"}, "some x is such that every y is such that some z is such that x sends y to z", true},
		{"repeated slot", []string{"every x is such that x admires x"}, "alice admires alice", true},
		{"identity body", []string{"every x is such that x is identical to x"}, "alice is identical to alice", true},
		{"negative body", []string{"no x is such that x admires blair"}, "it is not the case that (alice admires blair)", true},
		{"conjunction body", []string{"every x is such that (x is red) and (x admires blair)"}, "(alice is red) and (alice admires blair)", true},
		{"identity substitution", []string{"alice is identical to blair", "every x is such that alice admires x"}, "every x is such that blair admires x", true},
		{"four slots", []string{"every x is such that x sends x to x from x"}, "alice sends alice to alice from alice", true},
		{"deep alternation", []string{"every x is such that some y is such that every z is such that some u is such that every v is such that some w is such that every a is such that some b is such that x sends y to a from b"}, "some y is such that every z is such that some u is such that every v is such that some w is such that every a is such that some b is such that alice sends y to a from b", true},
		{"double object", []string{"some y is such that every x is such that x gave y a book"}, "some y is such that every x is such that x gave y a book", true},
		{"scope swap", []string{"every x is such that some y is such that x admires y"}, "some y is such that every x is such that x admires y", false},
		{"role swap", []string{"every x is such that x admires blair"}, "every x is such that blair admires x", false},
		{"existential witness escape", []string{"some x is such that x is red"}, "alice is red", false},
		{"universal strengthening", []string{"alice admires blair"}, "every x is such that x admires blair", false},
		{"negation flip", []string{"no x is such that x admires blair"}, "every x is such that x admires blair", false},
	}
}
func TestCatalogV24Exercises(t *testing.T) {
	start := time.Now()
	valid, invalid := 0, 0
	for _, tc := range v24Exercises() {
		t.Run(tc.name, func(t *testing.T) {
			ps := []*folFormula{}
			for _, s := range tc.p {
				p, e := parseFOL(s)
				if e != nil {
					t.Fatal(e)
				}
				ps = append(ps, p)
			}
			c, e := parseFOL(tc.c)
			if e != nil {
				t.Fatal(e)
			}
			var b strings.Builder
			refs := []string{}
			for i, s := range tc.p {
				fmt.Fprintf(&b, "%d. %s.\n", i+1, s)
				refs = append(refs, fmt.Sprint(i+1))
			}
			fmt.Fprintf(&b, "%d. Therefore, %s. [%s]", len(tc.p)+1, tc.c, strings.Join(refs, ","))
			_, result := CheckEntailmentGate(b.String())
			if result.Passed != tc.valid {
				t.Fatal(result.Message)
			}
			if finiteCountermodel(ps, c, 3) == tc.valid {
				t.Fatal("oracle disagrees")
			}
			if tc.valid {
				valid++
			} else {
				invalid++
			}
		})
	}
	t.Logf("%d valid, %d invalid; domains 1–3; oracle plus gate elapsed %s", valid, invalid, time.Since(start))
}
func TestCatalogV24Malformed(t *testing.T) {
	for _, s := range []string{"every x is such that y admires x", "every x is such that every x is such that x is red", "every x is such that (x is red) and (z is red)", "every x is such that x would admire alice"} {
		if _, e := parseFOL(s); e == nil {
			t.Fatal(s)
		}
		_, r := CheckEntailmentGate("1. " + s + ".\n2. Therefore, " + s + ". [1]")
		if r.Passed || !strings.Contains(r.Message, "nearest form") {
			t.Fatal(r.Message)
		}
	}
}

func TestCatalogV24ScopeAndIdentityGuards(t *testing.T) {
	invalid := []string{
		"1. every x is such that x is red.\n2. Therefore, x is red. [1]",
		"1. every x is such that (x is red) and (x is blue).\n2. Therefore, x is blue. [1]",
		"1. every object is red.\n| 2. Let witness be an arbitrary object.\n| 3. Therefore, witness is red. [1,2]\n4. Therefore, every x is such that x is red. [2-3]",
		"1. witness is red.\n| 2. Let witness be arbitrary.\n| 3. Therefore, witness is red. [1]\n4. Therefore, every x is such that x is red. [2-3]",
		"1. some x is such that x is red.\n| 2. Suppose witness is red.\n| 3. Therefore, witness is red. [2]\n4. Therefore, witness is red. [1,2-3]",
		"1. some x is such that x is red.\n| 2. Suppose (witness is red) and (witness is blue).\n| 3. Therefore, some x is such that x is blue. [2]\n4. Therefore, some x is such that x is blue. [1,2-3]",
	}
	for _, s := range invalid {
		_, r := CheckEntailmentGate(s)
		if r.Passed {
			t.Fatal(s)
		}
	}
	// Selective replacement, with the same name used twice in the atom.
	ps := []*folFormula{{op: "equal", terms: []string{"alice", "blair"}}, v24Q("every", "$x", &folFormula{op: "atom", symbol: "sends/to", terms: []string{"$x", "alice", "alice"}})}
	c := v24Q("every", "$y", &folFormula{op: "atom", symbol: "sends/to", terms: []string{"$y", "blair", "alice"}})
	// v35.0: the replaced name is an object, which the safe-position rule
	// refuses; the formula is still valid, so no countermodel exists.
	if _, r := CheckEntailmentGate(v24Flat(ps, c)); r.Passed {
		t.Fatal("object-position identity substitution certified")
	}
	if finiteCountermodel(ps, c, 3) {
		t.Fatal("selective substitution")
	}
}

func TestCatalogV24IndependentTranslations(t *testing.T) {
	// Hand-built meanings pin scope, repeated roles, and the article restrictor;
	// these do not obtain their semantics by parsing rendered source formulas.
	r := func(a, b string) *folFormula { return foAtom("outranks", a, b) }
	trans := v24Q("every", "$x", v24Q("every", "$y", v24Q("every", "$z", fol("implies", fol("and", r("$x", "$y"), r("$y", "$z")), r("$x", "$z")))))
	got, e := parseFOL(v24Exercises()[0].p[0])
	if e != nil || !folSame(got, trans) {
		t.Fatal("transitivity translation", e)
	}
	book := v24Q("some", "$y", v24Q("every", "$x", v24Q("some", "$b", fol("and", foAtom("is/book", "$b"), foAtom("gave/to", "$x", "$b", "$y")))))
	got, e = parseFOL("some y is such that every x is such that x gave y a book")
	if e != nil || !folSame(got, book) {
		t.Fatal("article translation", e)
	}
	// Pin at least six and eight nested quantifiers, independent of the seed.
	for _, n := range []int{6, 8} {
		f := foAtom("admires", "$x", "$y")
		for i := n - 1; i >= 0; i-- {
			kind := "every"
			if i%2 == 1 {
				kind = "some"
			}
			f = v24Q(kind, "$"+string("xyzuvwab"[i]), f)
		}
		got, e := parseFOL(f.Render())
		if e != nil || !folSame(f, got) {
			t.Fatal(n, e)
		}
	}
	// Complete the three universal eliminations and use the transitivity rule.
	argument := `1. every x is such that every y is such that every z is such that if x outranks y and y outranks z then x outranks z.
2. Therefore, every y is such that every z is such that if alice outranks y and y outranks z then alice outranks z. [1]
3. Therefore, every z is such that if alice outranks blair and blair outranks z then alice outranks z. [2]
4. Therefore, if (alice outranks blair) and (blair outranks casey) then (alice outranks casey). [3]
5. alice outranks blair.
6. blair outranks casey.
7. Therefore, (alice outranks blair) and (blair outranks casey). [5,6]
8. Therefore, alice outranks casey. [4,7]`
	v24Check(t, argument)
	v24LocalOracle(t, argument)
	if finiteCountermodel([]*folFormula{trans, r("alice", "blair"), r("blair", "casey")}, r("alice", "casey"), 3) {
		t.Fatal("transitivity oracle")
	}
}

// Check every local rule obligation. A discharged branch supplies its
// implication; eigenname obligations are universally closed, never asserted
// for one arbitrary fixed interpretation of that name.
func v24LocalOracle(t *testing.T, argument string) {
	t.Helper()
	arg, result := ParseNumberedArgument(argument, numberedArgumentOptions{})
	if !result.Passed {
		t.Fatal(result.Message)
	}
	steps := map[int]NumberedArgumentStep{}
	fs := map[int]*folFormula{}
	for _, s := range arg.Steps {
		steps[s.Number] = s
		f, e := parseFOL(fitchAssumptionStatement(s))
		if e != nil {
			t.Fatal(e, s.Statement)
		}
		fs[s.Number] = f
	}
	for _, s := range arg.Steps {
		if s.Kind != numberedArgumentStepConclusion {
			continue
		}
		ps := []*folFormula{}
		for _, id := range s.Citations {
			ps = append(ps, fs[id])
		}
		for _, span := range s.Ranges {
			first, last := steps[span[0]], fs[span[1]]
			start := fs[span[0]]
			if first.Kind == numberedArgumentStepArbitrary {
				m := arbitraryFOLPattern.FindStringSubmatch(first.Statement)
				if m == nil {
					t.Fatal("oracle unrestricted declaration expected")
				}
				ps = append(ps, v24Q("every", "$eigen", folSub(last, strings.ToLower(m[1]), "$eigen")))
				continue
			}
			branch := fol("implies", start, last)
			if len(s.Citations) == 1 && len(s.Ranges) == 1 && ps[0].op == "some" {
				name := folInstance(ps[0], start)
				if folSame(ps[0].children[0], start) {
					name = "unusedwitness"
				}
				branch = v24Q("every", "$eigen", folSub(branch, name, "$eigen"))
			}
			ps = append(ps, branch)
		}
		if finiteCountermodel(ps, fs[s.Number], 3) {
			t.Fatalf("local oracle step %d: %s", s.Number, argument)
		}
	}
}

func TestCatalogV24VacuousWitness(t *testing.T) {
	argument := `1. some x is such that alice is red.
| 2. Suppose alice is red.
| 3. Therefore, alice is red. [2]
4. Therefore, alice is red. [1,2-3]`
	v24Check(t, argument)
	v24LocalOracle(t, argument)
	argument = `| 1. Let witness be arbitrary.
| 2. Therefore, witness is identical to witness. [1]
3. Therefore, every x is such that x is identical to x. [1-2]
4. Therefore, alice is identical to alice. [3]`
	v24Check(t, argument)
	v24LocalOracle(t, argument)
}
