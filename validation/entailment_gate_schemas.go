package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const clauseWrapper = "it is the case that"

const variableArity = -1

var conditionalSplitPattern = regexp.MustCompile(`(?i),?\bthen\b,? `)
var separatedQuantifierPattern = regexp.MustCompile(`\b(every|no) thing\b`)

const schematicVariableMarker = "\u241f"

type clausePair struct {
	left  string
	right string
}

type entailmentSchema struct {
	name  string
	arity int
	match func([]NumberedArgumentStep, NumberedArgumentStep) schemaMatchOutcome
}

type schemaMatchOutcome struct {
	matched  bool
	score    int
	traction bool
	mismatch string
}

type clauseIdentity struct {
	leftLabel  string
	left       string
	rightLabel string
	right      string
}

var entailmentSchemas = []entailmentSchema{
	{name: "conjunctive modus ponens", arity: variableArity, match: matchConjunctiveModusPonens},
	{name: "modus ponens", arity: 2, match: matchModusPonens},
	{name: "modus tollens", arity: 2, match: matchModusTollens},
	{name: "hypothetical syllogism", arity: 2, match: matchHypotheticalSyllogism},
	{name: "conjunction introduction", arity: 2, match: matchConjunctionIntroduction},
	{name: "conjunction elimination", arity: 1, match: matchConjunctionElimination},
	{name: "disjunction introduction", arity: 1, match: matchDisjunctionIntroduction},
	{name: "disjunctive syllogism", arity: 2, match: matchDisjunctiveSyllogism},
	{name: "double negation", arity: 1, match: matchDoubleNegation},
	{name: "self-entailment", arity: 1, match: matchSelfEntailment},
	{name: "identity transitivity", arity: 2, match: matchIdentityTransitivity},
	{name: "universal syllogism", arity: 2, match: matchUniversalSyllogism},
	{name: "universal negative syllogism", arity: 2, match: matchUniversalNegativeSyllogism},
	{name: "universal negative instantiation", arity: 2, match: matchUniversalNegativeInstantiation},
	{name: "schematic instantiation", arity: 2, match: matchSchematicInstantiation},
	{name: "universal instantiation", arity: 2, match: matchUniversalInstantiation},
	{name: "necessity elimination", arity: 1, match: matchNecessityElimination},
}

func normalizeClause(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.Trim(value, ", ")
	if strings.HasSuffix(value, ".") {
		value = strings.TrimSuffix(value, ".")
	}
	value = strings.ToLower(value)
	return separatedQuantifierPattern.ReplaceAllString(value, "${1}thing")
}

func equalClauses(left, right string) bool {
	return comparisonClause(left) == comparisonClause(right)
}

func comparisonClause(value string) string {
	value = strings.ReplaceAll(stripClauseWrappers(value), ",", "")
	return strings.Join(strings.Fields(value), " ")
}

func stripClauseWrappers(value string) string {
	value = normalizeClause(value)
	for strings.HasPrefix(value, clauseWrapper) {
		remainder := value[len(clauseWrapper):]
		if remainder != "" && remainder[0] != ',' && remainder[0] != ' ' {
			break
		}
		value = strings.TrimLeft(remainder, ", ")
	}
	return value
}

func conditionalCandidates(statement string) []clausePair {
	return conditionalCandidatesWithNormalizer(statement, normalizeClause)
}

func conditionalCandidatesPreservingCase(statement string) []clausePair {
	return conditionalCandidatesWithNormalizer(statement, normalizeClausePreservingCase)
}

func conditionalCandidatesWithNormalizer(statement string, normalize func(string) string) []clausePair {
	statement = normalize(statement)
	if len(statement) < len("if ") || !strings.EqualFold(statement[:len("if ")], "if ") {
		return nil
	}
	remainder := statement[len("if "):]
	indexes := conditionalSplitPattern.FindAllStringIndex(remainder, -1)
	result := make([]clausePair, 0, len(indexes))
	for _, index := range indexes {
		left := normalize(remainder[:index[0]])
		right := normalize(remainder[index[1]:])
		if left != "" && right != "" {
			result = append(result, clausePair{left: left, right: right})
		}
	}
	return result
}

