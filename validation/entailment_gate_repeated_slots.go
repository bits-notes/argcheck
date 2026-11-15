package validation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// v30: a schematic slot that occurs more than once must bind one fixed thing.
//
// A rule such as "If X is wise and X is rich, then X is happy" is read as
// universally closed over X. Instantiating X with a referring term (a name, a
// definite description) is sound. Substituting a quantifier phrase is sound only
// when the quantifier can be pulled outside the whole rule without changing what
// each occurrence is about, and that depends on how often the slot occurs:
//
//   - Repeated in the antecedent: an existential, indefinite or disjunctive
//     binding lets each occurrence pick a different witness. "Some philosopher is
//     wise" and "Some philosopher is rich" do not give one philosopher who is
//     both. A universal binding is sound here (every philosopher is wise and
//     every philosopher is rich, so every philosopher is happy).
//   - Repeated in the consequent: a universal or disjunctive binding spreads one
//     instance into many. "Every philosopher admires himself" is not "every
//     philosopher admires every philosopher". An existential binding is sound
//     here (one witness that admires itself is a philosopher who admires some
//     philosopher).
//
// A slot that occurs once on each side keeps the v28.1 behaviour: every and some
// bind, scope operators do not.

var repeatedSlotUniversal = regexp.MustCompile(`^(?:every|each|all|any|everyone|everybody|everything|anyone|anybody|anything|whoever|whatever|whichever)\b`)

var repeatedSlotExistential = regexp.MustCompile(`^(?:some|someone|somebody|something|somewhere|a|an|another|one|two|three|four|five|six|seven|eight|nine|ten|several|many|most|certain|various|numerous|exactly|at least|more than|a few|a couple|lots of|plenty of)\b`)

var repeatedSlotDisjunctSingular = regexp.MustCompile(`^(?:(?:a|an|some|one|the|this|that)\s+\S+(?:\s+\S+){0,2}|\S+)$`)

var repeatedSlotRelativeThenVerb = regexp.MustCompile(`\b(?:that|who|which|whom|whose|where|when)\b`)

// A disjunctive term is a short coordination of singular terms ("Socrates or
// Plato", "either a cat or a dog"). Longer phrases that merely contain "or"
// (bare plurals, gerund clauses) are left to their generic reading.
func repeatedSlotDisjunctive(lower string) bool {
	if strings.HasPrefix(lower, "the ") {
		return false
	}
	trimmed := strings.TrimPrefix(lower, "either ")
	parts := strings.Split(trimmed, " or ")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		part = strings.TrimSuffix(strings.TrimSpace(part), ",")
		if !repeatedSlotDisjunctSingular.MatchString(part) {
			return false
		}
	}
	return true
}

// A binding that is itself a complete clause ("some systems are exposed")
// stands for a proposition, and a repeated propositional slot is sound.
func repeatedSlotClause(lower string) bool {
	words := strings.Fields(lower)
	for index, word := range words {
		if schematicBindingClauseVerbs[word] {
			return !repeatedSlotRelativeThenVerb.MatchString(strings.Join(words[:index], " "))
		}
	}
	return false
}

func repeatedSlotBindingKind(binding string) (universal, existential, disjunctive bool) {
	lower := strings.ToLower(strings.TrimSpace(binding))
	if repeatedSlotClause(lower) {
		return false, false, false
	}
	universal = repeatedSlotUniversal.MatchString(lower)
	existential = repeatedSlotExistential.MatchString(lower)
	disjunctive = repeatedSlotDisjunctive(lower)
	return universal, existential, disjunctive
}

func schematicMarkerCounts(template string) map[rune]int {
	counts := map[rune]int{}
	for _, r := range template {
		if v, ok := schematicVariableForMarker(r); ok {
			counts[v]++
		}
	}
	return counts
}

// repeatedSlotMismatch returns "" when every repeated slot binds soundly, and
// otherwise a reason naming the slot. Templates carry schematic markers.
func repeatedSlotMismatch(antecedentTemplate, consequentTemplate string, substitution schematicSubstitution) string {
	antecedent := schematicMarkerCounts(antecedentTemplate)
	consequent := schematicMarkerCounts(consequentTemplate)
	variables := make([]rune, 0, len(substitution))
	for variable := range substitution {
		variables = append(variables, variable)
	}
	sort.Slice(variables, func(i, j int) bool { return variables[i] < variables[j] })
	for _, variable := range variables {
		binding := substitution[variable]
		universal, existential, disjunctive := repeatedSlotBindingKind(binding)
		if antecedent[variable] >= 2 && (existential || disjunctive) {
			return fmt.Sprintf("schematic variable %c occurs %d times in the antecedent but binds %q, which does not pick out one fixed thing; separate occurrences may concern different witnesses", variable, antecedent[variable], binding)
		}
		if consequent[variable] >= 2 && (universal || disjunctive) {
			return fmt.Sprintf("schematic variable %c occurs %d times in the consequent but binds %q; one instance does not license every combination", variable, consequent[variable], binding)
		}
	}
	return ""
}
