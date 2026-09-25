---
id: SPEC-LAUNCHER-AUTOMODE-WORDING-001
title: "Correct the auto permission-mode requirement wording and the acceptEdits 'project default' claim in the moai cc / moai glm launchers and the launchers CLI reference"
version: "0.1.0"
status: draft
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, docs-site/content"
lifecycle: spec-anchored
tags: "permission-modes, auto-mode, acceptEdits, launcher-help, glm, docs-site, i18n, wording-drift"
tier: S
related_specs: [SPEC-AUT-PERMMODES-001]
---

# SPEC-LAUNCHER-AUTOMODE-WORDING-001 — auto-mode requirement wording + acceptEdits default claim

> Card: **t1182**. Evidence path for the card (lead-read, created at close, not now): `.moai/reports/t1182/verdict.md`.

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| "0.1.0" | 2026-09-25 | manager-spec | Initial plan-phase draft (Tier S: spec.md + plan.md + progress.md; AC inline in §3). Premise measured on branch `WT-auto-mode-help`, base develop `a520187f1`. |

## §1 Context and Problem

### §1.1 Official facts (the reference the current wording contradicts)

The official Claude Code permission-modes page (`https://code.claude.com/docs/en/permission-modes`, fetched by the orchestrator 2026-09-25 and passed in the dispatch — not re-fetched in this plan run) states:

- **Organization**: on Team and Enterprise, auto mode is available by default; administrators can turn it off (`permissions.disableAutoMode`).
- **Model**: on the Anthropic API and Claude Platform on AWS — Claude Opus 4.6 or later, Sonnet 4.6 or later, or a Fable model. On Amazon Bedrock, Google Cloud's Agent Platform, Microsoft Foundry, and signed-in Claude apps gateway sessions — only Claude Sonnet 5, Opus 4.7 or later, and the Fable models.
- **Built-in starting mode**: on Pro, Max, and Team plans, the built-in starting permission mode is auto mode (Claude Code v2.1.228+).

The eligibility matrix is provider-dependent and moves with every model release. Any fixed "plan X + model Y" phrase in a shipped binary goes stale on the next release.

### §1.2 Drifted surfaces (measured on this tree, base `a520187f1`)

| # | Location | Current text (verbatim) | Defect |
|---|---|---|---|
| T1 | `internal/cli/cc.go:103` | `auto               Background classifier checks actions (requires Team plan + Sonnet/Opus 4.6)` | "Team plan" contradicts the doc (Team/Enterprise org availability is one axis; Pro/Max start in auto); "Sonnet/Opus 4.6" omits the provider split and Fable. |
| T2 | `internal/cli/glm.go:319` | `auto mode requires Claude Sonnet 4.6 or Opus 4.6 running on Anthropic's API` | Same stale model list; the rejection itself is correct (GLM is not a supported model), only the stated reason is wrong. |
| T3-T6 | `docs-site/content/{ko,en,ja,zh}/cli-reference/launchers.md:47` | e.g. en: "`acceptEdits` (project default) … requires a Team plan + Sonnet/Opus 4.6 or later." | Same plan/model drift, plus the "(project default)" claim (T7). |
| T7 | `internal/cli/cc.go:101` and the same line 47 × 4 locales | `acceptEdits … (project default)` | False premise — see §1.3. |

Baseline counts (measured this run): `grep -c "Team plan\|Sonnet/Opus 4.6\|(project default)\|Sonnet 4.6 or Opus 4.6"` → `cc.go:2`, `glm.go:1`; `grep -c "4\.6"` → 1 in each of the 6 target files; the four locale-specific drifted-phrase patterns → 1 line in each `launchers.md`.

### §1.3 What "acceptEdits (project default)" actually is — repo evidence

