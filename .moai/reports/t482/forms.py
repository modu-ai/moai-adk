#!/usr/bin/env python3
"""Card t482 — reproduce SPEC-TODO-LANDING-ATTRIBUTION-001 v0.4.0 §A.4's six
positional forms over the pinned corpus, and emit the exhaustive residual.

Input : a subject stream on stdin (one commit subject per line), produced by
        `git log <pinned-corpus> --format=%s`.
Output: the reproduced counts plus the residual id list, and residual.json.

Every regex below is transcribed from the SPEC's own prose; where the prose is
ambiguous the transcription is noted in a comment, because an ambiguity that
the transcription resolves silently is itself a finding.
"""
import collections
import json
import re
import sys

subs = [line.rstrip("\n") for line in sys.stdin]

TOK = re.compile(r"\bt[0-9]+\b")

# --- the six forms, transcribed from §A.4 ---
F1 = re.compile(r"^[a-z]+\((t[0-9]+)\)!?:")                  # conventional-commit scope
F2 = re.compile(r"\((?:card )?(t[0-9]+)\)$")                 # trailing paren, group carries nothing else
F2B = re.compile(r"\(([^()]*t[0-9]+[^()]*)\) \(#[0-9]+\)$")  # trailing paren + reference group
F3A = re.compile(r"^Merge card (t[0-9]+)")
F3C = re.compile(r"^merge: (t[0-9]+)")

# form 3b: a merge whose NAMED TARGET is the branch the resolved landed ref
# names, AND whose trailing parenthetical group carries exactly one card token.
# The landed branch is `develop` for this repository (§A.5).
LANDED_BRANCH = "develop"
MERGE_TARGET = re.compile(r"^[Mm]erge\b.* into ([A-Za-z0-9/_.-]+)")
TRAILGRP = re.compile(r"\(([^()]*)\)$")

attributed = collections.defaultdict(set)
per_form = collections.Counter()

for s in subs:
    hits = []
    m = F1.search(s)
    if m:
        hits.append(("1", m.group(1)))
    m = F2.search(s)
    if m:
        hits.append(("2", m.group(1)))
    m = F2B.search(s)
    if m:
        toks = set(TOK.findall(m.group(1)))
        if len(toks) == 1:
            hits.append(("2b", toks.pop()))
    m = F3A.search(s)
    if m:
        hits.append(("3a", m.group(1)))
    m = F3C.search(s)
    if m:
        hits.append(("3c", m.group(1)))

    mt = MERGE_TARGET.search(s)
    if mt and mt.group(1) == LANDED_BRANCH:
        g = TRAILGRP.search(s)
        if g:
            toks = set(TOK.findall(g.group(1)))
            if len(toks) == 1:
                hits.append(("3b", toks.pop()))

    for form, cid in hits:
        attributed[cid].add(form)
        per_form[form] += 1

mentioned = set()
for s in subs:
    mentioned |= set(TOK.findall(s))

att = set(attributed)
resid = sorted(mentioned - att, key=lambda x: int(x[1:]))

print("subjects                              :", len(subs))
print("ids appearing anywhere in a subject   :", len(mentioned))
print("ids attributed (forms 1/2/2b/3a/3b/3c):", len(att))
print("subject-present but unattributed      :", len(resid))
print("per-form subject counts               :", dict(sorted(per_form.items())))
print()
print("RESIDUAL:", " ".join(resid))

with open(".moai/reports/t482/residual.json", "w", encoding="utf-8") as fh:
    json.dump({"residual": resid, "attributed": sorted(att)}, fh, indent=1)
