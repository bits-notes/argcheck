package validation

import "testing"

func TestCatalogV6BiconditionalForms(t *testing.T) {
	positives := []struct {
		name     string
		argument string
		form     string
	}{
		{
			name: "biconditional introduction",
			argument: `1. If the switch is up, then the lamp is lit.
2. If the lamp is lit, then the switch is up.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
			form: "biconditional introduction",
		},
		{
			name: "biconditional elimination forward",
			argument: `1. The switch is up if and only if the lamp is lit.
2. Therefore, if the switch is up, then the lamp is lit. [1]`,
			form: "biconditional elimination",
		},
		{
			name: "biconditional elimination reverse",
			argument: `1. The switch is up if and only if the lamp is lit.
2. Therefore, if the lamp is lit, then the switch is up. [1]`,
			form: "biconditional elimination",
		},
		{
			name: "biconditional modus ponens forward",
			argument: `1. The switch is up if and only if the lamp is lit.
2. The switch is up.
3. Therefore, the lamp is lit. [1,2]`,
			form: "biconditional modus ponens",
		},
		{
			name: "biconditional modus ponens reverse",
			argument: `1. The switch is up if and only if the lamp is lit.
2. The lamp is lit.
3. Therefore, the switch is up. [1,2]`,
			form: "biconditional modus ponens",
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

func TestCatalogV6BiconditionalFormsRejectAdversarialNearMisses(t *testing.T) {
	negatives := []struct {
		name     string
		argument string
	}{
		{
			name: "only if is not a biconditional conclusion",
			argument: `1. If the switch is up, then the lamp is lit.
2. If the lamp is lit, then the switch is up.
3. Therefore, the switch is up only if the lamp is lit. [1,2]`,
		},
		{
			name: "only if is not a biconditional premise for elimination",
			argument: `1. The switch is up only if the lamp is lit.
2. Therefore, if the switch is up, then the lamp is lit. [1]`,
		},
		{
			name: "only if is not a biconditional premise for modus ponens",
			argument: `1. The switch is up only if the lamp is lit.
2. The switch is up.
3. Therefore, the lamp is lit. [1,2]`,
		},
		{
			name: "one missing direction does not introduce a biconditional",
			argument: `1. If the switch is up, then the lamp is lit.
2. Therefore, the switch is up if and only if the lamp is lit. [1]`,
		},
		{
			name: "duplicate forward directions do not introduce a biconditional",
			argument: `1. If the switch is up, then the lamp is lit.
2. If the switch is up, then the lamp is lit.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "contrapositive does not substitute for reverse direction",
			argument: `1. If the switch is up, then the lamp is lit.
2. If it is not the case that the lamp is lit, then it is not the case that the switch is up.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "introduction rejects clause wording drift",
			argument: `1. If the switch is up, then the lamp is illuminated.
2. If the lamp is lit, then the switch is up.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "elimination rejects clause wording drift",
			argument: `1. The switch is up if and only if the lamp is lit.
2. Therefore, if the switch is raised, then the lamp is lit. [1]`,
		},
		{
			name: "modus ponens rejects clause wording drift",
			argument: `1. The switch is up if and only if the lamp is lit.
2. The switch is raised.
3. Therefore, the lamp is lit. [1,2]`,
		},
		{
			name: "embedded biconditional does not support elimination",
			argument: `1. The theorem states that the switch is up if and only if the lamp is lit.
2. Therefore, if the theorem states that the switch is up, then the lamp is lit. [1]`,
		},
		{
			name: "embedded biconditional does not support modus ponens",
			argument: `1. The theorem states that the switch is up if and only if the lamp is lit.
2. The theorem states that the switch is up.
3. Therefore, the lamp is lit. [1,2]`,
		},
		{
			name: "embedded biconditional conclusion is not introduction",
			argument: `1. If the theorem states that the switch is up, then the lamp is lit.
2. If the lamp is lit, then the theorem states that the switch is up.
3. Therefore, the theorem states that the switch is up if and only if the lamp is lit. [1,2]`,
		},
		{
			name: "prefatory attribution does not support elimination",
			argument: `1. According to the theorem, the switch is up if and only if the lamp is lit.
2. Therefore, if according to the theorem, the switch is up, then the lamp is lit. [1]`,
		},
		{
			name: "negated forward conclusion does not follow by biconditional modus ponens",
			argument: `1. The switch is up if and only if the lamp is lit.
2. The switch is up.
3. Therefore, it is not the case that the lamp is lit. [1,2]`,
		},
		{
			name: "negated reverse conclusion does not follow by biconditional modus ponens",
			argument: `1. The switch is up if and only if the lamp is lit.
2. The lamp is lit.
3. Therefore, it is not the case that the switch is up. [1,2]`,
		},
		{
			name: "negated conditional conclusion does not follow by biconditional elimination",
			argument: `1. The switch is up if and only if the lamp is lit.
2. Therefore, it is not the case that if the switch is up, then the lamp is lit. [1]`,
		},
		{
			name: "multiple biconditional connectives are not a single biconditional",
			argument: `1. If the switch is up, then the lamp is lit if and only if the alarm is armed.
2. If the lamp is lit if and only if the alarm is armed, then the switch is up.
3. Therefore, the switch is up if and only if the lamp is lit if and only if the alarm is armed. [1,2]`,
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
