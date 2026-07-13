---
id: SPEC-WEBCONF-SIMPLIFY-001
title: "moai web Configuration UI Simplification + Sub-Agent 4-Color Tier Redesign — Progress"
version: "0.2.0"
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

# Progress — SPEC-WEBCONF-SIMPLIFY-001

> Plan-phase artifact (6-artifact Tier L set: spec.md + plan.md + acceptance.md + progress.md + design.md + research.md). Run-phase evidence (§E.2), run-phase audit-ready signal (§E.3), and sync-phase audit-ready signal (§E.4) are populated by `manager-develop` and `manager-docs` in their respective phases. This file carries ONLY the §E.1 plan-phase signal at plan-phase authoring time; §E.2–§E.4 are placeholder headings (era-classification anchor, NOT populated content).

## §E.1 Plan-phase Audit-Ready Signal

**v0.2.0 (iter-1 audit-fix applied)**

- **Plan-phase artifacts (6-artifact Tier L set)**: `spec.md` (v0.2.0) + `plan.md` (v0.2.0) + `acceptance.md` (v0.2.0) + this `progress.md` (v0.2.0) + `design.md` (v0.2.0, NEW) + `research.md` (v0.2.0, NEW) — all six present in `.moai/specs/SPEC-WEBCONF-SIMPLIFY-001/`.
- **Frontmatter**: 12 canonical fields + `tier: L` present in all 6 files; `id` passes `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`; `status: draft`; `version: "0.2.0"`.
- **GEARS compliance**: all REQ-WC-001..014 use `shall` + (ubiquitous | `When` | `While` | `Where`); no legacy `IF/THEN`.
- **Out-of-Scope lint**: `spec.md` §D carries six `### Out of Scope — <topic>` H3 sub-headings, each with `-` bullets.
- **AC traceability**: `acceptance.md` §D matrix maps 21 ACs (AC-WC-001..021) to REQ-WC-001..014 + Constraints C-1..C-8.
- **Tier mapping decidability (Option A)**: plan.md §F M1.2 + design.md §C cover all 20 enumerated agents via the name-keyed lookup table — no agent unassigned, no ambiguity. The 13 M1.2-vs-actual-effort discrepancies counted by the iter-1 auditor are EXPECTED and IRRELEVANT under Option A (tier ≠ effort).
- **Clarification-marker zero-count [D1]**: plan.md §G carries NO unresolved clarification markers — OQ-1 resolved (quality_extras toggle on `launch` tab), OQ-2 resolved (`front-launch` is a phantom). OQ-3/OQ-4 retained as non-blocking design defaults.
- **Option A coherence [D3]**: REQ-WC-006, C-7, EC-1, M1.1, M1.2, GWT-5, GWT-7, GWT-16, AC-WC-005, AC-WC-007, AC-WC-017, AC-WC-021 all describe the SAME single model (name-keyed lookup table, display-only, effort files untouched, per-agent override via existing `agentfm.Patch`). Cross-checked for internal consistency.
- **§E preserve-all-other-keys [D4]**: every §E.N block in spec.md states the contract; spec.md §D.Δ enumerates 9 deliberate default-changes (workflow.execution_mode, security.strict_mode, cache.enabled, git-strategy merge_method/pre_push ×2, harness.effort_mapping ×3).
- **§E completeness [D5]**: spec.md §E now covers project (§E.9), handoff (§E.8), quality_extras (§E.10 — keys inside quality.yaml.tmpl) in addition to the original 7 sections.
- **front-launch phantom [D2]**: dropped from REQ-WC-002, GWT-2, M3; removal is a no-op.
- **Stale-comment sweep scoped [D13]**: plan.md M5 includes `fieldsets.templ:357` ("7 sub-agent"→20) and `agentfm.go:49` ("8 retained agents"→10 moai-custom).
- **SPEC ID pre-write self-check**: `SPEC-WEBCONF-SIMPLIFY-001` → regex PASS; decomposition `SPEC ✓ | WEBCONF ✓ | SIMPLIFY ✓ | 001 ✓ → PASS`.
- **Tier L artifacts**: `design.md` (Option A data-model design) + `research.md` (codebase map + live-template key inventory + 20-agent effort landscape) authored.

