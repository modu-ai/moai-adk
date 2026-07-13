---
id: SPEC-WEBCONF-SIMPLIFY-001
title: "moai web Configuration UI Simplification + Sub-Agent 4-Color Tier Redesign — Acceptance"
version: "0.3.0"
status: in-progress
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/web, internal/settings, internal/template/templates"
lifecycle: spec-anchored
tags: "web-ui, config, sub-agent, tier-model, template-defaults"
tier: L
---

## HISTORY

| Version | Date       | Change                                                                                                                                                      |
|---------|------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-07-13 | Initial plan-phase authoring.                                                                                                                               |
| 0.2.0   | 2026-07-13 | iter-1 audit-fix. GWT-2/REQ-WC-002 drop `front-launch` (D2). GWT-4/AC-WC-004 pin to launch tab (D7). GWT-5/EC-1 rewrite for name-keyed table (D3). AC-WC-017 drop parenthetical (D9). "21 ACs" (D11). Tier M → L. |
| 0.3.0   | 2026-07-13 | In-progress amendment. Refinement 1: GWT-17 + AC-WC-024 git_strategy core surface (`mode` + `merge_method` + `hooks.pre_push`). Refinement 2: GWT-18 + AC-WC-022 + AC-WC-023 per-option descriptions (REQ-WC-015, C-9). AC count 21 → 24. |

---

## §A. Acceptance Strategy

Every acceptance criterion is observable via a mechanical check: a `go test` assertion, a `grep` over shipped files, a `make build` exit code, a render smoke check, or a closed-set validation. Subjective criteria ("feels simpler") are excluded. ACs trace 1:1 to spec.md §B REQs via the §D matrix.

---

## §B. Severity Convention

- **MUST-PASS (blocker)**: failure blocks run-phase PASS. Harmonic-mean scoring treats any MUST-PASS failure as a zero-verdict.
- **SHOULD-PASS**: failure is technical debt, surfaced but non-blocking.
- **NICE-TO-HAVE**: informational; no gate.

---

## §C. Given-When-Then Core Scenarios

### GWT-1 (six-tab render — REQ-WC-001, MUST-PASS)

**Given** the `moai web` console is rendered (or its render test executes),
**When** an auditor enumerates the tab navigation,
**Then** exactly six tabs are present in the order `identity`, `language`, `launch`, `git_strategy`, `llm`, `agentfm`.

### GWT-2 (removed tabs absent — REQ-WC-002, MUST-PASS) [D2]

**Given** the `moai web` console is rendered,
**When** an auditor greps the rendered HTML (or the templ output) for the eleven removed tab identifiers (`project`, `quality_extras` full-tab, `workflow`, `harness`, `ralph`, `feedback`, `observability`, `security`, `mx`, `handoff`, `cache`),
**Then** none of the eleven appears as a top-level tab navigation entry (the `quality_extras` toggle exception is covered by GWT-4). The phantom `front-launch` identifier is NOT in the removed-tab list (no code identifier exists for it).

### GWT-3 (config keys persist — REQ-WC-003 + REQ-WC-005, MUST-PASS)

**Given** the baked template defaults in `internal/template/templates/.moai/config/sections/*.yaml`,
**When** an auditor greps each removed-tab section file for its headline keys,
**Then** every runtime-consumed key (e.g. `cacheStrategy.enabled`, `workflow.execution_mode`, `harness.default_profile`, `security.permission.strict_mode`) is present with the spec.md §E value (or the §D.Δ deliberate default-change value where flagged).

### GWT-4 (quality_extras toggle on launch tab — REQ-WC-004, MUST-PASS) [D7]

**Given** the `quality_extras` tab is removed,
**When** an auditor inspects the **`launch`** tab surface,
**Then** a single enable/disable toggle for the quality-extras feature is rendered on the `launch` tab, and toggling it persists a boolean to the config.

### GWT-5 (tier color from name-keyed table — REQ-WC-006, MUST-PASS) [D3]

