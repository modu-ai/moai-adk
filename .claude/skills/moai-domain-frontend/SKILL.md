---
name: moai-domain-frontend
version: 4.0.0
created: 2025-11-12
updated: 2025-11-12
status: active
tier: domain
description: "Enterprise-grade frontend architecture expertise with React 19, Next.js 15, Vue 3.5, modern state management (Zustand, Pinia, Jotai), performance optimization, and accessibility compliance; activates for Server Components, App Router, Composition API, and cutting-edge 2025 web development patterns. Enhanced with Context7 MCP for always-current documentation."
allowed-tools: "Read, Bash, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "frontend-expert"
secondary-agents: [alfred, qa-validator, doc-syncer]
keywords: [frontend, react, vue, nextjs, typescript, components, state-management, performance]
tags: [domain-expert, 2025-ready, enterprise]
orchestration:
  can_resume: true
  typical_chain_position: "middle"
  depends_on: []
---

# moai-domain-frontend

**Enterprise Frontend Architecture — 2025 Edition**

> **Primary Agent**: frontend-expert  
> **Secondary Agents**: alfred, qa-validator, doc-syncer  
> **Version**: 4.0.0 Enterprise  
> **Keywords**: frontend, react, vue, nextjs, typescript, components, state-management, performance

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Purpose**: Enterprise-grade frontend architecture expertise covering React 19 (Server Components, Actions, Compiler), Next.js 15 (App Router, Server Actions, Turbopack), Vue 3.5 (Composition API, Suspense, TypeScript), modern state management (Zustand 5.x, Pinia 3.x, Jotai 3.x), performance optimization, and WCAG 2.2 accessibility compliance. Always current via Context7 MCP integration.

**When to Use:**
- ✅ Building React 19 applications with Server Components and Actions
- ✅ Next.js 15 App Router with parallel routes and streaming
- ✅ Vue 3.5 Composition API with TypeScript and reactive transforms
- ✅ Modern state management (Zustand, Pinia, Jotai) implementation
- ✅ Performance optimization (React Compiler, code splitting, Suspense)
- ✅ Accessibility-first component architecture (ARIA, keyboard navigation)

**Quick Start Pattern:**

```typescript
// React 19 Server Component with Server Action
// app/products/page.tsx
import { Suspense } from 'react'
import ProductList from './ProductList'

async function getProducts() {
  const res = await fetch('https://api.example.com/products', { 
    cache: 'no-store' 
  })
  return res.json()
}

export default async function ProductsPage() {
  const products = await getProducts()
  
  return (
    <Suspense fallback={<ProductsSkeleton />}>
      <ProductList products={products} />
    </Suspense>
  )
}

// Server Action
'use server'
async function addToCart(formData: FormData) {
  const productId = formData.get('productId')
  // Server-side cart logic
  await db.cart.create({ productId })
  revalidatePath('/cart')
}
```

---

### Level 2: Practical Implementation (Common Patterns)

## 🚀 React 19 Server Components & Actions (2025 Best Practices)

### **Pattern 1: Server Components for Data Fetching**

**Architecture Overview**:
```
React 19 Server-First Architecture:
├── Server Components (Default)
│   ├── Direct database/API access
│   ├── Zero client JavaScript
│   ├── Server-side rendering
│   └── Streaming with Suspense
├── Client Components ('use client')
│   ├── Interactivity (useState, useEffect)
│   ├── Event handlers
│   ├── Browser APIs
│   └── Minimal boundaries
└── Server Actions ('use server')
    ├── Form mutations
    ├── Database updates
    ├── Server-side validation
    └── Progressive enhancement
```

**Example: Product Dashboard (Server Component)**:
```tsx
// app/dashboard/page.tsx (Server Component - default in React 19)
import { Suspense } from 'react'
import { cache } from 'react'

// Cached server-side fetch
const getAnalytics = cache(async () => {
  const res = await fetch('https://api.analytics.com/data', {
    cache: 'no-store', // Fresh data on every request
    headers: { 'Authorization': `Bearer ${process.env.API_KEY}` }
  })
  return res.json()
})

async function AnalyticsDashboard() {
  const analytics = await getAnalytics()
  
  return (
    <div className="dashboard">
      <h1>Sales Analytics</h1>
      <div className="metrics">
        <MetricCard title="Revenue" value={analytics.revenue} />
        <MetricCard title="Users" value={analytics.users} />
        <MetricCard title="Growth" value={analytics.growth} />
      </div>
    </div>
  )
}

export default function DashboardPage() {
  return (
    <Suspense fallback={<AnalyticsSkeleton />}>
      <AnalyticsDashboard />
    </Suspense>
  )
}
```

