package validation

import (
	"fmt"
	"strings"
	"testing"
)

// Independent full second-order semantics. The oracle owns both the AST and
// renderer; it never uses gate parsing or matching to decide truth. Names may
// co-refer. Every predicate extension and every assignment of three names is
// enumerated on domains 1..3. Property quantifiers range over ALL subsets,
// including the empty set and singleton identity properties.
type propertyOracleModel struct {
	n          int
	names      [3]int
	extensions [3]int
}

type propertyOracleFormula struct {
	op              string
	a, b, predicate int
	left, right     *propertyOracleFormula
}

func poAtom(op string, a, b, predicate int) *propertyOracleFormula {
	return &propertyOracleFormula{op: op, a: a, b: b, predicate: predicate}
}
func poNot(f *propertyOracleFormula) *propertyOracleFormula {
	return &propertyOracleFormula{op: "not", left: f}
}
func poBin(op string, a, b *propertyOracleFormula) *propertyOracleFormula {
	return &propertyOracleFormula{op: op, left: a, right: b}
}

func (f *propertyOracleFormula) eval(m propertyOracleModel, eigenproperty int) bool {
	member := func(name, extension int) bool { return extension&(1<<m.names[name]) != 0 }
	switch f.op {
	case "predicate", "possession":
		return member(f.a, m.extensions[f.predicate])
	case "identity":
		return m.names[f.a] == m.names[f.b]
	case "identity possession":
		return member(f.a, 1<<m.names[f.b])
	case "variable possession":
		return member(f.a, eigenproperty)
	case "inclusion":
		for p := 0; p < 1<<m.n; p++ {
			if member(f.b, p) && !member(f.a, p) {
				return false
			}
		}
		return true
	case "exists":
		for p := 0; p < 1<<m.n; p++ {
			if member(f.a, p) {
				return true
			}
		}
		return false
	case "distinguishing":
		for p := 0; p < 1<<m.n; p++ {
			if member(f.a, p) && !member(f.b, p) {
				return true
			}
		}
		return false
	case "shared":
		for p := 0; p < 1<<m.n; p++ {
			if member(f.a, p) && member(f.b, p) {
				return true
			}
		}
		return false
	case "forall property":
		for p := 0; p < 1<<m.n; p++ {
			if !f.left.eval(m, p) {
				return false
			}
		}
		return true
	case "not":
		return !f.left.eval(m, eigenproperty)
	case "and":
		return f.left.eval(m, eigenproperty) && f.right.eval(m, eigenproperty)
	case "or":
		return f.left.eval(m, eigenproperty) || f.right.eval(m, eigenproperty)
	case "if":
		return !f.left.eval(m, eigenproperty) || f.right.eval(m, eigenproperty)
	case "iff":
		return f.left.eval(m, eigenproperty) == f.right.eval(m, eigenproperty)
	}
	panic(f.op)
}

func (f *propertyOracleFormula) english() string {
	names := [3]string{"Socrates", "Plato", "Aristotle"}
	predicates := [3]string{"wise", "just", "calm"}
	a, b := names[f.a], names[f.b]
	switch f.op {
	case "predicate":
		return a + " is " + predicates[f.predicate]
	case "possession":
		return a + " has the property of being " + predicates[f.predicate]
	case "identity":
		return a + " is identical to " + b
	case "identity possession":
		return a + " has the property of being identical to " + b
	case "variable possession":
		return a + " has P"
	case "inclusion":
		return a + " has every property that " + b + " has"
	case "exists":
		return a + " has some property"
	case "distinguishing":
		return a + " has some property that " + b + " lacks"
	case "shared":
		return a + " shares some property with " + b
	case "not":
		return "It is not the case that " + f.left.english()
	case "and":
		return "It is the case both that " + f.left.english() + ", and that " + f.right.english()
	case "or":
		return "Either " + f.left.english() + " or " + f.right.english()
	case "if":
		return "If " + f.left.english() + ", then " + f.right.english()
	case "iff":
		return "(" + f.left.english() + ") if and only if (" + f.right.english() + ")"
	}
	panic(f.op)
}

type propertyOracleCandidate struct {
	name          string
	premises      []*propertyOracleFormula
	conclusion    *propertyOracleFormula
	argument      string
	wantForm      string
	mustAccept    bool
	targetStep    int  // zero means the final conclusion
	allProperties bool // enumerate every eigenproperty extension
}

