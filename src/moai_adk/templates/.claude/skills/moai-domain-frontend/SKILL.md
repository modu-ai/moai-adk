---
name: moai-domain-frontend
version: 4.0.0
status: stable
updated: 2025-11-20
description: Enterprise Frontend Development with modern architecture and scalable UI patterns
category: Domain
allowed-tools: [Read, Bash, Write, Edit, WebFetch]
---

# moai-domain-frontend: Enterprise Frontend Development

**AI-powered modern frontend architecture with React, Vue, Angular**

> Trust Score: 9.7/10 | Version: 4.0.0

---

## Overview

Enterprise Frontend Development expert with:

- **Modern Frameworks**: React 19, Vue 3.5, Angular 18, Svelte 5
- **State Management**: Zustand, TanStack Query, Redux Toolkit
- **Performance**: Code splitting, lazy loading, bundle optimization
- **Accessibility**: WCAG 2.1 AA compliance with ARIA support
- **TypeScript**: Full type safety with modern patterns

**Core Technologies**:

- React 19 with Server Components
- Next.js 16 with App Router and Turbopack
- Tailwind CSS and shadcn/ui
- Framer Motion for animations
- Zod for runtime validation

---

## React Component Architecture

### Modern Component with Hooks

```typescript
// components/UserList.tsx
import React, { useState, useMemo, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import { z } from "zod";

const UserSchema = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string().email(),
  role: z.enum(["admin", "user", "moderator"]),
});

type User = z.infer<typeof UserSchema>;

function useUsers(filters?: { role?: string; search?: string }) {
  return useQuery({
    queryKey: ["users", filters],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.role) params.append("role", filters.role);
      if (filters?.search) params.append("search", filters.search);

      const response = await fetch(`/api/users?${params}`);
      const data = await response.json();
      return z.array(UserSchema).parse(data);
    },
    staleTime: 5 * 60 * 1000,
  });
}

export const UserList: React.FC = React.memo(({ onUserSelect, filters }) => {
  const { data: users, isLoading, error } = useUsers(filters);

  const filteredUsers = useMemo(() => {
    if (!users) return [];
    return users.filter((user) => {
      if (filters?.role && user.role !== filters.role) return false;
      if (
        filters?.search &&
        !user.name.toLowerCase().includes(filters.search.toLowerCase())
      ) {
        return false;
      }
      return true;
    });
  }, [users, filters]);

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold">Users ({filteredUsers.length})</h2>
      {filteredUsers.map((user) => (
        <UserCard key={user.id} user={user} onClick={onUserSelect} />
      ))}
    </div>
  );
});
```

---

## State Management

### Zustand Store

```typescript
// stores/useAppStore.ts
import { create } from "zustand";
import { devtools } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";

interface AppState {
  currentUser: User | null;
  users: User[];
  theme: "light" | "dark";
  sidebarOpen: boolean;
  loading: { users: boolean; auth: boolean };
  errors: { users: string | null; auth: string | null };
}

interface AppActions {
  setCurrentUser: (user: User | null) => void;
  setTheme: (theme: "light" | "dark") => void;
  toggleSidebar: () => void;
  fetchUsers: () => Promise<void>;
}

export const useAppStore = create<AppState & AppActions>()(
  devtools(
    immer((set) => ({
      // Initial state
      currentUser: null,
      users: [],
      theme: "light",
      sidebarOpen: true,
      loading: { users: false, auth: false },
      errors: { users: null, auth: null },

      // Actions
      setCurrentUser: (user) => {
        set((state) => {
          state.currentUser = user;
        });
      },

      setTheme: (theme) => {
        set((state) => {
          state.theme = theme;
        });
      },

      toggleSidebar: () => {
        set((state) => {
          state.sidebarOpen = !state.sidebarOpen;
        });
      },

      fetchUsers: async () => {
        set((state) => {
          state.loading.users = true;
          state.errors.users = null;
        });

        try {
          const response = await fetch("/api/users");
          const users = await response.json();
          set((state) => {
            state.users = users;
            state.loading.users = false;
          });
        } catch (error) {
          set((state) => {
            state.errors.users = error.message;
            state.loading.users = false;
          });
        }
      },
    }))
  )
);
```

### TanStack Query

```typescript
// hooks/useUsers.ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

export function useUsers() {
  return useQuery({
    queryKey: ["users"],
    queryFn: async () => {
      const response = await fetch("/api/users");
      return response.json();
    },
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (userData: Omit<User, "id">) => {
      const response = await fetch("/api/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(userData),
      });
      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}
```

---

## Performance Optimization

### Code Splitting

```typescript
import dynamic from "next/dynamic";

// Lazy load heavy components
const HeavyChart = dynamic(() => import("./HeavyChart"), {
  loading: () => <div>Loading chart...</div>,
  ssr: false,
});

const AdminPanel = dynamic(() => import("./AdminPanel"));

// Route-based code splitting
const routes = [
  {
    path: "/dashboard",
    component: dynamic(() => import("./pages/Dashboard")),
  },
  {
    path: "/users",
    component: dynamic(() => import("./pages/Users")),
  },
];
```

