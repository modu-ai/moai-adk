---
name: moai-design-tools
description: >
  Design tool integration specialist covering Figma MCP, Pencil renderer, and Pencil-to-code export.
  Use when fetching design context from Figma, rendering Pencil DNA codes to .pen frames, exporting .pen
  designs to React/Tailwind code, or choosing design-to-code workflows. Supports design fetching (Figma),
  visual rendering (Pencil MCP), and code generation (React/Tailwind).
license: MIT
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob Bash WebFetch WebSearch mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.0.0"
  category: "domain"
  status: "active"
  updated: "2026-02-09"
  modularized: "false"
  tools: "Figma, Pencil MCP"
  tags: "figma, pencil, design to code, design export, render dna, pen frame, react from design, tailwind from design, design context, ui implementation"
  context7-libraries: "/figma/docs, /pencil/docs"
  related-skills: "moai-domain-uiux, moai-domain-frontend, moai-library-shadcn, moai-lang-typescript, moai-lang-react"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4500

# MoAI Extension: Triggers
triggers:
  keywords: ["figma", "pencil", "design to code", "design export", "render dna", "pen frame", "react from design", "tailwind from design", "design context", "ui implementation", "design fetching", "figma mcp", "pencil mcp", "component from design", "layout from design"]
  agents: ["expert-frontend", "team-designer"]
  phases: ["run"]
---

# Design Tools Integration Specialist

Comprehensive design-to-code workflow guidance covering three major capabilities: Figma MCP (design fetching), Pencil MCP (visual rendering), and Pencil-to-code export (React/Tailwind generation).

## Quick Tool Selection

### Figma MCP - Design Context Fetching

Figma integration for fetching design context, metadata, and screenshots from Figma files.

Best For: Fetching design tokens, component specifications, layout information, and style guides from existing Figma files. Extracting design system values and understanding design structure.

Key Strengths: Direct access to Figma file metadata, component hierarchy extraction, style guide generation, design token retrieval, screenshot capture for visual reference.

Workflow: Connect to Figma file → Fetch file metadata → Extract component tree → Retrieve design tokens → Generate style documentation.

Context7 Library: /figma/docs

### Pencil MCP - Visual Design Rendering

Pencil MCP integration for rendering DNA codes into visual .pen frames and creating design proposals.

Best For: Rapid prototyping, visual design iterations, creating UI mockups from text descriptions, collaborative design discussions, visual proposals before implementation.

Key Strengths: Text-to-design conversion, DNA code format for version control, iterative design refinement, visual preview without implementation, collaborative design workflow.

Workflow: Describe UI in natural language → Generate DNA code → Render to .pen frame → Visually review → Iterate on design → Export to code when ready.

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
- Want version-controllable design format (DNA codes)
- Collaborative design discussions with team

Choose Pencil-to-Code Export when:
- Design is finalized in .pen format
- Ready to implement visual designs as code
- Need React components with Tailwind styling
- Maintaining design fidelity is critical
- Rapid frontend development from designs

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
- Version DNA codes in repository
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

Covers DNA code format and syntax, text-to-design generation, .pen frame rendering, visual design iteration, collaborative workflows, and design version control.

Key sections: DNA code structure, natural language design prompts, rendering options, frame configuration, design refinement patterns, version control strategies, team collaboration workflows.

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
2. Review Common Design-to-Code Patterns for universal concepts
3. Open tool-specific reference file for implementation details
4. Refer to comparison.md when evaluating multiple tools
5. Use Context7 tools to access latest tool documentation

## Context7 Documentation Access

Access up-to-date tool documentation using Context7 MCP:

**Figma:**
- Use resolve-library-id with "figma" to get library ID
- Use get-library-docs with topic "mcp", "api", "design-tokens", "metadata"

**Pencil:**
- Use resolve-library-id with "pencil" to get library ID
- Use get-library-docs with topic "mcp", "dna-codes", "rendering", "export"

## Works Well With

- moai-domain-uiux: Design systems and component architecture
- moai-domain-frontend: React implementation patterns
- moai-library-shadcn: shadcn/ui component integration
- moai-lang-typescript: TypeScript for generated components
- moai-lang-react: React best practices
- moai-foundation-core: SPEC-driven development workflows

---

Status: Active
Version: 2.0.0 (Consolidated Design Tools Coverage)
Last Updated: 2026-02-09
Tools: Figma MCP, Pencil MCP, Pencil-to-Code Export
