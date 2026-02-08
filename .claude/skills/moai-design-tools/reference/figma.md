# Figma MCP Implementation Guide

Figma MCP integration for fetching design context, metadata, and specifications from Figma files.

## Overview

Figma MCP provides direct access to Figma file data through the Model Context Protocol, enabling automated extraction of design tokens, component hierarchies, and style information.

## MCP Configuration

### Server Setup

Add to `.mcp.json`:

```json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-figma"],
      "env": {
        "FIGMA_ACCESS_TOKEN": "your-token-here"
      }
    }
  }
}
```

### Authentication

1. Generate Figma Personal Access Token:
   - Go to Figma Settings → Account → Personal Access Tokens
   - Create new token with appropriate permissions
   - Copy token for MCP configuration

2. Environment Variables:
   - Set `FIGMA_ACCESS_TOKEN` in system environment or `.env` file
   - Token requires file read permissions for target files

3. Permission Scopes:
   - `file_read`: Read file content and metadata
   - `file_tokens`: Read design tokens and styles
   - `comments_read`: Read comments and feedback (optional)

## File Access Patterns

### Fetch File Metadata

```typescript
// Get file information
const metadata = await figma.get_file_metadata(fileKey);

// Response structure
{
  "name": "Design System",
  "description": "Main design system file",
  "lastModified": "2026-02-09T10:00:00Z",
  "thumbnailUrl": "https://...",
  "version": "1.0.0"
}
```

### Extract Component Tree

```typescript
// Get component hierarchy
const components = await figma.get_components(fileKey);

// Component structure
{
  "id": "1:2",
  "name": "Button/Primary",
  "type": "COMPONENT",
  "description": "Primary button component",
  "children": [
    {
      "id": "1:3",
      "name": "Label",
      "type": "TEXT"
    }
  ]
}
```

## Design Token Retrieval

### Color Tokens

```typescript
// Extract color styles
const colors = await figma.get_color_styles(fileKey);

// Token format
{
  "name": "primary/500",
  "value": "#3B82F6",
  "description": "Primary brand color"
}
```

### Typography Tokens

```typescript
// Extract text styles
const typography = await figma.get_text_styles(fileKey);

// Token format
{
  "name": "heading/1",
  "fontFamily": "Inter",
  "fontWeight": 700,
  "fontSize": 32,
  "lineHeight": 1.2
}
```

### Spacing Tokens

```typescript
// Extract spacing from layout grids
const spacing = await figma.get_spacing_tokens(fileKey);

// Token format
{
  "name": "spacing/md",
  "value": 16,
  "unit": "pixels"
}
```

## Screenshot Capture

### Capture Node Screenshot

```typescript
// Screenshot specific component
const screenshot = await figma.get_screenshot(
  fileKey,
  nodeId,
  {
    format: "png",
    scale: 2
  }
);

// Returns image URL or base64 data
```

### Batch Screenshots

```typescript
// Capture multiple components
const screenshots = await Promise.all([
  figma.get_screenshot(fileKey, "button-primary"),
  figma.get_screenshot(fileKey, "button-secondary"),
  figma.get_screenshot(fileKey, "input-field")
]);
```

## Style Guide Generation

### Component Documentation

```typescript
// Generate component documentation
const doc = {
  component: "Button/Primary",
  description: "Primary action button",
  props: {
    variant: "primary",
    size: "medium",
    disabled: false
  },
  states: ["default", "hover", "active", "disabled"],
  tokens: {
    background: "primary/500",
    color: "white/100",
    borderRadius: "radius/md",
    padding: "spacing/md"
  }
};
```

### Design System Documentation

```typescript
// Generate design system overview
const designSystem = {
  colors: await figma.get_color_styles(fileKey),
  typography: await figma.get_text_styles(fileKey),
  spacing: await figma.get_spacing_tokens(fileKey),
  components: await figma.get_components(fileKey)
};
```

## Best Practices

### Token Naming

Use semantic naming conventions:
- `color.primary.500` instead of `blue.500`
- `spacing.md` instead of `16px`
- `font.heading.1` instead of `32px bold`

### Version Control

Commit design token snapshots:
```bash
# Export tokens to JSON
figma export-tokens --output design-tokens.json

# Commit to repository
git add design-tokens.json
git commit -m "docs: update design tokens from Figma"
```

### Integration with Code

Map Figma tokens to code:
```css
/* Tailwind config */
module.exports = {
  theme: {
    colors: {
      primary: {
        500: '#3B82F6' /* Figma: primary/500 */
      }
    }
  }
}
```

## Common Workflows

### Workflow 1: Design System Sync

1. Fetch latest tokens from Figma
2. Compare with local token files
3. Update changed tokens
4. Generate documentation
5. Commit changes with diff summary

### Workflow 2: Component Documentation

1. Extract component metadata
2. Capture component screenshots
3. Generate props documentation
4. Create usage examples
5. Publish to component library

### Workflow 3: Style Guide Generation

1. Fetch all style definitions
2. Organize by category (colors, typography, spacing)
3. Generate visual examples
4. Create documentation site
5. Share with development team

## Error Handling

### Access Denied
```
Error: Failed to fetch file metadata (403)
Solution: Verify FIGMA_ACCESS_TOKEN has file read permission
```

### File Not Found
```
Error: File not found (404)
Solution: Verify fileKey is correct and file is shared with token owner
```

### Rate Limiting
```
Error: Rate limit exceeded (429)
Solution: Implement exponential backoff, cache responses
```

## Performance Optimization

### Caching Strategy

```typescript
// Cache design tokens for 1 hour
const cacheKey = `figma:${fileKey}:tokens`;
const cached = await cache.get(cacheKey);

if (cached) {
  return JSON.parse(cached);
}

const tokens = await figma.get_color_styles(fileKey);
await cache.set(cacheKey, JSON.stringify(tokens), 3600);
```

### Batch Requests

```typescript
// Fetch multiple resources in parallel
const [colors, typography, components] = await Promise.all([
  figma.get_color_styles(fileKey),
  figma.get_text_styles(fileKey),
  figma.get_components(fileKey)
]);
```

## Resources

- Figma REST API: https://www.figma.com/developers/api
- Figma MCP Server: https://github.com/modelcontextprotocol/servers/tree/main/src/figma
- Design Tokens Format: https://designtokens.org/format/
- W3C Design Tokens Community Group: https://www.w3.org/community/design-tokens/

---

Last Updated: 2026-02-09
Tool Version: Figma MCP 1.0.0
