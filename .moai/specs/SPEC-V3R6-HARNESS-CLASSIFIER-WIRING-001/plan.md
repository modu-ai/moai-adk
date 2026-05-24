---
id: SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001
title: "V3R4 Harness Classifier Runtime Wiring — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Evolution Loop Closure"
module: ".claude/skills/moai/workflows/harness.md, internal/cli/hook.go (Option A) or internal/hook/post_tool.go (Option B)"
lifecycle: spec-anchored
tags: "harness, classifier, wiring, runtime, v3r6, tier-s-minimal, plan"
---

# Implementation Plan — SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001

## §1. Overview

Tier S minimal SPEC: single milestone M1, 1-2 files, ≤300 LOC envelope. Mechanical wiring insertion to invoke the existing Go classifier API from the workflow body `status` verb. No new public cobra subcommand (preserves `BC-V3R4-HARNESS-001-CLI-RETIREMENT`).

The bulk of the implementation work is in selecting and implementing the wiring mechanism (one of three options below) and confirming fail-safe behavior. The classifier code itself is already shipped from V3R4-HARNESS-003 and exposes `AggregatePatterns`, `ClassifyTier`, `WritePromotion` as the public Go API.

## §2. Milestones

### M1 — Wiring Insertion (Single Milestone, Tier S Minimal)

**Goal**: Workflow body §2.1 `status` verb invokes the Go classifier on each call. Output includes a tier distribution table derived from `tier-promotions.jsonl`. Fail-open on classifier error. No-op when `learning.enabled: false`.

**Sub-tasks** (within M1, sequential):
1. Implement chosen wiring mechanism (Option A / B / C per §3 decision)
2. Insert workflow body Bash call at §2.1 status verb
3. Render tier distribution table in status output (read from `tier-promotions.jsonl`)
4. Add error annotation rendering for classifier failures
5. Add `learning.enabled` config gate check
6. Run AC-HCW-001..006 verification batch

**Exit criteria**: All ACs (1-5 mandatory + AC-006 Optional) PASS. `go test ./internal/harness/... ./internal/hook/... ./internal/cli/...` zero regression. `tier-promotions.jsonl` non-empty after `/moai:harness status` invocation on this project (97 usage-log events → expect ≥1 promotion entry per unique pattern key).

**Priority**: High (P1) — closes V3R4 learning loop on this project; unblocks subsequent V3R6 harness SPECs (PROPOSAL-GEN-001, RATELIMIT-001, SNAPSHOT-001).

## §3. Wiring Branch Trade-off Matrix (verbatim from orchestrator directive)

Three options for HOW the workflow body §2.1 invokes the Go classifier:

| Option | Mechanism | V3R4 CLI retirement impact | Latency | Complexity |
|--------|-----------|----------------------------|---------|------------|
| **A** | Workflow body Bash call to `moai hook harness-classify` (new hook subcommand, reuses `hook` namespace — `hook` is NOT retired) | Compatible — `hook` namespace orthogonal to retired `harness` namespace | <100ms (Go startup) | Low — 1 hook subcommand + 1 workflow body Bash insertion |
| **B** | PostToolUse hook (`internal/hook/post_tool.go`) invokes classifier in-process after raw event append, every N events (e.g., 10) | None — hook namespace untouched | <5ms in-process per N invocations | Medium — batching counter state, latency budget verification |
| **C** | SessionStart hook batch-runs classifier once per session | None — hook namespace untouched | +50-200ms on session start | Low — but ties classifier to session boundary, not status verb |

### Recommendation: Option A

**Rationale**:
1. **Cleanest separation**: status verb is the conceptual trigger; Option A makes the wiring explicit at the call site rather than implicit in event batching (B) or session lifecycle (C).
2. **Deterministic trigger**: User invocation of `/moai:harness status` reliably produces fresh promotions, which matches user expectation of the status verb as a "show me current state" command.
3. **Hook namespace already exists as orthogonal surface**: `hook` namespace is the canonical home for internal subcommands invoked by workflow body Bash calls (precedent: `moai hook session-start`, `moai hook subagent-stop`, etc.). Adding `moai hook harness-classify` follows existing convention zero-friction.
4. **Latency acceptable**: <100ms Go startup is dominated by `status` verb's other I/O (reading usage-log.jsonl, rendering tables). Marginal cost is invisible to user.
5. **Compatible with REQ-HCW-005 (Optional MAY)**: Option A's single-batch-per-invocation cadence matches the MAY clause exactly. Future PostToolUse incremental SPEC can layer Option B atop Option A without conflict.

**Why not Option B**:
- Batching counter state introduces persistent mutable state in hook layer (currently hook layer is stateless event recorder).
- N-event threshold tuning is a separate concern (5? 10? 25?) — best deferred to a dedicated incremental SPEC (`SPEC-V3R6-HARNESS-INCREMENTAL-001` candidate per spec.md §E).
- In-process latency budget verification adds AC complexity not justified for V3R6 baseline.

