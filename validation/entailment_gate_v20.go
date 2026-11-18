package validation

import (
	"regexp"
	"strings"
)

// Property terms name unary predicates; they are never individual bearers.
// These matchers do not recursively rewrite content or normalize synonyms.
const propertyTermPrefix = "the property of being "

var arbitraryPropertyPattern = regexp.MustCompile(`^Let ([A-Za-z][A-Za-z0-9_]*) be an arbitrary property that (.+) has$`)

func propertyBearer(s string) bool {
	s = normalizeClause(s)
	if strings.HasPrefix(s, propertyTermPrefix) {
		return false
	}
	// Definite that-clause subjects use the same whole-term boundary as v16.
	if thatNounPhrase(s) {
		return true
	}
	// Possessive descriptions are singular terms, including the design's
	// "his body". This does not resolve the pronoun or identify its referent.
	for _, prefix := range []string{"his ", "her ", "its ", "their "} {
		s = strings.TrimPrefix(s, prefix)
	}
	return referringTerm(s, false) && v13PlainClause(s)
}

// The complete bearer and predicate must be shared. In particular, a copula
// within a definite description cannot be mistaken for its outer copula.
func (policy strictRules) propertyCopulas(s string) []clausePair {
	var result []clausePair
	for _, pair := range policy.groundInstanceCandidates(s) {
		if propertyBearer(pair.left) && propertyPredicate(pair.right) {
			result = append(result, pair)
		}
	}
	return result
}

func propertyPredicate(s string) bool {
	if s == "" {
		return false
	}
	// A that-clause can be part of F: naming it does not rewrite its content.
	// Unscaffolded Boolean/conditional boundaries are deliberately refused.
	head := strings.SplitN(s, " that ", 2)[0]
	if strings.ContainsAny(head, ",;:\"`()") {
		return false
	}
	for _, word := range strings.Fields(head) {
		switch word {
		case "and", "or", "if", "then", "iff":
			return false
		}
	}
	return true
}

func propertyPossessions(s string) []clausePair {
	s = normalizeClause(s)
	var result []clausePair
	for _, pair := range splitPropertySurface(s, " has ") {
		if propertyBearer(pair.left) && strings.HasPrefix(pair.right, propertyTermPrefix) {
			predicate := strings.TrimPrefix(pair.right, propertyTermPrefix)
			if propertyPredicate(predicate) {
				result = append(result, clausePair{pair.left, predicate})
			}
		}
	}
	return result
}

func splitPropertySurface(s, separator string) []clausePair {
	var result []clausePair
	return appendSplitCandidates(result, normalizeClause(s), separator, false)
}

func propertyInclusions(s string) []clausePair {
	var result []clausePair
	for _, pair := range splitPropertySurface(s, " has every property that ") {
		if !strings.HasSuffix(pair.right, " has") {
			continue
		}
		b := strings.TrimSuffix(pair.right, " has")
		if propertyBearer(pair.left) && propertyBearer(b) {
			result = append(result, clausePair{pair.left, b})
		}
	}
	return result
}

// Exactly one canonical outer negation is allowed, on both sides together.
func propertyOuterNegation(s string) (string, bool) {
	s = normalizeClause(s)
	const prefix = "it is not the case that "
	return strings.TrimPrefix(s, prefix), strings.HasPrefix(s, prefix)
}

func (policy strictRules) matchPropertyAbstraction(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	a, an := propertyOuterNegation(p[0].Statement)
	b, bn := propertyOuterNegation(c.Statement)
	if an != bn {
		return schemaMatchOutcome{}
	}
	for _, pred := range policy.propertyCopulas(a) {
		for _, prop := range propertyPossessions(b) {
			if pred == prop {
				return schemaMatchOutcome{matched: true}
			}
		}
	}
	return schemaMatchOutcome{}
}

func (policy strictRules) matchPropertyConcretion(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return policy.matchPropertyAbstraction([]NumberedArgumentStep{c}, p[0])
}

func matchPropertyInstantiation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	for _, inclusion := range propertyInclusions(p[0].Statement) {
		for _, source := range propertyPossessions(p[1].Statement) {
			for _, target := range propertyPossessions(c.Statement) {
				if inclusion.right == source.left && inclusion.left == target.left && source.right == target.right {
					return schemaMatchOutcome{matched: true}
				}
			}
		}
	}
	return schemaMatchOutcome{}
}

func matchPropertyExistential(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	for _, prop := range propertyPossessions(p[0].Statement) {
		if normalizeClause(c.Statement) == prop.left+" has some property" {
			return schemaMatchOutcome{matched: true}
		}
	}
	return schemaMatchOutcome{}
}

func matchSharedProperty(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	for _, a := range propertyPossessions(p[0].Statement) {
		for _, b := range propertyPossessions(p[1].Statement) {
			if a.right == b.right && normalizeClause(c.Statement) == a.left+" shares some property with "+b.left {
				return schemaMatchOutcome{matched: true}
			}
		}
	}
	return schemaMatchOutcome{}
}

func matchIndiscernibilityOfIdenticals(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	for _, identity := range splitPropertySurface(p[0].Statement, " is identical to ") {
		if !propertyBearer(identity.left) || !propertyBearer(identity.right) {
			continue
		}
		for _, inclusion := range propertyInclusions(c.Statement) {
			if identity == inclusion {
				return schemaMatchOutcome{matched: true}
			}
		}
	}
	return schemaMatchOutcome{}
}

func (policy strictRules) propertyGeneralization(first, last, c NumberedArgumentStep, steps map[int]NumberedArgumentStep) (bool, string) {
	last.Statement = fitchAssumptionStatement(last)
	m := arbitraryPropertyPattern.FindStringSubmatch(first.Statement)
	if m == nil || !propertyBearer(m[2]) {
		return false, "arbitrary property must have a singular individual bearer"
	}
	if reason := policy.fitchFreshName(m[1], first.Number, c, steps); reason != "" {
		return false, reason
	}
	for _, inclusion := range propertyInclusions(c.Statement) {
		// Keep the eigenname case-sensitive, just as individual UG does. The
		// bearer is compared by ordinary clause normalization, without paraphrase.
		suffix := " has " + m[1]
		if inclusion.right == normalizeClause(m[2]) && strings.HasSuffix(last.Statement, suffix) &&
			normalizeClause(strings.TrimSuffix(last.Statement, suffix)) == inclusion.left {
			return true, ""
		}
	}
	return false, "final possession and property inclusion must match the arbitrary property and its bearer verbatim"
}

func v20PropertyForm(name string) bool {
	switch name {
	case "property abstraction", "property concretion", "property instantiation", "property existential", "shared property", "indiscernibility of identicals":
		return true
	}
	return false
}
