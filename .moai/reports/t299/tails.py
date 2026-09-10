import glob
import re
SHA = re.compile(r'^[0-9a-fA-F]{7,40}$')
ALPHA = re.compile(r'^[a-fA-F]{7,40}$')
nonhash = []
alpha = []
tot = 0
for f in sorted(glob.glob('.moai/specs/*/progress.md')):
    for line in open(f, encoding='utf-8', errors='replace'):
        if not line.startswith('sync_commit_sha:'):
            continue
        val = line[len('sync_commit_sha:'):].strip()
        parts = val.split(None, 1)
        if not parts:
            continue
        t = parts[0].strip('"\'')
        tail = parts[1] if len(parts) > 1 else ''
        if SHA.match(t):
            tot += 1
            if tail and '#' not in tail:
                nonhash.append((f, t, tail[:70]))
            if ALPHA.match(t):
                alpha.append((f, t))
print('sha-token values:', tot)
print('with non-# tail:', len(nonhash))
for r in nonhash:
    print('  ', r[0], r[1], '|', r[2])
print('all-alpha hex tokens (L1 candidates):', len(alpha), alpha[:5])
