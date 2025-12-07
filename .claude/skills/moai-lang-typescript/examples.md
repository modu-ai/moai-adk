# TypeScript Production-Ready Examples

## Full-Stack Application Setup

### Next.js 16 + tRPC + Prisma

```
my-app/
├── src/
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   │   └── page.tsx
│   │   │   └── register/
│   │   │       └── page.tsx
│   │   ├── (dashboard)/
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   ├── api/
│   │   │   ├── auth/[...nextauth]/
│   │   │   │   └── route.ts
│   │   │   └── trpc/[trpc]/
│   │   │       └── route.ts
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── providers.tsx
│   ├── components/
│   │   ├── ui/
│   │   └── features/
│   ├── lib/
│   │   ├── db.ts
│   │   ├── auth.ts
│   │   └── trpc.ts
│   └── server/
│       ├── routers/
│       │   ├── user.ts
│       │   ├── post.ts
│       │   └── _app.ts
│       ├── context.ts
│       └── trpc.ts
├── prisma/
│   └── schema.prisma
├── package.json
└── tsconfig.json
```

### package.json

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "next dev --turbo",
    "build": "next build",
    "start": "next start",
    "lint": "next lint",
    "test": "vitest",
    "test:e2e": "playwright test",
    "db:push": "prisma db push",
    "db:studio": "prisma studio",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "next": "^16.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "@trpc/server": "^11.0.0",
    "@trpc/client": "^11.0.0",
    "@trpc/react-query": "^11.0.0",
    "@tanstack/react-query": "^5.59.0",
    "@prisma/client": "^6.0.0",
    "next-auth": "^5.0.0",
    "zod": "^3.23.8",
    "zustand": "^5.0.0",
    "superjson": "^2.2.1"
  },
  "devDependencies": {
    "typescript": "^5.9.0",
    "@types/react": "^19.0.0",
    "@types/node": "^22.0.0",
    "prisma": "^6.0.0",
    "vitest": "^2.1.0",
    "@testing-library/react": "^16.0.0",
    "@playwright/test": "^1.48.0",
    "eslint": "^9.0.0",
    "eslint-config-next": "^16.0.0"
  }
}
```

### tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "ES2022"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

---

## Database Layer (Prisma)

### prisma/schema.prisma

```prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id            String    @id @default(cuid())
  email         String    @unique
  name          String?
  emailVerified DateTime?
  image         String?
  password      String?
  role          Role      @default(USER)
  posts         Post[]
  accounts      Account[]
  sessions      Session[]
  createdAt     DateTime  @default(now())
  updatedAt     DateTime  @updatedAt

  @@index([email])
}

model Post {
  id          String   @id @default(cuid())
  title       String
  content     String?
  published   Boolean  @default(false)
  authorId    String
  author      User     @relation(fields: [authorId], references: [id], onDelete: Cascade)
  tags        Tag[]
  createdAt   DateTime @default(now())
  updatedAt   DateTime @updatedAt

  @@index([authorId])
  @@index([published])
}

model Tag {
  id    String @id @default(cuid())
  name  String @unique
  posts Post[]
}

enum Role {
  USER
  ADMIN
}

// NextAuth models
model Account {
  id                String  @id @default(cuid())
  userId            String
  type              String
  provider          String
  providerAccountId String
  refresh_token     String?
  access_token      String?
  expires_at        Int?
  token_type        String?
  scope             String?
  id_token          String?
  session_state     String?
  user              User    @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@unique([provider, providerAccountId])
}

model Session {
  id           String   @id @default(cuid())
  sessionToken String   @unique
  userId       String
  expires      DateTime
  user         User     @relation(fields: [userId], references: [id], onDelete: Cascade)
}
```

### src/lib/db.ts

```typescript
import { PrismaClient } from "@prisma/client";

const globalForPrisma = globalThis as unknown as {
  prisma: PrismaClient | undefined;
};

export const db =
  globalForPrisma.prisma ??
  new PrismaClient({
    log: process.env.NODE_ENV === "development" ? ["query", "error", "warn"] : ["error"],
  });

if (process.env.NODE_ENV !== "production") {
  globalForPrisma.prisma = db;
}
```

---

## tRPC Setup

### src/server/trpc.ts

```typescript
import { initTRPC, TRPCError } from "@trpc/server";
import superjson from "superjson";
import { ZodError } from "zod";
import type { Context } from "./context";

