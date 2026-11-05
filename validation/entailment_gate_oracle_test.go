package validation

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
)

// The oracle owns its AST and semantics: no gate parsing, normalization, or
// schema matching is used to decide truth. All eight valuations are checked.
// This is bounded, deterministic coverage, not a completeness claim.
type gateOracleFormula struct {
	op          string
	atom        int
	left, right *gateOracleFormula
}

func gateOracleAtom(i int) *gateOracleFormula { return &gateOracleFormula{op: "atom", atom: i} }
func gateOracleNot(p *gateOracleFormula) *gateOracleFormula {
	return &gateOracleFormula{op: "not", left: p}
}
func gateOracleBinary(op string, p, q *gateOracleFormula) *gateOracleFormula {
	return &gateOracleFormula{op: op, left: p, right: q}
}

func (f *gateOracleFormula) depth() int {
	if f.op == "atom" {
		return 0
	}
	d := f.left.depth()
	if f.right != nil && f.right.depth() > d {
		d = f.right.depth()
	}
	return d + 1
}

func (f *gateOracleFormula) eval(valuation int) bool {
	switch f.op {
	case "atom":
		return valuation&(1<<f.atom) != 0
	case "not":
		return !f.left.eval(valuation)
	case "and":
		return f.left.eval(valuation) && f.right.eval(valuation)
	case "or":
		return f.left.eval(valuation) || f.right.eval(valuation)
	case "if":
		return !f.left.eval(valuation) || f.right.eval(valuation)
	case "iff":
		return f.left.eval(valuation) == f.right.eval(valuation)
	default:
		panic("unknown oracle connective: " + f.op)
	}
}

func (f *gateOracleFormula) literal() bool {
	return f.op == "atom" || (f.op == "not" && f.left.op == "atom")
}

// Render bounded formulas with explicit scope. Conjunctions containing any
// compound child use the both-that scaffold; a negated conjunction does too.
// Either joins literals or atomic conjunctions; nested conditionals occur only
// in consequents. Compound iff sides use the parenthesized v16 scaffold; no ambiguous
// flattened and/or tree is emitted.
func (f *gateOracleFormula) english() (string, bool) {
	if f.depth() > 2 {
		return "", false
	}
	if f.op == "atom" {
		return []string{"it rains", "it snows", "the street is wet"}[f.atom], true
	}
	left, ok := f.left.english()
	if !ok {
		return "", false
	}
	if f.op == "not" {
		if f.left.op == "and" {
			l, _ := f.left.left.english()
			r, _ := f.left.right.english()
			left = "it is the case both that " + l + ", and that " + r
		}
		return "it is not the case that " + left, true
	}
	right, ok := f.right.english()
	if !ok {
		return "", false
	}
	switch f.op {
	case "and":
		if f.left.op == "atom" && f.right.op == "atom" {
			return left + " and " + right, true
		}
		return "it is the case both that " + left + ", and that " + right, true
	case "or":
		if (f.left.literal() || f.left.op == "and") && (f.right.literal() || f.right.op == "and") {
			return "either " + left + " or " + right, true
		}
	case "if":
		if (f.left.literal() || f.left.op == "and" || f.left.op == "or") && (f.right.literal() || f.right.op == "if" || f.right.op == "and") {
			return "if " + left + ", then " + right, true
		}
	case "iff":
		if f.left.op == "atom" && f.right.op == "atom" {
			return left + " if and only if " + right, true
		}
		return "it holds that (" + left + ") if and only if (" + right + ")", true
	}
	return "", false
}

