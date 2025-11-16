# Advanced Examples & Production Patterns

## Example 1: Icon Selection Component

```tsx
import { Heart, Settings, Search, Bell, ChevronRight } from 'lucide-react'
import { useState } from 'react'

interface IconSelectorProps {
  selected?: string
  onSelect?: (iconName: string) => void
}

const availableIcons = {
  heart: Heart,
  settings: Settings,
  search: Search,
  bell: Bell,
  chevron: ChevronRight,
}

export function IconSelector({ selected, onSelect }: IconSelectorProps) {
  return (
    <div className="flex gap-3">
      {Object.entries(availableIcons).map(([name, Icon]) => (
        <button
          key={name}
          onClick={() => onSelect?.(name)}
          className={`p-3 rounded-lg transition-all ${
            selected === name
              ? 'bg-blue-500 text-white'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }`}
          title={name}
        >
          <Icon size={24} />
        </button>
      ))}
    </div>
  )
}
```

## Example 2: Icon with Badge

```tsx
import { Heart } from 'lucide-react'

interface IconWithBadgeProps {
  icon: React.ReactNode
  badgeCount?: number
  badgeColor?: string
}

export function IconWithBadge({
  icon,
  badgeCount,
  badgeColor = 'bg-red-500'
}: IconWithBadgeProps) {
  return (
    <div className="relative inline-block">
      {icon}
      {badgeCount !== undefined && (
        <span className={`
          absolute -top-2 -right-2
          ${badgeColor}
          text-white text-xs font-bold
          w-5 h-5 rounded-full
          flex items-center justify-center
        `}>
          {badgeCount > 99 ? '99+' : badgeCount}
        </span>
      )}
    </div>
  )
}

// Usage
export function BadgeExample() {
  return (
    <div className="flex gap-4">
      <IconWithBadge icon={<Heart size={24} />} badgeCount={5} />
      <IconWithBadge
        icon={<Heart size={24} />}
        badgeCount={150}
        badgeColor="bg-blue-500"
      />
    </div>
  )
}
```

## Example 3: Icon Grid with Search

```tsx
import { Search, Heart, Settings, Clock, AlertCircle, CheckCircle } from 'lucide-react'
import { useState, useMemo } from 'react'

const allIcons = [
  { name: 'heart', component: Heart },
  { name: 'settings', component: Settings },
  { name: 'clock', component: Clock },
  { name: 'alert', component: AlertCircle },
  { name: 'check', component: CheckCircle },
]

export function IconGrid() {
  const [search, setSearch] = useState('')

  const filtered = useMemo(() => {
    return allIcons.filter(icon =>
      icon.name.toLowerCase().includes(search.toLowerCase())
    )
  }, [search])

  return (
    <div className="space-y-4">
      <div className="relative">
        <Search className="absolute left-3 top-3 text-gray-400" size={20} />
        <input
          type="text"
          placeholder="Search icons..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div className="grid grid-cols-3 md:grid-cols-5 lg:grid-cols-6 gap-4">
        {filtered.map(({ name, component: Icon }) => (
          <button
            key={name}
            className="p-4 border rounded-lg hover:bg-gray-50 transition-colors text-center"
            title={name}
          >
            <Icon className="mx-auto mb-2 text-gray-600" size={28} />
            <span className="text-xs text-gray-600 break-words">{name}</span>
          </button>
        ))}
      </div>

      {filtered.length === 0 && (
        <p className="text-center text-gray-500">No icons found</p>
      )}
    </div>
  )
}
```

## Example 4: Loading States with Icons

