---
paths:
  - "**/*.js"
  - "**/*.mjs"
  - "**/*.cjs"
  - "**/package.json"
---

# JavaScript Rules

Version: ES2024+, Node.js 24 LTS (recommended)

## Tooling

- Linting: ESLint 9 flat config or Biome
- Formatting: Prettier or Biome
- Testing: Vitest (recommended) or Jest
- Runtime: Node.js 24 LTS, Bun 1.x, or Deno 2.x

## Best Practices (2026)

- Use ESM modules over CommonJS
- Use async/await (required for modern Node.js)
- Use `toWellFormed()` for Unicode string handling (ES2024)
- Prefer const over let, avoid var
- Use optional chaining (?.) and nullish coalescing (??)

## Framework Selection

- Large-scale apps: NestJS (TypeScript, enforced architecture)
- Lightweight APIs: Express.js or Fastify
- Edge/serverless: Hono (multi-runtime support)

## Security (from Node.js best practices)

- Use linter security rules (eslint-plugin-security)
- Limit concurrent requests with rate limiting middleware
- Extract secrets from config files (use environment variables)
- Prevent query injection with ORM/ODM libraries

## ES2024 Features

- `Object.groupBy()` and `Map.groupBy()` for grouping
- `Promise.withResolvers()` for promise creation
- `Atomics.waitSync()` for thread synchronization

## MoAI Integration

- Use Skill("moai-lang-javascript") for detailed patterns
- Follow TRUST 5 quality gates
