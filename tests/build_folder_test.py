"""tools/build_folder.py (card C-82, user decision 2026-11-17 "빌드 폴더를 한 곳으로 모으기").

An agent or audit build puts CARGO_TARGET_DIR outside the worktree, in one folder per worktree
under `<HOME>/.moai/build/<project>-<worktree name>/`, always with CARGO_INCREMENTAL=0, and
deletes that folder's cargo output when its task ends. `remove` deletes only the folders a
user picked from an `optimize-plan` preview. Every test uses a temporary HOME and a temporary
git repository; the real HOME is never read or written.
"""
import contextlib
import errno
import fcntl
import importlib.util
import io
import json
import os
import shlex
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest import mock

ROOT = Path(__file__).resolve().parent.parent
TOOL = ROOT / "tools" / "build_folder.py"


def git(cwd, *args):
    env = dict(os.environ, GIT_CONFIG_GLOBAL="/dev/null", GIT_CONFIG_SYSTEM="/dev/null")
    return subprocess.run(
        ["git", "-c", "user.name=t", "-c", "user.email=t@example.invalid", *args],
        cwd=cwd, env=env, check=True, capture_output=True, text=True,
    ).stdout


def cargo_output(folder):
    (folder / "debug" / "deps").mkdir(parents=True)
    (folder / "debug" / ".cargo-lock").write_text("")
    (folder / "debug" / "deps" / "lib.rlib").write_bytes(b"x" * 4096)
    (folder / "CACHEDIR.TAG").write_text("Signature: 8a477f597d28d172789f06886806bc55\n")
    (folder / "tmp").mkdir()


class Fixture:
    def __init__(self):
        self.tmp = tempfile.TemporaryDirectory()
        base = Path(os.path.realpath(self.tmp.name))
        self.home = base / "home"
        self.home.mkdir()
        self.repo = base / "repo"
        self.repo.mkdir()
        git(self.repo, "init", "-q", "-b", "main")
        (self.repo / "a.txt").write_text("a\n")
        git(self.repo, "add", "a.txt")
        git(self.repo, "commit", "-q", "-m", "init")
        self.wt = base / "wts" / "agent-7"
        git(self.repo, "worktree", "add", "-q", "-b", "card/1-x", str(self.wt))

    def run(self, *args, cwd=None, home=None, extra_env=None):
        env = {
            "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
            "HOME": str(home if home is not None else self.home),
            "GIT_CONFIG_GLOBAL": "/dev/null",
            "GIT_CONFIG_SYSTEM": "/dev/null",
            **(extra_env or {}),
        }
        return subprocess.run(
            ["python3", "-I", str(TOOL), *args],
            cwd=cwd or self.wt, env=env, capture_output=True, text=True,
        )

    def folder(self, name):
        return self.home / ".moai" / "build" / name

    def close(self):
        # Only HOME itself is ever made read-only (the unwritable-HOME test).
        self.home.chmod(0o755)
        self.tmp.cleanup()


def exports(stdout):
    out = {}
    for line in stdout.splitlines():
        words = shlex.split(line)
        assert words[0] == "export", line
        key, _, value = words[1].partition("=")
        out[key] = value
    return out


class EnvTest(unittest.TestCase):
    def setUp(self):
        self.fx = Fixture()
        self.addCleanup(self.fx.close)

    def test_env_names_one_folder_per_worktree_and_turns_incremental_off(self):
        out = self.fx.run("env")
        self.assertEqual(out.returncode, 0, out.stderr)
        env = exports(out.stdout)
        self.assertEqual(env["CARGO_TARGET_DIR"], str(self.fx.folder("repo-agent-7")))
        self.assertEqual(env["CARGO_INCREMENTAL"], "0")
        record = json.loads((self.fx.folder("repo-agent-7") / ".moai-build.json").read_text())
        self.assertEqual(record["schema_version"], 1)
        self.assertEqual(record["worktree"], str(self.fx.wt))
        self.assertEqual(record["git_common_dir"], str(self.fx.repo / ".git"))
        # The primary checkout gets the project name alone; two worktrees never share a folder.
        primary = exports(self.fx.run("env", cwd=self.fx.repo).stdout)
        self.assertEqual(primary["CARGO_TARGET_DIR"], str(self.fx.folder("repo")))
        self.assertNotEqual(primary["CARGO_TARGET_DIR"], env["CARGO_TARGET_DIR"])
        # Asking again changes nothing and prints the same lines.
        again = self.fx.run("env")
        self.assertEqual((again.returncode, again.stdout), (0, out.stdout))

    def test_env_refuses_a_linked_build_root_and_writes_nothing_through_it(self):
        elsewhere = self.fx.home / "elsewhere"
        elsewhere.mkdir()
        (self.fx.home / ".moai").mkdir()
        (self.fx.home / ".moai" / "build").symlink_to(elsewhere)
        out = self.fx.run("env")
        self.assertEqual(out.returncode, 1)
        self.assertEqual(out.stdout, "")
        self.assertIn("is a link", out.stderr)
        self.assertEqual(list(elsewhere.iterdir()), [])

    def test_env_refuses_a_linked_folder_or_record(self):
        target = self.fx.home / "t"
        target.mkdir()
        self.fx.folder("x").parent.mkdir(parents=True)
        self.fx.folder("repo-agent-7").symlink_to(target)
        out = self.fx.run("env")
        self.assertEqual((out.returncode, out.stdout), (1, ""))
        self.assertEqual(list(target.iterdir()), [])
        self.fx.folder("repo-agent-7").unlink()
        self.fx.folder("repo-agent-7").mkdir()
        (self.fx.folder("repo-agent-7") / ".moai-build.json").symlink_to(target / "r.json")
        out = self.fx.run("env")
        self.assertEqual((out.returncode, out.stdout), (1, ""))
        self.assertFalse((target / "r.json").exists())

    @unittest.skipIf(os.geteuid() == 0, "root can write anywhere")
    def test_env_refuses_an_unwritable_home(self):
        self.fx.home.chmod(0o500)
        out = self.fx.run("env")
        self.assertEqual((out.returncode, out.stdout), (1, ""))
        self.assertIn("not writable", out.stderr)
        self.assertFalse((self.fx.home / ".moai").exists())

    def test_env_refuses_a_relative_or_missing_home(self):
        for home in ["relative/home", ""]:
            out = self.fx.run("env", home=home)
            self.assertEqual((out.returncode, out.stdout), (1, ""), home)

    def test_home_takes_userprofile_on_windows_only_when_home_is_unset(self):
        # C-82 slot audit N1: the rule of `moai lane-env` (`key_vault::home_from`).
        import importlib.util
        spec = importlib.util.spec_from_file_location("build_folder", TOOL)
        bf = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(bf)
        with tempfile.TemporaryDirectory() as tmp:
            h, u = str(Path(tmp) / "h"), str(Path(tmp) / "u")
            self.assertEqual(bf.home({"USERPROFILE": u}, windows=True), Path(u))
            self.assertEqual(bf.home({"HOME": "", "USERPROFILE": u}, windows=True), Path(u))
            self.assertEqual(bf.home({"HOME": h, "USERPROFILE": u}, windows=True), Path(h))
            self.assertEqual(bf.home({"HOME": h, "USERPROFILE": u}, windows=False), Path(h))
            for environ, windows, says in [
                ({"USERPROFILE": u}, False, "HOME must be an absolute folder"),
                ({}, True, "neither HOME nor USERPROFILE"),
                ({"USERPROFILE": "rel"}, True, "HOME must be an absolute folder"),
            ]:
                with self.assertRaises(bf.Refused) as e:
                    bf.home(environ, windows=windows)
                self.assertIn(says, str(e.exception), environ)


