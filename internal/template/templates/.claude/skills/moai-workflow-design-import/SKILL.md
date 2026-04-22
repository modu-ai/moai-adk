---
name: moai-workflow-design-import
description: >
  Parses Claude Design handoff bundle (ZIP or HTML) and extracts design tokens,
  component manifests, and static assets for expert-frontend delegation. Validates
  bundle version against the supported_bundle_versions whitelist and returns
  structured error codes on failure with path B fallback guidance.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-20"
  tags: "design import, handoff bundle, claude design, design tokens, components"
  related-skills: "moai-domain-brand-design, moai-workflow-gan-loop"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["design import", "handoff bundle", "claude design", "bundle path", "design zip", "import bundle"]
  agents: ["expert-frontend"]
  phases: ["run"]
---

# moai-workflow-design-import

Handles Claude Design handoff bundle ingestion for the `/moai design` path A workflow. Validates bundle format, extracts design artifacts, and prepares structured output for `expert-frontend` delegation.

This skill is invoked only when the user selects path A (Claude Design) in `/moai design` and provides a local bundle file path.

---

## Quick Reference

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

## Works Well With

- `moai-domain-brand-design`: Fallback path when bundle import fails
- `moai-workflow-gan-loop`: Receives extracted tokens for quality evaluation
- `expert-frontend`: Primary consumer of extracted design artifacts

---

REQ coverage: REQ-SKILL-007, REQ-SKILL-008, REQ-SKILL-009, REQ-SKILL-010, REQ-SKILL-015
Version: 1.0.0
