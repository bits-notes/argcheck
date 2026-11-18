package validation

import "testing"

type v11Case struct {
	name, argument string
	form           string // expected form when it should pass; "" means it must fail
}

func runV11Cases(t *testing.T, cases []v11Case) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			want := tc.form != ""
			if report.Passed != want || result.Passed != want {
				t.Fatalf("passed = %v, want %v: %s %+v", report.Passed, want, result.Message, report.Steps)
			}
			if want {
				if got := report.Steps[len(report.Steps)-1].Form; got != tc.form {
					t.Fatalf("form = %q, want %q", got, tc.form)
				}
			}
		})
	}
}

// Two-placeholder schematic rules were already certified by v4's schematic
// instantiation (X, Y and Z bind under one consistent substitution). v11 pins
// the predicate-logic writer's usage. X and Y MAY bind the same phrase: the
// rule reads as for-all x for-all y, and instantiating both with one term is
// valid, exactly as the one-placeholder logic allows any phrase for X.
func TestCatalogV11TwoPlaceholderSchematicInstantiation(t *testing.T) {
	runV11Cases(t, []v11Case{
		{"two placeholders", "1. If X is a parent of Y, then Y is a child of X.\n2. Dana is a parent of Pat.\n3. Therefore, Pat is a child of Dana. [1,2]", "schematic instantiation"},
		{"two placeholders repeated", "1. If X trusts Y and Y trusts X, then X and Y are allies.\n2. Ada trusts Ben.\n3. Ben trusts Ada.\n4. Therefore, Ada and Ben are allies. [1,2,3]", "schematic conjunctive modus ponens"},
		{"two placeholders multiword phrases", "1. If X outranks Y, then Y reports to X.\n2. The chief engineer outranks the night supervisor.\n3. Therefore, the night supervisor reports to the chief engineer. [1,2]", "schematic instantiation"},
		{"identical phrases allowed", "1. If X admires Y, then Y trusts X.\n2. Alice admires Alice.\n3. Therefore, Alice trusts Alice. [1,2]", "schematic instantiation"},
		{"inconsistent substitution", "1. If X is a parent of Y, then Y is a child of X.\n2. Dana is a parent of Pat.\n3. Therefore, Pat is a child of Lee. [1,2]", ""},
		{"swapped in consequent", "1. If X is a parent of Y, then Y is a child of X.\n2. Dana is a parent of Pat.\n3. Therefore, Dana is a child of Pat. [1,2]", ""},
		{"swapped across conjuncts", "1. If X trusts Y and Y fears X, then X and Y are rivals.\n2. Ada trusts Ben.\n3. Ada fears Ben.\n4. Therefore, Ada and Ben are rivals. [1,2,3]", ""},
		{"partial match", "1. If X is a parent of Y, then Y is a child of X.\n2. Dana is a guardian of Pat.\n3. Therefore, Pat is a child of Dana. [1,2]", ""},
		{"consequent placeholder unbound", "1. If X is a parent of Pat, then Y is a child of X.\n2. Dana is a parent of Pat.\n3. Therefore, Pat is a child of Dana. [1,2]", ""},
	})
}

// Categorical existential introduction (a is S; a is P; therefore some S is
// P / some Ss are Ps) was already certified by v5's existential introduction.
func TestCatalogV11CategoricalExistentialIntroduction(t *testing.T) {
	runV11Cases(t, []v11Case{
		{"singular", "1. Rex is a dog.\n2. Rex is friendly.\n3. Therefore, some dog is friendly. [1,2]", "existential introduction"},
		{"plural", "1. Rex is a dog.\n2. Rex is friendly.\n3. Therefore, some dogs are friendly. [1,2]", "existential introduction"},
		{"different witnesses", "1. Rex is a dog.\n2. Tom is friendly.\n3. Therefore, some dog is friendly. [1,2]", ""},
		{"predicate drift", "1. Rex is a dog.\n2. Rex is friendly.\n3. Therefore, some dog is kind. [1,2]", ""},
	})
}

