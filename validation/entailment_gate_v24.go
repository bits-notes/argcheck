package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// folFormula is the function-free FOL= core. A $ prefix distinguishes bound
// variables from names. Predicate symbols include the verbatim prepositions.
// No search or semantic equivalence is used by the production rule checker.
type folFormula struct {
	op, symbol string
	terms      []string
	children   []*folFormula
}

func fol(op string, fs ...*folFormula) *folFormula { return &folFormula{op: op, children: fs} }

var explicitBinder = regexp.MustCompile(`\b(every|some|no) ([a-z]) is such that\b`)
var arbitraryFOLPattern = regexp.MustCompile(`^Let ([A-Za-z][A-Za-z0-9_]*) be arbitrary$`)
var folWord = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func folExplicit(s string) bool { return explicitBinder.MatchString(strings.ToLower(s)) }
func folStrip(s string) string {
	s = strings.TrimSpace(s)
	for len(s) > 1 && s[0] == '(' && s[len(s)-1] == ')' {
		depth, whole := 0, true
		for i, c := range s {
			if c == '(' {
				depth++
			}
			if c == ')' {
				depth--
			}
			if depth == 0 && i < len(s)-1 {
				whole = false
				break
			}
		}
		if !whole {
			break
		}
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}
func folSplit(s, sep string) (string, string, bool) {
	depth := 0
	for i, c := range s {
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
		}
		if depth == 0 && strings.HasPrefix(s[i:], sep) {
			return s[:i], s[i+len(sep):], true
		}
	}
	return "", "", false
}
func parseFOL(text string) (*folFormula, error) {
	s := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(text), "."))
	return parseFOLScope(s, map[string]bool{}, folExplicit(s))
}
func parseFOLScope(s string, bound map[string]bool, strict bool) (*folFormula, error) {
	s = folStrip(s)
	bad := func() (*folFormula, error) {
		return nil, fmt.Errorf("use every/some/no <variable> is such that, bind each variable once, and parenthesize compound scopes")
	}
	depth := 0
	for _, c := range s {
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
		}
		if depth < 0 {
			return bad()
		}
	}
	if depth != 0 {
		return bad()
	}
	if s == "contradiction" {
		return fol("bottom"), nil
	}
	if m := explicitBinder.FindStringSubmatchIndex(s); m != nil && m[0] == 0 {
		kind, v := s[m[2]:m[3]], s[m[4]:m[5]]
		if bound[v] {
			return bad()
		}
		next := map[string]bool{}
		for k, b := range bound {
			next[k] = b
		}
		next[v] = true
		body, err := parseFOLScope(strings.TrimSpace(s[m[1]:]), next, true)
		if err != nil {
			return nil, err
		}
		if kind == "no" {
			kind = "every"
			body = fol("not", body)
		}
		f := fol(kind, body)
		f.symbol = "$" + v
		return f, nil
	}
	if strings.HasPrefix(s, "it is not the case that ") {
		f, e := parseFOLScope(strings.TrimPrefix(s, "it is not the case that "), bound, strict)
		if e != nil {
			return nil, e
		}
		return fol("not", f), nil
	}
	if strings.HasPrefix(s, "if ") {
		a, b, ok := folSplit(s[3:], " then ")
		if !ok {
			a, b, ok = folSplit(s[3:], ", then ")
		}
		if ok {
			return folParsePair("implies", strings.TrimSuffix(a, ","), b, bound, strict)
		}
		return bad()
	}
	for _, p := range []struct{ sep, op string }{{" if and only if ", "iff"}, {" and ", "and"}, {" or ", "or"}} {
		t := s
		if p.op == "or" {
			t = strings.TrimPrefix(t, "either ")
		}
		if a, b, ok := folSplit(t, p.sep); ok {
			return folParsePair(p.op, a, b, bound, strict)
		}
	}
	term := func(t string) (string, bool) {
		t = strings.TrimSpace(t)
		if !folWord.MatchString(t) {
			return "", false
		}
		if bound[t] {
			return "$" + t, true
		}
		if strict && len(t) == 1 {
			return "", false
		}
		return t, true
	}
	if a, b, ok := strings.Cut(s, " is identical to "); ok {
		x, xok := term(a)
		y, yok := term(b)
		if !xok || !yok {
			return bad()
		}
		return &folFormula{op: "equal", terms: []string{x, y}}, nil
	}
	if a, b, ok := strings.Cut(s, " is "); ok {
		x, ok := term(a)
		if !ok || !folWord.MatchString(b) {
			return bad()
		}
		return &folFormula{op: "atom", symbol: "is/" + b, terms: []string{x}}, nil
	}
	w := strings.Fields(s)
	if len(w) < 3 {
		return bad()
	}
	a, ok := term(w[0])
	if !ok || !relationVerb(w[1]) || !folWord.MatchString(w[1]) {
		return bad()
	}
	// The readable double-object book example is an existential restrictor.
	if len(w) == 5 && w[1] == "gave" && (w[3] == "a" || w[3] == "an") && folWord.MatchString(w[4]) {
		b, ok := term(w[2])
		if !ok {
			return bad()
		}
		v := ""
		for _, letter := range "abcdefghijklmnopqrstuvwxyz" {
			if !bound[string(letter)] {
				v = "$" + string(letter)
				break
			}
		}
		if v == "" {
			return bad()
		}
		atom := &folFormula{op: "atom", symbol: "gave/to", terms: []string{a, v, b}}
		q := fol("some", fol("and", &folFormula{op: "atom", symbol: "is/" + w[4], terms: []string{v}}, atom))
		q.symbol = v
		return q, nil
	}
	b, ok := term(w[2])
	if !ok {
		return bad()
	}
	ts := []string{a, b}
	symbol := w[1]
	for i := 3; i < len(w); i += 2 {
		if i+1 >= len(w) || (w[i] != "to" && w[i] != "from" && w[i] != "with") {
			return bad()
		}
		t, ok := term(w[i+1])
		if !ok {
			return bad()
		}
		symbol += "/" + w[i]
		ts = append(ts, t)
	}
	if len(ts) > 4 {
		return bad()
	}
	return &folFormula{op: "atom", symbol: symbol, terms: ts}, nil
}
func folParsePair(op, a, b string, scope map[string]bool, strict bool) (*folFormula, error) {
	x, e := parseFOLScope(a, scope, strict)
	if e != nil {
		return nil, e
	}
	y, e := parseFOLScope(b, scope, strict)
	if e != nil {
		return nil, e
	}
	return fol(op, x, y), nil
}

