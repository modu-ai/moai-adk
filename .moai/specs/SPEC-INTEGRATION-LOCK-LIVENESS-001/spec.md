---
id: SPEC-INTEGRATION-LOCK-LIVENESS-001
title: "Release-integration lock: anchor holder liveness to the session process, not the acquire CLI process (card t298)"
version: "0.1.0"
status: draft
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban, internal/cli, internal/hook"
lifecycle: spec-anchored
tags: "integration-lock, liveness, kanban, factory, t298"
tier: M
---

# SPEC-INTEGRATION-LOCK-LIVENESS-001 — Integration Lock Holder-Liveness Anchor

## §A Context and Problem

The release-integration window lock (`internal/kanban/integration_lock.go`, card t194) is the
mechanical layer under the lane-serialization doctrine (`.claude/rules/moai/workflow/kanban-dispatch.md`
§ "Integration into the release branch is self-served"). Its design premise is that a window
record's validity is decided by "the recorded holder's liveness" — a premise that is currently
broken at the point of recording: `AcquireIntegrationLock` fills an unset `want.PID` with
`os.Getpid()` (integration_lock.go:174-176), which is the pid of the **short-lived
`moai integration acquire` CLI process**, not the lane session that holds the window. That
process exits at command return, so `Stale()` (integration_lock.go:91-99), which judges liveness
solely via `!FactoryProcessAlive(l.PID)`, reports true for **every** record immediately after it
is written. The window always reads "reclaimable"; takeover needs no `--force`; two lanes can
merge into the shared integration worktree concurrently — the exact failure the lock exists to
prevent.

Live reproduction by the lead (2026-08-27): `moai integration acquire --name probe-test` followed
immediately by `moai integration status` printed BOTH "held by a session that is gone
(reclaimable)" AND "holder: \<session> (pid 6397)" — pid 6397 was the CLI process itself.

The same defect shape was already solved once in this repository for the session work registry:
`internal/session/session_pid.go` exists precisely because "the registry is written from a hook
subprocess that exits within milliseconds, so its own PID is dead by the time any reader probes
it". Its resolution chain (`MOAI_SESSION_PID` env stamp naming a live process → nearest
non-wrapper ancestor → fallback) is the reusable primitive this SPEC anchors the integration
lock to.

## §B Verified Premises (baseline @ d29b8942e, branch WT-integration-lock)

All line numbers measured on this tree; re-verify before citing elsewhere.

1. `internal/kanban/integration_lock.go:71-79` — `IntegrationLock{SessionID, SessionName, PID,
   Branch, Worktree, AcquiredAt, Card}`. `SessionID`/`SessionName` exist but play no role in
   `Stale()`.
2. `internal/kanban/integration_lock.go:174-176` — `if want.PID == 0 { want.PID = os.Getpid() }`
   records the one-shot CLI process pid. This line is the defect.
3. `internal/kanban/integration_lock.go:91-99` — `Stale()`: `Held()` gate → `PID <= 0` returns
   false (treated live, conservative) → `!FactoryProcessAlive(l.PID)`.
4. `internal/kanban/integration_lock.go:84-90` — the doctrine comment on the asymmetry:
   "Indeterminate reads as LIVE, deliberately: treating an unprobeable holder as dead would
   clear a window that may still be in use, and the cost of a false 'stale' (two lanes merging
   at once) is the failure this lock exists to prevent, while the cost of a false 'live' is one
   operator asking the holder to release. **The asymmetry is not close.**" This SPEC treats that
   asymmetry as a binding constraint: every failure mode must resolve toward TREAT-AS-LIVE.
5. `internal/cli/integration.go:80-92` — `integrationSessionID(flag)`: `--session` flag →
   `config.EnvClaudeCodeSessionID` (`CLAUDE_CODE_SESSION_ID`) → the SessionStart side-channel
   file under the resolved lock root. Acquire errors when no id resolves.
6. `internal/cli/integration.go:148-152` — status prints "held by a session that is gone
   (reclaimable)" when `lock.Stale()`.
7. Consumers of `ReadIntegrationLock` / `Stale()` (grep @ this tree, non-test): 
   `internal/cli/integration.go:131,138,148` (status verb) and
   `internal/hook/integration_lock_guard.go:74,95` (PreToolUse guard, gated by
   `workflow.integration_lock.enabled`, default false, fail-open on uncertainty). The guard's
   stale branch allows with an advisory suggesting `moai integration acquire` reclamation
   (guard:95-97).
8. Liveness probes: `internal/kanban/factory_alive_unix.go` (signal 0; EPERM = live) and
   `factory_alive_windows.go` (OpenProcess + GetExitCodeProcess, STILL_ACTIVE; indeterminate →
   live). Windows parity exists and must be preserved.
