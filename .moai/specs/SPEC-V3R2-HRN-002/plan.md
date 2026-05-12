# Plan — SPEC-V3R2-HRN-002 Evaluator Memory Scope Amendment

> Phase 1B implementation plan. Based on research.md §8 gap analysis: ~95% greenfield work — only Sprint Contract durability is already in place. Plan organizes the FROZEN-zone amendment into 5 milestones with M5 carrying the CON-002 paperwork burden.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec (HRN-002 plan author) | Initial plan covering M1–M5 milestones for HRN-002. Mirrors SPC-001 plan structure (PR #870). |

---

## 1. Goal Recap

Insert a new `§11.4.1 Evaluator Memory Scope (Principle 4)` into the FROZEN `.claude/rules/moai/design/constitution.md`, declaring per-iteration ephemeral evaluator judgment memory and durable Sprint Contract state. Land the amendment via SPEC-V3R2-CON-002's 5-layer graduation protocol. Wire the corresponding config flag (`evaluator.memory_scope: per_iteration`) into both `design.yaml` and `harness.yaml`, extend `HarnessConfig` with a validator that rejects any non-`per_iteration` value with `HRN_EVAL_MEMORY_FROZEN`, and update evaluator-active agent body + moai-workflow-gan-loop SKILL.md with cross-references and iteration-handoff semantics. Source: Master §5.7 "single most important amendment" + Zhuge et al. 2024 R1 §9 (Agent-as-a-Judge fresh-memory finding).

## 2. Approach

Greenfield amendment shaped by SPC-001 precedent. Research confirms ~95% of work is net-new (Sprint Contract durability is the only pre-existing substrate). Plan tasks therefore consist of:

- Authoring the §11.4.1 constitution amendment text + version bump.
- Extending HarnessConfig with `Evaluator.MemoryScope` field and loader validator.
- Adding the config flag to both design.yaml and harness.yaml.
- Wiring the SKILL.md iteration-handoff step + agent body cross-reference.
- Adding the prior-judgment leak detection regression test.
- Filing CON-002 5-layer amendment evidence (FrozenGuard / Canary / ContradictionDetector / RateLimiter / HumanOversight) verbatim per SPC-001 pattern.
- Recording the amendment in `.moai/research/evolution-log.md` (core log) and `.moai/design/v3-research/evolution-log.md` (design cross-reference) per Decision D5.

## 3. Milestones

### M1 — Schema design + research delivery — Priority Critical

Owner: `manager-spec`
Deliverables:
- This plan.md, the research.md, the acceptance.md, the tasks.md, the progress.md, and the spec-compact.md as a 6-file plan bundle.
- Self-demonstrating hierarchical AC inside acceptance.md (REQ-HRN-002-XXX coverage via the SPC-001 schema).
- File:line anchor coverage (≥25 required, 50 achieved in research.md §10).

mx_plan tags:
- `@MX:NOTE` on research.md §11 D1-D10 decision block.
- `@MX:ANCHOR fan_in=3` on plan.md §3 M5 milestone (cross-referenced by HRN-003, CON-002 amendment instances).

Exit criteria:
- plan-auditor independent verification PASS at iteration ≤2.
- ≥25 file:line anchors in research.md (50 confirmed).
- All 19 REQs listed in spec.md §5 mapped to ≥1 AC in acceptance.md.

### M2 — Constitution amendment + HarnessConfig schema extension — Priority Critical

Owner: `manager-spec` (constitution edit), `expert-backend` (Go struct extension)

File:line anchors:
- `.claude/rules/moai/design/constitution.md:335` — insertion point for new `§11.4.1 Evaluator Memory Scope (Principle 4)` text (immediately after the Sprint Contract Protocol §11.4 closing rules line).
- `.claude/rules/moai/design/constitution.md:1-9` — HISTORY block appends amendment row.
- `.claude/rules/moai/design/constitution.md:405` — `Version: 3.4.0` → `Version: 3.5.0` (minor increment per amendment policy).
- `internal/config/types.go` — `HarnessConfig` struct (from HRN-001) gains `Evaluator EvaluatorConfig` sub-struct. New `EvaluatorConfig` struct declares `MemoryScope string` with `validate:"required,eq=per_iteration"` tag.
- `internal/config/loader.go` — `LoadHarnessConfig()` validator surfaces `HRN_EVAL_MEMORY_FROZEN` on any value other than `per_iteration`.

Tasks:
- Insert §11.4.1 verbatim text from spec.md §1.2 (lines 73-83 of spec.md).
- Append HISTORY row: `| 3.5.0 | 2026-XX-XX | HRN-002 | §11.4.1 Evaluator Memory Scope (Principle 4) inserted — per-iteration ephemeral judgment + durable Sprint Contract state |`.
- Bump version metadata at constitution.md:405.
- Extend `HarnessConfig` struct in `internal/config/types.go` with `Evaluator EvaluatorConfig` field.
- Declare new `EvaluatorConfig` struct: `MemoryScope string \`yaml:"memory_scope" validate:"required,eq=per_iteration"\``.
- Wire `HRN_EVAL_MEMORY_FROZEN` error sentinel in `internal/config/errors.go` (or wherever HRN-001 errors live).
- Update `LoadHarnessConfig()` and `LoadDesignConfig()` (if separate) to invoke validator.

mx_plan tags:
- `@MX:WARN reason="FROZEN-zone amendment per CON-002 §5 Layer 1 (Frozen Guard); modifications require full 5-layer cycle"` on the inserted §11.4.1 block.
- `@MX:NOTE` on `internal/config/types.go` `EvaluatorConfig.MemoryScope` field — "FROZEN at per_iteration per design-constitution §11.4.1 (SPEC-V3R2-HRN-002)".

Exit criteria:
- §11.4.1 text matches spec.md §1.2 verbatim (diff-byte-equal).
- Constitution version is 3.5.0.
- `go test ./internal/config/...` green with new validator tests covering `per_iteration` accept + `cumulative` reject + missing field reject (AC-HRN-002-04 leaf scenarios).
- No regression in existing HRN-001 loader tests.

### M3 — Agent body cross-reference + SKILL.md iteration-handoff + leak detection test — Priority High

Owner: `manager-docs` (agent body + SKILL.md) with `expert-backend` (leak detection test)

File:line anchors:
- `.claude/agents/moai/evaluator-active.md:91` — insert single cross-reference line immediately above `## Sprint Contract Negotiation` heading. Text: "Per design-constitution §11.4.1, evaluator judgment memory is ephemeral per iteration. The orchestrator MUST respawn evaluator-active via a fresh `Agent()` call at each GAN-loop iteration boundary; prior iteration's evaluator transcript MUST NOT appear in the new spawn prompt."
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:131-133` — insert new step between Phase 4 (Loop Decision) and Phase 5 (Iteration Feedback). Step title: "**Phase 4b: Iteration Handoff (REQ-HRN-002-009)**". Body declares fresh respawn with narrow 3-input prompt (BRIEF reference, Sprint Contract criterion states, artifact path).
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:147-158` — annotate stagnation-detection block: "Stagnation comparison reads score deltas from the durable Sprint Contract artifact, NOT from prior iteration's evaluator memory."
- `.claude/skills/moai-workflow-gan-loop/SKILL.md` — append a new "Solo Mode" subsection documenting the `{spec-id}/contract.yaml` alias (REQ-HRN-002-016).
- `internal/harness/evaluator_leak_test.go` — NEW Go integration test scanning evaluator spawn prompts for forbidden substrings (`Score:`, `Feedback:`, `Verdict:`, `Iteration N`). Fails with `HRN_EVAL_PRIOR_JUDGMENT_LEAK` per REQ-HRN-002-017.

Tasks:
- Insert evaluator-active cross-reference line.
- Insert SKILL.md Phase 4b iteration-handoff step.
- Annotate SKILL.md stagnation-detection block.
- Document `{spec-id}/contract.yaml` solo-mode alias.
- Author leak detection test fixture (mock spawn prompts with intentional `Score:` / `Feedback:` substrings) + assertion that the validator rejects them.

mx_plan tags:
- `@MX:NOTE` on agent body cross-reference line: "Cross-references design-constitution §11.4.1 (SPC-V3R2-HRN-002)".
- `@MX:NOTE` on SKILL.md Phase 4b step: "Per design-constitution §11.4.1; enforces REQ-HRN-002-009 fresh respawn".
- `@MX:WARN reason="prior-judgment leak detection"` on `internal/harness/evaluator_leak_test.go` test function.

Exit criteria:
- Agent body has exactly one new cross-reference line (no structural rewrite).
- SKILL.md Phase 4b step is present and references §11.4.1.
- `go test -run TestEvaluatorPriorJudgmentLeak ./internal/harness/...` green.
- All existing `internal/harness/...` tests remain green.

### M4 — Configuration propagation (design.yaml + harness.yaml + loader) — Priority High

Owner: `expert-backend`

File:line anchors:
- `.moai/config/sections/design.yaml:6-9` — under top-level `design:` namespace, insert sibling key `evaluator: \n  memory_scope: per_iteration` (sibling to `gan_loop:`, `brand_context:`, `claude_design:`, etc.).
- `.moai/config/sections/harness.yaml:5-8` — under top-level `harness:` namespace, insert sibling key `evaluator:\n  memory_scope: per_iteration` (sibling to `default_profile:`, `mode_defaults:`, `auto_detection:`, etc.).
- `internal/template/templates/.moai/config/sections/design.yaml` — template mirror (Template-First per CLAUDE.local.md §2).
- `internal/template/templates/.moai/config/sections/harness.yaml` — template mirror.
- `internal/config/loader_test.go` — fixtures covering valid `per_iteration` and invalid `cumulative` per AC-HRN-002-04.

Tasks:
- Add `evaluator.memory_scope: per_iteration` to both yaml files.
- Mirror in template directory.
- `make build` regenerates embedded files; verify byte-identical.
- Add loader test fixtures `internal/config/testdata/eval-memory-valid/harness.yaml` (value `per_iteration`) and `internal/config/testdata/eval-memory-frozen-violation/harness.yaml` (value `cumulative`).
- Assert validator returns `HRN_EVAL_MEMORY_FROZEN` on the violation fixture.
- Add a third fixture with missing `evaluator.memory_scope` key — confirm validator returns `HRN_EVAL_MEMORY_REQUIRED` (or whatever the canonical missing-field sentinel is in HRN-001 loader).

mx_plan tags:
- `@MX:NOTE` on the design.yaml insertion: "FROZEN at per_iteration per design-constitution §11.4.1; loader enforces via HRN_EVAL_MEMORY_FROZEN".
- `@MX:NOTE` on the harness.yaml insertion: same.

Exit criteria:
- Both YAML files contain the key with value `per_iteration`.
- Template mirrors byte-identical (verified by `make build`).
- Loader tests pass for both valid and invalid scenarios.
- `moai harness validate` (HRN-001 CLI subcommand) reports the new field correctly.

### M5 — REFACTOR + MX tags + completion gate (CON-002 paperwork) — Priority Critical

Owner (milestone-level): `manager-quality` for T-HRN002-10 (Canary) + T-HRN002-11 (MX tags); `manager-spec` for T-HRN002-12 (CON-002 evidence + completion gate). Per-task owner assignments are authoritative in `tasks.md`; this milestone header lists the union.

File:line anchors:
- `.moai/specs/SPEC-V3R2-HRN-002/spec.md:104-110` — §2.1 In Scope declares the CON-002 amendment graduation protocol invocation; M5 produces the evidence.
- `.moai/specs/SPEC-V3R2-HRN-002/con-002-amendment-evidence.md` — NEW (mirrors SPC-001 file at `.moai/specs/SPEC-V3R2-SPC-001/con-002-amendment-evidence.md`).
- `.moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt` — NEW Canary log (mirrors SPC-001 `canary-v2-reparse.txt`).
- `.moai/research/evolution-log.md` — NEW or appended (CON-002 core log).
- `.moai/design/v3-research/evolution-log.md` — NEW or appended (design subsystem cross-reference per Decision D5).
- `.moai/specs/SPEC-V3R2-HRN-002/progress.md` — final status `complete` after run-phase Canary + HumanOversight approval.
- `.claude/rules/moai/core/zone-registry.md` — append new `CONST-V3R2-NNN` entry referencing §11.4.1 as Frozen + canary_gate true (Decision D7 symmetric defense).

Tasks:
- **FrozenGuard evidence (Layer 1)**: confirm HRN-002 amendment is strictly additive (no existing rule modified; only new §11.4.1 inserted). Cite Decision D7 (symmetric defense via loader + FrozenGuard). Document in evidence block.
- **Canary evidence (Layer 2)**: select 3 past GAN-loop projects (or design completions) — if v3 corpus is empty, fall back to v2-legacy design completions or document `CanaryUnavailable` per REQ-CON-002-020. Simulate the amendment in shadow mode (run last 3 evaluator iterations under the new respawn protocol via test fixtures), verify no project score drops >0.10. Capture script output to `canary-fresh-memory-eval.txt`.
- **ContradictionDetector evidence (Layer 3)**: scan existing rules for conflicts. Expected verdict: none (§11.4 amendment is additive; the only related FROZEN clauses are §11 ordering FROZEN — unchanged, and pass_threshold floor — unchanged).
- **RateLimiter evidence (Layer 4)**: inventory v3.x FROZEN amendments used so far (1 of 3 used by SPC-001 per PR #870; HRN-002 would be #2). Confirm within cap.
- **HumanOversight evidence (Layer 5)**: AskUserQuestion bundles R1 §9 citation + canary verdict + Principle 4 reference + research.md §8 gap table as Option 1 ("(권장) Approve with full evidence reviewed"). Record approval timestamp + reviewer (Goos Kim) in landing PR description.
- Add `@MX:ANCHOR fan_in=3` to acceptance.md self-demonstrating example (referenced by HRN-003 hierarchical scoring + future Sprint Contract SPEC + CON-002 amendment pattern catalog).
- Add `@MX:NOTE` on `internal/config/types.go` `EvaluatorConfig.MemoryScope` field (from M2).
- Add `@MX:WARN reason="FROZEN-zone amendment"` on the §11.4.1 insertion point (from M2).
- Write evolution-log entry to both core + design logs per Decision D5.
- Register new CONST-V3R2-NNN entry in zone-registry.md for §11.4.1.

mx_plan tags:
- `@MX:ANCHOR fan_in=3` on acceptance.md self-demonstrating example (HRN-003 + Sprint Contract + CON-002 catalog).
- `@MX:NOTE` on `internal/config/types.go EvaluatorConfig` (from M2 — confirmed in M5 audit).
- `@MX:WARN reason="FROZEN-zone amendment per CON-002 §5 Layer 1 (Frozen Guard); modifications require full Canary + HumanOversight cycle"` on the §11.4.1 insertion point.

Exit criteria:
- progress.md `plan_status: complete` (after run-phase merge).
- `con-002-amendment-evidence.md` lists all 5 layers with verdict (4/5 PASS expected; Layer 5 PENDING-FINAL at PR open).
- `canary-fresh-memory-eval.txt` committed.
- `.moai/research/evolution-log.md` + `.moai/design/v3-research/evolution-log.md` both have EVO-HRN-002 entry (per REQ-HRN-002-010, Decision D5).
- `.claude/rules/moai/core/zone-registry.md` has new CONST-V3R2-NNN entry for §11.4.1.
- HumanOversight approval recorded in landing PR description.

---

## 4. Technical Approach

### 4.1 Stack & touch points

- Constitution rule files: `.claude/rules/moai/design/constitution.md` (§11.4.1 insertion + version bump + HISTORY), `.claude/rules/moai/core/zone-registry.md` (new CONST entry).
- Skill file: `.claude/skills/moai-workflow-gan-loop/SKILL.md` (Phase 4b iteration-handoff step + Solo Mode subsection).
- Agent file: `.claude/agents/moai/evaluator-active.md` (single cross-reference line at line 91).
- Go config subsystem: `internal/config/types.go` (HarnessConfig extension), `internal/config/loader.go` (validator wiring), `internal/config/errors.go` (sentinel), `internal/config/loader_test.go` (fixtures).
- Go harness module: `internal/harness/evaluator_leak_test.go` (new leak detection test, REQ-HRN-002-017).
- Config sections: `.moai/config/sections/design.yaml`, `.moai/config/sections/harness.yaml` (new evaluator.memory_scope key).
- Template mirror: `internal/template/templates/.moai/config/sections/{design,harness}.yaml`.
- Evolution logs: `.moai/research/evolution-log.md` (core), `.moai/design/v3-research/evolution-log.md` (design cross-reference).

### 4.2 Risk register (carried from research §9)

| # | Risk | Severity | Mitigation milestone |
|---|------|----------|-----------------------|
| R1 | <3 design projects in canary corpus | MEDIUM | M5 — fall back to v2-legacy or `CanaryUnavailable` |
| R2 | Fresh respawn overhead | MEDIUM | M2 — measure baseline; document if >5% |
| R3 | Human reviewer rubber-stamps without reading evidence | HIGH | M5 — bundle 4-component evidence in AskUser Option 1 |
| R4 | v2 sessions persist memory through resume | MEDIUM | M5 — BC-V3R2-010 release-notes entry + session-start hook retires v2 sessions |
| R5 | Sprint Contract state grows unbounded | LOW | M3 — document archival on SPEC completion |
| R6 | Subtle prior-context leak via Sprint Contract serialization | HIGH | M3 — leak detection test (REQ-HRN-002-017) |
| R7 | Rate-limiter quota exhausted | LOW | M5 — burst allowance per CON-002 design |
| R8 | Evolution-log write failure | MEDIUM | M5 — blocking REQ-HRN-002-010 |
| R9 | Third-party evaluator plugin breakage | LOW | M5 — release-notes notice |
| R10 | `memory: project` vs LLM-context confusion | MEDIUM | M2/M3 — amendment text + agent body cross-ref clarify |
| R11 | Solo-mode contract alias undocumented | LOW | M3 — SKILL.md documents alias |
| R12 | Dual evolution-log path disambiguation | MEDIUM | M1/M5 — Decision D5 codifies dual-write; M5 records both |

### 4.3 Dependencies

Blocked by (all landed on main):
- SPEC-V3R2-CON-001 — FROZEN/EVOLVABLE zone model + zone-registry.md infrastructure.
- SPEC-V3R2-CON-002 — 5-layer amendment graduation protocol.
- SPEC-V3R2-HRN-001 — `HarnessConfig` struct + loader.

Blocks:
- SPEC-V3R2-HRN-003 — hierarchical per-leaf scoring; depends on fresh-judgment semantics.
- SPEC-V3R2-WF-003 — multi-mode router; thorough harness uses Sprint Contract per HRN-002.

Related (no hard dependency):
- SPEC-V3R2-SPC-001 — recently merged (PR #870); HRN-002 reuses the CON-002 paperwork pattern + hierarchical AC schema.
- SPEC-DESIGN-CONST-AMEND-001 — v2-legacy amendment precedent; pattern source.
- SPEC-V3R2-EVAL-001 — v3-legacy evaluator profile schema (indirect consumer of memory_scope flag).
- SPEC-V3R2-EXT-004 — versioned migration auto-apply (BC-V3R2-010 session retirement).

### 4.4 Out-of-scope confirmations

Per spec.md §2.2 and research §11 decisions:
- Sprint Contract YAML schema changes (D4 — durability mechanics already encoded in existing §11.4; HRN-002 only enforces evaluator absence).
- Hierarchical scoring implementation (HRN-003).
- Per-SPEC evaluator profile overrides (harness-level only).
- Telemetry/metrics dashboards (deferred to Master §12 Open Question #3).
- Changing `max_iterations` (FROZEN at 5).
- Changing `escalation_after` (FROZEN at 3).
- Changing actor (expert-frontend) body.
- Cross-SPEC evaluator memory sharing (always SPEC-scoped).
- Allowing `cumulative` opt-in (FROZEN value; would require a new amendment cycle).
- Changing evaluator model (sonnet preserved).
- Go-side `internal/harness/gan_loop.go` runner module (D1 — runner enforcement via SKILL.md + leak test).
- Performance budget for spawn overhead (D8 — documented but not gated).

---

## 5. Quality Gates

### 5.1 Plan-phase gates (this PR)

- [ ] `plan-auditor` independent verification — PASS at iteration ≤2.
- [ ] research.md ≥25 file:line anchors — achieved 50.
- [ ] Every REQ in spec.md §5 mapped to ≥1 AC in acceptance.md (19 REQs verified).
- [ ] Every AC in acceptance.md cites ≥1 REQ on every leaf.
- [ ] tasks.md uses `T-HRN002-NN` naming with owner role per task.
- [ ] progress.md frontmatter `plan_status: audit-ready`.
- [ ] No `spec.md` modifications in this PR (read-only) — spec.md remains v0.1.0.
- [ ] ≥3 ACs in acceptance.md self-demonstrate hierarchical schema.

### 5.2 Run-phase gates (future PR, scoped by tasks.md)

- [ ] §11.4.1 amendment text byte-identical to spec.md §1.2.
- [ ] Constitution version bumped to 3.5.0; HISTORY appended.
- [ ] `go test ./internal/config/...` green with new validator tests.
- [ ] `go test -run TestEvaluatorPriorJudgmentLeak ./internal/harness/...` green.
- [ ] design.yaml + harness.yaml both have `evaluator.memory_scope: per_iteration`.
- [ ] Template mirrors byte-identical after `make build`.
- [ ] Canary evidence committed to `canary-fresh-memory-eval.txt`.
- [ ] CON-002 evidence committed to `con-002-amendment-evidence.md` (5 layers documented).
- [ ] Dual evolution-log entries written (core + design).
- [ ] zone-registry.md CONST entry added for §11.4.1.
- [ ] HumanOversight approval recorded in landing PR description.
- [ ] TRUST 5: Tested / Readable / Unified / Secured / Trackable all green.

### 5.3 Sync-phase gates

- [ ] CHANGELOG entry under `### Changed` referencing BC-V3R2-010 + R1 §9 citation.
- [ ] docs-site 4-language sync (per CLAUDE.local.md §17) for the design-constitution amendment reference (if a public docs page exposes §11).
- [ ] zone-registry.md cross-link updated.

---

## 6. Rollout Strategy

This is an **additive amendment** at the runtime layer (existing §11.4 text unchanged; new §11.4.1 inserted directly below), but a **schema-level semantic change** at the contract layer (evaluator memory scope was implicitly cumulative; becomes explicitly per-iteration). The `breaking: true` flag in spec.md frontmatter + BC-V3R2-010 declaration in Master §8 reflect this schema evolution.

Rollout proceeds as:

1. Plan PR (this) — research/plan/AC/tasks/progress/spec-compact artifacts only.
2. plan-auditor iteration → approval.
3. Plan PR merge.
4. Run PR — M2 (constitution amendment + HarnessConfig extension) + M3 (agent/SKILL/leak test) + M4 (config keys + loader validator).
5. Run PR merge.
6. Sync PR — M5 (CON-002 paperwork + evolution-log + zone-registry CONST entry + CHANGELOG).
7. HRN-003 (hierarchical scoring) plan kicks off, consuming the fresh-memory semantic.

Rollback: if the new respawn protocol causes regression (e.g., quality scores drop >0.10 on canary), revert by:
- Re-removing §11.4.1 text from constitution.md (CON-002-handled).
- Reverting `evaluator.memory_scope` keys from yaml files.
- Reverting HarnessConfig.Evaluator.MemoryScope field.
- Recording rollback in evolution-log (per CON-002 §5 Layer 5 Evolution Rollback).
- Marking BC-V3R2-010 as rolled-back; downstream HRN-003 plan halts.

Reverting the agent body cross-reference and SKILL.md Phase 4b text is recommended for cleanliness but not strictly required for the system to function (those are documentation; the runtime gate is the loader validator).

---

## 7. Resource Estimates (priority-based; no time predictions per agent-common-protocol)

| Milestone | Priority | Files touched (estimate) | LOC delta (estimate) |
|-----------|----------|--------------------------|----------------------|
| M1 | Critical | 6 (this plan bundle) | +1700 |
| M2 | Critical | 5 (constitution.md, types.go, loader.go, errors.go, loader_test.go) | +300 |
| M3 | High | 4 (evaluator-active.md, SKILL.md, evaluator_leak_test.go, leak test fixture) | +200 |
| M4 | High | 6 (design.yaml, harness.yaml, 2 template mirrors, 3 test fixtures, loader_test.go additions) | +120 |
| M5 | Critical | 5 (evidence.md, canary log, 2 evolution logs, zone-registry CONST entry, progress.md update) | +500 |

Total estimated delta across all milestones: ~2820 LOC of which ~1700 is plan documentation and ~1120 is code/config/test/paperwork additions.

---

## 8. Communication & Handoff

### 8.1 To plan-auditor

- Read research.md §8 gap analysis first — answers "is this work redundant with anything on main?" Answer: 95% greenfield.
- Read research.md §10 file:line anchors for grounded claims (50 anchors confirmed).
- Read research.md §11 decisions for the irreversible design choices baked into the plan (D1–D10).
- Verify acceptance.md self-demonstrates hierarchical schema on ≥3 ACs (per Decision D4).
- Confirm tasks.md owner roles align with manager-spec / expert-backend / manager-docs / manager-quality split.
- Verify REQ-to-AC traceability is 100% (every REQ in spec.md §5 has at least one AC entry).

### 8.2 To run-phase agent (manager-tdd or expert-backend)

- Start from M2 (constitution amendment) — highest-leverage, lowest-risk task. Insert §11.4.1 verbatim from spec.md §1.2.
- Then M2 Go-side (HarnessConfig extension + validator). Lands the loader-side defense (Decision D7).
- Then M3 (agent body + SKILL.md + leak test). Documents and enforces the runtime semantic.
- Then M4 (yaml config keys + loader fixtures). Activates the validator.
- Then M5 (CON-002 paperwork). Mirror SPC-001 pattern verbatim.
- Caveat: M2 constitution edit triggers FrozenGuard (Decision D7). Use the CON-002 amendment path documented in CON-002 spec.md §2.1; do NOT perform the edit as a raw file write.

### 8.3 To dependents (HRN-003, WF-003)

- HRN-003 plan-phase agent: hierarchical per-leaf scoring assumes fresh-judgment per leaf. HRN-002 ships the fresh-respawn protocol; HRN-003 builds the per-leaf scoring on top.
- WF-003 plan-phase agent: multi-mode router thorough mode requires Sprint Contract per HRN-002. HRN-002 ships the durable substrate (via existing §11.4 text + new §11.4.1 clarification).
- Both downstream SPECs may begin scoping in parallel after HRN-002 plan PR merges; run-phase coordination on main.

---

## 9. Done Definition

This plan is `audit-ready` when:
- All 6 plan files exist (`spec.md`, `research.md`, `plan.md`, `acceptance.md`, `tasks.md`, `progress.md`, `spec-compact.md`).
- plan-auditor PASS at iteration ≤2.
- PR `plan(spec): SPEC-V3R2-HRN-002 — Evaluator Memory Scope Amendment plan artifacts` opened against main.
- progress.md frontmatter `plan_status: audit-ready`.

End of plan.
