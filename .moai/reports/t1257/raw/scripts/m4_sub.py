#!/usr/bin/env python3
"""M4 vocabulary sweep — SPEC-ROLE-NAMING-DOCS-001 (card t1257).

Applies the lead->leader / worker->lane vocabulary substitutions per the M2
term table (progress.md §E.2), REQ-RND-021 first-occurrence qualifiers,
REQ-RND-023 auxiliary-role lines, REQ-RND-025 alias-line rewrites, and the
REQ-RND-011 kept-name sentence. Template copies first, then local copies.
Appends one ledger row per changed line per copy; verifies [HARD] counts;
re-keys surviving ledger rows across insertion shifts.
"""
import difflib
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
LEDGER = ROOT / ".moai/reports/t1257/ledger.tsv"

FOREMAN_LINE = ("`foreman` — an auxiliary role of the leader: the unattended watcher "
                "that dispatches the already-picked card to an isolated worker when no "
                "leader session holds the board.")
DEPUTY_LINE = ("`deputy` — an auxiliary role of the leader: the resident background "
               "manager-lead agent that reads raw-tree evidence and reports `RECOMMEND:` "
               "summaries, holding no power of consequence.")
KEPTNAME_LINE = ("The agent name `manager-lead` is kept, and the role it coordinates is "
                 "called the leader: this file is the leader's coordination agent — the "
                 "name is kept while the role is called leader.")

W = ".claude/rules/moai/workflow/"
C = ".claude/rules/moai/core/"

FILES = [
    (W + "kanban-dispatch.md", [
        ("ins_after", "## Deputy dispatch surface", DEPUTY_LINE),
        ("sweep",),
    ]),
    (W + "kanban-dispatch-detail.md", [
        ("pair", "A worker session launched by hand", "A session launched by hand"),
        ("pair",
         "Factory Mode companions are labelled `worker-1..worker-N` — `lane` stays the prose term for the slot.",
         "Factory Mode lanes are labelled `lane-1..lane-N` (`lane-<n>` is the session label, and `-f lane` joins as the next free one)."),
        ("pair", "one lead plus `worker-1..worker-N` sessions",
         "one leader plus `lane-1..lane-N` sessions"),
        ("sweep",),
    ]),
    (W + "kanban-dispatch-mechanics.md", [("sweep",)]),
    (W + "cross-session-messaging.md", [
        ("pair", "one coordinating session and workers that each own a stage of the pipeline",
         "one coordinating session and lane sessions that each own a stage of the pipeline"),
        ("pair", "each worker writes to an isolated tree so concurrent workers cannot collide",
         "each lane writes to an isolated tree so concurrent lanes cannot collide"),
        ("sweep",),
    ]),
    (W + "cross-session-messaging-detail.md", [("sweep",)]),
    (W + "orchestration-mode-selection.md", [
        ("pair", "one team per session; the lead is fixed",
         "one team per session; the team lead is fixed"),
        ("pair", "whether teammates inherit the lead's `ANTHROPIC_BASE_URL`",
         "whether teammates inherit the team lead's `ANTHROPIC_BASE_URL`"),
        ("pair",
         "runs a fleet of independent worker sessions (tmux panes) — N is an operator-side fleet size (the count-less worker entry, a bare `-k --name worker-<i>`, defaults to 8)",
         "runs a fleet of independent lane sessions (tmux panes) — N is an operator-side fleet size (the count-less lane entry, a bare `-k --name lane-<i>`, defaults to 8)"),
        ("pair", "the factory's own workers-registry / free-slot discipline",
         "the factory's own lane-registry / free-slot discipline"),
        ("sweep",),
    ]),
    (C + "agent-common-protocol.md", [("sweep",)]),
    (C + "moai-constitution-detail.md", [("sweep",)]),
    (".claude/skills/moai/workflows/gtd.md", [
        ("pair", "`auto-done` is a LEAD surface", "`auto-done` is a LEADER surface"),
        ("sweep",),
    ]),
    (".claude/output-styles/moai/moai.md", [("sweep",)]),
    (".claude/skills/moai-kanban-foreman/SKILL.md", [
        ("ins_after",
         "card classes live in the kanban dispatch rule (`.claude/rules/moai/workflow/kanban-dispatch.md`).",
         FOREMAN_LINE),
        ("sweep",),
    ]),
    (".claude/loop.md", [
        ("ins_after",
         "`/loop` in this project runs it at a self-paced interval.",
         FOREMAN_LINE),
        ("sweep",),
    ]),
    (".claude/agents/moai/manager-lead.md", [
        ("ins_after", "## Deputy dispatch surface (Role B extension)", DEPUTY_LINE),
        ("ins_after",
         "| Reference | this file (below) | `.claude/rules/moai/workflow/kanban-dispatch.md` |",
         KEPTNAME_LINE),
        ("pair",
         "(canonical label worker-1..worker-N; legacy agent-<n>/lane-<n> still parse as deprecated aliases)",
         "(canonical label `lane-1..lane-N`; `lane-<n>` is the session label, and `-f lane` joins as the next free one)"),
        ("pair",
         "lanes (-f: worker-1..worker-N; legacy agent-<n>/lane-<n> still parses)",
         "lanes (-f: lane-1..lane-N; `lane-<n>` is the session label)"),
        ("pair", "card: {id} | -> worker-{n}",
         "card: {id} | -> lane-{n}  "),
        ("pair", "# Lead Coordinator", "# Leader Coordinator"),
        ("pair", "**Lead-session posture.**", "**Leader-session posture.**"),
        ("sweep",),
    ]),
    (".claude/rules/local/gitflow-lane-protocol.md", [("sweep_ko",)]),
    (".claude/rules/local/repo-local-pr-policy.md", [
        ("pair",
         "the lead collects lane merge SHAs and batch-pushes `origin/develop` (리드 일괄, 2026-09-02)",
         "the factory leader collects lane merge SHAs and batch-pushes `origin/develop` (팩토리 리더 일괄, 2026-09-02)"),
        ("sweep_ko",),
    ]),
]

