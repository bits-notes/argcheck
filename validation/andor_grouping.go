package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// connectiveBoundary records only connectives outside parentheses and literal
// escapes. Offsets refer to the original text, so diagnostics retain spelling.
type connectiveBoundary struct {
	start, end int
	op         string
}

func groupingBoundaries(s string) ([]connectiveBoundary, []int) {
	syntax := strings.ToLower(groupingSyntax(s))
	depth := 0
	var bs []connectiveBoundary
	var commas []int
	for i := 0; i < len(syntax); i++ {
		switch syntax[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				commas = append(commas, i)
			}
		}
		if depth != 0 {
			continue
		}
		for _, op := range []string{"and", "or", "unless"} {
			sep := " " + op + " "
			if strings.HasPrefix(syntax[i:], sep) {
				bs = append(bs, connectiveBoundary{i, i + len(sep), op})
			}
		}
	}
	return bs, commas
}

// groupingPlan fixes the main connective only when new grouping syntax is
// present. Legacy single-connective lists keep their existing parsers.
func groupingPlan(s string) (op, left, right string, managed bool, reason string) {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	if !mixedGroupingWords(lower) && !strings.ContainsAny(s, "()") {
		return "", "", "", false, ""
	}
	if strings.ContainsAny(s, "()") && groupingParenthesisPlacement(s) != "" {
		// Existing parenthetical annotations inside an atomic clause are
		// not clause-grouping syntax. Preserve their legacy reading.
		return "", "", "", false, ""
	}
	// These established multiword operators own their connective tokens.
	_, _, topIff := folSplit(lower, " if and only if ")
	if topIff || strings.Contains(lower, "it is the case both that") {
		return "", "", "", false, ""
	}
	prefix := ""
	if strings.HasPrefix(lower, "either ") {
		prefix = "either "
		op = "or"
	}
	if strings.HasPrefix(lower, "both ") {
		prefix = "both "
		op = "and"
	}
	body := s[len(prefix):]
	bs, commas := groupingBoundaries(body)
	and, or, unless := false, false, false
	for _, b := range bs {
		and = and || b.op == "and"
		or = or || b.op == "or"
		unless = unless || b.op == "unless"
	}
	mixed := (and && or) || (unless && (and || or))
	managed = mixed || strings.ContainsAny(groupingSyntax(body), "()")
	if !managed {
		return "", "", "", false, ""
	}
	if mixed && len(commas) > 1 {
		return "", "", "", true, "ambiguous comma grouping: multiple commas mix list punctuation with connective boundaries"
	}
	selected := -1
	if prefix != "" {
		for i, b := range bs {
			if b.op == op {
				selected = i
				break
			}
		}
	} else if mixed {
		if unless {
			return "", "", "", true, "mixes top-level unless with and/or without explicit grouping"
		}
		return "", "", "", true, "mixes top-level and/or without explicit grouping"
	} else if len(bs) > 0 {
		selected = 0
	}
	if selected < 0 {
		return "", "", "", managed, ""
	}
	b := bs[selected]
	return b.op, strings.TrimSpace(strings.TrimSuffix(body[:b.start], ",")), strings.TrimSpace(body[b.end:]), true, ""
}

// Opaque complements and relative clauses have no asserted connective tree.
// Conservatively leave the entire containing scope opaque, just as the gate's
// lexical recognizers do. Explicit binders and logical wrappers are visited first.
func groupingOpaque(s string) bool {
	lower := strings.ToLower(s)
	return (strings.Contains(lower, " either ") && !strings.HasPrefix(lower, "either ")) || opaqueContextWords.MatchString(s) || strings.HasPrefix(lower, "unless ") || strings.Contains(lower, " who ") || strings.Contains(lower, " which ") || strings.ContainsAny(s, "\"`“”")
}

