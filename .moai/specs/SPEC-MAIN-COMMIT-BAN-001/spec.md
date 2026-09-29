---
id: SPEC-MAIN-COMMIT-BAN-001
title: "Local main Commit Ban — Doctrine Codification, Protected-Branch Commit Deny, Residue Disposition Procedure, Lead Push Threshold"
version: 0.1.0
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: high
phase: "v3.2.0"
module: hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, branch-guard, git-flow, doctrine, config, kanban"
related_specs: [SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001]
---

# SPEC-MAIN-COMMIT-BAN-001

## §A Problem / Motivation

Operator directive (card t1337): **local `main` is DEAD for commits.** Commits land on
develop-based card worktrees only; the lead pushes `develop` to `origin` in batches. Four gaps
make this policy today doc-only-at-best and unenforced at worst:

1. **The standing contract permits the exact act being banned.** `AGENTS.md` §2 (both the local
   copy and the template mirror `internal/template/templates/AGENTS.md.tmpl`) currently reads:
   *"Read-only inspection, `git fetch`, commits to the already-checked-out branch, and pushing it
   are permitted."* A session that finds the primary checkout sitting on `main` is contractually
   allowed to commit there.
2. **The mechanical guard has a commit-shaped hole.** `internal/hook/branch_guard.go`
   `branchStatePatterns` covers `switch` / `checkout <branch>` / `reset --hard` / `stash`
   / `rebase` / `merge` / mutating `git branch` — but NOT `git commit` (nor `git revert` /
   `git cherry-pick`). With the guard enabled, `git commit` on `main` in the primary checkout
   is allowed today.
3. **Residue exists on the primary checkout.** The primary checkout sits on `main` at
   `c8f245c2c` (measured: `git rev-list --count origin/main..main` = 1, exactly
   `c8f245c2c chore(policy): retire primary CLAUDE.local.md`) carrying develop-state files as
   uncommitted modifications to files still tracked on main. `CLAUDE.local.md` itself is
   UNTRACKED on main (deleted from tracking by that commit; `git cat-file -e
   main:CLAUDE.local.md` → absent) and gitignored — the operator-maintained working copy
   (whose §0.3/§0.4 record the retired-model doctrine) is primary-only and worktree-invisible.
   The tracked home of the local git-flow doctrine on develop is **`AGENTS.local.md` §4.1**
   (measured: tracked, `### §4.1 GitFlow 통합 체인 (develop)` present).
   No documented, guard-aware disposition procedure exists for retiring that residue.
4. **The lead's batch-push has no codified trigger.** `.moai/docs/gitflow-integration-chain.md`
   step 2 reads "push 시점 판단 — 리드가 배치를 닫을 시점을 정해 한 번 push한다" — the timing is
   pure lead judgement. The operator has fixed the trigger: **20 unpushed commits**. No
   machine-readable key carries that value anywhere.

## §B Scope

| # | Scope | Class |
|---|-------|-------|
| 1 | Doctrine update: AGENTS.md §2 (local + template mirror), gitflow doctrine docs | docs |
| 2 | Mechanical guard: protected-branch commit deny in the PreToolUse BranchGuard | code |
| 3 | Residue disposition procedure: documented operator-side retirement of local main residue | docs |
| 4 | Lead push trigger threshold: `20` unpushed commits, machine-readable | code+docs |

## §C Requirements

### REQ-1 — Doctrine: main is commit-dead (Scope 1)

- REQ-1.1: The local `AGENTS.md` §2 permitted-operations clause MUST stop listing unconditional
  "commits to the already-checked-out branch" as permitted; it MUST carry a protected-branch
  exception naming `main` as commit-forbidden in this repository, with a pointer to
  `.moai/docs/gitflow-integration-chain.md`.
- REQ-1.2: The template mirror `internal/template/templates/AGENTS.md.tmpl` MUST carry the
  parallel clause in GENERIC form (workflow-declared protected branch; git-strategy mode decides
  which) — NO card ids, NO internal dates, NO card-provenance narrative. Enforced leak classes
  (measured, `internal_content_leak_test.go` + doctrine §25.3): C1 SPEC-ID, C2 REQ/AC-token,
  C3 audit-citation, C4a date, C4b short-sha, C5 memory/archive-path; the card-id and
  CLAUDE.local-reference prohibitions are prose-level (no leak class) — sanitized-pair parity
  is the de facto guard for the rule pair.
- REQ-1.3: The develop-tracked local doctrine set MUST codify the ban with repo specifics —
  all four lane-editable in the worktree (existence measured): `AGENTS.local.md` §4.1,
  `.claude/rules/local/gitflow-lane-protocol.md` §1,
  `.moai/docs/gitflow-integration-chain.md`, and the git-flow revision note in
  `.moai/docs/git-workflow-doctrine.md` (line ~64 block).
