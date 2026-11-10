"""#2331 stage 1: structural checks on the statusline equivalence fixtures.

The goldens themselves are produced by tools/statusline_golden.py against a real
binary; this test only keeps the fixture set self-consistent so stage 2 can trust it.
"""

import json
import os
import pathlib
import subprocess
import sys
import tempfile
import unittest

FIXTURES = pathlib.Path(__file__).resolve().parent / "fixtures" / "statusline"
TOOL = pathlib.Path(__file__).resolve().parent.parent / "tools" / "statusline_golden.py"
sys.path.insert(0, str(TOOL.parent))
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
            # a12 holds invalid UTF-8 on purpose; Go decodes it with U+FFFD, so does this read.
            raw = (FIXTURES / c["stdin"]).read_bytes().decode("utf-8", "replace")
            if c["name"] == "a06-invalid-json":
                with self.assertRaises(json.JSONDecodeError):
                    json.loads(raw)
                continue
            json.loads(raw.replace("{PROJECT}", "/tmp/demo-project"))

    def test_late_recorded_cases_name_their_binary(self):
        # #2350 audit F1: cases recorded after stage 1 carry the Go binary that wrote them.
        for c in self.cases:
            rec = c.get("recorded_with")
            if rec is None:
                continue
            self.assertRegex(rec["version"], r"^\d+\.\d+\.\d+", c["name"])
            self.assertRegex(rec["sha256"], r"^[0-9a-f]{64}$", c["name"])

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


class BinaryVersionTest(unittest.TestCase):
    """#2350: `--check` without `--binary-version` reads the version from `<bin> --version`,
    and only the `🗿 v<version>` token is normalized."""

    def test_last_token_without_a_leading_v(self):
        parse = statusline_golden.parse_version_output
        self.assertEqual(parse(0, "moai-cli 2026.11.0-rc.1\n"), "2026.11.0-rc.1")
        self.assertEqual(parse(0, "moai version v3.2.1\n"), "3.2.1")

    def test_no_version_or_a_failed_run_is_none(self):
        parse = statusline_golden.parse_version_output
        self.assertIsNone(parse(0, "moai-cli unknown\n"))
        self.assertIsNone(parse(0, ""))
        self.assertIsNone(parse(1, "moai-cli 1.2.3\n"))

    def test_only_the_named_version_token_is_replaced(self):
        norm = statusline_golden.normalize_version
        self.assertEqual(norm("🤖 Opus | 🗿 v3.2.1\n", "3.2.1"), "🤖 Opus | 🗿 v<binary-version>\n")
        # b04: the template version stays visible; only the binary side is normalized.
        self.assertEqual(norm("🗿 v2.0.0 -> 🗿 v3.2.1", "3.2.1"), "🗿 v2.0.0 -> 🗿 v<binary-version>")
        self.assertEqual(norm("🗿 v3.2.10", "3.2.1"), "🗿 v<binary-version>0")
        self.assertEqual(norm("v3.2.1", "3.2.1"), "v3.2.1")

    def test_check_without_a_readable_version_exits_2(self):
        with tempfile.TemporaryDirectory() as tmp:
            not_a_binary = pathlib.Path(tmp) / "not-a-binary"
            not_a_binary.write_text("plain text\n", encoding="utf-8")
            proc = subprocess.run(
                [sys.executable, "-I", str(TOOL), "--bin", str(not_a_binary), "--check"],
                capture_output=True, text=True, timeout=60)
        self.assertEqual(proc.returncode, 2, proc.stdout + proc.stderr)
        self.assertIn("binary version unknown", proc.stdout)


@unittest.skipUnless(sys.platform == "darwin", "sandbox-exec is macOS only")
class SandboxProfileTest(unittest.TestCase):
    """#2350 AC-2/AC-5: the harness sandbox denies real-HOME reads, exec of anything but the
    binary under test and git, and network. Each probe has a control that must succeed, so a
    profile that fails everything does not pass."""

    def run_in(self, exe, *args):
        profile = statusline_golden.sandbox_profile(exe)
        return subprocess.run(["/usr/bin/sandbox-exec", "-p", profile, exe, *args],
                              capture_output=True, text=True, cwd="/", timeout=30)

    def test_real_home_is_not_readable(self):
        home = os.path.realpath(os.path.expanduser("~"))
        self.assertEqual(self.run_in("/bin/ls", "/").returncode, 0)
        denied = self.run_in("/bin/ls", home)
        self.assertNotEqual(denied.returncode, 0)
        self.assertIn("Operation not permitted", denied.stderr)

    def test_only_the_binary_and_git_are_executed(self):
        self.assertEqual(self.run_in("/bin/bash", "-c", "echo builtin").returncode, 0)
        denied = self.run_in("/bin/bash", "-c", "/bin/echo external")
        self.assertNotEqual(denied.returncode, 0)
        self.assertIn("Operation not permitted", denied.stderr)

    def test_network_is_denied(self):
        # Port 9 on loopback: without the deny this is `Connection refused`.
        denied = self.run_in("/bin/bash", "-c", "exec 3<>/dev/tcp/127.0.0.1/9")
        self.assertNotEqual(denied.returncode, 0)
        self.assertIn("Operation not permitted", denied.stderr)


if __name__ == "__main__":
    unittest.main()
