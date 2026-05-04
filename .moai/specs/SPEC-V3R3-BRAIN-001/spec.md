---
id: SPEC-V3R3-BRAIN-001
version: "0.1.0"
status: draft
created_at: 2026-05-04
updated_at: 2026-05-04
author: MoAI Plan Workflow
priority: P1
labels: [brain, ideation, workflow, handoff, claude-design, v3r3]
issue_number: null
phase: "v3.0.0 — Phase 8 — Brain Workflow Introduction"
module: "internal/cli/, .claude/skills/moai/workflows/, .claude/skills/moai-domain-ideation/, .claude/skills/moai-domain-research/, .claude/skills/moai-domain-design-handoff/, .claude/agents/manager-brain.md, .claude/commands/moai-brain.md, .moai/brain/"
dependencies: []
related_specs: [SPEC-V3R3-WEB-001]
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "brain, ideation, workflow, handoff, claude-design, v3r3"
harness: standard
---

# SPEC-V3R3-BRAIN-001: `/moai brain` — Idea-to-Item Workflow with Claude Design Handoff Package

## HISTORY

| Version | Date       | Author | Description |
|---------|------------|--------|-------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow | Initial SPEC — `/moai brain` pre-spec ideation workflow with 7 phases producing proposal.md + Claude Design handoff package. Self-bootstrap pattern: SPEC-V3R3-WEB-001 will be ideated USING this workflow once shipped. |

---

## 1. Goal

Introduce a brand-new pre-spec ideation workflow `/moai brain` that converts vague user ideas into validated product proposals AND produces a ready-to-use Claude Design handoff package. The workflow is the missing upstream phase that converts "I have a fuzzy idea" into "I have a concrete product proposal with SPEC decomposition candidates AND a paste-ready prompt for claude.com Design".

### 1.1 Workflow Chain (post-shipping)

```
/moai brain "idea"
    ↓ produces
.moai/brain/IDEA-XXX/{research.md, ideation.md, proposal.md, claude-design-handoff/}

[USER EXTERNAL ACTION: paste handoff package into claude.com Design]

/moai design --path A --bundle <claude-design-output>   (existing, reused)
/moai project --from-brain IDEA-XXX                     (proposal.md → product/structure/tech.md)
/moai plan SPEC-XXX-NNN                                  (per SPEC, may pre-populate from proposal.md decomposition)
[Run Gate: worktree | sequential | parallel]
/moai run SPEC-XXX-NNN
/moai sync
```

### 1.2 Cardinality

| Phase | Cardinality | Reason |
|-------|-------------|--------|
| brain | once per project | Vision is set once; if vision changes, run brain again with new IDEA-NNN |
| project | once per project | Product/structure/tech docs are project-scoped, refreshed on major changes |
| plan / run / sync | per SPEC (many) | SPECs are the unit of work — brain emits 2-10 candidates per IDEA |

---

## 2. Phase Structure

The brain workflow comprises 7 phases. Phases 1-6 produce the product proposal; Phase 7 produces the Claude Design handoff package.

| Phase | Purpose | Inputs | Outputs |
|-------|---------|--------|---------|
| 1. Discovery | Socratic interview to clarify vague idea | User input | clarified_intent (in-memory JSON) |
| 2. Diverge | Reuse `moai-foundation-thinking` Diverge-Converge — generate 5-15 angle variations | clarified_intent | divergent_concepts list (in-memory) |
| 3. Research | Parallel WebSearch + Context7 + market scan via `moai-domain-research` skill | divergent_concepts | `research.md` |
| 4. Converge | Reduce to 1-3 viable concepts; build Lean Canvas section | research.md + divergent_concepts | `ideation.md` (with Lean Canvas) |
| 5. Critical Evaluation | First Principles + Critical Evaluation from `moai-foundation-thinking` | ideation.md | evaluation_report inline in `ideation.md` |
| 6. Proposal | Final product proposal with SPEC decomposition candidate list | ideation.md + evaluation | `proposal.md` |
| 7. Claude Design Handoff Package | Generate paste-ready prompt + curated references + acceptance criteria for claude.com Design | proposal.md + brand context (if exists) | `claude-design-handoff/{prompt.md, context.md, references.md, acceptance.md, checklist.md}` |

