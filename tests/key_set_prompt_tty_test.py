"""`key-set PROVIDER --prompt` under a real terminal (card C-54 S1).

The value is typed at a hidden prompt: the terminal's echo is off while it is read, and its
modes (echo and line mode) are back at every exit (a stored value, a refusal, Ctrl-C, SIGTERM,
SIGHUP, SIGQUIT). Neither the terminal nor stdout ever shows the value, and nothing typed or
pasted is left in the terminal for the shell. One line is one value: anything after the line
end refuses the whole input. Without a terminal on stdin and on stderr `--prompt` is refused
and nothing is read or stored.

Every case uses a temporary HOME, a temporary working directory and a synthetic value made at
run time; the real home and any real key are never read. POSIX only (the `pty` module); the
binary is the one `self_audit_run.rust_bin` finds.
"""

import json
import os
import select
import signal
import subprocess
import sys
import tempfile
import termios
import time
import types
import unittest

from tests.self_audit_run import rust_bin

PROMPT_END = b": "
LFLAG = 3  # index of the local flags in what termios.tcgetattr returns
CC = 6  # index of the control characters


def env_for(home):
    return {"PATH": "/usr/bin:/bin", "HOME": home, "LANG": "C", "TERM": "xterm"}


def run_pipe(args, env, cwd, stdin=None):
    p = subprocess.run([str(rust_bin()), *args], input=stdin, capture_output=True, text=True,
                       env=env, cwd=cwd, timeout=60,
                       stdin=subprocess.DEVNULL if stdin is None else None)
    return p.stdout, p.stderr, p.returncode


def echo_on(fd):
    return bool(termios.tcgetattr(fd)[LFLAG] & termios.ECHO)


def left_in_queue(slave):
    """What the terminal still holds as input once the child is gone: what a shell would read
    next. Read without waiting, with the line mode off so a part of a line counts too."""
    attrs = termios.tcgetattr(slave)
    attrs[LFLAG] &= ~(termios.ICANON | termios.ECHO)
    attrs[CC][termios.VMIN] = 0
    attrs[CC][termios.VTIME] = 0
    termios.tcsetattr(slave, termios.TCSANOW, attrs)
    return os.read(slave, 65536)


# The stand-in for the shell a person runs `moai` from: it leads the session that owns the
# terminal and outlives the command, so the terminal can be looked at after the command ended
# (a terminal whose session leader exits is taken away from everyone). It starts the command
# in its own foreground group, reports the command's pid and then its exit code on
# `status_fd`, and stays until a byte arrives on `hold_fd`.
SHELL = r"""
import fcntl, os, signal, subprocess, sys, termios
status_fd, hold_fd = int(sys.argv[1]), int(sys.argv[2])
ignored = [int(s) for s in sys.argv[3].split(",") if s]
os.setsid()
fcntl.ioctl(0, termios.TIOCSCTTY, 0)
# A typed Ctrl-C goes to the whole foreground group: caught here (a handler, which the
# command does not inherit), so only the command ends.
for sig in (signal.SIGINT, signal.SIGQUIT, signal.SIGTERM, signal.SIGHUP, signal.SIGTSTP):
    signal.signal(sig, lambda *_: None)
def as_started():
    for sig in (signal.SIGINT, signal.SIGQUIT, signal.SIGTERM, signal.SIGHUP, signal.SIGTSTP):
        signal.signal(sig, signal.SIG_IGN if sig in ignored else signal.SIG_DFL)
child = subprocess.Popen(sys.argv[4:], preexec_fn=as_started)
os.write(status_fd, b"%d\n" % child.pid)
os.write(status_fd, b"%d\n" % child.wait())
os.read(hold_fd, 1)
"""


