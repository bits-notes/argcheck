package validation

import "testing"

func TestEntailmentGateV29Unless(t *testing.T) {
	tests := []struct {
		name     string
		argument string
		want     bool
	}{
		{"elimination", "1. The alarm sounds unless the battery is dead.\n2. It is not the case that the battery is dead.\n3. Therefore, the alarm sounds. [1,2]", true},
		{"elimination, unless first", "1. Unless the battery is dead, the alarm sounds.\n2. It is not the case that the battery is dead.\n3. Therefore, the alarm sounds. [1,2]", true},
		{"elimination, statute", "1. Any filing received after the statutory deadline is untimely unless the court granted an extension.\n2. It is not the case that the court granted an extension.\n3. Therefore, any filing received after the statutory deadline is untimely. [1,2]", true},
		{"conversion to conditional", "1. The alarm sounds unless the battery is dead.\n2. Therefore, if it is not the case that the battery is dead, then the alarm sounds. [1]", true},
		{"conversion from conditional", "1. If it is not the case that the battery is dead, then the alarm sounds.\n2. Therefore, the alarm sounds unless the battery is dead. [1]", true},

		{"no biconditional reading: Q gives not P", "1. The alarm sounds unless the battery is dead.\n2. The battery is dead.\n3. Therefore, it is not the case that the alarm sounds. [1,2]", false},
		{"no exclusive reading: P gives not Q", "1. The alarm sounds unless the battery is dead.\n2. The alarm sounds.\n3. Therefore, it is not the case that the battery is dead. [1,2]", false},
		{"no converse conditional", "1. The alarm sounds unless the battery is dead.\n2. Therefore, if the battery is dead, then it is not the case that the alarm sounds. [1]", false},
		{"pronoun in unless-clause", "1. Every student passes unless he fails the exam.\n2. It is not the case that he fails the exam.\n3. Therefore, every student passes. [1,2]", false},
		{"one conjunct of a compound unless-clause", "1. The alarm sounds unless the battery is dead and the backup is off.\n2. It is not the case that the battery is dead.\n3. Therefore, the alarm sounds. [1,2]", false},
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