## §E.2 Run-phase Evidence

**M1 + M2 evidence** (multi-milestone Tier L SPEC; M3–M9 append their own evidence in later delegation chunks). M1 verified the tier-model backend ACs; M2 verified the template-defaults baking ACs. All other ACs (tab reduction M3, UI M5, i18n M6, test fallout M7, docs M8, full-suite M9) are pending their owning milestone and enumerated at the bottom of this section.

### M1 AC matrix (tier-model backend)

| AC ID | Maps to | Severity | Status | Verification command | Actual output |
|-------|---------|----------|--------|----------------------|---------------|
| AC-WC-005 | REQ-WC-006 (badge from name table, 20 agents) | MUST-PASS | PASS | `go test -run TestAgentTier_All20ExpectedAgents ./internal/harness/v4manifest/` | `--- PASS: TestAgentTier_All20ExpectedAgents (0.00s)` — asserts each of the 20 expected (name→tier) pairs returned by `AgentTier` matches design.md §C. |
| AC-WC-006 | REQ-WC-007 (tier→suggested-(model,effort)) | MUST-PASS | PASS | `go test -run TestTierSuggestedModelEffort ./internal/harness/v4manifest/` | `--- PASS: TestTierSuggestedModelEffort (0.00s)` — 🔴→(opus,xhigh), 🟠→(opus,high), 🔵→(sonnet,medium), 🩵→(haiku,low) per design.md §D. |
| AC-WC-016 | REQ-WC-006 (catalog coverage: 20 agents, no orphan) | MUST-PASS | PASS | `go test -run 'TestAgentTier_CatalogFileCoverage|TestAgentTier_Distribution' ./internal/harness/v4manifest/` | `--- PASS: TestAgentTier_CatalogFileCoverage (0.00s)` + `--- PASS: TestAgentTier_Distribution (0.00s)` — enumerates `.claude/agents/{moai,harness}/*.md` (20 files), asserts every stem has a table entry AND no orphan entry; pins the 🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 distribution. |
| AC-WC-017 | C-7 (no `tier:` FM key — Option A) | MUST-PASS | PASS (PRESERVE) | `grep -rn '^tier:' .claude/agents/moai/ .claude/agents/harness/` | exit 1 (0 matches) — no agent frontmatter carries a `tier:` key. The 20 effort files are untouched by tier DISPLAY (Option A preserved). |
| AC-WC-021 | EC-1 (absent-effort agents get 🩵 from table) | MUST-PASS | PASS | `go test -run TestAgentTier_AbsentEffortAgents ./internal/harness/v4manifest/` | `--- PASS: TestAgentTier_AbsentEffortAgents (0.00s)` — the 3 `hns-oss-docs-*` agents (no `effort` frontmatter) each return `TierLightBlue` from the name table, NOT a fallback. |

**Supporting tests (same run, all PASS)**: `TestTierColor` (tier→emoji glyph map), `TestTierColor_UnknownTierReturnsEmpty`, `TestTierSuggestedModelEffort_UnknownTierReturnsEmpty`, `TestAgentTier_UnknownAgentReturnsNotOK` (v4manifest package); `TestTierForAgent_Delegates`, `TestTierForAgent_UnknownReturnsNotOK`, `TestTierSuggestedModelEffort_Delegates` (settings re-export, internal/settings/tier_test.go). Full suite line: `ok github.com/modu-ai/moai-adk/internal/harness/v4manifest 0.508s` (9/9 PASS) + `ok github.com/modu-ai/moai-adk/internal/settings 0.939s` (3/3 tier tests PASS).

### Option-A preservation evidence (C-7)

- `git status --short .claude/agents/moai/ .claude/agents/harness/` → empty (0 agent files modified).
- `grep -rn '^tier:' .claude/agents/moai/ .claude/agents/harness/` → 0 matches (no new FM key written).

