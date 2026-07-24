---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "moai init wizard restructure — 3-page topic layout + default recalibration (방안 A)"
version: "0.1.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: "internal/cli/wizard"
lifecycle: spec-anchored
tags: "cli, wizard, init, ux, model-policy, refactor"
tier: M
related_specs: [SPEC-V3R5-INIT-WIZARD-EXPANSION-001, SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001]
---

# SPEC-CLI-WIZARD-RESTRUCTURE-001 — moai init wizard restructure (방안 A)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | Initial plan-phase authoring. Encodes user-confirmed 방안 A: remove the advanced-settings gate, reorganize the `moai init` wizard into 3 topic-based pages shown to every user, recalibrate `model_policy` default High→Medium and `lsp_enabled` default false→true, and remove the `harness_profile` / `coverage_exemptions_enabled` questions (fix them at defaults). |

## §A — Overview

The `moai init` interactive wizard currently gates its richer settings behind a Quick-mode `advanced_bridge` confirm question plus `--standard` / `--advanced` flag modes and a reflection-based `advanced_gate.go` Phase-2 readiness stub. Most users answer "No" to the bridge and never see project mode, LSP, quality gates, or design settings. 방안 A (user-confirmed) removes the gate and presents a single, always-shown, three-page topic layout so every user makes the same informed choices, and recalibrates two defaults toward the more common rational choice.

This SPEC defines WHAT the restructured wizard observably does. The HOW (huh group assembly, exact edit sites) lives in `plan.md`.

## §B — Requirements (GEARS)

### §B.1 — Gate removal & 3-page structure

- **REQ-WIZ-001** (Ubiquitous): The init wizard shall present its questions as exactly three topic-based pages — **Basic**, **Model & Report**, **Quality & Workflow** — shown to every user regardless of any mode flag.
- **REQ-WIZ-002** (Unwanted): The init wizard shall not present an advanced-settings bridge question, and shall not require any confirmation step to reveal the Quality & Workflow settings.
- **REQ-WIZ-003** (Ubiquitous): Page 1 (Basic) shall contain, in order, the conversation-language, user-name, and project-name questions.
- **REQ-WIZ-004** (Ubiquitous): Page 2 (Model & Report) shall contain, in order, the model-policy and report-format questions.
- **REQ-WIZ-005** (Ubiquitous): Page 3 (Quality & Workflow) shall contain the LSP, quality-gate, project-mode, design, and Claude-Design questions.
- **REQ-WIZ-006** (State-driven): While the design workflow is disabled, the wizard shall hide the Claude-Design integration question; while it is enabled, the wizard shall show it.
- **REQ-WIZ-007** (Event-driven): When the user changes the conversation language on Page 1, the wizard shall render every subsequently displayed question title and description in the newly selected language.

### §B.2 — Default recalibration

- **REQ-WIZ-008** (Ubiquitous): The model-policy question shall pre-select the **Medium** tier as its default and carry the recommended marker on the Medium option (not the Max option).
- **REQ-WIZ-009** (Ubiquitous): The project-wide default model policy constant shall resolve to Medium, and every doc-comment or prose reference stating the default is "high" shall state "medium" instead.
- **REQ-WIZ-010** (Ubiquitous): The LSP-integration question shall default to enabled, and its title and description shall reflect the enabled-by-default state in all four supported locales (en/ko/ja/zh).
- **REQ-WIZ-011** (Unwanted): The wizard shall not retain "high" (ModelPolicyHigh) as the model-policy *default seed* anywhere, while preserving the Max/High tier as a user-selectable option.

### §B.3 — Question removal & fixed defaults

- **REQ-WIZ-012** (Ubiquitous): The wizard shall not present the harness-profile question, and the effective default harness profile shall resolve to `default`.
- **REQ-WIZ-013** (Ubiquitous): The wizard shall not present the coverage-exemptions question, and the effective coverage-exemptions setting shall resolve to `false`.
- **REQ-WIZ-014** (Ubiquitous): For each removed question, the wizard shall remove the corresponding ko/ja/zh translation entries so the translation-completeness regression test passes.

### §B.4 — Answer application & preservation

- **REQ-WIZ-015** (Ubiquitous): The init flow shall persist every Page 3 answer (LSP, quality-gate, project-mode, design, Claude-Design) to the project configuration without gating that application on a standard/advanced mode flag.
- **REQ-WIZ-016** (Ubiquitous): The reconfigure question set (used by `moai update --reconfigure`) shall preserve the Git question set and their ordering relative to the report-format question.
- **REQ-WIZ-017** (Unwanted): The wizard shall not leave orphaned answer-capture branches that no longer correspond to any presented question.

## §C — Exclusions

The following are explicitly out of scope for this SPEC. Each is routed to its correct home rather than expanded here.

### Out of Scope — new configuration fields

- No new config keys, no new `WizardResult` fields, and no new persisted settings are introduced. This SPEC only reorganizes and recalibrates the existing question set.

### Out of Scope — Git question redesign

- The 7 Git questions (`git_mode`, `git_provider`, tokens, usernames) and the remote-auto-detection path are untouched. Only their ordering-preservation in the reconfigure splice is asserted (REQ-WIZ-016).

### Out of Scope — model-routing matrix / profile persistence

- The per-agent profile matrix, `llm.profile` / `performance_tier` persistence, and the tier→effort/tier mapping are unchanged. Only the *default selection* of the model-policy tier moves High→Medium (REQ-WIZ-008/009). The Max, Medium, and Low tiers remain fully selectable.

### Out of Scope — advanced_gate / Phase 2 retirement decision

- Whether to fully retire `advanced_gate.go`, the `--standard` / `--advanced` flags, and the inert `Phase2Questions` stubs is deferred to a kickoff clarification (see `plan.md` §B `[NEEDS CLARIFICATION]`). This SPEC's core deliverable (3-page layout + default recalibration + question removal) does not depend on that decision.

### Out of Scope — wizard rendering engine / theme

- The huh v2 theme, styling tokens, and stepper mechanics (`styles.go`, `wizard.go` theme functions) are untouched except for the group-assembly changes needed to form three pages.
