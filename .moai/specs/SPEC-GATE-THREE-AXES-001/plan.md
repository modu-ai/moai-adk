# SPEC-GATE-THREE-AXES-001 — Implementation Plan

Sections are ordered by how likely each decision is to change under review. The data-model and substrate choices come first because they are the ones worth arguing about; the milestone sequencing and the mechanical steps follow.

All citations are at tree SHA **294b4b6ab** in `.claude/worktrees/t235`.

---

## §A Decisions that shape everything downstream

### A.1 The execution summary is a value the run builds, not a string steps append to

`Run` currently threads a single `passReason` string through five call sites (`internal/hook/quality/gate.go:340-402`), joined by `appendReason` (`internal/hook/quality/gate_typecheck.go:150-159`). Extending that string is the smaller diff and the wrong shape: a string cannot carry a duration, cannot distinguish "configured but never reached" from "skipped", and gives the tests nothing to assert on except substring matching — which is exactly the formulation `acceptance.md` rejects for AC-GTA-005.

The proposal is a per-step record accumulated by `Run`, rendered to text at the boundary:

- one record per **configured** step, seeded from the toolchain before execution begins, so a step the run never reaches is reportable as not reached (AC-GTA-007);
- each record carries the outcome, the measured elapsed time, the executed command line (`step.binary` plus `step.args`, **not** `gateStep.name` — the two diverge on two of the three Node resolution tiers), the step's label as its identity, and the existing notice text where a step produced one;
- `runStep` returns the observed facts to `executeStep`, which returns them to `Run`. Today `runStep` collapses the pass path to `(true, "")` at `gate.go:1020-1022`; that collapse is the change's centre of gravity.

**Reversibility note.** This changes an internal signature that `Run`, `executeStep`, and `runStep` share, and `Run`'s public `(bool, string)` return is consumed by both `internal/cli/gate.go:69` and the PreToolUse hook path. The public shape stays `(bool, string)` — the summary is rendered into the existing string — so no caller changes. If review prefers a structured public return instead, that is a wider change touching the hook path and should be settled before M1 starts.

**Alternative considered and rejected**: keeping `passReason` and appending formatted lines to it. Rejected because AC-GTA-007 requires reporting steps the run never reached, and an append-only accumulator has no entry for a step that never ran.

### A.2 Termination needs two mechanisms, not one

§A.2 of `spec.md` identifies two distinct failures behind one symptom:

1. `cmd.Wait` blocks past the deadline because a descendant holds the step's inherited output pipe;
2. the descendant itself is never signalled.

They need different remedies, and each remedy alone leaves the other failure standing. This is why `acceptance.md` splits them into AC-GTA-008 and AC-GTA-009.

- **For (1): `exec.Cmd.WaitDelay`.** The standard library's answer to exactly this case — it bounds how long `Wait` waits after the context is cancelled and closes the I/O pipes. Go 1.26.4 (`go.mod:3`); `grep -rn 'WaitDelay' internal/` returns 0, so nothing in the repository uses it yet. Per the simplicity ladder, the standard library is reached for before new code.
- **For (2): `Setpgid` at spawn plus a group signal at kill time**, so the whole tree the step started receives the termination. `golang.org/x/sys v0.47.0` (`go.mod:29`) is already a dependency and already carries the Unix lock implementation at `internal/spec/lock_unix.go`.

**Platform split.** `Setpgid` is Unix-only. The repository's convention is a build-tagged pair behind a small interface, with no naked syscall in the shared body: `internal/spec/lock.go:36-38` declares the interface, `internal/spec/lock_unix.go` and `internal/spec/lock_windows.go` implement it, and `internal/spec/lock.go:8-10` states the rule. This SPEC follows that pattern rather than inventing one — currently `ls internal/hook/quality/ | grep -c '_unix\|_windows'` is 0, so the pair is new to this package.

**Windows.** No process-group primitive is applied. `WaitDelay` still bounds the return (AC-GTA-008 Windows half), and the reported reason says descendants may survive rather than implying they were killed. A Windows job object would kill the tree properly and is a reasonable follow-up; it is not in this SPEC's scope because it is a second platform mechanism with its own failure modes and would double M2's surface.

