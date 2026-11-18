# Entailment Gate — Coverage and Known Limits (catalog v41.0)

**Type:** reference
**Status:** current as of catalog v41.0

Coverage claim: complete for propositional logic and for function-free
first-order logic with identity, within a controlled English (sentences with one
reading); a bounded finite-model oracle tests that the checker never accepts an
invalid argument; this is not proved.


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
| Counting and arithmetic | `Three senators voted yes and two voted no, so at least five voted.` | State the arithmetic as a premise |
| Tense and ordering | `It rained before the meeting.` + `The meeting was before lunch.` ⊢ `It rained before lunch.` | State transitivity of "before" as a premise |
| Verb predicates as properties | `Superman flies.` ⊢ `Superman has the property of being ...` | Write `Superman is someone who flies.` (copular) or add a premise |
| Counterfactual subproofs, `might` rules, bare `would V` | `If the match were struck, then the match would light.` has no plain form; `might` gets no rules | Bridging premises; `had gone` vs `went` also needs one |
| Modal logic beyond T | `Necessarily P.` ⊢ `Necessarily, necessarily P.` (S4) | Premise; which system holds is itself a substantive claim |
| Proving rules over predicates | concluding `If Plato is X, then Socrates is X.` (it can be used, not derived) | Use `has every property that` with property generalization (v20–v21), or assume the schema |
| Other property wordings | `Being just is a property that Socrates has.` | Use the canonical `Socrates has the property of being just.` |

## Out of scope for the gate

- **Premise truth:** the gate never judges whether premises are true.
- **Fidelity:** the gate never checks that an argument matches its source text.
- **Bridge squish:** an ad hoc `If [premises], then [conclusion]` always passes.
  Passing shows validity of the encoding, not that the reasoning was
  reconstructed.

## Writing so the gate can see the inference

Clauses match word for word after limited normalization (case, whitespace,
contractions, optional `then` after `if`, and simple elided copular
disjunctions). Synonym and paraphrase tolerance are outside the gate: if two
wordings should count as equivalent, state that connection as a visible identity
or conditional premise.

- **Canonical connectives:** write sentential negation as `It is not the case
  that X.`, disjunction as `Either X or Y.`, conjunction as `X and Y.`, and
  parenthesize nested biconditional sides as `(X) if and only if (Y)` (optionally
  prefixed by `It is the case that`). `It holds that` is ordinary text.
- **Quantifiers and agreement:** for singular instantiation use `Every man is
  mortal.` with `The traveler is a man.` (`All men are mortal.` does not match
  singular `a man`; use `All A are B.` + `All B are C.` for class syllogisms).
  Schematic placeholders (`U`, `V`, `W`, `X`, `Y`, `Z`) must agree in number with
  their bindings (`If X are fixed, ...` for plural subjects), and possessives
  (`X's`) and nested conditionals work.
- **Copular parsing and examples:** copular sentences split at their first ` is `,
  so give a short name to any subject containing an embedded `is` and state the
  identity as a premise. Keep parenthetical `such as ...` examples outside
  premises so they do not block placeholder binding.
- **General rules over ad hoc bridges:** to conclude one alternative rather than
  another, state the dichotomy (`Either we should use X or we should use Y.`),
  refute one side, and apply disjunctive syllogism. Replace ad hoc bridge
  conditionals with the general schematic or universal premise the argument
  relies on. See [`rules.md`](rules.md) for all 119 inference forms and worked
  examples.

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
sampling models. This bounded check tests that the checker never accepts an
invalid argument; this is not proved over unbounded domains.

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
The original core uses single-token names and predicates; v25 also provides
lossless name and atom escapes (below). `x gave y a book` is also supported: the article contributes
an existential book restrictor, inside the explicit binders. Arbitrary English
noun phrases, modal contexts, and relative clauses are not parsed by this core.
Use the existing readable grammar for those forms.

Use `(A) and (B)`, `either (A) or (B)`, `if (A) then (B)`,
`(A) if and only if (B)`, and `it is not the case that (A)` to fix boundaries.
Quantifiers can occur anywhere in these formulas. An unbound single-letter or indexed
term or rebinding an enclosing variable is an error in the fallback, including
when instantiating a quantified premise. Legacy single-letter names and
schematic placeholders retain their historical meaning outside fallback proofs;
they are not newly interpreted as free logical variables.

Normal phrasing remains the default. Legacy sentences keep their old parser
and schema priority. Explicit single-letter or indexed binders opt into the structured
parser, so legacy opaque matching cannot accept a malformed binder or split
its body into an unbound conjunct. Other structured forms are considered only
after the historical gate rejects the step.

The readable grammar retains its four-quantifier cap. The fallback does not
use that cap or enumerate scope readings. In v25 a variable is a letter
optionally followed by digits (`x`, `x1`, `y12`). The renderer uses all 26 bare
letters, then `x1`, `x2`, and so on. There is no numeric binder limit.
Disjoint scopes may reuse variables; rebinding an enclosing variable and free
variable occurrences remain errors. The implicit book witness also draws from
this unbounded alphabet. Use `name{x}` for a constant whose spelling looks like a variable.

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

## v25: notation equivalence and legacy audit

`notation equivalence` cites exactly one premise, with no subproof ranges. Both
sentences must have exactly one reading up to alpha-equivalence, and those
readings must be alpha-equivalent. It works in either direction. It does not
perform semantic-equivalence search, synonym matching, or paraphrase. Existing
forms retain priority and therefore may still report self-entailment or FOL
reiteration for a conversion they already accepted. The new form is append-only.
An ambiguous conversion reports: **use is such that to fix the reading**.
Even two universally quantified readings that commute logically are distinct
for this test: logical equivalence is weaker than alpha-equivalence.
The historical every-reading checks are unchanged.

The audit inventory in
[`legacy_v24_sentences.json`](../validation/testdata/legacy_v24_sentences.json)
contains every distinct readable sentence observed in accepted arguments during
the unchanged v24 validation suite at baseline `7636479`: fixtures, golden
cases, catalog cases, and generated oracle cases. Capture occurs at the gate
boundary, after expansion, so dynamically constructed arguments are included.
Arbitrary-name declarations and explicit-variable sentences are excluded.
The legacy capture was produced from a v24 checkout. Each row
records its sentence, source argument forms, and any named exception. Source
forms are provenance, not independent semantic coverage of every rule.

Current results are recorded in the table below. The conversion test directly
requires notation equivalence and separately passes each direction through the
public gate. Each successful rendering is checked against both the explicit
parser and the historical reading enumerator; neither may change its meaning.
New readable atoms cannot bypass the existing ambiguity checks through the FOL
fallback: a newly parseable unscaffolded formula must have one matching legacy
reading before it can be used by that fallback. Newly supported compound
fallback inputs must also use the renderer's parentheses; readable prose keeps
its historical one-step schemas and precedence guards.
The inventory records 106 source forms overall. Source-form provenance from a
modal/property argument does not establish a FOL translation of those operators.
Exact sentence-level exceptions remain in the JSON inventory.

**6,386 sentences; 3,550 bidirectional conversions; zero conversion failures;
2,424 ambiguous; 412 non-FOL; zero remaining lexical/opaque exceptions.**
All 3,434 prior conversions remain. The final 116 exceptions are resolved,
without reclassifying any sentence. Successful conversions cover 96 source forms,
with 177 bounded oracle samples (domains 1–3), plus hand-specified lexical cases
and propositional tautology/non-tautology checks.

| Category | Before | After |
| --- | ---: | ---: |
| Converted both ways | 3,434 | 3,550 |
| Multiple non-alpha-equivalent readings | 2,424 | 2,424 |
| Non-FOL modal, counterfactual, or property grammar | 412 | 412 |
| Variable-shaped constants | 89 | 0 |
| Lexical atoms without a unique explicit parse | 13 | 0 |
| Changed lexical segmentation or operator structure | 9 | 0 |
| Historical parenthesized-parser conflicts | 4 | 0 |
| Opaque scope/connective syntax conflict | 1 | 0 |

### Lossless lexical notation (v25 additions)

`name{x}` and `name{y2}` denote constants with those spellings, even inside a
binder using the same spelling: `every x is such that x admires name{x}`.
Bare legacy names retain their prior behavior outside explicit bodies.
For other lexical boundaries, `name{"claim that the store is nearby"}` quotes
a complete name. Quoted names use JSON string escaping; decoded names are
case-normalized. Names beginning with internal-marker characters remain constants,
never variables; `name{"$slot0"}` does not denote a bound slot.

`atom{"works/with", "alice", x}` fixes a verbatim predicate symbol and its
ordered terms. Quoted arguments are constants; unquoted arguments must be bound
variables. The number of arguments fixes the arity. A unary predicate uses the
existing `is/` symbol, as in `atom{"is/red", "alice"}`. With no arguments,
`atom{"either it rains"}` denotes exactly that normalized zero-place proposition.
JSON quotes protect commas, parentheses, connective words and scope text from
structural parsing. The renderer JSON-escapes brackets and repeated or non-ASCII
whitespace so numbered-argument preprocessing cannot treat them as citations or
collapse a literal label. Use `\u005b` / `\u005d` for literal square brackets. Curly braces keep this notation distinct from bracketed
step citations. No natural-language meaning is assigned to a newly quoted
predicate: the symbol and arity must match exactly.

