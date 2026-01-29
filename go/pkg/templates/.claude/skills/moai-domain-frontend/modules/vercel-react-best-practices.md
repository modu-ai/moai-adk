# Vercel React Best Practices

Comprehensive performance optimization guide for React and Next.js applications from Vercel Engineering. Contains 45 rules across 8 categories, prioritized by impact to guide automated refactoring and code generation.

---

## Rule Categories by Priority

| Priority | Category | Impact | Prefix |
|----------|----------|--------|--------|
| 1 | Eliminating Waterfalls | CRITICAL | `async-` |
| 2 | Bundle Size Optimization | CRITICAL | `bundle-` |
| 3 | Server-Side Performance | HIGH | `server-` |
| 4 | Client-Side Data Fetching | MEDIUM-HIGH | `client-` |
| 5 | Re-render Optimization | MEDIUM | `rerender-` |
| 6 | Rendering Performance | MEDIUM | `rendering-` |
| 7 | JavaScript Performance | LOW-MEDIUM | `js-` |
| 8 | Advanced Patterns | LOW | `advanced-` |

---

## 1. Eliminating Waterfalls (CRITICAL)

Waterfalls are the number one performance killer. Each sequential await adds full network latency. Eliminating them yields the largest gains.

### 1.1 Defer Await Until Needed

Move `await` operations into the branches where they're actually used to avoid blocking code paths that don't need them.

```tsx
// Incorrect: blocks both branches
async function handleRequest(userId: string, skipProcessing: boolean) {
  const userData = await fetchUserData(userId)

  if (skipProcessing) {
    return { skipped: true }
  }

  return processUserData(userData)
}

// Correct: only blocks when needed
async function handleRequest(userId: string, skipProcessing: boolean) {
  if (skipProcessing) {
    return { skipped: true }
  }

  const userData = await fetchUserData(userId)
  return processUserData(userData)
}
```

### 1.2 Promise.all() for Independent Operations

When async operations have no interdependencies, execute them concurrently using `Promise.all()`.

```tsx
// Incorrect: sequential execution, 3 round trips
const user = await fetchUser()
const posts = await fetchPosts()
const comments = await fetchComments()

// Correct: parallel execution, 1 round trip
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments()
])
```

### 1.3 Dependency-Based Parallelization

For operations with partial dependencies, use `better-all` to maximize parallelism.

```tsx
// Incorrect: profile waits for config unnecessarily
const [user, config] = await Promise.all([
  fetchUser(),
  fetchConfig()
])
const profile = await fetchProfile(user.id)

// Correct: config and profile run in parallel
import { all } from 'better-all'

const { user, config, profile } = await all({
  async user() { return fetchUser() },
  async config() { return fetchConfig() },
  async profile() {
    return fetchProfile((await this.$.user).id)
  }
})
```

### 1.4 Prevent Waterfall Chains in API Routes

In API routes and Server Actions, start independent operations immediately, even if you don't await them yet.

```tsx
// Incorrect: config waits for auth, data waits for both
export async function GET(request: Request) {
  const session = await auth()
  const config = await fetchConfig()
  const data = await fetchData(session.user.id)
  return Response.json({ data, config })
}

// Correct: auth and config start immediately
export async function GET(request: Request) {
  const sessionPromise = auth()
  const configPromise = fetchConfig()
  const session = await sessionPromise
  const [config, data] = await Promise.all([
    configPromise,
    fetchData(session.user.id)
  ])
  return Response.json({ data, config })
}
```

### 1.5 Strategic Suspense Boundaries

Use Suspense boundaries to show wrapper UI faster while data loads, rather than awaiting data in async components before returning JSX.

```tsx
// Incorrect: wrapper blocked by data fetching
async function Page() {
  const data = await fetchData()

  return (
    <div>
      <div>Sidebar</div>
      <div>Header</div>
      <div><DataDisplay data={data} /></div>
      <div>Footer</div>
    </div>
  )
}

// Correct: wrapper shows immediately, data streams in
function Page() {
  return (
    <div>
      <div>Sidebar</div>
      <div>Header</div>
      <div>
        <Suspense fallback={<Skeleton />}>
          <DataDisplay />
        </Suspense>
      </div>
      <div>Footer</div>
    </div>
  )
}

async function DataDisplay() {
  const data = await fetchData()
  return <div>{data.content}</div>
}
```

---

