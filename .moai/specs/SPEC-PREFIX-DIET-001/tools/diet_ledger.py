#!/usr/bin/env python3
"""Ledger tool for SPEC-PREFIX-DIET-001 (run phase, card t1450).

Commands (run from the worktree root):
  init                 build the three fixtures from the anchor blobs (git show <ANCHOR>:<path>)
  apply <style-file>   flip the planned `dropped` rows for one style and rewrite the template file
  sizes                print whole-file UTF-16 sizes and droppable totals

The unit extractor mirrors the Go helper in internal/template/output_style_diet_helpers_test.go
(spec.md section B). Unit texts concatenated reproduce the body byte for byte.
"""
import hashlib
import json
import re
import subprocess
import sys

ANCHOR = "5d5ff1aae"
STYLE_DIR = "internal/template/templates/.claude/output-styles/moai/"
LEDGER = "internal/template/testdata/output_style_ledger.json"
FROZEN = "internal/template/testdata/output_style_frozen.json"
LOCAL = "internal/template/testdata/output_style_localization.json"
FILES = ["moai.md", "moai-easy.md", "moai-learn.md"]

HEADING = re.compile(r"^#{1,6}\s")
TABLE = re.compile(r"^\s*\|")
LIST = re.compile(r"^\s*([-*+]|\d+[.)])\s")
FENCE = re.compile(r"^\s*```")

# Frozen handoff sections: (file, title line, level). The section runs from the heading unit up to
# (not including) the next heading unit of the same or a higher level, or the next `---` unit.
FROZEN_SECTIONS = {
    "moai.md": [("### Session Boundary Handoff [HARD]", 3), ("### Session Handoff [HARD]", 3)],
    "moai-easy.md": [("### Banner 7 — Picking Up Next Time (Session Handoff)", 3)],
    "moai-learn.md": [],
}

# Planned drops. index -> (kind, survivor, note). Only rationale / example units, never a unit that
# carries a binding token (enforced below). `---` units are separators of a dropped section.
SEP = "separator that bounds a dropped section; carries no content"


def _range(a, b):
    return list(range(a, b + 1))


PLAN = {
    "moai-easy.md": {
        # kind per range, survivor, note
        "groups": [
            (_range(44, 45), "rationale", "moai-easy.md section 5 'In plain words' / 'When I delegate vs. do it myself'",
             "reassurance for beginners; the delegation behaviour is stated in the surviving section 5 units"),
            (_range(53, 71), "example", "moai-easy.md section 6 'How it works' (the first-mention pattern and its blockquote example)",
             "sixteen sample glossary entries; the plain-language rule itself and a worked first-mention example survive"),
            (_range(72, 73), "rationale", "moai-easy.md section 11 table row 'Lost on a term' and section 6 'How it works'",
             "restates the pause-and-explain behaviour that the section 11 table row already specifies"),
            (_range(134, 142), "example", "moai-easy.md section 7 banner skeletons (Banner 1-6 templates)",
             "worked instances of the six banners; the skeleton of each banner survives in section 7"),
            (_range(158, 169), "rationale", "moai-easy.md section 1 (switching styles, who I am), section 4 (check step), section 5 (delegation), section 11 (I'm lost)",
             "FAQ restating answers already given in sections 1, 4, 5 and 11; switch mechanism survives in section 1"),
            (_range(170, 174), "rationale", "moai-easy.md section 2 'My Promise to You (Operating Principles)'",
             "teaching philosophy restates the operating principles of section 2"),
            (_range(175, 182), "rationale", "moai-easy.md section 1 sibling table and switch instruction",
             "style-switch quick reference duplicates the section 1 sibling table and the /output-style switch mechanism"),
            (_range(183, 184), "rationale", "moai-easy.md section 11 situation table",
             "friendly reminders restate the section 11 situation/response table"),
        ],
    },
    "moai-learn.md": {
        "groups": [
            (_range(5, 7), "rationale", "moai-learn.md section 1 'Core Mission' and section 3 'Phase 2 - Teach'",
             "principle quote and upfront-honesty paragraph restate the jargon-free first-pass rule of Phase 2"),
            (_range(122, 122), "rationale", "moai-learn.md section 8 anti-pattern catalogue intro (units before the table)",
             "root-cause narrative for the catalogue; the HARD violation statement and the catalogue itself survive"),
            (_range(137, 141), "rationale", "moai-learn.md section 1 'Core Mission' and section 3 phase units",
             "teaching philosophy restates the mission bullets and the five-phase protocol"),
        ],
    },
    "moai.md": {
        "groups": [
            (_range(6, 7), "rationale", "moai.md section 1 'Operating Principles'",
             "core-trait list restates principles 3 and 5 and the language rule of section 9"),
            (_range(64, 65), "rationale", "moai.md section 6 persistence paragraph (memory directory path) and .claude/rules/moai/workflow/moai-memory.md",
             "provenance claim and an explanatory note pointing at the memory rule; the persistence behaviour itself survives"),
            (_range(246, 250), "rationale", "moai.md section 1 'Operating Principles'",
             "service philosophy restates the operating principles of section 1"),
        ],
    },
}