The renderer first tries ordinary phrasing and checks that its parsed AST is
identical, including the historical parenthesized parser. When lexical boundaries
would change, it uses the atom escape. This resolves `works with` boundaries,
that-clause copulas, restricted-existential lexical leaves, opaque `either ...`
leaves, and the nested-relative scaffold in the audit without changing their
recorded readings. A scaffold represented by the legacy enumerator as an opaque
atom stays opaque; the escape does not invent internal quantifier semantics.
The audit still requires the public gate to accept both directions. Conversions
of that-clause or opaque labels do not establish their intended English semantics.

The escape syntax selects the explicit parser; existing prose recognizers and
inference forms are unchanged (introduced in v25 without adding an inference
form; the current catalog is **v41.0**). Tests cover capture,
arity/order distinctions, malformed/free variables, literal structural text,
reserved names, and JSON-escaped eigenname freshness.

Exceptions are pinned per sentence. The test fails if any classification changes;
`ARGCHECK_UPDATE_LEGACY=1` is an explicit ledger-update operation, not normal
verification. No unsupported alternative is discarded to obtain a unique
renderable reading. The bridge enumerates candidates conservatively; two
non-alpha-equivalent candidates suffice to refuse it.

G1 evidence: the original 2,000 v24 round trips remain unchanged. A new seeded
1,000-formula generator uses large variable sets for one fifth of its formulas:
200 formulas bind 30–40 distinct variables, all occurring below the full prefix.
Every formula round-trips; large formulas also pass gate reiteration and, when
universally headed, instantiation. Separate cases test depths 26, 27, 30, and 64,
indexed binding errors, disjoint scopes, and implicit witnesses after 26 binders.

## Completeness (FOL with identity)

**Scoped claim.** For every finite argument whose premises and conclusion
are sentences of the controlled grammar, each with exactly one closed,
function-free FOL= reading up to alpha-equivalence, if that argument is valid
over nonempty domains, the gate accepts some derivation of it. The derivation
may use catalog one-step forms, Fitch subproofs, notation equivalence, and
explicit-variable intermediate lines. This is existence of a finite derivation,
not a claim that any valid argument is accepted in one step or automatically
found by the checker. Modal, counterfactual, and property quantification are not
FOL= readings and are outside this claim, as is that-clause grammar.

**Verdict: Established over the stated scope.** This is a constructive
implementation-correspondence argument, not a machine-verified metatheorem and
not a claim about unrestricted English. “Reading” here means the gate's own
normalized controlled-grammar reading. The evidence comprises the round trips,
legacy bridge audit, rule inventory, bounded oracle, and this proof sketch:

1. **Notation bridge.** For a sentence with one closed FOL= reading, the legacy
   enumerator supplies its AST without discarding competing candidates.
   `notation equivalence` compares this AST with the rendering by
   alpha-equivalence. The literal atom notation specifies exactly a symbol,
   arity and ordered terms; name escapes distinguish constants from variables.
   Thus lexical segmentation and structural-looking opaque labels no longer
   prevent rendering. The audit exercises all 3,550 single-reading FOL rows in
   both directions, with no residual exceptions or reclassifications.
2. **Underlying calculus.** The completeness argument relies on the standard
   completeness of classical natural deduction for function-free FOL with
   identity over nonempty domains. Every valid finite sequent in that language
   therefore has a finite natural-deduction derivation.
3. **Rule translation.** Translate that derivation to the structured AST rules:
   conjunction/disjunction/conditional/biconditional introduction and elimination,
   negation and classical reductio, contradiction elimination, quantifier
   introduction/elimination, reflexivity and selective identity substitution.
   These rules inspect AST constructors, not predicate spellings or fixed
   quantifier depths. Fitch ranges express discharge. Fresh eigenname renaming
   meets the conservative requirement of freshness even in earlier closed
   subproofs; the finite derivation leaves infinitely many fresh names available.
   The unchanged inventory checks 50 instances of each of 18 rules (900 total).
   Decoded quoted names participate in freshness checks, so escaping cannot hide
   an eigenname in an earlier premise or the discharged conclusion.
4. **Intermediate formulas.** By structural induction, render every atom with
   its exact symbol and terms, equality with escaped names where needed, each
   connective with parenthesized children, and each quantifier with a fresh
   indexed variable. The variable alphabet is unbounded. Each finite formula
   has only finitely many forbidden lexical tokens; bound variables can be
   alpha-renamed to avoid them. Parsing reconstructs the same AST up to alpha
   renaming. Ordinary phrasing is only an optimization when the equality check
   succeeds; literal notation supplies the general atomic construction.
   Evidence includes the unchanged 2,000 v24 round trips, 1,000 v25 round trips
   (200 with 30–40 simultaneous binders), and literal-boundary tests.
5. **Composition.** Convert the premises to explicit notation, translate the
   finite derivation rule by rule, and convert its conclusion back. This gives
   an accepted finite derivation, not an automatic proof-search algorithm.

The finite tests support this construction and catch implementation mistakes;
they test that the checker never accepts an invalid argument (tested, not
proved), rather than proving completeness or validity over unbounded domains. The
argument depends on the stated parser/AST correspondence and standard natural-
deduction completeness, not on extrapolating a bounded oracle result.

Outside the claim:

- Ambiguous readable sentences retain every-reading checking. `is such that`
  fixes quantifier scope; the lossless lexical notation then crosses the bridge without paraphrase.
- Function terms are absent. FOL= with names and no function symbols still has
  a complete natural-deduction calculus for that signature; omitting functions
  is a signature restriction, not an incompleteness of that calculus.
- Counting bounds 1–5 are identity-based FOL abbreviations (v26). Larger bounds
  and general arithmetic remain outside the supported surface grammar.
- Unambiguous subject descriptions are identity-based FOL abbreviations (v27).
  Expansion and contraction use the exact Russellian AST with a negated
  two-distinct-witness uniqueness bound. This adds a notation bridge, not a
  stronger logic. Explicit narrow negation is unambiguous; surface "the F is
  not G" remains outside this single-reading completeness claim. Historical
  certificates retain their previous semantics rather than a retroactive
  Russellian every-reading check (see the v27 compatibility exception below).
- Modal, counterfactual, property and that-clause grammar are outside the
  function-free FOL claim. An opaque propositional letter carries no internal
  modal or attitude semantics.
- Paraphrase remains a permanent non-goal. Zero-place atoms render verbatim;
  different normalized clauses remain independent propositional letters.

A bounded finite-model oracle tests that the checker never accepts an invalid
argument; this is not proved: the oracle checks domains 1–3, empty predicate
extensions, and co-denoting names. v24 checks a fixed 10/50 sample of every
inventory rule, plus exercises and local Fitch obligations with eigenname
closure. v25 adds bidirectional conversion samples and independently hand-built
meanings for representative readable sentences, alongside ambiguity and
near-miss rejections. The identical-AST criterion preserves truth conditional on
faithful parsing; the tests do not prove faithfulness for arbitrary English or
that the checker never accepts an invalid argument over all domain sizes.

Bare opaque text containing an active bound-variable token remains conservatively
rejected inside that scope. This does **not** block the claim: use a fresh binder.
For example, `every x is such that x would admire alice` is rejected, whereas
`every y is such that x would admire alice` has an opaque zero-place body and
round-trips. Alpha-renaming preserves the formula; it does not turn the literal
`x` in that opaque label into an individual argument. The renderer avoids such
opaque-token collisions when selecting binders. Explicit atom quotes also mark
literal text unambiguously. Tests cover the rejection, fresh renaming, rendering
and public-gate instantiation.

Other practical limits are the readable four-quantifier cap, explicit-variable
intermediate lines, and ordinary finite input/stack/memory resources.

## v26: counting quantifiers

`At least n`, `at most n`, and `exactly n` accept digits 1–5 and words one–five.
The private parser expands these into n existential witnesses with pairwise
inequalities, the negation of n+1 such witnesses, or their conjunction.
Introduction/elimination converts both ways with explicit FOL definitions.
Other rules cover exact-count bounds, inconsistent bounds, uniqueness, and a
larger class versus its smaller subset. These are structural rules over identity
definitions; production never uses bounded model search for acceptance.
Fitch discharge reuses the existing FOL checker on expanded definitions.

The readable surface accepts subject counts, copular predicates, existing
relation vocabulary, and single relative clauses. `signed` and `wrote` are
recognized privately; historical lexical candidate sets are unchanged. Simple
plural classes use a final-s spelling. Other morphology, nested count phrases,
ambiguous scope, and general arithmetic require explicit formalization.