class CleanTest(unittest.TestCase):
    def setUp(self):
        self.fx = Fixture()
        self.addCleanup(self.fx.close)
        self.assertEqual(self.fx.run("env").returncode, 0)
        self.folder = self.fx.folder("repo-agent-7")
        cargo_output(self.folder)

    def test_clean_removes_cargo_output_and_the_folder_when_nothing_else_is_there(self):
        out = self.fx.run("clean")
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertFalse(self.folder.exists())
        self.assertEqual(json.loads(out.stdout)["kept"], [])

    def test_clean_keeps_evidence_and_says_so(self):
        (self.folder / "notes.md").write_text("evidence\n")
        out = self.fx.run("clean")
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertEqual(sorted(p.name for p in self.folder.iterdir()),
                         [".moai-build.json", "notes.md"])
        self.assertEqual(json.loads(out.stdout)["kept"], ["notes.md"])

    def test_clean_refuses_a_folder_another_worktree_owns(self):
        record = self.folder / ".moai-build.json"
        data = json.loads(record.read_text())
        data["worktree"] = "/somewhere/else"
        record.write_text(json.dumps(data))
        out = self.fx.run("clean")
        self.assertEqual(out.returncode, 1)
        self.assertTrue((self.folder / "debug").is_dir())

    def test_clean_without_a_folder_does_nothing(self):
        self.assertEqual(self.fx.run("clean").returncode, 0)
        out = self.fx.run("clean")
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertFalse(self.folder.exists())


def moai_bin():
    """The built binary `remove` reruns the preview with (`MOAI_CLI_BIN` or the cargo target
    folder), as tests/self_audit_run.py finds it; a missing binary fails the case."""
    explicit = os.environ.get("MOAI_CLI_BIN")
    if explicit:
        return Path(explicit)
    target = os.environ.get("CARGO_TARGET_DIR")
    path = (Path(target) if target else ROOT / "target") / "debug" / "moai-cli"
    if not path.exists():
        raise AssertionError(f"moai-cli not built at {path}; run cargo build first")
    return path


