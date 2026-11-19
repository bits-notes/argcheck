# Inference Rule Catalog (`v44.0`)

`english-arg-validator` checks each conclusion step against the 121 inference forms in gate catalog `v44.0` (`english-arg-validator --forms`). Clauses are matched verbatim after normalizing case, whitespace, contractions (including `it's (not) the case that` for `it is (not) the case that`), and edge punctuation; synonyms, paraphrases, and unstated domain assumptions are rejected.

The checker is complete for propositional logic and for function-free first-order logic with identity, within a controlled English (sentences with one reading); a bounded finite-model oracle tests that the checker never accepts an invalid argument; this is not proved.

Every `argcheck-accept <form>` and `argcheck-reject <form>` fenced block in this document is executed by `validation/docs_rules_test.go`.

## Numbered format

Premises have the form `1. Statement.`. Derived steps end with a period and
terminal citations: `3. Statement. [1,2]`. An optional leading `Therefore`,
`So`, `Thus`, or `Hence` is case-insensitive and may have a comma. A marker
without valid terminal citations is a syntax error. Markers match whole words;
ordinary words beginning with the same letters are unaffected. Degree phrases
such as `So many` without a comma remain statement content.

Only terminal citations identify a derived step. Inline references such as
`If [1,2], then P.` and `P if [1].` remain bridge premises. Re-rendered derived
steps use canonical `Therefore,`.

```argcheck-accept modus ponens
1. If it rains, then the ground is wet.
2. It rains.
3. So, the ground is wet. [1,2]
```

```argcheck-accept modus ponens
1. If it rains, then the ground is wet.
2. It rains.
3. The ground is wet. [1,2]
```

```argcheck-reject modus ponens
1. If it rains, then the ground is wet.
2. It rains.
3. So, the ground is wet.
```

## Step references

A premise may name earlier steps by number in one place only: a bridge premise whose whole antecedent, whole consequent, or both are step references.

- Syntax: `If [1], then C.`, `If [1] and [2], then C.`, `If [1], and [2], then C.` or `If [1,2], then C.` for the antecedent; `If A, then [1].` or `If A, then [1,2].` for the consequent; `If [1], then [2].` for both. A comma before `then` or `and` is optional.
- Expansion: each reference becomes `it is the case that <the step's text>`, copied verbatim, and several references are joined by `, and`, on either side of the conditional. So `If [1,2], then C.` is checked as `If it is the case that S1, and it is the case that S2, then C.`, and the step that uses it must still cite steps 1 and 2.
- What can be referenced: any earlier step at depth 0 or in a still-open enclosing subproof, including intermediate conclusions and discharged conditionals. A step inside a closed subproof cannot be referenced, and neither can a later step.
- Where references are rejected (syntax error "inline bracket references outside an explicit bridge premise"): under negation (`It is not the case that [1].`), in a disjunction (`Either [1] or [2].`), and mixed with other text on the same side (`If [1] and the sky is grey, then C.`, `If A, then [1] and the sky is grey.`). A bridge is a premise, so it cannot appear inside a subproof.
- Parenthesized numbers such as `(1)` are ordinary text. When a failing step uses one that names an earlier step, the failure message adds a hint: `to refer to step 1, write [1]`.
- The bridge caveat: a bridge premise asserts that the referenced steps yield C. The gate checks that the step follows from the bridge. It does not check that the bridge itself is true, so an ad hoc bridge carries the step's weight.

```argcheck-accept modus ponens
1. The street is wet.
2. If [1], then it rained.
3. Therefore, it rained. [1,2]
```

```argcheck-accept conjunctive modus ponens
1. The street is wet.
2. The sky is grey.
3. If [1,2], then it rained.
4. Therefore, it rained. [1,2,3]
```

A consequent reference lets the bridge feed modus tollens:

```argcheck-accept modus tollens
1. The street is wet.
2. If it rained, then [1].
3. It is not the case that the street is wet.
4. Therefore, it is not the case that it rained. [2,3]
```

## `schematic conjunctive modus ponens`

- **Schema:** `If Φ(U..Z) and ... and Ψ(U..Z), then Ω(U..Z)`; `Φ(t_U..t_Z)`; ...; `Ψ(t_U..t_Z)` ⊢ `Ω(t_U..t_Z)` (up to 10 antecedent conjuncts under one consistent substitution for uppercase placeholders `U`, `V`, `W`, `X`, `Y`, `Z`).
- **Valid form, refused by v44 strict grammar (`slot.value_singular`):**

```argcheck-refused slot.value_singular | Name the thing | schematic conjunctive modus ponens
1. If X is Y's mother and Z is X's brother, then Z is Y's uncle.
2. Dana is Pat's mother.
3. Lee is Dana's brother.
4. Therefore, Lee is Pat's uncle. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject schematic conjunctive modus ponens
1. If X is Y's mother and Z is X's brother, then Z is Y's uncle.
2. Dana is Pat's mother.
3. Lee is Pat's brother.
4. Therefore, Lee is Pat's uncle. [1,2,3]
```

- **What it does not cover:** Inconsistent bindings across conjuncts, unbound placeholders in the consequent, lowercase letters (`x`, `y`) used as schematic placeholders, more than 10 unscaffolded `and`-conjuncts, or extra cited premises that do not match an antecedent conjunct.

---

## `conjunctive modus ponens`

- **Schema:** `If P_1 and ... and P_n, then Q`; `P_1`; ...; `P_n` ⊢ `Q`.
- **Valid form, refused by v44 strict grammar (`scope.clause_start`):**

```argcheck-refused scope.clause_start | move the operator into each clause | conjunctive modus ponens
1. Rain falls.
2. Wind blows.
3. Snow melts.
4. If rain falls and wind blows and snow melts, then rivers rise.
5. Therefore, rivers rise. [1,2,3,4]
```

- **Rejected near-miss:**

```argcheck-reject conjunctive modus ponens
1. Rain falls.
2. Wind blows.
3. If rain falls and wind blows and snow melts, then rivers rise.
4. Therefore, rivers rise. [1,2,3]
```

- **What it does not cover:** Unestablished antecedent conjuncts, affirming the consequent, wording drift between a cited premise and its antecedent conjunct, or more than 12 plain `and` boundaries in the conditional antecedent.

---

## `modus ponens`

- **Schema:** `If P, then Q`; `P` ⊢ `Q` (also accepts `When P, Q`, `Whenever P, Q`, and `Q if P`).
- **Accepted:**

```argcheck-accept modus ponens
1. If rain falls, then streets get wet.
2. Rain falls.
3. Therefore, streets get wet. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject modus ponens
1. If rain falls, then streets get wet.
2. Rain falls.
3. Therefore, streets get damp. [1,2]
```

- **Postposed conditionals (v31.2):** `Y if X.` and `Y, if X.` read exactly as `If X, then Y.` wherever a conditional premise is accepted (modus ponens, modus tollens, hypothetical syllogism, conjunctive antecedents `Y if X and Z`, and step references `C if [1]`). The rewrite is skipped for `only if`, `even if`, `as if`, `if and only if`, any `unless`, embedded questions (including intervening objects/adverbs, progressive or perfect verbs, and preposition-final complements: `asks the robot if`, `wonders aloud if`, `is checking if`, `has forgotten if`, `is curious about if`), a consequent that reports or attributes (`says`, `thinks`, `that`, ...), a consequent containing `and`/`or`/`either` or a comma, an antecedent containing `or`, and statements with more than one `if`. The rewrite applies to whole numbered statements, not to clauses inside another statement, and only when the argument as written fails, so it never turns a passing argument into a failing one.

```argcheck-accept modus ponens
1. Steve will wear a raincoat tomorrow if it rains.
2. It rains.
3. Therefore, Steve will wear a raincoat tomorrow. [1,2]
```

- **What it does not cover:** Synonym substitution (`wet` to `damp`), affirming the consequent, `"only if"` treated as `"if"`, causal/temporal connectives (`because`, `since`, `after`, `unless`), or counterfactual detachment to indicative mood (see `counterfactual modus ponens`).


Indicative modal phrases are verbatim atoms in v31:

```argcheck-accept modus ponens
1. If the system is misaligned, then the system might seek power.
2. The system is misaligned.
3. Therefore, the system might seek power. [1,2]
```

```argcheck-reject modus ponens
1. If the system is misaligned, then the system might seek power.
2. The system is misaligned.
3. Therefore, the system seeks power. [1,2]
```

The modal cannot be dropped, introduced, or replaced with another modal. This also applies to modus tollens, conjunctive modus ponens, and schematic forms.

---

## `modus tollens`

- **Schema:** `If P, then Q`; `It is not the case that Q` ⊢ `It is not the case that P`.
- **Accepted:**

```argcheck-accept modus tollens
1. If rain falls, then streets get wet.
2. It is not the case that streets get wet.
3. Therefore, it is not the case that rain falls. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject modus tollens
1. If rain falls, then streets get wet.
2. It is not the case that rain falls.
3. Therefore, it is not the case that streets get wet. [1,2]
```

- **What it does not cover:** Denying the antecedent, lexical antonyms in place of explicit negation, or indicative negation applied to a counterfactual conditional (see `counterfactual modus tollens`).

---

## `hypothetical syllogism`

- **Schema:** `If P, then Q`; `If Q, then R` ⊢ `If P, then R`.
- **Accepted:**

```argcheck-accept hypothetical syllogism
1. If rain falls, then streets get wet.
2. If streets get wet, then traffic slows.
3. Therefore, if rain falls, then traffic slows. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject hypothetical syllogism
1. If rain falls, then streets get wet.
2. If streets get wet, then traffic slows.
3. Therefore, if rain falls, then traffic stops. [1,2]
```

- **What it does not cover:** Non-verbatim middle clauses, wording drift in the antecedent or consequent, or chaining counterfactual `would` conditionals (which is invalid for counterfactuals).

---

## `biconditional introduction`

- **Schema:** `If P, then Q`; `If Q, then P` ⊢ `P if and only if Q`.
- **Accepted:**

```argcheck-accept biconditional introduction
1. If the switch is up, then the lamp is lit.
2. If the lamp is lit, then the switch is up.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject biconditional introduction
1. If the switch is up, then the lamp is lit.
2. If it is not the case that the lamp is lit, then it is not the case that the switch is up.
3. Therefore, the switch is up if and only if the lamp is lit. [1,2]
```

- **What it does not cover:** Supplying a contrapositive in place of the reverse conditional without first contraposing it, `"only if"` conclusions, or unparenthesized compound sides around `if and only if`.

---

## `biconditional elimination`

- **Schema:** `P if and only if Q` ⊢ `If P, then Q` (or `If Q, then P`).
- **Accepted:**

```argcheck-accept biconditional elimination
1. The switch is up if and only if the lamp is lit.
2. Therefore, if the switch is up, then the lamp is lit. [1]
```

- **Rejected near-miss:**

```argcheck-reject biconditional elimination
1. The switch is up only if the lamp is lit.
2. Therefore, if the switch is up, then the lamp is lit. [1]
```

- **What it does not cover:** `"only if"` premises, biconditionals embedded inside attitude or attribution prefixes (`According to the theorem, ...`), or counterfactual mood inside the biconditional sides.

---

## `biconditional modus ponens`

- **Schema:** `P if and only if Q`; `P` ⊢ `Q` (and symmetrically `P if and only if Q`; `Q` ⊢ `P`).
- **Accepted:**

```argcheck-accept biconditional modus ponens
1. The switch is up if and only if the lamp is lit.
2. The lamp is lit.
3. Therefore, the switch is up. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject biconditional modus ponens
1. The switch is up if and only if the lamp is lit.
2. The switch is raised.
3. Therefore, the lamp is lit. [1,2]
```

- **What it does not cover:** Wording drift (`up` vs `raised`), `"only if"` premises, embedded biconditionals, or negated premises (use `biconditional modus tollens`).

---

## `conjunction introduction`

- **Schema:** `P`; `Q` ⊢ `P and Q` (or `It is the case both that P, and that Q`).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`):**

```argcheck-refused scope.clause_start | move the operator into each clause | conjunction introduction
1. Rain falls.
2. Wind blows.
3. Therefore, rain falls and wind blows. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject conjunction introduction
1. Rain falls.
2. Wind blows.
3. Therefore, rain falls and snow falls. [1,2]
```

- **What it does not cover:** Combining three or more premises in one step (handled by `n-ary conjunction introduction`), reworded conjuncts, or uncited premises.

---

## `n-ary conjunction introduction`

- **Schema:** `P_1`; `P_2`; ...; `P_n` ⊢ `P_1, P_2, ..., and P_n` (or `P_1 and P_2 and ... and P_n`, for `3 <= n <= 6`).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`):**

```argcheck-refused scope.clause_start | move the operator into each clause | n-ary conjunction introduction
1. Rain falls.
2. Wind blows.
3. Snow melts.
4. Therefore, rain falls, wind blows, and snow melts. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject n-ary conjunction introduction
1. Rain falls.
2. Wind blows.
3. Snow melts.
4. Therefore, rain falls, wind blows, and the snow melts. [1,2,3]
```

- **What it does not cover:** More than 6 premises in one step (chain binary or smaller conjunction steps instead), dropped or added conjuncts, or article/wording drift.

---

## `conjunction elimination`

- **Schema:** `P and Q` ⊢ `P` (or `Q`).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`):**

```argcheck-refused scope.clause_start | move the operator into each clause | conjunction elimination
1. Rain falls and wind blows.
2. Therefore, wind blows. [1]
```

- **Rejected near-miss:**

```argcheck-reject conjunction elimination
1. Rain falls and wind blows.
2. Therefore, snow falls. [1]
```

- **What it does not cover:** Conjunctions buried under negation, conditionals, or attitude verbs, or extracting a noun phrase from a coordinated subject (`Ada and Ben met`).

---

## `disjunction introduction`

- **Schema:** `P` ⊢ `Either P or Q` (or `P or Q`, in either disjunct position).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`):**

```argcheck-refused scope.clause_start | move the operator into each clause | disjunction introduction
1. Rain falls.
2. Therefore, either rain falls or snow falls. [1]
```

- **Rejected near-miss:**

```argcheck-reject disjunction introduction
1. Rain falls.
2. Therefore, either wind blows or snow falls. [1]
```

- **What it does not cover:** Disjunctions where neither disjunct matches the cited premise verbatim, or introducing disjunctions inside opaque or modal scopes without a modal rule.

---

## `disjunctive syllogism`

