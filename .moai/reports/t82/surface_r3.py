#!/usr/bin/env python3
"""Reproduce alwaysLoadedSurface() per-file, and price each file's movable (R3) mass.

R1 = clause-block [HARD] bytes (blocks.json, this file's rows).
R2 = lines carrying MUST / MUST NOT / shall that are not inside an R1 block.
R3 = everything else -- the only material the diet may move (REQ-AMC-002).
"""
import json
import os
import re

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
MEM_LINE_CAP, MEM_BYTE_CAP = 200, 25 * 1024
R2 = re.compile(r'\bMUST NOT\b|\bMUST\b|\bshall\b')


def frontmatter_has_paths(text):
    lines = text.split("\n")
    if not lines or lines[0].rstrip(" \t\r") != "---":
        return False
    for ln in lines[1:]:
        if ln.rstrip(" \t\r") == "---":
            return False
        if ln.startswith("paths:"):
            return True
    return False


def surface():
    rules = []
    for dp, _d, fs_ in os.walk(os.path.join(ROOT, ".claude", "rules", "moai")):
        for f in fs_:
            if not f.endswith(".md"):
                continue
            p = os.path.join(dp, f)
            if not frontmatter_has_paths(open(p, encoding="utf-8").read()):
                rules.append(p)
    return sorted(rules) + [
        os.path.join(ROOT, "CLAUDE.md"),
        os.path.join(ROOT, ".claude", "output-styles", "moai", "moai.md"),
        os.path.join(ROOT, "MEMORY.md"),
    ]


def memory_head(data):
    if len(data) > MEM_BYTE_CAP:
        data = data[:MEM_BYTE_CAP]
    n = 0
    for i, c in enumerate(data):
        if c == 0x0A:
            n += 1
            if n == MEM_LINE_CAP:
                return data[:i + 1]
    return data


blocks = json.load(open(os.path.join(HERE, "blocks.json"), encoding="utf-8"))
r1_lines = {}
r1_bytes = {}
for x in blocks:
    r1_bytes[x["file"]] = r1_bytes.get(x["file"], 0) + x["bytes"]
    r1_lines.setdefault(x["file"], set()).update(range(x["line"], x["endline"] + 1))

total = 0
rows = []
mem = os.path.join(ROOT, "MEMORY.md")
for p in surface():
    try:
        data = open(p, "rb").read()
    except OSError:
        data = b""
    if p == mem:
        data = memory_head(data)
    tok = len(data) // 4
    total += tok
    rel = os.path.relpath(p, ROOT)
    txt = data.decode("utf-8", "replace")
    lines = txt.split("\n")
    r1 = r1_bytes.get(rel, 0)
    covered = r1_lines.get(rel, set())
    r2 = 0
    for i, ln in enumerate(lines, 1):
        if i in covered:
            continue
        if R2.search(ln):
            r2 += len(ln.encode("utf-8")) + 1
    rows.append((rel, len(data), tok, r1, r2, len(data) - r1 - r2))

print("| file | bytes | tokens | R1 | R2 | R3 (movable) |")
print("|---|---:|---:|---:|---:|---:|")
for rel, b, t, r1, r2, r3 in sorted(rows, key=lambda r: -r[1]):
    print("| `%s` | %d | %d | %d | %d | %d |" % (rel, b, t, r1, r2, r3))
print()
print("surface files: %d   guard tokens (sum of per-file len/4): %d" % (len(rows), total))
sb = sum(r[1] for r in rows)
print("surface bytes: %d" % sb)
rules_only = [r for r in rows if r[0].startswith(".claude/rules/") or r[0] == "CLAUDE.md"]
print("rules+CLAUDE.md bytes: %d  R1 %d  R2 %d  R3 %d" % (
    sum(r[1] for r in rules_only), sum(r[3] for r in rules_only),
    sum(r[4] for r in rules_only), sum(r[5] for r in rules_only)))
