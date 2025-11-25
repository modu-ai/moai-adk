---
name: moai-core-claude-code
description: Canonical Claude Code skill authoring kit covering agent Skills, sub-agent templates, and best practices aligned with official documentation.
version: 1.0.0
modularized: false
tags:
  - claude-code
  - skills
  - sub-agents
  - best-practices
updated: 2025-11-25
status: active
---

# Claude Code Skills & Sub-Agent Authoring Kit

Unified reference for creating and maintaining Claude Code Skills and sub-agents that comply with official guidance.

**Official References**:  
- Agent Skills: https://code.claude.com/docs/en/skills  
- Skills best practices: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices  
- Sub-agents: https://code.claude.com/docs/en/sub-agents  
- Anthropic Engineering note: https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills

## Quick Reference (30 seconds)
- Use **kebab-case** names (max 64 chars), concise description with trigger scenarios.
- Store Skills in `~/.claude/skills/` (personal) or `.claude/skills/` (project); plugins are lowest priority.
- Keep SKILL.md under ~500 lines with **Progressive Disclosure** (Quick → Patterns → Examples → Troubleshooting).
- Declare minimal `allowed-tools`; avoid broad tool grants.
- Sub-agents must be invoked via **Task(subagent_type=...)** and cannot spawn other sub-agents.

## Agent Skill Template (Starter)
```markdown
---
name: my-skill
description: What the skill does and when to trigger it (keep specific)
version: 1.0.0
allowed-tools: Read, Edit, Bash, Grep   # minimal set only
tags: [domain, claude-code]
updated: 2025-11-25
status: active
---

# Title
**Quick What/When**: 2–3 lines on scope and trigger phrases.

## Quick Reference
- One responsibility, one clear trigger.
- Examples of when to auto-load.

## Patterns
- Core patterns (3–5) with code snippets.

## Examples
- 2–3 runnable examples with prompts and expected outputs.

## Troubleshooting
- Common issues, validation commands, how to debug Skill loading.
```

## Sub-Agent Template (Official Fields)
```markdown
---
name: backend-designer
description: Designs REST/GraphQL APIs with performance guardrails
tools: Read, Edit, Bash, Grep
model: sonnet   # sonnet/opus/haiku/inherit
permissionMode: default   # default/acceptEdits/dontAsk
skills: moai-core-claude-code, moai-core-agent-factory
---

# System Prompt
- Purpose: what this agent owns (single domain).
- Triggers: specific tasks it should handle.
- Guardrails: no sub-agent spawning; use Task() for delegation.
- Examples: 2–3 short command-to-action pairs.

## Invocation (required)
Task(subagent_type="backend-designer", description="Design REST API for orders service")
```

## Best Practices (from official docs)
1) **Single responsibility**: clear scope and domain; avoid overlapping names.  
2) **Specific descriptions**: include trigger phrases (e.g., “API schema”, “performance audit”).  
3) **Minimal tools**: principle of least privilege; avoid `Write` unless necessary.  
4) **Progressive disclosure**: keep SKILL.md concise; move deep dives to modules/examples.  
5) **Validation**: run `claude --debug` to check YAML/frontmatter; keep under 500 lines.  
6) **Versioning**: update `version`/`updated` fields; commit `.claude/skills` to VCS.  
7) **Sub-agent rules**: no nested sub-agents; always use Task() delegation.  
8) **Storage priority**: personal overrides project overrides plugins; avoid duplicate names.

## Checklists
- [ ] Name/description specific and non-overlapping.  
- [ ] Minimal `allowed-tools` declared.  
- [ ] Examples in English with language identifiers.  
- [ ] Sub-agent Task() invocation included.  
- [ ] Official links referenced for users.  
- [ ] Updated date/version set.

## Troubleshooting
- Skill not loading: validate frontmatter (YAML), check path `.claude/skills/<name>/SKILL.md`.
- Unexpected tool prompts: review `permissionMode` and `allowed-tools`.
- Duplicate triggers: rename Skill or tighten description to avoid collisions.
