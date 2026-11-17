"""tools/legacy_verdicts.py (SPEC-20261117-legacy-verdicts): the committed legacy asset map is
exactly what the script regenerates, every row gets a verdict, and a path no rule covers is
refused instead of guessed."""

import importlib.util
import pathlib
import subprocess
import sys
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOL = ROOT / "tools/legacy_verdicts.py"
spec = importlib.util.spec_from_file_location("legacy_verdicts", TOOL)
legacy_verdicts = importlib.util.module_from_spec(spec)
spec.loader.exec_module(legacy_verdicts)


class RegenerateTest(unittest.TestCase):
    def test_rerunning_the_script_reproduces_the_committed_map(self):
        out = subprocess.run(
            [sys.executable, "-I", str(TOOL)], capture_output=True, text=True, check=False
        )
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertIn('"rows": 1107', out.stdout)

    def test_every_row_has_a_verdict_and_sixteen_are_adopted(self):
        raw = legacy_verdicts.MAP.read_text(encoding="utf-8")
        text, counts = legacy_verdicts.regenerate(raw)
        self.assertNotIn('"status": "unchecked"', text)
        self.assertEqual(counts["adopted"], 16)
        self.assertEqual(sum(counts.values()), 1107)

    def test_a_path_without_a_rule_is_refused(self):
        with self.assertRaises(ValueError):
            legacy_verdicts.classify(".claude/unknown/new-file.md")

    def test_template_copies_get_the_verdict_of_their_source(self):
        for path in [".claude/skills/moai-workflow-loop/SKILL.md", ".claude/rules/moai/core/hooks-system.md"]:
            self.assertEqual(
                legacy_verdicts.classify(path),
                legacy_verdicts.classify("internal/template/templates/" + path),
            )
        self.assertEqual(
            legacy_verdicts.classify(".claude/skills/moai-workflow-loop/SKILL.md"),
            ("deferred", "design:edit-time-diagnostics-hook"),
        )

    def test_unknown_arguments_exit_2(self):
        self.assertEqual(legacy_verdicts.main(["legacy_verdicts.py", "--check"]), 2)


if __name__ == "__main__":
    unittest.main()
