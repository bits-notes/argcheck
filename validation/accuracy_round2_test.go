package validation

import (
	"strings"
	"testing"
)

func TestAccuracyRound2Cases(t *testing.T) {
	cases := []struct{ name, argument, reason string }{
		{"c1_02", `1. London is in England.
2. Therefore, it is not the case that it is not the case that London is in England. [1]
3. Therefore, either it is not the case that it is not the case that London is in England or two plus two equals five and Paris is in France. [2]
4. Therefore, if it is not the case that London is in England, then two plus two equals five and Paris is in France. [3]
5. Therefore, two plus two equals five and Paris is in France unless London is in England. [4]
6. Therefore, two plus two equals five. [5]`, "mixes top-level unless with and/or without explicit grouping"},
		{"c2_03", `1. Every billionaire is mortal.
2. Alice won the lottery if and only if Alice is a billionaire.
3. Therefore, Alice won the lottery if and only if Alice is mortal. [1,2]`, "top-level copula"},
		{"c2_05", `1. Every billionaire is mortal.
2. Alice is mortal.
| 3. Suppose Alice is a billionaire.
| 4. Therefore, Alice is a billionaire. [3]
5. Therefore, if Alice is a billionaire, then Alice is a billionaire. [3-4]
| 6. Suppose Alice is a billionaire.
| 7. Therefore, Alice is a billionaire. [6]
8. Therefore, if Alice is a billionaire, then Alice is a billionaire. [6-7]
9. Therefore, Alice is a billionaire if and only if Alice is a billionaire. [5,8]
10. Therefore, Alice is a billionaire if and only if Alice is mortal. [1,9]
11. Therefore, Alice is a billionaire. [2,10]`, "top-level copula"},
		{"c3_06", `1. Mars is one of the planets.
2. One of the planets is not rocky.
3. Therefore, it is not the case that Mars is rocky. [1,2]`, "quantifier phrase"},
		{"c3_07", `1. Socrates is a philosopher.
2. Someone in Athens is not a philosopher.
3. Therefore, it is not the case that Socrates is someone in Athens. [1,2]`, "quantifier phrase"},
		{"c3_08", `1. Mars is one of the planets.
2. One of the planets is not rocky.
3. Mars is rocky.
| 4. Suppose it is not the case that the moon is made of green cheese.
| 5. Therefore, it is not the case that Mars is rocky. [1,2]
| 6. Therefore, contradiction. [3,5]
7. Therefore, it is not the case that it is not the case that the moon is made of green cheese. [4-6]
8. Therefore, the moon is made of green cheese. [7]`, "quantifier phrase"},
		{"c3_09", `1. Every billionaire is mortal.
2. Almost no person is a billionaire.
3. Therefore, almost no person is mortal. [1,2]`, "quantifier phrase"},
		{"c3_10", `1. Every poet is mortal.
2. Just one senator is a poet.
3. Therefore, just one senator is mortal. [1,2]`, "quantifier phrase"},
		{"c7_14", `1. All princes are male.
2. Diana is a princess.
3. Therefore, Diana is male. [1,2]`, "does not match"},
		{"c7_15", `1. All corpses are dead.
2. The Marine Corps is a corps.
3. Therefore, the Marine Corps is dead. [1,2]`, "does not match"},
		{"r02_almostno", `1. Almost no philosopher is wise.
2. If X is wise, then X is happy.
3. Therefore, almost no philosopher is happy. [1,2]`, "quantifier phrase"},
		{"r10_veryfew", `1. Very few philosophers are wise.
2. If X are wise, then X are happy.
3. Therefore, very few philosophers are happy. [1,2]`, "quantifier phrase"},
		{"s01_lois", `1. Lois is looking for Superman.
2. Superman is identical to Clark.
3. Therefore, Lois is looking for Clark. [1,2]`, "subject of a plain sentence"},
		{"w01_planx", `1. If Plan X fails, then Plan X is abandoned.
2. Plan B fails.
3. Therefore, Plan B is abandoned. [1,2]`, "proper name"},
		{"w02_malcolm", `1. If Malcolm X is a man, then Malcolm X is mortal.
2. Malcolm Y is a man.
3. Therefore, Malcolm Y is mortal. [1,2]`, "proper name"},
		{"w04_genX", `1. If Generation X is cynical, then Generation X votes less.
2. Generation Z is cynical.
3. Therefore, Generation Z votes less. [1,2]`, "proper name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, report, msg := checkRulesDocArgument(tc.argument)
			if ok {
				t.Errorf("want NOT CERTIFIED; %+v", report.Steps)
			}
			if !strings.Contains(msg, tc.reason) {
				t.Errorf("want reason %q; got %s", tc.reason, msg)
			}
		})
	}

}

