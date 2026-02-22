# Pencil MCP Rendering Guide

Pencil MCP integration for creating and editing .pen design files with AI-assisted design generation.

## Overview

Pencil MCP provides a comprehensive set of tools for creating, editing, and managing .pen design files. The editor is specifically designed for web and mobile applications with AI-assisted design generation capabilities.

**Important Notes:**
- .pen file contents are encrypted and can ONLY be accessed via Pencil MCP tools
- NEVER use Read or Grep tools to read .pen file contents
- ALWAYS use Pencil MCP tools (batch_get, batch_design) for .pen file operations

## Pencil MCP Tools Reference

### Editor State and Document Management

#### get_editor_state()

Start with this tool to understand the current editor state:
- Currently active .pen file
- User's current selection
- Other essential context information

```
get_editor_state() → { activeFile, selection, ... }
```

#### open_document(filePathOrNew)

Open or create .pen files:
- `"new"` → Create new empty .pen file
- `"/path/to/file.pen"` → Open existing file

### Design Reading Tools

#### batch_get(patterns, nodeIds)

Retrieve nodes by searching patterns or reading specific node IDs:
- Use for discovering and understanding .pen file structure
- Supports pattern matching for efficient searching

#### get_screenshot()

Render a visual preview of a node in a .pen file:
- Use periodically to validate designs visually
- Returns image data for review

#### snapshot_layout()

Check the current layout structure of a .pen file:
- Examine computed layout rectangles
- Decide where to insert new nodes
- Understand spatial relationships

#### get_variables()

Extract current state of variables and themes:
- Design tokens
- Color definitions
- Theme configuration

### Design Creation and Modification

#### batch_design(operations)

Execute multiple design operations in a single call. **Maximum 25 operations per call recommended.**

**Operation Syntax:**

| Operation | Syntax | Description |
|-----------|--------|-------------|
| Insert | `foo=I("parent", { ... })` | Create new node |
| Copy | `baz=C("nodeid", "parent", { ... })` | Copy existing node |
| Replace | `foo2=R("nodeid1/nodeid2", {...})` | Replace node content |
| Update | `U(foo+"/nodeid", {...})` | Update existing node |
| Delete | `D("dfFAeg2")` | Remove node |
| Move | `M("nodeid3", "parent", 2)` | Move node to new parent |
| Generate Image | `G("baz", "ai", "...")` | AI image generation |

**Example:**
```
// Create a button component
button=I("root", {
  type: "frame",
  name: "Button",
  style: { backgroundColor: "#3B82F6", borderRadius: 6 }
})
U(button, { children: ["Click me"] })
```

#### set_variables()

Add or update variables in the .pen file:
- Define color tokens
- Set theme values
- Configure design system variables

### Layout and Space Management

#### find_empty_space_on_canvas()

Find available space on the canvas:
- Direction parameter for search direction
- Size parameter for desired dimensions
- Returns coordinates for placement

### Style Guide Integration

#### get_guidelines(topic)

Returns design guidelines and rules for working with .pen files.

**Available Topics:**
- `code` - Code-related design guidelines
- `table` - Table design patterns
- `tailwind` - Tailwind CSS integration
- `landing-page` - Landing page design patterns

#### get_style_guide_tags()

Returns all available style guide tags for filtering:
- Use to discover available style options
- Filter style guides by relevant tags

#### get_style_guide(tags, name)

Returns a style guide based on tags or specific name:
- Use when designing screens, websites, apps, or dashboards
- Apply consistent styling across designs

## Default Style: shadcn/ui Nova

### Nova Preset Configuration

When no specific style is requested, use the Nova preset with these defaults:

```
bunx --bun shadcn@latest create --preset "https://ui.shadcn.com/init?base=radix&style=nova&baseColor=neutral&theme=neutral&iconLibrary=hugeicons&font=noto-sans&menuAccent=bold&menuColor=default&radius=small&template=next&rtl=false" --template next
```

### Nova Style Tokens

