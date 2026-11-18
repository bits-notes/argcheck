package validation

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestConclusionMarkersModusPonens(t *testing.T) {
	for _, marker := range []string{"Therefore, ", "So, ", "Thus, ", "Hence, ", "Therefore ", "So ", "Thus ", "Hence ", "", "therefore, ", "so, ", "thus, ", "hence, ", "sO "} {
		t.Run(marker, func(t *testing.T) {
			input := "1. If it rains, then the ground is wet.\n2. It rains.\n3. " + marker + "the ground is wet. [1,2]"
			arg, result := ParseNumberedArgument(input, numberedArgumentOptions{})
			if !result.Passed {
				t.Fatal(result.Message)
			}
			step := arg.Steps[2]
			if step.Kind != numberedArgumentStepConclusion || step.Statement != "the ground is wet" || !reflect.DeepEqual(step.Citations, []int{1, 2}) {
				t.Fatalf("unexpected step: %+v", step)
			}
			if valid, _, msg := checkRulesDocArgument(input); !valid {
				t.Fatal(msg)
			}
			if got := numberedArgumentStepLine(step); got != "3. Therefore, the ground is wet. [1,2]" {
				t.Fatalf("rendered %q", got)
			}
		})
	}
}

func TestConclusionMarkersInlineReferencesStayPremises(t *testing.T) {
	for _, statement := range []string{"If [1,2], then P", "P if [1]", "If P, then [1]"} {
		t.Run(statement, func(t *testing.T) {
			step, result := parseNumberedArgumentLine("3. "+statement+".", 3)
			if !result.Passed {
				t.Fatal(result.Message)
			}
			if step.Kind != numberedArgumentStepPremise || step.Statement != statement || len(step.Citations) != 0 {
				t.Fatalf("unexpected step: %+v", step)
			}
		})
	}
}

func TestConclusionMarkersRequireTerminalCitations(t *testing.T) {
	for _, marker := range []string{"Therefore", "So", "Thus", "Hence", "so", "THUS"} {
		for _, separator := range []string{", ", " "} {
			for _, ending := range []string{".", ". [1, 2]", ". []", ". [1,2] trailing", ". [1,2]."} {
				input := "1. If it rains, then the ground is wet.\n2. It rains.\n3. " + marker + separator + "the ground is wet" + ending
				_, result := ParseNumberedArgument(input, numberedArgumentOptions{})
				want := fmt.Sprintf("line 3 starts with %s, but does not have valid terminal citations", marker)
				if result.Passed || !strings.HasPrefix(result.Message, want+": ") {
					t.Errorf("%q: got %+v, want %q", input, result, want)
				}
			}
		}
	}
}

func TestConclusionMarkersOrdinaryPremises(t *testing.T) {
	for _, statement := range []string{"Sophie is tall", "It rains", "So many lamps are lit", "Thusly is a word", "Henceforth it rains", "The label says therefore, then stops"} {
		step, result := parseNumberedArgumentLine("1. "+statement+".", 1)
		if !result.Passed || step.Kind != numberedArgumentStepPremise || step.Statement != statement {
			t.Errorf("%q: step=%+v result=%+v", statement, step, result)
		}
	}
}

func TestConclusionMarkersScopedRanges(t *testing.T) {
	for _, marker := range []string{"So, ", "thus ", ""} {
		input := "1. If it rains, then the ground is wet.\n| 2. Suppose it rains.\n| 3. " + marker + "the ground is wet. [1,2]\n4. " + marker + "if it rains, then the ground is wet. [2-3]"
		if valid, _, msg := checkRulesDocArgument(input); !valid {
			t.Fatalf("%q: %s", marker, msg)
		}
		arg, result := ParseNumberedArgument(input, numberedArgumentOptions{})
		if !result.Passed {
			t.Fatal(result.Message)
		}
		for _, step := range arg.Steps {
			reparsed, result := parseNumberedArgumentLine(numberedArgumentStepLine(step), step.Number)
			if !result.Passed || reparsed.Statement != step.Statement || reparsed.Depth != step.Depth || !reflect.DeepEqual(reparsed.Ranges, step.Ranges) {
				t.Fatalf("round trip: %+v %+v", reparsed, result)
			}
		}
	}
}
