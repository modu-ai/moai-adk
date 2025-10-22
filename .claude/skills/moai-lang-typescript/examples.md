# TypeScript 5.7 Code Examples

Production-ready examples for modern TypeScript development with Vitest 2.1.8, Biome 1.9.4, React 19, and TypeScript 5.7 features.

---

## Example 1: Vitest 2.1 with React 19 Component Testing

### Test File: `UserCard.test.tsx`

```typescript
// @TEST:UI-001 | SPEC: SPEC-UI-001.md | CODE: UserCard.tsx
import { render, screen } from '@testing/library'
import { describe, it, expect, vi } from 'vitest'
import userEvent from '@testing-library/user-event'
import { UserCard } from './UserCard'

describe('UserCard', () => {
  const mockUser = {
    id: 1,
    name: 'Alice Johnson',
    email: 'alice@example.com',
    role: 'admin'
  }

  it('renders user information correctly', () => {
    render(<UserCard user={mockUser} />)

    expect(screen.getByText('Alice Johnson')).toBeInTheDocument()
    expect(screen.getByText('alice@example.com')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /edit/i })).toBeInTheDocument()
  })

  it('calls onEdit when edit button is clicked', async () => {
    const handleEdit = vi.fn()
    const user = userEvent.setup()

    render(<UserCard user={mockUser} onEdit={handleEdit} />)

    const editButton = screen.getByRole('button', { name: /edit/i })
    await user.click(editButton)

    expect(handleEdit).toHaveBeenCalledWith(mockUser.id)
    expect(handleEdit).toHaveBeenCalledTimes(1)
  })

  it('displays admin badge for admin users', () => {
    render(<UserCard user={mockUser} />)

    expect(screen.getByText('Admin')).toBeInTheDocument()
  })

  it('does not display admin badge for regular users', () => {
    const regularUser = { ...mockUser, role: 'user' }

    render(<UserCard user={regularUser} />)

    expect(screen.queryByText('Admin')).not.toBeInTheDocument()
  })
})
```

### Component Implementation: `UserCard.tsx`

```typescript
// @CODE:UI-001 | SPEC: SPEC-UI-001.md | TEST: UserCard.test.tsx
import type { FC } from 'react'

interface User {
  id: number
  name: string
  email: string
  role: 'admin' | 'user'
}

interface UserCardProps {
  user: User
  onEdit?: (id: number) => void
}

export const UserCard: FC<UserCardProps> = ({ user, onEdit }) => {
  return (
    <div className="user-card">
      <div className="user-info">
        <h3>{user.name}</h3>
        <p>{user.email}</p>
        {user.role === 'admin' && <span className="badge">Admin</span>}
      </div>
      {onEdit && (
        <button
          type="button"
          onClick={() => onEdit(user.id)}
          aria-label="Edit user"
        >
          Edit
        </button>
      )}
    </div>
  )
}
```

**Key Features**:
- ✅ React Testing Library for user-centric tests
- ✅ Vitest mocking with `vi.fn()`
- ✅ `userEvent` for realistic user interactions
- ✅ Accessibility-focused queries (`getByRole`, `aria-label`)
- ✅ TypeScript strict typing

**Run Commands**:
```bash
vitest                             # Run tests in watch mode
vitest run                         # Run tests once
vitest --coverage                  # Run with coverage
vitest --ui                        # Open Vitest UI
```

---

## Example 2: TDD Workflow with API Service

### RED Phase: Write Failing Test First

```typescript
// services/userService.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { UserService } from './userService'
import type { User, CreateUserDTO } from '../types'

// @TEST:API-001 | SPEC: SPEC-API-001.md | CODE: userService.ts
describe('UserService', () => {
  let userService: UserService

  beforeEach(() => {
    userService = new UserService('https://api.example.com')
  })

  describe('createUser', () => {
    it('creates a user successfully', async () => {
      const newUser: CreateUserDTO = {
        name: 'Bob Smith',
        email: 'bob@example.com'
      }

      const createdUser = await userService.createUser(newUser)

      expect(createdUser).toEqual({
        id: expect.any(Number),
        name: 'Bob Smith',
        email: 'bob@example.com',
        createdAt: expect.any(String)
      })
    })

    it('throws error for invalid email', async () => {
      const invalidUser: CreateUserDTO = {
        name: 'Invalid User',
        email: 'invalid-email'
      }

      await expect(userService.createUser(invalidUser)).rejects.toThrow(
        'Invalid email format'
      )
    })

    it('throws error when API returns 400', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 400,
        json: async () => ({ error: 'Bad Request' })
      })

      const newUser: CreateUserDTO = {
        name: 'Test',
        email: 'test@example.com'
      }

      await expect(userService.createUser(newUser)).rejects.toThrow(
        'Bad Request'
      )
    })
  })
})
```

