"""Tests for tools/f2p_evidence.py (SPEC-20261121-fail-to-pass-evidence).

The tool lists the Rust integration tests (`tests/*.rs`) that a head adds or changes against a
base, runs them on the base source with the head's test files, and prints one JSON line whose
`summary` reads `f2p=<n> flagged=<m>`. Each test builds a temporary git repository (a base
commit and a head commit) and a fake `cargo`, a script that prints `test <name> ... ok|FAILED`
as the environment tells it, so no real build runs. The tool is started as a maintainer would
start it, through `python3 -I tools/f2p_evidence.py`. Written for Python 3.9.
"""

import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TOOL = REPO_ROOT / "tools" / "f2p_evidence.py"
PREFIX = "f2p_evidence: "

BASE_A = """#[test]
fn old_one() {
    assert_eq!(1 + 1, 2);
}

#[test]
fn unchanged() {
    assert!(true);
}
"""

HEAD_A = """#[test]
fn old_one() {
    assert_eq!(1 + 2, 3);
}

#[test]
fn unchanged() {
    assert!(true);
}

#[test]
fn added() {
    assert_eq!(2 + 2, 4);
}
"""

FAKE_CARGO = """#!%s
import json, os, sys
args = sys.argv[1:]
log = os.environ.get("F2P_FAKE_LOG")
if log:
    def read(rel):
        try:
            return open(rel, encoding="utf-8").read().strip()
        except OSError:
            return "absent"
    with open(log, "a", encoding="utf-8") as f:
        f.write(json.dumps({"args": args, "cwd": os.getcwd(),
                            "marker": read("src/marker.txt"),
                            "has_added": "fn added" in read("tests/a.rs")}) + "\\n")
if os.environ.get("F2P_FAKE_BUILD_ERROR"):
    sys.stderr.write("error[E0432]: unresolved import `x`\\n")
    sys.exit(101)
if os.environ.get("F2P_FAKE_SILENT"):
    sys.exit(0)
results = json.loads(os.environ.get("F2P_FAKE", "{}"))
names = args[args.index("--exact") + 1:] if "--exact" in args else []
failed = False
for n in names:
    r = results.get(n, "ok")
    print("test %%s ... %%s" %% (n, r))
    failed = failed or r != "ok"
sys.exit(101 if failed else 0)
""" % sys.executable


def run_git(cwd, *args):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env.update({"GIT_CONFIG_NOSYSTEM": "1", "HOME": str(cwd)})
    return subprocess.run(
        ["git", "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false"]
        + list(args),
        cwd=str(cwd), capture_output=True, text=True, env=env, check=True,
    ).stdout.strip()


def write(root, rel, text):
    path = Path(root) / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


class Fixture(unittest.TestCase):
    """A temporary repository with a base commit and a head commit, and a fake cargo."""

    def setUp(self):
        self.tmp = tempfile.mkdtemp(prefix="f2p-test-")
        self.addCleanup(shutil.rmtree, self.tmp, True)
        self.repo = Path(self.tmp) / "repo"
        self.repo.mkdir()
        run_git(self.repo, "init", "-q", "-b", "main")
        write(self.repo, "tests/a.rs", BASE_A)
        write(self.repo, "tests/b.rs", "#[test]\nfn only_in_b() {\n    assert!(true);\n}\n")
        write(self.repo, "src/marker.txt", "base-source\n")
        run_git(self.repo, "add", ".")
        run_git(self.repo, "commit", "-q", "-m", "base")
        self.base = run_git(self.repo, "rev-parse", "HEAD")
        write(self.repo, "tests/a.rs", HEAD_A)
        write(self.repo, "src/marker.txt", "head-source\n")
        run_git(self.repo, "add", ".")
        run_git(self.repo, "commit", "-q", "-m", "head")
        self.head = run_git(self.repo, "rev-parse", "HEAD")
        self.cargo = Path(self.tmp) / "fakecargo"
        self.cargo.write_text(FAKE_CARGO, encoding="utf-8")
        self.cargo.chmod(0o755)
        self.log = Path(self.tmp) / "calls.log"

    def calls(self):
        if not self.log.exists():
            return []
        return [json.loads(line) for line in self.log.read_text(encoding="utf-8").splitlines()]

    def tool(self, *extra, fake=None, build_error=False, silent=False, base=None, head=None, directory=None, cargo=True):
        env = {k: v for k, v in os.environ.items() if not k.startswith("F2P_")}
        env["F2P_FAKE_LOG"] = str(self.log)
        if fake is not None:
            env["F2P_FAKE"] = json.dumps(fake)
        if build_error:
            env["F2P_FAKE_BUILD_ERROR"] = "1"
        if silent:
            env["F2P_FAKE_SILENT"] = "1"
        argv = [sys.executable, "-I", str(TOOL), "--dir", str(directory or self.repo),
                "--base", base or self.base, "--head", head or self.head]
        if cargo:
            argv += ["--cargo", str(self.cargo)]
        argv += list(extra)
        return subprocess.run(argv, capture_output=True, text=True, env=env)

    def report(self, proc):
        self.assertEqual(proc.stdout.count("\n"), 1, proc.stdout + proc.stderr)
        return json.loads(proc.stdout)

    def snapshot(self):
        out = {}
        for p in sorted(self.repo.rglob("*")):
            if p.is_file() and ".git" not in p.relative_to(self.repo).parts:
                out[str(p.relative_to(self.repo))] = p.read_bytes()
        out["HEAD"] = run_git(self.repo, "rev-parse", "HEAD")
        out["status"] = run_git(self.repo, "status", "--porcelain")
        return out


