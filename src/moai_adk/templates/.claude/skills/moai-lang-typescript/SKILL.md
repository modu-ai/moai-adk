---
name: "moai-lang-typescript"
description: "Enterprise TypeScript with strict typing and modern ecosystem: TypeScript 5.9.3, Next.js 16, Turbopack, React 19, tRPC, Zod for type-safe schemas; activates for full-stack development, API contract definition, type safety enforcement, and framework-agnostic TypeScript patterns."
allowed-tools: 
version: "4.0.0"
status: stable
---

# Modern TypeScript Development — Enterprise v4.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Lines** | ~950 lines |
| **Size** | ~30KB |
| **Tier** | **3 (Professional)** |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | TypeScript strict types, Next.js patterns, full-stack development |
| **Trigger cues** | TypeScript, Next.js, React, tRPC, strict, types, API routes, SSR, full-stack |

## Technology Stack (November 2025 Stable)

### Core Language
- **TypeScript 5.9.3** (Released August 2025)
  - Deferred module evaluation (import defer)
  - Enhanced developer experience
  - Node.js 20 stable module resolution
  - Performance improvements (11% on large projects)

### Frontend Framework
- **React 19.x** (Latest stable)
  - Server Components
  - Actions (use server)
  - Ref as prop
  - New Hooks API

- **Next.js 16.x** (Full-stack framework)
  - Turbopack bundler (2x faster builds)
  - App Router (file-based routing)
  - Server Components by default
  - API Routes with middleware

### Build & Runtime
- **Node.js 22.11.0 LTS** (Long-term support until April 2027)
- **Turbopack** (Rust-based bundler, Webpack replacement)
- **Webpack 6.x** (Alternative bundler)
- **esbuild 0.23.x** (Alternative transpiler)

### Type-Safe APIs
- **tRPC 11.x** (End-to-end type-safe APIs)
  - Zero-cost abstractions
  - Automatic documentation
  - Client type inference

- **Zod 3.23.x** (Runtime schema validation)
  - Type inference from schemas
  - Custom error messages
  - Coercion support

- **OpenAPI 3.1.x** (API specification standard)

### Testing & Quality
- **Vitest 2.x** (Unit testing, Jest-compatible)
- **Playwright 1.48.x** (E2E testing)
- **TypeScript Compiler API** (For type checking)

### Package Management
- **npm 11.x** (Latest major version)
- **yarn 4.x** (Alternative package manager)
- **pnpm 9.x** (Fast, efficient package manager)

## Level 1: Fundamentals (High Freedom)

### 1. TypeScript 5.9 Type System

TypeScript 5.9 enhances type safety and developer experience:

**Basic Type Annotations**:
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

**Advanced Type Features**:
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

### 2. React 19 Component Patterns

React 19 introduces Server Components and new patterns:

**Server Components**:
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

**Client Components with Use**:
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

**Ref as Prop**:
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

### 3. Next.js 16 App Router

Next.js 16 with Turbopack provides modern full-stack development:

**Project Structure**:
```
app/
├── page.tsx                 // Home page
├── layout.tsx              // Root layout
├── api/
│   ├── users/
│   │   ├── route.ts        // GET, POST handlers
│   │   └── [id]/
│   │       └── route.ts    // Dynamic routes
│   └── health/
│       └── route.ts
├── dashboard/
│   ├── page.tsx            // /dashboard
│   ├── layout.tsx          // Dashboard layout
│   └── settings/
│       └── page.tsx        // /dashboard/settings
└── (auth)/
    ├── login/
    │   └── page.tsx
    └── register/
        └── page.tsx
```

