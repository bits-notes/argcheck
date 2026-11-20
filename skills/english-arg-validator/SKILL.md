---
name: english-arg-validator
description: Check a numbered argument for syntactically transparent validity with the deterministic `english-arg-validator` CLI (offline Go checker, no LLM). Use whenever drafting, checking, or repairing a premises-to-conclusion argument: write in the gate's numbered format, run `english-arg-validator`, read any step mismatch, and repair until it passes while keeping fidelity to the source text.
---

# english-arg-validator — deterministic validity check for numbered arguments

`english-arg-validator` runs two deterministic checks in Go (`numbered_argument_syntax`, then
the catalog `v45.1` entailment gate) with no LLM and no network access.

**Validity is syntactic: the gate is the standard. Never overrule a gate failure
with your own judgment that a step "obviously follows."** If you think a step is
valid, rephrase it into a catalog form or make the missing link a visible
premise. If no phrasing works, report that the inference is outside the catalog
rather than calling the step valid.

**Fidelity to the source is your job, not the gate's.** `english-arg-validator` checks only
whether each conclusion step follows syntactically from its cited steps; it
never checks whether premises are true or faithful to a text.

## Run it

```bash
english-arg-validator <<'EOF'
1. If rain falls, then the street is wet.
2. Rain falls.
3. Therefore, the street is wet. [1,2]
EOF
```

- `english-arg-validator file.txt` or stdin (`english-arg-validator < file.txt`, `cat file.txt | english-arg-validator -`).
- `--json` prints structured syntax, step reports, and provenance.
- `--quiet` prints only `VALID` or `NOT CERTIFIED`.
- `--forms` lists all 121 accepted inference forms in catalog `v45.1`.
- `--version` prints the catalog version and build stamp (`gate catalog v46.5; local english-arg-validator@<commit>`).
- Exit codes: `0` valid, `1` not certified, `2` input or operational error.
- `NOT CERTIFIED` means some step matched no accepted rule. It does not mean the
  argument is invalid; repair the step or add the missing premise.
- If `english-arg-validator` is not installed yet, install with
  `go install github.com/bits-notes/english-arg-validator/cmd/english-arg-validator@latest` or run
  `./install.sh` from a checkout of the repository.

Output reports each conclusion step as `ok: <form>` or
`FAIL: nearest form <form>: <exact mismatch>`. When a conclusion fails only
because it cites an extra step that does no work, a `hint:` line names the form
that passes on the used subset and the inert citation(s) to remove.

## Format (strict)

- **Premise:** `1. Statement.` — one line ending in a period, with no
  inference marker and no trailing citation brackets.
- **Conclusion:** `3. Therefore, statement. [1,2]` — requires trailing citations to
  earlier steps. The leading `Therefore`, `So`, `Thus`, or `Hence` is optional,
  case-insensitive, and accepts an optional comma. A marker without terminal
  citations is a syntax error. Inline references remain premises. Re-rendered
  steps use canonical `Therefore,`.
- **Numbering and layout:** Sequential from `1`, with no blank lines, no leading
  or trailing spaces, no code fences, and no inline commentary. The final step
  must be a conclusion.
- **Connectedness:** Every premise and every scoped step must support a later
  step (no orphan steps). Root-level premises may appear after intermediate
  conclusions as long as every citation points to an earlier step.
- **Postposed conditional:** `Y if X.` (or `Y, if X.`) reads as `If X, then Y.`. Not for `only if`, `even if`, `as if`, embedded questions, or a consequent with `and`/`or`.
- **Inline bridge premise:** `If [1], then C.` or `If [1,2], then C.` expands the
  referenced steps verbatim; `If A, then [1].` works too. References must fill a
  whole antecedent or consequent (not under negation or mixed with text); write
  `[1]`, not `(1)`. See "Step references" in `docs/rules.md`.
- **Fitch subproofs:** Prefix each active depth level with `| `. Open an
  assumption with `Suppose X.` (or `Let a be an arbitrary F.` /
  `Let witness be arbitrary.` for universal generalization). Discharge a
  subproof by citing its full range `[i-j]`; a reductio subproof ends with
  `Therefore, contradiction. [i,j]`. Steps outside a closed subproof cannot cite
  individual lines inside it.

