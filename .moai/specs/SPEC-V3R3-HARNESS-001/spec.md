---
id: SPEC-V3R3-HARNESS-001
title: Meta-Harness Skill — Static Core 22 + Dynamic ∞
version: "0.1.2"
status: completed
created_at: 2026-04-26
updated_at: 2026-04-27
author: manager-spec
priority: P0
phase: "v3.0.0 R3 — Phase C — Extreme Aggressive Core"
module: ".claude/skills/moai-meta-harness/, internal/template/templates/.claude/skills/moai-meta-harness/, internal/cli/update/, .moai/archive/skills/v2.16/"
dependencies:
  - SPEC-V3R3-HARNESS-LEARNING-001
related_specs:
  - SPEC-V3R3-PATTERNS-001
  - SPEC-V3R3-DESIGN-PIPELINE-001
  - SPEC-V3R3-PROJECT-HARNESS-001
  - SPEC-AGENCY-ABSORB-001
breaking: true
bc_id:
  - BC-V3R3-007
lifecycle: spec-anchored
labels: [harness, meta-skill, breaking-change, namespace-separation, v3r3, phase-c, extreme-aggressive, apache-2-0-attribution]
related_theme: "Phase C — Extreme Aggressive Core (meta-harness + namespace separation)"
target_release: v2.17.0
issue_number: null
depends_on:
  - SPEC-V3R3-HARNESS-LEARNING-001
---

# SPEC-V3R3-HARNESS-001: Meta-Harness Skill — Static Core 22 + Dynamic ∞

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial draft. v3R3 Phase C P0 — meta-harness skill 신설 + 16 정적 skills 제거 (BC-V3R3-007) + namespace 분리 (moai-*/my-harness-*) + revfactory/harness 7-Phase workflow 흡수 (Apache 2.0 attribution). |
| 0.1.1   | 2026-04-27 | orchestrator | D-1/D-2 plan-audit fix. §3 카테고리 라벨 라이브 트리 기준 정확화 (workflow=10, design=1, foundation=4 명시). plan.md §3.5 staticCoreAllowlist 11개 가공 skill명을 검증된 22개 실명으로 대체. plan.md §6 "11 REQs" → 10. |
| 0.1.2   | 2026-04-27 | orchestrator | T-M5-05 post-merge: status draft → completed. PR #724 머지 완료 (merge commit `ba4545981`). Wave A (M1) + Wave B (M2+M3) + Wave C (M4+M5) 전체 7/7 ACs 충족. main 브랜치에 통합됨. |

---

## 1. Goal (목적)

MoAI-ADK MUST evolve from a fixed-skill catalog (38 skills) to a meta-harness architecture where the static core retains exactly 22 base skills + 1 meta-harness skill (= 23 total), and project-specific skills are generated dynamically into a separated user namespace (`my-harness-*`). The meta-harness skill (`moai-meta-harness`) MUST absorb the revfactory/harness 7-Phase workflow (Apache 2.0 attribution mandatory) and integrate it with MoAI's existing manager/expert/builder agent ecosystem. Sixteen domain/framework/library/platform/tool skills MUST be removed (BC-V3R3-007 BREAKING) and auto-archived by the `moai update` migrator with a 1 minor release grace window.

### 1.1 Background

- The current 38-skill catalog includes domain-specific skills (backend, frontend, database, mobile, db-docs), framework skills (electron), library skills (shadcn, mermaid, nextra), platform skills (auth, deployment, chrome-extension), tool skills (ast-grep), and workflow skills (research, pencil-integration, formats-data). These skills are static and rarely match the actual project's domain.
- revfactory/harness (Apache 2.0) demonstrates a meta-skill pattern that generates project-specific harnesses dynamically via a 7-Phase workflow, with documented +60% effectiveness in A/B comparisons.
- The user-mandated namespace separation (handoff §0.1) draws a hard boundary: `moai-*` (static, MoAI-maintained) vs `my-harness-*` (dynamic, user-customized). This boundary prevents `moai update` from clobbering user customizations and gives users visual cues for what they own.
- SPEC-V3R3-HARNESS-LEARNING-001 (already drafted) provides the auto-evolution layer for the dynamic harness. This SPEC provides the generation layer that LEARNING-001 builds on.
- BREAKING change BC-V3R3-007 is unavoidable: the 16 removed skills no longer fit the meta-harness architecture. Migration is automated via `moai update` with a 1 minor release grace window and explicit `MIGRATION-v2.17.0.md`.

