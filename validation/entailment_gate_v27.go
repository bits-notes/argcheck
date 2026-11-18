package validation

import (
	"fmt"
	"strings"
	"unicode"
)

// Descriptions are a private fallback. Keep the v26 verdict and report whenever
// it succeeds, and keep its diagnostic whenever this extension declines.
func gateConclusionStepV27(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	old := gateConclusionStepV26(c, steps)
	if old.Passed {
		return old
	}
	if next := gateDescriptions(c, steps); next.Passed {
		return next
	}
	return old
}

// Convert only the controlled copular of-phrase to its ordered relation. The
// relation spelling and object are literal: no paraphrase or morphology bridge.
func descriptionRelations(f *folFormula) *folFormula {
	g := *f
	g.children = nil
	for _, child := range f.children {
		g.children = append(g.children, descriptionRelations(child))
	}
	if f.op == "atom" && len(f.terms) == 1 && strings.HasPrefix(f.symbol, "is/") {
		if a, b, ok := strings.Cut(strings.TrimPrefix(f.symbol, "is/"), " of "); ok {
			g.symbol = "is/" + a + "/of"
			g.terms = []string{f.terms[0], b}
		}
	}
	return &g
}

func descriptionParts(s string) (string, string, bool) {
	if !strings.HasPrefix(s, "the ") {
		return "", "", false
	}
	words := strings.Fields(strings.TrimPrefix(s, "the "))
	relative, seen := false, false
	for i, w := range words {
		if w == "that" || w == "who" || w == "which" {
			if relative {
				return "", "", false
			}
			relative = true
		}
		if w == "is" || relativeVerb(w) || w == "wrote" || w == "signed" {
			if !relative || seen {
				return strings.Join(words[:i], " "), strings.Join(words[i:], " "), i > 0
			}
			seen = true
		}
	}
	return "", "", false
}
func descriptionDefinition(g, p *folFormula) *folFormula {
	return fol("and", countQ("some", "$count", fol("and", g, p)), countDefinition("at most", 1, g))
}
func descriptionOpaque(s string) bool {
	// Relative markers are structural only in a leading description restrictor.
	if cl, tail, ok := descriptionParts(s); ok {
		// A restrictor that ends inside an embedded clause ("the observer
		// reckons alpha" + "is red") is not a description, whatever the verb.
		if clauseInsideTerm("the " + cl) {
			return true
		}
		for _, marker := range []string{" that ", " who ", " which "} {
			cl = strings.ReplaceAll(cl, marker, " ")
		}
		return opaqueContextWords.MatchString(cl + " " + tail)
	}
	return opaqueContextWords.MatchString(s)
}
func descriptionName(text string) string {
	text = strings.TrimSuffix(strings.TrimSpace(text), ".")
	i := strings.LastIndex(strings.ToLower(text), " is ")
	if i < 0 {
		return ""
	}
	rhs := strings.TrimSpace(text[i+4:])
	if rhs != "" && unicode.IsUpper([]rune(rhs)[0]) && relationWords(strings.ToLower(rhs)) {
		return strings.ToLower(rhs)
	}
	return ""
}
func descriptionGuard(cl string) (*folFormula, error) {
	if !relativeMarker(cl) {
		f, e := countAtom("countwitness is a " + cl)
		if e != nil {
			return nil, e
		}
		return folSub(f, "countwitness", "$count"), nil
	}
	facts, ok := relativeNP(cl, "countwitness")
	if !ok {
		return nil, fmt.Errorf("unsupported relative restrictor")
	}
	var gs []*folFormula
	for _, fact := range facts {
		f, e := countAtom(fact)
		if e != nil {
			return nil, e
		}
		gs = append(gs, folSub(f, "countwitness", "$count"))
	}
	return countAnd(gs...), nil
}
func descriptionParse(text string) ([]*folFormula, error) {
	s := folStrip(normalizeClause(text))
	original := folStrip(normalizeClausePreservingCase(text))
	if folExplicit(s) {
		f, e := parseFOL(s)
		if e != nil {
			return nil, e
		}
		return []*folFormula{f}, nil
	}
	if strings.HasPrefix(s, "it is not the case that ") {
		fs, e := descriptionParse(original[len("it is not the case that "):])
		if e != nil {
			return nil, e
		}
		for i, f := range fs {
			fs[i] = fol("not", f)
		}
		return fs, nil
	}
	if strings.HasPrefix(s, "if ") {
		if a, b, ok := folSplit(s[3:], " then "); ok {
			return descriptionPair("implies", strings.TrimSuffix(original[3:3+len(a)], ","), original[len(original)-len(b):])
		}
	}
	// This fallback has no precedence convention for readable compounds.
	// In particular, splitting AND before OR changes either P or Q and R
	// into (P or Q) and R. Leave those surfaces to the legacy gate; compound
	// description formulas must delimit both operands with parentheses.
	for _, marker := range []string{"either ", "both that ", " if and only if "} {
		if strings.Contains(s, marker) {
			return nil, fmt.Errorf("unsupported description connective scaffold")
		}
	}
	for _, pair := range []struct{ sep, op string }{{" and ", "and"}, {" or ", "or"}} {
		if a, b, ok := folSplit(s, pair.sep); ok {
			if folStrip(a) == a || folStrip(b) == b {
				return nil, fmt.Errorf("parenthesize description compound operands")
			}
			return descriptionPair(pair.op, original[:len(a)], original[len(original)-len(b):])
		}
	}
	// Opaque clauses remain whole propositions, with no replaceable name slots.
	for _, prefix := range []string{"an ", "a "} {
		if strings.HasPrefix(s, prefix) {
			fs, e := descriptionParse("The " + original[len(prefix):])
			if e != nil || len(fs) != 1 {
				return nil, fmt.Errorf("unsupported indefinite description")
			}
			if g, _ := descriptionInfo(fs[0]); g != nil {
				return []*folFormula{fs[0].children[0]}, nil
			}
			return nil, fmt.Errorf("unsupported indefinite description")
		}
	}
	if descriptionOpaque(s) {
		return []*folFormula{{op: "atom", symbol: "opaque/" + s}}, nil
	}
	if cl, tail, ok := descriptionParts(s); ok {
		g, e := descriptionGuard(cl)
		if e != nil {
			return nil, e
		}
		if name := descriptionName(text); name != "" {
			tail = "is identical to " + name
		}
		p, e := descriptionPredication(tail)
		if e != nil {
			return nil, e
		}
		g = descriptionRelations(g)
		if strings.HasPrefix(tail, "is not ") {
			positive, e := descriptionPredication("is " + strings.TrimPrefix(tail, "is not "))
			if e != nil {
				return nil, e
			}
			return []*folFormula{descriptionDefinition(g, p), fol("not", descriptionDefinition(g, positive))}, nil
		}
		return []*folFormula{descriptionDefinition(g, p)}, nil
	}
	// Bare-name copulas in this controlled extension denote equality. Ordinary
	// property words retain the legacy predicate analysis.
	if name := descriptionName(text); name != "" {
		if a, _, ok := strings.Cut(s, " is "); ok && !strings.Contains(a, " ") {
			return []*folFormula{{op: "equal", terms: []string{a, name}}}, nil
		}
	}
	f, e := countParse(s)
	if e != nil {
		return nil, e
	}
	return []*folFormula{descriptionRelations(f)}, nil
}
func descriptionPredication(tail string) (*folFormula, error) {
	// Capitalized bare-name copulas have already become explicit identities.
	f, e := countAtom("countwitness " + tail)
	if e != nil {
		return nil, e
	}
	// A successful legacy parse can still be an opaque zero-place clause.
	// It is a predication of the description only when the witness occupies
	// a real term slot in an atom (or its negation), never a connective tree.
	atom := f
	if atom.op == "not" {
		atom = atom.children[0]
	}
	if (atom.op != "atom" && atom.op != "equal") || folSame(f, folSub(f, "countwitness", "$count")) {
		return nil, fmt.Errorf("description requires a predicate of its witness")
	}
	return descriptionRelations(folSub(f, "countwitness", "$count")), nil
}
func descriptionPair(op, a, b string) ([]*folFormula, error) {
	as, e := descriptionParse(a)
	if e != nil {
		return nil, e
	}
	bs, e := descriptionParse(b)
	if e != nil {
		return nil, e
	}
	if len(as)*len(bs) > 16 {
		return nil, fmt.Errorf("too many description scope readings")
	}
	var out []*folFormula
	for _, x := range as {
		for _, y := range bs {
			out = append(out, fol(op, x, y))
		}
	}
	return out, nil
}

