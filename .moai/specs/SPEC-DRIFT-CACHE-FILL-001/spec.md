---
id: SPEC-DRIFT-CACHE-FILL-001
title: "Out-of-band drift cache fill for session start"
version: "0.4.0"
status: draft
created: 2026-09-18
updated: 2026-09-18
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: "internal/hook, internal/spec, internal/cli, internal/config, internal/template"
lifecycle: spec-anchored
tags: "session-start, drift-cache, latency, single-flight, advisory, card-t871"
tier: M
era: V3R6
depends_on: []
---

# SPEC-DRIFT-CACHE-FILL-001 — Out-of-band drift cache fill for session start

Delivering card: **t871**. Measurement provenance: card **t666 R3**.

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-18 | manager-spec | Initial plan-phase authoring — specifies the operator-adopted D3 (D1 + D2) decision for the never-filled drift cache |
| 0.4.0 | 2026-09-18 | manager-spec | plan-audit iter3 scoped revision (N1 completion + NEW-1/NEW-2): REQ-DCF-009 **re-amended** — the claim target moves from the record to a separate `<path>.lock`, and the reclaim is performed under it, closing the reclaim TOCTOU and the empty-record window; AC-DCF-006 clause (a) becomes cross-process (re-executed subprocesses, not goroutines) after the audit disproved the claim that clause (b) excluded the in-process-mutex mutant. Tier gate did not fire — count stays 16/16 |
| 0.3.0 | 2026-09-18 | manager-spec | plan-audit iter2 scoped revision (N1/N2/N3/N5): REQ-DCF-009 **amended** — the suppression write becomes an exclusive claim via `internal/atomicfile.Claim`, closing the lockless read-check-then-write race; REQ-DCF-013 regains "beyond the existing join bound"; TTL ≥ deadline + slack restored with both constants resolving from `internal/config/defaults.go`; the cache write reuses `internal/atomicfile.Replace`. Tier gate did not fire — no new requirement; count stays 16/16 |
| 0.2.0 | 2026-09-18 | manager-spec | plan-audit iter1 FAIL 0.55 revision: requirement set consolidated 22 → 16 (Tier M ceiling); suppression record moved to the parent (D2/D3); `internal/sessionmsg` adopted as the lock precedent (D4); config-field SSOT named (D5); torn-write requirement added (D9); `related_specs` frontmatter key dropped (D15) |

---

## §A Context and Motivation

### §A.1 The measured defect (card t666 R3 — carried, not re-derived)

`SessionStart`'s `Handle` spawns four deferred advisory scans in a background
goroutine and joins them with a 250 ms bound (`deferredScanJoinBound`,
`internal/hook/session_start.go`). The heaviest is the SPEC status-drift count
(`driftCountFn` = `spec.DriftCountCtx`, `internal/spec/drift.go`). Drift results
are cached keyed on the HEAD SHA at `.moai/state/drift-cache.json`
(`internal/spec/drift_cache.go`), and the save happens only **after** the
compute finishes.

Measured on a real `moai hook session-start` process (a `git clone --shared`
copy of this repository, 869 SPEC directories, darwin, load average 8-15):

| Observation | Result |
|---|---|
| exit arm (hook run repeatedly, cache cold) | 15 of 15 runs left the cache **ABSENT** |
| control arm (`moai spec drift` run to completion first) | 5 of 5 runs `present(head-match)`; the following hook run saw the hit |
| GIT_TRACE, cold hook run | `rev-parse HEAD` (cache-key lookup) at 262 ms; process **exits at 489 ms**, before the drift compute reaches its own git work |
| GIT_TRACE, slower run | `git log` started at 902 ms, cut 24 ms later by process exit |
| in-process decomposition | cold drift p50 **967 ms**; warm drift p50 **32 ms**; bounded join p50 **251 ms** on every cold run, delivering nothing; Handle p50 ~**390 ms** cold vs ~**176 ms** warm |

### §A.2 The mechanism

The compute is strictly longer than the join bound, and the process is a
short-lived CLI: when `Handle` returns, the runtime tears the goroutine down
wherever it happens to be — which is always before `saveDriftCache`. The cache
therefore has no writer on this path at all. Two consequences compound:

1. Every cold session pays the full 250 ms join and discards the result.
2. `status_drift_warning` is never delivered after a HEAD change — until a human
   runs `moai spec drift` by hand, and it regresses at the next commit.

The bounded join was designed as drop-mitigation for a scan that *usually*
finishes inside the bound. For drift it never does, so the bound is a pure cost.

