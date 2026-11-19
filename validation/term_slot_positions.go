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
func (policy strictRules) identitySubstitutionUnsafe(premises []string, conclusion string) (reason string) {
	if !policy.identity_safe_position {
		return ""
	}
	defer func() { _ = reason != "" && policy.hit("identity.safe_position") }()
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
func (policy strictRules) indefiniteBindingPositionSafe(concrete []rune, start int, segment string) bool {
	words := strings.Fields(strings.ToLower(segment))
	if len(words) == 0 {
		return true
	}
	prefix := strings.TrimSpace(strings.ToLower(string(concrete[:start])))
	if !indefiniteHeads[words[0]] {
		// v43.0 (slot.indefinite_written): a variable written just after an
		// indefinite in the rule ("greets a X", "a X arrived", "is a X") is
		// refused in every position. The instance need not be an article
		// plus a noun ("a lot of people", "a couple of neighbours"), and
		// telling the two apart takes a lexicon the gate does not have.
		return !(policy.slot_indefinite_written && writtenIndefiniteBefore(prefix) != "" && policy.hit("slot.indefinite_written"))
	}
	// v43.0 (slot.indefinite_value): a value that is to stand for one thing
	// must name one thing. After a copula the value is a predicate, and any
	// predicate phrase may be substituted.
	if policy.slot_indefinite_value && !afterCopula(prefix) && !indefiniteValueShapeOK(words) && policy.hit("slot.indefinite_value") {
		return false
	}
	return !(policy.slot_indefinite_object && !indefiniteSlotSafe(prefix) && policy.hit("slot.indefinite_object"))
}

// indefiniteSlotSafe is the v39.0 position test for an indefinite value:
// subject, predicate after a copula, or clause start. prefix is the
// lower-cased text before the value.
func indefiniteSlotSafe(prefix string) bool {
	if prefix == "" {
		return true
	}
	for _, lead := range append(append([]string{}, copulaLeads...), clauseSubjectLead...) {
		if prefix == lead || strings.HasSuffix(prefix, " "+lead) || strings.HasSuffix(prefix, ","+lead) {
			return true
		}
	}
	return strings.HasSuffix(prefix, ",")
}

var copulaLeads = []string{"is", "are", "was", "were", "be", "been", "being", "is not", "are not", "was not", "were not"}

// afterCopula reports whether prefix ends in a copula, so the value that
// follows is a predicate ("is a Bayesian agent with the universal prior").
func afterCopula(prefix string) bool {
	words := strings.Fields(prefix)
	// Existential "there is X" introduces a thing; it is not a predicate.
	if len(words) >= 2 && words[len(words)-2] == "there" {
		return false
	}
	for _, lead := range copulaLeads {
		if prefix == lead || strings.HasSuffix(prefix, " "+lead) || strings.HasSuffix(prefix, ","+lead) {
			return true
		}
	}
	return false
}

// indefiniteValueHeads may open an indefinite value that is then read as one
// thing: an article or "some"/"one" before a modified head noun.
var indefiniteValueHeads = wordSet("a an some one")

// indefinitePronounValues stand alone as an indefinite value.
var indefinitePronounValues = wordSet("someone something somebody anyone anything anybody")

// indefiniteValueStoppers end a noun head: anything after one of them
// ("a lot of people", "a man who is tall", "a dog and a cat") makes the
// phrase more than determiner + modifiers + head.
var indefiniteValueStoppers = wordSet("who whom whose which that where when whether than and or nor but if because as")

// indefiniteValueShapeOK (v43.0) accepts an indefinite value for a schematic
// variable only as a lone indefinite pronoun ("someone") or as a/an/some/one,
// optional modifiers, and one head word with nothing after it ("a red ball").
// A complement or quantity phrase ("a lot of people", "a friend of Bob",
// "a few runners") need not be one thing, so it is refused in every slot.
// Values led by other quantifiers keep their v42.0 treatment.
func indefiniteValueShapeOK(words []string) bool {
	if len(words) == 0 {
		return true
	}
	if indefinitePronounValues[words[0]] {
		return len(words) == 1
	}
	if !indefiniteValueHeads[words[0]] {
		// Other indefinite heads ("many robots", "several robots", "any
		// object ...") keep their v42.0 treatment.
		return true
	}
	if len(words) < 2 {
		return false
	}
	// Closed-class determiners that begin with an article ("a few", "a
	// little", "a great many", "a good many") quantify; they do not name one
	// thing. "a little" is refused even where it could be an adjective ("a
	// little dog"): the gate cannot tell the two apart.
	if words[1] == "few" || words[1] == "little" || (len(words) > 2 && (words[1] == "great" || words[1] == "good") && words[2] == "many") {
		return false
	}
	for _, word := range words[1:] {
		word = strings.Trim(word, ".,;:\"'()")
		if word == "" || indefiniteValueStoppers[word] || framePrepositions[word] || indefiniteHeads[word] {
			return false
		}
	}
	return true
}

// literalRuleSuggestion writes value in place of variable in the displayed
// template ("X arrived" -> "a lot of people arrived").
func literalRuleSuggestion(template string, variable rune, value string) string {
	display := schematicDisplay(template)
	name, ok := schematicVariableForMarker(variable)
	if !ok {
		name = variable
	}
	return strings.ReplaceAll(display, string(name), value)
}

// originalCaseSpan returns value as it is spelled in source, for messages: the
// matcher works on a lowercased comparison clause. It falls back to value when
// source does not contain it case-insensitively at the same byte length.
func originalCaseSpan(value, source string) string {
	lower := strings.ToLower(source)
	if len(lower) != len(source) {
		return value
	}
	if i := strings.Index(lower, value); i >= 0 {
		return source[i : i+len(value)]
	}
	return value
}

// writtenIndefiniteBefore returns the indefinite word (a, an, some ...) that
// ends prefix, or "" when prefix does not end in one.
func writtenIndefiniteBefore(prefix string) string {
	prefixWords := strings.Fields(prefix)
	if len(prefixWords) == 0 {
		return ""
	}
	last := strings.Trim(prefixWords[len(prefixWords)-1], "\"'(")
	if indefiniteHeads[last] {
		return last
	}
	return ""
}

// indefiniteSlotOverride is a reserved substitution key. When set, schematic
// unification skips the indefinite-position check; only the diagnostic pass
// below uses it, to explain a refusal. It is never a schematic marker.
const indefiniteSlotOverride rune = '\uFFFE'

// indefiniteSlotDiagnosis explains a failed schematic match that fails only
// because an indefinite would bind in an unsafe slot. It returns "" when the
// match fails for another reason too.
func (policy strictRules) indefiniteSlotDiagnosis(template string, premise NumberedArgumentStep, initial schematicSubstitution) string {
	relaxed := cloneSchematicSubstitution(initial)
	relaxed[indefiniteSlotOverride] = "1"
	concrete := []rune(policy.comparisonClause(premise.Statement))
	for _, substitution := range policy.unifySchematicTemplate(template, premise.Statement, relaxed) {
		for variable, value := range substitution {
			if variable == indefiniteSlotOverride || variable == singularLegacyOverride {
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
			marker := schematicVariableMarkers[variable]
			templateRunes := []rune(template)
			ti := strings.IndexRune(template, marker)
			if ti < 0 {
				continue
			}
			ti = len([]rune(template[:ti]))
			if !policy.singularBindingSafe(templateRunes, ti, concrete, runeStart, value) {
				shown := originalCaseSpan(value, premise.Statement)
				return fmt.Sprintf("slot.value_singular: premise %d binds a schematic variable to %q; a quantity, plural, group or indefinite phrase need not be one thing, so a rule about each thing does not apply to it. Name the thing (with capitalization evidenced away from sentence start or Therefore/So/Thus/Hence), use a singular definite description (the dog), or state existence as its own premise and then name it. Write the phrase itself into the rule instead of a variable (%q)", premise.Number, shown, literalRuleSuggestion(template, variable, shown))
			}
			if policy.indefiniteBindingPositionSafe(concrete, runeStart, value) {
				continue
			}
			if valueWords := strings.Fields(strings.ToLower(value)); len(valueWords) > 0 && indefiniteHeads[valueWords[0]] && !indefiniteValueShapeOK(valueWords) && indefiniteSlotSafe(strings.TrimSpace(strings.ToLower(string(concrete[:runeStart])))) {
				return fmt.Sprintf(
					"premise %d binds a schematic variable to %q (%q); an indefinite value is accepted only as a/an, some or one plus modifiers and one head noun with nothing after it (or a lone \"someone\"), because a quantity or complement phrase (\"a lot of people\", \"a friend of Bob\", \"a few runners\") need not be one thing. Write the phrase itself into the rule instead of a variable (%q), or name the individuals",
					premise.Number,
					value,
					schematicDisplay(template),
					literalRuleSuggestion(template, variable, value),
				)
			}
			if article := writtenIndefiniteBefore(strings.ToLower(string(concrete[:runeStart]))); article != "" {
				return fmt.Sprintf(
					"premise %d binds a schematic variable to %q just after %q, which the rule writes before the variable (%q); the phrase need not be %q plus a noun (\"a lot of people\"), so the schema cannot be instantiated with it. Drop %q from the rule and let the variable stand for the whole phrase (%q) when that phrase is an evidenced name or singular definite description; otherwise write the phrase itself into the rule",
					premise.Number,
					value,
					article,
					schematicDisplay(template),
					article,
					article,
					article+" "+value,
				)
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
	var next schematicSubstitution
	for _, premise := range premises {
		if premise.relaxIndefinite || premise.relaxClauseTerm {
			if next == nil {
				next = cloneSchematicSubstitution(initial)
			}
			if premise.relaxIndefinite {
				next[indefiniteSlotOverride] = "1"
			}
			if premise.relaxClauseTerm {
				next[clauseTermOverride] = "1"
			}
		}
	}
	if next == nil {
		return initial
	}
	return next
}

// clauseTermOverride is a reserved substitution key: when set, a term that
// reads as containing a clause may bind a variable. Diagnosis only.
const clauseTermOverride rune = '\uFFFD'

// clauseTermStepMessage names the binding refused because it reads as a term
// containing a clause ("physics exploits not known to humans": "exploits" can
// be a verb), for a step that passes only without that check.
func (policy strictRules) clauseTermStepMessage(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	for _, id := range c.Citations {
		conditional, ok := steps[id]
		if !ok {
			continue
		}
		for _, pair := range conditionalCandidatesPreservingCase(conditional.Statement) {
			for _, conjunct := range conjunctiveAntecedentCandidatesOrSelf(pair.left) {
				template, variables := policy.schematicTemplate(conjunct)
				if len(variables) == 0 {
					continue
				}
				for _, other := range c.Citations {
					premise, ok := steps[other]
					if !ok || other == id {
						continue
					}
					relaxed := schematicSubstitution{clauseTermOverride: "1"}
					for _, substitution := range policy.unifySchematicTemplate(template, premise.Statement, relaxed) {
						for variable, value := range substitution {
							if variable == clauseTermOverride || variable == indefiniteSlotOverride {
								continue
							}
							if policy.clauseInsideTerm(value) {
								return fmt.Sprintf("premise %d would bind a schematic variable to %q, which reads as containing a clause (a word in it can be a verb with its own complement), so it is not certified as a single term. Write the phrase into the rule instead of a variable, or reword the term (for example with a determiner or a relative clause) so that no word in it can be read as its verb", premise.Number, value)
							}
						}
					}
				}
			}
		}
	}
	return ""
}

// indefiniteSlotStepMessage names the refused binding for a step that passes
// only when an indefinite may bind in an object slot.
func (policy strictRules) indefiniteSlotStepMessage(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	for _, id := range c.Citations {
		conditional, ok := steps[id]
		if !ok {
			continue
		}
		pairs := conditionalCandidatesPreservingCase(conditional.Statement)
		// Counterfactual MP matches an auxiliary-stripped virtual conditional.
		// Diagnose that same surface while preserving lowercase literal names.
		marked, _ := replaceSchematicVariables(conditional.Statement)
		for _, pair := range counterfactualPlainPairs(marked) {
			for variable, marker := range schematicVariableMarkers {
				pair.left = strings.ReplaceAll(pair.left, string(marker), string(variable))
				pair.right = strings.ReplaceAll(pair.right, string(marker), string(variable))
			}
			pairs = append(pairs, pair)
		}
		for _, pair := range pairs {
			for _, conjunct := range append(conjunctiveAntecedentCandidatesOrSelf(pair.left), pair.right) {
				// Diagnosis must cover the same U/V/W/X/Y/Z alphabet as the
				// successful relaxed matcher, not only its older X/Y/Z subset.
				template, variables := policy.extendedSchematicTemplate(conjunct)
				if len(variables) == 0 {
					continue
				}
				for _, other := range append(append([]int(nil), c.Citations...), c.Number) {
					premise, ok := steps[other]
					if !ok || other == id {
						continue
					}
					if reason := policy.indefiniteSlotDiagnosis(template, premise, nil); reason != "" {
						return reason
					}
				}
			}
		}
	}
	return "an indefinite (a/an/some/someone ...) would bind a schematic variable where it need not pick out one thing: in an object position, right after an indefinite the rule writes (\"a X\"), or as a phrase with something after its noun (\"a lot of people\", \"a friend of Bob\"); state the existence claim as its own premise (\"There is a N that ...\"), drop the article from the rule, or write the phrase itself into the rule"
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

// Diagnostic only: an is/are mismatch must not hide a refused plural subject.
// Require both the antecedent tail and the instantiated consequent verbatim;
// this never relaxes unification or changes a verdict.
func (policy strictRules) pluralCopulaSlotMessage(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	for _, id := range c.Citations {
		for _, pair := range conditionalCandidatesPreservingCase(steps[id].Statement) {
			variable, tail, ok := strings.Cut(pair.left, " is ")
			if !ok || len(variable) != 1 || !strings.Contains("UVWXYZ", variable) {
				continue
			}
			for _, other := range c.Citations {
				if other == id {
					continue
				}
				value, predicate, ok := strings.Cut(scopeText(steps[other].Statement), " are ")
				if !ok || !strings.EqualFold(predicate, tail) || policy.singularValue(value) {
					continue
				}
				want := strings.ReplaceAll(pair.right, variable, value)
				if !strings.EqualFold(scopeText(want), scopeText(c.Statement)) {
					continue
				}
				return fmt.Sprintf("slot.value_singular: premise %d would bind %s to %q; a plural or bare value is refused for a term slot. Name the thing or write the phrase itself into the rule", other, variable, value)
			}
		}
	}
	return ""
}
