---
name: moai-lang-tailwind-css
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Enterprise Skill for advanced development
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# moai-lang-tailwind-css

**Tailwind CSS: Utility-First Framework for Rapid UI Development**

> **Primary Agent**: frontend-expert
> **Secondary Agent**: ui-ux-expert
> **Version**: 1.0.0 (Tailwind CSS  +)
> **Keywords**: tailwind, tailwind css, utility-first, css framework, responsive, customization, design tokens, performance

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Tailwind CSS** is a utility-first CSS framework that enables building custom designs directly in HTML:

| Concept | Description | Example |
|---------|-------------|---------|
| **Utility Classes** | Single-purpose classes for styling | `p-4`, `text-lg`, `bg-blue-500` |
| **Responsive** | Mobile-first with breakpoints | `md:text-lg`, `lg:p-8` |
| **Design Tokens** | Predefined scales (colors, spacing) | Colors: 50-950 scale |
| **Dark Mode** | Toggle dark/light themes | `dark:bg-gray-900` |
| **Customization** | Extend theme in config | `tailwind.config.js` |
| **Performance** | Purges unused styles at build | Only shipped classes in CSS |

**Installation ( +)**:
```bash
npm install -D tailwindcss

# Create config file
npx tailwindcss init
```

**Basic Setup**:
```css
/* app.css */
@import "tailwindcss";
```

**Key Breakpoints** (mobile-first):
| Prefix | Size | Min-Width |
|--------|------|-----------|
| (none) | Default | 0px |
| `sm` | Small | 640px |
| `md` | Medium | 768px |
| `lg` | Large | 1024px |
| `xl` | Extra Large | 1280px |
| `2xl` | 2X Large | 1536px |

---

### Level 2: Practical Implementation (Common Patterns)

#### Pattern 1: Basic Tailwind Setup with Design Tokens

```css
/* app.css */
@import "tailwindcss";

@theme {
  /* Colors - semantic naming */
  --color-primary: #0ea5e9;
  --color-primary-dark: #0284c7;
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;

  --color-neutral-50: #f9fafb;
  --color-neutral-100: #f3f4f6;
  --color-neutral-900: #111827;

  /* Typography */
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-mono: "Monaco", "Menlo", "Ubuntu Mono", monospace;

  /* Spacing - 8px base unit */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  --spacing-2xl: 3rem;

  /* Breakpoints */
  --breakpoint-sm: 640px;
  --breakpoint-md: 768px;
  --breakpoint-lg: 1024px;
  --breakpoint-xl: 1280px;
}
```

#### Pattern 2: Responsive Component with Tailwind

```jsx
// React component using Tailwind
function Card({ title, description, image }) {
  return (
    <div className="rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow">
      {/* Image - responsive sizing */}
      <img
        src={image}
        alt={title}
        className="w-full h-48 md:h-64 object-cover"
      />

      {/* Content - responsive padding */}
      <div className="p-4 md:p-6">
        <h3 className="text-lg md:text-xl font-bold text-gray-900 mb-2">
          {title}
        </h3>

        <p className="text-gray-600 text-sm md:text-base mb-4">
          {description}
        </p>

        {/* Button */}
        <button className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500 transition-colors">
          Learn More
        </button>
      </div>
    </div>
  );
}
```

#### Pattern 3: Tailwind Configuration with Custom Theme

```javascript
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
    "./index.html",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#f0f9ff",
          100: "#e0f2fe",
          500: "#0ea5e9",
          600: "#0284c7",
          900: "#0c2d4a",
        },
        semantic: {
          success: "#10b981",
          warning: "#f59e0b",
          error: "#ef4444",
        },
      },
      spacing: {
        xs: "0.25rem",
        sm: "0.5rem",
        md: "1rem",
        lg: "1.5rem",
        xl: "2rem",
        "2xl": "3rem",
      },
      fontFamily: {
        sans: [
          "-apple-system",
          "BlinkMacSystemFont",
          '"Segoe UI"',
          "sans-serif",
        ],
        mono: ["Monaco", "Menlo", "monospace"],
      },
      fontSize: {
        xs: ["0.75rem", { lineHeight: "1rem" }],
        sm: ["0.875rem", { lineHeight: "1.25rem" }],
        base: ["1rem", { lineHeight: "1.5rem" }],
        lg: ["1.125rem", { lineHeight: "1.75rem" }],
        xl: ["1.25rem", { lineHeight: "1.75rem" }],
      },
      borderRadius: {
        sm: "0.25rem",
        md: "0.375rem",
        lg: "0.5rem",
        xl: "0.75rem",
      },
      boxShadow: {
        sm: "0 1px 2px 0 rgba(0, 0, 0, 0.05)",
        md: "0 4px 6px 0 rgba(0, 0, 0, 0.1)",
        lg: "0 10px 15px 0 rgba(0, 0, 0, 0.1)",
      },
    },
  },
  plugins: [
    require("@tailwindcss/forms"),
    require("@tailwindcss/typography"),
  ],
};
```

#### Pattern 4: Dark Mode Toggle

