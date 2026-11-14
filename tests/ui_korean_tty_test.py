"""Korean human output on a terminal (CLI design S6b, Issue #2825; design
`docs/design-review/cli-s6-korean-2026-11-14.md` §6, leader decision 6870716572).

The S4 summary, the S5 progress lines and the `key-remove` confirm screen are compared with
per-language goldens under `tests/fixtures/ui/{en,ko}/`. The English goldens were captured from
the binary before S6b, so an equal English run means the English bytes did not change. A
stream that is not a terminal stays English whatever the language settings say (S6-2), and
the stdout envelope is the same bytes on a pipe, on an English terminal and on a Korean one
(D3). POSIX only (the `pty` module); the binary is the one `self_audit_run.rust_bin` finds.
"""

import json
import os
import re
import tempfile
import unicodedata
import unittest

from tests.self_audit_run import ROOT, rust_bin
from tests.ui_progress_confirm_test import run_pipe, run_tty
from tests.ui_summary_tty_test import PACKAGE, run

GOLDEN = ROOT / "tests" / "fixtures" / "ui"
DIGEST = re.compile(r"sha256:[0-9a-f]{64}")
ESC = re.compile(r"\x1b\[[0-9;]*m")
KOREAN = {"MOAI_LANG": "ko"}


def env_for(home, **extra):
    env = {"PATH": "/usr/bin:/bin", "HOME": home, "LANG": "en_US.UTF-8",
           "TERM": "xterm-256color", "COLUMNS": "60", "NO_COLOR": "1"}
    env.update(extra)
    return env


def normal(text, *paths):
    """`text` with each temporary path as `<home>` and every digest as `sha256:<digest>`."""
    for p in sorted({q for path in paths for q in (os.path.realpath(path), path)}, key=len,
                    reverse=True):
        text = text.replace(p, "<home>")
    return DIGEST.sub("sha256:<digest>", text)


def cells(line):
    """Display width: East Asian wide and full-width characters take two cells."""
    return sum(2 if unicodedata.east_asian_width(c) in "WF" else 1 for c in line)


def golden(lang, name):
    return (GOLDEN / lang / f"{name}.golden").read_text(encoding="utf-8")


def capture_summary(home, cwd=None, tty=True, **extra):
    """(stdout, stderr, exit) of `home-routing-plan` with stderr on a pty when `tty`."""
    return run(["home-routing-plan", home, PACKAGE], env_for(home, **extra), cwd or home, tty)


def capture_key_refusal(home, work, **extra):
    """(stdout, terminal text) of `key-remove` answered `n` on a terminal."""
    env = env_for(home, LANG="C", TERM="xterm", **extra)
    out, err, code = run_pipe(["key-set", "glm", "--shared"], env, work, stdin="not-a-real-key\n")
    assert code == 0, err
    out, term, code = run_tty(["key-remove", "glm", "--shared"], env, work, answer=b"n\n")
    assert code == 0, term
    return out, term


def capture_index_stages(root, home, **extra):
    """The progress lines `index-build` writes on a terminal."""
    env = env_for(home, LANG="C", TERM="xterm", **extra)
    _, term, code = run_tty(["index-build", root], env, root, stdin_text='{"protocol_version":1}')
    assert code == 0, term
    return "".join(line + "\n" for line in term.splitlines() if line.startswith("->"))


