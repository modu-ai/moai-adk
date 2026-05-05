---
spec_id: SPEC-V3R3-BRAIN-001
artifact: research
version: "0.1.0"
created_at: 2026-05-04
author: MoAI Plan Workflow
status: pre-spec research
---

# Research: SPEC-V3R3-BRAIN-001 — `/moai brain` Idea-to-Item Workflow

**Date**: 2026-05-04
**Author**: MoAI Plan Workflow
**Status**: pre-spec research

---

## 1. Codebase Context

### 1.1 Existing Thinking Foundation (reuse target)

**Skill**: `.claude/skills/moai-foundation-thinking/SKILL.md` (16,470 bytes)

The skill consolidates three integrated frameworks already used across MoAI:

| Framework | Module Path | Reused By |
|-----------|-------------|-----------|
| Diverge-Converge (5-phase: Requirements → Diverge 20-50 → Cluster → Converge → Document) | `modules/diverge-converge.md` | manager-strategy, team-reader analyst |
| Critical Evaluation (7-step: Restate → Evidence → Fallacies → Assumptions → Alternatives → Contradictions → Burden) | `modules/critical-evaluation.md` | manager-strategy, plan-auditor |
| Deep Questioning (6-layer progressive inquiry) | `modules/deep-questioning.md` | manager-spec, manager-strategy |
| First Principles (5-phase: Audit → Decompose → Generate → Trade-off → Bias Check) | `modules/first-principles.md` | manager-strategy |
| Sequential Thinking MCP | (`--deepthink` flag) | Architecture decisions |

**Key finding**: The brain workflow's Phase 2 (Diverge), Phase 4 (Converge), Phase 5 (Critical Evaluation), and Discovery Phase 1 (Deep Questioning) are **direct compositions** over `moai-foundation-thinking` — no new ideation logic is required at the framework layer. The new `moai-domain-ideation` skill is a thin orchestrator that invokes the foundation skill and adds artifact-shaping (Lean Canvas section assembly, SPEC-decomposition list extraction).

### 1.2 Existing Design Workflow (path A target)

**Skill**: `.claude/skills/moai-workflow-design-import/SKILL.md`
**Workflow**: `.claude/skills/moai/workflows/design.md`

`/moai design --path A --bundle <path>` already accepts a Claude Design handoff bundle and writes machine-generated artifacts to `.moai/design/`. The brain workflow's Phase 7 produces a bundle compatible with this existing import path:

```
.moai/brain/IDEA-001/claude-design-handoff/   ← brain Phase 7 output (this SPEC)
        │
        ▼ user external action (claude.com Design)
        │
.moai/design/                                  ← existing import workflow output
```

**Implication**: Phase 7 emits a directory shape that `moai-workflow-design-import` can ingest after the user's external claude.com Design execution. No changes to the import skill are required for this SPEC, only an additive bundle-resolution helper in `design.md` (deliverable #15) that auto-detects `.moai/brain/IDEA-XXX/claude-design-handoff/` as a candidate bundle source.

### 1.3 Existing Project Workflow (downstream consumer)

**Workflow**: `.claude/skills/moai/workflows/project.md` (46,614 bytes)
**Agent**: `.claude/agents/moai/manager-project.md`

`/moai project` currently generates `product.md`, `structure.md`, `tech.md` based on:
1. User natural language input
2. Existing codebase scanning (Glob, Grep)
3. Optional brand context from `.moai/project/brand/`

