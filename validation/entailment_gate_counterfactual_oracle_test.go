package validation

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// Independent finite VW semantics. No gate parser, content mapper, matcher,
// or normalization participates in evaluating truth. Ranks are contiguous
// total preorders, with the evaluation world tied at rank zero (weak, never
// strong, centering). Every valuation and every frame on 1..3 worlds is used.
type sphereFormula struct {
	op   string
	atom int
	a, b *sphereFormula
}

func sphereAtom(i int) *sphereFormula           { return &sphereFormula{op: "atom", atom: i} }
func sphereNot(a *sphereFormula) *sphereFormula { return &sphereFormula{op: "not", a: a} }
func sphereBin(op string, a, b *sphereFormula) *sphereFormula {
	return &sphereFormula{op: op, a: a, b: b}
}

type sphereModel struct {
	n         int
	ranks     [3][3]int
	valuation int
}

func (f *sphereFormula) truth(m sphereModel, atoms int, cache map[*sphereFormula]uint8) uint8 {
	if v, ok := cache[f]; ok {
		return v
	}
	full := uint8(1<<m.n) - 1
	var v uint8
	if f.op == "atom" {
		for w := 0; w < m.n; w++ {
			if m.valuation&(1<<(w*atoms+f.atom)) != 0 {
				v |= 1 << w
			}
		}
	} else {
		a := f.a.truth(m, atoms, cache)
		var b uint8
		if f.b != nil {
			b = f.b.truth(m, atoms, cache)
		}
		switch f.op {
		case "not":
			v = full ^ a
		case "and":
			v = a & b
		case "or":
			v = a | b
		case "if":
			v = (full ^ a) | b
		case "iff":
			v = full ^ (a ^ b)
		case "cf", "might":
			for w := 0; w < m.n; w++ {
				min := m.n
				var closest uint8
				for u := 0; u < m.n; u++ {
					if a&(1<<u) == 0 {
						continue
					}
					rank := m.ranks[w][u]
					if rank < min {
						min = rank
						closest = 0
					}
					if rank == min {
						closest |= 1 << u
					}
				}
				if (f.op == "cf" && closest & ^b == 0) || (f.op == "might" && closest&b != 0) {
					v |= 1 << w
				}
			}
		default:
			panic(f.op)
		}
	}
	cache[f] = v
	return v
}

// mood belongs to the conditional operator, not to the model's atomic facts.
func (f *sphereFormula) english(mood string) string {
	if f.op == "atom" {
		subject := []string{"the match", "the lamp", "the room"}[f.atom]
		predicate := []string{"struck", "lit", "warm"}[f.atom]
		aux := "was"
		switch mood {
		case "a":
			aux = "had been"
		case "c":
			aux = "would have been"
		case "might":
			aux = "might have been"
		}
		return subject + " " + aux + " " + predicate
	}
	if f.op == "not" {
		return "it is not the case that " + f.a.english(mood)
	}
	switch f.op {
	case "cf":
		return "if " + f.a.english("a") + ", then " + f.b.english("c")
	case "might":
		return "if " + f.a.english("a") + ", then " + f.b.english("might")
	case "if":
		left := f.a.english("")
		if f.a.op == "cf" || f.a.op == "if" {
			left = "(" + left + ")"
		}
		return "if " + left + ", then " + f.b.english("")
	case "iff":
		return "(" + f.a.english("") + ") if and only if (" + f.b.english("") + ")"
	case "and":
		return "it is the case both that " + f.a.english(mood) + ", and that " + f.b.english(mood)
	case "or":
		return "either " + f.a.english(mood) + " or " + f.b.english(mood)
	}
	panic(f.op)
}

type sphereCandidate struct {
	name string
	p    []*sphereFormula
	c    *sphereFormula
	want bool
}

