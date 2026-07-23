package validation

import "testing"

func TestUniversalNegativeInstantiation(t *testing.T) {
	positives := []struct {
		name     string
		argument string
	}{
		{
			name: "geoff-003 first person canonical form",
			argument: `1. I am aware of something.
2. No physical thing is aware of something.
3. Therefore, I am not a physical thing. [1,2]`,
		},
		{
			name: "wide-scope every-not E form",
			argument: `1. I am aware of something.
2. Every physical thing is not aware of something.
3. Therefore, I am not physical thing. [1,2]`,
		},
		{
			name: "plural instance and bare generic E form",
			argument: `1. These minds are aware of something.
2. Physical things are not aware of something.
3. Therefore, these minds are not physical things. [1,2]`,
		},
	}
	for _, test := range positives {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if !result.Passed || !report.Passed {
				t.Fatalf("gate failed: result=%#v report=%#v", result, report)
			}
			if got := report.Steps[0].Form; got != "universal negative instantiation" {
				t.Fatalf("form = %q, want universal negative instantiation", got)
			}
		})
	}

	negatives := []struct {
		name     string
		argument string
	}{
		{
			name: "contraries are not negation",
			argument: `1. The mind is invisible.
2. No physical thing is visible.
3. Therefore, the mind is not a physical thing. [1,2]`,
		},
		{
			name: "wrong middle",
			argument: `1. The mind is conscious.
2. No physical thing is extended.
3. Therefore, the mind is not a physical thing. [1,2]`,
		},
		{
			name: "conclusion excludes instance predicate instead of restricted class",
			argument: `1. The mind is conscious.
2. No physical thing is conscious.
3. Therefore, the mind is not conscious. [1,2]`,
		},
	}
	for _, test := range negatives {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if result.Passed || report.Passed {
				t.Fatalf("invalid form unexpectedly passed: result=%#v report=%#v", result, report)
			}
		})
	}
}

func TestSchematicInstantiation(t *testing.T) {
	positives := []struct {
		name     string
		argument string
	}{
		{
			name: "geoff-059 family",
			argument: `1. If X is conceivable, then X is metaphysically possible.
2. Any spatial object existing without any other spatial object existing is conceivable.
3. Therefore, any spatial object existing without any other spatial object existing is metaphysically possible. [1,2]`,
		},
		{
			name: "geoff-080 X form",
			argument: `1. If I am aware of having X, then I have X.
2. I am aware of having free will.
3. Therefore, I have free will. [1,2]`,
		},
		{
			name: "same substitution verifies multiple occurrences",
			argument: `1. If X resembles X, then X is self-similar.
2. The curve resembles the curve.
3. Therefore, the curve is self-similar. [1,2]`,
		},
	}
	for _, test := range positives {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if !result.Passed || !report.Passed {
				t.Fatalf("gate failed: result=%#v report=%#v", result, report)
			}
			if got := report.Steps[0].Form; got != "schematic instantiation" {
				t.Fatalf("form = %q, want schematic instantiation", got)
			}
		})
	}

	negatives := []struct {
		name     string
		argument string
	}{
		{
			name: "substitution mismatch",
			argument: `1. If X is conceivable, then X is metaphysically possible.
2. A cube is conceivable.
3. Therefore, a sphere is metaphysically possible. [1,2]`,
		},
		{
			name: "X only in antecedent",
			argument: `1. If X is conceivable, then something is metaphysically possible.
2. A cube is conceivable.
3. Therefore, something is metaphysically possible. [1,2]`,
		},
		{
			name: "lowercase x is not a variable",
			argument: `1. If x is conceivable, then x is metaphysically possible.
2. A cube is conceivable.
3. Therefore, a cube is metaphysically possible. [1,2]`,
		},
		{
			name: "X embedded in a word is not a variable",
			argument: `1. If Xylophone music is audible, then Xylophone music is detectable.
2. Bell music is audible.
3. Therefore, bell music is detectable. [1,2]`,
		},
	}
	for _, test := range negatives {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if result.Passed || report.Passed {
				t.Fatalf("invalid form unexpectedly passed: result=%#v report=%#v", result, report)
			}
		})
	}
}

func TestCatalogV3SurfaceTolerances(t *testing.T) {
	t.Run("clause equality ignores commas after parsing", func(t *testing.T) {
		if !equalClauses("Necessarily, the object exists.", "Necessarily the object exists.") {
			t.Fatal("clause equality did not ignore commas")
		}
		conditionals := conditionalCandidates("If P, then, Q.")
		if len(conditionals) != 1 || !equalClauses(conditionals[0].left, "P") || !equalClauses(conditionals[0].right, "Q") {
			t.Fatalf("conditional splitting broke under comma tolerance: %#v", conditionals)
		}
	})

	t.Run("separated thing quantifiers normalize", func(t *testing.T) {
		universals := universalCandidates("Every thing a person imagines is an idea.")
		if len(universals) != 1 || universals[0] != (clausePair{left: "a person imagines", right: "an idea"}) {
			t.Fatalf("every thing normalization candidates = %#v", universals)
		}
		negatives := universalNegativeCandidates("No thing a person imagines is a physical object.")
		if len(negatives) != 1 || negatives[0] != (clausePair{left: "a person imagines", right: "a physical object"}) {
			t.Fatalf("no thing normalization candidates = %#v", negatives)
		}
	})

	t.Run("am and is are equivalent copulas in terms", func(t *testing.T) {
		if !equalTerms("am aware of something", "is aware of something") {
			t.Fatal("term equality did not tolerate am/is copula variance")
		}
	})
}
