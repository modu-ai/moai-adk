"""The load probe (tools/mflow_probe.py, SPEC-20261120-mflow-measure).

The stubs keep every assertion off the wall clock: a fake target that exits with a
chosen code after a sleep, and one that kills itself with SIGTERM. The macOS sources
(`sysctl`, `vm_stat`, `ps`) are read by the probe itself; these tests only assert the
JSON shape the instruction names.
"""
import json
import os
import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
PROBE = ROOT / "tools" / "mflow_probe.py"


def run_probe(args, stub=None):
    command = [sys.executable, "-I", str(PROBE)] + list(args)
    if stub is not None:
        command += ["--", sys.executable, "-I", "-c", stub]
    return subprocess.run(command, capture_output=True, text=True, cwd=ROOT)


class MflowProbeTest(unittest.TestCase):
    def probe(self, *args, **kw):
        out = run_probe(list(args), **kw)
        self.assertEqual(out.returncode, 0, out.stderr or out.stdout)
        return json.loads(out.stdout)

    def test_exit_code_elapsed_samples_and_peak_come_back_in_the_answer(self):
        """REQ-1: the answer names the ending (exit 3), the elapsed wall time and the
        sampled load (median, max) and peak compile/test process count — asserted from
        the JSON, never from time."""
        answer = self.probe("--interval", "0.2", "--label", "stub", "--base", "a" * 40,
                            "--head", "b" * 40,
                            stub="import time; time.sleep(0.7); raise SystemExit(3)")
        self.assertEqual(answer["ended"]["kind"], "exit")
        self.assertEqual(answer["ended"]["code"], 3)
        self.assertGreater(answer["elapsed_seconds"], 0.5)
        self.assertGreaterEqual(answer["load"]["samples"], 1)
        self.assertIsInstance(answer["load"]["median"], float)
        self.assertIsInstance(answer["load"]["max"], float)
        self.assertIsInstance(answer["peak_processes"], int)
        self.assertIsInstance(answer["cpu"]["user_percent"], float)
        self.assertIsInstance(answer["cpu"]["system_percent"], float)
        self.assertIsInstance(answer["cpu"]["utilisation_percent"], float)

    def test_the_answer_names_the_run_and_the_machine_facts(self):
        """REQ-2: label, base, head, UTC and KST stamps, core count and the git/rust/CLI
        versions ride every answer."""
        answer = self.probe("--label", "before-1", "--base", "c" * 40, "--head", "d" * 40,
                            stub="pass")
        self.assertEqual(answer["label"], "before-1")
        self.assertEqual(answer["base"], "c" * 40)
        self.assertEqual(answer["head"], "d" * 40)
        self.assertIn("T", answer["utc"])
        self.assertTrue(answer["kst"].endswith("+09:00"), answer["kst"])
        self.assertGreater(answer["machine"]["ncpu"], 0)
        self.assertTrue(answer["versions"]["git"].startswith("git version"))
        self.assertTrue(answer["versions"]["rustc"].startswith("rustc "))
        self.assertIn("rust", answer["versions"]["cli"])

    def test_the_window_field_marks_an_above_threshold_run(self):
        """REQ-3: a threshold below the current load marks the run `above` (with the
        observed load named); a threshold nothing reaches marks it `ok`."""
        above = self.probe("--max-load", "0", "--label", "a", "--base", "a" * 40,
                           "--head", "b" * 40, stub="pass")
        self.assertEqual(above["window"], "above")
        self.assertGreaterEqual(above["window_observed_load"], 0)
        ok = self.probe("--max-load", "100000", "--label", "b", "--base", "a" * 40,
                        "--head", "b" * 40, stub="pass")
        self.assertEqual(ok["window"], "ok")

    def test_a_sigterm_ending_is_its_own_field_and_the_out_document_holds_the_series(self):
        """REQ-1 and REQ-4: a SIGTERM ending is `ended.kind == "signal"` — separate from
        any exit code — and the --out document holds the sample series (the summary's
        sample count equals the series length)."""
        with tempfile.TemporaryDirectory(prefix="mflow-probe-") as tmp:
            out_path = os.path.join(tmp, "run.json")
            answer = self.probe("--interval", "0.2", "--out", out_path, "--label", "sig",
                                "--base", "a" * 40, "--head", "b" * 40,
                                stub="import os, signal; os.kill(os.getpid(), signal.SIGTERM)")
            self.assertEqual(answer["ended"]["kind"], "signal")
            self.assertEqual(answer["ended"]["signal"], "SIGTERM")
            self.assertNotIn("code", answer["ended"])
            document = json.loads(pathlib.Path(out_path).read_text(encoding="utf-8"))
            series = document["series"]
            self.assertEqual(document["answer"]["load"]["samples"], len(series["load"]))
            self.assertEqual(len(series["load"]), len(series["swap_used_bytes"]))
            self.assertEqual(len(series["load"]), len(series["cpu_pcpu"]))
            self.assertEqual(len(series["load"]), len(series["compile_test_processes"]))

    def test_usage_without_a_target_command_is_an_error(self):
        """The probe always runs something: no target after the separators is a usage
        error on stderr, not an empty run."""
        out = run_probe(["--label", "empty"])
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("usage", out.stderr.lower())


if __name__ == "__main__":
    unittest.main()
