# progress.md — SPEC-CONFIG-MODE-MIGRATE-001

> Plan-phase skeleton. The §E.* headings below are the canonical markers the
> `era.go` classifier greps for (literal `§E.2` / `§E.3` / `§E.4` substrings).
> Do NOT author §E.5 (retired Mx-phase marker). Only §E.1 is populated at
> plan-phase; §E.2-§E.4 are placeholder headings left for the run/sync phases.

## §E.1 Plan-phase Audit-Ready Signal

- `plan_status`: _pending plan-auditor_
- `plan_complete_at`: _pending_
- `plan_artifact_count`: 2 (Tier S — spec.md + plan.md; this progress.md is emitted
  at every Tier and is not counted in the Tier total)
- `tier`: S
- `spec_lint_result`: _pending (run before plan-phase commit)_

## §E.2 Run-phase Evidence

Baseline HEAD (pre-run): `c028f3bd4` (branch `worktree-spec-config-mode-migrate-001`).
Implementation: `internal/cli/mode_migrate.go` + `internal/cli/mode_migrate_test.go`.
Core API: `IsWideningCandidate`, `ScanConfigDir`, `FormatDryRun`, `ApplyWidening`,
`runModeMigrate`; cobra wiring `moai config mode-migrate [--apply]` (new `config`
parent — no prior `moai config` existed; diagnostics live under `moai doctor config`).

### AC PASS/FAIL matrix (attribution: command + verbatim observed output + baseline)

Baseline for every row below: HEAD `c028f3bd4`, tree
`worktree-spec-config-mode-migrate-001`, command run `2026-08-13`.

| AC | Status | Verification command | Observed output (verbatim) |
|----|--------|---------------------|-----------------|
| AC-MIG-001 (dry-run no-op) | PASS | `go test -run TestModeMigrateDryRun_NoOp_OnDisk ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  0.759s` (modes on disk unchanged: 0600→0600, 0644→0644; output contains path + `0600` + `0644` + `--apply` + "No files were modified") |
| AC-MIG-002 (apply widens) | PASS | `go test -run TestModeMigrateApply_WidensNarrowed ./internal/cli/` | `ok ...` (0600 file → defs.FilePerm on disk, asserted via `os.Stat`) |
| AC-MIG-003 (only-widen) | PASS | `go test -run TestModeMigrateApply_OnlyWidens ./internal/cli/` | `ok ...` (0644 stays 0644; 0600 → 0644) |
| AC-MIG-004 (scope .moai/config/) | PASS | `go test -run TestModeMigrateApply_ScopeConfigOnly ./internal/cli/` | `ok ...` (`.claude/settings.json` at 0600 OUTSIDE `.moai/config` unchanged at 0600) |
| AC-MIG-005 (idempotent) | PASS | `go test -run TestModeMigrate_Idempotent ./internal/cli/` | `ok ...` (already-canonical tree → "0 candidate(s) found"; `--apply` no-op, exit 0) |
| AC-MIG-006 (helper routing) | PASS | `go test -run TestModeMigrate_HelperRouting ./internal/cli/` | `ok ...` (source contains `atomicfile.Write`; no bare `os.WriteFile(`; single `os.Chmod(` site references named `defs.FilePerm`; zero hardcoded mode literals) |
| AC-MIG-007 (0700 non-subset) | PASS | `go test -run TestModeMigrateApply_NonSubsetMode_Unchanged ./internal/cli/` | `ok ...` (0700 file unchanged post-apply; dry-run reports "1 candidate(s) found" — the 0600 file only; `IsWideningCandidate(0700)==false`) |
| AC-MIG-008 (symlink scope-leak) | PASS | `go test -run TestModeMigrate_SymlinkSkipped ./internal/cli/` | `ok ...` (dry-run reports "skipped (symlink)"; external symlink target stays 0600 post-apply — no `os.Chmod` landed on it) |

