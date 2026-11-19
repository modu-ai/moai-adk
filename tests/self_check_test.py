"""tools/self_check.sh (card C-77): the machine items of the pre-PR self-check from one command.

The script runs the machine-checkable items of skills/moai/references/internal/self-check.md
in the working folder, keeps going past a failure, prints one `self-check: <item> ...` line per
item with its exit code and ends with one summary line. These tests pin the item list and its
commands (whole `--bin` runs, never filtered), the keep-going rule, the skip and not_run rules,
the heavy-slot refusal (exit 75: not run, not failed), the package file bound at 65,536 / 65,537
bytes, the diff check, the Python files picked from the change, and the summary line. Real runs
use stub `cargo`, `rustfmt`, `python3` and `moai` in a temporary git repository with a
temporary HOME, so no cargo build runs here; `git` is real.
"""
import os
import re
import shutil
import subprocess
import tempfile
import time
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "tools" / "self_check.sh"
REFERENCE = ROOT / "skills" / "moai" / "references" / "internal" / "self-check.md"
PROFILE = ROOT / "src" / "profile.rs"
HEAVY_SLOT = ROOT / "src" / "heavy_slot.rs"

# The items in order with the exact command each runs through a stub (package-size is built
# in; diff-check runs the real git with the merge base in its arguments).
ITEMS = [
    ("bin-moai", "cargo test --offline --locked --bin moai"),
    ("bin-moai-cli", "cargo test --offline --locked --bin moai-cli"),
    ("clippy", "cargo clippy --offline --locked --all-targets -- -D warnings"),
    ("fmt", "cargo fmt --check"),
    ("fmt-ops", "rustfmt --check --edition 2021 src/ops/a.rs src/ops/b.rs"),
    ("git-env-scrub", "cargo test --offline --locked --test git_env_scrub"),
    ("home-isolation", "cargo test --offline --locked --test home_isolation_scan"),
    ("ac-command-lint", "cargo test --offline --locked --test ac_command_lint"),
    ("package-inspect",
     "cargo run --offline --locked --quiet --bin moai -- package-inspect skills/moai"),
    ("package-size", None),
    ("manifest", "python3 -I tools/source_manifest.py check"),
    ("diff-check", None),
    ("python", "python3 -m unittest tests/skill_references_test.py"),
]
LABELS = [label for label, _ in ITEMS]
COMMANDS = [cmd for _, cmd in ITEMS if cmd]
HEAVY = ["bin-moai", "bin-moai-cli", "clippy", "git-env-scrub", "home-isolation",
         "ac-command-lint"]
ITEM = re.compile(r"^self-check: (?P<label>[a-z-]+) (?P<rest>.*)$")
LAST = re.compile(
    r"^self-check: head=(?P<head>[0-9a-f]{40}|unknown) result=(?P<result>pass|fail|not_run) "
    r"(?P<dirty>dirty=1 )?(?P<dry>dry-run=1 )?failed=(?P<failed>none|[a-z,-]+) "
    r"not_run=(?P<not_run>none|[a-z,-]+) logs=(?P<logs>\S+)$")


class FakeRepo:
    """A temporary git repository with a copy of the script, the files each item needs, a base
    commit and one change on top, and stub `cargo` / `rustfmt` / `python3` on PATH that record
    their arguments. With slot=True a stub `moai` offers `heavy-slot-acquire`: it records its
    arguments and runs the command, or exits 75 when STUB_SLOT_REFUSE is set."""

    def __init__(self, slot=False):
        self.tmp = Path(tempfile.mkdtemp(prefix="self-check-"))
        self.repo = self.tmp / "repo"
        (self.repo / "tools").mkdir(parents=True)
        shutil.copy2(SCRIPT, self.repo / "tools" / "self_check.sh")
        (self.repo / "tools" / "source_manifest.py").write_text("# stub\n")
        (self.repo / "Cargo.toml").write_text("[package]\nname = \"x\"\n")
        (self.repo / "tests").mkdir()
        for name in ("git_env_scrub", "home_isolation_scan", "ac_command_lint"):
            (self.repo / "tests" / f"{name}.rs").write_text("\n")
        (self.repo / "tests" / "skill_references_test.py").write_text("# stub\n")
        (self.repo / "src" / "ops").mkdir(parents=True)
        for name in ("b", "a"):
            (self.repo / "src" / "ops" / f"{name}.rs").write_text("\n")
        (self.repo / "skills" / "moai").mkdir(parents=True)
        (self.repo / "skills" / "moai" / "SKILL.md").write_text("# moai\n")
        (self.repo / ".gitignore").write_text("target/\n")
        self.calls = self.tmp / "calls"
        self.slot_calls = self.tmp / "slot_calls"
        self.bin = self.tmp / "bin"
        self.bin.mkdir()
        for name in ("cargo", "rustfmt", "python3"):
            stub = self.bin / name
            stub.write_text(
                "#!/bin/sh\n"
                f"echo \"{name} $*\" >> '{self.calls}'\n"
                f"echo \"stub output of {name} $*\"\n"
                "case \"$*\" in *\"${STUB_FAIL:-@none@}\"*) exit 3;; esac\n"
                "exit 0\n")
            stub.chmod(0o755)
        if slot:
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
        self.git("init", "-q")
        self.git("add", "-A")
        self.git("commit", "-q", "-m", "base")
        self.base = self.git("rev-parse", "HEAD")
        (self.repo / "README.md").write_text("change\n")
        self.git("add", "-A")
        self.git("commit", "-q", "-m", "change")

    def env(self, **extra):
        env = {"PATH": f"{self.bin}:/usr/bin:/bin", "HOME": str(self.tmp), "LC_ALL": "C",
               "GIT_CONFIG_NOSYSTEM": "1", "GIT_AUTHOR_NAME": "t",
               "GIT_AUTHOR_EMAIL": "t@example.invalid", "GIT_COMMITTER_NAME": "t",
               "GIT_COMMITTER_EMAIL": "t@example.invalid"}
        env.update(extra)
        # A value of None removes the variable (an unset HOME is not an empty one).
        return {k: v for k, v in env.items() if v is not None}

    def git(self, *args):
        return subprocess.run(["git", *args], cwd=self.repo, env=self.env(), check=True,
                              capture_output=True, text=True).stdout.strip()

    def commit(self, message="more"):
        self.git("add", "-A")
        self.git("commit", "-q", "-m", message)

    def run(self, *args, **extra):
        # Started from outside the repository: the script must find its own root.
        extra.setdefault("SELF_CHECK_BASE", self.base)
        return subprocess.run(["bash", str(self.repo / "tools" / "self_check.sh"), *args],
                              cwd=self.tmp, env=self.env(**extra), capture_output=True,
                              text=True, timeout=60)

    def recorded(self):
        return self.calls.read_text().splitlines() if self.calls.exists() else []

    def slot_recorded(self):
        return self.slot_calls.read_text().splitlines() if self.slot_calls.exists() else []

    def close(self):
        shutil.rmtree(self.tmp, ignore_errors=True)


