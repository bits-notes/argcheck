# Entailment Gate — Coverage and Known Limits (catalog v26)

**Type:** reference
**Status:** current as of catalog v26 (2026-11-10)


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
`tools/capture_legacy.py` reproduces the capture in a v24 checkout. Each row
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
inference forms are unchanged. The catalog remains **v25**. Tests cover capture,
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
they are not themselves a proof of completeness or unbounded soundness. The
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
  Definite descriptions remain unimplemented (v27).
- Modal, counterfactual, property and that-clause grammar are outside the
  function-free FOL claim. An opaque propositional letter carries no internal
  modal or attitude semantics.
- Paraphrase remains a permanent non-goal. Zero-place atoms render verbatim;
  different normalized clauses remain independent propositional letters.

Soundness evidence is bounded, **not a proof**: the finite-model oracle checks
domains 1–3, empty predicate extensions, and co-denoting names. v24 checks a
fixed 10/50 sample of every inventory rule, plus exercises and local Fitch
obligations with eigenname closure. v25 adds bidirectional conversion samples
and independently hand-built meanings for representative readable sentences,
alongside ambiguity and near-miss rejections. The identical-AST criterion is
sound by construction conditional on faithful parsing; the tests do not prove
faithfulness for arbitrary English or soundness over all domain sizes.

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

Uniqueness requires membership in the counted class. The supplied b1 lacks
`Scott is a person`: a nonperson Scott and a distinct person may both write
Waverley. The two-object countermodel forces rejection. Adding that membership
to the cited ground conjunction yields the intended valid uniqueness step.
The supplied b3 is supported; definite-description b2 remains unsupported.

The v26 tests check accepted local and discharged proof obligations over domains
1–4, including empty extensions and co-denoting names, and compare all 15 count
abbreviations against independently assembled identity formulas. Counts 4 and
5 have vacuous lower-bound checks on smaller domains; these bounded checks do
not establish unbounded soundness. Invalid exercises each require a countermodel.

Compatibility: existing successful proofs keep their historical schema and
report; counting is a fallback for new inferences. Explicit out-of-range bounds
still receive the counting-bound diagnostic. The 290-argument corpus comparison
checks both verdicts and normalized report hashes.