- The project settings template carries **no** `defaultMode` key: `grep -n "defaultMode" internal/template/templates/.claude/settings.json.tmpl` → no match (exit 1). So there is no project-scope acceptEdits default.
- `internal/cli/launcher.go:1088` / `:1107-1108` (`syncPermissionModeToSettingsLocal`) treat `""` and `"acceptEdits"` as "no override": the profile's `defaultMode` is removed from `.claude/settings.local.json`, so whatever lower-precedence value exists applies. Its comment ("The project settings.json default is \"acceptEdits\"") is itself stale.
- Where acceptEdits really comes from: `moai init` calls `applyAutonomyTierBundleFn` (`internal/cli/init.go:932`) → `project.ApplyAutonomyTierBundle` (`internal/core/project/autonomy_bundle.go:71`), whose semi-auto / unset default writes **USER-scope** `defaultMode="acceptEdits"` to `~/.claude/settings.json` (`autonomy_bundle.go:22-24`, `config.TierDefaultMode` → `"acceptEdits"` at `internal/config/autonomy_tiers.go:86`; decided by SPEC-AUT-PERMMODES-001 REQ-002/REQ-004).
- Consequence: acceptEdits is the **`moai init` default** (written to user settings), not a project default. Absent that user-scope record, Claude Code's own built-in starting mode applies (auto on Pro/Max/Team per §1.1).

## §2 Requirements (GEARS)

### REQ-001 — cc help: auto eligibility defers to the official doc (Ubiquitous)

The `moai cc` help text shall describe auto mode's eligibility by deferring to the Claude Code permission-modes documentation (a supported model and plan are required), and shall name no specific plan tier and no model version number.

### REQ-002 — cc help: acceptEdits labelled as the moai init default (Ubiquitous)

The `moai cc` help text shall label `acceptEdits` as the `moai init` default instead of the project default.

### REQ-003 — glm: rejection kept, reason corrected (Event-driven)

**When** `moai glm` receives `--permission-mode auto` (space or equals form), the launcher shall reject the launch before starting Claude Code exactly as today (same returned error, same `moai cc --permission-mode auto` hint), and its stderr reason line shall state that auto mode requires a supported Claude model and that GLM models are not supported, naming no model version number.

### REQ-004 — launchers reference: four locales corrected together (Ubiquitous)

The `cli-reference/launchers.md` permission-mode sentence shall, in all four locales (ko / en / ja / zh) within one commit, label `acceptEdits` as the `moai init` default and state that the plans and models supporting `auto` are defined by the Claude Code permission-modes documentation, linked at `https://code.claude.com/docs/en/permission-modes`, with the same meaning in every locale and native idiom per locale.

### REQ-005 — no new eligibility literals in Go strings (Unwanted)

The change shall not introduce any plan-tier name (Pro / Max / Team / Enterprise) or model version literal into a Go user-facing string.

### REQ-006 — scope fence (Unwanted)

The change shall not modify any file other than the six targets (`internal/cli/cc.go`, `internal/cli/glm.go`, `docs-site/content/{ko,en,ja,zh}/cli-reference/launchers.md`), the tests that pin their wording, and this SPEC's own artifacts.

## §3 Acceptance Criteria (inline, Tier S)

Each AC is binary and mechanically checkable on the run-phase tree.

