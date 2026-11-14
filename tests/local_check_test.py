"""tools/local_check.sh (#2906 S5a, AC20): one local check set that the CI check job also runs.

The script runs the check steps of the CI `check` job in the same order with the same
arguments and ends with one machine-readable line. These tests pin the step list (so the
script and CI cannot drift apart), the final line's format, the dirty-tree rule and the
stop-at-first-failure rule. Real runs use stub `cargo`/`python3` in a temporary git
repository, so no cargo build runs here.
"""
import re
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "tools" / "local_check.sh"
CI = ROOT / ".github" / "workflows" / "ci.yml"
LINE = re.compile(
    r"^local-check: head=[0-9a-f]{40} result=(pass|fail) (dirty=1 )?"
    r"cmd=tools/local_check\.sh at=\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$")

# The check steps of the CI `check` job, quoted from .github/workflows/ci.yml at base
# cdb4b017c777530fb5126dd0a753930b738830a9 (the `run <label> <command>` lines of the
# "Check set" step), in order. Two differences, both on purpose: `python` became
# `python3` (setup-python's 3.12 answers to both on the runner; macOS has no `python`),
# and diff-check reads its range from LOCAL_CHECK_BASE / LOCAL_CHECK_DIFF_HEAD, which the
# workflow sets to the same `${{ ... }}` values it used inline.
CI_LINUX_STEPS = [
    ("fetch", "cargo fetch --locked"),
    ("test", "bash -c RUST_MIN_STACK=8388608 cargo nextest run --offline --locked --no-fail-fast"
             " 2>&1 | tee ci-test.log; exit \"${PIPESTATUS[0]}\""),
    ("doctest", "bash -c RUST_MIN_STACK=8388608 cargo test --offline --locked --doc"
                " 2>&1 | tee ci-doctest.log; exit \"${PIPESTATUS[0]}\""),
    ("clippy", "cargo clippy --offline --locked --all-targets -- -D warnings"),
    ("fmt", "cargo fmt --check"),
    ("unittest", "bash -c python3 -m unittest discover -v -s tests -p '*_test.py'"
                 " 2>&1 | tee ci-unittest.log; exit \"${PIPESTATUS[0]}\""),
    ("diff-check", "git diff --check BASE...HEAD"),
]
STEP = re.compile(r"^step (\S+): (.*)$")
INLINE_CHECK_COMMANDS = ("cargo fetch", "cargo nextest", "cargo test", "cargo clippy",
                         "cargo fmt", "unittest discover")


def git(cwd, *args):
    return subprocess.run(["git", *args], cwd=cwd, check=True, capture_output=True,
                          text=True).stdout.strip()


def last_line(text):
    lines = [line for line in text.splitlines() if line.strip()]
    return lines[-1] if lines else ""


class FakeRepo:
    """A temporary git repository holding a copy of the script, with stub tools on PATH."""

    def __init__(self, fail_label=None):
        self.tmp = Path(tempfile.mkdtemp(prefix="local-check-"))
        self.repo = self.tmp / "repo"
        (self.repo / "tools").mkdir(parents=True)
        shutil.copy2(SCRIPT, self.repo / "tools" / "local_check.sh")
        (self.repo / "README.md").write_text("one\n")
        (self.repo / ".gitignore").write_text("*.log\n")  # as the repository's own
        env = self.git_env()
        for args in (["init", "-q"], ["add", "-A"], ["commit", "-q", "-m", "one"]):
            subprocess.run(["git", *args], cwd=self.repo, env=env, check=True,
                           capture_output=True)
        (self.repo / "README.md").write_text("one\ntwo\n")
        for args in (["add", "-A"], ["commit", "-q", "-m", "two"]):
            subprocess.run(["git", *args], cwd=self.repo, env=env, check=True,
                           capture_output=True)
        self.calls = self.tmp / "calls"
        self.bin = self.tmp / "bin"
        self.bin.mkdir()
        # Each stub records its arguments; the one named by fail_label exits 1.
        for name in ("cargo", "python3"):
            stub = self.bin / name
            stub.write_text(
                "#!/bin/sh\n"
                f"echo \"{name} $*\" >> '{self.calls}'\n"
                f"case \"{name} $*\" in *'{fail_label or '@none@'}'*) exit 1;; esac\n"
                "exit 0\n")
            stub.chmod(0o755)

    def git_env(self):
        env = {"PATH": "/usr/bin:/bin", "HOME": str(self.tmp), "LC_ALL": "C",
               "GIT_CONFIG_NOSYSTEM": "1", "GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@example.invalid",
               "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@example.invalid"}
        return env

    def run(self, **extra):
        env = self.git_env()
        env["PATH"] = f"{self.bin}:/usr/bin:/bin"  # no `moai`: steps run without a slot
        env.update(extra)
        return subprocess.run(["bash", str(self.repo / "tools" / "local_check.sh")],
                              cwd=self.tmp, env=env, capture_output=True, text=True,
                              timeout=60)

    def recorded(self):
        return self.calls.read_text().splitlines() if self.calls.exists() else []

    def head(self):
        return git(self.repo, "rev-parse", "HEAD")

    def close(self):
        shutil.rmtree(self.tmp, ignore_errors=True)


