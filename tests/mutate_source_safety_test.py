"""Source-boundary regressions; compose the CLI fixture without inheriting tests."""

import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import unittest

import mutate_test as support


HIDDEN_BYTES = b"UNCOMMITTED SYNTHETIC BYTES\n"
PRIVATE_BYTES = b"PRIVATE SYNTHETIC SENTINEL: MUST NOT READ\n"
SOURCE_BYTES = (support.OLD + "\npub fn other() -> u32 { 9 }\n").encode()

# Keep the auditor's runpy/audit-hook boundary: observing only a rejection does
# not prove that the runner avoided reading the external file beforehand.
AUDITED_RUNNER = r'''import sys, pathlib, json, runpy, os, stat
if sys.platform == "darwin":
    import fcntl
sentinel = pathlib.Path(sys.argv[1]).resolve()
log = pathlib.Path(sys.argv[2])
runner, spec, ready = sys.argv[3:6]
def audit(event, args):
    if event == "open" and isinstance(args[0], (str, bytes, os.PathLike)):
        opened = pathlib.Path(os.fsdecode(args[0])).resolve()
        if opened == sentinel:
            with log.open("a") as stream:
                stream.write(json.dumps({"path": os.fsdecode(args[0]),
                                         "mode": args[1], "flags": args[2]}) + "\n")
sys.addaudithook(audit)

# The audit event reports a relative spelling but omits os.open's dir_fd.
# Observe the opened regular descriptor's metadata as well, without reading
# its contents or resolving any caller HOME/key paths.
original_open = os.open
def observed_open(path, flags, mode=0o777, *, dir_fd=None):
    fd = original_open(path, flags, mode, dir_fd=dir_fd)
    try:
        if stat.S_ISREG(os.fstat(fd).st_mode):
            actual = None
            if sys.platform.startswith("linux"):
                actual = os.readlink("/proc/self/fd/" + str(fd))
            elif sys.platform == "darwin":
                actual = os.fsdecode(
                    fcntl.fcntl(fd, 50, b"\0" * 1024).split(b"\0", 1)[0])
            # Native Windows FD observation is not exercised by this wrapper.
            if actual is not None and pathlib.Path(actual).resolve() == sentinel:
                with log.open("a") as stream:
                    stream.write(json.dumps({"path": actual, "mode": None,
                                             "flags": flags,
                                             "observer": "os.open-fd"}) + "\n")
    except BaseException:
        os.close(fd)
        raise
    return fd
os.open = observed_open
if original_open in os.supports_dir_fd:
    os.supports_dir_fd.add(observed_open)
pathlib.Path(ready).write_text("audit hook installed\n")
sys.argv = [runner, "--spec", spec]
runpy.run_path(runner, run_name="__main__")
'''