func groupingProblem(s string) string {
	if groupingParenthesisPlacement(s) != "" {
		return ""
	}
	s = folStrip(strings.TrimSpace(s))
	lower := strings.ToLower(s)
	if m := explicitBinder.FindStringSubmatchIndex(lower); m != nil && m[0] == 0 {
		return groupingProblem(s[m[1]:])
	}
	for _, prefix := range []string{"it is not the case that ", "it is the case that ", "necessarily, ", "necessarily ", "possibly, ", "possibly "} {
		if strings.HasPrefix(lower, prefix) {
			return groupingProblem(s[len(prefix):])
		}
	}
	if strings.HasPrefix(lower, "if ") || strings.HasPrefix(lower, "when ") || strings.HasPrefix(lower, "whenever ") {
		start := strings.Index(s, " ") + 1
		for _, sep := range []string{", then ", " then ", ", "} {
			if a, b, ok := folSplit(s[start:], sep); ok {
				if why := groupingProblem(a); why != "" {
					return why
				}
				return groupingProblem(b)
			}
		}
	}
	if a, b, ok := folSplit(s, " if and only if "); ok {
		if why := groupingProblem(a); why != "" {
			return why
		}
		return groupingProblem(b)
	}
	if groupingOpaque(s) {
		return ""
	}
	_, a, b, managed, why := groupingPlan(s)
	if why != "" && !strings.ContainsAny(groupingSyntax(s), "()") {
		// Phrase-level coordination ("a bad or catastrophic outcome",
		// "fumbling or bumbling") is not a clause connective. Judge it on the
		// original spelling, where names are still capitalized. The gate keeps
		// its legacy reading for such sentences (gateGroupingPlan).
		bs, _ := groupingBoundaries(s)
		kept, _ := clauseLevelBoundaries(s, bs, nil)
		and, or, unless := false, false, false
		for _, b := range kept {
			and = and || b.op == "and"
			or = or || b.op == "or"
			unless = unless || b.op == "unless"
		}
		if !(and && or) && !(unless && (and || or)) {
			return ""
		}
	}
	if why != "" {
		// A bare coordinated name/predicate is not a pair of whole clauses.
		// Keep such unregimented English atomic instead of inventing scope.
		if !strings.Contains(s, ",") {
			bs, _ := groupingBoundaries(s)
			start := 0
			for _, boundary := range bs {
				if groupingBareFragment(s[start:boundary.start]) {
					return ""
				}
				start = boundary.end
			}
			if groupingBareFragment(s[start:]) {
				return ""
			}
		}
		return why
	}
	if managed && a != "" && b != "" {
		if why := groupingProblem(a); why != "" {
			return why
		}
		return groupingProblem(b)
	}
	// Visit parenthesized scopes even when there is no outer connective.
	syntax := groupingSyntax(s)
	depth, start := 0, 0
	for i, c := range syntax {
		if c == '(' {
			if depth == 0 {
				start = i + 1
			}
			depth++
		}
		if c == ')' {
			depth--
			if depth == 0 {
				if why := groupingProblem(s[start:i]); why != "" {
					return why
				}
			}
		}
	}
	return ""
}

func validateStatementGrouping(s string, line int) Result {
	return allStrictRules().validateStatementGrouping(s, line)
}

func (policy strictRules) validateStatementGrouping(s string, line int) Result {
	if policy.operatorCoordinationAmbiguous(s) {
		return Result{Passed: false, Message: fmt.Sprintf("line %d sentence %q: %s", line, s+".", operatorGroupingMessage), Details: map[string]any{"line": line, "sentence": s}}
	}
	lower := strings.ToLower(s)
	if !strings.ContainsAny(s, "()") && !mixedGroupingWords(lower) {
		return Result{Passed: true}
	}
	depth := 0
	why := ""
	for _, c := range groupingSyntax(s) {
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
		}
		if depth < 0 {
			why = "unbalanced parentheses: unexpected closing parenthesis"
			break
		}
	}
	if why == "" && depth != 0 {
		why = "unbalanced parentheses: missing closing parenthesis"
	}
	if why == "" {
		if strings.Contains(s, "()") {
			why = "parentheses must contain a clause"
		}
	}
	if why == "" {
		why = groupingProblem(s)
	}
	if why == "" {
		return Result{Passed: true}
	}
	return Result{Passed: false, Message: fmt.Sprintf("line %d sentence %q: %s; use either...or / both...and, or parentheses to group", line, s+".", why), Details: map[string]any{"line": line, "sentence": s}}
}

