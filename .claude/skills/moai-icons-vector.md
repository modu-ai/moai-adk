# moai-icons-vector

**Vector Icon Libraries: Complete Ecosystem Guide (10+ Libraries, 200K+ Icons)**

> **Primary Agent**: frontend-expert
> **Secondary Agent**: ui-ux-expert
> **Version**: 1.1.0 (Lucide v0.4+, React Icons 35K+, Tabler v2.0+, Phosphor v1.4+, Heroicons v2.0+, Radix Icons v1.0+, Iconify v2.0+)
> **Keywords**: icons, vector icons, lucide, react icons, tabler icons, phosphor icons, heroicons, radix icons, iconify, svg icons, icon library, icon design system, accessibility

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Vector icons** are SVG-based, scalable icons that provide a modern alternative to emoji or font-based icons. Popular libraries span different use cases and design philosophies:

#### Tier 1: Ecosystem Leaders (1000+ icons)
| Library | Icons | Styles | Bundle | Best For | Install |
|---------|-------|--------|--------|----------|---------|
| **Lucide** | 1000+ | Single stroke | ~30KB | General UI, modern | `npm install lucide-react` |
| **React Icons** | 35K+ | Multiple sets | Modular | Multi-library support | `npm install react-icons` |
| **Tabler Icons** | 5900+ | 24px single stroke | ~22KB | Dashboard, consistent | `npm install @tabler/icons-react` |
| **Ionicons** | 1300+ | Material/iOS | ~25KB | Mobile + web | `npm install ionicons` |

#### Tier 2: Specialist Libraries (300-800 icons)
| Library | Icons | Styles | Best For | Install |
|---------|-------|--------|----------|---------|
| **Heroicons** | 300+ | Solid, outline | Tailwind projects | `npm install @heroicons/react` |
| **Phosphor** | 800+ | Thin-Bold, duotone | Flexible weights | `npm install @phosphor-icons/react` |
| **Material Design** | 900+ | Material style | Google design | `npm install @mui/icons-material` |
| **Bootstrap Icons** | 2000+ | SVG, webfont | Bootstrap ecosystem | `npm install bootstrap-icons` |

#### Tier 3: Compact & Specialized
| Library | Icons | Best For | Install |
|---------|-------|----------|---------|
| **Radix Icons** | 150+ | Precise 15x15px | `npm install @radix-ui/react-icons` |
| **Simple Icons** | 3300+ | Brand logos | `npm install simple-icons` |
| **Iconify** | 200K+ | Universal framework | `npm install @iconify/react` |

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

### Level 1.5: Icon Library Comparison Matrix

#### 선택 기준별 추천

**아이콘 개수 필요?**
- 100개 이상: React Icons (35K+), Tabler (5900+)
- 1000개 이상: Lucide (1000+), Ionicons (1300+)
- 기업 로고: Simple Icons (3300+)
- 모든 아이콘: Iconify (200K+)

**설계 스타일?**
- Stroke (일관성): Lucide, Tabler, Heroicons
- Weighted (다양성): Phosphor (thin~bold, duotone)
- Material Design: @mui/icons-material, Material Icons
- 간결함: Radix Icons (15x15px 정확)

**성능 중요?**
- 최소 번들: Radix Icons (~5KB), Heroicons (~10KB)
- 선택적 로드: React Icons (library별 import)
- Tree-shaking: 모든 라이브러리 지원
- 동적 로드: Iconify (on-demand CDN)

**프레임워크?**
- React only: 모든 라이브러리 지원
- React + Tailwind: Heroicons (공식 통합)
- Vue: Phosphor, Tabler, Bootstrap Icons
- 멀티-프레임워크: Iconify (React, Vue, Angular, Svelte)
- React Native: Tabler, Ionicons, Phosphor

#### 번들 크기 비교

```
Radix Icons:        ~5KB
Heroicons:         ~10KB
Lucide:            ~30KB (1000 icons)
Tabler Icons:      ~22KB (5900 icons)
Ionicons:          ~25KB (1300 icons)
React Icons:   Modular (fa: ~30KB, md: ~100KB, etc)
Phosphor:          ~25KB (800 icons with weights)
Simple Icons:      ~50KB (3300+ brand icons)
```

---

### Level 1.6: Library Selection Decision Tree

**Use this flowchart to choose the right icon library for your project:**

