package validation

import (
	"strings"
	"unicode"
)

// A letter following a capitalized name component is literal. Logical
// scaffolding is not a name component (If X, Both X, and so on).
// nameLetterJoiner marks a schema letter that belongs to a proper name
// ("Plan X", "Malcolm X", "Generation X"). protectNameLetters inserts it once
// per whole step statement, so matchers that see only a clause fragment still
// know the letter is literal.
const nameLetterJoiner = '\u2060'

func schemaLetterInName(rs []rune, index int) bool {
	return index > 0 && rs[index-1] == nameLetterJoiner
}

// protectNameLetters marks X/Y/Z as literal when every standalone occurrence
// of that letter in the statement follows the same capitalized word. "If X are
// demands and Answering X is ..." or "under SPC X knows" keep X a variable,
// because X also occurs on its own.
func protectNameLetters(statement string) string {
	rs := []rune(statement)
	literal := map[rune]bool{}
	for _, letter := range []rune{'X', 'Y', 'Z'} {
		name, seen, ok := "", false, true
		for i, r := range rs {
			if r != letter || (i > 0 && unicode.IsLetter(rs[i-1])) || (i < len(rs)-1 && unicode.IsLetter(rs[i+1])) {
				continue
			}
			before := nameBeforeLetter(rs, i)
			if before == "" || (seen && before != name) {
				ok = false
				break
			}
			name, seen = before, true
		}
		if seen && ok {
			literal[letter] = true
		}
	}
	if len(literal) == 0 {
		return statement
	}
	var out []rune
	for i, r := range rs {
		if literal[r] && (i == 0 || !unicode.IsLetter(rs[i-1])) && (i == len(rs)-1 || !unicode.IsLetter(rs[i+1])) {
			for len(out) > 0 && unicode.IsSpace(out[len(out)-1]) {
				out = out[:len(out)-1]
			}
			out = append(out, ' ', nameLetterJoiner)
		}
		out = append(out, r)
	}
	return string(out)
}

func protectArgumentNameLetters(argument NumberedArgument) NumberedArgument {
	steps := make([]NumberedArgumentStep, len(argument.Steps))
	for i, step := range argument.Steps {
		step.Statement = protectNameLetters(step.Statement)
		steps[i] = step
	}
	argument.Steps = steps
	return argument
}

func nameBeforeLetter(rs []rune, index int) string {
	prefix := strings.TrimRightFunc(string(rs[:index]), unicode.IsSpace)
	words := strings.Fields(prefix)
	if len(words) == 0 {
		return ""
	}
	prev := words[len(words)-1]
	switch strings.ToLower(prev) {
	case "if", "then", "both", "either", "and", "or", "not", "that", "of", "for", "when", "whenever":
		return ""
	}
	r := []rune(prev)
	if len(r) > 1 && unicode.IsUpper(r[0]) && unicode.IsLetter(r[len(r)-1]) {
		return prev
	}
	return ""
}

