"""Read-only drift analyzer for the zone-registry (t232 evidence tooling).

Usage: python3 analyze.py <project-root> <out.json>

Mirrors internal/constitution/validator.go's DRIFT check (normalizeWhitespace +
stripCodeFences + exact substring) and additionally resolves each entry's anchor
against the real heading slugs of its named file.
"""
import json
import os
import re
import sys

root = sys.argv[1]
reg = os.path.join(root, '.claude/rules/moai/core/zone-registry.md')
txt = open(reg).read()
entries = []
cur = None
for line in txt.split('\n'):
    m = re.match(r'^- id: (\S+)', line)
    if m:
        cur = {'id': m.group(1)}
        entries.append(cur)
        continue
    if cur is None:
        continue
    m = re.match(r'^  (\w+): (.*)$', line)
    if m:
        k, v = m.group(1), m.group(2).strip()
        if v.startswith('"') and v.endswith('"'):
            v = v[1:-1]
        cur[k] = v


def strip_fences(c):
    out = []
    inf = False
    for ln in c.split('\n'):
        if ln.strip().startswith('```'):
            inf = not inf
            out.append('')
            continue
        out.append("" if inf else ln)
    return '\n'.join(out)


def norm(s):
    return re.sub(r'\s+', ' ', s).strip()


def slug(h):
    h = h.strip().lstrip('#').strip().replace('`', '').lower()
    h = re.sub(r'[^a-z0-9\s-]', '', h)
    return '#' + re.sub(r'\s+', '-', h.strip())


cache = {}
res = []
for e in entries:
    f = e.get('file', '')
    p = os.path.join(root, f)
    if p not in cache:
        cache[p] = open(p).read() if os.path.exists(p) else None
    src = cache[p]
    if src is None:
        res.append({'id': e['id'], 'file': f, 'anchor': e.get('anchor'),
                    'clause_ok': False, 'anchor_ok': False,
                    'clause': e.get('clause', ''), 'missing_file': True})
        continue
    ok_clause = norm(e.get('clause', '')) in norm(strip_fences(src))
    slugs = set()
    inf = False
    for ln in src.split('\n'):
        if ln.strip().startswith('```'):
            inf = not inf
            continue
        if not inf and ln.startswith('#'):
            slugs.add(slug(ln))
    ok_anchor = e.get('anchor', '') in slugs
    res.append({'id': e['id'], 'file': f, 'anchor': e.get('anchor'),
                'clause_ok': ok_clause, 'anchor_ok': ok_anchor,
                'clause': e.get('clause', ''), 'missing_file': False})

print("entries=%d clause_fail=%d anchor_fail=%d" % (
    len(res),
    sum(1 for r in res if not r['clause_ok']),
    sum(1 for r in res if not r['anchor_ok'])))
print("both_fail=%d clause_fail_anchor_ok=%d clause_ok_anchor_fail=%d missing_file=%d" % (
    sum(1 for r in res if not r['clause_ok'] and not r['anchor_ok']),
    sum(1 for r in res if not r['clause_ok'] and r['anchor_ok']),
    sum(1 for r in res if r['clause_ok'] and not r['anchor_ok']),
    sum(1 for r in res if r['missing_file'])))
json.dump(res, open(sys.argv[2], 'w'), indent=1)
