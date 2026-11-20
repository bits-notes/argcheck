package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Legacy schemas may weaken positive counterfactual premises to material
// conditionals verbatim. Polarity guards below apply at the schema boundary,
// not at the splitter: refusing all counterfactual inputs loses valid proofs.
type gateConditionalKind int

const (
	materialConditional gateConditionalKind = iota
	wouldConditional
	mightConditional
)

var nonMaterialAuxiliary = regexp.MustCompile(`(?i)\b(would|might|could have)\b`)

var mightPerfectAuxiliary = regexp.MustCompile(`\bmight (not )?have\b`)

func conditionalKind(consequent, antecedent string) gateConditionalKind {
	s := v19Unparenthesize(stripClauseWrappers(consequent))
	antecedent = normalizeClause(antecedent)
	// A conditional inside the consequent is a separate connective. Its would
	// does not classify the enclosing material conditional as counterfactual.
	if strings.HasPrefix(s, "if ") || len(biconditionalCandidates(s)) > 0 {
		return materialConditional
	}
	if neg := negationCandidates(s); len(neg) > 0 {
		return conditionalKind(neg[0], antecedent)
	}
	for _, op := range []string{"conjunction", "disjunction"} {
		if parts := v19PolarityPairs(op, s); len(parts) > 0 {
			kind := materialConditional
			for _, part := range []string{parts[0].left, parts[0].right} {
				k := conditionalKind(part, antecedent)
				if k == mightConditional {
					return k
				}
				if k == wouldConditional {
					kind = k
				}
			}
			return kind
		}
	}
	kind := materialConditional
	antecedentMood := false
	for _, mood := range counterfactualMood.FindAllString(antecedent, -1) {
		// A bare would-clause in an antecedent is itself an atom; only
		// the existing had/were antecedent markers preserve this refusal.
		if strings.HasPrefix(mood, "had ") || strings.HasPrefix(mood, "were ") {
			antecedentMood = true
		}
	}
	for _, m := range nonMaterialAuxiliary.FindAllString(s, -1) {
		// v31: indicative might is part of the atom. Keep the existing
		// perfect-modal boundary and conservatively retain might with the
		// existing had/were mood markers in its antecedent. Neither this
		// classifier nor the material schemas remove modal words.
		if m == "might" && !mightPerfectAuxiliary.MatchString(s) && !antecedentMood {
			continue
		}
		if m != "would" {
			return mightConditional
		}
		kind = wouldConditional
	}
	return kind
}
func nonMaterialConditional(s string) bool {
	pairs := v19ConditionalPair(s)
	return len(pairs) > 0 && conditionalKind(pairs[0].right, pairs[0].left) != materialConditional
}
func counterfactualCandidates(s string) []clausePair {
	pairs := v19ConditionalPair(s)
	if len(pairs) == 0 || conditionalKind(pairs[0].right, pairs[0].left) != wouldConditional {
		return nil
	}
	// Only the first then closes the outer antecedent; nested consequents never
	// become alternative antecedents. Nested counterfactual rules are deferred.
	return pairs[:1]
}

// Plain content is local to a counterfactual operand, never a normalization
// of ordinary cited clauses. These are auxiliary operations only. In particular
// gone stays gone, and a bare would light has no plain candidate.
var counterfactualMood = regexp.MustCompile(`\b(would not have been|would have been|had not been|had been|would not have|would have|had not|had|would be|were) `)

var counterfactualUnsafe = regexp.MustCompile(`\b(if|then|and|or|either|that|who|which|whether|necessarily|possibly)\b|[,;:"` + "`" + `()?!]`)

func v19PlainClause(s string) bool { return s != "" && !counterfactualUnsafe.MatchString(s) }

