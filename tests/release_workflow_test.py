"""SPEC-20261119-headver-release-workflow (card C-93): the release workflow numbers, names and
checks a HeadVer release, and never publishes one.

Everything here is a fixture: the `run:` scripts of `.github/workflows/release.yml` are taken out
of the YAML and run in a temporary git repository (tags made there), with a fake binary for the
version check. No network, no real tag, no release, no real HOME state.
"""
import datetime
import os
import pathlib
import re
import subprocess
import tempfile
import unittest

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
RELEASE = ROOT / ".github/workflows/release.yml"
CI = ROOT / ".github/workflows/ci.yml"
SHA = "0123456789abcdef0123456789abcdef01234567"
OTHER_SHA = "fedcba9876543210fedcba9876543210fedcba98"
PINNED_USES = re.compile(r"^[\w.-]+/[\w./-]+@[0-9a-f]{40}$")
# What a release workflow of this card must never run: a tag, a release, an upload, a publish.
FORBIDDEN = re.compile(
    r"gh release|git tag|git push|upload-artifact|download-artifact|attest|cargo publish|"
    r"npm publish|docker push|gh api|gh pr|gh issue", re.I)


def epoch(year, month, day):
    """Noon UTC of a date, as the seconds `SOURCE_DATE_EPOCH` carries."""
    moment = datetime.datetime(year, month, day, 12, tzinfo=datetime.timezone.utc)
    return str(int(moment.timestamp()))


def git_env():
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env.update(GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_SYSTEM=os.devnull, GIT_CONFIG_NOSYSTEM="1",
               GIT_AUTHOR_NAME="t", GIT_AUTHOR_EMAIL="t@example.invalid",
               GIT_COMMITTER_NAME="t", GIT_COMMITTER_EMAIL="t@example.invalid", LC_ALL="C")
    return env


def flow():
    return yaml.safe_load(RELEASE.read_text(encoding="utf-8"))


def script(job, step_id):
    """The `run:` text of one step, as GitHub hands it to the shell."""
    (step,) = [s for s in flow()["jobs"][job]["steps"] if s.get("id") == step_id]
    return step["run"]


def bash(text, cwd, env):
    """What the runner does with a `run:` step under `shell: bash`."""
    return subprocess.run(["bash", "--noprofile", "--norc", "-eo", "pipefail", "-c", text],
                          cwd=cwd, env=env, capture_output=True, text=True, timeout=60)


class WorkflowShapeTest(unittest.TestCase):
    def test_the_workflow_starts_only_from_a_push_to_a_release_branch(self):
        """AC-1, AC-4: a person pushes `release/<sha>` or `release/rc/<sha>`; nothing else
        starts it, and a newer run never cancels an older one."""
        f = flow()
        self.assertEqual(f.get("on", f.get(True)), {"push": {"branches": ["release/**"]}})
        self.assertNotIn("concurrency", f)
        for job in f["jobs"].values():
            self.assertNotIn("concurrency", job)

    def test_the_token_is_read_only_and_no_secret_is_read(self):
        """AC-4: no write scope, no secret, no id-token."""
        text = RELEASE.read_text(encoding="utf-8")
        f = flow()
        self.assertEqual(f["permissions"], {"contents": "read"})
        for job_id, job in f["jobs"].items():
            with self.subTest(job=job_id):
                self.assertEqual(job["permissions"], {"contents": "read"})
                self.assertNotIn("secrets", job)
        self.assertIsNone(re.search(r"\$\{\{\s*secrets\.|secrets:\s*inherit|GITHUB_TOKEN|id-token", text))

    def test_no_step_tags_publishes_uploads_or_calls_the_api(self):
        """AC-4: the workflow only plans, builds and reports; the tag and the release are the
        person's own steps."""
        for job_id, job in flow()["jobs"].items():
            for step in job["steps"]:
                with self.subTest(job=job_id, step=step.get("name", step.get("uses"))):
                    self.assertIsNone(FORBIDDEN.search(step.get("run", "") + step.get("uses", "")))

    def test_values_reach_scripts_through_env_only(self):
        """A branch name is outside text: no `${{ }}` inside a `run:` script."""
        for job in flow()["jobs"].values():
            for step in job["steps"]:
                with self.subTest(step=step.get("name", step.get("uses"))):
                    self.assertNotIn("${{", step.get("run", ""))

    def test_actions_are_pinned_to_the_commit_ci_uses(self):
        ci_uses = {s["uses"] for j in yaml.safe_load(CI.read_text(encoding="utf-8"))["jobs"].values()
                   for s in j.get("steps", []) if "uses" in s}
        uses = [s["uses"] for j in flow()["jobs"].values() for s in j["steps"] if "uses" in s]
        self.assertTrue(uses)
        for use in uses:
            with self.subTest(uses=use):
                self.assertRegex(use, PINNED_USES)
                self.assertIn(use, ci_uses)

    def test_every_checkout_keeps_no_credential_and_the_plan_reads_the_tags(self):
        f = flow()
        for job_id, job in f["jobs"].items():
            (checkout,) = [s for s in job["steps"] if s.get("uses", "").startswith("actions/checkout@")]
            with self.subTest(job=job_id):
                self.assertIs(checkout["with"]["persist-credentials"], False)
        (plan_checkout,) = [s for s in f["jobs"]["plan"]["steps"] if "uses" in s]
        self.assertIs(plan_checkout["with"]["fetch-tags"], True)
        (build_checkout,) = [s for s in f["jobs"]["build"]["steps"]
                             if s.get("uses", "").startswith("actions/checkout@")]
        self.assertEqual(build_checkout["with"]["ref"], "${{ needs.plan.outputs.sha }}")

    def test_the_build_is_numbered_by_the_plan_and_never_a_local_build(self):
        """AC-3 (D4): a release build is a `CI build`: it gets `MOAI_BUILD_NUMBER` from the plan
        (and the plan's clock reading), and `MOAI_LOCAL_BUILD` is set nowhere."""
        f = flow()
        self.assertEqual(f["jobs"]["build"]["needs"], "plan")
        env = f["jobs"]["build"]["env"]
        self.assertEqual(env["MOAI_BUILD_NUMBER"], "${{ needs.plan.outputs.build }}")
        self.assertEqual(env["SOURCE_DATE_EPOCH"], "${{ needs.plan.outputs.epoch }}")
        for job in f["jobs"].values():
            self.assertNotIn("MOAI_LOCAL_BUILD", job.get("env", {}))
            for step in job["steps"]:
                self.assertNotIn("MOAI_LOCAL_BUILD", step.get("env", {}))
                self.assertNotIn("MOAI_LOCAL_BUILD", step.get("run", ""))
        self.assertNotIn("MOAI_LOCAL_BUILD", f.get("env", {}))

    def test_scripts_run_under_bash(self):
        f = flow()
        for job_id, job in f["jobs"].items():
            with self.subTest(job=job_id):
                self.assertEqual(job["defaults"]["run"]["shell"], "bash")


