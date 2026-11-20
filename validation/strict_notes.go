package validation

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

// StrictNote explains a conservative catalog refusal. A note is attached only
// when disabling this rule alone certifies the failed step. It is not a verdict.
type StrictNote struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Since       string   `json:"since"`
	Why         string   `json:"why"`
	AlsoRefuses []string `json:"also_refuses"`
	Fix         []string `json:"fix"`
}

// strictRetryCounter is an optional test counter at the diagnostic retry entry.
var strictRetryCounter atomic.Pointer[atomic.Int64]

// strictStepPasses repeats the gate and its final sentence guard for one step.
// Other conclusions are available as cited statements, never re-evaluated here.
// It repeats the verdict-relevant part of gateConclusionStep (v44 name
// evidence, term slots and round-2 term checks) and skips its diagnostic
// retries, which change a failing step's message but never its verdict or form.
func (policy strictRules) strictStepPasses(step NumberedArgumentStep, steps map[int]NumberedArgumentStep) bool {
	if counter := strictRetryCounter.Load(); counter != nil {
		counter.Add(1)
	}
	policy.slotNames = schematicNameEvidence(steps)
	policy.termSlots = ruleTermSlots(step, steps)
	policy.pluralWitnessBlocked = rulePluralWitnessBlocks(step, steps)
	retry := policy.gateConclusionStepStrict(step, steps)
	if policy.round2TermFailure(step, steps, retry) != "" {
		return false
	}
	return retry.Passed && policy.sentenceFrameFailure(step, steps, retry.Form) == ""
}

// strictStepPassesEitherReading also tries the gate's second, postposed-
// conditional reading, retrying only this step. In particular this exposes
// postposed-if refusals.
func (policy strictRules) strictStepPassesEitherReading(step NumberedArgumentStep, steps map[int]NumberedArgumentStep) bool {
	return policy.strictStepPasses(step, steps) || policy.rewrittenStepPasses(step, steps)
}

// rewrittenStepPasses is the second, postposed-conditional reading alone.
func (policy strictRules) rewrittenStepPasses(step NumberedArgumentStep, steps map[int]NumberedArgumentStep) bool {
	rewritten := make(map[int]NumberedArgumentStep, len(steps))
	changed := false
	for id, original := range steps {
		if s, ok := policy.postposedConditional(strings.TrimSuffix(strings.TrimSpace(original.Statement), ".")); ok {
			original.Statement = s
			changed = true
		}
		rewritten[id] = original
	}
	return changed && policy.strictStepPasses(rewritten[step.Number], rewritten)
}

// strictNoteWorkers bounds the goroutines one notes pass uses for re-checks.
const strictNoteWorkers = 4

