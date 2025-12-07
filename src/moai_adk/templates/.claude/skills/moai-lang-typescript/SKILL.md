---
name: moai-lang-typescript
description: TypeScript 5.9+ development specialist covering React 19, Next.js 16 App Router, type-safe APIs with tRPC, Zod validation, and modern TypeScript patterns. Use when developing TypeScript applications, React components, Next.js pages, or type-safe APIs.
version: 1.0.0
category: language
tags:
  - typescript
  - react
  - nextjs
  - frontend
  - fullstack
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

TypeScript 5.9+ Development Specialist - Modern TypeScript with React 19, Next.js 16, and type-safe API patterns.

Auto-Triggers: `.ts`, `.tsx`, `.mts`, `.cts` files, TypeScript configurations, React/Next.js projects

Core Stack:
- TypeScript 5.9: Deferred module evaluation, decorators, satisfies operator
- React 19: Server Components, use() hook, Actions, concurrent features
- Next.js 16: App Router, Server Actions, middleware, ISR/SSG/SSR
- Type-Safe APIs: tRPC 11, Zod 3.23, tanstack-query
- Testing: Vitest, React Testing Library, Playwright

Quick Commands:
```bash
# Create Next.js 16 project
npx create-next-app@latest --typescript --tailwind --app

# Install type-safe API stack
npm install @trpc/server @trpc/client @trpc/react-query zod @tanstack/react-query

# Install testing stack
npm install -D vitest @testing-library/react @playwright/test
```

---

## Implementation Guide (5 minutes)

### TypeScript 5.9 Key Features

Satisfies Operator - Type checking without widening:
```typescript
type Colors = "red" | "green" | "blue";
const palette = {
  red: [255, 0, 0],
  green: "#00ff00",
  blue: [0, 0, 255],
} satisfies Record<Colors, string | number[]>;

palette.red.map((n) => n * 2); // Works - red is number[]
palette.green.toUpperCase();   // Works - green is string
```

Deferred Module Evaluation:
```typescript
import defer * as analytics from "./heavy-analytics";
function trackEvent(name: string) {
  analytics.track(name); // Loads module on first use
}
```

Modern Decorators (Stage 3):
```typescript
function logged<T extends (...args: any[]) => any>(
  target: T,
  context: ClassMethodDecoratorContext
) {
  return function (this: ThisParameterType<T>, ...args: Parameters<T>) {
    console.log(`Calling ${String(context.name)}`);
    return target.apply(this, args);
  };
}

class API {
  @logged
  async fetchUser(id: string) { return fetch(`/api/users/${id}`); }
}
```

### React 19 Patterns

Server Components (Default in App Router):
```typescript
// app/users/[id]/page.tsx - Server Component
interface PageProps { params: Promise<{ id: string }>; }

export default async function UserPage({ params }: PageProps) {
  const { id } = await params;
  const user = await db.user.findUnique({ where: { id } });
  if (!user) notFound();
  return <main><h1>{user.name}</h1></main>;
}
```

use() Hook - Unwrap Promises and Context:
```typescript
"use client";
import { use } from "react";

function UserProfile({ userPromise }: { userPromise: Promise<User> }) {
  const user = use(userPromise); // Suspends until resolved
  return <div>{user.name}</div>;
}
```

Actions - Form Handling with Server Functions:
```typescript
// app/actions.ts
"use server";
import { revalidatePath } from "next/cache";
import { z } from "zod";

const CreateUserSchema = z.object({
  name: z.string().min(2),
  email: z.string().email(),
});

export async function createUser(formData: FormData) {
  const validated = CreateUserSchema.parse({
    name: formData.get("name"),
    email: formData.get("email"),
  });
  await db.user.create({ data: validated });
  revalidatePath("/users");
}
```

useActionState for Form Status:
```typescript
"use client";
import { useActionState } from "react";

export function CreateUserForm() {
  const [state, action, isPending] = useActionState(createUser, null);
  return (
    <form action={action}>
      <input name="name" disabled={isPending} />
      <button disabled={isPending}>{isPending ? "Creating..." : "Create"}</button>
      {state?.error && <p className="error">{state.error}</p>}
    </form>
  );
}
```

### Next.js 16 App Router

Route Structure:
```
app/
  layout.tsx          # Root layout
  page.tsx            # Home page (/)
  loading.tsx         # Loading UI
  error.tsx           # Error boundary
  api/route.ts        # API route (/api)
  users/
    page.tsx          # /users
    [id]/page.tsx     # /users/:id
  (marketing)/about/page.tsx  # /about (route group)
```

Metadata API:
```typescript
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: { default: "My App", template: "%s | My App" },
  description: "Modern TypeScript application",
};

export async function generateMetadata({ params }): Promise<Metadata> {
  const { id } = await params;
  const user = await getUser(id);
  return { title: user.name };
}
```

