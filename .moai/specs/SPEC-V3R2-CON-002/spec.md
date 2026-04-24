---
id: SPEC-V3R2-CON-002
title: "Constitutional amendment protocol with 5-layer safety gate"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 4 SPEC Writer
priority: P0 Critical
phase: "v3.0.0 — Phase 1 — Constitution & Foundation"
module: "internal/constitution/, .claude/rules/moai/core/, .moai/research/"
dependencies:
  - SPEC-V3R2-CON-001
related_gap: []
related_problem:
  - P-R02
related_pattern:
  - S-4
  - S-5
related_principle:
  - P1
  - P2
  - P12
related_theme: "Layer 1: Constitution"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "v3r2, constitution, amendment, safety-gate, graduation-protocol"
---

# SPEC-V3R2-CON-002: Constitutional amendment protocol with 5-layer safety gate

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | Wave 4 SPEC Writer | Initial draft from master v3 §4 Layer 1, design-constitution §5 generalization |

---

## 1. Goal (목적)

Define the protocol by which FROZEN clauses can be promoted to EVOLVABLE (demotion) and EVOLVABLE clauses can be promoted to FROZEN (promotion), and by which EVOLVABLE clauses can have their text modified under the 5-layer safety gate that already exists in `.claude/rules/moai/design/constitution.md` §5 (FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight).

This SPEC lifts the 5-layer gate from design subsystem scope to **core** constitution scope, so that the full moai rule tree (not just the design pipeline) is protected by the same graduation protocol. It also defines the evolution-log format (`.moai/research/evolution-log.md`) that records every applied amendment with human approval timestamp, canary verdict, and contradiction scan result.

CON-001 provides the zone labels. CON-002 provides the protocol that mutates those labels or their clauses.

## 2. Scope (범위)

### 2.1 In Scope

- `AmendmentProposal` Go type with fields: RuleID, Before, After, Evidence, CanaryResult, Contradicts, Approved, ApprovedBy, ApprovedAt.
- 5-layer interface surface: `FrozenGuard.Check`, `Canary.Evaluate`, `ContradictionDetector.Scan`, `RateLimiter.Admit`, `HumanOversight.Approve`.
- Rate-limiter defaults mirroring design-constitution §5 Layer 4: max 3 amendments per week, 24h cooldown, 50 active learnings cap.
- Canary evaluation runs proposed clause change against the last 3 completed SPECs (per design-constitution §5 Layer 2). Score drop threshold: 0.10.
- Evolution log format: append-only markdown at `.moai/research/evolution-log.md` with YAML-frontmatter entries keyed on `LEARN-YYYYMMDD-NNN`.
- `moai constitution amend` CLI subcommand wrapping the protocol.
- Integration with AskUserQuestion for HumanOversight layer (per design-constitution §5 Layer 5).
- Rollback protocol: if a graduated amendment causes regression (next SPEC score drop >0.10), automatic revert with entry in evolution-log.

### 2.2 Out of Scope

- Zone registry creation (→ SPEC-V3R2-CON-001).
- Rule-tree consolidation (→ SPEC-V3R2-CON-003).
- Evaluator rubric content (→ SPEC-V3R2-HRN-003).
- Non-constitutional memory amendments (lessons.md auto-capture uses a simpler workflow per SPEC-SLQG-001, not this 5-layer gate).
- Design-subsystem-specific graduation (already FROZEN in design-constitution §7; CON-002 cross-references, does not replace).

## 3. Environment (환경)

Current moai state (v2.13.2):

- `.claude/rules/moai/design/constitution.md` v3.3.0 §5 defines 5-layer safety architecture for the **design subsystem only**.
- `.claude/rules/moai/design/constitution.md` §7 defines the Knowledge Graduation Protocol (observation → heuristic → rule → graduated) for design learnings.
- No core-level amendment protocol exists. Changes to `.claude/rules/moai/core/moai-constitution.md` happen via direct git commits today.
- `.moai/research/observations/` directory exists per design-constitution §6 but holds only design-domain learnings. No central `evolution-log.md` yet.
- AskUserQuestion is available as a Claude Code tool (MoAI orchestrator only; subagents prohibited per agent-common-protocol.md).

