"""check_pr_self_audit.py fixtures (#2279 scope 3): the section with all four
measured lines passes; a missing section, any missing line, or a mutation line
that does not report FAILED exits 1; usage exits 2."""

import pathlib
import subprocess
import unittest

TOOL = pathlib.Path(__file__).resolve().parents[1] / "tools/check_pr_self_audit.py"

FULL = """Refs #2200

## Self-Audit

- mutation: claim-verify guard mutated (claim filter removed) -> check FAILED, measured
- clone: check set from a fresh clone, not the /tmp symlink path
- host: api.github.com exercised via fake gh executable, no network
- receipt: receipt digest re-read and matched before PASS recorded

## Notes
"""

MISSING_SECTION = "Refs #2200\nNo audit section here.\n"
MISSING_RECEIPT = FULL.replace("- receipt: receipt digest re-read and matched before PASS recorded\n", "")
UNMEASURED_MUTATION = FULL.replace("-> check FAILED, measured", "-> check passed")


class CheckPrSelfAuditTest(unittest.TestCase):
    def run_tool(self, text=None, stdin=True):
        return subprocess.run(
            ["python3", "-I", str(TOOL)],
            input=text,
            text=True,
            capture_output=True,
            stdin=None if stdin else subprocess.DEVNULL,
        )

    def test_full_section_passes(self):
        out = self.run_tool(FULL)
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertIn("self-audit: pass", out.stdout)

    def test_file_input_passes(self):
        import tempfile
        with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False) as f:
            f.write(FULL)
            path = f.name
        out = subprocess.run(
            ["python3", "-I", str(TOOL), "--file", path],
            text=True, capture_output=True,
        )
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertIn("self-audit: pass", out.stdout)

    def test_missing_section_blocked(self):
        out = self.run_tool(MISSING_SECTION)
        self.assertEqual(out.returncode, 1)
        self.assertIn("## self-audit", out.stdout)

    def test_missing_receipt_line_blocked(self):
        out = self.run_tool(MISSING_RECEIPT)
        self.assertEqual(out.returncode, 1)
        self.assertIn("receipt", out.stdout)

    def test_unmeasured_mutation_blocked(self):
        out = self.run_tool(UNMEASURED_MUTATION)
        self.assertEqual(out.returncode, 1)
        self.assertIn("mutation:FAILED", out.stdout)

    def test_empty_body_blocked(self):
        out = self.run_tool("")
        self.assertEqual(out.returncode, 1)

    def test_usage_exit_2(self):
        out = subprocess.run(
            ["python3", "-I", str(TOOL), "bogus"],
            text=True, capture_output=True, stdin=subprocess.DEVNULL,
        )
        self.assertEqual(out.returncode, 2)


if __name__ == "__main__":
    unittest.main()
