# Entailment Gate — Coverage and Known Limits (catalog v24)

**Type:** reference
**Status:** current as of catalog v24 (2026-11-10)


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
phrase, repeated bound-class references in a scaffold, and
embedded negations beneath multiple scaffolds are not supported by the new
grammar. Prefer the prepositional form when double-object noun boundaries are
unclear. Existing two-place and copular forms retain their previous grammar.

The finite-model oracle independently translates the exercises to first-order
formulas and exhaustively searches domains of sizes 1–3, allowing empty classes
and co-denoting names. Boolean unit propagation prunes the search without
sampling models. This bounded check is evidence of soundness, not a proof over
unbounded domains.

## v23: relative clauses

`F that A Rs` decomposes as F(x) and R(A,x); `F that Rs A` decomposes as
F(x) and R(x,A). `who` and `which` use the same rules. Unary clauses such as
`runner who runs` and n-place clauses such as `book that Alex gave Blair`
are supported. The filled clause must match verbatim; this path does not apply
the v22 double-object/to equivalence inside a relative clause.

Instantiation needs every conjunct of the restrictor. Existential introduction
needs both the restrictor and the body's facts. Predicate nominals can be
introduced from their two conjuncts or eliminated to either conjunct. Every
cited premise must contribute. Fitch existential witness assumptions state the
full restrictor and body; elimination exposes both class and relation, and the
existing fresh-name checks prevent witness escape. Universal generalization
also accepts relative restrictors.

Relative nouns may occupy subject or object positions under every/some/no.
The new path permits one quantifier per unscaffolded clause. To nest, use:

```text
Every hero that Sam admires is such that Pat admires every reader who visits that hero.
```

The scaffold fixes the outer scope, and instantiation preserves its entire
body. Unscaffolded relative clauses inside relative clauses receive a nearest
form diagnostic. Historical forms still run first, with unchanged parsing.

The new relative-clause verb vocabulary is deliberately bounded:
`lobbies`, `funds`, `admires`, `helps`, `visits`, `owns`, `reads`, `runs`,
`sleeps`, `smiles`, `works`, `gave`, `gives`, `sends`, `sent`, `sells`, `sold`,
`offers`, `offered`, `lends`, `lent`, `hands`, `handed`, `shows`, `showed`.
Other verbs, embedded negation, quantifiers inside the relative guard, and
unscaffolded nested relatives require separate explicit steps or a visible
bridging premise. There is no paraphrase, tense, or active/passive conversion.

The independent finite oracle translates relative nouns into conjunctions and
checks every new accepted exercise over domains 1–3, with empty classes and
co-denoting constants allowed. Invalid semantic look-alikes have countermodels;
Fitch checks cover local obligations and discharge. This is bounded evidence,
not a proof for arbitrary domain sizes. A redundant-citation rejection can
still be logically valid: class membership and a matching relation directly
entail their relative nominal, regardless of an unused universal premise.

## v24: explicit variables and structured FOL

The escape hatch binds a single lowercase letter with
`every x is such that BODY`, `some x is such that BODY`, or
`no x is such that BODY`. These mean unrestricted ∀x, ∃x, and ∀x¬ respectively.
Scope extends to the end of BODY; parentheses delimit a compound body or a
quantifier embedded in a connective. For example:

```text
1. every x is such that every y is such that every z is such that if x outranks y and y outranks z then x outranks z.
2. Therefore, every y is such that every z is such that if alice outranks y and y outranks z then alice outranks z. [1]
```

The body accepts identity, unary `NAME is PREDICATE`, verbatim binary verbs,
and three/four-place relations with `to`, `from`, or `with` separators.
Names and predicate words are single tokens; use names of at least two letters
in fallback proofs. `x gave y a book` is also supported: the article contributes
an existential book restrictor, inside the explicit binders. Arbitrary English
noun phrases, modal contexts, and relative clauses are not parsed by this core.
Use the existing readable grammar for those forms.

Use `(A) and (B)`, `either (A) or (B)`, `if (A) then (B)`,
`(A) if and only if (B)`, and `it is not the case that (A)` to fix boundaries.
Quantifiers can occur anywhere in these formulas. An unbound single-letter
term or rebinding an enclosing variable is an error in the fallback, including
when instantiating a quantified premise. Legacy single-letter names and
schematic placeholders retain their historical meaning outside fallback proofs;
they are not newly interpreted as free logical variables.

