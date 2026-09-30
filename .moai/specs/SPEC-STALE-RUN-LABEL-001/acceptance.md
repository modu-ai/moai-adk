# SPEC-STALE-RUN-LABEL-001 — Acceptance Criteria

Every AC is mechanically verifiable without a live factory run: unit or integration tests on a `t.TempDir()` factory DB, run env-scrubbed in ONE compound invocation (lane env falsifies env-reading guard tests locally):

```bash
unset MOAI_FACTORY_WORKERS MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL && go test ./internal/hook/... ./internal/factorymsg/...
```

Two-cell discipline (`verification-completeness.md` §2): each release-blocking AC carries a RED-now cell observed on the pre-implementation tree and a green-path cell naming the milestone that flips it. The RED-now measurements for the Go-test ACs are acquired at M1's RED step (the tests do not exist on the plan-phase tree — the plan-phase author cannot execute them without writing test files, which is out of the plan-phase scope); an AC whose RED is not observed at M1 loses release-blocking eligibility and is reclassified as a regression-guard, not recorded as a pass. AC-SRL-009's RED cell is different: its baseline is already measured and pinned below (audit iteration 1).

## D. AC Matrix

| ID | Requirement | Scenario (Given / When / Then) | Severity | Test (owning milestone) |
|---|---|---|---|---|
| AC-SRL-001 | REQ-SRL-001 | **Given** a `t.TempDir()` factory DB whose `runs` row for run `R` has `status='retired'`, and a session env carrying `MOAI_KANBAN_ID=R` + `MOAI_FACTORY_WORKER=worker-69` (legacy label) | **When** the hook prescription path runs (`registerFactoryHookPeer` / `staleRunNoticeFor`) | **Then** the returned notice contains no `runs --retire` prescription text — silence or the unbind notice only (assert via the verbatim `"stale run:"` prefix per locale, session_stale_run.go:65-66) | Release-blocking | `TestStaleRunNoticeSilentWhenRunRetired` (M1) |
| AC-SRL-002 | REQ-SRL-001 (positive control) | **Given** the same setup but `runs.status='active'` | **When** the prescription path runs | **Then** the stale-run prescription IS emitted (the gate measures state; it does not blanket-suppress) | Release-blocking | `TestStaleRunNoticeFiresWhenRunActive` (M1) |
| AC-SRL-003 | REQ-SRL-002 | **Given** an active run and a legacy-label session that has already received the prescription once in its session identity | **When** the prescription path runs again for the same session identity | **Then** the answer is silence (no second prescription) — and the dedup carrier is shared across ALL prescription surfaces of the identity, so startup + first prompt cannot both emit | Release-blocking | `TestStaleRunNoticeOncePerSession` (M1) |
| AC-SRL-004 | REQ-SRL-004 | **Given** a dead-run env (`MOAI_KANBAN_ID=R` retired, `MOAI_FACTORY_WORKER=worker-69`) and a SessionStart event with `source="clear"` | **When** the session-start handler evaluates factory identity | **Then** no factory peer binding is established and no effective dead-run binding is carried into the fresh session | Release-blocking | `TestClearSourceDeadRunEnvYieldsUnbound` (M2) |
| AC-SRL-005 | REQ-SRL-005, REQ-SRL-006 | **Given** a session env carrying an orphan label (`MOAI_FACTORY_WORKER=worker-69`, `MOAI_KANBAN_ID=R`, `runs.status='retired'`) | **When** the unbind path runs on that session, and again on its next turn | **Then** (a) exactly ONE unbind notice is emitted naming the orphan label and the measured run state, and every subsequent turn of that session identity is silent; (b) the notice names the `moai cc -f lane-<n>` re-bind entry iff an active run exists in the same root, and omits the line otherwise | Release-blocking | `TestUnbindNoticeThenSilence` + `TestUnbindNoticeRebindLinePresence` (M2) |
| AC-SRL-006 | REQ-SRL-007 | **Given** a session with a LIVE broker peer record in run `R`'s broker, and env carrying a legacy label for a run measured not active | **When** `factoryHookBatch` claims messages | **Then** claims are returned exactly as before the repair — the env-label judgment does not gate, filter, or alter claim delivery (worker-70 separation preserved) | Release-blocking | `TestInboundClaimIndependentOfEnvLabel` (M3) |
| AC-SRL-007 | REQ-SRL-008 | **Given** a current-vocabulary label (`lane-3`) and an active run | **When** the bind path runs | **Then** behavior is unchanged pre/post repair (bind succeeds through `ValidateActiveRun` → `BindLaunchPending`/`RegisterPeer`; characterization assertion) | High | `TestCurrentVocabularyBindPathUnchanged` (M3) |
| AC-SRL-008 | REQ-SRL-003 | **Given** any prescription-gate decision | **When** the gate reads run state | **Then** it reads through the shared tri-state accessor over `homestate.FactoryDBPath` (`runs.status`) — a measurement failure ("unavailable": busy DB, spent inspection budget) degrades to the existing degraded answer, never to a prescription and never to a hook error | Release-blocking | `TestPrescriptionGateUnavailableFailsOpen` (M1) |
| AC-SRL-009 | REQ-SRL-009 | **Given** the card's implementation diff (`git diff <card-base>...HEAD`) | **When** the diff-scoped literal extraction runs — added lines only, `internal/config/envkeys.go` excluded, added-line count logged | **Then** the extracted `MOAI_(FACTORY|KANBAN)[A-Z_]*` literal set is EMPTY | High | `TestNoNewEnvNameLiteralsInDiff` (M3) |

