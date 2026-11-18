package validation

import (
	"regexp"
	"strings"
	"unicode"
)

// These lexical fallbacks apply only to term extraction, not atomic premises.
// Relative/complement clauses on a nominal head remain part of that term.
var bindingAttitude = regexp.MustCompile(`\b(?:believ(?:e|es|ed|ing)|think(?:s|ing)?|thought|imagin(?:e|es|ed|ing)|dream(?:s|ed|ing)?|pretend(?:s|ed|ing)?|hop(?:e|es|ed|ing)|fear(?:s|ed|ing)?|say(?:s|ing)?|said|claim(?:s|ed|ing)?|know(?:s|ing)?|knew|doubt(?:s|ed|ing)?|wonder(?:s|ed|ing)?|suspect(?:s|ed|ing)?|suppos(?:e|es|ed|ing)|assum(?:e|es|ed|ing)|expect(?:s|ed|ing)?|wish(?:es|ed|ing)?|report(?:s|ed|ing)?|den(?:y|ies|ied)|assert(?:s|ed)?|argu(?:e|es|ed))\b`)
var sentenceOperator = regexp.MustCompile(`^(?:possibly|perhaps|maybe|allegedly|supposedly|reportedly|probably|apparently|usually|sometimes|yesterday|once|conceivably|presumably)\b`)
var impersonalEmbedding = regexp.MustCompile(`^it (?:is|was|seems|seemed)\b.*\bthat\b`)
var nominalAttitudeHead = regexp.MustCompile(`^(?:the|a|an) (?:(?:false|true|alleged|reported|original|same|different) )*(?:claim|claims|report|reports|thought|thoughts|fear|fears|hope|hopes|dream|dreams|wish|wishes|suspect|suspects)$`)
var nominalClause = regexp.MustCompile(`\s+(?:that|who|which|whom|whose)\s+`)

func embeddedBinding(s string) bool {
	return embeddedTerm(s, false)
}

// embeddedClauseValue is embeddedBinding for a variable in a clause slot
// ("... even when U"): the value is meant to be a whole clause, so the
// term-only check that refuses a verb plus complement inside a term
// (clauseInsideTerm) does not apply. Every other embedding check still does.
func embeddedClauseValue(s string) bool {
	return embeddedTermChecks(s, false, false)
}

// embeddedClauseHead is the stricter test used for coordination under a
// leading operator: only an attitude verb with a that/whether/if complement
// counts ("we should expect X to possess Y" is not an embedding).
func embeddedClauseHead(s string) bool {
	return embeddedTerm(s, true)
}

func embeddedTerm(s string, needComplementizer bool) bool {
	return embeddedTermChecks(s, needComplementizer, true)
}

func embeddedTermChecks(s string, needComplementizer, termSlot bool) bool {
	if termSlot && !needComplementizer && clauseInsideTerm(s) {
		return true
	}
	s = strings.ToLower(strings.TrimSpace(s))
	if sentenceOperator.MatchString(s) || impersonalEmbedding.MatchString(s) {
		return true
	}
	// Check the head before allowing a relative clause: "the observer believes
	// that ..." is not a nominal head. Negation inside a relative is harmless.
	head := s
	if loc := nominalClause.FindStringIndex(s); loc != nil {
		head = s[:loc[0]]
	}
	if len(strings.Fields(head)) == 1 || nominalAttitudeHead.MatchString(head) {
		return false
	}
	words := strings.Fields(s)
	headWords := len(strings.Fields(head))
	for i, w := range words[:headWords] {
		// The embedding verb follows the subject phrase directly. Stop at
		// the first preposition, subordinator or comma: "no reason at all
		// to think he is there" deep inside a description is not the main
		// verb of the term.
		if i > 0 && (attitudeScanStop[w] || strings.HasSuffix(words[i-1], ",")) {
			break
		}
		if !bindingAttitude.MatchString(w) || bindingAttitude.FindString(w) != w {
			continue
		}
		if attitudeUsedAsVerb(words, i, needComplementizer) {
			return true
		}
	}
	return false
}

