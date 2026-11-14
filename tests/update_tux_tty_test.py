"""`moai update` for a person at a terminal (Issue #2894, leader decision D3-U).

stdin, stdout and stderr all on one pseudo-terminal, the way a person runs it: screens, no JSON
envelope, one `[y/N]` question. A pipe keeps the envelope; `--json` keeps it on a terminal too.
Every run uses a temporary HOME with the binary in the release layout
`~/.moai-cli/versions/<v>/bin/moai-cli` and the repository's `skills/moai` and `agents/` beside
it, so the real HOME is never touched. POSIX only (the `pty` module).
"""

import json
import os
import select
import shutil
import subprocess
import tempfile
import time
import unittest

from tests.self_audit_run import ROOT, rust_bin

PROMPT_END = b"[y/N] "


def run_pty(args, env, cwd, answer=None, before_answer=None, stdin_tty=True):
    """Run with stdout and stderr (and stdin unless `stdin_tty` is false) on one pty. When the
    prompt shows, `before_answer()` runs, then `answer` is typed. Returns (text, exit)."""
    import pty

    master, slave = pty.openpty()
    try:
        stdin = slave if stdin_tty else subprocess.DEVNULL
        proc = subprocess.Popen([str(env["_BIN"]), *args], stdin=stdin, stdout=slave,
                                stderr=slave, cwd=cwd,
                                env={k: v for k, v in env.items() if k != "_BIN"})
        os.close(slave)
        slave = -1
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
                    if before_answer:
                        before_answer()
                    os.write(master, answer)
                    answered = True
            elif proc.poll() is not None:
                break
        proc.wait(timeout=60)
    finally:
        if slave != -1:
            os.close(slave)
        os.close(master)
    return seen.decode("utf-8", "replace").replace("\r\n", "\n"), proc.returncode


