---
description: "MoAI-ADK production release via Enhanced GitHub Flow (CLAUDE.local.md §18). Creates release/vX.Y.Z branch, version bump, CHANGELOG (bilingual), PR to main, merge commit (NOT squash), then scripts/release.sh for tag + GoReleaser. Hotfix support via --hotfix flag. All git operations delegated to manager-git. Quality failures escalate to expert-debug."
argument-hint: "[VERSION] [--hotfix] - optional target version (e.g., 2.1.0). If omitted, prompts for patch/minor/major selection."
type: local
allowed-tools: Skill
disable-model-invocation: true
version: 6.0.0
metadata:
  release_target: "production"
  branch: "main"
  tag_format: "vX.Y.Z"
  changelog_format: "english_first_bilingual"
  release_notes_format: "bilingual"
  git_delegation: "required"
  quality_escalation: "expert-debug"
  workflow_model: "enhanced-github-flow"
  merge_strategy: "merge-commit"
  reference_policy: "CLAUDE.local.md §18"
---

Use Skill("moai-workflow-release") with arguments: $ARGUMENTS
