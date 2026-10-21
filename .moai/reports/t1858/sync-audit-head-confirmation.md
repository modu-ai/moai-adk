auditor-model: claude-sonnet-5-5

verdict: PASS
audited_sha: 90c92ea98dde923ddfbebc9cff96f54a17931e3e
required_backend: codex pass
convergence_overall: pass

# t1858 HEAD-confirmation audit (absorb delta only)

Scope: confirm merged HEAD 90c92ea98 still carries exactly the change audited PASS-WITH-DEBT at 55cac554f. Closed defects F1-F6 and debts M1-M3 are not re-judged.

Audit plan (`moai verify audit-plan --project-root <tree>`): config_status ok, model codex, claude off, codex required, glm off, cross_model_active true, enforced_required ["codex"].

## Claim
1. H1: the card change is intact; code files are byte-identical to the audited commit.
2. H2: the CHANGELOG conflict was resolved keeping both entries, unmodified, with no conflict markers.
3. H3: nothing main changed is read by the card's code or tests; no new forbidden literal.
4. H4: scoped tests, gofmt, lint and required-checks validation are green at the merged HEAD.
5. Required codex backend passes.

## Evidence (verbatim, trimmed to deciding lines)
H1 `git diff --stat main HEAD`: 13 files, 551 insertions, 11 deletions. Files: .github/required-checks.yml, .github/test-input-filters.yml, .moai/docs/batch-failure-bisect-runbook.md, .moai/reports/t1858/{card-review,progress,sync-audit,verdict}.md, CHANGELOG.md, internal/config/required_checks.go, internal/config/required_checks_release_gate_test.go, internal/template/release_ci_gate_ssot_test.go, internal/template/test_input_filter_parity_test.go, scripts/release.sh. Only the card's own files.

H1 `git diff 55cac554f56b02bd4ace7e6c29952fa952bb1844 HEAD -- <the 7 code files> | wc -c` -> `0` (byte-identical).

H2 `git grep -n '^<<<<<<<\|^>>>>>>>'` hits only docs-site/content/{en,ja,ko,...}/worktree/{examples,faq}.md (e.g. `<<<<<<< HEAD` ... `>>>>>>> feature/SPEC-AUTH-001`), which are documentation examples of a conflict, not unresolved merges. Excluding docs-site/: no output. `grep -c '^<<<<<<<\|^>>>>>>>\|^=======$' CHANGELOG.md` -> `0`. `git diff main HEAD -- CHANGELOG.md` shows exactly one added line (the t1858 entry) directly above the t1854 entry. t1854 entry: `git show main:CHANGELOG.md | grep 'card t1854' | md5` = `3d573ed8cf84edbb8c480778548cb66a`, HEAD = `3d573ed8cf84edbb8c480778548cb66a`; count 1 each; t1858 entry count 1.

H3 `git diff --name-only 3642a3a9b main` (41 files): .claude/agents|rules|skills, .codex/agents manager-git, .moai/reports t1837/t1854, SPEC-LANDING-SURFACE-RULES-001, AGENTS.local.md, CHANGELOG.md, internal/auditreceipt/store.go, internal/homestate/card_audit_kickoff.go + test, internal/hook/audit_receipt_guard.go + test, internal/template/catalog.yaml, and template copies of manager-git/manager-spec/spec-frontmatter-schema/delivery-policy/factory-dispatch-*/spec-workflow/worktree-integration*/sync delivery. `... | grep -E 'scripts/|\.github/|internal/config|required_checks|test-input'` -> no output. `--diff-filter=A` for scripts/ -> no output. So no main-changed file is under scripts/, .github/, or internal/config.
`git grep -n 'Test (ubuntu-latest)' -- scripts internal/cli/branch_protection.go internal/cli/github_init.go` non-test hits: none; the only hits are scripts/ci-mirror/test/validate-required-checks_test.sh and scripts/ci-watch/test/run_test.sh fixtures, not brought by main (main touches neither).
`go test ./internal/config -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/config	15.369s` (includes TestNoHardcodedContexts).

H4:
- `go test ./internal/template -run 'TestReleaseScript|TestTestInputFilters|BranchProtection|TestNoHardcodedContexts' -count=1 -v`: PASS for TestReleaseScriptCIGateSSoTFailsClosed (24.16s), TestReleaseScriptRealSSoTDemandsEveryPinnedCheck, TestReleaseScriptCIGate, TestReleaseScriptCIGateWithoutGh, TestReleaseScriptTagsTheVerifiedCommit, TestReleaseScriptCharacterization, TestReleaseScriptHeadAdmission, TestReleaseScriptRcChangelogSkip, TestReleaseScriptMatrixGate, TestTestInputFiltersSingleSource, TestTestInputFiltersCoversGoTestInputs, TestBranchProtectionParity, TestBranchProtectionDetectFilterCoversParityInput; `ok  github.com/modu-ai/moai-adk/internal/template	71.254s`. (TestNoHardcodedContexts lives in internal/config and passed above.)
- `gofmt -l internal/config internal/template` -> no output.
- `golangci-lint run ./internal/config/ ./internal/template/` -> `0 issues.`
- `scripts/ci-mirror/validate-required-checks.sh` -> `✅ All validations passed` (Dimensions A/B/C green).

Cross-model: `mcp__moai__audit_multi` (card_id t1858, baseBranch, project_root = tree, focus as given):
overall_verdict: pass; gate_unmet: absent; plan_source: config; per_backend_verdicts: [{backend: codex, gate: required, verdict: pass, summary: "Verdict: pass", findings: [], next_steps: []}]; disagreement_flag null; fail_open_backends []; audit_receipt rcpt-dddc850b167379461c7d7aa8. This verdict is not final until the orchestrator's `moai verify audit-plan --result-file` check passes.

## Baseline-attribution
Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858, HEAD 90c92ea98dde923ddfbebc9cff96f54a17931e3e, all commands run in this audit run. Judging build for the moai MCP: ee981a932 (v3.2.0-rc.36-t1822c), reported by the tool as an ancestor of HEAD; the CLI/MCP output (audit-plan, audit_multi) was produced by that older build and the tool itself says to rebuild before trusting CLI results. Go tests, gofmt, golangci-lint and the validate script do not depend on the moai binary.

## Gaps
- The moai binary lag above: audit-plan and audit_multi ran on a build older than HEAD; the plan content matched what the lane stated, and codex judged the tree diff itself.
- Heavy lane re-measure and the full suite were deliberately not run (separate go-test-heavy slot); only the scoped tests listed were run.
- Live GitHub behaviour of release.sh was not exercised (tests use fakes, as in the audited round).
- TestNoHardcodedContexts was observed via the whole `internal/config` package run, not a separate `-run` line.
- Absorb delta of non-card files (rules, templates, hooks) was judged only for overlap with the card's inputs, not reviewed on its own merits.

## Residual-risk
Low. Main's changes touch no path the card reads; the only shared file (CHANGELOG.md) is verified entry-for-entry. A future main change to .github/required-checks.yml before landing would need a fresh confirmation.

## Scores
Functionality 1.00, Security 0.95, Craft 0.95, Consistency 0.95.

## Defect list
None. No BLOCKING finding (no regression from the absorb or from any code change since 55cac554f). The previously recorded debts M1-M3 remain as in sync-audit.md and are not re-judged.

AUDIT-VERDICT: PASS spec=t1858 receipts=rcpt-dddc850b167379461c7d7aa8
