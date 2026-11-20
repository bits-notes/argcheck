package validation

import (
	"os"
	"strings"
	"testing"
)

var v462PronounProbes = map[string]struct{ text, rule string }{
	"D2-4":                {"1. Alice owns the card.\n2. If it is valid, then the alarm sounds.\n3. Bob owns the code.\n4. It is valid.\n5. Therefore, the alarm sounds. [2,4]\n6. Therefore, it is the case both that Alice owns the card, and that Bob owns the code. [1,3]\n7. Therefore, it is the case both that the alarm sounds, and that it is the case both that Alice owns the card, and that Bob owns the code. [5,6]", "term.pronoun_cross_statement"},
	"D3-4":                {"1. John greets Bill.\n2. If he is happy, then the alarm sounds.\n3. Rex greets Tom.\n4. He is happy.\n5. Therefore, the alarm sounds. [2,4]\n6. Therefore, it is the case both that John greets Bill, and that Rex greets Tom. [1,3]\n7. Therefore, it is the case both that the alarm sounds, and that it is the case both that John greets Bill, and that Rex greets Tom. [5,6]", "term.pronoun_cross_statement"},
	"D6-5":                {"1. Mary does not love Bill.\n2. If Mary hates him, then the alarm sounds.\n3. Mary greets John.\n4. Mary hates him.\n5. Therefore, the alarm sounds. [2,4]\n6. Therefore, it is the case both that Mary does not love Bill, and that Mary greets John. [1,3]\n7. Therefore, it is the case both that the alarm sounds, and that it is the case both that Mary does not love Bill, and that Mary greets John. [5,6]", "term.pronoun_cross_statement"},
	"D11-2":               {"1. If John helps his wife, then the alarm sounds.\n2. John helps his wife.\n3. Therefore, the alarm sounds. [1,2]", "term.possessive_cross_statement"},
	"weather":             {"1. If it rains, then it is wet.\n2. It rains.\n3. Therefore, it is wet. [1,2]", "term.pronoun_cross_statement"},
	"consequent":          {"1. If the bell rings, then he runs.\n2. The bell rings.\n3. Therefore, he runs. [1,2]", "term.pronoun_cross_statement"},
	"unused schema value": {"1. If X is Y, then X is happy.\n2. The man is fond of his wife.\n3. Therefore, the man is happy. [1,2]", "term.possessive_cross_statement"},
	"schema literal":      {"1. If X is fond of his wife, then X is happy.\n2. The man is fond of his wife.\n3. Therefore, the man is happy. [1,2]", "term.possessive_cross_statement"},
	"schema value":        {"1. If Alice is X, then Bob is X.\n2. Alice is fond of his wife.\n3. Therefore, Bob is fond of his wife. [1,2]", "term.possessive_cross_statement"},
	"property comparison": {"1. Socrates is identical to himself.\n2. Therefore, Socrates has the property of being identical to himself. [1]", "term.pronoun_cross_statement"},
	"pattern matcher":     {"1. If he sings, then the bell rings.\n2. Therefore, if it is not the case that the bell rings, then it is not the case that he sings. [1]", "term.pronoun_cross_statement"},
	"universal predicate": {"1. Every cat is fond of him.\n2. Felix is a cat.\n3. Therefore, Felix is fond of him. [1,2]", "term.pronoun_cross_statement"},
	"modal pattern":       {"1. It is not the case that possibly, it is not the case that it rains.\n2. Therefore, necessarily, it rains. [1]", "term.pronoun_cross_statement"},
	"Fitch scope":         {"1. If he sings, then Alice sings.\n| 2. Suppose he sings.\n| 3. Therefore, Alice sings. [1,2]\n4. Therefore, if he sings, then Alice sings. [2-3]", "term.pronoun_cross_statement"},
	"self entailment":     {"1. They sing.\n2. Therefore, they sing. [1]", "term.pronoun_cross_statement"},
}

func TestV462PronounCrossStatement(t *testing.T) {
	for name, tc := range v462PronounProbes {
		t.Run(name, func(t *testing.T) {
			r, _ := CheckEntailmentGate(tc.text, notesOn)
			if r.Passed {
				t.Fatal("certified cross-statement pronoun")
			}
			found := false
			for _, step := range r.Steps {
				for _, note := range step.Notes {
					found = found || note.ID == tc.rule
				}
			}
			if !found {
				t.Fatalf("missing %s note: %+v", tc.rule, r.Steps)
			}
			if r, res := allStrictRules().without(tc.rule).CheckEntailmentGate(tc.text); !r.Passed {
				t.Fatalf("flag off did not restore acceptance: %s", res.Message)
			}
		})
	}
}

