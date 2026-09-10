#!/usr/bin/env python3
"""Card t482 — dump every subject that mentions each residual id, so the 38 can
be classified EXHAUSTIVELY rather than by whichever shapes a reader notices.

Input : subject stream on stdin, residual.json on disk.
Output: one block per residual id, every mentioning subject printed in full.
"""
import json
import re
import sys

subs = [line.rstrip("\n") for line in sys.stdin]
resid = json.load(open(".moai/reports/t482/residual.json", encoding="utf-8"))["residual"]

for cid in resid:
    pat = re.compile(r"\b" + cid + r"\b")
    hits = [s for s in subs if pat.search(s)]
    print("=" * 4, cid, f"({len(hits)} subject(s))")
    for s in hits:
        print("   ", s)
    print()
