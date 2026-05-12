# Tasks — SPEC-V3R2-HRN-002 Evaluator Memory Scope Amendment

> Run-phase task breakdown derived from plan.md milestones. Naming convention: `T-HRN002-NN`. Owner roles use the agent catalog (manager-spec / expert-backend / manager-docs / manager-quality).

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec | Initial T-HRN002-01 through T-HRN002-12 task list. |

---

## 1. Task Index

| Task ID | Milestone | Owner | Priority | Depends on |
|---------|-----------|-------|----------|------------|
| T-HRN002-01 | M1 | manager-spec | Critical | — |
| T-HRN002-02 | M1 | manager-spec | Critical | T-HRN002-01 |
| T-HRN002-03 | M2 | manager-spec | Critical | T-HRN002-02 |
| T-HRN002-04 | M2 | expert-backend | Critical | T-HRN002-02 |
| T-HRN002-05 | M3 | manager-docs | High | T-HRN002-03 |
| T-HRN002-06 | M3 | manager-docs | High | T-HRN002-03 |
| T-HRN002-07 | M3 | expert-backend | High | T-HRN002-04, T-HRN002-06 |
| T-HRN002-08 | M4 | expert-backend | High | T-HRN002-04 |
| T-HRN002-09 | M4 | expert-backend | High | T-HRN002-08 |
| T-HRN002-10 | M5 | manager-quality | Critical | T-HRN002-03, T-HRN002-07, T-HRN002-08 |
| T-HRN002-11 | M5 | manager-quality | Critical | T-HRN002-10 |
| T-HRN002-12 | M5 | manager-spec | Critical | T-HRN002-11 |

---

## 2. Task Detail

### M1 — Schema design + research delivery

#### T-HRN002-01 — Author research.md with codebase audit

Owner: manager-spec
Priority: Critical
Description: Produce research.md that captures the as-is state of evaluator-memory across constitution clauses, evaluator-active agent body, moai-workflow-gan-loop SKILL.md, harness Go config subsystem, and Sprint Contract storage. Compare against spec.md §5 (19 REQs) and produce a gap analysis. Surface Decision D1-D10 driving plan.md.
Inputs: spec.md, `.claude/rules/moai/design/constitution.md`, `.claude/agents/moai/evaluator-active.md`, `.claude/skills/moai-workflow-gan-loop/SKILL.md`, `internal/harness/*.go`, `.moai/config/sections/{design,harness}.yaml`, dependency SPECs (CON-001, CON-002, HRN-001).
Outputs: `.moai/specs/SPEC-V3R2-HRN-002/research.md` with ≥25 file:line anchors and §11 decision block.
Acceptance: research.md committed; ≥25 anchors verified by plan-auditor (50 confirmed).
Status: DONE (this PR).

#### T-HRN002-02 — Author plan.md, acceptance.md, tasks.md, progress.md, spec-compact.md

Owner: manager-spec
Priority: Critical
Description: Produce remaining 5 plan-phase artifacts. acceptance.md self-demonstrates hierarchical schema on ≥3 ACs. progress.md frontmatter `plan_status: audit-ready`.
Inputs: research.md (T-HRN002-01).
Outputs: 5 markdown files in `.moai/specs/SPEC-V3R2-HRN-002/`.
Acceptance: All 19 REQs mapped; tasks.md uses T-HRN002-NN naming; plan-auditor PASS at iteration ≤2.
Status: DONE (this PR).

### M2 — Constitution amendment + HarnessConfig schema

#### T-HRN002-03 — Insert §11.4.1 amendment text + version bump

Owner: manager-spec
Priority: Critical
Description: Insert the verbatim §11.4.1 text from spec.md §1.2 into `.claude/rules/moai/design/constitution.md` immediately after line 335 (end of §11.4 Sprint Contract Protocol rules). Bump version metadata at line 405 from 3.4.0 to 3.5.0. Append a HISTORY row (lines 1-9) recording the amendment.
Caveat: this edit triggers FrozenGuard (Decision D7); MUST go through CON-002 amendment path (T-HRN002-12 paperwork must run in parallel).
Inputs: spec.md §1.2 verbatim text (lines 73-83); CON-002 amendment infrastructure live on main.
Files: `.claude/rules/moai/design/constitution.md`.
Outputs: §11.4.1 section present; version 3.5.0; HISTORY row appended.
Acceptance: AC-HRN-002-01.a/b/c PASS (byte-identical text + insertion site verified); AC-HRN-002-02 PASS (version + HISTORY). Add `@MX:WARN reason="FROZEN-zone amendment per CON-002 §5 Layer 1"` to the inserted block.

