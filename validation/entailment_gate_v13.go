package validation

import (
	"fmt"
	"strings"
)

// Catalog v13 is appended after v12. These are one-step schemas, not a
// simplifier or a recursive proof search: each rule shares verbatim clauses
// across its premises and conclusion. Explosion and arbitrary tautology
// introduction are deliberately absent.
//
// Patterns describe schemas only. All English splitting delegates to the
// existing gate helpers, with conservative scope guards for nested connectives.
type v13Pattern struct {
	op       string
	children []v13Pattern
}

type v13Rule []v13Pattern // premises, then conclusion

var v13P = v13Pattern{op: "P"}
var v13Q = v13Pattern{op: "Q"}
var v13R = v13Pattern{op: "R"}

func v13Not(p v13Pattern) v13Pattern    { return v13Pattern{"negation", []v13Pattern{p}} }
func v13And(p, q v13Pattern) v13Pattern { return v13Pattern{"conjunction", []v13Pattern{p, q}} }
func v13Or(p, q v13Pattern) v13Pattern  { return v13Pattern{"disjunction", []v13Pattern{p, q}} }
func v13If(p, q v13Pattern) v13Pattern  { return v13Pattern{"conditional", []v13Pattern{p, q}} }
func v13Iff(p, q v13Pattern) v13Pattern { return v13Pattern{"biconditional", []v13Pattern{p, q}} }

// v13PlainClause prevents an unscaffolded connective from being lifted out
// of negation, a modal/report context, or a larger connective. Compound
// clauses remain usable as opaque variables, or inside explicit scaffolds.
func v13PlainClause(s string) bool {
	s = normalizeClause(s)
	if s == "" || strings.ContainsAny(s, ",;:\"`()") || opaqueContextWords.MatchString(s) {
		return false
	}
	for _, word := range strings.Fields(s) {
		switch word {
		case "if", "then", "and", "or", "either":
			return false
		}
	}
	return true
}

func v13Pairs(op, statement string) []clausePair {
	s := normalizeClause(statement)
	switch op {
	case "conditional":
		// The first then closes the outer antecedent. Later then tokens belong
		// to a nested consequent, never to an alternative outer split.
		pairs := conditionalCandidates(s)
		if len(pairs) > 0 {
			return pairs[:1]
		}
	case "biconditional":
		pairs := biconditionalCandidates(s)
		if len(pairs) > 0 && (isScaffoldedBiconditional(s) || (v13PlainClause(pairs[0].left) && v13PlainClause(pairs[0].right))) {
			return pairs
		}
	case "disjunction":
		if strings.HasPrefix(s, "either ") && strings.Count(s, " or ") == 1 {
			return disjunctionCandidates(s)
		}
	case "conjunction":
		pairs := conjunctionCandidates(s)
		if strings.HasPrefix(s, "it is the case both that ") && strings.Count(s, ", and that ") == 1 && len(pairs) > 0 {
			return pairs[:1]
		}
		if len(pairs) == 1 && v13PlainClause(pairs[0].left) && v13PlainClause(pairs[0].right) {
			return pairs
		}
	}
	return nil
}

func (policy strictRules) v13MatchPattern(pattern v13Pattern, statement, label string, bindings map[string]string) schemaMatchOutcome {
	if len(pattern.children) == 0 {
		if bound, ok := bindings[pattern.op]; ok {
			return policy.evaluateIdentities(clauseIdentity{label, statement, "bound clause " + pattern.op, bound})
		}
		bindings[pattern.op] = statement
		return schemaMatchOutcome{matched: true, traction: true}
	}
	var parts []string
	switch pattern.op {
	case "negation":
		parts = negationCandidates(statement)
	case "necessity":
		parts = necessityCandidates(statement)
	case "possibility":
		parts = possibilityCandidates(statement)
	default:
		pairs := v13Pairs(pattern.op, statement)
		if len(pairs) > 0 {
			parts = []string{pairs[0].left, pairs[0].right}
		}
	}
	if len(parts) != len(pattern.children) {
		return schemaMatchOutcome{mismatch: fmt.Sprintf("%s %q is not an explicitly scoped %s", label, normalizeClause(statement), pattern.op)}
	}
	outcome := schemaMatchOutcome{matched: true, traction: true}
	for i, child := range pattern.children {
		next := policy.v13MatchPattern(child, parts[i], fmt.Sprintf("%s %s clause %d", label, pattern.op, i+1), bindings)
		outcome.matched = outcome.matched && next.matched
		appendOutcome(&outcome, next)
	}
	return outcome
}