func (c propertyOracleCandidate) text() string {
	if c.argument != "" {
		return c.argument
	}
	var out strings.Builder
	refs := make([]string, len(c.premises))
	for i, p := range c.premises {
		fmt.Fprintf(&out, "%d. %s.\n", i+1, p.english())
		refs[i] = fmt.Sprint(i + 1)
	}
	fmt.Fprintf(&out, "%d. Therefore, %s. [%s]", len(c.premises)+1, c.conclusion.english(), strings.Join(refs, ","))
	return out.String()
}

func propertyOracleCandidates() []propertyOracleCandidate {
	var result []propertyOracleCandidate
	add := func(form string, c *propertyOracleFormula, ps ...*propertyOracleFormula) {
		result = append(result, propertyOracleCandidate{name: form, premises: ps, conclusion: c, wantForm: form, mustAccept: true})
	}
	for a := 0; a < 3; a++ {
		for b := 0; b < 3; b++ {
			for f := 0; f < 3; f++ {
				pred, prop, other := poAtom("predicate", a, 0, f), poAtom("possession", a, 0, f), poAtom("possession", b, 0, f)
				add("property abstraction", prop, pred)
				add("property concretion", pred, prop)
				add("property abstraction", poNot(prop), poNot(pred))
				add("property concretion", poNot(pred), poNot(prop))
				// Degenerate repeated premises may receive an older form (reiteration).
				if a != b {
					add("property instantiation", prop, poAtom("inclusion", a, b, 0), other)
					add("shared property", poAtom("shared", a, b, 0), prop, other)
					add("indiscernibility of identicals", poAtom("inclusion", a, b, 0), poAtom("identity", a, b, 0))
				}
				difference := poAtom("distinguishing", a, b, 0)
				add("distinguishing property", difference, prop, poNot(other))
				add("distinguishing property nonidentity", poNot(poAtom("identity", a, b, 0)), difference)
				add("distinguishing property non-inclusion", poNot(poAtom("inclusion", b, a, 0)), difference)
				if a != b {
					for _, bad := range []propertyOracleCandidate{
						{name: "distinguishing mismatched F", premises: []*propertyOracleFormula{prop, poNot(poAtom("possession", b, 0, (f+1)%3))}, conclusion: difference},
						{name: "distinguishing wrong bearer", premises: []*propertyOracleFormula{prop, poNot(other)}, conclusion: poAtom("distinguishing", a, 3-a-b, 0)},
						{name: "distinguishing positive identity", premises: []*propertyOracleFormula{difference}, conclusion: poAtom("identity", a, b, 0)},
						{name: "distinguishing positive inclusion", premises: []*propertyOracleFormula{difference}, conclusion: poAtom("inclusion", b, a, 0)},
					} {
						result = append(result, bad)
					}
				}
				add("property existential", poAtom("exists", a, 0, 0), prop)
				for _, bad := range []*propertyOracleFormula{poNot(prop), poAtom("possession", a, 0, (f+1)%3)} {
					result = append(result, propertyOracleCandidate{name: "invalid abstraction mutation", premises: []*propertyOracleFormula{pred}, conclusion: bad})
				}
			}
		}
	}
	// Arbitrary properties are universally quantified in the proof condition,
	// never interpreted as one lucky named extension.
	vpA, vpB := poAtom("variable possession", 0, 0, 0), poAtom("variable possession", 1, 0, 0)
	universal := &propertyOracleFormula{op: "forall property", left: poBin("if", vpB, vpA)}
	result = append(result, propertyOracleCandidate{name: "UG semantic discharge", premises: []*propertyOracleFormula{universal}, conclusion: poAtom("inclusion", 0, 1, 0), argument: v20UG, wantForm: "property generalization", mustAccept: true})
	result = append(result, propertyOracleCandidate{name: "UG full derivation", premises: []*propertyOracleFormula{poAtom("identity", 0, 1, 0)}, conclusion: poAtom("inclusion", 0, 1, 0), argument: v20UG, wantForm: "property generalization", mustAccept: true})
	result = append(result,
		propertyOracleCandidate{name: "property transitivity", premises: []*propertyOracleFormula{poAtom("inclusion", 1, 0, 0), poAtom("inclusion", 2, 1, 0)}, conclusion: poAtom("inclusion", 2, 0, 0), argument: v21Trans, wantForm: "property generalization", mustAccept: true},
		propertyOracleCandidate{name: "arbitrary instantiation assumption", premises: []*propertyOracleFormula{poAtom("inclusion", 1, 0, 0), vpA}, conclusion: vpB, argument: v21Trans, targetStep: 4, allProperties: true, wantForm: "property instantiation (arbitrary property)", mustAccept: true},
		propertyOracleCandidate{name: "arbitrary instantiation derived", premises: []*propertyOracleFormula{poAtom("inclusion", 2, 1, 0), vpB}, conclusion: poAtom("variable possession", 2, 0, 0), argument: v21Trans, targetStep: 5, allProperties: true, wantForm: "property instantiation (arbitrary property)", mustAccept: true},
		propertyOracleCandidate{name: "arbitrary distinguishing", premises: []*propertyOracleFormula{vpA, poNot(vpB)}, conclusion: poAtom("distinguishing", 0, 1, 0), argument: v21ArbitraryDiffer, targetStep: 3, allProperties: true, wantForm: "distinguishing property", mustAccept: true},
		propertyOracleCandidate{name: "arbitrary wrong bearer", premises: []*propertyOracleFormula{poAtom("inclusion", 1, 2, 0), vpA}, conclusion: vpB, argument: strings.Replace(v21Trans, "Plato has every property that Socrates has", "Plato has every property that Aristotle has", 1), targetStep: 4, allProperties: true},
		propertyOracleCandidate{name: "arbitrary distinguishing polarity", premises: []*propertyOracleFormula{vpA, vpB}, conclusion: poAtom("distinguishing", 0, 1, 0), argument: strings.ReplaceAll(v21ArbitraryDiffer, "It is not the case that Plato has P", "Plato has P"), targetStep: 3, allProperties: true},
	)
	// A free property condition does NOT suffice for generalization.
	result = append(result, propertyOracleCandidate{name: "invalid free property generalization", premises: []*propertyOracleFormula{poBin("if", poAtom("possession", 1, 0, 0), poAtom("possession", 0, 0, 0))}, conclusion: poAtom("inclusion", 0, 1, 0)})
	result = append(result, propertyOracleCandidate{name: "identity of indiscernibles", premises: []*propertyOracleFormula{poAtom("inclusion", 0, 1, 0)}, conclusion: poAtom("identity", 0, 1, 0), argument: v20IdentityOfIndiscernibles, wantForm: "property concretion", mustAccept: true})
	// Identity properties are singleton extensions, not independent predicates.
	add("property abstraction", poAtom("identity possession", 1, 1, 0), poAtom("identity", 1, 1, 0))
	add("property instantiation", poAtom("identity possession", 0, 1, 0), poAtom("inclusion", 0, 1, 0), poAtom("identity possession", 1, 1, 0))
	add("property concretion", poAtom("identity", 0, 1, 0), poAtom("identity possession", 0, 1, 0))
	// Descartes uses the same logical shape with arbitrary names and F.
	identity, inclusion := poAtom("identity", 0, 1, 0), poAtom("inclusion", 0, 1, 0)
	result = append(result, propertyOracleCandidate{name: "Descartes reductio", premises: []*propertyOracleFormula{poBin("if", identity, inclusion), poAtom("possession", 1, 0, 0), poNot(poAtom("possession", 0, 0, 0))}, conclusion: poNot(identity), argument: v20Descartes, wantForm: "modus tollens", mustAccept: true})
	return result
}