def git_show(path):
    return subprocess.run(["git", "show", "%s:%s" % (ANCHOR, path)], capture_output=True, text=True, check=True).stdout


def split_frontmatter(text):
    lines = text.split("\n")
    assert lines[0] == "---"
    for i in range(1, len(lines)):
        if lines[i] == "---":
            return "\n".join(lines[: i + 1]) + "\n", "\n".join(lines[i + 1 :])
    raise ValueError("frontmatter not closed")


def extract_units(body):
    lines = body.split("\n")
    units = []
    cur = None
    cur_kind = None
    in_fence = False
    gap = 0
    pending = []
    for ln in lines:
        blank = ln.strip() == ""
        if in_fence:
            cur.append(ln)
            if FENCE.match(ln):
                in_fence = False
            continue
        if blank:
            gap += 1
            pending.append(ln)
            continue
        is_head = bool(HEADING.match(ln))
        is_table = bool(TABLE.match(ln))
        is_list = bool(LIST.match(ln))
        is_fence = bool(FENCE.match(ln))
        if cur is None:
            new = True
        elif is_head or is_table:
            new = True
        elif cur_kind == "table":
            new = True
        elif gap == 0:
            new = False
        elif gap == 1 and (is_list or is_fence):
            new = False
        else:
            new = True
        if new:
            if cur is not None:
                cur.extend(pending)
                units.append(cur)
                pending = []
                cur = []
            else:
                cur = list(pending)
                pending = []
            cur_kind = "table" if is_table else ("heading" if is_head else "text")
        else:
            cur.extend(pending)
            pending = []
        cur.append(ln)
        if is_fence:
            in_fence = True
        gap = 0
    if cur is not None:
        cur.extend(pending)
        units.append(cur)
    out = []
    for i, u in enumerate(units):
        out.append("\n".join(u) + ("\n" if i < len(units) - 1 else ""))
    assert "".join(out) == body, "tiling failure"
    return out


def tokens(s):
    mn = s.count("MUST NOT")
    return [s.count("[HARD]"), mn, s.count("MUST") - mn, s.count("shall ")]


def u16(s):
    return len(s.encode("utf-16-le")) // 2


def sha(s):
    return hashlib.sha256(s.encode("utf-8")).hexdigest()


def first_line(u):
    for ln in u.split("\n"):
        if ln.strip():
            return ln
    return ""


def frozen_ranges(units, file):
    out = []
    for title, level in FROZEN_SECTIONS[file]:
        start = None
        for i, u in enumerate(units):
            if first_line(u) == title:
                start = i
                break
        assert start is not None, (file, title)
        end = start
        j = start + 1
        while j < len(units):
            fl = first_line(units[j])
            m = re.match(r"^(#{1,6})\s", fl)
            if fl.strip() == "---" or (m and len(m.group(1)) <= level):
                break
            end = j
            j += 1
        out.append((title, level, start, end))
    return out


def current_heading(units, i):
    h = ""
    for k in range(i + 1):
        fl = first_line(units[k])
        if HEADING.match(fl):
            h = fl
    return h


def dump_rows(rows):
    return "[\n" + ",\n".join(json.dumps(r, ensure_ascii=False) for r in rows) + "\n]"


def write_ledger(meta, rows):
    txt = "{\n"
    txt += '"anchor": %s,\n' % json.dumps(ANCHOR)
    txt += '"files": %s,\n' % json.dumps(meta, ensure_ascii=False, indent=1)
    txt += '"rows": %s\n}\n' % dump_rows(rows)
    open(LEDGER, "w", encoding="utf-8").write(txt)


