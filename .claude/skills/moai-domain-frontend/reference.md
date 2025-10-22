# Frontend Development Reference

> Official documentation for React, Angular, Vue, and modern frontend tooling

---

## Official Documentation Links

### Frameworks

| Framework | Version | Documentation | Status |
|-----------|---------|--------------|--------|
| **React** | 19.0 | https://react.dev/ | ✅ Current (2025) |
| **Next.js** | 15.1 | https://nextjs.org/docs | ✅ Current (2025) |
| **Angular** | 19.0 | https://angular.dev/ | ✅ Current (2025) |
| **Vue** | 3.5 | https://vuejs.org/ | ✅ Current (2025) |
| **Svelte** | 5.0 | https://svelte.dev/ | ✅ Current (2025) |

### Build Tools & Bundlers

| Tool | Version | Documentation | Status |
|------|---------|--------------|--------|
| **Vite** | 6.0 | https://vite.dev/ | ✅ Current (2025) |
| **Webpack** | 5.96 | https://webpack.js.org/ | ✅ Current (2025) |
| **Turbopack** | Latest | https://turbo.build/pack | ✅ Current (2025) |
| **esbuild** | 0.24 | https://esbuild.github.io/ | ✅ Current (2025) |

### State Management

| Library | Version | Documentation | Status |
|---------|---------|--------------|--------|
| **Zustand** | 5.0 | https://zustand-demo.pmnd.rs/ | ✅ Current (2025) |
| **Redux Toolkit** | 2.4 | https://redux-toolkit.js.org/ | ✅ Current (2025) |
| **Jotai** | 2.10 | https://jotai.org/ | ✅ Current (2025) |
| **Pinia** (Vue) | 2.3 | https://pinia.vuejs.org/ | ✅ Current (2025) |

---

## React Best Practices (2025)

### Modern Component Patterns

**Functional Components with Hooks**:
```tsx
import { useState, useEffect, useMemo } from 'react';

interface UserProps {
  userId: string;
}

export function UserProfile({ userId }: UserProps) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchUser(userId)
      .then(setUser)
      .finally(() => setLoading(false));
  }, [userId]);

  const displayName = useMemo(
    () => user ? `${user.firstName} ${user.lastName}` : '',
    [user]
  );

  if (loading) return <Skeleton />;
  if (!user) return <ErrorMessage />;

  return <div>{displayName}</div>;
}
```

### Server Components (React 19 + Next.js 15)

```tsx
// app/users/[id]/page.tsx
import { Suspense } from 'react';

// Server Component (async)
export default async function UserPage({ params }: { params: { id: string } }) {
  const user = await fetchUser(params.id);

  return (
    <div>
      <h1>{user.name}</h1>
      <Suspense fallback={<LoadingSkeleton />}>
        <UserPosts userId={user.id} />
      </Suspense>
    </div>
  );
}

// Client Component
'use client';
export function InteractiveButton() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(c => c + 1)}>{count}</button>;
}
```

### Performance Optimization

**Code Splitting**:
```tsx
import { lazy, Suspense } from 'react';

const HeavyComponent = lazy(() => import('./HeavyComponent'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <HeavyComponent />
    </Suspense>
  );
}
```

**Memoization**:
```tsx
import { memo } from 'react';

const ExpensiveComponent = memo(function ExpensiveComponent({ data }) {
  // Only re-renders if data changes
  return <div>{processData(data)}</div>;
});
```

---

## Vue 3 Best Practices

### Composition API

```vue
<script setup lang="ts">
import { ref, computed, watch } from 'vue';

interface User {
  id: number;
  name: string;
}

const user = ref<User | null>(null);
const loading = ref(true);

const displayName = computed(() => 
  user.value ? user.value.name.toUpperCase() : ''
);

watch(() => user.value?.id, (newId) => {
  if (newId) {
    fetchUserDetails(newId);
  }
});

async function loadUser() {
  loading.value = true;
  try {
    user.value = await fetchUser();
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div v-if="loading">Loading...</div>
  <div v-else-if="user">{{ displayName }}</div>
  <div v-else>No user found</div>
</template>
```

---

## Angular 19 Best Practices

### Standalone Components

```typescript
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div *ngIf="loading">Loading...</div>
    <div *ngIf="!loading && user">
      <h1>{{ user.name }}</h1>
      <p>{{ user.email }}</p>
    </div>
  `
})
export class UserProfileComponent {
  user: User | null = null;
  loading = true;

