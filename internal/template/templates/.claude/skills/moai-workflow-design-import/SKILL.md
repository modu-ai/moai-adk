---
name: moai-workflow-design-import
description: >
  Handles /moai design 3-path routing: Path A (Claude Design handoff bundle), Path B1
  (Figma extractor via meta-harness), Path B2 (Pencil MCP via meta-harness). Validates
  bundle version against supported_bundle_versions whitelist and returns structured error
  codes on failure. Depends on SPEC-V3R3-HARNESS-001 for Path B1/B2.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-27"
  tags: "design import, handoff bundle, claude design, design tokens, components, figma, pencil, meta-harness"
  related-skills: "moai-domain-brand-design, moai-workflow-gan-loop, moai-meta-harness"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["design import", "handoff bundle", "claude design", "bundle path", "design zip", "import bundle", "figma extractor", "pencil mcp", "path b1", "path b2"]
  agents: ["expert-frontend"]
  phases: ["run"]
---

# moai-workflow-design-import

Handles `/moai design` 3-path routing for design artifact ingestion. Each path produces
DTCG-validated design tokens at `.moai/design/tokens.json` for `expert-frontend` consumption.

**Path A** (Claude Design): Existing handoff bundle import — preserved verbatim.
**Path B1** (Figma): Meta-harness generates a project-scoped figma-extractor skill.
**Path B2** (Pencil): Meta-harness generates a project-scoped pencil-mcp skill.

Reserved output paths (design constitution §3.2): `tokens.json`, `components.json`,
`assets/`, `import-warnings.json`, `brief/BRIEF-*.md` — all written to `.moai/design/`.

---

## Path Selection

When `/moai design` requires path selection, present via AskUserQuestion:
1. **Path A — Claude Design** (권장): Handoff bundle (ZIP or HTML) from Claude Design.
2. **Path B1 — Figma**: Figma file access; meta-harness generates `my-harness-figma-extractor`.
3. **Path B2 — Pencil**: `.pen` files; meta-harness generates `my-harness-pencil-mcp`.

Selection is persisted to `.moai/design/path-selection.json` for audit.

Brand context (`.moai/project/brand/`) is the constitutional parent across all paths — no
path may override brand constraints (design constitution §3.1, §3.3).

---

---

## Path A — Claude Design Handoff Bundle

*This section is preserved verbatim from v1.0.0. Path A behavior is unchanged.*

### Supported Bundle Formats (Phase 1)

Primary supported formats:
- `ZIP`: Claude Design export containing `manifest.json`, `tokens.json`, `components/`, and `assets/`
- `HTML`: Single-file HTML export from Claude Design

Unsupported formats (Phase 2 roadmap):
- DOCX, PPTX, PDF, Canva link — return `DESIGN_IMPORT_UNSUPPORTED_FORMAT` and guide to path B.

### Version Whitelist

Before parsing, check the bundle's declared format version against `supported_bundle_versions` in `.moai/config/sections/design.yaml`.

Current default whitelist: `["1.0"]`

If the detected bundle version is not in the whitelist, return `DESIGN_IMPORT_UNSUPPORTED_VERSION` with the three required stderr fields (see Error Codes).

---

## Implementation Guide

### Bundle Parsing Flow

Step 1: Receive the bundle file path from the orchestrator.

Step 2: Validate file existence. If the path does not exist or is not readable, return `DESIGN_IMPORT_NOT_FOUND` immediately with manual path guidance.

Step 3: Validate file format. Inspect the file extension and magic bytes:
- `.zip`: ZIP magic bytes `PK\x03\x04`
- `.html`: HTML DOCTYPE or `<html` tag at start

If neither matches, return `DESIGN_IMPORT_UNSUPPORTED_FORMAT`.

