---
id: SPEC-INTEGRATION-LOCK-ATOMIC-001
title: "Release-integration lock: serialize the record's read-modify-write across processes (card t336)"
version: "0.1.1"
status: draft
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "integration-lock, atomicity, kanban, factory, t336"
tier: M
---

# SPEC-INTEGRATION-LOCK-ATOMIC-001 — Integration Lock Mutation Atomicity

## §A Context and Problem

The release-integration window lock (`internal/kanban/integration_lock.go`, card t194) is the
mechanical layer under the lane-serialization doctrine
(`.claude/rules/moai/workflow/kanban-dispatch.md` § "Integration into the release branch is
self-served"). Its package header states the discipline it intends to follow (L18-20, measured
at tree `15453140a`):

> "So the window is a RECORD whose validity is decided by the recorded holder's liveness, and
> the flock discipline is borrowed only to serialize mutations of that record."

That second clause is not backed by the code. No exclusion primitive is taken anywhere in the
file: `grep -n 'flock\|Flock' internal/kanban/integration_lock.go` at `15453140a` returns only
the prose lines of that header. The mutation path is an unserialized read-modify-write:

| Step | Line @ `15453140a` | Code |
|---|---|---|
| READ | `integration_lock.go:186` | `current, err := ReadIntegrationLock(projectRoot)` |
| DECIDE | `integration_lock.go:190-200` | held / stale / `force` branch |
| WRITE | `integration_lock.go:205` | `writeIntegrationLock(path, &want)` |

Nothing holds an exclusive lock across L186..L205. `ReleaseIntegrationLock` has the same shape
(read `:216` → holder check `:220-225` → `os.Remove` `:226`). `writeIntegrationLock:257`
additionally stages through a FIXED shared path (`path + ".tmp"`).

**Hypothesis, not yet a measurement.** The pre-flight for this card
(`.moai/reports/t336/preflight.md` §1, §6) states explicitly that **no concurrent acquire was
executed** on this tree. The double-hold above is read off the code path — two processes both
observing a free (or the same stale) record, both passing the decision, both writing, last write
wins — and it remains a **code-path hypothesis until run-phase RED observes two live holders**.
This SPEC is written on that basis: the RED milestone converts the hypothesis into an
observation, and nothing in this document may be cited as a measured defect before it does.

The consequence, if the hypothesis holds, is the exact failure the lock exists to prevent: two
lanes are each told the window is theirs, and both enter the shared release worktree. The
predecessor SPEC's own fail-direction constraint names this as the one hazard that does NOT
resolve toward the cheap direction.

### Ancestry — the residue of SPEC-INTEGRATION-LOCK-LIVENESS-001 (card t298)

t298 repaired the **identity of record** (which pid is written) and drew the boundary in its own
words (t298 `spec.md` §A):

