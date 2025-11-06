---
name: moai-lang-typescript
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "TypeScript best practices with modern frameworks, full-stack development, and type-safe patterns for 2025"
keywords: [typescript, javascript, frontend, backend, fullstack, react, nextjs, nodejs, type-safety]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# TypeScript Development Mastery

**Modern TypeScript Development with 2025 Best Practices**

> Comprehensive TypeScript development guidance covering full-stack applications, type-safe patterns, modern frameworks, and production-ready code using the latest tools and methodologies.

## What It Does

- **Full-Stack Development**: Next.js, React, Node.js with end-to-end type safety
- **Type-Safe APIs**: REST and GraphQL with TypeScript validation
- **Modern Frontend**: React 19+, Vue 3+, Angular 17+ with TypeScript
- **Backend Services**: Node.js, Deno, Bun with TypeScript optimization
- **Testing Strategies**: Vitest, Jest, Playwright with type-safe testing
- **Build & Deployment**: Modern bundlers, CI/CD, performance optimization
- **Code Quality**: ESLint, Prettier, automated type checking
- **Enterprise Patterns**: Monorepos, microservices, scalable architecture

## When to Use

### Perfect Scenarios
- **Building full-stack web applications**
- **Developing type-safe APIs and microservices**
- **Creating modern React/Next.js applications**
- **Enterprise-scale JavaScript applications**
- **Projects requiring strict type safety**
- **Teams with multiple JavaScript skill levels**
- **Applications with complex data models**

### Common Triggers
- "Create TypeScript API"
- "Set up Next.js project"
- "Type-safe React components"
- "TypeScript best practices"
- "Migrate JavaScript to TypeScript"
- "Full-stack TypeScript application"

## Tool Version Matrix (2025-11-06)

### Core TypeScript
- **TypeScript**: 5.6.x (latest) / 5.3.x (LTS)
- **Node.js**: 22.x (LTS) / 20.x (stable)
- **Package Managers**: pnpm 9.x, npm 10.x, yarn 4.x
- **Runtimes**: Node.js 22.x, Deno 2.x, Bun 1.1.x

### Frontend Frameworks
- **Next.js**: 15.x - Full-stack React framework
- **React**: 19.x - UI library with Server Components
- **Vue**: 3.5.x - Progressive framework
- **Angular**: 18.x - Enterprise framework
- **Svelte**: 5.x - Compiler-based framework

### Build Tools
- **Vite**: 6.x - Fast build tool
- **Webpack**: 5.x - Module bundler
- **esbuild**: 0.24.x - Fast bundler
- **SWC**: 1.7.x - Rust-based compiler

### Testing Tools
- **Vitest**: 2.x - Fast unit testing
- **Jest**: 30.x - Comprehensive testing
- **Playwright**: 1.48.x - E2E testing
- **Testing Library**: 16.x - Component testing

### Code Quality
- **ESLint**: 9.x - Linting (flat config)
- **Prettier**: 3.3.x - Code formatting
- **TypeScript Compiler**: Built-in type checking
- **Husky**: 9.x - Git hooks

### Backend Frameworks
- **Express**: 4.21.x - Minimal web framework
- **Fastify**: 5.x - Fast web framework
- **NestJS**: 10.x - Enterprise framework
- **tRPC**: 11.x - End-to-end typesafe APIs

## Ecosystem Overview

### Project Setup (2025 Best Practice)

```bash
# Modern TypeScript with pnpm (recommended)
pnpm create next-app@latest my-app --typescript --tailwind --eslint --app
cd my-app
pnpm add @types/node @types/react @types/react-dom
pnpm add -D vitest @vitest/ui jsdom @testing-library/react

# Alternative: Bun for ultra-fast development
bun create next-app my-app
cd my-app
bun add vitest @testing-library/react -d

# Monorepo with Nx or Turborepo
npx create-nx-workspace@latest my-workspace --preset=react-monorepo
# or
npx create-turbo@latest my-turbo-app
```

### Modern Project Structure