var clauseStartWords = map[string]bool{
	"the": true, "a": true, "an": true, "every": true, "each": true, "all": true, "some": true,
	"no": true, "any": true, "most": true, "many": true, "few": true, "both": true, "either": true,
	"neither": true, "it": true, "there": true, "he": true, "she": true, "they": true, "we": true,
	"i": true, "you": true, "this": true, "that": true, "these": true, "those": true, "his": true,
	"her": true, "their": true, "its": true, "our": true, "my": true, "your": true, "nothing": true,
	"nobody": true, "no one": true, "everyone": true, "everybody": true, "everything": true,
	"someone": true, "somebody": true, "something": true, "anyone": true, "anything": true,
	"if": true, "not": true, "necessarily": true, "possibly": true,
}

// clauseOperandStart reports whether text after a connective opens a clause:
// a capitalized word or single-letter term (names, schematic letters,
// variables), a pronoun, a determiner or quantifier, a parenthesis, or a
// logical operator. Lower-case content words ("a bad or catastrophic
// outcome", "fumbling or bumbling", "approaching or exceeding") continue a
// noun or verb phrase, so that connective is not a clause boundary.
func clauseOperandStart(rest string) bool {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return false
	}
	if rest[0] == '(' {
		return true
	}
	word := strings.Fields(rest)[0]
	word = strings.TrimRight(word, ",;:")
	if word == "" {
		return false
	}
	if c := word[0]; c >= 'A' && c <= 'Z' {
		return true
	}
	lower := strings.ToLower(word)
	if determinerWords[lower] {
		// "or a correct parametric model" continues a noun phrase; "or the
		// dog barks" opens a clause. Look for a verb cue before the next
		// connective or punctuation.
		return operandHasVerbCue(rest)
	}
	if len(word) == 1 {
		return true
	}
	return clauseStartWords[lower]
}

var determinerWords = map[string]bool{
	"a": true, "an": true, "the": true, "this": true, "that": true, "these": true, "those": true,
	"his": true, "her": true, "their": true, "its": true, "our": true, "my": true, "your": true,
}

var auxiliaryWords = map[string]bool{
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "has": true,
	"have": true, "had": true, "does": true, "do": true, "did": true, "will": true, "would": true,
	"can": true, "could": true, "should": true, "must": true, "may": true, "might": true, "shall": true,
	"cannot": true,
}

func operandHasVerbCue(rest string) bool {
	words := strings.Fields(strings.ToLower(rest))
	for i, w := range words {
		if i > 0 && (w == "and" || w == "or") {
			return false
		}
		bare := strings.TrimRight(w, ",;:.")
		if auxiliaryWords[bare] || (i >= 2 && (strings.HasSuffix(bare, "s") || strings.HasSuffix(bare, "ed"))) {
			return true
		}
		if bare != w {
			return false
		}
	}
	return false
}

// clauseLevelBoundaries keeps the connectives that join whole clauses, and
// the commas that sit before such a connective or before a clause start.
func clauseLevelBoundaries(s string, bs []connectiveBoundary, commas []int) ([]connectiveBoundary, []int) {
	var keep []connectiveBoundary
	for _, b := range bs {
		if clauseOperandStart(s[b.end:]) {
			keep = append(keep, b)
		}
	}
	var keptCommas []int
	for _, c := range commas {
		before := false
		for _, b := range keep {
			if b.start == c+1 {
				before = true
			}
		}
		if before || clauseOperandStart(s[c+1:]) {
			keptCommas = append(keptCommas, c)
		}
	}
	return keep, keptCommas
}

// gateGroupingPlan is groupingPlan as the entailment gate sees it. A sentence
// with a grouping problem reaches the gate only through a syntax exemption
// (opaque text, bare noun-phrase coordination), so the gate keeps the legacy
// reading for it instead of inventing a new scope.
func gateGroupingPlan(s string) (op, left, right string, managed bool, reason string) {
	op, left, right, managed, reason = groupingPlan(s)
	if reason != "" {
		return "", "", "", false, ""
	}
	return
}

