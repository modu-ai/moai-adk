"""tools/hook_latency.py (#2391 cost): the percentile math with literal inputs, the build
profile label and usage exits. No binary is run and no time is measured here: the tool's
numbers depend on machine load, so they are reported, never asserted."""

import importlib.util
import pathlib
import subprocess
import unittest

TOOL = pathlib.Path(__file__).resolve().parents[1] / "tools/hook_latency.py"
spec = importlib.util.spec_from_file_location("hook_latency", TOOL)
hook_latency = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hook_latency)


class PercentileTest(unittest.TestCase):
    def test_nearest_rank_on_ten_values(self):
        values = [10, 1, 9, 2, 8, 3, 7, 4, 6, 5]
        self.assertEqual(hook_latency.percentile(values, 50), 5)
        self.assertEqual(hook_latency.percentile(values, 95), 10)
        self.assertEqual(hook_latency.percentile(values, 90), 9)
        self.assertEqual(hook_latency.percentile(values, 100), 10)

    def test_rank_boundary_on_twenty_values(self):
        values = list(range(1, 21))
        # ceil(0.95 * 20) = 19 exactly; ceil(0.951 * 20) = 20.
        self.assertEqual(hook_latency.percentile(values, 95), 19)
        self.assertEqual(hook_latency.percentile(values, 95.1), 20)
        # ceil(0.5 * 20) = 10; one value more in the list moves the rank to 11 (ceil(10.5)).
        self.assertEqual(hook_latency.percentile(values, 50), 10)
        self.assertEqual(hook_latency.percentile(list(range(1, 22)), 50), 11)

    def test_single_value_and_invalid_inputs(self):
        self.assertEqual(hook_latency.percentile([7.5], 50), 7.5)
        self.assertEqual(hook_latency.percentile([7.5], 0.001), 7.5)
        with self.assertRaises(ValueError):
            hook_latency.percentile([], 50)
        with self.assertRaises(ValueError):
            hook_latency.percentile([1, 2], 0)
        with self.assertRaises(ValueError):
            hook_latency.percentile([1, 2], 100.01)

    def test_summary_keys(self):
        self.assertEqual(
            hook_latency.summarize([1.0, 2.0, 3.0, 4.0]),
            {"p50_ms": 2.0, "p95_ms": 4.0, "max_ms": 4.0},
        )


class LabelAndUsageTest(unittest.TestCase):
    def test_build_profile_from_path(self):
        self.assertEqual(hook_latency.profile_of("/x/target/release/moai-cli"), "release")
        self.assertEqual(hook_latency.profile_of("/x/target/debug/moai-cli"), "debug")
        self.assertEqual(hook_latency.profile_of("/usr/local/bin/moai-cli"), "unknown")

    def test_usage_exits_2(self):
        for argv in ([], ["--bin"], ["--bin", "x", "--samples", "0"], ["--bin", "x", "--samples", "n"]):
            with self.subTest(argv=argv):
                out = subprocess.run(["python3", "-I", str(TOOL), *argv],
                                     capture_output=True, text=True, stdin=subprocess.DEVNULL)
                self.assertEqual(out.returncode, 2, out.stdout)


if __name__ == "__main__":
    unittest.main()
