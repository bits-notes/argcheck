---
name: argcheck
description: Check a numbered argument for syntactically transparent validity with the deterministic `argcheck` CLI (offline Go checker, no LLM). Use whenever drafting, checking, or repairing a premises-to-conclusion argument: write in the gate's numbered format, run `argcheck`, read any step mismatch, and repair until it passes while keeping fidelity to the source text.
---

# argcheck — deterministic validity check for numbered arguments

`argcheck` runs two deterministic checks in Go (`numbered_argument_syntax`, then
the catalog `v31.2` entailment gate) with no LLM and no network access.

**Validity is syntactic: the gate is the standard. Never overrule a gate failure
with your own judgment that a step "obviously follows."** If you think a step is
valid, rephrase it into a catalog form or make the missing link a visible
premise. If no phrasing works, report that the inference is outside the catalog
rather than calling the step valid.

**Fidelity to the source is your job, not the gate's.** `argcheck` checks only
whether each conclusion step follows syntactically from its cited steps; it
never checks whether premises are true or faithful to a text.

## Run it

```bash
argcheck <<'EOF'
1. If it rains, then the street is wet.
2. It rains.
3. Therefore, the street is wet. [1,2]
EOF
```

- `argcheck file.txt` or stdin (`argcheck < file.txt`, `cat file.txt | argcheck -`).
- `--json` prints structured syntax, step reports, and provenance.
- `--quiet` prints only `VALID` or `INVALID`.
- `--forms` lists all 121 accepted inference forms in catalog `v31.2`.
- `--version` prints the catalog version and build stamp (`gate catalog v31.2; local argcheck@<commit>`).
- Exit codes: `0` valid, `1` invalid, `2` input or operational error.
- If `argcheck` is not installed yet, install with
  `go install github.com/bits-notes/argcheck/cmd/argcheck@latest` or run
  `./install-argcheck.sh` from a checkout of the repository.

Output reports each conclusion step as `ok: <form>` or
`FAIL: nearest form <form>: <exact mismatch>`. When a conclusion fails only
because it cites an extra step that does no work, a `hint:` line names the form
that passes on the used subset and the inert citation(s) to remove.

## Format (strict)

- **Premise:** `1. Statement.` — one line ending in a period, with no
  `Therefore,` prefix and no trailing brackets.
- **Conclusion:** `3. Therefore, statement. [1,2]` — requires both `Therefore,`
  and trailing citations to earlier steps.
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
1. If it rains, then the street is wet.
2. If the street is wet, then the cars slow.
| 3. Suppose it rains.
| 4. Therefore, the street is wet. [1,3]
| 5. Therefore, the cars slow. [2,4]
6. Therefore, if it rains, then the cars slow. [3-5]
```

## The repair loop

1. **Draft** the argument in the strict numbered format and run `argcheck`.
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
by watering down or overstating the source is a failure, not a fix.

## Passing is not the same as reconstructed

`If P1 and P2 and P3, then C.` plus `P1`–`P3` always passes as conjunctive modus
ponens, putting all the disputed weight into one ad hoc conditional. An argument
built out of `If [my premises], then [my conclusion]` bridges is a sufficiency
squish, not a reconstruction. Replace ad hoc bridges with the **general rule**
the author relies on (for example, `Every interval procedure that misses its
advertised coverage is unfit for reporting uncertainty.`), then instantiate it.

## Writing so the gate can see the inference

Clauses match **verbatim** after light normalization (case, whitespace,
contractions, optional `then`). Synonym and paraphrase tolerance are permanently
out of scope: if two wordings mean the same thing, state that equivalence as a
visible identity or conditional premise.

- **Repeat clauses word for word:** `the street is wet` does not match
  `the road is wet`.
- **Negation:** write `It is not the case that X.` Modus tollens fails on
  `The street is not wet` against `the street is wet`, and passes on
  `It is not the case that the street is wet`.
- **Universals and syllogisms:** `Every man is mortal.` + `Socrates is a man.`
  passes; `All men are mortal.` fails against singular `a man`. For class
  syllogisms, `Every A is a B.` / `All A are B.` and `No B is C.` work.
- **Connectives:** disjunction is `Either X or Y.`; conjunction is `X and Y.`;
  biconditional is `X if and only if Y.` or parenthesized `(X) if and only if (Y)`
  (optionally prefixed by `It is the case that`) when either side contains
  another connective. `It holds that` is ordinary text.
- **One hop per step:** chain intermediate conclusions and cite only the steps
  used by that single rule. Probe unfamiliar shapes with a 3-line test argument
  before writing a long chain.
- **"X rather than Y" conclusions:** state the dichotomy premise explicitly
  (`Either we should use X or we should use Y.`), refute one side, and apply
  disjunctive syllogism.

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

What `argcheck` does not do:

- **Premise truth:** it never judges whether premises are true.
- **Source fidelity:** it never checks whether an argument matches a source text.
- **Proof search or paraphrase:** it checks only the steps written, without
  synonym tolerance, general arithmetic (`n > 5`), modal systems beyond T,
  proportional quantifiers (`most`), or non-deductive inference.

See `docs/rules.md` in the `argcheck` repository for all 121 forms with accepted
and rejected examples, and `docs/coverage.md` for full grammar boundaries.
