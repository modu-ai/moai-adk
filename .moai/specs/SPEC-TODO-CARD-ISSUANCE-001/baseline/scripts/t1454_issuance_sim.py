import sqlite3, unicodedata, re, statistics
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text, st, seq in con.execute("select id,text,state,seq from items"):
    rows.append((int(cid[1:]), cid, text, 'live:' + st))
for cid, text, st, seq in con.execute("select id,text,state,seq from archived_items"):
    rows.append((int(cid[1:]), cid, text, 'archived'))
rows.sort()


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')
toks = {r[1]: frozenset(norm(DROP.sub('', r[2])).split()) for r in rows}
normed = {r[1]: norm(DROP.sub('', r[2])) for r in rows}


def jac(a, b):
    if not a or not b:
        return 0.0
    i = len(a & b)
    return i / (len(a) + len(b) - i)


def q(vals, p):
    vals = sorted(vals)
    return vals[int(p * (len(vals) - 1))]


def sim(window_lo):
    res = []
    exact = 0
    for n, cid, text, st in rows:
        if n < window_lo:
            continue
        prior = [r for r in rows if r[0] < n]
        sc = []
        ex = False
        for _, pid, _, pst in prior:
            if normed[pid] == normed[cid]:
                ex = True
                continue
            sc.append((jac(toks[cid], toks[pid]), pid))
        sc.sort(reverse=True)
        res.append([s for s, _ in sc[:3]])
        exact += 1 if ex else 0
    return res, exact


for lo in (1, 1300):
    res, exact = sim(lo)
    n = len(res)
    print('window ids >= t%d: cards %d, with an exact normalized duplicate among EARLIER cards (live+archived+dropped): %d' % (lo, n, exact))
    top1 = [r[0] for r in res if r]
    print('  top-1 non-exact Jaccard: median %.3f P75 %.3f P90 %.3f P95 %.3f max %.3f' % (q(top1, .5), q(top1, .75), q(top1, .9), q(top1, .95), max(top1)))
    for th in (0.8, 0.6, 0.5, 0.4, 0.3, 0.2):
        c1 = sum(1 for r in res if r and r[0] >= th)
        c3 = sum(1 for r in res if len(r) == 3 and r[2] >= th)
        print('  floor %.1f: cards with >=1 neighbor %d (%.1f%%), with all 3 slots filled %d (%.1f%%)' % (th, c1, 100 * c1 / n, c3, 100 * c3 / n))
