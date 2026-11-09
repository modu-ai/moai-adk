"""flow_metrics.py fixtures (#2279 scope 6): one counted gh call produces the
baseline file (counts, per-author, ages with a fixed --now), the call ceiling
refuses the call before it starts, and a failing gh exits 1. All through a fake
gh executable — no network."""

import json
import os
import pathlib
import stat
import subprocess
import tempfile
import unittest

TOOL = pathlib.Path(__file__).resolve().parents[1] / "tools/flow_metrics.py"

OPEN_PRS = json.dumps(
    [
        {"number": 2270, "author": {"login": "agent-1"}, "createdAt": "2026-11-08T03:00:00Z", "headRefName": "issue/2270-a"},
        {"number": 2273, "author": {"login": "agent-2"}, "createdAt": "2026-11-08T12:00:00Z", "headRefName": "issue/2273-b"},
        {"number": 2275, "author": {"login": "agent-1"}, "createdAt": "2026-11-09T04:00:00Z", "headRefName": "issue/2275-c"},
    ]
)

NOW = "2026-11-09T06:00:00Z"


def make_gh(temp: pathlib.Path, body: str, fail: bool = False) -> pathlib.Path:
    gh = temp / "gh-fake"
    gh.write_text(
        "#!/bin/sh\n"
        + (f'printf \'%s\' \'{body}\'\nexit 1\n' if fail else f'printf \'%s\' \'{body}\'\n')
    )
    gh.chmod(gh.stat().st_mode | stat.S_IXUSR)
    return gh


def run(gh, out, extra=()):
    return subprocess.run(
        [
            "python3", "-I", str(TOOL),
            "--owner", "modu-ai", "--repo", "moai-adk",
            "--out", str(out), "--gh-bin", str(gh),
            "--now", NOW, *extra,
        ],
        text=True, capture_output=True,
    )


class FlowMetricsTest(unittest.TestCase):
    def test_baseline_written_with_counts_and_ages(self):
        with tempfile.TemporaryDirectory() as d:
            temp = pathlib.Path(d)
            gh = make_gh(temp, OPEN_PRS)
            out = temp / "baseline.json"
            r = run(gh, out)
            self.assertEqual(r.returncode, 0, r.stderr)
            base = json.loads(out.read_text())
            self.assertEqual(base["open_pr_count"], 3)
            self.assertEqual(base["open_issues"], [2270, 2273, 2275])
            self.assertEqual(base["per_author"], {"agent-1": 2, "agent-2": 1})
            # Fixed --now makes ages deterministic: 27h, 18h, 2h
            # (27h = 1.125d; round() halves to even → 1.12).
            self.assertEqual(base["ages_days"]["2270"], 1.12)
            self.assertEqual(base["ages_days"]["2275"], 0.08)
            self.assertEqual(base["max_age_days"], 1.12)
            self.assertEqual(base["gh_calls_used"], 1)
            self.assertEqual(base["gh_call_ceiling"], 4)
            self.assertEqual(
                base["not_run"],
                ["pr_to_first_verdict", "pass_to_merge", "fail_ratio_and_causes", "sync_count"],
            )
            self.assertIn("calls=1/4", r.stdout)

    def test_call_ceiling_zero_refuses_before_call(self):
        with tempfile.TemporaryDirectory() as d:
            temp = pathlib.Path(d)
            gh = make_gh(temp, OPEN_PRS)
            out = temp / "baseline.json"
            r = run(gh, out, extra=["--max-gh-calls", "0"])
            self.assertEqual(r.returncode, 1)
            self.assertIn("ceiling", r.stderr)
            self.assertFalse(out.exists())

    def test_ceiling_one_is_exactly_enough(self):
        with tempfile.TemporaryDirectory() as d:
            temp = pathlib.Path(d)
            gh = make_gh(temp, OPEN_PRS)
            out = temp / "baseline.json"
            r = run(gh, out, extra=["--max-gh-calls", "1"])
            self.assertEqual(r.returncode, 0, r.stderr)
            self.assertEqual(json.loads(out.read_text())["gh_calls_used"], 1)

    def test_failing_gh_exits_1(self):
        with tempfile.TemporaryDirectory() as d:
            temp = pathlib.Path(d)
            gh = make_gh(temp, "[]", fail=True)
            out = temp / "baseline.json"
            r = run(gh, out)
            self.assertEqual(r.returncode, 1)
            self.assertFalse(out.exists())

    def test_usage_exit_2(self):
        r = subprocess.run(["python3", "-I", str(TOOL)], text=True, capture_output=True)
        self.assertEqual(r.returncode, 2)

    def test_tool_makes_exactly_one_call(self):
        # The stub fails hard (exit 7) on any invocation after the first: if the
        # tool ever made a second call, the run would exit 1, not 0.
        with tempfile.TemporaryDirectory() as d:
            temp = pathlib.Path(d)
            marker = temp / "second-call"
            gh = temp / "gh-two"
            gh.write_text(
                "#!/bin/sh\n"
                'if [ -f "%s" ]; then exit 7; fi\n'
                'touch "%s"\n'
                'printf \'%%s\' \'[]\'\n' % (marker, marker)
            )
            gh.chmod(gh.stat().st_mode | stat.S_IXUSR)
            out = temp / "baseline.json"
            r = run(gh, out, extra=["--max-gh-calls", "1"])
            self.assertEqual(r.returncode, 0, r.stderr)
            self.assertEqual(json.loads(out.read_text())["open_pr_count"], 0)


if __name__ == "__main__":
    unittest.main()