**Given** an agent's name (e.g. `manager-spec`) in the 20-agent catalog,
**When** the agentfm UI renders the agent row,
**Then** the row displays the color badge from the name-keyed lookup table (e.g. 🔴 for `manager-spec`), INDEPENDENT of the agent's current `effort` frontmatter value. Symmetrically: `manager-develop`→🟠, `manager-docs`→🔵, `hns-github-specialist`→🩵. The badge reflects the chosen tier, NOT the live effort.

### GWT-6 (tier suggested model+effort — REQ-WC-007, MUST-PASS)

**Given** the `tier → suggested-(model, effort)` table (plan.md §F M1.1 + design.md §D),
**When** an auditor reads the table,
**Then** 🔴→`opus`/`xhigh`, 🟠→`opus`/`high`, 🔵→`sonnet`/`medium`, 🩵→`haiku`/`low`.

### GWT-7 (per-agent override writes effort via agentfm.Patch — REQ-WC-008 + C-7, MUST-PASS) [D3]

**Given** an agent row in the agentfm UI,
**When** the editor selects a tier (color) AND then individually overrides model or effort,
**Then** the override is persisted via the existing `agentfm.Patch` mechanism — the per-agent `model` / `effort` frontmatter reflects the override (NOT the tier default), and NO new frontmatter key (e.g. `tier:`) is written. The 20 agent effort files are untouched by tier DISPLAY; only an explicit user override writes `effort`/`model`.

### GWT-8 (closed-set validation — REQ-WC-009, MUST-PASS)

**Given** a save payload carrying an `effort` or `model` value outside `V4EffortValues` / `V4ModelValues`,
**When** the agentfm validator processes the payload,
**Then** the save is rejected with a validation error naming the offending field and the allowed closed set.

### GWT-9 (deep doctrine preserved — REQ-WC-010, MUST-PASS)

**Given** the documentation cleanup is complete,
**When** an auditor checks the three named doctrine files,
**Then** `mx-tag-protocol.md`, `context-window-management.md`, and `cache-aware-execution.md` are present and byte-identical to their pre-SPEC state (or content-identical modulo unrelated drift).

### GWT-10 (atomic save — REQ-WC-011 + C-5, MUST-PASS)

**Given** a user saves the `launch` tab (a surviving tab),
**When** the `handleSave` handler processes the save,
**Then** no validator for a removed-tab field fires, and the save completes without referencing removed sections.

### GWT-11 (GLM carve-out intact — REQ-WC-012 + C-4, MUST-PASS)

**Given** the cache tab is removed and the cache config is baked,
**When** an auditor reads `internal/runtime/cache_control.go` `InjectCacheControl`,
**Then** the GLM-omit branch is intact (the no-op on GLM backends is preserved).

### GWT-12 (test fallout updated — REQ-WC-013 + C-6, MUST-PASS)

**Given** the schema-section and route-coverage tests,
**When** an auditor runs `go test ./internal/web/... ./internal/settings/...`,
**Then** the suite passes (tests were updated to the 6-tab + tier-model reality, not deleted).

### GWT-13 (4-locale i18n — REQ-WC-014, MUST-PASS)

**Given** the new tier-label strings,
**When** an auditor greps `internal/web/assets/i18n.js`,
**Then** every new tier label has entries in all four locales (en, ko, ja, zh).

### GWT-14 (make build — C-1, MUST-PASS)

**Given** the templ + template edits are complete,
**When** an auditor runs `make build`,
**Then** the build exits 0 and `*_templ.go` files are regenerated.

### GWT-15 (template-neutrality — C-2, MUST-PASS)

**Given** the baked template YAML,
**When** the template-neutrality CI guard (`.github/workflows/template-neutrality-check.yaml`) runs,
**Then** it passes — no SPEC IDs, REQ/AC tokens, commit SHAs, internal dates, or archive paths leak into the shipped YAML.

