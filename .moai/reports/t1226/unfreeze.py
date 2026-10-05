#!/usr/bin/env python3
"""Greedy unfreeze set for the escalation section (plan.md M6, card t1226).

Usage: unfreeze.py <candidates-<s>.tsv> <T_min>

For every REJECT bind>0 partition row, releasing ALL of its keyword lines
would newly admit the row's gross minus the chars already admitted from its
own non-binding paragraph rows (a row bound by several lines counts only when
every one of them is released — the overlap rule). Rows are taken in
descending order of newly-admitted chars per released line until
T_min - cumulative < 150000. The set is greedy, not minimal.
"""
import re
import sys

path, tmin = sys.argv[1], int(sys.argv[2])
rows = [l.rstrip("\n").split("\t") for l in open(path, encoding="utf-8")][1:]
para_admit = {}
for r in rows:
    m = re.match(r"^(.*) ¶[0-9]+$", r[1])
    if m and r[12] == "ADMIT":
        para_admit[(r[0], m.group(1))] = para_admit.get((r[0], m.group(1)), 0) + int(r[3])
cands = []
for r in rows:
    if r[13] == "bind>0":
        gain = int(r[2]) - para_admit.get((r[0], r[1]), 0)
        cands.append((gain / int(r[5]), gain, int(r[5]), r[0], r[1]))
cands.sort(key=lambda c: (-c[0], c[3], c[4]))
need = tmin - 149999
cum = lines = 0
print("need = %d  (T_min %d - 149999)" % (need, tmin))
print("rank\tlines\tgain\tcum_gain\tcum_lines\tfile\tsection")
for i, (per, gain, k, f, s) in enumerate(cands, 1):
    cum += gain
    lines += k
    print("%d\t%d\t%d\t%d\t%d\t%s\t%s" % (i, k, gain, cum, lines, f, s))
    if cum >= need:
        print("greedy_set_rows = %d\ngreedy_set_lines = %d\ngreedy_set_gain = %d\nT_min_after = %d" % (i, lines, cum, tmin - cum))
        break
else:
    print("greedy set does not reach 150000: total bind>0 gain = %d over %d lines" % (cum, lines))
print("all_bind_rows = %d\nall_bind_lines = %d\nall_bind_gain = %d" % (
    len(cands), sum(c[2] for c in cands), sum(c[1] for c in cands)))
