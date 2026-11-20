package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/bits-notes/english-arg-validator/validation/lexicon"
)

// Safe lexical entry points, shared by sentence-internal certificates. Bare
// noun/gerund subjects are admitted by a visible predicate, never merely by
// sentence-initial capitalization.
var frameOpeners = wordSet("the a an every some no all each any most many few several both either neither this that these those my your his her its our their mine yours hers ours theirs i you he she it we they me him us them who what someone somebody something anyone anybody anything everyone everybody everything nobody nothing if when whenever unless therefore so thus hence necessarily possibly")
var framePredicates = wordSet("is are was were am has have had does do did can could will would shall should must may might consists consist exists exist breathe eat drink resemble wrote saw met sent gave won lost took made knew struck")
var frameIntransitives = wordSet("ran run runs jumped jump jumps crashed crash crashes failed fail fails slept sleep sleeps walked walk walks sang sing sings swam swim swims arrived arrive arrives died die dies fell fall falls rose rise rises shone shine shines opened open opens closed close closes rang ring rings rained rain rains snowed snow snows worked work works succeeded succeed succeeds smiled smile smiles flew fly flies barked bark barks purred purr purrs stopped stop stops started start starts dances dance danced reads read blows blow blew melts melt melted thrives thrive thrived")

// Extensional transitive frames have nominal objects, never that/if/whether
// clauses. Unknown verbs do not gain a clause boundary by omission.
var frameTransitives = wordSet("admires admire admired respects respect respected loves love loved likes like liked owns own owned contains contain contained helps help helped caused cause causes leaves leave left outweighs outweigh outweighed solved solve solves wrote write writes gave give gives sent send sends")

// These nominal heads support adjective+noun bare subjects without treating
// any arbitrary two-word prefix as a noun phrase.
var frameNominalHeads = wordSet("value consciousness knowledge evidence information intelligence agency morality truth reality experience behavior behaviour welfare utility probability risk safety alignment learning reasoning theory theories equations dynamics")
var frameWord = regexp.MustCompile(`[A-Za-z][A-Za-z0-9'’-]*`)

func sentenceFrameOpener(s string, steps map[int]NumberedArgumentStep) string {
	return allStrictRules().sentenceFrameOpener(s, steps)
}

func (policy strictRules) sentenceFrameOpener(s string, steps map[int]NumberedArgumentStep) string {
	s = strings.TrimSpace(expandGateContractions(s))
	if s == "" || strings.HasPrefix(s, "(") || folExplicit(s) || onlyGerund(strings.ToLower(s)) {
		return ""
	}
	words := frameWord.FindAllString(s, -1)
	if len(words) == 0 {
		return ""
	}
	first := strings.ToLower(words[0])
	lower := strings.ToLower(s)
	// Explicit structure and determiners/logical openers are frames.
	if frameOpeners[first] || len(words[0]) == 1 || groupingLabel.MatchString(s) {
		return ""
	}
	for _, prefix := range []string{"more than ", "less than ", "at least ", "at most ", "exactly ", "almost every ", "almost all ", "nearly every ", "nearly all "} {
		if strings.HasPrefix(lower, prefix) {
			return ""
		}
	}
	if len(words) > 1 && strings.HasPrefix(s[len(words[0]):], ", ") && (strings.EqualFold(words[1], "who") || strings.EqualFold(words[1], "which")) {
		return ""
	}
	// Closed classes that may cover the whole sentence, checked before any
	// name or gerund evidence ("According" ends in -ing but is a preposition).
	// Strict: a fronted frame is refused with or without a comma.
	for _, phrase := range frameAdverbialPhrases {
		if policy.opener_frame && (strings.HasPrefix(lower+" ", phrase+" ") || strings.HasPrefix(lower, phrase+",")) && policy.hit("opener.frame") {
			return frameOpenerMessage(s[:len(phrase)])
		}
	}
	for _, phrase := range frameComplexPrepositions {
		if policy.opener_frame && strings.HasPrefix(lower+" ", phrase+" ") && policy.hit("opener.frame") {
			return frameOpenerMessage(frameSpan(s))
		}
	}
	if policy.opener_frame && framePrepositions[first] && !frameGerundSubject(strings.Fields(s)) && policy.hit("opener.frame") {
		return frameOpenerMessage(frameSpan(s))
	}
	if policy.opener_adverb && frameSentenceAdverbs[first] && policy.hit("opener.adverb") {
		return frameOpenerMessage(words[0])
	}
	if policy.opener_adverb && strings.HasSuffix(first, "ly") && !frameLyNouns[first] && (policy.opener_ly_exception || !round2OldLyExceptions[first]) && !frameLyAdjectiveSubject(strings.Fields(s)) && !frameDegreeAdjectiveSubject(strings.Fields(s)) && policy.hit("opener.adverb") {
		if policy.opener_ly_exception && round2OldLyExceptions[first] {
			policy.hit("opener.ly_exception")
		}
		return frameOpenerMessage(words[0])
	}
	// Bare-noun, gerund and adjective+noun subjects ("Beliefs ...", "Trying to
	// remove all assumptions ...", "Human value ...") and names stay subjects.
	return ""
}

