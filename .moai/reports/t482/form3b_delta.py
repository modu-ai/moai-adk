#!/usr/bin/env python3
"""Card t482 — locate the one subject separating this transcription of form 3b
(77) from the count SPEC-TODO-LANDING-ATTRIBUTION-001 v0.4.0 §A.4 records (76).

A transcription that reproduces 347/309/38 exactly but disagrees on a per-form
count means the residual is right and one form's WORDING is loose. Which
subject moves is the evidence for which clause is loose.
"""
import re
import sys

subs = [line.rstrip("\n") for line in sys.stdin]
TOK = re.compile(r"\bt[0-9]+\b")
TRAILGRP = re.compile(r"\(([^()]*)\)$")

# variant A — this card's transcription: `^[Mm]erge ... into develop`
A = re.compile(r"^[Mm]erge\b.* into develop\b")
# variant B — capital-M only, the spelling §A.4's examples all use
B = re.compile(r"^Merge\b.* into develop\b")

hits = {"A": [], "B": []}
for s in subs:
    g = TRAILGRP.search(s)
    if not g:
        continue
    toks = set(TOK.findall(g.group(1)))
    if len(toks) != 1:
        continue
    if A.search(s):
        hits["A"].append(s)
    if B.search(s):
        hits["B"].append(s)

print("variant A (^[Mm]erge ... into develop):", len(hits["A"]))
print("variant B (^Merge    ... into develop):", len(hits["B"]))
print()
print("subjects in A but not in B:")
for s in hits["A"]:
    if s not in hits["B"]:
        print("   ", s)
