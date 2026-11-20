package validation

import "strings"

const indefiniteTermMessage = "term.indefinite_singular: a/an noun phrases are ambiguous between generic and existential readings; use a name, every N, or some N"

// Only a determiner followed by a noun phrase is excluded: the arbitrary
// single-letter Fitch constant a and names such as Ann remain available.
func indefiniteSingularTerm(value string) bool {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	return len(words) > 1 && (words[0] == "a" || words[0] == "an")
}

func (policy strictRules) indefiniteTermRefused(value string) bool {
	return policy.term_indefinite_singular && indefiniteSingularTerm(value) && policy.hit("term.indefinite_singular")
}

// Predicate nouns still use legacyFOLTerm, whose article removal is valid for
// predication. Individual slots retain their surface until the formula gate.
func (policy strictRules) legacyFOLConstant(value string) string {
	if policy.term_indefinite_singular && indefiniteSingularTerm(value) {
		return normalizeClause(value)
	}
	return policy.legacyFOLTerm(value)
}

func (policy strictRules) indefiniteFormulaRefused(f *folFormula) bool {
	if f == nil {
		return false
	}
	for _, t := range f.terms {
		if policy.indefiniteTermRefused(t) {
			return true
		}
	}
	for _, child := range f.children {
		if policy.indefiniteFormulaRefused(child) {
			return true
		}
	}
	return false
}

// The non-copular UI fallback copies a verb phrase without relational parsing.
// Do not let that fallback hide an indefinite object after the verb or a
// preposition. Copular complements are handled as predicates elsewhere.
func (policy strictRules) indefinitePredicateObjectRefused(predicate string) bool {
	words := strings.Fields(predicate)
	for i := 1; i < len(words); i++ {
		if policy.indefiniteTermRefused(strings.Join(words[i:], " ")) {
			return true
		}
	}
	return false
}
