# Performance Optimization

Comprehensive patterns for frontend performance including Code Splitting, Dynamic imports, Image optimization, and Memoization strategies.

---

## Code Splitting

### Route-Based Splitting (Next.js)

```tsx
// Next.js App Router automatically code-splits by route
// Each page.tsx becomes a separate chunk

// app/dashboard/page.tsx
// This is automatically lazy-loaded when navigating to /dashboard
export default function DashboardPage() {
  return <Dashboard />
}

// For layouts, use streaming with Suspense
// app/dashboard/layout.tsx
import { Suspense } from 'react'
import { DashboardSkeleton } from '@/components/skeletons'

export default function DashboardLayout({ children }) {
  return (
    <div className="dashboard-layout">
      <Suspense fallback={<DashboardSkeleton />}>
        {children}
      </Suspense>
    </div>
  )
}
```

### Component-Level Splitting

```tsx
// React lazy loading
import { lazy, Suspense } from 'react'

// Lazy load heavy components
const HeavyChart = lazy(() => import('@/components/HeavyChart'))
const DataTable = lazy(() => import('@/components/DataTable'))

function Dashboard() {
  return (
    <div className="dashboard">
      <Suspense fallback={<ChartSkeleton />}>
        <HeavyChart data={chartData} />
      </Suspense>

      <Suspense fallback={<TableSkeleton />}>
        <DataTable rows={tableData} />
      </Suspense>
    </div>
  )
}

// Named exports with lazy
const { Modal } = await import('@/components/Modal')

// Preloading for faster navigation
const ChartComponent = lazy(() => import('@/components/Chart'))

function preloadChart() {
  import('@/components/Chart')
}

function Navigation() {
  return (
    <Link
      href="/analytics"
      onMouseEnter={preloadChart}
    >
      Analytics
    </Link>
  )
}
```

---

## Dynamic Imports

### Next.js Dynamic Import

```tsx
import dynamic from 'next/dynamic'

// Basic dynamic import
const DynamicChart = dynamic(() => import('@/components/Chart'), {
  loading: () => <ChartSkeleton />,
})

// Disable SSR for client-only components
const MapComponent = dynamic(
  () => import('@/components/Map'),
  {
    ssr: false,
    loading: () => <MapSkeleton />,
  }
)

// Named exports
const Modal = dynamic(
  () => import('@/components/Modal').then((mod) => mod.Modal),
  {
    loading: () => <ModalSkeleton />,
  }
)

// With custom loading component
const Editor = dynamic(
  () => import('@/components/RichTextEditor'),
  {
    loading: () => (
      <div className="editor-skeleton animate-pulse">
        <div className="h-8 bg-gray-200 rounded" />
        <div className="h-64 bg-gray-100 rounded mt-2" />
      </div>
    ),
    ssr: false,
  }
)

// Usage
function EditorPage() {
  const [showEditor, setShowEditor] = useState(false)

  return (
    <div>
      <button onClick={() => setShowEditor(true)}>
        Open Editor
      </button>

      {showEditor && <Editor />}
    </div>
  )
}
```

### Conditional Loading Pattern

```tsx
function FeatureFlags({ features }) {
  return (
    <>
      {features.analytics && (
        <Suspense fallback={<AnalyticsSkeleton />}>
          <DynamicAnalytics />
        </Suspense>
      )}

      {features.chat && (
        <Suspense fallback={<ChatSkeleton />}>
          <DynamicChat />
        </Suspense>
      )}
    </>
  )
}

// Load based on viewport
function LazySection({ children }) {
  const [isVisible, setIsVisible] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true)
          observer.disconnect()
        }
      },
      { rootMargin: '100px' }
    )

    if (ref.current) {
      observer.observe(ref.current)
    }

    return () => observer.disconnect()
  }, [])

  return (
    <div ref={ref}>
      {isVisible ? children : <Skeleton />}
    </div>
  )
}
```

---

## Image Optimization

### Next.js Image Component

