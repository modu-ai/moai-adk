"""Tests for tools/sandbox_write_probe.py (Issue #3068 P8, AC9).

Every case runs the probe on a temporary base folder this test creates, and the "real home" of
the refusal cases is a temporary folder handed to the tool through its home lookup. None runs
the tool on the real HOME or on any path the test did not make; the real-HOME measurement is
guarded (Q4) and is not a test.
"""

import errno
import importlib.util
import json
import os
import pwd
import re
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

REPO_ROOT = Path(__file__).resolve().parent.parent
TOOL = REPO_ROOT / "tools" / "sandbox_write_probe.py"

STEPS = ["mkdir_tree", "create_file", "sqlite_write", "sqlite_read", "hard_link", "rename",
         "advisory_lock", "fsync_dir", "cleanup"]
TOP_KEYS = {"tool", "version", "base", "project_id", "python", "uid", "steps", "all_ok",
            "left_behind", "writes_outside_base"}
STEP_KEYS = {"name", "ok", "errno", "error"}


def load_module():
    spec = importlib.util.spec_from_file_location("sandbox_write_probe", TOOL)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


MODULE = load_module()


def tmp_dir(test, prefix="moai-test-probe-"):
    """A temporary folder (real path) that is removed when the test ends."""
    holder = tempfile.TemporaryDirectory(prefix=prefix)
    test.addCleanup(holder.cleanup)
    return os.path.realpath(holder.name)


def snapshot(*roots):
    """Every path below each root with its type, so "nothing created" is a comparison."""
    seen = []

    def visit(folder):
        for name in sorted(os.listdir(folder)):
            path = os.path.join(folder, name)
            is_link = os.path.islink(path)
            seen.append((path, "link" if is_link else "x"))
            if not is_link and os.path.isdir(path):
                visit(path)

    for root in roots:
        visit(root)
    return sorted(seen)


def steps_by_name(out):
    return {s["name"]: s for s in out["steps"]}


def run_cli(*args):
    """The tool as a child process in isolated mode with an empty environment."""
    return subprocess.run([sys.executable, "-I", str(TOOL), *args], capture_output=True,
                          text=True, env={})


