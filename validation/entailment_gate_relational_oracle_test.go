package validation

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// This oracle deliberately owns its grammar, formulas, and truth definition.
// It does not call relationCandidates, relationNormal, or any gate matcher to
// interpret a sentence. Domains 1..3 are exhaustive but bounded, not a proof.
type relOracleTerm struct {
	kind  byte
	class int
} // a,b names; A,E,N quantifiers
func (t relOracleTerm) text() string {
	if t.kind == 'a' {
		return "Alice"
	}
	if t.kind == 'b' {
		return "Bob"
	}
	q := map[byte]string{'A': "every", 'E': "some", 'N': "no"}[t.kind]
	return q + " " + []string{"student", "teacher"}[t.class]
}

type relOracleFormula struct {
	terms                   [2]relOracleTerm
	order                   [2]int
	count                   int
	outerNot, innerNot      bool
	member                  bool
	memberClass, memberName int
}
type relOracleSurface struct {
	text     string
	readings []relOracleFormula
}
type relOracleModel struct{ size, student, teacher, relation, alice, bob int }

func (m relOracleModel) unary(class, name int) bool {
	mask := m.student
	if class == 1 {
		mask = m.teacher
	}
	return mask&(1<<name) != 0
}
func (f relOracleFormula) eval(m relOracleModel) bool {
	if f.member {
		name := m.alice
		if f.memberName == 1 {
			name = m.bob
		}
		return m.unary(f.memberClass, name)
	}
	values := [2]int{}
	for slot, t := range f.terms {
		if t.kind == 'a' {
			values[slot] = m.alice
		}
		if t.kind == 'b' {
			values[slot] = m.bob
		}
	}
	result := f.walk(m, values, 0)
	if f.outerNot {
		result = !result
	}
	return result
}
func (f relOracleFormula) walk(m relOracleModel, values [2]int, depth int) bool {
	result := false
	if depth == f.count {
		result = m.relation&(1<<(values[0]*m.size+values[1])) != 0
	} else {
		slot := f.order[depth]
		term := f.terms[slot]
		result = term.kind != 'E'
		for value := 0; value < m.size; value++ {
			if !m.unary(term.class, value) {
				continue
			}
			values[slot] = value
			body := f.walk(m, values, depth+1)
			if term.kind == 'A' && !body {
				result = false
				break
			}
			if term.kind == 'E' && body {
				result = true
				break
			}
			if term.kind == 'N' && body {
				result = false
				break
			}
		}
	}
	if depth == 1 && f.innerNot {
		result = !result
	}
	return result
}
func relOracleSentence(a, b relOracleTerm, scaffold int, outerNot, innerNot bool) relOracleSurface {
	f := relOracleFormula{terms: [2]relOracleTerm{a, b}, outerNot: outerNot, innerNot: innerNot}
	if scaffold >= 0 {
		f.order[f.count] = scaffold
		f.count++
	}
	for slot, t := range f.terms {
		if slot != scaffold && t.kind != 'a' && t.kind != 'b' {
			f.order[f.count] = slot
			f.count++
		}
	}
	surface := relOracleSurface{}
	left, right := a.text(), b.text()
	if scaffold >= 0 {
		bound := f.terms[scaffold]
		ref := "that " + []string{"student", "teacher"}[bound.class]
		if scaffold == 0 {
			left = ref
		} else {
			right = ref
		}
		body := left + " admires " + right
		if innerNot {
			body = "it is not the case that " + body
		}
		surface.text = bound.text() + " is such that " + body
	} else {
		surface.text = left + " admires " + right
	}
	if outerNot {
		surface.text = "it is not the case that " + surface.text
	}
	surface.readings = append(surface.readings, f)
	if scaffold < 0 && f.count == 2 {
		f.order = [2]int{1, 0}
		surface.readings = append(surface.readings, f)
	}
	return surface
}
func relOracleMember(name, class int) relOracleSurface {
	return relOracleSurface{fmt.Sprintf("%s is a %s", []string{"Alice", "Bob"}[name], []string{"student", "teacher"}[class]), []relOracleFormula{{member: true, memberName: name, memberClass: class}}}
}

