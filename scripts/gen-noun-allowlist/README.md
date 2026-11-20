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

* `quantity-words.txt` is unchanged: partitives, quantities, proportions,
  container measures, collectives, units, cardinals and ordinals remain denied,
  including their regular/WordNet irregular plurals.
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
* **Verb forms are not an exclusion.** Report overlaps with common verbs
  (at least five tagged WordNet verb uses) and rare/untagged verbs separately
  for review. Noun/verb overlap does not invalidate a noun; there is no global
  cross-POS first-sense ordering. For example, does also pluralizes doe.

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

This is isolated lexical data and lookup only: no validator imports it and no
acceptance/rejection rules change. Membership alone cannot establish POS or
countability in a sentence.

## Bounded verification

```sh
go vet ./validation/lexicon/...
go test -vet=off ./validation/lexicon ./validation -run 'Lexicon|DocsRules'
```

Regenerate twice and compare all generated files byte for byte. The lookup
remains package-private; testing `./validation` alone misses the lexicon tests.