// frameSpan is the fronted frame up to its comma, or its first four words.
func frameSpan(s string) string {
	if i := strings.Index(s, ","); i > 0 && i < 80 {
		return strings.TrimSpace(s[:i+1])
	}
	words := strings.Fields(s)
	return strings.Join(words[:min(len(words), 4)], " ")
}

// Prepositions are a closed class. Every single-word preposition, including
// -ing forms (according, regarding, concerning, ...) and subordinators that
// front a frame (after, before, since, until, as, because, although, though,
// while, whereas), refuses a sentence-internal step when it opens the sentence.
var framePrepositions = wordSet("aboard about above across after against along alongside amid amidst among amongst around as astride at atop barring because before behind below beneath beside besides between beyond but by circa concerning considering despite down during except excepting excluding following for from given in including inside into like minus near notwithstanding of off on onto opposite out outside over past pending per plus regarding round save since than through throughout till to toward towards under underneath unlike until unto up upon versus via with within without according although though while whereas whilst once whenever wherever lest")

// Complex prepositions (multi-word), longest first where they share a prefix.
var frameComplexPrepositions = []string{
	"according to", "as per", "as for", "as to", "as regards", "as far as", "in line with", "in view of",
	"in light of", "in accordance with", "in addition to", "in spite of", "in terms of", "in respect of",
	"in relation to", "in response to", "in the case of", "in the context of", "in the light of",
	"on behalf of", "on account of", "on the basis of", "with respect to", "with regard to",
	"with reference to", "by means of", "by virtue of", "by way of", "due to", "owing to", "thanks to",
	"apart from", "aside from", "instead of", "rather than", "prior to", "subsequent to", "contrary to",
	"relative to", "according as", "next to", "close to", "far from", "out of", "because of", "ahead of",
}

// Multi-word adverbials that comment on the whole sentence.
var frameAdverbialPhrases = []string{
	"of course", "no doubt", "without doubt", "at best", "at worst", "at first", "at last", "at the moment",
	"in theory", "in principle", "in practice", "in effect", "in fact", "in general", "in short", "in sum",
	"in other words", "so to speak", "for example", "for instance", "on the whole", "on balance",
	"on the other hand", "by and large", "all things considered", "as a rule", "as it were", "to be sure",
	"to some extent", "more or less", "sort of", "kind of", "if anything",
}

func frameOpenerMessage(opener string) string {
	return fmt.Sprintf("'%s' may cover the whole sentence; move it out, state the claim plainly, or reword the phrase so the sentence starts with its subject (e.g. 'Rationalist notions that are deeply held ...' instead of 'Deeply held rationalist notions ...').", opener)
}