class PlanTest(unittest.TestCase):
    """AC-1, AC-2: the `plan` step on fixture repositories."""

    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        self.repo = pathlib.Path(self._tmp.name) / "repo"
        self.repo.mkdir()
        self.out = pathlib.Path(self._tmp.name) / "github_output"
        self.git("init", "-q")
        (self.repo / "HEADVER").write_text("4\n", encoding="utf-8")
        self.git("add", "HEADVER")
        self.git("commit", "-q", "-m", "x")

    def git(self, *args):
        subprocess.run(["git", *args], cwd=self.repo, env=git_env(), check=True,
                       capture_output=True, text=True)

    def plan(self, ref=f"release/{SHA}", pushed=SHA, run="7", attempt="1", day=(2026, 11, 24),
             clock=False):
        self.out.write_text("", encoding="utf-8")
        env = git_env()
        env.update(REF_NAME=ref, PUSHED=pushed, RUN_NUMBER=run, RUN_ATTEMPT=attempt,
                   GITHUB_OUTPUT=str(self.out))
        env.pop("SOURCE_DATE_EPOCH", None)
        if not clock:
            env["SOURCE_DATE_EPOCH"] = epoch(*day)
        done = bash(script("plan", "plan"), self.repo, env)
        outputs = dict(line.split("=", 1) for line in self.out.read_text(encoding="utf-8").splitlines())
        return done, outputs

    def refused(self, done, outputs, words):
        self.assertNotEqual(done.returncode, 0, done.stdout + done.stderr)
        for word in words:
            self.assertIn(word, done.stderr)
        self.assertEqual(outputs, {}, "a refused plan hands nothing on")

    def test_a_release_branch_is_numbered_from_its_head_week_and_run_number(self):
        done, out = self.plan()
        self.assertEqual(done.returncode, 0, done.stderr)
        self.assertEqual(out, {"sha": SHA, "head": "4", "yearweek": "2648", "build": "7",
                               "version": "4.2648.7", "tag": "moai-adk-v4.2648.7",
                               "prerelease": "false", "epoch": epoch(2026, 11, 24)})

    def test_a_release_candidate_branch_is_a_pre_release_with_the_same_tag_shape(self):
        """AC-2: no suffix on the tag; the pre-release mark is what says it is a candidate."""
        done, out = self.plan(ref=f"release/rc/{SHA}")
        self.assertEqual(done.returncode, 0, done.stderr)
        self.assertEqual(out["tag"], "moai-adk-v4.2648.7")
        self.assertEqual(out["prerelease"], "true")

    def test_head_comes_from_the_tracked_file(self):
        (self.repo / "HEADVER").write_text("12\n", encoding="utf-8")
        done, out = self.plan()
        self.assertEqual(done.returncode, 0, done.stderr)
        self.assertEqual((out["version"], out["tag"]), ("12.2648.7", "moai-adk-v12.2648.7"))

    def test_the_week_is_the_iso_week_of_the_utc_day_at_the_year_boundary(self):
        for day, week in [((2026, 12, 31), "2653"), ((2027, 1, 1), "2653"), ((2027, 1, 3), "2653"),
                          ((2027, 1, 4), "2701"), ((2026, 11, 23), "2648"), ((2026, 11, 22), "2647")]:
            with self.subTest(day=day):
                done, out = self.plan(day=day)
                self.assertEqual(done.returncode, 0, done.stderr)
                self.assertEqual(out["yearweek"], week)

    def test_without_a_given_epoch_the_clock_is_read_once(self):
        before = int(datetime.datetime.now(datetime.timezone.utc).timestamp())
        done, out = self.plan(clock=True)
        after = int(datetime.datetime.now(datetime.timezone.utc).timestamp())
        self.assertEqual(done.returncode, 0, done.stderr)
        self.assertTrue(before <= int(out["epoch"]) <= after, out)

    def test_the_branch_name_must_name_the_pushed_commit(self):
        for ref in [f"release/{OTHER_SHA}", f"release/rc/{OTHER_SHA}"]:
            with self.subTest(ref=ref):
                done, out = self.plan(ref=ref)
                self.refused(done, out, ["names one commit", "the push is another"])

    def test_only_two_branch_shapes_are_accepted(self):
        bad = ["release/v1", f"release/{SHA[:39]}", f"release/{SHA}0", f"release/{SHA.upper()}",
               f"release/RC/{SHA}", f"release/rc/{SHA}/x", f"candidate/{SHA}", f"releases/{SHA}",
               f"release/rc/rc/{SHA}", "release/", "main"]
        for ref in bad:
            with self.subTest(ref=ref):
                done, out = self.plan(ref=ref)
                self.refused(done, out, ["release/<40 lowercase hex"])

    def test_a_rerun_is_refused_because_it_keeps_the_run_number(self):
        """AC-1: a re-run keeps `run_number`, so it would repeat a build number."""
        for attempt in ["2", "3", "0", "", "01", "x"]:
            with self.subTest(attempt=attempt):
                done, out = self.plan(attempt=attempt)
                self.refused(done, out, ["first attempt", "new release branch"])

    def test_the_run_number_has_the_bounds_of_the_build_variable(self):
        """The same values `MOAI_BUILD_NUMBER` takes (src/local_build.rs `number_of`)."""
        for good in ["1", "7", "99", "4294967295"]:
            with self.subTest(good=good):
                done, out = self.plan(run=good)
                self.assertEqual(done.returncode, 0, done.stderr)
                self.assertEqual(out["build"], good)
        for bad in ["0", "01", "-1", "+1", "x", "", "1.5", "4294967296", " 7"]:
            with self.subTest(bad=bad):
                done, out = self.plan(run=bad)
                self.refused(done, out, ["RUN_NUMBER"])

    def test_the_head_file_has_the_bounds_of_the_rust_reader(self):
        """The same rule as `headver::parse_head`: 0 to 9999, no sign, no leading zero."""
        for good, version in [("0\n", "0.2648.7"), ("4", "4.2648.7"), ("9999\n", "9999.2648.7")]:
            with self.subTest(good=good):
                (self.repo / "HEADVER").write_text(good, encoding="utf-8")
                done, out = self.plan()
                self.assertEqual(done.returncode, 0, done.stderr)
                self.assertEqual(out["version"], version)
        for bad in ["", "\n", "04\n", "10000\n", "-1\n", "4 5\n", "x\n", "4\n\n", "4.0\n"]:
            with self.subTest(bad=bad):
                (self.repo / "HEADVER").write_text(bad, encoding="utf-8")
                done, out = self.plan()
                self.refused(done, out, ["HEADVER"])

    def test_an_existing_tag_of_this_version_is_never_reused(self):
        """AC-1: two releases never share a version."""
        self.git("tag", "moai-adk-v4.2648.7")
        done, out = self.plan()
        self.refused(done, out, ["moai-adk-v4.2648.7", "already exists"])

    def test_the_build_number_only_goes_up_across_weeks_and_heads(self):
        """AC-1: a counter that restarted (the workflow file was recreated) must not repeat or
        fall behind a build number that already has a tag, in any week or head."""
        for tag in ["moai-adk-v4.2640.7", "moai-adk-v4.2640.9", "moai-adk-v3.2301.8",
                    "moai-adk-v5.2650.100"]:
            with self.subTest(existing=tag):
                self.git("tag", tag)
                done, out = self.plan()
                self.refused(done, out, ["moai-adk-v", "not above"])
                self.git("tag", "-d", tag)

    def test_lower_build_numbers_and_other_tag_families_do_not_block(self):
        for tag in ["moai-adk-v4.2640.6", "moai-adk-v4.2647.1", "v3.1.3", "go-v1.2.3",
                    "moai-cli-v2026.11.0-rc.1", "moai-adk-v4.2640.99-dev", "moai-adk-vx.y.z",
                    "xmoai-adk-v4.2640.50"]:
            self.git("tag", tag)
        done, out = self.plan()
        self.assertEqual(done.returncode, 0, done.stderr)
        self.assertEqual(out["tag"], "moai-adk-v4.2648.7")


