package validation

import "testing"

func TestCatalogV4SchematicInstantiation(t *testing.T) {
	// v44 by-design refusals: retained separately from the accepted controls.
	refused := map[string]strictRefusal{
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestCatalogV4SchematicInstantiation/single_variable_repeated_throughout": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestCatalogV4SchematicInstantiation/schematic_conjunctive_modus_ponens_with_one_variable": {"slot.value_singular", "Name the thing"},
		// slot.value_singular: term is outside the evidenced-name/singular-description grammar.
		"TestCatalogV4SchematicInstantiation/schematic_conjunctive_modus_ponens_with_three_variables": {"slot.value_singular", "Name the thing"},
	}

	positives := []struct {
		name     string
		argument string
		form     string
	}{
		{
			name: "single variable possessive consequent",
			argument: `1. If X is receiving power, then X's green indicator is illuminated.
2. The backup sensor is receiving power.
3. Therefore, the backup sensor's green indicator is illuminated. [1,2]`,
			form: "schematic instantiation",
		},
		{
			name: "single variable repeated throughout",
			argument: `1. If X's computed checksum exactly matches X's publisher's signed checksum, then X is unaltered.
2. This archive's computed checksum exactly matches this archive's publisher's signed checksum.
3. Therefore, this archive is unaltered. [1,2]`,
			form: "schematic instantiation",
		},
		{
			name: "multiple variables in two-premise instantiation",
			argument: `1. If X admires Y, then Y trusts X.
2. Alice admires Bob.
3. Therefore, Bob trusts Alice. [1,2]`,
			form: "schematic instantiation",
		},
		{
			name: "schematic conjunctive modus ponens with one variable",
			argument: `1. If X is a sealed container and X carries more than two thousand kilograms of Class 8 corrosive material, then X requires a red handling permit before X may enter the transfer yard.
2. Container K-17 is a sealed container.
3. Container K-17 carries more than two thousand kilograms of Class 8 corrosive material.
4. Therefore, container K-17 requires a red handling permit before container K-17 may enter the transfer yard. [1,2,3]`,
			form: "schematic conjunctive modus ponens",
		},
		{
			name: "schematic conjunctive modus ponens with three variables",
			argument: `1. If X is Y's mother and Z is X's brother, then Z is Y's uncle.
2. Dana is Pat's mother.
3. Lee is Dana's brother.
4. Therefore, Lee is Pat's uncle. [1,2,3]`,
			form: "schematic conjunctive modus ponens",
		},
	}

	for _, test := range positives {
		t.Run(test.name, func(t *testing.T) {
			if listedStrictRefusal(t, test.argument, refused) {
				return
			}
			report, result := CheckEntailmentGate(test.argument)
			if !result.Passed || !report.Passed {
				t.Fatalf("gate failed: result=%#v report=%#v", result, report)
			}
			if got := report.Steps[0].Form; got != test.form {
				t.Fatalf("form = %q, want %q", got, test.form)
			}
		})
	}
}

func TestCatalogV4SchematicInstantiationRejectsInvalidForms(t *testing.T) {
	negatives := []struct {
		name     string
		argument string
	}{
		{
			name: "inconsistent substitution across conjuncts",
			argument: `1. If X is Y's mother and Z is X's brother, then Z is Y's uncle.
2. Dana is Pat's mother.
3. Lee is Pat's brother.
4. Therefore, Lee is Pat's uncle. [1,2,3]`,
		},
		{
			name: "consequent variable absent from antecedent",
			argument: `1. If X is wet, then Y is slippery.
2. The deck is wet.
3. Therefore, the deck is slippery. [1,2]`,
		},
		{
			name: "instantiated consequent wording drift",
			argument: `1. If X is receiving power, then X's green indicator is illuminated.
2. The backup sensor is receiving power.
3. Therefore, the backup sensor's light is on. [1,2]`,
		},
		{
			name: "unused cited premise",
			argument: `1. If X is a sealed container and X carries corrosive material, then X requires a permit.
2. Container K-17 is a sealed container.
3. Container K-17 carries corrosive material.
4. Container K-17 is inspected.
5. Therefore, container K-17 requires a permit. [1,2,3,4]`,
		},
		{
			name: "empty substitution",
			argument: `1. If (X) is marked, then (X) is tracked.
2. () is marked.
3. Therefore, () is tracked. [1,2]`,
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

func TestCatalogV4PlainConjunctiveModusPonensIsUnchanged(t *testing.T) {
	argument := `1. If container K-17 is sealed and container K-17 carries corrosive material, then container K-17 requires a permit.
2. Container K-17 is sealed.
3. Container K-17 carries corrosive material.
4. Therefore, container K-17 requires a permit. [1,2,3]`

	// scope.clause_start: the strict grammar does not admit every bare clause opener.
	refused := []string{argument}
	for _, text := range refused {
		assertStrictRefusal(t, text, "scope.clause_start", "move the operator into each clause")
	}
}

func TestCatalogV4SchematicVariableBoundaries(t *testing.T) {
	template, variables := schematicTemplate("X's token, (Y), XY, Xenon, and Z")
	if variables['X'] != 1 || variables['Y'] != 1 || variables['Z'] != 1 || len(variables) != 3 {
		t.Fatalf("schematic variable counts = %#v, want one standalone X, Y, and Z", variables)
	}
	// This unit test isolates variable boundaries; provide the name evidence
	// that an argument-level caller normally collects before instantiation.
	policy := allStrictRules()
	names := map[string]bool{"alex": true, "beta": true, "gamma": true}
	policy.slotNames = &names
	if _, ok := policy.instantiateSchematicTemplate(template, schematicSubstitution{
		'X': "alex",
		'Y': "beta",
		'Z': "gamma",
	}); !ok {
		t.Fatal("standalone schematic variables did not instantiate")
	}
}
