package validation

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func v26Argument(ps []string, c string) string {
	var lines, refs []string
	for i, p := range ps {
		lines = append(lines, fmt.Sprintf("%d. %s.", i+1, p))
		refs = append(refs, fmt.Sprint(i+1))
	}
	return strings.Join(append(lines, fmt.Sprintf("%d. Therefore, %s. [%s]", len(ps)+1, c, strings.Join(refs, ","))), "\n")
}
func v26Check(t *testing.T, ps []string, c string, valid bool, domain int) {
	t.Helper()
	arg := v26Argument(ps, c)
	_, r := CheckEntailmentGate(arg)
	if r.Passed != valid {
		t.Fatalf("%s\n%s", arg, r.Message)
	}
	var fs []*folFormula
	for _, p := range ps {
		f, e := countParse(p)
		if e != nil {
			t.Fatal(e)
		}
		fs = append(fs, f)
	}
	f, e := countParse(c)
	if e != nil {
		t.Fatal(e)
	}
	if finiteCountermodel(fs, f, domain) == valid {
		t.Fatalf("oracle disagrees (domains 1-%d): %s", domain, arg)
	}
}

// Independently assemble the identity semantics, with a different association
// of conjunctions, rather than using the production counting constructor.
func v26OracleDefinition(kind string, n int, body *folFormula) *folFormula {
	lower := func(k int) *folFormula {
		var fs []*folFormula
		for i := 0; i < k; i++ {
			v := fmt.Sprintf("$oracle%d", i)
			fs = append(fs, folSub(body, "$count", v))
			for j := 0; j < i; j++ {
				fs = append(fs, foNot(foEqual(v, fmt.Sprintf("$oracle%d", j))))
			}
		}
		f := foAnd(fs...)
		for i := k - 1; i >= 0; i-- {
			f = v24Q("some", fmt.Sprintf("$oracle%d", i), f)
		}
		return f
	}
	switch kind {
	case "at least":
		return lower(n)
	case "at most":
		return foNot(lower(n + 1))
	default:
		return foAnd(lower(n), foNot(lower(n+1)))
	}
}
func TestCatalogV26Definitions(t *testing.T) {
	start := time.Now()
	for n, w := range []string{"one", "two", "three", "four", "five"} {
		n++
		for _, kind := range []string{"at least", "at most", "exactly"} {
			s := kind + " " + w + " labs signed the pledge"
			f, e := countParse(s)
			if e != nil {
				t.Fatal(e)
			}
			body := foAnd(foAtom("is/lab", "$count"), foAtom("signed", "$count", "pledge"))
			expected := v26OracleDefinition(kind, n, body)
			if finiteCountermodel([]*folFormula{f}, expected, 4) || finiteCountermodel([]*folFormula{expected}, f, 4) {
				t.Fatal("definition", s)
			}
			digits := fmt.Sprintf("%s %d labs signed the pledge", kind, n)
			v26Check(t, []string{s}, digits, true, 4)
			v26Check(t, []string{s}, f.Render(), true, 4)
			v26Check(t, []string{f.Render()}, s, true, 4)
		}
	}
	t.Logf("45 valid conversions; independently specified identity definitions; exhaustive domains 1-4; %s", time.Since(start))
}
func TestCatalogV26Exercises(t *testing.T) {
	start := time.Now()
	valid := 0
	for n := 1; n <= 5; n++ {
		lo := fmt.Sprintf("At least %d labs are red", n)
		hi := fmt.Sprintf("At most %d labs are red", n)
		ex := fmt.Sprintf("Exactly %d labs are red", n)
		v26Check(t, []string{ex}, lo, true, 4)
		v26Check(t, []string{ex}, hi, true, 4)
		v26Check(t, []string{lo, hi}, ex, true, 4)
		v26Check(t, []string{fmt.Sprintf("Exactly %d labs", n)}, fmt.Sprintf("At least %d labs", n), true, 4)
		valid += 4
	}
	cases := []struct {
		ps []string
		c  string
	}{
		{[]string{"At least two labs signed the pledge", "At most one lab that signed the pledge is American"}, "Some lab that signed the pledge is not American"},
		{[]string{"At least three students sent blair to casey", "At most two students that sent blair to casey are red"}, "Some student that sent blair to casey is not red"},
		{[]string{"At least three labs are red", "At most two labs are red"}, "Contradiction"},
		{[]string{"Exactly one person wrote Waverley", "Scott is a person and Scott wrote Waverley"}, "No person other than Scott wrote Waverley"},
		{[]string{"Exactly one student admires blair", "Ada is a student and Ada admires blair"}, "No student other than Ada admires blair"},
		{[]string{"Exactly one student", "Ada is a student"}, "No student other than Ada is a student"},
		{[]string{"Ada is a student"}, "At least one student"},
		{[]string{"Only Ada is a student"}, "Ada is a student and no student other than Ada is a student"},
		{[]string{"Ada is a student and no student other than Ada is a student"}, "Only Ada is a student"},
	}
	for _, tc := range cases {
		v26Check(t, tc.ps, tc.c, true, 4)
		valid++
	}
	t.Logf("%d fresh valid flat exercises; exhaustive domains 1-4; %s", valid, time.Since(start))
}
func TestCatalogV26InvalidOracle(t *testing.T) {
	start := time.Now()
	cases := []struct {
		ps []string
		c  string
	}{
		{[]string{"At least one lab is red"}, "At most one lab is red"},
		{[]string{"At most one lab is red"}, "At least one lab is red"},
		{[]string{"Ada is a student and Ada admires blair"}, "No student other than Ada admires blair"},
		{[]string{"At least one lab is red", "At most one lab is red"}, "Exactly two labs are red"},
		{[]string{"Exactly two labs are red"}, "At least three labs are red"},
		{[]string{"Exactly three labs are red"}, "At most two labs are red"},
		{[]string{"At least one lab is red", "At most one lab is red"}, "Contradiction"},
		{[]string{"At least two labs signed the pledge", "At most two labs that signed the pledge are American"}, "Some lab that signed the pledge is not American"},
		{[]string{"Exactly one person wrote Waverley", "Scott wrote Waverley"}, "No person other than Scott wrote Waverley"},
	}
	for _, tc := range cases {
		v26Check(t, tc.ps, tc.c, false, 4)
	}
	t.Logf("9 invalid exercises, each with countermodel in domains 1-4 (b1 in domain 2); %s", time.Since(start))
}
func TestCatalogV26Fitch(t *testing.T) {
	proofs := []string{
		`1. Exactly one student admires blair.
| 2. Suppose Ada is a student and Ada admires blair.
| 3. Therefore, no student other than Ada admires blair. [1,2]
4. Therefore, if Ada is a student and Ada admires blair then no student other than Ada admires blair. [2-3]`,
		`1. At most one lab is red.
| 2. Suppose at least two labs are red.
| 3. Therefore, contradiction. [1,2]
4. Therefore, it is not the case that at least two labs are red. [2-3]`,
	}
	for _, arg := range proofs {
		_, r := CheckEntailmentGate(arg)
		if !r.Passed {
			t.Fatal(r.Message)
		}
	}
	// Local subproof obligations plus discharged conclusions, independently
	// checked as closed entailments (the assumptions are explicit premises).
	v26Check(t, []string{"Exactly one student admires blair", "Ada is a student and Ada admires blair"}, "No student other than Ada admires blair", true, 4)
	v26Check(t, []string{"At most one lab is red", "At least two labs are red"}, "Contradiction", true, 4)
	for _, tc := range []struct{ p, c string }{{"Exactly one student admires blair", "if Ada is a student and Ada admires blair then no student other than Ada admires blair"}, {"At most one lab is red", "it is not the case that at least two labs are red"}} {
		p, _ := countParse(tc.p)
		c, e := countParse(tc.c)
		if e != nil || finiteCountermodel([]*folFormula{p}, c, 4) {
			t.Fatal("Fitch oracle", e)
		}
	}
}
func TestCatalogV26Bound(t *testing.T) {
	for _, n := range []string{"6", "six", "10", "zero", "0"} {
		_, r := CheckEntailmentGate(v26Argument([]string{"Exactly " + n + " labs are red"}, "At least "+n+" labs are red"))
		if r.Passed || !strings.Contains(r.Message, "bound is 5") {
			t.Fatal(r.Message)
		}
	}
	if EntailmentGateCatalogVersion != "v30" {
		t.Fatal(EntailmentGateCatalogVersion)
	}
}

func TestCatalogV26B1Membership(t *testing.T) {
	// Keep the specification conflict independently reproducible: the missing
	// class premise is essential, even though ordinary prose presupposes it.
	body := foAnd(foAtom("is/person", "$count"), foAtom("wrote", "$count", "waverley"))
	p := v26OracleDefinition("exactly", 1, body)
	q := foAtom("wrote", "scott", "waverley")
	c := v24Q("every", "$count", fol("implies", body, foEqual("$count", "scott")))
	if finiteCountermodel([]*folFormula{p, q}, c, 1) || !finiteCountermodel([]*folFormula{p, q}, c, 2) {
		t.Fatal("b1 must first have a countermodel at domain 2")
	}
	if finiteCountermodel([]*folFormula{p, q, foAtom("is/person", "scott")}, c, 4) {
		t.Fatal("membership repair")
	}
}
