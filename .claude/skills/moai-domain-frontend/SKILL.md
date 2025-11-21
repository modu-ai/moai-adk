---
name: moai-domain-frontend
description: Enterprise Frontend Development with React 19, Next.js 15, Vue 3.5, modern bundlers, and 2025 best practices
allowed-tools: [Read, Bash, WebFetch]
---

## Quick Reference (30 seconds)

# Enterprise Frontend Development Expert

**Latest Frameworks (2025)**:
- **React 19** - Server Components, use() hook, Suspense improvements
- **Next.js 15** - Partial Prerendering (PPR), Turbopack stable
- **Vue 3.5** - Signals-based reactivity, enhanced Composition API
- **Angular 19** - Standalone components standard, signals reactivity
- **Vite 5.x** - Lightning-fast build tool with HMR

**Key Capabilities**:
- Server Components for zero-client JavaScript
- Advanced rendering strategies (SSR, SSG, ISR, PPR)
- Modern state management with signals
- Optimized bundling with Turbopack/Vite
- Full-stack type safety with TypeScript

**When to Use**:
- Frontend architecture and framework selection
- Component design and state management
- Performance optimization and bundle analysis
- SEO and accessibility implementation

---

## Implementation Guide

### React 19 - Server Components & use() Hook

**Server Component with Data Fetching**:
```typescript
// app/users/page.tsx - React Server Component
import { Suspense } from 'react';
import { UserList } from './UserList';
import { UserSkeleton } from './UserSkeleton';

export default async function UsersPage() {
  // Fetch data on server - no client bundle
  const users = await fetch('https://api.example.com/users').then(r => r.json());
  
  return (
    <div>
      <h1>Users</h1>
      <Suspense fallback={<UserSkeleton />}>
        <UserList usersPromise={users} />
      </Suspense>
    </div>
  );
}
```

**Client Component with use() Hook**:
```typescript
'use client';

import { use, useState } from 'react';

interface UserListProps {
  usersPromise: Promise<User[]>;
}

export function UserList({ usersPromise }: UserListProps) {
  // use() hook unwraps promises and context
  const users = use(usersPromise);
  const [filter, setFilter] = useState('');
  
  const filteredUsers = users.filter(u => 
    u.name.toLowerCase().includes(filter.toLowerCase())
  );
  
  return (
    <div>
      <input
        type="text"
        placeholder="Filter users..."
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
      />
      <ul>
        {filteredUsers.map(user => (
          <li key={user.id}>{user.name}</li>
        ))}
      </ul>
    </div>
  );
}
```

**Advanced Suspense with Error Boundaries**:
```typescript
import { Suspense } from 'react';
import { ErrorBoundary } from 'react-error-boundary';

function ErrorFallback({ error, resetErrorBoundary }) {
  return (
    <div role="alert">
      <p>Something went wrong:</p>
      <pre>{error.message}</pre>
      <button onClick={resetErrorBoundary}>Try again</button>
    </div>
  );
}

export default function Page() {
  return (
    <ErrorBoundary FallbackComponent={ErrorFallback}>
      <Suspense fallback={<Loading />}>
        <DataComponent />
      </Suspense>
    </ErrorBoundary>
  );
}
```

### Next.js 15 - Partial Prerendering (PPR)

**Enable Incremental PPR**:
```typescript
// next.config.ts
import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  experimental: {
    ppr: 'incremental', // Enable PPR incrementally
  },
};

export default nextConfig;
```

**PPR with Dynamic Components**:
```typescript
// app/dashboard/page.tsx
import { Suspense } from 'react';
import { UserProfile } from './UserProfile';
import { ActivityFeed } from './ActivityFeed';

// Enable PPR for this route
export const experimental_ppr = true;

export default function DashboardPage() {
  return (
    <div>
      {/* Static part - prerendered */}
      <h1>Dashboard</h1>
      <nav>Navigation links...</nav>
      
      {/* Dynamic part - streams */}
      <Suspense fallback={<ProfileSkeleton />}>
        <UserProfile />
      </Suspense>
      
      <Suspense fallback={<ActivitySkeleton />}>
        <ActivityFeed />
      </Suspense>
    </div>
  );
}
```

**Server Actions for Data Mutations**:
```typescript
// app/actions.ts
'use server';

import { revalidatePath } from 'next/cache';

export async function createPost(formData: FormData) {
  const title = formData.get('title') as string;
  const content = formData.get('content') as string;
  
  // Server-side data mutation
  await db.posts.create({
    title,
    content,
    createdAt: new Date(),
  });
  
  // Revalidate cache
  revalidatePath('/blog');
  
  return { success: true };
}

// app/blog/create/page.tsx
'use client';

import { createPost } from '@/app/actions';

export function CreatePostForm() {
  return (
    <form action={createPost}>
      <input name="title" required />
      <textarea name="content" required />
      <button type="submit">Create Post</button>
    </form>
  );
}
```

### Vue 3.5 - Signals & Reactivity

**Signals-Based Reactivity**:
```typescript
<script setup lang="ts">
import { ref, computed, watch } from 'vue';

// Reactive state with signals
const count = ref(0);
const doubled = computed(() => count.value * 2);

// Watch with immediate execution
watch(count, (newValue) => {
  console.log(`Count changed to: ${newValue}`);
}, { immediate: true });

function increment() {
  count.value++;
}
</script>

<template>
  <div>
    <p>Count: {{ count }}</p>
    <p>Doubled: {{ doubled }}</p>
    <button @click="increment">Increment</button>
  </div>
</template>
```