```
Start: I need icons for my project
│
├─ Need 200K+ icons from 150+ sets?
│  ├─ YES → Iconify (완벽한 범용성)
│  └─ NO → Continue
│
├─ Building a dashboard or admin UI?
│  ├─ YES: Tabler Icons (5900+ 최적화된 아이콘)
│  └─ NO → Continue
│
├─ Using Tailwind CSS?
│  ├─ YES: Heroicons (공식 Tailwind 통합)
│  └─ NO → Continue
│
├─ Need weight variations (thin, light, bold, fill, etc.)?
│  ├─ YES: Phosphor Icons (6가지 무게 + duotone)
│  └─ NO → Continue
│
├─ Need 30K+ icons from multiple design systems?
│  ├─ YES: React Icons (Font Awesome + Material + Bootstrap + etc.)
│  └─ NO → Continue
│
├─ Prioritize smallest bundle size?
│  ├─ YES: Radix Icons (~5KB)
│  └─ NO → Continue
│
├─ Need brand logos primarily?
│  ├─ YES: Simple Icons (3300+ 브랜드 로고)
│  └─ NO → Continue
│
└─ Default recommendation: Lucide (1000+ 모던한 디자인)
```

**Quick Decision Matrix**:

| Scenario | Best Choice | Why |
|----------|-------------|-----|
| Want it all | Iconify | 200K+ icons, all frameworks |
| Dashboard app | Tabler Icons | 5900 optimized icons, 24px |
| Tailwind project | Heroicons | Official integration, 300+ icons |
| Flexible weights | Phosphor | 6 weights per icon, duotone |
| Multi-style | React Icons | 30+ design systems, 35K+ total |
| Minimal bundle | Radix Icons | 5KB, precise 15x15px |
| Brand logos | Simple Icons | 3300+ company logos |
| General UI | Lucide | 1000+ modern, well-designed |

---

### Level 2: Practical Implementation (Common Patterns)

#### Pattern 1: React Icons - Multi-Library Support (35K+ Icons)

**특징**: 30개+ 아이콘 라이브러리를 하나의 import로 통합 (Font Awesome, Material Design, Bootstrap, Feather, Ionicons 등)

```bash
# Installation
npm install react-icons
```

```tsx
// 다양한 라이브러리에서 아이콘 선택
import { FaBeer } from "react-icons/fa"           // Font Awesome (Solid)
import { FaRegClock } from "react-icons/fa"       // Font Awesome (Regular)
import { FaHouse } from "react-icons/fa6"         // Font Awesome v6
import { MdAccessibility } from "react-icons/md"  // Material Design
import { BsFolder, BsFillHouseFill } from "react-icons/bs"  // Bootstrap Icons
import { FiHome, FiSettings } from "react-icons/fi"        // Feather Icons
import { HiHome, HiOutlineCog } from "react-icons/hi"      // Heroicons v1
import { HiMiniHome } from "react-icons/hi2"     // Heroicons v2
import { IoMdHome, IoHome } from "react-icons/io" // Ionicons
import { AiFillHome } from "react-icons/ai"      // Ant Design Icons
import { RiHomeLine } from "react-icons/ri"      // Remix Icon
import { TbHome } from "react-icons/tb"          // Tabler Icons
import { LuHome } from "react-icons/lu"          // Lucide Icons (through react-icons)
import { GiSword } from "react-icons/gi"         // Game Icons
import { SiReact } from "react-icons/si"         // Simple Icons (brand logos)

export function MultiLibraryIcons() {
  return (
    <div className="flex flex-wrap gap-6">
      {/* Font Awesome */}
      <div className="flex flex-col items-center">
        <FaBeer size={32} className="text-yellow-600" />
        <span className="text-sm">Font Awesome</span>
      </div>

      {/* Material Design */}
      <div className="flex flex-col items-center">
        <MdAccessibility size={32} className="text-blue-600" />
        <span className="text-sm">Material Design</span>
      </div>

      {/* Bootstrap Icons */}
      <div className="flex flex-col items-center">
        <BsFillHouseFill size={32} className="text-green-600" />
        <span className="text-sm">Bootstrap Icons</span>
      </div>

      {/* Feather Icons */}
      <div className="flex flex-col items-center">
        <FiSettings size={32} className="text-purple-600" />
        <span className="text-sm">Feather Icons</span>
      </div>

      {/* Ant Design Icons */}
      <div className="flex flex-col items-center">
        <AiFillHome size={32} className="text-red-600" />
        <span className="text-sm">Ant Design Icons</span>
      </div>

      {/* Simple Icons (Brand Logos) */}
      <div className="flex flex-col items-center">
        <SiReact size={32} className="text-cyan-500" />
        <span className="text-sm">Brand Icons</span>
      </div>
    </div>
  )
}
```

