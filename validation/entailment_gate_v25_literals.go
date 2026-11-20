package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var folLiteralStart = regexp.MustCompile(`(?i)\b(?:atom|name)\{`)
var folNameEscape = regexp.MustCompile(`^name\{([a-z][0-9]*)\}$`)

// Hide literal contents from connective/binder scans. Positions remain byte-exact.
func folSyntax(s string) string {
	spans := folLiteralSpans(s)
	if len(spans) == 0 {
		return s
	}
	b := []byte(s)
	for _, span := range spans {
		for i := span[0]; i < span[1]; i++ {
			b[i] = ' '
		}
	}
	return string(b)
}
func folLiteralSpans(s string) [][2]int {
	if !strings.Contains(s, "{") {
		return nil
	}
	var spans [][2]int
	for pos := 0; pos < len(s); {
		m := folLiteralStart.FindStringIndex(s[pos:])
		if m == nil {
			break
		}
		start, end := pos+m[0], pos+m[1]
		quoted, escaped := false, false
		for ; end < len(s); end++ {
			c := s[end]
			if escaped {
				escaped = false
				continue
			}
			if quoted && c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				quoted = !quoted
				continue
			}
			if !quoted && c == '}' {
				end++
				break
			}
		}
		spans = append(spans, [2]int{start, end})
		pos = end
	}
	return spans
}

// Preserve literal payloads across the legacy prose normalizer.
func folNormalizeLiterals(s string) (string, bool) {
	spans := folLiteralSpans(s)
	if len(spans) == 0 {
		return "", false
	}
	masked := s
	replacements := []string{}
	for i := len(spans) - 1; i >= 0; i-- {
		span := spans[i]
		key := fmt.Sprintf("zzfolquotedliteral%dzz", i)
		for strings.Contains(strings.ToLower(s), key) {
			key += "z"
		}
		replacements = append(replacements, key, strings.ToLower(s[span[0]:span[1]]))
		masked = masked[:span[0]] + key + masked[span[1]:]
	}
	return strings.NewReplacer(replacements...).Replace(normalizeClause(masked)), true
}

// atom{"verbatim symbol", "constant", x} fixes lexical boundaries and arity.
// Strings are constants; unquoted variables must be bound in the current scope.
func folLiteralAtom(s string, bound map[string]bool) (*folFormula, error) {
	return allStrictRules().folLiteralAtom(s, bound)
}

func (policy strictRules) folLiteralAtom(s string, bound map[string]bool) (*folFormula, error) {
	bad := func() (*folFormula, error) {
		return nil, fmt.Errorf("malformed atom literal: quote its symbol and constants; bind variable arguments")
	}
	if !strings.HasPrefix(s, "atom{") || !strings.HasSuffix(s, "}") {
		return bad()
	}
	raw := s[5 : len(s)-1]
	var parts []string
	start := 0
	quoted, escaped := false, false
	for i, c := range raw {
		if escaped {
			escaped = false
			continue
		}
		if quoted && c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			quoted = !quoted
		}
		if !quoted && c == ',' {
			parts = append(parts, strings.TrimSpace(raw[start:i]))
			start = i + 1
		}
	}
	parts = append(parts, strings.TrimSpace(raw[start:]))
	var symbol string
	if json.Unmarshal([]byte(parts[0]), &symbol) != nil || symbol == "" {
		return bad()
	}
	f := &folFormula{op: "atom", symbol: strings.ToLower(symbol)}
	for _, p := range parts[1:] {
		if bound[p] {
			f.terms = append(f.terms, "$"+p)
			continue
		}
		var name string
		if json.Unmarshal([]byte(p), &name) != nil || name == "" || policy.indefiniteTermRefused(name) {
			return bad()
		}
		f.terms = append(f.terms, folLiteralConstant(name))
	}
	return f, nil
}

// Keep payload whitespace and brackets out of numbered-argument preprocessing.
func folQuote(s string) string {
	var out strings.Builder
	out.WriteByte('"')
	previousSpace := false
	for _, r := range s {
		space := unicode.IsSpace(r)
		if r == '[' || r == ']' || (space && (r != ' ' || previousSpace)) {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			encoded, _ := json.Marshal(string(r))
			out.Write(encoded[1 : len(encoded)-1])
		}
		previousSpace = space
	}
	out.WriteByte('"')
	return out.String()
}
func folRenderLiteral(f *folFormula) string {
	s := "atom{" + folQuote(f.symbol)
	for _, t := range f.terms {
		if strings.HasPrefix(t, "$") {
			s += ", " + t[1:]
		} else {
			s += ", " + folQuote(strings.TrimPrefix(t, "\x00"))
		}
	}
	return s + "}"
}

func folOpaqueWord(f *folFormula, word string) bool {
	if f.op == "atom" && len(f.terms) == 0 {
		for _, w := range strings.Fields(f.symbol) {
			if w == word {
				return true
			}
		}
	}
	for _, c := range f.children {
		if folOpaqueWord(c, word) {
			return true
		}
	}
	return false
}

// Quoted surface names cannot enter the internal bound-variable namespace.
func folLiteralConstant(s string) string {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "$") || strings.HasPrefix(s, "\x00") {
		return "\x00" + s
	}
	return s
}
func folParseName(s string) (string, bool) {
	if m := folNameEscape.FindStringSubmatch(s); m != nil {
		return m[1], true
	}
	if !strings.HasPrefix(s, "name{") || !strings.HasSuffix(s, "}") {
		return "", false
	}
	var name string
	if json.Unmarshal([]byte(s[5:len(s)-1]), &name) != nil || name == "" {
		return "", false
	}
	return folLiteralConstant(name), true
}

var folJSONString = regexp.MustCompile(`"(?:\\.|[^"\\])*"`)

// Fresh-name checks must inspect decoded quoted names, including JSON escapes.
func folDecodedLiteralText(s string) string {
	var out strings.Builder
	for _, span := range folLiteralSpans(s) {
		for _, quoted := range folJSONString.FindAllString(s[span[0]:span[1]], -1) {
			var value string
			if json.Unmarshal([]byte(quoted), &value) == nil {
				out.WriteString(value)
				out.WriteByte(' ')
			}
		}
	}
	return out.String()
}

func folSymbolContains(f *folFormula, text string) bool {
	if strings.Contains(f.symbol, text) {
		return true
	}
	for _, c := range f.children {
		if folSymbolContains(c, text) {
			return true
		}
	}
	return false
}

func (policy strictRules) folLiteralHasName(s, name string) bool {
	if len(folLiteralSpans(s)) == 0 && !strings.HasPrefix(name, "\x00") {
		return false
	}
	readings, _ := policy.legacyFOLReadings(s)
	for _, f := range readings {
		for _, n := range folNames(f) {
			if n == name {
				return true
			}
		}
	}
	return false
}
