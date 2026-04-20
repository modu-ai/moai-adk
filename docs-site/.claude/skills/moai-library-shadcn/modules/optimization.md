# shadcn/ui Performance Optimization

## Bundle Size Optimization

### Code Splitting Strategy

Pattern: Lazy Loading Components

```typescript
import React from "react"

// Lazy load heavy components
const DataTableComponent = React.lazy(() =>
 import("@/components/data-table").then(mod => ({ default: mod.DataTable }))
)

export function Dashboard() {
 return (
 <React.Suspense fallback={<p>Loading...</p>}>
 <DataTableComponent />
 </React.Suspense>
 )
}
```

### CSS File Size Optimization

shadcn/ui uses Tailwind CSS. Optimize CSS output by:

1. Purging unused styles in `tailwind.config.ts`:

```typescript
export default {
 content: [
 "./app//*.{js,ts,jsx,tsx}",
 "./components//*.{js,ts,jsx,tsx}",
 ],
 // Only include styles used in templates
}
```

2. Tree-shaking unused components - Import only what you need:

```typescript
// GOOD: Import specific components
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"

// AVOID: Importing entire component library
import * as UI from "@/components/ui"
```

### Component Import Optimization

```typescript
// Optimized imports
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

function Form() {
 return (
 <>
 <Input placeholder="Name" />
 <Button>Submit</Button>
 </>
 )
}
```

## Rendering Optimization

### React.memo for Components

```typescript
import React from "react"
import { Card } from "@/components/ui/card"

interface ItemProps {
 id: string
 title: string
 description: string
}

// Prevent re-renders when props haven't changed
export const ListItem = React.memo(function ListItem({ id, title, description }: ItemProps) {
 return (
 <Card className="p-4">
 <h3>{title}</h3>
 <p>{description}</p>
 </Card>
 )
}, (prevProps, nextProps) => {
 // Custom comparison
 return prevProps.id === nextProps.id &&
 prevProps.title === nextProps.title &&
 prevProps.description === nextProps.description
})
```

### useMemo for Expensive Computations

```typescript
import React, { useMemo } from "react"
import { Button } from "@/components/ui/button"

interface DataTableProps {
 rows: any[]
 filter: string
}

export function DataTable({ rows, filter }: DataTableProps) {
 // Memoize filtered results
 const filteredRows = useMemo(() => {
 return rows.filter(row => row.name.includes(filter))
 }, [rows, filter])

 return (
 <div>
 {filteredRows.map(row => (
 <div key={row.id}>{row.name}</div>
 ))}
 </div>
 )
}
```

### useCallback for Event Handlers

```typescript
import React, { useCallback } from "react"
import { Button } from "@/components/ui/button"

export function Form() {
 const handleSubmit = useCallback((e: React.FormEvent) => {
 e.preventDefault()
 // Expensive operation
 }, [])

 return (
 <form onSubmit={handleSubmit}>
 <Button type="submit">Submit</Button>
 </form>
 )
}
```

## Token Efficiency

### CSS Variable Optimization

```typescript
// Efficient design token usage
const tokenConfig = {
 colors: {
 primary: "hsl(var(--primary))",
 secondary: "hsl(var(--secondary))",
 },
 spacing: {
 xs: "var(--spacing-xs)",
 sm: "var(--spacing-sm)",
 },
}

// Use in components
export function Button() {
 return (
 <button className="bg-[var(--primary)] px-[var(--spacing-sm)]">
 Click me
 </button>
 )
}
```

## Rendering Optimization

### Virtual Scrolling for Large Lists

```typescript
import React, { useMemo } from "react"
import { FixedSizeList } from "react-window"
import { Card } from "@/components/ui/card"

interface Item {
 id: string
 name: string
}

export function VirtualList({ items }: { items: Item[] }) {
 const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
 <div style={style}>
 <Card className="p-2">
 <p>{items[index].name}</p>
 </Card>
 </div>
 )

 return (
 <FixedSizeList
 height={600}
 itemCount={items.length}
 itemSize={80}
 width="100%"
 >
 {Row}
 </FixedSizeList>
 )
}
```

### Pagination Pattern

```typescript
import React from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"

export function PaginatedList() {
 const [page, setPage] = React.useState(1)
 const itemsPerPage = 10

 const startIndex = (page - 1) * itemsPerPage
 const endIndex = startIndex + itemsPerPage

 return (
 <div>
 <div className="space-y-2">
 {/* Render only items for current page */}
 </div>
 <div className="flex gap-2">
 <Button onClick={() => setPage(page - 1)} disabled={page === 1}>
 Previous
 </Button>
 <span>Page {page}</span>
 <Button onClick={() => setPage(page + 1)}>
 Next
 </Button>
 </div>
 </div>
 )
}
```

## Network Optimization

### Image Optimization with Next.js

```typescript
import Image from "next/image"

export function OptimizedImage() {
 return (
 <Image
 src="/image.jpg"
 alt="Optimized image"
 width={400}
 height={300}
 priority={false}
 loading="lazy"
 />
 )
}
```

## Best Practices Summary

1. Lazy Load Components - Use React.lazy for non-critical components
2. Memoize Selectively - Don't over-memoize; profile first
3. Tree-shake Imports - Import specific components, not entire libraries
4. Optimize CSS - Configure Tailwind content properly
5. Virtual Scroll Large Lists - Use react-window for 1000+ items
6. Cache Computations - Use useMemo and useCallback judiciously
7. Monitor Bundle - Use bundle analyzer to identify bottlenecks

---

Version: 4.0.0
Last Updated: 2025-11-22
Token Focus: CSS file size, JavaScript bundle size, rendering performance
