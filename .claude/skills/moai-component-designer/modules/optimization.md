# Component Optimization Patterns

## Re-render Optimization

### React.memo for Preventing Re-renders

```typescript
interface ItemProps {
  id: string
  title: string
  onDelete: (id: string) => void
}

// Prevent re-renders when props haven't changed
export const TodoItem = React.memo(function TodoItem({ id, title, onDelete }: ItemProps) {
  return (
    <li>
      <span>{title}</span>
      <button onClick={() => onDelete(id)}>Delete</button>
    </li>
  )
})

// Custom comparison for complex props
export const UserCard = React.memo(
  function UserCard({ user }: { user: User }) {
    return <div>{user.name}</div>
  },
  (prevProps, nextProps) => {
    return prevProps.user.id === nextProps.user.id
  }
)
```

### useMemo for Expensive Computations

```typescript
import React, { useMemo } from "react"

export function DataProcessor({ data }: { data: number[] }) {
  // Memoize expensive computation
  const sorted = useMemo(() => {
    console.log("Sorting...")
    return [...data].sort((a, b) => a - b)
  }, [data])

  const average = useMemo(() => {
    return sorted.reduce((a, b) => a + b, 0) / sorted.length
  }, [sorted])

  return <div>Average: {average}</div>
}
```

### useCallback for Event Handlers

```typescript
import React, { useCallback } from "react"

export function ListComponent({ items, onItemClick }: { items: Item[]; onItemClick: (id: string) => void }) {
  // Memoize handler to prevent child re-renders
  const handleClick = useCallback(
    (id: string) => {
      onItemClick(id)
    },
    [onItemClick]
  )

  return (
    <ul>
      {items.map((item) => (
        <ListItem key={item.id} id={item.id} onClick={handleClick} />
      ))}
    </ul>
  )
}
```

## Code Splitting

### Lazy Loading Components

```typescript
import React from "react"

// Lazy load heavy components
const HeavyComponent = React.lazy(() => import("./HeavyComponent"))
const DataTable = React.lazy(() => import("./DataTable"))

export function Dashboard() {
  return (
    <>
      <React.Suspense fallback={<div>Loading...</div>}>
        <HeavyComponent />
      </React.Suspense>
      <React.Suspense fallback={<div>Loading table...</div>}>
        <DataTable />
      </React.Suspense>
    </>
  )
}
```

### Route-Based Code Splitting

```typescript
import { lazy } from "react"
import { BrowserRouter, Routes, Route } from "react-router-dom"

const Home = lazy(() => import("./pages/Home"))
const Dashboard = lazy(() => import("./pages/Dashboard"))
const Profile = lazy(() => import("./pages/Profile"))

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<React.Suspense fallback={<div>Loading...</div>}><Home /></React.Suspense>} />
        <Route path="/dashboard" element={<React.Suspense fallback={<div>Loading...</div>}><Dashboard /></React.Suspense>} />
        <Route path="/profile" element={<React.Suspense fallback={<div>Loading...</div>}><Profile /></React.Suspense>} />
      </Routes>
    </BrowserRouter>
  )
}
```

## Virtual Scrolling

### For Large Lists

```typescript
import React from "react"
import { FixedSizeList } from "react-window"

interface Item {
  id: string
  name: string
}

export function VirtualList({ items }: { items: Item[] }) {
  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style} className="border p-2">
      {items[index].name}
    </div>
  )

  return (
    <FixedSizeList
      height={600}
      itemCount={items.length}
      itemSize={50}
      width="100%"
    >
      {Row}
    </FixedSizeList>
  )
}
```

## Bundle Size Optimization

### Tree-Shaking Unused Code

```typescript
// ✅ GOOD: Named exports (tree-shakeable)
export function Button() {}
export function Card() {}
export function Input() {}

// ❌ AVOID: Default exports of objects
export default {
  Button,
  Card,
  Input,
}

// ✅ GOOD: Specific imports
import { Button } from "@/components"

// ❌ AVOID: Wildcard imports
import * as Components from "@/components"
```

### Dynamic Imports

```typescript
// Load component only when needed
async function loadHeavyComponent() {
  const module = await import("./HeavyComponent")
  return module.default
}

// Or in React
const DynamicComponent = React.lazy(() =>
  import("./HeavyComponent").then((mod) => ({ default: mod.HeavyComponent }))
)
```

## CSS Optimization

### CSS-in-JS Efficiency

```typescript
// ✅ GOOD: Create styles once
const buttonStyles = {
  base: "px-4 py-2 rounded",
  primary: "bg-blue-500 text-white",
  secondary: "bg-gray-200 text-gray-900",
}

export function Button({ variant = "primary" }: { variant: "primary" | "secondary" }) {
  return (
    <button className={`${buttonStyles.base} ${buttonStyles[variant]}`}>
      Click me
    </button>
  )
}

// ❌ AVOID: Creating new objects every render
export function BadButton() {
  return (
    <button
      className="px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"
      style={{ padding: "1rem" }}
    >
      Click me
    </button>
  )
}
```

## Animation Optimization

### useTransition for Non-blocking Updates

```typescript
import React, { useTransition, useState } from "react"

export function SearchableList({ items }: { items: Item[] }) {
  const [query, setQuery] = useState("")
  const [isPending, startTransition] = useTransition()

  const filtered = query
    ? items.filter((item) => item.name.includes(query))
    : items

  return (
    <>
      <input
        value={query}
        onChange={(e) => {
          startTransition(() => {
            setQuery(e.target.value)
          })
        }}
        placeholder="Search..."
      />
      {isPending && <p>Searching...</p>}
      <ul>
        {filtered.map((item) => (
          <li key={item.id}>{item.name}</li>
        ))}
      </ul>
    </>
  )
}
```

## Performance Monitoring

### Profiling Components

```typescript
import React, { Profiler } from "react"

function onRenderCallback(
  id: string,
  phase: "mount" | "update",
  actualDuration: number,
  baseDuration: number,
  startTime: number,
  commitTime: number
) {
  console.log(`${id} (${phase}): ${actualDuration}ms`)
}

export function ProfiledApp() {
  return (
    <Profiler id="app" onRender={onRenderCallback}>
      <App />
    </Profiler>
  )
}
```

## Best Practices

1. **Memoize Strategically** - Only memoize expensive renders
2. **Use Virtual Scrolling** - For lists with 100+ items
3. **Code Split Routes** - Load features on demand
4. **Tree-shake Exports** - Use named exports
5. **Monitor Performance** - Use React Profiler and DevTools
6. **Optimize Images** - Use next/image or responsive images
7. **Bundle Analysis** - Analyze bundle size regularly

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Focus**: Rendering performance, bundle size, code splitting efficiency