func TestAccuracyRound2ValidControls(t *testing.T) {
	cases := map[string]string{
		// v46.0: Socrates is now evidenced (names allowlist); valid.
		"schema":                "1. If X is a man, then X is mortal.\n2. Socrates is a man.\n3. Therefore, Socrates is mortal. [1,2]",
		"modified bare generic": "1. Red things are not blue things.\n2. The balls are red things.\n3. Therefore, it is not the case that the balls are blue things. [1,2]",

		"indefinite UI": "1. Every poet is mortal.\n2. A senator is a poet.\n3. Therefore, a senator is mortal. [1,2]",
		"definite UI":   "1. Every poet is mortal.\n2. The senator is a poet.\n3. Therefore, the senator is mortal. [1,2]",

		"unless": "1. The alarm sounds unless the battery is dead.\n2. Therefore, if it is not the case that the battery is dead, then the alarm sounds. [1]",
		"ui":     "1. Every billionaire is mortal.\n2. Alice is a billionaire.\n3. Therefore, Alice is mortal. [1,2]",
	}
	for name, arg := range cases {
		t.Run(name, func(t *testing.T) {
			ok, _, msg := checkRulesDocArgument(arg)
			if !ok {
				t.Fatal(msg)
			}
		})
	}

	// Refused by v44.0 strictness; retain the formerly valid controls.
	refused := []struct{ name, argument, hint string }{
		// term.inflection_verbatim: intentionally refused under v44.0 strictness.
		{"most every rule", "1. Every hunter is alert.\n2. Most cats are hunters.\n3. Therefore, most cats are alert. [1,2]", "write the rule with singular 'every N'"},
		// slot.letter_in_name: intentionally refused under v44.0 strictness.
		{"literal name conditional", "1. If Plan X fails, then Plan X is abandoned.\n2. Plan X fails.\n3. Therefore, Plan X is abandoned. [1,2]", "use a variable away from names, or write the name out"},
		// slot.letter_in_name: intentionally refused under v44.0 strictness.
		{"name plus variable", "1. If X supports Plan Y, then X recommends Plan Y.\n2. Alice supports Plan Y.\n3. Therefore, Alice recommends Plan Y. [1,2]", "use a variable away from names, or write the name out"},
		// slot.value_singular: intentionally refused under v44.0 strictness.
		{"theory", "1. If the theory of X is true, then X is useful.\n2. The theory of gravity is true.\n3. Therefore, gravity is useful. [1,2]", "Name the thing"},
		// term.inflection_verbatim: intentionally refused under v44.0 strictness.
		{"princes", "1. All princes are male.\n2. Charles is a prince.\n3. Therefore, Charles is male. [1,2]", "write the rule with singular 'every N'"},
		// slot.value_singular: intentionally refused under v44.0 strictness.
		{"long binding", "1. If X are useful, then X are valuable.\n2. Hand-coded objectives alongside algorithmic training signals are useful.\n3. Therefore, hand-coded objectives alongside algorithmic training signals are valuable. [1,2]", "Name the thing"},
		// slot.value_singular: intentionally refused under v44.0 strictness.
		{"gerund", "1. If X is useful, then X is valuable.\n2. Asking questions is useful.\n3. Therefore, asking questions is valuable. [1,2]", "Name the thing"},
		// slot.value_singular: intentionally refused under v44.0 strictness.
		{"description", "1. If X is useful, then X is valuable.\n2. The putative counterexample to Clark's analysis is useful.\n3. Therefore, the putative counterexample to Clark's analysis is valuable. [1,2]", "Name the thing"},
		// slot.value_singular: intentionally refused under v44.0 strictness.
		{"most schema", "1. If X are wise, then X are happy.\n2. Most philosophers are wise.\n3. Therefore, most philosophers are happy. [1,2]", "Name the thing"},
	}
	for _, tc := range refused {
		t.Run(tc.name, func(t *testing.T) {
			ok, _, msg := checkRulesDocArgument(tc.argument)
			if ok || !strings.Contains(msg, tc.hint) {
				t.Fatalf("want NOT CERTIFIED with fix hint %q; certified=%v: %s", tc.hint, ok, msg)
			}
		})
	}
}

