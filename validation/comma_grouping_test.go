package validation

import (
	"strings"
	"testing"
)

func TestCommaGroupingRefusalsAndRepairs(t *testing.T) {
	for _, tc := range []struct{ name, refused, accepted string }{
		{"C2-1", "1. Ann sings, and Bob dances or Cal reads.\n2. Therefore, Ann sings. [1]", "1. Both Ann sings and either Bob dances or Cal reads.\n2. Therefore, Ann sings. [1]"},
		{"C2-2", "1. Ann sings or Bob dances, and Cal reads.\n2. Therefore, Cal reads. [1]", "1. (Ann sings or Bob dances) and Cal reads.\n2. Therefore, Cal reads. [1]"},
		{"C2-3", "1. The lamp is lit, and the door is open or the alarm sounds.\n2. Therefore, The lamp is lit. [1]", "1. Both the lamp is lit and either the door is open or the alarm sounds.\n2. Therefore, The lamp is lit. [1]"},
		{"C2-4", "1. The lamp is lit or the door is open, and the alarm sounds.\n2. Therefore, The alarm sounds. [1]", "1. (The lamp is lit or the door is open) and the alarm sounds.\n2. Therefore, The alarm sounds. [1]"},
		{"C4-4", "1. Rex is happy, and Rex is rich or Rex is poor.\n2. Therefore, Rex is happy. [1]", "1. Both Rex is happy and either Rex is rich or Rex is poor.\n2. Therefore, Rex is happy. [1]"},
		{"H5-5", "1. Ann sings, and Bob dances or Cal reads or Dan sleeps.\n2. Therefore, Ann sings. [1]", "1. Ann sings and (Bob dances or Cal reads or Dan sleeps).\n2. Therefore, Ann sings. [1]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, result := CheckEntailmentGate(tc.refused)
			if result.Passed || !strings.Contains(result.Message, "mixes top-level and/or without explicit grouping; use either...or / both...and, or parentheses to group") {
				t.Fatalf("expected grouping syntax refusal, got %+v", result)
			}
			_, result = CheckEntailmentGate(tc.accepted)
			if !result.Passed {
				t.Fatal(result.Message)
			}
		})
	}
}

// A comma cannot override the explicit both scaffold.
func TestCommaGroupingScaffoldPrecedence(t *testing.T) {
	_, result := CheckEntailmentGate("1. Both Ann sings and either Bob dances, or Cal reads.\n2. Therefore, Ann sings. [1]")
	if !result.Passed {
		t.Fatal(result.Message)
	}
}

func TestCommaGroupingLeadingOperator(t *testing.T) {
	_, result := CheckEntailmentGate("1. Perhaps Ann sings, and Bob dances.\n2. Therefore, Bob dances. [1]")
	if result.Passed || !strings.Contains(result.Message, operatorGroupingMessage) {
		t.Fatalf("expected operator scope refusal, got %+v", result)
	}
}
