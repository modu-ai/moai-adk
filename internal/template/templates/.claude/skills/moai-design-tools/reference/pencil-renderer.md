# Pencil MCP Rendering Guide

Pencil MCP integration for creating and editing .pen design files with AI-assisted design generation.

## Overview

Pencil MCP provides a comprehensive set of tools for creating, editing, and managing .pen design files. The editor is specifically designed for web and mobile applications with AI-assisted design generation capabilities.

**Important Notes:**
- .pen files are pure JSON (Git diffable, mergeable) — schema version **2.9**
- ALWAYS use Pencil MCP tools (batch_get, batch_design) for .pen file operations — they provide structured access to the design graph
- Pencil MCP auto-configures — no manual `mcpServers` configuration required
- Pencil is a local-first tool — design files stay on your machine, MCP server runs locally

### Supported Node Types

| Type | Description |
|------|-------------|
| `Rectangle` | Basic rectangle shape |
| `Ellipse` | Elliptical/circular shape |
| `Line` | Single line |
| `Polygon` | Multi-sided polygon |
| `Path` | Bezier path (SVG-like) |
| `Text` | Text content with typography |
| `Frame` | Container with layout (flexbox) |
| `Group` | Visual grouping without layout |
| `Note` | Annotation/comment node |
| `Prompt` | AI prompt node |
| `Context` | Context information node |
| `IconFont` | Icon font glyph |
| `Ref` | Component instance (references a `reusable: true` component) |

### .pen Format Key Properties

| Property | Description |
|----------|-------------|
| `layout` | none / vertical / horizontal (flexbox) |
| `gap`, `padding` | Spacing values |
| `justifyContent`, `alignItems` | Flex alignment |
| `SizingBehavior` | Fixed or dynamic sizing |
| `textGrowth` | auto / fixed-width / fixed-width-height |
| `layoutPosition` | auto / absolute |
| `clip` | Boolean — clip content overflow |
| `opacity`, `flipX`, `flipY`, `rotation` | Transform properties |
| `enabled` | Boolean — visibility toggle |
| `metadata` | Custom context key-value pairs |
| `reusable` | Boolean — marks node as component origin |
| `cornerRadius` | Single value or 4-value array |

### Graphics Capabilities

**Fill types:** Solid color, linear gradient, radial gradient, angular gradient, image, mesh_gradient

**Stroke:** Single stroke with multiple fills; configurable thickness, alignment (inside/center/outside), join style (miter/round/bevel), dash pattern

**Effects:** Blur, background blur (frosted glass), inner shadow, outer shadow

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
- Returns JSON with component type, style, layout, children hierarchy

#### get_screenshot()

Render a visual preview of a node in a .pen file:
- Use periodically to validate designs visually
- Returns PNG image data for review

#### snapshot_layout()

Check the current layout structure of a .pen file:
- Examine computed layout rectangles with bounding boxes
- Decide where to insert new nodes
- Detect overlapping elements
- Understand spatial relationships

#### get_variables()

Extract current state of variables and themes:
- Design tokens (colors, spacing, border radii, sizes, fonts)
- Theme configuration (light/dark mode variables)
- Variable types: boolean, color, number, string
- Variables reference syntax: `$variable-name`

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

### Variable Creation Methods

Variables can be created through three methods:

1. **Manual Definition**: Define directly via `set_variables` or the variables panel
2. **From CSS**: AI agent extracts variables from `globals.css` (colors, spacing, fonts)
3. **From Figma**: Paste screenshots of variable tables or copy token values

**Bidirectional Sync**: Pencil supports two-way sync between .pen variables and CSS custom properties. AI assistant can update CSS based on .pen files, and CSS changes can be imported back.

### Property Search and Replace

#### search_all_unique_properties(nodeIds)

Recursively search for all unique properties on nodes:
- Discover unique property values across design elements
- Useful for auditing design consistency
- Find all colors, fonts, spacing values used in a design

#### replace_all_matching_properties(match, replace)

Recursively replace all matching properties on nodes:
- Bulk update design properties across the entire design
- Useful for theme changes, color updates, font replacements
- Maintains consistency when updating design tokens

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
- `design-system` - Design system guidelines and patterns

#### get_style_guide_tags()

Returns all available style guide tags for filtering:
- Use to discover available style options
- Filter style guides by relevant tags

#### get_style_guide(tags, name)

Returns a style guide based on tags or specific name:
- Use when designing screens, websites, apps, or dashboards
- Apply consistent styling across designs

## Component System

### Creating Components

Convert any design element (frame, shape, text) into a reusable component:
- Keyboard shortcut: **Cmd/Ctrl + Option/Alt + K**
- Component origins display with **magenta** bounding box
- Component instances display with **violet** bounding box
- "Go to component" button in properties panel navigates to origin

