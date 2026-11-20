package validation

import (
	"strings"
	"testing"
)

func TestV46Values(t *testing.T) {
	for _, tc := range []struct {
		value, verb string
		pass        bool
	}{
		{"subgoals", "are", true}, {"humanity", "is", true}, {"advanced machines", "are", true},
		{"the pump", "is", true}, {"the main valve", "is", true}, {"the death of the child", "is", true},
		{"lots", "are", false}, {"plenty", "are", false}, {"dozens", "are", false}, {"hundreds", "are", false},
		{"ages", "are", false}, {"droves", "are", false}, {"quokkas", "are", false}, {"numerous robots", "are", false},
		{"robots", "is", false}, {"water", "are", false}, {"the majority of voters", "is", false},
		{"the lot", "is", false}, {"the bulk", "is", false},
	} {
		t.Run(tc.value+"_"+tc.verb, func(t *testing.T) {
			text := "1. If X " + tc.verb + " dangerous, then X should be contained.\n2. " + tc.value + " " + tc.verb + " dangerous.\n3. Therefore, " + tc.value + " should be contained. [1,2]"
			r, res := CheckEntailmentGate(text)
			if r.Passed != tc.pass {
				t.Fatalf("pass=%v want=%v: %s", r.Passed, tc.pass, res.Message)
			}
		})
	}
}

func TestV46WitnessAndIdentity(t *testing.T) {
	for _, text := range []string{
		"1. If X is in the cup and X is in the bottle, then X is in both vessels.\n2. Water is in the cup.\n3. Water is in the bottle.\n4. Therefore, water is in both vessels. [1,2,3]",
		"1. If X is in the cup, then X is together.\n2. Water is in the cup.\n3. Therefore, water is together. [1,2]",
		"1. Robots are identical to Alpha.\n2. Alpha is identical to Beta.\n3. Therefore, robots are identical to Beta. [1,2]",
		"1. Robots are red.\n2. Therefore, something is red. [1]",
	} {
		if r, res := CheckEntailmentGate(text); r.Passed {
			t.Fatal("certified", text, res.Message)
		}
	}
}

func TestV46AgreementAndShape(t *testing.T) {
	p := allStrictRules()
	for _, v := range []string{"robots", "advanced machines", "small advanced machines", "water"} {
		if !p.singularValue(v) || p.without("slot.plural_mass_value").singularValue(v) {
			t.Error(v)
		}
	}
	for _, v := range []string{"very small advanced machines", "robots of Mars", "robots,", "Bob's robots", "full robots", "the flibbertigibbet", "the group of robots"} {
		if p.singularValue(v) {
			t.Error("accepted", v)
		}
	}
	for _, verb := range strings.Fields("is was has does runs works requires") {
		template, _ := p.schematicTemplate("X " + verb + " dangerous")
		if p.termValueAt([]rune(template), 0, "robots") {
			t.Error(verb)
		}
	}
	for _, verb := range strings.Fields("are were have do") {
		template, _ := p.schematicTemplate("X " + verb + " dangerous")
		if p.termValueAt([]rune(template), 0, "water") {
			t.Error(verb)
		}
	}
}

// These values exercise the new path end to end, rather than just membership.
func TestV46TwentyNewValues(t *testing.T) {
	accepted := strings.Split("subgoals|humans|machines|robots|systems|agents|humanity|intelligence|evidence|water|information|knowledge|research|software|advanced machines|aligned agents|trained robots|informed humans|reliable systems|dangerous subgoals", "|")
	refused := strings.Split("the heap|the pile|the stack|the ton|the load|the lack|the range|the variety|the gallon|the kind|the sort|the type|the unit|the measure|the measurement|the mile|the pound|the dollar|the year|the hour", "|")
	for _, value := range accepted {
		verb := "are"
		if strings.Contains("|humanity|intelligence|evidence|water|information|knowledge|research|software|", "|"+value+"|") {
			verb = "is"
		}
		text := "1. If X " + verb + " dangerous, then X should be contained.\n2. " + value + " " + verb + " dangerous.\n3. Therefore, " + value + " should be contained. [1,2]"
		if r, res := CheckEntailmentGate(text); !r.Passed {
			t.Errorf("%s: %s", value, res.Message)
		}
	}
	for _, value := range refused {
		text := "1. If X is dangerous, then X should be contained.\n2. " + value + " is dangerous.\n3. Therefore, " + value + " should be contained. [1,2]"
		if r, _ := CheckEntailmentGate(text); r.Passed {
			t.Errorf("certified %s", value)
		}
	}
	t.Log("20 newly accepted: " + strings.Join(accepted, ", "))
	t.Log("20 newly refused: " + strings.Join(refused, ", "))
}

func TestV46PredicateDescriptionAndWitnessBoundaries(t *testing.T) {
	for _, text := range []string{
		"1. If X is transparent or durable, then X is useful.\n2. The machine is transparent or durable.\n3. Therefore, the machine is useful. [1,2]",
		"1. If X is useful, then X is valuable.\n2. The durable transparent machine is useful.\n3. Therefore, the durable transparent machine is valuable. [1,2]",
	} {
		if r, res := CheckEntailmentGate(text); !r.Passed {
			t.Error(res.Message)
		}
	}
	for _, text := range []string{
		"1. If X is in the cup and X is in the bottle, then X is valuable.\n2. Water is in the cup.\n3. Water is in the bottle.\n4. Therefore, water is valuable. [1,2,3]",
		"1. If X is in the cup and X is in the bottle, then X is valuable.\n2. Water is in the cup and water is in the bottle.\n3. Therefore, water is valuable. [1,2]",
		"1. If X is valuable, then X are dangerous.\n2. Water is valuable.\n3. Therefore, water are dangerous. [1,2]",
		"1. If X is identical to Alpha, then X is valuable.\n2. Water is identical to Alpha.\n3. Therefore, water is valuable. [1,2]",
		"1. If Alpha is identical to X, then X is valuable.\n2. Alpha is identical to water.\n3. Therefore, water is valuable. [1,2]",
	} {
		if r, _ := CheckEntailmentGate(text); r.Passed {
			t.Error("certified", text)
		}
	}
}
