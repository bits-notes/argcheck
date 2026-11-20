package validation

import "testing"

// Pin the concrete examples and repairs in coverage.md's Known limits table.
func TestFinalDocumentedLimits(t *testing.T) {
	cases := []struct{ name, refused, repaired string }{
		// v46.1 refuses bare mass values, whether allowlisted or unlisted.
		{"mass", "1. If W is useful, then W is valuable.\n2. Silt is useful.\n3. Therefore, silt is valuable. [1,2]", "1. If silt is useful, then silt is valuable.\n2. Silt is useful.\n3. Therefore, silt is valuable. [1,2]"},
		{"long noun phrase", "1. If W are useful, then W are valuable.\n2. Signals from the model are useful.\n3. Therefore, signals from the model are valuable. [1,2]", "1. If signals from the model are useful, then signals from the model are valuable.\n2. Signals from the model are useful.\n3. Therefore, signals from the model are valuable. [1,2]"},
		{"name", "1. If X arrived, then X was seen.\n2. Rex arrived.\n3. Therefore, Rex was seen. [1,2]", "1. If Rex arrived, then Rex was seen.\n2. Rex arrived.\n3. Therefore, Rex was seen. [1,2]"},
		{"plural class", "1. All dogs are animals.\n2. The dog is a dog.\n3. Therefore, the dog is an animal. [1,2]", "1. Every dog is an animal.\n2. The dog is a dog.\n3. Therefore, the dog is an animal. [1,2]"},
		{"gerund clause", "1. Swimming is healthy and the lamp is lit.\n2. Therefore, swimming is healthy. [1]", "1. The activity is healthy and the lamp is lit.\n2. Therefore, the activity is healthy. [1]"},
		// v46.0 accepts several allowlisted adjectives ("the muddy wooden cart"); noun modifiers stay refused.
		{"description", "1. If X arrived, then X was seen.\n2. The brick garden wall arrived.\n3. Therefore, the brick garden wall was seen. [1,2]", "1. If the brick garden wall arrived, then the brick garden wall was seen.\n2. The brick garden wall arrived.\n3. Therefore, the brick garden wall was seen. [1,2]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, result := CheckEntailmentGate(tc.refused); result.Passed {
				t.Fatal("documented refusal certified")
			}
			if _, result := CheckEntailmentGate(tc.repaired); !result.Passed {
				t.Fatal(result.Message)
			}
		})
	}
}