### M2 evidence (template defaults baked — commit e0061ad34)

M2 baked the 11 removed tabs' values into `internal/template/templates/.moai/config/sections/*.yaml` so `moai init` / `moai update` distribute them as shipped defaults. 8 of 9 §D.Δ deliberate default-changes applied; 1 (merge_method capitalization) returned as a structured blocker (E7).

**§D.Δ deltas applied (8/9)**:

| Section.key | Old → New | File |
|-------------|-----------|------|
| `workflow.execution_mode` | `team` → `auto` | workflow.yaml |
| `security.permission.strict_mode` | `false` → `true` | security.yaml |
| `cache.cacheStrategy.enabled` | `false` → `true` | cache.yaml |
| `harness.effort_mapping.minimal` | `medium` → `low` | harness.yaml |
| `harness.effort_mapping.standard` | `high` → `medium` | harness.yaml |
| `harness.effort_mapping.thorough` | `xhigh` → `high` | harness.yaml |
| `git-strategy.personal.hooks.pre_push` | `warn` → `enforce` | git-strategy.yaml.tmpl |
| `git-strategy.team.hooks.pre_push` | `warn` → `enforce` | git-strategy.yaml.tmpl |

**§D.Δ delta NOT applied (1/9 — BLOCKER)**: `git-strategy.*.merge_method` `squash` → `Squash` (capitalized). `"Squash"` is outside the config closed set `validMergeMethods = {squash, merge, rebase}` (all lowercase, `internal/config/validation.go:274`). Baking it would ship a config that fails `validateGitStrategyMergeMethod` on every `moai init`. `merge_method` remains lowercase `squash` (valid) pending orchestrator/user resolution. Manual-mode `pre_push` correctly stays `warn` (§E.4 — not in §D.Δ).

**Preserve-all-other-keys (D4) verified**: `git diff --stat` = 5 files, 8 insertions / 8 deletions — every deletion is a value line, every insertion is the replacement value. Zero key deletions, zero structural changes. research.md §B live keys confirmed present post-edit via grep: harness (plan_audit_global/levels/model_upgrade_review.checklist/rate_limit = 8 matches), workflow (workflow_agents/model_routing/model_routing_profiles/session_name_pattern = 9 matches), security (extra_*_patterns/network_allowlist/env_scrub_extra = 9 matches), llm (performance_tier/plan_type/claude_models/base_url = 5 matches, no edit).

**§E keys-of-interest already matching live (no edit needed)**: harness default_profile/evaluator/mode_defaults/auto_detection/escalation/learning; workflow auto_clear/loop_prevention/token_budget/worktree; security sandbox; cache session_ttl; llm glm.models/mode/team_mode; handoff mode=manual/guide=false. ralph/feedback/observability/mx/project/quality bake as-is (§E.7/§E.9/§E.10 — no edit).

**M2-touched ACs**: AC-WC-003a (REQ-WC-003 keys persist) — M2 portion PASS (all keys present with baked values); AC-WC-003b (REQ-WC-005 values ship from template) — PASS (baked into `internal/template/templates/`); AC-WC-015 (C-2 template-neutrality) — PASS (`grep -rn 'SPEC-WEBCONF\|REQ-WC-\|AC-WC-' internal/template/templates/.moai/config/sections/` → 0 matches).

**Build evidence**: `make build` exit 0 (templates re-embedded via `//go:embed all:templates`); `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test ./internal/template/... ./internal/settings/...` all PASS (no schema break).

### M3 evidence (tab-set reduction + route reclassification)

M3 removed the 11 tabs from the web UI and reclassified the 8 former seam sections to `RouteExcluded` (config keys persist in baked YAML for runtime consumption — REQ-WC-003; web write path removed). The 6 surviving tabs: `identity, language, launch, git_strategy, llm, agentfm`.

