"""Tests for tools/audit_eval.py and the audit eval set (SPEC-20261119-audit-eval, card C-102).

The eval set replays real audit FAILs; the tool builds audit prompts (dry-run) and scores
findings against one-sentence answer keys. Everything here is local: no network, no model.
"""

import importlib.util
import json
import os
import subprocess
import sys
import tempfile
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
TOOL = os.path.join(REPO_ROOT, "tools", "audit_eval.py")
EVAL_DIR = os.path.join(REPO_ROOT, ".moai", "evals", "audit")


def load_module():
    spec = importlib.util.spec_from_file_location("audit_eval", TOOL)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


HEX40 = set("0123456789abcdef")


def is_sha40(text):
    return len(text) == 40 and all(c in HEX40 for c in text)


class SetTest(unittest.TestCase):
    """AC-1: twelve incidents, split 6/6 with no shared id, and size-matched controls."""

    def test_the_set_holds_twelve_incidents_and_six_controls(self):
        tuning = json.load(open(os.path.join(EVAL_DIR, "cases-tuning.json")))
        heldout = json.load(open(os.path.join(EVAL_DIR, "cases-heldout.json")))
        controls = json.load(open(os.path.join(EVAL_DIR, "controls.json")))
        ids = [c["id"] for c in tuning] + [c["id"] for c in heldout]
        self.assertEqual(len(ids), len(set(ids)), "a case id appears twice")
        self.assertGreaterEqual(len(ids), 10, "at least ten incidents")
        self.assertEqual(len(tuning), 6)
        self.assertEqual(len(heldout), 6)
        self.assertGreaterEqual(len(controls), len(ids) // 2, "one control per two incidents")
        for case in tuning + heldout:
            for field in ("pr", "audit_comment"):
                self.assertIsInstance(case[field], int, case["id"])
            for field in ("head", "base"):
                self.assertTrue(is_sha40(case[field]), f"{case['id']}.{field}")
            key = case["answer_key"]
            self.assertTrue(key.strip() and "\n" not in key, f"{case['id']}: one sentence")
        for control in controls:
            self.assertTrue(is_sha40(control["head"]), control["id"])
            self.assertTrue(is_sha40(control["base"]), control["id"])
            self.assertIn("matched_case", control, "each control names the case it matches")


class RunTest(unittest.TestCase):
    """AC-2: dry-run prompts carry head and base but never the answer key."""

    def setUp(self):
        self.audit_eval = load_module()
        self.tmp = tempfile.TemporaryDirectory()
        self.eval_dir = os.path.join(self.tmp.name, "evals")
        os.makedirs(self.eval_dir)
        self.out = os.path.join(self.tmp.name, "prompts")
        # Two real commits of this repository, so the ref check has something to find.
        head = subprocess.run(
            ["git", "-C", REPO_ROOT, "rev-parse", "HEAD"], capture_output=True, text=True
        ).stdout.strip()
        parent = subprocess.run(
            ["git", "-C", REPO_ROOT, "rev-parse", "HEAD~1"], capture_output=True, text=True
        ).stdout.strip()
        case = {
            "id": "case-ok",
            "pr": 1,
            "audit_comment": 1,
            "head": head,
            "base": parent,
            "answer_key": "The guard that refuses a broken marker is never exercised.",
        }
        json.dump([case], open(os.path.join(self.eval_dir, "cases-tuning.json"), "w"))
        json.dump([], open(os.path.join(self.eval_dir, "cases-heldout.json"), "w"))
        control = {
            "id": "ctl-ok",
            "pr": 2,
            "audit_comment": 2,
            "head": head,
            "base": parent,
        }
        json.dump([control], open(os.path.join(self.eval_dir, "controls.json"), "w"))
        self.case = case
        self.head, self.parent = head, parent

    def tearDown(self):
        self.tmp.cleanup()

    def cli(self, *args):
        return subprocess.run(
            [sys.executable, TOOL, *args], capture_output=True, text=True, cwd=REPO_ROOT
        )

    def test_dry_run_builds_prompts_without_the_answer_key(self):
        proc = self.cli("run", "--set", "all", "--eval-dir", self.eval_dir, "--out", self.out)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        summary = json.loads(proc.stdout)
        self.assertEqual(summary["written"], 2)
        self.assertEqual(summary["missing_ref"], [])
        for name in ("case-ok.md", "ctl-ok.md"):
            text = open(os.path.join(self.out, name)).read()
            self.assertIn(self.head, text, name)
            self.assertIn(self.parent, text, name)
        case_text = open(os.path.join(self.out, "case-ok.md")).read()
        self.assertNotIn(self.case["answer_key"], case_text)
        self.assertNotIn("answer_key", case_text)

    def test_a_missing_ref_is_reported_and_skipped(self):
        case = dict(self.case, id="case-missing", head="0" * 40)
        json.dump(
            [self.case, case],
            open(os.path.join(self.eval_dir, "cases-tuning.json"), "w"),
        )
        proc = self.cli("run", "--set", "tuning", "--eval-dir", self.eval_dir, "--out", self.out)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        summary = json.loads(proc.stdout)
        self.assertEqual(summary["missing_ref"], ["case-missing:head"])
        self.assertEqual(summary["written"], 1)

    def test_a_shared_id_between_the_files_is_refused(self):
        json.dump(
            [self.case], open(os.path.join(self.eval_dir, "cases-heldout.json"), "w")
        )
        proc = self.cli("run", "--set", "all", "--eval-dir", self.eval_dir, "--out", self.out)
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("case-ok", proc.stderr)
        self.assertIn("tuning", proc.stderr)
        self.assertIn("held-out", proc.stderr)

    def test_a_record_without_its_fields_is_refused(self):
        broken = dict(self.case, id="case-broken")
        del broken["answer_key"]
        json.dump([broken], open(os.path.join(self.eval_dir, "cases-tuning.json"), "w"))
        proc = self.cli("run", "--set", "tuning", "--eval-dir", self.eval_dir, "--out", self.out)
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("case-broken", proc.stderr)
        self.assertIn("answer_key", proc.stderr)


class ScoreTest(unittest.TestCase):
    """AC-3: only findings that describe the bug count; verdict recall is separate."""

    def setUp(self):
        self.audit_eval = load_module()
        self.tmp = tempfile.TemporaryDirectory()
        self.eval_dir = os.path.join(self.tmp.name, "evals")
        os.makedirs(self.eval_dir)
        head = subprocess.run(
            ["git", "-C", REPO_ROOT, "rev-parse", "HEAD"], capture_output=True, text=True
        ).stdout.strip()
        parent = subprocess.run(
            ["git", "-C", REPO_ROOT, "rev-parse", "HEAD~1"], capture_output=True, text=True
        ).stdout.strip()
        self.key = "The guard that refuses a broken marker is never exercised by any test."
        cases = []
        for cid in ("case-describe", "case-verdict-only", "case-miss", "case-raw-only"):
            cases.append(
                {
                    "id": cid,
                    "pr": 1,
                    "audit_comment": 1,
                    "head": head,
                    "base": parent,
                    "answer_key": self.key,
                }
            )
        json.dump(cases, open(os.path.join(self.eval_dir, "cases-tuning.json"), "w"))
        json.dump([], open(os.path.join(self.eval_dir, "cases-heldout.json"), "w"))
        json.dump(
            [{"id": "ctl-flag", "pr": 2, "audit_comment": 2, "head": head, "base": parent},
             {"id": "ctl-clean", "pr": 3, "audit_comment": 3, "head": head, "base": parent}],
            open(os.path.join(self.eval_dir, "controls.json"), "w"),
        )
        self.findings = os.path.join(self.tmp.name, "findings.json")
        self.head, self.parent = head, parent

    def tearDown(self):
        self.tmp.cleanup()

    def cli(self, rows):
        json.dump(rows, open(self.findings, "w"))
        return subprocess.run(
            [
                sys.executable,
                TOOL,
                "score",
                "--set",
                "tuning",
                "--eval-dir",
                self.eval_dir,
                "--findings",
                self.findings,
            ],
            capture_output=True,
            text=True,
            cwd=REPO_ROOT,
        )

    def test_recall_is_separate_and_only_describing_findings_count(self):
        described = "No test reaches the guard that refuses a broken marker, so a mutant removing it survives."
        rows = [
            {"case_id": "case-describe", "findings": [described], "verdict": "FAIL"},
            {"case_id": "case-verdict-only", "findings": ["Unrelated style remark"], "verdict": "FAIL"},
            {"case_id": "case-miss", "findings": ["Unrelated style remark"], "verdict": "PASS"},
            {"case_id": "case-raw-only", "findings": [described], "verdict": "PASS"},
            {"case_id": "ctl-flag", "findings": ["Anything at all"], "verdict": "FAIL"},
            {"case_id": "ctl-clean", "findings": [], "verdict": "PASS"},
        ]
        proc = self.cli(rows)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        report = json.loads(proc.stdout)
        # described: case-describe and case-raw-only → raw recall 2/4.
        self.assertEqual(report["cases"], 4)
        self.assertEqual(report["described"], 2)
        self.assertAlmostEqual(report["recall_raw"], 0.5)
        # verdict recall needs a described finding AND a FAIL verdict: 1/4.
        self.assertAlmostEqual(report["recall_verdict"], 0.25)
        # controls: one flagged of two.
        self.assertEqual(report["controls"], 2)
        self.assertEqual(report["controls_flagged"], 1)
        self.assertAlmostEqual(report["control_overflag"], 0.5)

    def test_the_token_fraction_is_a_constant(self):
        self.assertAlmostEqual(self.audit_eval.TOKEN_FRACTION, 0.5)

    def test_an_unknown_case_or_a_non_list_file_refuses(self):
        proc = self.cli([{"case_id": "case-unknown", "findings": [], "verdict": "PASS"}])
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("case-unknown", proc.stderr)
        json.dump({"not": "a list"}, open(self.findings, "w"))
        proc = subprocess.run(
            [
                sys.executable, TOOL, "score", "--set", "tuning",
                "--eval-dir", self.eval_dir, "--findings", self.findings,
            ],
            capture_output=True, text=True, cwd=REPO_ROOT,
        )
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("list", proc.stderr)


class ReadmeTest(unittest.TestCase):
    """AC-4: the README says how to add a case, how to run and score, and when to stop."""

    def test_the_readme_carries_the_three_sections(self):
        text = open(os.path.join(EVAL_DIR, "README.md")).read()
        for heading in ("## Adding a case", "## Running and scoring", "## When to stop tuning"):
            self.assertIn(heading, text)
        self.assertIn("held-out", text.split("## When to stop tuning", 1)[1])


class FollowupsTest(unittest.TestCase):
    """Card C-107 slice 1 (SPEC-20261119-audit-eval-followups): the secret-shape ignore rules
    cover .moai/evals/, the ref check pins GIT_NO_LAZY_FETCH in the child environment and the
    missing_ref entries name which ref of which case is missing, and a case id with a path
    separator is refused before anything is written."""

    def test_the_ignore_rules_cover_the_evals_folder(self):
        for name in ("x.key", "id_rsa_fixture", "nested/secrets.json"):
            path = f".moai/evals/{name}"
            r = subprocess.run(
                ["git", "check-ignore", "-v", path], capture_output=True, text=True, cwd=REPO_ROOT
            )
            self.assertEqual(r.returncode, 0, f"{path} is not ignored: {r.stdout}{r.stderr}")
            self.assertIn(path, r.stdout, path)

    def test_the_ref_check_sets_git_no_lazy_fetch(self):
        import textwrap

        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        bin_dir = os.path.join(tmp.name, "bin")
        os.makedirs(bin_dir)
        env_file = os.path.join(tmp.name, "env.txt")
        stub = os.path.join(bin_dir, "git")
        with open(stub, "w") as fh:
            fh.write(
                textwrap.dedent(
                    f"""\
                    #!/bin/sh
                    env | sort > "{env_file}"
                    exit 1
                    """
                )
            )
        os.chmod(stub, 0o755)
        # ref_missing builds the child environment from os.environ, so the stub reaches it
        # through PATH, not through a local dict.
        old_path = os.environ.get("PATH", "")
        os.environ["PATH"] = bin_dir + os.pathsep + old_path
        self.addCleanup(os.environ.__setitem__, "PATH", old_path)
        module = load_module()
        self.assertTrue(module.ref_missing(tmp.name, "0" * 40), "the stub fails the check")
        self.assertIn("GIT_NO_LAZY_FETCH=1", open(env_file).read())

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.eval_dir = os.path.join(self.tmp.name, "evals")
        os.makedirs(self.eval_dir)
        self.out = os.path.join(self.tmp.name, "prompts")
        head = subprocess.run(
            ["git", "-C", REPO_ROOT, "rev-parse", "HEAD"], capture_output=True, text=True
        ).stdout.strip()
        parent = subprocess.run(
            ["git", "-C", REPO_ROOT, "rev-parse", "HEAD~1"], capture_output=True, text=True
        ).stdout.strip()
        self.case = {
            "id": "case-ok",
            "pr": 1,
            "audit_comment": 1,
            "head": head,
            "base": parent,
            "answer_key": "The guard that refuses a broken marker is never exercised.",
        }
        json.dump([self.case], open(os.path.join(self.eval_dir, "cases-tuning.json"), "w"))
        json.dump([], open(os.path.join(self.eval_dir, "cases-heldout.json"), "w"))
        json.dump([], open(os.path.join(self.eval_dir, "controls.json"), "w"))
        self.head, self.parent = head, parent

    def cli(self, *args):
        return subprocess.run(
            [sys.executable, TOOL, *args], capture_output=True, text=True, cwd=REPO_ROOT
        )

    def test_a_missing_ref_names_which_ref(self):
        missing_head = dict(self.case, id="case-head", head="0" * 40)
        missing_base = dict(self.case, id="case-base", base="1" * 40)
        json.dump(
            [missing_head, missing_base],
            open(os.path.join(self.eval_dir, "cases-tuning.json"), "w"),
        )
        proc = self.cli("run", "--set", "tuning", "--eval-dir", self.eval_dir, "--out", self.out)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        summary = json.loads(proc.stdout)
        self.assertEqual(summary["missing_ref"], ["case-base:base", "case-head:head"], summary)
        self.assertEqual(summary["written"], 0)

    def test_a_path_separator_in_a_case_id_is_refused(self):
        for evil in ("../escape", "a/b", "a\\b", ".."):
            json.dump(
                [dict(self.case, id=evil)],
                open(os.path.join(self.eval_dir, "cases-tuning.json"), "w"),
            )
            proc = self.cli(
                "run", "--set", "tuning", "--eval-dir", self.eval_dir, "--out", self.out
            )
            self.assertNotEqual(proc.returncode, 0, evil)
            self.assertIn("path separator", proc.stderr, evil)
            self.assertFalse(
                os.path.isdir(self.out) and os.listdir(self.out), f"{evil}: wrote something"
            )


if __name__ == "__main__":
    unittest.main()