```text
1. If rain falls, then the street is wet.
2. If the street is wet, then the cars slow.
| 3. Suppose rain falls.
| 4. Therefore, the street is wet. [1,3]
| 5. Therefore, the cars slow. [2,4]
6. Therefore, if rain falls, then the cars slow. [3-5]
```

## The repair loop

1. **Draft** the argument in the strict numbered format and run `english-arg-validator`.
2. **Read each `FAIL`:** the mismatch quotes the exact clauses or bindings that
   failed to match. Fix the wording (align clauses verbatim, use canonical
   negation, or split a multi-hop jump into one inference per step) or fix the
   citations (drop inert citations named in `hint:`, add missing premises).
3. **Supply missing links visibly:** if the inference relies on an unstated
   premise, add it as an explicit premise and tell the user it was supplied.
   Never fold an unstated assumption silently into an existing premise.
4. **Re-run until `VALID`** and include both the final numbered argument and the
   version stamp line when reporting the result.

**Source fidelity while repairing:** when reconstructing an author's text, keep
every substantive qualifier and scope restriction the author wrote. Change only
surface syntax needed by the gate (word order, canonical connectives, pronoun
expansion). Leave hedges outside the argument and label them. Passing the gate
by watering down or overstating the source is a failure, not a fix. For the
full procedure (sentence ledger, anchored and supplied premises, bridge
premises, fidelity pass), see `skills/formalize-from-source/SKILL.md`.

## Passing is not the same as reconstructed

`If P1 and P2 and P3, then C.` plus `P1`–`P3` always passes as conjunctive modus
ponens, putting all the disputed weight into one ad hoc conditional. An argument
built out of `If [my premises], then [my conclusion]` bridges is a sufficiency
squish, not a reconstruction. Replace ad hoc bridges with the **general rule**
the author relies on (for example, `Every interval procedure that misses its
advertised coverage is unfit for reporting uncertainty.`), then instantiate it.

## Writing so the gate can see the inference

Clauses match **verbatim** after light normalization (case, whitespace,
contractions, optional `then`; `it's (not) the case that` = `it is (not) the case that`,
with straight or curly apostrophes, while `it's raining` stays an ordinary clause). Synonym and paraphrase tolerance are permanently
out of scope: if two wordings mean the same thing, state that equivalence as a
visible identity or conditional premise.

- **Pronouns across statements (v46.2):** matched clauses, terms, and schema
  values cannot contain he, she, it, they, him, her, them, himself, herself,
  itself, themselves, or possessive his, her, their, its. Matching is
  case-insensitive and whole-token. Only the closed v46.4 dummy-it allowlist is exempt (see docs/coverage.md); otherwise
  replace it with a name or noun phrase. There is no sole-antecedent exception.
  Pronouns in unused clauses or checks confined to one statement remain allowed.
  Logical scaffolds such as "it is not the case that" consume "it" as syntax.
  Refusals carry `term.pronoun_cross_statement` or
  `term.possessive_cross_statement` in strict notes.
- **Repeat clauses word for word:** `the street is wet` does not match
  `the road is wet`.
- **Negation:** write `It is not the case that X.` Modus tollens fails on
  `The street is not wet` against `the street is wet`, and passes on
  `It is not the case that the street is wet`.
- **Universals and syllogisms:** `Every man is mortal.` + `Socrates is a man.`
  passes; `All men are mortal.` fails against singular `a man`. For class
  syllogisms, `Every A is a B.` / `All A are B.` and `No B is C.` work.
- **Connectives:** disjunction is `Either X or Y.`; conjunction is `X and Y.` (or `Both X and Y.` when X ends in a verb phrase);
  biconditional is `X if and only if Y.` or parenthesized `(X) if and only if (Y)`
  (optionally prefixed by `It is the case that`) when either side contains
  another connective. `It holds that` is ordinary text.
- **One hop per step:** chain intermediate conclusions and cite only the steps
  used by that single rule. Probe unfamiliar shapes with a 3-line test argument
  before writing a long chain.