- **AC-001** (REQ-001, REQ-002) — Given the run-phase tree, When `grep -c "Team plan\|Sonnet/Opus 4.6\|(project default)\|4\.6" internal/cli/cc.go` runs, Then it prints `0`; and When `grep -c "(moai init default)" internal/cli/cc.go` and `grep -c "permission-modes docs" internal/cli/cc.go` run, Then each prints `1`.
- **AC-002** (REQ-003) — Given the run-phase tree, When `grep -c "Sonnet 4.6 or Opus 4.6\|4\.6" internal/cli/glm.go` runs, Then it prints `0`; and When `grep -c "supported Claude model" internal/cli/glm.go` and `grep -c "auto mode is not available with GLM" internal/cli/glm.go` run, Then each prints `1`.
- **AC-003** (REQ-004) — Given the four `launchers.md` files, When `grep -c "Team 플랜\|Team plan\|Team プラン\|Team 方案\|(프로젝트 기본)\|(project default)\|(プロジェクトデフォルト)\|(项目默认)\|4\.6"` runs over `docs-site/content/{ko,en,ja,zh}/cli-reference/launchers.md`, Then every file reports `0`.
- **AC-004** (REQ-004, parity) — Given the four `launchers.md` files, When `grep -c "code.claude.com/docs/en/permission-modes"` and `grep -c "acceptEdits\`.*moai init"` run on each, Then each file reports `1` for both, and all four matches sit on the same line number (the permission-mode sentence).
- **AC-005** (REQ-001, REQ-003) — Given the run-phase tree, When `go test ./internal/cli/ -run 'TestCharacterize_CC_HelpFlag|TestCharacterize_GLM_AutoMode'` runs (plus any test the run phase adds to pin the new wording, named in progress §E.2), Then it exits 0, and at least one test asserts the help output contains `moai init default` and `permission-modes docs` and at least one asserts the GLM stderr contains `supported Claude model` and does not contain `4.6`.
- **AC-006** (REQ-005) — Given the branch diff against its base, When `git diff a520187f1 -- internal/cli/cc.go internal/cli/glm.go | grep '^+' | grep -v '^+++' | grep -cE 'Sonnet|Opus|Fable|[0-9]\.[0-9]|\b(Pro|Max|Team|Enterprise)\b'` runs, Then it prints `0`.
- **AC-007** (REQ-006) — Given the branch diff against its base, When `git diff --name-only a520187f1...HEAD` runs, Then every listed path is one of the six targets, a `internal/cli/*_test.go` file, or under `.moai/specs/SPEC-LAUNCHER-AUTOMODE-WORDING-001/` (the card verdict under `.moai/reports/t1182/` is also permitted).

## §4 Exclusions

### Out of Scope — other drifted permission-mode surfaces (recorded as follow-up cards, not fixed here)

- `internal/cli/profile_setup_translations.go:192/288/384/480` `PermAuto` strings ("REQUIRES Max/Team/Enterprise/API plan + Sonnet 5+", × 4 locales) — also contradict §1.1.
- `internal/cli/profile_setup.go:31` `acceptEditsConfirmationLine` and its `acceptEditsConfirmationTexts` translations ("acceptEdits is the project default") — same false premise as T7, but pinned by `profile_setup_acceptEdits_test.go` under SPEC-AUT-PERMMODES-001's lineage (REQ-CCI-006); changing it touches a separate contract.
- `internal/cli/launcher.go:1088` and `:1107-1108` code comments claiming the project settings.json default is acceptEdits, and the `launcher_test.go:511` subtest name "matches project default" — non-user-facing, same premise.

### Out of Scope — behavior changes

- The GLM auto-mode rejection logic (`containsPermissionMode` check and returned error) — unchanged; only the stderr reason line is reworded.
- `syncPermissionModeToSettingsLocal` behavior, the init autonomy-tier bundle, and any `defaultMode` value written anywhere.
- Adding a runtime eligibility probe (plan / model / provider detection) for auto mode.

### Out of Scope — other documentation

- Other docs-site pages, README files, and `.claude/skills/moai-foundation-cc/reference/*` official-doc mirrors.
- Locale-specific Claude Code doc URLs (`/docs/ko/…`, etc.) — every non-en docs-site page in this repo links `/docs/en/` (121 occurrences measured), so the en URL is kept for all four locales.

## §5 Cross-references

- SPEC-AUT-PERMMODES-001 — decided acceptEdits as the init default written to USER scope (the fact T7's corrected wording states).
- `internal/core/project/autonomy_bundle.go`, `internal/cli/init.go:932`, `internal/cli/launcher.go:1082-1127` — §1.3 evidence.
- `.moai/docs/docs-site-i18n-rules.md` — 4-locale same-commit obligation.