```tsx
import Image from 'next/image'

// Basic optimized image
function Hero() {
  return (
    <Image
      src="/hero.jpg"
      alt="Hero image"
      width={1200}
      height={600}
      priority // Load immediately for LCP
    />
  )
}

// Responsive images
function ProductImage({ src, alt }) {
  return (
    <Image
      src={src}
      alt={alt}
      fill
      sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
      className="object-cover"
    />
  )
}

// With placeholder blur
function ArticleImage({ src, alt, blurDataURL }) {
  return (
    <Image
      src={src}
      alt={alt}
      width={800}
      height={450}
      placeholder="blur"
      blurDataURL={blurDataURL}
    />
  )
}

// Remote images with loader
function CloudinaryImage({ publicId, alt }) {
  return (
    <Image
      loader={({ src, width, quality }) =>
        `https://res.cloudinary.com/demo/image/upload/w_${width},q_${quality || 75}/${src}`
      }
      src={publicId}
      alt={alt}
      width={500}
      height={300}
    />
  )
}
```

### Lazy Loading Images

```tsx
// Native lazy loading
function GalleryImage({ src, alt }) {
  return (
    <img
      src={src}
      alt={alt}
      loading="lazy"
      decoding="async"
    />
  )
}

// Intersection Observer pattern
function LazyImage({ src, alt, className }) {
  const [isLoaded, setIsLoaded] = useState(false)
  const [isInView, setIsInView] = useState(false)
  const imgRef = useRef<HTMLImageElement>(null)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true)
          observer.disconnect()
        }
      },
      { rootMargin: '200px' }
    )

    if (imgRef.current) {
      observer.observe(imgRef.current)
    }

    return () => observer.disconnect()
  }, [])

  return (
    <div ref={imgRef} className={className}>
      {isInView && (
        <img
          src={src}
          alt={alt}
          onLoad={() => setIsLoaded(true)}
          className={cn(
            'transition-opacity duration-300',
            isLoaded ? 'opacity-100' : 'opacity-0'
          )}
        />
      )}
      {!isLoaded && <div className="skeleton" />}
    </div>
  )
}
```

---

## Memoization Strategies

### React Memoization Hooks

```tsx
import { memo, useMemo, useCallback, useDeferredValue } from 'react'

// Memoize expensive computations
function ExpensiveComponent({ items, filter }) {
  // Only recompute when items or filter changes
  const filteredItems = useMemo(() => {
    console.log('Computing filtered items...')
    return items
      .filter(item => item.name.includes(filter))
      .sort((a, b) => a.name.localeCompare(b.name))
  }, [items, filter])

  return (
    <ul>
      {filteredItems.map(item => (
        <li key={item.id}>{item.name}</li>
      ))}
    </ul>
  )
}

// Memoize callbacks to prevent child re-renders
function ParentComponent({ data }) {
  const [selected, setSelected] = useState<string | null>(null)

  // Stable callback reference
  const handleSelect = useCallback((id: string) => {
    setSelected(id)
  }, [])

  // Stable options reference
  const sortOptions = useMemo(() => ({
    key: 'name',
    direction: 'asc'
  }), [])

  return (
    <ChildComponent
      items={data}
      onSelect={handleSelect}
      sortOptions={sortOptions}
    />
  )
}

// Memoize child components
const ChildComponent = memo(function ChildComponent({
  items,
  onSelect,
  sortOptions
}: ChildProps) {
  return (
    <ul>
      {items.map(item => (
        <li key={item.id} onClick={() => onSelect(item.id)}>
          {item.name}
        </li>
      ))}
    </ul>
  )
})

// Deferred values for smooth input
function SearchList({ items }) {
  const [query, setQuery] = useState('')
  const deferredQuery = useDeferredValue(query)

  // Show stale results while computing
  const isStale = query !== deferredQuery

  const filteredItems = useMemo(() => {
    return items.filter(item =>
      item.name.toLowerCase().includes(deferredQuery.toLowerCase())
    )
  }, [items, deferredQuery])

  return (
    <div>
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search..."
      />

      <ul className={isStale ? 'opacity-50' : ''}>
        {filteredItems.map(item => (
          <li key={item.id}>{item.name}</li>
        ))}
      </ul>
    </div>
  )
}
```

### Custom Memoization Hook

```tsx
// hooks/useMemoizedCallback.ts
import { useRef, useCallback } from 'react'

