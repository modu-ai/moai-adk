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

**Last Updated**: 2025-10-22
**Framework Versions**: React 19.0, Vue 3.5, Angular 19.0, Next.js 15.1
