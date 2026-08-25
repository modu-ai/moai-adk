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
                # strip surrounding YAML double quotes (clause values are quoted scalars)
                if len(v) >= 2 and v[0] == '"' and v[-1] == '"':
                    v = v[1:-1].replace('\\"', '"')
                cur[f] = v
    return entries

def check(reg_path, root):
    entries = parse(reg_path)
    once = zero = multi = retired = selfref = 0
    fails = []
    empty = 0
    for e in entries:
        clause = e.get('clause', '')
        f = e.get('file', '')
        if not clause:
            empty += 1
            fails.append((e['id'], 'EMPTY-CLAUSE'))
            continue
        if f.endswith('zone-registry.md'):
            selfref += 1
            continue
        src = open(root + '/' + f, encoding='utf-8').read()
        n = src.count(clause)
        if '[SUPERSEDED' in clause:
            retired += 1
        elif n == 1:
            once += 1
        elif n == 0:
            zero += 1
            fails.append((e['id'], 'zero', clause[:60]))
        else:
            multi += 1
            fails.append((e['id'], 'multi-x%d' % n, clause[:60]))
    return dict(total=len(entries), once=once, zero=zero, multi=multi,
                retired_exempt=retired, selfref=selfref, empty=empty), fails

for label, reg, root in [
    ('LOCAL', '.claude/rules/moai/core/zone-registry.md', '.'),
    ('TMPL', 'internal/template/templates/.claude/rules/moai/core/zone-registry.md', 'internal/template/templates')]:
    stats, fails = check(reg, root)
    print(label, stats)
    for f in fails[:10]:
        print('  FAIL', f)
