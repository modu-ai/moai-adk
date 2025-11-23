# Design System Optimization

## Token Efficiency

### CSS Variable Optimization

```typescript
// Efficient token definition with CSS variables
:root {
  /* Colors */
  --color-primary-50: #f0f4ff;
  --color-primary-500: #4f46e5;
  --color-primary-900: #1e1b4b;

  /* Spacing */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;

  /* Typography */
  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;

  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);

  /* Transitions */
  --transition-base: 250ms ease-in-out;
}

/* Light theme (default) */
:root {
  --color-bg: #ffffff;
  --color-text: #1f2937;
}

/* Dark theme */
@media (prefers-color-scheme: dark) {
  :root {
    --color-bg: #1f2937;
    --color-text: #f9fafb;
  }
}
```

### Component Token Usage

```typescript
// Efficient component styling
.button {
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--font-size-base);
  background-color: var(--color-primary-500);
  color: white;
  border-radius: 0.375rem;
  transition: background-color var(--transition-base);
}

.button:hover {
  background-color: var(--color-primary-600);
}
```

## Build Time Optimization

### Token Generation

```typescript
// Generate tokens at build time
import { tokens } from "./design-tokens"
import * as fs from "fs"

function generateCSSVariables(tokens: Record<string, any>) {
  let css = ":root {\n"

  for (const [key, value] of Object.entries(tokens)) {
    css += `  --${key}: ${value};\n`
  }

  css += "}\n"
  return css
}

// Write to file once at build time
const cssVariables = generateCSSVariables(tokens)
fs.writeFileSync("src/styles/tokens.css", cssVariables)
```

## Runtime Optimization

### Memoized Token Access

```typescript
import React from "react"

// Cache token lookups
const tokenCache = new Map()

export function getToken(path: string) {
  if (tokenCache.has(path)) {
    return tokenCache.get(path)
  }

  // Expensive token lookup
  const token = lookupToken(path)
  tokenCache.set(path, token)
  return token
}

// Usage in components
export function optimizedButton() {
  const paddingToken = getToken("spacing.md")
  return <button style={{ padding: paddingToken }} />
}
```

## Component Library Optimization

### Tree-Shaking Components

```typescript
// ✅ GOOD: Import specific components
import { Button } from "@/components/button"
import { Card } from "@/components/card"

// ❌ AVOID: Import entire component library
import * as Components from "@/components"
```

### Lazy Load Component Variants

```typescript
// Lazy load variant definitions
const variants = React.lazy(() =>
  import("./button-variants").then(mod => ({
    default: mod.buttonVariants
  }))
)
```

## CSS Optimization

### Unused Style Removal

```typescript
// Tailwind config with proper content paths
export default {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}",
    "./pages/**/*.{js,ts,jsx,tsx}",
  ],
  // Only styles for used classes will be generated
}
```

### Critical CSS

```typescript
// Extract critical CSS for above-the-fold content
@media print {
  /* Styles only for critical path */
}
```

## Performance Patterns

### Responsive Image Optimization

```typescript
export function OptimizedImage({ src, alt }: { src: string; alt: string }) {
  return (
    <picture>
      <source media="(min-width: 1024px)" srcSet={`${src}-lg.jpg`} />
      <source media="(min-width: 768px)" srcSet={`${src}-md.jpg`} />
      <img src={`${src}-sm.jpg`} alt={alt} loading="lazy" decoding="async" />
    </picture>
  )
}
```

## Scaling Optimization

### Token Generation for Scale

```typescript
// Generate design tokens for different scales
const scales = {
  compact: 0.875,
  normal: 1,
  comfortable: 1.125,
}

function generateScale(baseFactor: number) {
  return {
    spacing: {
      xs: `${0.25 * baseFactor}rem`,
      sm: `${0.5 * baseFactor}rem`,
      md: `${1 * baseFactor}rem`,
    },
  }
}

// Generate for each scale
const scales_output = Object.entries(scales).reduce(
  (acc, [name, factor]) => ({
    ...acc,
    [name]: generateScale(factor),
  }),
  {}
)
```

## Best Practices Summary

1. **Use CSS Variables** - Enable dynamic theming without runtime JS
2. **Token Generation** - Generate tokens at build time
3. **Memoize Lookups** - Cache frequently accessed tokens
4. **Tree-shake Components** - Import only needed components
5. **Responsive Variants** - Use media queries for adaptation
6. **Lazy Load Variants** - Split variant definitions
7. **Unused CSS Removal** - Configure proper content paths
8. **Image Optimization** - Use responsive images and lazy loading

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Focus**: CSS file size, token efficiency, build-time optimization
