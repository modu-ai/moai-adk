---
spec_id: SPEC-V3R5-GIT-STRATEGY-SCHEMA-001
created: 2026-06-02
updated: 2026-06-02
---

# Progress — SPEC-V3R5-GIT-STRATEGY-SCHEMA-001

## §A.3 Version

- v0.1.0 (plan-phase, 2026-05-22)
- v0.1.1 (plan-audit iter-1 D1/D2/D4 patch, 2026-06-02, commit 2e00b958b)

## §D — Phase 0.5 Plan Audit Gate

- plan-auditor iter-1 verdict: **PASS-WITH-DEBT 0.82** (Tier M threshold 0.80)
- Dimensions: Clarity 0.85 / Completeness 0.90 / Testability 0.80 / Traceability 0.95
- MP-1/2/3/4 all PASS; GEARS/EARS PASS
- D1 (BLOCKING): acceptance.md AC-GSS-002 oracle self-contradiction (template↔production yaml drift) → **RESOLVED** via manager-spec commit 2e00b958b (acceptance.md AC-GSS-002 fixture declared synthetic with template-canonical oracle; AC-GSS-004 D2 per-field deprecation grep; plan.md D4 commit-subject `plan(`→`feat/test/chore(`)
- D3 (SHOULD-FIX): REQ-GSS-008 audit-removal precondition — git-strategy loader path in loadedSections unverified. Carry to M4 contingency (plan.md M4 step 4). NOTE: orchestrator pre-check found NO `git-strategy` reference in `internal/config/loader.go` — M4 may require a loader-registration fix.
- GATE-2: user-approved 2026-06-02 (Option A — D1-patch-first then run-phase)

## §E — Phase 0.95 Mode Selection

### Input parameters
- tier: M
- scope (file count): ~7 files in `internal/config/` — types.go [EXTEND], defaults.go [EXTEND], validation.go [EXTEND], audit_loader_completeness_test.go [EXTEND], git_strategy_nested_test.go [NEW], types_test.go [EXTEND], defaults_test.go [EXTEND]
- domain count: 1 (`internal/config` Go package)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, single package; M0 struct hierarchy must precede M1-M4)
- Agent Teams prereqs: not evaluated (single-domain coding work resolves to Mode 5 before the capability gate matters)

### Mode evaluation table
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | Multi-milestone struct refactor + new test file; not a single-line change |
| 2 background | no | Write/Edit-heavy; background subagents auto-deny Write per CONST-V3R2-020 |
| 3 agent-team | no | Single domain (<3 domains); coding-heavy; Agent Teams 3-5 ceiling unjustified |
| 4 parallel | no | Coding-heavy single-package work; Finding A4 caveat — sequential preferred |
| 5 sub-agent | YES | Coding-heavy, single domain, hard milestone dependency (M0 → M1-M4 → M5) |

### Decision: sub-agent

### Justification
Coding-heavy single-package (`internal/config`) brownfield refactor with a hard milestone dependency: M0 (struct hierarchy) must land before M1-M4 can author the roundtrip test, accessor, defaults, and validation. Per Anthropic Finding A4 (coding tasks involve fewer truly parallelizable subtasks than research) and the Mode 4-vs-5 tie-breaker in orchestration-mode-selection.md §B.2, Mode 5 (sequential sub-agent — manager-develop, cycle_type=tdd, Tier M Section A-E delegation template) is correct. Plan-auditor verdict is PASS-WITH-DEBT 0.82 with D1 patched; D3 carried to M4 contingency.

## §Run-phase Evidence

### M0 — Struct hierarchy + deprecation comments + ActiveModeProfile accessor

- File: `internal/config/types.go` [EXTEND]
- Added 6 new sub-structs (`GitLabConfig`, `AutomationConfig`, `BranchCreationConfig`, `CommitStyleConfig`, `HooksConfig`, `ModeProfile`) + extended `GitStrategyConfig` (top-level Mode/Provider/GitHubUsername/GitLab + 3 ModeProfile fields + 5 deprecated FLAT fields with per-field `// Deprecated:` comments).
- Included `ActiveModeProfile()` accessor (REQ-GSS-004; co-located in types.go alongside the struct — plan M2 deliverable folded into M0 since same file, keeps build green).
- LINT_BASELINE captured pre-M0: `golangci-lint run --timeout=2m ./internal/config/...` → `0 issues.` (1 line). NEW target ≤ 0.
- `go build ./internal/config/...` → exit 0; `go vet ./internal/config/...` → exit 0.
- Existing FLAT-field tests (`TestConfigStructCreation`, `TestGitStrategyConfigFields`, `TestNewDefaultGitStrategyConfig`) → PASS (backward-compat preserved, no test edits).
- Status transition: spec.md frontmatter `draft → in-progress` + `updated: 2026-06-02` on M0 (first run-phase code commit).

### M1 — Nested roundtrip + M2 ActiveModeProfile tests