```
my-typescript-project/
├── package.json
├── tsconfig.json              # TypeScript configuration
├── tsconfig.build.json        # Build-specific config
├── vite.config.ts             # Build tool configuration
├── eslint.config.js           # ESLint flat config
├── prettier.config.js         # Prettier configuration
├── playwright.config.ts       # E2E testing config
├── vitest.config.ts           # Unit testing config
├── src/
│   ├── app/                   # Next.js app router
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   ├── globals.css
│   │   └── api/               # API routes
│   ├── components/            # Reusable components
│   │   ├── ui/                # UI components
│   │   └── features/          # Feature components
│   ├── lib/                   # Utilities and configs
│   │   ├── utils.ts
│   │   ├── db.ts              # Database configuration
│   │   └── types.ts           # Global type definitions
│   ├── hooks/                 # Custom React hooks
│   └── types/                 # Domain-specific types
├── tests/
│   ├── __mocks__/             # Mock files
│   ├── setup.ts               # Test setup
│   └── integration/           # Integration tests
└── .github/
    └── workflows/             # CI/CD pipelines
```

## Modern TypeScript Patterns

### Advanced Type System Features (TypeScript 5.6)

```typescript
// Import attributes for better module management
import data from "./data.json" with { type: "json" };

// Template literal type improvements
type EventName<T extends string> = `on${Capitalize<T>}`;
type UserEvent = EventName<'click' | 'hover'>; // 'onClick' | 'onHover'

// Enhanced symbol key handling
const metaKey = Symbol('metadata');
interface Metadata {
  [metaKey]: string;
}

// Improved decorator metadata
function Entity<T extends { new(...args: any[]): {} }>(constructor: T) {
  return class extends constructor {
    createdAt = new Date();
  };
}

@Entity
class User {
  constructor(public name: string) {}
}
```

### Utility Types and Patterns

```typescript
// Advanced utility types
type StrictOmit<T, K extends keyof T> = T extends Record<string, unknown>
  ? Pick<T, Exclude<keyof T, K>>
  : never;

type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

type Branded<T, B> = T & { __brand: B };

// Type-safe IDs
type UserId = Branded<string, 'UserId'>;
type ProductId = Branded<string, 'ProductId'>;

function createUserId(id: string): UserId {
  return id as UserId;
}

// Type-safe API client
type ApiResponse<T, E = unknown> = 
  | { success: true; data: T }
  | { success: false; error: E };

interface User {
  id: UserId;
  name: string;
  email: string;
}

async function fetchUser(id: UserId): Promise<ApiResponse<User>> {
  try {
    const response = await fetch(`/api/users/${id}`);
    if (!response.ok) throw new Error('Failed to fetch');
    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    return { success: false, error };
  }
}
```

### React Patterns with TypeScript

```typescript
// Type-safe component props with forwardRef
import React, { forwardRef, useImperativeHandle } from 'react';

interface ButtonProps {
  variant: 'primary' | 'secondary' | 'danger';
  size: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  children: React.ReactNode;
  onClick?: () => void;
}

export interface ButtonRef {
  click: () => void;
  focus: () => void;
}

export const Button = forwardRef<ButtonRef, ButtonProps>(
  ({ variant, size, disabled, children, onClick }, ref) => {
    const buttonRef = React.useRef<HTMLButtonElement>(null);

    useImperativeHandle(ref, () => ({
      click: () => buttonRef.current?.click(),
      focus: () => buttonRef.current?.focus(),
    }));

    const baseClasses = 'font-semibold rounded-lg transition-colors';
    const variantClasses = {
      primary: 'bg-blue-600 text-white hover:bg-blue-700',
      secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300',
      danger: 'bg-red-600 text-white hover:bg-red-700',
    };
    const sizeClasses = {
      sm: 'px-3 py-1 text-sm',
      md: 'px-4 py-2 text-base',
      lg: 'px-6 py-3 text-lg',
    };

    return (
      <button
        ref={buttonRef}
        className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]}`}
        disabled={disabled}
        onClick={onClick}
      >
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
```

### Custom Hooks with TypeScript

```typescript
// Type-safe data fetching hook
interface UseApiResult<T, E = unknown> {
  data: T | null;
  error: E | null;
  loading: boolean;
  refetch: () => Promise<void>;
}