class LocalCheckStepsTest(unittest.TestCase):
    def dry_steps(self, repo):
        proc = repo.run(LOCAL_CHECK_DRY_RUN="1", LOCAL_CHECK_BASE="BASE",
                        LOCAL_CHECK_DIFF_HEAD="HEAD")
        self.assertEqual(proc.returncode, 1, proc.stderr)
        steps = [STEP.match(line).groups() for line in proc.stdout.splitlines()
                 if STEP.match(line)]
        return steps, proc

    def test_local_check_matches_the_ci_linux_steps(self):
        """AC20: the script's steps are the CI linux steps (labels, commands, order)."""
        repo = FakeRepo()
        self.addCleanup(repo.close)
        steps, _ = self.dry_steps(repo)
        self.assertEqual(steps, CI_LINUX_STEPS)

    def test_ci_check_job_runs_the_script_instead_of_inline_steps(self):
        """AC20: the CI check job calls the script for its full check set and repeats
        none of its commands inline; only the docs-only branch keeps its own diff-check."""
        text = CI.read_text(encoding="utf-8")
        check = text.split("\n  check:\n", 1)[1].split("\n  ci:\n", 1)[0]
        body = check.split("- name: Check set", 1)[1]
        docs_only, full = body.split("\n          fi\n", 1)
        self.assertIn('if [ "$CI_SCOPE" = "docs-only" ]; then', docs_only)
        self.assertEqual(len(re.findall(r'(?m)^\s+LOCAL_CHECK_KEEP_GOING=1 LOCAL_CHECK_BASE="[^"]+" '
                                        r'LOCAL_CHECK_DIFF_HEAD="[^"]+" bash tools/local_check\.sh '
                                        r'\|\| failed=1$', full)), 1)
        self.assertIn("github.event.pull_request.base.sha || 'HEAD~1'", full)
        self.assertIn("github.event.pull_request.head.sha || 'HEAD'", full)
        # Commands only: summary text (`echo ...`) and comments may name a command.
        def commands(text):
            return [line.strip() for line in text.splitlines()
                    if not line.strip().startswith(("#", "echo "))]
        for command in INLINE_CHECK_COMMANDS + ("git diff --check",):
            with self.subTest(command=command):
                self.assertEqual([c for c in commands(full) if command in c], [])
        for command in INLINE_CHECK_COMMANDS:
            with self.subTest(job_wide=command):
                self.assertEqual([c for c in commands(check)
                                  if command in c and c != "cargo nextest --version"], [])

    def test_real_run_executes_the_same_steps_in_order(self):
        repo = FakeRepo()
        self.addCleanup(repo.close)
        proc = repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(repo.recorded(), [
            "cargo fetch --locked",
            "cargo nextest run --offline --locked --no-fail-fast",
            "cargo test --offline --locked --doc",
            "cargo clippy --offline --locked --all-targets -- -D warnings",
            "cargo fmt --check",
            "python3 -m unittest discover -v -s tests -p *_test.py",
        ])