class ProbeRunTest(unittest.TestCase):
    def setUp(self):
        self.base = tmp_dir(self)

    def probe(self, *extra):
        return MODULE.execute(["--base", self.base, *extra], home_lookup=lambda: tmp_dir(self))

    def test_every_step_is_ok_and_nothing_is_left_behind(self):
        code, out = self.probe()
        self.assertEqual(code, 0)
        self.assertEqual([s["name"] for s in out["steps"]], STEPS)
        failed = [s for s in out["steps"] if not s["ok"]]
        self.assertEqual(failed, [])
        self.assertTrue(out["all_ok"])
        self.assertEqual(out["left_behind"], [])
        self.assertEqual(os.listdir(self.base), [])

    def test_json_shape(self):
        code, out = self.probe()
        self.assertEqual(code, 0)
        self.assertEqual(set(out), TOP_KEYS)
        self.assertEqual(out["tool"], "sandbox_write_probe")
        self.assertEqual(out["version"], 1)
        self.assertEqual(out["base"], self.base)
        self.assertEqual(out["uid"], os.getuid())
        self.assertEqual(out["python"], sys.version.split()[0])
        self.assertIs(out["writes_outside_base"], False)
        for step in out["steps"]:
            self.assertEqual(set(step), STEP_KEYS)
            self.assertIs(step["ok"], True)
            self.assertIsNone(step["errno"])
            self.assertIsNone(step["error"])
        json.dumps(out)

    def test_default_project_id_is_probe_and_eight_hex_digits(self):
        _, first = self.probe()
        _, second = self.probe()
        self.assertRegex(first["project_id"], r"^probe-[0-9a-f]{8}$")
        self.assertNotEqual(first["project_id"], second["project_id"])

    def test_given_project_id_is_used(self):
        code, out = self.probe("--project-id", "my.probe_1-x")
        self.assertEqual(code, 0)
        self.assertEqual(out["project_id"], "my.probe_1-x")

    def test_project_id_length_boundary(self):
        for length in (1, 64):
            code, out = self.probe("--project-id", "a" * length)
            self.assertEqual((code, out["all_ok"]), (0, True), length)
        before = snapshot(self.base)
        for length in (65, 100):
            code, out = self.probe("--project-id", "a" * length)
            self.assertEqual((code, out["refused"]), (2, "bad_project_id"), length)
        self.assertEqual(snapshot(self.base), before)

    def test_bad_project_ids_are_refused(self):
        before = snapshot(self.base)
        for bad in ("", ".", "..", "a/b", "a b", "a\0b", "../x", "é"):
            code, out = self.probe("--project-id", bad)
            self.assertEqual((code, out["refused"]), (2, "bad_project_id"), repr(bad))
        self.assertEqual(snapshot(self.base), before)

    def test_the_steps_really_make_the_ledger_files(self):
        seen = {}
        original = MODULE.Probe.fsync_dir

        def spy(probe):
            seen["names"] = sorted(os.listdir(probe.project))
            seen["modes"] = {
                "moai": os.stat(probe.moai).st_mode & 0o777,
                "projects": os.stat(probe.projects).st_mode & 0o777,
                "project": os.stat(probe.project).st_mode & 0o777,
                "ledger": os.stat(probe.path("ledger.db")).st_mode & 0o777,
                "lock": os.stat(probe.path(".recover.lock")).st_mode & 0o777,
            }
            seen["links"] = os.stat(probe.path("ledger.db")).st_nlink
            original(probe)

        with mock.patch.object(MODULE.Probe, "fsync_dir", spy):
            code, out = self.probe()
        self.assertEqual(code, 0)
        self.assertEqual(seen["names"], [".recover.lock", "ledger.db", "ledger.db.damaged-probe",
                                         "ledger.db.next"])
        self.assertEqual(seen["links"], 2)
        self.assertEqual(seen["modes"], {"moai": 0o700, "projects": 0o700, "project": 0o700,
                                         "ledger": 0o600, "lock": 0o600})
        self.assertTrue(out["all_ok"])

    def test_a_failed_step_is_reported_and_later_steps_still_run(self):
        err = OSError(errno.EPERM, "Operation not permitted")
        with mock.patch.object(MODULE.os, "link", side_effect=err):
            code, out = self.probe()
        steps = steps_by_name(out)
        self.assertEqual(code, 0)
        self.assertEqual(steps["hard_link"],
                         {"name": "hard_link", "ok": False, "errno": errno.EPERM,
                          "error": "Operation not permitted"})
        for name in STEPS:
            if name != "hard_link":
                self.assertTrue(steps[name]["ok"], name)
        self.assertFalse(out["all_ok"])
        self.assertEqual(out["left_behind"], [])
        self.assertEqual(os.listdir(self.base), [])

    def test_steps_that_need_a_failed_step_are_skipped_not_crashed(self):
        err = OSError(errno.EACCES, "Permission denied")
        with mock.patch.object(MODULE.Probe, "create_file", side_effect=err):
            code, out = self.probe()
        steps = steps_by_name(out)
        self.assertEqual(code, 0)
        self.assertEqual((steps["create_file"]["ok"], steps["create_file"]["errno"]),
                         (False, errno.EACCES))
        for name in ("sqlite_write", "sqlite_read", "hard_link"):
            self.assertFalse(steps[name]["ok"], name)
            self.assertIsNone(steps[name]["errno"], name)
            self.assertTrue(steps[name]["error"].startswith("skipped: "), name)
        for name in ("mkdir_tree", "rename", "advisory_lock", "fsync_dir", "cleanup"):
            self.assertTrue(steps[name]["ok"], name)
        self.assertFalse(out["all_ok"])
        self.assertEqual(os.listdir(self.base), [])

    @unittest.skipIf(os.geteuid() == 0, "root ignores folder permissions")
    def test_a_base_that_cannot_be_written_fails_mkdir_tree_with_eacces(self):
        os.chmod(self.base, 0o500)
        self.addCleanup(os.chmod, self.base, 0o700)
        code, out = self.probe()
        steps = steps_by_name(out)
        self.assertEqual(code, 0)
        self.assertEqual((steps["mkdir_tree"]["ok"], steps["mkdir_tree"]["errno"]),
                         (False, errno.EACCES))
        self.assertTrue(steps["mkdir_tree"]["error"])
        for name in STEPS[1:-1]:
            self.assertFalse(steps[name]["ok"], name)
            self.assertTrue(steps[name]["error"].startswith("skipped: "), name)
        self.assertTrue(steps["cleanup"]["ok"])
        self.assertFalse(out["all_ok"])
        self.assertEqual(out["left_behind"], [])
        self.assertEqual(os.listdir(self.base), [])

    def test_cleanup_reports_what_it_could_not_remove(self):
        real_unlink = os.unlink

        def deny_lock(path, *args, **kwargs):
            if str(path).endswith(".recover.lock"):
                raise PermissionError(errno.EACCES, "Permission denied")
            return real_unlink(path, *args, **kwargs)

        with mock.patch.object(MODULE.os, "unlink", deny_lock):
            code, out = self.probe("--project-id", "stuck")
        project = os.path.join(self.base, ".moai", "projects", "stuck")
        lock = os.path.join(project, ".recover.lock")
        self.addCleanup(lambda: [os.rmdir(d) for d in (
            project, os.path.dirname(project), os.path.dirname(os.path.dirname(project)))])
        self.addCleanup(lambda: os.path.lexists(lock) and os.unlink(lock))
        steps = steps_by_name(out)
        self.assertEqual(code, 0)
        self.assertFalse(steps["cleanup"]["ok"])
        self.assertEqual(steps["cleanup"]["errno"], errno.EACCES)
        self.assertFalse(out["all_ok"])
        self.assertIn(lock, out["left_behind"])
        self.assertIn(project, out["left_behind"])

    def test_cleanup_is_never_a_recursive_delete(self):
        original = MODULE.Probe.fsync_dir
        stray = {}

        def plant(probe):
            stray["path"] = probe.path("not-made-by-the-probe.txt")
            with open(stray["path"], "w", encoding="utf-8") as f:
                f.write("keep")
            original(probe)

        with mock.patch.object(MODULE.Probe, "fsync_dir", plant):
            code, out = self.probe("--project-id", "keepme")
        project = os.path.join(self.base, ".moai", "projects", "keepme")
        self.addCleanup(lambda: [os.rmdir(d) for d in (
            project, os.path.dirname(project), os.path.dirname(os.path.dirname(project)))])
        self.addCleanup(os.unlink, stray["path"])
        self.assertEqual(code, 0)
        self.assertTrue(os.path.exists(stray["path"]))
        self.assertIn(project, out["left_behind"])
        self.assertNotIn(stray["path"], out["left_behind"])
        self.assertFalse(steps_by_name(out)["cleanup"]["ok"])
        self.assertEqual(sorted(os.listdir(project)), ["not-made-by-the-probe.txt"])


