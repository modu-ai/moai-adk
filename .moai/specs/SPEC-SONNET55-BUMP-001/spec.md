---
id: SPEC-SONNET55-BUMP-001
title: "Promote Claude Sonnet 5.5 (claude-sonnet-5-5) across the moai product — alias table, web labels, template guidance, README/docs-site"
version: "0.1.1"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/template
lifecycle: spec-anchored
tags: model-policy, sonnet-5-5, alias, glm-slot, templates, docs-site
era: V3R6
tier: M
related_specs: [SPEC-AGENT-MODEL-INHERIT-001]
issue_number: null
---

# SPEC-SONNET55-BUMP-001 — Sonnet 5.5 product-wide bump

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-29 | Initial draft. Card t1322, operator directive 2026-09-29 ("update sonnet-5 to sonnet-5.5, moai web and settings etc. all"). Anchors re-measured on this tree at `a62a05764`; three card anchors corrected (§A.3). |
| 0.1.1 | 2026-09-29 | Plan-audit iteration-1 fixes: lifecycle enum corrected (spec-anchored); tier-fit note (§A.4); AC-SSB-011/012 added (t1315 record, production-id grep). |

## §A Background

Claude Sonnet 5.5 (`claude-sonnet-5-5`) is announced as the successor to Sonnet 5: 30%+ faster,
up to 30% cheaper per task, same input/output price as Sonnet 5 ($2/$10 per Mtok; cache read
$0.20, cache write $2.50). Effort is supported low→max with Medium as the Claude Code/apps
default. First Sonnet with cyber safeguards, distillation safety classifiers, and expanded
preserved thinking. The announcement states NO context-window figure — that is a research item
(§E.3, D-5), not an assumption.

moai's model surface is alias-keyed: users pick `sonnet` (or `sonnet[1m]`), and a single SSOT
table resolves it to a canonical Claude Code model id. Because the alias does not change, most
of the product needs either a one-line table edit or nothing at all; the real work is (a) the
table contents, (b) user-facing generation labels ("Sonnet 5" prose), and (c) moai's own
guidance/docs that state which generation `sonnet` means.

### A.1 Where the canonical id actually lives (measured at `a62a05764`)