// Render uses ordinary atoms and connectives; only unrestricted quantifiers
// need the explicit-variable escape hatch. Parentheses fix every body boundary.
func (f *folFormula) Render() string { return folRenderNames(f, nil, 0).renderSurface() }

// Canonical fresh binders also handle ASTs with shadowing when a closed
// formula is inserted beneath another quantifier. The surface forbids shadowing.
func folRenderNames(f *folFormula, env map[string]string, depth int) *folFormula {
	n := *f
	n.terms = append([]string{}, f.terms...)
	n.children = nil
	for i, t := range n.terms {
		if v, ok := env[t]; ok {
			n.terms[i] = v
		}
	}
	if f.op == "every" || f.op == "some" {
		if depth >= 26 {
			panic("FOL renderer requires more than 26 simultaneous binders")
		}
		next := map[string]string{}
		for k, v := range env {
			next[k] = v
		}
		n.symbol = "$" + string("xyzuvwabcdefghijklmnopqrst"[depth])
		next[f.symbol] = n.symbol
		n.children = []*folFormula{folRenderNames(f.children[0], next, depth+1)}
	} else {
		for _, c := range f.children {
			n.children = append(n.children, folRenderNames(c, env, depth))
		}
	}
	return &n
}
func (f *folFormula) renderSurface() string {
	term := func(t string) string { return strings.TrimPrefix(t, "$") }
	wrap := func(c *folFormula) string { return "(" + c.renderSurface() + ")" }
	switch f.op {
	case "bottom":
		return "contradiction"
	case "atom":
		p := strings.Split(f.symbol, "/")
		if len(f.terms) == 1 {
			return term(f.terms[0]) + " is " + p[len(p)-1]
		}
		s := term(f.terms[0]) + " " + p[0] + " " + term(f.terms[1])
		for i := 2; i < len(f.terms); i++ {
			s += " " + p[i-1] + " " + term(f.terms[i])
		}
		return s
	case "equal":
		return term(f.terms[0]) + " is identical to " + term(f.terms[1])
	case "not":
		return "it is not the case that " + wrap(f.children[0])
	case "and":
		return wrap(f.children[0]) + " and " + wrap(f.children[1])
	case "or":
		return "either " + wrap(f.children[0]) + " or " + wrap(f.children[1])
	case "implies":
		return "if " + wrap(f.children[0]) + " then " + wrap(f.children[1])
	case "iff":
		return wrap(f.children[0]) + " if and only if " + wrap(f.children[1])
	case "every", "some":
		return f.op + " " + term(f.symbol) + " is such that " + wrap(f.children[0])
	}
	panic("unknown FOL operator")
}
func folKey(f *folFormula, env map[string]string, depth int) string {
	if f == nil {
		return "nil"
	}
	s := f.op + ":" + f.symbol
	if f.op == "every" || f.op == "some" {
		next := map[string]string{}
		for k, v := range env {
			next[k] = v
		}
		next[f.symbol] = fmt.Sprint(depth)
		return f.op + "(" + folKey(f.children[0], next, depth+1) + ")"
	}
	for _, t := range f.terms {
		if v, ok := env[t]; ok {
			s += "/$" + v
		} else {
			s += "/" + t
		}
	}
	for _, c := range f.children {
		s += "(" + folKey(c, env, depth) + ")"
	}
	return s
}
func folSame(a, b *folFormula) bool { return folKey(a, nil, 0) == folKey(b, nil, 0) }
func folSub(f *folFormula, from, to string) *folFormula {
	n := *f
	n.terms = append([]string{}, f.terms...)
	n.children = nil
	for i, t := range n.terms {
		if t == from {
			n.terms[i] = to
		}
	}
	for _, c := range f.children {
		n.children = append(n.children, folSub(c, from, to))
	}
	return &n
}
func folNames(f *folFormula) []string {
	var out []string
	for _, t := range f.terms {
		if !strings.HasPrefix(t, "$") {
			out = append(out, t)
		}
	}
	for _, c := range f.children {
		out = append(out, folNames(c)...)
	}
	return out
}
func folInstance(q, instance *folFormula) string {
	for _, name := range append(folNames(instance), "witness") {
		if folSame(folSub(q.children[0], q.symbol, name), instance) {
			return name
		}
	}
	return ""
}
func folRule(ps []*folFormula, c *folFormula) string {
	if len(ps) == 1 {
		p := ps[0]
		if folSame(p, c) {
			return "reiteration"
		}
		if p.op == "and" && (folSame(p.children[0], c) || folSame(p.children[1], c)) {
			return "conjunction elimination"
		}
		if c.op == "or" && (folSame(p, c.children[0]) || folSame(p, c.children[1])) {
			return "disjunction introduction"
		}
		if p.op == "not" && p.children[0].op == "not" && folSame(p.children[0].children[0], c) {
			return "double negation"
		}
		if c.op == "not" && c.children[0].op == "not" && folSame(c.children[0].children[0], p) {
			return "double negation introduction"
		}
		if p.op == "bottom" {
			return "contradiction elimination"
		}
		if p.op == "every" && folInstance(p, c) != "" {
			return "universal instantiation"
		}
		if c.op == "some" && folInstance(c, p) != "" {
			return "existential introduction"
		}
		if c.op == "equal" && c.terms[0] == c.terms[1] {
			for _, name := range folNames(p) {
				if name == c.terms[0] {
					return "identity reflexivity"
				}
			}
		}
		if p.op == "iff" && c.op == "implies" && ((folSame(p.children[0], c.children[0]) && folSame(p.children[1], c.children[1])) || (folSame(p.children[1], c.children[0]) && folSame(p.children[0], c.children[1]))) {
			return "biconditional elimination"
		}
	}
	if len(ps) == 2 {
		for i := 0; i < 2; i++ {
			p, q := ps[i], ps[1-i]
			if c.op == "and" && folSame(p, c.children[0]) && folSame(q, c.children[1]) {
				return "conjunction introduction"
			}
			if p.op == "implies" && folSame(p.children[0], q) && folSame(p.children[1], c) {
				return "modus ponens"
			}
			if c.op == "bottom" && p.op == "not" && folSame(p.children[0], q) {
				return "contradiction introduction"
			}
			if c.op == "iff" && p.op == "implies" && q.op == "implies" && folSame(p.children[0], c.children[0]) && folSame(p.children[1], c.children[1]) && folSame(q.children[0], c.children[1]) && folSame(q.children[1], c.children[0]) {
				return "biconditional introduction"
			}
			if p.op == "equal" {
				for j := 0; j < 2; j++ {
					if !folSame(q, c) && folIdentityMatch(q, c, p.terms[j], p.terms[1-j]) {
						return "identity substitution"
					}
				}
			}
		}
	}
	return ""
}
func gateFOL(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	r := EntailmentGateStepReport{Step: c.Number, Depth: c.Depth, Citations: c.Citations, Ranges: c.Ranges, NearestForm: "explicit-variable FOL", Mismatch: "no one-step FOL rule matches; bind variables inside is such that and parenthesize scopes"}
	accept := func(s string) EntailmentGateStepReport {
		r.Passed = true
		r.Form = "FOL " + s
		r.NearestForm = ""
		r.Mismatch = ""
		return r
	}
	parseClosed := func(text string) (*folFormula, error) {
		return parseFOLScope(strings.ToLower(strings.TrimSuffix(strings.TrimSpace(text), ".")), map[string]bool{}, true)
	}
	target, e := parseClosed(c.Statement)
	if e != nil {
		r.Mismatch = e.Error()
		return r
	}
	parse := func(s NumberedArgumentStep) *folFormula {
		if s.Kind == numberedArgumentStepArbitrary {
			if m := arbitraryFOLPattern.FindStringSubmatch(s.Statement); m != nil {
				name := strings.ToLower(m[1])
				return &folFormula{op: "equal", terms: []string{name, name}}
			}
		}
		f, _ := parseClosed(s.Statement)
		return f
	}
	var ps []*folFormula
	for _, id := range c.Citations {
		f := parse(steps[id])
		if f == nil {
			return r
		}
		ps = append(ps, f)
	}
	if len(c.Ranges) == 0 {
		if form := folRule(ps, target); form != "" {
			return accept(form)
		}
		return r
	}
	if len(c.Ranges) == 1 {
		span := c.Ranges[0]
		first, last := steps[span[0]], steps[span[1]]
		end := parse(last)
		if end == nil {
			return r
		}
		if len(ps) == 0 && first.Kind == numberedArgumentStepArbitrary {
			m := arbitraryFOLPattern.FindStringSubmatch(first.Statement)
			if m != nil && target.op == "every" && folSame(folSub(target.children[0], target.symbol, strings.ToLower(m[1])), end) && fitchFreshName(m[1], first.Number, c, steps) == "" {
				return accept("universal generalization")
			}
		}
		start := parse(first)
		if start == nil || first.Kind != numberedArgumentStepAssumption {
			return r
		}
		if len(ps) == 0 {
			if target.op == "implies" && folSame(start, target.children[0]) && folSame(end, target.children[1]) {
				return accept("conditional proof")
			}
			if end.op == "bottom" && ((target.op == "not" && folSame(start, target.children[0])) || (start.op == "not" && folSame(start.children[0], target))) {
				return accept("reductio")
			}
		}
		if len(ps) == 1 && ps[0].op == "some" && folSame(end, target) {
			// A vacuous binder has no witness occurrence to recover. Choose
			// a genuinely fresh eigenname rather than an unrelated constant
			// already present in its body.
			if folSame(ps[0].children[0], start) {
				name := "eigenwitness"
				for fitchFreshName(name, first.Number, c, steps) != "" {
					name += "fresh"
				}
				if folSame(folSub(ps[0].children[0], ps[0].symbol, name), start) {
					return accept("existential elimination")
				}
			}
			name := folInstance(ps[0], start)
			if name != "" && fitchFreshName(name, first.Number, c, steps) == "" {
				return accept("existential elimination")
			}
		}
	}
	if len(c.Ranges) == 2 && len(ps) == 1 && ps[0].op == "or" {
		for swap := 0; swap < 2; swap++ {
			ok := true
			for i, span := range c.Ranges {
				first, last := steps[span[0]], steps[span[1]]
				if first.Kind != numberedArgumentStepAssumption || !folSame(parse(first), ps[0].children[(i+swap)%2]) || !folSame(parse(last), target) {
					ok = false
				}
			}
			if ok {
				return accept("proof by cases")
			}
		}
	}
	return r
}

// Equality elimination may replace any selection of free occurrences. Bound
// variables are alpha-normalized before comparing the unchanged structure.
func folIdentityMatch(a, b *folFormula, from, to string) bool {
	if a.op != b.op || len(a.terms) != len(b.terms) || len(a.children) != len(b.children) {
		return false
	}
	if a.op == "every" || a.op == "some" {
		// Rename both bound variables to the same fresh internal marker.
		marker := "$identity/" + a.symbol + "/" + b.symbol
		return folIdentityMatch(folSub(a.children[0], a.symbol, marker), folSub(b.children[0], b.symbol, marker), from, to)
	}
	if a.symbol != b.symbol {
		return false
	}
	for i, t := range a.terms {
		if t != b.terms[i] && !(t == from && b.terms[i] == to) {
			return false
		}
	}
	for i, c := range a.children {
		if !folIdentityMatch(c, b.children[i], from, to) {
			return false
		}
	}
	return true
}
