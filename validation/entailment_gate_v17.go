package validation

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// A relation has an ordered sequence of argument slots. Prefix entries bind slots, never
// substrings of the opaque verb. No predicate-logic engine or model search is
// used by the gate: the rules below are restricted quantifier proof rules.
type relationTerm struct{ quant, text string }
type relationQuantifier struct {
	kind, class string
	slot        int
}
type relationClause struct {
	terms              []relationTerm
	verb               string
	order              []int
	outerNot, innerNot bool
	scaffold           bool
	separators         []string
	boundCount         int
}
type relationNormal struct {
	names    []string
	verb     string
	prefix   []relationQuantifier
	negative bool
}

func relationWords(s string) bool {
	if s == "" || strings.ContainsAny(s, ",;:\"`()") || opaqueContextWords.MatchString(s) {
		return false
	}
	for _, w := range strings.Fields(s) {
		switch w {
		case "every", "each", "some", "no", "all", "any", "that", "not", "is", "are", "am", "if", "then", "and", "or", "either":
			return false
		}
	}
	return true
}
func relationTermCandidate(s string) (relationTerm, bool) {
	t := relationTerm{text: s}
	first, rest, _ := strings.Cut(s, " ")
	if (first == "a" || first == "an") && rest != "" {
		return relationTerm{}, false
	}
	switch first {
	case "every", "each", "some", "no":
		t.quant, t.text = first, rest
		if t.quant == "each" {
			t.quant = "every"
		}
	}
	return t, relationWords(t.text)
}
func relationVerb(s string) bool {
	if s == "is" || s == "are" || s == "am" {
		return false
	}
	// Copular relational phrases (is preferable to, is admired by) remain
	// opaque, and are never converted to active voice or to unary predicates.
	if strings.HasPrefix(s, "is ") {
		s = strings.TrimPrefix(s, "is ")
	}
	if !relationWords(s) {
		return false
	}
	first, _, _ := strings.Cut(s, " ")
	switch first {
	case "of", "who", "which", "whose", "with", "without", "in", "on", "from", "a", "an", "the":
		return false
	}
	return true
}

func relationCandidates(statement string) []relationClause {
	if out := nplaceCandidates(statement); len(out) > 0 {
		return out
	}
	s := normalizeClause(statement)
	outerNot := false
	for strings.HasPrefix(s, "it is not the case that ") {
		s = strings.TrimPrefix(s, "it is not the case that ")
		outerNot = !outerNot
	}
	var bound relationTerm
	scaffold, innerNot := false, false
	if head, body, ok := strings.Cut(s, " is such that "); ok {
		var valid bool
		bound, valid = relationTermCandidate(head)
		if !valid || bound.quant == "" {
			return nil
		}
		scaffold = true
		s = body
		for strings.HasPrefix(s, "it is not the case that ") {
			s = strings.TrimPrefix(s, "it is not the case that ")
			innerNot = !innerNot
		}
	}
	plain := s
	if scaffold {
		plain = strings.ReplaceAll(plain, "that "+bound.text, "boundterm")
	}
	if !v13PlainClause(plain) {
		return nil
	}
	words := strings.Fields(s)
	var out []relationClause
	for i := 1; i < len(words)-1; i++ {
		for j := i + 1; j < len(words); j++ {
			pieces := [2]string{strings.Join(words[:i], " "), strings.Join(words[j:], " ")}
			verb := strings.Join(words[i:j], " ")
			if !relationVerb(verb) {
				continue
			}
			r := relationClause{terms: make([]relationTerm, 2), verb: verb, outerNot: outerNot, innerNot: innerNot, scaffold: scaffold}
			boundSlot := -1
			valid := true
			for slot, piece := range pieces {
				if scaffold && piece == "that "+bound.text {
					if boundSlot != -1 {
						valid = false
						break
					}
					r.terms[slot], boundSlot = bound, slot
				} else {
					var ok bool
					r.terms[slot], ok = relationTermCandidate(piece)
					if !ok {
						valid = false
						break
					}
				}
			}
			if !valid || scaffold && boundSlot < 0 {
				continue
			}
			if scaffold {
				r.order = append(r.order, boundSlot)
			}
			for slot, t := range r.terms {
				if t.quant != "" && slot != boundSlot {
					r.order = append(r.order, slot)
				}
			}
			out = append(out, r)
		}
	}
	return out
}
func (r relationClause) readings() []relationNormal {
	orders := relationOrders(r)
	var out []relationNormal
	for _, order := range orders {
		n := relationNormal{names: make([]string, len(r.terms)), verb: r.verb, negative: r.outerNot}
		for i, t := range r.terms {
			if t.quant == "" {
				n.names[i] = t.text
			}
		}
		for i, slot := range order {
			if i == 1 && r.scaffold {
				n.negative = n.negative != r.innerNot
			}
			t := r.terms[slot]
			kind := t.quant
			if kind == "no" {
				kind = "some"
				n.negative = !n.negative
			}
			if n.negative {
				if kind == "every" {
					kind = "some"
				} else {
					kind = "every"
				}
			}
			n.prefix = append(n.prefix, relationQuantifier{kind, t.text, slot})
		}
		if r.scaffold && len(order) == 1 {
			n.negative = n.negative != r.innerNot
		}
		out = append(out, n)
	}
	return out
}