### 1.2 Non-Goals

- This SPEC does NOT implement the 5-layer integration mechanism (Layer 1 frontmatter triggers, Layer 2 workflow.yaml.harness, Layer 3 CLAUDE.md @import markers, Layer 4 workflow static import line, Layer 5 `.moai/harness/` user directory). That work is owned by SPEC-V3R3-PROJECT-HARNESS-001.
- This SPEC does NOT conduct the Socratic interview that drives harness generation. That is owned by SPEC-V3R3-PROJECT-HARNESS-001 (16 questions, 4 rounds).
- This SPEC does NOT implement the auto-evolution / self-learning loop. That is owned by SPEC-V3R3-HARNESS-LEARNING-001.
- This SPEC does NOT modify `.claude/agents/moai/` (22 agents stay unchanged).
- This SPEC does NOT touch `.claude/commands/moai/` routing (gate / context / security cleanup is owned by SPEC-V3R3-CMD-CLEANUP-001).
- This SPEC does NOT add or modify Vibe Design / DTCG token logic (owned by SPEC-V3R3-DESIGN-PIPELINE-001).

---

## 2. Scope

### 2.1 In Scope

- Create `.claude/skills/moai-meta-harness/SKILL.md` containing the revfactory/harness 7-Phase workflow with Apache 2.0 attribution at the top.
- Create the Template-First mirror at `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` so `moai init` deploys the meta-harness skill into every new project.
- Remove the 16 static skills enumerated in §3 (BC-V3R3-007).
- Mirror the removal in `internal/template/templates/.claude/skills/` so `moai init` does not redeploy the removed skills.
- Extend the `moai update` migrator (`internal/cli/update/`) to detect the 16 removed skills in user projects and auto-archive them under `.moai/archive/skills/v2.16/` while installing the new `moai-meta-harness` skill.
- Establish the namespace separation: static `moai-*` (22 base + 1 meta) vs dynamic `my-harness-*` (created by meta-harness, not by this SPEC).
- Author `MIGRATION-v2.17.0.md` migration guide explaining the BC-V3R3-007 breaking change with manual fallback steps.
- Verify new-session auto-activation: when CLAUDE.md is loaded, the system MUST recognize the dynamic harness if it exists (the actual 5-layer wiring is in PROJECT-HARNESS-001; this SPEC verifies the meta-harness skill itself does not block that recognition).
- Update `CHANGELOG.md`, `.moai/release/RELEASE-NOTES-v2.17.0.md` with the BC-V3R3-007 announcement.

### 2.2 Out of Scope

- Socratic interview implementation (PROJECT-HARNESS-001).
- 5-layer integration wiring (PROJECT-HARNESS-001).
- Self-learning evolution (LEARNING-001).
- DTCG token validator and Vibe Design pipeline (DESIGN-PIPELINE-001).
- Pattern Cookbook 6 reference docs (PATTERNS-001 in v2.16).
- Generation of any specific dynamic skill content (e.g., `my-harness-ios-patterns`) — that is the meta-harness's runtime responsibility, not a deliverable of this SPEC.

---

## 3. Sixteen Skills To Remove (BC-V3R3-007 verbatim list)

The following 16 skills MUST be removed (verbatim from handoff §4.2 / §6 v2.17.0 cluster):

| # | Skill ID | Category |
|---|---------|----------|
| 1 | `moai-domain-backend` | domain |
| 2 | `moai-domain-frontend` | domain |
| 3 | `moai-domain-database` | domain |
| 4 | `moai-domain-db-docs` | domain |
| 5 | `moai-domain-mobile` | domain |
| 6 | `moai-framework-electron` | framework |
| 7 | `moai-library-shadcn` | library |
| 8 | `moai-library-mermaid` | library |
| 9 | `moai-library-nextra` | library |
| 10 | `moai-tool-ast-grep` | tool |
| 11 | `moai-platform-auth` | platform |
| 12 | `moai-platform-deployment` | platform |
| 13 | `moai-platform-chrome-extension` | platform |
| 14 | `moai-workflow-research` | workflow |
| 15 | `moai-workflow-pencil-integration` | workflow |
| 16 | `moai-formats-data` | formats |

Total: domain (5) + framework (1) + library (3) + tool (1) + platform (3) + workflow (2) + formats (1) = **16 skills**.

