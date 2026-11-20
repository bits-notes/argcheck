package validation

import (
	"fmt"
	"strings"
)

// legacyFOLReadings is conservative: it never drops a candidate just because
// that candidate cannot be rendered. The old recognizers and their priority in
// the old gate are unchanged. Non-FOL operators are not treated as FOL atoms.
func legacyFOLReadings(text string) ([]*folFormula, string) {
	return allStrictRules().legacyFOLReadings(text)
}

func (policy strictRules) legacyFOLReadings(text string) ([]*folFormula, string) {
	s := folStrip(stripClauseWrappers(normalizeClause(text)))
	if folExplicit(s) {
		f, err := policy.parseFOL(s)
		if err != nil {
			return nil, "malformed explicit scope"
		}
		return []*folFormula{f}, ""
	}
	if len(necessityCandidates(s)) > 0 || len(possibilityCandidates(s)) > 0 || len(counterfactualCandidates(s)) > 0 || strings.Contains(s, "property") {
		return nil, "non-FOL modal, counterfactual, or property grammar"
	}
	if s == "contradiction" {
		return []*folFormula{fol("bottom")}, ""
	}
	// Parenthesized structured compounds already have a fixed boundary.
	if strings.ContainsAny(s, "()") {
		if f, err := policy.parseFOLScopeMode(s, map[string]bool{}, folExplicit(s), false); err == nil {
			return []*folFormula{f}, ""
		}
	}
	// A leading readable quantifier followed by a connective does not say
	// whether the connective is inside or outside its scope. Never select a
	// conjunction-first reading merely because that schema was tried first.
	for _, prefix := range []string{"every ", "all ", "some ", "no "} {
		if strings.HasPrefix(s, prefix) && (len(conjunctionCandidates(s)) > 0 || len(disjunctionCandidates(s)) > 0) {
			return nil, "ambiguous readable quantifier/connective scope"
		}
	}
	unique := func(fs []*folFormula) []*folFormula {
		out := []*folFormula{}
		seen := map[string]bool{}
		for _, f := range fs {
			k := folKey(f, nil, 0)
			if !seen[k] {
				seen[k] = true
				out = append(out, f)
				if len(out) == 2 {
					return out
				} // Two distinct readings suffice to refuse conversion.
			}
		}
		return out
	}
	pairs := func(op string, cs []clausePair) ([]*folFormula, string) {
		var out []*folFormula
		for _, c := range cs {
			as, ae := policy.legacyFOLReadings(c.left)
			bs, be := policy.legacyFOLReadings(c.right)
			if ae != "" {
				return nil, ae
			}
			if be != "" {
				return nil, be
			}
			for _, a := range as {
				for _, b := range bs {
					out = append(out, fol(op, a, b))
				}
			}
			out = unique(out)
			if len(out) > 1 {
				return out, ""
			}
		}
		return unique(out), ""
	}
	if ns := negationCandidates(s); len(ns) > 0 {
		fs, why := policy.legacyFOLReadings(ns[0])
		if why != "" {
			return nil, why
		}
		for i, f := range fs {
			fs[i] = fol("not", f)
		}
		return fs, ""
	}
	if cs := conditionalCandidates(s); len(cs) > 0 {
		return pairs("implies", cs)
	}
	if cs := biconditionalCandidates(s); len(cs) > 0 {
		return pairs("iff", cs)
	}
	if cs := conjunctionCandidates(s); len(cs) > 0 && !hasOuterGateOperator(s) {
		return pairs("and", cs)
	}
	if cs := disjunctionCandidates(s); len(cs) > 0 {
		return pairs("or", cs)
	}
	if relativeMarker(s) && !strings.Contains(s, " is such that ") {
		if q, ok := relativeQuant(s); ok {
			guards, why := policy.legacyFOLConjunction(q.guards)
			if why != "" {
				return nil, why
			}
			bodies, why := policy.legacyFOLReadings(q.body)
			if why != "" {
				return nil, why
			}
			var out []*folFormula
			for _, g := range guards {
				for _, b := range bodies {
					out = append(out, legacyFOLQuant(q.kind, "$relative", folSub(g, relativeHole, "$relative"), folSub(b, relativeHole, "$relative")))
				}
			}
			return unique(out), ""
		}
		if facts, ok := relativeGround(s); ok && !(len(facts) == 1 && facts[0] == s) {
			return policy.legacyFOLConjunction(facts)
		}
	}
	// Opaque predications stay zero-place atoms: exposing their subject as an
	// FOL term would permit identity substitution through an attitude or name.
	if identityOpaqueExtra.MatchString(s) {
		return []*folFormula{{op: "atom", symbol: s}}, ""
	}
	unary := func(pred, term string) *folFormula {
		return &folFormula{op: "atom", symbol: "is/" + policy.legacyFOLTerm(pred), terms: []string{term}}
	}
	// Preserve candidate enumeration, including embedded-copula alternatives.
	for _, family := range []struct {
		kind string
		cs   []clausePair
	}{
		{"no", policy.universalNegativeCandidates(s)}, {"every", universalCandidates(s)},
		{"some-not", particularNegativeCandidates(s)}, {"some", particularAffirmativeCandidates(s)},
	} {
		// Bare negative instances must not be promoted to universal negatives.
		if family.kind == "no" && !strings.HasPrefix(s, "no ") && !strings.HasPrefix(s, "every ") && !strings.HasPrefix(s, "all ") {
			continue
		}
		if strings.Contains(s, " is such that ") || len(family.cs) == 0 {
			continue
		}
		var out []*folFormula
		for _, c := range family.cs {
			b := unary(c.right, "$legacy")
			kind := family.kind
			if kind == "some-not" {
				kind = "some"
				b = fol("not", b)
			}
			out = append(out, legacyFOLQuant(kind, "$legacy", unary(c.left, "$legacy"), b))
		}
		return unique(out), ""
	}
	if cs := unrestrictedExistentialPredicateCandidates(s); len(cs) > 0 {
		var out []*folFormula
		for _, c := range cs {
			q := fol("some", unary(c, "$legacy"))
			q.symbol = "$legacy"
			out = append(out, q)
		}
		return unique(out), ""
	}
	if cs := identityCandidates(s); len(cs) > 0 {
		var out []*folFormula
		for _, c := range cs {
			// v28.2: an indefinite or quantified side is not a name. termCandidates
			// strips "a "/"an " for predicate nouns, which read "Tom is identical to
			// a cat" as tom = cat.
			if quantifiedIdentitySide(c.left) || quantifiedIdentitySide(c.right) {
				continue
			}
			out = append(out, &folFormula{op: "equal", terms: []string{policy.legacyFOLTerm(c.left), policy.legacyFOLTerm(c.right)}})
		}
		if len(out) == 0 {
			return nil, "identity needs names on both sides, not quantifier phrases"
		}
		return unique(out), ""
	}
	if cs := negativeInstanceCandidates(s); len(cs) > 0 {
		var out []*folFormula
		for _, c := range cs {
			out = append(out, fol("not", unary(c.right, policy.legacyFOLTerm(c.left))))
		}
		return unique(out), ""
	}
	if cs := policy.groundInstanceCandidates(s); len(cs) > 0 {
		var out []*folFormula
		for _, c := range cs {
			out = append(out, unary(c.right, policy.legacyFOLTerm(c.left)))
		}
		return unique(out), ""
	}
	if cs := relationCandidates(s); len(cs) > 0 {
		var out []*folFormula
		for _, c := range cs {
			orders := relationOrders(c)
			if len(orders) == 0 {
				return nil, "readable quantifier limit"
			}
			for _, order := range orders {
				terms := make([]string, len(c.terms))
				for i, t := range c.terms {
					terms[i] = policy.legacyFOLTerm(t.text)
					if t.quant != "" {
						terms[i] = fmt.Sprintf("$slot%d", i)
					}
				}
				symbol := strings.ReplaceAll(c.verb, "\x00", "/")
				f := &folFormula{op: "atom", symbol: symbol, terms: terms}
				for i := len(order) - 1; i >= 0; i-- {
					if i == 0 && c.innerNot {
						f = fol("not", f)
					}
					slot := order[i]
					term := c.terms[slot]
					f = legacyFOLQuant(term.quant, terms[slot], unary(term.text, terms[slot]), f)
				}
				if c.outerNot {
					f = fol("not", f)
				}
				out = append(out, f)
			}
		}
		return unique(out), ""
	}
	// Historical propositional schemas use verbatim opaque clauses. Keep them
	// zero-ary; never invent an individual or silently change predicate arity.
	return []*folFormula{{op: "atom", symbol: s}}, ""
}

