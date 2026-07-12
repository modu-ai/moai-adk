---
id: SPEC-ANALYZE-FIRST-ROUTING-001
title: "Analyze-First Routing Reform — language-independent intent analysis as default orchestration"
version: "0.1.0"
status: completed
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai, .claude/agents/moai, CLAUDE.md"
lifecycle: spec-anchored
tags: "agentic-core, routing, analyze-first, intent-analysis, agent-diet"
era: V3R6
tier: M
---

# SPEC-ANALYZE-FIRST-ROUTING-001 — Analyze-First Routing Reform

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-12 | manager-spec | Initial plan-phase authoring. Epic AGENTIC-CORE, SPEC 1 of 3 (entry SPEC; `GOAL-ENGINE-001` + `LOOP-SWEEP-001` depend on this). Findings consolidated in shared `research.md`. |

> **Epic**: AGENTIC-CORE (the frontmatter schema has no `epic:` field; recorded here in the body per the Epic naming taxonomy). This is SPEC 1 of 3. Downstream: `SPEC-GOAL-ENGINE-001` (depends_on this), `SPEC-LOOP-SWEEP-001` (depends_on GOAL-ENGINE).

## §A — Context and Motivation (WHY)

MoAI routing today is **English-biased and structurally gated**. The [HARD]
Intent Router lives inside the `/moai` skill (`.claude/skills/moai/SKILL.md` `:36`,
apply-instruction `:266`) and keys on English subcommand tokens (P1) and
English-only cue words (P3); `CLAUDE.md` §2 offers only a soft "Detect technology
keywords" hint for non-`/moai` natural-language requests (research.md §B.3-B.4).
Meanwhile 8 of 9 agents carry "MUST INVOKE when ANY of these keywords appear" +
EN/KO/JA/ZH keyword blocks (research.md §B.2) that (a) inflate always-loaded
context by an estimated ~5-6K tokens every session and (b) risk over-triggering
on Opus 4.7+/4.8 — directly contradicting the project's own authoring guidance at
`agent-authoring.md:215` ("aggressive language overtriggers tools/skills on the
latest models") and the fact that triggering is model-side SEMANTIC matching, not
literal keyword matching (research.md §B.1).

This SPEC makes **Analyze-First** the DEFAULT main-session orchestration behavior —
language-independent intent analysis for any input language, with or without
`/moai` — and diets the agent trigger blocks down to concise semantic scope prose.
It is docs/config only (no Go code). It is the Epic AGENTIC-CORE entry SPEC that
`SPEC-GOAL-ENGINE-001` and `SPEC-LOOP-SWEEP-001` build on.

### §A.1 Compatibility invariant (HARD)

`SPEC-HARNESS-EVOLVE-001` (research.md §D.2) attaches Phase −1/Ω routing-ledger
observation points to the CLAUDE.md §2 pipeline enumeration. This SPEC's §2
rewrite MUST PRESERVE an explicit ordered pipeline enumeration (the five stages
below) so those attachment points remain anchorable. The rewrite compacts the
prose; it does NOT delete the enumeration.

## §B — Scope (WHAT this SPEC delivers)

Five deliverables (D1-D5), all docs/config:

- **D1 — CLAUDE.md §2 rewrite**: Analyze-First pipeline as the default
  main-session behavior. Five ordered stages: ① intent analysis (any input
  language) → ② context-sufficiency check (existing Rule 5 AskUserQuestion) →
  ③ execution-plan composition (skills/agents/dynamic-workflow chain; Phase 0.95
  unchanged) → ④ approval gates (unchanged, incl. Implementation Kickoff Approval)
  → ⑤ execute → verify → iterate (goal evaluator from `SPEC-GOAL-ENGINE-001` when
  armed). Compact (~15 lines), remove the stale "Detect technology keywords"
  phrasing, point to the `/moai` skill for the structured router.
- **D2 — `/moai` SKILL.md Intent Router P3 language-independence clause**: add a
  clause stating cue words are English *exemplars*, not literal-match
  requirements, and that intent is classified for any `conversation_language`.
  P1 fast-path unchanged. Resolve the `lint` P3 collision (assign `lint` to `fix`;
  `gate` keeps "quality gate / pre-commit / check").
- **D3 — Agent description diet (all 9 `.claude/agents/moai/*.md`)**: replace
  "MUST INVOKE when ANY…" + 4-language keyword blocks with concise semantic scope
  prose + one line "Match user intent language-independently — do not require
  literal keyword matches." Keep the NOT-for (out-of-scope) clauses. Add
  equivalent trigger prose to `manager-design.md` (the one agent lacking it).
- **D4 — Hygiene**: fix `moai-foundation-core` `related-skills` drift (adopt the
  template value locally); update `skill-authoring.md` `triggers:` section to
  reflect semantic-matching reality (optional metadata, not a matcher);
  verify/fix SKILL.md `team/*.md` dead references; align `agent-authoring.md`
  examples (drop imperative-keyword style if present).
- **D5 — Template-First mirror**: every changed `.claude/` file mirrored in
  `internal/template/templates/.claude/` (§25 neutrality — no internal SPEC IDs
  in template bodies) + `make build` passes.

## §C — GEARS Requirements

> Notation: GEARS (current). `<subject>` is generalized. Each REQ maps to an AC in
> `acceptance.md`.

### §C.1 D1 — CLAUDE.md §2 Analyze-First default

- **REQ-AFR-001** (Ubiquitous): The `CLAUDE.md` §2 section shall define the
  Analyze-First pipeline as the default main-session orchestration behavior,
  enumerating the five ordered stages (intent analysis → context-sufficiency →
  execution-plan composition → approval gates → execute/verify/iterate) as an
  explicit ordered list.
- **REQ-AFR-002** (Ubiquitous): The `CLAUDE.md` §2 rewrite shall remove the stale
  "Detect technology keywords for agent matching" phrasing and replace it with a
  language-independent intent-analysis statement.
- **REQ-AFR-003** (State-driven): **While** the input is in any
  `conversation_language` (not only English), the §2 pipeline shall apply intent
  analysis without requiring an English `/moai` subcommand token.
- **REQ-AFR-004** (Ubiquitous): The §2 rewrite shall preserve an ordered pipeline
  enumeration usable as an attachment anchor by `SPEC-HARNESS-EVOLVE-001` Phase
  −1/Ω points, and shall not remove Phase 0.95 or the Implementation Kickoff
  Approval gate reference.

### §C.2 D2 — SKILL.md Intent Router language independence

- **REQ-AFR-005** (Ubiquitous): The `/moai` SKILL.md Intent Router shall carry a
  P3 clause stating cue words are English exemplars (not literal-match
  requirements) and that intent is classified for any `conversation_language`.
- **REQ-AFR-006** (Ubiquitous): The P1 first-word fast-path shall remain
  unchanged (English subcommand tokens continue to route directly).
- **REQ-AFR-007** (Unwanted behavior, event-detected): **When** the cue word
  `lint` is classified, the router shall route it to the `fix` bucket and shall
  not leave `lint` as a dual-membership cue shared with `gate` (the `gate` bucket
  retains "quality gate / pre-commit / check").

### §C.3 D3 — Agent description diet

- **REQ-AFR-008** (Ubiquitous): Each of the 9 `.claude/agents/moai/*.md` agent
  bodies shall express its delegation scope as concise semantic prose and shall
  not contain a "MUST INVOKE when ANY of these keywords appear" directive.
- **REQ-AFR-009** (Ubiquitous): Each of the 9 agent bodies shall contain the line
  "Match user intent language-independently — do not require literal keyword
  matches" (or a verbatim-equivalent single sentence) and shall retain its
  out-of-scope (NOT-for) clause.
- **REQ-AFR-010** (Ubiquitous): The `manager-design.md` agent shall gain concise
  trigger/scope prose equivalent to the other 8 agents (it currently has none).
- **REQ-AFR-011** (State-driven): **While** the diet is applied, the per-language
  (EN/KO/JA/ZH) keyword blocks shall be removed from every agent body, reducing
  the total trigger-block character count relative to the pre-diet baseline.

### §C.4 D4 — Hygiene

- **REQ-AFR-012** (Ubiquitous): The local `moai-foundation-core` SKILL.md
  `related-skills` value shall match the template value
  (`moai-foundation-cc, moai-foundation-thinking`), removing the
  `moai-foundation-context` (absorbed skill) drift.
- **REQ-AFR-013** (Ubiquitous): The `skill-authoring.md` `triggers:` section shall
  state that `triggers:` is optional metadata reflecting semantic matching, NOT a
  literal matcher.
- **REQ-AFR-014** (Event-detected): **When** the `/moai` SKILL.md references a
  `team/*.md` path that no longer resolves, that reference shall be removed or
  corrected (dead-reference cleanup).

### §C.5 D5 — Template-First mirror

- **REQ-AFR-015** (Ubiquitous): Every changed `.claude/` file shall have its
  corresponding `internal/template/templates/.claude/` mirror updated.
- **REQ-AFR-016** (Unwanted behavior): The template mirror bodies shall not
  contain any internal SPEC ID, REQ token, or audit citation (§25 neutrality).
- **REQ-AFR-017** (Event-driven): **When** the mirrors are updated, `make build`
  shall complete successfully (templates recompiled into the binary).

## §D — Exclusions (What NOT to Build)

[HARD] This SPEC explicitly does NOT deliver the following.

### §D.1 Out of Scope — Go routing code

- No changes to any Go source under `internal/` / `pkg/` / `cmd/`. Routing reform
  is docs/config only; the model-side semantic matcher is the runtime, not MoAI
  Go code (research.md §B.1).

### §D.2 Out of Scope — the `/moai goal` and `/moai loop` surfaces

- The `/moai goal` subcommand is owned by `SPEC-GOAL-ENGINE-001`. The `/moai loop`
  redefinition is owned by `SPEC-LOOP-SWEEP-001`. This SPEC only references the
  goal evaluator in the §2 pipeline stage ⑤ as a forward pointer; it does NOT
  build it.

### §D.3 Out of Scope — docs-site 4-locale translation

- Translating the reformed routing description into the docs-site
  `en/ja/zh/ko` locales is a DEFERRED follow-up (plan.md § Deferred), NOT
  run-phase scope.

### §D.4 Out of Scope — deep skill-catalog restructuring

- Retiring the deprecated `moai-meta-harness` skill, adding `triggers:` blocks to
  the ~30 skills that lack them, and any skill-body rewrite beyond the D4 hygiene
  items are out of scope (separate follow-up SPECs).

### §D.5 Out of Scope — CLAUDE.md HARD-rule set changes

- The §1 HARD Rules, the AskUserQuestion channel monopoly, and the Implementation
  Kickoff Approval gate are UNCHANGED. This SPEC rewrites only §2 (Request
  Processing Pipeline) prose.

## §E — Dependencies and Follow-ups

- **Blocks**: `SPEC-GOAL-ENGINE-001` (references the §2 pipeline stage ⑤ goal
  evaluator) and, transitively, `SPEC-LOOP-SWEEP-001`.
- **Compatibility**: `SPEC-HARNESS-EVOLVE-001` Phase −1/Ω attachment points
  (REQ-AFR-004 preserves the pipeline enumeration).
- **Doctrine anchors** (read, do not restate): shared `research.md`;
  `.claude/rules/moai/development/agent-authoring.md`;
  `.claude/rules/moai/development/skill-authoring.md`;
  `.claude/rules/moai/development/coding-standards.md` § Language Policy;
  `CLAUDE.local.md` §2 / §25.
