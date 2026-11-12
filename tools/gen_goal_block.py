#!/usr/bin/env python3
"""Deterministic /goal paste-block generator (A4 #2163, REQ-A4-01/02).

Reads a JSON fixture from stdin and prints the paste-ready block as JSON:

    {"milestone": "...", "issues": [{"number": N, "state": "open"}, ...],
     "prs": [...], "keep_set": [{"number": N, "reason": "..."}]}

Output kinds:
  leader  — milestone-wide completion condition (open issues+prs merged or
            closed; keep-set entries excluded with their reasons)
  agent   — loop_line: the lane-wide round condition (no open PR and no
            startable Issue, keep-set exclusions with reasons; the default goal
            of agent mode, #2515), and goal_lines: card-scoped conditions for
            the open issues (PR merged or reported blocked), used only when the
            user holds one card
  codex   — the same one-liner with the Codex caveat (effect unverified,
            docs/host-support-matrix.md goal rows)

The model cannot type `/goal` itself (autonomy-modes-spec.md §7, support table goal
rows); the only supported path is the user pasting this block. Stdlib only;
output is deterministic (sorted by number, stable wording).
"""

import json
import sys


def build(fixture: dict) -> dict:
    if not isinstance(fixture, dict):
        raise ValueError("expected an object")
    milestone = fixture.get("milestone", "")
    if not isinstance(milestone, str):
        raise ValueError("milestone must be a string")
    for group in ("issues", "prs", "keep_set"):
        items = fixture.get(group, [])
        if not isinstance(items, list):
            raise ValueError(f"{group} must be a list")
        for item in items:
            if not isinstance(item, dict):
                raise ValueError(f"{group} entries must be objects")
            if type(item.get("number")) is not int or item["number"] <= 0:
                raise ValueError(f"{group} entries require a positive integer number")
            if group == "keep_set" and not isinstance(item.get("reason"), str):
                raise ValueError("keep_set entries require a string reason")

    keep_set = sorted(fixture.get("keep_set", []), key=lambda k: k["number"])
    excluded = {k["number"] for k in keep_set}
    issues = sorted(
        (i for i in fixture.get("issues", [])
         if i.get("state") == "open" and i["number"] not in excluded),
        key=lambda i: i["number"],
    )
    prs = sorted(
        (p for p in fixture.get("prs", [])
         if p.get("state") == "open" and p["number"] not in excluded),
        key=lambda p: p["number"],
    )

    subject = "Issue·PR"
    if issues and not prs:
        subject = "Issue"
    elif prs and not issues:
        subject = "PR"

    if not issues and not prs:
        condition = f"milestone {milestone}의 열린 항목 없음"
    else:
        particle = "가" if subject == "Issue" else "이"
        condition = f"milestone {milestone}의 열린 {subject}{particle} 모두 병합되거나 종료됨"
    if keep_set:
        exclusions = ", ".join(f"#{k['number']}({k['reason']})" for k in keep_set)
        condition += f" — 다음은 제외: {exclusions}"
    leader_line = f"/goal {condition}" if issues or prs else None

    loop_line = None
    if issues or prs:
        loop_line = (f"/goal 저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임"
                     f"(milestone {milestone}, HOLD·needs:decision 제외)")
        if keep_set:
            loop_line += f" — 다음은 제외: {exclusions}"
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
        "agent": {"loop_line": loop_line, "goal_lines": agent_lines},
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
    try:
        block = build(fixture)
    except ValueError as exc:
        print(f"invalid fixture: {exc}", file=sys.stderr)
        return 1
    print(json.dumps(block, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
