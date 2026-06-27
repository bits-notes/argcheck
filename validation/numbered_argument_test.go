package validation

import (
	"context"
	"strings"
	"testing"
)

func TestNumberedArgumentSyntaxValidator_ValidExamples(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{
			name: "minimal",
			output: strings.Join([]string{
				"1. All bachelors are unmarried.",
				"2. Therefore, all bachelors are unmarried. [1]",
			}, "\n"),
		},
		{
			name: "simple",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. All animals are things.",
				"3. Therefore, all cats are things. [1,2]",
			}, "\n"),
		},
		{
			name: "multi step linear",
			output: strings.Join([]string{
				"1. Socrates is a person.",
				"2. All people are mortal.",
				"3. Therefore, Socrates is mortal. [1,2]",
				"4. All mortal things have parts.",
				"5. Therefore, Socrates has parts. [3,4]",
				"6. All things that have parts are made of particles.",
				"7. Therefore, Socrates is made of particles. [5,6]",
			}, "\n"),
		},
		{
			name: "inline references in bridge premise",
			output: strings.Join([]string{
				"1. The first premise is available.",
				"2. If [1], then the argument can refer to it.",
				"3. Therefore, the argument can refer to the first premise. [1,2]",
			}, "\n"),
		},
	}

	validator, err := Compile(BuiltinSpec(BuiltinNumberedArgumentSyntax))
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(context.Background(), tt.output)
			if !result.Passed {
				t.Fatalf("valid argument failed: %s", result.Message)
			}
		})
	}
}

func TestNumberedArgumentSyntaxValidator_InvalidAntiExamples(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantMessage string
	}{
		{
			name: "missing conclusion marker",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. All animals are things.",
				"3. All cats are things. [1,2]",
			}, "\n"),
			wantMessage: "line 3 is not a valid premise or conclusion",
		},
		{
			name: "forward reference",
			output: strings.Join([]string{
				"1. Therefore, Socrates is mortal. [2,3]",
				"2. Socrates is a person.",
				"3. All people are mortal.",
			}, "\n"),
			wantMessage: "step 1 cites current or future step 2",
		},
		{
			name: "extraneous material",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. All animals are things.",
				"Now, combining these two well-known facts:",
				"3. Therefore, all cats are things. [1,2]",
			}, "\n"),
			wantMessage: "line 3 is not a valid premise or conclusion",
		},
		{
			name: "missing therefore with terminal citations",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. All animals are things.",
				"3. All cats are things. [1,2]",
			}, "\n"),
			wantMessage: "line 3 is not a valid premise or conclusion",
		},
		{
			name: "dangling citation",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. All animals are things.",
				"3. Therefore, all cats are things. [1,2,4]",
			}, "\n"),
			wantMessage: "step 3 cites current or future step 4",
		},
		{
			name: "skipped numbering",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"3. Therefore, all cats are animals. [1]",
			}, "\n"),
			wantMessage: "numbering jumps from 1 to 3",
		},
		{
			name: "citation spaces",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. All animals are things.",
				"3. Therefore, all cats are things. [1, 2]",
			}, "\n"),
			wantMessage: "line 3 is not a valid premise or conclusion",
		},
		{
			name: "final premise",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. Therefore, all cats are animals. [1]",
				"3. The final line is a premise.",
			}, "\n"),
			wantMessage: "final step must be a conclusion",
		},
		{
			name: "blank line",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"",
				"2. Therefore, all cats are animals. [1]",
			}, "\n"),
			wantMessage: "line 2 is blank",
		},
		{
			name: "intermediate therefore missing terminal citations",
			output: strings.Join([]string{
				"1. All cats are animals.",
				"2. Therefore, cats are animals.",
				"3. Therefore, cats are animals. [2]",
			}, "\n"),
			wantMessage: "line 2 starts with Therefore, but does not have valid terminal citations",
		},
		{
			name: "source note bracket in premise",
			output: strings.Join([]string{
				"1. LSAG considered stable black-hole scenarios [LSAG 2008].",
				"2. Therefore, LSAG considered stable black-hole scenarios. [1]",
			}, "\n"),
			wantMessage: "line 1 contains non-numeric bracket reference",
		},
		{
			name: "source note bracket in conclusion statement",
			output: strings.Join([]string{
				"1. LSAG considered stable black-hole scenarios.",
				"2. Therefore, LSAG considered stable black-hole scenarios [LSAG 2008]. [1]",
			}, "\n"),
			wantMessage: "line 2 contains non-numeric bracket reference",
		},
		{
			name: "numeric source note bracket in premise",
			output: strings.Join([]string{
				"1. LSAG considered stable black-hole scenarios [2008].",
				"2. Therefore, LSAG considered stable black-hole scenarios. [1]",
			}, "\n"),
			wantMessage: "line 1 contains inline bracket references outside an explicit bridge premise",
		},
		{
			name: "numeric source page bracket in conclusion statement",
			output: strings.Join([]string{
				"1. LSAG considered stable black-hole scenarios.",
				"2. Therefore, LSAG considered stable black-hole scenarios [12]. [1]",
			}, "\n"),
			wantMessage: "line 2 contains inline bracket references outside an explicit bridge premise",
		},
		{
			name: "current-step inline reference",
			output: strings.Join([]string{
				"1. If [1], then the argument can refer to it.",
				"2. Therefore, the argument can refer to it. [1]",
			}, "\n"),
			wantMessage: "line 1 contains bracketed step reference",
		},
		{
			name: "source note inside bridge-looking premise",
			output: strings.Join([]string{
				"1. LSAG is authoritative.",
				"2. If LSAG [1] is authoritative, then the claim is supported.",
				"3. Therefore, the claim is supported. [1,2]",
			}, "\n"),
			wantMessage: "line 2 contains inline bracket references outside an explicit bridge premise",
		},
	}

	validator, err := Compile(BuiltinSpec(BuiltinNumberedArgumentSyntax))
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(context.Background(), tt.output)
			if result.Passed {
				t.Fatal("invalid argument passed")
			}
			if !strings.Contains(result.Message, tt.wantMessage) {
				t.Fatalf("message = %q, want containing %q", result.Message, tt.wantMessage)
			}
		})
	}
}

func TestNumberedArgumentSyntaxValidator_FencedBlockOption(t *testing.T) {
	output := strings.Join([]string{
		"```text",
		"1. All cats are animals.",
		"2. Therefore, all cats are animals. [1]",
		"```",
	}, "\n")

	disallow, err := Compile(BuiltinSpec(BuiltinNumberedArgumentSyntax))
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if result := disallow.Validate(context.Background(), output); result.Passed {
		t.Fatal("fenced block passed without allow_single_fenced_block")
	}

	allowSpec := BuiltinSpec(BuiltinNumberedArgumentSyntax)
	allowSpec.Options = []byte(`{"allow_single_fenced_block":true}`)
	allow, err := Compile(allowSpec)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if result := allow.Validate(context.Background(), output); !result.Passed {
		t.Fatalf("fenced block failed with allow_single_fenced_block: %s", result.Message)
	}
}

func TestValidate_UnknownSpecFailsClosed(t *testing.T) {
	result := Validate(context.Background(), Spec{Kind: "custom_code"}, "anything")
	if result.Passed {
		t.Fatal("unknown validator kind passed")
	}
	if !strings.Contains(result.Message, `unknown validator kind "custom_code"`) {
		t.Fatalf("message = %q", result.Message)
	}
}