class RemoveTest(unittest.TestCase):
    """`remove` deletes a picked folder only when it is a `trim` row of the handed preview AND
    of a fresh `optimize-plan` run on the same integration ref (audit N1: a hand-written
    preview is not enough)."""

    def setUp(self):
        self.fx = Fixture()
        self.addCleanup(self.fx.close)
        # repo-gone: its worktree is gone; repo-live: the fixture's fresh, live worktree.
        for name, worktree in [("repo-gone", self.fx.home / "gone-worktree"),
                               ("repo-live", self.fx.wt)]:
            self.fx.folder(name).mkdir(parents=True)
            cargo_output(self.fx.folder(name))
            (self.fx.folder(name) / ".moai-build.json").write_text(json.dumps({
                "schema_version": 1, "git_common_dir": str(self.fx.repo / ".git"),
                "worktree": str(worktree)}))
        self.preview = self.write_preview({"repo-gone": "trim", "repo-live": "skip"})

    def write_preview(self, actions, name="preview.json"):
        rows = [{"kind": "agent_build_folder", "path": str(self.fx.folder(n)), "action": a}
                for n, a in actions.items()]
        path = self.fx.home / name
        path.write_text(json.dumps(
            {"ok": True, "result": {"integration_ref": "main", "rows": rows}}))
        return path

    def remove(self, preview, *names):
        return self.fx.run("remove", "--preview", str(preview), "--moai", str(moai_bin()),
                           *names)

    def test_remove_deletes_only_a_picked_trim_row(self):
        out = self.remove(self.preview, "repo-gone")
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertFalse(self.fx.folder("repo-gone").exists())
        self.assertTrue(self.fx.folder("repo-live").is_dir())

    def test_remove_refuses_every_name_when_one_is_not_a_trim_row(self):
        for names in [["repo-live"], ["repo-gone", "repo-live"], ["../repo-gone"],
                      ["repo-missing"]]:
            out = self.remove(self.preview, *names)
            self.assertEqual(out.returncode, 1, names)
            self.assertTrue((self.fx.folder("repo-gone") / "debug").is_dir(), names)
            self.assertTrue((self.fx.folder("repo-live") / "debug").is_dir(), names)

    def test_remove_refuses_a_hand_written_preview_that_marks_a_live_folder_trim(self):
        forged = self.write_preview({"repo-gone": "trim", "repo-live": "trim"}, "forged.json")
        for names in [["repo-live"], ["repo-gone", "repo-live"]]:
            out = self.remove(forged, *names)
            self.assertEqual(out.returncode, 1, (names, out.stderr))
            self.assertIn("fresh preview", out.stderr)
            self.assertTrue((self.fx.folder("repo-live") / "debug").is_dir(), names)
            self.assertTrue((self.fx.folder("repo-gone") / "debug").is_dir(), names)

    def test_remove_refuses_a_folder_that_became_a_link(self):
        target = self.fx.home / "t"
        target.mkdir()
        (target / "keep.txt").write_text("k\n")
        os.rename(self.fx.folder("repo-gone"), self.fx.home / "moved")
        self.fx.folder("repo-gone").symlink_to(target)
        out = self.remove(self.preview, "repo-gone")
        self.assertEqual(out.returncode, 1)
        self.assertTrue((target / "keep.txt").is_file())

    def test_remove_refuses_a_linked_moai_folder(self):
        # Audit B1: `~/.moai` itself is a link to a folder holding a would-be leftover.
        elsewhere = self.fx.home / "elsewhere"
        os.rename(self.fx.home / ".moai", elsewhere)
        (self.fx.home / ".moai").symlink_to(elsewhere)
        out = self.remove(self.preview, "repo-gone")
        self.assertEqual(out.returncode, 1)
        self.assertIn("is a link", out.stderr)
        self.assertTrue((elsewhere / "build" / "repo-gone" / "debug").is_dir())
        self.assertTrue((elsewhere / "build" / "repo-gone" / "CACHEDIR.TAG").is_file())


def tree(folder):
    """Every entry under `folder` with its bytes (links by target), for before/after checks."""
    out = {}
    base = Path(folder)

    def visit(directory):
        for entry in sorted(os.scandir(directory), key=lambda e: e.name):
            path = Path(entry.path)
            rel = str(path.relative_to(base))
            if entry.is_symlink():
                out[rel] = ("link", os.readlink(path))
            elif entry.is_dir(follow_symlinks=False):
                out[rel] = ("dir",)
                visit(path)
            else:
                out[rel] = ("file", path.read_bytes())

    visit(base)
    return out