// Prefix proofs use only sound local laws: like quantifiers commute;
// exists/forall weakens to forall/exists; a universal can become existential
// when its class has an inhabitant. An earlier binder supplies an inhabitant
// inside its scope even when its class could be empty outside that scope.
func relationEntails(a, b relationNormal) bool { return relationEntailsWithMember(a, b, clausePair{}) }
func relationEntailsWithMember(a, b relationNormal, member clausePair) bool {
	if !slices.Equal(a.names, b.names) || a.verb != b.verb || a.negative != b.negative || len(a.prefix) != len(b.prefix) {
		return false
	}
	type prefixKey [4]relationQuantifier
	if len(a.prefix) > 4 {
		return false
	}
	key := func(p []relationQuantifier) prefixKey { var k prefixKey; copy(k[:], p); return k }
	target := key(b.prefix)
	queue := []prefixKey{key(a.prefix)}
	seen := map[prefixKey]bool{}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if p == target {
			return true
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		for i := 0; i+1 < len(a.prefix); i++ {
			if p[i].kind == p[i+1].kind || p[i].kind == "some" && p[i+1].kind == "every" {
				swapped := p
				swapped[i], swapped[i+1] = swapped[i+1], swapped[i]
				queue = append(queue, swapped)
			}
		}
		for i := range a.prefix {
			if p[i].kind != "every" {
				continue
			}
			inhabited := p[i].class == member.right
			for j := 0; j < i; j++ {
				inhabited = inhabited || p[j].class == p[i].class
			}
			if inhabited {
				weaker := p
				weaker[i].kind = "some"
				queue = append(queue, weaker)
			}
		}
	}
	return false
}

func relationRemove(n relationNormal, i int, name string) relationNormal {
	n.names = append([]string{}, n.names...)
	n.names[n.prefix[i].slot] = name
	n.prefix = append(append([]relationQuantifier{}, n.prefix[:i]...), n.prefix[i+1:]...)
	return n
}
func relationMembership(s string) []clausePair {
	var out []clausePair
	for _, m := range groundInstanceCandidates(s) {
		class := m.right
		for _, article := range []string{"a ", "an "} {
			class = strings.TrimPrefix(class, article)
		}
		if relationWords(m.left) && relationWords(class) && normalizeClause(s) == m.left+" is "+m.right {
			out = append(out, clausePair{m.left, class})
		}
	}
	return out
}

func relationHasQuantifier(s string) bool {
	for _, word := range strings.Fields(strings.ToLower(s)) {
		switch word {
		case "every", "each", "some", "no":
			return true
		}
	}
	return false
}