- **Schema:** `Either P or Q` (or `P or Q`); `It is not the case that P` ⊢ `Q` (and symmetrically when `Q` is negated).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`):**

```argcheck-refused scope.clause_start | move the operator into each clause | disjunctive syllogism
1. Rain falls or snow falls.
2. It is not the case that rain falls.
3. Therefore, snow falls. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject disjunctive syllogism
1. Rain falls or snow falls.
2. It is not the case that rain falls.
3. Therefore, hail falls. [1,2]
```

- **What it does not cover:** Exclusive-or inferences (affirming `P` to conclude `not Q`), three-way disjunction elimination in a single step without intermediate grouping, or elided disjunctions with quantified or ambiguous subjects.

---

## `double negation`

- **Schema:** `It is not the case that it is not the case that P` ⊢ `P`.
- **Accepted:**

```argcheck-accept double negation
1. It is not the case that it is not the case that rain falls.
2. Therefore, rain falls. [1]
```

- **Rejected near-miss:**

```argcheck-reject double negation
1. It is not the case that it is not the case that rain falls.
2. Therefore, snow falls. [1]
```

- **What it does not cover:** Lexical double negatives (`not uncommon` to `common`), or stripping double negation inside a subclause rather than at the outer sentence level.

---

## `self-entailment`

- **Schema:** `P` ⊢ `P` (reiteration up to case, whitespace, contraction, wrapper `it is the case that`, and supported conditional/elision surface variants).
- **Accepted:**

```argcheck-accept self-entailment
1. Rain falls.
2. Therefore, rain falls. [1]
```

- **Rejected near-miss:**

```argcheck-reject self-entailment
1. Rain falls.
2. Therefore, snow falls. [1]
```

- **What it does not cover:** Paraphrase, active/passive rewrites, or `"it holds that"` treated as a transparent wrapper.

---

## `identity transitivity`

- **Schema:** `a is identical to b`; `b is identical to c` ⊢ `a is identical to c` (also accepts `is identical with`).
- **Valid form, refused by v44 strict grammar (`identity.constant_sides`):**

```argcheck-refused identity.constant_sides | name the thing | identity transitivity
1. Alpha is identical to beta.
2. Beta is identical to gamma.
3. Therefore, alpha is identical to gamma. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject identity transitivity
1. Alpha is identical to beta.
2. Beta is identical to gamma.
3. Therefore, alpha is identical to delta. [1,2]
```

- **What it does not cover:** Bare copular `"is"` (`Alpha is beta`), negated identities, or chaining three or more identity premises in a single step, or quantifier phrases as terms: since v28.2 `Tom is identical to some cat` and `Some cat is identical to Felix` do not chain, and neither do `a`, `no`, `every` or `nobody` terms.

---

## `universal syllogism`

- **Schema:** `Every A is B` (or `All As are Bs`); `Every B is C` ⊢ `Every A is C` (Barbara).
- **Valid form, refused by v44 strict grammar (`term.inflection_verbatim`):**

```argcheck-refused term.inflection_verbatim | singular 'every N' | universal syllogism
1. All humans are mammals.
2. Every mammal is mortal.
3. Therefore, all humans are mortal. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject universal syllogism
1. All artists are creative.
2. All scientists are creative.
3. Therefore, all scientists are artists. [1,2]
```

- **What it does not cover:** Undistributed middle (`All A are C; All B are C`), existential import from universals (`Every A is B` to `Some A is B`), or non-copular verb predicates.

---

## `universal negative syllogism`

- **Schema:** `Every A is B`; `No B is C` (or `No C is B`) ⊢ `No A is C` (Celarent / Cesare).
- **Valid form, refused by v44 strict grammar (`term.inflection_verbatim`):**

```argcheck-refused term.inflection_verbatim | singular 'every N' | universal negative syllogism
1. All whales are mammals.
2. No mammal is a fish.
3. Therefore, no whale is a fish. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject universal negative syllogism
1. All whales are mammals.
2. No whale is a fish.
3. Therefore, no mammal is a fish. [1,2]
```

- **What it does not cover:** Illicit major/minor generalizations (`All A are B; No A is C` ⊢ `No B is C`), two negative premises, or particular premises (see `existential negative syllogism`).

---

## `universal negative instantiation`

- **Schema:** `No A is B` (or `No B is A`); `c is A` ⊢ `c is not B` / `It is not the case that c is B`.
- **Accepted:**

```argcheck-accept universal negative instantiation
1. I am aware of something.
2. No physical thing is aware of something.
3. Therefore, I am not a physical thing. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject universal negative instantiation
1. The mind is invisible.
2. No physical thing is visible.
3. Therefore, the mind is not a physical thing. [1,2]
```

- **What it does not cover:** Morphological contraries (`invisible` vs `visible`) without a bridging premise, downward-entailing or quantified subjects (`no cat`, `few cats`, `at most one cat`) treated as individual witnesses, or non-copular relational verbs (see `relational universal negative instantiation`).

---

## `schematic instantiation`

- **Schema:** `If Φ(U..Z), then Ψ(U..Z)`; `Φ(t_U..t_Z)` ⊢ `Ψ(t_U..t_Z)` (consistent substitution for standalone uppercase `U`, `V`, `W`, `X`, `Y`, `Z` across a single antecedent).
- **Accepted:**

```argcheck-accept schematic instantiation
1. If X admires Y, then Y trusts X.
2. Alice admires Bob.
3. Therefore, Bob trusts Alice. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject schematic instantiation
1. If x is conceivable, then x is metaphysically possible.
2. A cube is conceivable.
3. Therefore, a cube is metaphysically possible. [1,2]
```

```argcheck-reject schematic instantiation
1. It is not the case that Socrates is wise.
2. If X is wise, then X is happy.
3. Therefore, it is not the case that Socrates is happy. [1,2]
```

- **What it does not cover:** Lowercase variables (`x` is reserved for explicit `is such that` binders), letters embedded inside words (`Xylophone`), unbound placeholders in the consequent, or empty substitutions. Since v28.1 a placeholder also may not bind text that carries a scope operator: a binding that opens with negation, a non-upward quantifier or a falsity wrapper (`not every`, `no`, `nobody`, `at most`, `it is not`, `it is false`), or one that ends inside an attitude complement before its verb (`Plato denies that Socrates`). This applies to every schematic form. Since v30 a placeholder that occurs more than once in the antecedent may not bind an indefinite, existential or disjunctive term (`some philosopher`, `a cat`, `someone`, `two dogs`, `Socrates or Plato`): separate occurrences, matched against separate premises, could concern different witnesses. Names, definite descriptions and `every`/`each`/`all` phrases still bind there. A placeholder that occurs more than once in the consequent may not bind a universal or disjunctive term (`If X is wise, then X admires X` does not give `every philosopher admires every philosopher`).

---

## `non-copular universal instantiation`

- **Schema:** `Every A V` (or `All As V`); `c is an A` ⊢ `c V` (where `V` is a non-copular verb phrase).
- **Accepted:**

```argcheck-accept non-copular universal instantiation
1. Every cat purrs.
2. Felix is a cat.
3. Therefore, Felix purrs. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject non-copular universal instantiation
1. All men die.
2. Socrates is a man.
3. Therefore, Socrates dies. [1,2]
```

- **What it does not cover:** Irregular or inflected verb changes between plural and singular (`die` vs `dies`; write `Every man dies` so the verb phrase matches verbatim), attitude-embedded memberships (`Ada believes whales are mammals`), or dropping a relative-clause restrictor. Since v29.1 a singular instance (`Rex is a square`) may also take the third-person-singular verb (`Rex has four sides`); tense changes are not accepted.

---

## `universal instantiation`

- **Schema:** `Every A is B` (or `All As are B`); `c is an A` ⊢ `c is B`.
- **Valid form, refused by v44 strict grammar (`term.inflection_verbatim`):**

```argcheck-refused term.inflection_verbatim | singular 'every N' | universal instantiation
1. All humans are mortal.
2. Socrates is a human.
3. Therefore, Socrates is mortal. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject universal instantiation
1. All humans are mortal.
2. Socrates is mortal.
3. Therefore, Socrates is a human. [1,2]
```

- **What it does not cover:** Affirming the predicate (`Socrates is mortal` ⊢ `Socrates is a human`), instantiating from `Some` or `Few` or `No`, or non-copular predicates (handled by `non-copular universal instantiation`).

---

## `existential generalization`

- **Schema:** `c is P` ⊢ `Something is P` / `There is something that is P` / `There exists something that is P`.
- **Valid form, refused by v44 strict grammar (`eg.name_witness`):**

```argcheck-refused eg.name_witness | state the existential claim as its own premise | existential generalization
1. Sample R is defective.
2. Therefore, there exists something that is defective. [1]
```

- **Rejected near-miss:**

```argcheck-reject existential generalization
1. Every sample is defective.
2. Therefore, there exists something that is defective. [1]
```

- **What it does not cover:** Importing existence from a universal claim without a witness, generalizing out of negated, modal, or belief-embedded clauses, or introducing a restricted class without a membership premise (use `existential introduction`).

---

## `existential introduction`

- **Schema:** `c is an A`; `c is P_1`; ...; `c is P_k` ⊢ `Some A is P_1 and ... and P_k` (or `There exist As that are ...`).
- **Valid form, refused by v44 strict grammar (`eg.name_witness`):**

```argcheck-refused eg.name_witness | state the existential claim as its own premise | existential introduction
1. Rex is a dog.
2. Rex is friendly.
3. Therefore, some dog is friendly. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject existential introduction
1. Rex is a dog.
2. Tom is friendly.
3. Therefore, some dog is friendly. [1,2]
```

- **What it does not cover:** Distinct witnesses across the cited premises (`Rex` vs `Tom`), universal premises without an individual witness (`Every cat is a mammal; Every cat is furry`), or extra cited premises not reflected in the existential predicate.

---

## `existential counterexample`

- **Schema:** `There exist As that are not B` (or `There exist As that are R and not B`) ⊢ `It is not the case that every A (that is R) is B`.
- **Valid form, refused by v44 strict grammar (`term.inflection_verbatim`):**

```argcheck-refused term.inflection_verbatim | singular 'every N' | existential counterexample
1. There exist cats that are not asleep.
2. Therefore, it is not the case that every cat is asleep. [1]
```

- **Rejected near-miss:**

```argcheck-reject existential counterexample
1. There exist healthy adults that are sleep-deprived and not impaired.
2. Therefore, it is not the case that every healthy child that is sleep-deprived is impaired. [1]
```

- **What it does not cover:** Class or restrictor drift (`adult` vs `child`), affirmative smuggling (`There exist cats that are furry` ⊢ `Every cat is furry`), or concluding a universal negative (`No cat is asleep`).

---

## `necessity elimination`

- **Schema:** `Necessarily, P` ⊢ `P` (modal axiom **T**).
- **Accepted:**

```argcheck-accept necessity elimination
1. Necessarily, rain falls.
2. Therefore, rain falls. [1]
```

- **Rejected near-miss:**

```argcheck-reject necessity elimination
1. Rain falls.
2. Therefore, necessarily, rain falls. [1]
```

- **What it does not cover:** Inferring necessity from actuality (`P` ⊢ `Necessarily, P`), inferring actuality from possibility (`Possibly, P` ⊢ `P`), or S4/S5 iteration principles.

---

## `existential syllogism`

- **Schema:** `Some A is B`; `Every B is C` ⊢ `Some A is C` (Darii).
- **Accepted:**

```argcheck-accept existential syllogism
1. Some cat is a hunter.
2. Every hunter is alert.
3. Therefore, some cat is alert. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject existential syllogism
1. Some cat is a hunter.
2. Some hunter is alert.
3. Therefore, some cat is alert. [1,2]
```

- **What it does not cover:** Two existential premises (`Some A is B; Some B is C`), undistributed middle (`Every B is C; Some A is C`), or strengthening the conclusion to `Every A is C`.

---

## `existential negative syllogism`

- **Schema:** `Some A is B`; `No B is C` (or `No C is B`) ⊢ `Some A is not C` (Ferio / Festino).
- **Accepted:**

```argcheck-accept existential negative syllogism
1. Some cat is a hunter.
2. No hunter is lazy.
3. Therefore, some cat is not lazy. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject existential negative syllogism
1. Some cat is a hunter.
2. No hunter is lazy.
3. Therefore, no cat is lazy. [1,2]
```

- **What it does not cover:** Strengthening the conclusion to `No A is C`, converting a particular negative (`Some C is not A`), or starting from `Some A is not B` plus `Every B is C`.

---

## `schematic existential introduction`

- **Schema:** `Φ(a)` ⊢ `There is an X such that Φ(X)` (replacing every whole-term occurrence of ground individual `a` with uppercase placeholder `X`..`Z` or `U`..`W`).
- **Valid form, refused by v44 strict grammar (`eg.name_witness`):**

```argcheck-refused eg.name_witness | state the existential claim as its own premise | schematic existential introduction
1. Ada solved the puzzle.
2. Therefore, there is an X such that X solved the puzzle. [1]
```

- **Rejected near-miss:**

```argcheck-reject schematic existential introduction
1. Ada admires Ada's teacher.
2. Therefore, there is an X such that X admires Ada's teacher. [1]
```

- **What it does not cover:** Partial replacement of a repeated name (for selective replacement, use explicit-variable `FOL existential introduction`), generalizing from pronouns or quantifier phrases (`Nothing`, `Every dog`), opaque/modal/existence-predicate contexts, or chaining a second schematic quantifier over an already-schematic sentence.

---

## `identity substitution`

- **Schema:** `a is identical to b` (or `a is identical with b`); `Φ(a)` ⊢ `Φ(b)` (or `Φ(b)` ⊢ `Φ(a)`), replacing every whole-term occurrence of the ground term.
- **Valid form, refused by v44 strict grammar (`identity.constant_sides`):**

```argcheck-refused identity.constant_sides | name the thing | identity substitution
1. Hesperus is identical to Phosphorus.
2. Hesperus is visible at dusk.
3. Therefore, Phosphorus is visible at dusk. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject identity substitution
1. Hesperus is Phosphorus.
2. Hesperus is visible at dusk.
3. Therefore, Phosphorus is visible at dusk. [1,2]
```

- **What it does not cover:** Bare `"is"` (`Hesperus is Phosphorus`; for Russellian `"The F is a"` substitution, see `description substitution`), partial replacement of only some occurrences (supported in explicit-variable `FOL identity substitution`), attitude/modal/deontic opaque contexts, or substituting predicates/indefinites (`Ada is identical to a doctor`). Since v28.2 this also holds when the quantifier phrase is capitalized or opens the sentence (`Nobody is identical to Felix`, `A cat is identical to Felix`), in the readable and the explicit-variable forms. Identity substitution is also blocked under `is looking for`, `looks for`, `seeks`, `wants`, `needs`, `is searching for`, `hopes for`, `fears`, `believes`, `imagines`, and `worships`. Extensional predicates such as `painted`, `met`, `kicked`, and `is near` remain substitutable.

---

## `constructive dilemma`

- **Schema:** `Either P_1 or ... or P_n`; `If P_1, then Q_1`; ...; `If P_n, then Q_n` ⊢ `Either Q_1 or ... or Q_n` (`n = 2` or `3`).
- **Accepted:**

```argcheck-accept constructive dilemma
1. Either it rains or it snows.
2. If it rains, then the road is wet.
3. If it snows, then the road is icy.
4. Therefore, either the road is wet or the road is icy. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject constructive dilemma
1. Either the road is wet or the road is icy.
2. If it rains, then the road is wet.
3. If it snows, then the road is icy.
4. Therefore, either it rains or it snows. [1,2,3]
```

- **What it does not cover:** Affirming the consequents, missing a branch conditional, collapsing identical consequents into a single non-disjunctive claim (use `simple dilemma`), or more than three disjuncts.

---

## `simple dilemma`

- **Schema:** `Either P_1 or ... or P_n`; `If P_1, then Q`; ...; `If P_n, then Q` ⊢ `Q` (`n = 2` or `3`).
- **Accepted:**

```argcheck-accept simple dilemma
1. Either it rains or it snows.
2. If it rains, then the game is cancelled.
3. If it snows, then the game is cancelled.
4. Therefore, the game is cancelled. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject simple dilemma
1. Either it rains or it snows.
2. If it rains, then the game is cancelled.
3. If it snows, then the game is postponed.
4. Therefore, the game is cancelled. [1,2,3]
```

- **What it does not cover:** Branches with non-identical consequents, missing a conditional for one of the disjuncts, or multi-step case reasoning without material conditionals (use Fitch `proof by cases`).

---

## `destructive dilemma`

- **Schema:** `If P_1, then Q_1`; ...; `If P_n, then Q_n`; `Either it is not the case that Q_1 or ... or it is not the case that Q_n` ⊢ `Either it is not the case that P_1 or ... or it is not the case that P_n` (`n = 2` or `3`).
- **Accepted:**

```argcheck-accept destructive dilemma
1. If it rains, then the road is wet.
2. If it snows, then the road is icy.
3. Either it is not the case that the road is wet or it is not the case that the road is icy.
4. Therefore, either it is not the case that it rains or it is not the case that it snows. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject destructive dilemma
1. If it rains, then the road is wet.
2. If it snows, then the road is icy.
3. Either it is not the case that it rains or it is not the case that it snows.
4. Therefore, either it is not the case that the road is wet or it is not the case that the road is icy. [1,2,3]
```

- **What it does not cover:** Denying the antecedents, unnegated disjuncts in the disjunctive premise, or missing a conditional branch.

---

## `contraposition`

- **Schema:** `If P, then Q` ⊢ `If it is not the case that Q, then it is not the case that P` (and the reverse direction).
- **Accepted:**

```argcheck-accept contraposition
1. If it rains, then it snows.
2. Therefore, if it is not the case that it snows, then it is not the case that it rains. [1]
```

- **Rejected near-miss:**

```argcheck-reject contraposition
1. If it rains, then it snows.
2. Therefore, if it is not the case that it rains, then it is not the case that it snows. [1]
```

- **What it does not cover:** Negating both sides without reversing antecedent and consequent (inverse fallacy), or contraposing counterfactual `had`/`would` conditionals (`If the match had not lit, then the match would not have been struck`).

---

## `De Morgan`

- **Schema:** `It is not the case that (P and Q)` ⊣⊢ `Either (not P) or (not Q)` and `It is not the case that (either P or Q)` ⊣⊢ `(not P) and (not Q)`.
- **Accepted:**

```argcheck-accept De Morgan
1. It is not the case that it is the case both that it rains, and that it snows.
2. Therefore, either it is not the case that it rains or it is not the case that it snows. [1]
```

- **Rejected near-miss:**

```argcheck-reject De Morgan
1. It is not the case that it is the case both that it rains, and that it snows.
2. Therefore, it is the case both that it is not the case that it rains, and that it is not the case that it snows. [1]
```

- **What it does not cover:** Preserving the same connective instead of flipping `and`/`or`, dropping one of the inner negations, or unscaffolded ambiguous scope under negation.

---

## `commutation`

- **Schema:** `P and Q` ⊢ `Q and P` and `Either P or Q` ⊢ `Either Q or P`.
- **Accepted:**

```argcheck-accept commutation
1. Either it rains or it snows.
2. Therefore, either it snows or it rains. [1]
```

- **Rejected near-miss:**

```argcheck-reject commutation
1. Either it rains or it snows.
2. Therefore, either the street is wet or it rains. [1]
```

- **What it does not cover:** Commuting a conditional (`If P, then Q` to `If Q, then P`), commuting inside a negated scope without applying De Morgan, or commuting biconditionals (reported under `biconditional commutation`).

---

## `exportation`

- **Schema:** `If (P and Q), then R` ⊢ `If P, then if Q, then R`.
- **Accepted:**

```argcheck-accept exportation
1. If it is the case both that it rains, and that it snows, then the street is wet.
2. Therefore, if it rains, then if it snows, then the street is wet. [1]
```

- **Rejected near-miss:**

```argcheck-reject exportation
1. If either it rains or it snows, then the street is wet.
2. Therefore, if it rains, then if it snows, then the street is wet. [1]
```

- **What it does not cover:** Exporting a disjunctive antecedent, or exporting counterfactual conditionals.

---

## `importation`

- **Schema:** `If P, then if Q, then R` ⊢ `If (P and Q), then R`.
- **Accepted:**

```argcheck-accept importation
1. If it rains, then if it snows, then the street is wet.
2. Therefore, if it is the case both that it rains, and that it snows, then the street is wet. [1]
```

- **Rejected near-miss:**

```argcheck-reject importation
1. If it rains, then if it snows, then the street is wet.
2. Therefore, if either it rains or it snows, then the street is wet. [1]
```

- **What it does not cover:** Importing into a disjunctive antecedent, or importing nested counterfactual conditionals.

---

## `material implication`

- **Schema:** `If P, then Q` ⊣⊢ `Either it is not the case that P or Q`.
- **Accepted:**

```argcheck-accept material implication
1. If it rains, then it snows.
2. Therefore, either it is not the case that it rains or it snows. [1]
```

- **Rejected near-miss:**

```argcheck-reject material implication
1. If it rains, then it snows.
2. Therefore, either it rains or it snows. [1]
```

- **What it does not cover:** Omitting the negation on the antecedent, or turning a disjunction (`Either not P or Q`) into a counterfactual conditional (`If P had ..., then Q would ...`).

---

## `distribution`

- **Schema:** `P and (either Q or R)` ⊣⊢ `Either (P and Q) or (P and R)` and `Either P or (Q and R)` ⊣⊢ `(Either P or Q) and (Either P or R)`.
- **Accepted:**

```argcheck-accept distribution
1. It is the case both that it rains, and that either it snows or the street is wet.
2. Therefore, either it is the case both that it rains, and that it snows or it is the case both that it rains, and that the street is wet. [1]
```

- **Rejected near-miss:**

```argcheck-reject distribution
1. Either it rains or it is the case both that it snows, and that the street is wet.
2. Therefore, it is the case both that it rains, and that either it snows or the street is wet. [1]
```

- **What it does not cover:** Unscaffolded nested `and`/`or` sentences whose grouping is ambiguous, or swapping outer `and` and `or` without duplicating `P`.

---

## `absorption`

- **Schema:** `If P, then Q` ⊢ `If P, then (P and Q)` and `Either P or (P and Q)` ⊢ `P`.
- **Accepted:**

```argcheck-accept absorption
1. If it rains, then it snows.
2. Therefore, if it rains, then it is the case both that it rains, and that it snows. [1]
```

- **Rejected near-miss:**

```argcheck-reject absorption
1. Either it rains or it is the case both that it rains, and that it snows.
2. Therefore, it snows. [1]
```

- **What it does not cover:** Extracting the non-repeated conjunct `Q` from `Either P or (P and Q)`, or applying conditional absorption to counterfactuals.

---

## `double negation introduction`

- **Schema:** `P` ⊢ `It is not the case that it is not the case that P`.
- **Accepted:**

```argcheck-accept double negation introduction
1. It rains.
2. Therefore, it is not the case that it is not the case that it rains. [1]
```

- **Rejected near-miss:**

```argcheck-reject double negation introduction
1. It rains.
2. Therefore, it is not the case that it rains. [1]
```

- **What it does not cover:** Single negation, lexical antonyms, or inserting double negation inside a subclause rather than around the whole claim.

---

## `biconditional modus tollens`

- **Schema:** `P if and only if Q`; `It is not the case that Q` ⊢ `It is not the case that P` (and symmetrically when `P` is negated).
- **Accepted:**

```argcheck-accept biconditional modus tollens
1. It rains if and only if it snows.
2. It is not the case that it snows.
3. Therefore, it is not the case that it rains. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject biconditional modus tollens
1. It rains if and only if it snows.
2. It is not the case that it snows.
3. Therefore, it rains. [1,2]
```

- **What it does not cover:** Concluding the positive side instead of its negation, `"only if"` premises, or denying an unrelated clause.

---

## `biconditional chaining`

- **Schema:** `P if and only if Q`; `Q if and only if R` ⊢ `P if and only if R`.
- **Accepted:**

```argcheck-accept biconditional chaining
1. It rains if and only if it snows.
2. It snows if and only if the street is wet.
3. Therefore, it rains if and only if the street is wet. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject biconditional chaining
1. It rains if and only if it snows.
2. Ada sings if and only if the street is wet.
3. Therefore, it rains if and only if the street is wet. [1,2]
```

- **What it does not cover:** Non-matching middle clauses, chaining three or more biconditionals in one step, or mixing `if and only if` with one-way conditionals without first eliminating or introducing.

---

## `biconditional commutation`

- **Schema:** `P if and only if Q` ⊢ `Q if and only if P`.
- **Accepted:**

```argcheck-accept biconditional commutation
1. It rains if and only if it snows.
2. Therefore, it snows if and only if it rains. [1]
```

- **Rejected near-miss:**

```argcheck-reject biconditional commutation
1. It rains if and only if it snows.
2. Therefore, the street is wet if and only if it rains. [1]
```

- **What it does not cover:** Wording drift on either side, or unparenthesized compound clauses whose outer operator is not `if and only if`.

---

## `negated conditional`

- **Schema:** `It is not the case that if P, then Q` ⊢ `P` and `It is not the case that if P, then Q` ⊢ `It is not the case that Q`.
- **Accepted:**

```argcheck-accept negated conditional
1. It is not the case that if it rains, then it snows.
2. Therefore, it rains. [1]
```

- **Rejected near-miss:**

```argcheck-reject negated conditional
1. It is not the case that if it rains, then it snows.
2. Therefore, it is not the case that it rains. [1]
```

- **What it does not cover:** Concluding `not P` or `Q`, or applying material conditional negation to a counterfactual (`It is not the case that if P had ..., then Q would ...` does not entail `P` was true).

---

## `identity symmetry`

- **Schema:** `a is identical to b` ⊢ `b is identical to a` (also accepts `is identical with`).
- **Accepted:**

```argcheck-accept identity symmetry
1. Ada is identical to Ben.
2. Therefore, Ben is identical to Ada. [1]
```

- **Rejected near-miss:**

```argcheck-reject identity symmetry
1. Ada is identical to Ben.
2. Therefore, Cy is identical to Ada. [1]
```

- **What it does not cover:** Bare `"is"` (`Ada is Ben`), identities embedded inside belief or modal prefixes, or negated identities.

---

## `quantifier negation`

- **Schema:** `It is not the case that every A is B` ⊢ `Some A is not B` and `It is not the case that some A is B` ⊢ `No A is B`.
- **Accepted:**

```argcheck-accept quantifier negation
1. It is not the case that every man is mortal.
2. Therefore, some man is not mortal. [1]
```

- **Rejected near-miss:**

```argcheck-reject quantifier negation
1. It is not the case that every man is mortal.
2. Therefore, some man is mortal. [1]
```

- **What it does not cover:** Dropping the predicate negation (`Some A is B`), class/predicate drift, or non-copular relational quantifier negation (handled by `relational scope conversion`).

---

## `conversion`

- **Schema:** `Some A is a B` ⊢ `Some B is an A` and `No A is a B` ⊢ `No B is an A`.
- **Accepted:**

```argcheck-accept conversion
1. Some poet is a sailor.
2. Therefore, some sailor is a poet. [1]
```

- **Rejected near-miss:**

```argcheck-reject conversion
1. Every poet is a sailor.
2. Therefore, every sailor is a poet. [1]
```

- **What it does not cover:** Converting universal affirmatives (`Every A is a B`) or particular negatives (`Some A is not a B`), which are invalid.

---

## `universal modus tollens`

- **Schema:** `Every A is B`; `It is not the case that c is B` ⊢ `It is not the case that c is an A` (Camestres / Baroco instance).
- **Accepted:**

```argcheck-accept universal modus tollens
1. Every man is mortal.
2. It is not the case that Zeus is mortal.
3. Therefore, it is not the case that Zeus is a man. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject universal modus tollens
1. Every man is mortal.
2. It is not the case that Zeus is mortal.
3. Therefore, it is not the case that Zeus is a poet. [1,2]
```

- **What it does not cover:** Denying the subject class to deny the predicate (`Every man is mortal; Zeus is not a man` ⊢ `Zeus is not mortal`), or class/instance drift.

---

## `possibility introduction`

- **Schema:** `P` ⊢ `Possibly, P` and `Necessarily, P` ⊢ `Possibly, P`.
- **Accepted:**

```argcheck-accept possibility introduction
1. It rains.
2. Therefore, possibly, it rains. [1]
```

- **Rejected near-miss:**

```argcheck-reject possibility introduction
1. Possibly, it rains.
2. Therefore, necessarily, it rains. [1]
```

- **What it does not cover:** Strengthening possibility to actuality or necessity, or substituting identicals / generalizing witnesses inside `Possibly, ...`.

---

## `contradiction introduction`

- **Schema:** `P`; `It is not the case that P` ⊢ `contradiction`.
- **Accepted:**

```argcheck-accept contradiction introduction
1. It rains.
2. It is not the case that it rains.
3. Therefore, contradiction. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject contradiction introduction
1. It rains.
2. It is not the case that it snows.
3. Therefore, contradiction. [1,2]
```

- **What it does not cover:** Unmatched clauses, lexical contraries (`hot` and `cold` without a bridging premise), or cardinality clashes (`At least 4 ...; At most 2 ...`, handled by `counting contradiction`).

---

## `identity reflexivity`

- **Schema:** `Φ(a)` ⊢ `a is identical to a` (where `a` occurs as a whole term in a cited premise or in-scope step).
- **Accepted:**

```argcheck-accept identity reflexivity
1. Socrates is a man.
2. Therefore, Socrates is identical to Socrates. [1]
```

- **Rejected near-miss:**

```argcheck-reject identity reflexivity
1. Plato is a man.
2. Therefore, Socrates is identical to Socrates. [1]
```

- **What it does not cover:** Introducing `a is identical to a` when `a` does not appear as a whole word/phrase in the cited line (substrings like `Socratesian` do not count), or compound clauses in term position.

---

## `modal duality`

- **Schema:** `It is not the case that possibly, it is not the case that P` ⊣⊢ `Necessarily, P` and `It is not the case that necessarily, it is not the case that P` ⊣⊢ `Possibly, P`.
- **Accepted:**

```argcheck-accept modal duality
1. It is not the case that possibly, it is not the case that it rains.
2. Therefore, necessarily, it rains. [1]
```

- **Rejected near-miss:**

```argcheck-reject modal duality
1. It is not the case that possibly, it rains.
2. Therefore, necessarily, it rains. [1]
```

- **What it does not cover:** Missing either the outer or inner negation (`not possibly P` is `necessarily not P`, not `necessarily P`).

---

## `modal distribution`

- **Schema:** `Necessarily, if P, then Q` ⊢ `If necessarily, P, then necessarily, Q` (axiom **K**), `Necessarily, if P, then Q` ⊢ `If possibly, P, then possibly, Q`, and `Necessarily, (P and Q)` ⊣⊢ `(Necessarily, P) and (Necessarily, Q)`.
- **Accepted:**

```argcheck-accept modal distribution
1. Necessarily, if it rains, then it snows.
2. Therefore, if necessarily, it rains, then necessarily, it snows. [1]
```

- **Rejected near-miss:**

```argcheck-reject modal distribution
1. If necessarily, it rains, then necessarily, it snows.
2. Therefore, necessarily, if it rains, then it snows. [1]
```

- **What it does not cover:** Converse **K**, distributing `Possibly` over a conditional or conjunction, or mixed modalities (`If possibly P, then necessarily Q`).

---

## `relational universal instantiation`

- **Schema:** `R(..., every C_i, ...)`; `a_i is a C_i`; ... ⊢ `R(..., a_i, ...)` across binary through 5-place relational clauses (up to 4 quantifiers per sentence).
- **Accepted:**

```argcheck-accept relational universal instantiation
1. Every lab sells every model to every government.
2. Acme is a lab.
3. Japan is a government.
4. Therefore, Acme sells every model to Japan. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject relational universal instantiation
1. Alice sends some package to every recipient.
2. Parcel is a package.
3. Bob is a recipient.
4. Therefore, Alice sends Parcel to Bob. [1,2,3]
```

- **What it does not cover:** Instantiating an existential (`some`) slot with a chosen constant, swapping argument roles or prepositions (`to` vs `from`), active-to-passive conversion, `"any"` or bare plural `"all Xs"` in relational object slots, sentences with more than 4 quantifiers (use explicit-variable `FOL universal instantiation`), or an object of a verb that is not on the closed extensional allowlist (v38.0; `seeks`, `awaits`, `admires` may be intensional, so state the instance as a premise). Note: single-membership subject instantiation on non-copular verbs may be reported as `non-copular universal instantiation`.

---

## `relational existential generalization`

- **Schema:** `R(..., a, ...)`; `a is a C` ⊢ `R(..., some C, ...)`.
- **Valid form, refused by v44 strict grammar (`eg.subject_only`):**

```argcheck-refused eg.subject_only | state the existential claim as its own premise | relational existential generalization
1. Alice greets Bob.
2. Bob is a teacher.
3. Therefore, Alice greets some teacher. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject relational existential generalization
1. Alice greets Bob.
2. Bob is a student.
3. Therefore, Alice greets some teacher. [1,2]
```

- **What it does not cover:** Class mismatch (`student` vs `teacher`), swapping subject and object positions, verb changes, or extra unused membership premises.

---

## `relational universal negative instantiation`

- **Schema:** `R(..., no C, ...)`; `a is a C` ⊢ `It is not the case that R(..., a, ...)`.
- **Accepted:**

```argcheck-accept relational universal negative instantiation
1. Alice greets no tyrant.
2. Bob is a tyrant.
3. Therefore, it is not the case that Alice greets Bob. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject relational universal negative instantiation
1. Alice greets no tyrant.
2. Bob is a tyrant.
3. Therefore, it is not the case that Bob greets Alice. [1,2]
```

- **What it does not cover:** Reversing argument positions (`Bob greets Alice`), omitting the outer negation in the conclusion, or nested noun-phrase quantifiers (`every student of some school`).

---

## `relational scope conversion`

- **Schema:** Scope-preserving relational rewrites that hold on **every** reading of the premise, including commuting like quantifiers, weakening `Some B is such that every A ...` to `Every A is such that ... some B`, relational quantifier negation (`It is not the case that some A R b` ⊣⊢ `No A R b`), `is such that` scaffold introduction/elimination, and ditransitive alternation (`Alice gave Bob a book` ⊣⊢ `Alice gave a book to Bob`).
- **Accepted:**

```argcheck-accept relational scope conversion
1. Every student admires every teacher.
2. Therefore, every teacher is such that every student admires that teacher. [1]
```

- **Rejected near-miss:**

```argcheck-reject relational scope conversion
1. Every student admires some teacher.
2. Therefore, some teacher is such that every student admires that teacher. [1]
```

- **What it does not cover:** Quantifier-shift fallacies from a scope-ambiguous surface sentence (`Every student admires some teacher`) to its strong wide-existential reading; write the intended scope explicitly with `is such that`.

---

## `counterfactual modus ponens`

- **Schema:** `If P_cf, then Q_cf`; `P_ind` ⊢ `Q_ind` (detaching a subjunctive `had`/`were` ... `would have`/`would be` conditional via its indicative antecedent).
- **Accepted:**

```argcheck-accept counterfactual modus ponens
1. If the match had been struck, then the match would have lit.
2. The match was struck.
3. Therefore, the match lit. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual modus ponens
1. If the match had been struck, then the match would light.
2. The match was struck.
3. Therefore, the match lights. [1,2]
```

- **What it does not cover:** Bare `would <verb>` or irregular past-participle shifts (`had gone` vs `went`) where indicative mood cannot be recovered purely from auxiliaries (add a visible bridging premise), `might have` / `could have` conditionals, or `would` embedded inside an attitude report. Indicative `might`/`could`/`may`/`can` atoms instead use ordinary forms verbatim; they do not receive counterfactual mood rewriting. Bare `might` with an antecedent carrying the existing `had`/`were` mood markers remains conservatively counterfactual.

---

## `counterfactual modus tollens`

- **Schema:** `If P_cf, then Q_cf`; `It is not the case that Q_ind` ⊢ `It is not the case that P_ind`.
- **Accepted:**

```argcheck-accept counterfactual modus tollens
1. If the match had been struck, then the match would have lit.
2. It is not the case that the match lit.
3. Therefore, it is not the case that the match was struck. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual modus tollens
1. If Alice had gone home, then Bob would have lit a fire.
2. It is not the case that Bob lit a fire.
3. Therefore, it is not the case that Alice went home. [1,2]
```

- **What it does not cover:** Irregular verb stem rewrites (`had gone` to `went`) without a bridging premise, or concluding a counterfactual contrapositive (`If not Q_cf, then not P_cf`).

---

## `counterfactual to material`

- **Schema:** `If P_cf, then Q_cf` ⊢ `If P_ind, then Q_ind` (or `Either it is not the case that P_ind or Q_ind`), by weak centering.
- **Accepted:**

```argcheck-accept counterfactual to material
1. If the match had been struck, then the match would have lit.
2. Therefore, if the match was struck, then the match lit. [1]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual to material
1. If the match was struck, then the match lit.
2. Therefore, if the match had been struck, then the match would have lit. [1]
```

- **What it does not cover:** Strengthening an indicative conditional (or `P_ind; Q_ind` via strong centering) into a counterfactual conditional.

---

## `counterfactual consequent conjunction`

- **Schema:** `If P_cf, then Q_cf`; `If P_cf, then R_cf` ⊢ `If P_cf, then Q_cf and R_cf` (agglomeration under a verbatim shared antecedent).
- **Accepted:**

```argcheck-accept counterfactual consequent conjunction
1. If the match had been struck, then the match would have lit.
2. If the match had been struck, then the room would have warmed.
3. Therefore, if the match had been struck, then the match would have lit and the room would have warmed. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual consequent conjunction
1. If the match had been struck, then the match would have lit.
2. If the lighter had been lit, then the room would have warmed.
3. Therefore, if the match had been struck, then the match would have lit and the room would have warmed. [1,2]
```

- **What it does not cover:** Different antecedents, antecedent strengthening (`If P had ... and R had ..., then Q would ...`), or counterfactual transitivity.

---

## `counterfactual consequent conjunction elimination`

- **Schema:** `If P_cf, then Q_cf and R_cf` ⊢ `If P_cf, then Q_cf` (or `If P_cf, then R_cf`).
- **Accepted:**

```argcheck-accept counterfactual consequent conjunction elimination
1. If the match had been struck, then the match would have lit and the room would have warmed.
2. Therefore, if the match had been struck, then the match would have lit. [1]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual consequent conjunction elimination
1. If either the match had been struck or the lighter had been lit, then the room would have warmed.
2. Therefore, if the match had been struck, then the room would have warmed. [1]
```

- **What it does not cover:** Simplifying a disjunctive counterfactual antecedent (which is invalid in **VW** counterfactual logic), or extracting from an unscaffolded attitude report.

---

## `counterfactual consequent disjunction introduction`

- **Schema:** `If P_cf, then Q_cf` ⊢ `If P_cf, then either Q_cf or R_cf` (in either disjunct order, when both disjuncts carry `would` counterfactual mood).
- **Accepted:**

```argcheck-accept counterfactual consequent disjunction introduction
1. If the match had been struck, then the match would have lit.
2. Therefore, if the match had been struck, then either the match would have lit or the room would have warmed. [1]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual consequent disjunction introduction
1. If the match had been struck, then the match might have lit.
2. Therefore, if the match had been struck, then either the match would have lit or the room would have warmed. [1]
```

- **What it does not cover:** Upgrading counterfactual `might have` or `could have` to `would have` (v31 permits indicative modal atoms only in ordinary forms, without such an upgrade), or conditional excluded middle (`Either (if P_cf, then Q_cf) or (if P_cf, then not Q_cf)`).

---

## `counterfactual dilemma`

- **Schema:** Constructive, simple, or destructive dilemma combining an indicative disjunction with counterfactual conditionals via weak centering (e.g., `Either P_1_ind or P_2_ind`; `If P_1_cf, then Q_cf`; `If P_2_cf, then Q_cf` ⊢ `Q_ind`).
- **Accepted:**

```argcheck-accept counterfactual dilemma
1. Either the match was struck or the lighter was lit.
2. If the match had been struck, then the room would have warmed.
3. If the lighter had been lit, then the room would have warmed.
4. Therefore, the room warmed. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual dilemma
1. Either the match was struck or the lighter was lit.
2. If the match had been struck, then the match would have lit.
3. If the lighter had been lit, then the room would have warmed.
4. Therefore, the room warmed. [1,2,3]
```

- **What it does not cover:** Non-matching consequents in a simple dilemma, missing a branch conditional, or irregular verb stems without a bridging premise.

---

## `counterfactual reflexivity`

- **Schema:** `Φ(P_ind)` ⊢ `If P_cf, then P_would` (where `P_cf` and `P_would` map to the same indicative clause occurring in the cited step).
- **Accepted:**

```argcheck-accept counterfactual reflexivity
1. The match was struck.
2. Therefore, if the match had been struck, then the match would have been struck. [1]
```

- **Rejected near-miss:**

```argcheck-reject counterfactual reflexivity
1. The lighter was lit.
2. Therefore, if the match had been struck, then the match would have been struck. [1]
```

- **What it does not cover:** Introducing a reflexive counterfactual whose indicative clause does not occur in a cited premise, or non-matching antecedent and consequent clauses.

---

## `conditional proof`

- **Schema:** Subproof `i..j` starting with `Suppose P.` and ending with `Q.` ⊢ `If P, then Q. [i-j]`.
- **Accepted:**

```argcheck-accept conditional proof
1. If it rains, then the street is wet.
2. If the street is wet, then the cars slow.
| 3. Suppose it rains.
| 4. Therefore, the street is wet. [1,3]
| 5. Therefore, the cars slow. [2,4]
6. Therefore, if it rains, then the cars slow. [3-5]
```

- **Rejected near-miss:**

```argcheck-reject conditional proof
1. If it rains, then the street is wet.
2. If the street is wet, then the cars slow.
| 3. Suppose it rains.
| 4. Therefore, the street is wet. [1,3]
| 5. Therefore, the cars slow. [2,4]
6. Therefore, if it rains, then it snows. [3-5]
```

- **What it does not cover:** Discharging a subproof into a counterfactual `would` conditional (antecedent strengthening fails for counterfactuals), citing individual lines inside a closed subproof, or citing a partial range instead of the whole open-to-close block `[i-j]`.

---

## `reductio`

- **Schema:** Subproof `i..j` starting with `Suppose P.` and ending with `contradiction.` ⊢ `It is not the case that P. [i-j]` (and classical `Suppose it is not the case that P.` ... `contradiction.` ⊢ `P. [i-j]`).
- **Accepted:**

```argcheck-accept reductio
1. It is not the case that it rains.
| 2. Suppose it rains.
| 3. Therefore, contradiction. [1,2]
4. Therefore, it is not the case that it rains. [2-3]
```

- **Rejected near-miss:**

```argcheck-reject reductio
| 1. Suppose it rains.
| 2. Therefore, it rains. [1]
3. Therefore, it is not the case that it rains. [1-2]
```

- **What it does not cover:** Subproofs whose final line is not `contradiction`, or extra uncited ranges/premises attached to the discharge step.

---

## `proof by cases`

- **Schema:** `Either P_1 or ... or P_n`; subproofs `i_1..j_1` (`Suppose P_1.` ... `R.`) through `i_n..j_n` (`Suppose P_n.` ... `R.`) ⊢ `R. [d,i_1-j_1,...,i_n-j_n]`.
- **Accepted:**

```argcheck-accept proof by cases
1. Either it rains or it snows.
2. If it rains, then the street is wet.
3. If it snows, then the street is wet.
| 4. Suppose it rains.
| 5. Therefore, the street is wet. [2,4]
| 6. Suppose it snows.
| 7. Therefore, the street is wet. [3,6]
8. Therefore, the street is wet. [1,4-5,6-7]
```

- **Rejected near-miss:**

```argcheck-reject proof by cases
1. Either it rains or it snows.
2. If it rains, then the street is wet.
3. If it snows, then the road is icy.
| 4. Suppose it rains.
| 5. Therefore, the street is wet. [2,4]
| 6. Suppose it snows.
| 7. Therefore, the road is icy. [3,6]
8. Therefore, the street is wet. [1,4-5,6-7]
```

- **What it does not cover:** Branches ending in different conclusions, omitting a disjunct's subproof, or citing a step inside a sibling subproof.

---

## `universal generalization`

- **Schema:** Subproof `i..j` starting with `Let c be an arbitrary A.` (for fresh `c`) and ending with `c is B.` ⊢ `Every A is B. [i-j]`.
- **Accepted:**

```argcheck-accept universal generalization
1. Every man is mortal.
| 2. Let a be an arbitrary man.
| 3. Therefore, a is mortal. [1,2]
4. Therefore, every man is mortal. [2-3]
```

- **Rejected near-miss:**

```argcheck-reject universal generalization
1. Every man is mortal.
| 2. Let a be an arbitrary man.
| 3. Therefore, a is mortal. [1,2]
4. Therefore, a is mortal. [2-3]
```

- **What it does not cover:** Eigenname `c` occurring in any outer in-scope premise/assumption or escaping into the discharged conclusion, case mismatch (`a` vs `A`), or relational conclusions (handled by `relational universal generalization` or `FOL universal generalization`).

---

## `existential elimination`

- **Schema:** `Some A is B`; subproof `i..j` starting with `Suppose c is an A and c is B.` (for fresh `c`) and ending with `c`-free `Q.` ⊢ `Q. [k,i-j]`.
- **Valid form, refused by v44 strict grammar (`eg.name_witness`):**

```argcheck-refused eg.name_witness | state the existential claim as its own premise | existential elimination
1. Some man is mortal.
| 2. Suppose a is a man and a is mortal.
| 3. Therefore, a is mortal. [2]
| 4. Therefore, something is mortal. [3]
5. Therefore, something is mortal. [1,2-4]
```

- **Rejected near-miss:**

```argcheck-reject existential elimination
1. Some man is mortal.
| 2. Suppose a is a man and a is mortal.
| 3. Therefore, a is mortal. [2]
4. Therefore, a is mortal. [1,2-3]
```

- **What it does not cover:** Witness name `c` escaping into the conclusion `Q`, reusing a non-fresh name already in scope, or assuming extra facts about `c` beyond the existential premise.

---

## `relational universal generalization`

- **Schema:** Subproof `i..j` starting with `Let c be an arbitrary A.` (for fresh `c`) and ending with relational `R(..., c, ...).` ⊢ `R(..., every A, ...). [i-j]`.
- **Accepted:**

```argcheck-accept relational universal generalization
1. Every student admires every teacher.
| 2. Let b be an arbitrary student.
| 3. Therefore, b admires every teacher. [1,2]
4. Therefore, every student admires every teacher. [2-3]
```

- **Rejected near-miss:**

```argcheck-reject relational universal generalization
1. b admires every teacher.
| 2. Let b be an arbitrary student.
| 3. Therefore, b admires every teacher. [1]
4. Therefore, every student admires every teacher. [2-3]
```

- **What it does not cover:** Non-fresh eigennames, case changes on the eigenname inside the subproof, or strengthening scope so an inner existential improperly takes wide scope over `every A`.

---

## `relational existential elimination`

- **Schema:** `R(..., some A, ...)`; subproof `i..j` starting with `Suppose c is an A and R(..., c, ...).` (for fresh `c`) and ending with `c`-free `Q.` ⊢ `Q. [k,i-j]`.
- **Valid form, refused by v44 strict grammar (`scope.clause_start`):**

```argcheck-refused scope.clause_start | move the operator into each clause | relational existential elimination
1. Some student admires Bob.
| 2. Suppose c is a student and c admires Bob.
| 3. Therefore, c is a student. [2]
| 4. Therefore, c admires Bob. [2]
| 5. Therefore, some student admires Bob. [4,3]
6. Therefore, some student admires Bob. [1,2-5]
```

- **Rejected near-miss:**

```argcheck-reject relational existential elimination
1. Some student admires Bob.
| 2. Suppose c is a student and c admires Bob.
| 3. Therefore, c admires Bob. [2]
4. Therefore, c admires Bob. [1,2-3]
```

- **What it does not cover:** Witness name `c` escaping into `Q`, non-fresh witness names, or eliminating an inner existential that sits beneath a universal quantifier without first instantiating the outer universal.

---

## `property abstraction`

- **Schema:** `a is F` ⊢ `a has the property of being F` (and under outer `It is not the case that ...`).
- **Accepted:**

```argcheck-accept property abstraction
1. Socrates is wise.
2. Therefore, Socrates has the property of being wise. [1]
```

- **Rejected near-miss:**

```argcheck-reject property abstraction
1. Socrates is wise.
2. Therefore, Socrates has the property of being intelligent. [1]
```

- **What it does not cover:** Predicate paraphrase, non-copular verb predicates (`Superman flies` to `being flying`), quantified or higher-order bearers (`Every man`, `The property of being wise`), or rewriting inside conditionals, modals, or belief reports.

---

## `property concretion`

- **Schema:** `a has the property of being F` ⊢ `a is F` (and under outer `It is not the case that ...`).
- **Accepted:**

```argcheck-accept property concretion
1. Socrates has the property of being wise.
2. Therefore, Socrates is wise. [1]
```

- **Rejected near-miss:**

```argcheck-reject property concretion
1. Socrates has the property of being wise.
2. Therefore, Plato is wise. [1]
```

- **What it does not cover:** Changing the bearer, paraphrasing `F`, quantified bearers, or concreting below modal, conditional, or attitude operators.

---

## `property instantiation`

- **Schema:** `a has every property that b has`; `b has the property of being F` ⊢ `a has the property of being F`.
- **Accepted:**

```argcheck-accept property instantiation
1. Socrates has every property that Plato has.
2. Plato has the property of being wise.
3. Therefore, Socrates has the property of being wise. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject property instantiation
1. Socrates has every property that Plato has.
2. Socrates has the property of being wise.
3. Therefore, Plato has the property of being wise. [1,2]
```

- **What it does not cover:** Reversing the inclusion direction (`Socrates` to `Plato`), unabstracted `Plato is wise` without `the property of being ...`, or substituting names inside an opaque `such that ... believes that ...` predicate.

---

## `property existential`

- **Schema:** `a has the property of being F` ⊢ `a has some property`.
- **Accepted:**

```argcheck-accept property existential
1. Socrates has the property of being wise.
2. Therefore, Socrates has some property. [1]
```

- **Rejected near-miss:**

```argcheck-reject property existential
1. It is not the case that Socrates has the property of being wise.
2. Therefore, Socrates has some property. [1]
```

- **What it does not cover:** Negated property possession, changing the bearer, or quantified subjects (`Every man`).

---

## `shared property`

- **Schema:** `a has the property of being F`; `b has the property of being F` ⊢ `a shares some property with b`.
- **Accepted:**

```argcheck-accept shared property
1. Socrates has the property of being wise.
2. Plato has the property of being wise.
3. Therefore, Socrates shares some property with Plato. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject shared property
1. Socrates has the property of being wise.
2. Plato has the property of being just.
3. Therefore, Socrates shares some property with Plato. [1,2]
```

- **What it does not cover:** Non-identical predicates (`wise` vs `just`), quantified bearers, or unabstracted copular premises (`Socrates is wise`).

---

## `indiscernibility of identicals`

- **Schema:** `a is identical to b` ⊢ `a has every property that b has` (Leibniz's law).
- **Accepted:**

```argcheck-accept indiscernibility of identicals
1. Socrates is identical to Plato.
2. Therefore, Socrates has every property that Plato has. [1]
```

- **Rejected near-miss:**

```argcheck-reject indiscernibility of identicals
1. It is not the case that Socrates is identical to Plato.
2. Therefore, Socrates has every property that Plato has. [1]
```

- **What it does not cover:** Negated identity, bare `"is"` in place of `is identical to`, or quantified/higher-order terms as bearers. (Note: the converse — identity of indiscernibles — is derivable in multiple steps via `identity reflexivity`, `property abstraction`, `property instantiation`, and `property concretion`, not assumed as a primitive one-step rule.)

---

## `property generalization`

- **Schema:** Subproof `i..j` starting with `Let P be an arbitrary property that b has.` (for fresh `P`) and ending with `a has P.` ⊢ `a has every property that b has. [i-j]`.
- **Valid form, refused by v44 strict grammar (`identity.constant_sides`):**

```argcheck-refused identity.constant_sides | name the thing | property generalization
1. Socrates is identical to Plato.
| 2. Let P be an arbitrary property that Plato has.
| 3. Therefore, Socrates has P. [1,2]
4. Therefore, Socrates has every property that Plato has. [2-3]
```

- **Rejected near-miss:**

```argcheck-reject property generalization
1. Socrates is identical to Plato.
| 2. Let P be an arbitrary property that Plato has.
| 3. Therefore, Socrates has P. [1,2]
4. Therefore, Aristotle has every property that Plato has. [2-3]
```

- **What it does not cover:** Non-fresh property eigennames, `P` escaping into the conclusion, case changes (`P` vs `p`), or general mathematical induction.

---

## `property instantiation (arbitrary property)`

- **Schema:** `a has every property that b has`; `b has P` (or `Let P be an arbitrary property that b has`) ⊢ `a has P` (inside the open scope of `Let P be an arbitrary property ...`).
- **Accepted:**

```argcheck-accept property instantiation (arbitrary property)
1. Plato has every property that Socrates has.
2. Aristotle has every property that Plato has.
| 3. Let P be an arbitrary property that Socrates has.
| 4. Therefore, Plato has P. [1,3]
| 5. Therefore, Aristotle has P. [2,4]
6. Therefore, Aristotle has every property that Socrates has. [3-5]
```

- **Rejected near-miss:**

```argcheck-reject property instantiation (arbitrary property)
1. Plato has every property that Socrates has.
2. Socrates has P.
3. Therefore, Plato has P. [1,2]
```

- **What it does not cover:** Untyped nouns (`Socrates has courage`) or property letters `P` used outside an open `Let P be an arbitrary property ...` subproof.

---

## `distinguishing property`

- **Schema:** `a has the property of being F` (or in-scope arbitrary property `P`); `It is not the case that b has the property of being F` (or `P`) ⊢ `a has some property that b lacks`.
- **Accepted:**

```argcheck-accept distinguishing property
1. Socrates has the property of being wise.
2. It is not the case that Plato has the property of being wise.
3. Therefore, Socrates has some property that Plato lacks. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject distinguishing property
1. Socrates has the property of being wise.
2. It is not the case that Plato has the property of being wise.
3. Therefore, Plato has some property that Socrates lacks. [1,2]
```

- **What it does not cover:** Reversing who has and who lacks the property, alternate phrasings such as `that Plato does not have`, mismatched predicates, or untyped nouns.

---

## `distinguishing property nonidentity`

- **Schema:** `a has some property that b lacks` ⊢ `It is not the case that a is identical to b`.
- **Accepted example:**

```argcheck-accept distinguishing property nonidentity
1. Socrates has some property that Plato lacks.
2. Therefore, it is not the case that Socrates is identical to Plato. [1]
```

- **Rejected near-miss:**

```argcheck-reject distinguishing property nonidentity
1. Socrates has some property that Plato lacks.
2. Therefore, it is not the case that Socrates is identical to Aristotle. [1]
```

- **What it does not cover:** Concluding positive identity, changing either individual bearer, or modal lifting (`Necessarily, ...`).

---

## `distinguishing property non-inclusion`

- **Schema:** `a has some property that b lacks` ⊢ `It is not the case that b has every property that a has`.
- **Accepted:**

```argcheck-accept distinguishing property non-inclusion
1. Socrates has some property that Plato lacks.
2. Therefore, it is not the case that Plato has every property that Socrates has. [1]
```

- **Rejected near-miss:**

```argcheck-reject distinguishing property non-inclusion
1. Socrates has some property that Plato lacks.
2. Therefore, it is not the case that Socrates has every property that Plato has. [1]
```

- **What it does not cover:** Reversing the inclusion direction (`Socrates` may still have every property that `Plato` has), or quantified bearers.

---

## `relative clause inference`

- **Schema:** Single-level `that` / `who` / `which` relative-clause introduction, elimination, universal instantiation, existential introduction, and negative instantiation (e.g., `Every A who R b is P`; `c is an A`; `c R b` ⊢ `c is P`, and `c is an A that b R` ⊣⊢ `c is an A` + `b R c`).
- **Accepted:**

```argcheck-accept relative clause inference
1. Every reader who admires Pat is active.
2. Robin is a reader.
3. Robin admires Pat.
4. Therefore, Robin is active. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject relative clause inference
1. Every reader who admires Pat is active.
2. Robin is a reader.
3. Pat admires Robin.
4. Therefore, Robin is active. [1,2,3]
```

- **What it does not cover:** Swapped subject/object gaps (`Pat admires Robin` vs `Robin admires Pat`), omitting either the class premise or the relation premise, or unexpanded nested relative clauses (`a reader who visits a hero that Sam admires`; decompose with `is such that` or explicit FOL).

---

## `FOL reiteration`

- **Schema:** `Φ` ⊢ `Φ'` (up to bound-variable alpha-equivalence in explicit-variable first-order logic).
- **Accepted:**