### **Pattern 2: Server Actions for Mutations**

**Best Practice**: Use Server Actions for form submissions and data mutations, eliminating traditional API routes.

```tsx
// app/products/actions.ts
'use server'

import { revalidatePath } from 'next/cache'
import { z } from 'zod'

const productSchema = z.object({
  name: z.string().min(3),
  price: z.number().positive(),
  category: z.string()
})

export async function createProduct(formData: FormData) {
  // Server-side validation
  const data = {
    name: formData.get('name'),
    price: Number(formData.get('price')),
    category: formData.get('category')
  }
  
  const validated = productSchema.parse(data)
  
  // Direct database access (server-only)
  const product = await db.products.create({
    data: validated
  })
  
  // Revalidate cache
  revalidatePath('/products')
  
  return { success: true, product }
}

export async function deleteProduct(productId: string) {
  await db.products.delete({
    where: { id: productId }
  })
  
  revalidatePath('/products')
  return { success: true }
}
```

**Client Component Usage**:
```tsx
// app/products/ProductForm.tsx
'use client'

import { useFormState } from 'react'
import { createProduct } from './actions'

export function ProductForm() {
  const [state, formAction] = useFormState(createProduct, null)
  
  return (
    <form action={formAction}>
      <input name="name" required />
      <input name="price" type="number" required />
      <select name="category">
        <option value="electronics">Electronics</option>
        <option value="clothing">Clothing</option>
      </select>
      <button type="submit">Create Product</button>
      {state?.success && <p>Product created!</p>}
    </form>
  )
}
```

### **Pattern 3: React Compiler Optimization (Automatic Memoization)**

React 19's compiler automatically optimizes components—no manual `useMemo`, `useCallback`, or `React.memo` needed.

```tsx
// Before (React 18): Manual optimization
function ProductList({ products, onSelect }) {
  const filteredProducts = useMemo(
    () => products.filter(p => p.inStock),
    [products]
  )
  
  const handleSelect = useCallback(
    (id) => onSelect(id),
    [onSelect]
  )
  
  return <div>{/* ... */}</div>
}

// After (React 19): Compiler handles optimization automatically
function ProductList({ products, onSelect }) {
  const filteredProducts = products.filter(p => p.inStock)
  
  const handleSelect = (id) => onSelect(id)
  
  return <div>{/* ... */}</div>
}
```

---

## 🌐 Next.js 15 App Router Patterns (2025)

### **Pattern 4: Parallel Routes & Streaming**

**Architecture**:
```
app/
├── @analytics/
│   ├── loading.tsx
│   └── page.tsx
├── @feed/
│   ├── loading.tsx
│   └── page.tsx
├── layout.tsx  // Receives slots as props
└── page.tsx
```

**Implementation**:
```tsx
// app/layout.tsx
export default function DashboardLayout({
  children,
  analytics,
  feed
}: {
  children: React.ReactNode
  analytics: React.ReactNode
  feed: React.ReactNode
}) {
  return (
    <div className="dashboard-layout">
      <aside>{analytics}</aside>
      <main>{children}</main>
      <aside>{feed}</aside>
    </div>
  )
}

// app/@analytics/page.tsx (Parallel Route)
import { Suspense } from 'react'

async function getAnalytics() {
  const res = await fetch('https://api.analytics.com/stats')
  return res.json()
}

export default async function AnalyticsSlot() {
  const data = await getAnalytics()
  
  return (
    <Suspense fallback={<AnalyticsSkeleton />}>
      <AnalyticsPanel data={data} />
    </Suspense>
  )
}
```

### **Pattern 5: Server Actions with Progressive Enhancement**

