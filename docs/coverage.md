# Entailment Gate — Coverage and Known Limits (catalog v22)

**Type:** reference
**Status:** current as of catalog v22 (2026-11-10)


The gate checks whether each step of a numbered argument is an instance of a
catalog form, with clauses matched verbatim. Any valid argument can be made to
pass by stating the missing logic as a visible premise (a bridging conditional,
an identity, a schematic rule). This page lists which inferences the gate
**verifies itself** and which ones it can only accept as premises. Update it
whenever a catalog version changes coverage.

## What is covered

- **Propositional logic, completely.** Every valid argument built from if,
  and, or, not and if-and-only-if is derivable with the one-step forms (v0–v13)
  plus Fitch subproofs (v14): conditional proof, reductio, proof by cases.
  Explosion and excluded middle are derivable in a few steps; they are refused
  only as one-step jumps from unrelated premises.
- **Monadic and flat n-place first-order logic.** every/some/no over one-place
  predicates, two-place relations with a quantifier on either side and both
  scope readings checked (v17), extended to n-place atoms (v22), universal generalization and existential
  elimination with fresh-name checks (v14/v17).
- **Identity:** substitution (blocked inside that-clauses and modal/deontic
  contexts), symmetry, reflexivity.
- **Basic modal logic (system T):** possibility introduction, duality,
  distribution.
- **Counterfactuals (v19):** the Lewis VW rule set, with material rules
  blocked wherever they are invalid for counterfactuals.
- **Properties (v20–v21):** `a has the property of being F`, instantiation,
  "some property", shared and distinguishing properties, Leibniz's law in both
  directions, property generalization by subproof. Induction is usable as a
  schematic premise (`If zero is X and the successor of every number that is X
  is X, then every number is X.`).

## Permanent non-goals

These will not be added. The workaround is always a visible premise.

| Limit | Example that fails | Why it stays out | Workaround |
|---|---|---|---|
| Paraphrase and synonymy | `the road is wet` vs `the street is wet` | Verbatim matching is what makes the gate transparent; paraphrases require explicit premises | Premise stating the two are the same (identity or biconditional) |
| Properties of properties | `Socrates has every virtue.` + `Being courageous is a virtue.` ⊢ `Socrates is courageous.` | Bearers must be individuals; letting properties bear properties reintroduces Russell's paradox | Read "virtue" as an ordinary noun (`Courage is a virtue.` ⊢ `Socrates has courage.` passes), plus one premise per virtue linking `has courage` to `is courageous` |
| "Most" and other proportional quantifiers | `Most voters want X.` + `Most voters want Y.` ⊢ `Some voter wants X and Y.` | Not expressible with every/some/no at all | State the step as a conditional premise and apply modus ponens |
| Non-deductive inference | inductive, probabilistic, best-explanation arguments | The gate certifies deductive validity only | Write the inductive link as a premise (`If the evidence is E, then probably H.`) and the gate checks the structure around it |

## Not built yet (would fit the current design)

| Limit | Example that fails | Workaround today |
|---|---|---|
| Relative clauses inside relations | `Clark Kent admires every hero that Superman admires.` + `Batman is a hero that Superman admires.` ⊢ `Clark Kent admires Batman.` | Restate as a copular universal (`Every hero that Superman admires is a hero that Clark Kent admires.`) |
| Definite descriptions with uniqueness | `Exactly one person wrote Waverley.` + `Scott wrote Waverley.` ⊢ `No person other than Scott wrote Waverley.` | Identity premises (`Scott is identical to the author of Waverley.`) for substitution; uniqueness as a premise |
| Counting and arithmetic | `Three senators voted yes and two voted no, so at least five voted.` | State the arithmetic as a premise |
| Tense and ordering | `It rained before the meeting.` + `The meeting was before lunch.` ⊢ `It rained before lunch.` | State transitivity of "before" as a premise |
| Verb predicates as properties | `Superman flies.` ⊢ `Superman has the property of being ...` | Write `Superman is someone who flies.` (copular) or add a premise |
| Counterfactual subproofs, `might` rules, bare `would V` | `If the match were struck, then the match would light.` has no plain form; `might` gets no rules | Bridging premises; `had gone` vs `went` also needs one |
| Modal logic beyond T | `Necessarily P.` ⊢ `Necessarily, necessarily P.` (S4) | Premise; which system holds is itself a substantive claim |
| Proving rules over predicates | concluding `If Plato is X, then Socrates is X.` (it can be used, not derived) | Use `has every property that` with property generalization (v20–v21), or assume the schema |
| Other property wordings | `Being just is a property that Socrates has.` | Use the canonical `Socrates has the property of being just.` |

## Out of scope for the gate

- **Soundness:** the gate never judges whether premises are true.
- **Fidelity:** the gate never checks that an argument matches its source text.
- **Bridge squish:** an ad hoc `If [premises], then [conclusion]` always passes.
  Passing shows validity of the encoding, not that the reasoning was
  reconstructed.

## v22: n-place relations

Flat relations may have three, four, or more ordered argument positions. A
preposition separates each additional object position, for example
`Alice sends Parcel to Bob from Depot`. The verb and the ordered prepositions
must match verbatim. Changing an argument's role, dropping a position, changing
a preposition, or changing active to passive voice does not preserve the atom.

Universal instantiation, existential introduction, identity substitution, and
Fitch universal generalization and existential elimination apply to these
positions. Several membership premises can instantiate several universal
positions in one step; each citation must supply a used membership.

The sole new wording equivalence is the double-object / `to` alternation:
`Alice gave Bob a book` and `Alice gave a book to Bob` share the same three-place
atom. This applies to the verbatim forms give/gives/gave, send/sends/sent,
sell/sells/sold, offer/offers/offered, lend/lends/lent, hand/hands/handed, and
show/shows/showed. It does not equate different tenses or synonyms. Articles
`a` and `an` in these flat argument positions introduce existential binders.

Every ordering of the quantifiers is checked, up to **four quantifiers per
sentence** (at most 24 readings). A partially explicit `is such that` scaffold
fixes its outer binders; remaining quantifiers still receive every ordering.
Nested scaffolds can select a complete reading, for example:

```text
Some package is such that every courier is such that every recipient is such that that courier sends that package to that recipient.
```

More than four quantifiers yields a nearest-form error. There is no fixed
argument-position limit when at most four positions are quantified. This is a
restricted syntactic catalog, not general English parsing: new flat forms use
a single verb token, named or quantified noun phrases, and the preposition
vocabulary `to`, `from`, `with`, `for`, `at`, `in`, `on`, `into`, `onto`, `under`,
`over`, `through`, `between`, `beside`, and `by`. Prepositions inside a noun
phrase, relative clauses, repeated bound-class references in a scaffold, and
embedded negations beneath multiple scaffolds are not supported by the new
grammar. Prefer the prepositional form when double-object noun boundaries are
unclear. Existing two-place and copular forms retain their previous grammar.

The finite-model oracle independently translates the exercises to first-order
formulas and exhaustively searches domains of sizes 1–3, allowing empty classes
and co-denoting names. Boolean unit propagation prunes the search without
sampling models. This bounded check is evidence of soundness, not a proof over
unbounded domains.