const t = initTRPC.context<Context>().create({
  transformer: superjson,
  errorFormatter({ shape, error }) {
    return {
      ...shape,
      data: {
        ...shape.data,
        zodError:
          error.cause instanceof ZodError ? error.cause.flatten() : null,
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
  return next({
    ctx: {
      ...ctx,
      user: ctx.session.user,
    },
  });
});

// Admin-only procedure
export const adminProcedure = protectedProcedure.use(async ({ ctx, next }) => {
  if (ctx.user.role !== "ADMIN") {
    throw new TRPCError({ code: "FORBIDDEN" });
  }
  return next({ ctx });
});
```

### src/server/context.ts

```typescript
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";
import { db } from "@/lib/db";
import type { inferAsyncReturnType } from "@trpc/server";
import type { FetchCreateContextFnOptions } from "@trpc/server/adapters/fetch";

export async function createContext(opts: FetchCreateContextFnOptions) {
  const session = await getServerSession(authOptions);

  return {
    db,
    session,
    headers: opts.req.headers,
  };
}

export type Context = inferAsyncReturnType<typeof createContext>;
```

### src/server/routers/user.ts

```typescript
import { z } from "zod";
import { router, publicProcedure, protectedProcedure, adminProcedure } from "../trpc";
import { TRPCError } from "@trpc/server";
import bcrypt from "bcryptjs";

const UserSchema = z.object({
  id: z.string(),
  email: z.string().email(),
  name: z.string().nullable(),
  role: z.enum(["USER", "ADMIN"]),
  createdAt: z.date(),
});

const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(2).max(100),
  password: z.string().min(8),
});

const UpdateUserSchema = z.object({
  name: z.string().min(2).max(100).optional(),
  email: z.string().email().optional(),
});

export const userRouter = router({
  // Public: Get user by ID
  getById: publicProcedure
    .input(z.object({ id: z.string() }))
    .query(async ({ input, ctx }) => {
      const user = await ctx.db.user.findUnique({
        where: { id: input.id },
        select: {
          id: true,
          email: true,
          name: true,
          role: true,
          createdAt: true,
        },
      });

      if (!user) {
        throw new TRPCError({
          code: "NOT_FOUND",
          message: "User not found",
        });
      }

      return user;
    }),

  // Protected: Get current user profile
  me: protectedProcedure.query(async ({ ctx }) => {
    return ctx.db.user.findUnique({
      where: { id: ctx.user.id },
      select: {
        id: true,
        email: true,
        name: true,
        role: true,
        createdAt: true,
        _count: { select: { posts: true } },
      },
    });
  }),

  // Protected: Update current user
  update: protectedProcedure
    .input(UpdateUserSchema)
    .mutation(async ({ input, ctx }) => {
      return ctx.db.user.update({
        where: { id: ctx.user.id },
        data: input,
      });
    }),

  // Admin: List all users with pagination
  list: adminProcedure
    .input(
      z.object({
        page: z.number().min(1).default(1),
        limit: z.number().min(1).max(100).default(10),
        search: z.string().optional(),
      })
    )
    .query(async ({ input, ctx }) => {
      const { page, limit, search } = input;
      const skip = (page - 1) * limit;

      const where = search
        ? {
            OR: [
              { name: { contains: search, mode: "insensitive" as const } },
              { email: { contains: search, mode: "insensitive" as const } },
            ],
          }
        : {};

      const [users, total] = await Promise.all([
        ctx.db.user.findMany({
          where,
          skip,
          take: limit,
          orderBy: { createdAt: "desc" },
          select: {
            id: true,
            email: true,
            name: true,
            role: true,
            createdAt: true,
          },
        }),
        ctx.db.user.count({ where }),
      ]);

      return {
        users,
        pagination: {
          page,
          limit,
          total,
          totalPages: Math.ceil(total / limit),
        },
      };
    }),

  // Admin: Delete user
  delete: adminProcedure
    .input(z.object({ id: z.string() }))
    .mutation(async ({ input, ctx }) => {
      if (input.id === ctx.user.id) {
        throw new TRPCError({
          code: "BAD_REQUEST",
          message: "Cannot delete your own account",
        });
      }

      return ctx.db.user.delete({ where: { id: input.id } });
    }),
});
```

### src/server/routers/post.ts

```typescript
import { z } from "zod";
import { router, publicProcedure, protectedProcedure } from "../trpc";
import { TRPCError } from "@trpc/server";

const CreatePostSchema = z.object({
  title: z.string().min(1).max(200),
  content: z.string().optional(),
  tags: z.array(z.string()).optional(),
});

const UpdatePostSchema = CreatePostSchema.partial().extend({
  id: z.string(),
  published: z.boolean().optional(),
});

export const postRouter = router({
  // Public: List published posts
  list: publicProcedure
    .input(
      z.object({
        page: z.number().min(1).default(1),
        limit: z.number().min(1).max(50).default(10),
        tag: z.string().optional(),
      })
    )
    .query(async ({ input, ctx }) => {
      const { page, limit, tag } = input;
      const skip = (page - 1) * limit;

      const where = {
        published: true,
        ...(tag && { tags: { some: { name: tag } } }),
      };

      const [posts, total] = await Promise.all([
        ctx.db.post.findMany({
          where,
          skip,
          take: limit,
          orderBy: { createdAt: "desc" },
          include: {
            author: { select: { id: true, name: true } },
            tags: { select: { name: true } },
          },
        }),
        ctx.db.post.count({ where }),
      ]);

      return { posts, total, page, totalPages: Math.ceil(total / limit) };
    }),

  // Public: Get single post
  getById: publicProcedure
    .input(z.object({ id: z.string() }))
    .query(async ({ input, ctx }) => {
      const post = await ctx.db.post.findUnique({
        where: { id: input.id },
        include: {
          author: { select: { id: true, name: true } },
          tags: { select: { name: true } },
        },
      });

      if (!post) {
        throw new TRPCError({ code: "NOT_FOUND", message: "Post not found" });
      }

      // Only show unpublished to author
      if (!post.published && post.authorId !== ctx.session?.user?.id) {
        throw new TRPCError({ code: "NOT_FOUND", message: "Post not found" });
      }

      return post;
    }),

  // Protected: Create post
  create: protectedProcedure
    .input(CreatePostSchema)
    .mutation(async ({ input, ctx }) => {
      const { tags, ...data } = input;

      return ctx.db.post.create({
        data: {
          ...data,
          authorId: ctx.user.id,
          tags: tags?.length
            ? {
                connectOrCreate: tags.map((name) => ({
                  where: { name },
                  create: { name },
                })),
              }
            : undefined,
        },
        include: { tags: true },
      });
    }),

  // Protected: Update own post
  update: protectedProcedure
    .input(UpdatePostSchema)
    .mutation(async ({ input, ctx }) => {
      const { id, tags, ...data } = input;

      const post = await ctx.db.post.findUnique({ where: { id } });

      if (!post) {
        throw new TRPCError({ code: "NOT_FOUND" });
      }

      if (post.authorId !== ctx.user.id) {
        throw new TRPCError({ code: "FORBIDDEN" });
      }

      return ctx.db.post.update({
        where: { id },
        data: {
          ...data,
          tags: tags
            ? {
                set: [],
                connectOrCreate: tags.map((name) => ({
                  where: { name },
                  create: { name },
                })),
              }
            : undefined,
        },
        include: { tags: true },
      });
    }),

  // Protected: Delete own post
  delete: protectedProcedure
    .input(z.object({ id: z.string() }))
    .mutation(async ({ input, ctx }) => {
      const post = await ctx.db.post.findUnique({ where: { id: input.id } });

      if (!post) {
        throw new TRPCError({ code: "NOT_FOUND" });
      }

      if (post.authorId !== ctx.user.id) {
        throw new TRPCError({ code: "FORBIDDEN" });
      }

      return ctx.db.post.delete({ where: { id: input.id } });
    }),

  // Protected: Get user's drafts
  myDrafts: protectedProcedure.query(async ({ ctx }) => {
    return ctx.db.post.findMany({
      where: { authorId: ctx.user.id, published: false },
      orderBy: { updatedAt: "desc" },
      include: { tags: true },
    });
  }),
});
```

### src/server/routers/_app.ts

```typescript
import { router } from "../trpc";
import { userRouter } from "./user";
import { postRouter } from "./post";

export const appRouter = router({
  user: userRouter,
  post: postRouter,
});

export type AppRouter = typeof appRouter;
```

---

## Client Setup

### src/lib/trpc.ts

```typescript
import { createTRPCReact } from "@trpc/react-query";
import { httpBatchLink, loggerLink } from "@trpc/client";
import superjson from "superjson";
import type { AppRouter } from "@/server/routers/_app";

export const trpc = createTRPCReact<AppRouter>();

function getBaseUrl() {
  if (typeof window !== "undefined") return "";
  if (process.env.VERCEL_URL) return `https://${process.env.VERCEL_URL}`;
  return `http://localhost:${process.env.PORT ?? 3000}`;
}

export const trpcClient = trpc.createClient({
  links: [
    loggerLink({
      enabled: (opts) =>
        process.env.NODE_ENV === "development" ||
        (opts.direction === "down" && opts.result instanceof Error),
    }),
    httpBatchLink({
      url: `${getBaseUrl()}/api/trpc`,
      transformer: superjson,
    }),
  ],
});
```

### src/app/providers.tsx

```typescript
"use client";

import { useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { trpc, trpcClient } from "@/lib/trpc";
import { SessionProvider } from "next-auth/react";

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            refetchOnWindowFocus: false,
          },
        },
      })
  );

  return (
    <SessionProvider>
      <trpc.Provider client={trpcClient} queryClient={queryClient}>
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      </trpc.Provider>
    </SessionProvider>
  );
}
```

---

## React Components

### src/components/features/PostList.tsx

```typescript
"use client";