function useApi<T, E = unknown>(
  url: string,
  options?: RequestInit
): UseApiResult<T, E> {
  const [data, setData] = React.useState<T | null>(null);
  const [error, setError] = React.useState<E | null>(null);
  const [loading, setLoading] = React.useState(false);

  const fetchData = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(url, options);
      if (!response.ok) throw new Error(response.statusText);
      const result = await response.json() as T;
      setData(result);
    } catch (err) {
      setError(err as E);
    } finally {
      setLoading(false);
    }
  }, [url, options]);

  React.useEffect(() => {
    fetchData();
  }, [fetchData]);

  return { data, error, loading, refetch: fetchData };
}

// Usage
interface User {
  id: number;
  name: string;
  email: string;
}

function UserProfile({ userId }: { userId: number }) {
  const { data: user, loading, error } = useApi<User>(`/api/users/${userId}`);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  if (!user) return <div>No user found</div>;

  return (
    <div>
      <h1>{user.name}</h1>
      <p>{user.email}</p>
    </div>
  );
}
```

## Performance Optimization

### Code Splitting and Lazy Loading

```typescript
// Dynamic imports with type safety
import { lazy, Suspense } from 'react';

const HeavyComponent = lazy(() => 
  import('./HeavyComponent').then(module => ({
    default: module.HeavyComponent
  }))
);

// Route-based code splitting
import { createBrowserRouter } from 'react-router-dom';

const router = createBrowserRouter([
  {
    path: '/',
    element: <HomePage />,
  },
  {
    path: '/dashboard',
    element: (
      <Suspense fallback={<div>Loading dashboard...</div>}>
        <LazyDashboard />
      </Suspense>
    ),
    loader: () => import('./loaders/dashboardLoader').then(m => m.dashboardLoader()),
  },
]);

// Preloading strategies
const preloadComponent = () => {
  import('./HeavyComponent');
};

// Preload on hover
<button onMouseEnter={preloadComponent}>
  Load Component
</button>
```

### Bundle Optimization

```typescript
// vite.config.ts - Modern build configuration
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig({
  plugins: [
    react(),
    process.env.ANALYZE && visualizer({
      filename: 'dist/stats.html',
      open: true,
    }),
  ],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['react-router-dom'],
          ui: ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu'],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
  },
  server: {
    hmr: {
      overlay: true,
    },
  },
});
```

### Memory Management

```typescript
// Efficient state management with useReducer
interface State {
  items: Item[];
  filter: string;
  loading: boolean;
}

type Action =
  | { type: 'SET_ITEMS'; payload: Item[] }
  | { type: 'SET_FILTER'; payload: string }
  | { type: 'SET_LOADING'; payload: boolean };

function itemsReducer(state: State, action: Action): State {
  switch (action.type) {
    case 'SET_ITEMS':
      return { ...state, items: action.payload, loading: false };
    case 'SET_FILTER':
      return { ...state, filter: action.payload };
    case 'SET_LOADING':
      return { ...state, loading: action.payload };
    default:
      return state;
  }
}

// Memoized components for performance
import { memo, useMemo, useCallback } from 'react';

interface ExpensiveComponentProps {
  data: number[];
  onItemClick: (id: number) => void;
}

const ExpensiveComponent = memo<ExpensiveComponentProps>(({ data, onItemClick }) => {
  const processedData = useMemo(() => {
    return data.map(item => ({
      ...item,
      processed: expensiveCalculation(item),
    }));
  }, [data]);

  const handleClick = useCallback((id: number) => {
    onItemClick(id);
  }, [onItemClick]);

  return (
    <div>
      {processedData.map(item => (
        <Item key={item.id} item={item} onClick={handleClick} />
      ))}
    </div>
  );
});
```

## Testing Strategies

### Vitest Configuration

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react-swc';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'tests/',
        '**/*.d.ts',
        '**/*.config.*',
        'dist/',
      ],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
```

### Type-Safe Testing Patterns

