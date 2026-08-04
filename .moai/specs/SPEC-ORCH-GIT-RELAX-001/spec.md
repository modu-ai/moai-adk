---
id: SPEC-ORCH-GIT-RELAX-001
title: "Orchestrator-direct Tier S/M git ops + state-sensitive worktree recovery (manager-git relaxation)"
version: "0.1.1"
status: in-progress
created: 2026-08-04
updated: 2026-08-05
author: manager-spec (via orchestrator delegation)
priority: P1
phase: "v14.4.0 target"
module: ".claude/rules/moai, .claude/agents/moai, .moai/docs, .moai/config/sections, internal/hook"
lifecycle: spec-anchored
tags: "orchestrator, manager-git, git-op, relaxation, doctrine, branch-guard, worktree, context-sensitivity"
related_specs: [SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-V3R6-AGENT-TEAM-REBUILD-001]
---

# SPEC-ORCH-GIT-RELAX-001 — Orchestrator-direct Tier S/M git ops (manager-git relaxation)

## HISTORY

- **0.1.1** (2026-08-04) — iter-2 revision pass per independent plan-auditor verdict (FAIL 0.79 → addressed 5 defects D1–D5, none structural). D1: added `.claude/rules/moai/workflow/spec-workflow.md` Route B as change-surface location #13 (template-mirrored → mirror obligation) + CLAUDE.md:131 enumerated as #14 (verified no-change, already carve-out-scoped). D2: CLAUDE.md:131 under-count resolved (enumerated in Table 1). D3: extended traceability matrix to cover all 16 REQs (added AC-OGR-009 catalog-retention + frontmatter-preservation; AC-OGR-010 foreign-session auto-isolation; extended AC-OGR-002 to credit REQ-001/002/007). D4: resolved [NEEDS CLARIFICATION] marker #2 IN-SPEC — mandated `MOAI_BRANCH_GUARD_EXEMPT=1` per-invocation inline per design.md §3.3 + plan.md §G. D5: tightened AC-OGR-004 fuzzy qualifier. AC count 8 → 10; open clarifications 2 → 1; change-surface 12 → 14 (13 edit + 1 no-change).
- **0.1.0** (2026-08-04) — Initial plan-phase draft. Tier L redesign SPEC; first of a phased redesign ("순차 분할"). This SPEC covers ONLY the manager-git relaxation. Skills/hooks cleanup is a separate follow-up SPEC (out of scope). Triggering incident: PR #1338 handling (manager-git cross-swapped a concurrent session's worktree to main; orchestrator's direct recovery was correct).

---

## §A. User Story

**As a** MoAI orchestrator operating on a shared multi-session primary checkout,
**I want** Tier S/M push+PR and state-sensitive git ops (primary-branch changes, multi-worktree discrimination) handled directly by the orchestrator — the actor with maximum live context (current HEAD, active worktrees, transcript state, concurrent-session registry) — instead of by an isolated `manager-git` subagent holding only a stale session-start git snapshot,
**so that** the structural paradox documented in `.moai/reports/agent-skill-hook-redesign-evidence-20260804.md` §0 is resolved: the MOST context-sensitive operation is no longer forced onto the agent with the LEAST context.