Uniqueness requires membership in the counted class. Inferring `No person other
than Scott wrote Waverley` from `Exactly one person wrote Waverley` and `Scott
wrote Waverley` alone is invalid because it lacks
`Scott is a person`: a nonperson Scott and a distinct person may both write
Waverley. The two-object countermodel forces rejection. Adding that membership
as a separate cited premise (or in the cited ground conjunction) yields the
intended VALID uniqueness step.

The v26 tests check accepted local and discharged proof obligations over domains
1–4, including empty extensions and co-denoting names, and compare all 15 count
abbreviations against independently assembled identity formulas. Counts 4 and
5 have vacuous lower-bound checks on smaller domains; these bounded checks test
that the checker never accepts an invalid argument (tested, not proved). Invalid
exercises each require a countermodel.

Compatibility: existing successful proofs keep their historical schema and
report; counting is a fallback for new inferences. Explicit out-of-range bounds
receive the counting-bound diagnostic only after historical paths decline the
step. Non-numeric adverbial uses do not activate the counting quantifiers. The
290-argument corpus comparison checks both verdicts and normalized report hashes.

## v27: Russellian definite descriptions

`The F is G` expands to existence with predication plus at most one F:
`∃x(Fx ∧ Gx) ∧ ¬∃x∃y(Fx ∧ Fy ∧ x ≠ y)`. This is equivalent over nonempty
domains to Russell's `∃x(Fx ∧ ∀y(Fy → y = x) ∧ Gx)`. The oracle independently
uses the latter universal-identity analysis. Production reuses v26's counting
definition and checks syntactic rule instances; it performs no model search.

Supported forms are description definition, substitution, introduction,
elimination, uniqueness, and negation. Introduction
accepts existence-with-predication plus uniqueness, or a named member and its
predication plus an at-most-one/exactly-one bound. Elimination returns the
existential conjunct, uniqueness conjunct, existence alone, an exact count,
or membership of an identified witness. Definition expansion/contraction
connects descriptions with the explicit FOL calculus. Fitch discharge uses
the existing scope and eigenname checker on translated definitions.

Subject descriptions support one relative clause, including n-place relations,
and literal of-phrases: `author of Waverley` denotes the ordered author-of
relation to Waverley. No synonyms or differently written relative restrictors
are equated. Bare capitalized copular complements denote names; lowercase
complements denote predicates. Explicit `is identical to` removes this surface
convention. Nested descriptions, descriptions in arbitrary argument positions,
and unrestricted English remain outside this extension; write explicit FOL
intermediate lines. At most 16 readings per compound and 12 distinct clauses
per fallback step are considered.

The fallback requires a real witness term in an atomic predication (or its
negation); an opaque zero-place clause is not a description predicate. Binary
conjunctions and disjunctions in this fallback require parentheses around both
operands. Readable `either`, `both that`, and biconditional scaffolds remain
with the legacy gate, so description expansion cannot change their scope.

**Present King of France.** `The present king of France is not bald` admits
both existence/uniqueness with negative predication (wide) and negation of
the entire positive description (narrow). The description fallback requires
every combination of admissible premise and conclusion readings to validate;
repeated occurrences of the same clause share a reading. Explicit `It is not
the case that the present king of France is bald` fixes only the narrow
reading. Neither reading is silently discarded.

**Compatibility exception.** The v26 gate always runs first. Its successes
and full reports are preserved, and failed new attempts return the original
v26 diagnostic. Consequently, the old certificate from explicit narrow
negation to surface "the F is not G" still passes: v26 treats these as ordinary
clausal negation, whereas Russellian wide scope has a countermodel with no F.
A regression test pins both this unchanged report and its independent
Russellian countermodel. Thus legacy preservation and a retroactive universal
Russellian validity check cannot both hold; v27 preserves legacy behavior
as required. No description-specific error overrides an old VALID result.

Opaque that-clauses, attitudes, modal and deontic contexts expose no substitutable
terms. Invalid look-alikes cover missing existence, missing uniqueness, opaque
substitution, scope, unnamed witnesses, and changed restrictors.

Validation includes 20 valid and nine invalid fresh exercises, independent
finite-model checks over domains 1–3, 39 surface-reading equivalence checks
over domains 1–4, and Fitch local/discharge obligations. Empty extensions and
co-denoting constants are included; bounded checks test that the checker never
accepts an invalid argument (tested, not proved). The public corpus retains 24 sentences of fewer than 30 words.
Each is checked under modus ponens and
an unrelated-premise rejection, comparing complete v26/v27 step reports.
The source index records neutral source IDs and line numbers.

## v28: schematic substitution and premise placement

The form list is unchanged: these are extensions of identity substitution and
schematic modus ponens. Historical successful certificates retain priority.

Bare-and schematic antecedents now accept up to **ten conjuncts**, with one
used cited premise per conjunct. The matcher binds in antecedent order and
backtracks immediately on incompatible premises instead of allocating every
premise permutation. Repeated category conjuncts and `from Y` retain their
verbatim meanings. Unsplit `and` tokens can remain inside a conjunct if its
whole text matches a cited premise. Plain partition enumeration retains its
12-boundary bound. Existing explicit `, and it is the case that` scaffolds
retain their older supported lengths. A failed branch names its first unmatched
conjunct and the binding in force; diagnostics select the branch with most
matched conjuncts, deterministically. Exceeding the bare-and conjunct limit
reports `antecedent has N conjuncts; the limit is 10`.

`U`, `V`, and `W` join `X`, `Y`, and `Z` for schematic conditional instantiation.
The original alphabet is tried first, preserving old certificates that use the
new letters literally. Otherwise the extended alphabet requires the same
binding everywhere and forbids an unbound consequent variable. This extension
does not change explicit lowercase FOL binders or existential-placeholder syntax.

The identity fallback accepts `is identical with` as well as `is identical to`.
It recognizes whole gerund terms and coordinated proper names stipulated by an
identity, in a copular subject or a terminal object of the controlled transitive
register (`deserves`, `admires`, `praised`, `praises`, `helps`, `helped`,
`receives help from`). For example, `Standing as an equal` can replace
`being recognized as a peer` as the object of `Every citizen deserves ...`.
Coordinated proper names require capitalized words in each coordinand.

The fallback replaces a complete term, never a predicate or part of a longer
word/name. It retains quotation, attitude, sentential modal and bound-definite
safeguards. Bare adjectival `necessary`/`possible` conjuncts (optionally modified
by `not`, `always`, `often`, or `sometimes`) are read as predicates of the
subject, allowing `Ada is necessary and often obligatory`; this does not expose
names inside `It is necessary that ...`. The new fallback requires one complete
term occurrence on each side. Historical all-occurrence and structured FOL
identity rules retain their existing behavior, including selective substitution
where the structured rule explicitly permits it.

Root premises may follow inferences or completed Fitch subproofs. Citations
must still point backward and obey subproof visibility. Ordinary premises cannot
be inserted at a nonzero proof depth. Fresh-name checking also scans later open
premises before discharge: a late existential containing the witness name must
not license a circular existential elimination.

**Legacy narrow negation.** The historical clausal reading of
`the F is not G` remains deliberate. In particular, explicit
`not (the F is G)` still licenses that surface wording under the old certificate.
The Russellian description-wide reading instead asserts
`∃x(Fx ∧ ∀y(Fy → y=x) ∧ ¬Gx)`. In a nonempty model with **no F**, the historical
clausal negation is true and that Russellian reading is false. The inherited
negation regression and its independent finite countermodel remain pinned;
v28 does not retroactively impose the stronger reading. Concluding existence
of an F from the narrow negation remains rejected.

Verification compares the new regression fixtures against untouched v27 commit
`f9f0042`, plus targeted historical tests. The negative-star and
shared-category/`from Y` examples already pass v27; they are preservation pins.
New identity exercises have manually assigned meanings checked by the existing
finite-model oracle, domains 1–3. The six-to-ten-conjunct schemas also receive
exhaustive propositional truth-table checks, and repeated-category substitution
and late-premise eigenname rejection receive independent finite-model checks.
These bounded checks test that the checker never accepts an invalid argument;
this is not proved.

## v28.1: slot bindings may not carry scope operators

A schematic slot (X, Y, Z, W, U, V) stands for a term. Through v28 a slot could
bind any substring, so it could absorb an operator from the premise and leave the
rule's predicate inside that operator's scope. v28 accepted, for example:

```text
1. It is not the case that Socrates is wise.
2. If X is wise, then X is happy.
3. Therefore, it is not the case that Socrates is happy. [1,2]
```

with X bound to "it is not the case that Socrates": denying the antecedent. The
same capture worked with "no philosopher", "nobody", "not every philosopher",
"at most one philosopher", "it is false that Socrates" and "Plato denies that
Socrates". v28.1 rejects a binding that

- opens with negation, a non-upward quantifier or a falsity wrapper ("not every",
  "no", "nobody", "none", "nothing", "never", "neither", "few", "at most",
  "less than", "it is not", "it is false", "it is unlikely", "it is possible", …), or