func attachStrictNotes(argument NumberedArgument, report *EntailmentGateReport) {
	if report.Passed {
		return
	}
	expanded, _, err := ExpandNumberedArgumentBridges(argument)
	if err != nil {
		return
	}
	steps := make(map[int]NumberedArgumentStep, len(expanded.Steps))
	for _, step := range expanded.Steps {
		steps[step.Number] = step
	}
	// One re-check per failing step and candidate rule. Each re-check has its
	// own policy copy, so they run concurrently; notes are attached afterwards
	// in catalog order, as the serial pass did.
	type recheck struct {
		report int
		step   NumberedArgumentStep
		note   int
		passes bool
	}
	var jobs []recheck
	for i := range report.Steps {
		failed := &report.Steps[i]
		if failed.Passed {
			continue
		}
		step, ok := steps[failed.Step]
		if !ok || failed.NearestForm == "scope" || failed.NearestForm == "explicit-variable FOL" {
			continue
		}
		// One recorded check with every rule on. A rule that refused nothing
		// in it leaves the check unchanged when switched off, so only the
		// rules that fired need a retry. If the check passes, every rule is
		// retried as before. The gate has already recorded the first reading
		// of this step (it failed); only the second reading is checked here.
		recorder, fired := allStrictRules().recording()
		var plainPasses bool
		if failed.fired != nil {
			for id := range failed.fired.ids {
				fired.ids[id] = true
			}
			plainPasses = recorder.rewrittenStepPasses(step, steps)
		} else {
			plainPasses = recorder.strictStepPassesEitherReading(step, steps)
		}
		for n, note := range strictNotes {
			if plainPasses || fired.ids[note.ID] {
				jobs = append(jobs, recheck{report: i, step: step, note: n})
			}
		}
	}
	workers := min(strictNoteWorkers, runtime.GOMAXPROCS(0), len(jobs))
	next := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range next {
				job := &jobs[j]
				job.passes = allStrictRules().without(strictNotes[job.note].ID).strictStepPassesEitherReading(job.step, steps)
			}
		}()
	}
	for j := range jobs {
		next <- j
	}
	close(next)
	wg.Wait()
	for _, job := range jobs {
		failed := &report.Steps[job.report]
		if !job.passes || len(failed.Notes) == 3 {
			continue
		}
		note := strictNotes[job.note]
		// Return owned slices: library clients cannot mutate the catalog.
		note.AlsoRefuses = append([]string(nil), note.AlsoRefuses...)
		note.Fix = append([]string(nil), note.Fix...)
		failed.Notes = append(failed.Notes, note)
	}
}

