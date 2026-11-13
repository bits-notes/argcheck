# argcheck

argcheck checks the syntactically transparent validity of numbered arguments.
It runs deterministic syntax and entailment checks in Go, with no LLM and no
network access in local mode. Each conclusion must match an accepted inference
form using its cited steps. The current catalog is v28; `argcheck --forms` lists
its forms.

A passing result establishes validity under that catalog. It does not establish
that premises are true, that an argument faithfully represents a source, or that
an unstated assumption is justified. This is not a general theorem prover:
inferences outside the catalog can fail even when they are logically valid.

## Install

Requires Go 1.25.0 or later. From a checkout:

```sh
./install-argcheck.sh
```

The script installs to `$HOME/.local/bin`; set `BIN_DIR` to override it. It stamps
the binary with the current commit and adds `-dirty` when `cmd/argcheck` or
`validation` has uncommitted changes. Put the destination on your `PATH`.

Alternatively, from the checkout, install the latest published module version
(the module path is read from `go.mod`):

```sh
go install "$(go list -m)/cmd/argcheck@latest"
```

This uses Go's usual `GOBIN` or `GOPATH/bin` destination. Access to the module and
a published version are required. The script is preferable when you need an
explicit commit stamp.

## Numbered input format

```text
1. If it rains, then the street is wet.
2. It rains.
3. Therefore, the street is wet. [1,2]
```

- A premise is one line ending in a period, without `Therefore,` or citations.
- A conclusion uses `Therefore,` and citations after the period: `[1,2]`.
  Citations refer only to earlier steps.
- Number steps sequentially from 1. Do not include blank lines, leading spaces,
  code fences, or commentary in the argument. The final step is the conclusion.
- Every premise and intermediate conclusion must be cited by a later step.
- A bridge premise such as `If [1,2], then C.` expands those numbered steps into
  its antecedent. Such a bridge is an explicit assumption, not evidence for it.

Fitch subproofs use one `| ` prefix per depth level. `Suppose X.` opens an
assumption; discharge cites the entire subproof as `[i-j]`. For example:

```text
1. If it rains, then the street is wet.
2. If the street is wet, then the cars slow.
| 3. Suppose it rains.
| 4. Therefore, the street is wet. [1,3]
| 5. Therefore, the cars slow. [2,4]
6. Therefore, if it rains, then the cars slow. [3-5]
```

A step cannot cite into a closed subproof. Reductio subproofs end with
`Therefore, contradiction. [i,j]`. The catalog also supports proof by cases,
universal generalization (`Let a be an arbitrary F.`), and existential
elimination with fresh-name checks.

## CLI usage

```sh
argcheck argument.txt
argcheck < argument.txt
cat argument.txt | argcheck -
argcheck --json argument.txt
argcheck --quiet argument.txt
argcheck --forms
argcheck --version
```

Text output reports `VALID` or `INVALID`, the catalog and build stamp
(`gate catalog v28; local argcheck@<commit>`), and each conclusion's matched
form or mismatch. The verdict is always the first line; Fitch input is echoed
after the step report. An inert-citation
hint identifies extra citations whose removal would let a known form pass.

For scripts, use the exit code (0 valid, 1 invalid, 2 error, 3 divergent with
`--both`) or `--quiet`, which prints only `VALID`, `INVALID` or `DIVERGENT`:

```sh
argcheck --quiet argument.txt          # VALID
for f in args/*.txt; do argcheck --quiet "$f"; done | grep -c '^VALID$'
```
`--json` returns structured checks, details, and provenance.

Remote comparison is opt-in:

```sh
argcheck --prod argument.txt
argcheck --both argument.txt
argcheck --both --proxy http://127.0.0.1:17779/mcp argument.txt
```

`--prod` calls the remote validate endpoint (MCP); `--both` runs local and remote
checks and flags disagreement. The default endpoint is the localhost URL above.
It must be running and provide the expected validation tool. Remote results can
differ when the endpoint uses another catalog or implementation revision.

Exit codes: 0 valid, 1 invalid, 2 operational or input error, 3 local/remote
divergence. Put flags before the filename.

## Using the validation package from Go

Run syntax validation first, then the entailment gate, as the CLI does:

```go
package main

import (
    "context"
    "fmt"

    "github.com/bits-notes/argcheck/validation"
)

func main() {
    argument := "1. If it rains, then the street is wet.\n" +
        "2. It rains.\n" +
        "3. Therefore, the street is wet. [1,2]"
    syntax := validation.Validate(context.Background(),
        validation.BuiltinSpec(validation.BuiltinNumberedArgumentSyntax), argument)
    if !syntax.Passed {
        fmt.Println(syntax.Message)
        return
    }
    report, result := validation.CheckEntailmentGate(argument)
    fmt.Println(result.Passed, result.Message)
    fmt.Println(report)
}
```

