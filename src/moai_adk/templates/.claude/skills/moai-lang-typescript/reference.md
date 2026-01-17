# TypeScript Development Reference

## TypeScript 5.9 Complete Reference

### New Features Overview

| Feature | Description | Use Case |
|---------|-------------|----------|
| Deferred Module Evaluation | Lazy-load modules on first access | Performance optimization |
| Decorators (Stage 3) | Native decorator support | Logging, validation, DI |
| Satisfies Operator | Type check without widening | Precise type inference |
| Const Type Parameters | Infer literal types in generics | Configuration objects |
| NoInfer Utility Type | Control inference in generic positions | API design |

### Advanced Type Patterns

#### Conditional Types

```typescript
// Extract return type from async function
type Awaited<T> = T extends Promise<infer U> ? U : T;

// Create type based on condition
type NonNullable<T> = T extends null | undefined ? never : T;

// Distributive conditional types
type ToArray<T> = T extends any ? T[] : never;
type Result = ToArray<string | number>; // string[] | number[]
```

#### Mapped Types

```typescript
// Make all properties optional
type Partial<T> = { [P in keyof T]?: T[P] };

// Make all properties readonly
type Readonly<T> = { readonly [P in keyof T]: T[P] };

// Pick specific properties
type Pick<T, K extends keyof T> = { [P in K]: T[P] };

// Custom mapped type with key transformation
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

interface User {
  name: string;
  age: number;
}

type UserGetters = Getters<User>;
// { getName: () => string; getAge: () => number; }
```

#### Template Literal Types

```typescript
// Event handler types
type EventName = "click" | "focus" | "blur";
type EventHandler = `on${Capitalize<EventName>}`;
// "onClick" | "onFocus" | "onBlur"

// API route types
type HTTPMethod = "GET" | "POST" | "PUT" | "DELETE";
type APIRoute<M extends HTTPMethod, P extends string> = `${M} ${P}`;
type UserRoutes = APIRoute<"GET" | "POST", "/users">;

// CSS utility types
type CSSProperty = "margin" | "padding";
type CSSDirection = "top" | "right" | "bottom" | "left";
type CSSUtility = `${CSSProperty}-${CSSDirection}`;
```

#### Variadic Tuple Types

```typescript
// Concat tuple types
type Concat<T extends unknown[], U extends unknown[]> = [...T, ...U];

// First and rest
type First<T extends unknown[]> = T extends [infer F, ...unknown[]] ? F : never;
type Rest<T extends unknown[]> = T extends [unknown, ...infer R] ? R : never;

// Typed pipe function
type PipeFunction<I, O> = (input: I) => O;

declare function pipe<A, B>(fn1: PipeFunction<A, B>): PipeFunction<A, B>;
declare function pipe<A, B, C>(
  fn1: PipeFunction<A, B>,
  fn2: PipeFunction<B, C>
): PipeFunction<A, C>;
declare function pipe<A, B, C, D>(
  fn1: PipeFunction<A, B>,
  fn2: PipeFunction<B, C>,
  fn3: PipeFunction<C, D>
): PipeFunction<A, D>;
```

### Utility Types Deep Dive

```typescript
// Record - Create object type with specific keys and values
type PageInfo = { title: string };
type PageRecord = Record<"home" | "about" | "contact", PageInfo>;

// Exclude/Extract - Filter union types
type T1 = Exclude<"a" | "b" | "c", "a">; // "b" | "c"
type T2 = Extract<"a" | "b" | "c", "a" | "f">; // "a"

// Parameters/ReturnType - Function type utilities
function greet(name: string, age: number): string {
  return `Hello ${name}, you are ${age}`;
}
type Params = Parameters<typeof greet>; // [string, number]
type Return = ReturnType<typeof greet>; // string

// Awaited - Unwrap Promise types
type A = Awaited<Promise<string>>; // string
type B = Awaited<Promise<Promise<number>>>; // number

// NoInfer - Prevent type inference
function createState<T>(initial: NoInfer<T>): [T, (value: T) => void] {
  // Implementation
}
```

---

## React 19 Complete Reference

### Server Components Architecture

```
                    ┌─────────────────────────────────────┐
                    │           Server                     │
                    │  ┌───────────────────────────────┐  │
                    │  │   Server Component Tree       │  │
                    │  │   - Data fetching             │  │
                    │  │   - Database access           │  │
                    │  │   - Server-only logic         │  │
                    │  └───────────────────────────────┘  │
                    │               │                      │
                    │               ▼ RSC Payload          │
                    └───────────────┼─────────────────────┘
                                    │
                    ┌───────────────┼─────────────────────┐
                    │           Client                     │
                    │               ▼                      │
                    │  ┌───────────────────────────────┐  │
                    │  │   Client Component Tree       │  │
                    │  │   - Interactivity             │  │
                    │  │   - State management          │  │
                    │  │   - Event handlers            │  │
                    │  └───────────────────────────────┘  │
                    └─────────────────────────────────────┘
```

