# moai-design-tailwind-v4

Master Tailwind CSS 4 with performance engine, CSS-first configuration, and modern CSS features.

## Quick Start

Tailwind CSS 4 is a utility-first CSS framework with an optimized performance engine and support for CSS nesting and custom properties. Use this skill when styling components, creating design systems, implementing responsive layouts, or optimizing CSS output.

## Core Patterns

### Pattern 1: CSS-First Configuration with Design Tokens

**Pattern**: Define design tokens using CSS custom properties (theme variables).

```css
/* globals.css - Design tokens as CSS custom properties */
@theme {
  --color-primary: #3b82f6;
  --color-primary-dark: #1e40af;
  --color-primary-light: #dbeafe;

  --color-secondary: #ef4444;
  --color-secondary-dark: #991b1b;
  --color-secondary-light: #fee2e2;

  --color-gray: {
    50: #f9fafb;
    100: #f3f4f6;
    200: #e5e7eb;
    300: #d1d5db;
    400: #9ca3af;
    500: #6b7280;
    600: #4b5563;
    700: #374151;
    800: #1f2937;
    900: #111827;
  };

  --spacing: {
    0: 0;
    1: 0.25rem;
    2: 0.5rem;
    3: 0.75rem;
    4: 1rem;
    6: 1.5rem;
    8: 2rem;
    12: 3rem;
    16: 4rem;
    24: 6rem;
  };

  --font-family-sans: ui-sans-serif, system-ui, sans-serif;
  --font-family-mono: ui-monospace, SFMono-Regular, monospace;

  --font-size: {
    xs: 0.75rem;
    sm: 0.875rem;
    base: 1rem;
    lg: 1.125rem;
    xl: 1.25rem;
    2xl: 1.5rem;
    3xl: 1.875rem;
  };

  --border-radius: {
    none: 0;
    sm: 0.125rem;
    base: 0.25rem;
    md: 0.375rem;
    lg: 0.5rem;
    xl: 0.75rem;
    2xl: 1rem;
    full: 9999px;
  };
}

@layer base {
  :root {
    color-scheme: light;
  }

  @media (prefers-color-scheme: dark) {
    :root {
      color-scheme: dark;
      --color-primary: #60a5fa;
      --color-primary-dark: #3b82f6;
      --color-primary-light: #1e3a8a;
    }
  }
}
```

**When to use**:
- Building design systems with consistent colors, spacing, typography
- Implementing dark mode with CSS custom properties
- Creating maintainable stylesheets with reusable tokens
- Scaling design across multiple projects

**Key benefits**:
- Single source of truth for design values
- Easy theme switching without component changes
- Type-safe in TypeScript/JavaScript
- Performance optimized CSS output

### Pattern 2: Responsive Design with Mobile-First Approach

**Pattern**: Use Tailwind's responsive breakpoints for mobile-first responsive design.

```typescript
// components/ResponsiveGrid.tsx
export function ProductGrid() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {products.map((product) => (
        <article
          key={product.id}
          className="
            rounded-lg border border-gray-200 p-4
            hover:shadow-md transition-shadow
            dark:border-gray-700 dark:hover:shadow-lg
          "
        >
          <img
            src={product.image}
            alt={product.name}
            className="w-full h-48 object-cover rounded-md mb-3"
          />

          <h3 className="font-semibold text-sm md:text-base lg:text-lg truncate">
            {product.name}
          </h3>

          <p className="text-gray-500 text-xs md:text-sm line-clamp-2">
            {product.description}
          </p>

          <div className="flex items-center justify-between mt-4">
            <span className="font-bold text-lg md:text-xl text-primary">
              ${product.price}
            </span>

            <button className="
              px-3 py-1 md:px-4 md:py-2
              bg-primary text-white rounded-md
              hover:bg-primary-dark
              text-xs md:text-sm
              transition-colors
            ">
              Add to Cart
            </button>
          </div>
        </article>
      ))}
    </div>
  );
}

// Responsive typography
export function ResponsiveHeading() {
  return (
    <h1 className="
      text-2xl sm:text-3xl md:text-4xl lg:text-5xl xl:text-6xl
      font-bold leading-tight
      text-gray-900 dark:text-white
    ">
      Welcome to Our Store
    </h1>
  );
}
```

