---
description: "Repo-local policy override — moai-adk-go lands card work on local main (AGENTS.local.md §4.0): card work branches from local main and merges there with NO card-level PR; main direct push remains prohibited"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/skills/moai/workflows/moai.md"
---

# Repo-Local Branch Policy (moai-adk-go maintainer override)

[HARD] In THIS repository, direct push to `main` is DISABLED. `main` is protected with `enforce_admins: true` + required PR (verified via `gh api repos/modu-ai/moai-adk/branches/main/protection`), so a direct push to `main` is rejected even for admins. `main` advances ONLY through release pull requests (`release/vX.Y.Z` → `main`, merge-commit strategy; self-merge allowed at 0 required approvals once the required status checks pass).

[HARD] Card workflow (git-flow transition 2026-08-27; model: AGENTS.local.md §4.1; flow: AGENTS.local.md §4.0; rules: .claude/rules/local/gitflow-lane-protocol.md):
- Card worktrees branch from local `main` (never `origin/main`).
- Completed cards land on LOCAL `main` through the integration verb (`moai integration merge --card`; AGENTS.local.md §4.0 §2), never through a typed `git merge`. There are NO card-level PRs. Remote CI on `origin/main` is the verdict surface; lanes report their local merge SHA and do NOT push — the factory leader collects lane merge SHAs and closes the batch (`release/main-batch-YYYYMMDD` PR, AGENTS.local.md §4.0 §6–§7).
- The orchestrator MUST NOT instruct lane agents (`manager-develop` / `manager-docs` / per-spawn workers) to open card PRs against `main`, regardless of Tier.
- PR-based ceremony (spec-workflow Route B tier routing) applies ONLY to the release path above.

> ~~Prior policy (2026-07-20 → superseded 2026-08-26): ALL tiers (S / M / L) use Route B (PR): work lands on a feature branch and merges via PR~~ [RETIRED 2026-08-27 — replaced by the git-flow transition]

- This is a repo-local addendum (local-only, NOT mirrored to `internal/template/templates/`). The distributed template keeps the generic Route A / Route B choice for downstream users who have no branch protection.

Cross-reference: AGENTS.local.md §4.1; `.claude/rules/local/gitflow-lane-protocol.md`; `.moai/docs/git-workflow-doctrine.md` §18 / `.moai/docs/git-local-workflow-doctrine.md` §23 (header notices carry the transition framing).