func groupedCandidates(s, want string) ([]clausePair, bool) {
	op, a, b, managed, why := gateGroupingPlan(s)
	if !managed {
		return nil, false
	}
	if hasOuterGateOperator(strings.ToLower(s)) || why != "" || op != want || a == "" || b == "" {
		return nil, true
	}
	pairs := []clausePair{{left: a, right: b}}
	// Same-connective grouping is associative: (A and B) and C is also
	// A and (B and C). Offer those re-bracketings so a grouped operand that is
	// itself a conjunction (or disjunction) still splits at every boundary.
	sub := conjunctionCandidates
	if s != strings.ToLower(s) {
		sub = conjunctionCandidatesPreservingCase
	}
	if op == "or" {
		sub = disjunctionCandidates
	}
	for _, p := range sub(a) {
		pairs = append(pairs, clausePair{left: p.left, right: p.right + " " + op + " " + b})
	}
	for _, p := range sub(b) {
		pairs = append(pairs, clausePair{left: a + " " + op + " " + p.left, right: p.right})
	}
	return pairs, true
}

// Keep grouping visible in mismatch excerpts even though identity comparison
// can remove a redundant pair of outer parentheses.
func diagnosticClause(s string) string {
	if strings.ContainsAny(groupingSyntax(s), "()") {
		return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(s), "."))
	}
	return normalizeClause(s)
}

func groupingBareFragment(s string) bool {
	words := strings.Fields(strings.TrimSpace(s))
	return len(words) < 2 && !(len(words) == 1 && len(words[0]) == 1)
}

// Parentheses delimit whole operands, never a name, verb, or noun phrase
// within an atomic clause. Literal escapes are masked by folSyntax.
func groupingParenthesisPlacement(s string) string {
	syntax := strings.ToLower(groupingSyntax(s))
	for i, c := range syntax {
		if c == '(' {
			before := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(syntax[:i]), ","))
			allowed := before == "" || strings.HasSuffix(before, "(")
			for _, suffix := range []string{" and", " or", "either", "both", "if", "then", "that", "not", "necessarily", "possibly"} {
				allowed = allowed || before == strings.TrimSpace(suffix) || strings.HasSuffix(before, suffix)
			}
			if !allowed {
				return "parentheses must group whole clauses, not words inside a clause"
			}
			if i+1 < len(syntax) && strings.HasPrefix(strings.TrimSpace(syntax[i+1:]), ")") {
				return "parentheses must contain a clause"
			}
		}
		if c == ')' {
			after := strings.TrimSpace(syntax[i+1:])
			allowed := after == "" || strings.HasPrefix(after, ")")
			after = strings.TrimSpace(strings.TrimPrefix(after, ","))
			for _, prefix := range []string{"and ", "or ", "then ", "if ", "unless "} {
				allowed = allowed || strings.HasPrefix(after, prefix)
			}
			if !allowed {
				return "parentheses must group whole clauses, not words inside a clause"
			}
		}
	}
	return ""
}

// Enumerate only visible disjunction boundaries; an or inside a parenthesized
// conjunction must never become a separate dilemma case.
func groupedDisjunctionPartitions(s string, count int) [][]string {
	if count < 1 {
		return nil
	}
	if count == 1 {
		return [][]string{{s}}
	}
	var result [][]string
	for _, pair := range disjunctionCandidates(s) {
		for n := 1; n < count; n++ {
			for _, a := range groupedDisjunctionPartitions(pair.left, n) {
				for _, b := range groupedDisjunctionPartitions(pair.right, count-n) {
					result = append(result, append(append([]string(nil), a...), b...))
				}
			}
		}
	}
	return result
}

// Parenthetical identifiers in established prose are labels, not clause
// grouping. Preserve their historical atomic behavior (definition (a), etc.).
var groupingLabel = regexp.MustCompile(`(?i)\b(?:definition|condition|proposition) \([a-z]\)`)

func groupingSyntax(s string) string {
	syntax := folSyntax(s)
	if !strings.Contains(syntax, "(") {
		return syntax
	}
	return groupingLabel.ReplaceAllStringFunc(syntax, func(label string) string {
		return strings.NewReplacer("(", " ", ")", " ").Replace(label)
	})
}

func stripGroupedClause(s string) string {
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return s
	}
	return folStrip(s)
}

func mixedGroupingWords(s string) bool {
	a, o, u := strings.Contains(s, " and "), strings.Contains(s, " or "), strings.Contains(s, " unless ")
	return a && o || u && (a || o)
}