def init():
    meta = {}
    rows = []
    frozen_doc = {"anchor": ANCHOR, "sections": []}
    local_doc = {"anchor": ANCHOR, "tables": []}
    for f in FILES:
        full = git_show(STYLE_DIR + f)
        head, body = split_frontmatter(full)
        units = extract_units(body)
        franges = frozen_ranges(units, f)
        frozen_idx = set()
        meta_frozen = []
        for title, level, a, b in franges:
            text = "".join(units[a : b + 1])
            frozen_doc["sections"].append(
                {"file": f, "title": title, "level": level, "first_index": a, "last_index": b, "sha256": sha(text), "text": text}
            )
            meta_frozen.append({"title": title, "first_index": a, "last_index": b})
            frozen_idx.update(range(a, b + 1))
        meta[f] = {
            "anchor_body_sha256": sha(body),
            "anchor_frontmatter_sha256": sha(head),
            "anchor_units": len(units),
            "anchor_file_utf16": u16(full),
            "anchor_tokens": tokens(body),
            "frozen": meta_frozen,
        }
        planned = {}
        for idxs, kind, survivor, note in PLAN[f]["groups"]:
            for i in idxs:
                planned[i] = kind
        for i, u in enumerate(units):
            if i in frozen_idx:
                continue
            t = tokens(u)
            if sum(t):
                kind = "binding"
            elif i in planned:
                kind = planned[i]
            else:
                kind = "normative"
            assert not (sum(t) and i in planned), (f, i, "planned drop carries a binding token")
            rows.append(
                {
                    "id": "%s-%04d" % (f.replace(".md", ""), i),
                    "file": f,
                    "index": i,
                    "kind": kind,
                    "source": "%s | %s" % (f, current_heading(units, i)),
                    "before_text": u,
                    "after_text": u,
                    "treatment": "verbatim",
                    "survivor": "",
                    "note": "",
                }
            )
        # localization tables
        lines = body.split("\n")
        k = 0
        tno = 0
        while k < len(lines):
            if TABLE.match(lines[k]):
                j = k
                while j < len(lines) and TABLE.match(lines[j]):
                    j += 1
                block = lines[k:j]
                hdr = block[0]
                if "Korean" in hdr or "ko canonical" in hdr:
                    tno += 1
                    local_doc["tables"].append({"file": f, "id": "%s-t%02d" % (f.replace(".md", ""), tno), "rows": block})
                k = j
            else:
                k += 1
    write_ledger(meta, rows)
    open(FROZEN, "w", encoding="utf-8").write(json.dumps(frozen_doc, ensure_ascii=False, indent=1) + "\n")
    open(LOCAL, "w", encoding="utf-8").write(json.dumps(local_doc, ensure_ascii=False, indent=1) + "\n")
    print("init ok", {f: meta[f]["anchor_units"] for f in FILES}, "rows", len(rows))


def load_ledger():
    return json.load(open(LEDGER, encoding="utf-8"))


def reconstruct(doc, frozen_doc, f, after=True):
    meta = doc["files"][f]
    rows = {r["index"]: r for r in doc["rows"] if r["file"] == f}
    fro = {s["first_index"]: s for s in frozen_doc["sections"] if s["file"] == f}
    out = []
    i = 0
    n = meta["anchor_units"]
    while i < n:
        if i in fro:
            out.append(fro[i]["text"])
            i = fro[i]["last_index"] + 1
            continue
        r = rows[i]
        out.append(r["after_text"] if after else r["before_text"])
        i += 1
    return "".join(out)


def apply(f):
    doc = load_ledger()
    frozen_doc = json.load(open(FROZEN, encoding="utf-8"))
    for idxs, kind, survivor, note in PLAN[f]["groups"]:
        for r in doc["rows"]:
            if r["file"] == f and r["index"] in idxs:
                assert r["kind"] in ("rationale", "example"), (f, r["index"], r["kind"])
                assert sum(tokens(r["before_text"])) == 0
                is_sep = first_line(r["before_text"]).strip() == "---"
                r["treatment"] = "dropped"
                r["after_text"] = ""
                r["survivor"] = survivor
                r["note"] = (SEP + "; " + note) if is_sep else note
    write_ledger(doc["files"], doc["rows"])
    full = git_show(STYLE_DIR + f)
    head, _ = split_frontmatter(full)
    new = head + reconstruct(doc, frozen_doc, f, after=True)
    open(STYLE_DIR + f, "w", encoding="utf-8").write(new)
    print("apply ok", f, "utf16", u16(full), "->", u16(new))


def sizes():
    tot = 0
    for f in FILES:
        full = git_show(STYLE_DIR + f)
        head, body = split_frontmatter(full)
        units = extract_units(body)
        planned = {}
        for idxs, kind, survivor, note in PLAN[f]["groups"]:
            for i in idxs:
                planned[i] = kind
        d = sum(u16(units[i]) for i in planned)
        print("%-14s file=%d body=%d units=%d droppable=%d (rationale=%d example=%d) tokens=%s" % (
            f, u16(full), u16(body), len(units), d,
            sum(u16(units[i]) for i, k in planned.items() if k == "rationale"),
            sum(u16(units[i]) for i, k in planned.items() if k == "example"), tokens(body)))
        tot += d
    print("total droppable", tot)


if __name__ == "__main__":
    cmd = sys.argv[1]
    if cmd == "init":
        init()
    elif cmd == "apply":
        apply(sys.argv[2])
    elif cmd == "sizes":
        sizes()
    else:
        raise SystemExit("unknown command")
