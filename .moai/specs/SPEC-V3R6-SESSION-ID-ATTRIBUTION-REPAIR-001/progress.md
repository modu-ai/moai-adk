---
id: SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
title: "Progress — Session-ID attribution dead-feature repair"
version: "0.1.0"
status: in-progress
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/session.go; internal/hook/session_start.go; internal/session/registry.go"
lifecycle: spec-anchored
tags: "session, attribution, multi-session, coordination, race-attribution, doctrine"
era: V3R6
---

# Progress — SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001

This file is the canonical progress tracker. §E.1 is populated at plan-phase
close; §E.2-§E.5 are populated by downstream agents (manager-develop for §E.2/§E.3,
manager-docs for §E.4/§E.5) per the Status Transition Ownership Matrix.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase completed:** 2026-06-17
- **Artifacts emitted:** 4 (spec.md, plan.md, acceptance.md, research.md) + this progress.md
- **Frontmatter `status`:** in-progress (M1 commit transition `draft → in-progress` owned by manager-develop per Status Transition Ownership Matrix)
- **SPEC ID pre-write self-check:** PASS (decomposition printed in manager-spec turn)
- **Re-grep verification:** COMPLETE against HEAD `12e20d190` — 5 discrepancies found (1 count correction: 9→14 variants in broader-context sweep, 4→canonical-surface count in iter-2; 4 scope clarifications; zero invalidated citations). See research.md §F.
- **Era classification:** V3R6 (explicit `era: V3R6` in frontmatter to suppress EraAutoDetected INFO per lifecycle-sync-gate.md H-override).
- **Tier:** M (standard). cycle_type: tdd.
- **Predecessors:** SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001 (completed, track 2); SPEC-V3R6-MULTI-SESSION-COORD-001 (Layer 1 primitives).
- **Plan-auditor iter-1:** FAIL 0.78 (2 BLOCKING D1+D2, 1 SHOULD-FIX D3, 3 MINOR D4/D5/D6). See `.moai/reports/plan-audit/SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001-iter1.md`.
- **Plan-auditor iter-2:** PASS-WITH-DEBT 0.86 (≥ Tier M 0.80 threshold). D1-D6 resolved. Implementation Kickoff Approval GRANTED by user (run-phase 진입 권장).
- **Implementation Kickoff Approval:** GRANTED 2026-06-17 (user decision — "run-phase 진입 (권장)").

## §E.2 Run-phase Evidence

### Milestone status (M1-M3 complete, M4-M6 in progress)

- **M1 (P1 write-path investigation, GATING):** COMPLETE. Root cause documented in research.md §D.0 (HookOutput.Data `json:"-"` structural gap — empirically confirmed by grep against HEAD `12e20d190`) + §D.1 (empty-SessionID gate bypass — reproduced by characterization test `TestSessionStartEmptySessionIDEmitsWarning`). REQ-WPR-003 warning added (GREEN). `moai session doctor` diagnostic implemented (REQ-WPR-001/002). AC-WPR-001/002/003/004 PASS.
- **M2 (`moai session current`, P2 Stage 1):** COMPLETE. 7 subcommands (5 original + current + doctor). AC-RDP-001/003/006 PASS. AC-RDP-002 happy-path PASS (post-M3 side-channel: `TestSessionCurrentReadsSideChannel`).
- **M3 (SessionStart additionalContext injection, P2 Stage 2):** COMPLETE. `hookSpecificOutput.AdditionalContext` injection + side-channel file write (`.moai/state/current-session-id.txt`). AC-RDP-004/005 PASS. `TestSessionStartHandler_Handle` updated (old nil-assertion codified the K5 defect; new test reflects REQ-RDP-004 correct behavior).
- **M4+M5 (P3 fallback doctrine canonicalization):** IN PROGRESS.
- **M6 (resume template citation + final verification):** PENDING.

### AC PASS/FAIL matrix (M1-M3; M4-M6 pending)