---

## 3. Output Directory Layout

```
.moai/brain/
└── IDEA-001/                              ← auto-incremented per project
    ├── research.md                         ← Phase 3 output: market/tech research
    ├── ideation.md                         ← Phase 4-5 output: Lean Canvas + critical eval
    ├── proposal.md                         ← Phase 6 output: final proposal + SPEC decomposition
    └── claude-design-handoff/              ← Phase 7 output (5 files)
        ├── prompt.md                       ← paste-ready prompt for claude.com Design
        ├── context.md                      ← brand voice + Lean Canvas summary
        ├── references.md                   ← curated reference sites/screenshots
        ├── acceptance.md                   ← design quality criteria (WCAG, brand consistency)
        └── checklist.md                    ← 5-step user external action guide
```

IDEA-NNN auto-increments. Multiple IDEAs per project are permitted but conventionally one is used.

---

## 4. EARS Requirements

### 4.1 Workflow Execution

**REQ-BRAIN-001** (Event-Driven, Mandatory):
WHEN user invokes `/moai brain "<idea>"`, THE system SHALL execute Phase 1-7 sequentially, producing all artifacts under `.moai/brain/IDEA-XXX/`.

**REQ-BRAIN-002** (Event-Driven, Mandatory):
WHEN the user idea has a low clarity score, THE Discovery phase SHALL conduct Socratic interview rounds via AskUserQuestion before proceeding to Diverge, using the canonical 5-dimension clarity-scoring algorithm defined in `.claude/skills/moai/workflows/plan.md` Phase 0.3 for the score itself, but applying the brain-specific score-to-rounds mapping below (which inverts plan-workflow's mapping because brain's purpose is ideation depth, not requirements speed): score 1-3 → up to 5 rounds (high ambiguity needs deep clarification), score 4-6 → up to 3 rounds, score 7-10 → up to 1 round (verify intent only). Maximum 5 rounds total to comply with AskUserQuestion practical limits.

**REQ-BRAIN-003** (Event-Driven, Mandatory):
WHEN Phase 3 Research executes, THE system SHALL invoke WebSearch and Context7 MCP in parallel (single message, multiple tool calls) and produce `research.md` with all sources cited.

### 4.2 Artifact Contracts

**REQ-BRAIN-004** (Event-Driven, Mandatory):
WHEN Phase 6 Proposal is generated, `proposal.md` SHALL include a "SPEC Decomposition Candidates" section listing 2-10 candidate SPEC IDs with one-line scope each, using grammar `- SPEC-{DOMAIN}-{NUM}: {scope}`.

**REQ-BRAIN-005** (Event-Driven, Mandatory):
WHEN Phase 7 Handoff is generated, `claude-design-handoff/prompt.md` SHALL be paste-ready (no MoAI-specific tokens, no internal SPEC IDs) so user can copy-paste into claude.com Design without editing.

**REQ-BRAIN-006** (Optional, Conditional):
WHERE `.moai/project/brand/` exists with `brand-voice.md`, Phase 7 SHALL incorporate brand voice into `prompt.md` and `context.md`.

### 4.3 Downstream Integration

**REQ-BRAIN-007** (Event-Driven, Mandatory):
WHEN `/moai project --from-brain IDEA-XXX` is invoked AFTER brain completes, THE project workflow SHALL load `proposal.md` and use it as primary input for `product.md`/`structure.md`/`tech.md` generation, treating user codebase as secondary corroboration.

### 4.4 Constraints

**REQ-BRAIN-008** (Ubiquitous, Mandatory):
THE brain workflow SHALL be language-neutral — neither prompt templates nor research patterns SHALL favor specific tech stack or programming language. The 16 languages supported by MoAI-ADK (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift) are equally valid output targets.

