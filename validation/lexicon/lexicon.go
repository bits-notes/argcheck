// Package lexicon holds the v46 noun allowlist independently of validation rules.
package lexicon

import (
	_ "embed"
	"strings"
)

//go:embed plural_mass_nouns.txt
var pluralMassNouns string

// Construct once during package initialization. The map is never mutated after
// initialization, so concurrent lookups need no synchronization.
var pluralMassNounSet = func() map[string]struct{} {
	words := make(map[string]struct{})
	for _, line := range strings.Split(pluralMassNouns, "\n") {
		if line != "" && !strings.HasPrefix(line, "#") {
			words[line] = struct{}{}
		}
	}
	return words
}()

func isAllowlistedPluralMassNoun(word string) bool {
	_, ok := pluralMassNounSet[strings.ToLower(word)]
	return ok
}
