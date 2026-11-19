package validation

import "testing"

func TestCatalogV25LexicalOracle(t *testing.T) {
	cases := []struct {
		text string
		want *folFormula
	}{
		{"It rains", foAtom("it rains")},
		{"Snow melts", foAtom("snow melts")},
		{"If it snows, then it rains", fol("implies", foAtom("it snows"), foAtom("it rains"))},
		{"The FTC is active", foAtom("is/active", "ftc")},
		{"Japan's ministry is very active", foAtom("is/very active", "japan's ministry")},
		{"Alice sends blair into casey", foAtom("sends/into", "alice", "blair", "casey")},
		{"Every x is such that (x is very active) and (it rains)", v24Q("every", "$x", fol("and", foAtom("is/very active", "$x"), foAtom("it rains")))},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			fs, why := legacyFOLReadings(tc.text)
			if why != "" || len(fs) != 1 || !folSame(fs[0], tc.want) {
				t.Fatalf("unexpected reading: %s, %v", why, fs)
			}
			rendered, why := legacyFOLRender(tc.want)
			if why != "" {
				t.Fatal(why)
			}
			v25Conversion(t, tc.text, rendered)
			got, err := parseFOL(rendered)
			if err != nil {
				t.Fatal(err)
			}
			if finiteCountermodel([]*folFormula{tc.want}, got, 3) || finiteCountermodel([]*folFormula{got}, tc.want, 3) {
				t.Fatal("conversion countermodel")
			}
		})
	}
	for _, tc := range []struct {
		text  string
		valid bool
	}{
		{"If it rains, then it rains", true},
		{"Either it rains or it is not the case that it rains", true},
		{"If it rains, then snow melts", false},
		{"It rains and it is not the case that it rains", false},
	} {
		fs, why := legacyFOLReadings(tc.text)
		if why != "" || len(fs) != 1 {
			t.Fatal(tc.text, why)
		}
		rendered, why := legacyFOLRender(fs[0])
		if why != "" {
			t.Fatal(tc.text, why)
		}
		v25Conversion(t, tc.text, rendered)
		f, err := parseFOL(rendered)
		if err != nil {
			t.Fatal(err)
		}
		if finiteCountermodel(nil, f, 3) == tc.valid {
			t.Fatal(tc.text, "tautology verdict")
		}
	}
	if folSame(foAtom("it rains"), foAtom("rain falls")) {
		t.Fatal("paraphrase")
	}
	if !finiteCountermodel([]*folFormula{foAtom("it rains")}, foAtom("rain falls"), 3) {
		t.Fatal("opaque independence")
	}
}

func TestCatalogV25ZeroPlaceRules(t *testing.T) {
	v24Check(t, "1. Every x is such that it rains.\n2. Therefore, it rains. [1]")
	// explicit-variable FOL (no note id): v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit).
	assertStrictRefusal(t, "1. It rains.\n2. Therefore, some x is such that it rains. [1]", "explicit_variable_fol", "bind variables inside is such that")
	// scope.clause_start: v44 strict: lowercase constant / letter eigenvariable is not name evidence (B limit).
	assertStrictRefusal(t, "1. Every x is such that (x is red) and (snow melts).\n2. Therefore, (alice is red) and (snow melts). [1]\n3. Therefore, snow melts. [2]", "scope.clause_start", "move the operator into each clause")
	for _, text := range []string{"Every animal sleeps", "Every cat purrs", "either it rains"} {
		fs, why := legacyFOLReadings(text)
		if why != "" || len(fs) != 1 || len(fs[0].terms) != 0 {
			t.Fatal(text, why)
		}
		rendered, why := legacyFOLRender(fs[0])
		if why != "" {
			t.Fatal(text, why)
		}
		v25Conversion(t, text, rendered)
	}
}

func TestCatalogV25OpaqueScopeGuard(t *testing.T) {
	for _, text := range []string{"every x is such that", "every x is such that ()", "every x is such that x would admire alice"} {
		if _, err := parseFOL(text); err == nil {
			t.Fatal("malformed scope accepted", text)
		}
	}
}

func TestCatalogV25ReadableFallbackPrecedence(t *testing.T) {
	_, report := CheckEntailmentGate("1. Either it rains or it rains and it snows.\n2. Therefore, it snows. [1]")
	if report.Passed {
		t.Fatal("FOL fallback chose a mixed-connective reading")
	}
}
