# TypeScript CLI Reference

Quick reference for TypeScript 5.7, Vitest 2.1.8, Biome 1.9.4, npm/pnpm package management, and React 19.

---

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Purpose |
|------|---------|--------------|---------|
| **TypeScript** | 5.7.2 | 2025-01-15 | Language & Type System |
| **Vitest** | 2.1.8 | 2025-02-10 | Testing Framework |
| **Biome** | 1.9.4 | 2025-02-01 | Linter & Formatter |
| **Node.js** | 22.12.0 | 2025-02-05 | Runtime |
| **npm** | 10.9.2 | 2025-01-20 | Package Manager |
| **pnpm** | 9.15.1 | 2025-01-25 | Fast Package Manager |
| **React** | 19.0.0 | 2024-12-05 | UI Framework |
| **Vite** | 6.1.0 | 2025-01-30 | Build Tool |

---

## TypeScript 5.7

### Installation

```bash
# Install TypeScript globally
npm install -g typescript@5.7.2

# Install in project
npm install --save-dev typescript@5.7.2

# Verify installation
tsc --version
# Expected: Version 5.7.2
```

### Initialize Project

```bash
# Create tsconfig.json
tsc --init

# Or create custom tsconfig.json
npm init -y
npm install --save-dev typescript
npx tsc --init
```

### tsconfig.json (Recommended Settings)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "allowImportingTsExtensions": true,
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "jsx": "react-jsx",
    "types": ["vitest/globals"],
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

### Compiler Commands

```bash
# Compile all TypeScript files
tsc

# Compile specific file
tsc src/index.ts

# Watch mode (recompile on changes)
tsc --watch

# Type-check only (no output)
tsc --noEmit

# Show config
tsc --showConfig

# Project references
tsc --build
tsc -b --watch
```

### TypeScript 5.7 Features

**1. Inferred Type Predicates (5.5+)**

```typescript
function isString(value: unknown) {
  return typeof value === 'string'
  // Automatically inferred as: value is string
}
```

**2. Const Type Parameters (5.5+)**

```typescript
function processArray<const T extends readonly unknown[]>(arr: T) {
  return arr
}

const result = processArray([1, 2, 3] as const)
// type: readonly [1, 2, 3]
```

**3. Satisfies Operator**

```typescript
type Colors = 'red' | 'green' | 'blue'

const config = {
  background: 'red',
  foreground: 'blue'
} satisfies Record<string, Colors>

// Preserves literal types
const bg: 'red' = config.background // ✅
```

---

## Vitest 2.1.8

### Installation

```bash
# Install Vitest
npm install --save-dev vitest@2.1.8

# Install React Testing Library (for React projects)
npm install --save-dev @testing-library/react@16.1.0
npm install --save-dev @testing-library/user-event@14.5.2
npm install --save-dev @vitest/ui@2.1.8

# Install JSDOM (for browser environment)
npm install --save-dev jsdom@25.0.1
```

### Configuration (`vitest.config.ts`)

```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockData/',
        '**/types/'
      ],
      lines: 85,
      functions: 85,
      branches: 85,
      statements: 85
    }
  }
})
```

### Setup File (`src/test/setup.ts`)

```typescript
import { expect, afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'
import * as matchers from '@testing-library/jest-dom/matchers'

expect.extend(matchers)

afterEach(() => {
  cleanup()
})
```

### Common Commands

```bash
# Run tests (watch mode)
vitest

# Run tests once
vitest run

# Run specific test file
vitest run src/components/UserCard.test.tsx

# Run tests matching pattern
vitest run --testNamePattern="UserCard"

# Run with coverage
vitest run --coverage

# Open UI
vitest --ui

# Run in specific environment
vitest --environment jsdom

# Update snapshots
vitest run -u

# Show test output
vitest run --reporter=verbose
```

### Test API

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

describe('test suite', () => {
  beforeEach(() => {
    // Setup before each test
  })

  afterEach(() => {
    // Cleanup after each test
  })

  it('test name', () => {
    expect(1 + 1).toBe(2)
  })

  it.skip('skipped test', () => {
    // This test will be skipped
  })

  it.only('only this test runs', () => {
    // Only tests marked with .only will run
  })
})
```

### Assertions

```typescript
// Equality
expect(value).toBe(expected)
expect(value).toEqual(expected)
expect(value).toStrictEqual(expected)