**Constraint that binds this section.** t218's attribution fix lives in the same frame — the `parentBinds` branch at `internal/hook/quality/gate.go:996-1002`, guarded by two regression tests at `gate_timeout_attribution_test.go:44` and `:67`. Those tests read the reason string produced on the timeout path, and both mechanisms above execute on that same path. AC-GTA-010 forbids modifying either test body; any new text must be added around the existing reason, not in place of it.

### A.3 The lock substrate is `internal/spec/lock.go`'s pattern, not `internal/lockfile`

The card's premise that `internal/kanban/board_lock.go` consumes `internal/lockfile` is false (`spec.md` §A.4 item 2). That correction is not cosmetic — it removes the only precedent that would have justified `internal/lockfile` here, and inspecting the package directly confirms it is the wrong substrate:

| | `internal/lockfile` | `internal/spec/lock.go` pattern |
|---|---|---|
| API | `Lock` / `Unlock` only (`internal/lockfile/lockfile_unix.go:23-30`) | acquire with a contention sentinel, `Release` |
| Contention | `syscall.Flock(fd, LOCK_EX)` — **blocking**, no `LOCK_NB` | `LOCK_EX\|LOCK_NB`, returns `ErrSpecCloseLockHeld` immediately (`internal/spec/lock_unix.go:36-38`) |
| Windows | in-process mutex, documented as offering **no** cross-process protection | atomic `O_CREATE\|O_EXCL`, genuinely cross-process (`internal/spec/lock_windows.go:80-102`) |
| Holder identity | none | Windows impl only (`lock_windows.go:86` writes `pid=%d`); the Unix path opens `O_CREAT\|O_RDWR` and flocks, recording nothing (`lock_unix.go:31-40`). `board_lock.go` layers identity on **both** platforms via `BoardLockOwner{PID, CreatedAt}` (`board_lock.go:50-53`) |

A bounded wait cannot be built on a blocking primitive without a goroutine and a channel race around a syscall that has no cancellation — and on Windows it would serialize nothing at all across processes, which is precisely where two `moai gate` shells collide. The try-semantics substrate gives a bounded wait its natural form: attempt, sleep briefly, attempt again, until the budget expires.

**Chosen**: the `internal/spec/lock.go` pattern, with `internal/kanban/board_lock.go` as the closer precedent — it already layers holder identity (`board_lock.go:50-53`) and a bounded stale clear with a changed-hands abort (`board_lock.go:32-41`) on top of that pattern, which is exactly what REQ-GTA-014 needs. The changed-hands guard matters: a stale clear that unlinks a lock reacquired between inspection and removal would admit a second holder.

**Not chosen**: a new generic lock package. Two precedents already exist; a third abstraction earns nothing.

### A.4 Where the wait budget and grace budget live

Both are per-project knobs with sensible defaults, and both belong in `gate.yaml` beside the existing timeouts (`.moai/config/sections/gate.yaml:16-22`), flowing through `config.GateConfig` (`internal/config/types.go:764-782`) → `mapConfigGateToQuality` (`internal/cli/gate.go:136-170`) → `quality.GateConfig` (`internal/hook/quality/gate.go:20`), with defaults in `internal/config/defaults.go` and `quality.DefaultGateConfig`.

Open for review: whether the grace budget (A.2) is configurable at all, or a fixed small constant. A configurable grace is one more knob whose wrong value reintroduces the defect; a constant cannot be tuned when a legitimately slow shutdown needs longer. The plan proposes a constant for the grace and a configured value for the axis-3 wait, on the grounds that the first is a safety bound and the second is a policy.

---

## §B Milestones

Ordered so the self-contained change lands first and each execution-model change is independently verifiable.

### M1 — Axis 1: the run reports what it ran

Satisfies REQ-GTA-001 … REQ-GTA-007. Verified by AC-GTA-001 … AC-GTA-007.

