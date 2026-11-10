#! /usr/bin/env python3
"""C10a corpus/baseline consistency tests (Issue #2346, AC-2/AC-3).

The Rust side validates the fixture format; this side validates that the
committed baseline.json still matches a fresh rg-only run — thresholds are
pinned to the baseline, so loosening one without re-running fails here.
"""

import json
import os
import shlex
import shutil
import subprocess
import sys
import tempfile
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

    def test_partial_search_error_fails_check_and_build_without_writing_baseline(self):
        real_rg = shutil.which("rg")
        self.assertIsNotNone(real_rg, "rg is required; missing rg must not skip this test")
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            fixtures = root / "tests" / "fixtures" / "atlas"
            fixtures.mkdir(parents=True)
            (root / "tools").mkdir()
            runner = root / "tools" / RUNNER.name
            shutil.copy2(RUNNER, runner)
            (root / "src").mkdir()
            (root / "src" / "answer.rs").write_text(
                "fn searchable_answer() {}\n", encoding="utf-8"
            )
            (fixtures / "corpus.json").write_text(
                json.dumps(
                    {
                        "tasks": [
                            {
                                "id": "search-error",
                                "kind": "retrieval",
                                "pattern": "searchable_answer",
                                "ground_truth": {"files": ["src/answer.rs"]},
                            }
                        ]
                    }
                ),
                encoding="utf-8",
            )
            initial = subprocess.run(
                [sys.executable, "-I", str(runner), "--build"],
                cwd=root, capture_output=True, text=True,
            )
            self.assertEqual(initial.returncode, 0, initial.stdout + initial.stderr)
            baseline = fixtures / "baseline.json"
            original = baseline.read_bytes()
            fake_bin = root / "bin"
            fake_bin.mkdir()
            driver = fake_bin / "rg_error.py"
            diagnostic = "fixture source could not be read"
            driver.write_text(
                "import subprocess, sys\n"
                f"out = subprocess.run([{real_rg!r}, *sys.argv[1:]], capture_output=True)\n"
                "sys.stdout.buffer.write(out.stdout)\n"
                "sys.stderr.buffer.write(out.stderr)\n"
                "if '--version' in sys.argv[1:]:\n"
                "    raise SystemExit(out.returncode)\n"
                f"sys.stderr.write({diagnostic!r} + '\\n')\n"
                "raise SystemExit(2)\n",
                encoding="utf-8",
            )
            wrapper = fake_bin / "rg"
            wrapper.write_text(
                "#!/bin/sh\nexec "
                + shlex.quote(sys.executable) + " -I "
                + shlex.quote(str(driver)) + ' "$@"\n',
                encoding="utf-8",
            )
            wrapper.chmod(0o755)
            environment = dict(os.environ, PATH=str(fake_bin) + os.pathsep + os.environ.get("PATH", ""))
            for arguments in ([], ["--build"]):
                with self.subTest(arguments=arguments):
                    out = subprocess.run(
                        [sys.executable, "-I", str(runner), *arguments],
                        cwd=root, env=environment, capture_output=True, text=True,
                    )
                    self.assertNotEqual(out.returncode, 0, out.stdout + out.stderr)
                    self.assertIn(diagnostic, out.stderr)
                    self.assertEqual(baseline.read_bytes(), original)


if __name__ == "__main__":
    unittest.main()
