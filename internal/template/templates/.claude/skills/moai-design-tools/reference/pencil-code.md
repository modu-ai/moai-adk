# Pencil-to-Code Workflow Guide

Convert .pen designs to production-ready React/Tailwind code using a prompt-based workflow.

## Overview

Pencil-to-Code is a **prompt-based workflow**, not an export API. There is no `pencil.export_to_react()` function or `pencil.config.js` configuration file. Instead, you use Pencil MCP tools to read .pen frame data, analyze the JSON structure, and generate code that faithfully implements the design.

.pen files are pure JSON (Git diffable, mergeable), making them directly accessible via Pencil MCP tools. The Pencil MCP provides structured access to the design graph through well-defined tool calls.

## Workflow Steps

### Step 1: Retrieve Frame Data with batch_get

Use `batch_get` to read the .pen frame structure:

```
frames = batch_get(patterns: ["ComponentName"])
// or by node ID:
frames = batch_get(nodeIds: ["node-id-123"])
```

The returned JSON structure contains:
- Component type and name
- Style properties (colors, typography, spacing, borders)
- Layout configuration (flexbox direction, gap, alignment, dimensions)
- Child node hierarchy with nested components

### Step 2: Analyze Component Structure

Parse the JSON to understand:
- Nesting hierarchy and parent-child relationships
- Applied style values (map to design tokens where available)
- Component types (frame, text, image, button, input)
- Responsive behavior settings

Use `snapshot_layout` for computed layout information:

```
layout = snapshot_layout()
// Returns computed rectangles and spatial relationships
```

### Step 3: Map Pencil Components to React/Tailwind

Use the component mapping reference to translate Pencil node types to React primitives:

| Pencil Node Type | React Component | Tailwind Classes |
|------------------|-----------------|------------------|
| `frame` (flex row) | `<div>` | `flex flex-row` |
| `frame` (flex col) | `<div>` | `flex flex-col` |
| `frame` (grid) | `<div>` | `grid` |
| `text` (body) | `<p>` or `<span>` | `text-sm text-gray-700` |
| `text` (heading) | `<h1>`–`<h6>` | `text-xl font-semibold` |
| `image` | `<img>` or `next/image` | `object-cover` |
| `button` frame | `<button>` | `btn` pattern |
| `input` frame | `<input>` | `input` pattern |
| `card` frame | `<div>` | `rounded-lg border bg-white p-4` |

### Step 4: Generate Code with Design Token Mapping

Extract design tokens first, then generate code using those values:

```
vars = get_variables()
// Returns: { primary: "#3B82F6", spacing_md: 16, font_sans: "Noto Sans", ... }
```

Map tokens to Tailwind config:

```javascript
// tailwind.config.js — generated from get_variables() output
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: {
          50:  '#eff6ff',
          500: '#3b82f6',  /* From .pen design tokens */
          600: '#2563eb',
        }
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
      }
    }
  }
};
```

## Design Token Extraction

Use `get_variables` before generating code to avoid hardcoded values:

```
vars = get_variables()
```

Returns the current state of:
- Color definitions and palette
- Typography values (font family, size, weight)
- Spacing tokens
- Theme configuration (light/dark mode variables)

### Mapping to CSS Custom Properties

```css
/* From get_variables() output */
:root {
  --color-primary: #3B82F6;
  --color-primary-hover: #2563EB;
  --color-text-primary: #171717;
  --color-text-secondary: #525252;
  --spacing-md: 16px;
  --radius-md: 6px;
  --font-sans: 'Noto Sans', system-ui, sans-serif;
}
```

## Tailwind CSS Integration

### Responsive Classes

```tsx
// Responsive grid from .pen frame analysis
<div className="
  grid
  grid-cols-1
  md:grid-cols-2
  lg:grid-cols-3
  gap-4
  md:gap-6
  lg:gap-8
">
  {/* Cards */}
</div>
```

### Design Token Application

```tsx
// Nova style (default) — from get_variables() output
const card = (
  <div className="
    bg-white
    rounded-md
    border border-neutral-200
    p-4
    shadow-sm
  ">
    <h3 className="text-lg font-semibold text-neutral-900">
      Card Title
    </h3>
    <p className="text-sm text-neutral-500 mt-1">
      Card description text here.
    </p>
  </div>
);
```

## Component Code Patterns

### Pattern 1: Button Component

```tsx
// From .pen button frame analysis
export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'tertiary';
  size?: 'small' | 'medium' | 'large';
  isLoading?: boolean;
}

export const Button = ({
  variant = 'primary',
  size = 'medium',
  isLoading,
  children,
  ...props
}: ButtonProps) => {
  const base = 'inline-flex items-center justify-center font-medium rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2';

  const variants = {
    primary:   'bg-blue-600 text-white hover:bg-blue-700 focus-visible:ring-blue-500',
    secondary: 'bg-neutral-100 text-neutral-900 hover:bg-neutral-200 focus-visible:ring-neutral-500',
    tertiary:  'bg-transparent text-neutral-700 hover:bg-neutral-100 focus-visible:ring-neutral-500',
  };

  const sizes = {
    small:  'px-3 py-1.5 text-sm',
    medium: 'px-4 py-2 text-base',
    large:  'px-6 py-3 text-lg',
  };

  return (
    <button
      className={`${base} ${variants[variant]} ${sizes[size]} ${isLoading ? 'opacity-75 cursor-not-allowed' : ''}`}
      disabled={isLoading}
      {...props}
    >
      {children}
    </button>
  );
};
```

