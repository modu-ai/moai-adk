#!/usr/bin/env python3
"""Summarize clause blocks: per-file counts and bytes, clause-initial split."""
import collections
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
b = json.load(open(os.path.join(HERE, "blocks.json"), encoding="utf-8"))
print("blocks: %d  bytes: %d" % (len(b), sum(x["bytes"] for x in b)))
ci = [x for x in b if x["clause_initial"]]
pm = [x for x in b if not x["clause_initial"]]
print("clause-initial: %d  bytes: %d" % (len(ci), sum(x["bytes"] for x in ci)))
print("prose-mention : %d  bytes: %d" % (len(pm), sum(x["bytes"] for x in pm)))
print()
cnt = collections.Counter()
byt = collections.Counter()
for x in b:
    cnt[x["file"]] += 1
    byt[x["file"]] += x["bytes"]
for f, nb in byt.most_common():
    print("%3d %7d  %s" % (cnt[f], nb, f))