import { trpc } from "@/lib/trpc";
import { useState } from "react";

interface PostListProps {
  initialTag?: string;
}

export function PostList({ initialTag }: PostListProps) {
  const [page, setPage] = useState(1);
  const [tag, setTag] = useState(initialTag);

  const { data, isLoading, error } = trpc.post.list.useQuery(
    { page, limit: 10, tag },
    { keepPreviousData: true }
  );

  if (isLoading) {
    return <PostListSkeleton />;
  }

  if (error) {
    return <div className="text-red-500">Error: {error.message}</div>;
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4">
        {data?.posts.map((post) => (
          <PostCard key={post.id} post={post} />
        ))}
      </div>

      {data && data.totalPages > 1 && (
        <Pagination
          currentPage={page}
          totalPages={data.totalPages}
          onPageChange={setPage}
        />
      )}
    </div>
  );
}

function PostCard({ post }: { post: Post }) {
  return (
    <article className="p-6 bg-white rounded-lg shadow">
      <h2 className="text-xl font-semibold">
        <a href={`/posts/${post.id}`} className="hover:text-blue-600">
          {post.title}
        </a>
      </h2>
      <p className="mt-2 text-gray-600 line-clamp-2">{post.content}</p>
      <div className="mt-4 flex items-center gap-4">
        <span className="text-sm text-gray-500">By {post.author.name}</span>
        <div className="flex gap-2">
          {post.tags.map((tag) => (
            <span
              key={tag.name}
              className="px-2 py-1 text-xs bg-gray-100 rounded"
            >
              {tag.name}
            </span>
          ))}
        </div>
      </div>
    </article>
  );
}
```

### src/components/features/CreatePostForm.tsx

```typescript
"use client";

