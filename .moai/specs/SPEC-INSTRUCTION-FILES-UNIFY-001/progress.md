# SPEC-INSTRUCTION-FILES-UNIFY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: L (5 artifacts + progress.md). Artifacts written: spec.md, plan.md, acceptance.md,
  design.md, research.md.
- SPEC ID regex check executed as Bash, output `PASS`.
- **v0.3.0 — B1/B2 carve applied (operator decision, 2026-09-26).** Requirements: **16**
  (`REQ-IFU-001~006`, `013~019`, `023~025`; Tier L ceiling 25). Acceptance criteria: **20**
  (ceiling 25). The nine requirements touching a user-owned file (`REQ-IFU-007~012`,
  `020~022`) and the seven criteria covering them transferred verbatim to
  **`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`** (card **t1259** owns its plan phase). Nothing was
  deleted. Ids were deliberately NOT renumbered — the gaps are the carve's footprint (spec.md
  HISTORY).
- **The carve's cause is arithmetic, not the recorded debt.** The plan-audit judged the debt on
  the two folded criteria acceptable at the 0.85 threshold. What forced the split is that its D2
  finding needs two new criteria while the SPEC stood at 25/25 with no tier above L.
- Frontmatter repaired to the canonical 12-field schema (audit MP-3): `tags` quoted as a
  comma-separated string, `module` and `lifecycle` added, `title` / `version` quoted. `moai spec
  lint` now exits `0`.
- Five criteria that were vacuously passable (audit D1) repaired against verified symbols, and
  every test-invoking criterion now asserts `--- PASS: <TestName>` under `-v` rather than exit
  `0` alone. `AC-IFU-016`'s guard does not exist in `internal/hook` (verified) and the criterion
  is written to require its creation.
- Two criteria added for the previously uncovered requirements (audit D2): `AC-IFU-026`
  (`REQ-IFU-001`, contract deployment) and `AC-IFU-027` (`REQ-IFU-005`, launcher invariance).
  §D.2's traceability table is now derived from criterion bodies and carries a re-runnable
  verification command instead of a coverage claim.
- `AC-IFU-022` gained `CONTRACT_HEAD` as its positive control (audit D5): its absence makes a
  run INCONCLUSIVE, so a render that captured nothing can no longer resolve to the benign
  branch.
- The `AGENTS.md` mirror-divergence figure re-measured (audit D4): 57 template-only / 17
  root-only lines against `553e224f3` on 2026-09-26. The t925-era "46 lines" is retired.
- Source locations converted to symbol anchors throughout (the audit's staleness warning:
  `origin/develop` is ~93 commits ahead and card t1224 already moved
  `frozenInstructionFiles`).
- `AC-IFU-025` gained the always-loaded-budget clause (audit item 6) — root `AGENTS.md` sits
  inside that surface, card t1175 is retuning it concurrently, and neither named ceiling would
  have caught an overrun.
- The §A.5 confirmed-branch transfer to card t1219 is now asserted as a conditional Definition
  of Done item (audit D7) rather than resting on prose in three files.
- Status: `draft`. Plan phase only — no implementation, no commits by this agent.
- Run phase is blocked on card t1175 landing on develop (plan.md §B).
- Two run-phase measurements remain mandatory and carry criteria whichever way they come out:
  AC-IFU-021 (real linked worktree, ancestor discovery) and AC-IFU-022 (Codex discovery of
  `AGENTS.local.md`, with head and tail sentinels so silent truncation is observable).
- AC-IFU-022's command verb was verified in the plan phase, not inherited: `codex debug --help`
  against **codex-cli 0.157.0** lists `prompt-input` ("Render the model-visible prompt input
  list as JSON"). A run under a different codex-cli version re-confirms before relying on it.
- AC-IFU-004 asserts the Codex-discovered **filename set** recursively over the template tree,
  not a path the criterion already expects — a `-maxdepth 1` or path-presence form would keep
  passing after a rename back to `AGENTS.md`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
