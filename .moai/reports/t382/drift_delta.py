"""t382 AC-EH3-007 proposition-2 probe — per-row drift delta on the ex-V3R5 bucket.

Usage: drift_delta.py <drift-before.txt> <drift-after.txt> <population.txt>

For every SPEC that left the V3R5 bucket, prints its drift row before and after, its
frontmatter status, and whether the row is now flagged DRIFT. Exit 1 if any row is
missing from either drift output (a set mismatch that must be investigated on its own).
"""
import re
import sys


def rows(path):
    out = {}
    for line in open(path, encoding="utf-8"):
        parts = line.split()
        if len(parts) >= 3 and parts[0].startswith("SPEC-"):
            out[parts[0]] = line.rstrip("\n")
    return out


before = rows(sys.argv[1])
after = rows(sys.argv[2])
ids = sorted(set(re.findall(r"^(SPEC-[A-Z0-9-]+)\s", open(sys.argv[3], encoding="utf-8").read(), re.M)))

missing = [i for i in ids if i not in before or i not in after]
drifted = []
print("population: %d" % len(ids))
for i in ids:
    b = before.get(i, "<absent>")
    a = after.get(i, "<absent>")
    flag = "DRIFT" if re.search(r"\bDRIFT\b", a) else "-"
    if flag == "DRIFT":
        drifted.append(i)
    print("  %s\n    before: %s\n    after : %s\n    flag  : %s" % (i, b, a, flag))

print("\nrows now flagged DRIFT: %d %s" % (len(drifted), drifted))
if missing:
    print("MISSING from a drift output: %s" % missing)
sys.exit(1 if missing else 0)
