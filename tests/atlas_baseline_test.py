#! /usr/bin/env python3
"""C10a corpus/baseline consistency tests (Issue #2346, AC-2/AC-3).

The Rust side validates the fixture format; this side validates that the
committed baseline.json still matches a fresh rg-only run — thresholds are
pinned to the baseline, so loosening one without re-running fails here.
"""

import subprocess
import sys
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
RUNNER = ROOT / "tools" / "atlas_baseline.py"


class AtlasBaselineConsistency(unittest.TestCase):
    def test_committed_baseline_matches_a_fresh_run(self):
        out = subprocess.run(
            [sys.executable, "-I", str(RUNNER)],
            capture_output=True,
            text=True,
            cwd=ROOT,
        )
        self.assertEqual(out.returncode, 0, out.stdout + out.stderr)
        self.assertIn("thresholds pinned", out.stdout)


if __name__ == "__main__":
    unittest.main()
