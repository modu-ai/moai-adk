---
id: SPEC-MODEL-OPUS55-001
title: "Opus 5.5 adoption — opus alias to claude-opus-5-5, Opus 5 retired from the catalog, medium-effort default recommendation"
version: "0.1.2"
status: completed
created: 2026-09-23
updated: 2026-09-23
author: manager-spec (card t1089)
priority: P1
phase: "v3.2.0"
module: "internal/template,internal/cli,internal/cli/wizard,internal/web,internal/template/templates/.claude/rules,.claude/rules,.moai/project"
lifecycle: spec-anchored
tier: M
tags: "model-policy,opus-5-5,effort,profile-wizard,web-console,template"
related_specs: [SPEC-OPUS47-COMPAT-001]
amendment_of: SPEC-MODEL-OPUS55-001
---

# SPEC-MODEL-OPUS55-001 — Opus 5.5 adoption

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-23 | manager-spec (card t1089) | Initial plan-phase draft. Dispatched ID `SPEC-MODEL-OPUS-5-5-001` failed the SPEC-ID regex (middle segment `5` must start with a letter); canonical correction `SPEC-MODEL-OPUS55-001` (regex PASS, no collision). |
| 0.1.1 | 2026-09-23 | manager-spec (card t1089) | Plan-audit iter-1 FAIL 0.78 repair (report `.moai/reports/t1089/plan-audit-iter1.md`), scoped to defects D1–D15: REQ-OP55-007 rewritten to match the launch-effort resolution order (D1); high/xhigh default-effort conflict inventory added, REQ-OP55-010 extended to every false default statement incl. version-agnostic ones (D2); English-unified web model label kept (D3); positive + negative effort ACs (D4); AC-009 split per guard + alias-change mutant (D5, D14); `.moai/project/{tech,product}.md` brought into scope (D6); REQ-OP55-009 pinned to one canonical fact line (D7); launch `--model` change disclosed as K6 (D8); minor fixes D9–D13, D15. |
| 0.1.2 | 2026-09-23 | manager-spec (card t1089) | In-place amendment (see § Amendments). |

## Amendments

| Field | Value |
|-------|-------|
| prior completed version | 0.1.1 |
| prior_completed_sha | d0037fba8 |
| rationale | Sync-audit F2 (`.moai/reports/t1089/sync-audit.md`): §C said "Nothing else changes at runtime" and plan.md §E K4 said "Not changed here", but run commit `eb629efb5` changed how a resolved `max` effort reaches Claude Code. The SPEC body contradicted shipped behavior and no REQ/AC covered it. |
| scope | §C runtime-change bullet; REQ-OP55-007 gains the effort-delivery clause (REQ-OP55-007 (b)); plan.md §E K4 row; acceptance AC-OP55-007a/b + trace row. REQ/AC counts stay 16/16 (folded as sub-criteria). Operator-precedence wording is written for the correct behavior (argv anywhere, before or after `--`, space or `=` form); the run-phase F1 fix delivers it. Nothing else changed. |

## §A Context

### A.1 Problem

Claude Opus 5.5 was released 2026-09-22 and Claude Code's `opus` alias now resolves to it (Claude Code v2.1.280+). moai-adk still names Opus 5 (`claude-opus-5`) as the current Opus model in code, in every user-facing label (TUI wizard 4 locales, web console 4 locales), and in the template + local rule prose. The rule prose also states "Opus 5 defaults to `effort: high`" and "use `xhigh` for coding", which no longer describes the current default model: Opus 5.5 defaults to `medium`.

### A.2 Official basis (re-fetched by manager-spec 2026-09-23, not carried from dispatch)

