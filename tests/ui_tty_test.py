"""Human output under a real terminal (Issue #2803, design #2799 D2 §3 and §11).

Each case runs the built binary with stdout on a pseudo-terminal, so the TTY branches run as a
user would see them: the banner's shape and color depth by `TERM`, `COLORTERM`, `NO_COLOR` and
`COLUMNS`, and the stdout JSON envelope staying byte-identical between a pipe and a terminal.
POSIX only (the `pty` module); the binary is the one `self_audit_run.rust_bin` finds.
"""

import os
import re
import subprocess
import tempfile
import unittest

from tests.self_audit_run import rust_bin

ESC = re.compile(r"\x1b\[[0-9;]*m")
LOGO_CHAR = "█"  # the block letters of the logo


def base_env(home, **extra):
    env = {"PATH": "/usr/bin:/bin", "HOME": home, "LANG": "en_US.UTF-8"}
    env.update(extra)
    return env


def on_tty(args, env, cwd):
    """Run with stdout on a pty (stderr piped); return (stdout text, exit code)."""
    import pty  # POSIX only

    master, slave = pty.openpty()
    try:
        proc = subprocess.Popen(
            [str(rust_bin()), *args],
            stdin=subprocess.DEVNULL,
            stdout=slave,
            stderr=subprocess.PIPE,
            env=env,
            cwd=cwd,
        )
        os.close(slave)
        slave = -1
        chunks = []
        while True:
            try:
                data = os.read(master, 65536)
            except OSError:
                break
            if not data:
                break
            chunks.append(data)
        proc.wait(timeout=60)
        proc.stderr.close()
    finally:
        if slave != -1:
            os.close(slave)
        os.close(master)
    return b"".join(chunks).decode("utf-8", "replace").replace("\r\n", "\n"), proc.returncode


def on_pipe(args, env, cwd):
    p = subprocess.run(
        [str(rust_bin()), *args],
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
        env=env,
        cwd=cwd,
        timeout=60,
    )
    return p.stdout, p.returncode


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class TerminalBanner(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory()
        self.home = self.tmp.name

    def tearDown(self):
        self.tmp.cleanup()

    def help_on_tty(self, **extra):
        out, code = on_tty(["--help"], base_env(self.home, **extra), self.home)
        self.assertEqual(code, 0, out)
        return out

    def test_term_dumb_prints_one_ascii_line_and_no_escape(self):
        out = self.help_on_tty(TERM="dumb", COLUMNS="120")
        self.assertNotIn("\x1b", out)
        self.assertNotIn(LOGO_CHAR, out)
        self.assertTrue(out.startswith("MoAI-ADK "), out[:80])

    def test_a_narrow_terminal_gets_the_line_not_the_logo(self):
        out = self.help_on_tty(TERM="xterm-256color", COLUMNS="61")
        self.assertNotIn(LOGO_CHAR, out)
        self.assertTrue(out.startswith("MoAI-ADK "), out[:80])
        wide = self.help_on_tty(TERM="xterm-256color", COLUMNS="62")
        self.assertIn(LOGO_CHAR, wide)

    def test_256_color_terminals_get_no_24_bit_sequence(self):
        out = self.help_on_tty(TERM="xterm-256color", COLUMNS="120")
        self.assertIn("\x1b[38;5;", out)
        self.assertNotIn("38;2;", out)

    def test_truecolor_terminals_keep_the_24_bit_gradient(self):
        out = self.help_on_tty(TERM="xterm-256color", COLORTERM="truecolor", COLUMNS="120")
        self.assertIn("\x1b[38;2;217;119;87m", out)

    def test_sixteen_color_terminals_get_the_four_bit_accent(self):
        out = self.help_on_tty(TERM="xterm", COLUMNS="120")
        self.assertIn(LOGO_CHAR, out)
        self.assertNotIn("38;5;", out)
        self.assertNotIn("38;2;", out)

    def test_no_color_keeps_the_logo_without_escapes(self):
        out = self.help_on_tty(TERM="xterm-256color", NO_COLOR="1", COLUMNS="120")
        self.assertIn(LOGO_CHAR, out)
        self.assertNotIn("\x1b", out)

    def test_bare_run_follows_the_same_shape(self):
        out, code = on_tty([], base_env(self.home, TERM="dumb", COLUMNS="120"), self.home)
        self.assertEqual(code, 1)
        self.assertEqual(out, ESC.sub("", out))
        self.assertTrue(out.startswith("MoAI-ADK "), out[:80])


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class EnvelopeBytes(unittest.TestCase):
    """The stdout envelope does not depend on whether stdout is a terminal (D2 §1, §11)."""

    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory()
        self.home = self.tmp.name

    def tearDown(self):
        self.tmp.cleanup()

    def assert_same(self, args):
        for extra in ({"TERM": "xterm-256color", "COLORTERM": "truecolor"}, {"TERM": "dumb"}):
            env = base_env(self.home, COLUMNS="120", **extra)
            piped, pcode = on_pipe(args, env, self.home)
            shown, tcode = on_tty(args, env, self.home)
            self.assertEqual(shown, piped, f"{args} {extra}")
            self.assertEqual(tcode, pcode, f"{args} {extra}")
            self.assertTrue(piped.startswith('{"protocol_version":1,'), piped[:80])

    def test_a_read_only_operation(self):
        self.assert_same(["identity", self.home])

    def test_a_usage_error(self):
        self.assert_same(["claim-verify"])


if __name__ == "__main__":
    unittest.main()
