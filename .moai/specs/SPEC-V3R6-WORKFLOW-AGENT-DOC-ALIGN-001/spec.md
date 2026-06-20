---
id: SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001
title: "Workflow + Agent + Doc Consistency Cleanup (archived-agent purge, broken cross-refs, stale ground-truth)"
version: "0.1.0"
status: completed
created: 2026-06-20
updated: 2026-06-20
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai, .claude/agents/moai, .claude/rules/moai, internal/template/templates/.claude"
# Note: all agent bodies (incl. builder-harness.md) live under .claude/agents/moai/. There is NO .claude/agents/builder/ directory (verified: ls .claude/agents/ → harness/, local/, moai/).
lifecycle: spec-anchored
tags: "archived-agent, doc-alignment, template-mirror, consistency-cleanup, workflow"
era: V3R6
---

# SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001 — Workflow + Agent + Doc Consistency Cleanup

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-20 | manager-spec | Initial plan-phase authoring. Derived from §C of the comprehensive `/moai` skill+workflow audit (3 parallel read-only audits, 2026-06-20). Scope: archived-agent spawn purge (Group 1), broken cross-references/invalid dispatch (Group 2), stale ground-truth content (Group 3), orphaned RED test fold-in (Group 4). |
| 0.2.0 | 2026-06-20 | manager-spec | plan-auditor iter-1 revision (FAIL 0.71 → re-audit). D1: corrected nonexistent `.claude/agents/builder/` → `.claude/agents/moai/` throughout. D2: removed `researcher` from the archived-agent set (it is a LIVE workflow.yaml role_profile, not an archived agent) + added live-role_profile KEEP carve-out. D3: re-baselined counts (was 27/125 — the 125 figure had silently excluded the 14 researcher false-positives) + added `references/{reference,anti-patterns,mx-tag}.md` to scope. D4: rewrote REQ-WADA-016 to verify changed-line mirroring (not whole-file byte parity) + excluded pre-existing lifecycle divergence in manager-spec/docs/git.md + aligned exclusion with `devOnlyLocalFiles` (release.md / github.md / release-update.md). D6: demoted `claude-code-guide` to a 0-occurrence note. D8: disambiguated Round-split (team/run.md:69) vs sprint-contract (workflows/run.md:80) paths in research.md. |
| 0.3.0 | 2026-06-20 | manager-spec | plan-auditor iter-2 revision (PASS-WITH-DEBT 0.84 → clean-PASS push). D-NEW-2: replaced the literal placeholder `'ARCHIVED'` in AC-WADA-001/006/007/016-Step2 grep blocks with the executable inline-expanded `ARCHIVED11='...'` + `grep -rnE "$ARCHIVED11"` form (literal `'ARCHIVED'` matched 0 files = false-pass trap). D-NEW-3: removed the no-op `n=$(grep -c ...)` snippet from AC-WADA-016 Step 3 (computed-but-unused; Step 2 already covers the 3 baseline-diverged bodies). D-NEW-1: corrected file count 32 → **38 matching (36 archived-proper + 2 researcher-only)** across all 5 artifacts (139/14/125 occurrence arithmetic was correct; only the file count was wrong). D-NEW-4: added scope note to AC-WADA-016 Step 1 that `TestTemplateMirrorParity` walks workflows/-only and Step 2 covers agents/team/references parity. |

## §A — Context & Provenance

This SPEC remediates the documentation/agent-consistency defects surfaced by a comprehensive `/moai` skill + workflow audit conducted via three parallel read-only audits on 2026-06-20. The audit's groups A (3 RED tests, 2 of which fixed in commit `f4f29279e`) and B (dead-code removal, also in `f4f29279e`) are already resolved. This SPEC owns the audit's **group C** (markdown/agent defects) plus the **orphaned RED test** left behind by the now-closed SPEC-V3R6-ORCH-IGGDA-001.

The MoAI agent catalog was consolidated from 17 agents to 8 retained (7 MoAI-custom + Anthropic built-in `Explore`) by SPEC-V3R6-AGENT-TEAM-REBUILD-001. Twelve agents were archived. The orchestrator now rejects any `Agent()` spawn naming an archived agent per `.claude/rules/moai/workflow/archived-agent-rejection.md §C`. However, **139 textual references matching the archived-name alternation survive across 38 files** (36 files match the 11 archived-proper names + 2 files match `researcher` only; skill files under `.claude/skills/moai/` + 6 agent bodies under `.claude/agents/moai/`), plus several broken cross-references and stale ground-truth values that predate the consolidation.

