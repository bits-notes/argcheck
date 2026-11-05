package validation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// A counterfactual entails its material counterpart (weak centering), not
// conversely. Replacement is therefore safe in positive premise positions
// and negative conclusion positions. A biconditional uses each side both
// ways; a counterfactual antecedent is nonmonotone, so nested occurrences there
// are also visited both ways. Bare would clauses are ordinary atomic clauses.
type v19Occurrence struct {
	statement string
	positive  bool
	kind      gateConditionalKind
}

func v19Unparenthesize(s string) string {
	for strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		depth := 0
		whole := true
		for i, r := range s {
			if r == '(' {
				depth++
			}
			if r == ')' {
				depth--
			}
			if depth == 0 && i < len(s)-1 {
				whole = false
				break
			}
		}
		if !whole || depth != 0 {
			break
		}
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

// Only a then outside parentheses closes this conditional's antecedent.
// The legacy candidate enumerator remains unchanged for compatibility.
func v19ConditionalPair(s string) []clausePair {
	s = v19Unparenthesize(stripClauseWrappers(s))
	if !strings.HasPrefix(s, "if ") {
		return nil
	}
	body := s[3:]
	for _, index := range conditionalSplitPattern.FindAllStringIndex(body, -1) {
		depth := 0
		for _, r := range body[:index[0]] {
			if r == '(' {
				depth++
			}
			if r == ')' {
				depth--
			}
		}
		if depth == 0 {
			return []clausePair{{normalizeClause(body[:index[0]]), normalizeClause(body[index[1]:])}}
		}
	}
	return nil
}

// The polarity parser follows explicit scope. It never scans modal words
// globally, which would turn standalone would-clauses into connectives.
func v19PolarityPairs(op, s string) []clausePair {
	switch op {
	case "conjunction":
		if strings.HasPrefix(s, "it is the case both that ") {
			if pairs := conjunctionCandidates(s); len(pairs) > 0 {
				return pairs[:1]
			}
		}
		if !hasOuterGateOperator(s) && !strings.HasPrefix(s, "either ") {
			if pairs := conjunctionCandidates(s); len(pairs) == 1 {
				return pairs
			}
		}
	case "disjunction":
		if strings.HasPrefix(s, "either ") {
			if pairs := disjunctionCandidates(s); len(pairs) > 0 {
				return pairs[:1]
			}
		}
		if !hasOuterGateOperator(s) {
			if pairs := disjunctionCandidates(s); len(pairs) == 1 {
				return pairs
			}
		}
	}
	return nil
}
func v19Visit(s string, positive bool, visit func(v19Occurrence)) {
	s = v19Unparenthesize(stripClauseWrappers(s))
	if neg := negationCandidates(s); len(neg) > 0 {
		v19Visit(neg[0], !positive, visit)
		return
	}
	// Modal operators are monotone. Their atoms are not mood rewriting sites.
	for _, parts := range [][]string{necessityCandidates(s), possibilityCandidates(s)} {
		if len(parts) > 0 {
			v19Visit(parts[0], positive, visit)
			return
		}
	}
	if pairs := biconditionalCandidates(s); len(pairs) > 0 {
		for _, pair := range pairs {
			for _, side := range []string{pair.left, pair.right} {
				v19Visit(side, true, visit)
				v19Visit(side, false, visit)
			}
		}
		return
	}
	// Explicit Boolean scaffolds take precedence over embedded conditionals.
	for _, op := range []string{"conjunction", "disjunction"} {
		if pairs := v19PolarityPairs(op, s); len(pairs) > 0 {
			v19Visit(pairs[0].left, positive, visit)
			v19Visit(pairs[0].right, positive, visit)
			return
		}
	}
	pairs := v19ConditionalPair(s)
	if len(pairs) == 0 {
		return
	}
	// As in v13, the first then closes an unparenthesized outer antecedent.
	pair := pairs[0]
	kind := conditionalKind(pair.right)
	if kind != materialConditional {
		visit(v19Occurrence{s, positive, kind})
	}
	v19Visit(pair.left, !positive, visit)
	if kind != materialConditional {
		v19Visit(pair.left, positive, visit)
	}
	v19Visit(pair.right, positive, visit)
}
func v19PolarityFailure(p []NumberedArgumentStep, c NumberedArgumentStep) string {
	check := func(s string, conclusion bool) string {
		// Most legacy oracle/corpus steps contain no such connective at all.
		// This fast path changes no parse and avoids reparsing ordinary atoms.
		lower := strings.ToLower(s)
		if !strings.Contains(lower, "if") || !(strings.Contains(lower, "would") || strings.Contains(lower, "might") || strings.Contains(lower, "could")) {
			return ""
		}
		reason := ""
		v19Visit(s, true, func(o v19Occurrence) {
			if reason != "" {
				return
			}
			if o.kind == mightConditional {
				reason = fmt.Sprintf("%q is a might/could-have conditional; no material weakening is available", o.statement)
			} else if o.positive == conclusion {
				if conclusion {
					reason = fmt.Sprintf("%q occurs in a positive conclusion position; a material conditional does not entail a counterfactual", o.statement)
				} else {
					reason = fmt.Sprintf("%q occurs in a negative premise position; weakening there reverses the required entailment", o.statement)
				}
			}
		})
		return reason
	}
	for _, step := range p {
		if reason := check(step.Statement, false); reason != "" {
			return reason
		}
	}
	return check(c.Statement, true)
}

// An unchanged conditional may also be used as a whole proposition, even in
// negative positions (e.g. reiteration or contradiction). Such a certificate
// does not weaken it. Mask whole non-material conditionals consistently and
// ask the same schema to certify that purely opaque use. A fresh atom includes
// no copula or variable slot, so this cannot manufacture quantifier structure.
func v19OpaqueStep(p []NumberedArgumentStep, c NumberedArgumentStep, match func([]NumberedArgumentStep, NumberedArgumentStep) schemaMatchOutcome) bool {
	atoms := map[string]string{}
	prefix := "opaque_counterfactual_"
	all := c.Statement
	for _, step := range p {
		all += " " + step.Statement
	}
	for strings.Contains(strings.ToLower(all), prefix) {
		prefix += "fresh_"
	}
	// Mask exact occurrences, smallest first. This preserves the legacy
	// premise-guided grouping of an unscaffolded n-ary conjunction, including
	// a conditional conjunct, rather than imposing a new global parse.
	var occurrences []string
	seen := map[string]bool{}
	collect := func(statement string) {
		v19Visit(statement, true, func(o v19Occurrence) {
			if !seen[o.statement] {
				seen[o.statement] = true
				occurrences = append(occurrences, o.statement)
			}
		})
	}
	for _, step := range p {
		collect(step.Statement)
	}
	collect(c.Statement)
	sort.Slice(occurrences, func(i, j int) bool { return len(occurrences[i]) < len(occurrences[j]) })
	for _, statement := range occurrences {
		key := comparisonClause(statement)
		if atoms[key] == "" {
			atoms[key] = fmt.Sprintf("%s%d", prefix, len(atoms))
		}
	}
	mask := func(s string) string {
		s = normalizeClausePreservingCase(s)
		for _, statement := range occurrences {
			s = regexp.MustCompile("(?i)"+regexp.QuoteMeta(statement)).ReplaceAllString(s, atoms[comparisonClause(statement)])
		}
		return s
	}
	virtual := append([]NumberedArgumentStep(nil), p...)
	for i := range virtual {
		virtual[i].Statement = mask(virtual[i].Statement)
	}
	c.Statement = mask(c.Statement)
	return len(atoms) > 0 && match(virtual, c).matched
}

// Instantiation may copy a schematic cf as an opaque consequent. Instantiate
// the visible material schema first and compare both whole operands verbatim.
// This does not derive a cf by materializing its connective. In particular,
// do not reparse the instantiated antecedent: a noun phrase substituted for X
// may itself contain "if ... then ...".
func v19OpaqueInstantiation(p []NumberedArgumentStep, c NumberedArgumentStep) bool {
	if len(p) != 2 {
		return false
	}
	for _, pair := range conditionalCandidatesPreservingCase(p[0].Statement) {
		if conditionalKind(pair.right) != materialConditional {
			continue
		}
		ant, _ := schematicTemplate(pair.left)
		result, _ := schematicTemplate(pair.right)
		for _, sub := range unifySchematicTemplate(ant, p[1].Statement, nil) {
			a, okA := instantiateSchematicTemplate(ant, sub)
			b, okB := instantiateSchematicTemplate(result, sub)
			if !okA || !okB || !equalClauses(b, c.Statement) {
				continue
			}
			if equalClauses(a, p[1].Statement) {
				return true
			}
		}
	}
	return false
}

func init() {
	for i := range entailmentSchemas {
		schema := entailmentSchemas[i]
		if strings.HasPrefix(schema.name, "counterfactual") {
			continue
		}
		entailmentSchemas[i].match = func(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
			outcome := schema.match(p, c)
			if !outcome.matched {
				return outcome
			}
			reason := v19PolarityFailure(p, c)
			if reason == "" || v19OpaqueStep(p, c, schema.match) || (schema.name == "schematic instantiation" && v19OpaqueInstantiation(p, c)) {
				return outcome
			}
			return schemaMatchOutcome{refusal: true, traction: true, score: outcome.score, mismatch: schema.name + " refused: " + reason + ". " + counterfactualAvailable}
		}
	}
}