def items(stdout):
    """[(label, rest)] of the item lines, in order, and the last line."""
    lines = [line for line in stdout.splitlines() if line.strip()]
    found = []
    for line in lines[:-1]:
        m = ITEM.match(line)
        if m:
            found.append((m["label"], m["rest"]))
    return found, (lines[-1] if lines else "")


class SelfCheckTest(unittest.TestCase):
    def setUp(self):
        self.repo = FakeRepo()
        self.addCleanup(self.repo.close)

    def test_every_item_runs_in_order_with_one_line_and_the_run_passes(self):
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], LABELS)
        for label, rest in found:
            self.assertTrue(rest.startswith("exit=0 "), f"{label}: {rest}")
        self.assertEqual(self.repo.recorded(), COMMANDS)
        self.assertIn(f"cmd=git diff --check {self.repo.base}", dict(found)["diff-check"])
        m = LAST.match(last)
        self.assertIsNotNone(m, last)
        self.assertEqual(m["head"], self.repo.git("rev-parse", "HEAD"))
        self.assertEqual((m["result"], m["dirty"], m["dry"], m["failed"], m["not_run"]),
                         ("pass", None, None, "none", "none"))
        # Each item's output goes to its own log, not to the screen.
        self.assertNotIn("stub output", proc.stdout)
        log = self.repo.repo / "target" / "self-check" / "clippy.log"
        self.assertIn("stub output of cargo clippy", log.read_text())

    def test_whole_bin_runs_carry_no_test_name_filter(self):
        bins = [cmd for cmd in COMMANDS if " --bin moai" in cmd and cmd.startswith("cargo test")]
        self.assertEqual(bins, ["cargo test --offline --locked --bin moai",
                                "cargo test --offline --locked --bin moai-cli"])
        self.repo.run()
        recorded = self.repo.recorded()
        for cmd in bins:
            self.assertIn(cmd, recorded)

    def test_a_failing_item_is_reported_and_every_later_item_still_runs(self):
        proc = self.repo.run(STUB_FAIL="clippy")
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], LABELS)
        rest = dict(found)
        self.assertTrue(rest["clippy"].startswith("exit=3 "), rest["clippy"])
        self.assertTrue(rest["python"].startswith("exit=0 "), rest["python"])
        self.assertEqual(self.repo.recorded(), COMMANDS)
        m = LAST.match(last)
        self.assertEqual((m["result"], m["failed"], m["not_run"]), ("fail", "clippy", "none"))

    def test_two_failures_are_both_named(self):
        proc = self.repo.run(STUB_FAIL="--bin moai")
        self.assertEqual(proc.returncode, 1)
        m = LAST.match(items(proc.stdout)[1])
        # `--bin moai` matches bin-moai, bin-moai-cli and package-inspect.
        self.assertEqual(m["failed"], "bin-moai,bin-moai-cli,package-inspect")

    def test_package_file_bound_passes_at_65536_and_fails_at_65537_bytes(self):
        big = self.repo.repo / "skills" / "moai" / "workflows" / "big.md"
        big.parent.mkdir()
        big.write_bytes(b"x" * 65536)
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout)
        rest = dict(items(proc.stdout)[0])["package-size"]
        self.assertTrue(rest.startswith("exit=0 limit=65536 largest=65536 "), rest)
        self.assertIn(" over=0 ", rest)
        big.write_bytes(b"x" * 65537)
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 1, proc.stdout)
        found, last = items(proc.stdout)
        rest = dict(found)["package-size"]
        self.assertTrue(rest.startswith("exit=1 limit=65536 largest=65537 "), rest)
        self.assertIn(" over=1 ", rest)
        log = self.repo.repo / "target" / "self-check" / "package-size.log"
        # Built at run time: a literal fixture path would read as a broken skill reference to
        # tests/skill_references_test.py.
        self.assertIn(big.relative_to(self.repo.repo).as_posix(), log.read_text())
        self.assertEqual(LAST.match(last)["failed"], "package-size")

    def test_an_empty_package_folder_is_a_skip_not_a_pass(self):
        shutil.rmtree(self.repo.repo / "skills")
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout)
        rest = dict(items(proc.stdout)[0])
        for label in ("package-inspect", "package-size"):
            self.assertEqual(rest[label], "skip (no skills/moai/SKILL.md)")
        self.assertNotIn("package-inspect", " ".join(self.repo.recorded()))

    def test_an_item_whose_input_is_missing_is_skipped_with_the_reason(self):
        (self.repo.repo / "tests" / "git_env_scrub.rs").unlink()
        (self.repo.repo / "tools" / "source_manifest.py").unlink()
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout)
        rest = dict(items(proc.stdout)[0])
        self.assertEqual(rest["git-env-scrub"], "skip (no tests/git_env_scrub.rs)")
        self.assertEqual(rest["manifest"], "skip (no tools/source_manifest.py)")
        self.assertNotIn("--test git_env_scrub", " ".join(self.repo.recorded()))

    def test_without_cargo_toml_every_cargo_item_is_skipped(self):
        (self.repo.repo / "Cargo.toml").unlink()
        proc = self.repo.run()
        rest = dict(items(proc.stdout)[0])
        for label, cmd in ITEMS:
            if cmd and cmd.startswith("cargo "):
                self.assertEqual(rest[label], "skip (no Cargo.toml)", label)
        self.assertEqual([c for c in self.repo.recorded() if c.startswith("cargo")], [])

    def test_dry_run_runs_nothing_and_never_passes(self):
        proc = self.repo.run(SELF_CHECK_DRY_RUN="1")
        self.assertEqual(proc.returncode, 1)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], LABELS)
        self.assertEqual(self.repo.recorded(), [])
        m = LAST.match(last)
        self.assertEqual((m["result"], m["dry"]), ("fail", "dry-run=1 "))

    def test_an_uncommitted_change_is_flagged_on_the_last_line(self):
        (self.repo.repo / "new.txt").write_text("x\n")
        proc = self.repo.run()
        m = LAST.match(items(proc.stdout)[1])
        self.assertEqual(m["dirty"], "dirty=1 ")
        # The folder was checked; the line says it is not the committed head alone.
        self.assertEqual(proc.returncode, 0)

    def test_an_unknown_argument_is_a_usage_error(self):
        for args in (["--bin"], ["--full-python", "x"], ["--full-python=1"]):
            proc = self.repo.run(*args)
            self.assertEqual(proc.returncode, 2, args)
            self.assertIn("usage", proc.stderr)
        self.assertEqual(self.repo.recorded(), [])


