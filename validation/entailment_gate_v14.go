package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Scoped schemas consume ranges, never a flattened collection of their lines.
// The parser establishes ownership; GateNumberedArgument checks it again for
// callers constructing an AST directly. Inner conclusions are gated normally.
var fitchForms = []string{"conditional proof", "reductio", "proof by cases", "universal generalization", "existential elimination", "relational universal generalization", "relational existential elimination"}

func matchContradiction(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	if normalizeClause(c.Statement) != "contradiction" {
		return schemaMatchOutcome{}
	}
	for _, n := range negationCandidates(p[1].Statement) {
		if equalClauses(p[0].Statement, n) {
			return schemaMatchOutcome{matched: true}
		}
	}
	return schemaMatchOutcome{}
}

func fitchAssumptionStatement(s NumberedArgumentStep) string {
	if s.Kind == numberedArgumentStepArbitrary {
		if m := arbitraryFOLPattern.FindStringSubmatch(s.Statement); m != nil {
			return m[1] + " is identical to " + m[1]
		}
		if m := arbitraryPropertyPattern.FindStringSubmatch(s.Statement); m != nil {
			if propertyBearer(m[2]) {
				return m[2] + " has " + m[1]
			}
			return s.Statement
		}
		if m := arbitraryNamePattern.FindStringSubmatch(s.Statement); m != nil {
			return m[1] + " is a " + m[2]
		}
	}
	return s.Statement
}

