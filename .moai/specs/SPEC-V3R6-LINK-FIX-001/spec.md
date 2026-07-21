---
spec_id: SPEC-V3R6-LINK-FIX-001
title: Fix broken links — GitHub repo rename and stale Anthropic docs URL
status: completed
tier: S
created: 2026-07-19
updated: 2026-07-19
priority: medium
era: V3R6
---

# SPEC-V3R6-LINK-FIX-001: Fix broken links

## Background

Two categories of broken links exist across the repository:

1. **GitHub repo rename**: `modu-ai/moai-adk-go` → `modu-ai/moai-adk`. Four files contain old URLs

2. **Stale Anthropic docs URL**: `https://docs.anthropic.com/en/docs/claude-code` redirects away from the intended page. The canonical Claude Code documentation is now at `https://code.claude.com/docs/en`. Four README files need this fix.

Links to `adk.mo.ai.kr` pages that return 404 are excluded from scope: those are expected pre-deployment states, not broken link targets.

## Scope

### Working-Tree Changes
- `CONTRIBUTING.ko.md` — 3 GitHub URLs updated
- `CONTRIBUTING.md` — 3 GitHub URLs updated
- `docs-site/hugo.toml` — 2 GitHub URLs updated
- `docs-site/layouts/partials/site-header.html` — 1 GitHub URL updated

### Edit Required
- `README.md` line 558: replace stale Anthropic URL
- `README.ko.md` line 563: replace stale Anthropic URL
- `README.ja.md` line 558: replace stale Anthropic URL
- `README.zh.md` line 558: replace stale Anthropic URL

## Acceptance Criteria

- AC-1: No `github.com/modu-ai/moai-adk-go` URL remains in scope files
- AC-2: All four README files reference `https://code.claude.com/docs/en` (not `docs.anthropic.com`)
- AC-3: Changes committed and pushed to main
