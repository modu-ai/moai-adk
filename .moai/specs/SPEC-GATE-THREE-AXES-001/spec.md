---
id: SPEC-GATE-THREE-AXES-001
title: "Quality gate: report what ran, enforce the step timeout, serialize manual runs"
version: "0.1.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: GOOS
priority: P1
phase: "v3.1.4 target"
module: "internal/hook/quality, internal/cli"
lifecycle: spec-anchored
tags: "gate, quality, timeout, process-group, serialization, observability"
tier: M
era: V3R6
---

# SPEC-GATE-THREE-AXES-001 — Quality gate: report what ran, enforce the step timeout, serialize manual runs

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-08-24 | 0.1.0 | Plan-phase authoring. Three axes carried from card t235 / issue #1639, with the premise corrections from `.moai/reports/t235/premise-verification.md` folded in. | GOOS |

## §A Context — measured ground truth

Every citation below was re-read in this worktree at tree SHA **294b4b6ab** (`.claude/worktrees/t235`, branch `WT-gate-three-axes`). The card's own line numbers were superseded by the premise-verification report and are not reproduced here.

### A.1 A passing gate run emits 0 bytes

- `internal/hook/quality/gate.go:1020-1022` — `runStep` returns `(true, "")` when `cmd.Run()` returns nil. The child's captured stdout and stderr (`gate.go:1016-1018`) are discarded on the success path.
- `internal/hook/quality/gate.go:322-402` — `Run` accumulates a `passReason` across the vet, typecheck, lint, and ast-grep axes via `appendReason` (`internal/hook/quality/gate_typecheck.go:150-159`). That accumulator carries **skip notices only**, because `runStep` supplies nothing else on a pass.
- `internal/hook/quality/gate.go:397-399` — the test step is the one axis whose pass-side value is dropped outright: `if ok, out := g.executeStep(ctx, testStep, g.config.TestTimeout); !ok { return false, out }` never reads `out` on the success branch.
- `internal/cli/gate.go:69-81` — the CLI prints only when `output != ""`.
- `internal/hook/quality/gate.go:676-700` — `resolveNodeTestStep` tier (i) replaces the step with `npm run test:run`, delegating the verdict to a script the gate does not author.

Net effect: a run whose test step replayed a build cache and a run that genuinely executed the suite are byte-identical silence at the user's terminal. Nothing in the emitted output distinguishes them, and nothing names the command the verdict was actually delegated to.

### A.2 The step timeout is not enforced

- `internal/hook/quality/gate.go:1006` — `exec.CommandContext(stepCtx, name, args...)`. No `SysProcAttr` is set anywhere in the package: `grep -rn 'Setpgid\|Killpg\|killpg\|SysProcAttr' internal/hook/quality/` returns 0 lines (rc=1).
- `internal/hook/quality/gate.go:1016-1018` — `cmd.Stdout` and `cmd.Stderr` are `*bytes.Buffer`, not `*os.File`. `os/exec` therefore allocates OS pipes and copying goroutines, and `cmd.Wait` waits for those pipes to reach EOF **after** the process exits.
- `internal/hook/quality/gate.go:1020` — `err := cmd.Run()`. No `WaitDelay` is set; `grep -rn 'WaitDelay' internal/` returns 0 lines.

The context deadline signals only the direct child. A descendant that outlives it inherits the write end of the step's output pipe, so the pipe never reaches EOF and `Wait` blocks past the deadline that was supposed to bound it. The outer net at `internal/cli/gate.go:65` (`context.WithTimeout(..., 10*time.Minute)`) is the same kind of signal and cannot cut a blocked `Wait` either.

Consequence: the timeout the gate reports and the moment the gate actually returns are two different things, and no configured budget bounds the second.

### A.3 Manual `moai gate` runs are not serialized

- `internal/cli/gate.go` (full 215 lines read) — no lock is acquired. `grep -n 'Acquire.*Lock\|lockfile\.' internal/cli/gate.go` returns 0 lines.
- Two concurrent `moai gate` invocations in one checkout each run the full toolchain against the same working tree, competing for the same build cache and the same machine.

### A.4 Premise corrections recorded