- REQ-1.4: The primary-only `CLAUDE.local.md` working copy (untracked, gitignored —
  `.gitignore:276`; absent from worktrees by measurement) receives the same clause as an
  OPERATOR-side step at disposition time (REQ-4.6), never a lane edit — the lane writes only
  inside its worktree.

### REQ-2 — Mechanical guard: protected-branch commit deny (Scope 2)

- REQ-2.1: The PreToolUse BranchGuard MUST deny a commit-creating git command (`git commit`,
  `git revert`, `git cherry-pick`) when ALL of: (a) the command's actual cwd is the primary
  checkout (existing Seam A discriminant), (b) the resolved HEAD branch is in the configured
  protected list, (c) the invoking agent is not exempt (existing two exemption axes).
- REQ-2.2: Every deny MUST carry the `BRANCH_GUARD_VIOLATION:` sentinel prefix with a
  branch-specific suffix, and MUST NOT name a manager-git delegation as remediation (t43 rule).
- REQ-2.3: The deny MUST share the guard's fail-open norm: HEAD-resolution failure or git-context
  uncertainty → allow + stderr advisory + audit-log append. Detached HEAD → allow (no named
  branch to protect).
- REQ-2.4: The check MUST be a separate matcher/evaluator from `branchStatePatterns` (which is
  branch-agnostic) — the commit deny is branch-CONDITIONAL. It runs inside the same
  `checkBranchState` seam (same opt-in gate, same exemption, same normalization pipeline:
  quoted-argument, heredoc, comment, PowerShell handling).
- REQ-2.5: When the protected list is empty, the commit-family check MUST short-circuit to allow
  BEFORE any HEAD-resolving subprocess runs (zero added cost for unconfigured users).

### REQ-3 — Config surface (Scopes 2+4)

- REQ-3.1: `workflow.branch_guard.deny_commits_on` (list of branch names) MUST exist in
  `internal/config` (`BranchGuardConfig`), default `[]` in `defaults.go` (template-neutral,
  sibling of `Enabled: false`), mirrored in the template `workflow.yaml` (or its source) as
  `deny_commits_on: []`, and set to `[main]` in the LOCAL `.moai/config/sections/workflow.yaml`.
- REQ-3.2: The struct-YAML symmetry audit MUST stay green (`TestStructYAMLSymmetry_*`).
- REQ-3.3: `git_strategy.manual.lead_push_threshold` (int) MUST exist in `internal/config`
  git-strategy structs, default `0` (= disabled — manual-mode users ship `push_to_remote: false`),
  mirrored in the template `git-strategy.yaml`, and set to `20` in the LOCAL
  `.moai/config/sections/git-strategy.yaml` (operator-given initial value — recorded, not derived).
- REQ-3.4: No NEW opt-in boolean is introduced for the commit deny — it rides the existing
  `Workflow.BranchGuard.Enabled` gate (dogfood: enabled; template: false).

### REQ-4 — Residue disposition procedure (Scope 3)

- REQ-4.1: A documented OPERATOR-side procedure MUST retire local main residue, with FIXED
  ordering as a precondition (not a suggestion): (1) verify the existing preservation copy at
  primary `.moai/state/retired/CLAUDE.local.md` (sha256 `1db8d302…`) — reference-only; this SPEC
  creates NO new preservation artifact; (2) move the primary checkout to `develop` FIRST
  (`develop` checked out nowhere else — the integration worktree, if present, holds it);
  (3) only then, with `main` checked out NOWHERE, run `git fetch origin` + `git branch -f main
  origin/main` (zero working-copy impact).
- REQ-4.2: The reset target MUST be `origin/main` (NOT develop): main's role is the honest,
  synchronized reference of the release-PR landing surface; resetting to develop would
  fabricate a main-ahead-of-origin state — the `c8f245c2c` defect shape. Abort the procedure if
  the `origin/main` ref cannot resolve.
- REQ-4.3: The procedure MUST restate the standing prohibition: `git restore CLAUDE.local.md`
  is FORBIDDEN (CLAUDE.local.md §0.4) — restoring would regress the working copy to the retired
  third model.
- REQ-4.4: The procedure MUST carry a rehearsal-first step (scratch clone; observe git's actual
  refusals — the main→develop switch crosses a tracked-file/modified-working-copy boundary that
  MUST be observed, not assumed) and a documented-exception note: steps 2-3 are branch-state
  commands the guards deny to agent sessions BY DESIGN; the operator's own terminal is the
  execution channel, and a `BRANCH_GUARD_VIOLATION` on this procedure inside a session is
  correct behavior, not a bug.
