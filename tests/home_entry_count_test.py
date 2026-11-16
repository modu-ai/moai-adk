"""Tests for tools/home_entry_count.py (Issue #3068 P2, AC8).

Every case runs the tool on a temporary HOME this test creates. None runs it on a real HOME or
on any path the test did not make; the real-HOME half of AC8 (before and after a verification
run) is the leader's or operator's measurement and is not a test.
"""

import importlib.util
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TOOL = REPO_ROOT / "tools" / "home_entry_count.py"


def load_module():
    spec = importlib.util.spec_from_file_location("home_entry_count", TOOL)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class Home:
    """A temporary HOME with a `.moai` folder of the given entries."""

    def __init__(self, test):
        self.dir = tempfile.TemporaryDirectory(prefix="moai-test-home-")
        test.addCleanup(self.dir.cleanup)
        self.path = os.path.realpath(self.dir.name)

    def moai(self):
        return os.path.join(self.path, ".moai")

    def make(self, *names, secret="SECRET-CONTENT"):
        os.makedirs(self.moai(), exist_ok=True)
        for name in names:
            with open(os.path.join(self.moai(), name), "w", encoding="utf-8") as f:
                f.write(secret)
        return self


def run(*args, env=None):
    """The tool with `args`; the child environment holds only what `env` names."""
    return subprocess.run([sys.executable, "-I", str(TOOL), *args], capture_output=True,
                          text=True, env=env if env is not None else {})


def measure(home):
    out = run("--home", home.path)
    assert out.returncode == 0, out.stderr
    return json.loads(out.stdout)


class CountTest(unittest.TestCase):
    def test_empty_moai_folder_counts_zero(self):
        report = measure(Home(self).make())
        self.assertEqual(report["moai"], {"state": "dir", "count": 0, "names": []})

    def test_count_follows_the_entries_one_by_one(self):
        # N and N+1: the count moves by exactly one entry
        home = Home(self).make("a")
        self.assertEqual(measure(home)["moai"]["count"], 1)
        home.make("b")
        report = measure(home)
        self.assertEqual(report["moai"]["count"], 2)
        self.assertEqual(report["moai"]["names"], ["a", "b"])

    def test_names_are_sorted_and_hidden_entries_count(self):
        home = Home(self).make("zeta", ".hidden", "alpha")
        self.assertEqual(measure(home)["moai"]["names"], [".hidden", "alpha", "zeta"])

    def test_depth_is_one_a_folder_counts_once(self):
        home = Home(self).make("a")
        sub = os.path.join(home.moai(), "projects")
        os.makedirs(os.path.join(sub, "p1", "deep"))
        Path(sub, "p1", "ledger").write_text("x", encoding="utf-8")
        moai = measure(home)["moai"]
        self.assertEqual(moai["names"], ["a", "projects"])
        self.assertEqual(moai["count"], 2)

    def test_absent_moai_is_a_normal_answer_and_creates_nothing(self):
        home = Home(self)
        self.assertEqual(measure(home)["moai"], {"state": "absent", "count": 0, "names": []})
        self.assertEqual(os.listdir(home.path), [])

    def test_a_file_where_the_folder_belongs_is_reported_not_listed(self):
        home = Home(self)
        Path(home.moai()).write_text("not a folder", encoding="utf-8")
        self.assertEqual(measure(home)["moai"], {"state": "file", "count": 0, "names": []})

    def test_a_link_is_reported_and_not_followed(self):
        home = Home(self)
        target = Home(self).make("one", "two", "three")
        os.symlink(target.moai(), home.moai())
        self.assertEqual(measure(home)["moai"], {"state": "link", "count": 0, "names": []})

    def test_no_file_content_reaches_the_output(self):
        home = Home(self).make("a", secret="HOME-SECRET-MARKER")
        out = run("--home", home.path)
        self.assertEqual(out.returncode, 0)
        self.assertNotIn("HOME-SECRET-MARKER", out.stdout + out.stderr)

    def test_the_report_names_its_depth_and_the_home_it_was_given(self):
        home = Home(self).make("a")
        report = measure(home)
        self.assertEqual(report["home"], home.path)
        self.assertEqual(report["depth"], 1)
        self.assertIs(report["contents_read"], False)

    def test_the_tool_changes_nothing(self):
        home = Home(self).make("a", "b")
        before = sorted(os.listdir(home.moai())), os.stat(home.moai()).st_mtime_ns
        measure(home)
        run("--home", home.path, "--compare", os.devnull)
        after = sorted(os.listdir(home.moai())), os.stat(home.moai()).st_mtime_ns
        self.assertEqual(before, after)


