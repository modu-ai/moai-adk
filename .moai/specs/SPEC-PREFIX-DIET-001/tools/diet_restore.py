#!/usr/bin/env python3
"""Sync-audit repair (card t1450): restore dropped units whose survivor cannot be found, attach resolvable
survivor pointers (file + anchor substring) to the rest, and rebuild the three template style files from the ledger.
Run from the worktree root: python3 .moai/specs/SPEC-PREFIX-DIET-001/tools/diet_restore.py
"""
import json
import os
import sys

sys.path.insert(0, os.path.dirname(__file__))
import diet_ledger as dl  # noqa: E402

STY = "templates/.claude/output-styles/moai/"
RULES = "templates/.claude/rules/moai/core/moai-constitution.md"
M, E = STY + "moai.md", STY + "moai-easy.md"


def rng(a, b):
    return list(range(a, b + 1))


# Units restored verbatim: no survivor carries their user-facing instruction.
RESTORE = {
    "moai.md": rng(6, 7),
    "moai-easy.md": [44, 45, 158, 159, 167, 169] + rng(170, 174) + rng(182, 184),
    "moai-learn.md": rng(5, 7) + [122] + rng(137, 141),
}

# Survivor pointers for rows that stay dropped: index ranges -> (file, anchor).
POINTERS = {
    "moai.md": [
        ([64], M, "keep working right through auto-compaction"),
        ([65], RULES, "moai memory doctor"),
        ([246, 247], M, "### Operating Principles"),
        ([248], M, "pair programming partner"),
        ([249], M, "Intent-First"),
        ([250], M, "Verify Every Step"),
    ],
    "moai-easy.md": [
        ([72, 73], E, "Pause, explain X with an analogy, then continue"),
        (rng(53, 71), E, "A **function** (a reusable recipe"),
        (rng(134, 142), E, "### Banner 1 — Let's Begin (Step 1: Understand)"),
        ([160], E, "is plenty"),
        ([161], E, "one-line note on what just changed"),
        ([162], E, "never race past the confusion"),
        ([163], E, "change this"),
        ([164], E, "/output-style MoAI"),
        ([165], E, "| **MoAI-Learn** | Learning a concept deeply"),
        ([166], E, '"I don\'t understand" is always a perfectly good thing to say'),
        ([168], E, "Bring back the result"),
        (rng(175, 180), E, "### How I'm different from my siblings"),
        ([181], E, "just type `/output-style MoAI` right here in the chat"),
    ],
    "moai-learn.md": [],
}

doc = dl.load_ledger()
frozen_doc = json.load(open(dl.FROZEN, encoding="utf-8"))
for r in doc["rows"]:
    f, i = r["file"], r["index"]
    if i in RESTORE[f] and r["treatment"] == "dropped":
        r.update(treatment="verbatim", after_text=r["before_text"], survivor="", note="",
                 survivor_file="", survivor_anchor="")
        print("restored", r["id"], dl.first_line(r["before_text"])[:50])
    r.setdefault("survivor_file", "")
    r.setdefault("survivor_anchor", "")
for r in doc["rows"]:
    if r["treatment"] != "dropped":
        continue
    for idxs, fil, anchor in POINTERS[r["file"]]:
        if r["index"] in idxs:
            r["survivor_file"], r["survivor_anchor"] = fil, anchor
    assert r["survivor_file"], ("no pointer", r["id"])
dl.write_ledger(doc["files"], doc["rows"])
for f in dl.FILES:
    full = dl.git_show(dl.STYLE_DIR + f)
    head, _ = dl.split_frontmatter(full)
    new = head + dl.reconstruct(doc, frozen_doc, f, after=True)
    open(dl.STYLE_DIR + f, "w", encoding="utf-8").write(new)
    print(f, dl.u16(full), "->", dl.u16(new))
print("dropped rows", sum(1 for r in doc["rows"] if r["treatment"] == "dropped"))
