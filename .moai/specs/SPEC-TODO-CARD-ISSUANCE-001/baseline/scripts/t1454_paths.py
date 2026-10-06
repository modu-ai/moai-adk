import sqlite3, re, collections, itertools
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
live = [(r[0], r[1], r[2]) for r in con.execute("select id,text,state from items")]
arch = [(r[0], r[1], 'archived') for r in con.execute("select id,text from archived_items")]
# the path regex the existing triage helper uses (todo_triage.go todoTriagePathPattern)
PATH = re.compile(r'\b([a-z_]+/[a-z0-9_./-]+\.(?:go|md|yaml|sh))\b')
# a wider form: any path-like token with an extension, to bound the narrow form's recall
WIDE = re.compile(r'(?<![A-Za-z0-9_./-])((?:[.A-Za-z0-9_-]+/)+[A-Za-z0-9_.-]+\.(?:go|md|yaml|yml|sh|json|js|toml|templ|py|tmpl))')


def comp(p, depth=2):
    return '/'.join(p.split('/')[:depth])


def stats(name, cards):
    n = len(cards)
    narrow = sum(1 for _, t, _ in cards if PATH.search(t))
    wide = sum(1 for _, t, _ in cards if WIDE.search(t))
    npaths = [len(set(WIDE.findall(t))) for _, t, _ in cards if WIDE.search(t)]
    npaths.sort()
    med = npaths[len(npaths) // 2] if npaths else 0
    print('%s: %d cards; narrow(todoTriagePathPattern) names >=1 path: %d (%.0f%%); wide path form: %d (%.0f%%); median distinct paths among those: %d; max %d' % (
        name, n, narrow, 100 * narrow / n, wide, 100 * wide / n, med, max(npaths) if npaths else 0))


stats('all cards', live + arch)
stats('live non-dropped', [c for c in live if c[2] != 'dropped'])
stats('live open (queued+picked+hold)', [c for c in live if c[2] in ('queued', 'picked', 'hold')])
open_ = [c for c in live if c[2] in ('queued', 'picked', 'hold')]
for depth in (2, 3):
    keys = {cid: {comp(p, depth) for p in set(WIDE.findall(t))} for cid, t, _ in open_}
    pairs = [(a, b) for a, b in itertools.combinations(keys, 2) if keys[a] & keys[b]]
    withkey = sum(1 for k in keys.values() if k)
    print('open cards depth-%d component keys: cards with >=1 key %d of %d; pairs sharing >=1 key %d of %d' % (depth, withkey, len(keys), len(pairs), len(keys) * (len(keys) - 1) // 2))
fk = {cid: set(WIDE.findall(t)) for cid, t, _ in open_}
fpairs = [(a, b, sorted(fk[a] & fk[b])) for a, b in itertools.combinations(fk, 2) if fk[a] & fk[b]]
print('open cards pairs sharing >=1 exact path:', len(fpairs))
for a, b, s in fpairs[:8]:
    print('  ', a, b, s[:3])
cnt = collections.Counter()
for _, t, _ in live + arch:
    for p in set(WIDE.findall(t)):
        cnt[p] += 1
print('most-named paths across all cards:', cnt.most_common(8))
