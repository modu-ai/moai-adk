---
name: moai-lang-typescript
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Enterprise TypeScript with strict typing and modern ecosystem TypeScript
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# Modern TypeScript Development — Enterprise  

## Quick Summary

**Primary Focus**: TypeScript 5.9 with strict typing, Next.js 16, React 19, and type-safe APIs
**Best For**: Full-stack development, API contracts, type safety, component libraries
**Key Libraries**: React 19, Next.js 16, tRPC 11, Zod 3.23, Vitest 2.x
**Auto-triggers**: TypeScript, Next.js, React, tRPC, strict types, type safety, API routes

| Version | Release | Support |
|---------|---------|---------|
| TypeScript 5.9.3 | Aug 2025 | Active |
| React 19.x | Dec 2024 | Active |
| Next.js 16.x | 2025 | Active |
| Node.js 22 LTS | Oct 2024 | Apr 2027 |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core TypeScript concepts and patterns:
- **Type System**: Primitives, unions, intersections, generics, utility types
- **React 19**: Server components, transitions, ref as prop, new hooks
- **Next.js 16**: File-based routing, API routes, server actions, Turbopack
- **tRPC Basics**: Type-safe API definition and client usage
- **Examples**: See `examples.md` for full code samples

### Level 2: Advanced Patterns (See reference.md)

Production-ready enterprise patterns:
- **Advanced Types**: Conditional types, mapped types, generic constraints
- **Zod Validation**: Runtime schema validation with type inference
- **Component Testing**: Vitest with React Testing Library
- **API Design**: tRPC routers, middleware, context
- **Pattern Reference**: See `reference.md` for API details and best practices

### Level 3: Production Deployment (Consult security/performance skills)

Enterprise deployment and optimization:
- **Build Optimization**: Turbopack configuration, code splitting
- **Docker Deployment**: Multi-stage builds, environment setup
- **Performance**: Web Vitals, caching strategies, bundle analysis
- **Monitoring**: Error tracking, metrics, observability
- **Details**: Skill("moai-essentials-perf"), Skill("moai-security-backend")

---

## Technology Stack (November 2025 Stable)

### Language & Runtime
- **TypeScript 5.9.3** (August 2025, deferred module evaluation)
- **Node.js 22.11.0 LTS** (April 2027 support)
- **JavaScript ES2020** target for compatibility

### Frontend Framework
- **React 19.x** (Server Components, Actions, new Hooks API)
- **Next.js 16.x** (Turbopack, App Router, Server Components)
- **Turbopack** (Rust-based bundler, 2x faster builds)

### Type-Safe Ecosystem
- **tRPC 11.x** (End-to-end type safety without code generation)
- **Zod 3.23.x** (Runtime validation with TypeScript inference)
- **Vitest 2.x** (Unit testing with Jest compatibility)

### Package Management
- **npm 11.x**, **pnpm 9.x**, **yarn 4.x** (all supported)
- **Node Modules**: ESM modules with bundler resolution

---

## TypeScript Type System

### Basic Types
```typescript
// Primitives
const name: string = "John";
const age: number = 30;
const active: boolean = true;

// Union types for flexible APIs
type Status = "pending" | "approved" | "rejected";
type Result = string | number;

// Intersection types for combining
type Admin = User & { role: "admin"; permissions: string[] };
```

### Generics & Constraints
```typescript
// Generic function
function getFirstElement<T>(arr: T[]): T {
  return arr[0];
}

// Generic with constraints
function merge<T extends object, U extends object>(obj1: T, obj2: U): T & U {
  return { ...obj1, ...obj2 };
}
```

### Utility Types
```typescript
// Common patterns
type Readonly<T> = { readonly [K in keyof T]: T[K] };
type Partial<T> = { [K in keyof T]?: T[K] };
type Pick<T, K> = { [P in K]: T[P] };
type Omit<T, K> = Pick<T, Exclude<keyof T, K>>;
type Record<K, T> = { [P in K]: T };
```

---

## React 19 Components

