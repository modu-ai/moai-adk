---
title: Claude Design Handoff
description: Official Claude Design features and handoff bundle usage
weight: 30
draft: false
---

# Claude Design Handoff

## Claude Design Overview

**Claude Design** is an **AI-powered design generation tool** launched by Anthropic on April 17, 2026. Generate UI/UX designs in natural language via dedicated interface on Claude.ai.

- **Base Model:** Claude 3.5 Opus
- **Access:** https://claude.ai/design
- **Output Formats:** Design tokens, component specs, static assets, handoff bundle

## Supported Subscription Plans

| Plan | Claude Design Support | Notes |
|---|---|---|
| Free | Not supported | Cannot access Claude.ai/design |
| Pro | Supported | $20/month |
| Max | Supported | $200/month |
| Team | Supported (admin: off by default) | Billed per team |
| Enterprise | Supported (admin: off by default) | Contract-based |

**Note:** Team and Enterprise plans **disable feature by default** at admin level. Request admin to enable.

## Supported Input Formats

Claude Design accepts multiple input types:

| Format | Description |
|---|---|
| **Text** | Natural language design requirements |
| **Images** | Reference design mockups |
| **DOCX/PPTX** | Existing documents or presentations |
| **XLSX** | Data tables and structured info |
| **Web Capture** | Website screenshots from URL |
| **Figma** | Figma files and frames |
| **GitHub** | Repository code and README |

## Export Handoff Bundle

### Step 1: Open Claude.ai/design

Open browser to **https://claude.ai/design**

### Step 2: Create Design

In Claude Design interface:
- Enter design description in natural language
- Upload reference images/documents
- Real-time UI preview generation

Example prompt:
```
Create a landing page for a tech company.
- Hero section: large header, value proposition, CTA button
- Features section: 3 cards explaining key features
- Colors: dark blue (#1E40AF), light blue (#3B82F6)
- Typography: modern, clean aesthetic
```

### Step 3: Export Bundle

From Claude Design **Export** or **Share** menu:
- **ZIP format:** All design files, tokens, assets
- **PDF format:** Static document version (optional)
- **Canva/Figma format:** Continue editing in external tools (optional)
- **HTML/Claude Code:** Code snippets included

**Recommended:** Export as ZIP format

### Step 4: Save Locally

Save exported file to local filesystem:

```bash
# Example: ~/Downloads/my-design.zip
```

## Import Bundle into MoAI-ADK

### Step 5: Re-run /moai design

In Claude Code:

```
/moai design
```

Select Path A (Claude Design), then:

```
Enter bundle path: ~/Downloads/my-design.zip
```

### Step 6: Auto Conversion

`moai-workflow-design-import` skill:
- Parses bundle
- Converts design tokens to JSON
- Extracts component specs
- Copies static assets

Result files:
```
.moai/design/
├── tokens.json          # Design tokens (colors, typography, spacing)
├── components.json      # Component specifications
└── assets/              # Images, icons
```

## Bundle Version Support

Supported bundle formats:

| Bundle Version | Claude Design Release | Status | Notes |
|---|---|---|---|
| v1.0 (initial) | 2026-04-17 | Supported | Standard ZIP format |
| v1.1 | 2026-05-xx | Planned | Extended metadata |
| v2.0 (preview) | Future | Not supported | Manual compatibility update needed |

**Whitelist:** See `supported_bundle_versions` in `.moai/config/sections/design.yaml`

## Error Codes

When bundle import fails:

| Error Code | Cause | Solution |
|---|---|---|
| `DESIGN_IMPORT_NOT_FOUND` | Bundle file not found | Check path, verify file exists |
| `UNSUPPORTED_FORMAT` | Format other than ZIP | Re-export as ZIP format |
| `UNSUPPORTED_VERSION` | Bundle version not supported | Re-export from latest Claude Design |
| `SECURITY_REJECT` | Security check failed (malicious script detected) | Contact admin |
| `MISSING_MANIFEST` | Bundle structure corrupted | Create new bundle and retry |

## Fallback Path

When Claude Design unavailable or bundle import fails:

### Option 1: Switch to Path B

```
Bundle import failed. Switch to Path B?
→ [Yes] [No]
```

Select to auto-switch to code-based design workflow

### Option 2: Regenerate Bundle

Return to Claude.ai/design:
1. Modify existing design or create new
2. Re-export ZIP from Export menu
3. Re-run `/moai design` with new file

## Team Collaboration

Claude Design Team subscription features:

- **Real-time collaboration:** Multiple team members edit design simultaneously
- **Share Link:** Share design outside team (read-only)
- **Version History:** Restore previous versions
- **Comments:** Record feedback on design

**Note:** Disabled by default; team admin must enable

## Next Steps

- After bundle import, see [GAN Loop](./gan-loop.md) guide
- During code implementation, check [Sprint Contract Protocol](./gan-loop.md#sprint-contract-protocol)
- Review scoring criteria at [4-Dimensional Scoring](./gan-loop.md#4-dimensional-scoring)