**장점**: 여러 라이브러리를 한번에 사용 가능, 트리샤킹 지원
**단점**: 번들 크기가 라이브러리마다 다름 (선택적 설치 권장)

---

#### Pattern 1b: Lucide React - Basic Usage

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

#### Pattern 2: Phosphor Icons - 6 Weights + Duotone (800 Icons)

**특징**: 각 아이콘마다 6가지 무게(thin, light, regular, bold, fill, duotone) 제공

```bash
npm install @phosphor-icons/react
```

```tsx
import {
  Heart,
  Horse,
  Cube,
  Bell,
  GraduationCap
} from "@phosphor-icons/react"
import { IconContext } from "@phosphor-icons/react"

// 방법 1: 개별 아이콘 커스터마이징
export function PhosphorBasic() {
  return (
    <div className="space-y-4">
      {/* 기본 사용 */}
      <Heart />

      {/* 무게 선택 */}
      <Heart weight="thin" size={32} />
      <Heart weight="light" size={32} />
      <Heart weight="regular" size={32} />
      <Heart weight="bold" size={32} />
      <Heart weight="fill" size={32} color="#ff0000" />
      <Heart weight="duotone" size={32} color="#ff0000" />

      {/* 색상 + 무게 + 크기 조합 */}
      <Horse
        weight="bold"
        size={48}
        color="teal"
      />
    </div>
  )
}

// 방법 2: Context로 기본값 설정
export function PhosphorWithContext() {
  return (
    <IconContext.Provider
      value={{
        color: "limegreen",
        size: 32,
        weight: "bold",
        mirrored: false,
      }}
    >
      <div className="flex gap-4">
        <Heart />     {/* lime-green, 32px, bold */}
        <Horse />     {/* lime-green, 32px, bold */}
        <Cube />      {/* lime-green, 32px, bold */}
        {/* 개별 props로 오버라이드 가능 */}
        <Bell color="red" weight="fill" />
      </div>
    </IconContext.Provider>
  )
}

// 방법 3: 동적 무게 토글 (예: Rating)
export function PhosphorRating() {
  const [rating, setRating] = React.useState(0)

  return (
    <div className="flex gap-2">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          onClick={() => setRating(star)}
          className="hover:scale-110 transition-transform"
        >
          <Heart
            weight={star <= rating ? "fill" : "regular"}
            size={32}
            color={star <= rating ? "#ff0000" : "#ccc"}
          />
        </button>
      ))}
    </div>
  )
}
```

**장점**: 가장 유연한 무게 시스템, duotone 지원, RTL 미러링
**단점**: 무게당 파일 크기 증가 (하지만 선택적 로드 가능)

---

#### Pattern 3: Tabler Icons - Dashboard-Optimized (5900 Icons)

**특징**: 24x24px 기본 크기, 모두 2px stroke, 대시보드 UI에 최적화

```bash
npm install @tabler/icons-react
```

```tsx
import {
  IconArrowLeft,
  IconHome,
  IconHeart,
  IconAward,
  IconSearch,
  IconBell,
  IconSettings
} from "@tabler/icons-react"

export function TablerBasic() {
  return (
    <div className="space-y-4">
      {/* 기본 사용 */}
      <div className="flex items-center gap-2">
        <IconHome />
        <span>Home</span>
      </div>

      {/* 커스터마이징 */}
      <IconHeart
        size={36}
        color="red"
        stroke={3}  // stroke-width
        strokeLinejoin="miter"
      />

      {/* 대시보드 UI 예제 */}
      <div className="grid grid-cols-3 gap-4">
        <Card icon={<IconAward size={24} />} label="Awards" value="12" />
        <Card icon={<IconHome size={24} />} label="Homes" value="5" />
        <Card icon={<IconHeart size={24} />} label="Likes" value="240" />
      </div>
    </div>
  )
}

// 타입-안전 Card 컴포넌트
interface CardProps {
  icon: React.ReactNode
  label: string
  value: string
}

function Card({ icon, label, value }: CardProps) {
  return (
    <div className="p-4 border rounded-lg">
      <div className="flex items-center gap-2 mb-2">
        {icon}
        <span className="text-sm font-medium">{label}</span>
      </div>
      <span className="text-2xl font-bold">{value}</span>
    </div>
  )
}

// Tabler 아이콘 제목으로 사용 (흔한 패턴)
export function TablerHeadings() {
  return (
    <div className="space-y-6">
      <h1 className="flex items-center gap-2 text-3xl font-bold">
        <IconSearch size={40} className="text-blue-600" />
        Search Results
      </h1>

      <h2 className="flex items-center gap-2 text-2xl font-bold">
        <IconBell size={32} className="text-orange-600" />
        Notifications
      </h2>

      <div className="flex items-center gap-2 p-3 bg-blue-50 rounded-lg">
        <IconSettings size={20} className="text-blue-700 flex-shrink-0" />
        <span className="text-sm">System settings updated</span>
      </div>
    </div>
  )
}
```