## 2. Bundle Size Optimization (CRITICAL)

Reducing initial bundle size improves Time to Interactive and Largest Contentful Paint.

### 2.1 Avoid Barrel File Imports

Import directly from source files instead of barrel files to avoid loading thousands of unused modules.

```tsx
// Incorrect: imports entire library
import { Check, X, Menu } from 'lucide-react'
// Loads 1,583 modules, takes ~2.8s extra in dev

// Correct: imports only what you need
import Check from 'lucide-react/dist/esm/icons/check'
import X from 'lucide-react/dist/esm/icons/x'
import Menu from 'lucide-react/dist/esm/icons/menu'
// Loads only 3 modules

// Alternative: Next.js 13.5+
// next.config.js
module.exports = {
  experimental: {
    optimizePackageImports: ['lucide-react', '@mui/material']
  }
}
```

### 2.2 Dynamic Imports for Heavy Components

Use `next/dynamic` to lazy-load large components not needed on initial render.

```tsx
// Incorrect: Monaco bundles with main chunk ~300KB
import { MonacoEditor } from './monaco-editor'

function CodePanel({ code }: { code: string }) {
  return <MonacoEditor value={code} />
}

// Correct: Monaco loads on demand
import dynamic from 'next/dynamic'

const MonacoEditor = dynamic(
  () => import('./monaco-editor'),
  {
    loading: () => <EditorSkeleton />,
    ssr: false
  }
)

function CodePanel({ code }: { code: string }) {
  return <MonacoEditor value={code} />
}
```

### 2.3 Defer Non-Critical Third-Party Libraries

Analytics, logging, and error tracking don't block user interaction. Load them after hydration.

```tsx
// Incorrect: blocks initial bundle
import { Analytics } from '@vercel/analytics/react'

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  )
}

// Correct: loads after hydration
import dynamic from 'next/dynamic'

const Analytics = dynamic(
  () => import('@vercel/analytics/react').then(m => m.Analytics),
  { ssr: false }
)

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  )
}
```

### 2.4 Conditional Module Loading

Load large data or modules only when a feature is activated.

```tsx
function AnimationPlayer({ enabled }: { enabled: boolean }) {
  const [frames, setFrames] = useState<Frame[] | null>(null)

  useEffect(() => {
    if (enabled && !frames && typeof window !== 'undefined') {
      import('./animation-frames.js')
        .then(mod => setFrames(mod.frames))
        .catch(() => setEnabled(false))
    }
  }, [enabled, frames])

  if (!frames) return <Skeleton />
  return <Canvas frames={frames} />
}
```

### 2.5 Preload Based on User Intent

Preload components on hover or focus for perceived speed improvement.

```tsx
import dynamic from 'next/dynamic'

const ChartComponent = dynamic(() => import('./Chart'))

function preloadChart() {
  import('./Chart')
}

function Navigation() {
  return (
    <Link
      href="/analytics"
      onMouseEnter={preloadChart}
      onFocus={preloadChart}
    >
      Analytics
    </Link>
  )
}
```

---

## 3. Server-Side Performance (HIGH)

### 3.1 Per-Request Deduplication with React.cache()

Use `React.cache()` for per-request deduplication of expensive operations.

```tsx
import { cache } from 'react'

const getData = cache(async (id: string) => {
  const response = await fetch(`https://api.example.com/data/${id}`)
  return response.json()
})

// Multiple calls in the same request are deduplicated
async function UserPage({ params }: { params: { id: string } }) {
  const user = await getData(params.id)
  const posts = await getData(params.id) // Uses cached result
  return <div>{/* ... */}</div>
}
```

### 3.2 Cross-Request LRU Caching

Use LRU cache for cross-request caching when appropriate.

```tsx
import { LRUCache } from 'lru-cache'

const cache = new LRUCache<string, Data>({
  max: 500,
  ttl: 1000 * 60 * 5, // 5 minutes
})

async function getCachedData(key: string): Promise<Data> {
  if (cache.has(key)) {
    return cache.get(key)!
  }

  const data = await fetchData(key)
  cache.set(key, data)
  return data
}
```

### 3.3 Minimize Serialization at RSC Boundaries

Minimize data passed to client components to reduce serialization overhead.

```tsx
// Incorrect: passes entire object to client
async function ServerComponent() {
  const user = await fetchUser()
  return <ClientComponent user={user} />
}

