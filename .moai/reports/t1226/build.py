#!/usr/bin/env python3
"""Candidate-table builder for SPEC-ALWAYS-LOADED-HEADROOM-001 (card t1226).

Usage:
  build.py --surface {init|live} --root SURFACE_ROOT --out-tree DIR
           [--ref-root INIT_ROOT] [--hash POST_HASH]

Joins the mechanical rows of candidates.py with the human judgment layer
(judgments.tsv), the in-place compression attempts (m2/replacements.txt) and
the pointer lines (pointers.tsv), then:

  * always: writes the post-attempt surface (every ADMIT M1 row removed, every
    attempted M2 row replaced) under --out-tree so the frozen-multiset hash of
    the attempted state can be measured with the AC-ALH-004 shell pipeline;
  * with --hash: writes candidates-<s>.tsv, pointers-<s>.tsv and the evidence
    files, stamping post_hash = POST_HASH into every M2 evidence file.

Rules encoded here (acceptance.md §D TSV column contract):
  * a partition row with bind > 0 is REJECT bind>0 (mech none);
  * a yaml row is REJECT config-data;
  * every other row must have exactly one judgment (surface-specific rows win
    over "*"); a missing judgment is an error, never a default;
  * an M2 row planned "try" must have a replacement block; the attempt's
    chars = pre_chars - post_chars, and a longer replacement is an error;
  * a "*" replacement is reused on the live surface only when the live row's
    text is byte-identical to the init row it was authored against.
"""
import argparse
import hashlib
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import candidates as C  # noqa: E402

RULES = ".claude/rules/moai/"
REL = ".moai/reports/t1226/"


def short(rel):
    return rel[len(RULES):] if rel.startswith(RULES) else rel


def rid(f, sec):
    return hashlib.sha1(("%s\t%s" % (f, sec)).encode("utf-8")).hexdigest()[:10]


def read_rows(root, rel):
    text = open(os.path.join(root, rel), encoding="utf-8", newline="").read()
    lines = text.splitlines(keepends=True)
    if rel.endswith((".yaml", ".yml")):
        return lines, [("(전체)", 0, len(lines), False)]
    return lines, C.split_spans(lines)


def load_judgments():
    out = {}
    with open(os.path.join(HERE, "judgments.tsv"), encoding="utf-8") as fh:
        head = fh.readline().rstrip("\n").split("\t")
        for line in fh:
            if not line.strip():
                continue
            v = dict(zip(head, line.rstrip("\n").split("\t")))
            k = (v["surf"], v["file"], v["section"])
            if k in out:
                sys.exit("duplicate judgment %r" % (k,))
            out[k] = v
    return out


def load_blocks(path, fields):
    """Blocks: '@@ <surf> | <file> | <section>' ... '@@END'."""
    out = {}
    if not os.path.exists(path):
        return out
    cur, buf = None, []
    for line in open(path, encoding="utf-8"):
        if line.startswith("@@END"):
            out[cur] = "".join(buf)
            cur, buf = None, []
            continue
        if line.startswith("@@ ") and cur is None:
            parts = [p.strip() for p in line[3:].rstrip("\n").split(" | ")]
            if len(parts) != fields:
                sys.exit("bad block header: %r" % line)
            cur = tuple(parts)
            if cur in out:
                sys.exit("duplicate block %r" % (cur,))
            continue
        if cur is not None:
            buf.append(line)
    if cur is not None:
        sys.exit("unterminated block %r" % (cur,))
    return out


