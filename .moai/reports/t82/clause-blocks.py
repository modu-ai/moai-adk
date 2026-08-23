#!/usr/bin/env python3
"""Expand [HARD] markers to clause blocks over the always-loaded surface.

A clause block = the marker line plus its continuation to the next clause or
heading (SPEC-AGENTS-MD-CANON-001 AC-AMC-003).

Continuation rule:
  start at the marker line; consume following lines while
    - non-blank, not a heading, not a horizontal rule, not another [HARD] marker
    - a blank line is consumed only when the next non-blank line is
      structurally subordinate (list item, table row, fence, blockquote,
      indented 2+ spaces) -- i.e. the clause's own body continues
  stop at: heading (^#{1,6} ), horizontal rule, a new [HARD] marker line, or EOF.

Fenced code blocks are consumed whole (a fence opened inside a block runs to its
closing fence regardless of blank lines).

Marker position: a marker is "clause-initial" when [HARD] opens the line,
optionally preceded by list bullets, bold markers, blockquote, or a [ZONE:*]
tag. Otherwise it is a prose mention.

Surface: reproduces internal/config/token_budget_guard.go alwaysLoadedSurface()
rule enumeration -- .claude/rules/**/*.md without `paths:` frontmatter -- plus
CLAUDE.md. The output style is excluded by spec.md 5E.2 (render surface).
"""
import json
import os
import re

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", ".."))
HEAD = re.compile(r'^\s{0,3}#{1,6}\s')
HR = re.compile(r'^\s{0,3}(-{3,}|\*{3,}|_{3,})\s*$')
FENCE = re.compile(r'^\s*(```|~~~)')
SUB = re.compile(r'^(\s{2,}|[-*+]\s|\d+[.)]\s|\||>|\s*(```|~~~))')
INIT = re.compile(r'^\s*(>\s*)?([-*+]\s+|\d+[.)]\s+)?(\*{1,2})?\s*(\[ZONE:[^\]]*\]\s*)?(\*{1,2})?\s*\[HARD\]')


def frontmatter_has_paths(text):
    """Mirror token_budget_guard.go frontmatterHasPaths: top-level `paths:` inside
    a leading `---` frontmatter block only; body-level `paths:` is ignored."""
    lines = text.split("\n")
    if not lines or lines[0].rstrip(" \t\r") != "---":
        return False
    for ln in lines[1:]:
        if ln.rstrip(" \t\r") == "---":
            return False
        if ln.startswith("paths:"):
            return True
    return False


def always_loaded_rules():
    out = []
    for dirpath, _dirs, files in os.walk(os.path.join(ROOT, ".claude", "rules", "moai")):
        for f in files:
            if not f.endswith(".md"):
                continue
            p = os.path.join(dirpath, f)
            with open(p, encoding="utf-8") as fh:
                txt = fh.read()
            if frontmatter_has_paths(txt):
                continue
            out.append(p)
    return sorted(out)


def blocks(path):
    with open(path, encoding="utf-8") as fh:
        lines = fh.read().split("\n")
    res = []
    i = 0
    n = len(lines)
    while i < n:
        if "[HARD]" not in lines[i]:
            i += 1
            continue
        start = i
        buf = [lines[i]]
        j = i + 1
        # Heading-borne marker: the obligation is the section the heading opens,
        # so the block runs to the next heading of the same or higher level.
        hm = HEAD.match(lines[i])
        if hm:
            level = len(lines[i].strip().split(" ")[0])
            while j < n:
                h2 = HEAD.match(lines[j])
                if h2 and len(lines[j].strip().split(" ")[0]) <= level:
                    break
                buf.append(lines[j])
                j += 1
            text = "\n".join(buf).rstrip("\n") + "\n"
            res.append({
                "file": os.path.relpath(path, ROOT),
                "line": start + 1,
                "endline": start + len(buf),
                "bytes": len(text.encode("utf-8")),
                "clause_initial": True,
                "heading_borne": True,
                "text": text,
            })
            i = j
            continue
        while j < n:
            ln = lines[j]
            if FENCE.match(ln):
                buf.append(ln)
                j += 1
                while j < n and not FENCE.match(lines[j]):
                    buf.append(lines[j])
                    j += 1
                if j < n:
                    buf.append(lines[j])
                    j += 1
                continue
            if ln.strip() == "":
                k = j
                while k < n and lines[k].strip() == "":
                    k += 1
                if (k < n and SUB.match(lines[k]) and "[HARD]" not in lines[k]
                        and not HEAD.match(lines[k]) and not HR.match(lines[k])):
                    buf.extend(lines[j:k])
                    j = k
                    continue
                break
            if HEAD.match(ln) or HR.match(ln) or "[HARD]" in ln:
                break
            buf.append(ln)
            j += 1
        text = "\n".join(buf).rstrip("\n") + "\n"
        res.append({
            "file": os.path.relpath(path, ROOT),
            "line": start + 1,
            "endline": start + len(buf),
            "bytes": len(text.encode("utf-8")),
            "clause_initial": bool(INIT.match(lines[start])),
            "text": text,
        })
        i = j if j > i else i + 1
    return res


def main():
    files = always_loaded_rules() + [os.path.join(ROOT, "CLAUDE.md")]
    all_blocks = []
    for p in files:
        all_blocks.extend(blocks(p))
    print(json.dumps(all_blocks, ensure_ascii=False))


if __name__ == "__main__":
    main()
