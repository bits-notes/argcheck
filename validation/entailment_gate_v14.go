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

func (policy strictRules) matchContradiction(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	if normalizeClause(c.Statement) != "contradiction" {
		return schemaMatchOutcome{}
	}
	for _, n := range negationCandidates(p[1].Statement) {
		if policy.equalClauses(p[0].Statement, n) {
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

func (policy strictRules) gateFitchConclusion(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	report := EntailmentGateStepReport{Step: c.Number, Depth: c.Depth, Citations: c.Citations, Ranges: c.Ranges, NearestForm: "Fitch discharge", Mismatch: "ranges and cited premises do not match a v14 discharge rule"}
	accept := func(form string) EntailmentGateStepReport {
		report.Passed = true
		report.Form = form
		report.Paraphrase = readingScoped(form, c, steps)
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
				if policy.equalClauses(first.Statement, pair.left) && policy.equalClauses(last.Statement, pair.right) {
					if v19PolarityFailure(nil, c) == "" {
						return accept("conditional proof")
					}
				}
			}
			if normalizeClause(last.Statement) == "contradiction" {
				for _, n := range negationCandidates(c.Statement) {
					if policy.equalClauses(n, first.Statement) {
						return accept("reductio")
					}
				}
				for _, n := range negationCandidates(first.Statement) {
					if policy.equalClauses(n, c.Statement) {
						return accept("reductio")
					}
				}
			}
		}
		if first.Kind == numberedArgumentStepArbitrary {
			if arbitraryPropertyPattern.MatchString(first.Statement) {
				report.NearestForm = "property generalization"
				ok, reason := policy.propertyGeneralization(first, last, c, steps)
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
					if good && policy.relativeAll(guards, q.guards) && policy.relativeAll(q.guards, guards) && policy.equalClauses(last.Statement, relativeFill(q.body, normalizeClause(m[1]))) {
						if reason := policy.fitchFreshName(m[1], first.Number, c, steps); reason != "" {
							report.Mismatch = reason
							return report
						}
						return accept("universal generalization")
					}
				}
			}
			if m != nil && v13PlainClause(m[2]) {
				if reason := policy.fitchFreshName(m[1], first.Number, c, steps); reason != "" {
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
					if v13PlainClause(predicate) && policy.equalClauses(c.Statement, "Every "+m[2]+" is "+predicate) {
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
				if first.Kind != numberedArgumentStepAssumption || !policy.equalClauses(last.Statement, c.Statement) {
					matched = false
					break
				}
				found := false
				for i, p := range parts {
					if !used[i] && policy.equalClauses(first.Statement, p) {
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
		if first.Kind == numberedArgumentStepAssumption && policy.equalClauses(last.Statement, c.Statement) {
			if name := policy.relativeEEName(first, steps[c.Citations[0]]); name != "" {
				if reason := policy.fitchFreshName(name, first.Number, c, steps); reason != "" {
					report.Mismatch = reason
					return report
				}
				return accept("existential elimination")
			}
			if name := policy.relationalEEName(first, steps[c.Citations[0]]); name != "" {
				if reason := policy.fitchFreshName(name, first.Number, c, steps); reason != "" {
					report.Mismatch = reason
					return report
				}
				return accept("relational existential elimination")
			}
			// Reuse the existing existential introduction grammar in reverse: the
			// assumption supplies exactly the witness's membership and predicates.
			for _, pair := range conjunctionCandidates(first.Statement) {
				left := policy.groundInstanceCandidates(pair.left)
				right := policy.groundInstanceCandidates(pair.right)
				if len(left) != 1 || len(right) != 1 || !v13PlainClause(left[0].right) || !v13PlainClause(right[0].right) || !policy.equalClauses(left[0].left, right[0].left) {
					continue
				}
				name := left[0].left
				if !fitchNamePattern.MatchString(name) {
					continue
				}
				witness := []NumberedArgumentStep{{Statement: pair.left}, {Statement: pair.right}}
				if !policy.matchExistentialIntroduction(witness, steps[c.Citations[0]]).matched {
					continue
				}
				if reason := policy.fitchFreshName(name, first.Number, c, steps); reason != "" {
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

func (policy strictRules) fitchFreshName(name string, start int, conclusion NumberedArgumentStep, steps map[int]NumberedArgumentStep) string {
	match := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(name) + `\b`).MatchString
	contains := func(s string) bool {
		return match(s) || match(folDecodedLiteralText(s)) || policy.folLiteralHasName(s, strings.ToLower(name))
	}
	for n := 1; n < start; n++ {
		if contains(steps[n].Statement) {
			return fmt.Sprintf("fresh name %q at step %d already occurs in earlier step %d", name, start, n)
		}
	}
	// v28 permits later root premises. They must not smuggle an eigenname
	// into the existential premise used to discharge an earlier witness block.
	for n := start; n < conclusion.Number; n++ {
		if steps[n].Kind == numberedArgumentStepPremise && contains(steps[n].Statement) {
			return fmt.Sprintf("fresh name %q from step %d occurs in open premise %d", name, start, n)
		}
	}
	if contains(conclusion.Statement) {
		return fmt.Sprintf("fresh name %q from step %d occurs in conclusion step %d", name, start, conclusion.Number)
	}
	return ""
}
