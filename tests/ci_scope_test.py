"""Tests for tools/ci_scope.py (#2527, design comment 6822993459)."""

import importlib.util
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def load_module():
    spec = importlib.util.spec_from_file_location(
        "ci_scope", os.path.join(REPO_ROOT, "tools", "ci_scope.py")
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CiScopeTest(unittest.TestCase):
    def setUp(self):
        self.ci_scope = load_module()

    def run_stdin(self, text):
        proc = subprocess.run(
            [sys.executable, os.path.join(REPO_ROOT, "tools", "ci_scope.py")],
            input=text,
            capture_output=True,
            text=True,
        )
        return proc

    def test_all_docs_is_docs_only(self):
        scope, reasons = self.ci_scope.classify(
            ["docs/a.md", "docs/img/b.png", "docs/c.svg", "docs/d.txt"]
        )
        self.assertEqual(scope, "docs-only")
        self.assertEqual(reasons, [])

    def test_one_src_file_is_full(self):
        scope, _ = self.ci_scope.classify(["docs/a.md", "src/main.rs"])
        self.assertEqual(scope, "full")

    def test_exact_deny_files_are_full(self):
        for path in sorted(self.ci_scope.EXACT_DENY):
            scope, reasons = self.ci_scope.classify([path])
            self.assertEqual(scope, "full", path)
            self.assertTrue(any(path in r for r in reasons), path)

    def test_e2e_receipts_prefix_is_full(self):
        scope, _ = self.ci_scope.classify(["docs/e2e-receipts/receipts.jsonl"])
        self.assertEqual(scope, "full")

    def test_space_applied_record_alone_is_full(self):
        # Only the Space record changed: space_applied_record_test must still run (#2207).
        for path in ("docs/design-review/applied.md", "docs/design-review/pages.tsv"):
            scope, _ = self.ci_scope.classify([path])
            self.assertEqual(scope, "full", path)

    def test_design_review_fixes_prefix_is_full(self):
        # tests/space_applied_record_test.py hashes every fix file (#2207).
        scope, reasons = self.ci_scope.classify(["docs/design-review/fixes/group3-487fa473.md"])
        self.assertEqual(scope, "full")
        self.assertTrue(any("prefix-denied" in r for r in reasons), reasons)

    def test_non_doc_suffix_is_full(self):
        scope, _ = self.ci_scope.classify(["docs/x.rs"])
        self.assertEqual(scope, "full")

    def test_rename_from_src_is_full(self):
        scope, _ = self.ci_scope.classify(["src/a.rs", "docs/a.md"])
        self.assertEqual(scope, "full")

    def test_docs_deletion_only_is_docs_only(self):
        scope, _ = self.ci_scope.classify(["docs/old.md"])
        self.assertEqual(scope, "docs-only")

    def test_reports_path_is_docs_only(self):
        scope, _ = self.ci_scope.classify([".moai-cli/reports/issue-copy.md"])
        self.assertEqual(scope, "docs-only")

    def test_empty_input_is_full(self):
        scope, reasons = self.ci_scope.classify([])
        self.assertEqual(scope, "full")
        self.assertTrue(reasons)

    def test_single_blank_line_is_full(self):
        # The workflow pipes `git diff --name-only`; an empty diff yields one blank
        # line, which must fail safe to full exactly like empty stdin (#2527).
        scope, reasons = self.ci_scope.classify([""])
        self.assertEqual(scope, "full")
        self.assertTrue(reasons)
        proc = self.run_stdin("\n")
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "full")

    def test_garbage_input_is_full(self):
        proc = self.run_stdin("not a real path\n\x00weird\n")
        self.assertEqual(proc.returncode, 0)
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "full")

    def test_script_exit_zero_and_json(self):
        proc = self.run_stdin("docs/a.md\n")
        self.assertEqual(proc.returncode, 0)
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "docs-only")
        self.assertEqual(data["reasons"], [])

    def test_script_src_file_full(self):
        proc = self.run_stdin("src/a.rs\n")
        self.assertEqual(proc.returncode, 0)
        data = json.loads(proc.stdout)
        self.assertEqual(data["scope"], "full")
        self.assertTrue(data["reasons"])


class WorkflowContractTest(unittest.TestCase):
    def test_scope_steps_pass_rename_sources(self):
        # #2538 B1: plain `git diff --name-only` emits only the rename destination,
        # so renaming/deleting a test-consumed doc could pass as docs-only. The two
        # docs Scope steps keep their --no-renames diffs; the targets step reads its
        # changed paths from the files API, whose filename AND previous_filename
        # carry the rename sources to the base copy of the classifier.
        with open(
            os.path.join(REPO_ROOT, ".github", "workflows", "ci.yml"),
            "r",
            encoding="utf-8",
        ) as fh:
            yaml_text = fh.read()
        self.assertEqual(yaml_text.count("git diff --name-only --no-renames"), 2)
        self.assertNotIn('git diff --name-only "${base}', yaml_text)
        self.assertIn("previous_filename", yaml_text)


class MainExceptionPathTest(unittest.TestCase):
    def test_internal_error_fails_safe_to_full(self):
        # #2538 audit note (M3): the except branch in main() must answer "full".
        from unittest import mock
        from io import StringIO

        module = load_module()
        with mock.patch.object(module, "classify", side_effect=RuntimeError("boom")):
            with mock.patch.object(module.sys, "stdin", StringIO("docs/a.md\n")):
                with mock.patch.object(module.sys, "stdout", StringIO()) as out:
                    code = module.main()
        self.assertEqual(code, 0)
        data = json.loads(out.getvalue())
        self.assertEqual(data["scope"], "full")
        self.assertTrue(any("boom" in r for r in data["reasons"]))


