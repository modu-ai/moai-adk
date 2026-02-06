# Node Mapping Reference

Complete mapping from .pen node types to React/Tailwind output.

## Frame Node

Most common node type. Maps to `<div>` with flex/grid layout.

### Basic Frame
```json
{
  "type": "frame",
  "layout": "vertical",
  "gap": 16,
  "padding": [24, 32, 24, 32]
}
```
```tsx
<div className="flex flex-col gap-4 py-6 px-8">
  {children}
</div>
```

### Frame with Fill
```json
{
  "type": "frame",
  "fill": "#1A1A1A",
  "cornerRadius": [12, 12, 12, 12]
}
```
```tsx
<div className="bg-[#1A1A1A] rounded-xl">
  {children}
</div>
```

### Frame with Border
```json
{
  "type": "frame",
  "stroke": "#E5E5E5",
  "strokeThickness": 1,
  "cornerRadius": [8, 8, 8, 8]
}
```
```tsx
<div className="border border-gray-200 rounded-lg">
  {children}
</div>
```

### Frame with Gradient Fill
```json
{
  "type": "frame",
  "fill": {
    "type": "linear",
    "angle": 135,
    "stops": [
      {"position": 0, "color": "#1E3A5F"},
      {"position": 1, "color": "#0A1628"}
    ]
  }
}
```
```tsx
<div className="bg-gradient-to-br from-[#1E3A5F] to-[#0A1628]">
  {children}
</div>
```

### Grid Frame
```json
{
  "type": "frame",
  "layout": "grid",
  "columns": 3,
  "gap": 24
}
```
```tsx
<div className="grid grid-cols-3 gap-6">
  {children}
</div>
```

## Text Node

Maps to semantic HTML elements based on context.

### Heading Detection
```json
{
  "type": "text",
  "content": "Hero Title",
  "fontSize": 48,
  "fontWeight": 700
}
```
```tsx
// fontSize >= 32 && fontWeight >= 600 → heading
<h1 className="text-5xl font-bold">Hero Title</h1>

// Use h2-h6 based on hierarchy position in document
```

### Body Text
```json
{
  "type": "text",
  "content": "Description paragraph...",
  "fontSize": 16,
  "fontWeight": 400,
  "lineHeight": 1.6
}
```
```tsx
<p className="text-base font-normal leading-relaxed">
  Description paragraph...
</p>
```

### Styled Text
```json
{
  "type": "text",
  "content": "Label",
  "fontSize": 14,
  "fontWeight": 500,
  "textColor": "#6B7280"
}
```
```tsx
<span className="text-sm font-medium text-gray-500">Label</span>
```

### Text with Custom Font
```json
{
  "type": "text",
  "fontFamily": "Playfair Display",
  "fontSize": 64
}
```
```tsx
// Add to Tailwind theme
// --font-display: "Playfair Display", serif;
<h1 className="font-display text-6xl">...</h1>
```

## Rectangle Node

Simple shape, maps to styled `<div>`.

```json
{
  "type": "rectangle",
  "width": 100,
  "height": 4,
  "fill": "#2563EB",
  "cornerRadius": [2, 2, 2, 2]
}
```
```tsx
<div className="w-[100px] h-1 bg-blue-600 rounded-sm" />
```

## Ellipse Node

Maps to `<div>` with `rounded-full`.

```json
{
  "type": "ellipse",
  "width": 48,
  "height": 48,
  "fill": "#10B981"
}
```
```tsx
<div className="w-12 h-12 bg-emerald-500 rounded-full" />
```

## Image Node

Maps to `<img>` or background image.

### As Element
```json
{
  "type": "image",
  "src": "https://...",
  "width": 400,
  "height": 300
}
```
```tsx
<img
  src="https://..."
  alt=""
  className="w-[400px] h-[300px] object-cover"
/>
```

### As Background (frame with image fill)
```json
{
  "type": "frame",
  "fill": {"type": "image", "src": "..."},
  "width": 400,
  "height": 300
}
```
```tsx
<div
  className="w-[400px] h-[300px] bg-cover bg-center"
  style={{ backgroundImage: 'url(...)' }}
/>
```

