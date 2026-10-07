import collections, json, statistics
from lib import *
cards, findings, meta, assigns = load()

def header(t):
    """leading bracket tags (after stripping a [DROPPED ...] tag) or first 160 chars."""
    t = re.sub(r'^\[DROPPED[^\]]*\]\s*', '', t)
    m = re.match(r'(\s*\[[^\]]*\]\s*){1,4}', t)
    h = m.group(0) if m else ''
    return (h + ' ' + t[len(h):len(h) + 120]) if h else t[:160]

PAT = collections.OrderedDict([
    ('standing[PROJECT/GRAPH]', r'^\s*\[(PROJECT|GRAPH)\]'),
    ('audit-derived', r'감사\s*(후속|F\d|P\d|G\d|H\d)|sync-au|plan-au|WORKFLOW-AUDIT|hooks 감사|지침 감사|감사 후속|audit.{0,12}(F\d|finding|follow)|범위 밖 발견|codex 게이트|게이트 P\d'),
    ('follow-up/derived', r'후속|파생|follow-?up|분리분|분할|레인 상신|잔여|발견\]|발견 ·|재발행|이월|승계'),
    ('ci-repair', r'CI\s*(적색|상속|red|수리|실패)|적색|red 축|develop CI|회귀 수리|CI 수리'),
    ('report-milestone', r'보고서\s*[PM]\d|\[REPORT\]|보고서 M\d|설계 \d/\d'),
    ('external(issue/CC)', r'ISSUE #|GitHub #|^\s*\[#\d|CC \d+\.\d+|CC-UPDATE|PR #\d'),
    ('operator-named', r'운영자\s*(요청|지시|결정|승인|판정|골|선택|재현|배차)'),
    ('leader-issued', r'리드 발행|리드 기록|lead'),
])
lab = {}
multi = collections.Counter(); prim = collections.Counter()
for cid, c in cards.items():
    h = header(c['text'])
    ls = [k for k, p in PAT.items() if re.search(p, h, re.I)]
    lab[cid] = ls
    for l in ls: multi[l] += 1
    prim[ls[0] if ls else 'unlabeled'] += 1
N = len(cards)
print('N', N)
print('multi-label (header text pattern):')
for k in list(PAT) + []: print('  %-26s %4d %5.1f%%' % (k, multi[k], 100 * multi[k] / N))
print('primary (first match in priority order):')
for k, v in prim.most_common(): print('  %-26s %4d %5.1f%%' % (k, v, 100 * v / N))
both = sum(1 for l in lab.values() if 'leader-issued' in l and 'operator-named' in l)
print('leader-issued AND operator-named', both, '| leader-issued w/o operator mention', multi['leader-issued'] - both)

# ---- citation graph
allref = {cid: {r for r in refs(c['text'], cid)} for cid, c in cards.items()}
hdrref = {cid: {r for r in refs(header(c['text']), cid)} for cid, c in cards.items()}
DER = re.compile(r'후속|파생|follow-?up|분리|분할|발견|상신|감사|잔여|수리|재발행|이월|승계|회귀|실측', re.I)
parents = {}
for cid, c in cards.items():
    h = header(c['text'])
    ps = {r for r in hdrref[cid] if num(r) < num(cid)} if DER.search(h) else set()
    parents[cid] = ps
cite_any = sum(1 for v in allref.values() if v)
cite_earlier = sum(1 for cid, v in allref.items() if any(num(r) < num(cid) for r in v))
print('\ncards citing >=1 other card anywhere in text: %d (%.1f%%); citing an earlier card: %d (%.1f%%)' % (cite_any, 100 * cite_any / N, cite_earlier, 100 * cite_earlier / N))
der = [c for c, p in parents.items() if p]
print('cards with a derivation parent (header cites earlier card + derivation keyword): %d (%.1f%%)' % (len(der), 100 * len(der) / N))
dangling = sum(1 for c in der for p in parents[c] if p not in cards)
print('parent refs to ids with no store record:', dangling)

depth = {}
def d(cid, seen=()):
    if cid in depth: return depth[cid]
    ps = [p for p in parents.get(cid, ()) if p not in seen]
    v = 0 if not ps else 1 + max(d(p, seen + (cid,)) for p in ps)
    depth[cid] = v; return v
import sys; sys.setrecursionlimit(10000)
for c in sorted(cards, key=num): d(c)
dd = collections.Counter(depth[c] for c in cards)
print('derivation depth distribution (0 = root):', sorted(dd.items()))
ge2 = sum(v for k, v in dd.items() if k >= 2); ge3 = sum(v for k, v in dd.items() if k >= 3)
print('follow-up of follow-up (depth>=2): %d (%.1f%%); depth>=3: %d (%.1f%%)' % (ge2, 100 * ge2 / N, ge3, 100 * ge3 / N))
# by issue week: share derived
import datetime
wk = collections.defaultdict(lambda: [0, 0, 0])
for cid, c in cards.items():
    y, w, _ = datetime.date.fromisoformat(c['added_at'][:10]).isocalendar()
    k = 'W%02d' % w; wk[k][0] += 1; wk[k][1] += bool(parents[cid]); wk[k][2] += depth[cid] >= 2
