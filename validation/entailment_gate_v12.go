package validation

import (
	"fmt"
	"strings"
)

// Catalog v12 propositional forms: constructive, simple and destructive
// dilemmas, binary and n-ary. Each is appended after the v11 catalog so no
// previously certified step changes form.
//
// Shape shared by all three: exactly one cited premise is an "Either ... or
// ..." disjunction and every other cited premise is an "If ..., then ..."
// conditional. Citation order is free. Each conditional must serve exactly one
// disjunct (a bijection), so a dilemma with n disjuncts cites n conditionals
// and nothing else. Clauses compare verbatim after the gate's surface
// normalisation; negation is "it is not the case that", as in modus tollens.

const maxDilemmaDisjuncts = 6

type dilemmaKind int

const (
	constructiveDilemma dilemmaKind = iota
	simpleDilemma
	destructiveDilemma
)

func matchConstructiveDilemma(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	return matchDilemma(premises, conclusion, constructiveDilemma)
}

func matchSimpleDilemma(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	return matchDilemma(premises, conclusion, simpleDilemma)
}

func matchDestructiveDilemma(premises []NumberedArgumentStep, conclusion NumberedArgumentStep) schemaMatchOutcome {
	return matchDilemma(premises, conclusion, destructiveDilemma)
}

// eitherDisjunctPartitions splits an explicit "Either D1 or D2 ... or Dn"
// statement into exactly n disjuncts at " or " boundaries, enumerating every
// choice of boundaries (disjunction is associative, so any grouping is a
// sound reading of the top-level either-or). The leading "either" is required:
// it fixes the disjunction's scope at the whole statement, so "It is not the
// case that A or B" is never read as a disjunction here.
func eitherDisjunctPartitions(statement string, partCount int) [][]string {
	statement = normalizeClause(statement)
	if !strings.HasPrefix(statement, "either ") {
		return nil
	}
	body := statement[len("either "):]
	const separator = " or "
	count := strings.Count(body, separator)
	if count > 12 || partCount < 2 || partCount > count+1 {
		return nil
	}
	var boundaries []int
	for start := 0; ; {
		index := strings.Index(body[start:], separator)
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
			last := normalizeClause(body[start:])
			if last == "" {
				return
			}
			result = append(result, append(append([]string(nil), parts...), last))
			return
		}
		for i := next; i < len(boundaries); i++ {
			if len(boundaries)-i < partCount-1-len(parts) {
				break
			}
			part := normalizeClause(body[start:boundaries[i]])
			if part == "" {
				continue
			}
			visit(boundaries[i]+len(separator), i+1, append(parts, part))
		}
	}
	visit(0, 0, nil)
	return result
}

// negatedDisjuncts maps each disjunct to its negated clause, or returns false
// when any disjunct is not an explicit negation.
func negatedDisjuncts(parts []string) ([]string, bool) {
	result := make([]string, len(parts))
	for i, part := range parts {
		negated := negationCandidates(part)
		if len(negated) == 0 {
			return nil, false
		}
		result[i] = negated[0]
	}
	return result, true
}

func matchDilemma(premises []NumberedArgumentStep, conclusion NumberedArgumentStep, kind dilemmaKind) schemaMatchOutcome {
	n := len(premises) - 1
	if n < 2 || n > maxDilemmaDisjuncts {
		return schemaMatchOutcome{}
	}
	outcome := schemaMatchOutcome{}
	for d := range premises {
		conditionals := make([][]clausePair, 0, n)
		numbers := make([]int, 0, n)
		for i, premise := range premises {
			if i == d {
				continue
			}
			candidates := conditionalCandidates(premise.Statement)
			if len(candidates) == 0 {
				conditionals = nil
				break
			}
			conditionals = append(conditionals, candidates)
			numbers = append(numbers, premise.Number)
		}
		if conditionals == nil {
			continue
		}
		disjunctions := eitherDisjunctPartitions(premises[d].Statement, n)
		if len(disjunctions) == 0 {
			continue
		}
		// The premise disjunction and the cited conditionals parse: this step
		// has dilemma shape, so a failure below is a near-miss worth naming;
		// the shape alone earns one point so it outranks shapeless near-misses.
		outcome.traction = true
		outcome.score = 1

		var conclusionParts [][]string
		switch kind {
		case simpleDilemma:
			whole := normalizeClause(conclusion.Statement)
			parts := make([]string, n)
			for i := range parts {
				parts[i] = whole
			}
			conclusionParts = [][]string{parts}
		default:
			conclusionParts = eitherDisjunctPartitions(conclusion.Statement, n)
		}

		for _, disjuncts := range disjunctions {
			premiseSide := disjuncts
			if kind == destructiveDilemma {
				negated, ok := negatedDisjuncts(disjuncts)
				if !ok {
					continue
				}
				premiseSide = negated
			}
			for _, results := range conclusionParts {
				resultSide := results
				if kind == destructiveDilemma {
					negated, ok := negatedDisjuncts(results)
					if !ok {
						continue
					}
					resultSide = negated
				}
				if dilemmaBijection(conditionals, premiseSide, resultSide, kind, 0, make([]bool, n), make([]bool, n)) {
					return schemaMatchOutcome{matched: true, score: n + 1, traction: true}
				}
			}
		}
		if outcome.mismatch == "" {
			outcome.mismatch = dilemmaMismatch(kind, premises[d].Number, numbers, conclusion.Number)
		}
	}
	return outcome
}

// dilemmaBijection assigns conditional j (and every later one) to an unused
// premise-side clause and an unused conclusion-side clause.
//
//   - constructive/simple: antecedent = premise disjunct, consequent =
//     conclusion disjunct (for simple, every conclusion slot is the whole
//     conclusion).
//   - destructive: consequent = clause negated in a premise disjunct,
//     antecedent = clause negated in a conclusion disjunct.
func dilemmaBijection(conditionals [][]clausePair, premiseSide, resultSide []string, kind dilemmaKind, j int, usedPremise, usedResult []bool) bool {
	if j == len(conditionals) {
		return true
	}
	for _, conditional := range conditionals[j] {
		fromPremise, toResult := conditional.left, conditional.right
		if kind == destructiveDilemma {
			fromPremise, toResult = conditional.right, conditional.left
		}
		for p := range premiseSide {
			if usedPremise[p] || !equalClauses(fromPremise, premiseSide[p]) {
				continue
			}
			for r := range resultSide {
				if usedResult[r] || !equalClauses(toResult, resultSide[r]) {
					continue
				}
				usedPremise[p], usedResult[r] = true, true
				if dilemmaBijection(conditionals, premiseSide, resultSide, kind, j+1, usedPremise, usedResult) {
					return true
				}
				usedPremise[p], usedResult[r] = false, false
			}
		}
	}
	return false
}

func dilemmaMismatch(kind dilemmaKind, disjunction int, conditionals []int, conclusion int) string {
	cited := make([]string, len(conditionals))
	for i, number := range conditionals {
		cited[i] = fmt.Sprint(number)
	}
	list := strings.Join(cited, ", ")
	switch kind {
	case simpleDilemma:
		return fmt.Sprintf("conditionals %s do not each take one disjunct of premise %d to conclusion %d verbatim", list, disjunction, conclusion)
	case destructiveDilemma:
		return fmt.Sprintf("premise %d does not deny each consequent of conditionals %s exactly once, with conclusion %d denying their antecedents", disjunction, list, conclusion)
	default:
		return fmt.Sprintf("conditionals %s do not each take one disjunct of premise %d to one disjunct of conclusion %d verbatim", list, disjunction, conclusion)
	}
}
