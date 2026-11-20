package validation

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

var updateParaphrases = flag.Bool("update-paraphrases", false, "update certified reading golden snapshots")

func TestParaphraseDocsRulesGolden(t *testing.T) {
	raw, err := os.ReadFile("../docs/rules.md")
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, block := range parseRulesDocBlocks(t, string(raw)) {
		if block.kind != "accept" {
			continue
		}
		passed, report, msg := checkRulesDocArgument(block.body)
		if !passed {
			t.Fatalf("%s: %s", block.form, msg)
		}
		lines = append(lines, "FORM: "+block.form)
		for _, step := range report.Steps {
			if !step.Passed {
				continue
			}
			if step.Paraphrase == "" || step.ParaphraseForm != step.Form {
				t.Fatalf("missing reading: %+v", step)
			}
			if strings.Contains(step.Paraphrase, `\u`) || strings.Contains(step.Paraphrase, `["`) || strings.Contains(step.Paraphrase, ":  ⊢") {
				t.Errorf("escaped, bracketed or empty reading: %s", step.Paraphrase)
			}
			// Reuse the gate’s closed pronoun list and impersonal scaffold exemption.
			checked := step.Paraphrase
			for _, loc := range crossStatementWord.FindAllStringIndex(checked, -1) {
				word := checked[loc[0]:loc[1]]
				if strings.EqualFold(word, "it") && crossStatementScaffold(checked[loc[0]:]) {
					continue
				}
				if crossStatementPronoun(word) || crossStatementPossessive(word) {
					t.Errorf("personal pronoun %q in %s", word, step.Paraphrase)
				}
			}
			// Quoted lexical payloads are opaque, not newly asserted English.
			// Check every remaining connective with the production grouping checker.
			if why := groupingProblem(maskReadingLiterals(checked)); why != "" {
				t.Errorf("ungrouped output %s: %s", step.Paraphrase, why)
			}
			lines = append(lines, fmt.Sprintf("step %d: %s", step.Step, step.Paraphrase))
		}
		again, _ := CheckEntailmentGate(block.body)
		for i := range report.Steps {
			if report.Steps[i].Paraphrase != again.Steps[i].Paraphrase {
				t.Fatalf("nondeterministic %s", block.form)
			}
		}
		lines = append(lines, "")
	}
	got := strings.Join(lines, "\n") + "\n"
	path := "testdata/certified_readings.golden"
	if *updateParaphrases {
		if err := os.WriteFile(path, []byte(got), 0644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Fatal("certified readings differ; inspect and regenerate with -update-paraphrases")
	}
}

func maskReadingLiterals(s string) string {
	var b strings.Builder
	quoted, escape := false, false
	for _, r := range s {
		if quoted {
			if escape {
				escape = false
				continue
			}
			if r == '\\' {
				escape = true
				continue
			}
			if r == '"' {
				quoted = false
				b.WriteString("literal")
			}
			continue
		}
		if r == '"' {
			quoted = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestParaphraseRenderings(t *testing.T) {
	a := &folFormula{op: "atom", symbol: "man", terms: []string{"$x"}}
	b := &folFormula{op: "atom", symbol: "human", terms: []string{"$x"}}
	q := fol("every", fol("implies", a, b))
	q.symbol = "$x"
	if got := readingFOL(q); got != `∀X1 (("man"(X1) → "human"(X1)))` {
		t.Fatal(got)
	}
	if got := readingMP("Rex arrived", "Rex was seen"); got != `Premise: if "Rex arrived", then "Rex was seen". Premise: "Rex arrived". So: "Rex was seen".` {
		t.Fatal(got)
	}
	cases := []struct{ argument, contains string }{
		{"1. If X admires Y, then Y trusts X.\n2. Alice admires Bob.\n3. Therefore, Bob trusts Alice. [1,2]", "Bindings:"},
		{"1. Alice sings.\n2. Therefore, either Alice sings or Bob dances. [1]", "either"},
		{"1. Every x is such that x is red.\n2. Therefore, Rex is red. [1]", "∀X1"},
	}
	for _, tc := range cases {
		report, result := CheckEntailmentGate(tc.argument)
		if !result.Passed {
			t.Fatal(result.Message)
		}
		if !strings.Contains(report.Steps[0].Paraphrase, tc.contains) {
			t.Fatal(report.Steps[0].Paraphrase)
		}
	}
}

func TestParaphraseBoundNamesAndReadOnly(t *testing.T) {
	// Constant X1 and a nested shadowing binder must remain distinct.
	inner := fol("some", &folFormula{op: "equal", terms: []string{"$v", "X1"}})
	inner.symbol = "$v"
	outer := fol("every", fol("and", &folFormula{op: "atom", symbol: "P", terms: []string{"$v"}}, inner))
	outer.symbol = "$v"
	before := folKey(outer, nil, 0)
	want := `∀X1 (("P"(X1) ∧ ∃X2 ((X2 = "X1"))))`
	for i := 0; i < 10; i++ {
		if got := readingFOL(outer); got != want {
			t.Fatal(got)
		}
	}
	if before != folKey(outer, nil, 0) {
		t.Fatal("renderer mutated formula")
	}
}

func TestRefusalLegacyFields(t *testing.T) {
	raw, err := os.ReadFile("testdata/refusal_legacy.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name, Argument     string
		Syntax, Entailment map[string]any
	}
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, tc := range fixtures {
		t.Run(tc.Name, func(t *testing.T) {
			syntax := Validate(context.Background(), BuiltinSpec(BuiltinNumberedArgumentSyntax), tc.Argument)
			checkPinnedResult(t, syntax, tc.Syntax)
			report, result := CheckEntailmentGate(tc.Argument, EntailmentGateOptions{StrictNotes: true})
			if tc.Entailment != nil {
				checkPinnedResult(t, result, tc.Entailment)
			}
			found := false
			for _, s := range report.Steps {
				if s.Passed {
					if s.Refusal != nil {
						t.Fatal("refusal on passed step")
					}
					continue
				}
				found = true
				if s.Refusal == nil {
					t.Fatalf("missing refusal: %+v", s)
				}
				r := s.Refusal
				if r.Rule == "" || r.Hint == "" || r.Span.Text == "" {
					t.Fatalf("empty refusal: %+v", r)
				}
				switch tc.Name {
				case "pronoun":
					if r.Rule != "term.pronoun_cross_statement" || r.Span.Step != 2 || r.Span.Text != "it" || r.Span.Start != 3 || r.Span.End != 5 {
						t.Fatalf("pronoun span: %+v", r)
					}
				case "comma":
					if r.Rule != "scope.mixed_andor" || r.Span.Step != 1 || r.Span.Start != 0 || r.Hint != "use either...or / both...and, or parentheses to group" {
						t.Fatalf("comma span: %+v", r)
					}
				case "mismatch":
					if r.Rule != s.NearestForm || r.Hint != s.Mismatch {
						t.Fatalf("mismatch: %+v", r)
					}
				}
			}
			if !found {
				t.Fatal("no failed step")
			}
		})
	}
}
func checkPinnedResult(t *testing.T, result Result, want map[string]any) {
	t.Helper()
	raw, _ := json.Marshal(result)
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	stripReadingFields(got)
	// Version provenance is deliberately bumped; all old diagnostic fields stay.
	if d, ok := got["details"].(map[string]any); ok {
		if _, ok := d["catalog"]; ok {
			d["catalog"] = "v46.2"
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("legacy fields changed\ngot: %s\nwant: %v", raw, want)
	}
}
func stripReadingFields(v any) {
	switch x := v.(type) {
	case map[string]any:
		delete(x, "paraphrase")
		delete(x, "paraphrase_form")
		delete(x, "refusal")
		// Syntax errors had no steps before v46.3.
		if _, ok := x["sentence"]; ok {
			delete(x, "steps")
		}
		for _, c := range x {
			stripReadingFields(c)
		}
	case []any:
		for _, c := range x {
			stripReadingFields(c)
		}
	}
}

func TestRefusalRuneOffsets(t *testing.T) {
	s := NumberedArgumentStep{Number: 4, Statement: "Éva sees Ωmega"}
	span, ok := refusalExcerpt(s, "Ωmega")
	if !ok || span.Start != 9 || span.End != 14 || span.Text != "Ωmega" {
		t.Fatalf("%+v", span)
	}
}

func TestRefusalWithoutNotesAndValidateSurface(t *testing.T) {
	text := "1. If he sings, then the bell rings.\n2. He sings.\n3. Therefore, the bell rings. [1,2]"
	report, result := CheckEntailmentGate(text)
	if result.Passed || len(report.Steps) != 1 {
		t.Fatal(result)
	}
	s := report.Steps[0]
	if len(s.Notes) != 0 || s.Refusal == nil || s.Refusal.Rule != "term.pronoun_cross_statement" || s.Paraphrase != "" || s.ParaphraseForm != "" {
		t.Fatalf("%+v", s)
	}
	checked := Validate(context.Background(), BuiltinSpec(BuiltinEntailmentGate), text)
	raw, _ := json.Marshal(checked)
	if !strings.Contains(string(raw), `"refusal"`) {
		t.Fatal(string(raw))
	}
	accepted := "1. If the bell rings, then the door opens.\n2. The bell rings.\n3. Therefore, the door opens. [1,2]"
	checked = Validate(context.Background(), BuiltinSpec(BuiltinEntailmentGate), accepted)
	raw, _ = json.Marshal(checked)
	if !checked.Passed || !strings.Contains(string(raw), `"paraphrase_form":"modus ponens"`) {
		t.Fatal(string(raw))
	}
}

func TestParaphrasePostposedAndDischarge(t *testing.T) {
	text := "1. The door opens if the bell rings.\n2. The bell rings.\n3. Therefore, the door opens. [1,2]"
	report, result := CheckEntailmentGate(text)
	if !result.Passed {
		t.Fatal(result.Message)
	}
	if !strings.Contains(report.Steps[0].Paraphrase, `if "the bell rings", then "the door opens"`) {
		t.Fatal(report.Steps[0].Paraphrase)
	}
	text = "| 1. Suppose Alice sings.\n| 2. Therefore, Alice sings. [1]\n3. Therefore, if Alice sings, then Alice sings. [1-2]"
	report, result = CheckEntailmentGate(text)
	if !result.Passed {
		t.Fatal(result.Message)
	}
	if !strings.Contains(report.Steps[len(report.Steps)-1].Paraphrase, `assume "`) {
		t.Fatal(report.Steps)
	}
}

func TestParaphraseNecessitation(t *testing.T) {
	text := "1. Necessarily, if rain falls, then streets get wet.\n2. Necessarily, rain falls.\n3. Therefore, necessarily, streets get wet. [1,2]"
	r, result := CheckEntailmentGate(text)
	if !result.Passed {
		t.Fatal(result.Message)
	}
	s := r.Steps[0]
	if s.ParaphraseForm != "modus ponens (necessitation)" || !strings.Contains(s.Paraphrase, "□ on each premise and on the conclusion") {
		t.Fatalf("lost modal scope: %+v", s)
	}
}

func TestParaphraseCanonicalCategorical(t *testing.T) {
	cases := []struct{ argument, want string }{
		{"1. Every man is mortal.\n2. a is a man.\n3. Therefore, a is mortal. [1,2]", `For every X1: if X1 is a man, then X1 is mortal. "a" is a man. So: "a" is mortal.`},
		{"1. Every man is mortal.\n2. It is not the case that Zeus is mortal.\n3. Therefore, it is not the case that Zeus is a man. [1,2]", `For every X1: if X1 is a man, then X1 is mortal. it is not the case that "zeus" is mortal. So: it is not the case that "zeus" is a man.`},
		{"1. Some cat is a hunter.\n2. No hunter is lazy.\n3. Therefore, some cat is not lazy. [1,2]", `there is an X1 such that both (X1 is in the class "cat") and (X1 is a hunter). For every X1: if X1 is a hunter, then it is not the case that X1 is lazy. So: there is an X1 such that both (X1 is in the class "cat") and (it is not the case that X1 is lazy).`},
		{"1. Every cat is a hunter.\n2. Every hunter is alert.\n3. Therefore, every cat is alert. [1,2]", `For every X1: if X1 is in the class "cat", then X1 is a hunter. For every X1: if X1 is in the class "hunter", then X1 is alert. So: For every X1: if X1 is in the class "cat", then X1 is alert.`},
		{"1. Every cat is a hunter.\n2. No hunter is lazy.\n3. Therefore, no cat is lazy. [1,2]", `For every X1: if X1 is in the class "cat", then X1 is a hunter. For every X1: if X1 is a hunter, then it is not the case that X1 is lazy. So: For every X1: if X1 is in the class "cat", then it is not the case that X1 is lazy.`},
	}
	for _, tc := range cases {
		report, result := CheckEntailmentGate(tc.argument)
		if !result.Passed {
			t.Fatal(result.Message)
		}
		if got := report.Steps[0].Paraphrase; got != tc.want {
			t.Errorf("got %s\nwant %s", got, tc.want)
		}
	}
}

func TestParaphraseUsesRecordedSlots(t *testing.T) {
	// Deliberately unusable surface text: rendering must use only the witness.
	ps := []NumberedArgumentStep{{Number: 1, Statement: "unrelated"}, {Number: 2, Statement: "unrelated"}}
	c := NumberedArgumentStep{Number: 3, Statement: "unrelated"}
	ids := []clauseIdentity{
		{"premise 1 class", "man", "premise 2 predicate", "a man"},
		{"premise 2 instance", "a", "conclusion 3 instance", "a"},
		{"premise 1 predicate", "mortal", "conclusion 3 predicate", "mortal"},
	}
	want := `For every X1: if X1 is a man, then X1 is mortal. "a" is a man. So: "a" is mortal.`
	if got := readingApplication("universal instantiation", ps, c, ids); got != want {
		t.Fatal(got)
	}
	if got := readingApplication("universal instantiation", ps, c, nil); !strings.HasPrefix(got, "SEMI-FORMAL:") {
		t.Fatal(got)
	}
	for _, literal := range []string{"it is not the case that Zeus is mortal", "it is the case both that rain falls, and that snow falls", "I", "雪"} {
		if got := readingLiteral(literal); got != `"`+literal+`"` {
			t.Fatal(got)
		}
	}
	// English negation must not rewrite a symbol inside a quoted clause.
	if got := readingNegations(`¬"¬P"`); got != `it is not the case that ("¬P")` {
		t.Fatal(got)
	}
}

func TestParaphraseDocsExamplesAreGolden(t *testing.T) {
	doc, err := os.ReadFile("../docs/rules.md")
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/certified_readings.golden")
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(doc), "## Paraphrase of the certified reading")
	if !ok {
		t.Fatal("missing paraphrase documentation")
	}
	count := 0
	for _, line := range strings.Split(section, "\n") {
		if strings.HasPrefix(line, "step ") {
			count++
			if !strings.Contains("\n"+string(golden), "\n"+line+"\n") {
				t.Errorf("example absent from golden: %s", line)
			}
		}
	}
	if count != 4 {
		t.Fatalf("want four real examples, got %d", count)
	}
}
