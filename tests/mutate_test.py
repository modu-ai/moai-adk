"""Behavioral CLI contracts for the isolated mutation runner (stdlib only)."""

import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import time
import unittest

ROOT = Path(__file__).resolve().parent.parent
RUNNER = ROOT / "tools" / "mutate.py"
OLD = "pub fn answer() -> u32 { 41 }"
NEW = "pub fn answer() -> u32 { 42 }"

# This collaborator reads the actual checkout and Git state; it does not echo
# the spec or manufacture runner output. All its side effects stay in tempdirs.
FAKE_CARGO = r'''import json, os, pathlib, subprocess, sys
p = pathlib.Path.cwd()
def git(*args):
    return subprocess.check_output(["git", *args], cwd=p, text=True).strip()
record = {
    "cwd": str(p.resolve()), "argv": sys.argv[1:],
    "code": (p / "src/lib.rs").read_text(),
    "diff": git("diff", "--no-ext-diff", "-U0", "HEAD"),
    "head": git("rev-parse", "HEAD"),
    "git_dir": git("rev-parse", "--absolute-git-dir"),
    "common_dir": git("rev-parse", "--git-common-dir"),
    "branch": subprocess.run(["git", "symbolic-ref", "-q", "HEAD"], cwd=p,
                             capture_output=True).returncode,
    "target": os.environ.get("CARGO_TARGET_DIR"),
    "home": os.environ.get("HOME"), "tmp": os.environ.get("TMPDIR"),
    "xdg": os.environ.get("XDG_CONFIG_HOME"),
    "git_env": {k: v for k, v in os.environ.items() if k.startswith("GIT_")},
    "export_ignored": (p / "kept.txt").read_text(),
}
with open(os.environ["MUTATION_LOG"], "a") as f:
    f.write(json.dumps(record) + "\n")
mode = os.environ.get("FAKE_MODE", "killed")
if mode == "directory":
    (p / "src/lib.rs").unlink()
    (p / "src/lib.rs").mkdir()
elif mode == "symlink":
    (p / "src/lib.rs").unlink()
    (p / "src/lib.rs").symlink_to(os.environ["SENTINEL"])
elif mode == "index":
    subprocess.check_call(["git", "add", "src/lib.rs"], cwd=p)
elif mode in ("assume_unchanged", "skip_worktree"):
    (p / "kept.txt").write_text("hidden clone edit\n")
    flag = "--assume-unchanged" if mode == "assume_unchanged" else "--skip-worktree"
    subprocess.check_call(["git", "update-index", flag, "kept.txt"], cwd=p)
summary = "test result: FAILED. 1 passed; 1 failed; 0 ignored; 0 filtered out;"
def emit(value):
    print(json.dumps(value), flush=True)
kind = next(k for k in ("lib", "bin", "test", "example", "bench")
            if "--" + k in sys.argv)
name = "fixture" if kind == "lib" else sys.argv[sys.argv.index("--" + kind) + 1]
artifact = {"reason": "compiler-artifact", "package_id": "fixture 0.1.0",
            "manifest_path": str(p / "Cargo.toml"),
            "target": {"name": name, "kind": [kind], "crate_types": ["lib" if kind == "lib" else "bin"]},
            "profile": {"test": True}, "executable": str(p / "target/fixture"),
            "filenames": [], "fresh": False}
if mode.startswith("compiler_") or mode == "compile":
    if mode == "compiler_json":
        emit({"reason": "compiler-message", "message": {"rendered": summary + "\n",
              "message": summary, "level": "error"}, "target": artifact["target"]})
    elif mode == "compiler_stdout":
        print(summary)
    elif mode == "compiler_forged_build_json":
        emit(artifact)
        emit({"reason": "build-finished", "success": True})
        print(summary)
    elif mode == "compiler_multiline":
        print("error: diagnostic\n" + summary, file=sys.stderr)
    else:
        print(summary if mode == "compiler_stderr" else "error[E0308]: mismatched types",
              file=sys.stderr)
    emit({"reason": "build-finished", "success": False})
    sys.exit(101)
if mode == "unknown":
    print("toolchain unavailable", file=sys.stderr)
    sys.exit(7)
if mode != "no_artifact":
    if mode == "wrong_artifact":
        artifact["target"] = {"name": "unselected", "kind": ["bin"]}
    if mode == "non_test_artifact":
        artifact["profile"]["test"] = False
    if mode == "no_executable":
        artifact["executable"] = None
    emit(artifact)
    if mode == "ambiguous_artifact":
        emit(artifact)
if mode == "false_then_true":
    emit({"reason": "build-finished", "success": False})
if mode == "build_script":
    emit({"reason": "build-script-executed", "package_id": "fixture 0.1.0",
          "linked_libs": [], "linked_paths": [], "cfgs": [], "env": [],
          "out_dir": str(p / "target/build"), "rendered": summary})
    print(summary, file=sys.stderr)
if mode != "no_build_finished":
    emit({"reason": "build-finished", "success": True})
if mode in ("missing", "build_script"):
    sys.exit(0)
if mode == "runtime_stderr":
    print(summary, file=sys.stderr)
    sys.exit(101)
if mode == "zero":
    print("test result: ok. 0 passed; 0 failed; 0 ignored; 8 filtered out;")
    sys.exit(0)
if mode == "zero_failed":
    print("test result: FAILED. 0 passed; 1 failed; 0 ignored; 0 filtered out;")
    sys.exit(101)
if mode == "survived":
    print("test result: ok. 2 passed; 0 failed; 0 ignored; 0 filtered out;")
    sys.exit(0)
print(summary)
sys.exit(101)
'''