**Source edits (atomic — all in the same commit per KI-1/AP-3)**:
- `internal/web/schemaform.go`: `consoleTabs()` 16→6 entries; `schemaSectionMetas()` 12→2 entries (only `SectionGitStrategy` + `SectionLLM` survive as generic fieldsets).
- `internal/settings/sectionroute.go`: 8 former seam sections (workflow, harness, ralph, feedback, observability, security, handoff, cache) removed from `sectionRoutes` (zero-value `RouteExcluded`); `SeamSections()` → empty; `ExcludedSections()` += the 8 (11→19). `project`/`mx` were already zero-value `RouteExcluded`; `quality_extras` shares the `quality` file (stays `RouteTypedSave` — main SectionQuality D12-unaffected; M4 adds the toggle).
- `WriteSectionViaSeam` route gate now rejects all 8 (the web write path is gone); `ApplySchemaEdits` propagates the rejection. `allFields()` NOT pruned (config keys persist as field definitions; the route is the security boundary — forged writes blocked at the gate).

**Test updates (C-6 — UPDATED not deleted; 18 tests across 8 files)**: sectionroute_test.go (3 tests: route table, SeamSections len 0, ExcludedSections len 19); sectionwrite_test.go (golden success tests → rejection tests + 8 added to rejection list); schema_sections_test.go (TestSchemaSectionsRegistered M3-exemption + 2 round-trip tests skip RouteExcluded); web/schema_sections_test.go (5 tests: render-smoke trimmed to surviving sections, save tests for removed sections assert file-unchanged); web/schema_render_test.go (skip removed-section fields); web/m5b_verify_test.go (panel/tab counts 15→6); web/mx_rawview_test.go (mx markers present→absent); web/scope_contract_test.go (8 seam→RouteExcluded + added to exclusions).

**M3-touched ACs**: AC-WC-001 (6-tab render) — PASS; AC-WC-002 (11 removed tabs absent) — PASS; AC-WC-010 (atomic save intact — WriteSectionViaSeam rejects removed sections) — PASS.

**Build evidence**: `make build` exit 0; `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test ./internal/web/... ./internal/settings/...` all PASS; golangci-lint 0 issues; C-HRA-008 boundary 0 matches.

### M4 evidence (simplified surfaces + description mechanism)

M4 delivers (1) the per-option description rendering mechanism (REQ-WC-015, design.md §H option (a)), (2) verifies the already-simplified git_strategy/llm surfaces (REQ-WC-016), and (3) the quality_extras enable/disable toggle on the launch tab (REQ-WC-004 / AC-WC-004).

**Description mechanism (REQ-WC-015, en-only staging — AC-WC-022/023 partial-PASS)**: `FieldDef.Description string` added (schema.go) carrying the `fieldDesc.<sectionID>.<fieldID>` i18n key; `fieldDescription(f)` templ helper renders `.field-description` (conditional on Description != "") in schemaToggleRow + schemaSelectRow; per-option `<option data-i18n-title=...>` (native title tooltip) + app.js applyI18n resolves `data-i18n-title`; `.field-description` CSS; en i18n keys added (fieldDesc.git_strategy.mode + 3 options + fieldDesc.llm.glm.models.{high,medium,low,fable} + f.quality.quality_extras_enabled.{title,desc}); Description populated on git_strategy.mode + llm.glm.models.*. **Partial (en only)**: ko/ja/zh of all fieldDesc.* + agentfm descriptions + Description on remaining fields = M6.

**git_strategy surface (REQ-WC-016 / AC-WC-024 — PASS)**: verified at M4 target — exactly mode + {manual,personal,team}.{merge_method,hooks.pre_push} (7 fields); no per-provider nesting. TestM4GitStrategySurface PASS.

**llm surface (PASS)**: exactly glm.models.{high,medium,low,fable}; mode/team_mode read-only display. TestM4LLMSurface PASS.

**quality_extras toggle on launch (REQ-WC-004 / AC-WC-004 — PASS)**: new quality_extras_enabled FieldDef + QualityConfig.QualityExtrasEnabled + applyQualityKey case; toggle hand-built in fieldsetLaunch. TestM4QualityExtrasToggleOnLaunch PASS. TestM4DescriptionElementRenders PASS.