### Intersection Observer

```typescript
function useIntersectionObserver(
  ref: React.RefObject<Element>,
  options: IntersectionObserverInit = {}
) {
  const [isIntersecting, setIsIntersecting] = React.useState(false);

  React.useEffect(() => {
    const observer = new IntersectionObserver(([entry]) => {
      setIsIntersecting(entry.isIntersecting);
    }, options);

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => {
      if (ref.current) {
        observer.unobserve(ref.current);
      }
    };
  }, [ref, options]);

  return isIntersecting;
}
```

---

## Accessibility

### ARIA Implementation

```typescript
interface AccessibleButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  ariaLabel?: string;
  isLoading?: boolean;
  loadingText?: string;
}

export const AccessibleButton = forwardRef<
  HTMLButtonElement,
  AccessibleButtonProps
>(
  (
    { children, ariaLabel, isLoading, loadingText, disabled, ...props },
    ref
  ) => {
    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        aria-label={ariaLabel}
        aria-busy={isLoading}
        {...props}
      >
        {isLoading && (
          <>
            <span className="sr-only">{loadingText || "Loading"}</span>
            <div
              className="animate-spin rounded-full h-4 w-4 border-2"
              aria-hidden="true"
            />
          </>
        )}
        {!isLoading && children}
      </button>
    );
  }
);
```

### Focus Management

```typescript
export function useFocusTrap(isActive: boolean) {
  const containerRef = useRef<HTMLElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!isActive) return;

    previousFocusRef.current = document.activeElement as HTMLElement;

    const focusableElements = containerRef.current?.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    ) as NodeListOf<HTMLElement>;

    if (focusableElements?.length > 0) {
      focusableElements[0].focus();
    }

    return () => {
      if (previousFocusRef.current) {
        previousFocusRef.current.focus();
      }
    };
  }, [isActive]);

  return containerRef;
}
```

---

## Vue 3 Composition API

```vue
<!-- UserList.vue -->
<template>
  <div class="space-y-4">
    <div class="flex justify-between items-center">
      <h2 class="text-2xl font-bold">Users ({{ filteredUsers.length }})</h2>
      <button
        @click="refreshUsers"
        class="px-4 py-2 bg-blue-500 text-white rounded"
      >
        Refresh
      </button>
    </div>

    <TransitionGroup name="fade" tag="div" class="space-y-4">
      <UserCard
        v-for="user in filteredUsers"
        :key="user.id"
        :user="user"
        @select="handleUserSelect"
      />
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useUsers } from "@/composables/useUsers";

const props = defineProps<{
  filters?: { role?: string; search?: string };
}>();

const emit = defineEmits<{
  "user-select": [user: User];
}>();

const { data: users, refresh } = useUsers(props.filters);

const filteredUsers = computed(() => {
  if (!users.value) return [];
  return users.value.filter((user) => {
    if (props.filters?.role && user.role !== props.filters.role) return false;
    if (
      props.filters?.search &&
      !user.name.toLowerCase().includes(props.filters.search.toLowerCase())
    ) {
      return false;
    }
    return true;
  });
});

const handleUserSelect = (user: User) => {
  emit("user-select", user);
};

const refreshUsers = () => {
  refresh.value();
};
</script>
```

---

## Best Practices

### Component Design

```typescript
// 1. Use TypeScript for type safety
interface Props {
  user: User;
  onSelect: (user: User) => void;
}

// 2. Memoize expensive computations
const sortedUsers = useMemo(() => {
  return users.sort((a, b) => a.name.localeCompare(b.name));
}, [users]);

// 3. Use useCallback for event handlers
const handleClick = useCallback(() => {
  onSelect(user);
}, [onSelect, user]);

// 4. Implement proper error boundaries
class ErrorBoundary extends React.Component {
  componentDidCatch(error, errorInfo) {
    console.error("Error:", error, errorInfo);
  }

  render() {
    return this.props.children;
  }
}
```

### Performance Tips

- Use React.memo for expensive components
- Implement virtual scrolling for long lists
- Lazy load images with loading="lazy"
- Use Web Workers for heavy computations
- Optimize bundle size with tree shaking

---

## Related Skills

- `moai-lib-shadcn-ui`: Component library
- `moai-baas-vercel-ext`: Deployment platform
- `moai-domain-figma`: Design integration

---

## Validation Checklist

**Architecture**:

- [ ] Component structure defined
- [ ] State management configured
- [ ] Routing implemented
- [ ] API integration complete

**Performance**:

- [ ] Code splitting enabled
- [ ] Lazy loading implemented
- [ ] Bundle size optimized
- [ ] Performance metrics tracked

**Accessibility**:

- [ ] ARIA labels added
- [ ] Keyboard navigation working
- [ ] Focus management implemented
- [ ] Screen reader tested

**Quality**:

- [ ] TypeScript configured
- [ ] Tests written
- [ ] Error boundaries added
- [ ] Loading states handled

---

**Last Updated**: 2025-11-20
