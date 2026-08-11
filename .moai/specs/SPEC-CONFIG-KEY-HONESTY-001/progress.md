# SPEC-CONFIG-KEY-HONESTY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-12
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: [SPEC-CONFIG-TIER-PERSIST-001]
code_baseline: ed70e4354
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.72
    threshold: 0.80
    dimensions: {clarity: 0.78, completeness: 0.80, testability: 0.62, traceability: 0.72}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7, D8]
    deferred: [D9, D10, D11, D12, D13, D14, D15]
    report: .moai/reports/plan-audit/SPEC-CONFIG-KEY-HONESTY-001.md
  iteration_2:
    verdict: FAIL
    score: 0.78
    threshold: 0.80
    dimensions: {clarity: 0.75, completeness: 0.75, testability: 0.78, traceability: 0.85}
    must_pass: 7/7
    root_cause: baseline-drift (plan authored against d5336214e, audited against ed70e4354)
    report: .moai/reports/plan-audit/SPEC-CONFIG-KEY-HONESTY-001-review-2.md
  iteration_3:
    verdict: PASS
    score: 0.87
    threshold: 0.80
    dimensions: {clarity: 0.85, completeness: 0.90, testability: 0.83, traceability: 0.90}
    must_pass: 7/7
    refresh_type: baseline-reverification
    code_baseline: ed70e4354
    resolved: [D1..D8 stale citations, D3 adhoc-live deleted-file foundation, D4 AC-CKH-012 false baseline, D9 AC ceiling 23→15, D10 NFR floor 200→900, D11 M3-hold in spec/plan, D12 main-fork premise softened, D14 folded into D12]
    deferred: [D13 handoff-note-path placeholder (forward-ref by design), D14 AC-CKH-013 token-counting (minor)]
