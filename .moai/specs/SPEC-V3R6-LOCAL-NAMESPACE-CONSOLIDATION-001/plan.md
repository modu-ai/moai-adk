---
id: SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001
title: "Local Agent Namespace Consolidation — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.7.0"
module: ".claude/agents/local + .claude/skills/moai/workflows + internal/template/templates + .moai/docs"
lifecycle: spec-anchored
tags: "local-namespace, dev-only, agent-migration, template-refactor, claude-local-externalization, sprint-10-lane-b, thin-command-pattern"
depends_on: []
related_specs: []
---

# Plan — SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001

## A. Phase Lifecycle Table

| Phase | Owner | Status | Commit SHA | Notes |
|-------|-------|--------|------------|-------|
| Plan | manager-spec | in-progress | `<pending>` | This document; Phase 0.5 plan-auditor gate pending |
| Run | manager-develop | not-started | `<pending>` | M1-M6 expected; trunk-direct (no PR) per Hybrid Trunk policy |
| Sync | manager-docs | not-started | `<pending>` | CHANGELOG + frontmatter status → implemented |
| Mx | manager-docs or orchestrator | not-started | `<pending>` | EVALUATE-SKIP likely per mx-tag-protocol.md §a (markdown-only changes, 0 .go files) |

## B. Known Issues

(Populated by plan-auditor at Phase 0.5 audit gate. Per max-3 iteration contract, defects with `MUST-FIX` severity block run-phase entry; `SHOULD-FIX` defects are tracked here for iter-N resolution.)

| Iteration | Severity | Defect ID | Description | Resolution |
|-----------|----------|-----------|-------------|------------|
| (pending) | — | — | — | — |

## C. Milestone Breakdown

### M1 — Namespace Contract Documentation Update

Scope: Update three SSOT documents to register `.claude/agents/local/` as a user-owned (PRESERVE-list) namespace alongside the existing `.claude/agents/harness/`.

Files modified (3):
- `.claude/rules/moai/development/agent-authoring.md` — Add `.claude/agents/local/` row to the Agent Directory Convention table (between `meta/` row and `harness/` row, alphabetic and conceptual ordering). Update the [HARD] rule block to list local/ alongside harness/ in the PRESERVE clauses.
- `.claude/rules/moai/development/skill-authoring.md` — No structural namespace addition here (skills do not have a local/ namespace); but update the cross-reference section to mention the new local agent namespace as a related pattern.
- `.moai/docs/dev-only-commands-isolation.md` — Add agent-local-namespace verification checklist entries: `find internal/template/templates -path "*/agents/local/*"` returns empty (HARD), `find internal/template/templates -name "release-update-specialist.md"` returns empty, `find internal/template/templates -name "github-specialist.md"` returns empty. Update the "배포 금지 파일 목록" table with the two new agent body file rows.

Mirror requirement: same 3 paths under `internal/template/templates/.claude/rules/moai/development/` + `internal/template/templates/.moai/docs/` need parallel updates. Total file count for M1: 6 files (3 local + 3 mirror).

Acceptance verification (AC-LNC-001, AC-LNC-008, AC-LNC-011).

### M2 — Local Agent Body Authoring

Scope: Create two new agent body files under `.claude/agents/local/` and mirror the same two files (which is forbidden — see paragraph after this) — clarification: agent body files are USER-OWNED, mirror to `internal/template/templates/.claude/agents/local/` is PROHIBITED per REQ-LNC-012.

Files created (2, local-only — NO template mirror):
- `.claude/agents/local/release-update-specialist.md` — YAML frontmatter (name, description, tools, model, color, effort) + agent body containing the 8-phase CC upstream tracker workflow migrated verbatim from `.claude/skills/moai/workflows/release-update.md` (lines 34 onward). Approximate body LOC: 600 (matches predecessor skill body Phase 0-8 + Agent Delegation Map + Output Artifacts + Verification Gate + Anti-Patterns + References).
- `.claude/agents/local/github-specialist.md` — YAML frontmatter + agent body containing the GitHub issue/PR Agent Teams workflow migrated from `.claude/skills/moai/workflows/github.md` (approximate body LOC: 580).

Total LOC for M2: ~1184 LOC across 2 files.

Mirror requirement: NONE per REQ-LNC-012. Adding `internal/template/templates/.claude/agents/local/*.md` files at M2 is a HARD violation triggering AC-LNC-009 failure.

Acceptance verification (AC-LNC-003, AC-LNC-004, AC-LNC-009).

### M3 — Dev-Only Skill Removal + Thin Command Rewiring

Scope: Remove the two predecessor skill body files and update the two thin command wrappers to delegate to the new local agents.

Files deleted (2):
- `.claude/skills/moai/workflows/release-update.md`
- `.claude/skills/moai/workflows/github.md`

