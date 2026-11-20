package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Tokenize Unicode words rather than using ASCII \b: neither "theirs" nor
// "itinerary" nor a pronoun embedded in a Unicode name is a free pronoun.
var crossStatementWord = regexp.MustCompile(`[\pL\pN_]+`)

func crossStatementPronoun(word string) bool {
	switch strings.ToLower(word) {
	case "he", "she", "it", "they", "him", "her", "them", "himself", "herself", "itself", "themselves":
		return true
	}
	return false
}

func crossStatementPossessive(word string) bool {
	switch strings.ToLower(word) {
	case "his", "her", "their", "its":
		return true
	}
	return false
}

// Logical scaffolds consume "it" as syntax, not as a matched term. Preserve
// only the catalog's closed spellings, including its contraction aliases.
func crossStatementScaffold(s string) bool {
	s = strings.ToLower(strings.Join(strings.Fields(expandGateContractions(s)), " "))
	for _, prefix := range []string{"it is the case that", "it is not the case that", "it is the case both that"} {
		if strings.HasPrefix(s, prefix) {
			rest := s[len(prefix):]
			if rest != "" && (rest[0] == ' ' || rest[0] == ',') {
				return true
			}
		}
	}
	return false
}

// isolateStatementPronouns gives each (statement number, token) a private
// spelling. Repeated tokens within ONE statement retain their identity. No
// antecedent, gender, or name knowledge is used. The fresh prefix prevents a
// user-written word from accidentally matching an internal spelling.
func isolateStatementPronouns(s, prefix string, number int, selectWord func(string) bool) (string, bool) {
	var out strings.Builder
	last, changed := 0, false
	for _, loc := range crossStatementWord.FindAllStringIndex(s, -1) {
		word := s[loc[0]:loc[1]]
		if strings.EqualFold(word, "it") && crossStatementScaffold(s[loc[0]:]) || !selectWord(word) {
			continue
		}
		out.WriteString(s[last:loc[0]])
		// Letters only, so the spelling cannot be mistaken for a numeral or an
		// explicit schematic variable. Capitalization makes it a name-shaped term.
		n, suffix := number, ""
		for {
			suffix += string(rune('a' + n%26))
			n /= 26
			if n == 0 {
				break
			}
		}
		out.WriteString(prefix + suffix + strings.ToLower(word))
		last, changed = loc[1], true
	}
	if !changed {
		return s, false
	}
	out.WriteString(s[last:])
	return out.String(), true
}

// crossStatementFailure checks successful matches at their common boundary,
// including direct struct/string comparisons, schema/slot binding, and the
// FOL and scope fallbacks. A proof must survive making free pronouns local to
// numbered statements. This avoids a blanket ban on pronouns in a cited
// statement: an unused conjunct/disjunct is not a cross-statement match.
//
// Replay uses the same deterministic catalog with only these two rules off;
// it never runs the public validator or the diagnostic retries recursively.
func (policy strictRules) crossStatementFailure(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	if !policy.term_pronoun_cross_statement && !policy.term_possessive_cross_statement {
		return ""
	}
	personal, possessive := false, false
	selectWord := func(word string) bool {
		p := policy.term_pronoun_cross_statement && crossStatementPronoun(word)
		d := policy.term_possessive_cross_statement && crossStatementPossessive(word)
		personal, possessive = personal || p, possessive || d
		return p || d
	}
	prefix := "Argchecklocal"
	for {
		collision := strings.Contains(strings.ToLower(c.Statement), strings.ToLower(prefix))
		for _, step := range steps {
			if strings.Contains(strings.ToLower(step.Statement), strings.ToLower(prefix)) {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
		prefix += "local"
	}
	isolated := make(map[int]NumberedArgumentStep, len(steps))
	changed := false
	for id, step := range steps {
		var edited bool
		step.Statement, edited = isolateStatementPronouns(step.Statement, prefix, step.Number, selectWord)
		changed = changed || edited
		isolated[id] = step
	}
	target := c
	var edited bool
	target.Statement, edited = isolateStatementPronouns(c.Statement, prefix, c.Number, selectWord)
	if !changed && !edited {
		return ""
	}
	replay := policy.without("term.pronoun_cross_statement").without("term.possessive_cross_statement")
	replay.isolatedPronounPrefix = prefix
	replay.fired = nil // speculative failures must not leak into strict notes
	replay.slotNames = schematicNameEvidence(isolated)
	replay.termSlots = ruleTermSlots(target, isolated)
	replay.pluralWitnessBlocked = rulePluralWitnessBlocks(target, isolated)
	if replay.gateConclusionStepUnisolated(target, isolated).Passed {
		return ""
	}
	// Isolate both enabled classes TOGETHER: otherwise alternative schema
	// matches could each rely on the other class and evade both guards.
	var ids []string
	if personal && policy.hit("term.pronoun_cross_statement") {
		ids = append(ids, "term.pronoun_cross_statement")
	}
	if possessive && policy.hit("term.possessive_cross_statement") {
		ids = append(ids, "term.possessive_cross_statement")
	}
	return fmt.Sprintf("%s: a pronoun's referent depends on context and can differ between statements; replace it with a name or noun phrase", strings.Join(ids, ", "))
}

// A schematic value links a concrete statement to the rule statement even
// when that variable is absent from the consequent. Renaming alone cannot
// detect this edge: a variable could absorb the private spelling unchanged.
func (policy strictRules) crossStatementBindingRefused(value string) bool {
	if policy.isolatedPronounPrefix != "" && strings.Contains(strings.ToLower(value), strings.ToLower(policy.isolatedPronounPrefix)) {
		return true
	}
	refused := false
	for _, loc := range crossStatementWord.FindAllStringIndex(value, -1) {
		word := value[loc[0]:loc[1]]
		if strings.EqualFold(word, "it") && crossStatementScaffold(value[loc[0]:]) {
			continue
		}
		if policy.term_pronoun_cross_statement && crossStatementPronoun(word) && policy.hit("term.pronoun_cross_statement") {
			refused = true
		}
		if policy.term_possessive_cross_statement && crossStatementPossessive(word) && policy.hit("term.possessive_cross_statement") {
			refused = true
		}
	}
	return refused
}