```typescript
// Component testing with Testing Library
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, type MockedFunction } from 'vitest';
import { Button } from './Button';

describe('Button', () => {
  it('renders with correct variant styles', () => {
    render(
      <Button variant="primary" size="md" onClick={vi.fn()}>
        Click me
      </Button>
    );

    const button = screen.getByRole('button', { name: 'Click me' });
    expect(button).toBeInTheDocument();
    expect(button).toHaveClass('bg-blue-600', 'text-white');
  });

  it('handles click events', async () => {
    const handleClick = vi.fn();
    render(
      <Button variant="secondary" size="sm" onClick={handleClick}>
        Submit
      </Button>
    );

    const button = screen.getByRole('button', { name: 'Submit' });
    fireEvent.click(button);

    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('can be triggered via ref', () => {
    const handleClick = vi.fn();
    const ref = { current: null };

    render(
      <Button variant="primary" size="md" onClick={handleClick} ref={ref}>
        Click via ref
      </Button>
    );

    // TypeScript should infer the correct ref type
    if (ref.current) {
      ref.current.click();
    }

    expect(handleClick).toHaveBeenCalledTimes(1);
  });
});

// API testing with MSW
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { fetchUser } from './api';

const server = setupServer(
  rest.get('/api/users/1', (req, res, ctx) => {
    return res(
      ctx.json({
        id: 1,
        name: 'John Doe',
        email: 'john@example.com',
      })
    );
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('fetchUser', () => {
  it('fetches user data successfully', async () => {
    const userId = createUserId('1');
    const result = await fetchUser(userId);

    if (result.success) {
      expect(result.data.name).toBe('John Doe');
      expect(result.data.email).toBe('john@example.com');
    } else {
      fail('Expected successful fetch');
    }
  });
});
```

### E2E Testing with Playwright

```typescript
// tests/e2e/user-journey.spec.ts
import { test, expect, type Page } from '@playwright/test';

test.describe('User Authentication Flow', () => {
  test('should allow user to sign up and log in', async ({ page }) => {
    // Sign up
    await page.goto('/signup');
    
    await page.fill('[data-testid=email-input]', 'user@example.com');
    await page.fill('[data-testid=password-input]', 'securepassword123');
    await page.fill('[data-testid=name-input]', 'Test User');
    
    await page.click('[data-testid=signup-button]');
    
    // Should redirect to dashboard
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('[data-testid=user-name]')).toHaveText('Test User');
    
    // Log out
    await page.click('[data-testid=logout-button]');
    
    // Log in
    await page.goto('/login');
    await page.fill('[data-testid=email-input]', 'user@example.com');
    await page.fill('[data-testid=password-input]', 'securepassword123');
    await page.click('[data-testid=login-button]');
    
    // Should be back at dashboard
    await expect(page).toHaveURL('/dashboard');
  });

  test('should show validation errors for invalid input', async ({ page }) => {
    await page.goto('/signup');
    
    // Try to submit with empty fields
    await page.click('[data-testid=signup-button]');
    
    await expect(page.locator('[data-testid=email-error]')).toBeVisible();
    await expect(page.locator('[data-testid=password-error]')).toBeVisible();
    await expect(page.locator('[data-testid=name-error]')).toBeVisible();
  });
});
```

## Security Best Practices

### Input Validation and Sanitization

```typescript
// Type-safe validation with Zod
import { z } from 'zod';

const UserSchema = z.object({
  name: z.string().min(1).max(100).trim(),
  email: z.string().email(),
  age: z.number().min(13).max(120),
  bio: z.string().max(1000).optional(),
});

type User = z.infer<typeof UserSchema>;

// API route validation
export async function POST(request: Request) {
  try {
    const body = await request.json();
    const validatedUser = UserSchema.parse(body);
    
    // Process validated data
    await createUser(validatedUser);
    
    return Response.json({ success: true }, { status: 201 });
  } catch (error) {
    if (error instanceof z.ZodError) {
      return Response.json(
        { 
          success: false, 
          errors: error.flatten().fieldErrors 
        },
        { status: 400 }
      );
    }
    
    return Response.json(
      { success: false, error: 'Internal server error' },
      { status: 500 }
    );
  }
}

// XSS prevention
function sanitizeHtml(input: string): string {
  // Use DOMPurify or similar library
  return input.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');
}

// CSRF protection
function getCSRFToken(): string {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta?.getAttribute('content') || '';
}

// Secure API client
class SecureApiClient {
  private baseUrl: string;
  private csrfToken: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
    this.csrfToken = getCSRFToken();
  }

  async post<T>(endpoint: string, data: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.csrfToken,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json() as Promise<T>;
  }
}
```