Files modified (2):
- `.claude/commands/97-release-update.md` — YAML frontmatter `allowed-tools: Skill` → `allowed-tools: Agent`. Body line changes from `Use Skill("moai") with arguments: release-update $ARGUMENTS` to `Use the release-update-specialist subagent with arguments: $ARGUMENTS`. Body remains 1 line. Total LOC unchanged at 9.
- `.claude/commands/98-github.md` — Same pattern: `allowed-tools: Skill` → `allowed-tools: Agent`. Body line changes from `Use Skill("moai/workflows/github") with arguments: $ARGUMENTS` to `Use the github-specialist subagent with arguments: $ARGUMENTS`. Body remains 1 line. Total LOC unchanged at 9.

Mirror requirement: NONE. Per CLAUDE.local.md §2 Local-Only Files list, both `.claude/commands/97-*` and `.claude/commands/98-*` are dev-only and have never been under `internal/template/templates/`. Verification: `find internal/template/templates -name "97-*" -o -name "98-*"` returns empty BOTH before and after M3.

Acceptance verification (AC-LNC-002, AC-LNC-005, AC-LNC-006).

### M4 — Template Generic Refactor (Leak Removal)

Scope: Eliminate all 17 `CLAUDE.local.md` cross-references from approximately 13 files under `internal/template/templates/`. Per-leak rewrite strategy documented inline below.

Files modified (13, all under `internal/template/templates/`):

| File (template-side path) | Line(s) | Replacement strategy |
|---------------------------|---------|----------------------|
| `.claude/rules/moai/core/agent-common-protocol.md` | 339 | Rewrite race mitigation cross-ref to point at `.moai/docs/generic-patterns-guide.md` § Multi-Session Race Mitigation (W5 deliverable). |
| `.claude/rules/moai/development/agent-authoring.md` | 34 | Replace `CLAUDE.local.md §24.2 + §24.4` with same-file `§ Agent Directory Convention` and `.claude/skills/moai-meta-harness/SKILL.md § Namespace Separation` (both already cited verbatim elsewhere in the body). |
| `.claude/rules/moai/development/branch-origin-protocol.md` | 73 | Replace `CLAUDE.local.md §18.12 — dev-project specific notes` with generic `(see project-local maintenance documentation if applicable; stacked PR Case Study reference is §18.11)` — the §18.11 ref was already orphaned; reword to remove both. |
| `.claude/rules/moai/development/skill-authoring.md` | 264, 282, 305 | All three are `§15 / §24` language-neutrality + harness-namespace cross-refs. Replace with self-references to the same file's § Language Guidance Lives in Rules + § Skills Namespace Policy sections (which already contain the canonical content). |
| `.claude/rules/moai/workflow/moai-memory.md` | 17 | The line lists CLAUDE.local.md in a 5-level file inventory ("Local Instructions: CLAUDE.local.md (personal project, not committed)"). Rewrite to a generic 4-level inventory or note as "Optional local instructions file (e.g., CLAUDE.local.md if used; not committed)". |
| `.claude/output-styles/moai/moai.md` | 426, 458, 707 | All three are race-mitigation + namespace cross-refs. Rewrite to point at `.moai/docs/generic-patterns-guide.md` § Multi-Session Race Mitigation (lines 426, 458) and `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention (line 707). |
| `.claude/skills/moai/workflows/loop.md` | 215 | Language neutrality §22 cross-ref. Replace with `.claude/rules/moai/development/coding-standards.md` § Language Policy (which already contains the 16-language neutrality rule). |
| `.claude/skills/moai/workflows/project/doc-generation.md` | 139 | Same language-neutrality replacement as loop.md line 215. |
| `.claude/skills/moai-workflow-loop/SKILL.md` | 188 | Same language-neutrality replacement. |
| `.claude/skills/moai-workflow-loop/references/reference.md` | 789 | Same language-neutrality replacement. |
| `.claude/skills/moai-workflow-loop/references/examples.md` | 510 | Same language-neutrality replacement. |
| `.claude/skills/moai-meta-harness/SKILL.md` | 180 | Harness namespace §24 cross-ref. Replace with same-file § Namespace Separation (which is the canonical generator-side rule). |
| `.moai/config/sections/lsp.yaml.tmpl` | 4 | Language neutrality §22 cross-ref in a YAML comment. Replace with reference to `.claude/rules/moai/development/coding-standards.md` § Language Policy. |

Mirror requirement: same paths in local `.claude/` and `.moai/` need parallel updates so local copy matches template (Template-First Rule + immediate local sync per CLAUDE.local.md §2 Development Cycle). Total file count for M4: 26 files (13 template + 13 local mirror).

Note on local M4 mirrors: some local files under `.claude/` may have additional `CLAUDE.local.md` references that are valid in local context (because CLAUDE.local.md does exist in the maintainer's local project). M4 LOCAL scope only modifies the lines that correspond to the same 17 template-side leak sites — preserves local-only doctrinal refs.

Acceptance verification (AC-LNC-007, AC-LNC-010).

### M5 — Generic Patterns Guide Authoring (W5)

Scope: Create the new externalized generic patterns guide at `.moai/docs/generic-patterns-guide.md` (local) + `internal/template/templates/.moai/docs/generic-patterns-guide.md` (template).

Files created (2 = 1 local + 1 template mirror):
- `internal/template/templates/.moai/docs/generic-patterns-guide.md` — Template-First Rule compliance (CLAUDE.local.md §2). YAML-less Markdown body. Approximate body LOC: 250-400 covering 4 sections:
  1. Multi-Session Race Mitigation Procedure (generalized from CLAUDE.local.md §23.8)
  2. Hook Setup Procedure for New Machines (generalized from §23.1)
  3. Settings Intent Doctrine (generalized from §22)
  4. Late-Branch Phase D Recovery Procedure (generalized from §23.6)
- `.moai/docs/generic-patterns-guide.md` — Local mirror after `make build` regenerates embedded files.

Per-section content discipline: each section presents the pattern in user-audience prose. Maintainer-specific elements (e.g., 1-person OSS Hybrid Trunk, `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/` paths, `moai cg`/`moai glm` commands) are reframed as "if you adopt this policy" / "your project hash" / "your team's CG mode if configured" placeholders.

Acceptance verification (AC-LNC-010 implicit, content quality verified at sync-phase).

### M6 — Progress.md Backfill + Cross-Reference Audit

Scope: Update `.moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/progress.md` with M1-M5 status (in-progress → done as each milestone completes) and decision-log entries. Execute orchestrator-side verification batch to confirm all acceptance criteria pass before handoff to manager-docs for sync-phase.

Files modified (1):
- `.moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/progress.md` — §B Milestone Status table rows M1-M5 marked done; §C Decision Log appended with run-phase decisions; §E Run-phase Evidence section authored with verification command outputs.

Verification batch (run as a single multi-Bash call per `agent-common-protocol.md` §Parallel Execution):
```bash
# 1. Template leak elimination
grep -rln "CLAUDE.local.md\|CLAUDE\.local" internal/template/templates/

