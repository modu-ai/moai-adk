import json, collections, datetime
from lib import *
cards, findings, meta, assigns = load()
D = json.load(open(OUT + '/delivered.json')); G = json.load(open(OUT + '/graph.json'))
def ts(s): return datetime.datetime.fromisoformat(s.replace('Z', '+00:00'))
par = G['parents']
n = 0; res = collections.Counter(); small = 0; gaps = []
for c, ps in par.items():
    if not ps or c not in D: continue
    best = None
    for p in ps:
        if p not in D: continue
        fc, fp = set(D[c]['files']), set(D[p]['files'])
        gap = (ts(cards[c]['added_at']) - ts(D[p]['t'])).total_seconds() / 3600   # child issued minus parent landed
        ov = len(fc & fp); sub = bool(fc) and fc <= fp
        cand = (sub, ov, -abs(gap), gap, len(fc))
        if best is None or cand > best: best = cand
    if best is None: continue
    n += 1; sub, ov, _, gap, nf = best; gaps.append(gap)
    res['child issued BEFORE parent landed'] += gap < 0
    res['child issued within 24h after parent landed'] += 0 <= gap <= 24
    res['child issued within 48h of parent landing (either side)'] += abs(gap) <= 48
    res['child shares >=1 product file with parent'] += ov > 0
    res['child product files are a SUBSET of parent files'] += sub
    res['subset AND |gap|<=48h'] += sub and abs(gap) <= 48
    res['shares file AND |gap|<=48h'] += ov > 0 and abs(gap) <= 48
    res['shares file AND |gap|<=48h AND child <150 product lines'] += ov > 0 and abs(gap) <= 48 and D[c]['lines'] < 150
print('derived cards where both child and a parent have delivery commits: n=%d (of %d delivered cards in store)' % (n, sum(1 for c in D if c in cards)))
for k, v in res.items(): print('  %-60s %4d  %.0f%%' % (k, v, 100 * v / n))
gaps.sort(); print('  gap hours (child issue - parent landing): Q1 %.1f med %.1f Q3 %.1f' % (gaps[len(gaps) // 4], gaps[len(gaps) // 2], gaps[3 * len(gaps) // 4]))
# size by depth
dep = G['depth']
for k in (0, 1, 2, 3):
    L = sorted(D[c]['lines'] for c in D if c in dep and (dep[c] == k if k < 3 else dep[c] >= 3))
    F = sorted(len(D[c]['files']) for c in D if c in dep and (dep[c] == k if k < 3 else dep[c] >= 3))
    print('depth %s%d: n=%d product lines Q1/med/Q3 = %d/%d/%d ; files med %d ; <50 lines %.0f%%' % ('>=' if k == 3 else '', k, len(L), L[len(L) // 4], L[len(L) // 2], L[3 * len(L) // 4], F[len(F) // 2], 100 * sum(1 for x in L if x < 50) / len(L)))
# identical-text groups
nt = collections.defaultdict(list)
for c in cards.values():
    t = re.sub(r'^\[DROPPED[^\]]*\]\s*', '', c['text']); t = re.sub(r'\s+', ' ', t).strip().lower()
    nt[t].append(c['id'])
ex = [sorted(v, key=num) for v in nt.values() if len(v) > 1]
blk = [g for g in ex if 700 <= num(g[-1]) <= 800 and num(g[0]) < 700]
print('\nidentical-text groups %d; second copy in t700-t800 block (bulk re-issue): %d; other: %s' % (len(ex), len(blk), [g for g in ex if g not in blk][:12]))
st = collections.Counter((fate(cards[g[0]]), fate(cards[g[-1]])) for g in ex); print(' fate (first copy, last copy):', dict(st))
print(' added_at of second copies:', collections.Counter(cards[g[-1]]['added_at'][:10] for g in ex).most_common(4))
# jev findings coverage over time
jd = collections.Counter(f['at'][:10] for f in findings if f['source'] == 'jev'); print('\njev finding days:', sorted(jd.items()))
ad = collections.Counter(f['at'][:10] for f in findings if f['source'] == 'agent'); print('agent finding days:', sorted(ad.items()))
first = min(f['at'] for f in findings if f['source'] == 'jev')
since = [c for c in cards.values() if c['added_at'] >= first]
named = {x for f in findings for x in (f['subject_id'], f['related_id'])}
print('cards issued since first jev finding (%s): %d; of which named in any finding: %d (%.0f%%)' % (first[:10], len(since), sum(1 for c in since if c['id'] in named), 100 * sum(1 for c in since if c['id'] in named) / len(since)))
subj = collections.Counter(f['subject_id'] for f in findings)
print('findings where subject is the NEWER card: %d of %d' % (sum(1 for f in findings if num(f['subject_id']) > num(f['related_id'])), len(findings)))
lag = sorted((ts(f['at']) - ts(cards[f['subject_id']]['added_at'])).total_seconds() / 60 for f in findings if f['source'] == 'jev' and f['subject_id'] in cards)
print('jev finding lag after subject card issue (minutes): med %.1f Q3 %.1f max %.0f; within 1 min: %d of %d' % (lag[len(lag) // 2], lag[3 * len(lag) // 4], lag[-1], sum(1 for x in lag if x <= 1), len(lag)))
# high-score near-dup: fate of the newer card
hi = [f for f in findings if f['relation'] == 'near-duplicate' and f['score'] >= 0.8]
fc = collections.Counter(fate(cards[f['subject_id']]) if f['subject_id'] in cards else 'none' for f in hi)
print('near-dup >=0.8: fate of subject card:', dict(fc))
ov = json.load(open(OUT + '/overlap.json'))
print('\ngists for top overlap pairs:')
seen = set()
for w, j, a, b, sh in ov['top'][:16]:
    for x in (a, b):
        if x not in seen: seen.add(x); print('  %s: %s' % (x, ov['gist'][x]))