VERIFY_ONLY = [
    C + "moai-constitution.md",
    C + "moai-mcp-tools-catalogue.md",
    W + "worktree-integration.md",
    ".claude/skills/moai/workflows/factory.md",
    ".claude/agents/harness/hns-release-specialist.md",
    W + "session-handoff-examples.md",
    W + "session-handoff-format.md",
    W + "contract-autonomy.md",
]

PROTECT = [
    ("manager-lead", "\x00MGR\x00"),
    ("team-lead", "\x00TL1\x00"),
    ("team lead", "\x00TL2\x00"),
    ("LEAD-MERGE-APPROVED", "\x00LMA\x00"),
]
LEAD_RE = re.compile(r"\blead\b")


def sweep_lead(text):
    for pat, ph in PROTECT:
        text = text.replace(pat, ph)
    qual_pos = None
    for mm in LEAD_RE.finditer(text):
        pre = text[max(0, mm.start() - 8):mm.start()]
        if pre == "factory " or pre.endswith("-"):
            continue  # already qualified, or hyphenated compound (later occurrence)
        qual_pos = mm.start()
        break
    if qual_pos is not None:
        text = text[:qual_pos] + "factory " + text[qual_pos:]
    text = LEAD_RE.sub("leader", text)
    for pat, ph in PROTECT:
        text = text.replace(ph, pat)
    return text


def sweep_ko(text):
    n = text.count("리드")
    if n == 0:
        return text, n
    i = text.find("리드")
    text = text[:i] + "팩토리 리더" + text[i + 2:]
    return text.replace("리드", "리더"), n


def hard_count(text):
    return len(re.findall(r"\[HARD\]", text))


def apply_ops(text, ops):
    for op in ops:
        if op[0] == "pair":
            old, new = op[1], op[2]
            cnt = text.count(old)
            if cnt != 1:
                raise SystemExit(f"ANCHOR-FAIL pair: hits={cnt} for: {old[:80]}")
            text = text.replace(old, new)
        elif op[0] == "ins_after":
            anchor, ins = op[1], op[2]
            cnt = text.count(anchor)
            if cnt != 1:
                raise SystemExit(f"ANCHOR-FAIL ins: hits={cnt} for: {anchor[:80]}")
            text = text.replace(anchor, anchor + "\n" + ins, 1)
    return text


def line_map(old_lines, new_lines):
    """old line no -> new line no for lines whose content is unchanged."""
    sm = difflib.SequenceMatcher(None, old_lines, new_lines, autojunk=False)
    mapping = {}
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag == "equal":
            for k in range(i2 - i1):
                mapping[i1 + k + 1] = j1 + k + 1
    return mapping


def diff_rows(old_lines, new_lines):
    sm = difflib.SequenceMatcher(None, old_lines, new_lines, autojunk=False)
    rows = []
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag == "equal":
            continue
        for k in range(max(i2 - i1, j2 - j1)):
            o = old_lines[i1 + k] if i1 + k < i2 else ""
            nw = new_lines[j1 + k] if j1 + k < j2 else ""
            rows.append((j1 + k + 1, o, nw))
    return rows


