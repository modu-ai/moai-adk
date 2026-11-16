"""`key-set PROVIDER --prompt` under a real terminal (card C-54 S1).

The value is typed at a hidden prompt: the terminal's echo is off while it is read and back on
at every exit (a stored value, a refusal, Ctrl-C, SIGTERM). Neither the terminal nor stdout ever
shows the value. Without a terminal `--prompt` is refused and nothing is read or stored.

Every case uses a temporary HOME, a temporary working directory and a synthetic value made at
run time; the real home and any real key are never read. POSIX only (the `pty` module); the
binary is the one `self_audit_run.rust_bin` finds.
"""

import fcntl
import json
import os
import select
import signal
import subprocess
import tempfile
import termios
import time
import unittest

from tests.self_audit_run import rust_bin

PROMPT_END = b": "
LFLAG = 3  # index of the local flags in what termios.tcgetattr returns


def env_for(home):
    return {"PATH": "/usr/bin:/bin", "HOME": home, "LANG": "C", "TERM": "xterm"}


def run_pipe(args, env, cwd, stdin=None):
    p = subprocess.run([str(rust_bin()), *args], input=stdin, capture_output=True, text=True,
                       env=env, cwd=cwd, timeout=60,
                       stdin=subprocess.DEVNULL if stdin is None else None)
    return p.stdout, p.stderr, p.returncode


def echo_on(fd):
    return bool(termios.tcgetattr(fd)[LFLAG] & termios.ECHO)


def _own_terminal():
    """The child leads a new session with the pty (its stdin) as controlling terminal, so a
    typed Ctrl-C reaches it as SIGINT and nothing reaches the test runner."""
    os.setsid()
    fcntl.ioctl(0, termios.TIOCSCTTY, 0)


