import json, collections, statistics, datetime, re
from lib import *
cards, findings, meta, assigns = load()
commits = json.load(open(OUT + '/commits.json'))

def norm(p):
    return p[len('internal/template/templates/'):] if p.startswith('internal/template/templates/') else p
def process(p):
    return p.startswith('.moai/') or p in ('CHANGELOG.md',) or p.startswith('.claude/agent-memory')
per = {}
for c in commits:
    if not c['ids']: continue
    cid = c['ids'][0]
    e = per.setdefault(cid, {'files': collections.Counter(), 'allfiles': set(), 'add': 0, 'del': 0, 'padd': 0, 'pdel': 0, 'dates': [], 'n': 0, 'inner': 0, 'subj': c['subj']})
    e['n'] += 1; e['inner'] += c['ncommits']; e['dates'].append(c['date'])
    for p, a, d in c['files']:
        e['allfiles'].add(p)
        if process(p): e['padd'] += a; e['pdel'] += d
        else:
            e['files'][norm(p)] += a + d; e['add'] += a; e['del'] += d
print('cards with >=1 first-parent develop commit:', len(per), '| of which present in store:', sum(1 for c in per if c in cards))
done = [c for c in cards.values() if c['where'] == 'archived']
print('archived(done) cards:', len(done), 'with delivery commit found:', sum(1 for c in done if c['id'] in per), '(%.0f%%)' % (100 * sum(1 for c in done if c['id'] in per) / len(done)))

