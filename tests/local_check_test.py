"""tools/local_check.sh (#2906 S5a, AC20): one local check set that the CI check job also runs.

The script runs the check steps of the CI `check` job in the same order with the same
arguments and ends with one machine-readable line. These tests pin the step list (so the
script and CI cannot drift apart), the final line's format, the dirty-tree rule and the
stop-at-first-failure rule. Real runs use stub `cargo`/`python3`/`rustc` in a temporary git
repository, so no cargo build runs here.

#3066 P2: the default mode checks a candidate (the head merged with the latest tip of the
base branch) and prints the long line; `LOCAL_CHECK_MODE=in-place` (CI) keeps the short
line. The first classes below pin the in-place behaviour; `LocalCheckCandidateTest` pins the
candidate mode against a bare origin and a clone in temporary folders, with a temporary HOME.
"""
import hashlib
import os
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
# The long line of the candidate mode, as src/gate.rs `parse_local_check_evidence` reads it:
# column 0, single spaces, this field order, lowercase hex, no `dirty=1`.
LONG = re.compile(
    r"^local-check: head=(?P<head>[0-9a-f]{40}) result=(?P<result>pass|fail) "
    r"base=(?P<base>[0-9a-f]{40}|none) tree=(?P<tree>[0-9a-f]{40}|none) "
    r"script=(?P<script>[0-9a-f]{40}) rustc=(?P<rustc>[A-Za-z0-9._+-]{1,32}) "
    r"run=(?P<run>[0-9a-f]{12}) cmd=tools/local_check\.sh "
    r"at=(?P<at>\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ)$")
BASE_BRANCH = "v2026.11.0"

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
    # C-75 S0: build.rs includes src/ops/*.rs through `#[path]` in a generated OUT_DIR file,
    # so `cargo fmt` never reaches them; this step checks them with rustfmt directly.
    ("fmt-ops", "bash -c rustfmt --check --edition 2021 src/ops/*.rs"),
    ("unittest","bash -c python3 -m unittest discover -v -s tests -p '*_test.py'"
                 " 2>&1 | tee ci-unittest.log; exit \"${PIPESTATUS[0]}\""),
    ("diff-check", "git diff --check BASE...HEAD"),
]
STEP = re.compile(r"^step (\S+): (.*)$")
INLINE_CHECK_COMMANDS = ("cargo fetch", "cargo nextest", "cargo test", "cargo clippy",
                         "cargo fmt", "unittest discover")


def git(cwd, *args):
    return subprocess.run(["git", *args], cwd=cwd, check=True, capture_output=True,
                          text=True).stdout.strip()


def write_stubs(bin_dir, calls, where, fail_label):
    """Stub `cargo`, `python3`, `rustfmt` and `rustc` (no real build ever runs here).

    `cargo`/`python3`/`rustfmt` append their arguments to `calls`, and where they ran to `where`
    (tab separated: the physical cwd, its HEAD tree, its first and second parent); the one
    named by fail_label exits 1. They also exit 1 when a folder holds both
    `rule_forbids_forbidden.txt` and `forbidden.txt`: a rule the base branch adds that a head
    cannot know (the #3062 shape). `rustc` prints a version, or STUB_RUSTC_OUT with exit
    STUB_RUSTC_EXIT when that is set."""
    for name in ("cargo", "python3", "rustfmt"):
        stub = bin_dir / name
        stub.write_text(
            "#!/bin/sh\n"
            f"echo \"{name} $*\" >> '{calls}'\n"
            "printf '%s\\t%s\\t%s\\t%s\\n' \"$(pwd -P)\" \"$(git rev-parse 'HEAD^{tree}' 2>/dev/null)\""
            " \"$(git rev-parse 'HEAD^1' 2>/dev/null)\" \"$(git rev-parse 'HEAD^2' 2>/dev/null)\""
            f" >> '{where}'\n"
            "if [ -f rule_forbids_forbidden.txt ] && [ -f forbidden.txt ]; then\n"
            "  echo 'rule: forbidden.txt must not exist' >&2; exit 1\n"
            "fi\n"
            f"case \"{name} $*\" in *'{fail_label or '@none@'}'*) exit 1;; esac\n"
            "exit 0\n")
        stub.chmod(0o755)
    rustc = bin_dir / "rustc"
    rustc.write_text(
        "#!/bin/sh\n"
        "if [ -n \"${STUB_RUSTC_OUT+x}\" ]; then\n"
        "  printf '%s\\n' \"$STUB_RUSTC_OUT\"; exit \"${STUB_RUSTC_EXIT:-0}\"\n"
        "fi\n"
        "echo 'rustc 1.99.0 (stub 2026-01-01)'\n")
    rustc.chmod(0o755)


def last_line(text):
    lines = [line for line in text.splitlines() if line.strip()]
    return lines[-1] if lines else ""


