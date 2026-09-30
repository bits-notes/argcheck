package validation

import (
	"regexp"
	"strings"
	"unicode"
)

// SourceSentenceRange uses zero-based Index and Unicode code-point [Start, End) offsets.
// Workflow sentence references use the one-based position (Index + 1).
type SourceSentenceRange struct {
	Index int
	Start int
	End   int
}

var (
	abbreviationTokens = map[string]struct{}{
		"e.g": {}, "i.e": {}, "etc": {}, "vs": {},
		"dr": {}, "mr": {}, "mrs": {}, "ms": {},
		"prof": {}, "st": {},
	}
	// A bare list marker is only a sentence when nothing follows it. "3." and
	// "(a)" at the start of the input or a paragraph belong to the next sentence.
	bareListMarkerPattern = regexp.MustCompile(`^(?:[0-9]+\.|\([0-9]+\)\.?|\([A-Za-z]\)\.?)$`)
)

// SegmentSourceSentences preserves source offsets, including numbered-list prefixes.
func SegmentSourceSentences(source string) []SourceSentenceRange {
	runes := []rune(source)
	sentences := make([]SourceSentenceRange, 0)
	start := 0
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\n' {
			if bareListMarkerRange(runes, start, i) {
				continue
			}
			appendSentence(&sentences, runes, start, i)
			start = i + 1
			continue
		}
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		if r == '.' && skipDot(runes, i) {
			continue
		}
		end := i + 1
		for end < len(runes) && isCloser(runes[end]) {
			end++
		}
		if end < len(runes) && !unicode.IsSpace(runes[end]) {
			continue
		}
		if bareListMarkerRange(runes, start, end) {
			continue
		}
		appendSentence(&sentences, runes, start, end)
		for end < len(runes) && unicode.IsSpace(runes[end]) {
			end++
		}
		start = end
		i = end - 1
	}
	if start < len(runes) {
		appendSentence(&sentences, runes, start, len(runes))
	}
	return sentences
}

func bareListMarkerRange(runes []rune, from, to int) bool {
	if from < 0 {
		from = 0
	}
	if to > len(runes) {
		to = len(runes)
	}
	if from >= to {
		return false
	}
	return bareListMarkerPattern.MatchString(strings.TrimSpace(string(runes[from:to])))
}

func appendSentence(sentences *[]SourceSentenceRange, runes []rune, from, to int) {
	for from < to && unicode.IsSpace(runes[from]) {
		from++
	}
	for to > from && unicode.IsSpace(runes[to-1]) {
		to--
	}
	if from >= to {
		return
	}
	*sentences = append(*sentences, SourceSentenceRange{
		Index: len(*sentences),
		Start: from,
		End:   to,
	})
}

func skipDot(runes []rune, index int) bool {
	if index > 0 && index+1 < len(runes) && unicode.IsDigit(runes[index-1]) && unicode.IsDigit(runes[index+1]) {
		return true
	}
	return isAbbreviationDot(runes, index)
}

func isAbbreviationDot(runes []rune, dot int) bool {
	if dot <= 0 {
		return false
	}
	start := dot
	for start > 0 {
		r := runes[start-1]
		if unicode.IsLetter(r) || r == '.' {
			start--
			continue
		}
		break
	}
	if start == dot {
		return false
	}
	token := strings.ToLower(string(runes[start:dot]))
	if _, ok := abbreviationTokens[token]; ok {
		return true
	}
	return !strings.Contains(token, ".") && len([]rune(token)) == 1 && unicode.IsLetter([]rune(token)[0])
}

func isCloser(r rune) bool {
	switch r {
	case '"', '\'', '”', '’', ')', ']', '}', '»':
		return true
	default:
		return false
	}
}