#### T-HRN002-04 — Extend HarnessConfig with EvaluatorConfig.MemoryScope

Owner: expert-backend
Priority: Critical
Description: Add `Evaluator EvaluatorConfig` field to the `HarnessConfig` struct in `internal/config/types.go` (sub-struct of HRN-001 schema). Declare new `EvaluatorConfig` struct with single field `MemoryScope string` tagged `yaml:"memory_scope" validate:"required,eq=per_iteration"`. Wire `HRN_EVAL_MEMORY_FROZEN` error sentinel in `internal/config/errors.go`. Update `LoadHarnessConfig()` and `LoadDesignConfig()` (if separate) to invoke validator.
Inputs: HRN-001 `internal/config/types.go` (HarnessConfig struct), HRN-001 `internal/config/loader.go` (loader entry points).
Files: `internal/config/types.go`, `internal/config/loader.go`, `internal/config/errors.go`, `internal/config/loader_test.go` (fixture setup only — full fixtures land in T-HRN002-08).
Outputs: HarnessConfig.Evaluator.MemoryScope field; validator wired; sentinel declared.
Acceptance: `go test ./internal/config/...` green (sentinel exists, struct compiles); AC-HRN-002-04 schema-level verification deferred to T-HRN002-08 fixture landings. Add `@MX:NOTE` on EvaluatorConfig struct: "FROZEN at per_iteration per design-constitution §11.4.1 (SPEC-V3R2-HRN-002)".

### M3 — Agent body + SKILL.md + leak detection

#### T-HRN002-05 — Add evaluator-active cross-reference line

Owner: manager-docs
Priority: High
Description: Insert exactly one line into `.claude/agents/moai/evaluator-active.md` immediately above `## Sprint Contract Negotiation` heading (around line 91). Text: "Per design-constitution §11.4.1, evaluator judgment memory is ephemeral per iteration. The orchestrator MUST respawn evaluator-active via a fresh `Agent()` call at each GAN-loop iteration boundary; prior iteration's evaluator transcript MUST NOT appear in the new spawn prompt."
Inputs: §11.4.1 text landed via T-HRN002-03.
Files: `.claude/agents/moai/evaluator-active.md`.
Outputs: single new cross-reference line; no structural rewrite.
Acceptance: AC-HRN-002-05 PASS (line present + references §11.4.1). Add `@MX:NOTE` on the new line: "Cross-references design-constitution §11.4.1 (SPC-V3R2-HRN-002)".

#### T-HRN002-06 — Insert Phase 4b iteration-handoff step + Solo Mode subsection in SKILL.md

Owner: manager-docs
Priority: High
Description: Insert new step "Phase 4b: Iteration Handoff (REQ-HRN-002-009)" between Phase 4 (Loop Decision, around line 131) and Phase 5 (Iteration Feedback, around line 133) in `.claude/skills/moai-workflow-gan-loop/SKILL.md`. Body declares fresh respawn with narrow 3-input prompt. Cite design-constitution §11.4.1. Annotate stagnation-detection block (lines 147-158) to clarify score deltas read from Sprint Contract artifact (NOT prior evaluator memory). Append a new "Solo Mode" subsection documenting `{spec-id}/contract.yaml` alias per REQ-HRN-002-016.
Inputs: §11.4.1 text landed via T-HRN002-03.
Files: `.claude/skills/moai-workflow-gan-loop/SKILL.md`.
Outputs: Phase 4b step (~15 lines); stagnation annotation; Solo Mode subsection (~10 lines).
Acceptance: AC-HRN-002-06 PASS (Phase 4b cites §11.4.1); AC-HRN-002-13 PASS (Solo Mode subsection documents alias). Add `@MX:NOTE` on Phase 4b step: "Per design-constitution §11.4.1; enforces REQ-HRN-002-009 fresh respawn".

#### T-HRN002-07 — Author prior-judgment leak detection test

Owner: expert-backend
Priority: High
Description: Create new test file `internal/harness/evaluator_leak_test.go` implementing `TestEvaluatorPriorJudgmentLeak`. Test scans synthetic spawn-prompt strings for forbidden substrings (`Score:`, `Feedback:`, `Verdict:`, `Iteration N` regex). Validator function returns `HRN_EVAL_PRIOR_JUDGMENT_LEAK` error on detection. Cover positive cases (clean prompts pass), negative cases (forbidden substrings rejected), and depth-2 cases per AC-HRN-002-07.a.i + .a.ii (numbered iteration reference + paraphrased rationale).
Inputs: HarnessConfig struct from T-HRN002-04; AC-HRN-002-07 leaf scenarios from acceptance.md.
Files: `internal/harness/evaluator_leak_test.go` (new), `internal/harness/evaluator_leak.go` (new — validator function, ~50 LOC).
Outputs: test file + validator function.
Acceptance: AC-HRN-002-07.a, .a.i, .a.ii, .b, .c all PASS via `go test -run TestEvaluatorPriorJudgmentLeak ./internal/harness/...`. Add `@MX:WARN reason="prior-judgment leak detection per REQ-HRN-002-017"` on the test function.