var attitudeScanStop = map[string]bool{
	"of": true, "in": true, "from": true, "to": true, "after": true, "with": true, "about": true,
	"for": true, "on": true, "at": true, "by": true, "while": true, "because": true, "when": true,
	"but": true, "although": true, "though": true, "since": true, "until": true, "before": true,
	"where": true, "without": true, "under": true, "between": true, "across": true, "into": true,
}

var attitudeNounContext = map[string]bool{
	"the": true, "a": true, "an": true, "of": true, "its": true, "their": true, "his": true,
	"her": true, "our": true, "my": true, "your": true, "these": true, "those": true, "this": true,
	"that": true, "such": true, "any": true, "some": true, "no": true, "many": true, "main": true,
	"common": true, "widespread": true, "basic": true, "central": true, "key": true, "own": true,
}

var attitudeBaseSubject = map[string]bool{
	"i": true, "you": true, "we": true, "they": true, "people": true, "to": true, "will": true,
	"would": true, "may": true, "might": true, "can": true, "could": true, "should": true,
	"must": true, "do": true, "does": true, "did": true, "not": true,
}

// attitudeUsedAsVerb reports whether words[i] is an attitude or report verb
// taking a clause inside a term ("the observer believes alpha", "x claims
// that y"), rather than a noun ("the main claim", "fears of misalignment",
// "reports about y"). Base forms count only after a pronoun, "to" or an
// auxiliary.
func attitudeUsedAsVerb(words []string, i int, needComplementizer bool) bool {
	w := words[i]
	if i == 0 || i == len(words)-1 {
		return false
	}
	prev, next := words[i-1], strings.TrimRight(words[i+1], ",;:")
	complementizer := next == "that" || next == "whether" || next == "if"
	// A past form before a noun is usually an adjective ("imagined future
	// weight-editing"); count it only with an explicit complement clause.
	past := strings.HasSuffix(w, "ed") || w == "thought" || w == "said" || w == "knew"
	if (needComplementizer || past) && !complementizer {
		return false
	}
	if attitudeNounContext[prev] || strings.HasSuffix(prev, "-") || strings.HasSuffix(prev, "'s") || strings.HasSuffix(prev, "’s") || strings.Contains(prev, "-") {
		return false
	}
	switch next {
	case "of", "about", "for", "and", "or", "in", "on", "over", "to", "from", "with", "by", "as", "is", "are", "was", "were", "like", "such", "combined", "made":
		return false
	}
	// Without a complementizer, a clause complement ("believes alpha" +
	// "is red") leaves only the embedded subject inside the term. A verb
	// followed by a long object ("fears selfish ais due to ...") is an
	// ordinary transitive use.
	if !complementizer && len(words)-1-i > 4 {
		return false
	}
	finite := strings.HasSuffix(w, "s") || strings.HasSuffix(w, "ed") || w == "thought" || w == "said" || w == "knew"
	if !finite && !attitudeBaseSubject[prev] {
		return false
	}
	return true
}

// English pluralia tantum and -s nouns whose stem is a different word: a closed
// class, listed in full, so stripping -s never equates them with a singular.
var nonSingularStem = regexp.MustCompile(`(?:^|\s)(?:clothes|species|means|goods|news|odds|thanks|series|savings|earnings|belongings|remains|surroundings|axes|arms|customs|glasses|tongs|scissors|pants|trousers|shorts|jeans|tights|pyjamas|pajamas|knickers|briefs|spectacles|binoculars|goggles|pliers|shears|tweezers|clippers|forceps|scales|bellows|gallows|headquarters|barracks|crossroads|whereabouts|outskirts|proceeds|riches|wages|dregs|leftovers|annals|archives|congratulations|credentials|damages|grounds|lodgings|manners|minutes|morals|quarters|regards|spirits|stairs|troops|valuables|arrears|clippings|ashes|innings|alms|amends|doldrums|entrails|jitters|measles|mumps|rabies|diabetes|physics|mathematics|economics|ethics|politics|linguistics|statistics|athletics|gymnastics|electronics|genetics|logistics|dynamics|mechanics|semantics|optics|acoustics|phonetics|aesthetics|metaphysics)$`)
var identityOpaqueExtra = regexp.MustCompile(`(?i)\b(?:is looking for|looks for|is searching for|seeks|wants|needs|hopes for|fears|believes|imagines|worships|allegedly|supposedly|reportedly|dream(?:s|ed|ing)?|pretend(?:s|ed|ing)?|called|named|known as)\b`)
var questionStem = regexp.MustCompile(`(?i)\b(?:wonder\w*|ask\w*|investigat\w*|check\w*|forg\w*|debat\w*|determin\w*|decid\w*|test\w*|verif\w*|doubt\w*|question\w*|consider\w*|discuss\w*|explor\w*|examin\w*|stud\w*|know\w*|tell\w*|judg\w*|assess\w*|evaluat\w*|figur\w*|see|sees|saw|seen|knew|told|unsure|sure|aware)\b`)
var incompleteIfConsequent = regexp.MustCompile(`(?i)(?:\b(?:about|of|to|for|with|at|in|on|whether|aloud|out|up)|\b(?:(?:is|are|was|were) [a-z]+ing|(?:has|have|had) (?:[a-z]+ed|forgotten)))$`)