- REQ-4.5: The motivating incident MUST be recorded in this SPEC's background (`c8f245c2c`,
  card t1317) and the procedure home (`gitflow-integration-chain.md`) references it.
- REQ-4.6: The procedure MUST carry an operator-side doc step syncing the commit-ban clause
  into the primary-only `CLAUDE.local.md` §4.1 working copy (untracked — reachable only from
  the primary), and a post-condition note: after `main = origin/main`, main RE-TRACKS
  `CLAUDE.local.md` while the develop-side primary holds it untracked — any future primary
  switch to main refuses "untracked would be overwritten" (a re-armed boundary the procedure
  documents, not a defect to fix).

### REQ-5 — Lead push trigger threshold (Scope 4)

- REQ-5.1: `.moai/docs/gitflow-integration-chain.md` step 2 MUST change from free judgement to
  threshold-triggered: the lead reads `git rev-list --count origin/develop..develop` and closes
  the batch when the count ≥ the configured threshold.
- REQ-5.2: The threshold triggers the batch-close DECISION only; the push itself stays outside
  the integration window (standing rule: push는 창 밖 — 리드 일괄). An open window is never
  interrupted by the threshold.
- REQ-5.3: `AGENTS.local.md` §4.1 (the develop-tracked local doctrine, lane-editable) MUST cite
  the config key; the doctrine names the key and the counter command, with the config file
  authoritative for the current value. The primary-only `CLAUDE.local.md` copy receives the
  same citation via the REQ-4.6 operator-side sync.

## §D Out of Scope

### Out of Scope — Excluded by Design

- NOT denying `git pull` / `git fetch` on main: main stays a synchronized reference; syncing it
  is legitimate (lane protocol §1). Recorded as accepted residual.
- NOT denying non-commit-creating commands (`push`, `tag`, `notes`, `worktree`).
- NOT denying commits in card worktrees or the develop integration worktree — the ban is
  primary-checkout + protected-branch scoped.
- NOT executing the residue procedure on the real primary checkout within this SPEC's run phase —
  it is documented for the operator and rehearsed in a scratch clone only (REQ-4.4).
- NOT changing the template default of `Workflow.BranchGuard.Enabled` (stays false) and NOT
  shipping `deny_commits_on: [main]` to distributed users (stays `[]`).
- NOT extending the deny to `git merge`/`rebase` onto protected branches specifically — those are
  already denied in the primary checkout on ANY branch by the existing pattern set.

## §E Risks / Edge Cases

| # | Risk | Disposition |
|---|------|-------------|
| E-1 | Lexical commit-detection misses obfuscated forms (`bash -c "git commit …"`) AND adjacency forms: `git -C <primary-path> commit` executed from a worktree cwd classifies as worktree by the cwd-based discriminant and is NOT caught | Both accepted — same under-match/fail-open direction as the documented `git -C <path> branch` residual in the flagclass classifier; named in the plan's test notes |
| E-2 | The main→develop switch crosses the primary's REAL boundary: the develop-state working copy holds uncommitted modifications to files still tracked on main (modified-tracked-set) plus untracked working files that develop tracks (untracked-overwrite refusal). `CLAUDE.local.md` is NOT part of that boundary — untracked+gitignored on both sides (deleted from main's tracking by `c8f245c2c`), it survives any switch silently. A fresh clone at origin/main would observe silent removal, not a refusal — so the rehearsal MUST reproduce the primary's state class (develop-state content as unstaged modifications over main@c8f245c2c), not a clean clone | Rehearsal-first records whatever git actually refuses; the procedure is written from observed behavior (REQ-4.4) and names the modified-tracked-set boundary |
| E-3 | Symmetry audit cannot see the new keys (instrument is one level deep; workflow/git-strategy sections not in `symmetryCases` — measured) | REQ-3.2/REQ-3.3 satisfied by a DEDICATED template-presence test (plan D8), not by the existing symmetry harness |
| E-4 | Template edit leaks card provenance | Enforced leak classes C1 (SPEC-ID) / C3 (audit-citation) / C4a (date) / C4b (short-sha) / C5 (memory-path) + sanitized-pair parity for the rule pair; card-id and CLAUDE.local-reference prohibitions are prose-level (no leak class) and bind by review |
| E-5 | Threshold drift between doctrine prose and config | Config is authoritative (REQ-5.3); doctrine cites key name, states initial value as historical record |
| E-6 | `git branch -f main` refused because main is checked out somewhere | Ordering precondition REQ-4.1 (main checked out NOWHERE) + abort-and-report step |
