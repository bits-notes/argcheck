package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
)

var sourceQuoteCitations = regexp.MustCompile(`\[[0-9,\s]+\]`)
var sourceQuoteNegatedContraction = regexp.MustCompile(`(?i)n['’]t\b`)
var sourceQuoteWords = regexp.MustCompile(`[\p{L}\p{N}]+`)

// SourceContentAligned is deliberately stricter than the display aligner's
// partial-clause LCS match. Require >=80% of the statement's distinct content
// words AND >=60% Jaccard overlap: extra source detail is allowed, but neither
// a tiny common fragment nor a long quote can license mostly invented content.
// This is lexical evidence, not a semantic fidelity or entailment certificate.
func SourceContentAligned(statement, quote string) bool {
	a, b := sourceContentWords(statement), sourceContentWords(quote)
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	shared := 0
	for word := range a {
		if b[word] {
			shared++
		}
	}
	// Do not erase polarity or changed numbers under an otherwise high overlap.
	for _, words := range []map[string]bool{a, b} {
		for word := range words {
			if word == "not" || word == "no" || word == "never" || strings.IndexFunc(word, unicode.IsDigit) >= 0 {
				if a[word] != b[word] {
					return false
				}
			}
		}
	}
	return 5*shared >= 4*len(a) && 5*shared >= 3*(len(a)+len(b)-shared)
}

func sourceContentWords(text string) map[string]bool {
	text = sourceQuoteCitations.ReplaceAllString(text, " ")
	text = sourceQuoteNegatedContraction.ReplaceAllString(text, " not")
	words := sourceQuoteWords.FindAllString(cases.Fold().String(text), -1)
	result := map[string]bool{}
	for _, word := range words {
		switch word {
		case "a", "an", "the", "is", "are", "was", "were", "be", "been", "being", "of", "to", "in", "on", "at", "for", "by", "with", "and", "or", "it", "its", "that", "this", "then", "therefore", "so", "thus", "hence":
			continue
		}
		result[word] = true
	}
	return result
}

// ValidateSourceQuote returns half-open code-point offsets in the original
// source. A present quote must pass even when the mapping also has anchor lines.
func ValidateSourceQuote(source, statement, quote string) (int, int, error) {
	normalized, _ := normalizeSourceQuote(quote)
	normalized = strings.TrimRightFunc(normalized, func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSpace(r) })
	if utf8.RuneCountInString(normalized) < 12 {
		return 0, 0, fmt.Errorf("from_source_quote must contain at least 12 normalized characters")
	}
	haystack, offsets := normalizeSourceQuote(source)
	index := strings.Index(haystack, normalized)
	if index < 0 {
		return 0, 0, fmt.Errorf("from_source_quote is not an excerpt of source_material")
	}
	if !SourceContentAligned(statement, quote) {
		return 0, 0, fmt.Errorf("%s", sourceQuoteOverlapMessage(source, statement, quote))
	}
	start := utf8.RuneCountInString(haystack[:index])
	end := start + utf8.RuneCountInString(normalized)
	return offsets[start], offsets[end-1] + 1, nil
}

func normalizeSourceQuote(text string) (string, []int) {
	var chars []rune
	var offsets []int
	fold := cases.Fold()
	for i, r := range []rune(text) {
		if unicode.IsSpace(r) {
			if len(chars) == 0 || chars[len(chars)-1] == ' ' {
				continue
			}
			r = ' '
		}
		switch r {
		case '‘', '’', '‛', '‚':
			r = '\''
		case '“', '”', '„', '‟':
			r = '"'
		case '‐', '‑', '‒', '–', '—', '―', '−':
			r = '-'
		}
		for _, folded := range fold.String(string(r)) {
			chars = append(chars, folded)
			offsets = append(offsets, i)
		}
	}
	if len(chars) > 0 && chars[len(chars)-1] == ' ' {
		chars = chars[:len(chars)-1]
		offsets = offsets[:len(offsets)-1]
	}
	return string(chars), offsets
}