### Component Instances (Ref Type)

Copy a component origin to create an instance (`type: "ref"`):
- Instances reference the origin by ID
- Property overrides via `descendants` object using ID paths
- Full object replacement when `type` is specified in descendant
- Children replacement via `children` array in descendant
- Nested components are supported

### Slots

Slots are designated drop zones within components for flexible customization:
- Create: Select an **empty frame** in a component origin → "Make a slot" button
- Visual indicator: Diagonal lines on canvas mark slot areas
- Only empty frames in component origins can become slots

**Suggested Slot Components**: Mark other components as recommended content for specific slots. Provides guidance for both human designers and AI agents about intended content.

### Design Libraries (.lib.pen)

Reusable component collections shared across multiple .pen files:
- Create: Layers panel → Libraries icon → "Turn this file into a library"
- File extension: `.lib.pen` (irreversible designation)
- Import: Layers panel → Libraries icon → Select library
- Browse: Assets icon in layers panel (grid view + search by name)
- Default libraries are available out of the box
- Changes to library components auto-propagate to all instances

## UI Kit Options

Pencil MCP supports multiple UI kits for different design aesthetics. Choose the kit that best matches your project requirements:

| UI Kit | Description | Best For |
|--------|-------------|----------|
| **Shadcn UI** | Modern, accessible components with Radix UI primitives | Default for Nova preset; dashboards, admin panels |
| **Halo** | Glassmorphic design language with blur and transparency effects | Marketing sites, premium products |
| **Lunaris** | Dark-mode focused with high contrast and deep colors | Developer tools, creative applications |
| **Nitro** | Performance-optimized minimal design with minimal DOM footprint | High-traffic apps, performance-critical UIs |

Apply via style guide:
```
guide = get_style_guide(name: "shadcn-nova")  // or "halo", "lunaris", "nitro"
```

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

## Pencil CLI

**Install:** `npm install -g @pencil.dev/cli` (requires Node.js 18+)
**Latest Version:** 0.2.4

### Authentication

```
pencil login         # Interactive email/OTP login (stores in ~/.pencil/session-cli.json)
pencil status        # Verify session and account info
pencil version       # Show CLI version
```

For CI/CD, set `PENCIL_CLI_KEY` environment variable instead.

### Agent Mode

Run AI-assisted design generation from command line:

```
pencil --out login.pen --prompt "Create modern login page"
pencil --in dashboard.pen --out dashboard-v2.pen --prompt "Add sidebar"
pencil --out simple.pen --model claude-haiku-4-5 --prompt "404 page"
pencil --in design.pen --export hero.png --export-scale 2
```

| Flag | Description |
|------|-------------|
| `--in, -i <path>` | Input .pen file (optional) |
| `--out, -o <path>` | Output .pen file (required unless exporting) |
| `--prompt, -p <text>` | AI agent prompt |
| `--model, -m <id>` | Model: claude-opus-4-6 (default), claude-sonnet-4-6, claude-haiku-4-5 |
| `--export, -e <path>` | Export to PNG/JPEG/WEBP/PDF |
| `--export-scale <n>` | Scale multiplier for export |
| `--export-type <type>` | Image format |
| `--tasks, -t <path>` | Batch JSON file for multiple designs |
| `--custom, -c` | Custom Claude config |
| `--list-models` | List available models |
| `--verbose-mcp` | Full MCP error logging |

### Interactive Mode

```
pencil interactive              # Headless mode (local editor without GUI)
pencil interactive --app vscode # App mode (live WebSocket to running Pencil)
```

**Shell commands:**
- `tool_name({ key: value })` — Call MCP tool with parameters
- `save()` — Write changes to disk
- `exit()` — Exit shell

### Batch Processing

JSON task file format:
```json
{
  "tasks": [
    { "out": "page.pen", "prompt": "Create landing page", "model": "claude-haiku-4-5" }
  ]
}
```
Required: `out`, `prompt`. Optional: `in`, `model`.

### Export (CLI Only)

The `export_nodes` tool is available only through the CLI (not via MCP):
- Formats: PNG, JPEG, WEBP, PDF
- Scaling support via `--export-scale`

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `PENCIL_CLI_KEY` | Organization API key for CI/CD |
| `ANTHROPIC_API_KEY` | Anthropic authentication |
| `PENCIL_API_BASE` | Backend URL (default: https://api.pencil.dev) |
| `DEBUG` | Enable debug logging |

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

Last Updated: 2026-04-05
Tool Version: Pencil MCP (14 tools + export_nodes CLI-only)
CLI Version: @pencil.dev/cli 0.2.4
.pen Schema: 2.9
Default Style: shadcn/ui Nova (neutral, noto-sans, small radius)
