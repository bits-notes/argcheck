package validation

import "testing"

func TestCatalogV7MPDischargedBiconditionals(t *testing.T) {
	positives := []struct {
		name     string
		argument string
		form     string
	}{
		{
			name: "minimal modus ponens",
			argument: `1. If the switch is powered, then the switch is up if and only if the lamp is lit.
2. The switch is powered.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
			form: "modus ponens",
		},
		{
			name: "N1 steps 15 through 18 shape",
			argument: `1. Permutation Invariance holds.
2. There is a permutation pi such that wi = vpi(i) and vi = wpi(i) for every individual i.
3. If Permutation Invariance holds and there is a permutation pi such that wi = vpi(i) and vi = wpi(i) for every individual i, then w is strictly preferred to v if and only if v is strictly preferred to w.
4. Therefore, w is strictly preferred to v if and only if v is strictly preferred to w. [1,2,3]`,
			form: "conjunctive modus ponens",
		},
		{
			name: "N1 parallel steps 23 through 25 shape",
			argument: `1. There is a permutation sigma that exchanges distributions x and y.
2. If there is a permutation sigma that exchanges distributions x and y, then x is preferred to y if and only if y is preferred to x.
3. Therefore, x is preferred to y if and only if y is preferred to x. [1,2]`,
			form: "modus ponens",
		},
	}

	for _, test := range positives {
		t.Run(test.name, func(t *testing.T) {
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

func TestCatalogV7MPDischargedBiconditionalsRejectAdversarialNearMisses(t *testing.T) {
	negatives := []struct {
		name     string
		argument string
	}{
		{
			name: "consequent synonym",
			argument: `1. If the switch is powered, then the switch is raised if and only if the lamp is illuminated.
2. The switch is powered.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "consequent words reordered",
			argument: `1. If the switch is powered, then the switch is fully raised if and only if the lamp is brightly lit.
2. The switch is powered.
3. Therefore, the switch is raised fully if and only if the lamp is lit brightly. [1,2]`,
		},
		{
			name: "consequent sides swapped",
			argument: `1. If the switch is powered, then the switch is up if and only if the lamp is lit.
2. The switch is powered.
3. Therefore, the lamp is lit if and only if the switch is up. [1,2]`,
		},
		{
			name: "antecedent wording drift",
			argument: `1. If the switch is powered, then the switch is up if and only if the lamp is lit.
2. The switch has power.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "conjunctive antecedent not fully established",
			argument: `1. If the switch is powered and the controller is online, then the switch is up if and only if the lamp is lit.
2. The switch is powered.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "conclusion cites only the conditional",
			argument: `1. If the switch is powered, then the switch is up if and only if the lamp is lit.
2. Therefore, the switch is up if and only if the lamp is lit. [1]`,
		},
		{
			name: "available antecedent is not cited",
			argument: `1. If the switch is powered, then the switch is up if and only if the lamp is lit.
2. The switch is powered.
3. Therefore, the switch is up if and only if the lamp is lit. [1]`,
		},
		{
			name: "one directional consequent does not license iff",
			argument: `1. If the switch is powered, then the lamp is lit.
2. The switch is powered.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "two nonmatching directions still do not introduce iff",
			argument: `1. If the switch is raised, then the lamp is lit.
2. If the lamp is lit, then the switch is raised.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "iff is only one conjunct of the consequent",
			argument: `1. If the switch is powered, then the switch is up if and only if the lamp is lit and the controller is online.
2. The switch is powered.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "iff is embedded in the consequent attribution",
			argument: `1. If the switch is powered, then the theorem states that the switch is up if and only if the lamp is lit.
2. The switch is powered.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "extra cited premise is not antecedent support",
			argument: `1. If the switch is powered, then the switch is up if and only if the lamp is lit.
2. The switch is powered.
3. The controller is online.
4. Therefore, the switch is up if and only if the lamp is lit. [1,2,3]`,
		},
	}

	for _, test := range negatives {
		t.Run(test.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(test.argument)
			if result.Passed || report.Passed {
				t.Fatalf("adversarial near-miss unexpectedly passed: result=%#v report=%#v", result, report)
			}
		})
	}
}