// ValidateRestoredLineMappings is shared by prompt validation and transforms.
// Missing source fails closed for quotes; old outputs need no source context.
func ValidateRestoredLineMappings(source, argument string, mappings []FormalizationLineMapping) error {
	var parsed *NumberedArgument
	for _, mapping := range mappings {
		if mapping.FromSourceQuote == nil {
			continue
		}
		if parsed == nil {
			value, result := ParseNumberedArgumentAllowingNoConclusions(argument)
			if !result.Passed {
				return fmt.Errorf("parse restored argument: %s", result.Message)
			}
			parsed = &value
		}
		if mapping.RepairedLine < 1 || mapping.RepairedLine > len(parsed.Steps) {
			return fmt.Errorf("repaired line %d outside argument", mapping.RepairedLine)
		}
		if _, _, err := ValidateSourceQuote(source, parsed.Steps[mapping.RepairedLine-1].Statement, *mapping.FromSourceQuote); err != nil {
			return fmt.Errorf("repaired line %d: %w", mapping.RepairedLine, err)
		}
	}
	return nil
}

// SourceStatementAligned checks whole sentences at the quote overlap threshold.
// Split only terminal punctuation followed by whitespace/end, keeping decimal
// numbers intact. Do not fish for matching windows within unrelated clauses.
func SourceStatementAligned(source, statement string) bool {
	runes := []rune(source)
	start := 0
	for i, r := range runes {
		terminal := (r == '.' || r == '!' || r == '?') && (i+1 == len(runes) || unicode.IsSpace(runes[i+1]))
		if !terminal {
			continue
		}
		if SourceContentAligned(statement, string(runes[start:i+1])) {
			return true
		}
		start = i + 1
	}
	return SourceContentAligned(statement, string(runes[start:]))
}

// Diagnostics only: acceptance continues to use SourceContentAligned unchanged.
func sourceQuoteOverlapMessage(source, statement, quote string) string {
	message := "from_source_quote does not substantially overlap the repaired line's content"
	present := sourceContentWords(quote)
	content := sourceContentWords(statement)
	missing := []string{}
	seen := map[string]bool{}
	diagnosticStatement := sourceQuoteCitations.ReplaceAllString(statement, " ")
	diagnosticStatement = sourceQuoteNegatedContraction.ReplaceAllString(diagnosticStatement, " not")
	for _, word := range sourceQuoteWords.FindAllString(cases.Fold().String(diagnosticStatement), -1) {
		// Omit these function words from the content hint only; the overlap
		// check still counts them exactly as before.
		if word != "if" && word != "when" && word != "from" && content[word] && !present[word] && !seen[word] {
			missing = append(missing, word)
			seen[word] = true
		}
	}
	if len(missing) > 0 {
		message += "; content words missing from quote: " + strings.Join(missing, ", ")
	}
	return message + sourceQuoteSentenceSuggestion(source, quote)
}

func sourceSentences(source string) []string {
	runes := []rune(source)
	start := 0
	var sentences []string
	for i, r := range runes {
		if (r == '.' || r == '!' || r == '?') && (i+1 == len(runes) || unicode.IsSpace(runes[i+1])) {
			if sentence := strings.TrimSpace(string(runes[start : i+1])); sentence != "" {
				sentences = append(sentences, sentence)
			}
			start = i + 1
		}
	}
	if sentence := strings.TrimSpace(string(runes[start:])); sentence != "" {
		sentences = append(sentences, sentence)
	}
	return sentences
}

func sourceQuoteSentenceSuggestion(source, quote string) string {
	normalized, _ := normalizeSourceQuote(quote)
	normalized = strings.TrimRightFunc(normalized, func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSpace(r) })
	if normalized == "" {
		return ""
	}
	for _, sentence := range sourceSentences(source) {
		whole, _ := normalizeSourceQuote(sentence)
		trimmed := strings.TrimRightFunc(whole, func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSpace(r) })
		if strings.Contains(whole, normalized) && trimmed != normalized {
			return fmt.Sprintf("; quote the whole source sentence: %q", sentence)
		}
	}
	return ""
}
