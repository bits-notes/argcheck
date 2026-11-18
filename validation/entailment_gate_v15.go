package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Catalog v15 adds one-step system-T schemas, using the existing scoped
// pattern matcher. In particular, neither modal iteration nor K's converse
// is licensed, and an outer conjunction of modals needs the both-that scaffold.
func v15Box(p v13Pattern) v13Pattern { return v13Pattern{"necessity", []v13Pattern{p}} }
func v15Diamond(p v13Pattern) v13Pattern {
	return v13Pattern{"possibility", []v13Pattern{p}}
}

// Match the same optional comma as necessityCandidates. Keep the older
// possibility-introduction matcher unchanged to preserve its catalog boundary.
func possibilityCandidates(statement string) []string {
	s := normalizeClause(statement)
	for _, prefix := range []string{"possibly, ", "possibly "} {
		if strings.HasPrefix(s, prefix) {
			if rest := normalizeClause(strings.TrimPrefix(s, prefix)); rest != "" {
				return []string{rest}
			}
		}
	}
	return nil
}

func matchIdentityReflexivity(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{mismatch: fmt.Sprintf("conclusion %q must repeat the same term occurring as a whole word in premise %q", normalizeClause(c.Statement), normalizeClause(p[0].Statement))}
	for _, pair := range identityCandidates(c.Statement) {
		// Compare the terms verbatim after ordinary clause normalization, never
		// via morphology or substitution. Guards exclude compound pseudo-terms.
		if pair.left != pair.right || !v13PlainClause(pair.left) || !referringTerm(pair.left, false) {
			continue
		}
		contains := regexp.MustCompile(`(?i)(^|[^\pL\pN_])` + regexp.QuoteMeta(pair.left) + `($|[^\pL\pN_])`)
		if contains.MatchString(p[0].Statement) {
			return schemaMatchOutcome{matched: true, traction: true, score: 2}
		}
	}
	return best
}

func (policy strictRules) matchModalDuality(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c,
		v13Rule{v13Not(v15Diamond(v13Not(v13P))), v15Box(v13P)},
		v13Rule{v15Box(v13P), v13Not(v15Diamond(v13Not(v13P)))},
		v13Rule{v13Not(v15Box(v13Not(v13P))), v15Diamond(v13P)},
		v13Rule{v15Diamond(v13P), v13Not(v15Box(v13Not(v13P)))})
}

func (policy strictRules) matchModalDistribution(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.v13MatchRules(p, c,
		v13Rule{v15Box(v13If(v13P, v13Q)), v13If(v15Box(v13P), v15Box(v13Q))},
		v13Rule{v15Box(v13If(v13P, v13Q)), v13If(v15Diamond(v13P), v15Diamond(v13Q))},
		v13Rule{v15Box(v13And(v13P, v13Q)), v13And(v15Box(v13P), v15Box(v13Q))},
		v13Rule{v13And(v15Box(v13P), v15Box(v13Q)), v15Box(v13And(v13P, v13Q))})
}