// Non-ly sentence adverbs (closed class as far as English allows).
// "almost"/"nearly" before every/all are quantifier openers handled above.
var frameSentenceAdverbs = wordSet("perhaps maybe still even just only also too indeed surely certainly again already yet sometimes often once always never seldom soon now today tomorrow yesterday tonight however nevertheless nonetheless moreover furthermore meanwhile otherwise instead anyway anyhow regardless overall altogether hence thus therefore so then thereby accordingly consequently admittedly allegedly supposedly reportedly apparently evidently ostensibly arguably presumably possibly probably else almost nearly quite rather somewhat")
var frameLyNouns = wordSet("family italy reply assembly subassembly supply ally rally bully belly jelly lily july holy monopoly anomaly homily melancholy butterfly")

// frameLyAdjectives are -ly adjectives that are not used as sentence
// adverbs, so at the start of a step they open a noun-phrase subject
// ("Friendly action depends on ..."). Adverb-capable -ly words ("likely",
// "only", "early", "daily") are deliberately absent.
var frameLyAdjectives = wordSet("friendly unfriendly ugly lovely lonely costly elderly orderly cowardly scholarly worldly motherly fatherly brotherly sisterly neighborly heavenly ghostly")

// frameLyAdjectiveSubject reports an -ly adjective opening a noun-phrase
// subject: the adjective is followed directly (no comma) by another word.
func frameLyAdjectiveSubject(words []string) bool {
	if len(words) < 2 {
		return false
	}
	first := strings.ToLower(words[0])
	if strings.ContainsAny(first, ",;:") || !frameLyAdjectives[first] {
		return false
	}
	return true
}

const frameCoordinationMessage = "clause coordination after a verb is not certified; write 'Both ... and ...', use parentheses, or split into two premises."

// Only known complete, non-complement-taking predicates justify an unmarked
// split. A copula is accepted with a single predicate phrase, not a second
// finite clause. Explicit grouping supplies the boundary independently.
func framePlainClause(s string) bool {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return true
	}
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		all := true
		for _, part := range parts {
			all = all && framePlainClause(part)
		}
		if all {
			return true
		}
	}
	words := strings.Fields(strings.ToLower(s))
	if len(words) == 1 {
		return true
	} // propositional letters/opaque atoms
	for i, w := range words {
		if i == 0 {
			continue
		}
		if frameIntransitives[w] {
			return i == len(words)-1 && frameSubject(strings.Fields(s)[:i])
		}
		if frameTransitives[w] {
			if !frameSubject(strings.Fields(s)[:i]) || i+1 >= len(words) {
				return false
			}
			object := words[i+1:]
			if len(object) == 1 && len(object[0]) == 1 {
				return false
			}
			for _, token := range object {
				if token == "that" || token == "whether" || token == "if" || framePredicates[token] || frameIntransitives[token] {
					return false
				}
			}
			return true
		}
		if w == "is" || w == "are" || w == "was" || w == "were" || w == "am" {
			// A subject with an embedded finite predicate is not evidence of an
			// outer clause. The subject grammar permits nominal of-phrases.
			if !frameSubject(strings.Fields(s)[:i]) {
				return false
			}
			for _, p := range words[i+1:] {
				if framePredicates[p] {
					return false
				}
			}
			return i+1 < len(words)
		}
	}
	return false
}

func (policy strictRules) sentenceCoordinationUnsafe(s string) (unsafe bool) {
	if !policy.coordination_clause {
		return false
	}
	defer func() { _ = unsafe && policy.hit("coordination.clause") }()
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	for _, prefix := range []string{"both ", "either ", "if ", "when ", "whenever ", "unless ", "it is the case both that "} {
		if strings.HasPrefix(lower, prefix) {
			return false
		}
	}
	if strings.Contains(lower, " if and only if ") {
		return false
	}
	if op, left, right, managed, why := groupingPlan(s); managed && why == "" && (op == "and" || op == "or") {
		return policy.sentenceCoordinationUnsafe(left) || policy.sentenceCoordinationUnsafe(right)
	}
	bs, _ := groupingBoundaries(s)
	for i, b := range bs {
		if !clauseOperandStart(s[b.end:]) && i+1 < len(bs) {
			continue
		}
		if b.op != "and" && b.op != "or" {
			continue
		}
		left := strings.TrimSpace(strings.TrimSuffix(s[:b.start], ","))
		return !framePlainClause(left) || policy.sentenceCoordinationUnsafe(s[b.end:])
	}
	return false
}

