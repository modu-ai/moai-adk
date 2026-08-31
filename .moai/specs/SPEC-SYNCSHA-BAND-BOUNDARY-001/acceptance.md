# SPEC-SYNCSHA-BAND-BOUNDARY-001 — Acceptance Criteria

Tier S. **Eight criteria** (the Tier S ceiling). Every criterion that measures
rule behavior names the **exact mutation** that must turn it red, and the run
phase must observe that mutation red and then revert it — an unobserved mutation
claim is not evidence the criterion measures anything
(`verification-claim-integrity.md` §1).

Amended after the plan audit (`.moai/reports/t380/plan-audit.md`,
PASS-WITH-DEBT 0.9216): the former AC-SBB-005 and AC-SBB-008 are merged into one
criterion (audit D2), so the set is contiguous at 001..008; the merged criterion
carries the D1-relaxed delivery-shape obligation.

---

## §A. Fixture precondition (inherited, [HARD])

Every fixture under `internal/spec/testdata/syncsha/<case>/` carries
`status: in-progress` (non-terminal) and a `progress.md` satisfying era heuristic
H-4 (§E.2 + §E.4 + a non-empty `sync_commit_sha`), so nothing is demoted to
advisory. A mis-built fixture would break a criterion for a reason having nothing
to do with the band under test. The new fixtures follow the shape of
`internal/spec/testdata/syncsha/sha-short/{spec.md,progress.md}` measured in this
tree, differing only in the `sync_commit_sha` value and the title text.

Fixtures share the SPEC id `SPEC-SSFA-001` deliberately and are linted ONE AT A
TIME through `syncSHAFindings` (`lint_syncsha_test.go:50`), so
`DuplicateSPECIDRule` never sees two at once.

---

## §B. RED-now evidence ledger

Referenced by id from the criteria below. Two entries carry all four elements of
`verification-completeness.md` §2.1; two carry three, and the missing element is
named rather than inferred. **Document-level tree pin: `3f03d9c36`** — it binds
every entry and every criterion that carries no pin of its own.

### R-1 — floor narrowed (`{8,40}`), the t299 surviving mutant

- **Command**: `go test ./internal/spec/ -run TestSyncSHASlot_SilentOnSHA -count=1 -v`
- **Verbatim stdout**:
  ```
  --- FAIL: TestSyncSHASlot_SilentOnSHA (1.35s)
      lint_syncsha_test.go:128: sha-min7: expected 0 findings on a well-formed SHA, got 1: [{testdata/syncsha/sha-min7/progress.md 9 warning SyncSHASlotFormat `sync_commit_sha` holds "19b6f76", which is neither a commit SHA nor a recognized backfill placeholder, ...}]
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/spec	1.723s
  ```
- **Exit code**: not recorded as a separate field in the source
  (`.moai/reports/t380/probe-05-mutant-red.txt`); the `FAIL` line is the whole of
  the observation. Run phase records the exit code as its own field.
- **Tree**: `3f03d9c36` — plan-phase probe by this author, mutant fully reverted.

### R-2 — floor widened (`{6,40}`)

- **Command**: `go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1`
- **Verbatim stdout**:
  ```
  --- FAIL: TestSyncSHASlot_FlagsProse (0.52s)
      lint_syncsha_test.go:98: sha-below6: expected exactly 1 SyncSHASlotFormat finding, got 0: []
  ```
- **Exit code**: not recorded as a separate field by the source.
- **Tree**: `3f03d9c36` — planted by the plan audit, not by this author, and
  reverted there. Note the function name in this output is the audit's own M2
  reproduction, which the D1 ruling has since superseded: at run phase the same
  observation lands in `TestSyncSHASlot_FlagsOutOfBand`.

### R-3 — ceiling narrowed (`{7,39}`)

- **Command**: `go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1`
- **Verbatim stdout**: **absent.** The plan audit reports this mutant RED, killed
  by `sha-full`, in its results table only; it published no verbatim output for it.
- **Exit code**: absent, same reason.
- **Tree**: `3f03d9c36`. Attributed to `.moai/reports/t380/plan-audit.md`.

### R-4 — ceiling widened (`{7,41}`)

- **Command**: `go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1`
- **Verbatim stdout**: **absent**, same as R-3 — reported RED, killed by
  `sha-above41`, table entry only.
- **Exit code**: absent.
- **Tree**: `3f03d9c36`. Attributed to `.moai/reports/t380/plan-audit.md`.

**Gap, stated rather than papered over**: R-3 and R-4 are three-element entries.
No criterion below treats them as though they were four-element. Run phase
completes all four entries with its own measurement, including the exit code as
its own field.

---

## §C. Criteria

### AC-SBB-001 — the band floor, inside, is silent

