package validation

import (
	"strings"
	"unicode"
)

const mixedScopeHint = "use 'both … and' or 'either … or' to show the grouping"
const clauseStartHint = "each inspected clause must start with a determiner, proper name, pronoun or numeral; move the operator into each clause, or state the claim as an explicit premise"
const ingFrameHint = "move the frame into each clause, or state the framed claim as its own premise"
const copulaHint = "repeat the subject and verb: 'or the men are fish'"
const descriptionNegationHint = "add 'There is exactly one present king of France' and name it, or keep the wide negation"
const lyExceptionHint = "move the adverb into each clause, or state the framed claim as its own premise"

var round2OldLyExceptions = wordSet("only early daily")

// Closed predicate-list continuations preserve established list normalization;
// they are not evidence for an arbitrary bare-word clause.
var scopePredicateListWords = wordSet("light portable")
var scopeContractions = wordSet("here there let what where how why when who that it he she")

var scopeExpletiveVerbs = wordSet("is are was were be must may might can could will would should")

// scopeNameVerbs extends the frame verb lists for clause starts only: a closed
// allowlist of finite verbs (harvested from v44 corpus refusals) that may follow a
// capitalised subject. A finite verb after the subject is not a scope operator;
// sentence adverbs, -ly and -ing words are still excluded before this check.
var scopeNameVerbs = wordSet("requires require required makes make fails failed ranks rank ranked " +
	"entails entail entailed shares share shared violates violate violated seems seem seemed " +
	"exhibits exhibit exhibited represents represent represented implies imply implied holds hold held " +
	"satisfies satisfy satisfied faces face faced lacks lack lacked assumes assume assumed " +
	"illustrates illustrate illustrated conflicts conflict conflicted differs differ differed " +
	"admits admit admitted relies rely relied needs need needed provides provide provided " +
	"assigns assign assigned offers offer offered treats treat treated applies apply applied " +
	"operates operate operated possesses possess possessed dismisses dismiss dismissed " +
	"serves serve served imposes impose imposed belongs belong belonged shows show showed shown " +
	"counts count counted remains remain remained chooses choose chose depends depend depended " +
	"concerns concern concerned estimates estimate estimated permits permit permitted " +
	"increases increase increased suffices suffice sufficed seeks seek sought uses use used " +
	"takes take guarantees guarantee guaranteed produces produce produced proposes propose proposed " +
	"argues argue argued preserves preserve preserved greets greet greeted believes believe believed " +
	"claims claim claimed knows know thinks think thought says say said accepts accept accepted " +
	"rejects reject rejected supports support supported proves prove proved denies deny denied " +
	"predicts predict predicted prefers prefer preferred wants want wanted deserves deserve deserved " +
	"follows follow followed includes include included involves involve involved means mean meant " +
	"explains explain explained describes describe described defines define defined " +
	"determines determine determined affects affect affected achieves achieve achieved " +
	"receives receive received creates create created gets get got keeps keep kept " +
	"becomes become became goes go went comes come came sees see seen " +
	"intends intend intended believes asserts assert asserted denies " +
	"concludes conclude concluded infers infer inferred maintains maintain maintained " +
	"endorses endorse endorsed accepts values value valued reduces reduce reduced")

var scopeClauseWords = wordSet("the a an every some no all each any most many few several this that these those my your his her its our their i you he she it we they someone somebody something anyone anybody anything everyone everybody everything nobody nothing")

func scopeText(s string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "."))
}

// This final guard is independent of certificate names and precedes normalized
// equality (which can itself expand elision or move a negation). Literal reuse
// remains opaque. Conservative lexical splitting also covers the old grouping
// exemptions; they cannot license extraction from a containing sentence.
func (policy strictRules) round2ScopeFailure(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	target := scopeText(c.Statement)
	for _, id := range c.Citations {
		if strings.EqualFold(scopeText(steps[id].Statement), target) {
			return ""
		}
	}
	texts := []string{target}
	for _, id := range c.Citations {
		texts = append(texts, scopeText(fitchAssumptionStatement(steps[id])))
	}
	for _, s := range texts {
		if policy.scope_description_negation && unsafeDescriptionNegation(s) && descriptionConversionIn(s, texts) {
			return descriptionNegationHint
		}
		if policy.scope_elided_copula && expandGateElidedDisjunction(s) != s {
			_, right, _ := strings.Cut(strings.ToLower(s), " or ")
			if !singularPredicateHeads[right] {
				return copulaHint
			}
		}
		// A whole unchanged operand used in a larger formula is still opaque.
		opaque := false
		for _, other := range texts {
			if len(other) > len(s) && strings.Contains(strings.ToLower(other), strings.ToLower(s)) {
				opaque = true
			}
		}
		if opaque {
			continue
		}
		if scopePreservedPredicateList(s, texts) {
			continue
		}
		if why := policy.scopeExpressionFailure(s); why != "" {
			return why
		}
	}
	return ""
}

