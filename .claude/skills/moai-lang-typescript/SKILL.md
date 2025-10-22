---
name: moai-lang-typescript
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: TypeScript 5.8+ best practices with Vitest 3.x, Biome 2.0, React 19 patterns.
keywords: ['typescript', 'vitest', 'biome', 'react']
allowed-tools:
  - Read
  - Bash
---

# Lang TypeScript Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-typescript |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language |

---

## What It Does

TypeScript 5.8+ best practices with Vitest 3.x, Biome 2.0, and React 19 patterns. This skill provides comprehensive guidance for modern TypeScript development across web platforms, focusing on type safety, testing, code quality, and modern framework integration.

**Key capabilities**:
- ✅ TypeScript 5.8 strict mode and advanced types
- ✅ Vitest 3.x with React Testing Library
- ✅ Biome 2.0 type-aware linting and formatting
- ✅ React 19 patterns (hooks, suspense, concurrent features)
- ✅ Vite 6 build tooling
- ✅ Modern package management (pnpm/npm)
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-22)
- ✅ TDD workflow support with RED → GREEN → REFACTOR

---

## When to Use

**Automatic triggers**:
- TypeScript, React, or Vite file patterns detected (`.ts`, `.tsx`, `package.json`, `vite.config.ts`)
- SPEC implementation requiring TypeScript (`/alfred:2-run`)
- Code review requests for TypeScript codebases
- Quality gate validation during `/alfred:3-sync`

**Manual invocation**:
- Review TypeScript code for TRUST 5 compliance
- Design new React features with modern patterns
- Migrate to TypeScript 5.8 or React 19
- Troubleshoot type inference issues
- Optimize build performance with Vite

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status | Installation |
|------|---------|---------|--------|--------------|
| **TypeScript** | 5.8.3 | Compiler | ✅ Current | `pnpm add -D typescript` |
| **Vitest** | 3.2.0 | Testing | ✅ Current | `pnpm add -D vitest` |
| **Biome** | 2.0.0 | Linter/Formatter | ✅ Current | `pnpm add -D @biomejs/biome` |
| **React** | 19.1.0 | Framework | ✅ Current | `pnpm add react react-dom` |
| **Vite** | 6.0.0 | Build tool | ✅ Current | `pnpm add -D vite` |
| **React Testing Library** | 16.1.0 | Testing utilities | ✅ Current | `pnpm add -D @testing-library/react` |

**Compatibility matrix**:
- TypeScript 5.8+ requires Node.js 18+
- Vitest 3.x requires Vite 6 (also supports Vite 5)
- React 19 requires TypeScript 5.0+
- Biome 2.0 includes type-aware linter (no TypeScript compiler dependency)

---

## TypeScript 5.8 Core Principles

### 1. Strict Mode Configuration

**Always enable strict mode in `tsconfig.json`:**

```json
{
  "compilerOptions": {
    "strict": true,
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "esModuleInterop": true,
    "skipLibCheck": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "allowImportingTsExtensions": true,

    // Strict checks
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitReturns": true,
    "noImplicitOverride": true,

    // Path mapping
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
```

### 2. Type Safety Best Practices

```typescript
// ✅ GOOD: Explicit types for function parameters and returns
function calculateTotal(items: Item[], tax: number): number {
  return items.reduce((sum, item) => sum + item.price, 0) * (1 + tax);
}

// ✅ GOOD: Discriminated unions
type Result<T> =
  | { success: true; data: T }
  | { success: false; error: string };

function parseJSON(input: string): Result<unknown> {
  try {
    return { success: true, data: JSON.parse(input) };
  } catch (error) {
    return { success: false, error: String(error) };
  }
}

// ✅ GOOD: Type guards
function isUser(value: unknown): value is User {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    'name' in value
  );
}

// ❌ BAD: Using 'any'
function process(data: any) {  // Avoid!
  return data.something;
}

// ✅ BETTER: Use 'unknown' and validate
function process(data: unknown) {
  if (isUser(data)) {
    return data.name;
  }
  throw new Error('Invalid data');
}
```

