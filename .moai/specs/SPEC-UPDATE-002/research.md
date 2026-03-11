# SPEC-UPDATE-002 Research

## Overview

moai-domain-uiux와 moai-design-tools 스킬의 전체 업데이트를 위한 리서치 결과.

## 1. 현재 상태 분석

### moai-domain-uiux (~3,436 lines, 9 files)

| File | Lines | Version | Updated |
|------|-------|---------|---------|
| SKILL.md | 237 | 2.0.0 | 2026-01-11 |
| modules/accessibility-wcag.md | 261 | - | 2025-11-26 |
| modules/icon-libraries.md | 402 | 4.0.0 | 2025-11-13 |
| modules/component-architecture.md | 229 | - | 2025-11-26 |
| modules/theming-system.md | 374 | - | 2025-11-26 |
| modules/design-system-tokens.md | 441 | 1.0.0 | 2025-11-21 |
| modules/web-interface-guidelines.md | 688 | 1.0.0 | 2026-01-15 |
| references/examples.md | 560 | - | 2025-11-26 |
| references/reference.md | 244 | - | 2025-11-26 |

### moai-design-tools (~2,001 lines, 5 files)

| File | Lines | Version | Updated |
|------|-------|---------|---------|
| SKILL.md | 345 | 3.0.0 | 2026-02-21 |
| reference/pencil-renderer.md | 367 | - | 2026-02-21 |
| reference/figma.md | 318 | 1.0.0 | 2026-02-09 |
| reference/pencil-code.md | 546 | 1.0.0 | 2026-02-09 |
| reference/comparison.md | 425 | 1.0.0 | 2026-02-09 |

## 2. Gap Analysis

### 2.1 moai-domain-uiux MISSING

1. **Anti-AI Slop Prevention**: No guidance on avoiding generic AI-generated UI patterns
2. **Design Direction Framework**: No Purpose→Tone→Constraints→Differentiation process
3. **Motion/Microinteraction Design**: Only compliance-focused (prefers-reduced-motion), no design intent
4. **Mobile-First UX Patterns**: Only 12-line Touch section, missing touch targets, gestures, bottom sheets
5. **Nova Preset**: Not referenced despite moai-design-tools defaulting to it

### 2.2 moai-domain-uiux OUTDATED

1. TypeScript 5.5 → should be 5.9+
2. Tailwind CSS 3.4 → Tailwind CSS 4 released
3. Missing Hugeicons in icon-libraries.md (Nova preset default)

### 2.3 moai-domain-uiux DUPLICATED

1. design-system-tokens.md lines 194-306: Pencil MCP workflow duplicates moai-design-tools

### 2.4 moai-design-tools CRITICAL ERRORS

1. **figma.md**: Uses fictional API (figma.get_file_metadata, figma.get_components) — NOT real MCP tools
2. **pencil-renderer.md**: States ".pen files are encrypted" — INCORRECT, .pen is pure JSON
3. **pencil-code.md**: References fictional pencil.export_to_react API and pencil.config.js

### 2.5 moai-design-tools MISSING

1. Figma official remote MCP server (mcp.figma.com)
2. Code-to-Canvas (generate_figma_design)
3. Figma implement-design 7-step workflow
4. UI Kit documentation (Shadcn UI, Halo, Lunaris, Nitro)

## 3. External Research Findings

### 3.1 Pencil MCP (Verified)

- .pen files: Pure JSON, Git diffable/mergeable
- MCP auto-setup: No manual mcpServers config needed
- UI Kits: Shadcn UI, Halo, Lunaris, Nitro (all confirmed)
- search_all_unique_properties / replace_all_matching_properties: **UNVERIFIED** — omit from update

Confirmed tools: batch_design, batch_get, get_screenshot, snapshot_layout, get_editor_state, get_variables, set_variables, get_canvas_context, get_selected_frames, get_style_guide, export_frame_data

Unverified tools (need Context7 check): get_canvas_context, get_selected_frames, export_frame_data

### 3.2 Figma MCP (Verified from developers.figma.com)

- Remote server: https://mcp.figma.com/mcp
- Install: `claude plugin install figma@claude-plugins-official`
- Verified tools: generate_figma_design, get_design_context, get_screenshot, get_variable_defs, get_metadata, get_code_connect_map, add_code_connect_map, get_figjam, generate_diagram, create_design_system_rules, whoami
- Code-to-Canvas: v1 (Feb 2026), known issues with Japanese text and image dimensions
- implement-design skill: 7-step workflow from Figma official