9. Session work registry: `internal/session/registry.go` — `Entry{SessionID, SpecID, Phase,
   StartedAt, LastHeartbeat, PID, Host, CWD}` at `.moai/state/active-sessions.json`
   (`DefaultRegistryPath`, registry.go:35-39); `DefaultStaleMinutes = 30` (registry.go:62-65);
   verbs register/heartbeat/deregister/list/purge under `internal/cli/session.go`. Registration
   is driven by session hooks — a plain shell lane has no entry. The recorded `PID` is
   `resolveSessionPID()` (registry.go:196-198 comment: "The session PID, NOT os.Getpid()").
10. `internal/session/session_pid.go` — `resolveSessionPID()` resolution order: (1)
    `MOAI_SESSION_PID` (`config.EnvMoaiSessionPID`, envkeys.go:271) when it names a live
    process; (2) nearest ancestor that is not a wrapper shell (`sh/bash/zsh/dash/ksh/fish/
    csh/tcsh/env/moai`, platform ancestry via `proc_info_{linux,bsd,other}.go`); (3)
    `os.Getpid()` fallback — a fallback this SPEC must NOT inherit (it is the defect again).
11. `internal/cli/launch_session_pid.go` — the launcher stamps `MOAI_SESSION_PID` via
    `withSessionPID` on POSIX (`execve(2)`, launcher pid becomes the session's). Doc constraint:
    "The stamp belongs to callers that KNOW the session PID. A hook must never set it." Windows
    launcher stamp propagation is unverified (see §G Risks / plan open questions).
12. Existing test harness: `internal/cli/integration_lock_cli_test.go` — `runIntegration`
    (lines 26-43) executes the cobra command **in-process** with `CLAUDE_PROJECT_DIR` +
    `GIT_CEILING_DIRECTORIES` pinned to a temp root. In-process execution cannot simulate "the
    acquire CLI process exits" — the red test requires a real child binary (built via
    `go build -o`, NOT `go run`: an intermediate `go` process is not a wrapper name the
    ancestry walk skips, and it exits early).
13. Docs that currently enshrine the defect as a workaround (exist on THIS branch; main lacks
    them): `.claude/rules/local/gitflow-lane-protocol.md` §3 blockquote (lines 42-43:
    "[HARD] 현재 이 락은 직렬화를 보장하지 못한다 — 카드 t298 …") and `CLAUDE.local.md`
    §4.1/§4.1.2 (git-flow 통합 레인 운영 절차, lead-notice-only serialization prose).

## §C Requirements (GEARS)

Requirements are implementation-neutral: they fix the observable liveness contract, not the
code shape. REQ numbering: REQ-INL-NNN.

- **REQ-INL-001** (Ubiquitous) — The release-integration lock shall anchor holder liveness to
  the holding **session's** process lifetime, not the lifetime of the CLI process that wrote
  the record.

- **REQ-INL-002** (Event-driven) — **When** `moai integration acquire` records a holder, the
  acquire verb shall resolve the calling session's owner pid through the established owner-pid
  chain — an explicit `MOAI_SESSION_PID` env stamp naming a live process, else the nearest
  non-wrapper ancestor process — and shall record that pid together with an additive
  owner-anchor marker in the lock record.

- **REQ-INL-003** (Event-detected) — **When** no owner pid can be resolved at acquire time, the
  acquire verb shall record pid 0 with the owner-anchor marker, and the holder shall therefore
  read LIVE until explicit release or a recorded `--force` takeover — never treated as stale.

- **REQ-INL-004** (State-driven) — **While** the recorded owner process is alive, `Stale()` shall
  report false, and every consumer of the lock (`moai integration status`, acquire refusal, and
  the PreToolUse guard) shall treat the window as held and non-reclaimable.

- **REQ-INL-005** (Event-driven) — **When** the recorded owner process has exited, `Stale()`
  shall report true and the window shall be reclaimable by a foreign acquirer without
  `--force`.

- **REQ-INL-006** (Capability gate / backward compatibility) — **Where** an on-disk record
  predates the owner anchor (no owner-anchor marker present), the lock shall evaluate it with
  the pre-existing probe semantics — a dead recorded pid reads reclaimable, exactly as before —
  and the record schema change shall be additive-only (a new optional field; no existing field
  renamed, retyped, or removed; existing records remain parseable).

- **REQ-INL-007** (Capability gate / platform parity) — **Where** the platform cannot report
  process ancestry (Windows), the acquire verb shall degrade to the REQ-INL-003 conservative
  pid-0 record rather than to the acquire process's own pid, and the existing
  `FactoryProcessAlive` Windows probe behavior shall be preserved unchanged.

- **REQ-INL-008** (Capability gate / guard consumer) — **Where** `workflow.integration_lock.enabled`
  is set, the PreToolUse guard shall deny a foreign `git merge` in the release worktree while
  the anchored holder is live, and shall allow with an advisory when the anchored holder is
  gone — fail-open on uncertainty, unchanged.