```tsx
// app/products/[id]/page.tsx
import { updateProduct } from './actions'

export default async function ProductPage({ params }: { params: { id: string } }) {
  const product = await getProduct(params.id)
  
  return (
    <form action={updateProduct}>
      <input type="hidden" name="id" value={product.id} />
      <input name="name" defaultValue={product.name} />
      <input name="price" type="number" defaultValue={product.price} />
      <button type="submit">Update</button>
    </form>
  )
}

// Works without JavaScript (progressive enhancement)
// With JavaScript: optimistic updates via useOptimistic
```

### **Pattern 6: Next.js 15 Data Fetching (Cached vs Dynamic)**

```tsx
// Static data (cached by default)
async function getStaticProducts() {
  const res = await fetch('https://api.example.com/products', {
    cache: 'force-cache' // Default in Next.js 15
  })
  return res.json()
}

// Dynamic data (no cache)
async function getDynamicData() {
  const res = await fetch('https://api.example.com/live', {
    cache: 'no-store'
  })
  return res.json()
}

// Revalidated data (ISR)
async function getRevalidatedData() {
  const res = await fetch('https://api.example.com/products', {
    next: { revalidate: 60 } // Revalidate every 60 seconds
  })
  return res.json()
}
```

---

## 🎨 Vue 3.5 Composition API Patterns (2025)

### **Pattern 7: Vue 3.5 with TypeScript & Composition API**

```vue
<script setup lang="ts">
import { ref, computed, watch } from 'vue'

interface Product {
  id: string
  name: string
  price: number
  inStock: boolean
}

// Type-safe reactive state
const products = ref<Product[]>([])
const searchQuery = ref('')

// Computed with type inference
const filteredProducts = computed(() => 
  products.value.filter(p => 
    p.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
)

// Watcher with TypeScript
watch(searchQuery, async (newQuery) => {
  if (newQuery.length >= 3) {
    const results = await searchProducts(newQuery)
    products.value = results
  }
})

async function searchProducts(query: string): Promise<Product[]> {
  const res = await fetch(`/api/products?q=${query}`)
  return res.json()
}
</script>

<template>
  <div class="product-search">
    <input 
      v-model="searchQuery" 
      placeholder="Search products..."
      type="text"
    />
    <ul>
      <li v-for="product in filteredProducts" :key="product.id">
        {{ product.name }} - ${{ product.price }}
      </li>
    </ul>
  </div>
</template>
```

### **Pattern 8: Vue 3.5 Suspense & Async Components**

```vue
<script setup lang="ts">
import { ref } from 'vue'

// Async component with Suspense
const products = await fetchProducts() // Top-level await in setup

async function fetchProducts() {
  const res = await fetch('/api/products')
  return res.json()
}
</script>

<template>
  <Suspense>
    <template #default>
      <ProductList :products="products" />
    </template>
    <template #fallback>
      <ProductSkeleton />
    </template>
  </Suspense>
</template>
```

### **Pattern 9: Vue 3.5 Composables (Reusable Logic)**

```typescript
// composables/useProducts.ts
import { ref, computed } from 'vue'
import type { Ref } from 'vue'

interface Product {
  id: string
  name: string
  price: number
}

export function useProducts() {
  const products = ref<Product[]>([])
  const loading = ref(false)
  const error = ref<Error | null>(null)
  
  const totalValue = computed(() => 
    products.value.reduce((sum, p) => sum + p.price, 0)
  )
  
  async function fetchProducts() {
    loading.value = true
    error.value = null
    
    try {
      const res = await fetch('/api/products')
      products.value = await res.json()
    } catch (e) {
      error.value = e as Error
    } finally {
      loading.value = false
    }
  }
  
  return {
    products,
    loading,
    error,
    totalValue,
    fetchProducts
  }
}

// Usage in component
<script setup lang="ts">
import { onMounted } from 'vue'
import { useProducts } from '@/composables/useProducts'

const { products, loading, totalValue, fetchProducts } = useProducts()

onMounted(() => {
  fetchProducts()
})
</script>
```

---

## 🗂️ Modern State Management (2025)

### **Pattern 10: Zustand (Lightweight Global State)**

**Best for**: Medium-sized apps, simple global state, performance-critical scenarios.