  ngOnInit() {
    this.userService.getUser().subscribe({
      next: (user) => {
        this.user = user;
        this.loading = false;
      },
      error: (err) => console.error(err)
    });
  }
}
```

### Signals (Angular 19)

```typescript
import { Component, signal, computed } from '@angular/core';

@Component({
  selector: 'app-counter',
  template: `
    <button (click)="increment()">Count: {{ doubleCount() }}</button>
  `
})
export class CounterComponent {
  count = signal(0);
  doubleCount = computed(() => this.count() * 2);

  increment() {
    this.count.update(v => v + 1);
  }
}
```

---

## Styling Solutions

### Tailwind CSS

```tsx
export function Button({ children, variant = 'primary' }) {
  const baseClasses = 'px-4 py-2 rounded font-semibold transition';
  const variantClasses = {
    primary: 'bg-blue-600 text-white hover:bg-blue-700',
    secondary: 'bg-gray-200 text-gray-800 hover:bg-gray-300'
  };

  return (
    <button className={`${baseClasses} ${variantClasses[variant]}`}>
      {children}
    </button>
  );
}
```

### CSS Modules

```tsx
import styles from './Button.module.css';

export function Button({ children }) {
  return <button className={styles.button}>{children}</button>;
}
```

---

## Testing Strategies

### Vitest + React Testing Library

```tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Counter } from './Counter';

describe('Counter', () => {
  it('increments count on button click', () => {
    render(<Counter />);
    const button = screen.getByRole('button');
    
    expect(button).toHaveTextContent('Count: 0');
    
    fireEvent.click(button);
    expect(button).toHaveTextContent('Count: 1');
  });
});
```

### Playwright E2E Testing

```typescript
import { test, expect } from '@playwright/test';

test('user can login', async ({ page }) => {
  await page.goto('http://localhost:3000');
  await page.fill('[name="email"]', 'user@example.com');
  await page.fill('[name="password"]', 'password123');
  await page.click('button[type="submit"]');
  
  await expect(page.locator('text=Welcome')).toBeVisible();
});
```

---

## Additional Resources

- **React Docs**: https://react.dev/learn
- **Vue Mastery**: https://www.vuemastery.com/
- **Angular University**: https://angular-university.io/
- **Web.dev**: https://web.dev/

---

## Advanced State Management

### Zustand (Recommended for React)

```typescript
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

interface CartStore {
  items: CartItem[];
  addItem: (item: CartItem) => void;
  removeItem: (id: string) => void;
  clearCart: () => void;
  total: number;
}

export const useCartStore = create<CartStore>()(
  devtools(
    persist(
      (set, get) => ({
        items: [],
        addItem: (item) => set((state) => ({
          items: [...state.items, item]
        })),
        removeItem: (id) => set((state) => ({
          items: state.items.filter((i) => i.id !== id)
        })),
        clearCart: () => set({ items: [] }),
        get total() {
          return get().items.reduce((sum, item) => sum + item.price, 0);
        }
      }),
      { name: 'cart-storage' }
    )
  )
);

// Usage
function CartSummary() {
  const { items, total, removeItem } = useCartStore();

  return (
    <div>
      <h2>Total: ${total}</h2>
      {items.map(item => (
        <div key={item.id}>
          {item.name} - ${item.price}
          <button onClick={() => removeItem(item.id)}>Remove</button>
        </div>
      ))}
    </div>
  );
}
```

### Redux Toolkit

```typescript
import { createSlice, configureStore, PayloadAction } from '@reduxjs/toolkit';

interface CounterState {
  value: number;
}

const counterSlice = createSlice({
  name: 'counter',
  initialState: { value: 0 } as CounterState,
  reducers: {
    increment: (state) => {
      state.value += 1;
    },
    decrement: (state) => {
      state.value -= 1;
    },
    incrementByAmount: (state, action: PayloadAction<number>) => {
      state.value += action.payload;
    }
  }
});

export const { increment, decrement, incrementByAmount } = counterSlice.actions;

export const store = configureStore({
  reducer: {
    counter: counterSlice.reducer
  }
});

// Usage with React
import { useSelector, useDispatch } from 'react-redux';

function Counter() {
  const count = useSelector((state: RootState) => state.counter.value);
  const dispatch = useDispatch();

  return (
    <div>
      <button onClick={() => dispatch(decrement())}>-</button>
      <span>{count}</span>
      <button onClick={() => dispatch(increment())}>+</button>
    </div>
  );
}
```

### Jotai (Atomic State Management)

```typescript
import { atom, useAtom } from 'jotai';

// Define atoms
const countAtom = atom(0);
const doubleCountAtom = atom((get) => get(countAtom) * 2);

