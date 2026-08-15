---
description: "Repo-local PR policy override — moai-adk-go forces all-tier PR (Route B); Route A Hybrid-Trunk main-direct is disabled by branch protection"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/skills/moai/workflows/moai.md"
---

# Repo-Local PR Policy (moai-adk-go maintainer override)

[HARD] In THIS repository, the `spec-workflow.md` Route A ("Hybrid Trunk main-direct" for Tier S / M) is DISABLED. `main` is protected with `enforce_admins: true` + required PR (verified via `gh api repos/modu-ai/moai-adk/branches/main/protection`), so a direct push to `main` is rejected even for admins.

- ALL tiers (S / M / L) use **Route B (PR)**: work lands on a feature branch and merges via PR (self-merge allowed; 0 required approvals; CI must pass).
- The orchestrator MUST NOT instruct `manager-develop` to commit-and-push directly to `main`, regardless of Tier.
- This is a repo-local addendum (local-only, NOT mirrored to `internal/template/templates/`). The distributed template keeps the generic Route A / Route B choice for downstream users who have no branch protection.

Cross-reference: CLAUDE.local.md §23 (Hybrid Trunk RETIRED for this repo); `spec-workflow.md` § Route Selection.