// A mood auxiliary is not a report context, but all the older scope guards
// still apply to an unscaffolded Boolean operand ("thought P and Q", etc.).
func v19BooleanLeaf(s string) bool {
	return v19PlainClause(s) && !opaqueContextWords.MatchString(nonMaterialAuxiliary.ReplaceAllString(s, "auxiliary"))
}
func v19BooleanPairs(op, s string) []clausePair {
	pairs := v13Pairs(op, s)
	if len(pairs) > 0 {
		return pairs
	}
	if op == "conjunction" {
		pairs = conjunctionCandidates(s)
		if len(pairs) == 1 && v19BooleanLeaf(pairs[0].left) && v19BooleanLeaf(pairs[0].right) {
			return pairs
		}
	}
	return nil
}
func counterfactualPlain(s string) []string {
	s = stripClauseWrappers(s)
	if neg := negationCandidates(s); len(neg) > 0 {
		var out []string
		for _, p := range counterfactualPlain(neg[0]) {
			out = append(out, "it is not the case that "+p)
		}
		return out
	}
	// Preserve explicit Boolean scope and transform each complete operand.
	for _, op := range []string{"conjunction", "disjunction"} {
		pairs := v19BooleanPairs(op, s)
		if len(pairs) == 0 {
			continue
		}
		var out []string
		for _, l := range counterfactualPlain(pairs[0].left) {
			for _, r := range counterfactualPlain(pairs[0].right) {
				if op == "conjunction" {
					if strings.HasPrefix(s, "it is the case both that ") {
						out = append(out, "it is the case both that "+l+", and that "+r)
					} else {
						out = append(out, l+" and "+r)
					}
				} else {
					out = append(out, "either "+l+" or "+r)
				}
			}
		}
		return out
	}
	// Never rewrite an auxiliary buried inside another operator, relative
	// clause, quotation, or report. Recognition is conservative; content is not.
	if !v19PlainClause(s) {
		return nil
	}
	loc := counterfactualMood.FindStringSubmatchIndex(s)
	if loc == nil {
		if nonMaterialAuxiliary.MatchString(s) {
			return nil
		}
		return []string{s}
	}
	subject := strings.TrimSpace(s[:loc[0]])
	tail := strings.TrimSpace(s[loc[1]:])
	if subject == "" || tail == "" || opaqueContextWords.MatchString(subject) || nonMaterialAuxiliary.MatchString(subject+" "+tail) {
		return nil
	}
	auxiliary := s[loc[2]:loc[3]]
	var replacements []string
	switch auxiliary {
	case "had not been":
		replacements = []string{"was "}
	case "had been", "would have been", "would not have been":
		replacements = []string{"was ", "were "}
	case "were", "would be":
		replacements = []string{"is ", "are ", "am "}
	default:
		replacements = []string{""}
	}
	var out []string
	for _, replacement := range replacements {
		p := subject + " " + replacement + tail
		if strings.Contains(auxiliary, " not ") || strings.HasSuffix(auxiliary, " not") {
			p = "it is not the case that " + p
		}
		out = append(out, p)
	}
	return out
}
func counterfactualPlainPairs(s string) []clausePair {
	var out []clausePair
	for _, pair := range counterfactualCandidates(s) {
		for _, l := range counterfactualPlain(pair.left) {
			for _, r := range counterfactualPlain(pair.right) {
				out = append(out, clausePair{l, r})
			}
		}
	}
	return out
}
func v19CF(a, b v13Pattern) v13Pattern { return v13Pattern{"counterfactual", []v13Pattern{a, b}} }

// Enumerate mappings against the cited clauses. Each alternative gets its own
// bindings: a failed candidate must never bind a variable for the next one.
func (policy strictRules) v19Match(pattern v13Pattern, s string, bindings map[string]string, raw bool) []map[string]string {
	if len(pattern.children) == 0 {
		if bound, ok := bindings[pattern.op]; ok && !policy.equalClauses(bound, s) {
			return nil
		}
		next := make(map[string]string, len(bindings)+1)
		for k, v := range bindings {
			next[k] = v
		}
		next[pattern.op] = s
		return []map[string]string{next}
	}
	if pattern.op == "negation" {
		neg := negationCandidates(s)
		if len(neg) == 0 {
			return nil
		}
		return policy.v19Match(pattern.children[0], neg[0], bindings, raw)
	}
	var pairs []clausePair
	if pattern.op == "counterfactual" {
		if raw {
			pairs = counterfactualCandidates(s)
		} else {
			pairs = counterfactualPlainPairs(s)
		}
	} else {
		if pattern.op == "conditional" && nonMaterialConditional(s) {
			return nil
		}
		pairs = v19BooleanPairs(pattern.op, s)
	}
	var out []map[string]string
	for _, pair := range pairs {
		for _, b := range policy.v19Match(pattern.children[0], pair.left, bindings, raw) {
			out = append(out, policy.v19Match(pattern.children[1], pair.right, b, raw)...)
		}
	}
	return out
}