type relOracleObligation struct {
	premises   []relOracleSurface
	conclusion relOracleSurface
}

type relOracleCase struct {
	argument        string
	premises        []relOracleSurface
	conclusion      relOracleSurface
	noteObligations map[int]relOracleObligation
}

func relOracleCases() []relOracleCase {
	var cases []relOracleCase
	add := func(p []relOracleSurface, c relOracleSurface) {
		texts := make([]string, len(p))
		for i, s := range p {
			texts[i] = s.text
		}
		cases = append(cases, relOracleCase{relationArgument(texts, c.text), p, c, nil})
	}
	names := []relOracleTerm{{kind: 'a'}, {kind: 'b'}}
	terms := append([]relOracleTerm{}, names...)
	for _, kind := range []byte{'A', 'E', 'N'} {
		for class := 0; class < 2; class++ {
			terms = append(terms, relOracleTerm{kind, class})
		}
	}
	// Both positions, both names, both classes, every/no instantiation and
	// existential introduction, with wrong classes, positions, and polarities.
	for slot := 0; slot < 2; slot++ {
		for _, other := range terms {
			for class := 0; class < 2; class++ {
				for ni, name := range names {
					base := [2]relOracleTerm{other, other}
					base[slot] = name
					ground := relOracleSentence(base[0], base[1], -1, false, false)
					for _, kind := range []byte{'A', 'N', 'E'} {
						quantified := base
						quantified[slot] = relOracleTerm{kind, class}
						q := relOracleSentence(quantified[0], quantified[1], -1, false, false)
						for memberClass := 0; memberClass < 2; memberClass++ {
							m := relOracleMember(ni, memberClass)
							if kind == 'E' {
								add([]relOracleSurface{ground, m}, q)
							} else {
								target := relOracleSentence(base[0], base[1], -1, kind == 'N', false)
								add([]relOracleSurface{q, m}, target)
								add([]relOracleSurface{q, m}, relOracleSentence(base[1], base[0], -1, kind == 'N', false))
							}
						}
					}
				}
			}
		}
	}
	// Every mixed pair, both class assignments, both surface scopes, negation,
	// scaffold introduction/elimination, and each direction of a scope flip.
	for _, left := range terms[2:] {
		for _, right := range terms[2:] {
			for _, neg := range []bool{false, true} {
				ambiguous := relOracleSentence(left, right, -1, neg, false)
				subject := relOracleSentence(left, right, 0, neg, false)
				object := relOracleSentence(left, right, 1, neg, false)
				for _, pair := range [][2]relOracleSurface{{ambiguous, subject}, {ambiguous, object}, {subject, ambiguous}, {object, ambiguous}, {subject, object}, {object, subject}, {ambiguous, ambiguous}} {
					add([]relOracleSurface{pair[0]}, pair[1])
				}
			}
		}
	}
	for slot := 0; slot < 2; slot++ {
		for class := 0; class < 2; class++ {
			for _, name := range names {
				terms := [2]relOracleTerm{name, name}
				terms[slot] = relOracleTerm{'E', class}
				notSome := relOracleSentence(terms[0], terms[1], -1, true, false)
				terms[slot].kind = 'N'
				none := relOracleSentence(terms[0], terms[1], -1, false, false)
				add([]relOracleSurface{notSome}, none)
				add([]relOracleSurface{none}, notSome)
				terms[slot].kind = 'A'
				notEvery := relOracleSentence(terms[0], terms[1], -1, true, false)
				terms[slot].kind = 'E'
				negativeWitness := relOracleSentence(terms[0], terms[1], slot, false, true)
				add([]relOracleSurface{notEvery}, negativeWitness)
				add([]relOracleSurface{negativeWitness}, notEvery)
			}
		}
	}
	// Small Fitch proofs. Premises/conclusion are evaluated with the same
	// independent semantics; the local eigenname is not a domain constant.
	for class := 0; class < 2; class++ {
		for _, other := range terms {
			for slot := 0; slot < 2; slot++ {
				ts := [2]relOracleTerm{other, other}
				ts[slot] = relOracleTerm{'A', class}
				universal := relOracleSentence(ts[0], ts[1], -1, false, false)
				inst := ts
				inst[slot] = names[0]
				endpoint := relOracleSentence(inst[0], inst[1], -1, false, false)
				endpointText := strings.ReplaceAll(endpoint.text, "Alice", "z")
				cls := []string{"student", "teacher"}[class]
				argument := fmt.Sprintf("1. %s.\n| 2. Let z be an arbitrary %s.\n| 3. Therefore, %s. [1,2]\n4. Therefore, %s. [2-3]", universal.text, cls, endpointText, universal.text)
				cases = append(cases, relOracleCase{argument, []relOracleSurface{universal}, universal, map[int]relOracleObligation{
					// Diagnose the local rule that failed, not merely the
					// whole proof's endpoints (which can repeat a premise).
					3: {[]relOracleSurface{universal, relOracleMember(0, class)}, endpoint},
					4: {[]relOracleSurface{relOracleSentence(ts[0], ts[1], slot, false, false)}, universal},
				}})
				// Stale Alice is an actual premise name: generalizing her is unsound.
				member := relOracleMember(0, class)
				argument = fmt.Sprintf("1. %s.\n2. %s.\n| 3. Let Alice be an arbitrary %s.\n| 4. Therefore, %s. [1]\n5. Therefore, %s. [3-4]", endpoint.text, member.text, cls, endpoint.text, universal.text)
				cases = append(cases, relOracleCase{argument, []relOracleSurface{endpoint, member}, universal, nil})
				ts[slot].kind = 'E'
				existential := relOracleSentence(ts[0], ts[1], -1, false, false)
				// Use explicit outer scope so that the witness assumption is licensed.
				existential = relOracleSentence(ts[0], ts[1], slot, false, false)
				argument = fmt.Sprintf("1. %s.\n| 2. Suppose z is a %s and %s.\n| 3. Therefore, z is a %s. [2]\n| 4. Therefore, %s. [2]\n| 5. Therefore, %s. [3,4]\n6. Therefore, %s. [1,2-5]", existential.text, cls, endpointText, cls, endpointText, existential.text, existential.text)
				cases = append(cases, relOracleCase{argument, []relOracleSurface{existential}, existential, nil})
				// Escaping Alice: an existential witness cannot become a named conclusion.
				argument = fmt.Sprintf("1. %s.\n| 2. Suppose Alice is a %s and %s.\n| 3. Therefore, %s. [2]\n4. Therefore, %s. [1,2-3]", existential.text, cls, endpoint.text, endpoint.text, endpoint.text)
				cases = append(cases, relOracleCase{argument, []relOracleSurface{existential}, endpoint, nil})
			}
		}
	}
	return cases
}
func TestEntailmentGateRelationalOracle(t *testing.T) {
	started := time.Now()
	var models []relOracleModel
	for n := 1; n <= 3; n++ {
		for s := 0; s < 1<<n; s++ {
			for teacher := 0; teacher < 1<<n; teacher++ {
				for r := 0; r < 1<<(n*n); r++ {
					for a := 0; a < n; a++ {
						for b := 0; b < n; b++ {
							models = append(models, relOracleModel{n, s, teacher, r, a, b})
						}
					}
				}
			}
		}
	}
	if len(models) != 295944 {
		t.Fatalf("model enumeration: %d", len(models))
	}
	cache := map[relOracleFormula][]uint64{}
	truth := func(f relOracleFormula) []uint64 {
		if bits, ok := cache[f]; ok {
			return bits
		}
		bits := make([]uint64, (len(models)+63)/64)
		for i, m := range models {
			if f.eval(m) {
				bits[i/64] |= uint64(1) << uint(i%64)
			}
		}
		cache[f] = bits
		return bits
	}
	// This key function is independent of gate normalization. The generated
	// grammar uses only case/space normalization and an explicit outer not.
	// Negation wraps the same underlying clause, so P and not-P share scope.
	clauseKey := func(surface relOracleSurface) string {
		key := strings.ToLower(strings.Join(strings.Fields(surface.text), " "))
		for strings.HasPrefix(key, "it is not the case that ") {
			key = strings.TrimPrefix(key, "it is not the case that ")
		}
		return key
	}
	accepted, rejected, notes := 0, 0, 0
	cases := relOracleCases()
	for _, tc := range cases {
		report, result := CheckEntailmentGate(tc.argument)
		ambiguity := false
		for _, step := range report.Steps {
			ambiguity = ambiguity || step.AmbiguityNote != ""
		}
		if !result.Passed && !ambiguity {
			rejected++
			continue
		}
		checks := []relOracleObligation{}
		if result.Passed {
			checks = append(checks, relOracleObligation{tc.premises, tc.conclusion})
		} else {
			for _, step := range report.Steps {
				if step.AmbiguityNote == "" {
					continue
				}
				obligation, ok := tc.noteObligations[step.Step]
				if !ok {
					obligation = relOracleObligation{tc.premises, tc.conclusion}
				}
				checks = append(checks, obligation)
			}
		}
		for _, obligation := range checks {
			// Assign a reading once per distinct clause, then check every model
			// under every consistent assignment. OR-ing premise readings and
			// AND-ing conclusion readings separately would incorrectly uncouple
			// two occurrences of the same clause (including P versus not-P).
			surfaces := append(append([]relOracleSurface{}, obligation.premises...), obligation.conclusion)
			keys := []string{}
			counts := map[string]int{}
			for _, surface := range surfaces {
				key := clauseKey(surface)
				if _, ok := counts[key]; !ok {
					keys = append(keys, key)
					counts[key] = len(surface.readings)
				}
			}
			selected := map[string]int{}
			countermodel := -1
			var assignments func(int)
			assignments = func(index int) {
				if countermodel >= 0 {
					return
				}
				if index < len(keys) {
					key := keys[index]
					for reading := 0; reading < counts[key]; reading++ {
						selected[key] = reading
						assignments(index + 1)
					}
					return
				}
				conclusion := truth(obligation.conclusion.readings[selected[clauseKey(obligation.conclusion)]])
				premiseBits := make([][]uint64, len(obligation.premises))
				for i, p := range obligation.premises {
					premiseBits[i] = truth(p.readings[selected[clauseKey(p)]])
				}
				for word, c := range conclusion {
					bad := ^c
					for _, p := range premiseBits {
						bad &= p[word]
					}
					if word == len(conclusion)-1 && len(models)%64 != 0 {
						bad &= (uint64(1) << uint(len(models)%64)) - 1
					}
					if bad != 0 {
						for bit := 0; bit < 64; bit++ {
							if bad&(uint64(1)<<uint(bit)) != 0 {
								countermodel = word*64 + bit
								break
							}
						}
						return
					}
				}
			}
			assignments(0)
			if result.Passed && countermodel >= 0 {
				t.Fatalf("UNSOUND acceptance\n%s\nmodel=%+v\n%+v", tc.argument, models[countermodel], report)
			}
			if !result.Passed && countermodel < 0 {
				t.Fatalf("ambiguity note without a countermodel for its local obligation\n%s\n%+v", tc.argument, report)
			}
		}

		if result.Passed {
			accepted++
		} else {
			rejected++
			notes++
		}
	}
	if accepted < 200 || notes < 20 {
		t.Fatalf("coverage fell: accepted=%d notes=%d", accepted, notes)
	}
	t.Logf("exhaustive bounded oracle: domains 1–3, %d models, %d candidates, %d accepted, %d rejected (%d ambiguity notes), %d truth tables; elapsed %s (not a proof beyond this bound)", len(models), len(cases), accepted, rejected, notes, len(cache), time.Since(started))
}