# 2. Agent local namespace presence (local-side only)
ls -la .claude/agents/local/

# 3. Template local namespace ABSENCE (REQ-LNC-012)
find internal/template/templates -path "*/agents/local/*"

# 4. Thin command pattern compliance
wc -l .claude/commands/97-release-update.md .claude/commands/98-github.md

# 5. Dev-only skill removal
ls -la .claude/skills/moai/workflows/release-update.md .claude/skills/moai/workflows/github.md 2>&1

# 6. Generic patterns guide presence (both local and template)
ls -la .moai/docs/generic-patterns-guide.md internal/template/templates/.moai/docs/generic-patterns-guide.md

# 7. Full Go test suite (commands_audit_test.go in particular)
go test ./internal/template/...
```

Acceptance verification (all AC-LNC-001 through AC-LNC-011 final pass).

## D. Dependencies

- No SPEC-level dependencies (depends_on: []). The three predecessor scopes were prior-session investigations not formalized as SPECs.
- Implicit dependency on Thin Command Pattern doctrine in `.claude/rules/moai/development/coding-standards.md` lines 56-77 (must not regress).
- Implicit dependency on Template-First Rule (CLAUDE.local.md §2) and namespace separation contract (CLAUDE.local.md §24).
- No build-time dependency on `internal/cli/update.go` PRESERVE-list code path (deferred to follow-up SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001).

## E. Audit-Ready Signals

| Phase | Owner | Audit-Ready signal | Status |
|-------|-------|--------------------|--------|
| Plan | manager-spec | Phase 0.5 plan-auditor verdict ≥ 0.80 (Tier M PASS threshold) | `<pending>` |
| Run | manager-develop | All 7 verification batch commands return expected output; 11/11 AC-LNC PASS | `<pending>` |
| Sync | manager-docs | CHANGELOG entry created, frontmatter status `draft → implemented`, sync_commit_sha backfilled in all 4 artifacts | `<pending>` |
| Mx | manager-docs or orchestrator | Mx Step C EVALUATE-SKIP judgment per markdown-only criteria (0 .go files, 0 goroutines, 0 fan_in changes), `<moai>DONE</moai>` emit | `<pending>` |

## F. TRUST 5 Mapping

| Pillar | Application to this SPEC |
|--------|--------------------------|
| Tested | Markdown-only changes; verification via grep + ls + wc + go test ./internal/template/... batch. No new test files needed (existing commands_audit_test.go covers Thin Command Pattern regression). |
| Readable | All 13 REQ-LNC + 11 AC-LNC use GEARS notation per skill-authoring.md § GEARS-discipline; HISTORY tables in all 4 artifacts; cross-references resolve to canonical SSOT locations. |
| Unified | Single SPEC consolidates 3 scopes (W3-arch + W4 + W5); single CHANGELOG entry; single sync-phase frontmatter status transition. |
| Secured | No secrets, credentials, or auth code modified. Dev-only namespace separation enhances security boundary (maintainer-only agents not exposed to user projects). |
| Trackable | SPEC frontmatter 12-canonical-field validated; conventional commit subject pattern per Status Transition Ownership Matrix in spec-frontmatter-schema.md (plan-phase commit subject: `feat(SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001): plan-phase artifacts (Tier M Section A-G, 4 artifacts)`). |
