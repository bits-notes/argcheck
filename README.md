# english-arg-validator

`english-arg-validator` (formerly `argcheck`) is a deterministic checker in Go for transparently valid numbered
arguments. It runs offline with no LLM and no network access by default: every
conclusion step must instantiate a named inference form from the catalog
(currently **v41.0**, 121 forms) using only its cited earlier steps, with clauses
matched word for word after light surface normalization.

Coverage: complete for propositional logic and for function-free first-order
logic with identity, within a controlled English (sentences with one reading); a
bounded finite-model oracle tests that the checker never accepts an invalid
argument; this is not proved.

Passing `english-arg-validator` shows that an argument's steps follow under the catalog. It
does not establish that the premises are true, that an encoding faithfully
captures a source text, or that an unstated assumption is warranted. If two
wordings are meant to be equivalent, state that link as an explicit premise.

## Example

An eleven-step argument about an outage. Each "Therefore" line cites the earlier steps it uses.

```text
1. Either the server crashed or the network failed.
2. If the network failed, then the checkout request timed out.
3. It is not the case that the checkout request timed out.
4. Therefore, it is not the case that the network failed. [2,3]
5. Therefore, the server crashed. [1,4]
6. If the server crashed, then every on-call engineer was paged.
7. Therefore, every on-call engineer was paged. [5,6]
8. Alice is an on-call engineer.
9. Therefore, Alice was paged. [7,8]
10. If Alice was paged and the server crashed, then Alice owns the incident report.
11. Therefore, Alice owns the incident report. [5,9,10]
```

`english-arg-validator incident.txt` certifies every derived step and names the rule it matched:

```text
VALID  (gate catalog v41.0; local english-arg-validator@f28a796)
syntax: ok
step 4 [2,3]: ok: modus tollens
step 5 [1,4]: ok: disjunctive syllogism
step 7 [5,6]: ok: modus ponens
step 9 [7,8]: ok: non-copular universal instantiation
step 11 [5,9,10]: ok: conjunctive modus ponens
```

One natural slip: write premise 3 as "The checkout request did not time out." English readers see the same claim, but the checker matches clauses word for word, so step 4 is no longer certified:

```text
NOT CERTIFIED  (gate catalog v41.0; local english-arg-validator@f28a796)
syntax: ok
step 4 [2,3]: FAIL: nearest form modus ponens: premise 2 antecedent "the network failed" does not match premise 3 "the checkout request did not time out" verbatim
step 5 [1,4]: ok: disjunctive syllogism
step 7 [5,6]: ok: modus ponens
step 9 [7,8]: ok: non-copular universal instantiation
step 11 [5,9,10]: ok: conjunctive modus ponens
```

The fix is to word premise 3 the way premise 2 words its clause: "It is not the case that the checkout request timed out." A step that is not certified is not shown to be invalid. It means no catalog rule matched the wording as written.

## Install

Requires Go 1.25.0 or later.

Install the latest published version with `go install`:

```sh
go install github.com/bits-notes/english-arg-validator/cmd/english-arg-validator@latest
```

Or install from a local checkout (stamps the binary with the current git commit
into `$HOME/.local/bin`, overridable via `BIN_DIR`):

```sh
./install.sh
```

## Usage

```sh
english-arg-validator argument.txt        # check a file
english-arg-validator < argument.txt      # read from stdin
cat argument.txt | english-arg-validator -
english-arg-validator --quiet argument.txt
english-arg-validator --json argument.txt
english-arg-validator --forms             # list all 121 inference forms in catalog v41.0
english-arg-validator --version           # print catalog version and build stamp
```

Put flags before the file argument.

### Exit codes

| Code | Meaning |
| ---: | --- |
| `0` | `VALID` — syntax and every conclusion step pass |
| `1` | `NOT CERTIFIED` — syntax fails or at least one conclusion step matches no accepted rule. This does not mean the argument is invalid, only that english-arg-validator could not certify it. |
| `2` | Input, flag, or operational error (for example, empty input or missing remote configuration) |
| `3` | `DIVERGENT` — local and remote verdicts disagree (`--both`) |

## Worked examples

### Valid argument

Input (`valid.txt`):

```text
1. If it rains, then the street is wet.
2. It rains.
3. Therefore, the street is wet. [1,2]
```

Output (`english-arg-validator valid.txt`, exit `0`):

```text
VALID  (gate catalog v41.0; local english-arg-validator@f28a796)
syntax: ok
step 3 [1,2]: ok: modus ponens
```

### Rejected argument

Clauses must match verbatim (after case, whitespace, contraction (`isn't`, `it's not the case
that`), and optional
`then` normalization). Replacing `the street is wet` with `the road is wet` is
not certified (`uncertified.txt`):

```text
1. If it rains, then the street is wet.
2. It rains.
3. Therefore, the road is wet. [1,2]
```

Output (`english-arg-validator uncertified.txt`, exit `1`):

```text
NOT CERTIFIED  (gate catalog v41.0; local english-arg-validator@f28a796)
syntax: ok
step 3 [1,2]: FAIL: nearest form modus ponens: premise 1 consequent "the street is wet" does not match conclusion 3 "the road is wet" verbatim
```

When a conclusion fails only because it cites an extra step that does no work,
`english-arg-validator` prints an advisory `hint:` line naming the form that passes on the
used subset and the inert citation(s) to drop.

## The numbered format

- **Premises:** `1. Statement.` — one declarative sentence per line, ending in a
  period, with no inference marker and no trailing citation brackets.