// Truthiness
expect(value).toBeTruthy()
expect(value).toBeFalsy()
expect(value).toBeDefined()
expect(value).toBeUndefined()
expect(value).toBeNull()

// Numbers
expect(value).toBeGreaterThan(3)
expect(value).toBeGreaterThanOrEqual(3.5)
expect(value).toBeLessThan(5)
expect(value).toBeCloseTo(0.3, 5)

// Strings
expect(string).toMatch(/pattern/)
expect(string).toContain('substring')

// Arrays and iterables
expect(array).toContain(item)
expect(array).toHaveLength(3)

// Objects
expect(object).toHaveProperty('key')
expect(object).toMatchObject({ key: 'value' })

// Exceptions
expect(() => fn()).toThrow()
expect(() => fn()).toThrow('error message')
expect(() => fn()).toThrow(Error)

// Promises
await expect(promise).resolves.toBe(value)
await expect(promise).rejects.toThrow()

// Mocks
expect(mockFn).toHaveBeenCalled()
expect(mockFn).toHaveBeenCalledTimes(1)
expect(mockFn).toHaveBeenCalledWith(arg1, arg2)
```

### Mocking

```typescript
// Mock function
const mockFn = vi.fn()
mockFn.mockReturnValue(42)
mockFn.mockResolvedValue(Promise.resolve(42))
mockFn.mockRejectedValue(new Error('fail'))

// Mock implementation
mockFn.mockImplementation((x) => x + 1)

// Spy on object method
const spy = vi.spyOn(object, 'method')

// Mock module
vi.mock('./module', () => ({
  default: vi.fn(),
  namedExport: vi.fn()
}))

// Restore mocks
vi.restoreAllMocks()

// Clear mock calls
mockFn.mockClear()

// Reset mock implementation
mockFn.mockReset()
```

### Fake Timers

```typescript
// Use fake timers
vi.useFakeTimers()

// Advance time
vi.advanceTimersByTime(1000)

// Run all pending timers
vi.runAllTimers()

// Restore real timers
vi.restoreAllMocks()
```

---

## Biome 1.9.4

### Installation

```bash
# Install Biome
npm install --save-dev @biomejs/biome@1.9.4

# Initialize config
npx biome init
```

### Configuration (`biome.json`)

```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.4/schema.json",
  "vcs": {
    "enabled": true,
    "clientKind": "git",
    "useIgnoreFile": true
  },
  "files": {
    "ignoreUnknown": false,
    "ignore": ["node_modules", "dist", "build", ".next", "coverage"]
  },
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100,
    "lineEnding": "lf"
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "suspicious": {
        "noExplicitAny": "error",
        "noDoubleEquals": "error",
        "noDebugger": "error"
      },
      "style": {
        "useConst": "error",
        "useTemplate": "warn",
        "noNegationElse": "warn"
      },
      "correctness": {
        "noUnusedVariables": "error",
        "noUnusedImports": "error"
      },
      "complexity": {
        "noExcessiveCognitiveComplexity": "error"
      }
    }
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "trailingCommas": "es5",
      "semicolons": "asNeeded",
      "arrowParentheses": "always"
    }
  },
  "typescript": {
    "enabled": true
  }
}
```

### Common Commands

```bash
# Check all files (lint + format)
biome check .

# Check and auto-fix
biome check --write .

# Format only
biome format .

# Format and write
biome format --write .

# Lint only
biome lint .

# Lint and auto-fix
biome lint --write .

# Check specific files
biome check src/**/*.ts

# Show diagnostics without fixing
biome check --diagnostic-level=error .

# CI mode (no fixes)
biome ci .
```

---

## Package Management

### npm

```bash
# Initialize project
npm init -y

# Install dependencies
npm install                        # Install from package.json
npm install package-name           # Add production dependency
npm install --save-dev package-name  # Add dev dependency
npm install package@version        # Install specific version

# Update dependencies
npm update                         # Update all
npm update package-name            # Update specific package
npm outdated                       # Show outdated packages

# Remove dependencies
npm uninstall package-name

# Run scripts
npm run script-name
npm test
npm run build

# Audit security
npm audit
npm audit fix

