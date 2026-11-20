# Reproducible noun allowlist

Run from the repository root with Python 3 (standard library only):

```sh
python3 scripts/gen-noun-allowlist/generate.py
```

Downloads are cached in ignored `.cache/noun-allowlist/`. Pass `--cache DIR`
for an existing cache. Every file is SHA256-verified on every run, including
cached files; a mismatch fails before writing output. `sources.json` records
exact URLs, commits, licenses and checksums for all downloads, including both
frequency producer scripts retained for provenance review (never executed).
Generated data, reports and license notices are byte deterministic. No current
date or local path is emitted.

## Sources and licenses

* Princeton WordNet 3.0, mirrored in NLTK data at commit
  `550b6625bcef1f2abff2ff770a5a0d272c9c6b2a`, supplies noun lemmas, ordered noun
  senses, morphology, semantic relations, verbs and tagged sense counts.
  Its permissive license is reproduced in `validation/lexicon/LICENSE-WordNet.txt`.
* `brekker23/English-word-frequencies`, commit
  `815665f95ddab30520dff6df260ed3b6244b9638`, supplies `word_count.txt` under
  Apache-2.0. The producer counts English public-domain Gutenberg books, strips
  punctuation and filters against a word list. This ranking favors older
  literature; it is not a balanced contemporary corpus. Its notice is retained
  in `validation/lexicon/LICENSE-frequency.txt`.
* `IlyaSemenov/wikipedia-word-frequency`, commit
  `d11c4fab211156e96ab5943850dd6f98604707fb`, supplies the English Wikipedia
  frequency table `results/enwiki-2023-04-13.txt` under the repository's MIT
  license, retained verbatim in `validation/lexicon/LICENSE-wikipedia-frequency.txt`.
  The producer lowercases article text, normalizes dashes/apostrophes, rejects
  words with digits and requires occurrence in at least three articles. The
  table supplies modern encyclopedic vocabulary, with encyclopedia-specific
  bias. We use the published word counts, not Wikipedia article text; the MIT
  notice is the frequency repository's license, not a relicensing of Wikipedia.

No wordfreq, Google/LDC, Norvig-derived or Wiktionary frequency data is used.
Upstream notices remain with the derived data; this repository's MIT license
does not replace them. Upstream URLs are in the checksum manifest.

## Selection

1. Keep lowercase ASCII single-word WordNet noun lemmas from non-instance
   synsets (exclude capitalized names and named-instance senses).
2. Generate plurals: consonant-y → ies; s/x/z/ch/sh → es; otherwise s.
   WordNet `noun.exc` takes precedence. Reviewed overrides add people/persons
   and invariant sheep, deer, moose, aircraft, spacecraft, offspring, bison,
   salmon, trout, swine, grouse, squid and species. Retain frequency-attested
   forms. WordNet has no countability annotation: morphology plus attestation
   is an approximation, not proof of a count use.
3. Add reviewed plural-only nouns and `mass-nouns.txt` (each mass noun must
   have a WordNet noun synset). No broad suffix or abstract-noun heuristic.
4. Rank each source's lowercase ASCII words independently by descending count,
   breaking count ties alphabetically. Rank a word by its better (smaller)
   position in either source; absent words rank last. Break combined-rank ties
   alphabetically. Counts from different corpora are never compared directly.
5. Apply exclusions, select the best-ranked 9,000 eligible forms, and retain
   the 14 original required terms plus `credences` even below the cutoff.
   `credences` is checked against WordNet noun `credence` and its plural.
   `subgoals` remains the sole supplied count form absent from WordNet.
   The result contains **9,002 words**.

## Exclusions and overlap review

* `quantity-words.txt`: partitives, quantities, proportions, container
  measures, collectives, units, cardinals and ordinals are denied, including
  their regular/WordNet irregular plurals. It also denies nouns whose bare
  plural is a quantity idiom ("ages", "eons", "droves", "hordes", "legions",
  "swathes"). Concrete nouns with an "N of" quantity idiom (oceans, mountains,
  waves) stay in, because the slot rule refuses any word after the head.
* Semantic exclusion uses only the **first noun sense** in `index.noun` and
  its hypernym ancestry. The root `00033615` is quantity.n.01 / measure.n.02 /
  amount.n.03, with descendants including unit_of_measurement.n.01 and
  number.n.02. Later quantity senses do not disqualify a word. For inflections
  without their own noun entry, use the base's first sense; a surface form's
  own first noun sense takes precedence (species is not excluded as the
  generated plural of specie). There is no special information exception.
