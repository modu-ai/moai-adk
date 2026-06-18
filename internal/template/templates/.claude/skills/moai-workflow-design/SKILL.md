---
name: moai-workflow-design
description: >
  Unified design workflow skill — handles Path A (Claude Design handoff bundle import,
  via Figma extractor when needed) and design-brief context loading from .moai/design/
  (research, system, spec). Validates DTCG tokens, enforces brand-context constitutional
  priority. Use for /moai design workflow — NOT for general design system documentation.

when_to_use: >
  Use for the /moai design workflow: Path A Claude Design handoff-bundle
  import (via Figma extractor when needed), design-brief context loading
  from .moai/design/, DTCG token validation, and brand-context
  constitutional priority.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "0.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-22"
  tags: "design, import, handoff-bundle, claude-design, design-tokens, components, figma, meta-harness, brief, attach, context"
  related-skills: "moai-domain-brand-design, moai-workflow-gan-loop, moai-meta-harness, moai-domain-design-handoff"

progressive_disclosure:
  enabled: true
  level1_tokens: 120
  level2_tokens: 5000
---

# Design Workflow (`moai-workflow-design`)

Unified `/moai design` workflow skill. Handles two complementary responsibilities:

1. **Design artifact import** — Path A (Claude Design handoff bundle, ZIP/HTML) and Path
   B1 (Figma extractor via meta-harness). Produces DTCG-validated design tokens at
   `.moai/design/tokens.json` for `expert-frontend` consumption.
2. **Design-brief context loading** — Auto-loads human-authored briefs from `.moai/design/`
   (`spec.md`, `system.md`, `research.md`) into the orchestrator prompt before
   `expert-frontend` or `moai-domain-brand-design` runs.

Brand context (`.moai/project/brand/`) is the constitutional parent across all paths — no
path may override brand constraints (design constitution §3.1, §3.3).

## Quick Reference

**Reserved output paths** (design constitution §3.2, must not collide with human files):
`tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`,
`copy.json`, `path-selection.json` — all under `.moai/design/`.

**Path selection** (presented via AskUserQuestion when `/moai design` needs choice):
1. **Path A — Claude Design** (권장) — handoff bundle (ZIP or HTML)
2. **Path B1 — Figma** — meta-harness generates `moai-harness-figma-extractor` dynamically

Selection persisted to `.moai/design/path-selection.json`.

**Context-loading priority order** (REQ-2 / AC-4 from absorbed design-context skill):
`spec > system > research`. When token budget exceeded, drop in REVERSE priority — never
drop `spec`. Default `token_budget: 20000` from `design.yaml design_docs.token_budget`.

**Token estimation**: `estimated_tokens = ceiling(char_count / 4) * 1.10`.

## Implementation Guide

### Part 1 — Path A: Claude Design Handoff Bundle

**Supported formats (Phase 1)**:
- `ZIP` — Claude Design export with `manifest.json`, `tokens.json`, `components/`, `assets/`
- `HTML` — single-file Claude Design export

**Unsupported (Phase 2 roadmap)**: DOCX, PPTX, PDF, Canva link — return
`DESIGN_IMPORT_UNSUPPORTED_FORMAT` and guide to Path B.

**Version whitelist**: Check `manifest.json` `format_version` against
`supported_bundle_versions` in `.moai/config/sections/design.yaml`. Current default: `["1.0"]`.
Mismatch → `DESIGN_IMPORT_UNSUPPORTED_VERSION`.

