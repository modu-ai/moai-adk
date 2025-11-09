# moai-icons-vector

**Vector Icon Libraries: Lucide, Heroicons, Radix Icons for Modern Web Apps**

> **Primary Agent**: frontend-expert
> **Secondary Agent**: ui-ux-expert
> **Version**: 1.0.0 (Lucide v0.4+, Heroicons v2.0+, Radix Icons v1.0+)
> **Keywords**: icons, vector icons, lucide, heroicons, radix icons, svg icons, icon library, react icons, accessibility

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Vector icons** are SVG-based, scalable icons that provide a modern alternative to emoji or font-based icons. Three popular libraries dominate the React ecosystem:

| Library | Icons | Sizes | Styles | Best For | Install |
|---------|-------|-------|--------|----------|---------|
| **Lucide** | 1000+ | All | Consistent stroke | General UI, modern design | `npm install lucide-react` |
| **Heroicons** | 300+ | 16, 20, 24px | Solid, outline | Tailwind projects | `npm install @heroicons/react` |
| **Radix Icons** | 150+ | 15x15px | Consistent | Compact, precise UI | `npm install @radix-ui/react-icons` |

**Key Advantages Over Emoji**:
- ✅ Full control over color, size, stroke width
- ✅ Scalable vector graphics (no pixelation)
- ✅ Accessibility (proper ARIA labels)
- ✅ Semantic (not emoticons)
- ✅ Design system integration
- ✅ Dark mode support
- ✅ Animation capable

**When to Use Each**:
```
Large icon set needed? → Lucide (1000+ icons)
Tailwind CSS project? → Heroicons (official Tailwind icons)
Compact UI (15px)? → Radix Icons
Custom styling needed? → Lucide (most flexible)
Accessibility critical? → Any (all support ARIA)
```

---

### Level 2: Practical Implementation (Common Patterns)

#### Pattern 1: Lucide React - Basic Usage

```bash
# Installation
npm install lucide-react
```

```tsx
import {
  Activity,
  Heart,
  Search,
  Settings,
  ChevronRight,
  AlertCircle
} from 'lucide-react'

export function LucideExample() {
  return (
    <div className="space-y-6">
      {/* Basic icon (24px default) */}
      <div className="flex items-center gap-2">
        <Activity />
        <span>Activity Monitor</span>
      </div>

      {/* Custom size */}
      <div className="flex items-center gap-2">
        <Heart size={32} />
        <span>Large heart icon</span>
      </div>

      {/* Custom color */}
      <div className="flex items-center gap-2">
        <Search size={24} color="#0ea5e9" />
        <span>Search (blue)</span>
      </div>

      {/* With stroke width */}
      <div className="flex items-center gap-2">
        <AlertCircle size={24} strokeWidth={1.5} color="#ef4444" />
        <span>Alert (thin stroke)</span>
      </div>

      {/* Fill + Stroke */}
      <div className="flex items-center gap-2">
        <Heart
          size={28}
          fill="#ff0000"
          color="#ff0000"
          strokeWidth={2}
        />
        <span>Filled heart</span>
      </div>

      {/* With Tailwind classes */}
      <div className="flex items-center gap-2">
        <Settings className="w-6 h-6 text-gray-500 hover:text-gray-900 transition-colors" />
        <span>Settings (Tailwind styled)</span>
      </div>

      {/* Icon button */}
      <button className="p-2 rounded-lg hover:bg-gray-100 transition-colors">
        <ChevronRight size={20} className="text-gray-600" />
      </button>
    </div>
  )
}
```

#### Pattern 2: Heroicons with Tailwind CSS

```bash
# Installation
npm install @heroicons/react
```

```tsx
// Import from specific size/style paths
import { BeakerIcon } from '@heroicons/react/24/solid'
import { CheckIcon } from '@heroicons/react/20/solid'
import { ChevronRightIcon } from '@heroicons/react/16/solid'

export function HeroiconsExample() {
  return (
    <div className="space-y-4">
      {/* Solid 24px icon */}
      <div className="flex items-center gap-2">
        <BeakerIcon className="h-6 w-6 text-blue-500" />
        <span>Chemistry icon</span>
      </div>

      {/* Alert with conditional styling */}
      <div className="flex items-center gap-3 p-4 bg-green-50 rounded-lg">
        <CheckIcon className="h-5 w-5 text-green-600 flex-shrink-0" />
        <p className="text-sm text-green-800">Success message</p>
      </div>

      {/* Compact 16px icon for badge */}
      <span className="inline-flex items-center gap-1 px-2 py-1 bg-yellow-100 rounded text-xs">
        <ChevronRightIcon className="h-4 w-4 text-yellow-800" />
        <span>Status update</span>
      </span>
    </div>
  )
}
```

#### Pattern 3: Radix Icons - Compact Icons

```bash
# Installation
npm install @radix-ui/react-icons
```

