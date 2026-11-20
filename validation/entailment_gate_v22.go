package validation

import "strings"

// nplaceCandidates recognizes flat argument lists. Prepositions delimit slots;
// they remain part of the predicate's identity. Only the listed double-object
// verbs admit the to alternation. No active/passive or lexical rewriting occurs.
func nplaceCandidates(statement string) []relationClause {
	return allStrictRules().nplaceCandidates(statement)
}

func (policy strictRules) nplaceCandidates(statement string) []relationClause {
	// Most historical clauses cannot contain a new argument boundary. Avoid
	// running the new parser at all on that hot path.
	possible := false
	for _, word := range strings.Fields(strings.ToLower(statement)) {
		switch word {
		case "to", "from", "with", "for", "at", "in", "on", "into", "onto", "under", "over", "through", "between", "beside", "by",
			"give", "gives", "gave", "send", "sends", "sent", "sell", "sells", "sold", "offer", "offers", "offered", "lend", "lends", "lent", "hand", "hands", "handed", "show", "shows", "showed":
			possible = true
		}
	}
	if !possible {
		return nil
	}
	s := normalizeClause(statement)
	outer := false
	for strings.HasPrefix(s, "it is not the case that ") {
		s = strings.TrimPrefix(s, "it is not the case that ")
		outer = !outer
	}
	var bounds []relationTerm
	inner := false
	for strings.Contains(s, " is such that ") {
		head, body, _ := strings.Cut(s, " is such that ")
		t, ok := relationTermCandidate(head)
		if !ok || t.quant == "" {
			return nil
		}
		for _, b := range bounds {
			if b.text == t.text {
				return nil
			}
		}
		bounds = append(bounds, t)
		s = body
		// Negation under the first binder has the same semantics as v17.
		if len(bounds) == 1 {
			for strings.HasPrefix(s, "it is not the case that ") {
				s = strings.TrimPrefix(s, "it is not the case that ")
				inner = !inner
			}
		}
	}
	plain := s
	for _, b := range bounds {
		plain = strings.ReplaceAll(plain, "that "+b.text, "boundterm")
	}
	if !v13PlainClause(plain) {
		return nil
	}
	words := strings.Fields(s)
	var out []relationClause
	prep := func(w string) bool {
		switch w {
		case "to", "from", "with", "for", "at", "in", "on", "into", "onto", "under", "over", "through", "between", "beside", "by":
			return true
		}
		return false
	}
	ditransitive := func(w string) bool {
		switch w {
		case "give", "gives", "gave", "send", "sends", "sent", "sell", "sells", "sold", "offer", "offers", "offered", "lend", "lends", "lent", "hand", "hands", "handed", "show", "shows", "showed":
			return true
		}
		return false
	}
	term := func(piece string, subject bool) (relationTerm, bool) {
		if subject && policy.indefiniteTermRefused(piece) {
			return relationTerm{}, false
		}
		for _, b := range bounds {
			if piece == "that "+b.text {
				return b, true
			}
		}
		// In this closed object grammar, a/an introduces an existential
		// binder. It never supplies an individual constant shared by steps.
		for _, a := range []string{"a ", "an "} {
			if strings.HasPrefix(piece, a) {
				piece = "some " + strings.TrimPrefix(piece, a)
				break
			}
		}
		t, ok := relationTermCandidate(piece)
		for _, w := range strings.Fields(t.text) {
			if prep(w) || w == "a" || w == "an" {
				ok = false
			}
		}
		return t, ok
	}
	add := func(pieces, seps []string, verb string) {
		if len(pieces) < 3 {
			return
		}
		r := relationClause{verb: verb + "\x00" + strings.Join(seps, "\x00"), separators: seps, terms: make([]relationTerm, len(pieces)), outerNot: outer, innerNot: inner, scaffold: len(bounds) > 0, boundCount: len(bounds)}
		used := map[int]bool{}
		for i, p := range pieces {
			t, ok := term(p, i == 0)
			if !ok {
				return
			}
			r.terms[i] = t
		}
		for _, b := range bounds {
			slot := -1
			for i, p := range pieces {
				if p == "that "+b.text {
					if slot >= 0 {
						return
					}
					slot = i
				}
			}
			if slot < 0 {
				return
			}
			r.order = append(r.order, slot)
			used[slot] = true
		}
		for i, t := range r.terms {
			if t.quant != "" && !used[i] {
				r.order = append(r.order, i)
			}
		}
		out = append(out, r)
	}
	for i := 1; i+2 < len(words); i++ {
		verb := words[i]
		if !relationVerb(verb) || prep(verb) {
			continue
		}
		subject := strings.Join(words[:i], " ")
		rest := words[i+1:]
		var pieces, seps []string
		pieces = append(pieces, subject)
		start := 0
		for j, w := range rest {
			if prep(w) {
				pieces = append(pieces, strings.Join(rest[start:j], " "))
				seps = append(seps, w)
				start = j + 1
			}
		}
		pieces = append(pieces, strings.Join(rest[start:], " "))
		if len(seps) > 0 {
			add(pieces, seps, verb)
		}
		if len(seps) == 0 && ditransitive(verb) {
			for j := 1; j < len(rest); j++ {
				recipient, theme := strings.Join(rest[:j], " "), strings.Join(rest[j:], " ")
				add([]string{subject, theme, recipient}, []string{"to"}, verb)
			}
		}
	}
	return out
}

