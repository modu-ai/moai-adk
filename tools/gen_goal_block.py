#!/usr/bin/env python3
"""Deterministic /goal paste-block generator (A4 #2163, REQ-A4-01/02).

Reads a JSON fixture from stdin and prints the paste-ready block as JSON:

    {"milestone": "...", "issues": [{"number": N, "state": "open"}, ...],
     "prs": [...], "keep_set": [{"number": N, "reason": "..."}]}

Output kinds:
  leader  — milestone-wide completion condition (open issues+prs merged or
            closed; keep-set entries excluded with their reasons)
  agent   — card-scoped condition for the assigned issue (PR merged or
            reported blocked)
  codex   — the same one-liner with the Codex caveat (effect unverified,
            docs/host-support-matrix.md goal rows)

The model cannot type `/goal` itself (autonomy-modes-spec.md §7, 지원표 goal
rows); the only supported path is the user pasting this block. Stdlib only;
output is deterministic (sorted by number, stable wording).
"""

import json
import sys


def build(fixture: dict) -> dict:
    milestone = fixture.get("milestone", "")
    issues = sorted(
        (i for i in fixture.get("issues", []) if i.get("state") == "open"),
        key=lambda i: i["number"],
    )
    prs = sorted(
        (p for p in fixture.get("prs", []) if p.get("state") == "open"),
        key=lambda p: p["number"],
    )
    keep_set = sorted(fixture.get("keep_set", []), key=lambda k: k["number"])

    subject = "Issue·PR"
    if issues and not prs:
        subject = "Issue"
    elif prs and not issues:
        subject = "PR"

    open_counts = f"열린 {subject}"
    condition = f"milestone {milestone}의 {open_counts}이 모두 병합되거나 종료됨"
    if keep_set:
        exclusions = ", ".join(f"#{k['number']}({k['reason']})" for k in keep_set)
        condition += f" — 다음은 제외: {exclusions}"
    leader_line = f"/goal {condition}"

    agent_lines = []
    for i in issues:
        agent_lines.append(f"/goal #{i['number']}의 PR이 병합되거나 blocked로 보고됨")
    if not agent_lines:
        agent_lines = None

    codex_line = leader_line
    codex_note = (
        "Codex /goal의 진행 강제 여부는 미확인 — 효과를 보장하지 않는다"
        "(docs/host-support-matrix.md goal 행, 2026-11-09 확인)"
    )

    return {
        "leader": {"goal_line": leader_line, "condition": condition},
        "agent": {"goal_lines": agent_lines},
        "codex": {"goal_line": codex_line, "note": codex_note},
        "counts": {
            "open_issues": len(issues),
            "open_prs": len(prs),
            "keep_set": len(keep_set),
        },
        "keep_set_excluded": [
            {"number": k["number"], "reason": k["reason"]} for k in keep_set
        ],
        "model_can_input": False,
        "writes": False,
    }


def main(argv: list) -> int:
    if len(argv) != 1:
        print(__doc__, file=sys.stderr)
        return 2
    try:
        fixture = json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"invalid fixture JSON: {exc}", file=sys.stderr)
        return 1
    print(json.dumps(build(fixture), ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
