package validation

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func v27Parity(t *testing.T, text string) {
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
		want := gateConclusionStepV26(s, steps)
		got := gateConclusionStep(s, steps)
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("%s\nv26: %+v\nv27: %+v", text, want, got)
		}
	}
}
func TestCatalogV27CorpusRegression(t *testing.T) {
	data, e := os.ReadFile("testdata/descriptions/corpus-sentences.txt")
	if e != nil {
		t.Fatal(e)
	}
	sentences := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(sentences) != 24 {
		t.Fatal(len(sentences))
	}
	for i, s := range sentences {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			v27Parity(t, v26Argument([]string{"If p then " + s, "p"}, s))
			v27Parity(t, v26Argument([]string{"q"}, s))
		})
	}
}
func TestCatalogV27InheritedNegation(t *testing.T) {
	// scope.description_negation: the known empty-restrictor countermodel now requires refusal.
	text := v26Argument([]string{"It is not the case that the present king of France is bald"}, "The present king of France is not bald")
	v27Parity(t, text)
	refused := []string{text}
	for _, argument := range refused {
		assertStrictRefusal(t, argument, "scope.description_negation", "keep the wide negation")
	}
	g := foAtom("king", "$count")
	p := foAtom("bald", "$count")
	if !finiteCountermodel([]*folFormula{foNot(v27Russell(g, p))}, v27Russell(g, foNot(p)), 3) {
		t.Fatal("expected inherited scope countermodel")
	}
}

func TestCatalogV27PropositionalScopeRegression(t *testing.T) {
	// All ten accepted-invalid candidates from the propositional oracle.
	for _, tc := range []struct{ premise, conclusion string }{
		{"either it rains or it rains and the street is wet", "the street is wet"},
		{"either it rains or it snows and the street is wet", "the street is wet"},
		{"either it rains or the street is wet and it snows", "it snows"},
		{"either it rains or the street is wet and the street is wet", "the street is wet"},
		{"either it snows or it rains and the street is wet", "the street is wet"},
		{"either it snows or it snows and the street is wet", "the street is wet"},
		{"either it snows or the street is wet and it rains", "it rains"},
		{"either it snows or the street is wet and the street is wet", "the street is wet"},
		{"either the street is wet or the street is wet and it rains", "it rains"},
		{"either the street is wet or the street is wet and it snows", "it snows"},
	} {
		t.Run(tc.premise, func(t *testing.T) {
			text := v26Argument([]string{tc.premise}, tc.conclusion)
			_, r := CheckEntailmentGate(text)
			if r.Passed {
				t.Fatalf("invalid projection accepted: %s", text)
			}
			steps := map[int]NumberedArgumentStep{1: {Number: 1, Statement: tc.premise}}
			if gateDescriptions(NumberedArgumentStep{Number: 2, Statement: tc.conclusion, Citations: []int{1}}, steps).Passed {
				t.Fatal("description fallback accepted invalid projection")
			}
		})
	}
}

func TestCatalogV27PredicateTermRegression(t *testing.T) {
	for _, tail := range []string{"is wet or it rains", "is wet and it rains", "wobbles", "signed"} {
		if f, err := descriptionPredication(tail); err == nil {
			t.Errorf("non-atomic or opaque tail %q accepted as %s", tail, f.Render())
		}
	}
	for _, tail := range []string{"is wet", "is not wet", "wrote Ivanhoe", "is identical to Scott"} {
		if _, err := descriptionPredication(normalizeClause(tail)); err != nil {
			t.Errorf("term-bearing tail %q: %v", tail, err)
		}
	}
	// Explicit parentheses preserve the disjunction: its right conjunct cannot
	// be projected through the disjunction even with a real description term.
	text := v26Argument([]string{"(it rains) or ((the street is wet) and (it snows))"}, "it snows")
	_, r := CheckEntailmentGate(text)
	if r.Passed {
		t.Fatal("parenthesized disjunction projected a conjunct")
	}
}
