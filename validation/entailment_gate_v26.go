package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Recognize numeral tokens, not adverbs such as "at least as" or "exactly what".
// Larger numeral words are recognized only to give the bounded-surface diagnostic.
const countingNumeral = `(?:[0-9]+|zero|one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|thirteen|fourteen|fifteen|sixteen|seventeen|eighteen|nineteen|twenty|thirty|forty|fifty|sixty|seventy|eighty|ninety|hundred|thousand|million|billion|trillion)`

var countingPrefix = regexp.MustCompile(`^(at least|at most|exactly) (` + countingNumeral + `) (.+)$`)
var countingMarker = regexp.MustCompile(`\b(at least|at most|exactly) ` + countingNumeral + ` [a-z]|\bother than\b|^only `)

// Counting is a conservative, private surface extension. Its only semantic
// representation is ordinary FOL with equality, never an opaque count atom.
func countingActive(s string) bool                   { return countingMarker.MatchString(strings.ToLower(s)) }
func countQ(op, v string, f *folFormula) *folFormula { q := fol(op, f); q.symbol = v; return q }
func countAnd(fs ...*folFormula) *folFormula {
	f := fs[0]
	for _, g := range fs[1:] {
		f = fol("and", f, g)
	}
	return f
}
func countAtLeast(n int, body *folFormula) *folFormula {
	var fs []*folFormula
	for i := 0; i < n; i++ {
		v := fmt.Sprintf("$count%d", i)
		fs = append(fs, folSub(body, "$count", v))
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			fs = append(fs, fol("not", &folFormula{op: "equal", terms: []string{fmt.Sprintf("$count%d", i), fmt.Sprintf("$count%d", j)}}))
		}
	}
	f := countAnd(fs...)
	for i := n - 1; i >= 0; i-- {
		f = countQ("some", fmt.Sprintf("$count%d", i), f)
	}
	return f
}
func countDefinition(kind string, n int, body *folFormula) *folFormula {
	lo := countAtLeast(n, body)
	hi := fol("not", countAtLeast(n+1, body))
	switch kind {
	case "at least":
		return lo
	case "at most":
		return hi
	default:
		return fol("and", lo, hi)
	}
}
func countAtom(s string) (*folFormula, error) {
	// The two new lexical verbs are private: old candidate sets are unchanged.
	for _, verb := range []string{" signed ", " wrote "} {
		if a, b, ok := strings.Cut(s, verb); ok && !strings.Contains(b, verb) {
			if !relationWords(a) || !relationWords(b) {
				break
			}
			return &folFormula{op: "atom", symbol: strings.TrimSpace(verb), terms: []string{legacyFOLTerm(a), legacyFOLTerm(b)}}, nil
		}
	}
	fs, why := legacyFOLReadings(s)
	if why != "" || len(fs) != 1 {
		return nil, fmt.Errorf("counting requires one unambiguous predicate reading")
	}
	return fs[0], nil
}
func countPredicate(s string) (*folFormula, error) {
	words := strings.Fields(s)
	end := -1
	relative := false
	seen := false
	for i, w := range words {
		if w == "that" || w == "who" || w == "which" {
			if relative {
				return nil, fmt.Errorf("nested relative clause")
			}
			relative = true
		}
		if w == "is" || w == "are" || relativeVerb(w) || w == "signed" || w == "wrote" {
			if !relative || seen {
				end = i
				break
			}
			seen = true
		}
	}
	if end < 0 && !relative && relationWords(s) {
		class := s
		if strings.HasSuffix(class, "s") && !strings.HasSuffix(class, "ss") {
			class = strings.TrimSuffix(class, "s")
		}
		return &folFormula{op: "atom", symbol: "is/" + legacyFOLTerm(class), terms: []string{"$count"}}, nil
	}
	if end < 1 {
		return nil, fmt.Errorf("counting requires a class and predicate")
	}
	np := words[:end]
	classEnd := len(np)
	for i, w := range np {
		if w == "that" || w == "who" || w == "which" {
			classEnd = i
			break
		}
	}
	class := strings.Join(np[:classEnd], " ")
	if !relationWords(class) {
		return nil, fmt.Errorf("missing counting class")
	}
	// Controlled plural spelling, only on the newly recognized count surface.
	if strings.HasSuffix(class, "s") && !strings.HasSuffix(class, "ss") {
		class = strings.TrimSuffix(class, "s")
	}
	guard := &folFormula{op: "atom", symbol: "is/" + legacyFOLTerm(class), terms: []string{"$count"}}
	atom := func(text string) (*folFormula, error) {
		f, e := countAtom(text)
		if e != nil {
			return nil, e
		}
		return folSub(f, "countwitness", "$count"), nil
	}
	if classEnd < len(np) {
		tail := np[classEnd+1:]
		if len(tail) == 0 {
			return nil, fmt.Errorf("empty relative")
		}
		text := "countwitness " + strings.Join(tail, " ")
		if !(relativeVerb(tail[0]) || tail[0] == "signed" || tail[0] == "wrote") {
			text = strings.Join(tail, " ") + " countwitness"
		}
		f, e := atom(text)
		if e != nil {
			return nil, e
		}
		guard = fol("and", guard, f)
	}
	tail := append([]string{}, words[end:]...)
	if tail[0] == "are" {
		tail[0] = "is"
	}
	body, e := atom("countwitness " + strings.Join(tail, " "))
	if e != nil {
		return nil, e
	}
	if folSame(guard, body) {
		return guard, nil
	}
	return fol("and", guard, body), nil
}
func countParse(text string) (*folFormula, error) {
	s := folStrip(normalizeClause(text))
	if m := countingPrefix.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[2])
		for i, w := range []string{"one", "two", "three", "four", "five"} {
			if m[2] == w {
				n = i + 1
			}
		}
		if n < 1 || n > 5 {
			return nil, fmt.Errorf("counting bound is 5; use one to five or digits 1 to 5")
		}
		body, e := countPredicate(m[3])
		if e != nil {
			return nil, e
		}
		return countDefinition(m[1], n, body), nil
	}
	if strings.HasPrefix(s, "only ") {
		rest := strings.TrimPrefix(s, "only ")
		a, b, ok := strings.Cut(rest, " is ")
		if !ok {
			return nil, fmt.Errorf("only requires a named copular instance")
		}
		instance, e := countAtom(rest)
		if e != nil {
			return nil, e
		}
		p := strings.TrimPrefix(strings.TrimPrefix(b, "a "), "an ")
		body := &folFormula{op: "atom", symbol: "is/" + legacyFOLTerm(p), terms: []string{"$count"}}
		return fol("and", instance, countUnique(body, legacyFOLTerm(a))), nil
	}
	if strings.HasPrefix(s, "no ") && strings.Contains(s, " other than ") {
		cl, rest, _ := strings.Cut(strings.TrimPrefix(s, "no "), " other than ")
		words := strings.Fields(rest)
		at := -1
		for i, w := range words {
			if w == "is" || relativeVerb(w) || w == "signed" || w == "wrote" {
				at = i
				break
			}
		}
		if at < 1 {
			return nil, fmt.Errorf("other than requires a name and predicate")
		}
		name := legacyFOLTerm(strings.Join(words[:at], " "))
		body, e := countPredicate(cl + " " + strings.Join(words[at:], " "))
		if e != nil {
			return nil, e
		}
		return countUnique(body, name), nil
	}
	if folExplicit(s) {
		return parseFOL(s)
	}
	if strings.HasPrefix(s, "it is not the case that ") {
		f, e := countParse(strings.TrimPrefix(s, "it is not the case that "))
		if e != nil {
			return nil, e
		}
		return fol("not", f), nil
	}
	if strings.HasPrefix(s, "if ") {
		a, b, ok := folSplit(s[3:], " then ")
		if ok {
			return countPair("implies", strings.TrimSuffix(a, ","), b)
		}
	}
	for _, p := range []struct{ sep, op string }{{" if and only if ", "iff"}, {" and ", "and"}, {" or ", "or"}} {
		if a, b, ok := folSplit(s, p.sep); ok {
			return countPair(p.op, a, b)
		}
	}
	if strings.HasPrefix(s, "some ") {
		if body, e := countPredicate(strings.TrimPrefix(s, "some ")); e == nil {
			return countQ("some", "$count", body), nil
		}
	}
	return countAtom(s)
}
func countPair(op, a, b string) (*folFormula, error) {
	x, e := countParse(a)
	if e != nil {
		return nil, e
	}
	y, e := countParse(b)
	if e != nil {
		return nil, e
	}
	return fol(op, x, y), nil
}
func countUnique(body *folFormula, name string) *folFormula {
	return countQ("every", "$count", fol("implies", body, &folFormula{op: "equal", terms: []string{"$count", name}}))
}
func countLeaves(f *folFormula) []*folFormula {
	if f.op == "and" {
		return append(countLeaves(f.children[0]), countLeaves(f.children[1])...)
	}
	return []*folFormula{f}
}
func countSameBody(a, b *folFormula) bool {
	as, bs := countLeaves(a), countLeaves(b)
	for _, x := range as {
		found := false
		for _, y := range bs {
			found = found || folSame(x, y)
		}
		if !found {
			return false
		}
	}
	for _, y := range bs {
		found := false
		for _, x := range as {
			found = found || folSame(x, y)
		}
		if !found {
			return false
		}
	}
	return true
}

