# TypeScript 5.9 API Reference

Complete reference for TypeScript 5.9.3 development with November 2025 tool versions and best practices.

---

## Tool Versions (November 2025)

| Tool | Version | Release | Support |
|------|---------|---------|---------|
| **TypeScript** | 5.9.3 | Aug 2025 | Active |
| **Node.js LTS** | 22.11.0 | Oct 2024 | Apr 2027 |
| **React** | 19.x | Dec 2024 | Active |
| **Next.js** | 16.x | 2025 | Active |
| **tRPC** | 11.x | 2024 | Active |
| **Zod** | 3.23.x | 2025 | Active |
| **Vitest** | 2.x | 2024 | Active |
| **Turbopack** | Latest | 2025 | Active |

---

## TypeScript 5.9 Type System API

### Primitive Type Annotations
```typescript
// Strings
const name: string = "John";
const description: string = `Hello ${name}`;

// Numbers
const age: number = 30;
const pi: number = 3.14159;

// Booleans
const isActive: boolean = true;

// Arrays
const numbers: number[] = [1, 2, 3];
const strings: Array<string> = ["a", "b"];
const tuples: [string, number] = ["hello", 42];

// Any (avoid when possible)
const unknown: any = 42;

// Unknown (preferred over any)
const value: unknown = 42;
```

### Union & Intersection Types
```typescript
// Union types
type Status = "pending" | "approved" | "rejected";
type Result = string | number | boolean;

// Intersection types (combine multiple types)
type Admin = User & { role: "admin"; permissions: string[] };

// Discriminated unions
type SuccessResponse = { status: 'success'; data: unknown };
type ErrorResponse = { status: 'error'; error: string };
type Response = SuccessResponse | ErrorResponse;
```

### Generic Types
```typescript
// Generic interface
interface Container<T> {
  value: T;
  getValue(): T;
  setValue(value: T): void;
}

// Generic function
function getFirstElement<T>(arr: T[]): T {
  return arr[0];
}

// Generic constraints
function merge<T extends object, U extends object>(obj1: T, obj2: U): T & U {
  return { ...obj1, ...obj2 };
}
```

### Utility Types
```typescript
// Partial<T> - Make all properties optional
type PartialUser = Partial<User>;

// Pick<T, K> - Select specific properties
type UserPreview = Pick<User, 'name' | 'email'>;

// Omit<T, K> - Exclude specific properties
type UserWithoutPassword = Omit<User, 'password'>;

// Record<K, T> - Object with specific keys and value type
type RolePermissions = Record<'admin' | 'user' | 'guest', string[]>;

// Readonly<T> - Make all properties readonly
type ReadonlyUser = Readonly<User>;

// ReturnType<F> - Extract return type of function
type GetUserReturn = ReturnType<typeof getUser>;

// Parameters<F> - Extract parameter types of function
type GetUserParams = Parameters<typeof getUser>;
```

### Conditional Types
```typescript
// Basic conditional type
type IsString<T> = T extends string ? true : false;

// Extract array element type
type Flatten<T> = T extends Array<infer U> ? U : T;

// Nested conditionals
type Awaited<T> = T extends Promise<infer U> ? Awaited<U> : T;
```

### Mapped Types
```typescript
// Convert all properties to getters
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

// Make all properties readonly
type ReadonlyProps<T> = {
  readonly [K in keyof T]: T[K];
};

// Convert all properties to optional
type Optional<T> = {
  [K in keyof T]?: T[K];
};
```

---

## React 19 API

### Hooks
```typescript
// useState - Manage state
const [count, setCount] = useState<number>(0);

// useEffect - Side effects
useEffect(() => {
  console.log('Component mounted');
  return () => console.log('Component unmounted');
}, []);

// useContext - Access context
const theme = useContext(ThemeContext);

// useReducer - Complex state logic
const [state, dispatch] = useReducer(reducer, initialState);

// useCallback - Memoize function
const handleClick = useCallback(() => { /* ... */ }, []);

// useMemo - Memoize value
const memoizedValue = useMemo(() => expensiveCalculation(), [deps]);

// useTransition - Non-blocking state updates
const [isPending, startTransition] = useTransition();

// useRef - Access DOM elements
const inputRef = useRef<HTMLInputElement>(null);

// use - Unwrap promises (React 19)
const data = use(fetchData());
```