func round2Singular(word string) string {
	switch word {
	case "princess", "corps", "class", "bus", "glass", "chaos", "lens", "series", "species":
		return word
	case "princes":
		return "prince"
	case "corpses":
		return "corpse"
	case "princesses":
		return "princess"
	case "classes":
		return "class"
	case "buses":
		return "bus"
	case "glasses":
		return "glass"
	case "lenses":
		return "lens"
	case "analyses":
		return "analysis"
	case "theses":
		return "thesis"
	}
	if strings.HasSuffix(word, "ss") || strings.HasSuffix(word, "us") || strings.HasSuffix(word, "is") {
		return word
	}
	if strings.HasSuffix(word, "ches") || strings.HasSuffix(word, "shes") || strings.HasSuffix(word, "xes") || strings.HasSuffix(word, "sses") {
		return strings.TrimSuffix(word, "es")
	}
	if strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

// Preserve spelling for the positive term register. Quantifier phrases are
// never names merely because their first word starts a sentence.
func instanceSubjectAllowed(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	// Single-letter constants include Fitch arbitrary names such as a.
	if len(s) == 1 && s[0] >= 'a' && s[0] <= 'z' {
		return true
	}
	if !schematicBindingAllowed(lower) || quantifiedInstanceSubject(lower) {
		return false
	}
	for _, p := range []string{"the ", "a ", "an "} {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return false
	}
	// Only the first word decides: a quantifier or determiner other than
	// the/a/an is not a term ("almost no person", "just one senator", "one
	// of the planets", "someone in Athens"). Names, possessives, gerund
	// phrases ("trying to remove all assumptions") and bare noun phrases
	// ("consciousness", "kalbfleisch's distinction") are terms.
	first := strings.ToLower(strings.Trim(words[0], ",;:"))
	if nonReferringTermWords[first] || instanceQuantifierHeads[first] {
		return false
	}
	for _, phrase := range instanceQuantifierPhrases {
		if strings.HasPrefix(lower+" ", phrase) {
			return false
		}
	}
	return true
}

var instanceQuantifierHeads = map[string]bool{
	"almost": true, "nearly": true, "just": true, "very": true, "exactly": true, "barely": true,
	"hardly": true, "scarcely": true, "fewer": true, "less": true, "more": true, "at": true,
	"two": true, "three": true, "four": true, "five": true, "several": true, "numerous": true,
	"countless": true, "various": true, "certain": true,
}

var instanceQuantifierPhrases = []string{"one of ", "some of ", "many of ", "most of ", "all but "}

func uiInstanceCandidates(statement string) []clausePair {
	if hasOuterGateOperator(normalizeClause(statement)) {
		return nil
	}
	var result []clausePair
	for _, sep := range []string{" is ", " am ", " are "} {
		for _, pair := range appendSplitCandidatesPreservingCase(nil, statement, sep, false) {
			// The instance must be the premise's own top-level copula: the
			// subject cannot carry a sentential connective ("Alice won the
			// lottery if and only if Alice").
			if subjectHasConnective(pair.left) {
				continue
			}
			if instanceSubjectAllowed(pair.left) {
				result = append(result, clausePair{normalizeClause(pair.left), normalizeClause(pair.right)})
			}
		}
	}
	return result
}

func bareNegativeSubjectAllowed(statement string) bool {
	for _, sep := range []string{" is not ", " are not ", " am not "} {
		if a, _, ok := strings.Cut(statement, sep); ok {
			if instanceSubjectAllowed(a) {
				return true
			}
			// Bare plural class generics (including modified classes such as
			// "conscious experiences") remain supported, unlike quantified NPs.
			words := strings.Fields(strings.ToLower(a))
			if len(words) == 0 || !schematicBindingAllowed(a) {
				return false
			}
			for _, word := range words {
				if nonReferringTermWords[word] {
					return false
				}
			}
			return strings.HasSuffix(words[len(words)-1], "s")
		}
	}
	return false
}

// Diagnostic only: called after all certification routes failed. It must never
// veto a valid alternative proof using the same premises.
func accuracyRound2Failure(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) (string, string) {
	for _, p := range premises {
		s := normalizeClause(p.Statement)
		if identityOpaqueExtra.MatchString(s) && len(premises) > 1 {
			for _, q := range premises {
				if len(identityCandidates(q.Statement)) > 0 {
					if reason := identitySubstitutionUnsafe(stepStatements(premises), conclusion.Statement); reason != "" {
						return "identity substitution", reason
					}
					return "identity substitution", "identity substitution is blocked in an opaque intensional context"
				}
			}
		}
		if len(premises) == 2 {
			for _, q := range premises {
				if q.Number != p.Number && matchSchematicInstantiationAlphabet([]NumberedArgumentStep{p, q}, conclusion, nameBindingTemplate).matched {
					return "schematic instantiation", "schema letters inside a proper name are literal, not variables"
				}
			}
		}
		for _, prefix := range []string{"almost no ", "very few ", "just one ", "only one ", "one of ", "someone ", "something ", "somebody ", "most ", "few "} {
			if strings.HasPrefix(s, prefix) {
				return "term instantiation", "quantifier phrase cannot be used as an individual term or a universal class: " + prefix
			}
		}
		if len(universalCandidates(p.Statement)) > 0 {
			for _, q := range premises {
				if q.Number != p.Number && hasOuterGateOperator(normalizeClause(q.Statement)) {
					return "universal instantiation", "universal instantiation requires a whole top-level copula instance; it cannot rewrite inside a larger sentence"
				}
			}
		}
	}
	return "", ""
}

// Diagnostic counterfactual only: restore the old treatment of name suffixes.
// A successful match here (after normal certification failed) localizes the
// failure to the forbidden name binding rather than an unrelated mismatch.
func nameBindingTemplate(s string) (string, schematicVariableCounts) {
	rs := []rune(s)
	for i, r := range rs {
		if marker, ok := schematicVariableMarkers[r]; ok && (i == 0 || !unicode.IsLetter(rs[i-1])) && (i == len(rs)-1 || !unicode.IsLetter(rs[i+1])) && schemaLetterInName(rs, i) {
			rs[i] = marker
		}
	}
	// Counts must include the name markers as well as the real variables.
	template, counts := extendedSchematicTemplate(string(rs))
	for _, r := range rs {
		if variable, ok := schematicVariableForMarker(r); ok {
			counts[variable]++
		}
	}
	return template, counts
}

// Upward-monotone quantified subjects are lifted through a class inclusion,
// not treated as individual UI witnesses. Keep this in the schematic family.
func matchMonotoneUniversal(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	if len(universalCandidates(p[0].Statement)) == 0 {
		return schemaMatchOutcome{}
	}
	for _, statement := range []string{p[1].Statement, c.Statement} {
		if hasOuterGateOperator(normalizeClause(statement)) {
			return schemaMatchOutcome{}
		}
		bs, _ := groupingBoundaries(statement)
		bs, _ = clauseLevelBoundaries(statement, bs, nil)
		if len(bs) > 0 {
			return schemaMatchOutcome{}
		}
	}
	for _, instance := range instanceCandidates(p[1].Statement) {
		if !strings.HasPrefix(instance.left, "most ") && !strings.HasPrefix(instance.left, "more than ") {
			continue
		}
		if !schematicBindingAllowed(instance.left) || len(instanceCandidates(instance.left)) > 0 {
			continue
		}
		// Replacing only the whole subject preserves the quantifier; a compound
		// or embedded sentence cannot become a top-level witness instance this way.
		for _, result := range instanceCandidates(c.Statement) {
			if instance.left != result.left {
				continue
			}
			ps := append([]NumberedArgumentStep(nil), p...)
			ps[1].Statement = "Witness" + strings.TrimPrefix(normalizeClause(p[1].Statement), instance.left)
			cs := c
			cs.Statement = "Witness" + strings.TrimPrefix(normalizeClause(c.Statement), result.left)
			if out := matchUniversalInstantiation(ps, cs); out.matched {
				return out
			}
		}
	}
	return schemaMatchOutcome{}
}

func subjectHasConnective(subject string) bool {
	lower := " " + strings.ToLower(subject) + " "
	for _, c := range []string{" if and only if ", " iff ", " only if ", " unless ", " then ", " if "} {
		if strings.Contains(lower, c) {
			return true
		}
	}
	bs, _ := groupingBoundaries(subject)
	bs, _ = clauseLevelBoundaries(subject, bs, nil)
	for _, b := range bs {
		// A clause-level and/or inside the subject means the copula belongs
		// to the last conjunct, not to the whole premise.
		if operandHasVerbCue(subject[b.end:]) || segmentHasVerb(subject[:b.start]) {
			return true
		}
	}
	return false
}

// segmentHasVerb reports a finite verb cue after the first word of a clause
// ("Bob sings", "the plant has failed").
func segmentHasVerb(seg string) bool {
	words := strings.Fields(strings.ToLower(seg))
	for i, w := range words {
		w = strings.TrimRight(w, ",;:")
		if i == 0 || strings.HasSuffix(w, "'s") || strings.HasSuffix(w, "’s") {
			continue
		}
		if auxiliaryWords[w] || ((strings.HasSuffix(w, "s") || strings.HasSuffix(w, "ed")) && !strings.HasSuffix(w, "ss")) {
			return true
		}
	}
	return false
}

// cleanNameLetterMessage removes the invisible name marker from the result
// message and the step reports (which the details share), and on a failure
// that involved a marked letter says why the letter was literal.
func cleanNameLetterMessage(report *EntailmentGateReport, result Result) Result {
	note := "; schema letters inside a proper name are literal, not variables"
	clean := func(text string, failed bool) string {
		marked := strings.Contains(text, "\\u2060") || strings.ContainsRune(text, nameLetterJoiner)
		text = strings.ReplaceAll(text, " \\u2060", " ")
		text = strings.ReplaceAll(text, "\\u2060", "")
		text = strings.ReplaceAll(text, string(nameLetterJoiner), "")
		if marked && failed && !strings.Contains(text, "proper name") {
			text += note
		}
		return text
	}
	for i := range report.Steps {
		report.Steps[i].Mismatch = clean(report.Steps[i].Mismatch, !report.Steps[i].Passed)
		report.Steps[i].AmbiguityNote = clean(report.Steps[i].AmbiguityNote, false)
	}
	result.Message = clean(result.Message, !result.Passed)
	return result
}
