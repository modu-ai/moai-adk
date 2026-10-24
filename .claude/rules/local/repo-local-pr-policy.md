---
description: "Repo-local policy override — moai-adk-go runs agentic-trunk (AGENTS.local.md §4.0): card work branches from local main; per-card delivery follows the target-holder routing (card pull request merged by the leader with no auto-merge request where only the primary holds main; integration-window local merge where a dedicated worktree holds it); main direct push remains prohibited"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/skills/moai/workflows/moai.md"
---

# Repo-Local Branch Policy (moai-adk-go maintainer override)

[HARD] In THIS repository, direct push to `main` is DISABLED. `main` is protected with `enforce_admins: true` + required PR (verified via `gh api repos/modu-ai/moai-adk/branches/main/protection`), so a direct push to `main` is rejected even for admins. `main` advances ONLY through release pull requests (`release/vX.Y.Z` → `main`, merge-commit strategy; self-merge allowed at 0 required approvals once the required status checks pass).

[HARD] Card workflow (agentic-trunk; rescope 0.4.0 of SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001, 2026-10-21; model: AGENTS.local.md §4.0; flow: AGENTS.local.md §4.0; rules: .claude/rules/local/gitflow-lane-protocol.md):
- Card worktrees branch from local `main` (never `origin/main`).
- Per-card delivery follows the target-holder routing: where only the primary checkout holds `main` (this repository — the 2026-10-21 rescope abolished the dedicated integration worktree), the completed card lands through its pull request, which the leader merges with NO auto-merge request; where a dedicated integration worktree holds `main`, the card lands through the integration verb (`moai integration merge --card`; AGENTS.local.md §4.0 §2), never through a typed `git merge`. On either route a lane does not merge its own card. Remote CI on `origin/main` is the verdict surface; lanes do NOT push — the factory leader merges card pull requests (or collects lane merge SHAs on the local-merge route) and closes the release batch per `lead_push_threshold` (`release/main-batch-YYYYMMDD` PR, AGENTS.local.md §4.0 §6–§7); the release itself is tagged.
- Card pull requests are opened by the tool path (`moai factory complete` PR edge), never hand-run by a lane. The orchestrator MUST NOT instruct lane agents (`manager-develop` / `manager-docs` / per-spawn workers) to open or merge card PRs against `main` by hand, regardless of Tier.
- PR-based ceremony (spec-workflow Route B tier routing) applies ONLY to the release path above.

> ~~Prior policy (2026-07-20 → superseded 2026-08-26): ALL tiers (S / M / L) use Route B (PR): work lands on a feature branch and merges via PR~~ [RETIRED 2026-08-27 — replaced by the git-flow transition]

- This is a repo-local addendum (local-only, NOT mirrored to `internal/template/templates/`). The distributed template keeps the generic Route A / Route B choice for downstream users who have no branch protection.

Cross-reference: AGENTS.local.md §4.1; `.claude/rules/local/gitflow-lane-protocol.md`; `.moai/docs/git-workflow-doctrine.md` §18 / `.moai/docs/git-local-workflow-doctrine.md` §23 (header notices carry the transition framing).
