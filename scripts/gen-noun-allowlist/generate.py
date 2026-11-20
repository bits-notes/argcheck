#!/usr/bin/env python3
"""Generate the v46 noun lexicon; Python 3 standard library only."""
import argparse
from collections import defaultdict
import hashlib
import json
from pathlib import Path
import re
import urllib.request
import zipfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parent.parent
OUT = ROOT / 'validation' / 'lexicon'
REQUIRED = set('subgoals humans machines robots systems agents humanity intelligence evidence water information knowledge research software credences'.split())
LIMIT = 9000


def words(path):
    return set(' '.join(line.split('#')[0] for line in path.read_text().splitlines()).split())


def regular(word):
    if word.endswith('y') and len(word) > 1 and word[-2] not in 'aeiou':
        return word[:-1] + 'ies'
    if word.endswith(('s', 'x', 'z', 'ch', 'sh')):
        return word + 'es'
    return word + 's'


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--cache', type=Path, default=ROOT / '.cache' / 'noun-allowlist')
    args = ap.parse_args()
    args.cache.mkdir(parents=True, exist_ok=True)
    sources = json.loads((HERE / 'sources.json').read_text())
    for source in sources:
        dest = args.cache / source['file']
        if not dest.exists():
            with urllib.request.urlopen(source['url'], timeout=60) as response:
                data = response.read()
        else:
            data = dest.read_bytes()
        if hashlib.sha256(data).hexdigest() != source['sha256']:
            raise SystemExit('SHA256 mismatch: ' + source['file'])
        if not dest.exists():
            dest.write_bytes(data)

    with zipfile.ZipFile(args.cache / 'wordnet.zip') as z:
        def read(name):
            return z.read('wordnet/' + name).decode('utf-8')
        noun_data, exceptions, senses = read('data.noun'), read('noun.exc'), read('index.sense')
        # index.noun lists synsets in WordNet frequency order (within nouns).
        first_noun = {f[0]: f[6 + int(f[3])]
                      for line in read('index.noun').splitlines()
                      if line and not line.startswith(' ')
                      for f in [line.split()]}
        wn_license = read('LICENSE')
        verbs = {line.split()[0] for line in read('index.verb').splitlines()
                 if line and not line.startswith(' ')}
        verb_exc = read('verb.exc')

    # WordNet has POS and hypernyms, but no countability annotation. Require
    # attested plural morphology, or an explicitly reviewed mass noun.
    synsets = {}
    noun_senses = defaultdict(set)
    for line in noun_data.splitlines():
        if not line or not line[0].isdigit():
            continue
        fields = line.split('|')[0].split()
        offset, lexfile, count = fields[0], int(fields[1]), int(fields[3], 16)
        lemmas = fields[4:4 + 2 * count:2]
        i = 4 + 2 * count
        pointers = [fields[j:j+4] for j in range(i+1, i+1+4*int(fields[i]), 4)]
        parents = {p[1] for p in pointers if p[0] in ('@', '@i') and p[2] == 'n'}
        instance = any(p[0] == '@i' for p in pointers)
        synsets[offset] = (lemmas, parents, instance, lexfile)
        for lemma in lemmas:
            if re.fullmatch('[a-z]+', lemma) and not instance:
                noun_senses[lemma].add(offset)

    irregular = defaultdict(set)
    for line in exceptions.splitlines():
        inflected, *bases = line.split()
        for base in bases:
            irregular[base].add(inflected)
    # noun.exc is not exhaustive: it omits common invariant plurals and people.
    irregular['person'].update(('people', 'persons'))
    for base in 'sheep deer moose aircraft spacecraft offspring bison salmon trout swine grouse squid species'.split():
        irregular[base] = {base}

    def plurals(base):
        # Exceptions are authoritative (avoid childs, mans, mouses, etc.).
        return irregular.get(base, {regular(base)})

    def expand(bases):
        return set(bases) | {p for b in bases for p in plurals(b)}

    mass = words(HERE / 'mass-nouns.txt')
    missing_mass = mass - noun_senses.keys()
    if missing_mass:
        raise SystemExit('Mass terms missing from WordNet: ' + ', '.join(sorted(missing_mass)))
    candidates = set(mass)
    for base in noun_senses:
        candidates.update(plurals(base))
    # WordNet plural-only nouns, including goods and clothes; do not admit
    # singular -s nouns merely because their spelling ends in s.
    plural_only = set('police clergy personnel goods clothes trousers pants shorts scissors pliers tweezers tongs spectacles binoculars belongings remains surroundings outskirts earnings savings proceeds credentials annals archives entrails genitals groceries pajamas pyjamas thanks tidings valuables wages'.split())
    candidates.update(plural_only & noun_senses.keys())
    # Product vocabulary explicitly required by the brief; subgoal is absent
    # from WordNet 3.0. This is the sole manually supplied count-noun lemma.
    candidates.add('subgoals')
    if 'credence' not in noun_senses or 'credences' not in plurals('credence'):
        raise SystemExit('Required credences must be a WordNet-derived noun plural')

    # Rank each source independently; alphabetical order resolves equal counts.
    # Use the better observed rank, never compare counts across corpora.
    ranks = {}
    for filename in ('word_count.txt', 'enwiki-2023-04-13.txt'):
        freq = {}
        for line in (args.cache / filename).read_text().splitlines():
            w, n = line.split()
            if re.fullmatch('[a-z]+', w):
                freq[w] = int(n)
        for position, w in enumerate(sorted(freq, key=lambda w: (-freq[w], w)), 1):
            ranks[w] = min(ranks.get(w, position), position)
    rank = lambda w: (ranks.get(w, float('inf')), w)

    quantity = expand(words(HERE / 'quantity-words.txt'))
    # quantity.n.01 = measure.n.02 = amount.n.03 (00033615).
    # Its descendants include unit_of_measurement.n.01 and number.n.02.
    quantity_synsets = {'00033615'}
    while True:
        added = {s for s, (_, parents, _, _) in synsets.items() if parents & quantity_synsets}
        if added <= quantity_synsets:
            break
        quantity_synsets |= added
    quantity_bases = {w for w, first in first_noun.items()
                      if first in quantity_synsets and re.fullmatch('[a-z]+', w)}
    # Prefer a surface form's own noun entry over another lemma's inflection:
    # species is a taxonomic noun, even though regular(specie) also yields it.
    semantic_quantity = {w for w in expand(quantity_bases)
                         if first_noun.get(w) in quantity_synsets or w not in first_noun}

    closed = set()
    for name in ('singularPartitiveWords', 'termDeterminers', 'quantityNumerals', 'singularClosedWords'):
        matches = []
        for path in sorted((ROOT / 'validation').glob('*.go')):
            matches += re.findall(r'var\s+' + name + r'\s*=\s*wordSet\("([^"]*)"\)', path.read_text())
        if len(matches) != 1:
            raise SystemExit('Expected exactly one wordSet declaration for ' + name)
        closed.update(matches[0].split())
    closed.update('mine yours hers ours theirs myself yourself oneself thou thee thy thine ye who whom somebody someone something somewhere anybody anyone anything anywhere nobody noone nothing nowhere everybody everyone everything everywhere whatever whoever whomever whichever wherever whenever neither either each other another such that this these those what which whose whatsoever whomsoever whosoever'.split())
    closed = expand(closed)
    compounds = {prefix + head + suffix
                 for prefix in ('every', 'some', 'any', 'no')
                 for head in ('one', 'body', 'thing', 'where')
                 for suffix in ('', 's')}

    tags = defaultdict(int)
    for line in senses.splitlines():
        key, _, _, n = line.split()
        lemma, pos = key.split('%')
        if pos.startswith('2:'):
            tags[lemma] += int(n)
    # "Common verb" = at least five tagged uses over WordNet verb senses.
    # Inflected verbs are not separate WordNet sense entries; compare bases.
    common_verbs = {v for v, n in tags.items() if n >= 5}
    common_forms = {regular(v) for v in common_verbs}
    all_forms = {regular(v) for v in verbs}
    for line in verb_exc.splitlines():
        form, *bases = line.split()
        if form.endswith('s'):
            all_forms.add(form)
            if set(bases) & common_verbs:
                common_forms.add(form)
    common_forms.update(('is', 'has', 'does'))

    categories = [
        ('explicit quantity, proportion, partitive, measure and number words', quantity),
        ('WordNet first noun sense under quantity or measure', semantic_quantity),
        ('Go closed-word sets, pronouns, determiners and quantifiers', closed),
        ('indefinite pronoun compounds', compounds),
    ]
    # Report independent matches among all attested candidates plus required
    # terms. Categories overlap; counts are not a disjoint subtraction ledger.
    pool = candidates & ranks.keys() | REQUIRED
    excluded = set().union(*(values for _, values in categories))
    eligible = sorted(pool - excluded, key=rank)
    missing = REQUIRED - set(eligible)
    if missing:
        raise SystemExit('Required terms excluded: ' + ', '.join(sorted(missing)))
    chosen = set(eligible[:LIMIT]) | REQUIRED
    chosen_ranked = sorted(chosen, key=rank)
    OUT.mkdir(parents=True, exist_ok=True)
    header = ['# Generated noun allowlist; do not edit.',
              '# Command: python3 scripts/gen-noun-allowlist/generate.py',
              '# Sources (URLs and SHA256: scripts/gen-noun-allowlist/sources.json):']
    header += ['# ' + s['name'] + '; ' + s['version'] + '; ' + s['license'] for s in sources if s['file'] in ('word_count.txt', 'wordnet.zip', 'enwiki-2023-04-13.txt')]
    header += ['# Manual mass nouns and required domain vocabulary: see generator README.md.',
               '# Words: ' + str(len(chosen))]
    (OUT / 'plural_mass_nouns.txt').write_text('\n'.join(header + sorted(chosen)) + '\n')
    report = ['# Exclusions', '',
              'Generated by `python3 scripts/gen-noun-allowlist/generate.py`.', '',
              f'Counts are independent matches in {len(pool):,} frequency-attested candidates plus required terms; categories overlap.',
              'Examples are 30 entries from each global deny set, in frequency order (unattested entries last).',
              'The global deny sets also cover words that never qualified as candidates.', '']
    for label, values in categories:
        matches = sorted(pool & values, key=rank)
        report += [f'## {label}', '', f'Count: {len(matches)}. Global deny-set size: {len(values)}.', '',
                   ', '.join(sorted(values, key=rank)[:30]), '']
        print(f'{label}: {len(matches)}')
    unranked = sorted(chosen - ranks.keys())
    review = sorted({w for w in chosen if w.endswith('s')} & common_forms, key=rank)
    uncommon = sorted({w for w in chosen if w.endswith('s')} & (all_forms - common_forms), key=rank)
    report += ['## Selection boundary', '',
               f'Eligible: {len(eligible)}. Included: {len(chosen)}. Below size cutoff: {len(set(eligible) - chosen)}.', '',
               'Required terms without frequency observations: ' + (', '.join(unranked) or 'none') + '.', '',
               '## Verb/noun overlaps (informational only; never excluded)', '',
               f'Common verb overlaps in candidate pool: {len(pool & common_forms)}. Included: {len(review)}.', '',
               '## Included common verb forms ending in -s', '', ', '.join(review) or 'None.', '',
               '## Included rare or untagged WordNet verb forms ending in -s', '',
               ', '.join(uncommon) or 'None.', '',
               'All verb overlaps remain eligible. These verb lemmas have fewer than five tagged uses. This is a corpus-relative definition, not a guarantee of everyday rarity.', '',
               '## 50 most frequent included words', '', ', '.join(chosen_ranked[:50]), '',
               '## 50 least frequent included words', '', ', '.join(chosen_ranked[-50:]), '']
    (OUT / 'EXCLUDED.md').write_text('\n'.join(report))
    (OUT / 'LICENSE-WordNet.txt').write_text(wn_license)
    (OUT / 'LICENSE-frequency.txt').write_bytes((args.cache / 'frequency-LICENSE').read_bytes())
    (OUT / 'LICENSE-wikipedia-frequency.txt').write_bytes((args.cache / 'wikipedia-LICENSE').read_bytes())
    print(f'Final word count: {len(chosen)}')
    print('50 most frequent: ' + ', '.join(chosen_ranked[:50]))
    print('50 least frequent: ' + ', '.join(chosen_ranked[-50:]))
    print(f'Common verb overlaps: {len(pool & common_forms)} candidates, {len(review)} included (informational only)')
    print(f'Rare/untagged verb overlaps: {len(uncommon)} (full review list in EXCLUDED.md)')
    if not 8000 <= len(chosen) <= 10000:
        raise SystemExit('Lexicon outside required size range')


if __name__ == '__main__':
    main()
