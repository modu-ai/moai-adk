---
id: SPEC-V3R2-HRN-002
title: "Evaluator Memory Scope Amendment (per-iteration fresh judgment)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 5 — Harness + Evaluator"
module: ".claude/rules/moai/design/constitution.md, .moai/config/sections/design.yaml, .moai/config/sections/harness.yaml, .claude/agents/moai/evaluator-active.md, .claude/skills/moai/moai-workflow-gan-loop/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-CON-002
  - SPEC-V3R2-HRN-001
related_problem:
  - P-Z01
related_theme: "Layer 5 — Harness, Master §4.5, §5.7 Agent-as-Judge Without Memory, §8 BC-V3R2-010"
breaking: true
bc_id: [BC-V3R2-010]
lifecycle: spec-anchored
tags: "evaluator, agent-as-judge, sprint-contract, memory-scope, fresh-judgment, constitution-amendment, frozen-zone, v3r2"
---

# SPEC-V3R2-HRN-002: Evaluator Memory Scope Amendment (per-iteration fresh judgment)

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

The **single most important amendment to MoAI's design subsystem**, per Master §5.7: the current `.claude/rules/moai/design/constitution.md` §11.4 Sprint Contract Protocol retains evaluator memory across GAN-loop iterations. Agent-as-a-Judge (R1 §9 — Zhuge et al. 2024, arXiv:2410.10934) explicitly flags this as an anti-pattern:

> "Any errors in previous judgments could lead to a chain of errors... We hypothesize that the absence of a memory mechanism was beneficial to achieve more accurate and fair evaluations."

R1 reports that Agent-as-a-Judge matches human reliability (90%) **only when evaluator judgment memory is ephemeral per iteration**. Moai's current implementation cascades prior reasoning, compounding evaluator errors.

v3r2 amends the design constitution to decouple two kinds of state:

1. **Sprint Contract state is DURABLE**: passed criteria carry forward, failed criteria get refined, new criteria added when gaps appear. This is external, file-backed, cross-iteration state stored at `.moai/sprints/{team-id}/contract.yaml`.
2. **Evaluator judgment memory is EPHEMERAL**: each iteration spawns evaluator-active with a fresh Claude context that sees only (a) the BRIEF/SPEC, (b) the Sprint Contract criterion states, (c) the artifact under review. It MUST NOT see prior judgment rationales.

This SPEC is a **FROZEN-zone amendment**: design-constitution.md is FROZEN per its own §2 (and SPEC-V3R2-CON-001 generalizes this). The amendment MUST pass through SPEC-V3R2-CON-002's 5-layer safety gate (FrozenGuard + Canary + ContradictionDetector + RateLimiter + HumanOversight). This SPEC explicitly cites CON-002 as a blocking dependency and presents the before/after clause text for review.

### 1.1 Background

Master §5.7 identifies this as the single most important amendment. Master §Appendix A maps Principle 4 "Evaluator Judgments Fresh, Contract State Durable" to this SPEC. Master §8 declares `BC-V3R2-010`: "Evaluator memory scope per-iteration (design-constitution §11.4 amendment)" as a top-5 breaking change.

Current state:
- `.claude/rules/moai/design/constitution.md` §11 GAN Loop Contract is FROZEN per §2.
- §11.4 Sprint Contract Protocol declares iteration semantics but is silent on memory scope.
- `evaluator-active` agent body does not reset context between iterations; spawned via `Agent()` it inherits session context.
- Result: evaluator reasoning cascades; 2nd iteration judgment influenced by 1st; 3rd by 2nd. Fairness degrades.

v3r2 state:
- Amend §11.4 to split "Sprint Contract State" (durable, file-backed) from "Evaluator Judgment Memory" (ephemeral, per-iteration).
- Add `evaluator.memory_scope: per_iteration` to both `.moai/config/sections/design.yaml` (design-pipeline-specific) and `.moai/config/sections/harness.yaml` (general GAN loop).
- Wire SPEC-V3R2-HRN-001 loader to read the flag; wire the GAN loop runner to respawn evaluator-active per iteration with a narrow prompt carrying only (BRIEF, criterion states, artifact path) — NO transcripts of prior iterations' judgments.

### 1.2 Constitution amendment (before → after)

**Before (current §11.4 Sprint Contract Protocol, verbatim extract):**

