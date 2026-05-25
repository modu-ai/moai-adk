# Agent Archive — 2026-05-25

**Date**: 2026-05-25
**Originating SPEC**: [SPEC-V3R6-AGENT-TEAM-REBUILD-001](../../specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md)
**Milestone**: M3 — Archive 12 phantom and domain-expert agents
**Rationale**: Anthropic 2026 catalog consolidation 17→8 retained agents

## Effective Archive Count

Plan documented 12 archives. **Effective count: 11** because `.claude/agents/agency/researcher.md` does not exist in this project (already removed in prior cleanup or never created). Per plan.md §D.6 variance accommodation, the archive proceeds with the 11 files that actually exist.

## Archive Structure

```
agent-archive-2026-05-25/
├── README.md                  (this file)
├── core/                      (4 archived manager agents)
│   ├── manager-strategy.md
│   ├── manager-quality.md
│   ├── manager-brain.md
│   └── manager-project.md
├── meta/                      (1 archived meta agent)
│   └── claude-code-guide.md
└── expert/                    (6 archived domain-expert agents)
    ├── expert-backend.md
    ├── expert-frontend.md
    ├── expert-security.md
    ├── expert-devops.md
    ├── expert-performance.md
    └── expert-refactoring.md
```

## Per-Agent Replacement Pattern

| Archived agent | Replacement |
|----------------|-------------|
| manager-strategy | absorbed into manager-spec (planning role consolidated into single planning agent) |
| manager-quality | absorbed into Stop hook `.claude/hooks/moai/sync-phase-quality-gate.sh` (M4) + manager-develop self-verification + plan-auditor + evaluator-active |
| manager-brain | retired — ideation handled by orchestrator + WebSearch + Skill("moai-foundation-thinking") |
| manager-project | absorbed into manager-docs (initialization role consolidated into single documentation agent) |
| claude-code-guide | retired — upstream Claude Code investigation handled by orchestrator + WebSearch/WebFetch on demand |
| expert-backend | per-spawn specialization via `Agent(subagent_type: "general-purpose")` with backend role profile |
| expert-frontend | per-spawn specialization via `Agent(subagent_type: "general-purpose")` with frontend role profile |
| expert-security | per-spawn specialization via `Agent(subagent_type: "general-purpose")` with security role profile |
| expert-devops | per-spawn specialization via `Agent(subagent_type: "general-purpose")` with devops role profile |
| expert-performance | per-spawn specialization via `Agent(subagent_type: "general-purpose")` with performance role profile |
| expert-refactoring | per-spawn specialization via `Agent(subagent_type: "general-purpose")` with refactoring role profile |

## Anthropic 2026 Best-Practice Alignment

The archive resolves three architectural violations identified in the deep audit (`research.md` §Anthropic 2026 verbatim):

1. **"Subagents cannot spawn other subagents"** — the manager-strategy → manager-develop hierarchical chain was architecturally impossible; consolidating planning into manager-spec eliminates the chain.
2. **"3-5 teammates is the typical effective range"** — MoAI's 17-agent catalog was 2-5× over the recommended ceiling; the 8 retained agents (manager-spec, manager-develop, manager-docs, plan-auditor, evaluator-active, builder-harness, manager-git, Explore) fit within the recommendation.
3. **"Define when to keep spawning a subagent"** — 12 of the 17 agents had 0 invocations across the 4 most recent SPEC runs (phantom agents); archiving them removes the phantom problem.

## Template Mirror Handling

Template mirror archive handling is **deferred to M8** per the `moai update` namespace protection contract (CLAUDE.local.md §24). The 12 phantom files currently in `internal/template/templates/.claude/agents/` (core/4 + meta/1 + agency/1 + expert/6) will be **deleted** in M8 — the archive directory itself is **NOT** mirrored to the template.

Rationale: `.moai/backups/` is a project-local artifact, not template content. Mirroring archive directories would pollute new projects created via `moai init` with stale historical context they cannot reason about.

## Git History Preservation

All 11 archived files retain full commit history via `git mv` (not `mv` + `git add`). Use `git log --follow <archived-path>` to trace the file's evolution prior to archival.

Example:
```bash
git log --follow --oneline .moai/backups/agent-archive-2026-05-25/core/manager-strategy.md
```

## Restoration Procedure (if needed)

To restore an archived agent (e.g., if a future SPEC requires re-activation):

```bash
git mv .moai/backups/agent-archive-2026-05-25/core/manager-strategy.md \
       .claude/agents/core/manager-strategy.md
```

Then update the orchestrator's agent catalog references and re-add corresponding hook routing as needed.

## Related SPECs

- **SPEC-V3R6-AGENT-TEAM-REBUILD-001** — this archive's originating SPEC (M3)
- **SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001** — adjacent SPEC that introduced `.claude/agents/local/` namespace for dev-only agents
- **SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001** — predecessor SPEC that established GEARS notation as canonical (foundation principle layer)

---

**Audit-Ready**: This archive is reproducible from git history. Commit `<M3-SHA>` (see `git log -- .moai/backups/agent-archive-2026-05-25/`) captures the atomic transition.
