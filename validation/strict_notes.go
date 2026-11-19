package validation

import (
	"runtime"
	"strings"
	"sync"
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

// strictStepPasses repeats the gate and its final sentence guard for one step.
// Other conclusions are available as cited statements, never re-evaluated here.
// It calls the strict check directly: gateConclusionStep adds only diagnostic
// retries, which change a failing step's message but never its verdict or form.
func (policy strictRules) strictStepPasses(step NumberedArgumentStep, steps map[int]NumberedArgumentStep) bool {
	retry := policy.gateConclusionStepStrict(step, steps)
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
	defer func() {
		// The recorded rules are for this pass only; reports never carry them.
		for i := range report.Steps {
			report.Steps[i].fired = nil
		}
	}()
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
		AlsoRefuses: []string{"1. If X is safe, then X is useful.\n2. Highly reliable machinery is safe.\n3. Therefore, highly reliable machinery is useful. [1,2]"},
		Fix:         []string{"Start with the subject: \"Machinery that is highly reliable is safe.\" Use the same wording in the conclusion."},
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
		Why:         "A comma can change grouping ('A, and B or C' differs from 'A and B or C'), so clauses must repeat their punctuation exactly.",
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
		Fix:         []string{"Drop the article and let the variable stand for the whole phrase: \"If Rex is X, then Fido is X\" with X = \"a dog\". This works only when the phrase is an article, modifiers and one noun; otherwise write the phrase itself into the rule."},
	},
	{
		ID:    "slot.indefinite_value",
		Name:  "Indefinite value that need not name one thing",
		Since: "v43.0",
		Why:   "An indefinite value is accepted only as a lone 'someone' or as a/an, some or one plus modifiers and one noun with nothing after it. A quantity or complement phrase ('a lot of people', 'a friend of Bob', 'a few runners') need not be one thing. Known limit: a bare idiom such as 'a lot' still passes.",
		AlsoRefuses: []string{
			"1. If X arrived, then X is someone who arrived.\n2. A friend of Bob arrived.\n3. Therefore, a friend of Bob is someone who arrived. [1,2]",
		},
		Fix: []string{"Write the phrase itself into the rule: \"If a lot of people arrived, then a lot of people is someone who arrived.\" The rule then claims exactly what it says; no variable is filled with the phrase."},
	},
	{
		ID:          "slot.modified_quantifier",
		Name:        "Modified quantifier in a term slot",
		Since:       "v36.0",
		Why:         "A modified quantifier ('more than two robots', 'nearly all firms') is not an individual, so it cannot fill X or instantiate a universal.",
		AlsoRefuses: []string{"1. If X is red, then X is warm.\n2. Nearly every robot is red.\n3. Therefore, nearly every robot is warm. [1,2]"},
		Fix:         []string{"Use a name in the slot: \"Alpha is red.\" Conclude \"Alpha is warm.\" This certifies an individual instance rather than the quantified claim."},
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
		Fix:         []string{"Name the thing: \"The exploits are useful.\" Conclude \"The exploits are valuable.\""},
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
}
