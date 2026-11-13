#!/usr/bin/env python3
"""List the repo-wide scan tests a merge must run first (#2681, leader decision 6848942944).

Some tests read the checked-in tree itself (tools/repo_scans.json). Two branches that touch
disjoint files still fail once merged when one side changes a scan's rule files (the test, the
code that walks, an expected-set file) and the other side changes a path inside that scan's
scope: #2642 added a test that walks src/**/*.rs for copies of a rule, #2632 added a file with
such a copy, and the tip stayed red for 3 h 33 min.

The user decision 'no file overlap, no re-merge' stands. When this judge fires, the gate runs
the listed `checks` on the merge result (`git merge-tree --write-tree <tip> <head>`) before
merging; a red check holds the merge and goes back to the author.

Input, either:

    repo_scans.py --refs <tip> <head>     (git diff --name-only from their merge-base)
    repo_scans.py < {"pr": [...], "tip": [...]}

Output (JSON on stdout):

    {"remerge": bool, "hits": [{"id", "rules_side", "rules", "scope_side", "scope"}],
     "checks": ["command", ...], "reasons": [...]}

`remerge` keeps its name for the gate: true means "run `checks` on the merge result first".
Always exits 0. Any internal error fails safe to `remerge: true` with a reason and every check
in the table, so a broken judge never lets a merge skip the scans (same rule as
tools/ci_scope.py).
"""

import json
import subprocess
import sys
from pathlib import Path

TABLE = Path(__file__).resolve().parent / "repo_scans.json"


def load(path=TABLE):
    return json.loads(Path(path).read_text(encoding="utf-8"))["scans"]


def in_scope(path, scope):
    """Prefix match plus an optional suffix list, no globs (an empty prefix is every path)."""
    for item in scope:
        if not path.startswith(item["prefix"]):
            continue
        suffixes = item.get("suffixes")
        if not suffixes or path.endswith(tuple(suffixes)):
            return True
    return False


def judge(pr, tip, scans):
    """Return (remerge, hits): a hit is a scan whose rule files changed on one side while a
    path in its scope changed on the other side."""
    pr = [p for p in pr if p]
    tip = [p for p in tip if p]
    hits = []
    for scan in scans:
        rules = set(scan["rules"])
        for rules_side, rule_paths, scope_side, scope_paths in (
            ("pr", pr, "tip", tip),
            ("tip", tip, "pr", pr),
        ):
            changed_rules = sorted(p for p in rule_paths if p in rules)
            if not changed_rules:
                continue
            touched = sorted(p for p in scope_paths if in_scope(p, scan["scope"]))
            if touched:
                hits.append({
                    "id": scan["id"],
                    "rules_side": rules_side,
                    "rules": changed_rules,
                    "scope_side": scope_side,
                    "scope": touched,
                })
    return bool(hits), hits


def checks_for(hits, scans):
    """The check commands of every hit scan, in table order, without duplicates."""
    ids = {h["id"] for h in hits}
    out = []
    for scan in scans:
        if scan["id"] in ids:
            for command in scan["checks"]:
                if command not in out:
                    out.append(command)
    return out


def changed(base, ref):
    out = subprocess.run(["git", "diff", "--name-only", "--no-renames", f"{base}..{ref}"],
                         capture_output=True, text=True, check=True).stdout
    return [line for line in out.splitlines() if line]


def from_refs(tip, head):
    base = subprocess.run(["git", "merge-base", tip, head], capture_output=True, text=True,
                          check=True).stdout.strip()
    return changed(base, head), changed(base, tip)


def main(argv):
    scans = []
    try:
        scans = load()
        if argv[:1] == ["--refs"]:
            if len(argv) != 3:
                raise ValueError("usage: repo_scans.py --refs <tip> <head>")
            pr, tip = from_refs(argv[1], argv[2])
        elif argv:
            raise ValueError(f"unknown arguments: {argv}")
        else:
            data = json.load(sys.stdin)
            pr, tip = data["pr"], data["tip"]
            if not isinstance(pr, list) or not isinstance(tip, list):
                raise ValueError("pr and tip must be lists of paths")
        remerge, hits = judge(pr, tip, scans)
        checks = checks_for(hits, scans)
        reasons = [f"{h['id']}: rules changed on {h['rules_side']}, scope changed on {h['scope_side']}"
                   for h in hits]
    except Exception as exc:  # fail safe: any error means run every scan first
        remerge, hits, reasons = True, [], [f"judge internal error: {exc}"]
        checks = checks_for([{"id": s["id"]} for s in scans], scans)
    json.dump({"remerge": remerge, "hits": hits, "checks": checks, "reasons": reasons}, sys.stdout)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
