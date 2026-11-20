package validation

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestV465IndefiniteProbes(t *testing.T) {
	for _, text := range []string{
		"1. Every poet is mortal.\n2. A senator is a poet.\n3. Therefore, a senator is mortal. [1,2]",
		"1. Every guard is awake.\n2. A guard is a guard.\n3. Therefore, a guard is awake. [1,2]",
		"1. Every guard is awake.\n2. An officer is a guard.\n3. Therefore, an officer is awake. [1,2]",
		"1. Every guard who is awake is alert.\n2. An officer is a guard who is awake.\n3. Therefore, an officer is alert. [1,2]",
		"1. Every officer is a guard.\n2. Every guard is awake.\n3. Therefore, every officer is awake. [1,2]\n4. An officer is an officer.\n5. Therefore, an officer is awake. [3,4]",
	} {
		r, res := CheckEntailmentGate(text, notesOn)
		if r.Passed {
			t.Fatalf("certified: %s", text)
		}
		found := false
		for _, step := range r.Steps {
			for _, note := range step.Notes {
				found = found || note.ID == "term.indefinite_singular"
			}
		}
		if !found {
			t.Errorf("missing note: %s: %+v; %s", text, r.Steps, res.Message)
		}
	}
}

func TestV465Controls(t *testing.T) {
	for _, text := range []string{
		"1. Every guard is an officer.\n2. Rex is a guard.\n3. Therefore, Rex is an officer. [1,2]",
		"1. Some officer is a guard.\n2. Every guard is awake.\n3. Therefore, some officer is awake. [1,2]",
		"1. Every guard is awake.\n2. The officer is a guard.\n3. Therefore, the officer is awake. [1,2]",
		"1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]",
		"1. If a guard sleeps, then the bell rings.\n2. A guard sleeps.\n3. Therefore, the bell rings. [1,2]",
	} {
		if r, res := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("%s: %s", text, res.Message)
		}
	}
}

func TestV465ConstantRoutes(t *testing.T) {
	// Every route has a concrete name control; existing refusals keep their
	// original diagnostics and need not gain a second, redundant strict note.
	probes := map[string]string{
		"property":         "1. Rex is wise.\n2. Therefore, Rex has the property of being wise. [1]",
		"ui":               "1. Every guard is awake.\n2. Rex is a guard.\n3. Therefore, Rex is awake. [1,2]",
		"relative":         "1. Every guard who is awake is alert.\n2. Rex is a guard who is awake.\n3. Therefore, Rex is alert. [1,2]",
		"mt":               "1. Every guard is awake.\n2. Rex is not awake.\n3. Therefore, Rex is not a guard. [1,2]",
		"eg":               "1. The guard is identical to Rex.\n2. Rex is red.\n3. Therefore, something is red. [2]",
		"identity":         "1. McRex is identical to Bob.\n2. McRex is red.\n3. Therefore, Bob is red. [1,2]",
		"relation_subject": "1. Every guard admires Bob.\n2. Rex is a guard.\n3. Therefore, Rex admires Bob. [1,2]",
		"relation_object":  "1. Every guard admires Rex.\n2. Bob is a guard.\n3. Therefore, Bob admires Rex. [1,2]",
		"count":            "1. Exactly one guard.\n2. Rex is a guard.\n3. Therefore, no guard other than Rex is a guard. [1,2]",
		"fitch":            "1. Every guard is awake.\n| 2. Suppose Rex is a guard.\n| 3. Therefore, Rex is awake. [1,2]\n4. Therefore, if Rex is a guard, then Rex is awake. [2-3]",
		"literal":          "1. every x is such that atom{\"awake\", x}.\n2. Therefore, atom{\"awake\", \"Rex\"}. [1]",
	}
	for name, text := range probes {
		t.Run(name, func(t *testing.T) {
			if r, res := CheckEntailmentGate(text); !r.Passed {
				t.Fatalf("name control: %s", res.Message)
			}
			for _, term := range []string{"a guard", "An officer"} {
				probe := strings.ReplaceAll(strings.ReplaceAll(text, "McRex", term), "Rex", term)
				if r, _ := CheckEntailmentGate(probe); r.Passed {
					t.Errorf("certified %s: %+v", probe, r.Steps)
				}
			}
		})
	}
}