### M4 — Config propagation + loader fixtures

#### T-HRN002-08 — Add `evaluator.memory_scope: per_iteration` to design.yaml + harness.yaml (template-first)

Owner: expert-backend
Priority: High
Description: Add the `evaluator:\n  memory_scope: per_iteration` key block to `internal/template/templates/.moai/config/sections/design.yaml` (under top-level `design:` namespace, sibling to `gan_loop:`) and `internal/template/templates/.moai/config/sections/harness.yaml` (under top-level `harness:` namespace, sibling to `default_profile:`). Run `make build` to regenerate embedded files. Sync to local `.moai/config/sections/{design,harness}.yaml` to verify byte-identical.
Inputs: §11.4.1 text declares `per_iteration` is the only accepted value.
Files: `internal/template/templates/.moai/config/sections/design.yaml`, `internal/template/templates/.moai/config/sections/harness.yaml`, `.moai/config/sections/design.yaml`, `.moai/config/sections/harness.yaml`, generated `internal/template/embedded.go`.
Outputs: 4 YAML files updated (template + local mirror for both); embedded.go regenerated.
Acceptance: AC-HRN-002-03 PASS (both files contain the key; template mirrors byte-identical). Add `@MX:NOTE` on both YAML insertions: "FROZEN at per_iteration per design-constitution §11.4.1; loader enforces via HRN_EVAL_MEMORY_FROZEN".

#### T-HRN002-09 — Add loader test fixtures covering AC-HRN-002-04

Owner: expert-backend
Priority: High
Description: Add 3 fixtures under `internal/config/testdata/`:
- `eval-memory-valid/harness.yaml` — `evaluator.memory_scope: per_iteration` (passes validation).
- `eval-memory-frozen-violation/harness.yaml` — `evaluator.memory_scope: cumulative` (fails with `HRN_EVAL_MEMORY_FROZEN`).
- `eval-memory-missing/harness.yaml` — `evaluator:` block present but `memory_scope` key absent (fails with required-field sentinel).
Extend `internal/config/loader_test.go` with 3 corresponding test cases asserting each fixture's expected verdict.
Inputs: validator wired by T-HRN002-04.
Files: 3 new fixture directories under `internal/config/testdata/`; loader_test.go additions.
Outputs: 3 fixtures + 3 test cases.
Acceptance: AC-HRN-002-04.a/b/c PASS via `go test ./internal/config/...`. No regression in existing HRN-001 loader tests.

### M5 — CON-002 paperwork + completion gate

#### T-HRN002-10 — Run Canary against last 3 GAN-loop completions (or fallback)

Owner: manager-quality
Priority: Critical
Description: Execute the CON-002 Layer 2 Canary step. Select the last 3 completed GAN-loop projects from `.moai/design/` (or v2-legacy design completions if v3 corpus is empty). Simulate the new respawn protocol in shadow mode via fixtures + the leak detection validator (T-HRN002-07). Verify no project's quality score drops >0.10 vs the baseline (cumulative-memory) scoring. If <3 subjects available, emit `CanaryUnavailable` verdict per REQ-CON-002-020. Capture script output to `.moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt`.
Inputs: leak validator (T-HRN002-07), CON-002 protocol live on main.
Files: `.moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt` (new).
Outputs: Canary log committed; verdict recorded (PASS / CanaryUnavailable).
Acceptance: AC-HRN-002-09 Canary section populated. Mirror SPC-001 `canary-v2-reparse.txt` structure.

#### T-HRN002-11 — Add `@MX:ANCHOR` + `@MX:NOTE` + `@MX:WARN` tags

