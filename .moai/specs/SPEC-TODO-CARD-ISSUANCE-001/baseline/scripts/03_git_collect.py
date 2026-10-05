"""Collect numstat for every first-parent commit on develop since 2026-08-13 (read-only git)."""
import subprocess, re, json, collections
from lib import OUT
G = ['git', '-C', '/Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop']
R = re.compile(r'(?<![A-Za-z0-9])t(\d{1,4})(?![0-9A-Za-z])')
rows = [l.rstrip('\n').split('\t') for l in open(OUT + '/fp.tsv')]
out = []
for h, parents, date, subj in rows:
    ps = parents.split()
    if not ps: continue
    ids = ['t' + x for x in R.findall(subj)]
    src = 'subject'
    if not ids and len(ps) > 1:
        body = subprocess.run(G + ['log', '--format=%s%n%b', '%s..%s' % (ps[0], ps[1])], capture_output=True, text=True).stdout
        c = collections.Counter('t' + x for x in R.findall(body))
        if c: ids = [c.most_common(1)[0][0]]; src = 'branch-commits'
    ns = subprocess.run(G + ['diff', '--numstat', '--no-renames', ps[0], h], capture_output=True, text=True).stdout
    files = []
    for l in ns.splitlines():
        a, d, p = l.split('\t', 2)
        files.append((p, 0 if a == '-' else int(a), 0 if d == '-' else int(d)))
    ncommits = 1
    if len(ps) > 1:
        ncommits = int(subprocess.run(G + ['rev-list', '--count', '--no-merges', '%s..%s' % (ps[0], ps[1])], capture_output=True, text=True).stdout or 0)
    out.append({'h': h[:10], 'merge': len(ps) > 1, 'date': date, 'subj': subj, 'ids': ids, 'src': src, 'files': files, 'ncommits': ncommits})
json.dump(out, open(OUT + '/commits.json', 'w'))
print('commits', len(out), 'with card', sum(1 for o in out if o['ids']), 'by branch-commit fallback', sum(1 for o in out if o['src'] == 'branch-commits'))
