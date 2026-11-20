package validation

// This file only renders evidence returned by a successful matcher. It must
// never parse a statement or participate in selecting or rejecting a match.
import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// readingLiteral quotes an opaque matched slot without hiding its spelling.
func readingLiteral(s string) string {
	return strconv.Quote(strings.ReplaceAll(s, string(nameLetterJoiner), ""))
}

func readingMP(a, b string) string {
	return "Premise: if " + readingLiteral(a) + ", then " + readingLiteral(b) + ". Premise: " + readingLiteral(a) + ". So: " + readingLiteral(b) + "."
}

// A certificate application preserves opaque lexical inputs as literals. It
// does not invent a parse for forms whose matchers only expose a boolean.
// Equality witnesses, when available, expose the actual selected clause slots.
func readingApplication(form string, ps []NumberedArgumentStep, c NumberedArgumentStep, ids []clauseIdentity) string {
	if english := readingCategorical(form, ps, c, ids); english != "" {
		return english
	}
	parts := make([]string, 0, len(ps))
	for _, p := range ps {
		parts = append(parts, fmt.Sprintf("%d: %s", p.Number, readingLiteral(p.Statement)))
	}
	for _, span := range c.Ranges {
		parts = append(parts, fmt.Sprintf("range(%d..%d)", span[0], span[1]))
	}
	prefix := strings.Join(parts, "; ") + " ⊢ "
	if len(parts) == 0 {
		prefix = "therefore "
	}
	s := "SEMI-FORMAL: " + form + " {" + prefix + readingLiteral(c.Statement) + "}"
	for _, id := range ids {
		s += " {" + id.leftLabel + " = " + readingLiteral(id.left) + "; " + id.rightLabel + " = " + readingLiteral(id.right) + "}"
	}
	return s
}

func readingScoped(form string, c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	var ps []NumberedArgumentStep
	for _, id := range c.Citations {
		ps = append(ps, steps[id])
	}
	var parts []string
	for _, p := range ps {
		parts = append(parts, readingLiteral(p.Statement))
	}
	for _, span := range c.Ranges {
		parts = append(parts, fmt.Sprintf("assume %s (steps %d-%d); so %s", readingLiteral(steps[span[0]].Statement), span[0], span[1], readingLiteral(steps[span[1]].Statement)))
	}
	return "SEMI-FORMAL: " + form + ": " + strings.Join(parts, "; ") + "; therefore " + readingLiteral(c.Statement)
}