`EntailmentGateCatalogVersion` exposes the catalog version;
`EntailmentGateForms()` lists accepted form names. `CheckEntailmentGate` returns
both a step report and a validation result. Package calls do not contact a remote
service.

## Writing so the gate can see the inference

Clauses match word for word after limited normalization, including case,
whitespace, contractions, and optional `then`. Synonym and paraphrase tolerance
are deliberately outside the gate. If two wordings should count as equivalent,
state the connection as a visible identity or conditional premise.

- Repeat clauses exactly: `the street is wet` does not become `the road is wet`.
- Use canonical negation: `It is not the case that X.` For modus tollens,
  `The street is not wet` does not substitute for the canonical negation of
  `the street is wet`.
- For instantiation, use `Every man is mortal.` with `The traveler is a man.`
  `All men are mortal.` does not match the singular `a man`.
  For class syllogisms, use `All A are B.` and `All B are C.`
- Write disjunction as `Either X or Y.` and conjunction as `X and Y.`
  A conditional with a conjoined antecedent can use the separately cited
  conjuncts in conjunctive modus ponens.
- Make one inference hop per conclusion. Cite exactly the premises the form
  uses; extra citations can cause rejection.
- Copular parsing splits at the first ` is `. Give a short name to a subject
  containing an embedded `is`, then use that name consistently.
- Identity substitution supports whole gerund terms and coordinated proper names
  in its controlled argument positions, including predicates containing `and`.
  Examples appended with `such as` can still prevent placeholder binding; keep
  examples outside the numbered argument unless they are separate premises.

When a step fails, inspect the quoted mismatch, correct wording or citations,
and rerun. Test an unfamiliar shape in a small argument first. Do not silently
drop qualifiers or change a source's meaning to obtain a pass. Any supplied
inferential link should be a visible premise.

### Toolkit for general rules

Schematic conditionals can bind `U`, `V`, `W`, `X`, `Y`, and `Z` to terms or whole clauses:
`If X depends on a prior, then X is unfit for science.` instantiates against
`Bayesian inference depends on a prior.` Possessives (`X's`), conjunctive
antecedents, and nested conditionals work. Placeholders must agree in number
with their bindings: use `If X are fixed...` for a plural subject.

Universal syllogisms accept `Every A is a B.` plus `Every B is C.`, including
negative variants such as `No B is C.` Relative clauses and internal commas
can occur, but the repeated terms must match verbatim.

Copular existential introduction accepts: `J is preferable to M.`
and `J is an E.` can yield `Some E is preferable to M.` Non-copular verbs use the relational forms described below. Use a definite or bare individual subject, rather than
`A ...`. `Some A is a B.` plus `No B is C.` can yield `Some A is not C.`

To conclude one alternative rather than another, state the dichotomy
(`Either we should use X or we should use Y.`), refute one side, and apply
disjunctive syllogism. The dichotomy is a substantive premise to justify.
Likewise, a bridge saying that exactly these premises imply exactly this
conclusion can pass while carrying all the argument's disputed weight. Prefer
to state and instantiate the general rule that actually supports the inference.

### Catalog forms and scope

The catalog includes modus ponens and tollens, hypothetical and disjunctive
syllogism, conjunction introduction and elimination, identity substitution,
biconditionals, and dilemmas. v13 added forms including contraposition, De Morgan,
commutation, exportation/importation, material implication, distribution,
absorption, double-negation introduction, biconditional tollens and chaining,
negated conditional, identity symmetry, quantifier negation, conversion,
universal modus tollens, and possibility introduction. Explosion and excluded
middle from an unrelated premise are intentionally rejected. Conjunction
elimination does not split beneath an outer operator: negating `P and Q` does
not yield `Q`. v14 added the Fitch forms described above; flat-only package
consumers reject scoped input. Use `--forms` for the complete current catalog,
including modal forms.

Parenthesize biconditional sides to fix their boundaries:
`(X) if and only if (Y)`, optionally prefixed by `It is the case that`.
This handles sides containing conditionals. `It holds that` is ordinary text.

Relations use `TERM VERB TERM`, with every/some/no quantifiers in either position.
Fitch generalization and existential elimination also work over relations.
Sentences with two quantifiers are checked under both scope readings; a step
that depends on just one reading fails with a note. Make scope explicit, for
example `Some teacher is such that every student admires that teacher.`
Active and passive clauses do not automatically match.

