#!/usr/bin/env python3
"""Verify full-output snapshots and failure paths for A4 #2163."""

import copy
import json
import os
import pathlib
import subprocess
import types
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
NOTE = ("Codex /goal의 진행 강제 여부는 미확인 — 효과를 보장하지 않는다"
        "(docs/host-support-matrix.md goal 행, 2026-11-09 확인)")
ISSUES = {"milestone": "rc.2", "issues": [
    {"number": 204, "state": "open"}, {"number": 202, "state": "open"}],
    "prs": [], "keep_set": []}
PRS = {"milestone": "rc.2", "issues": [{"number": 210, "state": "closed"}],
       "prs": [{"number": 2200, "state": "open"}, {"number": 2199, "state": "open"},
               {"number": 2198, "state": "merged"}], "keep_set": []}
LOOP = "/goal 저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(milestone rc.2, HOLD·needs:decision 제외)"
KEPT = {"milestone": "rc.2", "issues": [
    {"number": 2016, "state": "open"}, {"number": 202, "state": "open"}],
    "prs": [{"number": 2199, "state": "open"}, {"number": 2017, "state": "open"}],
    "keep_set": [{"number": 2017, "reason": "운영자 보유"},
                 {"number": 2016, "reason": "운영자 보유"}]}


def rust_bin():
    explicit = os.environ.get("MOAI_CLI_BIN")
    if explicit:
        return pathlib.Path(explicit)
    target = os.environ.get("CARGO_TARGET_DIR")
    return (pathlib.Path(target) if target else ROOT / "target") / "debug" / "moai-cli"


def invoke(fixture):
    """`moai-cli goal-block` on `fixture` (#2708; the Python generator tools/gen_goal_block.py
    is gone, leader decision 6863668555). On success `stdout` is the envelope's `result` as
    JSON, the object the Python generator printed; on failure `stderr` holds the envelope and
    the error line."""
    text = json.dumps(fixture, ensure_ascii=False)
    binary = rust_bin()
    if not binary.exists():
        raise AssertionError(f"moai-cli not built at {binary}; run cargo build first")
    rs = subprocess.run([str(binary), "goal-block"], input=text, capture_output=True, text=True)
    if rs.returncode == 0:
        envelope = json.loads(rs.stdout)
        if envelope.get("ok") is not True:
            raise AssertionError(f"goal-block exit 0 without ok: {rs.stdout!r}")
        return types.SimpleNamespace(returncode=0, stdout=json.dumps(envelope["result"]),
                                     stderr=rs.stderr)
    return types.SimpleNamespace(returncode=rs.returncode, stdout="",
                                 stderr=rs.stdout + rs.stderr)


def run_generator(fixture):
    proc = invoke(fixture)
    if proc.returncode != 0:
        raise AssertionError(f"generator failed: {proc.stderr}")
    return json.loads(proc.stdout)


