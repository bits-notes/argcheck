package validation

import (
	"strings"
	"testing"
)

const v19MatchConditional = "If the match had been struck, then the match would have lit"

type v19GateCase struct{ name, argument, form, note string }

// Shared with the gap ledger: these are accepted forms, not gap placeholders.
var v19AcceptedCases = []v19GateCase{
	{"material contraposition", `1. If the match had been struck, then the match would have lit.
2. Therefore, if it is not the case that the match would have lit, then it is not the case that the match had been struck. [1]`, "contraposition", ""},
	{"predictive would verbatim MP", `1. If he left at noon, then he would arrive by three.
2. He left at noon.
3. Therefore, he would arrive by three. [1,2]`, "modus ponens", ""},
	{"biconditional introduction via weakening", `1. If the match had been struck, then the match would have lit.
2. If the match would have lit, then the match had been struck.
3. Therefore, (the match had been struck) if and only if (the match would have lit). [1,2]`, "biconditional introduction", ""},
	{"schematic verbatim MP", `1. If X had been struck, then X would have lit.
2. The match had been struck.
3. Therefore, the match would have lit. [1,2]`, "schematic instantiation", ""},

	{"schematic preserves lowercase constants", `1. If X had struck x, then X would have lit.
2. Alice struck x.
3. Therefore, Alice lit. [1,2]`, "counterfactual modus ponens", ""},

	{"reductio treats counterfactual as whole proposition", `1. If the match had been struck, then the match would have lit.
| 2. Suppose it is not the case that if the match had been struck, then the match would have lit.
| 3. Therefore, contradiction. [1,2]
4. Therefore, if the match had been struck, then the match would have lit. [2-3]`, "reductio", ""},

	{"MP conjunction content", `1. If the match had been struck, then the match would have lit and the room would have warmed.
2. The match was struck.
3. Therefore, the match lit and the room warmed. [1,2]`, "counterfactual modus ponens", ""},

	{"MP passive", `1. If the match had been struck, then the match would have lit.
2. The match was struck.
3. Therefore, the match lit. [1,2]`, "counterfactual modus ponens", ""},
	{"MP plural", `1. If the matches had been struck, then the matches would have been lit.
2. The matches were struck.
3. Therefore, the matches were lit. [1,2]`, "counterfactual modus ponens", ""},
	{"MP negative antecedent", `1. If Oswald had not shot Kennedy, then someone else would have shot Kennedy.
2. It is not the case that Oswald shot Kennedy.
3. Therefore, someone else shot Kennedy. [1,2]`, "counterfactual modus ponens", ""},
	{"MP negative passive", `1. If the match had not been struck, then the match would not have been lit.
2. It is not the case that the match was struck.
3. Therefore, it is not the case that the match was lit. [1,2]`, "counterfactual modus ponens", ""},
	{"MP negative perfect", `1. If the match had not lit, then the match would not have warmed the room.
2. It is not the case that the match lit.
3. Therefore, it is not the case that the match warmed the room. [1,2]`, "counterfactual modus ponens", ""},
	{"were and would be", `1. If I were ready, then I would be happy.
2. I am ready.
3. Therefore, I am happy. [1,2]`, "counterfactual modus ponens", ""},
	{"had lexical had", `1. If Alice had had a match, then Alice would have made a fire.
2. Alice had a match.
3. Therefore, Alice made a fire. [1,2]`, "counterfactual modus ponens", ""},
	{"had said", `1. If Alice had said hello, then Bob would have said goodbye.
2. Alice said hello.
3. Therefore, Bob said goodbye. [1,2]`, "counterfactual modus ponens", ""},
	{"MT", `1. If the match had been struck, then the match would have lit.
2. It is not the case that the match lit.
3. Therefore, it is not the case that the match was struck. [1,2]`, "counterfactual modus tollens", ""},
	{"material if", `1. If the match had been struck, then the match would have lit.
2. Therefore, if the match was struck, then the match lit. [1]`, "counterfactual to material", ""},
	{"material or", `1. If the match had been struck, then the match would have lit.
2. Therefore, either it is not the case that the match was struck or the match lit. [1]`, "counterfactual to material", ""},
	{"agglomeration", `1. If the match had been struck, then the match would have lit.
2. If the match had been struck, then the room would have warmed.
3. Therefore, if the match had been struck, then the match would have lit and the room would have warmed. [1,2]`, "counterfactual consequent conjunction", ""},
	{"conjunction elimination left", `1. If the match had been struck, then the match would have lit and the room would have warmed.
2. Therefore, if the match had been struck, then the match would have lit. [1]`, "counterfactual consequent conjunction elimination", ""},
	{"conjunction elimination right", `1. If the match had been struck, then the match would have lit and the room would have warmed.
2. Therefore, if the match had been struck, then the room would have warmed. [1]`, "counterfactual consequent conjunction elimination", ""},
	{"weakening", `1. If the match had been struck, then the match would have lit.
2. Therefore, if the match had been struck, then either the match would have lit or the room would have warmed. [1]`, "counterfactual consequent disjunction introduction", ""},
	{"bare verb closure", `1. If the match had been struck, then the match would light.
2. Therefore, if the match had been struck, then either the match would light or the room would warm. [1]`, "counterfactual consequent disjunction introduction", ""},
	{"simple dilemma", `1. Either the match was struck or the lighter was lit.
2. If the match had been struck, then the room would have warmed.
3. If the lighter had been lit, then the room would have warmed.
4. Therefore, the room warmed. [1,2,3]`, "counterfactual dilemma", ""},
	{"constructive dilemma", `1. Either the match was struck or the lighter was lit.
2. If the match had been struck, then the match would have lit.
3. If the lighter had been lit, then the room would have warmed.
4. Therefore, either the match lit or the room warmed. [1,2,3]`, "counterfactual dilemma", ""},
	{"destructive dilemma", `1. Either it is not the case that the match lit or it is not the case that the room warmed.
2. If the match had been struck, then the match would have lit.
3. If the lighter had been lit, then the room would have warmed.
4. Therefore, either it is not the case that the match was struck or it is not the case that the lighter was lit. [1,2,3]`, "counterfactual dilemma", ""},
	{"reflexivity", `1. The match was struck.
2. Therefore, if the match had been struck, then the match would have been struck. [1]`, "counterfactual reflexivity", ""},
	{"reflexivity occurrence", `1. Either the match was struck or the lighter was lit.
2. Therefore, if the match had been struck, then the match would have been struck. [1]`, "counterfactual reflexivity", ""},
	{"schematic substitution", `1. If X had been struck, then X would have lit.
2. The match was struck.
3. Therefore, the match lit. [1,2]`, "counterfactual modus ponens", ""},
}

