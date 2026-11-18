package validation

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// gateConclusionStepLegacy is the unchanged v25 entry point. Compare entire
// reports, including the selected form, not just the final validity verdict.
func v26AssertLegacyParity(t *testing.T, text string, accepted bool) {
	t.Helper()
	arg, syntax := ParseNumberedArgument(text, numberedArgumentOptions{})
	if !syntax.Passed {
		t.Fatal(syntax.Message)
	}
	steps := map[int]NumberedArgumentStep{}
	for _, s := range arg.Steps {
		steps[s.Number] = s
	}
	for _, s := range arg.Steps {
		if s.Kind != numberedArgumentStepConclusion {
			continue
		}
		want := gateConclusionStepLegacy(s, steps)
		got := gateConclusionStep(s, steps)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("step %d: got %+v; v25 %+v", s.Number, got, want)
		}
	}
	if accepted {
		_, report := CheckEntailmentGate(text)
		if !report.Passed {
			t.Fatal(report.Message)
		}
	}
}

func TestCatalogV26NonNumericParity(t *testing.T) {
	data, err := os.ReadFile("testdata/counting/non-numeric-phrases.txt")
	if err != nil {
		t.Fatal(err)
	}
	phrases := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	phrases = append(phrases, "at least as", "exactly the", "exactly what", "exactly that", "exactly this", "at least a", "at least in", "at least some", "at least such", "at most p", "at most alpha")
	for i, phrase := range phrases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if countingActive(phrase) {
				t.Fatalf("non-numeric counting activation: %q", phrase)
			}
			// Neutral subject makes even the truncated corpus snippets sentences.
			sentence := "The record says " + strings.TrimSpace(phrase)
			// A truncated corpus quotation contains a dangling closing
			// parenthesis. Grouping syntax now rejects it before parity checks.
			if i == 1 {
				_, syntax := ParseNumberedArgument(v26Argument([]string{"q"}, sentence), numberedArgumentOptions{})
				if syntax.Passed || !strings.Contains(syntax.Message, "unbalanced parentheses: unexpected closing parenthesis") {
					t.Fatalf("expected unmatched-parenthesis syntax failure: %s", syntax.Message)
				}
				return
			}
			v26AssertLegacyParity(t, v26Argument([]string{"If p then " + sentence, "p"}, sentence), false)
			// Also preserve legacy refusals and their diagnostics.
			v26AssertLegacyParity(t, v26Argument([]string{"q"}, sentence), false)
		})
	}
}

func TestCatalogV26RealArgumentParity(t *testing.T) {
	for _, name := range []string{"GB1.txt", "gettier-1963_A2.txt", "hempel-oppenheim-1948_A23.txt"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/regression/" + name)
			if err != nil {
				t.Fatal(err)
			}
			v26AssertLegacyParity(t, string(data), true)
		})
	}
}

func TestCatalogV26LegacyBeforeBound(t *testing.T) {
	for _, phrase := range []string{"At least one lab is red", "Exactly six labs are red", "At most 100 labs are red"} {
		v26AssertLegacyParity(t, v26Argument([]string{phrase}, phrase), true)
		v26AssertLegacyParity(t, v26Argument([]string{"If p then " + phrase, "p"}, phrase), true)
	}
}

func TestCatalogV26B1English(t *testing.T) {
	data, err := os.ReadFile("testdata/counting/b1-with-person.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, report := CheckEntailmentGate(string(data))
	if !report.Passed {
		t.Fatal(report.Message)
	}
	ps := []string{"Exactly one person wrote Waverley", "Scott wrote Waverley", "Scott is a person"}
	c := "No person other than Scott wrote Waverley"
	v26Check(t, ps, c, true, 4)
	v26Check(t, ps[:2], c, false, 2)
	v26Check(t, []string{ps[0], ps[1], "Ada is a person"}, c, false, 2)
	_, report = CheckEntailmentGate(v26Argument(append(ps, "Ada is red"), c))
	if report.Passed {
		t.Fatal("unrelated extra citation accepted")
	}
}
