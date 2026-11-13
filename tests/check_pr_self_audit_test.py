"""check_pr_self_audit.py fixtures (#2279 scope 3): the section with all four
measured lines passes; a missing section, any missing line, or a mutation line
that does not report FAILED exits 1; usage exits 2."""

import pathlib
import subprocess
import unittest

TOOL = pathlib.Path(__file__).resolve().parents[1] / "tools/check_pr_self_audit.py"

FULL = """Refs #2200

## Self-Audit

- mutation: claim-verify guard mutated (claim filter removed) -> check FAILED, measured
- clone: check set from a fresh clone, not the /tmp symlink path
- host: api.github.com exercised via fake gh executable, no network
- receipt: receipt digest re-read and matched before PASS recorded

## Notes
"""

MISSING_SECTION = "Refs #2200\nNo audit section here.\n"
MISSING_RECEIPT = FULL.replace("- receipt: receipt digest re-read and matched before PASS recorded\n", "")
UNMEASURED_MUTATION = FULL.replace("-> check FAILED, measured", "-> check passed")


class CheckPrSelfAuditTest(unittest.TestCase):
    def run_tool(self, text=None, stdin=True):
        return subprocess.run(
            ["python3", "-I", str(TOOL)],
            input=text,
            text=True,
            capture_output=True,
            stdin=None if stdin else subprocess.DEVNULL,
        )

    def test_full_section_passes(self):
        out = self.run_tool(FULL)
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertIn("self-audit: pass", out.stdout)

    def test_file_input_passes(self):
        import tempfile
        with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False) as f:
            f.write(FULL)
            path = f.name
        out = subprocess.run(
            ["python3", "-I", str(TOOL), "--file", path],
            text=True, capture_output=True,
        )
        self.assertEqual(out.returncode, 0, out.stderr)
        self.assertIn("self-audit: pass", out.stdout)

    def test_missing_section_blocked(self):
        out = self.run_tool(MISSING_SECTION)
        self.assertEqual(out.returncode, 1)
        self.assertIn("## self-audit", out.stdout)

    def test_missing_receipt_line_blocked(self):
        out = self.run_tool(MISSING_RECEIPT)
        self.assertEqual(out.returncode, 1)
        self.assertIn("receipt", out.stdout)

    def test_unmeasured_mutation_blocked(self):
        out = self.run_tool(UNMEASURED_MUTATION)
        self.assertEqual(out.returncode, 1)
        self.assertIn("mutation:FAILED", out.stdout)

    def test_empty_body_blocked(self):
        out = self.run_tool("")
        self.assertEqual(out.returncode, 1)

    def test_usage_exit_2(self):
        out = subprocess.run(
            ["python3", "-I", str(TOOL), "bogus"],
            text=True, capture_output=True, stdin=subprocess.DEVNULL,
        )
        self.assertEqual(out.returncode, 2)


HEAD40 = "72331be4fda1786c32c3189a6faa4a817324a200"


def with_ci(line):
    """FULL with one extra line inside the Self-Audit section."""
    return FULL.replace("\n## Notes", line + "\n\n## Notes")