```tsx
import { Loader2, CheckCircle, AlertCircle, RotateCcw } from 'lucide-react'

type LoadingState = 'loading' | 'success' | 'error' | 'idle'

interface StatusIndicatorProps {
  state: LoadingState
  message?: string
}

export function StatusIndicator({ state, message }: StatusIndicatorProps) {
  const configs = {
    loading: {
      icon: Loader2,
      color: 'text-blue-500',
      message: message || 'Loading...'
    },
    success: {
      icon: CheckCircle,
      color: 'text-green-500',
      message: message || 'Success!'
    },
    error: {
      icon: AlertCircle,
      color: 'text-red-500',
      message: message || 'Error occurred'
    },
    idle: {
      icon: RotateCcw,
      color: 'text-gray-400',
      message: message || 'Ready'
    }
  }

  const config = configs[state]
  const Icon = config.icon
  const isAnimating = state === 'loading'

  return (
    <div className="flex items-center gap-2">
      <Icon
        size={24}
        className={`${config.color} ${isAnimating ? 'animate-spin' : ''}`}
      />
      <span className="text-sm">{config.message}</span>
    </div>
  )
}
```

## Example 5: Icon Menu

```tsx
import { Menu, X, Home, Settings, User, LogOut } from 'lucide-react'
import { useState } from 'react'

interface MenuItem {
  label: string
  icon: React.ComponentType<{ size: number }>
  onClick: () => void
}

interface IconMenuProps {
  items: MenuItem[]
}

export function IconMenu({ items }: IconMenuProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="relative inline-block">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="p-2 rounded-lg hover:bg-gray-100"
        aria-label="Menu"
      >
        {isOpen ? <X size={24} /> : <Menu size={24} />}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-48 bg-white border rounded-lg shadow-lg z-10">
          {items.map((item, index) => {
            const Icon = item.icon
            return (
              <button
                key={index}
                onClick={() => {
                  item.onClick()
                  setIsOpen(false)
                }}
                className="w-full flex items-center gap-3 px-4 py-2 hover:bg-gray-100 text-left"
              >
                <Icon size={20} />
                <span>{item.label}</span>
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}

// Usage
export function MenuExample() {
  return (
    <IconMenu
      items={[
        { label: 'Home', icon: Home, onClick: () => console.log('Home') },
        { label: 'Profile', icon: User, onClick: () => console.log('Profile') },
        { label: 'Settings', icon: Settings, onClick: () => console.log('Settings') },
        { label: 'Logout', icon: LogOut, onClick: () => console.log('Logout') },
      ]}
    />
  )
}
```

## Example 6: Theme-Aware Icon System

```tsx
import { Sun, Moon, Settings } from 'lucide-react'
import { useState } from 'react'

type Theme = 'light' | 'dark' | 'auto'

interface ThemeAwareIconProps {
  size?: number
}

const iconColors = {
  light: '#fbbf24',
  dark: '#3b82f6',
  auto: '#6b7280'
}

export function ThemeAwareIcons({ size = 24 }: ThemeAwareIconProps) {
  const [theme, setTheme] = useState<Theme>('auto')

  const getIconColor = (t: Theme) => iconColors[t]

  return (
    <div className="space-y-4">
      <div className="flex gap-4">
        <button
          onClick={() => setTheme('light')}
          className={`p-3 rounded-lg ${theme === 'light' ? 'bg-yellow-100' : 'bg-gray-100'}`}
        >
          <Sun
            size={size}
            color={getIconColor('light')}
          />
        </button>

        <button
          onClick={() => setTheme('dark')}
          className={`p-3 rounded-lg ${theme === 'dark' ? 'bg-blue-100' : 'bg-gray-100'}`}
        >
          <Moon
            size={size}
            color={getIconColor('dark')}
          />
        </button>

        <button
          onClick={() => setTheme('auto')}
          className={`p-3 rounded-lg ${theme === 'auto' ? 'bg-gray-200' : 'bg-gray-100'}`}
        >
          <Settings
            size={size}
            color={getIconColor('auto')}
          />
        </button>
      </div>

      <p className="text-sm text-gray-600">
        Current theme: <strong>{theme}</strong>
      </p>
    </div>
  )
}
```

## Example 7: Icon Button Group