**API Routes with Middleware**:
```typescript
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { validateAuth } from '@/lib/auth';

// Middleware for API route
export const middleware = validateAuth;

export async function GET(request: NextRequest) {
  const users = await fetchUsers();
  return NextResponse.json({ users });
}

export async function POST(request: NextRequest) {
  const data = await request.json();
  const newUser = await createUser(data);
  return NextResponse.json({ user: newUser }, { status: 201 });
}

// Dynamic route with parameters
// app/api/users/[id]/route.ts
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

**Server Actions**:
```typescript
// app/actions/users.ts
'use server'

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

### 4. tRPC for Type-Safe APIs

tRPC provides end-to-end type safety without code generation:

**Router Definition**:
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

**Client Usage (Type-Safe)**:
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

## Level 2: Advanced Patterns (Medium Freedom)

### 1. Type-Safe Data Validation with Zod

Zod provides runtime schema validation with TypeScript inference:

**Schema Definition**:
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

**Custom Validation**:
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

### 2. Advanced TypeScript Patterns

**Generic Constraints**:
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

**Decorator Pattern (TypeScript 5.0+)**:
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

### 3. Testing with Vitest

Vitest provides Jest-compatible testing for TypeScript:

**Unit Tests**:
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

**Component Tests**:
```typescript
import { render, screen } from '@testing-library/react';
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

## Level 3: Production Deployment (Low Freedom, Expert Only)

### 1. Build Optimization with Turbopack

Turbopack (Rust-based bundler) provides 2x faster builds:

**next.config.js Configuration**:
```typescript
import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // Use Turbopack for dev mode
  experimental: {
    turbopack: {
      // Turbopack-specific configuration
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
  
  // API routes compression
  compress: true,
  
  // Security headers
  headers: async () => [
    {
      source: '/(.*)',
      headers: [
        { key: 'X-Content-Type-Options', value: 'nosniff' },
        { key: 'X-Frame-Options', value: 'DENY' }
      ]
    }
  ]
};

export default nextConfig;
```

### 2. Deployment Strategies

**Docker Deployment**:
```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY package*.json ./

EXPOSE 3000
CMD ["npm", "start"]
```

**Environment Configuration**:
```typescript
// lib/config.ts
const config = {
  apiUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000',
  apiKey: process.env.API_KEY,
  environment: process.env.NODE_ENV || 'development',
  isDev: process.env.NODE_ENV === 'development'
};

export default config;
```

### 3. Performance Monitoring

**Web Vitals Tracking**:
```typescript
// app/layout.tsx
import { reportWebVitals } from 'next/web-vitals';

export function reportWebVitals(metric: any) {
  console.log(`${metric.name}: ${metric.value}ms`);
  
  // Send to analytics service
  fetch('/api/metrics', {
    method: 'POST',
    body: JSON.stringify(metric)
  });
}
```

## Auto-Load Triggers

This Skill automatically activates when you:
- Work with TypeScript projects and strict type checking
- Develop full-stack applications with Next.js
- Implement type-safe APIs with tRPC
- Need data validation with Zod
- Set up React 19 Server Components
- Debug TypeScript type errors
- Optimize builds with Turbopack

## Best Practices Summary

1. **Enable strict mode** in tsconfig.json for full type safety
2. **Use Zod for runtime validation** combined with TypeScript types
3. **Prefer tRPC over REST** for end-to-end type safety
4. **Use Server Components** by default in Next.js 16
5. **Implement proper error handling** with discriminated unions
6. **Test with Vitest** for fast, type-safe unit tests
7. **Leverage TypeScript utility types** for DRY code
8. **Use conditional types** for advanced type manipulations
9. **Monitor Web Vitals** in production environments
10. **Build with Turbopack** for faster development cycles

## See Also

- **TypeScript 5.9 Release**: https://devblogs.microsoft.com/typescript/announcing-typescript-5-9/
- **React 19 Documentation**: https://react.dev/
- **Next.js 16 Documentation**: https://nextjs.org/docs
- **tRPC Documentation**: https://trpc.io/docs
- **Zod Documentation**: https://zod.dev/
- **Turbopack Documentation**: https://turbo.build/pack