class FmtOpsTest(unittest.TestCase):
    """N2: `cargo fmt` does not see src/ops/*.rs, so rustfmt checks them by name."""

    def setUp(self):
        self.repo = FakeRepo()
        self.addCleanup(self.repo.close)

    def test_every_operation_file_is_checked_in_name_order(self):
        proc = self.repo.run()
        self.assertIn("rustfmt --check --edition 2021 src/ops/a.rs src/ops/b.rs",
                      self.repo.recorded())
        self.assertTrue(dict(items(proc.stdout)[0])["fmt-ops"].startswith("exit=0 "))

    def test_a_badly_formatted_operation_file_fails_the_run(self):
        proc = self.repo.run(STUB_FAIL="src/ops/b.rs")
        found, last = items(proc.stdout)
        self.assertTrue(dict(found)["fmt-ops"].startswith("exit=3 "))
        self.assertEqual(LAST.match(last)["failed"], "fmt-ops")
        self.assertEqual(proc.returncode, 1)

    def test_no_operation_folder_is_a_skip(self):
        shutil.rmtree(self.repo.repo / "src")
        proc = self.repo.run()
        self.assertEqual(dict(items(proc.stdout)[0])["fmt-ops"], "skip (no src/ops/*.rs)")
        self.assertNotIn("rustfmt", " ".join(self.repo.recorded()))


class DiffCheckTest(unittest.TestCase):
    """N1: `git diff --check` from the merge base, committed and uncommitted changes alike."""

    def setUp(self):
        self.repo = FakeRepo()
        self.addCleanup(self.repo.close)

    def diff_line(self, proc):
        found, last = items(proc.stdout)
        return dict(found)["diff-check"], LAST.match(last)

    def test_a_committed_trailing_space_fails(self):
        (self.repo.repo / "README.md").write_text("change \n")
        self.repo.commit()
        rest, last = self.diff_line(self.repo.run())
        self.assertFalse(rest.startswith("exit=0 "), rest)
        self.assertEqual(last["failed"], "diff-check")

    def test_an_uncommitted_trailing_space_fails(self):
        (self.repo.repo / "README.md").write_text("change \n")
        rest, last = self.diff_line(self.repo.run())
        self.assertFalse(rest.startswith("exit=0 "), rest)
        self.assertEqual(last["failed"], "diff-check")

    def test_a_trailing_space_already_in_the_base_does_not_count(self):
        (self.repo.repo / "README.md").write_text("change \n")
        self.repo.commit()
        self.repo.base = self.repo.git("rev-parse", "HEAD")
        rest, last = self.diff_line(self.repo.run())
        self.assertTrue(rest.startswith("exit=0 "), rest)

    def test_a_base_that_does_not_resolve_is_not_run_never_a_pass(self):
        proc = self.repo.run(SELF_CHECK_BASE="no-such-branch")
        found, last = items(proc.stdout)
        rest = dict(found)
        for label in ("diff-check", "python"):
            self.assertTrue(rest[label].startswith("not_run (no base: 'no-such-branch'"),
                            rest[label])
        m = LAST.match(last)
        self.assertEqual((m["result"], m["failed"], m["not_run"]),
                         ("not_run", "none", "diff-check,python"))
        self.assertEqual(proc.returncode, 1)

    def test_an_option_shaped_base_is_never_passed_to_git(self):
        proc = self.repo.run(SELF_CHECK_BASE="--all")
        self.assertEqual(LAST.match(items(proc.stdout)[1])["not_run"], "diff-check,python")


