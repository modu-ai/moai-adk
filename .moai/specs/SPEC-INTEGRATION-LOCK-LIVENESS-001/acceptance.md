# SPEC-INTEGRATION-LOCK-LIVENESS-001 — Acceptance Criteria (card t298)

Baseline tree for every RED-now cell: `d29b8942e` @ `WT-integration-lock` (worktree t298);
cells re-measured after the develop merge name `c67a6ea64` and cite the drift check.
Two-cell discipline per `.claude/rules/moai/development/verification-completeness.md` §2: each
criterion carries a RED-now cell stating why it is red (or, for preserved-invariant criteria,
its mutant-guard pairing) and a green-path cell naming the flipping milestone. All tests run
under `t.TempDir()` with `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES` pinned; the real
primary-checkout state files are never touched.

## §D AC Matrix

| AC | Requirement | Kind | Flipping milestone |
|---|---|---|---|
| AC-INL-001 | REQ-INL-001, REQ-INL-002, REQ-INL-004 | RED-first core | M2 |
| AC-INL-002 | REQ-INL-002 | RED-first (env-stamp path) | M2 |
| AC-INL-003 | REQ-INL-005 | Mutant-guard pair | M2 |
| AC-INL-004 | REQ-INL-003 | Conservative invariant | M2 |
| AC-INL-005 | REQ-INL-010 | Preserved invariant | M3 |
| AC-INL-006 | REQ-INL-010 | Preserved invariant | M3 |
| AC-INL-007 | REQ-INL-006 | Backward compat | M3 |
| AC-INL-008 | REQ-INL-004, REQ-INL-008 | Guard consumer | M4 |
| AC-INL-009 | REQ-INL-007 | Platform parity | M4 |
| AC-INL-010 | REQ-INL-009 | Documentation | M5 |
| AC-INL-011 | REQ-INL-009 | Documentation | M5 |
| AC-INL-012 | REQ-INL-004, REQ-INL-010 | RED-first (acquire refusal) | M2 |
| AC-INL-013 | REQ-INL-009 | Negative assertion (doc wording bound) | M5 |

## §D.1 Criteria (Given-When-Then, two cells each)

### AC-INL-001 — held while the holder session is alive (cross-process)
**Given** a temp project root, a parent test process that remains alive, and a child process
running the real built `moai integration acquire --session <id>` (no env stamp, ancestry path);
**When** the child exits and the parent runs `moai integration status --json` on the same root;
**Then** BOTH hold — (a) the window reads held with `stale == false` (NOT "reclaimable"), and
(b) the record that acquire actually wrote carries the owner-anchor marker
`pid_source == "session-owner"`, read back from the on-disk
`.moai/state/integration-lock.json` under the temp root (or from the `status --json` output,
whichever the M2 shape surfaces) after the child has exited.
- **RED-now (a):** on `d29b8942e` the status output reads `stale: true` / "held by a session that
  is gone (reclaimable)" — the recorded pid is the exited child CLI's (integration_lock.go:174-176).
  The M1 test observing exactly this failure is the defect made mechanical; its observed
  failure output is the adoption evidence.
- **RED-now (b):** the marker field does not exist on the baseline tree —
  `grep -n 'PIDSource\|pid_source' internal/kanban/integration_lock.go` returns no match
  (rc=1; measured on `c67a6ea64`, whose `internal/kanban/integration_lock.go` is unchanged from
  `d29b8942e`), and the `IntegrationLock` struct (integration_lock.go:71-79) carries exactly
  `SessionID, SessionName, PID, Branch, Worktree, AcquiredAt, Card`. Clause (b) exists because
  clause (a) alone is mutant-satisfiable: an implementation that resolves the owner pid, writes
  it to `PID`, and never adds the marker field satisfies every other criterion in this set while
  violating REQ-INL-002's recording obligation and leaving REQ-INL-006's legacy discriminator
  unconstructible. The plan ranks the schema field as the least-reversible decision (plan.md
  §B consequence 1), so it is the one decision that must be measured directly.
- **Green path:** M2 (owner-pid anchor) flips both; passing output becomes `stale: false` with
  the recorded pid equal to the parent test process's pid AND `pid_source: "session-owner"`
  present in the written record.

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