class RefusalTest(unittest.TestCase):
    def refused(self, *args, env=None):
        out = run(*args, env=env)
        self.assertEqual(out.returncode, 2, (args, out.stdout, out.stderr))
        self.assertEqual(out.stdout, "", args)
        self.assertTrue(out.stderr.strip(), args)

    def test_no_arguments(self):
        self.refused()

    def test_the_environment_is_never_a_fallback(self):
        home = Home(self).make("a")
        self.refused(env={"HOME": home.path, "USERPROFILE": home.path})
        self.refused("--compare", os.devnull, env={"HOME": home.path})

    def test_home_must_be_an_explicit_absolute_existing_folder(self):
        home = Home(self).make("a")
        self.refused("--home")
        self.refused("--home", "")
        self.refused("--home", "~")
        self.refused("--home", "relative/path")
        self.refused("--home", os.path.join(home.path, "no-such-folder"))
        self.refused("--home", os.path.join(home.moai(), "a"))  # a file

    def test_unknown_and_repeated_options(self):
        home = Home(self)
        self.refused("--home", home.path, "--verbose")
        self.refused("--home", home.path, "--home", home.path)
        self.refused("--home", home.path, "stray")
        self.refused("--compare", os.devnull)

    def test_an_unreadable_moai_folder_is_refused_not_reported_empty(self):
        if os.geteuid() == 0:
            self.skipTest("root reads any folder")
        home = Home(self).make("a")
        os.chmod(home.moai(), 0)
        self.addCleanup(os.chmod, home.moai(), 0o700)
        self.refused("--home", home.path)


class CompareTest(unittest.TestCase):
    def snapshot(self, home):
        path = os.path.join(home.path, "before.json")
        with open(path, "w", encoding="utf-8") as f:
            json.dump(measure(home), f)
        return path

    def compare(self, home, before):
        out = run("--home", home.path, "--compare", before)
        return out.returncode, (json.loads(out.stdout) if out.stdout else None), out.stderr

    def test_unchanged_exits_zero_even_when_contents_change(self):
        home = Home(self).make("a", "b")
        before = self.snapshot(home)
        Path(home.moai(), "a").write_text("rewritten", encoding="utf-8")
        code, report, _ = self.compare(home, before)
        self.assertEqual(code, 0)
        self.assertIs(report["compare"]["unchanged"], True)
        self.assertEqual(report["compare"]["added"], [])
        self.assertEqual(report["compare"]["removed"], [])

    def test_one_entry_more_exits_one_and_names_it(self):
        home = Home(self).make("a", "b")
        before = self.snapshot(home)
        home.make("c")
        code, report, _ = self.compare(home, before)
        self.assertEqual(code, 1)
        self.assertEqual(report["compare"]["added"], ["c"])
        self.assertEqual(report["compare"]["count"], [2, 3])

    def test_one_entry_fewer_exits_one_and_names_it(self):
        home = Home(self).make("a", "b")
        before = self.snapshot(home)
        os.remove(os.path.join(home.moai(), "b"))
        code, report, _ = self.compare(home, before)
        self.assertEqual(code, 1)
        self.assertEqual(report["compare"]["removed"], ["b"])
        self.assertEqual(report["compare"]["count"], [2, 1])

    def test_a_rename_keeps_the_count_and_still_differs(self):
        home = Home(self).make("a", "b")
        before = self.snapshot(home)
        os.rename(os.path.join(home.moai(), "b"), os.path.join(home.moai(), "c"))
        code, report, _ = self.compare(home, before)
        self.assertEqual(code, 1)
        self.assertEqual(report["compare"]["count"], [2, 2])
        self.assertEqual(report["compare"]["added"], ["c"])
        self.assertEqual(report["compare"]["removed"], ["b"])

    def test_a_folder_appearing_where_none_was_differs(self):
        home = Home(self)
        before = self.snapshot(home)
        home.make()
        code, report, _ = self.compare(home, before)
        self.assertEqual(code, 1)
        self.assertEqual(report["compare"]["state"], ["absent", "dir"])
        self.assertEqual(report["compare"]["count"], [0, 0])

    def test_a_before_file_of_another_home_is_refused(self):
        one, other = Home(self).make("a"), Home(self).make("a")
        before = self.snapshot(one)
        code, report, stderr = self.compare(other, before)
        self.assertEqual((code, report), (2, None))
        self.assertIn("another HOME", stderr)

    def test_a_missing_or_malformed_before_file_is_refused(self):
        home = Home(self).make("a")
        bad = os.path.join(home.path, "bad.json")
        cases = [
            os.path.join(home.path, "no-such.json"),
            home.path,  # a folder
        ]
        for text in ("", "not json", "[]", "{}", json.dumps({"home": home.path}),
                     json.dumps({"home": home.path,
                                 "moai": {"state": "dir", "count": 2, "names": ["a"]}}),
                     json.dumps({"home": home.path,
                                 "moai": {"state": "dir", "count": 1, "names": [1]}})):
            Path(bad).write_text(text, encoding="utf-8")
            cases.append(bad)
            code, report, stderr = self.compare(home, bad)
            self.assertEqual((code, report), (2, None), text)
            self.assertTrue(stderr.strip(), text)
        for path in cases[:2]:
            code, report, _ = self.compare(home, path)
            self.assertEqual((code, report), (2, None), path)


class DocumentedTest(unittest.TestCase):
    def test_the_header_states_the_real_home_procedure(self):
        doc = " ".join(load_module().__doc__.split())
        for phrase in ("AC8", "before and after", "real HOME", "depth 1", "Exit 2",
                       "never defaults to the real HOME", "user confirms"):
            self.assertIn(phrase, doc)


if __name__ == "__main__":
    unittest.main()