### Component Types
```typescript
// Function component
type Props = { name: string; age?: number };
export const User: React.FC<Props> = ({ name, age }) => (
  <div>{name}, {age}</div>
);

// Component with ref
type RefProps = { name: string; ref?: React.Ref<HTMLDivElement> };
export const UserWithRef = React.forwardRef<HTMLDivElement, RefProps>(
  ({ name }, ref) => <div ref={ref}>{name}</div>
);

// Component with children
type LayoutProps = { children: React.ReactNode };
export const Layout: React.FC<LayoutProps> = ({ children }) => (
  <div>{children}</div>
);
```

---

## Next.js 16 API

### File-Based Routing
```
app/
├── page.tsx           # / (root page)
├── layout.tsx         # Root layout
├── dashboard/
│   ├── page.tsx       # /dashboard
│   ├── layout.tsx     # Dashboard layout
│   └── [id]/
│       └── page.tsx   # /dashboard/[id]
├── api/
│   └── users/
│       ├── route.ts   # /api/users (GET, POST)
│       └── [id]/
│           └── route.ts  # /api/users/[id]
└── (auth)/            # Route group (doesn't affect URL)
    ├── login/
    │   └── page.tsx   # /login
    └── register/
        └── page.tsx   # /register
```

### API Route Handlers
```typescript
// GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
export async function GET(request: NextRequest) {
  return NextResponse.json({ data: [] });
}

export async function POST(request: NextRequest) {
  const data = await request.json();
  return NextResponse.json(data, { status: 201 });
}

// Dynamic routes
export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  return NextResponse.json({ id: params.id });
}
```

### Server Functions
```typescript
'use server'

// Server action
export async function updateUser(formData: FormData) {
  const name = formData.get('name');
  // Database operation
  revalidatePath('/users');
  redirect('/users');
}

// Revalidate data
export async function revalidateUsers() {
  revalidatePath('/users');
  revalidateTag('users');
}
```

### Environment Variables
```typescript
// Public variables (exposed to browser)
process.env.NEXT_PUBLIC_API_URL

// Private variables (server-only)
process.env.API_SECRET
process.env.DATABASE_URL
```

---

## tRPC API

### Router Definition
```typescript
// Initialize tRPC
const t = initTRPC.context<Context>().create();

// Define router
export const router = t.router({
  // Query procedure
  getUser: t.procedure
    .input(z.object({ id: z.string() }))
    .query(async ({ input, ctx }) => {
      return await db.user.findUnique({ where: { id: input.id } });
    }),

  // Mutation procedure
  createUser: t.procedure
    .input(z.object({ name: z.string() }))
    .mutation(async ({ input, ctx }) => {
      return await db.user.create({ data: input });
    }),

  // Nested router
  admin: t.router({
    deleteUser: t.procedure
      .input(z.object({ id: z.string() }))
      .mutation(async ({ input }) => {
        return await db.user.delete({ where: { id: input.id } });
      })
  })
});

export type AppRouter = typeof router;
```

### Client Hooks
```typescript
// Query hook
const { data, isLoading, error } = trpc.getUser.useQuery({ id: '1' });

// Mutation hook
const createUserMutation = trpc.createUser.useMutation();
createUserMutation.mutate({ name: 'John' });

// Invalidate query cache
const utils = trpc.useUtils();
await utils.getUser.invalidate();

// Set query data
utils.getUser.setData({ id: '1' }, { name: 'John', email: 'john@example.com' });
```

---

## Zod Validation API