# Clean install
npm ci                             # Clean install (CI/CD)
```

### pnpm (Faster Alternative)

```bash
# Install pnpm
npm install -g pnpm

# Initialize project
pnpm init

# Install dependencies
pnpm install                       # Install from package.json
pnpm add package-name              # Add production dependency
pnpm add -D package-name           # Add dev dependency

# Update dependencies
pnpm update                        # Update all
pnpm update package-name           # Update specific package

# Remove dependencies
pnpm remove package-name

# Run scripts
pnpm run script-name
pnpm test

# Audit security
pnpm audit

# Clean install
pnpm install --frozen-lockfile     # CI/CD
```

### package.json Scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest",
    "test:ci": "vitest run --coverage",
    "test:ui": "vitest --ui",
    "type-check": "tsc --noEmit",
    "lint": "biome check .",
    "lint:fix": "biome check --write .",
    "format": "biome format --write ."
  }
}
```

---

## React 19

### Installation

```bash
# Install React 19
npm install react@19.0.0 react-dom@19.0.0

# Install types
npm install --save-dev @types/react@19.0.0 @types/react-dom@19.0.0
```

### Key React 19 Features

**1. use() Hook**

```typescript
import { use } from 'react'

function Component({ dataPromise }: { dataPromise: Promise<Data> }) {
  const data = use(dataPromise) // Unwrap promise in render
  return <div>{data.title}</div>
}
```

**2. useActionState() Hook**

```typescript
import { useActionState } from 'react'

function Form() {
  const [state, formAction, isPending] = useActionState(
    async (prevState, formData) => {
      // Server action
      return { success: true }
    },
    null
  )

  return (
    <form action={formAction}>
      <button disabled={isPending}>Submit</button>
    </form>
  )
}
```

**3. useOptimistic() Hook**

```typescript
import { useOptimistic } from 'react'

function Component({ messages }: { messages: Message[] }) {
  const [optimisticMessages, addOptimisticMessage] = useOptimistic(
    messages,
    (state, newMessage) => [...state, newMessage]
  )

  return (
    <div>
      {optimisticMessages.map((msg) => (
        <div key={msg.id}>{msg.text}</div>
      ))}
    </div>
  )
}
```

---

## Vite 6.1

### Installation

```bash
# Create new Vite project
npm create vite@latest my-app -- --template react-ts

# Install Vite in existing project
npm install --save-dev vite@6.1.0
```

### Configuration (`vite.config.ts`)

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    port: 3000,
    open: true
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom']
        }
      }
    }
  }
})
```

### Commands

```bash
# Development server
vite

# Build for production
vite build

# Preview production build
vite preview

# Optimize dependencies
vite optimize
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: TypeScript CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Lint
        run: npm run lint

      - name: Type check
        run: npm run type-check

      - name: Test
        run: npm run test:ci

      - name: Build
        run: npm run build

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage/coverage-final.json
```

---

## TRUST 5 Principles for TypeScript

### T - Test First (Coverage ≥85%)

```bash
# Run tests with coverage
vitest run --coverage

# Check coverage threshold
vitest run --coverage --coverage.lines=85
```

### R - Readable

```bash
# Format code
biome format --write .

# Lint code
biome lint --write .
```

### U - Unified (Type Safety)

```typescript
// Enable strict mode
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true
  }
}
```

### S - Secured

```bash
# Audit dependencies
npm audit

# Fix vulnerabilities
npm audit fix
```

### T - Trackable

```typescript
// @TAG integration
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: authService.test.ts
export class AuthService {
  // Implementation
}
```

---

## Quick Reference Commands

```bash
# Development
npm install                        # Install dependencies
npm run dev                        # Start dev server
tsc --noEmit                       # Type check
biome check --write .              # Lint & format

# Testing
vitest                             # Run tests (watch)
vitest run                         # Run tests once
vitest --coverage                  # Run with coverage
vitest --ui                        # Open UI

# Building
npm run build                      # Build for production
vite preview                       # Preview build

# Quality
tsc --noEmit                       # Type check
biome check .                      # Lint & format check
npm audit                          # Security audit
```

---

**Version**: 1.0.0 (2025-10-22)
**Updated**: Latest tool versions verified 2025-10-22
**Framework**: MoAI-ADK TypeScript Language Skill
**Status**: Production-ready