class PythonItemTest(unittest.TestCase):
    """N1: the unittest files the change touched plus the reference scan; --full-python runs
    CI's whole discover instead."""

    def setUp(self):
        self.repo = FakeRepo()
        self.addCleanup(self.repo.close)

    def python_calls(self):
        return [c for c in self.repo.recorded() if c.startswith("python3 -m unittest")]

    def test_changed_tests_and_the_tests_of_changed_tools_are_run(self):
        tests = self.repo.repo / "tests"
        tools = self.repo.repo / "tools"
        (tests / "zeta_test.py").write_text("# committed change\n")
        (tests / "bar_test.py").write_text("# test of a changed tool\n")
        self.repo.git("add", "-A")
        self.repo.git("commit", "-q", "-m", "tests")
        self.repo.base = self.repo.git("rev-parse", "HEAD")
        (tools / "bar.sh").write_text("# changed tool\n")
        (tools / "lonely.py").write_text("# a tool without a test\n")
        (tests / "helper.py").write_text("# not a test file\n")
        self.repo.commit()
        (tests / "zeta_test.py").write_text("# uncommitted change\n")
        (tests / "alpha_test.py").write_text("# untracked new test\n")
        proc = self.repo.run()
        self.assertEqual(self.python_calls(), [
            "python3 -m unittest tests/skill_references_test.py tests/alpha_test.py "
            "tests/bar_test.py tests/zeta_test.py"])
        self.assertTrue(dict(items(proc.stdout)[0])["python"].startswith("exit=0 "))

    def test_a_deleted_test_file_is_not_run(self):
        (self.repo.repo / "tests" / "gone_test.py").write_text("# x\n")
        self.repo.commit()
        self.repo.base = self.repo.git("rev-parse", "HEAD")
        (self.repo.repo / "tests" / "gone_test.py").unlink()
        self.repo.commit()
        self.repo.run()
        self.assertEqual(self.python_calls(),
                         ["python3 -m unittest tests/skill_references_test.py"])

    def test_the_reference_scan_is_listed_once_when_it_changed_too(self):
        (self.repo.repo / "tests" / "skill_references_test.py").write_text("# changed\n")
        self.repo.run()
        self.assertEqual(self.python_calls(),
                         ["python3 -m unittest tests/skill_references_test.py"])

    def test_nothing_to_run_is_a_skip(self):
        (self.repo.repo / "tests" / "skill_references_test.py").unlink()
        self.repo.commit()
        self.repo.base = self.repo.git("rev-parse", "HEAD")
        proc = self.repo.run()
        self.assertEqual(dict(items(proc.stdout)[0])["python"],
                         "skip (no changed Python test and no tests/skill_references_test.py)")
        self.assertEqual(self.python_calls(), [])

    def test_full_python_runs_ci_discover_and_only_that(self):
        proc = self.repo.run("--full-python")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(self.python_calls(),
                         ["python3 -m unittest discover -s tests -p *_test.py"])
        # It needs no base: CI's discover does not depend on the change.
        self.repo.calls.unlink()
        proc = self.repo.run("--full-python", SELF_CHECK_BASE="no-such-branch")
        self.assertEqual(self.python_calls(),
                         ["python3 -m unittest discover -s tests -p *_test.py"])
        self.assertEqual(LAST.match(items(proc.stdout)[1])["not_run"], "diff-check")


class HeavySlotTest(unittest.TestCase):
    """N3: the heavy cargo items run in a heavy slot under a per-run name; the slot's own
    refusal (exit 75) is not run, not a failure."""

    def setUp(self):
        self.repo = FakeRepo(slot=True)
        self.addCleanup(self.repo.close)

    def test_heavy_items_run_in_the_slot_under_a_per_run_name(self):
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        slot = self.repo.slot_recorded()
        self.assertEqual(len(slot), len(HEAVY))
        names = {re.search(r"--name (\S+) --", line)[1] for line in slot}
        self.assertEqual(len(names), 1, names)
        name = names.pop()
        self.assertRegex(name, r"^self-check-repo-\d+$")
        # Light items never take a slot; the heavy ones still ran their command.
        self.assertFalse(any("cargo fmt" in line or "package-inspect" in line for line in slot))
        self.assertEqual(self.repo.recorded(), COMMANDS)
        # Two runs, two names: runs never share the queue name `self-check`.
        self.repo.slot_calls.unlink()
        self.repo.run()
        other = re.search(r"--name (\S+) --", self.repo.slot_recorded()[0])[1]
        self.assertNotEqual(other, name)

    def test_the_name_can_be_set(self):
        self.repo.run(MOAI_HEAVY_SLOT_NAME="lane-7")
        for line in self.repo.slot_recorded():
            self.assertIn("--name lane-7 --", line)

    def test_a_slot_refusal_is_not_run_and_not_a_failure(self):
        proc = self.repo.run(STUB_SLOT_REFUSE="1")
        self.assertEqual(proc.returncode, 1)
        found, last = items(proc.stdout)
        rest = dict(found)
        for label in HEAVY:
            self.assertEqual(rest[label],
                             f"exit=75 slot-refused (not run) log=target/self-check/{label}.log")
        self.assertTrue(rest["fmt"].startswith("exit=0 "))
        m = LAST.match(last)
        self.assertEqual((m["result"], m["failed"], m["not_run"]),
                         ("not_run", "none", ",".join(HEAVY)))
        # No heavy command ran; the light ones did.
        self.assertFalse(any(c.startswith("cargo test") for c in self.repo.recorded()))

    def test_a_failure_beside_a_refusal_is_still_a_fail(self):
        # The stubs match their arguments only: `fmt --check` is the fmt item.
        proc = self.repo.run(STUB_SLOT_REFUSE="1", STUB_FAIL="fmt --check")
        m = LAST.match(items(proc.stdout)[1])
        self.assertEqual((m["result"], m["failed"]), ("fail", "fmt"))
        self.assertEqual(proc.returncode, 1)

    def test_a_command_failure_inside_the_slot_is_a_failure(self):
        proc = self.repo.run(STUB_FAIL="clippy")
        found, last = items(proc.stdout)
        self.assertTrue(dict(found)["clippy"].startswith("exit=3 "))
        self.assertEqual(LAST.match(last)["failed"], "clippy")