// Correct: passes only needed fields
async function ServerComponent() {
  const user = await fetchUser()
  const minimalUser = {
    id: user.id,
    name: user.name,
    avatar: user.avatar
  }
  return <ClientComponent user={minimalUser} />
}
```

### 3.4 Parallel Data Fetching with Component Composition

Restructure components to parallelize fetches rather than creating sequential chains.

```tsx
// Incorrect: sequential fetching
async function Page() {
  const user = await fetchUser()
  const posts = await fetchPosts(user.id)
  const comments = await fetchComments(posts.map(p => p.id))
  return <Dashboard user={user} posts={posts} comments={comments} />
}

// Correct: parallel fetching where possible
async function Page() {
  const [user, posts, comments] = await Promise.all([
    fetchUser(),
    fetchPosts(), // If posts doesn't strictly depend on user
    fetchComments() // If comments doesn't strictly depend on posts
  ])
  return <Dashboard user={user} posts={posts} comments={comments} />
}
```

### 3.5 Use after() for Non-Blocking Operations

Use `after()` for non-blocking operations that can run after response streaming starts.

```tsx
import { after } from 'next/server'

async function handler(request: Request) {
  const data = await fetchData()

  // Runs after response is sent
  after(() => {
    logAnalytics(data)
  })

  return Response.json(data)
}
```

---

## 4. Client-Side Data Fetching (MEDIUM-HIGH)

### 4.1 Deduplicate Global Event Listeners

Deduplicate global event listeners to prevent memory leaks and performance issues.

```tsx
// Incorrect: adds new listener on each render
function ChatComponent() {
  useEffect(() => {
    window.addEventListener('online', handleOnline)
    return () => window.removeEventListener('online', handleOnline)
  })
  return <div>Chat</div>
}

// Correct: single listener with dispatcher
const onlineListeners = new Set<() => void>()

function addOnlineListener(fn: () => void) {
  onlineListeners.add(fn)
}

function removeOnlineListener(fn: () => void) {
  onlineListeners.delete(fn)
}

if (typeof window !== 'undefined') {
  window.addEventListener('online', () => {
    onlineListeners.forEach(fn => fn())
  })
}

function ChatComponent() {
  useEffect(() => {
    const handler = () => console.log('online')
    addOnlineListener(handler)
    return () => removeOnlineListener(handler)
  })
  return <div>Chat</div>
}
```

### 4.2 Use SWR for Automatic Deduplication

Use SWR for automatic request deduplication and caching.

```tsx
import useSWR from 'swr'

const fetcher = (url: string) => fetch(url).then(r => r.json())

// Multiple components with the same key share the request
function UserSection() {
  const { data: user } = useSWR('/api/user', fetcher)
  return <div>{user?.name}</div>
}

function UserPosts() {
  const { data: user } = useSWR('/api/user', fetcher) // Deduplicated
  return <div>{user?.posts?.length} posts</div>
}
```

---

## 5. Re-render Optimization (MEDIUM)

### 5.1 Defer State Reads to Usage Point

Don't subscribe to state only used in callbacks.

```tsx
// Incorrect: re-renders on count changes
function Counter() {
  const [count, setCount] = useState(0)
  const handleClick = () => {
    console.log(count)
  }
  return <button onClick={handleClick}>Count: {count}</button>
}

// Correct: reads count only in render and callback
function Counter() {
  const [count, setCount] = useState(0)
  const handleClick = () => {
    setCount(c => {
      console.log(c)
      return c
    })
  }
  return <button onClick={handleClick}>Count: {count}</button>
}
```

### 5.2 Extract to Memoized Components

Extract expensive work into memoized components to enable early returns.

```tsx
// Incorrect: computes avatar even when loading
function Profile({ user, loading }: Props) {
  const avatar = useMemo(() => {
    const id = computeAvatarId(user)
    return <Avatar id={id} />
  }, [user])

  if (loading) return <Skeleton />
  return <div>{avatar}</div>
}

// Correct: skips computation when loading
const UserAvatar = memo(function UserAvatar({ user }: { user: User }) {
  const id = useMemo(() => computeAvatarId(user), [user])
  return <Avatar id={id} />
})

