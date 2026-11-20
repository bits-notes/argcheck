package validation

import (
	"strings"
	"testing"
)

var v461ReadingShiftProbes = map[string]string{
	"p05b": "1. If X are in the yard, then X are not in the house.\n2. Dogs are in the yard.\n3. Therefore, dogs are not in the house. [1,2]\n",
	"p05c": "1. If X is in the cup, then X is not in the bottle.\n2. Water is in the cup.\n3. Therefore, water is not in the bottle. [1,2]\n",
	"p05d": "1. If X are barking, then X are not asleep.\n2. Dogs are barking.\n3. Therefore, dogs are not asleep. [1,2]\n",
	"g01":  "1. If X are in the yard, then X are dangerous.\n2. Dogs are in the yard.\n3. Therefore, dogs are dangerous. [1,2]\n",
	"g02":  "1. If X is in the cup, then X is poisonous.\n2. Water is in the cup.\n3. Therefore, water is poisonous. [1,2]\n",
	"g03":  "1. If X are in the yard, then X cannot fly.\n2. Birds are in the yard.\n3. Therefore, birds cannot fly. [1,2]\n",
	"p03b": "1. If X are not in the house, then X are outside.\n2. Dogs are not in the house.\n3. Therefore, dogs are outside. [1,2]\n",
	"p04":  "1. If X is not in the house, then X is outside.\n2. Water is not in the house.\n3. Therefore, water is outside. [1,2]\n",
}

func assertV461SingularRefusal(t *testing.T, text string) {
	t.Helper()
	r, res := CheckEntailmentGate(text, notesOn)
	if r.Passed || res.Passed {
		t.Fatalf("certified: %s", text)
	}
	for _, step := range r.Steps {
		if strings.Contains(step.Mismatch, "slot.value_singular") {
			return
		}
		for _, note := range step.Notes {
			if note.ID == "slot.value_singular" {
				return
			}
		}
	}
	t.Fatalf("missing slot.value_singular refusal: %+v", r.Steps)
}

func TestV461ReadingShiftProbes(t *testing.T) {
	for name, text := range v461ReadingShiftProbes {
		t.Run(name, func(t *testing.T) { assertV461SingularRefusal(t, text) })
	}
}

func TestV461PluralMassFlagOptIn(t *testing.T) {
	p := allStrictRules()
	if p.slot_plural_mass_value {
		t.Fatal("loosening switch enabled by default")
	}
	p.slot_plural_mass_value = true
	if r, res := p.CheckEntailmentGate(v461ReadingShiftProbes["p05b"]); !r.Passed {
		t.Fatal(res.Message)
	}
	if r, _ := p.without("slot.plural_mass_value").CheckEntailmentGate(v461ReadingShiftProbes["p05b"]); r.Passed {
		t.Fatal("rule-id switch did not disable bare values")
	}
}

// The default singular-value guard masks this rule, but opting in must retain it.
func TestV461OptInPreservesPluralWitnessGuard(t *testing.T) {
	p := allStrictRules()
	p.slot_plural_mass_value = true
	text := "1. If X is in the cup, then X is in both vessels.\n2. Water is in the cup.\n3. Therefore, water is in both vessels. [1,2]"
	r, _ := p.CheckEntailmentGate(text)
	if r.Passed || len(r.Steps) != 1 || !strings.Contains(r.Steps[0].Mismatch, "slot.plural_witness") {
		t.Fatalf("missing witness refusal: %+v", r)
	}
	if r, res := p.without("slot.plural_witness").CheckEntailmentGate(text); !r.Passed {
		t.Fatal(res.Message)
	}
}