class CiCitationLineTest(unittest.TestCase):
    """#2385: a `ci:` line needs keyed `run=<digits>` and `head=<40 lowercase hex>`."""

    def run_tool(self, text):
        return subprocess.run(["python3", "-I", str(TOOL)], input=text, text=True, capture_output=True)

    def assert_pass(self, text):
        out = self.run_tool(text)
        self.assertEqual(out.returncode, 0, out.stdout + out.stderr)

    def assert_blocked(self, text, token):
        out = self.run_tool(text)
        self.assertEqual(out.returncode, 1, out.stdout)
        self.assertIn(token, out.stdout)

    def test_no_ci_line_still_passes(self):
        self.assert_pass(FULL)

    def test_full_citation_passes(self):
        self.assert_pass(with_ci("- ci: rc-ci run=42821157127 head=" + HEAD40))

    def test_both_workflow_spellings_follow_the_same_token_rules(self):
        # #2563: the workflow is renamed rc-ci -> ci. Open PRs keep the old word, new
        # ones write the new word; neither spelling relaxes the run=/head= checks.
        for word in ("ci", "rc-ci"):
            with self.subTest(word=word):
                self.assert_pass(with_ci(f"- ci: {word} run=42821157127 head=" + HEAD40))
                self.assert_blocked(with_ci(f"- ci: {word} run=42821157127 head=" + HEAD40[:39]), "ci:head")
                self.assert_blocked(with_ci(f"- ci: {word} run=42821157127 head=" + HEAD40 + "-dirty"), "ci:head")
                self.assert_blocked(with_ci(f"- ci: {word} rerun=42821157127 head=" + HEAD40), "ci:run")
                self.assert_blocked(with_ci(f"- ci: {word} run= head=" + HEAD40), "ci:run")

    def test_head_39_hex_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42821157127 head=" + HEAD40[:39]), "ci:head")

    def test_head_40_hex_passes(self):
        self.assert_pass(with_ci("- ci: rc-ci run=1 head=" + "a" * 40))

    def test_head_41_hex_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42821157127 head=" + HEAD40 + "0"), "ci:head")

    def test_uppercase_head_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42821157127 head=" + HEAD40.upper()), "ci:head")

    def test_empty_run_id_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run= head=" + HEAD40), "ci:run")

    def test_one_digit_run_id_passes(self):
        self.assert_pass(with_ci("- ci: rc-ci run=7 head=" + HEAD40))

    def test_unkeyed_run_id_blocked(self):
        # A bare number (or the digits inside the SHA) is not a run id.
        self.assert_blocked(with_ci("- ci: rc-ci 42821157127 head=" + HEAD40), "ci:run")

    def test_prefixed_run_key_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci rerun=42821157127 head=" + HEAD40), "ci:run")

    def test_missing_head_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42821157127 " + HEAD40), "ci:head")

    def test_not_run_with_reason_passes(self):
        self.assert_pass(with_ci("- ci: not_run (merge conflict, pull_request run did not start)"))

    def test_not_run_without_reason_blocked(self):
        self.assert_blocked(with_ci("- ci: not_run"), "ci:run")

    # Audit #2395 F1/N1/N2: suffixes and prefixes that are not whitespace must not pass.
    def test_head_40_hex_then_dash_suffix_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42 head=" + HEAD40 + "-dirty"), "ci:head")

    def test_head_40_hex_then_underscore_suffix_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42 head=" + HEAD40 + "_x"), "ci:head")

    def test_head_40_hex_then_dot_suffix_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42 head=" + HEAD40 + ".x"), "ci:head")

    def test_head_40_hex_then_non_hex_letter_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42 head=" + HEAD40 + "g"), "ci:head")

    def test_head_40_hex_at_line_end_passes(self):
        self.assert_pass(with_ci("- ci: rc-ci run=42 head=72331be4fda1786c32c3189a6faa4a817324a200"))

    def test_head_40_hex_then_space_passes(self):
        self.assert_pass(with_ci("- ci: rc-ci run=42 head=72331be4fda1786c32c3189a6faa4a817324a200 conclusion=success"))

    def test_head_with_dot_prefix_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42 x.head=" + HEAD40), "ci:head")

    def test_run_with_dot_prefix_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci .run=42 head=" + HEAD40), "ci:run")

    def test_run_with_suffix_blocked(self):
        self.assert_blocked(with_ci("- ci: rc-ci run=42.5 head=" + HEAD40), "ci:run")

    def test_not_run_with_punctuation_only_blocked(self):
        self.assert_blocked(with_ci("- ci: not_run:"), "ci:run")

    def test_not_run_with_one_word_reason_passes(self):
        self.assert_pass(with_ci("- ci: not_run conflict"))

    def test_every_ci_line_is_checked(self):
        good = "- ci: rc-ci run=42821157127 head=" + HEAD40
        bad = "- ci: rc-ci run=42821157128"
        self.assert_blocked(with_ci(good + "\n" + bad), "ci:head")