### §A.3 The adopted decision (D3 — operator, 2026-09-18; not re-opened here)

- **D1** — on a cache miss, do not wait for the drift scan at all.
- **D2** — on a cache miss, fill the cache out of band so the NEXT session gets a
  hit and the drift advisory returns.

This SPEC specifies the mechanism, requirements and acceptance for D1 + D2. The
choice among alternatives is settled and is recorded as accepted input.

### §A.4 Existing capability this reuses (simplicity ladder, before new code)

| Need | Existing capability |
|---|---|
| detached self-respawn, fire-and-forget | `internal/statusline/landed.go` `maybeRefreshLandedCounts` — `os.Executable()` + `isSelfInvocable` basename guard + nil streams + `Process.Release()` |
| non-blocking exclusive cross-process lock, **both platforms** | `internal/sessionmsg/lock_unix.go` (`unix.Flock`, `LOCK_EX\|LOCK_NB`) paired with `internal/sessionmsg/lock_windows.go` (`LockFileEx`, `LOCKFILE_EXCLUSIVE_LOCK\|LOCKFILE_FAIL_IMMEDIATELY`, documented in its own header as "parity with POSIX flock LOCK_EX \| LOCK_NB"). The pair is itself copied from `internal/session`; this SPEC copies it a third time rather than writing a variant |
| an exclusive "am I the one who proceeds" claim | `internal/atomicfile.Claim` — `os.O_CREATE\|os.O_EXCL\|os.O_WRONLY`, atomic on POSIX, `CREATE_NEW` on Windows, and deliberately never retried. `internal/verify/claim_lock.go` is the in-tree precedent for composing it with a second layer and for staleness reclaim |
| an atomic file replace | `internal/atomicfile.Replace` — the counterpart of `Claim`; the two are explicitly not substitutable ("make newpath be this content" vs "am I the one who gets to proceed") |
| self-bounding compute | `spec.DriftCountCtx` — already races the work against `ctx.Done()` |
| async/inline test seam | `deferredScansAsync` / `asyncDeferredScans()` / `registerDeferredScanSeam` |
| spawn-attempt test seam | `landedSpawnProbe` / `githubSpawnProbe` placement idiom |
| typed config key with a Go-side default | `internal/config/types.go` (`BranchGuard BranchGuardConfig \`yaml:"branch_guard"\``) + `internal/config/defaults.go` — the YAML files are overrides, never the SSOT |
| pull-based, never-awaited, persisted-for-a-later-activation refresh | `internal/hook/session_start_guard_liveness.go` — the same shape, one layer over |

Nothing here needs a new lock primitive, a new spawn idiom, or a new config
mechanism. The genuinely new surface is the child entry point and the
suppression record.

### §A.5 Tier judgment (Tier M, at ceiling)

16 requirements and 16 acceptance criteria — exactly the Tier M ceilings
(`spec-workflow.md` § SPEC Complexity Tier), reached by consolidation rather
than renumbering: the detach contract folded into REQ-DCF-001, the cache-resolve
pair into REQ-DCF-011, the fail-open pair into REQ-DCF-013, and the spawn-probe
seam moved to `plan.md` where an implementation affordance belongs. Tier stays M
on scope: the change touches roughly 10 files and well under 1000 LOC, so tiering
up to L would buy a higher audit threshold and two more artifacts without adding
any decision that `design.md` would hold.

---

## §B Requirements (GEARS)

`<subject>` is generalized. Named subjects: the **session-start handler** (the
parent), the **fill process** (the child), the **drift cache**, the **probe**.

### §B.1 M1 — Fill mechanism (out-of-band writer)

- **REQ-DCF-001** (Event-detected): **When** the session-start handler observes a
  drift-cache miss for the current HEAD, the handler **shall** start exactly one
  **detached** fill process and return without awaiting it. Detached means the
  child inherits none of the parent's stdin, stdout or stderr; the parent
  releases the process handle, never waits on it, and never reads its exit
  status.
- **REQ-DCF-002** (Ubiquitous): The fill process **shall** recompute drift for
  the current HEAD and persist the result to the HEAD-SHA-keyed drift cache.
- **REQ-DCF-003** (Ubiquitous): The fill process **shall** bound its own runtime
  by a deadline carried on its invocation and **shall** exit at that deadline
  whether or not the computation finished, so no fill outlives its bound.
- **REQ-DCF-004** (Where): **Where** the running executable's basename is not a
  `moai` binary, the handler **shall not** start a fill process.
