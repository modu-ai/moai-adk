# moai-lang-tailwind-shadcn

Integration patterns for Tailwind CSS and shadcn/ui theming, component customization, dark mode.

## Quick Start

Tailwind CSS and shadcn/ui work together seamlessly to provide a powerful styling system. Use this skill when creating themed component libraries, implementing design systems, or customizing shadcn/ui components for your brand.

## Core Patterns

### Pattern 1: Theme Configuration & Design Tokens

**Pattern**: Configure Tailwind CSS design tokens for consistent theming across shadcn/ui components.

```typescript
// tailwind.config.ts - Tailwind + shadcn/ui theme
import type { Config } from "tailwindcss"
import defaultTheme from "tailwindcss/defaultTheme"

const config = {
  darkMode: ["class"],
  content: [
    "./pages/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./app/**/*.{ts,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      fontFamily: {
        sans: ["var(--font-sans)", ...defaultTheme.fontFamily.sans],
        mono: ["var(--font-mono)", ...defaultTheme.fontFamily.mono],
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
} satisfies Config

export default config
```

```css
/* app/globals.css - CSS variables for theming */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --muted: 221.2 63.3% 97.8%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 221.2 83.2% 53.3%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --primary: 221.2 83.2% 53.3%;
    --primary-foreground: 210 40% 98%;
    --secondary: 217.2 91.2% 59.8%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --ring: 221.2 83.2% 53.3%;
    --radius: 0.5rem;

    --font-sans: system-ui, sans-serif;
    --font-mono: "Monaco", monospace;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 91.2% 59.8%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --primary: 217.2 91.2% 59.8%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 221.2 83.2% 53.3%;
    --secondary-foreground: 210 40% 98%;
    --ring: 212.7 26.8% 83.9%;
  }

  * {
    @apply border-border;
  }

  body {
    @apply bg-background text-foreground;
  }
}
```

**When to use**:
- Setting up color schemes for your application
- Creating dark mode support
- Brand customization with color tokens
- Multi-theme support

**Key benefits**:
- CSS variables enable runtime theme switching
- Single source of truth for colors
- Easy dark mode implementation
- Type-safe theme access in TypeScript

### Pattern 2: Custom Component Variants with shadcn/ui

**Pattern**: Create custom component variants by extending shadcn/ui base components.

```typescript
// components/ui/custom-button.tsx - Custom button variant
import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const customButtonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
        // Custom brand variants
        gradient: "bg-gradient-to-r from-primary to-accent text-white hover:opacity-90",
        subtle: "bg-muted text-muted-foreground hover:bg-muted/80",
        elevated: "bg-white shadow-md hover:shadow-lg dark:bg-slate-800",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
        xl: "h-12 rounded-lg px-10 text-lg",
        icon: "h-10 w-10",
        icon-sm: "h-8 w-8",
      },
      fullWidth: {
        true: "w-full",
        false: "",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
      fullWidth: false,
    },
  }
)

export interface CustomButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof customButtonVariants> {
  asChild?: boolean
}

const CustomButton = React.forwardRef<HTMLButtonElement, CustomButtonProps>(
  ({ className, variant, size, fullWidth, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(customButtonVariants({ variant, size, fullWidth, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
CustomButton.displayName = "CustomButton"

export { CustomButton, customButtonVariants }

// Usage example
export function ButtonShowcase() {
  return (
    <div className="space-y-4">
      <CustomButton variant="gradient">Gradient Button</CustomButton>
      <CustomButton variant="subtle" size="lg">Subtle Large</CustomButton>
      <CustomButton variant="elevated" fullWidth>Full Width</CustomButton>
      <CustomButton variant="link">Link Button</CustomButton>
    </div>
  )
}
```

**When to use**:
- Creating brand-specific component variants
- Building design system components
- Extending shadcn/ui for specific needs
- Maintaining consistency across components

**Key benefits**:
- Reusable component variants
- Type-safe prop combinations
- Maintainable custom components
- Easy to scale across projects

### Pattern 3: Responsive Component Patterns

**Pattern**: Build responsive components that adapt to different screen sizes.

```typescript
// components/responsive-layout.tsx - Responsive grid layout
'use client';

import { useMediaQuery } from '@/hooks/use-media-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export function ResponsiveLayout() {
  const isTablet = useMediaQuery('(min-width: 768px)');
  const isDesktop = useMediaQuery('(min-width: 1024px)');

  const columns = isDesktop ? 4 : isTablet ? 2 : 1;

  return (
    <div className="space-y-6">
      <div className={`grid gap-4 grid-cols-${columns}`}>
        {/* Dynamic columns based on screen size */}
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle className="text-sm md:text-base lg:text-lg">Card {i}</CardTitle>
              <CardDescription>Description text</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-xs md:text-sm text-muted-foreground">
                Content adapts to screen size
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Responsive navigation */}
      <nav className="flex flex-col md:flex-row gap-2">
        <Button variant="default" fullWidth={!isTablet}>
          Primary
        </Button>
        <Button variant="outline" fullWidth={!isTablet}>
          Secondary
        </Button>
      </nav>

      {/* Mobile-only elements */}
      {!isTablet && (
        <div className="p-4 bg-muted rounded-lg">
          <p className="text-sm">Mobile-only message</p>
        </div>
      )}
    </div>
  );
}

// hooks/use-media-query.ts - Custom hook for media queries
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = React.useState(false);

  React.useEffect(() => {
    const media = window.matchMedia(query);
    if (media.matches !== matches) {
      setMatches(media.matches);
    }

    const listener = () => setMatches(media.matches);
    media.addEventListener('change', listener);

    return () => media.removeEventListener('change', listener);
  }, [matches, query]);

  return matches;
}
```

**When to use**:
- Building mobile-responsive UIs
- Supporting multiple device types
- Creating flexible layouts
- Conditional rendering based on screen size

**Key benefits**:
- Mobile-first approach
- Smooth responsive transitions
- Better user experience across devices
- Single codebase for all screen sizes

## Progressive Disclosure

### Level 1: Basic Integration
- Use shadcn/ui components with Tailwind
- Apply basic Tailwind utilities to components
- Understand color token system
- Basic dark mode setup

### Level 2: Advanced Customization
- Create custom component variants
- Extend theme configuration
- Build responsive layouts
- Implement theme switching

### Level 3: Expert Design Systems
- Create comprehensive component libraries
- Design token management and extraction
- Multi-brand theming support
- Advanced animation and transition patterns

## Works Well With

- **Tailwind CSS 4**: Core styling framework
- **shadcn/ui**: Component library built on Tailwind
- **CVA**: Type-safe variant management
- **Next.js 16**: Server components and styles
- **TypeScript**: Type-safe theme access
- **Radix UI**: Accessible component primitives

## References

- **Tailwind Configuration**: https://tailwindcss.com/docs/configuration
- **shadcn/ui Docs**: https://ui.shadcn.com
- **Dark Mode**: https://tailwindcss.com/docs/dark-mode
- **Responsive Design**: https://tailwindcss.com/docs/responsive-design
- **CSS Variables**: https://developer.mozilla.org/en-US/docs/Web/CSS/--*