class SourceSafety(unittest.TestCase):
    def setUp(self):
        self.case = support.MutationCLI(
            "test_malformed_json_is_rejected_without_running_cargo")
        self.addCleanup(self.case.doCleanups)
        self.case.setUp()

    def physical_snapshot(self):
        """Walk physical source entries, never a replaced directory's target.

        The inherited snapshot lstat's leaves through arbitrary parents. Do not
        use it for these adversarial trees, even for the test's own assertions.
        """
        entries = {}

        def visit(directory, prefix=""):
            with os.scandir(directory) as children:
                for child in children:
                    if not prefix and child.name in (".git", "target"):
                        continue
                    name = prefix + child.name
                    mode = child.stat(follow_symlinks=False).st_mode
                    if stat.S_ISLNK(mode):
                        entries[name] = (mode, os.readlink(child.path))
                    elif stat.S_ISDIR(mode):
                        entries[name] = (mode, None)
                        visit(child.path, name + "/")
                    else:
                        self.assertTrue(stat.S_ISREG(mode), name)
                        entries[name] = (mode, Path(child.path).read_bytes())

        visit(self.case.source)
        git = self.case.source / ".git"
        return (entries, (git / "index").read_bytes(),
                (git / "HEAD").read_bytes(), self.case.git("show-ref"),
                self.case.git("rev-parse", "HEAD"))

    def invoke_blocked(self, *, sentinel=None, original=None):
        case = self.case
        spec = case.base / "source-safety-spec.json"
        spec.write_text(json.dumps(case.spec))
        before = self.physical_snapshot()
        external_before = sentinel.read_bytes() if sentinel is not None else None
        if sentinel is None:
            argv = [sys.executable, "-I", str(support.RUNNER), "--spec", str(spec)]
        else:
            log = case.base / "external-opens.jsonl"
            log.write_text("")
            ready = case.base / "audit-ready"
            argv = [sys.executable, "-I", "-c", AUDITED_RUNNER,
                    str(sentinel), str(log), str(support.RUNNER), str(spec), str(ready)]
        out = support.run_with_progress_budget(argv, cwd=case.source, env=case.env,
                                               log=case.log)
        # Check physical preservation before checking rejection: failures must
        # retain evidence even when the SUT unexpectedly reports completion.
        self.assertEqual(self.physical_snapshot(), before,
                         "runner changed physical source/index/HEAD/refs")
        self.assertEqual(case.sentinel.read_bytes(), b"external bytes must survive\n")
        if sentinel is not None:
            self.assertEqual(sentinel.read_bytes(), external_before)
            self.assertEqual(external_before, PRIVATE_BYTES)
            if original is not None:
                self.assertEqual(original.read_bytes(), SOURCE_BYTES)
            self.assertEqual(ready.read_text(), "audit hook installed\n")
            events = [json.loads(line) for line in log.read_text().splitlines()]
            self.assertEqual(events, [], "runner opened the external synthetic sentinel")
        report = json.loads(out.stdout)
        self.assertNotEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "blocked", report)
        self.assertTrue(report["errors"], report)
        self.assertEqual(case.calls(), [], "unsafe source reached Cargo")
        return report

    def hidden_neighbor(self, flags, *, changed=True, selected=False):
        name = "src/lib.rs" if selected else "kept.txt"
        for flag in flags:
            self.case.git("update-index", flag, name)
        if changed:
            (self.case.source / name).write_bytes(HIDDEN_BYTES)
        self.invoke_blocked()
        if changed:
            self.assertEqual((self.case.source / name).read_bytes(), HIDDEN_BYTES)
        if not selected:
            self.assertEqual((self.case.source / "src/lib.rs").read_bytes(), SOURCE_BYTES)

    def test_hidden_assume_unchanged_neighbor_is_blocked_without_cargo(self):
        self.hidden_neighbor(("--assume-unchanged",))

    def test_hidden_skip_worktree_neighbor_is_blocked_without_cargo(self):
        self.hidden_neighbor(("--skip-worktree",))

    def test_hidden_combined_flags_neighbor_is_blocked_without_cargo(self):
        self.hidden_neighbor(("--assume-unchanged", "--skip-worktree"))

    def test_clean_assume_unchanged_flag_is_rejected_without_index_rewrite(self):
        self.hidden_neighbor(("--assume-unchanged",), changed=False)

    def test_clean_skip_worktree_flag_is_rejected_without_index_rewrite(self):
        self.hidden_neighbor(("--skip-worktree",), changed=False)

    def test_clean_combined_flags_are_rejected_without_index_rewrite(self):
        self.hidden_neighbor(("--assume-unchanged", "--skip-worktree"), changed=False)

    def test_hidden_assume_unchanged_selected_file_is_blocked_without_cargo(self):
        self.hidden_neighbor(("--assume-unchanged",), selected=True)

    def test_hidden_skip_worktree_selected_file_is_blocked_without_cargo(self):
        self.hidden_neighbor(("--skip-worktree",), selected=True)

    def test_staged_ordinary_neighbor_is_blocked_without_index_rewrite(self):
        (self.case.source / "kept.txt").write_bytes(HIDDEN_BYTES)
        self.case.git("add", "kept.txt")
        self.invoke_blocked()
        self.assertEqual((self.case.source / "kept.txt").read_bytes(), HIDDEN_BYTES)

    def private_sentinel(self):
        private = self.case.base / "synthetic-private-home"
        private.mkdir()
        sentinel = private / "lib.rs"
        sentinel.write_bytes(PRIVATE_BYTES)
        return sentinel

    def test_directory_symlink_parent_blocks_with_zero_external_opens(self):
        sentinel = self.private_sentinel()
        source = self.case.source
        (source / "src").rename(source / "original-src")
        (source / "src").symlink_to(sentinel.parent, target_is_directory=True)
        self.invoke_blocked(sentinel=sentinel, original=source / "original-src/lib.rs")

    def test_multilevel_symlink_parent_blocks_with_zero_external_opens(self):
        source = self.case.source
        (source / "src").rename(source / "original-src")
        (source / "src/nested/deep").mkdir(parents=True)
        (source / "src/nested/deep/lib.rs").write_bytes(SOURCE_BYTES)
        self.case.git("add", ".")
        self.case.git("commit", "-qm", "multilevel source fixture")
        sentinel = self.private_sentinel()
        (source / "src/nested").rename(source / "original-nested")
        # The external tree matches the tracked multilevel suffix, ensuring an
        # unsafe reader would actually encounter the monitored sentinel.
        (sentinel.parent / "deep").mkdir()
        sentinel.rename(sentinel.parent / "deep/lib.rs")
        sentinel = sentinel.parent / "deep/lib.rs"
        (source / "src/nested").symlink_to(sentinel.parent.parent,
                                           target_is_directory=True)
        self.case.spec["mutations"][0]["file"] = "src/nested/deep/lib.rs"
        self.invoke_blocked(sentinel=sentinel, original=source / "original-nested/deep/lib.rs")

    def test_tracked_leaf_symlink_is_literal_with_zero_external_opens(self):
        sentinel = self.private_sentinel()
        source = self.case.source
        (source / "src").rename(source / "original-src")
        (source / "src").mkdir()
        (source / "src/lib.rs").symlink_to(sentinel)
        self.case.git("add", ".")
        self.case.git("commit", "-qm", "literal tracked leaf symlink fixture")
        self.invoke_blocked(sentinel=sentinel, original=source / "original-src/lib.rs")
        self.assertEqual(os.readlink(source / "src/lib.rs"), str(sentinel))

    def test_ordinary_blocked_input_final_snapshot_preserves_source(self):
        sentinel = self.private_sentinel()
        self.case.spec["mutations"][0]["old"] = "absent source line"
        self.invoke_blocked(sentinel=sentinel, original=self.case.source / "src/lib.rs")


if __name__ == "__main__":
    unittest.main()