var v19RejectedCases = []v19GateCase{
	{"genuine counterfactual contraposition", `1. If the match had been struck, then the match would have lit.
2. Therefore, if the match had not lit, then the match would not have been struck. [1]`, "", "contraposition"},
	{"HS plain-content middle still cannot produce cf", `1. If Hoover had been a communist, then Hoover would have been a traitor.
2. If Hoover had been a traitor, then Hoover would have been shot.
3. Therefore, if Hoover had been a communist, then Hoover would have been shot. [1,2]`, "", "hypothetical syllogism"},

	{"schematic lowercase constant is not a slot", `1. If X had struck x, then x would have lit.
2. Alice struck Alice.
3. Therefore, Alice lit. [1,2]`, "", "counterfactual"},

	{"embedded would report cannot detach", `1. If the match had been struck, then Ann thinks the match would have lit.
2. The match was struck.
3. Therefore, Ann thinks the match lit. [1,2]`, "", "visible bridging premise"},
	{"unscaffolded reported conjunction cannot weaken", `1. If the match had been struck, then Ann would have thought the lamp lit and the room warmed.
2. Therefore, if the match had been struck, then the room would have warmed. [1]`, "", "counterfactual"},

	{"conditional proof strengthening", `1. If the match had been struck, then the match would have lit.
| 2. Suppose the match had been struck and the match had been wet.
| 3. Therefore, the match had been struck. [2]
| 4. Therefore, the match would have lit. [1,3]
5. Therefore, if the match had been struck and the match had been wet, then the match would have lit. [2-4]`, "", "conditional proof"},
	{"conditional proof false antecedent", `1. It is not the case that the match had been struck.
| 2. Suppose the match had been struck.
| | 3. Suppose it is not the case that the match would have lit.
| | 4. Therefore, contradiction. [1,2]
| 5. Therefore, the match would have lit. [3-4]
6. Therefore, if the match had been struck, then the match would have lit. [2-5]`, "", "conditional proof"},
	{"exportation", `1. If the match had been struck and the match had been dry, then the match would have lit.
2. Therefore, if the match had been struck, then if the match had been dry, then the match would have lit. [1]`, "", "exportation"},
	{"importation", `1. If the match had been struck, then if the match had been dry, then the match would have lit.
2. Therefore, if the match had been struck and the match had been dry, then the match would have lit. [1]`, "", "importation"},
	{"negated conditional antecedent", `1. It is not the case that if Oswald had not shot Kennedy, then someone else would have shot Kennedy.
2. Therefore, Oswald had not shot Kennedy. [1]`, "", "negated conditional"},
	{"negated conditional consequent", `1. It is not the case that if Oswald had not shot Kennedy, then someone else would have shot Kennedy.
2. Therefore, it is not the case that someone else would have shot Kennedy. [1]`, "", "negated conditional"},
	{"or to if", `1. Either it is not the case that the match had been struck or the match would have lit.
2. Therefore, if the match had been struck, then the match would have lit. [1]`, "", "material implication"},
	{"conditional proof true consequent", `1. The match would have lit.
| 2. Suppose the match had been struck.
| 3. Therefore, the match would have lit. [1]
4. Therefore, if the match had been struck, then the match would have lit. [2-3]`, "", "conditional proof"},
	{"hypothetical syllogism", `1. If Hoover had been a communist, then Hoover would have been a traitor.
2. If Hoover would have been a traitor, then Hoover would have been loyal to the USSR.
3. Therefore, if Hoover had been a communist, then Hoover would have been loyal to the USSR. [1,2]`, "", "hypothetical syllogism"},
	{"material to counterfactual", `1. If the match was struck, then the match lit.
2. Therefore, if the match had been struck, then the match would have lit. [1]`, "", "counterfactual"},
	{"strong centering", `1. The match was struck.
2. The match lit.
3. Therefore, if the match had been struck, then the match would have lit. [1,2]`, "", "strong centering"},
	{"disjunctive antecedent simplification", `1. If either the match had been struck or the lighter had been lit, then the room would have warmed.
2. Therefore, if the match had been struck, then the room would have warmed. [1]`, "", "counterfactual"},
	{"conditional excluded middle", `1. The match was struck.
2. Therefore, either if the match had been struck, then the match would have lit or if the match had been struck, then it is not the case that the match would have lit. [1]`, "", "conditional excluded middle"},
	{"might to would", `1. If the match had been struck, then the match might have lit.
2. Therefore, if the match had been struck, then the match would have lit. [1]`, "", "Might/could-have"},
	{"might MP", `1. If the match had been struck, then the match might have lit.
2. The match had been struck.
3. Therefore, the match might have lit. [1,2]`, "", "modus ponens"},
	{"could have MP", `1. If the match had been struck, then the match could have lit.
2. The match had been struck.
3. Therefore, the match could have lit. [1,2]`, "", "modus ponens"},
	{"bare verb bridge", `1. If the match had been struck, then the match would light.
2. The match was struck.
3. Therefore, the match lights. [1,2]`, "", "visible bridging premise"},
	{"irregular verb bridge", `1. If Alice had gone home, then Bob would have lit a fire.
2. Alice went home.
3. Therefore, Bob lit a fire. [1,2]`, "", "visible bridging premise"},
	{"mood not globally normalized", `1. The match had been struck.
2. Therefore, the match was struck. [1]`, "", ""},
	{"no unrelated reflexivity", `1. The lighter was lit.
2. Therefore, if the match had been struck, then the match would have been struck. [1]`, "", "counterfactual"},
	{"biconditional elimination blocked", `1. (The match had been struck) if and only if (the match would have lit).
2. Therefore, if the match had been struck, then the match would have lit. [1]`, "", "biconditional elimination"},
	{"absorption blocked", `1. If the match had been struck, then the match would have lit.
2. Therefore, if the match had been struck, then the match had been struck and the match would have lit. [1]`, "", "absorption"},
}

