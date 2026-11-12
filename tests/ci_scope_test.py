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

    def test_space_applied_record_alone_is_full(self):
        # Only the Space record changed: space_applied_record_test must still run (#2207).
        for path in ("docs/design-review/applied.md", "docs/design-review/pages.tsv"):
            scope, _ = self.ci_scope.classify([path])
            self.assertEqual(scope, "full", path)

    def test_design_review_fixes_prefix_is_full(self):
        # tests/space_applied_record_test.py hashes every fix file (#2207).
        scope, reasons = self.ci_scope.classify(["docs/design-review/fixes/group3-487fa473.md"])
        self.assertEqual(scope, "full")
        self.assertTrue(any("prefix-denied" in r for r in reasons), reasons)

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


class DocumentationDriftTest(unittest.TestCase):
    """#2538 B2: every docs/ path a test, tool or src file names must be denied.

    A path referenced by name in tests/, tools/ or src/ is read, compared or
    embedded by that code; classifying it docs-only would skip the very check
    that consumes it. The scan is the same mechanical grep that produced
    EXACT_DENY, so adding a new reader without extending the denylist fails
    here first.
    """

    IGNORED_LITERAL_PREFIXES = (
        "docs/a.md", "docs/A.md", "docs/c.svg", "docs/d.txt", "docs/img/",
        "docs/x.rs", "docs/old.md", "docs/huge.md", "docs/ssh.md",
        "docs/spec.md", "docs/receipt.md", "docs/elsewhere.md",
        "docs/missing.md", "docs/untracked.md", "docs/staged.md",
        "docs/link.md", "docs/decisions-test.md", "docs/decisions.md",
        "docs/TBD-related.md", "docs/workflows/", "docs/wf",
    )


    def _fragmented_docs_paths(self):
        """Paths built from pieces like ROOT / "docs" / "x.md" or join("docs", "x.md").

        #2538 B2prime: the quoted-string scan misses these, so collect them with a
        separate pattern and deny their joined forms.
        """
        import re as _re
        import subprocess as _sp
        # Python: X / "docs" / "name"  and  os.path.join(..., "docs", "name")
        py = _sp.run(
            ["grep", "-rnE",
             r'["\x27]docs["\x27]\s*\/\s*["\x27]([A-Za-z0-9_./-]+)["\x27]',
             "tests", "tools", "src"],
            capture_output=True, text=True, cwd=REPO_ROOT,
        ).stdout
        # Line-filtered: skip this test's own docstrings and temp-dir fixtures
        # (a path under a tmp/scratch root never reads the repository docs/).
        keep = []
        for line in py.splitlines():
            if "ci_scope_test.py" in line:
                continue
            if "self.tmp" in line or "tmp.path" in line or "mkdtemp" in line:
                continue
            keep.append(line)
        py = "\n".join(keep)
        names = set(_re.findall(r'["\x27]docs["\x27]\s*/\s*["\x27]([A-Za-z0-9_./-]+)["\x27]', py))
        # Rust: .join("docs").join("x.md") or single .join("x.md") after a docs root
        rs_proc = _sp.run(
            ["grep", "-rnE",
             r'join\("docs"\)(?:\s*\.join\("[^"]+"\))*',
             "tests", "src"],
            capture_output=True, text=True, cwd=REPO_ROOT,
        )
        rs = "\n".join(
            line for line in rs_proc.stdout.splitlines()
            if "ci_scope_test.py" not in line
            and "self.tmp" not in line and "tmp.path" not in line
        )
        for chunk in rs.splitlines():
            for seg in _re.findall(r'join\("([A-Za-z0-9_./-]+)"\)', chunk)[1:]:
                if "." in seg:
                    names.add(seg)
        return {f"docs/{n}" for n in names if "." in n.rsplit("/", 1)[-1]}

    def test_fragmented_docs_paths_are_denied(self):
        found = self._fragmented_docs_paths()
        module = load_module()
        missing = sorted(p for p in found if p not in module.EXACT_DENY)
        self.assertEqual(missing, [], f"fragmented docs paths not denied: {missing}")

    def test_every_referenced_docs_path_is_denied(self):
        import re as _re
        import subprocess as _sp
        pattern = _re.compile(r'["\']((?:docs|\.\.\/docs)\/[A-Za-z0-9_./-]+)["\']')
        proc = _sp.run(
            ["grep", "-rhoE", r'["\x27](docs|\.\./docs)/[A-Za-z0-9_./-]+["\x27]',
             "tests", "tools", "src"],
            capture_output=True, text=True, cwd=REPO_ROOT,
        )
        quoted = set()
        for token in proc.stdout.splitlines():
            path = token.strip().strip('"').strip("'")
            path = path.replace("../docs/", "docs/", 1).rstrip("\\")
            if not path.startswith("docs/"):
                continue
            if any(path.startswith(p) for p in self.IGNORED_LITERAL_PREFIXES):
                continue
            module_now = load_module()
            if ("." not in path.rsplit("/", 1)[-1]
                    and f"{path}.md" in module_now.EXACT_DENY):
                continue  # prose shard of a denied .md file (usage/row strings)
            if path in ("docs/e2e-receipts", "docs/e2e-receipts/"):
                continue  # covered by PREFIX_DENY
            if path.startswith("docs/e2e-receipts/"):
                continue  # covered by PREFIX_DENY
            if any((path.rstrip("/") + "/").startswith(p) for p in module_now.PREFIX_DENY):
                continue  # a prefix-denied folder, named whole or by one of its files
            if path == "docs/support-matrix":
                continue  # prose shard of docs/support-matrix.md in usage/row strings; the .md path itself is denied
            if not path[5].isalnum() and path[5] != "_":
                continue  # trailing escape artifact from multi-line strings
            quoted.add(path)
        module = load_module()
        missing = sorted(p for p in quoted if p not in module.EXACT_DENY)
        self.assertEqual(missing, [], f"docs paths referenced by code but not denied: {missing}")


if __name__ == "__main__":
    unittest.main()