class FakeRepo:
    """A temporary git repository holding a copy of the script, with stub tools on PATH."""

    def __init__(self, fail_label=None, slot=False):
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
        self.where = self.tmp / "where"
        self.bin = self.tmp / "bin"
        self.bin.mkdir()
        write_stubs(self.bin, self.calls, self.where, fail_label)
        self.slot_calls = self.tmp / "slot_calls"
        if slot:
            # A stub `moai` offering `heavy-slot-acquire`: it records its arguments and runs the
            # command after `--`, or exits 75 (the slot's refusal) when STUB_SLOT_REFUSE is set.
            stub = self.bin / "moai"
            stub.write_text(
                "#!/bin/sh\n"
                "[ \"$1 $2\" = 'heavy-slot-acquire --help' ] && exit 0\n"
                f"echo \"$*\" >> '{self.slot_calls}'\n"
                "[ -n \"${STUB_SLOT_REFUSE:-}\" ] && exit 75\n"
                "while [ \"$#\" -gt 0 ] && [ \"$1\" != -- ]; do shift; done\n"
                "shift\n"
                "exec \"$@\"\n")
            stub.chmod(0o755)

    def git_env(self):
        env = {"PATH": "/usr/bin:/bin", "HOME": str(self.tmp), "LC_ALL": "C",
               "GIT_CONFIG_NOSYSTEM": "1", "GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@example.invalid",
               "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@example.invalid"}
        return env

    def run(self, **extra):
        env = self.git_env()
        env["PATH"] = f"{self.bin}:/usr/bin:/bin"  # no `moai` unless slot=True: no slot
        env.update(extra)
        # A value of None removes the variable (an unset HOME is not an empty one).
        env = {k: v for k, v in env.items() if v is not None}
        # These tests pin the in-place behaviour (CI's); the candidate mode has its own class.
        env.setdefault("LOCAL_CHECK_MODE", "in-place")
        return subprocess.run(["bash", str(self.repo / "tools" / "local_check.sh")],
                              cwd=self.tmp, env=env, capture_output=True, text=True,
                              timeout=60)

    def recorded(self):
        return self.calls.read_text().splitlines() if self.calls.exists() else []

    def slot_recorded(self):
        return self.slot_calls.read_text().splitlines() if self.slot_calls.exists() else []

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

    def test_ci_uses_the_in_place_mode_exactly_once(self):
        """#3066 P2 (M13): the workflow names LOCAL_CHECK_MODE=in-place on the one script call."""
        text = CI.read_text(encoding="utf-8")
        self.assertEqual(text.count("LOCAL_CHECK_MODE=in-place"), 1)
        self.assertEqual(len(re.findall(r"LOCAL_CHECK_MODE", text)), 1)
        self.assertEqual(len(re.findall(r"bash tools/local_check\.sh", text)), 1)
        line = next(l for l in text.splitlines() if "bash tools/local_check.sh" in l)
        self.assertIn("LOCAL_CHECK_MODE=in-place ", line)

    def test_ci_check_job_runs_the_script_instead_of_inline_steps(self):
        """AC20: the CI check job calls the script for its full check set and repeats
        none of its commands inline; only the docs-only branch keeps its own diff-check."""
        text = CI.read_text(encoding="utf-8")
        check = text.split("\n  check:\n", 1)[1].split("\n  ci:\n", 1)[0]
        body = check.split("- name: Check set", 1)[1]
        docs_only, full = body.split("\n          fi\n", 1)
        self.assertIn('if [ "$CI_SCOPE" = "docs-only" ]; then', docs_only)
        # #3066 P2: CI's checkout already is the merge ref, so it asks for the in-place mode;
        # without it the script would fetch and check a candidate on the runner.
        self.assertEqual(len(re.findall(r'(?m)^\s+LOCAL_CHECK_MODE=in-place LOCAL_CHECK_KEEP_GOING=1 '
                                        r'LOCAL_CHECK_BASE="[^"]+" '
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
            # No src/ops/ in the fake repository, so bash passes the pattern through as is.
            "rustfmt --check --edition 2021 src/ops/*.rs",
            "python3 -m unittest discover -v -s tests -p *_test.py",
        ])


class LocalCheckPythonSlotTest(unittest.TestCase):
    """SPEC-20261119-python-slot-pre-ready: the whole unittest discover waits in a python slot,
    a slot folder apart from the cargo slots; without `moai heavy-slot-acquire` it runs directly."""

    def setUp(self):
        self.repo = FakeRepo(slot=True)
        self.addCleanup(self.repo.close)
        self.default_dir = self.repo.tmp / ".cache" / "moai-cli" / "python-slots"

    def split(self):
        slot = self.repo.slot_recorded()
        return [l for l in slot if " --dir " in l], [l for l in slot if " --dir " not in l]

    def test_only_the_unittest_step_uses_the_python_slot_with_the_default_folder_and_limit(self):
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        python, cargo = self.split()
        self.assertEqual(len(python), 1, python)
        self.assertEqual(len(cargo), 3, cargo)  # test, doctest, clippy keep the heavy slot
        self.assertTrue(python[0].startswith(
            f"heavy-slot-acquire --dir {self.default_dir} --max 2 --min-free-gb 50 "
            "--name local-check-py -- bash -c python3 -m unittest discover"), python[0])
        self.assertTrue(all(l.startswith("heavy-slot-acquire --max 3 ") for l in cargo), cargo)
        # The command still ran inside the slot.
        self.assertIn("python3 -m unittest discover -v -s tests -p *_test.py",
                      self.repo.recorded())

    def test_xdg_cache_home_names_the_default_folder(self):
        xdg = self.repo.tmp / "xdg"
        self.repo.run(XDG_CACHE_HOME=str(xdg))
        python, _ = self.split()
        self.assertIn(f" --dir {xdg}/moai-cli/python-slots ", python[0])

    def test_the_folder_and_the_limit_can_be_set(self):
        self.repo.run(MOAI_PYTHON_SLOT_DIR="/x/y", MOAI_PYTHON_SLOT_MAX="1")
        python, _ = self.split()
        self.assertIn(" --dir /x/y --max 1 ", python[0])

    def test_empty_settings_mean_the_defaults(self):
        self.repo.run(MOAI_PYTHON_SLOT_DIR="", MOAI_PYTHON_SLOT_MAX="")
        python, _ = self.split()
        self.assertIn(f" --dir {self.default_dir} --max 2 ", python[0])

    def test_an_unset_or_empty_home_never_stops_the_script_and_the_step_runs_directly(self):
        # Audit B1: the folder default read $HOME at start-up and died with `unbound variable`
        # before the EXIT trap, so not even the last line was printed.
        for label, extra in (("unset", {"HOME": None}), ("empty", {"HOME": ""})):
            with self.subTest(home=label):
                repo = FakeRepo(slot=True)
                self.addCleanup(repo.close)
                proc = repo.run(**extra)
                self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
                self.assertNotIn("unbound variable", proc.stderr)
                self.assertIn("no python slot folder", proc.stderr)
                self.assertTrue(LINE.match(last_line(proc.stdout)), proc.stdout)
                self.assertEqual([l for l in repo.slot_recorded() if " --dir " in l], [])
                # Never the cargo slot folder either (REQ-12): no slot line carries the python run.
                self.assertEqual([l for l in repo.slot_recorded() if "unittest" in l], [])
                self.assertIn("python3 -m unittest discover -v -s tests -p *_test.py",
                              repo.recorded())

    def test_an_empty_xdg_cache_home_falls_back_to_home(self):
        self.repo.run(XDG_CACHE_HOME="")
        python, _ = self.split()
        self.assertIn(f" --dir {self.default_dir} --max 2 ", python[0])

    def test_the_folder_setting_alone_works_without_home(self):
        self.repo.run(HOME=None, MOAI_PYTHON_SLOT_DIR="/x/y")
        python, _ = self.split()
        self.assertIn(" --dir /x/y --max 2 ", python[0])

    def test_the_python_slot_name_is_the_heavy_name_with_a_suffix(self):
        self.repo.run(MOAI_HEAVY_SLOT_NAME="lane-7")
        python, cargo = self.split()
        self.assertIn(" --name lane-7-py -- ", python[0])
        for line in cargo:
            self.assertIn(" --name lane-7 -- ", line)

    def test_a_refusal_of_the_python_step_fails_the_run_with_its_exit_code(self):
        proc = self.repo.run(STUB_SLOT_REFUSE="1", LOCAL_CHECK_KEEP_GOING="1")
        self.assertEqual(proc.returncode, 1)
        self.assertIn("step unittest: exit 75", proc.stderr)
        self.assertIn(" result=fail ", last_line(proc.stdout))
        self.assertNotIn("python3 -m unittest", " ".join(self.repo.recorded()))

    def test_without_the_slot_command_the_step_runs_directly(self):
        repo = FakeRepo()
        self.addCleanup(repo.close)
        proc = repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(repo.slot_recorded(), [])
        self.assertIn("python3 -m unittest discover -v -s tests -p *_test.py", repo.recorded())


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
        self.assertEqual(len(repo.recorded()), 7)
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


def real_rustfmt():
    """The pinned toolchain's rustfmt binary (rust-toolchain.toml), not the rustup proxy: the
    fake repository runs with a temporary HOME, where a proxy would find no toolchain."""
    rustup = shutil.which("rustup")
    if rustup:
        proc = subprocess.run([rustup, "which", "rustfmt"], cwd=ROOT, capture_output=True,
                              text=True)
        if proc.returncode == 0 and proc.stdout.strip():
            return proc.stdout.strip()
    return shutil.which("rustfmt")


# rustfmt rewrites this to `fn misformatted() -> u8 {\n    1\n}`.
MISFORMATTED = "\nfn   misformatted( )->u8{1}\n"


class LocalCheckOpsFormatTest(unittest.TestCase):
    """C-75 S0: `cargo fmt --check` never reaches src/ops/*.rs (build.rs includes them by
    `#[path]` from a file it writes into OUT_DIR), so the fmt-ops step checks them with the
    real rustfmt. Here `cargo` stays a stub that passes, as `cargo fmt --check` does on such a
    file; the repository's own src/ops/*.rs are copied into the fake repository."""

    def repo_with_ops(self, change=None, ops=True):
        rustfmt = real_rustfmt()
        self.assertIsNotNone(rustfmt, "no rustfmt: install the rust-toolchain.toml toolchain")
        repo = FakeRepo()
        self.addCleanup(repo.close)
        (repo.bin / "rustfmt").unlink()
        (repo.bin / "rustfmt").symlink_to(rustfmt)
        if ops:
            target = repo.repo / "src" / "ops"
            target.mkdir(parents=True)
            for source in sorted((ROOT / "src" / "ops").glob("*.rs")):
                shutil.copy2(source, target / source.name)
            if change:
                change(target)
            env = repo.git_env()
            for args in (["add", "-A"], ["commit", "-q", "-m", "ops"]):
                subprocess.run(["git", *args], cwd=repo.repo, env=env, check=True,
                               capture_output=True)
        return repo

    def test_the_step_uses_the_crate_edition(self):
        edition = re.search(r'(?m)^edition = "(\d+)"$',
                            (ROOT / "Cargo.toml").read_text(encoding="utf-8")).group(1)
        script = SCRIPT.read_text(encoding="utf-8")
        self.assertEqual(script.count(f"rustfmt --check --edition {edition} src/ops/*.rs"), 1)
        self.assertEqual(script.count("rustfmt --check"), 1)

    def test_formatted_ops_files_pass(self):
        """The repository's src/ops/*.rs as they are: every file must already be formatted."""
        repo = self.repo_with_ops()
        proc = repo.run()
        drift = sorted(set(re.findall(r"(?m)^Diff in .*/src/ops/(\w+\.rs):", proc.stdout)))
        self.assertEqual(drift, [], "src/ops files rustfmt would change")
        self.assertNotIn("step fmt-ops: exit", proc.stderr)
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn(" result=pass cmd=", last_line(proc.stdout))

    def test_one_misformatted_ops_file_fails_the_step(self):
        def misformat(ops):
            with (ops / "package_inspect.rs").open("a", encoding="utf-8") as f:
                f.write(MISFORMATTED)
        repo = self.repo_with_ops(misformat)
        proc = repo.run(LOCAL_CHECK_KEEP_GOING="1")
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertIn(" result=fail cmd=", last_line(proc.stdout))
        self.assertIn("step fmt-ops: exit 1", proc.stderr)
        self.assertIn("package_inspect.rs", proc.stdout)
        # The cargo fmt step (a passing stub, like the real one on this file) is not the one
        # that failed.
        self.assertNotIn("step fmt: exit", proc.stderr)

    def test_no_ops_files_fails_the_step(self):
        """With nothing to check the pattern reaches rustfmt unexpanded; it must not pass."""
        repo = self.repo_with_ops(ops=False)
        proc = repo.run()
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertIn("step fmt-ops: exit 1", proc.stderr)


def long_fields(case, line):
    """The fields of a candidate-mode line; also the gate's rules for `none` (P1 parser)."""
    match = LONG.match(line)
    case.assertIsNotNone(match, f"not a long local-check line: {line!r}")
    fields = match.groupdict()
    if fields["result"] == "pass":
        case.assertNotIn("none", (fields["base"], fields["tree"]), line)
    if fields["base"] == "none":
        case.assertEqual(fields["tree"], "none", line)
    if "none" in (fields["base"], fields["tree"]):
        case.assertEqual(fields["result"], "fail", line)
    return fields


class CandidateRepo:
    """A bare origin, a clone on branch `feature`, stubs on PATH and a private TMPDIR."""

    def __init__(self, fail_label=None):
        self.tmp = Path(tempfile.mkdtemp(prefix="lc-cand-test-")).resolve()
        self.origin = self.tmp / "origin.git"
        self.repo = self.tmp / "repo"
        self.other = self.tmp / "other"
        self.tmpdir = self.tmp / "tmpdir"
        self.tmpdir.mkdir()
        self.calls = self.tmp / "calls"
        self.where = self.tmp / "where"
        self.bin = self.tmp / "bin"
        self.bin.mkdir()
        write_stubs(self.bin, self.calls, self.where, fail_label)
        self.git(self.tmp, "init", "-q", "--bare", "-b", BASE_BRANCH, str(self.origin))
        (self.repo / "tools").mkdir(parents=True)
        shutil.copy2(SCRIPT, self.repo / "tools" / "local_check.sh")
        self.git(self.repo, "init", "-q", "-b", BASE_BRANCH)
        self.git(self.repo, "remote", "add", "origin", str(self.origin))
        self.commit(self.repo, {"README.md": "one\ntwo\n", ".gitignore": "*.log\n"}, "base")
        self.git(self.repo, "push", "-q", "origin", BASE_BRANCH)
        self.git(self.repo, "checkout", "-q", "-b", "feature")

    def git_env(self):
        return {"PATH": "/usr/bin:/bin", "HOME": str(self.tmp), "LC_ALL": "C",
                "GIT_CONFIG_NOSYSTEM": "1", "GIT_AUTHOR_NAME": "t",
                "GIT_AUTHOR_EMAIL": "t@example.invalid", "GIT_COMMITTER_NAME": "t",
                "GIT_COMMITTER_EMAIL": "t@example.invalid"}

    def git(self, cwd, *args):
        return subprocess.run(["git", *args], cwd=cwd, env=self.git_env(), check=True,
                              capture_output=True, text=True).stdout.strip()

    def commit(self, cwd, files, message):
        for name, text in files.items():
            (Path(cwd) / name).write_text(text)
        self.git(cwd, "add", "-A")
        self.git(cwd, "commit", "-q", "-m", message)
        return self.git(cwd, "rev-parse", "HEAD")

    def _other(self):
        if not self.other.exists():
            self.git(self.tmp, "clone", "-q", str(self.origin), str(self.other))
        self.git(self.other, "fetch", "-q", "origin")
        self.git(self.other, "checkout", "-q", "-B", "work", f"origin/{BASE_BRANCH}")

    def advance_origin(self, files, message, branch=BASE_BRANCH):
        """A new commit on origin's `branch`, made by someone else (not the clone)."""
        self._other()
        self.commit(self.other, files, message)
        self.git(self.other, "push", "-q", "origin", f"HEAD:refs/heads/{branch}")
        return self.git(self.other, "rev-parse", "HEAD")

    def replace_origin_history(self, files):
        """origin's base branch becomes a root commit with no history in common."""
        self._other()
        self.git(self.other, "checkout", "-q", "--orphan", "rewritten")
        self.git(self.other, "rm", "-rfq", ".")
        self.commit(self.other, files, "unrelated root")
        self.git(self.other, "push", "-q", "--force", "origin", f"HEAD:refs/heads/{BASE_BRANCH}")
        return self.git(self.other, "rev-parse", "HEAD")

    def head(self):
        return self.git(self.repo, "rev-parse", "HEAD")

    def origin_tip(self):
        return self.git(self.repo, "ls-remote", "origin", f"refs/heads/{BASE_BRANCH}").split()[0]

    def local_origin_ref(self):
        return self.git(self.repo, "rev-parse", f"refs/remotes/origin/{BASE_BRANCH}")

    def run(self, **extra):
        env = self.git_env()
        env["PATH"] = f"{self.bin}:/usr/bin:/bin"  # no `moai`: steps run without a slot
        env["TMPDIR"] = str(self.tmpdir)
        env.update(extra)
        return subprocess.run(["bash", str(self.repo / "tools" / "local_check.sh")],
                              cwd=self.tmp, env=env, capture_output=True, text=True,
                              timeout=120)

    def recorded(self):
        return self.calls.read_text().splitlines() if self.calls.exists() else []

    def where_rows(self):
        if not self.where.exists():
            return []
        return [line.split("\t") for line in self.where.read_text().splitlines()]

    def leftovers(self):
        return sorted(p.name for p in self.tmpdir.iterdir())

    def worktrees(self):
        return [line for line in self.git(self.repo, "worktree", "list", "--porcelain").splitlines()
                if line.startswith("worktree ")]

    def snapshot(self):
        """Everything the user can see of the working folder, except the one ref a run may move."""
        status = self.git(self.repo, "status", "--porcelain=v2", "--branch",
                          "--untracked-files=all", "--ignored")
        files = {}
        for base, dirs, names in os.walk(self.repo):
            dirs[:] = [d for d in dirs if d != ".git"]
            for name in names:
                path = Path(base) / name
                files[str(path.relative_to(self.repo))] = (
                    oct(path.stat().st_mode), hashlib.sha1(path.read_bytes()).hexdigest())
        moving = f"refs/remotes/origin/{BASE_BRANCH} "
        refs = [r for r in self.git(self.repo, "for-each-ref",
                                    "--format=%(refname) %(objectname)").splitlines()
                if not r.startswith(moving)]
        return {
            "status": status,
            "files": files,
            "index": hashlib.sha1((self.repo / ".git" / "index").read_bytes()).hexdigest(),
            "ls-files": self.git(self.repo, "ls-files", "-s"),
            "head": self.git(self.repo, "rev-parse", "HEAD"),
            "branch": self.git(self.repo, "symbolic-ref", "HEAD"),
            "refs": refs,
            "stash": self.git(self.repo, "stash", "list"),
            "worktrees": self.git(self.repo, "worktree", "list", "--porcelain"),
        }

    def close(self):
        shutil.rmtree(self.tmp, ignore_errors=True)


class LocalCheckCandidateTest(unittest.TestCase):
    """#3066 P2: the default mode checks the head merged with the latest base tip."""

    def repo(self, **kwargs):
        repo = CandidateRepo(**kwargs)
        self.addCleanup(repo.close)
        return repo

    def test_head_passing_alone_fails_when_merged(self):
        """AC1 (the #3062 shape): origin gained a rule the head cannot know. In place the head
        passes (the hole); the candidate run fails, and runs outside the working folder."""
        repo = self.repo()
        repo.commit(repo.repo, {"forbidden.txt": "x\n"}, "head adds forbidden.txt")
        repo.advance_origin({"rule_forbids_forbidden.txt": "r\n"}, "base adds the rule")
        proc = repo.run()
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertEqual(fields["result"], "fail")
        self.assertEqual(fields["head"], repo.head())
        self.assertEqual(fields["base"], repo.origin_tip())
        self.assertNotEqual(fields["tree"], "none")
        self.assertIn("rule: forbidden.txt must not exist", proc.stderr)
        rows = repo.where_rows()
        self.assertEqual(len(rows), 1)  # the first step failed and the run stopped
        cwd = Path(rows[0][0])
        self.assertNotIn(repo.repo, [cwd, *cwd.parents])
        self.assertEqual(cwd.parent.parent, repo.tmpdir)
        self.assertTrue(cwd.parent.name.startswith("local-check-cand."), cwd)
        self.assertEqual(rows[0][1], fields["tree"])
        self.assertFalse(cwd.exists())
        same_head_in_place = repo.run(LOCAL_CHECK_MODE="in-place")
        self.assertEqual(same_head_in_place.returncode, 0, same_head_in_place.stderr)
        self.assertRegex(last_line(same_head_in_place.stdout), LINE)
        self.assertIn(" result=pass cmd=", last_line(same_head_in_place.stdout))

    def test_a_head_that_merges_cleanly_passes_with_one_long_line(self):
        repo = self.repo()
        repo.commit(repo.repo, {"feature.txt": "f\n"}, "head")
        proc = repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(fields["result"], "pass")
        self.assertEqual(sum(1 for out in proc.stdout.splitlines()
                             if out.startswith("local-check:")), 1)
        self.assertEqual(len(repo.recorded()), 7)
        self.assertEqual(repo.leftovers(), [])

    def test_base_moving_between_runs_changes_the_base_in_the_line(self):
        """AC2 (line): the base is the tip at the time of the run, and each run has its own
        candidate folder that is gone afterwards."""
        repo = self.repo()
        repo.commit(repo.repo, {"feature.txt": "f\n"}, "head")
        first = repo.run()
        self.assertEqual(first.returncode, 0, first.stdout + first.stderr)
        one = long_fields(self, last_line(first.stdout))
        self.assertEqual(one["base"], repo.origin_tip())
        new_tip = repo.advance_origin({"base_only.txt": "b\n"}, "origin moves")
        second = repo.run()
        self.assertEqual(second.returncode, 0, second.stdout + second.stderr)
        two = long_fields(self, last_line(second.stdout))
        self.assertEqual(two["base"], new_tip)
        self.assertNotEqual(one["base"], two["base"])
        self.assertNotEqual(one["tree"], two["tree"])
        self.assertNotEqual(one["run"], two["run"])
        self.assertEqual(one["head"], two["head"])
        cwds = {row[0] for row in repo.where_rows()}
        self.assertEqual(len(cwds), 2)
        for cwd in cwds:
            self.assertFalse(Path(cwd).exists(), cwd)

    def test_conflict_with_the_new_base_tip_is_a_fail_with_base_and_no_tree(self):
        """AC2 (conflict): base known, tree=none, exit 1, nothing ran, no folder left."""
        repo = self.repo()
        repo.commit(repo.repo, {"README.md": "one\ntwo-feature\n"}, "head edits line 2")
        repo.advance_origin({"README.md": "one\ntwo-origin\n"}, "origin edits line 2")
        proc = repo.run()
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertEqual((fields["result"], fields["tree"]), ("fail", "none"))
        self.assertEqual(fields["base"], repo.origin_tip())
        self.assertIn("conflicts with origin/" + BASE_BRANCH, proc.stderr)
        self.assertEqual(repo.recorded(), [])
        self.assertEqual(repo.leftovers(), [])
        self.assertEqual(len(repo.worktrees()), 1)

    def test_unrelated_history_is_a_fail_with_base_and_no_tree(self):
        repo = self.repo()
        repo.commit(repo.repo, {"feature.txt": "f\n"}, "head")
        root = repo.replace_origin_history({"only.txt": "o\n"})
        proc = repo.run()
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertEqual((fields["result"], fields["base"], fields["tree"]),
                         ("fail", root, "none"))
        self.assertEqual(repo.recorded(), [])
        self.assertEqual(repo.leftovers(), [])

    def test_line_names_the_base_and_tree_actually_tested(self):
        """AC3: base is the fetched tip, tree is the tree of an independent merge-tree, and
        the stubs saw that tree, base and head in the folder they ran in."""
        repo = self.repo()
        head = repo.commit(repo.repo, {"feature.txt": "f\n"}, "head")
        repo.advance_origin({"base_only.txt": "b\n"}, "origin moves")
        proc = repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(fields["head"], head)
        self.assertEqual(fields["base"], repo.origin_tip())
        self.assertEqual(fields["base"], repo.local_origin_ref())
        independent = repo.git(repo.repo, "merge-tree", "--write-tree",
                               fields["base"], head).splitlines()[0]
        self.assertEqual(fields["tree"], independent)
        rows = repo.where_rows()
        self.assertEqual(len(rows), 7)
        for _, tree, parent1, parent2 in rows:
            self.assertEqual((tree, parent1, parent2), (fields["tree"], fields["base"], head))
        self.assertEqual(fields["script"],
                         repo.git(repo.repo, "hash-object", "tools/local_check.sh"))
        self.assertEqual(fields["rustc"], "1.99.0")

    def test_run_leaves_working_folder_index_and_branches_unchanged(self):
        """AC5: folder, index, HEAD, branches, stash, worktree list and ignored files are
        identical after a pass, a conflict, a failing step and a failed fetch. Only
        refs/remotes/origin/<base> may move, and then to origin's tip."""
        for scenario in ("pass", "conflict", "step-failure", "fetch-failure"):
            with self.subTest(scenario=scenario):
                repo = self.repo(fail_label="clippy" if scenario == "step-failure" else None)
                edit = "two-feature\n" if scenario == "conflict" else "two\nthree\n"
                repo.commit(repo.repo, {"README.md": f"one\n{edit}"}, "head")
                repo.advance_origin({"README.md": "one\ntwo-origin\n" if scenario == "conflict"
                                     else "one\ntwo\n", "base_only.txt": "b\n"}, "origin moves")
                (repo.repo / "kept.log").write_text("ignored\n")
                (repo.repo / "stashed.txt").write_text("s\n")
                repo.git(repo.repo, "stash", "push", "-q", "-u", "-m", "keep me")
                repo.git(repo.repo, "branch", "other", "feature")
                if scenario == "fetch-failure":
                    repo.git(repo.repo, "remote", "set-url", "origin", str(repo.tmp / "gone.git"))
                before = repo.snapshot()
                stale = repo.local_origin_ref()
                proc = repo.run()
                after = repo.snapshot()
                self.assertEqual(after, before)
                self.assertEqual(proc.returncode, 0 if scenario == "pass" else 1,
                                 proc.stdout + proc.stderr)
                long_fields(self, last_line(proc.stdout))
                self.assertEqual(repo.leftovers(), [])
                if scenario == "fetch-failure":
                    self.assertEqual(repo.local_origin_ref(), stale)
                else:
                    self.assertEqual(repo.local_origin_ref(), repo.origin_tip())
                    self.assertNotEqual(stale, repo.origin_tip())

    def test_a_dirty_working_folder_fails_without_running_anything(self):
        repo = self.repo()
        (repo.repo / "new.txt").write_text("x\n")
        proc = repo.run()
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1)
        self.assertEqual((fields["result"], fields["base"], fields["tree"]),
                         ("fail", "none", "none"))
        self.assertNotIn("dirty=1", last_line(proc.stdout))  # the gate rejects such a line
        self.assertEqual(repo.recorded(), [])
        self.assertEqual(repo.leftovers(), [])

    def test_a_failed_fetch_never_falls_back_to_the_stale_local_ref(self):
        """No pass line with an old base: result=fail base=none tree=none, nothing ran."""
        repo = self.repo()
        repo.git(repo.repo, "remote", "set-url", "origin", str(repo.tmp / "gone.git"))
        self.assertRegex(repo.local_origin_ref(), r"^[0-9a-f]{40}$")  # the stale ref exists
        proc = repo.run()
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1)
        self.assertEqual((fields["result"], fields["base"], fields["tree"]),
                         ("fail", "none", "none"))
        self.assertIn("fetching origin/" + BASE_BRANCH + " failed", proc.stderr)
        self.assertEqual(repo.recorded(), [])
        self.assertEqual(repo.leftovers(), [])

    def test_base_branch_name_boundaries(self):
        repo = self.repo()
        repo.commit(repo.repo, {"feature.txt": "f\n"}, "head")
        for name in ("-x", "--upload-pack=x", "a..b", "a b", "a;b", "a$b", "a`b", "a*b", "a:b"):
            with self.subTest(rejected=name):
                proc = repo.run(LOCAL_CHECK_BASE_BRANCH=name)
                fields = long_fields(self, last_line(proc.stdout))
                self.assertEqual(proc.returncode, 1)
                self.assertEqual((fields["result"], fields["base"]), ("fail", "none"))
                self.assertIn("is not a branch name", proc.stderr)
                self.assertEqual(repo.recorded(), [])
        for name in ("release/2026.11_x-1", "A.b_c-d/e"):
            with self.subTest(accepted_but_missing=name):
                proc = repo.run(LOCAL_CHECK_BASE_BRANCH=name)
                long_fields(self, last_line(proc.stdout))
                self.assertEqual(proc.returncode, 1)
                self.assertNotIn("is not a branch name", proc.stderr)
                self.assertIn("failed", proc.stderr)
        tip = repo.advance_origin({"rel.txt": "r\n"}, "release branch", branch="release/2026.11")
        proc = repo.run(LOCAL_CHECK_BASE_BRANCH="release/2026.11")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(long_fields(self, last_line(proc.stdout))["base"], tip)

    def test_unknown_mode_exits_2_without_a_line(self):
        repo = self.repo()
        for mode in ("in_place", "In-Place", "candidate ", "x"):
            with self.subTest(mode=mode):
                proc = repo.run(LOCAL_CHECK_MODE=mode)
                self.assertEqual(proc.returncode, 2)
                self.assertNotIn("local-check:", proc.stdout)
                self.assertIn("LOCAL_CHECK_MODE must be candidate or in-place", proc.stderr)
        self.assertEqual(repo.recorded(), [])

    def test_candidate_mode_ignores_the_old_base_overrides(self):
        """An old base given by LOCAL_CHECK_BASE must not be able to make a pass line."""
        repo = self.repo()
        repo.commit(repo.repo, {"feature.txt": "f\n"}, "head")
        proc = repo.run(LOCAL_CHECK_BASE="HEAD~1", LOCAL_CHECK_DIFF_HEAD="HEAD~1")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn("ignores LOCAL_CHECK_BASE and LOCAL_CHECK_DIFF_HEAD", proc.stderr)
        self.assertEqual(long_fields(self, last_line(proc.stdout))["base"], repo.origin_tip())

    def test_the_old_base_overrides_cannot_hide_a_whitespace_error_in_the_diff_check(self):
        """An override that made the range empty (base = head) would turn the diff-check
        green; the candidate mode still checks the pinned base...head."""
        repo = self.repo()
        head = repo.commit(repo.repo, {"README.md": "one \ntwo\n"}, "trailing space")
        proc = repo.run(LOCAL_CHECK_BASE=head, LOCAL_CHECK_DIFF_HEAD=repo.origin_tip())
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertEqual(fields["result"], "fail")
        self.assertIn("step diff-check: exit", proc.stderr)

    def test_rustc_token_boundaries(self):
        """32 characters pass, 33 and other shapes (or no version) are a fail."""
        repo = self.repo()
        cases = [("rustc " + "a" * 32 + " (x 2026-01-01)", "a" * 32),
                 ("rustc 1.2.3-nightly+x_y (z)", "1.2.3-nightly+x_y"),
                 ("rustc " + "a" * 33 + " (x)", None),
                 ("rustc 1/2 (x)", None),
                 ("rustc", None),
                 ("", None)]
        for out, token in cases:
            with self.subTest(out=out):
                proc = repo.run(STUB_RUSTC_OUT=out)
                fields = long_fields(self, last_line(proc.stdout))
                if token is not None:
                    self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
                    self.assertEqual(fields["rustc"], token)
                else:
                    self.assertEqual(proc.returncode, 1)
                    self.assertEqual((fields["result"], fields["rustc"]), ("fail", "unknown"))
                    self.assertNotEqual(fields["tree"], "none")  # the candidate was built
                self.assertEqual(repo.leftovers(), [])
        no_rustc = repo.run(STUB_RUSTC_OUT="rustc 1.0.0", STUB_RUSTC_EXIT="1")
        self.assertEqual(no_rustc.returncode, 1)
        self.assertIn(" result=fail ", last_line(no_rustc.stdout))
        self.assertIn(" rustc=unknown ", last_line(no_rustc.stdout))

    def test_a_tmpdir_with_a_space_works(self):
        repo = self.repo()
        spaced = repo.tmp / "with space"
        spaced.mkdir()
        proc = repo.run(TMPDIR=str(spaced))
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        long_fields(self, last_line(proc.stdout))
        for row in repo.where_rows():
            self.assertTrue(row[0].startswith(str(spaced) + "/local-check-cand."), row[0])
        self.assertEqual(list(spaced.iterdir()), [])

    def test_an_earlier_candidate_folder_is_reported_and_not_deleted(self):
        repo = self.repo()
        old = repo.tmpdir / "local-check-cand.OLD"
        old.mkdir()
        (old / "keep.txt").write_text("k\n")
        proc = repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn(f"an earlier candidate folder exists: '{old}'", proc.stderr)
        self.assertEqual(repo.leftovers(), ["local-check-cand.OLD"])
        self.assertEqual((old / "keep.txt").read_text(), "k\n")

    def test_dry_run_checks_nothing_fetches_nothing_and_says_fail(self):
        repo = self.repo()
        repo.advance_origin({"base_only.txt": "b\n"}, "origin moves")
        stale = repo.local_origin_ref()
        proc = repo.run(LOCAL_CHECK_DRY_RUN="1")
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1)
        self.assertEqual((fields["result"], fields["base"], fields["tree"]),
                         ("fail", "none", "none"))
        steps = [STEP.match(line).groups() for line in proc.stdout.splitlines()
                 if STEP.match(line)]
        self.assertEqual([s[0] for s in steps], [label for label, _ in CI_LINUX_STEPS])
        self.assertEqual(steps[-1][1], "git diff --check <base>...<head>")
        self.assertEqual(repo.recorded(), [])
        self.assertEqual(repo.local_origin_ref(), stale)

    def test_candidate_mode_runs_the_same_steps_as_ci(self):
        """AC6: the same step list and order in the candidate; diff-check reads base...head."""
        repo = self.repo()
        head = repo.commit(repo.repo, {"feature.txt": "f\n"}, "head")
        proc = repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        steps = [STEP.match(line).groups() for line in proc.stdout.splitlines()
                 if STEP.match(line)]
        base = repo.origin_tip()
        expected = [(label, cmd.replace("BASE...HEAD", f"{base}...{head}"))
                    for label, cmd in CI_LINUX_STEPS]
        self.assertEqual(steps, expected)
        self.assertEqual([c.split()[0] + " " + c.split()[1] for c in repo.recorded()],
                         ["cargo fetch", "cargo nextest", "cargo test", "cargo clippy",
                          "cargo fmt", "rustfmt --check", "python3 -m"])

    def test_step_failure_keeps_real_evidence_and_cleans_up(self):
        repo = self.repo(fail_label="clippy")
        summary = repo.tmp / "summary.md"
        proc = repo.run(LOCAL_CHECK_KEEP_GOING="1", GITHUB_STEP_SUMMARY=str(summary))
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1)
        self.assertEqual(fields["result"], "fail")
        self.assertEqual(fields["base"], repo.origin_tip())
        self.assertNotEqual(fields["tree"], "none")
        self.assertEqual(len(repo.recorded()), 7)
        self.assertEqual(len(summary.read_text().splitlines()), 8)
        self.assertEqual(repo.leftovers(), [])
        self.assertEqual(len(repo.worktrees()), 1)

    def test_diff_check_in_the_candidate_catches_a_whitespace_error_in_the_head(self):
        repo = self.repo()
        repo.commit(repo.repo, {"README.md": "one \ntwo\n"}, "trailing space")
        proc = repo.run()
        fields = long_fields(self, last_line(proc.stdout))
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertEqual(fields["result"], "fail")
        self.assertIn("step diff-check: exit", proc.stderr)

    def test_the_line_regex_pins_the_field_widths(self):
        """The shape the gate reads: 39/41 hex, uppercase, non-hex, 33 rustc, 11/13 run and a
        dirty field are all outside it."""
        repo = self.repo()
        proc = repo.run()
        line = last_line(proc.stdout)
        fields = long_fields(self, line)
        self.assertEqual(len(fields["base"]), 40)
        for mutant in (
            line.replace(fields["base"], fields["base"][:39], 1),
            line.replace(fields["base"], fields["base"] + "0", 1),
            line.replace(fields["tree"], fields["tree"][:39], 1),
            line.replace(fields["tree"], fields["tree"] + "0", 1),
            line.replace(fields["tree"], fields["tree"].upper(), 1),
            line.replace(fields["base"], "g" + fields["base"][1:], 1),
            line.replace("rustc=1.99.0", "rustc=" + "a" * 33, 1),
            line.replace(fields["run"], fields["run"][:11], 1),
            line.replace(fields["run"], fields["run"] + "0", 1),
            line.replace(" base=", " dirty=1 base=", 1),
            line + " ",
            " " + line,
        ):
            with self.subTest(mutant=mutant):
                self.assertIsNone(LONG.match(mutant))


if __name__ == "__main__":
    unittest.main()
