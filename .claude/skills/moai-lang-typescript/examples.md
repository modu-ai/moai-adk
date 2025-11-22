# TypeScript - Practical Examples (10 Examples)

## Example 1: Interfaces
```typescript
interface User {
    id: number;
    name: string;
    email?: string;
}

function getUser(id: number): User {
    return { id, name: "John" };
}
```

## Example 2: Generics
```typescript
function identity<T>(arg: T): T {
    return arg;
}

const result = identity<string>("hello");
```

## Example 3: Union Types
```typescript
type Result = string | number;

function process(value: Result): void {
    if (typeof value === "string") {
        console.log(value.toUpperCase());
    } else {
        console.log(value * 2);
    }
}
```

## Example 4: Type Guards
```typescript
function isUser(obj: any): obj is User {
    return 'id' in obj && 'name' in obj;
}
```

## Example 5: Utility Types
```typescript
type Partial<T> = { [P in keyof T]?: T[P] };
type Required<T> = { [P in keyof T]-?: T[P] };
type Pick<T, K extends keyof T> = { [P in K]: T[P] };
```

## Example 6: Async/Await
```typescript
async function fetchUser(id: number): Promise<User> {
    const res = await fetch(`/api/users/${id}`);
    return res.json();
}
```

## Example 7: Classes
```typescript
class UserService {
    constructor(private readonly repo: UserRepository) {}
    
    async getUser(id: number): Promise<User> {
        return this.repo.findById(id);
    }
}
```

## Example 8: Enums
```typescript
enum Status {
    Active = "ACTIVE",
    Inactive = "INACTIVE"
}
```

## Example 9: Decorators
```typescript
function Log(target: any, key: string) {
    console.log(`${key} was called`);
}

class Example {
    @Log
    method() {}
}
```

## Example 10: Type Inference
```typescript
const user = { id: 1, name: "John" }; // inferred as { id: number, name: string }
```
