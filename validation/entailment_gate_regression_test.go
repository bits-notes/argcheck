package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Assert every expected ok: step, even when an earlier step has regressed.
// Do not select cases using the current gate verdict: that would silently
// exclude precisely the regressions this parity check is meant to detect.
func assertAcceptedStepParity(t *testing.T, argument NumberedArgument, forms []string) {
	t.Helper()
	report := GateNumberedArgument(argument)
	if !report.Passed {
		t.Errorf("expected VALID: %+v", report)
	}
	if len(report.Steps) != len(forms) {
		t.Fatalf("got %d derived steps, want %d", len(report.Steps), len(forms))
	}
	for i, step := range report.Steps {
		if !step.Passed {
			t.Errorf("step %d lost ok: %s", step.Step, step.Mismatch)
		}
		if forms[i] != "" && step.Form != forms[i] {
			t.Errorf("step %d form %q, want %q", step.Step, step.Form, forms[i])
		}
	}
}

func TestAcceptedFixtureStepParity(t *testing.T) {
	// v44 by-design refusals: retained separately from the accepted controls.
	refused := map[string]strictRefusal{
		// term.inflection_verbatim: class spellings differ.
		"TestAcceptedFixtureStepParity/battery.json/L-TV-barbara": {"term.inflection_verbatim", "singular 'every N'"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestAcceptedFixtureStepParity/GB1.txt": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestAcceptedFixtureStepParity/earman-norton-1987_A11.txt": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestAcceptedFixtureStepParity/earman-norton-1987_A5.txt": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestAcceptedFixtureStepParity/gettier-1963_A2.txt": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestAcceptedFixtureStepParity/goldman-1967_A7.txt": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected clause has no admitted opener.
		"TestAcceptedFixtureStepParity/hempel-oppenheim-1948_A23.txt": {"scope.clause_start", "move the operator into each clause"},
		// v46.0: Ada is evidenced; the embedded schematic clause still fails scope.clause_start.
		"TestAcceptedFixtureStepParity/every_parcel_in_a_crate": {"scope.clause_start", "move the operator into each clause"},
		// v46.0: Ada is evidenced; the embedded schematic clause still fails scope.clause_start.
		"TestAcceptedFixtureStepParity/a_parcel_with_no_label": {"scope.clause_start", "move the operator into each clause"},
	}

	// The 999 frozen accepted ASTs and their per-step forms use the same
	// assertion in TestEntailmentGateV16AcceptedCompatibility.
	check := func(t *testing.T, text string, pinned map[int]string) {
		t.Helper()
		if listedStrictRefusal(t, text, refused) {
			return
		}
		argument, result := ParseNumberedArgument(text, numberedArgumentOptions{})
		if !result.Passed {
			t.Fatal(result.Message)
		}
		var forms []string
		for _, step := range argument.Steps {
			if step.Kind == numberedArgumentStepConclusion {
				forms = append(forms, pinned[step.Number])
			}
		}
		assertAcceptedStepParity(t, argument, forms)
	}
	for _, path := range []string{
		"testdata/entailment_gate/battery.json",
		"testdata/entailment_gate/atomic-verification-cases.json",
	} {
		for _, c := range loadGoldGateCases(t, path) {
			if c.GoldStatus != "transparently_valid" || c.CatalogGap != "" || c.StructuralDefect != "" {
				continue
			}
			t.Run(filepath.Base(path)+"/"+c.ID, func(t *testing.T) {
				check(t, c.text(), nil)
			})
		}
	}
	fixtures, err := filepath.Glob("testdata/regression/*.txt")
	if err != nil || len(fixtures) < 3 {
		t.Fatalf("regression fixtures: %v (%d files)", err, len(fixtures))
	}
	pins := map[string][]int{
		"earman-norton-1987_A5.txt":  {7},
		"goldman-1967_A7.txt":        {7},
		"earman-norton-1987_A11.txt": {9, 11},
	}
	for _, path := range fixtures {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if listedStrictRefusal(t, string(data), refused) {
				return
			}
			if filepath.Base(path) == "earman-norton-1987_A5.txt" {
				// The fixture's premise has "theories, every", while its
				// schematic antecedent has "X every": that comma is lexical.
				report, _ := CheckEntailmentGate(string(data))
				if report.Passed || len(report.Steps) != 3 || !report.Steps[0].Passed || report.Steps[1].Passed || !report.Steps[2].Passed {
					t.Fatalf("expected only step 9 to lose its comma-mismatched certificate: %+v", report)
				}
				return
			}
			if filepath.Base(path) == "earman-norton-1987_A11.txt" {
				// v37.0: steps 6 and 9 rest on a premise that opens with the
				// prepositional frame "Under Leibniz equivalence", which may
				// cover the whole sentence, so they are not certified.
				report, _ := CheckEntailmentGate(string(data))
				for _, step := range report.Steps {
					framed := step.Step == 6 || step.Step == 9
					if framed && (step.Passed || !strings.Contains(step.Mismatch, "may cover the whole sentence")) {
						t.Fatalf("step %d: want the opener refusal, got %+v", step.Step, step)
					}
					if !framed && !step.Passed {
						t.Fatalf("step %d lost its certificate: %s", step.Step, step.Mismatch)
					}
				}
				return
			}
			pinned := map[int]string{}
			for _, n := range pins[filepath.Base(path)] {
				pinned[n] = "schematic conjunctive modus ponens"
			}
			check(t, string(data), pinned)
		})
	}
	// Minimal object-noun-phrase cases exercise both an unchanged conjunct
	// and a clause containing the variable. Both must retain their schematic
	// certificate when the n-place parser finds mixed quantifiers.
	for _, object := range []string{
		"every parcel in a crate",
		"a parcel with no label",
	} {
		text := "1. Ada is a courier.\n2. Ada carries " + object + ".\n" +
			"3. If X is a courier and X carries " + object + ", then X is ready.\n" +
			"4. Therefore, Ada is ready. [1,2,3]"
		t.Run(object, func(t *testing.T) {
			check(t, text, map[int]string{4: "schematic conjunctive modus ponens"})
			// A different substitution or object cannot inherit that certificate.
			for _, invalid := range []string{
				strings.Replace(text, "2. Ada carries", "2. Ben carries", 1),
				strings.Replace(text, "4. Therefore, Ada", "4. Therefore, Ben", 1),
				strings.Replace(text, "2. Ada carries "+object, "2. Ada carries another parcel", 1),
			} {
				report, _ := CheckEntailmentGate(invalid)
				if report.Passed {
					t.Errorf("accepted inconsistent schematic substitution: %s", invalid)
				}
			}
		})
	}
}