def q(v):
    v = sorted(v); n = len(v)
    return [v[0], v[n // 4], v[n // 2], v[3 * n // 4], v[int(n * .9)], v[-1]]
S = list(per.values())
lines = [e['add'] + e['del'] for e in S]
files = [len(e['files']) for e in S]
plines = [e['padd'] + e['pdel'] for e in S]
tot = [e['add'] + e['del'] + e['padd'] + e['pdel'] for e in S]
print('\n[min, Q1, median, Q3, P90, max]')
print('changed lines, product (excl .moai/**, CHANGELOG; template mirror paths folded):', q(lines))
print('files, product (mirror-folded):', q(files))
print('changed lines, process artifacts (.moai/** specs/reports etc):', q(plines))
print('changed lines total:', q(tot))
print('first-parent commits per card:', q([e['n'] for e in S]), '| inner branch commits per card:', q([e['inner'] for e in S]))
n = len(S)
def pc(x): return '%d (%.0f%%)' % (x, 100 * x / n)
print('product lines == 0 (process-only delivery):', pc(sum(1 for l in lines if l == 0)))
print('product lines < 50:', pc(sum(1 for l in lines if l < 50)))
print('product lines < 150:', pc(sum(1 for l in lines if l < 150)))
print('product files <= 3:', pc(sum(1 for f in files if f <= 3)))
print('product lines < 50 AND files <= 3:', pc(sum(1 for l, f in zip(lines, files) if l < 50 and f <= 3)))
tp = sum(lines); tpr = sum(plines)
print('total product lines %d vs process-artifact lines %d  (process share %.0f%%)' % (tp, tpr, 100 * tpr / (tp + tpr)))
ratio = sorted((e['padd'] + e['pdel']) / max(1, e['add'] + e['del']) for e in S if e['add'] + e['del'] > 0)
print('per-card process/product line ratio: median %.1f Q3 %.1f' % (ratio[len(ratio) // 2], ratio[3 * len(ratio) // 4]))

def ts(s): return datetime.datetime.fromisoformat(s)
for cid, e in per.items(): e['t'] = max(ts(d) for d in e['dates'])
ids = sorted(per, key=lambda c: per[c]['t'])
# same file set within 48h
fs = {c: frozenset(per[c]['files']) for c in ids}
pairs_exact = set(); pairs_j = set(); cards_j = set(); cards_sub = set()
for i, a in enumerate(ids):
    if not fs[a]: continue
    for b in ids[i + 1:]:
        if (per[b]['t'] - per[a]['t']).total_seconds() > 48 * 3600: break
        if not fs[b]: continue
        inter = len(fs[a] & fs[b]); un = len(fs[a] | fs[b])
        if fs[a] == fs[b]: pairs_exact.add((a, b))
        if inter / un >= 0.5: pairs_j.add((a, b)); cards_j |= {a, b}
        if inter and (fs[a] <= fs[b] or fs[b] <= fs[a]): cards_sub |= {a, b}
print('\nwithin 48h: exact same product file set pairs: %d (cards %d); Jaccard>=0.5 pairs: %d (cards %s); one set subset of the other: cards %s' % (
    len(pairs_exact), len({x for p in pairs_exact for x in p}), len(pairs_j), pc(len(cards_j)), pc(len(cards_sub))))

# primary file groups within 3 days
GENERIC = re.compile(r'(^|/)(go\.sum|go\.mod|CHANGELOG|README)|_templ\.go$|\.golden$|census|ac-count-baseline')
prim = {}
for c in ids:
    f = [(p, v) for p, v in per[c]['files'].items() if not GENERIC.search(p)]
    if f: prim[c] = max(f, key=lambda x: x[1])[0]
byfile = collections.defaultdict(list)
for c in ids:
    if c in prim: byfile[prim[c]].append(c)
groups = []
for f, cs in byfile.items():
    cur = [cs[0]]
    for c in cs[1:]:
        if (per[c]['t'] - per[cur[-1]]['t']).total_seconds() <= 72 * 3600: cur.append(c)
        else:
            if len(cur) > 1: groups.append((f, cur))
            cur = [c]
    if len(cur) > 1: groups.append((f, cur))
saved = sum(len(g) - 1 for _, g in groups)
print('same PRIMARY file (most-changed product file) chained within 72h: groups %d, cards %s, lane-sessions saveable if each group were one card: %d (%.0f%% of delivered cards)' % (
    len(groups), pc(sum(len(g) for _, g in groups)), saved, 100 * saved / n))
gs = collections.Counter(len(g) for _, g in groups); print('  group size distribution:', sorted(gs.items()))
# stricter: also require derivation link or text citation between members
graph = json.load(open(OUT + '/graph.json'))
def linked(a, b):
    return b in refs(cards[a]['text'], a) or a in refs(cards[b]['text'], b) if a in cards and b in cards else False
strict = 0; sg = 0
for f, g in groups:
    # union-find over citation links
    par = {x: x for x in g}
    def find(x):
        while par[x] != x: x = par[x]
        return x
    for i, a in enumerate(g):
        for b in g[i + 1:]:
            if linked(a, b): par[find(a)] = find(b)
    comp = collections.Counter(find(x) for x in g)
    s = sum(v - 1 for v in comp.values()); strict += s; sg += sum(1 for v in comp.values() if v > 1)
print('  stricter (same primary file within 72h AND one card cites the other): sub-groups %d, saveable lane-sessions %d (%.0f%%)' % (sg, strict, 100 * strict / n))
print('  top groups:')
for f, g in sorted(groups, key=lambda x: -len(x[1]))[:10]:
    print('   %2d  %s  %s  [%s..%s]' % (len(g), f[-60:], ' '.join(g[:14]), per[g[0]]['t'].date(), per[g[-1]]['t'].date()))
# hot files: distinct cards touching per file (any touch) & per 3-day window max
touch = collections.defaultdict(list)
for c in ids:
    for p in per[c]['files']:
        if not GENERIC.search(p): touch[p].append(c)
print('\nhottest product files (distinct cards touching, whole period; max cards in any 72h window):')
def maxwin(cs):
    best = 0; j = 0
    for i in range(len(cs)):
        while (per[cs[i]]['t'] - per[cs[j]]['t']).total_seconds() > 72 * 3600: j += 1
        best = max(best, i - j + 1)
    return best
for p, cs in sorted(touch.items(), key=lambda kv: -len(kv[1]))[:12]:
    print('   %3d  win72h=%2d  %s' % (len(cs), maxwin(cs), p[-70:]))
chk = [c for c in ('t1414', 't1441', 't1375') if c in per]
print('operator example launcher.go:', [(c, str(per[c]['t'])[:16], per[c]['files'].get('internal/cli/launcher.go')) for c in chk])

# completion per day (last delivery commit, UTC date of commit) and distinct WT branches
comp = collections.Counter(str(per[c]['t'].astimezone(datetime.timezone.utc).date()) for c in ids)
inv = json.load(open(OUT + '/inventory.json'))
days = sorted(set(comp) | set(inv['day']))
vals = [comp[d] for d in days if comp[d]]
print('\ncards delivered per day (git, last first-parent commit): active days %d median %s mean %.1f max %d' % (len(vals), statistics.median(vals), statistics.mean(vals), max(vals)))
print('last 14 days  date issued/delivered:', ' '.join('%s %d/%d' % (d[5:], inv['day'].get(d, 0), comp[d]) for d in days[-14:]))
wk = collections.defaultdict(lambda: [0, 0])
for d in days:
    y, w, _ = datetime.date.fromisoformat(d).isocalendar(); wk['W%02d' % w][0] += inv['day'].get(d, 0); wk['W%02d' % w][1] += comp[d]
print('week issued/delivered:', {k: tuple(v) for k, v in sorted(wk.items())})
json.dump({c: {'t': str(per[c]['t']), 'lines': per[c]['add'] + per[c]['del'], 'files': sorted(per[c]['files']), 'prim': prim.get(c)} for c in ids}, open(OUT + '/delivered.json', 'w'))
json.dump([(f, g) for f, g in groups], open(OUT + '/groups.json', 'w'))
