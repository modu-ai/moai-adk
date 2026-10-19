---
id: SPEC-AGENTS-CONTRACT-REVISE-001
title: "Agent-instruction contract revision — per-card PR flow, leader boundary, review gate, size diet"
version: "0.1.0"
status: completed
created: 2026-10-19
updated: 2026-10-20
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "AGENTS.md"
lifecycle: spec-anchored
tags: "agent-contract, pr-flow, review-gate, leader-boundary, size-diet, template-first, docs"
tier: M
related_specs: [SPEC-AGENTS-MD-CANON-001, SPEC-LOCAL-MAIN-FLOW-001]
---

# SPEC-AGENTS-CONTRACT-REVISE-001 — Agent-instruction contract revision

## §A Summary

This SPEC revises the agent-instruction contract surfaces so the written policy matches the operator-approved 2026-10-19 policy set: per-card fast draft-PR delivery, a coordination-only leader boundary, a review-gate contract derived from the PR #1841 merge-before-review incident, and an AGENTS.local.md size diet. It consolidates the abandoned cards t1884 (docs/PR-flow consistency) and t1886 (AGENTS.local.md size diet) into one revision, adopts/adapts/rejects each proposal group of the committed review inputs (`inputs/REVIEW.md`, `inputs/proposed.patch`), and encodes a strict policy-vs-implementation distinction: capabilities whose implementation has not landed (t1868 review gate, t1882 factory-complete draft/auto-merge path, t1895 config defaults) are described only as approved policy plus an in-progress task — never as shipped tooling.

Scope targets (operator-named): `AGENTS.md`, `AGENTS.local.md`, `internal/template/templates/AGENTS.md.tmpl`, the generation/deploy/update paths that render or sync these files (documentation of those paths only — code changes belong to the owning cards), and the directly-referenced documents the review inputs call out.

## §B Background and HISTORY

### HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-10-19 | 0.1.0 | Initial plan-phase authoring (factory card t1900, run tn1p4e, lane-44). Consolidates abandoned cards t1884 + t1886. Leader amendment (operator follow-up, board 559a) added the PR #1841 merge-before-review incident as motivating evidence and the review-gate rule block (REQ-ACR-004). |

The current contract documents still carry the superseded local-main batch model. Verified conflict sites in this tree (branch `WT-agents-agents-md`, HEAD `8249bff8c3`):

- `AGENTS.md:10` — "handle simple, bounded operations directly" (leader direct-execution exception).
- `AGENTS.md:34` — capability-table fallback "Do the bounded work inline".
- `AGENTS.md:200-205` — CodeRabbit-only review-completion contract.
- `AGENTS.local.md:70` — template-source list names `internal/template/templates/CLAUDE.md`, which does not exist in this tree (the template root carries `AGENTS.md.tmpl`).
- `AGENTS.local.md:446` — self-check question points at "CLAUDE.md Section 4"; the agent catalogue lives at `.moai/config/sections/delegation.yaml`.
- `AGENTS.local.md:458` — "Permitted direct execution" section (typo fixes, urgency, explicit requests, absence of a delegation target).
- `AGENTS.local.md:182-238` — §4.0 local-main integration flow with seven [HARD] items (branch base from local main, tool-path merge landing, clean-tree rule, resync, closing/disposal gated on batch landing, batch release procedure, batch-closing trigger).
- `AGENTS.local.md:239-292` — §4.1 integration chain (main).
- `.claude/rules/local/gitflow-lane-protocol.md:85` — "Lanes do not push `main` … passes the merge commit SHA … to the leader"; `:43` and `:53`/`:87`/`:93` carry the leader-bulk-push framing. A lane cited §4 as grounds to refuse a preservation push (measured operator report; its revision is in scope).
- `.claude/rules/local/repo-local-pr-policy.md` — card-workflow block: "There are NO card-level PRs … lanes report their local merge SHA and do NOT push — the factory leader collects lane merge SHAs and closes the batch".
- `.claude/rules/moai/workflow/factory-dispatch.md:20` ("the landing sequence is the lane's to run", "the lane closes its own card at completion"), `:183` ("dispose of none until the leader integrated it and the remote merge landed"), `:219`/`:222` (git-flow variant landing), and `.claude/rules/moai/workflow/factory-dispatch-mechanics.md:125` ("The lane's standard landing").

These sites contradict operator policies 1 (per-card draft-PR flow), 3 (wait only on real dependencies), and 4 (leader coordination only) — see §D.

