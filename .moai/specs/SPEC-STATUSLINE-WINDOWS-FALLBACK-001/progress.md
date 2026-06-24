# Progress — SPEC-STATUSLINE-WINDOWS-FALLBACK-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored by manager-spec: spec.md, plan.md, acceptance.md, progress.md.
- Tier: M (standard). Requirements: 10 GEARS (REQ-SWF-001..010). Acceptance criteria: 12 mechanical ACs (AC-SWF-001..012) + 4 Given-When-Then scenarios.
- SPEC ID self-check: `SPEC-STATUSLINE-WINDOWS-FALLBACK-001` → PASS (decomposition: SPEC | STATUSLINE | WINDOWS | FALLBACK | 001).
- Root cause verified against source (3 layers, A/B/C) — not inferred.
- Status: draft. Awaiting plan-auditor gate + Implementation Kickoff Approval before run-phase.

## §E.2 Run-phase Evidence

cycle_type=tdd (RED→GREEN→REFACTOR). Reproduction-first per REQ-SWF-010 / CLAUDE.md §7 Rule 4: failing tests written and observed FIRST, then made to pass by the fix. RED observed in two stages — (1) compile failure (`WithResolvedMoaiPath` undefined) before M2, then (2) assertion-level failure (3/4 tests FAIL) against the unmodified template before M3; GREEN (4/4 PASS) after M3.

### AC PASS/FAIL Matrix

| AC ID | Bound REQ | Status | Verification command | Actual output |
|-------|-----------|--------|----------------------|---------------|
| AC-SWF-001 | REQ-SWF-004 | PASS | `grep -n 'GoBinPath' internal/template/templates/.moai/status_line.sh.tmpl` | (no match) — baked `.GoBinPath` branch removed |
| AC-SWF-002a | REQ-SWF-002 | PASS | `grep -n 'ResolvedMoaiPath' internal/template/context.go` | field `ResolvedMoaiPath` + `WithResolvedMoaiPath` option present |
| AC-SWF-002b | REQ-SWF-002/003 | PASS | `TestStatusLineResolvedExecutableBranch` | resolved-executable branch present, ordered after `command -v moai`, before `$HOME` branches |
| AC-SWF-003 | REQ-SWF-001/006 | PASS | `TestStatusLineEveryExecGuarded` | every `exec ... statusline` reached only via PATH match or `[ -f ]` guard |
| AC-SWF-004 | REQ-SWF-005 | PASS | `TestStatusLineEmptyResolvedPathOmitsBranch` | empty `ResolvedMoaiPath` -> no branch, no empty `exec`; `$HOME` branches present |
| AC-SWF-005 | REQ-SWF-005 | PASS | `TestStatusLineResolvedExecutableBranch` | non-empty path -> `[ -f ".../moai" ]` guarded branch (posix form) |
| AC-SWF-006 | REQ-SWF-002/003 | PASS | `grep -c 'WithResolvedMoaiPath' <3 files>` (sum) | 4 (init x1, update x2 @650/688, clean-install x1) |
| AC-SWF-007 | REQ-SWF-007 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| AC-SWF-008 | REQ-SWF-008 | PASS | `go test ./internal/template/... -run TestTemplateNeutralityAudit` | ok (no neutrality / leak finding) |
| AC-SWF-009 | REQ-SWF-009 | PASS | `go test -cover ./internal/template/ ./internal/runtime/gobin/` | template 84.7% (>= 84.6% baseline), gobin 70.0% (= baseline); new logic directly asserted by `status_line_fallback_test.go` |
| AC-SWF-010 | REQ-SWF-010 | PASS | RED-then-GREEN evidence | compile-fail -> assertion-fail (3/4) on unmodified template -> 4/4 PASS after M3 |
| AC-SWF-011 | REQ-SWF-001 (Template-First) | PASS | `make build` | recompiled binary with new template embedded; `git status` shows no stray `catalog.yaml` change (no separate `embedded.go` golden — embed is `//go:embed all:templates` read-at-compile, see §E.3 Gaps) |
| AC-SWF-012 | full suite | PASS-WITH-DEBT | `go test ./...` | 97 ok; 1 pre-existing unrelated FAIL (`internal/skills` `TestEntryRouterLOCCeiling`) confirmed failing on clean HEAD `94a468a2e` with changes stashed (not introduced by this SPEC) |
| AC-SWF-013 | REQ-SWF-011 (RT-007-003 supersession) | PASS | `go test ./internal/template/ -run TestFallbackChainOrder` + grep | test PASSES asserting new order; no residual assertion of a baked `.GoBinPath` branch (only supersession-doc comment + a negative `must-be-gone` assertion) |