var v462PronounControls = map[string]string{
	"unused conjunct":        "1. Both Alice sings and he runs.\n2. Therefore, Alice sings. [1]",
	"new disjunct":           "1. Alice sings.\n2. Therefore, either Alice sings or she runs. [1]",
	"unused possessive":      "1. Both Alice sings and John helps his wife.\n2. Therefore, Alice sings. [1]",
	"local repeated pronoun": "1. Alice sings.\n2. Therefore, either Alice sings or he admires himself. [1]",
	// Existing v20 negative-abstraction and negative-concretion accepts:
	"concretion scaffold": "1. It is not the case that Socrates has the property of being wise.\n2. Therefore, it is not the case that Socrates is wise. [1]",
	"negation scaffold":   "1. It is not the case that Socrates is wise.\n2. Therefore, it is not the case that Socrates has the property of being wise. [1]",
	"whole tokens":        "1. If Heather visits the theater, then the alarm sounds.\n2. Heather visits the theater.\n3. Therefore, the alarm sounds. [1,2]",
}

func TestV462PronounLocalControls(t *testing.T) {
	for name, text := range v462PronounControls {
		t.Run(name, func(t *testing.T) {
			if r, res := CheckEntailmentGate(text, notesOn); !r.Passed {
				t.Fatal(res.Message)
			}
		})
	}
	// Optional local fixture output for the brief's binary verification.
	if dir := os.Getenv("ARGCHECK_V462_CONTROLS"); dir != "" {
		for name, text := range v462PronounControls {
			if err := os.WriteFile(dir+"/"+strings.ReplaceAll(name, " ", "-")+".txt", []byte(text+"\n"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestV462PronounTokens(t *testing.T) {
	for _, word := range []string{"he", "she", "it", "they", "him", "her", "them", "himself", "herself", "itself", "themselves", "his", "their", "its"} {
		text := "1. If Alice mentions " + strings.ToUpper(word) + ", then the bell rings.\n2. Alice mentions " + word + ".\n3. Therefore, the bell rings. [1,2]"
		if r, _ := CheckEntailmentGate(text); r.Passed {
			t.Errorf("accepted %s", word)
		}
		p := allStrictRules().without("term.pronoun_cross_statement").without("term.possessive_cross_statement")
		if r, res := p.CheckEntailmentGate(text); !r.Passed {
			t.Errorf("baseline %s: %s", word, res.Message)
		}
	}
	for _, word := range []string{"theirs", "itinerary", "shelter", "éhe", "heé", "he2", "he_name"} {
		if got, changed := isolateStatementPronouns(word, "Internal", 1, crossStatementPronoun); changed || got != word {
			t.Errorf("partial token: %s", word)
		}
	}
}

func TestV462PronounPolicyHits(t *testing.T) {
	for _, tc := range v462PronounProbes {
		p, fired := allStrictRules().recording()
		if r, _ := p.CheckEntailmentGate(tc.text); r.Passed || !fired.ids[tc.rule] {
			t.Fatalf("missing hit for %s: %v", tc.rule, fired.ids)
		}
	}
}

func TestV462PronounIsolation(t *testing.T) {
	for _, scaffold := range []string{"It is the case that, Alice sings", "It\tis not the case that Alice sings", "It's not the case that Alice sings", "It’s the case both that Alice sings, and that Bob dances"} {
		if got, changed := isolateStatementPronouns(scaffold, "Internal", 1, crossStatementPronoun); changed || got != scaffold {
			t.Errorf("changed scaffold %q to %q", scaffold, got)
		}
	}
	a, _ := isolateStatementPronouns("He sees him and HE likes him", "Internal", 1, crossStatementPronoun)
	b, _ := isolateStatementPronouns("He sees him and HE likes him", "Internal", 2, crossStatementPronoun)
	if a == b || !strings.Contains(a, "Internalbhe sees Internalbhim and Internalbhe likes Internalbhim") {
		t.Fatalf("isolation failed: %s / %s", a, b)
	}
	// Both lists include her. Each enabled flag independently refuses it.
	text := "1. Alice greets her.\n2. Therefore, Alice greets her. [1]"
	for _, id := range []string{"term.pronoun_cross_statement", "term.possessive_cross_statement"} {
		if r, _ := allStrictRules().without(id).CheckEntailmentGate(text); r.Passed {
			t.Errorf("other rule failed to refuse her with %s off", id)
		}
	}
}