# Load-independent child budget (#2502): the runner's real work is observable as
# appends to the mutation log (one JSON line per completed fake-cargo
# invocation), so the deadline resets on progress instead of expiring on
# wall-clock alone. A silent child fails the step deadline, a looping child
# fails the invocation budget, and slow-but-progressing work still meets the
# hard ceiling. Supersedes the bare 30-second timeout that raised five false
# subprocess.TimeoutExpired under load 167-218 with zero assertion failures
# (#2466 audit supplement 6811398145); rc-ci runs the same cases in 62 seconds.
PROGRESS_LOG_LINES_BUDGET = 64
STEP_DEADLINE_SECONDS = 120
HARD_CEILING_SECONDS = 600
_POLL_SECONDS = 0.2


class BudgetExceeded(RuntimeError):
    def __init__(self, reason, stdout="", stderr=""):
        super().__init__("mutation runner budget exceeded: " + reason)
        self.reason = reason
        self.stdout = stdout
        self.stderr = stderr


def run_with_progress_budget(argv, cwd, env, log,
                             step_seconds=STEP_DEADLINE_SECONDS,
                             ceiling_seconds=HARD_CEILING_SECONDS,
                             budget=PROGRESS_LOG_LINES_BUDGET,
                             clock=time.monotonic, sleep=time.sleep,
                             spawn=subprocess.Popen):
    """Run a child bounded by progress, not by a bare wall-clock timeout.

    The deadline resets whenever ``log`` gains a line, so a loaded machine only
    has to keep making progress. Child output is captured into temporary files
    (never undrained pipes, which would deadlock a chatty child and look like
    no progress). Both deadline comparisons allow one poll interval of grace so
    punctual progress is not killed by scheduling jitter. ``clock``, ``sleep``
    and ``spawn`` are seams: production uses the real ones and tests simulate
    time and children deterministically.
    """

    # The log is shared across invocations (MutationCLI keeps one file), so the
    # budget bounds this child's own entries: count from a start baseline and
    # check the final delta too, because a child can overspend and exit between
    # two polls — the exit branch must not accept it.
    def observed_lines():
        try:
            return len(Path(log).read_text().splitlines())
        except (FileNotFoundError, NotADirectoryError):
            return 0

    baseline = observed_lines()
    start = clock()
    deadline = start + step_seconds
    ceiling = start + ceiling_seconds
    seen = baseline
    with tempfile.TemporaryFile("w+", encoding="utf-8", errors="replace") as out_file, \
            tempfile.TemporaryFile("w+", encoding="utf-8", errors="replace") as err_file:
        child = spawn(argv, cwd=cwd, env=env, stdout=out_file, stderr=err_file,
                      text=True)
        try:
            while True:
                if child.poll() is not None:
                    child.wait()
                    if observed_lines() - baseline > budget:
                        raise BudgetExceeded(
                            f"invocation budget of {budget} progress lines exceeded")
                    break
                lines = observed_lines()
                if lines - baseline > budget:
                    raise BudgetExceeded(
                        f"invocation budget of {budget} progress lines exceeded")
                if lines > seen:
                    seen = lines
                    deadline = clock() + step_seconds
                now = clock()
                if now > deadline + _POLL_SECONDS:
                    raise BudgetExceeded(
                        f"no progress for {step_seconds} seconds")
                if now > ceiling + _POLL_SECONDS:
                    raise BudgetExceeded(
                        f"hang ceiling of {ceiling_seconds} seconds exceeded")
                sleep(_POLL_SECONDS)
        except BudgetExceeded as failure:
            child.kill()
            child.wait()
            out_file.seek(0)
            err_file.seek(0)
            failure.stdout = out_file.read()
            failure.stderr = err_file.read()
            raise
        out_file.seek(0)
        err_file.seek(0)
        return subprocess.CompletedProcess(argv, child.returncode,
                                           out_file.read(), err_file.read())