func matchRelationalRule(p []NumberedArgumentStep, c NumberedArgumentStep, rule string) schemaMatchOutcome {
	if !relationHasQuantifier(p[0].Statement) && !relationHasQuantifier(c.Statement) && len(nplaceCandidates(p[0].Statement)) == 0 && len(nplaceCandidates(c.Statement)) == 0 {
		return schemaMatchOutcome{}
	}

	for _, a := range relationCandidates(p[0].Statement) {
		for _, b := range relationCandidates(c.Statement) {
			if a.verb != b.verb {
				continue
			}
			memberships := []clausePair{{}}
			if len(p) == 2 {
				memberships = relationMembership(p[1].Statement)
			}
			for _, member := range memberships {
				all := true
				for _, x := range a.readings() {
					for _, y := range b.readings() {
						ok := false
						switch rule {
						case "scope":
							ok = relationEntails(x, y) && (len(x.prefix) > 0 || a.scaffold || b.scaffold || len(a.terms) > 2)
						case "UI", "negative UI":
							for i, q := range x.prefix {
								if q.kind == "every" && q.class == member.right {
									reduced := relationRemove(x, i, member.left)
									if relationEntailsWithMember(reduced, y, member) && (rule == "UI" && !x.negative || rule == "negative UI" && x.negative) {
										ok = true
									}
								}
							}
						case "EG":
							for i, q := range y.prefix {
								if q.kind == "some" && q.class == member.right && relationEntailsWithMember(x, relationRemove(y, i, member.left), member) {
									ok = true
								}
							}
						}
						if !ok {
							all = false
						}
					}
				}
				if all {
					return schemaMatchOutcome{matched: true}
				}
			}
		}
	}
	return schemaMatchOutcome{}
}
func matchRelationalUniversalInstantiation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return matchRelationalRule(p, c, "UI")
}
func matchRelationalNegativeInstantiation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return matchRelationalRule(p, c, "negative UI")
}
func matchRelationalExistentialGeneralization(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return matchRelationalRule(p, c, "EG")
}
func matchRelationalScope(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	return matchRelationalRule(p, c, "scope")
}

// Render the original (not dualized) quantifier reading. This is used only
// internally to make every scope choice explicit before invoking the catalog.
func (r relationClause) explicit(outer int) string {
	parts := [2]string{}
	for i, t := range r.terms {
		parts[i] = t.text
		if t.quant != "" {
			parts[i] = t.quant + " " + t.text
		}
	}
	parts[outer] = "that " + r.terms[outer].text
	s := r.terms[outer].quant + " " + r.terms[outer].text + " is such that " + parts[0] + " " + r.verb + " " + parts[1]
	if r.outerNot {
		s = "it is not the case that " + s
	}
	return s
}
func relationScopeChoices(s string) [][]string {
	if out := nplaceScopeChoices(s); len(out) > 0 {
		return out
	}
	var out [][]string
	seen := map[string]bool{}
	for _, r := range relationCandidates(s) {
		if len(r.terms) != 2 || r.scaffold || len(r.order) != 2 {
			continue
		}
		variants := []string{r.explicit(0), r.explicit(1)}
		key := strings.Join(variants, "\x00")
		if !seen[key] {
			seen[key] = true
			out = append(out, variants)
		}
	}
	return out
}

func relationNestedNoun(s string) bool {
	s = normalizeClause(s)
	for _, prep := range []string{" of ", " from ", " in ", " with "} {
		for _, q := range []string{"every ", "each ", "some ", "no "} {
			if before, _, ok := strings.Cut(s, prep+q); ok {
				for _, word := range strings.Fields(before) {
					switch word {
					case "every", "each", "some", "no":
						return len(relationCandidates(s)) == 0
					}
				}
			}
		}
	}
	return false
}

// Scope occurrences carry syntactic clause boundaries. They are discovered
// through the same connective splitters as the catalog, never by replacing an
// arbitrary substring inside a name, relation, or attitude report.
type relationScopeOccurrence struct {
	start, end int
	clause     string
	variants   [][]string
}