### 3. Advanced Type Patterns

```typescript
// ✅ GOOD: Utility types
type UserProfile = {
  id: string;
  name: string;
  email: string;
  age: number;
};

// Make all properties optional
type PartialUserProfile = Partial<UserProfile>;

// Pick specific properties
type UserSummary = Pick<UserProfile, 'id' | 'name'>;

// Omit specific properties
type UserWithoutEmail = Omit<UserProfile, 'email'>;

// Make properties readonly
type ImmutableUser = Readonly<UserProfile>;

// ✅ GOOD: Mapped types
type Nullable<T> = { [K in keyof T]: T[K] | null };
type NullableUser = Nullable<UserProfile>;

// ✅ GOOD: Conditional types
type ExtractArray<T> = T extends (infer U)[] ? U : never;
type Item = ExtractArray<string[]>;  // string

// ✅ GOOD: Template literal types
type Event = 'click' | 'focus' | 'blur';
type EventHandler = `on${Capitalize<Event>}`;  // 'onClick' | 'onFocus' | 'onBlur'
```

### 4. Generic Constraints

```typescript
// ✅ GOOD: Generic with constraints
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

const user = { id: '123', name: 'Alice' };
const id = getProperty(user, 'id');  // Type: string

// ✅ GOOD: Generic factory pattern
interface Repository<T extends { id: string }> {
  findById(id: string): Promise<T | null>;
  save(entity: T): Promise<T>;
  delete(id: string): Promise<void>;
}

class UserRepository implements Repository<User> {
  async findById(id: string): Promise<User | null> {
    // Implementation
    return null;
  }

  async save(user: User): Promise<User> {
    // Implementation
    return user;
  }

  async delete(id: string): Promise<void> {
    // Implementation
  }
}
```

---

## Vitest 3.x Best Practices

### 1. Configuration (`vitest.config.ts`)

```typescript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/**/*.test.{ts,tsx}',
        'src/**/*.spec.{ts,tsx}',
        'src/test/**',
        'src/main.tsx',
      ],
      all: true,
      lines: 85,
      functions: 85,
      branches: 85,
      statements: 85,
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
```

### 2. Test Setup (`src/test/setup.ts`)

```typescript
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';

// Cleanup after each test
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});
```

### 3. React Component Testing

```typescript
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { UserProfile } from './UserProfile';

describe('UserProfile', () => {
  const mockUser = {
    id: '123',
    name: 'Alice',
    email: 'alice@example.com',
  };

  // ✅ GOOD: Arrange-Act-Assert pattern
  it('displays user information', () => {
    // Arrange
    render(<UserProfile user={mockUser} />);

    // Act
    const nameElement = screen.getByText('Alice');

    // Assert
    expect(nameElement).toBeInTheDocument();
  });

  // ✅ GOOD: Async testing with waitFor
  it('loads user data asynchronously', async () => {
    const mockFetch = vi.fn().mockResolvedValue(mockUser);
    render(<UserProfile userId="123" onLoad={mockFetch} />);

    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });

    expect(mockFetch).toHaveBeenCalledWith('123');
  });

  // ✅ GOOD: User interaction testing
  it('handles form submission', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<UserForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText('Name'), 'Bob');
    await user.click(screen.getByRole('button', { name: 'Submit' }));

    expect(onSubmit).toHaveBeenCalledWith({ name: 'Bob' });
  });
});
```

### 4. Mocking with Vitest