```typescript
// store/productStore.ts
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'

interface Product {
  id: string
  name: string
  price: number
}

interface ProductStore {
  products: Product[]
  cart: Product[]
  addToCart: (product: Product) => void
  removeFromCart: (id: string) => void
  fetchProducts: () => Promise<void>
}

export const useProductStore = create<ProductStore>()(
  devtools(
    persist(
      (set) => ({
        products: [],
        cart: [],
        
        addToCart: (product) => 
          set((state) => ({ 
            cart: [...state.cart, product] 
          })),
        
        removeFromCart: (id) =>
          set((state) => ({
            cart: state.cart.filter(p => p.id !== id)
          })),
        
        fetchProducts: async () => {
          const res = await fetch('/api/products')
          const products = await res.json()
          set({ products })
        }
      }),
      { name: 'product-store' }
    )
  )
)

// Component usage
function ShoppingCart() {
  const cart = useProductStore(state => state.cart)
  const removeFromCart = useProductStore(state => state.removeFromCart)
  
  return (
    <ul>
      {cart.map(product => (
        <li key={product.id}>
          {product.name}
          <button onClick={() => removeFromCart(product.id)}>
            Remove
          </button>
        </li>
      ))}
    </ul>
  )
}
```

### **Pattern 11: Jotai (Atomic State Management)**

**Best for**: Complex state relationships, fine-grained reactivity, minimal re-renders.

```typescript
// atoms/productAtoms.ts
import { atom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'

interface Product {
  id: string
  name: string
  price: number
}

// Primitive atoms
export const productsAtom = atom<Product[]>([])
export const searchQueryAtom = atom('')

// Derived atom (computed)
export const filteredProductsAtom = atom((get) => {
  const products = get(productsAtom)
  const query = get(searchQueryAtom)
  
  return products.filter(p => 
    p.name.toLowerCase().includes(query.toLowerCase())
  )
})

// Async atom
export const productsFromAPIAtom = atom(async () => {
  const res = await fetch('/api/products')
  return res.json()
})

// Writable atom with side effects
export const addToCartAtom = atom(
  null,
  (get, set, product: Product) => {
    const cart = get(cartAtom)
    set(cartAtom, [...cart, product])
    
    // Side effect: analytics
    analytics.track('add_to_cart', { productId: product.id })
  }
)

// Persisted atom
export const cartAtom = atomWithStorage<Product[]>('cart', [])

// Component usage
import { useAtom, useAtomValue } from 'jotai'

function ProductSearch() {
  const [query, setQuery] = useAtom(searchQueryAtom)
  const filteredProducts = useAtomValue(filteredProductsAtom)
  
  return (
    <div>
      <input 
        value={query} 
        onChange={(e) => setQuery(e.target.value)} 
      />
      <ProductList products={filteredProducts} />
    </div>
  )
}
```

### **Pattern 12: Pinia (Vue State Management)**

```typescript
// stores/productStore.ts (Pinia 3.x for Vue 3.5)
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useProductStore = defineStore('products', () => {
  // State
  const products = ref<Product[]>([])
  const cart = ref<Product[]>([])
  
  // Getters (computed)
  const totalItems = computed(() => cart.value.length)
  const totalPrice = computed(() => 
    cart.value.reduce((sum, p) => sum + p.price, 0)
  )
  
  // Actions
  async function fetchProducts() {
    const res = await fetch('/api/products')
    products.value = await res.json()
  }
  
  function addToCart(product: Product) {
    cart.value.push(product)
  }
  
  function removeFromCart(id: string) {
    const index = cart.value.findIndex(p => p.id === id)
    if (index !== -1) {
      cart.value.splice(index, 1)
    }
  }
  
  return {
    products,
    cart,
    totalItems,
    totalPrice,
    fetchProducts,
    addToCart,
    removeFromCart
  }
})

// Component usage
<script setup lang="ts">
import { useProductStore } from '@/stores/productStore'

const store = useProductStore()
</script>

<template>
  <div>
    <p>Cart items: {{ store.totalItems }}</p>
    <p>Total: ${{ store.totalPrice }}</p>
    <button @click="store.fetchProducts()">
      Load Products
    </button>
  </div>
</template>
```

---

## ⚡ Performance Optimization Patterns (2025)

### **Pattern 13: Code Splitting & Lazy Loading**

