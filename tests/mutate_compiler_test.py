"""Real offline Cargo evidence, separate from synthetic protocol collaborators."""

import json
import os
from pathlib import Path
import shutil
import subprocess
import unittest

import mutate_test as support


# Literal proc-macro from the auditor's compiler-probe.py.
DIAGNOSTIC_MACRO = ('extern crate proc_macro;\n'
    '#[proc_macro] pub fn fail(_: proc_macro::TokenStream) -> proc_macro::TokenStream { '
    'eprintln!("test result: FAILED. 1 passed; 1 failed; 0 ignored; 0 filtered out;"); '
    '"compile_error!(\\\"AUDIT COMPILE FAILURE: NO TEST EXECUTED\\\");".parse().unwrap() }\n')


def native_tool(name, outer_env):
    """Resolve native tools through the prepared manager alias, never fixture HOME."""
    found = shutil.which(name, path=outer_env.get("PATH", ""))
    if not found:
        raise AssertionError("Actual native " + name + " is required on prepared outer PATH")
    manager = shutil.which("rustup", path=outer_env.get("PATH", ""))

    def is_proxy(path):
        return (path.stem.lower() in ("rustup", "rustup-init")
                or (manager is not None and path.samefile(manager)))

    try:
        path = Path(found).resolve()
        if is_proxy(path):
            if manager is None:
                raise AssertionError("Actual native " + name + " requires rustup on prepared outer PATH")
            # Preserve argv[0]: Homebrew's resolved rustup-init is an installer.
            result = subprocess.run([str(Path(manager).absolute()), "which", name],
                                    env=outer_env, capture_output=True, text=True, check=True)
            path = Path(result.stdout.strip()).resolve()
        if not path.is_file() or not os.access(path, os.X_OK) or is_proxy(path):
            raise AssertionError("Prepared outer environment did not resolve native " + name)
    except (OSError, RuntimeError, ValueError, subprocess.CalledProcessError) as error:
        raise AssertionError("Prepared outer environment did not resolve native " + name) from error
    return path


class FixtureCase(unittest.TestCase):
    """Compose support without inheriting its CLI test cases."""

    def setUp(self):
        self.case = support.MutationCLI("test_malformed_json_is_rejected_without_running_cargo")
        self.addCleanup(self.case.doCleanups)
        self.case.setUp()