> "Sprint Contract artifacts are stored in `.moai/sprints/` (from design.yaml `sprint_contract.artifact_dir`)"

(§11.4 does not explicitly address evaluator memory; the ambiguity is the problem.)

**After (v3r2 amendment text, to be inserted as new §11.4.1 Evaluator Memory Scope):**

> "**§11.4.1 Evaluator Memory Scope (Principle 4)**. Evaluator judgment memory SHALL be ephemeral per iteration. Each GAN-loop iteration invokes evaluator-active with a fresh LM context seeing only: (a) the SPEC + BRIEF, (b) the current Sprint Contract criterion states, (c) the artifact path under review. Prior iteration judgment rationales, scoring internals, or reflection traces MUST NOT appear in the evaluator's context window.
>
> Sprint Contract state SHALL be durable across iterations. Passed criteria carry forward (no regression allowed). Failed criteria carry forward with refined expectations. New criteria MAY be added by evaluator-active when the prior iteration revealed gaps.
>
> Implementation: the GAN loop runner SHALL respawn evaluator-active for each iteration (no session persistence). The Sprint Contract file at `.moai/sprints/{team-id}/contract.yaml` (or `.moai/sprints/{spec-id}/contract.yaml` for solo mode) is the sole cross-iteration carrier of evaluation state.
>
> Configuration: `evaluator.memory_scope: per_iteration` in both `design.yaml` and `harness.yaml` is FROZEN at `per_iteration`; any attempt to set another value (e.g., `cumulative`) fails loader validation with `HRN_EVAL_MEMORY_FROZEN`."

This insertion is classified FROZEN under SPEC-V3R2-CON-001 zone model; future changes require CON-002 graduation protocol.

*Source: R1 §9 Agent-as-a-Judge (Zhuge et al. 2024); design-principles.md Principle 4; problem-catalog.md P-Z01; major-v3-master.md §5.7, §8 BC-V3R2-010, §Appendix A.*

### 1.3 Non-Goals