func (policy strictRules) punctuationDifference(a, b string) string {
	a, b = policy.comparisonClause(a), policy.comparisonClause(b)
	if a != b && strings.ReplaceAll(a, ",", "") == strings.ReplaceAll(b, ",", "") {
		return fmt.Sprintf("differs only in punctuation: '%s' vs '%s'", a, b)
	}
	return ""
}

// Apply the guards at the final gate boundary so alternate FOL routes cannot
// bypass them. Preserve legacy diagnostics before checking explicit frames.
func (policy strictRules) sentenceFrameFailure(c NumberedArgumentStep, steps map[int]NumberedArgumentStep, form string) string {
	if why := policy.legacySentenceFrameFailure(c, steps, form); why != "" {
		return why
	}
	if policy.scope_mixed_andor {
		texts := []string{c.Statement}
		for _, id := range c.Citations {
			texts = append(texts, fitchAssumptionStatement(steps[id]))
		}
		for _, text := range texts {
			if hint := mixedFrameHint(text); hint != "" && policy.hit("scope.mixed_andor") {
				return hint
			}
		}
	}
	return ""
}

func (policy strictRules) legacySentenceFrameFailure(c NumberedArgumentStep, steps map[int]NumberedArgumentStep, form string) string {
	if why := policy.round2ScopeFailure(c, steps); why != "" {
		return why
	}
	for _, id := range c.Citations {
		if policy.equalClauses(steps[id].Statement, c.Statement) {
			return ""
		}
	}
	internal := false
	for _, key := range []string{"conjunct", "disjunct", "instantiation", "generalization", "identity substitution", "first-order", "notation equivalence", "schematic"} {
		internal = internal || strings.Contains(form, key)
	}
	if !internal {
		return ""
	}
	texts := []string{c.Statement}
	for _, id := range c.Citations {
		texts = append(texts, fitchAssumptionStatement(steps[id]))
	}
	for i, s := range texts {
		reused := false
		for j, other := range texts {
			if i != j && len(other) > len(s) && strings.Contains(" "+strings.ToLower(other)+" ", strings.ToLower(s)) {
				reused = true
			}
		}
		if reused {
			continue
		}
		if why := policy.sentenceFrameOpener(s, steps); why != "" {
			return why
		}
		if !strings.Contains(form, "n-ary") && (strings.Contains(form, "conjunction") || strings.Contains(form, "disjunction") || strings.Contains(form, "disjunctive syllogism")) && policy.sentenceCoordinationUnsafe(s) {
			return frameCoordinationMessage
		}
	}
	return ""
}

// Nominal subjects may contain determiners inside prepositional phrases.
func frameSubject(words []string) bool {
	if len(words) == 0 {
		return false
	}
	if len(words) == 1 {
		return true
	}
	if strings.HasSuffix(strings.ToLower(words[0]), "'s") || strings.HasSuffix(strings.ToLower(words[0]), "’s") {
		return true
	}
	for i, w := range words {
		if (w == "and" || w == "or") && i > 0 && i+1 < len(words) {
			return frameSubject(words[:i]) && frameSubject(words[i+1:])
		}
	}
	if !frameOpeners[strings.ToLower(words[0])] {
		names := true
		for _, w := range words {
			names = names && (unicode.IsUpper(firstRune(w)) || w == "and" || w == "or")
		}
		nominalHead := frameNominalHeads[strings.ToLower(words[len(words)-1])] || strings.Contains(words[len(words)-1], "-") || strings.HasSuffix(words[len(words)-1], "ing")
		for _, modifier := range words[:len(words)-1] {
			nominalHead = nominalHead && !framePredicates[strings.ToLower(modifier)] && !frameTransitives[strings.ToLower(modifier)] && !strings.HasSuffix(modifier, "ed") && !strings.HasSuffix(modifier, "s")
		}
		return names || strings.HasSuffix(strings.ToLower(words[0]), "ing") || nominalHead
	}
	for i, w := range words[1:] {
		if determinerWords[strings.ToLower(w)] {
			switch strings.ToLower(words[i]) {
			case "of", "in", "with", "from", "on", "for", "by", "at", "to":
			default:
				return false
			}
		}
	}
	return true
}