### The #1841 incident (motivating evidence; committed inputs, cited verbatim)

Measured and preserved in `.moai/specs/SPEC-AGENTS-CONTRACT-REVISE-001/inputs/REPORT.md` and `.moai/specs/SPEC-AGENTS-CONTRACT-REVISE-001/inputs/native-codex-rest-comment.json` (read-only evidence copies in this SPEC's inputs; do not edit):

- PR #1841 (modu-ai/moai-adk): `mergedAt=2026-10-19T05:13:31Z`, `mergeCommit=8249bff8c3933d647b4f99e59048595f9808e83a`, `headRefOid=d4ea064178412c0d8c1dc149fc83d4ff81470511`, `reviews=[]` (REPORT.md §"리뷰 전 병합 발견과 후속 정정").
- Native Codex comment (id 6260188619, author `chatgpt-codex-connector[bot]`, type Bot): created 2026-10-19T05:13:40Z; code review completed 2026-10-19T05:14:11Z, security review completed 2026-10-19T05:15:04Z; both rows read "Review trigger: Draft marked ready"; `codex-review-state:v1` body carries `code_review` and `security_review`, both `status:"completed"`, `has_any_findings:false`, `has_unresolved_findings:false`, `commit_sha:"d4ea0641…"` (native-codex-rest-comment.json).
- Timeline: merge 05:13:31Z precedes the code-review start (05:13:37Z), the security-review start (05:13:38Z), and both completions (05:14:11Z / 05:15:04Z). **A merge-before-review-completion violation; the post-hoc pass does not cure it** (REPORT.md: "사후 성공은 병합 시점의 누락을 소급해 통과시키지 않는다").

## §C Measured Facts (this tree)

All figures below were measured in THIS worktree (`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1900`, branch `WT-agents-agents-md`, HEAD `8249bff8c3933d647b4f99e59048595f9808e83a`, working tree clean except the untracked SPEC directory):

| # | Fact | Evidence (file:line) |
|---|------|----------------------|
| MF-1 | `.github/workflows/auto-merge.yml` is an ungated CI auto-merge path: trigger `workflow_run: workflows: ["CI"], types: [completed]` (L7-10); job condition only `event == 'pull_request' && conclusion == 'success'` (L23-26); no `draft` filter (grep for `draft` returns no hits); no configuration gate. | `.github/workflows/auto-merge.yml:7-10,23-26` |
| MF-2 | Merge-flag rule: `if [[ "$PR_BRANCH" == release/* \|\| "$PR_BRANCH" == hotfix/* ]]; then MERGE_FLAG="--merge" … else … "--squash"` — release/* OR hotfix/* → `--merge`, else `--squash`. An earlier SPEC omitted `hotfix/*` and failed its audit; do not repeat that. | `.github/workflows/auto-merge.yml:341-345` |
| MF-3 | The merge call carries `--match-head-commit "$TESTED_SHA" --delete-branch` — automatic branch deletion is active, which mismatches the no-automatic-branch-deletion policy; record as a limitation. | `.github/workflows/auto-merge.yml:355` |
| MF-4 | `moai factory complete` opens the PR without `--draft` and then requests an auto-merge: `factoryGHCall(wt, "pr", "create", "--base", target, …)` and `factoryGHCall(wt, "pr", "merge", <n>, "--auto", mergeFlag, "--match-head-commit", checkedTip)`. This is card t1882's implementation scope — docs must not describe the draft/auto-merge separation as implemented. | `internal/cli/factory_card_pr.go` (pr-create block ~L410, pr-merge block ~L444, read L390-460) |
| MF-5 | AGENTS.local.md measures **50,579 chars** (`wc -m`) — the committed copy at HEAD. The operator memo's 31,971-char figure is the uncommitted primary-checkout working copy, which AGENTS.local.md §0.2 declares non-citable as canonical; the canonical-copy discriminator is §0.1 (the tree lanes branch from). | `wc -m AGENTS.local.md` (this run); `AGENTS.local.md:16-47` |
| MF-6 | The four largest AGENTS.local.md chunks, measured over their line ranges in this tree (awk byte-lengths — C-locale `length()`, not `wc -m` counts): §4.0 (L182-238) ≈ 7,024; §4.1 (L239-292) ≈ 11,283; §2.0 (L91-120) ≈ 3,967; References (L482-500) ≈ 2,761 — together ≈ 25,035. Under the leader-corrected size model (decision-index), the four-chunk move covers ≈25,035 of the 112,650-char required always-loaded reduction; the residual (≈87,615) is card t1887's core-rules diet share. (Operator memo figures — 6,215 / 9,881 / 4,011 / 2,633 — were measured on the primary copy; both attributions are recorded.) | awk char-length over the named ranges (this run); decision-index measurement table (HEAD `9b1dcde024`) |
| MF-7 | Template-First mirror inventory: `AGENTS.md` ↔ `internal/template/templates/AGENTS.md.tmpl` (371 lines) and `.claude/rules/moai/workflow/factory-dispatch*.md` ↔ `internal/template/templates/.claude/rules/moai/workflow/` (all six files present) are mirrored; `AGENTS.local.md`, `.claude/rules/local/*`, and the six `.moai/docs/*` targets have NO template mirror (`internal/template/templates/.moai/docs/` carries six unrelated files; `templates/.claude/rules/local/` does not exist). Template edits ride the `make build` cycle. | `ls internal/template/templates/.claude/rules/moai/workflow/`; `ls internal/template/templates/.moai/docs/` |
| MF-8 | `AGENTS.md` = 367 lines / 23,780 bytes; `internal/template/templates/AGENTS.md.tmpl` = 371 lines / 23,784 bytes — maintained in parallel (the committed review patch edits both with parallel hunks). | `wc -c -l` (this run) |
| MF-9 | SPEC ID uniqueness: `SPEC-AGENTS-CONTRACT-REVISE-001` appears only as this SPEC's own (untracked) directory with `inputs/`; no other SPEC carries the ID. Nearest neighbors are unrelated (`SPEC-AGENTS-MD-CANON-001`, `SPEC-AGENTS-IGNORE-POLICY-001`, `SPEC-AGENTS-WORKTREE-ROW-001`). | `ls .moai/specs/` |

## §D Policy Baseline (operator-approved 2026-10-19)

The revision shall reflect, without re-litigating:

1. Per-card fast PR flow: short card branch/worktree → local verification → fast draft PR → real Codex review → safe handoff → next independent card in a fresh worktree.
2. `main` direct-push protection stays.
3. Wait only on real dependencies; isolate failures; return leases and test slots.
4. The leader is coordination/supervision only — never implements or tests cards itself.
5. Primary `main` carries existing changes preserved (t1890 context).
6. Deployment default is manual with staged independent opt-in; distinguish the local factory override from Web/executed config consistency (t1895). Capabilities not yet implemented are written ONLY as tasks/limitations — never described as implemented.
7. Auto-merge policy (source: `inputs/REVIEW.md:11`): conditional auto-merge applies only to a non-fork, non-bot PR authored by the verified connected GitHub user — identity looked up from the actual connected account, never a hardcoded product login. Other-author, fork, and bot PRs must not have auto-merge enabled and require BOTH the user's test verification confirmed AND explicit approval — both together, never one alone. Eligible own PRs still keep latest-SHA required CI, real Codex review, resolved findings, protection rules, and any third-party approvals. The repository stays public and automatic branch deletion is forbidden, own PRs included (ties to the `--delete-branch` / `delete_branch_on_merge` mismatch limitation). This is a local operation choice — not approval to change the product-wide manual default.
8. A Kanban simulation is never treated as implementation approval.
9. AGENTS.local.md size diet: [HARD] obligations stay in the body as one-line essence; procedures, examples, and references move to paths:-scoped companion docs; deleting an obligation is forbidden; every moved clause leaves a pointer. ~~Target ≤ 12,000 chars~~ [SUPERSEDED by leader correction 2026-10-19: the ≤12,000 figure came from a bad baseline (primary copy 1,587 commits behind main); the governing target is the always-loaded instruction total ≤ 140,000 chars (hard 150,000) with AGENTS.local.md's share derived — see decision-index measurement table. The four-chunk move is unaffected.]
10. Distinguish precisely: policy approval (2026-10-19) vs tool application. The factory-complete PR path lands with card t1882 — until then the flow is the manual `gh pr create --draft` path. The Codex review gate lands with card t1868 (in progress) — until then the gate is approved policy plus an in-progress task.
11. The batch release procedure (`release/main-batch-YYYYMMDD`, merge-commit) stays until #1830 processing; the merge-flag rule is release/* OR hotfix/* → `--merge`, else `--squash` (MF-2).

## §E Requirements (GEARS)

### REQ-ACR-001 — §4.0 current card-delivery policy

**While** a card's implementation is complete on its branch, the revised `AGENTS.local.md` §4.0 shall direct the officially assigned worker to finish scoped local verification, open a draft PR for the exact verified head, record the handoff (head SHA, evidence, review/repair owner, lease and test-slot disposition), and shall require receiver-side confirmation of that handoff before the worker takes the next independent card in a fresh worktree. The revised §4.0 shall state that opening the PR is not card completion — closure requires the confirmed remote merge and the required completion evidence — and that a preservation push of a card branch by its assigned worker is in-scope delivery, not a protocol violation.

### REQ-ACR-002 — Leader boundary

**The revised `AGENTS.md` §0, the `agent-spawning` capability-table row, and `AGENTS.local.md` §16 shall state the leader boundary**: a session assigned as leader or coordinator coordinates, delegates, reads evidence, and supervises, and shall not implement, repair, build, or run tests in any worktree. The revised text shall not carry an emergency, trivial-change, worker-absence, or explicit-request exception, and shall replace the "Permitted direct execution" section with the assigned-worker boundary.

### REQ-ACR-003 — Review contract rewrite

**When** a PR's required review is evaluated, the revised `AGENTS.md` §4 clause (and its template mirror) shall require the review to be observed on the PR's current head SHA from the configured review producer with its completed result and evidence URL, and shall not name CodeRabbit as the sole review producer. The revised clause shall cross-reference the standing Codex contract documents rather than restate their measurement recipe.

### REQ-ACR-004 — Review-gate policy block (incident-derived)

**The revised flow documents shall state, as approved policy, the review-gate rules** derived from the #1841 incident (§B), with observed triggers and rules recorded separately:

**Observed triggers and re-review behavior** (measured; PR-level observations attributed to the operator/leader board correction of 2026-10-19):

- (a) On PR #1841, the ready transition started the review.
- (b) On merged PRs #1831, #1832, #1834, an `@codex review` comment started the review (#1832 and #1834 have recorded results) — manual triggering is an observed trigger as well, and the documents shall not assert that manual triggering from draft is impossible.
- (c) On PRs #1836 and #1842, the head changed after ready with no automatic re-review observed.
- (d) Official Codex review triggers, per the Codex summary comment on PR #1844 (observed 2026-10-19T06:05Z): (i) opening a PR for review, (ii) changing a draft to ready, (iii) an `@codex review` comment, (iv) an `@codex security review` comment; an in-progress review shows 👀 and finishing without findings shows 👍.

**Rules:**

1. A merge is allowed only after confirming the **current** head SHA's `codex-review-state` two legs (`code_review` + `security_review`) are both `completed` with zero unresolved findings; the state set includes `Failed` (execution failure) — any state other than `completed` is not a pass and is a re-request target.
2. **When** the head SHA changed after review, the review shall be re-requested before merge.
3. An unprotected integration base is no exception to any of the above.
4. The leader is coordination/supervision only — it does not implement or test — and the #1841 merge-before-review-completion violation is the concrete failure this rule prevents.

**While** the gate implementation belongs to card t1868 (lane-22, in progress), the documents shall describe the gate as approved policy plus an in-progress implementation task and shall not describe it as already-implemented tooling.

### REQ-ACR-005 — Superseded-flow confinement

**The revised flow documents (`AGENTS.local.md` §4.0/§4.1, `.claude/rules/local/repo-local-pr-policy.md`, `.claude/rules/local/gitflow-lane-protocol.md`, the factory-dispatch rule family) shall describe the per-card draft-PR delivery as current card procedure, and shall confine the local-main-landing-mandatory model and the lanes-do-not-push rule to transition-scoped historical statements.** The historical statements shall preserve the prior decisions and evidence without execution authority over current card delivery; preservation, one-writer, ownership, and resource-lease safeguards shall remain binding.

### REQ-ACR-006 — Batch-release retention and merge-flag accuracy

**The revised documents shall retain the batch release procedure (`release/main-batch-YYYYMMDD`, merge-commit, leader-executed) as the current release-path procedure until #1830 processing**, and every merge-flag statement shall match the measured rule — release/* OR hotfix/* → `--merge`, else `--squash` (MF-2). A merge-flag statement that omits `hotfix/*` shall not land.

### REQ-ACR-007 — Policy-vs-implementation distinction

**Where** a documented capability has no landed implementation, the revised documents shall record it only as an approved policy plus an in-progress task or limitation, and shall not claim tool support that has not landed. **While** card t1882 has not landed the factory-complete PR path, the flow documents shall describe `gh pr create --draft` (manual) as the operative delivery path. **While** card t1895 has not landed the config-defaults change, the documents shall state the manual-by-default / staged-independent-opt-in deployment policy without describing it as the shipped default.

### REQ-ACR-008 — Auto-merge mechanism accuracy and conditions

**The revised documents' auto-merge description shall match the measured mechanism** — `workflow_run` CI-completed trigger, no draft filter, no configuration gate, `--match-head-commit` (MF-1) — **and shall state the auto-merge conditions in full**: conditional auto-merge only for a non-fork, non-bot PR authored by the verified connected GitHub user, identity looked up from the actual connected account (never a hardcoded product login); other-author, fork, and bot PRs shall not have auto-merge enabled and shall require BOTH the user's test verification confirmed AND explicit approval — both together, never one alone; eligible own PRs still keep latest-SHA required CI, real Codex review, resolved findings, protection rules, and any third-party approvals; the repository stays public and automatic branch deletion is forbidden, own PRs included. The documents shall record the active `--delete-branch` behavior (MF-3) against the no-automatic-branch-deletion policy as a recorded limitation, and shall state that the conditional own-PR setting is a local operation choice — not a change to the product-wide manual default (REQ-ACR-007). The documents shall not describe the mechanism with a draft filter or a configuration gate it does not have, and shall not present either of the two non-own-PR conditions (test verification, explicit approval) as sufficient alone.

### REQ-ACR-009 — AGENTS.local.md size diet (leader-corrected target)

**When** `AGENTS.local.md` is restructured, the governing target shall be the Claude Code always-loaded instruction TOTAL: ≤ 140,000 chars projected (hard limit 150,000), measured by `wc -m` over the always-loaded surfaces at the card's tree — with `AGENTS.local.md`'s share derived from that total rather than fixed, and card t1887's core-rules diet share accounted in the projection (decision-index measurement table; leader correction of 2026-10-19 — the earlier ≤12,000 figure was derived from a bad baseline and is withdrawn as a governing target). The four named chunks (§4.0, §4.1, §2.0, References — ≈25,035 awk-bytes at this tree) shall move into paths:-scoped companion docs regardless of the derived share. Every moved [HARD] clause shall leave a one-line pointer at its original location; a moved-clause reconciliation table (original section → new location) shall accompany the change; each new companion document shall be paths:-scoped such that its `paths:` glob verifiably loads for the files it claims to scope. **The revision shall preserve every [HARD] obligation as one-line body essence or pointer and shall not delete or weaken an obligation.** Compression beyond the four named chunks proceeds only as recorded in decision-index ("리더 정정" — leader correction, not operator-set).

### REQ-ACR-010 — Pointer and mechanism fixes

**The revised `AGENTS.local.md` shall point** the template-source list at `internal/template/templates/AGENTS.md.tmpl` (not `templates/CLAUDE.md`, MF/§B L70), the agent-catalogue self-check at `.moai/config/sections/delegation.yaml` and the current agent definitions (not "CLAUDE.md Section 4", §B L446), and the quick-start/embed description at the tracked per-file embed allowlist (`embed_manifest_gen.go`) with the explicit emit targets (`make agents-emit`, `make commands-emit`, `make embed-manifest`) plus `make build` as drift-check-and-compile — matching the measured code.

### REQ-ACR-011 — Historical marking of git doctrine documents

**The revised `.moai/docs/git-workflow-doctrine.md` and `.moai/docs/git-local-workflow-doctrine.md` shall mark** the historical branch-protection mutation examples, stash/reset recovery recipes, and hook-bypass alternatives as historical records without execution authority, and shall state that a historical protection receipt is not evidence of current remote settings and does not authorize changing those settings.

### REQ-ACR-012 — Update-preservation documentation

**The revised update-survival documents (`.moai/docs/update-local-file-survival.md`, `.moai/docs/harness-namespace-doctrine.md`, `.moai/docs/version-management.md`) shall present** the current classification/preservation contract as operative and shall confine the historical wipe/wholesale-restore procedures to incident-record framing without execution instructions. The `git-strategy.yaml` reset note shall be rewritten to the preservation contract (operator values and local factory override preserved, verified after update).

### REQ-ACR-013 — Primary-main preservation guidance

**The revised primary-main guidance (§4.0 and the local rules) shall require** inventory, preservation, and an explicit handoff plan before any recovery of the shared primary checkout, and shall not instruct `checkout` / `reset` / `stash` / force-checkout to align the primary with `main`. A clean dedicated main tree shall not be presented as proof the shared primary is safe to move.

### REQ-ACR-014 — Template-First mirror parity

**When** `AGENTS.md` changes, the same change shall land in `internal/template/templates/AGENTS.md.tmpl` and pass the template build cycle (`make build` exit 0). **When** the factory-dispatch rule family changes, its mirrors under `internal/template/templates/.claude/rules/moai/workflow/` shall change identically (MF-7). **The template mirrors carry producer-neutral rules only — incident references, card-id markers, and in-progress/status markers stay LOCAL-ONLY and never enter the template mirror** (C9 card-id leak guard: `\bt[0-9]{3,4}\b`, removal-only per-(file, literal) baseline, `internal/template/card_id_leak_test.go:33-56`). The card shall produce a changed-file + diff list covering every touched document and its template mirror.

### REQ-ACR-015 — Kanban prohibition

**The revised documents shall not describe a Kanban simulation as implementation approval, and shall not authorize Kanban-mode implementation or switching as part of this revision.**

### REQ-ACR-016 — CLI vs slash-command form distinction

**When** a revised document names the automated card-dispatch forms, it shall distinguish the CLI shell command `moai todo --auto` from the slash-skill commands `/moai todo --auto-lane` and `/moai todo --auto-leader`, and shall not write `moai todo --auto-lane` as a shell form. Code-side occurrences (SessionStart hook, codex launcher) are card t1695's delegated scope — out of this card's file changes (§H).

## §F Acceptance Criteria Overview

Tier M: 16 requirements → 16 acceptance criteria, 1:1 (`acceptance.md`, AC-ACR-001 … AC-ACR-016). The completion criteria the card must encode: (a) conflict grep with zero unsuperseded hits (AC-ACR-005), (b) changed-file + diff list per touched doc and mirror (AC-ACR-014), (c) receiver-confirmation requirement present in the flow docs (AC-ACR-001), (d) the leader-corrected size model — always-loaded total projection ≤ 140,000 with the decision-index measurement table and the four-chunk move executed (AC-ACR-009).

## §G Review-Proposal Disposition (adopt / adapt / reject)

Verified against the code and current docs of this tree. Tally: **ADOPT 6 · ADAPT 3 · REJECT 2** (11 groups). The review's 11 unapplied suggestions are not among the committed inputs and are out of scope (§H).

| # | Proposal group (review priority) | Disposition | Reason, verified in this tree |
|---|----------------------------------|-------------|-------------------------------|
| 1 | Leader role boundary; no direct-execution exception (High; `AGENTS.md:10,34`, `AGENTS.local.md:441-460`) | **ADOPT** | Matches policy 4; conflict sites verified (§B). Patch hunks are a sound basis. |
| 2 | Replace local-main batch/card-PR prohibition with draft-PR flow (High; `AGENTS.local.md:182-237`, `repo-local-pr-policy.md:10-14`, `gitflow-lane-protocol.md:83-115`) | **ADAPT** | Adopt the direction and the patch's §4.0/§4.0-H structure; the patch's wholesale supersession is narrowed — the batch-release procedure (§4.0 items 6-7) stays current for the release path per policy 11 and MF-2. |
| 3 | `factory_card_pr.go` missing `--draft` + `--auto` merge (High) | **REJECT for this card** | Implementation scope of card t1882 (MF-4). Docs record the limitation (REQ-ACR-007); code is untouched here. |
| 4 | CodeRabbit-only review contract → Codex review contract (High; `AGENTS.md:200-205`, tmpl 202-207) | **ADAPT** | Adopt with producer-neutral wording cross-referencing the standing contract docs (no duplication of t1874's contract surface), extended by the incident-derived gate rules of REQ-ACR-004. |
| 5 | PR submission is not done; closure after remote merge + evidence (High) | **ADOPT** | Matches policy 1/3 and the patch's handoff block; verified against `factory_card_pr.go` pr-open/merged-pr transitions (read L390-460). |
| 6 | Primary-main recovery: preservation first, no force-alignment (High) | **ADOPT** | Matches policy 5; patch item 6 adopted. |
| 7 | `defaults.go` config defaults audit (High; `internal/config/defaults.go:1333…1505`) | **REJECT for this card** | Card t1895's implementation scope (policy 6/10). Docs state the manual-default policy + limitation only; defaults code untouched here. |
| 8 | Nonexistent template `CLAUDE.md` / "CLAUDE.md Section 4" pointers (Medium; `AGENTS.local.md:70,446`) | **ADOPT** | Both pointers verified stale in this tree (§B); `templates/CLAUDE.md` absent, `AGENTS.md.tmpl` present. |
| 9 | Embed/build description drift (Medium; `local-dev-guide.md:111-125`, `AGENTS.local.md:50,149`) | **ADOPT** | Adopt the allowlist + explicit-emit wording; `make build` is drift-check-and-compile, not a regenerator of agent/command outputs. |
| 10 | Update-preservation contract vs historical wipe guidance (Medium) | **ADOPT** | Current preservation contract becomes operative; historical procedures become incident records (REQ-ACR-012). |
| 11 | Historical branch-protection/bypass examples (Medium; `git-local-workflow-doctrine.md:86-153`, `git-workflow-doctrine.md:253-307,367-375`) | **ADAPT** | Adopt the patch's header notices and add in-body historical scoping so the AC-ACR-005 conflict grep passes; history is preserved, not deleted. |

## §H Out of Scope

### Out of Scope — Go implementation of delivery tooling (t1882)

- No change to `internal/cli/factory_card_pr.go` or any generation/update Go code; the draft/auto-merge implementation belongs to card t1882.
- No documentation claim that `moai factory complete` opens draft PRs before t1882 lands.

### Out of Scope — Codex review-gate implementation (t1868)

- No implementation, contract-code change, or gate wiring for the `codex-review-state` merge gate; card t1868 (lane-22, in progress) owns it.
- The PR1828 gate's `reviews[]` APPROVED contract fix belongs to t1868; this SPEC only aligns the prose.

### Out of Scope — Configuration defaults change (t1895)

- No change to `internal/config/defaults.go` or shipped config defaults; manual-by-default is documented as policy, not switched on here.

### Out of Scope — Kanban mode

- No Kanban implementation or mode switching; a Kanban simulation is never treated as implementation approval.

### Out of Scope — Code-side dispatch strings (t1695)

- No changes to SessionStart hook or codex launcher strings for the `--auto` / `--auto-lane` / `--auto-leader` forms; card t1695 (lane-32) owns the code-side occurrences and is noted as delegated wherever the docs distinction is drawn.

### Out of Scope — Remote and management settings

- No branch-protection, visibility, credential, app-access, or cost-commitment changes; historical protection receipts are not authorization.
- No automatic branch deletion changes (`--delete-branch` / `delete_branch_on_merge` mismatch is recorded as a limitation only).

### Out of Scope — Primary-checkout recovery execution

- No execution of primary-checkout cleanup; inventory/preservation/handoff language only.

### Out of Scope — Unapplied review suggestions

- The 11 suggestions the reviewer left unapplied are not among the committed inputs (`inputs/` carries REVIEW.md, proposed.patch, REPORT.md, native-codex-rest-comment.json only); they are neither adopted nor rejected here and need their own evidence before any future adoption.

## §I Cross-References

- Inputs (committed evidence, read-only): `inputs/REVIEW.md`, `inputs/proposed.patch`, `inputs/REPORT.md`, `inputs/native-codex-rest-comment.json`.
- `AGENTS.local.md` §0.1/§0.2 (canonical-copy discriminator — the diet baseline follows it), `AGENTS.local.md` §4.0/§4.1 (revision targets).
- `.claude/rules/local/repo-local-pr-policy.md`, `.claude/rules/local/gitflow-lane-protocol.md` (local-only; no template mirror, MF-7).
- `.claude/rules/moai/workflow/factory-dispatch.md` (+ mechanics/cards/gates/detail — mirrored, MF-7).
- `SPEC-AGENTS-MD-CANON-001` (AGENTS.md canonical-copy rules), `SPEC-LOCAL-MAIN-FLOW-001` (the §4.0 basis this revision supersedes in part).
- Related cards: t1868 (Codex review gate, in progress), t1882 (factory-complete draft PR path), t1895 (config defaults), t1874 (Codex contract docs — non-duplication boundary), t1887 (core-rules diet — size-model co-owner), t1695 (code-side dispatch strings — delegated).