```tsx
import { Bold, Italic, Underline, AlignLeft, AlignCenter, AlignRight } from 'lucide-react'
import { useState } from 'react'

interface TextFormat {
  bold: boolean
  italic: boolean
  underline: boolean
  align: 'left' | 'center' | 'right'
}

export function TextFormatToolbar() {
  const [format, setFormat] = useState<TextFormat>({
    bold: false,
    italic: false,
    underline: false,
    align: 'left'
  })

  const toggleFormat = (key: keyof Omit<TextFormat, 'align'>) => {
    setFormat(prev => ({
      ...prev,
      [key]: !prev[key]
    }))
  }

  const setAlign = (align: TextFormat['align']) => {
    setFormat(prev => ({ ...prev, align }))
  }

  return (
    <div className="flex gap-1 p-2 border rounded-lg bg-gray-50">
      {/* Format buttons */}
      <button
        onClick={() => toggleFormat('bold')}
        className={`p-2 rounded ${format.bold ? 'bg-blue-500 text-white' : 'hover:bg-gray-200'}`}
        title="Bold"
      >
        <Bold size={18} />
      </button>

      <button
        onClick={() => toggleFormat('italic')}
        className={`p-2 rounded ${format.italic ? 'bg-blue-500 text-white' : 'hover:bg-gray-200'}`}
        title="Italic"
      >
        <Italic size={18} />
      </button>

      <button
        onClick={() => toggleFormat('underline')}
        className={`p-2 rounded ${format.underline ? 'bg-blue-500 text-white' : 'hover:bg-gray-200'}`}
        title="Underline"
      >
        <Underline size={18} />
      </button>

      {/* Divider */}
      <div className="w-px bg-gray-300 mx-1" />

      {/* Alignment buttons */}
      <button
        onClick={() => setAlign('left')}
        className={`p-2 rounded ${format.align === 'left' ? 'bg-blue-500 text-white' : 'hover:bg-gray-200'}`}
        title="Align Left"
      >
        <AlignLeft size={18} />
      </button>

      <button
        onClick={() => setAlign('center')}
        className={`p-2 rounded ${format.align === 'center' ? 'bg-blue-500 text-white' : 'hover:bg-gray-200'}`}
        title="Align Center"
      >
        <AlignCenter size={18} />
      </button>

      <button
        onClick={() => setAlign('right')}
        className={`p-2 rounded ${format.align === 'right' ? 'bg-blue-500 text-white' : 'hover:bg-gray-200'}`}
        title="Align Right"
      >
        <AlignRight size={18} />
      </button>
    </div>
  )
}
```

## Example 8: Notification with Icon

```tsx
import { AlertCircle, CheckCircle, AlertTriangle, Info, X } from 'lucide-react'

type NotificationType = 'info' | 'success' | 'warning' | 'error'

interface NotificationProps {
  type: NotificationType
  title: string
  message: string
  onClose?: () => void
}

export function Notification({
  type,
  title,
  message,
  onClose
}: NotificationProps) {
  const configs = {
    info: {
      icon: Info,
      bgColor: 'bg-blue-50',
      iconColor: 'text-blue-600',
      borderColor: 'border-blue-200'
    },
    success: {
      icon: CheckCircle,
      bgColor: 'bg-green-50',
      iconColor: 'text-green-600',
      borderColor: 'border-green-200'
    },
    warning: {
      icon: AlertTriangle,
      bgColor: 'bg-yellow-50',
      iconColor: 'text-yellow-600',
      borderColor: 'border-yellow-200'
    },
    error: {
      icon: AlertCircle,
      bgColor: 'bg-red-50',
      iconColor: 'text-red-600',
      borderColor: 'border-red-200'
    }
  }

  const config = configs[type]
  const Icon = config.icon

  return (
    <div className={`
      flex items-start gap-4 p-4
      ${config.bgColor}
      border ${config.borderColor}
      rounded-lg
    `}>
      <Icon className={`${config.iconColor} flex-shrink-0 mt-0.5`} size={20} />

      <div className="flex-1">
        <h3 className="font-semibold text-gray-900">{title}</h3>
        <p className="text-sm text-gray-600 mt-1">{message}</p>
      </div>

      {onClose && (
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-600 flex-shrink-0"
          aria-label="Close"
        >
          <X size={18} />
        </button>
      )}
    </div>
  )
}
```
