"""Tests for tools/shipped_text_ratchet.py (SPEC-20261121-shipped-text-ratchet).

The tool counts maintainer-only tokens (card numbers, issue numbers, tracker links, the name of
the maintainers' private rules file) in the text MoAI ships, and fails when a count rises above
the checked-in baseline. Each test builds a temporary tree and runs the tool as a maintainer
would, through `python3 -I tools/shipped_text_ratchet.py`; RepoTreeTest runs it on this
repository. Written for Python 3.9.
"""

import hashlib
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TOOL = REPO_ROOT / "tools" / "shipped_text_ratchet.py"
IMPROVE = REPO_ROOT / "tools" / "improve"
REFUSAL_PREFIX = "shipped_text_ratchet: "

TOKENS = {
    "version": 1,
    "paths": ["shipped/**", "agents/*.md", "src/help.rs"],
    "tokens": [
        {"id": "card", "pattern": r"\bC-[0-9]+\b", "why": "card number"},
        {"id": "link", "pattern": "example/repo", "why": "tracker link"},
    ],
}

# The files of the temporary tree and what the tokens match in each:
#   shipped/a.md       card 2, link 1
#   shipped/sub/b.md   card 1
#   src/help.rs        card 1
#   agents/x.md        nothing
# Outside the shipped paths and so never counted: agents/deep/y.md (agents/*.md is one level),
# other/z.md. shipped/bin.dat is not UTF-8 and is skipped.
TREE = {
    "shipped/a.md": "C-1 and C-2 and example/repo\n",
    "shipped/sub/b.md": "see C-3\n",
    "src/help.rs": "// C-4\n",
    "agents/x.md": "no tokens here\n",
    "agents/deep/y.md": "C-9\n",
    "other/z.md": "C-5 C-6 example/repo\n",
}

BASELINE = {
    "version": 1,
    "tokens": {"card": 4, "link": 1},
    "files": {
        "shipped/a.md": {"card": 2, "link": 1},
        "shipped/sub/b.md": {"card": 1},
        "src/help.rs": {"card": 1},
    },
}


def write(root, rel, text):
    path = Path(root) / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    if isinstance(text, bytes):
        path.write_bytes(text)
    else:
        path.write_text(text, encoding="utf-8")
    return path


def run(*args, cwd=None):
    proc = subprocess.run(
        [sys.executable, "-I", str(TOOL), *[str(a) for a in args]],
        capture_output=True,
        text=True,
        cwd=cwd,
    )
    return proc.returncode, proc.stdout, proc.stderr


class TreeCase(unittest.TestCase):
    """A temporary tree with the token file and the baseline beside it, outside the tree."""

    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        base = Path(self._tmp.name)
        self.root = base / "tree"
        self.cfg = base / "cfg"
        for rel, text in TREE.items():
            write(self.root, rel, text)
        write(self.cfg, "tokens.json", json.dumps(TOKENS))
        write(self.cfg, "baseline.json", json.dumps(BASELINE))
        self.tokens = self.cfg / "tokens.json"
        self.baseline = self.cfg / "baseline.json"

    def check(self):
        code, out, err = run(
            "check", "--root", self.root, "--tokens", self.tokens, "--baseline", self.baseline
        )
        return code, out, err

    def check_json(self):
        code, out, err = self.check()
        self.assertEqual(err, "", err)
        return code, json.loads(out)

    def set_baseline(self, data):
        write(self.cfg, "baseline.json", json.dumps(data))

    def append(self, rel, text):
        path = self.root / rel
        path.write_text(path.read_text(encoding="utf-8") + text, encoding="utf-8")


