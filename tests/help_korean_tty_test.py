"""Korean root help on a terminal (CLI design S6d, Issue #2825; design
`docs/design-review/cli-s6-korean-2026-11-14.md` §1 and §8).

Help goes to stdout, so its language is stdout's: Korean only when stdout is a terminal and the
language is `ko`, English bytes otherwise — a pipe whatever the settings say (S6-2), and a run
whose stdout is a pipe even when stderr is a terminal (the two streams choose apart, agent-22
6873424675). The pipe bytes are pinned by `tests/help_layout.rs`; here an English terminal must
give the same bytes, and the Korean screen is compared with `tests/fixtures/ui/ko/`. A
subcommand's help (S6e-1) keeps its own texts and takes a Korean frame: the usage, arguments,
options and examples headings and the help flag's line. POSIX only.
"""

import os
import re
import shutil
import subprocess
import tempfile
import unittest

from tests.self_audit_run import ROOT, rust_bin
from tests.ui_korean_tty_test import KOREAN, cells, env_for

GOLDEN = ROOT / "tests" / "fixtures" / "ui"
VERSION = re.compile(r"^MoAI-ADK \S+ — ", re.M)


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
    p = subprocess.run([moai_in(env["HOME"]), *args], stdin=subprocess.DEVNULL, capture_output=True,
                       text=True, env=env, cwd=env["HOME"], timeout=60)
    return p.stdout, p.stderr, p.returncode