Owner: manager-quality
Priority: Critical
Description: Annotate canonical sources with MX tags per plan.md §3 M5 mx_plan.
Files:
- `.moai/specs/SPEC-V3R2-HRN-002/acceptance.md` — `@MX:ANCHOR fan_in=3` on AC-HRN-002-01 (self-demonstrating example block, cross-referenced by HRN-003 + Sprint Contract + CON-002 catalog). Already present in this plan PR; verify in run phase.
- `internal/config/types.go` — `@MX:NOTE` on EvaluatorConfig struct: "FROZEN at per_iteration per design-constitution §11.4.1 (SPEC-V3R2-HRN-002)" (added by T-HRN002-04; verify retained).
- `.claude/rules/moai/design/constitution.md` (insertion point from T-HRN002-03) — `@MX:WARN reason="FROZEN-zone amendment per CON-002 §5 Layer 1 (Frozen Guard); modifications require full Canary + HumanOversight cycle"` (added by T-HRN002-03; verify retained).
Outputs: 3 MX-tag annotations verified.
Acceptance: `moai mx --validate` passes; no orphan tags; all `@MX:REASON` sub-lines present for WARN and ANCHOR.

#### T-HRN002-12 — File CON-002 amendment evidence + HumanOversight approval + dual evolution-log + zone-registry CONST entry

Owner: manager-spec
Priority: Critical
Description: Compile FrozenGuard / Canary / ContradictionDetector / RateLimiter / HumanOversight evidence per spec.md §2.1 + plan.md M5 task list. Mirror SPC-001 file structure (`.moai/specs/SPEC-V3R2-SPC-001/con-002-amendment-evidence.md` as the template). Record maintainer approval (Goos Kim) timestamp + reviewer in landing PR description. Write EVO-HRN-002 entries to BOTH `.moai/research/evolution-log.md` (CON-002 core log) AND `.moai/design/v3-research/evolution-log.md` (design subsystem cross-reference per Decision D5). Add new `CONST-V3R2-NNN` entry to `.claude/rules/moai/core/zone-registry.md` with `clause: "§11.4.1 Evaluator Memory Scope (Principle 4)"`, `zone: Frozen`, `canary_gate: true` per Decision D7. Update progress.md `plan_status: complete` after run-phase merges.
Inputs: Canary log from T-HRN002-10; MX tags verified by T-HRN002-11.
Files: `.moai/specs/SPEC-V3R2-HRN-002/con-002-amendment-evidence.md` (new), `.moai/research/evolution-log.md` (new or appended), `.moai/design/v3-research/evolution-log.md` (new or appended), `.claude/rules/moai/core/zone-registry.md` (CONST entry appended), `.moai/specs/SPEC-V3R2-HRN-002/progress.md` (status update).
Outputs: 5-section evidence document; dual evolution-log entries; zone-registry CONST entry; progress.md status `complete`.
Acceptance: AC-HRN-002-09 PASS (all 5 layers documented); AC-HRN-002-10 PASS (FrozenGuard rejection test passes post-CONST registration); AC-HRN-002-12 PASS (both evolution-log files have EVO-HRN-002 entry with `core_log_ref` pointer); HumanOversight approval received.

---

## 3. Task Dependency Graph

```
T-HRN002-01 (research.md)
    └─ T-HRN002-02 (plan/AC/tasks/progress/compact)
            ├─ T-HRN002-03 (§11.4.1 amendment text)
            │      ├─ T-HRN002-05 (evaluator-active cross-ref)
            │      ├─ T-HRN002-06 (SKILL.md Phase 4b + Solo Mode)
            │      └─ T-HRN002-10 (Canary)
            │             └─ T-HRN002-11 (MX tags)
            │                    └─ T-HRN002-12 (CON-002 evidence + dual log + CONST entry)
            └─ T-HRN002-04 (HarnessConfig + EvaluatorConfig struct)
                   ├─ T-HRN002-07 (leak detection test)
                   │      └─ T-HRN002-10 (Canary uses validator)
                   └─ T-HRN002-08 (yaml keys + template-first)
                          └─ T-HRN002-09 (loader test fixtures)
                                 └─ T-HRN002-10 (Canary)
```

---

## 4. Owner Role Reference

| Owner | Responsibilities |
|-------|------------------|
| manager-spec | Plan-phase artifact authoring (T-01, T-02), constitution amendment (T-03), CON-002 evidence + dual log + CONST entry (T-12). |
| expert-backend | Go-level implementation: HarnessConfig extension (T-04), leak detection test (T-07), yaml config + template (T-08), loader test fixtures (T-09). |
| manager-docs | Documentation amendments: evaluator-active cross-ref (T-05), SKILL.md Phase 4b + Solo Mode (T-06). |
| manager-quality | Run-phase quality gate: Canary (T-10), MX tag verification (T-11). |

---

## 5. Run-Phase Entry Criteria

This tasks.md becomes actionable when:
- [ ] Plan PR (this) merges to main.
- [ ] plan-auditor PASS recorded.
- [ ] manager-tdd or expert-backend assigned to lead M2-M5 execution in a fresh worktree (`moai worktree new SPEC-V3R2-HRN-002 --base origin/main`).

End of tasks.
