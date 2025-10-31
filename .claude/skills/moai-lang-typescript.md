# moai-lang-typescript

TypeScript 5.7+ best practices with strict typing, generics, advanced patterns, and type safety.

## Quick Start

TypeScript adds static type checking to JavaScript, catching errors before runtime and improving code maintainability. Use this skill when building applications that require type safety, working with complex data structures, or creating reusable libraries.

## Core Patterns

### Pattern 1: Strict Typing & Type Guards

**Pattern**: Enable strict mode and use type guards to ensure runtime safety.

```typescript
// tsconfig.json - Strict mode configuration
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "strict": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "strictBindCallApply": true,
    "strictPropertyInitialization": true,
    "noImplicitAny": true,
    "noImplicitThis": true,
    "alwaysStrict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
  }
}

// Type guards for runtime validation
type Admin = { role: 'admin'; permissions: string[] };
type User = { role: 'user'; id: string };
type Guest = { role: 'guest' };

type AccountType = Admin | User | Guest;

// Type guard function
function isAdmin(account: AccountType): account is Admin {
  return account.role === 'admin';
}

function isUser(account: AccountType): account is User {
  return account.role === 'user';
}

// Using type guards
function processAccount(account: AccountType) {
  if (isAdmin(account)) {
    // Type narrowed to Admin
    console.log('Admin permissions:', account.permissions);
  } else if (isUser(account)) {
    // Type narrowed to User
    console.log('User ID:', account.id);
  } else {
    // Type narrowed to Guest
    console.log('Guest access');
  }
}

// Discriminated unions
type Result<T> =
  | { success: true; data: T }
  | { success: false; error: string };

function handleResult<T>(result: Result<T>) {
  if (result.success) {
    // Type narrowed - data is available
    console.log(result.data);
  } else {
    // Type narrowed - error is available
    console.error(result.error);
  }
}
```

**When to use**:
- Building applications where runtime errors are expensive
- Working in teams where type safety improves collaboration
- Creating APIs that should have clear contracts
- Refactoring legacy code with confidence

**Key benefits**:
- Catch errors at compile time, not runtime
- Self-documenting code through types
- Improved IDE autocomplete and refactoring
- Gradual adoption of TypeScript possible

### Pattern 2: Generics & Advanced Type Patterns

**Pattern**: Use generics to create reusable, type-safe components and functions.

```typescript
// Generic function with constraints
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

const person = { name: 'Alice', age: 30 };
const name = getProperty(person, 'name'); // Type: string
const age = getProperty(person, 'age'); // Type: number

// Generic component with React
import { ReactNode } from 'react';

interface Props<T> {
  items: T[];
  renderItem: (item: T) => ReactNode;
  keyExtractor: (item: T) => string | number;
}

function List<T>({ items, renderItem, keyExtractor }: Props<T>) {
  return (
    <ul>
      {items.map((item) => (
        <li key={keyExtractor(item)}>{renderItem(item)}</li>
      ))}
    </ul>
  );
}

// Usage with type inference
<List
  items={[{ id: 1, name: 'Alice' }]}
  renderItem={(user) => <span>{user.name}</span>}
  keyExtractor={(user) => user.id}
/>

// Utility types for type transformation
type User = {
  id: string;
  name: string;
  email: string;
  password: string;
};

// Make all fields optional
type UserPreferences = Partial<User>;

// Make password and email required, others optional
type UserUpdate = Omit<User, 'id' | 'password'> & { password?: string };

// Extract specific fields
type UserPublic = Pick<User, 'id' | 'name' | 'email'>;

// Create readonly version
type ReadonlyUser = Readonly<User>;

// Create a mapping from User fields to strings
type UserErrors = Record<keyof User, string>;

const errors: UserErrors = {
  id: 'Invalid ID',
  name: 'Name is required',
  email: 'Invalid email',
  password: 'Password too weak',
};
```

**When to use**:
- Building component libraries with flexible APIs
- Creating type-safe database queries
- Working with complex data transformations
- Building frameworks or abstractions

**Key benefits**:
- Write once, use everywhere with generics
- Type safety across function signatures
- Reduced code duplication
- Better for large codebases

### Pattern 3: Advanced Patterns: Const Assertions & Type Inference

**Pattern**: Use const assertions and inference for better type precision.

```typescript
// Const assertion for literal types
const status = 'pending' as const; // Type: 'pending', not string

// Useful for object literals
const routes = {
  home: '/',
  about: '/about',
  contact: '/contact',
} as const; // Type: { readonly home: '/'; readonly about: '/about'; ... }

type RouteName = keyof typeof routes; // 'home' | 'about' | 'contact'

// Function to get values with strong typing
function navigate(route: RouteName) {
  const path = routes[route];
  // path has type: '/' | '/about' | '/contact'
}

// Template literal types for dynamic strings
type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';
type Route = `/${string}`;

type ApiRoute = `${HttpMethod} ${Route}`;

const getUsers: ApiRoute = 'GET /users'; // Valid
const createUser: ApiRoute = 'POST /users'; // Valid
// const invalid: ApiRoute = 'INVALID /users'; // Error!

// Conditional types for powerful abstractions
type Flatten<Type> = Type extends Array<infer Item> ? Item : Type;

type Str = Flatten<string[]>; // string
type Num = Flatten<number>; // number

// Inferring types from return values
function createUser(name: string) {
  return {
    id: Math.random(),
    name,
    createdAt: new Date(),
  };
}

type NewUser = ReturnType<typeof createUser>;
// Type: { id: number; name: string; createdAt: Date }
```

**When to use**:
- Working with configuration objects
- Creating type-safe string unions
- Building flexible, reusable components
- Extracting types from existing functions

**Key benefits**:
- More precise types (literal instead of general)
- Better refactoring - change value, type updates automatically
- Type inference reduces boilerplate
- Self-documenting through type definitions

## Progressive Disclosure

### Level 1: Basic Types
- Primitive types (string, number, boolean)
- Interfaces and type aliases
- Union and intersection types
- Optional and readonly properties

### Level 2: Advanced Types
- Generics with constraints
- Type guards and discriminated unions
- Mapped and conditional types
- Type inference with infer keyword

### Level 3: Expert Patterns
- Advanced utility types
- Recursive types
- Template literal types
- Type-level programming

## Works Well With

- **Next.js 16**: Full TypeScript support with Server Components
- **React 19**: Type-safe component props and hooks
- **Biome**: Fast linting and formatting with TypeScript
- **Vitest**: Unit testing with full type safety
- **Zod**: Runtime validation with inferred TypeScript types
- **tRPC**: End-to-end type safety for APIs

## References

- **Official Documentation**: https://www.typescriptlang.org/
- **Handbook**: https://www.typescriptlang.org/docs/handbook/
- **Advanced Types**: https://www.typescriptlang.org/docs/handbook/2/types-from-types.html
- **Utility Types**: https://www.typescriptlang.org/docs/handbook/utility-types.html
- **TypeScript in 5 Minutes**: https://www.typescriptlang.org/docs/handbook/typescript-in-5-minutes.html
