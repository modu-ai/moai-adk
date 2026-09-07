#!/usr/bin/env python3
"""Card t482 — measure the live-queue status of the 38 residual ids, which
SPEC-TODO-LANDING-ATTRIBUTION-001 v0.4.0 §A.4.2 records as UNMEASURED
("their live-queue status is likewise unmeasured ... so no operational-impact
figure is derivable from this number either").

Reads a COPY of the SQLite queue store, per the standing lesson that `moai
todo` and the disk store are known to disagree; the store is the record the
CLI reads, so it is the one cited here. The copy is taken so a concurrent lane
writing the live store cannot make this read fail or tear.

Input : /tmp/t482-backlog.db  (cp of .moai/state/todo/backlog.db)
        .moai/reports/t482/residual.json
"""
import collections
import json
import sqlite3

con = sqlite3.connect("/tmp/t482-backlog.db")
resid = json.load(open(".moai/reports/t482/residual.json", encoding="utf-8"))["residual"]

live = {r[0]: r[1] for r in con.execute("SELECT id, state FROM items")}
arch = {r[0] for r in con.execute("SELECT id FROM archived_items")}

print("live queue rows   :", len(live))
print("archived rows     :", len(arch))
print()

counts = collections.Counter()
for cid in resid:
    if cid in live:
        st = "live:" + str(live[cid])
    elif cid in arch:
        st = "archived"
    else:
        st = "absent"
    counts[st] += 1
    print(f"{cid:>6}  {st}")

print()
print("summary:", dict(counts))
