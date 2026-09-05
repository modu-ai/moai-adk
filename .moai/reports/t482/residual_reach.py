#!/usr/bin/env python3
"""Card t482 — for every residual id, count where in the corpus it is reachable
at all: subject occurrences and full-message occurrences.

An id whose ONLY corpus evidence is a subject the discriminator excludes has no
second chance: the landed verdict for it is `not-landed` on the whole corpus,
not merely on that one commit. That is the fact that separates "correctly
excluded, evidence lies elsewhere" from "silently wrong".

Input : /tmp/t482-full.txt   (git log --format='%h\\x1f%s\\x1f%B\\x1e')
        .moai/reports/t482/residual.json
"""
import json
import re

RECS = []
raw = open("/tmp/t482-full.txt", encoding="utf-8", errors="replace").read()
for rec in raw.split("\x1e"):
    rec = rec.strip("\n")
    if not rec:
        continue
    parts = rec.split("\x1f")
    if len(parts) < 3:
        continue
    RECS.append((parts[0], parts[1], parts[2]))

resid = json.load(open(".moai/reports/t482/residual.json", encoding="utf-8"))["residual"]

print(f"{'id':>6}  {'subj':>4}  {'msg':>4}   verdict")
for cid in resid:
    pat = re.compile(r"\b" + cid + r"\b")
    subj = sum(1 for _, s, _ in RECS if pat.search(s))
    msg = sum(1 for _, _, b in RECS if pat.search(b))
    verdict = "ONLY-IN-EXCLUDED-SUBJECTS" if msg == subj else "also in bodies"
    print(f"{cid:>6}  {subj:>4}  {msg:>4}   {verdict}")
