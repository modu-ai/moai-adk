---
id: SPEC-V3R5-INIT-WIZARD-EXPANSION-001
version: "0.1.0"
status: implemented
created_at: 2026-05-22
updated_at: 2026-06-02
author: manager-spec
priority: High
labels: [wizard, init, profile, ux, configuration]
tier: M
issue_number: null
related_specs: [SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001, SPEC-V3R5-GIT-STRATEGY-SCHEMA-001, SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001]
---

# SPEC-V3R5-INIT-WIZARD-EXPANSION-001 — `moai init` Wizard Decision-Point Expansion

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — Tier M, 3-artifact LEAN. v2 audit applied to 8 candidates B1-B8. Phase 1 (5 candidates) standalone; Phase 2 (3 candidates) deferred to post-P2/P4 run-phase. |

## 1. Background

The 2026-05-22 audit established that the current `moai init` wizard captures only **9 of 30+ yaml decision points** (~10% exposure ratio). Sister SPEC `SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001` (implemented v0.2.0) extended `moai profile setup` with statusline preset + segments. This SPEC extends the `moai init` interactive wizard to expose 5 additional project-level decisions that today require manual yaml editing.

### 1.1 v2 Audit Disposition Matrix (binding)

The 8 candidates from the seed analysis were verified via Step 1 grep + Step 2 SPEC catalog cross-check before scope inclusion:

| # | Key | Default | Consumer (verified) | Owner SPEC | Phase | Disposition |
|---|-----|---------|--------------------|-----------|------|-------------|
| B1 | `project.mode` (`personal`/`team`) | `personal` | `internal/config/types.go:82` (`yaml:"mode"`) + `initializer.go:393` writes "mode: personal" | SPEC-CONFIG-001 (config substrate) | 1 | INCLUDE |
| B2 | `harness.default_profile` | `default` | `internal/cli/harness_validate.go:107` + `internal/cli/harness_route.go:170` + 4 profiles in `.moai/config/evaluator-profiles/` (default/strict/lenient/frontend) | SPEC-V3R2-HRN-002 (harness substrate) | 1 | INCLUDE — dynamic enumeration from directory |
| B3 | `lsp.enabled` | `false` (opt-in) | `internal/lsp/gopls/config.go:72` + `pkg/models/config.go:76` + template `lsp.yaml.tmpl:45` master switch | SPEC-LSP-CORE-002 + SPEC-GOPLS-BRIDGE-001 | 1 | INCLUDE |
| B4 | `workflow.team.enabled` + `workflow.team.default_model` | `false`/`sonnet` | Template `workflow.yaml:27 default_model: sonnet` + 14 prose references in rules + skills (struct consumer pending P4) | SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 (P4, status:draft) | **2 (DEFER)** | DEFER — depends on P4 run-phase nested struct |
| B5 | `quality.enforce_quality` + `quality.coverage_exemptions.enabled` | `true`/`false` | `internal/core/quality/trust.go:200,303` + `internal/config/defaults.go:178-209` + `initializer.go:336` writes enforce_quality | SPEC-V3R2-QUALITY-001 (TRUST 5 substrate) | 1 | INCLUDE |
| B6 | `git-strategy.<mode>.branch_creation.{auto_enabled, prompt_always}` | `false`/`true` | Template `git-strategy.yaml` nested keys (struct consumer pending P2) | SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 (P2, status:draft) | **2 (DEFER)** | DEFER — depends on P2 run-phase nested struct |
| B7 | `git-strategy.<mode>.commit_style.{format, scope_required}` | `conventional`/`false` | Template nested keys (struct consumer pending P2) | SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 (P2, status:draft) | **2 (DEFER)** | DEFER — depends on P2 run-phase nested struct |
| B8 | `design.enabled` + `design.claude_design.enabled` | `true`/`true` | `internal/config/types.go:684,717` (`Design.ClaudeDesign`) + `internal/config/defaults.go:475` (`NewDefaultDesign`) + template `design.yaml:43` | SPEC-DESIGN-001 (design substrate) | 1 | INCLUDE |

**Phase 1 scope (this SPEC run-phase)**: B1, B2, B3, B5, B8 — 5 candidates, no external SPEC dependency.

**Phase 2 scope (deferred)**: B4, B6, B7 — 3 candidates blocked on P2 (`SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`) and P4 (`SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001`) run-phase completion. Plan.md milestones M5-M6 reserve scaffolding but cannot ship until prerequisites complete.

### 1.2 UX Cognitive-Load Mitigation

Adding 5 questions to the current 9-question wizard would push median completion time past empirical comfort threshold. The SPEC introduces a **3-tier mode flag** allowing graduated opt-in:

- **Quick mode** (default, no flag): existing 9 questions only — backward-compat preserved
- **Standard mode** (`moai init --standard`): Quick + Phase 1 candidates (5 questions added = 14 total)
- **Advanced mode** (`moai init --advanced`): Standard + Phase 2 candidates when ready (8 total added = 17 total)

Quick mode keeps the no-friction default. Standard/Advanced are opt-in for users who want full control without manual yaml editing.

## 2. Requirements (EARS)