```typescript
import { vi } from 'vitest';

// ✅ GOOD: Mock entire module
vi.mock('./api/userService', () => ({
  fetchUser: vi.fn().mockResolvedValue({ id: '123', name: 'Alice' }),
  updateUser: vi.fn().mockResolvedValue(true),
}));

// ✅ GOOD: Partial mock with actual implementation
vi.mock('./utils/logger', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./utils/logger')>();
  return {
    ...actual,
    logError: vi.fn(),  // Mock only this function
  };
});

// ✅ GOOD: Mock timers
it('debounces input', async () => {
  vi.useFakeTimers();
  const onChange = vi.fn();
  render(<DebouncedInput onChange={onChange} delay={300} />);

  await userEvent.type(screen.getByRole('textbox'), 'hello');
  expect(onChange).not.toHaveBeenCalled();

  vi.advanceTimersByTime(300);
  expect(onChange).toHaveBeenCalledWith('hello');

  vi.useRealTimers();
});
```

---

## Biome 2.0 Configuration

### 1. `biome.json` Setup

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
    "ignore": ["node_modules", "dist", "build", "coverage"]
  },
  "formatter": {
    "enabled": true,
    "formatWithErrors": false,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineEnding": "lf",
    "lineWidth": 100
  },
  "organizeImports": {
    "enabled": true
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "complexity": {
        "noBannedTypes": "error",
        "noExtraBooleanCast": "error",
        "noMultipleSpacesInRegularExpressionLiterals": "error",
        "noUselessCatch": "error",
        "noUselessConstructor": "error",
        "noUselessLoneBlockStatements": "error",
        "noUselessRename": "error",
        "noUselessTernary": "error",
        "noVoid": "error",
        "noWith": "error",
        "useFlatMap": "error",
        "useOptionalChain": "error",
        "useSimplifiedLogicExpression": "error"
      },
      "correctness": {
        "noConstAssign": "error",
        "noConstantCondition": "error",
        "noEmptyCharacterClassInRegex": "error",
        "noEmptyPattern": "error",
        "noGlobalObjectCalls": "error",
        "noInvalidConstructorSuper": "error",
        "noInvalidNewBuiltin": "error",
        "noNonoctalDecimalEscape": "error",
        "noPrecisionLoss": "error",
        "noSelfAssign": "error",
        "noSetterReturn": "error",
        "noSwitchDeclarations": "error",
        "noUndeclaredVariables": "error",
        "noUnreachable": "error",
        "noUnreachableSuper": "error",
        "noUnsafeFinally": "error",
        "noUnsafeOptionalChaining": "error",
        "noUnusedLabels": "error",
        "noUnusedVariables": "error",
        "useIsNan": "error",
        "useValidForDirection": "error",
        "useYield": "error"
      },
      "style": {
        "noArguments": "error",
        "noVar": "error",
        "useConst": "error",
        "useTemplate": "error"
      },
      "suspicious": {
        "noAsyncPromiseExecutor": "error",
        "noCatchAssign": "error",
        "noClassAssign": "error",
        "noCompareNegZero": "error",
        "noControlCharactersInRegex": "error",
        "noDebugger": "warn",
        "noDoubleEquals": "error",
        "noDuplicateCase": "error",
        "noDuplicateClassMembers": "error",
        "noDuplicateObjectKeys": "error",
        "noDuplicateParameters": "error",
        "noEmptyBlockStatements": "error",
        "noExplicitAny": "warn",
        "noExtraNonNullAssertion": "error",
        "noFallthroughSwitchClause": "error",
        "noFunctionAssign": "error",
        "noGlobalAssign": "error",
        "noImportAssign": "error",
        "noMisleadingCharacterClass": "error",
        "noPrototypeBuiltins": "error",
        "noRedeclare": "error",
        "noSelfCompare": "error",
        "noShadowRestrictedNames": "error",
        "noUnsafeDeclarationMerging": "error",
        "noUnsafeNegation": "error",
        "useGetterReturn": "error",
        "useValidTypeof": "error"
      }
    }
  },
  "javascript": {
    "formatter": {
      "jsxQuoteStyle": "double",
      "quoteProperties": "asNeeded",
      "trailingCommas": "es5",
      "semicolons": "always",
      "arrowParentheses": "always",
      "bracketSpacing": true,
      "bracketSameLine": false,
      "quoteStyle": "single",
      "attributePosition": "auto"
    }
  },
  "overrides": [
    {
      "include": ["*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx"],
      "linter": {
        "rules": {
          "suspicious": {
            "noExplicitAny": "off"
          }
        }
      }
    }
  ]
}
```

### 2. NPM Scripts

```json
{
  "scripts": {
    "format": "biome format --write .",
    "lint": "biome lint .",
    "lint:fix": "biome lint --write .",
    "check": "biome check .",
    "check:fix": "biome check --write .",
    "ci": "biome ci ."
  }
}
```

---

## React 19 Patterns

### 1. Server Components & Actions (Next.js 15+)

```typescript
// app/users/page.tsx - Server Component
import { db } from '@/lib/db';

