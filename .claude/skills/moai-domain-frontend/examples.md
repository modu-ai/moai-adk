# Frontend Development - Practical Examples

**Last updated**: 2025-11-22

## Example 1: React 19 Server Components

```tsx
// app/products/page.tsx - Server component
import { getProducts } from '@/lib/api';

export default async function ProductsPage() {
    const products = await getProducts();

    return (
        <main>
            <h1>Products</h1>
            <ProductGrid products={products} />
        </main>
    );
}

// app/components/ProductGrid.tsx - Client component
'use client';

import { useState } from 'react';
import ProductCard from './ProductCard';

export default function ProductGrid({ products }) {
    const [filter, setFilter] = useState('');

    const filtered = products.filter(p =>
        p.name.toLowerCase().includes(filter.toLowerCase())
    );

    return (
        <>
            <input
                placeholder="Filter products..."
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
            />
            <div className="grid">
                {filtered.map(product => (
                    <ProductCard key={product.id} product={product} />
                ))}
            </div>
        </>
    );
}
```

## Example 2: State Management with Zustand

```typescript
// lib/store.ts
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface AuthStore {
    user: User | null;
    isLoading: boolean;
    login: (email: string, password: string) => Promise<void>;
    logout: () => void;
}

export const useAuth = create<AuthStore>()(
    devtools(
        (set) => ({
            user: null,
            isLoading: false,
            login: async (email, password) => {
                set({ isLoading: true });
                try {
                    const response = await fetch('/api/login', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ email, password })
                    });
                    const user = await response.json();
                    set({ user, isLoading: false });
                } catch (error) {
                    set({ isLoading: false });
                    throw error;
                }
            },
            logout: () => {
                set({ user: null });
            }
        }),
        { name: 'auth-store' }
    )
);

// Usage in component
export function LoginForm() {
    const { login, isLoading } = useAuth();

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const form = e.currentTarget;
        const email = new FormData(form).get('email');
        const password = new FormData(form).get('password');
        await login(email as string, password as string);
    };

    return (
        <form onSubmit={handleSubmit}>
            <input name="email" type="email" required />
            <input name="password" type="password" required />
            <button type="submit" disabled={isLoading}>
                {isLoading ? 'Logging in...' : 'Login'}
            </button>
        </form>
    );
}
```

## Example 3: Advanced Hooks Pattern

```typescript
// hooks/useAsync.ts
import { useCallback, useEffect, useState } from 'react';

interface UseAsyncState<T> {
    status: 'idle' | 'loading' | 'success' | 'error';
    data: T | null;
    error: Error | null;
}

export function useAsync<T>(
    asyncFunction: () => Promise<T>,
    immediate = true
): UseAsyncState<T> & { execute: () => Promise<void> } {
    const [state, setState] = useState<UseAsyncState<T>>({
        status: 'idle',
        data: null,
        error: null
    });

    const execute = useCallback(async () => {
        setState({ status: 'loading', data: null, error: null });
        try {
            const response = await asyncFunction();
            setState({ status: 'success', data: response, error: null });
        } catch (error) {
            setState({ status: 'error', data: null, error: error as Error });
        }
    }, [asyncFunction]);

    useEffect(() => {
        if (immediate) {
            execute();
        }
    }, [execute, immediate]);

    return { ...state, execute };
}

// Usage
export function UserProfile({ userId }: { userId: string }) {
    const { status, data: user, error, execute } = useAsync(
        () => fetch(`/api/users/${userId}`).then(r => r.json())
    );

    if (status === 'loading') return <div>Loading...</div>;
    if (status === 'error') return <div>Error: {error?.message}</div>;
    if (!user) return null;

    return (
        <div>
            <h1>{user.name}</h1>
            <button onClick={execute}>Refresh</button>
        </div>
    );
}
```

## Example 4: Next.js 15 App Router API Routes

