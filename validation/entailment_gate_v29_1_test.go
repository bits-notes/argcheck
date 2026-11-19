package validation

import "testing"

func TestEntailmentGateV291NonCopularAgreement(t *testing.T) {
	refused := map[string]strictRefusal{
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV291NonCopularAgreement/have_to_has": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV291NonCopularAgreement/regular_-s": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV291NonCopularAgreement/consonant_y": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestEntailmentGateV291NonCopularAgreement/verbatim_plural_still_accepted": {"term.inflection_verbatim", "singular 'every N'"},
	}

	tests := []struct {
		name, argument string
		want           bool
	}{
		{"have to has", "1. All squares have four sides.\n2. Rex is a square.\n3. Therefore, Rex has four sides. [1,2]", true},
		{"regular -s", "1. All dogs bark loudly.\n2. Rex is a dog.\n3. Therefore, Rex barks loudly. [1,2]", true},
		{"consonant y", "1. All birds fly south.\n2. Tweety is a bird.\n3. Therefore, Tweety flies south. [1,2]", true},
		{"verbatim plural still accepted", "1. All squares have four sides.\n2. Rex is a square.\n3. Therefore, rex have four sides. [1,2]", true},
		{"tense change rejected", "1. All birds fly south.\n2. Tweety is a bird.\n3. Therefore, Tweety flew south. [1,2]", false},
		{"plural instance keeps plural", "1. All squares have four sides.\n2. Rex and Max are squares.\n3. Therefore, Rex and Max has four sides. [1,2]", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if listedStrictRefusal(t, tc.argument, refused) {
				return
			}
			if report, _ := CheckEntailmentGate(tc.argument); report.Passed != tc.want {
				t.Fatalf("accepted=%v, want %v\n%s", report.Passed, tc.want, tc.argument)
			}
		})
	}
}