// Diagnostic only: compare the common predicate suffix after different
// subjects (for example, a universal class and its named instance).
func (policy strictRules) punctuationSuffixDifference(a, b string) string {
	aw, bw := strings.Fields(policy.comparisonClause(a)), strings.Fields(policy.comparisonClause(b))
	n := 0
	for n < len(aw) && n < len(bw) && strings.Trim(aw[len(aw)-1-n], ",") == strings.Trim(bw[len(bw)-1-n], ",") {
		n++
	}
	if n < 3 {
		return ""
	}
	return policy.punctuationDifference(strings.Join(aw[len(aw)-n:], " "), strings.Join(bw[len(bw)-n:], " "))
}

var modifiedQuantifierWords = wordSet("no none nobody nothing one two three four five six seven eight nine ten few all every each most many half any some several dozens hundreds thousands millions")
var quantifierModifierWords = wordSet("next close well just almost nearly about around over under only exactly")

// modifiedQuantifierPhrase reports a quantifier phrase introduced by
// modifiers: any -ly adverb ("practically", "virtually", "precisely",
// "roughly"), or a listed modifier ("next to", "close to", "well over").
// "Practically no person", "next to no one" and "precisely one senator" are
// quantifiers, not terms, however the modifier is spelled. A closed class of
// quantifier words after an open class of modifiers keeps the slot strict
// without listing every adverb.
func modifiedQuantifierPhrase(lower string) bool {
	return allStrictRules().modifiedQuantifierPhrase(lower)
}

func (policy strictRules) modifiedQuantifierPhrase(lower string) (modified bool) {
	if !policy.slot_modified_quantifier {
		return false
	}
	defer func() { _ = modified && policy.hit("slot.modified_quantifier") }()
	words := strings.Fields(strings.ToLower(lower))
	i := 0
	for i < len(words) {
		w := strings.Trim(words[i], ",;:")
		if strings.HasSuffix(w, "ly") || quantifierModifierWords[w] || w == "to" || w == "than" {
			i++
			continue
		}
		break
	}
	if i == 0 || i >= len(words) {
		return false
	}
	return modifiedQuantifierWords[strings.Trim(words[i], ",;:")]
}

var quantityNumerals = wordSet("zero one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen seventeen eighteen nineteen twenty thirty forty fifty sixty seventy eighty ninety hundred thousand million billion dozen dozens hundreds thousands millions billions umpteen")
var quantityNouns = wordSet("minority majority fraction proportion percentage share handful couple number lot lots loads tons plenty bunch group dozen score third quarter half tenth fifth sixth eighth part portion bulk rest remainder")

// quantityPhrase reports a term that starts with a quantity expression: a
// numeral or digit ("zero robots", "two robots"), or a quantity noun
// ("a minority of robots", "half of the robots", "a handful of robots").
// Such a phrase is not an individual. The numerals and quantity nouns are a
// closed class.
func (policy strictRules) quantityPhrase(lower string) (quantity bool) {
	if !policy.slot_quantity {
		return false
	}
	defer func() { _ = quantity && policy.hit("slot.quantity") }()
	words := strings.Fields(strings.ToLower(lower))
	if len(words) == 0 {
		return false
	}
	first := strings.Trim(words[0], ",;:")
	if quantityNumerals[first] || (first != "" && first[0] >= '0' && first[0] <= '9') {
		// A numeral counts things only before a plural noun ("zero robots",
		// "two robots"); "zero probability" and "one mind" name a value or
		// an individual.
		if len(words) < 2 {
			return false
		}
		next := strings.Trim(words[1], ",;:")
		return strings.HasSuffix(next, "s") && !strings.HasSuffix(next, "ss")
	}
	if (first == "a" || first == "an" || first == "the") && len(words) > 2 {
		head := strings.Trim(words[1], ",;:")
		if strings.Trim(words[2], ",;:") != "of" || !quantityNouns[head] {
			return false
		}
		// "the proportion of X", "the number of X" name a value; with "the",
		// only the partitive heads quantify ("the majority of robots").
		if first == "the" {
			switch head {
			case "majority", "minority", "bulk", "rest", "remainder", "half":
				return true
			}
			return false
		}
		return true
	}
	return quantityNouns[first] && len(words) > 1 && strings.Trim(words[1], ",;:") == "of"
}