- ends inside the complement of an attitude verb or modal adjective before the
  embedded clause has its own verb ("Plato denies that Socrates", "it is likely
  that the store").

Negation inside a relative clause ("the philosopher who is not rich"), a complete
clause inside a noun phrase ("the claim that the store is nearby"), "every",
"some", names and conditional consequents bind as before. Re-gating the 1,482
arguments of the internal corpus changed no verdict. Tests:
`validation/entailment_gate_slot_binding_test.go`.

v30 closed the repeated-indefinite gap recorded here; see below.

## v28.2: identity terms must refer

The blind completeness review found three paths by which the identity rules took
quantifier phrases as names. v28.1 accepted, for example:

```text
1. Nobody is identical to Felix.
2. Nobody is black.
3. Therefore, Felix is black. [2,1]
```

- The v28 whole-term identity register compared capitalized words against a
  lowercase word list, so "Nobody" and "Some cat" passed as names.
- `identity transitivity` had no term check ("Tom is identical to some cat",
  "Some cat is identical to Felix", so "Tom is identical to Felix").
- The readable-to-FOL bridge stripped "a"/"an" from identity terms, so "Tom is
  identical to a cat" read as tom = cat and fed `FOL identity substitution`.

v28.2 rejects an identity side that opens with a determiner or quantifier
pronoun (every, all, each, any, some, no, none, nobody, nothing, someone,
everyone, anything, a, an, few, many, most, several, both, either, neither,
only, not) in all three paths. Names, definite descriptions ("the cat") and
gerund terms behave as before. Re-gating the 1,482-argument corpus, the 23
reverse spot-check derivations and 193 identity-bearing mutants changed no
verdict. Tests: `validation/entailment_gate_identity_terms_test.go`.

The intensional-transitive identity gap is closed for `is looking for`,
`looks for`, `seeks`, `wants`, `needs`, `is searching for`, `hopes for`,
`fears`, `believes`, `imagines`, and `worships`.
`TestEntailmentGateIntensionalVerbBlocked` records the rejection.

## v29: "unless"

v28.2 had no form for "unless", so the Formalizer could not certify a repair that
kept a statute's wording ("Any filing received after the statutory deadline is
untimely unless the court granted an extension"; review fixture domain-legal-01
failed with repaired verdict `non_atomic`). v29 adds two forms on the weak reading,
`P unless Q` = `if it is not the case that Q, then P` (equivalently `P or Q`):

- `unless elimination`: `P unless Q` (or `Unless Q, P`), `It is not the case that Q` ⊢ `P`;
- `unless conversion`: `P unless Q` ⊣⊢ `If it is not the case that Q, then P`.

The biconditional and exclusive readings are rejected: `P unless Q` with `Q` does not
give `not P`, and with `P` does not give `not Q`. An unless-clause containing a
personal pronoun is not matched, since the pronoun may point back into a quantified
main clause. "Unless" text is otherwise left verbatim, not normalized: a first
attempt that rewrote "P unless Q" during clause normalization changed 10 corpus
verdicts, because sub-clauses are normalized at different granularities. With
dedicated forms, the 1,482-argument corpus, 177 unless-bearing mutants, the
reverse spot-check and the earlier repros change no verdict. Tests:
`validation/entailment_gate_v29_test.go`.

## v29.1: verb agreement in non-copular instantiation

`non-copular universal instantiation` required the instance to repeat the plural
verb verbatim, so `All squares have four sides` + `Rex is a square` gave only
`rex have four sides`. v29.1 also accepts the third-person-singular verb when the
membership premise is singular (`Rex is a square` ⊢ `Rex has four sides`): `have` →
`has`, `do` → `does`, `go` → `goes`, and regular `-s`, `-es` and consonant-`y` →
`-ies` spellings. The verbatim plural stays accepted, so no earlier certificate
changes; tense changes (`flew`) and mismatched agreement stay rejected.

## v30: a repeated slot binds one fixed thing

A schematic rule is read as universally closed over its placeholders. When a
placeholder occurs more than once in the antecedent, each occurrence is matched
against its own premise (or its own position in one premise), so a quantifier
phrase bound there is re-scoped at each occurrence. v28.2 accepted

    1. Some philosopher is wise.
    2. Some philosopher is rich.
    3. If X is wise and X is rich, then X is happy.
    4. Therefore, some philosopher is happy. [1,2,3]

although the two premises may concern different philosophers. The same held for
`a philosopher`, `someone`, number phrases and disjunctions (`Socrates or Plato`),
for `U`/`V`/`W` placeholders, for two-variable rules (`X loves Y and Y is loyal`),
and for one premise matched twice (`Some philosopher admires some philosopher` +
`If X admires X, then X is vain`).

v30 rejects these bindings:

- a placeholder repeated in the antecedent may not bind an existential,
  indefinite or disjunctive term;
- a placeholder repeated in the consequent may not bind a universal or
  disjunctive term (`Every philosopher is wise` + `If X is wise, then X admires X`
  does not give `every philosopher admires every philosopher`).

Still accepted: names, definite descriptions (including one whose relative clause
contains `or`), `every`/`each`/`all` phrases repeated in the antecedent (every
philosopher is wise and every philosopher is rich, so every philosopher is happy),
a slot used once with `some` or `every`, and one premise that already says both
things of one witness (`Some philosopher is wise and rich` + `If X is wise and
rich, then X is happy`). Bare plurals (`Philosophers are wise`) take their generic
reading and still bind; this is a reading limit, not a proof of soundness for the
existential reading.

The other forms that combine two premises about one term were checked and already
reject the shape: predicate conjunction (`Some philosopher is wise` + `Some
philosopher is rich` does not give `some philosopher is wise and rich`), restricted
universals with a conjunctive relative clause, and the identity rules (v28.2).
`TestEntailmentGateSlotBindingScope` holds the repros and controls.

## v31: modal phrases are opaque claim parts

Catalog v31 retains 121 forms. In indicative conditionals, `might`, `could`,
`may`, and `can` belong to the verbatim atomic claim. For example, `If the
system is misaligned, then the system might seek power` and `The system is
misaligned` yield `The system might seek power`. They do not yield `The system
seeks power` or `The system could seek power`. The same matching policy applies
to antecedents, modus tollens using explicit clausal negation, conjunctive
modus ponens, and schematic forms. No new modal inference is licensed.

The boundary is conservative and syntactic, not a general English mood parser.
`conditionalKind` in `validation/entailment_gate_v19.go` now receives the
antecedent as well as the consequent. Bare `might` is material unless the
antecedent matches the existing `counterfactualMood` markers (`had`, `were`,
and their listed auxiliary variants; an antecedent `would` atom does not count). `might have` (including `might not have`)
and the existing `could have` remain non-material. Thus even an epistemic
`might have` is conservatively refused. Existing `would` classification and
counterfactual forms are unchanged; the change does not permit arbitrary
counterfactual reasoning. Existing `could not have` treatment is unchanged.
The recursive classification preserves explicit negation and Boolean scope;
`v19Visit` and opaque masking use the same boundary. `counterfactualUnsafe`
and mood rewriting are unchanged. Existing system-T forms on `possibly` and
`necessarily` still operate only on their explicit modal scaffolds.

`validation/entailment_gate_v31_test.go` tests eight representative unchanged corpus fixtures
in `validation/testdata/v31/`, ordinary and schematic forms, modal non-inferences,
counterfactual refusals, and system-T controls. It needs no `workspace/` files.

## v27 legacy negation

Decision A retains the historical clausal-negation verdict: from `It is not
the case that the present king of France is bald`, the gate accepts `The
present king of France is not bald`. The old certificate treats both as the
negation of the same clause, and is preserved before the description fallback.

This is not valid for the Russellian predicate-negation reading of the
conclusion. Let the domain contain one object and let there be no present king
of France. Then `not (there exists exactly one king and that king is bald)`
is true, but `there exists exactly one king and that king is not bald` is false.
The legacy verdict therefore does not establish existence of a king. The
existing `TestCatalogV27InheritedNegation` pins compatibility;
`TestV31V27LegacyNegationEmptyRestrictor` additionally forces the empty king
extension in a one-object countermodel. No negation verdict is changed in v31.

The full 36-fixture check passed 28: six still hit the existing repeated-indefinite
slot guard, one hits the counting-definition reading guard, and one contains
possessive `could have enough advantage`, which the conservative perfect-modal
boundary still refuses. v31 does not relax those boundaries to fit the corpus.

## v31.1: step references in a consequent

A bridge premise may now name earlier steps as its whole consequent (`If A, then [1].`), and a comma may stand before `and` between references (`If [1], and [2], then C.`). A consequent reference expands exactly as an antecedent one does, to `it is the case that <the step's text>`, so it can feed modus tollens and chains. A step reached only through a bridge reference now counts as used for the connectivity check. The bridge itself must still be cited.

This change only widens what is accepted. On the 1,482-argument internal corpus and the 281 freq-vs-bayes arguments, no verdict changed in either direction (against v30). Tests: `validation/step_references_test.go`, and the modus tollens example under "Step references" in `docs/rules.md`.