> "**Atomicity** — *how* the record is written — is the separate, unserialized
> read-modify-write in `AcquireIntegrationLock` (§G, and the iter-1 audit's D4), which this card
> does not close."

t298 §G carries it as a **RETAINED residual hazard** and closes with: "Serializing the record
mutation is explicitly OUT of scope for this card … it is a separate card, and widening this one
to cover it is the named anti-pattern." That separate card is this one. t298 also recorded why
the residue became urgent rather than academic: before its fix the defect was masked (every
record read reclaimable, so nothing was serialized anyway); after its fix "serialization becomes
real everywhere EXCEPT this window."

This SPEC therefore re-litigates none of t298's settled decisions. The liveness anchor, the
`PIDSourceSessionOwner` marker, the pid-0 conservative path, and the TREAT-AS-LIVE asymmetry are
inherited unchanged and are constraints here (§D), not open questions.

## §B Verified Premises (measured at tree `15453140a`, branch `WT-integration-lock-atomic`)

Every item below is attributed to this tree. Re-measure before citing elsewhere.

1. `internal/kanban/integration_lock.go:177-209` — `AcquireIntegrationLock`; the read/decide/write
   triple at `:186` / `:190-200` / `:205` with no exclusion between them.
2. `internal/kanban/integration_lock.go:215-230` — `ReleaseIntegrationLock`; read `:216`, holder
   check `:220-225`, `os.Remove` `:226`.
3. `internal/kanban/integration_lock.go:257` — `tmp := path + ".tmp"`, a fixed staging path shared
   by every concurrent writer.
4. `internal/kanban/integration_lock.go:13-26` — the package header's lifetime distinction: the
   board lock "spans one process's read-modify-write and is an flock, so it dies with the process
   that took it", whereas "an integration window spans many CLI invocations, many turns, and
   minutes of human-paced work — an fd cannot represent it." The two lifetimes are distinct and
   both are preserved by this SPEC (§D).
5. Substrate available in this package: `board_lock.go` (`AcquireBoardLock`, `BoardLock`,
   `boardLockImpl`, `newLockOwnerRecord`), `board_lock_unix.go:36` (`acquireBoardLockImpl` —
   flock, non-blocking, already path-parameterized), `board_lock_windows.go:55`
   (`acquireBoardLockImpl` — `O_CREATE|O_EXCL` with a bounded transient retry),
   `board_lock_clear_windows.go:75` (`ClearStaleBoardLock` — dead-owner clear with a pre-removal
   re-read; keyed to `boardLockPath(root)`), `board_lock_clear_unix.go:23` (gated-out no-op).
6. Existing bounded-contention policy in this package: `board_store.go:114-151`
   (`boardLockWaitBudget`, `boardLockWaitMin/Max/Step`, `boardLockRetryWait`) and
   `board_store.go:156` (`acquireBoardLockSerialized`).
7. Cross-process test harness in this package: `kanban_helper_test.go:22-113`
   (`TestKanbanHelperProcess`, dispatch on `MOAI_KANBAN_HELPER`) and
   `board_lock_cross_test.go:43-104` (two real child processes, exactly-one-succeeds, plus a
   positive control at `:84-104` proving the refusal is conditional on the invariant rather than
   on concurrency itself).
8. Non-test consumers of the record (grep at this tree): `internal/cli/integration.go:131`
   (status read), `:185` (acquire), `:234` (release), and
   `internal/hook/integration_lock_guard.go:74` (PreToolUse guard, read-only).
9. Layer split: the DENY layer (`internal/hook/integration_lock_guard.go`) is gated by
   `Workflow.IntegrationLock.Enabled` (distributed default `false`); the RECORD layer
   (`internal/kanban/integration_lock.go`) runs regardless. The defect is in the RECORD layer.

## §C Requirements (GEARS)

Requirements are implementation-neutral: they fix the observable mutation contract, not the code
shape. Numbering: REQ-ILA-NNN.

- **REQ-ILA-001** (Ubiquitous) — The integration-lock record's read-modify-write shall be
  serialized across OS processes: at most one process at a time shall be inside the
  read → decide → write sequence for a given project root.

- **REQ-ILA-002** (Event-driven) — **When** two sessions concurrently acquire the same free or
  stale window, exactly one shall be recorded as holder and the other shall be refused with
  `ErrIntegrationLockHeld`; the record shall never name a holder while a second caller has also
  been told it holds the window.

- **REQ-ILA-003** (Event-driven) — **When** a caller's mutation is serialized behind another
  caller's, it shall re-read the record after entering the critical section, so its held / stale
  / `force` decision is made against the state the previous mutation published, never against a
  read taken before the wait.

- **REQ-ILA-004** (State-driven) — **While** the mutation lock is contended, a contender shall
  retry within a bounded elapsed budget and shall surface a contention error that is
  DISTINGUISHABLE from `ErrIntegrationLockHeld` — a transient mutation-lock timeout is not a
  statement that another session owns the integration window, and reporting it as one would tell
  a lane the wrong thing about the board.

- **REQ-ILA-005** (Ubiquitous) — Single-threaded acquire and release semantics shall be
  unchanged. Re-acquire by the recorded holder shall still succeed and refresh the record; a
  stale record shall still be taken over and reported in `replaced`; `--force` shall still take
  over a live holder and report it; foreign release shall still be refused without `--force`;
  `ErrIntegrationLockNotHeld` shall still be returned for an empty release. This SPEC changes
  WHEN mutations interleave, never WHAT a given single-threaded call decides.

  **Amendment (deterministic-interleaving carve-out).** One addition to the mutation path is
  permitted and is NOT a semantic change: a package-level, **nil-by-default** test-only
  interleaving hook invoked once between the decision and the write
  (`integration_lock.go:204`), set only by the cross-process test child under an env flag. With
  the hook nil — every production path, every consumer in §B item 8 — behavior is
  byte-for-byte what this requirement states. The carve-out exists because the criteria this
  card lives on must be observable by construction rather than by luck (§G risk 1); it grants
  nothing else, and no production caller may set it.

- **REQ-ILA-006** (Capability gate / platform parity) — **Where** the platform substrate is
  atomic-create rather than flock (Windows), a mutation-lock holder that is killed mid-mutation
  shall not wedge the record permanently: a contender that exhausts its bounded budget shall
  clear the artifact only when the recorded owner is positively observed absent, re-reading the
  recorded identity immediately before removal and aborting on any mismatch, and shall then
  retry. On Unix the kernel releases flock on process exit, so no clear path is required and none
  is added.

- **REQ-ILA-007** (Ubiquitous) — The mutation lock's lifetime shall be one CLI invocation's
  critical section, distinct from the integration WINDOW's lifetime, which spans many
  invocations. The window shall remain a RECORD whose validity is decided by the recorded
  holder's liveness; the mutation lock shall never be the thing that represents the window.
  The mutation artifact shall NOT share the record's filename stem (`integration-lock`), so that
  a future reader cannot fold the two lifetimes back together by globbing
  `.moai/state/integration-lock*`; and discovering the record by glob rather than by its named
  path shall remain forbidden.

- **REQ-ILA-008** (Unwanted) — The mutation lock shall not be gated by
  `Workflow.IntegrationLock.Enabled` or by any other configuration flag. That flag gates the
  PreToolUse DENY layer only; the record layer runs in every project, and a double-hold corrupts
  coordination whether or not the deny is enabled.

- **REQ-ILA-009** (Unwanted) — The repair shall not make the release worktree unwritable, shall
  not remove `--force`, and shall not convert the lock into a capability boundary. It remains a
  coordination signal (`integration_lock.go:22-26`).

- **REQ-ILA-010** (Event-driven / secondary, defense-in-depth) — **When** the record is written,
  the staging path shall be unique to the writing call rather than a fixed shared name, so no two
  writers can share one staging file. Stated honestly: under a correct mutation lock (REQ-ILA-001)
  two concurrent writers cannot reach this path at all, so this is defense-in-depth against a
  future caller that writes outside the lock — not an independently observable defect, and no
  criterion in `acceptance.md` claims to observe a torn record.

- **REQ-ILA-011** (Event-driven / documentation) — **When** the repair lands, the package header's
  claim at `integration_lock.go:18-20` ("the flock discipline is borrowed only to serialize
  mutations of that record") shall be true of the code, either because the mechanism now exists or
  because the sentence is rewritten to describe what does exist. A header that describes an absent
  mechanism is the condition this card found.

## §D Constraints

- **Inherited fail-direction constraint (binding, from t298 §D and `integration_lock.go:99-112`):**
  every uncertainty path resolves toward TREAT-AS-LIVE. The double-hold this SPEC closes is the
  one hazard that fails the other way, which is why it is being closed.
- **Layer boundary [HARD]:** `internal/hook/**` and `internal/config/**` are OUT OF SCOPE and MUST
  NOT be modified. The repair is confined to the record layer.
- **Card boundary vs t320 [HARD]:** `moai integration release`'s error classification and message
  text are card t320's concern. If release's read-modify-write goes under the same mutation lock,
  its error sentinels and its user-facing strings stay byte-identical, so t320 lands on top
  unchanged.
- **Two lifetimes must not be conflated:** the mutation lock is short-lived (one CLI invocation);
  the window record is long-lived (many invocations, human-paced). The mutation lock is never
  persisted as the window and never consulted to decide who holds it.
- **Verification scope [HARD]:** `go test ./internal/kanban/...` (plus `./internal/cli/...` only if
  a CLI-level test is touched). A local full-suite run (`go test ./...`) is PROHIBITED in this
  repository; CI is the full-suite judge.
- **Cleanup obligation [HARD]:** any criterion that spawns child processes bounds them by kills
  registered with `t.Cleanup`, or by an external `timeout` / `exec.CommandContext` deadline. A
  trailing `kill` at the end of a test body is NOT acceptable: every early-return path skips it.
- **Test isolation [HARD]:** all tests isolate under `t.TempDir()`; no test touches the real
  primary checkout's `.moai/state/integration-lock.json`.
- **Windows parity:** `factory_alive_windows.go` behavior is untouched; a windows-tagged file that
  is added or changed obliges `GOOS=windows go vet ./internal/kanban/...` (a plain build does not
  compile `_test.go` files).

## §E Acceptance Matrix (summary)

Canonical enumeration lives in `acceptance.md` (AC-ILA-001..AC-ILA-009). The shape the lead
required is carried by the first three rows: AC-ILA-001 is the RED direction (two live holders
observed with the critical section disabled, and **attributed** — neither caller performed a
takeover — so a stale-reclaim cannot be mistaken for the race), AC-ILA-002 is the GREEN direction
(exactly one holder, the other refused with the named sentinel), and AC-ILA-003 is the positive
control (a non-conflicting concurrent pair that must BOTH succeed), which is what distinguishes
the repair from the lock refusing everything. AC-ILA-006 is the mutation guard that keeps
AC-ILA-002 from going vacuous later. AC-ILA-001, AC-ILA-002 and AC-ILA-006 all run ONE shipped
test function, differing only in whether the critical section is disabled — there is no
second, phantom test name.

## §F Dependencies and Related SPECs

- Predecessor: `SPEC-INTEGRATION-LOCK-LIVENESS-001` (card t298) — this SPEC is the residue that
  SPEC explicitly scoped out (its §G RETAINED hazard), re-measured as still open by an independent
  sync-auditor and carried into card t336.
- Adjacent, non-conflicting: card t320 (`moai integration release` empty-message defect), which
  touches the release OUTPUT layer. No `depends_on` is declared in either direction.
- Reuses the substrate of `SPEC-KANBAN-BOARD-001` (`board_lock*.go`) without changing its
  behavior.

## §G Risks

- **The RED observation must be deterministic, not lucky (resolved, not merely noted).** The window
  at `integration_lock.go:186..205` is one `os.ReadFile`, a branch, one `os.WriteFile` + `os.Rename`
  — tens of microseconds on a warm filesystem. A barrier-released pair hits it a few percent of the
  time at best, so a criterion that waits for the interleaving has no stop rule: widening until RED
  appears is tuning until luck arrives, and stopping early proceeds with no observation. This SPEC
  therefore does NOT rely on a probabilistic pair. The deterministic-interleaving hook permitted by
  REQ-ILA-005's amendment holds one child between its decision and its write while the second child
  runs its whole acquire, so the double-hold is produced **by construction** in a single round.
  Two consequences are owned rather than hidden:
  - The hook is an edit to the *unrepaired* path, so AC-ILA-001 is measured at HEAD with the
    critical section disabled — NOT at `15453140a`, where neither the hook nor the test exists.
  - The stalled child must be released on a bounded timeout, because under the repair the second
    child blocks on the mutation lock and would otherwise deadlock against the first. That timeout
    is a liveness bound, not a race window, and it MUST be shorter than the mutation-lock wait
    budget so the second child's outcome is a held-refusal rather than a busy-timeout.
- **Zero-observation handling (binds AC-ILA-001 AND AC-ILA-006).** With the hook in place a
  zero-observation run means the mechanism is wrong, not that the race is absent. Either criterion
  observing zero double-holds in **3 consecutive attempts** is a **blocker**: the run phase stops
  and escalates to the orchestrator with the attempt outputs, and MUST NOT proceed to M2 (for
  AC-ILA-001) or record AC-ILA-006 as satisfied. No widen-and-retry loop is authorized, and absence
  of the observation is never read as evidence the race does not exist.
- **A retry budget can mask contention as success.** REQ-ILA-004 requires the transient mutation
  timeout to be distinguishable from `ErrIntegrationLockHeld`; if the two are conflated, a lane
  hitting a slow peer would be told another session holds the window, which is false.
- **Windows clear reuse widens a shared helper.** REQ-ILA-006's clear is the same shape as
  `ClearStaleBoardLock`, which is currently keyed to `boardLockPath(root)`. Making it
  path-parameterized touches a tested board path. **The mitigation must not be misstated:**
  `board_lock_clear_windows.go` is `//go:build windows` (line 1), so on a darwin lane machine no
  test compiling that file runs at all — neither AC-ILA-005 nor `go test ./internal/kanban/...`
  can observe a regression in the code M3 edits, and citing them here would be a mitigation that
  cannot fire. What actually holds: **compilation** is checked locally by
  `GOOS=windows go vet ./internal/kanban/...`, and **behavioral non-regression** of
  `ClearStaleBoardLock` is judged by CI's windows job, which runs the board's existing windows
  clear criteria (`board_lock_clear_windows_test.go`) alongside the new ones — see AC-ILA-007(b).
  The blast radius is bounded in the same breath: the file is windows-tagged, so Unix behavior
  cannot change.
- **TOCTOU residual, inherited unchanged:** the pre-removal re-read NARROWS the clear's
  inspection-to-removal window; it does not close it (the same AP-29 residual
  `board_lock_clear_windows.go:66-74` states). No new engineering is attempted here.
- **PID reuse** in the clear's liveness probe is inherited from the existing design and fails
  toward TREAT-AS-LIVE (the mutation lock is retried, not cleared).

## §H History

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-08-29 | Initial draft (plan phase, card t336). Status set to `draft` on creation across all plan-phase artifacts. Race stated as a code-path hypothesis pending run-phase RED, per `.moai/reports/t336/preflight.md` §6. |
| 0.1.1 | 2026-08-29 | plan-audit iteration-1 repair (FAIL 0.803, six blocking findings). D1: the M1 helper's recorded PID fixed (never `os.Getpid()`) and a stale-takeover discriminator added to AC-ILA-001/002. D2: deterministic-interleaving hook adopted — REQ-ILA-005 amended with a nil-by-default carve-out, §G risk 1 rewritten, zero-observation clause given a 3-attempt ceiling and extended to AC-ILA-006. D3+D4: one shipped test function; AC-ILA-001 restated at HEAD-with-section-disabled with a positive pass condition. D5: `plan.md` §B item 1 corrected (`acquireBoardLockSerialized`, 1.65 s budget) and item 3 promoted to the deciding argument. D6: §G risk 3 mitigation restated as `GOOS=windows go vet` (compilation) + CI windows job (behavior). D7-D9 applied: pgrep self-match excluded, AC-ILA-001 Claim mood hedged, mutation artifact renamed off the record's stem. |

## Exclusions

This SPEC's exclusions are stated so the run phase cannot widen into them.

### Out of Scope — the deny layer and its flag
- `internal/hook/integration_lock_guard.go`, `internal/hook/pre_tool.go`, and every file under
  `internal/config/**` are untouched.
- Flipping `Workflow.IntegrationLock.Enabled`, or making the record layer conditional on it.

### Out of Scope — `moai integration release` output (card t320)
- The release verb's error classification, its empty-message defect, and its user-facing strings.
  If release's mutation goes under the lock, its sentinels and strings stay byte-identical.

### Out of Scope — holder-liveness semantics (settled by t298)
- `Stale()`, the `PIDSourceSessionOwner` marker, the pid-0 conservative path, owner-pid
  resolution, and the TREAT-AS-LIVE asymmetry. All inherited unchanged.
- Registry-authoritative staleness and heartbeat-freshness thresholds (t298 evaluated and
  rejected these).

### Out of Scope — capability enforcement
- Making the release worktree unwritable, OS-level locks on the worktree, or removing `--force`.
  The lock remains a coordination signal.

### Out of Scope — multi-host coordination
- The record carries no host field (t298 §G). A pid probed on another host stays meaningless; not
  addressed here.

### Out of Scope — doctrine and local documents
- `.claude/rules/moai/workflow/kanban-dispatch.md`, `CLAUDE.local.md`, and
  `.claude/rules/local/gitflow-lane-protocol.md` are not rewritten by this card. Only the package
  header's own false claim (REQ-ILA-011) is in scope.