import { useRouter } from "next/navigation";
import { trpc } from "@/lib/trpc";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

const formSchema = z.object({
  title: z.string().min(1, "Title is required").max(200),
  content: z.string().optional(),
  tags: z.string().optional(),
});

type FormData = z.infer<typeof formSchema>;

export function CreatePostForm() {
  const router = useRouter();
  const utils = trpc.useUtils();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
  });

  const createPost = trpc.post.create.useMutation({
    onSuccess: (post) => {
      utils.post.list.invalidate();
      router.push(`/posts/${post.id}`);
    },
  });

  const onSubmit = async (data: FormData) => {
    const tags = data.tags
      ?.split(",")
      .map((t) => t.trim())
      .filter(Boolean);

    await createPost.mutateAsync({
      title: data.title,
      content: data.content,
      tags,
    });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div>
        <label htmlFor="title" className="block text-sm font-medium">
          Title
        </label>
        <input
          id="title"
          {...register("title")}
          className="mt-1 block w-full rounded-md border p-2"
          disabled={isSubmitting}
        />
        {errors.title && (
          <p className="mt-1 text-sm text-red-500">{errors.title.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="content" className="block text-sm font-medium">
          Content
        </label>
        <textarea
          id="content"
          {...register("content")}
          rows={10}
          className="mt-1 block w-full rounded-md border p-2"
          disabled={isSubmitting}
        />
      </div>

      <div>
        <label htmlFor="tags" className="block text-sm font-medium">
          Tags (comma-separated)
        </label>
        <input
          id="tags"
          {...register("tags")}
          placeholder="react, typescript, nextjs"
          className="mt-1 block w-full rounded-md border p-2"
          disabled={isSubmitting}
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? "Creating..." : "Create Post"}
      </button>

      {createPost.error && (
        <p className="text-red-500">{createPost.error.message}</p>
      )}
    </form>
  );
}
```

---

## Testing Examples

### vitest.config.ts

```typescript
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: ["node_modules/", "src/test/"],
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
```

### src/test/setup.ts

```typescript
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach, vi } from "vitest";