Normal phrasing remains the default. Legacy sentences keep their old parser
and schema priority. Explicit single-letter binders opt into the structured
parser, so legacy opaque matching cannot accept a malformed binder or split
its body into an unbound conjunct. Other structured forms are considered only
after the historical gate rejects the step.

The readable grammar retains its four-quantifier cap. The fallback does not
use that cap or enumerate scope readings: it is uncapped up to the tested eight
nested quantifiers, with no numeric quantifier-depth check. The finite alphabet
and prohibition on rebinding nevertheless limit simultaneous binders to 26.
Disjoint scopes may reuse letters.

For unrestricted Fitch generalization use `Let witness be arbitrary.`.
The declaration introduces a fresh object name without a class assumption;
its usable fact is reflexive identity. Derive the body for that name and
cite the whole range to conclude `every x is such that BODY`.
`Let witness be an arbitrary F.` retains the older, class-restricted meaning;
it cannot establish an unrestricted universal. Existential elimination assumes
the whole instantiated body, and checks that the witness occurs neither in
any earlier line nor in the discharged conclusion. The freshness check is
conservative, including earlier closed subproofs; use a new name each time.

All propositional introduction/elimination rules, quantifier rules, and identity
rules operate on the structured AST. Equality substitution can replace a chosen
subset of name occurrences, including beneath quantifiers, while leaving bound
variables unchanged. It does not enter opaque or modal contexts. A cited
`contradiction` supports contradiction elimination. This does not change the
historical rejection of a direct two-premise explosion jump.

## Completeness (FOL with identity)

The intended argument has four parts:

1. Translate controlled sentences to closed FOL= formulas and render formulas
   back into controlled sentences, preserving alpha-equivalence. v24 shares the
   oracle's AST and tests 2,000 seeded random formulas of depth at most five with
   a budget of six quantifiers, including arities 1–4 and identity. Separate
   cases pin six and eight nested binders. Every generated formula round-trips.
2. Classical natural deduction with identity is complete for FOL= over nonempty
   domains. Thus each valid sequent has a finite natural-deduction derivation.
3. Translate each rule of that derivation into an accepted one-step rule or
   Fitch discharge. The inventory tests 50 instances each of ∧I, ∧E, ∨I, ∨E,
   →I, →E, ¬I, ¬E, ⊥E, double-negation elimination, ↔I, ↔E, ∀I, ∀E, ∃I, ∃E,
   =I and =E: 900 instances. Nested arbitrary formulas are used, not opaque
   propositional labels. The unrestricted declaration and selective equality
   substitution close two gaps found during implementation. Reflexivity uses
   a citation; a premise-free theorem can start an arbitrary-name subproof,
   derive reflexivity, generalize, and instantiate the desired name.
4. If every intermediate formula in the natural-deduction derivation is
   renderable, the translation supplies a derivation accepted by the gate,
   potentially using fallback wording for intermediate steps.

**This is not yet a proof of the unrestricted completeness claim.** Step 1 is
an expressiveness statement for the tested core, not for all FOL= formulas:
26 single-letter variables cannot express formulas needing more simultaneously
bound distinct variables without rebinding. Step 4 therefore still needs an
unbounded variable notation, or a separate theorem establishing that every
relevant derivation can stay within the available variable fragment. No such
fragment theorem is claimed here. The renderer also targets a normalized,
function-free signature of unary predicates and relations of arity 2–4, with
single-token names; it is not a renderer for arbitrary English grammar.

Nor is the entire historical grammar one-to-one with closed formulas up to
alpha-equivalence: some readable sentences admit multiple scope readings,
and equivalent surface forms can denote the same atom. The accurate property
for the structured core is `parse(Render(f)) ≡α f`, not a bijection for every
legacy sentence. Connecting every legacy grammatical form to this core by
accepted conversion steps remains unproved. The finite inventory supports the
rule correspondence, but cannot remove these expressiveness and translation
obligations. Do not describe the full catalog as complete for FOL= yet.

The finite oracle exhausts domains 1–3 and all interpretations, including
empty predicate extensions and co-denoting names. It checks all fallback
exercises, their invalid look-alikes, the full transitivity example, and a
fixed 10/50 sample of each inventory rule, including local Fitch obligations
with eigenname closure. It is bounded soundness evidence, not a proof for
unbounded domains. No LLM or mechanised proof is involved.

Remaining limits are verbatim matching, fallback intermediate lines, the
readable four-quantifier cap, the variable alphabet restriction above, and
ordinary input resource limits. Counting quantifiers and definite descriptions
are planned as defined abbreviations; neither is implemented by v24.
