package validation

import "testing"

func TestStrictNotesV45RefusalHits(t *testing.T) {
	for _, words := range [][]string{
		{"bob", "dog"},              // A name inside a description.
		{"brick", "garden", "wall"}, // v46.0: noun modifiers still fail; muddy/wooden are allowlisted adjectives.
		{"big", "red", "sees"},      // A finite verb as the head.
	} {
		p, fired := allStrictRules().recording()
		names := map[string]bool{"bob": true}
		p.slotNames = &names
		if p.descriptionWords(words, false) || !fired.ids["slot.description_shape"] {
			t.Errorf("description refusal was not recorded: %v", words)
		}
	}
	p, fired := allStrictRules().recording()
	if p.scopeOperandFailure("perhaps the lamp is lit") == "" || !fired.ids["scope.clause_start"] {
		t.Error("clause-start refusal was not recorded")
	}
}

func TestStrictNotesV45Parity(t *testing.T) {
	for _, value := range []string{"the Bob dog", "the muddy wooden cart", "the big red sees", "the group of robots", "Carlsmith's APS concept"} {
		text := "1. If X is red, then X is warm.\n2. " + value + " is red.\n3. Therefore, " + value + " is warm. [1,2]"
		notesParity(t, value, text)
	}
	notesParity(t, "plural copula", "1. If X is dangerous, then X should be studied.\n2. Subgoals are dangerous.\n3. Therefore, subgoals should be studied. [1,2]")
}
