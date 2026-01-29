---
name: moai-lang-typescript
description: >
  TypeScript 5.9+ development specialist covering React 19, Next.js 16 App Router, type-safe APIs with tRPC, Zod validation, and modern TypeScript patterns. Use when developing TypeScript applications, React components, Next.js pages, or type-safe APIs.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "typescript, react, nextjs, frontend, fullstack"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["TypeScript", "React", "Next.js", "tRPC", "Zod", ".ts", ".tsx", "tsconfig.json"]
  languages: ["typescript", "tsx"]
---

## Quick Reference (30 seconds)

TypeScript 5.9+ Development Specialist - Modern TypeScript with React 19, Next.js 16, and type-safe API patterns.

Auto-Triggers: Files with .ts, .tsx, .mts, or .cts extensions, TypeScript configurations, React or Next.js projects

Core Stack:

- TypeScript 5.9: Deferred module evaluation, decorators, satisfies operator
- React 19: Server Components, use hook, Actions, concurrent features
- Next.js 16: App Router, Server Actions, middleware, ISR/SSG/SSR
- Type-Safe APIs: tRPC 11, Zod 3.23, tanstack-query
- Testing: Vitest, React Testing Library, Playwright

Quick Commands:

Create Next.js 16 project using npx create-next-app with latest, typescript, tailwind, and app flags. Install type-safe API stack with npm install for trpc server, client, react-query, zod, and tanstack react-query. Install testing stack with npm install D flag for vitest, testing-library react, and playwright test.

---

## Implementation Guide (5 minutes)

### TypeScript 5.9 Key Features

Satisfies Operator for Type Checking Without Widening:

Define Colors type as union of red, green, and blue string literals. Create palette object with red as number array, green as hex string, and blue as number array. Apply satisfies operator with Record of Colors to string or number array. Now palette.red can use map method because it is inferred as number array, and palette.green can use toUpperCase because it is inferred as string.

Deferred Module Evaluation:

Use import defer with asterisk as namespace from heavy module path. In function that needs the module, access namespace property which loads module on first use.

Modern Decorators Stage 3:

Create logged function decorator that takes target function and ClassMethodDecoratorContext. Return function that logs method name then calls target with apply. Apply logged decorator to class method that fetches data.

### React 19 Patterns

Server Components Default in App Router:

For page component in app/users/[id]/page.tsx, define PageProps interface with params as Promise of object containing id string. Create async default function that awaits params, queries database for user, calls notFound if not found, and returns main element with user name.

use Hook for Unwrapping Promises and Context:

In client component marked with use client directive, import use from react. Create UserProfile component that takes userPromise prop as Promise of User type. Call use hook with the promise to suspend until resolved. Return div with user name.

Actions for Form Handling with Server Functions:

In server actions file marked with use server directive, import revalidatePath. Define CreateUserSchema with zod for name and email validation. Create async createUser function that takes FormData, parses with schema, creates user in database, and revalidates path.

useActionState for Form Status:

In client component, import useActionState. Create form component that destructures state, action, and isPending from useActionState called with createUser action. Return form with action prop, input disabled when pending, button with dynamic text, and error message from state.

### Next.js 16 App Router

Route Structure:

The app directory contains layout.tsx for root layout, page.tsx for home route, loading.tsx for loading UI, error.tsx for error boundary, and api/route.ts for API routes. Subdirectories like users contain page.tsx for list and [id]/page.tsx for dynamic routes. Route groups use parentheses like (marketing)/about/page.tsx.

Metadata API:

Import Metadata type. Export metadata object with title as object containing default and template, and description string. Export async generateMetadata function that takes params, awaits params, fetches user, and returns object with title set to user name.

Server Actions with Validation:

In server file, import zod, revalidatePath, and redirect. Define UpdateUserSchema with id, name, and email validation. Create async updateUser function taking prevState and formData. Parse with safeParse, return errors if failed, update database, revalidate path, and redirect.

### Type-Safe APIs with tRPC

Server Setup:

Import initTRPC and TRPCError from trpc server. Create t by calling initTRPC.context with Context type then create. Export router, publicProcedure, and protectedProcedure from t. The protectedProcedure uses middleware that checks session user and throws UNAUTHORIZED error if missing.

Router Definition:

Import zod. Create userRouter with router function. Define getById procedure using publicProcedure with input schema for id as uuid string, and query that finds user by id. Define create procedure using protectedProcedure with input schema for name and email, and mutation that creates user.

Client Usage:

In client component, create UserList function that calls trpc.user.list.useQuery with page parameter. Destructure data and isLoading. Create mutation with trpc.user.create.useMutation. Return loading state or list of users.

### Zod Schema Patterns

Complex Validation:

Create UserSchema with z.object containing id as uuid string, name with min and max length, email as email format, role as enum of admin, user, and guest, and createdAt with coerce.date. Apply strict method. Infer User type from schema. Create CreateUserSchema by omitting id and createdAt, extending with password and confirmPassword, and adding refine for password match validation with custom message and path.

### State Management

Zustand for Client State:

Import create from zustand and middleware. Define AuthState interface with user as User or null, login method, and logout method. Create useAuthStore with create function wrapped in devtools and persist middleware. Set initial user to null, login sets user, logout sets user to null. Persist uses auth-storage name.

Jotai for Atomic State:

Import atom from jotai and atomWithStorage from jotai/utils. Create countAtom with initial value 0. Create doubleCountAtom as derived atom that gets countAtom and multiplies by 2. Create themeAtom with atomWithStorage for light or dark theme persisted to storage.

---

## Advanced Patterns

For comprehensive documentation including advanced TypeScript patterns, performance optimization, testing strategies, and deployment configurations, see:

- reference.md for complete API reference, Context7 library mappings, and advanced type patterns
- examples.md for production-ready code examples, full-stack patterns, and testing templates

### Context7 Integration

For TypeScript documentation, use microsoft/TypeScript with decorators satisfies topics. For React 19, use facebook/react with server-components use-hook. For Next.js 16, use vercel/next.js with app-router server-actions. For tRPC, use trpc/trpc with procedures middleware. For Zod, use colinhacks/zod with schema validation.

---

## Works Well With

- moai-domain-frontend for UI components and styling patterns
- moai-domain-backend for API design and database integration
- moai-library-shadcn for component library integration
- moai-workflow-testing for testing strategies and patterns
- moai-foundation-quality for code quality standards
- moai-essentials-debug for debugging TypeScript applications

---

## Quick Troubleshooting

TypeScript Errors:

Run npx tsc with noEmit flag for type check only. Run npx tsc with generateTrace flag and output directory for performance trace.

React and Next.js Issues:

Run npm run build to check for build errors. Run npx next lint for ESLint check. Delete .next directory and run npm run dev to clear cache.

Type Safety Patterns:

Create assertNever function taking never type parameter that throws error for unexpected values, used in exhaustive switch statements. Create type guard function isUser that checks if value is object with id property and returns type predicate.

---

Last Updated: 2026-01-11
Status: Active (v1.1.0)