1. **The card's line numbers are stale.** Every axis-1 and axis-2 site sits roughly 20-24 lines below where the card places it. The citations in §A above are the measured ones, taken at 294b4b6ab.
2. **`internal/kanban/board_lock.go` is not a consumer of `internal/lockfile`.** The card lists it as one. It is not: the only importers are `internal/cli/settings.go:12`, `internal/cli/glm_tools.go:32`, and `internal/cli/taskledger/taskledger.go:16`. `internal/kanban/board_lock.go:6-11` states explicitly that it reuses the `internal/spec/lock.go` pattern and that "internal/lockfile's in-process mutex is neither used nor upgraded". The substrate consequence is worked in `plan.md` §D.3.
3. **Axis 1 is stronger than the card states.** The card describes `passReason` as carrying skip notices; the measured behaviour adds that the test step's pass-side output is discarded before it ever reaches the accumulator.
4. **The 915s survival figure is not reproduced here.** It is a load-dependent downstream observation. This SPEC rests on the structural facts in §A.2, which were measured directly; it makes no claim about a specific survival duration.

## §B Requirements (GEARS)

### Axis 1 — a completed run reports what it ran

**REQ-GTA-001** (Ubiquitous)
The quality gate shall produce, for every run that reaches a verdict, an execution summary listing each configured step together with the outcome that step actually reached: executed, skipped, or disabled.

**REQ-GTA-002** (event-driven)
When a step is executed, the quality gate shall record in the summary the wall-clock duration measured for that step's own execution.

**REQ-GTA-003** (state-driven)
While a step is not executed, the quality gate shall record in the summary the observed reason it was not executed, distinguishing each of the five paths by which the gate skips a step: (a) the step is turned off by configuration; (b) the step is optional and its binary is absent from PATH; (c) none of the step's named config files exist; (d) no staged file matches the step's declared extensions; (e) the project holds no source file of the step's declared source extensions.

**REQ-GTA-004** (capability gate)
Where a step's command is resolved to a substitute before execution, the quality gate shall name in the summary the command that was actually executed rather than the step as configured. "The command that was actually executed" means the full command line — the step's binary together with its arguments, as handed to the process launcher — and not the step's display label, which is a separate value that coincides with the command line only by accident.

> Two fields, not one. A step's **label** is its identity: it names which step a summary line is about, it is the key `gate.disabled_steps` is written against, and REQ-GTA-001's per-step listing is stated in terms of it. A step's **command** is what ran. REQ-GTA-004 binds the command field only. Wherever the summary reports a command — this requirement is the only place it does — it reports the executed command line.

**REQ-GTA-005** (Ubiquitous)
`moai gate` shall emit the execution summary on a passing run, so that a passing run is never indistinguishable from a run that did nothing.

**REQ-GTA-006** (unwanted)
The quality gate shall not drop an existing non-blocking notice when it emits the execution summary.

**REQ-GTA-007** (unwanted)
The quality gate shall not report a step outcome, duration, or resolved command that was not observed during the run being reported.

### Axis 2 — a step's timeout terminates the step

**REQ-GTA-008** (event-driven)
When a step's own deadline expires, the quality gate shall return a verdict within a bounded grace period, whether or not a descendant of the step still holds the step's inherited output stream.

**REQ-GTA-009** (event-driven, capability gate — compound)
When a step's deadline expires, and where the platform provides a process-group primitive, the quality gate shall signal termination to the step's entire process group rather than to the direct child alone; where the platform provides no such primitive, the quality gate shall instead report that descendants of the step may have survived.

**REQ-GTA-010** (unwanted)
The quality gate shall not alter which budget a timeout is attributed to: a step killed because the caller's budget ran out shall still be reported against the caller's budget, and a step killed by its own budget shall still be reported against its own.

**REQ-GTA-011** (unwanted)
The quality gate shall not change a within-deadline step's exit status, captured output, or working directory as a side effect of the termination machinery.

### Axis 3 — concurrent manual runs are serialized

**REQ-GTA-012** (state-driven — compound)
While another `moai gate` run in the same project holds the gate-run lock, a starting run shall wait for the lock for a bounded budget before doing anything else, and shall emit a notice naming the holder it is waiting on.

**REQ-GTA-013** (event-driven)
When the bounded wait expires without the lock being acquired, the run shall proceed without serialization and shall report that it did so.