class CountTest(TreeCase):
    def test_baseline_prints_the_counts_of_the_shipped_paths_only(self):
        code, out, err = run("baseline", "--root", self.root, "--tokens", self.tokens)
        self.assertEqual((code, err), (0, ""))
        self.assertEqual(json.loads(out), BASELINE)

    def test_baseline_output_is_stable_and_sorted(self):
        code, first, _ = run("baseline", "--root", self.root, "--tokens", self.tokens)
        self.assertEqual(code, 0)
        second = run("baseline", "--root", self.root, "--tokens", self.tokens)[1]
        self.assertEqual(first, second)
        self.assertEqual(first, json.dumps(json.loads(first), indent=2, sort_keys=True) + "\n")

    def test_check_lists_a_file_that_is_not_utf8_as_skipped_and_does_not_count_it(self):
        write(self.root, "shipped/bin.dat", b"\xff\xfe C-7 C-8")
        code, result = self.check_json()
        self.assertEqual(code, 0)
        self.assertEqual(result["skipped"], ["shipped/bin.dat"])
        self.assertEqual(result["failures"], [])

    def test_a_file_outside_the_shipped_paths_is_never_counted(self):
        write(self.root, "other/new.md", "C-10 C-11 C-12 example/repo")
        write(self.root, "agents/deeper/more.md", "C-13")
        code, result = self.check_json()
        self.assertEqual((code, result["failures"]), (0, []))


class BoundaryTest(TreeCase):
    def test_the_tree_the_baseline_was_taken_from_passes(self):
        code, result = self.check_json()
        self.assertEqual(code, 0)
        self.assertTrue(result["ok"])
        self.assertEqual((result["failures"], result["headroom"], result["skipped"]), ([], [], []))

    def test_one_more_match_in_a_listed_file_fails_and_names_it(self):
        self.append("shipped/a.md", "and C-8\n")
        code, result = self.check_json()
        self.assertEqual(code, 1)
        self.assertFalse(result["ok"])
        self.assertIn(
            {"kind": "file", "file": "shipped/a.md", "token": "card", "count": 3, "baseline": 2},
            result["failures"],
        )
        self.assertIn(
            {"kind": "total", "token": "card", "count": 5, "baseline": 4}, result["failures"]
        )

    def test_one_match_in_a_file_the_baseline_does_not_list_fails(self):
        write(self.root, "shipped/new.md", "C-8\n")
        code, result = self.check_json()
        self.assertEqual(code, 1)
        self.assertIn(
            {"kind": "file", "file": "shipped/new.md", "token": "card", "count": 1, "baseline": 0},
            result["failures"],
        )

    def test_a_listed_file_with_no_entry_for_a_token_counts_that_token_as_zero(self):
        self.append("shipped/sub/b.md", "example/repo\n")
        code, result = self.check_json()
        self.assertEqual(code, 1)
        self.assertIn(
            {"kind": "file", "file": "shipped/sub/b.md", "token": "link", "count": 1, "baseline": 0},
            result["failures"],
        )

    def test_a_match_moved_between_files_fails_on_the_file_rule_alone(self):
        write(self.root, "shipped/a.md", "C-1 and example/repo\n")  # card 2 -> 1
        self.append("shipped/sub/b.md", "and C-10\n")  # card 1 -> 2: the total stays 4
        code, result = self.check_json()
        self.assertEqual(code, 1)
        self.assertEqual(
            result["failures"],
            [{"kind": "file", "file": "shipped/sub/b.md", "token": "card", "count": 2, "baseline": 1}],
        )

    def test_a_total_above_its_baseline_fails_on_the_total_rule_alone(self):
        data = json.loads(json.dumps(BASELINE))
        data["tokens"]["card"] = 3  # the files add up to 4
        self.set_baseline(data)
        code, result = self.check_json()
        self.assertEqual(code, 1)
        self.assertEqual(
            result["failures"], [{"kind": "total", "token": "card", "count": 4, "baseline": 3}]
        )

    def test_a_token_the_baseline_leaves_out_counts_as_zero_everywhere(self):
        data = json.loads(json.dumps(BASELINE))
        del data["tokens"]["link"]
        del data["files"]["shipped/a.md"]["link"]
        self.set_baseline(data)
        code, result = self.check_json()
        self.assertEqual(code, 1)
        self.assertIn(
            {"kind": "total", "token": "link", "count": 1, "baseline": 0}, result["failures"]
        )
        self.assertIn(
            {"kind": "file", "file": "shipped/a.md", "token": "link", "count": 1, "baseline": 0},
            result["failures"],
        )


