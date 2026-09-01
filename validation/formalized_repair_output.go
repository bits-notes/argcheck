package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type formalizedRepairOutput struct {
	Argument    *string                         `json:"argument"`
	LineMapping *[]formalizedRepairLineMapping  `json:"line_mapping"`
	Dropped     *[]formalizedRepairDroppedEntry `json:"dropped"`
}

type formalizedRepairLineMapping struct {
	RepairedLine     int   `json:"repaired_line"`
	FromAsFoundLines []int `json:"from_as_found_lines"`
}

type formalizedRepairDroppedEntry struct {
	AsFoundLine int    `json:"as_found_line"`
	Kind        string `json:"kind"`
	Reason      string `json:"reason"`
}

type formalizedRepairOutputValidator struct{}

func (formalizedRepairOutputValidator) Validate(ctx context.Context, output string) Result {
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

	decoder := json.NewDecoder(bytes.NewBufferString(output))
	decoder.DisallowUnknownFields()
	var repair formalizedRepairOutput
	if err := decoder.Decode(&repair); err != nil {
		return Result{Passed: false, Message: "output must be one formalized repair JSON object: " + err.Error()}
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Result{Passed: false, Message: "output must contain only one formalized repair JSON object"}
	}
	if repair.Argument == nil {
		return Result{Passed: false, Message: "formalized repair output requires argument"}
	}
	if repair.LineMapping == nil {
		return Result{Passed: false, Message: "formalized repair output requires line_mapping"}
	}
	if repair.Dropped == nil {
		return Result{Passed: false, Message: "formalized repair output requires dropped"}
	}

	argumentResult := numberedArgumentSyntaxValidator{}.Validate(ctx, *repair.Argument)
	if !argumentResult.Passed {
		return Result{
			Passed:  false,
			Message: "formalized repair argument failed numbered_argument_syntax: " + argumentResult.Message,
			Details: map[string]any{"argument_validation": argumentResult.Details},
		}
	}
	if len(*repair.LineMapping) == 0 {
		return Result{Passed: false, Message: "formalized repair line_mapping must contain at least one entry"}
	}
	seenRepaired := make(map[int]bool, len(*repair.LineMapping))
	for index, mapping := range *repair.LineMapping {
		if mapping.RepairedLine < 1 {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].repaired_line must be a positive integer", index)}
		}
		if mapping.RepairedLine > argumentResult.Details["step_count"].(int) {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].repaired_line references line %d outside the argument", index, mapping.RepairedLine)}
		}
		if seenRepaired[mapping.RepairedLine] {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping has duplicate repaired_line %d", mapping.RepairedLine)}
		}
		seenRepaired[mapping.RepairedLine] = true
		if len(mapping.FromAsFoundLines) == 0 {
			return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].from_as_found_lines must contain at least one line", index)}
		}
		seenSources := make(map[int]bool, len(mapping.FromAsFoundLines))
		for _, line := range mapping.FromAsFoundLines {
			if line < 1 {
				return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].from_as_found_lines must contain only positive integers", index)}
			}
			if seenSources[line] {
				return Result{Passed: false, Message: fmt.Sprintf("line_mapping[%d].from_as_found_lines repeats line %d", index, line)}
			}
			seenSources[line] = true
		}
	}

	seenDropped := make(map[int]bool, len(*repair.Dropped))
	for index, dropped := range *repair.Dropped {
		if dropped.AsFoundLine < 1 {
			return Result{Passed: false, Message: fmt.Sprintf("dropped[%d].as_found_line must be a positive integer", index)}
		}
		if seenDropped[dropped.AsFoundLine] {
			return Result{Passed: false, Message: fmt.Sprintf("dropped has duplicate as_found_line %d", dropped.AsFoundLine)}
		}
		seenDropped[dropped.AsFoundLine] = true
		switch dropped.Kind {
		case DroppedContentKindPremise, DroppedContentKindConclusion, DroppedContentKindCitation:
		default:
			return Result{Passed: false, Message: fmt.Sprintf("dropped[%d].kind must be premise, conclusion, or citation", index)}
		}
		if strings.TrimSpace(dropped.Reason) == "" {
			return Result{Passed: false, Message: fmt.Sprintf("dropped[%d].reason must be non-empty", index)}
		}
	}

	return Result{
		Passed:       true,
		Message:      "output matches formalized_repair_output",
		Details:      map[string]any{"step_count": argumentResult.Details["step_count"], "mapping_count": len(*repair.LineMapping), "dropped_count": len(*repair.Dropped)},
		ParsedOutput: json.RawMessage(strings.TrimSpace(output)),
	}
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("extra JSON value")
		}
		return err
	}
	return nil
}
