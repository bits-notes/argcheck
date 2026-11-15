# argcheck

`argcheck` is a deterministic checker in Go for transparently valid numbered
arguments. It runs offline with no LLM and no network access by default: every
conclusion step must instantiate a named inference form from the catalog
(currently **v29.1**, 121 forms) using only its cited earlier steps, with clauses
matched word for word after light surface normalization.

Coverage: complete for propositional logic and for function-free first-order
logic with identity, within a controlled English (sentences with one reading); a
bounded finite-model oracle tests that the checker never accepts an invalid
argument; this is not proved.

Passing `argcheck` shows that an argument's steps follow under the catalog. It
does not establish that the premises are true, that an encoding faithfully
captures a source text, or that an unstated assumption is warranted. If two
wordings are meant to be equivalent, state that link as an explicit premise.

## Install

Requires Go 1.25.0 or later.

Install the latest published version with `go install`:

```sh
go install github.com/bits-notes/argcheck/cmd/argcheck@latest
```

Or install from a local checkout (stamps the binary with the current git commit
into `$HOME/.local/bin`, overridable via `BIN_DIR`):

```sh
./install-argcheck.sh
```

## Usage

```sh
argcheck argument.txt        # check a file
argcheck < argument.txt      # read from stdin
cat argument.txt | argcheck -
argcheck --quiet argument.txt
argcheck --json argument.txt
argcheck --forms             # list all 121 inference forms in catalog v29.1
argcheck --version           # print catalog version and build stamp
```

Put flags before the file argument.

### Exit codes

| Code | Meaning |
| ---: | --- |
| `0` | `VALID` — syntax and every conclusion step pass |
| `1` | `INVALID` — syntax fails or at least one conclusion step fails |
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

Output (`argcheck valid.txt`, exit `0`):

```text
VALID  (gate catalog v29.1; local argcheck@ef44658)
syntax: ok
step 3 [1,2]: ok: modus ponens
```

### Rejected argument

Clauses must match verbatim (after case, whitespace, contraction, and optional
`then` normalization). Replacing `the street is wet` with `the road is wet` is
rejected (`invalid.txt`):

```text
1. If it rains, then the street is wet.
2. It rains.
3. Therefore, the road is wet. [1,2]
```

Output (`argcheck invalid.txt`, exit `1`):

```text
INVALID  (gate catalog v29.1; local argcheck@ef44658)
syntax: ok
step 3 [1,2]: FAIL: nearest form modus ponens: premise 1 consequent "the street is wet" does not match conclusion 3 "the road is wet" verbatim
```

When a conclusion fails only because it cites an extra step that does no work,
`argcheck` prints an advisory `hint:` line naming the form that passes on the
used subset and the inert citation(s) to drop.

## The numbered format

- **Premises:** `1. Statement.` — one declarative sentence per line, ending in a
  period, with no `Therefore,` prefix and no trailing citation brackets.
- **Conclusions:** `3. Therefore, statement. [1,2]` — starts with `Therefore,`,
  ends in a period followed by bracketed citations to earlier steps.
- **Structure:** Steps are numbered sequentially from `1` with no blank lines,
  leading/trailing spaces, code fences, or commentary. The final step must be a
  conclusion. Every premise and every scoped step must support a later step (no
  unused premises). Root-level premises may appear after intermediate conclusions
  as long as every citation points to an earlier step.
- **Canonical phrasing:** Write sentential negation as `It is not the case that X.`,
  disjunction as `Either X or Y.`, conjunction as `X and Y.`, and identity as
  `A is identical to B.` (or `A is identical with B.`).
- **Inline bridge premises:** `If [1,2], then C.` expands steps `1` and `2`
  verbatim into a conjunctive antecedent. Note that an ad hoc bridge puts the
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
  "syntax": {
    "passed": true,
    "message": "output matches numbered_argument_syntax",
    "details": { "step_count": 3 }
  },
  "entailment": {
    "passed": true,
    "message": "output passes entailment_gate",
    "details": {
      "catalog": "v29.1",
      "expansions": [],
      "gated_steps": 1,
      "passed_steps": 1,
      "steps": [
        { "step": 3, "citations": [1, 2], "passed": true, "form": "modus ponens" }
      ]
    }
  },
  "entailment_catalog": "v29.1",
  "backend": "local",
  "version": "argcheck@ef44658"
}
```

`--forms` prints the catalog header (`entailment gate catalog v29.1 (...)`)
followed by all 121 accepted form names, covering propositional schemas,
syllogisms, schematic rules (`U`, `V`, `W`, `X`, `Y`, `Z` up to 10 conjuncts),
identity, system-T modals, counterfactuals, properties, n-place relations,
relative clauses, explicit-variable FOL with notation equivalence, counting
quantifiers (`1`–`5`), and Russellian definite descriptions.

- Every inference form, schema, accepted example, and rejected near-miss:
  [`docs/rules.md`](docs/rules.md)
- Detailed grammar rules, authoring guidance, and known limits:
  [`docs/coverage.md`](docs/coverage.md)
- Agent workflow skill: [`skills/argcheck/SKILL.md`](skills/argcheck/SKILL.md)

## Go library and development

```go
syntax := validation.Validate(ctx,
    validation.BuiltinSpec(validation.BuiltinNumberedArgumentSyntax), argument)
if syntax.Passed {
    report, entailment := validation.CheckEntailmentGate(argument)
    _ = report // per-step forms or mismatches; entailment.Passed is the verdict
}
```

Run tests locally with `go test ./...`.

## Advanced: remote MCP comparison

`--prod` sends the argument over HTTP JSON-RPC (`tools/call`) to a remote MCP
validation endpoint; `--both` runs local and remote checks and exits `3` if they
disagree. Both `ARGCHECK_REMOTE_URL` and `ARGCHECK_REMOTE_TOOL` must be set, either
in the environment or in `$XDG_CONFIG_HOME/argcheck/config` (default
`~/.config/argcheck/config`), one `KEY=VALUE` per line; the environment wins.
There are no built-in defaults; if either is unset, `argcheck` exits `2`.

```text
# ~/.config/argcheck/config
ARGCHECK_REMOTE_URL=https://example.com/mcp
ARGCHECK_REMOTE_TOOL=formalizer.validate
```

## License

[MIT](LICENSE)