func TestCatalogV11SchematicExistentialIntroduction(t *testing.T) {
	runV11Cases(t, []v11Case{
		{"simple", "1. Ada solved the puzzle.\n2. Therefore, there is an X such that X solved the puzzle. [1]", "schematic existential introduction"},
		{"repeated term", "1. Ada admires Ada's teacher.\n2. Therefore, there is an X such that X admires X's teacher. [1]", "schematic existential introduction"},
		{"object position", "1. Every critic praised Ada.\n2. Therefore, there is an X such that every critic praised X. [1]", "schematic existential introduction"},
		{"exists a Y", "1. The backup sensor is receiving power.\n2. Therefore, there exists a Y such that Y is receiving power. [1]", "schematic existential introduction"},
		{"negated matrix", "1. It is not the case that Ada sings.\n2. Therefore, there is an X such that it is not the case that X sings. [1]", "schematic existential introduction"},
		{"premise with placeholder unsupported", "1. Dana is a parent of Pat.\n2. Therefore, there is an X such that X is a parent of Pat. [1]\n3. Therefore, there is a Y such that there is an X such that X is a parent of Y. [2]", ""},
		{"instance mismatch", "1. Ada solved the puzzle.\n2. Therefore, there is an X such that X solved the riddle. [1]", ""},
		{"partial replacement", "1. Ada admires Ada's teacher.\n2. Therefore, there is an X such that X admires Ada's teacher. [1]", ""},
		{"no placeholder in matrix", "1. Ada solved the puzzle.\n2. Therefore, there is an X such that Ada solved the puzzle. [1]", ""},
		{"wrong bound placeholder", "1. Ada solved the puzzle.\n2. Therefore, there is an X such that Y solved the puzzle. [1]", ""},
		{"quantifier phrase is not a term", "1. Nothing is red.\n2. Therefore, there is an X such that X is red. [1]", ""},
		{"pronoun is not a term", "1. He solved the puzzle.\n2. Therefore, there is an X such that X solved the puzzle. [1]", ""},
		{"bound definite", "1. Each student's advisor praised the student.\n2. Therefore, there is an X such that each student's advisor praised X. [1]", ""},
		{"attitude context", "1. Lois believes Superman flies.\n2. Therefore, there is an X such that Lois believes X flies. [1]", ""},
		{"existence predicate", "1. Pegasus does not exist.\n2. Therefore, there is an X such that X does not exist. [1]", ""},
		{"modal context", "1. Ada might win.\n2. Therefore, there is an X such that X might win. [1]", ""},
		{"part of longer name", "1. Paris Hilton sings.\n2. Therefore, there is an X such that X Hilton sings. [1]", ""},
		{"quoted", "1. Ada wrote \"Ada\".\n2. Therefore, there is an X such that X wrote \"X\". [1]", ""},
		{"bare placeholder matrix", "1. Ada sings.\n2. Therefore, there is an X such that X. [1]", ""},
		{"term spanning a verb", "1. Ada solved the puzzle.\n2. Therefore, there is an X such that X solved X. [1]", ""},
		{"coordinated term", "1. Ada and Ben met.\n2. Therefore, there is an X such that X met. [1]", ""},
		{"universal does not license", "1. Every dog barks.\n2. Therefore, there is an X such that X barks. [1]", ""},
	})
}

func TestCatalogV11ExistentialSyllogisms(t *testing.T) {
	runV11Cases(t, []v11Case{
		{"darii singular", "1. Some cat is a hunter.\n2. Every hunter is alert.\n3. Therefore, some cat is alert. [1,2]", "existential syllogism"},
		{"darii plural", "1. Some cats are hunters.\n2. All hunters are alert.\n3. Therefore, some cats are alert. [1,2]", "existential syllogism"},
		{"darii premise order", "1. Every hunter is alert.\n2. Some cat is a hunter.\n3. Therefore, some cat is alert. [1,2]", "existential syllogism"},
		{"ferio", "1. Some cat is a hunter.\n2. No hunter is lazy.\n3. Therefore, some cat is not lazy. [1,2]", "existential negative syllogism"},
		{"ferio plural", "1. Some cats are hunters.\n2. No hunters are lazy.\n3. Therefore, some cats are not lazy. [1,2]", "existential negative syllogism"},
		{"festino converted E", "1. Some cat is a hunter.\n2. No lazy thing is a hunter.\n3. Therefore, some cat is not a lazy thing. [1,2]", "existential negative syllogism"},
		{"I + I", "1. Some cat is a hunter.\n2. Some hunter is alert.\n3. Therefore, some cat is alert. [1,2]", ""},
		{"undistributed middle", "1. Every hunter is alert.\n2. Some cat is alert.\n3. Therefore, some cat is a hunter. [1,2]", ""},
		{"middle as universal predicate", "1. Some cat is a hunter.\n2. Every alert thing is a hunter.\n3. Therefore, some cat is alert. [1,2]", ""},
		{"O premise is not I", "1. Some cat is not a hunter.\n2. Every hunter is alert.\n3. Therefore, some cat is not alert. [1,2]", ""},
		{"ferio affirmative conclusion", "1. Some cat is a hunter.\n2. No hunter is lazy.\n3. Therefore, some cat is lazy. [1,2]", ""},
		{"ferio universal conclusion", "1. Some cat is a hunter.\n2. No hunter is lazy.\n3. Therefore, no cat is lazy. [1,2]", ""},
		{"ferio O conversion", "1. Some cat is a hunter.\n2. No hunter is lazy.\n3. Therefore, some lazy thing is not a cat. [1,2]", ""},
		{"darii to universal", "1. Some cat is a hunter.\n2. Every hunter is alert.\n3. Therefore, every cat is alert. [1,2]", ""},
		{"term drift", "1. Some cat is a hunter.\n2. Every hunter is alert.\n3. Therefore, some dog is alert. [1,2]", ""},
	})
}

