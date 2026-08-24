# SPEC-GATE-THREE-AXES-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- Tree measured: **294b4b6ab** (`.claude/worktrees/t235`, branch `WT-gate-three-axes`). Every file:line citation in `spec.md` §A and `plan.md` was re-read at this SHA.
- SPEC ID regex check executed as Bash against the `internal/spec/lint.go` pattern — output: `PASS`.
- Duplicate check: no existing `.moai/specs/SPEC-GATE-THREE-AXES-001`, no prior reference in the SPEC catalogue.
- Grep-AC baselines measured at 294b4b6ab, all **0**: `Setpgid|Killpg|killpg|SysProcAttr` in `internal/hook/quality/` (rc=1); `WaitDelay` in `internal/`; lock acquisition in `internal/cli/gate.go`.
- Premise corrections recorded in `spec.md` §A.4 (4 items), with the substrate consequence worked in `plan.md` §A.3.
- Tier M budget: **16 requirements, 16 acceptance criteria** — exactly at the ceiling. Two requirements (REQ-GTA-009, REQ-GTA-012) are GEARS compound clauses rather than splits, which is what holds the count at budget without dropping an obligation.
- `moai spec lint .moai/specs/SPEC-GATE-THREE-AXES-001/spec.md` → `✓ No findings — all SPEC documents are valid`.
- Line-number citations were re-verified against the tree after authoring; four stale spans (the `runStep` success branch, the captured-output buffers, the `cmd.Run` call, and the dropped test-step value) were corrected to the measured ones.
- `era: V3R6` is pinned explicitly in frontmatter. Without it, auto-detection fires H-3 (`§E.2` present, `sync_commit_sha` absent — the normal state of a plan-phase SPEC) and classifies the SPEC as V3R5, which would grandfather-protect it out of modern-era drift detection. Measured before and after: `moai spec audit --filter-spec SPEC-GATE-THREE-AXES-001` reported `Grandfathered: 1 … [INFO] (V3R5) EraAutoDetected` before the pin, and `Modern-era clean: 1 / Drift findings: 0` after.
- Plan audit iter-1: **PASS-WITH-DEBT 0.85** (Tier M threshold 0.80), MUST-PASS 7/7, audited at `10e252834`. Report: `.moai/reports/t235/plan-audit-iter1.md`.

### iter-2 defect discharge

Three blocking defects fixed; three optional ones applied because each was cheap and none costs budget. Re-audit not run (lane owns it).