**IMPORTANT — `researcher` is NOT an archived agent (D2 correction)**: Of the 139 raw matches, **14 are legitimate `researcher` references** to the LIVE `role_profiles: researcher` profile in `.moai/config/sections/workflow.yaml` (verified at workflow.yaml:68, alongside analyst/architect/implementer/tester/designer/reviewer). Per NOTICE.md, the archived `researcher.md` MoAI agent file NEVER existed — `researcher` was an "originally absent" archive entry, so there is no agent file to purge and the role_profile references are CORRECT and MUST be preserved. The archived-agent-proper purge set is therefore the **11 actually-archived agents** (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide` [MoAI-custom file, 0 live occurrences — see §C REQ-WADA-001 note], `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`), NOT 12.

### §A.1 — Critical constraint: Template-First mirroring

Per CLAUDE.local.md §2 (Template-First Rule) and §24 (Harness Namespace), most affected skill/agent/rule files exist in BOTH the live `.claude/` tree AND the template source `internal/template/templates/.claude/`. The template source is the single source of truth (SSOT) for MIRRORED files. EVERY fix to a mirrored file MUST be applied to both locations (changed-line mirroring per REQ-WADA-016), and acceptance criteria MUST verify both mirrors. Verified at plan-phase: all in-scope skill/agent/rule files have a confirmed `internal/template/templates/.claude/` mirror **EXCEPT** the dev-only `devOnlyLocalFiles` set (`release.md`, `github.md`, `release-update.md`) which intentionally have NO mirror — their archived-purge applies to the LIVE file only.

Namespace boundary (§24): `.claude/agents/moai/` + `.claude/skills/moai/` are template-managed (sync overwrites). `.claude/agents/harness/` + `.claude/agents/local/` are user-owned — `.claude/agents/local/github-specialist.md` is dev-only and is NOT template-mirrored, so it is OUT OF SCOPE for the template-mirror verification (its archived-agent references, if any, are a local maintainer concern outside this SPEC's template-neutrality obligation).

### §A.2 — Evidence baseline (plan-phase grep, against working tree)

The alternation `ARCHIVED-11` below = `manager-strategy\|manager-quality\|manager-brain\|manager-project\|claude-code-guide\|expert-backend\|expert-frontend\|expert-security\|expert-devops\|expert-performance\|expert-refactoring` (the 11 actually-archived names — `researcher` is EXCLUDED because it is a live role_profile, see §A context). The raw alternation `ARCHIVED-12` adds `researcher` and is used only to show the total-with-false-positives count.

| Defect group | Evidence command | Observed at plan-phase |
|--------------|------------------|------------------------|
| G0 directory truth | `ls .claude/agents/` | `harness/ local/ moai/` — there is NO `builder/` dir; builder-harness.md is at `.claude/agents/moai/builder-harness.md` |
| G1 raw total (with researcher false-positives) | `grep -rn 'ARCHIVED-12' .claude/skills/moai/ .claude/agents/moai/` | 139 lines |
| G1 raw file count | `grep -rlnE 'ARCHIVED-12' .claude/skills/moai/ .claude/agents/moai/ \| wc -l` | 38 files (36 archived-proper + 2 researcher-only) |
| G1 researcher false-positives (KEEP) | `grep -rcn 'researcher' .claude/skills/moai/ .claude/agents/moai/` (summed) | 14 lines (live role_profile refs — NOT purged) |
| G1 archived-proper total (purge target) | `grep -rnE 'ARCHIVED-11' .claude/skills/moai/ .claude/agents/moai/` | 139 − 14 = 125 archived-proper occurrences across 36 files (the true purge count) |
| G1 archived refs (agents) | `grep -rln 'ARCHIVED-11' .claude/agents/moai/` | 6 files: builder-harness, manager-develop, manager-docs, manager-git, manager-spec, sync-auditor |
| G1 references/ files (newly added to scope) | `grep -rln 'ARCHIVED-11' .claude/skills/moai/references/` | reference.md, anti-patterns.md, mx-tag.md (carry archived refs incl. frontmatter `agents:` lists) |
| G2 team agents | `grep -rn 'team-reader\|team-validator' .claude/skills/moai/team/` | plan.md, review.md, debug.md, glm.md |
| G2 worktree glossary | `grep -n 'Terminology Glossary' worktree-integration.md` | 0 (section absent; 2 files cross-reference it) |
| G3 GLM models | `grep -n 'glm-5.1\|glm-4.7\|glm-4.5-air\|glm-4.7-flashx' team/glm.md` | lines 148-150, 168-172 |
| G3 retired terms | `grep -n 'Round split' team/run.md` + `grep -n 'sprint contract' workflows/run.md` | `team/run.md:69` (Round split) + `workflows/run.md:80` (sprint contract) — two DIFFERENT files (D8) |
| G3 SKILL.md version | `tail -30 SKILL.md \| grep -i version` | `Version: 2.6.0 / Last Updated: 2026-02-25` (body footer; SKILL.md frontmatter has no version field) |
| G4 RED test | `go test ./internal/skills/ -run TestEntryRouterLOCCeiling` | FAIL: run.md 246 LOC > 200 ceiling |
| D4 baseline divergence | `diff .claude/agents/moai/<f> internal/template/templates/.claude/agents/moai/<f>` for manager-spec/docs/git.md | ALL 3 already DIVERGED at baseline (pre-existing §E.5/Mx lifecycle divergence owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001 — OUT OF SCOPE here; whole-file byte parity is unsatisfiable for these 3) |
| D4 dev-only no-mirror | `ls internal/template/templates/.claude/skills/moai/workflows/release.md` | absent — `release.md` is dev-only (in `devOnlyLocalFiles`: release.md, github.md, release-update.md); its archived-purge applies to the LIVE file ONLY, no mirror obligation |

## §B — Goal

Restore consistency between the `/moai` skill+workflow+agent assets and the canonical 8-agent catalog + current ground-truth, applying every fix to both the live `.claude/` tree and the template SSOT, and restoring the test suite to GREEN by trimming `run.md` under its 200-LOC entry-router ceiling without losing routing logic.

## §C — GEARS Requirements

### Group 1 — Archived-agent spawn purge (HIGH)

- **REQ-WADA-001 (Ubiquitous)**: The `/moai` skill+workflow asset set shall not name any of the **11 actually-archived agents** (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide` [MoAI-custom file — 0 live occurrences, see REQ-WADA-001 note], `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`) as an `Agent()` spawn target or delegation owner in any template-managed file under `.claude/skills/moai/` or `.claude/agents/moai/`. (There is NO `.claude/agents/builder/` directory; builder-harness.md lives under `.claude/agents/moai/`.)
  - **KEEP carve-out — `researcher` live role_profile (REQ-WADA-001a, Ubiquitous)**: The asset set SHALL preserve every reference to the LIVE workflow role_profiles — `researcher`, `analyst`, `architect`, `implementer`, `tester`, `designer`, `reviewer` (defined in `.moai/config/sections/workflow.yaml role_profiles:`). These are NOT archived agents; they are dynamic-team spawn profiles used by `team/plan.md`, `team/run.md`, and the spec-workflow Team-Mode tables. Purging `researcher` would break team-mode plan-phase spawn. The run-phase MUST distinguish a `researcher` role_profile reference (KEEP) from a (nonexistent) `researcher` archived-agent spawn (there are none) — all 14 `researcher` occurrences are role_profile references and MUST be left untouched.
  - **`claude-code-guide` note (REQ-WADA-001b)**: `claude-code-guide` has **0 live occurrences** in the asset set (verified `grep -rn 'claude-code-guide' .claude/skills/moai/ .claude/agents/moai/` → 0). No purge action is required for it. The built-in disambiguation (the valid Anthropic built-in distinct from the archived MoAI file) is preserved per `archived-agent-rejection.md §C.3`; legitimate references TO that rule are NOT removed.