### GWT-16 (absent-effort agent still gets a badge — EC-1, MUST-PASS) [D3]

**Given** the 3 agents that lack an `effort` frontmatter key entirely (`hns-oss-docs-content-author-specialist`, `hns-oss-docs-locale-translator-specialist`, `hns-oss-docs-structure-curator-specialist`),
**When** the agentfm UI renders their rows,
**Then** each displays the 🩵 badge from the name-keyed lookup table — NOT a fallback, NOT "no badge". The name table is the source of truth; absent effort does not block tier display (Option A eliminates the iter-1 circular-dependency).

### GWT-17 (git_strategy core fields surfaced — REQ-WC-016, MUST-PASS) [Refinement 1]

**Given** the `git_strategy` tab is rendered,
**When** an auditor enumerates the top-level selectors on the tab,
**Then** exactly the three core fields `mode`, `merge_method`, and `hooks.pre_push` are surfaced as top-level selectors for the active mode, and the per-provider nesting (`branch_creation`, `automation`, `commit_style`, `github_integration`, `push_to_remote`, `draft_pr`, `required_reviews`, `branch_protection`) is NOT rendered (baked as template defaults).

### GWT-18 (per-field/option descriptions rendered — REQ-WC-015 + C-9, MUST-PASS) [Refinement 2]

**Given** a selectable field rendered on any of the six surviving tabs (`identity`, `language`, `launch`, `git_strategy`, `llm`, `agentfm`),
**When** an auditor inspects the rendered field,
**Then** the field displays a non-empty user-facing description in the active `conversation_language` below the field label, sourced from `internal/web/assets/i18n.js`; and where the field has selectable options, each option carries a per-option description.

---

## §D. AC Matrix (traceability — 24 ACs) [D11 + Refinement 1/2]

