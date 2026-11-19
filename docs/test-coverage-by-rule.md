# Test coverage by inference form

Positive is the number of gated steps the validation tests accepted with that form. Negative is the number of gated steps they rejected with that form as `nearest_form`.

Catalog version: `v44.0`.

The table lists every form `english-arg-validator --forms` prints (`validation.EntailmentGateForms`), in that order. Regenerate it with the commands below.

## Regenerate

```
ARGCHECK_FORM_COVERAGE=$PWD/tmp/form-coverage.json go test ./validation/...
go run ./tools/formcoverage tmp/form-coverage.json > docs/test-coverage-by-rule.md
```

| Form | Positive | Negative |
| --- | ---: | ---: |
| `schematic conjunctive modus ponens` | 41 | 24 |
| `conjunctive modus ponens` | 111 | 384 |
| `modus ponens` | 381 | 2162 |
| `modus tollens` | 83 | 57 |
| `hypothetical syllogism` | 98 | 13 |
| `biconditional introduction` | 209 | 7 |
| `biconditional elimination` | 362 | 1547 |
| `biconditional modus ponens` | 733 | 1536 |
| `conjunction introduction` | 84 | 67 |
| `n-ary conjunction introduction` | 14 | 0 |
| `conjunction elimination` | 7169 | 1078 |
| `disjunction introduction` | 85 | 1956 |
| `disjunctive syllogism` | 83 | 251 |
| `double negation` | 30 | 48 |
| `self-entailment` | 5788 | 8205 |
| `identity transitivity` | 6 | 5 |
| `universal syllogism` | 4 | 4 |
| `universal negative syllogism` | 5 | 4 |
| `universal negative instantiation` | 11 | 14 |
| `schematic instantiation` | 43 | 17 |
| `non-copular universal instantiation` | 128 | 141 |
| `universal instantiation` | 27 | 16 |
| `existential generalization` | 18 | 0 |
| `existential introduction` | 15 | 13 |
| `existential counterexample` | 7 | 4 |
| `necessity elimination` | 5 | 0 |
| `existential syllogism` | 9 | 3 |
| `existential negative syllogism` | 9 | 1 |
| `schematic existential introduction` | 14 | 0 |
| `identity substitution` | 63 | 34 |
| `constructive dilemma` | 133 | 1721 |
| `simple dilemma` | 181 | 0 |
| `destructive dilemma` | 44 | 0 |
| `contraposition` | 45 | 4 |
| `De Morgan` | 65 | 1 |
| `commutation` | 78 | 38 |
| `exportation` | 40 | 10 |
| `importation` | 40 | 7 |
| `material implication` | 72 | 30 |
| `distribution` | 137 | 1117 |
| `absorption` | 60 | 298 |
| `double negation introduction` | 27 | 0 |
| `biconditional modus tollens` | 471 | 3648 |
| `biconditional chaining` | 284 | 310 |
| `biconditional commutation` | 205 | 51 |
| `negated conditional` | 93 | 11 |
| `identity symmetry` | 4 | 3 |
| `quantifier negation` | 7 | 2 |
| `conversion` | 9 | 2 |
| `universal modus tollens` | 6 | 4 |
| `possibility introduction` | 6 | 0 |
| `contradiction introduction` | 12500 | 0 |
| `identity reflexivity` | 50 | 0 |
| `modal duality` | 47 | 0 |
| `modal distribution` | 53 | 3 |
| `relational universal instantiation` | 3 | 0 |
| `relational existential generalization` | 98 | 0 |
| `relational universal negative instantiation` | 39 | 0 |
| `relational scope conversion` | 221 | 310 |
| `counterfactual modus ponens` | 39 | 0 |
| `counterfactual modus tollens` | 19 | 0 |
| `counterfactual to material` | 39 | 0 |
| `counterfactual consequent conjunction` | 31 | 0 |
| `counterfactual consequent conjunction elimination` | 43 | 0 |
| `counterfactual consequent disjunction introduction` | 41 | 0 |
| `counterfactual dilemma` | 23 | 0 |
| `counterfactual reflexivity` | 21 | 0 |
| `conditional proof` | 642 | 6 |
| `reductio` | 1644 | 0 |
| `proof by cases` | 7 | 2 |
| `universal generalization` | 7 | 46 |
| `existential elimination` | 27 | 48 |
| `relational universal generalization` | 21 | 0 |
| `relational existential elimination` | 10 | 0 |
| `property abstraction` | 150 | 0 |
| `property concretion` | 145 | 0 |
| `property instantiation` | 31 | 0 |
| `property existential` | 30 | 0 |
| `shared property` | 21 | 0 |
| `indiscernibility of identicals` | 21 | 0 |
| `property generalization` | 23 | 11 |
| `property instantiation (arbitrary property)` | 21 | 0 |
| `distinguishing property` | 40 | 0 |
| `distinguishing property nonidentity` | 34 | 0 |
| `distinguishing property non-inclusion` | 30 | 0 |
| `relative clause inference` | 27 | 1 |
| `FOL reiteration` | 622 | 0 |
| `FOL conjunction introduction` | 52 | 0 |
| `FOL conjunction elimination` | 57 | 0 |
| `FOL disjunction introduction` | 51 | 0 |
| `FOL modus ponens` | 206 | 0 |
| `FOL contradiction introduction` | 103 | 0 |
| `FOL contradiction elimination` | 51 | 0 |
| `FOL double negation` | 48 | 0 |
| `FOL double negation introduction` | 1 | 0 |
| `FOL biconditional introduction` | 51 | 0 |
| `FOL biconditional elimination` | 51 | 0 |
| `FOL universal instantiation` | 340 | 0 |
| `FOL existential introduction` | 56 | 0 |
| `FOL identity reflexivity` | 37 | 0 |
| `FOL identity substitution` | 32 | 0 |
| `FOL conditional proof` | 51 | 0 |
| `FOL reductio` | 51 | 0 |
| `FOL universal generalization` | 52 | 0 |
| `FOL existential elimination` | 52 | 0 |
| `FOL proof by cases` | 51 | 0 |
| `notation equivalence` | 5813 | 13 |
| `counting definition` | 73 | 30 |
| `counting introduction` | 1 | 0 |
| `counting elimination` | 1 | 0 |
| `counting contradiction` | 1 | 0 |
| `counting uniqueness` | 8 | 0 |
| `counting subset` | 4 | 0 |
| `description definition` | 8 | 0 |
| `description substitution` | 6 | 0 |
| `description introduction` | 4 | 0 |
| `description elimination` | 3 | 0 |
| `description uniqueness` | 2 | 0 |
| `description negation` | 2 | 0 |
| `unless elimination` | 4 | 1 |
| `unless conversion` | 3 | 0 |