### AC-INL-011 — CLAUDE.local.md §4.1 states what the acquired hold guarantees
**Given** `CLAUDE.local.md` §4.1 carries the git-flow integration-lane model and its §4.1.2
procedure invokes `moai integration acquire` / `moai integration release` without stating what
the recorded hold guarantees; **When** the fix lands; **Then** §4.1 names the session-anchored
guarantee and the legacy re-acquire consequence, verified by
`awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md | grep -c '세션 프로세스'` → **≥1** and
`awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md | grep -c '재획득'` → **≥1**.
- **RED-now (re-measured; supersedes the earlier premise, which was false):** this file carries
  no lead-notice-only workaround prose at all. Measured on `c67a6ea64`
  (`git diff --stat d29b8942e HEAD -- CLAUDE.local.md` → empty, so the baseline text is
  unchanged): `awk '/^#### §4\.1\.2/,/^#### §4\.1\.3/' CLAUDE.local.md` is a bash procedure
  block whose only lock-related lines are `moai integration acquire --name <lane>` (line 312)
  and `moai integration release` (line 317); `grep -n '직렬화' CLAUDE.local.md` returns no
  match. What is red is the **absence**: the file never says whether the acquired hold is
  advisory or enforced, and the one pointer it gives for that answer — §4.1.1's closing line,
  "레인/리드가 따르는 운영 규칙 정본은 `.claude/rules/local/gitflow-lane-protocol.md`" — leads
  to the document whose §3 caveat AC-INL-010 deletes. Once that caveat is gone, this file's
  silence is the only surviving statement about the hold, so M5 ADDS the guarantee sentence
  here rather than rewriting anything. Baseline for the Then's two greps, measured in this run:
  both return **0**.
- **Green path:** M5; both grep counts flip to ≥1 with the added text inside §4.1.

### AC-INL-012 — a live holder's window is not transferred by a bare acquire (cross-process)
**Given** a temp project root, a parent test process that remains alive, and a child process
running the real built `moai integration acquire --session sess-a` that has since EXITED (the
same cross-process shape as AC-INL-001, so the recorded owner is the still-live parent);
**When** a second child runs `moai integration acquire --session sess-b` **without** `--force`;
**Then** BOTH hold — (a) that second acquire FAILS: it exits non-zero, its error carries the
`ErrIntegrationLockHeld` contention sentinel naming `sess-a`, and the on-disk
`.moai/state/integration-lock.json` under the temp root still records `session_id == "sess-a"`
(the window was not transferred); and (b) NO displacement record is produced by that refusal —
the refused acquire emits no `displaced:` line, and with `--json` its `replaced` field is
absent/null. Displaced-holder bookkeeping belongs to the takeover paths (`--force` and
stale-reclaim), never to a refusal.
- **Traceability:** clause (a) is REQ-INL-004's "acquire refusal" consumer leg — the one leg of
  that requirement no other criterion observes (AC-INL-001 observes the `status` consumer,
  AC-INL-008 the guard consumer). Clause (b) is REQ-INL-010's "shall not silently overwrite
  another lane's recorded trace", read in its refusal direction: after a refused acquire the
  recorded trace is still A's, and nothing reports it as displaced. **No new REQ was added** —
  the obligation is already carried by REQ-INL-004 and REQ-INL-010.
- **Why this row exists (the gap it closes):** nothing in AC-INL-001..011 gates the ACQUIRE
  refusal path. AC-INL-001 gates the STATUS read and AC-INL-003 gates the takeover direction, so
  a mutant that anchors the owner pid correctly for `Stale()`/`status` while leaving `acquire`
  willing to displace a live holder passes the entire existing set. That mutant is not
  hypothetical: it is the harm observed in the field on 2026-08-27 (spec.md §A, field
  reproduction) — a second lane's bare `acquire` displaced a live holder and both lanes were
  told the window was theirs.
- **RED-now (measured on `c67a6ea64`, this tree; the runtime leg is measured-by-test, see
  below):** the refusal branch (`integration_lock.go:161-168`) is reached only when
  `!force && !current.Stale()`, and `AcquireIntegrationLock` records the acquire CLI's own pid
  (`want.PID = os.Getpid()`, integration_lock.go:175), which is dead the moment that child
  exits — so `current.Stale()` is true and control takes the `case current.Stale():` reclaim
  arm (line 163) instead: the bare acquire SUCCEEDS and reports `displaced: sess-a`
  (`internal/cli/integration.go:196-199`). Both clauses therefore fail today. Existing coverage
  cannot see this: `runIntegration` (integration_lock_cli_test.go:26-43) executes the cobra
  command **in-process**, so the recorded pid is the
  still-alive `go test` process and the two existing refusal tests —
  `TestIntegrationAcquire_RefusesASecondLane` (integration_lock_cli_test.go:106) and
  `TestAcquireIntegrationLock_RefusesASecondLiveSession` (internal/kanban/integration_lock_test.go:60)
  — are green **vacuously** with respect to the field shape. That no cross-process refusal
  coverage exists is measured, not assumed:
  `grep -rln "exec.Command" internal/cli/integration_lock_cli_test.go internal/kanban/integration_lock_test.go`
  → no output, `rc=1`.
- **Measured-by-test, not by hand:** the runtime cell above is deliberately NOT backed by a live
  `moai integration acquire` run. Other lanes hold real merge windows in this repository right
  now, and the verb mutates the shared primary-checkout record; running it to produce a RED
  transcript would take a live lane's window. The failing observation is therefore produced by
  the M1 Go test named below, under `t.TempDir()` with `CLAUDE_PROJECT_DIR` +
  `GIT_CEILING_DIRECTORIES` pinned, and its verbatim output is recorded in progress.md §E.2 as
  the adoption evidence.
