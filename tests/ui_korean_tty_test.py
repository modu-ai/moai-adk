"""Korean human output on a terminal (CLI design S6b and S6c, Issue #2825; design
`docs/design-review/cli-s6-korean-2026-11-14.md` §6, leader decision 6870716572).

The S4 summary, the S5 progress lines, the `key-remove` confirm screen (S6b) and the S3 error
frame (S6c) are compared with per-language goldens under `tests/fixtures/ui/{en,ko}/`. The
English goldens were captured from the binary before each card, so an equal English run means
the English bytes did not change. A
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


# S6c (#2825, design 5 and 8): the S3 error frame. The frame's labels and next step follow the
# terminal's language; a usage error's first line is our sentence on a Korean terminal (clap's
# missing-argument names stay, nothing the user typed is put in); an operation error's message
# is the envelope's `error.message` and stays English.
ERROR_CASES = {
    "usage-missing": ["project-plan"],
    "usage-unknown": ["project-plan", "--nope", "{proj}"],
    "operation-permission": ["hooks-apply", "--host", "claude", "{proj}", "--plan-digest", "x",
                             "--permission-reference", ""],
}


def capture_error(name, home, tty=True, binary=None, **extra):
    """(exit, stdout bytes, stderr text) of ERROR_CASES[name] with `NO_COLOR=1`; stderr on a
    pty when `tty`. `binary` runs another build under the name `moai-cli` (the frame names the
    program as run)."""
    from tests import error_frame_tty_test as frame

    proj = os.path.join(home, "p")
    os.makedirs(proj, exist_ok=True)
    args = [a.replace("{proj}", proj) for a in ERROR_CASES[name]]
    env = frame.base_env(home, NO_COLOR="1", **extra)
    run_one = frame.stderr_on_tty if tty else frame.piped
    if binary is None:
        return run_one(args, env, home)
    link_dir = os.path.join(home, "bin")
    os.makedirs(link_dir, exist_ok=True)
    link = os.path.join(link_dir, "moai-cli")
    if not os.path.exists(link):
        os.symlink(binary, link)
    old = os.environ.get("MOAI_CLI_BIN")
    os.environ["MOAI_CLI_BIN"] = link
    try:
        return run_one(args, env, home)
    finally:
        if old is None:
            del os.environ["MOAI_CLI_BIN"]
        else:
            os.environ["MOAI_CLI_BIN"] = old


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class KoreanErrorFrame(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory(dir="/tmp")
        self.home = os.path.join(self.tmp.name, "home")
        os.makedirs(self.home)

    def tearDown(self):
        self.tmp.cleanup()

    def frame(self, name, **extra):
        code, out, err = capture_error(name, self.home, **extra)
        return code, out, normal(err, self.home)

    def test_english_frames_keep_their_bytes(self):
        for name in ERROR_CASES:
            _, _, err = self.frame(name)
            self.assertEqual(err, golden("en", f"error-{name}"), name)

    def test_korean_frames_and_the_same_envelope(self):
        for name in ERROR_CASES:
            code_en, out_en, _ = self.frame(name)
            code_ko, out_ko, err = self.frame(name, **KOREAN)
            self.assertEqual(err, golden("ko", f"error-{name}"), name)
            self.assertEqual((code_ko, out_ko), (code_en, out_en), f"{name}: same exit, envelope")
        # The typed argument never reaches a Korean frame.
        _, _, err = self.frame("usage-unknown", **KOREAN)
        self.assertNotIn("--nope", err)

    def test_a_pipe_keeps_the_english_bytes_in_any_language(self):
        for name in ERROR_CASES:
            plain = capture_error(name, self.home, tty=False)
            for extra in (KOREAN, {"LANG": "ko_KR.UTF-8", "LC_ALL": "ko_KR.UTF-8"}):
                self.assertEqual(capture_error(name, self.home, tty=False, **extra), plain,
                                 f"{name} {extra}")

    def test_a_control_character_in_the_program_name_is_shown_as_text(self):
        # The next step names the program as run; a name with an OSC title and BEL is data.
        evil_dir = os.path.join(self.tmp.name, "evil")
        os.makedirs(evil_dir)
        evil = os.path.join(evil_dir, "mo\x1b]0;T\x07ai")
        os.symlink(rust_bin(), evil)
        old = os.environ.get("MOAI_CLI_BIN")
        os.environ["MOAI_CLI_BIN"] = evil
        try:
            from tests import error_frame_tty_test as frame
            _, _, err = frame.stderr_on_tty(["project-plan"],
                                            frame.base_env(self.home, NO_COLOR="1", **KOREAN),
                                            self.home)
        finally:
            if old is None:
                del os.environ["MOAI_CLI_BIN"]
            else:
                os.environ["MOAI_CLI_BIN"] = old
        self.assertNotIn("\x1b]", err)
        self.assertNotIn("\x07", err)
        self.assertIn("`mo\\x1b]0;T\\x07ai project-plan --help` 로 인자를 확인하세요", err)


if __name__ == "__main__":
    unittest.main()