// Classical connective templates are independent of the production catalog.
// Each is tried with all eight mixtures of ordinary and property predicates.
// Explicit abstraction/concretion lines are added before or after every
// atomic endpoint. This tests real composed proofs, not implicit rewriting.
func propertyOracleCompositions() []propertyOracleCandidate {
	var result []propertyOracleCandidate
	for mask := 0; mask < 8; mask++ {
		atoms := make([]*propertyOracleFormula, 3)
		for i := range atoms {
			op := "predicate"
			if mask&(1<<i) != 0 {
				op = "possession"
			}
			atoms[i] = poAtom(op, i, 0, i)
		}
		p, q, r := atoms[0], atoms[1], atoms[2]
		n := poNot
		and := func(a, b *propertyOracleFormula) *propertyOracleFormula { return poBin("and", a, b) }
		or := func(a, b *propertyOracleFormula) *propertyOracleFormula { return poBin("or", a, b) }
		imp := func(a, b *propertyOracleFormula) *propertyOracleFormula { return poBin("if", a, b) }
		iff := func(a, b *propertyOracleFormula) *propertyOracleFormula { return poBin("iff", a, b) }
		rules := []struct {
			name string
			p    []*propertyOracleFormula
			c    *propertyOracleFormula
		}{
			{"modus ponens", []*propertyOracleFormula{imp(p, q), p}, q},
			{"modus tollens", []*propertyOracleFormula{imp(p, q), n(q)}, n(p)},
			{"hypothetical syllogism", []*propertyOracleFormula{imp(p, q), imp(q, r)}, imp(p, r)},
			{"conjunction introduction", []*propertyOracleFormula{p, q}, and(p, q)},
			{"conjunction elimination", []*propertyOracleFormula{and(p, q)}, p},
			{"disjunction introduction", []*propertyOracleFormula{p}, or(p, q)},
			{"disjunctive syllogism", []*propertyOracleFormula{or(p, q), n(p)}, q},
			{"double negation", []*propertyOracleFormula{n(n(p))}, p},
			{"double negation introduction", []*propertyOracleFormula{p}, n(n(p))},
			{"contraposition", []*propertyOracleFormula{imp(p, q)}, imp(n(q), n(p))},
			{"De Morgan", []*propertyOracleFormula{n(and(p, q))}, or(n(p), n(q))},
			{"De Morgan", []*propertyOracleFormula{n(or(p, q))}, and(n(p), n(q))},
			{"commutation", []*propertyOracleFormula{and(p, q)}, and(q, p)},
			{"commutation", []*propertyOracleFormula{or(p, q)}, or(q, p)},
			{"exportation", []*propertyOracleFormula{imp(and(p, q), r)}, imp(p, imp(q, r))},
			{"importation", []*propertyOracleFormula{imp(p, imp(q, r))}, imp(and(p, q), r)},
			{"material implication", []*propertyOracleFormula{imp(p, q)}, or(n(p), q)},
			{"material implication", []*propertyOracleFormula{or(n(p), q)}, imp(p, q)},
			{"distribution", []*propertyOracleFormula{and(p, or(q, r))}, or(and(p, q), and(p, r))},
			{"distribution", []*propertyOracleFormula{or(p, and(q, r))}, and(or(p, q), or(p, r))},
			{"absorption", []*propertyOracleFormula{or(p, and(p, q))}, p},
			{"absorption", []*propertyOracleFormula{imp(p, q)}, imp(p, and(p, q))},
			{"biconditional introduction", []*propertyOracleFormula{imp(p, q), imp(q, p)}, iff(p, q)},
			{"biconditional elimination", []*propertyOracleFormula{iff(p, q)}, imp(p, q)},
			{"biconditional modus ponens", []*propertyOracleFormula{iff(p, q), p}, q},
			{"biconditional modus tollens", []*propertyOracleFormula{iff(p, q), n(q)}, n(p)},
			{"biconditional chaining", []*propertyOracleFormula{iff(p, q), iff(q, r)}, iff(p, r)},
			{"biconditional commutation", []*propertyOracleFormula{iff(p, q)}, iff(q, p)},
			{"negated conditional", []*propertyOracleFormula{n(imp(p, q))}, p},
			{"negated conditional", []*propertyOracleFormula{n(imp(p, q))}, n(q)},
			{"simple dilemma", []*propertyOracleFormula{or(p, q), imp(p, r), imp(q, r)}, r},
			{"constructive dilemma", []*propertyOracleFormula{or(p, q), imp(p, r), imp(q, p)}, or(r, p)},
			{"destructive dilemma", []*propertyOracleFormula{imp(p, q), imp(q, r), or(n(q), n(r))}, or(n(p), n(q))},
		}
		for _, rule := range rules {
			base := propertyOracleCandidate{name: rule.name, premises: rule.p, conclusion: rule.c, mustAccept: true}
			result = append(result, base)
			flip := func(f *propertyOracleFormula) *propertyOracleFormula {
				neg := f.op == "not"
				inner := f
				if neg {
					inner = f.left
				}
				if inner.op != "predicate" && inner.op != "possession" {
					return nil
				}
				changed := *inner
				changed.op = "possession"
				if inner.op == "possession" {
					changed.op = "predicate"
				}
				if neg {
					return n(&changed)
				}
				return &changed
			}
			// Insert a conversion of an atomic premise before the old rule.
			for slot, premise := range rule.p {
				alt := flip(premise)
				if alt == nil {
					continue
				}
				var out strings.Builder
				ps := append([]*propertyOracleFormula{}, rule.p...)
				ps[slot] = alt
				refs := make([]string, len(ps))
				for i, f := range ps {
					fmt.Fprintf(&out, "%d. %s.\n", i+1, f.english())
					refs[i] = fmt.Sprint(i + 1)
				}
				fmt.Fprintf(&out, "%d. Therefore, %s. [%d]\n", len(ps)+1, premise.english(), slot+1)
				refs[slot] = fmt.Sprint(len(ps) + 1)
				fmt.Fprintf(&out, "%d. Therefore, %s. [%s]", len(ps)+2, rule.c.english(), strings.Join(refs, ","))
				result = append(result, propertyOracleCandidate{name: "rewrite then " + rule.name, premises: ps, conclusion: rule.c, argument: out.String(), mustAccept: true})
			}
			if alt := flip(rule.c); alt != nil {
				arg := base.text() + fmt.Sprintf("\n%d. Therefore, %s. [%d]", len(rule.p)+2, alt.english(), len(rule.p)+1)
				result = append(result, propertyOracleCandidate{name: rule.name + " then rewrite", premises: rule.p, conclusion: alt, argument: arg, mustAccept: true})
			}
		}
	}
	return result
}