class HeadroomTest(TreeCase):
    def test_one_match_fewer_passes_and_is_listed_as_headroom(self):
        write(self.root, "shipped/a.md", "C-1 and example/repo\n")
        code, result = self.check_json()
        self.assertEqual((code, result["ok"], result["failures"]), (0, True, []))
        self.assertIn(
            {"kind": "file", "file": "shipped/a.md", "token": "card", "count": 1, "baseline": 2},
            result["headroom"],
        )
        self.assertIn(
            {"kind": "total", "token": "card", "count": 3, "baseline": 4}, result["headroom"]
        )

    def test_a_deleted_listed_file_is_headroom_not_an_error(self):
        (self.root / "shipped/sub/b.md").unlink()
        code, result = self.check_json()
        self.assertEqual((code, result["failures"]), (0, []))
        self.assertIn(
            {"kind": "file", "file": "shipped/sub/b.md", "token": "card", "count": 0, "baseline": 1},
            result["headroom"],
        )


class RefusalTest(TreeCase):
    def refused(self, name):
        code, out, err = self.check()
        self.assertEqual(code, 2, "%s: exit %s, stdout %r, stderr %r" % (name, code, out, err))
        self.assertEqual(out, "", name)
        # Python itself exits 2 for a script it cannot open, so the message must be the tool's.
        self.assertTrue(err.startswith(REFUSAL_PREFIX), "%s: stderr %r" % (name, err))
        return err

    def baseline_with(self, change):
        data = json.loads(json.dumps(BASELINE))
        change(data)
        self.set_baseline(data)

    def tokens_with(self, change):
        data = json.loads(json.dumps(TOKENS))
        change(data)
        write(self.cfg, "tokens.json", json.dumps(data))

    def test_a_missing_token_file(self):
        self.tokens.unlink()
        self.refused("missing token file")

    def test_a_missing_baseline(self):
        self.baseline.unlink()
        self.refused("missing baseline")

    def test_a_token_file_that_is_not_json(self):
        write(self.cfg, "tokens.json", "not json {")
        self.refused("token file not json")

    def test_a_baseline_that_is_not_json(self):
        write(self.cfg, "baseline.json", "")
        self.refused("baseline not json")

    def test_an_extra_top_level_key_in_the_baseline(self):
        self.baseline_with(lambda d: d.update({"extra": 1}))
        self.refused("extra baseline key")

    def test_a_negative_count(self):
        self.baseline_with(lambda d: d["files"]["shipped/a.md"].update({"card": -1}))
        self.refused("negative count")

    def test_a_non_integer_count(self):
        self.baseline_with(lambda d: d["tokens"].update({"card": 4.5}))
        self.refused("float count")

    def test_a_boolean_is_not_a_count(self):
        self.baseline_with(lambda d: d["tokens"].update({"card": True}))
        self.refused("boolean count")

    def test_a_baseline_that_names_a_token_the_token_file_lacks(self):
        self.baseline_with(lambda d: d["tokens"].update({"gone": 1}))
        err = self.refused("unknown token in baseline")
        self.assertIn("gone", err)

    def test_a_baseline_file_entry_for_an_unknown_token(self):
        self.baseline_with(lambda d: d["files"]["shipped/a.md"].update({"gone": 1}))
        self.assertIn("gone", self.refused("unknown token in a file entry"))

    def test_a_version_other_than_one(self):
        self.baseline_with(lambda d: d.update({"version": 2}))
        self.refused("baseline version")

    def test_an_invalid_pattern_names_the_token(self):
        self.tokens_with(lambda d: d["tokens"][0].update({"pattern": "("}))
        self.assertIn("card", self.refused("invalid pattern"))

    def test_a_pattern_that_can_match_the_empty_string_names_the_token(self):
        self.tokens_with(lambda d: d["tokens"][1].update({"pattern": "x*"}))
        self.assertIn("link", self.refused("empty-matching pattern"))

    def test_paths_that_are_not_a_list_of_strings(self):
        self.tokens_with(lambda d: d.update({"paths": "shipped/**"}))
        self.refused("paths shape")

    def test_a_duplicate_token_id(self):
        self.tokens_with(lambda d: d["tokens"][1].update({"id": "card"}))
        self.refused("duplicate id")

    def test_baseline_refuses_a_bad_token_file_the_same_way(self):
        write(self.cfg, "tokens.json", "[]")
        code, out, err = run("baseline", "--root", self.root, "--tokens", self.tokens)
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(REFUSAL_PREFIX), err)


