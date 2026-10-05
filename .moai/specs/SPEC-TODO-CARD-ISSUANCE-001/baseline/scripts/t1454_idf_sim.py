import sqlite3, unicodedata, re, math, collections
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text, st in con.execute("select id,text,state from items"):
    rows.append((int(cid[1:]), cid, text))
for cid, text, st in con.execute("select id,text,state from archived_items"):
    rows.append((int(cid[1:]), cid, text))
rows.sort()


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')
T = {r[1]: frozenset(norm(DROP.sub('', r[2])).split()) for r in rows}
N = len(T)
df = collections.Counter(x for s in T.values() for x in s)
idf = {x: math.log(N / v) for x, v in df.items()}
normed = {r[1]: norm(DROP.sub('', r[2])) for r in rows}


def wj(a, b):
    inter = a & b
    u = a | b
    den = sum(idf[x] for x in u)
    return sum(idf[x] for x in inter) / den if den else 0.0


def q(vals, p):
    vals = sorted(vals)
    return vals[int(p * (len(vals) - 1))]


for lo in (1, 1300):
    top1 = []
    for n, cid, text in rows:
        if n < lo:
            continue
        best = 0.0
        for pn, pid, _ in rows:
            if pn >= n or normed[pid] == normed[cid]:
                continue
            s = wj(T[cid], T[pid])
            if s > best:
                best = s
        top1.append(best)
    m = len(top1)
    print('idf-weighted, window ids >= t%d: cards %d: top-1 median %.3f P75 %.3f P90 %.3f P95 %.3f P99 %.3f max %.3f' % (
        lo, m, q(top1, .5), q(top1, .75), q(top1, .9), q(top1, .95), q(top1, .99), max(top1)))
    for th in (0.5, 0.4, 0.3, 0.2, 0.15, 0.1):
        c = sum(1 for v in top1 if v >= th)
        print('  floor %.2f: cards with >=1 neighbor %d (%.1f%%)' % (th, c, 100 * c / m))
