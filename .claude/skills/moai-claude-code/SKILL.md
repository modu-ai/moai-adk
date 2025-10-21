---

name: moai-claude-code
description: Scaffolds and audits Claude Code agents, commands, skills, plugins, and settings with production templates. Use when configuring or reviewing Claude Code automation inside MoAI workflows.
allowed-tools:
  - Read
  - Write
  - Edit
---

# MoAI Claude Code Manager

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file) |
| Auto-load | SessionStart (Claude Code bootstrap) |
| Trigger cues | Agent/command/skill/plugin/settings authoring, Claude Code environment setup. |

Create and manage Claude Code's five core components according to official standards.

## Components covered

- **Agents** `.claude/agents/` — Persona, tools, and workflow definition
- **Commands** `.claude/commands/` — Slash command entry points
- **Skills** `.claude/skills/` — Reusable instruction capsules
- **Plugins** `settings.json › mcpServers` — MCP integrations
- **Settings** `.claude/settings.json` — Tool permissions, hooks, session defaults

## Reference files

- `reference.md` — Writing guide and checklist
- `examples.md` — Sample completed artifacts
- `templates/` — Markdown/JSON skeleton for five components
- `scripts/` — settings validation and template integrity check scripts

## Workflow

1. Analyze user requests to determine required components (Agent/Command/Skill/Plugin/Settings).
2. Copy the stub from `templates/` and replace the placeholders to suit your project context.
3. If necessary, run the `scripts/` verifier to check required fields, permissions, and follow-up links.

## Guardrails

- Maintain minimum privileges and progressive disclosure in line with Anthropic official guidelines.
- Rather than modifying the template directly, only update the hook/field guided by reference.md.
- The created files and settings.json are included in Git version management and a change history is left.

**Official documentation**: https://docs.claude.com/en/docs/claude-code/skills  
**Version**: 1.0.0

## Examples
```markdown
- Create a spec-builder agent and the /alfred command set in a new project.
- Review existing settings.json to update allowable tools and hook configurations.
```

## Best Practices
- The output template is designed to be idempotent so that it is safe even when reapplied.
- Detailed procedures are separated into reference.md (writing guide) and examples.md (sample artifacts) and loaded only when necessary.

## When to use
- Activates when someone asks to scaffold or audit Claude Code components.
- When bootstrapping a Claude Code configuration to a new project.
- When reexamining existing agents/commands/skills/plug-ins to meet standards.
- When verification of settings is required in initialization workflows such as `/alfred:0-project`.

## What it does
- Create and update the five core components as official templates.
- Verify accepted tools, model selection, and progressive disclosure links.
- Provides reusable stubs and verification procedures through templates/·scripts/ resources.

## Inputs
- User configuration requests (e.g. “add new command”, “review settings.json”) and current `.claude/` directory state.
- Project-specific template requirements or security/permissions policies.

## Outputs
- Official Markdown definition file under `.claude/agents|commands|skills/`.
- A summary of `.claude/settings.json` that reflects the latest settings and allowed tools and follow-up TODOs.

## Failure Modes
- If the template path or placeholder is different from the latest version, the result may be damaged.
- The settings.json permission policy may conflict with project rules or may block verification script execution.

## Dependencies
- When used with cc-manager, doc-syncer, and moai-foundation-git, the creation → verification → distribution flow is completed.
- You must have versioned resources in the templates/ and scripts/ directories for automation to work properly.

## References
- Anthropic. "Claude Code Style Guide." https://docs.claude.com/ (accessed 2025-03-29).
- Prettier. "Opinionated Code Formatter." https://prettier.io/docs/en/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added best practice structure to Claude code formatting skill.