class LocalCheckLineTest(unittest.TestCase):
    def test_dry_run_prints_the_line_and_never_a_pass(self):
        """A dry run checked nothing, so its line says fail and no cargo runs."""
        repo = FakeRepo()
        self.addCleanup(repo.close)
        proc = repo.run(LOCAL_CHECK_DRY_RUN="1")
        line = last_line(proc.stdout)
        self.assertRegex(line, LINE)
        self.assertIn(f"head={repo.head()} result=fail cmd=", line)
        self.assertEqual(proc.returncode, 1)
        self.assertEqual(repo.recorded(), [])

    def test_clean_pass_prints_one_pass_line_for_head(self):
        repo = FakeRepo()
        self.addCleanup(repo.close)
        proc = repo.run()
        line = last_line(proc.stdout)
        self.assertRegex(line, LINE)
        self.assertIn(f"head={repo.head()} result=pass cmd=tools/local_check.sh at=", line)
        self.assertEqual(sum(1 for out in proc.stdout.splitlines()
                             if out.startswith("local-check:")), 1)

    def test_dirty_tree_is_a_fail_with_dirty_flag(self):
        """(c) a tree with uncommitted changes is not the HEAD being checked."""
        for change in ("modified", "untracked"):
            with self.subTest(change=change):
                repo = FakeRepo()
                self.addCleanup(repo.close)
                if change == "modified":
                    (repo.repo / "README.md").write_text("changed\n")
                else:
                    (repo.repo / "new.txt").write_text("x\n")
                for extra in ({}, {"LOCAL_CHECK_DRY_RUN": "1"}):
                    proc = repo.run(**extra)
                    line = last_line(proc.stdout)
                    self.assertRegex(line, LINE)
                    self.assertIn(" result=fail dirty=1 cmd=", line)
                    self.assertNotEqual(proc.returncode, 0)

    def test_ignored_files_do_not_make_the_tree_dirty(self):
        repo = FakeRepo()
        self.addCleanup(repo.close)
        (repo.repo / "ci-test.log").write_text("x\n")
        proc = repo.run()
        self.assertIn(" result=pass cmd=", last_line(proc.stdout))

    def test_first_failure_stops_the_run_and_fails(self):
        repo = FakeRepo(fail_label="clippy")
        self.addCleanup(repo.close)
        proc = repo.run()
        self.assertEqual(proc.returncode, 1)
        self.assertIn(" result=fail cmd=", last_line(proc.stdout))
        self.assertEqual(repo.recorded()[-1],
                         "cargo clippy --offline --locked --all-targets -- -D warnings")

    def test_keep_going_runs_every_step_and_still_fails(self):
        """CI sets LOCAL_CHECK_KEEP_GOING=1 to keep its rule: record every step's exit."""
        repo = FakeRepo(fail_label="clippy")
        self.addCleanup(repo.close)
        summary = repo.tmp / "summary.md"
        proc = repo.run(LOCAL_CHECK_KEEP_GOING="1", GITHUB_STEP_SUMMARY=str(summary))
        self.assertEqual(proc.returncode, 1)
        self.assertIn(" result=fail cmd=", last_line(proc.stdout))
        self.assertEqual(len(repo.recorded()), 6)
        rows = summary.read_text().splitlines()
        self.assertEqual([r.split(" | ")[0] for r in rows],
                         ["| " + label for label, _ in CI_LINUX_STEPS])
        self.assertTrue(rows[3].endswith(" | 1 |"), rows[3])

    def test_diff_check_failure_is_a_fail(self):
        repo = FakeRepo()
        self.addCleanup(repo.close)
        (repo.repo / "README.md").write_text("one \n")
        env = repo.git_env()
        for args in (["add", "-A"], ["commit", "-q", "-m", "trailing space"]):
            subprocess.run(["git", *args], cwd=repo.repo, env=env, check=True)
        proc = repo.run(LOCAL_CHECK_BASE="HEAD~1")
        self.assertEqual(proc.returncode, 1)
        self.assertIn(" result=fail cmd=", last_line(proc.stdout))


if __name__ == "__main__":
    unittest.main()