func (policy strictRules) v13MatchRules(premises []NumberedArgumentStep, conclusion NumberedArgumentStep, rules ...v13Rule) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, rule := range rules {
		if len(rule) != len(premises)+1 {
			continue
		}
		bindings := map[string]string{}
		outcome := schemaMatchOutcome{matched: true}
		for i, pattern := range rule {
			step, label := conclusion, fmt.Sprintf("conclusion %d", conclusion.Number)
			if i < len(premises) {
				step = premises[i]
				label = fmt.Sprintf("premise %d", step.Number)
			}
			next := policy.v13MatchPattern(pattern, step.Statement, label, bindings)
			outcome.matched = outcome.matched && next.matched
			appendOutcome(&outcome, next)
		}
		if outcome.matched {
			outcome.reading = readingPatternRule(rule, bindings)
		}
		best = betterOutcome(best, outcome)
		if best.mismatch == "" && outcome.mismatch != "" {
			best.mismatch = outcome.mismatch
		}
		if best.matched {
			return best
		}
	}
	return best
}

func (policy strictRules) matchContraposition(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c,
		v13Rule{v13If(v13P, v13Q), v13If(v13Not(v13Q), v13Not(v13P))},
		v13Rule{v13If(v13Not(v13Q), v13Not(v13P)), v13If(v13P, v13Q)})
}

func (policy strictRules) matchDeMorgan(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c,
		v13Rule{v13Not(v13And(v13P, v13Q)), v13Or(v13Not(v13P), v13Not(v13Q))},
		v13Rule{v13Or(v13Not(v13P), v13Not(v13Q)), v13Not(v13And(v13P, v13Q))},
		v13Rule{v13Not(v13Or(v13P, v13Q)), v13And(v13Not(v13P), v13Not(v13Q))},
		v13Rule{v13And(v13Not(v13P), v13Not(v13Q)), v13Not(v13Or(v13P, v13Q))})
}

func (policy strictRules) matchCommutation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13And(v13P, v13Q), v13And(v13Q, v13P)}, v13Rule{v13Or(v13P, v13Q), v13Or(v13Q, v13P)})
}

func (policy strictRules) matchExportation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13If(v13And(v13P, v13Q), v13R), v13If(v13P, v13If(v13Q, v13R))})
}

func (policy strictRules) matchImportation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13If(v13P, v13If(v13Q, v13R)), v13If(v13And(v13P, v13Q), v13R)})
}

func (policy strictRules) matchMaterialImplication(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13If(v13P, v13Q), v13Or(v13Not(v13P), v13Q)}, v13Rule{v13Or(v13Not(v13P), v13Q), v13If(v13P, v13Q)})
}

func (policy strictRules) matchDistribution(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c,
		v13Rule{v13And(v13P, v13Or(v13Q, v13R)), v13Or(v13And(v13P, v13Q), v13And(v13P, v13R))},
		v13Rule{v13Or(v13And(v13P, v13Q), v13And(v13P, v13R)), v13And(v13P, v13Or(v13Q, v13R))},
		v13Rule{v13Or(v13P, v13And(v13Q, v13R)), v13And(v13Or(v13P, v13Q), v13Or(v13P, v13R))},
		v13Rule{v13And(v13Or(v13P, v13Q), v13Or(v13P, v13R)), v13Or(v13P, v13And(v13Q, v13R))})
}

func (policy strictRules) matchAbsorption(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13Or(v13P, v13And(v13P, v13Q)), v13P}, v13Rule{v13If(v13P, v13Q), v13If(v13P, v13And(v13P, v13Q))})
}

func (policy strictRules) matchDoubleNegationIntroduction(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13P, v13Not(v13Not(v13P))})
}

func (policy strictRules) matchBiconditionalModusTollens(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13Iff(v13P, v13Q), v13Not(v13Q), v13Not(v13P)}, v13Rule{v13Iff(v13P, v13Q), v13Not(v13P), v13Not(v13Q)})
}

func (policy strictRules) matchBiconditionalChaining(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13Iff(v13P, v13Q), v13Iff(v13Q, v13R), v13Iff(v13P, v13R)})
}

func (policy strictRules) matchBiconditionalCommutation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13Iff(v13P, v13Q), v13Iff(v13Q, v13P)})
}

func (policy strictRules) matchNegatedConditional(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c, v13Rule{v13Not(v13If(v13P, v13Q)), v13P}, v13Rule{v13Not(v13If(v13P, v13Q)), v13Not(v13Q)})
}

func (policy strictRules) matchIdentitySymmetry(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{mismatch: fmt.Sprintf("conclusion %q is not an identity with the same two terms reversed for premise %q", normalizeClause(c.Statement), normalizeClause(p[0].Statement))}
	for _, from := range identityCandidates(p[0].Statement) {
		for _, to := range identityCandidates(c.Statement) {
			if !v13PlainClause(from.left) || !v13PlainClause(from.right) || !v13PlainClause(to.left) || !v13PlainClause(to.right) {
				continue
			}
			if !referringTerm(from.left, false) || !referringTerm(from.right, false) || !referringTerm(to.left, false) || !referringTerm(to.right, false) {
				continue
			}
			best = betterOutcome(best, policy.evaluateIdentities(
				clauseIdentity{"premise identity subject", from.left, "conclusion identity object", to.right},
				clauseIdentity{"premise identity object", from.right, "conclusion identity subject", to.left}))
		}
	}
	return best
}