### Server & Client Components
```typescript
// Server Component
'use server'
export async function UserProfile({ userId }: { userId: string }) {
  const userData = await fetchUser(userId);
  return <div>{userData.name}</div>;
}

// Client Component with State
'use client'
export function Counter() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

### React Hooks
```typescript
// useState, useEffect, useContext
const [state, setState] = useState<number>(0);
useEffect(() => { /* ... */ }, []);

// useTransition for non-blocking updates
const [isPending, startTransition] = useTransition();

// useRef for DOM access
const inputRef = useRef<HTMLInputElement>(null);
```

---

## Next.js 16 Full-Stack Development

### File-Based Routing & Layouts
```
app/
├── page.tsx                  # Root page
├── layout.tsx                # Root layout
├── api/users/route.ts        # API route: /api/users
└── dashboard/
    ├── page.tsx              # /dashboard
    └── [id]/page.tsx         # /dashboard/[id]
```

### API Routes
```typescript
export async function GET(request: NextRequest) {
  return NextResponse.json({ users: [] });
}

export async function POST(request: NextRequest) {
  const data = await request.json();
  return NextResponse.json(data, { status: 201 });
}
```

### Server Actions
```typescript
'use server'

export async function createUser(formData: FormData) {
  const name = formData.get('name');
  await db.users.create({ data: { name } });
  revalidatePath('/users');
}
```

---

## Type-Safe APIs with tRPC

### Router Definition
```typescript
export const router = t.router({
  user: t.router({
    list: t.procedure.query(() => db.user.findMany()),

    getById: t.procedure
      .input(z.object({ id: z.string() }))
      .query(async ({ input }) =>
        db.user.findUnique({ where: { id: input.id } })
      )
  })
});

export type AppRouter = typeof router;
```

### Client Usage (Fully Typed)
```typescript
const { data: users } = trpc.user.list.useQuery();
// All types inferred from server router!
```

---

## Runtime Validation with Zod

### Schema Definition
```typescript
const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1).max(100),
  email: z.string().email()
});

type User = z.infer<typeof UserSchema>;

// Validation
const user = UserSchema.parse(data);  // Throws on error
const result = UserSchema.safeParse(data);  // Returns { success, data, error }
```

---

## Testing

### Unit Tests with Vitest
```typescript
import { describe, it, expect, vi } from 'vitest';

describe('Math', () => {
  it('adds numbers', () => {
    expect(add(2, 3)).toBe(5);
  });

  it('calls callback', () => {
    const fn = vi.fn();
    execute(fn);
    expect(fn).toHaveBeenCalled();
  });
});
```

### Component Tests
```typescript
import { render, screen } from '@testing-library/react';

it('renders button', () => {
  render(<Button>Click me</Button>);
  expect(screen.getByRole('button')).toBeInTheDocument();
});
```

---

## Production Best Practices

1. **Enable strict mode** in tsconfig.json for maximum type safety
2. **Use Zod** for runtime validation combined with TypeScript types
3. **Prefer tRPC** over REST for end-to-end type safety
4. **Use Server Components** by default in Next.js 16
5. **Implement proper error handling** with discriminated unions
6. **Test with Vitest** for fast, type-safe unit tests
7. **Leverage TypeScript utility types** for DRY code
8. **Use conditional types** for advanced type manipulations
9. **Monitor Web Vitals** in production environments
10. **Build with Turbopack** for faster development cycles

---

## Learn More

- **Examples**: See `examples.md` for React, Next.js, tRPC, Zod, and Vitest patterns
- **Reference**: See `reference.md` for API details, tool versions, and troubleshooting
- **TypeScript 5.9**: https://devblogs.microsoft.com/typescript/announcing-typescript-5-9/
- **React 19**: https://react.dev/
- **Next.js 16**: https://nextjs.org/docs
- **tRPC**: https://trpc.io/docs
- **Zod**: https://zod.dev/

---

**Skills**: Skill("moai-essentials-debug"), Skill("moai-essentials-perf"), Skill("moai-security-backend")
**Auto-loads**: TypeScript projects mentioning Next.js, React, tRPC, strict types, type safety

