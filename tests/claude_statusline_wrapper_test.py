"""#2383: the Claude Code statusline wrapper that reaches moai-cli without the PATH name `moai`.

tools/claude-statusline.sh is the repository source of the design §4.1 script
(docs/2383-statusline-switch-design.md). Every case runs the wrapper the way Claude Code
does (the file itself, through its shebang) in a temporary HOME whose name holds spaces and
shell metacharacters, with an environment built from scratch. The real HOME, the installed
moai-cli and the Go build are never run: the Rust and Go executables are fakes that record
their argv, cwd and stdin. PATH starts with a decoy `moai` executable that records a call
too, so any PATH lookup of the Go name shows up as a decoy record. The Rust name `moai-cli` is
looked up on PATH last (#2871), and only a test that calls `path_cli()` puts one there.

#2871 U3c: the install moves from `~/.moai-cli` to `~/.moai`. The wrapper tries
`~/.moai/path/moai-cli`, then `~/.moai-cli/path/moai-cli`, then `moai-cli` on PATH, then the Go
build; the cases below run it on the layout before the move, after it, and on both at once.

#3004 P6a: the one name is `moai`. Before those `moai-cli` places the wrapper tries
`~/.moai/path/moai`, then `~/.moai-cli/path/moai`, and runs one only when it is a link: the
install lays a link, while a plain file of that name can be an old Go shim that answers
`--version` as Rust and routes `statusline` to Go. The PATH name `moai` is still never looked up.
"""

import pathlib
import re
import shutil
import stat
import subprocess
import tempfile
import unittest

REPO = pathlib.Path(__file__).resolve().parent.parent
WRAPPER = REPO / "tools" / "claude-statusline.sh"
DESIGN = REPO / "docs" / "2383-statusline-switch-design.md"

RUST_REL = (".moai-cli", "path", "moai-cli")
MOVED_REL = (".moai", "path", "moai-cli")
GO_REL = ("go", "bin", "moai.old")
MOAI_MOVED_REL = (".moai", "path", "moai")
MOAI_OLD_REL = (".moai-cli", "path", "moai")
# Not valid UTF-8, no trailing newline, a NUL byte and a line break: the wrapper must pass
# the bytes through untouched (it never reads or re-serializes stdin).
STDIN = b'{"model":{"display_name":"Opus"},"cwd":"/x y"}\n\xff\xfe\x00tail'

FAKE = """#!/bin/sh
printf '%s|%s|%s\\n' "{name}" "$*" "$(pwd)" >> "$FAKE_LOG"
/bin/cat > "$FAKE_LOG.{name}.stdin"
printf 'out-{name}'
exit {code}
"""