### Pattern 2: Form Components

```tsx
// From .pen form frame analysis
export const FormField = ({
  label,
  error,
  id,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & { label: string; error?: string }) => (
  <div className="space-y-1">
    <label htmlFor={id} className="block text-sm font-medium text-neutral-700">
      {label}
    </label>
    <input
      id={id}
      className={`
        block w-full rounded-md border px-3 py-2 text-sm
        focus:outline-none focus:ring-2 focus:ring-offset-0
        ${error
          ? 'border-red-500 focus:ring-red-500'
          : 'border-neutral-300 focus:ring-blue-500'
        }
      `}
      aria-invalid={!!error}
      aria-describedby={error ? `${id}-error` : undefined}
      {...props}
    />
    {error && (
      <p id={`${id}-error`} className="text-sm text-red-600">
        {error}
      </p>
    )}
  </div>
);
```

### Pattern 3: Data Display

```tsx
// From .pen table frame analysis
export const DataTable = <T extends Record<string, unknown>>({
  columns,
  data,
}: {
  columns: { key: keyof T; label: string }[];
  data: T[];
}) => (
  <div className="overflow-x-auto rounded-md border border-neutral-200">
    <table className="min-w-full divide-y divide-neutral-200">
      <thead className="bg-neutral-50">
        <tr>
          {columns.map((col) => (
            <th
              key={String(col.key)}
              className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider"
            >
              {col.label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody className="bg-white divide-y divide-neutral-200">
        {data.map((row, i) => (
          <tr key={i}>
            {columns.map((col) => (
              <td key={String(col.key)} className="px-6 py-4 whitespace-nowrap text-sm text-neutral-900">
                {String(row[col.key])}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);
```

### Pattern 4: Responsive Layouts

```tsx
// From .pen grid frame analysis
export const Grid = ({
  children,
  cols = 3,
}: {
  children: React.ReactNode;
  cols?: 1 | 2 | 3 | 4;
}) => {
  const colMap = { 1: 'lg:grid-cols-1', 2: 'lg:grid-cols-2', 3: 'lg:grid-cols-3', 4: 'lg:grid-cols-4' };
  return (
    <div className={`grid gap-6 grid-cols-1 md:grid-cols-2 ${colMap[cols]}`}>
      {children}
    </div>
  );
};
```

## Props API Design

When generating components from .pen frames, create typed props interfaces:

```tsx
// Generated props interface from .pen component analysis
export interface CardProps {
  // Content
  children: React.ReactNode;
  title?: string;
  description?: string;

  // Styling — maps to .pen frame variants
  variant?: 'default' | 'bordered' | 'elevated';
  padding?: 'none' | 'small' | 'medium' | 'large';

  // State
  isLoading?: boolean;
  isDisabled?: boolean;

  // Events
  onClick?: () => void;

  // Accessibility
  'aria-label'?: string;
  role?: string;
}
```

## Best Practices

### Read Before Generate

Always read the .pen frame data before generating code:
1. `batch_get` to retrieve frame structure
2. `get_variables` to extract design tokens
3. `get_screenshot` to validate design intent
4. Generate code based on extracted data

### Design Token Priority

Prefer design tokens over hardcoded values:
- Use `get_variables` output as source of truth
- Map token names to CSS custom properties or Tailwind config
- Never hardcode colors, spacing, or typography values that exist in tokens

### Component Organization

```
src/
  components/
    ui/               # Generated from .pen frames
      Button.tsx
      Input.tsx
      Card.tsx
    features/         # Custom business logic components
      LoginForm.tsx   # Composes UI components
    index.ts          # Public API exports
```

### Validation

After generating code from .pen frames:
1. Use `get_screenshot` to capture the design
2. Run the application in a browser
3. Compare visual output against design screenshot
4. Iterate on styling until fidelity matches

## Error Handling

### Node Not Found

```
Error: batch_get returns empty results
Solution: Verify pattern names match exactly. Check with snapshot_layout for available nodes.
```

### Token Mismatch

```
Issue: Design tokens not applying correctly
Solution: Re-run get_variables and verify token key names.
         Check that tailwind.config.js uses exact token values from output.
```

### Layout Discrepancy

```
Issue: Generated layout differs from design
Solution: Use snapshot_layout to check computed rectangles.
         Verify flexbox direction, gap, and alignment values from batch_get output.
```

## Resources

- React Documentation: https://react.dev
- Tailwind CSS: https://tailwindcss.com
- shadcn/ui (Nova preset): https://ui.shadcn.com

---

Last Updated: 2026-03-11
Tool Version: Pencil MCP (Latest)
Workflow: Prompt-based (batch_get → analyze → generate)