- **REQ-WADA-002 (Event-driven)**: When an audited file delegates planning/strategy work formerly routed to `manager-strategy`, the file shall route that work to `manager-spec` per the `archived-agent-rejection.md §C` migration table (row 1).

- **REQ-WADA-003 (Event-driven)**: When an audited file delegates quality-gate/review work formerly routed to `manager-quality`, the file shall route that work to `sync-auditor` (independent quality scoring) OR an orchestrator read-only verification batch (lint + test + coverage) per the migration table (row 2).

- **REQ-WADA-004 (Event-driven)**: When an audited file delegates domain-specialist work formerly routed to an `expert-*` agent, the file shall route that work to a per-spawn `Agent(general-purpose)` invocation carrying the domain whitelist + domain instructions per the migration table (rows 7-12).

- **REQ-WADA-005 (Event-driven)**: When `feedback.md` (and any file whose sole archived executor was `manager-quality` for GitHub-issue creation) is remediated, the file shall route GitHub-issue creation to the orchestrator-direct `gh` CLI path (the migration table provides no retained-agent owner for issue creation), preserving the existing issue-creation behavior.

- **REQ-WADA-006 (Ubiquitous)**: The agent bodies (`manager-develop`, `manager-spec`, `manager-docs`, `manager-git`, `builder-harness`, `sync-auditor`) shall not reference archived agents in their "Delegation Protocol" / "OUT OF SCOPE" / cross-reference sections except where the reference is an explicit pointer TO the `archived-agent-rejection.md` migration table (a legitimate documentation reference, not a spawn instruction).