def tree(path):
    """Every file under `path` with its bytes, to prove nothing was written."""
    out = {}
    for root, _, files in os.walk(path):
        for name in files:
            p = os.path.join(root, name)
            with open(p, "rb") as f:
                out[os.path.relpath(p, path)] = f.read()
    return out


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class UpdateScreens(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory()
        base = os.path.realpath(self.tmp.name)
        self.home = os.path.join(base, "home")
        self.work = os.path.join(base, "work")
        os.makedirs(self.work)
        folder = os.path.join(self.home, ".moai-cli", "versions", "2026.11.0-rc.1")
        os.makedirs(os.path.join(folder, "bin"))
        self.bin = os.path.join(folder, "bin", "moai-cli")
        shutil.copy2(rust_bin(), self.bin)
        shutil.copytree(ROOT / "skills" / "moai", os.path.join(folder, "skills", "moai"))
        shutil.copytree(ROOT / "agents", os.path.join(folder, "agents"))
        self.package = os.path.join(folder, "skills", "moai")
        self.env = {"PATH": "/usr/bin:/bin", "HOME": self.home, "LANG": "en_US.UTF-8",
                    "TERM": "xterm-256color", "NO_COLOR": "1", "MOAI_LANG": "en",
                    "COLUMNS": "80", "_BIN": self.bin}

    def tearDown(self):
        self.tmp.cleanup()

    def pipe(self, *args):
        p = subprocess.run([self.bin, *args], capture_output=True, text=True, cwd=self.work,
                           env={k: v for k, v in self.env.items() if k != "_BIN"}, timeout=60)
        return p.stdout, p.stderr, p.returncode

    def skill_file(self):
        return os.path.join(self.home, ".claude", "skills", "moai", "SKILL.md")

    def test_a_pipe_keeps_the_envelope_and_json_is_the_same_bytes(self):
        """U-1: the pipe path is the envelope; `--json` on a pipe is the same bytes."""
        out, _, code = self.pipe("update")
        self.assertEqual(code, 0)
        env = json.loads(out)
        self.assertEqual(env["operation"], "update.run")
        self.assertEqual(env["result"]["mode"], "preview")
        out2, _, _ = self.pipe("update", "--json")
        self.assertEqual(out, out2)

    def test_a_terminal_shows_screens_and_no_envelope(self):
        """U-2: on a terminal the person sees the screen and the question, never JSON."""
        before = tree(self.home)
        text, code = run_pty(["update"], self.env, self.work, answer=b"n\n")
        self.assertEqual(code, 0, text)
        self.assertIn("MoAI update", text)
        self.assertIn("Checking", text)
        self.assertIn("[y/N]", text)
        self.assertIn("Nothing was written.", text)
        self.assertNotIn('"protocol_version"', text)
        self.assertNotIn("sha256:", text)
        self.assertEqual(tree(self.home), before, "N must write nothing")

    def test_json_on_a_terminal_prints_the_envelope(self):
        """U-2: `--json` keeps the envelope on a terminal."""
        text, code = run_pty(["update", "--json"], self.env, self.work, answer=b"n\n")
        self.assertEqual(code, 0, text)
        self.assertIn('"protocol_version"', text)
        self.assertIn('"operation":"update.run"', text)

    def test_yes_applies_and_ends_with_the_project_step(self):
        """U-4 and U-9: `y` writes, the end screen says what changed and sends each project to
        its own session's /moai update; a second run is all up to date with the same step."""
        text, code = run_pty(["update"], self.env, self.work, answer=b"y\n")
        self.assertEqual(code, 0, text)
        self.assertTrue(os.path.isfile(self.skill_file()), text)
        flat = text.replace("\n  ", " ").replace(" ", " ")
        self.assertIn("Done", text)
        self.assertIn("run /moai update in a session there", flat)
        self.assertNotIn('"protocol_version"', text)
        again, code = run_pty(["update"], self.env, self.work)
        self.assertEqual(code, 0, again)
        self.assertIn("Everything is up to date", again)
        self.assertNotIn("[y/N]", again)
        self.assertIn("$moai update", again.replace("\n  ", " ").replace(" ", " "))

    def test_end_of_input_writes_nothing(self):
        """U-4: end of input (Ctrl-D) at the question is a no."""
        before = tree(self.home)
        text, code = run_pty(["update"], self.env, self.work, answer=b"\x04")
        self.assertEqual(code, 0, text)
        self.assertEqual(tree(self.home), before)
        self.assertFalse(os.path.exists(self.skill_file()))

    def test_a_plan_that_changes_before_the_write_is_refused(self):
        """U-4: the package changes after the preview, then `y`: nothing is written and the
        new preview is shown."""
        def change_package():
            with open(os.path.join(self.package, "workflows", "update.md"), "a") as f:
                f.write("\nchanged after the preview\n")

        text, code = run_pty(["update"], self.env, self.work, answer=b"y\n",
                             before_answer=change_package)
        self.assertEqual(code, 0, text)
        self.assertIn("The plan changed after the preview", text)
        self.assertFalse(os.path.exists(self.skill_file()), text)

    def test_stdin_that_is_not_a_terminal_is_never_asked(self):
        """U-5: stdout on a terminal but stdin not: the screen and the confirming command,
        no question, no wait."""
        started = time.monotonic()
        text, code = run_pty(["update"], self.env, self.work, stdin_tty=False)
        self.assertEqual(code, 0, text)
        self.assertNotIn("[y/N]", text)
        self.assertIn("moai update --confirm sha256:", text)
        self.assertLess(time.monotonic() - started, 30)
        self.assertFalse(os.path.exists(self.skill_file()))

    def test_an_error_on_a_terminal_is_the_frame_not_json(self):
        """D3-U: a failed run on a terminal shows the error frame only; a pipe still gets the
        failure envelope."""
        shutil.rmtree(self.package)
        text, code = run_pty(["update"], self.env, self.work)
        self.assertEqual(code, 1, text)
        self.assertIn("no skill package installed", text)
        self.assertNotIn('"protocol_version"', text)
        out, _, code = self.pipe("update")
        self.assertEqual(code, 1)
        self.assertEqual(json.loads(out)["ok"], False)

    def test_korean_on_a_korean_terminal(self):
        """U item 8: the profile/terminal language picks Korean sentences."""
        env = dict(self.env, MOAI_LANG="ko")
        text, code = run_pty(["update"], env, self.work, answer=b"n\n")
        self.assertEqual(code, 0, text)
        self.assertIn("검사", text)
        self.assertIn("쓰지 않았어요.", text)


if __name__ == "__main__":
    unittest.main()
