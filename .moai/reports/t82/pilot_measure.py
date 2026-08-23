#!/usr/bin/env python3
"""Tally per-clause before/after bytes for the M1 pilot rewrite."""
import json
import os
import re
import statistics

HERE = os.path.dirname(os.path.abspath(__file__))
blocks = json.load(open(os.path.join(HERE, "blocks.json"), encoding="utf-8"))
src = open(os.path.join(HERE, "pilot-compressed.md"), encoding="utf-8").read()

after = {}
for m in re.finditer(r'<!--(B\d{3})-->\n(.*?)\n<!--/\1-->', src, re.S):
    after[m.group(1)] = m.group(2).strip() + "\n"

print("| id | origin | before B | after B | ratio |")
print("|---|---|---:|---:|---:|")
ratios = []
tb = ta = 0
for bid in sorted(after):
    x = blocks[int(bid[1:])]
    b = x["bytes"]
    a = len(after[bid].encode("utf-8"))
    tb += b
    ta += a
    ratios.append(a / b)
    print("| %s | `%s`:%d | %d | %d | %.3f |" % (bid, os.path.basename(x["file"]), x["line"], b, a, a / b))
print("| **aggregate** | | **%d** | **%d** | **%.3f** |" % (tb, ta, ta / tb))
print()
print("clauses: %d" % len(ratios))
print("aggregate compression ratio (after/before): %.4f  -> %.1f%% reduction" % (ta / tb, 100 * (1 - ta / tb)))
print("per-clause ratio: min %.3f  max %.3f  mean %.3f  median %.3f  stdev %.3f" % (
    min(ratios), max(ratios), statistics.mean(ratios), statistics.median(ratios),
    statistics.stdev(ratios)))