1. Introduce the per-step record and the accumulator in `Run` (A.1), seeded from the toolchain before the first step executes.
2. Return the observed facts from `runStep` — replacing the `(true, "")` collapse at `gate.go:1020-1022` — and thread them through `executeStep`. Each of `executeStep`'s five early-return skip paths must report *which* skip it took, per REQ-GTA-003, which now enumerates all five: `DisabledSteps` (`gate.go:778-780`), optional-binary `LookPath` (`:782-786`), `configFiles` absent (`:787-789`), `changedExts` no staged match (`:793-801`), `sourceExts` no source (`:806-816`).

   The `changedExts` path is conditional in a way the fixture must respect: it skips only when the staged-file lookup succeeded, and runs the step conservatively when `staged` is nil (`gate.go:796-800`) — which is what happens outside a git repository. Its fixture is therefore an initialized repository with at least one staged non-matching file, not a bare `t.TempDir()`. Carried in AC-GTA-003 as the fixture (d) caveat.
3. Record the **executed command line** for the Node test step (`gate.go:676-699`), per REQ-GTA-004. That resolution has **three** tiers, not two — `test:run` substitution, `nodeNonWatchFlag` flag-appending (`gate.go:729-744`, keyed on `nodeScriptWatchProne` at `:752-771`), and the unchanged fallback. AC-GTA-004 carries one fixture per tier with its exact `package.json` content.

   Take the command from `step.binary` + `step.args`, **not** from `gateStep.name`. `executeStep` passes both to `runStep` (`gate.go:818`) but only the latter pair reaches `exec.CommandContext` (`:1006`); the label diverges from argv on tiers (ii) and (iii), where `-- --passWithNoTests` is on the command line and absent from the label. Reporting the label there would drop the flag that decides whether an empty suite counts as a pass — output whose content is not the execution result, which is the defect this SPEC exists to remove.
4. Fix the dropped test-step pass value at `gate.go:397-399`.
5. Render the summary into `Run`'s existing string return, folding in — not replacing — the existing notice text (REQ-GTA-006).
6. `internal/cli/gate.go:79-81` already prints a non-empty pass-path output, so no CLI change is expected. Confirm rather than assume.

Touches: `internal/hook/quality/gate.go`, `internal/hook/quality/gate_typecheck.go`. No config change.

### M2 — Axis 2: the step timeout terminates the step

Satisfies REQ-GTA-008 … REQ-GTA-011. Verified by AC-GTA-008 … AC-GTA-010.

1. Add the build-tagged platform pair to `internal/hook/quality/` following `internal/spec/lock_{unix,windows}.go`: a small interface in the shared body, `Setpgid` at spawn and a group signal at kill on Unix, a no-op on Windows.
2. Set `WaitDelay` on the step command (A.2 mechanism 1).
3. Extend the timeout reason to say what was terminated, without disturbing the `parentBinds` branch at `gate.go:996-1002`.
4. Build the orphan-holding fixture (`acceptance.md` §D.2) by extending the existing sleeper helper at `gate_timeout_attribution_test.go:19-30` — a new helper mode that spawns a grandchild and exits, leaving the grandchild holding the inherited pipe. Additions to that file only; no existing line modified (AC-GTA-010).

Touches: `internal/hook/quality/gate.go`, two new platform files, one new test file, additions to `gate_timeout_attribution_test.go`.

**Hazard.** The fixture starts processes. Every one is registered with `t.Cleanup` or wrapped in an external `timeout`; a trailing `kill` is not cleanup, because every early return skips it. The sleeper's duration is bounded so a leaked process cannot outlive the suite by more than that bound. A library-level fork is out of the question here — `os.Executable()` under `go test` is the test binary, and re-executing it without a narrowing `-test.run` re-runs the suite.

### M3 — Axis 3: manual runs are serialized

Satisfies REQ-GTA-012 … REQ-GTA-016. Verified by AC-GTA-011 … AC-GTA-016.

1. Add the gate-run lock following `internal/kanban/board_lock.go`: acquire with a contention sentinel, holder identity recorded in the artifact, bounded stale clear with the changed-hands abort.
2. Wrap `runGate` (`internal/cli/gate.go:58-83`) in acquire → run → release, with the bounded wait loop and the one-way degradation flag.
3. Emit the waiting notice (REQ-GTA-012) and the degradation report (REQ-GTA-013) on stderr, alongside M1's summary.
4. Thread the wait budget through the config chain (A.4).
5. Every lock error path returns the gate's own verdict (REQ-GTA-015). The lock never decides the exit code.