ROOT = pathlib.Path(__file__).resolve().parents[1]
CI_FORMAT = "ci: ci run=<run id> head=<40-hex head SHA>"


class CheckRunRulesSingleSourceTest(unittest.TestCase):
    """#2385 AC-1: the rule table lives once in the VERIFY workflow; the other
    places point to it and repeat the same `ci:` line format."""

    def read(self, rel):
        return (ROOT / rel).read_text(encoding="utf-8")

    def test_generated_agents_md_points_at_the_single_source(self):
        # #2385 AC-1: the AGENTS.md managed-section template (both languages) refers to the
        # VERIFY workflow's "Check-run rules" instead of restating them, and that section exists.
        template = self.read("src/agents.rs")
        for line in (
            'Which checks run locally, and when an exact-head CI result may be cited instead, '
            'follow the \\"Check-run rules\\" section of the VERIFY workflow (the single source).',
            '어떤 검사를 로컬에서 돌리고 언제 정확한 head의 CI 결과를 대신 인용할 수 있는지는 '
            'VERIFY workflow의 \\"Check-run rules\\" 절(단일 출처)을 따른다.',
        ):
            self.assertEqual(template.count(line), 1, line)
        self.assertEqual(self.read("skills/moai/workflows/verify.md").count("\n## Check-run rules\n"), 1)
        # The template points; it never restates the rules (no citation format, no role table).
        self.assertNotIn(CI_FORMAT, template)
        self.assertNotIn("Delta auditor: merge-only delta", template)

    def test_verify_holds_the_rules_once(self):
        text = self.read("skills/moai/workflows/verify.md")
        self.assertEqual(text.count("\n## Check-run rules\n"), 1)
        self.assertIn(CI_FORMAT, text)
        for case in ("Author: new or changed code", "Author: a language or suite the change does not touch",
                     "Repair lane: the repair changes code",
                     "Auditor: the audited head changes code", "Delta auditor: merge-only delta",
                     "Author: a skill package file changes"):
            self.assertIn(case, text)
        # One build folder per worktree, and lanes report instead of deleting (#2385 follow-up).
        self.assertIn("never share `CARGO_TARGET_DIR` between worktrees", text)
        # WF-OBS repeat findings documented by Issue #2493, each with its ledger number.
        self.assertEqual(text.count("\n## Shell and helper rules\n"), 1)
        # The split retry is bounded (pre-audit FAIL 6814410637 B1): only a refusal of the
        # combined form, the refused effect stays forbidden, every step passes the same guard.
        for rule in ("Guard refusal (WF-OBS #17)",
                     "Only when the refusal is confirmed to be about the combined form",
                     "not that its effect is forbidden", "the refused effect itself stays forbidden",
                     "each step must pass the same guard on its own", "Splitting is never a bypass",
                     "Helper authentication failure (WF-OBS #19)", "record that review as `not_run`",
                     "Never record a key or token value", "zsh word splitting (WF-OBS #16)",
                     "for x in ${=var}", "never name a shell function after an existing alias"):
            self.assertIn(rule, text)
        # WF-OBS ledger rows 9, 25, 26, 29 and 30 (Issue #2401 CLOSE-TABLE 6841136693).
        self.assertEqual(text.count("\n## Change-specific rules\n"), 1)
        for rule in ("Base drift (WF-OBS #9)", "`git ls-remote origin refs/heads/<base>`",
                     "say which one you used",
                     "Platform-specific code (WF-OBS #25)",
                     "`cargo check --offline --locked --target x86_64-pc-windows-msvc`",
                     "record that check as `not_run`",
                     "Checker changes (WF-OBS #26)", "run the old and the new version over recent real inputs",
                     "report every verdict that changed",
                     "Absence in tests (WF-OBS #29)", "creates that absence in its own fixture",
                     "it never relies on the host lacking it",
                     "Manifest after a conflict (WF-OBS #30)", "take neither side's sha256",
                     "recompute it from the merged file", "check every entry of the manifest against its file"):
            self.assertIn(rule, text)
        self.assertNotIn("Reuse the existing target directory instead of creating a new one", text)
        # Cleaning removes cargo output only; `target/` and lane files stay (pre-audit B1).
        # The leader's canonical wording, Issue #2385 comment 6777290466, points 1-4.
        self.assertIn("Issue #2385 comment 6777290466", text)
        self.assertIn("removes only the cargo output it made itself", text)
        self.assertIn("the `debug`, `release` and `<target triple>` folders", text)
        self.assertIn("never removes `target/` itself or the notes, evidence and scripts under `target/<name>/`", text)
        self.assertIn("never touches another lane's worktree", text)
        self.assertIn("a script the leader writes after checking the list and the operator runs, never by a lane", text)
        self.assertIn("a worktree itself is removed only after a separate confirmation", text)
        self.assertNotIn("the leader removes them after checking the list", text)
        self.assertNotIn("The leader cleans the build folders of merged work", text)
        self.assertNotIn("A lane does not delete build folders", text)

    def test_other_places_reference_without_restating(self):
        for rel in ("skills/moai/workflows/audit.md", "skills/moai/workflows/sync.md",
                    ".github/pull_request_template.md"):
            with self.subTest(path=rel):
                text = self.read(rel)
                self.assertIn("Check-run rules", text)
                self.assertNotIn("| Role and change |", text)

    def test_pr_form_and_checker_use_the_same_ci_format(self):
        self.assertIn(CI_FORMAT, self.read(".github/pull_request_template.md"))
        self.assertIn(CI_FORMAT, self.read("tools/check_pr_self_audit.py"))


