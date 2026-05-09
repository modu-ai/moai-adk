# Wave 1 Task Decomposition

SPEC: SPEC-V3R3-CI-AUTONOMY-001
Wave: 1 (Quick Wins — T1 + T5)
Branch: `feat/SPEC-V3R3-CI-AUTONOMY-001-wave-1` (worktree: `/Users/goos/.moai/worktrees/moai-adk/ciaut-wave-1`, base: origin/main)
Methodology: TDD (RED-GREEN-REFACTOR)
Strategy reference: `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave1.md`

## Atomic Tasks (dependency-ordered)

| Task ID | Description | Requirement | Dependencies | Planned Files | Acceptance | Status |
|---------|-------------|-------------|--------------|---------------|------------|--------|
| W1-T07 | Required-checks SSoT YAML + Go loader | REQ-CIAUT-027 | none | `.github/required-checks.yml`, `internal/template/templates/.github/required-checks.yml`, `internal/config/required_checks.go`, `internal/config/required_checks_test.go` | AC-CIAUT-021 | pending |
| W1-T06 | github_init.go branch protection wiring + graceful auth-failure | REQ-CIAUT-025/026/028 | W1-T07 | `internal/cli/github_init.go` (modify), `internal/cli/github_init_test.go` (extend), `internal/cli/branch_protection.go`, `internal/cli/branch_protection_test.go`, `internal/template/templates/.github/branch-protection.json.tmpl` | AC-CIAUT-011, AC-CIAUT-022 | pending |
| W1-T04 | Makefile ci-local + pr-merge targets | REQ-CIAUT-003/004, REQ-CIAUT-029 | W1-T01 (cross-compile.sh ref) | `Makefile`, `internal/template/templates/Makefile` | (verified end-to-end via W1-T05) | pending |
| W1-T01 | scripts/ci-mirror/run.sh skeleton + _common.sh helpers | REQ-CIAUT-001/004/007 | none | `scripts/ci-mirror/run.sh`, `scripts/ci-mirror/lib/_common.sh`, `scripts/ci-mirror/cross-compile.sh`, `scripts/ci-mirror/run_test.sh`, mirrors under `internal/template/templates/scripts/ci-mirror/` | AC-CIAUT-002 | pending |
| W1-T02 | scripts/ci-mirror/lib/go.sh — Go pipeline (vet+lint+test+cross-compile) | REQ-CIAUT-003 | W1-T01 | `scripts/ci-mirror/lib/go.sh`, `scripts/ci-mirror/lib/go_test.sh`, mirror | AC-CIAUT-001 | pending |
| W1-T03 | 14 lightweight per-language modules + 4 spot-check tests | REQ-CIAUT-007 | W1-T01 | `scripts/ci-mirror/lib/{python,node,rust,java,kotlin,csharp,ruby,php,elixir,cpp,scala,r,flutter,swift}.sh` (14), `lib/{python,node,rust,swift}_test.sh` (4), mirrors | AC-CIAUT-002 | pending |
| W1-T05 | Pre-push hook (POSIX sh) + Go installer extension + invocation log | REQ-CIAUT-001/002/005/006(scoped)/030 | W1-T01..T04 | `internal/template/templates/.git_hooks/pre-push`, `internal/cli/hook_install.go` (extend), `internal/cli/hook_install_test.go` (extend), `scripts/ci-mirror/prepush_e2e_test.sh` | AC-CIAUT-001, AC-CIAUT-003, AC-CIAUT-013 | pending |

## Honest Scope Adjustments (user-approved)

1. **REQ-CIAUT-006** (`--no-verify` bypass logging): Hook cannot self-log when bypassed via `--no-verify` (hook isn't invoked). Wave 1 logs hook **invocations** (pass/fail) to `.moai/logs/prepush-bypass.log`. Real bypass detection (linking pushed commits to local CI absence) deferred to **Wave 2** CI watch loop.

2. **AskUserQuestion bridge for Go binary** (REQ-CIAUT-025): Go binary uses `Confirmer` interface (TTY default for direct invocation). MoAI orchestrator pre-confirms via AskUserQuestion then invokes `moai github init --yes-branch-protection` (new flag). This preserves the orchestrator-only HARD constraint without breaking direct CLI usage.

## Acceptance Criteria → Test File Map

| AC | Test Location |
|----|---------------|
| AC-CIAUT-001 | `scripts/ci-mirror/prepush_e2e_test.sh`, `scripts/ci-mirror/lib/go_test.sh` |
| AC-CIAUT-002 | `scripts/ci-mirror/run_test.sh` cases 1-3 + structural conformance loop |
| AC-CIAUT-003 | `scripts/ci-mirror/prepush_e2e_test.sh` (invocation log scoped) |
| AC-CIAUT-011 | `internal/cli/branch_protection_test.go::TestRenderBranchProtectionJSON` |
| AC-CIAUT-012 | `internal/cli/branch_protection_test.go::TestApplyBranchProtection` |
| AC-CIAUT-013 | `scripts/grep-no-release-automation.sh` (negative grep) |
| AC-CIAUT-021 | `internal/config/required_checks_test.go::TestNoHardcodedContexts` |
| AC-CIAUT-022 | `internal/cli/branch_protection_test.go::TestApplyBranchProtection_AuthFailure` |

## TRUST 5 Targets

- **Tested**: per-package coverage ≥85% (≥90% for `internal/config/required_checks.go`)
- **Readable**: shellcheck -s sh + golangci-lint clean delta from baseline
- **Unified**: SSoT pattern (W1-T07 → W1-T06)
- **Secured**: zero release/tag automation in Wave 1 (negative grep test)
- **Trackable**: Conventional Commits per task; CHANGELOG.md updated at end