- **REQ-DCF-005** (unwanted): The fill process **shall not** leave a
  partially-written drift cache file, whatever point its deadline interrupts it
  at.

### §B.2 M2 — Coordination: single-flight and burst suppression

- **REQ-DCF-006** (Ubiquitous): At most one fill process **shall** compute drift
  for a given project at a time.
- **REQ-DCF-007** (Event-driven): **When** a fill process starts, it **shall**
  attempt a non-blocking exclusive lock; **when** the attempt does not succeed it
  **shall** exit 0 immediately without computing.
- **REQ-DCF-008** (Ubiquitous): The lock **shall** be held by an operating-system
  handle whose release is performed by process exit — including abnormal
  termination — so no terminated fill can wedge later fills. The requirement is
  platform-neutral: it binds the POSIX and the Windows implementation equally.
- **REQ-DCF-009** (Event-driven) *(amended v0.3.0 — the write became a claim;
  re-amended v0.4.0 — the claim target is a separate lock, and the reclaim is
  performed under it)*: **When** the session-start handler decides to start a
  fill, the handler **shall** first acquire the suppression **lock** — a
  companion path beside the suppression record, acquired by **exclusive
  creation**, an operation at most one caller on the machine can win, which
  fails rather than overwrites when the path already exists, and which is never
  retried — and **shall**, holding that lock, read the record, judge it, write
  it, and release the lock; and **shall** start the child only on having written
  the record under the lock. **While** a record names the current HEAD, is not
  older than the suppression TTL, and is not stamped in the future, the handler
  **shall** start no further fill process.

  Two properties of the critical section are stated rather than assumed:

  - **Reclaim is concurrency-safe by containment.** Every read of the record,
    every judgement of its expiry, and every removal or rewrite happens **inside
    the lock**, so no handler can act on a record another handler has since
    replaced. The expiry judgement is made at the moment of removal, never
    carried over from a read taken before the lock was held.
  - **The record is never observed empty or half-written.** The lock carries no
    record content; the record itself is written whole. A record that is
    nonetheless unreadable — empty, truncated, or unparseable — is treated as
    **expired**, one stated disposition rather than two divergent readings.

  A record failing any of the three tests above is reclaimable: the handler,
  still holding the lock, removes it and writes its own. A handler that fails to
  acquire the lock starts no fill. A lock older than its own staleness bound is
  itself reclaimable — judged, like the record, at the moment of removal.

  > Amendment rationale (v0.3.0): v0.2.0 said "write … before starting the
  > child", which is a read-check-then-write and therefore lockless. The
  > ordering that matters is not *first write before any child* but *first write
  > before other reads* — N handlers released simultaneously all complete their
  > reads before the first write lands, and all spawn.
  >
  > Re-amendment rationale (v0.4.0): claiming the **record** left two holes the
  > iteration-3 audit found. (NEW-1, reclaim TOCTOU) removing an expired record
  > can destroy another handler's fresh claim — A reclaims and wins, B judging
  > on a pre-A read removes A's record and also wins. (NEW-2) an exclusive
  > create writes no content, so the record would exist **empty** for a window
  > and the disposition of an unparseable record was unspecified, diverging into
  > either a double spawn or a contradiction. Both close by claiming a separate
  > lock and doing the record work under it — the shape
  > `internal/verify/claim_lock.go` already uses. Still an amendment: same
  > actor, same record, same obligation; only the claim target and the placement
  > of the reclaim change.
- **REQ-DCF-010** (Event-detected): **When** a fill fails at any stage —
  including before the child executes at all — the handler-written record
  **shall** still bound respawn to at most one attempt per TTL per HEAD, so a
  persistently broken child cannot produce a per-session spawn loop.

### §B.3 M3 — Miss-path decoupling (D1)

- **REQ-DCF-011** (Event-driven): **When** the deferred advisory step runs, the
  session-start handler **shall** resolve the drift advisory from the HEAD-SHA
  cache alone — rendering `status_drift_warning` from an entry keyed to the
  current HEAD — and **shall not** compute drift in-band.
- **REQ-DCF-012** (Event-detected): **When** the cache lookup misses, the handler
  **shall** omit `status_drift_warning` for that session and **shall** spend no
  part of the join bound on drift computation.
- **REQ-DCF-013** (unwanted / Event-detected): No fill-related step **shall**
  block the handler, fail the hook, or extend the handler's return latency
  **beyond the existing join bound**; and
  **when** any fill step errors — executable resolution, record write, directory
  creation, lock open, or process start — the handler **shall** log at debug
  level and return its normal output.