## v31.2: postposed conditionals

A numbered statement of the form `Y if X.` (or `Y, if X.`) may now serve as the conditional `If X, then Y.`, so it feeds modus ponens, modus tollens, hypothetical syllogism, conjunctive antecedents and step references (`C if [1].`). The rewrite is applied to whole statements only, and only on a second reading after the argument as written fails, so it can add verdicts but never remove one. It is skipped where the reading is ambiguous: `only if`, `even if`, `as if`, `if and only if`, any `unless`, embedded questions (`asked if`, `curious if`, `whether`), consequents that report or attribute (`says`, `thinks`, `that`), a consequent with `and`/`or`/`either` or a comma, an antecedent with `or`, and more than one `if`. The v9 whitelist of very simple clauses (`The lamp is lit if the switch is up.`) still applies inside clauses, as before.

On the 1,482-argument internal corpus and the 281 freq-vs-bayes arguments, no verdict changed in either direction (against v31.1). Tests: `validation/postposed_if_test.go`, and the modus ponens example in `docs/rules.md`.

## v31.3: conclusion markers

A derived step is any numbered line that ends in a terminal citation list.
The leading marker is optional and may be `Therefore`, `So`, `Thus` or `Hence`,
with or without a comma, case-insensitive. The marker is stripped before
matching, so the verdict does not depend on which marker is used. An unmarked
`So` followed by `many`, `much`, `few`, `little`, `far`, `long`, `that` or
`called` is read as content, not as a marker. A line that starts with a marker
but has no valid terminal citations is a syntax failure. Inline references
(`If [1], then P.`) keep the line a premise. There are no verdict changes on
the corpus (2,476 arguments).

## v32.0: explicit and/or grouping

And/or grouping is checked at numbered-argument parsing, before entailment.
Whole-clause parentheses, a single main-connective comma, and leading
either/or or both/and fix scope; ambiguous mixtures and unbalanced parentheses
fail with line/sentence diagnostics. Single-connective lists and square-bracket
step references retain their behavior. `validation/andor_grouping_test.go`
covers these forms and opaque-context exemptions. The exemptions conservatively
cover containing clauses and can miss outer ambiguity; see the grouping
section of `docs/rules.md` for this known limitation.
Existing parenthetical definition/condition/proposition labels remain atomic;
truncated text with an unmatched parenthesis now fails syntax intentionally.
A connective counts only when it joins whole clauses (see the grouping
section of `docs/rules.md`), so phrase-level coordination such as `a bad or
catastrophic outcome` is not flagged. Corpus (2,476 arguments): 2 arguments
newly fail syntax, both because a single-letter term inside a noun phrase reads
as a clause start: `corpus-ai-safety-carlsmith-a6d0e28c4a` (`a version of X or
still less of Y or Z or U` beside an `and` chain) and `wilkinson-2022 A112`
(`Z − 2e or Z − 3e`; already invalid before).

## v32.1: NOT CERTIFIED verdict

A failing check now prints `NOT CERTIFIED` instead of `INVALID` (text output and
`--quiet`), and `--json` gains a `verdict` field (`"VALID"` or `"NOT CERTIFIED"`)
beside the existing boolean `valid`. A failure means that english-arg-validator found no
accepted rule for some step. It does not show that the argument is invalid. Exit
codes, step-level `FAIL` lines and messages are unchanged. No verdict changes.

## v32.2: specific syntax hints

Syntax failures now say what to fix for the common slips: a missing step
number, a sentence without a final period, citations without the period before
them, and a conclusion marker without terminal citations (the expected
`... wet. [1,2]` shape). A lowercase `therefore,` has been accepted since v31.3.
Message-only: no verdict changes on the corpus (2,476 arguments).

## v32.3: `it's the case that`

`it's the case that X` and `it's not the case that X` (straight or curly
apostrophe) are read exactly as `it is the case that X` and `it is not the case
that X`, wherever the full forms are accepted. Other uses of `it's` (`it's
raining`) stay ordinary clauses. No verdict changes on the corpus (2,476
arguments).

## v33.0: English accuracy boundaries (H001–H185 regression set)

- Embedded reports, attitudes (with or without `that`), impersonal clause
  introductions, and sentence adverbs cannot become schematic terms, witnesses,
  or universal-instantiation subjects.
- Non-upward quantifier phrases cannot become schematic terms or UI subjects.
  Existential-generalization controls retain their separate existence rules.
- Whole-clause coordination under a leading operator requires explicit scope:
  parentheses or `both...and` / `either...or`. The syntax diagnostic gives both
  parenthesized readings instead of silently choosing one.
- Listed plural-only, lexically distinct, and ambiguous `-s` nouns no longer
  match their stripped stems.
- Intensional adverbs, attitude complements, and naming predicates block
  identity substitution, including the readable FOL conversion route.
- A comma before `which`, `who`, `whom` or `whose` (a non-restrictive relative)
  survives clause and term comparison. Other commas now survive too (see the
  v36.0 audit below). Non-restrictive relative weakening is handled directionally.
- Postposed `if` refuses embedded-question verb families, intervening objects
  and adverbs, and incomplete preposition/progressive/perfect endings.

How the guards are scoped (after measuring them on the 2,476-argument corpus):
an attitude or report word blocks a term only when it is used as a verb near
the start of the term: after a short subject phrase, before any preposition,
subordinator or comma, and not after a determiner, possessive or adjective
compound. A past form (`imagined`, `expected`) needs a `that`/`whether`/`if`
complement, and a verb without a complementizer must leave at most four words
of the term after it. `only` before a gerund (`only asking obvious questions`)
is not a quantifier. Coordination under a leading operator is ambiguous only
when the operator is a sentence adverb, an `it is ... that` frame, or an
attitude verb with a `that`/`whether`/`if` complement, and no comma precedes
the connective.

Corpus impact (2,476 arguments): 5 valid arguments newly fail syntax under the
operator-coordination rule. Three are real ambiguities: `Clark's analysis says
that A and the counterexample is blocked ...` (corpus-analytic-philosophy-maps-56bc431d8d),
`once the instrument is known, A and B` (fvb F53) and `it seems plausible that
... and ...` (askell-2018 A1215). Two are false positives: `Smith knows that A
and S knows that B` (corpus-analytic-philosophy-maps-0dbf5734db) and
capitalized names joined by `or` inside a noun phrase (askell-2018 A622).

Remaining known gaps:
- These are surface recognizers, not a parser. The new coordination guard
  catches unlisted embedding verbs and prepositional subjects in clause
  extraction/building. Unlisted intensional relations and term-binding
  contexts remain open; see the v36.0 audit.
- The morphology exception list is finite.
- A postposed `if` sentence with question-verb vocabulary or an incomplete-
  looking ending may be refused. Write `If ..., then ...` instead.

Control-list discrepancy: H171–H173 already rejected in the untouched checkout.
H171 (`Even if A, then B`) now accepts its valid forward conditional step.
H172 (`Only if A, then B`) and H173 (`Unless A, then B`) remain rejected: neither
licenses the supplied forward modus ponens. Their inclusion in the requested
VALID-control range conflicts with their actual repro contents.


## v34.0: accuracy round 2 (16 held-out cases)

Accuracy round 2 closes the 16 supplied false positives with explicit unless
scope, literal schema letters in proper names, a positive UI subject register,
quantified-binding guards, whole-statement UI, conservative noun morphology,
and opaque intensional identity contexts. `TestAccuracyRound2Cases` checks all
16 verdicts and their reasons; `TestAccuracyRound2ValidControls` checks nearby
valid forms, including long schematic bindings and upward-monotone `most`.
Scope of each guard, as measured on the 2,476-argument corpus:
- A schema letter is literal only for X, Y or Z, and only when every standalone
  occurrence of that letter in the whole statement follows the same
  capitalized word (`Plan X ... Plan X`). `If X are demands and Answering X`,
  `under SPC X knows`, `by Additivity X` and `a Future V` keep their variables.
- The universal-instantiation subject is refused only when its first word is
  a quantifier or determiner other than the/a/an (`almost no`, `just one`,
  `one of`, `someone`, `most`). Names, possessives, gerund phrases (`trying to
  remove all assumptions`) and bare noun phrases stay terms. The instance must
  be the premise's top-level copula: the subject carries no `if and only if`,
  `if`, `then`, `unless` or clause-level `and`/`or`.
- In schematic slots, only the non-upward `almost no`, `very few` and `just
  one` are added to the leading-operator guard. Upward existentials (`someone`,
  `something`, `one of the`) stay bindable, since `Someone is red` and `If X
  is red, then X is warm` do give `Someone is warm`.

Corpus impact (2,476 arguments, against v33.0): 6 valid arguments now fail.
Five newly fail syntax under the `unless` grouping rule, all real mixtures of
`unless` with `and`/`or`: corpus-ai-safety-historical 66aa497c54, 817bd8ec7c
and d8cb1a03fe, fvb B13, and askell-diag-b A2. One is the `Lois is looking for
Superman` case in the earlier blind-check set, which is now correctly rejected.
There are no other verdict changes and no invalid-to-valid changes.

