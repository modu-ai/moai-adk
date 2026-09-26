#!/bin/sh
# Regenerates the SPEC-AGENT-MODEL-INHERIT-001 touch set (card t1246).
# Run from the worktree root: sh .moai/reports/t1246/touch-set.sh > .moai/reports/t1246/touch-set.txt
# Read-only: greps only, writes nothing itself. Output is C-collated (LC_ALL=C) so the
# run-entry gate's `comm -12` must also run under LC_ALL=C.
D='Per-Spawn Model Injection|moai model profile|agent_model_guard|agent-model-audit|Effort-Level Calibration|profile matrix|Profile Resolver|ResolveAgentModelEffort|agent_overrides|harness_agents|model_reasoning_effort|workflow_agents|model_routing|performance_tier|--model-policy|Agent\([^)]*(model|effort) *:|[Pp]er-[Ss]pawn [Mm]odel|model profile|per-spawn models|moai (init|update)[^|]{0,60}--profile'
G='ResolveAgentModelEffort|ResolveHarnessAgentModelEffort|DefaultProfileMatrix|ProfileMatrixAgents|AgentGroup\(|ApplyProfile|ApplyPerformanceTier|AgentOverrides|HarnessAgents|EffectiveProfile|AgentModelGuard|agent_model_guard|agent-model-audit|WorkflowAgents|ModelRouting|model_routing|IsValidPerformanceTier|NormalizeToTier|agentfm|cellguard'
{
  grep -rlE '^(model|effort):' .claude/agents internal/template/templates/.claude/agents
  ls internal/template/templates/.codex/agents/moai/*.toml
  grep -rlE -- "$D" .claude internal/template/templates docs-site/content .moai/docs
  grep -rlE "$G" --include='*.go' --include='*.templ' internal cmd pkg
  grep -lE '\b(model|effort) *:' .claude/workflows/*.js internal/template/templates/.claude/workflows/*.js
  grep -rlE '"(model|effort)"' .claude/commands/harness
  ls .moai/docs/agent-lint.md internal/template/templates/.moai/docs/agent-lint.md \
    .claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md \
    internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md \
    internal/template/templates/.moai/config/sections/workflow.yaml CHANGELOG.md
  # plan-audit iter-2 additions (N2 rosterguard-bound project docs, N3 verify-judge channel,
  # N7 docs residue, N9 profile-setup wording, shipped key inventory fixture)
  ls .moai/project/product.md .moai/project/tech.md \
    internal/config/testdata/shipped_key_inventory.yaml \
    .claude/rules/moai/workflow/verify-judge-effort-contract.md \
    .claude/hooks/tests/test-judge-effort-contract.sh \
    internal/cli/wizard/translations.go internal/cli/profile_setup_translations.go \
    docs-site/content/en/advanced/harness-v4-builder.md docs-site/content/ja/advanced/harness-v4-builder.md \
    docs-site/content/ko/advanced/harness-v4-builder.md docs-site/content/zh/advanced/harness-v4-builder.md \
    docs-site/content/en/cost-optimization/_index.md docs-site/content/ko/cost-optimization/_index.md \
    docs-site/content/en/multi-llm/_index.md \
    docs-site/content/ko/claude-code/foundations/features-overview.md
} | grep -vE '^\.claude/(agent-memory|worktrees)/|/testdata/codex-rollouts' | LC_ALL=C sort -u