## Ref Node (Component Instance)

Reusable component. Extract as separate component.

```json
{
  "type": "ref",
  "ref": "ButtonComponent",
  "descendants": {
    "label": {"content": "Submit"}
  }
}
```
```tsx
// 1. First, extract ButtonComponent from its definition
// 2. Use as:
<Button>Submit</Button>

// If customization needed:
<Button variant="primary">Submit</Button>
```

## Component Extraction

When a frame has `reusable: true`, extract as component:

### Source (.pen)
```json
{
  "id": "ButtonDef",
  "type": "frame",
  "reusable": true,
  "layout": "horizontal",
  "alignItems": "center",
  "padding": [12, 24, 12, 24],
  "cornerRadius": [8, 8, 8, 8],
  "fill": "#2563EB",
  "children": [
    {
      "id": "label",
      "type": "text",
      "content": "Button",
      "textColor": "#FFFFFF"
    }
  ]
}
```

### Output (React)
```tsx
// components/Button.tsx
interface ButtonProps {
  children: React.ReactNode
  className?: string
}

export function Button({ children, className }: ButtonProps) {
  return (
    <button
      className={cn(
        "flex items-center py-3 px-6 rounded-lg bg-blue-600 text-white",
        className
      )}
    >
      {children}
    </button>
  )
}
```

## Layout Algorithm

### Absolute → Flex Conversion

When `.pen` uses absolute positioning (`layout: "none"`) but children
have logical arrangement, convert to flex:

```json
// If children are vertically stacked with similar gaps
// Convert to flex-col with gap
{
  "type": "frame",
  "layout": "none",
  "children": [
    {"y": 0, ...},
    {"y": 50, ...},
    {"y": 100, ...}
  ]
}
```
```tsx
// Detect consistent 50px gaps → gap-[50px] or closest Tailwind
<div className="flex flex-col gap-12">...</div>
```

### Nested Flex

```json
{
  "type": "frame",
  "layout": "horizontal",
  "children": [
    {"type": "frame", "layout": "vertical", ...},
    {"type": "frame", "layout": "vertical", ...}
  ]
}
```
```tsx
<div className="flex flex-row">
  <div className="flex flex-col">...</div>
  <div className="flex flex-col">...</div>
</div>
```

## Semantic HTML Selection

| .pen Pattern | HTML Element |
|-------------|--------------|
| Large text (32px+), bold, at top | `<h1>` - `<h3>` |
| Medium text (20-28px), in section | `<h4>` - `<h6>` |
| Paragraph-length body text | `<p>` |
| Short labels or inline text | `<span>` |
| Frame with click handler noted | `<button>` |
| Frame wrapping navigation items | `<nav>` |
| Frame as page section | `<section>` |
| Frame as header area | `<header>` |
| Frame as footer area | `<footer>` |

## Accessibility Annotations

Add accessibility attributes based on design patterns:

```tsx
// Button-like frames
<button aria-label="Submit form">...</button>

// Icon-only elements
<span aria-hidden="true">...</span>

// Decorative images
<img alt="" role="presentation" />

// Interactive cards
<article role="article">...</article>
```

## Value Conversion Tables

### Gap → Tailwind
| .pen | Tailwind |
|------|----------|
| 4 | gap-1 |
| 8 | gap-2 |
| 12 | gap-3 |
| 16 | gap-4 |
| 20 | gap-5 |
| 24 | gap-6 |
| 32 | gap-8 |
| 40 | gap-10 |
| 48 | gap-12 |

### Font Size → Tailwind
| .pen | Tailwind |
|------|----------|
| 12 | text-xs |
| 14 | text-sm |
| 16 | text-base |
| 18 | text-lg |
| 20 | text-xl |
| 24 | text-2xl |
| 30 | text-3xl |
| 36 | text-4xl |
| 48 | text-5xl |
| 60 | text-6xl |
| 72 | text-7xl |

### Width/Height
```
fill_container → w-full / h-full
fixed number   → w-[Npx] / h-[Npx]
               → or closest Tailwind class (w-64, h-48, etc.)
```