function Profile({ user, loading }: Props) {
  if (loading) return <Skeleton />
  return <div><UserAvatar user={user} /></div>
}
```

### 5.3 Narrow Effect Dependencies

Use primitive dependencies in effects to avoid unnecessary re-runs.

```tsx
// Incorrect: runs on any user change
function UserProfile({ user }: { user: User }) {
  useEffect(() => {
    trackEvent('profile_view', { userId: user.id })
  }, [user]) // Runs on every user property change

  return <div>{user.name}</div>
}

// Correct: runs only when userId changes
function UserProfile({ user }: { user: User }) {
  useEffect(() => {
    trackEvent('profile_view', { userId: user.id })
  }, [user.id]) // Runs only when id changes

  return <div>{user.name}</div>
}
```

### 5.4 Subscribe to Derived State

Subscribe to derived booleans, not raw values.

```tsx
// Incorrect: re-renders on any user change
function UserMenu({ user }: { user: User | null }) {
  if (!user) return <LoginButton />

  return (
    <Menu>
      <MenuItem>Welcome, {user.name}</MenuItem>
    </Menu>
  )
}

// Correct: re-renders only when auth state changes
function UserMenu({ user }: { user: User | null }) {
  const isLoggedIn = user !== null

  if (!isLoggedIn) return <LoginButton />

  return (
    <Menu>
      <MenuItem>Welcome, {user.name}</MenuItem>
    </Menu>
  )
}
```

### 5.5 Use Functional setState Updates

Use functional setState for stable callbacks.

```tsx
// Incorrect: callback captures old state
function Counter() {
  const [count, setCount] = useState(0)
  const increment = () => setCount(count + 1)
  return <button onClick={increment}>{count}</button>
}

// Correct: callback always gets latest state
function Counter() {
  const [count, setCount] = useState(0)
  const increment = () => setCount(c => c + 1)
  return <button onClick={increment}>{count}</button>
}
```

### 5.6 Use Lazy State Initialization

Pass function to useState for expensive values.

```tsx
// Incorrect: runs on every render
function List({ items }: { items: Item[] }) {
  const [sorted, setSorted] = useState(
    items.sort((a, b) => a.name.localeCompare(b.name))
  )
  return <ul>{sorted.map(item => <li key={item.id}>{item.name}</li>)}</ul>
}

// Correct: runs only on initialization
function List({ items }: { items: Item[] }) {
  const [sorted, setSorted] = useState(() =>
    items.sort((a, b) => a.name.localeCompare(b.name))
  )
  return <ul>{sorted.map(item => <li key={item.id}>{item.name}</li>)}</ul>
}
```

### 5.7 Use Transitions for Non-Urgent Updates

Use `startTransition` for non-urgent updates.

```tsx
import { startTransition, useState } from 'react'

function SearchInput() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<Results[]>([])

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value

    // Urgent: update input immediately
    setQuery(value)

    // Non-urgent: defer expensive filtering
    startTransition(() => {
      setResults(filterResults(value))
    })
  }

  return (
    <div>
      <input value={query} onChange={handleChange} />
      <ResultList items={results} />
    </div>
  )
}
```

---

## 6. Rendering Performance (MEDIUM)

### 6.1 Animate SVG Wrapper Instead of SVG Element

Animate div wrapper, not SVG element for better performance.

```tsx
// Incorrect: animating SVG element
function AnimatedIcon() {
  return (
    <svg className="animate-spin">
      <path d="..." />
    </svg>
  )
}

// Correct: animating wrapper div
function AnimatedIcon() {
  return (
    <div className="animate-spin">
      <svg>
        <path d="..." />
      </svg>
    </div>
  )
}
```

### 6.2 CSS content-visibility for Long Lists

Use `content-visibility` for long lists to improve rendering performance.

```tsx
function LongList({ items }: { items: Item[] }) {
  return (
    <div>
      {items.map(item => (
        <div
          key={item.id}
          style={{
            contentVisibility: 'auto',
            containIntrinsicSize: '0 100px'
          }}
        >
          <h3>{item.title}</h3>
          <p>{item.description}</p>
        </div>
      ))}
    </div>
  )
}
```

### 6.3 Hoist Static JSX Elements

Extract static JSX outside components to avoid recreating on each render.

```tsx
// Incorrect: creates new objects on each render
function Component() {
  return (
    <div>
      <header className="fixed top-0 left-0 right-0 z-50">
        <Logo />
        <Navigation />
      </header>
      <main>{/* content */}</main>
    </div>
  )
}

