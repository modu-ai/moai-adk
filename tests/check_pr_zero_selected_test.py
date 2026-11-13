"""check_pr_self_audit.py zero-selected rules (#2403 item 5): a `0 passed`
result used as evidence without `not_run`, and a filtered `cargo test` call
that selects more than one target (or cargo's default set), are blocked in the
Self-Audit section and in markdown table rows. Boundaries are literal lines."""

import pathlib
import unittest
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from self_audit_run import run_check  # noqa: E402  (#2708: the Rust checker)



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
        return run_check((), text)

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

    def test_quoted_option_values_are_not_filters(self):
        for value in ['"slow test"', "'slow test'", '"slow || test)"']:
            with self.subTest(value=value):
                self.assert_pass(
                    body(f"- check: `cargo test --lib --test cli_a -- --skip {value}` exit 0\n")
                )
        self.assert_pass(
            body('- check: `cargo test --features "fast slow" --lib --test cli_a` exit 0\n')
        )

    def test_filter_with_quoted_skip_still_blocks_multiple_targets(self):
        self.assert_blocked(
            body('- check: `cargo test --lib --test cli_a needle -- --skip "slow test"` exit 0\n'),
            "filter-multi-target",
        )

    def test_filter_with_target_glob_is_blocked(self):
        for selector in [
            "--test 'cli_*'",
            "--test='cli_*'",
            "--bin 'app_?'",
            "--example 'demo_[ab]'",
            "--bench=bench_*",
        ]:
            with self.subTest(selector=selector):
                self.assert_blocked(
                    body(f"- check: `cargo test --offline --locked {selector} needle` exit 0\n"),
                    "filter-multi-target",
                )

    def test_target_glob_without_a_filter_passes(self):
        self.assert_pass(
            body('- check: `cargo test --test "cli_*" -- --skip "slow test"` exit 0\n')
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

    def test_unrelated_cargo_apostrophe_span_does_not_break_command_checks(self):
        note = "- note: `cargo's cache` was reused\n"
        self.assert_pass(body(note))
        self.assert_blocked(
            body(note + "- check: `cargo test --lib --test cli_a needle` exit 0\n"),
            "filter-multi-target",
        )

    def test_shell_redirections_preserve_real_filters_and_quoted_punctuation(self):
        redirections = [
            "> result.log",
            ">>result.log",
            '>"result log"',
            ">> 'result log'",
            "2> errors.log",
            '2>"error log"',
            "2>&1",
            "< input.txt",
            '<"input file"',
            "&> combined.log",
            '&>"combined log"',
        ]
        for redirection in redirections:
            with self.subTest(redirection=redirection, selection="whole-suite"):
                self.assert_pass(
                    body(f"- check: `cargo test {redirection}` exit 0\n")
                )
            for command in [
                f"cargo test --lib --test cli_a needle {redirection}",
                f"cargo test {redirection} --lib --test cli_a needle",
            ]:
                with self.subTest(redirection=redirection, command=command):
                    self.assert_blocked(
                        body(f"- check: `{command}` exit 0\n"),
                        "filter-multi-target",
                    )
        for value in ['">"', "'2>&1'", '"slow > test"', '"slow < test &> log"']:
            with self.subTest(quoted_option_value=value):
                self.assert_pass(
                    body(f"- check: `cargo test --lib --test cli_a -- --skip {value}` exit 0\n")
                )
        for value in ['">"', "'>>'", '"2>&1"', '"<"', '"&>"']:
            with self.subTest(quoted_filter=value):
                self.assert_blocked(
                    body(f"- check: `cargo test --lib --test cli_a {value}` exit 0\n"),
                    "filter-multi-target",
                )

    def test_cargo_global_options_do_not_hide_test_filters(self):
        prefixes = [
            "+nightly",
            "--offline",
            "--locked",
            "--frozen",
            "-q",
            "-v",
            "-vv",
            "-qv",
            "--color always",
            "--color=always",
            '--config "build.jobs=2"',
            "--config=build.jobs=2",
            "-Z unstable-options",
            "-Z=unstable-options",
            "-C workspace",
            "-C=workspace",
            '+nightly --offline --locked -vv --color=never --config "build.jobs=2"',
        ]
        for prefix in prefixes:
            with self.subTest(prefix=prefix, selection="two-targets"):
                self.assert_blocked(
                    body(f"- check: `cargo {prefix} test --lib --test cli_a needle` exit 0\n"),
                    "filter-multi-target",
                )
            with self.subTest(prefix=prefix, selection="one-target"):
                self.assert_pass(
                    body(f"- check: `cargo {prefix} test --test cli_a needle` exit 0\n")
                )
            with self.subTest(prefix=prefix, selection="unfiltered"):
                self.assert_pass(
                    body(f"- check: `cargo {prefix} test --lib --test cli_a` exit 0\n")
                )
        for command in [
            "cargo --offline build --lib --bin app needle",
            "cargo --color always check --lib --bin app needle",
            "cargo +nightly --locked fmt -- --check",
        ]:
            with self.subTest(non_test_command=command):
                self.assert_pass(body(f"- check: `{command}` exit 0\n"))

    def test_report_time_is_a_boolean_harness_flag(self):
        self.assert_blocked(
            body("- check: `cargo test --lib --test cli_a -- -Z unstable-options --report-time needle` exit 0\n"),
            "filter-multi-target",
        )
        self.assert_pass(
            body("- check: `cargo test --lib --test cli_a -- -Z unstable-options --report-time` exit 0\n")
        )

    def test_optional_leading_pipe_tables_are_evidence_but_plain_prose_is_not(self):
        layouts = [
            ("check | result", "--- | ---", "{check} | {result}"),
            ("check | result", ":--- | ---:", "{check} | {result}"),
            ("check | result |", ":---: | ---: |", "{check} | {result} |"),
            ("| check | result |", "| :--- | ---: |", "| {check} | {result} |"),
            ("check | result | notes", "--- | --- | ---", "{check} | {result}"),
            ("check | result", "--- | ---", "{check} | {result} | measured"),
        ]
        for header, delimiter, row in layouts:
            for check, result, tag in [
                ("hooks_cli", "ok. 0 passed", "zero-selected"),
                ("`cargo test --lib --test cli_a needle`", "exit 0", "filter-multi-target"),
            ]:
                table = "\n".join([header, delimiter, row.format(check=check, result=result), ""])
                with self.subTest(header=header, delimiter=delimiter, evidence=tag):
                    self.assert_blocked(body(table=table), tag)
            table = "\n".join([
                header,
                delimiter,
                row.format(check="hooks_cli", result="0 passed -> not_run"),
                "",
            ])
            with self.subTest(header=header, delimiter=delimiter, evidence="not_run"):
                self.assert_pass(body(table=table))
        for prose in [
            "Earlier a filtered run printed 0 passed; it was rerun per target.\n",
            "The example `cargo test --lib --test cli_a needle` is not execution evidence.\n",
            "Earlier 0 passed | historical result, not a table.\n",
        ]:
            with self.subTest(prose=prose):
                self.assert_pass(body(table=prose))


if __name__ == "__main__":
    unittest.main()