func relationOrders(r relationClause) [][]int {
	fixed := r.boundCount
	if r.scaffold && fixed == 0 {
		fixed = 1
	}
	if len(r.order) > 4 {
		return nil
	}
	var out [][]int
	order := append([]int{}, r.order...)
	var visit func(int)
	visit = func(i int) {
		if i == len(order) {
			out = append(out, append([]int{}, order...))
			return
		}
		for j := i; j < len(order); j++ {
			order[i], order[j] = order[j], order[i]
			visit(i + 1)
			order[i], order[j] = order[j], order[i]
		}
	}
	visit(fixed)
	return out
}

func nplaceScopeChoices(s string) [][]string {
	var out [][]string
	for _, r := range nplaceCandidates(s) {
		// All universal binders commute, as do all existential binders. No
		// textual scope rewrite is needed for these uniform readings, including
		// repeated class names whose scaffold references would be ambiguous.
		uniform := len(r.order) > 0
		kind := ""
		for _, slot := range r.order {
			q := r.terms[slot].quant
			if kind == "" {
				kind = q
			}
			uniform = uniform && q == kind && (q == "every" || q == "some")
		}
		if uniform || len(r.order)-r.boundCount < 2 || len(r.order) > 4 {
			continue
		}
		var variants []string
		for _, order := range relationOrders(r) {
			variants = append(variants, r.nplaceExplicit(order))
		}
		out = append(out, variants)
	}
	return out
}
func (r relationClause) nplaceExplicit(order []int) string {
	parts := make([]string, len(r.terms))
	for i, t := range r.terms {
		parts[i] = t.text
		if t.quant != "" {
			parts[i] = "that " + t.text
		}
	}
	verb, _, _ := strings.Cut(r.verb, "\x00")
	s := parts[0] + " " + verb + " " + parts[1]
	for i, prep := range r.separators {
		s += " " + prep + " " + parts[i+2]
	}
	for i := len(order) - 1; i >= 0; i-- {
		if i == 0 && r.innerNot {
			s = "it is not the case that " + s
		}
		t := r.terms[order[i]]
		s = t.quant + " " + t.text + " is such that " + s
	}
	if r.outerNot {
		s = "it is not the case that " + s
	}
	return s
}

// Multiple cited memberships may instantiate distinct universal slots in one
// step. Every membership must be consumed; existential slots are never named.
func (policy strictRules) matchNplaceInstantiation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	if len(p) < 3 || len(p) > 5 {
		return schemaMatchOutcome{}
	}
	for source, step := range p {
		for _, a := range policy.nplaceCandidates(step.Statement) {
			for _, b := range policy.nplaceCandidates(c.Statement) {
				var members []clausePair
				good := true
				for i, m := range p {
					if i != source {
						ms := policy.relationMembership(m.Statement)
						if len(ms) != 1 {
							good = false
							break
						}
						members = append(members, ms[0])
					}
				}
				if !good || len(a.order) > 4 || len(b.order) > 4 {
					continue
				}
				all := true
				for _, x := range a.readings() {
					for _, y := range b.readings() {
						var reduce func(relationNormal, int) bool
						reduce = func(n relationNormal, i int) bool {
							if i == len(members) {
								return relationEntails(n, y)
							}
							m := members[i]
							for j, q := range n.prefix {
								if q.kind == "every" && q.class == m.right && reduce(relationRemove(n, j, m.left), i+1) {
									return true
								}
							}
							return false
						}
						if !reduce(x, 0) {
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

func (policy strictRules) matchV22RelationalInstantiation(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	if len(p) > 2 {
		return policy.matchNplaceInstantiation(p, c)
	}
	for _, ordered := range orderedPremiseSelections(p, 2) {
		if out := policy.matchRelationalUniversalInstantiation(ordered, c); out.matched {
			return out
		}
	}
	return schemaMatchOutcome{}
}
