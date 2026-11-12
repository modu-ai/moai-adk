"""Handshake-driven deadline contracts: no wall-clock startup assumptions."""

from contextlib import redirect_stdout
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

import mutate_test as support


ROOT = Path(__file__).resolve().parent.parent
spec = importlib.util.spec_from_file_location("mutation_deadline_sut", ROOT / "tools/mutate.py")
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)

# This real child waits for ordinary IPC, not elapsed time or an OS kill.
HELPER = r'''
import json, sys
print("READY", flush=True)
assert sys.stdin.buffer.read(1) == b"X"
print(json.dumps({"reason": "compiler-artifact", "package_id": "fixture 0.1.0",
    "manifest_path": "Cargo.toml", "target": {"name": "fixture", "kind": ["lib"],
    "crate_types": ["lib"]}, "profile": {"test": True},
    "executable": "fixture", "filenames": [], "fresh": False}))
print(json.dumps({"reason": "build-finished", "success": True}))
print("test result: FAILED. 1 passed; 1 failed; 0 ignored; 0 filtered out;", flush=True)
sys.exit(101)
'''


class Clock:
    def __init__(self):
        self.value = 100.0

    def __call__(self):
        return self.value


class ProcessProxy:
    def __init__(self, process, case):
        self.process = process
        self.case = case
        self.calls = []
        self.released = False
        self.ready = False

    def __getattr__(self, name):
        return getattr(self.process, name)

    def release(self):
        if not self.released:
            self.case.assertTrue(self.ready, "cleanup must follow the READY handshake")
            self.process.stdin.write(b"X")
            self.process.stdin.flush()
            self.process.stdin.close()
            # Popen.communicate must not flush a closed stream.
            self.process.stdin = None
            self.released = True

    def communicate(self, *args, **kwargs):
        timeout = kwargs.get("timeout")
        self.calls.append(timeout)
        self.case.events.append(("communicate", timeout))
        self.case.assertTrue(self.ready)
        if self.case.expired_before_communicate:
            self.case.assertTrue(self.released,
                                 "expired deadline must stop before even calling communicate")
            self.case.assertIsNone(timeout)
        if timeout is not None and self.case.communication_error:
            raise OSError("controlled communicate failure")
        if timeout is not None and self.case.communicate_times_out:
            raise subprocess.TimeoutExpired(self.process.args, timeout)
        self.release()
        # The requested budget belongs to the injected clock, not OS wall time.
        # Timeout/error outcomes above are controlled; IPC drain is functional.
        return self.process.communicate()


class DeadlineHandshakeTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="mutation-deadline-")
        self.addCleanup(self.temp.cleanup)
        self.workspace = Path(self.temp.name)
        self.clock = Clock()
        self.events = []
        self.processes = []
        self.handshake_elapsed = 0
        self.spawn_elapsed = 0
        self.expired_before_communicate = False
        self.communicate_times_out = False
        self.communication_error = False
        self.refuse_cleanup = False
        self.addCleanup(self.cleanup_helpers)

    def cleanup_helpers(self):
        for proxy in self.processes:
            # Even a controlled cleanup-refusal test disposes of its own helper
            # through IPC. No denied host kill is retried or bypassed.
            if not proxy.ready:
                self.assertEqual(proxy.stdout.readline(), b"READY\n")
                proxy.ready = True
            proxy.release()
            proxy.process.communicate(timeout=30)
            self.assertIsNotNone(proxy.poll())

    def spawn(self, argv, **kwargs):
        self.events.append(("spawn", self.clock.value))
        kwargs["stdin"] = subprocess.PIPE
        process = subprocess.Popen(argv, **kwargs)
        proxy = ProcessProxy(process, self)
        self.processes.append(proxy)
        self.clock.value += self.spawn_elapsed
        return proxy

    def started(self, process):
        self.assertEqual(process.stdout.readline(), b"READY\n")
        process.ready = True
        self.events.append(("ready", self.clock.value))
        self.clock.value += self.handshake_elapsed

    def stop(self, process, force=False):
        self.events.append(("stop", force))
        if self.refuse_cleanup:
            raise runner.UnsafeProcess("controlled cleanup refusal")
        process.release()
        process.wait(timeout=30)

    def invoke(self, timeout=10, spawn=None):
        row, interrupted = runner.run_target(
            [sys.executable, "-I", "-c", HELPER, "--lib"], self.workspace,
            dict(os.environ), timeout, clock=self.clock, started=self.started,
            spawn=self.spawn if spawn is None else spawn, stop=self.stop)
        self.assertFalse(interrupted)
        self.assertIs(type(row["timed_out"]), bool)
        return row

    def assert_expired(self, elapsed):
        self.handshake_elapsed = elapsed
        self.expired_before_communicate = True
        row = self.invoke()
        self.assertTrue(row["timed_out"], row)
        self.assertEqual(row["status"], "error")
        self.assertIn("timed out", row["error"])
        self.assertIn(("stop", True), self.events)
        self.assertTrue(self.processes[0].released)
        self.assertTrue(all(timeout is None for timeout in self.processes[0].calls))
        # A summary produced during cleanup is not evidence of an executed target.
        self.assertNotEqual(row["status"], "FAILED")
        return row

    def test_expired_deadline_stops_before_communicate(self):
        self.assert_expired(11)

    def test_exact_deadline_is_expired_before_communicate(self):
        self.assert_expired(10)

    def test_not_expired_uses_remaining_budget(self):
        self.handshake_elapsed = 2
        row = self.invoke()
        self.assertFalse(row["timed_out"], row)
        self.assertEqual(self.processes[0].calls, [8])
        self.assertEqual((row["status"], row["passed"], row["failed"]), ("FAILED", 1, 1))
        self.assertTrue(any(event[0] == "stop" for event in self.events))

    def test_slow_ready_startup_expires_without_wall_clock_assumption(self):
        self.assert_expired(10000)
        self.assertEqual([event[0] for event in self.events[:2]], ["spawn", "ready"])

    def test_handshake_does_not_reset_original_deadline(self):
        self.handshake_elapsed = 7
        row = self.invoke()
        self.assertFalse(row["timed_out"])
        self.assertEqual(self.processes[0].calls, [3])

    def test_deadline_begins_after_owned_process_creation(self):
        self.spawn_elapsed = 10000
        self.handshake_elapsed = 2
        row = self.invoke()
        self.assertFalse(row["timed_out"])
        self.assertEqual(self.processes[0].calls, [8])

    def test_communicate_timeout_sets_observable_timed_out(self):
        self.handshake_elapsed = 2
        self.communicate_times_out = True
        row = self.invoke()
        self.assertTrue(row["timed_out"], row)
        self.assertEqual(row["status"], "error")
        self.assertEqual(self.processes[0].calls[0], 8)
        self.assertIn(("stop", True), self.events)

    def test_expired_deadline_cleanup_refusal_is_fail_closed(self):
        self.handshake_elapsed = 11
        self.expired_before_communicate = True
        self.refuse_cleanup = True
        row = self.invoke()
        self.assertTrue(row["timed_out"], row)
        self.assertEqual(row["status"], "error")
        self.assertIs(row["quiescent"], False)
        self.assertIn("controlled cleanup refusal", row["error"])
        self.assertEqual(self.processes[0].calls, [])
        self.assertFalse(self.processes[0].released)

    def test_communicate_exception_is_error_not_timeout(self):
        self.communication_error = True
        row = self.invoke()
        self.assertFalse(row["timed_out"], row)
        self.assertEqual(row["status"], "error")
        self.assertIn("controlled communicate failure", row["error"])

    def test_process_creation_error_is_not_timeout_and_has_no_child(self):
        def fail_spawn(*args, **kwargs):
            raise OSError("controlled creation failure")
        row = self.invoke(spawn=fail_spawn)
        self.assertFalse(row["timed_out"], row)
        self.assertEqual(row["status"], "error")
        self.assertIn("controlled creation failure", row["error"])
        self.assertEqual(self.processes, [])
        self.assertEqual(self.events, [])

    def test_actual_cli_original_point_two_timeout_preserves_source_and_restores_clone(self):
        fixture = support.MutationCLI("test_malformed_json_is_rejected_without_running_cargo")
        self.addCleanup(fixture.doCleanups)
        fixture.setUp()
        fixture.spec["timeout_seconds"] = 0.2
        spec_path = fixture.base / "spec.json"
        spec_path.write_text(json.dumps(fixture.spec))
        spec_bytes = spec_path.read_bytes()
        # Keep the auditor's Cargo command and source fixture, replacing only
        # its owned Cargo collaborator with the existing READY/IPC helper.
        (fixture.bin / "cargo").write_text("#!" + sys.executable + "\n" + HELPER)
        original_bytes = b"pub fn answer() -> u32 { 41 }\npub fn other() -> u32 { 9 }\n"
        self.assertEqual((fixture.source / "src/lib.rs").read_bytes(), original_bytes)
        before = fixture.snapshot()
        index = (fixture.source / ".git/index").read_bytes()
        head = (fixture.source / ".git/HEAD").read_bytes()
        refs = fixture.git("show-ref")
        self.handshake_elapsed = 0.2
        self.expired_before_communicate = True
        actual_run_target = runner.run_target
        observed = []

        def inject_private_seams(argv, workspace, env, timeout):
            self.assertEqual(timeout, 0.2, "the original parsed deadline must not be inflated")
            self.assertEqual((workspace / "src/lib.rs").read_text(),
                             support.NEW + "\npub fn other() -> u32 { 9 }\n")
            observed.append((list(argv), workspace, timeout))
            return actual_run_target(argv, workspace, env, timeout, clock=self.clock,
                                     started=self.started, spawn=self.spawn, stop=self.stop)

        capture = io.StringIO()
        previous_cwd = Path.cwd()
        try:
            os.chdir(fixture.source)
            with patch.dict(os.environ, dict(fixture.env, FAKE_MODE="timeout"), clear=True), \
                    patch.object(sys, "argv", ["mutate.py", "--spec", str(spec_path)]), \
                    patch.object(runner, "run_target", side_effect=inject_private_seams), \
                    redirect_stdout(capture):
                code = runner.main()
        finally:
            os.chdir(previous_cwd)
        report = json.loads(capture.getvalue())
        self.assertEqual(code, 1, report)
        self.assertEqual(report["status"], "error", report)
        self.assertIs(report["source_unchanged"], True)
        self.assertEqual(len(observed), 1)
        argv, workspace, timeout = observed[0]
        self.assertEqual(timeout, 0.2)
        mutation = report["mutations"][0]
        self.assertIs(mutation["restored"], True)
        self.assertEqual(mutation["restore_diff_exit"], 0)
        self.assertEqual(len(mutation["targets"]), 1)
        target = mutation["targets"][0]
        self.assertIs(target["timed_out"], True)
        self.assertEqual(target["status"], "error")
        self.assertNotEqual(target["status"], "FAILED")
        self.assertIn(("stop", True), self.events)
        self.assertTrue(self.processes[0].released)
        self.assertTrue(all(value is None for value in self.processes[0].calls))
        self.assertEqual(Path(report["workspace"]), workspace)
        self.assertEqual((workspace / "src/lib.rs").read_bytes(), original_bytes)
        self.assertEqual(fixture.git("diff", "HEAD", cwd=workspace), b"")
        self.assertEqual(fixture.snapshot(), before)
        self.assertEqual((fixture.source / "src/lib.rs").read_bytes(), original_bytes)
        self.assertEqual((fixture.source / ".git/index").read_bytes(), index)
        self.assertEqual((fixture.source / ".git/HEAD").read_bytes(), head)
        self.assertEqual(fixture.git("show-ref"), refs)
        self.assertEqual(spec_path.read_bytes(), spec_bytes)
        self.assertEqual(json.loads(spec_bytes)["timeout_seconds"], 0.2)
        self.assertEqual(fixture.sentinel.read_bytes(), b"external bytes must survive\n")


if __name__ == "__main__":
    unittest.main()