func gateFitchConclusion(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	report := EntailmentGateStepReport{Step: c.Number, Depth: c.Depth, Citations: c.Citations, Ranges: c.Ranges, NearestForm: "Fitch discharge", Mismatch: "ranges and cited premises do not match a v14 discharge rule"}
	accept := func(form string) EntailmentGateStepReport {
		report.Passed = true
		report.Form = form
		report.NearestForm = ""
		report.Mismatch = ""
		return report
	}
	if len(c.Ranges) == 1 && len(c.Citations) == 0 {
		r := c.Ranges[0]
		first, last := steps[r[0]], steps[r[1]]
		if first.Kind == numberedArgumentStepAssumption {
			if reason := v19PolarityFailure(nil, c); reason != "" {
				report.NearestForm = "conditional proof"
				report.Mismatch = "conditional proof refused: " + reason + ". " + counterfactualAvailable
				// Reductio below may still establish this whole proposition; it
				// does not interpret the counterfactual as a material conditional.
			}
			for _, pair := range v13Pairs("conditional", c.Statement) {
				if equalClauses(first.Statement, pair.left) && equalClauses(last.Statement, pair.right) {
					if v19PolarityFailure(nil, c) == "" {
						return accept("conditional proof")
					}
				}
			}
			if normalizeClause(last.Statement) == "contradiction" {
				for _, n := range negationCandidates(c.Statement) {
					if equalClauses(n, first.Statement) {
						return accept("reductio")
					}
				}
				for _, n := range negationCandidates(first.Statement) {
					if equalClauses(n, c.Statement) {
						return accept("reductio")
					}
				}
			}
		}
		if first.Kind == numberedArgumentStepArbitrary {
			if arbitraryPropertyPattern.MatchString(first.Statement) {
				report.NearestForm = "property generalization"
				ok, reason := propertyGeneralization(first, last, c, steps)
				if ok {
					return accept("property generalization")
				}
				report.Mismatch = reason
				return report
			}
			report.NearestForm = "universal generalization"
			m := arbitraryNamePattern.FindStringSubmatch(first.Statement)
			if m != nil && relativeMarker(normalizeClause(m[2])) {
				if q, ok := relativeQuant(c.Statement); ok && q.kind == "every" {
					guards, good := relativeNP(normalizeClause(m[2]), relativeHole)
					if good && relativeAll(guards, q.guards) && relativeAll(q.guards, guards) && equalClauses(last.Statement, relativeFill(q.body, normalizeClause(m[1]))) {
						if reason := fitchFreshName(m[1], first.Number, c, steps); reason != "" {
							report.Mismatch = reason
							return report
						}
						return accept("universal generalization")
					}
				}
			}
			if m != nil && v13PlainClause(m[2]) {
				if reason := fitchFreshName(m[1], first.Number, c, steps); reason != "" {
					report.Mismatch = reason
					return report
				}
				if matchRelationalUG(last, c, m[1], m[2]) {
					return accept("relational universal generalization")
				}
				// The eigenname must be the verbatim subject, not merely an occurrence
				// somewhere inside a copular clause or a case-normalized substitute.
				prefix := m[1] + " is "
				if strings.HasPrefix(last.Statement, prefix) {
					predicate := strings.TrimPrefix(last.Statement, prefix)
					if v13PlainClause(predicate) && equalClauses(c.Statement, "Every "+m[2]+" is "+predicate) {
						return accept("universal generalization")
					}
				}
			}
			report.Mismatch = "arbitrary class and final copular predicate must match Every F is G verbatim"
		}
	}
	if len(c.Citations) == 1 && len(c.Ranges) >= 2 {
		report.NearestForm = "proof by cases"
		disjunction := steps[c.Citations[0]].Statement
		for _, parts := range eitherDisjunctPartitions(disjunction, len(c.Ranges)) {
			used := make([]bool, len(parts))
			matched := true
			for _, r := range c.Ranges {
				first, last := steps[r[0]], steps[r[1]]
				if first.Kind != numberedArgumentStepAssumption || !equalClauses(last.Statement, c.Statement) {
					matched = false
					break
				}
				found := false
				for i, p := range parts {
					if !used[i] && equalClauses(first.Statement, p) {
						used[i] = true
						found = true
						break
					}
				}
				if !found {
					matched = false
					break
				}
			}
			if matched {
				return accept("proof by cases")
			}
		}
		report.Mismatch = "each disjunct must have one complete branch reaching the stated conclusion"
	}
	if len(c.Citations) == 1 && len(c.Ranges) == 1 {
		report.NearestForm = "existential elimination"
		r := c.Ranges[0]
		first, last := steps[r[0]], steps[r[1]]
		if first.Kind == numberedArgumentStepAssumption && equalClauses(last.Statement, c.Statement) {
			if name := relativeEEName(first, steps[c.Citations[0]]); name != "" {
				if reason := fitchFreshName(name, first.Number, c, steps); reason != "" {
					report.Mismatch = reason
					return report
				}
				return accept("existential elimination")
			}
			if name := relationalEEName(first, steps[c.Citations[0]]); name != "" {
				if reason := fitchFreshName(name, first.Number, c, steps); reason != "" {
					report.Mismatch = reason
					return report
				}
				return accept("relational existential elimination")
			}
			// Reuse the existing existential introduction grammar in reverse: the
			// assumption supplies exactly the witness's membership and predicates.
			for _, pair := range conjunctionCandidates(first.Statement) {
				left := groundInstanceCandidates(pair.left)
				right := groundInstanceCandidates(pair.right)
				if len(left) != 1 || len(right) != 1 || !v13PlainClause(left[0].right) || !v13PlainClause(right[0].right) || !equalClauses(left[0].left, right[0].left) {
					continue
				}
				name := left[0].left
				if !fitchNamePattern.MatchString(name) {
					continue
				}
				witness := []NumberedArgumentStep{{Statement: pair.left}, {Statement: pair.right}}
				if !matchExistentialIntroduction(witness, steps[c.Citations[0]]).matched {
					continue
				}
				if reason := fitchFreshName(name, first.Number, c, steps); reason != "" {
					report.Mismatch = reason
					return report
				}
				return accept("existential elimination")
			}
		}
		report.Mismatch = "existential witness assumption and branch conclusion must match the cited existential and conclusion"
	}
	return report
}

var fitchNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func fitchFreshName(name string, start int, conclusion NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	match := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(name) + `\b`).MatchString
	contains := func(s string) bool {
		return match(s) || match(folDecodedLiteralText(s)) || folLiteralHasName(s, strings.ToLower(name))
	}
	for n := 1; n < start; n++ {
		if contains(steps[n].Statement) {
			return fmt.Sprintf("fresh name %q at step %d already occurs in earlier step %d", name, start, n)
		}
	}
	if contains(conclusion.Statement) {
		return fmt.Sprintf("fresh name %q from step %d occurs in conclusion step %d", name, start, conclusion.Number)
	}
	return ""
}
