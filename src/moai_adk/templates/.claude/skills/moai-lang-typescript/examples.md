# TypeScript 5.9 Code Examples

Production-ready examples for modern TypeScript development with TypeScript 5.9.3, React 19, Next.js 16, tRPC 11, Zod 3.23, and Vitest 2.x.

---

## Example 1: TypeScript 5.9 Type System

### Basic Type Annotations
```typescript
// Primitive types
const name: string = "John";
const age: number = 30;
const active: boolean = true;

// Union types
type Status = "pending" | "approved" | "rejected";
const status: Status = "approved";

// Intersection types
type Admin = User & { role: "admin"; permissions: string[] };

// Generic types
interface Container<T> {
  value: T;
  getValue(): T;
  setValue(value: T): void;
}

const stringContainer: Container<string> = {
  value: "hello",
  getValue() { return this.value; },
  setValue(value) { this.value = value; }
};
```

### Advanced Type Features
```typescript
// Utility types
type Readonly<T> = { readonly [K in keyof T]: T[K] };
type Partial<T> = { [K in keyof T]?: T[K] };
type Record<K, T> = { [P in K]: T };

// Conditional types
type IsString<T> = T extends string ? true : false;
type Flatten<T> = T extends Array<infer U> ? U : T;

// Mapped types
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

interface Person {
  name: string;
  age: number;
}

type PersonGetters = Getters<Person>;
// Results in: { getName: () => string; getAge: () => number }
```

### Generic Constraints
```typescript
// Constraint T to objects with id property
function getById<T extends { id: string }>(
  items: T[],
  id: string
): T | undefined {
  return items.find(item => item.id === id);
}

// Constraint T to keyof another type
function pick<T, K extends keyof T>(
  obj: T,
  ...keys: K[]
): Pick<T, K> {
  const result = {} as Pick<T, K>;
  keys.forEach(key => {
    result[key] = obj[key];
  });
  return result;
}

const user = { id: '1', name: 'John', age: 30 };
const partial = pick(user, 'name', 'age');
// Type: { name: string; age: number }
```

---

## Example 2: React 19 Component Patterns

### Server Components
```typescript
// app/components/UserProfile.tsx
'use server'

import { getUserData } from '@/lib/db';

export async function UserProfile({ userId }: { userId: string }) {
  const userData = await getUserData(userId);

  return (
    <div className="user-profile">
      <h1>{userData.name}</h1>
      <p>{userData.bio}</p>
    </div>
  );
}
```

### Client Components with Transitions
```typescript
// app/components/Counter.tsx
'use client'

import { useState, useTransition } from 'react';

export function Counter() {
  const [count, setCount] = useState(0);
  const [isPending, startTransition] = useTransition();

  const handleIncrement = () => {
    startTransition(async () => {
      const newCount = await updateCountOnServer(count + 1);
      setCount(newCount);
    });
  };

  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={handleIncrement} disabled={isPending}>
        {isPending ? 'Updating...' : 'Increment'}
      </button>
    </div>
  );
}

async function updateCountOnServer(newCount: number): Promise<number> {
  'use server'
  // Server-side logic
  return newCount;
}
```

### Ref as Prop
```typescript
interface ComponentWithRefProps {
  ref?: React.Ref<HTMLInputElement>;
  value?: string;
}

// Now you can pass ref directly to component
const MyInput = React.forwardRef<
  HTMLInputElement,
  ComponentWithRefProps
>(({ value }, ref) => (
  <input ref={ref} defaultValue={value} />
));
```

---

## Example 3: Next.js 16 API Routes & Server Actions

### API Routes with Type Safety
```typescript
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { validateAuth } from '@/lib/auth';
import { z } from 'zod';

const UserSchema = z.object({
  name: z.string().min(1),
  email: z.string().email()
});

export async function GET(request: NextRequest) {
  const users = await fetchUsers();
  return NextResponse.json({ users });
}

export async function POST(request: NextRequest) {
  const data = await request.json();
  const validated = UserSchema.parse(data);
  const newUser = await createUser(validated);
  return NextResponse.json({ user: newUser }, { status: 201 });
}

// Dynamic route: app/api/users/[id]/route.ts
export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  const user = await getUserById(params.id);

  if (!user) {
    return NextResponse.json(
      { error: 'Not found' },
      { status: 404 }
    );
  }

  return NextResponse.json({ user });
}
```

### Server Actions
```typescript
// app/actions/users.ts
'use server'

import { revalidatePath } from 'next/cache';

export async function createUser(formData: FormData) {
  const name = formData.get('name');
  const email = formData.get('email');

  // Database operation
  const newUser = await db.users.create({
    name: String(name),
    email: String(email)
  });

  // Revalidate cache
  revalidatePath('/users');

  return newUser;
}

// Usage in component
// app/components/UserForm.tsx
'use client'

import { createUser } from '@/app/actions/users';

export function UserForm() {
  return (
    <form action={createUser}>
      <input name="name" required />
      <input name="email" type="email" required />
      <button type="submit">Create User</button>
    </form>
  );
}
```