- **REQ-WADA-007 (Ubiquitous)**: Skill/command frontmatter `agents:` (or `triggers.agents:`) lists shall not enumerate archived agents; each archived entry is removed or replaced with its canonical retained replacement.

### Group 2 — Broken cross-references & invalid dispatch (HIGH)

- **REQ-WADA-008 (Event-driven)**: When the `moai` skill dispatches a sub-workflow for the release path, it shall reference a workflow file that exists (`workflows/release.md`); no dispatch shall point at a non-existent `workflows/release-update.md`. (Plan-phase note: the `release-update` dispatch token was NOT found in the current `SKILL.md` working-tree state — see §G OQ-1; this REQ is conditional on its presence at run-phase.)

- **REQ-WADA-009 (Ubiquitous)**: The team skill files (`team/plan.md`, `team/review.md`, `team/debug.md`, `team/glm.md`) shall use `subagent_type: "general-purpose"` (the pattern established in `team/run.md`) rather than the non-existent `team-reader` / `team-validator` subagent types.

- **REQ-WADA-010 (Event-driven)**: When `spec-workflow.md` or `worktree-state-guard.md` cross-reference `worktree-integration.md § Terminology Glossary` for L1/L2/L3 layer definitions, the referenced `§ Terminology Glossary` section shall exist in `worktree-integration.md` (add the section) OR the cross-references shall be repointed to an existing section that defines L1/L2/L3 (design.md decides the direction).

### Group 3 — Stale ground-truth content (MED)

- **REQ-WADA-011 (Ubiquitous)**: The `team/glm.md` GLM model-name references shall reflect the current ground truth (`glm-5.2[1m]`) rather than the stale `glm-5.1` / `glm-4.7` / `glm-4.5-air` / `glm-4.7-flashx` values.

- **REQ-WADA-012 (Ubiquitous)**: Audited files shall use the canonical Epic/Milestone taxonomy per `.claude/rules/moai/development/sprint-round-naming.md`; retired terms ("Round split", "sprint contract") used as canonical grouping terms shall be replaced (Milestone / harness contract or equivalent canonical phrasing).

- **REQ-WADA-013 (Ubiquitous)**: The `SKILL.md` body version footer shall reflect a version/date consistent with the post-8-agent-consolidation state (bump from `2.6.0` / `2026-02-25`).

### Group 4 — Orphaned RED test fold-in

- **REQ-WADA-014 (Event-driven)**: When `run.md` is modified for the Group 1 archived-agent purge, the file shall be trimmed below the 200-LOC entry-router ceiling enforced by `internal/skills` `TestEntryRouterLOCCeiling` (`workflow_split_test.go:154`) without losing the entry-router argument-branching routing logic.

- **REQ-WADA-015 (Event-driven)**: When the run-phase completes, `go test ./internal/skills/ -run TestEntryRouterLOCCeiling` shall exit 0 (the suite shall be GREEN for this test).

### Cross-cutting — Template mirror

