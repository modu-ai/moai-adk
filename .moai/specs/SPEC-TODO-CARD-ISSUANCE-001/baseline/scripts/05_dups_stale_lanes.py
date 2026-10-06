import json, collections, statistics, datetime, math
from lib import *
cards, findings, meta, assigns = load()
delivered = json.load(open(OUT + '/delivered.json'))
NOW = datetime.datetime(2026, 10, 2, 11, 40, tzinfo=datetime.timezone.utc)
def ts(s):
    return datetime.datetime.fromisoformat(s.replace('Z', '+00:00'))
OPEN = {'queued', 'picked', 'hold'}
def isopen(cid): return cid in cards and cards[cid]['where'] == 'live' and cards[cid]['state'] in OPEN
def st(cid): return fate(cards[cid]) if cid in cards else 'no-record'

print('== 4. findings ==')
print('rows: live %d archived %d' % (sum(1 for f in findings if f['where'] == 'live'), sum(1 for f in findings if f['where'] == 'archived')))
c = collections.Counter((f['relation'], f['source']) for f in findings)
for k, v in sorted(c.items()): print('  %-15s %-6s %3d' % (k[0], k[1], v))
pairs = {}
for f in findings:
    k = (frozenset((f['subject_id'], f['related_id'])), f['relation'])
    pairs.setdefault(k, f)