export function useMemoizedCallback<T extends (...args: any[]) => any>(
  callback: T
): T {
  const ref = useRef(callback)

  // Update ref on each render
  ref.current = callback

  // Return stable function reference
  return useCallback((...args: Parameters<T>) => {
    return ref.current(...args)
  }, []) as T
}

// Usage - no need to specify dependencies
function Component({ onSubmit }) {
  const [value, setValue] = useState('')

  const handleSubmit = useMemoizedCallback(() => {
    onSubmit(value) // Always has access to latest value
  })

  return <button onClick={handleSubmit}>Submit</button>
}
```

---

## Bundle Optimization

### Vite Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],

  build: {
    target: 'es2022',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
    rollupOptions: {
      output: {
        manualChunks: {
          // Vendor chunks
          'vendor-react': ['react', 'react-dom'],
          'vendor-router': ['react-router-dom'],
          'vendor-ui': [
            '@radix-ui/react-dialog',
            '@radix-ui/react-dropdown-menu',
            '@radix-ui/react-tabs',
          ],
          'vendor-charts': ['recharts'],
        },
      },
    },
    chunkSizeWarningLimit: 500,
  },

  optimizeDeps: {
    include: ['react', 'react-dom'],
  },
})
```

### Tree Shaking Best Practices

```typescript
// Prefer named imports for tree shaking
// Good - tree shakeable
import { Button, Input } from '@/components/ui'

// Avoid - imports entire module
import * as UI from '@/components/ui'

// Barrel exports that preserve tree shaking
// components/ui/index.ts
export { Button } from './Button'
export { Input } from './Input'
export { Card } from './Card'

// Avoid default exports in library code
// Bad
export default Button

// Good
export { Button }
```

---

## Runtime Performance

### Virtualization for Long Lists

```tsx
import { useVirtualizer } from '@tanstack/react-virtual'

function VirtualList({ items }) {
  const parentRef = useRef<HTMLDivElement>(null)

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50,
    overscan: 5,
  })

  return (
    <div
      ref={parentRef}
      className="h-96 overflow-auto"
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            {items[virtualItem.index].name}
          </div>
        ))}
      </div>
    </div>
  )
}
```

### Debouncing and Throttling

```typescript
// hooks/useDebounce.ts
import { useState, useEffect } from 'react'

export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value)

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => clearTimeout(timer)
  }, [value, delay])

  return debouncedValue
}

// Usage
function SearchInput({ onSearch }) {
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebounce(query, 300)

  useEffect(() => {
    if (debouncedQuery) {
      onSearch(debouncedQuery)
    }
  }, [debouncedQuery, onSearch])

  return (
    <input
      value={query}
      onChange={(e) => setQuery(e.target.value)}
      placeholder="Search..."
    />
  )
}
```

---

## Performance Monitoring

### Web Vitals Tracking

```tsx
// lib/vitals.ts
import { onCLS, onFID, onLCP, onFCP, onTTFB } from 'web-vitals'

type Metric = {
  name: string
  value: number
  rating: 'good' | 'needs-improvement' | 'poor'
  id: string
}

function sendToAnalytics(metric: Metric) {
  // Send to your analytics service
  fetch('/api/analytics', {
    method: 'POST',
    body: JSON.stringify(metric),
  })
}

export function initVitals() {
  onCLS(sendToAnalytics)
  onFID(sendToAnalytics)
  onLCP(sendToAnalytics)
  onFCP(sendToAnalytics)
  onTTFB(sendToAnalytics)
}

// app/layout.tsx
import { initVitals } from '@/lib/vitals'

export default function RootLayout({ children }) {
  useEffect(() => {
    initVitals()
  }, [])

  return <html>{children}</html>
}
```

---

## Best Practices Summary

Code Splitting:
- Split by route automatically (Next.js App Router)
- Lazy load heavy components
- Preload on hover for faster navigation

Images:
- Use Next.js Image component
- Implement placeholder blur
- Set appropriate sizes attribute

Memoization:
- Use useMemo for expensive computations
- Use useCallback for stable callbacks
- Use memo for pure components
- Use useDeferredValue for smooth input

Bundle:
- Configure manual chunks
- Prefer named imports
- Monitor bundle size

Runtime:
- Virtualize long lists
- Debounce expensive operations
- Track Web Vitals

---

Version: 2.0.0
Last Updated: 2026-01-06