- `https://platform.claude.com/docs/en/models/opus-5-5/overview`: model ID `claude-opus-5-5` (no date suffix); 1M context; 128K max output; $4 / $20 per MTok; thinking "Adaptive (always on)" and cannot be turned off; default effort `medium`; released September 22, 2026. Breaking changes vs Opus 5: thinking cannot be disabled, forced tool use returns an error, thinking blocks tied to the producing model, `computer_20251124` not accepted.
- `https://code.claude.com/docs/en/model-config`: "Opus 5.5 requires Claude Code v2.1.280 or later"; the `opus` alias resolves to Opus 5.5 on the Anthropic API; default effort is `high` on every effort-capable model "except that Opus 5.5 defaults to `medium`"; a top-level `effortLevel` in project/local/managed settings or one passed with `--settings` "applies to every model", while a top-level `effortLevel` in the USER settings file does not count for Opus 5.5 (per-model `modelSettings` supersedes it); `claude-opus-5` remains selectable by full model name.

### A.3 Measured gap (see plan.md §B for commands + verbatim output)

- Code: `ModelIDOpus5 = "claude-opus-5"` is the `opus` alias target; `claude-opus-5` is not in the deprecated-id map.
- Current-behavior surfaces carrying an unattributed Opus 5 reference: 153 lines (probe P4 v2, plan.md §B.4) across Go production files, template rules/skills/config, local rules/skills/config, `.moai/project/{tech,product}.md`, web i18n, and three wizard golden files.
- Version-agnostic "default effort is `high`" statements: 11 hits (probe P6, plan.md §B.7), invisible to P4 because some name no Opus version.
- Web console: model and effort already have edit-and-save paths (select fields persisted to the profile store). The measured gap is the recommendation display: the effort empty option reads "(runtime default)" and no option is marked recommended; the model option labels read "Opus 5".

## §B Requirements (GEARS)

### B.1 Model identity

- **REQ-OP55-001** — The `opus` alias shall resolve to the canonical model id `claude-opus-5-5`, and the constant carrying that id shall be named for Opus 5.5.
- **REQ-OP55-002** — When a stored profile or preference value carries the full id `claude-opus-5`, the model-alias normalization shall resolve it to the `opus` alias, exactly as it already does for `claude-opus-4-6`, `claude-opus-4-7`, and `claude-opus-4-8`.
- **REQ-OP55-003** — The codebase shall not keep a named constant for `claude-opus-5`; the id survives only as a superseded row of the deprecated-id map.

### B.2 User-facing labels and recommendation

- **REQ-OP55-004** — The TUI profile wizard (en/ko/ja/zh), the model-policy question descriptions (en/ko/ja/zh), and the web console model option labels (en/ko/ja/zh) shall name Opus 5.5 wherever they name the model the `opus` alias resolves to.
- **REQ-OP55-005** — The TUI profile wizard and the web console settings screen shall mark `medium` as the recommended session effort level in all four locales.
- **REQ-OP55-006** — The TUI profile wizard and the web console settings screen shall mark the Opus 5.5 model option as the recommended model in all four locales.
- **REQ-OP55-007** — The web console effort field's empty option shall state the launch resolution order as it actually runs: the profile's model-policy effort when a model policy is stored, otherwise Claude Code's own model default (`medium` on Opus 5.5); the label shall not claim `medium` unconditionally, and the launch-effort resolution itself shall not change. (b) When the resolved launch effort is `max`, the launcher shall pass it as the `--effort max` launch argument and shall not write `max` into the injected `--settings` `effortLevel`; `low`/`medium`/`high`/`xhigh` shall keep travelling as the settings `effortLevel`; `CLAUDE_CODE_EFFORT_LEVEL` shall not be set; and when the operator's argv already carries `--effort` anywhere — before or after a `--` separator, in the space or `=` form — the launcher shall inject no `--effort` of its own on either the general or the kanban path.
- **REQ-OP55-008** — When the model the `opus` alias resolves to changes, the label-drift guards shall fail until every label names the new dotted marketing version, and shall not accept a label that names a bare "Opus 5" as satisfying "Opus 5.5".

### B.3 Rule prose (template + local pairs)

