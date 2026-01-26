---
paths:
  - "**/*.ts"
  - "**/*.tsx"
  - "**/tsconfig.json"
---

# TypeScript Rules

Version: TypeScript 5.9+

## Tooling

- Linting: ESLint 9 or Biome
- Formatting: Prettier or Biome
- Testing: Vitest or Jest
- Package management: pnpm or npm

## Preferred Patterns

- Enable strict mode in tsconfig.json
- Avoid any type - use unknown instead
- Use Zod for runtime validation
- Use React 19 Server Components by default

## MoAI Integration

- Use Skill("moai-lang-typescript") for detailed patterns
- Follow TRUST 5 quality gates