- **"X rather than Y" conclusions:** state the dichotomy premise explicitly
  (`Either we should use X or we should use Y.`), refute one side, and apply
  disjunctive syllogism.

### Strict conventions (catalog v35–v46.5)

The gate refuses phrasings that are ambiguous in English, even when the
inference is fine. Each refusal names its fix; apply exactly that fix.

- **Subject first.** No fronted adverb, `-ly` word or prepositional frame
  before the subject (`Clearly, ...`, `Of course ...`, `In theory ...`,
  `Under the treaty, ...`, `According to the report, ...`), with or without a
  comma. Move timing and place after the verb (`The firm cut prices in 2019.`).
  Keep an attribution as the subject: `The report says that the pump failed.`,
  plus `If the report says that the pump failed, then the pump failed.` when
  the bare claim is needed.
- **Clause coordination.** Join whole clauses with `Both A and B`,
  `Either A or B`, or parentheses `(A) and (B)`. A bare `A and B` after a verb
  phrase may be read as one coordinated object and is refused. Commas count:
  `A, and B` and `A and B` are different clauses, so repeat one form exactly.
- **Identity.** `Superman is identical to Clark.` substitutes only where the
  name is the subject of a clause. State any other substituted sentence as its
  own premise.
- **Schematic slots hold one singular thing (v44).** A variable in a term
  position (subject, object, object of a preposition, either side of an
  identity) takes only: a proper name (`Alpha`, `AIXI`); `the` + at most one
  modifier + a singular noun (`the pump`, `the main valve`); a name's
  possessive (`Bob's dog`); or `the N of D` (`the death of the child`). Not
  `a hero`, `some firm`, `someone`, bare nouns (`water`, `humanity`), plurals,
  quantities or numerals (`many firms`, `two robots`), coordinations
  (`Alice and Bob`), partitives, nested `of`, or longer descriptions.
- v46.5: `term.indefinite_singular` refuses a/an noun phrases as cross-statement constants: use a name, every N, or some N. Predicates such as “Rex is a guard” remain allowed.
- **Names need evidence away from sentence start.** A capital that appears
  only at the start of a sentence (or after `Therefore,`) does not show that a
  word is a name, so `Alpha is valuable.` alone cannot fill `X`. Make sure the
  name also appears mid-sentence somewhere in the argument
  (`The firm owns Alpha.`), or use `the N`. Internal capitals (`AIXI`,
  `McCarthy`) count as evidence anywhere.
- **Predicates are free.** After `is`/`are`/`was`/`were` a variable is a
  predicate and takes any phrase: `If Rex is X, then Fido is X.` with
  `X` = `a lot of fun`. A variable used as a term anywhere in the rule is held
  to the term rule everywhere.
- **Repair order for a refused value:** (1) write the phrase itself into the
  rule with no variable (`If humanity is valuable, then humanity should be
  protected.`); (2) name the thing or use a short `the N`; (3) for a claim
  about a kind, use `Every N is P.` with a named or `the N` instance.
- **Never write `a X` / `some X` in a rule.** Let the variable stand for the
  whole predicate (`If Rex is X, then Fido is X.`) or write the phrase in.
- **Classes match instances by spelling.** `Every human is mortal.` +
  `Socrates is a human.` passes; `All humans are mortal.` with
  `Socrates is a human.` does not (plural never matches singular).
- **Clause starts.** In a step that splits, swaps or distributes `and`/`or`
  (conjunction or disjunction elimination, commutation, De Morgan), each
  operand must begin with a determiner (`the`, `a`, `every`, `some`, `no`), a
  pronoun, a numeral, a name, a possessive, `there` + be, or
  `it is not the case that`. After a capitalised name the next word must be a
  common verb (`is`, `greets`, `believes`, `claims`, `requires`, …); for an
  unusual verb, declare an `atom{...}` literal. Bare nouns, adjectives,
  `-ing` words and adverbs are refused: write `The relation of numerical
  identity is transitive.`, `The act of signing up for cryonics is costly.`
  Simplest habit: start every clause this way. A sentence-initial `-ing`
  word is always refused.