func TestAccuracyRound2IdentityContexts(t *testing.T) {
	for _, verb := range []string{"is looking for", "looks for", "seeks", "wants", "needs", "is searching for", "hopes for", "fears", "believes", "imagines", "worships", "painted", "met", "kicked", "is near"} {
		t.Run(verb, func(t *testing.T) {
			arg := "1. Lois " + verb + " Superman.\n2. Superman is identical to Clark.\n3. Therefore, Lois " + verb + " Clark. [1,2]"
			// v35.0: identity substitution is certified only in identity
			// sentences and clause subjects, so every object position is
			// refused, extensional verbs included.
			want := false
			ok, _, msg := checkRulesDocArgument(arg)
			if ok != want {
				t.Fatalf("VALID=%v, want %v: %s", ok, want, msg)
			}
		})
	}
}

func TestAccuracyRound2TermGuards(t *testing.T) {
	for _, term := range []string{"almost no philosopher", "very few philosophers", "just one senator", "only one senator", "few philosophers"} {
		t.Run(term, func(t *testing.T) {
			if schematicBindingAllowed(term) {
				t.Fatal("quantifier accepted as schematic term")
			}
			if instanceSubjectAllowed(term) {
				t.Fatal("quantifier accepted as UI subject")
			}
		})
	}
	// Upward-monotone existentials are safe in a schematic slot ("Someone is
	// red" + "If X is red, then X is warm" gives "Someone is warm") but are
	// never individual UI subjects.
	for _, term := range []string{"one of the planets", "someone in Athens", "something red", "somebody wise"} {
		t.Run(term, func(t *testing.T) {
			if !schematicBindingAllowed(term) {
				t.Fatal("upward existential refused as schematic term")
			}
			if instanceSubjectAllowed(term) {
				t.Fatal("quantifier accepted as UI subject")
			}
		})
	}
	for _, word := range []string{"princess", "corps", "class", "bus", "glass", "chaos", "lens", "series", "species"} {
		t.Run(word, func(t *testing.T) {
			if got := round2Singular(word); got != word {
				t.Fatalf("singular changed to %q", got)
			}
		})
	}
	if instanceSubjectAllowed("Most philosophers") {
		t.Fatal("Most accepted as UI name")
	}
}

func TestAccuracyRound2WholeInstance(t *testing.T) {
	for _, arg := range []string{
		"1. Every billionaire is mortal.\n2. If Bob sings, then Alice is a billionaire.\n3. Therefore, if Bob sings, then Alice is mortal. [1,2]",
		"1. Every billionaire is mortal.\n2. Bob sings and Alice is a billionaire.\n3. Therefore, Bob sings and Alice is mortal. [1,2]",
		"1. Every billionaire is mortal.\n2. It is not the case that Alice is a billionaire.\n3. Therefore, it is not the case that Alice is mortal. [1,2]",
	} {
		ok, _, msg := checkRulesDocArgument(arg)
		if ok {
			t.Errorf("embedded UI accepted: %s (%s)", arg, msg)
		}
	}
}
