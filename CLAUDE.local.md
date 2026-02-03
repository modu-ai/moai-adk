# Local Development Rules

## Protected Directories (DO NOT MODIFY)

The following directories are for local development only and must NEVER be modified by agents or automated processes:

- `.claude/` - Local Claude Code configuration (agents, skills, rules, hooks, output-styles, settings)
- `.moai/` - Local MoAI project state (config, specs, memory, logs, manifest)

These directories contain the developer's active working environment. All template changes must be made exclusively in `internal/template/templates/`.

## Template Development

- Source of truth for distributed templates: `internal/template/templates/`
- Changes to template content go ONLY to `internal/template/templates/`
- Never sync from `.claude/` to `internal/template/templates/` or vice versa