func normalizeClausePreservingCase(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.Trim(value, ", ")
	if strings.HasSuffix(value, ".") {
		value = strings.TrimSuffix(value, ".")
	}
	return value
}

func negationCandidates(statement string) []string {
	statement = normalizeClause(statement)
	prefix := "it is not the case that"
	if !strings.HasPrefix(statement, prefix) {
		return nil
	}
	remainder := statement[len(prefix):]
	if remainder != "" && remainder[0] != ',' && remainder[0] != ' ' {
		return nil
	}
	remainder = normalizeClause(remainder)
	if remainder == "" {
		return nil
	}
	return []string{remainder}
}

func conjunctionCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	result := make([]clausePair, 0)
	if strings.HasPrefix(statement, "it is the case both that") {
		remainder := statement[len("it is the case both that"):]
		if remainder == "" || remainder[0] == ',' || remainder[0] == ' ' {
			remainder = strings.TrimLeft(remainder, ", ")
			result = appendSplitCandidates(result, remainder, ", and that ", true)
		}
	}
	result = appendSplitCandidates(result, statement, ", and that ", true)
	result = appendSplitCandidates(result, statement, ", and ", false)
	result = appendSplitCandidates(result, statement, " and ", false)
	return result
}

func conjunctiveAntecedentCandidates(statement string) [][]string {
	statement = normalizeClause(statement)
	result := make([][]string, 0)
	separator := ", and " + clauseWrapper
	if strings.Contains(statement, separator) {
		parts := strings.Split(statement, separator)
		for index := 1; index < len(parts); index++ {
			parts[index] = clauseWrapper + " " + parts[index]
		}
		result = append(result, parts)
	}
	for _, conjunction := range conjunctionCandidates(statement) {
		result = append(result, []string{conjunction.left, conjunction.right})
	}
	return result
}

func disjunctionCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	statement = strings.TrimPrefix(statement, "either ")
	return appendSplitCandidates(nil, statement, " or ", false)
}

func universalCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	result := make([]clausePair, 0)
	if strings.HasPrefix(statement, "every ") {
		result = appendSplitCandidates(result, strings.TrimPrefix(statement, "every "), " is ", false)
	}
	if strings.HasPrefix(statement, "all ") {
		result = appendSplitCandidates(result, strings.TrimPrefix(statement, "all "), " are ", false)
	}
	for _, prefix := range []string{"everything ", "everyone "} {
		if strings.HasPrefix(statement, prefix) {
			result = appendSplitCandidates(result, strings.TrimPrefix(statement, prefix), " is ", false)
		}
	}
	return result
}

// universalNegativeCandidates parses E-propositions ("no S is P") into
// subject/predicate pairs. The regimented register stipulates wide-scope
// negation: "every S is not P" and bare generics ("Ss are not Ps") read as
// universal negatives, never as "not every S is P".
func universalNegativeCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	result := make([]clausePair, 0)
	for _, prefix := range []string{"no ", "nothing ", "no one "} {
		if strings.HasPrefix(statement, prefix) {
			remainder := strings.TrimPrefix(statement, prefix)
			result = appendSplitCandidates(result, remainder, " is ", false)
			result = appendSplitCandidates(result, remainder, " are ", false)
		}
	}
	for _, prefix := range []string{"every ", "all ", "everything ", "everyone "} {
		if strings.HasPrefix(statement, prefix) {
			remainder := strings.TrimPrefix(statement, prefix)
			result = appendSplitCandidates(result, remainder, " is not ", false)
			result = appendSplitCandidates(result, remainder, " am not ", false)
			result = appendSplitCandidates(result, remainder, " are not ", false)
		}
	}
	bare := true
	for _, prefix := range []string{"some ", "most ", "many ", "few ", "several ", "at least ", "at most ", "there ", "if ", "no ", "not ", "it is not the case"} {
		if strings.HasPrefix(statement, prefix) {
			bare = false
			break
		}
	}
	if bare {
		result = appendSplitCandidates(result, statement, " are not ", false)
		result = appendSplitCandidates(result, statement, " is not ", false)
		result = appendSplitCandidates(result, statement, " am not ", false)
	}
	return result
}

func negativeInstanceCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	for _, prefix := range []string{"every ", "all ", "everything ", "everyone ", "no ", "nothing ", "no one "} {
		if strings.HasPrefix(statement, prefix) {
			return nil
		}
	}
	result := make([]clausePair, 0)
	for _, separator := range []string{" is not ", " am not ", " are not "} {
		result = appendSplitCandidates(result, statement, separator, false)
	}
	return result
}

func necessityCandidates(statement string) []string {
	statement = normalizeClause(statement)
	for _, prefix := range []string{"necessarily, ", "necessarily "} {
		if strings.HasPrefix(statement, prefix) {
			remainder := normalizeClause(strings.TrimPrefix(statement, prefix))
			if remainder != "" {
				return []string{remainder}
			}
		}
	}
	return nil
}

func equalTerms(left, right string) bool {
	leftCandidates := termCandidates(left)
	rightCandidates := termCandidates(right)
	for _, leftCandidate := range leftCandidates {
		for _, rightCandidate := range rightCandidates {
			if leftCandidate == rightCandidate {
				return true
			}
		}
	}
	return false
}

func termCandidates(value string) []string {
	value = normalizeClause(value)
	value = strings.ReplaceAll(value, ",", "")
	value = strings.Join(strings.Fields(value), " ")
	for _, copula := range []string{"is ", "am ", "are "} {
		if strings.HasPrefix(value, copula) {
			value = strings.TrimPrefix(value, copula)
			break
		}
	}
	for _, article := range []string{"a ", "an ", "the "} {
		if strings.HasPrefix(value, article) {
			value = strings.TrimPrefix(value, article)
			break
		}
	}
	result := []string{value}
	if strings.HasSuffix(value, "s") && len(value) > 1 {
		result = append(result, strings.TrimSuffix(value, "s"))
	}
	if strings.HasSuffix(value, "es") && len(value) > 2 {
		result = append(result, strings.TrimSuffix(value, "es"))
	}
	return result
}

func instanceCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	result := make([]clausePair, 0)
	for _, separator := range []string{" is ", " am ", " are "} {
		result = appendSplitCandidates(result, statement, separator, false)
	}
	return result
}

func replaceSchematicVariables(value, replacement string) (string, int) {
	runes := []rune(value)
	var result strings.Builder
	count := 0
	for index, current := range runes {
		if current == 'X' &&
			(index == 0 || !unicode.IsLetter(runes[index-1])) &&
			(index == len(runes)-1 || !unicode.IsLetter(runes[index+1])) {
			result.WriteString(replacement)
			count++
			continue
		}
		result.WriteRune(current)
	}
	return result.String(), count
}

func schematicTemplate(value string) (string, int) {
	replaced, count := replaceSchematicVariables(value, schematicVariableMarker)
	return comparisonClause(replaced), count
}

func deriveSchematicSubstitution(antecedentTemplate, instance string) (string, bool) {
	segments := strings.Split(antecedentTemplate, schematicVariableMarker)
	if len(segments) < 2 {
		return "", false
	}
	instance = comparisonClause(instance)
	if !strings.HasPrefix(instance, segments[0]) {
		return "", false
	}
	remainder := instance[len(segments[0]):]

	var substitution string
	if len(segments) == 2 && segments[1] == "" {
		substitution = remainder
	} else {
		nextLiteral := segments[1]
		if nextLiteral == "" {
			return "", false
		}
		nextIndex := strings.Index(remainder, nextLiteral)
		if nextIndex < 0 {
			return "", false
		}
		substitution = remainder[:nextIndex]
	}
	if substitution == "" {
		return "", false
	}
	return substitution, true
}

func identityCandidates(statement string) []clausePair {
	return appendSplitCandidates(nil, normalizeClause(statement), " is identical to ", false)
}

func appendSplitCandidates(result []clausePair, statement, separator string, stripRightThat bool) []clausePair {
	start := 0
	for {
		index := strings.Index(statement[start:], separator)
		if index < 0 {
			return result
		}
		index += start
		left := normalizeClause(statement[:index])
		right := normalizeClause(statement[index+len(separator):])
		if stripRightThat {
			right = strings.TrimPrefix(right, "that ")
		}
		if left != "" && right != "" {
			result = append(result, clausePair{left: left, right: right})
		}
		start = index + len(separator)
	}
}

