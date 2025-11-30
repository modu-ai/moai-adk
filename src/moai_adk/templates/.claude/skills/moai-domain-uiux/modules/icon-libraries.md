---
name: moai-foundation-uiux
description: Vector icon libraries ecosystem guide covering 10+ major libraries with 200K+ icons, including React Icons (35K+), Lucide (1000+), Tabler Icons (5900+), Iconify (200K+), Heroicons, Phosphor, and Radix Icons with implementation patterns, decision trees, and best practices.
version: 1.0.0
modularized: false
tags:
 - enterprise
 - development
 - vector
updated: 2025-11-24
status: active
---

## Quick Reference (30 seconds)

# Icon Libraries

Vector Icon Libraries: Enterprise Guide (10+ Libraries, 200K+ Icons)

> Primary Agent: frontend-expert
> Secondary Agent: ui-ux-expert
> Version: 4.0.0 (Lucide v0.4+, React Icons 35K+, Tabler v2.0+, Phosphor v1.4+)
> Keywords: icons, vector icons, lucide, react icons, iconify, svg icons, accessibility

## Level 1: Quick Reference

### Library Selection Guide

Ecosystem Leaders (1000+ icons):
- Lucide (1000+): General UI, modern design, ~30KB
- React Icons (35K+): Multi-library support, modular bundles
- Tabler Icons (5900+): Dashboard optimized, ~22KB
- Ionicons (1300+): Mobile + web compatibility

Specialist Libraries (300-800 icons):
- Heroicons (300+): Official Tailwind CSS icons
- Phosphor (800+): 6 weights + duotone variations
- Material Design (900+): Google design system
- Bootstrap Icons (2000+): Bootstrap ecosystem

Compact & Specialized:
- Radix Icons (150+): Precise 15x15px, ~5KB
- Simple Icons (3300+): Brand logos only
- Iconify (200K+): Universal framework, CDN-based

### Quick Decision Matrix

| Scenario | Best Choice | Why |
|----------|-------------|-----|
| Want maximum icons | Iconify | 200K+ icons from 150+ sets |
| Dashboard application | Tabler Icons | 5900 optimized icons, 24px |
| Tailwind CSS project | Heroicons | Official integration |
| Flexible styling needed | Phosphor | 6 weights + duotone |
| Minimal bundle size | Radix Icons | 5KB, precise 15x15px |
| Brand logos | Simple Icons | 3300+ company logos |
| General purpose UI | Lucide | 1000+ modern, well-designed |

### Bundle Size Comparison

```
Radix Icons: ~5KB (150 icons)
Heroicons: ~10KB (300 icons)
Tabler Icons: ~22KB (5900 icons)
Ionicons: ~25KB (1300 icons)
Phosphor: ~25KB (800 icons with weights)
Lucide: ~30KB (1000 icons)
Simple Icons: ~50KB (3300+ brand icons)
React Icons: Modular (varies by library)
```

## Quick Installation Commands

```bash
# Core libraries
npm install lucide-react
npm install @heroicons/react
npm install @phosphor-icons/react
npm install @tabler/icons-react
npm install @radix-ui/react-icons

# Multi-library support
npm install react-icons
npm install @iconify/react

# Brand icons
npm install simple-icons
```

Version: 4.0.0 Enterprise
Last Updated: 2025-11-13
Status: Production Ready
Enterprise Grade: Full Enterprise Support

## Implementation Guide

## Level 2: Practical Implementation

### Core Library Patterns

#### Lucide React - General Purpose (1000+ icons)

```tsx
import { Heart, Search, Settings, ChevronRight } from 'lucide-react'

export function LucideExample() {
 return (
 <div className="space-y-4">
 {/* Basic usage (24px default) */}
 <div className="flex items-center gap-2">
 <Search />
 <span>Search</span>
 </div>

 {/* Custom styling */}
 <Heart size={32} color="#ff0000" fill="#ff0000" />

 {/* Tailwind integration */}
 <Settings className="w-6 h-6 text-gray-500 hover:text-gray-900" />

 {/* Icon button */}
 <button className="p-2 rounded-lg hover:bg-gray-100">
 <ChevronRight size={20} />
 </button>
 </div>
 )
}
```

#### React Icons - Multi-Library (35K+ icons)

```tsx
import { FaHome } from "react-icons/fa" // Font Awesome
import { MdHome } from "react-icons/md" // Material Design
import { BsHouse } from "react-icons/bs" // Bootstrap
import { FiHome } from "react-icons/fi" // Feather
import { SiReact } from "react-icons/si" // Brand logos

export function MultiLibraryExample() {
 return (
 <div className="flex gap-4">
 <FaHome size={32} className="text-blue-600" />
 <MdHome size={32} className="text-green-600" />
 <BsHouse size={32} className="text-purple-600" />
 <FiHome size={32} className="text-orange-600" />
 <SiReact size={32} className="text-cyan-500" />
 </div>
 )
}
```

#### Phosphor Icons - Weight Variations (800+ icons)

```tsx
import { Heart, Star } from "@phosphor-icons/react"

export function PhosphorExample() {
 const [rating, setRating] = React.useState(3)

 return (
 <div className="space-y-4">
 {/* Weight variations */}
 <div className="flex gap-2">
 <Heart weight="thin" />
 <Heart weight="light" />
 <Heart weight="regular" />
 <Heart weight="bold" />
 <Heart weight="fill" />
 </div>

 {/* Interactive rating */}
 <div className="flex gap-1">
 {[1, 2, 3, 4, 5].map((star) => (
 <button key={star}>
 <Star
 weight={star <= rating ? "fill" : "regular"}
 size={24}
 color={star <= rating ? "#fbbf24" : "#d1d5db"}
 />
 </button>
 ))}
 </div>
 </div>
 )
}
```