@unittest.skipUnless(os.name == "posix", "pseudo-terminals are POSIX only")
class KoreanHelp(unittest.TestCase):
    def setUp(self):
        if not rust_bin().exists():
            raise AssertionError(f"moai-cli not built at {rust_bin()}; run cargo build first")
        self.tmp = tempfile.TemporaryDirectory(dir="/tmp")
        self.home = self.tmp.name

    def tearDown(self):
        self.tmp.cleanup()

    def help(self, args=("--help",), **extra):
        out, err, code = run_help(list(args), env_for(self.home, **extra))
        self.assertEqual(code, 0, err)
        return normal(out)

    def test_an_english_terminal_gets_the_pipe_bytes(self):
        for args in (["--help"], ["help", "--all"]):
            piped, _, code = pipe(args, env_for(self.home))
            self.assertEqual(code, 0)
            self.assertEqual(self.help(args), normal(piped), args)

    def test_the_korean_root_help(self):
        self.assertEqual(self.help(**KOREAN),
                         (GOLDEN / "ko" / "help-root.golden").read_text(encoding="utf-8"))

    def test_the_locale_chooses_korean_help(self):
        self.assertEqual(self.help(LANG="ko_KR.UTF-8"), self.help(**KOREAN))
        self.assertEqual(self.help(LANG="ko_KR.UTF-8", MOAI_LANG="en"), self.help())

    def test_korean_help_fits_the_terminal(self):
        for columns in ("60", "80"):
            for args in (["--help"], ["help", "--all"]):
                text = self.help(args, COLUMNS=columns, **KOREAN)
                self.assertIn("사용법:", text)
                for line in text.splitlines():
                    self.assertLessEqual(cells(line), int(columns), (columns, args, line))

    def test_the_full_list_has_korean_headings_and_options(self):
        text = self.help(["help", "--all"], **KOREAN)
        for heading in ("시작", "프로젝트", "업데이트와 이동", "키", "진단", "레인과 리더",
                        "검사", "내부"):
            self.assertIn(f"\n{heading}\n", text)
        self.assertIn("\n옵션:\n", text)
        self.assertNotIn("\nOptions:\n", text)

    def test_a_pipe_stays_english_whatever_the_language_says(self):
        english, _, _ = pipe(["--help"], env_for(self.home))
        for extra in (KOREAN, {"LANG": "ko_KR.UTF-8", "LC_ALL": "ko_KR.UTF-8"}):
            out, err, code = pipe(["--help"], env_for(self.home, **extra))
            self.assertEqual(code, 0, err)
            self.assertEqual(out, english, extra)

    def test_a_subcommand_help_has_a_korean_frame(self):
        """S6e-1: the usage, arguments, options and examples headings and the help flag's line
        read Korean; a command of a group not moved yet keeps its own texts."""
        text = self.help(["identity", "--help"], **KOREAN)
        for head in ("\n사용법: moai identity <PATH>\n", "\n인자:\n", "\n옵션:\n",
                     "\n예시:\n  moai identity .\n"):
            self.assertIn(head, text)
        self.assertIn("도움말 보기 (요약은 '-h')", text)
        self.assertIn("Any path inside a clone or worktree", text)
        for word in ("Usage:", "Arguments:", "Options:", "Examples:", "Print help"):
            self.assertNotIn(word, text)
        short = self.help(["key-set", "-h"], **KOREAN)
        self.assertIn("도움말 보기 (자세히는 '--help')", short)
        self.assertIn("\n옵션:\n      --shared", short)

    def test_the_project_group_reads_its_own_texts_in_korean(self):
        """S6e-2: the Project group's root commands — about, long about, argument help."""
        detect = self.help(["project-detect", "--help"], **KOREAN)
        self.assertIn("`/moai project` 경로(#2496)", detect)
        self.assertIn("\n          프로젝트 루트\n", detect)
        self.assertNotIn("Detect the project's facts", detect)
        setup = self.help(["language-setup", "--help"], **KOREAN)
        self.assertIn("호스트 질문 도구용 질문 묶음 찍기", setup)
        self.assertIn("대화 기본 항목의 대화 언어 코드", setup)
        self.assertTrue(self.help(["index-build", "-h"], **KOREAN).startswith("코드 지도 색인 만들기\n"))
        apply = self.help(["project-apply", "--help"], **KOREAN)
        self.assertIn("호출자가 준 권한 참조", apply)
        for columns in ("60", "80"):
            for word in ("project-detect", "project-plan", "project-apply", "language-setup",
                         "index-build"):
                for flag in ("-h", "--help"):
                    text = self.help([word, flag], COLUMNS=columns, **KOREAN)
                    for line in text.splitlines():
                        if line.startswith("사용법: ") or line.startswith("  moai "):
                            continue  # clap never wraps a usage or an example, in English too
                        self.assertLessEqual(cells(line), int(columns),
                                             (columns, word, flag, line))

    def test_the_keys_group_reads_its_own_texts_in_korean(self):
        """S6e-3: the Keys group's root commands — about, long about, argument help — and
        every line inside 60 and 80 cells but the usage and example lines clap never wraps."""
        text = self.help(["key-set", "--help"], **KOREAN)
        self.assertIn("공급자 키를 암호화해 금고에 저장해요(#2393)", text)
        self.assertIn("공급자 이름, 예: `glm`", text)
        self.assertIn("모든 프로젝트가 쓰는", text)
        imp = self.help(["key-import-plan", "--help"], **KOREAN)
        self.assertIn("프로젝트 대신 `~/.moai-cli/vault/_shared/`", imp)
        self.assertIn("묻지 않고 지우기", self.help(["key-remove", "--help"], **KOREAN))
        for columns in ("60", "80"):
            for word in ("key-set", "key-status", "key-remove", "key-import-plan",
                         "key-import-apply"):
                for flag in ("-h", "--help"):
                    for line in self.help([word, flag], COLUMNS=columns, **KOREAN).splitlines():
                        if line.startswith("사용법: ") or line.startswith("  moai "):
                            continue
                        self.assertLessEqual(cells(line), int(columns),
                                             (columns, word, flag, line))

    def test_no_korean_line_of_any_command_help_runs_past_sixty_cells(self):
        """Every command's `-h` and `--help` at 60 columns: no line with Hangul past 60 cells
        (clap counts a Hangul letter as one cell, so a Korean help beside a wide name ran to 72
        before the Korean frame put every help on its own line)."""
        listing, _, _ = pipe(["help", "--all"], env_for(self.home))
        words = sorted({line.split()[0] for line in listing.splitlines()
                        if line.startswith("  ") and not line.startswith("   ")
                        and not line.split()[0].startswith("-")})
        self.assertGreater(len(words), 100)
        over = []
        for word in words:
            for flag in ("-h", "--help"):
                for line in self.help([word, flag], COLUMNS="60", **KOREAN).splitlines():
                    if line.startswith("사용법: ") or line.startswith("  moai "):
                        continue  # clap never wraps a usage or an example, in English too
                    if cells(line) > 60 and any("가" <= c <= "힣" for c in line):
                        over.append((word, flag, cells(line), line))
        self.assertEqual(over, [])

    def test_an_english_subcommand_help_keeps_its_bytes(self):
        for args in (["project-detect", "--help"], ["key-set", "-h"], ["cc", "--help"]):
            piped, _, code = pipe(args, env_for(self.home))
            self.assertEqual(code, 0)
            self.assertEqual(self.help(args), normal(piped), args)
            ko_piped, _, _ = pipe(args, env_for(self.home, **KOREAN))
            self.assertEqual(ko_piped, piped, f"{args}: a pipe stays English")

    def test_korean_subcommand_help_fits_the_terminal(self):
        for columns in ("60", "80"):
            for args in (["project-detect", "--help"], ["key-set", "--help"], ["cc", "--help"],
                         ["update", "-h"]):
                text = self.help(args, COLUMNS=columns, **KOREAN)
                self.assertIn("사용법:", text)
                for line in text.splitlines():
                    self.assertLessEqual(cells(line), int(columns), (columns, args, line))

    def test_a_usage_error_on_a_piped_stderr_stays_english(self):
        """Audit #2909 B1: clap prints the root usage in a usage error, so with stdout a Korean
        terminal and stderr a pipe the error bytes are the all-pipe bytes."""
        # Both ways to Korean (re-audit 6874932388 note: the locale was only probed).
        for korean in (KOREAN, {"LANG": "ko_KR.UTF-8"}):
            for args in (["--bogus"], ["bogus"], ["project-detect"]):
                _, piped_err, piped_code = pipe(args, env_for(self.home))
                out, err, code = run_help(args, env_for(self.home, **korean), stdout_tty=True,
                                          stderr_tty=False)
                self.assertEqual(code, piped_code, (korean, args))
                self.assertNotEqual(code, 0, (korean, args))
                self.assertEqual(err, piped_err, (korean, args))
                self.assertNotIn("에이전트", err, (korean, args))

    def test_help_follows_stdout_not_stderr(self):
        """stdout a pipe and stderr a terminal: English help; stdout a terminal and stderr a
        pipe: Korean help. Each stream chooses its own language."""
        out, _, code = run_help(["--help"], env_for(self.home, **KOREAN), stdout_tty=False,
                                stderr_tty=True)
        self.assertEqual(code, 0)
        english, _, _ = pipe(["--help"], env_for(self.home))
        self.assertEqual(out, english)
        self.assertIn("\nUsage: moai <command>", out)
        out, err, code = run_help(["--help"], env_for(self.home, **KOREAN), stdout_tty=True,
                                  stderr_tty=False)
        self.assertEqual(code, 0)
        self.assertIn("사용법:", out)
        self.assertEqual(err, "")


if __name__ == "__main__":
    unittest.main()
