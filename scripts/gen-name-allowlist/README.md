# Reproducible name allowlist

From the repository root, run one command (Python 3, standard library only):

```sh
python3 scripts/gen-name-allowlist/generate.py
```

The checked-in noun/adjective lists are exclusion inputs. If changing those,
regenerate them first with `python3 scripts/gen-noun-allowlist/generate.py`.
Downloads default to ignored `workspace/name-cache/`; `--cache DIR` overrides
this, and `--output-dir DIR` writes a separate reproducibility result. Every
source is commit-pinned and SHA256-verified, even on cache hits; mismatches fail
before output writes. There are no timestamps, local paths, network searches,
private data, or manually supplied names in the generated files.

## Sources and attribution

- Mark Kantrowitz's Names Corpus 1.3 (1994-03-29), with additions by Bill Ross,
  from `nltk/nltk_data` commit `550b6625bcef1f2abff2ff770a5a0d272c9c6b2a`,
  `packages/corpora/names.zip`. Its permissive attribution/redistribution
  notice is reproduced in `validation/lexicon/LICENSE-names.txt`.
- FiveThirtyEight's `most-common-name/surnames.csv`, US Census Bureau surname
  data, at `fivethirtyeight/data` commit
  `4c1ff5e3aef1816ae04af63218015066e186c147`. We retain the repository's
  CC-BY-4.0 license and attribution in `LICENSE-census-names.txt`. The pinned
  source README records Census provenance; do not label this copy Census 2010.
  We change case, filter, deduplicate, and combine it with the NLTK corpus.
- WordNet 3.0 at the same NLTK commit, under the existing
  `LICENSE-WordNet.txt`, supplies the ordinary-word exclusion vocabulary.
- `brekker23/English-word-frequencies` at
  `815665f95ddab30520dff6df260ed3b6244b9638` supplies common-word exclusions
  under Apache-2.0 (existing `LICENSE-frequency.txt`).

`sources.json` contains every URL, pin, license and checksum, including the
Census-copy README and licenses. No external non-GitHub download is required.

## Selection and soundness

Use all NLTK first names and Census surnames with **count >= 1,000**. Keep
lowercase ASCII alphabetic words of at least three letters. Reject overlap with:

- the noun generator's quantity words, closed/determiner/pronoun sets and
  compounds, including inflections;
- all four checked-in lexical lists (plural/mass, singular, mass, adjective);
- every WordNet POS lemma, multiword-lemma component, exception form and
  conservatively generated inflection (plural, verb tense, comparative,
  superlative; including y/e/doubling and f/fe variations);
- the **top 10,000** ASCII lowercase English-word-frequencies entries, ranked
  by decreasing count with alphabetic tie-breaking.

Morphology deliberately overgenerates, so legitimate names can be excluded.
This is conservative evidence, not a comprehensive register of people.
The result has **22,369 names**, from 29,326 distinct candidates. Exclusion
categories overlap; `EXCLUDED-names.md` records each count and the unique total.
Leah, Amanda, Jennifer, Megan, Olivia, Samantha and Patel survive. Sarah, Emily,
Michael, Johnson and Nguyen do not survive the WordNet/common-word exclusions.
Yudkowsky is absent from these sources; it is not hand-added.

`names.txt` supplies `lexicon.IsName`, a lowercase membership lookup. Runtime
callers still require a capitalized token, after stripping possessives.
`name_forbidden.txt` supplies `IsNameForbidden`: quantity/closed forms cannot
become names merely by being capitalized in a mid-sentence term position.
Internal-capital evidence retains its existing separate treatment.

The tests regenerate twice from the cache and compare all name outputs against
the checkout. Missing cache skips regeneration rather than fetching implicitly.
