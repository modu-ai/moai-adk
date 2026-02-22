---
name: moai-design-tools
description: >
  Design tool integration specialist covering Figma MCP, Pencil renderer, and Pencil-to-code export.
  Use when fetching design context from Figma, rendering Pencil DNA codes to .pen frames, exporting .pen
  designs to React/Tailwind code, or choosing design-to-code workflows. Supports design fetching (Figma),
  visual rendering (Pencil MCP), and code generation (React/Tailwind).
license: MIT
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob Bash WebFetch WebSearch mcp__context7__resolve-library-id mcp__context7__get-library-docs mcp__pencil__batch_design mcp__pencil__batch_get mcp__pencil__get_screenshot mcp__pencil__snapshot_layout mcp__pencil__get_editor_state mcp__pencil__get_variables mcp__pencil__set_variables mcp__pencil__get_guidelines mcp__pencil__get_style_guide mcp__pencil__get_style_guide_tags mcp__pencil__open_document mcp__pencil__find_empty_space_on_canvas
user-invocable: false
metadata:
  version: "3.0.0"
  category: "domain"
  status: "active"
  updated: "2026-02-21"
  modularized: "false"
  tools: "Figma, Pencil MCP"
  tags: "figma, pencil, design to code, design export, render dna, pen frame, react from design, tailwind from design, design context, ui implementation"
  context7-libraries: "/figma/docs, /pencil/docs"
  related-skills: "moai-domain-uiux, moai-domain-frontend, moai-library-shadcn, moai-lang-typescript, moai-lang-react"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5500

# MoAI Extension: Triggers
triggers:
  keywords: ["figma", "pencil", "design to code", "design export", "render dna", "pen frame", "react from design", "tailwind from design", "design context", "ui implementation", "design fetching", "figma mcp", "pencil mcp", "component from design", "layout from design"]
  agents: ["expert-frontend", "team-designer"]
  phases: ["run"]
---

# Design Tools Integration Specialist

Comprehensive design-to-code workflow guidance covering three major capabilities: Figma MCP (design fetching), Pencil MCP (visual rendering), and Pencil-to-code export (React/Tailwind generation).

## Default Design Style (shadcn/ui Nova)

When no specific design style is requested, use the **shadcn/ui Nova** preset with Notion-style neutral color scheme:

```
bunx --bun shadcn@latest create --preset "https://ui.shadcn.com/init?base=radix&style=nova&baseColor=neutral&theme=neutral&iconLibrary=hugeicons&font=noto-sans&menuAccent=bold&menuColor=default&radius=small&template=next&rtl=false" --template next
```

### Nova Style Configuration

| Property | Value | Description |
|----------|-------|-------------|
| Style | `nova` | Modern, clean design language |
| Base Color | `neutral` | Notion-style grayscale palette |
| Theme | `neutral` | Consistent neutral theming |
| Icon Library | `hugeicons` | Comprehensive icon set |
| Font | `noto-sans` | Clean, readable sans-serif |
| Radius | `small` | Subtle rounded corners |
| Menu Accent | `bold` | Strong menu highlighting |

### When to Use Default Style

Apply the Nova preset when:
- User requests "clean", "modern", or "minimalist" design without specifics
- No brand guidelines or design system specified
- Creating dashboards, admin panels, or productivity tools
- Building documentation or content-focused interfaces

## Quick Tool Selection

### Figma MCP - Design Context Fetching

Figma integration for fetching design context, metadata, and screenshots from Figma files.

Best For: Fetching design tokens, component specifications, layout information, and style guides from existing Figma files. Extracting design system values and understanding design structure.

Key Strengths: Direct access to Figma file metadata, component hierarchy extraction, style guide generation, design token retrieval, screenshot capture for visual reference.

Workflow: Connect to Figma file → Fetch file metadata → Extract component tree → Retrieve design tokens → Generate style documentation.

Context7 Library: /figma/docs

### Pencil MCP - Visual Design Rendering

Pencil MCP integration for creating and editing .pen design files with AI-assisted design generation.

Best For: Rapid prototyping, visual design iterations, creating UI mockups from text descriptions, collaborative design discussions, visual proposals before implementation.

Key Strengths: Text-to-design conversion, batch design operations, style guide integration, visual preview without implementation, collaborative design workflow.

**Available Pencil MCP Tools:**

