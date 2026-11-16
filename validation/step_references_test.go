package validation

import (
	"encoding/json"
	"strings"
	"testing"
)

// Bracketed step references ([n], [n,m]) are allowed only as the whole
// antecedent of a bridge premise: "If [1], then C.", "If [1] and [2], then C."
// or "If [1,2], then C.". Each reference expands to "it is the case that <step
// text>". Every case cites its steps correctly, so a failure comes from the
// reference itself. Parenthesized step numbers are plain text; a failing step
// that uses one gets a hint (message only, no verdict change).
func TestStepReferencePositions(t *testing.T) {
	tests := []struct {
		name     string
		argument string
		valid    bool
		message  string
	}{
		{"antecedent, single reference", "1. The street is wet.\n2. If [1], then it rained.\n3. Therefore, it rained. [1,2]", true, ""},
		{"antecedent, references joined by and", "1. The street is wet.\n2. The sky is grey.\n3. If [1] and [2], then it rained.\n4. Therefore, it rained. [1,2,3]", true, ""},
		{"antecedent, list reference", "1. The street is wet.\n2. The sky is grey.\n3. If [1,2], then it rained.\n4. Therefore, it rained. [1,2,3]", true, ""},
		{"antecedent, no comma before then", "1. The street is wet.\n2. If [1] then it rained.\n3. Therefore, it rained. [1,2]", true, ""},
		{"reference to an intermediate conclusion", "1. The street is wet.\n2. If the street is wet, then the road is slick.\n3. Therefore, the road is slick. [1,2]\n4. If [3], then drivers slow down.\n5. Therefore, drivers slow down. [3,4]", true, ""},
		{"reference to an earlier derived step", "1. The street is wet.\n2. If [1], then it rained.\n3. Therefore, it rained. [1,2]\n4. If [2], then the forecast was right.\n5. Therefore, the forecast was right. [2,4]", true, ""},
		{"reference to a discharged conditional", "1. If the street is wet, then it rained.\n| 2. Suppose the street is wet.\n| 3. Therefore, it rained. [1,2]\n4. Therefore, if the street is wet, then it rained. [2-3]\n5. If [4], then the inference holds.\n6. Therefore, the inference holds. [4,5]", true, ""},
		{"comma before and between references", "1. The street is wet.\n2. The sky is grey.\n3. If [1], and [2], then it rained.\n4. Therefore, it rained. [1,2,3]", false, "outside an explicit bridge premise"},
		{"reference in the consequent", "1. The street is wet.\n2. If it rained, then [1].\n3. It rained.\n4. Therefore, the street is wet. [2,3]", false, "outside an explicit bridge premise"},
		{"reference under negation", "1. The street is wet.\n2. It is not the case that [1].\n3. Therefore, the street is wet. [1]", false, "outside an explicit bridge premise"},
		{"reference in a disjunction", "1. The street is wet.\n2. The sky is grey.\n3. Either [1] or [2].\n4. Therefore, the street is wet. [1]", false, "outside an explicit bridge premise"},
		{"reference mixed with text in the antecedent", "1. The street is wet.\n2. If [1] and the sky is grey, then it rained.\n3. The sky is grey.\n4. Therefore, it rained. [1,3,2]", false, "outside an explicit bridge premise"},
		{"reference to a later step", "1. If [2], then it rained.\n2. The street is wet.\n3. Therefore, it rained. [1,2]", false, "does not refer to a prior step"},
		{"referenced step not otherwise used", "1. The street is wet.\n2. If [1], then it rained.\n3. Therefore, it rained. [2]", false, "antecedent \"it is the case that the street is wet\" is not matched verbatim by any cited premise"},
		{"reference into a closed subproof", "1. If the street is wet, then it rained.\n| 2. Suppose the street is wet.\n| 3. Therefore, it rained. [1,2]\n4. Therefore, if the street is wet, then it rained. [2-3]\n5. If [3], then the game is off.\n6. Therefore, the game is off. [3,5]", false, "inside the closed subproof"},
		{"bridge premise inside a subproof", "1. If the street is wet, then it rained.\n| 2. Suppose the street is wet.\n| 3. If [2], then the ground is damp.\n| 4. Therefore, the ground is damp. [2,3]\n5. Therefore, if the street is wet, then the ground is damp. [2-4]", false, "must be at depth 0"},
		{"parenthesized (1) gets a hint", "1. The street is wet.\n2. If (1), then it rained.\n3. Therefore, it rained. [1,2]", false, "to refer to step 1, write [1]"},
		{"parenthesized (1,2) gets a hint", "1. The street is wet.\n2. The sky is grey.\n3. If (1,2), then it rained.\n4. Therefore, it rained. [1,2,3]", false, "to refer to steps 1 and 2, write [1,2]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if report.Passed != tc.valid || result.Passed != tc.valid {
				t.Fatalf("valid=%v/%v, want %v: %s\n%s", report.Passed, result.Passed, tc.valid, result.Message, tc.argument)
			}
			if tc.message == "" {
				return
			}
			blob, _ := json.Marshal(report.Steps)
			if !strings.Contains(result.Message+string(blob), tc.message) {
				t.Fatalf("message lacks %q: %s %s", tc.message, result.Message, blob)
			}
		})
	}
}

func TestStepReferenceHintOnlyForEarlierSteps(t *testing.T) {
	for _, tc := range []struct {
		statement string
		step      int
		want      bool
	}{
		{"If (1), then it rained", 2, true},
		{"If (1, 2), then it rained", 3, true},
		{"If (3), then it rained", 2, false},
		{"The ratio is (0)", 2, false},
		{"Theorem (12) holds", 4, false},
		{"See note (a)", 3, false},
	} {
		if got := stepReferenceHint(tc.statement, tc.step) != ""; got != tc.want {
			t.Fatalf("hint for %q at step %d = %v, want %v", tc.statement, tc.step, got, tc.want)
		}
	}
}