### REQ-IWE-001 (Ubiquitous, project.mode)
The `moai init` wizard **shall** include a Select question for `project.mode` with options `personal` (recommended) and `team`, persisting the choice to `.moai/config/sections/project.yaml` under key `project.mode`.

### REQ-IWE-002 (Ubiquitous, harness.default_profile)
The `moai init` wizard **shall** include a Select question for `harness.default_profile` dynamically enumerated from `.moai/config/evaluator-profiles/*.md` filenames (currently `default`, `strict`, `lenient`, `frontend`), persisting the choice to `.moai/config/sections/harness.yaml` under key `harness.default_profile`.

### REQ-IWE-003 (Ubiquitous, lsp.enabled)
The `moai init` wizard **shall** include a Confirm question for `lsp.enabled` (default `false`, opt-in to `true`), persisting the choice to the rendered `.moai/config/sections/lsp.yaml` under key `lsp.enabled`.

### REQ-IWE-004 (Ubiquitous, quality gates)
The `moai init` wizard **shall** include two Confirm questions for `quality.enforce_quality` (default `true`) and `quality.coverage_exemptions.enabled` (default `false`), persisting choices to `.moai/config/sections/quality.yaml` under the existing `constitution:` block.

### REQ-IWE-005 (Ubiquitous, design opt-in)
The `moai init` wizard **shall** include two Confirm questions for `design.enabled` (default `true`) and `design.claude_design.enabled` (default `true`), persisting choices to `.moai/config/sections/design.yaml`.

### REQ-IWE-006 (Event-driven, mode flag)
**When** `moai init` is invoked with `--standard`, the wizard **shall** present Phase 1 candidates (REQ-IWE-001 through REQ-IWE-005) in addition to the existing 9 questions. **When** invoked without `--standard` (and without `--advanced`), Phase 1 questions **shall** be skipped and yaml defaults retained (backward-compatible Quick mode).

### REQ-IWE-007 (Event-driven, advanced flag)
**When** `moai init` is invoked with `--advanced`, the wizard **shall** present both Phase 1 and Phase 2 candidates. If P2 (`SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`) or P4 (`SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001`) run-phases are not yet complete (no nested struct in source), the wizard **shall** skip the corresponding Phase 2 questions and emit a stderr warning naming the missing dependency.

### REQ-IWE-008 (State-driven, non-interactive)
**While** `--non-interactive` is set, the wizard **shall** skip all interactive prompts (Phase 1 and Phase 2) and write yaml defaults unchanged. CLI flags `--enforce-quality`, `--enable-lsp`, `--harness-profile <name>`, `--project-mode <personal|team>`, `--enable-design` **shall** override defaults non-interactively.

### REQ-IWE-009 (Optional, grouping)
Where the wizard uses `huh.NewForm` groups for visual organization, Phase 1 questions **shall** be grouped into a single "Project Defaults" group displayed after the existing Git settings group, with each question prefixed by a step indicator consistent with the current `tui.Stepper(current, total)` pattern.

### REQ-IWE-010 (Ubiquitous, persistence guarantee)
The wizard **shall** persist user selections to yaml files exclusively through `internal/core/project/initializer.go` write paths, never through direct stdout redirection or shell pipes, ensuring all writes are testable and auditable.

### REQ-IWE-011 (Ubiquitous, test coverage)
Every new Question definition and every new yaml-write code path **shall** have unit-test coverage ≥85% in `internal/cli/wizard/expansion_test.go` and `internal/core/project/initializer_expansion_test.go` per project policy (`.moai/config/sections/quality.yaml`).

## 3. Out of Scope

### 3.1 Out of Scope

- Modification of the existing 9 wizard questions or their order (Quick mode is frozen for backward-compat)
- Modification of `moai profile setup` 12 fields (separately governed by SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001)
- Phase 2 candidates B4, B6, B7 implementation (deferred to post-P2/P4 run-phase; this SPEC may scaffold but not ship)
- New evaluator-profile creation (B2 only exposes selection from the existing 4 profiles — adding a new profile is a separate concern)
- LSP server installation guidance beyond the boolean toggle (B3 only exposes `lsp.enabled`; per-language server activation remains in `lsp.yaml.tmpl`)
- Migration of pre-existing projects (`moai update` path) — wizard expansion only affects `moai init` scaffold path. Existing projects retain their yaml unchanged.
- Localization of new question titles/descriptions into ko/ja/zh in this SPEC — only English strings ship; localization is a follow-up SPEC
- Web-UI or TUI redesign — wizard remains terminal-only via existing `huh`/`bubbletea` stack

## 4. Glossary

- **Quick mode**: default `moai init` behavior with 9 questions, no opt-in flags required
- **Standard mode**: `moai init --standard` with Quick + Phase 1 = 14 questions
- **Advanced mode**: `moai init --advanced` with Standard + Phase 2 (when prerequisites complete) = 17 questions
- **Phase 1 candidates**: B1, B2, B3, B5, B8 — no external SPEC dependency
- **Phase 2 candidates**: B4, B6, B7 — blocked on P2/P4 run-phase
- **Wizard exposure ratio**: fraction of yaml decision points captured interactively; current ~10%, post-Phase 1 ~26%, post-Phase 2 ~37%