* Read `singularPartitiveWords`, `termDeterminers`, `quantityNumerals` and
  `singularClosedWords` directly from `validation/*.go`; fail on missing or
  ambiguous declarations. Deny these and their plurals, supplemental closed
  words and every/some/any/no + one/body/thing/where compounds.
* **Verb-dominant forms are excluded; ordinary noun/verb overlap is not.**
  A third-person verb form is denied when its verb senses are tagged more
  than three times as often as the noun senses it would pluralise (and at
  least five times), plus is/has/does/was/says/goes. This removes "says",
  "does" (also the plural of doe), "means" and "forecasts", and keeps
  works, values, risks, costs and claims. Other overlaps with common verbs
  (at least five tagged WordNet verb uses) and rare/untagged verbs are
  reported for review only.

`EXCLUDED.md` gives overlapping candidate exclusion counts, global deny-set
sizes, frequency-ordered examples, selection boundaries, both verb-overlap
review lists, and the highest/lowest 50 included words.

## Core probe exceptions

The must-include test covers the core argument probe except:

* `nuclear`: adjective only.
* `datasets`: no dataset noun lemma in WordNet 3.0.
* `qualia`: noun.exc lists qualia → quale, but quale has no noun synset.
* `norms`: norm's first sense → standard → system_of_measurement → quantity.
* `money`: first sense → medium_of_exchange → standard → system_of_measurement
  → quantity.
* `probabilities`: probability's first sense directly descends from quantity.

The last three remain excluded by the requested first-sense rule; no semantic
exceptions are supplied. The probe therefore reports these six missing words.

The v46 validator imports these allowlists. Membership alone cannot establish
POS or countability in a sentence; binding checks supply context and agreement.

## Bounded verification

```sh
go vet ./validation/lexicon/...
go test -vet=off ./validation/lexicon ./validation -run 'Lexicon|DocsRules'
```

Regenerate twice and compare all generated files byte for byte. Testing `./validation` alone misses the lexicon tests. Public lookups are
`IsPluralMassNoun`, `IsSingularNoun`, `IsAdjective`, and `IsMassNoun`.

## Singular nouns, adjectives and agreement (v46.0)

The generator also selects the top 14,000 frequency-attested WordNet singular
noun lemmas (including mass nouns), with exactly the plural-list exclusions and
without plural-only nouns. Required core vocabulary is retained below the cutoff.
All `singularPartitiveWords` are excluded. The resulting list has 14,000 entries.

Adjectives include WordNet adjectives and satellites, plus attested regular
past participles of verbs with at least five tagged uses. WordNet verb exceptions
ending in -ed or -en take precedence; other irregular past forms are not inferred
as participles. Take the best-ranked 5,000 eligible words plus required legacy
predicate/description vocabulary: 5,007 entries. `adjective-exclusions.txt`
excludes quantity, determiner and identity readings even where an adjective sense
exists, including cardinals/ordinals; Go closed words and pronoun compounds are
also excluded. Explicit exclusions take priority over legacy modifiers: `former`
and ordinals are not adjective members. Ordinals retain their separate description
rule. `adjective-supplement.txt` supplies APS, ultraintelligent and the legacy
participial attributive parking. All three are explicit reviewed additions.

`mass_nouns.txt` in the output is the intersection of the reviewed mass list and
the selected plural/mass vocabulary (303 entries), used for agreement rather than
suffix guessing. Singular/plural homographs use the reviewed mass classification
when present; it is deliberately not a full contextual countability parser.

Use `--output-dir DIR` to regenerate outside the checkout. The deterministic
regeneration test compares two fresh outputs against every checked-in lexical
list and exclusion report when the pinned source cache is available. Missing
cache skips that test; it never downloads implicitly in a test run.

## Name evidence

The companion [name generator](../gen-name-allowlist/README.md) produces the
conservative `names.txt` allowlist and quantity/closed-word runtime veto.
`IsName` is a lowercase lookup; validation additionally requires capitalization.
The adjective required vocabulary includes WordNet's `deferential` for the
narrow degree-adverb opener exception; the adjective list now has 5,007 entries.
