---
description: "Repo-local card delivery policy (moai-adk-go) — per-card draft-PR flow with the current-head review gate, conditional own-PR auto-merge, and the coordination-only leader boundary; direct main push remains prohibited"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/skills/moai/workflows/moai.md"
---

# Repo-Local Branch Policy (moai-adk-go maintainer override)

> **2026-10-19 revision (SPEC-AGENTS-CONTRACT-REVISE-001).** The card-delivery model this file
> used to state — cards merged on local `main` with no card-level PRs, lanes reporting merge
> SHAs without pushing, the leader collecting merge SHAs and closing batches — is superseded
> for card delivery and retained only as the transition-scoped historical record at the bottom
> of this file. Preservation, one-writer, ownership, and resource-lease safeguards are
> unchanged.

[HARD] Current card delivery is the per-card draft-PR flow of `AGENTS.local.md` §4.0: an
officially assigned worker uses a fresh card worktree, finishes scoped local verification,
opens a draft PR for the exact verified head — `gh pr create --draft` is the operative
delivery path; the draft/auto-merge separation in `moai factory complete` is an in-progress
implementation task (card t1882, not landed) and is not described here as shipped — and records a safe handoff (head
SHA, evidence, review/repair owner, lease and test-slot disposition). The next independent
card follows only after receiver-side confirmation of that handoff. Opening the PR is not card
completion: closure requires the confirmed remote merge and the required completion evidence.
A preservation push of a card branch by its assigned worker is in-scope delivery, not a
protocol violation.

[HARD] Review gate (incident-derived, approved policy). A merge is allowed only after
confirming the current head SHA's `codex-review-state` two legs — `code_review` and
`security_review` — are both `completed` with zero unresolved findings; any other state
(including `Failed`) is not a pass and is a re-request target; when the head changed after the
review, the review is re-requested before merge (a completed review stays valid while the head
SHA is unchanged); an unprotected integration base is no exception. The full trigger narration
and incident evidence live in `AGENTS.local.md` §4.0 item 3. The gate implementation is an
in-progress task — this file states approved policy, not shipped tooling.

[HARD] Conditional auto-merge applies only to a non-fork, non-bot PR authored by the verified
connected GitHub user — identity resolved from the actual connected account, never a hardcoded
login. Other-author, fork, and bot PRs must not have auto-merge enabled and require BOTH the
user's test verification confirmed AND explicit approval — both together, never one alone.
Eligible own PRs still keep latest-SHA required CI, real Codex review, resolved findings,
protection rules, and any third-party approvals. The repository stays public and automatic
branch deletion is forbidden, own PRs included — the active `--delete-branch` /
`delete_branch_on_merge` behavior is a recorded limitation. This is a local operation choice,
not a change to the product-wide manual default.

[HARD] The leader coordinates and supervises only. It never implements, repairs, builds, or
runs tests in any worktree; emergency, trivial-change, worker-absence, and explicit-request
framings are not exceptions. Preserve an existing diff and its provenance before handing it to
the officially assigned worker; if none is available, report the assignment blocker.

[HARD] Direct push to `main` remains disabled. Repository-protection claims require fresh
remote evidence — a historical protection receipt is not evidence of current remote settings
and does not authorize changing those settings.

**Historical record — transition-scoped, no execution authority over current card delivery.**
~~[HARD] Card workflow (git-flow transition 2026-08-27; model: AGENTS.local.md §4.1; flow:
AGENTS.local.md §4.0; rules: .claude/rules/local/gitflow-lane-protocol.md): Card worktrees
branch from local `main` (never `origin/main`). Completed cards land on LOCAL `main` through
the integration verb (`moai integration merge --card`; AGENTS.local.md §4.0 §2), never through
a typed `git merge`. There are NO card-level PRs. Remote CI on `origin/main` is the verdict
surface; lanes report their local merge SHA and do NOT push — the factory leader collects lane
merge SHAs and closes the batch (`release/main-batch-YYYYMMDD` PR, AGENTS.local.md §4.0
§6–§7). The orchestrator MUST NOT instruct lane agents (`manager-develop` / `manager-docs` /
per-spawn workers) to open card PRs against `main`, regardless of Tier. PR-based ceremony
(spec-workflow Route B tier routing) applies ONLY to the release path above.~~
[RETIRED 2026-10-19 — superseded by the per-card draft-PR flow at the top of this file. The
integration verb and its window remain the mechanism behind the retained release-path
resynchronization (`AGENTS.local.md` §4.0-R); the batch release procedure stays the
release-path procedure until #1830 processing. The no-card-PR instruction to lane agents is
lifted: card delivery now IS a (draft) PR.]

~~Prior policy (2026-07-20 → superseded 2026-08-26): ALL tiers (S / M / L) use Route B (PR): work lands on a feature branch and merges via PR~~ [RETIRED 2026-08-27 — replaced by the git-flow transition, itself superseded 2026-10-19 by the per-card draft-PR flow]

- This is a repo-local addendum (local-only, NOT mirrored to `internal/template/templates/`).
  The distributed template keeps the generic Route A / Route B choice for downstream users who
  have no branch protection, and starts manual with staged independent opt-ins per the
  distributed automation policy.

Cross-reference: `AGENTS.local.md` §4.0 (current policy) · §4.0-R (release path) · §4.0-H
(historical block); `.claude/rules/local/gitflow-lane-protocol.md`;
`.claude/rules/local/agents-local-detail.md` (moved detail).