- **REQ-WADA-016 (Ubiquitous)**: Every change THIS SPEC makes to a file under `.claude/skills/moai/`, `.claude/agents/moai/`, or `.claude/rules/moai/` that HAS a template mirror shall be applied identically to its `internal/template/templates/.claude/...` mirror — i.e., the **specific lines this SPEC changes** shall be present and identical in both trees after run-phase. This is a **changed-line mirror requirement, NOT a whole-file byte-parity requirement**, because three agent bodies (`manager-spec.md`, `manager-docs.md`, `manager-git.md`) are ALREADY DIVERGED live↔template at baseline on the 3-phase-vs-legacy `§E.5`/Mx lifecycle content (owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001, OUT OF SCOPE here). For those three files, the run-phase verifies ONLY that this SPEC's archived-purge edit landed in BOTH trees; it does NOT attempt to reconcile the pre-existing lifecycle divergence. Dev-only files with no template mirror (`devOnlyLocalFiles`: `release.md`, `github.md`, `release-update.md`) are EXCLUDED from the mirror requirement entirely — their archived-purge applies to the LIVE file only.

- **REQ-WADA-017 (Ubiquitous)**: The template source shall remain content-neutral per CLAUDE.local.md §25 — the replacement text introduced by this SPEC shall not leak internal SPEC IDs, REQ tokens, audit citations, internal dates, or commit SHAs into `internal/template/templates/`.

## §D — Acceptance Criteria Reference

See `acceptance.md` for the full AC matrix (AC-WADA-001 … AC-WADA-017) with verifiable grep/test commands proving both live + template mirrors are fixed.

## §E — Scope Boundaries (What/Why, not How)

This SPEC defines WHAT must be consistent (no archived spawn targets, valid cross-references, current ground-truth, GREEN test) and WHY (orchestrator rejects archived spawns; broken refs misroute work; stale values mislead; RED test blocks main). It does NOT prescribe the exact replacement prose for each of the ~125 archived-proper occurrences (139 raw − 14 researcher role_profile KEEPs) — that is a run-phase implementation decision constrained by the migration table.

## §F — Exclusions

The following are explicitly out of scope for this SPEC. Each excluded item is a routing decision (it belongs elsewhere) or a deliberate deferral.

### Out of Scope — Run-phase code & dead-code removal

- The audit's Group A (3 RED test fixtures) and Group B (~600 LOC dead-code removal) are already resolved in commit `f4f29279e` and are NOT re-touched here.
- No production Go code under `internal/` or `pkg/` is modified by this SPEC EXCEPT the unavoidable consequence of trimming `run.md` (a markdown asset, not Go) to satisfy `TestEntryRouterLOCCeiling`. The test itself is NOT modified — the fix is to the markdown asset it measures.

### Out of Scope — User-owned & dev-only assets

- `.claude/agents/local/github-specialist.md` (dev-only, not template-mirrored) is NOT remediated for archived-agent references — it is a local maintainer asset outside the template-neutrality obligation.
- `.claude/agents/harness/**` and `.claude/skills/harness-*/**` (user-owned per §24) are NOT touched.
- `.claude/commands/97-*.md`, `98-*.md`, `99-*.md` (dev-only maintainer commands per CLAUDE.local.md §21) are NOT remediated.

### Out of Scope — Dev-only no-mirror files & pre-existing template divergence

- The dev-only `workflows/release.md` (and `github.md`, `release-update.md` — the `devOnlyLocalFiles` set in `internal/skills/workflow_split_test.go:184`) have NO template mirror by design. Their archived-agent purge applies to the LIVE file ONLY; they are EXCLUDED from REQ-WADA-016 mirror verification and from the `TestTemplateMirrorParity` exclusion set.
- The pre-existing live↔template divergence in `manager-spec.md` / `manager-docs.md` / `manager-git.md` (3-phase-vs-legacy `§E.5`/Mx lifecycle content) is owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001 and is NOT reconciled by this SPEC. This SPEC only verifies its own archived-purge edits landed in both trees for those three files (per REQ-WADA-016).

### Out of Scope — Mechanical-detection layer

- No new Go lint rule or CI guard to mechanically PREVENT future archived-agent references is created by this SPEC. This SPEC remediates the current textual state; a forward-looking mechanical detector (e.g., an `ArchivedAgentReference` lint rule) is a separate future SPEC.

### Out of Scope — Archived-agent revival

- No archived agent is revived, un-archived, or re-introduced. The `archived-agent-rejection.md` migration table and the `claude-code-guide` built-in disambiguation (a valid Anthropic built-in distinct from the archived MoAI file) are preserved verbatim — legitimate references TO that rule are NOT removed.

### Out of Scope — Behavior changes

- This is a documentation/consistency cleanup. No workflow behavior, agent capability, or routing semantics changes EXCEPT correcting routing that was already broken (e.g., `feedback.md` whose sole executor was an archived agent). No new features are added.