class GoalBlockSnapshotTest(unittest.TestCase):
    def test_snapshot_issues_only(self):
        self.assertEqual(run_generator(ISSUES), {
            "leader": {
                "goal_line": "/goal milestone rc.2의 열린 Issue가 모두 병합되거나 종료됨",
                "condition": "milestone rc.2의 열린 Issue가 모두 병합되거나 종료됨"},
            "agent": {"loop_line": LOOP, "goal_lines": [
                "/goal #202의 PR이 병합되거나 blocked로 보고됨",
                "/goal #204의 PR이 병합되거나 blocked로 보고됨"]},
            "codex": {"goal_line": "/goal milestone rc.2의 열린 Issue가 모두 병합되거나 종료됨", "note": NOTE},
            "counts": {"open_issues": 2, "open_prs": 0, "keep_set": 0},
            "keep_set_excluded": [], "model_can_input": False, "writes": False,
            "startable": [202, 204], "held": [],
            "end_condition": {"open_prs": 0, "startable": 2, "met": False},
        })

    def test_snapshot_prs_only(self):
        self.assertEqual(run_generator(PRS), {
            "leader": {
                "goal_line": "/goal milestone rc.2의 열린 PR이 모두 병합되거나 종료됨",
                "condition": "milestone rc.2의 열린 PR이 모두 병합되거나 종료됨"},
            "agent": {"loop_line": LOOP, "goal_lines": None},
            "codex": {"goal_line": "/goal milestone rc.2의 열린 PR이 모두 병합되거나 종료됨", "note": NOTE},
            "counts": {"open_issues": 0, "open_prs": 2, "keep_set": 0},
            "keep_set_excluded": [], "model_can_input": False, "writes": False,
            "startable": [], "held": [],
            "end_condition": {"open_prs": 2, "startable": 0, "met": False},
        })

    def test_snapshot_keep_set_included(self):
        self.assertEqual(run_generator(KEPT), {
            "leader": {
                "goal_line": "/goal milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨 — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)",
                "condition": "milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨 — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)"},
            "agent": {"loop_line": LOOP + " — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)",
                      "goal_lines": ["/goal #202의 PR이 병합되거나 blocked로 보고됨"]},
            "codex": {
                "goal_line": "/goal milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨 — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)",
                "note": NOTE},
            "counts": {"open_issues": 1, "open_prs": 1, "keep_set": 2},
            "keep_set_excluded": [{"number": 2016, "reason": "운영자 보유"},
                                  {"number": 2017, "reason": "운영자 보유"}],
            "model_can_input": False, "writes": False,
            "startable": [202],
            "held": [{"issue": 2016, "reason": "keep_set", "keep_reason": "운영자 보유"}],
            "end_condition": {"open_prs": 1, "startable": 1, "met": False},
        })

    def test_no_open_items_does_not_offer_an_already_satisfied_goal(self):
        for fixture in [{"milestone": "rc.2"}, {
            "milestone": "rc.2", "issues": [{"number": 1, "state": "open"}],
            "prs": [{"number": 2, "state": "open"}],
            "keep_set": [{"number": 1, "reason": "보유"}, {"number": 2, "reason": "보유"}],
        }]:
            with self.subTest(fixture=fixture):
                out = run_generator(fixture)
                self.assertIsNone(out["leader"]["goal_line"])
                self.assertIsNone(out["codex"]["goal_line"])
                self.assertIsNone(out["agent"]["goal_lines"])
                self.assertIsNone(out["agent"]["loop_line"])
                self.assertIn("열린 항목 없음", out["leader"]["condition"])
                self.assertEqual(out["counts"]["open_issues"], 0)
                self.assertEqual(out["counts"]["open_prs"], 0)

    def test_the_agent_default_is_the_lane_wide_round_not_one_card(self):
        # #2515: a card-scoped goal ends the session when that card ends; the loop line
        # holds until no open PR and no startable Issue remain, whatever card is held.
        out = run_generator(PRS)
        self.assertIsNone(out["agent"]["goal_lines"])
        self.assertEqual(out["agent"]["loop_line"], LOOP)
        self.assertNotIn("#", out["agent"]["loop_line"].split("—")[0])

    def test_held_cards_alone_meet_the_end_condition(self):
        # #2901 AC5: open Issues that are all on hold or waiting for a decision, and no open
        # PR: the loop's end condition is met, so no loop line is offered.
        out = run_generator({"milestone": "rc.2", "issues": [
            {"number": 2345, "state": "open", "labels": ["needs:decision"]},
            {"number": 2346, "state": "open", "labels": ["HOLD", "P1"]}]})
        self.assertEqual(out.get("end_condition"), {"open_prs": 0, "startable": 0, "met": True})
        self.assertEqual(out.get("startable"), [])
        self.assertEqual(out.get("held"), [
            {"issue": 2345, "reason": "held", "labels": ["needs:decision"]},
            {"issue": 2346, "reason": "held", "labels": ["HOLD"]}])
        self.assertIsNone(out["agent"]["loop_line"])

    def test_an_open_pr_or_a_startable_card_keeps_the_loop(self):
        # #2901 AC5: one open PR, or one startable Issue next to held ones: not met.
        held = {"number": 2345, "state": "open", "labels": ["needs:decision"]}
        for fixture, want in [
            ({"milestone": "rc.2", "issues": [held], "prs": [{"number": 9, "state": "open"}]},
             {"open_prs": 1, "startable": 0, "met": False}),
            ({"milestone": "rc.2", "issues": [held, {"number": 7, "state": "open", "labels": ["threshold"]}]},
             {"open_prs": 0, "startable": 1, "met": False}),
        ]:
            with self.subTest(fixture=fixture):
                out = run_generator(fixture)
                self.assertEqual(out.get("end_condition"), want)
                self.assertEqual(out["agent"]["loop_line"], LOOP)

    def test_each_issue_field_is_judged(self):
        # #2901: title (epic), open sub-issues, a body Blocked-By line naming an open Issue,
        # a blocker GitHub reports; a Blocked-By line naming a closed Issue blocks nothing.
        out = run_generator({"milestone": "rc.2", "issues": [
            {"number": 1, "state": "open", "title": "[rc.2][EP-A] epic"},
            {"number": 2, "state": "open", "open_children": 2},
            {"number": 3, "state": "open", "body": "Blocked-By: #5"},
            {"number": 4, "state": "open", "blocked_by": [5]},
            {"number": 5, "state": "open", "body": "Blocked-By: #6"},
            {"number": 6, "state": "closed"}]})
        self.assertEqual(out.get("held"), [
            {"issue": 1, "reason": "epic", "open_children": 0},
            {"issue": 2, "reason": "epic", "open_children": 2},
            {"issue": 3, "reason": "blocked_by", "blocked_by": [5]},
            {"issue": 4, "reason": "blocked_by", "blocked_by": [5]}])
        self.assertEqual(out.get("startable"), [5])
        for bad in [{"labels": "HOLD"}, {"labels": [1]}, {"title": 3}, {"body": None},
                    {"open_children": -1}, {"blocked_by": [0]}, {"blocked_by": "5"}]:
            with self.subTest(bad=bad):
                proc = invoke({"issues": [dict({"number": 1, "state": "open"}, **bad)]})
                self.assertEqual(proc.returncode, 1)
                self.assertIn("invalid fixture", proc.stderr)

    def test_input_permutations_produce_identical_output(self):
        for fixture in [ISSUES, PRS, KEPT]:
            reversed_fixture = copy.deepcopy(fixture)
            for key in ["issues", "prs", "keep_set"]:
                reversed_fixture[key].reverse()
            self.assertEqual(invoke(fixture).stdout, invoke(reversed_fixture).stdout)

    def test_bad_items_are_rejected_without_traceback(self):
        fixtures = [[], {"milestone": 1}, {"issues": None}, {"issues": [None]}, {"prs": [{}]},
                    {"issues": [{"state": "open"}]}, {"keep_set": [{"number": 1}]},
                    {"keep_set": [{"number": 1, "reason": 3}]}]
        for number in [True, 0, -1, "2016", 1.5]:
            fixtures.append({"issues": [{"number": number, "state": "open"}]})
        for fixture in fixtures:
            with self.subTest(fixture=fixture):
                proc = invoke(fixture)
                self.assertEqual(proc.returncode, 1)
                self.assertEqual(proc.stdout, "")
                self.assertIn("invalid fixture", proc.stderr)
                self.assertNotIn("Traceback", proc.stderr)


if __name__ == "__main__":
    unittest.main()
