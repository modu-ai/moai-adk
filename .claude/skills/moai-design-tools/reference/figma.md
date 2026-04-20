# Figma MCP Integration Guide

Figma MCP integration for fetching design context, creating designs, and managing design systems through the official Figma MCP remote server.

## Overview

Figma MCP provides direct access to Figma file data and design generation capabilities through the Model Context Protocol. The official Figma MCP server enables design context extraction, variable management, FigJam collaboration, and AI-powered design creation (Code-to-Canvas).

## Setup

### Installation

Install the official Figma MCP plugin:

```
claude plugin install figma@claude-plugins-official
```

No manual `mcpServers` configuration is required. The plugin handles connection automatically.

### Remote MCP Server

The official Figma MCP server is hosted at:

```
https://mcp.figma.com/mcp
```

### Authentication

1. Authenticate via the Figma MCP plugin prompt when first connecting
2. Your Figma account credentials are used for authorization
3. Access is scoped to files and projects your account can view

## Figma MCP Tools Reference

### Design Context and Reading

#### get_design_context

Extract design context from Figma files:
- Retrieve component hierarchy and structure
- Understand layout, spacing, and design relationships
- Get detailed specifications for implementing designs

```
get_design_context(fileKey, nodeId?) → { components, layout, styles, ... }
```

#### get_screenshot

Capture screenshots of Figma frames for visual reference:
- Render specific frames as images
- Use as visual reference during implementation
- Compare design intent with code output

```
get_screenshot(fileKey, nodeId) → image data
```

#### get_variable_defs

Extract design variables and tokens from Figma files:
- Color tokens and palettes
- Typography definitions
- Spacing and sizing values
- Theme configurations

```
get_variable_defs(fileKey) → { colors, typography, spacing, ... }
```

#### get_metadata

Get file metadata and structural information:
- File name, description, and timestamps
- Page structure and frame hierarchy
- Component library references

```
get_metadata(fileKey) → { name, description, lastModified, pages, ... }
```

#### whoami

Get current authenticated user information:
- Verify authentication status
- Check user identity and permissions

```
whoami() → { id, name, email, ... }
```

### Code Connect

#### get_code_connect_map

Retrieve code connect mappings that link Figma components to code implementations:
- Map Figma component IDs to code component names
- Reference existing design-to-code connections

```
get_code_connect_map(fileKey) → { componentId: codeComponent, ... }
```

#### add_code_connect_map

Add new code connect mappings to link Figma components with code:
- Register code implementations for Figma components
- Enable bidirectional design-code traceability

```
add_code_connect_map(fileKey, mappings) → confirmation
```

### FigJam and Diagrams

#### get_figjam

Access FigJam boards for collaboration content:
- Retrieve sticky notes, shapes, and text
- Extract workflow diagrams and user flows
- Access collaborative brainstorming sessions

```
get_figjam(fileKey) → { boards, elements, ... }
```

#### generate_diagram

Create diagrams in FigJam from text descriptions:
- Generate flowcharts and architecture diagrams
- Create user journey maps
- Build system design visualizations

```
generate_diagram(description, fileKey?) → { diagramId, ... }
```

### Design Generation

#### generate_figma_design

Create designs in Figma from text descriptions (Code-to-Canvas workflow):
- Generate UI components from natural language
- Create layouts and screens from specifications
- Automate design creation from requirements

```
generate_figma_design(description, targetFileKey?) → { frameId, ... }
```