- **REQ-INL-009** (Event-driven / documentation) — **When** the fix lands, the two local
  documents that currently enshrine the defect as a workaround
  (`.claude/rules/local/gitflow-lane-protocol.md` §3 blockquote; `CLAUDE.local.md` §4.1/§4.1.2)
  shall be rewritten to describe the restored serialization guarantee.

- **REQ-INL-010** (Unwanted) — The acquire verb shall not silently overwrite another lane's
  recorded trace: every takeover (stale-reclaim or `--force`) shall continue reporting the
  displaced holder, and foreign release shall continue to be refused without `--force`.

## §D Constraints

- **Fail-direction constraint (binding, quoted from the code itself):** every new uncertainty
  path resolves toward TREAT-AS-LIVE. integration_lock.go:84-90: "the cost of a false 'stale'
  (two lanes merging at once) is the failure this lock exists to prevent, while the cost of a
  false 'live' is one operator asking the holder to release. The asymmetry is not close."
- **Heartbeat freshness must not decide staleness.** The session registry's 30-minute
  `DefaultStaleMinutes` purge window would make a quiet-but-alive holder read stale — the
  false-"stale" direction. Registry data may inform, never decide, `Stale()`.
- **Additive-only JSON schema** for `.moai/state/integration-lock.json`; no breaking change to
  records already on disk.
- **`MOAI_SESSION_PID` stamping discipline** (launch_session_pid.go): the stamp belongs to
  callers that know the session pid; the acquire CLI may READ it, but nothing in this SPEC adds
  a new writer of it.
- **Windows parity** preserved (`factory_alive_windows.go` untouched in behavior; the
  conservative pid-0 path is the Windows degradation, not `os.Getpid()`).
- **Test isolation [HARD]:** all tests isolate under `t.TempDir()` with `CLAUDE_PROJECT_DIR` +
  `GIT_CEILING_DIRECTORIES` pinned; no test touches the real primary-checkout
  `.moai/state/integration-lock.json` or `active-sessions.json`.
- **Local-only docs:** both doc targets are local-only files (no template mirror, no
  `make build` obligation).

## §E Acceptance Matrix (summary)

Canonical enumeration lives in `acceptance.md` (AC-INL-001..AC-INL-011), each criterion carrying
the two-cell discipline: a RED-now cell pinned to tree `d29b8942e` stating why it is red, and a
green-path cell naming the milestone that flips it. Highlights: AC-INL-001 is the required
red-first cross-process reproduction (child acquire exits; parent session alive; status must
read HELD — fails today); AC-INL-003 is its mutant-guard pair (owner genuinely gone →
reclaimable), which together exclude both the vacuous and the impossible direction.

## §F Dependencies and Related SPECs

- Predecessor: card t194 (integration lock mechanical layer) and t181 (announcement doctrine) —
  no SPEC ID carried forward; this SPEC repairs t194's liveness anchor.
- Reuses the owner-pid resolution discipline of `internal/session/session_pid.go` (session work
  registry lineage).
- No `depends_on` entries; no other SPEC in this worktree's `.moai/specs/` shares the ID space.

## §G Risks

- **Windows launcher stamp propagation unverified** — if the Windows launcher does not stamp
  `MOAI_SESSION_PID` with a session-lifetime pid, Windows acquires degrade to pid-0
  conservative (serialization preserved, reclaim requires `--force`). Acceptable per the
  asymmetry doctrine; documented, not hidden.
- **Ancestry walk finds the wrong long-lived process** when the acquire runs under an unknown
  wrapper (e.g. a version-manager shim). Mitigation: the wrapper-name set is the existing,
  tested one from session_pid.go; the `MOAI_SESSION_PID` env stamp short-circuits the walk
  wherever the launcher stamped it.
- **Legacy in-flight windows at upgrade** read reclaimable exactly as today (no wedge, no
  regression); a lane that acquired pre-fix must re-acquire post-upgrade — noted in the doc
  rewrite.

## §H History

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-08-27 | Initial draft (plan phase, card t298). Status set to `draft` on creation across all four plan-phase artifacts. |

## Exclusions

### Out of Scope — deny-layer enablement
- Flipping `workflow.integration_lock.enabled` to true in any config (the guard's opt-in
  posture is unchanged; this SPEC only fixes the liveness the guard consumes).

### Out of Scope — registry-authoritative staleness
- Making the session work registry (`active-sessions.json`) the decision input for `Stale()`,
  including heartbeat-freshness thresholds, registry purge triggers, or auto-deregistration of
  lock holders. Evaluated as design Option A and rejected (see plan.md §B).

### Out of Scope — release-worktree writability enforcement
- Any capability-boundary work (making the worktree unwritable, OS-level locks). The lock
  remains a coordination signal; `--force` remains the deliberate bypass.

### Out of Scope — launcher/hook changes
- New writers of `MOAI_SESSION_PID`, launcher rework, or SessionStart/SessionEnd hook changes.
  The owner-pid chain is consumed read-only at acquire time.
