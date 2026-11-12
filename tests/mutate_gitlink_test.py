"""Gitlink cleanliness contracts, composed without inheriting the CLI tests."""

from contextlib import contextmanager
import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import unittest

import mutate_test as support


COMMITTED_BYTES = b"COMMITTED SYNTHETIC SUBMODULE\n"
HIDDEN_BYTES = b"UNCOMMITTED SYNTHETIC BYTES\n"


class GitlinkSafety(unittest.TestCase):
    @contextmanager
    def fixture(self):
        case = support.MutationCLI(
            "test_malformed_json_is_rejected_without_running_cargo")
        try:
            case.setUp()
            case.env.update(GIT_CONFIG_GLOBAL=os.devnull,
                            GIT_CONFIG_SYSTEM=os.devnull,
                            GIT_TERMINAL_PROMPT="0")
            yield case
        finally:
            case.doCleanups()

    def git(self, case, *args, cwd=None):
        # Readbacks must not refresh either source index as a side effect.
        return case.git("--no-optional-locks", *args, cwd=cwd)

    def child_repository(self, case, path):
        path.mkdir()
        self.git(case, "init", "-q", cwd=path)
        self.git(case, "config", "user.name", "Fixture", cwd=path)
        self.git(case, "config", "user.email", "fixture@example.invalid", cwd=path)
        (path / "lib.rs").write_bytes(COMMITTED_BYTES)
        self.git(case, "add", "lib.rs", cwd=path)
        self.git(case, "commit", "-qm", "synthetic gitlink child", cwd=path)
        return self.git(case, "rev-parse", "HEAD", cwd=path).decode().strip()

    def attach_gitlink(self, case, parent, name, oid, *, ignore="all"):
        (parent / ".gitmodules").write_text(
            '[submodule "' + name + '"]\n\tpath = ' + name +
            '\n\turl = file:///synthetic/unused\n')
        self.git(case, "add", ".gitmodules", cwd=parent)
        self.git(case, "update-index", "--add", "--cacheinfo", "160000",
                 oid, name, cwd=parent)
        self.git(case, "commit", "-qm", "synthetic tracked gitlink", cwd=parent)
        self.git(case, "config", "submodule." + name + ".ignore", ignore,
                 cwd=parent)
        self.assertEqual(
            self.git(case, "ls-files", "--stage", "--", name, cwd=parent),
            ("160000 " + oid + " 0\t" + name + "\n").encode())

    def initialized_child(self, case, *, ignore="all"):
        child = case.source / "vendor"
        oid = self.child_repository(case, child)
        self.attach_gitlink(case, case.source, "vendor", oid, ignore=ignore)
        return child

    def uninitialized_child(self, case):
        # Obtain a real commit ID without initializing/fetching the gitlink.
        oid = self.child_repository(case, case.base / "synthetic-child")
        child = case.source / "vendor"
        child.mkdir()
        self.attach_gitlink(case, case.source, "vendor", oid)
        self.assertFalse((child / ".git").exists())
        return child

    def literal_tree(self, root, *, source=False):
        entries = {}

        def visit(directory, prefix=""):
            with os.scandir(directory) as children:
                for child in children:
                    if source and (child.name == ".git" or
                                   (not prefix and child.name == "target")):
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

        visit(root)
        return entries

    def snapshot(self, case, repositories):
        metadata = {}
        for repository in repositories:
            git_dir = repository / ".git"
            raw = {name: (git_dir / name).read_bytes()
                   if (git_dir / name).exists() else None
                   for name in ("index", "HEAD", "config", "packed-refs")}
            metadata[str(repository)] = (
                raw, self.literal_tree(git_dir / "refs"),
                self.git(case, "rev-parse", "HEAD", cwd=repository),
                self.git(case, "show-ref", cwd=repository),
                self.git(case, "ls-files", "--stage", "-v", "-z", cwd=repository),
                self.git(case, "status", "--porcelain=v1", "--untracked-files=all",
                         cwd=repository))
        return self.literal_tree(case.source, source=True), metadata

    def invoke(self, case, *children):
        spec = case.base / "gitlink-spec.json"
        spec.write_text(json.dumps(case.spec))
        repositories = (case.source, *children)
        before = self.snapshot(case, repositories)
        out = subprocess.run(
            [sys.executable, "-I", str(support.RUNNER), "--spec", str(spec)],
            cwd=case.source, env=dict(case.env, FAKE_MODE="killed"),
            capture_output=True, text=True, timeout=30)
        # Preserve literal evidence even when a dirty source is wrongly accepted.
        self.assertEqual(self.snapshot(case, repositories), before,
                         "runner changed parent/child bytes, raw indexes, HEADs, refs or config")
        self.assertEqual(case.sentinel.read_bytes(), b"external bytes must survive\n")
        try:
            report = json.loads(out.stdout)
        except ValueError:
            self.fail("CLI must emit one JSON object: " +
                      repr((out.returncode, out.stdout, out.stderr)))
        self.assertEqual(report["schema_version"], 1)
        return out, report

    def assert_blocked(self, case, *children):
        out, report = self.invoke(case, *children)
        self.assertNotEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "blocked", report)
        self.assertTrue(report["errors"], report)
        self.assertEqual(case.calls(), [], "dirty gitlink source reached Cargo")

    def assert_allowed(self, case, *children):
        out, report = self.invoke(case, *children)
        self.assertEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "completed", report)
        self.assertEqual(report["errors"], [], report)
        self.assertTrue(report["source_unchanged"], report)
        self.assertEqual(len(case.calls()), 1)
        self.assertTrue(report["mutations"][0]["restored"], report)
        self.assertEqual(report["mutations"][0]["targets"][0]["status"], "FAILED")

    def test_dirty_child_hidden_by_ignore_all_is_blocked_without_cargo(self):
        with self.fixture() as case:
            child = self.initialized_child(case)
            (child / "lib.rs").write_bytes(HIDDEN_BYTES)
            self.assertEqual(self.git(case, "status", "--porcelain=v1",
                                      "--untracked-files=all"), b"")
            self.assertIn(b"vendor", self.git(case, "status", "--porcelain=v1",
                                             "--ignore-submodules=none"))
            self.assertIn(b"lib.rs", self.git(case, "status", "--porcelain=v1",
                                             cwd=child))
            self.assert_blocked(case, child)

    def test_visible_dirty_child_is_blocked_without_cargo(self):
        with self.fixture() as case:
            child = self.initialized_child(case, ignore="none")
            (child / "lib.rs").write_bytes(HIDDEN_BYTES)
            self.assertIn(b"vendor", self.git(case, "status", "--porcelain=v1"))
            self.assert_blocked(case, child)

    def test_clean_initialized_child_is_allowed_without_source_changes(self):
        with self.fixture() as case:
            child = self.initialized_child(case)
            self.assertEqual(self.git(case, "status", "--porcelain=v1", cwd=child), b"")
            self.assert_allowed(case, child)

    def test_child_head_index_staged_and_untracked_changes_are_blocked(self):
        for state in ("head", "index", "staged", "untracked"):
            with self.subTest(state=state), self.fixture() as case:
                child = self.initialized_child(case)
                if state == "head":
                    (child / "lib.rs").write_bytes(HIDDEN_BYTES)
                    self.git(case, "add", "lib.rs", cwd=child)
                    self.git(case, "commit", "-qm", "child-only commit", cwd=child)
                    self.assertEqual(self.git(case, "status", "--porcelain=v1",
                                              cwd=child), b"")
                elif state == "index":
                    # Keep HEAD and physical bytes intact; remove only the index entry.
                    self.git(case, "update-index", "--force-remove", "lib.rs", cwd=child)
                    self.assertEqual((child / "lib.rs").read_bytes(), COMMITTED_BYTES)
                elif state == "staged":
                    (child / "lib.rs").write_bytes(HIDDEN_BYTES)
                    self.git(case, "add", "lib.rs", cwd=child)
                else:
                    (child / "user-note").write_bytes(HIDDEN_BYTES)
                self.assertEqual(self.git(case, "status", "--porcelain=v1"), b"")
                self.assert_blocked(case, child)

    def test_dirty_child_hidden_by_ignore_dirty_is_blocked_without_cargo(self):
        with self.fixture() as case:
            child = self.initialized_child(case, ignore="dirty")
            (child / "lib.rs").write_bytes(HIDDEN_BYTES)
            self.assertEqual(self.git(case, "status", "--porcelain=v1"), b"")
            self.assert_blocked(case, child)

    def test_hidden_child_index_flags_are_blocked_without_rewriting_index(self):
        flags = (("--assume-unchanged",), ("--skip-worktree",),
                 ("--assume-unchanged", "--skip-worktree"))
        for flag_set in flags:
            for changed in (False, True):
                with self.subTest(flags=flag_set, changed=changed), self.fixture() as case:
                    child = self.initialized_child(case, ignore="none")
                    for flag in flag_set:
                        self.git(case, "update-index", flag, "lib.rs", cwd=child)
                    if changed:
                        (child / "lib.rs").write_bytes(HIDDEN_BYTES)
                    self.assertEqual(self.git(case, "status", "--porcelain=v1", cwd=child), b"")
                    self.assertEqual(self.git(case, "status", "--porcelain=v1",
                                              "--ignore-submodules=none"), b"")
                    self.assert_blocked(case, child)

    def test_nested_gitlink_hidden_bytes_are_blocked_without_cargo(self):
        with self.fixture() as case:
            child = case.source / "vendor"
            self.child_repository(case, child)
            nested = child / "nested"
            nested_oid = self.child_repository(case, nested)
            self.attach_gitlink(case, child, "nested", nested_oid)
            child_oid = self.git(case, "rev-parse", "HEAD", cwd=child).decode().strip()
            self.attach_gitlink(case, case.source, "vendor", child_oid)
            self.git(case, "update-index", "--assume-unchanged", "lib.rs", cwd=nested)
            (nested / "lib.rs").write_bytes(HIDDEN_BYTES)
            self.assertEqual(self.git(case, "status", "--porcelain=v1", cwd=nested), b"")
            self.assertEqual(self.git(case, "status", "--porcelain=v1", cwd=child), b"")
            self.assertEqual(self.git(case, "status", "--porcelain=v1"), b"")
            self.assert_blocked(case, child, nested)

    def test_empty_uninitialized_gitlink_is_allowed_as_metadata_only(self):
        with self.fixture() as case:
            child = self.uninitialized_child(case)
            self.assert_allowed(case)
            self.assertEqual(list(child.iterdir()), [], "runner initialized the gitlink")

    def test_nonempty_uninitialized_gitlink_is_blocked_without_cargo(self):
        with self.fixture() as case:
            child = self.uninitialized_child(case)
            (child / "lib.rs").write_bytes(HIDDEN_BYTES)
            self.assertEqual(self.git(case, "status", "--porcelain=v1"), b"")
            self.assert_blocked(case)
            self.assertFalse((child / ".git").exists())

    def test_clone_gitlink_poison_blocks_restoration_and_next_mutation(self):
        with self.fixture() as case:
            child = self.uninitialized_child(case)
            case.spec["mutations"].append(case.mutation(
                id="other", old="pub fn other() -> u32 { 9 }",
                new="pub fn other() -> u32 { 8 }"))
            # Only this fixture's Cargo collaborator poisons the clone. Keep
            # the common harness unchanged and let it record real clone state.
            cargo = case.bin / "cargo"
            cargo.write_text("#!" + sys.executable + "\n" +
                             support.FAKE_CARGO.replace(
                                 'mode = os.environ.get("FAKE_MODE", "killed")',
                                 '(p / "vendor/clone-note").write_bytes('
                                 'b"UNCOMMITTED CLONE GITLINK BYTES\\n")\n'
                                 'mode = os.environ.get("FAKE_MODE", "killed")'))
            out, report = self.invoke(case)
            self.assertNotEqual(out.returncode, 0, report)
            self.assertEqual(report["status"], "blocked", report)
            self.assertTrue(report["source_unchanged"], report)
            self.assertTrue(report["errors"], report)
            self.assertEqual(len(case.calls()), 1, "runner continued after clone gitlink poison")
            self.assertEqual(len(report["mutations"]), 1, report)
            mutation = report["mutations"][0]
            self.assertFalse(mutation["restored"], report)
            self.assertNotEqual(mutation["restore_diff_exit"], 0, report)
            workspace = Path(report["workspace"])
            self.assertEqual((workspace / "vendor/clone-note").read_bytes(),
                             b"UNCOMMITTED CLONE GITLINK BYTES\n")
            self.assertEqual(list(child.iterdir()), [], "clone poison reached source gitlink")


if __name__ == "__main__":
    unittest.main()
