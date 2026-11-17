"""tools/source_manifest.py (Issue #2837) and its git merge driver tools/merge_driver_manifest.py
(card C-74): temporary git repositories only."""
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
TOOL = ROOT / "tools" / "source_manifest.py"
DRIVER = ROOT / "tools" / "merge_driver_manifest.py"
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

    def env(self):
        return {
            "PATH": os.environ.get("PATH", ""),
            "HOME": str(self.home),
            "GIT_CONFIG_NOSYSTEM": "1",
            "GIT_CONFIG_GLOBAL": os.devnull,
            "LC_ALL": "C",
        }

    def git(self, *args, check=True, cwd=None):
        cfg = ["-c", "user.name=t", "-c", "user.email=t@example.invalid",
               "-c", "commit.gpgsign=false", "-c", "merge.conflictStyle=merge"]
        return subprocess.run(["git", *cfg, *args], cwd=cwd or self.root, env=self.env(),
                              capture_output=True, text=True, check=check)

    def with_driver_files(self):
        """Commit this repository's `.gitattributes` and both tools into the temporary one, the
        layout `--install` registers the driver for (a missing source is left out, so a test
        before the driver exists fails on its own assertion)."""
        for rel in (".gitattributes", "tools/source_manifest.py", "tools/merge_driver_manifest.py"):
            if (ROOT / rel).is_file():
                (self.root / rel).parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(ROOT / rel, self.root / rel)
        self.commit("driver files")

    def driver(self, *args):
        """`tools/merge_driver_manifest.py <args>` from the temporary repository's own copy."""
        return subprocess.run([sys.executable, "-I", str(self.root / "tools" / DRIVER.name), *args],
                              cwd=self.root, env=self.env(), capture_output=True, text=True)

    def install_driver(self):
        out = self.driver("--install")
        if out.returncode != 0:
            raise AssertionError(f"--install: exit {out.returncode}\n{out.stdout}{out.stderr}")

    def unmerged(self):
        return sorted({line.split("\t", 1)[1] for line in
                       self.git("ls-files", "-u").stdout.splitlines()})

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

    def branches(self, ours, theirs, cwd=None):
        """Run `ours(self)` on main and `theirs(self)` on a side branch, commit each (with the
        manifest rewritten), then merge the side branch into main (from `cwd`, default the
        repository root). Returns the merge result."""
        self.git("checkout", "-q", "-b", "side")
        theirs(self)
        self.assert_tool("write", 0)
        self.commit("theirs")
        self.git("checkout", "-q", "main")
        ours(self)
        self.assert_tool("write", 0)
        self.commit("ours")
        return self.git("merge", "-q", "--no-edit", "side", check=False, cwd=cwd)

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


def add_file(name):
    """A new package file `references/<name>.md` with its manifest entry."""
    def run(repo):
        repo.put(f"references/{name}.md", f"{name}\n")
        data = repo.manifest()
        data["files"][f"references/{name}.md"] = {"sha256": "", "origin": origin(name)}
        repo.manifest(data)
    return run


def rename_file(old, new):
    """`git mv` a package file and move its manifest entry to the end under the new name."""
    def run(repo):
        repo.git("mv", str(PKG / old), str(PKG / new))
        data = repo.manifest()
        data["files"][new] = data["files"].pop(old)
        repo.manifest(data)
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

    def test_a_fifo_in_the_package_is_refused_without_reading_it(self):
        # #2849 audit B1: reading a FIFO would block forever; the timeout only turns a hang
        # into a failure (a normal run takes well under a second).
        r = self.repo
        os.mkfifo(r.root / PKG / "workflows/pipe")
        for command in ("check", "write"):
            out = subprocess.run([sys.executable, "-I", str(TOOL), command, "--root", str(r.root)],
                                 capture_output=True, text=True, timeout=60)
            self.assertEqual(out.returncode, 2, out.stdout + out.stderr)
            self.assertIn("workflows/pipe is not a regular file", out.stdout)

    def test_a_linked_path_to_the_manifest_is_refused_and_nothing_outside_is_written(self):
        # #2849 audit B2: a linked package root would send the write outside the repository.
        for linked in ("skills", "skills/moai", "skills/moai/references",
                       "skills/moai/references/source-manifest.json"):
            with self.subTest(linked=linked), tempfile.TemporaryDirectory() as tmp:
                r = Repo(tmp)
                outside = Path(tmp) / "outside"
                real = r.root / linked
                os.rename(real, outside)
                os.symlink(outside, real)
                r.put("workflows/a.md", LINES + "more\n")
                before = {p: p.read_bytes() for p in outside.rglob("*") if p.is_file()} \
                    if outside.is_dir() else {outside: outside.read_bytes()}
                for command in ("check", "write"):
                    report = r.assert_tool(command, 2)
                    self.assertIn(f"{linked} is a link", report["needs_person"])
                after = {p: p.read_bytes() for p in outside.rglob("*") if p.is_file()} \
                    if outside.is_dir() else {outside: outside.read_bytes()}
                self.assertEqual(after, before, "nothing outside the repository changes")

    def test_write_keeps_the_manifest_mode(self):
        # #2849 audit N1: mkstemp makes 0600; the rewritten manifest keeps its own mode.
        r = self.repo
        path = r.root / PKG / MANIFEST
        os.chmod(path, 0o640)
        r.put("workflows/a.md", LINES + "more\n")
        self.assertTrue(r.assert_tool("write", 0)["written"])
        self.assertEqual(os.stat(path).st_mode & 0o777, 0o640)

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


