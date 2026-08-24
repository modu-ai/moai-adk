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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
