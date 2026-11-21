"""The golden harness runs its checks on a bounded number of workers (tools/statusline_golden.py).

The binary and the cases are stand-ins: a temporary fixture folder holds a small manifest and its
goldens, and `run_case` is replaced by a function that returns each case's text and records how
many calls were running at once. Nothing here starts the real binary.
"""
import contextlib
import importlib.util
import io
import json
import os
import shutil
import sys
import tempfile
import threading
import time
import unittest
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
TOOL = REPO / "tools" / "statusline_golden.py"
CASES = [f"case-{n}" for n in range(1, 9)]
VARIANTS = ["plain", "ansi", "ansi256"]


def load_tool():
    spec = importlib.util.spec_from_file_location("statusline_golden_under_test", TOOL)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class Fixture:
    """A fixture folder, a stand-in binary and a fake `run_case` on a loaded harness module."""

    def __init__(self, test, pause=0.01):
        self.tool = load_tool()
        self.tmp = Path(tempfile.mkdtemp(prefix="golden-parallel-"))
        test.addCleanup(shutil.rmtree, self.tmp, ignore_errors=True)
        fixtures = self.tmp / "fixtures"
        (fixtures / "golden").mkdir(parents=True)
        recorded = {v: {"version": "1.0.0", "sha256": "x", "stdout": "pipe"} for v in VARIANTS}
        manifest = {"cases": [{"name": c, "stdin": "unused.json",
                               "recorded_with": {"variants": recorded}} for c in CASES]}
        (fixtures / "cases.json").write_text(json.dumps(manifest), encoding="utf-8")
        for c in CASES:
            for v in VARIANTS:
                (fixtures / "golden" / f"{c}.{v}.txt").write_text(self.text(c, v), encoding="utf-8")
        self.binary = self.tmp / "stand-in"
        self.binary.write_bytes(b"not a program")
        self.tool.FIXTURES = fixtures
        self.pause = pause
        self.lock = threading.Lock()
        self.active = 0
        self.peak = 0
        self.runs = 0
        self.contract_runs = 0
        self.changed = {}
        self.tool.run_case = self.run_case
        self.tool.check_color_contract = lambda binary, argv: self.contract("color")
        self.tool.check_input_boundaries = lambda binary, argv: self.contract("boundaries")

    @staticmethod
    def text(case, variant):
        return f"🗿 v1.0.0 {case} {variant}\n"

    def run_case(self, binary, argv, case, variant, stdin_bytes=None, stdout_tty=None):
        with self.lock:
            self.active += 1
            self.peak = max(self.peak, self.active)
            self.runs += 1
        time.sleep(self.pause)
        with self.lock:
            self.active -= 1
        # The stand-in output is the golden's text, or another text where a test changed it.
        return 0, self.changed.get((case["name"], variant), self.text(case["name"], variant)), ""

    def contract(self, name):
        with self.lock:
            self.contract_runs += 1
        return f"ok   native {name} contract"

    def run_main(self, *args, env=None):
        argv = ["statusline_golden.py", "--bin", str(self.binary), "--binary-version", "1.0.0",
                "--check", *args]
        out = io.StringIO()
        old_env = dict(os.environ)
        os.environ.pop("MOAI_SPAWN_TEST_CAP", None)
        os.environ.update(env or {})
        try:
            with mock_argv(argv), contextlib.redirect_stdout(out):
                code = self.tool.main()
        finally:
            os.environ.clear()
            os.environ.update(old_env)
        return code, out.getvalue()


@contextlib.contextmanager
def mock_argv(argv):
    old = sys.argv
    sys.argv = argv
    try:
        yield
    finally:
        sys.argv = old


