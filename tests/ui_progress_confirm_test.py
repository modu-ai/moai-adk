"""The confirm prompt and static progress lines under a real terminal (Issue #2806, design
#2799 D2 §5·§7·§11, leader decision 6866110172).

`key-remove` without `--yes` asks only when stdin and stderr are both terminals; a pipe keeps
the old non-interactive answer byte for byte. `index-build` prints one line per stage only
when stderr is a terminal, and its stdout envelope does not change with it. POSIX only (the
`pty` module); the binary is the one `self_audit_run.rust_bin` finds.
"""

import json
import os
import re
import select
import subprocess
import tempfile
import time
import unittest

from tests.self_audit_run import rust_bin

PROMPT_END = b"[y/N] "
ESC = re.compile(r"\x1b\[[0-9;]*m")


def env_for(home, **extra):
    env = {"PATH": "/usr/bin:/bin", "HOME": home, "LANG": "C", "TERM": "xterm"}
    env.update(extra)
    return env


def run_pipe(args, env, cwd, stdin=""):
    p = subprocess.run([str(rust_bin()), *args], input=stdin, capture_output=True, text=True,
                       env=env, cwd=cwd, timeout=60)
    return p.stdout, p.stderr, p.returncode


def run_tty(args, env, cwd, answer=None, stdin_tty=True, stdin_text=None):
    """stderr (and stdin unless `stdin_tty` is false) on one pty, stdout on a pipe. When the
    prompt shows, `answer` is typed. Returns (stdout, terminal text, exit code)."""
    import pty  # POSIX only

    master, slave = pty.openpty()
    try:
        if stdin_text is not None:
            stdin = subprocess.PIPE
        else:
            stdin = slave if stdin_tty else subprocess.DEVNULL
        proc = subprocess.Popen([str(rust_bin()), *args], stdin=stdin,
                                stdout=subprocess.PIPE, stderr=slave, env=env, cwd=cwd)
        os.close(slave)
        slave = -1
        if stdin_text is not None:
            proc.stdin.write(stdin_text.encode())
            proc.stdin.close()
        seen = b""
        answered = False
        deadline = time.monotonic() + 60
        while time.monotonic() < deadline:
            ready, _, _ = select.select([master], [], [], 0.2)
            if ready:
                try:
                    data = os.read(master, 65536)
                except OSError:
                    break
                if not data:
                    break
                seen += data
                if answer is not None and not answered and seen.endswith(PROMPT_END):
                    os.write(master, answer)
                    answered = True
            elif proc.poll() is not None:
                break
        out = proc.stdout.read().decode()
        proc.wait(timeout=60)
    finally:
        if slave != -1:
            os.close(slave)
        os.close(master)
    text = seen.decode("utf-8", "replace").replace("\r\n", "\n")
    return out, ESC.sub("", text), proc.returncode


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class KeyRemovePrompt(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory()
        self.home = os.path.join(self.tmp.name, "home")
        self.cwd = os.path.join(self.tmp.name, "work")
        os.makedirs(self.home)
        os.makedirs(self.cwd)
        self.env = env_for(self.home)

    def tearDown(self):
        self.tmp.cleanup()

    def set_key(self):
        out, err, code = run_pipe(["key-set", "glm", "--shared"], self.env, self.cwd,
                                  stdin="not-a-real-key\n")
        self.assertEqual(code, 0, err)

    def test_a_pipe_never_asks_and_keeps_the_old_answer(self):
        self.set_key()
        out, err, code = run_pipe(["key-remove", "glm", "--shared"], self.env, self.cwd)
        self.assertEqual(code, 0, err)
        self.assertNotIn("[y/N]", err)
        self.assertEqual(json.loads(out)["result"]["removed"], False)
        # stdin piped, stderr a terminal: still no question (both must be terminals).
        out2, term, _ = run_tty(["key-remove", "glm", "--shared"], self.env, self.cwd,
                                stdin_tty=False)
        self.assertNotIn("[y/N]", term)
        self.assertEqual(out2, out, "the stdout envelope is the same bytes")

    def test_no_on_the_terminal_removes_nothing_and_names_the_flag(self):
        self.set_key()
        piped, _, _ = run_pipe(["key-remove", "glm", "--shared"], self.env, self.cwd)
        out, term, code = run_tty(["key-remove", "glm", "--shared"], self.env, self.cwd,
                                  answer=b"n\n")
        self.assertEqual(code, 0, term)
        self.assertIn("key-remove deletes 1 file", term)
        self.assertIn("glm key, shared scope:", term)
        self.assertIn("! Remove it? [y/N] ", term)
        self.assertIn("next: run with --yes to remove it without asking", term)
        self.assertEqual(out, piped, "a refusal answers with the preview envelope as before")
        self.assertNotIn("\x1b[", out)

    def test_yes_on_the_terminal_equals_the_yes_flag(self):
        self.set_key()
        out, term, code = run_tty(["key-remove", "glm", "--shared"], self.env, self.cwd,
                                  answer=b"y\n")
        self.assertEqual(code, 0, term)
        self.assertEqual(json.loads(out)["result"]["removed"], True)
        self.set_key()
        flag, _, _ = run_pipe(["key-remove", "glm", "--shared", "--yes"], self.env, self.cwd)
        self.assertEqual(out, flag, "an interactive yes answers like --yes, byte for byte")

    def test_nothing_to_remove_asks_nothing(self):
        out, term, code = run_tty(["key-remove", "glm", "--shared"], self.env, self.cwd,
                                  answer=b"y\n")
        self.assertEqual(code, 0, term)
        self.assertNotIn("[y/N]", term)
        self.assertEqual(json.loads(out)["result"]["exists"], False)


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class IndexBuildProgress(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory()
        self.home = os.path.join(self.tmp.name, "home")
        os.makedirs(self.home)
        self.env = env_for(self.home)

    def tearDown(self):
        self.tmp.cleanup()

    def project(self, name):
        root = os.path.join(self.tmp.name, name)
        os.makedirs(root)
        with open(os.path.join(root, "a.md"), "w") as f:
            f.write("# a\n")
        git = {"HOME": self.home, "PATH": "/usr/bin:/bin", "GIT_CONFIG_NOSYSTEM": "1",
               "GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@example.invalid",
               "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@example.invalid",
               "GIT_AUTHOR_DATE": "2026-01-01T00:00:00Z",
               "GIT_COMMITTER_DATE": "2026-01-01T00:00:00Z"}
        for args in (["init", "-q"], ["add", "a.md"], ["commit", "-qm", "a"]):
            subprocess.run(["git", *args], cwd=root, env=git, check=True, capture_output=True)
        return root

    @staticmethod
    def stable(text):
        doc = json.loads(text)
        doc.get("result", {}).pop("built_at", None)
        return doc

    def test_stages_show_on_a_terminal_and_never_on_a_pipe(self):
        a = self.project("a")
        b = self.project("b")
        req = '{"protocol_version":1}'
        piped, err, code = run_pipe(["index-build", a], self.env, a, stdin=req)
        self.assertEqual(code, 0, err)
        self.assertEqual(err, "", "no progress byte on a pipe")
        # stdin carries the request, so only stderr is the terminal here.
        out, term, code = run_tty(["index-build", b], self.env, b, stdin_text=req)
        self.assertEqual(code, 0, term)
        lines = [line for line in term.splitlines() if line.startswith("->")]
        self.assertEqual(lines, ["-> scanning tracked files (1/3)",
                                 "-> building the views (2/3)",
                                 "-> publishing the generation (3/3)"])
        self.assertNotIn("\r", term.replace("\n", ""))
        self.assertEqual(self.stable(out).keys(), self.stable(piped).keys())

    def test_moai_output_quiet_drops_the_stages_and_plain_drops_color_and_symbols(self):
        req = '{"protocol_version":1}'
        utf8 = dict(self.env, LANG="en_US.UTF-8", LC_ALL="en_US.UTF-8")
        stages = ("scanning tracked files (1/3)", "building the views (2/3)",
                  "publishing the generation (3/3)")

        def shown(name, **extra):
            root = self.project(name)
            out, term, code = run_tty(["index-build", root], dict(utf8, **extra), root,
                                      stdin_text=req)
            self.assertEqual(code, 0, term)
            return out, term

        _, normal = shown("n")
        self.assertIn("→ scanning tracked files (1/3)", ESC.sub("", normal))
        _, plain = shown("p", MOAI_OUTPUT="plain")
        self.assertEqual([line for line in plain.splitlines() if line.startswith("->")],
                         [f"-> {s}" for s in stages])
        self.assertNotIn("\x1b", plain)
        self.assertNotIn("→", plain)
        out, quiet = shown("q", MOAI_OUTPUT="quiet")
        self.assertFalse(any(s in quiet for s in stages), quiet)
        self.assertTrue(json.loads(out)["ok"])


if __name__ == "__main__":
    unittest.main()
