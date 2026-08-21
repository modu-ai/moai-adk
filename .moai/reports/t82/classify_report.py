#!/usr/bin/env python3
"""Join blocks.json with classification.tsv; emit the per-block table + subtotals."""
import collections
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
blocks = json.load(open(os.path.join(HERE, "blocks.json"), encoding="utf-8"))
cls = {}
basis = {}
with open(os.path.join(HERE, "classification.tsv"), encoding="utf-8") as fh:
    next(fh)
    for ln in fh:
        if not ln.strip():
            continue
        bid, c, b = ln.rstrip("\n").split("\t")
        cls[bid] = c
        basis[bid] = b

rows = []
for i, x in enumerate(blocks):
    bid = "B%03d" % i
    if bid not in cls:
        sys.exit("unclassified: " + bid)
    rows.append((bid, cls[bid], x["file"], x["line"], x["endline"], x["bytes"], basis[bid]))

tot = collections.Counter()
cnt = collections.Counter()
for _bid, c, _f, _l, _e, b, _r in rows:
    tot[c] += b
    cnt[c] += 1

if "--table" in sys.argv:
    print("| id | class | origin | bytes | basis |")
    print("|---|---|---|---:|---|")
    for bid, c, f, ln, e, b, r in rows:
        print("| %s | %s | `%s`:%d-%d | %d | %s |" % (bid, c, f, ln, e, b, r))
    print()

print("blocks: %d  total bytes: %d" % (len(rows), sum(t for t in tot.values())))
for c in ("C", "K", "P"):
    print("  %s: %3d blocks  %7d B" % (c, cnt[c], tot[c]))
print()
per = collections.defaultdict(lambda: collections.Counter())
for bid, c, f, ln, e, b, r in rows:
    per[f][c] += b
    per[f][c + "n"] += 1
print("| file | C blocks | C bytes | K blocks | K bytes | P bytes |")
print("|---|---:|---:|---:|---:|---:|")
for f in sorted(per, key=lambda k: -(per[k]["C"] + per[k]["K"] + per[k]["P"])):
    p = per[f]
    print("| `%s` | %d | %d | %d | %d | %d |" % (f, p["Cn"], p["C"], p["Kn"], p["K"], p["P"]))