**Given** the fixture `internal/spec/testdata/syncsha/sha-min7/progress.md` whose
`sync_commit_sha` value token is exactly 7 hex characters,
**When** `go test ./internal/spec/ -run TestSyncSHASlot_SilentOnSHA -count=1` runs,
**Then** `SyncSHASlotFormatRule` produces **0** findings for that fixture.

Mutation that must turn it red: narrow the floor by one — replace the
`isCommitSHAToken` call at `internal/spec/lint_syncsha.go:103` with an inlined
`^[0-9a-fA-F]{8,40}$` match (the t299 surviving mutant, verbatim). The criterion
must report `sha-min7: expected 0 findings on a well-formed SHA, got 1`.

RED-now: **R-1** (§B). Four-element except the exit code, which is named absent
there and is recorded at run phase.

Maps REQ-SBB-001, REQ-SBB-003.

### AC-SBB-002 — the band ceiling, inside, is silent

**Given** the EXISTING fixture `internal/spec/testdata/syncsha/sha-full/progress.md`,
whose value token was measured in this tree at 40 hex characters and which is
already registered in `TestSyncSHASlot_SilentOnSHA` (`lint_syncsha_test.go:126`,
resolved at base `3f03d9c36`; this card's M4 commit moved that line to `:138`),
**When** the same command runs,
**Then** `SyncSHASlotFormatRule` produces **0** findings for that fixture.

Mutation that must turn it red: narrow the ceiling by one — inline
`^[0-9a-fA-F]{7,39}$` at `lint_syncsha.go:103`. The criterion must report
`sha-full: expected 0 findings on a well-formed SHA, got 1`.

No new fixture is created for this criterion: `sha-full` already sits at the
inside edge of the ceiling. A duplicate would catch nothing it does not.

RED-now: **R-3** (§B) — three-element; the verbatim stdout is absent and is
supplied at run phase.

Maps REQ-SBB-001, REQ-SBB-003.

### AC-SBB-003 — one below the floor is flagged

**Given** the fixture `internal/spec/testdata/syncsha/sha-below6/progress.md` whose
`sync_commit_sha` value token is exactly 6 hex characters,
**When** `go test ./internal/spec/ -run TestSyncSHASlot_FlagsOutOfBand -count=1` runs,
**Then** `SyncSHASlotFormatRule` produces **exactly 1** finding, of severity
`warning`, against the sibling `progress.md`, non-advisory, on the
`sync_commit_sha` line.

Mutation that must turn it red: widen the floor by one — inline
`^[0-9a-fA-F]{6,40}$` at `lint_syncsha.go:103`. The finding count must drop to 0.

RED-now: **R-2** (§B). The cited output names `TestSyncSHASlot_FlagsProse`, which
was the audit's reproduction under the pre-ruling design; the observation is about
`sha-below6` and transfers unchanged to the new function.

Maps REQ-SBB-002, REQ-SBB-004.

### AC-SBB-004 — one above the ceiling is flagged

**Given** the fixture `internal/spec/testdata/syncsha/sha-above41/progress.md` whose
`sync_commit_sha` value token is exactly 41 hex characters,
**When** `go test ./internal/spec/ -run TestSyncSHASlot_FlagsOutOfBand -count=1` runs,
**Then** `SyncSHASlotFormatRule` produces **exactly 1** finding, of severity
`warning`, against the sibling `progress.md`, non-advisory, on the
`sync_commit_sha` line.

Mutation that must turn it red: widen the ceiling by one — inline
`^[0-9a-fA-F]{7,41}$` at `lint_syncsha.go:103`. The finding count must drop to 0.

RED-now: **R-4** (§B) — three-element; verbatim stdout supplied at run phase.

Maps REQ-SBB-002, REQ-SBB-004.

### AC-SBB-005 — delivery shape, and no production source change

*(Merged from the pre-audit AC-SBB-005 and AC-SBB-008 per the D2 ruling: both are
read off the same `git diff --stat` and neither measures runtime behavior. Both
obligations are carried below; neither is dropped.)*

**Given** the run-phase diff,
**When** `git diff --stat` is read against the base `3f03d9c36`,
**Then** all three hold:

1. **No production source change.** None of `internal/spec/syncsha.go`,
   `internal/spec/lint_syncsha.go`, or `internal/spec/closer.go` appears in the
   diff. The delivery touches only files under
   `internal/spec/testdata/syncsha/` plus `internal/spec/lint_syncsha_test.go`.
   The planted mutants of AC-SBB-006 / AC-SBB-007 are transient and reverted; a
   surviving edit to `lint_syncsha.go` at close time fails this criterion.
2. **No duplicated assertion body.** `TestSyncSHASlot_FlagsOutOfBand` carries its
   own case list against the shared helper `syncSHAFindings`
   (`lint_syncsha_test.go:50`); it does not copy the assertion body of
   `TestSyncSHASlot_FlagsProse` or of any other existing function. Verified by
   reading the diff, not by counting lines.
3. **`TestSyncSHASlot_FlagsProse` is untouched in name and behavior.** It carries
   the prose case only, so t299's AC-SSF-001 command string
   (`go test ./internal/spec/ -run TestSyncSHASlot_FlagsProse`) remains true.

The `func Test` count in `lint_syncsha_test.go` becomes **4** (measured at 3 in
this tree at `3f03d9c36` via `grep -c '^func Test'`). That increase is expected
and is **not** a violation: the pre-audit form of this criterion pinned the count
at 3, and the D1 ruling relaxed it precisely because that pin made the
semantically correct home for the outside-band cases unreachable. The relaxed
constraint is this card's own, not another SPEC's.

Mutation that must turn it red: none — this criterion measures the shape of the
delivery, not runtime behavior, and is verified by reading the diff.

Maps REQ-SBB-005, REQ-SBB-007.

### AC-SBB-006 — a band-WIDENING mutant is observed red across the outside pair

**Given** a single planted mutant widening the band at `lint_syncsha.go:103`
(either `{6,40}` or `{7,41}`),
**When** `go test ./internal/spec/ -run TestSyncSHASlot -count=1` runs,
**Then** at least one of `sha-below6` / `sha-above41` reports a criterion failure,
the verbatim FAIL output **and the exit code as its own field** are captured to
`.moai/reports/t380/`, and the mutant is reverted with the suite observed GREEN
again afterwards.

Both widenings are planted and observed separately — one run per direction. A
single planted mutant satisfying only one direction does not discharge this
criterion.

RED-now: **R-2** (floor widened, four-element less the exit code) and **R-4**
(ceiling widened, three-element). Run phase re-measures both under its own hands
and completes the missing elements; the ledger entries are prior observations,
not a substitute for the run-phase measurement this criterion requires.

Maps REQ-SBB-006.

### AC-SBB-007 — a band-NARROWING mutant is observed red across the inside pair

**Given** a single planted mutant narrowing the band at `lint_syncsha.go:103`
(either `{8,40}` or `{7,39}`),
**When** the same command runs,
**Then** at least one of `sha-min7` / `sha-full` reports a criterion failure, the
verbatim FAIL output **and the exit code as its own field** are captured to
`.moai/reports/t380/`, and the mutant is reverted with the suite observed GREEN
again afterwards.

Both narrowings are planted and observed separately. Together with AC-SBB-006
this discharges the bidirectional obligation of REQ-SBB-006: four planted
mutants, four observed reds, four reverts, all four verbatim outputs and exit
codes recorded.

The `{8,40}` narrowing is the t299 surviving mutant itself; observing it red is
what closes debt D1.

RED-now: **R-1** (floor narrowed) and **R-3** (ceiling narrowed).

Maps REQ-SBB-006.

### AC-SBB-008 — AC-SSF-007 is untouched and its distinction is preserved

**Given** `.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/acceptance.md`,
**When** `git diff` is read against the base,
**Then** that file is unchanged, AC-SSF-007's grep criterion is neither edited nor
restated here, and this SPEC's §C records — in prose a later reader can act on —
that AC-SSF-007's remaining gap is CI **automation**, a different gap from the
fixture-corpus gap this card closes.

*(Renumbered from the pre-audit ninth criterion. The pre-audit eighth criterion's
obligation did not disappear with the renumbering: it is clause 1 of the merged
AC-SBB-005.)*

Maps REQ-SBB-008.

---

## §D. Definition of Done

- [ ] Three new fixture directories exist (`sha-min7`, `sha-below6`,
      `sha-above41`), each with `spec.md` + `progress.md` matching the §A shape.
- [ ] `sha-full` is reused for the ceiling-inside case; no fourth fixture is added.
- [ ] `sha-min7` is registered in `TestSyncSHASlot_SilentOnSHA`'s case list.
- [ ] `TestSyncSHASlot_FlagsOutOfBand` exists, carries `sha-below6` and
      `sha-above41`, and duplicates no existing assertion body.
- [ ] `TestSyncSHASlot_FlagsProse` is unchanged in name and behavior.
- [ ] Four mutants planted, each observed red, each reverted; four verbatim FAIL
      outputs plus their exit codes under `.moai/reports/t380/`, completing
      ledger entries R-1..R-4.
- [ ] `go test ./internal/spec/... -count=1` GREEN on the clean tree, exit code
      recorded.
- [ ] `go vet ./internal/spec/...` clean.
- [ ] AC-SBB-001..008 each marked PASS or FAIL with its evidence path in
      `progress.md` §E.2, in the four-element form rather than a summary matrix.
