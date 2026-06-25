package validation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestRegexMatchValidator(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindRegexMatch,
		Spec:          json.RawMessage(`{"pattern":"[A-Z]{2}-[0-9]{3}"}`),
	}

	result := Validate(context.Background(), spec, "AB-123")
	if !result.Passed {
		t.Fatalf("Passed = false, message = %q", result.Message)
	}

	result = Validate(context.Background(), spec, "AB-123\nextra")
	if result.Passed {
		t.Fatal("regex_match allowed a partial match")
	}
	if !strings.Contains(result.Message, "does not match regex_match") {
		t.Fatalf("message = %q", result.Message)
	}
}

func TestRegexMatchValidatorRejectsUnsafeSpecs(t *testing.T) {
	if _, err := Compile(Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindRegexMatch,
		Spec:          json.RawMessage(`{"pattern":"` + strings.Repeat("a", maxRegexPatternBytes+1) + `"}`),
	}); err == nil || !strings.Contains(err.Error(), "pattern exceeds") {
		t.Fatalf("long pattern error = %v", err)
	}

	if _, err := Compile(Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindRegexMatch,
		Spec:          json.RawMessage(`{"pattern":"[", "unexpected": true}`),
	}); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field error = %v", err)
	}
}

func TestLineRulesValidatorDocsExample(t *testing.T) {
	validator, err := Compile(numberedArgumentLineRulesSpec())
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	output := strings.Join([]string{
		"1. Socrates is a person.",
		"2. All people are mortal.",
		"3. Therefore, Socrates is mortal. [1,2]",
	}, "\n")
	result := validator.Validate(context.Background(), output)
	if !result.Passed {
		t.Fatalf("Passed = false, message = %q", result.Message)
	}
	if result.Details["line_count"] != 3 {
		t.Fatalf("line_count = %#v, want 3", result.Details["line_count"])
	}
}

func TestLineRulesValidatorRejectsMalformedOutput(t *testing.T) {
	validator, err := Compile(numberedArgumentLineRulesSpec())
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	tests := []struct {
		name        string
		output      string
		wantMessage string
	}{
		{
			name: "unmatched line",
			output: strings.Join([]string{
				"1. Socrates is a person.",
				"not a numbered line",
			}, "\n"),
			wantMessage: "line 2 matches 0 patterns",
		},
		{
			name: "skipped numbering",
			output: strings.Join([]string{
				"1. Socrates is a person.",
				"3. Therefore, Socrates is mortal. [1]",
			}, "\n"),
			wantMessage: "numbering jumps from 1 to 3",
		},
		{
			name: "final premise",
			output: strings.Join([]string{
				"1. Socrates is a person.",
				"2. Socrates has a name.",
			}, "\n"),
			wantMessage: `final line must match pattern "conclusion"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(context.Background(), tt.output)
			if result.Passed {
				t.Fatal("Passed = true, want false")
			}
			if !strings.Contains(result.Message, tt.wantMessage) {
				t.Fatalf("message = %q, want containing %q", result.Message, tt.wantMessage)
			}
		})
	}
}

func TestLineRulesValidatorRejectsUnsafeSpecs(t *testing.T) {
	t.Run("too many patterns", func(t *testing.T) {
		patterns := make([]string, 0, maxLineRulePatterns+1)
		for i := 0; i < maxLineRulePatterns+1; i++ {
			patterns = append(patterns, `"p`+string(rune('a'+i))+`":"^.+$"`)
		}
		raw := `{"patterns":{` + strings.Join(patterns, ",") + `},"rules":[{"type":"all_lines_match_one_pattern"}]}`
		if _, err := Compile(Spec{SchemaVersion: SchemaVersionV1, Kind: KindLineRules, Spec: json.RawMessage(raw)}); err == nil || !strings.Contains(err.Error(), "patterns exceed") {
			t.Fatalf("too many patterns error = %v", err)
		}
	})

	t.Run("too many rules", func(t *testing.T) {
		rules := make([]string, 0, maxLineRuleRules+1)
		for i := 0; i < maxLineRuleRules+1; i++ {
			rules = append(rules, `{"type":"all_lines_match_one_pattern"}`)
		}
		raw := `{"patterns":{"line":"^.+$"},"rules":[` + strings.Join(rules, ",") + `]}`
		if _, err := Compile(Spec{SchemaVersion: SchemaVersionV1, Kind: KindLineRules, Spec: json.RawMessage(raw)}); err == nil || !strings.Contains(err.Error(), "rules exceed") {
			t.Fatalf("too many rules error = %v", err)
		}
	})

	t.Run("unknown rule type", func(t *testing.T) {
		raw := `{"patterns":{"line":"^.+$"},"rules":[{"type":"run_javascript"}]}`
		if _, err := Compile(Spec{SchemaVersion: SchemaVersionV1, Kind: KindLineRules, Spec: json.RawMessage(raw)}); err == nil || !strings.Contains(err.Error(), "unknown type") {
			t.Fatalf("unknown rule error = %v", err)
		}
	})

	t.Run("pattern with missing capture", func(t *testing.T) {
		raw := `{"patterns":{"line":"^[0-9]+\\. .+$"},"rules":[{"type":"sequential_numbering","capture":1}]}`
		if _, err := Compile(Spec{SchemaVersion: SchemaVersionV1, Kind: KindLineRules, Spec: json.RawMessage(raw)}); err == nil || !strings.Contains(err.Error(), "does not have capture 1") {
			t.Fatalf("missing capture error = %v", err)
		}
	})
}

func TestLineRulesValidatorRejectsLineLimit(t *testing.T) {
	spec := Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindLineRules,
		Spec: json.RawMessage(`{
			"patterns": {"line": "^.+$"},
			"rules": [{"type": "all_lines_match_one_pattern"}]
		}`),
	}
	validator, err := Compile(spec)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	lines := make([]string, 0, maxLineRuleLines+1)
	for i := 0; i < maxLineRuleLines+1; i++ {
		lines = append(lines, "line")
	}
	result := validator.Validate(context.Background(), strings.Join(lines, "\n"))
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
	if !strings.Contains(result.Message, "line validation limit") {
		t.Fatalf("message = %q", result.Message)
	}
}

func numberedArgumentLineRulesSpec() Spec {
	return Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindLineRules,
		Spec: json.RawMessage(`{
			"allow_blank_lines": false,
			"patterns": {
				"premise": "([1-9][0-9]*)\\. .+\\.",
				"conclusion": "([1-9][0-9]*)\\. Therefore, .+\\. \\[([1-9][0-9]*(,[1-9][0-9]*)*)\\]"
			},
			"rules": [
				{"type": "all_lines_match_one_pattern"},
				{"type": "sequential_numbering", "capture": 1},
				{"type": "final_line_pattern", "pattern": "conclusion"}
			]
		}`),
	}
}