// upwardQuantity is the part of quantityPhrase that is safe in a schematic
// slot ("a majority of robots are red" and "if X is red, X is warm" give "a
// majority of robots are warm"). Numerals ("two" may mean exactly two), zero,
// minorities, fractions and halves are not.
func upwardQuantity(lower string) bool {
	words := strings.Fields(strings.ToLower(lower))
	if len(words) < 3 {
		return false
	}
	i := 0
	if words[0] == "a" || words[0] == "an" || words[0] == "the" {
		i = 1
	}
	switch words[i] {
	case "majority", "lot", "lots", "loads", "tons", "plenty", "bulk":
		return true
	}
	return false
}

// frameIngPrepositions are -ing words that are both prepositions ("Considering
// the cost, we stop") and gerunds ("Considering brain emulation invites
// caution").
var frameIngPrepositions = wordSet("considering regarding concerning following including excluding barring pending excepting")

var frameGerundFiniteVerbs = wordSet("is was has does can will would should may might must remains seems requires invites")

var frameGerundBlockers = wordSet("we they he she it i you one there this that these those")

// frameGerundSubject reports an -ing preposition used as a gerund subject:
// no comma before the main verb, no other subject in between, and a
// singular finite verb within six words. Anything else stays an opener.
func frameGerundSubject(words []string) bool {
	if len(words) < 3 || strings.ContainsAny(words[0], ",;:") || !frameIngPrepositions[strings.ToLower(words[0])] {
		return false
	}
	for index := 1; index < len(words) && index <= 6; index++ {
		raw := words[index]
		if strings.ContainsAny(words[index-1], ",;:") {
			return false
		}
		word := strings.ToLower(strings.Trim(raw, ",;:."))
		if frameGerundBlockers[word] {
			return false
		}
		if index >= 2 && (frameGerundFiniteVerbs[word] || (strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") && frameGerundVerbLike(words, index))) {
			return true
		}
	}
	return false
}

// frameGerundVerbLike accepts an -s word as the finite verb only when the
// word before it is not a determiner or preposition (so "the costs" stays a noun).
func frameGerundVerbLike(words []string, index int) bool {
	previous := strings.ToLower(strings.Trim(words[index-1], ",;:."))
	switch previous {
	case "the", "a", "an", "its", "their", "his", "her", "our", "my", "your", "of", "for", "to", "in", "on", "with", "by", "and", "or":
		return false
	}
	return true
}

// Degree modifiers only: sentence adverbs and hedges are deliberately absent.
var frameDegreeAdverbs = wordSet("purely entirely wholly fully highly completely totally extremely")

func frameDegreeAdjectiveSubject(words []string) bool {
	return len(words) >= 2 && frameDegreeAdverbs[strings.ToLower(words[0])] && lexicon.IsAdjective(words[1])
}

// Scope admission additionally requires a noun head and a visible predicate.
func frameDegreeAdjectiveClause(words []string) bool {
	if !frameDegreeAdjectiveSubject(words) {
		return false
	}
	for i := 2; i+1 < len(words); i++ {
		w := strings.ToLower(words[i])
		next := strings.ToLower(words[i+1])
		if (lexicon.IsSingularNoun(w) || lexicon.IsPluralMassNoun(w)) && (framePredicates[next] || frameIntransitives[next] || frameTransitives[next]) {
			return true
		}
		if !lexicon.IsAdjective(w) {
			return false
		}
	}
	return false
}
