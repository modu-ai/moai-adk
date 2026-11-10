"""#2331 stage 1: structural checks on the statusline equivalence fixtures.

The goldens themselves are produced by tools/statusline_golden.py against a real
binary; this test only keeps the fixture set self-consistent so stage 2 can trust it.
"""

import json
import pathlib
import sys
import tempfile
import unittest

FIXTURES = pathlib.Path(__file__).resolve().parent / "fixtures" / "statusline"
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent / "tools"))
import statusline_golden  # noqa: E402


class StatuslineFixturesTest(unittest.TestCase):
    def setUp(self):
        self.cases = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))["cases"]

    def test_case_names_are_unique_and_classified(self):
        names = [c["name"] for c in self.cases]
        self.assertEqual(len(names), len(set(names)))
        for c in self.cases:
            self.assertIn(c["class"], ("a", "b"), c["name"])
            self.assertTrue(c["name"].startswith(c["class"]), c["name"])
            self.assertTrue(c.get("covers"), c["name"])

    def test_every_stdin_exists_and_parses_except_the_invalid_case(self):
        for c in self.cases:
            raw = (FIXTURES / c["stdin"]).read_text(encoding="utf-8")
            if c["name"] == "a06-invalid-json":
                with self.assertRaises(json.JSONDecodeError):
                    json.loads(raw)
                continue
            json.loads(raw.replace("{PROJECT}", "/tmp/demo-project"))

    def test_every_case_has_a_plain_and_an_ansi_golden(self):
        for c in self.cases:
            plain = FIXTURES / "golden" / f"{c['name']}.plain.txt"
            ansi = FIXTURES / "golden" / f"{c['name']}.ansi.txt"
            self.assertTrue(plain.is_file(), plain)
            self.assertTrue(ansi.is_file(), ansi)
            self.assertNotIn("\x1b[", plain.read_text(encoding="utf-8"), plain)

    def test_every_case_has_an_ansi256_golden_without_truecolor(self):
        # #2350: the ANSI256 path (COLORTERM unset) is recorded as its own variant.
        for c in self.cases:
            golden = FIXTURES / "golden" / f"{c['name']}.ansi256.txt"
            self.assertTrue(golden.is_file(), golden)
            self.assertNotIn("38;2;", golden.read_text(encoding="utf-8"), golden)

    def test_no_stray_goldens(self):
        names = {c["name"] for c in self.cases}
        for g in (FIXTURES / "golden").iterdir():
            self.assertIn(g.name.rsplit(".", 2)[0], names, g)


class BinaryVersionDetectionTest(unittest.TestCase):
    """#2350: `--check` without `--binary-version` reads the version from `<bin> --version`."""

    def detect(self, script):
        with tempfile.TemporaryDirectory() as tmp:
            exe = pathlib.Path(tmp) / "fake-bin"
            exe.write_text("#!/bin/sh\n" + script + "\n", encoding="utf-8")
            exe.chmod(0o755)
            return statusline_golden.detect_binary_version(exe)

    def test_last_token_without_a_leading_v(self):
        self.assertEqual(self.detect('echo "moai-cli 2026.11.0-rc.1"'), "2026.11.0-rc.1")
        self.assertEqual(self.detect('echo "moai version v3.2.1"'), "3.2.1")

    def test_no_version_or_a_failed_run_is_none(self):
        self.assertIsNone(self.detect('echo "moai-cli unknown"'))
        self.assertIsNone(self.detect("exit 0"))
        self.assertIsNone(self.detect('echo "moai-cli 1.2.3"; exit 1'))


if __name__ == "__main__":
    unittest.main()