### Authentication and Authorization

```typescript
// JWT-based authentication
interface JWTPayload {
  userId: string;
  email: string;
  role: 'user' | 'admin';
  exp: number;
}

class AuthService {
  private static readonly TOKEN_KEY = 'auth_token';

  static setToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
  }

  static getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  static removeToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
  }

  static isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1])) as JWTPayload;
      return Date.now() >= payload.exp * 1000;
    } catch {
      return true;
    }
  }

  static getCurrentUser(): JWTPayload | null {
    const token = this.getToken();
    if (!token || this.isTokenExpired(token)) {
      return null;
    }

    try {
      return JSON.parse(atob(token.split('.')[1])) as JWTPayload;
    } catch {
      return null;
    }
  }
}

// Protected routes
interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: 'user' | 'admin';
}

function ProtectedRoute({ children, requiredRole = 'user' }: ProtectedRouteProps) {
  const user = AuthService.getCurrentUser();
  
  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (requiredRole === 'admin' && user.role !== 'admin') {
    return <Navigate to="/unauthorized" replace />;
  }

  return <>{children}</>;
}

// API middleware
function withAuth<T extends unknown[], R>(
  handler: (...args: T) => Promise<R>
) {
  return async (...args: T): Promise<R> => {
    const token = AuthService.getToken();
    
    if (!token || AuthService.isTokenExpired(token)) {
      throw new Error('Unauthorized');
    }

    return handler(...args);
  };
}
```

## Integration Patterns

### Full-Stack Type Safety with tRPC

```typescript
// server/router.ts
import { initTRPC } from '@trpc/server';
import { z } from 'zod';
import type { CreateHTTPContextOptions } from '@trpc/server/adapters/standalone';

const t = initTRPC.context<CreateHTTPContextOptions>().create();

export const appRouter = t.router({
  getUser: t.procedure
    .input(z.object({ id: z.string() }))
    .query(async ({ input, ctx }) => {
      const user = await ctx.db.user.findUnique({
        where: { id: input.id },
      });
      return user;
    }),
    
  createUser: t.procedure
    .input(z.object({
      name: z.string().min(1),
      email: z.string().email(),
    }))
    .mutation(async ({ input, ctx }) => {
      const user = await ctx.db.user.create({
        data: input,
      });
      return user;
    }),
});

export type AppRouter = typeof appRouter;

// client/trpc.ts
import { createTRPCProxyClient, httpBatchLink } from '@trpc/client';
import type { AppRouter } from '../server/router';

const trpc = createTRPCProxyClient<AppRouter>({
  links: [
    httpBatchLink({
      url: 'http://localhost:3000/trpc',
    }),
  ],
});

// Usage in React component
function UserProfile({ userId }: { userId: string }) {
  const { data: user, isLoading } = trpc.getUser.useQuery({ id: userId });
  const createUser = trpc.createUser.useMutation();

  if (isLoading) return <div>Loading...</div>;

  return (
    <div>
      <h1>{user?.name}</h1>
      <p>{user?.email}</p>
      <button 
        onClick={() => createUser.mutate({
          name: 'New User',
          email: 'new@example.com',
        })}
      >
        Create User
      </button>
    </div>
  );
}
```

### Database Integration with Prisma

```typescript
// schema.prisma
model User {
  id        String   @id @default(cuid())
  email     String   @unique
  name      String
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
  posts     Post[]
}

model Post {
  id        String   @id @default(cuid())
  title     String
  content   String
  published Boolean  @default(false)
  author    User     @relation(fields: [authorId], references: [id])
  authorId  String
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
}

// lib/db.ts
import { PrismaClient } from '@prisma/client';

const globalForPrisma = globalThis as unknown as {
  prisma: PrismaClient | undefined;
};

export const prisma = globalForPrisma.prisma ?? new PrismaClient();

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = prisma;

// API routes with type safety
import { z } from 'zod';

const createPostSchema = z.object({
  title: z.string().min(1).max(200),
  content: z.string().min(1),
  authorId: z.string().cuid(),
});

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const { title, content, authorId } = createPostSchema.parse(body);
    
    const post = await prisma.post.create({
      data: {
        title,
        content,
        authorId,
      },
      include: {
        author: true,
      },
    });
    
    return Response.json(post, { status: 201 });
  } catch (error) {
    if (error instanceof z.ZodError) {
      return Response.json(
        { errors: error.flatten().fieldErrors },
        { status: 400 }
      );
    }
    
    return Response.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
```