```jsx
function DarkModeToggle() {
  const [isDark, setIsDark] = React.useState(false);

  return (
    <div className={isDark ? "dark" : ""}>
      <button
        onClick={() => setIsDark(!isDark)}
        className="px-4 py-2 bg-gray-200 dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg"
      >
        {isDark ? "🌙" : "☀️"}
      </button>

      {/* Component that respects dark mode */}
      <div className="bg-white dark:bg-gray-900 text-gray-900 dark:text-white p-6">
        <h1 className="text-2xl font-bold">Dark Mode Example</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          This text changes color in dark mode.
        </p>
      </div>
    </div>
  );
}
```

#### Pattern 5: Responsive Grid Layout

```jsx
// Responsive grid: 1 column mobile, 2 columns tablet, 3 columns desktop
function ProductGrid({ products }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
      {products.map((product) => (
        <div
          key={product.id}
          className="rounded-lg border border-gray-200 overflow-hidden hover:shadow-lg transition-shadow"
        >
          <img
            src={product.image}
            alt={product.name}
            className="w-full h-40 object-cover"
          />
          <div className="p-4">
            <h3 className="font-bold text-gray-900">{product.name}</h3>
            <p className="text-gray-600 text-sm mt-1">{product.description}</p>
            <div className="mt-4 flex justify-between items-center">
              <span className="font-bold text-primary-600">${product.price}</span>
              <button className="px-3 py-1 bg-primary-500 text-white rounded-md hover:bg-primary-600 text-sm">
                Add to Cart
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
```

---

### Level 3: Advanced Patterns (Expert Reference)

#### Advanced Pattern 1: Custom Variants and Plugins

```javascript
// tailwind.config.js - Custom variant plugin
const plugin = require("tailwindcss/plugin");

module.exports = {
  theme: {
    extend: {},
  },
  plugins: [
    // Custom variant: group-hover-children
    plugin(function ({ addVariant }) {
      addVariant("group-hover-children", ":where(.group:hover) & > *");
    }),

    // Custom utility: safe area
    plugin(function ({ matchUtilities, theme }) {
      matchUtilities(
        {
          safe: (value) => ({
            paddingLeft: `max(${value}, env(safe-area-inset-left))`,
            paddingRight: `max(${value}, env(safe-area-inset-right))`,
          }),
        },
        {
          values: theme("spacing"),
        }
      );
    }),
  ],
};
```

Usage in HTML:
```html
<!-- Custom variant -->
<div class="group">
  <div class="group-hover-children:text-blue-500">Text changes on parent hover</div>
</div>

<!-- Custom utility -->
<div class="safe-4">Safe area padding</div>
```

#### Advanced Pattern 2: Dynamic Color System with CSS Variables

```javascript
// Create dynamic color system
const generateColorShades = (hex) => {
  // Convert hex to HSL, generate shades
  const shades = {
    50: lighten(hex, 0.95),
    100: lighten(hex, 0.9),
    500: hex,
    900: darken(hex, 0.9),
  };
  return shades;
};

module.exports = {
  theme: {
    colors: {
      primary: generateColorShades("#0ea5e9"),
      secondary: generateColorShades("#8b5cf6"),
    },
  },
};
```

#### Advanced Pattern 3: Container Queries (Modern Approach)

```css
/* app.css */
@import "tailwindcss";

@theme {
  --container-sm: 24rem;
  --container-md: 28rem;
  --container-lg: 32rem;
  --container-xl: 36rem;
}
```

```jsx
// Usage
function ResponsiveCard() {
  return (
    <div className="@container">
      {/* Responsive to container, not viewport */}
      <div className="@sm:grid @sm:grid-cols-2 @md:grid-cols-3">
        {/* Content */}
      </div>
    </div>
  );
}
```

---

## 🎯 Performance Best Practices

### 1. Purging Unused CSS
```javascript
// tailwind.config.js
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
    // Include all files that may contain Tailwind classes
    "./components/**/*.{js,jsx}",
    "./pages/**/*.{js,jsx}",
  ],
};
```

### 2. Avoid Dynamic Class Names
```javascript
// ❌ BAD - Tailwind can't scan this
const padding = "p-4"; // Dynamic string
className={`p-${size}`}

// ✅ GOOD - Static class names
className={size === "sm" ? "p-2" : "p-4"}
```

### 3. Extract Reusable Components
```css
/* Instead of repeating classes, use @layer components */
@layer components {
  .btn-primary {
    @apply px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors;
  }

  .card {
    @apply rounded-lg shadow-md p-6 bg-white;
  }
}
```

### 4. Tree-Shaking and Code Splitting
```javascript
// Webpack config - enable tree-shaking
mode: "production", // Automatically enables minification
optimization: {
  usedExports: true,
}
```

---

## 📚 Official References

- **Tailwind CSS Docs**: https://tailwindcss.com/docs
- **Tailwind UI**: https://tailwindui.com/ (Component examples)
- **Tailwind IntelliSense**: VS Code extension for autocomplete
- **Tailwind CSS Plugins**: https://tailwindcss.com/docs/plugins
- **Headless UI**: https://headlessui.com/ (Unstyled, accessible components)

---

## 🔗 Related Skills

- `Skill("moai-lang-html-css")` – Semantic HTML & CSS foundations
- `Skill("moai-lib-shadcn-ui")` – Pre-built Tailwind components
- `Skill("moai-domain-frontend")` – Full frontend architecture with Tailwind