**Enhanced Composition API**:
```typescript
// composables/useUserData.ts
import { ref, computed } from 'vue';

export function useUserData(userId: string) {
  const user = ref(null);
  const loading = ref(false);
  const error = ref(null);
  
  const isAdmin = computed(() => user.value?.role === 'admin');
  
  async function fetchUser() {
    loading.value = true;
    try {
      const response = await fetch(`/api/users/${userId}`);
      user.value = await response.json();
    } catch (e) {
      error.value = e.message;
    } finally {
      loading.value = false;
    }
  }
  
  // Auto-fetch on mount
  fetchUser();
  
  return {
    user,
    loading,
    error,
    isAdmin,
    refetch: fetchUser
  };
}

// Usage in component
<script setup lang="ts">
import { useUserData } from '@/composables/useUserData';

const props = defineProps<{ userId: string }>();
const { user, loading, error, isAdmin, refetch } = useUserData(props.userId);
</script>
```

### Modern Bundlers

**Vite 5.x Configuration**:
```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    // Code splitting optimization
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom'],
          'ui-vendor': ['@mui/material', '@emotion/react'],
        },
      },
    },
    // Modern build target
    target: 'es2020',
    // Minification
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
      },
    },
  },
  // Development optimizations
  server: {
    hmr: {
      overlay: true,
    },
  },
});
```

**Turbopack (Next.js 15)**:
```typescript
// next.config.ts - Turbopack is now stable
const nextConfig = {
  // Turbopack enabled by default in Next.js 15
  experimental: {
    turbo: {
      // Custom module rules
      rules: {
        '*.svg': {
          loaders: ['@svgr/webpack'],
        },
      },
    },
  },
};
```

### CSS-in-JS & Styling

**Tailwind CSS 4.0 (JIT Standard)**:
```typescript
// tailwind.config.ts
import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./app/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      // Custom design tokens
      colors: {
        brand: {
          50: '#f0f9ff',
          500: '#0ea5e9',
          900: '#0c4a6e',
        },
      },
      // Custom animations
      animation: {
        'fade-in': 'fadeIn 0.3s ease-in-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
};

export default config;
```

**CSS Modules with TypeScript**:
```typescript
// Button.module.css
.button {
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  font-weight: 500;
}

.primary {
  background-color: var(--color-primary);
  color: white;
}

// Button.tsx
import styles from './Button.module.css';

interface ButtonProps {
  variant?: 'primary' | 'secondary';
  children: React.ReactNode;
}

export function Button({ variant = 'primary', children }: ButtonProps) {
  return (
    <button className={`${styles.button} ${styles[variant]}`}>
      {children}
    </button>
  );
}
```

---

## Advanced Patterns

### Performance Optimization

**React Performance Monitoring**:
```typescript
import { Profiler, ProfilerOnRenderCallback } from 'react';

const onRenderCallback: ProfilerOnRenderCallback = (
  id,
  phase,
  actualDuration,
  baseDuration,
  startTime,
  commitTime
) => {
  console.log(`Component ${id} (${phase}) rendered in ${actualDuration}ms`);
  
  // Send to analytics
  if (actualDuration > 50) {
    analytics.track('slow-render', {
      componentId: id,
      duration: actualDuration,
      phase,
    });
  }
};

export function App() {
  return (
    <Profiler id="App" onRender={onRenderCallback}>
      <YourApp />
    </Profiler>
  );
}
```

**Code Splitting with React.lazy**:
```typescript
import { lazy, Suspense } from 'react';

// Lazy load heavy components
const HeavyComponent = lazy(() => import('./HeavyComponent'));
const AdminPanel = lazy(() => import('./AdminPanel'));

export function App() {
  const { user } = useAuth();
  
  return (
    <Suspense fallback={<LoadingSpinner />}>
      {user?.isAdmin && <AdminPanel />}
      <Suspense fallback={<ComponentSkeleton />}>
        <HeavyComponent />
      </Suspense>
    </Suspense>
  );
}
```

### State Management (2025)

**Zustand with Persist**:
```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface UserStore {
  user: User | null;
  setUser: (user: User) => void;
  logout: () => void;
}

export const useUserStore = create<UserStore>()(
  persist(
    (set) => ({
      user: null,
      setUser: (user) => set({ user }),
      logout: () => set({ user: null }),
    }),
    {
      name: 'user-storage',
      partialize: (state) => ({ user: state.user }),
    }
  )
);
```

**TanStack Query (React Query v5)**:
```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const res = await fetch('/api/users');
      return res.json();
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useCreateUser() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (user: NewUser) => {
      const res = await fetch('/api/users', {
        method: 'POST',
        body: JSON.stringify(user),
      });
      return res.json();
    },
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });
}
```

---

## Reference & Resources

### Context7 Documentation Access

**Latest Framework Patterns**:
- React: `/facebook/react` - Server Components, hooks, patterns
- Next.js: `/vercel/next.js` - App Router, PPR, caching strategies
- Vue: `/vuejs/core` - Composition API, reactivity, performance

---

## Best Practices

### DO
- ✅ Use Server Components for zero-client JavaScript
- ✅ Implement code splitting for better performance
- ✅ Optimize images with next/image or similar
- ✅ Use TypeScript for type safety
- ✅ Implement proper error boundaries
- ✅ Monitor Web Vitals (LCP, FID, CLS)
- ✅ Use CSS-in-JS or utility-first CSS

### DON'T
- ❌ Load entire app as client-side bundle
- ❌ Skip accessibility (a11y) requirements
- ❌ Ignore bundle size optimization
- ❌ Use prop drilling instead of context
- ❌ Forget to memoize expensive computations
- ❌ Skip performance monitoring
- ❌ Use inline styles without optimization

---

**Last Updated**: 2025-11-22
**Version**: 5.0.0
**Status**: Production Ready (2025 Standards)
