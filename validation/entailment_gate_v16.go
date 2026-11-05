package validation

import "strings"

// Parentheses alone delimit both sides of a scaffolded biconditional, even when
// either side contains another scaffold: "(X) if and only if (Y)". A leading
// "it is the case that" is the ordinary clause wrapper and is stripped first, so
// "It is the case that (X) if and only if (Y)" is the same clause. No other
// prefix is special: "it holds that" is ordinary text (catalog v18).
func scaffoldedBiconditional(statement string) (string, bool) {
	s := stripClauseWrappers(normalizeClause(statement))
	return s, strings.HasPrefix(s, "(")
}

func isScaffoldedBiconditional(statement string) bool {
	_, ok := scaffoldedBiconditional(statement)
	return ok
}

func scaffoldedBiconditionalCandidates(s string) []clausePair {
	group := func(s string) (string, string, bool) {
		if !strings.HasPrefix(s, "(") {
			return "", "", false
		}
		depth := 0
		for i, r := range s {
			switch r {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					clause := normalizeClause(s[1:i])
					return clause, s[i+1:], clause != ""
				}
			}
		}
		return "", "", false
	}
	left, rest, ok := group(s)
	const separator = " if and only if "
	if !ok || !strings.HasPrefix(rest, separator) {
		return nil
	}
	right, rest, ok := group(strings.TrimPrefix(rest, separator))
	if !ok || rest != "" {
		return nil
	}
	return []clausePair{{left: left, right: right}}
}

// thatNounPhrase recognizes a definite head with a complement, not a term
// inside that complement. Context boundaries are checked separately below.
func thatNounPhrase(term string) bool {
	head, complement, ok := strings.Cut(normalizeClause(term), " that ")
	return ok && strings.HasPrefix(head, "the ") && len(strings.Fields(head)) >= 2 &&
		referringTerm(head, false) && complement != "" && !strings.ContainsAny(term, "\"`“”‘")
}

// wholeTermContext is deliberately a small surface register: a complete
// copular subject or the terminal object of a simple predication. Masking ONLY
// that whole span allows its complement vocabulary, while the surrounding
// context still undergoes every existing opacity/quantifier check. A partial
// subject, an object inside a that-clause, and a possessive modifier fail.
func wholeTermContext(statement, term string) bool {
	s, term := normalizeClause(statement), normalizeClause(term)
	plainTerm := term
	if strings.HasPrefix(term, "the ") || strings.Contains(term, "'s ") || strings.Contains(term, "’s ") {
		plainTerm = strings.TrimSuffix(strings.TrimSuffix(term, " claims"), " claim")
	}
	plainTransparent, _ := transparentContext(plainTerm)
	if !(thatNounPhrase(term) || (referringTerm(term, false) && plainTransparent)) || strings.Count(s, term) != 1 {
		return false
	}
	if strings.HasPrefix(s, term+" ") {
		rest := strings.TrimPrefix(s, term)
		for _, copula := range []string{" is ", " am ", " are "} {
			if strings.HasPrefix(rest, copula) {
				predicate := strings.TrimPrefix(rest, copula)
				// Another copula would make this a partial embedded subject.
				if len(instanceCandidates(predicate)) != 0 {
					return false
				}
				// In a copular predicate, "claim" is a noun, not the
				// attitude verb caught by opaqueContextWords ("Ann claims").
				context := "term" + rest
				context = strings.TrimSuffix(strings.TrimSuffix(context, " claims"), " claim")
				ok, quantified := transparentContext(context)
				return ok && !quantified
			}
		}
	}
	if strings.HasSuffix(s, " "+term) {
		prefix := strings.TrimSuffix(s, term)
		ok, quantified := transparentContext(prefix + "term")
		if !ok || quantified {
			return false
		}
		// One ordinary verb after a referring subject (Ada repeats ...),
		// or an explicit copular relation (Ada is opposed to ...).
		words := strings.Fields(prefix)
		if len(words) >= 2 && referringTerm(strings.Join(words[:len(words)-1], " "), false) {
			switch words[len(words)-1] {
			case "of", "the", "a", "an", "and", "or", "if", "then":
				return false
			}
			return true
		}
		for _, pair := range instanceCandidates(prefix + "term") {
			if referringTerm(pair.left, false) && strings.HasSuffix(pair.right, " to term") {
				return true
			}
		}
	}
	return false
}

func substitutionContext(statement, term string) (bool, bool) {
	if wholeTermContext(statement, term) {
		return true, false
	}
	return transparentContext(statement)
}

func substitutionTerm(term string, hasQuantifier bool) bool {
	return referringTerm(term, hasQuantifier) || (!hasQuantifier && thatNounPhrase(term))
}
