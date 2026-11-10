"""check_pr_self_audit.py zero-selected rules (#2403 item 5): a `0 passed`
result used as evidence without `not_run`, and a filtered `cargo test` call
that selects more than one target (or cargo's default set), are blocked in the
Self-Audit section and in markdown table rows. Boundaries are literal lines."""

import pathlib
import subprocess
import unittest

TOOL = pathlib.Path(__file__).resolve().parents[1] / "tools/check_pr_self_audit.py"

SECTION = """## Self-Audit

- mutation: guard removed -> check FAILED, measured
- clone: check set from a fresh clone
- host: gh exercised via a fake executable, no network
- receipt: digest re-read before PASS
{extra}
## Notes
"""


def body(extra="", table=""):
    return "Refs #2403\n\n" + table + "\n" + SECTION.format(extra=extra)


class ZeroSelectedTest(unittest.TestCase):
    def run_tool(self, text):
        return subprocess.run(
            ["python3", "-I", str(TOOL)], input=text, text=True, capture_output=True
        )

    def assert_pass(self, text):
        out = self.run_tool(text)
        self.assertEqual(out.returncode, 0, out.stdout + out.stderr)
        self.assertIn("self-audit: pass", out.stdout)

    def assert_blocked(self, text, tag):
        out = self.run_tool(text)
        self.assertEqual(out.returncode, 1, out.stdout + out.stderr)
        self.assertIn(tag, out.stdout)

    # --- `0 passed` evidence ---------------------------------------------

    def test_zero_passed_in_section_is_blocked(self):
        self.assert_blocked(
            body("- check: test result: ok. 0 passed; 0 failed; 12 filtered out\n"),
            "zero-selected",
        )

    def test_one_passed_in_section_passes(self):
        self.assert_pass(body("- check: test result: ok. 1 passed; 0 failed\n"))

    def test_ten_and_twenty_passed_are_not_zero(self):
        self.assert_pass(body("- check: 10 passed and 20 passed\n"))

    def test_zero_passed_marked_not_run_passes(self):
        self.assert_pass(body("- check: hook_dispatch_cli 0 passed -> not_run\n"))

    def test_zero_passed_in_table_row_is_blocked(self):
        table = "| check | result |\n|---|---|\n| hooks_cli | ok. 0 passed |\n"
        self.assert_blocked(body(table=table), "zero-selected")

    def test_zero_passed_in_plain_prose_outside_section_is_not_checked(self):
        prose = "Earlier a filtered run printed 0 passed; it was rerun per target.\n"
        self.assert_pass(body(table=prose))

    # --- filtered cargo test target count ------------------------------------

    def test_filter_with_one_target_passes(self):
        self.assert_pass(body("- check: `cargo test --offline --locked --test index_cli spec` exit 0\n"))

    def test_filter_with_lib_only_passes(self):
        self.assert_pass(body("- check: `cargo test --lib hook_engine` exit 0\n"))

    def test_filter_with_two_targets_is_blocked(self):
        self.assert_blocked(
            body("- check: `cargo test --lib hook_engine --test hook_dispatch_cli` exit 0\n"),
            "filter-multi-target",
        )

    def test_filter_with_three_targets_is_blocked(self):
        self.assert_blocked(
            body("- check: `cargo test --lib --test a --test b name` exit 0\n"),
            "filter-multi-target",
        )

    def test_filter_without_target_is_blocked(self):
        self.assert_blocked(
            body("- check: `cargo test operations_doc` exit 0\n"), "filter-multi-target"
        )

    def test_filter_with_all_targets_is_blocked(self):
        self.assert_blocked(
            body("- check: `cargo test --all-targets name` exit 0\n"), "filter-multi-target"
        )

    def test_one_target_plus_a_many_selector_is_blocked(self):
        # `--tests` selects every integration target, so with `--lib` the filter
        # reaches more than one target even though only one single selector is named.
        self.assert_blocked(
            body("- check: `cargo test --lib --tests name` exit 0\n"), "filter-multi-target"
        )

    def test_filter_after_dashdash_counts(self):
        self.assert_blocked(
            body("- check: `cargo test --lib --test x -- name --exact` exit 0\n"),
            "filter-multi-target",
        )

    def test_two_targets_without_filter_passes(self):
        self.assert_pass(body("- check: `cargo test --offline --locked --lib --test x` exit 0\n"))

    def test_whole_suite_without_filter_passes(self):
        self.assert_pass(body("- check: `cargo test --offline --locked` exit 0\n"))

    def test_option_values_are_not_filters(self):
        self.assert_pass(
            body("- check: `cargo test -p moai-cli --features x --test a -- --skip slow --test-threads 1` exit 0\n")
        )

    def test_equals_form_option_value_is_not_a_filter(self):
        self.assert_pass(body("- check: `cargo test --target-dir=/t --test a` exit 0\n"))

    def test_bin_with_value_and_filter_passes(self):
        self.assert_pass(body("- check: `cargo test --bin moai-cli operations_doc` exit 0\n"))

    def test_second_command_on_the_line_is_checked(self):
        self.assert_blocked(
            body("- check: `cargo test --test a x` then `cargo test y` exit 0\n"),
            "filter-multi-target",
        )

    def test_prose_cargo_test_count_is_not_a_command(self):
        # Real PR bodies (#2365, #2367, #2374, #2377) say "cargo test 671 passed"
        # in prose; that is a result, not a filtered command.
        self.assert_pass(
            body("- clone: fresh clone, cargo test 671 passed / 0 failed, python Ran 124 OK\n")
        )

    def test_unquoted_command_is_not_read(self):
        self.assert_pass(body("- check: ran cargo test --lib x --test y by hand\n"))

    def test_filtered_multi_target_in_table_row_is_blocked(self):
        table = "| cmd | exit |\n|---|---|\n| `cargo test --lib x --test y` | 0 |\n"
        self.assert_blocked(body(table=table), "filter-multi-target")


if __name__ == "__main__":
    unittest.main()
