package validation

import (
	"fmt"
	"regexp"
	"strings"
)

const clauseWrapper = "it is the case that"

var conditionalSplitPattern = regexp.MustCompile(`(?i),?\bthen\b,? `)

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
	{name: "universal instantiation", arity: 2, match: matchUniversalInstantiation},
}

func normalizeClause(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.Trim(value, ", ")
	if strings.HasSuffix(value, ".") {
		value = strings.TrimSuffix(value, ".")
	}
	return strings.ToLower(value)
}

func equalClauses(left, right string) bool {
	return stripClauseWrappers(left) == stripClauseWrappers(right)
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
	statement = normalizeClause(statement)
	if !strings.HasPrefix(statement, "if ") {
		return nil
	}
	remainder := strings.TrimPrefix(statement, "if ")
	indexes := conditionalSplitPattern.FindAllStringIndex(remainder, -1)
	result := make([]clausePair, 0, len(indexes))
	for _, index := range indexes {
		left := normalizeClause(remainder[:index[0]])
		right := normalizeClause(remainder[index[1]:])
		if left != "" && right != "" {
			result = append(result, clausePair{left: left, right: right})
		}
	}
	return result
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

func disjunctionCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	statement = strings.TrimPrefix(statement, "either ")
	return appendSplitCandidates(nil, statement, " or ", false)
}

func universalCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	if !strings.HasPrefix(statement, "every ") {
		return nil
	}
	return appendSplitCandidates(nil, strings.TrimPrefix(statement, "every "), " is ", false)
}

func instanceCandidates(statement string) []clausePair {
	return appendSplitCandidates(nil, normalizeClause(statement), " is ", false)
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
	outcome := schemaMatchOutcome{traction: true}
	for _, identity := range identities {
		if equalClauses(identity.left, identity.right) {
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

func betterOutcome(current, candidate schemaMatchOutcome) schemaMatchOutcome {
	if candidate.matched || candidate.score > current.score || (candidate.score == current.score && candidate.traction && !current.traction) {
		return candidate
	}
	return current
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
				candidate := evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d class", premises[0].Number), universal.left, fmt.Sprintf("premise %d predicate", premises[1].Number), instance.right},
					clauseIdentity{fmt.Sprintf("premise %d instance", premises[1].Number), instance.left, fmt.Sprintf("conclusion %d instance", conclusion.Number), result.left},
					clauseIdentity{fmt.Sprintf("premise %d predicate", premises[0].Number), universal.right, fmt.Sprintf("conclusion %d predicate", conclusion.Number), result.right},
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

func scoreUniversalPredicateSwap(premise, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, source := range universalCandidates(premise.Statement) {
		for _, target := range universalCandidates(conclusion.Statement) {
			candidate := evaluateIdentities(
				clauseIdentity{fmt.Sprintf("premise %d class", premise.Number), source.left, fmt.Sprintf("conclusion %d class", conclusion.Number), target.left},
				clauseIdentity{fmt.Sprintf("premise %d predicate", premise.Number), "is " + source.right, fmt.Sprintf("conclusion %d predicate", conclusion.Number), "is " + target.right},
			)
			best = betterOutcome(best, candidate)
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
