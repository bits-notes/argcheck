package validation

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

// finiteFormula is independent of the English grammar and the gate's prefix
// rules. Terms starting with $ are variables; all others are constants.
// Quantifiers are unrestricted FOL; class restrictions are ordinary formulas.
type finiteFormula = folFormula

func foAtom(p string, ts ...string) *finiteFormula {
	return &finiteFormula{op: "atom", symbol: p, terms: ts}
}
func foNot(f *finiteFormula) *finiteFormula {
	return &finiteFormula{op: "not", children: []*finiteFormula{f}}
}
func foAnd(fs ...*finiteFormula) *finiteFormula { return &finiteFormula{op: "and", children: fs} }
func foOr(fs ...*finiteFormula) *finiteFormula  { return &finiteFormula{op: "or", children: fs} }
func foQuant(kind, variable, class string, body *finiteFormula) *finiteFormula {
	if kind == "no" {
		return foNot(foQuant("some", variable, class, body))
	}
	guard := foAtom(class, variable)
	if kind == "every" {
		body = foOr(foNot(guard), body)
	} else {
		body = foAnd(guard, body)
	}
	q := &finiteFormula{op: kind, symbol: variable, children: []*finiteFormula{body}}
	return q
}
func foEqual(a, b string) *finiteFormula { return &finiteFormula{op: "equal", terms: []string{a, b}} }

// Finite grounding produces a Boolean formula over ALL ground predicate
// tuples, including unary extensions. Tseitin clauses preserve satisfiability.
// Exhaustive binary search with unit propagation prunes irrelevant assignments;
// there is no sampling, unique-name assumption, or bound on predicate density.
type finiteCNF struct {
	clauses [][]int
	atoms   map[string]int
	next    int
	unary   []int
}

