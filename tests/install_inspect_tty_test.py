"""`install-inspect`'s PATH note under a real terminal (#2867 R5a, PR #2881 audit B1).

When the first `moai` on PATH is not this generation's, the note is printed once on stderr
only when stderr is a terminal; a lane or a script that pipes stderr reads no extra line. The
stdout JSON envelope is byte-identical either way and carries the note in
`path_first_moai.note`. POSIX only (the `pty` module); the binary is the one
`self_audit_run.rust_bin` finds.
"""

import os
import shutil
import subprocess
import tempfile
import unittest

from tests.self_audit_run import rust_bin

NOTE = (
    "moai-cli: hooks are generated with moai-cli until R5; put a Rust moai first on PATH "
    "to let them use moai (nothing was changed)\n"
)


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
class InstallInspectTty(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        root = os.path.realpath(self.tmp.name)
        self.home = os.path.join(root, "home")
        self.proj = os.path.join(root, "proj")
        os.makedirs(self.home)
        os.makedirs(self.proj)
        self.go = os.path.join(root, "go-bin")
        os.makedirs(self.go)
        shim = os.path.join(self.go, "moai")
        with open(shim, "w") as f:
            f.write("#!/bin/sh\necho 'moai 3.1.5 (go)'\n")
        os.chmod(shim, 0o755)
        self.rust = os.path.join(root, "rust-bin")
        os.makedirs(self.rust)
        target = os.path.join(self.rust, "moai")
        try:
            os.link(rust_bin(), target)
        except OSError:
            shutil.copy2(rust_bin(), target)

    def tearDown(self):
        self.tmp.cleanup()

    def env(self, first_dir):
        return {
            "PATH": f"{first_dir}:/usr/bin:/bin",
            "HOME": self.home,
            "LANG": "en_US.UTF-8",
            "TERM": "xterm-256color",
        }

    def run_both(self, first_dir):
        args = ["install-inspect", "--json"]
        piped_run = piped(args, self.env(first_dir), self.proj)
        tty_run = stderr_on_tty(args, self.env(first_dir), self.proj)
        return piped_run, tty_run

    def test_a_go_first_moai_prints_the_note_once_on_a_terminal_only(self):
        (code_p, out_p, err_p), (code_t, out_t, err_t) = self.run_both(self.go)
        self.assertEqual((code_p, code_t), (0, 0), err_p + err_t)
        self.assertEqual(err_p, "", "a piped stderr carries no note")
        self.assertEqual(err_t, NOTE, "the terminal shows the note exactly once")
        self.assertEqual(out_t, out_p, "the envelope is byte-identical")
        self.assertIn(b'"generation":"go"', out_p)
        self.assertIn(NOTE.removeprefix("moai-cli: ").rstrip("\n").encode(), out_p)

    def test_a_rust_first_moai_prints_nothing_on_a_terminal(self):
        (code_p, out_p, err_p), (code_t, out_t, err_t) = self.run_both(self.rust)
        self.assertEqual((code_p, code_t), (0, 0), err_p + err_t)
        self.assertEqual(err_p, "")
        self.assertEqual(err_t, "", "no note when the first moai is this generation's")
        self.assertEqual(out_t, out_p, "the envelope is byte-identical")
        self.assertIn(b'"generation":"rust"', out_p)


if __name__ == "__main__":
    unittest.main()
