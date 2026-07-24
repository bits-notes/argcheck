package validation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONSchemaValidatorPassesAndReturnsParsedOutput(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindJSONSchema,
		Spec: json.RawMessage(`{
			"type": "object",
			"required": ["status", "count"],
			"properties": {
				"status": {"type": "string", "enum": ["ok", "failed"]},
				"count": {"type": "integer", "minimum": 1}
			},
			"additionalProperties": false
		}`),
	}

	result := Validate(context.Background(), spec, `{"status":"ok","count":2}`)
	if !result.Passed {
		t.Fatalf("Passed = false, message = %q", result.Message)
	}
	if len(result.ParsedOutput) == 0 {
		t.Fatal("ParsedOutput is empty")
	}
	var parsed map[string]any
	if err := json.Unmarshal(result.ParsedOutput, &parsed); err != nil {
		t.Fatalf("ParsedOutput is invalid JSON: %v", err)
	}
	if parsed["status"] != "ok" {
		t.Fatalf("parsed status = %#v, want ok", parsed["status"])
	}
}

func TestJSONSchemaValidatorRejectsMixedProse(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindJSONSchema,
		Spec:          json.RawMessage(`{"type":"object"}`),
	}

	result := Validate(context.Background(), spec, `Here is JSON: {"status":"ok"}`)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
}

func TestJSONSchemaValidatorRejectsAdditionalProperties(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindJSONSchema,
		Spec: json.RawMessage(`{
			"type": "object",
			"properties": {"status": {"type": "string"}},
			"additionalProperties": false
		}`),
	}

	result := Validate(context.Background(), spec, `{"status":"ok","extra":true}`)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
}

func TestJSONSchemaValidatorReportsArrayInsteadOfStringPath(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindJSONSchema,
		Spec: json.RawMessage(`{
			"type": "object",
			"required": ["remediation_instructions"],
			"properties": {
				"remediation_instructions": {"type": "string"}
			},
			"additionalProperties": false
		}`),
	}

	result := Validate(context.Background(), spec, `{"remediation_instructions":["fix the citations"]}`)
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "$.remediation_instructions has type array, expected string") {
		t.Fatalf("message = %q, want path-specific type error", result.Message)
	}
}

func TestJSONSchemaValidatorEnforcesOneOfObjectVariants(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindJSONSchema,
		Spec: json.RawMessage(`{
			"type": "object",
			"required": ["op", "reason"],
			"properties": {
				"op": {"type": "string", "enum": ["replace", "split"]},
				"reason": {"type": "string", "minLength": 1},
				"atom_id": {"type": "string"},
				"statement": {"type": "string"},
				"statements": {"type": "array", "items": {"type": "string"}}
			},
			"additionalProperties": false,
			"oneOf": [
				{
					"required": ["op", "atom_id", "statement", "reason"],
					"properties": {
						"op": {"const": "replace"},
						"atom_id": {},
						"statement": {},
						"reason": {}
					},
					"additionalProperties": false
				},
				{
					"required": ["op", "atom_id", "statements", "reason"],
					"properties": {
						"op": {"const": "split"},
						"atom_id": {},
						"statements": {},
						"reason": {}
					},
					"additionalProperties": false
				}
			]
		}`),
	}

	for _, output := range []string{
		`{"op":"replace","atom_id":"a1","statement":"P","reason":"repair"}`,
		`{"op":"split","atom_id":"a1","statements":["P","Q"],"reason":"repair"}`,
	} {
		if result := Validate(context.Background(), spec, output); !result.Passed {
			t.Fatalf("valid variant rejected: %s", result.Message)
		}
	}

	result := Validate(
		context.Background(),
		spec,
		`{"op":"replace","statement":"P","statements":["P","Q"],"reason":"repair"}`,
	)
	if result.Passed {
		t.Fatal("malformed cross-variant object passed")
	}
	if !strings.Contains(result.Message, "does not match any oneOf variant") ||
		!strings.Contains(result.Message, `missing required property "atom_id"`) {
		t.Fatalf("unexpected oneOf error: %q", result.Message)
	}
}

func TestJSONSchemaValidatorRequiresExactlyOneOneOfMatch(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindJSONSchema,
		Spec: json.RawMessage(`{
			"type": "integer",
			"oneOf": [
				{"minimum": 0},
				{"maximum": 10}
			]
		}`),
	}
	result := Validate(context.Background(), spec, `5`)
	if result.Passed {
		t.Fatal("value matching two oneOf variants passed")
	}
	if !strings.Contains(result.Message, "matches 2 oneOf variants") {
		t.Fatalf("unexpected oneOf overlap error: %q", result.Message)
	}
}
