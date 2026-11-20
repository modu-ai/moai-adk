"""tools/build_folder.py (card C-82, user decision 2026-11-17 "빌드 폴더를 한 곳으로 모으기").

An agent or audit build puts CARGO_TARGET_DIR outside the worktree, in one folder per worktree
under `<HOME>/.moai/build/<project>-<worktree name>/`, always with CARGO_INCREMENTAL=0, and
deletes that folder's cargo output when its task ends. `remove` deletes only the folders a
user picked from an `optimize-plan` preview. Every test uses a temporary HOME and a temporary
git repository; the real HOME is never read or written.
"""
import fcntl
import json
import os
import shlex
import subprocess
import tempfile
import unittest
from pathlib import Path

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
    for path in sorted(Path(folder).rglob("*")):
        rel = str(path.relative_to(folder))
        if path.is_symlink():
            out[rel] = ("link", os.readlink(path))
        elif path.is_dir():
            out[rel] = ("dir",)
        else:
            out[rel] = ("file", path.read_bytes())
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
