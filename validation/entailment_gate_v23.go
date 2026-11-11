package validation

import "strings"

// Relative clauses use a deliberately bounded single-token verb vocabulary.
// Keeping their parser private prevents changes to historical candidate sets.
const relativeHole = "zzrelativeholezz"

func relativeVerb(w string) bool {
	switch w {
	case "lobbies", "funds", "admires", "helps", "visits", "owns", "reads", "runs", "sleeps", "smiles", "works", "gave", "gives", "sends", "sent", "sells", "sold", "offers", "offered", "lends", "lent", "hands", "handed", "shows", "showed":
		return true
	}
	return false
}
func relativeMarker(s string) bool {
	for _, m := range []string{" that ", " who ", " which "} {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}
func relativeNP(s, name string) ([]string, bool) {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "a "), "an ")
	words := strings.Fields(s)
	at := -1
	for i, w := range words {
		if w == "that" || w == "who" || w == "which" {
			if at >= 0 {
				return nil, false
			}
			at = i
		}
	}
	if at < 0 {
		if !relationWords(s) {
			return nil, false
		}
		return []string{name + " is a " + s}, true
	}
	if at == 0 || at+1 == len(words) {
		return nil, false
	}
	class := strings.Join(words[:at], " ")
	if !relationWords(class) {
		return nil, false
	}
	tail := words[at+1:]
	verb := -1
	for i, w := range tail {
		if relativeVerb(w) {
			if verb >= 0 {
				return nil, false
			}
			verb = i
		}
	}
	if verb < 0 {
		return nil, false
	}
	var clause string
	if verb == 0 {
		clause = name + " " + strings.Join(tail, " ")
	} else {
		clause = strings.Join(tail, " ") + " " + name
	}
	// No hidden quantifier, connective, or second relative clause in a guard.
	for _, w := range tail {
		switch w {
		case "every", "some", "no", "each", "a", "an", "is", "not", "and", "or", "if", "then":
			return nil, false
		}
	}
	if !v13PlainClause(clause) {
		return nil, false
	}
	return []string{name + " is a " + class, clause}, true
}
func relativeMask(words []string) []string {
	for i, w := range words {
		if w == "that" {
			words[i] = "relativeconnector"
		}
	}
	return words
}

// A cited premise must contribute to the proof, even when its unused atoms
// may be discarded by conjunction elimination.
func relativeUsed(p []NumberedArgumentStep, want []string) bool {
	for _, step := range p {
		facts, ok := relativeGround(step.Statement)
		if !ok {
			return false
		}
		used := false
		for _, f := range facts {
			if relativeContains(want, f) {
				used = true
			}
		}
		if !used {
			return false
		}
	}
	return true
}
func relativePlain(s string) bool {
	return v13PlainClause(strings.Join(relativeMask(strings.Fields(s)), " ")) && !strings.Contains(s, "it is not the case")
}
func relativeGround(s string) ([]string, bool) {
	s = normalizeClause(s)
	if !relativePlain(s) {
		return nil, false
	}
	if subject, pred, ok := strings.Cut(s, " is "); ok && relativeMarker(pred) && !relativeMarker(subject) {
		if !strings.HasPrefix(pred, "a ") && !strings.HasPrefix(pred, "an ") {
			return nil, false
		}
		if t, ok := relationTermCandidate(subject); !ok || t.quant != "" {
			return nil, false
		}
		return relativeNP(pred, subject)
	}
	return []string{s}, true
}

type relativeQuantified struct {
	kind   string
	guards []string
	body   string
}

