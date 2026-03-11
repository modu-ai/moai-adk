# SPEC-UPDATE-002 Implementation Plan

## Phase 1: Foundation Fixes (Critical Errors)

### 1.1 reference/figma.md — REWRITE
- Remove all fictional API calls (figma.get_file_metadata, figma.get_components, etc.)
- Replace with verified Figma MCP tools from developers.figma.com
- Add installation: `claude plugin install figma@claude-plugins-official`
- Add Remote MCP server: https://mcp.figma.com/mcp
- Document Code-to-Canvas (generate_figma_design) with v1 limitations
- Add implement-design 7-step workflow
- Add get_variable_defs for design token extraction
- Estimated: 318→350 lines

### 1.2 reference/pencil-renderer.md — UPDATE
- Fix ".pen files are encrypted" → ".pen files are pure JSON (Git diffable/mergeable)"
- Remove manual mcpServers configuration section
- Add Pencil MCP auto-setup note
- Add UI Kit section (Shadcn UI, Halo, Lunaris, Nitro)
- Verify tool names via Context7 before adding new tools
- Estimated: 367→380 lines

### 1.3 reference/pencil-code.md — REWRITE
- Remove fictional pencil.export_to_react API
- Remove fictional pencil.config.js configuration
- Replace with actual prompt-based workflow: batch_get → analyze structure → generate React/Tailwind
- Keep design-to-code patterns and Tailwind mapping guidance
- Estimated: 546→400 lines

## Phase 2: Design-Tools Shell Updates

### 2.1 moai-design-tools/SKILL.md — UPDATE
- Add verified Pencil tools to allowed-tools if confirmed via Context7
- Update Pencil MCP Tools table to match verified tool list
- Add note about .pen pure JSON format
- Add UI Kit section reference
- Bump version to 4.0.0
- Update metadata.updated to 2026-03-11
- Estimated: 345→360 lines

### 2.2 reference/comparison.md — UPDATE
- Align Figma column with new official MCP tools
- Update Pencil column with corrected .pen format info
- Adjust workflow recommendations based on updated capabilities
- Estimated: 425→430 lines

## Phase 3: Domain-UiUx Core Updates

### 3.1 modules/web-interface-guidelines.md — UPDATE (Major)
- ADD new section: "Design Direction & Anti-AI Slop" (~120 lines)
  - Design thinking process: Purpose → Tone → Constraints → Differentiation
  - Banned patterns (fonts, colors, layouts)
  - Style extremes guide
  - Atmospheric backgrounds
- ADD new section: "Motion & Microinteraction Design" (~40 lines)
  - Transition timing and easing curves
  - Entrance/exit animation patterns
  - Stagger patterns
  - Scroll-triggered effects
- EXPAND existing Touch section: "Mobile-First UX" (~30 lines)
  - Touch target sizing (min 44x44px)
  - Gesture design patterns
  - Bottom sheet, pull-to-refresh patterns
- Estimated: 688→810 lines

### 3.2 modules/design-system-tokens.md — UPDATE
- REMOVE Pencil MCP Integration Workflow section (lines 194-306, ~115 lines)
- ADD 3-line cross-reference to moai-design-tools
- Estimated: 441→330 lines

### 3.3 moai-domain-uiux/SKILL.md — UPDATE
- TypeScript 5.5 → 5.9+
- Tailwind CSS 3.4 → 4.x
- Add "Anti-AI Slop" and "design direction" to triggers keywords
- Add Nova preset cross-reference in Quick Reference
- Bump version to 3.0.0
- Update metadata.updated
- Estimated: 237→250 lines

## Phase 4: Minor Updates

### 4.1 modules/theming-system.md — UPDATE
- Add Nova preset cross-reference (link to moai-design-tools)
- Update Tailwind CSS references to v4 where applicable
- Estimated: 374→385 lines

### 4.2 modules/icon-libraries.md — UPDATE
- Add Hugeicons to icon library table
- Note as default for Nova/shadcn preset
- Estimated: 402→415 lines

### 4.3 references/reference.md — UPDATE
- Add Figma MCP official docs link
- Add Anthropic Frontend Design skill reference
- Estimated: 244→250 lines

### 4.4 references/examples.md — UPDATE
- Update version references only
- Estimated: 560→560 lines

## Dependencies

- Phase 2 depends on Phase 1 (figma/pencil corrections inform SKILL.md)
- Phase 3.2 and 3.3 can run in parallel with Phase 3.1
- Phase 4 can run in parallel with Phase 3
- Context7 verification should run before Phase 2 (tool name confirmation)

## Post-Implementation

1. `make build` — regenerate internal/template/embedded.go
2. `go test ./internal/template/...` — verify template system
3. Grep verification: no remnants of fictional APIs
4. Copy to local .claude/skills/ for immediate testing