Predicate unit (spec.md §D.2 enumeration): `go test -run TestIsWideningCandidate_Enumeration ./internal/cli/` → `ok` (0600/0640 candidate; 0700/0660/0644/0664/0666/0500 not).

### E2 Cross-platform build
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E5 Lint (NEW vs baseline)
```
$ golangci-lint run --timeout=3m ./internal/cli/  → 0 issues
```
NEW issues introduced: 0 (the initial errcheck/staticcheck findings during GREEN
were resolved in the same iteration: `_, _ = fmt.Fprint*` discard pattern + the
empty-branch test refactor; final run is clean).

### E8 RED failing-test output (TDD falsifiability — captured BEFORE GREEN)
```
$ go test -run TestModeMigrateDryRun_NoOp_OnDisk ./internal/cli/
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/mode_migrate_test.go:67:12: undefined: runModeMigrate
internal/cli/mode_migrate_test.go:103:12: undefined: runModeMigrate
...
internal/cli/mode_migrate_test.go:249:6: undefined: IsWideningCandidate
internal/cli/mode_migrate_test.go:271:13: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```
Tests were authored test-first (M1 locks the API surface before M2 apply path).
GREEN then implemented `IsWideningCandidate` / `ScanConfigDir` / `FormatDryRun` /
`ApplyWidening` / `runModeMigrate` + cobra wiring in `mode_migrate.go`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-08-13"
run_commit_sha: "5817e9173"   # backfilled post-commit (D3 self-referential-hazard exemption)
run_status: "audit-ready"
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0   # internal/config/atomicfile/*.go untouched (sibling CLOSED SPEC owned); real .moai/config/sections/*.yaml untouched (t.TempDir only)
l44_pre_commit_fetch: "0 0"        # worktree branch vs origin/main: clean (no parallel-session race)
l44_post_push_fetch: "deferred to push (orchestrator-owned)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: "exit 0"
  windows_amd64: "exit 0"
total_run_phase_files: 2   # internal/cli/mode_migrate.go + internal/cli/mode_migrate_test.go
m1_to_mN_commit_strategy: "single M1 commit carries spec.md status draft→in-progress + plan-audit amendments + progress §E + both impl files (Tier S — M1/M2/M3 collapsed into one cohesive commit since the API surface, apply path, and wiring are one small unit)"
```

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase — `sync_commit_sha:` field is populated by the single sync commit
at sync-phase close (manager-docs)._

## §F Phase 4 Mode Selection

**Input parameters**:
- tier: S
- scope (file count): ~3-4 (internal/cli/config/mode_migrate.go + mode_migrate_test.go + cmd.go wiring)
- domain count: 1 (Go config CLI)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy; M1→M2→M3 sequential dependency)

**Mode evaluation**:
| Mode | Selected | Rationale |
|------|----------|-----------|
| trivial | no | multi-file semantic new code |
| background | no | coding-heavy, not read-only |
| agent-team | n/a | RETIRED (Mode 3 tombstone) |
| parallel | no | coding-heavy per Anthropic parallelism caveat |
| sub-agent | YES | Tier S coding-heavy, sequential milestones |
| workflow | no | not high-volume mechanical |

**Decision**: sub-agent (Mode 5)

**Justification**: Tier S coding-heavy work with sequential milestone dependency
(M1 locks the scan API surface → M2 apply path → M3 wiring). Anthropic's coding-task
parallelism caveat favors sequential sub-agent over parallel fan-out. manager-develop
with cycle_type=tdd, delegated per-milestone under an armed ac_converge goal.

**Progression mode**: Autonomous (ac_converge goal armed for AC-MIG-001~008 convergence —
AC-MIG-008 symlink-skip added in the plan-audit D5 fix batch).

**Implementation Kickoff Approval**: PASSED — user approved run-phase entry with
autonomous progression (this session, worktree-spec-config-mode-migrate-001).