### State Management with Zustand

```typescript
// store/userStore.ts
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

interface User {
  id: string;
  name: string;
  email: string;
}

interface UserStore {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  updateProfile: (updates: Partial<User>) => void;
}

export const useUserStore = create<UserStore>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        isLoading: false,
        
        login: async (email: string, password: string) => {
          set({ isLoading: true });
          try {
            const response = await fetch('/api/auth/login', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ email, password }),
            });
            
            if (response.ok) {
              const user = await response.json();
              set({ user, isLoading: false });
            } else {
              throw new Error('Login failed');
            }
          } catch (error) {
            set({ isLoading: false });
            throw error;
          }
        },
        
        logout: () => {
          set({ user: null });
        },
        
        updateProfile: (updates) => {
          const currentUser = get().user;
          if (currentUser) {
            set({ user: { ...currentUser, ...updates } });
          }
        },
      }),
      {
        name: 'user-storage',
        partialize: (state) => ({ user: state.user }),
      }
    )
  )
);

// Usage in components
function ProfilePage() {
  const { user, isLoading, updateProfile } = useUserStore();

  if (isLoading) return <div>Loading...</div>;
  if (!user) return <div>Please log in</div>;

  return (
    <div>
      <h1>Profile</h1>
      <p>Name: {user.name}</p>
      <p>Email: {user.email}</p>
      <button 
        onClick={() => updateProfile({ name: 'Updated Name' })}
      >
        Update Name
      </button>
    </div>
  );
}
```

## Modern Development Workflow

### Configuration Files

```json
// package.json
{
  "name": "my-typescript-app",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "next dev --turbopack",
    "build": "next build",
    "start": "next start",
    "lint": "eslint . --max-warnings 0",
    "lint:fix": "eslint . --fix",
    "type-check": "tsc --noEmit",
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui"
  },
  "dependencies": {
    "next": "^15.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "@trpc/client": "^11.0.0",
    "@trpc/server": "^11.0.0",
    "@trpc/react-query": "^11.0.0",
    "zod": "^3.23.0",
    "zustand": "^5.0.0"
  },
  "devDependencies": {
    "@types/node": "^22.0.0",
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "typescript": "^5.6.0",
    "eslint": "^9.0.0",
    "prettier": "^3.3.0",
    "vitest": "^2.0.0",
    "@testing-library/react": "^16.0.0",
    "playwright": "^1.48.0"
  }
}
```

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "ES2022"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@/components/*": ["./src/components/*"],
      "@/lib/*": ["./src/lib/*"],
      "@/types/*": ["./src/types/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

```javascript
// eslint.config.js (ESLint Flat Config)
import js from '@eslint/js';
import typescript from '@typescript-eslint/eslint-plugin';
import typescriptParser from '@typescript-eslint/parser';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';

export default [
  js.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: typescriptParser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
    plugins: {
      '@typescript-eslint': typescript,
      'react': react,
      'react-hooks': reactHooks,
    },
    rules: {
      ...typescript.configs.recommended.rules,
      '@typescript-eslint/no-unused-vars': 'error',
      '@typescript-eslint/no-explicit-any': 'warn',
      'react/react-in-jsx-scope': 'off',
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
  },
];
```

### CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        node-version: [20.x, 22.x]
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: 'pnpm'
      
      - name: Install pnpm
        uses: pnpm/action-setup@v3
        with:
          version: 9
      
      - name: Install dependencies
        run: pnpm install --frozen-lockfile
      
      - name: Type check
        run: pnpm type-check
      
      - name: Lint
        run: pnpm lint
      
      - name: Test
        run: pnpm test
      
      - name: E2E tests
        run: pnpm test:e2e
      
      - name: Build
        run: pnpm build
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage/lcov.info
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**TypeScript Target**: 5.6+ with latest language features  

This skill provides comprehensive TypeScript development guidance with 2025 best practices, covering everything from basic type safety to advanced full-stack patterns and enterprise-grade applications.
