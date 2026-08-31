import glob
import re
SHA = re.compile(r'^[0-9a-fA-F]{7,40}$')
PH = re.compile(r'^pending-backfill(-[A-Za-z0-9-]+)?$')


def token(value):
    parts = value.split()
    if not parts:
        return ''
    t = parts[0]
    if t[:1] in ('"', "'"):
        t = t[1:]
    if t[-1:] in ('"', "'"):
        t = t[:-1]
    return t


sha = []
ph = []
bad = []
for f in sorted(glob.glob('.moai/specs/*/progress.md')):
    for i, line in enumerate(open(f, encoding='utf-8', errors='replace'), 1):
        if not line.startswith('sync_commit_sha:'):
            continue
        val = line[len('sync_commit_sha:'):].strip()
        t = token(val)
        rec = (f, i, val[:80], t)
        (sha if SHA.match(t) else ph if PH.match(t) else bad).append(rec)

print("total", len(sha) + len(ph) + len(bad), "| SHA", len(sha),
      "| PLACEHOLDER", len(ph), "| FLAGGED", len(bad))
print("--- FLAGGED ---")
for r in bad:
    print(r[0], r[1], "token=", repr(r[3]), "val=", r[2])
print("--- PLACEHOLDER ---")
for r in ph:
    print(r[0], "token=", repr(r[3]))
