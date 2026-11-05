package validation

import "strings"

func v21PropertyForm(name string) bool {
	switch name {
	case "property instantiation (arbitrary property)", "distinguishing property", "distinguishing property nonidentity", "distinguishing property non-inclusion":
		return true
	}
	return false
}

// Reconstruct the enclosing blocks, including sibling closures. Depth alone
// cannot establish ownership. The public gate validates the scope first.
func enclosingPropertyNames(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) map[string]bool {
	active := []NumberedArgumentStep{}
	for n := 1; n < c.Number; n++ {
		s := steps[n]
		assumption := s.Kind == numberedArgumentStepAssumption || s.Kind == numberedArgumentStepArbitrary
		for len(active) > 0 && (active[len(active)-1].Depth > s.Depth || (assumption && active[len(active)-1].Depth == s.Depth)) {
			active = active[:len(active)-1]
		}
		if assumption {
			active = append(active, s)
		}
	}
	names := map[string]bool{}
	for _, s := range active {
		if s.Depth <= c.Depth && s.Kind == numberedArgumentStepArbitrary {
			if m := arbitraryPropertyPattern.FindStringSubmatch(s.Statement); m != nil && propertyBearer(m[2]) {
				names[m[1]] = true
			}
		}
	}
	return names
}

// Preserve case for the eigenname; only its bearer uses clause normalization.
// A bare name never becomes a property merely because it occurs after "has".
func arbitraryPropertyPossessions(s string, names map[string]bool) []clausePair {
	s = normalizeClausePreservingCase(s)
	var result []clausePair
	for name := range names {
		suffix := " has " + name
		if strings.HasSuffix(s, suffix) {
			bearer := strings.TrimSuffix(s, suffix)
			if propertyBearer(bearer) {
				result = append(result, clausePair{normalizeClause(bearer), name})
			}
		}
	}
	return result
}

func matchNoPropertyScope(_ []NumberedArgumentStep, _ NumberedArgumentStep) schemaMatchOutcome {
	return schemaMatchOutcome{}
}

func matchArbitraryPropertyInstantiation(p []NumberedArgumentStep, c NumberedArgumentStep, steps map[int]NumberedArgumentStep) schemaMatchOutcome {
	names := enclosingPropertyNames(c, steps)
	for _, inclusion := range propertyInclusions(p[0].Statement) {
		for _, source := range arbitraryPropertyPossessions(p[1].Statement, names) {
			for _, target := range arbitraryPropertyPossessions(c.Statement, names) {
				if inclusion.right == source.left && inclusion.left == target.left && source.right == target.right {
					return schemaMatchOutcome{matched: true}
				}
			}
		}
	}
	return schemaMatchOutcome{}
}

func distinguishingProperties(s string) []clausePair {
	var result []clausePair
	for _, pair := range splitPropertySurface(s, " has some property that ") {
		if strings.HasSuffix(pair.right, " lacks") {
			b := strings.TrimSuffix(pair.right, " lacks")
			if propertyBearer(pair.left) && propertyBearer(b) {
				result = append(result, clausePair{pair.left, b})
			}
		}
	}
	return result
}

func matchDistinguishingProperty(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return matchScopedDistinguishingProperty(p, c, nil)
}

func matchScopedDistinguishingProperty(p []NumberedArgumentStep, c NumberedArgumentStep, steps map[int]NumberedArgumentStep) schemaMatchOutcome {
	// Check the canonical negation without lowercasing its property name.
	negative := normalizeClausePreservingCase(p[1].Statement)
	const prefix = "it is not the case that "
	if !strings.HasPrefix(strings.ToLower(negative), prefix) {
		return schemaMatchOutcome{}
	}
	negative = negative[len(prefix):]
	names := enclosingPropertyNames(c, steps)
	// Keep named predicates and arbitrary properties separate: an eigenname
	// "wise" must not be identified with "the property of being wise".
	for _, possessions := range []func(string) []clausePair{
		propertyPossessions,
		func(s string) []clausePair { return arbitraryPropertyPossessions(s, names) },
	} {
		for _, a := range possessions(p[0].Statement) {
			for _, b := range possessions(negative) {
				for _, target := range distinguishingProperties(c.Statement) {
					if a.right == b.right && target == (clausePair{a.left, b.left}) {
						return schemaMatchOutcome{matched: true}
					}
				}
			}
		}
	}
	return schemaMatchOutcome{}
}

func matchDistinguishingPropertyNonidentity(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	for _, pair := range distinguishingProperties(p[0].Statement) {
		if normalizeClause(c.Statement) == "it is not the case that "+pair.left+" is identical to "+pair.right {
			return schemaMatchOutcome{matched: true}
		}
	}
	return schemaMatchOutcome{}
}

func matchDistinguishingPropertyNonInclusion(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	for _, pair := range distinguishingProperties(p[0].Statement) {
		if normalizeClause(c.Statement) == "it is not the case that "+pair.right+" has every property that "+pair.left+" has" {
			return schemaMatchOutcome{matched: true}
		}
	}
	return schemaMatchOutcome{}
}