func TestEntailmentGatePropertyOracle(t *testing.T) {
	// identity.constant_sides: the historical UG surface lacks strict constant evidence.
	// Keep every independent model check, including target steps in a refused proof.
	refused := map[string]strictRefusal{
		"1. Socrates is identical to Plato.\n| 2. Let P be an arbitrary property that Plato has.\n| 3. Therefore, Socrates has P. [1,2]\n4. Therefore, Socrates has every property that Plato has. [2-3]": {"identity.constant_sides", "name the thing"},
	}

	candidates := append(propertyOracleCandidates(), propertyOracleCompositions()...)
	accepted := make([]bool, len(candidates))
	counterexamples := make([]bool, len(candidates))
	byForm := map[string]int{}
	for i, c := range candidates {
		report, result := CheckEntailmentGate(c.text())
		accepted[i] = report.Passed
		form := ""
		if len(report.Steps) > 0 {
			form = report.Steps[len(report.Steps)-1].Form
		}
		if c.targetStep != 0 {
			found := false
			for _, step := range report.Steps {
				if step.Step == c.targetStep {
					accepted[i], form, found = step.Passed, step.Form, true
				}
			}
			if !found {
				t.Fatalf("missing target step for %s: %s", c.name, result.Message)
			}
		}
		_, strictRefused := refused[c.text()]
		if strictRefused {
			r := refused[c.text()]
			assertStrictRefusal(t, c.text(), r.rule, r.hint)
		}
		if c.mustAccept && !strictRefused && !report.Passed {
			t.Fatalf("%s refused: %s\n%s", c.name, result.Message, c.text())
		}
		if !strictRefused && c.wantForm != "" && form != c.wantForm {
			t.Fatalf("%s: %+v", c.name, report)
		}
		for _, step := range report.Steps {
			if step.Passed {
				byForm[step.Form]++
			}
		}
	}
	models := 0
	for size := 1; size <= 3; size++ {
		for a := 0; a < size; a++ {
			for b := 0; b < size; b++ {
				for c := 0; c < size; c++ {
					for f := 0; f < 1<<size; f++ {
						for g := 0; g < 1<<size; g++ {
							for h := 0; h < 1<<size; h++ {
								m := propertyOracleModel{size, [3]int{a, b, c}, [3]int{f, g, h}}
								models++
								for i, candidate := range candidates {
									extensions := 1
									if candidate.allProperties {
										extensions = 1 << size
									}
									for eigen := 0; eigen < extensions; eigen++ {
										premisesTrue := true
										for _, p := range candidate.premises {
											if !p.eval(m, eigen) {
												premisesTrue = false
												break
											}
										}
										if premisesTrue && !candidate.conclusion.eval(m, eigen) {
											counterexamples[i] = true
											// Unconditional: any acceptance is checked, including mutated cases.
											if accepted[i] {
												t.Fatalf("unsound %s in %+v\n%s", candidate.name, m, candidate.text())
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	rejected := 0
	for i, c := range candidates {
		if !c.mustAccept {
			rejected++
			if !counterexamples[i] {
				t.Fatalf("invalid control lacks a countermodel: %s", c.name)
			}
		}
	}
	for _, name := range []string{"property abstraction", "property concretion", "property instantiation", "property existential", "shared property", "indiscernibility of identicals", "property generalization", "property instantiation (arbitrary property)", "distinguishing property", "distinguishing property nonidentity", "distinguishing property non-inclusion"} {
		if byForm[name] == 0 {
			t.Fatalf("uncovered form %s", name)
		}
	}
	if models != 14344 {
		t.Fatalf("model count %d", models)
	}
	t.Logf("property oracle: %d models, %d candidates, %d invalid controls; forms=%v", models, len(candidates), rejected, byForm)
}
