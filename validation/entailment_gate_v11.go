package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Catalog v11 predicate-logic forms: Darii/Ferio existential syllogisms,
// schematic existential introduction, and identity substitution. Each is
// appended after the v10 catalog so no previously certified step changes form.

// particularAffirmativeCandidates parses I-propositions "some S is/are M".
// Predicates opening with "not" are O-propositions and are never read as I.
func particularAffirmativeCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	if !strings.HasPrefix(statement, "some ") {
		return nil
	}
	remainder := strings.TrimPrefix(statement, "some ")
	result := make([]clausePair, 0)
	for _, separator := range []string{" is ", " are "} {
		for _, split := range appendSplitCandidates(nil, remainder, separator, false) {
			if split.right == "not" || strings.HasPrefix(split.right, "not ") {
				continue
			}
			result = append(result, split)
		}
	}
	return result
}

// particularNegativeCandidates parses O-propositions "some S is/are not P".
func particularNegativeCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	if !strings.HasPrefix(statement, "some ") {
		return nil
	}
	remainder := strings.TrimPrefix(statement, "some ")
	result := make([]clausePair, 0)
	for _, separator := range []string{" is not ", " are not "} {
		result = appendSplitCandidates(result, remainder, separator, false)
	}
	return result
}

// matchExistentialSyllogism certifies Darii: some S is M; every M is P;
// therefore some S is P. The I premise is matched in its written orientation
// only and the middle term must be the universal's subject (distributed).
func matchExistentialSyllogism(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, particular := range particularAffirmativeCandidates(premises[0].Statement) {
		for _, universal := range universalCandidates(premises[1].Statement) {
			for _, result := range particularAffirmativeCandidates(conclusion.Statement) {
				candidate := evaluateTermIdentities(
					clauseIdentity{fmt.Sprintf("premise %d predicate", premises[0].Number), particular.right, fmt.Sprintf("premise %d class", premises[1].Number), universal.left},
					clauseIdentity{fmt.Sprintf("premise %d subject", premises[0].Number), particular.left, fmt.Sprintf("conclusion %d subject", conclusion.Number), result.left},
					clauseIdentity{fmt.Sprintf("premise %d predicate", premises[1].Number), universal.right, fmt.Sprintf("conclusion %d predicate", conclusion.Number), result.right},
				)
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

// matchExistentialNegativeSyllogism certifies Ferio (and, by E-conversion,
// Festino): some S is M; no M is P (or no P is M); therefore some S is not P.
// The O conclusion is never converted.
func matchExistentialNegativeSyllogism(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, particular := range particularAffirmativeCandidates(premises[0].Statement) {
		for _, negative := range universalNegativeCandidates(premises[1].Statement) {
			for _, result := range particularNegativeCandidates(conclusion.Statement) {
				for _, terms := range [][2]string{{negative.left, negative.right}, {negative.right, negative.left}} {
					candidate := evaluateTermIdentities(
						clauseIdentity{fmt.Sprintf("premise %d predicate", premises[0].Number), particular.right, fmt.Sprintf("premise %d middle term", premises[1].Number), terms[0]},
						clauseIdentity{fmt.Sprintf("premise %d subject", premises[0].Number), particular.left, fmt.Sprintf("conclusion %d subject", conclusion.Number), result.left},
						clauseIdentity{fmt.Sprintf("premise %d excluded term", premises[1].Number), terms[1], fmt.Sprintf("conclusion %d excluded term", conclusion.Number), result.right},
					)
					best = betterOutcome(best, candidate)
					if best.matched {
						return best
					}
				}
			}
		}
	}
	return best
}

var schematicExistentialPattern = regexp.MustCompile(`^(?i:there (?:is|exists) (?:an?|some)) ([XYZ]) (?i:such that),? (.+)$`)

// Words that make a singular term a quantifier phrase, a pronoun, or a
// variable rather than a referring name; generalizing on or substituting for
// such a "term" is not first-order existential introduction/Leibniz's law.
var nonReferringTermWords = map[string]bool{
	"every": true, "all": true, "each": true, "any": true, "some": true, "no": true, "none": true,
	"nothing": true, "nobody": true, "something": true, "someone": true, "somebody": true,
	"everything": true, "everyone": true, "everybody": true, "anything": true, "anyone": true,
	"anybody": true, "a": true, "an": true, "few": true, "many": true, "most": true, "several": true,
	"both": true, "either": true, "neither": true, "only": true, "not": true, "whoever": true,
	"whatever": true, "who": true, "which": true, "that": true, "what": true,
	"he": true, "she": true, "it": true, "they": true, "him": true, "her": true, "them": true,
	"his": true, "hers": true, "its": true, "their": true, "theirs": true, "himself": true,
	"herself": true, "itself": true, "themselves": true, "one": true, "x": true, "y": true, "z": true,
	// Verbs, auxiliaries and connectives mean the "term" spans a clause boundary.
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
	"am": true, "has": true, "have": true, "had": true, "does": true, "do": true, "did": true,
	"will": true, "and": true, "or": true, "if": true, "then": true, "than": true, "nor": true,
}

// Quantifier determiners whose presence in the context could bind an
// anaphoric definite ("each student's advisor praised the student").
var contextQuantifierWords = map[string]bool{
	"every": true, "all": true, "each": true, "any": true, "some": true, "no": true, "none": true,
	"nothing": true, "nobody": true, "something": true, "someone": true, "somebody": true,
	"everything": true, "everyone": true, "everybody": true, "anything": true, "anyone": true,
	"anybody": true, "few": true, "many": true, "most": true, "several": true, "both": true,
}

// Intensional, modal, existence and complement-clause vocabulary. Inside such
// contexts neither substitution of co-referring names nor existential
// generalization on a name is truth-preserving, so the forms stay opaque.
var opaqueContextWords = regexp.MustCompile(`(?i)\b(?:believes?|believed|thinks?|thought|knows?|knew|known|wants?|wanted|hopes?|hoped|fears?|feared|seeks?|sought|wish|wishes|wished|doubts?|doubted|imagines?|imagined|supposes?|supposed|says?|said|claims?|claimed|asserts?|asserted|denies|deny|denied|expects?|expected|intends?|intended|realizes?|realized|recognizes?|recognized|notices?|noticed|learns?|learned|wonders?|wondered|prefers?|preferred|necessarily|possibly|necessary|possible|might|may|could|would|should|must|can|exists?|existed|existence|real|fictional|imaginary|called|named|that|whether)\b`)

var truthFunctionalWrapper = regexp.MustCompile(`(?i)\bit is (?:not )?the case that\b`)

// referringTerm reports whether term (in comparison form) is a candidate
// singular term. Definite descriptions are refused when the context carries
// a quantifier that could bind them.
func referringTerm(term string, contextHasQuantifier bool) bool {
	words := strings.Fields(term)
	if len(words) == 0 || strings.TrimSpace(term) != term {
		return false
	}
	for _, word := range words {
		if nonReferringTermWords[strings.Trim(word, "'’")] {
			return false
		}
	}
	if contextHasQuantifier {
		switch words[0] {
		case "the", "this", "that", "these", "those", "such":
			return false
		}
	}
	return true
}

func transparentContext(statement string) (ok bool, hasQuantifier bool) {
	if strings.ContainsAny(statement, "\"`“”‘") || strings.Contains(statement, " '") || strings.HasPrefix(statement, "'") {
		return false, false
	}
	// The truth-functional wrappers are the only complement "that" the
	// register treats as transparent.
	if opaqueContextWords.MatchString(truthFunctionalWrapper.ReplaceAllString(statement, " ")) {
		return false, false
	}
	for _, word := range strings.Fields(strings.ToLower(statement)) {
		if contextQuantifierWords[strings.Trim(word, ",;:()")] {
			hasQuantifier = true
		}
	}
	return true, hasQuantifier
}

// replaceTermOccurrences replaces every word-bounded, case-insensitive
// occurrence of term in text. It refuses (ok=false) when an occurrence is
// adjacent to a capitalized word or a hyphen, where the phrase may be part of a
// longer name ("Paris Hilton", "Ada-Lovelace") rather than the term itself.
func replaceTermOccurrences(text, term, replacement string) (result string, count int, ok bool) {
	lower := strings.ToLower(text)
	term = strings.ToLower(term)
	if term == "" || len(lower) != len(text) {
		return "", 0, false
	}
	var builder strings.Builder
	position := 0
	for position <= len(text) {
		index := strings.Index(lower[position:], term)
		if index < 0 {
			break
		}
		start := position + index
		end := start + len(term)
		if (start > 0 && wordRune(lastRune(text[:start]))) || (end < len(text) && wordRune(firstRune(text[end:]))) {
			builder.WriteString(text[position : start+1])
			position = start + 1
			continue
		}
		if start > 0 && (text[start-1] == '-' || (text[start-1] == ' ' && startsCapitalizedWord(lastWord(text[:start-1])))) {
			return "", 0, false
		}
		if end < len(text) && (text[end] == '-' || (text[end] == ' ' && startsCapitalizedWord(text[end+1:]))) {
			return "", 0, false
		}
		builder.WriteString(text[position:start])
		builder.WriteString(replacement)
		count++
		position = end
	}
	if position < len(text) {
		builder.WriteString(text[position:])
	}
	return builder.String(), count, true
}

func wordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func lastRune(value string) rune {
	r, _ := utf8.DecodeLastRuneInString(value)
	return r
}

func firstRune(value string) rune {
	r, _ := utf8.DecodeRuneInString(value)
	return r
}

func lastWord(value string) string {
	if index := strings.LastIndexAny(value, " "); index >= 0 {
		return value[index+1:]
	}
	return value
}

func startsCapitalizedWord(value string) bool {
	return value != "" && unicode.IsUpper(firstRune(value))
}

// matchSchematicExistentialIntroduction certifies A(t); therefore there is an
// X such that A(X), where A(X) is A(t) with EVERY word-bounded occurrence of
// the singular term t replaced by the one bound placeholder.
func matchSchematicExistentialIntroduction(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	match := schematicExistentialPattern.FindStringSubmatch(normalizeClausePreservingCase(conclusion.Statement))
	if match == nil {
		return schemaMatchOutcome{}
	}
	variable := []rune(match[1])[0]
	replacedMatrix, counts := replaceSchematicVariables(match[2])
	outcome := schemaMatchOutcome{traction: true}
	if len(counts) != 1 || counts[variable] == 0 {
		outcome.mismatch = fmt.Sprintf("conclusion %d matrix must use the bound placeholder %c and no other placeholder", conclusion.Number, variable)
		return outcome
	}
	template := comparisonClause(replacedMatrix)
	if !strings.ContainsFunc(template, unicode.IsLetter) {
		outcome.mismatch = fmt.Sprintf("conclusion %d matrix has no predicate besides the placeholder", conclusion.Number)
		return outcome
	}

	premiseText := normalizeClausePreservingCase(premises[0].Statement)
	if _, premiseVariables := replaceSchematicVariables(premiseText); len(premiseVariables) != 0 {
		outcome.mismatch = fmt.Sprintf("premise %d contains a standalone placeholder", premises[0].Number)
		return outcome
	}

	marker := string(schematicVariableMarkers[variable])
	outcome.mismatch = fmt.Sprintf("conclusion %d matrix is not premise %d with every occurrence of one singular term replaced by %c", conclusion.Number, premises[0].Number, variable)
	for _, substitution := range unifySchematicTemplate(template, premiseText, nil) {
		term := substitution[variable]
		transparent, hasQuantifier := substitutionContext(premiseText, term)
		if !transparent || !substitutionTerm(term, hasQuantifier) {
			continue
		}
		replaced, count, ok := replaceTermOccurrences(premiseText, term, marker)
		if !ok || count != counts[variable] {
			continue
		}
		if comparisonClause(replaced) == template {
			return schemaMatchOutcome{matched: true, score: 1, traction: true}
		}
	}
	return outcome
}

// matchIdentitySubstitution certifies a is identical to b; A(a); therefore
// A(b) (or, by symmetry of identity, A(b); therefore A(a)), where EVERY
// word-bounded occurrence of the replaced name is substituted. Only the
// explicit "is identical to" relation counts: a bare "a is b" cannot be told
// apart from predication ("Ada is a doctor").
func matchIdentitySubstitution(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	identities := identityCandidates(premises[0].Statement)
	if len(identities) == 0 {
		return schemaMatchOutcome{}
	}
	premiseText := normalizeClausePreservingCase(premises[1].Statement)
	outcome := schemaMatchOutcome{traction: true}

	outcome.mismatch = fmt.Sprintf("conclusion %d is not premise %d with every occurrence of one identical term replaced by the other", conclusion.Number, premises[1].Number)
	for _, identity := range identities {
		for _, direction := range [][2]string{{identity.left, identity.right}, {identity.right, identity.left}} {
			transparent, premiseQuantifier := substitutionContext(premiseText, direction[0])
			conclusionTransparent, conclusionQuantifier := substitutionContext(normalizeClausePreservingCase(conclusion.Statement), direction[1])
			hasQuantifier := premiseQuantifier || conclusionQuantifier
			if !transparent || !conclusionTransparent || !substitutionTerm(identity.left, hasQuantifier) || !substitutionTerm(identity.right, hasQuantifier) {
				continue
			}
			replaced, count, ok := replaceTermOccurrences(premiseText, direction[0], direction[1])
			if !ok || count == 0 {
				continue
			}
			if comparisonClause(replaced) == comparisonClause(conclusion.Statement) {
				return schemaMatchOutcome{matched: true, score: 1, traction: true}
			}
		}
	}
	return outcome
}

// quantifiedInstanceSubject reports whether an instance-position subject is a
// quantifier phrase that universal (negative) instantiation must not treat as
// an individual. Before v11, "No cat is a hunter; every hunter is alert;
// therefore no cat is alert" passed as universal instantiation because "no
// cat" was read as a name. Downward- and non-monotone determiners are refused
// outright; "some" subjects are now certified by the dedicated existential
// syllogism forms (Darii/Ferio/Festino). Upward-monotone determiners (every,
// all, most, many, at least) keep their v10 behaviour, which is valid.
func quantifiedInstanceSubject(subject string) bool {
	subject = normalizeClause(subject)
	if strings.HasPrefix(subject, "it is not the case") {
		return true
	}
	first, _, _ := strings.Cut(subject, " ")
	switch first {
	case "some", "no", "nobody", "nothing", "none", "noone", "few", "fewer", "less", "exactly", "neither", "not", "only", "hardly", "scarcely", "never", "little":
		return true
	case "at":
		return strings.HasPrefix(subject, "at most ")
	}
	return false
}
