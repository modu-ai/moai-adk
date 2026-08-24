# SPEC-GATE-THREE-AXES-001 — Acceptance Criteria

16 criteria, at the Tier M ceiling. Every one is binary-testable and carries a **Mutant** line: an implementation that would satisfy the criterion while violating its REQ. Where the mutant is not killed by the criterion as written, the criterion is wrong and is rewritten until it is.

Two standing rules govern this file:

- **A check is complete only when a failing input has been observed actually failing.** Each criterion names its failing input under **RED**. A criterion whose RED was never observed failing is not met, regardless of what the green run reports.
- **A grep-shaped criterion is admissible only when its token returns 0 on the pre-implementation tree.** No criterion below is grep-shaped: every measured candidate token (`Setpgid` / `Killpg` / `SysProcAttr` / `WaitDelay` in `internal/hook/quality/`, lock acquisition in `internal/cli/gate.go`) has a baseline of 0 at tree SHA **294b4b6ab** and would therefore have been admissible, but each is only a structural marker — present-in-a-comment satisfies it while the behaviour stays broken. The behavioural criteria below carry the verification instead. The measured baselines are recorded in `progress.md §E.1` so a later reader can see the check was made rather than skipped.

---

## §D.1 Milestone 1 — Axis 1: a completed run reports what it ran

### AC-GTA-001 — the summary is complete and varies with what executed

**Given** two copies of one Go fixture that differ only in `gate.disabled_steps`, one with the test step disabled and one without,
**When** the gate runs to a verdict against each,
**Then** each summary names every step the toolchain configured, each with exactly one outcome token drawn from {executed, skipped, disabled}; **and** the two summaries differ in the test step's outcome — `disabled` in the first, `executed` in the second — while every other step's outcome line is identical between them.

**Mutant A**: a summary rendered from the configured toolchain table, emitted identically on every run. **Killed**: it produces byte-identical summaries for two runs that executed different work, so the required difference is absent.
**Mutant B**: reporting only the steps that executed, leaving a silently skipped step invisible — the original defect in a new shape. **Killed**: the *configured* set is required, so a skipped step's absence fails it.
**RED**: on the pre-implementation tree both runs emit the empty string; both halves of the assertion fail.

### AC-GTA-002 — the reported duration is measured

**Given** a fixture whose step command is the re-executed test-binary sleeper already used by `internal/hook/quality/gate_timeout_attribution_test.go:19-30`, sleeping a controlled duration D, and a second fixture whose step command returns immediately,
**When** the gate executes each,
**Then** the summary reports the first step's duration as ≥ D **and** the second step's duration as < D.

**Mutant**: printing a constant, the configured timeout budget, or any value not derived from the clock. **Killed**: the two-sided bound. No single constant satisfies both `≥ D` and `< D`, and echoing the budget fails the second bound because both fixtures share one budget.
**RED**: pre-implementation, no duration is reported at all.

### AC-GTA-003 — a skip reason names the observation, not a generic

**Given** a fixture where an optional step's binary is absent from PATH, and a fixture where that same step is turned off through `disabled_steps`,
**When** the gate reaches a verdict against each,
**Then** the two summaries carry distinct reason text for that step, and the config-disabled case's text does not claim the tool was missing.

**Mutant**: one generic `skipped` token for every non-executed step. **Killed**: distinctness is required, and the mutant emits the same token twice.
**RED**: pre-implementation, neither case is reported.

### AC-GTA-004 — a resolved command is reported as resolved

**Given** a Node fixture whose `package.json` declares `scripts."test:run"`, and a second Node fixture that declares only `scripts.test`,
**When** the gate reaches a verdict against each,
**Then** the first summary names `npm run test:run` as the command executed, and the second names the unresolved step form.

**Mutant A**: always printing the configured step name. **Killed** by the first fixture.
**Mutant B**: hardcoding `npm run test:run` for every Node run. **Killed** by the second fixture.
**RED**: pre-implementation, no command is named in either case.

### AC-GTA-005 — `moai gate` emits the summary on a pass

**Given** a project whose gate passes,
**When** `moai gate` runs and exits 0,
**Then** its stderr contains the per-step outcome lines that AC-GTA-001 requires for the detected toolchain.

**Mutant**: emitting any non-empty string on the pass path — a banner, a version line, a bare `ok`. **Killed**: the criterion binds to the outcome lines of a named toolchain, not to output length. An assertion of the form "stderr is non-empty" would pass this mutant and is explicitly rejected as a formulation of this criterion.
**RED**: pre-implementation, `internal/cli/gate.go:79-81` prints nothing because `output` is empty.