class VerifyTest(unittest.TestCase):
    """AC-2, AC-3: the `verify` step compares the built binaries with the plan and writes the
    release note."""

    VERSION = "4.2648.7"

    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        self.dir = pathlib.Path(self._tmp.name)
        self.summary = self.dir / "summary.md"
        self.bin = self.dir / "target/release"
        self.bin.mkdir(parents=True)

    def binary(self, name, first, second):
        path = self.bin / name
        path.write_text(f"#!/bin/sh\nprintf '%s\\n' '{first}' '{second}'\n", encoding="utf-8")
        path.chmod(0o755)

    def good(self, name, version=None, build="7"):
        self.binary(name, f"moai-cli {version or self.VERSION} (rust)",
                    f"CI build {build} (MOAI_BUILD_NUMBER), built 2026-11-24 UTC "
                    f"(ISO week 2026-W48), package 2026.11.0-rc.1")

    def verify(self, tag=None, prerelease="false", version=None, build="7"):
        env = dict(os.environ, VERSION=version or self.VERSION, BUILD=build,
                   TAG=tag or f"moai-adk-v{version or self.VERSION}", SHA=SHA,
                   PRERELEASE=prerelease, GITHUB_STEP_SUMMARY=str(self.summary))
        self.summary.write_text("", encoding="utf-8")
        return bash(script("build", "verify"), self.dir, env)

    def test_binaries_that_report_the_plans_version_and_a_ci_build_pass(self):
        self.good("moai-cli")
        self.good("moai")
        done = self.verify()
        self.assertEqual(done.returncode, 0, done.stderr)
        note = self.summary.read_text(encoding="utf-8")
        for text in ["moai-adk-v4.2648.7", "4.2648.7", SHA, "CI build 7 (MOAI_BUILD_NUMBER)",
                     "pre-release: no"]:
            self.assertIn(text, note)

    def test_a_candidate_note_says_pre_release(self):
        self.good("moai-cli")
        self.good("moai")
        done = self.verify(prerelease="true")
        self.assertEqual(done.returncode, 0, done.stderr)
        self.assertIn("pre-release: yes", self.summary.read_text(encoding="utf-8"))

    def test_a_binary_of_another_version_fails(self):
        self.good("moai-cli")
        self.good("moai", version="4.2648.6")
        done = self.verify()
        self.assertNotEqual(done.returncode, 0, done.stdout)
        self.assertIn("target/release/moai", done.stderr)
        self.assertEqual(self.summary.read_text(encoding="utf-8"), "", "no note for a failed check")

    def test_a_local_build_line_fails(self):
        """AC-3: a release binary that says `local build` or `unnumbered build` is not a release."""
        for second in ["local build 7 (MOAI_LOCAL_BUILD), built 2026-11-24 UTC",
                       "unnumbered build, built 2026-11-24 UTC",
                       "CI build 8 (MOAI_BUILD_NUMBER), built 2026-11-24 UTC",
                       "CI build 7 (MOAI_BUILD_NUMBER)x", ""]:
            with self.subTest(second=second):
                self.good("moai")
                self.binary("moai-cli", f"moai-cli {self.VERSION} (rust)", second)
                done = self.verify()
                self.assertNotEqual(done.returncode, 0, done.stdout)
                self.assertIn("CI build 7 (MOAI_BUILD_NUMBER), built", done.stderr)

    def test_a_tag_that_does_not_spell_the_version_fails(self):
        self.good("moai-cli")
        self.good("moai")
        for tag in ["moai-adk-v4.2648.8", "moai-cli-v4.2648.7", "v4.2648.7", "moai-adk-v4.2648.7-rc.1"]:
            with self.subTest(tag=tag):
                done = self.verify(tag=tag)
                self.assertNotEqual(done.returncode, 0, done.stdout)
                self.assertIn("moai-adk-v4.2648.7", done.stderr)

    def test_a_missing_binary_fails(self):
        self.good("moai-cli")
        done = self.verify()
        self.assertNotEqual(done.returncode, 0, done.stdout)


if __name__ == "__main__":
    unittest.main()
