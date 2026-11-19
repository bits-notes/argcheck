package validation

import (
	"fmt"
	"testing"
)

// Included in the central gap ledger so accepted and refused boundaries remain
// visible there, rather than only in this version's regression tests.
var catalogV16GapCases = []struct{ name, argument, wantForm string }{
	{"v16 compound biconditional elimination", `1. (If it rains, then the street is wet) if and only if (The city is prepared).
2. Therefore, If If it rains, then the street is wet, then The city is prepared. [1]`, "biconditional elimination"},
	{"v16 embedded copula existential introduction", `1. The claim that the store is nearby is false.
2. The claim that the store is nearby is a claim.
3. Therefore, Some claim is false. [1,2]`, "existential introduction"},
	{"v16 embedded copula universal instantiation", `1. Every false claim is retracted.
2. The claim that the store is nearby is a false claim.
3. Therefore, The claim that the store is nearby is retracted. [1,2]`, "universal instantiation"},
	{"v16 embedded copula universal negative instantiation", `1. No false claim is retracted.
2. The claim that the store is nearby is a false claim.
3. Therefore, It is not the case that The claim that the store is nearby is retracted. [1,2]`, "universal negative instantiation"},
	{"v16 embedded copula old negative spelling", `1. The claim that the store is nearby is a false claim.
2. No retracted item is a false claim.
3. Therefore, The claim that the store is nearby is not a retracted item. [1,2]`, "universal negative instantiation"},
	{"v16 embedded copula existential generalization", `1. The claim that the store is nearby is false.
2. Therefore, Something is false. [1]`, "existential generalization"},
	{"v16 whole subject schematic existential", `1. The claim that the store is nearby is false.
2. Therefore, There is an X such that X is false. [1]`, "schematic existential introduction"},
	{"v16 embedded copula existential syllogism", `1. Some claim that the store is nearby is a false thing.
2. Every false thing is retracted.
3. Therefore, Some claim that the store is nearby is retracted. [1,2]`, "existential syllogism"},
	{"v16 embedded copula existential negative syllogism", `1. Some claim that the store is nearby is a false thing.
2. No false thing is retracted.
3. Therefore, Some claim that the store is nearby is not retracted. [1,2]`, "existential negative syllogism"},
	{"v16 whole noun phrase substitution", `1. The claim that the store is nearby is identical to Ngo's claim.
2. Ngo's claim is false.
3. Therefore, The claim that the store is nearby is false. [1,2]`, "identity substitution"},
	{"v16 whole noun phrase reverse substitution", `1. The claim that the store is nearby is identical to Ngo's claim.
2. The claim that the store is nearby is false.
3. Therefore, Ngo's claim is false. [1,2]`, "identity substitution"},
	{"v16 whole object substitution", `1. The claim that the store is nearby is identical to Ngo's claim.
2. Ada repeats The claim that the store is nearby.
3. Therefore, Ada repeats Ngo's claim. [1,2]`, ""}, // v35.0: object position
	{"v16 whole object reverse substitution", `1. The claim that the store is nearby is identical to Ngo's claim.
2. Ada repeats Ngo's claim.
3. Therefore, Ada repeats The claim that the store is nearby. [1,2]`, ""}, // v35.0: object position
	{"v16 mismatched witnesses", `1. The claim that the store is nearby is false.
2. The claim that the station is nearby is a claim.
3. Therefore, Some claim is false. [1,2]`, ""},
	{"v16 truncated witness", `1. The claim that the store is nearby is false.
2. The claim that the store is a claim.
3. Therefore, Some claim is false. [1,2]`, ""},
	{"v16 invented embedded witness", `1. Ada is convinced that Hesperus is bright.
2. Hesperus is a planet.
3. Therefore, Some planet is bright. [1,2]`, ""},
	{"v16 embedded copula drops subject", `1. Every claim is false.
2. The claim that the store is nearby is a claim.
3. Therefore, The store is false. [1,2]`, ""},
	{"v16 negative instantiation drops subject", `1. No false claim is retracted.
2. The claim that the store is nearby is a false claim.
3. Therefore, It is not the case that the store is retracted. [1,2]`, ""},
	{"v16 no existential from embedded belief", `1. Ann believes that Hesperus is bright.
2. Therefore, Something is bright. [1]`, ""},
	{"v16 no existential from partial that subject", `1. The claim that the store is nearby is false.
2. Therefore, Something is nearby is false. [1]`, ""},
	{"v16 no schematic existential inside noun phrase", `1. The claim that the store is nearby is false.
2. Therefore, There is an X such that the claim that X is nearby is false. [1]`, ""},
	{"v16 belief stays opaque", `1. Hesperus is identical to Phosphorus.
2. Ann believes that Hesperus is bright.
3. Therefore, Ann believes that Phosphorus is bright. [1,2]`, ""},
	{"v16 substitution inside noun phrase stays opaque", `1. Hesperus is identical to Phosphorus.
2. The claim that Hesperus is bright is false.
3. Therefore, The claim that Phosphorus is bright is false. [1,2]`, ""},
	{"v16 whole term within belief stays opaque", `1. The claim that the store is nearby is identical to Ngo's claim.
2. Ann believes that The claim that the store is nearby is false.
3. Therefore, Ann believes that Ngo's claim is false. [1,2]`, ""},
	{"v16 whole term object within belief stays opaque", `1. The claim that the store is nearby is identical to Ngo's claim.
2. Ann believes The claim that the store is nearby.
3. Therefore, Ann believes Ngo's claim. [1,2]`, ""},
	{"v16 whole term possessive not entire object", `1. The claim that the store is nearby is identical to Ngo's claim.
2. Ada repeats The claim that the store is nearby's wording.
3. Therefore, Ada repeats Ngo's claim's wording. [1,2]`, ""},
	{"v16 lexical drift", `1. (If it rains, then the street is wet) if and only if (The city is prepared).
2. The city is prepared.
3. Therefore, If it rains, then the road is wet. [1,2]`, ""},
	{"v16 unscaffolded conditional remains outer", `1. If it rains, then the street is wet if and only if the city is prepared.
2. it rains.
3. Therefore, the street is wet if and only if the city is prepared. [1,2]`, "modus ponens"},
	{"v16 unscaffolded left conditional not outer iff", `1. If it rains, then the street is wet if and only if the city is prepared.
2. If it rains, then the street is wet.
3. Therefore, The city is prepared. [1,2]`, ""},
	{"v16 should stays opaque", `1. The claim that the store is nearby is identical to Ngo's claim.
2. The claim that the store is nearby should be retracted.
3. Therefore, Ngo's claim should be retracted. [1,2]`, ""},
	{"v16 must stays opaque", `1. The claim that the store is nearby is identical to Ngo's claim.
2. The claim that the store is nearby must be retracted.
3. Therefore, Ngo's claim must be retracted. [1,2]`, ""},
	{"v16 can stays opaque", `1. The claim that the store is nearby is identical to Ngo's claim.
2. The claim that the store is nearby can be retracted.
3. Therefore, Ngo's claim can be retracted. [1,2]`, ""},
	{"v16 Necessarily,  no witness", `1. Necessarily, The claim that the store is nearby is false.
2. Therefore, Something is false. [1]`, ""},
	{"v16 Possibly,  no witness", `1. Possibly, The claim that the store is nearby is false.
2. Therefore, Something is false. [1]`, ""},
	{"v16 It is not the case that  no witness", `1. It is not the case that The claim that the store is nearby is false.
2. Therefore, Something is false. [1]`, ""},
}