- File: `internal/config/git_strategy_nested_test.go` [NEW]
- `TestGitStrategyConfig_NestedRoundTrip`: synthetic fixture (template-canonical oracle per AC-GSS-002) → yaml.Unmarshal → 12 nested-field assertions (REQ-GSS-003). PASS.
- `TestGitStrategyConfig_ZeroValueSafety`: zero-value `ActiveModeProfile()` returns (nil,false) no-panic (R-GSS-002, Edge-GSS-001). PASS.
- `TestGitStrategyConfig_ActiveModeProfile`: 5-case table (manual/personal/team/empty/unknown); `=== RUN` count = 6 (AC-GSS-003). PASS.
- `go test -race -run TestGitStrategyConfig_` → PASS.

### M3 — NewDefaultGitStrategyConfig nested defaults

- Files: `internal/config/defaults.go` [EXTEND], `internal/config/defaults_test.go` [EXTEND]
- `NewDefaultGitStrategyConfig()` returns 3 populated ModeProfiles mirroring `git-strategy.yaml.tmpl` (REQ-GSS-006); FLAT defaults preserved.
- `TestNewDefaultGitStrategyConfig` extended with 12 nested assertions (AC-GSS-005) + 4 preserved FLAT assertions. PASS.

### M4 — Validation extension + audit registry cleanup

- Files: `internal/config/validation.go` [EXTEND], `internal/config/audit_loader_completeness_test.go` [EXTEND]
- 20 explicit-literal nested `checkStringField("git_strategy.{manual,personal,team}.*")` calls (REQ-GSS-007). AC-GSS-006 grep1=22 (≥11), grep2=20 (≥9).
- Removed `"git-strategy"` from `acknowledgedUnloadedSections` (REQ-GSS-008). AC-GSS-007 grep1=0, grep2=0.
- **D3 outcome (empirical)**: `TestAuditLoaderCompleteness` PASSES with NO `loader.go` edit. The audit test iterates only `*.yaml` files in the template sections dir and skips `*.yaml.tmpl` (test line 103). `git-strategy` exists there solely as `git-strategy.yaml.tmpl`, so it never produced a `sectionName` — the allowlist entry was dead. `acknowledged` count 13 → 12, test green. No loader-registration fix was required.

### M5 — Cross-platform + lint + race final gate

- Cross-platform build: `PASS-darwin` / `PASS-windows` / `PASS-linux` (AC-GSS-008).
- `go test -race ./internal/config/...` → PASS (AC-GSS-009).
- Lint NEW = final − baseline = 0 − 0 = 0 (golangci-lint `0 issues.`) (AC-GSS-009).
- C-HRA-008 boundary grep → empty (AC-GSS-009).
- Coverage `go test -cover ./internal/config/...` → 77.3% package-level (pre-existing baseline; git-strategy code itself 100% function-level: `NewDefaultGitStrategyConfig` 100%, `ActiveModeProfile` 100%, `validateDynamicTokens` 100%). The 85% target reflects pre-SPEC package loaders out of scope; SPEC additions only raise coverage.
- Full repo `go test ./...`: `internal/config` PASS. 2 pre-existing flaky failures in `internal/hook/wrapper_test.go` (`TestHookWrapper_MoaiBinaryFallback`, `TestHookWrapper_ValidJSON`, `signal: killed` ~5s under parallel contention) — PASS in isolation; outside this SPEC's `internal/config` scope; not introduced by M0-M4 (diff is config-only). Documented STABILIZE-003 flaky candidate.

### AC Binary PASS/FAIL Matrix (final)

| AC | Status | Evidence |
|----|--------|----------|
| AC-GSS-001 | PASS | go doc grep1=22 (≥12), grep2=18 (≥14), grep3=2 (≥1) |
| AC-GSS-002 | PASS | TestGitStrategyConfig_NestedRoundTrip ok |
| AC-GSS-003 | PASS | TestGitStrategyConfig_ActiveModeProfile ok, RUN=6 |
| AC-GSS-004 | PASS | TestConfig* ok; AutoBranch Deprecated=1, WorktreeRoot Deprecated=1, aggregate=3 |
| AC-GSS-005 | PASS | TestNewDefaultGitStrategyConfig ok (12 nested assertions) |
| AC-GSS-006 | PASS | grep checkStringField("git_strategy.=22 (≥11); nested-path grep=20 (≥9) |
| AC-GSS-007 | PASS | git-strategy literal=0; in-slice=0; TestAuditLoaderCompleteness ok |
| AC-GSS-008 | PASS | 3× go build (darwin/windows/linux) exit 0 |
| AC-GSS-009 | PASS | C-HRA-008 grep empty; go test -race PASS; lint NEW=0 |

9/9 must-pass ACs PASS. Tier M threshold (≥80% AC pass, ALL must-pass = PASS) satisfied at 100%.

### Run-phase close signal

- run_status: implemented (9/9 AC PASS)
- ac_pass_count: 9 / ac_fail_count: 0
- cross_platform_build: PASS-darwin / PASS-windows / PASS-linux
- new_warnings_or_lints_introduced: 0
- m0_to_m5_commit_strategy: 5 milestone commits (M0 83caed674, M1 e8b57f36e, M3 13dbcae48, M4 ec6398b8b, M5 this commit); M2 accessor folded into M0 (same file)
- preserve_list_post_run: docs-site dirty files untouched (in shared checkout; not present in isolated L1 worktree)