print('distinct (pair, relation):', len(pairs), '| distinct cards named:', len({x for k in pairs for x in k[0]}), 'of', len(cards))
nd = [f for (k, rel), f in pairs.items() if rel == 'near-duplicate']
sc = sorted(f['score'] for f in nd)
print('near-duplicate (jev) scores: n=%d min %.2f Q1 %.2f med %.2f Q3 %.2f max %.2f; >=0.8: %d; <0.5: %d' % (len(sc), sc[0], sc[len(sc) // 4], sc[len(sc) // 2], sc[3 * len(sc) // 4], sc[-1], sum(1 for s in sc if s >= .8), sum(1 for s in sc if s < .5)))
days = collections.Counter(f['at'][:10] for f in pairs.values())
print('finding dates: first %s last %s; days with findings %d' % (min(days), max(days), len(days)))
out = collections.Counter(); both_open = []; noaction = []
for (k, rel), f in pairs.items():
    a, b = f['subject_id'], f['related_id']
    sa, sb = st(a), st(b)
    cite = (a in cards and b in refs(cards[a]['text'], a)) or (b in cards and a in refs(cards[b]['text'], b))
    dropped = 'dropped' in (sa, sb)
    if isopen(a) and isopen(b): both_open.append((a, b, rel, f['score'])); out['both still open'] += 1
    elif dropped: out['one side dropped'] += 1
    elif isopen(a) or isopen(b): out['one open, one done'] += 1
    else:
        out['both done'] += 1
    if not dropped and not cite and sa.startswith('done') and sb.startswith('done'):
        noaction.append((a, b, rel, f['score']))
    out['texts cite each other'] += cite
print('pair outcome:', dict(out))
print('both done, neither dropped, neither text cites the other (no trace of the finding being acted on):', len(noaction),
      '| of which near-dup score>=0.8:', sum(1 for x in noaction if x[2] == 'near-duplicate' and x[3] >= .8))
print('both open pairs:', both_open)
# were both-done near-dup pairs delivered separately touching same files?
same = 0; meas = 0
for a, b, rel, s in noaction:
    if a in delivered and b in delivered:
        meas += 1
        fa, fb = set(delivered[a]['files']), set(delivered[b]['files'])
        if fa and fb and len(fa & fb) / len(fa | fb) >= 0.3: same += 1
print('  of those, both have delivery commits: %d; product file-set Jaccard>=0.3: %d' % (meas, same))

DUP = re.compile(r'중복|동일|쌍둥이|흡수|대체|재발행|id 충돌|duplicate|absorb|supersed|이미 (착지|해소|존재|구현|shipped)|already', re.I)
PREM = re.compile(r'전제 (반증|소멸|무효)|premise', re.I)
ACC = re.compile(r'오타|우발|실수|쓰레기|사고', re.I)
dr = [c for c in cards.values() if c['state'] == 'dropped']
dc = collections.Counter()
for c in dr:
    m = re.match(r'\[DROPPED[^\]]*\]', c['text']); h = m.group(0) if m else c['text'][:120]
    if ACC.search(h): dc['accidental (typo/verb misuse)'] += 1
    elif DUP.search(h): dc['duplicate / absorbed / already done'] += 1
    elif PREM.search(h): dc['premise dead'] += 1
    elif re.search(r'운영자', h): dc['operator decision (other)'] += 1
    else: dc['other/unstated'] += 1
print('dropped cards %d by reason-tag pattern: %s' % (len(dr), dict(dc)))
absorbed_done = sum(1 for c in cards.values() if c['where'] == 'archived' and re.search(r'^\s*\[[^\]]*(흡수|중복|대체|ABSORBED|SUPERSEDED)', c['text']))
print('archived cards whose leading tag says absorbed/duplicate/replaced:', absorbed_done)
nt = collections.defaultdict(list)
for c in cards.values():
    t = re.sub(r'^\[DROPPED[^\]]*\]\s*', '', c['text']); t = re.sub(r'\s+', ' ', t).strip().lower()
    nt[t].append(c['id'])
ex = [v for v in nt.values() if len(v) > 1]
print('byte-identical (normalized) card texts: groups %d cards %d e.g. %s' % (len(ex), sum(len(v) for v in ex), ex[:4]))

print('\n== open-card text overlap (no recorded relation) ==')
op = [c for c in cards.values() if isopen(c['id'])]
TOK = re.compile(r'[A-Za-z_][A-Za-z0-9_./\-]{2,}|[가-힣]{2,}')
STOP = set('리드 발행 운영자 승인 지시 클래스 tier 카드 절차 근거 범위 run plan sync 후속 class 그리고 대한 위한 있다 없다 한다'.split())
def toks(t):
    t = re.sub(r'(?<![A-Za-z0-9])t\d{1,4}(?![0-9A-Za-z])', ' ', t)
    return {x.lower() for x in TOK.findall(t)} - STOP
T = {c['id']: toks(c['text']) for c in cards.values()}
df = collections.Counter(x for s in T.values() for x in s)
N = len(T)
idf = {x: math.log(N / v) for x, v in df.items()}
PATH = re.compile(r'[a-z0-9_./\-]+\.(go|md|sh|yaml|yml|json|js|toml)|spec-[a-z0-9\-]+|[a-z_]+/[a-z_/.\-]+')
rel = {frozenset((f['subject_id'], f['related_id'])) for f in findings}
res = []
for i, a in enumerate(op):
    for b in op[i + 1:]:
        A, B = T[a['id']], T[b['id']]
        inter = A & B
        w = sum(idf[x] for x in inter); u = sum(idf[x] for x in A | B)
        j = len(inter) / max(1, len(A | B))
        ids_shared = sorted(x for x in inter if PATH.fullmatch(x) and df[x] <= 25)
        res.append((w / u if u else 0, j, a['id'], b['id'], ids_shared))
res.sort(reverse=True)
print('open cards', len(op), 'pairs', len(res), '| pairs with a recorded finding among open cards:', sum(1 for r in res if frozenset((r[2], r[3])) in rel))
wj = sorted(r[0] for r in res)
print('idf-weighted Jaccard over open pairs: median %.3f P95 %.3f P99 %.3f max %.3f' % (wj[len(wj) // 2], wj[int(len(wj) * .95)], wj[int(len(wj) * .99)], wj[-1]))
shown = 0
for w, j, a, b, sh in res:
    if frozenset((a, b)) in rel: continue
    cite = b in refs(cards[a]['text'], a) or a in refs(cards[b]['text'], b)
    print(' %.2f/%.2f %s[%s] ~ %s[%s] cite=%s shared-ids=%d %s' % (w, j, a, cards[a]['state'], b, cards[b]['state'], 'Y' if cite else 'n', len(sh), ','.join(s[-28:] for s in sh[:3])))
    shown += 1
    if shown >= 15: break
gist = {}
for w, j, a, b, sh in res[:40]:
    for x in (a, b): gist[x] = re.sub(r'\s+', ' ', re.sub(r'^(\s*\[[^\]]*\]\s*)+', '', cards[x]['text']))[:70]
json.dump({'top': [(round(w, 3), round(j, 3), a, b, sh[:5]) for w, j, a, b, sh in res[:40]], 'gist': gist}, open(OUT + '/overlap.json', 'w'), ensure_ascii=False)

print('\n== 5. stale / blocked ==')
for s in ('picked', 'hold', 'queued'):
    L = [c for c in op if c['state'] == s]
    ages = sorted((NOW - ts(c['added_at'])).total_seconds() / 86400 for c in L)
    print('%-7s n=%d age-since-issue days: median %.1f max %.1f; >3d: %d; >7d: %d' % (s, len(L), ages[len(ages) // 2], ages[-1], sum(1 for a in ages if a > 3), sum(1 for a in ages if a > 7)))
pk = [c for c in op if c['state'] == 'picked']
pa = [(c['id'], (NOW - ts(c['picked_at'])).total_seconds() / 86400) for c in pk if c['picked_at']]
print('picked with picked_at: %d of %d; picked >3d ago: %s; picked_at missing: %s' % (len(pa), len(pk), [(i, round(d, 1)) for i, d in pa if d > 3], [c['id'] for c in pk if not c['picked_at']]))
print('hold:', [(c['id'], c['added_at'][:10], c.get('picked_by')) for c in op if c['state'] == 'hold'])
WAIT = re.compile(r'(선행|의존|대기|blocked|이후 착수|병합 후|착지 후|완료 후|다음)\s*[:：]?\s*[^.。\n]{0,40}?(?<![A-Za-z0-9])t\d{2,4}|(?<![A-Za-z0-9])t\d{2,4}[^.。\n]{0,30}(선행|의존|대기|착지 후|병합 후|완료 후|이후)')
w = [c['id'] for c in op if WAIT.search(c['text'])]
print('open cards whose text says they wait on / depend on another card (pattern):', len(w), w[:25])
wopen = 0
for cid in w:
    for r in refs(cards[cid]['text'], cid):
        if isopen(r): wopen += 1; break
print('  ...of which cite at least one card that is itself still open:', wopen)
OPQ = re.compile(r'운영자 (결정|판정|승인) (큐|대기|필수|필요)|운영자 결정 큐|운영자 승인 대기|\[보류')
oq = [c['id'] for c in op if OPQ.search(c['text'])]
print('open cards with operator-decision markers (pattern):', len(oq), oq)
# time to first pick
w1 = [(ts(c['picked_at']) - ts(c['added_at'])).total_seconds() / 3600 for c in cards.values() if c['picked_at']]
w1.sort()
print('issue->pick hours (only %d cards carry picked_at; stamp exists since ~09-30): min %.1f Q1 %.1f med %.1f Q3 %.1f max %.1f' % (len(w1), w1[0], w1[len(w1) // 4], w1[len(w1) // 2], w1[3 * len(w1) // 4], w1[-1]))
lt = sorted((ts(delivered[c]['t']) - ts(cards[c]['added_at'])).total_seconds() / 3600 for c in delivered if c in cards)
lt = [x for x in lt if x >= 0]
print('issue->landed-on-develop hours (git, n=%d): Q1 %.1f med %.1f Q3 %.1f P90 %.1f; <6h: %.0f%%; <24h: %.0f%%; >72h: %.0f%%' % (
    len(lt), lt[len(lt) // 4], lt[len(lt) // 2], lt[3 * len(lt) // 4], lt[int(len(lt) * .9)],
    100 * sum(1 for x in lt if x < 6) / len(lt), 100 * sum(1 for x in lt if x < 24) / len(lt), 100 * sum(1 for x in lt if x > 72) / len(lt)))

print('\n== 6. lanes ==')
print('items.picked_by non-null rows:', sum(1 for c in cards.values() if c.get('picked_by')), 'of', len(cards))
print('todo_runtime_assignments by owner_label:', dict(collections.Counter(a['owner_label'] for a in assigns)))
# proxy: distinct cards with commits per day (all commits reachable from develop)
R = re.compile(r'(?<![A-Za-z0-9])t(\d{1,4})(?![0-9A-Za-z])')
act = collections.defaultdict(set)
for l in open(OUT + '/allc.tsv'):
    h, d, an, tr, subj = l.rstrip('\n').split('\t', 4)
    m = R.search(subj)
    if m and not subj.lower().startswith('merge'):
        act[ts(d).astimezone(datetime.timezone.utc).date().isoformat()].add('t' + m.group(1))
ds = sorted(act)
v = [len(act[d]) for d in ds]
print('PROXY distinct cards with >=1 non-merge commit per UTC day: days %d median %s mean %.1f P90 %d max %d' % (len(v), statistics.median(v), statistics.mean(v), sorted(v)[int(len(v) * .9)], max(v)))
print('last 14 days:', ' '.join('%s=%d' % (d[5:], len(act[d])) for d in ds[-14:]))
# peak hourly concurrency proxy: cards with commits within the same 3h bucket
hb = collections.defaultdict(set)
for l in open(OUT + '/allc.tsv'):
    h, d, an, tr, subj = l.rstrip('\n').split('\t', 4)
    m = R.search(subj)
    if m and not subj.lower().startswith('merge'):
        t = ts(d).astimezone(datetime.timezone.utc)
        hb[(t.date().isoformat(), t.hour // 3)].add('t' + m.group(1))
hv = sorted(len(x) for x in hb.values())
print('PROXY distinct cards committing within one 3h bucket: median %d P90 %d max %d' % (hv[len(hv) // 2], hv[int(len(hv) * .9)], hv[-1]))