func relativeQuant(s string) (relativeQuantified, bool) {
	s = normalizeClause(s)
	var q relativeQuantified
	if strings.Contains(s, relativeHole) {
		return q, false
	}
	if !relativePlain(s) {
		return q, false
	}
	if head, body, ok := strings.Cut(s, " is such that "); ok {
		kind, np, ok := strings.Cut(head, " ")
		if !ok || (kind != "every" && kind != "some" && kind != "no") {
			return q, false
		}
		guards, ok := relativeNP(np, relativeHole)
		if !ok {
			return q, false
		}
		base := np
		for _, m := range []string{" that ", " who ", " which "} {
			base, _, _ = strings.Cut(base, m)
		}
		// A repeated class would shadow the outer demonstrative reference.
		for _, kind := range []string{"every ", "some ", "no "} {
			if strings.Contains(" "+body, " "+kind+base+" ") {
				return q, false
			}
		}
		ref := "that " + base
		if !strings.Contains(" "+body+" ", " "+ref+" ") {
			return q, false
		}
		return relativeQuantified{kind, guards, strings.TrimSpace(strings.ReplaceAll(" "+body+" ", " "+ref+" ", " "+relativeHole+" "))}, true
	}
	words := strings.Fields(s)
	qi := -1
	for i, w := range words {
		if w == "every" || w == "some" || w == "no" {
			if qi >= 0 {
				return q, false
			}
			qi = i
		}
	}
	if qi < 0 {
		return q, false
	}
	end := len(words)
	if qi == 0 {
		rel := false
		seenVerb := false
		for i := 1; i < len(words); i++ {
			w := words[i]
			if w == "that" || w == "who" || w == "which" {
				if rel {
					return q, false
				}
				rel = true
			}
			if w == "is" || relativeVerb(w) {
				if !rel || seenVerb {
					end = i
					break
				}
				seenVerb = true
			}
		}
		if end == len(words) {
			return q, false
		}
	}
	guards, ok := relativeNP(strings.Join(words[qi+1:end], " "), relativeHole)
	if !ok {
		return q, false
	}
	body := strings.Join(append(append(append([]string{}, words[:qi]...), relativeHole), words[end:]...), " ")
	if qi > 0 && len(relationCandidates(body)) == 0 {
		return q, false
	}
	return relativeQuantified{words[qi], guards, body}, true
}
func relativeFill(s, name string) string { return strings.ReplaceAll(s, relativeHole, name) }
func relativeFacts(p []NumberedArgumentStep) ([]string, bool) {
	var facts []string
	for _, s := range p {
		f, ok := relativeGround(s.Statement)
		if !ok {
			return nil, false
		}
		facts = append(facts, f...)
	}
	return facts, true
}
func relativeContains(facts []string, s string) bool {
	for _, f := range facts {
		if equalClauses(f, s) {
			return true
		}
	}
	return false
}
func relativeAll(facts, want []string) bool {
	for _, w := range want {
		if !relativeContains(facts, w) {
			return false
		}
	}
	return true
}
func relativeNames(facts []string) []string {
	var names []string
	for _, f := range facts {
		if a, _, ok := strings.Cut(f, " is "); ok {
			if t, ok := relationTermCandidate(a); ok && t.quant == "" {
				names = append(names, a)
			}
		}
	}
	return names
}
func matchRelativeClauses(p []NumberedArgumentStep, c NumberedArgumentStep) schemaMatchOutcome {
	active := relativeMarker(strings.ToLower(c.Statement))
	for _, s := range p {
		active = active || relativeMarker(strings.ToLower(s.Statement))
	}
	if !active {
		return schemaMatchOutcome{}
	}
	accept := func() schemaMatchOutcome { return schemaMatchOutcome{matched: true} }
	facts, ok := relativeFacts(p)
	if ok {
		if target, ok := relativeGround(c.Statement); ok && relativeAll(facts, target) && relativeUsed(p, target) {
			return accept()
		}
		if q, ok := relativeQuant(c.Statement); ok && q.kind == "some" {
			for _, name := range relativeNames(facts) {
				wants := []string{}
				for _, g := range q.guards {
					wants = append(wants, relativeFill(g, name))
				}
				body, ok := relativeGround(relativeFill(q.body, name))
				if ok && relativeAll(facts, append(wants, body...)) && relativeUsed(p, append(wants, body...)) {
					return accept()
				}
			}
		}
	}
	for i, s := range p {
		q, ok := relativeQuant(s.Statement)
		if !ok || q.kind == "some" {
			continue
		}
		rest := append(append([]NumberedArgumentStep{}, p[:i]...), p[i+1:]...)
		facts, ok := relativeFacts(rest)
		if !ok {
			continue
		}
		for _, name := range relativeNames(facts) {
			wants := []string{}
			for _, g := range q.guards {
				wants = append(wants, relativeFill(g, name))
			}
			if !relativeAll(facts, wants) || !relativeUsed(rest, wants) {
				continue
			}
			body := relativeFill(q.body, name)
			if q.kind == "no" {
				if equalClauses(c.Statement, "it is not the case that "+body) {
					return accept()
				}
				continue
			}
			if equalClauses(body, c.Statement) {
				return accept()
			}
			have, ok := relativeGround(body)
			want, ok2 := relativeGround(c.Statement)
			if ok && ok2 && relativeAll(have, want) {
				return accept()
			}
		}
	}
	return schemaMatchOutcome{}
}

func relativeEEName(first, existential NumberedArgumentStep) string {
	if !relativeMarker(normalizeClause(existential.Statement)) {
		return ""
	}
	q, ok := relativeQuant(existential.Statement)
	if !ok || q.kind != "some" {
		return ""
	}
	// A witness assumption states the full instantiated formula, including the
	// restrictor. Match both directions so no additional fact can be assumed.
	parts := []string{first.Statement}
	if pairs := conjunctionCandidates(first.Statement); len(pairs) > 0 {
		parts = []string{pairs[0].left, pairs[0].right}
	}
	var premises []NumberedArgumentStep
	for _, s := range parts {
		premises = append(premises, NumberedArgumentStep{Statement: s})
	}
	facts, ok := relativeFacts(premises)
	if !ok {
		return ""
	}
	for _, name := range relativeNames(facts) {
		wants := []string{}
		for _, g := range q.guards {
			wants = append(wants, relativeFill(g, name))
		}
		body, bodyOK := relativeGround(relativeFill(q.body, name))
		if fitchNamePattern.MatchString(name) && bodyOK && relativeAll(facts, append(wants, body...)) && relativeAll(append(wants, body...), facts) {
			return name
		}
	}
	return ""
}

func relativeNested(s string) bool {
	s = normalizeClause(s)
	if strings.Contains(s, " is such that ") || !relativePlain(s) {
		return false
	}
	// Two relative introducers inside the same nominal require a scaffold.
	for _, part := range strings.Split(s, " is ") {
		count := 0
		for _, w := range strings.Fields(part) {
			if w == "that" || w == "who" || w == "which" {
				count++
			}
		}
		if count > 1 {
			return true
		}
	}
	return false
}