```tsx
import {
  FaceIcon,
  SunIcon,
  MoonIcon,
  CheckIcon,
  ExitIcon,
  DotsHorizontalIcon
} from '@radix-ui/react-icons'

export function RadixIconsExample() {
  return (
    <div className="space-y-4">
      {/* Basic Radix Icons (15x15px) */}
      <div className="flex items-center gap-2">
        <FaceIcon />
        <span>Profile</span>
      </div>

      {/* Theme toggle */}
      <div className="flex gap-2">
        <button className="p-2 rounded hover:bg-gray-100">
          <SunIcon />
        </button>
        <button className="p-2 rounded hover:bg-gray-100">
          <MoonIcon />
        </button>
      </div>

      {/* Status indicators */}
      <div className="flex items-center gap-2">
        <CheckIcon className="text-green-600" />
        <span>Verified</span>
      </div>

      {/* Menu button */}
      <button className="p-2 rounded hover:bg-gray-100">
        <DotsHorizontalIcon />
      </button>

      {/* With Tailwind sizing */}
      <div className="flex gap-2">
        <button className="p-2 text-gray-500 hover:text-gray-900 hover:bg-gray-100 rounded">
          <ExitIcon className="w-4 h-4" />
        </button>
      </div>
    </div>
  )
}
```

#### Pattern 4: Icon Button Component (Type-Safe)

```tsx
import {
  ReactNode,
  SVGProps,
  FC
} from 'react'
import { Activity, Heart, Settings } from 'lucide-react'

// Icon type definition
type IconType = FC<SVGProps<SVGSVGElement>>

interface IconButtonProps {
  icon: IconType
  label: string
  onClick?: () => void
  variant?: 'primary' | 'secondary' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
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
  onClick,
  variant = 'ghost',
  size = 'md',
  disabled = false,
}: IconButtonProps) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      title={label}
      aria-label={label}
      className={`
        p-2 rounded-lg transition-all
        ${variantMap[variant]}
        ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
      `}
    >
      <Icon className={sizeMap[size]} />
    </button>
  )
}

// Usage examples
export function IconButtonDemo() {
  return (
    <div className="flex gap-2">
      <IconButton icon={Activity} label="Activity" variant="primary" />
      <IconButton icon={Heart} label="Favorite" variant="secondary" size="lg" />
      <IconButton icon={Settings} label="Settings" variant="ghost" />
    </div>
  )
}
```

#### Pattern 5: Dynamic Icon Component (By Name)

```tsx
import {
  Heart,
  Settings,
  Search,
  AlertCircle,
  Activity,
  Clock
} from 'lucide-react'
import { useMemo } from 'react'

const iconMap = {
  heart: Heart,
  settings: Settings,
  search: Search,
  alert: AlertCircle,
  activity: Activity,
  clock: Clock,
} as const

type IconName = keyof typeof iconMap

interface DynamicIconProps {
  name: IconName
  size?: number
  color?: string
  className?: string
}

export function DynamicIcon({
  name,
  size = 24,
  color = 'currentColor',
  className = ''
}: DynamicIconProps) {
  const Icon = iconMap[name]

  if (!Icon) {
    console.warn(`Icon "${name}" not found`)
    return null
  }

  return <Icon size={size} color={color} className={className} />
}

// Usage
export function DynamicIconDemo() {
  const icons: IconName[] = ['heart', 'settings', 'search']

  return (
    <div className="flex gap-4">
      {icons.map((iconName) => (
        <DynamicIcon
          key={iconName}
          name={iconName}
          size={32}
          className="text-blue-500"
        />
      ))}
    </div>
  )
}
```

#### Pattern 6: Accessible Icon with Label

```tsx
import { AlertCircle, CheckCircle } from 'lucide-react'

interface AccessibleIconProps {
  icon: React.ReactNode
  label: string
  ariaLabel?: string
  type?: 'success' | 'error' | 'warning' | 'info'
}

export function AccessibleIcon({
  icon,
  label,
  ariaLabel,
  type = 'info'
}: AccessibleIconProps) {
  const colorMap = {
    success: 'text-green-600',
    error: 'text-red-600',
    warning: 'text-yellow-600',
    info: 'text-blue-600',
  }

  return (
    <div className="flex items-center gap-2">
      <div
        className={colorMap[type]}
        role="img"
        aria-label={ariaLabel || label}
      >
        {icon}
      </div>
      <span className="text-sm font-medium">{label}</span>
    </div>
  )
}

// Usage
export function AccessibleIconDemo() {
  return (
    <div className="space-y-2">
      <AccessibleIcon
        icon={<CheckCircle size={20} />}
        label="Payment successful"
        ariaLabel="Success: Payment was processed"
        type="success"
      />
      <AccessibleIcon
        icon={<AlertCircle size={20} />}
        label="Verification required"
        ariaLabel="Warning: Please verify your email"
        type="warning"
      />
    </div>
  )
}
```

---

### Level 3: Advanced Patterns (Expert Reference)

#### Advanced Pattern 1: Custom Icon Component with TypeScript

