# Acceptance Criteria — SPEC-V3R2-HRN-002 Evaluator Memory Scope Amendment

> Self-demonstrating: this file uses the **hierarchical** AC format that SPC-001 introduces (parent → `.a/.b/.c` children inheriting parent Given). At least three ACs (AC-HRN-002-01, AC-HRN-002-04, AC-HRN-002-07) are authored as parent/child trees, recursively proving the SPC-001 schema works for HRN-002's own paperwork.
>
> Flat ACs from spec.md §6 remain canonical and unchanged; this file augments them with hierarchical Given/When/Then breakdown for each REQ as plan-phase deliverable.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec (HRN-002 plan author) | Initial hierarchical AC authoring covering all 19 REQs across 5 EARS modalities. Self-demonstrates SPC-001 schema. |

---

## 1. Format Conventions (this file)

- Top-level: `AC-HRN-002-NN` (e.g., `AC-HRN-002-01`).
- Depth-1 children: `.a / .b / .c` (lowercase letters).
- Depth-2 grandchildren: `.a.i / .a.ii` (lowercase Roman).
- Maximum depth = 3 (top + 2 child levels), enforced by `internal/spec/ears.go:21` `MaxDepth = 3`.
- Each leaf carries `(maps REQ-...)`. Intermediate nodes MAY omit when all leaves carry a tail.
- Children inherit parent's Given when the child's own Given is empty (`internal/spec/ears.go:77-81`).

## 1.1 Drift Reconciliation Notes (post-SPC-001 merge, 2026-05-13)

