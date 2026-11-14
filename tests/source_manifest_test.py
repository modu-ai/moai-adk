"""tools/source_manifest.py (Issue #2837): temporary git repositories only."""
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
TOOL = ROOT / "tools" / "source_manifest.py"
PKG = Path("skills") / "moai"
MANIFEST = "references/source-manifest.json"  # package-relative, like every path below
sys.path.insert(0, str(ROOT / "tools"))
import source_manifest as sm  # noqa: E402

LINES = "".join(f"line {i}\n" for i in range(1, 21))


def origin(name, reason="first"):
    return {"kind": "design-spec", "template_id": name, "adoption": "new", "reason": reason}


class Repo:
    """A repository with a three-file package and a scrubbed git setup. Paths given to `put`,
    `read` and `edit_line` are relative to the package folder."""

    def __init__(self, tmp):
        self.root = Path(tmp) / "repo"
        self.home = Path(tmp) / "home"
        self.home.mkdir()
        self.root.mkdir()
        self.git("init", "-q", "-b", "main")
        self.put("SKILL.md", "# moai\n")
        self.put("workflows/a.md", LINES)
        self.put("workflows/b.md", LINES)
        self.manifest({
            "schema_version": 1,
            "covers": "every package file except this manifest",
            "files": {
                rel: {"sha256": "0" * 64, "origin": origin(rel)}
                for rel in ["SKILL.md", "workflows/a.md", "workflows/b.md"]
            },
        })
        self.assert_tool("write", 0)
        self.commit("base")

    def git(self, *args, check=True):
        env = {
            "PATH": os.environ.get("PATH", ""),
            "HOME": str(self.home),
            "GIT_CONFIG_NOSYSTEM": "1",
            "GIT_CONFIG_GLOBAL": os.devnull,
            "LC_ALL": "C",
        }
        cfg = ["-c", "user.name=t", "-c", "user.email=t@example.invalid",
               "-c", "commit.gpgsign=false", "-c", "merge.conflictStyle=merge"]
        return subprocess.run(["git", *cfg, *args], cwd=self.root, env=env,
                              capture_output=True, text=True, check=check)

    def put(self, rel, text):
        path = self.root / PKG / rel
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text, encoding="utf-8")

    def read(self, rel):
        return (self.root / PKG / rel).read_text(encoding="utf-8")

    def manifest(self, data=None):
        if data is not None:
            self.put(MANIFEST, json.dumps(data, indent=2, ensure_ascii=False) + "\n")
        return json.loads(self.read(MANIFEST))

    def set_reason(self, rel, reason):
        data = self.manifest()
        data["files"][rel]["origin"]["reason"] = reason
        self.manifest(data)

    def commit(self, message):
        self.git("add", "-A")
        self.git("commit", "-q", "-m", message)

    def tool(self, command):
        return subprocess.run([sys.executable, "-I", str(TOOL), command, "--root", str(self.root)],
                              capture_output=True, text=True,
                              env={**os.environ, "HOME": str(self.home),
                                   "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.devnull})

    def assert_tool(self, command, code):
        out = self.tool(command)
        if out.returncode != code:
            raise AssertionError(f"{command}: exit {out.returncode}\n{out.stdout}{out.stderr}")
        return json.loads(out.stdout)

    def branches(self, ours, theirs):
        """Run `ours(self)` on main and `theirs(self)` on a side branch, commit each (with the
        manifest rewritten), then merge the side branch into main. Returns the merge result."""
        self.git("checkout", "-q", "-b", "side")
        theirs(self)
        self.assert_tool("write", 0)
        self.commit("theirs")
        self.git("checkout", "-q", "main")
        ours(self)
        self.assert_tool("write", 0)
        self.commit("ours")
        return self.git("merge", "-q", "--no-edit", "side", check=False)

    def fresh(self):
        """The manifest built from scratch for the files now on disk: a copy whose sha256
        values are blanked, rewritten by the tool outside git."""
        with tempfile.TemporaryDirectory() as tmp:
            copy = Path(tmp)
            for path in (self.root / "skills").rglob("*"):
                if path.is_file():
                    dest = copy / path.relative_to(self.root)
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    dest.write_bytes(path.read_bytes())
            data = json.loads((copy / PKG / MANIFEST).read_text(encoding="utf-8"))
            for entry in data["files"].values():
                entry["sha256"] = "0" * 64
            (copy / PKG / MANIFEST).write_text(sm.render(data), encoding="utf-8")
            sm.write(copy)
            return (copy / PKG / MANIFEST).read_text(encoding="utf-8")


def edit_line(rel, n, text):
    def run(repo):
        lines = repo.read(rel).splitlines(keepends=True)
        lines[n - 1] = text + "\n"
        repo.put(rel, "".join(lines))
    return run


def both(*steps):
    def run(repo):
        for step in steps:
            step(repo)
    return run


class SourceManifestTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Repo(self.tmp.name)

    def tearDown(self):
        self.tmp.cleanup()

    def test_check_names_a_stale_entry_and_write_fixes_it(self):
        r = self.repo
        self.assertTrue(r.assert_tool("check", 0)["ok"])
        r.put("workflows/a.md", LINES + "more\n")
        report = r.assert_tool("check", 1)
        self.assertEqual(report["changed"], ["workflows/a.md"])
        before = r.read(MANIFEST)
        self.assertTrue(r.assert_tool("write", 0)["written"])
        self.assertNotEqual(r.read(MANIFEST), before)
        self.assertFalse(r.assert_tool("write", 0)["written"])
        r.assert_tool("check", 0)

    def test_ac1_different_skill_files_on_each_side(self):
        r = self.repo
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           edit_line("workflows/b.md", 20, "theirs"))
        if merge.returncode != 0:
            r.assert_tool("write", 0)
        r.assert_tool("check", 0)
        self.assertEqual(r.read(MANIFEST), r.fresh())

    def test_ac1_same_skill_file_conflicts_only_in_its_hash_and_one_write_resolves_it(self):
        r = self.repo
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           edit_line("workflows/a.md", 20, "theirs"))
        self.assertNotEqual(merge.returncode, 0, "the hash line conflicts")
        self.assertIn("<<<<<<< ", r.read(MANIFEST))
        self.assertNotIn("<<<<<<< ", r.read("workflows/a.md"))
        report = r.assert_tool("write", 0)
        self.assertTrue(report["conflicted"])
        self.assertEqual(r.read(MANIFEST), r.fresh())
        r.assert_tool("check", 0)

    def test_an_unmerged_manifest_is_read_from_the_index_stages(self):
        r = self.repo
        r.branches(edit_line("workflows/a.md", 1, "ours"),
                   edit_line("workflows/a.md", 20, "theirs"))
        stages = sm.index_stages(r.root)
        self.assertIsNotNone(stages)
        want = tuple(r.git("show", f":{n}:{(PKG / MANIFEST).as_posix()}").stdout for n in (1, 2, 3))
        self.assertEqual(stages, want)
        self.assertEqual(len({json.loads(t)["files"]["workflows/a.md"]["sha256"] for t in want}), 3)
        r.assert_tool("write", 0)
        r.git("add", str(PKG / MANIFEST))
        self.assertIsNone(sm.index_stages(r.root), "a resolved manifest has no stages")

    def test_an_origin_changed_on_one_side_is_kept_three_way(self):
        r = self.repo
        merge = r.branches(
            both(edit_line("workflows/a.md", 1, "ours"),
                 lambda repo: repo.set_reason("workflows/a.md", "first, then ours")),
            edit_line("workflows/a.md", 20, "theirs"),
        )
        self.assertNotEqual(merge.returncode, 0)
        r.assert_tool("write", 0)
        self.assertEqual(r.manifest()["files"]["workflows/a.md"]["origin"]["reason"],
                         "first, then ours")
        r.assert_tool("check", 0)

    def test_an_origin_changed_on_both_sides_is_left_to_a_person(self):
        r = self.repo
        r.branches(
            both(edit_line("workflows/a.md", 1, "ours"),
                 lambda repo: repo.set_reason("workflows/a.md", "first, ours")),
            both(edit_line("workflows/a.md", 20, "theirs"),
                 lambda repo: repo.set_reason("workflows/a.md", "first, theirs")),
        )
        before = r.read(MANIFEST)
        report = r.assert_tool("write", 2)
        self.assertIn("workflows/a.md", report["needs_person"])
        self.assertIn("reason", report["needs_person"])
        self.assertEqual(r.read(MANIFEST), before, "nothing is written")

    def test_a_skill_file_still_in_conflict_stops_the_write(self):
        r = self.repo
        r.branches(edit_line("workflows/a.md", 10, "ours"),
                   edit_line("workflows/a.md", 10, "theirs"))
        before = r.read(MANIFEST)
        report = r.assert_tool("write", 2)
        self.assertIn("conflict markers remain in: workflows/a.md", report["needs_person"])
        self.assertEqual(r.read(MANIFEST), before)

    def test_entries_added_on_both_sides_are_joined(self):
        r = self.repo

        def add(name):
            def run(repo):
                repo.put(f"references/{name}.md", f"{name}\n")
                data = repo.manifest()
                data["files"][f"references/{name}.md"] = {"sha256": "", "origin": origin(name)}
                repo.manifest(data)
            return run

        r.branches(add("x"), add("y"))
        r.assert_tool("write", 0)
        keys = list(r.manifest()["files"])
        self.assertEqual(keys[-2:], ["references/x.md", "references/y.md"], keys)
        r.assert_tool("check", 0)

    def test_a_new_file_without_an_entry_and_a_deleted_file(self):
        r = self.repo
        r.put("references/new.md", "new\n")
        report = r.assert_tool("write", 2)
        self.assertIn("no entry (write its origin first): references/new.md", report["needs_person"])
        (r.root / PKG / "references/new.md").unlink()
        (r.root / PKG / "workflows/b.md").unlink()
        report = r.assert_tool("write", 0)
        self.assertEqual(report["dropped"], ["workflows/b.md"])
        self.assertNotIn("workflows/b.md", r.manifest()["files"])

    def test_a_link_in_the_package_is_refused(self):
        r = self.repo
        os.symlink("a.md", r.root / PKG / "workflows/link.md")
        report = r.assert_tool("write", 2)
        self.assertIn("workflows/link.md is a link", report["needs_person"])

    def test_a_markdown_heading_underline_is_not_a_conflict(self):
        r = self.repo
        r.put("workflows/a.md", "Title\n=======\n")
        r.assert_tool("write", 0)
        r.assert_tool("check", 0)

    def test_diff3_markers_without_a_git_index_are_joined_three_way(self):
        base = {"schema_version": 1, "files": {"a.md": {"sha256": "1", "origin": origin("a")}}}
        ours = json.loads(json.dumps(base))
        ours["files"]["a.md"]["origin"]["reason"] = "ours"
        theirs = json.loads(json.dumps(base))
        theirs["files"]["a.md"]["sha256"] = "2"
        texts = [sm.render(x).splitlines(keepends=True) for x in (ours, base, theirs)]
        # One block covering the whole object, the way diff3 writes it.
        text = ("<<<<<<< ours\n" + "".join(texts[0]) + "||||||| base\n" + "".join(texts[1])
                + "=======\n" + "".join(texts[2]) + ">>>>>>> theirs\n")
        joined = sm.joined(text)
        self.assertEqual(joined["files"]["a.md"]["origin"]["reason"], "ours")
        # Without a base section only equal sides pass.
        no_base = ("<<<<<<< ours\n" + "".join(texts[0]) + "=======\n" + "".join(texts[2])
                   + ">>>>>>> theirs\n")
        with self.assertRaises(sm.NeedsPerson):
            sm.joined(no_base)

    def test_this_repository_manifest_is_what_the_tool_writes(self):
        out = subprocess.run([sys.executable, "-I", str(TOOL), "check", "--root", str(ROOT)],
                             capture_output=True, text=True)
        self.assertEqual(out.returncode, 0, out.stdout + out.stderr)


if __name__ == "__main__":
    unittest.main()
