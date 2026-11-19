package validation

import (
	"strings"
	"testing"
)

func TestCounterfactualWeakeningPolarity(t *testing.T) {
	cf := v19MatchConditional
	atom := "the room is warm"
	not := func(s string) string { return "it is not the case that " + s }
	iff := func(a, b string) string { return "(" + a + ") if and only if (" + b + ")" }
	imp := func(a, b string) string { return "if (" + a + "), then " + b }
	and := func(a, b string) string { return "it is the case both that " + a + ", and that " + b }
	for _, tc := range []struct {
		name, s                 string
		premiseOK, conclusionOK bool
	}{
		{"positive", cf, true, false},
		{"negative", not(cf), false, true},
		{"double negative", not(not(cf)), true, false},
		{"material antecedent", imp(cf, atom), false, true},
		{"material consequent", imp(atom, cf), true, false},
		{"nested antecedent flips twice", imp(imp(cf, atom), atom), true, false},
		{"conjunction", and(cf, atom), true, false},
		{"disjunction", "either " + cf + " or " + atom, true, false},
		{"negated conjunction", not(and(cf, atom)), false, true},
		{"iff positive and negative", iff(cf, atom), false, false},
		{"negated iff still both", not(iff(cf, atom)), false, false},
		{"standalone would atom", "the death of the child would be a very bad thing", true, true},
		{"would atom in conjunction", and("the child would be safe", atom), true, true},
		{"could not have remains ordinary text", "if the threat is severe, then he could not have avoided the action", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := []NumberedArgumentStep{{Statement: tc.s}}
			if got := v19PolarityFailure(p, NumberedArgumentStep{Statement: atom}) == ""; got != tc.premiseOK {
				t.Errorf("premise safe=%v, want %v", got, tc.premiseOK)
			}
			if got := v19PolarityFailure(nil, NumberedArgumentStep{Statement: tc.s}) == ""; got != tc.conclusionOK {
				t.Errorf("conclusion safe=%v, want %v", got, tc.conclusionOK)
			}
		})
	}
}
func TestCounterfactualMixedMaterialRules(t *testing.T) {
	for _, tc := range []v19GateCase{
		{"material contradiction entails negative cf", `1. It is not the case that if the match was struck, then the match lit.
| 2. Suppose if the match had been struck, then the match would have lit.
| 3. Therefore, if the match was struck, then the match lit. [2]
| 4. Therefore, contradiction. [1,3]
5. Therefore, it is not the case that if the match had been struck, then the match would have lit. [2-4]`, "reductio", ""},

		{"schematic nested if with omitted then", `1. The principle that if help is possible, then we should help takes no account of distance.
2. If X takes no account of distance, then if X were acted upon, even in its qualified form, our world would be changed.
3. Therefore, if the principle that if help is possible, then we should help were acted upon, even in its qualified form, our world would be changed. [1,2]`, "schematic instantiation", ""},

		{"opaque schematic cf consequent", `1. Alice is ready.
2. If X is ready, then if X had been struck, then X would have lit.
3. Therefore, if Alice had been struck, then Alice would have lit. [1,2]`, "schematic instantiation", ""},
		{"opaque n-ary conjunction", `1. If the match had been struck, then the match would have lit.
2. The child would be safe.
3. The room is warm.
4. Therefore, if the match had been struck, then the match would have lit and the child would be safe and the room is warm. [1,2,3]`, "n-ary conjunction introduction", ""},

		{"mixed HS", `1. If the match had been struck, then the match would have lit.
2. If the match would have lit, then the room is warm.
3. Therefore, if the match had been struck, then the room is warm. [1,2]`, "hypothetical syllogism", ""},
		{"verbatim MP", `1. If the match had been struck, then the match would have lit.
2. The match had been struck.
3. Therefore, the match would have lit. [1,2]`, "modus ponens", ""},
		{"verbatim MT", `1. If the match had been struck, then the match would have lit.
2. It is not the case that the match would have lit.
3. Therefore, it is not the case that the match had been struck. [1,2]`, "modus tollens", ""},
		{"negative conclusion via double negation", `1. It is not the case that if the match had been struck, then the match would have lit.
2. Therefore, it is not the case that it is not the case that it is not the case that if the match had been struck, then the match would have lit. [1]`, "double negation introduction", ""},
		{"positive copied conjunct", `1. If the match had been struck, then the match would have lit.
2. The room is warm.
3. Therefore, it is the case both that if the match had been struck, then the match would have lit, and that the room is warm. [1,2]`, "conjunction introduction", ""},
		{"negative conditional use", `1. It is not the case that if the match had been struck, then the match would have lit.
2. Therefore, the match had been struck. [1]`, "", "negative premise position"},
		{"positive nested conclusion", `1. If the match had been struck and the match had been dry, then the match would have lit.
2. Therefore, if the match had been struck, then if the match had been dry, then the match would have lit. [1]`, "", "positive conclusion position"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// slot.value_singular: the embedded-conditional noun phrase is not an admitted term.
			refused := map[string]strictRefusal{"TestCounterfactualMixedMaterialRules/schematic_nested_if_with_omitted_then": {"slot.value_singular", "Name the thing"}}
			if listedStrictRefusal(t, tc.argument, refused) {
				return
			}
			report, result := CheckEntailmentGate(tc.argument)
			if len(report.Steps) == 0 {
				t.Fatal(result.Message)
			}
			last := report.Steps[len(report.Steps)-1]
			if tc.form != "" {
				if !report.Passed || last.Form != tc.form {
					t.Fatalf("want %s: %+v", tc.form, report)
				}
			} else if report.Passed || !strings.Contains(last.Mismatch, tc.note) {
				t.Fatalf("want refusal %q: %+v", tc.note, report)
			}
		})
	}
}
