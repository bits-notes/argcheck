package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	SchemaVersionV1 = "validator/v1"

	KindBuiltin    = "builtin"
	KindJSONSchema = "json_schema"
	KindRegexMatch = "regex_match"
	KindLineRules  = "line_rules"

	BuiltinNone                     = "none"
	BuiltinNonEmpty                 = "non_empty"
	BuiltinNumberedArgumentSyntax   = "numbered_argument_syntax"
	BuiltinDisputationChallenge     = "disputation_challenge"
	BuiltinDisputationProseResponse = "disputation_prose_response"
	BuiltinConditionalityPreserve   = "conditionality_preservation"

	defaultMaxOutputBytes = 128 * 1024
	defaultDeadline       = 250 * time.Millisecond
)

// Spec is the JSON contract for deterministic output validation.
type Spec struct {
	SchemaVersion string          `json:"schema_version,omitempty"`
	Kind          string          `json:"kind"`
	Name          string          `json:"name,omitempty"`
	Options       json.RawMessage `json:"options,omitempty"`
	Spec          json.RawMessage `json:"spec,omitempty"`
}

// Result is the product-facing answer returned by a validator.
type Result struct {
	Passed       bool            `json:"passed"`
	Message      string          `json:"message"`
	Details      map[string]any  `json:"details,omitempty"`
	ParsedOutput json.RawMessage `json:"parsed_output,omitempty"`
}

// Error carries a failed validation result through legacy ValidateFunc paths.
type Error struct {
	Result Result
}

func (e *Error) Error() string {
	if e == nil {
		return "validation failed"
	}
	return "validation failed: " + e.Result.Message
}

// OutputValidator validates an LLM output deterministically.
type OutputValidator interface {
	Validate(ctx context.Context, output string) Result
}

// LegacyKindSpec converts the older validator_kind string to a full spec.
func LegacyKindSpec(kind string) Spec {
	switch strings.TrimSpace(kind) {
	case "", BuiltinNonEmpty:
		return BuiltinSpec(BuiltinNonEmpty)
	case BuiltinNone:
		return BuiltinSpec(BuiltinNone)
	case BuiltinNumberedArgumentSyntax:
		return BuiltinSpec(BuiltinNumberedArgumentSyntax)
	case BuiltinDisputationChallenge:
		return BuiltinSpec(BuiltinDisputationChallenge)
	case BuiltinDisputationProseResponse:
		return BuiltinSpec(BuiltinDisputationProseResponse)
	case BuiltinConditionalityPreserve:
		return BuiltinSpec(BuiltinConditionalityPreserve)
	default:
		return Spec{Kind: strings.TrimSpace(kind)}
	}
}

// BuiltinSpec returns the standard JSON shape for a named built-in validator.
func BuiltinSpec(name string) Spec {
	return Spec{
		SchemaVersion: SchemaVersionV1,
		Kind:          KindBuiltin,
		Name:          name,
	}
}

// Normalize fills defaults and accepts compact legacy shapes.
func Normalize(spec Spec) Spec {
	spec.SchemaVersion = strings.TrimSpace(spec.SchemaVersion)
	spec.Kind = strings.TrimSpace(spec.Kind)
	spec.Name = strings.TrimSpace(spec.Name)

	if spec.Kind == "" {
		spec.Kind = KindBuiltin
		spec.Name = BuiltinNonEmpty
	}
	if spec.SchemaVersion == "" {
		spec.SchemaVersion = SchemaVersionV1
	}
	if spec.Kind == BuiltinNonEmpty || spec.Kind == BuiltinNone || spec.Kind == BuiltinNumberedArgumentSyntax || spec.Kind == BuiltinDisputationChallenge || spec.Kind == BuiltinDisputationProseResponse || spec.Kind == BuiltinConditionalityPreserve {
		spec.Name = spec.Kind
		spec.Kind = KindBuiltin
	}
	if spec.Kind == KindBuiltin && spec.Name == "" {
		spec.Name = BuiltinNonEmpty
	}
	return spec
}

// Compile checks a validator spec and returns an executable validator.
func Compile(spec Spec) (OutputValidator, error) {
	spec = Normalize(spec)
	if spec.SchemaVersion != SchemaVersionV1 {
		return nil, fmt.Errorf("unsupported validator schema_version %q", spec.SchemaVersion)
	}

	switch spec.Kind {
	case KindBuiltin:
		return compileBuiltin(spec)
	case KindJSONSchema:
		return compileJSONSchema(spec)
	case KindRegexMatch:
		return compileRegexMatch(spec)
	case KindLineRules:
		return compileLineRules(spec)
	default:
		return nil, fmt.Errorf("unknown validator kind %q", spec.Kind)
	}
}

// Validate compiles and runs a validator in one call.
func Validate(ctx context.Context, spec Spec, output string) Result {
	validator, err := Compile(spec)
	if err != nil {
		return Result{Passed: false, Message: err.Error()}
	}
	return validator.Validate(ctx, output)
}

func compileBuiltin(spec Spec) (OutputValidator, error) {
	switch spec.Name {
	case BuiltinNone:
		return noneValidator{}, nil
	case BuiltinNonEmpty:
		if len(spec.Options) > 0 {
			return nil, fmt.Errorf("validator %q does not support options", spec.Name)
		}
		return nonEmptyValidator{}, nil
	case BuiltinNumberedArgumentSyntax:
		options, err := parseNumberedArgumentOptions(spec.Options)
		if err != nil {
			return nil, err
		}
		return numberedArgumentSyntaxValidator{options: options}, nil
	case BuiltinDisputationChallenge:
		options, err := parseDisputationChallengeOptions(spec.Options)
		if err != nil {
			return nil, err
		}
		return disputationChallengeValidator{options: options}, nil
	case BuiltinDisputationProseResponse:
		options, err := parseDisputationProseResponseOptions(spec.Options)
		if err != nil {
			return nil, err
		}
		return disputationProseResponseValidator{options: options}, nil
	case BuiltinConditionalityPreserve:
		options, err := parseConditionalityPreservationOptions(spec.Options)
		if err != nil {
			return nil, err
		}
		return conditionalityPreservationValidator{options: options}, nil
	default:
		return nil, fmt.Errorf("unknown built-in validator %q", spec.Name)
	}
}

func withValidationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, defaultDeadline)
}

func checkOutputLimit(output string) Result {
	if len(output) > defaultMaxOutputBytes {
		return Result{
			Passed:  false,
			Message: fmt.Sprintf("output exceeds %d byte validation limit", defaultMaxOutputBytes),
		}
	}
	return Result{Passed: true}
}

type noneValidator struct{}

func (noneValidator) Validate(ctx context.Context, output string) Result {
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	select {
	case <-ctx.Done():
		return Result{Passed: false, Message: "validation timed out"}
	default:
		return Result{Passed: true, Message: "validation skipped"}
	}
}

type nonEmptyValidator struct{}

func (nonEmptyValidator) Validate(ctx context.Context, output string) Result {
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Passed: false, Message: "validation timed out"}
	default:
	}
	if strings.TrimSpace(output) == "" {
		return Result{Passed: false, Message: "output is empty"}
	}
	return Result{Passed: true, Message: "output matches non_empty"}
}
