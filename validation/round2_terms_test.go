package validation

import (
	"strings"
	"testing"
)

func TestRound2TermsAudit(t *testing.T) {
	cases := []struct{ probe, hint, argument string }{
		{"binding-irregular-1", "name", "1. If X is red, then X is warm.\n2. The observer dreamt Alpha is red.\n3. Therefore, the observer dreamt Alpha is warm. [1,2]\n"}, {"clause-slot-1", "write the phrase into the rule", "1. If when X is red Bob smiles, then X is warm.\n2. When the observer reckons Alpha is red Bob smiles.\n3. Therefore, the observer reckons Alpha is warm. [1,2]\n"},
		{"intensional-eg-1", "state the existential claim as its own premise", "1. Lois craves Zeus.\n2. Therefore, there is an X such that Lois craves X. [1]\n"},
		{"context-binding-1", "state the existential claim as its own premise", "1. Twelve students greet the teacher.\n2. Therefore, there is an X such that twelve students greet X. [1]\n"},
		{"quantity-ui-1", "every N", "1. Every rodent is an animal.\n2. Six mice are rodents.\n3. Therefore, Six mice are animals. [1,2]\n"}, {"identity-pronoun-1", "name the thing", "1. He is identical to Alpha.\n2. He is identical to Beta.\n3. Therefore, Alpha is identical to Beta. [1,2]\n"},
		{"identity-nonsingular-2", "name the thing", "1. Zero mice is identical to Alpha.\n2. Zero mice is identical to Beta.\n3. Therefore, Alpha is identical to Beta. [1,2]\n"}, {"identity-intensional-2", "state the substituted claim as a premise", "1. Superman is identical to Clark.\n2. Superman is famous as a hero.\n3. Therefore, Clark is famous as a hero. [1,2]\n"},
		{"lexical-letter-1", "write the name out", "1. If Model U is red, then Model U is warm.\n2. Model Alpha is red.\n3. Therefore, Model Alpha is warm. [1,2]\n"}, {"proper-name-mixed-1", "write the name out", "1. If Malcolm X is red, then X is warm.\n2. Malcolm Alpha is red.\n3. Therefore, Alpha is warm. [1,2]\n"},
		{"morphology-5", "every N", "1. All breeches are wet.\n2. Alpha is a breech.\n3. Therefore, Alpha is wet. [1,2]\n"},
	}
	for _, tc := range cases {
		t.Run(tc.probe, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if report.Passed || result.Passed {
				t.Fatalf("audit certified: %+v", report)
			}
			if !strings.Contains(strings.ToLower(result.Message), strings.ToLower(tc.hint)) {
				t.Fatalf("missing rewrite %q: %s", tc.hint, result.Message)
			}
		})
	}
}

func TestRound2TermsControls(t *testing.T) {
	controls := []string{
		"1. If X is red, then X is warm.\n2. The dog is red.\n3. Therefore, the dog is warm. [1,2]",
		"1. The dog is identical to Rex.\n2. Rex is red.\n3. Therefore, something is red. [2]",
		"1. Every rodent is an animal.\n2. The mouse is a rodent.\n3. Therefore, the mouse is an animal. [1,2]",
		"1. The dog is identical to Alpha.\n2. The dog is identical to Beta.\n3. Therefore, Alpha is identical to Beta. [1,2]",
		"1. The man is identical to Clark.\n2. The man is red.\n3. Therefore, Clark is red. [1,2]",
		"1. If the model is red, then the model is warm.\n2. The model is red.\n3. Therefore, the model is warm. [1,2]",
		"1. Every breech is wet.\n2. Alpha is a breech.\n3. Therefore, Alpha is wet. [1,2]",
	}
	for _, s := range controls {
		r, res := CheckEntailmentGate(s)
		if !r.Passed || !res.Passed {
			t.Errorf("control failed: %s\n%s", s, res.Message)
		}
	}
}

func TestRound2TermsWitnessPaths(t *testing.T) {
	for _, s := range []string{
		"1. Lois greets Zeus.\n2. Therefore, there is an X such that Lois greets X. [1]",
		"1. Zeus is sought by Lois.\n2. Therefore, something is sought by Lois. [1]",
		"1. Lois stands near Zeus.\n2. Therefore, there is an X such that Lois stands near X. [1]",
		"1. The dog is red.\n2. Therefore, something is red. [1]",
		"1. Every student greets Zeus.\n2. Therefore, there is an X such that every student greets X. [1]",
		"1. Lois greets Zeus.\n2. Therefore, some x is such that Lois greets x. [1]",
	} {
		r, _ := CheckEntailmentGate(s)
		if r.Passed {
			t.Errorf("unsafe witness certified: %s", s)
		}
	}
}

func TestRound2TermsClosedClasses(t *testing.T) {
	p := allStrictRules()
	for _, n := range []string{"zero", "six", "twelve", "twenty", "thirty", "hundred", "thousand", "million", "dozen", "6", "123"} {
		if p.instanceSubjectAllowed(n+" mice") || p.equalTerms(n+" mice", n+" mice") {
			t.Errorf("numeral term admitted: %s", n)
		}
	}
	for _, letter := range []string{"U", "V", "W", "X", "Y", "Z"} {
		if !letterNearName("If Model "+letter+" is red, then the dog is warm") || !letterNearName("If "+letter+" Model is red, then the dog is warm") {
			t.Errorf("letter not protected: %s", letter)
		}
	}
	for _, pair := range [][2]string{{"breeches", "breech"}, {"dogs", "dog"}, {"analyses", "analysis"}, {"princesses", "princess"}} {
		if p.equalTerms(pair[0], pair[1]) {
			t.Errorf("inflection admitted: %v", pair)
		}
		if !p.without("term.inflection_verbatim").equalTerms(pair[0], pair[1]) {
			t.Errorf("switch failed: %v", pair)
		}
	}
}

func TestRound2TermsIdentityRoutes(t *testing.T) {
	for _, s := range []string{
		"1. He is identical with Alpha.\n2. He is red.\n3. Therefore, Alpha is red. [1,2]",
		"1. The man is Clark.\n2. The man is famous as a hero.\n3. Therefore, Clark is famous as a hero. [1,2]",
		"1. The man is identical to Clark.\n2. The man greets Lois.\n3. Therefore, Clark greets Lois. [1,2]",
		"1. The man is identical to Clark.\n2. The man can fly.\n3. Therefore, Clark can fly. [1,2]",
	} {
		r, _ := CheckEntailmentGate(s)
		if r.Passed {
			t.Errorf("identity bypass: %s", s)
		}
	}
}