The existing valid `most` and `more than` class-inclusion inferences now report
schematic instantiation rather than treating a quantified subject as an
individual UI witness. Fitch arbitrary names and modified bare plural class
generics retain their accepted readings.

## v35.0: identity substitution only in safe positions

Identity substitution (Leibniz's law) is certified only in known-safe positions,
an allowlist instead of a list of opaque verbs. A position is safe when the
substituted premise and the conclusion are explicit identity sentences
(identity chains), or when every replaced occurrence is the grammatical subject
of a plain clause. That means the start of the sentence, or of a clause after
`if`, `then`, `and`, `or`, `it is (not) the case that`, `such that` or an
opening parenthesis, and not followed by a passive (`is admired by`) or raising
form (`is believed to`). A capitalized proper name stays substitutable as the
subject of a modal sentence (`Superman can fly` gives `Clark can fly`). Every
other position is not certified, extensional objects (`Lois kissed Superman`)
included, and the message names the reason. The check runs on every route that
certifies identity substitution: the surface rule, the whole-term identity
register and the FOL rule. This closes the `is looking for` / `thinking about` /
`admires` class without listing verbs.

Schematic instantiation into a non-subject slot was left open in v36.0 and is
closed in v39.0; see below.

Corpus impact (2,476 arguments, against v34.0): 2 valid arguments now fail.
One is the rules.md example `every x is such that (x admires alice)`, which
substitutes into an object (the example now uses a subject); the other is
askell-2018 A609 (substitution inside `is a qualitative duplicate of ...`).
0 invalid-to-valid changes.

## v36.0: sentence-internal frames (variant A)

Steps that look inside a sentence (splitting or building and/or, identity
substitution, universal, existential and schematic instantiation) now need a
visible frame:

1. Openers. A sentence that opens with a sentence adverb (any `-ly` word, or
   perhaps/maybe/sometimes/often/once/always/never/yesterday/today) or a fronted
   framing phrase (a preposition, then a comma within the first few words: `In
   the story,`, `Under the law of likelihood,`) is not certified for these
   steps. The message suggests moving it out or rewording so the sentence
   starts with its subject.
2. Marked clause coordination. An unmarked `A and B` / `A or B` is split or
   built only when the left clause is a recognized complete clause (copula,
   listed intransitive, or listed extensional transitive with a nominal object).
   Otherwise write `Both ... and ...`, `Either ... or ...`, parentheses, or
   separate premises. This applies to elimination and introduction alike.
3. Commas count. Clause and term comparison keeps commas, except structural
   ones (`If A, then B`, `Therefore,`, the and/or grouping comma). H130-H134
   are closed.
4. Modified quantifiers (from the held-out list): a slot binding or
   universal-instantiation subject that opens with modifiers plus a quantifier
   word (`practically no`, `virtually no`, `next to no`, `precisely one`,
   `roughly one`) is refused.

Corpus impact (2,476 arguments, against v35.0): 597 valid arguments become NOT
CERTIFIED, chosen as variant A by the owner after a sample. 0 invalid-to-valid,
0 steps fail-to-pass. By rule: openers (mostly fronted frames) about 370;
marked coordination about 175; comma-exact 3; the rest are schematic bindings
that no longer match once commas count.

## v38.0: the remaining audit rows

- Relational instantiation and existential generalization into an object
  position are certified only for a closed allowlist of extensional verbs
  (greets, sends, gives, causes, outweighs, contains, ...; not `admires`, v38.1) and copular frames
  (`is preferable to`, `is taller than`, ...), and their passives. Any other
  verb's object may be intensional (`Lois awaits every guest`, `seeks`,
  `craves`), so it is refused. The message suggests stating the instance as a
  premise. Subject positions are unchanged.
- Quantity expressions are a closed class: numerals before a plural noun
  (`zero robots`, `two objections`) and quantity nouns (`a minority of`, `a
  handful of`, `half of`). They are never individuals for instantiation and,
  except the upward-monotone `a majority of`, `lots of`, `plenty of`, never fill
  a schematic slot. `Zero probability` and `the proportion of ...` name values
  and stay terms.
- The unless-anaphora check uses the full closed class of pronouns,
  demonstratives and pro-forms (`this`, `one`, `such`, ...).
- Adjectives that take an if/whether complement (`unsure`, `hesitant`, ...)
  block the postposed `C if A` reading; the class is listed in full.

## v37.0: closed-class openers and structural term rules

- Openers: every preposition (single-word, including -ing forms such as
  `according`, and complex: `according to`, `as per`, `in view of`, `with
  respect to`, ...), the listed non-`-ly` sentence adverbs and multi-word
  adverbials (`of course`, `in theory`, `so to speak`, ...), and every `-ly`
  word refuse a sentence-internal step, with or without a comma. `According to
  Bob, A and B. Therefore, B.` is NOT CERTIFIED. (v36.0 let `According` through
  as if it were a gerund subject.)
- Terms: a term that contains a finite verb (or a base verb after a plural
  subject) followed within three words by a complement is refused in schematic
  slots, universal instantiation, existential generalization and definite
  descriptions, whatever the verb (`the observer reckons Alpha` + `is red`).
- Raising predicates (`seems to`, `appears to`, `happens to`, `tends to`, ...)
  and irregular passive participles block identity substitution.
- Pluralia tantum (`tongs`, `scissors`, `trousers`, ..., `-ics` fields) never
  match a stripped singular.
- A postposed `C if A` is not read as a conditional when the word before `if`
  is a verb that may take the if-clause as its object (`ponders if`).

Corpus impact (2,476 arguments, against v36.0): 391 valid arguments become NOT
CERTIFIED, under the strict standard (ambiguity is refused even at a corpus
cost). Most are prepositional openers without a comma (`In infinite worlds
...`, `Under X ...`) and terms that end inside an embedded clause. One argument
becomes VALID again: `Nearly all ...` is a quantifier opener, as `Almost every`
already was, which restores its v35.0 verdict.

## v36.0 blocklist audit

This audit distinguishes a positive grammar recognizer (**allow**) from a
negative lexical filter (**block**). “VALID path” includes parsers and
normalizers whose output can feed a successful certificate; a refusal-only
helper is not automatically harmless when *absence* of its words permits a
parse. Runtime sets of seen formulas, bound variables, names and caches are
not lexical word lists. File names below are relative to `validation/`.

**Audit status (v38.0): every row settled** as (a) a closed class enumerated in full, (b) a structural rule, or (c) a refusal-only list whose accepting path is fixed elsewhere, each with a regression test.
and complement-coordination examples, and comma-exact comparison closes
H130–H134. The rows marked **open** are genuine remaining obligations; this
change does not claim that an incomplete lexical blocklist proves a safe
frame. They need a separate positive grammar or an explicit owner decision
about the narrower accepted language before the audit requirement is met.