### Schema Types
```typescript
// Primitives
z.string()
z.number()
z.boolean()
z.date()
z.enum(['a', 'b', 'c'])

// Collections
z.array(z.string())
z.object({ name: z.string() })
z.record(z.string(), z.number())
z.set(z.string())
z.map(z.string(), z.number())
z.tuple([z.string(), z.number()])

// Utility
z.union([z.string(), z.number()])
z.discriminatedUnion('type', [/* ... */])
z.lazy(() => schema)  // Circular references
```

### Validation Methods
```typescript
// Parse (throws on error)
const user = UserSchema.parse(data);

// SafeParse (returns result object)
const result = UserSchema.safeParse(data);
if (result.success) {
  console.log(result.data);
} else {
  console.log(result.error);
}

// Type inference
type User = z.infer<typeof UserSchema>;
```

### String Validations
```typescript
z.string()
  .min(3, 'Minimum 3 characters')
  .max(50, 'Maximum 50 characters')
  .email('Invalid email')
  .url('Invalid URL')
  .regex(/pattern/, 'Invalid format')
  .transform(val => val.toLowerCase())
  .refine(val => !val.includes('bad'), 'Contains banned word')
```

---

## Testing with Vitest

### Test Syntax
```typescript
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

describe('Math', () => {
  beforeEach(() => { /* setup */ });
  afterEach(() => { /* cleanup */ });

  it('adds numbers', () => {
    expect(1 + 1).toBe(2);
  });

  it('throws on invalid input', () => {
    expect(() => divide(1, 0)).toThrow();
  });
});
```

### Matchers
```typescript
// Equality
expect(value).toBe(expected);          // ===
expect(value).toEqual(expected);       // deep equal

// Truthiness
expect(value).toBeTruthy();
expect(value).toBeFalsy();
expect(value).toBeNull();
expect(value).toBeUndefined();

// Numbers
expect(value).toBeGreaterThan(5);
expect(value).toBeLessThanOrEqual(10);

// Strings
expect(text).toMatch(/pattern/);
expect(text).toContain('substring');

// Functions
expect(fn).toHaveBeenCalled();
expect(fn).toHaveBeenCalledWith('arg');
expect(fn).toHaveBeenCalledTimes(3);
```

### Mocking
```typescript
// Mock function
const mockFn = vi.fn();
const mockFn = vi.fn((x) => x * 2);

// Mock module
vi.mock('@/lib/db', () => ({
  getUser: vi.fn()
}));

// Mock return value
mockFn.mockReturnValue(42);
mockFn.mockResolvedValue(data);
mockFn.mockRejectedValue(new Error('Failed'));
```

---

## tsconfig.json Essential Options

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "baseUrl": ".",
    "paths": {
      "@/*": ["./*"]
    },
    "allowSyntheticDefaultImports": true
  },
  "include": ["src", "next-env.d.ts"],
  "exclude": ["node_modules", "dist", ".next"]
}
```

---

## Common Commands

```bash
# Build and run
npm run dev          # Start dev server
npm run build        # Build for production
npm start            # Start production server

# Testing
npm test             # Run tests
npm run test:watch   # Watch mode

# Type checking
npx tsc --noEmit     # Check types

# Linting
npm run lint         # Run ESLint
npm run format       # Format code with Prettier

# Database (if using Prisma)
npx prisma generate  # Generate Prisma client
npx prisma migrate   # Run migrations
```

---

## Performance Best Practices

1. **Code splitting**: Use dynamic imports for lazy loading
2. **Image optimization**: Use Next.js Image component
3. **Caching**: Leverage React Query or SWR for data
4. **Type safety**: Enable strict mode in tsconfig
5. **Bundle analysis**: Use webpack-bundle-analyzer
6. **Hydration**: Match server and client output
7. **Suspense**: Use for async components and transitions
8. **Server Components**: Default in Next.js 16 for better performance

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Types not found | Check tsconfig.json paths, run `npm install` |
| Module not found | Verify import path, check barrel exports |
| Type mismatch | Review function signature, check z.infer<> |
| Build error | Clear .next/, node_modules, run build again |
| Hydration mismatch | Ensure server/client output matches |
| Slow builds | Check Turbopack config, analyze bundle |

