#!/usr/bin/env python3
"""Exercise the diagnostic receipt against actual Go event-stream shapes."""

import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("require-named-test.py")
PACKAGE = "github.com/modu-ai/moai-adk/internal/privatefile"
TEST = "TestSQLitePinnedSidecarLifetime"


def event(action, test=None, package=PACKAGE):
    row = {"Action": action, "Package": package}
    if test is not None:
        row["Test"] = test
    return row


class NamedTestReceipt(unittest.TestCase):
    def invoke(self, rows, raw=None):
        with tempfile.TemporaryDirectory() as temp:
            stream = Path(temp) / "events.jsonl"
            stream.write_text(raw if raw is not None else "\n".join(map(json.dumps, rows)), encoding="utf-8")
            return subprocess.run(
                [sys.executable, str(SCRIPT), "--events", str(stream), "--package", PACKAGE, "--test", TEST],
                capture_output=True, text=True, check=False,
            )

    def test_named_run_pass_and_package_pass_produce_receipt(self):
        result = self.invoke([event("start"), event("run", TEST), event("output", TEST), event("pass", TEST), event("pass")])
        self.assertEqual(result.returncode, 0, result.stderr)
        receipt = json.loads(result.stdout)
        self.assertEqual(receipt["package"], PACKAGE)
        self.assertEqual(receipt["test"], TEST)
        self.assertEqual(receipt["run"], 1)
        self.assertEqual(receipt["pass"], 1)
        self.assertEqual(receipt["package_pass"], 1)

    def test_zero_matched_tests_is_rejected(self):
        self.assertNotEqual(self.invoke([event("start"), event("pass")]).returncode, 0)

    def test_each_required_event_is_indispensable(self):
        rows = [event("run", TEST), event("pass", TEST), event("pass")]
        for missing in range(len(rows)):
            with self.subTest(missing=missing):
                self.assertNotEqual(self.invoke(rows[:missing] + rows[missing + 1:]).returncode, 0)

    def test_any_failure_or_skip_is_rejected(self):
        for action in ["fail", "skip"]:
            for test in [None, TEST, TEST + "/child"]:
                with self.subTest(action=action, test=test):
                    rows = [event("run", TEST), event(action, test), event("pass", TEST), event("pass")]
                    self.assertNotEqual(self.invoke(rows).returncode, 0)

    def test_another_test_or_package_cannot_supply_receipt(self):
        for rows in [
            [event("run", "TestOther"), event("pass", "TestOther"), event("pass")],
            [event("run", TEST, "other"), event("pass", TEST, "other"), event("pass", package="other")],
        ]:
            self.assertNotEqual(self.invoke(rows).returncode, 0)

    def test_duplicate_receipt_is_rejected(self):
        rows = [event("run", TEST), event("pass", TEST), event("pass")]
        for duplicate in rows:
            with self.subTest(duplicate=duplicate):
                self.assertNotEqual(self.invoke(rows + [duplicate]).returncode, 0)

    def test_events_must_follow_execution_order(self):
        self.assertNotEqual(self.invoke([event("pass", TEST), event("run", TEST), event("pass")]).returncode, 0)
        self.assertNotEqual(self.invoke([event("run", TEST), event("pass"), event("pass", TEST)]).returncode, 0)

    def test_malformed_or_non_object_json_is_rejected(self):
        for raw in ["not-json\n", "true\n", "[]\n", ""]:
            with self.subTest(raw=raw):
                self.assertNotEqual(self.invoke([], raw).returncode, 0)


if __name__ == "__main__":
    unittest.main()