```argcheck-accept FOL reiteration
1. every x is such that (x is red).
2. Therefore, every y is such that (y is red). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL reiteration
1. every x is such that (x is red).
2. Therefore, every x is such that (x is blue). [1]
```

- **What it does not cover:** Changing predicate symbols, unbound single-letter variables (`every x1 is such that (x2 is red)`), or shadowing an outer variable with the same binder name.

---

## `FOL conjunction introduction`

- **Schema:** `Φ`; `Ψ` ⊢ `(Φ) and (Ψ)` (when explicit-variable FOL is active).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`; lowercase constant or letter eigenvariable is not name evidence):**

```argcheck-refused scope.clause_start | move the operator into each clause | FOL conjunction introduction
1. every x is such that (x is red).
2. alice is blue.
3. Therefore, (every x is such that (x is red)) and (alice is blue). [1,2]
```

- **Rejected near-miss:**

```argcheck-reject FOL conjunction introduction
1. every x is such that (x is red).
2. alice is blue.
3. Therefore, (every x is such that (x is red)) and (blair is blue). [1,2]
```

- **What it does not cover:** Three-or-more-premise conjunction in a single `FOL` step (chain binary steps), unparenthesized compound scopes, or mismatched conjuncts.

---

## `FOL conjunction elimination`

- **Schema:** `(Φ) and (Ψ)` ⊢ `Φ` (or `Ψ`).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`; lowercase constant or letter eigenvariable is not name evidence):**