// Quantified/categorical forms stay in the existing copular register. Reject
// embedded connectives/copulas rather than mistake part of a predicate for a
// new class. equalTerms supplies the gate's existing a/an and number handling.
func v13ClassPair(pair clausePair) bool {
	return v13PlainClause(pair.left) && v13PlainClause(pair.right) && len(instanceCandidates(pair.left)) == 0 && len(instanceCandidates(pair.right)) == 0
}

func (policy strictRules) v13CompareClasses(from, to clausePair, swap bool) schemaMatchOutcome {
	if !v13ClassPair(from) || !v13ClassPair(to) {
		return schemaMatchOutcome{}
	}
	if swap {
		to = clausePair{to.right, to.left}
	}
	return policy.evaluateTermIdentities(
		clauseIdentity{"premise class", from.left, "conclusion class", to.left},
		clauseIdentity{"premise predicate", from.right, "conclusion predicate", to.right})
}

func (policy strictRules) matchQuantifierNegation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{mismatch: fmt.Sprintf("conclusion %q is not the corresponding particular negative or universal negative of the negated quantifier for premise %q", normalizeClause(c.Statement), normalizeClause(p[0].Statement))}
	for _, negated := range negationCandidates(p[0].Statement) {
		for _, pair := range universalCandidates(negated) {
			// The ledger's affirmative universal only: do not read 'not B' as B.
			if strings.HasPrefix(pair.right, "not ") {
				continue
			}
			for _, result := range particularNegativeCandidates(c.Statement) {
				best = betterOutcome(best, policy.v13CompareClasses(pair, result, false))
			}
		}
		for _, pair := range particularAffirmativeCandidates(negated) {
			if !strings.HasPrefix(normalizeClause(c.Statement), "no ") {
				continue
			}
			for _, result := range policy.universalNegativeCandidates(c.Statement) {
				best = betterOutcome(best, policy.v13CompareClasses(pair, result, false))
			}
		}
	}
	return best
}

func (policy strictRules) matchConversion(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{mismatch: fmt.Sprintf("conclusion %q is not a particular affirmative or no-class statement with subject and predicate reversed for premise %q", normalizeClause(c.Statement), normalizeClause(p[0].Statement))}
	parsers := []func(string) []clausePair{particularAffirmativeCandidates, policy.universalNegativeCandidates}
	for i, parse := range parsers {
		if i == 1 && (!strings.HasPrefix(normalizeClause(p[0].Statement), "no ") || !strings.HasPrefix(normalizeClause(c.Statement), "no ")) {
			continue
		}
		for _, pair := range parse(p[0].Statement) {
			for _, result := range parse(c.Statement) {
				best = betterOutcome(best, policy.v13CompareClasses(pair, result, true))
			}
		}
	}
	return best
}

func (policy strictRules) matchUniversalModusTollens(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{mismatch: fmt.Sprintf("conclusion %q is not the same denied instance of the universal class for premise %q", normalizeClause(c.Statement), normalizeClause(p[0].Statement))}
	for _, universal := range universalCandidates(p[0].Statement) {
		if !v13ClassPair(universal) || strings.HasPrefix(universal.right, "not ") {
			continue
		}
		for _, denied := range negationCandidates(p[1].Statement) {
			for _, instance := range policy.groundInstanceCandidates(denied) {
				if !v13ClassPair(instance) || quantifiedInstanceSubject(instance.left) {
					continue
				}
				for _, resultNegated := range negationCandidates(c.Statement) {
					for _, result := range policy.groundInstanceCandidates(resultNegated) {
						if !v13ClassPair(result) || quantifiedInstanceSubject(result.left) {
							continue
						}
						candidate := policy.evaluateTermIdentities(
							clauseIdentity{"universal predicate", universal.right, "denied instance predicate", instance.right},
							clauseIdentity{"universal class", universal.left, "denied conclusion class", result.right})
						subject := policy.evaluateIdentities(clauseIdentity{"premise instance", instance.left, "conclusion instance", result.left})
						candidate.matched = candidate.matched && subject.matched
						appendOutcome(&candidate, subject)
						best = betterOutcome(best, candidate)
					}
				}
			}
		}
	}
	return best
}

func (policy strictRules) matchPossibilityIntroduction(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	s := normalizeClause(c.Statement)
	const prefix = "possibly, "
	if !strings.HasPrefix(s, prefix) || len(s) == len(prefix) {
		return schemaMatchOutcome{mismatch: fmt.Sprintf("conclusion %q does not begin with Possibly, followed by a clause", s)}
	}
	result := s[len(prefix):]
	candidates := append([]string{p[0].Statement}, necessityCandidates(p[0].Statement)...)
	best := schemaMatchOutcome{}
	for _, clause := range candidates {
		best = betterOutcome(best, policy.evaluateIdentities(clauseIdentity{"premise clause", clause, "possible conclusion clause", result}))
	}
	return best
}
