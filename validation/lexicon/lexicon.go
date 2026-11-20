// Package lexicon holds generated, immutable lexical allowlists.
package lexicon

import (
	_ "embed"
	"strings"
)

//go:embed plural_mass_nouns.txt
var pluralMassNouns string

//go:embed singular_nouns.txt
var singularNouns string

//go:embed adjectives.txt
var adjectives string

//go:embed mass_nouns.txt
var massNouns string

func makeSet(data string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, line := range strings.Split(data, "\n") {
		if line != "" && !strings.HasPrefix(line, "#") {
			result[line] = struct{}{}
		}
	}
	return result
}

var pluralMassNounSet = makeSet(pluralMassNouns)
var singularNounSet = makeSet(singularNouns)
var adjectiveSet = makeSet(adjectives)
var massNounSet = makeSet(massNouns)

func contains(set map[string]struct{}, word string) bool {
	_, ok := set[strings.ToLower(word)]
	return ok
}

// IsPluralMassNoun reports membership, without inferring inflections.
func IsPluralMassNoun(word string) bool { return contains(pluralMassNounSet, word) }

// IsSingularNoun reports membership in the singular description-head list.
func IsSingularNoun(word string) bool { return contains(singularNounSet, word) }

// IsAdjective reports membership in the adjective and participle list.
func IsAdjective(word string) bool { return contains(adjectiveSet, word) }

// IsMassNoun distinguishes reviewed mass nouns for slot agreement.
func IsMassNoun(word string) bool                  { return contains(massNounSet, word) }
func isAllowlistedPluralMassNoun(word string) bool { return IsPluralMassNoun(word) }

//go:embed names.txt
var names string

//go:embed name_forbidden.txt
var nameForbidden string

var nameSet = makeSet(names)
var nameForbiddenSet = makeSet(nameForbidden)

// IsName reports lexical membership; callers must separately require capitals.
func IsName(word string) bool { return contains(nameSet, word) }

// IsNameForbidden vetoes quantity and closed words even with contextual capitals.
func IsNameForbidden(word string) bool { return contains(nameForbiddenSet, word) }