### §B.4 M4 — Opt-out and test isolation

- **REQ-DCF-014** (Where): **Where** the resolved `workflow.drift_cache_fill`
  setting is disabled, the handler **shall** start no fill process; and **where**
  no project configuration declares the key, the setting **shall** resolve to
  enabled.
- **REQ-DCF-015** (State-driven): **While** the deferred-scan async seam is off —
  the test binary's `deferredScansAsync=false`, or a handler carrying
  `WithSynchronousDeferredScans` — the handler **shall** start no fill process.

### §B.5 M5 — An instrument that carries a verdict

- **REQ-DCF-016** (Ubiquitous): The drift-cache probe **shall** return an
  explicit verdict: failing when any run in its swept set left the cache absent,
  and distinguishing an empty swept set from a pass.

---

## §C Exclusions

### Out of Scope — drift computation itself

- No change to the drift algorithm, its git-log pass, or its record semantics.
- No change to the cache key: it stays the HEAD SHA. Frontmatter-keyed or
  content-hashed invalidation is a separate concern.
- No change to `moai spec drift`'s default behaviour, output, or `--no-cache`.

### Out of Scope — session-start scope beyond drift

- The other three deferred advisory scans (telemetry prune, stale-memory wrap,
  pending-proposal summary) keep the existing bounded join unchanged.
- No change to `deferredScanJoinBound`, to the hook's `timeout`, or to the
  `SessionStart` wiring in `.claude/settings.json`.

### Out of Scope — process and platform ambitions

- No long-lived daemon, no scheduler, no cross-machine coordination, no
  cadence-driven refresh.
- No warm-up of the cache at install, update, or commit time.
- No retry-with-backoff state machine beyond the single TTL-bounded record.

### Out of Scope — surfacing changes

- No change to how `status_drift_warning` renders, or to the `Data` map's
  `json:"-"` internal-only contract.
- No user-facing notification that a fill is running or has failed.

---

## §D Assumptions and accepted costs

1. The drift cache's only reader of consequence on this path is the deferred
   advisory step.
2. A one-session delay in the drift advisory after a HEAD change is acceptable —
   strictly better than today, where the advisory never arrives.
3. `os.Executable()` resolves to the installed `moai` binary in the hook's
   runtime, as it already does for the statusline refresh children.
4. **Shipping enabled is not neutral, and that is the accepted cost.** The
   `workflow.*` guard family defaults to `false` because those features ship
   *inert* — `internal/config/types.go` grounds the family default on
   **neutrality**, not on "adds a deny". This feature is not inert when enabled:
   on every user's machine, every cache miss starts an unsolicited child process.
   The cost is accepted because the child is short-lived, self-bounded,
   single-flight, TTL-suppressed, and silent, and because a default-off setting
   would leave the measured defect in place for every user who never reads the
   config. Users who object set `workflow.drift_cache_fill.enabled: false`
   (REQ-DCF-014). This is a decision to challenge at Kickoff, not an oversight.
5. A torn cache write is prevented rather than tolerated (REQ-DCF-005), even
   though `loadDriftCache` already fails open on unparseable JSON — because the
   deadline is engineered to fire near completion, which is exactly when a
   non-atomic write is most likely in flight.

---

## §E Cross-References

- `internal/hook/session_start.go` — the bounded join and the deferred seam.
- `internal/spec/drift_cache.go`, `internal/spec/drift.go` — the cache and the
  self-bounding `DriftCountCtx`.
- `internal/statusline/landed.go` — the detached-child idiom this reuses.
- `internal/sessionmsg/lock_unix.go`, `internal/sessionmsg/lock_windows.go` —
  the non-blocking exclusive lock pair this reuses on both platforms.
- `internal/config/types.go`, `internal/config/defaults.go` — where the opt-out
  key's typed field and its default actually live.
- `.github/workflows/template-neutrality-check.yaml` — the CI guard the
  template-side edit crosses.
- Related SPECs (moved here from frontmatter per the schema's optional-field
  list): SPEC-SESSIONSTART-PERF-001, SPEC-DRIFT-001,
  SPEC-HOOK-SESSIONSTART-PROBE-001, SPEC-GUARD-LIVENESS-001,
  SPEC-STATUSLINE-PROFILE-RESPECT-001.
- `.claude/rules/moai/development/verification-completeness.md` — the discipline
  `acceptance.md` is authored against.

🗿 MoAI
