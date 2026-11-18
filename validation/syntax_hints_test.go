package validation

import "testing"

func TestSyntaxHints(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"lowercase therefore accepted", "1. If it rains, then it pours.\n2. It rains.\n3. therefore, it pours. [1,2]", ""},
		{"premise without period", "1. If it rains, then it pours.\n2. It rains\n3. Therefore, it pours. [1,2]", "line 2 is not a valid premise or conclusion: end the sentence with a period"},
		{"no step number", "If it rains, then it pours.\n2. It rains.\n3. Therefore, it pours. [1,2]", `line 1 is not a valid premise or conclusion: start the line with its step number, a period and a space, e.g. "1. "`},
		{"citations without period", "1. It rains.\n2. It rains [1]", `line 2 is not a valid premise or conclusion: end the sentence with a period before the citations, e.g. "... wet. [1,2]"`},
		{"marker without citations", "1. It rains.\n2. Therefore, it rains.", `line 2 starts with Therefore, but does not have valid terminal citations: end the line with a period, a space and the cited step numbers in square brackets, e.g. "... wet. [1,2]"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, result := ParseNumberedArgument(tc.input, numberedArgumentOptions{})
			if tc.want == "" {
				if !result.Passed {
					t.Fatalf("unexpected failure: %s", result.Message)
				}
				return
			}
			if result.Passed || result.Message != tc.want {
				t.Fatalf("got %v %q, want %q", result.Passed, result.Message, tc.want)
			}
		})
	}
}