The static core after removal MUST contain exactly (verified against live tree on 2026-04-27):
- `moai-foundation-*` (4): cc, core, quality, thinking
- `moai-workflow-*` (10 — note: research and pencil-integration removed from this group): ddd, design-context, design-import, gan-loop, loop, project, spec, tdd, testing, worktree
- `moai-ref-*` (5): api-patterns, git-workflow, owasp-checklist, react-patterns, testing-pyramid
- `moai-design-*` (1): design-system
- `moai-domain-{copywriting,brand-design}` (2 FROZEN)
- `moai-meta-harness` (1 NEW)

= 22 base + 1 meta = **23 total skills**. Math: 4 + 10 + 5 + 1 + 2 + 1 = 23.

---

## 4. Stakeholders

| Role | Interest |
|------|----------|
| Solo developer | Removes irrelevant skills, gets project-specific harness via 1 command (`/moai project`). |
| Team lead | Each project gets a harness sized to its actual domain rather than a one-size-fits-all catalog. |
| MoAI maintainer | Reduced surface area to maintain (22 base instead of 38), clearer separation between core and project. |
| Plan-auditor | Verifiable boundary between `moai-*` (FROZEN-equivalent for users) and `my-harness-*` (mutable). |
| User upgrading from v2.16.x | Automatic migration via `moai update` with 1 minor grace window, archived skills available for restore. |

---

## 5. Exclusions (What NOT to Build)

[HARD] This SPEC explicitly EXCLUDES the following — building any of these is a scope violation:

1. The Socratic interview UI / questions (owned by SPEC-V3R3-PROJECT-HARNESS-001).
2. The 5-layer integration mechanism (Layer 1-5 listed in handoff §0.2 — owned by SPEC-V3R3-PROJECT-HARNESS-001).
3. The self-learning auto-evolution loop (owned by SPEC-V3R3-HARNESS-LEARNING-001).
4. Generation of specific `my-harness-*` skill content (the meta-harness skill provides the *capability*; running it is a runtime concern, not a SPEC deliverable).
5. Removal of any skill not in the verbatim 16-skill list (e.g., `moai-foundation-*` MUST be preserved).
6. Modification of any agent under `.claude/agents/moai/` (this SPEC is skill-only).
7. Replacement of the `/moai project` command (only PROJECT-HARNESS-001 extends `project.md` Phase 5+).
8. Auto-deletion of user `my-harness-*` skills during `moai update` (user customizations are inviolate).

---

## 6. Requirements (EARS format)

### REQ-HARNESS-001 (Ubiquitous — Meta-Harness Skill Creation + Apache 2.0 Attribution)

The system **shall** include `.claude/skills/moai-meta-harness/SKILL.md` containing the revfactory/harness 7-Phase workflow adapted to MoAI conventions, and the SKILL.md **shall** carry a top-of-file attribution line crediting `revfactory/harness` under Apache 2.0. The Template-First mirror at `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` **shall** be byte-identical to the local copy except for variable substitution placeholders.

### REQ-HARNESS-002 (Unwanted — 16 Static Skills Removal Enforcement)

**If** any of the 16 skills enumerated in §3 are present under `.claude/skills/` after this SPEC merges, **then** the build **shall** fail. The removal **shall** apply to both the local project tree and `internal/template/templates/.claude/skills/`. No exception is permitted: the verbatim 16-skill list is the authoritative source.

### REQ-HARNESS-003 (Ubiquitous — Namespace Separation)

The system **shall** maintain two disjoint skill prefixes: `moai-*` (static, MoAI-maintained, the only prefix permitted under MoAI core skill management) and `my-harness-*` (dynamic, user-managed, generated by `moai-meta-harness`). The `moai update` command **shall not** modify any file under `.claude/skills/my-harness-*/` or `.claude/agents/my-harness/`. The Doctor command **shall** detect and warn on any skill whose name begins with `moai-` outside the static core or with `my-harness-` outside the user area.

### REQ-HARNESS-004 (Ubiquitous — revfactory/harness 7-Phase Workflow Absorption)

`moai-meta-harness/SKILL.md` **shall** document the seven phases of the revfactory/harness workflow (briefly: Discovery, Analysis, Synthesis, Skeleton, Customization, Evaluation, Iteration) adapted to MoAI vocabulary. The skill body **shall** reference MoAI's existing agents (manager-spec, manager-tdd, manager-ddd, expert-*, evaluator-active) as the execution substrate, not introduce new agents.

### REQ-HARNESS-005 (Event-Driven — moai update Migrator Extension)