**When to use**:
- Building responsive web applications
- Supporting mobile, tablet, and desktop screens
- Implementing flexible layouts
- Creating accessible text sizing

**Key benefits**:
- Mobile-first approach reduces CSS
- Breakpoints align with real device sizes
- Consistent spacing and sizing
- Easy to test on actual devices

### Pattern 3: Advanced CSS Features (Nesting, Gradients, Animations)

**Pattern**: Use CSS nesting, gradients, and custom animations for advanced styling.

```css
/* components/Card.module.css - CSS nesting */
.card {
  @apply rounded-lg border border-gray-200 p-6 bg-white;

  &:hover {
    @apply shadow-lg border-primary;
  }

  &.featured {
    @apply border-2 border-primary bg-primary/5;
  }

  .header {
    @apply flex items-center justify-between mb-4;

    .title {
      @apply text-lg font-bold;
    }

    .icon {
      @apply w-6 h-6 text-primary;
    }
  }

  .content {
    @apply text-gray-600 leading-relaxed;

    &.dense {
      @apply text-sm;
    }
  }

  .footer {
    @apply flex gap-2 mt-4 pt-4 border-t;

    .button {
      @apply px-3 py-1 rounded text-sm font-medium;

      &.primary {
        @apply bg-primary text-white hover:bg-primary-dark;
      }

      &.secondary {
        @apply bg-gray-100 text-gray-900 hover:bg-gray-200;
      }
    }
  }
}

/* Gradient backgrounds */
.gradient-primary {
  @apply bg-gradient-to-r from-primary via-primary-light to-primary-dark;
}

.gradient-text {
  @apply bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent;
}

/* Custom animations */
@keyframes shimmer {
  0% {
    background-position: -1000px 0;
  }
  100% {
    background-position: 1000px 0;
  }
}

.loading {
  background: linear-gradient(
    90deg,
    #f0f0f0 25%,
    #e0e0e0 50%,
    #f0f0f0 75%
  );
  background-size: 1000px 100%;
  animation: shimmer 2s infinite;
}
```

**When to use**:
- Creating complex UI patterns
- Building loading states and skeletons
- Implementing smooth transitions and animations
- Creating visually appealing designs

**Key benefits**:
- CSS nesting reduces specificity issues
- Native CSS gradients are performant
- Keyframe animations for smooth interactions
- Composable utility classes

## Progressive Disclosure

### Level 1: Basic Utilities
- Using utility classes for styling
- Responsive design with breakpoints (sm, md, lg, xl)
- Color palette and spacing system
- Hover and focus states

### Level 2: Advanced Configuration
- Customizing theme with design tokens
- Creating component classes
- Dark mode support
- Custom breakpoints and theme extensions

### Level 3: Expert Optimization
- CSS nesting for complex selectors
- Performance optimization (PurgeCSS, tree-shaking)
- Advanced animation and transition patterns
- Plugin development for custom utilities

## Works Well With

- **Next.js 16**: Official styling approach for Next.js
- **React 19**: Component-scoped styling with Tailwind
- **shadcn/ui**: Built entirely on Tailwind CSS
- **TypeScript**: Type-safe with tailwind-merge and clsx
- **Headless UI**: Unstyled components + Tailwind for styling
- **PostCSS**: CSS transformation pipeline

## References

- **Official Documentation**: https://tailwindcss.com
- **Tailwind CSS v4**: https://tailwindcss.com/blog/tailwindcss-v4
- **Configuration**: https://tailwindcss.com/docs/configuration
- **Responsive Design**: https://tailwindcss.com/docs/responsive-design
- **Customization**: https://tailwindcss.com/docs/customizing-your-theme
- **Dark Mode**: https://tailwindcss.com/docs/dark-mode
