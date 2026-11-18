package validation

import (
	"fmt"
	"sort"
	"strings"
)

// Ten conjuncts are supported without allocating a factorial list of assignments.
const schematicConjunctLimit = 10

func schematicBindingText(state schematicSubstitution) string {
	var keys []int
	for k := range state {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%c=%q", rune(k), state[rune(k)]))
	}
	if len(parts) == 0 {
		return "{}"
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
func schematicDisplay(template string) string {
	return strings.Map(func(r rune) rune {
		if v, ok := schematicVariableForMarker(r); ok {
			return v
		}
		return r
	}, template)
}

// A branch uses each cited line at most once. Equivalent states with the same
// remaining premises have identical continuations; memoization changes no verdict.
func (policy strictRules) searchSchematicConjuncts(ps []NumberedArgumentStep, ci int, ant, cons string, conjuncts []string, c NumberedArgumentStep) schemaMatchOutcome {
	n := len(conjuncts)
	if n > len(ps)-1 {
		return schemaMatchOutcome{traction: true, mismatch: fmt.Sprintf("premise %d antecedent has %d conjuncts but only %d other cited premises", ps[ci].Number, n, len(ps)-1)}
	}
	best := schemaMatchOutcome{}
	used := map[int]bool{ps[ci].Number: true}
	seen := map[string]bool{}
	var visit func(int, schematicSubstitution) bool
	visit = func(at int, state schematicSubstitution) bool {
		var ids []int
		for id := range used {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		key := fmt.Sprintf("%d/%v/%s", at, ids, schematicBindingText(state))
		if seen[key] {
			return false
		}
		seen[key] = true
		candidate := schemaMatchOutcome{traction: true, score: at}
		if at == n {
			want, ok := instantiateSchematicTemplate(cons, state)
			_, antOK := instantiateSchematicTemplate(ant, state)
			if !ok || !antOK {
				candidate.mismatch = "unbound schematic variable under binding " + schematicBindingText(state)
			} else if reason := repeatedSlotMismatch(ant, cons, state); reason != "" {
				candidate.mismatch = reason
			} else {
				appendOutcome(&candidate, policy.evaluateIdentities(clauseIdentity{fmt.Sprintf("premise %d instantiated consequent", ps[ci].Number), want, fmt.Sprintf("conclusion %d", c.Number), c.Statement}))
				if candidate.mismatch == "" && n+1 != len(ps) {
					candidate.mismatch = "cited premises unused by the conjunctive antecedent"
				}
				candidate.matched = candidate.mismatch == ""
			}
			best = betterOutcome(best, candidate)
			return candidate.matched
		}
		found := false
		for _, p := range ps {
			if used[p.Number] {
				continue
			}
			for _, next := range policy.unifySchematicTemplate(conjuncts[at], p.Statement, relaxedInitial(state, p)) {
				found = true
				used[p.Number] = true
				if visit(at+1, next) {
					return true
				}
				delete(used, p.Number)
			}
		}
		if !found {
			candidate.mismatch = fmt.Sprintf("premise %d antecedent conjunct %q has no matching unused cited premise under binding %s", ps[ci].Number, schematicDisplay(conjuncts[at]), schematicBindingText(state))
			best = betterOutcome(best, candidate)
		}
		return false
	}
	visit(0, schematicSubstitution{})
	return best
}

// New identity surfaces are admitted only in complete argument positions.
// Masking the entire stipulated term keeps gerund/coordination vocabulary out
// of the surrounding opacity check; it never exposes a term inside an attitude.
func v28IdentityTerm(term string) bool {
	// v28.2: check quantifier words case-insensitively. The case-preserving split
	// handed "Nobody" and "Some cat" to a lowercase word list, so they passed as
	// names ("Nobody is identical to Felix" + "Nobody is black" => "Felix is black").
	if quantifiedIdentitySide(term) {
		return false
	}
	if substitutionTerm(term, false) {
		return true
	}
	s := normalizeClause(term)
	words := strings.Fields(s)
	if len(words) < 2 {
		return false
	}
	if strings.HasSuffix(words[0], "ing") {
		for _, w := range words[1:] {
			if nonReferringTermWords[w] && w != "a" && w != "an" {
				return false
			}
		}
		return !strings.ContainsAny(s, "\"`“”‘,;:()")
	}
	// A coordinated proper name is treated as one name by the explicit identity.
	// Finite clauses and quantified coordinands are not names.
	if strings.Contains(s, " and ") {
		for _, part := range strings.Split(term, " and ") {
			for _, w := range strings.Fields(part) {
				if !startsCapitalizedWord(w) {
					return false
				}
			}
			if !referringTerm(normalizeClause(part), false) {
				return false
			}
		}
		return true
	}
	return false
}

func v28IdentityContext(statement, term string) bool {
	masked, count, ok := replaceTermOccurrences(statement, term, "v28term")
	if !ok || count != 1 {
		return false
	}
	s := normalizeClause(masked)
	_, quantified := transparentContext(s)
	if quantified {
		first, _, _ := strings.Cut(normalizeClause(term), " ")
		switch first {
		case "the", "this", "that", "these", "those", "such":
			return false
		}
	}
	// All surrounding vocabulary remains subject to the historical opacity rule.
	// Bare adjectival necessity here predicates of the subject, not of a clause.
	if strings.HasPrefix(s, "v28term is ") {
		predicate := strings.TrimPrefix(s, "v28term is ")
		parts := strings.Split(predicate, " and ")
		if len(parts) > 1 {
			for i, part := range parts {
				adjective := part
				for _, modifier := range []string{"not ", "always ", "often ", "sometimes "} {
					adjective = strings.TrimPrefix(adjective, modifier)
				}
				if adjective == "necessary" || adjective == "possible" {
					parts[i] = "v28adjective"
				}
			}
			if safe, _ := transparentContext("v28term is " + strings.Join(parts, " and ")); safe {
				return true
			}
		}
		safe, _ := transparentContext(s)
		return safe
	}
	if !strings.HasSuffix(s, " v28term") {
		return false
	}
	safe, _ := transparentContext(s)
	if !safe {
		return false
	}
	prefix := strings.TrimSuffix(s, " v28term")
	// A small transitive register prevents masking predicate vocabulary or a
	// quantifier restrictor as though it were an individual argument.
	for _, verb := range []string{"deserves", "admires", "praised", "praises", "helps", "helped", "receives help from"} {
		subject, found := strings.CutSuffix(prefix, " "+verb)
		if !found {
			continue
		}
		for _, q := range []string{"every ", "some ", "no "} {
			subject = strings.TrimPrefix(subject, q)
		}
		if referringTerm(subject, false) {
			return true
		}
	}
	return false
}

func (policy strictRules) matchV28Identity(ps []NumberedArgumentStep, c NumberedArgumentStep) bool {
	if len(ps) != 2 {
		return false
	}
	for i := range ps {
		identity := normalizeClausePreservingCase(ps[i].Statement)
		for _, sep := range []string{" is identical to ", " is identical with "} {
			pairs := appendSplitCandidatesPreservingCase(nil, identity, sep, false)
			if len(pairs) != 1 {
				continue
			}
			pair := pairs[0]
			if !v28IdentityTerm(pair.left) || !v28IdentityTerm(pair.right) {
				continue
			}
			for _, d := range [][2]string{{pair.left, pair.right}, {pair.right, pair.left}} {
				premise := normalizeClausePreservingCase(ps[1-i].Statement)
				target := normalizeClausePreservingCase(c.Statement)
				if !v28IdentityContext(premise, d[0]) || !v28IdentityContext(target, d[1]) {
					continue
				}
				replaced, n, ok := replaceTermOccurrences(premise, d[0], d[1])
				if ok && n > 0 && policy.equalClauses(replaced, target) {
					return true
				}
			}
		}
	}
	return false
}

func (policy strictRules) extendedSchematicTemplate(value string) (string, schematicVariableCounts) {
	replaced, counts := replaceSchematicAlphabet(value, true)
	return policy.comparisonClause(replaced), counts
}

// v28 preserves every successful historical certificate before trying the new
// whole-term identity register. Ranged and explicit proofs keep their own policy.
func gateConclusionStep(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	return allStrictRules().gateConclusionStep(c, steps)
}

func (policy strictRules) gateConclusionStep(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	report := policy.gateConclusionStepStrict(c, steps)
	if report.Passed {
		return report
	}
	// task_9b65b028: if the step passes only when an indefinite may bind a
	// schematic variable in an object slot, say so instead of reporting
	// whichever other form came nearest.
	relaxed := make(map[int]NumberedArgumentStep, len(steps))
	for id, step := range steps {
		step.relaxIndefinite = policy.slot_indefinite_object
		relaxed[id] = step
	}
	if retry := policy.gateConclusionStepStrict(c, relaxed); retry.Passed {
		report.NearestForm = retry.Form
		report.Mismatch = policy.indefiniteSlotStepMessage(c, steps)
		return report
	}
	// task_eb737ad9: a term refused because a word in it can be read as a verb
	// stays refused, but the message names the binding.
	clauseRelaxed := make(map[int]NumberedArgumentStep, len(steps))
	for id, step := range steps {
		step.relaxClauseTerm = policy.term_clause_inside
		clauseRelaxed[id] = step
	}
	if retry := policy.gateConclusionStepStrict(c, clauseRelaxed); retry.Passed {
		if message := policy.clauseTermStepMessage(c, steps); message != "" {
			report.NearestForm = retry.Form
			report.Mismatch = message
		}
	}
	return report
}

func (policy strictRules) gateConclusionStepStrict(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	old := policy.gateConclusionStepV27(c, steps)
	if old.Passed {
		return old
	}
	if len(c.Ranges) == 0 && !folExplicit(c.Statement) {
		var ps []NumberedArgumentStep
		for _, id := range c.Citations {
			p := steps[id]
			p.Statement = fitchAssumptionStatement(p)
			ps = append(ps, p)
		}
		if policy.matchV28Identity(ps, c) {
			if reason := policy.identitySubstitutionUnsafe(stepStatements(ps), c.Statement); reason != "" {
				old.NearestForm, old.Mismatch = "identity substitution", reason
				return old
			}
			old.Passed = true
			old.Form = "identity substitution"
			old.NearestForm = ""
			old.Mismatch = ""
			return old
		}
	}
	return old
}
