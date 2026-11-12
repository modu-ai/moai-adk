"""Executable alias-resolution boundaries, not evidence of real Rust compilation.

The six real compiler regressions remain in mutate_compiler_test.py. These
synthetic native tools implement only deterministic file-producing commands.
"""

import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import unittest

import mutate_compiler_test as compiler
import mutate_test as support


PROXY = r'''
import json
import os
from pathlib import Path
import sys

name = Path(sys.argv[0]).name
args = sys.argv[1:]
with open(os.environ["ALIAS_LOG"], "a") as stream:
    stream.write(json.dumps({"name": name, "args": args,
                             "home": os.environ["HOME"],
                             "cargo_home": os.environ["CARGO_HOME"],
                             "rustup_home": os.environ["RUSTUP_HOME"]}) + "\n")
# The installer rejects even `which`: only the preserved manager alias may look up.
if name == "rustup" and len(args) == 2 and args[0] == "which":
    paths = json.loads(Path(os.environ["ALIAS_LOOKUP"]).read_text())
    print(paths[args[1]])
    sys.exit(0)
if name in ("cargo", "rustc"):
    native = str(Path(os.environ["ALIAS_NATIVE"]) / name)
    os.execve(native, [native] + args, os.environ)
print("FIXTURE_INSTALLER_OR_INVALID_DISPATCH: " + name, file=sys.stderr)
sys.exit(64)
'''

NATIVE = r'''
import hashlib
import json
import os
from pathlib import Path
import sys

name = Path(sys.argv[0]).name
args = sys.argv[1:]
with open(os.environ["NATIVE_LOG"], "a") as stream:
    stream.write(json.dumps({"name": name, "args": args,
                             "home": os.environ["HOME"]}) + "\n")
# Intentional fake file protocol: neither command compiles real Rust.
if name == "cargo" and args == ["generate-lockfile", "--offline"]:
    digest = hashlib.sha256(Path("Cargo.toml").read_bytes()).hexdigest()
    Path("Cargo.lock").write_text("synthetic-lock:" + digest + "\n")
elif name == "rustc" and len(args) == 4 and args[1:3] == ["--emit=metadata", "-o"]:
    digest = hashlib.sha256(Path(args[0]).read_bytes()).digest()
    Path(args[3]).write_bytes(b"synthetic-metadata:" + digest)
else:
    print("FIXTURE_NATIVE_UNSUPPORTED_COMMAND", file=sys.stderr)
    sys.exit(65)
'''


