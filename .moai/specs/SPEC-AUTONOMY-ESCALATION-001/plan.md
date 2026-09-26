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
2. Re-check every 「A1 plan-audit 통과본으로 재확인」 requirement against the plan-audit-passed A1
   schema (spec.md §F lists each field as drafted at `8f77d9a33`); if a field is renamed or
   missing, or an open item O1-O9 resolved differently, stop and return a blocker to the
   orchestrator for a mid-run spec amendment.
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
  question Q1). Read `workflow.autonomy.mode` through A1's config reader (inert under `guided`);
  add only the `escalation.new_api_detector` key.
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

- Classes 5, 6, 7, 8, 9 and invalid-contract handling (REQ-AE-020).
- ACs: AC-AE-004, AC-AE-014, AC-AE-015, AC-AE-016, AC-AE-017, AC-AE-018, AC-AE-022.

### M6 — Template default and documentation (mechanical)

- Template `workflow.yaml` gains `autonomy.escalation.new_api_detector: graph` under the block
  A1 introduces (A1 owns `mode: guided` and `budget_default`); `make build`.

## §G — Risks

| Risk | Mitigation |
|---|---|
| A1 plan-audit changes the `8f77d9a33` draft | Tagged requirements + spec.md §F open items + pre-flight step 2 |
| PreToolUse latency under contract mode | Pure in-memory checks on the hot path; checkpoints for heavy work |
| Class 4 false positives on refactors that move declarations | Report lists additions only by kind; sync-audit remains the backstop (C6) |
| Class 6 regex misses an obfuscated push | Reported as residual risk; blocking belongs to A3 |
| Invariant commands never run → never checked | Recorded known limit (design.md §C.2) |

## §H — Open questions (for the lead)

- **Q1** Where does the needs-decision record live — card evidence area, or the A1/F1 shared
  record layer? Blocks M1.
- **Q2** Is a new CLI verb for on-demand checkpoints acceptable, given it is itself a class-4 event?
- **Q3** [RESOLVED by the A1 draft `8f77d9a33`] A1 ships the AC counter Go port and verify's
  measured values; class 1 consumes them (spec.md §F O7).
- **Q4** Sync-audit retry ceiling source — no config key was located; is it the same tier map?
- **Q5** Does "recorded CI failure" have an existing on-disk producer, or is class 5's CI limb
  expected to stay not-observed until one exists?
- **Q6** A signed contract that turns `signed-invalid` mid-run for a non-acceptance reason
  (e.g. `contract_digest_mismatch`) is specified as not-armed (REQ-AE-020). Should it instead
  trip an escalation? The design-source ("change => contract void") leans that way, but no
  class token covers it and adding one would be a seventh `escalate_on` token owned by A1.
- **Q7** Which file set does the A1 `frozen-files` invariant resolve to (spec.md §F O2)?
