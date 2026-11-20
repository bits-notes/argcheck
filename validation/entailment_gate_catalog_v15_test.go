package validation

import (
	"fmt"
	"strings"
	"testing"
)

func TestEntailmentGateCatalogV15(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_forward/commas0": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_forward/commas1": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_forward/commas2": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_forward/commas3": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_reverse/commas0": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_reverse/commas1": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_reverse/commas2": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_reverse/commas3": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_inner_scaffold/commas0": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_inner_scaffold/commas1": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_inner_scaffold/commas2": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestEntailmentGateCatalogV15/conjunction_inner_scaffold/commas3": {"scope.clause_start", "move the operator into each clause"},
	}

	for _, tc := range []struct{ name, premise, conclusion, form string }{
		{"reflexivity", "Socrates is a man", "Socrates is identical to Socrates", "identity reflexivity"},
		{"case insensitive occurrence", "SOCRATES is a man", "Socrates is identical to Socrates", "identity reflexivity"},
		{"multiword term", "Mary Jane sings", "Mary Jane is identical to Mary Jane", "identity reflexivity"},
		{"punctuation boundary", "Ada admires (Socrates)", "Socrates is identical to Socrates", "identity reflexivity"},
		{"later whole word", "Socratesian admires Socrates", "Socrates is identical to Socrates", "identity reflexivity"},
		{"duality necessity forward", "It is not the case that possibly, it is not the case that the rain falls", "Necessarily, the rain falls", "modal duality"},
		{"duality necessity reverse", "Necessarily, the rain falls", "It is not the case that possibly, it is not the case that the rain falls", "modal duality"},
		{"duality possibility forward", "It is not the case that necessarily, it is not the case that the rain falls", "Possibly, the rain falls", "modal duality"},
		{"duality possibility reverse", "Possibly, the rain falls", "It is not the case that necessarily, it is not the case that the rain falls", "modal duality"},
		{"K", "Necessarily, if the rain falls, then the snow falls", "If necessarily, the rain falls, then necessarily, the snow falls", "modal distribution"},
		{"possibility monotonicity", "Necessarily, if the rain falls, then the snow falls", "If possibly, the rain falls, then possibly, the snow falls", "modal distribution"},
		{"conjunction forward", "Necessarily, the rain falls and the snow falls", "It is the case both that necessarily, the rain falls, and that necessarily, the snow falls", "modal distribution"},
		{"conjunction reverse", "It is the case both that necessarily, the rain falls, and that necessarily, the snow falls", "Necessarily, the rain falls and the snow falls", "modal distribution"},
		{"conjunction inner scaffold", "Necessarily, it is the case both that the rain falls, and that the snow falls", "It is the case both that necessarily, the rain falls, and that necessarily, the snow falls", "modal distribution"},
		{"earlier self entailment", "Socrates is identical to Socrates", "Socrates is identical to Socrates", "self-entailment"},
	} {
		// Exercise every combination of optional commas on premise/conclusion.
		for variant := 0; variant < 4; variant++ {
			t.Run(fmt.Sprintf("%s/commas%d", tc.name, variant), func(t *testing.T) {
				p, c := tc.premise, tc.conclusion
				strip := strings.NewReplacer("Necessarily, ", "Necessarily ", "necessarily, ", "necessarily ", "Possibly, ", "Possibly ", "possibly, ", "possibly ")
				if variant&1 != 0 {
					p = strip.Replace(p)
				}
				if variant&2 != 0 {
					c = strip.Replace(c)
				}
				argument := fmt.Sprintf("1. %s.\n2. Therefore, %s. [1]", p, c)
				if listedStrictRefusal(t, argument, refused) {
					return
				}
				report, result := CheckEntailmentGate(argument)
				if !report.Passed || !result.Passed || len(report.Steps) != 1 || report.Steps[0].Form != tc.form {
					t.Fatalf("want %q: %s %+v", tc.form, result.Message, report)
				}
			})
		}
	}
}

