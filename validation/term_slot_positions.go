package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Identity substitution (Leibniz's law) is certified only in known-safe
// positions: (a) when the substituted premise is itself an identity sentence
// (identity chains), or (b) when every occurrence of the replaced term is the
// grammatical subject of a plain clause. Objects of verbs ("Lois is looking
// for Superman", "Lois admires Superman"), passives ("is admired by") and
// raising forms ("is believed to") are not certified. An allowlist of
// positions stays strict where a list of opaque verbs could never be complete.

var clauseSubjectLead = []string{"if", "then", "and", "or", "it is the case that", "it is not the case that", "both that", "and that", "either", "both"}

var passiveOrRaising = regexp.MustCompile(`^\s*(?:(?:is|are|was|were|be|been|being)\s+(?:not\s+)?(?:said|known|thought|believed|rumou?red|reported|alleged|supposed|expected|considered|claimed|assumed|presumed|seen|sought|felt|held|likely|unlikely|certain|sure|bound|apt|liable|bought|taught|caught|brought|fought|made|built|sent|spent|lent|found|ground|wound|hung|led|read|paid|laid|kept|left|dealt|meant|met|sold|told|struck|stuck|shot|hit|put|set|cut|hurt|let|shut|won|spun|begun|sung|rung|drunk|swum|held|heard|loved|owned|[a-z]+ed|[a-z]+en)\s+(?:by|to)\b|(?:does\s+not\s+|did\s+not\s+)?(?:seems?|seemed|appears?|appeared|happens?|happened|tends?|tended|proves?|proved|turns?\s+out|turned\s+out)\s+to\b)`)

// termSubjectPositions reports whether every word-bounded occurrence of term
// in text is the subject of a plain clause, and how many occurrences there are.
func termSubjectPositions(text, term string) (safe bool, count int) {
	lower := strings.ToLower(text)
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return false, 0
	}
	position := 0
	safe = true
	for {
		index := strings.Index(lower[position:], term)
		if index < 0 {
			break
		}
		start := position + index
		end := start + len(term)
		position = start + 1
		if (start > 0 && wordRune(lastRune(lower[:start]))) || (end < len(lower) && wordRune(firstRune(lower[end:]))) {
			continue
		}
		count++
		prefix := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(lower[:start]), ","))
		lead := prefix == "" || strings.HasSuffix(prefix, "(") || strings.HasSuffix(prefix, "such that")
		for _, word := range clauseSubjectLead {
			if prefix == word || strings.HasSuffix(prefix, " "+word) || strings.HasSuffix(prefix, ","+word) {
				lead = true
			}
		}
		if strings.HasSuffix(strings.TrimSpace(lower[:start]), ",") && !lead {
			// A comma-led clause ("..., Superman flies") also starts a clause.
			lead = true
		}
		rest := lower[end:]
		if strings.HasPrefix(rest, "'s ") || strings.HasPrefix(rest, "’s ") {
			// A possessive inside the subject phrase ("Superman's cape is red").
			rest = rest[strings.Index(rest, " "):]
			if idx := strings.Index(rest, " is "); idx >= 0 {
				rest = rest[idx:]
			}
		}
		if !lead || passiveOrRaising.MatchString(rest) {
			safe = false
		}
	}
	return safe && count > 0, count
}

// identitySubstitutionUnsafe returns a reason when a step that substitutes
// identicals does so outside the safe positions, and "" when it is safe.
func identitySubstitutionUnsafe(premises []string, conclusion string) string {
	for i, identityPremise := range premises {
		if !explicitIdentity(identityPremise) {
			continue
		}
		identities := identityCandidates(identityPremise)
		if len(identities) == 0 {
			continue
		}
		for j, other := range premises {
			if j == i {
				continue
			}
			if explicitIdentity(other) && explicitIdentity(conclusion) {
				return "" // identity chain
			}
			otherText := normalizeClausePreservingCase(other)
			for _, identity := range identities {
				for _, d := range [][2]string{{identity.left, identity.right}, {identity.right, identity.left}} {
					if !strings.Contains(strings.ToLower(otherText), strings.ToLower(d[0])) {
						continue
					}
					safe, count := termSubjectPositions(otherText, d[0])
					if count == 0 {
						continue
					}
					conclusionSafe, _ := termSubjectPositions(normalizeClausePreservingCase(conclusion), d[1])
					if safe && conclusionSafe {
						return ""
					}
					return fmt.Sprintf("substitution is certified only in identity sentences and in the subject of a plain sentence; %q here is not the subject of a plain clause (an object, passive or raising position). If the step is sound, state it as a premise", d[0])
				}
			}
		}
	}
	return ""
}

func stepStatements(steps []NumberedArgumentStep) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.Statement
	}
	return out
}

var explicitIdentityPattern = regexp.MustCompile(`(?i)\b(?:is|are|was|were) (?:not )?(?:identical (?:to|with)|the same (?:person|thing|object|individual) as|one and the same as)\b|\s=\s`)

// explicitIdentity reports an identity sentence written as one ("A is
// identical to B"), as opposed to any copular "A is B".
func explicitIdentity(statement string) bool {
	return explicitIdentityPattern.MatchString(statement)
}

var indefiniteHeads = map[string]bool{
	"a": true, "an": true, "some": true, "any": true, "someone": true, "something": true,
	"somebody": true, "anyone": true, "anything": true, "anybody": true, "several": true,
	"many": true, "one": true,
}