class NoAuthorCiLineTest(unittest.TestCase):
    """Decision D4 (Issue #1975 comment 6840449133, Issue #2612): the author writes no `ci:`
    line; the auditor reads the head's CI and cites it in the verdict, and the merge gate reads
    the head's check runs. During the transition a body with or without the line passes."""

    def read(self, rel):
        return (ROOT / rel).read_text(encoding="utf-8")

    def test_rules_say_the_auditor_cites_ci(self):
        verify = self.read("skills/moai/workflows/verify.md")
        self.assertIn("The auditor reads the run back itself and writes it in the verdict as one line", verify)
        self.assertIn("the author's `## Self-Audit` section needs no such line, and no verdict waits for one", verify)
        self.assertIn("passes a body with or without that line", verify)
        self.assertIn("without waiting for an author's `ci:` line (decision D4",
                      self.read("skills/moai/workflows/audit.md"))
        self.assertIn("no author `ci:` line is waited for (decision D4",
                      self.read("skills/moai/workflows/mode-leader.md"))
        self.assertIn("it needs no `ci:` line in its body (decision D4",
                      self.read("skills/moai/workflows/mode-agent.md"))
        self.assertIn("저자는 `## Self-Audit` 절에 `ci:` 줄을 적지 않아도 된다",
                      self.read(".github/pull_request_template.md"))

    def test_no_merge_rule_waits_for_an_author_line(self):
        for rel in ("skills/moai/workflows/mode-agent.md", "skills/moai/workflows/mode-leader.md",
                    "skills/moai/workflows/sync.md"):
            text = self.read(rel)
            self.assertNotIn("its `ci:` line", text, rel)
            self.assertIn("a non-author PASS that cites that run", text, rel)
        leader = self.read("skills/moai/workflows/mode-leader.md")
        self.assertNotIn("top-priority next act", leader)
        self.assertNotIn("never posts a verdict over a stale `ci:` line", leader)


if __name__ == "__main__":
    unittest.main()