```javascript
// Color Palette (Neutral/Notion-style)
const novaColors = {
  // Backgrounds
  background: "#FFFFFF",
  surface: "#FAFAFA",
  surfaceHover: "#F5F5F5",
  surfaceActive: "#EBEBEB",

  // Borders
  border: "#E5E5E5",
  borderHover: "#D4D4D4",

  // Text
  textPrimary: "#171717",
  textSecondary: "#525252",
  textTertiary: "#A3A3A3",

  // Accent (subtle blue)
  accent: "#3B82F6",
  accentHover: "#2563EB",
  accentLight: "#EFF6FF"
};

// Spacing (4px base)
const novaSpacing = {
  xs: "4px",
  sm: "8px",
  md: "16px",
  lg: "24px",
  xl: "32px"
};

// Border Radius (small, subtle)
const novaRadius = {
  sm: "4px",
  md: "6px",
  lg: "8px"
};

// Typography
const novaTypography = {
  fontFamily: "'Noto Sans', system-ui, sans-serif",
  fontSize: {
    xs: "12px",
    sm: "14px",
    md: "16px",
    lg: "18px",
    xl: "20px"
  },
  fontWeight: {
    normal: 400,
    medium: 500,
    semibold: 600,
    bold: 700
  }
};
```

### Applying Nova Style in batch_design

```
// Create a Nova-styled card
card=I("parent", {
  type: "frame",
  name: "Card",
  style: {
    backgroundColor: "#FFFFFF",
    borderRadius: 6,
    borderWidth: 1,
    borderColor: "#E5E5E5",
    padding: 16,
    boxShadow: "0 1px 3px rgba(0,0,0,0.1)"
  }
})

// Add heading with Nova typography
heading=I(card, {
  type: "text",
  content: "Card Title",
  style: {
    fontFamily: "'Noto Sans', sans-serif",
    fontSize: 18,
    fontWeight: 600,
    color: "#171717"
  }
})

// Add body text
body=I(card, {
  type: "text",
  content: "Card description text here.",
  style: {
    fontFamily: "'Noto Sans', sans-serif",
    fontSize: 14,
    fontWeight: 400,
    color: "#525252"
  }
})
```

## Workflow Patterns

### Starting a New Design

1. **Initialize Editor State**
   ```
   state = get_editor_state()
   ```

2. **Create or Open Document**
   ```
   open_document("new")  // or path to existing
   ```

3. **Get Style Guidelines**
   ```
   tags = get_style_guide_tags()
   guide = get_style_guide(tags: ["minimalist", "neutral"])
   ```

4. **Set Design Tokens**
   ```
   set_variables(novaColors)
   set_variables(novaSpacing)
   ```

### Creating a Component

1. **Find Space on Canvas**
   ```
   space = find_empty_space_on_canvas(direction: "right", size: { w: 400, h: 300 })
   ```

2. **Design with batch_design**
   ```
   component=I("root", { ... })
   U(component, { ... })
   ```

3. **Visual Validation**
   ```
   screenshot = get_screenshot()
   ```

4. **Iterate as Needed**
   ```
   U(component + "/child", { ... })
   ```

### Analyzing Existing Designs

1. **Get Layout Structure**
   ```
   layout = snapshot_layout()
   ```

2. **Read Design Elements**
   ```
   nodes = batch_get(patterns: ["Button", "Card"])
   ```

3. **Extract Variables**
   ```
   vars = get_variables()
   ```

## Best Practices

### batch_design Operations

- Maximum 25 operations per call
- Group related operations together
- Use variable references for node IDs
- Build incrementally, validate with screenshots

### Style Consistency

- Always use get_style_guide before designing
- Apply Nova preset as default
- Maintain consistent spacing and typography
- Use design tokens from get_variables

### Performance

- Batch operations efficiently
- Use patterns in batch_get for searching
- Cache style guide information
- Minimize redundant screenshot calls

## Error Handling

### Common Issues

| Issue | Solution |
|-------|----------|
| "Cannot read .pen file" | Use batch_get, never Read tool |
| "Node not found" | Check node ID with batch_get |
| "Invalid operation syntax" | Verify batch_design syntax |
| "Style not applied" | Check variable names match |

### Validation Pattern

```
// Always validate after batch operations
batch_design([...])
screenshot = get_screenshot()
// Review screenshot for correctness
```

## Resources

- Pencil Official: https://pencil.dev
- Pencil Documentation: https://docs.pencil.dev
- Pencil AI Integration: https://docs.pencil.dev/getting-started/ai-integration
- shadcn/ui: https://ui.shadcn.com
- shadcn Nova Style: https://ui.shadcn.com/docs/components

---

Last Updated: 2026-02-21
Tool Version: Pencil MCP (Latest)
Default Style: shadcn/ui Nova (neutral, noto-sans, small radius)
