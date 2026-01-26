---
paths:
  - "**/*.ts"
  - "**/*.tsx"
  - "**/tsconfig.json"
---

# TypeScript Rules

Version: TypeScript 5.9+

## Tooling

- Linting: ESLint 9 flat config or Biome
- Formatting: Prettier or Biome
- Testing: Vitest (recommended) or Jest
- Package management: pnpm (recommended) or npm

## Best Practices (2026)

- Enable strict mode: `"strict": true` in tsconfig.json
- Use Typed Routes for compile-time route safety (Next.js 15.5+)
- Avoid `any` type - use `unknown` for truly unknown types
- Use `satisfies` operator for type checking without widening
- Use Zod for runtime validation with `z.infer<>` for type inference

## React 19 Patterns

- Default to Server Components, use 'use client' only when needed
- Use `useActionState` for form actions
- Use `useOptimistic` for optimistic UI updates
- Use `use()` for reading promises and context
- Enable React Compiler for automatic memoization

## Next.js 15.5

- Use App Router with route groups for organization
- Fetch data in Server Components
- Use `React.cache()` for request memoization
- Implement streaming with Suspense boundaries

## MoAI Integration

- Use Skill("moai-lang-typescript") for detailed patterns
- Follow TRUST 5 quality gates