Step 4: Security scan (before any extraction):
- List all ZIP entries (for ZIP bundles) without extracting
- Reject if any entry contains: executable extensions (`.sh`, `.exe`, `.bat`, `.cmd`, `.ps1`, `.py`, `.rb`, `.pl`), symbolic links, path traversal sequences (`../`, `..\`), or absolute paths
- If any security issue detected, return `DESIGN_IMPORT_SECURITY_REJECT` without extracting any content

Step 5: Read `manifest.json` from the bundle root:
- Extract `format_version` field
- Compare against `supported_bundle_versions` from `design.yaml`
- If version not in whitelist, return `DESIGN_IMPORT_UNSUPPORTED_VERSION`

Step 6: Extract and parse:
- `tokens.json` → `.moai/design/tokens.json`
- `components/*.html` or `components/*.json` → `.moai/design/components.json` (component manifest)
- `assets/**` → `.moai/design/assets/` (images, fonts, icons)
- `copy.json` (if present) → `.moai/design/copy.json`

Step 7: Validate extracted tokens structure:
- Required keys: `colors`, `typography`, `spacing`
- If any required key is missing, add a warning to the output but do not fail

Step 8: Report extraction results to the orchestrator.

---

### ZIP Bundle Expected Structure

```
bundle.zip
  manifest.json          # Required: format_version, claude_design_version, created_at
  tokens.json            # Required: colors, typography, spacing, radii, shadows
  components/            # Optional: component HTML or JSON specs
    hero.html
    navigation.html
    card.html
  assets/                # Optional: images, fonts, icons
    images/
    fonts/
    icons/
  copy.json              # Optional: structured copy output
```

### HTML Bundle Expected Structure

Single-file HTML export. Extract:
- Inline `<style>` CSS custom properties as color and spacing tokens
- Inline `<script>` JSON blocks tagged with `data-design-tokens`
- `<link>` tags referencing external assets (list only, do not fetch)

---

### Output Artifacts

All output is written to `.moai/design/`:

**`.moai/design/tokens.json`** — Normalized design tokens:
```
{
  "colors": { "primary": "...", "secondary": "...", ... },
  "typography": { "fontFamily": {...}, "fontSize": {...}, ... },
  "spacing": { "base": 4, "scale": {...} },
  "radii": { "sm": "4px", ... },
  "shadows": { "sm": "...", ... },
  "source": "claude-design-bundle",
  "bundle_version": "1.0"
}
```

**`.moai/design/components.json`** — Component manifest:
```
{
  "components": [
    { "name": "Hero", "file": "hero.html", "variants": [...] },
    ...
  ]
}
```

**`.moai/design/assets/`** — Static assets directory (images, fonts, icons extracted verbatim).

---

### Error Codes

All errors are returned as structured responses with an error code and human-readable message.

**`DESIGN_IMPORT_NOT_FOUND`**
- Trigger: Bundle file path does not exist or is not readable.
- Action: Return error immediately. Output manual guidance: "Provide the correct local file path, or switch to path B (moai-domain-brand-design)."

**`DESIGN_IMPORT_UNSUPPORTED_FORMAT`**
- Trigger: File is not ZIP, HTML, or magic bytes do not match.
- Action: Return error with supported format list. Guide to path B.

**`DESIGN_IMPORT_UNSUPPORTED_VERSION`**
- Trigger: Bundle `manifest.json` `format_version` is not in `supported_bundle_versions` whitelist.
- Required stderr output (all three fields mandatory):
  1. Detected version: `"Detected bundle version: v<N>"`
  2. Supported versions: `"Supported versions: <list from design.yaml>"`
  3. Fallback guidance: `"Switch to path B: run /moai design and select 'Code-based brand design'"`
- Do not create any partial output files.

**`DESIGN_IMPORT_SECURITY_REJECT`**
- Trigger: Bundle contains executable files, symbolic links, path traversal sequences, or absolute paths.
- Action: Reject without extracting any content. List the offending entries in the error message.
- Do not create `.moai/design/` directory.

**`DESIGN_IMPORT_MISSING_MANIFEST`**
- Trigger: ZIP bundle does not contain `manifest.json`.
- Action: Return error. Cannot determine bundle version without manifest. Guide to path B.

---

### Fallback Guidance

When any error code is returned, always append this guidance:

```
To continue with code-based design (path B):
1. Run /moai design
2. Select "Code-based brand design (moai-domain-brand-design)"
3. Ensure .moai/project/brand/visual-identity.md is complete
```

---

## Advanced Patterns

### Partial Bundle Recovery

If the bundle is valid but missing optional components (e.g., no `components/` directory):
- Extract what is available
- Add warnings to the output manifest
- Do not fail — proceed with partial output
- Note missing sections in `.moai/design/import-warnings.json`

### Token Normalization

Input bundles may use different naming conventions. Normalize to the MoAI token schema:

| Bundle field | Normalized token |
| --- | --- |
| `primary_color` | `colors.primary` |
| `brand_color` | `colors.primary` |
| `heading_font` | `typography.fontFamily.heading` |
| `base_spacing` | `spacing.base` |

Normalization rules are applied silently. Log all renamed fields in the import warnings.

### Asset Safety

When extracting assets:
- Validate image MIME types (accept: png, jpg, gif, webp, svg, ico)
- Validate font formats (accept: woff2, woff, ttf, otf)
- Reject archives within archives (no nested ZIPs)
- Strip metadata from SVG files that contains script tags

---

---

## Path B1 — Figma Extractor (via Meta-Harness)

**Prerequisite**: SPEC-V3R3-HARNESS-001 meta-harness skill (`moai-meta-harness`) must be
available. Path B1 does NOT ship a static Figma skill — it is generated dynamically.

### Invocation

When the user selects Path B1, invoke `moai-meta-harness` to generate:

```
.claude/skills/my-harness-figma-extractor/SKILL.md
```

The generated skill is project-scoped and user-owned (`my-harness-*` prefix). It will never
be overwritten by `moai update`.

### Generation Contract

`moai-meta-harness` Phase 5 (Customization) produces `my-harness-figma-extractor` with:

- **Figma file ID**: project-specific file identifier (collected during Socratic interview)
- **Page selectors**: which Figma pages map to which token categories
- **Credential reference**: environment variable name holding the Figma API token
  (e.g., `FIGMA_TOKEN`). The token value is NEVER stored in the skill file.

The generated extractor skill produces `tokens.json` + `components.json` at `.moai/design/`
matching the reserved file paths in design constitution §3.2.

### Output

After `my-harness-figma-extractor` runs, this skill receives the produced artifacts at
`.moai/design/` and proceeds to DTCG validation before `expert-frontend` consumption.

---

## Path B2 — Pencil MCP (via Meta-Harness)

**Prerequisite**: SPEC-V3R3-HARNESS-001 meta-harness skill (`moai-meta-harness`) must be
available. Path B2 is implemented via dynamically generated `my-harness-pencil-mcp` skill
(SPEC-V3R3-HARNESS-001 BC-V3R3-007 16-skill removal — legacy static skill archived).

### Invocation

When the user selects Path B2, invoke `moai-meta-harness` to generate:

```
.claude/skills/my-harness-pencil-mcp/SKILL.md
```

### Generation Contract

`moai-meta-harness` Phase 5 produces `my-harness-pencil-mcp` with:

- **`.pen` file locations**: glob patterns for the project's Pencil design files
- **Pencil MCP server endpoint**: local server URL (default: `http://localhost:11111`)
- **Export configuration**: which Pencil frames/pages map to token categories

The generated skill uses `mcp__pencil__*` tools (already listed in `moai-design-system`
allowed-tools) to extract design data and write to `.moai/design/`.

### Output

After `my-harness-pencil-mcp` runs, this skill receives artifacts at `.moai/design/` and
proceeds to DTCG validation before `expert-frontend` consumption.

---

## Works Well With

- `moai-domain-brand-design`: Fallback path when bundle import fails
- `moai-workflow-gan-loop`: Receives extracted tokens for quality evaluation
- `expert-frontend`: Primary consumer of extracted design artifacts
- `moai-meta-harness`: Generates `my-harness-figma-extractor` and `my-harness-pencil-mcp`
  for Path B1/B2 (SPEC-V3R3-HARNESS-001)
- `moai-design-system`: DTCG token validation guidance and design system context

---

REQ coverage: REQ-DPL-001, REQ-DPL-002, REQ-DPL-003, REQ-SKILL-007, REQ-SKILL-008, REQ-SKILL-009, REQ-SKILL-010, REQ-SKILL-015
Version: 1.1.0 (SPEC-V3R3-DESIGN-PIPELINE-001 Wave C.1)
