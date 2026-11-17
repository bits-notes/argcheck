package validation

// ParseNumberedArgumentLine parses a single numbered line without graph checks.
func ParseNumberedArgumentLine(line string, number int) (NumberedArgumentStep, Result) {
	return parseNumberedArgumentLine(line, number)
}

// CheckConclusionStep checks a conclusion against its indexed premises.
func CheckConclusionStep(step NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	return gateConclusionStep(step, steps)
}

// ComparisonClause returns the gate's normalized clause spelling.
func ComparisonClause(value string) string { return comparisonClause(value) }

// ConditionalClauses returns the gate's possible antecedent/consequent pairs.
func ConditionalClauses(statement string) [][2]string {
	var result [][2]string
	for _, pair := range conditionalCandidates(statement) {
		result = append(result, [2]string{pair.left, pair.right})
	}
	return result
}

// ConjunctiveAntecedentCandidates returns possible conjunction splits.
func ConjunctiveAntecedentCandidates(statement string, count int) [][]string {
	return conjunctiveAntecedentCandidates(statement, count)
}

// ParseDefaultNumberedArgument parses with the default syntax options.
func ParseDefaultNumberedArgument(text string) (NumberedArgument, Result) {
	return ParseNumberedArgument(text, numberedArgumentOptions{})
}