// Correct: static JSX hoisted outside
const header = (
  <header className="fixed top-0 left-0 right-0 z-50">
    <Logo />
    <Navigation />
  </header>
)

function Component() {
  return (
    <div>
      {header}
      <main>{/* content */}</main>
    </div>
  )
}
```

### 6.4 Optimize SVG Precision

Reduce SVG coordinate precision to reduce file size.

```tsx
// Before optimization
<path d="M123.456 789.012 L345.678 901.234" />

// After optimization
<path d="M123.5 789 L346 901.2" />
```

### 6.5 Prevent Hydration Mismatch Without Flickering

Use inline script for client-only data to prevent hydration mismatch.

```tsx
export default function RootLayout() {
  return (
    <html>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `
              window.theme = localStorage.getItem('theme') || 'light'
              document.documentElement.classList.add(window.theme)
            `
          }}
        />
      </head>
      <body>{children}</body>
    </html>
  )
}
```

### 6.6 Use Activity Component for Show/Hide

Use Activity component for show/hide to preserve state.

```tsx
import { Activity } from 'react-activity'

function TabPanel({ activeTab }: { activeTab: string }) {
  return (
    <Activity mode={activeTab}>
      <Activity.Mode mode="profile">
        <ProfileTab />
      </Activity.Mode>
      <Activity.Mode mode="settings">
        <SettingsTab />
      </Activity.Mode>
    </Activity>
  )
}
```

### 6.7 Use Explicit Conditional Rendering

Use ternary, not `&&` for conditionals to avoid rendering zeros.

```tsx
// Incorrect: renders 0 when count is 0
function ItemCount({ count }: { count: number }) {
  return <div>{count && <span>{count} items</span>}</div>
}

// Correct: uses ternary for explicit rendering
function ItemCount({ count }: { count: number }) {
  return <div>{count > 0 ? <span>{count} items</span> : null}</div>
}
```

---

## 7. JavaScript Performance (LOW-MEDIUM)

### 7.1 Batch DOM CSS Changes

Group CSS changes via classes or cssText.

```tsx
// Incorrect: multiple reflows
element.style.paddingTop = '10px'
element.style.paddingRight = '10px'
element.style.paddingBottom = '10px'
element.style.paddingLeft = '10px'

// Correct: single reflow
element.style.cssText = 'padding: 10px'
// or
element.className = 'padded'
```

### 7.2 Build Index Maps for Repeated Lookups

Build Map for repeated lookups.

```tsx
// Incorrect: O(n) lookup repeated
const users = [{ id: 1, name: 'Alice' }, { id: 2, name: 'Bob' }]
function getUserName(id: number) {
  return users.find(u => u.id === id)?.name
}

// Correct: O(1) lookup with Map
const userMap = new Map(users.map(u => [u.id, u]))
function getUserName(id: number) {
  return userMap.get(id)?.name
}
```

### 7.3 Cache Property Access in Loops

Cache object properties in loops.

```tsx
// Incorrect: property access on each iteration
for (let i = 0; i < items.length; i++) {
  processItem(items[i], options.enabled, options.verbose)
}

// Correct: cached property access
const { enabled, verbose } = options
for (let i = 0; i < items.length; i++) {
  processItem(items[i], enabled, verbose)
}
```

### 7.4 Cache Repeated Function Calls

Cache function results in module-level Map.

```tsx
const expensiveCache = new Map<string, Result>()

function expensiveComputation(input: string): Result {
  if (expensiveCache.has(input)) {
    return expensiveCache.get(input)!
  }

  const result = compute(input)
  expensiveCache.set(input, result)
  return result
}
```

### 7.5 Cache Storage API Calls

Cache localStorage/sessionStorage reads.

```tsx
// Incorrect: reads from storage on every call
function getTheme(): string {
  return localStorage.getItem('theme') || 'light'
}

// Correct: cached with invalidation
let cachedTheme: string | null = null

function getTheme(): string {
  if (cachedTheme === null) {
    cachedTheme = localStorage.getItem('theme') || 'light'
  }
  return cachedTheme
}

function setTheme(theme: string): void {
  localStorage.setItem('theme', theme)
  cachedTheme = theme
}
```

### 7.6 Combine Multiple Array Iterations

Combine multiple filter/map into one loop.

```tsx
// Incorrect: multiple iterations
const active = items.filter(item => item.active)
const sorted = active.sort((a, b) => a.name.localeCompare(b.name))
const transformed = sorted.map(item => ({ id: item.id, label: item.name }))