def run_prompt(args, env, cwd, typed=None, at_prompt=None):
    """stdin and stderr on one pty, stdout on a pipe. When the prompt shows, `typed` is written
    in small pieces, or `at_prompt(master, proc)` runs. Returns (stdout, terminal bytes, exit
    code, echo while asked, echo after exit)."""
    import pty  # POSIX only

    master, slave = pty.openpty()
    asked = False
    echo_asked = None
    pending = b""
    try:
        proc = subprocess.Popen([str(rust_bin()), *args], stdin=slave, stdout=subprocess.PIPE,
                                stderr=slave, env=env, cwd=cwd, preexec_fn=_own_terminal)
        seen = b""
        deadline = time.monotonic() + 60
        while time.monotonic() < deadline:
            # The typed bytes go in only as fast as the child takes them: a write the
            # terminal has no room for waits its turn instead of blocking this loop.
            ready, writable, _ = select.select([master], [master] if pending else [], [], 0.2)
            if ready:
                try:
                    data = os.read(master, 65536)
                except OSError:
                    break
                if not data:
                    break
                seen += data
                if not asked and seen.endswith(PROMPT_END):
                    asked = True
                    echo_asked = echo_on(master)
                    if at_prompt is not None:
                        at_prompt(master, proc)
                    if typed is not None:
                        pending = typed
            if pending and writable:
                pending = pending[os.write(master, pending[:256]):]
            if not ready and not pending and proc.poll() is not None:
                break
        if proc.poll() is None:
            # Past the deadline a child still waiting on the terminal fails the test instead
            # of hanging it.
            proc.kill()
        out = proc.stdout.read().decode()
        proc.stdout.close()
        proc.wait(timeout=60)
        echo_after = echo_on(master)
    finally:
        os.close(slave)
        os.close(master)
    return out, seen, proc.returncode, echo_asked, echo_after


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class KeySetPrompt(unittest.TestCase):
    ARGS = ["key-set", "typesafe", "--shared", "--prompt"]

    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory()
        self.home = os.path.join(self.tmp.name, "home")
        self.cwd = os.path.join(self.tmp.name, "work")
        os.makedirs(self.home)
        os.makedirs(self.cwd)
        self.env = env_for(self.home)
        # Test data, made at run time: not a real key and not shaped like one.
        self.value = f"moai-prompt-test-{os.getpid()}-{time.monotonic_ns()}"

    def tearDown(self):
        self.tmp.cleanup()

    def vault(self):
        return os.path.join(self.home, ".moai-cli", "vault")

    def entry(self):
        return os.path.join(self.vault(), "_shared", "typesafe.key")

    def status(self):
        out, err, code = run_pipe(["key-status", "typesafe", "--shared"], self.env, self.cwd)
        self.assertEqual(code, 0, err)
        return json.loads(out)["result"]["entry"]

    def assert_hidden(self, value, out, term):
        self.assertNotIn(value.encode(), term, "the typed value reached the terminal")
        self.assertNotIn(value, out, "the typed value reached stdout")

    def test_the_typed_value_is_stored_and_never_shown(self):
        out, term, code, echo_asked, echo_after = run_prompt(
            self.ARGS, self.env, self.cwd, typed=self.value.encode() + b"\n")
        self.assertEqual(code, 0, term)
        self.assert_hidden(self.value, out, term)
        self.assertIs(echo_asked, False, "echo is off while the prompt waits")
        self.assertTrue(echo_after, "echo is back on after the value is stored")
        self.assertIn(b"typesafe", term, "the prompt names the provider")
        result = json.loads(out)["result"]
        # The answer of a piped `key-set`, field for field: no value, length or fingerprint.
        self.assertEqual(sorted(result), ["encrypted", "folder", "path", "provider", "scope",
                                          "value_shown", "written"])
        self.assertEqual(result["value_shown"], False)
        self.assertEqual(result["written"], True)
        # The entry decrypts, and no stored file holds the value in clear.
        self.assertEqual(self.status(), "present")
        with open(self.entry(), "rb") as f:
            self.assertNotIn(self.value.encode(), f.read())

    def test_the_prompt_answers_like_the_pipe(self):
        piped, err, code = run_pipe(["key-set", "typesafe", "--shared"], self.env, self.cwd,
                                    stdin=self.value + "\n")
        self.assertEqual(code, 0, err)
        out, term, code, _, _ = run_prompt(self.ARGS, self.env, self.cwd,
                                           typed=self.value.encode() + b"\n")
        self.assertEqual(code, 0, term)
        self.assertEqual(out, piped, "the stdout envelope is the same bytes")

    def test_without_a_terminal_the_prompt_is_refused_and_nothing_is_stored(self):
        for stdin in (self.value + "\n", None):
            out, err, code = run_pipe(self.ARGS, self.env, self.cwd, stdin=stdin)
            self.assertEqual(code, 1, err)
            self.assertNotIn(self.value, out + err)
            self.assertEqual(
                json.loads(out)["error"]["message"],
                "--prompt reads the value from a terminal with echo off, and stdin is not a "
                "terminal; nothing was read or stored (run it in your own terminal window, or "
                "drop --prompt and pipe the value in)")
            self.assertFalse(os.path.exists(self.vault()))

    def test_ctrl_c_leaves_echo_on_and_stores_nothing(self):
        def interrupt(master, proc):
            os.write(master, self.value.encode()[:8] + b"\x03")

        out, term, code, echo_asked, echo_after = run_prompt(
            self.ARGS, self.env, self.cwd, at_prompt=interrupt)
        self.assertEqual(code, -signal.SIGINT, term)
        self.assertIs(echo_asked, False)
        self.assertTrue(echo_after, "Ctrl-C left the terminal without echo")
        self.assertNotIn(self.value.encode()[:8], term)
        self.assertFalse(os.path.exists(self.vault()))

    def test_sigterm_and_sighup_leave_echo_on_and_store_nothing(self):
        for sig in (signal.SIGTERM, signal.SIGHUP):
            out, term, code, echo_asked, echo_after = run_prompt(
                self.ARGS, self.env, self.cwd,
                at_prompt=lambda master, proc, sig=sig: proc.send_signal(sig))
            self.assertEqual(code, -sig, term)
            self.assertIs(echo_asked, False)
            self.assertTrue(echo_after, f"signal {sig} left the terminal without echo")
            self.assertFalse(os.path.exists(self.vault()))

    def test_empty_and_control_character_input_is_refused_with_echo_back_on(self):
        # Enter alone, spaces, end of input (Ctrl-D), a control character inside the value,
        # and an arrow key (an escape sequence).
        for typed in (b"\n", b"   \n", b"\x04", b"ab\x01cd\n", b"ab\x1b[Dcd\n"):
            out, term, code, _, echo_after = run_prompt(self.ARGS, self.env, self.cwd,
                                                        typed=typed)
            self.assertEqual(code, 1, (typed, term))
            self.assertEqual(json.loads(out)["ok"], False, typed)
            self.assertTrue(echo_after, typed)
            self.assertNotIn(b"abcd", term)
            self.assertFalse(os.path.exists(self.vault()), typed)

    def test_backspace_edits_the_hidden_line(self):
        # A wrong character and its backspace: stored as if never typed. A stray control
        # byte would be refused (the case above), so a stored entry shows the edit was read.
        out, term, code, _, _ = run_prompt(self.ARGS, self.env, self.cwd,
                                           typed=b"x\x7f" + self.value.encode() + b"\n")
        self.assertEqual(code, 0, term)
        self.assert_hidden(self.value, out, term)
        self.assertEqual(self.status(), "present")

    def test_the_size_limit_holds_at_the_prompt(self):
        # 65,535 bytes is the largest value stored (a 65,583-byte entry); 65,536 is refused.
        longest = b"k" * 65_535
        out, term, code, _, echo_after = run_prompt(self.ARGS, self.env, self.cwd,
                                                    typed=longest + b"\n")
        self.assertEqual(code, 0, term[-200:])
        self.assertNotIn(b"kkkk", term)
        self.assertNotIn("65535", out.replace(",", ""))
        self.assertEqual(os.path.getsize(self.entry()), 65_583)
        self.assertEqual(self.status(), "present")
        os.remove(self.entry())
        out, term, code, _, echo_after = run_prompt(self.ARGS, self.env, self.cwd,
                                                    typed=longest + b"k\n")
        self.assertEqual(code, 1, term[-200:])
        self.assertNotIn(b"kkkk", term)
        self.assertTrue(echo_after)
        self.assertFalse(os.path.exists(self.entry()))

    def test_a_bad_provider_name_is_refused_before_anything_is_asked(self):
        out, term, code, echo_asked, echo_after = run_prompt(
            ["key-set", "Not_A_Name", "--shared", "--prompt"], self.env, self.cwd,
            typed=self.value.encode() + b"\n")
        self.assertEqual(code, 1, term)
        self.assertIsNone(echo_asked, "no prompt for a name that cannot be stored")
        self.assertTrue(echo_after)
        self.assertFalse(os.path.exists(self.vault()))


if __name__ == "__main__":
    unittest.main()