class DocumentationDriftTest(unittest.TestCase):
    """#2538 B2: every docs/ path a test, tool or src file names must be denied.

    A path referenced by name in tests/, tools/ or src/ is read, compared or
    embedded by that code; classifying it docs-only would skip the very check
    that consumes it. The scan is the same mechanical grep that produced
    EXACT_DENY, so adding a new reader without extending the denylist fails
    here first.
    """

    IGNORED_LITERAL_PREFIXES = (
        "docs/a.md", "docs/A.md", "docs/c.svg", "docs/d.txt", "docs/img/",
        "docs/x.rs", "docs/old.md", "docs/huge.md", "docs/ssh.md",
        "docs/spec.md", "docs/receipt.md", "docs/elsewhere.md",
        "docs/missing.md", "docs/untracked.md", "docs/staged.md",
        "docs/link.md", "docs/decisions-test.md", "docs/decisions.md",
        "docs/TBD-related.md", "docs/workflows/", "docs/wf",
    )


    def _fragmented_docs_paths(self):
        """Paths built from pieces like ROOT / "docs" / "x.md" or join("docs", "x.md").

        #2538 B2prime: the quoted-string scan misses these, so collect them with a
        separate pattern and deny their joined forms.
        """
        import re as _re
        import subprocess as _sp
        # Python: X / "docs" / "name"  and  os.path.join(..., "docs", "name")
        py = _sp.run(
            ["grep", "-rnE",
             r'["\x27]docs["\x27]\s*\/\s*["\x27]([A-Za-z0-9_./-]+)["\x27]',
             "tests", "tools", "src"],
            capture_output=True, text=True, cwd=REPO_ROOT,
        ).stdout
        # Line-filtered: skip this test's own docstrings and temp-dir fixtures
        # (a path under a tmp/scratch root never reads the repository docs/).
        keep = []
        for line in py.splitlines():
            if "ci_scope_test.py" in line:
                continue
            if "self.tmp" in line or "tmp.path" in line or "mkdtemp" in line:
                continue
            keep.append(line)
        py = "\n".join(keep)
        names = set(_re.findall(r'["\x27]docs["\x27]\s*/\s*["\x27]([A-Za-z0-9_./-]+)["\x27]', py))
        # Rust: .join("docs").join("x.md") or single .join("x.md") after a docs root
        rs_proc = _sp.run(
            ["grep", "-rnE",
             r'join\("docs"\)(?:\s*\.join\("[^"]+"\))*',
             "tests", "src"],
            capture_output=True, text=True, cwd=REPO_ROOT,
        )
        rs = "\n".join(
            line for line in rs_proc.stdout.splitlines()
            if "ci_scope_test.py" not in line
            and "self.tmp" not in line and "tmp.path" not in line
        )
        for chunk in rs.splitlines():
            for seg in _re.findall(r'join\("([A-Za-z0-9_./-]+)"\)', chunk)[1:]:
                if "." in seg:
                    names.add(seg)
        return {f"docs/{n}" for n in names if "." in n.rsplit("/", 1)[-1]}

    def test_fragmented_docs_paths_are_denied(self):
        found = self._fragmented_docs_paths()
        module = load_module()
        missing = sorted(p for p in found if p not in module.EXACT_DENY)
        self.assertEqual(missing, [], f"fragmented docs paths not denied: {missing}")

    def test_every_referenced_docs_path_is_denied(self):
        import re as _re
        import subprocess as _sp
        pattern = _re.compile(r'["\']((?:docs|\.\.\/docs)\/[A-Za-z0-9_./-]+)["\']')
        proc = _sp.run(
            ["grep", "-rhoE", r'["\x27](docs|\.\./docs)/[A-Za-z0-9_./-]+["\x27]',
             "tests", "tools", "src"],
            capture_output=True, text=True, cwd=REPO_ROOT,
        )
        quoted = set()
        for token in proc.stdout.splitlines():
            path = token.strip().strip('"').strip("'")
            path = path.replace("../docs/", "docs/", 1).rstrip("\\")
            if not path.startswith("docs/"):
                continue
            if any(path.startswith(p) for p in self.IGNORED_LITERAL_PREFIXES):
                continue
            module_now = load_module()
            if ("." not in path.rsplit("/", 1)[-1]
                    and f"{path}.md" in module_now.EXACT_DENY):
                continue  # prose shard of a denied .md file (usage/row strings)
            if path in ("docs/e2e-receipts", "docs/e2e-receipts/"):
                continue  # covered by PREFIX_DENY
            if path.startswith("docs/e2e-receipts/"):
                continue  # covered by PREFIX_DENY
            if any((path.rstrip("/") + "/").startswith(p) for p in module_now.PREFIX_DENY):
                continue  # a prefix-denied folder, named whole or by one of its files
            if path == "docs/support-matrix":
                continue  # prose shard of docs/support-matrix.md in usage/row strings; the .md path itself is denied
            if not path[5].isalnum() and path[5] != "_":
                continue  # trailing escape artifact from multi-line strings
            quoted.add(path)
        module = load_module()
        missing = sorted(p for p in quoted if p not in module.EXACT_DENY)
        self.assertEqual(missing, [], f"docs paths referenced by code but not denied: {missing}")