| Surface | Evidence |
|---|---|
| Alias SSOT | `internal/template/model_policy.go` `ModelAliasTable` (~:70): `"sonnet": "claude-sonnet-5"`. `@MX:ANCHOR` comment names its consumers: `launcher.go` `expandModelString` (:1142), `profile_setup.go` `normalizeModel` (:65), settings schema `modelOptions`. |
| Deprecated-id home | same file `ModelDeprecatedCanonicalIDs` (:82-90): carries `"claude-sonnet-4-6": "sonnet"` — the precedent this bump follows for `claude-sonnet-5`. |
| Named constants | `ModelIDOpus55` (:42), `ModelIDOpus48` (:48) exist; no `ModelIDSonnet*` constant exists. |
| Reverse lookup | `ModelAliasFromCanonicalID` (:112) checks `ModelAliasTable` first, then `ModelDeprecatedCanonicalIDs` — so after the bump both old and new canonical ids resolve to alias `sonnet` with no consumer change. |
| GLM slot mapping | `internal/template/glm_effort_overlay.go` `GLMSlotForModel` (:280) switches on the **alias** (`ModelAliasFromCanonicalID(base)`), not the canonical id. `sonnet → GLMSlotMedium` survives the bump unchanged. The card's item 4 ("update the claude-sonnet-5→Medium row") is **alias-keyed and needs no code edit** — verification only. |
| Web console | `internal/web/validate.go` model options flow from `settings.ModelOptionValues()` + `ModelAliasTable` keys (:52-69) — alias-driven, no code change. The real web surface is `assets/i18n.js` user-facing labels: `f.model.opt.sonnet: "Sonnet 5"` (:344-345) and generation prose (44 case-insensitive hits total). `internal/web/agentfm.go` does not exist; the agentfm model-options surface lives in `handlers.go` + `schemaform.go` and is alias-driven. |
| Wizard labels | `internal/cli/profile_setup_translations.go`: `"sonnet (Sonnet 5, balanced)"` (:174-175), effort-support prose (:181, :277, and ja/zh equivalents), `"Sonnet 5+"` plan gate prose (:192, :288). |
| Template guidance | 26 template files match `sonnet`; the moai-owned ones carrying the generation statement: `rules/moai/development/model-policy.md` ("sonnet = Sonnet 5 … native 1M"), `rules/moai/workflow/context-window-management.md` ("Sonnet 5 (1M)" row), `quality.yaml.tmpl` + `rules/moai/core/moai-constitution.md` + `context-window-management-detail.md` + `development/prompting-best-practices.md` + `agents/moai/super-advisor.md` ("available on Opus 5.5, Sonnet 5, …" effort lists), `development/skill-authoring.md`, `rules/moai/core/settings-management.md`, `rules/moai/core/moai-mcp-tools-catalogue.md`, `.moai/config/sections/{llm,workflow}.yaml` (alias + slot prose only). |
| Reference-only mirrors | `skills/moai-foundation-cc/reference/**` (5 files) and `skills/moai-foundation-core/**` carry `claude-sonnet-5` as verbatim copies of official Claude Code documentation — upstream-doc mirrors, not moai guidance (model-policy.md :53 says so explicitly). Excluded from edits. |
| README/docs-site | README.ko.md / README.md carry 3 sonnet mentions each (benchmark table row `sonnet-5 [max]` :242, prose :244, GLM slot table :711). docs-site `content/{en,ja,ko,zh}/` carry 116 files with sonnet mentions (~29 per locale). |
| Test literals | `internal/cli/glm_slot_effort_test.go:48`, `internal/cli/profile_setup_normalize_test.go:61`, `internal/template/glm_slot_test.go:30`, `internal/web/session_telemetry_cells_test.go:103,110` pin `"claude-sonnet-5"` as the *current* canonical; `internal/hook/served_model_test.go` / `served_model_stop_test.go` use it as an arbitrary fixture model string. `internal/cli/launcher_test.go` resolves via `template.ModelAliasCanonicalID("sonnet")` (:688, :692) — alias-driven, self-adjusts, no edit expected. |

### A.2 t1246 absorption already landed

Card t1246 (per-agent model+effort profile removal) merged into this base at `15560e663`.
`internal/template/profile_matrix.go` no longer exists — its remnant header lives in
`apply_harness.go` ("This file is what remains of the former profile_matrix.go") and carries no
sonnet literal. The card's item 2 surface therefore reduces to: `apply_harness.go` (verify-only),
`profile_setup_translations.go`, `profile_setup.go normalizeModel`, settings schema
`modelOptions` — all alias-driven except the translation strings.

### A.3 Card-anchor corrections (tree wins)

| Card claimed | Tree says |
|---|---|
| `profile_matrix.go` carries the profile matrix | Absorbed into `apply_harness.go` (t1246); no sonnet literal there. Verify-only. |
| GLM slot: "update the claude-sonnet-5→Medium row for claude-sonnet-5-5" | `GLMSlotForModel` is alias-keyed (`model_policy.go` reverse map feeds it); no row edit exists to make. Verification-only milestone item. |
| `internal/web/agentfm.go` is a sonnet surface | File does not exist; surface is `handlers.go` + `schemaform.go`, alias-driven. Verify-only. |

### A.4 Tier-fit note

The raw touch surface (code + templates + 4-locale docs) exceeds the typical Tier M file count,
but the decision count is small — one alias-table edit fans out mechanically — and the
requirement/AC volume (13 REQ / 12 AC) is within Tier M ceilings. Tier M is retained; the docs
fan-out in M4 is derivative prose, not independent design decisions.

## §B Requirements (GEARS)

