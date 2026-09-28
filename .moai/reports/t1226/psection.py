#!/usr/bin/env python3
"""P절 reconciliation arithmetic for card t1226 (AC-ALH-005).

Reads sec-172ef22eb.txt (the committed sec.py rerun on the original tree) and
the design.md §4.0 pool column / §4.3 rejection table of the predecessor SPEC,
and prints every figure the verdict's P절 section cites.
"""
import os
import re

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.normpath(os.path.join(HERE, "..", "..", ".."))
DESIGN = os.path.join(ROOT, ".moai/specs/SPEC-ALWAYS-LOADED-DIET-002/design.md")

pool, cur = {}, None
for line in open(os.path.join(HERE, "sec-172ef22eb.txt"), encoding="utf-8"):
    m = re.match(r"^=== (\S+)", line)
    if m:
        cur = m.group(1)
        continue
    m = re.match(r"^\s+(\d+)\s+b=(\d+)\s+L(\d)\s+(.*)$", line)
    if m and m.group(2) == "0" and m.group(3) != "1":
        pool[cur] = pool.get(cur, 0) + int(m.group(1))
        if cur.endswith("kanban-dispatch.md") and m.group(4).startswith("Scope"):
            print("rerun kanban '## Scope — when this rule is live' = %s (b=0, L2)" % m.group(1))
print("rerun per-file bind-0 section pool (L1 excluded):")
for f in sorted(pool):
    print("  %6d  %s" % (pool[f], f))
P = sum(pool.values())
print("rerun P절 = %d" % P)

text = open(DESIGN, encoding="utf-8").read()
t40 = re.search(r"^\| 파일 \| §3\.2 gross 목표.*?^\| \*\*합계\*\*.*?$", text, re.M | re.S).group(0)
col = []
for row in t40.splitlines()[2:-1]:
    cells = [c.strip().strip("`*").replace(",", "") for c in row.strip("|").split("|")]
    col.append((cells[0], int(cells[2])))
print("design.md §4.0 pool column: rows=%d sum=%d" % (len(col), sum(c[1] for c in col)))
print("design.md §4.0 kanban cell = %d" % dict(col)["kanban-dispatch.md"])
t43 = re.search(r"^### §4\.3.*?^---$", text, re.M | re.S).group(0)
J, n, kan = 0, 0, False
for row in t43.splitlines():
    cells = [c.strip() for c in row.strip("|").split("|")]
    if len(cells) >= 4 and re.fullmatch(r"[0-9,]+", cells[2]):
        J += int(cells[2].replace(",", ""))
        n += 1
        if "kanban-dispatch.md" in cells[0] and "Scope" in cells[1]:
            kan = True
print("design.md §4.3 J = %d over %d rows with a char count; includes kanban Scope row: %s" % (J, n, "yes" if kan else "no"))

B, A_agents, P_M2, R = 246943, 8489, 77206, 8400
y = 3092 / 28717
m1 = P - J - A_agents
base = P_M2 + J + A_agents
m2 = round(base * y)
print("M1 cap = %d - %d - %d = %d" % (P, J, A_agents, m1))
print("M2 base = %d + %d + %d = %d ; M2 cap = round(%d x 3092/28717) = %d" % (P_M2, J, A_agents, base, base, m2))
print("F(172ef22eb) = %d - %d - %d + %d = %d" % (B, m1, m2, R, B - m1 - m2 + R))
