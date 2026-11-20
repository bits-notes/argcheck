package validation

import (
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/bits-notes/english-arg-validator/validation/lexicon"
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

// Ordinary capitals need either lexical evidence or a capitalized mid-sentence
// term occurrence with no lowercase occurrence anywhere in the argument.
// Quantities/closed words never supply either route. Internal capitals retain
// their separate historical evidence rule, including at sentence start.
func schematicNameEvidence(steps map[int]NumberedArgumentStep) *map[string]bool {
	names := map[string]bool{}
	lowercase := map[string]bool{}
	baseWord := func(raw string) string {
		w := strings.Trim(raw, ".,;:!?\"()[]")
		return strings.TrimSuffix(strings.TrimSuffix(w, "'s"), "’s")
	}
	for _, step := range steps {
		for _, raw := range strings.Fields(step.Statement) {
			base := baseWord(raw)
			if base == strings.ToLower(base) {
				lowercase[base] = true
			}
		}
	}
	for _, step := range steps {
		words := strings.Fields(step.Statement)
		for i, raw := range words {
			w := strings.Trim(raw, ".,;:!?\"()[]")
			base := baseWord(raw)
			lower := strings.ToLower(base)
			if base != w && unicode.IsUpper(firstRune(base)) && !scopeContractions[lower] {
				names["possessive:"+lower] = true
			}
			if internalCapital(w) {
				names[strings.ToLower(w)] = true
				names[lower] = true
				continue
			}
			// Plain capitalisation away from sentence start (the pre-v46 test)
			// still marks a word that may be a name. It never makes a value
			// acceptable; descriptions use it to keep refusing name modifiers
			// ("the Bob dog") that strict evidence no longer records.
			if unicode.IsUpper(firstRune(base)) && i > 0 && !strings.ContainsAny(words[i-1], ".!?") {
				names["cap:"+lower] = true
			}
			if !unicode.IsUpper(firstRune(base)) || singularClosedWord(lower) || lexicon.IsNameForbidden(lower) {
				continue
			}
			evidence := lexicon.IsName(lower)
			if !evidence && !lowercase[lower] && i > 0 && !strings.ContainsAny(words[i-1], ".!?") {
				prev := strings.ToLower(strings.Trim(words[i-1], ",:"))
				if prev != "therefore" && prev != "so" && prev != "thus" && prev != "hence" {
					// An allowlisted count plural ("Robots") stays ambiguous even
					// when capitalised in a term position.
					evidence = nameTermPosition(words, i) && !(lexicon.IsPluralMassNoun(lower) && !lexicon.IsMassNoun(lower))
				}
			}
			if evidence {
				names[strings.ToLower(w)] = true
				names[lower] = true
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
	return policy.singularOnlyValue(value) || (policy.slot_plural_mass_value && policy.pluralMassValue(value))
}

// Keep identity and description recursion restricted to singular terms.
func (policy strictRules) singularOnlyValue(value string) bool {
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
			if !policy.singularOnlyValue(strings.Join(words[:k], " ")) {
				return false
			}
			if policy.singularOnlyValue(strings.Join(right, " ")) {
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
	if policy.slot_one_type && policy.termSlots[variable] && clauseSlotBefore(template, ti) && policy.hit("slot.one_type") {
		return policy.termValueAt(template, ti, value)
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
	if !policy.slot_value_singular || !policy.hit("slot.value_singular") {
		return true
	}
	before := strings.TrimSpace(string(template[:ti]))
	after := strings.TrimSpace(string(template[ti+1:]))
	if before == "" && after == "" {
		return true
	}
	variable, _ := schematicVariableForMarker(template[ti])
	if policy.slot_one_type && policy.termSlots[variable] && policy.hit("slot.one_type") {
		return policy.termValueAt(template, ti, value)
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
	return policy.termValueAt(template, ti, value)
}

func (policy strictRules) hasSlotName(w string) bool {
	return policy.slotNames != nil && (*policy.slotNames)[w]
}

// Diagnostic provenance: a variable already refused by an older check must
// stay attributable to that check when it is repeated in the consequent.
const singularLegacyOverride rune = '\uFFFD'

func legacySingularRefusal(concrete []rune, ci int, value string) bool {
	legacy := allStrictRules()
	return !legacy.indefiniteBindingPositionSafe(concrete, ci, value) || !legacyBindingAllowedAs(value, clauseSlotBefore(concrete, ci)) || (ci == 0 && legacy.sentenceFrameOpener(string(concrete), nil) != "")
}

// legacyBindingAllowedAs is schematicBindingAllowedAs under the full default
// rules with no recorder, which depends only on its arguments. Unification asks
// it for every candidate segment of every binding, so results are cached; past
// the limit, values are checked without caching.
var (
	legacyBindingCache [2]sync.Map
	legacyBindingCount atomic.Int64
)

const legacyBindingCacheLimit = 50000

func legacyBindingAllowedAs(value string, clauseSlot bool) bool {
	slot := 0
	if clauseSlot {
		slot = 1
	}
	if cached, ok := legacyBindingCache[slot].Load(value); ok {
		return cached.(bool)
	}
	allowed := allStrictRules().schematicBindingAllowedAs(value, clauseSlot)
	if legacyBindingCount.Load() < legacyBindingCacheLimit {
		if _, loaded := legacyBindingCache[slot].LoadOrStore(value, allowed); !loaded {
			legacyBindingCount.Add(1)
		}
	}
	return allowed
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
			if (policy.hasSlotName(w) || policy.hasSlotName("cap:"+w)) && !(w == "aps" && i < len(words)-1) {
				policy.hit("slot.description_shape")
				return false
			}
			// v44.0 accepted one unrestricted modifier ("the hard-takeoff concept");
			// Preserve that modifier behavior; the allowlist and the verb-head
			// check apply only to descriptions with two or more modifiers.
			if len(words) > 2 && i < len(words)-1 && !ordinal && !lexicon.IsAdjective(w) {
				policy.hit("slot.description_shape")
				return false
			}
			if len(words) > 2 && i == len(words)-1 && (framePredicates[w] || frameTransitives[w] || frameIntransitives[w] || scopeNameVerbs[w]) {
				policy.hit("slot.description_shape")
				return false
			}
		}
	}
	head := words[len(words)-1]
	if plural {
		return lexicon.IsPluralMassNoun(head)
	}
	return lexicon.IsSingularNoun(head) && !singularPluralHead(head)
}