| File and list / recognizer | Kind | VALID path? | Disposition / reason |
| --- | --- | --- | --- |
| `sentence_frames.go`: `frameOpeners` | allow | yes | Determiners, pronouns and logical entry points end the opener check early. |
| `sentence_frames.go`: `-ly` opener rule | closed-class rule | yes | Every `-ly` word at the start is refused (morphological, so unlisted adverbs are covered); `frameLyNouns` lists the few `-ly` nouns that stay subjects. |
| `sentence_frames.go`: `frameSentenceAdverbs` + `frameAdverbialPhrases` | closed class | yes | v37.0: the non-`-ly` sentence adverbs and multi-word adverbials (of course, no doubt, in theory, in practice, so to speak, in effect, ...) are listed as a closed class; every `-ly` word is refused by morphology. Settled. |
| `sentence_frames.go`: `framePrepositions` + `frameComplexPrepositions` | closed class | yes | v37.0: every single-word preposition (including -ing forms such as according/regarding and fronting subordinators) and the complex prepositions (according to, as per, in view of, with respect to, ...) refuse a sentence-internal step with or without a comma. Settled. |
| `sentence_frames.go`: name evidence (capitalized mid-sentence elsewhere) | allow (exception) | yes | Lets a sentence-initial name through the opener check; it cannot turn an adverb into a name unless the same word is capitalized mid-sentence. |
| `sentence_frames.go`: `frameTransitives`, `frameIntransitives`, `framePredicates`, `frameNominalHeads` | allow | yes | Rule 2: an unmarked `A and B` split needs a recognized complete left clause; unknown verbs require explicit grouping (`Both`, `Either`, parentheses, separate premises). |
| `sentence_frames.go`: coordination prefixes | allow | yes | Explicit both/either grouping and conditional scaffolds are handled by their own parsers. |
| `sentence_frames.go`: `modifiedQuantifierPhrase` | closed-class rule | yes | A slot binding or UI subject that opens with modifiers (any `-ly` adverb, `next to`, `close to`, `well`, `just`, ...) followed by a closed class of quantifier words (`no`, `one`, `few`, `all`, numerals, ...) is refused: `practically no`, `virtually no`, `next to no`, `precisely one`, `roughly one`. The quantifier words are a closed class; the modifier side is open by morphology. |
| `entailment_gate_accuracy.go`: `bindingAttitude` | block | yes | Backstopped by the v37.0 structural rule `clauseInsideTerm`: a term containing any finite verb (or a base verb after a plural subject) followed within three words by a complement is refused, whatever the verb. The list remains as an extra refusal. Settled. |
| same: `sentenceOperator` | block | yes | Additional negative filter; new opener recognizer rejects unlisted productive `-ly` adverbs. Non-`-ly` nominal/adverb ambiguity remains a residual. |
| same: `impersonalEmbedding` | block | yes | Backstopped by `clauseInsideTerm` ("it appears that alpha": finite verb + that). Settled. |
| same: `nominalAttitudeHead` | allow (exception) | yes | An exception inside the list check only; `clauseInsideTerm` runs first and does not consult it. Settled. |
| same: `nominalClause` | allow (boundary) | yes | Closed relative/complement markers; full nominal parse still depends on other guards. |
| same: `attitudeScanStop` | allow (exception) | yes | Exception inside the list check only; `clauseInsideTerm` scans the whole term. Settled. |
| same: `attitudeNounContext` | allow (exception) | yes | Exception inside the list check only; `clauseInsideTerm` uses its own determiner/preposition rule. Settled. |
| same: `attitudeBaseSubject` | allow (verb cue) | yes | Superseded by `clauseInsideTerm`, which treats a base verb after a plural subject as finite. Settled. |
| same: `attitudeUsedAsVerb` next-token switch | allow (exception) | yes | Exception inside the list check only; `clauseInsideTerm` decides independently. Settled. |
| same: `nonSingularStem` | block | yes | v37.0: English pluralia tantum and lexically distinct -s nouns listed in full (tongs, scissors, trousers, ..., -ics fields) as a closed class. Settled. |
| same: `identityOpaqueExtra` | block | yes | v35 subject-position guard independently protects legacy identity substitution; **open** in other FOL/term-extraction uses. |
| same: `questionStem`, `incompleteIfConsequent` | block | yes | Backstopped by a structural rule (v37.0): a postposed `C if A` is not read as a conditional when the word before `if` is a finite verb not on the intransitive list (and not a participle after a copula), or any word after a pronoun subject. Settled. |
| same: `operatorCoordinationAmbiguous` prefix exemptions | allow (exception) | yes | Structural operators own their connectives; this early syntax guard still has legacy opaque-context exemptions. |
| same: `relativeComma` | allow | no | Legacy unused relative-marker recognizer; it no longer controls comparison. |
| same: `nonrestrictiveRelative` | allow | yes | Positive directional relative weakening pattern; general matching retains lexical commas. |
| same: `structuralModalComma`, `structuralCoordinationComma` | allow | yes | Modal prefix commas and visible complete-clause coordination boundaries only; explicit literal payloads are masked. |
| `term_slot_positions.go`: `clauseSubjectLead` | allow | yes | Explicit clause/subject boundary markers. |
| same: `passiveOrRaising` | block | yes | v37.0: raising predicates (seem, appear, happen, tend, prove, turn out ... to) are a closed class; passives are caught by -ed/-en morphology plus a closed list of irregular participles. Settled. |
| same: `explicitIdentityPattern` | allow | yes | Only explicit identity phrases or equality trigger identity-chain treatment. |
| `entailment_gate_surface.go`: contraction alternatives | allow | yes | Fixed lexical expansions, not an open-ended synonym rule. |
| same: `gateSimpleClause` | allow | yes | Positive finite grammar for the old inline conditional rewrite. |
| same: `gateUnsafeSurface` | block | yes | Cannot independently admit a clause: `simpleGateClause` also requires `gateSimpleClause`. Additional exclusions only narrow that allowlist. |
| same: `normalizeGateConditional` if/when/whenever prefixes | allow | yes | Fixed structural operators. |
| same: `postposedIfBlockedBefore` | block | yes | v38.0 (a)+(b): the adjectives that take an if/whether complement (unsure, hesitant, curious, aware, ...) are listed in full, and since v37.0 any verb before `if` refuses structurally. Test: `TestAuditPart2Rules/hesitant if`. Settled. |
| same: `postposedConsequentUnsafe`, `postposedAntecedentUnsafe` | block | yes | (c) Refusal-only lists; the accepting path for `C if A` is fixed by the structural verb-before-`if` rule and the closed adjective class in the row above, so these lists cannot certify anything. Settled. |
| same: `gateElidedCopula`, `gateBarePredicate` | allow | yes | Positive copular and predicate shapes; lexical homonymy is outside these surface recognizers. |
| `andor_grouping.go`: connective list in `groupingBoundaries` | allow | yes | Exactly and/or/unless. |
| same: `groupingOpaque` exemptions | block | yes | (c) Grouping is read only on clauses that pass the v37.0 opener rule and the structural term rule, so an attitude or framing context around the coordination is refused first. Test: `grouping in belief`. Settled. |
| same: `groupingProblem` operator prefixes and conditional separators | allow | yes | Structural grammar, not lexical scope inference. |
| same: `clauseStartWords`, `determinerWords`, `auxiliaryWords` | allow | yes | Positive operand cues; suffix-based verb cues remain heuristic. |
| same: `groupingParenthesisPlacement` before/after operator lists | allow | yes | Only whole-operand boundary positions are recognized as grouping. |
| same: `groupingLabel` | allow | yes | Exact definition/condition/proposition label annotation syntax. |
| `entailment_gate_schemas.go`: `conditionalSplitPattern`, `separatedQuantifierPattern` | allow | yes | Explicit then boundary and every/no thing spelling. Structural comma consumption is localized here. |
| same: `hasOuterGateOperator` prefixes | allow (operator recognition) | yes | (a)+(b) A sentence operator in front of a clause is either a listed operator or, since v37.0, refused: every -ly word, every preposition and the closed class of non-ly sentence adverbs. Tests: `outer necessarily`, `outer possibly`. Settled. |
| same: universal/negative/particular/existential quantifier prefixes and copula separators | allow | yes | Enumerated quantifier grammar; other strings are not silently assigned that quantifier. |
| same: negative-instance and ground-instance excluded prefixes | block | yes | (a) v38.0: a quantifier-led term is recognized from closed classes in full: determiners (`termDeterminers`), modified quantifiers (v36.0), and quantity expressions (numerals before a plural noun, quantity nouns: minority, majority, handful, half, ...). Tests: `minority UI`, `zero schematic`. Settled. |
| same: necessity and wrapper prefixes | allow | yes | Fixed logical scaffolds. |
| same: `termCandidates` article/copula prefixes | allow | yes | Fixed term normalization; residual morphology is listed separately. |
| same: `schematicBindingLeadingOperators` | block | yes | (a)+(b) Leading operators are the closed opener classes (v37.0) and the quantifier classes above (v38.0). Settled. |
| same: `schematicBindingComplementTail` | block | yes | (b) Superseded by the v37.0 structural rule `clauseInsideTerm`: any finite verb with a complement inside a captured term refuses, whatever the verb. Settled. |
| same: `schematicBindingClauseVerbs` | allow (exception) | yes | A finite-verb cue allows a complete relative inside a term; does not by itself establish nominality. |
| same: `thirdPersonSingularPredicate` quantified-object switch | block | yes | (c) The switch changes only the verb agreement of the subject instance (`has` for `have`); the object is copied verbatim and since v38.0 object positions are instantiated only for allowlisted extensional verbs. Settled. |
| same: have/do/go and inflection endings | allow / productive rule | yes | Fixed irregular agreement and regular spelling; semantic predicate text is otherwise unchanged. |
| same: schema form-name lists | allow | yes | Internal rule identifiers, not natural-language vocabulary. |
| `accuracy_round2.go`: `nameBeforeLetter` operator switch | allow (exception) | yes | Keeps logical operators from being treated as a multiword name prefix. |
| same: `round2Singular` exception switch and suffix rules | block / productive rule | yes | v37.0: English pluralia tantum and lexically distinct -s nouns listed in full (tongs, scissors, trousers, ..., -ics fields) as a closed class. Settled. |
| same: `instanceQuantifierHeads`, `instanceQuantifierPhrases` | block | yes | (a) Same closed quantifier classes as the instance-prefix row (v38.0 `quantityPhrase`). Settled. |
| same: `accuracyRound2Failure` quantifier prefix list | block | no | Diagnostic only, after all candidate schemas fail. Cannot make a step VALID. |
| same: `subjectHasConnective` separators | block | yes | Closed connective syntax, but the subject recognizer also uses open-ended lexical cues. |
| `entailment_gate_v11.go`: `nonReferringTermWords` | block | yes | (a) Same closed quantifier and pronoun classes (v38.0 `quantityPhrase`, full pronoun class). Settled. |
| same: `contextQuantifierWords` | block | yes | (a) Same closed quantifier classes; a description bound by an omitted context is refused by the v37.0 term rule. Settled. |
| same: `opaqueContextWords` | block | yes | (a) v38.0: relational instantiation and generalization into an object position are certified only for a closed allowlist of extensional verbs and copular frames (`extensionalVerbHeads`, `extensionalCopulas`), so an unlisted intensional verb (seeks, awaits, craves, worships) is never substituted; passive and raising subjects are blocked by `passiveOrRaising`. Tests: `object UI awaits`, `object UI craves`, `object EG seeks`. Settled. |
| same: `truthFunctionalWrapper`, schematic-existential pattern | allow | yes | Explicit logical scaffolds only. |
| same: `referringTerm` definite-head switch | block | yes | Closed definite heads, but the surrounding context-quantifier detection remains open. |
| `entailment_gate_v14.go`: `fitchForms`, `fitchNamePattern` | allow | yes | Rule identifiers and identifier syntax. |
| `entailment_gate_v15.go`: possibility prefixes | allow | yes | Fixed logical operator. |
| `entailment_gate_v16.go`: `leadingModal` | allow | yes | Fixed modal auxiliaries; does not remove arbitrary adverbs. |
| `entailment_gate_v17.go`: `relationWords` excluded operators | block | yes | (a) Same extensional allowlist as `opaqueContextWords`. Settled. |
| same: `relationTermCandidate`, `relationHasQuantifier` switches | allow | yes | every/each/some/no bind slots; others do not bind through this grammar. |
| same: `relationVerb` excluded first words | block | yes | (a) Same extensional allowlist; an unlisted verb can no longer receive an object substitution. Settled. |
| same: `relationNestedNoun` prepositions/quantifiers | allow | yes | Recognizes explicitly enumerated nested scope positions. |
| `entailment_gate_v19.go`: `nonMaterialAuxiliary`, `mightPerfectAuxiliary`, `counterfactualMood` | allow | yes | Explicit modal/mood syntax. |
| same: `counterfactualUnsafe` | block | yes | (c) No counterfactual form is in the catalog; tests show antecedent strengthening, contraposition and chaining on `had ... would have` conditionals are refused. Tests: `cf strengthening`, `cf contraposition`, `cf chaining`. Settled. |
| same: copular tense replacement alternatives | allow | yes | Local, explicitly enumerated counterfactual auxiliary conversions. |
| `entailment_gate_v20.go`: `arbitraryPropertyPattern`, possessive prefixes | allow | yes | Explicit property-binder and possession syntax. |
| `entailment_gate_v22.go`: n-place preposition and ditransitive switches | allow | yes | Only enumerated slot boundaries and double-object alternations. Other lexical safety inherits `relationWords`. |
| `entailment_gate_v23.go`: relative markers, quantifier and relation-verb switches | allow | yes | Explicit relative grammar; inherits the lexical relation limitation. |
| `entailment_gate_v24.go`: `explicitBinder`, variable/name/word patterns | allow | yes | Explicit FOL syntax; atom names are uninterpreted. Readable fallback inherits older lexical limits. |
| `entailment_gate_v25.go`: readable quantifier families, `identityQuantifierWords` | allow / block | yes | Quantifier families are positive; identity-term exclusion is **open** without independent positive referring-term grammar. |
| `entailment_gate_v25_literals.go`: `folLiteralStart`, `folNameEscape`, JSON literal syntax | allow | yes | Explicit uninterpreted atom/name syntax, with literal payload preserved. |
| `entailment_gate_v26.go`: counting prefixes/numerals, signed/wrote predicates | allow | yes | Finite counting grammar; class stemming still shares the morphology limitation. |
| `entailment_gate_v27.go`: relative/article/grouping marker lists | allow / block | yes | Positive description grammar plus negative ambiguity exclusions; **open** where readable relation parsers supply lexical meaning. |
| `entailment_gate_v28.go`: modifier list, relational verb list, quantifier and identity separator lists | allow | yes | Explicit rewrite candidates; identity substitutions also pass the subject-position check. |
| `entailment_gate_v29.go`: `unlessAnaphor` | block | yes | (a) v38.0: the full closed class of English pronouns, demonstratives and pro-forms (this, that, one, such, same, there, ...). Tests: `unless one`, `unless this`. Settled. |
| `entailment_gate_repeated_slots.go`: `repeatedSlotUniversal`, `repeatedSlotExistential` | allow / block | yes | Explicit universal recognition; repeated-slot existential exclusion remains **open** for unlisted quantificational expressions. |
| same: `repeatedSlotDisjunctSingular`, `repeatedSlotRelativeThenVerb` | allow / block | yes | Positive short-term shape and relative boundary exclusions; not a full nominal grammar. |
| `numbered_argument.go`: conclusion markers and their noun-continuation exceptions | allow / block | yes | Fixed markers; continuations disambiguate ordinary “so ...” statements. They affect parsing, not an entailment by themselves. |
| same: premise/conclusion/reference/fence patterns | allow | yes | Numbered-argument syntax and explicit numeric references. |
| `numbered_argument_scope.go`: arbitrary-name pattern | allow | yes | Explicit binder syntax. |
| `entailment_gate_expand.go`: bridge pattern | allow | yes | Explicit cited-step expansion grammar. |
| `step_reference_hint.go`: parenthesized reference pattern | allow | no | Diagnostic hint only. |

