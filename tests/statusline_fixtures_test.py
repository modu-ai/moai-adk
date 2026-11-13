"""Behavioral checks for the native Claude statusline harness and its sandbox.

Fixture input/output coverage is exercised by the real-binary golden integration test.
"""

import os
import pathlib
import subprocess
import sys
import tempfile
import unittest

TOOL = pathlib.Path(__file__).resolve().parent.parent / "tools" / "statusline_golden.py"
sys.path.insert(0, str(TOOL.parent))
import statusline_golden  # noqa: E402


class BinaryVersionTest(unittest.TestCase):
    """#2350: `--check` without `--binary-version` reads the version from `<bin> --version`,
    and only the `🗿 v<version>` token is normalized."""

    def test_last_token_without_a_leading_v(self):
        parse = statusline_golden.parse_version_output
        self.assertEqual(parse(0, "moai-cli 2026.11.0-rc.1\n"), "2026.11.0-rc.1")
        self.assertEqual(parse(0, "moai version v3.2.1\n"), "3.2.1")
        # #2790: the Rust build's line ends with its generation mark.
        self.assertEqual(parse(0, "moai-cli 2026.11.0-rc.1 (rust)\n"), "2026.11.0-rc.1")
        self.assertEqual(parse(0, "moai-cli 2026.11.0-rc.1+local.3 (rust)\n"), "2026.11.0-rc.1+local.3")
        self.assertIsNone(parse(0, "(rust)\n"))

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


class DeveloperEnvTest(unittest.TestCase):
    """#2699: the sandbox allows the developer folder `xcrun --find git` names in this process,
    so every child run under the sandbox must resolve git with the same `DEVELOPER_DIR`.
    Without it the macOS golden cases with git (b01, b02, b08-b10) differed on the tip after
    `.cargo/config.toml` started setting the variable. Runs on every platform."""

    def test_the_variable_is_carried_only_when_set(self):
        from unittest import mock
        with mock.patch.dict(os.environ, {"DEVELOPER_DIR": "/Library/Developer/CommandLineTools"}):
            self.assertEqual(statusline_golden.developer_env(),
                             {"DEVELOPER_DIR": "/Library/Developer/CommandLineTools"})
        with mock.patch.dict(os.environ, {}, clear=True):
            self.assertEqual(statusline_golden.developer_env(), {})

    def test_both_sandboxed_child_environments_carry_it(self):
        source = TOOL.read_text()
        # The case run and the `--version` probe are the two children built from scratch.
        self.assertEqual(source.count('"PATH": "/usr/bin:/bin"'), 2, "two child environments")
        self.assertEqual(source.count("**developer_env()"), 2, "each carries DEVELOPER_DIR")


if __name__ == "__main__":
    unittest.main()
