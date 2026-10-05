#!/usr/bin/env python3
"""t1256 role-token census over internal/, cmd/, pkg/ Go sources.

Reproducible: run from the repository (or worktree) root. Writes one TSV row per
token occurrence to the path given by --out (default: .moai/reports/t1256/raw/
census.tsv, an ignored path, so the tracked tree is never modified) and prints
per-class counts. Heuristic line-level classifier; each classify() branch below
is the rule a reader can challenge.

Usage: python3 .moai/specs/SPEC-ROLE-NAMING-CODE-001/census.py [--out PATH]
"""
import argparse
import os
import re
import sys
from collections import Counter, defaultdict

ROOTS = ["internal", "cmd", "pkg"]
SKIP_DIR = os.path.join("internal", "template", "templates")  # docs card t1257

# token -> regex. Boundary: not preceded by a lowercase letter (so "subagent",
# "plane", "mislead" never match) unless the token starts uppercase (camelCase).
TOK = {
    "leader": re.compile(r"(?:(?<![A-Za-z])lead|(?<=[a-z])Lead|(?<![A-Za-z])Lead|LEAD)(?:er|ER)s?(?![a-z])"),
    "lead": re.compile(r"(?:(?<![A-Za-z])lead|(?<=[a-z])Lead|(?<![A-Za-z])Lead|(?<![A-Z])LEAD)(?!er|ER)(?:s|ing)?(?![a-z])"),
    "worker": re.compile(r"(?:(?<![A-Za-z])worker|(?<=[a-z0-9])Worker|(?<![A-Za-z])Worker|WORKER)(?:s|S)?(?![a-z])"),
    "agent": re.compile(r"(?:(?<![A-Za-z])agent|(?<=[a-z0-9])Agent|(?<![A-Za-z])Agent|AGENT)(?:s|S|ic)?(?![a-z])"),
    "cjk-lead": re.compile(r"리더|리드(?!미)|リーダー|主导"),
    "cjk-worker": re.compile(r"워커|ワーカー"),
    "cjk-lane": re.compile(r"레인|レーン|泳道"),
    "lane": re.compile(r"(?:(?<![A-Za-z])lane|(?<=[a-z0-9])Lane|(?<![A-Za-z])Lane|LANE)(?:s|S)?(?![a-z])"),
}

# Packages where "worker" is a goroutine-pool word, not a session role.
WORKER_POOL_PKGS = ("internal/hook/mx", "internal/lsp", "internal/mx", "internal/graph",
                    "internal/navigator", "internal/hook/quality", "internal/spec",
                    "internal/profile")
AGENT_ROLE = re.compile(
    r"factory[A-Za-z]*agent|legacy[A-Za-z]*agent|agent-<n>|agent-%d|-f agent|FactoryAgent|"
    r"LegacyAgent|isAgent|\"agent\" \|\||== \"agent\"|agent-\d|lead-agent|agent label|"
    r"MOAI_FACTORY_ROLE", re.I)
LEAD_VERB = re.compile(r"\blead(s|ing)? (to|the way|into)\b|\bleading\b|\blead time\b", re.I)
ENV_KEY = re.compile(r"MOAI_[A-Z_]*(LEAD|WORKER|LANE|AGENT|ROLE)[A-Z_]*|EnvMoai(KanbanLead|FactoryWorker|FactoryRole)")
SQL = re.compile(r"\b(CREATE TABLE|ALTER TABLE|INSERT INTO|INSERT OR|SELECT |UPDATE |DELETE FROM|WHERE |ADD COLUMN|FROM [a-z_]+)")
SPEC_REF = re.compile(r"SPEC-[A-Z0-9-]+")
JSON_TAG = re.compile(r'`[^`]*json:"[^"]*"[^`]*`')
PERSIST_HINT = re.compile(r"workers\.json|lead_pid|lead_process_start|lead_session_id|lane_slot|lane_handoff|"
                          r"lane_dispatch|lane_message|lane_endpoint|\bRoleLead\s*=|\bRoleLane\s*=|FROM workers|"
                          r"INTO workers|\"workers\.json\"|role='lead'|\"lead\", \"lead\"")