def flatten(s):
    return s.replace("\t", " ").replace("\n", " ")[:240]


def classify(nw):
    if "lane-1..lane-N" in nw:
        return "alias-rewrite"
    if nw in (FOREMAN_LINE, DEPUTY_LINE):
        return "auxiliary-role-line"
    if nw == KEPTNAME_LINE:
        return "kept-name-sentence"
    return "role-noun-substitution"


def transform(text, ops):
    new = apply_ops(text, ops)
    if any(o[0] == "sweep" for o in ops):
        new = sweep_lead(new)
    elif any(o[0] == "sweep_ko" for o in ops):
        new = sweep_ko(new)[0]
    return new


def main():
    dry = "--dry-run" in sys.argv
    ledger_rows = []
    rekeys = []  # (path_in_ledger, old_line, new_line)
    for rel, ops in FILES:
        loc = ROOT / rel
        tpl = loc if rel.startswith(".claude/rules/local/") else ROOT / "internal/template/templates" / rel
        old_tpl = tpl.read_text()
        new_tpl = transform(old_tpl, ops)
        if hard_count(new_tpl) < hard_count(old_tpl):
            raise SystemExit(f"HARD-DECREASE {rel}")
        old_loc = loc.read_text()
        identical = old_tpl == old_loc
        if identical:
            new_loc = new_tpl
        else:
            new_loc = transform(old_loc, ops)
            if hard_count(new_loc) < hard_count(old_loc):
                raise SystemExit(f"HARD-DECREASE local {rel}")
        if dry:
            print(f"OK(dry) {rel}: hard {hard_count(old_tpl)}->{hard_count(new_tpl)}; "
                  f"pair {'IDENT' if identical else 'DIFFER'}")
            continue
        tpl.write_text(new_tpl)
        loc.write_text(new_loc)
        tmap = line_map(old_tpl.split("\n"), new_tpl.split("\n"))
        rekeys += [(rel, o, n) for o, n in tmap.items() if o != n]
        for ln, o, nw in diff_rows(old_tpl.split("\n"), new_tpl.split("\n")):
            ledger_rows.append((rel, ln, flatten(o), flatten(nw), classify(nw)))
        if not identical:
            lmap = line_map(old_loc.split("\n"), new_loc.split("\n"))
            rekeys += [(rel, o, n) for o, n in lmap.items() if o != n]
            for ln, o, nw in diff_rows(old_loc.split("\n"), new_loc.split("\n")):
                ledger_rows.append((rel, ln, flatten(o), flatten(nw), classify(nw)))
        print(f"OK {rel}: hard {hard_count(old_tpl)}->{hard_count(new_tpl)}; "
              f"pair {'IDENT' if identical else 'DIFFER'}")
    if dry:
        return
    for rel in VERIFY_ONLY:
        p = ROOT / rel
        if not p.exists():
            print(f"WARN missing verify-only target {rel}")
            continue
        t = p.read_text()
        for pat, ph in PROTECT:
            t = t.replace(pat, ph)
        hits = len(LEAD_RE.findall(t)) + t.count("리드") + len(
            re.findall(r"worker-1\.\.worker-N|-f worker", t))
        ledger_rows.append((rel, 0, "",
                            f"verify-only this run: role-noun residue hits={hits} (0 expected; identifiers manager-lead/team-lead/leaf-worker excluded)",
                            "verify-only"))
    # re-key surviving rows (content unchanged, shifted by insertions)
    lines = LEDGER.read_text().split("\n")
    keyed = {}
    for path, o, n in rekeys:
        keyed.setdefault(path, {})[o] = n
    out = []
    for row in lines:
        if not row:
            out.append(row)
            continue
        parts = row.split("\t")
        m = keyed.get(parts[0])
        if m and parts[1].isdigit() and int(parts[1]) in m:
            parts[1] = str(m[int(parts[1])])
        out.append("\t".join(parts))
    LEDGER.write_text("\n".join(out))
    with LEDGER.open("a") as f:
        for path, ln, o, nw, cls in ledger_rows:
            f.write(f"{path}\t{ln}\t{o}\t{nw}\t{cls}\n")
    print(f"ledger rows appended: {len(ledger_rows)}; re-keyed rows applied")


if __name__ == "__main__":
    main()