## Summary

0 forms have 0 positive. 67 forms have 0 negative.

## Forms counted outside `--forms`

A name in this list was reported for a gated step and is not printed by `english-arg-validator --forms`. Reported names can differ from catalog names by a suffix such as ` (necessitation)`.

- `Fitch discharge`: positive 0, negative 13810
- `antecedent strengthening`: positive 0, negative 2
- `counterfactual content`: positive 0, negative 125
- `counterfactual rules`: positive 0, negative 357
- `explicit-variable FOL`: positive 0, negative 44
- `material to counterfactual`: positive 0, negative 4
- `modus ponens (necessitation)`: positive 2, negative 0
- `none`: positive 0, negative 480
- `relational grammar`: positive 0, negative 4
- `scope`: positive 0, negative 1
- `simplification of disjunctive antecedents`: positive 0, negative 5
- `strong centering`: positive 0, negative 3
- `universal negative syllogism (necessitation)`: positive 4, negative 0

### v31 modal atom regression

`entailment_gate_v31_test.go` covers indicative modal atoms in modus ponens,
modus tollens, conjunctive modus ponens and schematic instantiation, plus eight representative
real arguments in `testdata/v31/`. Negative controls reject converse,
modal introduction/elimination/substitution and counterfactual might detachment.
System-T controls and the explicit v27 empty-restrictor countermodel pin the
unchanged boundaries. The modus ponens entry in `docs/rules.md` includes
executable accepted/rejected modal examples. Catalog v31 still has 121 forms.

## Catalog v42.1 strictness notes

`TestStrictNotesExamples` covers all thirteen rule IDs, unchanged baseline
verdicts and mismatch text, and certifying fixes. `TestDocsRulesStrictNotes`
keeps the generated [strictness table](rules.md#strictness-notes) synchronized
with the Go catalog.