**Build evidence**: make build (templ + assets re-embedded) exit 0; go build exit 0; windows/amd64 cross-build exit 0; web/settings tests all PASS; golangci-lint 0 issues; C-HRA-008 boundary 0 matches.

### Remaining ACs (pending their owning milestone)

| AC ID | Owning milestone | Status |
|-------|------------------|--------|
| AC-WC-001 / AC-WC-002 (6-tab render, removed tabs absent) | M3 | PASS (consoleTabs 6 entries; 11 removed tabs absent from render + route reclassified) |
| AC-WC-003a / AC-WC-003b (config keys persist + ship from template) | M2 | PASS (M2 portion — keys present with baked values in `internal/template/templates/`) |
| AC-WC-004 (quality_extras toggle on launch tab) | M4 | PASS (toggle renders on launch tab + persists via typed quality path) |
| AC-WC-007 (tier-click auto-suggest writes via agentfm.Patch) | M5 | PENDING (backend accessor `TierSuggestedModelEffort` ready; UI wiring is M5) |
| AC-WC-008 (closed-set validation) | M5 | PENDING (closed sets `V4EffortValues`/`V4ModelValues` already exist; validator wiring is M5) |
| AC-WC-009 (deep doctrine preserved) | M8 | PENDING (no doctrine file touched in M1) |
| AC-WC-010 / AC-WC-011 (atomic save, GLM carve-out) | M3 / M2 | PENDING |
| AC-WC-012 (web/settings test fallout) | M7 | PENDING |
| AC-WC-013 (4-locale i18n) | M6 | PENDING |
| AC-WC-014 (make build — templ regenerate) | M5+ | PENDING (no templ edit in M1; `go build ./...` exit 0 confirms Go compile green) |
| AC-WC-015 (template-neutrality CI guard) | M2 | PASS (M2 — `grep` 0 SPEC-ID matches in baked sections) |
| AC-WC-018 (max/inherit neutral badge) | M5 | PENDING (UI layer; M1 table excludes sentinels per design.md §D) |
| AC-WC-019 (full-suite regression `go test ./...`) | M9 | PENDING |
| AC-WC-020 (render smoke) | M9 | PENDING |

## §E.3 Run-phase Audit-Ready Signal

**M1 + M2 + M3 partial** — multi-milestone Tier L SPEC; run-phase in-progress (M3 of M1–M9). This signal is completed at M9. The block below records the M1 + M2 + M3 portion.