v22 extends relations to three or more places, including prepositional forms
such as `Alice sends Parcel to Bob from Depot`. For example:

```text
1. Alice gave Bob a book.
2. Bob is a child.
3. Therefore, Alice gave some child a book. [1,2]
```

`Alice gave Bob a book` and `Alice gave a book to Bob` are the same atom:
the double-object / `to` alternation is the only new wording equivalence.
Argument roles, verbs, and prepositions otherwise match verbatim. All scope
orders are checked, with at most four quantifiers per sentence. Nested
`is such that` scaffolds fix scope. See [coverage and known limits](docs/coverage.md)
for the supported vocabulary, examples, and restrictions.

### Relative clauses (v23)

Relative nouns decompose into their class and relation. `F that A admires`
requires both `x is an F` and `A admires x`; `F who admires A` instead requires
`x admires A`. `that`, `who`, and `which` support the same gap rules.

```text
1. Every hero that Sam admires is active.
2. Robin is a hero.
3. Sam admires Robin.
4. Therefore, Robin is active. [1,2,3]
```

The same decomposition supports existential introduction, predicate-nominal
introduction and elimination, and fresh witnesses in Fitch subproofs. Relative
nouns can occur in either relation position under every/some/no, including
`book that Alex gave Blair`. Relations match verbatim with the gap filled.
Use `is such that` to delimit nested scopes. The new grammar has a bounded verb
vocabulary; see [coverage](docs/coverage.md#v23-relative-clauses).

### Explicit variables and notation equivalence (v25)

For nested formulas with no readable phrasing, use `every x is such that BODY`
or `some x is such that BODY`. For example, `every x is such that every y is
such that every z is such that if x outranks y and y outranks z then x outranks z`.
Parentheses fix connective boundaries; variables must be bound without rebinding.
Use `name{x}` or `name{y2}` for constants spelled like variables; bare names
outside explicit bodies keep their previous behavior. The four-quantifier readable limit
does not apply. Indexed variables (`x1`, `x2`, `y12`) give an unbounded
alphabet; round trips test 30–40 simultaneous binders and a 64-binder boundary.
Indexed variables must be bound, just like bare letters, in explicit proofs.

The structured core supports propositional rules, identity substitution, and
Fitch quantifier rules for nested formulas. `Let witness be arbitrary.` opens
an unrestricted fresh-name subproof; the older `Let witness be an arbitrary F.`
remains class-restricted. See [the completeness scope and evidence](docs/coverage.md#completeness-fol-with-identity).
The new **notation equivalence** form cites one premise and requires both
sentences to have one reading, identical up to bound-variable renaming. It
converts readable FOL phrasing to the explicit core and back. Ambiguous readings
are refused; use `is such that` to fix scope. Existing matching forms keep their
priority. This is grammatical conversion, not paraphrase.

Opaque propositional clauses now render verbatim as zero-place atoms, including
inside explicit-variable bodies. Multiword names and predicates and additional
relation prepositions use the readable grammar's lexical recognizers. Distinct
normalized clauses are distinct atoms; this adds no paraphrase matching.

When ordinary phrasing cannot preserve a lexical boundary, use
`atom{"works/with", "alice", x}`: the quoted symbol and ordered arguments specify
one atom, quoted arguments are constants, and unquoted variables must be bound.
`atom{"either it rains"}` has no arguments and is a zero-place proposition.
`name{"claim that the store is nearby"}` quotes a complete name. Quoted strings
use JSON escaping, and decoded names participate in eigenname freshness checks.
These escapes preserve exact symbols and arity; they do not add paraphrase rules.
The renderer uses them only when needed. Curly braces are distinct from citation
brackets. This is still catalog **v25**, with no new inference form.

The legacy audit converts **3,550 of 6,386** distinct sentences both ways with
zero conversion failures. All **116** final lexical/structural residuals are
resolved; the **2,424 ambiguous** and **412 non-FOL** sentences are unchanged.
There are no reclassifications.

The completeness verdict is **Established over the stated scope**: every valid
finite argument with one closed, function-free FOL= reading per sentence under
the gate's controlled grammar has some accepted derivation over nonempty domains.
The argument uses a lossless notation bridge, structural rendering, and the
standard completeness of natural deduction for FOL=. It does not assert that
the checker finds proofs automatically. Fresh variable names avoid collisions
with literal opaque text. Ambiguous readings, modal, counterfactual, property
and that-clause semantics remain outside the claim. Bounded oracle checks are
soundness evidence, not a soundness proof; see the coverage page for the proof
sketch and its implementation-correspondence assumptions.

### Counterfactuals

A conditional whose consequent contains `would` is treated as counterfactual.
Ordinary rules can use it as a premise, but cannot conclude a counterfactual or
use one inside a negation. Counterfactual conclusions require the dedicated
catalog forms, preventing ordinary strengthening, false-antecedent inference,
and unsupported chaining.

Counterfactual modus ponens and tollens compare mood-stripped content:
`had been struck` can match `was struck`. Only auxiliaries are rewritten;
`had gone` and `went` need a visible bridge. `would light` has no plain form,
and `might` has no rules.

### Properties

For copular predicates, `a is F.` and `a has the property of being F.` convert
in either direction, also under a top-level canonical negation. `F` must match
verbatim. Property inclusion (`a has every property that b has.`) permits
instantiation from `b has the property of being F.` The catalog also supports
property existence, shared properties, and indiscernibility from identity.

For property generalization, open a Fitch subproof with
`Let P be an arbitrary property that b has.`, which provides `b has P.`;
derive `a has P.` and discharge to `a has every property that b has.` v21 also
allows property-inclusion chaining for such an arbitrary property. A property
possessed by one individual and denied of another establishes a distinguishing
property, supporting nonidentity or non-inclusion.

Property bearers must be individuals, not quantified phrases or properties
themselves. These conversions do not rewrite inside modals, conditionals, or
counterfactuals. `Being F is a property that a has` remains ordinary text.

## Development

```sh
go build ./...
go vet ./...
go test -vet=off ./...
```

Test fixtures live inside `validation/testdata`; no sibling checkout is needed.
See `LICENSE` for the current rights notice.

### Counting (v26)

Counts one–five (also 1–5) expand into FOL with identity and distinct witnesses.

```text
1. At least two labs signed the pledge.
2. At most one lab that signed the pledge is American.
3. Therefore, some lab that signed the pledge is not American. [1,2]
```

```text
1. Exactly one student admires blair.
| 2. Suppose Ada is a student and Ada admires blair.
| 3. Therefore, no student other than Ada admires blair. [1,2]
4. Therefore, if Ada is a student and Ada admires blair then no student other than Ada admires blair. [2-3]
```

An exact count entails both bounds. Counting definitions convert to and from
explicit-variable FOL. Larger bounds report the bound of 5; general arithmetic
remains unsupported. Uniqueness must establish that
the named witness belongs to the counted class.

### Definite descriptions (v27)

Russellian descriptions assert existence, uniqueness, and predication:

```text
1. The author of Waverley is Scott.
2. The author of Waverley wrote Ivanhoe.
3. Therefore, Scott wrote Ivanhoe. [1,2]
```

The fallback supports substitution, introduction from a witnessed predication
and an at-most-one bound, elimination, and uniqueness inferences. It works
inside Fitch subproofs and converts to explicit FOL with identity. Subject
descriptions can contain literal of-phrases and one relative clause. A bare
capitalized copular complement denotes a name; use `is identical to` for
explicit identity, and lowercase predicate wording for properties. Matching
restrictors stay verbatim; opaque attitude, modal, and deontic clauses never
expose terms for description substitution.

For the present King of France case, `The present king of France is not bald`
has both a wide reading (a unique king who is not bald) and a narrow reading
(not a unique king who is bald). New description steps must hold on every
reading. `It is not the case that the present king of France is bald` fixes
the narrow reading.

Compatibility takes priority: v26 successes retain their exact reports. In
particular, its old narrow-to-surface-negation certificate is preserved even
though it does not establish the new wide reading. The every-reading guarantee
applies to the new fallback, not retroactively to historical certificates.
See [coverage and limits](docs/coverage.md#v27-russellian-definite-descriptions).

### Substitution and CLI provenance (v28)

Schematic conjunctive modus ponens accepts six through ten bare-and conjuncts
and the additional placeholders `U`, `V`, and `W`. Failures report the unmatched
conjunct and current binding. Explicit identity substitution supports whole
gerund terms, coordinated proper names, and `identical with` in controlled
argument positions. Root premises can follow inferences, provided citations
still point backward; Fitch visibility and fresh-name safeguards remain active.

`argcheck --version` prints the catalog and build commit without reading input.
The install script already embeds the commit; module version or `unknown` is
used when build provenance is unavailable. The legacy narrow-negation choice
is unchanged. See [v28 coverage and limits](docs/coverage.md#v28-schematic-substitution-and-premise-placement).
