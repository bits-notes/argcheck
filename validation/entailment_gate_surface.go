package validation

import (
	"regexp"
	"strings"
)

var gateMarkerLine = regexp.MustCompile(`(?m)^([1-9][0-9]*\. )(?i:so|thus|hence|therefore),? (.+\. \[[1-9][0-9]*(?:,[1-9][0-9]*)*\])$`)

// In the numbered proof register, a leading inference marker on a cited step
// announces the inference; it contributes nothing to the asserted proposition.
// Keep this adaptation local to the gate, not the strict syntax validator.
func normalizeGateMarkers(output string) string {
	output = strings.ReplaceAll(strings.ReplaceAll(output, "\r\n", "\n"), "\r", "\n")
	return gateMarkerLine.ReplaceAllStringFunc(output, func(line string) string {
		match := gateMarkerLine.FindStringSubmatch(line)
		lower := strings.ToLower(match[2])
		// Degree/manner phrases are content, not inference markers.
		for _, prefix := range []string{"many ", "much ", "few ", "little ", "far ", "long ", "that ", "called "} {
			if strings.HasPrefix(lower, prefix) {
				return line
			}
		}
		return match[1] + "Therefore, " + match[2]
	})
}

var gateContraction = regexp.MustCompile(`(?i)\b(?:doesn't|don't|didn't|isn't|aren't|wasn't|weren't|hasn't|haven't|hadn't|won't|wouldn't|can't|couldn't|shouldn't|mustn't|needn't|cannot)\b`)

// These are spelling expansions of the same auxiliary and negation, not
// changes of scope or verb inflection. In particular, cannot = can not.
func expandGateContractions(value string) string {
	// Quoted spellings are mentioned rather than asserted; leave them opaque.
	if strings.ContainsAny(value, "\"`“”‘") || strings.Contains(value, " '") || strings.HasPrefix(value, "'") {
		return value
	}
	value = strings.ReplaceAll(value, "’", "'")
	return gateContraction.ReplaceAllStringFunc(value, func(word string) string {
		lower := strings.ToLower(word)
		expanded := ""
		switch lower {
		case "won't":
			expanded = "will not"
		case "can't", "cannot":
			expanded = "can not"
		default:
			expanded = strings.TrimSuffix(lower, "n't") + " not"
		}
		if word[0] >= 'A' && word[0] <= 'Z' {
			expanded = strings.ToUpper(expanded[:1]) + expanded[1:]
		}
		return expanded
	})
}

// A deliberately small grammar makes infix if a clause connective rather than
// the complement of "asks/wonders/knows if", a quotation, or a relative clause.
var gateSimpleClause = regexp.MustCompile(`(?i)^(?:[a-z]|(?:[a-z]+|the [a-z]+(?: [a-z]+)?) (?:(?:is|are|was|were) (?:not )?(?:lit|up|down|open|closed|wet|dry|ready|red|green|hot|cold|active|inactive)|(?:rises|rise|falls|fall|shines|shone|opens|open|closes|close|rings|ring|rains|snows|sounds|starts|stops|succeeds|fails|works|sleeps|smiles|walks|sings|flies|barks|purrs)))$`)
var gateUnsafeSurface = regexp.MustCompile(`(?i)\b(?:if|when|whenever|unless|since|because|after|before|that|who|which|whether|and|or|only|then)\b|["';:?!().,“”‘’]`)

func simpleGateClause(value string) bool {
	return !gateUnsafeSurface.MatchString(value) && gateSimpleClause.MatchString(value)
}

// In the gate's conditional register, the comma in "If/When/Whenever A, B"
// marks the same antecedent/consequent boundary as ", then". Only one comma
// and no nested connective are accepted, so no alternative scope is invented.
func normalizeGateConditional(value string) string {
	lower := strings.ToLower(value)
	for _, prefix := range []string{"if ", "when ", "whenever "} {
		if strings.HasPrefix(lower, prefix) && strings.Count(value, ",") == 1 {
			parts := strings.SplitN(value[len(prefix):], ",", 2)
			left, right := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if strings.HasPrefix(strings.ToLower(right), "then ") {
				right = right[len("then "):]
			}
			if left != "" && right != "" && !gateUnsafeSurface.MatchString(left) && !gateUnsafeSurface.MatchString(right) {
				return "If " + left + ", then " + right
			}
		}
	}
	// Do not rewrite "only if" yet: the v6 regression contract explicitly
	// rejects its forward use, and v9 must preserve those negative verdicts.
	// "B if A" states exactly A -> B when both operands are independent,
	// simple declarative clauses; embedded/interrogative uses remain opaque.
	if strings.Count(lower, " if ") == 1 {
		index := strings.Index(lower, " if ")
		left, right := value[:index], value[index+4:]
		if simpleGateClause(left) && simpleGateClause(right) {
			return "If " + right + ", then " + left
		}
	}
	return value
}

var gateElidedCopula = regexp.MustCompile(`(?i)^([a-z]+|the [a-z]+(?: [a-z]+)?) (is|are|was|were) (.+)$`)
var gateBarePredicate = regexp.MustCompile(`(?i)^(?:[a-z]+|(?:on|in|at) (?:the )?[a-z]+)$`)

// With a single unquantified subject, positive copula and two bare predicates,
// coordination elides only the repeated subject/copula: S is P or S is Q.
// Negation, quantifiers, embedded clauses and multiple subjects are excluded.
func expandGateElidedDisjunction(statement string) string {
	if !strings.HasPrefix(strings.ToLower(statement), "either ") || strings.Count(strings.ToLower(statement), " or ") != 1 {
		return statement
	}
	index := strings.Index(strings.ToLower(statement), " or ")
	left, right := statement[len("either "):index], statement[index+4:]
	match := gateElidedCopula.FindStringSubmatch(left)
	if match == nil || !gateBarePredicate.MatchString(match[3]) || !gateBarePredicate.MatchString(right) {
		return statement
	}
	for _, part := range []string{match[1], match[3], right} {
		if gateUnsafeSurface.MatchString(part) {
			return statement
		}
		for _, word := range strings.Fields(strings.ToLower(part)) {
			switch word {
			case "not", "no", "every", "all", "some", "something", "someone", "everyone", "everything", "nothing", "nobody", "neither", "either", "there", "a", "an", "any", "each", "many", "most", "few", "several", "both", "none", "everybody", "somebody", "anybody", "anyone", "anything", "whoever":
				return statement
			}
		}
	}
	return statement[:index+4] + match[1] + " " + match[2] + " " + right
}