| Tool | Purpose |
|------|---------|
| `batch_design` | Create, modify, and manipulate design elements in batches |
| `batch_get` | Read design components and hierarchy by patterns or node IDs |
| `get_screenshot` | Render design previews as images |
| `snapshot_layout` | Analyze computed layout structure |
| `get_editor_state` | Get current editor context and active file |
| `get_variables` | Read design tokens and theme variables |
| `set_variables` | Update design tokens and theme variables |
| `get_guidelines` | Get design guidelines for code, tables, Tailwind, or landing pages |
| `get_style_guide` | Get style guide by name or tags |
| `get_style_guide_tags` | List all available style guide tags |
| `open_document` | Open existing .pen file or create new one |
| `find_empty_space_on_canvas` | Find available space for new elements |

Workflow: Describe UI in natural language → Generate design with batch_design → Visually review with get_screenshot → Iterate on design → Export to code when ready.

Context7 Library: /pencil/docs

### Pencil-to-Code Export - React/Tailwind Generation

Export .pen designs to production-ready React and Tailwind CSS code.

Best For: Converting approved .pen designs to implementation, generating React components with Tailwind styling, maintaining design fidelity in code, rapid frontend development from visual designs.

Key Strengths: Design-to-code automation, React component generation, Tailwind CSS styling, responsive layout handling, component structure preservation, design system integration.

Workflow: Finalize .pen design → Configure export options → Generate React components → Apply Tailwind classes → Review generated code → Integrate into project.

## Quick Decision Guide

Choose Figma MCP when:
- Need to extract design context from existing Figma files
- Working with designers who use Figma
- Required to fetch design tokens and component specifications
- Need screenshots or visual references from Figma
- Documenting existing design systems

Choose Pencil MCP when:
- Creating new designs from scratch
- Rapid prototyping and visual iteration needed
- Text-based design workflow preferred
- Want AI-assisted design generation
- Collaborative design discussions with team

Choose Pencil-to-Code Export when:
- Design is finalized in .pen format
- Ready to implement visual designs as code
- Need React components with Tailwind styling
- Maintaining design fidelity is critical
- Rapid frontend development from designs

## Pencil MCP Workflow

### Starting a Design Session

1. **Check Editor State**
   ```
   get_editor_state() → Determine active .pen file and user selection
   ```

2. **Open or Create Document**
   ```
   open_document(filePathOrNew: "new") → Create new .pen file
   open_document(filePathOrNew: "/path/to/file.pen") → Open existing
   ```

3. **Get Design Guidelines**
   ```
   get_guidelines(topic: "code" | "table" | "tailwind" | "landing-page")
   get_style_guide_tags() → Get available style tags
   get_style_guide(tags: ["minimalist", "dashboard"], name: "nova")
   ```

### Creating Designs

1. **Generate with batch_design**

   Use batch_design for efficient batch operations. Syntax:
   ```
   foo=I("parent", { ... })    // Insert new node
   baz=C("nodeid", "parent", { ... })  // Copy node
   foo2=R("nodeid1/nodeid2", {...})    // Replace node
   U(foo+"/nodeid", {...})     // Update node
   D("dfFAeg2")               // Delete node
   M("nodeid3", "parent", 2)   // Move node
   G("baz", "ai", "...")       // Generate image with AI
   ```

2. **Design with Default Nova Style**

   When creating components without user-specified style:
   - Use neutral color palette (grays, whites)
   - Apply small radius (4-6px)
   - Use Noto Sans or system sans-serif
   - Maintain clean, minimal aesthetic
   - Apply consistent 4px/8px spacing grid

3. **Review with get_screenshot**
   ```
   get_screenshot() → Visual validation of design
   ```

### Managing Design Tokens

1. **Read Variables**
   ```
   get_variables() → Current design tokens and themes
   ```

2. **Update Variables**
   ```
   set_variables(variables: { primary: "#3B82F6", ... })
   ```

### Layout Analysis

```
snapshot_layout() → Analyze computed layout rectangles
find_empty_space_on_canvas(direction: "right", size: { w: 200, h: 100 })
```

## Common Design-to-Code Patterns

### Universal Patterns

These patterns apply across all three tools with tool-specific implementations.

**Design Token Management:**