func unsafeDescriptionNegation(s string) bool {
	lower := strings.ToLower(s)
	if !strings.HasPrefix(lower, "it is not the case that ") {
		return false
	}
	body := strings.TrimSpace(s[len("it is not the case that "):])
	subject, _, ok := strings.Cut(body, " is ")
	if !ok {
		return false
	}
	words := strings.Fields(subject)
	if len(words) == 0 {
		return true
	}
	for _, w := range words {
		if scopeClauseWords[strings.ToLower(w)] || !unicode.IsUpper(firstRune(w)) {
			return true
		}
	}
	return false
}

func (policy strictRules) scopeExpressionFailure(s string) string {
	s = strings.TrimSpace(stripGroupedClause(s))
	if scopeLiteralAtom(s) {
		return ""
	}
	if expanded := expandGateElidedDisjunction(s); expanded != s {
		_, right, _ := strings.Cut(strings.ToLower(s), " or ")
		if policy.scope_elided_copula && !singularPredicateHeads[right] {
			return copulaHint
		}
		s = expanded
	}
	lower := strings.ToLower(s)
	// The catalog's explicit "both that ... , and that ..." scaffold owns
	// its separator. Treat its operands like parenthesised operands, rather
	// than rescanning their internal and/or as ungrouped top-level scope.
	const bothThat = "it is the case both that "
	if strings.HasPrefix(lower, bothThat) && strings.Count(lower, ", and that ") == 1 {
		i := strings.Index(lower, ", and that ")
		if i > len(bothThat) {
			if why := policy.scopeOperandFailure(s[len(bothThat):i]); why != "" {
				return why
			}
			return policy.scopeOperandFailure(s[i+len(", and that "):])
		}
	}
	// A disjunction of two such complete scaffolds has an explicit outer
	// "either" and a unique separator before the second scaffold.
	const orBothThat = " or it is the case both that "
	if strings.HasPrefix(lower, "either "+bothThat) && strings.Count(lower, orBothThat) == 1 {
		i := strings.Index(lower, orBothThat)
		if why := policy.scopeExpressionFailure(s[len("either "):i]); why != "" {
			return why
		}
		return policy.scopeExpressionFailure(s[i+len(" or "):])
	}
	// "If P, then Q": the antecedent and consequent are separate expressions;
	// "if" is not part of P's clause start.
	if strings.HasPrefix(lower, "if ") {
		if i := strings.Index(lower, ", then "); i > len("if ") {
			if why := policy.scopeExpressionFailure(s[len("if "):i]); why != "" {
				return why
			}
			return policy.scopeExpressionFailure(s[i+len(", then "):])
		}
	}
	// The and in this established operator is not coordination.
	if i := strings.Index(lower, " if and only if "); i >= 0 {
		if why := policy.scopeExpressionFailure(s[:i]); why != "" {
			return why
		}
		return policy.scopeExpressionFailure(s[i+len(" if and only if "):])
	}
	for _, prefix := range []string{"it is not the case that ", "it is the case that "} {
		if strings.HasPrefix(lower, prefix) {
			return policy.scopeExpressionFailure(s[len(prefix):])
		}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	first := strings.ToLower(strings.Trim(words[0], ","))
	// Closed-class clause starters such as "something" are not -ing frames.
	if policy.opener_ing_frame && !scopeClauseWords[first] && strings.HasSuffix(first, "ing") {
		return ingFrameHint
	}
	if policy.opener_ly_exception && round2OldLyExceptions[first] {
		return lyExceptionHint
	}
	// Nor has the same ambiguity as or, but is absent from the older splitter.
	scan := strings.ReplaceAll(strings.ToLower(s), " nor ", " or  ")
	bs, _ := groupingBoundaries(scan)
	and, or := false, false
	for _, b := range bs {
		and = and || b.op == "and"
		or = or || b.op == "or"
	}
	if op, left, right, managed, why := groupingPlan(s); managed && why == "" && left != "" && (op == "and" || op == "or") {
		if why := policy.scopeOperandFailure(left); why != "" {
			return why
		}
		return policy.scopeOperandFailure(right)
	}
	if and && or {
		explicit := strings.HasPrefix(lower, "both ") || strings.HasPrefix(lower, "either ")
		if !explicit && policy.scope_mixed_andor {
			return mixedScopeHint
		}
	}
	if len(bs) == 0 {
		return ""
	}
	body := s
	for _, prefix := range []string{"both ", "either ", "neither "} {
		if strings.HasPrefix(lower, prefix) {
			body = s[len(prefix):]
			break
		}
	}
	// Recompute offsets after removing the scaffold.
	bs, _ = groupingBoundaries(strings.ReplaceAll(strings.ToLower(body), " nor ", " or  "))
	start := 0
	for _, b := range bs {
		if why := policy.scopeOperandFailure(body[start:b.start]); why != "" {
			return why
		}
		start = b.end
	}
	return policy.scopeOperandFailure(body[start:])
}

func (policy strictRules) scopeOperandFailure(s string) string {
	s = strings.TrimSpace(strings.Trim(s, ","))
	if scopeLiteralAtom(strings.TrimSpace(stripGroupedClause(s))) {
		return ""
	}
	if why := policy.scopeExpressionFailure(s); why != "" {
		return why
	}
	s = strings.TrimSpace(stripGroupedClause(s))
	// A compound operand ("(B) or C") was split and every part checked by
	// scopeExpressionFailure above; its first word is not a clause start.
	if bs, _ := groupingBoundaries(strings.ReplaceAll(strings.ToLower(s), " nor ", " or  ")); len(bs) > 0 {
		return ""
	}
	lower := strings.ToLower(s)
	// scopeExpressionFailure has already checked the antecedent and
	// consequent separately. An explicit conditional is a complete operand,
	// not a nominal clause whose first word must be a subject.
	if strings.HasPrefix(lower, "if ") && strings.Contains(lower, ", then ") {
		return ""
	}
	// Expletive "there" opens a clause only before be or a modal ("There is
	// a dog", "There must be ..."); "There the lamp is lit" stays refused.
	if w := strings.Fields(lower); len(w) > 1 && w[0] == "there" && scopeExpletiveVerbs[w[1]] {
		return ""
	}
	if strings.HasPrefix(lower, "both ") || strings.HasPrefix(lower, "either ") || strings.HasPrefix(lower, "neither ") {
		return ""
	}
	if strings.HasPrefix(lower, "it is not the case that ") {
		return policy.scopeOperandFailure(s[len("it is not the case that "):])
	}
	words := strings.Fields(s)
	if !policy.scope_clause_start || len(words) == 0 {
		return ""
	}
	first := strings.ToLower(words[0])
	if scopeClauseWords[first] || unicode.IsDigit(firstRune(first)) {
		return ""
	}
	// A possessive ("Birnbaum's proof", "society's norms") opens a noun
	// phrase, never an adverb or frame. Verb contractions are a closed class.
	if base := strings.TrimSuffix(strings.TrimSuffix(first, "'s"), "’s"); base != first && !scopeContractions[base] {
		return ""
	}
	// A capitalized token alone is not name evidence: require a visible finite
	// predicate after the capitalized name, or a propositional letter.
	if len(words) == 1 && len(words[0]) == 1 {
		return ""
	}
	i := 0
	for i < len(words) && unicode.IsUpper(firstRune(words[i])) && !frameSentenceAdverbs[strings.ToLower(words[i])] && !strings.HasSuffix(strings.ToLower(words[i]), "ly") && !strings.HasSuffix(strings.ToLower(words[i]), "ing") {
		i++
	}
	if i > 0 && i < len(words) && (framePredicates[strings.ToLower(words[i])] || frameIntransitives[strings.ToLower(words[i])] || frameTransitives[strings.ToLower(words[i])] || scopeNameVerbs[strings.ToLower(words[i])]) {
		return ""
	}
	return clauseStartHint
}

// A closed atom literal declares an opaque proposition, not English whose
// first word or internal connectives determine scope. Malformed literals and
// unbound variable arguments do not qualify for this exception.
func scopeLiteralAtom(s string) bool {
	if !strings.HasPrefix(s, "atom{") {
		return false
	}
	_, err := folLiteralAtom(s, nil)
	return err == nil
}

// Restrict the description guard to a wide/narrow conversion pair. Merely
// carrying a wide negative premise into De Morgan or modus tollens is safe.
func descriptionConversionIn(wide string, texts []string) bool {
	body := strings.TrimPrefix(strings.ToLower(wide), "it is not the case that ")
	subject, predicate, ok := strings.Cut(body, " is ")
	if !ok {
		return false
	}
	narrow := subject + " is not " + predicate
	for _, s := range texts {
		if strings.EqualFold(s, narrow) {
			return true
		}
	}
	return false
}

// Predicate lists are opaque when the entire closed-class list survives in
// another statement. This does not license extracting a bare adjective as a
// clause; in that case every operand must satisfy scopeOperandFailure.
func scopePreservedPredicateList(s string, texts []string) bool {
	words := strings.Fields(strings.ToLower(s))
	if len(words) == 0 || strings.HasSuffix(words[0], "ing") || strings.HasSuffix(words[0], "ly") {
		return false
	}
	for i, w := range words {
		if w != "is" && w != "are" {
			continue
		}
		tail := words[i+1:]
		hasAnd, hasOr := false, false
		for _, raw := range tail {
			word := strings.Trim(raw, ",")
			if word == "and" {
				hasAnd = true
				continue
			}
			if word == "or" {
				hasOr = true
				continue
			}
			if !singularPredicateHeads[word] && !scopePredicateListWords[word] {
				return false
			}
		}
		if hasAnd == hasOr {
			return false
		}
		suffix := " " + w + " " + strings.Join(tail, " ")
		for _, other := range texts {
			if other != s && strings.HasSuffix(strings.ReplaceAll(strings.ToLower(other), ",", ""), strings.ReplaceAll(suffix, ",", "")) {
				return true
			}
		}
		return false
	}
	return false
}