func (c *finiteCNF) fresh() int { c.next++; return c.next }
func (c *finiteCNF) combine(op string, ls []int) int {
	// Hash-cons identical Boolean subformulas. Scope permutations otherwise
	// duplicate large grounded subtrees and obscure immediate contradictions.
	ls = append([]int{}, ls...)
	sort.Ints(ls)
	unique := ls[:0]
	for _, l := range ls {
		if len(unique) == 0 || unique[len(unique)-1] != l {
			unique = append(unique, l)
		}
	}
	ls = unique
	if len(ls) == 1 {
		return ls[0]
	}
	key := op + fmt.Sprint(ls)
	if v := c.atoms[key]; v != 0 {
		return v
	}

	v := c.fresh()
	c.atoms[key] = v
	if op == "and" {
		for _, l := range ls {
			c.clauses = append(c.clauses, []int{-v, l})
		}
		clause := []int{v}
		for _, l := range ls {
			clause = append(clause, -l)
		}
		c.clauses = append(c.clauses, clause)
	} else {
		for _, l := range ls {
			c.clauses = append(c.clauses, []int{v, -l})
		}
		clause := []int{-v}
		clause = append(clause, ls...)
		c.clauses = append(c.clauses, clause)
	}
	return v
}
func (c *finiteCNF) ground(f *finiteFormula, size int, env map[string]int) int {
	switch f.op {
	case "atom":
		key := f.symbol + fmt.Sprint(len(f.terms)) + ":"
		for _, t := range f.terms {
			key += fmt.Sprint(env[t]) + ","
		}
		if v := c.atoms[key]; v != 0 {
			return v
		}
		v := c.fresh()
		c.atoms[key] = v
		if len(f.terms) == 1 {
			c.unary = append(c.unary, v)
		}
		return v
	case "equal":
		v := c.fresh()
		lit := v
		if env[f.terms[0]] != env[f.terms[1]] {
			lit = -v
		}
		c.clauses = append(c.clauses, []int{lit})
		return v
	case "bottom":
		v := c.fresh()
		c.clauses = append(c.clauses, []int{-v})
		return v
	case "implies":
		return c.combine("or", []int{-c.ground(f.children[0], size, env), c.ground(f.children[1], size, env)})
	case "iff":
		a, b := c.ground(f.children[0], size, env), c.ground(f.children[1], size, env)
		return c.combine("and", []int{c.combine("or", []int{-a, b}), c.combine("or", []int{-b, a})})
	case "not":
		return -c.ground(f.children[0], size, env)
	case "every", "some":
		old, had := env[f.symbol]
		var ls []int
		for value := 0; value < size; value++ {
			env[f.symbol] = value
			ls = append(ls, c.ground(f.children[0], size, env))
		}
		if had {
			env[f.symbol] = old
		} else {
			delete(env, f.symbol)
		}
		op := "or"
		if f.op == "every" {
			op = "and"
		}
		return c.combine(op, ls)
	case "and", "or":
		var ls []int
		for _, child := range f.children {
			ls = append(ls, c.ground(child, size, env))
		}
		return c.combine(f.op, ls)
	default:
		panic("unknown FOL operator: " + f.op)
	}
}
func finiteSAT(clauses [][]int, values []int8, priority []int) bool {
	for {
		changed := false
		for _, clause := range clauses {
			satisfied := false
			open, last := 0, 0
			for _, lit := range clause {
				v := lit
				sign := int8(1)
				if v < 0 {
					v = -v
					sign = -1
				}
				if values[v] == sign {
					satisfied = true
					break
				}
				if values[v] == 0 {
					open++
					last = lit
				}
			}
			if satisfied {
				continue
			}
			if open == 0 {
				return false
			}
			if open == 1 {
				v := last
				sign := int8(1)
				if v < 0 {
					v = -v
					sign = -1
				}
				values[v] = sign
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	branch := 0
	shortest := int(^uint(0) >> 1)
	for _, clause := range clauses {
		satisfied := false
		open, last := 0, 0
		for _, lit := range clause {
			v := lit
			sign := int8(1)
			if v < 0 {
				v = -v
				sign = -1
			}
			if values[v] == sign {
				satisfied = true
				break
			}
			if values[v] == 0 {
				open++
				last = v
			}
		}
		if !satisfied && open < shortest {
			shortest = open
			branch = last
		}
	}
	if branch == 0 {
		return true
	}
	for _, v := range priority {
		if values[v] == 0 {
			branch = v
			break
		}
	}
	for _, value := range []int8{1, -1} {
		next := append([]int8{}, values...)
		next[branch] = value
		if finiteSAT(clauses, next, priority) {
			return true
		}
	}
	return false
}

// finiteCountermodel takes formulas, so subsequent catalogs can reuse it.
// It exhausts domains 1..maxSize and every constant interpretation.
func finiteCountermodel(premises []*finiteFormula, conclusion *finiteFormula, maxSize int) bool {
	root := foAnd(append(append([]*finiteFormula{}, premises...), foNot(conclusion))...)
	names := map[string]bool{}
	var collect func(*finiteFormula)
	collect = func(f *finiteFormula) {
		for _, t := range f.terms {
			if !strings.HasPrefix(t, "$") {
				names[t] = true
			}
		}
		for _, child := range f.children {
			collect(child)
		}
	}
	collect(root)
	var constants []string
	for name := range names {
		constants = append(constants, name)
	}
	sort.Strings(constants)
	for size := 1; size <= maxSize; size++ {
		env := map[string]int{}
		var assign func(int) bool
		assign = func(i int) bool {
			if i < len(constants) {
				for value := 0; value < size; value++ {
					env[constants[i]] = value
					if assign(i + 1) {
						return true
					}
				}
				return false
			}
			cnf := finiteCNF{atoms: map[string]int{}}
			lit := cnf.ground(root, size, env)
			cnf.clauses = append(cnf.clauses, []int{lit})
			return finiteSAT(cnf.clauses, make([]int8, cnf.next+1), cnf.unary)
		}
		if assign(0) {
			return true
		}
	}
	return false
}

type nOracleSurface struct {
	text     string
	readings []*finiteFormula
}

func nGround(text, pred string, terms ...string) nOracleSurface {
	return nOracleSurface{text, []*finiteFormula{foAtom(pred, terms...)}}
}
func nIdentity(text, a, b string) nOracleSurface {
	return nOracleSurface{text, []*finiteFormula{foEqual(a, b)}}
}

// q maps slots to "every class" / "some class". An explicit order fixes
// scope; nil enumerates every order, independently of relationOrders.
func nQuantified(text, pred string, terms []string, q map[int]string, explicit []int) nOracleSurface {
	var slots []int
	for slot := range q {
		slots = append(slots, slot)
	}
	sort.Ints(slots)
	var orders [][]int
	if explicit != nil {
		orders = append(orders, explicit)
	} else {
		var perm func([]int, []int)
		perm = func(prefix, rest []int) {
			if len(rest) == 0 {
				orders = append(orders, prefix)
				return
			}
			for i, v := range rest {
				next := append([]int{}, rest[:i]...)
				next = append(next, rest[i+1:]...)
				perm(append(append([]int{}, prefix...), v), next)
			}
		}
		perm(nil, slots)
	}
	out := nOracleSurface{text: text}
	for _, order := range orders {
		ts := append([]string{}, terms...)
		for slot := range q {
			ts[slot] = fmt.Sprintf("$v%d", slot)
		}
		f := foAtom(pred, ts...)
		for i := len(order) - 1; i >= 0; i-- {
			slot := order[i]
			kind, class, _ := strings.Cut(q[slot], " ")
			f = foQuant(kind, ts[slot], class, f)
		}
		out.readings = append(out.readings, f)
	}
	return out
}
func nOracleInvalid(p []nOracleSurface, c nOracleSurface) bool {
	selected := map[string]int{}
	var walk func(int, []*finiteFormula) bool
	walk = func(i int, ps []*finiteFormula) bool {
		if i == len(p) {
			for j, f := range c.readings {
				if old, ok := selected[c.text]; ok && old != j {
					continue
				}
				if finiteCountermodel(ps, f, 3) {
					return true
				}
			}
			return false
		}
		for j, f := range p[i].readings {
			if old, ok := selected[p[i].text]; ok && old != j {
				continue
			}
			old, had := selected[p[i].text]
			selected[p[i].text] = j
			if walk(i+1, append(ps, f)) {
				return true
			}
			if had {
				selected[p[i].text] = old
			} else {
				delete(selected, p[i].text)
			}
		}
		return false
	}
	return walk(0, nil)
}
func TestCatalogV22FiniteOracle(t *testing.T) {
	start := time.Now()
	valid, invalid := 0, 0
	for _, tc := range nplaceExercises() {
		t.Run(tc.name, func(t *testing.T) {
			bad := nOracleInvalid(tc.p, tc.c)
			if bad == tc.valid {
				t.Fatalf("countermodel=%v, expected valid=%v", bad, tc.valid)
			}
			if tc.valid {
				valid++
			} else {
				invalid++
			}
		})
	}
	t.Logf("domains 1–3, %d valid and %d invalid exercises, exhaustive finite grounding and binary model search; elapsed %s", valid, invalid, time.Since(start))
}
func TestFiniteOracleSanity(t *testing.T) {
	p := foAtom("R", "a", "b", "c", "d")
	if finiteCountermodel([]*finiteFormula{p}, p, 3) {
		t.Fatal("reflexivity")
	}
	if !finiteCountermodel([]*finiteFormula{p}, foAtom("R", "b", "a", "c", "d"), 3) {
		t.Fatal("role swap")
	}
	if finiteCountermodel([]*finiteFormula{p, foEqual("a", "b")}, foAtom("R", "b", "a", "c", "d"), 3) {
		t.Fatal("identity")
	}
	// Forces exactly the reason size three matters: at least three distinct
	// members, even though constants are free to co-denote in smaller models.
	distinct := foAnd(foNot(foEqual("a", "b")), foNot(foEqual("b", "c")), foNot(foEqual("a", "c")))
	if finiteCountermodel(nil, foNot(distinct), 2) || !finiteCountermodel(nil, foNot(distinct), 3) {
		t.Fatal("domain bound")
	}
}

// Check the local obligations inside the four Fitch exercises as well as
// their overall endpoints. In particular, repeating a premise at the final
// line cannot hide an unsound instantiation in the subproof.
func TestCatalogV22FitchLocalOracle(t *testing.T) {
	for arity := 3; arity <= 4; arity++ {
		names := []string{"Alice", "Parcel", "Bob", "Depot"}[:arity]
		classes := []string{"courier", "package", "recipient", "station"}[:arity]
		pred := "sends/to"
		if arity == 4 {
			pred += "/from"
		}
		slot := arity - 1
		vars := append([]string{}, names...)
		vars[slot] = "$z"
		body := foAtom(pred, vars...)
		universal := foQuant("every", "$z", classes[slot], body)
		existential := foQuant("some", "$z", classes[slot], body)
		witness := append([]string{}, names...)
		witness[slot] = "z"
		atom := foAtom(pred, witness...)
		member := foAtom(classes[slot], "z")
		assumption := foAnd(member, atom)
		checks := []struct {
			p []*finiteFormula
			c *finiteFormula
		}{
			{[]*finiteFormula{universal, member}, atom},
			{[]*finiteFormula{assumption}, member},
			{[]*finiteFormula{assumption}, atom},
			{[]*finiteFormula{member, atom}, existential},
		}
		for i, check := range checks {
			if finiteCountermodel(check.p, check.c, 3) {
				t.Fatalf("arity %d local obligation %d", arity, i)
			}
		}
		// The discharge obligations bind the eigenname; it is not a constant in
		// either the UG result or the EE result.
		alpha := append([]string{}, names...)
		alpha[slot] = "$fresh"
		generalized := foQuant("every", "$fresh", classes[slot], foAtom(pred, alpha...))
		if finiteCountermodel([]*finiteFormula{generalized}, universal, 3) {
			t.Fatal("UG discharge")
		}
		implication := foOr(foNot(foAtom(pred, alpha...)), existential)
		closedSubproof := foQuant("every", "$fresh", classes[slot], implication)
		if finiteCountermodel([]*finiteFormula{existential, closedSubproof}, existential, 3) {
			t.Fatal("EE discharge")
		}
	}
}