**Relaxation scope (not abolition).** `manager-git` is RETAINED for genuinely complex flows where isolation overhead is justified by recurrence + decomposability + verifiability: Tier L release PRs, multi-step merges, and the Late-Branch 4-Phase closure. `manager-git` has demonstrably performed standard push+PR well (PRs #1319, #1321, #1326, #1330) — this relaxation targets ONLY the context-sensitive Tier S/M + state-op class where its isolation is a liability, not an asset.

---

## §B. Problem Analysis — the structural paradox

The evidence base report (`.moai/reports/agent-skill-hook-redesign-evidence-20260804.md`) establishes the paradox in §0 and §1.3:

1. **manager-git is the ONLY context-SENSITIVE-vulnerable retained agent.** Its operating accuracy depends on live git state (HEAD position, worktree count, branch identity) that its isolated subagent context cannot observe beyond a session-start snapshot. All other 9 retained agents (manager-spec/develop/docs/design, auditors, super-advisor, builder-harness, e2e-tester) operate on self-contained prompts where isolated context is an asset (no stale state risk).
2. **PR #1338 incident = the precise failure signature.** manager-git, holding a stale snapshot, could not restore primary→main and cross-swapped a concurrent session's worktree branch to main. The orchestrator's DIRECT recovery (full context: live HEAD, active worktree registry, transcript) was correct. Forcing the most context-sensitive op onto the agent with least context caused the failure.
3. **Quantitative grounding** (evidence §5.1, arXiv:2512.08296): sequential state-coupled ops degrade −39% to −70% under delegation; multi-agent invocation costs ≈15× tokens vs single chat. Below 45% baseline single-actor accuracy, additional agents produce negative returns.

---

## §C. Requirements (GEARS notation)

### Context-sensitivity inversion principle

**REQ-OGR-001** (Ubiquitous) The MoAI orchestrator SHALL route state-sensitive git operations (primary-branch changes, multi-worktree discrimination, branch recovery after a concurrent-session race) to the full-context actor (the orchestrator directly), and SHALL route context-insensitive recurring git workflows (Tier L release, multi-step merge, Late-Branch 4-Phase closure) to the isolated `manager-git` subagent.

**REQ-OGR-002** (Ubiquitous) The `<subject>` of a state-sensitive git operation (the actor who must hold live HEAD + active worktree count + concurrent-session registry to decide correctly) SHALL be the orchestrator, not an isolated subagent, because reachability of state is the precondition for correct state mutation.

### Orchestrator-direct Tier S/M push+PR

**REQ-OGR-003** (State-driven) **While** the change is Tier S (< 300 LOC, < 5 files) or Tier M (300–1000 LOC, 5–15 files) **AND** no explicit `--pr` flag requests heavy ceremony, the orchestrator SHALL create the feature branch, push it, and open the PR directly via Bash (`git switch -c` / `git push -u` / `gh pr create`) without spawning `manager-git`.

**REQ-OGR-004** (Capability gate) **Where** the orchestrator performs a direct Tier S/M push+PR on the primary checkout, the orchestrator SHALL set the `MOAI_BRANCH_GUARD_EXEMPT=1` sentinel environment variable for the duration of the git-op sequence so the `branch_guard` PreToolUse hook admits the branch-state mutation, and SHALL unset it when the sequence completes.

**REQ-OGR-005** (Event-driven) **When** the orchestrator is about to perform a direct Tier S/M push+PR, the orchestrator SHALL first run the Pre-Spawn/Pre-Edit Sync Check (`.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check + § Pre-Edit Sync Check — `git fetch origin main`, `git rev-list --count --left-right origin/main...HEAD`, and `moai session list --json` active-sessions query) and SHALL halt on any divergence (`N 0` origin-ahead, `N M` diverged) or any foreign active session on the same SPEC scope.

**REQ-OGR-006** (Event-driven) **When** the Pre-Spawn/Pre-Edit Sync Check detects ≥1 foreign active session on the same checkout during the direct push+PR sequence, the orchestrator SHALL auto-isolate into a worktree (per `worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation) OR surface an `AskUserQuestion` round (isolate / wait / abort) rather than proceeding on the shared tree.

### manager-git retention (relaxation, not abolition)

**REQ-OGR-007** (State-driven) **While** the change is Tier L (> 1000 LOC OR > 15 files OR constitutional) OR the user has explicitly requested `--pr` heavy ceremony, the orchestrator SHALL delegate push+PR to `manager-git` as today (no behavior change for the complex-flow class).

**REQ-OGR-008** (Ubiquitous) The `manager-git` subagent SHALL remain the canonical owner of: (a) Tier L release PR creation, (b) multi-step merge execution (e.g. `--auto-merge` team-mode flows), and (c) Late-Branch 4-Phase closure (`manager-git.md` § Late-Branch Invocation Pattern, Phases A–D).

**REQ-OGR-009** (Unwanted) The `manager-git` subagent SHALL NOT be spawned for Tier S/M push+PR in the absence of the `--pr` flag; spawning it in that context is the anti-pattern this SPEC retires.

### Doctrine consistency across all 12 change surfaces

**REQ-OGR-010** (Ubiquitous) All 13 change-surface locations requiring edits (§D Table 1 rows #1–#13; row #14 is verified no-change) SHALL be updated to reflect the relaxation consistently; a cross-reference audit (grep) SHALL confirm zero residual statements of the form "all tiers route push+PR through manager-git" outside the Tier L / `--pr` carve-out. The grep scope SHALL include `.claude/rules/moai/workflow/spec-workflow.md` (Route B description — location #13) and its template mirror.

### Branch-guard exemption path verification

**REQ-OGR-011** (Capability gate) **Where** the orchestrator sets `MOAI_BRANCH_GUARD_EXEMPT=1` for a direct Tier S/M git-op sequence, the `isExemptAgent` function (`internal/hook/branch_guard.go:144-155`) SHALL admit the orchestrator-direct op WITHOUT requiring a Go code change, because the env-var branch of the exemption predicate already returns true before the `AgentType == "manager-git"` identity check is reached.

**REQ-OGR-012** (Event-driven) **When** verification shows the env-var exemption path does NOT admit orchestrator-direct Tier S/M (e.g. the env var is not read in the relevant code path, or the hook is not on the orchestrator's PreToolUse surface), the run-phase SHALL implement the minimal Go change to make it admit — but SHALL NOT widen the exemption beyond the env-var predicate (no new identity branch for "orchestrator").

### Regression coverage of the incident class

**REQ-OGR-013** (Event-driven) **When** the run-phase verification suite runs, a regression test SHALL reproduce the PR-#1338 incident class (multi-worktree state op: orchestrator-with-context restores primary→main correctly vs an isolated snapshot-holder that would cross-swap) and SHALL demonstrate that the orchestrator-direct path resolves it.

### Template-First + namespace + neutrality

**REQ-OGR-014** (Capability gate) **Where** a doctrine change touches a template-mirrored file (CLAUDE.md, `.claude/agents/moai/manager-git.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/rules/moai/workflow/main-checkout-branch-guard.md`, `.moai/config/sections/delegation.yaml`), the change SHALL be mirrored to `internal/template/templates/` counterpart, `make build` SHALL regenerate embedded assets, and the template-neutrality guard (`internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml`) SHALL pass.

**REQ-OGR-015** (Unwanted) The doctrine update SHALL NOT mirror to `internal/template/templates/` for the LOCAL-ONLY files (`CLAUDE.local.md`, `.moai/docs/git-local-workflow-doctrine.md`, `.moai/docs/repo-local-pr-policy.md`) — those carry no template counterpart by design and mirroring them would violate §24/§25.

**REQ-OGR-016** (Unwanted) The relaxation SHALL NOT remove `manager-git` from the retained agent catalog (11 agents), SHALL NOT change its `model: sonnet` / `effort: low` / `permissionMode: bypassPermissions` frontmatter, and SHALL NOT alter its Late-Branch 4-Phase Pattern body (Phases A–D stay verbatim).

---

## §D. Change-surface map (14 enumerated locations — 13 edits + 1 verified no-change, after evidence §4 + iter-2 audit)

Table 1 — the 14 enumerated locations (13 requiring edits + 1 verified no-change). Iter-2 audit expanded the census beyond evidence §4's original 12: location #13 (`spec-workflow.md` Route B) was missed by evidence §4 (template-mirrored → mirror obligation), and location #14 (`CLAUDE.md:131`) was an unenumerated 3rd CLAUDE.md site already carve-out-scoped (no edit needed). Verbatim file:line in the evidence report + iter-2 audit.

| # | Location | Current rule | After relaxation | Owner milestone |
|---|----------|--------------|------------------|-----------------|
| 1 | `CLAUDE.md:75` (selection tree row 6) | "Tier L OR `--pr` → manager-git" | Tier S/M default orchestrator-direct; Tier L / `--pr` only → manager-git | M1 |
| 2 | `CLAUDE.md:90` (catalog table row) | "PR creation per Tier + Late-Branch" | "Tier L / `--pr` PR + Late-Branch 4-Phase; Tier S/M orchestrator-direct" | M1 |
| 3 | `.moai/docs/git-local-workflow-doctrine.md:149` (§23.9, HARD) | "모든 tier PR 경유 manager-git" (HARD) | Tier S/M orchestrator-direct; Tier L / `--pr` → manager-git | M1 (LOCAL-ONLY, no mirror) |
| 4 | `.moai/docs/git-local-workflow-doctrine.md:153-156` (Tier table) | S/M/L rows all "manager-git" | S/M rows → orchestrator-direct; L row stays manager-git | M1 (LOCAL-ONLY) |
| 5 | `.moai/docs/git-local-workflow-doctrine.md:160` (REQ-ATR-020) | "모든 tier PR 생성은 manager-git 책임" | Carve-out: Tier S/M exempted; Tier L / `--pr` retained | M1 (LOCAL-ONLY) |
| 6 | `.moai/docs/git-local-workflow-doctrine.md:162` (Late-Branch 4-Phase) | Tier L 4-Phase | UNCHANGED (Tier L only — stays) | (no change) |
| 7 | `.claude/agents/moai/manager-git.md:5` (invocation gate, description) | "push + PR always delegated to manager-git" | "Tier L / `--pr` delegated; Tier S/M orchestrator-direct" | M2 |
| 8 | `.claude/agents/moai/manager-git.md:88-129` (Late-Branch body) | Phases A–D | UNCHANGED (Tier L only — stays) | (no change) |
| 9 | `.moai/config/sections/delegation.yaml:74` | note "manager-git is Tier-L / --pr conditional" | Refresh note: Tier S/M orchestrator-direct; Tier L / `--pr` → manager-git | M2 |
| 10 | `.claude/rules/moai/workflow/main-checkout-branch-guard.md:104,124` | exemption = `AgentType=="manager-git"` OR env | Document that orchestrator-direct Tier S/M uses the env path (`MOAI_BRANCH_GUARD_EXEMPT=1`); Go unchanged expected | M2 |
| 11 | `.claude/rules/moai/core/agent-common-protocol.md` §Pre-Spawn / §Pre-Edit | manager-git identity exemption context | Add orchestrator-direct Tier S/M context (env-path exemption + pre-session sync check) | M2 |
| 12 | `internal/hook/branch_guard.go:140-154` (`isExemptAgent`) | env OR identity exemption | VERIFY env path admits orchestrator-direct Tier S/M with no Go change; minimal change only if verification fails | M3 |
| 13 | `.claude/rules/moai/workflow/spec-workflow.md:26` (Route B description — iter-2 addition; **template-mirrored**) | "Route B — PR route (Tier L OR `--pr`): `manager-git` creates a feature branch and opens a PR per phase" | Qualify Route B actor: Tier L / `--pr` → `manager-git`; Tier S/M Route B (per `repo-local-pr-policy.md` ALL-tier Route B override) → orchestrator-direct per SPEC-ORCH-GIT-RELAX-001. **Mirror obligation**: `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` edited in the SAME milestone. | M1 (edit + mirror) |
| 14 | `CLAUDE.md:131` (§5 Agent Chain — iter-2 enumeration) | `[optional Tier L OR --pr] PR (manager-git)` — 3rd CLAUDE.md site (catalog at #1/#2 listed only lines 75, 90) | **NO CHANGE** — already carve-out-scoped (`[optional Tier L OR --pr]` prefix limits to the carve-out). Enumerated for census completeness (resolves the under-count); no edit required. | (no change) |

---

## §E. Constraints

- **16-language template neutrality** (CLAUDE.local.md §15): doctrine changes in `internal/template/templates/` must pass the template-neutrality CI guard.
- **Template-First** (CLAUDE.local.md §2 [HARD]): every `.claude/` or template-`.moai/` change → mirror to `internal/template/templates/` + `make build`. LOCAL-ONLY files (CLAUDE.local.md, git-local-workflow-doctrine.md, repo-local-pr-policy.md) MUST NOT be mirrored.
- **Namespace separation** (§24): no `moai-*` / `harness-*` confusion introduced.
- **Incremental & traceable**: no retroactive rewriting of unrelated doctrine; existing Late-Branch body (Phases A–D) stays verbatim.
- **"완화 not 폐지"**: `manager-git` is retained; the SPEC sculpts its responsibility surface, it does not abolish the agent.
- **No `--no-verify` shortcuts**: the `MOAI_BRANCH_GUARD_EXEMPT=1` sentinel is the sanctioned exemption path; bypassing pre-commit/pre-push hooks (`git commit --no-verify`, `git push --no-verify`) remains prohibited per `coding-standards.md` § Bash Risk-Amplifier Doctrine.
- **Implementation Kickoff Approval unchanged**: this relaxation does not weaken the plan→run human gate; it changes which actor performs the run-side git-op, not whether the user approved run-phase entry.

---

## §F. Out of Scope

### Out of Scope — Skills catalog cleanup

- TRUST 5 framework 4-fold duplication consolidation (`moai-foundation-core` vs `moai-foundation-quality` vs `moai-workflow-testing` vs `moai-foundation-core/modules`).
- `moai-workflow-testing` scope reduction (33-module kitchen-sink → testing-first principles).
- Orphan skill triage (`hns-moaiadk-dev-reference`, `moai-workflow-docs-claim-check`, `moai-workflow-ci-loop`).
- `moai-meta-harness` stale reference correction (2 sites: `moai-harness-learner/SKILL.md:122`, `moai/workflows/harness.md:265`).
- `moai workflow lint` coverage extension (orphans / scope-creep / stale-ref detection).
- These belong to a **separate follow-up SPEC** (skills-lens redesign).

### Out of Scope — Hooks cleanup

- 5 possible-dead hooks verification (`handle-elicitation`, `handle-elicitation-result`, `handle-notification`, `handle-session-start-compact`, `handle-task-created`).
- `team-ac-verify.sh` stub — implementation or removal decision.
- `security-{scan,turn,commit}` 3-hook duplication deep verification.
- These belong to a **separate follow-up SPEC** (hooks-lens redesign).

### Out of Scope — manager-git frontmatter fields

- `manager-git` `model: sonnet` pin (the only non-`inherit` pin in the catalog) — the evidence report flags this as "independent of the relaxation" and it is tracked separately under model-policy.
- `manager-git` `permissionMode: bypassPermissions` — retained; the relaxation routes work AWAY from manager-git for Tier S/M, so the permission path sees less traffic, not different traffic.
- `manager-git` `effort: low` — retained (matrix anchor).

### Out of Scope — output-style banner drift

- 3-persona banner drift quantification (evidence §6 D, flagged thin) — belongs to a docs/sync-lens follow-up.

### Out of Scope — Late-Branch 4-Phase Pattern rewrite

- The Late-Branch Phases A–D body in `manager-git.md` § Late-Branch Invocation Pattern stays verbatim. It is the canonical Tier L closure and is explicitly retained (REQ-OGR-008).

### Out of Scope — branch_guard opt-in gate flip

- `Workflow.BranchGuard.Enabled` defaults to `false` (distributed default). This SPEC does not flip the opt-in gate; it documents that WHEN the guard is enabled, the env-var exemption path admits orchestrator-direct Tier S/M. Flipping the distributed default is a separate decision.

---

## §G. Acceptance Criteria (summary — full Given-When-Then in acceptance.md)

- **AC-OGR-001** — orchestrator creates+pushes+PRs a Tier S/M branch directly with `MOAI_BRANCH_GUARD_EXEMPT=1` + Pre-Spawn/Pre-Edit Sync Check, WITHOUT spawning manager-git.
- **AC-OGR-002** — manager-git is NOT invoked for Tier S/M (only Tier L / explicit `--pr`).
- **AC-OGR-003** — all 13 edit locations updated consistently (cross-reference grep audit covers §D Table 1 rows #1–#13 + their template mirrors; returns 0 residual "all tiers → manager-git" statements outside the Tier L / `--pr` carve-out).
- **AC-OGR-004** — branch-guard exemption path verified — orchestrator-direct Tier S/M admitted via `MOAI_BRANCH_GUARD_EXEMPT=1` (Go unchanged or minimal, env-path cited).
- **AC-OGR-005** — full gate passes (agent lint, workflow lint, template-neutrality, mirror-parity, branch-guard tests, `make build`).
- **AC-OGR-006** — regression test proves orchestrator-direct resolves the PR-#1338 incident class (multi-worktree state op).
- **AC-OGR-007** — Late-Branch 4-Phase body (Phases A–D) remains verbatim (byte-identical) in `manager-git.md`.
- **AC-OGR-008** — LOCAL-ONLY files (CLAUDE.local.md, git-local-workflow-doctrine.md, repo-local-pr-policy.md) carry NO `internal/template/templates/` mirror.
- **AC-OGR-009** — `manager-git` remains in the CLAUDE.md §4 retained-agent catalog AND its frontmatter `model`/`effort`/`permissionMode` fields are unchanged post-relaxation (REQ-OGR-016 catalog-retention + frontmatter-preservation limbs).
- **AC-OGR-010** — when the Pre-Spawn/Pre-Edit Sync Check detects ≥1 foreign active session on the same checkout + SPEC scope during a direct Tier S/M push+PR, the orchestrator halts and either auto-isolates to a worktree OR surfaces an `AskUserQuestion` round (REQ-OGR-006 foreign-session auto-isolation).

---

## §H. Cross-References

- **Evidence base**: `.moai/reports/agent-skill-hook-redesign-evidence-20260804.md` (§0 paradox, §1.3 context-sensitivity classification, §4 12-location change surface, §5.1 quantitative grounding, §5.2 7 delegation-vs-direct principles).
- **Triggering incident**: PR #1338 handling (manager-git primary→main restore failure + concurrent-session worktree cross-swap).
- **Precedent SPECs**: `SPEC-WORKTREE-BRANCH-GUARD-001` (the branch-state guard this relaxation uses), `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001` (the opt-in gate), `SPEC-V3R6-AGENT-TEAM-REBUILD-001` (REQ-ATR-020 — the doctrine this SPEC carves out).
- **Doctrine SSOTs updated**: CLAUDE.md §4 + selection tree; `.moai/docs/git-local-workflow-doctrine.md` §23.7/§23.9 (LOCAL-ONLY); `.claude/agents/moai/manager-git.md`; `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn/Pre-Edit Sync Check; `.claude/rules/moai/workflow/main-checkout-branch-guard.md`; `.moai/config/sections/delegation.yaml`.
- **Web research**: arXiv:2512.08296 (multi-agent scaling limits), arXiv:2604.14228 (Dive-into-Claude-Code), Anthropic multi-agent guidance (≈15× token overhead). Verbatim URLs in `research.md`.