| AC ID      | Maps to REQ        | Severity    | Verification mechanism                                                                                           | GWT   |
|------------|--------------------|-------------|------------------------------------------------------------------------------------------------------------------|-------|
| AC-WC-001  | REQ-WC-001         | MUST-PASS   | Render test asserts tab count == 6 and tab IDs match the ordered list.                                           | GWT-1 |
| AC-WC-002  | REQ-WC-002         | MUST-PASS   | Render test asserts none of the 11 removed tab IDs appears in tab nav.                                           | GWT-2 |
| AC-WC-003a | REQ-WC-003         | MUST-PASS   | `grep` each removed-tab section YAML for headline keys; all present.                                             | GWT-3 |
| AC-WC-003b | REQ-WC-005         | MUST-PASS   | `grep` confirms the values live in `internal/template/templates/.moai/config/sections/*.yaml` (NOT only local). | GWT-3 |
| AC-WC-004  | REQ-WC-004         | MUST-PASS   | Render test asserts the quality_extras toggle is present on the **`launch`** tab (OQ-1 resolved).                | GWT-4 |
| AC-WC-005  | REQ-WC-006         | MUST-PASS   | Agentfm render test: for each of the 20 agents, badge color matches the name-keyed lookup table (plan.md M1.2).  | GWT-5 |
| AC-WC-006  | REQ-WC-007         | MUST-PASS   | Unit test asserts the tier→suggested-(model,effort) table matches the 4-row spec.                                | GWT-6 |
| AC-WC-007  | REQ-WC-008 + C-7   | MUST-PASS   | Agentfm flow test: tier-select auto-suggests model+effort; individual override persists via `agentfm.Patch`; NO new FM key written. | GWT-7 |
| AC-WC-008  | REQ-WC-009         | MUST-PASS   | Unit test: out-of-set effort/model value is rejected by the validator.                                           | GWT-8 |
| AC-WC-009  | REQ-WC-010         | MUST-PASS   | `git diff` confirms the three doctrine files are untouched; grep confirms presence.                              | GWT-9 |
| AC-WC-010  | REQ-WC-011 + C-5   | MUST-PASS   | Save-handler test: saving a surviving tab invokes no removed-tab validator.                                      | GWT-10|
| AC-WC-011  | REQ-WC-012 + C-4   | MUST-PASS   | `grep 'InjectCacheControl' internal/runtime/cache_control.go` + unit test confirms GLM-omit branch intact.       | GWT-11|
| AC-WC-012  | REQ-WC-013 + C-6   | MUST-PASS   | `go test ./internal/web/... ./internal/settings/...` exits 0.                                                    | GWT-12|
| AC-WC-013  | REQ-WC-014         | MUST-PASS   | `grep` i18n.js: each new tier label has 4-locale entries.                                                        | GWT-13|
| AC-WC-014  | C-1                | MUST-PASS   | `make build` exits 0; `*_templ.go` regenerated.                                                                  | GWT-14|
| AC-WC-015  | C-2                | MUST-PASS   | Template-neutrality CI guard passes.                                                                             | GWT-15|
| AC-WC-016  | REQ-WC-006 (20-agent coverage) | MUST-PASS | Unit test enumerates all 20 agents in `.claude/agents/{moai,harness}/` and asserts each has exactly one tier entry in the name table. | GWT-5 |
| AC-WC-017  | C-7 (no new FM key) | MUST-PASS  | `grep -rn '^tier:' .claude/agents/moai/ .claude/agents/harness/` returns no matches — no agent has a `tier:` frontmatter key (Option A: name table, not a FM field). [D9] | —     |
| AC-WC-018  | C-8 (max/inherit override) | SHOULD-PASS | UI renders a neutral "custom" badge when effort is `max` or model is `inherit` (not a 5th color).             | —     |
| AC-WC-019  | full-suite regression | MUST-PASS | `go test ./...` exits 0 (no cascade beyond web/settings; covers effort_mapping triple KI-8).                     | —     |
| AC-WC-020  | render smoke        | SHOULD-PASS | Manual `moai web` launch confirms 6 tabs render and agentfm shows color badges.                                  | —     |
| AC-WC-021  | absent-effort badge (EC-1) | MUST-PASS | The 3 hns-oss-docs-* agents (no `effort` FM) each render a 🩵 badge from the name table. [D3]                | GWT-16|
| AC-WC-022  | REQ-WC-015 + C-9 (descriptions 4-locale) | MUST-PASS | Test enumerates every selectable field across the 6 surviving tabs and asserts each renders a non-empty description in all 4 locales (en/ko/ja/zh). [Refinement 2] | GWT-18|
| AC-WC-023  | REQ-WC-015 (description i18n key convention) | MUST-PASS | `grep` i18n.js: description strings live under the `fieldDesc.<sectionID>.<fieldID>` (and per-option `fieldDesc.<sectionID>.<fieldID>.option.<value>`) convention; no ad-hoc keys. [Refinement 2] | GWT-18|
| AC-WC-024  | REQ-WC-016 (git_strategy core surface) | MUST-PASS | Render test asserts the git_strategy tab surfaces exactly `mode` + `merge_method` + `hooks.pre_push` as top-level selectors; per-provider nesting is absent. [Refinement 1] | GWT-17|

**Total: 24 ACs** (AC-WC-003 split into 003a/003b; AC-WC-021 added for absent-effort under Option A; AC-WC-022/023/024 added by Refinement 1 + Refinement 2). [D11]

---

## §E. Edge Cases

