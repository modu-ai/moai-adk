import glob
import re
SHA = re.compile(r'^[0-9a-fA-F]{7,40}$')
tot = sha_hash = all_hash = 0
for f in sorted(glob.glob('.moai/specs/*/progress.md')):
    for line in open(f, encoding='utf-8', errors='replace'):
        if not line.startswith('sync_commit_sha:'):
            continue
        val = line[len('sync_commit_sha:'):].strip()
        tot += 1
        if '#' in val:
            all_hash += 1
        parts = val.split(None, 1)
        t = parts[0].strip('"\'') if parts else ''
        if SHA.match(t) and '#' in val:
            sha_hash += 1
print('total', tot, '| any-value with #:', all_hash, '| SHA-token values with #:', sha_hash)