class MutationCLI(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="mutate-contract-")
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name).resolve()
        self.source = self.base / "source"
        self.source.mkdir()
        self.home = self.base / "caller-home"
        self.home.mkdir()
        self.bin = self.base / "bin"
        self.bin.mkdir()
        self.log = self.base / "cargo.jsonl"
        self.sentinel = self.base / "sentinel"
        self.sentinel.write_bytes(b"external bytes must survive\n")
        self.env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
        self.env.update(HOME=str(self.home), XDG_CONFIG_HOME=str(self.home / "xdg"),
                        TMPDIR=str(self.base), GIT_CONFIG_NOSYSTEM="1",
                        PATH=str(self.bin) + os.pathsep + os.environ.get("PATH", ""),
                        MUTATION_LOG=str(self.log), SENTINEL=str(self.sentinel),
                        CARGO_TARGET_DIR=str(self.source / "target/mutation-cargo"))
        cargo = self.bin / "cargo"
        cargo.write_text("#!" + sys.executable + "\n" + FAKE_CARGO)
        cargo.chmod(0o755)
        (self.source / "src").mkdir()
        (self.source / "src/lib.rs").write_text(OLD + "\npub fn other() -> u32 { 9 }\n")
        (self.source / "Cargo.toml").write_text('[package]\nname="fixture"\nversion="0.1.0"\nedition="2021"\n')
        (self.source / "Cargo.lock").write_text("version = 3\n")
        (self.source / ".gitignore").write_text("/target/\n")
        (self.source / ".gitattributes").write_text("kept.txt export-ignore\n")
        (self.source / "kept.txt").write_text("HEAD bytes, not archive output\n")
        self.git("init", "-q")
        self.git("config", "user.name", "Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        self.git("add", ".")
        self.git("commit", "-qm", "fixture")
        self.spec = {"schema_version": 1, "mutations": [self.mutation()]}

    def git(self, *args, cwd=None):
        return subprocess.check_output(["git", *args], cwd=cwd or self.source,
                                       env=self.env, stderr=subprocess.STDOUT)

    def mutation(self, **changes):
        value = {"id": "answer", "file": "src/lib.rs", "old": OLD,
                 "new": NEW, "targets": [{"kind": "lib"}]}
        value.update(changes)
        return value

    def snapshot(self):
        import stat

        index = Path(os.fsdecode(self.git("rev-parse", "--git-path", "index").strip()))
        if not index.is_absolute():
            index = self.source / index
        tracked = self.git("ls-files", "-z").split(b"\0")
        contents = {}
        for name in tracked:
            if not name:
                continue
            path = self.source / os.fsdecode(name)
            mode = path.lstat().st_mode
            contents[os.fsdecode(name)] = (
                mode, os.readlink(path) if stat.S_ISLNK(mode) else path.read_bytes())
        return (self.git("rev-parse", "HEAD"),
                hashlib.sha256(index.read_bytes()).hexdigest(),
                contents,
                self.git("--no-optional-locks", "status", "--porcelain=v1", "--untracked-files=all"),
                self.git("show-ref"))

    def calls(self):
        return [json.loads(line) for line in self.log.read_text().splitlines()] if self.log.exists() else []

    def invoke(self, mode="killed", environment=None, spec=None):
        path = self.base / "spec.json"
        path.write_text(json.dumps(self.spec if spec is None else spec))
        before = self.snapshot()
        env = dict(self.env, FAKE_MODE=mode)
        if environment:
            env.update(environment)
        out = run_with_progress_budget(
            [sys.executable, "-I", str(RUNNER), "--spec", str(path)],
            cwd=self.source, env=env, log=self.log)
        self.assertEqual(self.snapshot(), before, "runner changed source HEAD/index/bytes/status/refs")
        self.assertEqual(self.sentinel.read_bytes(), b"external bytes must survive\n")
        try:
            report = json.loads(out.stdout)
        except ValueError:
            self.fail("CLI must emit exactly one JSON object: " + repr((out.returncode, out.stdout, out.stderr)))
        self.assertIsInstance(report, dict)
        self.assertEqual(report["schema_version"], 1)
        return out, report

    def assert_rejected(self, **kwargs):
        out, report = self.invoke(**kwargs)
        self.assertNotEqual(out.returncode, 0, report)
        self.assertIn(report["status"], ("blocked", "error"))
        self.assertTrue(report["errors"], report)
        self.assertEqual(self.calls(), [], "invalid input reached cargo")

    def test_code_line_can_be_replaced_by_a_comment(self):
        replacement = "// guard deliberately removed"
        out, report = self.invoke(spec={
            "schema_version": 1, "mutations": [self.mutation(new=replacement)]})
        self.assertEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "completed")
        self.assertTrue(report["source_unchanged"])
        calls = self.calls()
        self.assertEqual(len(calls), 1)
        self.assertEqual(calls[0]["code"], replacement + "\npub fn other() -> u32 { 9 }\n")
        mutation = report["mutations"][0]
        for diff in (calls[0]["diff"], mutation["diff"]):
            changes = [line for line in diff.splitlines()
                       if line.startswith(("+", "-")) and not line.startswith(("+++", "---"))]
            self.assertEqual(changes, ["-" + OLD, "+" + replacement])
        self.assertTrue(mutation["restored"])
        self.assertEqual(mutation["restore_diff_exit"], 0)

    def test_clone_restoration_uses_its_own_checkout_mode(self):
        source_file = self.source / "src/lib.rs"
        source_file.chmod(0o600)
        source_mode = source_file.stat().st_mode
        previous_umask = os.umask(0o022)
        try:
            out, report = self.invoke()
        finally:
            os.umask(previous_umask)
        self.assertEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "completed")
        self.assertTrue(report["source_unchanged"])
        self.assertEqual(source_file.stat().st_mode, source_mode)
        row = report["mutations"][0]
        self.assertTrue(row["restored"])
        self.assertEqual(row["restore_diff_exit"], 0)
        workspace = Path(report["workspace"])
        self.assertTrue(workspace.is_dir())
        self.assertEqual(self.git("diff", "HEAD", cwd=workspace), b"")

    def test_mutant_is_observed_once_per_explicit_target_in_independent_clone(self):
        targets = [{"kind": "lib", "filter": "answer"}] + [
            {"kind": kind, "name": "fixture_" + kind, "filter": "answer"}
            for kind in ("bin", "test", "example", "bench")]
        self.spec["mutations"][0]["targets"] = targets
        out, report = self.invoke()
        self.assertEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "completed")
        self.assertTrue(report["source_unchanged"])
        head = self.git("rev-parse", "HEAD").decode().strip()
        self.assertEqual(report["head"], head)
        calls = self.calls()
        self.assertEqual(len(calls), len(targets))
        workspace = Path(report["workspace"]).resolve()
        owned = (self.source / "target/mutation-cargo").resolve()
        actual_target = Path(report["cargo_target_dir"]).resolve()
        self.assertIn(self.source / "target", workspace.parents)
        self.assertIn(owned, actual_target.parents)
        self.assertNotEqual(actual_target, owned)
        self.assertEqual(len({call["target"] for call in calls}), 1)
        for call, target in zip(calls, targets):
            argv = call["argv"]
            self.assertEqual(argv[0], "test")
            for flag in ("--offline", "--locked", "--message-format=json"):
                self.assertIn(flag, argv)
            selectors = [flag for flag in argv
                         if flag in ("--lib", "--bin", "--test", "--example", "--bench")]
            self.assertEqual(selectors, ["--" + target["kind"]])
            if "name" in target:
                self.assertEqual(argv[argv.index(selectors[0]) + 1], target["name"])
            self.assertIn("answer", argv)
            self.assertNotIn("--no-run", argv, "one Cargo invocation must execute the selected tests")
            self.assertEqual(Path(call["cwd"]), workspace)
            self.assertEqual(call["code"], NEW + "\npub fn other() -> u32 { 9 }\n")
            self.assertIn("-" + OLD, call["diff"])
            self.assertIn("+" + NEW, call["diff"])
            self.assertEqual(call["head"], head)
            self.assertNotEqual(call["branch"], 0, "clone must be detached")
            self.assertEqual(Path(call["git_dir"]), workspace / ".git")
            common = Path(call["common_dir"])
            self.assertEqual((workspace / common).resolve() if not common.is_absolute() else common.resolve(), workspace / ".git")
            self.assertEqual(Path(call["target"]), actual_target)
            for key in ("home", "tmp", "xdg"):
                self.assertTrue(call[key], key)
                self.assertIn(self.source / "target", Path(call[key]).resolve().parents)
            self.assertEqual(call["export_ignored"], "HEAD bytes, not archive output\n")
        self.assertFalse((workspace / ".git/objects/info/alternates").exists())
        for original in (self.source / ".git/objects").glob("*/*"):
            clone_object = workspace / ".git/objects" / original.relative_to(self.source / ".git/objects")
            if original.is_file() and clone_object.exists():
                self.assertFalse(os.path.samefile(original, clone_object), "object hardlink aliases source")
        mutation = report["mutations"][0]
        self.assertEqual((mutation["id"], mutation["file"]), ("answer", "src/lib.rs"))
        self.assertIn("-" + OLD, mutation["diff"])
        self.assertIn("+" + NEW, mutation["diff"])
        self.assertTrue(mutation["restored"])
        self.assertEqual(mutation["restore_diff_exit"], 0)
        self.assertEqual((workspace / "src/lib.rs").read_bytes(), (self.source / "src/lib.rs").read_bytes())
        self.assertEqual(self.git("diff", "HEAD", cwd=workspace), b"")
        self.assertEqual(len(mutation["targets"]), len(targets))
        for result, call in zip(mutation["targets"], calls):
            self.assertEqual(result["argv"], ["cargo"] + call["argv"])
            self.assertEqual(result["status"], "FAILED")
            self.assertEqual((result["passed"], result["failed"], result["exit"]), (1, 1, 101))
            self.assertIn("test result: FAILED", result["stdout"])
            self.assertIsInstance(result["stderr"], str)

    def test_status_is_based_on_real_libtest_counts_and_exit_not_any_failure(self):
        cases = [("survived", "survived", 0, 2, 0),
                 ("compile", "error", 1, None, None),
                 ("unknown", "error", 1, None, None),
                 ("missing", "not_run", 1, None, None),
                 ("zero", "not_run", 1, 0, 0),
                 ("zero_failed", "not_run", 1, 0, 1)]
        for mode, status, exit_code, passed, failed in cases:
            with self.subTest(mode=mode):
                if self.log.exists():
                    self.log.unlink()
                out, report = self.invoke(mode)
                self.assertEqual(out.returncode, exit_code, report)
                result = report["mutations"][0]["targets"][0]
                self.assertEqual(result["status"], status)
                self.assertEqual((result["passed"], result["failed"]), (passed, failed))
                self.assertTrue(report["mutations"][0]["restored"])
                self.assertEqual(len(self.calls()), 1)
                if mode == "zero_failed":
                    self.assertIn("0 passed; 1 failed", result["stdout"])
                if mode == "compile":
                    self.assertIn("E0308", result["stderr"])

    def test_restoration_failure_stops_before_next_mutation_and_never_follows_symlink(self):
        for mode in ("directory", "symlink", "index"):
            with self.subTest(mode=mode):
                if self.log.exists():
                    self.log.unlink()
                self.spec["mutations"] = [self.mutation(), self.mutation(
                    id="other", old="pub fn other() -> u32 { 9 }", new="pub fn other() -> u32 { 8 }")]
                out, report = self.invoke(mode)
                self.assertEqual(out.returncode, 1)
                self.assertIn(report["status"], ("blocked", "error"))
                self.assertFalse(report["mutations"][0]["restored"])
                self.assertNotEqual(report["mutations"][0]["restore_diff_exit"], 0)
                self.assertEqual(len(self.calls()), 1)
                self.assertTrue(report["errors"])

    def test_hidden_index_flags_block_restoration_and_next_mutation(self):
        for mode in ("assume_unchanged", "skip_worktree"):
            with self.subTest(mode=mode):
                if self.log.exists():
                    self.log.unlink()
                self.spec["mutations"] = [self.mutation(), self.mutation(
                    id="other", old="pub fn other() -> u32 { 9 }", new="pub fn other() -> u32 { 8 }")]
                out, report = self.invoke(mode)
                self.assertEqual(out.returncode, 1, report)
                self.assertIn(report["status"], ("blocked", "error"))
                self.assertTrue(report["source_unchanged"])
                self.assertFalse(report["mutations"][0]["restored"])
                self.assertNotEqual(report["mutations"][0]["restore_diff_exit"], 0)
                self.assertEqual(len(self.calls()), 1)
                self.assertTrue(report["errors"], report)

    def test_repeated_runs_use_distinct_retained_workspaces_and_targets(self):
        _, first = self.invoke()
        _, second = self.invoke()
        self.assertNotEqual(first["workspace"], second["workspace"])
        self.assertNotEqual(first["cargo_target_dir"], second["cargo_target_dir"])
        self.assertTrue(Path(first["workspace"]).is_dir())
        self.assertTrue(Path(second["workspace"]).is_dir())

    def test_invalid_spec_and_mutation_inputs_are_rejected_before_cargo(self):
        invalid = [
            {"schema_version": 2, "mutations": [self.mutation()]},
            {"schema_version": 1, "mutations": []},
        ]
        for changes in ({"file": "../sentinel"}, {"file": str(self.sentinel)},
                        {"file": "missing.rs"}, {"old": "absent"},
                        {"new": OLD}, {"old": ""}, {"new": ""},
                        {"old": OLD + "\nsecond"}, {"new": NEW + "\nsecond"},
                        {"targets": []}, {"targets": [{"kind": "all"}]},
                        {"targets": [{"kind": "lib", "name": "bad"}]},
                        {"targets": [{"kind": "test"}]},
                        {"targets": [{"kind": "test", "name": "--all"}]},
                        {"targets": [{"kind": "bin", "name": "../bad"}]},
                        {"targets": [{"kind": "bench", "name": "*"}]},
                        {"targets": [{"kind": "example", "name": "a;touch sentinel"}]},
                        {"targets": [{"kind": "lib", "filter": "--ignored"}]},
                        {"targets": [{"kind": "lib", "filter": ""}]}):
            invalid.append({"schema_version": 1, "mutations": [self.mutation(**changes)]})
        for timeout in (0, -1, "soon", float("inf")):
            invalid.append(dict(self.spec, timeout_seconds=timeout))
        for value in invalid:
            with self.subTest(spec=value):
                self.assert_rejected(spec=value)

    def test_comment_lines_and_duplicate_old_lines_cannot_be_mutated(self):
        original = (self.source / "src/lib.rs").read_text()
        for line in ("// comment", "    /// docs", " //! docs", "  # comment", OLD):
            with self.subTest(line=line):
                content = original + line + "\n"
                (self.source / "src/lib.rs").write_text(content)
                self.git("add", "src/lib.rs")
                self.git("commit", "-qm", "invalid matching line")
                self.assert_rejected(spec={"schema_version": 1, "mutations": [self.mutation(old=line)]})

    def test_tracked_symlink_and_symlink_parent_are_rejected(self):
        (self.source / "alias.rs").symlink_to("src/lib.rs")
        (self.source / "aliasdir").symlink_to("src", target_is_directory=True)
        self.git("add", "alias.rs", "aliasdir")
        self.git("commit", "-qm", "symlinks")
        for file in ("alias.rs", "aliasdir/lib.rs"):
            with self.subTest(file=file):
                self.assert_rejected(spec={"schema_version": 1, "mutations": [self.mutation(file=file)]})

    def test_dirty_source_is_rejected_without_resetting_user_changes(self):
        for state in ("unstaged", "staged", "untracked"):
            with self.subTest(state=state):
                if state == "untracked":
                    (self.source / "user-note").write_text("retain me")
                else:
                    (self.source / "src/lib.rs").write_text(OLD + "\n// user's " + state + " edit\n")
                    if state == "staged":
                        self.git("add", "src/lib.rs")
                self.assert_rejected()
                if state == "untracked":
                    self.assertEqual((self.source / "user-note").read_text(), "retain me")
                # Commit fixture edits rather than resetting the source under test.
                self.git("add", ".")
                self.git("commit", "-qm", state)

    def test_target_must_be_strictly_inside_owned_target_with_real_path_boundaries(self):
        outside = self.base / "outside"
        outside.mkdir()
        (self.source / "target").mkdir()
        (self.source / "target/escape").symlink_to(outside, target_is_directory=True)
        values = [str(self.source / "target"), str(outside),
                  str(self.source / "target-sibling/cache"),
                  str(self.source / "target/escape/cache"), ""]
        for target in values:
            with self.subTest(target=target):
                self.assert_rejected(environment={"CARGO_TARGET_DIR": target})
        del self.env["CARGO_TARGET_DIR"]
        self.assert_rejected()

    def test_target_root_symlink_cannot_promote_an_external_directory_to_owned_target(self):
        outside = self.base / "outside"
        outside.mkdir()
        (self.source / "target").symlink_to(outside, target_is_directory=True)
        self.assert_rejected(environment={"CARGO_TARGET_DIR": str(self.source / "target/cache")})

    def test_malformed_json_is_rejected_without_running_cargo(self):
        path = self.base / "broken.json"
        path.write_text('{"schema_version":')
        before = self.snapshot()
        out = run_with_progress_budget(
            [sys.executable, "-I", str(RUNNER), "--spec", str(path)],
            cwd=self.source, env=self.env, log=self.log)
        self.assertEqual(self.snapshot(), before)
        self.assertNotEqual(out.returncode, 0)
        report = json.loads(out.stdout)
        self.assertIn(report["status"], ("blocked", "error"))
        self.assertTrue(report["errors"])
        self.assertEqual(self.calls(), [])

    def test_unborn_head_is_rejected_without_creating_a_commit(self):
        unborn = self.base / "unborn"
        unborn.mkdir()
        self.git("init", "-q", cwd=unborn)
        (unborn / ".gitignore").write_text("/target/\n")
        self.git("add", ".gitignore", cwd=unborn)
        index = (unborn / ".git/index").read_bytes()
        head = (unborn / ".git/HEAD").read_bytes()
        spec = self.base / "unborn-spec.json"
        spec.write_text(json.dumps(self.spec))
        env = dict(self.env, CARGO_TARGET_DIR=str(unborn / "target/cargo"))
        out = run_with_progress_budget(
            [sys.executable, "-I", str(RUNNER), "--spec", str(spec)],
            cwd=unborn, env=env, log=self.log)
        self.assertNotEqual(out.returncode, 0)
        report = json.loads(out.stdout)
        self.assertIn(report["status"], ("blocked", "error"))
        self.assertTrue(report["errors"])
        self.assertEqual((unborn / ".git/index").read_bytes(), index)
        self.assertEqual((unborn / ".git/HEAD").read_bytes(), head)
        self.assertEqual(self.calls(), [])

    def test_inherited_git_redirects_cannot_mutate_or_select_another_repository(self):
        decoy = self.base / "decoy"
        self.git("clone", "--no-hardlinks", str(self.source), str(decoy))
        decoy_before = (decoy / "src/lib.rs").read_bytes()
        out, report = self.invoke(environment={
            "GIT_DIR": str(decoy / ".git"), "GIT_WORK_TREE": str(decoy),
            "GIT_INDEX_FILE": str(decoy / ".git/index"),
            "GIT_OBJECT_DIRECTORY": str(decoy / ".git/objects")})
        self.assertEqual(out.returncode, 0, report)
        self.assertEqual((decoy / "src/lib.rs").read_bytes(), decoy_before)
        call = self.calls()[0]
        for key in ("GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY"):
            self.assertNotIn(key, call["git_env"])
        self.assertEqual(call["code"], NEW + "\npub fn other() -> u32 { 9 }\n")


if __name__ == "__main__":
    unittest.main()