class SeedTest(unittest.TestCase):
    """`seed --from <name>`: the worktree's own folder starts as a copy-on-write clone of a lane's
    built folder. Where the platform or filesystem cannot clone, the answer is `seeded: false`
    with a reason and nothing is copied."""

    def setUp(self):
        self.fx = Fixture()
        self.addCleanup(self.fx.close)
        self.source = self.fx.folder("lane-src")
        self.source.mkdir(parents=True)
        cargo_output(self.source)
        (self.source / "notes.md").write_text("lane evidence\n")
        (self.source / ".moai-build.json").write_text(json.dumps(
            {"schema_version": 1, "git_common_dir": "/lane/.git", "worktree": "/lane"}))
        self.own = self.fx.folder("repo-agent-7")

    def seed(self, name="lane-src", **kw):
        return self.fx.run("seed", "--from", name, **kw)

    def answer(self, out):
        self.assertEqual(out.returncode, 0, out.stderr)
        return json.loads(out.stdout)

    def seeded(self):
        """The answer of a seed that is expected to clone; skips where the filesystem cannot."""
        got = self.answer(self.seed())
        if not got["seeded"] and got["reason"] in ("unsupported", "clone_failed"):
            self.skipTest(f"no copy-on-write clone here: {got}")
        return got

    def test_seed_clones_the_cargo_output_and_nothing_else(self):
        got = self.seeded()
        self.assertTrue(got["seeded"], got)
        self.assertIn(got["method"], ("clonefile", "reflink"))
        self.assertEqual(Path(got["folder"]), self.own)
        self.assertEqual((self.own / "debug" / "deps" / "lib.rlib").read_bytes(), b"x" * 4096)
        self.assertTrue((self.own / "CACHEDIR.TAG").is_file())
        # Evidence stays with the lane, and the owner record is this worktree's, not the lane's.
        self.assertFalse((self.own / "notes.md").exists())
        record = json.loads((self.own / ".moai-build.json").read_text())
        self.assertEqual(record["worktree"], os.path.realpath(self.fx.wt))
        self.assertEqual(sorted(p.name for p in self.own.iterdir()),
                         [".moai-build.json", "CACHEDIR.TAG", "debug", "tmp"])

    def test_seed_leaves_the_source_unchanged(self):
        before = tree(self.source)
        self.seeded()
        self.assertEqual(tree(self.source), before)

    def test_seed_writes_nothing_of_its_own_into_the_clone(self):
        self.seeded()
        # A second seed finds the folder built and changes nothing.
        before = tree(self.own)
        got = self.answer(self.seed())
        self.assertEqual((got["seeded"], got["reason"]), (False, "already_built"))
        self.assertEqual(tree(self.own), before)

    def test_seed_does_nothing_when_the_folder_is_already_built(self):
        self.assertEqual(self.fx.run("env").returncode, 0)
        cargo_output(self.own)
        (self.own / "debug" / "deps" / "mine.rlib").write_bytes(b"mine")
        before = tree(self.own)
        got = self.answer(self.seed())
        self.assertEqual((got["seeded"], got["reason"]), (False, "already_built"))
        self.assertEqual(tree(self.own), before)

    def test_seed_copies_nothing_when_cloning_is_switched_off(self):
        got = self.answer(self.seed(extra_env={"BUILD_FOLDER_NO_COW": "1"}))
        self.assertEqual((got["seeded"], got["reason"]), (False, "disabled"))
        self.assertFalse((self.own / "debug").exists())
        self.assertEqual(sorted(p.name for p in self.own.iterdir()), [".moai-build.json"])

    def test_seed_does_not_clone_a_folder_that_is_being_built(self):
        with open(self.source / "debug" / ".cargo-lock", "w") as lock:
            fcntl.flock(lock, fcntl.LOCK_EX)
            got = self.answer(self.seed())
        self.assertEqual((got["seeded"], got["reason"]), (False, "source_building"))
        self.assertFalse((self.own / "debug").exists())

    def test_seed_without_a_source_folder_says_so(self):
        got = self.answer(self.seed("no-such-folder"))
        self.assertEqual((got["seeded"], got["reason"]), (False, "source_missing"))
        self.assertFalse((self.own / "debug").exists())

    def test_seed_refuses_a_linked_source_and_names_that_are_not_one_folder(self):
        elsewhere = self.fx.home / "elsewhere"
        elsewhere.mkdir()
        cargo_output(elsewhere)
        os.symlink(elsewhere, self.fx.folder("linked"))
        for name in ("linked", "../x", "a/b", "", ".."):
            out = self.seed(name)
            self.assertEqual(out.returncode, 1, (name, out.stdout, out.stderr))
            self.assertEqual(out.stdout, "")
        self.assertFalse(self.own.exists())

    def test_seed_refuses_the_worktrees_own_folder(self):
        self.assertEqual(self.fx.run("env").returncode, 0)
        out = self.seed("repo-agent-7")
        self.assertEqual(out.returncode, 1, out.stderr)

    def test_seed_refuses_a_linked_build_root(self):
        real = self.fx.home / "elsewhere-build"
        real.mkdir()
        shutil_rmtree(self.fx.home / ".moai" / "build")
        os.symlink(real, self.fx.home / ".moai" / "build")
        out = self.seed()
        self.assertEqual(out.returncode, 1, out.stderr)

    def test_seed_needs_a_source_name(self):
        self.assertEqual(self.fx.run("seed").returncode, 2)
        self.assertEqual(self.fx.run("seed", "--from").returncode, 2)

    def in_the_way(self, prepare):
        """Own folder with something already at a path the clone would take: nothing is cloned and
        nothing that was there is touched."""
        self.assertEqual(self.fx.run("env").returncode, 0)
        prepare(self.own)
        before = tree(self.own)
        got = self.answer(self.seed())
        self.assertEqual((got["seeded"], got["reason"]), (False, "target_in_the_way"), got)
        self.assertEqual(tree(self.own), before)

    def test_seed_leaves_a_file_that_is_in_the_way_alone(self):
        # An evidence file named like a cargo folder is not cargo output, so it does not make the
        # folder "already built"; it must survive a seed that cannot use its name.
        self.in_the_way(lambda own: (own / "tmp").write_text("my evidence\n"))

    def test_seed_leaves_a_folder_that_is_in_the_way_alone(self):
        def prepare(own):
            (own / "debug").mkdir()
            (own / "debug" / "evidence.txt").write_text("my evidence\n")
        self.in_the_way(prepare)

    def test_seed_leaves_a_link_that_is_in_the_way_alone(self):
        def prepare(own):
            target = own.parent / "elsewhere.txt"
            target.write_text("not mine to delete\n")
            os.symlink(target, own / "debug")
        self.in_the_way(prepare)
        self.assertEqual((self.own.parent / "elsewhere.txt").read_text(), "not mine to delete\n")

    def test_seed_says_nothing_was_cloned_when_the_source_has_no_cargo_output(self):
        empty = self.fx.folder("lane-empty")
        empty.mkdir()
        (empty / "notes.md").write_text("only evidence\n")
        got = self.answer(self.seed("lane-empty"))
        self.assertEqual((got["seeded"], got["reason"]), (False, "source_empty"), got)
        self.assertEqual(sorted(p.name for p in self.own.iterdir()), [".moai-build.json"])

    def test_seed_refuses_the_own_folder_written_in_another_letter_case(self):
        self.assertEqual(self.fx.run("env").returncode, 0)
        if not self.fx.folder("REPO-AGENT-7").exists():
            self.skipTest("this filesystem tells letter cases apart")
        out = self.seed("REPO-AGENT-7")
        self.assertEqual(out.returncode, 1, (out.stdout, out.stderr))
        self.assertEqual(out.stdout, "")

    def test_seed_names_the_cargo_home_condition_of_a_warm_start(self):
        got = self.seeded()
        self.assertIn("CARGO_HOME", got["note"], got)

    def test_seed_without_a_build_root_is_not_an_error(self):
        shutil_rmtree(self.fx.home / ".moai")
        got = self.answer(self.seed())
        self.assertEqual((got["seeded"], got["reason"]), (False, "source_missing"), got)

    def test_clean_removes_a_clone_of_a_read_only_source(self):
        deps = self.source / "debug" / "deps"
        deps.chmod(0o555)
        (deps / "lib.rlib").chmod(0o444)
        self.addCleanup(lambda: deps.chmod(0o755))
        self.seeded()
        out = self.fx.run("clean")
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertFalse(self.own.exists())