class PythonSlotTest(unittest.TestCase):
    """SPEC-20261119-python-slot-pre-ready: a whole `unittest discover` waits in a python slot,
    a slot folder apart from the cargo slots; the touched-files item stays out of any slot."""

    def setUp(self):
        self.repo = FakeRepo(slot=True)
        self.addCleanup(self.repo.close)
        self.default_dir = self.repo.tmp / ".cache" / "moai-cli" / "python-slots"

    def python_lines(self):
        return [l for l in self.repo.slot_recorded() if " --dir " in l]

    def test_full_python_runs_in_the_python_slot_with_the_default_folder_and_limit(self):
        proc = self.repo.run("--full-python")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        lines = self.python_lines()
        self.assertEqual(len(lines), 1, lines)
        self.assertRegex(lines[0], (
            rf"^heavy-slot-acquire --dir {re.escape(str(self.default_dir))} --max 2 "
            r"--min-free-gb 50 --name self-check-repo-\d+-py -- "
            r"python3 -m unittest discover -s tests -p \*_test\.py$"))
        # The cargo items still use the heavy slot, without a folder of their own.
        self.assertEqual(len(self.repo.slot_recorded()) - 1, len(HEAVY))

    def test_the_touched_files_item_takes_no_slot(self):
        self.repo.run()
        self.assertEqual(self.python_lines(), [])

    def test_xdg_cache_home_and_the_two_settings_change_folder_and_limit(self):
        xdg = self.repo.tmp / "xdg"
        self.repo.run("--full-python", XDG_CACHE_HOME=str(xdg))
        self.assertIn(f" --dir {xdg}/moai-cli/python-slots ", self.python_lines()[0])
        self.repo.slot_calls.unlink()
        self.repo.run("--full-python", MOAI_PYTHON_SLOT_DIR="/x/y", MOAI_PYTHON_SLOT_MAX="1")
        self.assertIn(" --dir /x/y --max 1 ", self.python_lines()[0])

    def test_empty_settings_mean_the_defaults(self):
        self.repo.run("--full-python", MOAI_PYTHON_SLOT_DIR="", MOAI_PYTHON_SLOT_MAX="")
        self.assertIn(f" --dir {self.default_dir} --max 2 ", self.python_lines()[0])

    def test_an_unset_or_empty_home_never_stops_the_script_and_the_item_runs_directly(self):
        # Audit B1: the folder default read $HOME at start-up and died with `unbound variable`.
        for label, extra in (("unset", {"HOME": None}), ("empty", {"HOME": ""})):
            with self.subTest(home=label):
                self.repo.calls.unlink(missing_ok=True)
                self.repo.slot_calls.unlink(missing_ok=True)
                proc = self.repo.run("--full-python", **extra)
                self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
                self.assertNotIn("unbound variable", proc.stderr)
                self.assertIn("no python slot folder", proc.stderr)
                self.assertEqual(self.python_lines(), [])
                # Never the cargo slot folder either (REQ-12): no slot line carries the python run.
                self.assertEqual([l for l in self.repo.slot_recorded() if "unittest" in l], [])
                self.assertIn("python3 -m unittest discover -s tests -p *_test.py",
                              self.repo.recorded())

    def test_an_empty_xdg_cache_home_falls_back_to_home(self):
        self.repo.run("--full-python", XDG_CACHE_HOME="")
        self.assertIn(f" --dir {self.default_dir} --max 2 ", self.python_lines()[0])

    def test_the_folder_setting_alone_works_without_home(self):
        self.repo.run("--full-python", HOME=None, MOAI_PYTHON_SLOT_DIR="/x/y")
        self.assertIn(" --dir /x/y --max 2 ", self.python_lines()[0])

    def test_the_python_slot_never_uses_the_cargo_slot_folder(self):
        self.repo.run("--full-python")
        for line in self.repo.slot_recorded():
            if " --dir " not in line:
                self.assertNotIn("python-slots", line)
        self.assertNotIn("heavy-slots", self.python_lines()[0])

    def test_a_refusal_of_the_python_slot_is_not_run_not_a_failure(self):
        proc = self.repo.run("--full-python", STUB_SLOT_REFUSE="1")
        self.assertEqual(proc.returncode, 1)
        found, last = items(proc.stdout)
        self.assertEqual(dict(found)["python"],
                         "exit=75 slot-refused (not run) log=target/self-check/python.log")
        m = LAST.match(last)
        self.assertEqual((m["result"], m["failed"], m["not_run"]),
                         ("not_run", "none", ",".join(HEAVY + ["python"])))

    def test_without_the_slot_command_full_python_runs_directly(self):
        repo = FakeRepo()
        self.addCleanup(repo.close)
        repo.run("--full-python")
        self.assertEqual(repo.slot_recorded(), [])
        self.assertIn("python3 -m unittest discover -s tests -p *_test.py", repo.recorded())


