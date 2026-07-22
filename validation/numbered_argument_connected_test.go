package validation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func parseForTest(t *testing.T, arg string) NumberedArgument {
	t.Helper()
	parsed, res := ParseNumberedArgument(arg, numberedArgumentOptions{})
	if !res.Passed {
		t.Fatalf("parse failed: %s", res.Message)
	}
	return parsed
}

func TestNumberedArgumentUnusedSteps(t *testing.T) {
	cases := []struct {
		name string
		arg  string
		want []int
	}{
		{
			name: "fully connected",
			arg:  "1. All men are mortal.\n2. Socrates is a man.\n3. Therefore, Socrates is mortal. [1,2]",
			want: []int{},
		},
		{
			name: "uncited premise",
			arg:  "1. All men are mortal.\n2. Socrates is a man.\n3. The sky is blue.\n4. Therefore, Socrates is mortal. [1,2]",
			want: []int{3},
		},
		{
			name: "two independent conclusions, every premise used",
			// A valid multi-conclusion argument: premise 1 feeds conclusion 2, premises
			// 4,5 feed conclusion 6. No premise is dead even though 1/2 don't feed the
			// final conclusion.
			arg:  "1. The box is red.\n2. Therefore, the box has a color. [1]\n3. Steam is water.\n4. All water is wet.\n5. Therefore, steam is wet. [3,4]",
			want: nil,
		},
		{
			name: "dead premise among multiple conclusions",
			arg:  "1. The box is red.\n2. The moon is far.\n3. Therefore, the box has a color. [1]\n4. Steam is water.\n5. All water is wet.\n6. Therefore, steam is wet. [4,5]",
			want: []int{2},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NumberedArgumentUnusedSteps(parseForTest(t, tc.arg))
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("unused = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNumberedArgumentSyntaxRejectsUnusedSteps(t *testing.T) {
	v := numberedArgumentSyntaxValidator{}
	// connected -> passes
	if res := v.Validate(context.Background(),
		"1. All men are mortal.\n2. Socrates is a man.\n3. Therefore, Socrates is mortal. [1,2]"); !res.Passed {
		t.Fatalf("connected argument rejected: %s", res.Message)
	}
	// dead premise -> fails, and reports the unused step
	res := v.Validate(context.Background(),
		"1. All men are mortal.\n2. Socrates is a man.\n3. The sky is blue.\n4. Therefore, Socrates is mortal. [1,2]")
	if res.Passed {
		t.Fatal("argument with an uncited premise was accepted")
	}
	unused, ok := res.Details["unused_steps"].([]int)
	if !ok || !reflect.DeepEqual(unused, []int{3}) {
		t.Fatalf("unused_steps detail = %v (ok=%v), want [3]", res.Details["unused_steps"], ok)
	}

	allowDisconnected, err := Compile(Spec{
		Kind:    KindBuiltin,
		Name:    BuiltinNumberedArgumentSyntax,
		Options: json.RawMessage(`{"allow_disconnected_steps":true}`),
	})
	if err != nil {
		t.Fatalf("compile allow_disconnected_steps validator: %v", err)
	}
	if res := allowDisconnected.Validate(context.Background(),
		"1. All men are mortal.\n2. Socrates is a man.\n3. The sky is blue.\n4. Therefore, Socrates is mortal. [1,2]"); !res.Passed {
		t.Fatalf("argument with an uncited premise rejected with allow_disconnected_steps: %s", res.Message)
	}
}