const operatorGroupingMessage = `coordination under a leading operator is ambiguous; use parentheses, e.g. "(Perhaps A) and B" or "Perhaps (A and B)", or explicit both...and / either...or`

func operatorCoordinationAmbiguous(s string) bool {
	lowerSentence := strings.ToLower(s)
	if !strings.Contains(lowerSentence, " and ") && !strings.Contains(lowerSentence, " or ") {
		return false
	}
	for _, prefix := range []string{"if ", "when ", "whenever ", "unless ", "it is the case that ", "it is not the case that ", "it is the case both that "} {
		if strings.HasPrefix(strings.ToLower(s), prefix) {
			return false
		}
	}
	if strings.Contains(strings.ToLower(s), "whether or not") || strings.ContainsAny(s, "\"`“”") {
		return false
	}
	bs, _ := groupingBoundaries(s)
	if len(bs) == 0 {
		return false
	}
	if bs[0].start > 0 && s[bs[0].start-1] == ',' {
		// A comma before the connective makes it the main one (v32.0), so
		// the operator's scope ends at the comma.
		return false
	}
	first := strings.TrimSpace(s[:bs[0].start])
	lower := strings.ToLower(first)
	if strings.HasPrefix(lower, "both ") || strings.HasPrefix(lower, "either ") || strings.HasPrefix(lower, "(") {
		return false
	}
	if strings.Contains(lower, " both ") || strings.Contains(lower, " either ") {
		return false
	}
	if !sentenceOperator.MatchString(lower) && !impersonalEmbedding.MatchString(lower) && !embeddedClauseHead(lower) {
		return false
	}
	// Only whole-clause coordination; preserve coordinated noun phrases.
	return clauseOperandStart(s[bs[0].end:]) && !groupingBareFragment(s[bs[0].end:])
}

var relativeComma = regexp.MustCompile(`, (which|who|whom|whose) `)

// Commas belong to atoms unless a structural parser consumes them.
var structuralModalComma = regexp.MustCompile(`^(necessarily|possibly),\s*`)

var structuralCoordinationComma = regexp.MustCompile(`,\s+(and|or)\b`)

func comparisonCommas(s string) string {
	if !strings.Contains(s, ",") {
		return s
	}
	s = structuralModalComma.ReplaceAllString(s, "$1 ")
	// Only a visible clause boundary owns its comma. In particular, an
	// and/or inside atom{name} literal payloads is never syntax, and a
	// noun/verb phrase list is not evidence of two asserted clauses.
	matches := structuralCoordinationComma.FindAllStringIndex(groupingSyntax(s), -1)
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		left, right := strings.TrimSpace(s[:m[0]]), strings.TrimSpace(s[m[1]:])
		rightClause := clauseOperandStart(right) || (len(strings.Fields(right)) > 1 && framePlainClause(right))
		if framePlainClause(left) && rightClause {
			s = s[:m[0]] + s[m[0]+1:]
		}
	}
	return s
}

var nonrestrictiveRelative = regexp.MustCompile(`^(the [^,]+), (which|who|whom|whose) ([^,]+), ((?:is|are|was|were) [^,]+)$`)

