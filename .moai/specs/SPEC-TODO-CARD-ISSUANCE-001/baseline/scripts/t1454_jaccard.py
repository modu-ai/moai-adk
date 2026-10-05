import sqlite3, unicodedata, re, collections, statistics, sys
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text, st in con.execute("select id,text,state from items"):
    rows.append((cid, text, 'live:' + st))
for cid, text, st in con.execute("select id,text,state from archived_items"):
    rows.append((cid, text, 'archived'))


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')


def toks_stripped(text):
    return frozenset(norm(DROP.sub('', text)).split())


print('rows', len(rows))
g = collections.defaultdict(list)
for cid, text, st in rows:
    g[norm(DROP.sub('', text))].append(cid)
dups = [v for v in g.values() if len(v) > 1]
print('identical normalized text groups (drop prefix stripped):', len(dups), 'cards', sum(len(v) for v in dups))
g2 = collections.defaultdict(list)
for cid, text, st in rows:
    g2[norm(text)].append(cid)
dups2 = [v for v in g2.values() if len(v) > 1]
print('identical normalized text groups (as stored):', len(dups2), 'cards', sum(len(v) for v in dups2))
T = [(cid, toks_stripped(text), st) for cid, text, st in rows]
n = len(T)
best = []
top3 = []
for i in range(n):
    sc = []
    a = T[i][1]
    for j in range(n):
        if i == j:
            continue
        b = T[j][1]
        if not a or not b:
            continue
        inter = len(a & b)
        u = len(a) + len(b) - inter
        sc.append(inter / u)
    sc.sort(reverse=True)
    best.append(sc[0] if sc else 0)
    top3.append(sc[:3])


def q(vals, p):
    vals = sorted(vals)
    return vals[int(p * (len(vals) - 1))]


print('best-neighbor Jaccard: median %.3f P75 %.3f P90 %.3f P95 %.3f P99 %.3f max %.3f' % (
    q(best, .5), q(best, .75), q(best, .9), q(best, .95), q(best, .99), max(best)))
for th in (0.8, 0.7, 0.6, 0.5, 0.4, 0.3):
    c = sum(1 for b in best if b >= th)
    print('cards with a neighbor >= %.1f: %d (%.1f%%)' % (th, c, 100 * c / n))
print('cards with neighbor in [0.80,1.0):', sum(1 for b in best if 0.8 <= b < 1.0), ' ==1.0:', sum(1 for b in best if b >= 1.0))
for th in (0.5, 0.4, 0.3, 0.2):
    print('cards whose 3rd best neighbor >= %.1f: %d' % (th, sum(1 for t in top3 if len(t) == 3 and t[2] >= th)))
tc = [len(t[1]) for t in T]
print('tokens per card: median %d P90 %d max %d' % (statistics.median(tc), q(tc, .9), max(tc)))