```argcheck-refused scope.clause_start | move the operator into each clause | FOL conjunction elimination
1. (every x is such that (x is red)) and (alice is blue).
2. Therefore, every x is such that (x is red). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL conjunction elimination
1. (every x is such that (x is red)) and (alice is blue).
2. Therefore, some x is such that (x is red). [1]
```

- **What it does not cover:** Weakening a conjunct while eliminating it (`every` to `some`), or extracting a conjunct from underneath an outer quantifier without first instantiating the quantifier.

---

## `FOL disjunction introduction`

- **Schema:** `Φ` ⊢ `either (Φ) or (Ψ)` (or `either (Ψ) or (Φ)`).
- **Valid form, refused by v44 strict grammar (`scope.clause_start`; lowercase constant or letter eigenvariable is not name evidence):**

```argcheck-refused scope.clause_start | move the operator into each clause | FOL disjunction introduction
1. every x is such that (x is red).
2. Therefore, either (every x is such that (x is red)) or (alice is blue). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL disjunction introduction
1. every x is such that (x is red).
2. Therefore, either (some x is such that (x is red)) or (alice is blue). [1]
```

- **What it does not cover:** Introducing a disjunction where neither disjunct is alpha-equivalent to the cited premise, or introducing an unbound variable in the added disjunct.

