package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const clauseWrapper = "it is the case that"

const variableArity = -1
const variableArityAtLeastTwo = -2

var conditionalSplitPattern = regexp.MustCompile(`(?i),?\bthen\b,? `)
var separatedQuantifierPattern = regexp.MustCompile(`\b(every|no) thing\b`)

const (
	schematicVariableMarkerX = '\u241d'
	schematicVariableMarkerY = '\u241e'
	schematicVariableMarkerZ = '\u241f'
)

var schematicVariableMarkers = map[rune]rune{
	'X': schematicVariableMarkerX,
	'Y': schematicVariableMarkerY,
	'Z': schematicVariableMarkerZ,
}

type schematicVariableCounts map[rune]int

type schematicSubstitution map[rune]string

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
	{name: "schematic conjunctive modus ponens", arity: variableArity, match: matchSchematicConjunctiveModusPonens},
	{name: "conjunctive modus ponens", arity: variableArity, match: matchConjunctiveModusPonens},
	{name: "modus ponens", arity: 2, match: matchModusPonens},
	{name: "modus tollens", arity: 2, match: matchModusTollens},
	{name: "hypothetical syllogism", arity: 2, match: matchHypotheticalSyllogism},
	{name: "biconditional introduction", arity: 2, match: matchBiconditionalIntroduction},
	{name: "biconditional elimination", arity: 1, match: matchBiconditionalElimination},
	{name: "biconditional modus ponens", arity: 2, match: matchBiconditionalModusPonens},
	{name: "conjunction introduction", arity: 2, match: matchConjunctionIntroduction},
	{name: "n-ary conjunction introduction", arity: variableArity, match: matchNaryConjunctionIntroduction},
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
	{name: "non-copular universal instantiation", arity: 2, match: matchNonCopularUniversalInstantiation},
	{name: "universal instantiation", arity: 2, match: matchUniversalInstantiation},
	{name: "existential generalization", arity: 1, match: matchExistentialGeneralization},
	{name: "existential introduction", arity: variableArityAtLeastTwo, match: matchExistentialIntroduction},
	{name: "existential counterexample", arity: 1, match: matchExistentialCounterexample},
	{name: "necessity elimination", arity: 1, match: matchNecessityElimination},
	{name: "existential syllogism", arity: 2, match: matchExistentialSyllogism},
	{name: "existential negative syllogism", arity: 2, match: matchExistentialNegativeSyllogism},
	{name: "schematic existential introduction", arity: 1, match: matchSchematicExistentialIntroduction},
	{name: "identity substitution", arity: 2, match: matchIdentitySubstitution},
	{name: "constructive dilemma", arity: variableArity, match: matchConstructiveDilemma},
	{name: "simple dilemma", arity: variableArity, match: matchSimpleDilemma},
	{name: "destructive dilemma", arity: variableArity, match: matchDestructiveDilemma},
	// v13 forms follow all prior entries to preserve reported form names.
	{name: "contraposition", arity: 1, match: matchContraposition},
	{name: "De Morgan", arity: 1, match: matchDeMorgan},
	{name: "commutation", arity: 1, match: matchCommutation},
	{name: "exportation", arity: 1, match: matchExportation},
	{name: "importation", arity: 1, match: matchImportation},
	{name: "material implication", arity: 1, match: matchMaterialImplication},
	{name: "distribution", arity: 1, match: matchDistribution},
	{name: "absorption", arity: 1, match: matchAbsorption},
	{name: "double negation introduction", arity: 1, match: matchDoubleNegationIntroduction},
	{name: "biconditional modus tollens", arity: 2, match: matchBiconditionalModusTollens},
	{name: "biconditional chaining", arity: 2, match: matchBiconditionalChaining},
	{name: "biconditional commutation", arity: 1, match: matchBiconditionalCommutation},
	{name: "negated conditional", arity: 1, match: matchNegatedConditional},
	{name: "identity symmetry", arity: 1, match: matchIdentitySymmetry},
	{name: "quantifier negation", arity: 1, match: matchQuantifierNegation},
	{name: "conversion", arity: 1, match: matchConversion},
	{name: "universal modus tollens", arity: 2, match: matchUniversalModusTollens},
	{name: "possibility introduction", arity: 1, match: matchPossibilityIntroduction},
	{name: "contradiction introduction", arity: 2, match: matchContradiction},
	// v15 forms follow every existing one-step entry.
	{name: "identity reflexivity", arity: 1, match: matchIdentityReflexivity},
	{name: "modal duality", arity: 1, match: matchModalDuality},
	{name: "modal distribution", arity: 1, match: matchModalDistribution},
}

// normalizeClause is the gate's only clause normalization: contractions, case,
// whitespace and edge punctuation. Never extend it toward paraphrase matching
// (synonyms, stemming, similarity, LLM equivalence). Equating two wordings is a
// premise the argument must state visibly. See the "Non-goal: paraphrase
// matching" section of repo_wiki/concepts/arguments/entailment-gate-spec.md.
func normalizeClause(value string) string {
	value = expandGateContractions(value)
	value = strings.Join(strings.Fields(value), " ")
	value = strings.Trim(value, ", ")
	if strings.HasSuffix(value, ".") {
		value = strings.TrimSuffix(value, ".")
	}
	value = normalizeGateConditional(value)
	value = expandGateElidedDisjunction(value)
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

func biconditionalCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	const connective = " if and only if "
	if strings.Count(statement, connective) != 1 {
		return nil
	}
	index := strings.Index(statement, connective)
	left := normalizeClause(statement[:index])
	right := normalizeClause(statement[index+len(connective):])
	if left == "" || right == "" {
		return nil
	}
	// The infix alone cannot disambiguate a top-level biconditional from one
	// embedded under a report, attribution, or larger punctuated clause. Reject
	// those surface boundaries rather than risk certifying the embedded reading.
	for _, clause := range []string{stripClauseWrappers(left), stripClauseWrappers(right)} {
		if strings.Contains(clause, " that ") || strings.Contains(clause, " according to ") || strings.ContainsAny(clause, ",;:") {
			return nil
		}
	}
	return []clausePair{{left: left, right: right}}
}