func (policy strictRules) v19Rules(p []NumberedArgumentStep, c NumberedArgumentStep, raw bool, rules ...v13Rule) schemaMatchOutcome {
	for _, rule := range rules {
		if len(rule) != len(p)+1 {
			continue
		}
		states := []map[string]string{{}}
		for i, pattern := range rule {
			s := c.Statement
			if i < len(p) {
				s = p[i].Statement
			}
			var next []map[string]string
			for _, bindings := range states {
				next = append(next, policy.v19Match(pattern, s, bindings, raw)...)
			}
			states = next
			if len(states) == 0 {
				break
			}
		}
		if len(states) > 0 {
			return schemaMatchOutcome{matched: true, traction: true, score: len(rule), reading: readingPatternRule(rule, states[0])}
		}
	}
	return schemaMatchOutcome{}
}

func (policy strictRules) matchCounterfactualMP(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	result := policy.v19Rules(p, c, false, v13Rule{v19CF(v13P, v13Q), v13P, v13Q})
	if result.matched {
		return result
	}
	// Protect uppercase schematic slots before normalization, just as the
	// older schematic rules do. Restoring every lowercase x/y/z afterwards
	// would incorrectly turn ordinary constant names into variables.
	marked, _ := replaceSchematicVariables(p[0].Statement)
	for _, pair := range counterfactualPlainPairs(marked) {
		for variable, marker := range schematicVariableMarkers {
			pair.left = strings.ReplaceAll(pair.left, string(marker), string(variable))
			pair.right = strings.ReplaceAll(pair.right, string(marker), string(variable))
		}
		virtual := append([]NumberedArgumentStep(nil), p...)
		virtual[0].Statement = "If " + pair.left + ", then " + pair.right
		if policy.matchSchematicInstantiation(virtual, c).matched {
			return schemaMatchOutcome{matched: true, traction: true, score: 3}
		}
	}
	return result
}

func (policy strictRules) matchCounterfactualMT(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v19Rules(p, c, false, v13Rule{v19CF(v13P, v13Q), v13Not(v13Q), v13Not(v13P)})
}

func (policy strictRules) matchCounterfactualMaterial(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v19Rules(p, c, false, v13Rule{v19CF(v13P, v13Q), v13If(v13P, v13Q)}, v13Rule{v19CF(v13P, v13Q), v13Or(v13Not(v13P), v13Q)})
}

func (policy strictRules) v19ConsequentRules(p []NumberedArgumentStep, c NumberedArgumentStep, rules ...v13Rule) schemaMatchOutcome {
	out := policy.v19Rules(p, c, false, rules...)
	if out.matched {
		return out
	}
	// Closure of consequents can compare the would clauses verbatim even when
	// bare-verb content is unavailable. This mode never detaches plain content.
	return policy.v19Rules(p, c, true, rules...)
}

func (policy strictRules) matchCounterfactualConjunction(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v19ConsequentRules(p, c, v13Rule{v19CF(v13P, v13Q), v19CF(v13P, v13R), v19CF(v13P, v13And(v13Q, v13R))})
}

func (policy strictRules) matchCounterfactualConjunctionElimination(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v19ConsequentRules(p, c, v13Rule{v19CF(v13P, v13And(v13Q, v13R)), v19CF(v13P, v13Q)}, v13Rule{v19CF(v13P, v13And(v13Q, v13R)), v19CF(v13P, v13R)})
}

func (policy strictRules) matchCounterfactualDisjunction(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v19ConsequentRules(p, c, v13Rule{v19CF(v13P, v13Q), v19CF(v13P, v13Or(v13Q, v13R))}, v13Rule{v19CF(v13P, v13Q), v19CF(v13P, v13Or(v13R, v13Q))})
}