SENTINEL = re.compile(r"\b[A-Z][A-Z0-9]*[_-](LEAD|LEADER|WORKER|LANE|AGENT)[A-Z0-9_-]*\b|\b(LEAD|LANE|WORKER|AGENT)[_-][A-Z][A-Z0-9_-]+\b")
STRING_LIT = re.compile(r'"(?:[^"\\]|\\.)*"|`[^`]*`')
COMMENT = re.compile(r"^\s*//")
PERSIST_PKGS = ("internal/homestate", "internal/factorymsg", "internal/kanban")


def classify(path, line, tok, m):
    s = line
    if tok.startswith("cjk-"):
        if tok == "cjk-lead" and "리드" in m.group(0) and not path.startswith(("internal/hook", "internal/cli", "internal/kanban")):
            return "unrelated"
        if tok != "cjk-lane" and not path.startswith(("internal/hook", "internal/cli", "internal/kanban", "internal/web")):
            return "unrelated"
        if COMMENT.match(s):
            return "identifier-internal"
        return "user-facing"
    if tok == "agent" and not AGENT_ROLE.search(s):
        return "unrelated"
    if tok == "worker" and path.startswith(WORKER_POOL_PKGS):
        return "unrelated"
    if tok == "lead":
        word = m.group(0).lower()
        after = s[m.end():m.end() + 12].lower()
        if word.endswith("ing") or after.startswith((" to ", " the way", " into", " time")):
            return "unrelated"
    for am in re.finditer(r"manager-lead|ManagerLead|manager_lead", s):
        if am.start() <= m.start() < am.end():
            return "claude-agent-name"
    for sp in SPEC_REF.finditer(s):
        if sp.start() <= m.start() < sp.end():
            return "unrelated"
    if ENV_KEY.search(s):
        return "env-config-key"
    in_json_tag = any(tok in t.lower() for t in JSON_TAG.findall(s))
    is_comment = bool(COMMENT.match(s))
    if (not is_comment) and (SQL.search(s) or PERSIST_HINT.search(s) or (in_json_tag and path.startswith(PERSIST_PKGS))):
        return "persisted-state"
    if SENTINEL.search(s):
        return "sentinel"
    if not COMMENT.match(s):
        for lit in STRING_LIT.finditer(s):
            if lit.start() <= m.start() < lit.end():
                return "user-facing"
    return "identifier-internal"


def main():
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--out", default=os.path.join(".moai", "reports", "t1256", "raw", "census.tsv"),
                    help="TSV output path (keep it outside the tracked tree)")
    args = ap.parse_args()
    rows = []
    for root in ROOTS:
        for dp, _dn, fn in os.walk(root):
            if dp.startswith(SKIP_DIR):
                continue
            for f in fn:
                if not f.endswith(".go"):
                    continue
                p = os.path.join(dp, f)
                with open(p, encoding="utf-8", errors="replace") as fh:
                    for i, line in enumerate(fh, 1):
                        for tok, rx in TOK.items():
                            for m in rx.finditer(line):
                                cls = classify(p, line, tok, m)
                                rows.append((tok, p, i, cls, "test" if f.endswith("_test.go") else "src",
                                             line.strip()[:200].replace("\t", " ")))
    out_dir = os.path.dirname(args.out)
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)
    with open(args.out, "w") as out:
        out.write("token\tfile\tline\tclass\tkind\ttext\n")
        for r in rows:
            out.write("\t".join(map(str, r)) + "\n")
    c = Counter((r[0], r[3], r[4]) for r in rows)
    classes = ["user-facing", "identifier-internal", "env-config-key", "persisted-state", "sentinel", "claude-agent-name", "unrelated"]
    print("token\tkind\t" + "\t".join(classes) + "\ttotal")
    for tok in TOK:
        for kind in ("src", "test"):
            vals = [c[(tok, k, kind)] for k in classes]
            print(f"{tok}\t{kind}\t" + "\t".join(map(str, vals)) + f"\t{sum(vals)}")
    files = defaultdict(set)
    for r in rows:
        if r[3] != "unrelated":
            files[r[0]].add(r[1])
    print("role-sense files per token:", {k: len(v) for k, v in files.items()})
    print("total rows:", len(rows))


if __name__ == "__main__":
    sys.exit(main())