**REQ-GTA-014** (capability gate)
Where the lock artifact records a holder process that is no longer alive, a starting run shall clear the artifact and acquire the lock rather than waiting out the full budget.

**REQ-GTA-015** (unwanted)
`moai gate` shall not fail a run, and shall not block without bound, because the gate-run lock could not be acquired.

**REQ-GTA-016** (unwanted)
A run that has degraded to unserialized execution shall not re-acquire the lock for the remainder of that run; the degradation is one-way.

## §C Acceptance criteria

Enumerated in `acceptance.md`, together with the mutant analysis required for each criterion.

## §D Exclusions

### Out of Scope — the lint axis

- Any change to how lint steps are selected, resolved, or executed. Card t233 / issue #1631 touches the same `executeStep` frame and is not yet dispatched; overlapping edits here would collide with it.
- Any change to `RunAstGrepGateV2` or the ast-grep rule set.

### Out of Scope — gate policy and verdicts

- Which steps run for which language, and the contents of the per-language toolchain table.
- Whether a failure blocks a commit (`BlockOnError` / `WarnOnlyMode` semantics).
- The per-step timeout *values* in `gate.yaml`. This SPEC makes the configured value enforceable; it does not retune it.
- The pytest exit-5 and dotnet-restore special cases.

### Out of Scope — serialization beyond the manual CLI

- Serializing the PreToolUse hook path or any other in-process caller of `QualityGate.Run`. Axis 3 is scoped to the `moai gate` CLI entry point, where two runs are two separate OS processes.
- Cross-machine or cross-checkout coordination. The lock is per-project-directory.
- Queueing, fairness, or priority among waiting runs. The wait is a bounded wait, not a queue.

### Out of Scope — output format as a contract

- Making the execution summary a machine-parseable format that third parties may depend on. It is human-facing diagnostic output. A `--json` surface, if wanted, is a separate SPEC.

## §E Traceability

Per-requirement, not per-range. Two axes are non-1:1 (axis 2 is 4 REQs → 3 ACs, axis 3 is 5 REQs → 6 ACs), so a range-level table would leave a run-phase reader to re-derive which criterion discharges which requirement. Each AC additionally carries a `Verifies:` line naming the same mapping from the other side.

| REQ | Discharged by | Axis | Milestone |
|-----|---------------|------|-----------|
| REQ-GTA-001 | AC-GTA-001 | 1 | M1 |
| REQ-GTA-002 | AC-GTA-002 | 1 | M1 |
| REQ-GTA-003 | AC-GTA-003 (five fixtures, one per skip path) | 1 | M1 |
| REQ-GTA-004 | AC-GTA-004 (three fixtures, one per resolution tier) | 1 | M1 |
| REQ-GTA-005 | AC-GTA-005 | 1 | M1 |
| REQ-GTA-006 | AC-GTA-006 | 1 | M1 |
| REQ-GTA-007 | AC-GTA-007 | 1 | M1 |
| REQ-GTA-008 | AC-GTA-008 | 2 | M2 |
| REQ-GTA-009 | AC-GTA-009 (group-signal branch) + AC-GTA-008 Windows half (report branch) | 2 | M2 |
| REQ-GTA-010 | AC-GTA-010, first half | 2 | M2 |
| REQ-GTA-011 | AC-GTA-010, second half | 2 | M2 |
| REQ-GTA-012 | AC-GTA-011 (non-overlap conjunct) + AC-GTA-012 (notice conjunct) | 3 | M3 |
| REQ-GTA-013 | AC-GTA-013 | 3 | M3 |
| REQ-GTA-014 | AC-GTA-014 | 3 | M3 |
| REQ-GTA-015 | AC-GTA-015 ("shall not fail") + AC-GTA-013's upper bound ("shall not block without bound") | 3 | M3 |
| REQ-GTA-016 | AC-GTA-016 | 3 | M3 |

No requirement is uncovered and no criterion is orphaned. AC-GTA-010 is the one genuine merge; its two halves stay separable so a failure attributes to REQ-GTA-010 or REQ-GTA-011 rather than to the pair.

Tier M budget: 16 requirements and 16 acceptance criteria. This SPEC carries exactly 16 of each — at the ceiling, not over it. Two requirements are stated as GEARS compound clauses (REQ-GTA-009, REQ-GTA-012) rather than split, which is what keeps the count at budget without dropping any obligation.