---

## `FOL modus ponens`

- **Schema:** `if (Φ) then (Ψ)`; `Φ` ⊢ `Ψ`.
- **Accepted:**

```argcheck-accept FOL modus ponens
1. if (every x is such that (x is red)) then (alice is blue).
2. every x is such that (x is red).
3. Therefore, alice is blue. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject FOL modus ponens
1. if (every x is such that (x is red)) then (alice is blue).
2. alice is blue.
3. Therefore, every x is such that (x is red). [1,2]
```

- **What it does not cover:** Affirming the consequent, modus tollens in one step on explicit-variable formulas (derive via `FOL reductio` + `FOL modus ponens`), or unparenthesized compound scopes.

---

## `FOL contradiction introduction`

- **Schema:** `Φ`; `it is not the case that (Φ)` ⊢ `contradiction`.
- **Accepted:**

```argcheck-accept FOL contradiction introduction
1. every x is such that (x is red).
2. it is not the case that (every x is such that (x is red)).
3. Therefore, contradiction. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject FOL contradiction introduction
1. every x is such that (x is red).
2. it is not the case that (some x is such that (x is red)).
3. Therefore, contradiction. [1,2]
```

- **What it does not cover:** Non-alpha-equivalent formulas (`every x` vs `some x`), or negating only the inner matrix instead of the whole formula without first instantiating.

---

## `FOL contradiction elimination`

- **Schema:** `contradiction` ⊢ `Φ` (ex falso quodlibet for any well-formed closed FOL formula `Φ`).
- **Accepted:**

```argcheck-accept FOL contradiction elimination
1. every x is such that (x is red).
2. it is not the case that (every x is such that (x is red)).
3. Therefore, contradiction. [1,2]
4. Therefore, some y is such that (y is blue). [3]
```

- **Rejected near-miss:**

```argcheck-reject FOL contradiction elimination
1. it is not the case that (every x is such that (x is red)).
2. Therefore, some y is such that (y is blue). [1]
```

- **What it does not cover:** Citing a negated formula rather than `contradiction`, or concluding an open formula with an unbound variable.

---

## `FOL double negation`

- **Schema:** `it is not the case that (it is not the case that (Φ))` ⊢ `Φ`.
- **Accepted:**

```argcheck-accept FOL double negation
1. it is not the case that (it is not the case that (every x is such that (x is red))).
2. Therefore, every x is such that (x is red). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL double negation
1. it is not the case that (every x is such that (x is red)).
2. Therefore, every x is such that (x is red). [1]
```

- **What it does not cover:** Removing a single negation, or eliminating double negation buried inside a quantifier body without instantiating and re-generalizing.

---

## `FOL double negation introduction`

- **Schema:** `Φ` ⊢ `it is not the case that (it is not the case that (Φ))`.
- **Accepted:**

```argcheck-accept FOL double negation introduction
1. every x is such that (x is red).
2. Therefore, it is not the case that (it is not the case that (every x is such that (x is red))). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL double negation introduction
1. every x is such that (x is red).
2. Therefore, it is not the case that (every x is such that (x is red)). [1]
```

- **What it does not cover:** Single negation, or wrapping only a subformula rather than the whole cited formula.

---

## `FOL biconditional introduction`

- **Schema:** `if (Φ) then (Ψ)`; `if (Ψ) then (Φ)` ⊢ `(Φ) if and only if (Ψ)`.
- **Accepted:**

```argcheck-accept FOL biconditional introduction
1. if (every x is such that (x is red)) then (alice is blue).
2. if (alice is blue) then (every x is such that (x is red)).
3. Therefore, (every x is such that (x is red)) if and only if (alice is blue). [1,2]
```

- **Rejected near-miss:**

```argcheck-reject FOL biconditional introduction
1. if (every x is such that (x is red)) then (alice is blue).
2. Therefore, (every x is such that (x is red)) if and only if (alice is blue). [1]
```

- **What it does not cover:** Supplying only one direction, or unparenthesized compound sides around `if and only if`.

---

## `FOL biconditional elimination`

- **Schema:** `(Φ) if and only if (Ψ)` ⊢ `if (Φ) then (Ψ)` (or `if (Ψ) then (Φ)`).
- **Accepted:**

```argcheck-accept FOL biconditional elimination
1. (every x is such that (x is red)) if and only if (alice is blue).
2. Therefore, if (every x is such that (x is red)) then (alice is blue). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL biconditional elimination
1. (every x is such that (x is red)) if and only if (alice is blue).
2. Therefore, (every x is such that (x is red)) and (alice is blue). [1]
```

- **What it does not cover:** Direct biconditional detachment in a single step on explicit-variable formulas (eliminate to `if (...) then (...)` first, then apply `FOL modus ponens`).

---

## `FOL universal instantiation`

- **Schema:** `every v is such that (Φ(v))` ⊢ `Φ(c)` (for any closed constant name `c`; note `no v is such that (Φ(v))` parses as `every v is such that (it is not the case that (Φ(v)))`).
- **Accepted:**

```argcheck-accept FOL universal instantiation
1. every x is such that (some y is such that (x admires y)).
2. Therefore, some y is such that (alice admires y). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL universal instantiation
1. every x is such that (x is red).
2. Therefore, x is red. [1]
```

- **What it does not cover:** Instantiating with a bare variable token (`x`, `y`, `x1` — write a multi-letter constant such as `alice` or `witness`, or `name{x}`), or instantiating two nested binders in a single step.

---

## `FOL existential introduction`

- **Schema:** `Φ(c)` ⊢ `some v is such that (Φ(v))` (where replacing `v` by `c` in `Φ(v)` yields `Φ(c)`, allowing selective replacement of some or all occurrences of `c`).
- **Valid form, refused by v44 strict grammar (`explicit_variable_fol`; lowercase constant or letter eigenvariable is not name evidence):**

```argcheck-refused explicit_variable_fol | bind variables inside is such that | FOL existential introduction
1. alice admires alice.
2. Therefore, some x is such that (x admires alice). [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL existential introduction
1. alice admires blair.
2. Therefore, every x is such that (x admires blair). [1]
```

- **What it does not cover:** Universal strengthening from a ground instance, introducing two existential binders in one step, or function symbols (`f(x)`).

---

## `FOL identity reflexivity`

- **Schema:** `Ψ(c)` ⊢ `c is identical to c` (when `c` occurs as a constant in the cited FOL premise).
- **Valid form, refused by v44 strict grammar (`identity.constant_sides`; lowercase constant or letter eigenvariable is not name evidence):**

```argcheck-refused identity.constant_sides | name the thing | FOL identity reflexivity
1. (every x is such that (x is red)) and (alice is red).
2. Therefore, alice is identical to alice. [1]
```

- **Rejected near-miss:**

```argcheck-reject FOL identity reflexivity
1. (every x is such that (x is red)) and (alice is red).
2. Therefore, blair is identical to blair. [1]
```

- **What it does not cover:** Reflexive identity for a constant absent from the cited step (unless inside an open `Let c be arbitrary.` subproof, where citing the declaration step supplies `c is identical to c`).

---

## `FOL identity substitution`

- **Schema:** `a is identical to b`; `Φ(a)` ⊢ `Φ(a/b)` (replacing one or more free occurrences of `a` with `b`, or `b` with `a`, including inside quantifier scopes), where every replaced occurrence is in a safe position (see "Identity substitution: safe positions" below).
- **Valid form, refused by v44 strict grammar (`identity.constant_sides`; lowercase constant or letter eigenvariable is not name evidence):**

```argcheck-refused identity.constant_sides | name the thing | FOL identity substitution
1. alice is identical to blair.
2. every x is such that (alice admires x).
3. Therefore, every x is such that (blair admires x). [1,2]
```

- **Rejected near-miss:**

```argcheck-reject FOL identity substitution
1. alice is identical to blair.
2. every x is such that (alice admires x).
3. Therefore, every x is such that (casey admires x). [1,2]
```

```argcheck-reject FOL identity substitution
1. alice is identical to blair.
2. every x is such that (x admires alice).
3. Therefore, every x is such that (x admires blair). [1,2]
```

- **What it does not cover:** Zero replacements (`Φ(a)` unchanged; use `FOL reiteration`), bare `"is"` instead of `is identical to`, or substituting non-identical constants.

### Identity substitution: safe positions (v35.0)

Identity substitution is certified only in known-safe positions, an allowlist
rather than a list of opaque verbs:

- in identity sentences (`A is identical to B`, so identity chains work);
- where every replaced occurrence is the grammatical subject of a plain clause:
  the start of the sentence, or of a clause after `if`, `then`, `and`, `or`,
  `it is (not) the case that`, `such that` or an opening parenthesis. A proper
  name stays substitutable as the subject of a modal sentence (`Superman can
  fly` gives `Clark can fly`).

Every other position is not certified: objects of verbs (`Lois is looking for
Superman`, `Lois admires Superman`, `Lois kissed Superman`), passives (`is
admired by`) and raising forms (`is believed to`, `is said to`), and anything
inside a that-clause. The message names the reason. If such a step is sound,
state it as a premise.

```argcheck-refused identity.constant_sides | name the thing | identity substitution
1. Superman can fly.
2. Superman is identical to Clark.
3. Therefore, Clark can fly. [1,2]
```

```argcheck-reject identity substitution
1. Lois is looking for Superman.
2. Superman is identical to Clark.
3. Therefore, Lois is looking for Clark. [1,2]
```

```argcheck-reject identity substitution
1. Superman is admired by Lois.
2. Superman is identical to Clark.
3. Therefore, Clark is admired by Lois. [1,2]
```

---

## `FOL conditional proof`

- **Schema:** Subproof `i..j` starting with `Suppose Φ.` and ending with `Ψ.` ⊢ `if (Φ) then (Ψ). [i-j]`.
- **Accepted:**

