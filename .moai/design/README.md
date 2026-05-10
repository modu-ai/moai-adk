# .moai/design/

This directory holds human-authored design context for the `/moai design` workflow (Path B,
code-based brand design). Agents load these files as constitutional input before generating
copy, design tokens, and frontend code.

## Attachment Targets (Auto-load)

When `/moai design` runs, the following files are automatically attached in priority order:

**Auto-load priority**: `spec.md > system.md > research.md > pencil-plan.md`

Files not in this list (e.g. `wireframes/`, `assets/`) are excluded from auto-attach.

## File Roles

| File | Role | Editable by humans |
|------|------|--------------------|
| `research.md` | Competitor analysis, user insights, adopt/avoid patterns | YES |
| `system.md` | Design tokens, typography, spacing, accessibility rules | YES |
| `spec.md` | Information architecture, frame-by-frame spec | YES |
| `pencil-plan.md` | Pencil MCP batch operation plan (optional) | YES |
| `wireframes/` | PNG/SVG human reference images | YES |

## Reserved Filenames (Auto-generated — Do NOT Create Manually)

The following names are reserved for auto-generated artifacts from `moai-workflow-design-import`
and `manager-spec`. Creating user files with these names causes `moai update` to exit with an error.

**Exact reserved names:**
- `tokens.json`
- `components.json`
- `import-warnings.json`

**Glob reserved pattern:**
- `brief/BRIEF-*.md`

## Update Policy

- **Human-authored files** (`research.md`, `system.md`, `spec.md`, `pencil-plan.md`):
  `moai update` preserves user edits detected via SHA-256 hash mismatch.
- **Auto-generated files** (`tokens.json`, `components.json`, etc.):
  Overwritten by import tools on each run. Do not edit manually.
- **Re-init safety**: If `.moai/design/` already contains any regular file,
  `moai init` logs a warning and skips template deployment.
