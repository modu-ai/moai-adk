# SPEC-DIVECC-INVENTORY-VIEW-001 — Progress

> Run-phase progress + evidence for the unified `moai inventory [--json]` command
> (Epic Dive-into-CC N6, final candidate). cycle_type=tdd (RED-GREEN-REFACTOR).

---

## §E Mode Selection (Phase 0.95)

Per `.claude/rules/moai/workflow/orchestration-mode-selection.md` §D logging contract.

- **tier**: M (3-artifact set)
- **scope (file count)**: 3 files (1 new source + 1 new test + 1-line root.go edit) + SPEC dir
- **domain count**: 1 (Go CLI — `internal/cli`)
- **file language mix**: 100% Go
- **concurrency benefit**: LOW (coding-heavy, single-domain, single new file)
- **Agent Teams prereqs status**: not evaluated (single-domain, below the ≥3-domain threshold)

**Decision: sub-agent** (Mode 5).

Rationale: this is a coding-heavy, single-domain, single-new-file SPEC. Per the
Anthropic coding-task parallelism caveat, sequential sub-agent (Mode 5) is the
correct default — there is no multi-domain fan-out benefit (Mode 4) and the scope
is far below the Mode 6 mechanical-fan-out threshold (≥ ~30 uniform-transform
files). The orchestrator delegated the run-phase to a single `manager-develop`
sub-agent (cycle_type=tdd).

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-06-23
plan_status: audit-ready
plan_auditor_verdict: PASS
plan_auditor_score: 0.86
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
```

---

## §E.2 Run-phase Evidence

Run-phase TDD implementation of `moai inventory [--json]`. All 8 AC-INV criteria
(10 AC rows including a/b/c sub-criteria) verified with the acceptance.md
verification commands. Every reported result is the verbatim output of a command
actually run against this tree (`.claude/rules/moai/core/verification-claim-integrity.md`
§1.1 surface 2).

### AC PASS/FAIL Matrix

| AC | Status | Verification command | Actual output (observed) |
|----|--------|----------------------|--------------------------|
| AC-INV-001 | PASS | `grep -n newInventoryCmd internal/cli/root.go` + `go test -run TestNewInventory_CommandShape\|TestInventory_RegisteredOnRoot ./internal/cli/` | `root.go:113: rootCmd.AddCommand(newInventoryCmd())`; tests PASS (ok internal/cli) |
| AC-INV-002 | PASS | `grep -n 'session.QueryActiveWork\|provider.List\|harness.ListHarnesses' internal/cli/inventory.go` + `go test -run TestInventory_ComposesThreeSurfaces ./internal/cli/` | `QueryActiveWork("")` @237, `provider.List()` @283, `harness.ListHarnesses(projectRoot)` @302; test PASS |
| AC-INV-003 | PASS | `go test -run TestInventory_JSONShape\|TestInventory_JSONOutputUnmarshals ./internal/cli/` | 3 top-level keys (sessions/worktrees/harnesses) + per-row key fields round-trip; tests PASS |
| AC-INV-004 | PASS | `go test -run TestInventory_HumanReadableSummary\|TestInventory_RenderTextPopulated ./internal/cli/` | rendered output contains 3 surface labels + counts + key-field rows; tests PASS |
| AC-INV-005a | PASS | `go test -run TestInventory_EmptySurfacesGraceful ./internal/cli/` | absent registry + absent harness dir → both count 0, exit 0, no error; PASS |
| AC-INV-005b | PASS | `go test -run TestInventory_PerSurfaceErrorDegrades ./internal/cli/` | nil WorktreeProvider → worktrees.error set, sessions+harnesses still render; PASS |
| AC-INV-005c | PASS | `go test -run TestInventory_WorktreeListErrorDegrades ./internal/cli/` | non-nil stub List() returns "fatal: not a git repository" → worktrees.error set, others count 0, exit 0; PASS |
| AC-INV-006a | PASS | `grep -n resolveProjectRoot internal/cli/inventory.go` + `go test -run TestInventory_ProjectRootResolution ./internal/cli/` | `resolveProjectRoot(cmd)` @42; --project-root resolves harness dir from the flagged root (not cwd); PASS |
| AC-INV-006b | PASS | `git status --porcelain \| grep internal/(session\|cli/worktree\|cli/harness\|core/git)/` + `grep '"json"' internal/cli/worktree/list.go` | no backing-package file changed (only root.go + inventory.go + inventory_test.go + SPEC dir); no --json added to worktree list |
| AC-INV-007 | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/inventory.go` + `go test -run TestNewInventory_NoAskUserQuestion ./internal/cli/` | 0 matches in inventory.go; static guard test PASS |