func TestCatalogV19(t *testing.T) {
	if EntailmentGateCatalogVersion != "v19" {
		t.Fatalf("catalog = %s", EntailmentGateCatalogVersion)
	}
	for _, tc := range append(append([]v19GateCase{}, v19AcceptedCases...), v19RejectedCases...) {
		t.Run(tc.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if len(report.Steps) == 0 {
				t.Fatalf("no gated steps: %s", result.Message)
			}
			last := report.Steps[len(report.Steps)-1]
			if tc.form != "" {
				if !report.Passed || last.Form != tc.form {
					t.Fatalf("want %s: %+v (%s)", tc.form, report, result.Message)
				}
			} else {
				if last.Passed {
					t.Fatalf("excluded final step passed: %+v", last)
				}
				if tc.note != "" && !strings.Contains(last.Mismatch, tc.note) {
					t.Fatalf("note missing %q: %+v", tc.note, last)
				}
			}
		})
	}
}

func TestCatalogV19AuxiliaryMapping(t *testing.T) {
	for _, tc := range []struct {
		clause string
		want   []string
	}{
		{"the match had been struck", []string{"the match was struck", "the match were struck"}},
		{"Oswald had not shot Kennedy", []string{"it is not the case that oswald shot kennedy"}},
		{"the match would have lit", []string{"the match lit"}},
		{"Alice had gone", []string{"alice gone"}},
		{"the match would light", nil},
		{"the match might have lit", nil},
	} {
		got := counterfactualPlain(tc.clause)
		if strings.Join(got, "|") != strings.Join(tc.want, "|") {
			t.Errorf("%q: %v, want %v", tc.clause, got, tc.want)
		}
	}
}

