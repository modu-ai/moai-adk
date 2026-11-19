"""Offline tests for tools/check_guard.py (C-97, SPEC-20261119-check-guard).

The guard decides whether a pull request touches a protected check path and whether it
declared `audit-tier: 2`. Everything runs in temporary directories with fixture lists;
no network, no `gh` — the label step goes through the module's `run_gh` seam, which these
tests replace with a recorder. The verdict contract is REQ-3 to REQ-9 of the SPEC.
"""
import contextlib
import importlib.util
import io
import json
import pathlib
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
GUARD = ROOT / "tools" / "check_guard.py"

LABEL_ARGS = ["--pr", "9", "--repo", "example/repo", "--label", "self-modifying check"]


def load_guard():
    spec = importlib.util.spec_from_file_location("check_guard", GUARD)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


class TempCase(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.tmp = pathlib.Path(self._tmp.name)
        self.addCleanup(self._tmp.cleanup)
        self.list_text = json.dumps(LIST, indent=2) + "\n"

    def write(self, name, text):
        path = self.tmp / name
        path.write_text(text, encoding="utf-8")
        return str(path)

    def run_guard(self, mod, files, body, extra=None):
        args = [
            "--list", self.write("list.json", self.list_text),
            "--files", self.write("files.txt", files),
            "--body", self.write("body.md", body),
            "--summary", str(self.tmp / "summary.md"),
        ]
        if extra is not None:
            args += extra
        out = io.StringIO()
        with contextlib.redirect_stdout(out):
            code = mod.main(args)
        path = self.tmp / "summary.md"
        summary = path.read_text(encoding="utf-8") if path.exists() else ""
        try:
            verdict = json.loads(out.getvalue())
        except json.JSONDecodeError:
            verdict = None
        return code, verdict, summary


LIST = {
    "schema_version": 1,
    "protected": [
        ".github/workflows/**",
        "tools/ci_scope.py",
        "tools/check_guard.py",
        "tools/protected_checks.json",
        "src/gate*.rs",
        "src/audit_gate.rs",
        "REVIEW.md",
        "agents/audit.md",
    ],
}


class CheckGuardPassTest(TempCase):
    """AC-1 (REQ-3, REQ-4): protected paths + a declared tier 2 passes."""

    def test_protected_paths_with_declared_tier_2_pass_and_label_is_attempted(self):
        mod = load_guard()
        calls = []
        mod.run_gh = lambda args: calls.append(args) or (0, "")
        code, verdict, summary = self.run_guard(
            mod,
            "tools/ci_scope.py\nREVIEW.md\n",
            "audit-tier: 2 gate and review instructions\n\nbody\n",
            extra=LABEL_ARGS,
        )
        self.assertEqual(code, 0)
        self.assertTrue(verdict["protected"])
        self.assertEqual(verdict["declared_tier"], 2)
        self.assertEqual(verdict["verdict"], "pass")
        self.assertIn("tools/ci_scope.py", verdict["protected_paths"])
        self.assertIn("REVIEW.md", verdict["protected_paths"])
        self.assertEqual(verdict["label"], "attempted")
        self.assertEqual(len(calls), 2, "label create then PR add")
        self.assertIn("self-modifying check", summary)
        self.assertIn("tools/ci_scope.py", summary)

    def test_protected_list_edit_with_tier_2_still_passes(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(
            mod,
            "tools/protected_checks.json\nsrc/gate_collect.rs\n",
            "audit-tier: 2\n",
        )
        self.assertEqual(code, 0)
        self.assertTrue(verdict["protected"])


class CheckGuardForceTest(TempCase):
    """AC-2 (REQ-5): protected paths without an exact tier 2 declaration fail."""

    def test_protected_paths_without_declaration_fail_with_declare_reason(self):
        mod = load_guard()
        code, verdict, summary = self.run_guard(
            mod, "REVIEW.md\n", "a body that never declares\n", extra=LABEL_ARGS
        )
        self.assertEqual(code, 1)
        self.assertEqual(verdict["declared_tier"], None)
        self.assertEqual(verdict["verdict"], "fail")
        self.assertIn("declare audit-tier: 2", summary)

    def test_protected_paths_with_declared_zero_and_one_fail(self):
        mod = load_guard()
        for tier in ("0", "1"):
            with self.subTest(tier=tier):
                code, verdict, summary = self.run_guard(
                    mod, "agents/audit.md\n", f"audit-tier: {tier}\n"
                )
                self.assertEqual(code, 1)
                self.assertEqual(verdict["declared_tier"], int(tier))
                self.assertIn("declare audit-tier: 2", summary)

    def test_malformed_declaration_counts_as_absent(self):
        mod = load_guard()
        for body in ("audit-tier: 3\n", "audit-tier:0\n", "see audit-tier: 0\n"):
            with self.subTest(body=body):
                code, verdict, _ = self.run_guard(mod, "REVIEW.md\n", body)
                self.assertEqual(code, 1)
                self.assertEqual(verdict["declared_tier"], None)

    def test_declaration_is_read_from_the_first_ten_lines_only(self):
        mod = load_guard()
        late = "\n" * 10 + "audit-tier: 2\n"
        code, verdict, _ = self.run_guard(mod, "REVIEW.md\n", late)
        self.assertEqual(code, 1)


class CheckGuardMatchTest(TempCase):
    """AC-3 (REQ-3): glob, prefix and exact entry shapes."""

    def test_glob_entry_matches_gate_modules(self):
        mod = load_guard()
        for path in ("src/gate.rs", "src/gate_collect.rs"):
            with self.subTest(path=path):
                code, verdict, _ = self.run_guard(mod, f"{path}\n", "audit-tier: 2\n")
                self.assertEqual(code, 0)
                self.assertIn(path, verdict["protected_paths"])
        code, verdict, _ = self.run_guard(mod, "src/audit_gate.rs\n", "audit-tier: 2\n")
        self.assertEqual(code, 0)
        self.assertIn("src/audit_gate.rs", verdict["protected_paths"])

    def test_workflow_prefix_locks_only_its_folder(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(mod, ".github/workflows/ci.yml\n", "audit-tier: 2\n")
        self.assertEqual(code, 0)
        self.assertTrue(verdict["protected"])
        code, verdict, _ = self.run_guard(mod, ".github/moai/labels.yml\n", "audit-tier: 0\n")
        self.assertEqual(code, 0)
        self.assertFalse(verdict["protected"])

    def test_exact_entries_do_not_match_subpaths(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(mod, "docs/notes/REVIEW.md\n", "audit-tier: 0\n")
        self.assertEqual(code, 0)
        self.assertFalse(verdict["protected"])


class CheckGuardSelfEditTest(TempCase):
    """AC-4: editing the protected list or REVIEW.md is itself flagged; the base
    list is the only list the guard reads, so a weakened copy changes nothing."""

    def test_pr_editing_the_list_itself_is_flagged(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(
            mod, "tools/protected_checks.json\n", "audit-tier: 0\n"
        )
        self.assertEqual(code, 1)
        self.assertTrue(verdict["protected"])

    def test_pr_weakening_review_md_still_hits_the_base_list(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(mod, "REVIEW.md\n", "audit-tier: 0\n")
        self.assertEqual(code, 1)
        self.assertTrue(verdict["protected"])

    def test_pr_list_copy_never_reaches_the_guard(self):
        mod = load_guard()
        weaker = {"schema_version": 1, "protected": ["docs/**"]}
        self.write("weaker.json", json.dumps(weaker))
        code, verdict, _ = self.run_guard(mod, "REVIEW.md\n", "audit-tier: 0\n")
        self.assertEqual(code, 1, "the guard never reads the PR's list copy")


class CheckGuardRenameTest(TempCase):
    """AC-5: rename old and new names and deletions are caught."""

    def test_rename_into_a_protected_path_is_flagged_by_the_new_name(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(mod, "src/gate.rs\n", "audit-tier: 0\n")
        self.assertEqual(code, 1)

    def test_rename_out_of_a_protected_path_is_flagged_by_the_old_name(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(mod, "src/audit_gate.rs\n", "audit-tier: 0\n")
        self.assertEqual(code, 1)

    def test_deleting_a_protected_path_is_flagged(self):
        mod = load_guard()
        code, verdict, _ = self.run_guard(mod, "tools/ci_scope.py\n", "audit-tier: 2\n")
        self.assertEqual(code, 0)
        self.assertTrue(verdict["protected"])


class CheckGuardUnprotectedTest(TempCase):
    """AC-6 (REQ-6): no protected path means a pass."""

    def test_docs_only_changes_pass_at_any_tier(self):
        mod = load_guard()
        for body in ("audit-tier: 0\n", "no declaration at all\n"):
            with self.subTest(body=body):
                code, verdict, summary = self.run_guard(mod, "docs/guide.md\n", body)
                self.assertEqual(code, 0)
                self.assertFalse(verdict["protected"])
                self.assertIn("no protected path", summary)


class CheckGuardSafeTest(TempCase):
    """AC-7 (REQ-7, REQ-8, REQ-9): fail safe and the label seam."""

    def test_broken_list_json_fails_safe(self):
        mod = load_guard()
        self.list_text = "{ not json"
        code, verdict, summary = self.run_guard(mod, "docs/guide.md\n", "audit-tier: 0\n")
        self.assertEqual(code, 1)
        self.assertIsNone(verdict)
        self.assertIn("protected list", summary)

    def test_missing_list_file_fails_safe(self):
        mod = load_guard()
        args = [
            "--list", str(self.tmp / "absent.json"),
            "--files", self.write("files.txt", "docs/guide.md\n"),
            "--body", self.write("body.md", "audit-tier: 0\n"),
            "--summary", str(self.tmp / "summary.md"),
        ]
        with contextlib.redirect_stdout(io.StringIO()) as out:
            code = mod.main(args)
        self.assertEqual(code, 1)
        self.assertNotIn('"pass"', out.getvalue())

    def test_empty_or_unreadable_file_list_fails_safe(self):
        mod = load_guard()
        for files in ("", "\n\n"):
            with self.subTest(files=repr(files)):
                code, verdict, summary = self.run_guard(mod, files, "audit-tier: 2\n")
                self.assertEqual(code, 1)
                self.assertIsNone(verdict)
                self.assertIn("file list", summary)

    def test_control_characters_in_a_path_fail_safe(self):
        mod = load_guard()
        code, verdict, summary = self.run_guard(mod, "docs/a\x07b.md\n", "audit-tier: 2\n")
        self.assertEqual(code, 1)
        self.assertIn("readable text", summary)

    def test_unreadable_body_fails_safe(self):
        mod = load_guard()
        args = [
            "--list", self.write("list.json", self.list_text),
            "--files", self.write("files.txt", "docs/guide.md\n"),
            "--body", str(self.tmp / "absent-body.md"),
            "--summary", str(self.tmp / "summary.md"),
        ]
        with contextlib.redirect_stdout(io.StringIO()):
            code = mod.main(args)
        self.assertEqual(code, 1)

    def test_label_failure_cannot_flip_the_verdict(self):
        mod = load_guard()

        def failing(args):
            return (1, "gh: permission denied")

        mod.run_gh = failing
        code, verdict, summary = self.run_guard(
            mod, "REVIEW.md\n", "audit-tier: 2\n", extra=LABEL_ARGS
        )
        self.assertEqual(code, 0)
        self.assertEqual(verdict["verdict"], "pass")
        self.assertEqual(verdict["label"], "failed")
        self.assertIn("permission denied", summary)

    def test_label_flags_absent_means_no_label_call(self):
        mod = load_guard()
        calls = []
        mod.run_gh = lambda args: calls.append(args) or (0, "")
        code, verdict, _ = self.run_guard(mod, "REVIEW.md\n", "audit-tier: 2\n")
        self.assertEqual(code, 0)
        self.assertEqual(verdict["label"], "skipped")
        self.assertEqual(calls, [])

    def test_usage_exit_2_when_a_required_option_is_missing(self):
        mod = load_guard()
        with contextlib.redirect_stdout(io.StringIO()) as out:
            with self.assertRaises(SystemExit) as caught:
                mod.main(["--files", self.write("files.txt", "docs/a.md\n")])
        self.assertEqual(caught.exception.code, 2)
        self.assertIn("usage", out.getvalue().lower())


if __name__ == "__main__":
    unittest.main()
