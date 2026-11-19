package validation

import "strings"

// Closed predicate vocabulary avoids suffix homographs ("criminal", "general",
// "cannibal") that may denote nouns rather than adjective predicates.
var coordinatedAdjectives = wordSet("reliable possible feasible risky safe unsafe aligned causal informational dangerous useful happy red blue green warm cold wise rich poor friendly mortal alive dead hungry tired valuable eligible registered ready late early wet dry small large tall short round square honest kind brave light portable")
var coordinatedAdverbs = wordSet("financially very quite extremely deeply highly")

func coordinatedAdjective(w string) bool { return coordinatedAdjectives[w] }

func (policy strictRules) coordinatedPredicate(s string) bool {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		if w != "is" && w != "are" && w != "was" && w != "were" {
			continue
		}
		subject := strings.Join(words[:i], " ")
		// Schematic X/Y/Z subjects use exactly the same singular values when bound.
		schematic := len(subject) == 1 && strings.Contains("uvwxyz", subject)
		if !schematic && !policy.singularValue(subject) {
			return false
		}
		tail := words[i+1:]
		start := 0
		op := ""
		count := 0
		check := func(part []string) bool {
			if len(part) == 2 && coordinatedAdverbs[part[0]] {
				part = part[1:]
			}
			return len(part) == 1 && !policy.hasSlotName(part[0]) && coordinatedAdjective(part[0])
		}
		for j, v := range tail {
			if v != "and" && v != "or" {
				continue
			}
			if op != "" && op != v {
				return false
			}
			if !check(tail[start:j]) {
				return false
			}
			op = v
			count++
			start = j + 1
		}
		return count > 0 && check(tail[start:])
	}
	return false
}
