package validation

import "testing"

// The motivating case: a binary constructive dilemma from a metaethics
// reconstruction that no earlier catalog certified.
func TestCatalogV12MotivatingConstructiveDilemma(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV12MotivatingConstructiveDilemma/metaethics": {"scope.clause_start", "move the operator into each clause"},
	}
	runV44 := func(t *testing.T, cases []v11Case) { runV11Strict(t, cases, refused) }

	runV44(t, []v11Case{
		{"metaethics", "1. Either moral facts depend on attitudes or moral facts do not depend on attitudes.\n2. If moral facts depend on attitudes, then robust attitude-independent realism is false.\n3. If moral facts do not depend on attitudes, then the proposed response-dependent account is false.\n4. Therefore, either robust attitude-independent realism is false or the proposed response-dependent account is false. [1,2,3]", "constructive dilemma"},
	})
}

func TestCatalogV12ConstructiveDilemma(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV12ConstructiveDilemma/ternary": {"scope.clause_start", "move the operator into each clause"},
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV12ConstructiveDilemma/ternary_conclusion_permuted": {"scope.clause_start", "move the operator into each clause"},
	}
	runV44 := func(t *testing.T, cases []v11Case) { runV11Strict(t, cases, refused) }

	runV44(t, []v11Case{
		{"binary", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. Therefore, either the road is wet or the road is icy. [1,2,3]", "constructive dilemma"},
		{"conditionals first", "1. If it rains, then the road is wet.\n2. If it snows, then the road is icy.\n3. Either it rains or it snows.\n4. Therefore, either the road is wet or the road is icy. [1,2,3]", "constructive dilemma"},
		{"citation order shuffled", "1. If it snows, then the road is icy.\n2. Either it rains or it snows.\n3. If it rains, then the road is wet.\n4. Therefore, either the road is wet or the road is icy. [3,1,2]", "constructive dilemma"},
		{"conclusion order swapped", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. Therefore, either the road is icy or the road is wet. [1,2,3]", "constructive dilemma"},
		{"ternary", "1. Either Ada stays or Ben stays or Cy stays.\n2. If Ada stays, then the lab opens.\n3. If Ben stays, then the shop opens.\n4. If Cy stays, then the cafe opens.\n5. Therefore, either the lab opens or the shop opens or the cafe opens. [1,2,3,4]", "constructive dilemma"},
		{"ternary conclusion permuted", "1. If Cy stays, then the cafe opens.\n2. Either Ada stays or Ben stays or Cy stays.\n3. If Ada stays, then the lab opens.\n4. If Ben stays, then the shop opens.\n5. Therefore, either the cafe opens or the lab opens or the shop opens. [1,2,3,4]", "constructive dilemma"},

		{"conclusion disjunct without conditional", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. Therefore, either the road is wet or the road is flooded. [1,2,3]", ""},
		{"conditional for wrong disjunct", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it hails, then the road is icy.\n4. Therefore, either the road is wet or the road is icy. [1,2,3]", ""},
		{"one conditional serves both disjuncts", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it rains, then the road is icy.\n4. Therefore, either the road is wet or the road is icy. [1,2,3]", ""},
		{"missing conditional", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. Therefore, either the road is wet or the road is icy. [1,2]", ""},
		{"ternary missing conditional", "1. Either Ada stays or Ben stays or Cy stays.\n2. If Ada stays, then the lab opens.\n3. If Ben stays, then the shop opens.\n4. Therefore, either the lab opens or the shop opens or the cafe opens. [1,2,3]", ""},
		{"ternary conclusion drops a disjunct", "1. Either Ada stays or Ben stays or Cy stays.\n2. If Ada stays, then the lab opens.\n3. If Ben stays, then the shop opens.\n4. If Cy stays, then the cafe opens.\n5. Therefore, either the lab opens or the shop opens. [1,2,3,4]", ""},
		{"affirming the consequents", "1. Either the road is wet or the road is icy.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. Therefore, either it rains or it snows. [1,2,3]", ""},
		{"conjunction is not disjunction", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. Therefore, the road is wet and the road is icy. [1,2,3]", ""},
		{"disjunction without either", "1. It is not the case that it rains or it snows.\n2. If it is not the case that it rains, then the road is dry.\n3. If it snows, then the road is icy.\n4. Therefore, either the road is dry or the road is icy. [1,2,3]", ""},
		{"extra unused premise", "1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. If it hails, then the road is white.\n5. Therefore, either the road is wet or the road is icy. [1,2,3,4]", ""},
	})
}

func TestCatalogV12SimpleDilemma(t *testing.T) {
	runV11Cases(t, []v11Case{
		{"binary", "1. Either it rains or it snows.\n2. If it rains, then the game is cancelled.\n3. If it snows, then the game is cancelled.\n4. Therefore, the game is cancelled. [1,2,3]", "simple dilemma"},
		{"order shuffled", "1. If it snows, then the game is cancelled.\n2. If it rains, then the game is cancelled.\n3. Either it rains or it snows.\n4. Therefore, the game is cancelled. [1,2,3]", "simple dilemma"},
		{"ternary", "1. Either it rains or it snows or it hails.\n2. If it rains, then the game is cancelled.\n3. If it snows, then the game is cancelled.\n4. If it hails, then the game is cancelled.\n5. Therefore, the game is cancelled. [1,2,3,4]", "simple dilemma"},

		{"consequents differ", "1. Either it rains or it snows.\n2. If it rains, then the game is cancelled.\n3. If it snows, then the game is postponed.\n4. Therefore, the game is cancelled. [1,2,3]", ""},
		{"conditional for wrong disjunct", "1. Either it rains or it snows.\n2. If it rains, then the game is cancelled.\n3. If it hails, then the game is cancelled.\n4. Therefore, the game is cancelled. [1,2,3]", ""},
		{"missing conditional", "1. Either it rains or it snows.\n2. If it rains, then the game is cancelled.\n3. Therefore, the game is cancelled. [1,2]", ""},
		{"affirming the consequent", "1. Either the game is cancelled or the game is cancelled.\n2. If it rains, then the game is cancelled.\n3. If it snows, then the game is cancelled.\n4. Therefore, it rains. [1,2,3]", ""},
	})
}

func TestCatalogV12DestructiveDilemma(t *testing.T) {
	// Explicit v44 refusal list; original case texts remain in the table below.
	refused := map[string]strictRefusal{
		// scope.clause_start: an inspected operand lacks the admitted clause-start shape.
		"TestCatalogV12DestructiveDilemma/ternary": {"scope.clause_start", "move the operator into each clause"},
	}
	runV44 := func(t *testing.T, cases []v11Case) { runV11Strict(t, cases, refused) }

	runV44(t, []v11Case{
		{"binary", "1. If it rains, then the road is wet.\n2. If it snows, then the road is icy.\n3. Either it is not the case that the road is wet or it is not the case that the road is icy.\n4. Therefore, either it is not the case that it rains or it is not the case that it snows. [1,2,3]", "destructive dilemma"},
		{"disjunction first, swapped", "1. Either it is not the case that the road is icy or it is not the case that the road is wet.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. Therefore, either it is not the case that it snows or it is not the case that it rains. [1,2,3]", "destructive dilemma"},
		{"ternary", "1. If Ada stays, then the lab opens.\n2. If Ben stays, then the shop opens.\n3. If Cy stays, then the cafe opens.\n4. Either it is not the case that the lab opens or it is not the case that the shop opens or it is not the case that the cafe opens.\n5. Therefore, either it is not the case that Ada stays or it is not the case that Ben stays or it is not the case that Cy stays. [1,2,3,4]", "destructive dilemma"},

		{"denying the antecedents", "1. If it rains, then the road is wet.\n2. If it snows, then the road is icy.\n3. Either it is not the case that it rains or it is not the case that it snows.\n4. Therefore, either it is not the case that the road is wet or it is not the case that the road is icy. [1,2,3]", ""},
		{"conclusion disjunct without conditional", "1. If it rains, then the road is wet.\n2. If it snows, then the road is icy.\n3. Either it is not the case that the road is wet or it is not the case that the road is icy.\n4. Therefore, either it is not the case that it rains or it is not the case that it hails. [1,2,3]", ""},
		{"crossed pairing", "1. If it rains, then the road is wet.\n2. If it snows, then the road is wet.\n3. Either it is not the case that the road is wet or it is not the case that the road is icy.\n4. Therefore, either it is not the case that it rains or it is not the case that it snows. [1,2,3]", ""},
		{"missing conditional", "1. If it rains, then the road is wet.\n2. Either it is not the case that the road is wet or it is not the case that the road is icy.\n3. Therefore, either it is not the case that it rains or it is not the case that it snows. [1,2]", ""},
		{"unnegated disjunct", "1. If it rains, then the road is wet.\n2. If it snows, then the road is icy.\n3. Either it is not the case that the road is wet or the road is icy.\n4. Therefore, either it is not the case that it rains or it is not the case that it snows. [1,2,3]", ""},
	})
}

func TestCatalogV12NearMissNamesDilemma(t *testing.T) {
	report, _ := CheckEntailmentGate("1. Either it rains or it snows.\n2. If it rains, then the road is wet.\n3. If it snows, then the road is icy.\n4. Therefore, either the road is wet or the road is flooded. [1,2,3]")
	step := report.Steps[len(report.Steps)-1]
	if step.Passed || step.NearestForm != "constructive dilemma" {
		t.Fatalf("step = %+v, want failing constructive dilemma near-miss", step)
	}
}