class ClaudeStatuslineWrapperTest(unittest.TestCase):
    def setUp(self):
        # resolve(): on macOS the temp dir sits under the /var -> /private/var link, and the
        # fakes report the physical cwd.
        self.tmp = pathlib.Path(tempfile.mkdtemp(prefix="moai-2383-")).resolve()
        self.addCleanup(shutil.rmtree, self.tmp, ignore_errors=True)
        self.home = self.tmp / "home dir $(x);'q\"&"
        self.home.mkdir()
        self.cwd = self.tmp / "project dir"
        self.cwd.mkdir()
        self.log = self.tmp / "calls.log"
        decoys = self.tmp / "decoy-bin"
        decoys.mkdir()
        self._fake(decoys / "moai", "decoy-moai", 0)
        self.pathbin = self.tmp / "path-bin"
        self.pathbin.mkdir()
        self.env = {"HOME": str(self.home), "PATH": f"{self.pathbin}:{decoys}:/usr/bin:/bin",
                    "FAKE_LOG": str(self.log)}

    def _fake(self, path, name, code, mode=0o755):
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(FAKE.format(name=name, code=code), encoding="utf-8")
        path.chmod(mode)
        return path

    def rust(self, code=0, mode=0o755):
        return self._fake(self.home.joinpath(*RUST_REL), "rust", code, mode)

    def moved(self, name="rust", code=0, mode=0o755):
        return self._fake(self.home.joinpath(*MOVED_REL), name, code, mode)

    def path_cli(self, code=0, mode=0o755):
        return self._fake(self.pathbin / "moai-cli", "pathcli", code, mode)

    def moai_link(self, rel, name, code=0, mode=0o755):
        """The installed shape of `path/moai`: a link to a build under `versions/`."""
        link = self.home.joinpath(*rel)
        target = self._fake(link.parent.parent / "versions" / f"v-{name}" / "bin" / "moai", name, code, mode)
        link.parent.mkdir(parents=True, exist_ok=True)
        link.symlink_to(target)
        return link

    def go(self, code=0, mode=0o755):
        return self._fake(self.home.joinpath(*GO_REL), "go", code, mode)

    def run_wrapper(self):
        return subprocess.run([str(WRAPPER)], input=STDIN, cwd=self.cwd, env=self.env,
                              capture_output=True, timeout=30)

    def calls(self):
        if not self.log.exists():
            return []
        return [line.split("|") for line in self.log.read_text(encoding="utf-8").splitlines()]

    def assert_called_once(self, name):
        self.assertEqual(self.calls(), [[name, "statusline", str(self.cwd)]])
        self.assertEqual(pathlib.Path(f"{self.log}.{name}.stdin").read_bytes(), STDIN)

    def test_rust_is_preferred_with_stdin_argv_and_cwd_passed_through(self):
        self.rust()
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-rust", b""))
        self.assert_called_once("rust")

    def test_a_failing_rust_keeps_its_exit_code_and_go_is_not_run_as_well(self):
        self.rust(code=3)
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (3, b"out-rust"))
        self.assert_called_once("rust")

    def test_rust_through_a_symlink_like_the_installed_link(self):
        target = self._fake(self.home / ".moai-cli" / "versions" / "v1" / "bin" / "moai-cli", "rust", 0)
        link = self.home.joinpath(*RUST_REL)
        link.parent.mkdir(parents=True)
        link.symlink_to(target)
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_a_non_executable_rust_falls_back_to_go(self):
        self.rust(mode=0o644)
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    def test_without_rust_go_renders_by_absolute_path(self):
        self.go(code=5)
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (5, b"out-go"))
        self.assert_called_once("go")

    def test_a_non_executable_go_is_not_executed(self):
        self.go(mode=0o644)
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"", b""))
        self.assertEqual(self.calls(), [])

    # #2503 (audit N1 on PR #2482): `-x` is true for a searchable directory, so a directory at
    # either path used to reach `exec` and end with 126 instead of falling through.
    def test_a_directory_at_the_rust_path_falls_back_to_go(self):
        self.home.joinpath(*RUST_REL).mkdir(parents=True)
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    def test_a_link_to_a_directory_at_the_rust_path_falls_back_to_go(self):
        target = self.home / ".moai-cli" / "versions" / "v1"
        target.mkdir(parents=True)
        link = self.home.joinpath(*RUST_REL)
        link.parent.mkdir(parents=True)
        link.symlink_to(target)
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    def test_directories_at_both_paths_exit_0_silently(self):
        self.home.joinpath(*RUST_REL).mkdir(parents=True)
        self.home.joinpath(*GO_REL).mkdir(parents=True)
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"", b""))
        self.assertEqual(self.calls(), [])

    def test_nothing_installed_exits_0_silently_and_never_uses_path(self):
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"", b""))
        self.assertEqual(self.calls(), [])

    # #2871 U3c: the layout after `home-install-apply` moved the install to ~/.moai.
    def test_after_the_move_the_moved_cli_renders(self):
        self.moved()
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-rust", b""))
        self.assert_called_once("rust")

    def test_after_the_move_the_installed_link_renders(self):
        target = self._fake(self.home / ".moai" / "versions" / "v1" / "bin" / "moai-cli", "rust", 0)
        link = self.home.joinpath(*MOVED_REL)
        link.parent.mkdir(parents=True)
        link.symlink_to(target)
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_the_moved_cli_wins_when_both_layouts_exist(self):
        self.rust()
        self.moved(name="moved")
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-moved"))
        self.assert_called_once("moved")

    def test_a_directory_at_the_moved_path_falls_back_to_the_old_layout(self):
        self.home.joinpath(*MOVED_REL).mkdir(parents=True)
        self.rust()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-rust", b""))
        self.assert_called_once("rust")

    def test_a_non_executable_moved_cli_falls_back_to_the_old_layout(self):
        self.moved(name="moved", mode=0o644)
        self.rust()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_a_failing_moved_cli_keeps_its_exit_code_and_nothing_else_runs(self):
        self.moved(code=4)
        self.rust(code=0)
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (4, b"out-rust"))
        self.assert_called_once("rust")

    # #3004 P6a: the link `path/moai` is tried before every `moai-cli` place.
    def test_the_moved_moai_link_renders_and_wins_over_the_moai_cli_places(self):
        self.moai_link(MOAI_MOVED_REL, "moai-moved")
        self.moved()
        self.rust()
        self.path_cli()
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-moai-moved", b""))
        self.assert_called_once("moai-moved")

    def test_the_old_layout_moai_link_renders_and_wins_over_its_moai_cli(self):
        self.moai_link(MOAI_OLD_REL, "moai-old")
        self.rust()
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-moai-old"))
        self.assert_called_once("moai-old")

    def test_the_moved_moai_link_wins_over_the_old_layout_moai_link(self):
        self.moai_link(MOAI_OLD_REL, "moai-old")
        self.moai_link(MOAI_MOVED_REL, "moai-moved")
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-moai-moved"))
        self.assert_called_once("moai-moved")

    def test_a_plain_file_named_moai_is_never_run_it_may_be_an_old_go_shim(self):
        for rel in (MOAI_MOVED_REL, MOAI_OLD_REL):
            self._fake(self.home.joinpath(*rel), "shim", 0)
        self.rust()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-rust", b""))
        self.assert_called_once("rust")

    def test_a_plain_moai_file_alone_falls_through_to_go(self):
        self._fake(self.home.joinpath(*MOAI_MOVED_REL), "shim", 0)
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-go"))
        self.assert_called_once("go")

    def test_a_non_executable_moai_link_falls_through(self):
        self.moai_link(MOAI_MOVED_REL, "moai-moved", mode=0o644)
        self.moved()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_a_moai_link_to_a_directory_falls_through(self):
        target = self.home / ".moai" / "versions" / "v1"
        target.mkdir(parents=True)
        link = self.home.joinpath(*MOAI_MOVED_REL)
        link.parent.mkdir(parents=True, exist_ok=True)
        link.symlink_to(target)
        self.moved()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_a_failing_moai_link_keeps_its_exit_code_and_nothing_else_runs(self):
        self.moai_link(MOAI_MOVED_REL, "moai-moved", code=6)
        self.moved()
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (6, b"out-moai-moved"))
        self.assert_called_once("moai-moved")

    def test_the_decoy_moai_on_path_is_still_never_run(self):
        self.moai_link(MOAI_MOVED_REL, "moai-moved")
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-moai-moved"))
        self.assertNotIn("decoy-moai", [c[0] for c in self.calls()])

    # The name `moai-cli` on PATH is the last place; the Go name `moai` is never looked up.
    def test_moai_cli_on_path_renders_when_no_home_layout_has_it(self):
        self.path_cli()
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-pathcli", b""))
        self.assert_called_once("pathcli")

    def test_a_home_layout_wins_over_moai_cli_on_path(self):
        self.path_cli()
        self.moved()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_the_old_layout_wins_over_moai_cli_on_path(self):
        self.path_cli()
        self.rust()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_a_non_executable_moai_cli_on_path_falls_back_to_go(self):
        self.path_cli(mode=0o644)
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    # #3008 audit B1: with `.` or an empty entry in PATH, `command -v moai-cli` answers the bare
    # name, and the status line runs with the project folder as its cwd. A file named moai-cli
    # that sits there must never be executed.
    def _plant_in_cwd(self):
        return self._fake(self.cwd / "moai-cli", "planted", 0)

    def _path_with(self, entry_first):
        base = self.env["PATH"]
        self.env["PATH"] = f"{entry_first}{base}"

    def test_a_dot_in_path_never_runs_a_file_from_the_current_folder(self):
        self._plant_in_cwd()
        self._path_with(".:")
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    def test_an_empty_path_entry_never_runs_a_file_from_the_current_folder(self):
        self._plant_in_cwd()
        self._path_with(":")
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    def test_a_trailing_empty_path_entry_never_runs_a_file_from_the_current_folder(self):
        self._plant_in_cwd()
        self.env["PATH"] = self.env["PATH"] + ":"
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    def test_a_real_moai_cli_before_a_dot_entry_in_path_still_wins(self):
        self._plant_in_cwd()
        self.env["PATH"] = self.env["PATH"] + ":."
        self.path_cli()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-pathcli"))
        self.assert_called_once("pathcli")

    def test_a_directory_named_moai_cli_on_path_falls_back_to_go(self):
        (self.pathbin / "moai-cli").mkdir()
        self.go()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"out-go", b""))
        self.assert_called_once("go")

    def test_the_go_env_file_is_never_sourced(self):
        # The Go wrapper sourced ~/.moai/.env.glm; design §4.1 drops it (no credentials in a render).
        env_file = self.home / ".moai" / ".env.glm"
        env_file.parent.mkdir()
        env_file.write_text('printf "sourced|x|x\\n" >> "$FAKE_LOG"\n', encoding="utf-8")
        self.rust()
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout), (0, b"out-rust"))
        self.assert_called_once("rust")

    def test_the_wrapper_is_an_executable_bash_script_that_parses(self):
        self.assertTrue(WRAPPER.stat().st_mode & stat.S_IXUSR, "the settings command runs the file itself")
        self.assertTrue(WRAPPER.read_bytes().startswith(b"#!/bin/bash\n"))
        r = subprocess.run(["/bin/bash", "-n", str(WRAPPER)], capture_output=True, timeout=30)
        self.assertEqual((r.returncode, r.stderr), (0, b""))

    def test_only_moai_cli_is_looked_up_on_path_no_sourced_env_and_no_temporary_file(self):
        lines = WRAPPER.read_text(encoding="utf-8").splitlines()
        code = [line for line in lines if not line.lstrip().startswith("#")]
        body = "\n".join(code)
        self.assertEqual(re.findall(r"command -v ([\w-]+)", body), ["moai-cli"])
        self.assertNotIn("which ", body)
        self.assertNotRegex(body, r"(^|\s)(source|\.)\s", "the Go .env.glm source is dropped (design §4.1)")
        self.assertNotIn("mktemp", body)
        self.assertEqual(re.findall(r'exec "\$\w+" statusline', body),
                         ['exec "$MOAI_CLI" statusline', 'exec "$MOAI_GO" statusline'])

    def test_the_design_section_4_1_block_is_the_wrapper_byte_for_byte(self):
        doc = DESIGN.read_text(encoding="utf-8")
        section = doc.split("### 4.1 ", 1)[1]
        block = re.search(r"```bash\n(.*?)```", section, re.S).group(1)
        self.assertEqual(block, WRAPPER.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