```argcheck-accept FOL conditional proof
| 1. Suppose every x is such that (x is red).
| 2. Therefore, alice is red. [1]
3. Therefore, if (every x is such that (x is red)) then (alice is red). [1-2]
```

- **Rejected near-miss:**

```argcheck-reject FOL conditional proof
| 1. Suppose every x is such that (x is red).
| 2. Therefore, alice is red. [1]
3. Therefore, if (every x is such that (x is red)) then (blair is red). [1-2]
```

- **What it does not cover:** Mismatched antecedent or consequent formulas, or extra cited steps alongside the `[i-j]` range.

---

## `FOL reductio`

- **Schema:** Subproof `i..j` starting with `Suppose Φ.` and ending with `contradiction.` ⊢ `it is not the case that (Φ). [i-j]` (and classical `Suppose it is not the case that (Φ).` ... `contradiction.` ⊢ `Φ. [i-j]`).
- **Accepted:**

```argcheck-accept FOL reductio
1. it is not the case that (every x is such that (x is red)).
| 2. Suppose every x is such that (x is red).
| 3. Therefore, contradiction. [1,2]
4. Therefore, it is not the case that (every x is such that (x is red)). [2-3]
```

- **Rejected near-miss:**

```argcheck-reject FOL reductio
| 1. Suppose every x is such that (x is red).
| 2. Therefore, alice is red. [1]
3. Therefore, it is not the case that (every x is such that (x is red)). [1-2]
```

- **What it does not cover:** Subproofs not ending in `contradiction`, or citing extra premises on the discharge line.

---

## `FOL universal generalization`

- **Schema:** Subproof `i..j` starting with unrestricted `Let c be arbitrary.` (for fresh `c`) and ending with `Φ(c).` ⊢ `every v is such that (Φ(v)). [i-j]`.
- **Accepted:**

```argcheck-accept FOL universal generalization
1. every x is such that (x is red).
| 2. Let witness be arbitrary.
| 3. Therefore, witness is red. [1]
4. Therefore, every x is such that (x is red). [2-3]
```

- **Rejected near-miss:**

```argcheck-reject FOL universal generalization
1. (every x is such that (x is blue)) and (witness is red).
| 2. Let witness be arbitrary.
| 3. Therefore, witness is red. [1]
4. Therefore, every x is such that (x is red). [2-3]
```

- **What it does not cover:** Non-fresh eigennames occurring in outer active steps or in the conclusion, or restricted `Let c be an arbitrary A.` declarations when discharging an unrestricted `every x is such that ...` quantifier.

---

## `FOL existential elimination`

- **Schema:** `some v is such that (Φ(v))`; subproof `i..j` starting with `Suppose Φ(c).` (for fresh `c`) and ending with `c`-free `Ψ.` ⊢ `Ψ. [k,i-j]`.
- **Valid form, refused by v44 strict grammar (`scope.clause_start`; lowercase constant or letter eigenvariable is not name evidence):**

```argcheck-refused scope.clause_start | move the operator into each clause | FOL existential elimination
1. some x is such that ((x is red) and (some y is such that (y is blue))).
| 2. Suppose (witness is red) and (some y is such that (y is blue)).
| 3. Therefore, some y is such that (y is blue). [2]
4. Therefore, some y is such that (y is blue). [1,2-3]
```

- **Rejected near-miss:**

```argcheck-reject FOL existential elimination
1. some x is such that (x is red).
| 2. Suppose witness is red.
| 3. Therefore, witness is red. [2]
4. Therefore, witness is red. [1,2-3]
```

- **What it does not cover:** Witness constant `c` escaping into the conclusion `Ψ`, reusing a constant already in scope, or assuming extra conjuncts about `c` not licensed by `Φ(v)`.

---

## `FOL proof by cases`

- **Schema:** `either (Φ) or (Ψ)`; subproof `i_1..j_1` (`Suppose Φ.` ... `Ω.`) and subproof `i_2..j_2` (`Suppose Ψ.` ... `Ω.`) ⊢ `Ω. [d,i_1-j_1,i_2-j_2]`.
- **Accepted:**

```argcheck-accept FOL proof by cases
1. either (every x is such that (x is red)) or (every x is such that (x is blue)).
2. if (every x is such that (x is red)) then (some y is such that (y is green)).
3. if (every x is such that (x is blue)) then (some y is such that (y is green)).
| 4. Suppose every x is such that (x is red).
| 5. Therefore, some y is such that (y is green). [2,4]
| 6. Suppose every x is such that (x is blue).
| 7. Therefore, some y is such that (y is green). [3,6]
8. Therefore, some y is such that (y is green). [1,4-5,6-7]
```

- **Rejected near-miss:**

```argcheck-reject FOL proof by cases
1. either (every x is such that (x is red)) or (every x is such that (x is blue)).
2. if (every x is such that (x is red)) then (some y is such that (y is green)).
3. if (every x is such that (x is blue)) then (some y is such that (y is yellow)).
| 4. Suppose every x is such that (x is red).
| 5. Therefore, some y is such that (y is green). [2,4]
| 6. Suppose every x is such that (x is blue).
| 7. Therefore, some y is such that (y is yellow). [3,6]
8. Therefore, some y is such that (y is green). [1,4-5,6-7]
```

- **What it does not cover:** Branches ending in non-equivalent formulas, or three-way disjunctions without nested binary `either (...) or (...)` elimination.

---

## `notation equivalence`

- **Schema:** `S_readable` ⊣⊢ `S_FOL` (one-step bidirectional conversion between a single-reading controlled-English sentence and its canonical explicit-variable FOL rendering).
- **Accepted:**

```argcheck-accept notation equivalence
1. Every student is red.
2. Therefore, every x is such that (if (x is student) then (x is red)). [1]
```

- **Rejected near-miss:**

```argcheck-reject notation equivalence
1. Every student admires some teacher.
2. Therefore, every x is such that (if (x is student) then (some y is such that ((y is teacher) and (x admires y)))). [1]
```

- **What it does not cover:** Multi-reading scope-ambiguous sentences (`Every student admires some teacher`, `Every student is red and alice is blue` — disambiguate with `is such that` first), or modal, counterfactual, and typed-property sentences outside FOL.

---

## `counting definition`

- **Schema:** Definitional equivalences and direct FOL consequences of bounded cardinality expansions (`1`–`5` / `one`–`five`, `at least`, `at most`, `exactly`, `only a is F`, and `no F other than a ...`).
- **Accepted:**

```argcheck-accept counting definition
1. At least one lab is red.
2. Therefore, at least 1 lab is red. [1]
```

- **Rejected near-miss:**

```argcheck-reject counting definition
1. At least one lab is red.
2. Therefore, at most one lab is red. [1]
```

- **What it does not cover:** Counts outside `1`–`5` (`zero`, `0`, `6`, `six` and higher are rejected with a bound diagnostic), proportional quantifiers (`most`, `many`, `few`), or `at least n` entailing `at most n`.

---

## `counting introduction`

- **Schema:** `At least n F ...`; `At most n F ...` ⊢ `Exactly n F ...` (`1 <= n <= 5`, over matching counted predicates).
- **Accepted:**

```argcheck-accept counting introduction
1. At least 1 lab that visits Blair admires Casey.
2. At most 1 lab that visits Blair admires Casey.
3. Therefore, exactly 1 lab that admires Casey visits Blair. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject counting introduction
1. At least 1 lab is red.
2. At most 1 lab is red.
3. Therefore, exactly 2 labs are red. [1,2]
```

- **What it does not cover:** Mismatched bounds (`1` vs `2`), counts above `5`, or non-equivalent restrictor/predicate clauses. (Note: when the premise and conclusion bodies are already syntactically identical without conjunct reordering, the step may be certified directly as `counting definition`.)

---

## `counting elimination`

- **Schema:** `Exactly n F ...` ⊢ `At least n F ...` (or `At most n F ...`, for `1 <= n <= 5`).
- **Accepted:**

```argcheck-accept counting elimination
1. Exactly 1 lab that visits Blair admires Casey.
2. Therefore, at least 1 lab that admires Casey visits Blair. [1]
```

- **Rejected near-miss:**

```argcheck-reject counting elimination
1. Exactly 2 labs are red.
2. Therefore, at least 3 labs are red. [1]
```

- **What it does not cover:** Strengthening `Exactly n` to `At least (n+1)` or `At most (n-1)`, or counts above `5`. (As with `counting introduction`, identical-body projections may be reported as `counting definition`.)

---

## `counting contradiction`

- **Schema:** `At least m F ...`; `At most n F ...` ⊢ `contradiction` (where `1 <= n < m <= 5`).
- **Accepted:**

```argcheck-accept counting contradiction
1. At least 4 labs are red.
2. At most 2 labs are red.
3. Therefore, contradiction. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject counting contradiction
1. At least 1 lab is red.
2. At most 1 lab is red.
3. Therefore, contradiction. [1,2]
```

- **What it does not cover:** Compatible bounds (`m <= n`), mismatched counted predicates, or counts above `5`. (Note: adjacent `m = n + 1` clashes with identical bodies may be reported as `counting definition` because `At most n` expands directly to the negation of `At least (n+1)`.)

---

## `counting uniqueness`

- **Schema:** `Exactly one F P` (or `At most one F P`); `a is an F and a P` (or separate cited ground premises `a is an F`; `a P`) ⊢ `No F other than a P`.
- **Accepted:**

```argcheck-accept counting uniqueness
1. Exactly one student admires Blair.
2. Ada is a student and Ada admires Blair.
3. Therefore, no student other than Ada admires Blair. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject counting uniqueness
1. Exactly one person wrote Waverley.
2. Scott wrote Waverley.
3. Therefore, no person other than Scott wrote Waverley. [1,2]
```

- **What it does not cover:** Omitting the class-membership premise (`Scott is a person` is required alongside `Scott wrote Waverley`; otherwise `Scott` could be a non-person while some person also wrote Waverley), or counts greater than `1`.

---

## `counting subset`

- **Schema:** `At least m F R`; `At most n F that R is P` ⊢ `Some F that R is not P` (pigeonhole / subset difference for `1 <= n < m <= 5`).
- **Valid form, refused by v44 strict grammar (`eg.name_witness`):**

```argcheck-refused eg.name_witness | state the existential claim as its own premise | counting subset
1. At least two labs signed the pledge.
2. At most one lab that signed the pledge is red.
3. Therefore, some lab that signed the pledge is not red. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject counting subset
1. At least two labs signed the pledge.
2. At most two labs that signed the pledge are red.
3. Therefore, some lab that signed the pledge is not red. [1,2]
```

- **What it does not cover:** `m <= n` (where every `F that R` could still be `P`), mismatched relative clauses, or counts above `5`.

---

## `description definition`

- **Schema:** Russellian expansion equivalences and direct FOL consequences of `The G P` (`some x (G(x) and P(x)) and at most 1 x G(x)`).
- **Accepted:**

```argcheck-accept description definition
1. The lab is red.
2. Therefore, some lab is red. [1]
```

- **Rejected near-miss:**

```argcheck-reject description definition
1. The author of Waverley wrote Ivanhoe.
2. Therefore, Scott wrote Ivanhoe. [1]
```

- **What it does not cover:** Naming a specific individual without an identity premise (`The author of Waverley is Scott`), non-verbatim restrictor rewrites (`who` vs `that`), or descriptions inside attitude, modal, or deontic operators.

---

## `description substitution`

- **Schema:** `The G is a`; `The G P` ⊢ `P(a)` (Russellian definite-description substitution for capitalized constant `a` and verbatim restrictor `G`).
- **Valid form, refused by v44 strict grammar (`identity.extensional_context`):**

```argcheck-refused identity.extensional_context | state the substituted claim as a premise | description substitution
1. The author of Waverley is Scott.
2. The author of Waverley wrote Ivanhoe.
3. Therefore, Scott wrote Ivanhoe. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject description substitution
1. An author of Waverley is Scott.
2. Some author of Waverley wrote Ivanhoe.
3. Therefore, Scott wrote Ivanhoe. [1,2]
```

- **What it does not cover:** Indefinite descriptions (`An author ...`), relative-pronoun changes between the two `The ...` phrases (`The courier who ...` vs `The courier that ...`), extra unrelated citations, or opaque contexts (`Ann believes that ...`, `It is necessary that ...`, `It is obligatory that ...`).

---

## `description introduction`

- **Schema:** `Some G P` + `At most one G` ⊢ `The G P`, or `a is a G` + `P(a)` + `At most one G` (or `Exactly one G`) ⊢ `The G P`.
- **Accepted:**

```argcheck-accept description introduction
1. Ada is a lab.
2. Ada is red.
3. At most one lab.
4. Therefore, the lab is red. [1,2,3]
```

- **Rejected near-miss:**

```argcheck-reject description introduction
1. Ada is a lab.
2. At most one lab.
3. Therefore, the lab is red. [1,2]
```

- **What it does not cover:** Missing uniqueness (`At most one G` / `Exactly one G`), missing existence or predication (`Ada is red`), or uncapitalized witness constants where a proper name is required.

---

## `description elimination`

- **Schema:** `The G is a` ⊢ `a is a G`, or `The G P` ⊢ `Some G` / `Exactly one G`.
- **Valid form, refused by v44 strict grammar (`identity.extensional_context`):**

```argcheck-refused identity.extensional_context | state the substituted claim as a premise | description elimination
1. The author of Waverley is Scott.
2. Therefore, Scott is an author of Waverley. [1]
```

- **Rejected near-miss:**

```argcheck-reject description elimination
1. At most one author of Waverley.
2. Scott wrote Ivanhoe.
3. Therefore, Scott is an author of Waverley. [1,2]
```

- **What it does not cover:** Concluding membership from `At most one G` without existence, or inferring wide-scope negated existence from an ambiguous `The G is not P` premise.

---

## `description uniqueness`

- **Schema:** `The G is a`; `b is a G` ⊢ `b is a` (identifying any member of `G` with the referent of `The G`).
- **Valid form, refused by v44 strict grammar (`identity.constant_sides`):**

```argcheck-refused identity.constant_sides | name the thing | description uniqueness
1. The author of Waverley is Scott.
2. Blair is an author of Waverley.
3. Therefore, Blair is Scott. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject description uniqueness
1. An author of Waverley is Scott.
2. Blair is an author of Waverley.
3. Therefore, Blair is Scott. [1,2]
```

- **What it does not cover:** Indefinite `An G is a`, mismatched restrictors (`author of Waverley` vs `author of Ivanhoe`), or uncapitalized predicates mistaken for proper names in `b is a`.

---

## `description negation`

- **Schema:** Scope-safe reasoning with negated definite descriptions, such as `The G is a`; `It is not the case that the G P` (or `The G is not P`) ⊢ `It is not the case that P(a)` (or `a is not P`).
- **Valid form, refused by v44 strict grammar (`identity.extensional_context`):**

```argcheck-refused identity.extensional_context | state the substituted claim as a premise | description negation
1. The author of Waverley is Scott.
2. It is not the case that the author of Waverley is red.
3. Therefore, it is not the case that Scott is red. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject description negation
1. The present king of France is not bald.
2. Therefore, some present king of France is not bald. [1]
```

- **What it does not cover:** Inferring existence (`Some G is not P`) from surface `The G is not P` (which has both a wide-scope Russellian reading and a narrow-scope `It is not the case that the G is P` reading that holds when no `G` exists), or inferring wide-scope `The G is not P` from narrow-scope `It is not the case that the G is P` plus `At most one G`.

---

## `unless elimination`

- **Schema:** `P unless Q` (or `Unless Q, P`); `It is not the case that Q` ⊢ `P`. "Unless" has its weak reading, `if it is not the case that Q, then P` (equivalently `P or Q`).
- **Accepted:**

```argcheck-accept unless elimination
1. The alarm sounds unless the battery is dead.
2. It is not the case that the battery is dead.
3. Therefore, the alarm sounds. [1,2]
```

- **Rejected near-miss:**

```argcheck-reject unless elimination
1. The alarm sounds unless the battery is dead.
2. The battery is dead.
3. Therefore, it is not the case that the alarm sounds. [1,2]
```

- **What it does not cover:** The biconditional or exclusive reading (`P unless Q`, `Q` ⊢ `not P`; `P unless Q`, `P` ⊢ `not Q`), an unless-clause containing a personal pronoun (`Every student passes unless he fails the exam`, whose pronoun may point back into the main clause), or a compound unless-clause matched against one conjunct.

---

## `unless conversion`

- **Schema:** `P unless Q` ⊣⊢ `If it is not the case that Q, then P` (either direction).
- **Accepted:**

```argcheck-accept unless conversion
1. The alarm sounds unless the battery is dead.
2. Therefore, if it is not the case that the battery is dead, then the alarm sounds. [1]
```

- **Rejected near-miss:**

```argcheck-reject unless conversion
1. The alarm sounds unless the battery is dead.
2. Therefore, if the battery is dead, then it is not the case that the alarm sounds. [1]
```