func (c sphereCandidate) argument() string {
	var b strings.Builder
	citations := make([]string, len(c.p))
	for i, p := range c.p {
		fmt.Fprintf(&b, "%d. %s.\n", i+1, p.english(""))
		citations[i] = fmt.Sprint(i + 1)
	}
	fmt.Fprintf(&b, "%d. Therefore, %s. [%s]", len(c.p)+1, c.c.english(""), strings.Join(citations, ","))
	return b.String()
}
func sphereRanks(n, w int) [][3]int {
	var out [][3]int
	limit := 1
	for i := 0; i < n; i++ {
		limit *= n
	}
	for code := 0; code < limit; code++ {
		x := code
		var r [3]int
		seen := 0
		max := 0
		for i := 0; i < n; i++ {
			r[i] = x % n
			x /= n
			seen |= 1 << r[i]
			if r[i] > max {
				max = r[i]
			}
		}
		if r[w] == 0 && seen == (1<<(max+1))-1 {
			out = append(out, r)
		}
	}
	return out
}
func sphereCandidates() ([]sphereCandidate, []sphereCandidate) {
	a, b, c := sphereAtom(0), sphereAtom(1), sphereAtom(2)
	cf := func(p, q *sphereFormula) *sphereFormula { return sphereBin("cf", p, q) }
	and := func(p, q *sphereFormula) *sphereFormula { return sphereBin("and", p, q) }
	or := func(p, q *sphereFormula) *sphereFormula { return sphereBin("or", p, q) }
	imp := func(p, q *sphereFormula) *sphereFormula { return sphereBin("if", p, q) }
	var candidates []sphereCandidate
	add := func(name string, p []*sphereFormula, c *sphereFormula) {
		candidates = append(candidates, sphereCandidate{name, p, c, true})
	}
	for _, p := range []*sphereFormula{a, b, c, sphereNot(a)} {
		for _, q := range []*sphereFormula{a, b, c, sphereNot(b)} {
			add("MP", []*sphereFormula{cf(p, q), p}, q)
			add("MT", []*sphereFormula{cf(p, q), sphereNot(q)}, sphereNot(p))
			add("to material", []*sphereFormula{cf(p, q)}, imp(p, q))
			add("to disjunction", []*sphereFormula{cf(p, q)}, or(sphereNot(p), q))
			add("conjunction", []*sphereFormula{cf(p, q), cf(p, c)}, cf(p, and(q, c)))
			add("elimination", []*sphereFormula{cf(p, and(q, c))}, cf(p, q))
			add("weakening", []*sphereFormula{cf(p, q)}, cf(p, or(q, c)))
			add("dilemma", []*sphereFormula{or(p, c), cf(p, q), cf(c, q)}, q)
			add("reflexivity", []*sphereFormula{or(p, q)}, cf(p, p))
		}
	}
	excluded := []sphereCandidate{
		{"contraposition", []*sphereFormula{cf(a, b)}, cf(sphereNot(b), sphereNot(a)), false},
		{"hypothetical syllogism", []*sphereFormula{cf(a, b), cf(b, c)}, cf(a, c), false},
		{"antecedent strengthening / conditional proof", []*sphereFormula{cf(a, c)}, cf(and(a, b), c), false},
		{"false antecedent / conditional proof reductio", []*sphereFormula{sphereNot(a)}, cf(a, c), false},
		{"exportation", []*sphereFormula{cf(and(a, b), c)}, imp(a, cf(b, c)), false},
		{"importation", []*sphereFormula{imp(a, cf(b, c))}, cf(and(a, b), c), false},
		{"negated conditional antecedent", []*sphereFormula{sphereNot(cf(a, b))}, a, false},
		{"negated conditional consequent", []*sphereFormula{sphereNot(cf(a, b))}, sphereNot(b), false},
		{"material to counterfactual", []*sphereFormula{imp(a, b)}, cf(a, b), false},
		{"disjunction to counterfactual", []*sphereFormula{or(sphereNot(a), b)}, cf(a, b), false},
		{"true consequent / conditional proof", []*sphereFormula{b}, cf(a, b), false},
		{"strong centering", []*sphereFormula{a, b}, cf(a, b), false},
		{"simplification of disjunctive antecedents", []*sphereFormula{cf(or(a, b), c)}, cf(a, c), false},
		{"conditional excluded middle", []*sphereFormula{c}, or(cf(a, b), cf(a, sphereNot(b))), false},
		{"might to would", []*sphereFormula{sphereBin("might", a, b)}, cf(a, b), false},
	}
	// These use both directions of polarity and mixed cf/material nesting.
	mixed := []sphereCandidate{
		{"mixed HS to material", []*sphereFormula{cf(a, b), imp(b, c)}, imp(a, c), false},
		{"material contraposition", []*sphereFormula{cf(a, b)}, imp(sphereNot(b), sphereNot(a)), false},
		{"material disjunction weakening", []*sphereFormula{cf(a, b)}, or(sphereNot(a), b), false},
		{"negative cf conclusion", []*sphereFormula{sphereNot(or(sphereNot(a), b))}, sphereNot(cf(a, b)), false},
		{"opaque negative reiteration", []*sphereFormula{sphereNot(cf(a, b))}, sphereNot(cf(a, b)), true},
		{"opaque positive conjunction", []*sphereFormula{cf(a, b), c}, and(cf(a, b), c), true},
		{"opaque disjunction", []*sphereFormula{cf(a, b)}, or(cf(a, b), c), true},
		{"negative antecedent cannot be weakened", []*sphereFormula{imp(cf(a, b), c), imp(a, b)}, c, false},
		{"positive nested conclusion", []*sphereFormula{cf(and(a, b), c)}, imp(a, cf(b, c)), false},
		{"positive conjunct cannot be manufactured", []*sphereFormula{imp(a, b), c}, and(cf(a, b), c), false},
		{"positive disjunct cannot be manufactured", []*sphereFormula{imp(a, b)}, or(cf(a, b), c), false},
		{"negative disjunction premise", []*sphereFormula{sphereNot(or(cf(a, b), c))}, a, false},
		{"biconditional premise both polarities", []*sphereFormula{sphereBin("iff", cf(a, b), c), imp(a, b)}, c, false},
		{"biconditional conclusion both polarities", []*sphereFormula{cf(a, b), imp(b, a)}, sphereBin("iff", a, b), false},
		{"nested antecedent double flip", []*sphereFormula{cf(a, b)}, imp(imp(cf(a, b), c), c), false},
	}
	candidates = append(candidates, mixed...)
	candidates = append(candidates, excluded...)
	// Seeded near misses and cross-connective candidates prevent an oracle
	// consisting solely of handpicked successes. Every acceptance is checked.
	pool := []*sphereFormula{a, b, c, sphereNot(a), sphereNot(b), cf(a, b), cf(b, c), cf(a, c), cf(a, and(b, c)), cf(or(a, b), c), cf(a, or(b, c)), cf(a, sphereNot(b)), imp(a, b), or(a, b), and(a, b), sphereNot(cf(a, b)), sphereBin("might", a, b)}
	pool = append(pool, imp(cf(a, b), c), imp(a, cf(b, c)), and(cf(a, b), c), or(cf(a, b), c), sphereNot(or(cf(a, b), c)), sphereBin("iff", cf(a, b), c), sphereNot(or(sphereNot(a), b)))
	rng := rand.New(rand.NewSource(19))
	for i := 0; i < 250; i++ {
		p := []*sphereFormula{pool[rng.Intn(len(pool))]}
		if i%2 == 0 {
			p = append(p, pool[rng.Intn(len(pool))])
		}
		candidates = append(candidates, sphereCandidate{"seeded near miss", p, pool[rng.Intn(len(pool))], false})
	}
	return candidates, excluded
}
func TestCounterfactualSphereOracle(t *testing.T) {
	candidates, excluded := sphereCandidates()
	type rendered struct {
		candidate sphereCandidate
		text      string
	}
	var accepted []rendered
	forms := map[string]int{}
	renderings := 0
	for _, candidate := range candidates {
		texts := []string{candidate.argument()}
		for _, words := range [][]string{
			{"the match had been struck", "the lamp would have been lit", "the room was warm"},
			{"the match had been struck", "the lamp would have been lit", "the room would have been warm"},
		} {
			if text, ok := candidate.verbatimArgument(words); ok {
				texts = append(texts, text)
			}
		}
		for i, text := range texts {
			renderings++
			report, result := CheckEntailmentGate(text)
			if i == 0 && candidate.want && !report.Passed {
				t.Fatalf("intended %s failed: %s\n%s", candidate.name, result.Message, text)
			}
			if report.Passed {
				accepted = append(accepted, rendered{candidate, text})
				forms[report.Steps[0].Form]++
			}
		}
	}
	found := make([]bool, len(excluded))
	models, points, frames := 0, 0, 0
	for atoms := 2; atoms <= 3; atoms++ {
		for n := 1; n <= 3; n++ {
			var choices [3][][3]int
			for w := 0; w < n; w++ {
				choices[w] = sphereRanks(n, w)
			}
			var enumerate func(int, sphereModel)
			enumerate = func(w int, m sphereModel) {
				if w < n {
					for _, rank := range choices[w] {
						m.ranks[w] = rank
						enumerate(w+1, m)
					}
					return
				}
				frames++
				for valuation := 0; valuation < 1<<(atoms*n); valuation++ {
					m.valuation = valuation
					models++
					points += n
					cache := map[*sphereFormula]uint8{}
					check := func(candidate sphereCandidate) uint8 {
						// Candidates using the third atom belong to the three-atom run.
						var usesThird func(*sphereFormula) bool
						usesThird = func(f *sphereFormula) bool {
							return f != nil && ((f.op == "atom" && f.atom >= atoms) || usesThird(f.a) || usesThird(f.b))
						}
						if usesThird(candidate.c) {
							return 0
						}
						for _, p := range candidate.p {
							if usesThird(p) {
								return 0
							}
						}
						premises := uint8(1<<n) - 1
						for _, p := range candidate.p {
							premises &= p.truth(m, atoms, cache)
						}
						return premises & ^candidate.c.truth(m, atoms, cache)
					}
					for _, accepted := range accepted {
						candidate := accepted.candidate
						// UNCONDITIONAL: no expected-valid flag or rule-name filter can exempt
						// a gate-certified step from the soundness assertion.
						if bad := check(candidate); bad != 0 {
							t.Fatalf("unsound certificate: %s\n%s\nmodel=%+v atoms=%d worlds=%03b", candidate.name, accepted.text, m, atoms, bad)
						}
					}
					for i, candidate := range excluded {
						if !found[i] && check(candidate) != 0 {
							found[i] = true
						}
					}
				}
			}
			enumerate(0, sphereModel{n: n})
		}
	}
	for i, ok := range found {
		if !ok {
			t.Errorf("oracle self-check: no countermodel for %s", excluded[i].name)
		}
	}
	t.Logf("%d candidates, %d renderings, %d accepted, %d frames across 2/3 atoms, %d models, %d pointed worlds; %d excluded patterns have countermodels; forms=%v", len(candidates), renderings, len(accepted), frames, models, points, len(excluded), forms)
}

