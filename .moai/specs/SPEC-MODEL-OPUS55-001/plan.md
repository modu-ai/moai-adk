# plan.md — SPEC-MODEL-OPUS55-001

## §A Context

- Card t1089 (class C, Tier M). Worktree `.claude/worktrees/t1089`, branch `WT-opus-55-default`, dispatched base `17f71a13d`; plan authored at HEAD `6e75b74db` (the lead's first-measure commit on top of that base).
- First measurements by the lead: `.moai/reports/t1089/first-measure-20260923.md` (zone-registry anchors, web edit path, settings template effort key, model_policy.go shape).
- Official facts re-fetched by manager-spec on 2026-09-23 — spec.md §A.2.

## §B Measured inventory (commands + verbatim output; tree `6e75b74db`)

### B.1 Lead baseline re-measured — it does not reproduce

Lead claim: 14 code / 14 template / 13 local files. Re-measured with the literal pattern `opus-5\|opus 5` (case-insensitive):

```
$ grep -rli 'opus-5\|opus 5' internal pkg cmd --include='*.go' --exclude-dir=templates | wc -l
      15
$ grep -rli 'opus-5\|opus 5' internal pkg cmd --exclude-dir=templates   (all file types)
      19   (the 15 .go files + internal/web/assets/i18n.js + 3 internal/cli/testdata/profilewizard/groups-{ja,ko,zh}.golden)
$ grep -rli 'opus-5\|opus 5' internal/template/templates | wc -l
      15
$ grep -rli 'opus-5\|opus 5' .claude CLAUDE.md AGENTS.md .moai/config   (agent-memory excluded)
      15
```

Template 15: `CLAUDE.md`, `.moai/config/sections/quality.yaml.tmpl`, and under `.claude/`: `rules/moai/core/{moai-constitution,moai-constitution-detail,settings-management,zone-registry}.md`, `rules/moai/development/{agent-authoring,model-policy,prompting-best-practices,skill-authoring}.md`, `rules/moai/workflow/{context-window-management,worktree-integration}.md`, `skills/moai-foundation-core/modules/token-optimization.md`, `skills/moai-foundation-thinking/SKILL.md`, `skills/moai-workflow-spec/references/reference.md`.

Local 15: the same 13 `.claude/` paths + `CLAUDE.md` + `.moai/config/sections/harness.yaml` (local carries the Opus 5 mention in `harness.yaml:47`; the template carries it in `quality.yaml.tmpl:9` instead — the two `.moai/config` trees already differ here).

Go files with a `ModelIDOpus5` or Opus-5 hit (18, union of both regexes): `internal/template/{model_policy,profile_matrix}.go` + tests `{model_policy,profile_matrix,glm_slot}_test.go`; `internal/cli/{profile_setup_translations}.go` + tests `{codex_session,glm_slot_effort,launcher,profile_setup_translations,tokens}_test.go`; `internal/cli/wizard/{questions,translations}.go` + `model_policy_matrix_agreement_test.go`; `internal/statusline/session_telemetry_payload_test.go`; `internal/web/{appbar_context,console_ux_fix,session_telemetry_cells}_test.go`.

### B.2 The constant and its consumers

```
$ grep -rn 'ModelIDOpus5' internal --include='*.go' | grep -v 'model_policy.go'
internal/template/glm_slot_test.go:28:		{ModelIDOpus5, GLMSlotHigh},
internal/template/model_policy_test.go:59/60/62/63  (asserts ModelIDOpus5 == "claude-opus-5" and opus alias == ModelIDOpus5)
internal/cli/glm_slot_effort_test.go:47:		{template.ModelIDOpus5, "e-high"},
internal/cli/launcher_test.go:683, 687, 934
```

Production: `internal/template/model_policy.go:51 const ModelIDOpus5 = "claude-opus-5"`, `:78 "opus": ModelIDOpus5`; deprecated map carries `claude-opus-4-6`, `claude-opus-4-7`, `ModelIDOpus48`, `claude-sonnet-4-6` — no `claude-opus-5`.

### B.3 Web edit path (measured: exists; gap is display only)

`internal/settings/schema.go:358-378` — `model` and `effort_level` are `TypeSelect` fields persisted to the profile store; effort options `low/medium/high/xhigh/max` (`:231-234`); effort `EmptyLabelKey: "opt.runtime_default"`. `internal/web/assets/i18n.js`: `"f.model.opt.opus[1m]": "Opus 5"` at 356/1250/2029/2808; `"f.effort_level.opt.medium"` at 443/1337/2116/2895 carries no recommendation marker; `"opt.runtime_default"` at 435/1329/2108/2887.

### B.4 Current-behavior probe P4 (the negative-AC instrument, RED-now)

v1 (iter-1, superseded by v2):

```
$ find internal pkg cmd .claude/rules .claude/skills .moai/config CLAUDE.md -type f \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.tmpl' -o -name '*.js' -o -name '*.golden' \) ! -name '*_test.go' -exec awk 'tolower($0) ~ /opus[ -]5([^.0-9-]|\.[^0-9]|\.$|$)/ && tolower($0) !~ /superseded|measured on opus 5/ {print FILENAME ":" FNR}' {} +
exit=0, 147 lines (full verbatim list: .moai/reports/t1089/p4-rednow-6e75b74db.txt; grouped below)
```

v2 (iter-2, the AC instrument — D11 + D6 repair): adds the roots `.claude/agents .claude/commands .claude/output-styles .moai/project AGENTS.md`, catches `Opus5` (`opus[ -]?5`), and narrows the allow-list from the bare word `superseded` to the literal markers `(superseded)`, `// superseded`, and `measured on Opus 5`:

```
$ find internal pkg cmd .claude/rules .claude/skills .claude/agents .claude/commands .claude/output-styles .moai/config .moai/project CLAUDE.md AGENTS.md -type f \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.tmpl' -o -name '*.js' -o -name '*.golden' \) ! -name '*_test.go' -exec awk 'tolower($0) ~ /opus[ -]?5([^.0-9-]|\.[^0-9]|\.$|$)/ && tolower($0) !~ /\(superseded\)|\/\/ superseded|measured on opus 5/ {print FILENAME ":" FNR}' {} +
exit=0, 153 lines (verbatim sorted list: .moai/reports/t1089/p4v2-rednow-6e75b74db.txt)
```

The 6 lines v2 adds over v1: `.moai/project/product.md:151`, `.moai/project/tech.md:13`, `:14`, `:15`, `internal/template/model_policy.go:54` (the `ModelIDOpus48` doc comment "now replaced by ModelIDOpus5", D15), `internal/template/model_policy.go:78` (`"opus": ModelIDOpus5`). Known limit: an "Opus 5-era" style phrase (`opus 5-e…`) is not matched, because the `-` continuation is excluded to keep `claude-opus-5-5` silent.

Per-file offending line count: profile_setup_translations.go 24; model-policy.md 15 (template) + 15 (local); wizard/translations.go 13; web/assets/i18n.js 8; moai-constitution.md 5 + 5; groups-ko/ja.golden 5 each, groups-zh.golden 4; moai-foundation-thinking/SKILL.md 3 + 3; agent-authoring.md 3 + 3; template/model_policy.go 3; wizard/questions.go 3; token-optimization.md 2 + 2; context-window-management.md 2 + 2; moai-constitution-detail.md 2 + 2; settings-management.md 2 + 2; template/profile_matrix.go 2; one each: CLAUDE.md (template + local), quality.yaml.tmpl, local harness.yaml, reference.md (template + local), worktree-integration.md (template + local), skill-authoring.md (template + local), prompting-best-practices.md (template + local).

Pattern notes: `opus[ -]?5` followed by `.`+digit is excluded so "Opus 5.5" / "claude-opus-5-5" never match; a trailing sentence period ("… Opus 5.") still matches. Allowed survivors (v2) carry the literal marker `(superseded)` in prose, `// superseded` in Go (the deprecated-id row), or `measured on Opus 5` (benchmark attributions, REQ-OP55-012). Zone-registry anchors (`#opus-5-48-…`) are NOT matched by P4 (followed by `-4`) — they have their own AC.

### B.7 Version-agnostic high-default probe P6 (RED-now) — D2

P4 cannot see a default-effort claim that names no Opus version. P6 searches for the statement shapes themselves:

```
$ grep -rnoE '`high` \(default\)|high: the default|defaults? to `effort: high`|default effort: high' .claude/rules .claude/skills internal/template/templates/.claude internal/template/templates/CLAUDE.md CLAUDE.md .moai/project
.claude/rules/moai/core/moai-constitution.md:57:default to `effort: high`
.claude/rules/moai/development/agent-authoring.md:380:defaults to `effort: high`
.claude/rules/moai/development/model-policy.md:19:default effort: high
.claude/rules/moai/development/model-policy.md:230:high: the default
.claude/rules/moai/workflow/dynamic-workflows.md:126:`high` (default)
internal/template/templates/.claude/rules/moai/core/moai-constitution.md:57:default to `effort: high`
internal/template/templates/.claude/rules/moai/development/agent-authoring.md:380:defaults to `effort: high`
internal/template/templates/.claude/rules/moai/development/model-policy.md:19:default effort: high
internal/template/templates/.claude/rules/moai/development/model-policy.md:230:high: the default
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:126:`high` (default)
.moai/project/tech.md:15:`high` (default)
exit=0 (11 lines)
```

Companion probe P5 (the wrong-direction rewrite, "Opus 5.5 … default … high"): `grep -rnE 'Opus 5\.5[^.|]*default[^.|]*high' .claude/rules .claude/skills internal/template/templates/.claude internal/template/templates/CLAUDE.md CLAUDE.md .moai/project` → no output, exit 1 (regression-guard; red only under the mutant "Opus 5.5 defaults to `effort: high`").

Fact-line probe P7 (REQ-OP55-009, single invocation):
```
$ awk '/^- opus = Opus 5\.5/ && /claude-opus-5-5/ && /2\.1\.280/ && /1M/ && /128K/ && /\$4/ && /\$20/ && /medium/ && /always on/ {n++} END {print n+0}' .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md
0
exit=0
```

### B.5 Local/template pair identity (pre-state)

```
$ cmp -s .claude/<f> internal/template/templates/.claude/<f>  (per pair)
SAME rules/moai/core/zone-registry.md
SAME rules/moai/development/model-policy.md
SAME rules/moai/development/prompting-best-practices.md
SAME rules/moai/workflow/context-window-management.md
SAME skills/moai-foundation-thinking/SKILL.md
SAME rules/moai/workflow/dynamic-workflows.md   (measured in iter-2: cmp rc=0)
DIFF rules/moai/core/moai-constitution.md, moai-constitution-detail.md, settings-management.md
DIFF rules/moai/development/agent-authoring.md, skill-authoring.md
DIFF rules/moai/workflow/worktree-integration.md
DIFF skills/moai-foundation-core/modules/token-optimization.md, skills/moai-workflow-spec/references/reference.md
DIFF CLAUDE.md (root vs template)
```

The six SAME pairs must stay SAME (AC-OP55-010). DIFF pairs are edited on both sides, each against its own text.

### B.6 Other measured facts

- `grep -n -i effort internal/template/templates/.claude/settings.json.tmpl` → no output, rc=1 (no effort key today).
- `settings-management.md:93` (both copies): `effortLevel` "Intentionally NOT shipped in `settings.json.tmpl`"; the launcher passes the profile effort as `effortLevel` in the transient `--settings` file.
- Zone-registry: `anchor: "#opus-5-48-prompt-philosophy"` at lines 311 and 319, both copies; the anchors are resolved by `internal/constitution/registry_sync_test.go` (`TestRegistrySyncGuard`, `TestRegistrySyncMirrorsIdentical`).
- Label guards that derive the version from the canonical id: `internal/cli/wizard/model_policy_matrix_agreement_test.go:64-66` (`strings.TrimPrefix(opusID, "claude-opus-")`) and `internal/cli/profile_setup_translations_test.go` (`wantVersion != "5"` hard check). With `claude-opus-5-5` the TrimPrefix yields `5-5`, not `5.5` — see §E K2.
- docs-site: `grep -rli 'opus-5\|opus 5' docs-site | grep -v node_modules | wc -l` → 49 files (4 locales).
- No moai Go path sends `tool_choice` / `budget_tokens` / `computer_20251124` to the Anthropic API: hits are `internal/cli/mcp_glm.go` (z.ai GLM), `internal/cli/agentlint/agent_lint.go` (lint rule), and tests.

## §C Decisions

### C.1 (a) `claude-opus-5` joins the deprecated-id map → `opus`

Decision: YES. Remove the `ModelIDOpus5` constant; add `ModelIDOpus55 = "claude-opus-5-5"` as the `opus` target; add `"claude-opus-5": "opus"` as a string-literal row of the deprecated-id map (same form as the 4-6 / 4-7 rows), with a comment carrying the word `superseded`.
Why: without the row a prefs file that carries the full `claude-opus-5` id stops normalizing to the `opus` alias. Precedent, stated accurately (D15): the deprecated-map row is the pattern at every previous bump, but the 4.8 → 5 bump ALSO kept a named constant (`ModelIDOpus48`); this SPEC deliberately does not keep one for `claude-opus-5` (REQ-OP55-003). Reading of the card: "remove the opus-5 constant/table row" is taken as "remove the constant and repoint the alias-table row"; the added deprecated-map row is a different table and is what keeps stored `claude-opus-5` values resolving. The `ModelIDOpus48` doc comment ("now replaced by ModelIDOpus5") is updated to name Opus 5.5; AC-OP55-003's word grep catches it if missed. Trade-off stated: a user who deliberately stored `claude-opus-5` is normalized to the alias (now Opus 5.5) by wizard migration; Claude Code itself still accepts `claude-opus-5` typed as a full name, so the id is not blocked — only removed from the moai catalog, which is the operator's decision.

### C.2 (b) No template `effortLevel` injection

Decision: NO injection. `settings.json.tmpl` stays free of any effort key (REQ-OP55-013).
Why: (1) `settings-management.md:93` records this as intentional; (2) the official model-config page says a top-level `effortLevel` in project settings "applies to every model" — injecting `medium` would also pin Sonnet 5 and Fable 5.1, whose native default is `high`, so it changes behavior beyond the target model; (3) Opus 5.5 already starts at `medium` natively, so the injection adds nothing for the target; (4) `effortLevel` is runtime-managed per CLAUDE.local.md §2 (settings.local separation). The medium recommendation lives at the profile layer (wizard + web), where the launcher's transient `--settings` path already carries an explicit choice.

### C.3 (c) docs-site + READMEs split to a separate card

Decision: SPLIT. 49 docs-site files × 4-locale same-PR obligation + Vercel deploy is its own oss-docs harness job; the README benchmark tables are measured facts on Opus 5 and must not be rewritten. The orchestrator issues the follow-up card (queue mutation is the lead's act).

### C.4 Constitution heading and anchor

Rename `## Opus 5 / 4.8 Prompt Philosophy` → `## Opus 5.5 Prompt Philosophy` in constitution + constitution-detail (both copies); zone-registry anchors for CONST-V3R2-028/029 become `#opus-55-prompt-philosophy` (both copies); every "§ Opus 5 / 4.8 Prompt Philosophy" cross-reference (agent-authoring.md:222, prompting-best-practices.md:62, constitution-detail description) follows. Zone Evolvable, `canary_gate: false` — no constitution-amend gate. The two clause literals ("Principle 4 — …", "Principle 5 — …") stay byte-identical.

### C.5 Recommendation scope

The recommendation is a label/marker (REQ-OP55-005/006/007), not a behavior change: the launch-effort fallback and the per-agent matrix are untouched (spec.md §C).

### C.6 High/xhigh effort-statement inventory (D2)

Source: `grep -rnE "xhigh|\bhigh\b.{0,20}default|default.{0,30}\bhigh\b"` over `.claude/rules/moai`, the thinking skill, `CLAUDE.md`, `.moai/config/sections/{harness,quality}.yaml`, `.moai/project/{tech,product}.md`, `internal/template/profile_matrix.go`, filtered to effort/opus/default lines (GLM reasoning-effort lines excluded), plus P6 (§B.7). Line numbers are local. Template line numbers were measured identical for the REWRITE rows 1-5 and row 7 (constitution:57-58, agent-authoring:380, model-policy:19/230/234 via the SAME pair, dynamic-workflows:126 via the SAME pair); KEEP rows 9-14 were not line-pinned in the template copy because nothing is edited there.

| # | Location (both copies unless noted) | Current text (abridged) | Decision | Reason | AC |
|---|---|---|---|---|---|
| 1 | moai-constitution.md:57-60 | "Opus 5 and 4.8 default to `effort: high` everywhere; … Use `xhigh` for coding and agentic work, keep a minimum of `high` …" | REWRITE | False default claim. New text: Opus 5.5 defaults to `medium` (other effort-capable models `high`); moai's recommended session effort is `medium`; raise per role (`high`/`xhigh`/`max`) where the work needs it | AC-OP55-006, P6, P5 |
| 2 | agent-authoring.md:380 | "Opus 5 defaults to `effort: high` …; raise to `xhigh` for coding/agentic work" | REWRITE | False default claim; per-role raise stays as advice | AC-OP55-006, P6, P5 |
| 3 | model-policy.md:19 | "(default effort: high across all surfaces …; set xhigh explicitly for coding/agentic work …)" | REWRITE | This is the canonical fact line (REQ-OP55-009) | AC-OP55-006 (P7), P6 |
| 4 | model-policy.md:230 | "high: the default on Opus 5 / Opus 4.8 …; minimum for intelligence-sensitive work" | REWRITE | False default claim; the "minimum for intelligence-sensitive work" advice is kept | AC-OP55-006, P6 |
| 5 | dynamic-workflows.md:126 (SAME pair) | "`low`, `medium`, `high` (default), `xhigh`, `max`" | REWRITE | Version-agnostic claim false on Opus 5.5; becomes "`high` (default on most models; `medium` on Opus 5.5)" | P6, AC-OP55-006 |
| 6 | `.moai/project/tech.md:15` (local only) | "`high` (default)" + "Opus 5 carries …" | REWRITE | Same claim; project doc brought into scope (D6) | P6, P4 v2 |
| 7 | model-policy.md:234 | "On Opus 5, `low` and `medium` are stronger than on earlier Opus models" | KEEP, attribute | Vendor statement about Opus 5; mark `(superseded)` rather than transfer it to Opus 5.5 unverified | P4 v2 |
| 8 | model-policy.md:157 + profile_matrix.go:280 | "`xhigh` … on Opus scores the same as `high`" / "on Opus 5 it scores the same as" | KEEP, attribute | Benchmark measured on Opus 5 (REQ-OP55-012): wording "measured on Opus 5" | P4 v2 |
| 9 | agent-authoring.md:219 | "(xhigh for coding/agentic, minimum high for intelligence-sensitive)" | KEEP | Per-agent escalation guidance, not a default claim; says nothing about `medium`, so it does not contradict the `medium` session default | — (P6 does not match) |
| 9b | prompting-best-practices.md:35 — local `.claude/rules/moai/development/prompting-best-practices.md:35` and template `internal/template/templates/.claude/rules/moai/development/prompting-best-practices.md:35` (SAME pair, measured at `084722b91`) | "**Effort calibration**: `xhigh` for coding/agentic work, minimum `high` for intelligence-sensitive work, `medium`/`low` only for speed-critical or simple tasks." | REWRITE (plan-audit iter-2 N1) | Framing `medium` as reserved for speed-critical/simple tasks contradicts the `medium` recommendation (REQ-OP55-005). New sentence: `medium` is Opus 5.5's default and moai's recommended session effort; raise per role (`high`/`xhigh`/`max`) where the work needs it; `low` for speed-critical or simple tasks | AC-OP55-006e |
| 10 | dynamic-workflows.md:135-138 | per-purpose table (research high/xhigh, implement xhigh, …) | KEEP | Explicit per-spawn effort for workflow agents, not a default | — |
| 11 | CLAUDE.md:141/143, thinking SKILL.md:294, session-handoff.md:88, session-handoff-examples.md:147/283, model-policy.md:240 | `ultrathink` → `effort: xhigh`; `/effort ultracode` → xhigh | KEEP (CLAUDE.md:141 only drops "Opus 5/4.8" via P4) | Explicit per-turn keyword; no default claim | P4 v2 (CLAUDE.md:141) |
| 12 | verify-judge-effort-contract.md:11 | "falls back to `high`" | KEEP | moai judge contract default, not a model default | — |
| 13 | moai-mcp-tools-catalogue.md:70 | "defaults to `sonnet/high`" | KEEP | codex tool default, unrelated to Opus | — |
| 14 | agent-authoring.md:61/83, skill-authoring.md:25, coding-standards.md:103, worktree-integration.md:456, quality.yaml.tmpl:9, harness.yaml:47 | effort value lists / availability | KEEP (Opus-5 names updated via P4) | Value lists, not default claims | P4 v2 |

### C.7 Effort empty-option wording (D1)

Measured: `resolveLaunchEffort` (launcher.go:1171) returns the explicit effort, else `MapModelPolicyToEffort(model_policy)`, else `""`; `applyLaunchEffort` (launch_effort_settings.go:52) injects the non-empty result as `effortLevel` via `--settings`. `TestResolveLaunchEffort` (launcher_test.go:900) already pins the three fallback cases (high→high, medium→medium, low→low) and the both-empty case. Decision: the web empty-option label (key `opt.runtime_default`, whose only non-test consumer is `internal/settings/schema.go:374`) states the order honestly — "model policy effort if set, otherwise Claude Code's default (medium on Opus 5.5)" — in each locale. The resolution code is not changed; `TestResolveLaunchEffort` is the behavioral pin that keeps the label and the code aligned (AC-OP55-007).

### C.8 Web model label (D3)

Decision: Option A — keep the English-unified invariant of `TestModelOptLabelsEnglishUnified` (internal/web/console_ux_fix_test.go:197). The `f.model.opt.opus[1m]` label becomes the one English string `Opus 5.5 (Recommended)` in all four locales, and that test's expected value is updated to it in M2. Effort option labels are already localized, so the medium marker is per-locale there: `Medium (Recommended)` / `중간 (권장)` / `中 (推奨)` / `中 (推荐)`.

## §D Constraints and Template-First order

1. Template copy first, then local copy, for every paired rule/skill/config file.
2. `make build` after template edits (embedded FS). No template agent `.md` change is planned, so `make agents-emit` is not required; if one does change, run `make agents-emit` and commit the emitted `.codex/agents/moai/*.toml`.
3. The three wizard golden files (`internal/cli/testdata/profilewizard/groups-{ja,ko,zh}.golden`) are regenerated with the test's `-update-golden` flag (profile_setup_golden_test.go:90), never hand-edited.
4. Template neutrality: no SPEC IDs, REQ tokens, internal dates, commit SHAs, or CLAUDE.local references in template edits; the dated official facts stay out of templates (state "Claude Code v2.1.280+" — a version, not a date).
5. Stage by explicit pathspec; every commit names card t1089.

## §E Known issues for the run phase

- **K1 — substring trap.** "Opus 5.5" contains "Opus 5"; "claude-opus-5-5" contains "claude-opus-5". Every negative assertion (tests and greps) must use a boundary that excludes the dotted/hyphenated continuation — see P4 in §B.4.
- **K2 — version derivation.** Both label guards derive the marketing version by trimming `claude-opus-`; for `claude-opus-5-5` that yields `5-5`. The guard must map hyphens to dots (`5.5`) and its negative side must reject a bare "Opus 5" (REQ-OP55-008). `profile_setup_translations_test.go` also hard-checks `wantVersion != "5"`.
- **K3 — fixtures stay.** `tokens_test.go`, `session_telemetry_*_test.go`, `appbar_context_test.go`, `codex_session_test.go` use `claude-opus-5` as an arbitrary valid model string; leave them unless a test asserts the alias target.
- **K4 — `max` in settings.** The model-config page lists `effortLevel` values `low/medium/high/xhigh`; the launcher may pass `max`. Not changed here; recorded as residual risk.
- **K5 — local `.moai/config` drift.** Local `harness.yaml:47` and template `quality.yaml.tmpl:9` carry different Opus-5 comments; edit each where it lives; do not "reconcile" the two trees in this SPEC.
- **K6 — launch `--model` change (D8).** REQ-OP55-001 is a runtime change, not only a label change: `expandModelString` (launcher.go:1139-1152) expands a profile model `opus[1m]` to `claude-opus-5-5[1m]`. On Claude Code below v2.1.280 that id is not a known model. moai has no Claude Code version floor in Go, so this SPEC adds none; the risk is stated in the rule prose (the v2.1.280 minimum in the fact line) and accepted as residual.
- **K7 — two independent label guards (D5).** `TestGetProfileText_OpusAliasValues` (internal/cli/profile_setup_translations_test.go) reads `profile_setup_translations.go`; the wizard agreement test (internal/cli/wizard/model_policy_matrix_agreement_test.go) reads `wizard/questions.go` + `wizard/translations.go`. Each needs its own mutant (AC-OP55-009a/b), and both must follow an alias-target change (AC-OP55-009c).

## §F Milestones (ordered by decision-reversibility, highest first)

- **M1 (Priority High) — model identity + legacy map.** `internal/template/model_policy.go` (constant rename, alias row, deprecated row, doc comments incl. price $4/$20 and CC 2.1.280); consuming tests in `internal/template`, `internal/cli` (launcher, glm_slot_effort), with RED-first tests for REQ-OP55-001/002/003.
- **M2 (Priority High) — recommendation display (user-facing).** TUI wizard labels (`profile_setup_translations.go`, `wizard/questions.go`, `wizard/translations.go`) + web i18n (`internal/web/assets/i18n.js`) for Opus 5.5 labels, medium-recommended, Opus 5.5 recommended, effort empty-option wording (C.7); `TestModelOptLabelsEnglishUnified` expected value updated to `Opus 5.5 (Recommended)` (C.8); label-drift guards per K2/K7; golden regeneration (3 files).
- **M3 (Priority Medium) — rule prose, template first.** The 13 `.claude/` template files + `dynamic-workflows.md` + `CLAUDE.md` + `quality.yaml.tmpl`, then their local copies + local `harness.yaml` + `.moai/project/{tech,product}.md` (local-only project docs, D6); constitution heading + zone-registry anchors (C.4); every REWRITE row of C.6 (REQ-OP55-010), including row 9b `prompting-best-practices.md:35` in both copies (AC-OP55-006e); benchmark attributions (REQ-OP55-012) incl. the `profile_matrix.go` comment.
- **M4 (Priority Medium) — rebuild and verify.** `make build`; acceptance probes AC-OP55-001..016; scoped tests (template, cli/wizard, web, settings, constitution) + `internal/cli` under a slot lease with `-timeout 25m`.

## §G Anti-patterns

- Rewriting Opus 5 benchmark numbers as Opus 5.5 numbers (fabrication).
- Touching `docs-site/`, `README*.md`, `CHANGELOG.md`, or other SPEC directories.
- Injecting `effortLevel` into any template settings file.
- A negative grep for "Opus 5" without the continuation boundary (passes or fails for the wrong reason).

## §H Open items

- 0 open clarification markers.
- Follow-up card to be issued by the lead: docs-site 4-locale + README prose update for Opus 5.5 (C.3).
