package validation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCanonicalizeSpecJSONMakesObjectOrderAndWhitespaceInvariant(t *testing.T) {
	left := Spec{
		Kind:    KindJSONSchema,
		Options: json.RawMessage(` { "z": [ { "b": 2, "a": 1 } ], "a": true } `),
		Spec:    json.RawMessage(`{"required":["value"],"properties":{"value":{"maxLength":9,"type":"string"}},"type":"object"}`),
	}
	right := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindJSONSchema,
		Options:       json.RawMessage(`{"a":true,"z":[{"a":1,"b":2}]}`),
		Spec:          json.RawMessage("{\n  \"type\": \"object\",\n  \"properties\": {\"value\": {\"type\": \"string\", \"maxLength\": 9}},\n  \"required\": [\"value\"]\n}"),
	}

	canonicalLeft, err := CanonicalizeSpecJSON(left)
	if err != nil {
		t.Fatal(err)
	}
	canonicalRight, err := CanonicalizeSpecJSON(right)
	if err != nil {
		t.Fatal(err)
	}

	leftJSON, err := json.Marshal(canonicalLeft)
	if err != nil {
		t.Fatal(err)
	}
	rightJSON, err := json.Marshal(canonicalRight)
	if err != nil {
		t.Fatal(err)
	}
	if string(leftJSON) != string(rightJSON) {
		t.Fatalf("canonical specs differ:\nleft:  %s\nright: %s", leftJSON, rightJSON)
	}
	if got, want := string(canonicalLeft.Options), `{"a":true,"z":[{"a":1,"b":2}]}`; got != want {
		t.Fatalf("canonical options = %s, want %s", got, want)
	}
}

func TestCanonicalizeSpecJSONPreservesArrayOrderAndNumberLexemes(t *testing.T) {
	first, err := CanonicalizeSpecJSON(Spec{
		Kind: KindJSONSchema,
		Spec: json.RawMessage(`{"enum":[9007199254740993,1e+09],"required":["a","b"]}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := CanonicalizeSpecJSON(Spec{
		Kind: KindJSONSchema,
		Spec: json.RawMessage(`{"required":["b","a"],"enum":[9007199254740993,1e+09]}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(first.Spec) == string(second.Spec) {
		t.Fatal("array order must remain identity-significant")
	}
	if !strings.Contains(string(first.Spec), "9007199254740993") || !strings.Contains(string(first.Spec), "1e+09") {
		t.Fatalf("number lexemes were coerced: %s", first.Spec)
	}
}

func TestCanonicalizeSpecJSONReturnsFieldSpecificErrors(t *testing.T) {
	_, err := CanonicalizeSpecJSON(Spec{
		Kind:    KindBuiltin,
		Name:    BuiltinNumberedArgumentSyntax,
		Options: json.RawMessage(`{"allow_blank_lines":true} trailing`),
	})
	if err == nil || !strings.Contains(err.Error(), "validator options") {
		t.Fatalf("options error = %v, want field-specific error", err)
	}

	_, err = CanonicalizeSpecJSON(Spec{
		Kind: KindJSONSchema,
		Spec: json.RawMessage(`{"type":"object"} {"type":"array"}`),
	})
	if err == nil || !strings.Contains(err.Error(), "validator spec") {
		t.Fatalf("spec error = %v, want field-specific error", err)
	}
}

func TestCanonicalizeSpecJSONPreservesValidatorBehavior(t *testing.T) {
	original := Spec{
		Kind:    KindJSONSchema,
		Options: json.RawMessage(`{ "allow_markdown_fence": true }`),
		Spec: json.RawMessage(`{
			"required": ["status"],
			"properties": {"status": {"enum": ["ok"], "type": "string"}},
			"additionalProperties": false,
			"type": "object"
		}`),
	}
	canonical, err := CanonicalizeSpecJSON(original)
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{
		`{"status":"ok"}`,
		"```json\n{\"status\":\"ok\"}\n```",
		`{"status":"wrong"}`,
		`{"status":"ok","extra":true}`,
	} {
		before := Validate(context.Background(), original, output)
		after := Validate(context.Background(), canonical, output)
		if before.Passed != after.Passed || before.Message != after.Message {
			t.Fatalf("behavior changed for %q: before=%+v after=%+v", output, before, after)
		}
	}

	builtin := Spec{
		Kind: KindBuiltin,
		Name: BuiltinNumberedArgumentSyntax,
		Options: json.RawMessage(`{
			"allow_single_fenced_block": true,
			"allow_blank_lines": true
		}`),
	}
	canonicalBuiltin, err := CanonicalizeSpecJSON(builtin)
	if err != nil {
		t.Fatal(err)
	}
	output := "```\n1. A.\n\n2. Therefore, A. [1]\n```"
	before := Validate(context.Background(), builtin, output)
	after := Validate(context.Background(), canonicalBuiltin, output)
	if before.Passed != after.Passed || before.Message != after.Message {
		t.Fatalf("built-in behavior changed: before=%+v after=%+v", before, after)
	}
}