**Why not Option C**:
- Couples classifier to session boundary rather than status verb. Conflicts with status verb's "show me current state" UX contract.
- SessionStart hook is already busy with GLM credential setup and teammateMode injection — adding classifier invocation grows the critical-path session boot latency.
- Session-bounded promotions miss events that occur within a long-running session (which is the common case for this project's multi-SPEC sprints).

### Lock-in Decision Boundary

This `plan.md §3` is the **recommendation**, not the lock-in. Lock-in occurs in run-phase manager-strategy:
- If plan-auditor concurs with Option A → run-phase implements Option A directly.
- If plan-auditor surfaces evidence favoring B or C (e.g., latency benchmark showing >500ms on `status` invocation) → run-phase manager-strategy makes the final binding decision via AskUserQuestion to orchestrator before implementation begins.

`spec.md §G` is the canonical pointer to this `plan.md §3` decision rule.

## §4. Technical Approach (Option A Concrete Design)

Conditional on Option A being chosen (recommended), the implementation has two artifacts:

### 4.1 New hook subcommand: `moai hook harness-classify`

**File**: `internal/cli/hook.go` (existing file, add new subcommand)
**LOC**: ~80-150 Go LOC
**Behavior**:
1. Read `.moai/harness/harness.yaml` → check `learning.enabled`. If `false`, exit 0 (no-op per REQ-HCW-004).
2. Call `harness.AggregatePatterns(.moai/harness/usage-log.jsonl)` → returns `[]PatternKey`.
3. Call `harness.ClassifyTier(patterns)` → returns `[]TierAssignment`.
4. Call `harness.WritePromotion(.moai/harness/learning-history/tier-promotions.jsonl, assignments)` → appends promotions.
5. Print one-line summary to stderr (e.g., `harness-classify: 4 patterns → 4 promotions written`).
6. Exit 0 on success. Exit 1 on classifier error (workflow body interprets exit 1 as the trigger for REQ-HCW-003 fail-open annotation; workflow body itself does NOT abort).

**Cobra registration**: Add as subcommand under existing `hookCmd` in `internal/cli/hook.go`. No new top-level command. Hidden from help if desired (optional polish; not required for ACs).

### 4.2 Workflow body §2.1 Bash insertion

**File**: `.claude/skills/moai/workflows/harness.md` (existing file, edit §2.1)
**LOC**: ~20 Bash LOC (within the existing §2.1 status verb section)
**Behavior**:
1. Before rendering the existing status sections, invoke: `moai hook harness-classify 2>&1` (capture stderr into a buffer).
2. If exit code 0: render the classifier summary line above the existing status sections.
3. If exit code 1: render an error annotation block (`> ⚠ classifier error: <stderr>`) above the existing status sections, then continue rendering the rest. Do NOT `exit 1` from the workflow body.
4. Render the new "Tier Distribution" table by reading `.moai/harness/learning-history/tier-promotions.jsonl` and grouping by tier (tier 1 / 2 / 3 / 4 counts).
5. Existing status sections (usage-log summary, etc.) render unchanged after the new sections.

## §5. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Option A choice locked-in here conflicts with future incremental SPEC | Medium | Low | REQ-HCW-005 (Optional MAY) explicitly leaves cadence open. Option A's single-batch cadence does not prevent Option B layering. |
| `tier-promotions.jsonl` file does not exist on first invocation | High (expected on this project) | Low | `WritePromotion` already handles file creation per V3R4-HARNESS-003. Verify via AC-HCW-001 (`wc -l > 0` after first invocation). |
| `learning.enabled: false` config gate not honored by classifier API | Low | Medium | AC-HCW-004 explicitly tests this. If classifier API does not check the gate itself, the new hook subcommand checks the gate before invoking the API. |
| Classifier error on corrupt usage-log entry crashes status verb | Low | High (regression) | AC-HCW-003 explicitly injects corrupt entry and verifies fail-open behavior. Implementation MUST capture classifier panic + return exit 1 (NOT propagate panic). |
| New subcommand `moai hook harness-classify` conflicts with retired `moai harness` namespace | None | None | `hook` namespace orthogonal to `harness` namespace per BC-V3R4-HARNESS-001-CLI-RETIREMENT. Verified by lint test on hook subcommand registration. |
| Workflow body Bash insertion breaks status verb on Windows (no Bash) | Low | Low | Existing workflow body already uses Bash; no new platform requirement introduced. |

## §6. Verification (M1 Exit Criteria)

After M1 completion, run the canonical 7-item read-only verification batch (per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution):

```bash
# 1. Full test suite (Go)
go test ./...

# 2. Coverage report
go test -coverprofile=cover.out ./internal/harness/... ./internal/hook/... ./internal/cli/...

# 3. Subagent-boundary grep
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go"

# 4. Sentinel-key audit
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ | head -20

# 5. CLI smoke check
go run ./cmd/moai --version
go run ./cmd/moai hook --help  # verify harness-classify subcommand listed

# 6. Benchmark micro-suite (optional)
go test -bench=. -benchmem -run=^$ ./internal/harness/...

# 7. Lint baseline
golangci-lint run --timeout=2m
```

All 7 invocations issued in parallel within a single orchestrator response turn per HARD batching rule.

## §7. Cross-references

- `spec.md §G` — wiring mechanism decision pointer to this §3
- `acceptance.md` — AC-HCW-001..006 binary criteria
- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution — 7-item verification batch
- `SPEC-V3R4-HARNESS-001` — `BC-V3R4-HARNESS-001-CLI-RETIREMENT` contract preserved
- `SPEC-V3R4-HARNESS-002` — observer hooks shipped (event source)
- `SPEC-V3R4-HARNESS-003` — classifier API shipped (`AggregatePatterns`, `ClassifyTier`, `WritePromotion`)
- `CLAUDE.local.md §24` — Harness namespace separation policy
