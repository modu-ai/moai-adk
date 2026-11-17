"""tools/ci_superseded.py (SPEC-20261117-ci-superseded-neutral, card C-83): a post-merge run whose
only non-passing target is a macOS job replaced while still waiting by a newer merge is
`superseded`, so the `ci` aggregate is skipped (neutral) instead of red. Everything else stays
`fail`: a failed target, a cancel with any other reason (timeout, manual), a cancel of a job
that started, any event but push. The recorded run is the survey's I6 example."""

import copy
import importlib.util
import json
import os
import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOL = ROOT / "tools/ci_superseded.py"
RECORDED = ROOT / "tests/fixtures/ci_superseded/run-43728008190.json"
spec = importlib.util.spec_from_file_location("ci_superseded", TOOL)
ci_superseded = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ci_superseded)


def recorded():
    return json.loads(RECORDED.read_text(encoding="utf-8"))


def job(run, name):
    return next(j for j in run["jobs"] if j["name"] == name)


def decide(run):
    return ci_superseded.decide(run["event"], run["check_result"], run["jobs"])[0]


class Decide(unittest.TestCase):
    def test_the_recorded_superseded_run_is_neutral(self):
        self.assertEqual(decide(recorded()), "superseded")

    def test_a_failed_target_beside_the_superseded_one_fails(self):
        run = recorded()
        job(run, "ci / linux-x86_64")["conclusion"] = "failure"
        run["check_result"] = "failure"
        self.assertEqual(decide(run), "fail")
        run["check_result"] = "cancelled"
        self.assertEqual(decide(run), "fail")

    def test_a_cancel_for_any_other_reason_fails(self):
        for message in [
            "The job running on runner GitHub Actions 1 has exceeded the maximum execution time of 45 minutes.",
            "The operation was canceled.",
            "Canceling since a higher priority waiting request for ci-push-macos-arm64 exists",
            "Canceling since a higher priority waiting request for ci-push-macos-x86_64 exists!",
        ]:
            with self.subTest(message=message):
                run = recorded()
                job(run, "ci / macos-x86_64")["annotations"] = [message]
                self.assertEqual(decide(run), "fail")
        run = recorded()
        job(run, "ci / macos-x86_64")["annotations"] = []
        self.assertEqual(decide(run), "fail")
        # The message as a bare string (not a list) is not read as a substring match.
        run = recorded()
        job(run, "ci / macos-x86_64")["annotations"] = (
            "Canceling since a higher priority waiting request for ci-push-macos-x86_64 exists")
        self.assertEqual(decide(run), "fail")

    def test_a_cancelled_job_that_started_fails(self):
        run = recorded()
        job(run, "ci / macos-x86_64")["steps"] = [{"name": "Set up job", "conclusion": "success"}]
        self.assertEqual(decide(run), "fail")
        run = recorded()
        job(run, "ci / macos-x86_64")["runner_name"] = "GitHub Actions 7"
        self.assertEqual(decide(run), "fail")

    def test_only_a_macos_target_can_be_superseded(self):
        run = recorded()
        linux = job(run, "ci / linux-x86_64")
        linux.update(conclusion="cancelled", runner_name="", steps=[],
                     annotations=["Canceling since a higher priority waiting request for ci-push-linux-x86_64 exists"])
        self.assertEqual(decide(run), "fail")

    def test_only_a_push_run_can_be_superseded(self):
        for event in ["pull_request", "workflow_dispatch", "merge_group", ""]:
            with self.subTest(event=event):
                run = recorded()
                run["event"] = event
                self.assertEqual(decide(run), "fail")

    def test_the_aggregate_result_must_be_cancelled(self):
        for result in ["success", "failure", "skipped", ""]:
            with self.subTest(result=result):
                run = recorded()
                run["check_result"] = result
                self.assertEqual(decide(run), "fail")

    def test_no_target_or_no_superseded_target_fails(self):
        run = recorded()
        run["jobs"] = [j for j in run["jobs"] if not j["name"].startswith("ci / ")]
        self.assertEqual(decide(run), "fail")
        run = recorded()
        job(run, "ci / macos-x86_64").update(conclusion="success")
        self.assertEqual(decide(run), "fail")

    def test_unreadable_jobs_fail_instead_of_raising(self):
        for jobs in [None, "x", [None], [{"name": 3}], [{"name": "ci / macos-arm64"}]]:
            with self.subTest(jobs=jobs):
                self.assertEqual(ci_superseded.decide("push", "cancelled", jobs)[0], "fail")


class CommandLine(unittest.TestCase):
    def run_tool(self, run):
        with tempfile.TemporaryDirectory() as tmp:
            jobs = pathlib.Path(tmp) / "jobs.json"
            jobs.write_text(json.dumps(run["jobs"]), encoding="utf-8")
            output = pathlib.Path(tmp) / "output"
            env = {k: v for k, v in os.environ.items() if not k.startswith(("GITHUB_", "GH_"))}
            env.update(EVENT=run["event"], CHECK_RESULT=run["check_result"], GITHUB_OUTPUT=str(output))
            out = subprocess.run([sys.executable, "-I", str(TOOL), "--jobs-file", str(jobs)],
                                 capture_output=True, text=True, env=env, check=False)
            return out, output.read_text(encoding="utf-8") if output.exists() else ""

    def test_the_state_line_goes_to_the_step_output(self):
        out, written = self.run_tool(recorded())
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertEqual(written, "state=superseded\n")
        self.assertIn("ci / macos-x86_64", out.stdout)
        run = recorded()
        job(run, "ci / linux-x86_64")["conclusion"] = "failure"
        out, written = self.run_tool(run)
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertEqual(written, "state=fail\n")


if __name__ == "__main__":
    unittest.main()