- Changing the Sprint Contract file format (SPEC-V3R2-HRN-003 and existing constitution §11.4 declare the YAML shape).
- Implementing fresh-memory evaluator without the constitution amendment (illegal — FROZEN zone).
- Modifying the evaluator-active agent body beyond a single cross-reference to the new §11.4.1.
- Implementing hierarchical acceptance criteria scoring (SPEC-V3R2-HRN-003).
- Changing `pass_threshold` floor (stays 0.60 FROZEN).
- Implementing additional evaluator profiles (content out of scope).
- Per-iteration telemetry collection (deferred to Master §12 Open Question #3).
- CG-mode evaluator spawning (same rules apply; no mode-specific behavior here).

---

## 2. Scope (범위)

### 2.1 In Scope

- Run the CON-002 amendment graduation protocol for design-constitution §11.4.1 insertion:
  - Step 1: Proposal Generation — this SPEC document is the proposal.
  - Step 2: Canary Validation — shadow-eval against last 3 completed design projects (if any exist under `.moai/design/`).
  - Step 3: Contradiction Check — scan existing rules for conflicts with the new §11.4.1.
  - Step 4: Human Review — AskUserQuestion gate surfaced by the MoAI orchestrator.
  - Step 5: Application — on approval, write the new §11.4.1 text and update design-constitution version to v3.4.0.
- Insert the new `§11.4.1 Evaluator Memory Scope` text verbatim as declared in §1.2.
- Add `evaluator.memory_scope: per_iteration` as a FROZEN-value key in `.moai/config/sections/design.yaml` and `.moai/config/sections/harness.yaml`.
- Update `HarnessConfig` struct (from SPEC-V3R2-HRN-001) to include `Evaluator.MemoryScope string` with validator enforcing `per_iteration` as the only accepted value in v3.0.
- Update the GAN loop runner (`.claude/skills/moai/moai-workflow-gan-loop/SKILL.md` + associated Go-side runner if applicable) to respawn evaluator-active per iteration with a narrow prompt.
- Update evaluator-active agent body with a single cross-reference line pointing at §11.4.1.
- Add regression test: running 3 iterations of GAN loop with intentionally-divergent artifacts produces 3 independent score cards (no visible cascade).
- Template-first: all file edits land in `internal/template/templates/` first; `make build` regenerates; local tree byte-identical.
- Update `.moai/design/v3-research/evolution-log.md` (or create) with the amendment record per constitution §6.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Changing the Sprint Contract YAML schema (durability mechanics already encoded in existing §11.4).
- Implementing per-SPEC evaluator profile overrides (harness-level only).
- Implementing hierarchical scoring (SPEC-V3R2-HRN-003).
- Adding evaluator telemetry or metrics dashboards.
- Changing `max_iterations` (stays at 5 per design-constitution §11).
- Changing `escalation_after` (stays at 3).
- Changing the actor (expert-frontend) body.
- Implementing actor memory scope (actor memory is project-level per Master §4.4 agent frontmatter).
- Cross-SPEC evaluator memory sharing (Non-Goal; evaluator is always SPEC-scoped).
- Allowing `cumulative` memory mode as a future opt-in (the constitution floor is `per_iteration` only; opt-out would require a new amendment cycle).
- Changing the evaluator model (stays at opus per current frontmatter).

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-beta.1+ (Phase 5)
- Claude Code v2.1.111+ (Opus 4.7 Adaptive Thinking required for fresh-context evaluator per HRN-003 rubric anchoring)
- Target constitution file: `.claude/rules/moai/design/constitution.md` (FROZEN zone, amended here)
- Target config files: `.moai/config/sections/design.yaml`, `.moai/config/sections/harness.yaml`
- Target evaluator agent: `.claude/agents/moai/evaluator-active.md`
- Target skill: `.claude/skills/moai/moai-workflow-gan-loop/SKILL.md`
- Target Go module: `internal/harness/` + `internal/config/types.go`
- Version bump: design-constitution v3.3.0 → v3.4.0 (amendment increments minor)
- Migration: BC-V3R2-010 (AUTO per Master §8 — config flag addition; evaluator-active respawn per iteration; old evaluator sessions retired on first upgrade)

---

## 4. Assumptions (가정)

- SPEC-V3R2-CON-001 has landed; the FROZEN/EVOLVABLE zone model is live.
- SPEC-V3R2-CON-002 has landed; the 5-layer amendment graduation protocol (FrozenGuard + Canary + ContradictionDetector + RateLimiter + HumanOversight) is implemented and can be invoked.
- SPEC-V3R2-HRN-001 has landed; the `HarnessConfig` struct is live and can be extended with `Evaluator.MemoryScope`.
- The amendment can be safely rolled out because the Sprint Contract file already persists cross-iteration state (mechanism already in place; v3r2 just enforces evaluator memory absence).
- Existing design projects under `.moai/design/` provide ≥3 canary subjects; if fewer, the Canary step reports `insufficient corpus` and waits.
- Human approval will explicitly cite R1 §9 evidence and Principle 4 in the AskUserQuestion decision record.
- No third-party tool or workflow relies on evaluator cross-iteration memory (Master §6 Problems intentionally deferred and §10 Risk Register R4 both flag the amendment as the primary change in beta.1).
- Respawning evaluator-active per iteration is supported by Claude Code Agent() primitives; no new runtime feature required.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-HRN-002-001 (Ubiquitous) — 헌법 수정 적용**
The file `.claude/rules/moai/design/constitution.md` **shall** contain a new section `§11.4.1 Evaluator Memory Scope (Principle 4)` with the verbatim text declared in §1.2 of this SPEC.

**REQ-HRN-002-002 (Ubiquitous) — 버전 증가**
The design-constitution header metadata **shall** be updated from `Version: 3.3.0` to `Version: 3.4.0` with the amendment recorded in the HISTORY section.

**REQ-HRN-002-003 (Ubiquitous) — design.yaml 키 추가**
The file `.moai/config/sections/design.yaml` **shall** contain the key `evaluator.memory_scope: per_iteration`; the value is FROZEN.

**REQ-HRN-002-004 (Ubiquitous) — harness.yaml 키 추가**
The file `.moai/config/sections/harness.yaml` **shall** contain the key `evaluator.memory_scope: per_iteration` (mirror of design.yaml for harness-routed runs); the value is FROZEN.

**REQ-HRN-002-005 (Ubiquitous) — 구조체 확장**
The `HarnessConfig` struct in `internal/config/types.go` **shall** include `Evaluator.MemoryScope string` with validator tag enforcing the value `per_iteration` as the only accepted string in v3.0.

**REQ-HRN-002-006 (Ubiquitous) — 에이전트 바디 참조**
The file `.claude/agents/moai/evaluator-active.md` body **shall** contain a single cross-reference line pointing at design-constitution §11.4.1 stating that evaluator memory is ephemeral per iteration.

**REQ-HRN-002-007 (Ubiquitous) — GAN loop 스킬 갱신**
The skill `.claude/skills/moai/moai-workflow-gan-loop/SKILL.md` **shall** declare that each iteration spawns evaluator-active with a fresh context; the skill body **shall** cite design-constitution §11.4.1.

### 5.2 Event-Driven (이벤트 기반)

**REQ-HRN-002-008 (Event-Driven) — 수정안 승인 흐름**
**When** this SPEC is submitted for merge, the SPEC-V3R2-CON-002 amendment graduation protocol **shall** execute all 5 layers in order (FrozenGuard, Canary with last 3 design projects if available, ContradictionDetector, RateLimiter respecting 3/week cap, HumanOversight via AskUserQuestion); approval is required before the §11.4.1 text is committed.

**REQ-HRN-002-009 (Event-Driven) — 반복 시작 시 리스폰**
**When** the GAN loop begins a new iteration, the runner **shall** invoke a new `Agent()` spawn of `evaluator-active` with a spawn-prompt containing exactly three inputs: the SPEC/BRIEF reference, the current Sprint Contract criterion states, and the artifact path; the prompt **shall** NOT include any prior iteration's evaluator transcript.

**REQ-HRN-002-010 (Event-Driven) — 쓰루풋 수정 로그**
**When** the amendment is applied, the runner **shall** append a record to `.moai/design/v3-research/evolution-log.md` (or create the file if absent) with fields `{id: EVO-HRN-002, timestamp, before_snippet, after_snippet, canary_verdict, approver, rationale_cite: "R1 §9 + Principle 4"}`.

**REQ-HRN-002-011 (Event-Driven) — 유효성 위반 처리**
**When** `LoadHarnessConfig()` or `LoadDesignConfig()` encounters `evaluator.memory_scope` value other than `per_iteration`, the loader **shall** return error `HRN_EVAL_MEMORY_FROZEN` refusing to start the runtime.

### 5.3 State-Driven (상태 기반)

**REQ-HRN-002-012 (State-Driven) — 컨트랙트 상태 내구성**
**While** the GAN loop is running across iterations, the Sprint Contract file **shall** carry all criterion states forward (passed, failed, refined, new); no criterion may regress from passed to failed across iterations except through explicit human override.

**REQ-HRN-002-013 (State-Driven) — 판결 메모리 휘발성**
**While** the GAN loop is running, the evaluator-active agent context window **shall** contain no prior iteration's judgment rationale; the fresh respawn is the enforcement mechanism.

**REQ-HRN-002-014 (State-Driven) — FROZEN 값 보존**
**While** the v3.0.0 minor cycle is active, `evaluator.memory_scope` **shall** remain fixed at `per_iteration`; any attempt to change the value requires a new CON-002 amendment cycle.

### 5.4 Optional (선택)

**REQ-HRN-002-015 (Optional) — 수정 전 카나리 결과 공개**
**Where** the Canary step runs successfully with ≥3 design project subjects, the Canary verdict (score delta summary) **may** be attached to the AskUserQuestion decision record as supporting evidence for the HumanOversight approval.

**REQ-HRN-002-016 (Optional) — Sprint Contract 위치 별칭**
**Where** the GAN loop runs in solo mode (no team ID), the Sprint Contract file location **may** be `.moai/sprints/{spec-id}/contract.yaml` instead of `.moai/sprints/{team-id}/contract.yaml`; both locations are equivalent for loader purposes.

### 5.5 Unwanted Behavior

**REQ-HRN-002-017 (Unwanted Behavior) — 이전 반복 전사 주입 금지**
**If** the GAN loop runner constructs an evaluator spawn-prompt containing any substring that includes a prior iteration's `Score:` or `Feedback:` line from the evaluator's previous output, **then** CI (via a new integration test in HRN-001 test suite) **shall** fail with error `HRN_EVAL_PRIOR_JUDGMENT_LEAK`.

**REQ-HRN-002-018 (Unwanted Behavior) — cumulative 값 재도입 금지**
**If** any config file or schema introduces `evaluator.memory_scope: cumulative` (or any value other than `per_iteration`), **then** the loader validator **shall** fail with error `HRN_EVAL_MEMORY_FROZEN` (REQ-011) and the commit **shall** be rejected at CI.

**REQ-HRN-002-019 (Unwanted Behavior) — 수정 문구 수정 금지**
**If** a PR modifies the inserted §11.4.1 text without running the CON-002 graduation protocol, **then** the FrozenGuard (from SPEC-V3R2-CON-001) **shall** block the write and emit error `CON_FROZEN_GUARD_REJECTED`.

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-HRN-002-01**: `.claude/rules/moai/design/constitution.md` contains §11.4.1 with the exact text from §1.2 of this SPEC.
- **AC-HRN-002-02**: The constitution header shows `Version: 3.4.0` and HISTORY has a row for this amendment.
- **AC-HRN-002-03**: Both `design.yaml` and `harness.yaml` contain `evaluator.memory_scope: per_iteration`.
- **AC-HRN-002-04**: `HarnessConfig.Evaluator.MemoryScope` field parses and validates; setting value `"cumulative"` fails with `HRN_EVAL_MEMORY_FROZEN` (REQ-011, AC link).
- **AC-HRN-002-05**: evaluator-active agent body references §11.4.1.
- **AC-HRN-002-06**: moai-workflow-gan-loop SKILL.md cites §11.4.1 and declares fresh-respawn semantics.
- **AC-HRN-002-07**: Running 3 consecutive GAN iterations with deliberately divergent artifacts produces 3 independent ScoreCards; no ScoreCard shows substring content from another iteration's Feedback (leak test).
- **AC-HRN-002-08**: Sprint Contract YAML persists across 3 iterations; passed criteria carry forward; test fixture covers pass/fail/refined/new states.
- **AC-HRN-002-09**: CON-002 graduation protocol executes successfully; `.moai/design/v3-research/evolution-log.md` contains an EVO-HRN-002 record with canary verdict and approver.
- **AC-HRN-002-10**: Attempting to bypass FrozenGuard by editing §11.4.1 via plain file write fails with `CON_FROZEN_GUARD_REJECTED`.
- **AC-HRN-002-11**: Backward-compat test: v2 evaluator sessions (memory-retentive) are detected and retired on first v3.0.0-beta.1 upgrade with log `EVAL_SESSION_UPGRADED`.

---

## 7. Constraints (제약)

- [HARD] FROZEN constitution amendment — changes to §11.4.1 after this SPEC land MUST pass through CON-002 graduation.
- [HARD] FROZEN value for `evaluator.memory_scope` — only `per_iteration` accepted (REQ-014, REQ-018).
- [HARD] Sprint Contract state durability preserved (REQ-012).
- [HARD] No prior judgment leak in evaluator prompts (REQ-017).
- [HARD] pass_threshold floor 0.60 preserved (cross-reference SPEC-V3R2-HRN-001 REQ-012).
- [HARD] Template-First (CLAUDE.local.md §2) — all file edits land in template tree first.
- [HARD] Human approval required for this amendment (CON-002 5th layer).
- [HARD] evolution-log.md record is FROZEN once written (per constitution §6).
- [HARD] Breaking change declared (BC-V3R2-010 per Master §8); AUTO migration rolls forward; no manual step required.
- [HARD] 16-language neutrality — amendment is language-agnostic.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                               | Impact | Mitigation                                                                                                 |
|--------------------------------------------------------------------|--------|------------------------------------------------------------------------------------------------------------|
| Canary step finds <3 design projects (insufficient corpus)         | MEDIUM | Document the shortage in the amendment record; proceed with explicit note; revisit after first post-amendment design runs |
| Fresh respawn introduces overhead (model init per iteration)       | MEDIUM | Opus 4.7 spawn is fast (<2s); overhead < 5% of typical iteration duration; acceptable for correctness gain |
| Human reviewer approves without reading R1 §9 evidence             | HIGH   | AskUserQuestion bundles R1 §9 citation + canary verdict + design-principles.md Principle 4 as mandatory context |
| v2 users rely on cumulative evaluator insight (undocumented)       | MEDIUM | BC-V3R2-010 listed in release notes; CHANGELOG entry explains the quality tradeoff; Sprint Contract still carries criteria state |
| Sprint Contract state grows unbounded                              | LOW    | Criteria list is bounded by SPEC acceptance count (typically <20); archive on SPEC completion              |
| evaluator-active prompt reconstruction subtly leaks prior context  | HIGH   | REQ-017 integration test scans every spawn prompt substring; fails CI on leak                              |
| CON-002 rate limiter (3/week cap) delays the amendment             | LOW    | This SPEC is high-priority; rate-limiter burst allowance per CON-002 design                                |
| Evolution-log.md write fails and amendment unrecorded              | MEDIUM | REQ-010 is blocking; amendment not considered complete until log write succeeds                            |
| Third-party evaluator skills or plugins break on fresh-context     | LOW    | No known third-party evaluator; document the behavior in release notes                                     |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE zone model — design-constitution must be classified FROZEN)
- **SPEC-V3R2-CON-002** (5-layer amendment graduation protocol — the gate this SPEC passes through)
- **SPEC-V3R2-HRN-001** (HarnessConfig struct — extended with Evaluator.MemoryScope field)