- **What it does not cover:** The converse conditional `if Q, then not P` (the exclusive reading), or an unless-clause containing a personal pronoun.

## Grouping conjunctions and disjunctions

A clause cannot mix top-level `and` and `or` without visible scope. This is a
syntax failure, before entailment checking, on premises and conclusions alike,
including conditional antecedents/consequents and explicit quantified scopes.
The diagnostic names the line and quotes the original sentence, then says
`use either...or / both...and, a comma, or parentheses to group`.

Use whole-clause parentheses: `(Ann sings and Bob dances) or Cal reads` is a
disjunction; `Ann sings and (Bob dances or Cal reads)` is a conjunction.
Parentheses must balance and remain visible in statement echoes and mismatch
excerpts. Square brackets such as `[1]` still mean step references, not grouping.

A single comma immediately before the main connective also fixes scope:
`Ann sings and Bob dances, or Cal reads` means `(Ann sings and Bob dances) or
Cal reads`; `Ann sings, and Bob dances or Cal reads` means `Ann sings and
(Bob dances or Cal reads)`. Multiple commas in a mixed clause, including lists
such as `Ann sings, Bob dances and Cal reads, or Dan sleeps`, are ambiguous;
use parentheses instead. Plain lists containing only one kind of connective
retain their existing behavior, including `Ann sings, Bob dances, and Cal reads`
and `Ann sings, Bob dances, or Cal reads`.

A leading `either` or `both` opens the first operand of its connective, so
`Either Ann sings and Bob dances or Cal reads` is a disjunction whose first
disjunct is `Ann sings and Bob dances`. A single comma still marks the main
connective: `Either Ann sings and Bob dances, or Cal reads` is a disjunction,
and so is `Both Ann sings and Bob dances, or Cal reads` (its left disjunct is
`both Ann sings and Bob dances`). `Either Ann sings, or Bob dances and Cal reads`
is a disjunction with `Bob dances and Cal reads` as its right disjunct. With no
comma, the either/both connective is the main one and its second operand runs
to the end: `Either Ann sings or Bob dances and Cal reads` means `Ann sings or
(Bob dances and Cal reads)`, and `Both Ann sings and Bob dances or Cal reads`
means `Ann sings and (Bob dances or Cal reads)`. Parenthesize
operands if a longer scaffold has further mixed connectives: each operand must
itself have unambiguous scope.

```argcheck-accept conjunction elimination
1. Ann sings and (Bob dances or Cal reads).
2. Therefore, Ann sings. [1]
```

```argcheck-accept conjunction elimination
1. Ann sings, and Bob dances or Cal reads.
2. Therefore, Ann sings. [1]
```

```argcheck-accept disjunctive syllogism
1. (Ann sings and Bob dances) or Cal reads.
2. It is not the case that Ann sings and Bob dances.
3. Therefore, Cal reads. [1,2]
```

```argcheck-reject conjunction elimination
1. Ann sings and Bob dances or Cal reads.
2. Therefore, Ann sings. [1]
```

```argcheck-reject conjunction elimination
1. Ann sings and Bob dances, or Cal reads.
2. Therefore, Ann sings. [1]
```

```argcheck-reject conjunction elimination
1. Both Ann sings and Bob dances, or Cal reads.
2. Therefore, Ann sings. [1]
```

```argcheck-reject conjunction elimination
1. Ann sings, and Bob dances, or Cal reads.
2. Therefore, Ann sings. [1]
```

Leading attitudes, impersonal `It is <adjective> that` reports, and sentence
adverbs now require explicit grouping when followed by whole-clause
coordination. For example, `Perhaps A and B` must be written `(Perhaps A) and B`
or `Perhaps (A and B)`, or with explicit `both...and` / `either...or`.
Since v44.0, relative clauses, quotations, lowercase names and single-letter terms cannot exempt a containing mixed connective sentence from the final extraction guard. An unchanged whole atom can still be reused. Explicit parentheses or both/either scaffolds must establish the grouping before its parts can be inspected.

Existing parenthetical identifiers such as `definition (a)`, `condition (b)`,
and `proposition (e)` remain atomic labels, not logical grouping. Parenthetical annotations
inside atomic clauses, such as `significant (p < 0.01)`, retain their old opaque
reading. Parentheses around words, such as `Ann (sings)`, do not create a
logical group. Only parentheses at whole-clause boundaries are interpreted as
logical grouping; all parentheses must still balance.

## English extraction and matching boundaries

Schematic X/Y/Z slots and universal-instantiation subjects cannot swallow a
clause-embedding operator. Existential witnesses have the same embedding guard.
Attitude/report verbs, including complements without `that`, impersonal
`It is <adjective> that` clauses, and leading sentence adverbs are guarded by
lexical recognition. Nominal relative clauses retain their internal negation;
whole attitude reports remain usable as atomic premises. This is a conservative
surface register, not a complete English noun-phrase parser.

Schematic bindings and universal-instantiation subjects reject `exactly N`,
`only`, `all but`, `almost every/all`, `nearly`, `at most`, `fewer than`, `few`,
`no`, `none`, and `neither`. Upward determiner forms remain available, including
`every`, `each`, `all`, `some`, `a`, `an`, `the`, `most`, `many`, `several`,
`more than N`, and `at least N`. Existential generalization has its own existence
requirements; for example, the accepted `almost every` witness form is preserved.

Plural matching does not strip suffixes from `clothes`, `species`, `means`,
`goods`, `news`, `odds`, `thanks`, `series`, `savings`, `earnings`, `belongings`,
`remains`, `surroundings`, `axes`, `arms`, `customs`, or `glasses`. These words
still match themselves. Ordinary `premises` / `premise` matching is unchanged.

Identity substitution also refuses intensional adverbs (`allegedly`,
`supposedly`, `reportedly`), attitude complements such as `dreams about`, and
naming predicates (`is called`, `is named`, `is known as`). These predications
must not acquire substitutable subject terms during readable FOL conversion.
Extensional predicates such as `painted`, `near`, `similar to`, `was`,
`almost red`, and `fake robot` retain their previous behavior.

A comma before `which`, `who`, `whom` or `whose` (a non-restrictive relative)
is significant for restatement and identity matching. Other commas are not,
so verb-list punctuation (`eats, shoots, and leaves`) is a known gap.
Non-restrictive to restrictive relative-clause weakening is directional, not
clause identity.

```argcheck-reject schematic instantiation
1. If X is red, then X is warm.
2. The observer imagines Alpha is red.
3. Therefore, The observer imagines Alpha is warm. [1,2]
```

```argcheck-refused scope.clause_start | move the operator into each clause | notation equivalence
1. Alpha eats, shoots, and leaves.
2. Therefore, Alpha eats, shoots and leaves. [1]
```

```argcheck-accept modus ponens
1. The ground is wet if it rains.
2. It rains.
3. Therefore, The ground is wet. [1,2]
```

Leading concessive `Even if A, then B` supplies the ordinary conditional
commitment `If A, then B`, so modus ponens is accepted. This does not extend the
postposed `B even if A` register. `Only if A, then B` and `Unless A, then B`
still do not license the forward step from A to B.

### Accuracy boundaries for term instantiation and unless grouping

A schema letter immediately following a capitalized name component (for example,
`Plan X`) is literal. A conditional without real schema variables remains an
ordinary conditional. `If X is a man, then X is mortal` and `the theory of X`
retain their schematic readings.

Universal instantiation requires a whole top-level copula instance, with a proper
name, definite description (`the N ...`), or plain indefinite (`a/an N ...`)
subject (including single-letter arbitrary names in Fitch proofs). It cannot rewrite a copula embedded in a conditional, biconditional,
conjunction, or negation. Quantified phrases such as `almost no person`,
`just one senator`, `one of the planets`, and `someone in Athens` are not
individual terms or bare universal classes. Schematic slots also reject
`almost no`, `very few`, `just one`, `only one`, `one of`, `someone`, `something`,
`somebody`, and quantifying `few`. Upward-monotone `most` remains available in
schematic instantiation, but not as a UI subject. The existing upward-monotone
`most` and `more than` inferences from an every/all class inclusion are also
certified under the schematic form, preserving the entire quantified subject.
Long nominal and gerund bindings
remain supported. Plural matching preserves singular `princess`, `corps`,
`class`, `bus`, `glass`, `chaos`, `lens`, `series`, and `species`.

Mixing top-level `unless` with clause-level `and` or `or` requires parentheses,
a comma before the main connective, or `both`/`either` grouping. The syntax
reason is exactly `mixes top-level unless with and/or without explicit grouping`.
The existing `Either A or B and C` reading remains `A or (B and C)`.

```argcheck-accept unless conversion
1. (The alarm sounds and the lamp flashes) unless the battery is dead.
2. Therefore, if it is not the case that the battery is dead, then the alarm sounds and the lamp flashes. [1]
```

```argcheck-accept unless conversion
1. The alarm sounds and the lamp flashes, unless the battery is dead.
2. Therefore, if it is not the case that the battery is dead, then the alarm sounds and the lamp flashes. [1]
```

```argcheck-reject unless conversion
1. The alarm sounds and the lamp flashes unless the battery is dead.
2. Therefore, if it is not the case that the battery is dead, then the alarm sounds and the lamp flashes. [1]
```

## Sentence-internal frames

A certificate that extracts or builds a conjunction/disjunction, substitutes
identicals, or instantiates a quantifier or schema refuses a sentence that opens
with something that may cover the whole sentence: a sentence adverb (any word
ending in `-ly`, such as `Arguably`, `Purportedly`, `Hypothetically`, and
`perhaps`, `maybe`, `sometimes`, `often`, `once`, `always`, `never`,
`yesterday`, `today`), or a fronted frame that starts with any preposition, single-word or complex
(`In the story`, `Under the law of likelihood`, `According to Bob`, `As per the
log`, `In view of ...`), with or without a comma. A capitalized word that
also appears capitalized mid-sentence elsewhere in the argument is a name, not
an adverb under the older recognizer. Since v44.0, initial -ing words are
refused even with a gerund-subject reading, and inspected clauses must use
controlled clause starts. The message suggests moving the adverb or
frame out, or rewording so the sentence starts with its subject (`Rationalist
notions that are deeply held ...` instead of `Deeply held rationalist notions
...`).
Whole-sentence reuse does not inspect the reused sentence.

```argcheck-reject conjunction elimination
1. Arguably the server crashed and the network failed.
2. Therefore, the network failed. [1]
```

Unmarked clause coordination requires a recognized complete left clause.
Copular predications and listed intransitive predicates support ordinary
`A and B`; other verbs require explicit whole-clause grouping. Parentheses
inside the complement of a verb do not assert that complement. Write
`Both ... and ...`, `Either ... or ...`, parenthesize whole operands, or split
assertions into numbered premises. A coordinated noun or verb phrase does not
by itself license clause extraction.

```argcheck-accept conjunction elimination
1. The server crashed and the network failed.
2. Therefore, the network failed. [1]
```

```argcheck-reject conjunction elimination
1. Bob reckons the server crashed and the network failed.
2. Therefore, the network failed. [1]
```

Commas in clause and term comparisons are significant. A structural parser may
consume the comma in `If A, then B`, a conclusion marker such as `Therefore,`,
or an explicitly recognized coordination boundary. Other commas, including
relative, appositive and vocative commas, remain part of the compared text.
Whitespace and a final period do not distinguish clauses.

```argcheck-reject self-entailment
1. Alpha eats shoots and leaves.
2. Therefore, Alpha eats, shoots, and leaves. [1]
```

## Strictness notes

A failed step can include up to three structured notes. A note appears only
when retrying that step with the named strictness rule disabled, and all other
rules enabled, certifies it. Rules that block a step together may produce no
note. Notes never change the verdict or the existing mismatch explanation.
Passing steps have no notes. A suggested new premise supplies a claim rather
than proving it; replacing a quantified phrase with a name changes the claim.

The table is generated from `validation/strict_notes.go`. Regenerate it with
`go test ./validation -run '^TestDocsRulesStrictNotes$' -update-strict-notes`.
The same test without the update flag fails if the table differs from the Go
data. `TestStrictNotesExamples` checks every example and its concrete fix.