func normalizeClausePreservingCase(value string) string {
	value = expandGateContractions(value)
	value = strings.Join(strings.Fields(value), " ")
	value = strings.Trim(value, ", ")
	if strings.HasSuffix(value, ".") {
		value = strings.TrimSuffix(value, ".")
	}
	return expandGateElidedDisjunction(normalizeGateConditional(value))
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

// hasOuterGateOperator recognizes explicit scopes which must not be split
// at an embedded and/or. Bare legacy conjunction/disjunction spellings remain
// supported; explicit either/both scaffolds are handled by their own splitters.
func hasOuterGateOperator(statement string) bool {
	for _, prefix := range []string{"if ", "it is not the case that ", "necessarily, ", "necessarily ", "possibly, ", "possibly "} {
		if strings.HasPrefix(statement, prefix) {
			return true
		}
	}
	return strings.Contains(statement, " if and only if ") && !strings.HasPrefix(statement, "it is the case both that")
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

func conjunctiveAntecedentCandidates(statement string, partCount int) [][]string {
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
	return append(result, plainAndPartitions(statement, partCount)...)
}

func conjunctiveAntecedentCandidatesPreservingCase(statement string, partCount int) [][]string {
	statement = normalizeClausePreservingCase(statement)
	result := make([][]string, 0)
	separator := ", and " + clauseWrapper
	if strings.Contains(strings.ToLower(statement), separator) {
		parts := splitPreservingCase(statement, separator)
		for index := 1; index < len(parts); index++ {
			parts[index] = clauseWrapper + " " + parts[index]
		}
		result = append(result, parts)
	}
	for _, conjunction := range conjunctionCandidatesPreservingCase(statement) {
		result = append(result, []string{conjunction.left, conjunction.right})
	}
	// The schematic evaluator also enumerates premise assignments (k!).
	// Bound the additional plain-and search independently of legacy forms.
	if partCount <= 5 {
		result = append(result, plainAndPartitions(statement, partCount)...)
	}
	return result
}

// plainAndPartitions preserves case and enumerates each choice of boundaries
// once. Unsplit "and" tokens remain inside a conjunct, whose entire text must
// still match a cited premise. The cap bounds the search to at most 2^12 choices.
func plainAndPartitions(statement string, partCount int) [][]string {
	const separator = " and "
	lower := strings.ToLower(statement)
	count := strings.Count(lower, separator)
	if count > 12 || partCount < 3 || partCount > count+1 {
		return nil
	}
	var boundaries []int
	for start := 0; ; {
		index := strings.Index(lower[start:], separator)
		if index < 0 {
			break
		}
		index += start
		boundaries = append(boundaries, index)
		start = index + len(separator)
	}
	var result [][]string
	var visit func(int, int, []string)
	visit = func(start, next int, parts []string) {
		if len(parts) == partCount-1 {
			result = append(result, append(append([]string(nil), parts...), statement[start:]))
			return
		}
		for i := next; i < len(boundaries); i++ {
			if len(boundaries)-i < partCount-1-len(parts) {
				break
			}
			boundary := boundaries[i]
			visit(boundary+len(separator), i+1, append(parts, statement[start:boundary]))
		}
	}
	visit(0, 0, nil)
	return result
}

func conjunctionCandidatesPreservingCase(statement string) []clausePair {
	statement = normalizeClausePreservingCase(statement)
	result := make([]clausePair, 0)
	lower := strings.ToLower(statement)
	if strings.HasPrefix(lower, "it is the case both that") {
		remainder := statement[len("it is the case both that"):]
		if remainder == "" || remainder[0] == ',' || remainder[0] == ' ' {
			remainder = strings.TrimLeft(remainder, ", ")
			result = appendSplitCandidatesPreservingCase(result, remainder, ", and that ", true)
		}
	}
	result = appendSplitCandidatesPreservingCase(result, statement, ", and that ", true)
	result = appendSplitCandidatesPreservingCase(result, statement, ", and ", false)
	result = appendSplitCandidatesPreservingCase(result, statement, " and ", false)
	return result
}

func splitPreservingCase(statement, separator string) []string {
	lower := strings.ToLower(statement)
	parts := make([]string, 0)
	start := 0
	for {
		index := strings.Index(lower[start:], separator)
		if index < 0 {
			return append(parts, normalizeClausePreservingCase(statement[start:]))
		}
		index += start
		parts = append(parts, normalizeClausePreservingCase(statement[start:index]))
		start = index + len(separator)
	}
}

func disjunctionCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	if hasOuterGateOperator(statement) || strings.HasPrefix(statement, "it is the case both that") {
		return nil
	}
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

// groundInstanceCandidates narrows instanceCandidates to an unquantified
// witness with its first copula as the subject/predicate boundary. This keeps
// the existential forms from importing existence from universals or treating
// an embedded copula as a second, invented witness.
func groundInstanceCandidates(statement string) []clausePair {
	statement = normalizeClause(statement)
	for _, prefix := range []string{
		"a ", "an ", "all ", "any ", "each ", "every ", "everything ", "everyone ",
		"few ", "if ", "it is not the case", "many ", "most ", "necessarily ",
		"no ", "no one ", "nothing ", "several ", "some ", "someone ", "something ",
		"there ", "at least ", "at most ",
	} {
		if strings.HasPrefix(statement, prefix) && !(prefix == "a " && strings.HasPrefix(statement, "a is ")) {
			return nil
		}
	}

	candidates := instanceCandidates(statement)
	if len(candidates) == 0 {
		return nil
	}
	shortestSubject := len(statement) + 1
	for _, candidate := range candidates {
		if len(candidate.left) < shortestSubject {
			shortestSubject = len(candidate.left)
		}
	}
	result := make([]clausePair, 0, 1)
	for _, candidate := range candidates {
		if len(candidate.left) == shortestSubject {
			result = append(result, candidate)
		}
	}
	return result
}

type existentialClause struct {
	class        string
	matrix       string
	unrestricted bool
}

func unrestrictedExistentialPredicateCandidates(statement string) []string {
	statement = normalizeClause(statement)
	result := make([]string, 0, 1)
	for _, prefix := range []string{
		"there exists something that is ",
		"there is something that is ",
		"something is ",
	} {
		if !strings.HasPrefix(statement, prefix) {
			continue
		}
		predicate := normalizeClause(strings.TrimPrefix(statement, prefix))
		if predicate != "" {
			result = append(result, predicate)
		}
	}
	return result
}

func thereExistRestrictedCandidates(statement string) []existentialClause {
	statement = normalizeClause(statement)
	if !strings.HasPrefix(statement, "there exist ") {
		return nil
	}
	remainder := strings.TrimPrefix(statement, "there exist ")
	result := make([]existentialClause, 0)
	for _, separator := range []string{" that are ", " that is ", " that "} {
		for _, split := range appendSplitCandidates(nil, remainder, separator, false) {
			result = append(result, existentialClause{class: split.left, matrix: split.right})
		}
	}
	return result
}

func existentialIntroductionConclusionCandidates(statement string) []existentialClause {
	result := thereExistRestrictedCandidates(statement)
	statement = normalizeClause(statement)
	if !strings.HasPrefix(statement, "some ") {
		return result
	}
	remainder := strings.TrimPrefix(statement, "some ")
	for _, separator := range []string{" is ", " are "} {
		for _, split := range appendSplitCandidates(nil, remainder, separator, false) {
			result = append(result, existentialClause{class: split.left, matrix: split.right})
		}
	}
	return result
}

func conjunctionPartitions(statement string, partCount int) [][]string {
	statement = normalizeClause(statement)
	if statement == "" || partCount < 1 {
		return nil
	}
	if partCount == 1 {
		return [][]string{{statement}}
	}

	result := make([][]string, 0)
	for _, split := range conjunctionCandidates(statement) {
		for leftCount := 1; leftCount < partCount; leftCount++ {
			leftPartitions := conjunctionPartitions(split.left, leftCount)
			rightPartitions := conjunctionPartitions(split.right, partCount-leftCount)
			for _, left := range leftPartitions {
				for _, right := range rightPartitions {
					partition := append(append([]string(nil), left...), right...)
					result = append(result, partition)
				}
			}
		}
	}
	return result
}

func replaceSchematicVariables(value string) (string, schematicVariableCounts) {
	runes := []rune(value)
	var result strings.Builder
	counts := schematicVariableCounts{}
	for index, current := range runes {
		marker, isVariable := schematicVariableMarkers[current]
		if isVariable &&
			(index == 0 || !unicode.IsLetter(runes[index-1])) &&
			(index == len(runes)-1 || !unicode.IsLetter(runes[index+1])) {
			result.WriteRune(marker)
			counts[current]++
			continue
		}
		result.WriteRune(current)
	}
	return result.String(), counts
}

func schematicTemplate(value string) (string, schematicVariableCounts) {
	replaced, counts := replaceSchematicVariables(value)
	return comparisonClause(replaced), counts
}

func unifySchematicTemplate(template, concrete string, initial schematicSubstitution) []schematicSubstitution {
	templateRunes := []rune(comparisonClause(template))
	concreteRunes := []rune(comparisonClause(concrete))
	return unifySchematicRunes(templateRunes, concreteRunes, 0, 0, cloneSchematicSubstitution(initial))
}

func unifySchematicRunes(
	template []rune,
	concrete []rune,
	templateIndex int,
	concreteIndex int,
	substitution schematicSubstitution,
) []schematicSubstitution {
	if templateIndex == len(template) {
		if concreteIndex == len(concrete) {
			return []schematicSubstitution{substitution}
		}
		return nil
	}

	variable, isMarker := schematicVariableForMarker(template[templateIndex])
	if !isMarker {
		if concreteIndex >= len(concrete) || template[templateIndex] != concrete[concreteIndex] {
			return nil
		}
		return unifySchematicRunes(template, concrete, templateIndex+1, concreteIndex+1, substitution)
	}

	if bound, ok := substitution[variable]; ok {
		boundRunes := []rune(bound)
		if len(boundRunes) == 0 || concreteIndex+len(boundRunes) > len(concrete) {
			return nil
		}
		for index, current := range boundRunes {
			if concrete[concreteIndex+index] != current {
				return nil
			}
		}
		return unifySchematicRunes(template, concrete, templateIndex+1, concreteIndex+len(boundRunes), substitution)
	}

	result := make([]schematicSubstitution, 0)
	for end := concreteIndex + 1; end <= len(concrete); end++ {
		segment := concrete[concreteIndex:end]
		if containsSchematicMarker(segment) {
			continue
		}
		next := cloneSchematicSubstitution(substitution)
		next[variable] = string(segment)
		result = append(result, unifySchematicRunes(template, concrete, templateIndex+1, end, next)...)
	}
	return result
}

func schematicVariableForMarker(marker rune) (rune, bool) {
	for variable, candidate := range schematicVariableMarkers {
		if marker == candidate {
			return variable, true
		}
	}
	return 0, false
}

func containsSchematicMarker(value []rune) bool {
	for _, current := range value {
		if _, ok := schematicVariableForMarker(current); ok {
			return true
		}
	}
	return false
}

func cloneSchematicSubstitution(source schematicSubstitution) schematicSubstitution {
	result := make(schematicSubstitution, len(source))
	for variable, value := range source {
		result[variable] = value
	}
	return result
}

func instantiateSchematicTemplate(template string, substitution schematicSubstitution) (string, bool) {
	var result strings.Builder
	for _, current := range template {
		variable, isMarker := schematicVariableForMarker(current)
		if !isMarker {
			result.WriteRune(current)
			continue
		}
		value, ok := substitution[variable]
		if !ok || value == "" {
			return "", false
		}
		result.WriteString(value)
	}
	return result.String(), true
}

func schematicVariablesSubset(subset, superset schematicVariableCounts) bool {
	for variable := range subset {
		if superset[variable] == 0 {
			return false
		}
	}
	return true
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

func appendSplitCandidatesPreservingCase(result []clausePair, statement, separator string, stripRightThat bool) []clausePair {
	lower := strings.ToLower(statement)
	start := 0
	for {
		index := strings.Index(lower[start:], separator)
		if index < 0 {
			return result
		}
		index += start
		left := normalizeClausePreservingCase(statement[:index])
		right := normalizeClausePreservingCase(statement[index+len(separator):])
		if stripRightThat && len(right) >= len("that ") && strings.EqualFold(right[:len("that ")], "that ") {
			right = right[len("that "):]
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
			conjunctiveCandidates := conjunctiveAntecedentCandidates(conditional.left, len(premises)-1)
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

func matchSchematicConjunctiveModusPonens(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	if len(premises) < 3 {
		return schemaMatchOutcome{}
	}

	best := schemaMatchOutcome{}
	for conditionalIndex, premise := range premises {
		for _, conditional := range conditionalCandidatesPreservingCase(premise.Statement) {
			antecedentTemplate, antecedentVariables := schematicTemplate(conditional.left)
			consequentTemplate, consequentVariables := schematicTemplate(conditional.right)
			if len(antecedentVariables) == 0 && len(consequentVariables) == 0 {
				continue
			}

			candidate := schemaMatchOutcome{traction: true}
			if !schematicVariablesSubset(consequentVariables, antecedentVariables) {
				candidate.mismatch = fmt.Sprintf(
					"premise %d consequent contains a schematic variable absent from its antecedent",
					premise.Number,
				)
				best = betterOutcome(best, candidate)
				continue
			}

			conjunctiveCandidates := conjunctiveAntecedentCandidatesPreservingCase(conditional.left, len(premises)-1)
			if len(conjunctiveCandidates) == 0 {
				candidate.mismatch = fmt.Sprintf(
					"premise %d antecedent %q does not split into two or more conjuncts",
					premise.Number,
					normalizeClause(conditional.left),
				)
				best = betterOutcome(best, candidate)
				continue
			}

			for _, conjuncts := range conjunctiveCandidates {
				conjunctTemplates := make([]string, len(conjuncts))
				for index, conjunct := range conjuncts {
					conjunctTemplates[index], _ = schematicTemplate(conjunct)
				}
				candidate = evaluateSchematicConjunctiveModusPonensCandidate(
					premises,
					conditionalIndex,
					antecedentTemplate,
					consequentTemplate,
					conjunctTemplates,
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

func evaluateSchematicConjunctiveModusPonensCandidate(
	premises []NumberedArgumentStep,
	conditionalIndex int,
	antecedentTemplate string,
	consequentTemplate string,
	conjunctTemplates []string,
	conclusion NumberedArgumentStep,
) schemaMatchOutcome {
	conditionalPremise := premises[conditionalIndex]
	assignments := schematicPremiseAssignments(premises, conditionalIndex, len(conjunctTemplates))
	if len(assignments) == 0 {
		return schemaMatchOutcome{
			traction: true,
			mismatch: fmt.Sprintf(
				"premise %d antecedent has %d conjuncts but only %d other cited premises",
				conditionalPremise.Number,
				len(conjunctTemplates),
				len(premises)-1,
			),
		}
	}

	best := schemaMatchOutcome{}
	for _, assignment := range assignments {
		states := []schematicSubstitution{{}}
		candidate := schemaMatchOutcome{traction: true}
		for conjunctIndex, premiseIndex := range assignment {
			nextStates := make([]schematicSubstitution, 0)
			for _, state := range states {
				nextStates = append(
					nextStates,
					unifySchematicTemplate(conjunctTemplates[conjunctIndex], premises[premiseIndex].Statement, state)...,
				)
			}
			if len(nextStates) == 0 {
				candidate.mismatch = fmt.Sprintf(
					"premise %d antecedent conjunct %q does not instantiate cited premise %d %q under one consistent substitution",
					conditionalPremise.Number,
					normalizeClause(conjunctTemplates[conjunctIndex]),
					premises[premiseIndex].Number,
					normalizeClause(premises[premiseIndex].Statement),
				)
				states = nil
				break
			}
			states = nextStates
			candidate.score++
		}

		if len(states) == 0 {
			best = betterOutcome(best, candidate)
			continue
		}

		for _, state := range states {
			instantiatedAntecedent, antecedentOK := instantiateSchematicTemplate(antecedentTemplate, state)
			instantiatedConsequent, consequentOK := instantiateSchematicTemplate(consequentTemplate, state)
			if !antecedentOK || !consequentOK {
				continue
			}
			stateCandidate := schemaMatchOutcome{traction: true, score: len(conjunctTemplates)}
			consequentOutcome := evaluateIdentities(clauseIdentity{
				fmt.Sprintf("premise %d instantiated consequent", conditionalPremise.Number), instantiatedConsequent,
				fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement,
			})
			appendOutcome(&stateCandidate, consequentOutcome)

			unused := firstUnusedSchematicPremise(premises, conditionalIndex, assignment)
			if stateCandidate.mismatch == "" && unused.Number != 0 {
				stateCandidate.mismatch = fmt.Sprintf(
					"cited premise %d %q is unused by premise %d conjunctive antecedent %q",
					unused.Number,
					normalizeClause(unused.Statement),
					conditionalPremise.Number,
					normalizeClause(instantiatedAntecedent),
				)
			}
			stateCandidate.matched = stateCandidate.mismatch == "" && len(conjunctTemplates)+1 == len(premises)
			best = betterOutcome(best, stateCandidate)
			if best.matched {
				return best
			}
		}
	}
	return best
}

func schematicPremiseAssignments(premises []NumberedArgumentStep, conditionalIndex, conjunctCount int) [][]int {
	available := make([]int, 0, len(premises)-1)
	for index, premise := range premises {
		if index != conditionalIndex && premise.Number != premises[conditionalIndex].Number {
			available = append(available, index)
		}
	}
	if conjunctCount > len(available) {
		return nil
	}

	result := make([][]int, 0)
	var enumerate func([]int, []bool, map[int]bool)
	enumerate = func(current []int, used []bool, usedNumbers map[int]bool) {
		if len(current) == conjunctCount {
			result = append(result, append([]int(nil), current...))
			return
		}
		for availableIndex, premiseIndex := range available {
			if used[availableIndex] || usedNumbers[premises[premiseIndex].Number] {
				continue
			}
			used[availableIndex] = true
			usedNumbers[premises[premiseIndex].Number] = true
			enumerate(append(current, premiseIndex), used, usedNumbers)
			delete(usedNumbers, premises[premiseIndex].Number)
			used[availableIndex] = false
		}
	}
	enumerate(nil, make([]bool, len(available)), map[int]bool{premises[conditionalIndex].Number: true})
	return result
}

func firstUnusedSchematicPremise(
	premises []NumberedArgumentStep,
	conditionalIndex int,
	assignment []int,
) NumberedArgumentStep {
	used := make([]bool, len(premises))
	used[conditionalIndex] = true
	for _, premiseIndex := range assignment {
		used[premiseIndex] = true
	}
	for premiseIndex, premise := range premises {
		if !used[premiseIndex] {
			return premise
		}
	}
	return NumberedArgumentStep{}
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

func matchBiconditionalIntroduction(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, forward := range conditionalCandidates(premises[0].Statement) {
		for _, reverse := range conditionalCandidates(premises[1].Statement) {
			for _, biconditional := range biconditionalCandidates(conclusion.Statement) {
				candidate := evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d antecedent", premises[0].Number), forward.left, fmt.Sprintf("conclusion %d left clause", conclusion.Number), biconditional.left},
					clauseIdentity{fmt.Sprintf("premise %d consequent", premises[0].Number), forward.right, fmt.Sprintf("conclusion %d right clause", conclusion.Number), biconditional.right},
					clauseIdentity{fmt.Sprintf("premise %d antecedent", premises[1].Number), reverse.left, fmt.Sprintf("conclusion %d right clause", conclusion.Number), biconditional.right},
					clauseIdentity{fmt.Sprintf("premise %d consequent", premises[1].Number), reverse.right, fmt.Sprintf("conclusion %d left clause", conclusion.Number), biconditional.left},
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

// MatchesBiconditionalIntroduction reports whether a biconditional is licensed
// as one atomic step. A conclusion must be introduced either from two verbatim
// directional conditionals or by ordinary/conjunctive modus ponens from a
// conditional whose full consequent is the verbatim biconditional conclusion.
// A biconditional embedded in an anchored conditional premise is the author's
// premise commitment, rather than a derived biconditional introduction.
//
// Keeping every case in this predicate ensures the catalog and atomicity scan
// share the exact parsing and clause-identity rules.
func MatchesBiconditionalIntroduction(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) bool {
	if conclusion.Kind == numberedArgumentStepPremise {
		return biconditionalEmbeddedInConditional(conclusion.Statement)
	}
	if len(biconditionalCandidates(conclusion.Statement)) == 0 {
		return false
	}
	for _, ordered := range orderedPremiseSelections(premises, 2) {
		if matchBiconditionalIntroduction(ordered, conclusion).matched {
			return true
		}
		if matchModusPonens(ordered, conclusion).matched {
			return true
		}
	}
	return matchConjunctiveModusPonens(premises, conclusion).matched
}

func biconditionalEmbeddedInConditional(statement string) bool {
	const connective = " if and only if "
	for _, conditional := range conditionalCandidates(statement) {
		if strings.Contains(conditional.left, connective) || strings.Contains(conditional.right, connective) {
			return true
		}
	}
	return false
}

func matchBiconditionalElimination(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, biconditional := range biconditionalCandidates(premises[0].Statement) {
		for _, conditional := range conditionalCandidates(conclusion.Statement) {
			for _, direction := range []clausePair{biconditional, {left: biconditional.right, right: biconditional.left}} {
				candidate := evaluateIdentities(
					clauseIdentity{fmt.Sprintf("premise %d antecedent", premises[0].Number), direction.left, fmt.Sprintf("conclusion %d antecedent", conclusion.Number), conditional.left},
					clauseIdentity{fmt.Sprintf("premise %d consequent", premises[0].Number), direction.right, fmt.Sprintf("conclusion %d consequent", conclusion.Number), conditional.right},
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

func matchBiconditionalModusPonens(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, biconditional := range biconditionalCandidates(premises[0].Statement) {
		for _, direction := range []clausePair{biconditional, {left: biconditional.right, right: biconditional.left}} {
			candidate := evaluateIdentities(
				clauseIdentity{fmt.Sprintf("premise %d antecedent", premises[0].Number), direction.left, fmt.Sprintf("premise %d", premises[1].Number), premises[1].Statement},
				clauseIdentity{fmt.Sprintf("premise %d consequent", premises[0].Number), direction.right, fmt.Sprintf("conclusion %d", conclusion.Number), conclusion.Statement},
			)
			best = betterOutcome(best, candidate)
			if best.matched {
				return best
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
	// Raw conjunction candidates also serve bridge antecedents, where cited
	// clauses disambiguate scope. Elimination has no such evidence: an outer
	// negation, conditional, disjunction or modal must remain intact.
	statement := normalizeClause(premises[0].Statement)
	if hasOuterGateOperator(statement) || strings.HasPrefix(statement, "either ") {
		return schemaMatchOutcome{}
	}
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

// The non-copular register uses a one-word class/subject and an opaque verb phrase.
// Do not guess noun/verb boundaries in modified classes, relative clauses, or
// coordinated/conditional predicates. Predicate agreement is exact: no verb
// inflection or termCandidates tolerance is applied to the verb phrase.
func matchNonCopularUniversalInstantiation(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	statement := normalizeClause(premises[0].Statement)
	remainder := ""
	for _, prefix := range []string{"all ", "every "} {
		if strings.HasPrefix(statement, prefix) {
			remainder = strings.TrimPrefix(statement, prefix)
			break
		}
	}
	class, predicate, ok := strings.Cut(remainder, " ")
	if !ok || class == "" || predicate == "" {
		return schemaMatchOutcome{}
	}
	first, _, _ := strings.Cut(predicate, " ")
	switch first {
	case "is", "are", "am", "who", "which", "that", "whose", "with", "without", "of", "in", "on", "from", "and", "or":
		return schemaMatchOutcome{}
	}
	if strings.ContainsAny(remainder, ",;:()") {
		return schemaMatchOutcome{}
	}
	for _, connective := range []string{" and ", " or ", " if ", " unless ", " then ", " who ", " which ", " that "} {
		if strings.Contains(predicate, connective) {
			return schemaMatchOutcome{}
		}
	}
	best := schemaMatchOutcome{}
	for _, instance := range groundInstanceCandidates(premises[1].Statement) {
		// A complex subject could hide an attitude report ("Ada believes whales
		// are mammals"); substituting inside it would not be instantiation.
		if len(strings.Fields(instance.left)) != 1 {
			continue
		}
		// This schema requires is/are membership, not the broader am register.
		membership := normalizeClause(premises[1].Statement)
		if !strings.HasPrefix(membership, instance.left+" is ") && !strings.HasPrefix(membership, instance.left+" are ") {
			continue
		}
		candidate := evaluateTermIdentities(clauseIdentity{"universal class", class, "membership class", instance.right})
		appendOutcome(&candidate, evaluateIdentities(clauseIdentity{"instantiated predicate", instance.left + " " + predicate, "conclusion", conclusion.Statement}))
		candidate.matched = candidate.score == 2
		best = betterOutcome(best, candidate)
		if best.matched {
			return best
		}
	}
	return best
}

func matchUniversalInstantiation(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, universal := range universalCandidates(premises[0].Statement) {
		for _, instance := range instanceCandidates(premises[1].Statement) {
			if quantifiedInstanceSubject(instance.left) {
				continue
			}
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
		if quantifiedInstanceSubject(instance.left) {
			continue
		}
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

// matchSchematicInstantiation certifies If A(vars), then B(vars); A(sigma);
// therefore B(sigma), where vars are standalone uppercase X, Y, or Z and one
// consistent non-empty substitution verifies every occurrence.
func matchSchematicInstantiation(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, conditional := range conditionalCandidatesPreservingCase(premises[0].Statement) {
		antecedentTemplate, antecedentVariables := schematicTemplate(conditional.left)
		consequentTemplate, consequentVariables := schematicTemplate(conditional.right)
		candidate := schemaMatchOutcome{traction: true}
		if len(antecedentVariables) == 0 || len(consequentVariables) == 0 {
			switch {
			case len(antecedentVariables) == 0 && len(consequentVariables) == 0:
				candidate.mismatch = fmt.Sprintf("premise %d conditional has no standalone uppercase X, Y, or Z in either side", premises[0].Number)
			case len(antecedentVariables) == 0:
				candidate.score = 1
				candidate.mismatch = fmt.Sprintf("premise %d antecedent has no standalone uppercase X, Y, or Z", premises[0].Number)
			default:
				candidate.score = 1
				candidate.mismatch = fmt.Sprintf("premise %d consequent has no standalone uppercase X, Y, or Z", premises[0].Number)
			}
			best = betterOutcome(best, candidate)
			continue
		}

		if !schematicVariablesSubset(consequentVariables, antecedentVariables) {
			candidate.mismatch = fmt.Sprintf(
				"premise %d consequent contains a schematic variable absent from its antecedent",
				premises[0].Number,
			)
			best = betterOutcome(best, candidate)
			continue
		}

		substitutions := unifySchematicTemplate(antecedentTemplate, premises[1].Statement, nil)
		if len(substitutions) == 0 {
			candidate.score = 1
			candidate.mismatch = fmt.Sprintf(
				"premise %d does not provide one consistent non-empty substitution for premise %d antecedent template",
				premises[1].Number,
				premises[0].Number,
			)
			best = betterOutcome(best, candidate)
			continue
		}

		for _, substitution := range substitutions {
			instantiatedAntecedent, antecedentOK := instantiateSchematicTemplate(antecedentTemplate, substitution)
			instantiatedConsequent, consequentOK := instantiateSchematicTemplate(consequentTemplate, substitution)
			if !antecedentOK || !consequentOK {
				continue
			}
			candidate = evaluateIdentities(
				clauseIdentity{
					fmt.Sprintf("premise %d instantiated antecedent", premises[0].Number),
					instantiatedAntecedent,
					fmt.Sprintf("premise %d", premises[1].Number),
					premises[1].Statement,
				},
				clauseIdentity{
					fmt.Sprintf("premise %d instantiated consequent", premises[0].Number),
					instantiatedConsequent,
					fmt.Sprintf("conclusion %d", conclusion.Number),
					conclusion.Statement,
				},
			)
			best = betterOutcome(best, candidate)
			if best.matched {
				return best
			}
		}
	}
	return best
}

func matchExistentialGeneralization(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, instance := range groundInstanceCandidates(premises[0].Statement) {
		for _, predicate := range unrestrictedExistentialPredicateCandidates(conclusion.Statement) {
			candidate := evaluateIdentities(clauseIdentity{
				fmt.Sprintf("premise %d predicate", premises[0].Number), instance.right,
				fmt.Sprintf("conclusion %d existential predicate", conclusion.Number), predicate,
			})
			best = betterOutcome(best, candidate)
			if best.matched {
				return best
			}
		}
	}
	return best
}

func matchExistentialIntroduction(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, existential := range existentialIntroductionConclusionCandidates(conclusion.Statement) {
		for witnessIndex, witnessPremise := range premises {
			for _, witness := range groundInstanceCandidates(witnessPremise.Statement) {
				candidate := evaluateTermIdentities(clauseIdentity{
					fmt.Sprintf("premise %d witness class", witnessPremise.Number), witness.right,
					fmt.Sprintf("conclusion %d existential class", conclusion.Number), existential.class,
				})
				candidate.matched = false
				candidate.traction = true

				for _, predicates := range conjunctionPartitions(existential.matrix, len(premises)-1) {
					if !matchExistentialPredicatePremises(premises, witnessIndex, witness.left, predicates) {
						continue
					}
					candidate.score += len(predicates)
					if candidate.mismatch == "" {
						candidate.matched = true
						return candidate
					}
				}

				if candidate.mismatch == "" {
					candidate.mismatch = fmt.Sprintf(
						"the remaining cited premises do not give subject %q every conclusion predicate verbatim",
						normalizeClause(witness.left),
					)
				}
				best = betterOutcome(best, candidate)
			}
		}
	}
	return best
}

func matchExistentialPredicatePremises(
	premises []NumberedArgumentStep,
	witnessIndex int,
	subject string,
	predicates []string,
) bool {
	used := make([]bool, len(premises))
	used[witnessIndex] = true
	var matchNext func(int) bool
	matchNext = func(predicateIndex int) bool {
		if predicateIndex == len(predicates) {
			return true
		}
		for premiseIndex, premise := range premises {
			if used[premiseIndex] {
				continue
			}
			for _, instance := range groundInstanceCandidates(premise.Statement) {
				if !equalClauses(subject, instance.left) || !equalClauses(predicates[predicateIndex], instance.right) {
					continue
				}
				used[premiseIndex] = true
				if matchNext(predicateIndex + 1) {
					return true
				}
				used[premiseIndex] = false
			}
		}
		return false
	}
	return matchNext(0)
}

type existentialCounterexampleClause struct {
	class        string
	restrictor   string
	negated      string
	unrestricted bool
}

func existentialCounterexampleCandidates(statement string) []existentialCounterexampleClause {
	statement = normalizeClause(statement)
	sources := thereExistRestrictedCandidates(statement)
	const unrestrictedPrefix = "there exists something that is "
	if strings.HasPrefix(statement, unrestrictedPrefix) {
		matrix := normalizeClause(strings.TrimPrefix(statement, unrestrictedPrefix))
		if matrix != "" {
			sources = append(sources, existentialClause{matrix: matrix, unrestricted: true})
		}
	}

	result := make([]existentialCounterexampleClause, 0)
	for _, source := range sources {
		if strings.HasPrefix(source.matrix, "not ") {
			negated := normalizeClause(strings.TrimPrefix(source.matrix, "not "))
			if negated != "" {
				result = append(result, existentialCounterexampleClause{
					class: source.class, negated: negated, unrestricted: source.unrestricted,
				})
			}
		}
		for _, conjunction := range conjunctionCandidates(source.matrix) {
			if !strings.HasPrefix(conjunction.right, "not ") {
				continue
			}
			negated := normalizeClause(strings.TrimPrefix(conjunction.right, "not "))
			if conjunction.left == "" || negated == "" {
				continue
			}
			result = append(result, existentialCounterexampleClause{
				class: source.class, restrictor: conjunction.left, negated: negated, unrestricted: source.unrestricted,
			})
		}
	}
	return result
}

type universalCounterexampleClause struct {
	class        string
	restrictor   string
	consequent   string
	unrestricted bool
}

func universalCounterexampleCandidates(statement string) []universalCounterexampleClause {
	statement = normalizeClause(statement)
	result := make([]universalCounterexampleClause, 0)
	for _, universal := range universalCandidates(statement) {
		switch {
		case strings.HasPrefix(statement, "everything "):
			if !strings.HasPrefix(universal.left, "that is ") {
				continue
			}
			restrictor := normalizeClause(strings.TrimPrefix(universal.left, "that is "))
			if restrictor != "" {
				result = append(result, universalCounterexampleClause{
					restrictor: restrictor, consequent: universal.right, unrestricted: true,
				})
			}
		case strings.HasPrefix(statement, "every "):
			restricted := appendSplitCandidates(nil, universal.left, " that is ", false)
			for _, split := range restricted {
				result = append(result, universalCounterexampleClause{
					class: split.left, restrictor: split.right, consequent: universal.right,
				})
			}
			if len(restricted) == 0 && !strings.Contains(universal.left, " that") {
				result = append(result, universalCounterexampleClause{
					class: universal.left, consequent: universal.right,
				})
			}
		}
	}
	return result
}

func matchExistentialCounterexample(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	best := schemaMatchOutcome{}
	for _, negatedConclusion := range negationCandidates(conclusion.Statement) {
		for _, existential := range existentialCounterexampleCandidates(premises[0].Statement) {
			for _, universal := range universalCounterexampleCandidates(negatedConclusion) {
				candidate := schemaMatchOutcome{traction: true}
				switch {
				case existential.unrestricted && universal.unrestricted:
					candidate.score++
				case !existential.unrestricted && !universal.unrestricted:
					appendOutcome(&candidate, evaluateTermIdentities(clauseIdentity{
						fmt.Sprintf("premise %d existential class", premises[0].Number), existential.class,
						fmt.Sprintf("conclusion %d universal class", conclusion.Number), universal.class,
					}))
				default:
					candidate.mismatch = "the existential and universal do not have the same restricted or unrestricted domain"
				}
				appendOutcome(&candidate, evaluateIdentities(
					clauseIdentity{
						fmt.Sprintf("premise %d restrictor", premises[0].Number), existential.restrictor,
						fmt.Sprintf("conclusion %d restrictor", conclusion.Number), universal.restrictor,
					},
					clauseIdentity{
						fmt.Sprintf("premise %d negated predicate", premises[0].Number), existential.negated,
						fmt.Sprintf("conclusion %d consequent", conclusion.Number), universal.consequent,
					},
				))
				candidate.matched = candidate.mismatch == "" && candidate.score == 3
				best = betterOutcome(best, candidate)
				if best.matched {
					return best
				}
			}
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

// matchNaryConjunctionIntroduction (catalog v10) certifies a conclusion that is
// literally the conjunction of all of its three or more cited premises, in any
// order: "A, B, and C", "A, B and C" or "A and B and C". The comparison is on
// whole normalized clauses, so no conjunct is split or reinterpreted.
func matchNaryConjunctionIntroduction(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	if len(premises) < 3 || len(premises) > 6 {
		return schemaMatchOutcome{}
	}
	target := comparisonClause(conclusion.Statement)
	parts := make([]string, len(premises))
	for i, premise := range premises {
		parts[i] = comparisonClause(premise.Statement)
	}
	matched := false
	permuteClauses(parts, 0, func(order []string) bool {
		last := len(order) - 1
		listed := strings.Join(order[:last], " ") + " and " + order[last]
		chained := strings.Join(order, " and ")
		matched = target == listed || target == chained
		return matched
	})
	if matched {
		return schemaMatchOutcome{matched: true, score: len(premises), traction: true}
	}
	return schemaMatchOutcome{mismatch: fmt.Sprintf("conclusion %d is not exactly the conjunction of its cited premises", conclusion.Number)}
}

// permuteClauses visits orderings of parts; visit returns true to stop.
func permuteClauses(parts []string, from int, visit func([]string) bool) bool {
	if from == len(parts) {
		return visit(parts)
	}
	for i := from; i < len(parts); i++ {
		parts[from], parts[i] = parts[i], parts[from]
		stop := permuteClauses(parts, from+1, visit)
		parts[from], parts[i] = parts[i], parts[from]
		if stop {
			return true
		}
	}
	return false
}

// EntailmentGateForms lists the names of the accepted inference forms in
// match order, for tools that describe the catalog.
func EntailmentGateForms() []string {
	names := make([]string, 0, len(entailmentSchemas))
	for _, schema := range entailmentSchemas {
		names = append(names, schema.name)
	}
	return append(names, fitchForms...)
}