func gateOraclePool() []*gateOracleFormula {
	pool := []*gateOracleFormula{gateOracleAtom(0), gateOracleAtom(1), gateOracleAtom(2)}
	seen := make(map[string]bool)
	for _, f := range pool {
		s, _ := f.english()
		seen[s] = true
	}
	add := func(f *gateOracleFormula) {
		if s, ok := f.english(); ok && !seen[s] {
			seen[s] = true
			pool = append(pool, f)
		}
	}
	for depth := 1; depth <= 2; depth++ {
		previous := append([]*gateOracleFormula(nil), pool...)
		for _, p := range previous {
			add(gateOracleNot(p))
			for _, q := range previous {
				for _, op := range []string{"and", "or", "if", "iff"} {
					add(gateOracleBinary(op, p, q))
				}
			}
		}
	}
	return pool
}

type gateOracleCandidate struct {
	premises   []*gateOracleFormula
	conclusion *gateOracleFormula
}

func (c gateOracleCandidate) argument() (string, bool) {
	var b strings.Builder
	citations := make([]string, len(c.premises))
	for i, p := range c.premises {
		s, ok := p.english()
		if !ok {
			return "", false
		}
		fmt.Fprintf(&b, "%d. %s.\n", i+1, s)
		citations[i] = fmt.Sprint(i + 1)
	}
	s, ok := c.conclusion.english()
	if !ok {
		return "", false
	}
	fmt.Fprintf(&b, "%d. Therefore, %s. [%s]", len(c.premises)+1, s, strings.Join(citations, ","))
	return b.String(), true
}

// counterexample returns the first valuation satisfying all premises and
// falsifying the conclusion, or -1 when the inference is classically valid.
func (c gateOracleCandidate) counterexample() int {
	for valuation := 0; valuation < 8; valuation++ {
		all := true
		for _, premise := range c.premises {
			all = all && premise.eval(valuation)
		}
		if all && !c.conclusion.eval(valuation) {
			return valuation
		}
	}
	return -1
}