**장점**: 5900개 아이콘, 일관된 크기, 대시보드 최적화, 번들 작음
**단점**: 24px 고정, 무게 변화 없음

---

#### Pattern 4: Iconify - Universal Icon Framework (200K+ Icons)

**특징**: 150개+ 아이콘 세트를 하나의 API로 접근 (CDN 기반 동적 로드)

```bash
npm install @iconify/react
# 또는 HTML의 경우 CDN 사용
```

```tsx
import { Icon } from "@iconify/react"
import homeIcon from "@iconify-icons/mdi/home"
import accountIcon from "@iconify-icons/mdi/account"

// 방법 1: 아이콘 문자열로 참조 (CDN 동적 로드)
export function IconifyStringBased() {
  return (
    <div className="space-y-4">
      {/* FontAwesome 아이콘 */}
      <Icon icon="fa:home" width="32" height="32" />

      {/* Material Design Icons */}
      <Icon icon="mdi:home" width="32" height="32" />

      {/* Bootstrap Icons */}
      <Icon icon="bi:house" width="32" height="32" />

      {/* Feather Icons */}
      <Icon icon="feather:home" width="32" height="32" />

      {/* 색상 + 크기 */}
      <Icon
        icon="eva:people-outline"
        width="48"
        height="48"
        style={{ color: "#0ea5e9" }}
      />
    </div>
  )
}

// 방법 2: 가져온 아이콘 컴포넌트
export function IconifyImported() {
  return (
    <div className="flex gap-4">
      <Icon icon={homeIcon} width="32" />
      <Icon icon={accountIcon} width="32" />
    </div>
  )
}

// 방법 3: 다양한 아이콘 세트 비교
export function IconifyMultipleSets() {
  const iconName = "home"

  return (
    <div className="grid grid-cols-3 gap-4">
      <div className="text-center">
        <Icon
          icon={`fa:${iconName}`}
          width="40"
          height="40"
          className="mb-2"
        />
        <span className="text-xs">Font Awesome</span>
      </div>

      <div className="text-center">
        <Icon
          icon={`mdi:${iconName}`}
          width="40"
          height="40"
          className="mb-2"
        />
        <span className="text-xs">Material Design</span>
      </div>

      <div className="text-center">
        <Icon
          icon={`heroicons-outline:${iconName}`}
          width="40"
          height="40"
          className="mb-2"
        />
        <span className="text-xs">Heroicons</span>
      </div>
    </div>
  )
}
```

**HTML/CSS로도 사용 가능** (JavaScript 최소화):
```html
<script src="https://code.iconify.design/1/1.0.8/iconify.min.js"></script>

<!-- FontAwesome 아이콘 -->
<span class="iconify" data-icon="fa:home"></span>

<!-- Material Design -->
<span class="iconify" data-icon="mdi:home"></span>

<!-- 색상 + 크기 제어 -->
<span
  class="iconify"
  data-icon="eva:people-outline"
  style="color: #0ea5e9; font-size: 48px;"
></span>
```

**장점**: 200K+ 아이콘, 150+ 세트 지원, 동적 로드, 다중 프레임워크
**단점**: CDN 의존성, 네트워크 요청

---

#### Pattern 2b: Heroicons with Tailwind CSS

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

### Library Comparison Matrix

