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

## MUST

- Enable strict mode in tsconfig.json
- Use explicit return types for exported functions
- Use Zod for runtime validation
- Prefer const assertions for literals
- Use discriminated unions for state management
- Handle all Promise rejections

## MUST NOT

- Use any type (use unknown instead)
- Use @ts-ignore without explanation
- Export mutable variables
- Use non-null assertion (!) without validation
- Disable strict checks in tsconfig
- Mix async/await with .then() chains

## File Conventions

- *.test.ts or *.spec.ts for test files
- index.ts for barrel exports
- Use PascalCase for components (*.tsx)
- Use camelCase for utilities
- Use kebab-case for file names

## Testing

- Use Vitest for fast unit tests
- Use Testing Library for component tests
- Mock modules with vi.mock() or jest.mock()
- Use msw for API mocking

## React Specific (when applicable)

- Use React 19 Server Components by default
- Prefer function components over class components
- Use React.memo() for expensive renders
- Avoid prop drilling with Context or state libraries

## MoAI Integration

- Use Skill("moai-lang-typescript") for detailed patterns
- Follow TRUST 5 quality gates
