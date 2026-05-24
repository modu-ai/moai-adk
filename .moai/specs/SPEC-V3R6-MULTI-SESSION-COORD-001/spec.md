---
id: SPEC-V3R6-MULTI-SESSION-COORD-001
title: "Multi-Session Coordination — 4-Layer Race Mitigation Architecture"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: "GOOS행님"
priority: P1
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "multi-session, coordination, registry, hook, race-mitigation"
---

# SPEC-V3R6-MULTI-SESSION-COORD-001 — Multi-Session Coordination

## HISTORY

- 2026-05-24: Plan-phase artifacts created. Tier M. 4-layer architecture (Go primitive + CLI + hook + pre-spawn rule extension). Origin: ARR-001/SIV-001 race incident 2026-05-24 + L52 lesson + CLAUDE.local.md §23.8.
- 2026-05-24: Plan-phase iter-2 — D1 broken AC reference (iter-1 cited a non-existent AC ID) resolved by adding AC-COORD-013 (CLI 5 verbs verification, REQ-COORD-021 trace). D2 six uncovered REQs resolved: REQ-COORD-006/018/020/021 covered by new AC-COORD-013/014/015/016; REQ-COORD-012/024 documented as L48 trace-orphan in §C.5. Case 3 staging-area race (commit `24cb6ad4b`, 20× scope drift) added to §A.1 as 3rd empirical case + §F.6 mitigation extended with L4 scope reinforcement (`git diff --cached --name-only` pre-commit assertion).

## §A Background

### §A.1 Motivating Example (Verbatim — ARR-001 3-Session Race)

SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (ARR-001)의 4-phase lifecycle가 2026-05-24일 **3개 orchestrator 세션에 걸쳐 분산 실행**되었다:

- **Session A** (earlier in day): plan + run + run-backfill — commits `e2fbe4d60` + `e6ad82031` + `e48af1792`
- **Session B** (`f3d5f57e-4620-48a3-9a65-a9fe32b2816c`, ≈20:44 KST): sync + sync-backfill — commits `11abb9a30` + `a25476e7e`
- **Session C** (`cd8d8946-e06f-4c76-a3ab-869eba092356`, ≈21:25 KST): Mx-chore EVALUATE-PASS — commit `e0c334e18`

**Race signal**: Session C의 pre-spawn `git fetch origin main && git rev-list --count --left-right origin/main...HEAD`은 `0 0` (clean ahead)을 반환했다. 이유는 Session B의 sync commits가 이미 Session C의 local main에 fetch 시점에 **silently fast-forward**되었기 때문이다. Session C는 `/moai sync SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001`을 다시 위임할 뻔했고, 만약 그랬다면 다음 race가 발생했을 것:

1. **CHANGELOG.md 중복 entry race** — manager-docs가 동일 `[Unreleased]` 섹션에 ARR-001 entry를 다시 push
2. **4 frontmatter status field overwrite race** — `status: implemented` → `status: implemented` (no-op이지만 commit 발생)
3. **progress.md §E.4 sync block overwrite race** — sync_complete_at 타임스탬프 재기록 + sync_commit_sha 추적 혼란

**Detection retrospective**: 본 race는 자동화된 coordination signal로 감지된 것이 아니라, Session C에서 사용자 의심 → `git log --all --oneline | grep ARR` 수동 호출 → 11abb9a30 + a25476e7e 발견의 순서로 **retrospective**하게만 감지되었다. 같은 패턴은 SIV-001 (SPEC-V3R6-SPEC-ID-VALIDATION-001)에서도 plan-phase와 run-phase가 분리된 세션에서 진행되며 유사 race 정황이 관찰되었다 (L52 NEW lesson).

