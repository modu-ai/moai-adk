"""Issue #2388 AC-3b: explicit re-audit CLI completeness checks."""

import pathlib
import tempfile
import unittest
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from self_audit_run import run_check  # noqa: E402  (#2708: the Rust checker)



ROOT = pathlib.Path(__file__).resolve().parents[1]
SCOPE = "tools/check_pr_self_audit.py의 F1 필드 검증과 변경한 CLI 경계를 수리하고 확인했다."
SCOPE_LINE = "- 수리 확인 범위: " + SCOPE + "\n"
FINDINGS_LINE = "- 새 지적 분류: 없음\n"
REVIEW = "## 독립 review\n\n" + SCOPE_LINE + FINDINGS_LINE + "\n"
CI_LINE = "- ci: rc-ci run=42821157127 head=72331be4fda1786c32c3189a6faa4a817324a200\n"
FULL = """Refs #2388

## Self-Audit

- mutation: claim-verify guard mutated (claim filter removed) -> check FAILED, measured
- clone: check set from a fresh clone, not the /tmp symlink path
- host: api.github.com exercised via fake gh executable, no network
- receipt: receipt digest re-read and matched before PASS recorded
- ci: rc-ci run=42821157127 head=72331be4fda1786c32c3189a6faa4a817324a200

## 독립 review

- 수리 확인 범위: tools/check_pr_self_audit.py의 F1 필드 검증과 변경한 CLI 경계를 수리하고 확인했다.
- 새 지적 분류: 없음

## Notes

Repair verification is recorded above.
"""