- **Existential-elimination witnesses** (and constants in explicit FOL) should
  be capitalised fresh names (`Suppose Cyd is a student and Cyd greets Bob.`),
  not single letters or lowercase words.
- **Negating a description:** converting between `The F is not G` and
  `It is not the case that the F is G` works only for names.
- **Existential generalisation** works only from a non-passive subject, with
  a name as the witness; otherwise state the existential claim as a premise.
- **Objects of verbs.** `Lois admires every hero.` does not give
  `Lois admires Bruce.` for most verbs, since the object of `admires`, `seeks`,
  `wants` may not pick anything out. State the instance as a premise.
- **When the message says "state ... as a premise",** add that line as a
  premise and keep the rest of the argument unchanged.

### Schematic rules and definitions

- **Placeholders:** `U`, `V`, `W`, `X`, `Y`, and `Z` bind terms or whole clauses
  across up to **10 antecedent conjuncts** (and 10 cited premises). Failed
  instantiations report the first unmatched conjunct and current binding, e.g.
  `conjunct "X is sane" has no matching unused cited premise under binding {X="ada"}`.
  Above 10 conjuncts, split the rule into a chain.
- **Agreement and parsing traps:** placeholders must agree in number with their
  bindings (`If X are fixed, ...` for plural subjects). Copular sentences split
  at the first ` is `, so name any subject containing an embedded `is` via an
  identity premise rather than embedding it directly. Keep `such as ...`
  examples outside premises so they do not block placeholder binding.
- **Definitions as identities:** write `T is identical to D.` or
  `T is identical with D.` (bare `is` is predication, not identity). Whole
  gerund phrases (`Standing as an equal is identical to being recognized as a peer.`)
  and coordinated proper names substitute in transparent argument positions,
  replacing every occurrence of the term. Substitution is refused inside opaque
  contexts (`should`, `must`, `can`, attitudes, quotations, `that`-clauses);
  bridge opaque contexts with a visible schematic premise if needed.

### Relations, relative clauses, explicit FOL, counting, and descriptions

- **Relations (`TERM VERB TERM` and n-place):** quantifiers (`every`/`some`/`no`)
  work in subject, object, and prepositional slots (`to`, `from`, `with`, etc.).
  `Alice gave Bob a book.` and `Alice gave a book to Bob.` share one 3-place
  atom; active and passive voice remain distinct. Readable sentences are capped
  at **4 quantifiers**, and every scope reading must validate; disambiguate
  scope with `is such that` (`Some teacher is such that every student admires that teacher.`).
- **Relative clauses:** `F that a Rs` and `F who Rs a` (`that`/`who`/`which`)
  decompose into `F(x) ∧ R(...)` over the supported verb vocabulary, with the
  filled relation matched verbatim. Use `is such that` to nest relative clauses.
- **Explicit-variable FOL and notation equivalence:** for deep quantifier
  alternations or >4 quantifiers, write `every x is such that BODY`,
  `some x is such that BODY`, or `no x is such that BODY` using bare or indexed
  variables (`x`, `x1`, `y12`) and parenthesized compound scopes. Use
  `Let witness be arbitrary.` for unrestricted universal generalization.
  `notation equivalence` converts in one step between an unambiguous readable
  sentence and its explicit-variable form; `name{x}` / `name{"..."}` and
  `atom{"..."}` preserve exact lexical boundaries inside explicit bodies.
- **Counting quantifiers (`1`–`5` or `one`–`five`):** `at least n F`,
  `at most n F`, `exactly n F`, `no F other than a ...`, and `only a is ...`
  expand to FOL with identity. Uniqueness (`no person other than Scott ...`)
  requires citing class membership (`Scott is a person.`) alongside the relation.
- **Russellian definite descriptions:** `The F is G` expands to existence,
  uniqueness (`at most one F`), and predication, supporting substitution,
  introduction, elimination, uniqueness, and negation. Parenthesize compound
  operands `(P) and (Q)` in the description fallback. Under negation, `The F is not G`
  is checked on both wide and narrow readings in the description fallback (while
  preserving historical narrow-to-surface certificates); use
  `It is not the case that the F is G` to fix the narrow reading.