### Residual frame limits

The new guard is a surface recognizer, not a part-of-speech parser. Bare-noun
and gerund recognition uses predicate cues; noun/verb homonyms and complex
nominal subjects remain uncertain. The final guard currently classifies rule
families by certificate name; a new rule that inspects sentences must be added
to that classification. Explicit grouping of a whole operand permits that
operand to remain opaque. A complement's own parentheses do not make its
content asserted. Unknown intransitive verbs can be conservatively refused;
use explicit grouping for those cases. No corpus impact percentage was
measured in this run.

## v41.0: "Both A and B" conjunctions

- The clause-coordination refusal (v36.0) tells the writer to use `Both ...
  and ...`, but conjunction introduction and elimination did not split that
  form, so following the hint still failed. `Both A and B` now has conjuncts A
  and B. Mixed `Both A and B or C` and a `both` inside an attitude report stay
  refused.

## v40.0: admires is not extensional

- `admires` comes off the extensional allowlist for relational instantiation
  and generalization into an object position: admiration may be intensional
  (Lois admires Superman, not Clark). `Lois admires every hero. Bruce is a
  hero. Therefore, Lois admires Bruce.` is NOT CERTIFIED. `greets` is added,
  and the relational docs examples use it.
- Corpus impact against v38.0: 1 valid argument becomes NOT CERTIFIED.

## v39.0: indefinites bind schematic variables only as subject or predicate

The gap left open in v36.0 is closed. An indefinite or quantifier phrase
(`a`, `an`, `some`, `any`, `someone`, `something`, `somebody`, `anyone`,
`anything`, `anybody`, `several`, `many`, `one` ...) can bind a schematic
variable only at the start of a clause (its subject) or after a copula (`is a
hero`). In any other position, the object of a verb or a preposition, inside an
infinitive or that-clause, or a complement, the phrase need not pick anything
out, so the schema is not instantiated with it:

```text
1. If Lois is looking for X, then X exists.
2. Lois is looking for a unicorn.
3. Therefore, a unicorn exists. [1,2]
```

is NOT CERTIFIED, and the message names the binding and the slot:
`premise 2 binds a schematic variable to "a unicorn" in an object position
("lois is looking for X"); an indefinite there need not pick anything out, so
the schema cannot be instantiated with it. State the existence claim as its own
premise, e.g. "There is a unicorn that Lois is looking for."`

The message is used whenever a step would pass but for this rule, even if
another form came nearer. Names and definite descriptions (`Superman`, `the
mayor`) still bind in any slot.

The rule is positional, not lexical, so it also refuses the extensional
`If Lois admires X, then X is admirable; Lois admires a hero; therefore a hero
is admirable`, which is valid on every reading. A narrower rule that lists
intensional verbs (`seeks`, `wants`, `needs`, `imagines`, `is looking for`) was
not adopted: the class is open, and a list cannot guarantee that every case
invalid on some reading is still refused.

Repair: state the existence claim as its own premise, as the message suggests,
or write the phrase into the supplied rule instead of a variable
(`If Lois is looking for a unicorn, then a unicorn exists.`), which makes the
conditional an explicit premise rather than a schema instance.

Corpus impact (2,476 arguments, against v38.0): 57 valid arguments now fail;
0 invalid-to-valid changes. Mutants (8,109): 2 valid-to-invalid; 0 new
acceptances. Every one of the 57 gets the slot message above. The 57 by slot:
object of a preposition 36, object of a transitive verb 9, comparison or
predicative complement 5, inside an infinitive or that-clause 4, object of an
intensional verb 2, embedded subject after an adjunct 1 (counted as
non-initial by this surface rule). In the measured run, 56 of the 57 passed
again after a supplied-step rewrite; the other then hit the separate
nested-quantifier limit of the relational grammar.