### AC-SRL-009 instrument (D2 restatement — replaces the whole-tree sweep)

The verdict path is diff-scoped added-line extraction, executed as `TestNoNewEnvNameLiteralsInDiff` so it cannot silently not run (§1.3 continued firing); the test logs the added-line count it swept so an empty sweep is visible, never silent (§1.1):

```bash
git diff <card-base>...HEAD -- internal/hook internal/factorymsg internal/cli ':!*envkeys.go' \
  | grep -E '^\+' | grep -vE '^\+\+\+' | grep -oE 'MOAI_(FACTORY|KANBAN)[A-Z_]*' | sort -u
```

Assert: empty output. (Test files are excluded by the diff scope of the owning package tests; `_test.go` additions carrying env names in fixtures are acceptable and are excluded by the same `:!` pathspec pattern extended with `':!*_test.go'`.)

**RED cell — measured, pinned** (audit iteration 1, plan-audit-r1.md §3, tree @ d194083fb): the same extraction run WHOLE-TREE (not diff-scoped) returns **28** literals — internal/hook + internal/factorymsg **12**, internal/cli **16** — including measured false-positive sources (i18n copy literals at `session_start_factory_i18n.go:116`, comment mention at `contract_sign_guard.go:20`). The whole-tree count is the positive control proving the extractor has signal; it is also why the verdict path is diff-scoped — the old whole-tree sweep instrument returned 28 on a tree with zero new literals (report-not-verdict, verification-completeness.md §1.1) and its `grep -v 'config\.'` exclusion admitted new literals on any config-referencing line (false-negative direction). Mutate probe at M3: a working-tree edit adding one `MOAI_FACTORY_*` literal flips the diff-scoped extraction non-empty — the failure observed on a known failing input before the AC is adopted.

### RED-now discipline note

AC-SRL-001's RED-now on the pre-implementation tree: the current `registerFactoryHookPeer` legacy branch returns the retire prescription unconditionally, so the (future) test asserting silence fails red for the RIGHT stated reason — the branch returns at `factory_messages.go:64-68` before any run-state measurement. The verbatim command, stdout, exit code, and tree SHA are recorded at M1 RED per `verification-completeness.md` §2.1 and pinned in the M1 evidence file. AC-SRL-009's RED cell above is already measured and attributed (auditor, this tree @ d194083fb).

## D.1 Severity Scale

Release-blocking: the defect class this SPEC exists to kill (infinite prescription, dead-run binding, separation coupling). High: quality/preservation guards.

## D.2 Edge Cases

- Run DB file absent entirely (never a factory project) → not-active verdict → unbind path or silence, never a prescription.
- Inspection budget spent (busy DB, 200ms deadline) → tri-state "unavailable" → existing degraded answer (fail-open), never a prescription, never a hook error.
- Unbind notice when NO active run exists in the root → notice omits the re-bind line (REQ-SRL-006 negative branch).
- Env label current-vocabulary but run retired → same unbound semantics as the legacy-label case (the gate keys on run state, not label vocabulary — the legacy branch is one caller of the shared gate).
- Session identity changes via `clear` after an unbind notice → the new identity may receive one unbind notice again (correct: it is a new session).
- Diff contains zero added Go lines → instrument logs the empty sweep and PASSES only with the logged count recorded in the verdict; a silent pass is invalid.

## D.3 Quality Gates

- `go vet` + `golangci-lint run` clean on changed packages.
- Package coverage per `quality.yaml` `test_coverage_target` on `internal/hook`, `internal/factorymsg`.
- E4 separation diff review: no env-label reference added inside `factoryHookBatch`'s call tree.
- E5 diff-scoped env-literal instrument (AC-SRL-009) with its logged swept count.

## D.4 Definition of Done

1. All release-blocking ACs PASS with verbatim test output recorded (E1).
2. RED-now cells observed and pinned for every release-blocking AC (or the AC reclassified regression-guard with reason).
3. Worker-70 separation test green and the diff review clean (AC-SRL-006).
4. Card t1345 self-healing scope untouched: no `moai factory relaunch` verb, no auto re-registration logic in the diff.
5. No template-tree changes (verify: `git diff --name-only | grep internal/template/templates` empty).