- **REQ-SSB-001** (Ubiquitous) The alias table `ModelAliasTable` shall resolve the `sonnet` alias to the canonical id `claude-sonnet-5-5`.
- **REQ-SSB-002** (Ubiquitous) `ModelDeprecatedCanonicalIDs` shall map the superseded id `claude-sonnet-5` to the `sonnet` alias, so historical prefs files carrying the old canonical id keep resolving.
- **REQ-SSB-003** (Ubiquitous) The promoted entry shall reference a named constant `ModelIDSonnet55 = "claude-sonnet-5-5"` declared beside `ModelIDOpus55`, and no consumer file shall hardcode the literal id where the constant or the alias table is importable.
- **REQ-SSB-004** (Ubiquitous) The downstream surfaces — `launcher.go expandModelString`, `profile_setup.go normalizeModel`, settings schema `modelOptions`, `web/validate.go modelOptionList`, `glm_effort_overlay.go GLMSlotForModel` — shall resolve the promoted id through `ModelAliasTable` / `ModelAliasFromCanonicalID` without per-consumer literals.
- **REQ-SSB-005** (Event-driven) When the GLM slot resolver receives `claude-sonnet-5-5`, `claude-sonnet-5`, or `sonnet` (any `[1m]` variant included), it shall resolve the medium slot for all three.
- **REQ-SSB-006** (Ubiquitous) User-facing generation labels — `internal/web/assets/i18n.js` picker labels and generation prose, `internal/cli/profile_setup_translations.go` model/effort strings (all locales) — shall state Sonnet 5.5 as the current generation and shall not claim an effort-support ceiling below what Sonnet 5.5 supports (low→max).
- **REQ-SSB-007** (Ubiquitous) Template mirrors that state moai's own guidance about the current Sonnet generation (`model-policy.md`, `context-window-management.md` (+detail), `quality.yaml.tmpl`, `moai-constitution.md`, `prompting-best-practices.md`, `skill-authoring.md`, `settings-management.md`, `moai-mcp-tools-catalogue.md`, `agents/moai/super-advisor.md`, config-section prose) shall state Sonnet 5.5; verbatim official-doc mirrors under `skills/moai-foundation-cc/reference/**` and `skills/moai-foundation-core/**` shall not be edited by this SPEC.
- **REQ-SSB-008** (Ubiquitous) Template edits shall land template-first: `make build` shall succeed after every template edit batch, and `make agents-emit` shall run before build whenever a file under `templates/.claude/agents/moai/` changes (super-advisor.md is in scope).
- **REQ-SSB-009** (Ubiquitous) README (ko canonical + en/ja/zh derived) and docs-site pages (4 locales, same-PR parity) shall state Sonnet 5.5 wherever they state the current Sonnet generation; benchmark-table rows recording historical Sonnet 5 measurements shall keep their historical values and gain a generation label, not a rewritten number.
- **REQ-SSB-010** (Event-driven) When a test asserts the *current* canonical resolution of the `sonnet` alias, it shall assert `claude-sonnet-5-5`; tests using `claude-sonnet-5` as an arbitrary fixture string (`internal/hook/served_model_*`) and alias-driven tests (`launcher_test.go`) shall remain unchanged.
- **REQ-SSB-011** (Event-driven) When card t1315 (GitHub #1730, live resolution of `expandModelString`) merges after this SPEC, its absorber shall take this SPEC's table contents; when it merges before, this SPEC shall land on top of it unchanged in mechanism. Neither side shall rewrite the other's mechanism.
- **REQ-SSB-012** (Event-driven) When the official model docs (docs.anthropic.com models overview / pricing) state the Sonnet 5.5 context window, the research record shall cite that figure with its URL, and the statusline/context-window tables (`context-window-management.md`(+detail), docs-site tokenomics pages) shall be updated **only if** the figure differs from what they currently state for Sonnet; if the docs state no figure, the tables shall keep their Sonnet 5 row values and the research record shall say so.
- **REQ-SSB-013** (Ubiquitous) The Sonnet-5.5 thinking-off migration note (users running Sonnet with thinking off must switch to the `between_tools` thinking setting before moving to 5.5) shall be surfaced in docs at minimum: the `model-policy.md` template rule and the docs-site model-policy page (all 4 locales). No code or config surface reads or enforces this setting in this SPEC.

## §C Decisions

- **D-1 (constant)** Introduce `ModelIDSonnet55`; the alias table references it. Mirrors the `ModelIDOpus55` precedent; keeps the id greppable to one declaration.
- **D-2 (GLM slot is verification-only)** The slot mapping is alias-keyed and needs no edit. M1 carries a test that proves REQ-SSB-005 rather than a code change. If verification contradicts this (a literal `claude-sonnet-5` row is found in the effort overlay), the test change becomes a code change in the same milestone.
- **D-3 (template scoping)** Only files stating moai's own guidance change. The litmus test, from model-policy.md :53: files that "mirror official Claude Code documentation" stay byte-identical. `skills/moai-foundation-cc/reference/**` and `skills/moai-foundation-core/**` are excluded (AC-SSB-007).
- **D-4 (t1315 coordination)** Mechanism vs contents: t1315 changes how the table is read (compile-time snapshot → live resolution), this SPEC changes what the table says. Whichever lands second absorbs; no semantic conflict is expected. Recorded as REQ-SSB-011 and carried in plan.md §F.
- **D-5 (context window is research-gated)** Not assumed. M3 opens with the research step (research.md §2 procedure) and its output gates whether the context-window tables change. Default when undocumented: keep existing row values, record the citation gap.
- **D-6 (between_tools minimal surface)** Docs-only: template `model-policy.md` note + docs-site model-policy page. No runtime setting, no wizard field, no validation.

## §D Non-Functional Constraints

- Fixture-string principle: `claude-sonnet-5` occurrences that act as arbitrary test fixtures (served_model tests) stay unchanged — the SPEC changes behavior, not test scaffolding.
- Template neutrality CI (`.github/workflows/template-neutrality-check.yaml`) must stay green: no card provenance (t1322), no internal dates beyond what the edited prose already carries, no macOS-bias paths introduced.
- README 4-locale same-PR parity and docs-site 4-locale parity are HARD (hns-oss-docs rules): every generation-label edit lands in all four locales in the same change.
- Go code/comments in English; error wrapping via `%w`; no new hardcoded URLs (citation lives in docs prose, which is allowed).

## §E Research Items

- **E.1 Sonnet 5.5 pricing** — VERIFIED via the announcement (input $2 / output $10 / cache read $0.20 / cache write $2.50 per Mtok, same input/output as Sonnet 5). Used for docs prose only; no code reads pricing.
- **E.2 Effort support** — VERIFIED: low→max supported; Claude Code/apps default Medium. Matches the existing `GLMSlotMedium` wiring (D-2) — no change.
- **E.3 Context window** — NOT stated in the announcement. Research procedure in `research.md` §2; gates the context-window table decision (REQ-SSB-012, D-5).
- **E.4 `between_tools` migration note** — VERIFIED as an announcement migration note; surfaced per REQ-SSB-013.

## §F Out of Scope

### Out of Scope — verbatim official-doc mirrors
- `internal/template/templates/.claude/skills/moai-foundation-cc/reference/**` — upstream Claude Code documentation copies; they change when upstream changes, not when moai bumps.
- `internal/template/templates/.claude/skills/moai-foundation-core/**` — same class.

### Out of Scope — mechanism changes
- Converting `expandModelString` to live resolution (owned by card t1315 / GitHub #1730). This SPEC changes table contents only (REQ-SSB-011).
- Any change to `ModelAliasPickerValues`, the `[1m]` unification policy, or the `opusplan` routing alias.
- Any new runtime handling of the `between_tools` thinking setting beyond the docs note (REQ-SSB-013).

### Out of Scope — model surfaces unrelated to Sonnet
- Opus, Fable, Haiku rows in `ModelAliasTable`, `ModelDeprecatedCanonicalIDs`, and every prose mention of their generation labels — untouched unless a shared sentence lists them alongside Sonnet 5.5 (only the Sonnet token changes there).
- GLM model ids (`glm-5.3-flash` etc.) and the GLM slot model/effort config keys.
- Historical benchmark figures in README/docs-site (REQ-SSB-009 labels them; it does not rewrite them).

### Out of Scope — runtime behavior
- Statusline rendering logic, served-model gate/stop verdict logic (`internal/hook/served_model*.go` production files) — model-agnostic already; their test fixtures stay.
- `moai web` server routes and validation semantics (`validate.go` is alias-driven; zero expected diff beyond nothing).