// Usage
function Counter() {
  const [count, setCount] = useAtom(countAtom);
  const [doubled] = useAtom(doubleCountAtom);

  return (
    <div>
      <p>Count: {count}</p>
      <p>Doubled: {doubled}</p>
      <button onClick={() => setCount(c => c + 1)}>Increment</button>
    </div>
  );
}
```

---

## Form Management

### React Hook Form

```typescript
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const schema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  age: z.number().min(18, 'Must be 18 or older')
});

type FormData = z.infer<typeof schema>;

function RegistrationForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm<FormData>({
    resolver: zodResolver(schema)
  });

  const onSubmit = async (data: FormData) => {
    await fetch('/api/register', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('email')} type="email" />
      {errors.email && <span>{errors.email.message}</span>}

      <input {...register('password')} type="password" />
      {errors.password && <span>{errors.password.message}</span>}

      <input {...register('age', { valueAsNumber: true })} type="number" />
      {errors.age && <span>{errors.age.message}</span>}

      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Submitting...' : 'Register'}
      </button>
    </form>
  );
}
```

### VeeValidate (Vue)

```vue
<script setup lang="ts">
import { useForm } from 'vee-validate';
import * as yup from 'yup';

const schema = yup.object({
  email: yup.string().required().email(),
  password: yup.string().required().min(8),
});

const { handleSubmit, errors } = useForm({
  validationSchema: schema,
});

const onSubmit = handleSubmit(async (values) => {
  await fetch('/api/register', {
    method: 'POST',
    body: JSON.stringify(values),
  });
});
</script>

<template>
  <form @submit="onSubmit">
    <Field name="email" type="email" />
    <ErrorMessage name="email" />

    <Field name="password" type="password" />
    <ErrorMessage name="password" />

    <button type="submit">Register</button>
  </form>
</template>
```

---

## Data Fetching

### React Query (TanStack Query)

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

interface User {
  id: number;
  name: string;
  email: string;
}

// Fetch data
function UserProfile({ userId }: { userId: number }) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['user', userId],
    queryFn: async () => {
      const res = await fetch(`/api/users/${userId}`);
      if (!res.ok) throw new Error('Failed to fetch');
      return res.json() as Promise<User>;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return <div>{data.name}</div>;
}

// Mutate data
function UpdateUserForm({ userId }: { userId: number }) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (newName: string) => {
      const res = await fetch(`/api/users/${userId}`, {
        method: 'PATCH',
        body: JSON.stringify({ name: newName }),
      });
      return res.json();
    },
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ['user', userId] });
    },
  });

  return (
    <button onClick={() => mutation.mutate('New Name')}>
      Update Name
    </button>
  );
}
```

### SWR (Stale-While-Revalidate)

```typescript
import useSWR from 'swr';

const fetcher = (url: string) => fetch(url).then(res => res.json());

function Profile() {
  const { data, error, isLoading } = useSWR('/api/user', fetcher, {
    revalidateOnFocus: false,
    revalidateOnReconnect: true,
  });

  if (error) return <div>Failed to load</div>;
  if (isLoading) return <div>Loading...</div>;

  return <div>Hello {data.name}!</div>;
}
```

---

## Routing

### Next.js App Router (Recommended)

```typescript
// app/layout.tsx (Root layout)
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}

// app/page.tsx (Home page)
export default function Home() {
  return <h1>Welcome</h1>;
}

// app/users/[id]/page.tsx (Dynamic route)
export default async function UserPage({
  params
}: {
  params: { id: string }
}) {
  const user = await fetchUser(params.id);

  return <div>{user.name}</div>;
}

// app/api/users/route.ts (API route)
import { NextResponse } from 'next/server';

export async function GET() {
  const users = await fetchUsers();
  return NextResponse.json(users);
}

export async function POST(request: Request) {
  const body = await request.json();
  const user = await createUser(body);
  return NextResponse.json(user, { status: 201 });
}
```

### React Router v6

```typescript
import { createBrowserRouter, RouterProvider, Link } from 'react-router-dom';

const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Home />,
      },
      {
        path: 'users/:id',
        element: <UserProfile />,
        loader: async ({ params }) => {
          return fetch(`/api/users/${params.id}`);
        },
      },
    ],
  },
]);

function App() {
  return <RouterProvider router={router} />;
}

// Usage
function UserProfile() {
  const user = useLoaderData() as User;
  return <div>{user.name}</div>;
}
```

### Vue Router

```typescript
import { createRouter, createWebHistory } from 'vue-router';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: Home,
    },
    {
      path: '/users/:id',
      component: UserProfile,
      props: true,
    },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: 'dashboard',
          component: AdminDashboard,
        },
      ],
    },
  ],
});

// Navigation guard
router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !isAuthenticated()) {
    next('/login');
  } else {
    next();
  }
});
```

