---
id: SPEC-GITFLOW-DOCTRINE-ALIGN-001
title: "Align the bodies of three local-only git-doctrine documents with the 2026-08-27 GitHub Flow -> git-flow transition"
version: "0.1.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: ".moai/docs"
lifecycle: spec-anchored
tags: "git-flow, doctrine-alignment, documentation-only, develop-integration, pr-policy, local-only"
tier: S
related_specs: [SPEC-V3R6-AGENT-TEAM-REBUILD-001]
---

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-08-27 | manager-spec | Initial creation — plan-phase artifacts for kanban card t310 (Class C, Tier S). Scope fixed by the lane dispatch: repair body inconsistencies left in three local-only doctrine documents at the 2026-08-27 GitHub Flow -> git-flow transition. Defect inventory measured on this tree (base d29b8942e); see research.md. |
| 2026-08-27 | manager-spec | Plan-audit iter-1 amendment (verdict PASS 0.85; report `.moai/reports/t310/plan-audit-iter1.md`) — D1: REQ-GDA-004 coverage extended to the §23.7 line-151 always-PR route-prescription clause; D2: marker-census command fixed to `grep -F` form; D3: AC-GDA-007 red-count restated (branch-model mentions vs incidental identifier); D4: M1 draft backtick typo corrected; D5: closure-gate token de-literalized. Research inventory gains D9 (L151) and F-7 (title framings, consciously excluded). |

## §A Context and Problem

On 2026-08-27 the operator directed this repository to switch its development model from GitHub Flow to git-flow. The canonical model is written in exactly two places, neither of which this SPEC touches:

1. `CLAUDE.local.md` §4.1 (model + rationale)
2. `.claude/rules/local/gitflow-lane-protocol.md` (lane operating rules)

At transition time, THREE downstream local-only doctrine documents received only ~7-line header notices stating that the new model supersedes parts of what follows — but their BODIES were left asserting the OPPOSITE of the new model. A reader who greps a section without reading the header receives inverted [HARD] rules:

- "create a `develop` branch" listed as FORBIDDEN (it is now mandatory for card work),
- "ALL tiers open PRs against `main`" presented as standing procedure (card-level PRs no longer exist),
- the 2026-08-14-era rationale that Gitflow's dual-management burden outweighed its benefits stated as CURRENT fact (the transition reverses that judgment — and each file's own header says so).

This is a documentation-alignment defect with routing consequences: `.claude/rules/local/repo-local-pr-policy.md` carries a `paths:` frontmatter keying it to load whenever `.moi/specs/**`, `run.md`, `sync.md`, or `moai.md` workflows are touched — an inaccurate body there actively misroutes lane workflows.

Canonical invariants the repaired bodies must state (from the two sources above):

- Card worktrees branch FROM `develop`, never `main`.
- Completed cards merge into local `develop` via `git merge --no-ff` inside the single integration worktree `.claude/worktrees/develop`; there are NO card-level PRs; remote CI on `origin/develop` is the verdict surface; lanes push `origin/develop`.
- rc builds are cut from `develop` on operator request; `release/vX.Y.Z` branches from `develop` and is the ONLY path into `main`; `main` direct push remains forbidden (`enforce_admins: true`, `required_approving_review_count: 0` — fully binding).

## §B Requirements (GEARS)

### REQ-GDA-001 — Operative-model restatement (Ubiquitous)

Every body passage of the three target documents (`.moai/docs/git-workflow-doctrine.md`, `.moai/docs/git-local-workflow-doctrine.md`, `.claude/rules/local/repo-local-pr-policy.md`) that states where card work branches or how completed card work integrates shall present the git-flow model: card worktrees branch from `develop`, integrate into local `develop` by merge without a card-level pull request, and their integration verdict is CI on `origin/develop`.

### REQ-GDA-002 — Premise correction or dated annotation (Event-driven)

**When** a reader consults a pre-transition premise that stands un-annotated as current rule in `.moai/docs/git-workflow-doctrine.md` — the rationale paragraph at line 15 (which asserts, as standing fact, the judgment the transition reversed), the §18.0.1 forbidden list, and the §18.3.1 tier-based PR routing rows — the document shall either restate the corrected current rule at that premise's own location, or mark the superseded content with a dated `[RETIRED 2026-08-27]` annotation following the file's established `[RETIRED 2026-07-20]` convention.

### REQ-GDA-003 — Forbidden-list truthfulness (Unwanted)