### AC-GTA-006 — an existing notice survives alongside the summary

**Given** a fixture where the ast-grep scanner is absent, so that step emits its existing skip notice,
**When** the gate reaches a passing verdict,
**Then** the emitted output contains both that notice's text and the execution summary.

**Mutant**: replacing `passReason` with the summary, dropping the notice. **Killed**: both are required to be present.
**RED**: pre-implementation the notice is emitted alone; the conjunction fails.

### AC-GTA-007 — no unobserved value reaches the summary

**Given** a fixture whose second step fails, aborting the run before the later steps are reached,
**When** the summary is inspected,
**Then** every later step is reported as not reached, none carries an `executed` outcome, and none carries a non-zero duration.

**Mutant**: pre-populating the summary from configuration and overwriting entries as steps complete, leaving a pre-populated `executed` on a step the run never reached. **Killed**: the aborting fixture exposes exactly that residue.
**RED**: pre-implementation, no summary exists to inspect.

---

## §D.2 Milestone 2 — Axis 2: a step's timeout terminates the step

The failing input for this milestone is one fixture, used by AC-GTA-008 and AC-GTA-009 alike: **a step command that spawns a descendant which outlives it and inherits the step's stdout and stderr**. This is the exact shape §A.2 of `spec.md` identifies as blocking `cmd.Wait`, and it is the input that must be observed failing before either criterion counts as met.

### AC-GTA-008 — the gate returns within a bounded grace period, on both platforms

**Given** the orphan-holding fixture above, a step timeout T, and a descendant that sleeps for a duration far exceeding T plus the grace budget G,
**When** the gate runs that step,
**Then** the call returns within T + G as measured by the caller's wall clock — on Unix and on Windows alike; **and** on Windows the reported reason additionally states that descendants of the step may have survived.

**Mutant A**: raising T so the descendant finishes first. **Killed**: the descendant's sleep is specified relative to T + G, so no value of T rescues it.
**Mutant B**: asserting only that the call "eventually returns". **Killed**: the bound is explicit and is the assertion.
**Mutant C**: a Windows build that compiles but never applies the bound. **Killed** only by executing the criterion on Windows.
**Verification note**: `GOOS=windows go vet ./...` proves the platform file compiles and **is not evidence that the behaviour holds**. The Windows half of this criterion is admissible only from the CI Windows matrix run; until that run is observed, the Windows half stands as a gap, not a pass.
**RED**: on the pre-implementation tree this test does not return within T + G — it blocks until the test binary's own `-timeout` fires. That observed hang is the failing input.

### AC-GTA-009 — the descendant is actually terminated (Unix)

**Given** the same fixture, with the descendant's PID recorded,
**When** the step's deadline expires,
**Then** within the grace window a liveness probe of that PID reports the process gone.

**Mutant**: bounding `Wait` alone. This unblocks the parent and satisfies AC-GTA-008 in full while leaving the orphan running forever — which is why AC-GTA-008 and AC-GTA-009 are two criteria and not one. **Killed**: the liveness probe still finds the PID alive.
**RED**: pre-implementation, the descendant survives indefinitely; the probe finds it alive.

### AC-GTA-010 — nothing on the existing timeout or happy path regresses

**Given** the two existing regression tests `TestRunStep_ParentDeadlineIsNotBlamedOnTheStep` (`internal/hook/quality/gate_timeout_attribution_test.go:44`) and `TestRunStep_StepDeadlineStillBlamesTheStep` (`:67`), and a step that writes to both stdout and stderr, exits non-zero, and completes well inside its deadline,
**When** the suite runs against the post-change tree,
**Then** both regression tests pass **and** `git diff` against 294b4b6ab shows no modified or deleted line in that test file; **and** the within-deadline step's captured combined output and derived failure message are byte-identical to the pre-change tree's, with its working-directory resolution unchanged.

**Mutant A**: adjusting the two test bodies until they agree with whatever the new code does. **Killed**: the diff constraint. Additions to the file are permitted; modification and deletion are not.
**Mutant B**: a process-group or pipe change that detaches the child from the configured `cmd.Dir`, reorders the stderr-then-stdout concatenation, or truncates output at the grace boundary. **Killed**: byte-identity plus the working-directory assertion.
**RED**: not applicable — a preservation criterion. Its failing inputs are the mutations named above, each of which the assertions detect.