---

## Authentication

### JWT Authentication with Next.js

```typescript
// lib/auth.ts
import { SignJWT, jwtVerify } from 'jose';

const secret = new TextEncoder().encode(process.env.JWT_SECRET);

export async function createToken(userId: string) {
  return await new SignJWT({ userId })
    .setProtectedHeader({ alg: 'HS256' })
    .setExpirationTime('24h')
    .sign(secret);
}

export async function verifyToken(token: string) {
  try {
    const { payload } = await jwtVerify(token, secret);
    return payload;
  } catch {
    return null;
  }
}

// middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export async function middleware(request: NextRequest) {
  const token = request.cookies.get('token')?.value;

  if (!token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  const payload = await verifyToken(token);
  if (!payload) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: '/dashboard/:path*',
};
```

### Auth Context (React)

```typescript
import { createContext, useContext, useState, useEffect } from 'react';

interface User {
  id: string;
  email: string;
  name: string;
}

interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is logged in
    fetch('/api/auth/me')
      .then(res => res.json())
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setIsLoading(false));
  }, []);

  const login = async (email: string, password: string) => {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) throw new Error('Login failed');

    const user = await res.json();
    setUser(user);
  };

  const logout = () => {
    fetch('/api/auth/logout', { method: 'POST' });
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

// Usage
function ProtectedPage() {
  const { user, logout } = useAuth();

  if (!user) return <Navigate to="/login" />;

  return (
    <div>
      <p>Welcome, {user.name}</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

---

## Accessibility (a11y)

### ARIA Attributes

```tsx
function Modal({ isOpen, onClose, children }: ModalProps) {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      className="fixed inset-0 z-50 overflow-y-auto"
    >
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />
      <div className="relative bg-white rounded-lg p-6">
        <h2 id="modal-title">Modal Title</h2>
        {children}
        <button
          onClick={onClose}
          aria-label="Close modal"
          className="absolute top-2 right-2"
        >
          ×
        </button>
      </div>
    </div>
  );
}
```

### Keyboard Navigation

```tsx
function Dropdown({ items }: { items: string[] }) {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex(i => Math.min(i + 1, items.length - 1));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex(i => Math.max(i - 1, 0));
        break;
      case 'Enter':
        e.preventDefault();
        // Select item
        break;
      case 'Escape':
        setIsOpen(false);
        break;
    }
  };

  return (
    <div onKeyDown={handleKeyDown}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
      >
        Select Item
      </button>
      {isOpen && (
        <ul role="listbox" tabIndex={-1}>
          {items.map((item, index) => (
            <li
              key={item}
              role="option"
              aria-selected={index === selectedIndex}
              tabIndex={0}
            >
              {item}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

---

## Performance Optimization

### Code Splitting

```typescript
// React
import { lazy, Suspense } from 'react';

const HeavyComponent = lazy(() => import('./HeavyComponent'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <HeavyComponent />
    </Suspense>
  );
}

// Next.js dynamic imports
import dynamic from 'next/dynamic';

const DynamicChart = dynamic(() => import('../components/Chart'), {
  loading: () => <p>Loading chart...</p>,
  ssr: false, // Disable SSR for this component
});
```

### Image Optimization

```tsx
// Next.js Image component
import Image from 'next/image';

function Hero() {
  return (
    <Image
      src="/hero.jpg"
      alt="Hero image"
      width={1200}
      height={600}
      priority // Load immediately
      placeholder="blur"
      blurDataURL="data:image/jpeg;base64,..."
    />
  );
}

// Responsive images
<Image
  src="/photo.jpg"
  alt="Photo"
  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
  width={800}
  height={600}
/>
```

### Virtualization

```typescript
import { useVirtualizer } from '@tanstack/react-virtual';
import { useRef } from 'react';

function VirtualList({ items }: { items: string[] }) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 35,
    overscan: 5,
  });

  return (
    <div ref={parentRef} style={{ height: '400px', overflow: 'auto' }}>
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.index}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            {items[virtualItem.index]}
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## Internationalization (i18n)

### next-intl (Next.js)

```typescript
// messages/en.json
{
  "HomePage": {
    "title": "Welcome",
    "description": "This is the home page"
  }
}

// app/[locale]/page.tsx
import { useTranslations } from 'next-intl';

export default function HomePage() {
  const t = useTranslations('HomePage');

  return (
    <div>
      <h1>{t('title')}</h1>
      <p>{t('description')}</p>
    </div>
  );
}

// Middleware
import createMiddleware from 'next-intl/middleware';

export default createMiddleware({
  locales: ['en', 'de', 'ja'],
  defaultLocale: 'en',
});
```

### vue-i18n

```vue
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t, locale } = useI18n();

function changeLanguage(lang: string) {
  locale.value = lang;
}
</script>

<template>
  <div>
    <h1>{{ t('welcome') }}</h1>
    <p>{{ t('greeting', { name: 'Alice' }) }}</p>

    <button @click="changeLanguage('en')">English</button>
    <button @click="changeLanguage('ja')">日本語</button>
  </div>
</template>

<i18n>
{
  "en": {
    "welcome": "Welcome",
    "greeting": "Hello, {name}!"
  },
  "ja": {
    "welcome": "ようこそ",
    "greeting": "こんにちは、{name}さん!"
  }
}
</i18n>
```

---

## SEO Optimization

### Next.js Metadata API

```typescript
// app/layout.tsx
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: {
    default: 'My App',
    template: '%s | My App',
  },
  description: 'Welcome to my application',
  openGraph: {
    title: 'My App',
    description: 'Welcome to my application',
    images: ['/og-image.jpg'],
  },
};

// app/blog/[slug]/page.tsx
export async function generateMetadata({
  params
}: {
  params: { slug: string }
}): Promise<Metadata> {
  const post = await getPost(params.slug);

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: {
      title: post.title,
      description: post.excerpt,
      images: [post.coverImage],
    },
  };
}
```

### Structured Data (JSON-LD)

```tsx
function BlogPost({ post }: { post: Post }) {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'BlogPosting',
    headline: post.title,
    image: post.coverImage,
    datePublished: post.publishedAt,
    author: {
      '@type': 'Person',
      name: post.author.name,
    },
  };

  return (
    <article>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <h1>{post.title}</h1>
      {/* Rest of content */}
    </article>
  );
}
```

---

## Progressive Web App (PWA)

### Service Worker Registration

```typescript
// public/sw.js
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open('v1').then((cache) => {
      return cache.addAll([
        '/',
        '/styles.css',
        '/script.js',
      ]);
    })
  );
});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request);
    })
  );
});