**REQ-BRAIN-009** (Event-Driven, Mandatory):
WHEN brain completes Phase 7 successfully, THE system SHALL emit AskUserQuestion offering: (a) proceed to `/moai project --from-brain IDEA-XXX` (Recommended), (b) review artifacts manually first, (c) regenerate handoff package with adjustments.

### 4.5 Negative Requirements (Unwanted Behavior)

**REQ-BRAIN-010** (Unwanted, Mandatory):
THE system SHALL NOT auto-execute `/moai project` after brain completes — user explicit choice via REQ-BRAIN-009 AskUserQuestion is required.

**REQ-BRAIN-011** (Unwanted, Mandatory):
THE Phase 6 proposal.md SHALL NOT specify implementation language or framework versions — tech selection is deferred to `/moai project` and `/moai plan`.

**REQ-BRAIN-012** (Unwanted, Mandatory):
THE Phase 1 Discovery SHALL NOT use free-form prose questions — all clarification questions MUST go through AskUserQuestion with ToolSearch preload (per `.claude/rules/moai/core/askuser-protocol.md`).

---

## 5. Files to Create / Modify

**Total: 17 deliverables (10 NEW + 7 PATCH)**

> Deliverable count anchors at 17 logical entries. Deliverable #12 is a single template directory whose 8 sub-files (enumerated in §5.5) are components of that one deliverable. Actual file-system artifact count = 16 single-file entries (#1–#11, #13–#17) + 8 sub-files of #12 = 24 files. Audit accounting uses the deliverable count (17).

### 5.1 Skills (5)

| # | Path | Type | Purpose |
|---|------|------|---------|
| 1 | `.claude/skills/moai/workflows/brain.md` | NEW (~600 LOC) | Workflow orchestration; phase contract; AskUserQuestion sequences |
| 2 | `.claude/skills/moai-domain-ideation/SKILL.md` | NEW (~400 LOC) | Diverge/Converge/Lean Canvas patterns (thin orchestrator over `moai-foundation-thinking`) |
| 3 | `.claude/skills/moai-domain-research/SKILL.md` | NEW (~350 LOC) | WebSearch/Context7/market-scan parallel patterns |
| 4 | `.claude/skills/moai-domain-design-handoff/SKILL.md` | NEW (~450 LOC) | Claude Design prompt template + brand voice integration |
| 5 | `.claude/skills/moai/SKILL.md` | PATCH | Add `brain` to subcommand router (Priority 1 routing) |

### 5.2 Agents (1)

| # | Path | Type | Purpose |
|---|------|------|---------|
| 6 | `.claude/agents/manager-brain.md` | NEW (~250 LOC) | Brain workflow coordinator; delegates to ideation/research/handoff skills |

### 5.3 Commands (2)

| # | Path | Type | Purpose |
|---|------|------|---------|
| 7 | `.claude/commands/moai-brain.md` | NEW (Thin Command, ≤20 LOC body) | `/moai:brain` slash command wrapper |
| 8 | `.claude/commands/moai.md` | PATCH | Recognize `brain` as subcommand of `/moai` |

### 5.4 Go CLI (2)

| # | Path | Type | Purpose |
|---|------|------|---------|
| 9 | `internal/cli/brain.go` | NEW (~150 LOC) | `moai brain` terminal command (delegates to slash by spawning Claude Code session OR prints instructions) |
| 10 | `internal/cli/root.go` | PATCH | Register `brainCmd` with cobra root |

### 5.5 Output Directory Templates (2)

| # | Path | Type | Purpose |
|---|------|------|---------|
| 11 | `internal/template/templates/.moai/brain/.gitkeep` | NEW | Ensure dir exists in scaffolded projects |
| 12 | `internal/template/templates/.moai/brain/IDEA-EXAMPLE/` (directory; contains 8 sub-files enumerated below) | NEW | Worked example handoff package for user reference |

Sub-files of deliverable #12 (each ~50 LOC, all NEW; per `plan.md` §A6.2):