// Recognize the identity definition, including explicit-variable input. Surface
// wording never licenses a rule independently of these AST comparisons.
func countInfo(f *folFormula) (string, int, *folFormula) {
	kind := "at least"
	base := f
	if f.op == "and" {
		kind = "exactly"
		base = f.children[0]
	}
	if f.op == "not" {
		kind = "at most"
		base = f.children[0]
	}
	n := 0
	var v string
	for base.op == "some" {
		if n == 0 {
			v = base.symbol
		}
		n++
		base = base.children[0]
	}
	if n == 0 || n > 6 {
		return "", 0, nil
	}
	if kind == "at most" {
		n--
	}
	if n < 1 || n > 5 {
		return "", 0, nil
	}
	// The leftmost conjuncts are the first witness's predicate. Try all
	// conjunction prefixes; verify the entire expanded definition afterwards.
	var candidates []*folFormula
	var walk func(*folFormula)
	walk = func(g *folFormula) {
		candidates = append(candidates, g)
		if g.op == "and" {
			walk(g.children[0])
		}
	}
	walk(base)
	for _, g := range candidates {
		body := folSub(g, v, "$count")
		if folSame(f, countDefinition(kind, n, body)) {
			return kind, n, body
		}
	}
	return "", 0, nil
}
func countingRule(ps []*folFormula, c *folFormula) string {
	if form := folRule(ps, c); form != "" {
		return "definition"
	}
	// The counted witness may be established by separate cited ground facts.
	// Match every remaining premise to the full instantiated body: neither
	// missing membership nor unrelated extra citations license uniqueness.
	if len(ps) >= 2 && c.op == "every" && c.children[0].op == "implies" {
		b := c.children[0]
		if eq := b.children[1]; eq.op == "equal" && eq.terms[0] == c.symbol {
			for i, p := range ps {
				pk, pn, pb := countInfo(p)
				if (pk != "exactly" && pk != "at most") || pn != 1 {
					continue
				}
				ground := append([]*folFormula{}, ps[:i]...)
				ground = append(ground, ps[i+1:]...)
				if countSameBody(pb, folSub(b.children[0], c.symbol, "$count")) &&
					countSameBody(folSub(pb, "$count", eq.terms[1]), countAnd(ground...)) {
					return "uniqueness"
				}
			}
		}
	}
	ck, cn, cb := countInfo(c)
	if len(ps) == 1 {
		pk, pn, pb := countInfo(ps[0])
		if pk == "exactly" && pn == cn && (ck == "at least" || ck == "at most") && countSameBody(pb, cb) {
			return "elimination"
		}
		// only a is F uses the same uniqueness definition as no F other than a is F.
		if folSame(ps[0], c) {
			return "definition"
		}
	}
	if len(ps) == 2 {
		for i := 0; i < 2; i++ {
			p, q := ps[i], ps[1-i]
			pk, pn, pb := countInfo(p)
			qk, qn, qb := countInfo(q)
			if pk == "at least" && qk == "at most" && countSameBody(pb, qb) {
				if cn == pn && pn == qn && ck == "exactly" && countSameBody(pb, cb) {
					return "introduction"
				}
				if pn > qn && c.op == "bottom" {
					return "contradiction"
				}
			}
			// A class with more members than its B subset has a member outside B.
			if (pk == "at least" || pk == "exactly") && (qk == "at most" || qk == "exactly") && pn > qn && c.op == "some" && qb.op == "and" {
				b := folSub(c.children[0], c.symbol, "$count")
				if b.op == "and" && b.children[1].op == "not" && countSameBody(pb, qb.children[0]) && countSameBody(pb, b.children[0]) && folSame(qb.children[1], b.children[1].children[0]) {
					return "subset"
				}
			}
		}
	}
	return ""
}
func gateCounting(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	r := EntailmentGateStepReport{Step: c.Number, Depth: c.Depth, Citations: c.Citations, Ranges: c.Ranges, NearestForm: "counting definition", Mismatch: "requires the identity-based counting definition and matching predicates"}
	target, e := countParse(c.Statement)
	if e != nil {
		r.Mismatch = e.Error()
		return r
	}
	var ps []*folFormula
	for _, id := range c.Citations {
		p, e := countParse(steps[id].Statement)
		if e != nil {
			r.Mismatch = e.Error()
			return r
		}
		ps = append(ps, p)
	}
	if len(c.Ranges) == 0 {
		if form := countingRule(ps, target); form != "" {
			r.Passed = true
			r.Form = "counting " + form
			r.NearestForm = ""
			r.Mismatch = ""
		}
		return r
	}
	// Reuse the established Fitch checker on lossless explicit definitions;
	// all eigenname and scope checks remain in force.
	translated := map[int]NumberedArgumentStep{}
	for id, s := range steps {
		if s.Kind != numberedArgumentStepArbitrary {
			f, e := countParse(s.Statement)
			if e == nil {
				s.Statement = f.Render()
			}
		}
		translated[id] = s
	}
	c.Statement = target.Render()
	return gateFOL(c, translated)
}
