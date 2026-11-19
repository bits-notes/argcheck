package validation

import (
	"strings"
	"testing"
)

func TestRound2ScopeProbes(t *testing.T) {
	cases := []struct{ file, hint, control string }{
		{"group-more-4", mixedScopeHint, "1. Both Ann is happy and either Bob is happy or Cal reads.\n2. Therefore, Ann is happy. [1]"},
		{"lowercase-actual-1", mixedScopeHint, "1. Both Ann sings and Bob dances.\n2. Therefore, Ann sings. [1]"},
		{"frame-omitted-2", clauseStartHint, "1. It is not the case that the lamp is lit and the door is open.\n2. Therefore, either it is not the case that the lamp is lit or it is not the case that the door is open. [1]"},
		{"gerund-frame-1", ingFrameHint, "1. Both Alice is happy and Bob is sad.\n2. Therefore, Bob is sad. [1]"},
		{"homonym-actual-1", copulaHint, "1. Either the men are quiet or the men are fish.\n2. Therefore, either the men are quiet or the men are fish. [1]"},
		{"negation-1", descriptionNegationHint, "1. It is not the case that Louis is bald.\n2. Therefore, Louis is not bald. [1]"},
		{"ly-exception-2", lyExceptionHint, "1. The lamp is lit and the door is open.\n2. Therefore, the door is open and the lamp is lit. [1]"},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			b := scopeAuditProbes[tc.file]
			r, result := CheckEntailmentGate(string(b))
			if r.Passed || !strings.Contains(result.Message, tc.hint) {
				t.Fatalf("probe: %+v / %s", r, result.Message)
			}
			r, result = CheckEntailmentGate(tc.control)
			if !r.Passed {
				t.Fatalf("control: %+v / %s", r, result.Message)
			}
		})
	}
}

func TestRound2ScopeNestedAndReverse(t *testing.T) {
	for _, input := range []string{
		"1. The present king of France is not bald.\n2. Therefore, it is not the case that the present king of France is bald. [1]",
		"1. It is not the case that daily the lamp is lit and the door is open.\n2. Therefore, either it is not the case that daily the lamp is lit or it is not the case that the door is open. [1]",
		"1. Following the script Alice is happy and Bob is sad.\n2. Therefore, Bob is sad and following the script Alice is happy. [1]",
	} {
		r, _ := CheckEntailmentGate(input)
		if r.Passed {
			t.Errorf("unsafe conversion passed: %s", input)
		}
	}
	for _, input := range []string{
		"1. Either the men are happy or red.\n2. Therefore, either the men are happy or the men are red. [1]",
		"1. It is not the case that perhaps the lamp is lit and the door is open.\n2. Therefore, it is not the case that perhaps the lamp is lit and the door is open. [1]",
	} {
		r, result := CheckEntailmentGate(input)
		if !r.Passed {
			t.Errorf("control failed: %s: %s", input, result.Message)
		}
	}
	// Check both switches independently at the recursive expression boundary;
	// Daily's audit refusal intentionally has overlapping strict causes.
	if allStrictRules().without("opener.ly_exception").scopeExpressionFailure("Daily the lamp is lit and the door is open") == "" {
		t.Error("clause-start guard missed Daily")
	}
	if allStrictRules().without("scope.clause_start").scopeExpressionFailure("Daily the lamp is lit and the door is open") == "" {
		t.Error("ly guard missed Daily")
	}
}

// Verbatim audit inputs, embedded so the tests do not depend on workspace files.
var scopeAuditProbes = map[string]string{
	"group-more-4":       "1. Ann is a singer who dances and Bob is happy or Cal reads.\n2. Therefore, Ann is a singer who dances. [1]\n",
	"lowercase-actual-1": "1. Ann sings and Bob dances or cal reads.\n2. Therefore, Ann sings. [1]\n",
	"frame-omitted-2":    "1. It is not the case that perhaps the lamp is lit and the door is open.\n2. Therefore, either it is not the case that perhaps the lamp is lit or it is not the case that the door is open. [1]\n",
	"gerund-frame-1":     "1. Following the script Alice is happy and Bob is sad.\n2. Therefore, Bob is sad. [1]\n",
	"homonym-actual-1":   "1. Either the men are quiet or fish.\n2. Therefore, either the men are quiet or the men are fish. [1]\n",
	"negation-1":         "1. It is not the case that the present king of France is bald.\n2. Therefore, the present king of France is not bald. [1]\n",
	"ly-exception-2":     "1. Daily the lamp is lit and the door is open.\n2. Therefore, the door is open and daily the lamp is lit. [1]\n",
}

func TestRound2ScopeClosedClassIngStarts(t *testing.T) {
	for _, word := range []string{"something", "anything", "everything", "nothing"} {
		t.Run(word, func(t *testing.T) {
			if why := allStrictRules().scopeOperandFailure(word + " is red"); why != "" {
				t.Fatalf("closed-class clause start refused: %s", why)
			}
		})
	}
}

// A parenthesised disjunct followed by another disjunct is a compound operand
// that is checked part by part; expletive "there" before be or a modal is a
// clause start. Operators inside a part stay refused.
func TestRound2ScopeCompoundOperands(t *testing.T) {
	for _, text := range []string{
		"1. The lamp is lit.\n2. Therefore, the lamp is lit or (the door is open) or the bell rings. [1]",
		"1. The lamp is lit.\n2. Therefore, either (the lamp is lit) or (the door is open) or (the bell rings). [1]",
		"1. There must be a dog.\n2. Therefore, either there must be a dog or the door is open. [1]",
		"1. Birnbaum's proof is valid and the door is open.\n2. Therefore, Birnbaum's proof is valid. [1]",
	} {
		if _, r := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("should certify: %s\n%s", r.Message, text)
		}
	}
	for _, text := range []string{
		"1. The lamp is lit.\n2. Therefore, the lamp is lit or (perhaps the door is open) or the bell rings. [1]",
		"1. There the lamp is lit and the door is open.\n2. Therefore, the door is open. [1]",
		"1. Here's the lamp and the door is open.\n2. Therefore, the door is open. [1]",
	} {
		if _, r := CheckEntailmentGate(text); r.Passed {
			t.Errorf("certified: %s", text)
		}
	}
}

// A conditional's antecedent and consequent are checked as separate
// expressions; "If" does not count as the first clause's start.
func TestRound2ScopeConditionalParts(t *testing.T) {
	for _, text := range []string{
		"1. If the lamp is lit and the door is open, then the room is bright.\n2. The lamp is lit.\n3. The door is open.\n4. Therefore, the room is bright. [1,2,3]",
		"1. The lamp is lit and the door is open.\n2. If the lamp is lit and the door is open, then the room is bright.\n3. Therefore, the room is bright. [1,2]",
		"1. If Alice runs and Bob jumps, then Carol cheers.\n2. Alice runs.\n3. Bob jumps.\n4. Therefore, Carol cheers. [1,2,3]",
	} {
		if _, r := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("should certify: %s\n%s", r.Message, text)
		}
	}
	for _, text := range []string{
		"1. If perhaps the lamp is lit and the door is open, then the room is bright.\n2. Perhaps the lamp is lit.\n3. The door is open.\n4. Therefore, the room is bright. [1,2,3]",
		"1. If the lamp is lit and the door is open or the bell rings, then the room is bright.\n2. The lamp is lit.\n3. The door is open.\n4. Therefore, the room is bright. [1,2,3]",
	} {
		if _, r := CheckEntailmentGate(text); r.Passed {
			t.Errorf("certified: %s", text)
		}
	}
}