func TestV465ParserBoundaries(t *testing.T) {
	policy := allStrictRules()
	for _, term := range []string{"a guard", "An officer", " A\tguard ", "AN OFFICER"} {
		if !indefiniteSingularTerm(term) {
			t.Errorf("missed %q", term)
		}
	}
	for _, term := range []string{"a", "Ann", "Rex", "the guard", "some guard", "Rex is a guard"} {
		if indefiniteSingularTerm(term) {
			t.Errorf("overmatched %q", term)
		}
	}
	for _, pair := range [][2]string{
		{"Rex signed the pledge", "a guard signed the pledge"},
		{"Rex wrote Atlas", "Rex wrote an officer"},
		{"no guard other than Rex is awake", "no guard other than a guard is awake"},
	} {
		if _, err := policy.countParse(pair[0]); err != nil {
			t.Errorf("count control %q: %v", pair[0], err)
		}
		if _, err := policy.countParse(pair[1]); err == nil {
			t.Errorf("count constant %q accepted", pair[1])
		}
	}
	for _, pair := range [][2]string{
		{`atom{"awake", "Rex"}`, `atom{"awake", "a guard"}`},
		{`name{"Rex"} is red`, `name{"an officer"} is red`},
	} {
		if _, err := policy.parseFOL(pair[0]); err != nil {
			t.Errorf("FOL control: %v", err)
		}
		if _, err := policy.parseFOL(pair[1]); err == nil {
			t.Errorf("FOL constant %q accepted", pair[1])
		}
	}
	if _, err := policy.descriptionParse("A guard is red"); err == nil {
		t.Error("ambiguous indefinite description accepted")
	}
	if _, err := policy.descriptionParse("Rex is red"); err != nil {
		t.Errorf("description name control: %v", err)
	}
	if len(policy.nplaceCandidates("a guard gave Bob Atlas")) != 0 {
		t.Error("generic/existential subject accepted")
	}
	if len(policy.nplaceCandidates("Rex gave Bob Atlas")) == 0 {
		t.Error("n-place name control refused")
	}
	// In the closed double-object syntax, an indefinite object is an explicit
	// existential binder, never a matched constant. Preserve that mapping.
	fs, why := policy.legacyFOLReadings("Alice gave Bob a book")
	if why != "" || len(fs) != 1 || len(folNames(fs[0])) != 2 {
		t.Errorf("existential object changed: %v %s", fs, why)
	}
}

func TestV465StructuredRefusal(t *testing.T) {
	text := "1. Every guard is awake.\n2. An officer is a guard.\n3. Therefore, an officer is awake. [1,2]"
	for _, options := range [][]EntailmentGateOptions{nil, {notesOn}} {
		r, _ := CheckEntailmentGate(text, options...)
		if r.Passed || len(r.Steps) != 1 || r.Steps[0].Refusal == nil || r.Steps[0].Refusal.Rule != "term.indefinite_singular" {
			t.Errorf("missing structured rule: %+v", r.Steps)
		}
	}
}

func TestV465LegacyBareArticles(t *testing.T) {
	for _, sentence := range []string{
		"A tensor vanishes exactly when its components vanish in every coordinate system",
		"An object vanishes exactly when its components vanish in every coordinate system",
		"container K-17 requires a permit",
	} {
		if fs, why := legacyFOLReadings(sentence); len(fs) != 0 || why != indefiniteTermMessage {
			t.Errorf("%q: %d readings, %s", sentence, len(fs), why)
		}
	}
	for _, sentence := range []string{`atom{"awake", "a"}`, `every x is such that atom{"admires", x, "a"}`} {
		if fs, why := legacyFOLReadings(sentence); len(fs) != 1 || why != "" {
			t.Errorf("explicit constant %q: %d readings, %s", sentence, len(fs), why)
		}
	}
}

func TestV465LegacyArticleAudit(t *testing.T) {
	data, err := os.ReadFile("testdata/legacy_v24_sentences.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []v25LegacySentence
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		fs, why := legacyFOLReadings(row.Sentence)
		if why != "" || len(fs) != 1 {
			continue
		}
		var visit func(*folFormula)
		visit = func(f *folFormula) {
			for _, term := range f.terms {
				if term == "a" || term == "an" {
					rendered, _ := legacyFOLRender(f)
					t.Logf("article-shaped term: %q => %s", row.Sentence, rendered)
					// Copular Fitch a and explicitly written FOL constants are
					// legitimate. English relation splitting cannot invent them.
					if f.op == "atom" && !strings.HasPrefix(f.symbol, "is/") && !folExplicit(row.Sentence) {
						t.Errorf("bare article in legacy relation: %s", rendered)
					}
				}
			}
			for _, child := range f.children {
				visit(child)
			}
		}
		visit(fs[0])
	}
}