References: master-v3 §4 Layer 1 Go type sketch; design-principles.md §P12 Constitutional Governance with FROZEN/EVOLVABLE zones; pattern-library.md §S-5 5-Layer Safety.

## 4. Assumptions (가정)

- The 5-layer gate proven in the design subsystem transfers cleanly to core constitution. No new safety mechanisms needed beyond those listed in design-constitution §5.
- Canary evaluation requires access to the last 3 completed SPECs. In a fresh install with <3 SPECs, Canary emits `CanaryUnavailable` and the amendment cannot auto-apply; human-override only.
- HumanOversight layer uses AskUserQuestion. This means amendment proposals can only be confirmed by the MoAI orchestrator, not by subagents (matches existing design constitution §5 Layer 5).
- Evolution log is version-controlled; accidental rollbacks via git revert are acceptable, but the log itself is append-only within a session.
- SPEC-V3R2-CON-001 has shipped and the zone registry exists; CON-002 reads it to identify the target of each amendment.

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-CON-002-001: The system SHALL provide a Go type `internal/constitution.AmendmentProposal` with fields RuleID, Before, After, Evidence, CanaryResult, Contradicts, Approved, ApprovedBy, ApprovedAt.
- REQ-CON-002-002: The system SHALL implement the 5-layer safety gate as a pipeline: FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight. Failure at any layer SHALL halt the pipeline with structured error.
- REQ-CON-002-003: The system SHALL persist applied amendments to `.moai/research/evolution-log.md` as append-only markdown entries with YAML frontmatter.
- REQ-CON-002-004: Each evolution-log entry SHALL contain: `id` (LEARN-YYYYMMDD-NNN), `rule_id` (CONST-V3R2-NNN), `zone_before`, `zone_after`, `clause_before`, `clause_after`, `canary_verdict`, `contradictions`, `approved_by`, `approved_at`.
- REQ-CON-002-005: The FrozenGuard layer SHALL reject any amendment whose target rule has `zone: Frozen` in the registry, unless the amendment is specifically a demotion (Frozen → Evolvable) with supporting evidence.
- REQ-CON-002-006: The RateLimiter SHALL enforce at most 3 accepted amendments per rolling 7-day window, a 24-hour cooldown between acceptances, and a cap of 50 active learnings (per design-constitution §5 Layer 4).
- REQ-CON-002-007: The HumanOversight layer SHALL invoke AskUserQuestion with the proposal diff, canary verdict, and contradiction report before applying the amendment.

### 5.2 Event-driven

- REQ-CON-002-010: WHEN a learning reaches 5 observations with confidence ≥ 0.80 (per design-constitution §6 graduation tier), the system SHALL auto-generate an AmendmentProposal targeting the relevant zone-registry entry.
- REQ-CON-002-011: WHEN an AmendmentProposal passes all 5 safety layers, the system SHALL apply the change to the source rule file, update the zone-registry entry, and append an evolution-log entry within a single atomic operation (either all three succeed or none).
- REQ-CON-002-012: WHEN an AmendmentProposal fails any safety layer, the system SHALL log the proposal to `.moai/research/rejected-amendments/LEARN-YYYYMMDD-NNN.md` with the failure reason and SHALL NOT modify the source file or registry.
- REQ-CON-002-013: WHEN a graduated amendment causes the next SPEC score to drop by more than 0.10 versus the pre-amendment baseline, the system SHALL trigger an automatic rollback: the clause reverts, the evolution-log entry gets a `rolled_back: true` marker, and the learning enters a 30-day cooldown (per design-constitution §14 + `staleness_window_days`).

### 5.3 State-driven

- REQ-CON-002-020: WHILE fewer than 3 completed SPECs exist in `.moai/specs/`, the Canary layer SHALL emit verdict `CanaryUnavailable` and auto-apply paths SHALL be disabled; only explicit human-override amendments proceed.
- REQ-CON-002-021: WHILE an amendment is being applied, concurrent amendment attempts SHALL be rejected with `AmendmentInProgress` error (single-writer lock on evolution-log).
- REQ-CON-002-022: WHILE the evolution log contains a `rolled_back: true` entry for Rule ID X within the last 30 days, new amendment proposals targeting X SHALL be blocked by the RateLimiter.

