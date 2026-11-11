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
	if len(sentences) != 40 {
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
	// Compatibility is deliberately stronger than retroactive Russell semantics:
	// this old negation certificate must retain its complete historical report.
	text := v26Argument([]string{"It is not the case that the present king of France is bald"}, "The present king of France is not bald")
	v27Parity(t, text)
	_, r := CheckEntailmentGate(text)
	if !r.Passed {
		t.Fatal(r.Message)
	}
	g := foAtom("king", "$count")
	p := foAtom("bald", "$count")
	if !finiteCountermodel([]*folFormula{foNot(v27Russell(g, p))}, v27Russell(g, foNot(p)), 3) {
		t.Fatal("expected inherited scope countermodel")
	}
}