### 9.2 Blocks

- **SPEC-V3R2-HRN-003** (Hierarchical acceptance scoring — builds on fresh-judgment per-sub-criterion semantics)
- **SPEC-V3R2-WF-003** (Multi-mode router — thorough harness mode uses Sprint Contract per this SPEC)

### 9.3 Related

- **SPEC-DESIGN-CONST-AMEND-001** (v2 legacy constitution amendment precedent) — provides the amendment-writing pattern.
- **SPEC-V3R2-EVAL-001** (v3-legacy evaluator profile) — profile schema consumers this flag.
- **SPEC-V3R2-EXT-004** (Versioned migration auto-apply) — BC-V3R2-010 migration runs via the session-start hook.

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01; REQ-002 → AC-02; REQ-003 → AC-03; REQ-004 → AC-03; REQ-005 → AC-04; REQ-006 → AC-05; REQ-007 → AC-06; REQ-008 → AC-09; REQ-009 → AC-07; REQ-010 → AC-09; REQ-011 → AC-04; REQ-012 → AC-08; REQ-013 → AC-07; REQ-014 → AC-04; REQ-015 → AC-09 (canary evidence optional); REQ-016 → solo-mode regression; REQ-017 → AC-07 (leak test); REQ-018 → AC-04; REQ-019 → AC-10.
- Total REQ count: 19 (Ubiquitous 7, Event-Driven 4, State-Driven 3, Optional 2, Unwanted 3)
- Expected AC count: 11
- Wave 1/2 sources:
  - `r1-ai-harness-papers.md` §9 Agent-as-a-Judge (Zhuge et al. 2024, arXiv:2410.10934) — anti-pattern cite
  - `design-principles.md` Principle 4 "Evaluator Judgments Fresh, Contract State Durable"
  - `problem-catalog.md` P-Z01 (HIGH, Evaluator memory cascade)
  - `major-v3-master.md` §5.7 Agent-as-Judge Without Memory (single most important amendment), §8 BC-V3R2-010, §Appendix A (Principle 4 maps to this SPEC)
  - `pattern-library.md` E-1 Agent-as-a-Judge (Intermediate scoring + fresh memory, ADOPT priority 9), E-2 Sprint Contract (Criteria-state negotiation, ADOPT priority 6)
  - `.claude/rules/moai/design/constitution.md` §11 GAN Loop Contract (amendment target)
- Code-side paths:
  - `.claude/rules/moai/design/constitution.md` (amended — §11.4.1 inserted, version bumped, REQ-001, REQ-002)
  - `.moai/config/sections/design.yaml` (modified, REQ-003)
  - `.moai/config/sections/harness.yaml` (modified, REQ-004)
  - `internal/config/types.go` (modified, REQ-005)
  - `internal/config/loader.go` (modified, REQ-011)
  - `internal/config/loader_test.go` (modified, AC-04 fixture)
  - `.claude/agents/moai/evaluator-active.md` (modified, REQ-006)
  - `.claude/skills/moai/moai-workflow-gan-loop/SKILL.md` (modified, REQ-007)
  - `internal/harness/gan_loop.go` (new or modified, REQ-009, REQ-017)
  - `internal/harness/gan_loop_test.go` (new, AC-07 leak test)
  - `.moai/design/v3-research/evolution-log.md` (new or modified, REQ-010)
  - `internal/template/templates/...` (template-first mirrors)

---

End of SPEC.