class MergeDriverTest(unittest.TestCase):
    """tools/merge_driver_manifest.py (card C-74). Git runs the driver before it writes any
    merged file to the working tree, so the driver never reads the package from disk: it
    merges the three manifests and, for a file both sides changed, merges that file's three
    versions itself (found by their sha256 in the history) to learn the merged hash."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Repo(self.tmp.name)
        self.repo.with_driver_files()

    def tearDown(self):
        self.tmp.cleanup()

    def assert_clean_and_correct(self, merge):
        r = self.repo
        self.assertEqual(merge.returncode, 0, merge.stdout + merge.stderr)
        self.assertEqual(r.unmerged(), [])
        self.assertNotIn("<<<<<<< ", r.read(MANIFEST))
        r.assert_tool("check", 0)
        self.assertEqual(r.read(MANIFEST), r.fresh())

    def test_this_repository_names_the_driver_for_the_manifest(self):
        out = subprocess.run(["git", "-C", str(ROOT), "check-attr", "merge", "--",
                              (PKG / MANIFEST).as_posix()], capture_output=True, text=True)
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertEqual(out.stdout.strip(),
                         f"{(PKG / MANIFEST).as_posix()}: merge: moai-manifest")

    def test_install_is_idempotent_and_check_reports_it(self):
        r = self.repo
        before = r.driver("--check")
        self.assertEqual(before.returncode, 1, before.stdout + before.stderr)
        self.assertFalse(json.loads(before.stdout)["registered"])
        r.install_driver()
        r.install_driver()
        values = r.git("config", "--local", "--get-all", "merge.moai-manifest.driver").stdout
        self.assertEqual(len(values.splitlines()), 1, values)
        self.assertIn("tools/merge_driver_manifest.py", values)
        after = r.driver("--check")
        self.assertEqual(after.returncode, 0, after.stdout + after.stderr)
        self.assertTrue(json.loads(after.stdout)["registered"])

    def test_check_reports_a_driver_line_that_differs(self):
        r = self.repo
        r.git("config", "--local", "merge.moai-manifest.driver", "true")
        out = r.driver("--check")
        self.assertEqual(out.returncode, 1, out.stdout + out.stderr)
        self.assertFalse(json.loads(out.stdout)["registered"])

    def test_same_file_edited_apart_merges_with_the_driver(self):
        # 61 of the 62 historical manifest-only conflicts: both sides edit one skill file at
        # lines that merge cleanly, so only its sha256 line conflicts.
        r = self.repo
        r.install_driver()
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           edit_line("workflows/a.md", 20, "theirs"))
        self.assert_clean_and_correct(merge)
        self.assertEqual(r.read("workflows/a.md").splitlines()[::19], ["ours", "theirs"])

    def test_without_the_driver_registered_git_keeps_its_normal_conflict(self):
        r = self.repo
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           edit_line("workflows/a.md", 20, "theirs"))
        self.assertNotEqual(merge.returncode, 0)
        self.assertEqual(r.unmerged(), [(PKG / MANIFEST).as_posix()])
        self.assertIn("<<<<<<< ", r.read(MANIFEST))

    def test_registered_but_missing_from_the_checkout_falls_back_to_a_normal_conflict(self):
        # An older branch without the tool: the registered command runs git's own text merge.
        r = self.repo
        r.install_driver()
        r.git("rm", "-q", "tools/merge_driver_manifest.py")
        r.commit("no driver here")
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           edit_line("workflows/a.md", 20, "theirs"))
        self.assertNotEqual(merge.returncode, 0)
        self.assertEqual(r.unmerged(), [(PKG / MANIFEST).as_posix()])
        self.assertIn("<<<<<<< ", r.read(MANIFEST))
        r.assert_tool("write", 0)
        r.assert_tool("check", 0)

    def test_different_files_on_each_side_merge_with_the_driver(self):
        r = self.repo
        r.install_driver()
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           edit_line("workflows/b.md", 20, "theirs"))
        self.assert_clean_and_correct(merge)

    def test_a_merge_started_in_a_subfolder_uses_the_driver(self):
        r = self.repo
        r.install_driver()
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           edit_line("workflows/a.md", 20, "theirs"),
                           cwd=r.root / PKG / "workflows")
        self.assert_clean_and_correct(merge)

    def test_a_real_conflict_in_the_skill_file_leaves_the_manifest_unresolved(self):
        # The driver does not mark the manifest resolved: both stay conflicted until the
        # skill file is resolved and `write` runs.
        r = self.repo
        r.install_driver()
        merge = r.branches(edit_line("workflows/a.md", 10, "ours"),
                           edit_line("workflows/a.md", 10, "theirs"))
        self.assertNotEqual(merge.returncode, 0)
        self.assertEqual(r.unmerged(), [(PKG / MANIFEST).as_posix(), (PKG / "workflows/a.md").as_posix()])
        self.assertIn("<<<<<<< ", r.read(MANIFEST))
        self.assertIn("<<<<<<< ", r.read("workflows/a.md"))
        r.assert_tool("write", 2)
        r.put("workflows/a.md", LINES.replace("line 10\n", "both\n"))
        r.assert_tool("write", 0)
        r.assert_tool("check", 0)

    def test_new_files_on_both_sides_merge_with_the_driver(self):
        r = self.repo
        r.install_driver()
        merge = r.branches(add_file("x"), add_file("y"))
        self.assert_clean_and_correct(merge)
        self.assertEqual(list(r.manifest()["files"])[-2:], ["references/x.md", "references/y.md"])

    def test_a_rename_on_one_side_merges_with_the_driver(self):
        r = self.repo
        r.install_driver()
        merge = r.branches(edit_line("workflows/a.md", 1, "ours"),
                           both(rename_file("workflows/b.md", "workflows/c.md"),
                                edit_line("workflows/a.md", 20, "theirs")))
        self.assert_clean_and_correct(merge)
        files = r.manifest()["files"]
        self.assertNotIn("workflows/b.md", files)
        self.assertIn("workflows/c.md", files)

    def test_a_rename_against_an_edit_is_left_to_write(self):
        # Git carries our edit of b.md into theirs' c.md; the driver cannot know that merged
        # text from the manifests, so it leaves the conflict and `write` reads the files.
        r = self.repo
        r.install_driver()
        merge = r.branches(edit_line("workflows/b.md", 1, "ours"),
                           rename_file("workflows/b.md", "workflows/c.md"))
        self.assertNotEqual(merge.returncode, 0)
        self.assertEqual(r.unmerged(), [(PKG / MANIFEST).as_posix()])
        self.assertEqual(r.read("workflows/c.md").splitlines()[0], "ours")
        r.assert_tool("write", 0)
        r.assert_tool("check", 0)

    def test_an_origin_changed_on_both_sides_is_left_to_a_person(self):
        r = self.repo
        r.install_driver()
        merge = r.branches(
            both(edit_line("workflows/a.md", 1, "ours"),
                 lambda repo: repo.set_reason("workflows/a.md", "first, ours")),
            both(edit_line("workflows/a.md", 20, "theirs"),
                 lambda repo: repo.set_reason("workflows/a.md", "first, theirs")),
        )
        self.assertNotEqual(merge.returncode, 0)
        self.assertEqual(r.unmerged(), [(PKG / MANIFEST).as_posix()])
        self.assertIn("<<<<<<< ", r.read(MANIFEST))

    def test_a_stale_side_manifest_still_fails_check_after_the_merge(self):
        # The driver trusts each side's sha256 values; when one side committed a skill edit
        # without `write`, the merge result is wrong and `check` (CI) fails on it.
        r = self.repo
        r.install_driver()

        def theirs(repo):
            edit_line("workflows/a.md", 20, "theirs")(repo)
            repo.assert_tool("write", 0)
            repo.commit("theirs, manifest written")
            edit_line("workflows/a.md", 19, "theirs, no write")(repo)

        r.git("checkout", "-q", "-b", "side")
        theirs(r)
        r.commit("theirs, manifest stale")
        r.git("checkout", "-q", "main")
        edit_line("workflows/a.md", 1, "ours")(r)
        r.assert_tool("write", 0)
        r.commit("ours")
        merge = r.git("merge", "-q", "--no-edit", "side", check=False)
        self.assertEqual(merge.returncode, 0, merge.stdout + merge.stderr)
        self.assertEqual(r.assert_tool("check", 1)["changed"], ["workflows/a.md"])


if __name__ == "__main__":
    unittest.main()