| Defect | Severity | Fix |
|--------|----------|-----|
| D1 — REQ-GTA-003 scope contradiction (two-way REQ vs five-way plan) | major, blocking | **Widened the requirement, did not narrow the plan** (lead's decision). REQ-GTA-003 now enumerates all five skip paths (a)-(e) including `changedExts`, which the earlier text never named. AC-GTA-003 rewritten from two fixtures to five — one per path, with mutual distinctness required — and the collapse-into-one-reason mutant it previously admitted is now named explicitly as Mutant B. |
| D2 — AC-GTA-004's second fixture underdetermined | major, blocking | `resolveNodeTestStep` has three branches; "declares only `scripts.test`" admitted two of them, so a natural fixture (`"test": "vitest"`) would fail against a correct implementation. Replaced with a three-row table naming each fixture's exact `package.json` content, the branch it lands in, and the expected command. Added Mutant C (collapsing tiers (i) and (ii)), which fixture B now kills. |
| D3 — `disabled_steps` polarity inverted and unnamed | minor, blocking | Polarity stated in both Givens: an entry whose value is **FALSE** skips the step. AC-GTA-001's fixture is now `{"go test": false}` explicitly, with the source citations (`gate.go:778-780`, `types.go:775-779`, `cli/gate.go:150-152`) and the failure mode the intuitive polarity would have produced. |
| D4 — range-level REQ→AC mapping | minor, optional | Applied. `spec.md` §E is now per-requirement, and every AC carries a `Verifies:` line naming the same mapping from the other side. |
| D5 — overstated holder-identity row | minor, optional | Applied. `plan.md` §A.3 now qualifies the row: PID is recorded by the Windows impl only; `board_lock.go` layers identity on both platforms. |
| D6 — AC-GTA-008's RED is an unbounded hang | minor, optional | Applied. Added a mandatory RED-isolation clause: observed with a narrowing `-test.run` and an explicit `-test.timeout`, never inside a package run. |

Budget after the D1 widening: **16 requirements, 16 acceptance criteria** — unchanged. D1 strengthened an existing requirement and an existing criterion rather than adding either, so the Tier M ceiling still binds exactly.

Two nuances discovered while building the D1 fixtures, carried into both artifacts so run-phase does not rediscover them: the `changedExts` path skips only when the staged-file lookup succeeded and runs conservatively when `staged` is nil (`gate.go:796-800`), so its fixture must be a real git repository with a staged non-matching file; and `nodeScriptWatchProne` treats a bare `vitest` first token as watch-prone (`gate.go:752-771`), which is what makes the D2 fixture choice load-bearing.

Post-fix verification: `moai spec lint .moai/specs/SPEC-GATE-THREE-AXES-001/spec.md` → `✓ No findings`; `moai spec audit --filter-spec SPEC-GATE-THREE-AXES-001` → modern-era clean 1, drift 0.

### iter-2 audit and D7-D9 discharge

Plan audit iter-2: **PASS-WITH-DEBT 0.92** (+0.07 from 0.85, monotonic), MUST-PASS 7/7, all six iter-1 defects verified resolved against source, traceability 0.80 → 0.95. Audited at `f8186b172`. Report: `.moai/reports/t235/plan-audit-iter2.md`. Iteration 2 of the Tier M ceiling of 2 — no iteration-3 audit follows; the lead verifies this discharge directly.

| Defect | Severity | Fix |
|--------|----------|-----|
| D7 — AC-GTA-004 asserted `gateStep.name`, REQ-GTA-004 requires the command actually executed | major, blocking | **Expected values are now the full argv** (lead's ruling), not the label. Verified the divergence at source: `executeStep` ends `return g.runStep(ctx, step.name, timeout, step.binary, step.args...)` (`gate.go:818`) — the label travels as `runStep`'s `stepName` and reaches only the reason strings, while `binary`+`args` become `exec.CommandContext(stepCtx, name, args...)` (`:1006`). Tier (i) argv `npm run test:run` coincides with its label; tier (ii) argv is `npm test -- --passWithNoTests --run` against label `npm test --run`; tier (iii) argv is `npm test -- --passWithNoTests` against label `npm test`. The AC table now carries both columns, with the label column explicitly marked **not asserted** so Mutant D has something to fail against. New Mutant D: reporting `name` in place of argv — it passes fixture A, where the two coincide, and drops `-- --passWithNoTests` on the other two. |
| D8 — four off-by-one citation boundaries | minor, optional | Applied. `gate.go:778-815` → `:778-816`; tier (i) `:683-689` → `:684-689`; tier (ii) `:690-696` → `:691-696`; pass-through `:697` → `:698`. |
| D9 — AC-GTA-003 omitted the guards' sequential ordering | minor, optional | Applied. Added a guard-ordering caveat covering all five fixtures: the paths are ordered guards at `:778`/`:782`/`:787`/`:793`/`:806`, first match wins, so each fixture's step must clear every preceding guard. Named the overlap that makes it real (25 steps declare `optional:`, 11 declare `configFiles:`) and the failure shape — the fixture stays green while the path it was written for goes unverified. |

**Why argv rather than narrowing the requirement.** Tiers (ii) and (iii) both pass `--passWithNoTests`, which makes an empty suite report success. Whether that flag was on the command line therefore decides what a "pass" on that summary line means. A summary reading `npm test` while `npm test -- --passWithNoTests` ran is output whose content is not the execution result — the card's named axis-1 mutant, reappearing at smaller scale inside the criterion written to close it. Narrowing REQ-GTA-004 to "the resolved step form" would have made the SPEC self-consistent by accepting exactly the thing it exists to remove.

**Consistency sweep (requested).** Checked every other place the summary could report a command. REQ-GTA-004 is the only requirement that binds one; no other AC asserts a reported command. The `"go test"` strings in AC-GTA-001 and AC-GTA-003 are `disabled_steps` **keys** — the step's label used as a config key — and are correct as they stand. To stop the two from being conflated in run-phase, REQ-GTA-004 now defines "the command" as binary + arguments and carries a note separating the two fields: the label is the step's identity (what REQ-GTA-001's listing is stated in terms of, and what `disabled_steps` is keyed on), the command is what ran. `plan.md` §A.1's record shape and M1 step 3 were updated to match, both naming `step.binary` + `step.args` and rejecting `gateStep.name`.

Budget after D7-D9: **16 requirements, 16 acceptance criteria** — unchanged. All three fixes are wording within existing numbered items; no REQ or AC was added, removed, or renumbered.

Post-fix verification: `moai spec lint .moai/specs/SPEC-GATE-THREE-AXES-001/spec.md` → `✓ No findings`; `moai spec audit --filter-spec SPEC-GATE-THREE-AXES-001` → modern-era clean 1, drift 0.

_<pending Implementation Kickoff Approval>_

## §E.2 Run-phase Evidence

Milestone **M1** (axis 1 — a completed run reports what it ran). Every command below was run in `.claude/worktrees/t235` on branch `WT-gate-three-axes`, against the working tree stacked on **5ee95c5e8**. M2 and M3 are untouched.

### Continuity note — this milestone was finished by a second agent

A first `manager-develop` spawn was interrupted by a session limit before it committed or wrote any evidence. It left `internal/hook/quality/gate.go` modified and two new files uncommitted, and filed no report. Nothing it produced is recorded here on its own authority: the residue was re-read against the SPEC, and every one of its tests was held against the mutant its criterion names (table below) before being accepted. AC-GTA-005 had no test at all — that gap is closed by `internal/cli/gate_summary_cli_test.go`, added in this milestone.

### Deliverables

| File | State | Lines |
|------|-------|-------|
| `internal/hook/quality/gate.go` | modified | +65 / −7 |
| `internal/hook/quality/gate_summary.go` | new | 259 |
| `internal/hook/quality/gate_summary_test.go` | new | 537 |
| `internal/cli/gate_summary_cli_test.go` | new | 142 |

`git status --short` at close lists exactly these four paths and nothing else.

### AC matrix

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-GTA-001 | PASS | `go test ./internal/hook/quality/ -run '^TestSummaryIsCompleteAndVariesWithWhatExecuted$' -count=1` | `--- PASS: TestSummaryIsCompleteAndVariesWithWhatExecuted` |
| AC-GTA-002 | PASS | `go test ./internal/hook/quality/ -run '^TestSummaryDurationIsMeasuredNotConstant$' -count=1` | `--- PASS: TestSummaryDurationIsMeasuredNotConstant (1.21s)` |
| AC-GTA-003 | PASS | `go test ./internal/hook/quality/ -run '^TestSummaryDistinguishesAllFiveSkipPaths$' -v -count=1` | all five sub-tests PASS: `a-config-off`, `b-optional-binary-absent`, `c-config-files-absent`, `d-no-staged-match`, `e-no-project-source` |
| AC-GTA-004 | PASS | `go test ./internal/hook/quality/ -run '^TestSummaryReportsExecutedCommandLineNotLabel$' -v -count=1` | all three tiers PASS: `i-test-run-script`, `ii-watch-prone`, `iii-unchanged` |
| AC-GTA-005 | PASS | `go test ./internal/cli/ -run '^TestGateCmd_PassPathEmitsExecutionSummary$' -v -count=1` | `--- PASS: TestGateCmd_PassPathEmitsExecutionSummary (0.13s)` + sub-test `a_bare_banner_does_not_satisfy_the_criterion` PASS |
| AC-GTA-006 | **PASS-WITH-DEBT** | `go test ./internal/hook/quality/ -run '^TestSummaryDoesNotDropAnExistingNotice$' -v -count=1` | `rules-dir-unconfigured` PASS; `scanner-absent` **SKIP** — `sg is installed; the scanner-absent notice is unreachable here` |
| AC-GTA-007 | PASS | `go test ./internal/hook/quality/ -run '^TestSummaryReportsUnreachedStepsAsNotReached$' -count=1` | `--- PASS: TestSummaryReportsUnreachedStepsAsNotReached` |

AC-GTA-006 is PASS-WITH-DEBT rather than PASS because the criterion names the *scanner-absent* fixture specifically, and that fixture is unreachable on a machine with `sg` installed — this one. The conjunction it asserts (an existing notice AND the summary both present in one output) is verified on the sibling `rules-dir-unconfigured` notice, which travels the same `passReason` path. The named fixture is reachable in CI where `sg` is absent; until that run is observed the named-fixture half is a gap, not a pass.

### RED evidence

**AC-GTA-005 — RED observed directly.** `internal/hook/quality/gate.go` was reverted to `5ee95c5e8` (`git show HEAD:…`) and the new CLI test run against it:

```
gate_summary_cli_test.go:69: stderr carries no per-step summary rows; got:
    typecheck: skipped (no default for this language; set gate.typecheck.command to enable one)
    ast-grep scan skipped: ast_grep_gate.rules_dir is empty (not configured); …
--- FAIL: TestGateCmd_PassPathEmitsExecutionSummary (0.17s)
```

That output is also the criterion's rejected formulation, measured: the pre-implementation tree emits a **non-empty** string on the pass path, so an assertion of the form "stderr is non-empty" would have passed against the broken tree. Binding to the outcome lines of a named toolchain is what makes the criterion falsifiable, and the sub-test `a_bare_banner_does_not_satisfy_the_criterion` holds the predicate to it mechanically.

**AC-GTA-001 … 007 (the residue's six tests) — RED established by mutation.** These tests were written by the interrupted agent, so no pre-GREEN failing output exists for them and a passing run alone proves nothing: a test written after its implementation can pass while verifying nothing. Each was instead held against the mutant its own criterion names — the mutant applied to the source, the test run, the source restored and its SHA-1 re-checked against the pre-mutation copy. All six were killed:

| # | Mutant applied | Criterion's name for it | Target test | rc | Observed |
|---|----------------|-------------------------|-------------|----|----------|
| 1 | `runSummary.mark` made a no-op — summary renders from configuration, identical every run | AC-GTA-001 Mutant A | `TestSummaryIsCompleteAndVariesWithWhatExecuted` | 1 | `step "go test" outcome field "not reached" carries 0 outcome tokens, want exactly 1` |
| 2 | reported duration replaced by the constant `time.Second` | AC-GTA-002 Mutant | `TestSummaryDurationIsMeasuredNotConstant` | 1 | `step that slept 1.2s reported 1s; the reported duration is not measured` |
| 3 | all five skip reasons collapsed to the token `skipped` | AC-GTA-003 Mutant A | `TestSummaryDistinguishesAllFiveSkipPaths` | 1 | `reason "skipped" does not name its own observation (missing "disabled_steps")` — and the same on paths (b) and (c) |
| 4 | `markExecuted(stepName, stepName, …)` — report the label in place of argv | AC-GTA-004 Mutant D | `TestSummaryReportsExecutedCommandLineNotLabel` | 1 | tier (i) survives (label and argv coincide there); tiers (ii) and (iii) fail on the dropped `-- --passWithNoTests` — exactly the signature the criterion predicts |
| 5 | `withSummary` returns the summary alone, dropping the accumulated notices | AC-GTA-006 Mutant | `TestSummaryDoesNotDropAnExistingNotice` | 1 | `the ast-grep notice was dropped when the summary was emitted` |
| 6 | steps seeded as `outcomeExecuted` instead of `outcomeNotReached` | AC-GTA-007 Mutant | `TestSummaryReportsUnreachedStepsAsNotReached` | 1 | `step "golangci-lint" was never reached but is reported as "executed"` |

Mutant 4's asymmetry is the load-bearing detail: it is the one mutant a weaker fixture set would have missed, and the criterion's three-tier table is what catches it.

### End-to-end observation

`CLAUDE_PROJECT_DIR=/tmp/t235-fixture go run ./cmd/moai gate` against a passing two-file Go module, exit **0**:

```
quality gate steps (4 configured):
  - go vet: executed in 58ms — go vet ./...
  - typecheck: skipped — no default for this language; set gate.typecheck.command to enable one
  - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, …)
  - go test: executed in 46ms — go test ./...
```

Before this milestone the same invocation printed the two notices and nothing else. `internal/cli/gate.go` needed no change, as `plan.md` M1 step 6 predicted — confirmed by running the path, not by reading it.

### Suite, lint, coverage, cross-platform

| Check | Command | Result |
|-------|---------|--------|
| quality suite | `go test ./internal/hook/quality/ -count=1 -timeout 300s` | rc=0 — `ok … 5.979s` |
| cli suite | `go test ./internal/cli/... -count=1 -timeout 1200s` | rc=0 — 17 packages `ok`; `internal/cli` alone 286.962s |
| vet | `go vet ./internal/hook/quality/... ./internal/cli/...` | rc=0, no output |
| lint | `golangci-lint run --timeout=5m ./internal/hook/quality/... ./internal/cli/...` | rc=0 — `0 issues.` |
| coverage | `go test ./internal/hook/quality/ -count=1 -cover` | `coverage: 90.3% of statements` (threshold 85%) |
| windows compile | `GOOS=windows go vet ./...` | rc=0 — compilation evidence only, never behavioural |
| linux compile | `GOOS=linux go vet ./...` | rc=0 — same standing |
| subagent boundary | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/quality/ internal/cli/gate.go internal/cli/gate_summary_cli_test.go` | rc=1, no matches |

The 1200s floor on `internal/cli` is not optional: the package alone measures 287s here, and a 300s bound fails a tree that is fine.

### What was NOT observed

- **AC-GTA-006's named fixture.** `sg` is installed on this machine, so the scanner-absent notice is unreachable and its sub-test skipped. The AC's conjunction is verified on the sibling notice only.
- **The full suite.** `go test ./...` was not run and is prohibited here. Affected packages only; the full-suite verdict is CI's, against the pull-request head.
- **Any platform but darwin/arm64.** The two `GOOS=` runs are compilation evidence. No behaviour was observed on Windows or Linux.
- **`golangci-lint` executing rather than skipping.** Every fixture built here lacks a `.golangci.*` config, so that step took its config-files-absent skip path in all runs. The AC-GTA-005 assertion deliberately excludes it for that reason.
- **M2 and M3 territory.** No `WaitDelay`, no `Setpgid`, no lock. `grep -rn 'Setpgid\|Killpg\|SysProcAttr\|WaitDelay' internal/hook/quality/` still returns nothing — the §E.1 baselines are unchanged.
- **A transient hang the lead reported.** Immediately after the first agent died, one package run hit a 303s timeout with a stack showing `exec.Cmd.Start` blocked on a pipe read. It did not reproduce in any run of this milestone: the package is green in 5.98s and no orphan was found. Treated as contention with the dying process, not as a defect — and explicitly NOT as evidence for the axis-2 defect M2 exists to fix.

### Where the SPEC and the source disagreed

Nothing needing a SPEC change. Two observations worth carrying forward:

1. **The ast-grep axis is not a summary row.** REQ-GTA-001 binds "each configured step", and ast-grep is not a `gateStep` in any toolchain table — it is a separate gate whose notice travels the `passReason` path. It is therefore reported as a notice above the summary rather than as a fifth row. Consistent with the requirement as written; recorded because a reader counting steps may expect otherwise.
2. **"Executed" includes a step that failed to launch.** `markExecuted` fires from a `defer` in `runStep`, so a step whose binary is missing is still reported as executed with its command line. This matches REQ-GTA-004's own definition — the command "as handed to the process launcher" — and is what lets AC-GTA-004's Node fixtures assert the argv on a machine without `npm`. It is a deliberate reading, not an oversight.

---

## §E.2 Run-phase Evidence — Milestone M2

Milestone **M2** (axis 2 — a step's timeout terminates the step). Every command below was run in `.claude/worktrees/t235` on branch `WT-gate-three-axes`, against the working tree stacked on **ba22f41cf**. M3 is untouched.

### Deliverables

| File | State | Lines |
|------|-------|-------|
| `internal/hook/quality/gate.go` | modified | +40 / −3 |
| `internal/hook/quality/step_process_group.go` | new | 27 |
| `internal/hook/quality/step_process_group_unix.go` | new (`//go:build !windows`) | 34 |
| `internal/hook/quality/step_process_group_windows.go` | new (`//go:build windows`) | 20 |
| `internal/hook/quality/gate_step_termination_test.go` | new | 174 |
| `internal/hook/quality/gate_step_termination_unix_test.go` | new (`//go:build !windows`) | 47 |

`git status --short` at close lists exactly these six paths and nothing else. `internal/cli` is unmodified, as M2 expected.

### RED evidence — observed before any implementation existed

Both criteria were written and observed failing against the pre-implementation tree, with only the two constants (`stepWaitGrace`, `descendantTerminationNote`) added so the tests would compile. `runStep` was untouched at that point: no `WaitDelay`, no `Setpgid`, no group signal.

AC-GTA-008, observed in isolation per the criterion's mandatory RED-isolation clause — a narrowing `-test.run` plus an explicit `-test.timeout`, never inside a package run:

```
$ go test -count=1 -run '^TestRunStep_ReturnsWithinGraceWhenADescendantHoldsTheStream$' -timeout 20s ./internal/hook/quality/
panic: test timed out after 20s
	running tests:
		TestRunStep_ReturnsWithinGraceWhenADescendantHoldsTheStream (20s)
...
goroutine 7 [chan receive]:
os/exec.(*Cmd).awaitGoroutines(0x7400a433e680, 0x0)
	.../src/os/exec/exec.go:981 +0x14c
os/exec.(*Cmd).Wait(0x7400a433e680)
	.../src/os/exec/exec.go:948 +0x140
rc=1
```

That stack is the defect `spec.md` §A.2 describes, reproduced rather than assumed: the step's own 2s deadline came and went, the direct child had already exited, and `Wait` stayed blocked in `awaitGoroutines` — the pipe-copying goroutines — because the grandchild still held the write end. It blocked for the full 20s until the test binary's own alarm fired.

AC-GTA-009, same isolation:

```
$ go test -count=1 -run '^TestRunStep_TerminatesTheStepsDescendant$' -timeout 20s ./internal/hook/quality/
panic: test timed out after 20s
	running tests:
		TestRunStep_TerminatesTheStepsDescendant (20s)
rc=1
```

AC-GTA-010 has no RED — it is a preservation criterion. Its expected values were **measured on the pre-change tree** rather than asserted from the new code: `TestRunStep_WithinDeadlineOutputAndCwdAreUnchanged` was run and passed against the untouched `runStep` (`rc=0`, `--- PASS … (0.01s)`) before any of the M2 edits landed, which is what makes "byte-identical to the pre-change tree's" an observation instead of a claim.

### AC matrix

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-GTA-008 (Unix half) | PASS | `go test -count=1 -run '^TestRunStep_ReturnsWithinGraceWhenADescendantHoldsTheStream$' -timeout 60s -v ./internal/hook/quality/` | `--- PASS: TestRunStep_ReturnsWithinGraceWhenADescendantHoldsTheStream (2.45s)` — returned in 2.45s against the T+G = 4s bound |
| AC-GTA-008 (Windows half) | **GAP** | not runnable here | see § What was NOT observed |
| AC-GTA-009 | PASS | `go test -count=1 -run '^TestRunStep_TerminatesTheStepsDescendant$' -timeout 60s -v ./internal/hook/quality/` | `--- PASS: TestRunStep_TerminatesTheStepsDescendant (2.57s)` |
| AC-GTA-010 (first half — attribution) | PASS | `go test -count=1 -run '^TestRunStep_(ParentDeadlineIsNotBlamedOnTheStep\|StepDeadlineStillBlamesTheStep)$' -timeout 120s -v ./internal/hook/quality/` | `--- PASS: TestRunStep_ParentDeadlineIsNotBlamedOnTheStep (0.10s)` / `--- PASS: TestRunStep_StepDeadlineStillBlamesTheStep (0.10s)` |
| AC-GTA-010 (diff constraint) | PASS | `git diff 294b4b6ab -- internal/hook/quality/gate_timeout_attribution_test.go \| wc -c` | `0` — the protected file is byte-identical to the tree the SPEC cites; not one line was added, modified, or deleted |
| AC-GTA-010 (second half — happy path) | PASS | `go test -count=1 -run '^TestRunStep_WithinDeadlineOutputAndCwdAreUnchanged$' -timeout 60s -v ./internal/hook/quality/` | `--- PASS (0.01s)` — same message and same `cwd=` as the pre-change measurement |

The two t218 regression tests stayed green without being touched because both assert with `strings.Contains`, not equality: the descendant note is **appended** to the existing reason rather than replacing it. That property was checked in the source before the edit, not assumed from the instruction.

### Mutation evidence

Each criterion was held against the mutant its own text names, and the mutant was actually built and run.

| Mutant | Applied as | Result |
|--------|-----------|--------|
| AC-GTA-009's named mutant — *bounding `Wait` alone* | `Setpgid = false` + the group signal removed; `WaitDelay` left in place | **AC-GTA-008 PASSED (2.33s), AC-GTA-009 FAILED**: `descendant 18108 is still alive 2s after the step returned`. The separation the SPEC insists on is empirically load-bearing: a `WaitDelay`-only implementation satisfies AC-GTA-008 in full while leaving the orphan running. |
| The converse — *group signal alone, no `WaitDelay`* | `cmd.WaitDelay = 0`; process group left in place | **AC-GTA-008 FAILED** — hung to the 20s alarm, the RED shape again. Neither mechanism substitutes for the other, in either direction. |
| AC-GTA-010 Mutant B — *reordered stream concatenation* | `stdout.String() + stderr.String()` | **FAILED**: `got "…\n\nstdout-line cwd=…\nstderr-line"` vs `want "…\n\nstderr-line\nstdout-line cwd=…"` |
| AC-GTA-010 Mutant B — *working directory detached* | `cmd.Dir = os.TempDir()` | **FAILED**: `got "…cwd=/var/folders/kt/…/T"` vs `want "…cwd=…/internal/hook/quality"` |
| AC-GTA-010 Mutant A — *editing the two test bodies to agree* | not applicable | Structurally excluded by the 0-byte diff above rather than by a test run. |
| AC-GTA-008 Mutant A — *raising T so the descendant finishes first* | not applicable | Structurally excluded: the assertion bound is `T + stepWaitGrace` and the descendant sleeps 30s, so no value of T rescues it. |

Every mutant was reverted immediately after its run; the tree at commit time carries none of them.

### Suite, vet, lint

| Check | Command | Result |
|-------|---------|--------|
| affected package | `go test -count=1 -timeout 900s ./internal/hook/quality/...` | `ok  github.com/modu-ai/moai-adk/internal/hook/quality  49.976s`, rc=0 — see the contention note below for a later re-run |
| vet | `go vet ./internal/hook/quality/...` | rc=0, no output |
| lint | `golangci-lint run --timeout=5m ./internal/hook/quality/...` | `0 issues.`, rc=0 (one `errcheck` finding on an unchecked `fmt.Fprintf` in the fixture was fixed first) |
| cross-compile | `GOOS=windows go vet ./...` / `GOOS=linux go vet ./...` | both rc=0 — **compilation evidence only**, no behavioural claim rests on it |
| fixture hygiene | `pgrep -fl 'TestHelperOrphanChild'` | rc=1, no match — no fixture process leaked, including after the two RED runs that died to the test alarm and therefore skipped `t.Cleanup` |
| `internal/cli` gate subset | `go test -count=1 -run 'Gate' -timeout 600s ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  156.895s`, rc=0 |
| `internal/cli` whole package | `go test -count=1 -timeout 1200s ./internal/cli/...` | **inconclusive** — see the contention note below |

### Contention note — two measurements this machine could not make cleanly

Four other worktree lanes were running `go test ./internal/cli/...` against this same machine while these checks ran; `uptime` read a load average of **46.07**. Two results are attributable to that load rather than to the tree, and both are reported as gaps, not as passes:

1. **`internal/cli` whole-package: `FAIL … 1201.914s`.** No individual `--- FAIL` appears anywhere in the log — the package exhausted its 20-minute wall clock, and the test running at the alarm had been going 4 seconds. Every sub-package (`agentlint`, `harness`, `pr`, `worktree`, …) reported `ok`. The single known-good range for this package is 287-866s; four concurrent runs of it is the difference. M2 modifies no file under `internal/cli`, and the gate-related subset of that package passes in 156.895s (row above). Re-running it now would have added a fifth concurrent run of the same package, which the milestone's own load discipline forbids, so it was not re-run. **The verdict for this package is CI's.**
   Worth recording separately: the backgrounded wrapper around that run reported `exit code 0` while its body read `FAIL`. The verdict here is read from the `FAIL` line, never from the wrapper's status.
2. **A later `./internal/hook/quality/...` re-run failed on an M1 test, not an M2 one.** `--- FAIL: TestSummaryDurationIsMeasuredNotConstant (3.09s)` — `step that returned immediately reported 1.820353667s, at or above the 1.2s the other step slept`. That criterion (AC-GTA-002) asserts a two-sided duration bound, so under enough load a process spawn alone crosses the lower fixture's ceiling: the test is measuring the machine. On the same tree, isolated, it passes 3/3 (`go test -count=3 -run '^TestSummaryDurationIsMeasuredNotConstant$'` → three `--- PASS`, 2.25s / 1.64s / 1.70s), and the same package passed whole at `49.976s` earlier in this milestone before the other lanes started. Nothing in M2 touches the duration measurement. Flagged for the lead as an M1 test that is contention-sensitive by construction.

### Fixture hygiene

The orphan fixture starts two processes per test and every one is bounded twice: the grandchild carries its own `-test.timeout=60s` (an external bound that holds even when the test binary panics and `t.Cleanup` never runs), and the test registers a `t.Cleanup` kill. No trailing `kill` line exists anywhere in the fixture — every early return would skip it. No background load is spawned by any recipe here. Both re-executions of the test binary narrow with `-test.run`, so neither can re-enter the suite.

### What was NOT observed

- **Windows behaviour, in full.** `GOOS=windows go vet ./...` says the platform file compiles and nothing more. The Windows half of AC-GTA-008 — that `WaitDelay` bounds the return there, and that the reason states descendants may have survived — is **a gap, not a pass**, until the CI Windows matrix run is read. AC-GTA-008's own Mutant C ("a Windows build that compiles but never applies the bound") is by construction not killable from this machine.
- **Linux behaviour.** Same standing as Windows: compilation only. The Unix implementation is exercised on darwin/arm64 alone.
- **The full suite.** `go test ./...` was not run and is prohibited here. The full-suite verdict is CI's, against the pull-request head.
- **A Windows job object.** Out of scope per `plan.md` §A.2; the Windows reason therefore reports that descendants may have survived rather than implying they were killed.
- **A clean `internal/cli` whole-package run.** Not obtained on this machine — see the contention note above. Only the gate-related subset was measured cleanly.
- **M3 territory.** No lock was added. `internal/cli/gate.go` is unmodified.
- **The lint axis (card t233 / issue #1631).** Untouched, per `spec.md` §D.

### Where the SPEC and the source disagreed

One place, and it changes an implementation detail rather than the SPEC.

**The deadline is not the only exit that leaves a descendant behind.** REQ-GTA-009 and AC-GTA-009 are both phrased around "when the step's deadline expires", but the fixture the SPEC itself specifies in §D.2 — a child that spawns a grandchild and *exits* — never reaches the deadline at all: the direct child is gone in milliseconds, so `exec`'s cancellation path, and with it `cmd.Cancel`, never runs. Implementing the group signal only in `Cancel` would have passed AC-GTA-008 and left AC-GTA-009's orphan alive. The group is therefore swept on **every** exit from `runStep` (a `defer`), and additionally in `Cancel` for the case where the child really does outlive its deadline. This satisfies REQ-GTA-009's stated case and the case its own fixture actually produces; it does not widen the blast radius, because the group is created per step at spawn and so contains only that step's descendants.

Two smaller notes:

1. **`WaitDelay` fires on a path the timeout branch does not cover.** When the child exits cleanly but a descendant holds the pipes, `cmd.Run` returns `exec.ErrWaitDelay` and `stepCtx.Err()` is nil — so the existing `DeadlineExceeded` branch never sees it. A dedicated branch reports it as a failure: the captured output is provably incomplete, and passing on truncated output would restore exactly the silence axis 1 was written to remove.
2. **The grace budget is a constant, not a config key.** `stepWaitGrace = 2s` in `step_process_group.go`, per `plan.md` §A.4's proposal — a safety bound rather than a policy. No `gate.yaml` key was added, so no template mirror was needed in M2.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_milestone: M1
run_status: complete-with-debt      # AC-GTA-006's named fixture is unreachable locally
run_complete_at: 2026-08-25
run_commit_sha: 3c441782b           # M1; backfilled in the follow-up commit
run_base_sha: 5ee95c5e8
run_branch: WT-gate-three-axes
run_worktree: .claude/worktrees/t235

ac_pass_count: 6                    # AC-GTA-001..005, 007
ac_pass_with_debt_count: 1          # AC-GTA-006 (scanner-absent fixture skipped: sg installed)
ac_fail_count: 0
ac_out_of_scope: [AC-GTA-008, AC-GTA-009, AC-GTA-010, AC-GTA-011, AC-GTA-012, AC-GTA-013, AC-GTA-014, AC-GTA-015, AC-GTA-016]

red_evidence:
  ac_gta_005: observed              # gate.go reverted to 5ee95c5e8; test failed
  ac_gta_001_004_006_007: mutation  # residue tests; 6/6 named mutants killed, no pre-GREEN output exists
  mutants_applied: 6
  mutants_killed: 6
  mutants_survived: 0

total_run_phase_files: 4
files_modified: 1
files_added: 3
preserve_list_post_run_count: 0     # git status --short lists only the 4 SPEC-scope paths

new_warnings_or_lints_introduced: 0
lint_command: "golangci-lint run --timeout=5m ./internal/hook/quality/... ./internal/cli/..."
lint_result: "0 issues."
coverage_internal_hook_quality: 90.3
coverage_threshold: 85

cross_platform_build:
  darwin_arm64_test: pass           # behavioural
  goos_windows_vet: pass            # compilation only
  goos_linux_vet: pass              # compilation only

pre_commit_fetch: "git fetch origin main -> rc=0; git rev-list --count --left-right origin/main...HEAD -> 9 5"
post_push_fetch: not-applicable     # WT-gate-three-axes is unpushed; the worktree holds the only copy
full_suite_verdict: deferred-to-ci  # go test ./... prohibited locally

m1_to_mN_commit_strategy: one-commit-per-milestone
next_milestone: M2                  # axis 2 — not started, deliberately untouched
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