def pick(table, s, f, sec):
    return table.get((s, f, sec)) or table.get(("*", f, sec))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--surface", required=True, choices=["init", "live"])
    ap.add_argument("--root", required=True)
    ap.add_argument("--out-tree", required=True)
    ap.add_argument("--ref-root")
    ap.add_argument("--hash")
    a = ap.parse_args()
    s = a.surface

    judg = load_judgments()
    repl = {}
    for name in sorted(os.listdir(os.path.join(HERE, "m2"))):
        if name.startswith("replacements") and name.endswith(".txt"):
            part = load_blocks(os.path.join(HERE, "m2", name), 3)
            dup = set(part) & set(repl)
            if dup:
                sys.exit("duplicate block across files: %r" % (sorted(dup),))
            repl.update(part)
    ptrs = load_blocks(os.path.join(HERE, "pointers.txt"), 3)
    sizes = {}
    for line in open(os.path.join(HERE, "dest-sizes.tsv"), encoding="utf-8").readlines()[1:]:
        d, lv, tv = line.rstrip("\n").split("\t")
        sizes[d] = (lv, tv)
    paths = [p.strip() for p in open(os.path.join(HERE, "count-set-%s.txt" % s), encoding="utf-8") if p.strip()]

    used_j, used_r, used_p = set(), set(), set()
    rows_out, m1_pairs, errors = [], {}, []
    for rel in paths:
        f = short(rel)
        lines, spans = read_rows(a.root, rel)
        ref = {}
        if a.ref_root and os.path.exists(os.path.join(a.ref_root, rel)):
            rl, rsp = read_rows(a.ref_root, rel)
            ref = {n: "".join(rl[x:y]) for n, x, y, _ in rsp}
        new = list(lines)  # line-slot rewrite: each slot keeps its index
        for name, x, y, isp in spans:
            body = lines[x:y]
            text = "".join(body)
            gross = len(text)
            bind = C.nbind(body)
            row = {"file": rel, "section": name, "gross": gross, "chars": "-", "mech": "none",
                   "bind": bind, "gov": "N", "c1": "N", "c2": "NA", "c3": "NA", "c4": "NA",
                   "dest": "-", "verdict": "REJECT", "reason": "-", "evidence": ""}
            if f.endswith(".yaml"):
                row.update(c1="Y", reason="config-data", evidence=REL + "evidence/config-data.md")
                rows_out.append(row)
                continue
            if not isp and bind > 0:
                row.update(reason="bind>0", evidence=REL + "evidence/bind-%s.md" % s,
                           _lines=[l for l in body if C.BIND.search(l)])
                rows_out.append(row)
                continue
            j = pick(judg, s, f, name)
            if j is None:
                errors.append("no judgment: %s | %s" % (f, name))
                continue
            used_j.add((j["surf"], j["file"], j["section"]))
            row.update(mech=j["mech"], gov=j["gov"], c2=j["c2"], c3=j["c3"], c4=j["c4"], dest=j["dest"])
            row["c1"] = "Y" if (bind == 0 and j["gov"] == "N") else "N"
            plan = j["plan"]
            if plan.startswith("REJECT:"):
                row.update(verdict="REJECT", reason=plan[7:], evidence=REL + "evidence/rejects.md")
            elif plan.startswith("UNTRIED:"):
                row.update(verdict="UNTRIED", reason="untried:" + plan[8:], evidence=REL + "evidence/untried.md")
            elif plan == "move":
                ev = REL + "evidence/m1/%s/%s.md" % (s, rid(f, name))
                row.update(verdict="ADMIT", chars=gross, reason="-", evidence=ev, _text=text)
                m1_pairs.setdefault((f, j["dest"]), []).append(row)
                for i in range(x, y):
                    new[i] = ""
            elif plan == "try":
                r = repl.get((s, f, name))
                key = (s, f, name)
                if r is None:
                    r = repl.get(("*", f, name))
                    key = ("*", f, name)
                    if r is not None and s == "live" and ref.get(name) != text:
                        errors.append("live text differs from init; needs a live block: %s | %s" % (f, name))
                        continue
                if r is None:
                    errors.append("no replacement: %s | %s" % (f, name))
                    continue
                used_r.add(key)
                body_txt = text.rstrip("\n")
                suffix = text[len(body_txt):]
                rtxt = r.rstrip("\n")
                post = (rtxt + suffix) if rtxt else ""
                if C.BIND.search(post):
                    errors.append("replacement introduces a keyword line: %s | %s" % (f, name))
                if len(post) > gross:
                    errors.append("replacement longer than original: %s | %s" % (f, name))
                ev = REL + "evidence/m2/%s/%s.md" % (s, rid(f, name))
                row.update(verdict="ADMIT", chars=gross - len(post), reason="-", evidence=ev,
                           _pre=text, _post=post)
                new[x] = post
                for i in range(x + 1, y):
                    new[i] = ""
            else:
                errors.append("bad plan %r: %s | %s" % (plan, f, name))
                continue
            rows_out.append(row)
        dst = os.path.join(a.out_tree, rel)
        os.makedirs(os.path.dirname(dst), exist_ok=True)
        with open(dst, "w", encoding="utf-8", newline="") as fh:
            fh.write("".join(new))

    # pointers: one per (file, dest) pair of ADMIT M1 rows
    ptr_rows = []
    for (f, d), rs in sorted(m1_pairs.items()):
        p = ptrs.get((s, f, d)) or ptrs.get(("*", f, d))
        if p is None:
            errors.append("no pointer for %s -> %s" % (f, d))
            continue
        used_p.add((s, f, d) if (s, f, d) in ptrs else ("*", f, d))
        ptxt = p.rstrip("\n")
        ptr_rows.append((rs[0]["file"], d, len(ptxt), ptxt, rs))

    if errors:
        sys.exit("ERRORS:\n  " + "\n  ".join(errors))

    total = sum(r["gross"] for r in rows_out if " ¶" not in r["section"] or not re.search(r" ¶[0-9]+$", r["section"]))
    A = sum(r["chars"] for r in rows_out if r["verdict"] == "ADMIT")
    U = sum(r["gross"] for r in rows_out if r["verdict"] == "UNTRIED")
    R = sum(p[2] for p in ptr_rows)
    print("surface=%s total=%d A_adm=%d R=%d U=%d T_min=%d" % (s, total, A, R, U, total - A + R))
    x = ".claude/rules/moai/workflow/skill-routing.md"
    t17 = sum(r["gross"] for r in rows_out if r["file"] != x and not re.search(r" ¶[0-9]+$", r["section"]))
    A17 = sum(r["chars"] for r in rows_out if r["file"] != x and r["verdict"] == "ADMIT")
    U17 = sum(r["gross"] for r in rows_out if r["file"] != x and r["verdict"] == "UNTRIED")
    R17 = sum(p[2] for p in ptr_rows if p[0] != x)
    print("surface=%s_17 total=%d A_adm=%d R=%d U=%d T_min=%d" % (s, t17, A17, R17, U17, t17 - A17 + R17))
    m2 = [r for r in rows_out if r["verdict"] == "ADMIT" and r["mech"] == "M2"]
    print("m2_attempted=%d m2_pre=%d m2_saved=%d m1_rows=%d m1_chars=%d" % (
        len(m2), sum(r["gross"] for r in m2), sum(r["chars"] for r in m2),
        sum(len(p[4]) for p in ptr_rows), sum(r["chars"] for p in ptr_rows for r in p[4])))

    unused = [k for k in judg if k not in used_j and k[0] in ("*", s)]
    if unused:
        print("note: judgments unused on this surface: %d" % len(unused))
    if not a.hash:
        return
    write_outputs(s, rows_out, ptr_rows, sizes, a.hash, judg)


