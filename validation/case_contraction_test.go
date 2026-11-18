package validation

import "testing"

func TestCaseContractions(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		valid      bool
	}{
		{"modus tollens", "1. If it rains, then the game is off.\n2. It's not the case that the game is off.\n3. Therefore, it is not the case that it rains. [1,2]", true},
		{"curly apostrophe", "1. If it rains, then the game is off.\n2. It’s not the case that the game is off.\n3. Therefore, it’s not the case that it rains. [1,2]", true},
		{"positive wrapper", "1. It's the case that it rains.\n2. If it rains, then the game is off.\n3. Therefore, the game is off. [1,2]", true},
		{"conclusion contracted", "1. If it rains, then the game is off.\n2. It is not the case that the game is off.\n3. Therefore, it's not the case that it rains. [1,2]", true},
		{"double negation", "1. It's not the case that it's not the case that it rains.\n2. Therefore, it rains. [1]", true},
		{"in an antecedent", "1. If it's not the case that it rains, then the game is on.\n2. It is not the case that it rains.\n3. Therefore, the game is on. [1,2]", true},
		{"other it's stays atomic", "1. If it is raining, then the game is off.\n2. It's raining.\n3. Therefore, the game is off. [1,2]", false},
		{"negation still matters", "1. If it rains, then the game is off.\n2. It's the case that the game is off.\n3. Therefore, it's not the case that it rains. [1,2]", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, result := CheckEntailmentGate(tc.text)
			if result.Passed != tc.valid {
				t.Fatalf("got %v: %s", result.Passed, result.Message)
			}
		})
	}
}