// Recover only our exact Russellian definition; arbitrary conjunctions cannot
// acquire description rules merely by having a similar surface spelling.
func descriptionInfo(f *folFormula) (g, p *folFormula) {
	if f.op != "and" || f.children[0].op != "some" {
		return nil, nil
	}
	ex := f.children[0]
	body := folSub(ex.children[0], ex.symbol, "$count")
	if body.op != "and" {
		return nil, nil
	}
	g, p = body.children[0], body.children[1]
	if !folSame(f.children[1], countDefinition("at most", 1, g)) {
		return nil, nil
	}
	return g, p
}
func descriptionRule(ps []*folFormula, c *folFormula) string {
	if form := folRule(ps, c); form != "" {
		return "definition"
	}
	if len(ps) == 1 {
		g, p := descriptionInfo(ps[0])
		if g != nil && p.op == "not" && c.op == "not" {
			h, q := descriptionInfo(c.children[0])
			if h != nil && folSame(g, h) && folSame(p.children[0], q) {
				return "negation"
			}
		}
		if g != nil {
			for _, part := range []*folFormula{countQ("some", "$count", g), countQ("some", "$count", p), countDefinition("exactly", 1, g)} {
				if folSame(c, part) {
					return "elimination"
				}
			}
			if p.op == "equal" && p.terms[0] == "$count" && folSame(c, folSub(g, "$count", p.terms[1])) {
				return "elimination"
			}
		}
	}
	if len(ps) == 2 {
		for i := 0; i < 2; i++ {
			g, p := descriptionInfo(ps[i])
			q := ps[1-i]
			if g != nil && p.op == "equal" && p.terms[0] == "$count" {
				name := p.terms[1]
				h, r := descriptionInfo(q)
				if q.op == "not" {
					ng, np := descriptionInfo(q.children[0])
					if ng != nil && folSame(g, ng) && folSame(c, fol("not", folSub(np, "$count", name))) {
						return "negation"
					}
				}
				if h != nil && folSame(g, h) && folSame(c, folSub(r, "$count", name)) {
					return "substitution"
				}
				for _, other := range folNames(q) {
					if folSame(q, folSub(g, "$count", other)) && c.op == "equal" && ((c.terms[0] == other && c.terms[1] == name) || (c.terms[1] == other && c.terms[0] == name)) {
						return "uniqueness"
					}
				}
			}
			cg, cp := descriptionInfo(c)
			if cg != nil {
				if folSame(ps[i], countQ("some", "$count", fol("and", cg, cp))) && folSame(q, countDefinition("at most", 1, cg)) {
					return "introduction"
				}
				// Exactly one F plus a witnessed F and G establishes the description.
				if folSame(ps[i], countDefinition("exactly", 1, cg)) {
					for _, name := range folNames(q) {
						if countSameBody(q, folSub(fol("and", cg, cp), "$count", name)) {
							return "introduction"
						}
					}
				}
			}
		}
	}
	if len(ps) == 3 {
		g, p := descriptionInfo(c)
		if g != nil {
			for i, u := range ps {
				if !folSame(u, countDefinition("at most", 1, g)) && !folSame(u, countDefinition("exactly", 1, g)) {
					continue
				}
				rest := append([]*folFormula{}, ps[:i]...)
				rest = append(rest, ps[i+1:]...)
				for _, name := range folNames(countAnd(rest...)) {
					if countSameBody(countAnd(rest...), folSub(fol("and", g, p), "$count", name)) {
						return "introduction"
					}
				}
			}
		}
	}
	return ""
}
func descriptionSurfaceKey(text string) string {
	return normalizeClause(text) + "|name=" + descriptionName(text)
}
func descriptionSurfaceGuards(text string, spellings map[string]string) bool {
	s := folStrip(normalizeClause(text))
	if folExplicit(s) {
		return true
	}
	if strings.HasPrefix(s, "it is not the case that ") {
		return descriptionSurfaceGuards(s[len("it is not the case that "):], spellings)
	}
	if strings.HasPrefix(s, "if ") {
		if a, b, ok := folSplit(s[3:], " then "); ok {
			return descriptionSurfaceGuards(strings.TrimSuffix(a, ","), spellings) && descriptionSurfaceGuards(b, spellings)
		}
	}
	if cl, _, ok := descriptionParts(s); ok {
		g, e := descriptionGuard(cl)
		if e != nil {
			return false
		}
		key := folKey(descriptionRelations(g), nil, 0)
		if old, ok := spellings[key]; ok && old != cl {
			return false
		}
		spellings[key] = cl
		return true
	}
	for _, sep := range []string{" and ", " or "} {
		if a, b, ok := folSplit(s, sep); ok {
			return descriptionSurfaceGuards(a, spellings) && descriptionSurfaceGuards(b, spellings)
		}
	}
	return true
}
func descriptionPresent(f *folFormula) bool {
	if g, _ := descriptionInfo(f); g != nil {
		return true
	}
	for _, child := range f.children {
		if descriptionPresent(child) {
			return true
		}
	}
	return false
}
func gateDescriptions(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	r := EntailmentGateStepReport{Step: c.Number, Depth: c.Depth, Citations: c.Citations, Ranges: c.Ranges}
	ids := append([]int{}, c.Citations...)
	for _, span := range c.Ranges {
		ids = append(ids, span[0], span[1])
	}
	texts := []string{c.Statement}
	active := false
	for _, id := range ids {
		texts = append(texts, steps[id].Statement)
	}
	for _, s := range texts {
		active = active || strings.Contains(normalizeClause(s), "the ")
	}
	if !active {
		return r
	}
	// Assign readings once per verbatim clause, including repeated citations.
	spellings := map[string]string{}
	for _, s := range texts {
		if !descriptionSurfaceGuards(s, spellings) {
			return r
		}
	}
	hasDefinition := false
	keys := []string{}
	choices := map[string][]*folFormula{}
	for _, s := range texts {
		key := descriptionSurfaceKey(s)
		if _, ok := choices[key]; ok {
			continue
		}
		fs, e := descriptionParse(s)
		if e != nil {
			return r
		}
		for _, f := range fs {
			hasDefinition = hasDefinition || descriptionPresent(f)
		}
		choices[key] = fs
		keys = append(keys, key)
	}
	if !hasDefinition || len(keys) > 12 {
		return r
	}
	selected := map[string]*folFormula{}
	form := ""
	var check func(int) bool
	check = func(at int) bool {
		if at < len(keys) {
			for _, f := range choices[keys[at]] {
				selected[keys[at]] = f
				if !check(at + 1) {
					return false
				}
			}
			return true
		}
		target := selected[descriptionSurfaceKey(c.Statement)]
		if len(c.Ranges) == 0 {
			var ps []*folFormula
			for _, id := range c.Citations {
				ps = append(ps, selected[descriptionSurfaceKey(steps[id].Statement)])
			}
			form = descriptionRule(ps, target)
			return form != ""
		}
		translated := map[int]NumberedArgumentStep{}
		for id, s := range steps {
			if f := selected[descriptionSurfaceKey(s.Statement)]; f != nil {
				s.Statement = f.Render()
			}
			translated[id] = s
		}
		next := c
		next.Statement = target.Render()
		result := gateFOL(next, translated)
		form = "definition"
		return result.Passed
	}
	if check(0) {
		r.Passed = true
		r.Form = "description " + form
	}
	return r
}
