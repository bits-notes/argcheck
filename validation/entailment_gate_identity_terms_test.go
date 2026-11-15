package validation

import "testing"

// Found by the blind completeness review (Nov 2026): the identity rules took
// quantifier phrases as names. The v28 whole-term register compared
// capitalized words ("Nobody", "Some") against a lowercase word list,
// identity transitivity had no term check, and the readable FOL bridge stripped
// "a"/"an" so "a cat" became the constant cat. Each rejected case is invalid and
// was accepted by v28.1.
func TestEntailmentGateIdentityTermsMustRefer(t *testing.T) {
	tests := []struct {
		name     string
		argument string
		want     bool
	}{
		{"substitution through nobody", "1. Nobody is identical to Felix.\n2. Nobody is black.\n3. Therefore, Felix is black. [2,1]", false},
		{"substitution through some", "1. Some cat is identical to Felix.\n2. Some cat is black.\n3. Therefore, Felix is black. [2,1]", false},
		{"substitution through a", "1. Tom admires a cat.\n2. A cat is identical to Felix.\n3. Therefore, Tom admires Felix. [1,2]", false},
		{"transitivity through some", "1. Tom is identical to some cat.\n2. Some cat is identical to Felix.\n3. Therefore, Tom is identical to Felix. [1,2]", false},
		{"transitivity through a", "1. Tom is identical to a cat.\n2. A cat is identical to Felix.\n3. Therefore, Tom is identical to Felix. [1,2]", false},
		{"transitivity through no", "1. Tom is identical to no cat.\n2. No cat is identical to Felix.\n3. Therefore, Tom is identical to Felix. [1,2]", false},
		{"transitivity through every", "1. Tom is identical to every cat.\n2. Every cat is identical to Felix.\n3. Therefore, Tom is identical to Felix. [1,2]", false},

		{"transitivity with names", "1. Tom is identical to Felix.\n2. Felix is identical to Garfield.\n3. Therefore, Tom is identical to Garfield. [1,2]", true},
		{"transitivity with a definite description", "1. Tom is identical to the cat.\n2. The cat is identical to Felix.\n3. Therefore, Tom is identical to Felix. [1,2]", true},
		{"substitution with names", "1. Tom is black.\n2. Tom is identical to Felix.\n3. Therefore, Felix is black. [1,2]", true},
		{"indefinite outside identity", "1. A cat is black.\n2. A cat is white.\n3. Therefore, a cat is black and a cat is white. [1,2]", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report, _ := CheckEntailmentGate(tc.argument)
			if report.Passed != tc.want {
				t.Fatalf("accepted=%v, want %v\n%s", report.Passed, tc.want, tc.argument)
			}
		})
	}
}

// Known reading limit, not fixed: an intensional transitive verb is read as an
// ordinary relation, so identity substitution goes through it.
func TestEntailmentGateIntensionalVerbGap(t *testing.T) {
	argument := "1. Lois is looking for Superman.\n2. Superman is identical to Clark.\n3. Therefore, Lois is looking for Clark. [1,2]"
	if report, _ := CheckEntailmentGate(argument); !report.Passed {
		t.Fatal("intensional-verb gap is now rejected; move it into a rejection test")
	}
}