# C-76 (SPEC-20261119-macos-on-pr) AC-4: the file lists of the failed ci push runs
# on v2026.11.0, read 2026-11-19 (gh run list; merged-delta lists from local git,
# generated by the C-76 worktree scratch target/c76). 47 of the 68 failures the
# 2026-11-17 analysis counted were macOS-only; this set is every failed push run,
# so it holds those 47 (the per-run job split would need one API call per run).
# The lists are historical data, not files this suite reads, so they sit in a
# quote-free text blob: DocumentationDriftTest flags quoted docs/ literals as
# test-consumed documents, and these were never read. skill paths are stored
# with an uppercase SKILLS/ prefix for the same reason — the repo-wide scan of
# tests/skill_references_test.py reads `skills/…` in tracked files as current
# references, and some of these historical paths have moved since (they are
# restored to their recorded lowercase form at parse time below). One line per
# run: the run id, then the merged-delta file list.
_HISTORICAL_FAILED_RUNS_TEXT = """
44142006309 .moai/specs/SPEC-20261117-edit-diagnostics-followups/spec.md .moai/specs/SPEC-20261117-edit-diagnostics/spec.md src/edit_diagnostics.rs tests/edit_diagnostics_cli.rs
44070852987 .moai/specs/SPEC-20261117-slot-build-folder/spec.md docs/design-review/appendix/interfaces.md SKILLS/moai/references/internal/build-folder.md SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/verify.md src/heavy_slot.rs src/lane_env.rs src/main.rs tests/heavy_slot_cli.rs tests/lane_env_cli.rs tests/lane_rules_2879.rs
43968482790 AGENTS.local.md AGENTS.md
43813502221 docs/design-review/appendix/operations/project_update_plan.md docs/rust-product-path.md src/go_state.rs src/project_update.rs src/state_root.rs tests/go_settings_cli.rs tests/project_update_plan_cli.rs tests/update_end_to_end_cli.rs
43812466125 docs/design-review/appendix/operations/card.md src/card/dispatch.rs src/card/input.rs tests/card_dispatch_cli.rs
43811281841 docs/design-review/appendix/ledger_recover.md docs/design-review/appendix/operations/ledger_recover_apply.md docs/design-review/appendix/operations/ledger_recover_plan.md docs/design-review/appendix/operations/ledger_status.md docs/rust-product-path.md src/help.rs src/ledger/command.rs src/ledger/mod.rs src/ledger/open.rs src/ledger/recover.rs src/ledger/schema.rs src/ledger_recover.rs src/lib.rs src/ops/ledger_recover_apply.rs src/ops/ledger_recover_plan.rs tests/fixtures/help/all-60.txt tests/fixtures/help/all-80.txt tests/fixtures/help/all-dumb.txt tests/fixtures/help/all.txt tests/ledger_recover.rs tests/ledger_recover_cli.rs
43797628428 docs/rust-product-path.md tests/common/home.rs tests/fixtures/home_isolation/binary_without_home.txt tests/home_entry_count_test.py tests/home_isolation_scan.rs tools/home_entry_count.py tools/repo_scans.json
43769634663 docs/claude-adapter-contract.md docs/claude-owned-update.md docs/claude-preview.md docs/design-review/appendix/operations/project_update_plan.md docs/hooks-generation.md docs/rust-product-path.md src/claude.rs src/install.rs src/lib.rs src/main.rs src/project_update.rs src/status_line.rs tests/go_project_real_run_cli.rs tests/hooks_pre_p4_fixtures.rs tests/project_update_plan_cli.rs tests/repo_scans_test.py tests/status_line_cli.rs
43673152768 SKILLS/moai/references/adapters/orca.md SKILLS/moai/references/source-manifest.json tests/orca_optional_3005.rs tools/repo_scans.json
43672559962 src/help.rs src/main.rs tests/cli_name_moai.rs tests/fixtures/help/all-80.txt tests/fixtures/help/root-60.txt tests/fixtures/help/root-80.txt tests/fixtures/help/root-dumb.txt tests/fixtures/help/root.txt tests/fixtures/ui/en/error-usage-invalid.golden tests/fixtures/ui/en/error-usage-missing.golden tests/fixtures/ui/en/error-usage-unknown.golden tests/fixtures/ui/ko/error-usage-invalid.golden tests/fixtures/ui/ko/error-usage-missing.golden tests/fixtures/ui/ko/error-usage-unknown.golden tests/help_layout.rs tests/ui_korean_tty_test.py
43671971040 src/go_state.rs src/home_layout.rs src/hook_config.rs tests/go_state_cli.rs tests/home_retire_cli.rs tests/hook_dispatch_cli.rs
43670311516 SKILLS/moai/SKILL.md SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-agent.md SKILLS/moai/workflows/mode-leader.md tests/mode_workflow_steps.rs
43670163716 src/agents_local_merge.rs src/contract.rs src/go_state.rs src/main.rs src/project_update.rs tests/error_frame_tty_test.py tests/error_wording_cli.rs
43670143020 docs/design-review/appendix/operations/project_skill_apply.md docs/rust-product-path.md src/project_skills.rs tests/project_skill_apply_gate_cli.rs
43583918376 docs/design-review/appendix/operations/gate_loop.md src/gate_loop.rs src/ops/gate_loop.rs tests/gate_loop.rs
43582959619 src/help.rs src/main.rs tests/help_korean_tty_test.py
43546563292 docs/design-review/appendix/operations/home_state_plan.md src/home_state.rs src/ops/home_state_plan.rs tests/home_state_selective.rs
43543900942 src/optimize.rs tests/optimize_plan_cli.rs
43540852379 tests/heavy_check_slot_test.py tools/heavy_check_slot.py
43464560363 docs/design-review/appendix/operations/goal_block.md SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-agent.md SKILLS/moai/workflows/mode-leader.md src/goal_block.rs src/lib.rs src/ops/goal_block.rs tests/goal_block_test.py tests/mode_workflow_steps.rs tests/product_cli.rs
43462634624 docs/design-review/appendix/operations/optimize_plan.md docs/rust-product-path.md src/lib.rs src/ops/optimize_plan.rs src/optimize.rs src/worktree_gc.rs tests/optimize_plan_cli.rs
43442382726 src/contract.rs tests/revert_twice_cli.rs
43442295551 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/project.md src/agents.rs src/language.rs tests/language_setting.rs tests/product_cli.rs
43442213224 CONTRIBUTING.md
43438498219 docs/design-review/appendix/interfaces.md SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/audit.md SKILLS/moai/workflows/verify.md src/lib.rs src/main.rs src/self_audit.rs tests/check_pr_reaudit_test.py tests/check_pr_self_audit_test.py tests/check_pr_zero_selected_test.py tests/self_audit_cli.rs tests/self_audit_parity.py tools/repo_scans.json
43432289902 docs/design-review/appendix/operations/gate_check.md src/gate_collect.rs src/lib.rs src/ops/gate_check.rs tests/gate_collect.rs
43431927721 adapters/claude/install.py adapters/claude/output-styles/LICENSE adapters/claude/output-styles/NOTICE.md adapters/claude/output-styles/moai-easy.md adapters/claude/output-styles/moai-learn.md adapters/claude/output-styles/moai.md adapters/claude/settings_update.py adapters/claude/statusline.py docs/claude-adapter-contract.md docs/claude-owned-update.md docs/claude-preview.md docs/rust-product-path.md experiments/claude_settings_plan.py templates/claude/output-styles/LICENSE templates/claude/output-styles/NOTICE.md templates/claude/output-styles/moai-easy.md templates/claude/output-styles/moai-learn.md templates/claude/output-styles/moai.md tests/claude_preview_test.py tests/claude_python_adapter_removed.rs tests/claude_recovery_test.py tests/claude_settings_mirror_parity_test.py tests/claude_settings_plan_test.py tests/claude_transaction_boundary_test.py tests/claude_update_test.py tests/repo_scans_test.py tools/repo_scans.json
43431575652 src/worktree_include.rs
43430310003 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/project.md src/generation.rs tests/product_cli.rs
43430262688 docs/design-review/appendix/operations/lane_safe_to_reset.md src/lane_reset.rs tests/lane_safe_to_reset.rs
43429051093 tools/repo_scans.json
43426897993 .cargo/config.toml tests/agents_home_cli.rs tests/boundary_output_scope.rs tests/cargo_env_developer_dir.rs tests/codex_config_cli.rs tests/common/mod.rs tests/dispatch_cli.rs tests/doc_reference_boundaries.rs tests/exe_link_cli.rs tests/go_state_cli.rs tests/lane_safe_to_reset.rs tests/ledger_forge_unconnected_cli.rs tests/queue_query_cli.rs tests/removed_skill_sources.rs tests/scope_symlink_cli.rs tests/statusline_counts_refresh.rs tests/statusline_git_env.rs tests/statusline_home_state.rs tests/update_command_cli.rs tests/update_route.rs tests/update_user_layer_cli.rs tests/update_version_folders.rs tests/usage_limit_cli.rs
43426552074 tests/test_isolation_rules.rs
43425859452 .moai-cli/manifests/github-issues.json .moai-cli/reports/rc2_issue_gen.py docs/decisions-2026-11-08.md docs/spikes/pi-v1.5.0/README.md docs/spikes/pi-v1.5.0/measure_pi.py docs/worktree-strategy-2026-11-10.md
43425480778 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-leader.md tests/repo_scans_test.py tools/repo_scans.json tools/repo_scans.py
43393497616 docs/design-review/appendix/interfaces.md docs/rust-product-path.md SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-leader.md src/forge/labels.rs src/forge/mod.rs src/forge/write.rs src/main.rs templates/github/labels.yml templates/github/manifest.json tests/labels_sync.rs tests/mode_workflow_steps.rs
43388742908 src/statusline/render.rs
43386397109 src/contract.rs src/hook_engine.rs src/hooks.rs tests/hook_dispatch_cli.rs tests/hooks_cli.rs
43383431587 SKILLS/moai/references/adapters/aside.md SKILLS/moai/references/adapters/orca.md SKILLS/moai/references/source-manifest.json tests/product_cli.rs
43382861289 .github/moai/README.md docs/rust-product-path.md SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-agent.md SKILLS/moai/workflows/mode-leader.md src/github_templates.rs templates/github/ISSUE_TEMPLATE/moai-followup.yml templates/github/manifest.json tests/github_cli.rs tests/github_profile_cli.rs tests/github_templates_test.py
43381272316 docs/design-review/appendix/operations/agents_home_plan.md docs/host-support-matrix.md src/agent_defs.rs src/agent_install.rs tests/agent_defs_package.rs tests/agents_home_cli.rs tests/launcher_three_hosts.rs
43379894594 src/agent_install.rs
43376991802 REVIEW.md tests/review_md_repo.rs
43357487428 SKILLS/moai/references/agents/audit.md SKILLS/moai/references/agents/docs.md SKILLS/moai/references/agents/impl.md SKILLS/moai/references/agents/plan.md SKILLS/moai/references/agents/research.md SKILLS/moai/references/chains/index.md SKILLS/moai/references/internal/decide.md SKILLS/moai/references/internal/git-safety.md SKILLS/moai/references/internal/interview.md SKILLS/moai/references/internal/review.md SKILLS/moai/references/internal/test-design.md SKILLS/moai/references/routes.json SKILLS/moai/references/source-manifest.json src/agent_defs.rs tests/agent_defs_package.rs
43356372858 docs/design-review/appendix/operations/agents_home_apply.md docs/design-review/appendix/operations/agents_home_plan.md docs/design-review/appendix/operations/agents_home_revert.md src/agent_install.rs src/lib.rs src/ops/agents_home_apply.rs src/ops/agents_home_plan.rs src/ops/agents_home_revert.rs tests/agents_home_cli.rs
43355820983 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/project.md tests/product_cli.rs
43355783326 src/agents.rs src/apply.rs src/archive.rs src/claude.rs src/contract.rs src/decision.rs src/decision_log.rs src/forge/audit_status.rs src/forge/report.rs src/generation.rs src/github_templates.rs src/hooks.rs src/hosts/codex.rs src/hosts/codex_config.rs src/key_vault.rs src/layout.rs src/message.rs src/migrate.rs src/profile.rs src/recovery.rs src/update.rs src/worktree_gc.rs
43355514348 src/review.rs tests/review_links_cli.rs
43353177277 .github/pull_request_template.md SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/audit.md SKILLS/moai/workflows/mode-agent.md SKILLS/moai/workflows/mode-leader.md SKILLS/moai/workflows/sync.md SKILLS/moai/workflows/verify.md tests/check_pr_self_audit_test.py tools/check_pr_self_audit.py
43349413854 SKILLS/moai/references/agents/audit.md SKILLS/moai/references/agents/docs.md SKILLS/moai/references/agents/impl.md SKILLS/moai/references/agents/plan.md SKILLS/moai/references/agents/research.md SKILLS/moai/references/source-manifest.json src/agent_defs.rs src/generation.rs src/lib.rs tests/agent_defs_package.rs tests/product_cli.rs
43349122304 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/verify.md tests/check_pr_self_audit_test.py
43347893226 build.rs docs/design-review/appendix/interfaces.md docs/design-review/appendix/operations/package_inspect.md src/main.rs src/ops/mod.rs src/ops/package_inspect.rs tests/hot_file_ops_merge.rs tools/ci_scope.py
43346656734 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-agent.md SKILLS/moai/workflows/mode-leader.md src/claim_verify.rs tests/claim_verify_cli.rs tests/mode_workflow_steps.rs
43336237181 src/statusline/counts.rs src/statusline/render.rs tests/fixtures/statusline-claude/cases.json tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.ansi.txt tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.ansi256.txt tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.plain.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.ansi.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.ansi256.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.plain.txt tests/statusline_counts_refresh.rs
43333197201 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-agent.md SKILLS/moai/workflows/mode-leader.md tests/product_cli.rs
43333156739 src/statusline/counts.rs
43331030812 src/statusline/counts.rs src/statusline/state.rs tests/statusline_counts_refresh.rs
43329276140 docs/rust-product-path.md src/update.rs tests/update_command_cli.rs
43329169497 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/project.md src/project_detect.rs tests/project_detect_cli.rs
43328146722 docs/rust-product-path.md src/audit.rs tests/boundary_repro.rs
43325687166 src/statusline/render.rs tests/fixtures/statusline-claude/cases.json tests/fixtures/statusline-claude/golden/a01-minimal.ansi.txt tests/fixtures/statusline-claude/golden/a01-minimal.ansi256.txt tests/fixtures/statusline-claude/golden/a02-doc-full.ansi.txt tests/fixtures/statusline-claude/golden/a02-doc-full.ansi256.txt tests/fixtures/statusline-claude/golden/a03-context-null.ansi.txt tests/fixtures/statusline-claude/golden/a03-context-null.ansi256.txt tests/fixtures/statusline-claude/golden/a04-context-high.ansi.txt tests/fixtures/statusline-claude/golden/a04-context-high.ansi256.txt tests/fixtures/statusline-claude/golden/a05-pr-approved-effort.ansi.txt tests/fixtures/statusline-claude/golden/a05-pr-approved-effort.ansi256.txt tests/fixtures/statusline-claude/golden/a06-invalid-json.ansi.txt tests/fixtures/statusline-claude/golden/a06-invalid-json.ansi256.txt tests/fixtures/statusline-claude/golden/a08-model-string-pr-changes.ansi.txt tests/fixtures/statusline-claude/golden/a08-model-string-pr-changes.ansi256.txt tests/fixtures/statusline-claude/golden/a09-top-level-null.ansi.txt tests/fixtures/statusline-claude/golden/a09-top-level-null.ansi256.txt tests/fixtures/statusline-claude/golden/a10-unpaired-surrogate.ansi.txt tests/fixtures/statusline-claude/golden/a10-unpaired-surrogate.ansi256.txt tests/fixtures/statusline-claude/golden/a11-surrogate-session-name.ansi.txt tests/fixtures/statusline-claude/golden/a11-surrogate-session-name.ansi256.txt tests/fixtures/statusline-claude/golden/a12-invalid-utf8.ansi.txt tests/fixtures/statusline-claude/golden/a12-invalid-utf8.ansi256.txt tests/fixtures/statusline-claude/golden/a13-duplicate-top-level-keys.ansi.txt tests/fixtures/statusline-claude/golden/a13-duplicate-top-level-keys.ansi256.txt tests/fixtures/statusline-claude/golden/a14-duplicate-nested-keys.ansi.txt tests/fixtures/statusline-claude/golden/a14-duplicate-nested-keys.ansi256.txt tests/fixtures/statusline-claude/golden/a15-case-folded-keys.ansi.txt tests/fixtures/statusline-claude/golden/a15-case-folded-keys.ansi256.txt tests/fixtures/statusline-claude/golden/a16-session-five-hour-only.ansi.txt tests/fixtures/statusline-claude/golden/a16-session-five-hour-only.ansi256.txt tests/fixtures/statusline-claude/golden/a17-week-only.ansi.txt tests/fixtures/statusline-claude/golden/a17-week-only.ansi256.txt tests/fixtures/statusline-claude/golden/b01-git-clean.ansi.txt tests/fixtures/statusline-claude/golden/b01-git-clean.ansi256.txt tests/fixtures/statusline-claude/golden/b02-git-dirty-ahead.ansi.txt tests/fixtures/statusline-claude/golden/b02-git-dirty-ahead.ansi256.txt tests/fixtures/statusline-claude/golden/b03-config-latte-subset.ansi.txt tests/fixtures/statusline-claude/golden/b03-config-latte-subset.ansi256.txt tests/fixtures/statusline-claude/golden/b04-template-version.ansi.txt tests/fixtures/statusline-claude/golden/b04-template-version.ansi256.txt tests/fixtures/statusline-claude/golden/b05-task-state.ansi.txt tests/fixtures/statusline-claude/golden/b05-task-state.ansi256.txt tests/fixtures/statusline-claude/golden/b06-goal-armed.ansi.txt tests/fixtures/statusline-claude/golden/b06-goal-armed.ansi256.txt tests/fixtures/statusline-claude/golden/b07-goal-not-armed.ansi.txt tests/fixtures/statusline-claude/golden/b07-goal-not-armed.ansi256.txt tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.ansi.txt tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.ansi256.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.ansi.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.ansi256.txt tests/fixtures/statusline-claude/golden/b10-forge-suppressed.ansi.txt tests/fixtures/statusline-claude/golden/b10-forge-suppressed.ansi256.txt tests/fixtures/statusline-claude/golden/b12-session-without-bars.ansi.txt tests/fixtures/statusline-claude/golden/b12-session-without-bars.ansi256.txt
43324666231 .gitignore docs/rust-product-path.md src/profile.rs tests/profile_apply_engine.rs
43321683860 docs/rust-product-path.md src/update.rs tests/update_version_folders.rs
43317577477 docs/design-review/appendix/interfaces.md docs/rust-product-path.md src/go_state.rs src/lib.rs src/main.rs tests/go_state_cli.rs
43317529117 docs/rust-product-path.md src/claude.rs tests/claude_settings_cli.rs
43317484313 docs/release-rehearsal-2026-11-12.md tests/release_rehearsal_test.py tools/ci_scope.py
43271078539 src/statusline/render.rs tests/fixtures/statusline-claude/cases.json tests/fixtures/statusline-claude/golden/a02-doc-full.ansi.txt tests/fixtures/statusline-claude/golden/a02-doc-full.ansi256.txt tests/fixtures/statusline-claude/golden/a02-doc-full.plain.txt tests/fixtures/statusline-claude/golden/a05-pr-approved-effort.ansi.txt tests/fixtures/statusline-claude/golden/a05-pr-approved-effort.ansi256.txt tests/fixtures/statusline-claude/golden/a05-pr-approved-effort.plain.txt tests/fixtures/statusline-claude/golden/a11-surrogate-session-name.ansi.txt tests/fixtures/statusline-claude/golden/a11-surrogate-session-name.ansi256.txt tests/fixtures/statusline-claude/golden/a11-surrogate-session-name.plain.txt tests/fixtures/statusline-claude/golden/b03-config-latte-subset.ansi.txt tests/fixtures/statusline-claude/golden/b03-config-latte-subset.ansi256.txt tests/fixtures/statusline-claude/golden/b03-config-latte-subset.plain.txt tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.ansi.txt tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.ansi256.txt tests/fixtures/statusline-claude/golden/b08-forge-zero-counts.plain.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.ansi.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.ansi256.txt tests/fixtures/statusline-claude/golden/b09-forge-counts-hidden-fields-enabled.plain.txt tests/fixtures/statusline-claude/golden/b10-forge-suppressed.ansi.txt tests/fixtures/statusline-claude/golden/b10-forge-suppressed.ansi256.txt tests/fixtures/statusline-claude/golden/b10-forge-suppressed.plain.txt tests/fixtures/statusline-claude/golden/b12-session-without-bars.ansi.txt tests/fixtures/statusline-claude/golden/b12-session-without-bars.ansi256.txt tests/fixtures/statusline-claude/golden/b12-session-without-bars.plain.txt
43271018829 SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/verify.md tests/heavy_check_slot_test.py tools/heavy_check_slot.py
43266681360 .cargo/config.toml .github/moai/README.md .github/pull_request_template.md .github/workflows/ci.yml .github/workflows/rc-ci.yml CONTRIBUTING.md docs/queue-query.md docs/release-tag-ruleset-moai-cli-v.md docs/rust-product-path.md docs/support-matrix.md rust-toolchain.toml SKILLS/moai/references/source-manifest.json SKILLS/moai/workflows/mode-agent.md SKILLS/moai/workflows/verify.md src/flow.rs src/leader.rs src/main.rs src/queue_query.rs src/scope.rs tests/check_pr_self_audit_test.py tests/ci_install_ripgrep_test.py tests/ci_scope_test.py tests/flow_cli.rs tests/github_assets_test.py tests/leader_cli.rs tests/queue_query_cli.rs tools/check_pr_self_audit.py
"""