def load_tool():
    spec = importlib.util.spec_from_file_location("build_folder_under_test", TOOL)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class SeedRollbackTest(unittest.TestCase):
    """A clone that fails part-way is taken back, and only what this run made. The failure is put
    in by replacing `clone_entry` in a loaded copy of the tool."""

    def setUp(self):
        self.fx = Fixture()
        self.addCleanup(self.fx.close)
        self.source = self.fx.folder("lane-src")
        self.source.mkdir(parents=True)
        cargo_output(self.source)
        (self.source / "notes.md").write_text("lane evidence\n")
        self.own = self.fx.folder("repo-agent-7")
        self.tool = load_tool()
        self.real_clone = self.tool.clone_entry

    def seed_with(self, clone):
        env = {"HOME": str(self.fx.home), "GIT_CONFIG_GLOBAL": "/dev/null",
               "GIT_CONFIG_SYSTEM": "/dev/null"}
        out = io.StringIO()
        old = os.getcwd()
        os.chdir(self.fx.wt)
        try:
            with mock.patch.dict(os.environ, env), mock.patch.object(self.tool, "clone_entry", clone), \
                    contextlib.redirect_stdout(out):
                rc = self.tool.cmd_seed("lane-src")
        finally:
            os.chdir(old)
        return rc, json.loads(out.getvalue())

    def can_clone(self):
        probe = self.fx.home / "probe-src"
        probe.mkdir()
        try:
            self.real_clone(probe, self.fx.home / "probe-dst")
        except OSError:
            self.skipTest("no copy-on-write clone here")

    def test_a_failure_after_the_first_entry_takes_back_that_entry_only(self):
        self.can_clone()
        before = tree(self.source)
        calls = []

        def clone(src, dst):
            calls.append(Path(src).name)
            if len(calls) == 2:
                raise OSError(errno.ENOSPC, "injected: no space left")
            return self.real_clone(src, dst)

        rc, got = self.seed_with(clone)
        self.assertEqual(rc, 0)
        self.assertEqual((got["seeded"], got["reason"]), (False, "clone_failed"), got)
        self.assertGreaterEqual(len(calls), 2)
        self.assertEqual(sorted(p.name for p in self.own.iterdir()), [".moai-build.json"])
        self.assertEqual(tree(self.source), before)

    def test_a_half_written_entry_of_the_failed_clone_is_taken_back(self):
        # `cp --reflink=always` can leave a part of the destination behind when it fails.
        before = tree(self.source)

        def clone(src, dst):
            Path(dst).mkdir()
            (Path(dst) / "half.o").write_text("half written\n")
            raise OSError(errno.EIO, "injected: failed after writing a part")

        rc, got = self.seed_with(clone)
        self.assertEqual((rc, got["seeded"], got["reason"]), (0, False, "clone_failed"), got)
        self.assertEqual(sorted(p.name for p in self.own.iterdir()), [".moai-build.json"])
        self.assertEqual(tree(self.source), before)

    def test_a_read_only_entry_of_the_failed_clone_is_taken_back(self):
        # a clone keeps the read-only modes of its source: the same hook as `clean` removes it
        def clone(src, dst):
            Path(dst).mkdir()
            (Path(dst) / "ro.o").write_text("read only\n")
            os.chmod(Path(dst) / "ro.o", 0o444)
            os.chmod(dst, 0o555)
            raise OSError(errno.EIO, "injected: failed with a read-only part")

        self.addCleanup(lambda: [os.chmod(p, 0o755) for p in self.own.iterdir() if p.is_dir()])
        rc, got = self.seed_with(clone)
        self.assertEqual((rc, got["seeded"], got["reason"]), (0, False, "clone_failed"), got)
        self.assertNotIn("could not take back", got["detail"])
        self.assertEqual(sorted(p.name for p in self.own.iterdir()), [".moai-build.json"])

    def test_an_entry_that_cannot_be_taken_back_is_named_in_the_answer(self):
        made = []

        def clone(src, dst):
            made.append(Path(dst).name)
            Path(dst).mkdir()
            raise OSError(errno.EIO, "injected: failed")

        def stuck(*_args, **_kwargs):
            raise PermissionError(errno.EACCES, "injected: cannot remove")

        with mock.patch.object(self.tool.shutil, "rmtree", stuck):
            rc, got = self.seed_with(clone)
        self.assertEqual((rc, got["seeded"], got["reason"]), (0, False, "clone_failed"), got)
        self.assertIn("injected: failed", got["detail"])
        self.assertEqual(len(made), 1)
        self.assertIn(f"could not take back: {made[0]}", got["detail"])

    def test_the_linux_copy_messages_map_to_the_reasons(self):
        def failing(stderr):
            run = mock.Mock(return_value=mock.Mock(returncode=1, stderr=stderr))
            with mock.patch.object(self.tool.sys, "platform", "linux"), \
                    mock.patch.object(self.tool.subprocess, "run", run):
                with self.assertRaises(OSError) as ctx:
                    self.tool.clone_entry(self.source / "debug", self.own / "debug")
            return ctx.exception.errno

        self.assertEqual(failing("cp: failed to clone: Operation not supported"), errno.ENOTSUP)
        self.assertEqual(failing("cp: failed to clone: Invalid cross-device link"), errno.ENOTSUP)
        self.assertEqual(failing("cp: failed to clone: No space left on device"), errno.EIO)
        self.assertEqual(failing(""), errno.EIO)

    def test_the_source_stays_locked_while_it_is_cloned_and_is_free_afterwards(self):
        self.can_clone()
        lock = self.source / "debug" / ".cargo-lock"
        during = []

        def try_exclusive():
            with open(lock, "w") as f:
                try:
                    fcntl.flock(f, fcntl.LOCK_EX | fcntl.LOCK_NB)
                    return "free"
                except OSError:
                    return "held"

        def clone(src, dst):
            during.append(try_exclusive())
            return self.real_clone(src, dst)

        rc, got = self.seed_with(clone)
        self.assertEqual((rc, got["seeded"]), (0, True), got)
        # A lane that starts to build while the clone is made has to wait for it.
        self.assertEqual(set(during), {"held"}, during)
        self.assertEqual(try_exclusive(), "free")

    def test_a_platform_that_cannot_clone_is_told_apart_from_a_failed_clone(self):
        for number in (errno.ENOTSUP, errno.EXDEV):
            def clone(src, dst, number=number):
                raise OSError(number, "injected")
            rc, got = self.seed_with(clone)
            self.assertEqual((rc, got["seeded"], got["reason"]), (0, False, "unsupported"), number)
            self.assertEqual(sorted(p.name for p in self.own.iterdir()), [".moai-build.json"])