func evaluateIdentities(identities ...clauseIdentity) schemaMatchOutcome {
	return evaluateIdentityComparisons(equalClauses, identities...)
}

func evaluateTermIdentities(identities ...clauseIdentity) schemaMatchOutcome {
	return evaluateIdentityComparisons(equalTerms, identities...)
}

func evaluateIdentityComparisons(equal func(string, string) bool, identities ...clauseIdentity) schemaMatchOutcome {
	outcome := schemaMatchOutcome{traction: true}
	for _, identity := range identities {
		if equal(identity.left, identity.right) {
			outcome.score++
			continue
		}
		if outcome.mismatch == "" {
			outcome.mismatch = fmt.Sprintf(
				"%s %q does not match %s %q verbatim",
				identity.leftLabel,
				normalizeClause(identity.left),
				identity.rightLabel,
				normalizeClause(identity.right),
			)
		}
	}
	outcome.matched = outcome.score == len(identities)
	return outcome
}

func appendOutcome(current *schemaMatchOutcome, addition schemaMatchOutcome) {
	current.score += addition.score
	current.traction = current.traction || addition.traction
	if current.mismatch == "" {
		current.mismatch = addition.mismatch
	}
}

func betterOutcome(current, candidate schemaMatchOutcome) schemaMatchOutcome {
	if candidate.matched || candidate.score > current.score || (candidate.score == current.score && candidate.traction && !current.traction) {
		return candidate
	}
	return current
}

