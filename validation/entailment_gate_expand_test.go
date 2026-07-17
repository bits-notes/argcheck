package validation

import (
	"reflect"
	"testing"
)

func TestExpandNumberedArgumentBridges(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		step       int
		want       string
		wantRecord BridgeExpansion
	}{
		{
			name:  "single reference",
			input: "1. Rain falls.\n2. If [1], then streets get wet.\n3. Therefore, streets get wet. [1,2]",
			step:  2,
			want:  "If it is the case that Rain falls, then streets get wet",
			wantRecord: BridgeExpansion{
				Step: 2, Original: "If [1], then streets get wet", Expanded: "If it is the case that Rain falls, then streets get wet",
			},
		},
		{
			name:  "comma list",
			input: "1. Rain falls.\n2. Wind blows.\n3. If [1,2], then the storm arrives.\n4. Therefore, the storm arrives. [1,3]",
			step:  3,
			want:  "If it is the case that Rain falls, and it is the case that Wind blows, then the storm arrives",
			wantRecord: BridgeExpansion{
				Step: 3, Original: "If [1,2], then the storm arrives", Expanded: "If it is the case that Rain falls, and it is the case that Wind blows, then the storm arrives",
			},
		},
		{
			name:  "and list",
			input: "1. Rain falls.\n2. Wind blows.\n3. If [1] and [2], then the storm arrives.\n4. Therefore, the storm arrives. [1,3]",
			step:  3,
			want:  "If it is the case that Rain falls, and it is the case that Wind blows, then the storm arrives",
			wantRecord: BridgeExpansion{
				Step: 3, Original: "If [1] and [2], then the storm arrives", Expanded: "If it is the case that Rain falls, and it is the case that Wind blows, then the storm arrives",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			argument, result := ParseNumberedArgument(test.input, numberedArgumentOptions{})
			if !result.Passed {
				t.Fatalf("parse failed: %s", result.Message)
			}
			expanded, records, err := ExpandNumberedArgumentBridges(argument)
			if err != nil {
				t.Fatalf("expand: %v", err)
			}
			if got := expanded.Steps[test.step-1].Statement; got != test.want {
				t.Fatalf("expanded statement = %q, want %q", got, test.want)
			}
			if !reflect.DeepEqual(records, []BridgeExpansion{test.wantRecord}) {
				t.Fatalf("records = %#v, want %#v", records, []BridgeExpansion{test.wantRecord})
			}
		})
	}
}

func TestExpandNumberedArgumentBridgesChainedAndIdempotent(t *testing.T) {
	input := "1. Rain falls.\n2. If [1], then streets get wet.\n3. If [2], then traffic slows.\n4. Therefore, traffic slows. [2,3]"
	argument, result := ParseNumberedArgument(input, numberedArgumentOptions{})
	if !result.Passed {
		t.Fatalf("parse failed: %s", result.Message)
	}

	expanded, records, err := ExpandNumberedArgumentBridges(argument)
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	want := "If it is the case that If it is the case that Rain falls, then streets get wet, then traffic slows"
	if got := expanded.Steps[2].Statement; got != want {
		t.Fatalf("chained statement = %q, want %q", got, want)
	}
	if len(records) != 2 {
		t.Fatalf("expansion count = %d, want 2", len(records))
	}

	again, secondRecords, err := ExpandNumberedArgumentBridges(expanded)
	if err != nil {
		t.Fatalf("second expansion: %v", err)
	}
	if len(secondRecords) != 0 {
		t.Fatalf("second expansion records = %#v, want none", secondRecords)
	}
	if !reflect.DeepEqual(again, expanded) {
		t.Fatal("expansion is not idempotent")
	}
}