func relationScopeOccurrences(statement string) []relationScopeOccurrence {
	var out []relationScopeOccurrence
	seen := map[string]bool{}
	var visit func(string, int)
	visit = func(s string, offset int) {
		if !relationHasQuantifier(s) {
			return
		}
		key := fmt.Sprintf("%d:%s", offset, s)
		if seen[key] {
			return
		}
		seen[key] = true
		child := func(part string, last bool) {
			if part == "" || len(part) >= len(s) {
				return
			}
			index := strings.Index(s, part)
			if last {
				index = strings.LastIndex(s, part)
			}
			if index >= 0 {
				visit(part, offset+index)
			}
		}
		// Negation/modal operators wrap a clause; its scope choice is shared with
		// the same positive/unmodalized clause elsewhere (not a new occurrence).
		for _, parts := range [][]string{negationCandidates(s), necessityCandidates(s), possibilityCandidates(s)} {
			if len(parts) > 0 {
				for _, part := range parts {
					child(part, false)
				}
				return
			}
		}
		if variants := relationScopeChoices(s); len(variants) > 0 {
			out = append(out, relationScopeOccurrence{offset, offset + len(s), s, variants})
			return
		}
		if unwrapped := stripClauseWrappers(s); unwrapped != s {
			child(unwrapped, false)
			return
		}
		for _, pairs := range [][]clausePair{conditionalCandidates(s), biconditionalCandidates(s)} {
			if len(pairs) > 0 {
				for _, pair := range pairs {
					child(pair.left, false)
					child(pair.right, true)
				}
				return
			}
		}
		if hasOuterGateOperator(s) || strings.Contains(s, " is such that ") {
			return
		}
		for _, pairs := range [][]clausePair{conjunctionCandidates(s), disjunctionCandidates(s)} {
			for _, pair := range pairs {
				child(pair.left, false)
				child(pair.right, true)
			}
		}
	}
	visit(normalizeClause(statement), 0)
	sort.Slice(out, func(i, j int) bool {
		if out[i].start == out[j].start {
			return out[i].end > out[j].end
		}
		return out[i].start < out[j].start
	})
	return out
}
func rewriteRelationScopes(statement string, occurrences []relationScopeOccurrence, bindings map[string]string) string {
	if len(occurrences) == 0 {
		return statement
	}
	s := normalizeClause(statement)
	var b strings.Builder
	end := 0
	for _, o := range occurrences {
		if o.start < end {
			continue
		} // duplicated candidate decompositions
		b.WriteString(s[end:o.start])
		b.WriteString(bindings[o.clause])
		end = o.end
	}
	b.WriteString(s[end:])
	return b.String()
}