HISTORICAL_FAILED_RUNS = [
    (parts[0], parts[1:])
    for parts in (
        line.split()
        for line in _HISTORICAL_FAILED_RUNS_TEXT.replace("SKILLS/", "skills/").strip().splitlines()
    )
]

# AC-4: the runs the frozen rule table leaves linux-only, named by run id. Each
# one's file list falls outside the six measured categories (skills text, ledger
# and gate logic, docs); some of the 47 macOS-only failures may sit here — the
# gap is recorded in the SPEC's Result, and widening the table is a follow-up.
HISTORICAL_FALSE_RUN_IDS = frozenset({
    "44142006309", "43968482790", "43812466125", "43811281841", "43769634663",
    "43673152768", "43670311516", "43670143020", "43583918376", "43543900942",
    "43464560363", "43442382726", "43442295551", "43442213224", "43438498219",
    "43432289902", "43431927721", "43431575652", "43430310003", "43430262688",
    "43429051093", "43426552074", "43425859452", "43425480778", "43393497616",
    "43383431587", "43382861289", "43376991802", "43357487428", "43355820983",
    "43355514348", "43353177277", "43349413854", "43349122304", "43346656734",
    "43333197201", "43329169497", "43328146722", "43324666231", "43317529117",
})


class CiScopePlatformTest(unittest.TestCase):
    """C-76 (SPEC-20261119-macos-on-pr): platform-sensitive path classification."""

    def setUp(self):
        self.ci_scope = load_module()

    def test_platform_sensitive_paths_need_macos(self):
        # One representative per measured category; the reason names the category.
        cases = {
            "pty/terminal": ["src/launcher_terminal.rs"],
            "pty/terminal (in-process pty)": ["tests/help_layout.rs"],
            "pty/terminal (tty test)": ["tests/ui_summary_tty_test.py"],
            "process/signal": ["src/heavy_slot.rs"],
            "process/signal (tool)": ["tools/heavy_check_slot.py"],
            "fs links/permissions": ["src/worktree_gc.rs"],
            "fs links/permissions (test)": ["tests/exe_link_cli.rs"],
            "fs links/permissions (home layout)": ["src/home_layout.rs"],
            "fs links/permissions (go state)": ["src/go_state.rs"],
            "fs links/permissions (version folders)": ["src/versioning.rs"],
            "fs links/permissions (agents home)": ["src/ops/agents_home_apply.rs"],
            "launcher": ["src/launcher_init.rs"],
            "launcher (test)": ["tests/launcher_terminal_output.rs"],
            "launcher (omp)": ["tests/omp_launcher.rs"],
            "statusline": ["src/statusline/render.rs"],
            "statusline (test)": ["tests/statusline_cards.rs"],
            "statusline (fixture)": ["tests/fixtures/statusline-claude/cases.json"],
            "build/toolchain": ["rust-toolchain.toml"],
            "build/toolchain (cargo)": ["Cargo.lock"],
            "build/toolchain (workflow)": [".github/workflows/ci.yml"],
            "build/toolchain (ci tool)": ["tools/ci_scope.py"],
            "build/toolchain (installer)": ["install/install.sh"],
        }
        for label, paths in cases.items():
            needed, reasons = self.ci_scope.macos_needed(paths)
            self.assertTrue(needed, label)
            self.assertTrue(reasons, label)

    def test_non_platform_paths_stay_linux_only(self):
        for paths in (
            ["docs/llms.txt"],
            ["docs/a.md", "docs/img/b.png"],
            ["src/contract.rs"],
            ["skills/moai/workflows/verify.md"],
            ["tools/repo_scans.json"],
            ["AGENTS.md"],
            ["src/gate_loop.rs", "docs/rust-product-path.md"],
        ):
            needed, reasons = self.ci_scope.macos_needed(paths)
            self.assertFalse(needed, paths)
            self.assertEqual(reasons, [], paths)

    def test_platform_prefix_does_not_leak_outside_the_family(self):
        # A similar name outside a family must not match: the statusline prefix is
        # the directory, not a name stem (docs/old.md is a synthetic path), and the
        # Python adapter's statusline is not the Rust statusline module.
        needed, _ = self.ci_scope.macos_needed(
            ["src/statusline_notes.rs", "docs/old.md", "adapters/claude/statusline.py"]
        )
        self.assertFalse(needed)

    def test_empty_input_needs_macos(self):
        needed, reasons = self.ci_scope.macos_needed([])
        self.assertTrue(needed)
        self.assertTrue(reasons)
        proc = self.run_proc("\n")
        data = json.loads(proc.stdout)
        self.assertTrue(data["macos_needed"])

    def test_unusable_path_fails_safe_to_macos(self):
        # A path with control characters cannot be judged: fail safe to the wider
        # matrix (same direction as the docs-only fail-safe).
        needed, reasons = self.ci_scope.macos_needed(["src/ok.rs", "bad\x00path"])
        self.assertTrue(needed)
        self.assertTrue(reasons)
        proc = self.run_proc("not a real path\n\x00weird\n")
        data = json.loads(proc.stdout)
        self.assertTrue(data["macos_needed"])

    def test_json_carries_platform_keys(self):
        proc = self.run_proc("docs/a.md\n")
        data = json.loads(proc.stdout)
        self.assertFalse(data["macos_needed"])
        self.assertEqual(data["macos_reasons"], [])
        proc = self.run_proc("src/statusline/render.rs\n")
        data = json.loads(proc.stdout)
        self.assertTrue(data["macos_needed"])
        self.assertTrue(any("statusline" in r for r in data["macos_reasons"]))

    def test_internal_error_fails_safe_to_macos(self):
        from io import StringIO
        from unittest import mock

        module = load_module()
        with mock.patch.object(module, "macos_needed", side_effect=RuntimeError("boom")):
            with mock.patch.object(module.sys, "stdin", StringIO("docs/a.md\n")):
                with mock.patch.object(module.sys, "stdout", StringIO()) as out:
                    code = module.main()
        self.assertEqual(code, 0)
        data = json.loads(out.getvalue())
        self.assertTrue(data["macos_needed"])
        self.assertTrue(any("boom" in r for r in data["macos_reasons"]))

    def test_historical_failed_merge_file_lists(self):
        # AC-4: every historical failed push run's file list is classified; the
        # linux-only ones are exactly the named set above, no silent drops.
        module = load_module()
        self.assertEqual(len(HISTORICAL_FAILED_RUNS), 69)
        got_false = {
            run_id
            for run_id, files in HISTORICAL_FAILED_RUNS
            if not module.macos_needed(files)[0]
        }
        self.assertEqual(got_false, set(HISTORICAL_FALSE_RUN_IDS))

    def run_proc(self, text):
        import subprocess

        return subprocess.run(
            [sys.executable, os.path.join(REPO_ROOT, "tools", "ci_scope.py")],
            input=text,
            capture_output=True,
            text=True,
        )


