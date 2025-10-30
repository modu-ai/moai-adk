---
title: Tailwind CSS 4 Features & Optimization
description: Master Tailwind CSS 4 performance engine, CSS-first configuration, and modern CSS features
freedom_level: high
tier: domain
updated: 2025-10-31
---

# Tailwind CSS 4 Features & Optimization

## Overview

Tailwind CSS v4 is a complete reimagining of the framework with a new high-performance engine (5x faster full builds, 100x faster incremental builds), CSS-first configuration, and modern CSS features like cascade layers, @property, and color-mix(). This skill covers v4's breaking changes, migration patterns, and optimization strategies.

## Key Patterns

### 1. CSS-First Configuration (Breaking Change)

**Pattern**: Configure Tailwind directly in CSS instead of tailwind.config.js.

```css
/* app/globals.css */
@import "tailwindcss";

/* Define custom theme in CSS */
@theme {
  /* Custom colors */
  --color-brand: #0066cc;
  --color-brand-dark: #004080;
  
  /* Custom spacing */
  --spacing-xl: 3rem;
  --spacing-xxl: 5rem;
  
  /* Custom breakpoints */
  --breakpoint-xs: 375px;
  --breakpoint-3xl: 1920px;
  
  /* Custom fonts */
  --font-display: 'Inter Display', sans-serif;
  --font-mono: 'Fira Code', monospace;
}

/* Use custom tokens */
.hero {
  background-color: var(--color-brand);
  padding: var(--spacing-xxl);
  font-family: var(--font-display);
}
```

**Migration from v3**:
```javascript
// ❌ OLD: tailwind.config.js (v3)
module.exports = {
  theme: {
    extend: {
      colors: {
        brand: '#0066cc'
      }
    }
  }
}

// ✅ NEW: CSS @theme block (v4)
@theme {
  --color-brand: #0066cc;
}
```

### 2. Automatic Content Detection

**Pattern**: No more content configuration - Tailwind v4 automatically detects template files.

```css
/* v4: No configuration needed! */
@import "tailwindcss";

/* Tailwind automatically scans:
 * - app/**/*.{js,ts,jsx,tsx}
 * - components/**/*.{js,ts,jsx,tsx}
 * - src/**/*.{js,ts,jsx,tsx}
 */
```

**Migration**:
```javascript
// ❌ OLD: Manual content paths (v3)
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}'
  ]
}

// ✅ NEW: Automatic detection (v4)
// Nothing to configure!
```

### 3. P3 Wide-Gamut Colors

**Pattern**: Use vivid P3 colors for modern displays.

```css
@theme {
  /* P3 colors for modern screens */
  --color-electric-blue: color(display-p3 0 0.5 1);
  --color-vivid-red: color(display-p3 1 0 0.3);
  --color-neon-green: color(display-p3 0.2 1 0.4);
}

/* Automatic fallback for older browsers */
.button-primary {
  /* Browsers without P3 support get sRGB fallback */
  background-color: var(--color-electric-blue);
}
```

**Utility classes**:
```html
<!-- v4 includes P3 variants of default colors -->
<div class="bg-blue-500">Standard blue</div>
<div class="bg-blue-500/p3">Vivid P3 blue (if supported)</div>
```

### 4. Container Queries (Built-in)

**Pattern**: Use @container for component-based responsive design.

```css
/* Define container */
.card-wrapper {
  container-type: inline-size;
  container-name: card;
}

/* Container query utilities */
@container card (min-width: 400px) {
  .card-content {
    display: grid;
    grid-template-columns: 1fr 2fr;
  }
}
```

**Utility classes** (v4 built-in, no plugin needed):
```html
<div class="@container">
  <div class="@lg:grid @lg:grid-cols-2">
    <div>Sidebar</div>
    <div>Content</div>
  </div>
</div>
```

**Migration**: Remove `@tailwindcss/container-queries` plugin - now built-in!

### 5. not-* Variant for Negation

**Pattern**: Style elements when they DON'T match a condition.

```html
<!-- Style when NOT hovering -->
<button class="not-hover:opacity-50 hover:opacity-100">
  Hover me
</button>

<!-- Style when NOT disabled -->
<input class="not-disabled:cursor-pointer disabled:cursor-not-allowed" />

<!-- Style when NOT first child -->
<div class="not-first:border-t">
  <p class="not-first:mt-4">Item 1</p>
  <p class="not-first:mt-4">Item 2 (has top border + margin)</p>
</div>

<!-- Style when NOT in dark mode -->
<div class="not-dark:bg-white dark:bg-black">
  Adapts to theme
</div>
```

