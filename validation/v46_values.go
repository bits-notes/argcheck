package validation

import (
	"strings"

	"github.com/bits-notes/english-arg-validator/validation/lexicon"
)

func (p strictRules) pluralMassValue(value string) bool {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	if len(words) < 1 || len(words) > 3 || strings.ContainsAny(value, ",;:'’\"") {
		return false
	}
	head := words[len(words)-1]
	if p.hasSlotName(head) || singularClosedWord(head) || !lexicon.IsPluralMassNoun(head) {
		return false
	}
	for _, w := range words[:len(words)-1] {
		if singularClosedWord(w) || !lexicon.IsAdjective(w) {
			return false
		}
	}
	return true
}

// This check runs at every occurrence, including consequent instantiation.
func (p strictRules) termValueAt(template []rune, ti int, value string) bool {
	if p.singularOnlyValue(value) {
		return true
	}
	if !p.slot_plural_mass_value || !p.pluralMassValue(value) {
		return false
	}
	// A bare value cannot become either side of an identity, even in a schema.
	if p.identity_constant_sides {
		for _, pair := range p.round2IdentityPairs(string(template)) {
			marker := string(template[ti])
			if strings.Contains(pair.left, marker) || strings.Contains(pair.right, marker) {
				p.hit("identity.constant_sides")
				return false
			}
		}
	}
	v, _ := schematicVariableForMarker(template[ti])
	if p.slot_plural_witness && p.pluralWitnessBlocked[v] {
		p.hit("slot.plural_witness")
		return false
	}
	after := strings.Fields(strings.TrimSpace(string(template[ti+1:])))
	if len(after) == 0 {
		return true
	}
	verb := strings.ToLower(after[0])
	words := strings.Fields(strings.ToLower(value))
	if lexicon.IsMassNoun(words[len(words)-1]) {
		return verb != "are" && verb != "were" && verb != "have" && verb != "do"
	}
	singularVerb := verb == "is" || verb == "was" || verb == "has" || verb == "does" ||
		(strings.HasSuffix(verb, "s") && (framePredicates[verb] || frameTransitives[verb] || frameIntransitives[verb] || scopeNameVerbs[verb]))
	return !singularVerb
}

// Conservative per-step witness guard. Repeated antecedent occurrences are
// refused even if nested scope would make them harmless. Multiple non-rule
// citations are refused for bare values because no portion identity is known.
func rulePluralWitnessBlocks(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) [256]bool {
	out := [256]bool{}
	facts := 0
	for _, id := range c.Citations {
		candidates := conditionalCandidatesPreservingCase(steps[id].Statement)
		if len(candidates) == 0 {
			facts++
		}
		for _, cond := range candidates {
			_, counts := replaceSchematicAlphabet(cond.left, true)
			dangerous := false
			for _, w := range strings.Fields(strings.ToLower(cond.right)) {
				w = strings.Trim(w, ".,;:!?()\"")
				if w == "same" || w == "both" || w == "together" {
					dangerous = true
				}
			}
			for v, count := range counts {
				if count > 1 || dangerous {
					out[v] = true
				}
			}
		}
	}
	if facts > 1 {
		for i := range out {
			out[i] = true
		}
	}
	return out
}