```tsx
// React lazy loading
import { lazy, Suspense } from 'react'

const HeavyComponent = lazy(() => import('./HeavyComponent'))

function App() {
  return (
    <Suspense fallback={<Spinner />}>
      <HeavyComponent />
    </Suspense>
  )
}

// Next.js dynamic imports
import dynamic from 'next/dynamic'

const DynamicMap = dynamic(() => import('./Map'), {
  loading: () => <p>Loading map...</p>,
  ssr: false // Disable SSR for client-only components
})

// Vue async components
import { defineAsyncComponent } from 'vue'

const AsyncChart = defineAsyncComponent(() => 
  import('./Chart.vue')
)
```

### **Pattern 14: React Suspense for Data Fetching**

```tsx
// React 19 Suspense pattern
import { Suspense, use } from 'react'

// Resource pattern (fetches data)
const productsResource = fetch('/api/products').then(r => r.json())

function ProductList() {
  const products = use(productsResource) // Suspends until resolved
  
  return (
    <ul>
      {products.map(p => (
        <li key={p.id}>{p.name}</li>
      ))}
    </ul>
  )
}

export default function App() {
  return (
    <Suspense fallback={<ProductsSkeleton />}>
      <ProductList />
    </Suspense>
  )
}
```

---

## ♿ Accessibility Patterns (WCAG 2.2)

### **Pattern 15: Accessible Form with ARIA**

```tsx
function AccessibleForm() {
  const [errors, setErrors] = useState<Record<string, string>>({})
  
  return (
    <form aria-labelledby="form-title">
      <h2 id="form-title">Product Information</h2>
      
      <div>
        <label htmlFor="product-name">
          Product Name <span aria-label="required">*</span>
        </label>
        <input
          id="product-name"
          type="text"
          required
          aria-required="true"
          aria-invalid={!!errors.name}
          aria-describedby={errors.name ? "name-error" : undefined}
        />
        {errors.name && (
          <p id="name-error" role="alert" className="error">
            {errors.name}
          </p>
        )}
      </div>
      
      <button type="submit" aria-busy={isSubmitting}>
        {isSubmitting ? 'Saving...' : 'Save Product'}
      </button>
    </form>
  )
}
```

### **Pattern 16: Keyboard Navigation**

```tsx
function AccessibleDropdown() {
  const [isOpen, setIsOpen] = useState(false)
  const [selected, setSelected] = useState<number>(0)
  const itemsRef = useRef<HTMLButtonElement[]>([])
  
  const handleKeyDown = (e: KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelected(prev => (prev + 1) % items.length)
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelected(prev => (prev - 1 + items.length) % items.length)
        break
      case 'Enter':
      case ' ':
        e.preventDefault()
        handleSelect(items[selected])
        break
      case 'Escape':
        setIsOpen(false)
        break
    }
  }
  
  return (
    <div onKeyDown={handleKeyDown}>
      <button
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        onClick={() => setIsOpen(!isOpen)}
      >
        {selectedItem.name}
      </button>
      
      {isOpen && (
        <ul role="listbox" aria-activedescendant={`item-${selected}`}>
          {items.map((item, index) => (
            <li
              key={item.id}
              id={`item-${index}`}
              role="option"
              aria-selected={index === selected}
              ref={(el) => itemsRef.current[index] = el!}
            >
              {item.name}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
```

---

### Level 3: Advanced Patterns (Expert Reference)

## 🎯 Advanced Patterns

### **Pattern 17: React Server Components with Streaming**

```tsx
// app/dashboard/page.tsx
import { Suspense } from 'react'

async function SlowComponent() {
  await new Promise(resolve => setTimeout(resolve, 3000))
  return <div>Slow data loaded</div>
}

async function FastComponent() {
  await new Promise(resolve => setTimeout(resolve, 500))
  return <div>Fast data loaded</div>
}

export default function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>
      
      {/* Fast component streams first */}
      <Suspense fallback={<Skeleton />}>
        <FastComponent />
      </Suspense>
      
      {/* Slow component streams later */}
      <Suspense fallback={<Skeleton />}>
        <SlowComponent />
      </Suspense>
    </div>
  )
}
```

### **Pattern 18: Next.js 15 Intercepting Routes (Modals)**