class CiDecideScriptTest(unittest.TestCase):
    """C-76 re-audit B-3 (and N-5, N-7): run the targets job's decide step itself.

    Text pins cannot bind the decided matrix to the job output — `$matrix`
    flipped to `$linux` passed every earlier test. Here the decide step's
    script is extracted from ci.yml, its `${{ … }}` expressions are filled
    with the case's values, and it runs with `bash` in a temporary git
    repository, exactly as the runner would. The `matrix=` line it appends to
    $GITHUB_OUTPUT is the assertion.
    """

    FULL = (
        '[{"target":"linux-x86_64","runner":"ubuntu-24.04","timeout":30,"leg":"nextest"},'
        '{"target":"linux-x86_64","runner":"ubuntu-24.04","timeout":30,"leg":"rest"},'
        '{"target":"macos-arm64","runner":"macos-15","timeout":45,"leg":"nextest"},'
        '{"target":"macos-arm64","runner":"macos-15","timeout":45,"leg":"rest"}]'
    )
    LINUX = (
        '[{"target":"linux-x86_64","runner":"ubuntu-24.04","timeout":30,"leg":"nextest"},'
        '{"target":"linux-x86_64","runner":"ubuntu-24.04","timeout":30,"leg":"rest"}]'
    )
    # SPEC-20261120-release-pr-three-os: the release list is today's full list plus the
    # two windows-x86_64 legs — written as the join the decide step performs
    # (`release="${full%]},${windows}]"`), so the constant and the script cannot drift.
    RELEASE = (
        FULL[:-1]
        + ','
        + '{"target":"windows-x86_64","runner":"windows-2025","timeout":45,"leg":"nextest"},'
        + '{"target":"windows-x86_64","runner":"windows-2025","timeout":45,"leg":"rest"}]'
    )

    @classmethod
    def setUpClass(cls):
        import yaml

        with open(
            os.path.join(REPO_ROOT, ".github", "workflows", "ci.yml"),
            encoding="utf-8",
        ) as fh:
            flow = yaml.safe_load(fh)
        cls.script = next(
            s for s in flow["jobs"]["targets"]["steps"] if s.get("id") == "decide"
        )["run"]

    @staticmethod
    def git(repo, *args):
        subprocess.run(
            ["git", *args], cwd=repo, check=True, capture_output=True
        )

    def repo(self, changes, classifier=None):
        """A git repository with a base commit and a head commit changing
        `changes` (a path-to-body map); `classifier` overrides tools/ci_scope.py.
        Returns (tmpdir, base_sha, head_sha)."""
        tmp = tempfile.mkdtemp(prefix="c76-decide-")
        self.addCleanup(shutil.rmtree, tmp, ignore_errors=True)
        tools = os.path.join(tmp, "tools")
        os.makedirs(tools)
        with open(os.path.join(tools, "ci_scope.py"), "w", encoding="utf-8") as fh:
            if classifier is not None:
                fh.write(classifier)
            else:
                shutil.copyfile(
                    os.path.join(REPO_ROOT, "tools", "ci_scope.py"), fh.name
                )
        self.git(tmp, "init", "-q", "-b", "main")
        self.git(tmp, "config", "user.name", "fixture")
        self.git(tmp, "config", "user.email", "fixture@example.invalid")
        with open(os.path.join(tmp, "seed.txt"), "w", encoding="utf-8") as fh:
            fh.write("base\n")
        self.git(tmp, "add", "-A")
        self.git(tmp, "commit", "-q", "-m", "base")
        for rel, body in changes.items():
            path = os.path.join(tmp, rel)
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, "w", encoding="utf-8") as fh:
                fh.write(body)
        self.git(tmp, "add", "-A")
        self.git(tmp, "commit", "-q", "--allow-empty", "-m", "head")
        base = subprocess.run(
            ["git", "rev-parse", "HEAD~1"], cwd=tmp, capture_output=True, text=True
        ).stdout.strip()
        head = subprocess.run(
            ["git", "rev-parse", "HEAD"], cwd=tmp, capture_output=True, text=True
        ).stdout.strip()
        names = sorted(changes)
        return tmp, base, head, names

    def decide(self, repo, event, base, head, gh_files=None, gh_total=None,
               gh_broken=False, head_ref=None):
        """Run the decide step as the runner would; return (proc, matrix). On a
        pull_request a fake `gh` answers the files API from `gh_files`
        (`gh_total` overrides the true total; `gh_broken` makes gh fail), the
        way the runner's token would reach the real API."""
        script = self.script.replace("${{ github.event_name }}", event)
        script = script.replace("${{ github.event.pull_request.base.sha }}", base)
        script = script.replace("${{ github.event.pull_request.number }}", "7")
        script = script.replace("${{ github.repository }}", "owner/repo")
        script = script.replace("${{ github.token }}", "fake-token")
        out = os.path.join(repo, "github-output")
        if os.path.exists(out):
            os.remove(out)
        runner_temp = os.path.join(repo, "runner-temp")
        os.makedirs(runner_temp, exist_ok=True)
        env = dict(
            os.environ,
            GITHUB_OUTPUT=out,
            GITHUB_SHA=head,
            RUNNER_TEMP=runner_temp,
            PR_NUMBER="7",
            REPO="owner/repo",
            GH_TOKEN="fake-token",
        )
        if head_ref is not None:
            # The decide step receives the head branch name through env: (SPEC-20261120-
            # release-pr-three-os); the fake runner hands it over the same way.
            env["HEAD_REF"] = head_ref
        if gh_files is not None or gh_broken:
            # A bare path list is a shorthand for plain items; renames pass items
            # with `previous_filename` directly.
            if gh_files and isinstance(gh_files[0], str):
                gh_files = [{"filename": p} for p in gh_files]
            bindir = os.path.join(repo, "fakebin")
            os.makedirs(bindir, exist_ok=True)
            gh = os.path.join(bindir, "gh")
            # The fake runs each call's `--jq` expression through the real jq
            # when one is on PATH — the expression itself is judged, so a mutant
            # that drops `previous_filename` from the files expression fails the
            # rename scene (the N2 follow-up). Without jq it falls back to an
            # emulation that splits filename and previous_filename the same way,
            # and the expression itself is then not judged. `gh_broken` names
            # which call fails: "files", "count" (the item count), "view" (the
            # total) or "all" (True keeps working as "all").
            with open(gh, "w", encoding="utf-8") as fh:
                fh.write(
                    "#!/usr/bin/env python3\n"
                    "import json, os, shutil, subprocess, sys\n"
                    "args = sys.argv[1:]\n"
                    "expr = args[args.index('--jq') + 1] if '--jq' in args else ''\n"
                    "is_view = 'view' in args\n"
                    "is_api = 'api' in args\n"
                    "broken = os.environ.get('GH_FAKE_BROKEN', '')\n"
                    "if (broken == 'all' or (broken == 'api' and is_api)\n"
                    "        or (broken == 'count' and is_api and 'length' in expr)\n"
                    "        or (broken == 'files' and is_api and 'length' not in expr)\n"
                    "        or (broken == 'view' and is_view)):\n"
                    "    sys.exit(1)\n"
                    "items = os.environ.get('GH_FAKE_FILES', '[]')\n"
                    "jq = shutil.which('jq')\n"
                    "if jq is None:\n"
                    "    if is_view:\n"
                    "        print(os.environ.get('GH_FAKE_TOTAL', '0'))\n"
                    "    elif is_api:\n"
                    "        for it in json.loads(items):\n"
                    "            print(it['filename'])\n"
                    "            if it.get('previous_filename'):\n"
                    "                print(it['previous_filename'])\n"
                    "    else:\n"
                    "        sys.exit(1)\n"
                    "    sys.exit(0)\n"
                    "if is_view:\n"
                    "    payload = json.dumps({'changedFiles': int(os.environ.get('GH_FAKE_TOTAL', '0'))})\n"
                    "elif is_api:\n"
                    "    payload = items\n"
                    "else:\n"
                    "    sys.exit(1)\n"
                    "run = subprocess.run([jq, '-r', expr], input=payload,\n"
                    "                     capture_output=True, text=True)\n"
                    "sys.stdout.write(run.stdout)\n"
                    "sys.stderr.write(run.stderr)\n"
                    "sys.exit(run.returncode)\n"
                )
            os.chmod(gh, 0o755)
            env["PATH"] = bindir + os.pathsep + env["PATH"]
            env["GH_FAKE_FILES"] = json.dumps(gh_files or [])
            env["GH_FAKE_TOTAL"] = str(len(gh_files or []) if gh_total is None else gh_total)
            if gh_broken is True:
                gh_broken = "all"
            elif gh_broken is False:
                gh_broken = ""
            env["GH_FAKE_BROKEN"] = str(gh_broken)
        # The runner executes the step with `bash --noprofile --norc -eo pipefail`,
        # so the decide script runs under exactly that shell here too (audit: a
        # plainer shell hid that a failing call kills the step instead of failing
        # safe).
        proc = subprocess.run(
            ["bash", "--noprofile", "--norc", "-eo", "pipefail", "-c", script],
            cwd=repo, env=env, capture_output=True, text=True,
        )
        matrix = None
        if os.path.exists(out):
            with open(out, encoding="utf-8") as fh:
                for line in fh:
                    if line.startswith("matrix="):
                        matrix = line.rstrip("\n")[len("matrix="):]
        return proc, matrix

    def test_push_and_dispatch_get_the_full_matrix(self):
        for event in ("push", "workflow_dispatch"):
            with self.subTest(event=event):
                repo, base, head, _names = self.repo({"docs/a.md": "x\n"})
                proc, matrix = self.decide(repo, event, base, head)
                self.assertEqual(proc.returncode, 0, proc.stderr)
                self.assertEqual(matrix, self.FULL)

    def test_a_platform_sensitive_pr_gets_macos(self):
        repo, base, head, names = self.repo({"src/statusline/render.rs": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_files=names
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)
        self.assertIn("target-reason:", proc.stdout)

    def test_a_plain_pr_stays_linux_only(self):
        repo, base, head, names = self.repo({"docs/a.md": "x\n", "README.md": "x\n"})
        proc, matrix = self.decide(repo, "pull_request", base, head, gh_files=names)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.LINUX)

    def test_a_broken_classifier_falls_back_to_the_full_matrix(self):
        # The classifier exits 1: the decide step must still succeed with the
        # full matrix (fail safe), not fail or narrow to linux-only.
        repo, base, head, names = self.repo(
            {"docs/a.md": "x\n"}, classifier="import sys\nsys.exit(1)\n"
        )
        proc, matrix = self.decide(repo, "pull_request", base, head, gh_files=names)
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)

    def test_a_truncated_files_api_answer_falls_back_to_the_full_matrix(self):
        # The files API serves at most 3000 files; a truncated listing must not
        # judge a partial list — the fail-safe full matrix runs instead.
        repo, base, head, names = self.repo({"src/statusline/render.rs": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_files=names, gh_total=len(names) + 1
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)
        self.assertIn("fail-safe", proc.stderr)

    def test_an_empty_files_api_answer_falls_back_to_the_full_matrix(self):
        repo, base, head, _names = self.repo({"docs/a.md": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_files=[], gh_total=0
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)

    def test_a_broken_files_api_falls_back_to_the_full_matrix(self):
        repo, base, head, _names = self.repo({"src/statusline/render.rs": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_broken=True
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)

    def test_a_rename_serves_both_paths_from_the_jq(self):
        # The fake gh applies the decide step's own `--jq`: a rename item yields
        # its new and its old name, so both paths reach the judge.
        repo, base, head, _names = self.repo({"docs/a.md": "x\n"})
        proc, _matrix = self.decide(
            repo, "pull_request", base, head,
            gh_files=[
                {"filename": "notes/new.md", "previous_filename": "notes/old.md"},
                {"filename": "docs/a.md"},
            ],
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        with open(
            os.path.join(repo, "runner-temp", "ci-scope-files.txt"),
            encoding="utf-8",
        ) as fh:
            paths = set(fh.read().splitlines())
        self.assertIn("notes/new.md", paths)
        self.assertIn("notes/old.md", paths)

    def test_the_truncation_guard_counts_api_items_not_path_lines(self):
        # audit B1: a rename serves two path lines for one API item. Three items
        # (two renames) against a true total of four must read as truncated —
        # counting lines would report 6 >= 4 and judge a partial list. A
        # linux-only partial list judged complete would narrow the matrix.
        repo, base, head, _names = self.repo({"docs/a.md": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head,
            gh_files=[
                {"filename": "a/new1", "previous_filename": "a/old1"},
                {"filename": "a/new2", "previous_filename": "a/old2"},
                {"filename": "docs/a.md"},
            ],
            gh_total=4,
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)
        self.assertIn("served 3 of 4", proc.stderr)

    def test_a_failed_files_call_falls_back_in_the_runner_shell(self):
        # The runner runs the step with -e and pipefail: a failing gh call must
        # land in the fail-safe full matrix, not kill the step.
        repo, base, head, names = self.repo({"src/statusline/render.rs": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_files=names, gh_broken="files"
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)

    def test_a_failed_item_count_call_falls_back_in_the_runner_shell(self):
        repo, base, head, names = self.repo({"src/statusline/render.rs": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_files=names, gh_broken="count"
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)

    def test_a_failed_total_call_falls_back_in_the_runner_shell(self):
        repo, base, head, names = self.repo({"src/statusline/render.rs": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_files=names, gh_broken="view"
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.FULL)

    def test_a_release_head_branch_pins_the_three_os_set(self):
        """SPEC-20261120-release-pr-three-os: a pull_request whose head branch name starts
        with release/v decides the joined release list — the step itself runs here, so the
        guard is bound to its output, not its text. The release PR skips the classifier
        entirely: the same docs-only change that narrows an ordinary PR stays the six-leg
        list."""
        for head_ref in ("release/v4.2647.10", "release/v4.2647.11"):
            with self.subTest(head_ref=head_ref):
                repo, base, head, names = self.repo({"docs/a.md": "x\n"})
                proc, matrix = self.decide(
                    repo, "pull_request", base, head, gh_files=names,
                    head_ref=head_ref,
                )
                self.assertEqual(proc.returncode, 0, proc.stderr)
                self.assertEqual(matrix, self.RELEASE)
                entries = json.loads(matrix)
                self.assertEqual(
                    sum(1 for e in entries if e["target"] == "windows-x86_64"), 2
                )
                self.assertEqual(
                    {e["runner"] for e in entries if e["target"] == "windows-x86_64"},
                    {"windows-2025"},
                )

    def test_a_non_release_head_branch_keeps_todays_narrowing(self):
        # Any head branch name but release/v* keeps today's rule: the same
        # docs-only change narrows to linux through the classifier.
        repo, base, head, names = self.repo({"docs/a.md": "x\n"})
        proc, matrix = self.decide(
            repo, "pull_request", base, head, gh_files=names,
            head_ref="feature/plain-work",
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertEqual(matrix, self.LINUX)


class CiWorkflowMatrixContractTest(unittest.TestCase):
    """C-76: the ci.yml matrix contract — a `targets` job decides, `check`
    consumes, the aggregate keeps the name `ci` and the push branch unchanged."""

    @classmethod
    def setUpClass(cls):
        with open(
            os.path.join(REPO_ROOT, ".github", "workflows", "ci.yml"),
            "r",
            encoding="utf-8",
        ) as fh:
            cls.yaml_text = fh.read()

    def test_targets_job_decides_the_matrix(self):
        self.assertIn("  targets:", self.yaml_text)
        self.assertIn("matrix: ${{ steps.decide.outputs.matrix }}", self.yaml_text)
        self.assertIn("macos_needed", self.yaml_text)

    def test_check_consumes_the_targets_output(self):
        self.assertEqual(
            self.yaml_text.count("fromJSON(needs.targets.outputs.matrix)"), 1
        )
        self.assertIn("needs: [targets]", self.yaml_text)
        # The old event ternary no longer picks the matrix inside `check`.
        self.assertNotIn(
            "github.event_name == 'pull_request'\n          && '[{\"target\":\"linux-x86_64\"",
            self.yaml_text,
        )

    def test_full_matrix_literal_lives_in_targets(self):
        full = (
            '[{"target":"linux-x86_64","runner":"ubuntu-24.04","timeout":30,"leg":"nextest"},'
            '{"target":"linux-x86_64","runner":"ubuntu-24.04","timeout":30,"leg":"rest"},'
            '{"target":"macos-arm64","runner":"macos-15","timeout":45,"leg":"nextest"},'
            '{"target":"macos-arm64","runner":"macos-15","timeout":45,"leg":"rest"}]'
        )
        self.assertEqual(self.yaml_text.count(full), 1)

    def test_aggregate_ci_name_and_needs_unchanged(self):
        self.assertIn("\n  ci:\n", self.yaml_text)
        self.assertIn("needs: [check, ci-verdict]", self.yaml_text)
        self.assertIn(
            "github.event_name == 'push' && needs.check.result == 'cancelled'",
            self.yaml_text,
        )
        self.assertIn("name: ci", self.yaml_text)

    def test_windows_build_still_push_only(self):
        # SPEC-20261120-release-pr-three-os: the condition gains one release-PR branch —
        # push and dispatch keep running it (their head_ref is empty, so the startswith
        # test never fires there), and an ordinary PR still skips the job.
        self.assertIn(
            "if: ${{ github.event_name != 'pull_request' "
            "|| startsWith(github.head_ref, 'release/v') }}",
            self.yaml_text,
        )

    def test_targets_paths_come_from_the_files_api(self):
        # The targets changed paths pass rename sources too — through the files
        # API (filename AND previous_filename), the base copy of the scope tool
        # judging them. The two docs Scope steps keep their rename-safe git diffs.
        self.assertEqual(
            self.yaml_text.count("git diff --name-only --no-renames"), 2
        )
        self.assertIn("previous_filename", self.yaml_text)
        self.assertIn(
            'gh api "repos/$REPO/pulls/$PR_NUMBER/files?per_page=100" --paginate',
            self.yaml_text,
        )

    def test_decide_script_pins_its_fail_safe_branches(self):
        # The branches a text mutation could silently invert: linux-only is
        # chosen only when the classifier answered (ok=1) with macos False, and a
        # failed diff never fails the job (matrix stays the full one).
        self.assertIn(
            'if [ "$ok" = 1 ] && [ "$macos" = "False" ]; then', self.yaml_text
        )
        self.assertIn("full matrix (fail-safe)", self.yaml_text)
        self.assertIn('matrix="$full"', self.yaml_text)

    def test_decide_script_classifies_only_pull_requests(self):
        # Audit B-2: only a pull_request run enters the classifier, and the diff
        # base is the PR's base commit. `if true` or a negated comparison keeps
        # both matrix literals and every other test green, so the exact guard is
        # pinned here too (rc_ci_matrix in tests/github_assets_test.py pins the
        # same line where it reads the lists).
        self.assertIn(
            'if [ "${{ github.event_name }}" = "pull_request" ]; then',
            self.yaml_text,
        )
        self.assertIn('${{ github.event.pull_request.base.sha }}', self.yaml_text)
        self.assertNotIn('github.event_name }}" != ', self.yaml_text)
        self.assertNotIn('if true; then', self.yaml_text)


if __name__ == "__main__":
    unittest.main()