export default async function UsersPage() {
  // ✅ GOOD: Direct database access in Server Component
  const users = await db.user.findMany();

  return (
    <div>
      <h1>Users</h1>
      <UserList users={users} />
    </div>
  );
}

// ✅ GOOD: Server Action
'use server';

export async function createUser(formData: FormData) {
  const name = formData.get('name') as string;
  const email = formData.get('email') as string;

  await db.user.create({
    data: { name, email },
  });

  revalidatePath('/users');
}
```

### 2. Modern Hooks Patterns

```typescript
import { useState, useEffect, useCallback, useMemo, useTransition } from 'react';

// ✅ GOOD: Custom hook with TypeScript
function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}

// ✅ GOOD: useTransition for non-blocking updates
function SearchResults() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Result[]>([]);
  const [isPending, startTransition] = useTransition();

  const handleSearch = (value: string) => {
    setQuery(value);  // Immediate update

    startTransition(() => {
      // Non-blocking update
      fetchResults(value).then(setResults);
    });
  };

  return (
    <div>
      <input value={query} onChange={(e) => handleSearch(e.target.value)} />
      {isPending && <Spinner />}
      <ResultsList results={results} />
    </div>
  );
}

// ✅ GOOD: useMemo for expensive computations
function DataTable({ data }: { data: DataItem[] }) {
  const sortedData = useMemo(() => {
    return [...data].sort((a, b) => a.name.localeCompare(b.name));
  }, [data]);

  return <Table data={sortedData} />;
}

// ✅ GOOD: useCallback for stable references
function Parent() {
  const [count, setCount] = useState(0);

  const handleClick = useCallback(() => {
    setCount((c) => c + 1);
  }, []);

  return <ExpensiveChild onClick={handleClick} />;
}
```

### 3. Context & State Management

```typescript
import { createContext, useContext, useReducer, type ReactNode } from 'react';

// ✅ GOOD: Typed context
type AuthState = {
  user: User | null;
  isAuthenticated: boolean;
};

type AuthAction =
  | { type: 'LOGIN'; payload: User }
  | { type: 'LOGOUT' };

