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

**M1-only evidence** (multi-milestone Tier L SPEC; M2–M9 append their own evidence in later delegation chunks). The ACs verified below are the M1-relevant tier-model backend ACs — the name→tier lookup table, the accessors, and the Option-A preservation assertions. All other ACs (tab reduction M3, template baking M2, UI M5, i18n M6, test fallout M7, docs M8, full-suite M9) are pending their owning milestone and enumerated at the bottom of this section.

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

### Non-M1 ACs (pending their owning milestone)

| AC ID | Owning milestone | Status |
|-------|------------------|--------|
| AC-WC-001 / AC-WC-002 (6-tab render, removed tabs absent) | M3 | PENDING |
| AC-WC-003a / AC-WC-003b / AC-WC-004 (config keys persist, quality_extras toggle) | M2 / M4 | PENDING |
| AC-WC-007 (tier-click auto-suggest writes via agentfm.Patch) | M5 | PENDING (backend accessor `TierSuggestedModelEffort` ready; UI wiring is M5) |
| AC-WC-008 (closed-set validation) | M5 | PENDING (closed sets `V4EffortValues`/`V4ModelValues` already exist; validator wiring is M5) |
| AC-WC-009 (deep doctrine preserved) | M8 | PENDING (no doctrine file touched in M1) |
| AC-WC-010 / AC-WC-011 (atomic save, GLM carve-out) | M3 / M2 | PENDING |
| AC-WC-012 (web/settings test fallout) | M7 | PENDING |
| AC-WC-013 (4-locale i18n) | M6 | PENDING |
| AC-WC-014 (make build — templ regenerate) | M5+ | PENDING (no templ edit in M1; `go build ./...` exit 0 confirms Go compile green) |
| AC-WC-015 (template-neutrality CI guard) | M2 | PENDING (no template edit in M1) |
| AC-WC-018 (max/inherit neutral badge) | M5 | PENDING (UI layer; M1 table excludes sentinels per design.md §D) |
| AC-WC-019 (full-suite regression `go test ./...`) | M9 | PENDING |
| AC-WC-020 (render smoke) | M9 | PENDING |

## §E.3 Run-phase Audit-Ready Signal

**M1 partial** — multi-milestone Tier L SPEC; run-phase in-progress (M1 of M1–M9). This signal is completed at M9. The block below records the M1 portion.

```yaml
run_status: in-progress
run_complete_at: null   # pending — run-phase not complete (M1 of 9 milestones)
run_commit_sha: pending-backfill-m1   # self-referential — backfilled in a follow-up commit (D3 placeholder exemption)
m1_to_mN_commit_strategy: per-milestone commit + push to main (Route A — Hybrid Trunk 1-person OSS)
ac_pass_count: 5          # M1-relevant ACs PASS: AC-WC-005, 006, 016, 017, 021
ac_fail_count: 0
preserve_list_post_run_count: 20   # Option A — all 20 agent .md files under .claude/agents/{moai,harness}/ untouched
cross_platform_build:
  native: "go build ./... exit 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
coverage:
  v4manifest: "98.7% of statements"
  settings: "88.5% of statements"
new_warnings_or_lints_introduced: 0   # golangci-lint: 0 issues (baseline was also 0)
subagent_boundary: "grep -rn 'AskUserQuestion|mcp__askuser' internal/harness/v4manifest/ internal/settings/ | grep -v _test.go → 0 matches (C-HRA-008 clean)"
l44_pre_commit_fetch: pending   # pre-push git fetch result recorded post-push
l44_post_push_fetch: pending
milestones:
  M1:
    subject: "feat(SPEC-WEBCONF-SIMPLIFY-001): M1 sub-agent 4-color tier data-model + name-keyed lookup table"
    sha: pending-backfill-m1
    acs: [AC-WC-005, AC-WC-006, AC-WC-016, AC-WC-017, AC-WC-021]
    status: PASS
  M2: { status: pending, owner: template-defaults-baking }
  M3: { status: pending, owner: tab-reduction }
  M4: { status: pending, owner: surviving-tab-surfaces }
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