PRE_READY = [
    ("diff-check", None),
    ("fmt", "cargo fmt --check"),
    ("fmt-ops", "rustfmt --check --edition 2021 src/ops/a.rs src/ops/b.rs"),
    ("manifest", "python3 -I tools/source_manifest.py check"),
    ("spec-check", "moai-cli spec check SPEC-20261119-demo --closure --project ."),
    ("py-source-manifest", "python3 -m unittest tests.source_manifest_test"),
    ("py-repo-scans", "python3 -m unittest tests.repo_scans_test"),
    ("skill-package",
     "cargo test --offline --locked --test product_cli skill_package"),
]
PRE_READY_LABELS = [label for label, _ in PRE_READY]
PRE_READY_COMMANDS = [cmd for _, cmd in PRE_READY if cmd]
BIN_ITEMS = ["spec-check", "py-repo-scans"]
SPEC_PATH = ".moai/specs/SPEC-20261119-demo/spec.md"


def prepare_pre_ready(repo, spec=True, binary=True):
    """Adds what the --pre-ready items need to a FakeRepo: the two test files, product_cli, one
    changed SPEC, and a stub built `moai-cli` newer than every source file. The stub records
    its arguments and answers STUB_SPEC_OUT with exit STUB_SPEC_EXIT."""
    root = repo.repo
    (root / "tests" / "source_manifest_test.py").write_text("# stub\n")
    (root / "tests" / "repo_scans_test.py").write_text("# stub\n")
    (root / "tests" / "product_cli.rs").write_text("\n")
    if spec:
        (root / SPEC_PATH).parent.mkdir(parents=True)
        (root / SPEC_PATH).write_text("# demo\n")
    repo.commit("pre-ready files")
    past = time.time() - 100
    for path in (root / "Cargo.toml", root / "src" / "ops" / "a.rs", root / "src" / "ops" / "b.rs"):
        os.utime(path, (past, past))
    if binary:
        built = root / "target" / "debug" / "moai-cli"
        built.parent.mkdir(parents=True)
        built.write_text(
            "#!/bin/sh\n"
            f"echo \"moai-cli $*\" >> '{repo.calls}'\n"
            "if [ -n \"${STUB_SPEC_ONLY:-}\" ]; then\n"
            "  case \"$*\" in *\"$STUB_SPEC_ONLY\"*) ;; *) exit 0 ;; esac\n"
            "fi\n"
            "printf '%s\\n' \"${STUB_SPEC_OUT:-}\"\n"
            "exit \"${STUB_SPEC_EXIT:-0}\"\n")
        built.chmod(0o755)


