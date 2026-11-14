"""Help is English on a terminal whatever the language settings say (Issue #2825 R1, user
decision 6878265192: the CLI's human output is English only; the user's language belongs to the
LLM session).

The pipe bytes are pinned by `tests/help_layout.rs`; here a terminal, with no language setting
and with `MOAI_LANG=ko` or a Korean locale, must give those same bytes — for the root help,
`help --all`, a subcommand's `-h` and `--help`, and the stderr of a usage error. POSIX only.
"""

import os
import re
import shutil
import subprocess
import tempfile
import unittest

from tests.self_audit_run import rust_bin
from tests.ui_korean_tty_test import KOREAN, env_for

VERSION = re.compile(r"^MoAI-ADK \S+ — ", re.M)
KOREAN_SETTINGS = (KOREAN, {"LANG": "ko_KR.UTF-8"}, {"LC_ALL": "ko_KR.UTF-8"})


def moai_in(home):
    """The binary copied as `<home>/bin/moai`, so help names `moai` as a person types it."""
    path = os.path.join(home, "bin", "moai")
    if not os.path.exists(path):
        os.makedirs(os.path.dirname(path), exist_ok=True)
        shutil.copy2(rust_bin(), path)
    return path


def run_help(args, env, stdout_tty=True, stderr_tty=False):
    """(stdout, stderr, exit) with stdout and/or stderr on a pty (one pty for both when both)."""
    import pty  # POSIX only

    master, slave = pty.openpty()
    try:
        proc = subprocess.Popen([moai_in(env["HOME"]), *args], stdin=subprocess.DEVNULL,
                                stdout=slave if stdout_tty else subprocess.PIPE,
                                stderr=slave if stderr_tty else subprocess.PIPE,
                                env=env, cwd=env["HOME"])
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
        out = "" if stdout_tty else proc.stdout.read().decode()
        err = "" if stderr_tty else proc.stderr.read().decode()
        proc.wait(timeout=60)
    finally:
        if slave != -1:
            os.close(slave)
        os.close(master)
    # The pty turns each "\n" into "\r\n"; under a full output queue macOS has been seen to send
    # one "\r" twice ("\r\r\n" in a 15 KB `help --all`). Help text carries no "\r" of its own.
    term = b"".join(chunks).decode("utf-8", "replace").replace("\r", "")
    if stdout_tty:
        out = term
    else:
        err = term
    return out, err, proc.returncode


def normal(text):
    """`text` from its help title on — the startup banner a terminal gets before the root help
    is the banner's own (#2803), not help — with the version as `<version>`."""
    m = VERSION.search(text)
    return VERSION.sub("MoAI-ADK <version> — ", text[m.start():] if m else text)


def pipe(args, env):
    p = subprocess.run([moai_in(env["HOME"]), *args], stdin=subprocess.DEVNULL,
                       capture_output=True, text=True, env=env, cwd=env["HOME"], timeout=60)
    return p.stdout, p.stderr, p.returncode


def hangul(text):
    return any("가" <= c <= "힣" for c in text)


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class EnglishHelp(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory(dir="/tmp")
        self.home = self.tmp.name

    def tearDown(self):
        self.tmp.cleanup()

    def help(self, args, **extra):
        out, err, code = run_help(list(args), env_for(self.home, **extra))
        self.assertEqual(code, 0, err)
        return normal(out)

    def test_a_terminal_gets_the_pipe_bytes_in_every_language_setting(self):
        for args in (["--help"], ["help", "--all"], ["project-detect", "--help"],
                     ["key-set", "-h"], ["cc", "--help"], ["update", "-h"]):
            piped, _, code = pipe(args, env_for(self.home))
            self.assertEqual(code, 0, args)
            self.assertEqual(self.help(args), normal(piped), args)
            for extra in KOREAN_SETTINGS:
                self.assertEqual(self.help(args, **extra), normal(piped), (args, extra))

    def test_no_command_help_carries_korean_under_a_korean_setting(self):
        listing, _, _ = pipe(["help", "--all"], env_for(self.home))
        words = sorted({line.split()[0] for line in listing.splitlines()
                        if line.startswith("  ") and not line.startswith("   ")
                        and not line.split()[0].startswith("-")})
        self.assertGreater(len(words), 100)
        korean = []
        for word in words:
            english, _, _ = pipe([word, "--help"], env_for(self.home))
            text = self.help([word, "--help"], **KOREAN)
            # A command's own English about may quote a Korean phrase; it is in the pipe too.
            if hangul(text) and text != normal(english):
                korean.append(word)
        self.assertEqual(korean, [])

    def test_a_usage_error_is_english_on_a_terminal_too(self):
        for extra in KOREAN_SETTINGS:
            for args in (["--bogus"], ["bogus"], ["project-detect"]):
                _, piped_err, piped_code = pipe(args, env_for(self.home))
                _, err, code = run_help(args, env_for(self.home, **extra), stdout_tty=True,
                                        stderr_tty=False)
                self.assertEqual((code, err), (piped_code, piped_err), (args, extra))


if __name__ == "__main__":
    unittest.main()
