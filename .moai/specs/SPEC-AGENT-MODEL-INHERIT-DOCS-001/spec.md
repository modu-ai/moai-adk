---
id: SPEC-AGENT-MODEL-INHERIT-DOCS-001
title: "docs-site model/effort docs rewrite to the inheritance narrative + wizard H24 residue (card t1300)"
version: "0.2.1"
status: in-progress
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
| 0.2.0 | 2026-09-29 | manager-spec | plan-audit iter-1 (FAIL 0.75, Tier M threshold 0.80) delta revision, coordinator-relayed and re-verified on this tree. D1 premise-collapse: REQ-AMD-007 reconceptualized as a pure regression guard — the wizard model-policy rewrite already landed pre-baseline in t1246 commit `41cf11c4d`; the 5 `model_policy` occurrences are 1 comment (translations.go:404) + 4 map keys (417/429/441/453) whose values already carry the main-session inheritance wording in all 4 locales, and the init wizard model question itself is retired (`question_removal_test.go` sharedInitRemovedIDs comment). M4 demoted to verify-only. D2: outside-cluster retirement-narrative pages added to the M3 disposition list (autonomy-tier 16/116/117, config-sections 95, self-evolving 96, token-budget 107, advanced/_index 49-50). D4: tokenomics inbound profile-matrix links corrected to 4 (27/51/109/131). |
| 0.2.1 | 2026-09-29 | coordinator (lead-relayed audit D3+D6) | D3 related-specs re-adjustment recorded: this SPEC absorbs the docs surface of SPEC-MODEL-MATRIX-DOCS-001, which the parent SPEC §C Supersession reverses — "This SPEC reverses the target of … SPEC-MODEL-MATRIX-CORE-001 (in-progress), SPEC-MODEL-MATRIX-CONFIG-001 (draft), SPEC-MODEL-MATRIX-SURFACES-001 (draft), and SPEC-MODEL-MATRIX-DOCS-001 (in-progress)" (SPEC-AGENT-MODEL-INHERIT-001/spec.md:106). D6: REQ-AMD-005 label typo `(Event-detected)` → `(Event-driven)` (label harmonized with REQ-AMD-002/003/010). |

---

## 1. Overview

SPEC-AGENT-MODEL-INHERIT-001 (t1246, landed) removed per-agent model/effort assignment from the
product: subagents inherit the main session's model and effort, MoAI agent definitions declare
neither, and the `moai model profile` accessor is deleted. The docs layer still teaches the retired
per-agent profile-matrix narrative: the docs-site A-cluster pages (5 pages × 4 locales), the
getting-started flag docs, and the prompt-caching caveats all predate the change. The Go-side
wizard strings are the exception — the H24 rewrite landed pre-baseline in t1246's own commit
`41cf11c4d` and needs only regression-guard verification (REQ-AMD-007).

This SPEC is the docs-layer sweep. The Go-side H24 wizard rewrite is NOT in scope as new work: it
landed pre-baseline in t1246's own commit `41cf11c4d` — this SPEC only verifies it survives. Its
canonical wording source is the landed doctrine sentence at
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
- `internal/cli/wizard/translations.go` carries 5 `model_policy` occurrences: 1 comment listing
  map keys (line 404) + 4 `model_policy` map keys (417/429/441/453) whose values ALREADY carry the
  main-session inheritance narrative in all 4 locales (en "Session model policy"; ko "세션 모델 정책";
  ja "セッションモデルポリシー"; zh "会话模型策略"). The rewrite landed pre-baseline in t1246 commit
  `41cf11c4d`; per-agent wording count in the file is 0. The init wizard model question itself is
  retired (the `sharedInitRemovedIDs` comment in `question_removal_test.go`). The phase-① survey
  §5's rewrite scope for this file is therefore OBSOLETE — what remains is regression-guard
  verification (REQ-AMD-007).
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

- REQ-AMD-005 (Event-driven): When the run phase decides to remove any docs-site page path, the
  run shall add a corresponding `vercel.json` redirect entry (pattern:
  `.moai/docs/docs-site-i18n-rules.md` §17.1) in the same PR, symmetric across the affected
  locales' URLs. The default disposition recorded by this SPEC is zero page removals and therefore
  zero new redirects; any removal is an explicit deviation that this requirement catches.

- REQ-AMD-006 (Ubiquitous): The prompt-caching pages shall not present per-spawn model injection as
  a live cache-behavior; the measured bullet (`cost-optimization/prompt-caching.md:219` en) and its
  "per-agent model injection" link description (line 308 en) shall be dropped or reduced to the
  inheritance default, and the `claude-code/context-memory/prompt-caching.md` sibling shall be
  re-verified per locale (0 hits on en at baseline). Outside the A-cluster, the pages carrying the
  retirement narrative shall receive a recorded disposition (rewrite the link description to the
  inheritance narrative, or record no-change with its verifying evidence): en
  `advanced/autonomy-tier.md` lines 16/116/117, `advanced/config-sections.md` line 95,
  `advanced/self-evolving.md` line 96, `advanced/token-budget.md` line 107, `advanced/_index.md`
  lines 49-50 — each located by grep, ×4 locales.

- REQ-AMD-007 (Regression guard — no rewrite, no churn): The landed wizard model-policy wording in
  `internal/cli/wizard/translations.go` shall SURVIVE the run phase unchanged. The landed baseline
  (t1246 commit `41cf11c4d`, measured on tree `8a969dfc0`) is, verbatim — line 417 (en):
  `"model_policy": {Title: "Session model policy", Description: "Sets the default reasoning effort
  of the Claude session launched with this profile when no effort level is chosen. Subagents
  inherit the session's model and effort."}`; line 429 (ko): `"model_policy": {Title: "세션 모델
  정책", Description: "추론 강도를 따로 고르지 않았을 때, 이 프로필로 실행하는 Claude 세션의 기본 추론
  강도를 정합니다. 서브에이전트는 세션의 모델과 추론 강도를 그대로 따릅니다."}`; line 441 (ja):
  `"model_policy": {Title: "セッションモデルポリシー", Description: "推論強度を個別に選ばなかったとき、このプロファイルで起動する
  Claude セッションの既定の推論強度を決めます。サブエージェントはセッションのモデルと推論強度をそのまま引き継ぎます。"}`;
  line 453 (zh): `"model_policy": {Title: "会话模型策略", Description: "未单独选择推理强度时，决定使用此配置文件启动的
  Claude 会话的默认推理强度。子代理沿用会话的模型与推理强度。"}`. The run agent MUST NOT edit
  `translations.go` at all, and MUST NOT rename the `model_policy` map keys — they are shared
  config keys; renaming breaks config compatibility. The same guard holds for
  `internal/cli/profile_setup_translations.go` (baseline: line 165 `"Session model policy"`).

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
