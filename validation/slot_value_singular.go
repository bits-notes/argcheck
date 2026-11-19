package validation

import (
	"strings"
	"unicode"
)

// Retain case evidence before comparisonClause folds it away. Evidence is
// evaluation-local, shared by normal matching and one-rule diagnostic retries.
func internalCapital(w string) bool {
	for i, r := range w {
		if i > 0 && unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func schematicNameEvidence(steps map[int]NumberedArgumentStep) *map[string]bool {
	names := map[string]bool{}
	for _, step := range steps {
		words := strings.Fields(step.Statement)
		for i, raw := range words {
			w := strings.Trim(raw, ".,;:!?\"()[]")
			// A capital after the first letter (AIXI, McCarthy, GPT-4) is name
			// evidence even at sentence start: an ordinary word written there
			// ("Lots", "Most") has only its first letter capitalised.
			base := strings.TrimSuffix(strings.TrimSuffix(w, "'s"), "’s")
			possessiveName := base != w && unicode.IsUpper(firstRune(base)) && !scopeContractions[strings.ToLower(base)]
			if possessiveName {
				names["possessive:"+strings.ToLower(base)] = true
			}
			if internalCapital(w) {
				names[strings.ToLower(w)] = true
				names[strings.TrimSuffix(strings.TrimSuffix(strings.ToLower(w), "'s"), "’s")] = true
				continue
			}
			if i == 0 || strings.ContainsAny(words[i-1], ".!?") {
				continue
			}
			prev := strings.ToLower(strings.Trim(words[i-1], ",:"))
			if prev == "therefore" || prev == "so" || prev == "thus" || prev == "hence" {
				continue
			}
			if unicode.IsUpper(firstRune(w)) {
				names[strings.ToLower(w)] = true
				names[strings.TrimSuffix(strings.TrimSuffix(strings.ToLower(w), "'s"), "’s")] = true
			}
		}
	}
	return &names
}

var singularClosedWords = wordSet("half enough plenty little fewer less more much least none nobody nothing nowhere everybody everything everywhere somebody something somewhere anybody anything anywhere he she it they we you i me him her us them himself herself itself themselves ourselves yourselves whoever whatever whichever")

// singularPartitiveWords (v44.0) are quantificational partitive nouns: "the
// majority", "the rest", "the bulk" pick out a part of some things by quantity,
// not one thing, so a description containing one is refused (pb/chief
// 2026-11-18: partitives are quantifiers). A closed class; "the parking lot" is
// refused too.
var singularPartitiveWords = wordSet("majority minority rest remainder bulk most half third quarter fraction part portion share percentage percent proportion number amount quantity lot lots bunch plenty couple handful dozen score plurality pair few several many all some none any each both either neither whole balance residue")

func singularClosedWord(w string) bool {
	if singularPartitiveWords[w] || termDeterminers[w] || nonReferringTermWords[w] || indefiniteHeads[w] || quantityNumerals[w] || singularClosedWords[w] {
		return true
	}
	for _, p := range []string{"every", "some", "any", "no"} {
		if strings.HasPrefix(w, p) {
			return true
		}
	}
	for _, r := range w {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func singularPluralHead(w string) bool {
	return nonSingularStem.MatchString(w) || round2Singular(w) != w
}

func (policy strictRules) singularName(words []string) bool {
	if len(words) < 1 || len(words) > 4 {
		return false
	}
	for _, w := range words {
		if singularClosedWord(w) || indefiniteValueStoppers[w] || framePrepositions[w] || strings.ContainsAny(w, ",;:") || !policy.hasSlotName(w) {
			return false
		}
	}
	return true
}

func (policy strictRules) singularValue(value string) bool {
	value = strings.Trim(strings.TrimSpace(strings.ToLower(value)), ".;:!?\"()")
	if strings.Contains(value, ",") {
		return false
	}
	words := strings.Fields(value)
	if len(words) == 0 {
		return false
	}
	if policy.singularName(words) {
		return true
	}
	// One of-complement only. Bare plural complements are limited to the
	// closed collective heads group/set; the description still denotes one thing.
	if words[0] == "the" {
		for k := 2; k < len(words)-1; k++ {
			if words[k] != "of" {
				continue
			}
			right := words[k+1:]
			for _, w := range right {
				if w == "of" {
					return false
				}
			}
			if !policy.singularValue(strings.Join(words[:k], " ")) {
				return false
			}
			if policy.singularValue(strings.Join(right, " ")) {
				return true
			}
			return policy.descriptionWords(right, words[k-1] == "group" || words[k-1] == "set")
		}
	}

	start := 1
	if words[0] != "the" {
		start = 0
		for i, w := range words {
			if strings.HasSuffix(w, "'s") || strings.HasSuffix(w, "’s") {
				owner := append([]string(nil), words[:i+1]...)
				owner[i] = strings.TrimSuffix(strings.TrimSuffix(w, "'s"), "’s")
				if !policy.singularName(owner) && !(len(owner) == 1 && !singularClosedWord(owner[0]) && policy.hasSlotName("possessive:"+owner[0])) {
					return false
				}
				start = i + 1
				break
			}
		}
		if start == 0 {
			return false
		}
	}
	if start >= len(words) {
		return false
	}
	return policy.descriptionWords(words[start:], false)
}

// The template, not the length of a proposed match, decides whole-clause
// slots. In a copular complement a known bare predicate retains the old behavior;
// noun-phrase shapes (including identity complements) must pass the allowlist.
func (policy strictRules) singularBindingSafe(template []rune, ti int, concrete []rune, ci int, value string) bool {
	variable, _ := schematicVariableForMarker(template[ti])
	if policy.slot_one_type && policy.termSlots[variable] && clauseSlotBefore(template, ti) {
		return policy.singularValue(value)
	}
	// During unification the older checks actually run. Preserve their
	// attribution in one-rule-off retries, but never invent such an exemption
	// when merely generating a consequent.
	if legacySingularRefusal(concrete, ci, value) {
		return true
	}
	return policy.singularBindingShapeSafe(template, ti, concrete, ci, value)
}

func (policy strictRules) singularBindingShapeSafe(template []rune, ti int, concrete []rune, ci int, value string) bool {
	if !policy.slot_value_singular {
		return true
	}
	before := strings.TrimSpace(string(template[:ti]))
	after := strings.TrimSpace(string(template[ti+1:]))
	if before == "" && after == "" {
		return true
	}
	variable, _ := schematicVariableForMarker(template[ti])
	if policy.slot_one_type && policy.termSlots[variable] {
		return policy.singularValue(value)
	}
	if clauseSlotBefore(template, ti) && repeatedSlotClause(value) {
		return true
	}
	words := strings.Fields(strings.ToLower(value))
	if len(words) == 0 {
		return false
	}
	// After a copula the variable stands for a predicate ("If Rex is X, then
	// Fido is X"), so any predicate phrase may be substituted, as in v43.0
	// (afterCopula; existential "there is" is not a copula). A variable that
	// also fills a term slot was already held to the singular allowlist above.
	if afterCopula(strings.ToLower(before)) {
		return true
	}
	return policy.singularValue(value)
}

func (policy strictRules) hasSlotName(w string) bool {
	return policy.slotNames != nil && (*policy.slotNames)[w]
}

// Diagnostic provenance: a variable already refused by an older check must
// stay attributable to that check when it is repeated in the consequent.
const singularLegacyOverride rune = '\uFFFD'

func legacySingularRefusal(concrete []rune, ci int, value string) bool {
	legacy := allStrictRules()
	return !legacy.indefiniteBindingPositionSafe(concrete, ci, value) || !legacy.schematicBindingAllowedAs(value, clauseSlotBefore(concrete, ci)) || (ci == 0 && legacy.sentenceFrameOpener(string(concrete), nil) != "")
}

// The gate has no general noun/adjective parser. Preserve a closed set of
// unambiguous predicate surfaces; unknown copular values fall back to the
// singular-term allowlist rather than being presumed adjectives.
var singularPredicateHeads = wordSet("happy red blue green warm cold wise rich poor friendly mortal alive dead hungry tired safe useful valuable eligible registered ready late early wet dry small large tall short round square honest kind brave")
var singularPredicateModifiers = wordSet("very quite extremely deeply highly")

func singularPredicateValue(words []string) bool {
	if len(words) == 1 {
		return singularPredicateHeads[words[0]]
	}
	return len(words) == 2 && singularPredicateModifiers[words[0]] && singularPredicateHeads[words[1]]
}

// Unknown modifiers are refused. APS is a closed attributive acronym, not
// permission to insert arbitrary proper names into descriptions.
var descriptionModifiers = wordSet("first second third fourth fifth sixth seventh eighth ninth tenth ordinary ultraintelligent aps big red small old young present former parking")

func (policy strictRules) descriptionWords(words []string, plural bool) bool {
	if len(words) == 0 {
		return false
	}
	for i, w := range words {
		ordinal := descriptionModifiers[w] && (w == "first" || w == "second" || w == "third" || w == "fourth" || w == "fifth" || w == "sixth" || w == "seventh" || w == "eighth" || w == "ninth" || w == "tenth") && i < len(words)-1
		if (!ordinal && singularClosedWord(w)) || indefiniteValueStoppers[w] || framePrepositions[w] || strings.ContainsAny(w, "'’.,;:") {
			return false
		}
		if policy.slot_description_shape {
			if policy.hasSlotName(w) && !(w == "aps" && i < len(words)-1) {
				return false
			}
			if i < len(words)-1 && !descriptionModifiers[w] && !coordinatedAdjective(w) {
				return false
			}
			if i == len(words)-1 && (framePredicates[w] || frameTransitives[w] || frameIntransitives[w] || scopeNameVerbs[w]) {
				return false
			}
		}
	}
	return plural || !singularPluralHead(words[len(words)-1])
}