The `.moai/docs/git-workflow-doctrine.md` forbidden lists (§18.0.1 "금지 사항" and §18.10 [HARD]) shall not carry an un-annotated bullet declaring creation of a `develop` branch forbidden; each such bullet shall be corrected IN PLACE to name a prohibition that is true under the current model (for example: branching a card worktree from `main`; opening card-level PRs against `main`).

### REQ-GDA-004 — §23.7 / §23.9 re-scoping (State-driven)

**While** `.moai/docs/git-local-workflow-doctrine.md` §23.7 [HARD] bullets and the §23.9 Tier-based PR Routing section describe change-routing — INCLUDING the always-PR route-prescription tail of the `git push origin main` 금지 bullet (line 151 at d29b8942e; research.md D9), whose feat/chore/docs enumeration presents every change type, card work included, as PR-opening — they shall scope PR-based ceremony to main-touching (release-path) changes — `release/vX.Y.Z` -> `main` with merge-commit strategy — and shall preserve the retired card-level per-tier PR content behind `[RETIRED 2026-08-27]` annotations, per the file's existing historical-preservation convention. The bullet's TRUE half (main direct push rejected server-side) survives un-annotated.

### REQ-GDA-005 — Origin-premise rule rewrite (Ubiquitous)

The `.claude/rules/local/repo-local-pr-policy.md` body shall (a) retain, intact in substance, the [HARD] prohibition of direct push to `main` under `enforce_admins: true`; (b) state the card model — cards branch from `develop`, merge into `develop` via `git merge --no-ff` inside the integration worktree, no card-level PR; and (c) remain coherent with its distributed-template-neutrality paragraph (this file is local-only, never mirrored).

### REQ-GDA-006 — Hard scope fence (Unwanted)

The alignment edits shall not modify: `CLAUDE.local.md` (owned by card t308), `.claude/rules/local/gitflow-lane-protocol.md`, `.moai/docs/template-internal-isolation-doctrine.md`, any path under `internal/template/templates/**`, any CI workflow under `.github/workflows/**`, or any Go source file.

REQ count: 6 (Tier S ceiling: 8).

## §C Verifiable Scope Summary

Full AC enumeration (Given-When-Then, mechanically grep-checkable, with RED-now baselines measured at d29b8942e) lives in `acceptance.md` — the canonical AC layer for this SPEC. Eight criteria, AC-GDA-001 .. AC-GDA-008, each pairing a concrete verification command with its pre-edit observation.

## Out of Scope

### Out of Scope — Files owned elsewhere or mechanically mirrored

- `CLAUDE.local.md` including its §4.1 and §18 cross-reference stubs — card t308 owns it.
- `.claude/rules/local/gitflow-lane-protocol.md` — canonical source, read-only for this SPEC.
- `internal/template/templates/**` — the three target files are repo-local-only and NOT mirrored; no template parity obligation arises.
- `.github/workflows/**` — the push-trigger additions described in the dispatch already exist upstream of this SPEC.

### Out of Scope — Additional un-flipped premises found during planning

Research surfaced further passages still written for the pre-transition model (see research.md §F, items F-1..F-5): the §18.1 branch diagram, the §18.3 merge-strategy table's feature->main rows, §18.5 hotfix branching, the §18.8 release-cut example (`git checkout -b release/vX.Y.Z main`, measured at line 299), and §23.8's `origin/main`-based race-mitigation references. Each is covered reader-side by its file's header notice; repairing them is deferred rather than silently bundled, to keep this card's diff reviewable.

### Out of Scope — Behavior, code, and configuration

- No Go code, config file, hook script, or gate behavior changes — documentation only.
- No renaming or renumbering of existing §18.x / §23.x subsection headings (existing cross-references must keep resolving).
- No decision here on whether an operator-explicit `--pr` card escape hatch survives the transition; the canonical files say card work makes no PRs, and the body text follows them (question recorded in research.md §F-6, non-blocking).

## Cross-references

- Canonical model: `CLAUDE.local.md` §4.1 + `.claude/rules/local/gitflow-lane-protocol.md`
- Targets: `.moai/docs/git-workflow-doctrine.md` (§18), `.moai/docs/git-local-workflow-doctrine.md` (§23), `.claude/rules/local/repo-local-pr-policy.md`
- Annotation convention precedent: `[RETIRED 2026-07-20]` blockquotes and `~~strike-through~~` marks already present in the two doctrine files (left by the 2026-07-20 PR-mandatory change)
