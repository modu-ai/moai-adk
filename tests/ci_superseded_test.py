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


FAKE_GH = """#!{python}
import json, os, sys
with open(os.environ["FAKE_GH_LOG"], "a", encoding="utf-8") as log:
    log.write(json.dumps(sys.argv[1:]) + "\\n")
answers = json.loads(open(os.environ["FAKE_GH_ANSWERS"], encoding="utf-8").read())
path = sys.argv[2] if sys.argv[1:2] == ["api"] and len(sys.argv) == 3 else None
if path not in answers:
    print("fake gh: no answer for " + repr(sys.argv[1:]), file=sys.stderr)
    sys.exit(1)
print(json.dumps(answers[path]))
"""


class Fetch(unittest.TestCase):
    """C-83 audit N1: in CI the jobs come from `gh api`. The read is scoped to the run's own
    attempt, asks for up to 100 jobs, and fetches annotations only for cancelled targets."""

    def run_fetch(self, answers):
        with tempfile.TemporaryDirectory() as tmp:
            tmp = pathlib.Path(tmp)
            gh = tmp / "bin" / "gh"
            gh.parent.mkdir()
            gh.write_text(FAKE_GH.format(python=sys.executable), encoding="utf-8")
            gh.chmod(0o755)
            (tmp / "answers.json").write_text(json.dumps(answers), encoding="utf-8")
            env = {k: v for k, v in os.environ.items() if not k.startswith(("GITHUB_", "GH_"))}
            env.update(PATH=f"{gh.parent}{os.pathsep}{env.get('PATH', '')}",
                       FAKE_GH_LOG=str(tmp / "log"), FAKE_GH_ANSWERS=str(tmp / "answers.json"),
                       GITHUB_REPOSITORY="owner/repo", GITHUB_RUN_ID="111", GITHUB_RUN_ATTEMPT="2",
                       GITHUB_OUTPUT=str(tmp / "output"), EVENT="push", CHECK_RESULT="cancelled")
            out = subprocess.run([sys.executable, "-I", str(TOOL)], capture_output=True, text=True,
                                 env=env, check=False)
            calls = [json.loads(line) for line in (tmp / "log").read_text().splitlines()] \
                if (tmp / "log").exists() else []
            written = (tmp / "output").read_text() if (tmp / "output").exists() else ""
            return out, calls, written

    def answers(self):
        run = recorded()
        jobs = [{k: v for k, v in j.items() if k != "annotations"} for j in run["jobs"]]
        answers = {"repos/owner/repo/actions/runs/111/attempts/2/jobs?per_page=100": {"jobs": jobs}}
        for j in run["jobs"]:
            if j["annotations"]:
                answers[f"repos/owner/repo/check-runs/{j['id']}/annotations"] = [
                    {"message": m, "annotation_level": "failure"} for m in j["annotations"]]
        return answers

    def test_the_read_is_attempt_scoped_and_annotations_are_read_only_for_cancelled_targets(self):
        out, calls, written = self.run_fetch(self.answers())
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertEqual(calls, [
            ["api", "repos/owner/repo/actions/runs/111/attempts/2/jobs?per_page=100"],
            ["api", "repos/owner/repo/check-runs/132854068092/annotations"],
        ])
        self.assertEqual(written, "state=superseded\n")

    def test_a_failed_read_writes_no_state_and_exits_2(self):
        answers = self.answers()
        del answers["repos/owner/repo/check-runs/132854068092/annotations"]
        out, _calls, written = self.run_fetch(answers)
        self.assertEqual(out.returncode, 2)
        self.assertEqual(written, "")


if __name__ == "__main__":
    unittest.main()
