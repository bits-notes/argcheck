package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
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
		if strings.HasPrefix(lower+" ", phrase+" ") || strings.HasPrefix(lower, phrase+",") {
			return frameOpenerMessage(s[:len(phrase)])
		}
	}
	for _, phrase := range frameComplexPrepositions {
		if strings.HasPrefix(lower+" ", phrase+" ") {
			return frameOpenerMessage(frameSpan(s))
		}
	}
	if framePrepositions[first] {
		return frameOpenerMessage(frameSpan(s))
	}
	if frameSentenceAdverbs[first] {
		return frameOpenerMessage(words[0])
	}
	if strings.HasSuffix(first, "ly") && !frameLyNouns[first] {
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
var frameLyNouns = wordSet("family italy reply assembly supply ally rally bully belly jelly lily july only early daily holy")

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

func sentenceCoordinationUnsafe(s string) bool {
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
		return sentenceCoordinationUnsafe(left) || sentenceCoordinationUnsafe(right)
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
		return !framePlainClause(left) || sentenceCoordinationUnsafe(s[b.end:])
	}
	return false
}

func punctuationDifference(a, b string) string {
	a, b = comparisonClause(a), comparisonClause(b)
	if a != b && strings.ReplaceAll(a, ",", "") == strings.ReplaceAll(b, ",", "") {
		return fmt.Sprintf("differs only in punctuation: '%s' vs '%s'", a, b)
	}
	return ""
}

// Whole-sentence reuse is never an inspection of that sentence. Apply the
// guards at the final gate boundary so alternate FOL routes cannot bypass it.
func sentenceFrameFailure(c NumberedArgumentStep, steps map[int]NumberedArgumentStep, form string) string {
	for _, id := range c.Citations {
		if equalClauses(steps[id].Statement, c.Statement) {
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
		if why := sentenceFrameOpener(s, steps); why != "" {
			return why
		}
		if !strings.Contains(form, "n-ary") && (strings.Contains(form, "conjunction") || strings.Contains(form, "disjunction") || strings.Contains(form, "disjunctive syllogism")) && sentenceCoordinationUnsafe(s) {
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
func punctuationSuffixDifference(a, b string) string {
	aw, bw := strings.Fields(comparisonClause(a)), strings.Fields(comparisonClause(b))
	n := 0
	for n < len(aw) && n < len(bw) && strings.Trim(aw[len(aw)-1-n], ",") == strings.Trim(bw[len(bw)-1-n], ",") {
		n++
	}
	if n < 3 {
		return ""
	}
	return punctuationDifference(strings.Join(aw[len(aw)-n:], " "), strings.Join(bw[len(bw)-n:], " "))
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