func gateOracleCandidates(pool []*gateOracleFormula) []gateOracleCandidate {
	var candidates []gateOracleCandidate
	seen := make(map[string]bool)
	add := func(premises []*gateOracleFormula, conclusion *gateOracleFormula) {
		c := gateOracleCandidate{premises: premises, conclusion: conclusion}
		if text, ok := c.argument(); ok && !seen[text] {
			seen[text] = true
			candidates = append(candidates, c)
		}
	}
	atoms := pool[:3]
	not := gateOracleNot
	and := func(p, q *gateOracleFormula) *gateOracleFormula { return gateOracleBinary("and", p, q) }
	or := func(p, q *gateOracleFormula) *gateOracleFormula { return gateOracleBinary("or", p, q) }
	implies := func(p, q *gateOracleFormula) *gateOracleFormula { return gateOracleBinary("if", p, q) }
	iff := func(p, q *gateOracleFormula) *gateOracleFormula { return gateOracleBinary("iff", p, q) }

	// Exhaustive substitutions in common schemas, plus nearby conclusions.
	// Degenerate substitutions (P=Q) are intentional, and validity is still
	// decided by the truth table, never by the template's name.
	for _, p := range atoms {
		for _, q := range atoms {
			for _, r := range atoms {
				seeds := []gateOracleCandidate{
					// v16: (P -> Q) <-> R, its valid uses and invalid projections.
					{[]*gateOracleFormula{iff(implies(p, q), r), not(r)}, not(implies(p, q))},
					{[]*gateOracleFormula{iff(implies(p, q), r), not(implies(p, q))}, not(r)},
					{[]*gateOracleFormula{iff(implies(p, q), r), iff(r, p)}, iff(implies(p, q), p)},
					{[]*gateOracleFormula{iff(implies(p, q), r), r}, implies(p, q)},
					{[]*gateOracleFormula{iff(implies(p, q), r), implies(p, q)}, r},
					{[]*gateOracleFormula{iff(implies(p, q), r)}, iff(r, implies(p, q))},
					{[]*gateOracleFormula{iff(implies(p, q), r)}, p},
					{[]*gateOracleFormula{iff(implies(p, q), r)}, q},
					{[]*gateOracleFormula{iff(implies(p, q), r)}, r},
					// Scope controls: a nested connective is not a top-level one.
					{[]*gateOracleFormula{q}, not(or(p, q))},
					{[]*gateOracleFormula{q}, and(p, or(r, q))},
					{[]*gateOracleFormula{implies(p, and(q, r))}, r},
					{[]*gateOracleFormula{or(p, and(q, r))}, r},
					{[]*gateOracleFormula{implies(p, q)}, implies(not(q), not(p))},
					{[]*gateOracleFormula{implies(not(q), not(p))}, implies(p, q)},
					{[]*gateOracleFormula{not(and(p, q))}, or(not(p), not(q))},
					{[]*gateOracleFormula{or(not(p), not(q))}, not(and(p, q))},
					{[]*gateOracleFormula{not(or(p, q))}, and(not(p), not(q))},
					{[]*gateOracleFormula{and(not(p), not(q))}, not(or(p, q))},
					{[]*gateOracleFormula{and(p, q)}, and(q, p)},
					{[]*gateOracleFormula{or(p, q)}, or(q, p)},
					{[]*gateOracleFormula{implies(and(p, q), r)}, implies(p, implies(q, r))},
					{[]*gateOracleFormula{implies(p, implies(q, r))}, implies(and(p, q), r)},
					{[]*gateOracleFormula{implies(p, q)}, or(not(p), q)},
					{[]*gateOracleFormula{or(not(p), q)}, implies(p, q)},
					{[]*gateOracleFormula{and(p, or(q, r))}, or(and(p, q), and(p, r))},
					{[]*gateOracleFormula{or(and(p, q), and(p, r))}, and(p, or(q, r))},
					{[]*gateOracleFormula{or(p, and(q, r))}, and(or(p, q), or(p, r))},
					{[]*gateOracleFormula{and(or(p, q), or(p, r))}, or(p, and(q, r))},
					{[]*gateOracleFormula{or(p, and(p, q))}, p},
					{[]*gateOracleFormula{implies(p, q)}, implies(p, and(p, q))},
					{[]*gateOracleFormula{iff(p, q), not(q)}, not(p)},
					{[]*gateOracleFormula{iff(p, q), not(p)}, not(q)},
					{[]*gateOracleFormula{iff(p, q), iff(q, r)}, iff(p, r)},
					{[]*gateOracleFormula{iff(p, q)}, iff(q, p)},
					{[]*gateOracleFormula{not(implies(p, q))}, p},
					{[]*gateOracleFormula{not(implies(p, q))}, not(q)},
					{[]*gateOracleFormula{implies(p, q), p}, q},
					{[]*gateOracleFormula{implies(p, q), not(q)}, not(p)},
					{[]*gateOracleFormula{implies(p, q), implies(q, r)}, implies(p, r)},
					{[]*gateOracleFormula{p, q}, and(p, q)},
					{[]*gateOracleFormula{and(p, q)}, p},
					{[]*gateOracleFormula{p}, or(p, q)},
					{[]*gateOracleFormula{or(p, q), not(p)}, q},
					{[]*gateOracleFormula{not(not(p))}, p},
					{[]*gateOracleFormula{p}, not(not(p))},
					{[]*gateOracleFormula{implies(p, q), implies(q, p)}, iff(p, q)},
					{[]*gateOracleFormula{iff(p, q)}, implies(q, p)},
					{[]*gateOracleFormula{iff(p, q), p}, q},
					{[]*gateOracleFormula{implies(and(p, q), r), p, q}, r},
					{[]*gateOracleFormula{or(p, q), implies(p, r), implies(q, r)}, r},
					{[]*gateOracleFormula{or(p, q), implies(p, r), implies(q, p)}, or(r, p)},
					{[]*gateOracleFormula{implies(p, q), implies(q, r), or(not(q), not(r))}, or(not(p), not(q))},
				}
				for _, seed := range seeds {
					add(seed.premises, seed.conclusion)
					for _, replacement := range atoms {
						add(seed.premises, replacement)
						add(seed.premises, not(replacement))
					}
				}
			}
		}
	}
	// Cover every rendered formula, probing proper subformulas as conclusions.
	for _, f := range pool {
		add([]*gateOracleFormula{f}, f)
		if f.left != nil {
			add([]*gateOracleFormula{f}, f.left)
		}
		if f.right != nil {
			add([]*gateOracleFormula{f}, f.right)
		}
	}
	// Fixed-seed shared-subformula sampling, plus unrelated negative controls.
	// Each arity 1..3 is sampled equally; no wall-clock seed or rejection loop.
	rng := rand.New(rand.NewSource(12012026))
	for i := 0; i < 1800; i++ {
		f := pool[rng.Intn(len(pool))]
		premises := []*gateOracleFormula{f}
		for len(premises) < 1+i%3 {
			p := pool[rng.Intn(len(pool))]
			if len(premises) == 1 && i%2 == 0 && f.left != nil {
				p = f.left
			}
			premises = append(premises, p)
		}
		conclusion := pool[rng.Intn(len(pool))]
		if i%2 == 0 && f.right != nil {
			conclusion = f.right
		}
		add(premises, conclusion)
	}
	return candidates
}