### 5.4 Optional

- REQ-CON-002-030: WHERE the configuration key `constitution.amendment.auto_apply: false` is set in `.moai/config/sections/constitution.yaml`, the pipeline SHALL halt after ContradictionDetector and surface the proposal via AskUserQuestion even if RateLimiter would admit it.
- REQ-CON-002-031: WHERE `MOAI_CONSTITUTION_DRY_RUN=1` is set, the pipeline SHALL execute all 5 layers but not write to any file, returning only the proposal's predicted outcome.

### 5.5 Complex

- REQ-CON-002-040: IF an AmendmentProposal targets a clause whose registry `canary_gate: false`, THEN the Canary layer SHALL be skipped but the other 4 layers SHALL still execute; ELSE the full 5-layer pipeline runs.
- REQ-CON-002-041: WHILE the ContradictionDetector scans and finds a conflict, WHEN the user confirms via AskUserQuestion "supersede the conflicting rule", THEN the pipeline SHALL produce a compound amendment that modifies both rules atomically and records each modification as a separate evolution-log entry with a shared `batch_id`.
- REQ-CON-002-042: IF the HumanOversight layer times out (no user response within 24 hours), THEN the proposal SHALL be marked as `pending_approval` and preserved in `.moai/research/pending-amendments/`; subsequent pipeline runs SHALL surface pending proposals first.

## 6. Acceptance Criteria

- AC-CON-002-01: Given a valid AmendmentProposal targeting an EVOLVABLE rule, When the pipeline executes and all 5 layers pass, Then the source rule file, zone registry, and evolution-log are updated atomically. (maps REQ-CON-002-011)
- AC-CON-002-02: Given a proposal targeting a FROZEN rule with no explicit demotion evidence, When FrozenGuard.Check runs, Then the proposal is rejected with `FrozenClauseImmutable` error. (maps REQ-CON-002-005)
- AC-CON-002-03: Given 3 amendments applied within the past 7 days, When a 4th proposal reaches RateLimiter, Then RateLimiter.Admit returns false with `RateLimitExceeded`. (maps REQ-CON-002-006)
- AC-CON-002-04: Given a project with 2 completed SPECs only, When Canary.Evaluate is invoked, Then the verdict is `CanaryUnavailable` and the auto-apply path halts. (maps REQ-CON-002-020)
- AC-CON-002-05: Given an applied amendment and the next completed SPEC scores 0.15 below baseline, When the rollback watcher runs, Then the clause is reverted, the evolution-log entry gains `rolled_back: true`, and the learning enters 30-day cooldown. (maps REQ-CON-002-013, REQ-CON-002-022)
- AC-CON-002-06: Given AskUserQuestion is invoked for HumanOversight with proposal diff, When the user selects "Reject", Then the proposal is moved to `rejected-amendments/` and no source file is modified. (maps REQ-CON-002-007, REQ-CON-002-012)
- AC-CON-002-07: Given `MOAI_CONSTITUTION_DRY_RUN=1`, When a full pipeline is invoked, Then no file is written and the returned structure contains the predicted outcome of each layer. (maps REQ-CON-002-031)
- AC-CON-002-08: Given two concurrent `moai constitution amend` invocations, When both attempt to write evolution-log, Then exactly one succeeds and the other exits with `AmendmentInProgress`. (maps REQ-CON-002-021)
- AC-CON-002-09: Given a proposal with `registry.canary_gate: false`, When the pipeline runs, Then layers 1, 3, 4, 5 execute and Canary is skipped. (maps REQ-CON-002-040)
- AC-CON-002-10: Given a ContradictionDetector finds a conflict and the user confirms "supersede", When the compound amendment is applied, Then both affected rules change atomically and two evolution-log entries share a `batch_id`. (maps REQ-CON-002-041)
- AC-CON-002-11: Given HumanOversight times out without user response, When 24 hours elapse, Then the proposal is persisted at `.moai/research/pending-amendments/` and appears first on the next pipeline invocation. (maps REQ-CON-002-042)
- AC-CON-002-12: Given an auto-generated proposal from a 5-observation learning, When the full pipeline succeeds, Then the evolution-log entry links to the source observation (`source_observation_id: LEARN-YYYYMMDD-NNN`). (maps REQ-CON-002-010, REQ-CON-002-004)
- AC-CON-002-13: Given `constitution.amendment.auto_apply: false` in config, When RateLimiter would admit but the policy forbids auto-apply, Then HumanOversight is invoked even for rules that passed all prior layers. (maps REQ-CON-002-030)

