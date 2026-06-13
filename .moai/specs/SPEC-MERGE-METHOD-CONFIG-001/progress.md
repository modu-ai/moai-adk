# Progress — SPEC-MERGE-METHOD-CONFIG-001

> Tier M. GitHub Issue #1061 — configurable sync-phase PR merge method (`git_strategy.<mode>.merge_method`, one of squash/merge/rebase, default `squash`). Built on top of the #1064 git_strategy Save isolation fix (the `merge_method` field rides the same git_strategy section through the dirty-flag gate).

## §E.1 Run-phase Milestones

| Milestone | Description | Status |
|-----------|-------------|--------|
| M1 | `merge_method` config field in git-strategy.yaml per-mode (default `squash`) — defaults.go + types.go + template | DONE |
| M2 | loader/validator wiring — enum `squash\|merge\|rebase` via the `validateGitConventionConfig` enum pattern (not `checkStringField`), field-path error | DONE |
| M3 | `internal/github` MergeMethod abstraction wired conceptually (config value set ≡ {squash,merge,rebase}) | DONE |
| M4 | 3 hardcoded `gh pr merge --squash` template/doc sites updated to honor config (delivery.md, manager-git.md, moai-ref-git-workflow/skill.md) | DONE |
| M5 | FROZEN `spec-workflow.md` lifecycle-table widening (squash → configured `merge_method` default squash) — local copy + template mirror byte-identical | DONE |
| M6 | catalog.yaml hash cascade + neutrality + mirror-parity gates | DONE |

## §E.2 Run-phase Evidence

13/13 AC PASS (acceptance.md SSOT, AC-MMC-001..013). Highlights:

- AC-MMC-001 default `squash` all 3 modes; AC-MMC-002/003/004 loader absent/present/partial; AC-MMC-005/006 enum validation + field-path + empty fail-safe.
- AC-MMC-007/008 template sites honor config, squash-default renders `gh pr merge --squash --delete-branch` byte-equivalent.
- AC-MMC-009 neutrality, AC-MMC-010 FROZEN SSOT+mirror byte-identical (`TestRuleTemplateMirrorDrift`), AC-MMC-011 template seed `merge_method: squash` ×3, AC-MMC-012 full suite green, AC-MMC-013 skill.md config pointer.
- Edge cases: EC-1 empty→squash no error, EC-3 case-sensitivity (`Squash` rejected).

run-phase quality: coverage internal/config 78.1% (no regression), `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0, `golangci-lint run` 0 issues. No SEC-HARDEN file touched (scope = internal/config + internal/template + spec).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-13
run_commit_sha: "639fd7a5e"   # cherry-picked from L1 worktree 0c017c411 onto SEC-HARDEN base 443fad912
run_status: green
ac_pass_count: 13
ac_fail_count: 0
coverage_internal_config: "78.1%"
cross_platform_build:
  host: pass
  windows: pass
new_warnings_or_lints_introduced: 0   # validation.go:394 unusedparams is PRE-EXISTING (line drift from 332)
plan_audit_verdict: "PASS-WITH-DEBT 0.82 (Tier M threshold 0.80), defects D2-D7 patched in ca4056509"
total_run_phase_files: 14
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-13
sync_commit_sha: "(backfilled below)"
sync_status: green
status_transition: in-progress → implemented
changelog_entry: added              # CHANGELOG.md [Unreleased] § Changed
spec_lint_fix: "§E.1 Out of Scope h3 added — MissingExclusions resolved (orch-direct, B6 idiom)"
```

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_complete_at: 2026-06-13
mx_commit_sha: "(backfilled below)"
status_transition: implemented → completed
four_phase_close: true             # plan + run + sync + Mx
github_issue: 1061                  # configurable merge method
frozen_amendment: "spec-workflow.md lifecycle table widened (squash → configured merge_method default squash); GATE-2 approved"
```