### Component Patterns

#### Server Component with Streaming

```typescript
// app/dashboard/page.tsx
import { Suspense } from "react";
import { DashboardSkeleton, ChartSkeleton } from "./skeletons";

export default function DashboardPage() {
  return (
    <main>
      <h1>Dashboard</h1>
      <Suspense fallback={<DashboardSkeleton />}>
        <DashboardMetrics />
      </Suspense>
      <Suspense fallback={<ChartSkeleton />}>
        <AnalyticsChart />
      </Suspense>
    </main>
  );
}

async function DashboardMetrics() {
  const metrics = await fetchMetrics(); // Streamed independently
  return <MetricsGrid data={metrics} />;
}

async function AnalyticsChart() {
  const data = await fetchAnalytics(); // Streamed independently
  return <Chart data={data} />;
}
```

#### Client Component Patterns

```typescript
"use client";

import { useState, useTransition, useOptimistic } from "react";

interface Message {
  id: string;
  text: string;
  sending?: boolean;
}

export function MessageList({ messages }: { messages: Message[] }) {
  const [isPending, startTransition] = useTransition();
  const [optimisticMessages, addOptimistic] = useOptimistic(
    messages,
    (state, newMessage: Message) => [...state, { ...newMessage, sending: true }]
  );

  async function sendMessage(formData: FormData) {
    const text = formData.get("text") as string;
    const tempId = crypto.randomUUID();

    addOptimistic({ id: tempId, text });

    startTransition(async () => {
      await submitMessage(text);
    });
  }

  return (
    <div>
      <ul>
        {optimisticMessages.map((msg) => (
          <li key={msg.id} style={{ opacity: msg.sending ? 0.5 : 1 }}>
            {msg.text}
          </li>
        ))}
      </ul>
      <form action={sendMessage}>
        <input name="text" required />
        <button type="submit" disabled={isPending}>Send</button>
      </form>
    </div>
  );
}
```

### Hooks Reference

| Hook | Purpose | Server/Client |
|------|---------|---------------|
| use() | Unwrap Promise/Context | Both |
| useState | Component state | Client |
| useEffect | Side effects | Client |
| useContext | Access context | Client |
| useRef | Mutable reference | Client |
| useMemo | Memoize computation | Client |
| useCallback | Memoize callback | Client |
| useTransition | Non-blocking updates | Client |
| useOptimistic | Optimistic UI | Client |
| useActionState | Form action state | Client |
| useFormStatus | Form submission status | Client |
| useDeferredValue | Defer expensive updates | Client |
| useId | Generate unique IDs | Both |

---

## Next.js 16 Complete Reference

### Rendering Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| SSR | Server-Side Rendering | Dynamic, personalized content |
| SSG | Static Site Generation | Blog posts, documentation |
| ISR | Incremental Static Regeneration | E-commerce, frequently updated |
| CSR | Client-Side Rendering | Interactive dashboards |
| PPR | Partial Prerendering | Mixed static/dynamic |

### Route Configuration

```typescript
// Static generation
export const dynamic = "force-static";

// Dynamic on every request
export const dynamic = "force-dynamic";

// Revalidate every 60 seconds
export const revalidate = 60;

// Disable revalidation
export const revalidate = false;

// Runtime selection
export const runtime = "edge"; // or "nodejs"

// Preferred region for edge functions
export const preferredRegion = ["iad1", "sfo1"];

// Maximum duration for serverless functions
export const maxDuration = 30;
```

### Data Fetching Patterns

```typescript
// Parallel data fetching
async function Dashboard() {
  const [users, posts, comments] = await Promise.all([
    getUsers(),
    getPosts(),
    getComments(),
  ]);

  return <DashboardView users={users} posts={posts} comments={comments} />;
}

// Sequential data fetching (when dependent)
async function UserPosts({ userId }: { userId: string }) {
  const user = await getUser(userId);
  const posts = await getPosts(user.id); // Depends on user

  return <PostList posts={posts} />;
}

// With caching
import { unstable_cache } from "next/cache";

const getCachedUser = unstable_cache(
  async (id: string) => db.user.findUnique({ where: { id } }),
  ["user"],
  { revalidate: 3600, tags: ["user"] }
);

// Revalidation
import { revalidatePath, revalidateTag } from "next/cache";

export async function updateUser(id: string, data: UserData) {
  await db.user.update({ where: { id }, data });
  revalidatePath(`/users/${id}`);
  revalidateTag("user");
}
```

### Middleware Patterns

