package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Identity substitution (Leibniz's law) is certified only in known-safe
// positions: (a) when the substituted premise is itself an identity sentence
// (identity chains), or (b) when every occurrence of the replaced term is the
// grammatical subject of a plain clause. Objects of verbs ("Lois is looking
// for Superman", "Lois admires Superman"), passives ("is admired by") and
// raising forms ("is believed to") are not certified. An allowlist of
// positions stays strict where a list of opaque verbs could never be complete.

var clauseSubjectLead = []string{"if", "then", "and", "or", "it is the case that", "it is not the case that", "both that", "and that", "either", "both"}

var passiveOrRaising = regexp.MustCompile(`^\s*(?:is|are|was|were|be|been|being)\s+(?:not\s+)?(?:said|known|thought|believed|rumou?red|reported|alleged|supposed|expected|considered|claimed|assumed|presumed|seen|sought|felt|held|likely|certain|[a-z]+ed|[a-z]+en)\s+(?:by|to)\b`)

// termSubjectPositions reports whether every word-bounded occurrence of term
// in text is the subject of a plain clause, and how many occurrences there are.
func termSubjectPositions(text, term string) (safe bool, count int) {
	lower := strings.ToLower(text)
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return false, 0
	}
	position := 0
	safe = true
	for {
		index := strings.Index(lower[position:], term)
		if index < 0 {
			break
		}
		start := position + index
		end := start + len(term)
		position = start + 1
		if (start > 0 && wordRune(lastRune(lower[:start]))) || (end < len(lower) && wordRune(firstRune(lower[end:]))) {
			continue
		}
		count++
		prefix := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(lower[:start]), ","))
		lead := prefix == "" || strings.HasSuffix(prefix, "(") || strings.HasSuffix(prefix, "such that")
		for _, word := range clauseSubjectLead {
			if prefix == word || strings.HasSuffix(prefix, " "+word) || strings.HasSuffix(prefix, ","+word) {
				lead = true
			}
		}
		if strings.HasSuffix(strings.TrimSpace(lower[:start]), ",") && !lead {
			// A comma-led clause ("..., Superman flies") also starts a clause.
			lead = true
		}
		rest := lower[end:]
		if strings.HasPrefix(rest, "'s ") || strings.HasPrefix(rest, "’s ") {
			// A possessive inside the subject phrase ("Superman's cape is red").
			rest = rest[strings.Index(rest, " "):]
			if idx := strings.Index(rest, " is "); idx >= 0 {
				rest = rest[idx:]
			}
		}
		if !lead || passiveOrRaising.MatchString(rest) {
			safe = false
		}
	}
	return safe && count > 0, count
}

// identitySubstitutionUnsafe returns a reason when a step that substitutes
// identicals does so outside the safe positions, and "" when it is safe.
func identitySubstitutionUnsafe(premises []string, conclusion string) string {
	for i, identityPremise := range premises {
		if !explicitIdentity(identityPremise) {
			continue
		}
		identities := identityCandidates(identityPremise)
		if len(identities) == 0 {
			continue
		}
		for j, other := range premises {
			if j == i {
				continue
			}
			if explicitIdentity(other) && explicitIdentity(conclusion) {
				return "" // identity chain
			}
			otherText := normalizeClausePreservingCase(other)
			for _, identity := range identities {
				for _, d := range [][2]string{{identity.left, identity.right}, {identity.right, identity.left}} {
					if !strings.Contains(strings.ToLower(otherText), strings.ToLower(d[0])) {
						continue
					}
					safe, count := termSubjectPositions(otherText, d[0])
					if count == 0 {
						continue
					}
					conclusionSafe, _ := termSubjectPositions(normalizeClausePreservingCase(conclusion), d[1])
					if safe && conclusionSafe {
						return ""
					}
					return fmt.Sprintf("substitution is certified only in identity sentences and in the subject of a plain sentence; %q here is not the subject of a plain clause (an object, passive or raising position). If the step is sound, state it as a premise", d[0])
				}
			}
		}
	}
	return ""
}

func stepStatements(steps []NumberedArgumentStep) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.Statement
	}
	return out
}

var explicitIdentityPattern = regexp.MustCompile(`(?i)\b(?:is|are|was|were) (?:not )?(?:identical (?:to|with)|the same (?:person|thing|object|individual) as|one and the same as)\b|\s=\s`)

// explicitIdentity reports an identity sentence written as one ("A is
// identical to B"), as opposed to any copular "A is B".
func explicitIdentity(statement string) bool {
	return explicitIdentityPattern.MatchString(statement)
}