| AC ID | Severity | Status | Evidence |
|-------|----------|--------|----------|
| AC-WPR-001 | MUST | PASS | `TestSessionDoctorRegistryAbsent` + `TestSessionDoctorRegistryPresentWithEntry` — `moai session doctor` reports registry_exists + entry_count + root_cause_candidates |
| AC-WPR-002 | MUST | PASS | `doctorRootCauses()` enumerates 3 candidates (empty session_id, hook wrapper silent-exit, registry write failure); `TestSessionDoctorRegistryAbsent` asserts candidates non-empty + mention session_id/hook/wrapper |
| AC-WPR-003 | MUST | PASS | `TestSessionStartEmptySessionIDEmitsWarning` — stderr warning emitted when SessionID==""; hook exits 0 (non-blocking) |
| AC-WPR-004 | MUST (GATE) | PASS | research.md §D populated with empirically-reproduced root cause (D.0 structural + D.1 gate bypass); M2-M6 unblocked |
| AC-RDP-001 | MUST | PASS | `TestSessionCurrentListedInHelp` — 7 subcommands listed (register/heartbeat/deregister/list/purge/current/doctor) |
| AC-RDP-002 | MUST | PASS | `TestSessionCurrentReadsSideChannel` — UUID resolved from side-channel file post-M3; `available:true` |
| AC-RDP-003 | MUST | PASS | `TestSessionCurrentFallbackWhenNoSideChannel` — exit 0 + canonical fallback when no side-channel; `TestSessionCurrentJSONFallback` — available:false, source:fallback |
| AC-RDP-004 | MUST | PASS | `TestSessionStartInjectsAdditionalContext` — hookSpecificOutput.AdditionalContext carries UUID; `TestSessionStartWritesSideChannel` — side-channel file written |
| AC-RDP-005 | MUST | PASS | `TestSessionStartAdditionalContextStrictlyAdditive` — existing multi_session_register=ok marker preserved + new injection present |
| AC-RDP-006 | SHOULD | PASS | `TestSessionCurrentFallbackWhenNoSideChannel` + `TestSessionCurrentShowFallbackFlag` — canonical fallback emitted |
| AC-FBC-001 | MUST | PENDING (M4+M5) | — |
| AC-FBC-002 | MUST | PENDING (M4+M5) | — |
| AC-FBC-003 | MUST | PENDING (M4+M5) | — |
| AC-FBC-004 | MUST | PENDING (M4+M5) | — |
| AC-FBC-005 | SHOULD | PENDING (M4+M5) | — |
| AC-MSC-001 | MUST | PASS | existing session tests pass unchanged (`TestSessionRegisterSmoke`, `TestSessionListSmoke`, `TestSessionHeartbeatSmoke`, `TestSessionDeregisterSmoke`, `TestSessionPurgeSmoke`, `TestSessionFiveVerbsHelp`) |
| AC-MSC-002 | MUST | PASS | `FormatStderrReminder` unchanged (no edits to registry.go L424-448); existing `TestSessionStartMultiSessionProtocol*` pass |
| AC-MSC-003 | SHOULD | PENDING (M6) | — |

### Verbatim test/build/vet evidence

**go build ./... (2026-06-17, HEAD with M1-M3 changes):**
```
$ go build ./...
(exit 0, no output)
```

**go vet ./... (2026-06-17):**
```
$ go vet ./...
(exit 0, no output)
```

**go test ./internal/session/... ./internal/cli/... (2026-06-17):**
```
$ go test ./internal/session/... ./internal/cli/...
ok  	github.com/modu-ai/moai-adk/internal/session	8.730s
ok  	github.com/modu-ai/moai-adk/internal/cli	10.190s
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	(cached)
ok  	github.com/modu-ai/moai-adk/internal/cli/pr	(cached)
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	(cached)
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	(cached)
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	2.276s
```

**go test -count=1 ./internal/hook/ (2026-06-17, SessionStart + wrapper):**
```
$ go test -count=1 ./internal/hook/ -run 'TestSessionStart|TestHookWrapper'
ok  	github.com/modu-ai/moai-adk/internal/hook	0.793s
```

### Gaps (verification-claim-integrity §3.4)

- **Pre-existing baseline failure (OUT OF SCOPE):** `TestCollectMemory` + `TestCollectMemory_AutoCompactScaling` in `internal/statusline` fail on the clean `12e20d190` baseline (confirmed via `git stash` + retest). These are a leftover from the immediately-preceding STATUSLINE-PRESET-RETIRE SPEC, NOT caused by this SPEC's changes. Scope discipline: session-id attribution only — these are NOT fixed here.
- **AC-FBC-001..005, AC-MSC-003:** PENDING M4+M5+M6 (doctrine canonicalization + resume template citation).
- **Full-suite statusline failure does NOT affect this SPEC's in-scope packages** (session/cli/hook all PASS).

### Residual-risk (verification-claim-integrity §3.5)

- The `additionalContext` injection is lost after `/clear`/compaction (spec.md §F.2). The side-channel file (`.moai/state/current-session-id.txt`) persists across compaction, so `moai session current` can re-read the UUID post-compaction — but only if the SessionStart hook ran with a non-empty SessionID in this project directory.
- Headless `-p` invocations without hooks bypass SessionStart entirely; `moai session current` returns the canonical fallback (REQ-RDP-006) in that case.

## §E.3 Run-phase Audit-Ready Signal

_<pending — manager-develop populates after M4-M6 completion + final full-suite verification>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates with sync_commit_sha after CHANGELOG/README/docs sync>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — manager-docs populates with mx_commit_sha after 4-phase close>_
