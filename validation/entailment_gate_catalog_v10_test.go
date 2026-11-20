package validation

import "testing"

func TestCatalogV10NaryConjunctionIntroduction(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV10NaryConjunctionIntroduction/oxford_comma": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV10NaryConjunctionIntroduction/no_oxford_comma": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV10NaryConjunctionIntroduction/chained_and": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV10NaryConjunctionIntroduction/any_order": {"scope.clause_start", "move the operator into each clause"},
	}

	tests := []struct {
		name, argument string
		want           bool
	}{
		{"oxford comma", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, rain falls, wind blows, and snow melts. [1,2,3]", true},
		{"no oxford comma", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, rain falls, wind blows and snow melts. [1,2,3]", true},
		{"chained and", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, rain falls and wind blows and snow melts. [1,2,3]", true},
		{"any order", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, snow melts, rain falls, and wind blows. [1,2,3]", true},
		{"premise with internal list", "1. The firm has no employee, no manager, and no building in common with the firm's founder.\n2. The firm moved twice.\n3. The charter survives.\n4. Therefore, the firm has no employee, no manager, and no building in common with the firm's founder, the firm moved twice, and the charter survives. [1,2,3]", true},
		{"missing conjunct", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, rain falls and wind blows. [1,2,3]", false},
		{"extra conjunct", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, rain falls, wind blows, snow melts, and ice cracks. [1,2,3]", false},
		{"reworded conjunct", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, rain falls, wind blows, and the snow melts. [1,2,3]", false},
		{"disjunction is not conjunction", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Therefore, rain falls, wind blows, or snow melts. [1,2,3]", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if listedStrictRefusal(t, tc.argument, refused) {
				return
			}
			report, detail := CheckEntailmentGate(tc.argument)
			if report.Passed != tc.want {
				t.Fatalf("passed = %v, want %v: %s", report.Passed, tc.want, detail.Message)
			}
		})
	}
}