---

## Example 4: tRPC Type-Safe APIs

### Router Definition
```typescript
// server/trpc.ts
import { initTRPC } from '@trpc/server';
import { z } from 'zod';

export const t = initTRPC.create();

export const router = t.router({
  user: t.router({
    list: t.procedure.query(async () => {
      return await db.user.findMany();
    }),

    byId: t.procedure
      .input(z.object({ id: z.string() }))
      .query(async ({ input }) => {
        return await db.user.findUnique({
          where: { id: input.id }
        });
      }),

    create: t.procedure
      .input(z.object({
        name: z.string(),
        email: z.string().email()
      }))
      .mutation(async ({ input }) => {
        return await db.user.create({
          data: input
        });
      })
  })
});

export type AppRouter = typeof router;
```

### Client Usage (Fully Typed)
```typescript
// client/trpc.ts
import { createTRPCReact } from '@trpc/react-query';
import type { AppRouter } from '@/server/trpc';

export const trpc = createTRPCReact<AppRouter>();

// Component usage - fully typed!
export function UserList() {
  const { data: users, isLoading } = trpc.user.list.useQuery();

  if (isLoading) return <div>Loading...</div>;

  return (
    <ul>
      {users?.map(user => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  );
}
```

---

## Example 5: Zod Validation

### Schema Definition
```typescript
import { z } from 'zod';

const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1).max(100),
  email: z.string().email(),
  age: z.number().int().min(0).max(150),
  roles: z.array(z.enum(['admin', 'user', 'guest'])).default(['user']),
  createdAt: z.date().default(() => new Date())
});

// Infer TypeScript type from schema
type User = z.infer<typeof UserSchema>;

// Validation
const userData = { /* ... */ };
const user = UserSchema.parse(userData); // Throws on error
const result = UserSchema.safeParse(userData); // Returns { success, data, error }
```

### Custom Validation
```typescript
const PasswordSchema = z.string()
  .min(8, 'Password must be at least 8 characters')
  .regex(/[A-Z]/, 'Must contain uppercase')
  .regex(/[0-9]/, 'Must contain numbers')
  .refine(
    (pwd) => !commonPasswords.includes(pwd),
    'Password is too common'
  );

const RegisterSchema = z.object({
  email: z.string().email(),
  password: PasswordSchema,
  confirmPassword: z.string()
}).refine(
  (data) => data.password === data.confirmPassword,
  { message: 'Passwords must match', path: ['confirmPassword'] }
);
```

---

## Example 6: Testing with Vitest

### Unit Tests
```typescript
import { describe, it, expect } from 'vitest';
import { add, multiply } from '@/lib/math';

describe('Math Utils', () => {
  it('should add numbers', () => {
    expect(add(2, 3)).toBe(5);
  });

  it('should multiply numbers', () => {
    expect(multiply(2, 3)).toBe(6);
  });
});
```

### Component Tests
```typescript
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';
import { Button } from '@/components/Button';

describe('Button Component', () => {
  it('renders with text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button')).toHaveTextContent('Click me');
  });

  it('calls onClick handler', async () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click</Button>);

    await userEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalled();
  });
});
```

---

## Example 7: Decorators (TypeScript 5.0+)

### Method Memoization
```typescript
// Enable experimentalDecorators in tsconfig.json

function Memoize(
  target: any,
  propertyKey: string,
  descriptor: PropertyDescriptor
) {
  const originalMethod = descriptor.value;
  const cache = new Map();

  descriptor.value = function(...args: any[]) {
    const key = JSON.stringify(args);
    if (!cache.has(key)) {
      cache.set(key, originalMethod.apply(this, args));
    }
    return cache.get(key);
  };

  return descriptor;
}

class MathUtils {
  @Memoize
  fibonacci(n: number): number {
    if (n <= 1) return n;
    return this.fibonacci(n - 1) + this.fibonacci(n - 2);
  }
}
```

---

## Example 8: Next.js 16 Configuration

### Full-Featured next.config.js
```typescript
import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // Use Turbopack for faster dev builds
  experimental: {
    turbopack: {
      resolveAlias: {
        '@/*': './*'
      }
    }
  },

  // Image optimization
  images: {
    formats: ['image/avif', 'image/webp'],
    remotePatterns: [
      { hostname: 'cdn.example.com' }
    ]
  },

  // Compression
  compress: true,

  // Security headers
  headers: async () => [
    {
      source: '/(.*)',
      headers: [
        { key: 'X-Content-Type-Options', value: 'nosniff' },
        { key: 'X-Frame-Options', value: 'DENY' },
        { key: 'X-XSS-Protection', value: '1; mode=block' }
      ]
    }
  ]
};

export default nextConfig;
```

---

**Learn More**: See `reference.md` for API details, tool versions, and troubleshooting guides.