func legacyFOLTerm(s string) string { return allStrictRules().legacyFOLTerm(s) }

func (policy strictRules) legacyFOLTerm(s string) string {
	t := policy.termCandidates(s)[0]
	// Surface names must never collide with the AST's internal variable namespace.
	if strings.HasPrefix(t, "$") || strings.HasPrefix(t, "\x00") {
		return "\x00" + t
	}
	return t
}
func legacyFOLQuant(kind, v string, guard, body *folFormula) *folFormula {
	if kind == "no" {
		kind = "every"
		body = fol("not", body)
	}
	if kind == "every" {
		body = fol("implies", guard, body)
	} else {
		body = fol("and", guard, body)
	}
	q := fol(kind, body)
	q.symbol = v
	return q
}

func (policy strictRules) legacyFOLConjunction(texts []string) ([]*folFormula, string) {
	var out []*folFormula
	for _, s := range texts {
		fs, why := policy.legacyFOLReadings(s)
		if why != "" {
			return nil, why
		}
		if len(out) == 0 {
			out = fs
			continue
		}
		var next []*folFormula
		for _, a := range out {
			for _, b := range fs {
				next = append(next, fol("and", a, b))
			}
		}
		out = next
	}
	return out, ""
}

// Rendering failures remain named coverage exceptions, not alternative readings
// to discard. In particular a one-letter legacy name is not a bound variable.
func legacyFOLRender(f *folFormula) (string, string) { return allStrictRules().legacyFOLRender(f) }

