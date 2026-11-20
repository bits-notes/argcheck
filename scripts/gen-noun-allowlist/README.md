# Reproducible noun allowlist

Run from the repository root with Python 3 (standard library only):

```sh
python3 scripts/gen-noun-allowlist/generate.py
```

Downloads are cached in ignored `.cache/noun-allowlist/`. Pass `--cache DIR`
for an existing cache. Every file is SHA256-verified on every run, including
cached files; a mismatch fails before writing output. `sources.json` records
exact URLs, commits, licenses and checksums for every downloaded file, including
the frequency producer's script retained in the cache for provenance review.
The generator does not execute that script. No raw source data is committed.
The generated list, exclusion report and license notices are byte deterministic.
Frequency ties use alphabetical order. No current date or local path is emitted.

## Sources and licenses

* Princeton WordNet 3.0, mirrored in NLTK data at commit
  `550b6625bcef1f2abff2ff770a5a0d272c9c6b2a`, provides noun lemmas, irregular
  morphology, semantic relations, verb lemmas and tagged sense counts.
  Its permissive Princeton WordNet license is reproduced in
  `validation/lexicon/LICENSE-WordNet.txt`.
* `brekker23/English-word-frequencies` at commit
  `815665f95ddab30520dff6df260ed3b6244b9638` provides `word_count.txt`.
  The repository licenses the material under Apache-2.0 and describes the input
  as public-domain books. Its producer script counts English Gutenberg texts,
  strips punctuation and filters against a word list. The source's Apache-2.0
  license is reproduced in `validation/lexicon/LICENSE-frequency.txt`.
  This corpus is biased toward older literature; it is not a contemporary
  balanced-English ranking. Source claims about the entire English vocabulary
  are not assumed. We use only observed counts.

These are permissive sources; no wordfreq, CC-BY-SA, Google/LDC frequency data,
Norvig-derived frequency data, or Wiktionary data is used. The selected sources'
notices remain with the derived data; the repository's MIT license does not
replace their licenses. Upstream URLs are in the checksum manifest.

## Selection

1. Parse WordNet noun synsets. Keep lowercase ASCII single-word lemmas from
   non-instance synsets (exclude capitalized names and named-instance senses).
2. Generate regular plurals: consonant-y → ies; s/x/z/ch/sh → es; otherwise s.
   WordNet `noun.exc` takes precedence for irregular plurals, including invariant
   forms. Retain only frequency-attested forms. This combines morphological and
   corpus evidence; WordNet does **not** annotate countability, so this is an
   approximation rather than a proof that each lemma has a count use.
3. Add a small explicitly reviewed list of plural-only nouns, and the reviewed
   uncountable senses in `mass-nouns.txt`. Every mass noun must exist in WordNet;
   no broad "abstract noun" or suffix heuristic admits singular count nouns.
4. Apply all exclusions, then select the 8,000 highest-frequency eligible forms.
   Include the brief's 14 required terms even below that cutoff. `subgoals`
   is the sole supplied count-noun form absent from WordNet and from the ranking.
   It receives count zero. The result currently contains **8,001 words**.

## Exclusions

* `quantity-words.txt` supplies partitives, quantities, proportions, container
  measures, collectives usable as quantities, units, cardinal and ordinal words.
  Both each supplied form and its regular/WordNet irregular plurals are denied.
* Deny every lemma (and its plurals) in WordNet's `noun.quantity` category or any
  descendant of synset `00033615` (measure/quantity/amount). This catches obscure
  measures, currencies and time quantities too. Any such sense disqualifies the
  word, even when another sense is a normal count noun. The filter deliberately
  over-excludes ambiguous words such as feet, seconds and glasses.
  **One explicit conflict resolution:** the brief requires `information`, but
  WordNet gives it an information-theory numerical-measure sense. That word alone
  is exempted from the semantic quantity filter; its ordinary mass use is kept.
* Read `singularPartitiveWords`, `termDeterminers`, `quantityNumerals` and
  `singularClosedWords` directly from `validation/*.go`; fail on missing or
  ambiguous declarations. Deny those words and their plurals, plus a supplemental
  closed-class list and every/some/any/no + one/body/thing/where compounds.
* Deny third-person singular forms of common verbs, defined reproducibly as
  verb lemmas with at least five tagged occurrences across WordNet verb senses.
  Use regular morphology plus WordNet's irregular verb forms ending in s.
  This removes means, runs, works, leads, claims and results. It leaves machines
  (required by the brief), whose verb lemma has zero tagged occurrences.
  **Goods stays:** good is adjectival, but goods is an established plural noun
  and not a common third-person verb form. We do not compare "first senses"
  across POS: WordNet orders senses separately within POS, not in one global
  noun/verb/adjective order. Tagged counts are sparse and historical, so the
  full list of remaining rare/untagged verb overlaps is provided for review.

`EXCLUDED.md` reports overlapping exclusion counts in the candidate pool,
30 highest-frequency examples per category (all if fewer qualify), the global
deny-set sizes, the selection boundary, and both common and rare verb-overlap
review lists. Exclusions are enforced even when no candidate demonstrates them.
The final 50 highest- and lowest-frequency words are printed and recorded there.

This slice supplies data and an isolated lookup only. No validator calls it and
no acceptance/rejection rule changes. Future integration must still validate
syntax and context; lexical membership alone cannot establish countability or POS.
