---
name: moai-lang-typescript
description: Enterprise TypeScript 5.9+ with strict typing, Next.js 16, React 19, tRPC, Zod for type-safe full-stack development and modern web applications
version: 1.0.0
modularized: false
tags:
  - programming-language
  - enterprise
  - typescript
  - development
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: react, lang, moai, typescript  


## Quick Reference

Enterprise TypeScript 5.9+ with strict type checking, Next.js 16 app router, React 19 server components, tRPC 11 for end-to-end type safety, and Zod 3.23 for runtime validation.

**Key Facts**:
- **Language**: TypeScript 5.9.3 (Aug 2025) with deferred module evaluation
- **Runtime**: Node.js 22 LTS (support until Apr 2027)
- **Frontend**: React 19 with server components, Next.js 16 with Turbopack
- **Type Safety**: tRPC 11 for type-safe APIs, Zod for runtime schema validation
- **Testing**: Vitest 2.x for fast, type-safe unit testing
- **When to Use**: Full-stack apps, type-safe APIs, component libraries, enterprise codebases

---

## What It Does

TypeScript provides compile-time type safety for JavaScript with powerful type inference. It excels in:

- **Type Safety**: Catch errors at compile time, not runtime, with strict type checking
- **Full-Stack Development**: Next.js 16 for server-rendered React apps with file-based routing
- **Type-Safe APIs**: tRPC for end-to-end type safety without code generation
- **Runtime Validation**: Zod schemas for runtime type checking and data validation
- **Component Libraries**: React 19 with TypeScript for type-safe UI components
- **Enterprise Patterns**: Advanced types (generics, conditional, mapped) for complex domains

TypeScript's static type system prevents common runtime errors, improves IDE support with autocomplete and refactoring, and enables confident refactoring in large codebases.

---

## When to Use

**Use TypeScript when**:
- Building large-scale applications requiring type safety and maintainability
- Creating full-stack apps with Next.js and React server components
- Implementing type-safe APIs with tRPC and runtime validation
- Developing component libraries and design systems
- Working in teams where code consistency and documentation are critical

**Avoid TypeScript when**:
- Rapid prototyping where type annotations slow initial development
- Small scripts where JavaScript is sufficient
- Build time overhead is unacceptable (though usually minimal with Turbopack)

---

## Key Features

1. **Static Type Checking**: Compile-time error detection with strict mode
2. **Type Inference**: Automatic type deduction reduces annotation burden
3. **Generics**: Reusable type-safe functions and classes with type parameters
4. **Union Types**: Model multiple possible types with `string | number`
5. **Intersection Types**: Combine types with `User & Admin`
6. **Conditional Types**: Advanced type transformations based on conditions
7. **Utility Types**: Built-in types like `Partial<T>`, `Pick<T, K>`, `Omit<T, K>`
8. **Type Guards**: Runtime type narrowing with `typeof`, `instanceof`, custom guards

---

## Works Well With

- `moai-lang-javascript` — JavaScript runtime and ecosystem compatibility
  - Best for: Migrating JavaScript codebases to TypeScript incrementally

- `moai-domain-frontend` — React, Next.js, Vue frontend frameworks
  - Best for: Type-safe component development and state management

- `moai-domain-backend` — Node.js backend with Express, Fastify, tRPC
  - Best for: Type-safe REST APIs and GraphQL servers

- `moai-domain-database` — TypeORM, Prisma for type-safe database access
  - Best for: Database operations with compile-time type checking

- `moai-essentials-review` — Code review and quality validation patterns
  - Best for: Enforcing type safety and best practices in PRs

---

## Core Concepts

### Type Inference vs Explicit Types
TypeScript infers types automatically in many cases (`const x = 5` → `number`). Explicit types are needed for function parameters, complex types, and when inference is ambiguous. Use inference when obvious, explicit types for clarity.

### Structural Typing (Duck Typing)
TypeScript uses structural typing: two types are compatible if their structures match, regardless of names. If `User` and `Person` have identical properties, they're interchangeable. This differs from nominal typing in Java/C#.

### Any vs Unknown vs Never
`any` disables type checking (avoid), `unknown` requires type narrowing before use (safe), `never` represents impossible values (function that never returns, exhaustive switch cases).

---

## Best Practices

### ✅ DO

1. **Enable Strict Mode**: Use `"strict": true` in tsconfig.json for maximum type safety
   - Reason: Catches more errors, enforces null checks, prevents implicit any

2. **Use Zod for Runtime Validation**: Combine TypeScript types with Zod schemas
   - Reason: Compile-time and runtime type safety together

3. **Prefer tRPC Over REST**: Use tRPC for type-safe APIs without code generation
   - Reason: Automatic type inference from server to client, fewer bugs

4. **Leverage Utility Types**: Use `Partial`, `Pick`, `Omit` to avoid duplication
   - Reason: DRY principle, maintainable type definitions

5. **Type Function Returns**: Always annotate function return types explicitly
   - Reason: Catches incorrect return values, better documentation

### ❌ DON'T

1. **Avoid `any` Type**: Never use `any` to bypass type checking
   - Reason: Defeats purpose of TypeScript, hides errors

2. **Don't Ignore Type Errors**: Fix type errors rather than using `@ts-ignore`
   - Reason: Type errors indicate real bugs or design issues

3. **Avoid Type Assertions**: Don't use `as` unless absolutely necessary
   - Reason: Overrides type system, can introduce runtime errors

4. **Don't Duplicate Types**: Avoid defining same type multiple times
   - Reason: Use `type` aliases or `interface` to share definitions

5. **Avoid Non-Null Assertions**: Don't use `!` operator without validation
   - Reason: Can cause runtime null pointer exceptions

---

## Implementation Guide

(See previous content for implementation details - TypeScript type system, React 19, Next.js 16, tRPC, Zod patterns)

---

## Advanced Patterns

(See previous content for advanced conditional types, mapped types, generic constraints, enterprise patterns)

---

## Context7 Integration

### Related Libraries & Tools
- [TypeScript](/microsoft/TypeScript): Typed superset of JavaScript
- [Next.js](/vercel/next.js): React framework with server-side rendering
- [React](/facebook/react): UI library for building user interfaces
- [tRPC](/trpc/trpc): End-to-end type-safe APIs
- [Zod](/colinhacks/zod): TypeScript-first schema validation

### Official Documentation
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Next.js Documentation](https://nextjs.org/docs)
- [React 19 Documentation](https://react.dev/)
- [tRPC Documentation](https://trpc.io/docs)
- [Zod Documentation](https://zod.dev/)

### Version-Specific Guides
Latest stable version: TypeScript 5.9.3, Next.js 16.x, React 19.x
- [TypeScript 5.9 Release Notes](https://devblogs.microsoft.com/typescript/announcing-typescript-5-9/)
- [Next.js 16 Upgrade Guide](https://nextjs.org/docs/app/building-your-application/upgrading)
- [React 19 Migration](https://react.dev/blog/2024/12/05/react-19)

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready  
**Version**: 4.0.0