```typescript
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  // Authentication check
  const token = request.cookies.get("auth-token");
  if (!token && request.nextUrl.pathname.startsWith("/dashboard")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  // Rate limiting headers
  const response = NextResponse.next();
  response.headers.set("X-RateLimit-Limit", "100");
  response.headers.set("X-RateLimit-Remaining", "99");

  // Geolocation-based routing
  const country = request.geo?.country || "US";
  if (request.nextUrl.pathname === "/" && country === "DE") {
    return NextResponse.rewrite(new URL("/de", request.url));
  }

  return response;
}

export const config = {
  matcher: [
    "/((?!api|_next/static|_next/image|favicon.ico).*)",
  ],
};
```

---

## tRPC 11 Complete Reference

### Router Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    tRPC Architecture                     │
├─────────────────────────────────────────────────────────┤
│  Client                                                  │
│  ┌────────────────────────────────────────────────────┐ │
│  │  trpc.user.getById.useQuery({ id: "123" })         │ │
│  │                     │                               │ │
│  │                     ▼ Type-safe call                │ │
│  └────────────────────────────────────────────────────┘ │
│                        │                                 │
│                        │ HTTP/WebSocket                  │
│                        ▼                                 │
│  Server                                                  │
│  ┌────────────────────────────────────────────────────┐ │
│  │  userRouter.getById                                │ │
│  │    .input(z.object({ id: z.string() }))           │ │
│  │    .query(({ input }) => db.user.find(input.id))  │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Complete Setup

```typescript
// server/context.ts
import { getServerSession } from "next-auth";
import { db } from "@/lib/db";

export async function createContext({ req, res }: { req: Request; res: Response }) {
  const session = await getServerSession();
  return { db, session, req, res };
}

export type Context = Awaited<ReturnType<typeof createContext>>;

// server/trpc.ts
import { initTRPC, TRPCError } from "@trpc/server";
import { Context } from "./context";
import superjson from "superjson";

const t = initTRPC.context<Context>().create({
  transformer: superjson,
  errorFormatter({ shape, error }) {
    return {
      ...shape,
      data: {
        ...shape.data,
        zodError: error.cause instanceof ZodError ? error.cause.flatten() : null,
      },
    };
  },
});

export const router = t.router;
export const publicProcedure = t.procedure;
export const protectedProcedure = t.procedure.use(async ({ ctx, next }) => {
  if (!ctx.session?.user) {
    throw new TRPCError({ code: "UNAUTHORIZED" });
  }
  return next({ ctx: { ...ctx, user: ctx.session.user } });
});

// Middleware for logging
const loggerMiddleware = t.middleware(async ({ path, type, next }) => {
  const start = Date.now();
  const result = await next();
  console.log(`${type} ${path} - ${Date.now() - start}ms`);
  return result;
});

export const loggedProcedure = publicProcedure.use(loggerMiddleware);
```

### Subscriptions (WebSocket)

```typescript
// server/routers/notifications.ts
import { observable } from "@trpc/server/observable";
import { EventEmitter } from "events";

const ee = new EventEmitter();

export const notificationRouter = router({
  onNewMessage: protectedProcedure.subscription(({ ctx }) => {
    return observable<Message>((emit) => {
      const handler = (message: Message) => {
        if (message.userId === ctx.user.id) {
          emit.next(message);
        }
      };

      ee.on("message", handler);
      return () => ee.off("message", handler);
    });
  }),

  sendMessage: protectedProcedure
    .input(z.object({ text: z.string() }))
    .mutation(async ({ input, ctx }) => {
      const message = await db.message.create({
        data: { text: input.text, userId: ctx.user.id },
      });
      ee.emit("message", message);
      return message;
    }),
});
```

---

## Zod 3.23 Complete Reference

### Schema Types

| Type | Example | Description |
|------|---------|-------------|
| string | z.string() | String validation |
| number | z.number() | Number validation |
| boolean | z.boolean() | Boolean validation |
| date | z.date() | Date object validation |
| enum | z.enum(["a", "b"]) | Literal union |
| nativeEnum | z.nativeEnum(MyEnum) | TS enum validation |
| array | z.array(z.string()) | Array validation |
| object | z.object({...}) | Object validation |
| union | z.union([...]) | Type union |
| discriminatedUnion | z.discriminatedUnion(...) | Tagged union |
| tuple | z.tuple([...]) | Fixed-length array |
| record | z.record(z.string()) | Record type |
| map | z.map(z.string(), z.number()) | Map validation |
| set | z.set(z.string()) | Set validation |
| literal | z.literal("hello") | Exact value |
| null | z.null() | Null type |
| undefined | z.undefined() | Undefined type |
| any | z.any() | Any type |
| unknown | z.unknown() | Unknown type |
| never | z.never() | Never type |