---

## §D.3 Milestone 3 — Axis 3: concurrent manual runs are serialized

### AC-GTA-011 — two runs do not overlap

**Given** two `moai gate` processes started against one project directory, the first executing a step that runs for a duration D,
**When** the second starts within D,
**Then** the second's first executed step, as timestamped in its own AC-GTA-002 summary, begins after the first run released the lock.

**Mutant**: emitting a "waiting…" line while running anyway. **Killed**: the criterion asserts on the execution timestamps the M1 summary records, not on the presence of a message.
**RED**: on the pre-implementation tree the two runs' execution windows overlap.

### AC-GTA-012 — the waiting notice names the holder

**Given** a gate-run lock held by a process whose PID P is recorded in the lock artifact,
**When** a second run starts and begins waiting,
**Then** its stderr notice contains P.

**Mutant**: a generic "another run is in progress". **Killed**: P is required, and P is knowable only by reading the artifact.
**RED**: pre-implementation, no lock and no notice exist.

### AC-GTA-013 — the wait is bounded and degrades

**Given** a lock held by a live process for longer than the wait budget W,
**When** a second run starts,
**Then** it begins executing its first step at ≥ W and < W plus a stated slack, it returns the gate's own pass/fail verdict, and its output states that it ran unserialized.

**Mutant A**: waiting indefinitely. **Killed** by the upper bound.
**Mutant B**: failing the run when the wait expires. **Killed** by the requirement that the gate's own verdict is returned.
**RED**: pre-implementation, no wait exists, so the lower bound `≥ W` fails immediately.

### AC-GTA-014 — a dead holder does not cost the full budget

**Given** a lock artifact recording a PID that is not alive,
**When** a run starts,
**Then** it acquires the lock within a bound far below W, and emits no degradation report.

**Mutant**: treating every held artifact as live and waiting out W before degrading. **Killed**: the acquisition bound is stated as far below W, and the degradation report must be absent.
**RED**: pre-implementation, no artifact and no clear path exist.

### AC-GTA-015 — a lock failure never fails the run

**Given** a project directory in which the lock artifact's directory cannot be created,
**When** `moai gate` runs,
**Then** it executes the gate, returns the gate's own verdict, and reports that the lock was unavailable.

**Mutant**: propagating the lock error as the command's error. **Killed**: the verdict must be the gate's own.
**RED**: the failing input is the read-only lock directory, which must be observed producing the gate's verdict rather than an error. Pre-implementation there is no lock to fail, so this RED is observed against the M3 tree before the error-path handling is added.

### AC-GTA-016 — degradation is one-way

**Given** a run that has degraded to unserialized execution at W,
**When** the original holder releases the lock while the degraded run is still executing,
**Then** a third run started at that moment acquires the lock immediately.

**Mutant**: opportunistically re-acquiring after degrading. **Killed**: the third run's immediate acquisition is impossible if the degraded run took the lock.
**RED**: pre-implementation, no lock exists and the third run's acquisition is trivially immediate, so this criterion is meaningful only against the M3 tree and must be run there.

---

## §D.4 Severity and blocking status

| Criterion | Severity | Blocks milestone close |
|-----------|----------|------------------------|
| AC-GTA-001 … 007 | blocking | M1 |
| AC-GTA-008 (Unix half) , 009, 010 | blocking | M2 |
| AC-GTA-008 (Windows half) | blocking, evidence from the CI Windows matrix only | M2 |
| AC-GTA-011 … 016 | blocking | M3 |

## §D.5 Definition of Done

A milestone is done when all of the following are observed, each with the command run and its output cited:

1. Every blocking criterion for that milestone passes, and each one's **RED** was observed failing first.
2. `go test ./internal/hook/quality/...` passes. `go test -timeout 1200s ./internal/cli/...` passes.
3. `go vet ./internal/hook/quality/... ./internal/cli/...` is clean.
4. `GOOS=windows go vet ./...` and `GOOS=linux go vet ./...` compile — recorded as compilation evidence only, never as behavioural evidence.
5. The full suite verdict comes from CI against the pull-request head, not from a local run. The local suite is never run in full (`go test ./...` is prohibited here).
6. No verification recipe in the milestone spawns background load, and every process a recipe starts is bounded by `t.Cleanup` or an external `timeout` wrapper — not by a trailing `kill` line.