class Command:
    """The command the stand-in shell started: its pid and exit code as they are reported."""

    def __init__(self, status_fd):
        self.status_fd = status_fd
        self.text = b""

    def _lines(self, wait):
        ready, _, _ = select.select([self.status_fd], [], [], wait)
        if ready:
            self.text += os.read(self.status_fd, 4096)
        return self.text.split(b"\n")[:-1]

    def pid(self):
        deadline = time.monotonic() + 30
        while time.monotonic() < deadline:
            lines = self._lines(0.1)
            if lines:
                return int(lines[0])
        raise AssertionError("the command did not start")

    def poll(self):
        lines = self._lines(0)
        return int(lines[1]) if len(lines) > 1 else None

    def send_signal(self, sig):
        try:
            os.kill(self.pid(), sig)
        except ProcessLookupError:
            pass  # already gone; the exit code tells


def run_prompt(args, env, cwd, typed=None, at_prompt=None, stderr_tty=True, ignored=()):
    """stdin and stderr on one pty (stderr on nothing when `stderr_tty` is false), stdout on a
    pipe, under a stand-in shell (`SHELL`). When the prompt shows, `at_prompt(master, command)`
    runs and `typed` is written in small pieces; `typed` is bytes, or a list of `(seconds to
    wait first, bytes)`. The command starts with the signals in `ignored` ignored, as under
    `nohup`.

    Returns a namespace: `out` (stdout), `term` (terminal bytes), `code`, `echo_asked` (echo
    while the prompt waited, None when it never showed), `modes_back` (echo and line mode on
    after the exit), `left` (input the terminal still holds after the exit)."""
    import pty  # POSIX only

    chunks = [(0, typed)] if isinstance(typed, bytes) else list(typed or [])
    master, slave = pty.openpty()
    status_r, status_w = os.pipe()
    hold_r, hold_w = os.pipe()
    asked = False
    echo_asked = None
    pending = b""
    due = None
    try:
        shell = subprocess.Popen(
            [sys.executable, "-c", SHELL, str(status_w), str(hold_r),
             ",".join(str(int(s)) for s in ignored), str(rust_bin()), *args],
            stdin=slave, stdout=subprocess.PIPE,
            stderr=slave if stderr_tty else subprocess.DEVNULL,
            env=env, cwd=cwd, pass_fds=(status_w, hold_r))
        os.close(status_w)
        os.close(hold_r)
        status_w = hold_r = -1
        command = Command(status_r)
        seen = b""
        deadline = time.monotonic() + 60
        while time.monotonic() < deadline:
            # The typed bytes go in only as fast as the terminal takes them: a write it has
            # no room for waits its turn instead of blocking this loop.
            ready, writable, _ = select.select([master], [master] if pending else [], [], 0.1)
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
                        at_prompt(master, command)
                    due = time.monotonic()
            if pending and writable:
                pending = pending[os.write(master, pending[:256]):]
                if not pending:
                    due = time.monotonic()
            if not pending and chunks and due is not None:
                # The next piece goes in after its wait, whether or not the command is still
                # there: what a paste sends late still reaches the terminal.
                if time.monotonic() >= due + chunks[0][0]:
                    pending = chunks.pop(0)[1]
                continue
            if not ready and not pending and command.poll() is not None:
                if not chunks or not asked:
                    break
        if command.poll() is None:
            # Past the deadline a command still waiting on the terminal fails the test
            # instead of hanging it.
            os.kill(command.pid(), signal.SIGKILL)
            while command.poll() is None:
                time.sleep(0.05)
        code = command.poll()
        # The stand-in shell still holds the terminal: this is what the person's shell finds.
        flags = termios.tcgetattr(slave)[LFLAG]
        modes_back = bool(flags & termios.ECHO) and bool(flags & termios.ICANON)
        left = left_in_queue(slave)
        os.write(hold_w, b"x")
        out = shell.stdout.read().decode()
        shell.stdout.close()
        shell.wait(timeout=60)
    finally:
        for fd in (slave, master, status_r, hold_w, status_w, hold_r):
            if fd != -1:
                os.close(fd)
    return types.SimpleNamespace(out=out, term=seen, code=code, echo_asked=echo_asked,
                                 modes_back=modes_back, left=left)


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

    def prompt(self, **how):
        return run_prompt(how.pop("args", self.ARGS), self.env, self.cwd, **how)

    def vault(self):
        return os.path.join(self.home, ".moai-cli", "vault")

    def entry(self):
        return os.path.join(self.vault(), "_shared", "typesafe.key")

    def status(self):
        out, err, code = run_pipe(["key-status", "typesafe", "--shared"], self.env, self.cwd)
        self.assertEqual(code, 0, err)
        return json.loads(out)["result"]["entry"]

    def assert_hidden(self, value, r):
        self.assertNotIn(value.encode(), r.term, "the typed value reached the terminal")
        self.assertNotIn(value, r.out, "the typed value reached stdout")

    def assert_terminal_as_before(self, r, why=""):
        self.assertTrue(r.modes_back, f"echo and line mode are not back {why}")
        self.assertEqual(r.left, b"", f"input was left in the terminal for the shell {why}")

    def test_the_typed_value_is_stored_and_never_shown(self):
        r = self.prompt(typed=self.value.encode() + b"\n")
        self.assertEqual(r.code, 0, r.term)
        self.assert_hidden(self.value, r)
        self.assertIs(r.echo_asked, False, "echo is off while the prompt waits")
        self.assert_terminal_as_before(r)
        self.assertIn(b"typesafe", r.term, "the prompt names the provider")
        result = json.loads(r.out)["result"]
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
        # A terminal that ends a line with CR LF sends one more line end: still one line.
        r = self.prompt(typed=self.value.encode() + b"\r\n")
        self.assertEqual(r.code, 0, r.term)
        self.assertEqual(r.out, piped, "the stdout envelope is the same bytes")
        self.assert_terminal_as_before(r)

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

    def test_without_a_terminal_to_ask_on_the_prompt_is_refused(self):
        # stdin is a terminal, stderr is not: the question could not be seen.
        r = self.prompt(typed=self.value.encode() + b"\n", stderr_tty=False)
        self.assertEqual(r.code, 1, r.out)
        self.assertEqual(
            json.loads(r.out)["error"]["message"],
            "--prompt shows its question on stderr, and stderr is not a terminal; nothing was "
            "read or stored (run it in your own terminal window, or drop --prompt and pipe the "
            "value in)")
        self.assertIsNone(r.echo_asked)
        self.assertFalse(os.path.exists(self.vault()))

    def test_ctrl_c_leaves_the_terminal_as_it_was_and_stores_nothing(self):
        def interrupt(master, proc):
            os.write(master, self.value.encode()[:8] + b"\x03")

        r = self.prompt(at_prompt=interrupt)
        self.assertEqual(r.code, -signal.SIGINT, r.term)
        self.assertIs(r.echo_asked, False)
        self.assert_terminal_as_before(r, "after Ctrl-C")
        self.assertNotIn(self.value.encode()[:8], r.term)
        self.assertFalse(os.path.exists(self.vault()))

    def test_ending_signals_leave_the_terminal_as_it_was_and_store_nothing(self):
        # SIGQUIT ends the process without a core file (it could hold typed bytes): the exit
        # status is the conventional 128 + signal instead of the signal itself.
        for sig, code in ((signal.SIGTERM, -signal.SIGTERM), (signal.SIGHUP, -signal.SIGHUP),
                          (signal.SIGQUIT, 128 + signal.SIGQUIT)):
            def end(master, proc, sig=sig):
                os.write(master, self.value.encode()[:8])
                proc.send_signal(sig)

            r = self.prompt(at_prompt=end)
            self.assertEqual(r.code, code, (sig, r.term))
            self.assertIs(r.echo_asked, False)
            self.assert_terminal_as_before(r, f"after signal {sig}")
            self.assertNotIn(self.value.encode()[:8], r.term)
            self.assertFalse(os.path.exists(self.vault()))

    def test_an_ignored_hangup_stays_ignored_and_a_stop_signal_does_not_stop_the_prompt(self):
        # Under `nohup` SIGHUP is ignored and stays so; SIGTSTP is ignored while echo is off,
        # since a stopped process could not put the terminal's modes back.
        def poke(master, proc):
            proc.send_signal(signal.SIGHUP)
            proc.send_signal(signal.SIGTSTP)

        r = self.prompt(at_prompt=poke, typed=[(0.3, self.value.encode() + b"\n")],
                        ignored=(signal.SIGHUP,))
        self.assertEqual(r.code, 0, r.term)
        self.assert_terminal_as_before(r)
        self.assertEqual(self.status(), "present")

    def test_empty_and_control_character_input_is_refused_with_the_terminal_as_it_was(self):
        # Enter alone, spaces, a control character inside the value, and an arrow key (an
        # escape sequence).
        for typed in (b"\n", b"   \n", b"ab\x01cd\n", b"ab\x1b[Dcd\n"):
            r = self.prompt(typed=typed)
            self.assertEqual(r.code, 1, (typed, r.term))
            self.assertEqual(json.loads(r.out)["ok"], False, typed)
            self.assert_terminal_as_before(r, repr(typed))
            self.assertNotIn(b"abcd", r.term)
            self.assertFalse(os.path.exists(self.vault()), typed)

    def test_ctrl_d_cancels_and_stores_nothing(self):
        # On an empty line and after a part of a value alike: Ctrl-D never ends a value.
        for typed in (b"\x04", self.value.encode() + b"\x04"):
            r = self.prompt(typed=typed)
            self.assertEqual(r.code, 1, (typed, r.term))
            self.assertIn("cancelled", json.loads(r.out)["error"]["message"])
            self.assert_hidden(self.value, r)
            self.assert_terminal_as_before(r, repr(typed))
            self.assertFalse(os.path.exists(self.vault()), typed)

    def test_a_second_line_refuses_the_whole_input(self):
        # A paste of two lines, in one piece and with the second line a moment later: nothing
        # is stored, neither line is shown, and nothing is left for the shell to run.
        for typed in (b"AAAfirst\nBBBsecond\n",
                      [(0, b"AAAfirst\n"), (0.3, b"BBBsecond\n")]):
            r = self.prompt(typed=typed)
            self.assertEqual(r.code, 1, (typed, r.term))
            self.assertIn("more than one line", json.loads(r.out)["error"]["message"])
            self.assertNotIn(b"AAAfirst", r.term, typed)
            self.assertNotIn(b"BBBsecond", r.term, typed)
            self.assertNotIn("AAAfirst", r.out)
            self.assertNotIn("BBBsecond", r.out)
            self.assert_terminal_as_before(r, repr(typed))
            self.assertFalse(os.path.exists(self.vault()), typed)

    def test_backspace_edits_the_hidden_line(self):
        # A wrong character and its backspace: stored as if never typed. A stray control
        # byte would be refused (the case above), so a stored entry shows the edit was read.
        r = self.prompt(typed=b"x\x7f" + self.value.encode() + b"\n")
        self.assertEqual(r.code, 0, r.term)
        self.assert_hidden(self.value, r)
        self.assertEqual(self.status(), "present")

    def test_the_size_limit_holds_at_the_prompt(self):
        # 65,535 bytes is the largest value stored (a 65,583-byte entry); 65,536 is refused.
        longest = b"k" * 65_535
        r = self.prompt(typed=longest + b"\n")
        self.assertEqual(r.code, 0, r.term[-200:])
        self.assertNotIn(b"kkkk", r.term)
        self.assertNotIn("65535", r.out.replace(",", ""))
        self.assertEqual(os.path.getsize(self.entry()), 65_583)
        self.assertEqual(self.status(), "present")
        os.remove(self.entry())
        r = self.prompt(typed=longest + b"k\n")
        self.assertEqual(r.code, 1, r.term[-200:])
        self.assertNotIn(b"kkkk", r.term)
        self.assert_terminal_as_before(r)
        self.assertFalse(os.path.exists(self.entry()))

    def test_a_bad_provider_name_is_refused_before_anything_is_asked(self):
        r = self.prompt(args=["key-set", "Not_A_Name", "--shared", "--prompt"],
                        typed=self.value.encode() + b"\n")
        self.assertEqual(r.code, 1, r.term)
        self.assertIsNone(r.echo_asked, "no prompt for a name that cannot be stored")
        self.assertTrue(r.modes_back)
        self.assertFalse(os.path.exists(self.vault()))


if __name__ == "__main__":
    unittest.main()