func matchConjunctiveModusPonens(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	if len(premises) < 3 {
		return schemaMatchOutcome{}
	}

	best := schemaMatchOutcome{}
	for conditionalIndex, premise := range premises {
		for _, conditional := range conditionalCandidates(premise.Statement) {
			conjunctiveCandidates := conjunctiveAntecedentCandidates(conditional.left)
			if len(conjunctiveCandidates) == 0 {
				candidate := schemaMatchOutcome{
					traction: true,
					mismatch: fmt.Sprintf(
						"premise %d antecedent %q does not split into two or more conjuncts",
						premise.Number,
						normalizeClause(conditional.left),
					),
				}
				best = betterOutcome(best, candidate)
				continue
			}

			for _, conjuncts := range conjunctiveCandidates {
				candidate := evaluateConjunctiveModusPonensCandidate(
					premises,
					conditionalIndex,
					conditional,
					conjuncts,
					conclusion,
				)
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

func evaluateConjunctiveModusPonensCandidate(
	premises []NumberedArgumentStep,
	conditionalIndex int,
	conditional clausePair,
	conjuncts []string,
	conclusion NumberedArgumentStep,
) schemaMatchOutcome {
	conditionalPremise := premises[conditionalIndex]
	outcome := evaluateIdentities(clauseIdentity{
		fmt.Sprintf("premise %d consequent", conditionalPremise.Number), conditional.right,
		fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement,
	})
	used := make([]bool, len(premises))
	used[conditionalIndex] = true
	usedPremiseNumbers := map[int]bool{conditionalPremise.Number: true}

	for _, conjunct := range conjuncts {
		matchedIndex := -1
		for premiseIndex, premise := range premises {
			if used[premiseIndex] || usedPremiseNumbers[premise.Number] || !equalClauses(conjunct, premise.Statement) {
				continue
			}
			matchedIndex = premiseIndex
			break
		}
		if matchedIndex >= 0 {
			used[matchedIndex] = true
			usedPremiseNumbers[premises[matchedIndex].Number] = true
			outcome.score++
			continue
		}
		if outcome.mismatch == "" {
			unused := make([]NumberedArgumentStep, 0, len(premises))
			for premiseIndex, premise := range premises {
				if !used[premiseIndex] {
					unused = append(unused, premise)
				}
			}
			nearest := nearestPremiseToClause(unused, 0, conjunct)
			if nearest.Number == 0 {
				outcome.mismatch = fmt.Sprintf(
					"premise %d antecedent conjunct %q is not matched verbatim by any unused cited premise",
					conditionalPremise.Number,
					normalizeClause(conjunct),
				)
			} else {
				outcome.mismatch = fmt.Sprintf(
					"premise %d antecedent conjunct %q does not match cited premise %d %q verbatim",
					conditionalPremise.Number,
					normalizeClause(conjunct),
					nearest.Number,
					normalizeClause(nearest.Statement),
				)
			}
		}
	}

	for premiseIndex, premise := range premises {
		if used[premiseIndex] {
			continue
		}
		if outcome.mismatch == "" {
			outcome.mismatch = fmt.Sprintf(
				"cited premise %d %q is unused by premise %d conjunctive antecedent",
				premise.Number,
				normalizeClause(premise.Statement),
				conditionalPremise.Number,
			)
		}
	}

	outcome.matched = outcome.mismatch == "" && len(conjuncts)+1 == len(premises)
	return outcome
}

func matchModusPonens(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, conditional := range conditionalCandidates(premises[0].Statement) {
		candidate := evaluateIdentities(
			clauseIdentity{fmt.Sprintf("premise %d antecedent", premises[0].Number), conditional.left, fmt.Sprintf("premise %d", premises[1].Number), premises[1].Statement},
			clauseIdentity{fmt.Sprintf("premise %d consequent", premises[0].Number), conditional.right, fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement},
		)
		best = betterOutcome(best, candidate)
		if best.matched {
			return best
		}
	}
	return best
}

func matchModusTollens(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	negatedPremises := negationCandidates(premises[1].Statement)
	negatedConclusions := negationCandidates(conclusion.Statement)
	for _, conditional := range conditionalCandidates(premises[0].Statement) {
		for _, negatedPremise := range negatedPremises {
			for _, negatedConclusion := range negatedConclusions {
				candidate := evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d consequent", premises[0].Number), conditional.right, fmt.Sprintf("premise %d negated clause", premises[1].Number), negatedPremise},
					clauseIdentity{fmt.Sprintf("premise %d antecedent", premises[0].Number), conditional.left, fmt.Sprintf("conclusion %d negated clause", conclusion.Number), negatedConclusion},
				)
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

func matchHypotheticalSyllogism(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, first := range conditionalCandidates(premises[0].Statement) {
		for _, second := range conditionalCandidates(premises[1].Statement) {
			for _, result := range conditionalCandidates(conclusion.Statement) {
				candidate := evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d consequent", premises[0].Number), first.right, fmt.Sprintf("premise %d antecedent", premises[1].Number), second.left},
					clauseIdentity{fmt.Sprintf("premise %d antecedent", premises[0].Number), first.left, fmt.Sprintf("conclusion %d antecedent", conclusion.Number), result.left},
					clauseIdentity{fmt.Sprintf("premise %d consequent", premises[1].Number), second.right, fmt.Sprintf("conclusion %d consequent", conclusion.Number), result.right},
				)
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

func matchConjunctionIntroduction(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, conjunction := range conjunctionCandidates(conclusion.Statement) {
		candidate := evaluateIdentities(
			clauseIdentity{fmt.Sprintf("premise %d", premises[0].Number), premises[0].Statement, fmt.Sprintf("conclusion %d left conjunct", conclusion.Number), conjunction.left},
			clauseIdentity{fmt.Sprintf("premise %d", premises[1].Number), premises[1].Statement, fmt.Sprintf("conclusion %d right conjunct", conclusion.Number), conjunction.right},
		)
		best = betterOutcome(best, candidate)
		if best.matched {
			return best
		}
	}
	return best
}

func matchConjunctionElimination(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, conjunction := range conjunctionCandidates(premises[0].Statement) {
		for _, part := range []struct {
			label string
			value string
		}{{"left conjunct", conjunction.left}, {"right conjunct", conjunction.right}} {
			candidate := evaluateIdentities(clauseIdentity{
				fmt.Sprintf("premise %d %s", premises[0].Number, part.label), part.value,
				fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement,
			})
			best = betterOutcome(best, candidate)
			if best.matched {
				return best
			}
		}
	}
	return best
}

func matchDisjunctionIntroduction(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, disjunction := range disjunctionCandidates(conclusion.Statement) {
		for _, part := range []struct {
			label string
			value string
		}{{"left disjunct", disjunction.left}, {"right disjunct", disjunction.right}} {
			candidate := evaluateIdentities(clauseIdentity{
				fmt.Sprintf("premise %d", premises[0].Number), premises[0].Statement,
				fmt.Sprintf("conclusion %d %s", conclusion.Number, part.label), part.value,
			})
			best = betterOutcome(best, candidate)
			if best.matched {
				return best
			}
		}
	}
	return best
}

func matchDisjunctiveSyllogism(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, disjunction := range disjunctionCandidates(premises[0].Statement) {
		for _, negated := range negationCandidates(premises[1].Statement) {
			variants := [][2]string{{disjunction.left, disjunction.right}, {disjunction.right, disjunction.left}}
			for _, variant := range variants {
				candidate := evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d negated clause", premises[1].Number), negated, fmt.Sprintf("premise %d disjunct", premises[0].Number), variant[0]},
					clauseIdentity{fmt.Sprintf("premise %d remaining disjunct", premises[0].Number), variant[1], fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement},
				)
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

func matchDoubleNegation(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, once := range negationCandidates(premises[0].Statement) {
		for _, twice := range negationCandidates(once) {
			candidate := evaluateIdentities(clauseIdentity{
				fmt.Sprintf("premise %d doubly negated clause", premises[0].Number), twice,
				fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement,
			})
			best = betterOutcome(best, candidate)
		}
	}
	return best
}

func matchSelfEntailment(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	return evaluateIdentities(clauseIdentity{
		fmt.Sprintf("premise %d", premises[0].Number), premises[0].Statement,
		fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement,
	})
}

func matchIdentityTransitivity(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, first := range identityCandidates(premises[0].Statement) {
		for _, second := range identityCandidates(premises[1].Statement) {
			for _, result := range identityCandidates(conclusion.Statement) {
				candidate := evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d middle term", premises[0].Number), first.right, fmt.Sprintf("premise %d middle term", premises[1].Number), second.left},
					clauseIdentity{fmt.Sprintf("premise %d first term", premises[0].Number), first.left, fmt.Sprintf("conclusion %d first term", conclusion.Number), result.left},
					clauseIdentity{fmt.Sprintf("premise %d final term", premises[1].Number), second.right, fmt.Sprintf("conclusion %d final term", conclusion.Number), result.right},
				)
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

func matchUniversalInstantiation(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, universal := range universalCandidates(premises[0].Statement) {
		for _, instance := range instanceCandidates(premises[1].Statement) {
			for _, result := range instanceCandidates(conclusion.Statement) {
				candidate := schemaMatchOutcome{}
				appendOutcome(&candidate, evaluateTermIdentities(
					clauseIdentity{fmt.Sprintf("premise %d class", premises[0].Number), universal.left, fmt.Sprintf("premise %d predicate", premises[1].Number), instance.right},
				))
				appendOutcome(&candidate, evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d instance", premises[1].Number), instance.left, fmt.Sprintf("conclusion %d instance", conclusion.Number), result.left},
				))
				appendOutcome(&candidate, evaluateTermIdentities(
					clauseIdentity{fmt.Sprintf("premise %d predicate", premises[0].Number), universal.right, fmt.Sprintf("conclusion %d predicate", conclusion.Number), result.right},
				))
				candidate.matched = candidate.score == 3
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

func scoreUniversalPredicateSwap(premise, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, source := range universalCandidates(premise.Statement) {
		for _, target := range universalCandidates(conclusion.Statement) {
			candidate := evaluateTermIdentities(
				clauseIdentity{fmt.Sprintf("premise %d class", premise.Number), source.left, fmt.Sprintf("conclusion %d class", conclusion.Number), target.left},
				clauseIdentity{fmt.Sprintf("premise %d predicate", premise.Number), "is " + source.right, fmt.Sprintf("conclusion %d predicate", conclusion.Number), "is " + target.right},
			)
			best = betterOutcome(best, candidate)
		}
	}
	return best
}

func matchUniversalSyllogism(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, first := range universalCandidates(premises[0].Statement) {
		for _, second := range universalCandidates(premises[1].Statement) {
			for _, result := range universalCandidates(conclusion.Statement) {
				candidate := evaluateTermIdentities(
					clauseIdentity{fmt.Sprintf("premise %d class", premises[0].Number), first.left, fmt.Sprintf("conclusion %d class", conclusion.Number), result.left},
					clauseIdentity{fmt.Sprintf("premise %d predicate", premises[0].Number), first.right, fmt.Sprintf("premise %d class", premises[1].Number), second.left},
					clauseIdentity{fmt.Sprintf("premise %d predicate", premises[1].Number), second.right, fmt.Sprintf("conclusion %d predicate", conclusion.Number), result.right},
				)
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

// matchUniversalNegativeSyllogism certifies the valid A/E second- and
// first-figure syllogisms (Celarent, Cesare, Camestres, Camenes) in one form.
// Because E-propositions convert (no A is B entails no B is A), the E premise
// and E conclusion are matched in either orientation; the middle term must be
// the A premise's predicate, never its subject.
func matchUniversalNegativeSyllogism(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, universal := range universalCandidates(premises[0].Statement) {
		for _, negative := range universalNegativeCandidates(premises[1].Statement) {
			for _, result := range universalNegativeCandidates(conclusion.Statement) {
				negativeOrientations := [][2]string{{negative.left, negative.right}, {negative.right, negative.left}}
				resultOrientations := [][2]string{{result.left, result.right}, {result.right, result.left}}
				for _, negativeTerms := range negativeOrientations {
					for _, resultTerms := range resultOrientations {
						candidate := evaluateTermIdentities(
							clauseIdentity{fmt.Sprintf("premise %d predicate", premises[0].Number), universal.right, fmt.Sprintf("premise %d middle term", premises[1].Number), negativeTerms[0]},
							clauseIdentity{fmt.Sprintf("premise %d class", premises[0].Number), universal.left, fmt.Sprintf("conclusion %d subject term", conclusion.Number), resultTerms[0]},
							clauseIdentity{fmt.Sprintf("premise %d excluded term", premises[1].Number), negativeTerms[1], fmt.Sprintf("conclusion %d excluded term", conclusion.Number), resultTerms[1]},
						)
						best = betterOutcome(best, candidate)
						if best.matched {
							return best
						}
					}
				}
			}
		}
	}
	return best
}

// matchUniversalNegativeInstantiation certifies the fixed-orientation
// disjoint-class inference a is F; no G is F; therefore a is not G. Unlike
// universal negative syllogism, neither conversion nor term morphology is
// permitted: the shared predicate, instance, and excluded class are compared
// as clauses. The conclusion alone may add an indefinite article before G.
func matchUniversalNegativeInstantiation(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, instance := range instanceCandidates(premises[0].Statement) {
		for _, negative := range universalNegativeCandidates(premises[1].Statement) {
			for _, result := range negativeInstanceCandidates(conclusion.Statement) {
				candidate := schemaMatchOutcome{}
				appendOutcome(&candidate, evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d predicate", premises[0].Number), instance.right, fmt.Sprintf("premise %d predicate", premises[1].Number), negative.right},
					clauseIdentity{fmt.Sprintf("premise %d instance", premises[0].Number), instance.left, fmt.Sprintf("conclusion %d instance", conclusion.Number), result.left},
				))
				appendOutcome(&candidate, evaluateIdentityComparisons(
					equalClausesWithOptionalIndefiniteArticle,
					clauseIdentity{fmt.Sprintf("premise %d restricted class", premises[1].Number), negative.left, fmt.Sprintf("conclusion %d excluded class", conclusion.Number), result.right},
				))
				candidate.matched = candidate.score == 3
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
		}
	}
	return best
}

func equalClausesWithOptionalIndefiniteArticle(left, right string) bool {
	return stripOptionalIndefiniteArticle(left) == stripOptionalIndefiniteArticle(right)
}

func stripOptionalIndefiniteArticle(value string) string {
	value = comparisonClause(value)
	for _, article := range []string{"a ", "an "} {
		if strings.HasPrefix(value, article) {
			return strings.TrimPrefix(value, article)
		}
	}
	return value
}

// matchSchematicInstantiation certifies If A(X), then B(X); A(t);
// therefore B(t), where X is a standalone uppercase variable in both sides
// and the same non-empty substitution verifies every occurrence.
func matchSchematicInstantiation(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, conditional := range conditionalCandidatesPreservingCase(premises[0].Statement) {
		antecedentTemplate, antecedentVariables := schematicTemplate(conditional.left)
		consequentTemplate, consequentVariables := schematicTemplate(conditional.right)
		candidate := schemaMatchOutcome{traction: true}
		if antecedentVariables == 0 || consequentVariables == 0 {
			switch {
			case antecedentVariables == 0 && consequentVariables == 0:
				candidate.mismatch = fmt.Sprintf("premise %d conditional has no standalone uppercase X in either side", premises[0].Number)
			case antecedentVariables == 0:
				candidate.score = 1
				candidate.mismatch = fmt.Sprintf("premise %d antecedent has no standalone uppercase X", premises[0].Number)
			default:
				candidate.score = 1
				candidate.mismatch = fmt.Sprintf("premise %d consequent has no standalone uppercase X", premises[0].Number)
			}
			best = betterOutcome(best, candidate)
			continue
		}

		substitution, ok := deriveSchematicSubstitution(antecedentTemplate, premises[1].Statement)
		if !ok {
			candidate.score = 1
			candidate.mismatch = fmt.Sprintf(
				"premise %d does not provide a non-empty substitution for premise %d antecedent template",
				premises[1].Number,
				premises[0].Number,
			)
			best = betterOutcome(best, candidate)
			continue
		}

		candidate = evaluateIdentities(
			clauseIdentity{
				fmt.Sprintf("premise %d instantiated antecedent", premises[0].Number),
				strings.ReplaceAll(antecedentTemplate, schematicVariableMarker, substitution),
				fmt.Sprintf("premise %d", premises[1].Number),
				premises[1].Statement,
			},
			clauseIdentity{
				fmt.Sprintf("premise %d instantiated consequent", premises[0].Number),
				strings.ReplaceAll(consequentTemplate, schematicVariableMarker, substitution),
				fmt.Sprintf("conclusion %d", conclusion.Number),
				conclusion.Statement,
			},
		)
		best = betterOutcome(best, candidate)
		if best.matched {
			return best
		}
	}
	return best
}

// matchNecessityElimination certifies "Necessarily, P. Therefore, P." — the T
// axiom, sound for the alethic reading of "necessarily" the regimented
// register stipulates.
func matchNecessityElimination(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, stripped := range necessityCandidates(premises[0].Statement) {
		candidate := evaluateIdentities(clauseIdentity{
			fmt.Sprintf("premise %d de-necessitated clause", premises[0].Number), stripped,
			fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement,
		})
		best = betterOutcome(best, candidate)
		if best.matched {
			return best
		}
	}
	return best
}

func conjunctiveAntecedentModusPonensFailure(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) (string, bool) {
	for _, premise := range premises {
		for _, conditional := range conditionalCandidates(premise.Statement) {
			if !equalClauses(conditional.right, conclusion.Statement) {
				continue
			}
			matchedAntecedent := false
			for _, cited := range premises {
				if cited.Number != premise.Number && equalClauses(conditional.left, cited.Statement) {
					matchedAntecedent = true
					break
				}
			}
			if matchedAntecedent {
				continue
			}
			nearest := nearestPremiseToClause(premises, premise.Number, conditional.left)
			if nearest.Number == 0 {
				return fmt.Sprintf(
					"premise %d antecedent %q is not matched verbatim by any cited premise",
					premise.Number,
					normalizeClause(conditional.left),
				), true
			}
			return fmt.Sprintf(
				"premise %d antecedent %q is not matched verbatim by any cited premise; nearest is step %d %q",
				premise.Number,
				normalizeClause(conditional.left),
				nearest.Number,
				normalizeClause(nearest.Statement),
			), true
		}
	}
	return "", false
}

func nearestPremiseToClause(premises []NumberedArgumentStep, excluded int, clause string) NumberedArgumentStep {
	clauseWords := wordSet(stripClauseWrappers(clause))
	bestScore := -1
	best := NumberedArgumentStep{}
	for _, premise := range premises {
		if premise.Number == excluded {
			continue
		}
		score := 0
		for word := range wordSet(stripClauseWrappers(premise.Statement)) {
			if clauseWords[word] {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			best = premise
		}
	}
	return best
}

func wordSet(value string) map[string]bool {
	words := make(map[string]bool)
	for _, word := range strings.Fields(value) {
		words[word] = true
	}
	return words
}