**Result: 10/10 AC rows PASS.** No FAIL, no PASS-WITH-DEBT.

### Edge-case evidence (acceptance.md §B)

| # | Edge case | Result |
|---|-----------|--------|
| E1 | No `.moai/state/active-sessions.json` | sessions.count == 0, exit 0 (TestInventory_EmptySurfacesGraceful) |
| E2 | No `.claude/commands/harness/` | harnesses.count == 0, exit 0 (TestInventory_EmptySurfacesGraceful) |
| E3 | Only main checkout | worktrees lists provider baseline (live smoke: 76 worktrees incl. main) |
| E4 | `WorktreeProvider == nil` | worktrees.error set, others render (TestInventory_PerSurfaceErrorDegrades) |
| E5 | Harness `manifest_missing == true` | row rendered with manifest_missing:true (TestInventory_CollectHarnessesRows + live smoke) |
| E6 | Out of git repo | sessions 0 + harnesses 0 + worktrees.error set, exit 0 (TestInventory_OutOfProjectExitsZero) |
| E7 | --json vs default consistent counts | identical surface counts (TestInventory_ConsistentCountsAcrossModes) |

### Cross-platform build (E2)

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### Coverage (E3)

```
$ go test -coverprofile ./internal/cli/  → inventory.go statement coverage 91.2% (83/91 statements)
```
Target ≥ 85% — met. Uncovered statements are genuine-error branches (JSON marshal
failure on a valid struct; deps-present EnsureGit-success path; session/harness
backing read-error).

### Lint (E5)

```
$ golangci-lint run --timeout=2m ./internal/cli/...  → 0 issues
$ gofmt -l <new files>                                → (empty — formatted)
$ go vet ./internal/cli/                              → exit 0
```
No NEW findings. Baseline was 0 issues before the change.

### Live smoke (CLI end-to-end)

```
$ go run ./cmd/moai inventory --json
  → sessions.count 0, worktrees.count 76 (HEAD short-form e.g. c463257b),
    harnesses.count 3 (incl. manifest_missing harnesses), exit 0
$ go run ./cmd/moai inventory
  → 3 rounded-border cards (Sessions (0) / Worktrees (76) / Harnesses (3))
```

### Implementation summary

- **CREATE** `internal/cli/inventory.go` (~330 LOC): `UnifiedInventoryReport` +
  3 sub-report structs + 3 surface collectors (graceful degradation) +
  `newInventoryCmd()` factory + JSON/text rendering + `ensureWorktreeProvider`
  lazy wiring.
- **CREATE** `internal/cli/inventory_test.go`: 19 table/unit tests covering
  AC-INV-001..010, edge cases E1-E7, the subagent-boundary static guard, and
  the collector/render/wiring paths.
- **EXTEND** `internal/cli/root.go`: 1-line `rootCmd.AddCommand(newInventoryCmd())`
  in `init()`.
- **PRESERVE invariant (REQ-INV-011/012)**: ZERO modification to
  `internal/session/`, `internal/cli/worktree/`, `internal/cli/harness/`,
  `internal/core/git/`. The unified view CALLS their exported functions only.

### Cascading-failure note (CLAUDE.local.md §6)

`go test ./...` showed two `internal/hook` failures (`TestHookWrapper_ValidJSON`,
`TestHookWrapper_MoaiBinaryFallback`) under parallel full-suite load — each a
5-second timeout flake on the moai-binary-exec hook tests. Both PASS
deterministically in isolation (0.17s / 0.33s); `internal/hook` is NOT in this
SPEC's changeset. This is a pre-existing environmental flake unrelated to
SPEC-DIVECC-INVENTORY-VIEW-001 (verification-claim-integrity: observed, not
assumed — re-run in isolation confirmed PASS).

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-23
run_commit_sha: be6bbc238cf97253d4ea756c3bd8de174daeb903
run_status: implemented
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 4   # internal/session, internal/cli/worktree, internal/cli/harness, internal/core/git — all unchanged
l44_pre_commit_fetch: "2 1 → rebased onto origin/main c463257b (disjoint: parallel SPEC touched only STEERING-ALIGN-GUARDRAIL-HOOK-001 + CHANGELOG.md, zero overlap) → 0 1 clean"
l44_post_push_fetch: "0 0 (pushed to main, synced)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: exit 0
  windows_amd64: exit 0
total_run_phase_files: 4   # inventory.go (new) + inventory_test.go (new) + root.go (1-line) + spec.md (frontmatter)
m1_to_mN_commit_strategy: "single run-phase commit (M1-M5 squashed — Tier M single-new-file scope)"
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
# (sync-phase — orchestrator-direct 3-phase close)
sync_commit_sha: <PENDING-SYNC>
sync_status: complete
```
