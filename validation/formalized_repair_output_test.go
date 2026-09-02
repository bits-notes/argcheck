package validation

import (
	"context"
	"strings"
	"testing"
)

const validFormalizedRepairOutput = `{
  "argument": "1. If A, then B.\n2. A.\n3. Therefore, B. [1,2]",
  "coverage_anchor": "source",
  "line_mapping": [
    {"repaired_line": 1, "from_as_found_lines": [1]},
    {"repaired_line": 2, "from_as_found_lines": [2]},
    {"repaired_line": 3, "from_as_found_lines": [3]}
  ],
  "dropped": []
}`

func TestFormalizedRepairOutputValidatorAcceptsContractAndParsesJSON(t *testing.T) {
	result := Validate(context.Background(), BuiltinSpec(BuiltinFormalizedRepairOutput), validFormalizedRepairOutput)
	if !result.Passed || len(result.ParsedOutput) == 0 || result.Details["step_count"] != 3 {
		t.Fatalf("validation result = %+v", result)
	}
}

func TestFormalizedRepairOutputValidatorRejectsMalformedContract(t *testing.T) {
	tests := []struct {
		name, output, want string
	}{
		{name: "missing dropped", output: `{"argument":"1. A.\n2. Therefore, A. [1]","coverage_anchor":"source","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]}]}`, want: "requires dropped"},
		{name: "missing coverage anchor", output: strings.Replace(validFormalizedRepairOutput, "  \"coverage_anchor\": \"source\",\n", "", 1), want: "requires coverage_anchor"},
		{name: "unknown top-level field", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [], "extra": true`, 1), want: "unknown field"},
		{name: "unknown dropped field", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [{"as_found_line":4,"kind":"premise","reason":"unused","text":"model-authored"}]`, 1), want: "unknown field"},
		{name: "invalid kind", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [{"as_found_line":4,"kind":"other","reason":"unused"}]`, 1), want: "kind must be"},
		{name: "empty reason", output: strings.Replace(validFormalizedRepairOutput, `"dropped": []`, `"dropped": [{"as_found_line":4,"kind":"premise","reason":" "}]`, 1), want: "reason must be non-empty"},
		{name: "empty mapping sources", output: strings.Replace(validFormalizedRepairOutput, `"from_as_found_lines": [1]`, `"from_as_found_lines": []`, 1), want: "at least one line"},
		{name: "mapping outside argument", output: strings.Replace(validFormalizedRepairOutput, `"repaired_line": 3`, `"repaired_line": 4`, 1), want: "outside the argument"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Validate(context.Background(), BuiltinSpec(BuiltinFormalizedRepairOutput), testCase.output)
			if result.Passed || !strings.Contains(result.Message, testCase.want) {
				t.Fatalf("validation result = %+v, want message containing %q", result, testCase.want)
			}
		})
	}
}

func TestFormalizedRepairOutputValidatorAppliesNumberedArgumentSyntax(t *testing.T) {
	invalid := strings.Replace(validFormalizedRepairOutput, "3. Therefore, B. [1,2]", "3. B. [1,2]", 1)
	result := Validate(context.Background(), BuiltinSpec(BuiltinFormalizedRepairOutput), invalid)
	if result.Passed || !strings.Contains(result.Message, "failed numbered_argument_syntax") {
		t.Fatalf("validation result = %+v", result)
	}
}

func TestFormalizedRepairOutputValidatorRejectsOptions(t *testing.T) {
	spec := BuiltinSpec(BuiltinFormalizedRepairOutput)
	spec.Options = []byte(`{"allow_blank_lines":true}`)
	if _, err := Compile(spec); err == nil || !strings.Contains(err.Error(), "does not support options") {
		t.Fatalf("Compile error = %v", err)
	}
}

func TestFormalizedRepairOutputValidatorEnforcesSourceCoverageFromRepairPrompt(t *testing.T) {
	prompt := `Repair.
<source_material>
1. If it rains, then the ground is wet.
2. It rains.
3. Therefore, the ground is wet. [1,2]
4. If the sun shines, then the sky is bright.
5. The sun shines.
6. Therefore, the sky is bright. [4,5]
</source_material>
<as_found_argument>
1. If it rains, then the ground is wet.
2. It rains.
3. Therefore, the ground is wet. [1,2]
</as_found_argument>`
	output := `{
  "argument":"1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]",
  "coverage_anchor":"source",
  "line_mapping":[
    {"repaired_line":1,"from_as_found_lines":[1]},
    {"repaired_line":2,"from_as_found_lines":[2]},
    {"repaired_line":3,"from_as_found_lines":[3]}
  ],
  "dropped":[
    {"as_found_line":4,"kind":"premise","reason":"separate sub-argument"},
    {"as_found_line":5,"kind":"premise","reason":"separate sub-argument"},
    {"as_found_line":6,"kind":"conclusion","reason":"separate sub-argument"}
  ]
}`
	result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), output, prompt)
	if !result.Passed || result.Details["coverage_anchor"] != FormalizationCoverageAnchorSource {
		t.Fatalf("source-anchored validation = %+v", result)
	}

	orphan := strings.Replace(output, `,
    {"as_found_line":6,"kind":"conclusion","reason":"separate sub-argument"}`, "", 1)
	result = ValidateFormalizedRepairOutputWithPrompt(context.Background(), orphan, prompt)
	if result.Passed || !strings.Contains(result.Message, "source line 6 is not covered") {
		t.Fatalf("orphan source validation = %+v", result)
	}
}

func TestFormalizedRepairOutputValidatorAcceptsDeadPremiseAndDerivedConclusionFixtures(t *testing.T) {
	deadPrompt := `<source_material>
1. If it rains, then the ground is wet.
2. It rains.
3. The moon orbits the Earth.
4. Therefore, the ground is wet. [1,2]
</source_material>
<as_found_argument>
1. If it rains, then the ground is wet.
2. It rains.
3. Therefore, the ground is wet. [1,2]
</as_found_argument>`
	deadOutput := `{"argument":"1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]","coverage_anchor":"source","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]},{"repaired_line":2,"from_as_found_lines":[2]},{"repaired_line":3,"from_as_found_lines":[4]}],"dropped":[{"as_found_line":3,"kind":"premise","reason":"dead premise"}]}`
	if result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), deadOutput, deadPrompt); !result.Passed {
		t.Fatalf("dead-premise repair rejected: %+v", result)
	}

	derivedPrompt := `<source_material>
1. A thermometer recorded 38.4 degrees.
2. The measurement conditions were valid.
3. Therefore, the patient had pyrexia. [1,2]
</source_material>
<as_found_argument>
1. A thermometer recorded 38.4 degrees.
2. The measurement conditions were valid.
3. Therefore, the patient had pyrexia. [1,2]
</as_found_argument>`
	derivedOutput := `{"argument":"1. A thermometer recorded 38.4 degrees.\n2. The measurement conditions were valid.\n3. If a thermometer recorded 38.4 degrees under valid conditions, then the temperature exceeded 38 degrees.\n4. Therefore, the temperature exceeded 38 degrees. [1,2,3]\n5. If the temperature exceeded 38 degrees, then the patient had pyrexia.\n6. Therefore, the patient had pyrexia. [4,5]","coverage_anchor":"source","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]},{"repaired_line":2,"from_as_found_lines":[2]},{"repaired_line":6,"from_as_found_lines":[3]}],"dropped":[]}`
	if result := ValidateFormalizedRepairOutputWithPrompt(context.Background(), derivedOutput, derivedPrompt); !result.Passed {
		t.Fatalf("derived-conclusion repair rejected: %+v", result)
	}
}