Server Actions with Validation:
```typescript
"use server";
import { z } from "zod";
import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

const UpdateUserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(2).max(100),
  email: z.string().email(),
});

export async function updateUser(prevState: any, formData: FormData) {
  const result = UpdateUserSchema.safeParse(Object.fromEntries(formData));
  if (!result.success) {
    return { errors: result.error.flatten().fieldErrors };
  }
  await db.user.update({ where: { id: result.data.id }, data: result.data });
  revalidatePath(`/users/${result.data.id}`);
  redirect(`/users/${result.data.id}`);
}
```

### Type-Safe APIs with tRPC

Server Setup:
```typescript
// server/trpc.ts
import { initTRPC, TRPCError } from "@trpc/server";

const t = initTRPC.context<Context>().create();

export const router = t.router;
export const publicProcedure = t.procedure;
export const protectedProcedure = t.procedure.use(async ({ ctx, next }) => {
  if (!ctx.session?.user) throw new TRPCError({ code: "UNAUTHORIZED" });
  return next({ ctx: { ...ctx, user: ctx.session.user } });
});
```

Router Definition:
```typescript
import { z } from "zod";
import { router, publicProcedure, protectedProcedure } from "../trpc";

export const userRouter = router({
  getById: publicProcedure
    .input(z.object({ id: z.string().uuid() }))
    .query(({ input, ctx }) => ctx.db.user.findUnique({ where: { id: input.id } })),

  create: protectedProcedure
    .input(z.object({ name: z.string().min(2), email: z.string().email() }))
    .mutation(({ input, ctx }) => ctx.db.user.create({ data: input })),
});
```

Client Usage:
```typescript
"use client";
export function UserList() {
  const { data, isLoading } = trpc.user.list.useQuery({ page: 1 });
  const createUser = trpc.user.create.useMutation();
  if (isLoading) return <div>Loading...</div>;
  return <ul>{data?.map((u) => <li key={u.id}>{u.name}</li>)}</ul>;
}
```

### Zod Schema Patterns

Complex Validation:
```typescript
import { z } from "zod";

const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(2).max(100),
  email: z.string().email(),
  role: z.enum(["admin", "user", "guest"]),
  createdAt: z.coerce.date(),
}).strict();

type User = z.infer<typeof UserSchema>;

const CreateUserSchema = UserSchema.omit({ id: true, createdAt: true })
  .extend({ password: z.string().min(8), confirmPassword: z.string() })
  .refine((d) => d.password === d.confirmPassword, {
    message: "Passwords don't match", path: ["confirmPassword"],
  });
```

### State Management

Zustand for Client State:
```typescript
import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

interface AuthState {
  user: User | null;
  login: (user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(persist((set) => ({
    user: null,
    login: (user) => set({ user }),
    logout: () => set({ user: null }),
  }), { name: "auth-storage" }))
);
```

Jotai for Atomic State:
```typescript
import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

const countAtom = atom(0);
const doubleCountAtom = atom((get) => get(countAtom) * 2);
const themeAtom = atomWithStorage<"light" | "dark">("theme", "light");
```

---

## Advanced Patterns

For comprehensive documentation including advanced TypeScript patterns, performance optimization, testing strategies, and deployment configurations, see:

- [reference.md](reference.md) - Complete API reference, Context7 library mappings, advanced type patterns
- [examples.md](examples.md) - Production-ready code examples, full-stack patterns, testing templates

### Context7 Integration

```typescript
// TypeScript - mcp__context7__get_library_docs("/microsoft/TypeScript", "decorators satisfies", 1)
// React 19 - mcp__context7__get_library_docs("/facebook/react", "server-components use-hook", 1)
// Next.js 16 - mcp__context7__get_library_docs("/vercel/next.js", "app-router server-actions", 1)
// tRPC - mcp__context7__get_library_docs("/trpc/trpc", "procedures middleware", 1)
// Zod - mcp__context7__get_library_docs("/colinhacks/zod", "schema validation", 1)
```

---

## Works Well With

- `moai-domain-frontend` - UI components, styling patterns
- `moai-domain-backend` - API design, database integration
- `moai-library-shadcn` - Component library integration
- `moai-workflow-testing` - Testing strategies and patterns
- `moai-foundation-quality` - Code quality standards
- `moai-essentials-debug` - Debugging TypeScript applications

---

## Quick Troubleshooting

TypeScript Errors:
```bash
npx tsc --noEmit                    # Type check only
npx tsc --generateTrace ./trace     # Performance trace
```

React/Next.js Issues:
```bash
npm run build                       # Check for build errors
npx next lint                       # ESLint check
rm -rf .next && npm run dev         # Clear cache
```

Type Safety:
```typescript
// Exhaustive check
function assertNever(x: never): never { throw new Error(`Unexpected: ${x}`); }

// Type guard
function isUser(v: unknown): v is User {
  return typeof v === "object" && v !== null && "id" in v;
}
```

---

Version: 1.0.0
Last Updated: 2025-12-07
Status: Production Ready
