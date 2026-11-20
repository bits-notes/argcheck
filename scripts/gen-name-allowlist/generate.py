#!/usr/bin/env python3
"""Deterministic conservative name evidence; standard-library Python only."""
import argparse
import csv
import hashlib
import io
import json
from pathlib import Path
import re
import urllib.request
import zipfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
NOUN = ROOT / 'scripts/gen-noun-allowlist'


def words(path):
    return set(' '.join(line.split('#')[0] for line in path.read_text().splitlines()).split())


def inflections(w):
    # Conservative overgeneration is intentional: ambiguity must lose evidence.
    forms = {w, w+'s', w+'es', w+'ed', w+'ing', w+'er', w+'est'}
    if w.endswith('y'):
        forms |= {w[:-1]+s for s in ('ies', 'ied', 'ier', 'iest')}
    if w.endswith('e'):
        forms |= {w+'d', w[:-1]+'ing', w+'r', w+'st'}
    if len(w) > 2 and w[-1] not in 'aeiouwxy' and w[-2] in 'aeiou':
        forms |= {w+w[-1]+s for s in ('ed', 'ing', 'er', 'est')}
    if w.endswith('ie'):
        forms.add(w[:-2]+'ying')
    if w.endswith('f'):
        forms.add(w[:-1]+'ves')
    if w.endswith('fe'):
        forms.add(w[:-2]+'ves')
    return forms


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--cache', type=Path, default=ROOT/'workspace/name-cache')
    ap.add_argument('--output-dir', type=Path, default=ROOT/'validation/lexicon')
    args = ap.parse_args()
    args.cache.mkdir(parents=True, exist_ok=True)
    args.output_dir.mkdir(parents=True, exist_ok=True)
    for src in json.loads((HERE/'sources.json').read_text()):
        path = args.cache/src['file']
        data = path.read_bytes() if path.exists() else urllib.request.urlopen(src['url'], timeout=60).read()
        if hashlib.sha256(data).hexdigest() != src['sha256']:
            raise SystemExit('SHA256 mismatch: '+src['file'])
        if not path.exists():
            path.write_bytes(data)
    with zipfile.ZipFile(args.cache/'names.zip') as z:
        first = set((z.read('names/male.txt')+z.read('names/female.txt')).decode().lower().split())
        notice = z.read('names/README').decode()
    rows = list(csv.DictReader(io.StringIO((args.cache/'surnames.csv').read_text())))
    surnames = {r['name'].lower() for r in rows if int(r['count']) >= 1000}
    candidates = first | surnames
    # WordNet's index.* files are lowercased, so case comes from the synset
    # lines in data.*: a lemma written lowercase in any synset is an ordinary
    # word (bill, rose, clement); one written only capitalised is a proper noun
    # (Alice, Scott) and is not, by itself, evidence of a common word.
    wordnet = set()
    lower_lemmas, proper_lemmas = set(), set()
    with zipfile.ZipFile(args.cache/'wordnet.zip') as z:
        for pos in ('noun','verb','adj','adv'):
            for line in z.read('wordnet/data.'+pos).decode().splitlines():
                if not line or line.startswith(' '):
                    continue
                f = line.split()
                for k in range(int(f[3], 16)):
                    lemma = re.sub(r'\(.*\)$', '', f[4+2*k])
                    (lower_lemmas if lemma == lemma.lower() else proper_lemmas).add(lemma.lower())
            wordnet.update(z.read('wordnet/'+pos+'.exc').decode().lower().split())
    for lemma in lower_lemmas:
        wordnet.update(inflections(lemma))
        # Components of multiword lemmas can also be common words.
        wordnet.update(re.split('[_-]', lemma))
    proper_only = proper_lemmas - wordnet
    quantity = set().union(*(inflections(w) for w in words(NOUN/'quantity-words.txt')))
    closed = set()
    for name in ('singularPartitiveWords', 'termDeterminers', 'quantityNumerals', 'singularClosedWords'):
        matches = []
        for path in sorted((ROOT/'validation').glob('*.go')):
            matches += re.findall(r'var\s+'+name+r'\s*=\s*wordSet\("([^"]*)"\)', path.read_text())
        if len(matches) != 1:
            raise SystemExit('Expected exactly one wordSet declaration for '+name)
        closed.update(matches[0].split())
    closed.update('mine yours hers ours theirs myself yourself oneself thou thee thy thine ye who whom somebody someone something somewhere anybody anyone anything anywhere nobody noone nothing nowhere everybody everyone everything everywhere whatever whoever whomever whichever wherever whenever neither either each other another such that this these those what which whose whatsoever whomsoever whosoever'.split())
    closed = set().union(*(inflections(w) for w in closed))
    closed |= {p+h+s for p in ('every','some','any','no') for h in ('one','body','thing','where') for s in ('','s')}
    freq = {}
    for line in (args.cache/'word_count.txt').read_text().splitlines():
        w,n = line.split()
        if re.fullmatch('[a-z]+',w):
            freq[w] = int(n)
    # The frequency list is lowercased, so frequent names (alice, scott) look
    # common; exempt only words WordNet knows solely as capitalised proper nouns.
    common = set(sorted(freq,key=lambda w: (-freq[w],w))[:10000]) - proper_only
    categories = [('non-ASCII, non-letter or at most two letters', {w for w in candidates if not re.fullmatch('[a-z]{3,}',w)}), ('quantity words and inflections',quantity), ('closed words and inflections',closed)]
    for file in ('plural_mass_nouns.txt','singular_nouns.txt','mass_nouns.txt','adjectives.txt'):
        categories.append((file,words(ROOT/'validation/lexicon'/file)))
    categories += [('WordNet lowercase lemmas (all POS), components and inflections',wordnet),('English-word-frequencies top 10000, less WordNet proper-only nouns',common)]
    denied = set().union(*(s for _,s in categories))
    selected = candidates-denied
    def write(name,data):
        (args.output_dir/name).write_text(data)
    write('names.txt',''.join(w+'\n' for w in sorted(selected)))
    write('name_forbidden.txt',''.join(w+'\n' for w in sorted(quantity|closed) if re.fullmatch('[a-z]+',w)))
    write('LICENSE-names.txt',notice)
    write('LICENSE-census-names.txt','FiveThirtyEight data: most-common-name/surnames.csv.\nDerived from US Census Bureau surname data; frequency floor 1000.\nSource and pin: scripts/gen-name-allowlist/sources.json.\nChanges: lowercase, filtering, deduplication, combined with NLTK names.\n\n'+(args.cache/'census-LICENSE').read_text())
    report = ['# Name allowlist exclusions','',f'First names: {len(first)}; surnames with count >= 1000: {len(surnames)}.',f'Unique candidates: {len(candidates)}; selected: {len(selected)}.', '', 'Overlapping candidate exclusion counts (not additive):','']
    report += [f'- {label}: {len(candidates & denied_set)}' for label,denied_set in categories]
    report += ['',f'Total unique excluded: {len(candidates-selected)}.',f'leah: {"included" if "leah" in selected else "excluded"}; yudkowsky: {"included" if "yudkowsky" in selected else "absent"}.','']
    write('EXCLUDED-names.md','\n'.join(report))
    print('\n'.join(report))


if __name__ == '__main__':
    main()