**Parsing flow**:
1. Receive bundle file path from orchestrator
2. Validate file existence → `DESIGN_IMPORT_NOT_FOUND` if missing
3. Validate format (extension + magic bytes: `PK\x03\x04` for ZIP, DOCTYPE/`<html` for HTML)
4. **Security scan before extraction** — list ZIP entries; reject executables (`.sh`, `.exe`,
   `.bat`, `.cmd`, `.ps1`, `.py`, `.rb`, `.pl`), symlinks, path traversal (`../`, `..\`),
   absolute paths → `DESIGN_IMPORT_SECURITY_REJECT`
5. Read `manifest.json`, validate version
6. Extract: `tokens.json` → `.moai/design/tokens.json`; `components/` → `components.json`;
   `assets/**` → `.moai/design/assets/`; `copy.json` → `.moai/design/copy.json`
7. Validate token structure (required keys: `colors`, `typography`, `spacing`); missing
   keys → warning, not failure
8. Report extraction results

**Expected ZIP structure**: `manifest.json` (format_version, claude_design_version,
created_at) + `tokens.json` (colors, typography, spacing, radii, shadows) + optional
`components/` (HTML or JSON specs) + optional `assets/` (images, fonts, icons) + optional
`copy.json` (structured copy).

**Output token schema** (normalized to MoAI): top-level keys `colors`, `typography`,
`spacing`, `radii`, `shadows`, plus `source: "claude-design-bundle"` and `bundle_version`.

**Field normalization** (silent rename, logged in import-warnings.json):
`primary_color`/`brand_color` → `colors.primary`; `heading_font` →
`typography.fontFamily.heading`; `base_spacing` → `spacing.base`.

**Asset safety**: Validate image MIME (png, jpg, gif, webp, svg, ico) and font formats
(woff2, woff, ttf, otf). Reject nested ZIPs. Strip script tags from SVG metadata.

### Part 2 — Path B1: Figma Extractor (Meta-Harness)

**Prerequisite**: SPEC-V3R3-HARNESS-001 `moai-meta-harness`. Path B1 does NOT ship a
static Figma skill — it is generated dynamically. When user selects Path B1, invoke
`moai-meta-harness` to generate `.claude/skills/harness-figma-extractor/SKILL.md`
(project-scoped and user-owned via `harness-*` prefix — `moai update` never
overwrites). Meta-harness Phase 5 (Customization) collects via Socratic interview:
Figma file ID, page selectors mapping pages to token categories, credential reference
(env var name like `FIGMA_TOKEN`; value NEVER stored in skill file). Generated extractor
produces `tokens.json` + `components.json` at `.moai/design/`; DTCG validation runs before
`expert-frontend` consumption.

### Part 3 — Design-Brief Context Loading

Auto-loads human-authored briefs during Phase B2.5 of `/moai design` when
`design_docs.auto_load_on_design_command: true`. Can also be invoked standalone with
explicit `dir` argument.

**Configuration resolution**: Read `design_docs` from `.moai/config/sections/design.yaml`.
If absent, use compiled-in defaults:
- `dir: .moai/design`
- `auto_load_on_design_command: true`
- `token_budget: 20000`
- `priority: [spec, system, research]`

Log `design_docs not configured — using defaults` when key absent.

**Bare-token → filename mapping**:
- `spec` → `<dir>/spec.md`
- `system` → `<dir>/system.md`
- `research` → `<dir>/research.md`

**Steps**:
1. **Directory check**: Glob `<dir>/`. Missing → emit header only and log
   `design docs not initialized — run /moai init or SPEC-DESIGN-DOCS-001 to create`.
2. **Auto-load gate**: From Phase B2.5, check `auto_load_on_design_command`. False → skip.
3. **Parallel Read**: Issue all candidate file Reads in a single batched parallel tool-call set.
4. **Filter `_TBD_` files**: A file with only scaffold content (lines blank, `_TBD_`,
   headings without bodies, or `<!--`/`>` comments) is skipped. Log
   `skip: <token> — _TBD_ only`.
5. **Token budget enforcement**: Include in priority order until cumulative
   `estimated_tokens` would exceed budget. Overflow → drop lowest priority (`research`
   first, then `system`; never `spec`). Single file too large → truncate at nearest
   `##`/`###` boundary and append `> truncated: <filename> at char_offset=N`.
6. **Build output block** — first non-empty line MUST be exactly `## Design Context (from
   .moai/design/)`. For each file, prepend `> source: .moai/design/<filename>` then
   content (or truncated).
7. **Warnings section** (when unreadable files encountered): append
   `> warnings: [<token1> unreadable: <reason>, ...]` after the content.

**All-`_TBD_` case**: header-only output + log
`design docs present but all are _TBD_ — no content loaded`.

### Error Codes (Path A)

- `DESIGN_IMPORT_NOT_FOUND` — bundle path missing → guide to Path B
- `DESIGN_IMPORT_UNSUPPORTED_FORMAT` — non-ZIP/HTML → guide to Path B
- `DESIGN_IMPORT_UNSUPPORTED_VERSION` — version not in whitelist. Required stderr (all 3
  lines mandatory): `Detected bundle version: v<N>`; `Supported versions: <list from
  design.yaml>`; `Switch to path B: run /moai design and select 'Code-based brand design'`.
- `DESIGN_IMPORT_SECURITY_REJECT` — executables/symlinks/traversal/absolute paths
  detected. List offending entries. Do NOT create `.moai/design/` directory.
- `DESIGN_IMPORT_MISSING_MANIFEST` — ZIP without `manifest.json` → guide to Path B

**Fallback guidance** appended to every error: instruct user to run `/moai design` and
select "Code-based brand design (moai-domain-brand-design)" after ensuring
`.moai/project/brand/visual-identity.md` is complete.

### Partial Bundle Recovery

Valid bundle missing optional components → extract what's available, log warnings to
`.moai/design/import-warnings.json`, proceed with partial output. Never silent failure.

## Works Well With

`moai-domain-brand-design` (Path B fallback / context consumer), `moai-domain-design-handoff`
(produces `claude-design-handoff/` for Path A), `moai-workflow-gan-loop` (uses tokens +
context as baseline), `moai-meta-harness` (generates figma extractor for Path B1),
`expert-frontend` (primary consumer), `.claude/rules/moai/design/constitution.md` (brand
priority + reserved paths).

## Common Rationalizations

- "Skip security scan for trusted bundles" — "trusted" is unverifiable. Scan every bundle, no exceptions.
- "Drop spec.md when budget tight" — spec.md is priority 1, never dropped. Drop research → system → escalate.
- "_TBD_ files contain useful context" — `_TBD_` means scaffold-only. Skip to avoid polluting the prompt.
- "Path B1 needs a hardcoded Figma extractor" — Path B1 uses meta-harness generation. Static Figma skill prohibited.
- "Brand context is one input among many" — brand context is the constitutional parent; conflicts resolve in favor of brand.

## Red Flags

- Bundle parse proceeds without security scan
- ZIP entries containing `../`, symlinks, or executables accepted
- `manifest.json` version validation bypassed
- Design context block missing canonical header `## Design Context (from .moai/design/)`
- `spec.md` dropped when budget exceeded (priority violation)
- Figma API token value stored inside skill file (only env var name allowed)
- Output written outside `.moai/design/` reserved path set

## Verification

- [ ] Path A security scan rejects fixture with `..` and symlinks
- [ ] Path A produces `.moai/design/tokens.json` with normalized schema
- [ ] Path B1 invocation triggers `moai-meta-harness` (no static skill)
- [ ] Context-load output starts with `## Design Context (from .moai/design/)`
- [ ] Budget truncation appends `> truncated: <filename> at char_offset=N`
- [ ] All-`_TBD_` case emits header + log only
- [ ] DTCG validation runs on Path A and Path B1 outputs
- [ ] `internal/design/dtcg/frozen_guard_test.go` references this skill name

REQ coverage: REQ-DPL-001..003, REQ-SKILL-007..015 (Path A); REQ-1..16 (context).

<!-- absorbed from moai-workflow-design-import + moai-workflow-design-context per the skill consolidation policy -->