def tree_hash(root):
    digest = hashlib.sha256()
    for dirpath, dirnames, filenames in os.walk(str(root)):
        dirnames.sort()
        for name in sorted(filenames):
            path = Path(dirpath) / name
            digest.update(str(path.relative_to(root)).encode("utf-8"))
            digest.update(path.read_bytes())
    return digest.hexdigest()


class ReadOnlyTest(TreeCase):
    def test_check_and_baseline_write_nothing_in_the_tree_or_beside_it(self):
        before = (tree_hash(self.root), tree_hash(self.cfg))
        passed = run("check", "--root", self.root, "--tokens", self.tokens, "--baseline", self.baseline)
        printed = run("baseline", "--root", self.root, "--tokens", self.tokens)
        self.assertEqual((passed[0], printed[0]), (0, 0))
        self.append("shipped/a.md", "C-99\n")
        mid = (tree_hash(self.root), tree_hash(self.cfg))
        failed = run("check", "--root", self.root, "--tokens", self.tokens, "--baseline", self.baseline)
        self.assertEqual(failed[0], 1)
        self.assertEqual(mid, (tree_hash(self.root), tree_hash(self.cfg)))
        self.assertNotEqual(before[0], mid[0])  # the append was the only change
        self.assertEqual(before[1], mid[1])

    def test_a_refused_run_writes_nothing_either(self):
        self.baseline.unlink()
        before = (tree_hash(self.root), tree_hash(self.cfg))
        code, out, err = run(
            "check", "--root", self.root, "--tokens", self.tokens, "--baseline", self.baseline
        )
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(REFUSAL_PREFIX), err)
        self.assertEqual(before, (tree_hash(self.root), tree_hash(self.cfg)))

    def test_the_repository_is_untouched_by_a_run_on_it(self):
        watched = [REPO_ROOT / "skills" / "moai", REPO_ROOT / "agents", IMPROVE]
        before = [tree_hash(p) for p in watched if p.exists()]
        code, out, err = run("check")
        self.assertEqual((code, err), (0, ""), out)
        self.assertEqual(before, [tree_hash(p) for p in watched if p.exists()])


class RepoTreeTest(unittest.TestCase):
    def test_check_on_this_repository_passes_with_the_checked_in_files(self):
        code, out, err = run("check")
        self.assertEqual((code, err), (0, ""), out)
        result = json.loads(out)
        self.assertTrue(result["ok"])
        self.assertEqual(result["failures"], [])

    def test_the_token_file_names_the_three_shipped_paths(self):
        path = IMPROVE / "shipped-text-tokens.json"
        self.assertTrue(path.is_file(), path)
        data = json.loads(path.read_text(encoding="utf-8"))
        for needed in ("skills/moai/**", "agents/*.md", "src/help.rs"):
            self.assertIn(needed, data["paths"])

    def test_no_file_of_the_scan_is_shipped(self):
        files = [
            "tools/shipped_text_ratchet.py",
            "tools/improve/shipped-text-tokens.json",
            "tools/improve/shipped-text-baseline.json",
            "tests/shipped_text_ratchet_test.py",
        ]
        manifest = json.loads(
            (REPO_ROOT / "skills/moai/references/source-manifest.json").read_text(encoding="utf-8")
        )
        entries = list(manifest["files"])
        for rel in files:
            self.assertTrue((REPO_ROOT / rel).is_file(), rel)
            self.assertFalse(rel.startswith(("skills/moai/", "agents/")), rel)
            self.assertFalse(any(rel.endswith(e) or e.endswith(rel) for e in entries), rel)
        self.assertFalse([e for e in entries if "ratchet" in e or "shipped-text" in e], entries)


if __name__ == "__main__":
    unittest.main()
