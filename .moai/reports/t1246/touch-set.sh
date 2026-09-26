#!/bin/sh
# Regenerates the SPEC-AGENT-MODEL-INHERIT-001 touch set (card t1246).
# Run from the worktree root: sh .moai/reports/t1246/touch-set.sh > .moai/reports/t1246/touch-set.txt
# Read-only: greps only, writes nothing itself.
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
} | grep -vE '^\.claude/(agent-memory|worktrees)/|/testdata/' | sort -u