func TestCatalogV19BridgeNamesBothClauses(t *testing.T) {
	for _, tc := range []struct{ argument, left, right string }{
		{`1. If the match had been struck, then the match would light.
2. The match was struck.
3. Therefore, the match lights. [1,2]`, "the match would light", "the match lights"},
		{`1. If Alice had gone home, then Bob would have lit a fire.
2. It is not the case that Bob lit a fire.
3. Therefore, it is not the case that Alice went home. [1,2]`, "alice had gone home", "alice went home"},
	} {
		report, result := CheckEntailmentGate(tc.argument)
		if report.Passed || len(report.Steps) != 1 {
			t.Fatalf("expected failed step: %s", result.Message)
		}
		note := strings.ToLower(report.Steps[0].Mismatch)
		for _, want := range []string{tc.left, tc.right, "visible bridging premise"} {
			if !strings.Contains(note, want) {
				t.Errorf("note missing %q: %s", want, note)
			}
		}
	}
}

func TestCatalogV19WouldInAntecedentDoesNotChangeKind(t *testing.T) {
	argument := `1. If Alice would have lit a fire, then the room is warm.
2. The room is warm.
3. Therefore, Alice lit a fire. [1,2]`
	report, result := CheckEntailmentGate(argument)
	if report.Passed || len(report.Steps) != 1 {
		t.Fatalf("expected a failed material inference: %s", result.Message)
	}
	if strings.Contains(report.Steps[0].Mismatch, "is a counterfactual") {
		t.Fatalf("antecedent mood determined the kind: %+v", report.Steps[0])
	}
}