class ProbeRefusalTest(unittest.TestCase):
    def setUp(self):
        self.base = tmp_dir(self)
        self.home = tmp_dir(self, "moai-test-fakehome-")

    def refuse(self, argv, word, watch=()):
        """The run is refused with `word`, exit 2, and nothing below `watch` changed."""
        before = snapshot(*watch)
        code, out = MODULE.execute(argv, home_lookup=lambda: self.home)
        self.assertEqual(code, 2, out)
        self.assertEqual(out["refused"], word, out)
        self.assertEqual(set(out), {"refused", "detail"})
        self.assertEqual(snapshot(*watch), before)
        return out

    def test_missing_base(self):
        self.refuse([], "base_missing", [self.base])
        self.refuse(["--base", ""], "base_missing", [self.base])

    def test_relative_base(self):
        for rel in ("relative/dir", ".", "..", "./x", "x"):
            self.refuse(["--base", rel], "base_not_absolute", [self.base])

    def test_root_base(self):
        for root in ("/", "//", "/."):
            self.refuse(["--base", root], "root")

    def test_base_that_is_not_an_existing_directory(self):
        self.refuse(["--base", os.path.join(self.base, "absent")], "base_not_a_directory",
                    [self.base])
        file_path = os.path.join(self.base, "a-file")
        Path(file_path).write_text("x", encoding="utf-8")
        self.refuse(["--base", file_path], "base_not_a_directory", [self.base])

    def test_moai_already_in_the_base(self):
        for make in (os.mkdir, lambda p: Path(p).write_text("x", encoding="utf-8"),
                     lambda p: os.symlink(os.path.join(self.base, "nowhere"), p)):
            base = tmp_dir(self)
            make(os.path.join(base, ".moai"))
            self.refuse(["--base", base], "moai_exists", [base])

    def test_base_that_is_a_symlink(self):
        target = tmp_dir(self)
        link = os.path.join(self.base, "link")
        os.symlink(target, link)
        self.refuse(["--base", link], "symlink", [self.base, target])
        dangling = os.path.join(self.base, "dangling")
        os.symlink(os.path.join(self.base, "nowhere"), dangling)
        self.refuse(["--base", dangling], "symlink", [self.base])

    def test_the_real_home_itself(self):
        self.refuse(["--base", self.home], "real_home", [self.home])

    def test_real_home_wins_over_an_existing_moai(self):
        os.mkdir(os.path.join(self.home, ".moai"))
        self.refuse(["--base", self.home], "real_home", [self.home])

    def test_each_protected_folder_and_what_is_inside_it(self):
        for name in (".moai", ".moai-cli", ".claude", ".codex", "go"):
            folder = os.path.join(self.home, name)
            os.makedirs(os.path.join(folder, "deep", "er"))
            for base in (folder, os.path.join(folder, "deep"), os.path.join(folder, "deep", "er")):
                self.refuse(["--base", base], "real_home", [self.home])

    def test_names_next_to_the_protected_ones_are_allowed(self):
        # the edge on the allowed side: `go2`, `gopher`, `.moaix` are not `go` or `.moai`
        for name in ("go2", "gopher", ".moaix", ".claude-other", "work"):
            base = os.path.join(self.home, name)
            os.mkdir(base)
            code, out = MODULE.execute(["--base", base], home_lookup=lambda: self.home)
            self.assertEqual((code, out["all_ok"]), (0, True), name)
            self.assertEqual(os.listdir(base), [], name)

    def test_a_case_variant_of_a_protected_folder_is_refused_on_a_case_blind_volume(self):
        os.mkdir(os.path.join(self.home, "go"))
        if not os.path.exists(os.path.join(self.home, "GO")):
            self.skipTest("this volume is case-sensitive")
        self.refuse(["--base", os.path.join(self.home, "GO")], "real_home", [self.home])

    def test_real_home_is_found_through_a_symlinked_parent(self):
        alias = os.path.join(tmp_dir(self), "alias")
        os.symlink(self.home, alias)
        os.mkdir(os.path.join(self.home, ".claude"))
        self.refuse(["--base", os.path.join(alias, ".claude")], "real_home", [self.home])

    def test_a_home_that_cannot_be_looked_up_refuses(self):
        def broken():
            raise KeyError("getpwuid(): uid not found")

        before = snapshot(self.base)
        code, out = MODULE.execute(["--base", self.base], home_lookup=broken)
        self.assertEqual((code, out["refused"]), (2, "real_home"))
        self.assertEqual(snapshot(self.base), before)

    def test_usage_errors(self):
        self.refuse(["--base", self.base, "--base", self.base], "usage", [self.base])
        self.refuse(["--base"], "usage")
        self.refuse(["--base", self.base, "--project-id"], "usage", [self.base])
        self.refuse(["--home", self.base], "usage", [self.base])
        self.refuse([self.base], "usage", [self.base])

    def test_the_default_home_lookup_is_the_account_entry_not_the_environment(self):
        entry = pwd.getpwuid(os.getuid()).pw_dir
        with mock.patch.dict(os.environ, {"HOME": self.home}):
            self.assertEqual(MODULE.real_home(), entry)
            self.assertNotEqual(MODULE.real_home(), self.home)


