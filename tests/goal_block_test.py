#!/usr/bin/env python3
"""Snapshot fixtures for the /goal paste-block generator (A4 #2163).

Three fixtures from the acceptance criteria — open issues only, open PRs
only, keep-set included — pin the generator's full output. Any wording or
ordering change must update these snapshots explicitly in the same commit.
"""

import json
import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
GENERATOR = ROOT / "tools" / "gen_goal_block.py"


def run_generator(fixture: dict) -> dict:
    proc = subprocess.run(
        ["python3", "-I", str(GENERATOR)],
        input=json.dumps(fixture, ensure_ascii=False),
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        raise AssertionError(f"generator failed: {proc.stderr}")
    return json.loads(proc.stdout)


class GoalBlockSnapshotTest(unittest.TestCase):
    def test_snapshot_issues_only(self):
        fixture = {
            "milestone": "rc.2",
            "issues": [
                {"number": 204, "state": "open"},
                {"number": 202, "state": "open"},
            ],
            "prs": [],
            "keep_set": [],
        }
        out = run_generator(fixture)
        self.assertEqual(
            out,
            {
                "leader": {
                    "goal_line": "/goal milestone rc.2의 열린 Issue이 모두 병합되거나 종료됨",
                    "condition": "milestone rc.2의 열린 Issue이 모두 병합되거나 종료됨",
                },
                "agent": {
                    "goal_lines": [
                        "/goal #202의 PR이 병합되거나 blocked로 보고됨",
                        "/goal #204의 PR이 병합되거나 blocked로 보고됨",
                    ]
                },
                "codex": {
                    "goal_line": "/goal milestone rc.2의 열린 Issue이 모두 병합되거나 종료됨",
                    "note": (
                        "Codex /goal의 진행 강제 여부는 미확인 — 효과를 보장하지 않는다"
                        "(docs/host-support-matrix.md goal 행, 2026-11-09 확인)"
                    ),
                },
                "counts": {"open_issues": 2, "open_prs": 0, "keep_set": 0},
                "keep_set_excluded": [],
                "model_can_input": False,
                "writes": False,
            },
        )

    def test_snapshot_prs_only(self):
        fixture = {
            "milestone": "rc.2",
            "issues": [
                {"number": 210, "state": "closed"},
                {"number": 205, "state": "open"},
            ],
            "prs": [{"number": 2199, "state": "open"}],
            "keep_set": [],
        }
        out = run_generator(fixture)
        # closed issues are not counted; subject becomes PR-only
        self.assertEqual(out["counts"], {"open_issues": 1, "open_prs": 1, "keep_set": 0})
        self.assertEqual(
            out["leader"]["goal_line"],
            "/goal milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨",
        )
        self.assertEqual(out["keep_set_excluded"], [])

    def test_snapshot_keep_set_included(self):
        fixture = {
            "milestone": "rc.2",
            "issues": [{"number": 202, "state": "open"}],
            "prs": [{"number": 2199, "state": "open"}],
            "keep_set": [
                {"number": 2016, "reason": "운영자 보유"},
                {"number": 2017, "reason": "운영자 보유"},
            ],
        }
        out = run_generator(fixture)
        self.assertEqual(
            out["leader"]["condition"],
            (
                "milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨"
                " — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)"
            ),
        )
        self.assertEqual(out["leader"]["goal_line"], "/goal milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨 — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)")
        self.assertEqual(out["keep_set_excluded"], [{"number": 2016, "reason": "운영자 보유"}, {"number": 2017, "reason": "운영자 보유"}])
        self.assertFalse(out["model_can_input"])
        self.assertFalse(out["writes"])


if __name__ == "__main__":
    unittest.main()