class ListTest(Fixture):
    def test_the_list_holds_the_added_and_the_changed_test_and_not_the_unchanged_one(self):
        proc = self.tool("--list-only")
        self.assertEqual(proc.returncode, 0, proc.stderr)
        tests = self.report(proc)["tests"]
        got = sorted((t["file"], t["name"], t["kind"]) for t in tests)
        self.assertEqual(got, [("tests/a.rs", "added", "new"), ("tests/a.rs", "old_one", "changed")])

    def test_list_only_runs_no_program(self):
        proc = self.tool("--list-only")
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(self.calls(), [])

    def test_a_file_the_head_leaves_alone_is_not_listed(self):
        tests = self.report(self.tool("--list-only"))["tests"]
        self.assertNotIn("tests/b.rs", [t["file"] for t in tests])

    def test_a_deleted_test_is_not_listed(self):
        write(self.repo, "tests/a.rs", HEAD_A.replace("#[test]\nfn unchanged() {\n    assert!(true);\n}\n\n", ""))
        run_git(self.repo, "add", ".")
        run_git(self.repo, "commit", "-q", "-m", "drop one")
        head = run_git(self.repo, "rev-parse", "HEAD")
        tests = self.report(self.tool("--list-only", head=head))["tests"]
        self.assertNotIn("unchanged", [t["name"] for t in tests])

    def test_a_new_test_file_lists_all_its_tests_as_new(self):
        write(self.repo, "tests/c.rs", "#[test]\nfn c_one() {}\n\n#[test]\nfn c_two() {}\n")
        run_git(self.repo, "add", ".")
        run_git(self.repo, "commit", "-q", "-m", "add c")
        head = run_git(self.repo, "rev-parse", "HEAD")
        tests = self.report(self.tool("--list-only", head=head))["tests"]
        got = sorted((t["name"], t["kind"]) for t in tests if t["file"] == "tests/c.rs")
        self.assertEqual(got, [("c_one", "new"), ("c_two", "new")])


class RunTest(Fixture):
    def test_the_summary_counts_the_failing_and_the_flagged_tests(self):
        proc = self.tool(fake={"added": "FAILED", "old_one": "ok"})
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        report = self.report(proc)
        self.assertEqual(report["summary"], "f2p=1 flagged=1")
        results = {t["name"]: t["base"] for t in report["tests"]}
        self.assertEqual(results, {"added": "fail", "old_one": "pass"})

    def test_all_failing_on_the_base_exits_zero(self):
        proc = self.tool(fake={"added": "FAILED", "old_one": "FAILED"})
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(self.report(proc)["summary"], "f2p=2 flagged=0")

    def test_the_head_test_file_runs_on_the_base_source(self):
        proc = self.tool(fake={"added": "FAILED", "old_one": "FAILED"})
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        calls = self.calls()
        self.assertEqual(len(calls), 1, calls)
        self.assertEqual(calls[0]["marker"], "base-source")
        self.assertTrue(calls[0]["has_added"], "the head's tests/a.rs must be the file that runs")

    def test_one_cargo_call_per_test_file_with_exact_names(self):
        write(self.repo, "tests/c.rs", "#[test]\nfn c_one() {}\n")
        run_git(self.repo, "add", ".")
        run_git(self.repo, "commit", "-q", "-m", "add c")
        head = run_git(self.repo, "rev-parse", "HEAD")
        self.tool(fake={"added": "FAILED", "old_one": "FAILED", "c_one": "FAILED"}, head=head)
        argvs = sorted(" ".join(c["args"]) for c in self.calls())
        self.assertEqual(len(argvs), 2, argvs)
        self.assertTrue(any(a.startswith("test --test a") and "--exact" in a and "added" in a and "old_one" in a for a in argvs), argvs)
        self.assertTrue(any(a.startswith("test --test c") and a.endswith("c_one") for a in argvs), argvs)

    def test_a_base_that_does_not_build_fails_every_listed_test_with_build_error(self):
        proc = self.tool(build_error=True)
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        report = self.report(proc)
        self.assertEqual(report["summary"], "f2p=2 flagged=0")
        self.assertEqual(sorted(t["reason"] for t in report["tests"]), ["build_error", "build_error"])

    def test_every_test_passing_on_the_base_flags_them_all(self):
        proc = self.tool(fake={})
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        self.assertEqual(self.report(proc)["summary"], "f2p=0 flagged=2")