`spec.md` was authored 2026-04-23 against `.claude/rules/moai/design/constitution.md` v3.3.0. Between authorship and plan delivery (2026-05-13), SPEC-V3R2-SPC-001 (PR #870, main `07dabe011`) advanced the constitution to v3.4.0. `spec.md` remains FROZEN per CON-001 zone model; this plan adapts the deltas downstream.

- **Version delta**: spec.md REQ-HRN-002-002 / AC-HRN-002-02 cite "Version: 3.3.0 → 3.4.0". This plan targets `Version: 3.4.0 → 3.5.0` (the next bump after the already-landed v3.4.0). The intent of REQ-HRN-002-002 (HISTORY entry + version bump for the evaluator memory amendment) is preserved; only the version numerals are adapted to current main state. AC-HRN-002-02 verbiage in this acceptance.md (`3.4.0 → 3.5.0`) supersedes the stale spec.md figure for run-phase verification.
- **SKILL.md path**: spec.md §3 module list cites `.claude/skills/moai/moai-workflow-gan-loop/SKILL.md`. The actual path on main is `.claude/skills/moai-workflow-gan-loop/SKILL.md` (no intermediate `/moai/` subdirectory). plan.md, research.md, and tasks.md use the correct path; spec.md path is preserved verbatim per FROZEN rule.
- **No further drift expected**: research.md §11 D1–D10 captured all other plan-time decisions; this reconciliation block is the only deviation from spec.md and is bounded by the two items above.

## 2. REQ ↔ AC Traceability Matrix

| REQ ID | EARS modality | Mapped AC(s) | Notes |
|--------|---------------|--------------|-------|
| REQ-HRN-002-001 | Ubiquitous | AC-HRN-002-01 | §11.4.1 text presence |
| REQ-HRN-002-002 | Ubiquitous | AC-HRN-002-02 | Version bump 3.4.0 → 3.5.0 |
| REQ-HRN-002-003 | Ubiquitous | AC-HRN-002-03 | design.yaml key |
| REQ-HRN-002-004 | Ubiquitous | AC-HRN-002-03 | harness.yaml key |
| REQ-HRN-002-005 | Ubiquitous | AC-HRN-002-04 | HarnessConfig struct extension |
| REQ-HRN-002-006 | Ubiquitous | AC-HRN-002-05 | Agent body cross-ref |
| REQ-HRN-002-007 | Ubiquitous | AC-HRN-002-06 | SKILL.md update |
| REQ-HRN-002-008 | Event-Driven | AC-HRN-002-09 | CON-002 5-layer execution |
| REQ-HRN-002-009 | Event-Driven | AC-HRN-002-07 | Fresh respawn at iteration start |
| REQ-HRN-002-010 | Event-Driven | AC-HRN-002-09, AC-HRN-002-12 | evolution-log entry |
| REQ-HRN-002-011 | Event-Driven | AC-HRN-002-04 | Loader validator `HRN_EVAL_MEMORY_FROZEN` |
| REQ-HRN-002-012 | State-Driven | AC-HRN-002-08 | Sprint Contract state durability |
| REQ-HRN-002-013 | State-Driven | AC-HRN-002-07 | Judgment memory volatility |
| REQ-HRN-002-014 | State-Driven | AC-HRN-002-04 | FROZEN value preservation |
| REQ-HRN-002-015 | Optional | AC-HRN-002-09 | Canary verdict in AskUser record |
| REQ-HRN-002-016 | Optional | AC-HRN-002-13 | Solo-mode `{spec-id}/contract.yaml` alias |
| REQ-HRN-002-017 | Unwanted | AC-HRN-002-07 | Prior-judgment leak rejected |
| REQ-HRN-002-018 | Unwanted | AC-HRN-002-04 | `cumulative` value rejected |
| REQ-HRN-002-019 | Unwanted | AC-HRN-002-10 | FrozenGuard blocks unauthorized §11.4.1 edits |

Coverage check: 19 REQs declared in spec.md §5 (REQ-HRN-002-001 through 019), every REQ has ≥1 mapped AC entry above. Total ACs in this file: 13 (11 carried from spec.md §6 + 2 plan-phase additions — AC-HRN-002-12 for evolution-log dual-write per Decision D5, AC-HRN-002-13 for solo-mode alias per REQ-016).

---

## 3. Acceptance Criteria (hierarchical)

### 3.1 §11.4.1 amendment text presence (hierarchical — 3 children)

<!-- @MX:ANCHOR fan_in=3 -->
<!-- @MX:REASON: "self-demonstrating hierarchical AC reference cited by HRN-003 + Sprint Contract + CON-002 catalog" -->
- AC-HRN-002-01: Given the FROZEN file `.claude/rules/moai/design/constitution.md` after M2 lands
  - AC-HRN-002-01.a: When the reader scans for section `§11.4.1 Evaluator Memory Scope (Principle 4)`, Then the section exists immediately after the §11.4 Sprint Contract Protocol closing rules (insertion site `.claude/rules/moai/design/constitution.md:335`). (maps REQ-HRN-002-001)
  - AC-HRN-002-01.b: When the reader compares the §11.4.1 body text against spec.md §1.2 verbatim block (spec.md lines 73-83), Then the text is byte-identical. (maps REQ-HRN-002-001)
  - AC-HRN-002-01.c: When the reader walks the four-paragraph structure of §11.4.1, Then they find: (i) ephemeral judgment declaration, (ii) durable Sprint Contract declaration, (iii) implementation (fresh respawn) declaration, (iv) configuration (FROZEN per_iteration) declaration. (maps REQ-HRN-002-001)

### 3.2 Constitution version + HISTORY

- AC-HRN-002-02: Given the constitution file post-M2, When the reader scans the version footer (`.claude/rules/moai/design/constitution.md:405`), Then it reads `Version: 3.5.0`. When the reader scans the HISTORY block (lines 1-9), Then a new row exists with version `3.5.0`, date, and description citing "§11.4.1 Evaluator Memory Scope (Principle 4) inserted — per-iteration ephemeral judgment + durable Sprint Contract state". (maps REQ-HRN-002-002)

### 3.3 Configuration key presence (both yaml files)

- AC-HRN-002-03: Given the config files post-M4, When the reader greps for `memory_scope` in `.moai/config/sections/design.yaml` AND `.moai/config/sections/harness.yaml`, Then both files contain `evaluator:\n  memory_scope: per_iteration` under their top-level namespace. When the reader checks the template mirrors at `internal/template/templates/.moai/config/sections/{design,harness}.yaml`, Then both mirrors are byte-identical after `make build`. (maps REQ-HRN-002-003, REQ-HRN-002-004)

### 3.4 HarnessConfig schema + loader validator (hierarchical — 3 children)

- AC-HRN-002-04: Given the Go config subsystem post-M2 + M4
  - AC-HRN-002-04.a: When `LoadHarnessConfig()` is called with a fixture where `evaluator.memory_scope: per_iteration`, Then it returns a valid `*HarnessConfig` with `cfg.Evaluator.MemoryScope == "per_iteration"`. (maps REQ-HRN-002-005)
  - AC-HRN-002-04.b: When `LoadHarnessConfig()` is called with a fixture where `evaluator.memory_scope: cumulative`, Then it returns a wrapped error containing the sentinel `HRN_EVAL_MEMORY_FROZEN`. (maps REQ-HRN-002-011, REQ-HRN-002-018, REQ-HRN-002-014)
  - AC-HRN-002-04.c: When `LoadHarnessConfig()` is called with a fixture missing the `evaluator.memory_scope` key entirely, Then it returns a wrapped error containing the required-field sentinel (per HRN-001 loader convention). (maps REQ-HRN-002-005)

### 3.5 evaluator-active cross-reference

- AC-HRN-002-05: Given `.claude/agents/moai/evaluator-active.md` post-M3, When the reader scans the body for a reference to `§11.4.1`, Then exactly one line near the `## Sprint Contract Negotiation` section (around line 91) declares that evaluator judgment memory is ephemeral per iteration and the orchestrator MUST respawn evaluator-active at each iteration boundary. (maps REQ-HRN-002-006)

### 3.6 moai-workflow-gan-loop SKILL.md update

- AC-HRN-002-06: Given `.claude/skills/moai-workflow-gan-loop/SKILL.md` post-M3, When the reader scans for "Phase 4b" or "Iteration Handoff", Then a new step exists between Phase 4 (Loop Decision) and Phase 5 (Iteration Feedback) that declares fresh-respawn semantics with a narrow 3-input prompt (BRIEF reference, Sprint Contract criterion states, artifact path). When the reader scans for a citation, Then the step references `design-constitution §11.4.1` verbatim. (maps REQ-HRN-002-007)

### 3.7 Fresh respawn behavior + prior-judgment leak detection (hierarchical — 3 children + 2 grandchildren)

- AC-HRN-002-07: Given the GAN loop runner post-M3 leak-detection test landing
  - AC-HRN-002-07.a: When `internal/harness/evaluator_leak_test.go` `TestEvaluatorPriorJudgmentLeak` is invoked with a synthetic spawn prompt containing forbidden substrings (`Score:`, `Feedback:`, `Verdict:`), Then the test assertion fails the prompt construction with `HRN_EVAL_PRIOR_JUDGMENT_LEAK` error. (maps REQ-HRN-002-017)
    - AC-HRN-002-07.a.i: When the spawn prompt contains substring `Iteration N` referring to a numbered prior iteration, Then the validator flags it as a leak. (maps REQ-HRN-002-017)
    - AC-HRN-002-07.a.ii: When the spawn prompt contains a paraphrased rationale (e.g., "the previous evaluator noted..."), Then the validator flags it as a leak via heuristic substring scan. (maps REQ-HRN-002-017)
  - AC-HRN-002-07.b: When the test runs with a clean spawn prompt containing only BRIEF + Sprint Contract criterion states + artifact path, Then the validator returns no error and the test passes. (maps REQ-HRN-002-009, REQ-HRN-002-013)
  - AC-HRN-002-07.c: When 3 consecutive iterations are simulated under the new respawn protocol via fixtures, Then each iteration's spawn prompt is independent and the third iteration's prompt MUST NOT contain any substring lifted from iterations 1 or 2's evaluator output. (maps REQ-HRN-002-009, REQ-HRN-002-013)

### 3.8 Sprint Contract state durability

- AC-HRN-002-08: Given a Sprint Contract file at `.moai/sprints/{team-id}/contract.yaml` (or solo-mode `{spec-id}/contract.yaml`), When the GAN loop runs 3 iterations with at least one passed criterion in iteration 1, Then iteration 3's contract YAML still carries the passed criterion as `status: passed` and the criterion's pass state has not regressed without explicit human override. When a previously failed criterion has been refined, Then the refined criterion status is `refined` (not regressed to `passed` without test evidence). (maps REQ-HRN-002-012)

### 3.9 CON-002 5-layer execution + Canary attachment

- AC-HRN-002-09: Given the CON-002 amendment graduation protocol invoked for this SPEC, When the protocol runs end-to-end in run-phase, Then `.moai/specs/SPEC-V3R2-HRN-002/con-002-amendment-evidence.md` exists with 5 layer sections (FrozenGuard / Canary / ContradictionDetector / RateLimiter / HumanOversight) each carrying verdict + supporting evidence. When the Canary layer runs successfully with ≥3 design subjects, Then the verdict (score delta summary) appears in the AskUserQuestion Option 1 description per REQ-HRN-002-015. When the Canary layer cannot find ≥3 subjects, Then the verdict is `CanaryUnavailable` and human override is the only path forward per REQ-CON-002-020. (maps REQ-HRN-002-008, REQ-HRN-002-010, REQ-HRN-002-015)

### 3.10 FrozenGuard blocks unauthorized §11.4.1 edits

- AC-HRN-002-10: Given the §11.4.1 text landed post-M2 and a registered CONST-V3R2-NNN entry in `.claude/rules/moai/core/zone-registry.md` with `canary_gate: true`, When an unauthorized PR attempts to modify §11.4.1 via plain file write (without invoking the CON-002 graduation protocol), Then FrozenGuard (from SPEC-V3R2-CON-001 + CON-002) rejects the write with `CON_FROZEN_GUARD_REJECTED`. (maps REQ-HRN-002-019)

### 3.11 v2 evaluator session upgrade

- AC-HRN-002-11: Given a v2 evaluator session with cumulative-memory state persists at upgrade time, When v3.0.0-beta.1 starts and the session-start hook executes the BC-V3R2-010 migration (AUTO per Master §8), Then any in-flight v2 evaluator session is retired and a log entry `EVAL_SESSION_UPGRADED` appears in the session log. Existing Sprint Contract files at `.moai/sprints/` remain intact (cross-iteration substrate is unaffected). (maps REQ-HRN-002-014, BC-V3R2-010)

### 3.12 Dual evolution-log entry

- AC-HRN-002-12: Given the M5 paperwork completes, When the reader scans `.moai/research/evolution-log.md` (CON-002 core log), Then it contains an `EVO-HRN-002` entry with fields `{id, timestamp, before_snippet, after_snippet, canary_verdict, approver, rationale_cite: "R1 §9 + Principle 4"}` per REQ-HRN-002-010. When the reader scans `.moai/design/v3-research/evolution-log.md` (design subsystem cross-reference per Decision D5), Then it contains a cross-reference entry with `core_log_ref: EVO-HRN-002` pointer. (maps REQ-HRN-002-010)

### 3.13 Solo-mode contract alias

- AC-HRN-002-13: Given the GAN loop runs in solo mode (no team ID), When the orchestrator looks up the Sprint Contract location, Then `.moai/sprints/{spec-id}/contract.yaml` is accepted as equivalent to `.moai/sprints/{team-id}/contract.yaml`. When the SKILL.md is read post-M3 update, Then the alias is documented in a dedicated "Solo Mode" subsection. (maps REQ-HRN-002-016)

---

## 4. Edge Cases

| # | Scenario | Expected behaviour | Anchor |
|---|----------|--------------------|--------|
| E1 | Both `team-id` and `spec-id` contract files exist | Solo-mode `{spec-id}/contract.yaml` is the authoritative source; runner logs warning and ignores `{team-id}/contract.yaml` if both present | `.claude/skills/moai-workflow-gan-loop/SKILL.md` M3 Solo Mode subsection |
| E2 | `evaluator.memory_scope` declared at per-level (e.g., `levels.thorough.evaluator.memory_scope`) rather than top-level | Loader treats per-level declaration as schema violation; validator errors with `HRN_EVAL_MEMORY_INVALID_SCOPE` | `internal/config/loader.go` validator |
| E3 | Two consecutive iterations produce identical scores (no progress, no regression) | Stagnation detection (REQ-HRN-002-013 ancillary) fires via Sprint Contract score delta, NOT via evaluator memory comparison | `.claude/skills/moai-workflow-gan-loop/SKILL.md:147-158` |
| E4 | Evaluator-active SubagentStop hook fires while iteration is mid-flight | Hook handler is idempotent; logs completion regardless of mid-flight state; no state leak between iterations because each iteration is a fresh `Agent()` spawn | `.claude/agents/moai/evaluator-active.md:20-25` |
| E5 | YAML `memory_scope` value uses non-canonical case (e.g., `Per_Iteration`, `PER_ITERATION`) | Validator strict equality match against `per_iteration` (lowercase); other case forms are rejected with `HRN_EVAL_MEMORY_FROZEN` | `internal/config/types.go EvaluatorConfig.MemoryScope validate:"eq=per_iteration"` |
| E6 | Sprint Contract YAML carries forward a stale `negotiation_history` entry from an iteration that was rolled back | The Sprint Contract artifact persists; rollback records a new `iteration_status: rolled_back` row without removing prior rows; evaluator reads canonical state from latest contract revision | `.claude/skills/moai-workflow-gan-loop/SKILL.md:210-240` |
| E7 | <3 design projects exist in `.moai/design/` for Canary | Per REQ-CON-002-020, Canary emits `CanaryUnavailable`; AskUserQuestion still runs but Option 1 description carries the "Canary insufficient corpus" notice; user can override | `.moai/specs/SPEC-V3R2-CON-002/spec.md:111-113` |

---

## 5. Quality Gate Criteria (Definition of Done)

### 5.1 Plan-phase DoD (this PR)

- [ ] All 13 ACs above carry at least one `(maps REQ-...)` reference on every leaf.
- [ ] At least three ACs (AC-HRN-002-01, AC-HRN-002-04, AC-HRN-002-07) self-demonstrate the hierarchical schema with explicit `.a/.b/.c` children.
- [ ] AC-HRN-002-07 demonstrates depth-2 nesting (`.a.i`, `.a.ii`) to prove the MaxDepth=3 schema works.
- [ ] Every REQ in spec.md §5 (REQ-HRN-002-001 through REQ-HRN-002-019) appears at least once in §2 traceability matrix.
- [ ] Plan-auditor PASS at iteration ≤2.

### 5.2 Run-phase DoD (future PR per tasks.md)

- [ ] §11.4.1 text byte-identical to spec.md §1.2.
- [ ] Constitution version 3.5.0; HISTORY appended.
- [ ] `go test ./internal/config/...` green (covers AC-HRN-002-04.a/b/c).
- [ ] `go test -run TestEvaluatorPriorJudgmentLeak ./internal/harness/...` green (covers AC-HRN-002-07).
- [ ] design.yaml + harness.yaml both contain `evaluator.memory_scope: per_iteration`.
- [ ] Template mirrors byte-identical after `make build`.
- [ ] Sprint Contract durability test fixture covers AC-HRN-002-08 (3 iterations, 1 passed criterion carries forward).
- [ ] CON-002 evidence committed (5 layers documented in `con-002-amendment-evidence.md`).
- [ ] Canary log committed at `canary-fresh-memory-eval.txt`.
- [ ] Dual evolution-log entries (core + design) written per AC-HRN-002-12.
- [ ] zone-registry CONST entry added for §11.4.1 (canary_gate: true per Decision D7).
- [ ] HumanOversight approval recorded in landing PR description.

### 5.3 Sync-phase DoD

- [ ] CHANGELOG entry under `### Changed` cites BC-V3R2-010 + R1 §9 Agent-as-a-Judge.
- [ ] CHANGELOG entry documents the v2 → v3 session retirement behavior (AC-HRN-002-11).
- [ ] zone-registry.md cross-link complete.
- [ ] docs-site 4-language sync if §11 is publicly exposed (CLAUDE.local.md §17).

---

## 6. Self-Demonstration Notice

This file uses the new hierarchical AC schema (depth 0 → 1 → 2) on AC-HRN-002-01, AC-HRN-002-04, AC-HRN-002-07. Specifically:

- **AC-HRN-002-01** has 3 depth-1 children covering insertion site, byte-identical text, and 4-paragraph structure.
- **AC-HRN-002-04** has 3 depth-1 children covering valid value accept, invalid value reject, and missing-key reject.
- **AC-HRN-002-07** has 3 depth-1 children covering leak rejection, clean prompt accept, and 3-iteration independence. AC-HRN-002-07.a has 2 depth-2 grandchildren (`.a.i`, `.a.ii`) covering specific leak substrings — this exercises the MaxDepth=3 boundary.

The remaining ACs are flat — demonstrating that flat and hierarchical co-exist within the same SPEC (REQ-SPC-001-040 → AC-SPC-001-09 from the SPC-001 plan-phase precedent).

Recursive proof: HRN-002's acceptance.md uses the hierarchical AC schema that SPC-001 introduced. If the SPC-001 parser (`internal/spec/parser.go`) successfully reads this file with `go test ./internal/spec/...` against the live fixture, then the SPC-001 schema works for HRN-002's own paperwork. This is the same recursive proof SPC-001's acceptance.md provides for itself.

End of acceptance.