class NativeToolAliasTests(unittest.TestCase):
    """Compose the shared fixture without inheriting unrelated CLI tests."""

    def setUp(self):
        self.case = support.MutationCLI("test_malformed_json_is_rejected_without_running_cargo")
        self.addCleanup(self.case.doCleanups)
        self.case.setUp()
        self.root = self.case.base / "native-alias"
        self.bin = self.root / "bin"
        self.proxy = self.root / "proxy"
        self.native = self.root / "native"
        self.project = self.root / "project"
        self.outer_home = self.root / "outer-home"
        self.inner_home = self.root / "sut-home"
        for path in (self.bin, self.proxy, self.native, self.project,
                     self.outer_home, self.inner_home):
            path.mkdir(parents=True)
        self.alias_log = self.root / "alias.jsonl"
        self.native_log = self.root / "native.jsonl"
        self.lookup = self.root / "lookup.json"
        self.lookup.write_text(json.dumps({name: str(self.native / name)
                                          for name in ("cargo", "rustc")}))
        # No inherited PATH, HOME, Rust configuration, keys, or host manager.
        self.outer = dict(PATH=str(self.bin), HOME=str(self.outer_home),
                          CARGO_HOME=str(self.outer_home / "cargo"),
                          RUSTUP_HOME=str(self.outer_home / "rustup"),
                          TMPDIR=str(self.root), ALIAS_LOG=str(self.alias_log),
                          NATIVE_LOG=str(self.native_log), ALIAS_LOOKUP=str(self.lookup),
                          ALIAS_NATIVE=str(self.native))
        for name in ("cargo", "rustc"):
            self.executable(self.native / name, NATIVE)
        (self.project / "Cargo.toml").write_text(
            '[package]\nname="synthetic_alias_protocol"\nversion="0.0.0"\n')
        (self.project / "lib.rs").write_text("pub fn synthetic_boundary() {}\n")

    def executable(self, path, source):
        path.write_text("#!" + sys.executable + "\n" + source)
        path.chmod(0o755)

    def aliases(self, target="rustup-init", hardlink=False, manager=True):
        payload = self.proxy / target
        self.executable(payload, PROXY)
        for name in ("cargo", "rustc", "rustup") if manager else ("cargo", "rustc"):
            alias = self.bin / name
            if hardlink:
                try:
                    os.link(payload, alias)
                except (NotImplementedError, OSError) as error:
                    self.skipTest("temporary filesystem does not support hardlinks: " + str(error))
            else:
                alias.symlink_to(payload)
        return payload

    def events(self, path):
        return [json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []

    def run_tool(self, tool, args, inner=False):
        environment = dict(self.outer)
        if inner:
            environment.update(HOME=str(self.inner_home),
                               CARGO_HOME=str(self.inner_home / "cargo"),
                               RUSTUP_HOME=str(self.inner_home / "rustup"))
        return support.run_with_progress_budget(
            [str(tool)] + args, cwd=self.project, env=environment,
            log=self.alias_log)

    def assert_command(self, name, tool):
        if name == "cargo":
            output = self.project / "Cargo.lock"
            args = ["generate-lockfile", "--offline"]
            digest = hashlib.sha256((self.project / "Cargo.toml").read_bytes()).hexdigest()
            expected = ("synthetic-lock:" + digest + "\n").encode()
        else:
            output = self.project / "lib.rmeta"
            args = ["lib.rs", "--emit=metadata", "-o", str(output)]
            expected = b"synthetic-metadata:" + hashlib.sha256(
                (self.project / "lib.rs").read_bytes()).digest()
        result = self.run_tool(tool, args, inner=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(output.read_bytes(), expected)
        self.assertEqual(self.events(self.native_log)[-1],
                         {"name": name, "args": args, "home": str(self.inner_home)})

    def assert_no_installer(self):
        self.assertFalse(any(event["name"] not in ("rustup", "cargo", "rustc")
                             for event in self.events(self.alias_log)),
                         "native_tool invoked the resolved installer rather than its manager alias")

    def assert_native_selection(self):
        self.addCleanup(self.assert_no_installer)
        for name in ("cargo", "rustc"):
            with self.subTest(tool=name):
                selected = compiler.native_tool(name, self.outer)
                # Fail before execution if the helper returns an installer/proxy.
                self.assertEqual(selected, self.native / name)
                self.assert_command(name, selected)
        for event in self.events(self.alias_log):
            self.assertEqual(event["name"], "rustup")
            self.assertIn(event["args"], (["which", "cargo"], ["which", "rustc"]))
            self.assertEqual(event["home"], str(self.outer_home))
            self.assertEqual(event["cargo_home"], self.outer["CARGO_HOME"])
            self.assertEqual(event["rustup_home"], self.outer["RUSTUP_HOME"])
        self.assertEqual({tuple(event["args"]) for event in self.events(self.alias_log)},
                         {("which", "cargo"), ("which", "rustc")})

    def test_ordinary_rustup_proxy_resolves_executable_native_tools(self):
        self.aliases("rustup")
        self.assert_native_selection()

    def test_homebrew_installer_alias_resolves_executable_native_tools(self):
        self.aliases()
        self.assert_native_selection()

    def test_hardlinked_proxies_use_samefile_manager_alias(self):
        self.aliases("proxy-payload", hardlink=True)
        self.assert_native_selection()

    def test_native_path_tools_take_precedence_over_unrelated_manager(self):
        alternate = self.root / "unrelated-toolchain"
        alternate.mkdir()
        for name in ("cargo", "rustc"):
            self.executable(alternate / name,
                            'import sys\n'
                            'sys.stderr.write("FIXTURE_UNRELATED_TOOLCHAIN\\n")\n'
                            'sys.exit(66)\n')
        self.lookup.write_text(json.dumps({name: str(alternate / name)
                                          for name in ("cargo", "rustc")}))
        for name in ("cargo", "rustc"):
            (self.bin / name).symlink_to(self.native / name)
        unrelated = self.proxy / "rustup-init"
        self.executable(unrelated, PROXY)
        (self.bin / "rustup").symlink_to(unrelated)
        self.addCleanup(self.assert_no_installer)
        for name in ("cargo", "rustc"):
            with self.subTest(tool=name):
                selected = compiler.native_tool(name, self.outer)
                # Exercise the consumer first: an unrelated lookup result fails
                # the file-producing protocol, not merely a dispatch-log check.
                self.assert_command(name, selected)
                self.assertEqual(selected, self.native / name)
        self.assertEqual(self.events(self.alias_log), [], "unrelated manager overrode native PATH tools")

    def test_proxy_without_manager_reports_clear_native_prerequisite_error(self):
        self.aliases(manager=False)
        self.addCleanup(self.assert_no_installer)
        for name in ("cargo", "rustc"):
            with self.subTest(tool=name):
                with self.assertRaises(AssertionError) as error:
                    compiler.native_tool(name, self.outer)
                self.assertIn(name, str(error.exception))
                self.assertRegex(str(error.exception), r"(?i)native|prerequisite")
        self.assertEqual(self.events(self.alias_log), [])
        self.assertEqual(self.events(self.native_log), [])

    def test_manager_returning_proxy_or_unusable_path_reports_clear_error(self):
        installer = self.aliases()
        self.addCleanup(self.assert_no_installer)
        nonexecutable = self.native / "not-executable"
        nonexecutable.write_text("not a tool\n")
        nonexecutable.chmod(0o644)
        candidates = {"proxy": installer, "missing": self.native / "missing",
                      "directory": self.native, "nonexecutable": nonexecutable}
        for label, candidate in candidates.items():
            self.lookup.write_text(json.dumps({name: str(candidate) for name in ("cargo", "rustc")}))
            for name in ("cargo", "rustc"):
                with self.subTest(returned=label, tool=name):
                    with self.assertRaises(AssertionError) as error:
                        compiler.native_tool(name, self.outer)
                    self.assertIn(name, str(error.exception))
                    self.assertRegex(str(error.exception), r"(?i)native|prerequisite")
        self.assertEqual(self.events(self.native_log), [])


if __name__ == "__main__":
    unittest.main()