**Use Cases**: Simplify conditional styling, reduce custom CSS, cleaner component code.

### 6. Dynamic Utility Values (Arbitrary Values Enhanced)

**Pattern**: Use dynamic values without extending configuration.

```html
<!-- v4: Arbitrary values work everywhere -->
<div class="grid-cols-[200px_1fr_200px]">
  Three column layout
</div>

<div class="bg-[#c0ffee]">
  Custom hex color
</div>

<div class="text-[clamp(1rem,2vw,2rem)]">
  Responsive text with clamp
</div>

<!-- Dynamic CSS variables -->
<div style="--card-height: 400px" class="h-[var(--card-height)]">
  Dynamic height from CSS variable
</div>
```

### 7. Performance Optimization Strategy

**Pattern**: Leverage v4's speed for optimal build times.

```json
// package.json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "build:css": "tailwindcss -i ./app/globals.css -o ./dist/output.css --minify"
  }
}
```

**Benchmark Results** (from official v4 release):
- Full builds: **5x faster** than v3
- Incremental builds: **100x faster** (measured in microseconds)
- Real project: **3.5x faster** full rebuilds, **8x faster** incremental

**Optimization Tips**:
```css
/* Use @layer for better organization */
@layer components {
  .btn-primary {
    @apply px-4 py-2 bg-blue-500 text-white rounded;
  }
}

@layer utilities {
  .text-balance {
    text-wrap: balance;
  }
}
```

### 8. Vite Plugin Integration

**Pattern**: Use official Vite plugin for tight integration.

```javascript
// vite.config.js
import tailwindcss from '@tailwindcss/vite'

export default {
  plugins: [
    tailwindcss()
  ]
}
```

**Benefits**: Zero configuration, instant HMR, optimized production builds.

### 9. Cascade Layers for Specificity Control

**Pattern**: v4 uses CSS cascade layers for predictable specificity.

```css
/* Tailwind v4 automatically organizes into layers */
@layer base, components, utilities;

/* Your custom layers */
@layer components {
  .card {
    @apply rounded-lg shadow-md p-4;
  }
}

@layer utilities {
  .text-pretty {
    text-wrap: pretty;
  }
}

/* Override behavior */
@layer overrides {
  /* These styles take precedence over Tailwind utilities */
  .force-blue {
    color: blue !important;
  }
}
```

### 10. Registered Custom Properties (@property)

**Pattern**: v4 uses @property for type-safe CSS variables.

```css
@property --color-primary {
  syntax: '<color>';
  inherits: true;
  initial-value: #3b82f6;
}

@property --spacing-unit {
  syntax: '<length>';
  inherits: true;
  initial-value: 1rem;
}

/* Animate custom properties smoothly */
.animated-gradient {
  background: linear-gradient(var(--color-primary), var(--color-secondary));
  transition: --color-primary 0.3s ease;
}
```

## Checklist

- [ ] Migrate `tailwind.config.js` to CSS `@theme` blocks
- [ ] Remove `content` configuration (automatic detection in v4)
- [ ] Remove `@tailwindcss/container-queries` plugin (built-in now)
- [ ] Update build scripts to use Tailwind v4 CLI
- [ ] Install Vite plugin: `npm install -D @tailwindcss/vite`
- [ ] Replace `!important` with cascade layers where possible
- [ ] Test P3 colors on modern displays (Safari, Chrome on macOS)
- [ ] Use `not-*` variants to simplify negative conditions
- [ ] Benchmark build times: expect 3-5x speedup
- [ ] Review arbitrary values: ensure they work in all contexts

## Resources

- **Official v4 Release**: https://tailwindcss.com/blog/tailwindcss-v4
- **Migration Guide**: https://tailwindcss.com/docs/upgrade-guide
- **CSS-First Configuration**: https://tailwindcss.com/docs/configuration
- **Container Queries**: https://tailwindcss.com/docs/hover-focus-and-other-states#container-queries
- **Best Practices (2025)**: https://www.bootstrapdash.com/blog/tailwind-css-best-practices
- **Performance Analysis**: https://medium.com/@asierr/tailwind-css-v4-whats-new-and-why-it-matters-for-developers-in-2025-5df81fd2b8b5

---

**Version**: 1.0.0  
**Last Updated**: 2025-10-31  
**Model Recommendation**: Sonnet (deep reasoning for migration strategy and performance optimization)