func TestEntailmentGateTruthTableOracle(t *testing.T) {
	pool := gateOraclePool()
	candidates := gateOracleCandidates(pool)
	type counts struct{ total, valid, accepted int }
	total := counts{}
	byShape := make(map[string]*counts)
	byForm := make(map[string]int)
	for _, c := range candidates {
		argument, ok := c.argument()
		if !ok || len(c.premises) < 1 || len(c.premises) > 3 {
			t.Fatalf("invalid oracle fixture: %+v", c)
		}
		// Check syntax separately so parser regressions cannot silently reduce
		// soundness coverage to a collection of trivial rejections.
		if _, parsed := ParseNumberedArgument(argument, numberedArgumentOptions{}); !parsed.Passed {
			t.Fatalf("oracle syntax failure: %s\n%s", parsed.Message, argument)
		}
		report, result := CheckEntailmentGate(argument)
		if len(report.Steps) != 1 || report.Passed != result.Passed {
			t.Fatalf("inconsistent oracle gate result: %+v %s\n%s", report, result.Message, argument)
		}
		counterexample := c.counterexample()
		shape := fmt.Sprintf("%d-premise/%s/depth-%d", len(c.premises), c.conclusion.op, c.conclusion.depth())
		if byShape[shape] == nil {
			byShape[shape] = &counts{}
		}
		for _, count := range []*counts{&total, byShape[shape]} {
			count.total++
			if counterexample == -1 {
				count.valid++
			}
			if report.Passed {
				count.accepted++
			}
		}
		if report.Passed {
			byForm[report.Steps[0].Form]++
			if counterexample != -1 {
				t.Errorf("UNSOUND form=%q valuation=(rains=%t, snows=%t, wet=%t)\n%s", report.Steps[0].Form,
					counterexample&1 != 0, counterexample&2 != 0, counterexample&4 != 0, argument)
			}
		}
	}
	// Coverage is a separate requirement from the unconditional soundness
	// assertion above. Every new propositional matcher must actually fire.
	for _, form := range []string{
		"contraposition", "De Morgan", "commutation", "exportation", "importation",
		"material implication", "distribution", "absorption", "double negation introduction",
		"biconditional modus tollens", "biconditional chaining", "biconditional commutation",
		"negated conditional",
	} {
		if byForm[form] == 0 {
			t.Errorf("oracle did not exercise new form %q", form)
		}
	}
	if total.accepted == 0 || total.valid == total.total {
		t.Fatal("oracle needs both gate acceptances and truth-table-invalid controls")
	}
	t.Logf("oracle: formulas=%d candidates=%d valid=%d accepted=%d invalid=%d", len(pool), total.total, total.valid, total.accepted, total.total-total.valid)
	shapes := make([]string, 0, len(byShape))
	for shape := range byShape {
		shapes = append(shapes, shape)
	}
	sort.Strings(shapes)
	for _, shape := range shapes {
		c := byShape[shape]
		rate := "n/a (no valid candidates)"
		if c.valid > 0 {
			rate = fmt.Sprintf("%.1f%%", 100*float64(c.accepted)/float64(c.valid))
		}
		t.Logf("shape %s: candidates=%d valid=%d accepted=%d accepted/valid=%s", shape, c.total, c.valid, c.accepted, rate)
	}
	forms := make([]string, 0, len(byForm))
	for form := range byForm {
		forms = append(forms, form)
	}
	sort.Strings(forms)
	for _, form := range forms {
		t.Logf("form %q: accepted=%d", form, byForm[form])
	}
}

