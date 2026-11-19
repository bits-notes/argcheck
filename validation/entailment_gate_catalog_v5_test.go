package validation

import "testing"

func TestCatalogV5SafeQuantifierForms(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestCatalogV5SafeQuantifierForms/existential_counterexample_restricted": {"term.inflection_verbatim", "singular 'every N'"},
		// term.inflection_verbatim: plural and singular class spellings must match verbatim.
		"TestCatalogV5SafeQuantifierForms/existential_counterexample_without_restrictor": {"term.inflection_verbatim", "singular 'every N'"},
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestCatalogV5SafeQuantifierForms/existential_generalization_there_exists": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestCatalogV5SafeQuantifierForms/existential_generalization_there_is": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestCatalogV5SafeQuantifierForms/existential_generalization_something_is": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestCatalogV5SafeQuantifierForms/restricted_existential_introduction_single_predicate": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestCatalogV5SafeQuantifierForms/restricted_existential_introduction_conjunctive_predicate": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestCatalogV5SafeQuantifierForms/restricted_existential_introduction_some_singular": {"eg.name_witness", "state the existential claim as its own premise"},
		// eg.name_witness: the existential witness lacks the required name evidence.
		"TestCatalogV5SafeQuantifierForms/restricted_existential_introduction_some_plural": {"eg.name_witness", "state the existential claim as its own premise"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV5SafeQuantifierForms/existential_counterexample_unrestricted": {"scope.clause_start", "move the operator into each clause"},
	}

	positives := []struct {
		name     string
		argument string
		form     string
	}{
		{
			name: "existential generalization there exists",
			argument: `1. Sample R is defective.
2. Therefore, there exists something that is defective. [1]`,
			form: "existential generalization",
		},
		{
			name: "existential generalization there is",
			argument: `1. Sample R is defective.
2. Therefore, there is something that is defective. [1]`,
			form: "existential generalization",
		},
		{
			name: "existential generalization something is",
			argument: `1. Sample R is defective.
2. Therefore, something is defective. [1]`,
			form: "existential generalization",
		},
		{
			name: "restricted existential introduction single predicate",
			argument: `1. Dana is a healthy adult.
2. Dana is sleeping six hours.
3. Therefore, there exist healthy adults that are sleeping six hours. [1,2]`,
			form: "existential introduction",
		},
		{
			name: "restricted existential introduction conjunctive predicate",
			argument: `1. Lee is a natural short sleeper.
2. Lee is habitually sleeping six hours.
3. Lee is showing no health harm.
4. Therefore, there exist natural short sleepers that are habitually sleeping six hours and showing no health harm. [1,2,3]`,
			form: "existential introduction",
		},
		{
			name: "restricted existential introduction some singular",
			argument: `1. Dana is a healthy adult.
2. Dana is sleeping six hours.
3. Therefore, some healthy adult is sleeping six hours. [1,2]`,
			form: "existential introduction",
		},
		{
			name: "restricted existential introduction some plural",
			argument: `1. Dana is a healthy adult.
2. Dana is sleeping six hours.
3. Therefore, some healthy adults are sleeping six hours. [1,2]`,
			form: "existential introduction",
		},
		{
			name: "existential counterexample restricted",
			argument: `1. There exist healthy adults that are sleep-deprived and not impaired.
2. Therefore, it is not the case that every healthy adult that is sleep-deprived is impaired. [1]`,
			form: "existential counterexample",
		},
		{
			name: "existential counterexample without restrictor",
			argument: `1. There exist cats that are not asleep.
2. Therefore, it is not the case that every cat is asleep. [1]`,
			form: "existential counterexample",
		},
		{
			name: "existential counterexample unrestricted",
			argument: `1. There exists something that is sleep-deprived and not impaired.
2. Therefore, it is not the case that everything that is sleep-deprived is impaired. [1]`,
			form: "existential counterexample",
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

func TestCatalogV5SafeQuantifierFormsRejectInvalidForms(t *testing.T) {
	negatives := []struct {
		name     string
		argument string
	}{
		{
			name: "existential generalization predicate paraphrase",
			argument: `1. Sample R is broken.
2. Therefore, there exists something that is defective. [1]`,
		},
		{
			name: "existential generalization does not import existence from a universal",
			argument: `1. Every sample is defective.
2. Therefore, there exists something that is defective. [1]`,
		},
		{
			name: "existential introduction different subjects",
			argument: `1. Dana is a healthy adult.
2. Lee is sleeping six hours.
3. Therefore, there exist healthy adults that are sleeping six hours. [1,2]`,
		},
		{
			name: "existential introduction class mismatch",
			argument: `1. Dana is a cat.
2. Dana is sleeping six hours.
3. Therefore, there exist dogs that are sleeping six hours. [1,2]`,
		},
		{
			name: "existential introduction unused cited premise",
			argument: `1. Dana is a healthy adult.
2. Dana is sleeping six hours.
3. Dana is cheerful.
4. Therefore, there exist healthy adults that are sleeping six hours. [1,2,3]`,
		},
		{
			name: "existential introduction does not combine universal subjects as witnesses",
			argument: `1. Every cat is a mammal.
2. Every cat is furry.
3. Therefore, there exist mammals that are furry. [1,2]`,
		},
		{
			name: "existential counterexample consequent mismatch",
			argument: `1. There exist healthy adults that are sleep-deprived and not tired.
2. Therefore, it is not the case that every healthy adult that is sleep-deprived is impaired. [1]`,
		},
		{
			name: "existential counterexample class mismatch",
			argument: `1. There exist healthy adults that are sleep-deprived and not impaired.
2. Therefore, it is not the case that every healthy child that is sleep-deprived is impaired. [1]`,
		},
		{
			name: "existential counterexample restrictor mismatch",
			argument: `1. There exist healthy adults that are sleep-deprived and not impaired.
2. Therefore, it is not the case that every healthy adult that is exhausted is impaired. [1]`,
		},
		{
			name: "existential affirmative smuggling",
			argument: `1. There exist cats that are furry.
2. Therefore, every cat is furry. [1]`,
		},
		{
			name: "some does not entail every",
			argument: `1. Some cat is furry.
2. Therefore, every cat is furry. [1]`,
		},
		{
			name: "restricted existential requires a class witness",
			argument: `1. Sample R is defective.
2. Therefore, there exist healthy adults that are defective. [1]`,
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