// One reading is assigned to each distinct normalized clause, including
// occurrences inside connectives, negations and Fitch range endpoints.
func gateRelationalScopes(c NumberedArgumentStep, steps map[int]NumberedArgumentStep) EntailmentGateStepReport {
	raw := gateConclusionStepUnscoped(c, steps)
	ids := append([]int{}, c.Citations...)
	for _, r := range c.Ranges {
		ids = append(ids, r[0], r[1])
	}
	ids = append(ids, c.Number)
	type choice struct {
		id       int
		original string
		variants [][]string
	}
	var choices []choice
	occurrences := map[int][]relationScopeOccurrence{}
	clauseSeen := map[string]bool{}
	originals := map[int]NumberedArgumentStep{}
	seen := map[int]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		s := steps[id]
		if id == c.Number {
			s = c
		}
		if !relationHasQuantifier(s.Statement) && len(nplaceCandidates(s.Statement)) == 0 {
			continue
		}
		quantifiers := 0
		if v13PlainClause(s.Statement) || strings.Contains(normalizeClause(s.Statement), " is such that ") {
			for _, word := range strings.Fields(normalizeClause(s.Statement)) {
				switch word {
				case "every", "each", "some", "no":
					quantifiers++
				}
			}
		}
		for _, r := range nplaceCandidates(s.Statement) {
			if len(r.order) > quantifiers {
				quantifiers = len(r.order)
			}
		}
		if quantifiers > 4 {
			raw.Passed = false
			raw.Form = ""
			raw.NearestForm = "relational grammar"
			raw.Mismatch = "the relational grammar supports at most four quantifiers per sentence; use separate explicit steps"
			return raw
		}
		if relationNestedNoun(s.Statement) {
			raw.Passed = false
			raw.Form = ""
			raw.NearestForm = "relational grammar"
			raw.Mismatch = "noun phrases containing nested quantifiers are outside the relational grammar; use an explicit is such that scaffold"
			return raw
		}
		originals[id] = s
		occurrences[id] = relationScopeOccurrences(s.Statement)
		for _, o := range occurrences[id] {
			if !clauseSeen[o.clause] {
				clauseSeen[o.clause] = true
				choices = append(choices, choice{id, o.clause, o.variants})
			}
		}
	}
	if len(choices) == 0 {
		return raw
	}
	local := make(map[int]NumberedArgumentStep, len(steps))
	for id, s := range steps {
		local[id] = s
	}
	applyBindings := func(bindings map[string]string) NumberedArgumentStep {
		for id, original := range originals {
			changed := original
			changed.Statement = rewriteRelationScopes(original.Statement, occurrences[id], bindings)
			local[id] = changed
		}
		target := c
		if changed, ok := local[c.Number]; ok {
			target = changed
		}
		return target
	}
	// A certificate that still works after replacing each ambiguous clause
	// with a distinct opaque atom is uniform in its meaning. This proves all
	// consistent scope assignments at once and preserves the original form.
	// It also avoids imposing relational scaffold syntax on propositional rules.
	if raw.Passed {
		atoms := map[string]string{}
		for i, ch := range choices {
			atom := fmt.Sprintf("v17scopeatom%d", i)
			for {
				collision := false
				for _, step := range steps {
					collision = collision || strings.Contains(strings.ToLower(step.Statement), atom)
				}
				if !collision && !strings.Contains(strings.ToLower(c.Statement), atom) {
					break
				}
				atom += "x"
			}
			atoms[ch.original] = atom
		}
		if proof := gateConclusionStepUnscoped(applyBindings(atoms), local); proof.Passed {
			return raw
		}
	}
	bindings := map[string]string{}
	any, all := false, false
	var failed EntailmentGateStepReport
	var successful []string
	selected := make([]string, len(choices))
	parses := make([][]string, len(choices))
	// Splits are existential (premise-guided); scope choices are universal.
	// Keep a split fixed across all readings, so a proof cannot switch what its
	// verb or noun denotes when the quantifier order changes.
	var chooseParses func(int)
	chooseParses = func(i int) {
		if all {
			return
		}
		if i < len(choices) {
			for _, pair := range choices[i].variants {
				parses[i] = pair
				chooseParses(i + 1)
			}
			return
		}
		thisAll := true
		var visit func(int)
		visit = func(j int) {
			if j == len(choices) {
				target := applyBindings(bindings)
				r := gateConclusionStepUnscoped(target, local)
				if r.Passed {
					any = true
					if successful == nil {
						successful = append([]string{}, selected...)
					}
				} else {
					failed = r
					thisAll = false
				}
				return
			}
			ch := choices[j]
			for _, variant := range parses[j] {
				bindings[ch.original] = variant
				selected[j] = variant
				visit(j + 1)
			}
		}
		visit(0)
		if thisAll {
			all = true
		}
	}
	chooseParses(0)
	if all {
		if !raw.Passed {
			raw.Passed = true
			raw.Form = "relational scope conversion"
			raw.NearestForm = ""
			raw.Mismatch = ""
		}
		return raw
	}
	raw.Passed = false
	raw.Form = ""
	if any {
		ch := choices[0]
		// The scaffold itself is a plain-English statement of the selected reading.
		reading := successful[0]
		if normalizeClause(ch.original) == "every student admires some teacher" && reading == "some teacher is such that every student admires that teacher" {
			reading = "some teacher is admired by every student"
		}
		scaffold := successful[0]
		scaffold = strings.ToUpper(scaffold[:1]) + scaffold[1:] + "."
		raw.AmbiguityNote = fmt.Sprintf("step %d: %q has two scope readings; this step follows only if %s. To mean that, write %q", c.Number, normalizeClause(ch.original), reading, scaffold)
		if len(choices) > 1 {
			var selectedReadings []string
			for i, choice := range choices {
				sentence := successful[i]
				sentence = strings.ToUpper(sentence[:1]) + sentence[1:] + "."
				selectedReadings = append(selectedReadings, fmt.Sprintf("step %d: %q", choice.id, sentence))
			}
			raw.AmbiguityNote = fmt.Sprintf("step %d: %q has two scope readings; this step follows only under some combinations of readings. One certified combination uses these explicit scopes: %s", c.Number, normalizeClause(ch.original), strings.Join(selectedReadings, "; "))
		}
		raw.NearestForm = "relational scope conversion"
		if len(nplaceCandidates(choices[0].original)) > 0 {
			raw.AmbiguityNote = strings.ReplaceAll(raw.AmbiguityNote, "two scope readings", "multiple scope readings")
		}
		raw.Mismatch = raw.AmbiguityNote
	} else if raw.NearestForm == "" {
		raw.NearestForm = failed.NearestForm
		raw.Mismatch = failed.Mismatch
	}
	return raw
}