- **Counterfactuals and properties:** conditionals with `would` in the
  consequent use dedicated counterfactual rules (ordinary material rules cannot
  conclude or negate them; only auxiliaries are mood-stripped, e.g. `had been struck`
  vs `was struck`). Copular predicates convert via
  `a is F.` ⟺ `a has the property of being F.` and support property inclusion
  (`a has every property that b has.`), distinguishing properties, and Fitch
  property generalization (`Let P be an arbitrary property that b has.`).

## Contradictions between encodings vs contradictions between authors

Two passing encodings that contradict each other do not establish that two
authors disagree. Before claiming an author-level contradiction:

- Verify each contradicting step against quotes from the source text (exact
  quantifiers, scope restrictions, and terms). Apparent contradictions regularly
  come from dropped qualifiers, vocabulary aligned across authors without textual
  warrant, or a devil's-advocate persona encoded as the author's voice.
- Align vocabulary across two arguments only when both sources support it.
- Keep hypothetical or persona arguments labeled and separate from the author's
  own commitments.

## Coverage and limits

Coverage: complete for propositional logic and for function-free first-order
logic with identity, within a controlled English (sentences with one reading); a
bounded finite-model oracle tests that the checker never accepts an invalid
argument; this is not proved.

What `english-arg-validator` does not do:

- **Premise truth:** it never judges whether premises are true.
- **Source fidelity:** it never checks whether an argument matches a source text.
- **Proof search or paraphrase:** it checks only the steps written, without
  synonym tolerance, general arithmetic (`n > 5`), modal systems beyond T,
  proportional quantifiers (`most`), or non-deductive inference.

See `docs/rules.md` in the `english-arg-validator` repository for all 121 forms with accepted
and rejected examples, and `docs/coverage.md` for full grammar boundaries.

## Make and/or scope explicit

Never mix top-level `and` and `or` without grouping, including inside `If ...,
then ...` or `every x is such that ...`. The syntax diagnostic identifies the
line and sentence and asks you to use `either...or / both...and, or parentheses to group`.
Parentheses group whole clauses and must balance; `[n]` remains a step reference.
Keep parentheses in source quotations and repairs.

- `(Ann sings and Bob dances) or Cal reads` is a disjunction and does not
  license extracting `Ann sings`.
- `Ann sings and (Bob dances or Cal reads)` and `Both Ann sings and either Bob dances
  or Cal reads` are conjunctions and do license that extraction.
- Leading `either`/`both` selects its matching main connective; its second
  operand runs to the end. Each operand must itself have unambiguous scope.
- A `both ... and` frame does not group an unmarked inner `or`; add inner
  `either ... or` or parentheses. Conversely, inside `either ... or`, group
  an inner `and` with `both ... and` or parentheses. Apply this recursively.
- Since v46.2, commas never fix mixed and/or scope. `Ann sings, and Bob dances
  or Cal reads` is refused. Use both/either or parentheses. Single-connective
  lists retain their existing behavior.

Opaque complements, relative clauses, `unless`, quoted text, `whether or not`,
and noun-phrase coordination are conservatively exempt. The exemption can
also hide ambiguity in an outer connective in the same clause; do not treat
syntax acceptance as evidence of a transparent internal logical structure.
See `docs/rules.md` for executable examples and the scope limitation.
Existing identifiers such as `definition (a)` and parenthetical annotations
inside atomic clauses retain their opaque reading. Only whole-clause boundary
parentheses create logical groups; parentheses must always balance.


### Read back the certified reading (v46.3)

Use `--paraphrase` or each passed step's JSON `paraphrase` and
`paraphrase_form`. Read every step's paraphrase back against the source sentence
it formalizes. A changed quantifier, scope, hedge, or referent is a fidelity
error to fix even though the step certified. Semi-formal renderings preserve
opaque lexical slots and the actual rule/formula; do not treat an opaque slot
as an independently parsed English claim. Failed steps add a structured
`refusal` with the existing fix and rune offsets into the identified statement.