def write_outputs(s, rows, ptr_rows, sizes, post_hash, judg):
    base = os.path.join(HERE)
    cols = C.COLS
    with open(os.path.join(base, "candidates-%s.tsv" % s), "w", encoding="utf-8") as fh:
        fh.write("\t".join(cols) + "\n")
        for r in rows:
            fh.write("\t".join(str(r[c]) for c in cols) + "\n")
    with open(os.path.join(base, "pointers-%s.tsv" % s), "w", encoding="utf-8") as fh:
        fh.write("file\tdest\tpointer_chars\n")
        for f, d, n, _, _ in ptr_rows:
            fh.write("%s\t%s\t%d\n" % (f, d, n))
    ev = os.path.join(base, "evidence")
    for sub in ("m1/%s" % s, "m2/%s" % s, "pointers/%s" % s):
        os.makedirs(os.path.join(ev, sub), exist_ok=True)
    for f, d, n, ptxt, rs in ptr_rows:
        pid = rid(f, d)
        with open(os.path.join(ev, "pointers", s, pid + ".txt"), "w", encoding="utf-8") as fh:
            fh.write(ptxt)
        for r in rs:
            lv, tv = sizes[d]
            with open(os.path.join(HERE, "..", "..", "..", r["evidence"]), "w", encoding="utf-8") as fh:
                fh.write("# M1 — %s | %s\n\n" % (r["file"], r["section"]))
                fh.write("surface = %s\nmech = M1\ngross = %d\nchars = %d\ndest = %s\n" % (s, r["gross"], r["chars"], d))
                fh.write("pointer_file = %sevidence/pointers/%s/%s.txt\npointer_chars = %d\n" % (REL, s, pid, n))
                fh.write("c2 = Y (a paragraph row has no heading, so no `§ <title>` citation can target it)\n")
                fh.write("c3 = Y (destination paths: see evidence/dest-paths.txt)\n")
                fh.write("c4 = Y (dest size live %s / template %s; cumulative check: AC-ALH-003 (4))\n" % (lv, tv))
                fh.write("note = %s\n\n" % judg.get(("*", short(r["file"]), r["section"]), {}).get("note", ""))
                fh.write("## moved text\n\n````\n%s````\n" % "".join(_row_text(r)))
    with open(os.path.join(ev, "bind-%s.md" % s), "w", encoding="utf-8") as fh:
        fh.write("# REJECT bind>0 — surface %s\n\nEach partition row below holds keyword lines "
                 "(`\\[HARD\\]|MUST|shall `); condition 1 fails, so the row is not relocatable or "
                 "compressible under the freeze. Its non-binding paragraphs are separate `¶n` rows.\n\n" % s)
        for r in rows:
            if r["reason"] == "bind>0":
                fh.write("## %s | %s (gross %d, bind %d)\n\n" % (r["file"], r["section"], r["gross"], r["bind"]))
                for l in r["_lines"]:
                    fh.write("- `%s`\n" % l.strip().replace("`", "'")[:200])
                fh.write("\n")
    with open(os.path.join(ev, "rejects.md"), "w", encoding="utf-8") as fh:
        fh.write("# REJECT rows from the judgment layer (both surfaces)\n\n"
                 "Source of truth: `.moai/reports/t1226/judgments.tsv` (columns gov / plan / note). "
                 "gov=a: a keyword line depends on the row (table, heading, enumeration, definition, "
                 "loading/applicability scope); gov=b: the row narrows a keyword line's scope "
                 "(exception, exemption, carve-out, boundary). citation-breaks rows: see "
                 "`evidence/citations-claude.txt`. Rows flagged in design.md §4.3 keep their §4.3 "
                 "classification.\n\n| surf | file | section | gov | reason | why |\n|---|---|---|---|---|---|\n")
        for k, j in judg.items():
            if j["plan"].startswith("REJECT:"):
                fh.write("| %s | %s | %s | %s | %s | %s |\n" % (j["surf"], j["file"], j["section"].replace("|", "/"),
                                                              j["gov"], j["plan"][7:], j["note"].replace("|", "/")))
    with open(os.path.join(ev, "config-data.md"), "w", encoding="utf-8") as fh:
        fh.write("# REJECT config-data\n\n`.moai/config/sections/user.yaml` and `language.yaml` are "
                 "configuration data imported by CLAUDE.md §9, not instruction prose; REQ-ALH-003 records "
                 "them as `config-data` rejections.\n")
    with open(os.path.join(ev, "untried.md"), "w", encoding="utf-8") as fh:
        fh.write("# UNTRIED rows\n\n")
        for k, j in judg.items():
            if j["plan"].startswith("UNTRIED:"):
                fh.write("- %s | %s — untried:%s — %s\n" % (j["file"], j["section"], j["plan"][8:], j["note"]))
    for r in rows:
        if r["verdict"] == "ADMIT" and r["mech"] == "M2":
            with open(os.path.join(HERE, "..", "..", "..", r["evidence"]), "w", encoding="utf-8") as fh:
                fh.write("# M2 attempt — %s | %s\n\n" % (r["file"], r["section"]))
                fh.write("surface = %s\npre_chars = %d\npost_chars = %d\nchars = %d\npost_hash = %s\n\n" % (
                    s, len(r["_pre"]), len(r["_post"]), r["chars"], post_hash))
                fh.write("post_hash is the frozen-multiset sha256 of the whole post-attempt surface "
                         "(every attempted M2 row replaced, every ADMIT M1 row removed), measured with the "
                         "AC-ALH-004 pipeline on $SCRATCH/post-%s.\n\n" % s)
                fh.write("## before\n\n````\n%s````\n\n## after\n\n````\n%s````\n" % (r["_pre"], r["_post"]))


def _row_text(r):
    return r["_text"]


if __name__ == "__main__":
    main()