- **EC-1** [D3 rewritten — no longer circular]: An agent frontmatter file lacks an `effort` key entirely (the 3 `hns-oss-docs-*` agents). Under Option A the tier comes from the name-keyed lookup table directly — the badge renders 🩵 from the table, NOT from any effort fallback. Absent effort does not block display and is not a special case. Covered by GWT-16 / AC-WC-021.
- **EC-2**: An agent frontmatter carries `effort: max`. The UI shows a neutral "custom" badge (AC-WC-018); the 4-color concept is not extended to a 5th color. The name-table tier is still known but the badge reflects the override sentinel.
- **EC-3**: An agent frontmatter carries `model: inherit`. The model selector shows `inherit` selected; no tier-suggested model is forced. The name-table tier badge is unaffected.
- **EC-4**: A user hand-edited the local `.moai/config/sections/cache.yaml` after `moai update`. The local value is preserved (KI-2); the SPEC does not auto-overwrite.
- **EC-5**: The `front-launch` tab identifier (OQ-2) does not exist in the current schema (confirmed phantom). The removal is a no-op; the run-phase agent documents this in `progress.md §E.2`. Any stale prose reference to `front-launch` is deleted as dead code.
- **EC-6**: A future agent is added to `.claude/agents/` after this SPEC lands. The name-keyed tier table does not auto-assign; the new agent renders NO badge (or a neutral "unmapped" badge) until a tier is explicitly added to the table. Surface as SHOULD-PASS follow-up, not a blocker for this SPEC.

---

## §F. Indirect Verification

Where direct observation is impractical, indirect evidence is accepted:

- **AC-WC-020 (render smoke)**: if `moai web` cannot be launched in the test environment, a headless render test (Go test executing the templ) is acceptable indirect evidence.
- **AC-WC-009 (doctrine preserved)**: `git diff --name-only origin/main...HEAD -- .claude/rules/moai/workflow/mx-tag-protocol.md .claude/rules/moai/workflow/context-window-management.md .claude/rules/moai/workflow/cache-aware-execution.md` returning empty is acceptable indirect evidence.

---

## §G. Quality Gate Criteria (Definition of Done)

The SPEC is DONE when ALL of the following hold:

1. All MUST-PASS ACs (AC-WC-001 through AC-WC-017, AC-WC-019, AC-WC-021, AC-WC-022, AC-WC-023, AC-WC-024) are green with observed command output cited in `progress.md §E.2`.
2. `make build` exits 0.
3. `go test ./...` exits 0.
4. Template-neutrality CI guard passes.
5. The 20-agent name→tier table (plan.md §F M1.2 + design.md §C) is fully implemented and asserted by AC-WC-016.
6. No agent frontmatter file carries a new `tier:` key (AC-WC-017) — the 20 effort files are untouched by tier DISPLAY.
7. The three deep-doctrine files are untouched (AC-WC-009).
8. SHOULD-PASS ACs (AC-WC-018, AC-WC-020) are either green OR carried forward as documented debt with a `@MX:DEBT` annotation and a follow-up owner.

---

## §H. Forward-Looking Checks (post-close)

- **FL-1**: A follow-up SPEC may productize the 4-color tier model beyond `moai web` (docs-site tokenomics page, README catalog table). This SPEC explicitly does NOT do that (spec.md §D).
- **FL-2**: If the agent catalog grows beyond 20, the name→tier table (plan.md §F M1.2 + design.md §C) must be extended; a CI assertion that "agent count == 20" would be too brittle, so AC-WC-016 enumerates the actual catalog dynamically and checks that every catalog agent HAS a table entry (not that the count is exactly 20).
- **FL-3**: The stale memory `project_agent_token_cost_color_tiers.md` (10-agent split) should be superseded by a fresh memory entry recording the 20-agent name→tier mapping once M1 lands.

---

## §I. Cross-References

- `spec.md` §B (REQs) / §C (Constraints) / §D (Out of Scope) / §D.Δ (deliberate default-changes) / §E (baked keys-of-interest).
- `design.md` §A–§H — Option A name-keyed table design + description-source mechanism (§H).
- `research.md` §A–§C — codebase map + live-template key inventory + 20-agent effort landscape.
- `plan.md` §F (Milestones M1..M9) / §G (Open Questions — all resolved) / §H (Anti-Patterns).
- `CLAUDE.local.md` §2 / §15 / §25 (Template-First, language neutrality, template-internal-isolation).