// indefiniteBindingPositionSafe limits schematic bindings that are indefinite
// or quantifier phrases ("a unicorn", "someone") to safe positions in the
// instance clause: its subject, or a predicate after a copula ("is a hero").
// In an object slot ("Lois is looking for a unicorn") the phrase need not pick
// out anything, so binding it would let "a unicorn exists" through. Names,
// definite descriptions and other bindings are unaffected.
func indefiniteBindingPositionSafe(concrete []rune, start int, segment string) bool {
	words := strings.Fields(strings.ToLower(segment))
	if len(words) == 0 || !indefiniteHeads[words[0]] {
		return true
	}
	prefix := strings.TrimSpace(strings.ToLower(string(concrete[:start])))
	if prefix == "" {
		return true
	}
	for _, lead := range append([]string{"is", "are", "was", "were", "be", "been", "is not", "are not", "was not", "were not"}, clauseSubjectLead...) {
		if prefix == lead || strings.HasSuffix(prefix, " "+lead) || strings.HasSuffix(prefix, ","+lead) {
			return true
		}
	}
	return strings.HasSuffix(prefix, ",")
}

// indefiniteSlotOverride is a reserved substitution key. When set, schematic
// unification skips the indefinite-position check; only the diagnostic pass
// below uses it, to explain a refusal. It is never a schematic marker.
const indefiniteSlotOverride rune = '\uFFFE'

// indefiniteSlotDiagnosis explains a failed schematic match that fails only
// because an indefinite would bind in an unsafe slot. It returns "" when the
// match fails for another reason too.
func indefiniteSlotDiagnosis(template string, premise NumberedArgumentStep, initial schematicSubstitution) string {
	relaxed := cloneSchematicSubstitution(initial)
	relaxed[indefiniteSlotOverride] = "1"
	concrete := []rune(comparisonClause(premise.Statement))
	for _, substitution := range unifySchematicTemplate(template, premise.Statement, relaxed) {
		for variable, value := range substitution {
			if variable == indefiniteSlotOverride {
				continue
			}
			if _, preset := initial[variable]; preset {
				continue
			}
			start := strings.Index(string(concrete), value)
			if start < 0 {
				continue
			}
			runeStart := len([]rune(string(concrete)[:start]))
			if indefiniteBindingPositionSafe(concrete, runeStart, value) {
				continue
			}
			return fmt.Sprintf(
				"premise %d binds a schematic variable to %q in an object position (%q); an indefinite there need not pick anything out, so the schema cannot be instantiated with it. State the existence claim as its own premise, e.g. %q",
				premise.Number,
				value,
				schematicDisplay(template),
				indefiniteRestatement(premise.Statement, value),
			)
		}
	}
	return ""
}

// indefiniteRestatement suggests "There is a unicorn that Lois is looking for."
// for the premise "Lois is looking for a unicorn." and the phrase "a unicorn".
func indefiniteRestatement(statement, phrase string) string {
	body := strings.TrimRight(strings.TrimSpace(statement), ".")
	lower := strings.ToLower(body)
	index := strings.Index(lower, strings.ToLower(phrase))
	if index < 0 {
		return "There is " + phrase + " such that " + body + "."
	}
	rest := strings.Join(strings.Fields(body[:index]+" "+body[index+len(phrase):]), " ")
	return "There is " + phrase + " that " + rest + "."
}

// relaxedInitial returns the initial substitution for a schematic match,
// carrying indefiniteSlotOverride when any premise is a diagnostic copy.
func relaxedInitial(initial schematicSubstitution, premises ...NumberedArgumentStep) schematicSubstitution {
	for _, premise := range premises {
		if premise.relaxIndefinite {
			next := cloneSchematicSubstitution(initial)
			next[indefiniteSlotOverride] = "1"
			return next
		}
	}
	return initial
}

// indefiniteSlotStepMessage names the refused binding for a step that passes
// only when an indefinite may bind in an object slot.
func indefiniteSlotStepMessage(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	for _, id := range c.Citations {
		conditional, ok := steps[id]
		if !ok {
			continue
		}
		for _, pair := range conditionalCandidatesPreservingCase(conditional.Statement) {
			for _, conjunct := range conjunctiveAntecedentCandidatesOrSelf(pair.left) {
				template, variables := schematicTemplate(conjunct)
				if len(variables) == 0 {
					continue
				}
				for _, other := range c.Citations {
					premise, ok := steps[other]
					if !ok || other == id {
						continue
					}
					if reason := indefiniteSlotDiagnosis(template, premise, nil); reason != "" {
						return reason
					}
				}
			}
		}
	}
	return "an indefinite (a/an/some/someone ...) would bind a schematic variable in an object position, where it need not pick anything out; state the existence claim as its own premise (\"There is a N that ...\")"
}

// conjunctiveAntecedentCandidatesOrSelf lists an antecedent and every conjunct
// of its two- to four-way conjunctive splits, for diagnosis only.
func conjunctiveAntecedentCandidatesOrSelf(antecedent string) []string {
	seen := map[string]bool{antecedent: true}
	out := []string{antecedent}
	for parts := 2; parts <= 4; parts++ {
		for _, split := range conjunctiveAntecedentCandidates(antecedent, parts) {
			for _, conjunct := range split {
				if !seen[conjunct] {
					seen[conjunct] = true
					out = append(out, conjunct)
				}
			}
		}
	}
	return out
}