```
app/
├── photos/
│   ├── [id]/
│   │   └── page.tsx        # Full photo page
│   └── page.tsx            # Photo grid
├── @modal/
│   └── (.)photos/
│       └── [id]/
│           └── page.tsx    # Modal view (intercepts)
└── layout.tsx
```

```tsx
// app/layout.tsx
export default function Layout({
  children,
  modal
}: {
  children: React.ReactNode
  modal: React.ReactNode
}) {
  return (
    <>
      {children}
      {modal}
    </>
  )
}

// app/@modal/(.)photos/[id]/page.tsx (Intercepting route)
export default function PhotoModal({ params }: { params: { id: string } }) {
  return (
    <dialog open>
      <PhotoDetail id={params.id} />
    </dialog>
  )
}
```

### **Pattern 19: Vue 3.5 Teleport for Modals**

```vue
<script setup lang="ts">
import { ref } from 'vue'

const isOpen = ref(false)
</script>

<template>
  <button @click="isOpen = true">Open Modal</button>
  
  <!-- Teleport to body (escapes component hierarchy) -->
  <Teleport to="body">
    <div v-if="isOpen" class="modal-overlay" @click="isOpen = false">
      <div class="modal-content" @click.stop>
        <h2>Modal Title</h2>
        <p>Modal content here</p>
        <button @click="isOpen = false">Close</button>
      </div>
    </div>
  </Teleport>
</template>
```

### **Pattern 20: Advanced TypeScript Patterns**

```typescript
// Generic component with constraints
function DataTable<T extends { id: string }>(props: {
  data: T[]
  columns: Array<keyof T>
  onRowClick: (row: T) => void
}) {
  return (
    <table>
      <thead>
        <tr>
          {props.columns.map(col => (
            <th key={String(col)}>{String(col)}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {props.data.map(row => (
          <tr key={row.id} onClick={() => props.onRowClick(row)}>
            {props.columns.map(col => (
              <td key={String(col)}>{String(row[col])}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}

// Usage with type inference
interface Product {
  id: string
  name: string
  price: number
}

<DataTable
  data={products}
  columns={['name', 'price']}  // Type-safe: only Product keys allowed
  onRowClick={(product) => console.log(product.name)}
/>
```

---

## 🎯 Best Practices Checklist

**Must-Have (2025 Standards):**
- ✅ Use React 19 Server Components for data-heavy pages
- ✅ Implement Server Actions for form submissions (eliminate API routes)
- ✅ Apply React Compiler (avoid manual useMemo/useCallback)
- ✅ Use Next.js 15 App Router (not Pages Router)
- ✅ Adopt Vue 3.5 Composition API + TypeScript (not Options API)
- ✅ Implement code splitting and lazy loading
- ✅ Use Suspense boundaries for streaming UI
- ✅ Optimize Core Web Vitals (LCP < 2.5s, CLS < 0.1, INP < 200ms)

**State Management:**
- ✅ Zustand for simple global state (medium apps)
- ✅ Jotai for complex atomic state (fine-grained reactivity)
- ✅ Pinia for Vue applications (official state management)
- ✅ Context API only for simple, infrequent updates

**Accessibility (WCAG 2.2):**
- ♿ Use semantic HTML (button, nav, main, article)
- ♿ Provide ARIA labels for dynamic content
- ♿ Ensure keyboard navigation (Tab, Enter, Escape, Arrow keys)
- ♿ Test with screen readers (NVDA, JAWS, VoiceOver)
- ♿ Maintain color contrast ratio ≥ 4.5:1 (WCAG AA)

**Security:**
- 🔒 Validate all input server-side (Zod, Yup)
- 🔒 Use Server Actions for sensitive operations
- 🔒 Implement CSP (Content Security Policy)
- 🔒 Sanitize user-generated content (DOMPurify)
- 🔒 Avoid exposing API keys in client code

---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with React, Next.js, or Vue
- Need latest framework documentation (React 19, Next.js 15, Vue 3.5)
- Verifying API changes and deprecations
- Exploring new features (Server Components, App Router, Composition API)

**Example Usage:**