**When** `moai update` runs against a project containing any of the 16 removed skills, the migrator **shall** copy each removed skill verbatim to `.moai/archive/skills/v2.16/<skill-id>/` (preserving directory structure, including `SKILL.md`, `modules/`, `examples.md`, `reference.md`), then remove the skill from `.claude/skills/`, then install `moai-meta-harness` from the embedded template. The migrator **shall** print a one-line status per archived skill and a final summary count.

### REQ-HARNESS-006 (Ubiquitous — Grace Window + Migration Guide)

The BC-V3R3-007 breaking change **shall** carry a 1 minor release grace window, meaning v2.17.x continues to recognize the archived skill paths in `.moai/archive/skills/v2.16/` as restorable via `moai migrate restore-skill <skill-id>`. The migration guide **shall** be authored at `.moai/release/MIGRATION-v2.17.0.md` and **shall** include: the verbatim 16-skill list, automatic migration steps, manual fallback steps, restore command syntax, and the deprecation timeline. CHANGELOG.md and RELEASE-NOTES-v2.17.0.md **shall** prominently announce BC-V3R3-007.

### REQ-HARNESS-007 (Ubiquitous — Archive Directory Structure)

The archive directory `.moai/archive/skills/v2.16/` **shall** preserve the full original skill content (frontmatter, body, modules, examples, reference) so that `moai migrate restore-skill <skill-id>` can recreate the original `.claude/skills/<skill-id>/` directory byte-for-byte (excluding ignored runtime artifacts). The archive layout **shall** be: `.moai/archive/skills/v2.16/<skill-id>/SKILL.md` plus any sibling files and subdirectories the original skill carried.

### REQ-HARNESS-008 (Unwanted — User Customization Preservation)

**If** the user has files under `.moai/harness/`, `.claude/agents/my-harness/`, or `.claude/skills/my-harness-*/` at the time of `moai update`, **then** the migrator **shall not** modify, delete, archive, or rename any of those files. User customizations are inviolate.

### REQ-HARNESS-009 (Event-Driven — Generated Harness Validation Hook)

**When** the meta-harness skill produces a new `my-harness-*` skill at runtime, the system **shall** invoke `evaluator-active` against the generated artifact using the harness Sprint Contract protocol (design constitution §11.5). The +60% effectiveness baseline from revfactory/harness A/B data **shall** be cited in the evaluator profile rationale; this REQ does not require runtime measurement, only that the evaluator profile references the +60% target as the design intent.

### REQ-HARNESS-010 (Ubiquitous — New Session Auto-Activation Mechanism)

The system **shall** ensure that when a new Claude Code session starts in a project that has been initialized with a dynamic harness (i.e., `.moai/harness/main.md` exists), the meta-harness skill itself does not block the recognition path. The actual 5-layer wiring (CLAUDE.md @import markers, workflow.yaml.harness section, frontmatter triggers, workflow static import line, `.moai/harness/` directory) is implemented in SPEC-V3R3-PROJECT-HARNESS-001; this SPEC's responsibility is to ensure the meta-harness skill's frontmatter declares `triggers.phases: ["plan", "run", "sync"]` and `triggers.keywords: ["harness", "project-init", "meta-skill"]` so that the skill can be auto-loaded when needed without explicit invocation.

---

## 7. Acceptance Coverage Map

| AC ID | Covers REQ-IDs |
|-------|----------------|
| AC-HARNESS-01 | REQ-HARNESS-001, REQ-HARNESS-004 |
| AC-HARNESS-02 | REQ-HARNESS-002 |
| AC-HARNESS-03 | REQ-HARNESS-003 |
| AC-HARNESS-04 | REQ-HARNESS-007, REQ-HARNESS-008 |
| AC-HARNESS-05 | REQ-HARNESS-005 |
| AC-HARNESS-06 | REQ-HARNESS-006 |
| AC-HARNESS-07 | REQ-HARNESS-009 |

Coverage: 10 REQs ↔ 7 ACs, 100% (every REQ appears in at least one AC). REQ-HARNESS-010 is verified indirectly through AC-HARNESS-01 (frontmatter inspection) and is fully validated under PROJECT-HARNESS-001 acceptance.

---

## 8. Constraints