class PreReadyTest(unittest.TestCase):
    """--pre-ready: the whole-tree checks that held pull requests back before READY, fast ones
    first; the two items that read the built moai-cli are not_run while it is missing or stale."""

    def setUp(self):
        self.repo = FakeRepo()
        self.addCleanup(self.repo.close)
        prepare_pre_ready(self.repo)

    def lines(self, proc):
        found, last = items(proc.stdout)
        return dict(found), LAST.match(last)

    def test_it_runs_only_its_items_in_order_and_passes(self):
        proc = self.repo.run("--pre-ready")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], PRE_READY_LABELS)
        for label, rest in found:
            self.assertTrue(rest.startswith("exit=0 "), f"{label}: {rest}")
        self.assertEqual(self.repo.recorded(), PRE_READY_COMMANDS)
        m = LAST.match(last)
        self.assertEqual((m["result"], m["failed"], m["not_run"]), ("pass", "none", "none"))
        # None of the ordinary items ran.
        self.assertFalse(any(c.startswith("cargo test --offline --locked --bin")
                             for c in self.repo.recorded()))

    def test_python_items_use_the_python_slot_and_the_cargo_item_the_heavy_slot(self):
        repo = FakeRepo(slot=True)
        self.addCleanup(repo.close)
        prepare_pre_ready(repo)
        proc = repo.run("--pre-ready")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        slot = repo.slot_recorded()
        python = [l for l in slot if " --dir " in l]
        cargo = [l for l in slot if " --dir " not in l]
        self.assertEqual(len(python), 2, slot)
        self.assertIn("-- python3 -m unittest tests.source_manifest_test", python[0])
        self.assertIn("-- python3 -m unittest tests.repo_scans_test", python[1])
        self.assertEqual(len(cargo), 1, slot)
        self.assertIn("-- cargo test --offline --locked --test product_cli skill_package",
                      cargo[0])
        # The light items take no slot.
        self.assertFalse(any("cargo fmt" in l or "source_manifest.py" in l for l in slot))

    def test_a_missing_binary_is_not_run_for_the_two_items_that_read_it(self):
        (self.repo.repo / "target" / "debug" / "moai-cli").unlink()
        proc = self.repo.run("--pre-ready")
        self.assertEqual(proc.returncode, 1)
        rest, m = self.lines(proc)
        for label in BIN_ITEMS:
            self.assertEqual(rest[label],
                             "not_run (no built moai-cli: target/debug/moai-cli)", label)
        self.assertEqual((m["result"], m["failed"], m["not_run"]),
                         ("not_run", "none", ",".join(BIN_ITEMS)))
        recorded = " ".join(self.repo.recorded())
        self.assertNotIn("tests.repo_scans_test", recorded)
        self.assertNotIn("spec check", recorded)
        # The items that need no binary still ran.
        for label in ("py-source-manifest", "skill-package"):
            self.assertTrue(rest[label].startswith("exit=0 "), label)

    def test_a_binary_older_than_a_source_file_is_stale_not_a_pass(self):
        future = time.time() + 100
        os.utime(self.repo.repo / "src" / "ops" / "a.rs", (future, future))
        proc = self.repo.run("--pre-ready")
        rest, m = self.lines(proc)
        for label in BIN_ITEMS:
            self.assertEqual(
                rest[label],
                "not_run (stale moai-cli: target/debug/moai-cli is older than src/ops/a.rs)",
                label)
        self.assertEqual(m["not_run"], ",".join(BIN_ITEMS))

    def test_the_binary_path_can_be_named(self):
        elsewhere = self.repo.tmp / "elsewhere" / "moai-cli"
        elsewhere.parent.mkdir()
        shutil.move(self.repo.repo / "target" / "debug" / "moai-cli", elsewhere)
        proc = self.repo.run("--pre-ready", MOAI_CLI_BIN=str(elsewhere))
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(self.repo.recorded(), PRE_READY_COMMANDS)

    def test_a_draft_specs_closure_status_alone_passes_spec_check(self):
        proc = self.repo.run("--pre-ready", STUB_SPEC_EXIT="1",
                             STUB_SPEC_OUT=f"spec check refused: 1 error(s): {SPEC_PATH} closure_status")
        rest, m = self.lines(proc)
        self.assertTrue(rest["spec-check"].startswith("exit=0 "), rest["spec-check"])
        self.assertEqual(m["failed"], "none")

    def test_any_other_code_fails_spec_check_even_beside_closure_status(self):
        for out in (f"{SPEC_PATH}:31 req_gears; {SPEC_PATH} closure_status",
                    f"{SPEC_PATH} closure_result",
                    f"{SPEC_PATH} closure_status; {SPEC_PATH} closure_result"):
            with self.subTest(out=out):
                proc = self.repo.run("--pre-ready", STUB_SPEC_EXIT="1", STUB_SPEC_OUT=out)
                rest, m = self.lines(proc)
                self.assertTrue(rest["spec-check"].startswith("exit=1 "), rest["spec-check"])
                self.assertEqual(m["failed"], "spec-check")
                self.assertEqual(proc.returncode, 1)

    def test_a_refusal_with_no_readable_code_fails_spec_check(self):
        proc = self.repo.run("--pre-ready", STUB_SPEC_EXIT="2", STUB_SPEC_OUT="crashed")
        self.assertEqual(self.lines(proc)[1]["failed"], "spec-check")

    def test_no_changed_spec_is_a_skip_that_needs_no_binary(self):
        repo = FakeRepo()
        self.addCleanup(repo.close)
        prepare_pre_ready(repo, spec=False, binary=False)
        proc = repo.run("--pre-ready")
        rest, m = self.lines(proc)
        self.assertEqual(rest["spec-check"], "skip (no changed SPEC)")
        self.assertEqual(m["not_run"], "py-repo-scans")

    def test_a_base_that_does_not_resolve_is_not_run_for_diff_and_spec_checks(self):
        proc = self.repo.run("--pre-ready", SELF_CHECK_BASE="no-such-branch")
        rest, m = self.lines(proc)
        for label in ("diff-check", "spec-check"):
            self.assertTrue(rest[label].startswith("not_run (no base: 'no-such-branch'"), label)
        self.assertEqual(m["not_run"], "diff-check,spec-check")

    def test_a_missing_test_file_is_a_skip_with_the_reason(self):
        (self.repo.repo / "tests" / "repo_scans_test.py").unlink()
        rest, m = self.lines(self.repo.run("--pre-ready"))
        self.assertEqual(rest["py-repo-scans"], "skip (no tests/repo_scans_test.py)")
        self.assertNotIn("tests.repo_scans_test", " ".join(self.repo.recorded()))

    def test_a_failing_item_is_named_and_the_later_items_still_run(self):
        proc = self.repo.run("--pre-ready", STUB_FAIL="fmt --check")
        rest, m = self.lines(proc)
        self.assertEqual(m["failed"], "fmt")
        self.assertTrue(rest["skill-package"].startswith("exit=0 "))

    def test_dry_run_lists_the_items_and_runs_nothing(self):
        proc = self.repo.run("--pre-ready", SELF_CHECK_DRY_RUN="1")
        self.assertEqual(proc.returncode, 1)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], PRE_READY_LABELS)
        self.assertEqual(self.repo.recorded(), [])
        self.assertEqual(LAST.match(last)["dry"], "dry-run=1 ")

    def test_every_source_file_named_in_the_stale_rule_makes_the_binary_stale(self):
        # N/N+1 of the rule: a file newer than the binary is stale, one older is not.
        for name in ("Cargo.toml", "Cargo.lock", "build.rs", "src/ops/b.rs"):
            with self.subTest(file=name):
                repo = FakeRepo()
                self.addCleanup(repo.close)
                prepare_pre_ready(repo)
                path = repo.repo / name
                path.write_text(path.read_text() if path.exists() else "x\n")
                future = time.time() + 100
                os.utime(path, (future, future))
                rest, m = self.lines(repo.run("--pre-ready"))
                self.assertEqual(
                    rest["py-repo-scans"],
                    f"not_run (stale moai-cli: target/debug/moai-cli is older than {name})")
        for name in ("Cargo.lock", "build.rs"):
            with self.subTest(older=name):
                repo = FakeRepo()
                self.addCleanup(repo.close)
                prepare_pre_ready(repo)
                path = repo.repo / name
                path.write_text("x\n")
                past = time.time() - 100
                os.utime(path, (past, past))
                rest, m = self.lines(repo.run("--pre-ready"))
                self.assertTrue(rest["py-repo-scans"].startswith("exit=0 "), rest)

    def test_a_new_uncommitted_spec_is_checked_too(self):
        spec = self.repo.repo / ".moai" / "specs" / "SPEC-20261119-fresh" / "spec.md"
        spec.parent.mkdir(parents=True)
        spec.write_text("# fresh\n")
        proc = self.repo.run("--pre-ready")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertIn("(SPEC-20261119-demo SPEC-20261119-fresh)", proc.stdout)
        self.assertIn("moai-cli spec check SPEC-20261119-fresh --closure --project .",
                      self.repo.recorded())

    def test_one_failing_spec_among_two_fails_the_item_and_both_are_checked(self):
        second = self.repo.repo / ".moai" / "specs" / "SPEC-20261119-second" / "spec.md"
        second.parent.mkdir(parents=True)
        second.write_text("# second\n")
        proc = self.repo.run("--pre-ready", STUB_SPEC_ONLY="SPEC-20261119-second",
                             STUB_SPEC_EXIT="1", STUB_SPEC_OUT=f"{SPEC_PATH} closure_result")
        rest, m = self.lines(proc)
        self.assertEqual(m["failed"], "spec-check")
        calls = [c for c in self.repo.recorded() if c.startswith("moai-cli spec check")]
        self.assertEqual(len(calls), 2, calls)

    def test_it_cannot_be_combined_with_full_python(self):
        for args in (["--pre-ready", "--full-python"], ["--full-python", "--pre-ready"]):
            proc = self.repo.run(*args)
            self.assertEqual(proc.returncode, 2, args)
            self.assertIn("usage", proc.stderr)
        self.assertEqual(self.repo.recorded(), [])


