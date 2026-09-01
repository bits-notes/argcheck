package validation

import (
	"context"
	"strings"
	"testing"
)

const validFormalizedRepairOutput = `{
  "argument": "1. If A, then B.\n2. A.\n3. Therefore, B. [1,2]",
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
		{name: "missing dropped", output: `{"argument":"1. A.\n2. Therefore, A. [1]","line_mapping":[{"repaired_line":1,"from_as_found_lines":[1]}]}`, want: "requires dropped"},
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
