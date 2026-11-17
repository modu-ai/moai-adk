"""tools/build_folder.py (card C-82, user decision 2026-11-17 "빌드 폴더를 한 곳으로 모으기").

An agent or audit build puts CARGO_TARGET_DIR outside the worktree, in one folder per worktree
under `<HOME>/.moai/build/<project>-<worktree name>/`, always with CARGO_INCREMENTAL=0, and
deletes that folder's cargo output when its task ends. `remove` deletes only the folders a
user picked from an `optimize-plan` preview. Every test uses a temporary HOME and a temporary
git repository; the real HOME is never read or written.
"""
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

    def run(self, *args, cwd=None, home=None):
        env = {
            "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
            "HOME": str(home if home is not None else self.home),
            "GIT_CONFIG_GLOBAL": "/dev/null",
            "GIT_CONFIG_SYSTEM": "/dev/null",
        }
        return subprocess.run(
            ["python3", "-I", str(TOOL), *args],
            cwd=cwd or self.wt, env=env, capture_output=True, text=True,
        )

    def folder(self, name):
        return self.home / ".moai" / "build" / name

    def close(self):
        for p in self.home.rglob("*"):
            if p.is_dir() and not p.is_symlink():
                p.chmod(0o755)
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


class RemoveTest(unittest.TestCase):
    def setUp(self):
        self.fx = Fixture()
        self.addCleanup(self.fx.close)
        for name in ["repo-gone", "repo-live"]:
            self.fx.folder(name).mkdir(parents=True)
            cargo_output(self.fx.folder(name))
            (self.fx.folder(name) / ".moai-build.json").write_text(json.dumps({
                "schema_version": 1, "git_common_dir": str(self.fx.repo / ".git"),
                "worktree": str(self.fx.home / name)}))
        rows = [
            {"kind": "agent_build_folder", "path": str(self.fx.folder("repo-gone")),
             "action": "trim"},
            {"kind": "agent_build_folder", "path": str(self.fx.folder("repo-live")),
             "action": "skip"},
        ]
        self.preview = self.fx.home / "preview.json"
        self.preview.write_text(json.dumps({"ok": True, "result": {"rows": rows}}))

    def test_remove_deletes_only_a_picked_trim_row(self):
        out = self.fx.run("remove", "--preview", str(self.preview), "repo-gone")
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertFalse(self.fx.folder("repo-gone").exists())
        self.assertTrue(self.fx.folder("repo-live").is_dir())

    def test_remove_refuses_every_name_when_one_is_not_a_trim_row(self):
        for names in [["repo-live"], ["repo-gone", "repo-live"], ["../repo-gone"],
                      ["repo-missing"]]:
            out = self.fx.run("remove", "--preview", str(self.preview), *names)
            self.assertEqual(out.returncode, 1, names)
            self.assertTrue((self.fx.folder("repo-gone") / "debug").is_dir(), names)
            self.assertTrue((self.fx.folder("repo-live") / "debug").is_dir(), names)

    def test_remove_refuses_a_folder_that_became_a_link(self):
        target = self.fx.home / "t"
        target.mkdir()
        (target / "keep.txt").write_text("k\n")
        os.rename(self.fx.folder("repo-gone"), self.fx.home / "moved")
        self.fx.folder("repo-gone").symlink_to(target)
        out = self.fx.run("remove", "--preview", str(self.preview), "repo-gone")
        self.assertEqual(out.returncode, 1)
        self.assertTrue((target / "keep.txt").is_file())


class ScriptsTest(unittest.TestCase):
    def test_the_check_scripts_turn_incremental_output_off(self):
        for rel in ["tools/local_check.sh", "tools/self_check.sh"]:
            lines = (ROOT / rel).read_text().splitlines()
            self.assertIn("export CARGO_INCREMENTAL=0", lines, rel)


if __name__ == "__main__":
    unittest.main()
