---
id: SPEC-V3R6-HOOK-ASYNC-EXPAND-001
title: "Hook async 확대 — Implementation Plan (Tier M, M1-M5)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook, internal/hook/testutil"
lifecycle: spec-anchored
tags: "hook, async, goroutine, performance, plan, sprint-2"
tier: M
---

# Implementation Plan — SPEC-V3R6-HOOK-ASYNC-EXPAND-001

## Section A — Context

### A.1 Design Doc Quote (verbatim, § 4 Layer 4 둘째 항목)

> **Async 확대**:
> - 현재 `PostToolUse`만 async
> - v3.0: `FileChanged`, `ConfigChange`, `TaskCreated`, `Notification`도 async 전환
> - 효과: blocking 시간 감소 (사용자 인지 응답 속도 ↑)

Source: `.moai/research/v3.0-design-2026-05-22.md` lines 249-252.

### A.2 SPEC artifacts

- `.moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/spec.md` (~205 lines, 6 REQs)
- `.moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/plan.md` (this file)
- `.moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/acceptance.md` (8 binary ACs)

### A.3 In-scope files (8-10 files)

**Production code (4)**:
- `internal/hook/file_changed.go` — REQ-HAE-001 refactor
- `internal/hook/config_change.go` — REQ-HAE-002 refactor
- `internal/hook/task_created.go` — REQ-HAE-003 refactor (conditional)
- `internal/hook/notification.go` — REQ-HAE-004 refactor (conditional)

**Test infrastructure (1 new)**:
- `internal/hook/testutil/wait_async.go` — REQ-HAE-006 `WaitForAsync` helper (NEW package directory)

**Test files (4)**:
- `internal/hook/file_changed_test.go` — extend with `WaitForAsync` + benchmark
- `internal/hook/config_change_test.go` — extend with `WaitForAsync` + benchmark
- `internal/hook/task_created_test.go` — extend with `WaitForAsync` + benchmark + observability matrix
- `internal/hook/notification_test.go` — extend with `WaitForAsync` + benchmark + observability matrix

**Goleak verification (1)**:
- `internal/hook/main_test.go` or `internal/hook/leak_test.go` — `goleak.VerifyTestMain` integration

### A.4 Dependencies

- `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001` MUST merge first (gates REQ-HAE-003/004 observability flag consumption)
- `go.uber.org/goleak` module — confirm presence in `go.mod` (already used elsewhere in MoAI-ADK testing; verify in pre-flight)

### A.5 PRESERVE list (DO NOT MODIFY)

- All other hook handlers (`stop.go`, `subagent_stop.go`, `user_prompt_submit.go`, `pre_tool_use.go`, `session_start.go`, `post_tool_use.go`, `worktree_create.go`, `worktree_remove.go`, `pre_compact.go`, `post_compact.go`, `cwd_changed.go`, `elicitation.go`, `task_completed.go`, `teammate_idle.go`, `instructions_loaded.go`, `stop_failure.go`)
- `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` (hook registration — no schema change needed; async is purely a Go-side concern)
- `.claude/hooks/moai/handle-*.sh` shell wrappers (stdin/stdout contract preserved; goroutine is invisible to wrapper)
- `internal/hook/contract.go`, `internal/hook/dual_parse.go`, `internal/hook/errors.go`, `internal/hook/generic_handler.go` (shared infrastructure)
- `internal/hook/.moai/` working-tree leak from `subagent_stop.go` (owned by SPEC-V3R6-HOOK-CONTRACT-FIX-001 REQ-HCF-005)

## Section B — Known Issues Auto-Injection

### B1. Cross-platform Build Tags

The four target handlers use no `syscall` package directly. `context.WithTimeout` + `go func()` + `sync.WaitGroup` are all portable. Verification: `GOOS=windows GOARCH=amd64 go build ./internal/hook/...` MUST pass.

The new `internal/hook/testutil/wait_async.go` MUST NOT use `syscall`. If a platform-specific code path emerges (unlikely), use `//go:build !windows` + `//go:build windows` file split per lessons #21.

### B2. Cross-SPEC Policy Conflict Pre-scan