func (policy strictRules) legacyFOLRender(f *folFormula) (string, string) {
	s := policy.renderFOL(f)
	got, err := policy.parseFOLScope(s, map[string]bool{}, true)
	if err != nil || !folSame(f, got) {
		why := policy.legacyFOLRenderReason(folRenderNames(f, nil, 0), map[string]bool{})
		if why == "" {
			why = "rendered compound changes connective scope"
		}
		return "", why
	}
	readings, why := policy.legacyFOLReadings(s)
	if why != "" || len(readings) != 1 || !folSame(f, readings[0]) {
		return "", "rendered surface conflicts with historical parenthesized parsing"
	}
	return s, ""
}

// Diagnose failures at their lexical leaf, retaining a reason on every ledger
// row. A variable-shaped legacy constant cannot silently become a variable.
func (policy strictRules) legacyFOLRenderReason(f *folFormula, bound map[string]bool) string {
	for _, term := range f.terms {
		if folVariable.MatchString(term) {
			return "legacy constant looks like an explicit variable (requires a name escape)"
		}
		if strings.HasPrefix(term, "\x00") {
			return "legacy constant uses a reserved internal marker"
		}
	}
	if f.op == "every" || f.op == "some" {
		next := map[string]bool{}
		for k, v := range bound {
			next[k] = v
		}
		next[strings.TrimPrefix(f.symbol, "$")] = true
		return policy.legacyFOLRenderReason(f.children[0], next)
	}
	for _, child := range f.children {
		if why := policy.legacyFOLRenderReason(child, bound); why != "" {
			return why
		}
	}
	if f.op == "atom" || f.op == "equal" {
		got, err := policy.parseFOLScope(policy.renderFOLSurface(f), bound, true)
		if err != nil {
			if len(f.terms) == 0 {
				return "opaque clause conflicts with explicit scope or connective syntax"
			}
			return "legacy lexical atom has no unique explicit parse"
		}
		if !folSame(f, got) {
			return "rendered atom changes lexical segmentation or operator structure"
		}
	}
	return ""
}

func notationEquivalence(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	return allStrictRules().notationEquivalence(c, steps)
}

func (policy strictRules) notationEquivalence(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	r := EntailmentGateStepReport{Step: c.Number, Depth: c.Depth, Citations: c.Citations, Ranges: c.Ranges, NearestForm: "notation equivalence", Mismatch: "requires one premise and identical single readings"}
	if len(c.Citations) != 1 || len(c.Ranges) != 0 {
		return r
	}
	a, ae := policy.legacyFOLReadings(steps[c.Citations[0]].Statement)
	b, be := policy.legacyFOLReadings(c.Statement)
	if len(a) > 1 || len(b) > 1 || strings.Contains(ae, "ambiguous") || strings.Contains(be, "ambiguous") {
		r.Mismatch = "ambiguous sentence: use is such that to fix the reading"
		return r
	}
	if ae != "" || be != "" || len(a) != 1 || len(b) != 1 {
		return r
	}
	if folSame(a[0], b[0]) {
		r.Passed = true
		r.Form = "notation equivalence"
		r.NearestForm = ""
		r.Mismatch = ""
	}
	return r
}

// identityQuantifierWords are determiners and quantifier pronouns that cannot
// open a name. Personal and reflexive pronouns are left to the existing readings.
var identityQuantifierWords = map[string]bool{
	"every": true, "all": true, "each": true, "any": true, "some": true, "no": true, "none": true,
	"nothing": true, "nobody": true, "something": true, "someone": true, "somebody": true,
	"everything": true, "everyone": true, "everybody": true, "anything": true, "anyone": true,
	"anybody": true, "a": true, "an": true, "few": true, "many": true, "most": true, "several": true,
	"both": true, "either": true, "neither": true, "only": true, "not": true,
}

func quantifiedIdentitySide(term string) bool {
	words := strings.Fields(normalizeClause(term))
	return len(words) > 0 && identityQuantifierWords[strings.Trim(words[0], "'’")]
}
