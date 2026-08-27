# SPEC-INTEGRATION-LOCK-LIVENESS-001 — Acceptance Criteria (card t298)

Baseline tree for every RED-now cell: `d29b8942e` @ `WT-integration-lock` (worktree t298).
Two-cell discipline per `.claude/rules/moai/development/verification-completeness.md` §2: each
criterion carries a RED-now cell stating why it is red (or, for preserved-invariant criteria,
its mutant-guard pairing) and a green-path cell naming the flipping milestone. All tests run
under `t.TempDir()` with `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES` pinned; the real
primary-checkout state files are never touched.

## §D AC Matrix

| AC | Requirement | Kind | Flipping milestone |
|---|---|---|---|
| AC-INL-001 | REQ-INL-001, 002, 004 | RED-first core | M2 |
| AC-INL-002 | REQ-INL-002 | RED-first (env-stamp path) | M2 |
| AC-INL-003 | REQ-INL-005 | Mutant-guard pair | M2 |
| AC-INL-004 | REQ-INL-003 | Conservative invariant | M2 |
| AC-INL-005 | REQ-INL-010 | Preserved invariant | M3 |
| AC-INL-006 | REQ-INL-010 | Preserved invariant | M3 |
| AC-INL-007 | REQ-INL-006 | Backward compat | M3 |
| AC-INL-008 | REQ-INL-004, 008 | Guard consumer | M4 |
| AC-INL-009 | REQ-INL-007 | Platform parity | M4 |
| AC-INL-010 | REQ-INL-009 | Documentation | M5 |
| AC-INL-011 | REQ-INL-009 | Documentation | M5 |

## §D.1 Criteria (Given-When-Then, two cells each)

### AC-INL-001 — held while the holder session is alive (cross-process)
**Given** a temp project root, a parent test process that remains alive, and a child process
running the real built `moai integration acquire --session <id>` (no env stamp, ancestry path);
**When** the child exits and the parent runs `moai integration status --json` on the same root;
**Then** the window reads held with `stale == false` (NOT "reclaimable").
- **RED-now:** on `d29b8942e` the status output reads `stale: true` / "held by a session that
  is gone (reclaimable)" — the recorded pid is the exited child CLI's (integration_lock.go:174-176).
  The M1 test observing exactly this failure is the defect made mechanical; its observed
  failure output is the adoption evidence.
- **Green path:** M2 (owner-pid anchor) flips it; passing output becomes `stale: false` with
  the recorded pid equal to the parent test process's pid.

### AC-INL-002 — env-stamp resolution path
**Given** the same cross-process shape with the child env carrying
`MOAI_SESSION_PID=<parent pid>` (parent alive); **When** the child acquire exits and status is
read; **Then** the window reads held / not reclaimable, and the recorded pid equals the parent's.
- **RED-now:** current acquire ignores `MOAI_SESSION_PID` entirely and records the child's own
  pid → status reads reclaimable. Red for the right stated reason (missing resolution chain,
  not test-harness accident).
- **Green path:** M2 step 1 of the chain flips it. This variant needs no ancestry walk, so it
  is also the Windows-semantics proxy (see AC-INL-009).

### AC-INL-003 — reclaimable when the owner is genuinely gone (mutant-guard pair)
**Given** an anchored record whose owner pid was a spawned dummy process that has exited;
**When** a foreign session runs `moai integration acquire` (no `--force`) and then status;
**Then** the takeover succeeds without `--force` and the window reads held by the new session.
- **RED-now (by pairing):** green today only vacuously — every holder reads reclaimable on
  `d29b8942e`, so this criterion alone asserts nothing. It is adopted as the second half of the
  pair whose first half (AC-INL-001) is red now: a mutant that satisfies AC-INL-001 by always
  recording pid 0 (always-live) fails this criterion, and the pair together exclude both the
  vacuous and the impossible direction.
