package validation

import (
	"strings"
	"unicode"
)

var termNumeralWords = wordSet("zero one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen seventeen eighteen nineteen twenty thirty forty fifty sixty seventy eighty ninety hundred thousand million dozen")

func numeralTerm(s string) bool {
	w := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	if len(w) == 0 {
		return false
	}
	if termNumeralWords[w[0]] {
		return true
	}
	return unicode.IsDigit(firstRune(w[0]))
}

// A rule's variables have one type across its complete statement, including
// antecedent/consequent fragments subsequently passed to separate matchers.
func ruleTermSlots(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) [256]bool {
	out := [256]bool{}
	for _, id := range c.Citations {
		text, _ := replaceSchematicAlphabet(steps[id].Statement, true)
		rs := []rune(strings.ToLower(text))
		for i, r := range rs {
			v, ok := schematicVariableForMarker(r)
			if !ok {
				continue
			}
			before := strings.TrimSpace(string(rs[:i]))
			after := strings.TrimSpace(string(rs[i+1:]))
			if (before == "then" || strings.HasSuffix(before, " then")) && (after == "" || after == ".") {
				continue
			}
			if afterCopula(before) {
				continue
			}
			if !clauseSlotBefore(rs, i) {
				out[v] = true
			}
		}
	}
	return out
}

// Names are evidenced before case-folding. Quantified surrounding clauses are
// refused even when the witness itself is a name: a definite can covary there.
func (p strictRules) egWitnessSafe(statement, term string) bool {
	s := normalizeClause(statement)
	t := normalizeClause(term)
	if p.eg_name_witness {
		if !p.singularName(strings.Fields(t)) && p.hit("eg.name_witness") {
			return false
		}
		for _, w := range strings.Fields(s) {
			if (termNumeralWords[w] || termDeterminers[w] || numeralTerm(w)) && p.hit("eg.name_witness") {
				return false
			}
		}
	}
	if !p.eg_subject_only || !p.hit("eg.subject_only") {
		return true
	}
	for _, pair := range identityCandidates(statement) {
		if normalizeClause(pair.left) == t || normalizeClause(pair.right) == t {
			return true
		}
	}
	// Copular identity requires two positively evidenced terms.
	for _, pair := range instanceCandidates(statement) {
		if p.singularOnlyValue(pair.left) && p.singularOnlyValue(pair.right) && (pair.left == t || pair.right == t) {
			return true
		}
	}
	if !strings.HasPrefix(s, t+" ") {
		return false
	}
	tail := strings.TrimPrefix(s, t+" ")
	// Closed predicate grammar avoids guessing whether an unfamiliar copula
	// complement is a passive participle. Active verbs use the existing register.
	if extensionalTail(tail) {
		return true
	}
	words := strings.Fields(tail)
	if len(words) == 0 || !extensionalVerbHeads[words[0]] {
		return false
	}
	_, n, ok := replaceTermOccurrences(s, t, "round2witness")
	return ok && n == 1
}

func (p strictRules) egFormulaSafe(f *folFormula, witness string) bool {
	if p.eg_name_witness && !p.singularName(strings.Fields(witness)) && p.hit("eg.name_witness") {
		return false
	}
	if f.op == "equal" {
		return true
	}
	// Conservative: a compound or quantified matrix needs its own existence
	// premise; an atom must expose its subject as the first argument.
	if f.op != "atom" {
		subjectAllowed := !p.eg_subject_only || !p.hit("eg.subject_only")
		nameAllowed := !p.eg_name_witness || !p.hit("eg.name_witness")
		return subjectAllowed && nameAllowed
	}
	if !p.eg_subject_only || !p.hit("eg.subject_only") {
		return true
	}
	if len(f.terms) == 0 || f.terms[0] != witness {
		return false
	}
	for _, t := range f.terms[1:] {
		if t == witness {
			return false
		}
	}
	if strings.HasPrefix(f.symbol, "is/") {
		return extensionalTail("is " + strings.TrimPrefix(f.symbol, "is/"))
	}
	return extensionalVerbHeads[f.symbol]
}