// app/layout.tsx
useEffect(() => {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
  }
}, []);
```

### Web App Manifest

```json
{
  "name": "My PWA",
  "short_name": "MyPWA",
  "description": "A progressive web application",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#000000",
  "icons": [
    {
      "src": "/icon-192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/icon-512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

---

## WebSocket Integration

### Real-time Updates

```typescript
import { useEffect, useState } from 'react';

function useWebSocket(url: string) {
  const [messages, setMessages] = useState<string[]>([]);
  const [ws, setWs] = useState<WebSocket | null>(null);

  useEffect(() => {
    const socket = new WebSocket(url);

    socket.onmessage = (event) => {
      setMessages(prev => [...prev, event.data]);
    };

    socket.onclose = () => {
      console.log('WebSocket closed, reconnecting...');
      setTimeout(() => setWs(new WebSocket(url)), 3000);
    };

    setWs(socket);

    return () => socket.close();
  }, [url]);

  const sendMessage = (message: string) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(message);
    }
  };

  return { messages, sendMessage };
}

// Usage
function ChatRoom() {
  const { messages, sendMessage } = useWebSocket('ws://localhost:3000');
  const [input, setInput] = useState('');

  return (
    <div>
      <div>
        {messages.map((msg, i) => (
          <p key={i}>{msg}</p>
        ))}
      </div>
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            sendMessage(input);
            setInput('');
          }
        }}
      />
    </div>
  );
}
```

---

## Error Handling

### Error Boundaries (React)

```typescript
import { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
    // Log to error reporting service
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div>
            <h1>Something went wrong</h1>
            <details>
              <summary>Error details</summary>
              <pre>{this.state.error?.message}</pre>
            </details>
          </div>
        )
      );
    }

    return this.props.children;
  }
}

// Usage
<ErrorBoundary fallback={<ErrorPage />}>
  <App />
</ErrorBoundary>
```

### Next.js Error Handling

```typescript
// app/error.tsx
'use client';

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={() => reset()}>Try again</button>
    </div>
  );
}

// app/global-error.tsx
'use client';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html>
      <body>
        <h2>Global error occurred</h2>
        <button onClick={() => reset()}>Try again</button>
      </body>
    </html>
  );
}
```

---

**Last Updated**: 2025-10-22
**Framework Versions**: React 19.0, Vue 3.5, Angular 19.0, Next.js 15.1