<!-- strictness-notes:start -->
| rule id | name | since | why | also refuses | fix |
| --- | --- | --- | --- | --- | --- |
| `opener.frame` | Fronted frame | v36.0 | A phrase before the subject may scope over the whole sentence: 'According to the report, A and B' attributes both claims, so A alone is not certified. | 1. If X is registered, then X is eligible.<br>2. In 2019 the firm is registered.<br>3. Therefore, in 2019 the firm is eligible. [1,2] | Move the time phrase after the predicate, including in the schema: "If X is registered in 2019, then X is eligible in 2019." |
| `opener.adverb` | Fronted adverb | v37.0 | A sentence adverb may weaken or qualify the whole claim: 'Arguably A and B' may assert only that A and B is arguable. | 1. If X is safe, then X is useful.<br>2. Highly reliable machinery is safe.<br>3. Therefore, highly reliable machinery is useful. [1,2] | Start with the subject: "The reliable machine is safe." Use the same wording in the conclusion. |
| `coordination.clause` | Clause coordination after a verb | v36.0 | In 'The observer saw A and B left', 'and' may join two objects or two whole clauses, so splitting off 'B left' is not certified. | 1. The pump malfunctioned.<br>2. The valve stuck.<br>3. Therefore, the pump malfunctioned and the valve stuck. [1,2] | Write "Both the pump malfunctioned and the valve stuck." |
| `comma.exact` | Commas must match | v36.0 | A comma can change grouping ('A, and B or C' differs from 'A and B or C'), so clauses must repeat their punctuation exactly. | 1. Every device is small, light and portable.<br>2. Alpha is a device.<br>3. Therefore, Alpha is small, light, and portable. [1,2] | Repeat the clause with exactly the same commas wherever it is reused. |
| `identity.safe_position` | Identity substitution only in subjects | v35.0 | Substituting co-referring names inside an attitude or other opaque context can change truth. The catalog therefore permits identity substitution only in identity sentences and plain clause subjects. | 1. Alpha is identical to Beta.<br>2. The robot kicks Alpha.<br>3. Therefore, the robot kicks Beta. [1,2] | State the substituted sentence as a premise: "The robot kicks Beta." This supplies the missing claim; it does not prove it from the original premises. |
| `slot.indefinite_object` | Indefinite in an object position | v39.0 | An indefinite object need not pick anything out: 'The robot is looking for a unicorn' does not establish that a unicorn exists. | 1. If the robot admires X, then X is admirable.<br>2. The robot admires a hero.<br>3. Therefore, a hero is admirable. [1,2] | Use a name in the slot: "The robot admires Alpha." Conclude "Alpha is admirable." This changes the premise to identify a particular hero. |
| `slot.indefinite_written` | Indefinite written before a variable | v43.0 | When the rule writes 'a X', the instance need not be an article plus a noun: 'Bob greets a lot of people' would make X 'lot of people'. The catalog has no lexicon to tell that from 'a dog', so such a variable is refused in every position. | 1. If Rex is a X, then Fido is a X.<br>2. Rex is a dog.<br>3. Therefore, Fido is a dog. [1,2] | Write the phrase itself into the rule ("If Rex is a dog, then Fido is a dog"), or use a name or singular definite description. |
| `slot.indefinite_value` | Indefinite value that need not name one thing | v43.0 | The v43.0 shape check requires a lone indefinite pronoun or a/an, some or one plus modifiers and one noun with nothing after it. A quantity or complement phrase ('a lot of people', 'a friend of Bob', 'a few runners') need not be one thing. Since v44.0, slot.value_singular also refuses the remaining indefinite values. | 1. If X arrived, then X is someone who arrived.<br>2. A friend of Bob arrived.<br>3. Therefore, a friend of Bob is someone who arrived. [1,2] | Write the phrase itself into the rule: "If a lot of people arrived, then a lot of people is someone who arrived." The rule then claims exactly what it says; no variable is filled with the phrase. |
| `slot.value_singular` | Schematic value must be a singular term | v44.0 | A quantity, plural, group or indefinite phrase (a lot, most people, Bob and Carol, a dog) need not be one thing, so a rule about each thing does not apply to it. Only evidenced proper names and singular definite descriptions fill term slots. Predicate-only positions after a copula accept any value; variables also used as terms still require singular values. | 1. If X arrived, then X was seen.<br>2. A dog arrived.<br>3. Therefore, a dog was seen. [1,2] | Name the thing (Rex arrived), with internal capitals or capitalization evidenced somewhere other than sentence start or after Therefore/So/Thus/Hence; or use a singular definite description (the dog). Alternatively state existence as its own premise and then name it, or write the phrase itself into the rule instead of a variable. |
| `slot.modified_quantifier` | Modified quantifier in a term slot | v36.0 | A modified quantifier ('more than two robots', 'nearly all firms') is not an individual, so it cannot fill X or instantiate a universal. | 1. If X is red, then X is warm.<br>2. Nearly every robot is red.<br>3. Therefore, nearly every robot is warm. [1,2] | Use a singular definite description: "The robot is red." Conclude "The robot is warm." This certifies an individual instance rather than the quantified claim. |
| `slot.quantity` | Quantity expression in a term slot | v38.0 | A quantity phrase ('two robots', 'a minority of voters', 'zero cases') may mean exactly that many, so it is not an individual for X or for instantiation. Since v43.0 the rule refuses only readings that are ambiguous ('exactly two'); collective quantity phrases such as 'a group of robots' are refused by slot.indefinite_value instead, so no step valid on every reading is refused by this rule alone. |  | Name the group: "The team is registered." Conclude "The team is eligible." |
| `term.clause_inside` | Clause inside a term | v37.0 | A term that contains a finite verb with a complement may hide an attitude or embedded claim: 'the observer reckons Alpha is red' does not give that Alpha is red. | 1. If X are useful, then X are valuable.<br>2. Software exploits not known to humans are useful.<br>3. Therefore, software exploits not known to humans are valuable. [1,2] | Name the thing: "The exploit is useful." Conclude "The exploit is valuable." |
| `noun.pluralia_tantum` | Plural-only noun | v37.0 | Some -s nouns are not plurals of the stripped word ('tongs' is not more than one 'tong'), so singular/plural matching is refused for them. | 1. All glasses are fragile.<br>2. The cup is a glass.<br>3. Therefore, the cup is fragile. [1,2] | Use the singular noun consistently: "Every glass is fragile." |
| `conditional.postposed_if` | Postposed 'if' after a verb or adjective | v37.0 | A verb or adjective before 'if' can introduce a whether-question rather than a condition; postposed conditionals are recognized only in conservative surface forms. | 1. The machine vibrates if the motor runs.<br>2. The motor runs.<br>3. Therefore, the machine vibrates. [1,2] | Write "If the motor runs, then the machine vibrates." |
| `relation.object_extensional` | Instantiation into an object position | v38.0 | The object of an unlisted verb may be intensional: awaiting every guest does not license the same substitutions as greeting every guest. Object instantiation is restricted to a closed verb list. | 1. The robot admires every hero.<br>2. Alpha is a hero.<br>3. Therefore, the robot admires Alpha. [1,2] | State the instance as a premise: "The robot admires Alpha." This supplies the missing claim; it does not prove it from the original premises. |
| `unless.anaphor` | Pronoun across 'unless' | v38.0 | A pronoun after 'unless' may point back into the main clause, so 'A unless it is dead' cannot be split into independent clauses. | 1. The alarm sounds unless this battery is dead.<br>2. It is not the case that this battery is dead.<br>3. Therefore, the alarm sounds. [1,2] | Replace "this battery" with "the battery" in both premises. |
| `slot.description_shape` | Controlled definite descriptions | v45.0 | Description modifiers must be recognized adjectives or closed-list modifiers; names and finite clauses are refused. | 1. If X is red, then X is warm.<br>2. The Bob dog is red.<br>3. Therefore, the Bob dog is warm. [1,2] | Name it, or use 'the dog'. |
| `slot.one_type` | One type per rule variable | v44.0 | A clause-slot occurrence must not smuggle a clause into another occurrence used as a term. | 1. If when X is red bob smiles, then X is warm.<br>2. When the dog is red bob smiles.<br>3. Therefore, the dog is warm. [1,2] | Name the thing in every occurrence, or write the phrase into the rule. |
| `eg.subject_only` | Subject witnesses only | v44.0 | An object or passive subject may occur in an intensional position without asserting its existence. | 1. Lois greets Zeus.<br>2. Therefore, there is an X such that Lois greets X. [1] | State the existential claim as its own premise. |
| `eg.name_witness` | Named existential witnesses | v44.0 | A definite may covary with another quantifier; descriptions do not fix one witness. | 1. The dog is red.<br>2. Therefore, something is red. [1] | State the existential claim as its own premise, or use an evidenced proper name in an unquantified subject clause. |
| `term.numeral` | No numeral-led terms | v44.0 | A numeral phrase can state an exact quantity rather than identify one constant. | 1. All rodents are animals.<br>2. Six mice are rodents.<br>3. Therefore, six mice are animals. [1,2] | Name the thing, or use 'the mouse'. |
| `identity.constant_sides` | Constant identity sides | v44.0 | Pronouns can shift referent and quantities need not identify one object. | 1. He is identical to Alpha.<br>2. He is identical to Beta.<br>3. Therefore, Alpha is identical to Beta. [1,2] | Name the thing, or use 'the dog' on each side. |
| `identity.extensional_context` | Closed extensional identity contexts | v44.0 | An unlisted predicate may describe an intensional role under which co-referring names cannot be exchanged. | 1. The man is identical to Clark.<br>2. The man is famous as a hero.<br>3. Therefore, Clark is famous as a hero. [1,2] | State the substituted claim as a premise. |
| `slot.letter_in_name` | Letters adjacent to names | v44.0 | A schematic letter may be a literal component of Model U or Malcolm X. | 1. If Model U is red, then Model U is warm.<br>2. Model U is red.<br>3. Therefore, Model U is warm. [1,2] | Use a variable away from names, or write the name out. |
| `term.inflection_verbatim` | Verbatim noun inflection | v44.0 | A plural surface may not denote instances of the stem guessed by deleting its suffix. | 1. All dogs are wet.<br>2. The dog is a dog.<br>3. Therefore, the dog is wet. [1,2] | Write the rule with singular 'every N' (Every dog is wet). |
| `scope.mixed_andor` | Mixed connective scope | v44.0 | A and B or C may mean A and (B or C) or (A and B) or C; relative clauses and lowercase names do not disambiguate it. | 1. Ann is a singer who dances and Bob is happy or Cal reads.<br>2. Therefore, Ann is a singer who dances. [1] | Use 'both … and' or 'either … or' to show the grouping. |
| `scope.clause_start` | Controlled clause starts | v44.0 | Each inspected clause must start with a determiner, proper name, pronoun or numeral. An unrecognized start may be an operator with wider scope. | 1. It is not the case that perhaps the lamp is lit and the door is open.<br>2. Therefore, either it is not the case that perhaps the lamp is lit or it is not the case that the door is open. [1] | Move the operator into each clause, or state the claim as an explicit premise. |
| `opener.ing_frame` | Initial -ing frame | v44.0 | Following the script may qualify the entire conjunction rather than name its subject. Even genuine gerund subjects are refused when inspected. | 1. It is not the case that it is not the case that swimming is healthy.<br>2. Therefore, swimming is healthy. [1] | Move the frame into each clause, or state the framed claim as its own premise. |
| `scope.elided_copula` | Closed copula elision | v44.0 | In Either the men are quiet or fish, fish may be a verb rather than a predicate with an omitted copula. Unknown predicates are not expanded. | 1. Either the men are quiet or fish.<br>2. Therefore, either the men are quiet or the men are fish. [1] | Repeat the subject and verb: 'or the men are fish'. |
| `scope.description_negation` | Description negation scope | v44.0 | Wide negation need not assert that a description denotes; narrow predicate negation can presuppose existence. Only proper names permit the conversion. | 1. It is not the case that the present king of France is bald.<br>2. Therefore, the present king of France is not bald. [1] | Add 'There is exactly one present king of France' and name it, or keep the wide negation. |
| `opener.ly_exception` | Adverb-capable noun exceptions | v44.0 | Daily, early and only can be adverbs with scope over an entire sentence, even when they also have nominal uses. | 1. It is not the case that it is not the case that daily the lamp is lit.<br>2. Therefore, daily the lamp is lit. [1] | Move the adverb into each clause, or state the framed claim as its own premise. |
<!-- strictness-notes:end -->

## v44.0: singular schematic term values

`slot.value_singular` is enabled by default. Term slots admit only evidenced
proper names (1–4 tokens) or singular definite descriptions without
coordination or relative clauses; the bounded “the N of D” shape below is allowed. Indefinites are refused even
when they have the v43.0 article-plus-noun shape. Pure whole-clause variables retain their matching behavior; predicate positions
after a copula take any value. Variables also used as terms remain constrained. See the strict note above and the v44.0 coverage entry.

Partial noun-head slots such as `the X` remain refused when X alone is not
an allowed term; put the entire description in X or write it literally.

### v44.0 strict term and binding rules

- `slot.description_shape`: Description modifiers must be recognized adjectives or closed-list modifiers; names and finite clauses are refused. Name it, or use 'the dog'.
- `slot.one_type`: A clause-slot occurrence must not smuggle a clause into another occurrence used as a term. Name the thing in every occurrence, or write the phrase into the rule.
- `eg.subject_only`: An object or passive subject may occur in an intensional position without asserting its existence. State the existential claim as its own premise.
- `eg.name_witness`: A definite may covary with another quantifier; descriptions do not fix one witness. State the existential claim as its own premise, or use an evidenced proper name in an unquantified subject clause.
- `term.numeral`: A numeral phrase can state an exact quantity rather than identify one constant. Name the thing, or use 'the mouse'.
- `identity.constant_sides`: Pronouns can shift referent and quantities need not identify one object. Name the thing, or use 'the dog' on each side.
- `identity.extensional_context`: An unlisted predicate may describe an intensional role under which co-referring names cannot be exchanged. State the substituted claim as a premise.
- `slot.letter_in_name`: A schematic letter may be a literal component of Model U or Malcolm X. Use a variable away from names, or write the name out.
- `term.inflection_verbatim`: A plural surface may not denote instances of the stem guessed by deleting its suffix. Write the rule with singular 'every N' (Every dog is wet).

Since v45.0 descriptions admit any number of recognized modifiers plus a head (see below). Mixed subordinate-clause/term schemas are conservatively refused outright. EG uses a closed active-verb/copular grammar and refuses compound matrices and other quantified material. Noun inflection is verbatim in ordinary matching and counting classes; rewrite plural rules with singular “every N”. Ordinary initial-capital name evidence must appear away from sentence start and Therefore/So/Thus/Hence; internal capitals also count at sentence start. These intentional costs and the three inflection acceptance sites are detailed in the v44.0 coverage entry.
## v44.0: strict scope and connective rules

- **`scope.mixed_andor` — Mixed connective scope.** A and B or C may mean A and (B or C) or (A and B) or C; relative clauses and lowercase names do not disambiguate it. Use 'both … and' or 'either … or' to show the grouping.
- **`scope.clause_start` — Controlled clause starts.** Each inspected clause must start with a determiner, proper name, pronoun or numeral. An unrecognized start may be an operator with wider scope. Move the operator into each clause, or state the claim as an explicit premise.
- **`opener.ing_frame` — Initial -ing frame.** Following the script may qualify the entire conjunction rather than name its subject. Even genuine gerund subjects are refused when inspected. Closed-class clause starters such as `something`, `anything`, `everything`, and `nothing` are not frames. Move the frame into each clause, or state the framed claim as its own premise.
- **`scope.elided_copula` — Closed copula elision.** In Either the men are quiet or fish, fish may be a verb rather than a predicate with an omitted copula. Unknown predicates are not expanded. Repeat the subject and verb: 'or the men are fish'.
- **`scope.description_negation` — Description negation scope.** Wide negation need not assert that a description denotes; narrow predicate negation can presuppose existence. Only proper names permit the conversion. Add 'There is exactly one present king of France' and name it, or keep the wide negation.
- **`opener.ly_exception` — Adverb-capable noun exceptions.** Daily, early and only can be adverbs with scope over an entire sentence, even when they also have nominal uses. Move the adverb into each clause, or state the framed claim as its own premise.

These final-boundary checks precede normalized equality and apply independently of the certificate name. Whole literal reiteration remains opaque. The conservative scope scan may also refuse nominal coordination, quoted connectives, lowercase names, bare noun subjects, and genuine gerund subjects; use explicit grouping and full clauses. Copula elision requires one word from `singularPredicateHeads`.

### v44.0 refinements after round 2

Predicate positions after a copula accept any value, as in v43.0's
`afterCopula`: “If Rex is X, then Fido is X” admits “euphoric” and “a hero”.
A variable also used as a term still needs a singular value at that occurrence.
Internal capitals (AIXI, McCarthy) provide name evidence even at sentence
start; an ordinary initial capital alone does not.

The antecedent and consequent of “If P, then Q” are checked separately for
scope. A compound parenthesised operand is checked part by part, so grouping
cannot hide an unsafe clause start. Expletive “there” followed by be or a
modal, and a possessive subject, open a clause. A verb contraction is not a
possessive: “John's dog” and “John's running” have different shapes.

“The N of D” is a singular term when the head is singular and D is an
evidenced name or an admitted singular definite description, for example
“the death of the child”. Nested of-complements, indefinite complements,
plural heads and quantity/partitive heads remain refused.

Refusal diagnostics cover U/V/W as well as X/Y/Z. When a single v44 rule
accounts for a failed binding, its explanation and rewrite take priority over
the older indefinite-object fallback. This changes messages, not verdicts.

Closed `atom{...}` literals remain opaque to the English scope scan. The
canonical “both that …, and that …” scaffold, explicit comma grouping, and
a conditional used as a complete operand retain their own boundaries. Wide
negation around an identity is checked around its two terms, rather than
being mistaken for part of the left term. A conservative witness refusal
does not claim that an alternative scope reading has a countermodel.

## v45.0: bounded predicate and description loosenings

A singular term may have a coordinated adjective predicate after `is`, `are`,
`was`, or `were`. Every member must be one adjective, optionally preceded by
one listed adverb. Only a uniform `and` or `or` is accepted; mixed coordination,
quantified/plural subjects, and clause continuations receive no new exemption.
The adjective allowlist is `reliable possible feasible risky safe unsafe aligned
causal informational dangerous useful happy red blue green warm cold wise rich
poor friendly mortal alive dead hungry tired valuable eligible registered ready
late early wet dry small large tall short round square honest kind brave light
portable`. The adverbs are `financially very quite extremely deeply highly`.
Unknown adjectives and suffix homographs remain refused: suffixes alone cannot
exclude nouns such as “criminal” or “general”.

```argcheck-accept schematic instantiation
1. If X is informational or causal, then X is a link.
2. The channel is informational or causal.
3. Therefore, the channel is a link. [1,2]
```

```argcheck-accept schematic instantiation
1. If X is possible and financially feasible, then X will be built.
2. The machine is possible and financially feasible.
3. Therefore, the machine will be built. [1,2]
```

Definite descriptions admit any number of recognized modifiers and one singular
head. Modifiers are the adjective list above plus `first second third fourth
fifth sixth seventh eighth ninth tenth ordinary ultraintelligent aps big red
small old young present former parking`. `APS` is an explicit attributive
acronym exception. Other names, determiners, quantifiers, cardinal numerals,
connectives, relatives, prepositions and finite verbs are excluded. Ordinals
are permitted only as modifiers, not as quantity heads. Unknown modifiers are
refused. Capitalized possessive owners supply evidence for that owner only;
they do not establish a bare name elsewhere.

A single `of` complement may be an existing allowed singular name/definite or
a bare singular phrase with the same modifier restrictions. The closed
collective heads `group` and `set` also permit bare plural complements.
Nested `of` remains refused.

```argcheck-accept schematic instantiation
1. If X is solved, then X is boring.
2. The game of ordinary chess is solved.
3. Therefore, the game of ordinary chess is boring. [1,2]
```

```argcheck-accept schematic instantiation
1. If X is a risk, then X deserves attention.
2. The first ultraintelligent machine is a risk.
3. Therefore, the first ultraintelligent machine deserves attention. [1,2]
```

```argcheck-accept schematic instantiation
1. If X is solved, then X is boring.
2. Carlsmith's APS concept is solved.
3. Therefore, Carlsmith's APS concept is boring. [1,2]
```

```argcheck-accept schematic instantiation
1. If X is solved, then X is boring.
2. The group of robots is solved.
3. Therefore, the group of robots is boring. [1,2]
```

The closed quantity/partitive exclusion is `majority minority rest remainder
bulk most half third quarter fraction part portion share percentage percent
proportion number amount quantity lot lots bunch plenty couple handful dozen
score plurality pair few several many all some none any each both either neither
whole balance residue`. These words remain refused as heads, including in
possessives; existing exclusions in other positions are retained.

```argcheck-refused slot.value_singular | Name the thing | schematic instantiation
1. If X is solved, then X is boring.
2. The majority of voters is solved.
3. Therefore, the majority of voters is boring. [1,2]
```

Predicate-only variables after `being`, `need/needs/have/has/had to be`, and
modal `be` use the same copula classifier for binding and whole-rule slot types.
A variable also used as a term still needs a singular term value.

```argcheck-accept schematic instantiation
1. If the machine needs to be X, then the channel is X.
2. The machine needs to be ultraparallel.
3. Therefore, the channel is ultraparallel. [1,2]
```

Bare plurals and mass values remain refused. An `is`/`are` mismatch in the
plural friction probe now reports `slot.value_singular` without certifying it.
Clause-start failures quote both the offending clause and its first token.