class SourceTest(unittest.TestCase):
    def test_the_tool_reads_no_environment_and_opens_no_network(self):
        text = TOOL.read_text(encoding="utf-8")
        self.assertIsNone(re.search(r"os\.environ|getenv|environb", text))
        self.assertIsNone(re.search(r"^\s*(import|from)\s+(socket|ssl|urllib|http|requests)\b",
                                    text, re.M))


class ProcessTest(unittest.TestCase):
    def test_runs_in_isolated_mode_with_an_empty_environment(self):
        base = tmp_dir(self)
        proc = run_cli("--base", base)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(proc.stderr, "")
        self.assertEqual(len(proc.stdout.strip().splitlines()), 1)
        out = json.loads(proc.stdout)
        self.assertEqual(set(out), TOP_KEYS)
        self.assertTrue(out["all_ok"], out)
        self.assertEqual(out["left_behind"], [])
        self.assertEqual(os.listdir(base), [])

    def test_a_refusal_exits_2_with_one_json_line_and_creates_nothing(self):
        base = tmp_dir(self)
        before = snapshot(base)
        for args, word in ((("--base", "relative"), "base_not_absolute"), ((), "base_missing"),
                           (("--base", os.path.join(base, "absent")), "base_not_a_directory")):
            proc = run_cli(*args)
            self.assertEqual(proc.returncode, 2, args)
            lines = proc.stdout.strip().splitlines()
            self.assertEqual(len(lines), 1, args)
            self.assertEqual(json.loads(lines[0])["refused"], word)
        self.assertEqual(snapshot(base), before)

    def test_a_failing_step_still_exits_0(self):
        if os.geteuid() == 0:
            self.skipTest("root ignores folder permissions")
        base = tmp_dir(self)
        os.chmod(base, 0o500)
        self.addCleanup(os.chmod, base, 0o700)
        proc = run_cli("--base", base)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        out = json.loads(proc.stdout)
        self.assertFalse(out["all_ok"])
        self.assertEqual(steps_by_name(out)["mkdir_tree"]["errno"], errno.EACCES)


if __name__ == "__main__":
    unittest.main()