// Directional weakening of a positive definite-subject predication, never identity.
func nonrestrictiveWeakening(premise, conclusion string) bool {
	p := normalizeClause(premise)
	if !nonrestrictiveRelative.MatchString(p) || hasOuterGateOperator(p) {
		return false
	}
	return equalClauses(nonrestrictiveRelative.ReplaceAllString(p, "$1 $2 $3 $4"), conclusion)
}

var termDeterminers = wordSet("the a an this that these those his her its their my your our every some each any no all many most few several both either neither such what which whose one two three own other another")
var termPhraseLinks = wordSet("and or but nor than as to who which whom whose where when while")
var termClauseStart = wordSet("that whether if the a an this these those his her its their my your our some every all no any he she they we i you it")
var irregularFiniteVerbs = wordSet("thought said knew saw felt found held told heard meant took gave made kept left believed understood forgot remembered")

// clauseInsideTerm reports a term that contains a finite verb followed by a
// complement (a capitalized word, a pronoun, a determiner, or that/whether/
// if): "the observer reckons alpha", "the observer fancies that alpha". Such a
// term ends inside an embedded clause, whatever the verb, so this is a
// structural rule rather than a list of attitude verbs. Text after a relative
// marker (who/which/that) is part of a noun phrase and is not scanned.
func clauseInsideTerm(term string) bool {
	raw := strings.Fields(strings.TrimSpace(term))
	for i := 1; i < len(raw)-1; i++ {
		w := strings.ToLower(strings.Trim(raw[i], ",;:"))
		if w == "who" || w == "which" || w == "whom" || w == "whose" || (w == "that" && i > 0 && !finiteVerbCue(strings.ToLower(raw[i-1]))) {
			return false
		}
		// A base-form verb after a plural subject ("the observers reckon
		// alpha") is also a finite verb.
		pluralSubject := strings.HasSuffix(strings.ToLower(raw[i-1]), "s") && !strings.HasSuffix(strings.ToLower(raw[i-1]), "ss") && !termDeterminers[w] && !framePrepositions[w] && !termPhraseLinks[w]
		if !(finiteVerbCue(w) || pluralSubject) || strings.Contains(w, "-") {
			continue
		}
		// A verb follows the head of a noun phrase, never a determiner or a
		// preposition: "an embedded agent", "the systems engineer" are noun
		// phrases.
		prev := strings.ToLower(strings.Trim(raw[i-1], ",;:"))
		if termDeterminers[prev] || framePrepositions[prev] || strings.HasSuffix(prev, "'s") || strings.HasSuffix(prev, "’s") {
			continue
		}
		// The hole is a term that ends just inside an embedded clause ("the
		// observer reckons alpha" + "is red"). A long run after the verb is a
		// whole clause bound to a propositional letter, which is fine.
		if len(raw)-1-i > 3 {
			continue
		}
		if strings.HasSuffix(raw[i], ",") {
			continue
		}
		next := strings.Trim(raw[i+1], ",;:")
		nl := strings.ToLower(next)
		if nl == "that" && i+2 < len(raw) {
			// "hypotheses that are true" is a subject relative (that + verb);
			// "reckons that alpha is red" is a complement (that + subject).
			after := strings.ToLower(strings.Trim(raw[i+2], ",;:"))
			if auxiliaryWords[after] || finiteVerbCue(after) || after == "not" {
				continue
			}
		}
		if termClauseStart[nl] || (len(next) > 0 && unicode.IsUpper([]rune(next)[0])) {
			return true
		}
		// Terms arrive lower-cased, so a following name looks like any word.
		// Strict: any word other than a preposition, conjunction, relative
		// marker or "to" opens a complement after a finite verb cue.
		if !framePrepositions[nl] && !termPhraseLinks[nl] {
			return true
		}
	}
	return false
}

func finiteVerbCue(w string) bool {
	if irregularFiniteVerbs[w] {
		return true
	}
	if len(w) < 4 || strings.HasSuffix(w, "ss") || strings.HasSuffix(w, "'s") || strings.HasSuffix(w, "us") || strings.HasSuffix(w, "is") {
		return false
	}
	return strings.HasSuffix(w, "s") || strings.HasSuffix(w, "ed")
}
