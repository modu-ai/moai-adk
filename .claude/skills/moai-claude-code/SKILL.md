---
name: moai-claude-code
description: Claude Code Creation and management of 5 components - Agent, Command, Skill, Plugin, Settings Precise file creation based on template
allowed-tools:
  - Read
  - Write
  - Edit
---

# MoAI Claude Code Manager

Create and manage Claude Code's five core components according to official standards.

## Support components

1. **Agent** (.claude/agents/) - Professional agent
2. **Command** (.claude/commands/) - Slash command
3. **Skill** (.claude/skills/) - Reusable function module
4. **Plugin** (mcpServers in settings.json) - MCP server integration
5. **Settings** (.claude/settings.json) - Permissions and hook settings

## Template Features

MoAI-ADK integrated production-grade templates (5)

- Fully detailed (complete and practical for use)
- MoAI-ADK workflow integration
- Copy-paste out of the box
- Includes verification and troubleshooting guide

## How to use

### Create Agent
"Please create a spec-builder Agent"

### Settings optimization
"Please create settings.json for Python project"

### Full Verification
"Please verify all Claude Code settings"

## Detailed Documentation

- **reference.md**: Component-specific writing guide
- **examples.md**: Collection of practical examples
- **templates/**: 5 production-grade templates
- **scripts/**: Python verification script (optional)

## How it works

1. Analyzing user requests → Identify component types
2. Select an appropriate template (templates/ directory)
3. Placeholder substitution and file creation
4. Automatic verification (optional, run scripts/)

## Core principles

- **Compliance with official standards**: Full compliance with Anthropic guidelines
- **Avoid hallucination**: Use only verified templates
- **Minimum privileges**: Specify only necessary tools
- **Security priority**: Manage sensitive information environment variables

## Official best practices (2024-12)

- Keep `SKILL.md` concise and push optional materials into referenced files (reference.md, examples.md, scripts/) so Claude loads only what it needs.
- Include trigger phrases in `description` (≤1024 chars) and use gerund/action names to improve automatic discovery.
- Link reference files directly from `SKILL.md` and avoid multi-hop references; aim for ≤1 depth for progressive disclosure.
- Test skills with Haiku, Sonnet, and any target models to confirm instruction density is appropriate.
- Store each skill as `.claude/skills/{skill-name}/SKILL.md`; Claude scans only first-level directories under `skills/`.

---

**Official Documentation**: https://docs.claude.com/en/docs/claude-code/skills
**Version**: 1.0.0