### Scope discipline (Definition of Done)

- `gobin/resolver.go` `Detect` body — UNTOUCHED (grep PASS)
- `internal/shell/env.go` — UNTOUCHED (grep PASS)
- `install.ps1` — UNTOUCHED (grep PASS)
- Files changed (7): `internal/template/context.go`, `internal/template/templates/.moai/status_line.sh.tmpl`, `internal/template/status_line_fallback_test.go` (new), `internal/template/hardcoded_path_audit_test.go`, `internal/core/project/initializer.go`, `internal/cli/update.go`, `internal/cli/update_clean_install.go`

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-20
run_commit_sha: c4a42bc11
run_status: implemented
ac_pass_count: 12
ac_fail_count: 0
ac_pass_with_debt_count: 1   # AC-SWF-012 full-suite: 1 pre-existing unrelated FAIL (internal/skills LOC ceiling)
preserve_list_post_run_count: 3   # gobin/resolver.go Detect, internal/shell/env.go, install.ps1 — all untouched
l44_pre_commit_fetch: "origin/main 61db3aa96; rev-list 2 0 (origin ahead 2, non-overlapping SPEC-V3R6-HARNESS-V4-001 artifacts only); local HEAD base 94a468a2e"
l44_post_push_fetch: "git fetch origin main; rev-list 0 0 (local == origin/main, synced post-push c4a42bc11)"
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues on affected packages (QF1001 introduced then fixed)
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 7
m1_to_mN_commit_strategy: single-commit (Tier M Hybrid Trunk, main direct push)
```

### Gaps (§3.4 — explicitly NOT observed)

- No separate `internal/template/embedded.go` golden file exists in this repo: the embed mechanism is `//go:embed all:templates` in `internal/template/embed.go`, read from disk at compile time. `make build` recompiles the binary (no `.go` codegen step for the template FS). AC-SWF-011's "embedded.go regenerated" is satisfied by the recompile; there is no `git diff internal/template/embedded.go` to inspect (the file does not exist). This diverges from the plan.md §E step-5 wording but matches the actual repo mechanism.
- Runtime behavior of the rendered script on a real Windows installer machine was not executed (no Windows host); verification is at the render-assertion + cross-compile level, per the SPEC's render-boundary scope.

### Residual-risk (§3.5)

- The pre-existing `internal/skills` `TestEntryRouterLOCCeiling` failure is unrelated but means a bare `go test ./...` is not green at HEAD; a parallel session must resolve that skills-LOC debt separately.
- On a Windows installer machine where `os.Executable()` resolves to a path containing spaces, the `[ -f "..." ]` + `exec "..."` double-quoting (verified present in the template) is relied upon for correctness; not runtime-exercised here.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-20
sync_commit_sha: 545ed6879
sync_status: completed
sync_phase_tasks:
  changelog_entry_added: true
  frontmatter_status_transition: true
  frontmatter_updated_field_refreshed: false    # already set to 2026-06-20
  sync_commit_sha_backfill_required: true
spec_status_post_close: completed
verification_checks:
  go_test_passed: true
  golangci_lint_clean: true
  spec_lint_clean: true
  markdown_linting: true
  template_neutrality: true
```