func extensionalTail(tail string) bool {
	w := strings.Fields(strings.ToLower(tail))
	if len(w) < 2 {
		return false
	}
	switch w[0] {
	case "is", "are", "was", "were":
	default:
		return false
	}
	w = w[1:]
	if len(w) > 0 && w[0] == "not" {
		w = w[1:]
	}
	if len(w) == 1 {
		return singularPredicateHeads[w[0]]
	}
	if len(w) == 2 && (w[0] == "a" || w[0] == "an") {
		for _, r := range w[1] {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return !singularClosedWord(w[1])
	}
	return len(w) >= 3 && w[0] == "identical" && w[1] == "to"
}

func extensionalIdentityContext(statement, term string) bool {
	s := normalizeClause(statement)
	t := normalizeClause(term)
	if explicitIdentity(s) {
		return true
	}
	if !strings.HasPrefix(s, t+" ") {
		return false
	}
	_, n, ok := replaceTermOccurrences(s, t, "round2term")
	return ok && n == 1 && extensionalTail(strings.TrimPrefix(s, t+" "))
}

func letterNearName(statement string) bool {
	statement = strings.ReplaceAll(statement, string(nameLetterJoiner), "")
	w := strings.Fields(statement)
	for i, raw := range w {
		v := []rune(strings.Trim(raw, ".,;:!?()[]"))
		if len(v) != 1 {
			continue
		}
		if _, ok := schematicVariableMarkers[v[0]]; !ok {
			continue
		}
		if i > 0 {
			prefix := strings.Join(w[:i], " ") + " "
			if nameBeforeLetter([]rune(prefix+string(v)), len([]rune(prefix))) != "" {
				return true
			}
		}
		if i+1 < len(w) {
			next := strings.Trim(w[i+1], ".,;:!?()[]")
			if len([]rune(next)) > 1 && unicode.IsUpper(firstRune(next)) {
				return true
			}
		}
	}
	return false
}

// This final guard covers identity routes in the legacy, FOL, readable,
// counting and description catalogs. It runs before any successful result is
// exposed; it does not rely on which matcher won the schema search.
func (p strictRules) round2TermFailure(c NumberedArgumentStep, steps map[int]NumberedArgumentStep, report EntailmentGateStepReport) string {
	if p.slot_one_type {
		for _, id := range c.Citations {
			text, _ := replaceSchematicAlphabet(steps[id].Statement, true)
			rs := []rune(strings.ToLower(text))
			for i, r := range rs {
				v, ok := schematicVariableForMarker(r)
				if !ok || !p.termSlots[v] {
					continue
				}
				before := strings.TrimSpace(string(rs[:i]))
				if clauseSlotBefore(rs, i) && !(before == "if" || strings.HasSuffix(before, " if")) && p.hit("slot.one_type") {
					return "slot.one_type: mixed subordinate-clause and term uses are ambiguous; write the phrase into the rule, or name the thing in a rule without mixed slots"
				}
			}
		}
	}
	if p.slot_letter_in_name {
		for _, id := range c.Citations {
			if letterNearName(steps[id].Statement) && p.hit("slot.letter_in_name") {
				return "slot.letter_in_name: a schematic letter adjacent to a capitalized word may be part of a proper name; use a variable away from names, or write the name out"
			}
		}
	}
	if !report.Passed {
		return ""
	}
	form := strings.ToLower(report.Form)
	if strings.Contains(form, "identity") || strings.Contains(form, "substitution") || strings.Contains(form, "description") || strings.Contains(form, "counting") {
		statements := []string{c.Statement}
		for _, id := range c.Citations {
			statements = append(statements, steps[id].Statement)
		}
		if p.identity_constant_sides {
			for _, s := range statements {
				for _, pair := range p.round2IdentityPairs(s) {
					if (!p.singularOnlyValue(pair.left) || !p.singularOnlyValue(pair.right)) && p.hit("identity.constant_sides") {
						return "identity.constant_sides: identity sides must be names or singular definites; name the thing, or use 'the dog'; alternatively, state the substituted claim as a premise"
					}
				}
			}
		}
		if p.identity_extensional_context {
			for _, s := range statements[1:] {
				for _, pair := range p.round2IdentityPairs(s) {
					for _, other := range statements {
						if len(p.round2IdentityPairs(other)) > 0 {
							continue
						}
						for _, t := range []string{pair.left, pair.right} {
							_, n, ok := replaceTermOccurrences(other, t, "round2term")
							if !ok || n == 0 {
								continue
							}
							// Older positional refusals retain their diagnostic attribution.
							if !p.identity_safe_position {
								if safe, _ := termSubjectPositions(other, t); !safe {
									continue
								}
							}
							if !extensionalIdentityContext(other, t) && p.hit("identity.extensional_context") {
								return "identity.extensional_context: substitution requires a closed extensional copular predicate; state the substituted claim as a premise"
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func formulaNumeralTerm(f *folFormula) bool {
	for _, term := range f.terms {
		if numeralTerm(term) {
			return true
		}
	}
	for _, child := range f.children {
		if formulaNumeralTerm(child) {
			return true
		}
	}
	return false
}

// Include the v28 synonym and the v27 capitalized copular-identity surface;
// both can otherwise reach the FOL identity matcher under a bridge form name.
func (p strictRules) round2IdentityPairs(statement string) []clausePair {
	s := strings.NewReplacer(" is identical with ", " is identical to ", " are identical with ", " is identical to ", " are identical to ", " is identical to ").Replace(normalizeClause(statement))
	// Wide negation is a logical wrapper, not part of the left identity
	// term. Keep checking both terms of the enclosed identity.
	for strings.HasPrefix(s, "it is not the case that ") {
		s = normalizeClause(strings.TrimPrefix(s, "it is not the case that "))
	}
	pairs := identityCandidates(s)
	if len(pairs) > 0 {
		return pairs
	}
	for _, pair := range instanceCandidates(s) {
		if p.singularName(strings.Fields(pair.right)) {
			pairs = append(pairs, pair)
		}
	}
	return pairs
}

func (p strictRules) round2RewriteHint(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	s := normalizeClause(c.Statement)
	if (p.eg_subject_only || p.eg_name_witness) && (strings.HasPrefix(s, "there ") || strings.HasPrefix(s, "some ") || strings.HasPrefix(s, "something ")) {
		return "; state the existential claim as its own premise"
	}
	for _, id := range c.Citations {
		s = normalizeClause(steps[id].Statement)
		if p.term_inflection_verbatim && strings.HasPrefix(s, "there ") {
			return "; use the same singular noun spelling; write universal rules with singular 'every N' and existential claims as 'There is a N ...'"
		}
		if p.term_inflection_verbatim && (strings.HasPrefix(s, "all ") || strings.HasPrefix(s, "every ") || strings.HasPrefix(s, "no ") || strings.HasPrefix(s, "some ")) {
			return "; write the rule with singular 'every N', using the same noun spelling; name any numeral-led term"
		}
	}
	return ""
}
