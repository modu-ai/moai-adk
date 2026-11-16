"""Tests for tools/candidate_summary.py (#2554 PR-A, candidate verification summary)."""

import importlib.util
import json
import os
import subprocess
import sys
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SCRIPT = os.path.join(REPO_ROOT, "tools", "candidate_summary.py")
SHA = "0123456789abcdef0123456789abcdef01234567"
OTHER = "fedcba9876543210fedcba9876543210fedcba98"
TARGETS = ["linux-x86_64", "linux-arm64", "macos-arm64", "macos-x86_64", "windows-x86_64"]


def load_module():
    spec = importlib.util.spec_from_file_location("candidate_summary", SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def target(sha=SHA, run="4000", passed="4000", skipped="12", result="success"):
    return {
        "result": result,
        "outputs": {
            "sha": sha,
            "tests_run": run,
            "tests_passed": passed,
            "tests_skipped": skipped,
        },
    }


def needs_all_good():
    needs = {t: target() for t in TARGETS}
    needs["pin"] = {"result": "success", "outputs": {"sha": SHA}}
    return needs


class CandidateSummaryTest(unittest.TestCase):
    def setUp(self):
        self.cs = load_module()

    def judge(self, needs, expected=SHA):
        return self.cs.judge(needs, expected)

    def test_five_targets_on_the_pinned_sha_with_tests_run_pass(self):
        ok, text = self.judge(needs_all_good())
        self.assertTrue(ok, text)
        self.assertIn("verdict: PASS", text)
        for t in TARGETS:
            self.assertIn(t, text)
        self.assertEqual(text.count(SHA), 6)  # heading + five rows

    def test_a_missing_target_fails(self):
        for t in TARGETS:
            with self.subTest(missing=t):
                needs = needs_all_good()
                del needs[t]
                ok, text = self.judge(needs)
                self.assertFalse(ok)
                self.assertIn("verdict: FAIL", text)
                self.assertIn(f"{t}: missing", text)

    def test_any_result_other_than_success_fails(self):
        for result in ("failure", "cancelled", "skipped", ""):
            with self.subTest(result=result):
                needs = needs_all_good()
                needs["macos-x86_64"]["result"] = result
                ok, text = self.judge(needs)
                self.assertFalse(ok)
                self.assertIn("macos-x86_64", text)

    def test_a_target_that_checked_out_another_sha_fails(self):
        needs = needs_all_good()
        needs["linux-arm64"]["outputs"]["sha"] = OTHER
        ok, text = self.judge(needs)
        self.assertFalse(ok)
        self.assertIn("linux-arm64", text)
        self.assertIn(OTHER, text)

    def test_zero_or_unreadable_tests_never_pass_windows_included(self):
        for bad in ("0", "", "none", "-1", "1e3", " 4000", "4000 "):
            with self.subTest(tests_run=bad):
                needs = needs_all_good()
                needs["windows-x86_64"]["outputs"]["tests_run"] = bad
                ok, text = self.judge(needs)
                self.assertFalse(ok, text)
                self.assertIn("windows-x86_64", text)
        needs = needs_all_good()
        needs["windows-x86_64"]["outputs"]["tests_run"] = "0"
        needs["windows-x86_64"]["outputs"]["tests_passed"] = "0"
        _, text = self.judge(needs)
        self.assertIn("compiling is not a pass", text)

    def test_passed_must_equal_run(self):
        needs = needs_all_good()
        needs["macos-arm64"]["outputs"]["tests_passed"] = "3999"
        ok, text = self.judge(needs)
        self.assertFalse(ok)
        self.assertIn("macos-arm64", text)

    def test_skipped_is_reported_and_may_be_zero(self):
        needs = needs_all_good()
        needs["linux-x86_64"]["outputs"]["tests_skipped"] = "0"
        needs["windows-x86_64"]["outputs"]["tests_skipped"] = "31"
        ok, text = self.judge(needs)
        self.assertTrue(ok, text)
        self.assertIn("31", text)

    def test_the_pin_job_must_agree_with_the_expected_sha(self):
        needs = needs_all_good()
        needs["pin"]["outputs"]["sha"] = OTHER
        ok, _ = self.judge(needs)
        self.assertFalse(ok)
        needs = needs_all_good()
        del needs["pin"]
        ok, _ = self.judge(needs)
        self.assertFalse(ok)
        # The expected SHA is the run's own, not the pin job's claim.
        ok, _ = self.judge(needs_all_good(), expected=OTHER)
        self.assertFalse(ok)

    def test_a_malformed_expected_sha_fails(self):
        for bad in ("", "abc", SHA.upper(), SHA + "0", "g" * 40):
            with self.subTest(expected=bad):
                needs = needs_all_good()
                ok, _ = self.judge(needs, expected=bad)
                self.assertFalse(ok)

    def test_unknown_shapes_fail_instead_of_raising(self):
        for needs in ({}, {"pin": None}, {t: "x" for t in TARGETS}, {t: {} for t in TARGETS}):
            with self.subTest(needs=needs):
                ok, _ = self.judge(needs)
                self.assertFalse(ok)

    def run_cli(self, needs, github_sha=SHA):
        env = {"PATH": os.environ.get("PATH", ""), "NEEDS": json.dumps(needs)}
        if github_sha is not None:
            env["GITHUB_SHA"] = github_sha
        return subprocess.run(
            [sys.executable, "-I", SCRIPT], env=env, capture_output=True, text=True
        )

    def test_cli_exit_code_follows_the_verdict(self):
        proc = self.run_cli(needs_all_good())
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn("verdict: PASS", proc.stdout)
        needs = needs_all_good()
        needs["windows-x86_64"]["result"] = "failure"
        proc = self.run_cli(needs)
        self.assertEqual(proc.returncode, 1)
        self.assertIn("verdict: FAIL", proc.stdout)

    def test_cli_without_inputs_fails_closed(self):
        proc = subprocess.run(
            [sys.executable, "-I", SCRIPT],
            env={"PATH": os.environ.get("PATH", "")},
            capture_output=True,
            text=True,
        )
        self.assertEqual(proc.returncode, 1)
        self.assertIn("verdict: FAIL", proc.stdout)
        proc = self.run_cli(needs_all_good(), github_sha=None)
        self.assertEqual(proc.returncode, 1)
        env = {"PATH": os.environ.get("PATH", ""), "NEEDS": "not json", "GITHUB_SHA": SHA}
        proc = subprocess.run(
            [sys.executable, "-I", SCRIPT], env=env, capture_output=True, text=True
        )
        self.assertEqual(proc.returncode, 1)


if __name__ == "__main__":
    unittest.main()