class ParallelGoldenTest(unittest.TestCase):
    def test_jobs_four_and_one_print_the_same_lines_in_manifest_order(self):
        serial = Fixture(self)
        code_one, out_one = serial.run_main("--jobs", "1")
        parallel = Fixture(self)
        code_four, out_four = parallel.run_main("--jobs", "4")
        self.assertEqual((code_one, code_four), (0, 0), out_four)
        # The binary path differs per fixture; every other line, in order, is the same.
        keep = lambda out: [l for l in out.splitlines() if not l.startswith("binary sha256")]
        self.assertEqual(keep(out_one), keep(out_four))
        lines = [l for l in out_four.splitlines() if l.startswith("ok   case-")]
        want = [f"ok   {c} {v}" for c in CASES for v in VARIANTS]
        self.assertEqual(lines, want)
        self.assertIn(f"checked {len(CASES) * len(VARIANTS)} golden outputs", out_four)

    def test_the_workers_are_bounded_by_jobs(self):
        one = Fixture(self)
        one.run_main("--jobs", "1")
        self.assertEqual(one.peak, 1)
        four = Fixture(self)
        four.run_main("--jobs", "4")
        # Four workers ran at once at some moment, never more: the cap holds and it is used.
        self.assertEqual(four.peak, 4)

    def test_a_different_golden_fails_the_same_way_serial_and_parallel(self):
        results = []
        for jobs in ("1", "4"):
            fx = Fixture(self)
            fx.changed[("case-3", "ansi")] = "🗿 v1.0.0 case-3 ansi CHANGED\n"
            results.append(fx.run_main("--jobs", jobs))
        (code_one, out_one), (code_four, out_four) = results
        self.assertEqual((code_one, code_four), (1, 1))
        for out in (out_one, out_four):
            self.assertIn("FAIL case-3 ansi: golden differs", out)
            self.assertEqual(out.count("FAIL "), 1, out)
        self.assertEqual(
            [l for l in out_one.splitlines() if not l.startswith("binary sha256")],
            [l for l in out_four.splitlines() if not l.startswith("binary sha256")])

    def test_the_runtime_contracts_run_once_whatever_jobs_is(self):
        for jobs in ("1", "4", "16"):
            fx = Fixture(self)
            code, out = fx.run_main("--jobs", jobs)
            self.assertEqual(code, 0, out)
            self.assertEqual(fx.contract_runs, 2, f"--jobs {jobs}: color and boundaries once each")
            self.assertEqual(out.count("native color contract"), 1)
            self.assertEqual(out.count("native boundaries contract"), 1)

    def test_a_failing_contract_is_one_fail_line_and_exit_one(self):
        fx = Fixture(self)

        def broken(binary, argv):
            raise ValueError("plain pipe failed")

        fx.tool.check_color_contract = broken
        code, out = fx.run_main("--jobs", "4")
        self.assertEqual(code, 1)
        self.assertIn("FAIL native runtime contract: plain pipe failed", out)
        self.assertEqual(fx.contract_runs, 0, "a failed color contract still skips the second")

    def test_without_jobs_the_spawn_cap_decides_and_four_is_the_fallback(self):
        unset = Fixture(self)
        unset.run_main()
        self.assertEqual(unset.peak, 4)
        capped = Fixture(self)
        capped.run_main(env={"MOAI_SPAWN_TEST_CAP": "2"})
        self.assertEqual(capped.peak, 2)
        # An explicit --jobs wins over the environment.
        explicit = Fixture(self)
        explicit.run_main("--jobs", "3", env={"MOAI_SPAWN_TEST_CAP": "2"})
        self.assertEqual(explicit.peak, 3)

    def test_a_bad_jobs_value_is_refused_before_any_check_runs(self):
        for value in ("0", "abc", "-1", "2.5", ""):
            fx = Fixture(self)
            code, out = fx.run_main(f"--jobs={value}")
            self.assertEqual(code, 2, f"--jobs={value!r}: {out}")
            self.assertIn("--jobs must be a whole number of at least 1", out)
            self.assertEqual((fx.runs, fx.contract_runs), (0, 0), f"--jobs={value!r} ran something")
        fx = Fixture(self)
        code, out = fx.run_main(env={"MOAI_SPAWN_TEST_CAP": "0"})
        self.assertEqual(code, 2, out)
        self.assertIn("MOAI_SPAWN_TEST_CAP must be a whole number of at least 1", out)
        self.assertEqual((fx.runs, fx.contract_runs), (0, 0))

    def test_resolve_jobs_edges(self):
        tool = load_tool()
        self.assertEqual(tool.resolve_jobs("1", {}), 1)
        self.assertEqual(tool.resolve_jobs("64", {}), 64)
        self.assertEqual(tool.resolve_jobs(None, {}), 4)
        self.assertEqual(tool.resolve_jobs(None, {"MOAI_SPAWN_TEST_CAP": "7"}), 7)
        with self.assertRaises(ValueError):
            tool.resolve_jobs("0", {})


if __name__ == "__main__":
    unittest.main()