**Case 2 (concurrent plan-phase race, 2026-05-24 ≈22:00 KST)**: 본 SPEC의 plan-phase delegation 자체가 진행되는 동안 (orchestrator spawn `21:53` → manager-spec return `22:00`), 또 다른 orchestrator session이 SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 plan-phase를 동시 진행했다. 두 session은 동일 project root + 동일 memory hash 위에서 ≈7분간 coordination signal 없이 작업했다. 충돌은 발생하지 않았다 (두 SPEC이 disjoint한 디렉토리만 수정) — 하지만 7분 race window 자체가 *"중복되지 않는 작업을 사용자가 자발적으로 선택한다"*는 가정에 의존하고 있음을 입증한다. 이 가정은 scale에서 또는 공유 rule 파일 (예: `agent-common-protocol.md`, `session-handoff.md`, `output-style/moai.md`)을 동시 수정하려는 SPEC들에서는 깨진다.

Commits: `e5b2859a9` (HARNESS-PROPOSAL-GEN-001 plan) + `2b99be826` (progress.md backfill).
Cross-reference: 본 SPEC plan.md §E PRESERVE list snapshot이 spawn 시점에는 `?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` (untracked)였는데, manager-spec return 시점에는 main에 통합되어 status에서 사라진 것이 동시 활동의 증거.

**Case 3 (staging-area race, 2026-05-24 ≈22:54 KST)**: 본 SPEC plan-phase 직후 progress.md `§C plan-auditor iter-1` row backfill을 위한 chore commit이 STAGING AREA RACE를 일으켰다. 의도는 `git add .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md` (1-file scope)였으나, 실제 commit `24cb6ad4b`는 14 files (1881 insertions vs intended ~92 insertions, **20× scope drift**)를 포함하여 origin/main에 push되었다. 흡수된 13 files는 concurrent session의 SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 run-phase + sync-phase + Mx-phase 산출물이었다 (e5b2859a9 + 2b99be826 + 추가 commits의 working tree 잔여물).

**Race signal**: `git status`가 본 commit 직전 14 modified/added files을 보였고, `git add .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md` 단일 path만 staging 의도였으나, 본 세션의 prior `git add -A` 또는 staging-area carry-over로 13 files이 silent하게 staged 상태였다. `git commit -m "chore..."`이 staged area 전체를 commit하면서 race가 표면화되었다. Push range mismatch (`535b5b6ae..24cb6ad4b` 의도 1 file vs 실제 14 files)로 retrospectively 감지되었다.

**이 케이스는 L44 HARD (pre-spawn fetch obligation)의 한계를 입증한다**: fetch는 동일 commit-base를 보장할 뿐 staging-area scope-drift는 다루지 않는다. §F.6 mitigation에서 별도 다룬다.

Commit reference: `24cb6ad4b` (본 SPEC chore commit이지만 13 PROPOSAL-GEN-001 files 흡수); `bdee48858` (governance documentation of this incident).

### §A.2 Cross-References (Verbatim Citation Sources)

- `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/progress.md` §E.5 `multi_session_coordination_note`
- `.moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/progress.md` §F `multi_session_coordination_note`
- `MEMORY.md` L52 lesson — multi-session race coordination (SendMessage background resume vs user AskUserQuestion 응답 race 패턴, sha256 결정성으로 결과 동등 + post-hoc canonical verify)
- `CLAUDE.local.md` §23.8 Multi-Session Race Mitigation — defense-in-depth policy at user-facing layer
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check — 현재 2-command batch (L1 primitive)
- `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume Pattern — L2/L3 worktree as race-elimination alternative

### §A.3 Problem Statement

동일 project root + 동일 memory hash (`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)를 공유하는 2개 이상 Claude Code 세션이:

1. **공유 git working tree**를 동시 modify할 수 있음 — 현재 mitigation은 `git fetch + git rev-list` 2-command pre-spawn batch (L1, advisory만)
2. **공유 MEMORY.md**의 paste-ready resume을 양쪽 세션이 동시 consume할 수 있음 — 현재 mitigation은 사용자 수동 규율 (CLAUDE.local.md §23.8 정책 3)
3. **동일 SPEC**에 대한 work allocation의 자동화된 coordination 없음 — 본 SPEC이 해결할 핵심 갭