- **Green path:** authored RED in **M1** (`TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder`,
  alongside the AC-INL-001/002 cross-process tests, sharing their `buildMoaiBinary` /
  `runIntegrationChild` helpers); flipped green by **M2** (owner-pid anchor), after which the
  live parent's pid is what `Stale()` probes, the refusal branch is reached, and the passing
  output shows the non-zero exit naming `sess-a`, the record still holding `sess-a`, and no
  `displaced:` line.

### AC-INL-013 — neither shipped document claims a serialization guarantee (negative assertion)
**Given** plan.md §F M5 carries a [HARD] wording bound on the prose this SPEC ships — forbidding
"직렬화를 보장한다" and any equivalent unqualified guarantee claim, "동시 acquire는 불가능하다",
and any wording implying the record is a capability boundary rather than a coordination signal —
and no criterion in AC-INL-001..012 measures that bound; **When** M5 lands; **Then** the
forbidden-phrase sweep over BOTH shipped documents returns **0**:

```bash
grep -hoE '직렬화를[[:space:]]*보장|동시.*acquire.*불가능|acquire.*동시.*불가능' \
  .claude/rules/local/gitflow-lane-protocol.md CLAUDE.local.md | wc -l
```

Scope is exactly those two paths — the SPEC artifacts under `.moai/specs/` necessarily quote the
forbidden strings to define them and are never swept.

- **Why this row exists (the gap it closes, and the harm behind it):** spec.md §G records that
  the acquire path remains an unserialized read-modify-write which this SPEC does **not** fix. The
  lead announcement is therefore currently the FIRST serialization layer, and two lanes have
  already read `moai integration acquire` output as window ownership and collided (spec.md §A,
  field reproduction, 2026-08-27). A shipped document claiming the lock now guarantees
  serialization would cause lanes to drop the announcement layer — converting a documented,
  operator-visible coordination step into a false mechanical guarantee. The surviving mutant this
  row kills: a documentation edit that deletes the §3 caveat (satisfying AC-INL-010), adds
  AC-INL-011's two required tokens, **and also** writes 직렬화를 보장한다 passes every criterion in
  AC-INL-001..012 while landing that false guarantee in a [HARD] doctrine document.
- **Traceability:** REQ-INL-009 (the documentation requirement) — this row measures the negative
  half of what M5 ships, where AC-INL-010/011 measure the positive half. REQ-INL-010 does **not**
  apply: it governs the acquire verb's displaced-holder reporting, not shipped prose. **No new REQ
  was added.**
- **RED-now (measured on this tree, `6d3f8251c` @ `WT-integration-lock`):** the command above
  returns **1** — `.claude/rules/local/gitflow-lane-protocol.md:42`, the §3 caveat blockquote
  "현재 이 락은 직렬화를 보장하지 못한다 — 카드 t298 …". `CLAUDE.local.md` contributes **0**.
  The first pattern arm is deliberately **polarity-agnostic**: it matches `직렬화를 보장` whatever
  follows, so it hits today's negated caveat AND a future positive claim. That choice is what
  makes the criterion non-vacuous — the positive-only form `직렬화를[[:space:]]*보장한다` measures
  **0** on both files on this tree (measured in this run), so a positive-only pattern would be
  green at arrival and would assert nothing about the work.
- **Consequence for M5's replacement prose (binding):** the shipped replacement must state the
  mechanism positively — the hold survives the acquire CLI's exit and is released by the holder's
  death, an explicit release, or a recorded `--force`; the lead announcement remains the doctrine's
  first layer — and must **not** restate the honest negated form 직렬화를 보장하지 못한다 either.
  None of plan.md §F M5's required-instead wording contains a swept token, so writing it as
  prescribed does not trip this check.
- **Mutant probe (run against scratch copies; the real files were not modified):** correct-M5
  simulation (line 42 deleted, nothing added) → **0**; mutant A (deletion + `이 락은 이제
  직렬화를 보장한다.`) → **1**; mutant B (mutant A + `동시 acquire는 불가능하다.` in
  CLAUDE.local.md) → **2**. The criterion fails on both mutants and passes only the correct edit.
- **Reviewer-judgment sub-clause (not greppable):** the third prohibition — wording implying the
  record is a capability boundary rather than a coordination signal — is a judgment about meaning,
  not a token match. It is asserted by review of the M5 diff, not by the command above; the grep
  is the mechanical floor, not the whole bound.
- **Green path:** M5; the same command returns 0 after the caveat is replaced by the bounded
  statement, with AC-INL-010/011 independently confirming the required wording is present.

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

- MUST-PASS: AC-INL-001, 002, 003, 004, 007, 008, 012, 013 (the contract, its mutant guards,
  backward compat, the guard consumer, the acquire-refusal path, and the doc wording bound —
  013 is MUST-PASS because its failure mode is a false mechanical guarantee in a [HARD] doctrine
  document, which would retire the announcement layer that is currently the first one).
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