afterEach(() => {
  cleanup();
});

// Mock next/navigation
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    back: vi.fn(),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
}));
```

### src/components/features/__tests__/PostCard.test.tsx

```typescript
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PostCard } from "../PostCard";

const mockPost = {
  id: "1",
  title: "Test Post",
  content: "This is test content",
  published: true,
  author: { id: "1", name: "John Doe" },
  tags: [{ name: "react" }, { name: "typescript" }],
  createdAt: new Date(),
};

describe("PostCard", () => {
  it("renders post title and content", () => {
    render(<PostCard post={mockPost} />);

    expect(screen.getByText("Test Post")).toBeInTheDocument();
    expect(screen.getByText("This is test content")).toBeInTheDocument();
  });

  it("displays author name", () => {
    render(<PostCard post={mockPost} />);

    expect(screen.getByText(/John Doe/)).toBeInTheDocument();
  });

  it("renders all tags", () => {
    render(<PostCard post={mockPost} />);

    expect(screen.getByText("react")).toBeInTheDocument();
    expect(screen.getByText("typescript")).toBeInTheDocument();
  });

  it("links to post detail page", () => {
    render(<PostCard post={mockPost} />);

    const link = screen.getByRole("link", { name: /Test Post/i });
    expect(link).toHaveAttribute("href", "/posts/1");
  });
});
```

### E2E Test: playwright/posts.spec.ts

```typescript
import { test, expect } from "@playwright/test";

test.describe("Posts", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
  });

  test("should display list of posts", async ({ page }) => {
    await expect(page.getByRole("article")).toHaveCount.above(0);
  });

  test("should navigate to post detail", async ({ page }) => {
    const firstPost = page.getByRole("article").first();
    const title = await firstPost.getByRole("heading").textContent();

    await firstPost.getByRole("link").click();

    await expect(page).toHaveURL(/\/posts\/.+/);
    await expect(page.getByRole("heading", { level: 1 })).toHaveText(title!);
  });

  test("should filter posts by tag", async ({ page }) => {
    await page.getByRole("button", { name: "react" }).click();

    const posts = page.getByRole("article");
    for (const post of await posts.all()) {
      await expect(post.getByText("react")).toBeVisible();
    }
  });
});

test.describe("Authenticated User", () => {
  test.use({ storageState: "playwright/.auth/user.json" });

  test("should create new post", async ({ page }) => {
    await page.goto("/posts/new");

    await page.getByLabel("Title").fill("My New Post");
    await page.getByLabel("Content").fill("This is my new post content.");
    await page.getByLabel("Tags").fill("test, e2e");

    await page.getByRole("button", { name: "Create Post" }).click();

    await expect(page).toHaveURL(/\/posts\/.+/);
    await expect(page.getByRole("heading", { level: 1 })).toHaveText(
      "My New Post"
    );
  });
});
```

---

## Environment Configuration

### .env.example

```env
# Database
DATABASE_URL="postgresql://user:password@localhost:5432/myapp?schema=public"

# NextAuth
NEXTAUTH_URL="http://localhost:3000"
NEXTAUTH_SECRET="your-secret-key-min-32-chars"

# OAuth Providers
GITHUB_ID=""
GITHUB_SECRET=""
GOOGLE_CLIENT_ID=""
GOOGLE_CLIENT_SECRET=""

# App
NEXT_PUBLIC_APP_URL="http://localhost:3000"
```

### src/lib/env.ts

```typescript
import { z } from "zod";

const envSchema = z.object({
  DATABASE_URL: z.string().url(),
  NEXTAUTH_URL: z.string().url(),
  NEXTAUTH_SECRET: z.string().min(32),
  NODE_ENV: z.enum(["development", "production", "test"]).default("development"),
  NEXT_PUBLIC_APP_URL: z.string().url().optional(),
});

const parsed = envSchema.safeParse(process.env);

if (!parsed.success) {
  console.error(
    "Invalid environment variables:",
    JSON.stringify(parsed.error.format(), null, 2)
  );
  throw new Error("Invalid environment variables");
}

export const env = parsed.data;
```

---

Version: 1.0.0
Last Updated: 2025-12-07