## 7. Constraints (제약)

- 5-layer order is FROZEN and may not be reordered (mirrors design-constitution §5 ordering).
- Pass threshold floor remains 0.60 per master-v3 §1.3 FROZEN (cannot be lowered by the amendment protocol itself — preventing the amendment system from weakening its own safety threshold).
- Evolution log is append-only at the file level within a session; edits to older entries are prohibited (separate `rolled_back: true` marker mechanism).
- Rate limiter windows and thresholds (3/week, 24h, 50) are default config values settable via `constitution.yaml` but bounded by FROZEN minima (min cooldown 1h, max 7/week, max cap 100).
- Go module dependencies: no new third-party dependency. AskUserQuestion is invoked via the existing orchestrator channel (MoAI only, not subagents).
- Performance: full pipeline dry-run (dry_run mode) must complete within 5 seconds for proposals targeting rules with ≤3 canary SPECs.

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Approval fatigue — users reject all proposals to avoid friction | MEDIUM | MEDIUM | Defaults favor FROZEN; EVOLVABLE clauses are rare; surface clear proposal summaries with evidence count |
| Canary regressions propagate before rollback | LOW | HIGH | Rollback trigger on >0.10 score drop; 30-day cooldown prevents rapid re-proposal |
| Evolution-log corruption (interrupted write) | LOW | HIGH | Single-writer lock per REQ-CON-002-021; atomic rename on write |
| Users bypass protocol with direct git edits | HIGH | LOW | CI check against zone-registry hash; detect drift via `moai doctor constitution` |
| 5-layer pipeline too slow for dev loop | LOW | MEDIUM | DRY_RUN mode short-circuits; Canary is the only network-call-free-but-compute-heavy layer |

## 9. Dependencies

### 9.1 Blocked by

- SPEC-V3R2-CON-001 (must know which rules are FROZEN vs EVOLVABLE before amending them).

### 9.2 Blocks

- SPEC-V3R2-HRN-002 (evaluator fresh-memory amendment uses this protocol).
- SPEC-V3R2-WF-005 (language rules vs skills boundary decision may trigger CON-002 amendments).

### 9.3 Related

- `.claude/rules/moai/design/constitution.md` §5-7, §14 — existing 5-layer architecture and rollback protocol in design subsystem.
- SPEC-SLQG-001 (lessons.md auto-capture) — distinct simpler workflow; CON-002 is the heavier constitutional path.
- design-principles.md §P12, pattern-library.md §S-5.

## 10. Traceability

- Theme: Layer 1 Constitution (master-v3 §4).
- Principles: P1 SPEC as Constitutional Contract; P2 Constitutional Governance (design-principles.md §P12); P12 File-First Primitives (evolution log is a markdown file).
- Problems: P-R02 Constitutional sprawl (amendment protocol prevents unilateral additions without gate).
- Patterns: S-4 FROZEN + Graduation (pattern-library.md §S-4); S-5 5-Layer Safety (pattern-library.md §S-5).
- Wave 1 sources: R1 §18 Constitutional AI; R1 §16 ADAS anti-pattern flag (meta-agents can drift unsafe — 5-layer gate is directly appropriate defense).
- Wave 2 sources: design-principles.md §P12, pattern-library.md §S-5, design-constitution §5-7 prototype.