**핵심 미해결 갭**: orchestrator는 자신의 session_id조차 자기 자신의 work에 attribution하지 못하며, 다른 세션이 동일 SPEC을 작업 중인지 query할 수단이 없다.

## §B Goals / Non-Goals

### §B.1 Goals

1. **Active-sessions registry**: 4-layer 아키텍처의 L1 — 동일 project root에서 작동하는 모든 Claude Code 세션이 자신의 `{session_id, spec_id, phase, started_at, last_heartbeat, pid, host, cwd}`를 atomic-write 방식으로 등록/heartbeat/조회/deregister 가능한 Go primitive + CLI 노출
2. **Paste-ready session_id tagging**: L2 — paste-ready resume 메시지 + auto-memory `project_*.md` 파일이 `source_session_id` 필드를 표면화하여 사용자가 어느 세션이 resume을 생성했는지 시각적으로 식별 가능
3. **SessionStart hook 통합**: L3 — 새 세션 시작 시 자동 RegisterSession + 30분 임계값 zombie purge + 다른 active 세션 surface (stderr system-reminder)
4. **Pre-spawn HARD rule 확장**: L4 — 현재 git fetch + git rev-list 2-command batch에 `moai session list --json --filter-spec=<SPEC-ID>` 3rd command 추가, 다른 세션 활동 감지 시 AskUserQuestion (wait/override/abort)

### §B.2 Non-Goals (out of scope)

1. **Cross-machine sync** — registry는 per-machine local file (`.moai/state/active-sessions.json`)에만 작동, 다른 머신의 세션과 coordination 없음
2. **Distributed lock** — registry는 advisory만 (orchestrator가 query 후 자율 판단), strong mutual exclusion lock 미제공
3. **Multi-user concurrent** — single user의 multi-session 패턴만 지원, multi-tenant scenarios out of scope
4. **MEMORY.md write coordination** — paste-ready resume tagging만 다루며, MEMORY.md 자체의 concurrent write conflict는 별도 SPEC (현재 git working tree commit으로 우회)
5. **Crash recovery beyond stale purge** — 30분 heartbeat threshold 초과 entry만 purge, sophisticated crash detection (PID liveness check, signal handlers) out of scope
6. **GUI/dashboard** — CLI subcommand만 노출, web UI 미제공

### §B.3 Out of Scope

