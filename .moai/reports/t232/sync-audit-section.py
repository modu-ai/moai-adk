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

def heading_slug(text):
    # six-step rule from spec.md 2.2
    text = text.strip()
    text = text.replace('`', '')
    text = text.lower()
    text = ''.join(c if (c.isalnum() and ord(c) < 128) or c in ' -' else '' for c in text)
    text = re.sub(r'\s+', '-', text.strip())
    return '#' + text

def headings_of(path):
    # list of (line_no, level, slug, title) excluding fenced code blocks
    out = []
    fenced = False
    for i, line in enumerate(open(path, encoding='utf-8'), 1):
        if line.lstrip().startswith('```'):
            fenced = not fenced
            continue
        if fenced:
            continue
        m = re.match(r'^(#{1,6})\s+(.*)$', line.rstrip())
        if m:
            title = m.group(2)
            out.append((i, len(m.group(1)), heading_slug(title), title))
    return out

def section_span(hs, slug):
    # span = from the anchored heading line to the next heading of same-or-lower level
    for idx, (i, lvl, s, t) in enumerate(hs):
        if s == slug:
            end = None
            for (j, l2, s2, t2) in hs[idx + 1:]:
                if l2 <= lvl:
                    end = j
                    break
            return (i, end if end is not None else 10**9)
    return None

def nearest_heading(hs, lineno):
    best = None
    for (i, lvl, slug, title) in hs:
        if i <= lineno:
            best = (i, slug, title)
        else:
            break
    return best

reg = '.claude/rules/moai/core/zone-registry.md'
entries = parse(reg)
anchor_fail = []
section_mismatch = []
checked = 0
for e in entries:
    f = e.get('file', '')
    anchor = e.get('anchor', '')
    clause = e.get('clause', '')
    hs = headings_of(f)
    slugs = set(s for (_, _, s, _) in hs)
    if anchor not in slugs:
        anchor_fail.append((e['id'], anchor, f))
        continue
    if '[SUPERSEDED' in clause or not clause:
        continue
    src_lines = open(f, encoding='utf-8').read().splitlines()
    hit = next((i + 1 for i, line in enumerate(src_lines) if clause in line), None)
    if hit is None:
        section_mismatch.append((e['id'], 'NO-HIT', anchor))
        continue
    checked += 1
    span = section_span(hs, anchor)
    if span is None or not (span[0] <= hit < span[1]):
        nh = nearest_heading(hs, hit)
        section_mismatch.append((e['id'], 'clause-under=%s' % (nh[1] if nh else 'NONE'), 'anchor=%s' % anchor, clause[:50]))

print('entries=%d anchor_unresolved=%d clause_section_checked=%d section_mismatch=%d' % (
    len(entries), len(anchor_fail), checked, len(section_mismatch)))
for x in anchor_fail:
    print(' ANCHOR-FAIL', x)
for x in section_mismatch:
    print(' MISMATCH', x)
