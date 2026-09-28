"""Render the file x token x class table as markdown from raw TSVs."""
import sys, os, collections
RAW = sys.argv[1]
rows = [l.split('\t') for l in open(os.path.join(RAW, 'per-file-class.tsv')).read().split('\n')[1:] if l]
hard = collections.Counter()
for l in open(os.path.join(RAW, 'hard-lines.tsv')).read().split('\n')[1:]:
    if l:
        p, ln, toks, _ = l.split('\t', 3)
        for t in toks.split(','):
            hard[(p, t)] += 1
anch = collections.defaultdict(set)
for l in open(os.path.join(RAW, 'anchors.tsv')).read().split('\n')[1:]:
    if l:
        p, ln, toks, h, n, _ = l.split('\t', 5)
        if int(n) > 0:
            for t in toks.split(','):
                anch[(p, t)].add(n)
agg = collections.defaultdict(collections.Counter)
sfm = {}
for p, sf, tok, cls, n in rows:
    agg[(p, tok)][cls] += int(n)
    sfm[p] = sf
ORDER = ['T:.claude/rules', 'T:.claude/agents', 'T:.claude/skills', 'T:.claude/output-styles', 'T:.claude/commands', 'T:.claude/loop.md',
         'T:.codex', 'T:.agents', 'T:.moai', 'T:root', 'L:.claude/rules', 'L:.claude/agents', 'L:.claude/skills',
         'L:.claude/output-styles', 'L:.claude/commands', 'L:root', 'L:.moai/config', 'docs:en', 'docs:ko', 'docs:ja', 'docs:zh', 'README']
TOKO = ['lead', 'leader', 'lane', 'worker', 'companion', 'foreman', 'deputy', 'coordinator']
def key(k):
    p, t = k
    sf = sfm[p]
    return (ORDER.index(sf) if sf in ORDER else 99, p, TOKO.index(t))
print('| # | file | surface | token | total | role | ident | compound | other-meaning | plain-en | frozen | HARD | anchor |')
print('|---|---|---|---|---|---|---|---|---|---|---|---|---|')
for i, k in enumerate(sorted(agg, key=key), 1):
    c = agg[k]
    p, t = k
    role = c['role'] + c['role:subagent-worker']
    role_s = str(role) + (' (sub %d)' % c['role:subagent-worker'] if c['role:subagent-worker'] else '')
    ident = sum(v for kk, v in c.items() if kk.startswith('ident:'))
    ident_s = str(ident) + (' (' + ','.join('%s %d' % (kk[6:], v) for kk, v in sorted(c.items()) if kk.startswith('ident:')) + ')' if ident else '')
    other = sum(v for kk, v in c.items() if kk.startswith('other-meaning'))
    other_s = str(other) + (' (' + ','.join('%s %d' % (kk.split(':')[1], v) for kk, v in sorted(c.items()) if kk.startswith('other-meaning')) + ')' if other else '')
    print('| %d | `%s` | %s | %s | %d | %s | %s | %d | %s | %d | %d | %s | %s |' % (
        i, p, sfm[p], t, sum(c.values()), role_s, ident_s, c['compound(slug-or-prose)'], other_s,
        c['plain-english'], c['frozen/historical'], hard[k] or '', 'Y' if anch[k] else ''))