// Fitch binding must remove an outer quantifier. Commuting like quantifiers
// is safe; moving a witness across a preceding universal is not.
func relationOuterRemovals(n relationNormal, kind, class, name string) []relationNormal {
	var out []relationNormal
	for i, q := range n.prefix {
		if q.kind != kind {
			break
		}
		if q.class == class {
			out = append(out, relationRemove(n, i, name))
		}
	}
	return out
}
func matchRelationalUG(last, c NumberedArgumentStep, name, class string) bool {
	for _, word := range strings.Fields(last.Statement) {
		if strings.EqualFold(word, name) && word != name {
			return false
		}
	}

	for _, end := range relationCandidates(last.Statement) {
		for _, goal := range relationCandidates(c.Statement) {
			all := true
			for _, x := range end.readings() {
				for _, y := range goal.readings() {
					ok := false
					// Generalize the eigenname at outer scope first, then
					// apply the same sound prefix laws as a one-step scope
					// conversion. Merely removing a binder from the goal
					// misses valid weakenings licensed by an enclosing class.
					for slot, term := range x.names {
						if term != normalizeClause(name) {
							continue
						}
						generalized := x
						generalized.names = append([]string{}, x.names...)
						generalized.names[slot] = ""
						generalized.prefix = append([]relationQuantifier{{"every", normalizeClause(class), slot}}, x.prefix...)
						if relationEntails(generalized, y) {
							ok = true
						}
					}
					if !ok {
						all = false
					}
				}
			}
			if all {
				return true
			}
		}
	}
	return false
}
func relationalEEName(first, existential NumberedArgumentStep) string {
	for _, pair := range conjunctionCandidates(first.Statement) {
		for _, sides := range [][2]string{{pair.left, pair.right}, {pair.right, pair.left}} {
			for _, m := range relationMembership(sides[0]) {
				if !fitchNamePattern.MatchString(m.left) {
					continue
				}
				for _, source := range relationCandidates(existential.Statement) {
					for _, witness := range relationCandidates(sides[1]) {
						all := true
						for _, x := range source.readings() {
							for _, y := range witness.readings() {
								ok := false
								for _, reduced := range relationOuterRemovals(x, "some", m.right, m.left) {
									if relationEntails(reduced, y) {
										ok = true
									}
								}
								if !ok {
									all = false
								}
							}
						}
						if all {
							return m.left
						}
					}
				}
			}
		}
	}
	return ""
}