Touches: `internal/cli/gate.go`, a new lock file (package placement to be settled in run-phase: alongside the CLI, or a small package of its own), `internal/config/types.go`, `internal/config/defaults.go`, `.moai/config/sections/gate.yaml` and its template mirror.

**Template-First.** `gate.yaml` ships to user projects. Any key added in M3 is added at `internal/template/templates/.moai/config/sections/gate.yaml` first, then `make build`, then the local copy — and the neutrality audit (`go test ./internal/template/...`) runs before the milestone closes. Template content carries no SPEC ID, no REQ token, and no internal date.

---

## §C Run-phase verification constraints

These bind every verification recipe in this SPEC. They are constraints, not suggestions; a recipe that violates one is rejected regardless of what it reports.

1. **Never run the full local suite.** `go test ./...` is prohibited. Affected packages only:
   - `go test ./internal/hook/quality/...`
   - `go test -timeout 1200s ./internal/cli/...`
   - `go test ./internal/template/...` (M3 only, for the neutrality audit)
2. **`internal/cli` needs `-timeout 1200s`.** That package alone measures roughly 336s; a 300s ceiling fails a tree that is fine.
3. **Never spawn background load in a verification recipe.** No contention harness, no parallel load generator. Where a test starts a process, the process is bounded by `t.Cleanup` or an external `timeout` wrapper.
4. **The full-suite verdict comes from CI**, against the pull-request head, in a clean environment and across the platform matrix. A local green is an early signal and is never cited as the full-suite result.
5. **`GOOS=<os> go vet` is compilation evidence only.** It says the platform file builds. It says nothing about behaviour, and no behavioural claim may cite it.
6. **A pipeline's exit code is the last command's.** `find … | wc -l` reports rc=0 with a count of 0 even when `find` died. Any counting recipe either sets `pipefail` or checks the producing command's status separately.
7. **Environment scrubbing for local-only failures.** A failure that reproduces locally but not in CI is an environment hypothesis until the session's own variables are scrubbed. The scrub and the command travel as one compound invocation — `unset VAR1 VAR2 && go test ./internal/hook/quality/...` — because each Bash call is a fresh process.

---

## §D Risks

| Risk | Where it bites | Mitigation |
|------|----------------|------------|
| The summary changes what callers of `Run` see | `internal/cli/gate.go:69`, the PreToolUse hook path | Public return shape unchanged (A.1). Both call sites read and print; neither parses. Confirm by reading them, not by assuming. |
| M2 disturbs t218's attribution fix | `gate.go:996-1002` | AC-GTA-010 forbids modifying either regression test; both must pass unmodified. |
| A group kill reaches further than intended | any step | The group is created per step at spawn, so it contains only that step's descendants. AC-GTA-010 asserts the within-deadline path is byte-identical. |
| The stale clear unlinks a live lock | M3 | The changed-hands abort from `board_lock.go:32-41` is carried over, not reinvented. |
| M2's fixture leaks processes onto the developer's machine | local runs | Bounded sleeps plus `t.Cleanup`-registered kills, per §C item 3. |
| Overlap with card t233 / issue #1631 | `executeStep` | t233 is not yet dispatched. If it starts before M1 lands, the two must be sequenced rather than merged — both edit the same frame. |

---

## §E Cross-references

- `.moai/reports/t235/premise-verification.md` — the measured premise report this SPEC is built on (tree SHA 294b4b6ab).
- `internal/spec/lock.go`, `lock_unix.go`, `lock_windows.go` — the platform-split and try-semantics precedent.
- `internal/kanban/board_lock.go` — holder identity and bounded stale clear on that same pattern.
- `internal/hook/quality/gate_timeout_attribution_test.go` — the t218 regression tests M2 must leave intact.
- Card t233 / issue #1631 — the lint axis, explicitly out of scope.