func (policy strictRules) matchCounterfactualDilemma(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	// Reuse only the bijection/Boolean shape of all three existing dilemmas.
	// Every conditional must be would, and every side must have plain content.
	choices := make([][]clausePair, len(p))
	count := 0
	for i, step := range p {
		choices[i] = counterfactualPlainPairs(step.Statement)
		if len(choices[i]) > 0 {
			count++
		}
	}
	if count != len(p)-1 {
		return schemaMatchOutcome{}
	}
	virtual := append([]NumberedArgumentStep(nil), p...)
	var visit func(int) bool
	visit = func(i int) bool {
		if i == len(p) {
			for _, kind := range []dilemmaKind{simpleDilemma, constructiveDilemma, destructiveDilemma} {
				if policy.matchDilemma(virtual, c, kind).matched {
					return true
				}
			}
			return false
		}
		if len(choices[i]) == 0 {
			return visit(i + 1)
		}
		for _, pair := range choices[i] {
			virtual[i].Statement = "If " + pair.left + ", then " + pair.right
			if visit(i + 1) {
				return true
			}
		}
		return false
	}
	return schemaMatchOutcome{matched: visit(0)}
}

func (policy strictRules) matchCounterfactualReflexivity(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	for _, pair := range counterfactualPlainPairs(c.Statement) {
		if !policy.equalClauses(pair.left, pair.right) {
			continue
		}
		re := regexp.MustCompile(`(^|[^\pL\pN_])` + regexp.QuoteMeta(normalizeClause(pair.left)) + `($|[^\pL\pN_])`)
		if re.MatchString(normalizeClause(p[0].Statement)) {
			return schemaMatchOutcome{matched: true, traction: true, score: 2}
		}
		for _, source := range counterfactualPlainPairs(p[0].Statement) {
			if policy.equalClauses(pair.left, source.left) || policy.equalClauses(pair.left, source.right) {
				return schemaMatchOutcome{matched: true, traction: true, score: 2}
			}
		}
	}
	return schemaMatchOutcome{}
}

const counterfactualAvailable = "Positive counterfactual premises may be weakened verbatim to material conditionals. Counterfactual steps the gate accepts: modus ponens, modus tollens, counterfactual to material, consequent conjunction, consequent weakening, dilemma, and reflexivity."

func counterfactualRuleNote(rule, statement string) string {
	reason := "material rules do not describe closest-world counterfactuals"
	switch rule {
	case "contraposition":
		reason = "the contrapositive concerns different closest worlds"
	case "hypothetical syllogism":
		reason = "the closest worlds can change at each link"
	case "conditional proof":
		reason = "ordinary subproofs allow antecedent strengthening and false-antecedent/true-consequent proofs; counterfactual subproofs are deferred to v2"
	case "exportation", "importation":
		reason = "nested counterfactuals shift the closest worlds"
	case "negated conditional":
		reason = "denying a counterfactual entails neither its antecedent nor its negated consequent"
	case "strong centering":
		reason = "weak centering permits other equally close worlds where the consequent is false"
	case "antecedent strengthening", "simplification of disjunctive antecedents":
		reason = "restricting the antecedent changes which worlds are closest"
	case "conditional excluded middle":
		reason = "equally close antecedent-worlds may disagree on the consequent"
	case "might to would":
		reason = "some closest antecedent-worlds satisfying the consequent does not mean all do"
	case "modus ponens", "modus tollens", "schematic instantiation", "conjunctive modus ponens", "schematic conjunctive modus ponens":
		reason = "counterfactual detachment requires the permitted plain-content mapping, not verbatim mood clauses"
	case "material implication", "material to counterfactual":
		reason = "a false antecedent or true consequent does not establish a counterfactual"
	}
	return fmt.Sprintf("%q is a counterfactual or might/could-have conditional; %s cannot justify this counterfactual inference: %s. %s Might/could-have conditionals have no counterfactual rules in v19.", statement, rule, reason, counterfactualAvailable)
}