class ReasonTest(Fixture):
    def test_a_reason_turns_the_flag_into_a_recorded_note_and_exit_zero(self):
        proc = self.tool("--reason", "pure refactor, behaviour unchanged", fake={"added": "FAILED", "old_one": "ok"})
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        report = self.report(proc)
        self.assertEqual(report["summary"], "f2p=1 flagged=0")
        self.assertEqual(report["reason"], "pure refactor, behaviour unchanged")

    def test_without_a_reason_the_same_run_exits_one(self):
        proc = self.tool(fake={"added": "FAILED", "old_one": "ok"})
        self.assertEqual(proc.returncode, 1)
        self.assertNotIn("reason", self.report(proc))

    def test_an_empty_reason_is_refused(self):
        proc = self.tool("--reason", "  ", fake={"added": "FAILED"})
        self.assertEqual(proc.returncode, 2, proc.stdout)
        self.assertTrue(proc.stderr.startswith(PREFIX), proc.stderr)
        self.assertEqual(proc.stdout, "")


class RefusalTest(Fixture):
    def refused(self, proc):
        self.assertEqual(proc.returncode, 2, proc.stdout + proc.stderr)
        self.assertEqual(proc.stdout, "")
        self.assertTrue(proc.stderr.startswith(PREFIX), proc.stderr)

    def test_an_unknown_base_is_refused(self):
        self.refused(self.tool(base="0" * 40))

    def test_an_unknown_head_is_refused(self):
        self.refused(self.tool(head="no-such-branch"))

    def test_a_missing_directory_is_refused(self):
        self.refused(self.tool(directory=Path(self.tmp) / "nowhere"))

    def test_a_base_that_is_not_an_ancestor_of_the_head_is_refused(self):
        self.refused(self.tool(base=self.head, head=self.base))

    def test_a_run_that_prints_no_verdict_for_a_listed_test_is_refused(self):
        self.refused(self.tool(silent=True))

    def test_a_missing_cargo_program_is_refused(self):
        proc = subprocess.run(
            [sys.executable, "-I", str(TOOL), "--dir", str(self.repo), "--base", self.base,
             "--head", self.head, "--cargo", str(Path(self.tmp) / "no-cargo-here")],
            capture_output=True, text=True,
        )
        self.refused(proc)


class CleanTest(Fixture):
    def test_the_repository_is_untouched_after_a_run(self):
        before = self.snapshot()
        proc = self.tool(fake={"added": "FAILED", "old_one": "FAILED"})
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(self.snapshot(), before)

    def test_the_repository_is_untouched_after_a_list(self):
        before = self.snapshot()
        proc = self.tool("--list-only")
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertEqual(self.snapshot(), before)

    def test_the_temporary_build_folder_is_gone_afterwards(self):
        proc = self.tool(fake={"added": "FAILED", "old_one": "FAILED"})
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertTrue(self.calls(), "the fake cargo must have been called")
        cwd = Path(self.calls()[0]["cwd"])
        self.assertFalse(cwd.exists(), "the folder the base was built in must be removed: %s" % cwd)

    def test_the_temporary_folder_is_gone_after_a_build_error_too(self):
        proc = self.tool(build_error=True)
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        self.assertTrue(self.calls(), "the fake cargo must have been called")
        cwd = Path(self.calls()[0]["cwd"])
        self.assertFalse(cwd.exists(), "the folder the base was built in must be removed: %s" % cwd)


if __name__ == "__main__":
    unittest.main()