### 3.3 Anti-AI Slop (Anthropic frontend-design, 140.1K installs)

Design thinking process: Purpose → Tone → Constraints → Differentiation

Banned patterns:
- Fonts: Inter, Roboto, Arial, system fonts, Space Grotesk
- Colors: Purple gradients on white backgrounds
- Patterns: Predictable layouts, cookie-cutter components

Mandated alternatives:
- Distinctive font pairings
- Dominant colors + sharp accents
- Atmospheric backgrounds (gradient meshes, noise textures, grain overlays)
- High-impact motion (page loads + scroll triggers)
- Asymmetric spatial composition

Style extremes: brutally minimal, maximalist chaos, retro-futuristic, organic/natural, luxury/refined, playful/toy-like, editorial/magazine, brutalist/raw, art deco/geometric, soft/pastel, industrial/utilitarian

## 4. Technical Design

### 4.1 File Impact Analysis (14 files total)

**moai-design-tools (5 files):**

| File | Action | Lines | Change |
|------|--------|-------|--------|
| reference/figma.md | REWRITE | 318→350 | Replace fictional API with verified Figma MCP |
| reference/pencil-renderer.md | UPDATE | 367→380 | Fix .pen format, remove manual MCP config |
| reference/pencil-code.md | REWRITE | 546→400 | Remove fictional export API |
| SKILL.md | UPDATE | 345→360 | Add tools to allowed-tools, bump v4.0.0 |
| reference/comparison.md | UPDATE | 425→430 | Align with figma/pencil updates |

**moai-domain-uiux (9 files):**

| File | Action | Lines | Change |
|------|--------|-------|--------|
| SKILL.md | UPDATE | 237→250 | TS 5.9+, TW 4, Anti-AI Slop trigger, v3.0.0 |
| modules/design-system-tokens.md | UPDATE | 441→330 | Remove 115-line Pencil duplicate |
| modules/web-interface-guidelines.md | UPDATE | 688→810 | Add Anti-AI Slop, motion, mobile-first |
| modules/theming-system.md | UPDATE | 374→385 | Nova cross-reference, Tailwind v4 |
| modules/icon-libraries.md | UPDATE | 402→415 | Add Hugeicons |
| references/examples.md | UPDATE | 560→560 | Minor version updates |
| references/reference.md | UPDATE | 244→250 | Add new resource links |
| modules/component-architecture.md | NO CHANGE | 229 | - |
| modules/accessibility-wcag.md | NO CHANGE | 261 | - |

### 4.2 Implementation Order

1. **Phase 1 (Foundation Fixes)**: figma.md, pencil-renderer.md, pencil-code.md — fix critical errors first
2. **Phase 2 (Design-Tools Shell)**: SKILL.md, comparison.md — update frontmatter and tool list
3. **Phase 3 (Domain-UiUx Core)**: web-interface-guidelines.md (Anti-AI Slop), design-system-tokens.md (dedup), SKILL.md
4. **Phase 4 (Minor Updates)**: theming-system.md, icon-libraries.md, examples.md, reference.md

### 4.3 Token Budget Impact

- moai-design-tools: -166 lines net (-950 tokens) — pencil-code.md shrinks significantly
- moai-domain-uiux: +11 lines net (+600 tokens) — Anti-AI Slop adds, Pencil dedup removes
- Total: Net +600 tokens (~3.75% increase), within L2 budget

### 4.4 Risks

1. **HIGH**: 3 Pencil tools (get_canvas_context, get_selected_frames, export_frame_data) need Context7 verification
2. **MEDIUM**: Figma tool list should be verified via Context7 during implementation
3. **MEDIUM**: Code-to-Canvas v1 has known issues — document limitations clearly
4. **LOW**: generate_figma_design may be unavailable in some sessions (rollout)

### 4.5 Post-Implementation Verification

1. `make build` — regenerate embedded templates
2. `go test ./internal/template/...` — verify template system
3. Grep for fictional API remnants (figma.get_file_metadata, pencil.export_to_react, pencil.config.js)
4. Context7 verification of Pencil and Figma tool names
