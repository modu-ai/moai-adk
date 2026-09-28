---
id: SPEC-AGENT-MODEL-INHERIT-DOCS-001
title: "docs-site model/effort docs rewrite to the inheritance narrative + wizard H24 residue (card t1300)"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site, internal/cli/wizard"
lifecycle: spec-anchored
tags: "docs-site, model-inheritance, i18n, profile-matrix-retirement, wizard-translations, h24, card-t1300"
tier: M
era: V3R6
related_specs: [SPEC-AGENT-MODEL-INHERIT-001, SPEC-AGENT-MODEL-ENFORCE-001, SPEC-ROLE-NAMING-DOCS-001, SPEC-MODEL-MATRIX-DOCS-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-29 | manager-spec | Initial draft for card t1300 (Tier M, Class C, split from t1246 M8). Scope measured in `.moai/reports/t1300/phase1-survey.md` (155 pages × 4 locales = 620 surveyed; A/B/C classification with en line-number edit maps) and re-verified on tree `8a969dfc0` at plan authoring. Baseline RED-now counts recorded in acceptance.md. |

---

## 1. Overview

SPEC-AGENT-MODEL-INHERIT-001 (t1246, landed) removed per-agent model/effort assignment from the
product: subagents inherit the main session's model and effort, MoAI agent definitions declare
neither, and the `moai model profile` accessor is deleted. The docs layer still teaches the retired
per-agent profile-matrix narrative: the docs-site A-cluster pages (5 pages × 4 locales), the
getting-started flag docs, the prompt-caching caveats, and the init-wizard UI strings (t1246 design
H24 residue) all predate the change.

This SPEC is the docs-layer sweep. Its canonical wording source is the landed doctrine sentence at
`.claude/rules/moai/core/agent-common-protocol.md` § "Subagent Model and Effort":

> Subagents inherit the main session's model and effort: pass neither `model` nor `effort` when
> spawning a subagent, and MoAI agent definitions declare neither.

All rewrites MUST derive from this wording and from the actual measured CLI surface — not invent a
parallel narrative.

## 2. Measured Baseline (tree `8a969dfc0`, branch `WT-model-docs-sweep`)

- `internal/cli/model.go` is deleted — the `moai model profile` accessor no longer exists.
- `internal/cli/init.go:105-109` registers `--model-policy`, `--high`, `--medium-alias`, `--low`,
  `--profile` with `deprecatedAgentModelFlagUsage`; `warnDeprecatedAgentModelFlags` (init.go:355)
  emits the deprecation warning pointing at `moai profile setup`. The flags survive only as stubs.
- `docs-site/content/en/cli-reference/profile.md` documents the SURVIVING `moai profile` command
  family (`profile list` / `profile setup` / `profile current`, CLAUDE_CONFIG_DIR isolation) — it
  does NOT document the dead accessor. Measured directly; the phase-① survey's uncertainty about
  this page is resolved by this measurement.
- `internal/cli/wizard/translations.go` carries 5 `model_policy` occurrences (grep count 5).
  `internal/cli/profile_setup_translations.go` is already clean (only line 165
  `"Session model policy"` — main-session wording landed by t1302).
- A-cluster RED-now hits (en): `multi-llm/model-policy.md` 7 hits for the retired markers
  ("The three profiles" / "Per-agent assignment table" / "moai model profile");
  `advanced/profile-matrix.md` 3 hits. The prompt-caching page carries 1 per-spawn-injection
  bullet (en cost-optimization/prompt-caching.md:219) plus a "per-agent model injection" link
  description (line 308); its claude-code/context-memory sibling carries 0 on en.
- 4-locale existence symmetry for every in-scope page: verified in the phase-① survey.

## 3. Requirements (GEARS)

- REQ-AMD-001 (Ubiquitous): The docs-site pages that describe agent model and effort assignment
  shall describe the inheritance narrative — subagents inherit the main session's model and effort,
  and neither `model` nor `effort` is passed or declared — derived from the canonical doctrine
  wording quoted in §1, and shall not present per-agent profile assignment as current behavior.

- REQ-AMD-002 (Event-driven): When a reader opens any of the five A-cluster pages
  (`multi-llm/model-policy`, `advanced/profile-matrix`, `advanced/agent-guide`,
  `advanced/no-haiku-3tier`, `advanced/tokenomics-overview`) in any of the four locales, the page
  shall present the measured sections (edit map: plan.md §F M1) rewritten to the inheritance
  narrative, with the 4-locale same-change obligation satisfied through the ko→en→ja/zh chain.

- REQ-AMD-003 (Event-driven): When a reader consults the getting-started and core-concepts pages
  that document the `moai init` flags (`getting-started/cli.md`, `getting-started/init-wizard.md`,
  `getting-started/introduction.md`, `getting-started/faq.md`,
  `core-concepts/what-is-moai-adk.md`), the page shall document `--model-policy` and `--profile`
  as deprecated stubs that emit a deprecation warning directing the user to `moai profile setup`,
  in all four locales.

- REQ-AMD-004 (Ubiquitous): The `cli-reference/profile.md` page shall be retained in all four
  locales and shall document only the surviving `moai profile` command family; When residual
  model-policy-matrix wording is measured in it, the page shall be rewritten in place — no page
  removal and no vercel.json redirect for this path.

- REQ-AMD-005 (Event-detected): When the run phase decides to remove any docs-site page path, the
  run shall add a corresponding `vercel.json` redirect entry (pattern:
  `.moai/docs/docs-site-i18n-rules.md` §17.1) in the same PR, symmetric across the affected
  locales' URLs. The default disposition recorded by this SPEC is zero page removals and therefore
  zero new redirects; any removal is an explicit deviation that this requirement catches.

- REQ-AMD-006 (Ubiquitous): The prompt-caching pages shall not present per-spawn model injection as
  a live cache-behavior; the measured bullet (`cost-optimization/prompt-caching.md:219` en) and its
  "per-agent model injection" link description (line 308 en) shall be dropped or reduced to the
  inheritance default, and the `claude-code/context-memory/prompt-caching.md` sibling shall be
  re-verified per locale (0 hits on en at baseline).

- REQ-AMD-007 (Ubiquitous): The five `model_policy` wizard strings in
  `internal/cli/wizard/translations.go` (en/ko/ja/zh question title and description, survey §5
  line map) shall be rewritten to main-session wording — the policy feeds only the main-session
  effort fallback via `MapModelPolicyToEffort` — in native wording per locale; and
  `internal/cli/profile_setup_translations.go` shall not regress from its landed main-session
  wording (baseline: line 165 `"Session model policy"`).

- REQ-AMD-008 (Ubiquitous): Every docs-site edit shall follow the HARD rules of
  `.moai/docs/docs-site-i18n-rules.md`: 4-locale same-PR, ko canonical source chain
  (ko→en→ja/zh), `{{< icon >}}` shortcode instead of body emoji, Mermaid TD-only, emphasis-marker
  spacing, and the §17.1 URL blacklist.

- REQ-AMD-009 (Ubiquitous): Korean docs copy shall read as clean native written register (문어) —
  professional native idiom, no translationese (no English-syntax carry-over, no calqued figurative
  stock).

- REQ-AMD-010 (Event-driven): When the run phase completes the edits, the `hns-oss-docs-verify`
  skill's verify recipe shall pass — warning-free hugo build, sitemap existence, URL-blacklist
  grep, Mermaid direction grep, 4-locale file-existence and section parity, body-emoji scan — and
  `go build ./...` plus the affected `go test ./internal/cli/...` packages shall pass.

## 4. Exclusions

### Out of Scope — Go behavior changes
- No changes to `internal/cli/init.go` flag registration or the `deprecatedAgentModelFlagUsage`
  warning behavior — the deprecation stubs stay exactly as landed.
- No change to `MapModelPolicyToEffort` or any model/effort resolution logic.
- No removal of `internal/cli/model.go`-adjacent code beyond the string rewrites of REQ-AMD-007.

### Out of Scope — surfaces already landed by sibling cards
- `README.md` / `README.ko.md` — taken by card t1302 (merge `8a969dfc0`).
- `.claude/rules/**`, `.claude/agents/**`, template mirrors — taken by SPEC-AGENT-MODEL-INHERIT-001.
- docs-site `cli-reference/launchers.md` and `tokens.md` role-vocabulary edits — taken by t1257
  (merge `afecf81e9`); this SPEC only touches the A/B/C clusters measured in the phase-① survey.

### Out of Scope — profile system behavior
- The `moai profile` command family (CLAUDE_CONFIG_DIR isolation) is live product surface — its
  behavior and its cli-reference page survive unchanged except for model-policy residue cleanup.
- Page removals and vercel.json redirects are out of the default path (REQ-AMD-005 records the
  conditional); in-place rewrite is the disposition for every in-scope page.