func TestEntailmentGateV15NearMisses(t *testing.T) {
	for _, tc := range []struct{ name, premise, conclusion string }{
		{"different identity terms", "Socrates is a man", "Socrates is identical to Plato"},
		{"absent identity term", "Plato is a man", "Socrates is identical to Socrates"},
		{"substring suffix", "Socratesian sings", "Socrates is identical to Socrates"},
		{"substring prefix", "antisocrates sings", "Socrates is identical to Socrates"},
		{"unicode boundary", "éSocrates sings", "Socrates is identical to Socrates"},
		{"word underscore", "Socrates_jr sings", "Socrates is identical to Socrates"},
		{"morphology is not identity", "The cat sings", "The cat is identical to The cats"},
		{"compound pseudo term", "The rain falls and the snow falls", "The rain falls and the snow falls is identical to the rain falls and the snow falls"},
		{"duality missing inner not necessity", "It is not the case that possibly, the rain falls", "Necessarily, the rain falls"},
		{"duality missing inner not possibility", "It is not the case that necessarily, the rain falls", "Possibly, the rain falls"},
		{"duality reverse missing not necessity", "Necessarily, the rain falls", "It is not the case that possibly, the rain falls"},
		{"duality reverse missing not possibility", "Possibly, the rain falls", "It is not the case that necessarily, the rain falls"},
		{"duality wrong clause", "It is not the case that possibly, it is not the case that the rain falls", "Necessarily, the snow falls"},
		{"converse K", "If necessarily, the rain falls, then necessarily, the snow falls", "Necessarily, if the rain falls, then the snow falls"},
		{"possibility to necessity", "Possibly, the rain falls", "Necessarily, the rain falls"},
		{"possible conditional", "Possibly, if the rain falls, then the snow falls", "If possibly, the rain falls, then possibly, the snow falls"},
		{"reverse possibility monotonicity", "If possibly, the rain falls, then possibly, the snow falls", "Necessarily, if the rain falls, then the snow falls"},
		{"mixed modality", "Necessarily, if the rain falls, then the snow falls", "If possibly, the rain falls, then necessarily, the snow falls"},
		{"K reversed clauses", "Necessarily, if the rain falls, then the snow falls", "If necessarily, the snow falls, then necessarily, the rain falls"},
		{"S4 outside T", "Necessarily, the rain falls", "Necessarily, necessarily, the rain falls"},
		{"S5 outside T", "Possibly, the rain falls", "Necessarily, possibly, the rain falls"},
		{"conjunction different clause", "Necessarily, the rain falls and the snow falls", "It is the case both that necessarily, the rain falls, and that necessarily, the hail falls"},
		{"conjunction one merely possible", "It is the case both that necessarily, the rain falls, and that possibly, the snow falls", "Necessarily, the rain falls and the snow falls"},
		{"ambiguous outer conjunction", "Necessarily, the rain falls and the snow falls", "Necessarily, the rain falls, and necessarily, the snow falls"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			argument := fmt.Sprintf("1. %s.\n2. Therefore, %s. [1]", tc.premise, tc.conclusion)
			if _, syntax := ParseNumberedArgument(argument, numberedArgumentOptions{}); !syntax.Passed {
				t.Fatalf("bad fixture: %s", syntax.Message)
			}
			report, result := CheckEntailmentGate(argument)
			if report.Passed || result.Passed {
				t.Fatalf("accepted: %s %+v", argument, report)
			}
		})
	}
}

func TestEntailmentGateV15FitchDuality(t *testing.T) {
	argument := `| 1. Suppose it is not the case that possibly, it is not the case that the rain falls.
| 2. Therefore, necessarily, the rain falls. [1]
3. Therefore, if it is not the case that possibly, it is not the case that the rain falls, then necessarily, the rain falls. [1-2]`
	report, result := CheckEntailmentGate(argument)
	if !report.Passed || !result.Passed || len(report.Steps) != 2 || report.Steps[0].Form != "modal duality" || report.Steps[1].Form != "conditional proof" {
		t.Fatalf("%s %+v", result.Message, report)
	}
}

func TestEntailmentGateV15IdentityScope(t *testing.T) {
	inScope := `| 1. Suppose Socrates is a man.
| 2. Therefore, Socrates is identical to Socrates. [1]
3. Therefore, if Socrates is a man, then Socrates is identical to Socrates. [1-2]`
	report, result := CheckEntailmentGate(inScope)
	if !report.Passed || !result.Passed || len(report.Steps) != 2 || report.Steps[0].Form != "identity reflexivity" {
		t.Fatalf("in-scope term rejected: %s %+v", result.Message, report)
	}
	outOfScope := inScope + "\n4. Therefore, Socrates is identical to Socrates. [1]"
	if report, result := CheckEntailmentGate(outOfScope); report.Passed || result.Passed {
		t.Fatalf("closed assumption accepted: %+v", report)
	}
	// A name occurring elsewhere in the proof does not satisfy the cited-line rule.
	wrongCitation := `1. Socrates is a man.
2. Plato is a man.
3. Therefore, Socrates is identical to Socrates. [2]
4. Therefore, Socrates is a man and Socrates is identical to Socrates. [1,3]`
	if report, result := CheckEntailmentGate(wrongCitation); report.Passed || result.Passed {
		t.Fatalf("uncited name accepted: %+v", report)
	}
}