All tools support design token extraction and management. Figma MCP extracts tokens from existing files, Pencil MCP generates tokens during design creation, Pencil-to-code exports tokens as CSS variables or Tailwind config.

**Component Architecture:**

All tools maintain component hierarchy. Figma MCP reads component structure from Figma, Pencil MCP creates component structure in DNA codes, Pencil-to-code generates React components preserving hierarchy.

**Responsive Design:**

All tools handle responsive layouts. Figma MCP extracts responsive variants, Pencil MCP defines responsive breakpoints in DNA, Pencil-to-code generates Tailwind responsive classes.

**Style Consistency:**

All tools ensure design consistency. Figma MCP validates against design system, Pencil MCP enforces design tokens, Pencil-to-code applies consistent Tailwind classes.

### Workflow Best Practices

Applicable to all tools:

**Design System Integration:**
- Define design tokens before starting design work
- Use consistent naming conventions across tools
- Maintain single source of truth for design values
- Document token usage and component patterns

**Version Control:**
- Commit Figma metadata snapshots for reference
- Version .pen files in repository
- Track design iterations with git
- Document design decisions in code comments

**Collaboration:**
- Use Figma comments for design feedback
- Share .pen frames for visual review
- Create pull requests for design changes
- Maintain design documentation alongside code

**Quality Assurance:**
- Validate design tokens against style guide
- Test responsive breakpoints
- Verify accessibility compliance
- Review generated code for optimization

## Tool-Specific Implementation

For detailed tool-specific implementation guidance, see the reference files:

### Figma MCP Implementation

File: reference/figma.md

Covers Figma MCP connection setup, file metadata fetching, component tree extraction, design token retrieval, screenshot capture, and style guide generation.

Key sections: MCP configuration, authentication setup, file access patterns, metadata queries, component hierarchy parsing, token extraction formats, screenshot API usage, design system documentation.

### Pencil MCP Rendering

File: reference/pencil-renderer.md

Covers batch_design operations, style guide integration, .pen frame rendering, visual design iteration, collaborative workflows, and design version control.

Key sections: batch_design syntax, natural language design prompts, rendering options, frame configuration, design refinement patterns, version control strategies, team collaboration workflows.

### Pencil-to-Code Export

File: reference/pencil-code.md

Covers .pen design export to React components, Tailwind CSS generation, component structure preservation, responsive layout handling, and design system integration.

Key sections: Export configuration, React component generation, Tailwind class application, props API design, state management integration, testing generated components, optimization strategies.

### Tool Comparison

File: reference/comparison.md

Provides detailed comparison matrix covering use cases, workflow patterns, integration complexity, and when to use each tool.

Key sections: Feature comparison table, workflow decision matrix, tool integration patterns, migration strategies, ecosystem compatibility, team workflow recommendations.

## Navigation Guide

When working with design-to-code features:

1. Start with Quick Tool Selection (above) if choosing a tool
2. Apply Default Nova Style when no style specified
3. Review Common Design-to-Code Patterns for universal concepts
4. Open tool-specific reference file for implementation details
5. Refer to comparison.md when evaluating multiple tools
6. Use Context7 tools to access latest tool documentation

## Context7 Documentation Access

Access up-to-date tool documentation using Context7 MCP:

**Figma:**
- Use resolve-library-id with "figma" to get library ID
- Use get-library-docs with topic "mcp", "api", "design-tokens", "metadata"

**Pencil:**
- Use resolve-library-id with "pencil" to get library ID
- Use get-library-docs with topic "mcp", "dna-codes", "rendering", "export"

## Official Documentation

- Pencil Official: https://pencil.dev
- Pencil Docs: https://docs.pencil.dev
- Pencil AI Integration: https://docs.pencil.dev/getting-started/ai-integration

## Works Well With

- moai-domain-uiux: Design systems and component architecture
- moai-domain-frontend: React implementation patterns
- moai-library-shadcn: shadcn/ui component integration (Nova preset)
- moai-lang-typescript: TypeScript for generated components
- moai-lang-react: React best practices
- moai-foundation-core: SPEC-driven development workflows

---

Status: Active
Version: 3.0.0 (Nova Style + Pencil MCP Tools Integration)
Last Updated: 2026-02-21
Tools: Figma MCP, Pencil MCP, Pencil-to-Code Export
Default Style: shadcn/ui Nova (neutral, noto-sans, small radius)
