"""The stderr error frame under a real terminal (card #2804, design #2799 D2 §4 and §11).

Each case runs the built binary with stderr on a pseudo-terminal (stdout piped), so the frame
is drawn as a user would see it, and again with both streams piped, as a lane or a script runs
it. The pipe keeps today's bytes; the stdout JSON envelope is byte-identical in both. POSIX
only (the `pty` module); the binary is the one `self_audit_run.rust_bin` finds.
"""

import os
import re
import subprocess
import tempfile
import unittest

from tests.self_audit_run import rust_bin

ESC = re.compile(r"\x1b\[[0-9;]*m")


def base_env(home, **extra):
    env = {"PATH": "/usr/bin:/bin", "HOME": home, "LANG": "en_US.UTF-8", "TERM": "xterm-256color"}
    env.update(extra)
    return env


def piped(args, env, cwd):
    p = subprocess.run(
        [str(rust_bin()), *args],
        stdin=subprocess.DEVNULL,
        capture_output=True,
        env=env,
        cwd=cwd,
        timeout=60,
    )
    return p.returncode, p.stdout, p.stderr.decode("utf-8", "replace")


def stderr_on_tty(args, env, cwd):
    """Run with stderr on a pty (stdout piped); return (exit, stdout bytes, stderr text)."""
    import pty  # POSIX only

    master, slave = pty.openpty()
    try:
        proc = subprocess.Popen(
            [str(rust_bin()), *args],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=slave,
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
        out = proc.stdout.read()
        proc.wait(timeout=60)
        proc.stdout.close()
    finally:
        if slave != -1:
            os.close(slave)
        os.close(master)
    return proc.returncode, out, b"".join(chunks).decode("utf-8", "replace").replace("\r\n", "\n")


@unittest.skipUnless(os.name == "posix", "pty is POSIX only")
class ErrorFrameTty(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.home = self.tmp.name

    def tearDown(self):
        self.tmp.cleanup()

    def env(self, **extra):
        return base_env(self.home, **extra)

    def test_an_operation_error_is_framed_on_a_terminal_and_unchanged_on_a_pipe(self):
        args = ["claim-verify", "."]  # empty stdin: `invalid typed JSON input`
        code_p, out_p, err_p = piped(args, self.env(), self.home)
        self.assertEqual(code_p, 1)
        self.assertEqual(err_p, "invalid typed JSON input\n")  # today's bytes
        code_t, out_t, err_t = stderr_on_tty(args, self.env(), self.home)
        self.assertEqual(code_t, 1)
        self.assertEqual(out_t, out_p, "the envelope is byte-identical")
        self.assertEqual(err_t, "\x1b[31merror:\x1b[0m invalid typed JSON input\n")

    def test_plain_modes_draw_the_same_frame_without_escapes(self):
        args = ["claim-verify", "."]
        for extra in ({"NO_COLOR": "1"}, {"TERM": "dumb"}, {"COLUMNS": "60", "NO_COLOR": "1"}):
            _, _, err = stderr_on_tty(args, self.env(**extra), self.home)
            self.assertEqual(err, "error: invalid typed JSON input\n", extra)

    def test_a_missing_argument_keeps_its_name_and_ends_with_the_next_step(self):
        args = ["project-plan"]
        code_p, out_p, err_p = piped(args, self.env(), self.home)
        self.assertEqual(code_p, 2)
        self.assertEqual(
            err_p, "error: the following required arguments were not provided:\n  <PROJECT>\n"
        )
        code_t, out_t, err_t = stderr_on_tty(args, self.env(), self.home)
        self.assertEqual(code_t, 2)
        self.assertEqual(out_t, out_p, "the envelope is byte-identical")
        plain = ESC.sub("", err_t)
        lines = plain.splitlines()
        self.assertEqual(lines[0], "error: the following required arguments were not provided:")
        self.assertEqual(lines[1], "  <PROJECT>")
        self.assertTrue(lines[-1].startswith("  next: run `"), plain)
        self.assertIn(" project-plan --help` to see its arguments", lines[-1])
        self.assertIn("\x1b[36mnext:\x1b[0m", err_t)

    def test_a_key_command_keeps_its_fixed_text_on_a_terminal(self):
        # #2393: a key command's stray argument may be a key; nothing of it is echoed.
        args = ["key-status", "--sk-ant-pasted"]
        code_t, out_t, err_t = stderr_on_tty(args, self.env(), self.home)
        self.assertEqual(code_t, 2)
        self.assertEqual(err_t, "unexpected argument; see --help\n")
        self.assertNotIn(b"sk-ant", out_t)

    def test_the_color_flag_comes_before_the_environment(self):
        # #2799 ③ (D2 order): `--color=never|always` first, then NO_COLOR, TERM=dumb, the TTY.
        args = ["claim-verify", "."]
        _, out_auto, _ = piped(args, self.env(), self.home)
        for flag in (["--color", "never"], ["--color=never"]):
            code, out, err = stderr_on_tty([*flag, *args], self.env(), self.home)
            self.assertEqual(code, 1, flag)
            self.assertEqual(out, out_auto, "the envelope is byte-identical")
            self.assertEqual(err, "error: invalid typed JSON input\n", flag)
        for extra in ({"NO_COLOR": "1"}, {"TERM": "dumb"}):
            _, _, err = stderr_on_tty(["--color", "always", *args], self.env(**extra), self.home)
            self.assertEqual(err, "\x1b[31merror:\x1b[0m invalid typed JSON input\n", extra)
        _, _, err = stderr_on_tty(["--color", "auto", *args], self.env(NO_COLOR="1"), self.home)
        self.assertEqual(err, "error: invalid typed JSON input\n")

    def test_the_color_flag_keeps_the_pipe_bytes(self):
        args = ["claim-verify", "."]
        code_p, out_p, err_p = piped(args, self.env(), self.home)
        for flag in ("always", "never", "auto"):
            code, out, err = piped(["--color", flag, *args], self.env(), self.home)
            self.assertEqual((code, out, err), (code_p, out_p, err_p), flag)

    def test_a_flag_before_the_command_keeps_the_next_step_of_that_command(self):
        code, _, err = stderr_on_tty(["--color", "never", "project-plan"], self.env(), self.home)
        self.assertEqual(code, 2)
        self.assertIn(" project-plan --help` to see its arguments", err.splitlines()[-1])

    def test_help_all_reads_the_line_without_a_leading_color_flag(self):
        # Audit #2952 B2: `help --all` is answered before parsing from position 1.
        want = piped(["help", "--all"], self.env(), self.home)
        self.assertEqual(want[0], 0)
        for flag in (["--color", "never"], ["--color=always"]):
            self.assertEqual(piped([*flag, "help", "--all"], self.env(), self.home), want, flag)

    def test_a_hook_word_after_a_leading_color_flag_is_still_a_hook(self):
        # Audit #2952 B3: hook mode is judged from position 1 (`hook-receipt`), so a leading
        # `--color` must not turn a hook call into an ordinary usage error.
        want = piped(["hook-receipt"], self.env(), self.home)
        got = piped(["--color", "never", "hook-receipt"], self.env(), self.home)
        self.assertEqual(got, want)

    def test_an_unknown_color_value_is_a_usage_error(self):
        code, out, err = piped(["--color", "bogus", "claim-verify", "."], self.env(), self.home)
        self.assertEqual(code, 2)
        self.assertEqual(out, b"")
        self.assertIn("bogus", err)
        for value in ("auto", "always", "never"):
            self.assertIn(value, err)


if __name__ == "__main__":
    unittest.main()
