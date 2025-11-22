# Advanced Design System Patterns

## Design System Architecture

### Token-Based Design System

A modern design system is built on design tokens - single-source-of-truth values for design decisions.

**Pattern: Token Definition and Usage**

```typescript
// tokens/colors.ts
export const colors = {
  primary: {
    50: "#f0f4ff",
    100: "#e6ebff",
    500: "#4f46e5",
    600: "#4338ca",
    900: "#1e1b4b",
  },
  semantic: {
    success: "var(--color-green-500)",
    error: "var(--color-red-500)",
    warning: "var(--color-yellow-500)",
  },
}

// Usage in components
export function Alert({ type }: { type: "success" | "error" | "warning" }) {
  return (
    <div style={{
      backgroundColor: colors.semantic[type],
      padding: "1rem",
    }}>
      Alert content
    </div>
  )
}
```

### Component Token System

```typescript
// tokens/spacing.ts
export const spacing = {
  xs: "0.25rem",   // 4px
  sm: "0.5rem",    // 8px
  md: "1rem",      // 16px
  lg: "1.5rem",    // 24px
  xl: "2rem",      // 32px
}

// tokens/typography.ts
export const typography = {
  heading1: {
    fontSize: "2.5rem",
    lineHeight: "1.2",
    fontWeight: 700,
  },
  body: {
    fontSize: "1rem",
    lineHeight: "1.5",
    fontWeight: 400,
  },
  caption: {
    fontSize: "0.875rem",
    lineHeight: "1.25",
    fontWeight: 400,
  },
}
```

## Component Composition Patterns

### Slot-Based Component Architecture

```typescript
// Button component with flexible slots
interface ButtonProps {
  variant?: "primary" | "secondary"
  size?: "sm" | "md" | "lg"
  leftIcon?: React.ReactNode
  rightIcon?: React.ReactNode
  children: React.ReactNode
}

export function Button({
  variant = "primary",
  size = "md",
  leftIcon,
  rightIcon,
  children,
}: ButtonProps) {
  const variantClasses = {
    primary: "bg-blue-500 text-white",
    secondary: "bg-gray-200 text-gray-900",
  }

  const sizeClasses = {
    sm: "px-2 py-1 text-sm",
    md: "px-4 py-2",
    lg: "px-6 py-3 text-lg",
  }

  return (
    <button className={`${variantClasses[variant]} ${sizeClasses[size]} flex items-center gap-2`}>
      {leftIcon && <span>{leftIcon}</span>}
      {children}
      {rightIcon && <span>{rightIcon}</span>}
    </button>
  )
}
```

### Polymorphic Components

```typescript
// Generic component that accepts any element type
interface PolymorphicProps<T extends React.ElementType> {
  as?: T
  children: React.ReactNode
}

export function Card<T extends React.ElementType = "div">({
  as: Component = "div",
  children,
  ...props
}: PolymorphicProps<T> & React.ComponentPropsWithoutRef<T>) {
  return (
    <Component className="rounded-lg border shadow" {...props}>
      {children}
    </Component>
  )
}

// Usage
<Card as="section">Section Card</Card>
<Card as="article">Article Card</Card>
```

## Design Variant Patterns

### Responsive Variants

```typescript
// Design system with responsive support
const responsiveDesign = {
  container: {
    default: "w-full",
    md: "md:max-w-2xl",
    lg: "lg:max-w-4xl",
  },
  grid: {
    cols: {
      default: "grid-cols-1",
      md: "md:grid-cols-2",
      lg: "lg:grid-cols-3",
    },
  },
}

export function ResponsiveGrid() {
  return (
    <div className={`${responsiveDesign.grid.cols.default} ${responsiveDesign.grid.cols.md} ${responsiveDesign.grid.cols.lg}`}>
      {/* Grid content */}
    </div>
  )
}
```

### State Variants

```typescript
interface StateVariants {
  default: string
  hover: string
  active: string
  disabled: string
  focus: string
}

const buttonStates: StateVariants = {
  default: "bg-blue-500 text-white",
  hover: "hover:bg-blue-600",
  active: "active:bg-blue-700",
  disabled: "disabled:opacity-50 disabled:cursor-not-allowed",
  focus: "focus:ring-2 focus:ring-blue-300",
}
```

## Color System

### Semantic Color Palette

```typescript
export const semanticColors = {
  success: {
    light: "#d4edda",
    main: "#28a745",
    dark: "#1e7e34",
  },
  error: {
    light: "#f8d7da",
    main: "#dc3545",
    dark: "#bd2130",
  },
  warning: {
    light: "#fff3cd",
    main: "#ffc107",
    dark: "#e0a800",
  },
  info: {
    light: "#d1ecf1",
    main: "#17a2b8",
    dark: "#0c5460",
  },
}
```

## Typography System

### Font Scale

```typescript
export const fontSizes = {
  xs: "0.75rem",     // 12px
  sm: "0.875rem",    // 14px
  base: "1rem",      // 16px
  lg: "1.125rem",    // 18px
  xl: "1.25rem",     // 20px
  "2xl": "1.5rem",   // 24px
  "3xl": "1.875rem", // 30px
  "4xl": "2.25rem",  // 36px
  "5xl": "3rem",     // 48px
}

export const lineHeights = {
  tight: 1.1,
  snug: 1.375,
  normal: 1.5,
  relaxed: 1.625,
  loose: 2,
}
```

## Spacing System

### Modular Scale

```typescript
export const baseSpacing = 8 // 8px base unit

export const spacing = {
  0: "0",
  xs: `${baseSpacing * 0.25}px`,      // 2px
  sm: `${baseSpacing * 0.5}px`,       // 4px
  md: `${baseSpacing * 1}px`,         // 8px
  lg: `${baseSpacing * 1.5}px`,       // 12px
  xl: `${baseSpacing * 2}px`,         // 16px
  "2xl": `${baseSpacing * 2.5}px`,    // 20px
  "3xl": `${baseSpacing * 3}px`,      // 24px
  "4xl": `${baseSpacing * 4}px`,      // 32px
}
```

## Border Radius System

```typescript
export const borderRadius = {
  none: "0",
  sm: "0.125rem",    // 2px
  base: "0.25rem",   // 4px
  md: "0.375rem",    // 6px
  lg: "0.5rem",      // 8px
  xl: "0.75rem",     // 12px
  "2xl": "1rem",     // 16px
  full: "9999px",
}
```

## Shadow System

```typescript
export const shadows = {
  none: "none",
  sm: "0 1px 2px 0 rgba(0, 0, 0, 0.05)",
  md: "0 4px 6px -1px rgba(0, 0, 0, 0.1)",
  lg: "0 10px 15px -3px rgba(0, 0, 0, 0.1)",
  xl: "0 20px 25px -5px rgba(0, 0, 0, 0.1)",
}
```

## Transition System

```typescript
export const transitions = {
  fast: "150ms ease-in-out",
  base: "250ms ease-in-out",
  slow: "350ms ease-in-out",
}

export function TransitionExample() {
  return (
    <div
      style={{
        transition: `all ${transitions.base}`,
      }}
      className="hover:bg-blue-500"
    >
      Hover me
    </div>
  )
}
```

## Zindex Scale

```typescript
export const zIndex = {
  hide: -1,
  auto: "auto",
  base: 0,
  dropdown: 1000,
  sticky: 1020,
  fixed: 1030,
  backdrop: 1040,
  offcanvas: 1050,
  modal: 1060,
  popover: 1070,
  tooltip: 1080,
}
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
