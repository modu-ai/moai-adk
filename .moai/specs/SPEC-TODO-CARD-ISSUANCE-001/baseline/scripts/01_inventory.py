import json, collections, statistics
from lib import *
cards, findings, meta, assigns = load()
print('last_seq', meta.get('last_seq'), 'rows', len(cards))
ids = sorted(num(c) for c in cards)
missing = sorted(set(range(1, int(meta['last_seq']) + 1)) - set(ids))
print('ids present', len(ids), 'min', ids[0], 'max', ids[-1], 'missing', len(missing))
# missing ranges
rng = []; s = None
for m in missing:
    if s is None: s = p = m
    elif m == p + 1: p = m
    else: rng.append((s, p)); s = p = m
if s is not None: rng.append((s, p))
big = sorted(rng, key=lambda r: r[0] - r[1])[:8]
print('largest missing id ranges', big, 'n ranges', len(rng))
leg = json.load(open(LEGACY))
legids = {i['id'] for i in leg['items']} | {a['item']['id'] for a in leg['archived']}
print('legacy json ids', len(legids), 'not in sqlite', len(legids - set(cards)), 'legacy last_seq', leg['last_seq'])
print('missing covered by legacy', len({'t%d' % m for m in missing} & legids))

print('\n== state ==')
cnt = collections.Counter((c['where'], c['state']) for c in cards.values())
for k, v in sorted(cnt.items()): print(k, v)
arch = [c for c in cards.values() if c['where'] == 'archived']
print('archived with landing_verdict', sum(1 for c in arch if c['landing_verdict']),
      'landing', sum(1 for c in arch if c['landing']), 'archived_at', sum(1 for c in arch if c['archived_at']),
      'picked_at', sum(1 for c in arch if c['picked_at']))
drop_txt = collections.Counter()
for c in cards.values():
    if c['state'] == 'dropped':
        m = re.match(r'\[DROPPED\s*[—-]\s*([^\]]{0,40})\]', c['text'])
        drop_txt[m.group(1).strip()[:25] if m else 'other'] += 1
print('dropped reasons (text prefix, pattern):', drop_txt.most_common(12))

print('\n== issuance per day (added_at, UTC) ==')
day = collections.Counter(c['added_at'][:10] for c in cards.values())
days = sorted(day)
print('first', days[0], 'last', days[-1], 'active days', len(days))
vals = [day[d] for d in days]
print('per active day: median', statistics.median(vals), 'mean %.1f' % statistics.mean(vals), 'max', max(vals))
wk = collections.Counter()
import datetime
for d, v in day.items():
    y, w, _ = datetime.date.fromisoformat(d).isocalendar(); wk['%d-W%02d' % (y, w)] += v
for k in sorted(wk): print(k, wk[k])
print('last 14 days:', [(d[5:], day[d]) for d in days[-14:]])
json.dump({'day': day, 'missing': missing}, open(OUT + '/inventory.json', 'w'))
# id vs added_at monotonic? (ids near 1453 in ~3-4 weeks claim)
bynum = sorted(cards.values(), key=lambda c: num(c['id']))
for n in (1, 100, 300, 500, 700, 900, 1100, 1300, 1450):
    c = min(bynum, key=lambda c: abs(num(c['id']) - n)); print('id', c['id'], 'added', c['added_at'][:10])