- **REQ-OP55-009** — The canonical fact line for the current Opus model — the `- opus = …` line under "Current model generation mapping" in `model-policy.md`, both copies — shall state Opus 5.5, `claude-opus-5-5`, Claude Code minimum v2.1.280, 1M context, 128K output, $4 / $20 per MTok, default effort `medium`, and adaptive thinking always on; every other rule, skill, config, and project-doc mention of the current Opus model shall name Opus 5.5 and shall not contradict that line.
- **REQ-OP55-010** — Every statement that the current model's default effort is `high` — including version-agnostic ones such as "`high` (default)" — shall be rewritten so that Opus 5.5 defaults to `medium` and other effort-capable models keep their `high` default; the full list, with keep/rewrite decisions for every high/xhigh effort statement, is plan.md §C.6.
- **REQ-OP55-011** — When the constitution's prompt-philosophy heading is renamed, both zone-registry copies (template + local) shall carry the new anchor for CONST-V3R2-028 and CONST-V3R2-029, and no registry anchor shall point at a heading that no longer exists.
- **REQ-OP55-012** — Where a benchmark or measurement was taken on Opus 5, the prose shall keep that attribution explicitly (the literal phrase "measured on Opus 5") rather than relabel the measurement as Opus 5.5.

### B.4 Preserved behavior (unwanted-behavior guards)

- **REQ-OP55-013** — The template `settings.json.tmpl` shall not carry an `effortLevel` key (or any effort key).
- **REQ-OP55-014** — The per-agent profile matrix cells (model + effort per agent per profile column) shall not change.
- **REQ-OP55-015** — Historical records shall not be modified: `CHANGELOG.md`, other SPEC directories under `.moai/specs/`, pre-existing files under `.moai/reports/` (this card's own evidence directory `.moai/reports/t1089/` excepted), `.moai/research/`, `.moai/docs/`, `docs-site/`, and the four `README*.md` files.
- **REQ-OP55-016** — The template edits shall carry no forbidden internal-content class (SPEC IDs, REQ tokens, internal dates, commit SHAs, CLAUDE.local references) and shall keep the 16 programming languages neutral; the embedded template shall be rebuilt after template edits.

## §C Exclusions

### Out of Scope — docs-site and READMEs (split to a separate card)

- The 49 `docs-site/content/**` files that mention Opus 5 (4 locales, same-PR obligation, Vercel deploy) move to a separate oss-docs card.
- The four `README*.md` files: their Opus 5 benchmark tables are measured facts on Opus 5 and stay unchanged; any prose update belongs to the same separate docs card.
- `.moai/docs/mcp-recipes.md` (one Opus 5 mention) goes with that docs card.

### Out of Scope — runtime changes other than the alias target

- This SPEC makes two runtime changes. (1) The `opus` alias target: a profile model of `opus[1m]` launches as `--model claude-opus-5-5[1m]`, which needs Claude Code v2.1.280+ (plan.md §E K6). (2) Delivery of a resolved `max` effort (REQ-OP55-007 (b), plan.md §E K4): Claude Code does not accept `max` as a settings `effortLevel` ("`max` isn't accepted as a level in either key", code.claude.com/docs/en/model-config), so `max` is passed as the session-scoped `--effort max` launch argument on the general and kanban paths, while the other four levels stay on the settings path; an operator-supplied `--effort` anywhere in argv wins. Nothing else changes at runtime.
- No change to the launch effort fallback (explicit profile effort wins, else the model-policy-derived effort, else no override).
- No re-derivation of the per-agent profile matrix cells (operator-settled; the matrix comment forbids re-derivation).
- No `effortLevel` injection into any template settings file, and no `modelSettings` support.
- No handling of the Opus 5.5 API breaking changes (forced tool use, thinking disable, `computer_20251124`): no moai code path sends Anthropic API requests with those parameters; the GLM `tool_choice` path in `internal/cli/mcp_glm.go` targets z.ai and is unaffected.

### Out of Scope — historical and test-fixture surfaces

- Test fixtures that use `claude-opus-5` as an arbitrary model string in token, telemetry, and codex-servability tests stay as they are (Claude Code still accepts the id).
- Completed SPECs, CHANGELOG history, and reports keep their Opus 5 wording.

## §D Constraints

- Template-First: every template edit precedes its local copy; `make build` after template edits; `make agents-emit` only if a template agent file changes (none planned).
- Local/template pairs that are byte-identical today stay byte-identical (measured list in plan.md §B.5).
- `internal/cli` full-package test runs take a slot lease and `-timeout 25m` (lane protocol §8).
