"""Tests for tools/ci_scope.py (#2527, design comment 6822993459)."""

import importlib.util
import json
import os
import subprocess
import sys
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def load_module():
    spec = importlib.util.spec_from_file_location(
        "ci_scope", os.path.join(REPO_ROOT, "tools", "ci_scope.py")
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CiScopeTest(unittest.TestCase):
    def setUp(self):
        self.ci_scope = load_module()

    def run_stdin(self, text):
        proc = subprocess.run(
            [sys.executable, os.path.join(REPO_ROOT, "tools", "ci_scope.py")],
            input=text,
            capture_output=True,
            text=True,
        )
        return proc

    def test_all_docs_is_docs_only(self):
        scope, reasons = self.ci_scope.classify(
            ["docs/a.md", "docs/img/b.png", "docs/c.svg", "docs/d.txt"]
        )
        self.assertEqual(scope, "docs-only")
        self.assertEqual(reasons, [])

    def test_one_src_file_is_full(self):
        scope, _ = self.ci_scope.classify(["docs/a.md", "src/main.rs"])
        self.assertEqual(scope, "full")

    def test_exact_deny_files_are_full(self):
        for path in sorted(self.ci_scope.EXACT_DENY):
            scope, reasons = self.ci_scope.classify([path])
            self.assertEqual(scope, "full", path)
            self.assertTrue(any(path in r for r in reasons), path)

    def test_e2e_receipts_prefix_is_full(self):
        scope, _ = self.ci_scope.classify(["docs/e2e-receipts/receipts.jsonl"])
        self.assertEqual(scope, "full")

    def test_non_doc_suffix_is_full(self):
        scope, _ = self.ci_scope.classify(["docs/x.rs"])
        self.assertEqual(scope, "full")

    def test_rename_from_src_is_full(self):
        scope, _ = self.ci_scope.classify(["src/a.rs", "docs/a.md"])
        self.assertEqual(scope, "full")

    def test_docs_deletion_only_is_docs_only(self):
        scope, _ = self.ci_scope.classify(["docs/old.md"])
        self.assertEqual(scope, "docs-only")

    def test_reports_path_is_docs_only(self):
        scope, _ = self.ci_scope.classify([".moai-cli/reports/issue-copy.md"])
        self.assertEqual(scope, "docs-only")

    def test_empty_input_is_full(self):
        scope, reasons = self.ci_scope.classify([])
        self.assertEqual(scope, "full")
        self.assertTrue(reasons)

    def test_single_blank_line_is_full(self):
        # The workflow pipes `git diff --name-only`; an empty diff yields one blank
        # line, which must fail safe to full exactly like empty stdin (#2527).
        scope, reasons = self.ci_scope.classify([""])
        self.assertEqual(scope, "full")
        self.assertTrue(reasons)
        proc = self.run_stdin("\n")
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "full")

    def test_garbage_input_is_full(self):
        proc = self.run_stdin("not a real path\n\x00weird\n")
        self.assertEqual(proc.returncode, 0)
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "full")

    def test_script_exit_zero_and_json(self):
        proc = self.run_stdin("docs/a.md\n")
        self.assertEqual(proc.returncode, 0)
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "docs-only")
        self.assertEqual(data["reasons"], [])

    def test_script_src_file_full(self):
        proc = self.run_stdin("src/a.rs\n")
        self.assertEqual(proc.returncode, 0)
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "full")
        self.assertTrue(data["reasons"])


class WorkflowContractTest(unittest.TestCase):
    def test_scope_steps_pass_rename_sources(self):
        # #2538 B1: plain `git diff --name-only` emits only the rename destination,
        # so renaming/deleting a test-consumed doc could pass as docs-only. Both
        # Scope steps must use --no-renames so old paths reach the classifier.
        with open(
            os.path.join(REPO_ROOT, ".github", "workflows", "rc-ci.yml"),
            "r",
            encoding="utf-8",
        ) as fh:
            yaml_text = fh.read()
        self.assertEqual(yaml_text.count("git diff --name-only --no-renames"), 2)
        self.assertNotIn('git diff --name-only "${base}', yaml_text)


class MainExceptionPathTest(unittest.TestCase):
    def test_internal_error_fails_safe_to_full(self):
        # #2538 audit note (M3): the except branch in main() must answer "full".
        from unittest import mock
        from io import StringIO

        module = load_module()
        with mock.patch.object(module, "classify", side_effect=RuntimeError("boom")):
            with mock.patch.object(module.sys, "stdin", StringIO("docs/a.md\n")):
                with mock.patch.object(module.sys, "stdout", StringIO()) as out:
                    code = module.main()
        self.assertEqual(code, 0)
        data = json.loads(out.getvalue())
        self.assertEqual(data["scope"], "full")
        self.assertTrue(any("boom" in r for r in data["reasons"]))


if __name__ == "__main__":
    unittest.main()