func TestGateOracleTruthSemantics(t *testing.T) {
	p, q := gateOracleAtom(0), gateOracleAtom(1)
	// Bit n records truth in valuation n; these are independent truth-table
	// constants, not expectations calculated with the evaluator under test.
	for _, tc := range []struct {
		formula *gateOracleFormula
		mask    uint8
	}{
		{p, 0xaa}, {q, 0xcc}, {gateOracleAtom(2), 0xf0},
		{gateOracleNot(p), 0x55}, {gateOracleNot(gateOracleNot(p)), 0xaa},
		{gateOracleBinary("and", p, q), 0x88},
		{gateOracleBinary("or", p, q), 0xee},
		{gateOracleBinary("if", p, q), 0xdd},
		{gateOracleBinary("iff", p, q), 0x99},
	} {
		var mask uint8
		for v := 0; v < 8; v++ {
			if tc.formula.eval(v) {
				mask |= 1 << v
			}
		}
		if mask != tc.mask {
			t.Errorf("%s truth mask = %02x, want %02x", tc.formula.op, mask, tc.mask)
		}
	}
	for _, tc := range []struct {
		name  string
		c     gateOracleCandidate
		valid bool
	}{
		{"modus ponens", gateOracleCandidate{[]*gateOracleFormula{gateOracleBinary("if", p, q), p}, q}, true},
		{"affirming consequent", gateOracleCandidate{[]*gateOracleFormula{gateOracleBinary("if", p, q), q}, p}, false},
		{"inconsistent premises", gateOracleCandidate{[]*gateOracleFormula{p, gateOracleNot(p)}, q}, true},
	} {
		if got := tc.c.counterexample() == -1; got != tc.valid {
			t.Errorf("%s valid=%v, want %v", tc.name, got, tc.valid)
		}
	}
}