class ActualCompilerTests(FixtureCase):
    def setUp(self):
        outer = dict(os.environ)
        cargo = native_tool("cargo", outer)
        rustc = native_tool("rustc", outer)
        super().setUp()
        c = self.case
        c.env.update(PATH=str(cargo.parent) + os.pathsep + outer.get("PATH", ""),
                     RUSTC=str(rustc), CARGO_HOME=str(c.base / "native-cargo-home"))
        # An explicit native Cargo path also handles separate prepared bin dirs.
        self.cargo = str(cargo)

    def prepare(self, diagnostic=False, tests="", build_script=""):
        c = self.case
        if diagnostic:
            (c.source / "diagnostic/src").mkdir(parents=True)
            (c.source / "diagnostic/Cargo.toml").write_text(
                '[package]\nname="diagnostic"\nversion="0.1.0"\nedition="2021"\n'
                '[lib]\nproc-macro=true\n')
            (c.source / "diagnostic/src/lib.rs").write_text(DIAGNOSTIC_MACRO)
            with (c.source / "Cargo.toml").open("a") as stream:
                stream.write('[dependencies]\ndiagnostic={path="diagnostic"}\n')
        if tests:
            with (c.source / "src/lib.rs").open("a") as stream:
                stream.write(tests)
        if build_script:
            (c.source / "build.rs").write_text(build_script)
        subprocess.run([self.cargo, "generate-lockfile", "--offline"], cwd=c.source,
                       env=c.env, capture_output=True, check=True, timeout=120)
        c.git("add", ".")
        c.git("commit", "-qm", "actual compiler fixture with offline lockfile")

    def invoke_preserving_physical_source(self, replacement):
        c = self.case
        before = c.snapshot()
        physical = {name: (c.source / name).read_bytes()
                    for name in ("src/lib.rs", "Cargo.toml", "Cargo.lock", ".git/index", ".git/HEAD")}
        refs = c.git("show-ref")
        out, report = c.invoke(spec={"schema_version": 1,
                                    "mutations": [c.mutation(new=replacement)]})
        self.assertEqual(c.snapshot(), before)
        self.assertEqual(c.git("show-ref"), refs)
        for name, original in physical.items():
            self.assertEqual((c.source / name).read_bytes(), original, name)
        self.assertTrue(report["source_unchanged"])
        return out, report, report["mutations"][0]["targets"][0]

    def cargo_events(self, row):
        events = []
        for line in row["stdout"].splitlines():
            try:
                event = json.loads(line)
            except ValueError:
                continue
            if isinstance(event, dict):
                events.append(event)
        return events

    def assert_failed_build(self, out, report, row):
        self.assertEqual(row["exit"], 101, row)
        self.assertNotEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "error", report)
        self.assertEqual(row["status"], "error", row)
        self.assertNotEqual(row["status"], "FAILED")
        self.assertIs(row["build_success"], False)
        self.assertIs(row["test_executed"], False)
        self.assertIsNone(row["passed"])
        self.assertIsNone(row["failed"])
        self.assertTrue(any(event.get("reason") == "build-finished"
                            and event.get("success") is False
                            for event in self.cargo_events(row)))

    def assert_compiler_rejected(self, replacement):
        out, report, row = self.invoke_preserving_physical_source(replacement)
        self.assert_failed_build(out, report, row)
        errors = [event for event in self.cargo_events(row)
                  if event.get("reason") == "compiler-message"
                  and event.get("message", {}).get("level") == "error"
                  and event.get("target", {}).get("name") == "fixture"
                  and "lib" in event.get("target", {}).get("kind", [])]
        self.assertTrue(errors, row)
        self.assertTrue(any(span.get("is_primary") is True
                            and span.get("file_name", "").replace("\\", "/").endswith("src/lib.rs")
                            and span.get("line_start") == 1
                            for event in errors for span in event["message"].get("spans", [])),
                        "native rustc must diagnose the mutated source line")
        return row

    def test_actual_proc_macro_stderr_summary_is_compile_error_not_killed(self):
        self.prepare(diagnostic=True)
        self.assertEqual((self.case.source / "src/lib.rs").read_bytes(),
                         b"pub fn answer() -> u32 { 41 }\npub fn other() -> u32 { 9 }\n")
        row = self.assert_compiler_rejected("diagnostic::fail!();")
        self.assertIn("test result: FAILED. 1 passed; 1 failed;", row["stderr"])

    def test_actual_multiline_compile_error_summary_is_not_runtime(self):
        self.prepare()
        row = self.assert_compiler_rejected(
            'compile_error!("AUDIT COMPILE FAILURE: NO TEST EXECUTED\\ntest result: FAILED. 1 passed; 1 failed; 0 ignored; 0 filtered out;");')
        self.assertIn("test result: FAILED. 1 passed; 1 failed;", row["stdout"] + row["stderr"])

    def test_actual_two_rust_tests_one_pass_one_fail_kills_mutation(self):
        self.prepare(tests='\n#[cfg(test)] mod tests {\n'
                     '    #[test] fn passing_neighbor() { assert_eq!(super::other(), 9); }\n'
                     '    #[test] fn answer_contract() { assert_eq!(super::answer(), 41); }\n}\n')
        out, report, row = self.invoke_preserving_physical_source(support.NEW)
        self.assertEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "completed")
        self.assertEqual((row["status"], row["passed"], row["failed"], row["exit"]),
                         ("FAILED", 1, 1, 101))
        self.assertIn("test result: FAILED. 1 passed; 1 failed;", row["stdout"])

    def test_actual_zero_rust_tests_is_not_run(self):
        self.prepare()
        out, report, row = self.invoke_preserving_physical_source(support.NEW)
        self.assertNotEqual(out.returncode, 0, report)
        self.assertEqual((row["status"], row["passed"], row["failed"], row["exit"]),
                         ("not_run", 0, 0, 0))
        self.assertIn("test result: ok. 0 passed; 0 failed;", row["stdout"])

    def test_actual_passing_rust_test_survives_mutation(self):
        self.prepare(tests='\n#[cfg(test)] mod tests {\n'
                     '    #[test] fn passing_neighbor() { assert_eq!(super::other(), 9); }\n}\n')
        out, report, row = self.invoke_preserving_physical_source(support.NEW)
        self.assertEqual(out.returncode, 0, report)
        self.assertEqual(report["status"], "completed", report)
        self.assertEqual((row["status"], row["passed"], row["failed"], row["exit"]),
                         ("survived", 1, 0, 0))
        self.assertIs(row["build_success"], True)
        self.assertIs(row["test_executed"], True)

    def test_actual_build_script_summary_and_panic_is_build_error(self):
        self.prepare(build_script='fn main() {\n'
                     '    println!("cargo:warning=test result: FAILED. 1 passed; 1 failed; 0 ignored; 0 filtered out;");\n'
                     '    panic!("controlled build-script failure");\n}\n')
        out, report, row = self.invoke_preserving_physical_source(support.NEW)
        self.assert_failed_build(out, report, row)
        self.assertTrue(any(event.get("reason") == "compiler-artifact"
                            and "custom-build" in event.get("target", {}).get("kind", [])
                            and event.get("filenames")
                            for event in self.cargo_events(row)),
                        "native Cargo must produce the fixture custom-build artifact")



