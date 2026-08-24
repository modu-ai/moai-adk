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

_<pending Implementation Kickoff Approval>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