```

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- Code baseline `ed70e4354` (iteration-3 refresh). The prior `d5336214e` baseline was 12 days stale
  at iteration-2 audit; all file:line citations were re-verified and updated against `ed70e4354`.
- **Iteration-3 refresh (2026-08-12).** Plan-audit iteration 2 returned FAIL 0.78 (threshold 0.80)
  with the root cause identified as baseline drift — the plan was authored against `d5336214e` and
  audited against HEAD `ed70e4354`. The iteration-3 refresh resolved all 15 named defects (D1-D15
  from the iteration-2 report): D1/D2/D5/D6/D7/D8 stale file:line citations re-verified and updated;
  D3 `adhoc-live` class foundation re-derived (sole confirmed instance file deleted at `5792fc755`,
  class retained as forward-looking with zero current instances, `tmux_preferred` reclassified
  dead); D4 AC-CKH-012 baseline re-derived against HEAD (`isHookOptInEnabled` refactored to
  delegator at `e3f8dd463`, inline-struct readers now at `routing_ledger.go:104` +
  `update.go:1140`); D9 Tier M AC ceiling (23→15 via consolidation, documented in acceptance.md §A
  clause 8); D10 NFR-CKH-002 floor raised (200→900 keys / 200→250 fields); D11 M3-hold reflected in
  spec.md §B.1 + plan.md §F M3 (not only progress.md); D12 main-fork premise softened to
  conditional; D14 AC-CKH-016 token-counting folded into the consolidated AC-CKH-010 (minor,
  carried forward). The §A discipline framework, §C falsification design, and class taxonomy were
  validated as strong by the auditor and preserved unchanged.
- Findings F1-F7 each re-verified against `ed70e4354`; one drift recorded
  (spec.md §A.8 — shipped `workflow.yaml` worktree toggles contradict `internal/config/defaults.go`).
- **F3 re-derived path-resolved at the plan-audit revision** (D3): 287 distinct `yaml:`-tagged field
  names, **174** with zero production reads and 4 accessor-only; 161 map to a shipped key across 188
  (file, key) occurrences. This supersedes the earlier bare-field-name figures (122 / 5 / 121),
  which were produced by the very method this SPEC forbids as AP-3. The re-derivation used a
  throwaway `go/packages` selector-resolution probe over `./internal/... ./pkg/... ./cmd/...`
  (106 packages, 0 type errors); M2's guard is its durable implementation.
- The recomputation resolved the §A.3 ↔ §A.6 contradiction rather than papering over it:
  `auto_merge` is dead under path resolution and now appears in the family table, matching §A.6's
  "AutoMerge has zero reads". 43 names flip live→dead in total.
- Prose-consumer discriminator measured: dotted-path fixed-string probe yields 0-1 hits per key
  versus up to 46 for the bare leaf key.
- SPEC ID regex self-check executed: `PASS`.
- Status: `draft`. Awaiting plan-audit re-run and Implementation Kickoff Approval.

### Epic run order (dependency sequencing)

`depends_on: [SPEC-CONFIG-TIER-PERSIST-001]` records that this SPEC reads the tier-resolution
contract E3 owns — a shipped key's liveness is judged against the loader that actually resolves it,
and E3 fixes which tier that is. The edge is a **read** edge: nothing in M1-M6 writes a surface E3
also writes.

The run-phase `Depends_on Pre-flight Check` treats a dependency as fulfilled only at
`status: completed`. Every SPEC in this Epic is currently `draft`, so entering `/moai run` on this
SPEC before E3 closes raises the 3-option wait / override / abort blocker. **The dependency is
satisfied by sequencing, not by an `--ignore-deps` bypass** — the run order below is the mechanism,
and it is consistent with the orders recorded in `SPEC-UPDATE-DATA-SURVIVAL-001` §E.1 and
`SPEC-CONFIG-TIER-PERSIST-001` §E.1:

| Order | SPEC | Gate to clear before the next entry |
|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) | reaches `status: completed` — REQ-RIL2-015/016 landed |
| 2 | `SPEC-UPDATE-DATA-SURVIVAL-001` (E2) | reaches `status: completed` — backup coverage + failure contract landed |
| 3 | `SPEC-CONFIG-TIER-PERSIST-001` (E3) | reaches `status: completed` — tier precedence + atomic write landed |
| 4 | **`SPEC-CONFIG-KEY-HONESTY-001`** (this SPEC) | — |
| 5+ | remaining Epic SPECs (`SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-UPDATE-CI-GUARD-001`, `SPEC-UPDATE-DOC-DRIFT-001`) | no `depends_on` edge to this SPEC |

Do **not** invoke `/moai run` on this SPEC with `--ignore-deps`. If E3 slips and starting this SPEC
early becomes necessary, the correct move is to run **M1, M2, M4, M5, and M6** — none reads E3's
tier contract — and hold **M3** open, since `quality.yaml`'s parse path is the one place where which
tier resolved the block changes the answer. That is a scope decision for the orchestrator to surface
via `AskUserQuestion`, not a flag the run-phase agent may set on its own.

One ordering constraint is internal to this SPEC and independent of the Epic order: M1 must land
before M2, because M2's guard fails on any `dead` / `unresolved` / `unbound` key absent from M1's
**P** / **R** allowlists — with no inventory, every shipped key fails at once.

### Deferred audit defects (D9-D15) — iteration-3 resolution status

Recorded so the next iteration does not re-derive them. As of iteration-3 refresh:
- D9 (`qualityFileWrapper` citation at `types.go:1174`): **RESOLVED** → updated to `types.go:1312`.
- D10 (§A.6 grep transcript pre-filtered): **RESOLVED** → §A.6 prose documents the filter; NFR floor
  tightened (see D11 below).
- D11 (NFR-CKH-002's 200-key floor against ~1020 shipped keys): **RESOLVED** → floor raised to
  900 keys / 250 fields.
- D12 (AC counts tokens rather than asserting meaning): **ACKNOWLEDGED, carried forward** — the
  consolidated AC-CKH-010 Part B retains the token-count form as a necessary-but-not-sufficient
  proxy; a semantic grep was considered but rejected as brittle against prose rewording.
- D13 (`<handoff-note-path>` placeholder): **ACKNOWLEDGED, forward-ref by design** — the path is
  concrete by run-phase; no plan-time action possible.
- D14 (`main-fork/` premise false in worktree): **RESOLVED** → plan.md §B2 softened to conditional
  ("MAY exist in some checkouts"); AP-4 hazard documented as a general rule.
- D15 (`depends_on` target in `draft`): **ADDRESSED** by the run-order table above + M3-hold
  mechanism, not by frontmatter change.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Logged by the orchestrator before the first run-phase `Agent()` spawn (per
orchestration-mode-selection.md §D).

### Input parameters

- **tier**: M
- **scope (file count)**: ~10-14 files (config Go source + template YAML + test
  files + `.moai/docs` triage rule + `CLAUDE.local.md` + testdata inventory)
- **domain count**: 4 (`internal/config` Go, `internal/template` YAML, test
  guard, `.moai/docs` + docs prose)
- **file language mix**: Go + YAML + markdown (no frontend, no shell)
- **concurrency benefit**: LOW — coding-heavy with data dependencies (M1
  inventory → M2 guard; M1 W/P/R/D classification → M4/M5/M6 consumption)

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | NO | 5 milestones, multi-file, semantic classification work |
| 2 background | NO | coding-heavy write work, not read-only async |
| 3 agent-team | RETIRED | Mode 3 tombstone (Agent Teams static layer retired) |
| 4 parallel | NO | coding-heavy violates Anthropic's coding-task parallelism caveat |
| 5 sub-agent | **YES** | sequential per-milestone delegation; data deps + semantic judgment |
| 6 workflow | NO | not high-volume mechanical (W/P/R/D classification + guard logic is semantic) |

### Decision

`sub-agent`

### Justification

Tier M coding-heavy work with 5 milestones (M1/M2/M4/M5/M6) touching config Go,
template YAML, test files, and docs prose. Per Anthropic's coding-task
parallelism caveat, sequential sub-agent delegation (Mode 5) is the correct
default: the milestones have hard data dependencies (M1's inventory +
allowlists are consumed by M2's guard, and M1's W/P/R/D classification is
consumed by M4/M5/M6), and the work involves semantic judgment (classifying
each shipped key into W/P/R/D), not a uniform mechanical transform. Progression
mode: **autonomous (goal-armed ac_converge)** — selected by the user at the
Implementation Kickoff Approval gate. The `/moai goal` ac_converge condition is
armed alongside M1 delegation; the loop continues across milestones without
per-milestone checkpoints until all active ACs (M3-hold excluded) converge.