#### Iconify - Universal Framework (200K+ icons)

```tsx
import { Icon } from "@iconify/react"

export function IconifyExample() {
 return (
 <div className="space-y-4">
 {/* String-based (CDN loaded) */}
 <Icon icon="fa:home" width="32" height="32" />
 <Icon icon="mdi:account" width="32" height="32" />
 <Icon icon="bi:house" width="32" height="32" />

 {/* Custom styling */}
 <Icon
 icon="heroicons:heart"
 width="48"
 height="48"
 style={{ color: "#ef4444" }}
 />
 </div>
 )
}
```

### Type-Safe Icon Button

```tsx
import { FC, SVGProps } from 'react'

type IconType = FC<SVGProps<SVGSVGElement>>

interface IconButtonProps {
 icon: IconType
 label: string
 variant?: 'primary' | 'secondary' | 'ghost'
 size?: 'sm' | 'md' | 'lg'
 onClick?: () => void
}

const sizeMap = {
 sm: 'w-4 h-4',
 md: 'w-5 h-5',
 lg: 'w-6 h-6',
}

const variantMap = {
 primary: 'bg-blue-500 text-white hover:bg-blue-600',
 secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300',
 ghost: 'text-gray-600 hover:text-gray-900 hover:bg-gray-100',
}

export function IconButton({
 icon: Icon,
 label,
 variant = 'ghost',
 size = 'md',
 onClick,
}: IconButtonProps) {
 return (
 <button
 onClick={onClick}
 aria-label={label}
 title={label}
 className={`
 p-2 rounded-lg transition-all
 ${variantMap[variant]}
 `}
 >
 <Icon className={sizeMap[size]} />
 </button>
 )
}
```

## Performance & Best Practices

### Performance Optimization

```tsx
// Good: Tree-shake specific icons
import { Heart, Star } from 'lucide-react'

// Bad: Import entire library
import * as Icons from 'lucide-react'
const Icon = Icons[iconName]

// Good: Dynamic imports for large sets
const LazyIcon = React.lazy(() => import('lucide-react').then(module => ({
 default: module[iconName]
})))

// Good: Memoize components
const MemoHeart = React.memo(Heart)
```

### Bundle Optimization Strategies

1. Choose right library size: Use Radix Icons for minimal bundles
2. Import specific icons: Avoid `import *` patterns
3. Dynamic loading: Load icons on-demand for large sets
4. Icon subsets: Create custom bundles per feature
5. Tree-shaking: Use ES modules and bundler optimization

### Accessibility Essentials

- Use `aria-label` for icon-only buttons
- Ensure 4.5:1 color contrast ratio
- Support high contrast mode with `currentColor`
- Don't rely on color alone for meaning
- Use semantic HTML structure
- Test with screen readers

## Library Comparison Summary

| Library | Icons | Bundle Size | Best For |
|---------|-------|-------------|----------|
| Lucide | 1000+ | ~30KB | General purpose UI |
| Heroicons | 300+ | ~10KB | Tailwind CSS projects |
| Phosphor | 800+ | ~25KB | Weight flexibility needed |
| Tabler | 5900+ | ~22KB | Dashboard applications |
| Radix | 150+ | ~5KB | Minimal bundle size |
| React Icons | 35K+ | Modular | Multi-library support |
| Iconify | 200K+ | CDN | Maximum icon variety |

## Advanced Patterns

## Level 3: Advanced Integration

### Custom Icon Component

```tsx
import { forwardRef, SVGProps } from 'react'

interface CustomIconProps extends SVGProps<SVGSVGElement> {
 isActive?: boolean
 tooltip?: string
}

export const CustomIcon = forwardRef<SVGSVGElement, CustomIconProps>(
 ({ isActive, tooltip, className = '', ...props }, ref) => (
 <svg
 ref={ref}
 viewBox="0 0 24 24"
 width="24"
 height="24"
 className={`
 ${isActive ? 'text-blue-500' : 'text-gray-400'}
 ${tooltip ? 'cursor-help' : ''}
 ${className}
 transition-colors duration-200
 `}
 title={tooltip}
 {...props}
 >
 <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z" />
 </svg>
 )
)
```

### Icon Theme System

```tsx
import { Heart, Settings } from 'lucide-react'

type IconTheme = 'light' | 'dark' | 'accent'

interface IconThemeConfig {
 color: string
 strokeWidth: number
 opacity: number
}

const themeConfig: Record<IconTheme, IconThemeConfig> = {
 light: { color: '#e5e7eb', strokeWidth: 2, opacity: 1 },
 dark: { color: '#1f2937', strokeWidth: 2, opacity: 1 },
 accent: { color: '#0ea5e9', strokeWidth: 2.5, opacity: 1 },
}

export function ThemedIcon({ theme, size = 24 }: { theme: IconTheme; size?: number }) {
 const config = themeConfig[theme]

 return (
 <div className="flex gap-4">
 <Heart size={size} color={config.color} strokeWidth={config.strokeWidth} />
 <Settings size={size} color={config.color} strokeWidth={config.strokeWidth} />
 </div>
 )
}
```

### Icon Animation

```tsx
import { Heart } from 'lucide-react'
import { useState } from 'react'

export function AnimatedIcon() {
 const [isActive, setIsActive] = useState(false)

 return (
 <button onClick={() => setIsActive(!isActive)} className="p-4">
 <Heart
 size={32}
 className={`
 text-red-500 transition-all duration-300
 ${isActive ? 'scale-125 animate-pulse' : 'scale-100'}
 `}
 fill={isActive ? '#ff0000' : 'none'}
 />
 </button>
 )
}
```