- **Green path:** M2 + M3 regression test; passing output shows the stale takeover reporting
  the displaced holder.

### AC-INL-004 — unresolvable owner is conservative, not dead
**Given** an anchored record (`pid_source: "session-owner"`) with pid 0; **When** any consumer
evaluates it; **Then** `Stale()` is false — the holder reads live until explicit release or
`--force`.
- **RED-now (mutant cell):** behaviorally green today only because `PID <= 0` already returns
  false (integration_lock.go:95-97) and no anchored record can exist yet. The guarded-against
  mutant is the M2 refactor itself going wrong ("fill pid 0 with `os.Getpid()` so the field is
  never empty") — that mutant passes AC-INL-001's letter and fails this criterion's world. The
  test pins the conservative path against regression during the fix.
- **Green path:** M2/M3; passing output shows held/live for the pid-0 anchored record.

### AC-INL-005 — foreign release refused (preserved invariant)
**Given** a window held by session A; **When** session B runs `moai integration release`
without `--force`; **Then** the release fails with the foreign sentinel
(`IsIntegrationLockForeign`).
- **RED-now (mutant cell):** green today at the in-process layer (existing test
  `TestIntegrationRelease_*` lineage) and unrelated to the pid defect — preserved invariant,
  protected because the M2 rewrite of the record path could plausibly disturb it. Its
  mutant: any schema/rewrite change that drops the SessionID comparison.
- **Green path:** M3 regression test remains green across the change.

### AC-INL-006 — takeover never silently overwrites a trace (preserved invariant)
**Given** a window held by session A (live, anchored); **When** session B acquires with
`--force`; **Then** the output reports the displaced holder (A's id/pid/since), and a stale
takeover likewise reports what it cleared.
- **RED-now (mutant cell):** green today for `--force` (existing
  `TestIntegrationAcquire_ForceReportsWhatItDisplaced`); preserved invariant with the same
  rationale as AC-INL-005. Mutant: the anchor refactor dropping the `replaced` return path.
- **Green path:** M3 extends the existing test to the anchored-record shapes.

### AC-INL-007 — legacy on-disk records behave as before (backward compat)
**Given** a temp root containing a hand-written old-format record (no `pid_source`, pid of an
exited process); **When** `moai integration status` and a foreign `acquire` run;
**Then** the record parses (no error), reads reclaimable, and the takeover succeeds — exactly
the pre-fix semantics; additionally, re-acquiring a window the caller already holds still
refreshes it.
- **RED-now (mutant cell):** green today by definition (legacy IS today's format). The
  criterion pins the additive-only schema contract; its mutants are (a) a field rename/retype
  that breaks parsing of old records, and (b) a semantics change that makes legacy records
  live-forever (upgrade wedge). Both are plausible M2 outcomes this row excludes.
- **Green path:** M3; passing output shows parse + reclaimable + successful takeover.

### AC-INL-008 — PreToolUse guard follows the anchored liveness
**Given** `workflow.integration_lock.enabled` semantics under test (the guard function
directly, per existing `integration_lock_guard_test.go` style) and a `git merge --no-ff
WT-x` Bash command from a foreign session; **When** the lock record's anchored holder is live;
**Then** the guard denies with the `INTEGRATION_LOCK_VIOLATION` sentinel; **When** the anchored
holder is gone; **Then** the guard allows with the reclaim advisory; **When** the anchored
record has pid 0; **Then** the guard denies (conservative).
- **RED-now:** the live-holder denial leg is red in substance today — a genuinely live holder's
  record reads stale, so the guard's stale branch (guard:95-97) allows what it should deny.
  Red for the right stated reason (the shared `Stale()` defect, not guard-local logic).
- **Green path:** M4 guard tests; all three legs observed on the same run.

### AC-INL-009 — Windows parity (compile + conservative degradation)
**Given** the changed packages; **When** `GOOS=windows GOARCH=amd64 go vet ./internal/...`
runs; **Then** it completes clean (no compile errors on the Windows liveness files); and the
platform-neutral env-stamp test (AC-INL-002) plus the pid-0 conservative test (AC-INL-004)
demonstrate the Windows degradation path: ancestry unavailable → conservative record, never
the acquire process's own pid.
- **RED-now:** `go vet` compiles today too; the RED cell for this row is the behavioral
  premise — the current code has no Windows degradation path to preserve because it has no
  anchor at all. Adopted via its mutant: any implementation that falls back to `os.Getpid()`
  where ancestry is unavailable (session_pid.go's step 3) reintroduces the defect on Windows
  and fails AC-INL-004's pid-0 contract.
- **Green path:** M4; `go vet` output clean + `git diff --stat` shows
  `factory_alive_windows.go` untouched.

### AC-INL-010 — gitflow-lane-protocol §3 caveat rewritten
**Given** `.claude/rules/local/gitflow-lane-protocol.md` on this branch carries the §3
blockquote "현재 이 락은 직렬화를 보장하지 못한다 — 카드 t298 …" (grep-hit at lines 42-43 on
`d29b8942e`); **When** the fix lands; **Then** that caveat text is gone, replaced by prose
describing the restored session-anchored guarantee and the legacy re-acquire note.
- **RED-now:** the caveat text exists on the baseline tree (grep returns ≥1 hit).
- **Green path:** M5; the same grep returns 0 hits and the replacement wording is present.

### AC-INL-011 — CLAUDE.local.md §4.1/§4.1.2 prose updated
**Given** `CLAUDE.local.md` §4.1/§4.1.2 documents lead-notice-only serialization as the
operating workaround; **When** the fix lands; **Then** the prose describes the restored
two-layer serialization (lead announcement + mechanically enforced record).
- **RED-now:** §4.1.2's operating procedure on `d29b8942e` still instructs the
  lead-notice-only workaround (no mention of the restored guarantee).
- **Green path:** M5; the section names the record-backed serialization.

## §D.2 Edge Cases

- Acquire run from a plain shell (no Claude env, no launcher stamp): ancestry walk finds the
  invoking shell's parent — an interactive terminal process that is long-lived; if genuinely
  unresolvable → pid 0 conservative. Either outcome honors the asymmetry.
- Acquire run with `MOAI_SESSION_PID` naming a DEAD process: the chain must skip the stamp
  (liveness-checked) and fall through, not record a dead pid.
- Record written by a lane that then `/clear`s and re-acquires: self-refresh must keep working
  (SessionID comparison, unchanged).
- Ancestry under an unknown wrapper (version-manager shim not in the wrapper set): worst case
  the walk anchors to a shorter-lived process → holder reads stale sooner than truth; bounded
  by the wrapper set's accuracy, identical exposure to the session registry's existing walk.

## §D.3 Severity and Closure Gates

- MUST-PASS: AC-INL-001, 002, 003, 004, 007, 008 (the contract, its mutant guards, backward
  compat, and the guard consumer).
- SHOULD-PASS: AC-INL-005, 006, 009, 010, 011 (preserved invariants, parity compile, docs).
- Definition of Done: all MUST-PASS observed green on the implementation tree with verbatim
  outputs recorded in progress.md §E.2; both doc rows landed; CI green on the PR head.

## §D.4 Indirect Verification

- `internal/session` seam unit tests (M2) cover the resolution precedence directly; the
  cross-process tests cover it end-to-end through the real binary.
- The guard is verified through its function-level tests rather than a live Claude Code hook
  run (the hook wrapper layer is unchanged by this SPEC).

## §D.5 Forward-Looking Checks

- A future migration of the lock record schema must treat `pid_source` as additive lineage
  (values may grow; absent must keep meaning "legacy").
- If the session registry ever becomes authoritative for lock staleness, spec.md §D's
  heartbeat-freshness constraint must be revisited explicitly, not silently.
