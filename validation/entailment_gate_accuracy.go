package validation

import (
	"regexp"
	"strings"
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

// embeddedClauseHead is the stricter test used for coordination under a
// leading operator: only an attitude verb with a that/whether/if complement
// counts ("we should expect X to possess Y" is not an embedding).
func embeddedClauseHead(s string) bool {
	return embeddedTerm(s, true)
}

func embeddedTerm(s string, needComplementizer bool) bool {
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

var nonSingularStem = regexp.MustCompile(`(?:^|\s)(?:clothes|species|means|goods|news|odds|thanks|series|savings|earnings|belongings|remains|surroundings|axes|arms|customs|glasses)$`)
var identityOpaqueExtra = regexp.MustCompile(`(?i)\b(?:allegedly|supposedly|reportedly|dream(?:s|ed|ing)?|pretend(?:s|ed|ing)?|called|named|known as)\b`)
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

// comparisonCommas keeps the one comma that changes meaning in matching: a
// comma before which/who/whom/whose marks a non-restrictive relative ("The
// robots, which are red, are warm" also says every robot is red). Every other
// comma stays insignificant, as before. Known gap: a comma that separates
// verbs ("eats, shoots, and leaves") is still ignored.
func comparisonCommas(s string) string {
	if !strings.Contains(s, ",") {
		return s
	}
	marked := relativeComma.ReplaceAllString(s, " \x00$1 ")
	marked = strings.ReplaceAll(marked, ",", "")
	return strings.ReplaceAll(marked, "\x00", ", ")
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