// Correct: single iteration
const result = items
  .filter(item => item.active)
  .sort((a, b) => a.name.localeCompare(b.name))
  .map(item => ({ id: item.id, label: item.name }))
```

### 7.7 Early Length Check for Array Comparisons

Check array length before expensive comparison.

```tsx
// Incorrect: compares even if lengths differ
function arraysEqual(a: unknown[], b: unknown[]): boolean {
  return a.every((item, i) => item === b[i])
}

// Correct: early exit on length mismatch
function arraysEqual(a: unknown[], b: unknown[]): boolean {
  if (a.length !== b.length) return false
  return a.every((item, i) => item === b[i])
}
```

### 7.8 Early Return from Functions

Return early from functions.

```tsx
// Incorrect: nested conditions
function process(input: string | null) {
  if (input !== null) {
    if (input.length > 0) {
      // process input
    } else {
      return { error: 'empty' }
    }
  } else {
    return { error: 'null' }
  }
}

// Correct: early returns
function process(input: string | null) {
  if (input === null) return { error: 'null' }
  if (input.length === 0) return { error: 'empty' }
  // process input
}
```

### 7.9 Hoist RegExp Creation

Hoist RegExp creation outside loops.

```tsx
// Incorrect: creates RegExp on each iteration
for (const str of strings) {
  const match = str.match(/\d{3}-\d{3}-\d{4}/)
  // ...
}

// Correct: RegExp created once
const phoneRegex = /\d{3}-\d{3}-\d{4}/
for (const str of strings) {
  const match = str.match(phoneRegex)
  // ...
}
```

### 7.10 Use Loop for Min/Max Instead of Sort

Use loop for min/max instead of sort.

```tsx
// Incorrect: O(n log n) for simple max
function max(arr: number[]): number {
  return arr.sort((a, b) => b - a)[0]
}

// Correct: O(n) for simple max
function max(arr: number[]): number {
  let max = arr[0]
  for (let i = 1; i < arr.length; i++) {
    if (arr[i] > max) max = arr[i]
  }
  return max
}
```

### 7.11 Use Set/Map for O(1) Lookups

Use Set/Map for O(1) lookups.

```tsx
// Incorrect: O(n) array.includes
const allowed = ['admin', 'user', 'guest']
function hasPermission(role: string): boolean {
  return allowed.includes(role)
}

// Correct: O(1) Set.has
const allowedSet = new Set(['admin', 'user', 'guest'])
function hasPermission(role: string): boolean {
  return allowedSet.has(role)
}
```

### 7.12 Use toSorted() Instead of sort() for Immutability

Use `toSorted()` for immutability.

```tsx
// Incorrect: mutates original array
const sorted = items.sort((a, b) => a.name.localeCompare(b.name))

// Correct: returns new sorted array
const sorted = items.toSorted((a, b) => a.name.localeCompare(b.name))
```

---

## 8. Advanced Patterns (LOW)

### 8.1 Store Event Handlers in Refs

Store event handlers in refs to avoid stale closures.

```tsx
function Component() {
  const [count, setCount] = useState(0)
  const handlerRef = useRef(() => {
    console.log(count)
  })

  // Update ref without triggering re-render
  handlerRef.current = () => {
    console.log(count)
  }

  useEffect(() => {
    const handler = () => handlerRef.current()
    window.addEventListener('click', handler)
    return () => window.removeEventListener('click', handler)
  }, [])

  return <button onClick={() => setCount(c => c + 1)}>Count: {count}</button>
}
```

### 8.2 useLatest for Stable Callback Refs

Use `useLatest` for stable callback refs.

```tsx
function useLatest<T>(value: T): { readonly current: T } {
  const ref = useRef(value)
  ref.current = value
  return ref
}

function Component() {
  const [count, setCount] = useState(0)
  const latestCount = useLatest(count)

  useEffect(() => {
    const interval = setInterval(() => {
      console.log(latestCount.current)
    }, 1000)
    return () => clearInterval(interval)
  }, [latestCount])

  return <button onClick={() => setCount(c => c + 1)}>Count: {count}</button>
}
```

---

## Reference

For the complete guide with all rules expanded, refer to the Vercel React Best Practices repository and documentation.

---

Version: 1.0.0
Last Updated: 2026-01-15
Source: Vercel Engineering
