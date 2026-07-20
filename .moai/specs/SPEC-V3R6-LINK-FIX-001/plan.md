---
spec_id: SPEC-V3R6-LINK-FIX-001
title: Fix broken links — GitHub repo rename and stale Anthropic docs URL
status: draft
phase: plan
created: 2026-07-19
updated: 2026-07-19
---

# Plan: SPEC-V3R6-LINK-FIX-001

## Milestones

### M1: Stage working-tree changes
- `git add CONTRIBUTING.md docs-site/hugo.toml docs-site/layouts/partials/site-header.html`
- `CONTRIBUTING.ko.md` is already staged — no action needed

### M2: Fix stale Anthropic URL in README files
Replace `https://docs.anthropic.com/en/docs/claude-code` with `https://code.claude.com/docs/en`:
- `README.md` line 558
- `README.ko.md` line 563
- `README.ja.md` line 558
- `README.zh.md` line 558

### M3: Commit and push
```
git add README.md README.ko.md README.ja.md README.zh.md
git commit -m "fix(docs): fix broken links — repo rename and stale Claude Code URL"
git push origin main
```

## Implementation Notes
- M1 files: content already correct, staging only
- M2 is a single-line text replacement in each of 4 files
- No tests required for URL-only changes
- Push to main (Hybrid Trunk Tier S — no PR branch)