- [HARD] All instruction files in English (per `.claude/rules/moai/development/coding-standards.md` Language Policy).
- [HARD] Template-First rule: every change to `.claude/skills/moai-meta-harness/` MUST be mirrored under `internal/template/templates/.claude/skills/moai-meta-harness/`.
- [HARD] FROZEN zone preservation: `.claude/rules/moai/design/constitution.md` §2 is NOT modified by this SPEC.
- [HARD] No file under `.claude/agents/moai/` is created, modified, or deleted by this SPEC.
- [HARD] EARS format mandatory for all REQs ("WHEN ... THEN ...", "IF ... THEN ...", "The system shall ...").
- [HARD] Conventional Commits with Korean body for all commits originating from this SPEC.
- Apache 2.0 attribution: `moai-meta-harness/SKILL.md` MUST include a top-of-file credit block referencing `https://github.com/revfactory/harness` and the Apache 2.0 license.
- 16-skill removal list is verbatim — adding or removing entries from §3 changes the SPEC ID.

---

## 9. Risks

| Risk | Mitigation |
|------|------------|
| User has heavily customized one of the 16 removed skills | Auto-archive preserves the original content under `.moai/archive/skills/v2.16/<id>/`; restore via `moai migrate restore-skill <id>`. Migration guide documents this path. |
| Confusion between static `moai-*` and dynamic `my-harness-*` | Doctor command warns on namespace violations. CLAUDE.md `<!-- moai:harness-start -->` markers (PROJECT-HARNESS-001) provide visual cues. |
| Apache 2.0 license compliance | Top-of-file attribution block in SKILL.md. Migration guide and CHANGELOG.md cite revfactory/harness. |
| moai update silently deletes user customizations | REQ-HARNESS-008 prohibits modification of any user-area file. CI test verifies preservation. |
| Grace window too short for production users | 1 minor release window is the minimum; users on a longer cadence can pin v2.16.x. Migration guide documents pin syntax. |
| revfactory/harness 7-Phase doesn't fit MoAI vocabulary | SKILL.md adapts terminology (e.g., harness Phase 1 "Discovery" → MoAI "manager-spec PLAN phase"). Adaptation table inline in SKILL.md. |
| evaluator-active runtime cost on every harness generation | evaluator-active runs once per harness generation (project init), not per command — amortized over project lifetime. |

---

## 10. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| SPEC-V3R3-HARNESS-LEARNING-001 | Hard prerequisite (depends_on) | LEARNING-001 self-evolution loop builds on the meta-harness skill this SPEC creates. LEARNING-001 was drafted first to lock the user-area boundary. |
| SPEC-V3R3-PROJECT-HARNESS-001 | Sibling (related_specs) | PROJECT-HARNESS-001 implements the Socratic interview + 5-layer wiring that drive `moai-meta-harness` at runtime. |
| SPEC-V3R3-DESIGN-PIPELINE-001 | Sibling (related_specs) | DESIGN-PIPELINE-001 references `moai-workflow-pencil-integration` removal (in this SPEC's 16-skill list). |
| SPEC-V3R3-PATTERNS-001 | Sibling (v2.16) | PATTERNS-001 absorbs revfactory/harness reference docs into `.claude/rules/moai/`; this SPEC absorbs the harness *workflow* into the meta-skill. |
| SPEC-AGENCY-ABSORB-001 | Reference | Establishes the agency-domain skill absorption pattern this SPEC reuses (FROZEN zones for `moai-domain-{copywriting,brand-design}`). |

---

## 11. Glossary

- **Meta-Harness**: A skill that generates other skills. The meta-harness skill (`moai-meta-harness`) is invoked by `/moai project` to produce project-specific `my-harness-*` skills via the 7-Phase workflow.
- **Static core**: The 22 base skills + 1 meta-harness skill (= 23) maintained by MoAI upstream and updated via `moai update`.
- **Dynamic harness**: The set of `.claude/agents/my-harness/*`, `.claude/skills/my-harness-*/*`, and `.moai/harness/*` artifacts generated per project. Inviolate to `moai update`.
- **BC-V3R3-007**: Breaking change identifier for the 16-skill removal. Carries 1 minor grace window.
- **Archive**: Read-only preservation of removed skills under `.moai/archive/skills/v2.16/<skill-id>/` for restore via `moai migrate restore-skill`.
- **revfactory/harness**: External Apache 2.0 project (`https://github.com/revfactory/harness`) that originated the 7-Phase meta-skill pattern. Cited per Apache 2.0 §4(c) attribution requirement.
- **Sprint Contract**: The evaluator-active negotiation protocol from design constitution §11.5, reused here to validate generated harness artifacts.