// Identity substitution uses only the explicit "is identical to" relation:
// bare "a is b" cannot be told apart from predication ("Ada is a doctor").
func TestCatalogV11IdentitySubstitution(t *testing.T) {
	runV11Cases(t, []v11Case{
		{"forward", "1. Hesperus is identical to Phosphorus.\n2. Hesperus is visible at dusk.\n3. Therefore, Phosphorus is visible at dusk. [1,2]", "identity substitution"},
		{"symmetric", "1. Hesperus is identical to Phosphorus.\n2. Phosphorus is visible at dawn.\n3. Therefore, Hesperus is visible at dawn. [1,2]", "identity substitution"},
		// v35.0: identity substitution is certified only in identity sentences
		// and clause subjects, so these object positions are refused.
		{"every occurrence", "1. Ada is identical to the countess.\n2. Ada praised Ada's tutor.\n3. Therefore, the countess praised the countess's tutor. [1,2]", ""},
		{"object under quantifier", "1. Ada is identical to Lovelace.\n2. Every critic praised Ada.\n3. Therefore, every critic praised Lovelace. [1,2]", ""},
		{"subject substitution", "1. Ada is identical to Lovelace.\n2. Ada praised every critic.\n3. Therefore, Lovelace praised every critic. [1,2]", "identity substitution"},
		{"partial replacement", "1. Ada is identical to the countess.\n2. Ada praised Ada's tutor.\n3. Therefore, the countess praised Ada's tutor. [1,2]", ""},
		{"wrong term", "1. Hesperus is identical to Phosphorus.\n2. Hesperus is visible at dusk.\n3. Therefore, Mars is visible at dusk. [1,2]", ""},
		{"predication is not identity", "1. Ada is a doctor.\n2. Ada praised Ben.\n3. Therefore, a doctor praised Ben. [1,2]", ""},
		{"bare is not identity", "1. Hesperus is Phosphorus.\n2. Hesperus is visible at dusk.\n3. Therefore, Phosphorus is visible at dusk. [1,2]", ""},
		{"attitude context", "1. Superman is identical to Clark Kent.\n2. Lois believes Superman flies.\n3. Therefore, Lois believes Clark Kent flies. [1,2]", ""},
		{"knowledge context", "1. Hesperus is identical to Phosphorus.\n2. The ancients knew that Hesperus is Hesperus.\n3. Therefore, the ancients knew that Hesperus is Phosphorus. [1,2]", ""},
		{"term absent", "1. Hesperus is identical to Phosphorus.\n2. Mars is red.\n3. Therefore, Mars is red. [1,2]", ""},
		{"part of longer name", "1. Paris is identical to the capital.\n2. Paris Hilton sings.\n3. Therefore, the capital Hilton sings. [1,2]", ""},
		{"indefinite is not a term", "1. Ada is identical to a doctor.\n2. Ada praised Ben.\n3. Therefore, a doctor praised Ben. [1,2]", ""},
		{"non-identity", "1. It is not the case that Hesperus is identical to Mars.\n2. Hesperus is visible at dusk.\n3. Therefore, Mars is visible at dusk. [1,2]", ""},
	})
}

// Soundness fix: v10 universal (negative) instantiation read quantifier
// phrases such as "no cat" as individuals, certifying "No cat is a hunter;
// every hunter is alert; therefore no cat is alert". Downward/non-monotone
// subjects now fail; "some" subjects route to the existential syllogisms.
func TestCatalogV11QuantifiedSubjectsAreNotInstances(t *testing.T) {
	runV11Cases(t, []v11Case{
		{"no subject", "1. No cat is a hunter.\n2. Every hunter is alert.\n3. Therefore, no cat is alert. [1,2]", ""},
		{"nothing subject", "1. Nothing is a hunter.\n2. Every hunter is alert.\n3. Therefore, nothing is alert. [1,2]", ""},
		{"few subject", "1. Few cats are hunters.\n2. All hunters are alert.\n3. Therefore, few cats are alert. [1,2]", ""},
		{"at most subject", "1. At most one cat is a hunter.\n2. Every hunter is alert.\n3. Therefore, at most one cat is alert. [1,2]", ""},
		{"only subject", "1. Only cats are hunters.\n2. All hunters are alert.\n3. Therefore, only cats are alert. [1,2]", ""},
		{"no subject negative instantiation", "1. No cat is a hunter.\n2. No lazy thing is a hunter.\n3. Therefore, no cat is not a lazy thing. [1,2]", ""},
		{"individual unchanged", "1. Every hunter is alert.\n2. Tom is a hunter.\n3. Therefore, Tom is alert. [1,2]", "universal instantiation"},
		{"most subject lifted schematically", "1. Every hunter is alert.\n2. Most cats are hunters.\n3. Therefore, most cats are alert. [1,2]", "schematic instantiation"},
	})
}