class CheckPrReauditTest(unittest.TestCase):
    def run_tool(self, text, *args):
        return run_check(args, text, cwd=ROOT)

    def assert_pass(self, text, *args):
        out = self.run_tool(text, *args)
        self.assertEqual(out.returncode, 0, out.stdout + out.stderr)
        self.assertIn("self-audit: pass", out.stdout)
        return out

    def assert_blocked(self, text, token):
        out = self.run_tool(text, "--reaudit")
        self.assertEqual(out.returncode, 1, out.stdout + out.stderr)
        self.assertIn(token, out.stdout)

    def with_scope(self, value):
        return FULL.replace(SCOPE_LINE, "- 수리 확인 범위: " + value + "\n")

    def with_findings(self, value):
        return FULL.replace(FINDINGS_LINE, "- 새 지적 분류: " + value + "\n")

    def test_explicit_reaudit_stdin_passes(self):
        self.assert_pass(FULL, "--reaudit")

    def test_explicit_reaudit_file_passes_in_both_flag_orders(self):
        with tempfile.TemporaryDirectory(prefix="reaudit-test-", dir=ROOT) as directory:
            path = pathlib.Path(directory) / "body.md"
            path.write_text(FULL, encoding="utf-8")
            for args in (("--reaudit", "--file", str(path)),
                         ("--file", str(path), "--reaudit")):
                with self.subTest(args=args):
                    self.assert_pass(None, *args)

    def test_initial_audit_without_review_remains_accepted(self):
        self.assert_pass(FULL.replace(REVIEW, ""))

    def test_initial_audit_does_not_infer_reaudit_from_placeholders_or_epoch(self):
        initial = FULL.replace(REVIEW, "")
        baseline = self.assert_pass(initial)
        candidates = (
            FULL.replace(SCOPE, "<수리 범위>").replace("새 지적 분류: 없음", "새 지적 분류: TODO"),
            initial.replace("Refs #2388", "Refs #2388\nclaim2388#6782371193 epoch3"),
        )
        for body in candidates:
            with self.subTest(body=body):
                out = self.assert_pass(body)
                self.assertEqual((out.stdout, out.stderr), (baseline.stdout, baseline.stderr))

    def test_markdown_list_and_whitespace_conventions_pass(self):
        for marker in ("", "- ", "* ", "\t*\t"):
            with self.subTest(marker=marker):
                body = FULL.replace(SCOPE_LINE, marker + "수리 확인 범위:  " + SCOPE + "  \n")
                body = body.replace(FINDINGS_LINE, marker + "새 지적 분류:  없음  \n")
                self.assert_pass(body, "--reaudit")

    def test_valid_new_finding_classifications_pass(self):
        for classification in (
            "없음",
            "차단: 수리한 경계에서 계약 위반으로 데이터 손상이 발생한다.",
            "후속 Issue #2388: 계약을 유지하는 진단 메시지 개선을 추적한다.",
            "후속 Issue #1: 설명 보강을 추적한다.",
        ):
            with self.subTest(classification=classification):
                self.assert_pass(self.with_findings(classification), "--reaudit")


    def test_completed_descriptions_with_marker_like_references_pass(self):
        descriptions = (
            "src/todo.rs 경계와 호출자를 수리하고 확인했다.",
            "TODO.md 경계와 호출자를 수리하고 확인했다.",
            "TBD-related.md 경계와 호출자를 수리하고 확인했다.",
            "`ci: not_run` 분류를 보존하는 CLI 경계를 수리하고 확인했다.",
            "0 < n > 10 비교 경계를 수리하고 확인했다.",
        )
        for description in descriptions:
            with self.subTest(description=description):
                self.assert_pass(self.with_scope(description), "--reaudit")
                for classification in ("차단: ", "후속 Issue #2388: "):
                    self.assert_pass(
                        self.with_findings(classification + description), "--reaudit",
                    )

    def test_marker_named_files_remain_valid_evidence(self):
        for path in ("TODO.md", "TBD-related.md", "src/TODO.md", "docs/TBD-related.md"):
            with self.subTest(path=path):
                self.assert_pass(self.with_scope(path), "--reaudit")
                self.assert_pass(self.with_findings("차단: " + path), "--reaudit")

    def test_punctuation_suffixed_bare_markers_are_blocked(self):
        for marker in ("TODO", "TBD", "not_run"):
            for suffix in (".", "-", "...", ".-"):
                for value in (marker + suffix, "`" + marker + "`" + suffix):
                    with self.subTest(value=value):
                        self.assert_blocked(self.with_scope(value), "reaudit:scope")
                        for classification in ("차단: ", "후속 Issue #2388: "):
                            self.assert_blocked(
                                self.with_findings(classification + value),
                                "reaudit:findings",
                            )

    def test_inline_comment_delimiters_in_code_explanations_pass(self):
        for code in (
            "`<!--`",
            "``<!--``",
            "```<!--```",
            "`` `<!--` ``",
            "``` ``<!--`` ```",
            "`-->`",
        ):
            description = (
                "HTML comment opener " + code +
                " 인식 경계를 수리하고 회귀 시험을 확인했다."
            )
            with self.subTest(code=code):
                self.assert_pass(self.with_scope(description), "--reaudit")
                for classification in ("차단: ", "후속 Issue #2388: "):
                    self.assert_pass(
                        self.with_findings(classification + description), "--reaudit",
                    )

    def test_real_comments_after_inline_code_still_hide_required_fields(self):
        for code in ("`<!--`", "`` `<!--` ``", "```<!--```"):
            for line, token in (
                (SCOPE_LINE, "reaudit:scope"),
                (FINDINGS_LINE, "reaudit:findings"),
            ):
                with self.subTest(code=code, field=token):
                    hidden = code + " <!--\n" + line + "-->\n"
                    self.assert_blocked(FULL.replace(line, hidden), token)

    def test_unmatched_or_mismatched_code_runs_do_not_hide_real_comments(self):
        for prefix in ("`", "``", "``literal` ", "`literal`` ", "\\`literal` "):
            with self.subTest(prefix=prefix):
                hidden = prefix + "<!--\n" + FINDINGS_LINE + "-->\n"
                self.assert_blocked(
                    FULL.replace(FINDINGS_LINE, hidden), "reaudit:findings",
                )

    def test_backticks_inside_actual_comments_do_not_expose_fields(self):
        hidden = "<!-- `\n" + FINDINGS_LINE + "` -->\n"
        self.assert_blocked(FULL.replace(FINDINGS_LINE, hidden), "reaudit:findings")

    def test_hidden_review_sections_cannot_supply_evidence(self):
        examples = (
            "<!--\n" + REVIEW + "-->\n",
            "<!-- " + REVIEW.replace("\n", "\n<!-- -->") + " -->\n",
            "```markdown\n" + REVIEW + "```\n",
            "~~~~markdown\n" + REVIEW + "~~~~\n",
            "".join("    " + line + "\n" for line in REVIEW.splitlines()),
            "".join("\t" + line + "\n" for line in REVIEW.splitlines()),
        )
        initial = FULL.replace(REVIEW, "")
        empty_review = "## 독립 review\n\n"
        for example in examples:
            with self.subTest(example=example):
                self.assert_blocked(initial + example, "## 독립 review")
                self.assert_blocked(
                    initial + example + empty_review, "reaudit:scope",
                )

    def test_commented_fields_cannot_complete_visible_review(self):
        for line, token in (
            (SCOPE_LINE, "reaudit:scope"),
            (FINDINGS_LINE, "reaudit:findings"),
        ):
            for hidden in ("<!-- " + line.rstrip() + " -->\n",
                           "<!--\n" + line + "-->\n"):
                with self.subTest(hidden=hidden):
                    self.assert_blocked(FULL.replace(line, hidden), token)

    def test_duplicate_visible_review_headings_are_ambiguous(self):
        for duplicate in (REVIEW, "## 독립 review\n\n"):
            for body in (FULL + duplicate, FULL.replace(REVIEW, duplicate + REVIEW)):
                with self.subTest(body=body):
                    self.assert_blocked(body, "## 독립 review")

    def test_genuine_review_ignores_examples_and_inline_comments(self):
        examples = (
            "<!--\n" + REVIEW + "-->\n",
            "<!-- ## 독립 review -->\n",
            "````markdown\n```\n" + REVIEW + "````\n",
            "~~~~markdown\n~~~\n" + REVIEW + "~~~~~\n",
            "    ## 독립 review\n    - 수리 확인 범위: TODO\n",
            "\t## 독립 review\n\t수리 확인 범위: TODO\n",
            "<!--\n    -->\n",
        )
        for example in examples:
            with self.subTest(example=example):
                body = FULL.replace(SCOPE_LINE, example + SCOPE_LINE)
                self.assert_pass(body, "--reaudit")
        body = FULL.replace(SCOPE_LINE, SCOPE_LINE.rstrip() + " <!-- TODO -->\n")
        self.assert_pass(body, "--reaudit")

    def test_short_or_wrong_fence_closers_do_not_expose_examples(self):
        for opener, false_closer, closer in (
            ("````", "```", "````"),
            ("~~~~", "~~~", "~~~~"),
            ("```", "~~~", "```"),
        ):
            with self.subTest(opener=opener, false_closer=false_closer):
                example = opener + "\n" + false_closer + "\n" + REVIEW + closer + "\n"
                self.assert_blocked(FULL.replace(REVIEW, example), "## 독립 review")

    def test_bare_code_wrapped_placeholders_remain_blocked(self):
        for value in ("`TODO`", "`TBD`", "`not_run`", "`<path>`"):
            with self.subTest(value=value):
                self.assert_blocked(self.with_scope(value), "reaudit:scope")
                for classification in ("차단: ", "후속 Issue #2388: "):
                    self.assert_blocked(
                        self.with_findings(classification + value), "reaudit:findings",
                    )

    def test_missing_review_section_is_blocked(self):
        self.assert_blocked(FULL.replace(REVIEW, ""), "## 독립 review")

    def test_only_exact_review_heading_counts(self):
        for heading in ("### 독립 review", "## 독립 reviewer", "## 독립 검토"):
            with self.subTest(heading=heading):
                self.assert_blocked(FULL.replace("## 독립 review", heading), "## 독립 review")

    def test_missing_repair_scope_is_blocked(self):
        self.assert_blocked(FULL.replace(SCOPE_LINE, ""), "reaudit:scope")

    def test_missing_finding_classification_is_blocked(self):
        self.assert_blocked(FULL.replace(FINDINGS_LINE, ""), "reaudit:findings")

    def test_empty_repair_scope_is_blocked(self):
        for value in ("", " \t "):
            with self.subTest(value=value):
                self.assert_blocked(self.with_scope(value), "reaudit:scope")

    def test_empty_finding_classification_is_blocked(self):
        for value in ("", " \t "):
            with self.subTest(value=value):
                self.assert_blocked(self.with_findings(value), "reaudit:findings")

    def test_placeholder_and_punctuation_scope_is_blocked(self):
        for value in (
            "TODO", "todo: verify F1", "TBD", "F1 boundary TBD", "not_run",
            "not_run (verification unavailable)", "<수리 범위>",
            "F1 <path> boundary repaired", "...", "---", "…", "()",
        ):
            with self.subTest(value=value):
                self.assert_blocked(self.with_scope(value), "reaudit:scope")

    def test_placeholder_and_punctuation_classification_is_blocked(self):
        for value in ("TODO", "TBD", "not_run", "<새 지적 분류>", "...", "---", "…"):
            with self.subTest(value=value):
                self.assert_blocked(self.with_findings(value), "reaudit:findings")

    def test_scope_outside_designated_section_does_not_count(self):
        body = FULL.replace(SCOPE_LINE, "")
        candidates = (
            SCOPE_LINE + "\n" + body,
            body.replace("## 독립 review", SCOPE_LINE + "\n## 독립 review"),
            body + "\n" + SCOPE_LINE,
            body.replace("## Notes", "### Repair details\n" + SCOPE_LINE + "\n## Notes"),
        )
        for candidate in candidates:
            with self.subTest(body=candidate):
                self.assert_blocked(candidate, "reaudit:scope")

    def test_classification_outside_designated_section_does_not_count(self):
        body = FULL.replace(FINDINGS_LINE, "")
        candidates = (
            FINDINGS_LINE + "\n" + body,
            body.replace("## 독립 review", FINDINGS_LINE + "\n## 독립 review"),
            body + "\n" + FINDINGS_LINE,
            body.replace("## Notes", "### Repair details\n" + FINDINGS_LINE + "\n## Notes"),
        )
        for candidate in candidates:
            with self.subTest(body=candidate):
                self.assert_blocked(candidate, "reaudit:findings")

    def test_conflicting_fields_outside_section_do_not_invalidate_valid_review(self):
        body = "- 수리 확인 범위: TODO\n- 새 지적 분류: TBD\n\n" + FULL
        body += "\n- 수리 확인 범위: <path>\n- 새 지적 분류: not_run\n"
        self.assert_pass(body, "--reaudit")

    def test_review_ends_at_next_heading_of_any_level(self):
        for heading in ("# Follow-up", "## Follow-up", "### Follow-up"):
            with self.subTest(heading=heading):
                body = FULL.replace(FINDINGS_LINE, heading + "\n" + FINDINGS_LINE)
                self.assert_blocked(body, "reaudit:findings")

    def test_duplicate_repair_scope_is_blocked(self):
        for duplicate in (SCOPE_LINE, "- 수리 확인 범위: 다른 경계만 수리했다.\n", "* 수리 확인 범위: TODO\n"):
            with self.subTest(duplicate=duplicate):
                body = FULL.replace(SCOPE_LINE, SCOPE_LINE + duplicate)
                self.assert_blocked(body, "reaudit:scope")

    def test_duplicate_finding_classification_is_blocked(self):
        for duplicate in (
            FINDINGS_LINE,
            "- 새 지적 분류: 차단: 데이터 손상 경로가 남아 있다.\n",
            "* 새 지적 분류: 후속 Issue #2388: 진단 개선을 추적한다.\n",
        ):
            with self.subTest(duplicate=duplicate):
                body = FULL.replace(FINDINGS_LINE, FINDINGS_LINE + duplicate)
                self.assert_blocked(body, "reaudit:findings")

    def test_unknown_finding_classification_is_blocked(self):
        for value in ("none", "보류", "통과", "새 지적 없음", "없음: 후속 확인 필요"):
            with self.subTest(value=value):
                self.assert_blocked(self.with_findings(value), "reaudit:findings")

    def test_blocking_classification_needs_concrete_reason(self):
        for value in ("차단", "차단:", "차단: \t", "차단: ...", "차단: ---", "차단: …", "차단: ()"):
            with self.subTest(value=value):
                self.assert_blocked(self.with_findings(value), "reaudit:findings")

    def test_blocking_placeholder_reason_is_blocked(self):
        for reason in ("TODO", "TBD", "not_run", "<reason>", "TODO: 데이터 경계 확인", "데이터 <boundary> 확인"):
            with self.subTest(reason=reason):
                self.assert_blocked(self.with_findings("차단: " + reason), "reaudit:findings")

    def test_followup_classification_needs_concrete_reason(self):
        for value in (
            "후속 Issue #2388", "후속 Issue #2388:", "후속 Issue #2388: \t",
            "후속 Issue #2388: ...", "후속 Issue #2388: ---", "후속 Issue #2388: …",
            "후속 Issue #2388: ()",
        ):
            with self.subTest(value=value):
                self.assert_blocked(self.with_findings(value), "reaudit:findings")

    def test_followup_placeholder_reason_is_blocked(self):
        for reason in ("TODO", "TBD", "not_run", "<reason>", "TODO: 진단 개선", "진단 <path> 개선"):
            with self.subTest(reason=reason):
                self.assert_blocked(self.with_findings("후속 Issue #2388: " + reason), "reaudit:findings")

    def test_followup_issue_requires_positive_ascii_integer(self):
        for number in (
            "", "0", "-1", "+1", "2388.0", "2e3", "２３８８", "٢٣٨٨",
            "2388suffix", "2388-1", "2388_1", "2388/1", "2388 9", "2388#",
        ):
            with self.subTest(number=number):
                value = "후속 Issue #" + number + ": 진단 개선을 추적한다."
                self.assert_blocked(self.with_findings(value), "reaudit:findings")

    def test_malformed_followup_syntax_is_blocked(self):
        for value in (
            "후속 Issue 2388: 진단 개선을 추적한다.",
            "후속 #2388: 진단 개선을 추적한다.",
            "후속 Issue #2388 진단 개선을 추적한다.",
            "후속 Issue #2388; 진단 개선을 추적한다.",
        ):
            with self.subTest(value=value):
                self.assert_blocked(self.with_findings(value), "reaudit:findings")

    def test_reaudit_still_requires_self_audit_section(self):
        self.assert_blocked(REVIEW, "## self-audit")

    def test_reaudit_still_requires_each_measured_self_audit_field(self):
        lines = (
            ("- mutation: claim-verify guard mutated (claim filter removed) -> check FAILED, measured\n", "mutation"),
            ("- clone: check set from a fresh clone, not the /tmp symlink path\n", "clone"),
            ("- host: api.github.com exercised via fake gh executable, no network\n", "host"),
            ("- receipt: receipt digest re-read and matched before PASS recorded\n", "receipt"),
        )
        for line, token in lines:
            with self.subTest(field=token):
                self.assert_blocked(FULL.replace(line, ""), token)

    def test_reaudit_still_requires_failed_mutation_check(self):
        self.assert_blocked(FULL.replace("-> check FAILED, measured", "-> check passed"), "mutation:FAILED")

    def test_reaudit_still_enforces_ci_citation_guards(self):
        cases = (
            ("- ci: rc-ci run=42821157127\n", "ci:head"),
            ("- ci: rc-ci head=72331be4fda1786c32c3189a6faa4a817324a200\n", "ci:run"),
            ("- ci: rc-ci run=42 head=72331be4fda1786c32c3189a6faa4a817324a200-dirty\n", "ci:head"),
            ("- ci: not_run\n", "ci:run"),
        )
        for line, token in cases:
            with self.subTest(line=line):
                self.assert_blocked(FULL.replace(CI_LINE, line), token)

    def test_reaudit_keeps_ci_citation_optional(self):
        self.assert_pass(FULL.replace(CI_LINE, ""), "--reaudit")

    def test_unknown_flags_exit_usage(self):
        for args in (("--unknown",), ("--reaudit", "--unknown"), ("--unknown", "--reaudit")):
            with self.subTest(args=args):
                out = self.run_tool(FULL, *args)
                self.assertEqual(out.returncode, 2, out.stdout + out.stderr)

    def test_duplicate_flags_exit_usage(self):
        with tempfile.TemporaryDirectory(prefix="reaudit-test-", dir=ROOT) as directory:
            path = pathlib.Path(directory) / "body.md"
            path.write_text(FULL, encoding="utf-8")
            cases = (
                ("--reaudit", "--reaudit"),
                ("--reaudit", "--file", str(path), "--reaudit"),
                ("--file", str(path), "--file", str(path)),
                ("--reaudit", "--file", str(path), "--file", str(path)),
            )
            for args in cases:
                with self.subTest(args=args):
                    out = self.run_tool(FULL, *args)
                    self.assertEqual(out.returncode, 2, out.stdout + out.stderr)

    def test_incomplete_file_arguments_exit_usage(self):
        for args in (("--file",), ("--reaudit", "--file"), ("--file", "--reaudit")):
            with self.subTest(args=args):
                out = self.run_tool(FULL, *args)
                self.assertEqual(out.returncode, 2, out.stdout + out.stderr)


if __name__ == "__main__":
    unittest.main()
