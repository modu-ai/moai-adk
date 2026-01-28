---
paths:
  - "**/*.js"
  - "**/*.mjs"
  - "**/*.cjs"
  - "**/package.json"
---

# JavaScript Rules

Version: ES2024+, Node.js 22 LTS

## Tooling

- Linting: ESLint 9 or Biome
- Formatting: Prettier or Biome
- Testing: Vitest or Jest
- Runtime: Node.js, Bun 1.x, or Deno 2.x

## Preferred Patterns

- Use ESM modules over CommonJS
- Use async/await over callbacks
- Prefer const over let

## MoAI Integration

- Use Skill("moai-lang-javascript") for detailed patterns
- Follow TRUST 5 quality gates