def git_project(base, home, name):
    import subprocess

    root = os.path.join(base, name)
    os.makedirs(root)
    with open(os.path.join(root, "a.md"), "w") as f:
        f.write("# a\n")
    git = {"HOME": home, "PATH": "/usr/bin:/bin", "GIT_CONFIG_NOSYSTEM": "1",
           "GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@example.invalid",
           "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@example.invalid",
           "GIT_AUTHOR_DATE": "2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE": "2026-01-01T00:00:00Z"}
    for args in (["init", "-q"], ["add", "a.md"], ["commit", "-qm", "a"]):
        subprocess.run(["git", *args], cwd=root, env=git, check=True, capture_output=True)
    return root


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class KoreanTerminal(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        # Under /tmp the HOME path is short enough that the 60-column title never cuts it, so
        # the golden's `<home>` stands for the whole path on macOS and Linux alike.
        self.tmp = tempfile.TemporaryDirectory(dir="/tmp")
        self.home = os.path.join(self.tmp.name, "home")
        self.work = os.path.join(self.tmp.name, "work")
        os.makedirs(self.home)
        os.makedirs(self.work)

    def tearDown(self):
        self.tmp.cleanup()

    def summary(self, **extra):
        out, err, code = capture_summary(self.home, **extra)
        self.assertEqual(code, 0, out)
        return out, normal(ESC.sub("", err), self.home)

    def test_the_english_summary_keeps_its_bytes(self):
        _, err = self.summary()
        self.assertEqual(err, golden("en", "home-routing-plan"))

    def test_moai_lang_ko_gives_the_korean_summary_and_the_same_envelope(self):
        en_out, _ = self.summary()
        ko_out, err = self.summary(**KOREAN)
        self.assertEqual(err, golden("ko", "home-routing-plan"))
        self.assertEqual(ko_out, en_out, "the stdout envelope does not follow the language")

    def test_the_locale_chooses_korean_when_nothing_else_does(self):
        _, err = self.summary(LANG="ko_KR.UTF-8")
        self.assertEqual(err, golden("ko", "home-routing-plan"))
        # MOAI_LANG beats the locale.
        _, err = self.summary(LANG="ko_KR.UTF-8", MOAI_LANG="en")
        self.assertEqual(err, golden("en", "home-routing-plan"))

    def test_the_project_profile_chooses_korean(self):
        project = os.path.join(self.tmp.name, "proj")
        os.makedirs(project)
        _, err, code = run_pipe(["init"], env_for(self.home), project)
        self.assertEqual(code, 0, err)
        # The profile lives under a hidden state folder, which `glob` skips; walk instead.
        files = [os.path.join(d, "project.yaml") for d, _, names in os.walk(project)
                 if "project.yaml" in names]
        self.assertEqual(len(files), 1, files)
        with open(files[0]) as f:
            profile = json.load(f)
        profile["language"] = {"user_chat": "ko"}
        with open(files[0], "w") as f:
            json.dump(profile, f)
        out, err, code = capture_summary(self.home, cwd=project)
        self.assertEqual(code, 0, out)
        self.assertEqual(normal(ESC.sub("", err), self.home), golden("ko", "home-routing-plan"))

    def test_a_pipe_stays_english_whatever_the_language_says(self):
        tty_out, _ = self.summary()
        for extra in (KOREAN, {"LANG": "ko_KR.UTF-8", "LC_ALL": "ko_KR.UTF-8"}):
            out, err, code = capture_summary(self.home, tty=False, **extra)
            self.assertEqual(code, 0, out)
            self.assertEqual(err, "", f"{extra}: nothing human on a pipe")
            self.assertEqual(out, tty_out, f"{extra}: the same envelope bytes")

    def test_korean_lines_fit_sixty_cells(self):
        _, err = self.summary(**KOREAN)
        for line in err.splitlines():
            if line.strip().startswith("--plan-digest"):
                continue  # copied whole, never cut (the S4 rule)
            self.assertLessEqual(cells(line), 60, line)

    def test_plain_output_keeps_the_language(self):
        _, err = self.summary(MOAI_OUTPUT="plain", **KOREAN)
        self.assertNotIn("\x1b", err)
        self.assertIn("동작", err)
        self.assertIn("-> 다음:", err)

    def test_the_key_remove_screen_in_english_and_korean(self):
        out, term = capture_key_refusal(self.home, self.work)
        self.assertEqual(normal(term, self.home), golden("en", "key-remove-refused"))
        out_ko, term_ko = capture_key_refusal(self.home, self.work, **KOREAN)
        self.assertEqual(normal(term_ko, self.home), golden("ko", "key-remove-refused"))
        self.assertEqual(out_ko, out, "the refusal answers with the same envelope")
        # The key is still there (both answers were no): a pipe never asks, in any language,
        # and answers with the same preview envelope.
        piped, piped_err, code = run_pipe(["key-remove", "glm", "--shared"],
                                          env_for(self.home, LANG="C", **KOREAN), self.work)
        self.assertEqual(code, 0, piped_err)
        self.assertEqual(json.loads(piped)["result"]["exists"], True)
        self.assertNotIn("[y/N]", piped_err)
        self.assertNotIn("지울까요", piped_err)
        self.assertEqual(piped, out)
        # `MOAI_OUTPUT=quiet` still asks (without it the answer is no), in the same language.
        _, quiet = capture_key_refusal(self.home, self.work, MOAI_OUTPUT="quiet", **KOREAN)
        self.assertEqual(normal(quiet, self.home), golden("ko", "key-remove-refused"))

    def test_the_index_stages_in_english_and_korean(self):
        a = git_project(self.tmp.name, self.home, "a")
        self.assertEqual(capture_index_stages(a, self.home), golden("en", "index-build-stages"))
        b = git_project(self.tmp.name, self.home, "b")
        self.assertEqual(capture_index_stages(b, self.home, **KOREAN),
                         golden("ko", "index-build-stages"))


if __name__ == "__main__":
    unittest.main()
