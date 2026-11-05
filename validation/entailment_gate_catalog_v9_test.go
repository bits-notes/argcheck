package validation

import (
	"fmt"
	"strings"
	"testing"
)

func TestCatalogV9SurfaceForms(t *testing.T) {
	tests := []struct {
		name, argument string
		want           bool
	}{
		{"if comma", "1. If interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", true},
		{"when", "1. When interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", true},
		{"whenever", "1. Whenever interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", true},
		{"final if", "1. The lamp is lit if the switch is up.\n2. The switch is up.\n3. Therefore, the lamp is lit. [1,2]", true},
		{"final if rain", "1. The ground is wet if it rains.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]", true},
		{"final if alarm", "1. The alarm sounds if the door is open.\n2. The door is open.\n3. Therefore, the alarm sounds. [1,2]", true},
		{"symbolic final if", "1. B if A.\n2. A.\n3. Therefore, B. [1,2]", true},
		{"conditional identity", "1. When the switch is up, the lamp is lit.\n2. Therefore, if the switch is up, then the lamp is lit. [1]", true},
		{"elision identity", "1. Either the meeting is on Tuesday or on Thursday.\n2. Therefore, either the meeting is on Tuesday or the meeting is on Thursday. [1]", true},
		{"only if converse", "1. The switch is up only if the lamp is lit.\n2. The lamp is lit.\n3. Therefore, the switch is up. [1,2]", false},
		{"curious interrogative", "1. Ada is curious if the switch is up.\n2. The switch is up.\n3. Therefore, Ada is curious. [1,2]", false},
		{"contracted clause negation", "1. If the switch is up, then the lamp is lit.\n2. It isn't the case that the lamp is lit.\n3. Therefore, it is not the case that the switch is up. [1,2]", true},
		{"no main verb scope guess", "1. Ada doesn't always sing.\n2. Therefore, it is not the case that Ada sings. [1]", false},
		{"when affirming consequent", "1. When interest rates rise, mortgages become more expensive.\n2. Mortgages become more expensive.\n3. Therefore, interest rates rise. [1,2]", false},
		{"unless", "1. Unless interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", false},
		{"because", "1. Because interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", false},
		{"since", "1. Since interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", false},
		{"after", "1. After interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", false},
		{"before", "1. Before interest rates rise, mortgages become more expensive.\n2. Interest rates rise.\n3. Therefore, mortgages become more expensive. [1,2]", false},
		{"interrogative if", "1. Ada wonders if the switch is up.\n2. The switch is up.\n3. Therefore, Ada wonders. [1,2]", false},
		{"unsure if", "1. Ada is unsure if the switch is up.\n2. The switch is up.\n3. Therefore, Ada is unsure. [1,2]", false},
		{"embedded if", "1. The report says the lamp is lit if the switch is up.\n2. The switch is up.\n3. Therefore, the report says the lamp is lit. [1,2]", false},
		{"multiple commas", "1. When the switch is up, according to Ada, the lamp is lit.\n2. The switch is up.\n3. Therefore, according to Ada, the lamp is lit. [1,2]", false},
		{"meeting elision", "1. Either the meeting is on Tuesday or on Thursday.\n2. It is not the case that the meeting is on Tuesday.\n3. Therefore, the meeting is on Thursday. [1,2]", true},
		{"adjective elision", "1. Either the lamp is red or green.\n2. It is not the case that the lamp is red.\n3. Therefore, the lamp is green. [1,2]", true},
		{"elision introduction", "1. The lamp is green.\n2. Therefore, either the lamp is red or green. [1]", true},
		{"ambiguous subject", "1. Either Ada or Ben is ready or tired.\n2. It is not the case that Ada or Ben is ready.\n3. Therefore, Ada or Ben is tired. [1,2]", false},
		{"embedded subject", "1. Either Ada thinks the lamp is red or green.\n2. It is not the case that Ada thinks the lamp is red.\n3. Therefore, Ada thinks the lamp is green. [1,2]", false},
		{"quantified subject", "1. Either every lamp is red or green.\n2. It is not the case that every lamp is red.\n3. Therefore, every lamp is green. [1,2]", false},
		{"negated copula", "1. Either the lamp is not red or green.\n2. It is not the case that the lamp is not red.\n3. Therefore, the lamp is not green. [1,2]", false},
		{"contraction does not affirm", "1. Ada doesn't sing.\n2. Therefore, Ada sings. [1]", false},
		{"so degree", "1. Many birds fly.\n2. So many birds fly. [1]", false},
		{"quoted contraction", "1. The password is \"can't\".\n2. Therefore, the password is \"can not\". [1]", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tt.argument)
			if report.Passed != tt.want || result.Passed != tt.want {
				t.Fatalf("want %v: %+v; %+v", tt.want, report, result)
			}
		})
	}
}

func TestCatalogV9Contractions(t *testing.T) {
	for _, pair := range [][2]string{{"doesn't", "does not"}, {"isn't", "is not"}, {"didn't", "did not"}, {"won't", "will not"}, {"can't", "can not"}, {"cannot", "can not"}, {"can’t", "cannot"}, {"aren't", "are not"}, {"don't", "do not"}, {"wasn't", "was not"}, {"weren't", "were not"}, {"hasn't", "has not"}, {"haven't", "have not"}, {"hadn't", "had not"}, {"wouldn't", "would not"}, {"couldn't", "could not"}, {"shouldn't", "should not"}, {"mustn't", "must not"}, {"needn't", "need not"}} {
		t.Run(pair[0], func(t *testing.T) {
			subject, complement := "Ada", "agree"
			switch pair[1] {
			case "is not", "was not":
				complement = "ready"
			case "are not", "were not":
				subject, complement = "They", "ready"
			case "do not":
				subject = "They"
			case "have not":
				subject, complement = "They", "agreed"
			case "has not", "had not":
				complement = "agreed"
			}
			for _, direction := range [][2]string{pair, {pair[1], pair[0]}} {
				argument := fmt.Sprintf("1. If %s %s %s, then Ben is ready.\n2. %s %s %s.\n3. Therefore, Ben is ready. [1,2]", subject, direction[0], complement, subject, direction[1], complement)
				report, result := CheckEntailmentGate(argument)
				if !report.Passed || !result.Passed {
					t.Fatalf("%s: %+v", argument, result)
				}
			}
		})
	}
}

func TestCatalogV9Markers(t *testing.T) {
	for _, marker := range []string{"So", "Thus", "Hence", "Therefore", "So,", "Thus,", "Hence,", "Therefore,"} {
		t.Run(marker, func(t *testing.T) {
			argument := fmt.Sprintf("1. The lamp is lit.\n2. %s the lamp is lit. [1]\n3. %s the lamp is lit. [2]", marker, marker)
			for _, newline := range []string{"\n", "\r\n", "\r"} {
				report, result := CheckEntailmentGate(strings.ReplaceAll(argument, "\n", newline))
				if !report.Passed || !result.Passed {
					t.Fatalf("%+v", result)
				}
			}
		})
	}
}