func TestCatalogV16BiconditionalCompoundSides(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/1/7/intro": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/1/7/elim_reverse": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/2/7/intro": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/2/7/elim_reverse": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/7/1/intro": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/7/1/elim": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/7/2/intro": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV16BiconditionalCompoundSides/7/2/elim": {"scope.clause_start", "move the operator into each clause"},
	}

	iff := func(p, q string) string { return "(" + p + ") if and only if (" + q + ")" }
	neg := func(p string) string { return "It is not the case that " + p }
	cond := func(p, q string) string { return "If " + p + ", then " + q }
	compounds := []string{
		"If it rains, then the street is wet",
		"It is the case both that it rains, and that it snows",
		"Either it rains or it snows",
		"It is not the case that it rains",
		"Necessarily, it rains",
		"Possibly, it rains",
		"It rains if and only if it snows",
		iff("If it rains, then it snows", "The street is wet"),
		"Every cat is mortal",
	}
	for i, p := range compounds {
		for j, q := range compounds {
			r := "The city is prepared"
			for _, tc := range []struct {
				name             string
				premises         []string
				conclusion, form string
			}{
				{"intro", []string{cond(p, q), cond(q, p)}, iff(p, q), "biconditional introduction"},
				{"elim", []string{iff(p, q)}, cond(p, q), "biconditional elimination"},
				{"elim reverse", []string{iff(p, q)}, cond(q, p), "biconditional elimination"},
				{"MP", []string{iff(p, q), p}, q, "biconditional modus ponens"},
				{"MP reverse", []string{iff(p, q), q}, p, "biconditional modus ponens"},
				{"MT", []string{iff(p, q), neg(q)}, neg(p), "biconditional modus tollens"},
				{"MT reverse", []string{iff(p, q), neg(p)}, neg(q), "biconditional modus tollens"},
				{"chain", []string{iff(p, q), iff(q, r)}, iff(p, r), "biconditional chaining"},
				{"commute", []string{iff(p, q)}, iff(q, p), "biconditional commutation"},
			} {
				t.Run(fmt.Sprintf("%d/%d/%s", i, j, tc.name), func(t *testing.T) {
					argument := ""
					refs := ""
					for k, premise := range tc.premises {
						argument += fmt.Sprintf("%d. %s.\n", k+1, premise)
						if k > 0 {
							refs += ","
						}
						refs += fmt.Sprint(k + 1)
					}
					argument += fmt.Sprintf("%d. Therefore, %s. [%s]", len(tc.premises)+1, tc.conclusion, refs)
					if listedStrictRefusal(t, argument, refused) {
						return
					}
					report, result := CheckEntailmentGate(argument)
					if !report.Passed || !result.Passed {
						t.Fatalf("%s\n%s", result.Message, argument)
					}
					// Identical sides can trigger an earlier valid form (self-entailment).
					if i != j && report.Steps[0].Form != tc.form {
						t.Fatalf("got %q, want %q", report.Steps[0].Form, tc.form)
					}
				})
			}
		}
	}
}

func TestCatalogV16MalformedScaffold(t *testing.T) {
	for _, s := range []string{
		"() if and only if (it rains)",
		"(it rains) if and only if ()",
		"(it rains if and only if (it snows)",
		"(it rains) if and only if (it snows))",
		"(it rains) if and only if (it snows) and it hails",
		"It is the case that () if and only if (it rains)",
	} {
		if got := biconditionalCandidates(s); len(got) != 0 {
			t.Errorf("malformed scaffold parsed: %s: %+v", s, got)
		}
	}
}
