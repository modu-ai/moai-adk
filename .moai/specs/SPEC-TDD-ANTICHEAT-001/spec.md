---
id: SPEC-TDD-ANTICHEAT-001
title: "TDD Test-First Anti-Cheat Enforcement"
version: "0.1.0"
status: in-progress
created: 2026-07-24
updated: 2026-07-29
author: manager-spec
priority: P1
phase: "v3.0.2 target"
module: ".claude/ (moai-workflow-tdd skill, manager-develop rule + agent) + internal/template/templates mirrors"
lifecycle: spec-anchored
tags: "tdd, test-first, anti-cheat, harness, quality-gate, red-phase"
tier: S
---

# SPEC-TDD-ANTICHEAT-001 — TDD Test-First Anti-Cheat Enforcement

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-24 | manager-spec | Initial draft — plan-phase authoring (spec/plan/acceptance/progress) |

## §A Context

The harness has the *vocabulary* of RED-first TDD but no *anti-cheat enforcement*. Three concrete gaps make test-first **unfalsifiable** today:

- The run-phase agent instruction says "confirm RED state", but this is advisory only — nothing consumes the RED signal as completion evidence.
- The run-phase self-verification matrix (the `§E` items in the manager-develop prompt template) never requires the RED failure output. A run whose tests were written *after* the implementation produces an identical, fully-passing self-verification matrix. Test-first therefore cannot be distinguished from test-after by the completion report.
- No "delete implementation code written before its failing test" rule exists anywhere in the harness. The Red Flags / Verification checklists in the TDD skill sit in advisory (evolvable) blocks and are consumed by no completion gate.

This SPEC closes those three gaps with the **smallest change that makes test-first falsifiable**: promote the existing advisory checklists into a HARD invariant, add exactly one new self-verification item that requires observed RED evidence, and extend the run-phase agent's forbidden list and RED-phase step. It introduces no new files and no new agents.

## §B Requirements (GEARS)

- **REQ-TDD-001** (Ubiquitous): The `moai-workflow-tdd` skill shall carry a HARD "Test-First Anti-Cheat" section that promotes the existing Red Flags / Verification advisory content into enforced invariants.
- **REQ-TDD-002** (Ubiquitous — invariant i): The Test-First Anti-Cheat section shall require that the RED failure output be **observed and shown as completion evidence** (the verbatim failing-test output, captured before any implementation makes it pass).
- **REQ-TDD-003** (Ubiquitous — invariant ii): The Test-First Anti-Cheat section shall require that **any implementation code written before its failing test be deleted and re-derived test-first**.
- **REQ-TDD-004** (Ubiquitous): The manager-develop self-verification (`§E`) shall carry exactly one new self-verification item requiring the **verbatim RED failing-test output captured before GREEN**, so that test-first becomes falsifiable in the completion matrix.
- **REQ-TDD-005** (State-driven): While adding the new self-verification item, the `§E` structure shall **preserve the existing E1-E7 items unchanged** — the change is additive (ADD, not restructure).
- **REQ-TDD-006** (Ubiquitous): The manager-develop agent Behavioral Contract Forbidden list shall include **"writing implementation before its failing test"**.
- **REQ-TDD-007** (Ubiquitous): The manager-develop agent TDD Cycle RED-phase step (STEP 2) shall carry the **RED-evidence-plus-delete-pre-test-code invariant**.
- **REQ-TDD-008** (Where — falsifiability): Where the new RED-evidence self-verification item is present, a completion matrix that omits observed pre-GREEN RED failing output shall be **structurally incomplete** — a run that skipped RED cannot be reported as clean.

## §C Constraints (DO NOT VIOLATE)

- **C1 — Tier S / minimal**: Smallest change that works. No new files, no new agents. The `§E` matrix gains exactly one item; the existing E1-E7 items are not restructured. Reject scope creep.
- **C2 — Dual-write (Template-First)**: Each of the three affected files has a template-source mirror under `internal/template/templates/.claude/...`. Every edit is a mandatory dual-write to BOTH the operational `.claude/...` copy AND its mirror (6 file edits total), followed by `make build`.
- **C3 — Neutrality (CI-guarded)**: The template-mirror copies MUST NOT introduce internal SPEC IDs, REQ/AC tokens tied to this SPEC, internal dates, or commit SHAs. The new prose is generic mechanism only. The authoritative guard is `TestTemplateNoInternalContentLeak` (`internal/template/internal_content_leak_test.go`), backed by the `template-neutrality-check.yaml` workflow.
- **C4 — Language**: The new prose (skill body, rule, agent body) is authored in English per the instruction-document language policy.
- **C5 — Simplicity bound**: Exactly three logical changes across exactly six files. The skill change is a *promotion* of the already-present Red Flags / Verification content, not net-new bulk. If the implementation exceeds this envelope, stop and re-scope.

## §D Exclusions

This section states what is explicitly **out of scope** for this SPEC.

### Out of Scope — mechanical RED-phase enforcement
- No hook, lint rule, or CI check that mechanically parses a run transcript to prove RED was observed. This SPEC makes test-first *falsifiable via the completion matrix* (a required evidence field); it does not build an automated detector.
- No git-history gate asserting a failing-test commit precedes the implementation commit.

### Out of Scope — §E restructuring
- No renumbering, merging, or rewording of the existing E1-E7 self-verification items. The change is strictly additive (one new item).

### Out of Scope — new agents or files
- No new agent, skill, rule, or command file is created. No changes to any file beyond the three named files and their template mirrors.

### Out of Scope — DDD cycle
- The ANALYZE-PRESERVE-IMPROVE (DDD) cycle path is untouched. This SPEC binds the RED-GREEN-REFACTOR (TDD) surface only.

## §E Cross-References

- `.claude/skills/moai-workflow-tdd/SKILL.md` — Red Flags / Verification advisory blocks to promote (invariant source).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` `§E` — Self-Verification Deliverables E1-E7 (add one item).
- `.claude/agents/moai/manager-develop.md` — Behavioral Contract Forbidden + TDD Cycle STEP 2.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the "no unobserved-verification-claim" invariant this SPEC operationalizes for the RED phase.
- `internal/template/internal_content_leak_test.go` `TestTemplateNoInternalContentLeak` — neutrality CI guard.