var strictNotes = []StrictNote{
	{
		ID:          "opener.frame",
		Name:        "Fronted frame",
		Since:       "v36.0",
		Why:         "A phrase before the subject may scope over the whole sentence: 'According to the report, A and B' attributes both claims, so A alone is not certified.",
		AlsoRefuses: []string{"1. If X is registered, then X is eligible.\n2. In 2019 the firm is registered.\n3. Therefore, in 2019 the firm is eligible. [1,2]"},
		Fix:         []string{"Move the time phrase after the predicate, including in the schema: \"If X is registered in 2019, then X is eligible in 2019.\""},
	},
	{
		ID:          "opener.adverb",
		Name:        "Fronted adverb",
		Since:       "v37.0",
		Why:         "A sentence adverb may weaken or qualify the whole claim: 'Arguably A and B' may assert only that A and B is arguable.",
		AlsoRefuses: []string{"1. If X is safe, then X is useful.\n2. Arguably reliable machinery is safe.\n3. Therefore, arguably reliable machinery is useful. [1,2]"},
		Fix:         []string{"Start with the subject: \"The reliable machine is safe.\" Use the same wording in the conclusion."},
	},
	{
		ID:          "coordination.clause",
		Name:        "Clause coordination after a verb",
		Since:       "v36.0",
		Why:         "In 'The observer saw A and B left', 'and' may join two objects or two whole clauses, so splitting off 'B left' is not certified.",
		AlsoRefuses: []string{"1. The pump malfunctioned.\n2. The valve stuck.\n3. Therefore, the pump malfunctioned and the valve stuck. [1,2]"},
		Fix:         []string{"Write \"Both the pump malfunctioned and the valve stuck.\""},
	},
	{
		ID:          "comma.exact",
		Name:        "Commas must match",
		Since:       "v36.0",
		Why:         "Commas do not group mixed and/or; use both...and, either...or or parentheses instead. Clauses must repeat their remaining punctuation exactly.",
		AlsoRefuses: []string{"1. Every device is small, light and portable.\n2. Alpha is a device.\n3. Therefore, Alpha is small, light, and portable. [1,2]"},
		Fix:         []string{"Repeat the clause with exactly the same commas wherever it is reused."},
	},
	{
		ID:          "identity.safe_position",
		Name:        "Identity substitution only in subjects",
		Since:       "v35.0",
		Why:         "Substituting co-referring names inside an attitude or other opaque context can change truth. The catalog therefore permits identity substitution only in identity sentences and plain clause subjects.",
		AlsoRefuses: []string{"1. Alpha is identical to Beta.\n2. The robot kicks Alpha.\n3. Therefore, the robot kicks Beta. [1,2]"},
		Fix:         []string{"State the substituted sentence as a premise: \"The robot kicks Beta.\" This supplies the missing claim; it does not prove it from the original premises."},
	},
	{
		ID:          "slot.indefinite_object",
		Name:        "Indefinite in an object position",
		Since:       "v39.0",
		Why:         "An indefinite object need not pick anything out: 'The robot is looking for a unicorn' does not establish that a unicorn exists.",
		AlsoRefuses: []string{"1. If the robot admires X, then X is admirable.\n2. The robot admires a hero.\n3. Therefore, a hero is admirable. [1,2]"},
		Fix:         []string{"Use a name in the slot: \"The robot admires Alpha.\" Conclude \"Alpha is admirable.\" This changes the premise to identify a particular hero."},
	},
	{
		ID:          "slot.indefinite_written",
		Name:        "Indefinite written before a variable",
		Since:       "v43.0",
		Why:         "When the rule writes 'a X', the instance need not be an article plus a noun: 'Bob greets a lot of people' would make X 'lot of people'. The catalog has no lexicon to tell that from 'a dog', so such a variable is refused in every position.",
		AlsoRefuses: []string{"1. If Rex is a X, then Fido is a X.\n2. Rex is a dog.\n3. Therefore, Fido is a dog. [1,2]"},
		Fix:         []string{"Write the phrase itself into the rule (\"If Rex is a dog, then Fido is a dog\"), or use a name or singular definite description."},
	},
	{
		ID:    "slot.indefinite_value",
		Name:  "Indefinite value that need not name one thing",
		Since: "v43.0",
		Why:   "The v43.0 shape check requires a lone indefinite pronoun or a/an, some or one plus modifiers and one noun with nothing after it. A quantity or complement phrase ('a lot of people', 'a friend of Bob', 'a few runners') need not be one thing. Since v44.0, slot.value_singular also refuses the remaining indefinite values.",
		AlsoRefuses: []string{
			"1. If X arrived, then X is someone who arrived.\n2. A friend of Bob arrived.\n3. Therefore, a friend of Bob is someone who arrived. [1,2]",
		},
		Fix: []string{"Write the phrase itself into the rule: \"If a lot of people arrived, then a lot of people is someone who arrived.\" The rule then claims exactly what it says; no variable is filled with the phrase."},
	},
	{
		ID:          "slot.value_singular",
		Name:        "Schematic value must be a singular term",
		Since:       "v44.0",
		Why:         "A quantity, plural, group or indefinite phrase (a lot, most people, Bob and Carol, a dog) need not be one thing, so a rule about each thing does not apply to it. A bare plural or mass term (dogs, water) has no fixed quantity and its reading can shift between premise and conclusion. Since v46.1, these values are refused even when allowlisted. Evidenced proper names and singular definite descriptions fill term slots. Predicate-only positions after a copula accept any value; variables also used as terms still require singular values.",
		AlsoRefuses: []string{"1. If X arrived, then X was seen.\n2. A dog arrived.\n3. Therefore, a dog was seen. [1,2]"},
		Fix:         []string{"Name the thing (Rex arrived), with internal capitals or capitalization evidenced somewhere other than sentence start or after Therefore/So/Thus/Hence; or use a singular definite description (the dog). For a plural or mass claim, rewrite with an explicit quantifier (some dogs, every dog); do not substitute a quantified phrase as a singular value. Alternatively state existence as its own premise and then name it, or write the phrase itself into the rule instead of a variable."},
	},
	{
		ID:          "slot.modified_quantifier",
		Name:        "Modified quantifier in a term slot",
		Since:       "v36.0",
		Why:         "A modified quantifier ('more than two robots', 'nearly all firms') is not an individual, so it cannot fill X or instantiate a universal.",
		AlsoRefuses: []string{"1. If X is red, then X is warm.\n2. Nearly every robot is red.\n3. Therefore, nearly every robot is warm. [1,2]"},
		Fix:         []string{"Use a singular definite description: \"The robot is red.\" Conclude \"The robot is warm.\" This certifies an individual instance rather than the quantified claim."},
	},
	{
		ID:          "slot.quantity",
		Name:        "Quantity expression in a term slot",
		Since:       "v38.0",
		Why:         "A quantity phrase ('two robots', 'a minority of voters', 'zero cases') may mean exactly that many, so it is not an individual for X or for instantiation. Since v43.0 the rule refuses only readings that are ambiguous ('exactly two'); collective quantity phrases such as 'a group of robots' are refused by slot.indefinite_value instead, so no step valid on every reading is refused by this rule alone.",
		AlsoRefuses: []string{},
		Fix:         []string{"Name the group: \"The team is registered.\" Conclude \"The team is eligible.\""},
	},
	{
		ID:          "term.clause_inside",
		Name:        "Clause inside a term",
		Since:       "v37.0",
		Why:         "A term that contains a finite verb with a complement may hide an attitude or embedded claim: 'the observer reckons Alpha is red' does not give that Alpha is red.",
		AlsoRefuses: []string{"1. If X are useful, then X are valuable.\n2. Software exploits not known to humans are useful.\n3. Therefore, software exploits not known to humans are valuable. [1,2]"},
		Fix:         []string{"Name the thing: \"The exploit is useful.\" Conclude \"The exploit is valuable.\""},
	},
	{
		ID:          "noun.pluralia_tantum",
		Name:        "Plural-only noun",
		Since:       "v37.0",
		Why:         "Some -s nouns are not plurals of the stripped word ('tongs' is not more than one 'tong'), so singular/plural matching is refused for them.",
		AlsoRefuses: []string{"1. All glasses are fragile.\n2. The cup is a glass.\n3. Therefore, the cup is fragile. [1,2]"},
		Fix:         []string{"Use the singular noun consistently: \"Every glass is fragile.\""},
	},
	{
		ID:          "conditional.postposed_if",
		Name:        "Postposed 'if' after a verb or adjective",
		Since:       "v37.0",
		Why:         "A verb or adjective before 'if' can introduce a whether-question rather than a condition; postposed conditionals are recognized only in conservative surface forms.",
		AlsoRefuses: []string{"1. The machine vibrates if the motor runs.\n2. The motor runs.\n3. Therefore, the machine vibrates. [1,2]"},
		Fix:         []string{"Write \"If the motor runs, then the machine vibrates.\""},
	},
	{
		ID:          "relation.object_extensional",
		Name:        "Instantiation into an object position",
		Since:       "v38.0",
		Why:         "The object of an unlisted verb may be intensional: awaiting every guest does not license the same substitutions as greeting every guest. Object instantiation is restricted to a closed verb list.",
		AlsoRefuses: []string{"1. The robot admires every hero.\n2. Alpha is a hero.\n3. Therefore, the robot admires Alpha. [1,2]"},
		Fix:         []string{"State the instance as a premise: \"The robot admires Alpha.\" This supplies the missing claim; it does not prove it from the original premises."},
	},
	{
		ID:          "unless.anaphor",
		Name:        "Pronoun across 'unless'",
		Since:       "v38.0",
		Why:         "A pronoun after 'unless' may point back into the main clause, so 'A unless it is dead' cannot be split into independent clauses.",
		AlsoRefuses: []string{"1. The alarm sounds unless this battery is dead.\n2. It is not the case that this battery is dead.\n3. Therefore, the alarm sounds. [1,2]"},
		Fix:         []string{"Replace \"this battery\" with \"the battery\" in both premises."},
	},
	{ID: "slot.description_shape", Name: "Controlled definite descriptions", Since: "v45.0", Why: "Description modifiers must be recognized adjectives or closed-list modifiers; names and finite clauses are refused.", AlsoRefuses: []string{"1. If X is red, then X is warm.\n2. The Bob dog is red.\n3. Therefore, the Bob dog is warm. [1,2]"}, Fix: []string{"Name it, or use 'the dog'."}},
	{ID: "slot.one_type", Name: "One type per rule variable", Since: "v44.0", Why: "A clause-slot occurrence must not smuggle a clause into another occurrence used as a term.", AlsoRefuses: []string{"1. If when X is red bob smiles, then X is warm.\n2. When the dog is red bob smiles.\n3. Therefore, the dog is warm. [1,2]"}, Fix: []string{"Name the thing in every occurrence, or write the phrase into the rule."}},
	{ID: "eg.subject_only", Name: "Subject witnesses only", Since: "v44.0", Why: "An object or passive subject may occur in an intensional position without asserting its existence.", AlsoRefuses: []string{"1. Lois greets Zeus.\n2. Therefore, there is an X such that Lois greets X. [1]"}, Fix: []string{"State the existential claim as its own premise."}},
	{ID: "eg.name_witness", Name: "Named existential witnesses", Since: "v44.0", Why: "A definite may covary with another quantifier; descriptions do not fix one witness.", AlsoRefuses: []string{"1. The dog is red.\n2. Therefore, something is red. [1]"}, Fix: []string{"State the existential claim as its own premise, or use an evidenced proper name in an unquantified subject clause."}},
	{ID: "term.numeral", Name: "No numeral-led terms", Since: "v44.0", Why: "A numeral phrase can state an exact quantity rather than identify one constant.", AlsoRefuses: []string{"1. All rodents are animals.\n2. Six mice are rodents.\n3. Therefore, six mice are animals. [1,2]"}, Fix: []string{"Name the thing, or use 'the mouse'."}},
	{ID: "identity.constant_sides", Name: "Constant identity sides", Since: "v44.0", Why: "Pronouns can shift referent and quantities need not identify one object.", AlsoRefuses: []string{"1. He is identical to Alpha.\n2. He is identical to Beta.\n3. Therefore, Alpha is identical to Beta. [1,2]"}, Fix: []string{"Name the thing, or use 'the dog' on each side."}},
	{ID: "identity.extensional_context", Name: "Closed extensional identity contexts", Since: "v44.0", Why: "An unlisted predicate may describe an intensional role under which co-referring names cannot be exchanged.", AlsoRefuses: []string{"1. The man is identical to Clark.\n2. The man is famous as a hero.\n3. Therefore, Clark is famous as a hero. [1,2]"}, Fix: []string{"State the substituted claim as a premise."}},
	{ID: "slot.letter_in_name", Name: "Letters adjacent to names", Since: "v44.0", Why: "A schematic letter may be a literal component of Model U or Malcolm X.", AlsoRefuses: []string{"1. If Model U is red, then Model U is warm.\n2. Model U is red.\n3. Therefore, Model U is warm. [1,2]"}, Fix: []string{"Use a variable away from names, or write the name out."}},
	{ID: "term.inflection_verbatim", Name: "Verbatim noun inflection", Since: "v44.0", Why: "A plural surface may not denote instances of the stem guessed by deleting its suffix.", AlsoRefuses: []string{"1. All dogs are wet.\n2. The dog is a dog.\n3. Therefore, the dog is wet. [1,2]"}, Fix: []string{"Write the rule with singular 'every N' (Every dog is wet)."}},
	{ID: "scope.mixed_andor", Name: "Mixed connective scope", Since: "v44.0", Why: "A and B or C may mean A and (B or C) or (A and B) or C; relative clauses and lowercase names do not disambiguate it.", AlsoRefuses: []string{"1. Ann is a singer who dances and Bob is happy or Cal reads.\n2. Therefore, Ann is a singer who dances. [1]"}, Fix: []string{"Use 'both … and' or 'either … or' to show the grouping."}},
	{ID: "scope.clause_start", Name: "Controlled clause starts", Since: "v44.0", Why: "Each inspected clause must start with a determiner, proper name, pronoun or numeral. An unrecognized start may be an operator with wider scope.", AlsoRefuses: []string{"1. It is not the case that perhaps the lamp is lit and the door is open.\n2. Therefore, either it is not the case that perhaps the lamp is lit or it is not the case that the door is open. [1]"}, Fix: []string{"Move the operator into each clause, or state the claim as an explicit premise."}},
	{ID: "opener.ing_frame", Name: "Initial -ing frame", Since: "v44.0", Why: "Following the script may qualify the entire conjunction rather than name its subject. Even genuine gerund subjects are refused when inspected.", AlsoRefuses: []string{"1. It is not the case that it is not the case that swimming is healthy.\n2. Therefore, swimming is healthy. [1]"}, Fix: []string{"Move the frame into each clause, or state the framed claim as its own premise."}},
	{ID: "scope.elided_copula", Name: "Closed copula elision", Since: "v44.0", Why: "In Either the men are quiet or fish, fish may be a verb rather than a predicate with an omitted copula. Unknown predicates are not expanded.", AlsoRefuses: []string{"1. Either the men are quiet or fish.\n2. Therefore, either the men are quiet or the men are fish. [1]"}, Fix: []string{"Repeat the subject and verb: 'or the men are fish'."}},
	{ID: "scope.description_negation", Name: "Description negation scope", Since: "v44.0", Why: "Wide negation need not assert that a description denotes; narrow predicate negation can presuppose existence. Only proper names permit the conversion.", AlsoRefuses: []string{"1. It is not the case that the present king of France is bald.\n2. Therefore, the present king of France is not bald. [1]"}, Fix: []string{"Add 'There is exactly one present king of France' and name it, or keep the wide negation."}},
	{ID: "opener.ly_exception", Name: "Adverb-capable noun exceptions", Since: "v44.0", Why: "Daily, early and only can be adverbs with scope over an entire sentence, even when they also have nominal uses.", AlsoRefuses: []string{"1. It is not the case that it is not the case that daily the lamp is lit.\n2. Therefore, daily the lamp is lit. [1]"}, Fix: []string{"Move the adverb into each clause, or state the framed claim as its own premise."}},
	{ID: "slot.plural_witness", Name: "No repeated plural or mass witness", Since: "v46.0", Why: "Water is in the cup and water is in the bottle does not establish the same water. Repeated antecedent occurrences, multiple fact citations, and consequents containing same, both or together cannot reuse a bare plural or mass witness.", AlsoRefuses: []string{"1. If X is in the cup, then X is in both vessels.\n2. Water is in the cup.\n3. Therefore, water is in both vessels. [1,2]"}, Fix: []string{"Name each portion or state the conclusion as a premise."}},
	{
		ID:          "term.pronoun_cross_statement",
		Name:        "Personal pronouns across statements",
		Since:       "v46.2",
		Why:         "a pronoun's referent depends on context and can differ between statements",
		AlsoRefuses: []string{"1. If it rains, then it is wet.\n2. It rains.\n3. Therefore, it is wet. [1,2]"},
		Fix:         []string{"replace the pronoun with the name or noun phrase it refers to, e.g. 'the card'"},
	},
	{
		ID:          "term.possessive_cross_statement",
		Name:        "Possessive determiners across statements",
		Since:       "v46.2",
		Why:         "a possessive determiner's referent depends on context and can differ between statements",
		AlsoRefuses: []string{"1. If John helps his wife, then the alarm sounds.\n2. John helps his wife.\n3. Therefore, the alarm sounds. [1,2]"},
		Fix:         []string{"replace 'his wife' with a name or a phrase that does not depend on a pronoun"},
	},
}
