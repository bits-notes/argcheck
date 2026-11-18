package validation

import (
	"regexp"
	"strings"
)

// "it's (not) the case that" is the same wrapper as "it is (not) the case
// that". Other uses of "it's" ("it's raining") are left alone.
var gateCaseContraction = regexp.MustCompile(`(?i)\bit's (not )?the case\b`)

var gateContraction = regexp.MustCompile(`(?i)\b(?:doesn't|don't|didn't|isn't|aren't|wasn't|weren't|hasn't|haven't|hadn't|won't|wouldn't|can't|couldn't|shouldn't|mustn't|needn't|cannot)\b`)

// These are spelling expansions of the same auxiliary and negation, not
// changes of scope or verb inflection. In particular, cannot = can not.
func expandGateContractions(value string) string {
	// Quoted spellings are mentioned rather than asserted; leave them opaque.
	if strings.ContainsAny(value, "\"`“”‘") || strings.Contains(value, " '") || strings.HasPrefix(value, "'") {
		return value
	}
	value = strings.ReplaceAll(value, "’", "'")
	value = gateCaseContraction.ReplaceAllStringFunc(value, func(phrase string) string {
		expanded := "it is " + strings.ToLower(phrase[len("it's "):])
		if phrase[0] == 'I' {
			expanded = "I" + expanded[1:]
		}
		return expanded
	})
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

// v31.2: "Y if X." (or "Y, if X.") states exactly "If X, then Y." This
// rewrite is applied to whole step statements only, and only on a second pass
// after the argument as written fails (see checkEntailmentGate), so it can
// add verdicts but never remove one. It applies only when the reading is
// unambiguous:
//   - exactly one "if", not "only if", "even if", "as if", "if and only if",
//     and no "unless" anywhere;
//   - not an embedded question ("asked if", "wonder if", "curious if", ...);
//   - the consequent is not a report or attitude ("says", "thinks", "that"),
//     which could scope over the "if";
//   - the consequent has no and/or/either/comma (scope would be ambiguous);
//   - the antecedent has no further if/then/or (a conjunctive "X and Z" is fine).
var postposedIfBlockedBefore = map[string]bool{
	"only": true, "even": true, "as": true, "and": true, "or": true, "but": true,
	"ask": true, "asks": true, "asked": true, "asking": true,
	"wonder": true, "wonders": true, "wondered": true, "wondering": true,
	"know": true, "knows": true, "knew": true, "see": true, "sees": true, "saw": true,
	"check": true, "checks": true, "checked": true, "decide": true, "decides": true, "decided": true,
	"tell": true, "tells": true, "told": true, "doubt": true, "doubts": true, "doubted": true,
	"unsure": true, "sure": true, "whether": true, "question": true, "determine": true, "determines": true,
	"determined": true, "test": true, "tests": true, "tested": true,
	"curious": true, "uncertain": true, "unclear": true, "clear": true, "matter": true, "matters": true,
	"inquire": true, "inquires": true, "inquired": true, "learn": true, "learns": true, "learned": true,
	"find": true, "finds": true, "found": true, "remember": true, "remembers": true, "remembered": true,
	"forget": true, "forgets": true, "forgot": true, "care": true, "cares": true, "cared": true,
}

var postposedConsequentUnsafe = regexp.MustCompile(`(?i)\b(?:and|or|either|if|then|unless|whether|that|says?|said|claims?|claimed|thinks?|thought|believes?|believed|reports?|reported|argues?|argued|states|stated|suggests?|suggested|insists?|insisted|denies|deny|denied|hopes?|hoped|fears?|feared|expects?|expected|predicts?|predicted|shows?|showed|shown|proves?|proved|according)\b|[,;:?!"()“”]`)
var postposedAntecedentUnsafe = regexp.MustCompile(`(?i)\b(?:if|then|or|unless|whether)\b|[;:?!"()“”]`)

func postposedConditional(value string) (string, bool) {
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "if ") || strings.Contains(lower, "if and only if") || strings.Contains(lower, "unless") {
		return "", false
	}
	if strings.Count(lower, " if ") != 1 {
		return "", false
	}
	index := strings.Index(lower, " if ")
	left := strings.TrimSpace(value[:index])
	right := strings.TrimSpace(value[index+len(" if "):])
	left = strings.TrimSpace(strings.TrimSuffix(left, ","))
	if left == "" || right == "" {
		return "", false
	}
	words := strings.Fields(strings.ToLower(left))
	if postposedIfBlockedBefore[strings.Trim(words[len(words)-1], ",")] {
		return "", false
	}
	if postposedConsequentUnsafe.MatchString(left) || postposedAntecedentUnsafe.MatchString(right) {
		return "", false
	}
	return "If " + right + ", then " + left, true
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
