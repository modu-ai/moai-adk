"""#2383: the Claude Code statusline wrapper that reaches moai-cli without the PATH name `moai`.

tools/claude-statusline.sh is the repository source of the design §4.1 script
(docs/2383-statusline-switch-design.md). Every case runs the wrapper the way Claude Code
does (the file itself, through its shebang) in a temporary HOME whose name holds spaces and
shell metacharacters, with an environment built from scratch. The real HOME, the installed
moai-cli and the Go build are never run: the Rust and Go executables are fakes that record
their argv, cwd and stdin. PATH starts with decoy `moai` and `moai-cli` executables that
record a call too, so any PATH lookup of either name shows up as a decoy record.
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
GO_REL = ("go", "bin", "moai.old")
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
        for name in ("moai", "moai-cli"):
            self._fake(decoys / name, "decoy-" + name, 0)
        self.env = {"HOME": str(self.home), "PATH": f"{decoys}:/usr/bin:/bin", "FAKE_LOG": str(self.log)}

    def _fake(self, path, name, code, mode=0o755):
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(FAKE.format(name=name, code=code), encoding="utf-8")
        path.chmod(mode)
        return path

    def rust(self, code=0, mode=0o755):
        return self._fake(self.home.joinpath(*RUST_REL), "rust", code, mode)

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

    def test_nothing_installed_exits_0_silently_and_never_uses_path(self):
        r = self.run_wrapper()
        self.assertEqual((r.returncode, r.stdout, r.stderr), (0, b"", b""))
        self.assertEqual(self.calls(), [])

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

    def test_no_path_lookup_no_sourced_env_and_no_temporary_file(self):
        lines = WRAPPER.read_text(encoding="utf-8").splitlines()
        code = [line for line in lines if not line.lstrip().startswith("#")]
        body = "\n".join(code)
        self.assertNotIn("command -v", body)
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