class SelfCheckSourcesTest(unittest.TestCase):
    def test_the_bound_is_the_profile_limit(self):
        self.assertIn("pub const LIMIT: u64 = 65536;", PROFILE.read_text())
        self.assertRegex(SCRIPT.read_text(), r"(?m)^limit=65536 ")

    def test_the_refusal_code_is_the_slot_busy_code(self):
        self.assertIn("pub const BUSY: u8 = 75;", HEAVY_SLOT.read_text())
        self.assertRegex(SCRIPT.read_text(), r"(?m)^slot_refused=75 ")

    def test_the_reference_names_the_runner_and_every_item(self):
        text = REFERENCE.read_text()
        self.assertIn("`tools/self_check.sh`", text)
        self.assertIn("`--full-python`", text)
        for label in LABELS:
            self.assertIn(f"`{label}`", text, label)


MODE_AGENT = ROOT / "skills" / "moai" / "workflows" / "mode-agent.md"
VERIFY = ROOT / "skills" / "moai" / "workflows" / "verify.md"


def paragraph(path, start):
    """The one line of `path` that starts with `start` (the skill files keep a rule on one line)."""
    lines = [l for l in path.read_text().splitlines() if l.startswith(start)]
    if len(lines) != 1:
        raise AssertionError(f"{path.name}: {len(lines)} lines start with {start!r}, want 1")
    return lines[0]


class SkillDocsNameThePythonSlotTest(unittest.TestCase):
    """SPEC-20261119-skills-python-slot-lines: the three documents lanes and audits read name
    what the scripts do (the python slot and `--pre-ready`), so a later edit that drops a name
    turns a test red. Only names are read, never a whole sentence (SPEC decision D-3)."""

    def test_self_check_lists_the_pre_ready_items_in_the_scripts_order(self):
        text = REFERENCE.read_text()
        self.assertIn("`--pre-ready`", text)
        at = -1
        for label in PRE_READY_LABELS:
            found = text.find(f"`{label}`", at + 1)
            self.assertGreater(found, at, f"{label} is not named after the item before it")
            at = found

    def test_self_check_says_a_head_passes_the_list_before_ready(self):
        text = REFERENCE.read_text()
        self.assertRegex(text, r"(?i)before the lane reports (it )?ready|before (a )?READY")

    def test_self_check_names_the_python_slot_and_its_two_variables(self):
        text = REFERENCE.read_text()
        for name in ("python slot", "MOAI_PYTHON_SLOT_DIR", "MOAI_PYTHON_SLOT_MAX"):
            self.assertIn(name, text, name)
        self.assertIn("not run", text)

    def test_mode_agent_heavy_checks_names_pre_ready_and_the_python_slot(self):
        line = paragraph(MODE_AGENT, "7. **heavy-checks**")
        self.assertIn("`tools/self_check.sh --pre-ready`", line)
        self.assertIn("python slot", line)

    def test_verify_serialised_paragraph_names_the_python_slot(self):
        line = paragraph(VERIFY, "Heavy checks are serialised")
        self.assertIn("python slot", line)

    def test_the_names_the_documents_use_are_the_scripts_own(self):
        script = SCRIPT.read_text()
        for name in ("--pre-ready", "MOAI_PYTHON_SLOT_DIR", "MOAI_PYTHON_SLOT_MAX"):
            self.assertIn(name, script, name)


if __name__ == "__main__":
    unittest.main()