- Cross-machine session coordination (only same-host sessions sharing project memory hash)
- Distributed lock service (no Redis, no Postgres, no remote coordinator)
- Multi-user / multi-tenant support (single-user, single-machine assumption)
- Cross-project session isolation (only within one project root)
- Session ID generation (rely on Claude Code's existing UUID assignment)
- Modification of paste-ready emit logic in output-style §8 beyond `source_session_id` metadata addition
- Hook event additions beyond `SessionStart` (Notification / PostToolUse / Stop out of scope)
- UI/TUI changes (CLI text output + `--json` flag only, no web dashboard)
- MEMORY.md concurrent write coordination (paste-ready tagging only)
- Crash recovery beyond 30-minute stale heartbeat purge (no PID liveness check, no signal handlers)
- Registry schema versioning / migration tooling (REQ-COORD-024 freezes schema)
- Inter-session messaging (Agent Teams handles that, not this SPEC)

## §C Requirements (EARS Format)

### §C.1 L1 — Active-Sessions Registry (Go Primitive + CLI)

**REQ-COORD-001** (Ubiquitous): The system shall provide a Go package `internal/session/` that exposes `RegisterSession`, `Heartbeat`, `DeregisterSession`, `QueryActiveWork`, `PurgeStale` functions operating on a JSON-backed registry file at `.moai/state/active-sessions.json`.

**REQ-COORD-002** (Ubiquitous): The registry entry schema shall be `{session_id: UUID-string, spec_id: string, phase: enum(plan|run|sync|mx|none), started_at: ISO8601-string, last_heartbeat: ISO8601-string, pid: int, host: string, cwd: string}`. The registry file shall be a JSON array of such entries.

**REQ-COORD-003** (Event-Driven): When `RegisterSession(session_id, spec_id, phase)` is invoked, the system shall atomically append a new entry with `started_at = last_heartbeat = time.Now().UTC()`, `pid = os.Getpid()`, `host = os.Hostname()`, `cwd = os.Getwd()` using lockfile-based atomic write semantics.

**REQ-COORD-004** (Event-Driven): When `Heartbeat(session_id)` is invoked, the system shall atomically update the matching entry's `last_heartbeat` field to `time.Now().UTC()` without modifying other fields.

**REQ-COORD-005** (Event-Driven): When `DeregisterSession(session_id)` is invoked, the system shall atomically remove the matching entry from the registry; the operation shall be idempotent (no error if entry already removed).

**REQ-COORD-006** (Event-Driven): When `QueryActiveWork(opt_spec_id)` is invoked, the system shall return a slice of entries; if `opt_spec_id` is non-empty, only entries matching `spec_id == opt_spec_id` shall be returned.

**REQ-COORD-007** (State-Driven): While the registry contains entries whose `last_heartbeat` is older than 30 minutes, when `PurgeStale(threshold_minutes=30)` is invoked, the system shall atomically remove those stale entries and return the count of purged entries.

**REQ-COORD-008** (Ubiquitous): The system shall use lockfile-based atomic write semantics (`.moai/state/active-sessions.json.lock` flock-based or `O_CREATE|O_EXCL` sentinel) to prevent corruption under concurrent registry mutation from multiple sessions on the same host.

### §C.2 L2 — Paste-ready session_id Tagging

**REQ-COORD-009** (Ubiquitous): The system shall extend `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format to require a `source_session_id: <UUID>` line within the paste-ready resume message preamble (immediately after the top cut-line marker or as a Block 2 sub-field).

**REQ-COORD-010** (Ubiquitous): The system shall extend `.claude/output-styles/moai/moai.md` (or equivalent orchestrator output-style template) §8 Session Handoff section to direct the orchestrator to emit `source_session_id` in paste-ready resume messages.

**REQ-COORD-011** (Ubiquitous): The system shall require new auto-memory files (`project_*.md`) to include a `source_session_id: <UUID>` field in their YAML-like header section or first prose line. MEMORY.md index entries SHALL include a `(session: <UUID-8-char-prefix>)` parenthetical annotation.

**REQ-COORD-012** (Optional): Where existing memory files lack `source_session_id` (pre-implementation entries), the system shall continue to function correctly; absence of the field shall NOT be treated as an error by any consumer.

### §C.3 L3 — SessionStart Hook Multi-Session Surface

**REQ-COORD-013** (Event-Driven): When `internal/hook/session_start.go` is invoked by Claude Code's SessionStart hook event, the system shall call `session.RegisterSession(stdin_session_id, "(none)", "(none)")` as Step 1 to register the new session with no SPEC scope initially.

**REQ-COORD-014** (Event-Driven): When the SessionStart hook completes Step 1 RegisterSession, the system shall call `session.PurgeStale(30)` as Step 2 to remove zombie entries from crashed sessions.

**REQ-COORD-015** (Event-Driven): When the SessionStart hook completes Step 2 PurgeStale, the system shall call `session.QueryActiveWork()` as Step 3; if the result contains other entries (excluding the just-registered current session), the system shall emit a system-reminder to stderr describing `(session_id_prefix, spec_id, phase, age_minutes)` for each other active session.

**REQ-COORD-016** (Ubiquitous): The hook wrapper script `.claude/hooks/moai/handle-session-start.sh` shall pass the Claude Code session_id from stdin JSON to `moai hook session-start` invocation so that `RegisterSession` receives the canonical Claude Code session_id (not a freshly generated UUID).

### §C.4 L4 — Pre-Spawn HARD Rule Extension

**REQ-COORD-017** (Ubiquitous): The system shall extend `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check to add a 3rd command `moai session list --json --filter-spec=<SPEC-ID>` after the existing 2-command batch (git fetch + git rev-list).

**REQ-COORD-018** (Event-Driven): When the orchestrator invokes the 3-command pre-spawn batch and the 3rd command's output is `[]` (empty JSON array), the orchestrator shall proceed normally (no other session on this SPEC).

**REQ-COORD-019** (Event-Driven): When the orchestrator invokes the 3-command pre-spawn batch and the 3rd command's output contains entries from sessions other than the current session, the orchestrator shall STOP, surface the entries via prose summary, and ask the user via AskUserQuestion to choose between (wait / override / abort).

**REQ-COORD-020** (Ubiquitous): The current 2-command batch (git fetch + git rev-list) shall be preserved verbatim; the 3rd command is additive only. Backward compatibility with sessions not yet running the new hook (no registry entries) is automatic via REQ-COORD-018 (empty result behavior).

### §C.5 Cross-cutting

**REQ-COORD-021** (Ubiquitous): The CLI subcommand `moai session` shall expose 5 verbs: `register`, `heartbeat`, `deregister`, `list`, `purge`. Each verb shall accept human-readable output by default and `--json` for machine-readable output.

**REQ-COORD-022** (Ubiquitous): All registry write operations shall be cross-platform compatible across linux, darwin, windows; specifically, atomic-write semantics shall use platform-appropriate primitives (POSIX `O_CREATE|O_EXCL` + rename on linux/darwin, equivalent on windows).

**REQ-COORD-023** (Ubiquitous): The Go package and CLI shall pass `go vet ./internal/session/... ./cmd/moai/...` and `golangci-lint run --timeout=2m` with zero issues.

**REQ-COORD-024** (Unwanted): The system shall NOT modify the registry file format or schema after initial implementation without a follow-up SPEC; the schema in REQ-COORD-002 is the canonical contract.

**REQ-COORD-012 + REQ-COORD-024 — Trace-orphan tolerance**: REQ-COORD-012 (Optional: backward compat for memory files lacking source_session_id) and REQ-COORD-024 (Unwanted: schema freeze post-implementation) are deliberately NOT covered by dedicated ACs per L48 SSOT canonical principle (spec.md is single source of truth for Requirements; Optional/Unwanted variants may remain trace-orphan if their behavioral assertion is captured implicitly by other ACs or by the absence of failure modes). REQ-COORD-012 is implicitly verified by AC-COORD-015 (orchestrator proceed-on-empty behavior — which applies equally when entries lack source_session_id as a degenerate empty-result case). REQ-COORD-024 is a future-looking prohibition with no observable behavior at run-phase; its enforcement is meta (any SPEC that modifies the schema must explicitly cite REQ-COORD-024 as superseded), and therefore lives outside the AC matrix scope.

## §D Architecture

### §D.1 Layer Breakdown

```
┌─────────────────────────────────────────────────────────────────┐
│ L4 — Pre-Spawn HARD Rule (agent-common-protocol.md)              │
│   3-command batch: git fetch + git rev-list + moai session list  │
│   Triggers AskUserQuestion (wait/override/abort) on conflict     │
└────────────────────┬────────────────────────────────────────────┘
                     │ depends on
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ L3 — SessionStart Hook (internal/hook/session_start.go)          │
│   Step 1: RegisterSession(stdin_session_id, "(none)", "(none)")  │
│   Step 2: PurgeStale(30)                                         │
│   Step 3: QueryActiveWork() → stderr system-reminder if others   │
└────────────────────┬────────────────────────────────────────────┘
                     │ depends on
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ L2 — Paste-ready session_id Tagging (rules + output-style)       │
│   source_session_id in resume preamble                           │
│   project_*.md auto-memory frontmatter                           │
│   MEMORY.md index entry (session: <prefix>) annotation           │
└────────────────────┬────────────────────────────────────────────┘
                     │ depends on
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ L1 — Active-Sessions Registry (Go primitive + CLI)               │
│   .moai/state/active-sessions.json (atomic-write JSON)           │
│   internal/session/registry.go (~150 LOC)                        │
│   cmd/moai/session.go (~80 LOC) — 5 subcommands                  │
└─────────────────────────────────────────────────────────────────┘
```

### §D.2 Data Flow (Example: 2 Sessions, 1 SPEC)

```
Session A (SPEC-X plan-phase):
  SessionStart hook → RegisterSession(A_uuid, "(none)", "(none)")
  /moai plan SPEC-X → orchestrator calls RegisterSession(A_uuid, "SPEC-X", "plan")
  ... work in progress, periodic Heartbeat(A_uuid) ...

Session B (concurrent, attempts SPEC-X sync-phase):
  SessionStart hook → RegisterSession(B_uuid, "(none)", "(none)")
                    → PurgeStale(30) → 0 removed
                    → QueryActiveWork() → [A entry] → stderr: "session A_prefix active on SPEC-X phase=plan age=15m"
  /moai sync SPEC-X → orchestrator pre-spawn batch:
    1. git fetch origin → 0 0
    2. git rev-list... → 0 0
    3. moai session list --json --filter-spec=SPEC-X → [{session_id: A_uuid, spec_id: "SPEC-X", phase: "plan", ...}]
  → orchestrator detects A entry, STOP, AskUserQuestion (wait/override/abort)
  → user chooses "wait" → orchestrator polls QueryActiveWork periodically
  → user chooses "override" → orchestrator proceeds + logs warning + records override in progress.md
  → user chooses "abort" → orchestrator returns blocker
```

### §D.3 Atomic-Write Strategy

Registry write operations use the following pattern (cross-platform):

1. Acquire advisory lock on `.moai/state/active-sessions.json.lock` (POSIX `flock` LOCK_EX or windows equivalent)
2. Read current registry file (or treat missing file as empty array `[]`)
3. Parse JSON into `[]Entry` slice
4. Apply mutation (append for Register, update for Heartbeat, filter for Deregister/PurgeStale)
5. Marshal back to JSON with stable key ordering
6. Write to temp file `.moai/state/active-sessions.json.tmp`
7. `os.Rename` temp file over canonical file (atomic on POSIX, mostly-atomic on windows via MoveFileEx with MOVEFILE_REPLACE_EXISTING)
8. Release lock

This pattern guarantees no partial-write corruption even if multiple Go processes on the same host attempt concurrent mutation.

### §D.4 Backward Compatibility

- **Sessions without hook installed**: SessionStart hook is a Claude Code feature; sessions running before hook installation simply don't register. `QueryActiveWork()` returns only registered sessions. New 3rd pre-spawn command returns `[]` for un-tracked sessions → orchestrator proceeds normally (no false positive).
- **Memory files without source_session_id**: REQ-COORD-012 explicitly allows this. No consumer rejects entries lacking the field.
- **Existing 2-command batch**: REQ-COORD-020 preserves verbatim. Only additive 3rd command introduced.

## §E Implementation Phases

### §E.1 Milestone Sequence (M1 → M5, dependency order)

| Milestone | Deliverable | Dependency | Verification |
|-----------|-------------|------------|--------------|
| M1 | `internal/session/registry.go` + `internal/session/registry_test.go` Go primitive with 5 functions + atomic-write semantics + unit tests | none | AC-COORD-001..004 + AC-COORD-011..012 |
| M2 | `cmd/moai/session.go` CLI subcommand (5 verbs, --json flag) + `internal/cli/root.go` registration | M1 | AC-COORD-013 (verified via `moai session --help` + per-verb smoke test) |
| M3 | `internal/hook/session_start.go` modification (3-step protocol) + `.claude/hooks/moai/handle-session-start.sh` modification (session_id pass-through) | M1 (registry) | AC-COORD-007..008 |
| M4 | `.claude/rules/moai/core/agent-common-protocol.md` extension + `.claude/rules/moai/workflow/session-handoff.md` extension + `.claude/output-styles/moai/moai.md` extension (paste-ready tagging) | none (documentation) | AC-COORD-005..006 + AC-COORD-009..010 |
| M5 | progress.md finalization + frontmatter status `draft → implemented` for all 4 artifacts + run-phase evidence/audit-ready signal | M1-M4 | progress.md §D run-phase complete + B12 self-test |

### §E.2 Critical Path

M1 (Go primitive) → M2 (CLI) and M3 (hook) in parallel → M4 (rule extension) → M5 (progress finalization).

M4 may be authored in parallel with M1-M3 since it is documentation-only with no code dependency.

## §F Risks

### §F.1 Atomic-Write Portability

**Risk**: `os.Rename` is atomic on POSIX but has edge cases on windows (file in use by other process). Lockfile semantics also differ.

**Mitigation**:
- Use `golang.org/x/sys/windows` for windows-specific MoveFileEx with proper flags
- Test on linux + darwin + windows CI matrix (AC-COORD-011)
- Document any windows-specific behavior in `internal/session/doc.go`

### §F.2 Hook Timeout

**Risk**: SessionStart hook 3-step protocol (Register + Purge + Query) must complete within Claude Code hook timeout (5 seconds default per CLAUDE.local.md §7).

**Mitigation**:
- Benchmark hook execution time during M3 implementation
- If approaching 5 sec, parallelize Register + Purge + Query (Purge and Query are independent of Register's specific entry)
- If hook times out, registry write may be lost; orchestrator falls back to "no registry" behavior (REQ-COORD-018 empty result)

### §F.3 Stale Heartbeat Threshold Tuning

**Risk**: 30-minute threshold may be too aggressive (false-positive removal of legitimate long-running session) or too lax (zombie entries linger blocking other sessions).

**Mitigation**:
- 30 min is initial default
- Make threshold configurable via `.moai/config/sections/session.yaml` (future SPEC if needed)
- Active sessions are expected to Heartbeat every 5-10 min (driven by orchestrator periodic refresh, not part of this SPEC)
- Empirical tuning post-merge

### §F.4 False-Positive: Legitimate Parallel Sessions

**Risk**: User intentionally runs 2 sessions on different SPECs simultaneously. Pre-spawn check might surface noise.

**Mitigation**:
- REQ-COORD-017 filters by `--filter-spec=<SPEC-ID>` — only same-SPEC conflicts surface
- Different-SPEC parallel work emits no warning from pre-spawn batch
- SessionStart hook Step 3 stderr message is informational only (not blocking)

### §F.5 Session_id Collision

**Risk**: Two sessions accidentally receive same session_id (extremely unlikely with UUIDs but theoretically possible if Claude Code reuses session_id across reconnect).

**Mitigation**:
- Claude Code session_id is UUID v4 (collision probability ~10^-37)
- RegisterSession is idempotent on duplicate session_id (update existing entry instead of duplicate) — minor schema decision deferred to run-phase
- Heartbeat/Deregister are also idempotent on missing entry

### §F.6 Multi-Session Race on Implementation of THIS SPEC

**Risk**: This SPEC modifies `internal/session/registry.go` which is the very mechanism it implements. Implementation cannot use the registry to coordinate its own implementation.

**Mitigation**:
- Implementation SPEC is single-session (this SPEC's run-phase delegated to manager-develop in one session)
- Pre-spawn check (current 2-command batch) suffices for this SPEC's run-phase
- Post-merge, all subsequent SPECs benefit from the 3-command batch
- **L4 scope reinforcement (added per Case 3 empirical evidence)**: pre-commit assertion `git diff --cached --name-only | sort -u` MUST be verified against the intended scope before EVERY commit on shared working tree. When the cached file set diverges from intent, the session MUST `git reset` (atomic clear of staging area, no destructive operation) and re-stage with explicit per-path `git add <specific-path>` invocations. This complements but does NOT replace L44 HARD pre-spawn fetch.
- **Cross-reference**: `CLAUDE.local.md` §23.8 Multi-Session Race Mitigation (defense-in-depth policy at user-facing layer) is updated in parallel by orchestrator to reflect Case 3 evidence.

## §G Cross-References

### §G.1 Related SPECs

- **SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001** (implemented) — Motivating race incident (§A.1)
- **SPEC-V3R6-SPEC-ID-VALIDATION-001** (implemented) — Sibling SPEC also affected by L52 race pattern
- **SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001** (implemented) — Hook system canonical reference (`internal/hook/` package patterns)
- **SPEC-V3R6-LEGACY-CLEANUP-001** (implemented) — Original 2026-05-23 race incident (PR push range mismatch detection)

### §G.2 Related Rules

- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check — current 2-command batch (extended by L4)
- `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format — extended by L2 to require source_session_id
- `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol — `[SUPERSEDED by ...]` convention used by memory tagging (L2)
- `CLAUDE.local.md` §23.8 Multi-Session Race Mitigation — defense-in-depth policy (this SPEC operationalizes Layer 1 of the policy)

### §G.3 Related Memory / Lessons

- MEMORY.md L52 — multi-session race coordination pattern (canonical source of motivation)
- `project_sprint8_arr001_run_complete` — ARR-001 4-phase commit history reference
- `project_sprint8_siv001_sync_complete` — SIV-001 sync-phase reference

### §G.4 Related Constitution / Frozen Zones

- **ZONE:Frozen**: REQ-COORD-002 (registry schema) — schema modifications require follow-up SPEC per REQ-COORD-024
- **ZONE:Evolvable**: 30-min heartbeat threshold (REQ-COORD-007) — empirically tunable

## §H Exclusions (What NOT to Build)

[HARD] The following are explicitly out of scope and SHALL NOT be implemented in this SPEC:

1. **Cross-machine sync** — registry is per-machine only. No network protocol, no remote query, no cluster mode.
2. **Distributed lock / strong mutex** — registry is advisory. Orchestrator may override after AskUserQuestion. No `flock` on shared filesystem, no etcd/zookeeper.
3. **Multi-user / multi-tenant** — single user's multi-session pattern only. No user identity, no permission model.
4. **MEMORY.md concurrent write coordination** — only paste-ready resume tagging (L2). MEMORY.md mutual exclusion deferred.
5. **Crash recovery beyond stale purge** — 30-min threshold only. No PID liveness check, no signal handlers, no zombie reaping.
6. **GUI / web dashboard** — CLI only. No HTML output, no TUI beyond plain text + JSON.
7. **Session resume / replay** — registry tracks "what's active now". No history, no replay of past sessions.
8. **Inter-session messaging** — registry is read-only query + write self. No SendMessage between sessions (Agent Teams handles that).
9. **Automatic session_id generation** — SessionStart hook receives session_id from Claude Code stdin. CLI `moai session register` accepts session_id as required argument (no auto-gen for CLI invocations outside hook).
10. **Schema versioning / migration** — REQ-COORD-024 prohibits schema change without follow-up SPEC. No migration tooling.

[HARD] Build-time SHALL NOT introduce dependencies on: external lock services (etcd, consul, redis), distributed databases, network sockets, web frameworks, or any non-stdlib package beyond `golang.org/x/sys` (for windows MoveFileEx).