| Sub-path | Purpose |
|----------|---------|
| `IDEA-EXAMPLE/research.md` | Worked example: market research for hypothetical SaaS |
| `IDEA-EXAMPLE/ideation.md` | Worked example: Lean Canvas filled in |
| `IDEA-EXAMPLE/proposal.md` | Worked example: 3 SPEC decomposition candidates |
| `IDEA-EXAMPLE/claude-design-handoff/prompt.md` | Worked example: paste-ready Claude Design prompt |
| `IDEA-EXAMPLE/claude-design-handoff/context.md` | Worked stub: brand voice + Lean Canvas summary |
| `IDEA-EXAMPLE/claude-design-handoff/references.md` | Worked stub: curated reference URLs/screenshots |
| `IDEA-EXAMPLE/claude-design-handoff/acceptance.md` | Worked stub: design quality criteria |
| `IDEA-EXAMPLE/claude-design-handoff/checklist.md` | Worked stub: 5-step user external-action guide |

### 5.6 Workflow Integration Patches (3)

| # | Path | Type | Purpose |
|---|------|------|---------|
| 13 | `.claude/skills/moai/workflows/project.md` | PATCH | Accept `--from-brain IDEA-XXX` flag; consume proposal.md when generating product/structure/tech.md |
| 14 | `.claude/skills/moai/workflows/plan.md` | PATCH | When proposal.md detected, pre-populate SPEC candidates from decomposition list |
| 15 | `.claude/skills/moai/workflows/design.md` | PATCH | Auto-detect `.moai/brain/IDEA-XXX/claude-design-handoff/` as path A bundle source |

### 5.7 Tests (2)

| # | Path | Type | Purpose |
|---|------|------|---------|
| 16 | `internal/cli/brain_test.go` | NEW (~200 LOC) | Unit tests for brain CLI command |
| 17 | `internal/template/commands_audit_test.go` | PATCH | Verify new `moai-brain.md` matches Thin Command Pattern |

### 5.8 Documentation (deferred to `/moai sync`)

NOT in this SPEC's deliverables — flagged for follow-up sync:
- `docs-site/content/{ko,en,ja,zh}/v3/workflows/brain.md` (4-locale)

---

## 6. Constraints

[HARD] **16-language neutrality** (REQ-BRAIN-008): Phase 3 research must not auto-suggest specific tech stacks. Even if user mentions "Python", proposal.md describes the proposal at the product level.