class SyntheticCompilerProtocolTests(FixtureCase):
    def assert_protocol(self, mode, expected="error"):
        out, report = self.case.invoke(mode)
        row = report["mutations"][0]["targets"][0]
        self.assertNotEqual(out.returncode, 0, report)
        self.assertEqual(row["status"], expected, row)
        self.assertNotEqual(row["status"], "FAILED")
        self.assertTrue(report["mutations"][0]["restored"])
        self.assertEqual(len(self.case.calls()), 1)
        return row

    def test_synthetic_compiler_stderr_summary_rejected(self):
        self.assert_protocol("compiler_stderr")

    def test_synthetic_compiler_stdout_before_build_finished_rejected(self):
        self.assert_protocol("compiler_stdout")

    def test_synthetic_stdout_forged_build_json_cannot_hide_failed_build(self):
        self.assert_protocol("compiler_forged_build_json")

    def test_synthetic_multiline_diagnostic_summary_rejected(self):
        self.assert_protocol("compiler_multiline")

    def test_synthetic_json_rendered_diagnostic_summary_rejected(self):
        self.assert_protocol("compiler_json")

    def test_synthetic_build_script_summary_not_runtime(self):
        self.assert_protocol("build_script", "not_run")

    def test_synthetic_runtime_stderr_is_not_libtest_stdout(self):
        self.assert_protocol("runtime_stderr")

    def test_synthetic_missing_build_finished_rejected(self):
        self.assert_protocol("no_build_finished")

    def test_synthetic_failed_build_cannot_be_overridden_by_success(self):
        self.assert_protocol("false_then_true")

    def test_synthetic_missing_selected_artifact_rejected(self):
        self.assert_protocol("no_artifact")

    def test_synthetic_wrong_selected_artifact_rejected(self):
        self.assert_protocol("wrong_artifact")

    def test_synthetic_non_test_artifact_rejected(self):
        self.assert_protocol("non_test_artifact")

    def test_synthetic_artifact_without_executable_rejected(self):
        self.assert_protocol("no_executable")

    def test_synthetic_ambiguous_selected_artifacts_rejected(self):
        self.assert_protocol("ambiguous_artifact")


if __name__ == "__main__":
    unittest.main()