This SPEC adds a fourth, highest-priority input source: `.moai/brain/IDEA-XXX/proposal.md` consumed via the new `--from-brain IDEA-XXX` flag (deliverable #13). When present, proposal.md becomes the primary source of truth for product scope; codebase scanning becomes secondary corroboration.

**Boundary**: brain output does NOT replace project output. They are sequential phases:
- brain = WHO/WHY (vision, market, user, decomposition candidates)
- project = WHAT (product/structure/tech docs grounded in brain proposal)

### 1.4 Existing Plan Workflow (downstream consumer)

**Workflow**: `.claude/skills/moai/workflows/plan.md` (36,153 bytes)
**Agent**: `.claude/agents/moai/manager-spec.md`

`/moai plan` currently accepts a free-form description and produces a single SPEC. The patch in deliverable #14 makes plan aware of `proposal.md`'s "SPEC Decomposition Candidates" section: when a user runs `/moai plan SPEC-XXX-001` (or no SPEC ID at all) within a project that has `.moai/brain/IDEA-XXX/proposal.md`, plan can pre-populate candidate suggestions from the proposal's decomposition list.

**Constraint**: This is **suggestion only** — the user retains full control over which SPECs to commit to. plan.md's existing AskUserQuestion confirmation gate is preserved.

### 1.5 Workflow Router

**Skill**: `.claude/skills/moai/SKILL.md`
**Command**: `.claude/commands/moai/moai.md`

Subcommand routing currently recognizes: `plan`, `run`, `sync`, `design`, `db`, `project`, `fix`, `loop`, `mx`, `feedback`, `review`, `clean`, `codemaps`, `coverage`, `e2e`. Adding `brain` to this router (deliverable #5 + #8) follows the established pattern — the router parses the first positional arg, dispatches to `.claude/skills/moai/workflows/brain.md`, and surfaces no other behavioral change.

### 1.6 Agent Catalog

**Existing managers** (`.claude/agents/moai/manager-*.md`):
- manager-spec, manager-ddd, manager-docs, manager-git, manager-project, manager-strategy
- manager-quality, manager-tdd (planned)

**Pattern observation**: Each manager owns one workflow phase. `manager-brain` (deliverable #6) extends this pattern, owning phases 1-7 of the brain workflow. It delegates to:
- `moai-foundation-thinking` (Phase 1 Discovery, Phase 2 Diverge, Phase 4 Converge, Phase 5 Critical Eval)
- `moai-domain-research` (Phase 3 Research) — new skill in this SPEC
- `moai-domain-ideation` (artifact assembly, Lean Canvas, decomposition list) — new skill in this SPEC
- `moai-domain-design-handoff` (Phase 7) — new skill in this SPEC

### 1.7 Template-First Discipline

[HARD] Every new file under `.claude/`, `.moai/` MUST also exist in `internal/template/templates/`.

The 17 deliverables include `internal/template/templates/.moai/brain/.gitkeep` and a worked example directory `IDEA-EXAMPLE/` so newly scaffolded projects (`moai init`) include the brain output directory. Skills, agents, and commands all have template mirrors.

### 1.8 Thin Command Pattern

**Rule** (coding-standards.md): Command body under 20 LOC, delegate to skill via "Use the {skill} skill to execute {workflow}".

Deliverable #7 (`.claude/commands/moai-brain.md`) follows this pattern. The rich workflow logic lives in `.claude/skills/moai/workflows/brain.md` (deliverable #1, ~600 lines).

---

## 2. Anthropic / External Best-Practice References

### 2.1 Ideation-to-Spec Pipelines

**Pattern (industry-standard)**: Lean Canvas (Ash Maurya, 2010) and Business Model Canvas (Osterwalder) are the de facto ideation-to-proposal artifacts in startup methodology. Both decompose vague ideas into 9 testable blocks: Problem, Customer Segments, Unique Value Proposition, Solution, Channels, Revenue Streams, Cost Structure, Key Metrics, Unfair Advantage. The brain workflow's Phase 4 (Converge → ideation.md with Lean Canvas) adopts this canonical format.

**Source**: https://leanstack.com/lean-canvas (publicly documented)

**Adaptation**: For software-development context (MoAI's domain), the brain workflow weights "Problem" + "Customer Segments" + "Solution" + "Key Metrics" most heavily. Revenue/Cost blocks remain present but are deferred when the product is internal tooling.

### 2.2 First Principles Decomposition (Phase 5 Critical Evaluation)

**Pattern**: Elon Musk's first-principles methodology (popularized via SpaceX/Tesla decisions) breaks problems to fundamental truths rather than reasoning by analogy. `moai-foundation-thinking/modules/first-principles.md` already encodes this as a 5-phase process. Phase 5 of the brain workflow invokes this module on the ideation.md candidates surfaced in Phase 4.

### 2.3 Claude Design Handoff Best Practices

**Source**: Anthropic's Claude Design product documentation (https://claude.com/product/claude-design — publicly accessible)

**Key practices** the Phase 7 prompt template adopts:
1. **Concrete reference URLs**: paste URLs to existing sites that exemplify desired feel/aesthetic — Claude Design is multimodal and will consume the visuals.
2. **Explicit acceptance criteria**: state non-negotiable design quality gates (WCAG AA, mobile-first responsive breakpoints, brand color hex codes).
3. **Brand voice section**: paste the brand voice excerpts directly into the prompt — Claude Design does not have access to `.moai/project/brand/`.
4. **Out-of-scope list**: explicitly state what is NOT to be designed (e.g., "do not design admin/internal pages", "do not include illustrations beyond hero").
5. **Tech-stack neutral**: do NOT specify React vs Vue — Claude Design produces design artifacts (Figma-equivalent), not code. Tech stack is decided downstream in `/moai design --path A`.

**Implication for prompt.md template**: 5-section structure (Goal, References, Brand Voice, Acceptance, Out-of-scope) maps directly onto these 5 best practices. The handoff package's `references.md` and `acceptance.md` files are extracted-out versions of sections 2 and 4 to enable per-section regeneration if the user wants to iterate.

### 2.4 Parallel Tool-Call Best Practice

**Source**: Anthropic's Claude documentation on tool use — https://docs.anthropic.com/en/docs/build-with-claude/tool-use

When multiple independent tool calls are needed, issuing them in a single message (multiple `<tool_use>` blocks) is documented as 50-70% faster than sequential calls. REQ-BRAIN-003 mandates this for Phase 3 Research (parallel WebSearch + Context7).

### 2.5 Socratic Interview / Round Limits

**Source**: AskUserQuestion protocol (this project's `.claude/rules/moai/core/askuser-protocol.md`)

- Maximum 4 questions per AskUserQuestion call (Claude Code hard limit)
- Maximum 4 options per question
- First option must carry `(권장)` / `(Recommended)` suffix
- Free-form prose questions prohibited

REQ-BRAIN-002 inherits these limits. Phase 1 Discovery is bounded to 5 rounds × 4 questions = max 20 clarifying questions. Practice (referenced in §3.3 below) suggests 1-2 rounds typically achieves 100% intent clarity for ideation-stage requests, so 5 is a safe upper bound.

---

## 3. Decisions Locked by Conversation Rounds

The fourth Socratic round (referenced in user's spec spawn prompt) locked these design decisions:

### 3.1 Workflow Position

`brain` is a NEW upstream phase BEFORE `project`. Order:
```
brain (once) → [user external claude.com Design] → design (path A) → project (once) → plan (per SPEC) → run → sync
```

`brain` and `project` are run-once artifacts; `plan/run/sync` repeat per SPEC.

### 3.2 Output Format

Markdown only (decision in user's spawn prompt §"Out of Scope"). No JSON canvas, no YAML proposal. Rationale: human review velocity > machine parsing — proposal.md is read by humans before consumed by `/moai project --from-brain`. The decomposition list inside proposal.md is the only machine-parsed section, and uses a simple bullet-list grammar.

### 3.3 Discovery Round Cap

Max 5 rounds (REQ-BRAIN-002). User decision: ideation-stage requests rarely exceed 2-3 rounds in practice; 5 is a safety ceiling, not a target.

### 3.4 Phase 7 Mandatory Inclusion

Phase 7 (Claude Design Handoff Package) is NOT optional. User correction during the fourth round: "you forgot Claude Design handoff — that's the whole point of the workflow". This SPEC treats the handoff package as a first-class deliverable with its own 5-file structure (prompt.md / context.md / references.md / acceptance.md / checklist.md).

### 3.5 Tech-Stack Agnostic Output (16-language neutrality)

REQ-BRAIN-008 enforces that brain artifacts do not assume a specific programming language or framework. Even if the user's idea mentions "Python web app", the proposal.md should describe the proposal at the product level (not "FastAPI + PostgreSQL"). Tech selection is deferred to `/moai project` and `/moai plan`.

### 3.6 Self-Bootstrap (Web SPEC will use Brain)

The Web dashboard SPEC (SPEC-V3R3-WEB-001) will be ideated USING the brain workflow once this SPEC ships. This is the canonical proof-of-concept. Note: this dependency direction means SPEC-V3R3-WEB-001 cannot start until SPEC-V3R3-BRAIN-001 is merged.

---

## 4. Risk Surface

### 4.1 Phase 7 Prompt Quality Risk (HIGH)

**Risk**: A poorly-templated prompt.md produces a Claude Design output that misses the user's brand voice or feature scope, requiring manual editing — defeating the "paste-ready" promise of REQ-BRAIN-005.

**Mitigation**:
- prompt.md template includes a self-review checklist before final write (manager-brain validates structure presence)
- Phase 7 acceptance gate: AskUserQuestion confirms user reviewed handoff package before workflow exit
- IDEA-EXAMPLE/ template ships with a worked example so users see the expected output shape
- `--regenerate` flag (REQ-BRAIN-009 option c) allows iteration without re-running phases 1-6

### 4.2 Brand Context Absent (MEDIUM)

**Risk**: First-time projects without `.moai/project/brand/` populated produce generic-voice prompts. REQ-BRAIN-006 specifies brand integration but does not define the absent-fallback.

**Mitigation**:
- When brand context absent, Phase 7 emits a "Brand Voice (default — please customize)" section with a placeholder warning
- AskUserQuestion offers user the option to run brand interview before Phase 7 (or accept default)
- acceptance.md scenario #2 verifies graceful fallback (covered in this SPEC's acceptance.md)

### 4.3 Claude.com Design External Dependency (MEDIUM)

**Risk**: Claude Design product UX changes break the prompt template assumptions; the workflow has no direct integration test with claude.com.

**Mitigation**:
- prompt.md template is designed to be human-readable instructions, not machine-formatted — robust to UX changes
- references.md and acceptance.md are static design-quality artifacts independent of Claude Design's UI
- This SPEC's acceptance scenario #1 stops at "handoff package generated" — does not require Claude Design execution

### 4.4 Self-Bootstrap Order (LOW)

**Risk**: Pressure to start SPEC-V3R3-WEB-001 before BRAIN ships could lead to brain workflow being designed around web-specific assumptions, breaking generality.

**Mitigation**:
- This SPEC explicitly lists "Web UI for brain" in Out-of-Scope (web SPEC depends on brain, not vice versa)
- REQ-BRAIN-008 (16-language neutrality) is testable at acceptance — see scenario #5

### 4.5 SPEC Decomposition Format Drift (LOW)

**Risk**: proposal.md's "SPEC Decomposition Candidates" section format diverges from what `/moai plan --from-brain` expects.

**Mitigation**:
- proposal.md template fixes the section grammar: `### SPEC Decomposition Candidates` heading + bullet list `- SPEC-XXX-NNN: <one-line scope>`
- plan.md patch (deliverable #14) implements a defensive parser that surfaces non-conforming entries as warnings, never errors
- Acceptance scenario #4 verifies parseability

---

## 5. Reuse Inventory

| Asset | Status | Used By |
|-------|--------|---------|
| moai-foundation-thinking (Diverge-Converge) | exists | brain Phase 2, 4 |
| moai-foundation-thinking (Critical Evaluation) | exists | brain Phase 5 |
| moai-foundation-thinking (Deep Questioning) | exists | brain Phase 1 Discovery |
| moai-foundation-thinking (First Principles) | exists | brain Phase 5 (alongside Critical Eval) |
| moai-workflow-design-import | exists | downstream consumer of Phase 7 output |
| moai-workflow-spec | exists | downstream consumer of proposal.md |
| moai-workflow-project | exists | downstream consumer (--from-brain flag patch) |
| AskUserQuestion + ToolSearch protocol | exists | Phase 1 Discovery, Phase 6 confirmation, Phase 7 next-action |
| MoAI thin command pattern | exists | deliverable #7 conformance |
| Template-First scaffold pattern | exists | deliverable #11, #12 |

**No new infrastructure is required**. The 17 deliverables are pure compositions over existing primitives.

---

## 6. Estimated Effort

| Component | LOC | Confidence |
|-----------|-----|------------|
| `.claude/skills/moai/workflows/brain.md` (orchestration) | ~600 | High |
| `moai-domain-ideation/SKILL.md` (Lean Canvas + decomposition assembly) | ~400 | High |
| `moai-domain-research/SKILL.md` (parallel WebSearch + Context7) | ~350 | Medium |
| `moai-domain-design-handoff/SKILL.md` (prompt template) | ~450 | Medium |
| `manager-brain.md` (agent definition) | ~250 | High |
| `moai-brain.md` (thin command, ≤20 LOC body + frontmatter) | ~80 | High |
| `moai.md` patch (subcommand router) | ~30 | High |
| `internal/cli/brain.go` (Go entry) | ~150 | High |
| `internal/cli/root.go` patch | ~10 | High |
| `internal/template/templates/.moai/brain/IDEA-EXAMPLE/` (5 files × ~50 lines) | ~250 | Medium |
| `project.md` patch (--from-brain flag) | ~80 | High |
| `plan.md` patch (decomposition parser) | ~100 | Medium |
| `design.md` patch (bundle auto-detect) | ~40 | High |
| `internal/cli/brain_test.go` | ~200 | High |
| `internal/template/commands_audit_test.go` extension | ~50 | High |
| **Total** | **~3,040** | Medium-High |

(Plan-spawn estimate of ~2,400 was a floor; this SPEC adds explicit research and handoff sub-skills which lift the realistic estimate. Both estimates within the same order of magnitude — no SPEC-split needed.)

---

## 7. Open Questions Resolved by This Research

1. **Q**: Does brain replace `/moai plan` ideation? **A**: No. plan retains its current ideation hooks; brain is upstream-only. (decision §3.1)
2. **Q**: Is JSON output supported? **A**: No, markdown only. (decision §3.2)
3. **Q**: Can users skip Phase 7? **A**: No, mandatory per user correction. (decision §3.4)
4. **Q**: Should brain auto-trigger `/moai project`? **A**: No, AskUserQuestion offers it as REQ-BRAIN-009 option (a) — user controls. (decision §3.6 + REQ-BRAIN-009)
5. **Q**: How is brand context absence handled? **A**: Default-voice fallback with user customization prompt. (mitigation §4.2)

---

## 8. References

- `.claude/skills/moai-foundation-thinking/SKILL.md` (Diverge-Converge, Critical Evaluation, Deep Questioning, First Principles)
- `.claude/skills/moai-workflow-design-import/SKILL.md` (downstream import path)
- `.claude/skills/moai/workflows/project.md` (downstream project consumer)
- `.claude/skills/moai/workflows/plan.md` (downstream plan consumer)
- `.claude/skills/moai/workflows/design.md` (path A integration)
- `.claude/rules/moai/core/askuser-protocol.md` (Socratic interview rules)
- `.claude/rules/moai/development/coding-standards.md` (Thin Command Pattern, frontmatter schema)
- Lean Canvas (Ash Maurya, 2010) — https://leanstack.com/lean-canvas
- Anthropic Claude Design product documentation — https://claude.com/product/claude-design
- Anthropic tool-use parallel-call documentation — https://docs.anthropic.com/en/docs/build-with-claude/tool-use

---

**Status**: Research complete. Spec.md / plan.md / acceptance.md follow.