### GREEN Phase: Implement Minimum Code

```typescript
// services/userService.ts
// @CODE:API-001 | SPEC: SPEC-API-001.md | TEST: userService.test.ts
import type { User, CreateUserDTO } from '../types'

export class UserService {
  constructor(private baseUrl: string) {}

  async createUser(dto: CreateUserDTO): Promise<User> {
    // Validate email
    if (!this.isValidEmail(dto.email)) {
      throw new Error('Invalid email format')
    }

    const response = await fetch(`${this.baseUrl}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(dto)
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to create user')
    }

    return response.json()
  }

  private isValidEmail(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
  }
}
```

### REFACTOR Phase: Extract Validation & Error Handling

```typescript
// utils/validation.ts
export const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function validateEmail(email: string): void {
  if (!emailRegex.test(email)) {
    throw new Error('Invalid email format')
  }
}

// services/apiClient.ts
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public body?: unknown
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export async function handleApiResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new ApiError(
      error.error || 'API request failed',
      response.status,
      error
    )
  }
  return response.json()
}

// services/userService.ts (refactored)
import { validateEmail } from '../utils/validation'
import { handleApiResponse } from './apiClient'

export class UserService {
  constructor(private baseUrl: string) {}

  async createUser(dto: CreateUserDTO): Promise<User> {
    validateEmail(dto.email)

    const response = await fetch(`${this.baseUrl}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(dto)
    })

    return handleApiResponse<User>(response)
  }
}
```

**Refactoring Improvements**:
- ✅ Extracted validation logic to reusable module
- ✅ Created custom `ApiError` class
- ✅ Abstracted API response handling
- ✅ Better separation of concerns

---

## Example 3: TypeScript 5.7 Advanced Features

### Type Predicates & Narrowing

```typescript
// types/guards.ts
interface Cat {
  type: 'cat'
  meow: () => void
}

interface Dog {
  type: 'dog'
  bark: () => void
}

type Animal = Cat | Dog

// Type predicate
export function isCat(animal: Animal): animal is Cat {
  return animal.type === 'cat'
}

// Usage
function makeSound(animal: Animal): void {
  if (isCat(animal)) {
    animal.meow() // TypeScript knows it's Cat
  } else {
    animal.bark() // TypeScript knows it's Dog
  }
}
```

### Template Literal Types

```typescript
// types/routes.ts
type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'
type Route = '/users' | '/posts' | '/comments'

// Combine types to create allowed API endpoints
type APIEndpoint = `${HTTPMethod} ${Route}`

// Usage
const validEndpoint: APIEndpoint = 'GET /users' // ✅
// const invalidEndpoint: APIEndpoint = 'PATCH /users' // ❌ Type error
```

### Satisfies Operator (TypeScript 5.x)

```typescript
// config.ts
type Config = {
  apiUrl: string
  timeout: number
  retries?: number
}

// Enforce type while preserving literal types
const config = {
  apiUrl: 'https://api.example.com',
  timeout: 5000,
  retries: 3
} satisfies Config

// TypeScript infers literal types
const url: 'https://api.example.com' = config.apiUrl // ✅

// Still type-safe
// const invalid = { apiUrl: 123 } satisfies Config // ❌
```

---

## Example 4: Vitest Mocking Patterns

### Mocking Modules

```typescript
// services/logger.test.ts
import { describe, it, expect, vi } from 'vitest'
import { Logger } from './logger'

vi.mock('./logger', () => ({
  Logger: vi.fn().mockImplementation(() => ({
    info: vi.fn(),
    error: vi.fn(),
    warn: vi.fn()
  }))
}))

describe('Logger', () => {
  it('logs info messages', () => {
    const logger = new Logger()
    logger.info('Test message')

    expect(logger.info).toHaveBeenCalledWith('Test message')
  })
})
```

### Spying on Functions

```typescript
// utils/analytics.test.ts
import { describe, it, expect, vi } from 'vitest'
import * as analytics from './analytics'

describe('analytics', () => {
  it('tracks page view', () => {
    const spy = vi.spyOn(analytics, 'trackPageView')

    analytics.trackPageView('/home')

    expect(spy).toHaveBeenCalledWith('/home')

    spy.mockRestore()
  })
})
```

### Fake Timers

```typescript
// utils/delay.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { delay } from './delay'

describe('delay', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('resolves after specified time', async () => {
    const promise = delay(1000)

    vi.advanceTimersByTime(999)
    expect(promise).not.toHaveResolved()

    vi.advanceTimersByTime(1)
    await expect(promise).resolves.toBeUndefined()
  })
})
```

---

## Example 5: React 19 with TypeScript

### Server Components (React 19)

```typescript
// app/users/page.tsx
import type { FC } from 'react'

interface User {
  id: number
  name: string
}

// Server Component (default in Next.js 14+)
const UsersPage: FC = async () => {
  const response = await fetch('https://api.example.com/users', {
    cache: 'no-store' // Opt out of caching
  })
  const users: User[] = await response.json()

  return (
    <div>
      <h1>Users</h1>
      <ul>
        {users.map((user) => (
          <li key={user.id}>{user.name}</li>
        ))}
      </ul>
    </div>
  )
}

export default UsersPage
```

### use() Hook (React 19)

```typescript
// components/UserProfile.tsx
'use client'

import { use, type FC } from 'react'

interface UserProfileProps {
  userPromise: Promise<User>
}

export const UserProfile: FC<UserProfileProps> = ({ userPromise }) => {
  // use() unwraps promises in render
  const user = use(userPromise)

  return (
    <div>
      <h2>{user.name}</h2>
      <p>{user.email}</p>
    </div>
  )
}
```

### Form Actions (React 19)

```typescript
// app/users/create/page.tsx
'use client'

import { useActionState, type FC } from 'react'

async function createUser(prevState: unknown, formData: FormData) {
  const name = formData.get('name') as string
  const email = formData.get('email') as string

  try {
    const response = await fetch('/api/users', {
      method: 'POST',
      body: JSON.stringify({ name, email })
    })

    if (!response.ok) throw new Error('Failed to create user')

    return { success: true, message: 'User created!' }
  } catch (error) {
    return { success: false, message: 'Error creating user' }
  }
}

const CreateUserPage: FC = () => {
  const [state, formAction, isPending] = useActionState(createUser, null)

  return (
    <form action={formAction}>
      <input name="name" required />
      <input name="email" type="email" required />
      <button type="submit" disabled={isPending}>
        {isPending ? 'Creating...' : 'Create User'}
      </button>
      {state?.message && <p>{state.message}</p>}
    </form>
  )
}

export default CreateUserPage
```

---

## Example 6: Biome 1.9 Configuration

### Configuration File (`biome.json`)

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
    "ignore": ["node_modules", "dist", "build", ".next"]
  },
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "suspicious": {
        "noExplicitAny": "error",
        "noDoubleEquals": "error"
      },
      "style": {
        "useConst": "error",
        "useTemplate": "warn"
      },
      "correctness": {
        "noUnusedVariables": "error"
      }
    }
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "trailingCommas": "es5",
      "semicolons": "asNeeded"
    }
  },
  "typescript": {
    "enabled": true
  }
}
```

**Run Commands**:
```bash
biome check .                      # Check all files
biome check --write .              # Check and auto-fix
biome format .                     # Format only
biome lint .                       # Lint only
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/typescript.yml
name: TypeScript CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '22'

      - name: Install dependencies
        run: npm ci

      - name: Run Biome
        run: npx biome check .

      - name: Type check
        run: npx tsc --noEmit

      - name: Run tests
        run: npx vitest run --coverage

      - name: Check coverage
        run: |
          coverage=$(jq -r '.total.lines.pct' coverage/coverage-summary.json)
          if (( $(echo "$coverage < 85" | bc -l) )); then
            echo "Coverage $coverage% is below 85%"
            exit 1
          fi

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

# View coverage report
open coverage/index.html

# Check coverage threshold
vitest run --coverage --coverage.lines=85
```

### R - Readable

```bash
# Format code with Biome
biome format --write .

# Lint code
biome lint .
```

### U - Unified (Type Safety)

```typescript
// Enable strict mode in tsconfig.json
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

# Update dependencies
npm update

# Check for outdated packages
npm outdated
```

### T - Trackable

```typescript
// @TAG integration in code
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: authService.test.ts
export class AuthService {
  async login(credentials: Credentials): Promise<User> {
    // Implementation
  }
}
```

---

**Version**: 1.0.0 (2025-10-22)
**Updated**: Latest tool versions verified 2025-10-22
**Framework**: MoAI-ADK TypeScript Language Skill
**Status**: Production-ready