print('week: issued / derived / depth>=2')
for k in sorted(wk): print(' ', k, wk[k], '%.0f%%' % (100 * wk[k][1] / wk[k][0]))

# children count (fan-out)
children = collections.defaultdict(set)
for c, ps in parents.items():
    for p in ps: children[p].add(c)
fan = sorted(children.items(), key=lambda kv: -len(kv[1]))[:12]
print('\ntop fan-out parents (direct derived children):')
for p, ch in fan:
    t = cards[p]['text'][:90].replace('\n', ' ') if p in cards else '<no record>'
    print('  %s -> %d : %s' % (p, len(ch), t))

# repair-of-repair
REP = re.compile(r'수리|적색|회귀|regress|복구|깨|고장|결함|버그|bug|fix|red\b', re.I)
isrep = {cid: bool(REP.search(header(c['text']))) for cid, c in cards.items()}
r1 = [c for c in der if isrep[c]]
r2 = [c for c in r1 if any(isrep.get(p) and parents.get(p) for p in parents[c])]
r2b = [c for c in r1 if any(isrep.get(p) for p in parents[c])]
print('\nderived cards whose header has a repair keyword: %d; whose parent is also a repair-keyword card: %d; ...and grandparent exists: %d' % (len(r1), len(r2b), len(r2)))
aud = [c for c in cards if 'audit-derived' in lab[c]]
aud_der = [c for c in aud if parents[c]]
opt = [c for c in cards if re.search(r'선택\s*(발견|항목|사항)|optional|비차단|non-?blocking|P[23]\b|minor|경미|PASS-WITH-DEBT|부채', header(cards[c]['text']), re.I)]
print('audit-derived header: %d (of which cite a parent card: %d); headers flagged optional/minor/P2-3/debt: %d; both: %d' % (len(aud), len(aud_der), len(opt), len(set(aud) & set(opt))))

# longest chains: reconstruct
def chain(cid):
    out = [cid]
    while parents.get(out[-1]):
        ps = [p for p in parents[out[-1]] if p not in out]
        if not ps: break
        out.append(max(ps, key=lambda p: depth.get(p, 0)))
    return out
tops = sorted(cards, key=lambda c: -depth[c])
used = set(); shown = 0
print('\nlongest derivation chains (leaf <- ... <- root):')
for c in tops:
    ch = chain(c)
    if set(ch) & used: continue
    used |= set(ch); shown += 1
    print(' depth %d: %s' % (depth[c], ' <- '.join(ch)))
    for x in ch:
        t = re.sub(r'\s+', ' ', cards[x]['text'])[:105] if x in cards else '<no record>'
        print('     %s [%s] %s' % (x, fate(cards[x]) if x in cards else '-', t))
    if shown >= 6: break

# connected components over any-citation graph
adj = collections.defaultdict(set)
for c, rs in allref.items():
    for r in rs:
        if r in cards: adj[c].add(r); adj[r].add(c)
seen = set(); comps = []
for c in cards:
    if c in seen: continue
    st = [c]; comp = []
    seen.add(c)
    while st:
        x = st.pop(); comp.append(x)
        for y in adj[x]:
            if y not in seen: seen.add(y); st.append(y)
    comps.append(comp)
sz = collections.Counter(len(c) for c in comps)
print('\nany-citation components: n=%d singletons=%d size2-5=%d size6-20=%d size>20=%d largest=%d' % (
    len(comps), sz[1], sum(v for k, v in sz.items() if 2 <= k <= 5), sum(v for k, v in sz.items() if 6 <= k <= 20), sum(v for k, v in sz.items() if k > 20), max(sz)))
# derivation-only components (tighter topic clusters)
adj2 = collections.defaultdict(set)
for c, ps in parents.items():
    for p in ps:
        if p in cards: adj2[c].add(p); adj2[p].add(c)
seen = set(); comps2 = []
for c in cards:
    if c in seen or c not in adj2: continue
    st = [c]; comp = []; seen.add(c)
    while st:
        x = st.pop(); comp.append(x)
        for y in adj2[x]:
            if y not in seen: seen.add(y); st.append(y)
    comps2.append(sorted(comp, key=num))
comps2.sort(key=len, reverse=True)
sz2 = [len(c) for c in comps2]
print('derivation components: n=%d cards=%d sizes top10=%s median=%s' % (len(comps2), sum(sz2), sz2[:10], statistics.median(sz2)))
for comp in comps2[:7]:
    root = comp[0]
    span = (cards[comp[0]]['added_at'][:10], cards[comp[-1]]['added_at'][:10])
    st = collections.Counter(fate(cards[x]) for x in comp)
    print('  size %d root %s span %s..%s states %s' % (len(comp), root, span[0], span[1], dict(st)))
    print('     root:', re.sub(r'\s+', ' ', cards[root]['text'])[:110])
    print('     ids:', ' '.join(comp[:28]), '...' if len(comp) > 28 else '')
json.dump({'parents': {k: sorted(v) for k, v in parents.items()}, 'depth': depth, 'labels': lab,
           'comps2': comps2}, open(OUT + '/graph.json', 'w'))