### Advanced Patterns

```typescript
// Discriminated unions for type-safe variants
const EventSchema = z.discriminatedUnion("type", [
  z.object({ type: z.literal("click"), x: z.number(), y: z.number() }),
  z.object({ type: z.literal("keypress"), key: z.string() }),
  z.object({ type: z.literal("scroll"), delta: z.number() }),
]);

// Recursive types
interface Category {
  name: string;
  subcategories: Category[];
}

const CategorySchema: z.ZodType<Category> = z.lazy(() =>
  z.object({
    name: z.string(),
    subcategories: z.array(CategorySchema),
  })
);

// Branded types for type safety
const UserId = z.string().uuid().brand<"UserId">();
type UserId = z.infer<typeof UserId>;

// Error customization
const EmailSchema = z.string().email({
  message: "Please enter a valid email adddess",
}).refine((email) => !email.includes("+"), {
  message: "Email aliases are not allowed",
});

// Preprocessing
const DateSchema = z.preprocess(
  (val) => (typeof val === "string" ? new Date(val) : val),
  z.date()
);

// Coercion
const CoercedNumber = z.coerce.number(); // "42" -> 42
const CoercedDate = z.coerce.date(); // "2024-01-01" -> Date
```

---

## Context7 Library Mappings

### Primary Libraries

```
/microsoft/TypeScript       - TypeScript language and compiler
/facebook/react             - React library
/vercel/next.js             - Next.js framework
/trpc/trpc                  - tRPC type-safe APIs
/colinhacks/zod             - Zod schema validation
```

### State Management

```
/pmndrs/zustand             - Zustand state management
/pmndrs/jotai               - Jotai atomic state
/reduxjs/redux-toolkit      - Redux Toolkit
```

### UI Libraries

```
/shadcn-ui/ui               - shadcn/ui components
/tailwindlabs/tailwindcss   - Tailwind CSS
/radix-ui/primitives        - Radix UI primitives
/chakra-ui/chakra-ui        - Chakra UI
```

### Testing

```
/vitest-dev/vitest          - Vitest testing framework
/testing-library/react-testing-library - React Testing Library
/microsoft/playwright       - Playwright E2E testing
```

### Build Tools

```
/vercel/turbo               - Turborepo monorepo
/evanw/esbuild              - esbuild bundler
/privatenumber/tsup         - tsup bundler
/biomejs/biome              - Biome linter/formatter
```

---

## Performance Optimization

### Bundle Optimization

```typescript
// Dynamic imports for code splitting
const HeavyComponent = dynamic(() => import("./HeavyComponent"), {
  loading: () => <Skeleton />,
  ssr: false,
});

// Tree-shaking friendly exports
// utils/index.ts - BAD
export * from "./math";
export * from "./string";

// utils/index.ts - GOOD
export { add, subtract } from "./math";
export { capitalize } from "./string";
```

### React Optimization

```typescript
// Memo for expensive components
const ExpensiveList = memo(function ExpensiveList({ items }: Props) {
  return items.map((item) => <ExpensiveItem key={item.id} item={item} />);
}, (prev, next) => prev.items.length === next.items.length);

// useMemo for expensive calculations
const sortedItems = useMemo(
  () => items.sort((a, b) => a.name.localeCompare(b.name)),
  [items]
);

// useCallback for stable references
const handleClick = useCallback((id: string) => {
  setSelectedId(id);
}, []);
```

### TypeScript Compilation

```json
// tsconfig.json optimizations
{
  "compilerOptions": {
    "incremental": true,
    "tsBuildInfoFile": ".tsbuildinfo",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "isolatedModules": true
  }
}
```

---

## Security Best Practices

### Input Validation

```typescript
// Always validate on server
export async function createUser(formData: FormData) {
  const result = UserSchema.safeParse({
    name: formData.get("name"),
    email: formData.get("email"),
  });

  if (!result.success) {
    return { error: "Invalid input" }; // Don't expose details
  }

  // Proceed with validated data
}
```

### Environment Variables

```typescript
// env.ts
import { z } from "zod";

const envSchema = z.object({
  DATABASE_URL: z.string().url(),
  NEXTAUTH_SECRET: z.string().min(32),
  NEXTAUTH_URL: z.string().url(),
  NODE_ENV: z.enum(["development", "production", "test"]),
});

export const env = envSchema.parse(process.env);
```

### Authentication

```typescript
// Protect server actions
"use server";

import { getServerSession } from "next-auth";
import { redirect } from "next/navigation";

export async function protectedAction() {
  const session = await getServerSession();
  if (!session) {
    redirect("/login");
  }

  // Proceed with authenticated action
}
```

---

Version: 1.0.0
Last Updated: 2025-12-07