| Feature | Lucide | React Icons | Tabler | Heroicons | Phosphor | Radix | Iconify |
|---------|--------|-------------|--------|-----------|----------|-------|---------|
| **Icon Count** | 1000+ | 35K+ | 5900+ | 300+ | 800+ | 150+ | 200K+ |
| **Default Size** | 24px | Variable | 24px | 16/20/24 | 24px | 15px | Variable |
| **Styles** | Single | Multiple | Single stroke | Outline, Solid | 6 weights + duotone | Single | Multiple |
| **TypeScript** | Full | Full | Full | Full | Full | Full | Full |
| **Tree-Shaking** | Yes | Partial | Yes | Yes | Yes | Yes | Via CDN |
| **Bundle Size (min+gzip)** | ~30KB | Modular | ~22KB | ~10KB | ~25KB | ~5KB | CDN |
| **Customization** | High | Medium | High | Medium | Very High | Low | High |
| **Weight Support** | No | No | No | No | Yes | No | Yes |
| **Dark Mode** | Via classes | Via classes | Via classes | Via classes | Via colors | Via classes | Via style |
| **React Native** | No | Partial | Yes | No | Yes | No | Yes |
| **Framework Support** | React only | React mainly | React, Vue, Svelte | React, Vue | React, Vue, Svelte | React, Vue | All frameworks |
| **Best Use Case** | General UI | Multi-library | Dashboard UI | Tailwind CSS | Flexible design | Compact UI | Universal |

### Detailed Feature Comparison

**Customization Flexibility**:
- 🥇 **Phosphor**: 6 weight variants + duotone per icon
- 🥈 **Lucide**: Full color, size, stroke control
- 🥉 **Tabler/React Icons**: Good control, limited variants

**Bundle Size Efficiency**:
- 🥇 **Radix Icons**: ~5KB (smallest)
- 🥇 **Heroicons**: ~10KB (official Tailwind icons)
- 🥈 **Lucide/Tabler**: ~22-30KB (good balance)
- 🥉 **React Icons**: Variable per sub-library
- ⚠️ **Simple Icons**: ~50KB (many brand logos)
- 🌐 **Iconify**: CDN-based (no local bundle)

**Icon Coverage**:
- 🌐 **Iconify**: 200K+ (complete coverage)
- 📚 **React Icons**: 35K+ (multi-library aggregator)
- 📊 **Tabler**: 5900+ (dashboard-optimized)
- 🎨 **Lucide**: 1000+ (modern, well-designed)
- 🏷️ **Simple Icons**: 3300+ (brand logos)

**Framework Compatibility**:
- ✅ **React-only**: Lucide, Heroicons, React Icons, Radix
- ✅ **Multi-framework**: Tabler, Phosphor, Bootstrap Icons, Iconify
- ✅ **React Native**: Tabler, Phosphor, Ionicons

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

### Tier 1: Ecosystem Leaders

- **Lucide Icons**: https://lucide.dev/
- **Lucide React Docs**: https://lucide.dev/guide/packages/lucide-react
- **React Icons**: https://react-icons.github.io/
- **React Icons GitHub**: https://github.com/react-icons/react-icons
- **Tabler Icons**: https://tabler-icons.io/
- **Tabler Icons React**: https://github.com/tabler/tabler-icons
- **Ionicons**: https://ionicons.com/
- **Ionicons Docs**: https://ionicons.com/usage

### Tier 2: Specialist Libraries

- **Heroicons**: https://heroicons.com/
- **Heroicons React**: https://github.com/tailwindlabs/heroicons
- **Phosphor Icons**: https://phosphor.designsystem.com/
- **Phosphor React**: https://github.com/phosphor-icons/phosphor-react
- **Material Design Icons**: https://www.npmjs.com/package/@mui/icons-material
- **Bootstrap Icons**: https://icons.getbootstrap.com/

### Tier 3: Specialized Libraries

- **Radix Icons**: https://radix-ui.com/icons
- **Radix Icons React**: https://github.com/radix-ui/icons
- **Simple Icons**: https://simpleicons.org/
- **Simple Icons React**: https://www.npmjs.com/package/simple-icons
- **Iconify**: https://iconify.design/
- **Iconify React**: https://iconify.design/docs/icon-components/react/

### Additional Resources

- **Icon Performance Comparison**: https://bundlephobia.com/ (compare library bundle sizes)
- **Accessibility in SVG**: https://www.w3.org/WAI/tutorials/graphics/
- **Icon Design Systems**: https://www.designsystems.com/icons/
- **Web Accessibility**: https://www.w3.org/WAI/WCAG21/quickref/

---

## 🔗 Related Skills

- `Skill("moai-lang-tailwind-css")` – Styling icons with Tailwind
- `Skill("moai-lib-shadcn-ui")` – shadcn/ui uses Lucide by default
- `Skill("moai-lang-html-css")` – SVG accessibility basics
