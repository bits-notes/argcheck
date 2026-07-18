package validation

import (
	"context"
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
			name: "orphan intermediate conclusion drags its private premise out",
			// step 3 is a conclusion nothing cites; premise 1 feeds only step 3; the
			// final conclusion (5) reaches neither, so both 1 and 3 are unused.
			arg:  "1. The moon is made of rock.\n2. Socrates is a man.\n3. Therefore, the moon is a body. [1]\n4. All men are mortal.\n5. Therefore, Socrates is mortal. [2,4]",
			want: []int{1, 3},
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
}