// Diagnose only after every genuine schema has failed. Masking the mood words
// here is a shape probe for the old material rule, never a proof or a rewrite
// passed back into the gate. It names every formerly certifiable refusal.
func (policy strictRules) counterfactualFailure(p []NumberedArgumentStep, c NumberedArgumentStep) (string, string) {
	all := append(append([]NumberedArgumentStep(nil), p...), c)
	var source string
	for _, step := range all {
		s := normalizeClause(step.Statement)
		if start := strings.Index(s, "if "); start >= 0 && nonMaterialConditional(s[start:]) {
			source = step.Statement
			break
		}
	}
	if source == "" {
		return "", ""
	}
	mask := func(s string) string {
		return nonMaterialAuxiliary.ReplaceAllStringFunc(s, func(m string) string { return strings.ReplaceAll(m, " ", "_") + "_v19_opaque" })
	}
	virtual := append([]NumberedArgumentStep(nil), p...)
	for i := range virtual {
		virtual[i].Statement = mask(normalizeClausePreservingCase(p[i].Statement))
	}
	conclusion := c
	conclusion.Statement = mask(normalizeClausePreservingCase(c.Statement))
	for _, schema := range policy.schemas() {
		if strings.HasPrefix(schema.name, "counterfactual") || !schemaAcceptsPremiseCount(schema, len(p)) {
			continue
		}
		for _, ordered := range orderedPremiseSelections(virtual, schema.arity) {
			if schema.match(ordered, conclusion).matched {
				return schema.name, counterfactualRuleNote(schema.name, source)
			}
		}
	}
	// Natural moods can conceal the corresponding material shape. Diagnose
	// the invalid VW patterns against their plain operands as well.
	for _, excluded := range []struct {
		name string
		rule v13Rule
	}{
		{"contraposition", v13Rule{v19CF(v13P, v13Q), v19CF(v13Not(v13Q), v13Not(v13P))}},
		{"hypothetical syllogism", v13Rule{v19CF(v13P, v13Q), v19CF(v13Q, v13R), v19CF(v13P, v13R)}},
		{"antecedent strengthening", v13Rule{v19CF(v13P, v13R), v19CF(v13And(v13P, v13Q), v13R)}},
		{"simplification of disjunctive antecedents", v13Rule{v19CF(v13Or(v13P, v13Q), v13R), v19CF(v13P, v13R)}},
		{"strong centering", v13Rule{v13P, v13Q, v19CF(v13P, v13Q)}},
		{"material to counterfactual", v13Rule{v13If(v13P, v13Q), v19CF(v13P, v13Q)}},
		{"material implication", v13Rule{v13Or(v13Not(v13P), v13Q), v19CF(v13P, v13Q)}},
		{"negated conditional", v13Rule{v13Not(v19CF(v13P, v13Q)), v13P}},
		{"negated conditional", v13Rule{v13Not(v19CF(v13P, v13Q)), v13Not(v13Q)}},
	} {
		if len(excluded.rule) != len(p)+1 {
			continue
		}
		for _, ordered := range orderedPremiseSelections(p, len(p)) {
			if policy.v19Rules(ordered, c, false, excluded.rule).matched {
				return excluded.name, counterfactualRuleNote(excluded.name, source)
			}
		}
	}
	// For content mismatches name the two actual clauses, never merely say
	// "not verbatim". Even would V -> Vs requires a visible bridging premise.
	for i, step := range p {
		if len(conditionalCandidatesWithNormalizer(c.Statement, normalizeClause)) > 0 {
			break
		}
		for _, pair := range counterfactualCandidates(step.Statement) {
			target := c.Statement
			clause := pair.right
			if len(p) == 2 {
				other := p[1-i].Statement
				matches := func(surface, plain string) bool {
					for _, candidate := range counterfactualPlain(surface) {
						if policy.equalClauses(candidate, plain) {
							return true
						}
					}
					return false
				}
				if !matches(pair.left, other) {
					negatedOther, negatedConclusion := negationCandidates(other), negationCandidates(c.Statement)
					if len(negatedOther) > 0 && len(negatedConclusion) > 0 {
						clause, target = pair.right, negatedOther[0]
						if matches(pair.right, negatedOther[0]) {
							clause, target = pair.left, negatedConclusion[0]
						}
					} else {
						clause, target = pair.left, other
					}
				}
			}
			return "counterfactual content", fmt.Sprintf("Counterfactual clauses %q and %q are not linked by the permitted auxiliary mapping; a visible bridging premise is needed. No verb morphology or paraphrase matching is performed. %s", clause, target, counterfactualAvailable)
		}
	}
	return "counterfactual rules", counterfactualRuleNote("antecedent strengthening, strong centering, simplification of disjunctive antecedents, conditional excluded middle, or might to would", source)
}
