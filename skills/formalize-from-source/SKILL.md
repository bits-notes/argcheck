---
name: formalize-from-source
description: Turn a source passage into a faithful numbered argument. Build a sentence ledger that gives every source sentence a disposition, anchor each premise to its sentences, mark supplied premises, then check validity with argcheck and fidelity against the ledger. Use when reconstructing an author's argument from a text, not when writing your own.
---

# Formalize an argument from a source text

`argcheck` decides validity. It never checks that an argument matches its
source. This skill is the fidelity half: what the author said, what you added,
and what you left out, each visible and checkable.

## 1. Build the sentence ledger

Split the passage into sentences. Split a sentence further where it carries
more than one proposition ("A, but B"; "A, so B"). Give each unit a stable
locator: `S1`, `S2`, … or `page.sequence` (`12.3`) for long texts.

Give every unit exactly one disposition:

| Disposition | Meaning |
|---|---|
| `used` | appears as a premise or conclusion, reworded only as section 3 allows |
| `supports` | backs a step without being one (an example, an illustration, data) |
| `context` | sets the scene or defines a term the argument relies on |
| `signpost` | structure only ("I now turn to…", "In summary…") |
| `excluded` | left out, with a stated reason |

No unit is left without a disposition. An exclusion always gives a reason, for
example "reports another author's view without endorsing it", "aside on method"
or "a hedge, kept outside the argument (see section 3)".

## 2. Keep the exact source text beside the ledger

Copy each unit's exact words next to its locator. Do not paraphrase in the
ledger. Checking fidelity means comparing the argument with these words, so
they must be the author's.

```text
S1  Most residents rely on the river for drinking water.        used
S2  The council has approved a chemical plant upstream.         used
S3  Plants of this kind have leaked in three nearby towns.      supports
S4  I will not discuss the economic case here.                  excluded: signpost of scope
S5  So the council should require an independent spill audit.   used (conclusion)
```

## 3. Anchor each premise and mark what you supplied

- Tag every premise with its locators, e.g. `[S1]`, in a note beside the
  argument (not inside the numbered lines, where `[n]` means a step reference).
- A premise the author did not state, but the inference needs, is **supplied**.
  Mark it as supplied and say why it is needed. Never fold an unstated
  assumption into a stated premise.
- Keep every qualifier and scope restriction the author wrote ("most",
  "in this case", "unless the plant is inspected"). The only changes allowed
  are word order, some tense, and replacing pronouns with what they refer to.
- Keep hedges ("I suspect", "it may be that") outside the argument and label
  them. Do not harden "I believe P" into "P", or into "one has reason to
  believe P".

## 4. Bridge premises: state the author's link, not a re-join

A supplied conditional must carry the connection the author relies on. A rule
whose consequent only restates its own antecedents makes any list of claims
"valid" and shows nothing.

Bad (the supplied rule re-joins its premises):

```text
1. Most residents rely on the river for drinking water.
2. The council has approved a chemical plant upstream.
3. If most residents rely on the river for drinking water and the council has approved a chemical plant upstream, then most residents rely on the river for drinking water and a chemical plant has been approved upstream of it.   (supplied)
4. Therefore, most residents rely on the river for drinking water and a chemical plant has been approved upstream of it. [1,2,3]
```

Good (the supplied rule states why the facts matter, as the passage implies):

```text
1. Most residents rely on the river for drinking water.
2. The council has approved a chemical plant upstream.
3. If most residents rely on the river for drinking water and the council has approved a chemical plant upstream, then the council should require an independent spill audit.   (supplied: the author's link between S1-S2 and S5)
4. Therefore, the council should require an independent spill audit. [1,2,3]
```

Both pass `argcheck`; only the second is faithful. The `(supplied…)` tags are
notes for the reader, so remove them before running the checker.

If the passage only reports or lists claims and never draws a conclusion from
them, do not force an inference. Record those units as `context` or `excluded`
and say so.

## 5. Check validity

Run `argcheck` on the numbered argument and repair until it prints `VALID`. See
the `argcheck` skill for the format and for repairs. Repairs must still obey
section 3: passing the gate by weakening or overstating the source is a failure,
not a fix.

## 6. Close with a fidelity pass against the ledger

Before reporting the argument, check both directions:

- Every `used` unit appears in the argument, with its qualifiers intact.
- Every premise and conclusion has a locator or a `supplied` mark.
- Every `excluded` unit has a reason, and no unit was skipped.
- Every supplied premise states a real link (section 4), not a restatement.
- Hedges are outside the argument and labelled.

Report the ledger, the argument with its locator notes, the supplied premises
with their reasons, and the `argcheck` verdict line.