func readingPattern(p v13Pattern, bindings map[string]string) string {
	if len(p.children) == 0 {
		return readingLiteral(bindings[p.op])
	}
	cs := make([]string, len(p.children))
	for i, c := range p.children {
		cs[i] = readingPattern(c, bindings)
	}
	return readingConnective(p.op, cs...)
}
func readingPatternRule(rule v13Rule, bindings map[string]string) string {
	parts := make([]string, len(rule))
	for i, p := range rule {
		parts[i] = readingPattern(p, bindings)
	}
	if readingPatternEnglish(rule) {
		return readingEnglishRule(parts[len(parts)-1], parts[:len(parts)-1]...)
	}
	return "SEMI-FORMAL: " + strings.Join(parts[:len(parts)-1], "; ") + " ⊢ " + parts[len(parts)-1]
}
func readingSubstitution(sub schematicSubstitution) string {
	keys := make([]rune, 0, len(sub))
	for k := range sub {
		if _, variable := schematicVariableMarkers[k]; variable {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	var parts []string
	for _, k := range keys {
		parts = append(parts, schematicDisplay(string(k))+" = "+readingLiteral(sub[k]))
	}
	return "Bindings: {" + strings.Join(parts, "; ") + "}. "
}

// Render the already checked FOL AST. Bound variables and quoted constants
// occupy separate namespaces. Each binder gets a fresh deterministic Xn.
func readingFOL(f *folFormula) string {
	next := 0
	var render func(*folFormula, map[string]string) string
	render = func(f *folFormula, env map[string]string) string {
		if f == nil {
			return "⊥"
		}
		term := func(s string) string {
			if v, ok := env[s]; ok {
				return v
			}
			return readingLiteral(s)
		}
		if f.op == "every" || f.op == "some" {
			next++
			v := fmt.Sprintf("X%d", next)
			nested := make(map[string]string, len(env)+1)
			for k, s := range env {
				nested[k] = s
			}
			nested[f.symbol] = v
			q := "∀"
			if f.op == "some" {
				q = "∃"
			}
			return q + v + " (" + render(f.children[0], nested) + ")"
		}
		switch f.op {
		case "atom":
			ts := make([]string, len(f.terms))
			for i, t := range f.terms {
				ts[i] = term(t)
			}
			return readingLiteral(f.symbol) + "(" + strings.Join(ts, ", ") + ")"
		case "equal":
			return "(" + term(f.terms[0]) + " = " + term(f.terms[1]) + ")"
		case "bottom":
			return "⊥"
		case "not":
			return "¬(" + render(f.children[0], env) + ")"
		}
		ops := map[string]string{"and": "∧", "or": "∨", "implies": "→", "iff": "↔"}
		if op, ok := ops[f.op]; ok {
			return "(" + render(f.children[0], env) + " " + op + " " + render(f.children[1], env) + ")"
		}
		cs := make([]string, len(f.children))
		for i, c := range f.children {
			cs[i] = render(c, env)
		}
		return f.op + "(" + strings.Join(cs, "; ") + ")"
	}
	return render(f, nil)
}
func readingFOLRule(ps []*folFormula, c *folFormula) string {
	// Only readingUniversal creates this private binder; actual FOL stays semi-formal.
	if c != nil && c.symbol == "$reading" {
		var parts []string
		for _, p := range ps {
			parts = append(parts, readingUniversalEnglish(p))
		}
		return readingEnglishRule(readingUniversalEnglish(c), parts...)
	}
	var parts []string
	for _, p := range ps {
		parts = append(parts, readingFOL(p))
	}
	if len(parts) == 0 {
		return "SEMI-FORMAL: therefore " + readingFOL(c)
	}
	return "SEMI-FORMAL: " + strings.Join(parts, "; ") + " ⊢ " + readingFOL(c)
}

func readingConjunctiveMP(conjuncts []string, consequent string) string {
	parts := make([]string, len(conjuncts))
	for i, c := range conjuncts {
		parts[i] = readingLiteral(c)
	}
	conjunction := parts[0]
	for _, part := range parts[1:] {
		conjunction = readingConnective("conjunction", conjunction, part)
	}
	return readingEnglishRule(readingLiteral(consequent), append([]string{readingConnective("conditional", conjunction, readingLiteral(consequent))}, parts...)...)
}
func readingUniversal(pair clausePair) *folFormula {
	a := &folFormula{op: "atom", symbol: "is/" + pair.left, terms: []string{"$reading"}}
	b := &folFormula{op: "atom", symbol: "is/" + pair.right, terms: []string{"$reading"}}
	f := fol("every", fol("implies", a, b))
	f.symbol = "$reading"
	return f
}

// Substitution for display only: never call a policy guard while rendering.
func readingBoundTemplate(template string, bindings schematicSubstitution) string {
	var b strings.Builder
	for _, r := range template {
		if variable, ok := schematicVariableForMarker(r); ok {
			b.WriteString(bindings[variable])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func readingAtom(s string) string { return readingLiteral(s) }
func readingBinary(op string, pair clausePair) string {
	return readingConnective(map[string]string{"→": "conditional", "↔": "biconditional", "∧": "conjunction", "∨": "disjunction"}[op], readingAtom(pair.left), readingAtom(pair.right))
}
func readingSequent(conclusion string, premises ...string) string {
	return readingEnglishRule(conclusion, premises...)
}

// Negation markers here were emitted by the rendering call sites, never parsed
// from a source sentence. Stop at quoted lexical slots.
func readingNegations(s string) string {
	if strings.HasPrefix(s, "¬") {
		return "it is not the case that (" + readingNegations(s[len("¬"):]) + ")"
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return "(" + readingNegations(s[1:len(s)-1]) + ")"
	}
	return s
}
func readingEnglishRule(conclusion string, premises ...string) string {
	var parts []string
	for _, p := range premises {
		parts = append(parts, readingNegations(p)+".")
	}
	return strings.Join(parts, " ") + " So: " + readingNegations(conclusion) + "."
}
func readingConnective(op string, cs ...string) string {
	switch op {
	case "negation":
		return "it is not the case that (" + cs[0] + ")"
	case "conditional":
		return "if (" + cs[0] + "), then (" + cs[1] + ")"
	case "biconditional":
		return "(" + cs[0] + ") if and only if (" + cs[1] + ")"
	case "conjunction":
		return "both (" + cs[0] + ") and (" + cs[1] + ")"
	case "disjunction":
		return "either (" + cs[0] + ") or (" + cs[1] + ")"
	default:
		return op + "(" + strings.Join(cs, "; ") + ")"
	}
}
func readingPatternEnglish(rule v13Rule) bool {
	var clean func(v13Pattern) bool
	clean = func(p v13Pattern) bool {
		if len(p.children) == 0 {
			return true
		}
		switch p.op {
		case "negation", "conditional", "biconditional", "conjunction", "disjunction":
		default:
			return false
		}
		for _, c := range p.children {
			if !clean(c) {
				return false
			}
		}
		return true
	}
	for _, p := range rule {
		if !clean(p) {
			return false
		}
	}
	return true
}

// Class positions are known from the selected schema. Predicates retain their
// recorded spelling; class nouns receive an article only in subject-class slots.
func readingClass(s string) string {
	if strings.HasPrefix(s, "a ") || strings.HasPrefix(s, "an ") || strings.HasPrefix(s, "the ") {
		return s
	}
	// Quoting a class label also handles plural and modified class names without
	// guessing morphology or treating the class as a named individual.
	return "in the class " + readingLiteral(s)
}
func readingMember(subject, predicate string) string { return subject + " is " + predicate }
func readingAll(class, predicate string, negative bool) string {
	consequent := readingMember("X1", predicate)
	if negative {
		consequent = "it is not the case that " + consequent
	}
	return "For every X1: if " + readingMember("X1", readingClass(class)) + ", then " + consequent
}
func readingSome(class, predicate string, negative bool) string {
	consequent := readingMember("X1", predicate)
	if negative {
		consequent = "it is not the case that " + consequent
	}
	return "there is an X1 such that both (" + readingMember("X1", readingClass(class)) + ") and (" + consequent + ")"
}
func readingUniversalEnglish(f *folFormula) string {
	pair := f.children[0]
	return readingAll(strings.TrimPrefix(pair.children[0].symbol, "is/"), strings.TrimPrefix(pair.children[1].symbol, "is/"), false)
}

// Only consume named slots from the successful equality witnesses. If a legacy
// path returned only a boolean, retain its honest semi-formal application.
func readingCategorical(form string, ps []NumberedArgumentStep, c NumberedArgumentStep, ids []clauseIdentity) string {
	slots := map[string]string{}
	for _, id := range ids {
		slots[id.leftLabel] = id.left
		slots[id.rightLabel] = id.right
	}
	p := func(i int, role string) string {
		if i >= len(ps) {
			return ""
		}
		return slots[fmt.Sprintf("premise %d %s", ps[i].Number, role)]
	}
	result := func(role string) string { return slots[fmt.Sprintf("conclusion %d %s", c.Number, role)] }
	switch form {
	case "universal instantiation":
		class, predicate, instance := p(1, "predicate"), p(0, "predicate"), p(1, "instance")
		if class == "" || predicate == "" || instance == "" {
			return ""
		}
		return readingEnglishRule(readingMember(readingLiteral(result("instance")), result("predicate")), readingAll(class, predicate, false), readingMember(readingLiteral(instance), class))
	case "universal modus tollens":
		class, predicate, instance := slots["denied conclusion class"], slots["universal predicate"], slots["premise instance"]
		if class == "" || predicate == "" || instance == "" {
			return ""
		}
		return readingEnglishRule("it is not the case that "+readingMember(readingLiteral(instance), class), readingAll(class, predicate, false), "it is not the case that "+readingMember(readingLiteral(instance), predicate))
	case "universal negative instantiation":
		class, predicate, instance := result("excluded class"), p(0, "predicate"), p(0, "instance")
		if class == "" || predicate == "" || instance == "" {
			return ""
		}
		return readingEnglishRule("it is not the case that "+readingMember(readingLiteral(instance), class), readingMember(readingLiteral(instance), predicate), readingAll(class, predicate, true))
	case "existential syllogism", "existential negative syllogism":
		negative := form == "existential negative syllogism"
		class, middle, predicate := p(0, "subject"), p(0, "predicate"), p(1, "predicate")
		if negative {
			predicate = p(1, "excluded term")
		}
		if class == "" || middle == "" || predicate == "" {
			return ""
		}
		return readingEnglishRule(readingSome(class, predicate, negative), readingSome(class, middle, false), readingAll(middle, predicate, negative))
	case "universal negative syllogism":
		class, middle, predicate := p(0, "class"), p(0, "predicate"), p(1, "excluded term")
		if class == "" || middle == "" || predicate == "" {
			return ""
		}
		return readingEnglishRule(readingAll(class, predicate, true), readingAll(class, middle, false), readingAll(middle, predicate, true))
	}
	return ""
}

func readingFOLScoped(ps []*folFormula, c *folFormula, discharges []string) string {
	var parts []string
	for _, p := range ps {
		parts = append(parts, readingFOL(p))
	}
	parts = append(parts, discharges...)
	return "SEMI-FORMAL: " + strings.Join(parts, "; ") + "; therefore " + readingFOL(c)
}