// A second surface uses one fixed spelling per atom, so material schemas can
// match the old verbatim would/had clauses. The independent renderer admits a
// surface only if its grammatical conditional kinds agree with the model AST.
// It does not call the gate's classifier or parser.
func (f *sphereFormula) verbatimEnglish(words []string) (string, bool) {
	if f.op == "atom" {
		return words[f.atom], true
	}
	a, ok := f.a.verbatimEnglish(words)
	if !ok {
		return "", false
	}
	if f.op == "not" {
		return "it is not the case that " + a, true
	}
	b, ok := f.b.verbatimEnglish(words)
	if !ok {
		return "", false
	}
	var wouldContent func(*sphereFormula) bool
	wouldContent = func(x *sphereFormula) bool {
		switch x.op {
		case "atom":
			return strings.Contains(words[x.atom], " would ")
		case "not":
			return wouldContent(x.a)
		case "and", "or":
			return wouldContent(x.a) || wouldContent(x.b)
		default:
			return false
		}
	}
	switch f.op {
	case "cf", "if":
		if wouldContent(f.b) != (f.op == "cf") {
			return "", false
		}
		if f.a.op == "cf" || f.a.op == "if" {
			a = "(" + a + ")"
		}
		return "if " + a + ", then " + b, true
	case "and":
		return "it is the case both that " + a + ", and that " + b, true
	case "or":
		return "either " + a + " or " + b, true
	case "iff":
		return "(" + a + ") if and only if (" + b + ")", true
	}
	return "", false
}
func (c sphereCandidate) verbatimArgument(words []string) (string, bool) {
	var out strings.Builder
	var cites []string
	for i, p := range c.p {
		text, ok := p.verbatimEnglish(words)
		if !ok {
			return "", false
		}
		fmt.Fprintf(&out, "%d. %s.\n", i+1, text)
		cites = append(cites, fmt.Sprint(i+1))
	}
	text, ok := c.c.verbatimEnglish(words)
	if !ok {
		return "", false
	}
	fmt.Fprintf(&out, "%d. Therefore, %s. [%s]", len(c.p)+1, text, strings.Join(cites, ","))
	return out.String(), true
}