```bash
grep -rn "Retired\|TestHarnessRetirement\|deprecation-marker\|superseded" internal/hook/
```

Expected: zero matches in the four target files. `SPEC-V3R6-HOOK-CONTRACT-FIX-001` (PR #1044 already MERGED) touches `worktree_create.go`/`worktree_remove.go` and `subagent_stop.go` — none of which overlap with this SPEC.

### B3. C-HRA-008 / Subagent Boundary Discipline

`internal/hook/` is subagent-domain code. AskUserQuestion / mcp__askuser invocations are PROHIBITED.

Verification:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
```
Expected: zero matches.

### B4. Frontmatter Canonical Schema

All three SPEC artifacts use the 12-field schema (`created:`, `updated:`, `tags:` — NOT `created_at:`, `updated_at:`, `labels:`) per `.claude/rules/moai/development/spec-frontmatter-schema.md`. `tier: M` field included.

### B5. CI 3-tier Awareness

- `spec-lint` MUST PASS on all three artifacts (frontmatter + heading structure + Out of Scope h3 sub-sections).
- `golangci-lint run --timeout=2m` MUST report 0 NEW issues over baseline.
- `Test (ubuntu-latest / macos-latest / windows-latest)` MUST exit 0 with `go test -race` clean.

### B6. spec-lint Heading Convention

`## Exclusions (What NOT to Build)` (h2 in spec.md § 6) followed by three `### 6.1 Out of Scope: ...` h3 sub-sections — satisfies `OutOfScopeRule` h3 heading text match (`internal/spec/lint.go:704`).

### B7. observer.go / capture path resolution

Not applicable to this SPEC. The four target handlers do not call `os.Getwd()` for `obsPath` resolution. Working-tree leak risk owned by `SPEC-V3R6-HOOK-CONTRACT-FIX-001` REQ-HCF-005.

### B8. Working Tree Hygiene

Runtime-managed files MUST NOT be modified by this SPEC:
- `.moai/harness/usage-log.jsonl`
- `.moai/state/*`
- `internal/hook/.moai/` (leak path; out-of-scope)

`.moai/state/hook-async-warnings.jsonl` is a NEW write path introduced by REQ-HAE-005. It is created lazily by the goroutine deadline-exceeded path; tests MUST use `t.TempDir()` for isolation.

## Section C — Pre-flight Check List

manager-develop MUST execute before any code change:

```bash
# 1. Branch + HEAD baseline
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build pre-verification
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Existing lint baseline (NEW vs pre-existing分)
golangci-lint run --timeout=2m 2>&1 | tail -10

# 4. Dependency confirmation (goleak)
go list -m go.uber.org/goleak || echo "MISSING: add to go.mod"

# 5. PRESERVE 대상 enumeration
ls internal/hook/*.go | grep -v -E "(file_changed|config_change|task_created|notification|_test)" | head -30

# 6. Cross-SPEC conflict scan
grep -rn "Retired\|TestHarnessRetirement\|superseded" internal/hook/ || echo "no conflicts"

# 7. Observability dependency check (HOOK-OBSERVE-OPT-IN-001 merged?)
gh pr list --state merged --search "SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001" --limit 1
ls .moai/config/sections/observability.yaml 2>&1 || echo "observability.yaml not yet present"
```

## Section D — Constraints (DO NOT VIOLATE)

### D.1 Explicit prohibitions

- DO NOT modify `.claude/settings.json` or its template (`internal/template/templates/.claude/settings.json.tmpl`). Async is a Go-side concern only.
- DO NOT modify shell wrappers under `.claude/hooks/moai/`. Their stdin/stdout contract is preserved.
- DO NOT push to `origin/main` without an open PR (Tier M run-phase follows feat-branch + squash merge per `git-strategy.yaml`).
- DO NOT use `--no-verify`, `--amend`, or force-push to main.
- DO NOT execute `git pull --rebase origin main` autonomously during run-phase (per lessons #1 from V3R6-CODE-COMMENTS-EN-001 Wave 2 — manager-develop autonomous rebase rewrites sibling SPEC SHAs).
- DO NOT run `AskUserQuestion` from within the manager-develop subagent (C-HRA-008 boundary).

### D.2 Required behaviors

- USE `context.WithTimeout(ctx, 5*time.Second)` for every spawned goroutine.
- USE `sync.WaitGroup`-or-equivalent (registered by production code, awaited by `WaitForAsync` test helper) — NOT `time.Sleep` in tests.
- USE Conventional Commits with `🗿 MoAI` trailer.
- USE `go.uber.org/goleak.VerifyTestMain(m, ...)` in `internal/hook/main_test.go` or equivalent for AC-HAE-007.
- USE `t.TempDir()` for any test that exercises `.moai/state/hook-async-warnings.jsonl` write paths.

### D.3 Conditional behavior (REQ-HAE-003/004)

- WHERE `observability.enabled == false` → handler returns empty payload, NO goroutine spawned (zero-overhead path).
- WHERE `observability.enabled == true` → handler launches goroutine, returns within ≤ 100 ms.

The config read MUST be cheap (cached or lazy-loaded once per process). Re-reading `observability.yaml` on every hook invocation defeats the latency goal.

## Section E — Self-Verification Deliverables

manager-develop's completion report MUST include:

### E.1 AC Binary PASS/FAIL Matrix

All 8 ACs (AC-HAE-001..008) per `acceptance.md`, each row populated with verification command + actual output.

### E.2 Cross-Platform Build

```
$ go build ./...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
$ GOOS=linux GOARCH=amd64 go build ./...    → exit 0
```

### E.3 Coverage (≥ 85% threshold for `internal/hook/`)

```
$ go test -cover ./internal/hook/...
```

### E.4 Subagent Boundary Grep (C-HRA-008)

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
(no output expected)
```

### E.5 Lint Status (NEW vs baseline)

```
$ golangci-lint run --timeout=2m
# NEW issues explicitly enumerated; pre-existing baseline separately marked
```

### E.6 Race Detection

```
$ go test -race ./internal/hook/...
PASS — exit 0 across darwin, linux, windows runners
```

### E.7 Goroutine Leak Verification

```
$ go test -count=1 -timeout=30s ./internal/hook/...
# goleak.VerifyTestMain reports 0 leaks
```

### E.8 Branch HEAD + PR State

- New commits SHA list (one per Milestone)
- `gh pr view <PR-number> --json state,mergeable,statusCheckRollup`
- PR title format: `feat(SPEC-V3R6-HOOK-ASYNC-EXPAND-001): Tier M run-phase — hook async expansion across 4 events`

### E.9 Blocker Report (if applicable)

If user decisions are required (e.g., HOOK-OBSERVE-OPT-IN-001 not yet merged), return structured blocker report — DO NOT call AskUserQuestion.

## Section F — Milestones (M1-M5)

This is a Tier M SPEC. Five milestones, executed in priority order. No time estimates per `agent-common-protocol.md` § Time Estimation.

### M1 — Foundation: `WaitForAsync` test helper (REQ-HAE-006, AC-HAE-008)

**Priority**: P1 (foundational — all subsequent test work depends on it)

**Deliverable**: `internal/hook/testutil/wait_async.go` (new package directory).

**Approach**:
- Package `testutil` with one exported function: `WaitForAsync(t *testing.T, wg *sync.WaitGroup, deadline time.Duration)`.
- Internally uses `sync.WaitGroup.Wait()` wrapped in a `select { case <-done: case <-time.After(deadline) }` pattern.
- On deadline timeout: `t.Fatalf("WaitForAsync: %v exceeded", deadline)`.
- Production handlers expose their `*sync.WaitGroup` via package-private accessor for tests; production code does NOT need to import `testutil`.

**Verification**: `goimports` clean, `golangci-lint` clean, no production-code imports of `testing` package.

**Self-AC**: package directory exists; one exported function; godoc complete.

### M2 — FileChanged async transition (REQ-HAE-001, AC-HAE-001/002)

**Priority**: P2 (unconditional, no observability gating)

**Deliverable**:
- `internal/hook/file_changed.go` — wrap MX delta scan in `go func()` with `context.WithTimeout(ctx, 5*time.Second)`.
- `internal/hook/file_changed_test.go` — extend with:
  - `TestFileChanged_AsyncReturn_Under100ms` (benchmark using `testing.B` with 10 concurrent invocations, asserts p95 ≤ 100 ms via `b.ReportMetric`).
  - `TestFileChanged_SideEffectsCompleted` using `WaitForAsync`.

**Self-AC**: AC-HAE-001 grep match + AC-HAE-002 benchmark passes.

### M3 — ConfigChange async transition (REQ-HAE-002, AC-HAE-003)

**Priority**: P2 (unconditional, parallel to M2 in concept but sequenced to avoid same-file contention with M1)

**Deliverable**:
- `internal/hook/config_change.go` — wrap diff-aware reload + validation in `go func()` with deadline.
- `internal/hook/config_change_test.go` — benchmark + side-effect verification.

**Self-AC**: AC-HAE-001 grep match (cumulative ≥ 2) + AC-HAE-003 benchmark passes.

### M4 — TaskCreated + Notification conditional async (REQ-HAE-003/004, AC-HAE-004/005)

**Priority**: P3 (conditional on HOOK-OBSERVE-OPT-IN-001 merge — blocker if not yet merged)

**Deliverable**:
- `internal/hook/task_created.go` — conditional `go func()` based on `observability.enabled` read.
- `internal/hook/notification.go` — same pattern.
- Both test files — observability matrix `{enabled=false, enabled=true}` × benchmark.
- Lightweight config-cache helper (likely in `internal/hook/observability.go` new file, OR via existing config loader if HOOK-OBSERVE-OPT-IN-001 provides one — coordinate with that SPEC's deliverables).

**Self-AC**: AC-HAE-001 grep match (cumulative ≥ 4) + AC-HAE-004 + AC-HAE-005 benchmarks pass; observability=false produces zero-goroutine path verified.

### M5 — Goleak + race + deadline + chore (REQ-HAE-005, AC-HAE-006/007 + status update)

**Priority**: P4 (closing verification + housekeeping)

**Deliverable**:
- `internal/hook/main_test.go` (or `internal/hook/leak_test.go`) — `goleak.VerifyTestMain(m)` integration.
- `internal/hook/async_warning.go` (new file) — implements `LogAsyncWarning(event, elapsed)` that appends a JSON line to `.moai/state/hook-async-warnings.jsonl`.
- `internal/hook/async_warning_test.go` — verify warning emission when goroutine exceeds 5s deadline (use `context.WithDeadline` + injected slow side-effect for determinism).
- `.moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/spec.md` — status `draft → implemented`, version `0.1.0 → 0.2.0`.
- `.moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/progress.md` — M1~M5 evidence capture (commit SHAs, AC PASS evidence, baseline residual notes).

**Self-AC**: AC-HAE-006 race clean + AC-HAE-007 goleak clean; spec.md status reflects completion.

## Section G — Technical Approach Summary

### G.1 Goroutine pattern (canonical)

```go
// Pseudocode — actual implementation per file
func (h *handler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
    // ... fast-path validation ...

    // Spawn side-effect goroutine with deadline
    asyncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    h.wg.Add(1)
    go func() {
        defer cancel()
        defer h.wg.Done()
        start := time.Now()
        if err := h.doSideEffect(asyncCtx, input); err != nil {
            // log structured warning (NOT panic, NOT propagate)
        }
        if elapsed := time.Since(start); elapsed > 5*time.Second {
            LogAsyncWarning(h.EventType(), elapsed)
        }
    }()

    return &HookOutput{Continue: true, ExitCode: 0}, nil
}
```

Key invariants:
- `context.Background()` decoupled from request `ctx` — request ctx cancellation should NOT cancel the side-effect (it MUST run to completion or deadline).
- `defer cancel()` releases context resources.
- `h.wg` is a package-private `*sync.WaitGroup` accessor returned by a test helper; production code does not expose it via the public API.

### G.2 Observability gating (REQ-HAE-003/004)

```go
// internal/hook/observability.go (new, M4)
var (
    observabilityOnce sync.Once
    observabilityEnabled bool
)

func IsObservabilityEnabled() bool {
    observabilityOnce.Do(func() {
        observabilityEnabled = loadObservabilityFlag() // reads .moai/config/sections/observability.yaml
    })
    return observabilityEnabled
}
```

The flag is loaded once per process (sync.Once) — cheap on the hot path.

### G.3 Deadline warning log format

`.moai/state/hook-async-warnings.jsonl`, one JSON object per line:

```json
{"timestamp":"2026-05-23T14:23:18Z","event":"FileChanged","deadline_ms":5000,"elapsed_ms":5247,"message":"goroutine exceeded deadline"}
```

`moai doctor` (future enhancement, OUT OF SCOPE here) MAY surface last 7 days of warnings. This SPEC only defines the write path.

## Section H — Risks (detail)

See spec.md § 5 for the risk table. Detailed mitigation:

### R1 mitigation — Goroutine leak prevention

- Every `go func()` uses `context.WithTimeout(context.Background(), 5*time.Second)` with `defer cancel()`.
- `goleak.VerifyTestMain(m, ...)` in `internal/hook/main_test.go` runs ALL `internal/hook/` tests under leak detection. AC-HAE-007 enforces 0 leaks.
- M5 deliverable `async_warning_test.go` includes a stress test that spawns 100 goroutines and verifies all complete or self-cancel within `5s + 100ms` slack.

### R2 mitigation — Test race conditions

- REQ-HAE-006 mandates `WaitForAsync` for every assertion against async side-effects.
- AC-HAE-006 enforces `go test -race ./internal/hook/...` exits 0.
- M1 `WaitForAsync` is the foundation — manager-develop must implement M1 before any test extension in M2-M5.

### R3 mitigation — Cross-Sprint with HOOK-OBSERVE-OPT-IN-001

- `depends_on: [SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001]` declared in this SPEC's frontmatter.
- Section C Pre-flight Check 7 verifies HOOK-OBSERVE-OPT-IN-001 merge state.
- If not merged: M1-M3 (unconditional async for FileChanged + ConfigChange) PROCEED; M4 (conditional async for TaskCreated + Notification) BLOCKS with structured blocker report.
- REQ-HAE-003/004 use the EXACT config schema (`.moai/config/sections/observability.yaml`, key `observability.enabled`) — coordinate with HOOK-OBSERVE-OPT-IN-001's SPEC body for naming alignment.

### R4 mitigation — Claude Code SDK contract

- Contract test in M5 (or in each M2-M4 test file): verify `HookOutput.Continue: true, ExitCode: 0` is the unchanged response shape.
- Async transition is INVISIBLE to Claude Code — the shell wrapper still receives the same stdin/stdout JSON. Only the timing-from-invocation-to-response shifts.
- If Claude Code SDK introduces a `synchronous: true` hook flag in a future version, that is OUT OF SCOPE; this SPEC's premise is the CURRENT (2026-05) contract.

### R5 mitigation — Silent async failure

- REQ-HAE-005 mandates warning log to `.moai/state/hook-async-warnings.jsonl`.
- Side-effect errors are NOT propagated to the response (would defeat async); they ARE logged.
- Future `moai doctor` enhancement surfaces these warnings — OUT OF SCOPE here but PATH defined.

## Section I — Acceptance Verification (cross-ref)

See [acceptance.md](./acceptance.md) for the 8 binary AC matrix with verification commands and traceability.

## Out of Scope (also see spec.md § 6)

### Out of Scope: Single mega-dispatcher consolidation
Deferred to a future Wave 4+ SPEC. This plan does not touch shell wrappers or CLI subcommand registry.

### Out of Scope: External queue / message broker
In-process goroutine + `context.WithTimeout` only. No Redis/Kafka/disk-queue integration.

### Out of Scope: Async transition for other hook events
Stop/SubagentStop/UserPromptSubmit/PreToolUse/SessionStart/WorktreeCreate/WorktreeRemove retain their current profile (PostToolUse is already async). Only the 4 events in § A.3 are modified.