const AuthContext = createContext<{
  state: AuthState;
  dispatch: React.Dispatch<AuthAction>;
} | null>(null);

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'LOGIN':
      return { user: action.payload, isAuthenticated: true };
    case 'LOGOUT':
      return { user: null, isAuthenticated: false };
    default:
      return state;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, {
    user: null,
    isAuthenticated: false,
  });

  return (
    <AuthContext.Provider value={{ state, dispatch }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
```

### 4. Error Boundaries (Class Component)

```typescript
import { Component, type ReactNode, type ErrorInfo } from 'react';

type Props = {
  children: ReactNode;
  fallback?: ReactNode;
};

type State = {
  hasError: boolean;
  error: Error | null;
};

// ✅ GOOD: Error boundary with TypeScript
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error boundary caught:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || <div>Something went wrong</div>;
    }

    return this.props.children;
  }
}
```

---

## Project Structure

### Recommended File Organization (Vite + React)

```
my-app/
├── public/
│   └── assets/
├── src/
│   ├── app/                    # Next.js 15 App Router (if applicable)
│   │   ├── layout.tsx
│   │   └── page.tsx
│   ├── components/
│   │   ├── common/             # Reusable components
│   │   ├── features/           # Feature-specific components
│   │   └── layout/             # Layout components
│   ├── hooks/                  # Custom hooks
│   ├── lib/                    # Utilities, helpers
│   ├── services/               # API clients
│   ├── stores/                 # State management
│   ├── types/                  # Type definitions
│   ├── test/                   # Test utilities
│   │   └── setup.ts
│   ├── main.tsx                # Entry point
│   └── App.tsx
├── tests/                      # E2E tests (Playwright)
├── .env.example
├── biome.json
├── package.json
├── tsconfig.json
├── vite.config.ts
└── vitest.config.ts
```

---

## Package Management (pnpm)

### `package.json` Template

```json
{
  "name": "my-app",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest --coverage",
    "lint": "biome lint .",
    "lint:fix": "biome lint --write .",
    "format": "biome format --write .",
    "check": "biome check .",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0"
  },
  "devDependencies": {
    "@biomejs/biome": "^2.0.0",
    "@testing-library/jest-dom": "^6.6.3",
    "@testing-library/react": "^16.1.0",
    "@testing-library/user-event": "^14.5.2",
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "@vitejs/plugin-react": "^4.3.4",
    "@vitest/ui": "^3.2.0",
    "jsdom": "^25.0.1",
    "typescript": "^5.8.3",
    "vite": "^6.0.0",
    "vitest": "^3.2.0"
  }
}
```

### Installation Commands

```bash
# Install pnpm globally
npm install -g pnpm

# Initialize new project
pnpm create vite my-app --template react-ts
cd my-app

# Install dependencies
pnpm install

# Add Biome
pnpm add -D @biomejs/biome
pnpm biome init

# Add Vitest + React Testing Library
pnpm add -D vitest @vitest/ui jsdom
pnpm add -D @testing-library/react @testing-library/jest-dom @testing-library/user-event

# Run development server
pnpm dev
```

---

## Common Patterns

### 1. API Client with Typed Responses

```typescript
// lib/api.ts
type ApiResponse<T> =
  | { success: true; data: T }
  | { success: false; error: string };

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const data = await response.json();
      return { success: true, data };
    } catch (error) {
      return { success: false, error: String(error) };
    }
  }

  async post<T, B = unknown>(
    endpoint: string,
    body: B
  ): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const data = await response.json();
      return { success: true, data };
    } catch (error) {
      return { success: false, error: String(error) };
    }
  }
}

export const api = new ApiClient(import.meta.env.VITE_API_URL);
```

### 2. Form Handling with TypeScript

```typescript
import { useState, type FormEvent } from 'react';
import { z } from 'zod';

// ✅ GOOD: Zod schema for validation
const userSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Invalid email format'),
  age: z.number().min(18, 'Must be at least 18'),
});

type UserFormData = z.infer<typeof userSchema>;

function UserForm() {
  const [formData, setFormData] = useState<Partial<UserFormData>>({});
  const [errors, setErrors] = useState<Partial<Record<keyof UserFormData, string>>>({});

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();

    const result = userSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Partial<Record<keyof UserFormData, string>> = {};
      result.error.errors.forEach((err) => {
        const path = err.path[0] as keyof UserFormData;
        fieldErrors[path] = err.message;
      });
      setErrors(fieldErrors);
      return;
    }

    // Submit valid data
    console.log('Valid data:', result.data);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        name="name"
        value={formData.name || ''}
        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
      />
      {errors.name && <span>{errors.name}</span>}
      {/* ... */}
    </form>
  );
}
```

### 3. Typed Environment Variables

```typescript
// src/env.ts
import { z } from 'zod';

const envSchema = z.object({
  VITE_API_URL: z.string().url(),
  VITE_APP_NAME: z.string().min(1),
  VITE_ENABLE_ANALYTICS: z.enum(['true', 'false']).transform((val) => val === 'true'),
});

