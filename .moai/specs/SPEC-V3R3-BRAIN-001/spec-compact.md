---
spec_id: SPEC-V3R3-BRAIN-001
artifact: spec-compact
version: "0.1.0"
source: spec.md
created_at: 2026-05-04
updated_at: 2026-05-04
---

# SPEC-V3R3-BRAIN-001 — Compact Summary

**Title**: `/moai brain` — Idea-to-Item Workflow with Claude Design Handoff Package
**Priority**: P1
**Status**: draft
**Phase**: v3.0.0 — Phase 8 — Brain Workflow Introduction
**Author**: MoAI Plan Workflow
**created_at**: 2026-05-04 | **updated_at**: 2026-05-04
**Harness**: standard

---

## Goal (1 sentence)

Introduce a pre-spec ideation workflow `/moai brain` that converts vague user ideas into validated product proposals with SPEC decomposition candidates AND a paste-ready Claude Design handoff package.

## Workflow Position

```
brain (NEW, once) → [external claude.com Design] → design (path A) → project (once) → plan (per SPEC) → run → sync
```

## 7 Phases

1. Discovery — Socratic interview (clarity score ≤ 3 triggers up to 5 rounds)
2. Diverge — 5-15 angle variations (reuses moai-foundation-thinking)
3. Research — parallel WebSearch + Context7 → research.md
4. Converge — Lean Canvas → ideation.md
5. Critical Evaluation — First Principles + Critical Eval (appended to ideation.md)
6. Proposal — proposal.md with "SPEC Decomposition Candidates" section
7. Handoff — claude-design-handoff/ (5 files: prompt/context/references/acceptance/checklist)

## Output Layout

```
.moai/brain/IDEA-NNN/
├── research.md
├── ideation.md
├── proposal.md
└── claude-design-handoff/{prompt, context, references, acceptance, checklist}.md
```

## EARS Requirements (12 total)

| ID | Type | Summary |
|----|------|---------|
| REQ-BRAIN-001 | Event-Driven | 7 phases execute sequentially |
| REQ-BRAIN-002 | Event-Driven | ≤5 Discovery rounds when clarity ≤ 3 |
| REQ-BRAIN-003 | Event-Driven | Parallel WebSearch + Context7 in single message |
| REQ-BRAIN-004 | Event-Driven | proposal.md "SPEC Decomposition Candidates" section, 2-10 entries |
| REQ-BRAIN-005 | Event-Driven | prompt.md is paste-ready (no MoAI tokens) |
| REQ-BRAIN-006 | Optional | Brand voice integrated when present |
| REQ-BRAIN-007 | Event-Driven | `/moai project --from-brain` consumes proposal.md |
| REQ-BRAIN-008 | Ubiquitous | 16-language neutrality |
| REQ-BRAIN-009 | Event-Driven | Phase 7 exit AskUserQuestion (3 options) |
| REQ-BRAIN-010 | Unwanted | NO auto-execution of /moai project |
| REQ-BRAIN-011 | Unwanted | NO tech-stack in proposal.md |
| REQ-BRAIN-012 | Unwanted | NO prose questions (AskUserQuestion only) |

## Files (17 deliverables = 10 NEW + 7 PATCH)

> Deliverable #12 is a directory containing 8 worked-example sub-files; deliverable count anchors on the directory entry.

**Skills (5)**: workflows/brain.md (NEW), moai-domain-ideation (NEW), moai-domain-research (NEW), moai-domain-design-handoff (NEW), moai/SKILL.md (PATCH)
**Agents (1)**: manager-brain.md (NEW)
**Commands (2)**: moai-brain.md (NEW, Thin Pattern), commands/moai.md (PATCH)
**Go CLI (2)**: internal/cli/brain.go (NEW), root.go (PATCH)
**Templates (2)**: .moai/brain/.gitkeep (NEW), IDEA-EXAMPLE/ (NEW directory; 8 sub-files)
**Workflow Patches (3)**: project.md (PATCH, --from-brain), plan.md (PATCH, decomposition parser), design.md (PATCH, bundle auto-detect)
**Tests (2)**: internal/cli/brain_test.go (NEW), commands_audit_test.go (PATCH)

NEW set: #1, #2, #3, #4, #6, #7, #9, #11, #12, #16 (10)
PATCH set: #5, #8, #10, #13, #14, #15, #17 (7)

## Reuse Inventory

- moai-foundation-thinking: Diverge-Converge, Critical Evaluation, Deep Questioning, First Principles
- moai-workflow-design-import: downstream consumer of Phase 7 output
- AskUserQuestion + ToolSearch protocol
- MoAI Thin Command Pattern, Template-First scaffold

**No new infrastructure required** — pure composition over existing primitives.

## Risk Top 3

1. Phase 7 prompt quality (HIGH) — mitigated by self-review + IDEA-EXAMPLE + --regenerate
2. Brand absent (MEDIUM) — mitigated by default-voice fallback + AskUserQuestion offer
3. Claude.com Design external (MEDIUM) — mitigated by human-readable template, not machine-parsed

## Estimated LOC

~3,050 (skills/agent ~2,180 + Go CLI ~410 + tests ~250 + examples/patches ~210)

## Out of Scope

Web UI for brain (separate SPEC-V3R3-WEB-001), automated claude.com execution, brain v0.2 extensions (Persona, Competitor Matrix, GAN loop), multi-IDEA orchestration, replacing /moai plan, JSON output, non-software ideas, 4-locale docs (deferred to /moai sync).

## Dependencies

- Upstream: NONE (all foundation skills exist)
- Downstream: SPEC-V3R3-WEB-001 (will be self-bootstrapped using brain)

## Acceptance Scenarios

1. Happy path — full 7 phases (REQ-BRAIN-001/005/009)
2. Brand absent — graceful default-voice (REQ-BRAIN-006)
3. WebSearch failure — partial-result tolerance (REQ-BRAIN-003)
4. proposal.md decomposition parseable by /moai plan (REQ-BRAIN-004/007)
5. Language neutrality — Python idea → tech-agnostic proposal (REQ-BRAIN-008/011)
6. AskUserQuestion enforcement — no prose questions (REQ-BRAIN-012)

Plus 5 edge cases (multi-IDEA, mid-workflow interrupt, Phase 7 regenerate, empty decomposition, concurrent invocations).

---

**See**: research.md, spec.md, plan.md, acceptance.md