```typescript
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
    const searchParams = request.nextUrl.searchParams;
    const page = searchParams.get('page') || '1';

    try {
        const users = await db.users.findMany({
            skip: (parseInt(page) - 1) * 10,
            take: 10
        });

        return NextResponse.json(users);
    } catch (error) {
        return NextResponse.json(
            { error: 'Failed to fetch users' },
            { status: 500 }
        );
    }
}

export async function POST(request: NextRequest) {
    const body = await request.json();

    try {
        const user = await db.users.create({ data: body });
        return NextResponse.json(user, { status: 201 });
    } catch (error) {
        return NextResponse.json(
            { error: 'Failed to create user' },
            { status: 400 }
        );
    }
}
```

## Example 5: Component Testing with Vitest

```typescript
// components/Button.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Button from './Button';

describe('Button Component', () => {
    it('renders with correct text', () => {
        render(<Button>Click me</Button>);
        expect(screen.getByText('Click me')).toBeInTheDocument();
    });

    it('calls onClick when clicked', async () => {
        const user = userEvent.setup();
        const handleClick = vi.fn();

        render(<Button onClick={handleClick}>Click</Button>);
        await user.click(screen.getByRole('button'));

        expect(handleClick).toHaveBeenCalledOnce();
    });

    it('disables button when disabled prop is true', () => {
        render(<Button disabled>Click me</Button>);
        expect(screen.getByRole('button')).toBeDisabled();
    });
});
```

## Example 6: Vue 3.5 Composition API

```typescript
// composables/useFetch.ts
import { ref, computed, onMounted } from 'vue';

export function useFetch(url: string) {
    const data = ref(null);
    const loading = ref(false);
    const error = ref(null);

    const fetch_data = async () => {
        loading.value = true;
        try {
            const response = await fetch(url);
            data.value = await response.json();
        } catch (err) {
            error.value = err;
        } finally {
            loading.value = false;
        }
    };

    onMounted(() => fetch_data());

    return {
        data: computed(() => data.value),
        loading: computed(() => loading.value),
        error: computed(() => error.value),
        refetch: fetch_data
    };
}

// components/UserList.vue
<template>
    <div>
        <div v-if="loading">Loading...</div>
        <div v-if="error">{{ error }}</div>
        <ul v-else>
            <li v-for="user in data" :key="user.id">
                {{ user.name }}
            </li>
        </ul>
        <button @click="refetch">Refresh</button>
    </div>
</template>

<script setup lang="ts">
import { useFetch } from '@/composables/useFetch';

const { data, loading, error, refetch } = useFetch('/api/users');
</script>
```

## Example 7: Tailwind CSS with Dynamic Classes

```typescript
// components/Card.tsx
import clsx from 'clsx';

interface CardProps {
    variant?: 'primary' | 'secondary';
    size?: 'sm' | 'md' | 'lg';
    children: React.ReactNode;
}

export function Card({ variant = 'primary', size = 'md', children }: CardProps) {
    return (
        <div
            className={clsx(
                'rounded-lg p-4 shadow-md',
                {
                    'bg-blue-50 border-blue-200': variant === 'primary',
                    'bg-gray-50 border-gray-200': variant === 'secondary'
                },
                {
                    'p-2': size === 'sm',
                    'p-4': size === 'md',
                    'p-6': size === 'lg'
                }
            )}
        >
            {children}
        </div>
    );
}
```

## Example 8: Form Handling with React Hook Form

```typescript
// components/UserForm.tsx
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const userSchema = z.object({
    email: z.string().email(),
    name: z.string().min(1),
    age: z.number().min(18)
});

type UserFormData = z.infer<typeof userSchema>;

export function UserForm() {
    const {
        control,
        handleSubmit,
        formState: { errors }
    } = useForm<UserFormData>({
        resolver: zodResolver(userSchema)
    });

    const onSubmit = async (data: UserFormData) => {
        const response = await fetch('/api/users', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        // Handle response
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Controller
                name="email"
                control={control}
                render={({ field }) => (
                    <input {...field} type="email" />
                )}
            />
            {errors.email && <span>{errors.email.message}</span>}

            <button type="submit">Submit</button>
        </form>
    );
}
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