export const env = envSchema.parse(import.meta.env);

// Usage
console.log(env.VITE_API_URL);  // Type-safe!
```

---

## TRUST 5 Integration

### T - Test First (Vitest)

```bash
# Run all tests
pnpm test

# Run tests with UI
pnpm test:ui

# Coverage report
pnpm test:coverage

# Watch mode (default)
pnpm test --watch
```

**Target**: 85%+ test coverage.

### R - Readable (Biome)

```bash
# Lint only
pnpm lint

# Format only
pnpm format

# Check everything (lint + format)
pnpm check

# Auto-fix
pnpm check:fix
```

### U - Unified (TypeScript Strict Mode)

Enable all strict checks in `tsconfig.json`:
```json
{
  "compilerOptions": {
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true
  }
}
```

### S - Secured (Static Analysis)

```bash
# Type checking
pnpm typecheck

# Audit dependencies
pnpm audit

# Use Biome security rules
# (Biome 2.0 includes security-focused linting)
```

### T - Trackable (@TAG System)

```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: auth.test.ts
export class AuthService {
  // Implementation tied to SPEC-AUTH-001
}
```

---

## Performance Optimization

### 1. Code Splitting (Vite)

```typescript
import { lazy, Suspense } from 'react';

// ✅ GOOD: Lazy load heavy components
const Dashboard = lazy(() => import('./Dashboard'));
const Settings = lazy(() => import('./Settings'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Suspense>
  );
}
```

### 2. Vite Build Optimization

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          utils: ['lodash-es', 'date-fns'],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
  },
});
```

### 3. React Performance

```typescript
import { memo } from 'react';

// ✅ GOOD: Memoize expensive components
export const ExpensiveComponent = memo(function ExpensiveComponent({
  data,
}: {
  data: Data[];
}) {
  return <div>{/* Expensive render logic */}</div>;
});
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: TypeScript CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: pnpm/action-setup@v4
        with:
          version: 9

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Type check
        run: pnpm typecheck

      - name: Lint
        run: pnpm lint

      - name: Test
        run: pnpm test:coverage

      - name: Build
        run: pnpm build

      - name: Upload coverage
        uses: codecov/codecov-action@v5
        with:
          files: ./coverage/coverage-final.json
```

---

## References (Latest Documentation)

**Official Resources** (Updated 2025-10-22):
- TypeScript Handbook: https://www.typescriptlang.org/docs/
- Vitest Documentation: https://vitest.dev/
- Biome Documentation: https://biomejs.dev/
- React Documentation: https://react.dev/
- Vite Documentation: https://vite.dev/

**Community Resources**:
- TypeScript Discord: https://discord.gg/typescript
- React Community: https://react.dev/community
- Awesome TypeScript: https://github.com/dzharii/awesome-typescript

---

## Changelog

- **v2.0.0** (2025-10-22): Major expansion with TypeScript 5.8, Vitest 3.x, Biome 2.0, React 19, 1,200+ lines, comprehensive patterns, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates, TRUST 5 validation)
- `moai-alfred-code-reviewer` (code review automation)
- `moai-essentials-debug` (debugging support for TypeScript)
- `moai-domain-frontend` (React/UI patterns)
- `moai-domain-web-api` (API integration patterns)

---

## Best Practices Summary

✅ **DO**:
- Enable TypeScript strict mode
- Use Vitest for fast, modern testing
- Configure Biome 2.0 for linting and formatting
- Maintain 85%+ test coverage
- Leverage React 19 concurrent features
- Use type guards instead of 'any'
- Follow functional programming principles
- Write composable, reusable components

❌ **DON'T**:
- Use 'any' type liberally
- Skip type annotations on public APIs
- Ignore TypeScript errors
- Mix JavaScript and TypeScript without justification
- Skip testing async code
- Use class components for new code (prefer hooks)
- Mutate props or state directly
- Over-optimize prematurely

---

**End of Skill** | Total: 1,280+ lines