// Wrap every existing one-step candidate in a conditional proof, and derive
// reductios by contradicting its conclusion. The expected semantics use only
// the oracle AST; neither acceptance nor a presumed valid inner schema filters
// candidates. Wrong endpoints and wrong reductio conclusions are also tested.
func TestEntailmentGateFitchTruthTableOracle(t *testing.T) {
	pool := gateOraclePool()
	total, valid, accepted := 0, 0, 0
	byForm := map[string]int{}
	check := func(argument string, premises []*gateOracleFormula, conclusion *gateOracleFormula) {
		t.Helper()
		if _, r := ParseNumberedArgument(argument, numberedArgumentOptions{}); !r.Passed {
			t.Fatalf("oracle syntax: %s\n%s", r.Message, argument)
		}
		report, result := CheckEntailmentGate(argument)
		if report.Passed != result.Passed {
			t.Fatal("inconsistent gate result")
		}
		counterexample := (gateOracleCandidate{premises: premises, conclusion: conclusion}).counterexample()
		total++
		if counterexample == -1 {
			valid++
		}
		if report.Passed {
			accepted++
			byForm[report.Steps[len(report.Steps)-1].Form]++
			if counterexample != -1 {
				t.Errorf("UNSOUND Fitch valuation=%d\n%s", counterexample, argument)
			}
		}
	}
	for _, inner := range gateOracleCandidates(pool) {
		// Bounded inner derivations use the same 1..3-premise candidate pool.
		n := len(inner.premises)
		if n < 1 || n > 3 {
			continue
		}
		var prefix strings.Builder
		for i, p := range inner.premises[1:] {
			text, _ := p.english()
			fmt.Fprintf(&prefix, "%d. %s.\n", i+1, text)
		}
		assumption, _ := inner.premises[0].english()
		end, _ := inner.conclusion.english()
		refs := make([]string, n)
		for i := range refs {
			refs[i] = fmt.Sprint(i + 1)
		}
		body := fmt.Sprintf("%s| %d. Suppose %s.\n| %d. Therefore, %s. [%s]\n", prefix.String(), n, assumption, n+1, end, strings.Join(refs, ","))
		for _, target := range []*gateOracleFormula{inner.conclusion, pool[0]} {
			final := gateOracleBinary("if", inner.premises[0], target)
			text, ok := final.english()
			if !ok {
				continue
			}
			argument := body + fmt.Sprintf("%d. Therefore, %s. [%d-%d]", n+2, text, n, n+1)
			check(argument, inner.premises[1:], final)
		}
		// Outer premises include not Q. Within P, infer Q and contradiction;
		// discharge P to not P. Its classical reverse starts with not P seeds.
		negEnd := gateOracleNot(inner.conclusion)
		negText, ok := negEnd.english()
		if !ok {
			continue
		}
		var outer strings.Builder
		outer.WriteString(prefix.String())
		fmt.Fprintf(&outer, "%d. %s.\n", n, negText)
		refs[n-1] = fmt.Sprint(n + 1)
		reductioBody := fmt.Sprintf("%s| %d. Suppose %s.\n| %d. Therefore, %s. [%s]\n| %d. Therefore, contradiction. [%d,%d]\n", outer.String(), n+1, assumption, n+2, end, strings.Join(refs, ","), n+3, n, n+2)
		targets := []*gateOracleFormula{gateOracleNot(inner.premises[0]), pool[0]}
		if inner.premises[0].op == "not" {
			targets = append(targets, inner.premises[0].left)
		}
		for _, target := range targets {
			text, ok := target.english()
			if !ok {
				continue
			}
			premises := append(append([]*gateOracleFormula(nil), inner.premises[1:]...), negEnd)
			check(reductioBody+fmt.Sprintf("%d. Therefore, %s. [%d-%d]", n+4, text, n+1, n+3), premises, target)
		}
	}
	for _, form := range []string{"conditional proof", "reductio"} {
		if byForm[form] == 0 {
			t.Errorf("no accepted %s", form)
		}
	}
	if total == valid {
		t.Fatal("oracle needs invalid controls")
	}
	t.Logf("Fitch oracle: formulas=%d candidates=%d valid=%d accepted=%d invalid=%d forms=%v", len(pool), total, valid, accepted, total-valid, byForm)
}

// A separate modal AST interprets formulas on all three-world reflexive
// frames, including non-transitive and non-Euclidean frames. It shares no
// parsing or matching code with the gate. This is bounded soundness coverage.
type gateModalOracleFormula struct {
	op          string
	left, right *gateModalOracleFormula
}

func (f *gateModalOracleFormula) english() string {
	switch f.op {
	case "p":
		return "it rains"
	case "q":
		return "it snows"
	case "not":
		return "it is not the case that " + f.left.english()
	case "box":
		return "necessarily, " + f.left.english()
	case "diamond":
		return "possibly, " + f.left.english()
	case "if":
		return "if " + f.left.english() + ", then " + f.right.english()
	case "and":
		return "it is the case both that " + f.left.english() + ", and that " + f.right.english()
	default:
		panic(f.op)
	}
}