[HARD] **Template-First**: Every new file under `.claude/`, `.moai/` MUST also exist in `internal/template/templates/`. (Deliverables #11, #12 enforce this for the brain output directory.)

[HARD] **Thin Command Pattern**: `moai-brain.md` body under 20 LOC (per `.claude/rules/moai/development/coding-standards.md`).

[HARD] **Frontmatter 9-field canonical schema** for spec.md: `id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number`. NEVER `created`/`updated` (legacy, rejected).

[HARD] **Reuse over re-invention**: `moai-foundation-thinking` provides Diverge/Converge/First Principles/Critical Eval. The new ideation skill is a thin orchestrator, NOT a re-implementation.

[HARD] **Parallel execution** (REQ-BRAIN-003): Phase 3 research MUST issue WebSearch + Context7 calls in a single message (multiple tool blocks), not sequentially.

[HARD] **AskUserQuestion enforcement** (REQ-BRAIN-012): All user-facing decision points (Discovery rounds, Phase 6 confirmation, Phase 7 next-action) use AskUserQuestion with ToolSearch preload. Prose questions ending with `?` are prohibited.

[HARD] **harness level: standard**: plan-auditor auto-engaged at Phase 2.3 of `/moai plan` workflow.

---

## 7. Out of Scope (Exclusions — What NOT to Build)

The following are explicitly excluded from THIS SPEC's deliverables:

1. **Web UI for brain**: `moai web` browser dashboard is a SEPARATE SPEC (SPEC-V3R3-WEB-001) that depends on this one. The web SPEC will itself be ideated using this brain workflow (self-bootstrap pattern).
2. **Automated execution of claude.com Design**: User manual external action remains. The handoff package is paste-ready; the actual claude.com Design execution is a human step.
3. **Phase extensions for v0.2**: User Persona Generator, Competitor Matrix, GAN-loop refinement of proposal — deferred to brain v0.2.
4. **Multi-IDEA orchestration**: Parallel ideation across many IDEA-NNN simultaneously — out of scope, single-IDEA workflow only.
5. **Replacing `/moai plan` ideation hooks**: plan.md retains its current ideation behavior. The brain SPEC is upstream-only and additive.
6. **Auto-execution of `/moai plan` after `/moai project`**: User explicitly controls when to enter plan phase.
7. **Alternative IDEA file formats**: Only markdown — no JSON canvas, no YAML proposal.
8. **Brain GAN loop**: Iterative quality refinement of proposal.md via evaluator-active — deferred to brain v0.2.
9. **Brain output for non-software ideas**: brain is scoped to software/product ideation. Non-software brainstorming (e.g., business strategy without software component) is out of scope.
10. **4-locale documentation in this SPEC**: docs-site updates deferred to `/moai sync` after merge.

---

## 8. Acceptance Criteria

See `acceptance.md` for full Given-When-Then scenarios. Minimum 5 scenarios required:

1. Happy path — vague idea → 7 phases → all artifacts produced (REQ-BRAIN-001)
2. Brand context absent — Phase 7 graceful default-voice fallback (REQ-BRAIN-006)
3. Research phase — WebSearch failure handled gracefully (REQ-BRAIN-003)
4. SPEC decomposition list parseable by `/moai plan --from-brain` (REQ-BRAIN-004)
5. Language neutrality — Python-mentioning idea produces tech-stack-agnostic proposal (REQ-BRAIN-008)

---

## 9. Definition of Done

- [ ] All 17 files created/patched per §5
- [ ] All 12 EARS requirements (REQ-BRAIN-001..012) verified via acceptance scenarios
- [ ] Frontmatter validation: 9 required fields present; `created_at`/`updated_at` (NOT `created`/`updated`); labels is YAML array
- [ ] `make build` completes successfully (embedded templates regenerated)
- [ ] `go test ./internal/cli/...` passes including `brain_test.go`
- [ ] `go test ./internal/template/...` passes including extended `commands_audit_test.go`
- [ ] `golangci-lint run ./...` zero warnings
- [ ] plan-auditor PASS on Phase 2.3 of `/moai plan SPEC-V3R3-BRAIN-001`
- [ ] PR merged to main; no follow-up regressions in `/moai run`, `/moai sync`, `/moai project`, `/moai plan` (smoke tests)
- [ ] IDEA-EXAMPLE/ template verified to scaffold correctly via `moai init` → `ls .moai/brain/`

---

## 10. Dependencies and Sequencing

**Upstream dependencies**: None. All foundation skills (`moai-foundation-thinking`, `moai-workflow-design-import`) already exist.

**Downstream blockers** (this SPEC unblocks):
- SPEC-V3R3-WEB-001 (Web dashboard SPEC) — will be ideated using brain workflow once shipped (self-bootstrap proof-of-concept)
- Future ideation-stage SPECs across all v3 phases

**Parallel SPECs**: None — this is a single-domain SPEC with no shared file ownership conflicts.

---

## 11. References

- `research.md` (this SPEC's siblings) — codebase context, Anthropic best practices, decision log
- `plan.md` (this SPEC's siblings) — implementation phases, risk analysis, MX tag plan
- `acceptance.md` (this SPEC's siblings) — 5+ Given-When-Then scenarios
- `.claude/rules/moai/development/coding-standards.md` — Thin Command Pattern, frontmatter schema
- `.claude/rules/moai/core/askuser-protocol.md` — Socratic interview rules
- Lean Canvas (Ash Maurya, 2010) — https://leanstack.com/lean-canvas
- Anthropic Claude Design — https://claude.com/product/claude-design

---

**Status**: draft (audit-ready post-plan-auditor PASS)
