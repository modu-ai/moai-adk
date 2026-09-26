# plan.md — SPEC-AUTONOMY-ESCALATION-001

Milestones are ordered by decision reversibility: the decisions most likely to change come
first, mechanical wiring last. No time estimates; priority labels only.

## §A — Context

- Base: `develop` at `ca1d5dc43`; worktree `.claude/worktrees/t1235`, branch `WT-escalation-detector`.
- Tier L (9 detection classes across PreToolUse, PostToolUse and checkpoint surfaces, a new
  record layer, config, template default; expected > 15 files).
- Run-phase entry gate: card t1234 (A1, contract schema) merged into `develop`. Plan may proceed
  in parallel; run may not (card text).
- Measured premises: `research.md` §B. Proposed mechanics: `design.md`.

## §B — Known issues carried into run

- **B-1** Queue store forbids a fourth state (`backlog_schema_freeze_test.go:3-5,70`); the
  needs-decision marking must not touch it.
- **B-2** `graph.FileAPI` is working-tree only and CGO/grammar-dependent (`codequery.go:65-95`).
- **B-3** `harness.escalation` is a homonym (`internal/config/types.go:1199-1243`); do not reuse.
- **B-4** `MOAI_AUTONOMY_TIER` stays untouched (`internal/config/envkeys.go:139-148`).
- **B-5** Hook bind budget: no subprocess on the per-tool-call path; git reads only at checkpoints.
- **B-6** Changing any `acceptance.md` requires `./internal/spec` in the re-measure scope (AC
  snapshot guard).

## §C — Pre-flight (run phase)

1. `git merge-base --is-ancestor <A1 merge commit> HEAD` → exit 0 (A1 present).
2. Re-read every 「A1 스키마 확정 후 재조정」 requirement against A1's landed schema; if a field in
   spec.md §F is renamed or missing, stop and return a blocker to the orchestrator for a
   mid-run spec amendment.
3. Capture the AC-AE-001 golden baseline in its own commit before any detector code.

## §D — Constraints

- Detection and reporting only; no gate rewiring (A3), no closure report (A4), no schema design (A1).
- Default config `guided`; template carries no card id, SPEC id, date, or SHA.
- One writer per tree; change-scoped tests only; no background load.

## §E — Self-verification (per milestone)

E1 AC matrix with command + verbatim output + HEAD; E2 cross-platform build; E3 coverage;
E4 subagent-boundary grep; E5 lint delta; E8 RED output before GREEN.

## §F — Milestones

### M1 — Record layer and activation gate (Priority High, most change-prone)

- Decide the escalation record location and the needs-decision derivation (design.md §A; open
  question Q1). Add `workflow.autonomy` config parsing with the inert default.
- ACs: AC-AE-001, AC-AE-002.

### M2 — Report writer, dedup, fault handling, acceptance-change

- Report shape (design.md §D), occurrence counting, `not-armed` / `not-checked` audit lines,
  queue-untouched guarantee, class 1.
- ACs: AC-AE-003, AC-AE-005, AC-AE-006, AC-AE-007, AC-AE-019, AC-AE-020, AC-AE-021.

### M3 — Path and command classes (PreToolUse / PostToolUse)

- Classes 2a, 2b, 3 wired on the hook paths without altering existing decisions.
- ACs: AC-AE-008, AC-AE-009, AC-AE-010.

### M4 — New-architecture/API detector (Priority Medium, heuristic)

- Base-blob extraction route, package / exported-decl / CLI-verb / MCP-tool / config-key diff,
  `off` switch, not-observed labelling for unsupported extraction.
- ACs: AC-AE-011, AC-AE-012, AC-AE-013.

### M5 — Evidence, irreversible action, operational trips, escalate_on gating

- Classes 5, 6, 7, 8, 9 and per-class disabling from `escalate_on`.
- ACs: AC-AE-004, AC-AE-014, AC-AE-015, AC-AE-016, AC-AE-017, AC-AE-018, AC-AE-022.

### M6 — Template default and documentation (mechanical)

- Template `workflow.yaml` gains the `autonomy` block with `mode: guided`; `make build`.

## §G — Risks

| Risk | Mitigation |
|---|---|
| A1 schema diverges from the assumed fields | Tagged requirements + pre-flight step 2 |
| PreToolUse latency under contract mode | Pure in-memory checks on the hot path; checkpoints for heavy work |
| Class 4 false positives on refactors that move declarations | Report lists additions only by kind; sync-audit remains the backstop (C6) |
| Class 6 regex misses an obfuscated push | Reported as residual risk; blocking belongs to A3 |
| Invariant commands never run → never checked | Recorded known limit (design.md §C.2) |

## §H — Open questions (for the lead)

- **Q1** Where does the needs-decision record live — card evidence area, or the A1/F1 shared
  record layer? Blocks M1.
- **Q2** Is a new CLI verb for on-demand checkpoints acceptable, given it is itself a class-4 event?
- **Q3** Should the AC counting rule be shared with the local-only ac-baseline guard, or
  re-implemented in the product path?
- **Q4** Sync-audit retry ceiling source — no config key was located; is it the same tier map?
- **Q5** Does "recorded CI failure" have an existing on-disk producer, or is class 5's CI limb
  expected to stay not-observed until one exists?