**Code-to-Canvas (v1) Known Limitations:**
- Japanese text rendering may have issues
- Image dimensions may not exactly match specifications
- Available via Remote MCP server only (https://mcp.figma.com/mcp)

### Design System

#### create_design_system_rules

Create design system rules and guidelines within Figma:
- Define component usage patterns
- Establish naming conventions
- Document design principles

```
create_design_system_rules(rules) → { ruleId, ... }
```

## implement-design Workflow

When implementing a Figma design as code, follow this 7-step workflow:

### Step 1: Get Design Context

```
get_design_context(fileKey, nodeId)
```

Understand the component structure, layout relationships, and design intent before writing any code.

### Step 2: Get Visual Reference

```
get_screenshot(fileKey, nodeId)
```

Capture a screenshot to use as visual reference throughout implementation. Compare final code output against this image.

### Step 3: Extract Design Tokens

```
get_variable_defs(fileKey)
```

Extract all design variables (colors, typography, spacing). Use these values in your implementation instead of hardcoding.

### Step 4: Analyze Component Structure

Review the design context to identify:
- Component hierarchy and nesting
- Responsive behavior and breakpoints
- Interactive states (hover, active, disabled)
- Reusable sub-components

### Step 5: Generate React/Tailwind Code

Implement the component using extracted design context:
- Map Figma components to React components
- Apply design token values from get_variable_defs
- Use Tailwind classes for styling
- Handle responsive layouts with Tailwind breakpoints

### Step 6: Apply Design Tokens

Map extracted variables to Tailwind config:

```typescript
// tailwind.config.js — populated from get_variable_defs output
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: {
          500: '#3B82F6',  /* Figma: colors.primary */
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      }
    }
  }
}
```

### Step 7: Verify Against Screenshot

Compare the implemented component against the screenshot from Step 2:
- Check visual accuracy (colors, spacing, typography)
- Validate responsive behavior at all breakpoints
- Verify interactive states

## Design Token Extraction

### Extracting Color Variables

```
vars = get_variable_defs(fileKey)
// Returns: { colors: { primary: "#3B82F6", ... }, ... }
```

Map to CSS custom properties:
```css
:root {
  --color-primary: #3B82F6;
  --color-primary-hover: #2563EB;
}
```

### Extracting Typography Variables

```
vars = get_variable_defs(fileKey)
// Returns: { typography: { heading1: { fontFamily: "Inter", fontSize: 32, fontWeight: 700 } } }
```

Map to Tailwind:
```typescript
module.exports = {
  theme: {
    extend: {
      fontSize: {
        'heading-1': ['2rem', { lineHeight: '1.2', fontWeight: '700' }],
        'heading-2': ['1.5rem', { lineHeight: '1.3', fontWeight: '600' }],
      }
    }
  }
}
```

## Design-to-Code Workflow Examples

### Example 1: Component Implementation

```
1. get_metadata(fileKey) → Identify target frame and page
2. get_design_context(fileKey, nodeId) → Understand component structure
3. get_screenshot(fileKey, nodeId) → Capture visual reference
4. get_variable_defs(fileKey) → Extract all design tokens
5. Implement React component with extracted specifications
6. Compare implementation screenshot with design screenshot
```

### Example 2: Design System Setup

```
1. get_variable_defs(fileKey) → All design tokens
2. get_code_connect_map(fileKey) → Discover existing code mappings
3. Generate tailwind.config.js from extracted tokens
4. Map Figma component IDs to React component names
5. add_code_connect_map(fileKey, mappings) → Register connections
```

### Example 3: FigJam Workflow Import

```
1. get_figjam(fileKey) → Access workflow diagrams and user flows
2. Parse user journey from FigJam content
3. Implement screens following the user flow
4. Use generate_diagram to create updated architecture docs
```

## Best Practices

### Design Context First

- Always call get_design_context before implementing any Figma design
- Use get_screenshot as a persistent visual reference checkpoint
- Extract all variables with get_variable_defs — never hardcode design values

### Token Naming Conventions

Use semantic naming when mapping Figma variables to code:
- `color.primary.500` instead of `blue`
- `spacing.md` instead of `16px`
- `font.heading.1` instead of `32px bold`

### Code Connect Usage

Register code-component mappings for full traceability:
- Use add_code_connect_map when implementing new components
- Use get_code_connect_map to discover if components are already mapped

### Design Verification

- Always verify implementation against get_screenshot output
- Use create_design_system_rules to document component usage patterns

## Error Handling

### Authentication Issues

```
Error: Not authenticated to Figma
Solution: Re-run: claude plugin install figma@claude-plugins-official
```

### File Not Found

```
Error: File not found or access denied
Solution: Verify fileKey is correct and your Figma account has file access
```

### Node Not Found

```
Error: Node ID not found in file
Solution: Use get_metadata to discover valid node IDs and page structure
```

### Code-to-Canvas Unavailable

```
Error: generate_figma_design not available
Solution: generate_figma_design requires the Remote MCP server (https://mcp.figma.com/mcp)
```

## Resources

- Figma MCP Remote Server: https://mcp.figma.com/mcp
- Figma Developers: https://www.figma.com/developers
- Design Tokens Format: https://designtokens.org/format/

---

Last Updated: 2026-03-11
Tool Version: Figma MCP (Official Remote Server)
