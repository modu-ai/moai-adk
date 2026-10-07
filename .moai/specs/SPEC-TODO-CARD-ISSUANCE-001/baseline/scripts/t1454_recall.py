import sqlite3, unicodedata, re, math, collections
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text in con.execute("select id,text from items"):
    rows.append((int(cid[1:]), cid, text))
for cid, text in con.execute("select id,text from archived_items"):
    rows.append((int(cid[1:]), cid, text))
rows.sort()
num = {r[1]: r[0] for r in rows}


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')
T = {r[1]: frozenset(norm(DROP.sub('', r[2])).split()) for r in rows}
N = len(T)
df = collections.Counter(x for s in T.values() for x in s)
idf = {x: math.log(N / v) for x, v in df.items()}


def jac(a, b):
    i = len(a & b)
    u = len(a) + len(b) - i
    return i / u if u else 0.0


def wj(a, b):
    den = sum(idf[x] for x in a | b)
    return sum(idf[x] for x in a & b) / den if den else 0.0


pairs = []
for tbl in ('findings', 'archived_findings'):
    for s, r, rel, src in con.execute("select subject_id, related_id, relation, source from %s" % tbl):
        pairs.append((s, r, rel, src))
pairs = list({(s, r, rel, src) for s, r, rel, src in pairs})
print('recorded finding rows (distinct):', len(pairs), collections.Counter((rel, src) for _, _, rel, src in pairs))


def rank(measure, subject, floor):
    me = num[subject]
    sc = []
    for n, cid, _ in rows:
        if cid == subject or n >= me:
            continue
        sc.append((measure(T[subject], T[cid]), cid))
    sc.sort(reverse=True)
    return [c for s, c in sc[:3] if s >= floor]


usable = [(s, r, rel, src) for s, r, rel, src in pairs if s in num and r in num and num[r] < num[s]]
print('pairs where the related card is earlier than the subject (surfaceable at issuance):', len(usable), 'of', len(pairs))
for name, m in (('token-set Jaccard', jac), ('idf-weighted Jaccard', wj)):
    for floor in (0.0, 0.2, 0.3, 0.4, 0.5):
        hit = sum(1 for s, r, rel, src in usable if r in rank(m, s, floor))
        print('%s floor %.1f: recorded counterpart inside top-3: %d of %d (%.0f%%)' % (name, floor, hit, len(usable), 100 * hit / len(usable)))
