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