```yaml
run_status: in-progress
run_complete_at: null   # pending — run-phase not complete (M3 of 9 milestones)
run_commit_sha: pending-backfill-m4   # latest run-phase commit (M4); M1=7b11d68bc, M2=e0061ad34, M3=cca120c70
m1_to_mN_commit_strategy: per-milestone commit + push to main (Route A — Hybrid Trunk 1-person OSS)
ac_pass_count: 11         # M1 (005,006,016,017,021) + M2 (003a,003b,015) + M3 (001,002) + M4 (004) = 11 PASS; AC-WC-022/023/024 partial (en-only staging)
ac_fail_count: 0
ac_pass_with_debt_count: 1   # M2 merge_method blocker (§D.Δ row 4 squash→Squash outside closed set) — debt, not fail
preserve_list_post_run_count: 20   # Option A — all 20 agent .md files under .claude/agents/{moai,harness}/ untouched
cross_platform_build:
  native: "go build ./... exit 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
coverage:
  v4manifest: "98.7% of statements"
  settings: "88.5% of statements"
new_warnings_or_lints_introduced: 0   # golangci-lint: 0 issues (baseline was also 0)
subagent_boundary: "grep -rn 'AskUserQuestion|mcp__askuser' internal/harness/v4manifest/ internal/settings/ | grep -v _test.go → 0 matches (C-HRA-008 clean)"
l44_pre_commit_fetch: "origin/main...HEAD divergence = 0 1 (local ahead by 1 — clean, no parallel race at push)"
l44_post_push_fetch: "3d535837b..7b11d68bc main -> main (push succeeded; Route A admin-bypass for branch protection + 5s pre-push warn — normal)"
milestones:
  M1:
    subject: "feat(SPEC-WEBCONF-SIMPLIFY-001): M1 sub-agent 4-color tier data-model + name-keyed lookup table"
    sha: 7b11d68bc
    acs: [AC-WC-005, AC-WC-006, AC-WC-016, AC-WC-017, AC-WC-021]
    status: PASS
  M2:
    subject: "feat(SPEC-WEBCONF-SIMPLIFY-001): M2 bake removed-tab defaults into template sections"
    sha: e0061ad34
    acs: [AC-WC-003a, AC-WC-003b, AC-WC-015]
    status: PASS-WITH-DEBT
    debt: "§D.Δ row 4 merge_method squash→Squash NOT applied — capitalized 'Squash' is outside validMergeMethods={squash,merge,rebase} (internal/config/validation.go:274). merge_method remains lowercase 'squash' (valid). BLOCKER returned for orchestrator/user resolution."
  M3:
    subject: "feat(SPEC-WEBCONF-SIMPLIFY-001): M3 remove 11 tabs from web UI + reclassify routes to RouteExcluded"
    sha: cca120c70
    acs: [AC-WC-001, AC-WC-002, AC-WC-010]
    status: PASS
  M4:
    subject: "feat(SPEC-WEBCONF-SIMPLIFY-001): M4 simplified surfaces + description rendering mechanism"
    sha: pending-backfill-m4
    acs: [AC-WC-004, AC-WC-024, AC-WC-022-partial, AC-WC-023-partial]
    status: PASS-WITH-DEBT
    debt: "AC-WC-022/023 (description mechanism) en-only — ko/ja/zh fieldDesc.* + agentfm descriptions + full Description population = M6"
  M5: { status: pending, owner: agentfm-ui-redesign }
  M6: { status: pending, owner: i18n-assets }
  M7: { status: pending, owner: test-fallout }
  M8: { status: pending, owner: docs-cleanup }
  M9: { status: pending, owner: final-verification }
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_

## §F Phase 4 Mode Selection

**Input parameters**: tier=L; scope≈20+ files (web handlers/schemaform/fieldsets/agentfm + settings schema/route/sections + 13 template YAML + 3 frontend assets + multiple tests); domains≥6 (web/settings/template/agents/assets/tests); language mix=Go+YAML+JS+CSS+templ; concurrency benefit=LOW (coding-heavy, net-new data-model concept at M1).

**Mode evaluation**:
- Mode 1 trivial: not selected — multi-file semantic change.
- Mode 2 background: not selected — write-capable implementation, not read-only.
- Mode 3 agent-team: RETIRED — never selected.
- Mode 4 parallel: not selected — coding-heavy (Anthropic coding-task parallelism caveat → Mode 5 over Mode 4).
- Mode 5 sub-agent: **SELECTED** — coding-heavy Tier L, sequential per-milestone delegation with verification gates.
- Mode 6 workflow: not selected — not mechanical-uniform (varied edits across web/settings/template/agents/assets; net-new concept needs per-milestone judgment).

**Decision**: sub-agent (Mode 5)

**Justification**: The SPEC is coding-heavy with a net-new data-model concept (M1 name-keyed tier table). Per Anthropic's coding-task parallelism caveat, Mode 5 sequential sub-agent is the correct default over Mode 4 parallel. Mode 6 admits only genuinely-parallel high-volume mechanical transforms; this SPEC's varied multi-domain edits + net-new concept need per-milestone judgment, so Mode 5. Route A (Hybrid Trunk main-direct) per the 1-person-OSS policy (CLAUDE.local.md §23) — manager-develop commits/pushes directly to main, no PR. Run-phase progression: M1 (foundational tier data-model) → verification gate → M2+M3+M4 (simplification core) → M5+M6 (tier UI) → M7+M8+M9 (finalization), 4 delegation chunks. User armed autonomous `/goal ac_converge` for continuation between chunks.