func (f *gateModalOracleFormula) eval(world, frame, valuation int) bool {
	switch f.op {
	case "p":
		return valuation&(1<<world) != 0
	case "q":
		return valuation&(1<<(world+3)) != 0
	case "not":
		return !f.left.eval(world, frame, valuation)
	case "if":
		return !f.left.eval(world, frame, valuation) || f.right.eval(world, frame, valuation)
	case "and":
		return f.left.eval(world, frame, valuation) && f.right.eval(world, frame, valuation)
	case "box", "diamond":
		for next := 0; next < 3; next++ {
			if frame&(1<<(3*world+next)) == 0 {
				continue
			}
			truth := f.left.eval(next, frame, valuation)
			if f.op == "box" && !truth {
				return false
			}
			if f.op == "diamond" && truth {
				return true
			}
		}
		return f.op == "box"
	default:
		panic(f.op)
	}
}

func TestEntailmentGateV15ModalOracle(t *testing.T) {
	p, q := &gateModalOracleFormula{op: "p"}, &gateModalOracleFormula{op: "q"}
	unary := func(op string, a *gateModalOracleFormula) *gateModalOracleFormula {
		return &gateModalOracleFormula{op: op, left: a}
	}
	box := func(a *gateModalOracleFormula) *gateModalOracleFormula { return unary("box", a) }
	diamond := func(a *gateModalOracleFormula) *gateModalOracleFormula { return unary("diamond", a) }
	not := func(a *gateModalOracleFormula) *gateModalOracleFormula { return unary("not", a) }
	imp := func(a, b *gateModalOracleFormula) *gateModalOracleFormula {
		return &gateModalOracleFormula{op: "if", left: a, right: b}
	}
	and := func(a, b *gateModalOracleFormula) *gateModalOracleFormula {
		return &gateModalOracleFormula{op: "and", left: a, right: b}
	}
	for _, tc := range []struct {
		name     string
		from, to *gateModalOracleFormula
		valid    bool
	}{
		{"duality box forward", not(diamond(not(p))), box(p), true},
		{"duality box reverse", box(p), not(diamond(not(p))), true},
		{"duality diamond forward", not(box(not(p))), diamond(p), true},
		{"duality diamond reverse", diamond(p), not(box(not(p))), true},
		{"K", box(imp(p, q)), imp(box(p), box(q)), true},
		{"possibility monotonicity", box(imp(p, q)), imp(diamond(p), diamond(q)), true},
		{"conjunction forward", box(and(p, q)), and(box(p), box(q)), true},
		{"conjunction reverse", and(box(p), box(q)), box(and(p, q)), true},
		{"converse K", imp(box(p), box(q)), box(imp(p, q)), false},
		{"S4", box(p), box(box(p)), false},
		{"S5", diamond(p), box(diamond(p)), false},
		{"possibility to necessity", diamond(p), box(p), false},
		{"possible conditional", diamond(imp(p, q)), imp(diamond(p), diamond(q)), false},
		{"missing inner negation", not(diamond(p)), box(p), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			counterexamples := 0
			const reflexive = 1<<0 | 1<<4 | 1<<8
			for frame := 0; frame < 1<<9; frame++ {
				if frame&reflexive != reflexive {
					continue
				}
				for valuation := 0; valuation < 1<<6; valuation++ {
					for world := 0; world < 3; world++ {
						if tc.from.eval(world, frame, valuation) && !tc.to.eval(world, frame, valuation) {
							counterexamples++
						}
					}
				}
			}
			if (counterexamples == 0) != tc.valid {
				t.Fatalf("valid=%v counterexamples=%d", tc.valid, counterexamples)
			}
			argument := fmt.Sprintf("1. %s.\n2. Therefore, %s. [1]", tc.from.english(), tc.to.english())
			report, result := CheckEntailmentGate(argument)
			if report.Passed != tc.valid || result.Passed != tc.valid {
				t.Fatalf("valid=%v: %s %+v", tc.valid, result.Message, report)
			}
		})
	}
}