- **Conclusions:** `3. Therefore, statement. [1,2]` — ends in a period followed by
  bracketed citations to earlier steps. The leading `Therefore`, `So`, `Thus`,
  or `Hence` is optional, case-insensitive, and accepts an optional comma.
  A marker without terminal citations is a syntax error. Inline references
  (`If [1,2], then P.` or `P if [1].`) remain premises. Re-rendered steps use
  canonical `Therefore,`.
- **Structure:** Steps are numbered sequentially from `1` with no blank lines,
  leading/trailing spaces, code fences, or commentary. The final step must be a
  conclusion. Every premise and every scoped step must support a later step (no
  unused premises). Root-level premises may appear after intermediate conclusions
  as long as every citation points to an earlier step.
- **Canonical phrasing:** Write sentential negation as `It is not the case that X.`,
  disjunction as `Either X or Y.`, conjunction as `X and Y.`, and identity as
  `A is identical to B.` (or `A is identical with B.`).
- **Conditionals:** `If X, then Y.` (also `If X, Y.`). The postposed form
  `Y if X.` (or `Y, if X.`) reads the same. `only if`, `even if`, `as if`,
  embedded questions (`asked if`), reports (`says … if`) and a consequent containing `and`/`or` are not
  read as postposed conditionals.
- **Inline bridge premises:** `If [1], then C.`, `If [1,2], then C.` or
  `If A, then [1].` expands the referenced steps verbatim. References may fill
  a whole antecedent or consequent and nowhere else; see "Step references" in `docs/rules.md`. An ad hoc bridge puts the
  inferential weight into that conditional premise rather than proving it.

### Fitch subproofs

Prefix each active subproof level with `| `. Open an assumption with
`Suppose X.` (or `Let a be an arbitrary F.` / `Let x be arbitrary.` for
universal generalization) and discharge the closed subproof by citing its full
line range `[i-j]`:

```text
1. If it rains, then the street is wet.
2. If the street is wet, then the cars slow.
| 3. Suppose it rains.
| 4. Therefore, the street is wet. [1,3]
| 5. Therefore, the cars slow. [2,4]
6. Therefore, if it rains, then the cars slow. [3-5]
```

Steps outside a closed subproof cannot cite individual lines inside it. Reductio
subproofs conclude `Therefore, contradiction. [i,j]`.

## `--json` and `--forms`

`--json` emits structured syntax and entailment results:

```json
{
  "valid": true,
  "verdict": "VALID",
  "syntax": {
    "passed": true,
    "message": "output matches numbered_argument_syntax",
    "details": { "step_count": 3 }
  },
  "entailment": {
    "passed": true,
    "message": "output passes entailment_gate",
    "details": {
      "catalog": "v41.0",
      "expansions": [],
      "gated_steps": 1,
      "passed_steps": 1,
      "steps": [
        { "step": 3, "citations": [1, 2], "passed": true, "form": "modus ponens" }
      ]
    }
  },
  "entailment_catalog": "v41.0",
  "backend": "local",
  "version": "english-arg-validator@ef44658"
}
```

`verdict` is `"VALID"` or `"NOT CERTIFIED"`; the boolean `valid` is kept for
existing clients. A failing check means english-arg-validator found no accepted rule for some
step, not that the argument is invalid.

`--forms` prints the catalog header (`entailment gate catalog v41.0 (...)`)
followed by all 121 accepted form names, covering propositional schemas,
syllogisms, schematic rules (`U`, `V`, `W`, `X`, `Y`, `Z` up to 10 conjuncts),
identity, system-T modals, counterfactuals, properties, n-place relations,
relative clauses, explicit-variable FOL with notation equivalence, counting
quantifiers (`1`–`5`), and Russellian definite descriptions.

- Every inference form, schema, accepted example, and rejected near-miss:
  [`docs/rules.md`](docs/rules.md)
- Detailed grammar rules, authoring guidance, and known limits:
  [`docs/coverage.md`](docs/coverage.md)
- Agent workflow skill: [`skills/english-arg-validator/SKILL.md`](skills/english-arg-validator/SKILL.md)
- Formalizing from a source text (sentence ledger, supplied premises, fidelity
  pass): [`skills/formalize-from-source/SKILL.md`](skills/formalize-from-source/SKILL.md)

## Go library and development

The library contains the deterministic entailment checker and numbered-argument
parser and syntax validator. LLM output contracts and reconstruction, repair,
canonicalization, and disputation workflows belong in their consuming application.

```go
syntax := validation.Validate(ctx,
    validation.BuiltinSpec(validation.BuiltinNumberedArgumentSyntax), argument)
if syntax.Passed {
    report, entailment := validation.CheckEntailmentGate(argument)
    _ = report // per-step forms or mismatches; entailment.Passed is the verdict
}
```

Run tests locally with `go test -timeout 30m ./...`. The full suite takes about 6 to 10 minutes, mostly in the truth-table and Fitch oracle tests in `validation/`, which is close to Go's default 10-minute per-package timeout.

## Advanced: remote MCP comparison

`--prod` sends the argument over HTTP JSON-RPC (`tools/call`) to a remote MCP
validation endpoint; `--both` runs local and remote checks and exits `3` if they
disagree. Both `EAV_REMOTE_URL` and `EAV_REMOTE_TOOL` must be set, either
in the environment or in `$XDG_CONFIG_HOME/english-arg-validator/config` (default
`~/.config/english-arg-validator/config`), one `KEY=VALUE` per line; the environment wins.
There are no built-in defaults; if either is unset, `english-arg-validator` exits `2`.

```text
# ~/.config/english-arg-validator/config
EAV_REMOTE_URL=https://example.com/mcp
EAV_REMOTE_TOOL=formalizer.validate
```

## License

[MIT](LICENSE)
