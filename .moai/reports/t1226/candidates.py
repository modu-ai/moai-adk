#!/usr/bin/env python3
"""Candidate enumerator for SPEC-ALWAYS-LOADED-HEADROOM-001 (card t1226).

Usage:
  candidates.py [--keys] --paths COUNT_SET_FILE SURFACE_ROOT

Splits every file of a surface's count set into a flat, non-overlapping
partition whose per-file gross sum equals `wc -m` of that file:

  * markdown: the preamble (text before the first heading, frontmatter
    included) is one row "(서문)"; every heading of level <= 3 outside a
    code fence opens a section row that runs to the next such heading
    (the same boundary rule as sec.py);
  * yaml: one row "(전체)".

For every markdown section that holds at least one binding line
(\\[HARD\\]|MUST|shall ), each blank-line-delimited paragraph after the
heading line that holds zero binding lines is emitted as an extra row
"<section> ¶n" (n = ordinal of the paragraph among all paragraphs of the
section, fences kept whole). Paragraph rows are NOT part of the partition.

gross counts characters (code points, line endings included), which is what
`wc -m` reports under a UTF-8 locale. --keys prints only file, section,
gross, bind; otherwise the 15-column TSV with placeholder judgment columns.
"""
import re
import sys

BIND = re.compile(r'\[HARD\]|MUST|shall ')
HEAD = re.compile(r'^(#{1,6}) (.*)$')
MAXLVL = 3
COLS = ["file", "section", "gross", "chars", "mech", "bind", "gov", "c1", "c2",
        "c3", "c4", "dest", "verdict", "reason", "evidence"]


def nbind(lines):
    return sum(1 for l in lines if BIND.search(l))


def split_markdown(text):
    """Return a list of (section_name, lines, is_paragraph) rows."""
    lines = text.splitlines(keepends=True)
    infence = False
    heads = []
    for i, l in enumerate(lines):
        if l.lstrip().startswith("```"):
            infence = not infence
            continue
        if infence:
            continue
        m = HEAD.match(l.rstrip("\r\n"))
        if m and len(m.group(1)) <= MAXLVL:
            heads.append((i, m.group(2).strip()))
    rows = []
    first = heads[0][0] if heads else len(lines)
    if first > 0:
        rows.append(("(서문)", lines[:first], False))
    seen = {}
    bounds = [h[0] for h in heads] + [len(lines)]
    for (a, title), b in zip(heads, bounds[1:]):
        seen[title] = seen.get(title, 0) + 1
        name = title if seen[title] == 1 else "%s (#%d)" % (title, seen[title])
        body = lines[a:b]
        rows.append((name, body, False))
        if nbind(body) > 0:
            for n, para in enumerate(paragraphs(body[1:]), start=1):
                if nbind(para) == 0:
                    rows.append(("%s ¶%d" % (name, n), para, True))
    return rows


def paragraphs(lines):
    """Blank-line-delimited blocks; a fenced block is never split."""
    out, cur, infence = [], [], False
    for l in lines:
        if l.lstrip().startswith("```"):
            infence = not infence
            cur.append(l)
            continue
        if not infence and l.strip() == "":
            if cur:
                out.append(cur)
                cur = []
            continue
        cur.append(l)
    if cur:
        out.append(cur)
    return out


def rows_for(root, rel):
    text = open("%s/%s" % (root, rel), encoding="utf-8", newline="").read()
    if rel.endswith((".yaml", ".yml")):
        return [("(전체)", text.splitlines(keepends=True), False)]
    return split_markdown(text)


def main(argv):
    keys = "--keys" in argv
    argv = [a for a in argv if a != "--keys"]
    if len(argv) != 3 or argv[0] != "--paths":
        sys.exit(__doc__)
    paths = [p.strip() for p in open(argv[1], encoding="utf-8") if p.strip()]
    root = argv[2]
    if not keys:
        print("\t".join(COLS))
    for rel in paths:
        for name, body, _ in rows_for(root, rel):
            gross = sum(len(l) for l in body)
            bind = nbind(body)
            if keys:
                print("%s\t%s\t%d\t%d" % (rel, name, gross, bind))
            else:
                print("\t".join([rel, name, str(gross), "-", "none", str(bind), "N", "-",
                                 "NA", "NA", "NA", "-", "UNTRIED", "-", "-"]))


if __name__ == "__main__":
    main(sys.argv[1:])