```python
# Fetch latest React 19 documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()

# React 19 Server Components
react_docs = await helper.get_docs(
    library_id="/facebook/react/v19_2_0",
    topic="Server Components, Actions, useFormState",
    tokens=8000
)

# Next.js 15 App Router
nextjs_docs = await helper.get_docs(
    library_id="/vercel/next.js/v15.1.8",
    topic="App Router, Server Actions, Streaming",
    tokens=8000
)

# Vue 3.5 Composition API
vue_docs = await helper.get_docs(
    library_id="/vuejs/docs",
    topic="Composition API, Suspense, TypeScript",
    tokens=6000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| React 19 | `/facebook/react/v19_2_0` | Server Components, Actions, Compiler |
| Next.js 15 | `/vercel/next.js/v15.1.8` | App Router, Server Actions, Turbopack |
| Vue 3 | `/vuejs/docs` | Composition API, Suspense, TypeScript |
| React Router | `/remix-run/react-router` | Client-side routing patterns |
| VueUse | `/vueuse/vueuse` | Vue composables library |

---

## 📊 Decision Tree

**When to use moai-domain-frontend:**

```
Start: Need modern frontend?
  ├─ React/Next.js project?
  │   ├─ YES → Use React 19 + Next.js 15 patterns
  │   │         (Server Components, App Router, Server Actions)
  │   └─ NO → Vue project?
  │           ├─ YES → Use Vue 3.5 Composition API patterns
  │           └─ NO → Consider framework selection
  │
  ├─ State management needed?
  │   ├─ Simple global state → Zustand
  │   ├─ Complex atomic state → Jotai
  │   ├─ Vue application → Pinia
  │   └─ Rare updates → Context API
  │
  ├─ Performance critical?
  │   ├─ YES → Server Components + Streaming + Code splitting
  │   └─ NO → Standard patterns
  │
  └─ Accessibility required?
      ├─ YES → ARIA patterns + keyboard navigation
      └─ NO → Still implement basics (semantic HTML)
```

---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("moai-foundation-typescript") – Type-safe development foundation
- Skill("moai-foundation-git") – Version control for component libraries

**Complementary Skills:**
- Skill("moai-domain-backend") – Full-stack integration patterns
- Skill("moai-domain-testing") – Component testing strategies (Vitest, Playwright)
- Skill("moai-domain-ui") – Design system integration
- Skill("moai-domain-performance") – Core Web Vitals optimization

**Next Steps:**
- Skill("moai-domain-deployment") – Frontend deployment strategies
- Skill("moai-domain-monitoring") – Performance monitoring & analytics

---

## 📚 Official References

### **React 19**
- [React 19 Official Docs](https://react.dev/)
- [React Server Components](https://react.dev/reference/rsc/server-components)
- [React Compiler](https://react.dev/learn/react-compiler)

### **Next.js 15**
- [Next.js 15 Docs](https://nextjs.org/docs)
- [App Router](https://nextjs.org/docs/app)
- [Server Actions](https://nextjs.org/docs/app/building-your-application/data-fetching/server-actions-and-mutations)

### **Vue 3.5**
- [Vue 3 Docs](https://vuejs.org/)
- [Composition API](https://vuejs.org/guide/typescript/composition-api)
- [TypeScript with Vue](https://vuejs.org/guide/typescript/overview)

### **State Management**
- [Zustand](https://zustand-demo.pmnd.rs/)
- [Jotai](https://jotai.org/)
- [Pinia](https://pinia.vuejs.org/)

### **Testing**
- [Vitest](https://vitest.dev/)
- [Testing Library](https://testing-library.com/)
- [Playwright](https://playwright.dev/)

---

## 📈 Version History

**v4.0.0** (2025-11-12) — Enterprise Edition
- ✨ React 19 Server Components & Actions patterns
- ✨ Next.js 15 App Router with parallel routes
- ✨ Vue 3.5 Composition API + TypeScript
- ✨ Modern state management (Zustand, Jotai, Pinia)
- ✨ React Compiler optimization patterns
- ✨ 20+ comprehensive code examples
- ✨ WCAG 2.2 accessibility patterns
- ✨ Context7 MCP integration
- ✨ 2025 best practices
- 📏 Lines: 950+ (Enterprise standard)
- 📦 Size: 32KB

---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (frontend-expert)  
**Status**: ✅ Production Ready — Enterprise v4.0
