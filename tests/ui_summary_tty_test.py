"""The plan/apply summary on a terminal's stderr (Issue #2805, design #2799 D2 §1, §6, §11).

stderr goes to a pseudo-terminal and stdout to a pipe, so the summary branch runs as a user sees
it while the envelope is captured byte for byte. POSIX only (the `pty` module); the binary is the
one `self_audit_run.rust_bin` finds.
"""

import os
import re
import subprocess
import tempfile
import unittest

from tests.self_audit_run import ROOT, rust_bin

ESC = re.compile(r"\x1b\[[0-9;]*m")
PACKAGE = str(ROOT / "skills" / "moai")


def env_for(home, **extra):
    env = {"PATH": "/usr/bin:/bin", "HOME": home, "LANG": "en_US.UTF-8",
           "TERM": "xterm-256color", "COLUMNS": "100"}
    env.update(extra)
    return env


def run(args, env, cwd, stderr_tty):
    """(stdout, stderr, exit): stdout always piped; stderr on a pty when `stderr_tty`."""
    if not stderr_tty:
        p = subprocess.run([str(rust_bin()), *args], stdin=subprocess.DEVNULL,
                           capture_output=True, text=True, env=env, cwd=cwd, timeout=60)
        return p.stdout, p.stderr, p.returncode
    import pty  # POSIX only

    master, slave = pty.openpty()
    try:
        proc = subprocess.Popen([str(rust_bin()), *args], stdin=subprocess.DEVNULL,
                                stdout=subprocess.PIPE, stderr=slave, env=env, cwd=cwd)
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
        out = proc.stdout.read().decode()
        proc.stdout.close()
        proc.wait(timeout=60)
    finally:
        if slave != -1:
            os.close(slave)
        os.close(master)
    err = b"".join(chunks).decode("utf-8", "replace").replace("\r\n", "\n")
    return out, err, proc.returncode


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class PlanSummary(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory()
        self.home = self.tmp.name

    def tearDown(self):
        self.tmp.cleanup()

    def routing(self, stderr_tty, **extra):
        return run(["home-routing-plan", self.home, PACKAGE], env_for(self.home, **extra),
                   self.home, stderr_tty)

    def test_a_terminal_gets_the_table_and_the_envelope_stays_the_same(self):
        piped_out, piped_err, pcode = self.routing(False)
        tty_out, tty_err, tcode = self.routing(True)
        self.assertEqual(tty_out, piped_out)
        self.assertEqual(tcode, pcode)
        self.assertEqual(piped_err, "", "no summary when stderr is a pipe")
        plain = ESC.sub("", tty_err)
        self.assertIn("home-routing-plan", plain)
        self.assertRegex(plain, r"ACTION\s+HOST")
        self.assertRegex(plain, r"create\s+claude")
        self.assertRegex(plain, r"create\s+codex")
        self.assertIn("writes nothing", plain)
        self.assertIn("next:", plain)
        self.assertIn("home-routing-apply", plain)
        self.assertIn("\x1b[", tty_err, "color on a 256-color terminal")
        self.assertNotIn("38;5;", tty_err)
        self.assertNotIn("38;2;", tty_err)

    def test_quiet_turns_the_summary_off(self):
        _, err, _ = self.routing(True, MOAI_OUTPUT="quiet")
        self.assertEqual(err, "")

    def test_plain_and_dumb_terminals_get_ascii_without_escapes(self):
        for extra in ({"MOAI_OUTPUT": "plain"}, {"TERM": "dumb"}, {"NO_COLOR": "1", "LANG": "C"}):
            _, err, _ = self.routing(True, **extra)
            self.assertTrue(err, extra)
            self.assertNotIn("\x1b", err, extra)
            self.assertTrue(err.isascii(), f"{extra}: {err!r}")

    def test_narrow_terminals_keep_every_line_inside_the_width(self):
        _, err, _ = self.routing(True, COLUMNS="60")
        lines = ESC.sub("", err).splitlines()
        self.assertTrue(any("--plan-digest" in line for line in lines), lines)
        for line in lines:
            if line.strip().startswith("--plan-digest"):
                continue  # the digest is copied whole, never cut
            self.assertLessEqual(len(line), 60, line)

    def test_a_result_without_rows_is_one_line(self):
        out, err, code = run(["home-state-plan", "--home", self.home, "--now", "2026-11-13T19:00:00Z"],
                             env_for(self.home), self.home, True)
        self.assertEqual(code, 0, out)
        lines = ESC.sub("", err).splitlines()
        self.assertEqual(len(lines), 4, lines)  # title, totals, next, digest
        self.assertIn("writes nothing", lines[1])
        self.assertIn("home-state-apply", lines[2])
        self.assertTrue(lines[3].strip().startswith("--plan-digest sha256:"), lines[3])

    def test_control_characters_in_a_folder_name_never_reach_the_terminal(self):
        # #2818 audit B1 (CWE-150): a HOME folder named with a window-title OSC and a
        # clear-screen CSI; the summary shows them as text, the envelope escapes them as JSON.
        evil = os.path.join(self.home, "h\x1b]0;PWNED-TITLE\x07\x1b[2Jz")
        os.makedirs(evil)
        args = ["home-state-plan", "--home", evil, "--now", "2026-11-13T19:00:00Z"]
        piped, _, pcode = run(args, env_for(evil), self.home, False)
        out, err, code = run(args, env_for(evil), self.home, True)
        self.assertEqual(code, 0, out)
        self.assertEqual(out, piped)
        self.assertIn("\\u001b]0;PWNED-TITLE", out, "the envelope escapes it as JSON")
        self.assertNotIn("\x1b]", err)
        self.assertNotIn("\x1b[2J", err)
        self.assertNotIn("\x07", err)
        self.assertIn("\\x1b]0;PWNED-TITLE\\x07\\x1b[2Jz", ESC.sub("", err))

    def test_other_operations_print_no_summary(self):
        out, err, code = run(["identity", self.home], env_for(self.home), self.home, True)
        self.assertEqual(code, 0, out)
        self.assertEqual(err, "")


if __name__ == "__main__":
    unittest.main()
