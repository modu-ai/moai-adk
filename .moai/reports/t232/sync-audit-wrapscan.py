import re

def parse(path):
    entries = []
    cur = None
    for line in open(path, encoding='utf-8'):
        m = re.match(r'^- id: (CONST-\S+)', line)
        if m:
            cur = {'id': m.group(1)}
            entries.append(cur)
            continue
        if cur is None:
            continue
        for f in ('file', 'anchor', 'clause'):
            m = re.match(r'^  %s: (.*)$' % f, line)
            if m:
                v = m.group(1).strip()
                if len(v) >= 2 and v[0] == '"' and v[-1] == '"':
                    v = v[1:-1].replace('\\"', '"')
                cur[f] = v
    return entries

TERM = tuple('.!?:;"”’)]`')  # sentence-terminal characters
reg = '.claude/rules/moai/core/zone-registry.md'
wrap = []
midline = []
for e in parse(reg):
    clause = e.get('clause', '')
    if '[SUPERSEDED' in clause or not clause:
        continue
    lines = open(e['file'], encoding='utf-8').read().splitlines()
    hit_idx = next((i for i, line in enumerate(lines) if clause in line), None)
    if hit_idx is None:
        continue
    line = lines[hit_idx]
    pos = line.index(clause) + len(clause)
    rest = line[pos:].strip()
    if rest:
        midline.append((e['id'], 'rest=%r' % rest[:40]))
        continue
    if clause.rstrip()[-1:] in TERM:
        continue
    nxt = next((line for line in lines[hit_idx + 1:] if line.strip()), '')
    if nxt:
        wrap.append((e['id'], 'clause-tail=%r' % clause[-25:], 'next=%r' % nxt.strip()[:45]))

print('midline-truncation=%d wrap-suspect=%d' % (len(midline), len(wrap)))
for x in midline:
    print(' MIDLINE', x)
for x in wrap:
    print(' WRAP', x)