```tsx
import { LucideProps } from 'lucide-react'
import { forwardRef, SVGProps } from 'react'

interface CustomIconProps extends LucideProps {
  // Custom props
  isActive?: boolean
  tooltip?: string
}

export const CustomIcon = forwardRef<
  SVGSVGElement,
  CustomIconProps
>(({ isActive, tooltip, className = '', ...props }, ref) => {
  return (
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
      {/* SVG path content */}
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2m0 18c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8m3.5-9c.83 0 1.5-.67 1.5-1.5S16.33 8 15.5 8 14 8.67 14 9.5s.67 1.5 1.5 1.5m-7 0c.83 0 1.5-.67 1.5-1.5S9.33 8 8.5 8 7 8.67 7 9.5 7.67 11 8.5 11m3.5 6.5c2.33 0 4.31-1.46 5.11-3.5H6.89c.8 2.04 2.78 3.5 5.11 3.5z" />
    </svg>
  )
})

CustomIcon.displayName = 'CustomIcon'
```

#### Advanced Pattern 2: Icon Theme System

```tsx
import { Heart, Settings, Bell } from 'lucide-react'

type IconTheme = 'light' | 'dark' | 'accent'

interface IconThemeConfig {
  color: string
  strokeWidth: number
  opacity: number
}

const themeConfig: Record<IconTheme, IconThemeConfig> = {
  light: {
    color: '#e5e7eb',
    strokeWidth: 2,
    opacity: 1,
  },
  dark: {
    color: '#1f2937',
    strokeWidth: 2,
    opacity: 1,
  },
  accent: {
    color: '#0ea5e9',
    strokeWidth: 2.5,
    opacity: 1,
  },
}

interface ThemedIconProps {
  theme: IconTheme
  size?: number
}

export function ThemedIcon({ theme, size = 24 }: ThemedIconProps) {
  const config = themeConfig[theme]

  return (
    <div className="flex gap-4">
      <Heart
        size={size}
        color={config.color}
        strokeWidth={config.strokeWidth}
        style={{ opacity: config.opacity }}
      />
      <Settings
        size={size}
        color={config.color}
        strokeWidth={config.strokeWidth}
        style={{ opacity: config.opacity }}
      />
      <Bell
        size={size}
        color={config.color}
        strokeWidth={config.strokeWidth}
        style={{ opacity: config.opacity }}
      />
    </div>
  )
}
```

#### Advanced Pattern 3: Icon Animation

```tsx
import { Heart } from 'lucide-react'
import { useState } from 'react'

export function AnimatedIcon() {
  const [isAnimating, setIsAnimating] = useState(false)

  return (
    <button
      onClick={() => setIsAnimating(!isAnimating)}
      className="p-4"
    >
      <Heart
        size={32}
        className={`
          text-red-500 transition-all duration-300
          ${isAnimating ? 'scale-125 animate-pulse' : 'scale-100'}
        `}
        fill={isAnimating ? '#ff0000' : 'none'}
      />
    </button>
  )
}
```

---

## 🎯 Comparison & Best Practices

### Library Comparison

| Feature | Lucide | Heroicons | Radix |
|---------|--------|-----------|-------|
| **Icon Count** | 1000+ | 300+ | 150+ |
| **Default Size** | 24px | Multiple | 15px |
| **Styles** | Single | Solid, Outline | Single |
| **TypeScript** | Full | Full | Full |
| **Tree-Shaking** | Yes | Yes | Yes |
| **Bundle Size** | Small | Small | Smallest |
| **Customization** | High | Medium | Low |

### Accessibility Checklist

- ✅ Use `aria-label` for icon-only buttons
- ✅ Wrap icons with text in semantically meaningful containers
- ✅ Use `role="img"` only when necessary (icon is content)
- ✅ Ensure adequate color contrast (4.5:1 for text)
- ✅ Don't use color alone to convey meaning (pair with text/icon variation)
- ✅ Support high contrast mode (use `currentColor` when possible)

### Performance Best Practices

```tsx
// ✅ Good: Tree-shake unused icons
import { Heart } from 'lucide-react'

// ❌ Bad: Import entire library
import * as Icons from 'lucide-react'
const Icon = Icons[iconName]

// ✅ Good: Use dynamic imports for large icon sets
const Icon = React.lazy(() =>
  import('lucide-react').then(module => ({
    default: module[iconName]
  }))
)

// ✅ Good: Memoize icon components
const MemoIcon = React.memo(Heart)
```

---

## 📚 Official References

- **Lucide Icons**: https://lucide.dev/
- **Lucide React Docs**: https://lucide.dev/guide/packages/lucide-react
- **Heroicons**: https://heroicons.com/
- **Heroicons React**: https://github.com/tailwindlabs/heroicons
- **Radix Icons**: https://radix-ui.com/icons
- **Radix Icons React**: https://github.com/radix-ui/icons

---

## 🔗 Related Skills

- `Skill("moai-lang-tailwind-css")` – Styling icons with Tailwind
- `Skill("moai-lib-shadcn-ui")` – shadcn/ui uses Lucide by default
- `Skill("moai-lang-html-css")` – SVG accessibility basics