class AuditPlaceTest(unittest.TestCase):
    """`audit-place --name <n> --head <sha>`: one fixed audit worktree and its build folder, moved
    to each audited head instead of cloned again, so files that did not change keep their
    modification times and cargo rebuilds only the crates that did."""

    def setUp(self):
        self.fx = Fixture()
        self.addCleanup(self.fx.close)
        repo = self.fx.repo
        (repo / ".gitignore").write_text("target/\n")
        (repo / "b.txt").write_text("b1\n")
        git(repo, "add", "-A")
        git(repo, "commit", "-q", "-m", "one")
        self.c1 = git(repo, "rev-parse", "HEAD").strip()
        (repo / "b.txt").write_text("b2\n")
        git(repo, "commit", "-q", "-am", "two")
        self.c2 = git(repo, "rev-parse", "HEAD").strip()
        (repo / "b.txt").write_text("b3\n")
        (repo / "c.txt").write_text("c3\n")
        git(repo, "add", "-A")
        git(repo, "commit", "-q", "-m", "three")
        self.c3 = git(repo, "rev-parse", "HEAD").strip()
        self.place = self.fx.home / ".moai" / "worktrees" / "repo" / "audit-a1"
        self.folder = self.fx.folder("repo-audit-a1")

    def call(self, head, name="a1", **kw):
        return self.fx.run("audit-place", "--name", name, "--head", head, **kw)

    def answer(self, out):
        self.assertEqual(out.returncode, 0, out.stderr)
        return json.loads(out.stdout)

    def head_in_place(self):
        return git(self.place, "rev-parse", "HEAD").strip()

    def test_a_first_call_makes_the_place_and_its_build_folder(self):
        got = self.answer(self.call(self.c1))
        self.assertTrue(got["created"], got)
        self.assertIsNone(got["previous"])
        self.assertEqual(Path(got["place"]), self.place)
        self.assertEqual(Path(got["folder"]), self.folder)
        self.assertEqual(got["head"], self.c1)
        self.assertEqual(got["tree"], git(self.fx.repo, "rev-parse", f"{self.c1}^{{tree}}").strip())
        self.assertEqual(self.head_in_place(), self.c1)
        self.assertEqual(git(self.place, "branch", "--show-current").strip(), "")  # detached
        record = json.loads((self.folder / ".moai-build.json").read_text())
        self.assertEqual(record["worktree"], os.path.realpath(self.place))
        self.assertEqual((self.place / "b.txt").read_text(), "b1\n")

    def test_a_second_call_moves_the_same_folder_and_keeps_unchanged_files_as_they_are(self):
        self.answer(self.call(self.c2))
        old = 1_000_000_000
        os.utime(self.place / "a.txt", (old, old))
        inode = os.stat(self.place / "a.txt").st_ino
        got = self.answer(self.call(self.c3))
        self.assertFalse(got["created"], got)
        self.assertEqual(got["previous"], self.c2)
        self.assertEqual(got["head"], self.c3)
        self.assertEqual(got["moved_files"], 2)  # b.txt changed, c.txt added
        self.assertEqual(self.head_in_place(), self.c3)
        self.assertEqual((self.place / "b.txt").read_text(), "b3\n")
        # The file that did not change was not rewritten: same modification time, same inode.
        self.assertEqual(int(os.stat(self.place / "a.txt").st_mtime), old)
        self.assertEqual(os.stat(self.place / "a.txt").st_ino, inode)

    def test_the_same_head_twice_changes_nothing(self):
        self.answer(self.call(self.c2))
        old = 1_000_000_000
        os.utime(self.place / "b.txt", (old, old))
        got = self.answer(self.call(self.c2))
        self.assertFalse(got["created"])
        self.assertEqual(got["moved_files"], 0)
        self.assertEqual(got["taken_back"], 0)
        self.assertEqual(int(os.stat(self.place / "b.txt").st_mtime), old)

    def test_a_move_takes_back_edits_and_strays_but_not_ignored_files(self):
        self.answer(self.call(self.c1))
        (self.place / "a.txt").write_text("edited\n")  # a tracked file changed
        (self.place / "stray_test.rs").write_text("fn main() {}\n")  # would be a test crate
        (self.place / "target").mkdir()
        (self.place / "target" / "evidence.txt").write_text("keep\n")  # ignored
        got = self.answer(self.call(self.c2))
        self.assertEqual(got["taken_back"], 2, got)
        self.assertEqual((self.place / "a.txt").read_text(), "a\n")
        self.assertFalse((self.place / "stray_test.rs").exists())
        self.assertEqual((self.place / "target" / "evidence.txt").read_text(), "keep\n")
        self.assertEqual(git(self.place, "status", "--porcelain").strip(), "")
        self.assertEqual(got["tree"], git(self.place, "rev-parse", "HEAD^{tree}").strip())

    def test_a_bad_head_or_name_is_refused_and_writes_nothing(self):
        for head in ["abc", self.c1.upper(), "0" * 40, self.c1 + "0", self.c1[:12], "--help", ""]:
            out = self.call(head)
            self.assertEqual(out.returncode, 1, (head, out.stdout, out.stderr))
            self.assertEqual(out.stdout, "")
        for name in ["a/b", "..", "a b", "", "-x"]:
            out = self.call(self.c1, name=name)
            self.assertEqual(out.returncode, 1, (name, out.stdout, out.stderr))
        self.assertFalse((self.fx.home / ".moai" / "worktrees").exists())
        self.assertFalse((self.fx.home / ".moai" / "build").exists())

    def test_a_place_that_is_a_link_a_file_or_not_a_worktree_is_refused(self):
        self.place.parent.mkdir(parents=True)
        elsewhere = self.fx.home / "elsewhere"
        elsewhere.mkdir()
        os.symlink(elsewhere, self.place)
        out = self.call(self.c1)
        self.assertEqual(out.returncode, 1, out.stderr)
        self.assertTrue(os.path.islink(self.place))
        self.assertEqual(list(elsewhere.iterdir()), [])
        os.unlink(self.place)
        self.place.write_text("a file\n")
        out = self.call(self.c1)
        self.assertEqual(out.returncode, 1, out.stderr)
        self.assertEqual(self.place.read_text(), "a file\n")
        self.place.unlink()
        self.place.mkdir()
        (self.place / "mine.txt").write_text("not a worktree\n")
        out = self.call(self.c1)
        self.assertEqual(out.returncode, 1, out.stderr)
        self.assertEqual((self.place / "mine.txt").read_text(), "not a worktree\n")

    def test_a_clone_or_a_link_to_another_worktree_is_never_moved(self):
        # A full clone has every commit, so only the worktree registry tells it is not a place.
        self.place.parent.mkdir(parents=True)
        git(self.fx.home, "clone", "-q", str(self.fx.repo), str(self.place))
        git(self.place, "checkout", "-q", "--detach", self.c1)
        out = self.call(self.c3)
        self.assertEqual(out.returncode, 1, out.stderr)
        self.assertEqual(self.head_in_place(), self.c1)
        # A link to a real worktree of this repository must not move that worktree either.
        import shutil
        shutil.rmtree(self.place)
        os.symlink(self.fx.wt, self.place)
        before = git(self.fx.wt, "rev-parse", "HEAD").strip()
        out = self.call(self.c3)
        self.assertEqual(out.returncode, 1, out.stderr)
        self.assertTrue(os.path.islink(self.place))
        self.assertEqual(git(self.fx.wt, "rev-parse", "HEAD").strip(), before)

    def test_a_head_that_changes_a_gitignore_is_refused_while_ignored_files_are_there(self):
        # Audit N1: `clean` after the move reads the new head's rules, which would delete files the
        # old rules ignored. The move is refused and nothing changes.
        repo = self.fx.repo
        git(repo, "rm", "-q", ".gitignore")
        git(repo, "commit", "-q", "-m", "drop the ignore file")
        no_ignore = git(repo, "rev-parse", "HEAD").strip()
        self.answer(self.call(self.c3))
        (self.place / "target").mkdir()
        (self.place / "target" / "evidence.log").write_text("keep\n")
        out = self.call(no_ignore)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn("ignore_rules_changed", out.stderr)
        self.assertEqual(self.head_in_place(), self.c3)
        self.assertEqual((self.place / "target" / "evidence.log").read_text(), "keep\n")
        # Without ignored files the same move is allowed.
        (self.place / "target" / "evidence.log").unlink()
        (self.place / "target").rmdir()
        got = self.answer(self.call(no_ignore))
        self.assertEqual(got["head"], no_ignore)

    def test_a_changed_gitignore_in_a_subfolder_counts_too(self):
        repo = self.fx.repo
        (repo / "sub").mkdir()
        (repo / "sub" / ".gitignore").write_text("*.log\n")
        (repo / "sub" / "x.txt").write_text("x\n")
        git(repo, "add", "-A")
        git(repo, "commit", "-q", "-m", "sub ignore")
        with_rule = git(repo, "rev-parse", "HEAD").strip()
        (repo / "sub" / ".gitignore").write_text("")
        git(repo, "commit", "-q", "-am", "sub ignore dropped")
        without_rule = git(repo, "rev-parse", "HEAD").strip()
        self.answer(self.call(with_rule))
        (self.place / "sub" / "evidence.log").write_text("keep\n")  # ignored by sub/.gitignore
        out = self.call(without_rule)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn("ignore_rules_changed", out.stderr)
        self.assertEqual((self.place / "sub" / "evidence.log").read_text(), "keep\n")
        self.assertEqual(self.head_in_place(), with_rule)

    def test_a_place_whose_git_file_points_at_another_repository_is_refused(self):
        # Audit N2: the path is registered here, but its `.git` leads to a clone's worktree.
        self.answer(self.call(self.c1))
        clone = self.fx.home / "clone"
        git(self.fx.home, "clone", "-q", str(self.fx.repo), str(clone))
        other = self.fx.home / "other-wt"
        git(clone, "worktree", "add", "-q", "--detach", str(other), self.c1)
        (self.place / ".git").write_text((other / ".git").read_text())
        out = self.call(self.c3)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn("another repository", out.stderr)
        self.assertEqual(git(other, "rev-parse", "HEAD").strip(), self.c1)
        self.assertEqual(git(other, "status", "--porcelain").strip(), "")

    def test_a_refusal_after_the_worktree_was_made_takes_it_back(self):
        # Audit N3: a build folder of another worktree, and a file in the way of the owner record.
        other = self.folder
        other.mkdir(parents=True)
        (other / ".moai-build.json").write_text(json.dumps(
            {"schema_version": 1, "git_common_dir": "/x/.git", "worktree": "/some/other/worktree"}))
        out = self.call(self.c1)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertFalse(self.place.exists())
        self.assertEqual(len(git(self.fx.repo, "worktree", "list", "--porcelain").split("worktree ")), 3)
        (other / ".moai-build.json").unlink()
        (other / ".moai-build.json.tmp").write_text("in the way")
        out = self.call(self.c1)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertFalse(self.place.exists(), "a place made by a refused run is taken back")
        self.assertEqual(len(git(self.fx.repo, "worktree", "list", "--porcelain").split("worktree ")), 3)
        (other / ".moai-build.json.tmp").unlink()
        self.assertEqual(self.answer(self.call(self.c1))["head"], self.c1)  # not stuck

    def test_a_linked_build_root_refuses_before_a_place_is_made(self):
        elsewhere = self.fx.home / "elsewhere"
        elsewhere.mkdir()
        (self.fx.home / ".moai").mkdir()
        os.symlink(elsewhere, self.fx.home / ".moai" / "build")
        out = self.call(self.c1)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertFalse(self.place.exists())
        self.assertEqual(list(elsewhere.iterdir()), [])

    def test_an_untracked_nested_repository_is_taken_back_too(self):
        # Audit N4: `git clean -fd` keeps a nested repository and the place would be stuck.
        self.answer(self.call(self.c1))
        nested = self.place / "nested"
        nested.mkdir()
        git(nested, "init", "-q")
        (nested / "f.txt").write_text("x\n")
        got = self.answer(self.call(self.c2))
        self.assertGreaterEqual(got["taken_back"], 1)
        self.assertFalse(nested.exists())
        self.assertEqual(git(self.place, "status", "--porcelain").strip(), "")

    def test_only_a_detached_place_made_by_audit_place_is_moved(self):
        # Audit N5: a branch checked out in the place, or a worktree someone else made there.
        self.answer(self.call(self.c1))
        git(self.place, "checkout", "-q", "-b", "feat")
        (self.place / "a.txt").write_text("edited\n")
        out = self.call(self.c2)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn("branch checked out", out.stderr)
        self.assertEqual((self.place / "a.txt").read_text(), "edited\n")
        git(self.place, "checkout", "-q", "--detach", "-f", self.c1)
        marker = self.folder / ".moai-audit-place.json"
        marker.unlink()
        out = self.call(self.c2)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn("not made by audit-place", out.stderr)
        self.assertEqual(self.head_in_place(), self.c1)

    def test_a_place_folder_deleted_by_hand_is_made_again(self):
        # Audit N6: git still lists the deleted folder as a worktree.
        import shutil
        self.answer(self.call(self.c1))
        shutil.rmtree(self.place)
        got = self.answer(self.call(self.c2))
        self.assertTrue(got["created"], got)
        self.assertEqual(self.head_in_place(), self.c2)

    def test_the_owner_marker_names_the_place(self):
        self.answer(self.call(self.c1, extra_env={"CARGO_HOME": "/cargo/one"}))
        mark = json.loads((self.folder / ".moai-audit-place.json").read_text())
        self.assertEqual(mark["place"], os.path.realpath(self.place))
        self.assertEqual(mark["cargo_home"], "/cargo/one")
        self.assertEqual(mark["head"], self.c1)

    def test_a_held_cargo_lock_stops_the_move(self):
        self.answer(self.call(self.c1))
        cargo_output(self.folder)
        fd = os.open(self.folder / "debug" / ".cargo-lock", os.O_RDWR)
        self.addCleanup(os.close, fd)
        fcntl.flock(fd, fcntl.LOCK_EX)
        out = self.call(self.c2)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn("place_busy", out.stderr)
        self.assertEqual(self.head_in_place(), self.c1)
        self.assertEqual((self.place / "b.txt").read_text(), "b1\n")

    def test_cargo_home_is_recorded_and_a_change_is_reported_once(self):
        got = self.answer(self.call(self.c1, extra_env={"CARGO_HOME": "/cargo/one"}))
        self.assertEqual(got["cargo_home"], "/cargo/one")
        self.assertNotIn("cargo_home_changed", got)
        got = self.answer(self.call(self.c1, extra_env={"CARGO_HOME": "/cargo/two"}))
        self.assertTrue(got["cargo_home_changed"], got)
        self.assertEqual(got["cargo_home"], "/cargo/two")
        self.assertEqual(got["cargo_home_was"], "/cargo/one")
        got = self.answer(self.call(self.c1, extra_env={"CARGO_HOME": "/cargo/two"}))
        self.assertNotIn("cargo_home_changed", got)
        got = self.answer(self.call(self.c1))
        self.assertEqual(got["cargo_home"], str(self.fx.home / ".cargo"))
        self.assertTrue(got["cargo_home_changed"])

    def test_the_calling_checkout_and_the_cargo_output_are_left_alone(self):
        self.answer(self.call(self.c1))
        cargo_output(self.folder)
        before = tree(self.folder / "debug")
        wt_before = tree(self.fx.wt)
        self.answer(self.call(self.c3))
        self.assertEqual(tree(self.folder / "debug"), before)
        self.assertEqual(tree(self.fx.wt), wt_before)
        self.assertTrue(self.place.is_dir())

    def test_a_clean_in_the_place_still_removes_only_cargo_output(self):
        self.answer(self.call(self.c1))
        cargo_output(self.folder)
        out = subprocess.run(
            ["python3", "-I", str(TOOL), "clean"], cwd=self.place, capture_output=True, text=True,
            env={"PATH": os.environ.get("PATH", "/usr/bin:/bin"), "HOME": str(self.fx.home),
                 "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_SYSTEM": "/dev/null"})
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertFalse((self.folder / "debug").exists())
        self.assertTrue(self.place.is_dir())
        # The place's own state record is not cargo output, so the folder and the record stay.
        self.assertTrue((self.folder / ".moai-audit-place.json").is_file())


def shutil_rmtree(path):
    import shutil
    shutil.rmtree(path)


class ScriptsTest(unittest.TestCase):
    def test_the_check_scripts_turn_incremental_output_off(self):
        for rel in ["tools/local_check.sh", "tools/self_check.sh"]:
            lines = (ROOT / rel).read_text().splitlines()
            self.assertIn("export CARGO_INCREMENTAL=0", lines, rel)


if __name__ == "__main__":
    unittest.main()
