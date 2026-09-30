package validation

import (
	"fmt"
	"strings"
	"unicode"
)

// FallbackBridge renders the supplied connecting premise and the conclusion
// text that follows "Therefore, ". The word after "then " keeps its capital
// unless it is ordinary: not an all-caps token of length >= 2 or "I", not a
// label followed by a single capital letter or digit, and either a function
// word (so "then If" becomes "then if") or a word whose lowercase form occurs
// as a whole word elsewhere in the argument. The stored text is exactly the
// text the entailment gate certifies.
func FallbackBridge(antecedent, conclusion, argumentText string) (bridge, renderedConclusion string) {
	conclusion = strings.TrimSpace(conclusion)
	rendered := lowercaseLeadingOrdinaryWord(conclusion, argumentText)
	return fmt.Sprintf("If %s, then %s", antecedent, rendered), rendered
}

func lowercaseLeadingOrdinaryWord(phrase, argument string) string {
	fields := strings.Fields(phrase)
	if len(fields) == 0 {
		return phrase
	}
	stripped := stripLeadingOrdinaryWord(fields[0])
	if stripped == "" || !startsWithCapital(stripped) {
		return phrase
	}
	next := ""
	if len(fields) > 1 {
		next = fields[1]
	}
	if (len([]rune(stripped)) == 1 && stripped != "A") || isAllCapsToken(stripped) || isLabelToken(next) {
		return phrase
	}
	lower := strings.ToLower(stripped)
	// The leading word is still capitalized in argument, so searching for the
	// lowercase form does not count this leading position.
	if !isOrdinaryFunctionWord(lower) && !containsLowercaseWholeWord(argument, lower) {
		return phrase
	}
	return lowerFirstLetter(phrase)
}

func stripLeadingOrdinaryWord(word string) string {
	trimmed := strings.TrimFunc(word, unicode.IsPunct)
	for _, suffix := range []string{"'s", "’s"} {
		if strings.HasSuffix(trimmed, suffix) {
			trimmed = strings.TrimSuffix(trimmed, suffix)
			break
		}
	}
	return strings.TrimFunc(trimmed, unicode.IsPunct)
}

func startsWithCapital(word string) bool {
	for _, r := range word {
		return unicode.IsLetter(r) && unicode.IsUpper(r)
	}
	return false
}

func isAllCapsToken(word string) bool {
	if len([]rune(word)) < 2 {
		return false
	}
	sawLetter := false
	for _, r := range word {
		if !unicode.IsLetter(r) {
			continue
		}
		sawLetter = true
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return sawLetter
}

func isLabelToken(field string) bool {
	stripped := strings.TrimFunc(field, unicode.IsPunct)
	if stripped == "" {
		return false
	}
	runes := []rune(stripped)
	if len(runes) == 1 && unicode.IsUpper(runes[0]) {
		return true
	}
	for _, r := range runes {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isOrdinaryFunctionWord(word string) bool {
	switch word {
	case "a", "an", "the", "if", "in", "on", "by", "with", "without", "for",
		"from", "to", "of", "at", "as", "and", "or", "but", "so", "yet",
		"not", "no", "all", "some", "any", "each", "every", "either", "neither", "both",
		"even", "only", "rather", "instead", "under", "over", "after", "before", "because", "since",
		"although", "though", "while", "when", "whenever", "where", "there", "this", "that", "these",
		"those", "it", "its", "we", "they", "you", "he", "she", "one", "such", "most",
		"many", "much", "more", "less", "then", "thus", "therefore", "hence", "given", "unless",
		"until", "once", "whether", "what", "which", "who", "how", "why", "via", "per",
		"into", "onto", "upon", "about", "against", "between", "through", "during", "within", "among",
		"across":
		return true
	default:
		return false
	}
}

func containsLowercaseWholeWord(text, word string) bool {
	if text == "" || word == "" {
		return false
	}
	haystack := []rune(text)
	needle := []rune(word)
	limit := len(haystack) - len(needle)
	for i := 0; i <= limit; i++ {
		if i > 0 && bridgeWordRune(haystack[i-1]) {
			continue
		}
		matched := true
		for j, r := range needle {
			if haystack[i+j] != r {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		end := i + len(needle)
		if end < len(haystack) && bridgeWordRune(haystack[end]) {
			continue
		}
		return true
	}
	return false
}

func lowerFirstLetter(phrase string) string {
	runes := []rune(phrase)
	for i, r := range runes {
		if !unicode.IsLetter(r) {
			continue
		}
		if !unicode.IsUpper(r) {
			return phrase
		}
		runes[i] = unicode.ToLower(r)
		return string(runes)
	}
	return phrase
}

func bridgeWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
