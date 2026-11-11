#!/usr/bin/env python3
"""Run in a clean v24 checkout to prepare the read-only Go test overlay.

Then run, with the usual sandbox cache variables:
  ARGCHECK_CAPTURE=$PWD/.cache/legacy-capture.jsonl go test -vet=off -overlay tmp/capture_overlay.json ./validation

This captures accepted runtime arguments, including dynamically constructed
catalog tests and golden fixtures, without adding hooks to production code.
Use --collect afterward to write the distinct readable sentence inventory.
Run the v25 coverage test with ARGCHECK_UPDATE_LEGACY=1 to record exceptions.
The output's forms are argument-level provenance, not a claim that each
sentence independently exercises each of those rules. Explicit v24 binders
and arbitrary-name declarations are excluded; they are not legacy sentences.
"""
import sys
import json,pathlib
root=pathlib.Path.cwd()
(root/'tmp').mkdir(exist_ok=True)
(root/'.cache').mkdir(exist_ok=True)
(root/'tmp/go.mod').write_text('module scratch\n\ngo 1.25.0\n')
if '--collect' in sys.argv:
    import re
    rows = {}
    for line in (root/'.cache/legacy-capture.jsonl').read_text().splitlines():
        record = json.loads(line)
        forms = {s['form'] for s in record['Report']['steps'] if s.get('form')}
        for step in record['Argument']['steps']:
            sentence = step['statement']
            if step['kind'] == 'arbitrary' or re.search(r'\b(every|some|no) [a-z] is such that\b', sentence, re.I):
                continue
            rows.setdefault(sentence, set()).update(forms)
    output = [{'sentence': s, 'forms': sorted(forms)} for s, forms in sorted(rows.items())]
    (root/'validation/testdata/legacy_v24_sentences.json').write_text(json.dumps(output, indent=2, ensure_ascii=False)+'\n')
    print(f'{len(output)} distinct readable sentences')
    sys.exit(0)
p=root/'validation/entailment_gate.go'
if 'EntailmentGateCatalogVersion = "v24"' not in p.read_text():
    sys.exit('Capture must run against the v24 baseline, before bridge changes.')
(root/'.cache/legacy-capture.jsonl').write_text('')
s=p.read_text().replace('"context"','"context"\n "encoding/json"\n "os"\n "sync"')
s=s.replace('func GateNumberedArgument(arg NumberedArgument) EntailmentGateReport {','func GateNumberedArgument(arg NumberedArgument) (captured EntailmentGateReport) {\n defer func() { if captured.Passed { captureMu.Lock(); defer captureMu.Unlock(); f, err := os.OpenFile(os.Getenv("ARGCHECK_CAPTURE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600); if err != nil { panic(err) }; defer f.Close(); json.NewEncoder(f).Encode(struct { Argument NumberedArgument; Report EntailmentGateReport }{arg, captured}) } }()')
s+='\nvar captureMu sync.Mutex\n'
out=root/'tmp/capture_gate.go';out.write_text(s)
(root/'tmp/capture_overlay.json').write_text(json.dumps({'Replace':{str(p):str(out)}}))
