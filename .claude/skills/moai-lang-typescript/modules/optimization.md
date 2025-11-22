# TypeScript Performance

## Strict Mode
```json
{
    "compilerOptions": {
        "strict": true,
        "noImplicitAny": true,
        "strictNullChecks": true
    }
}
```

## Type Narrowing
```typescript
function process(value: string | number) {
    if (typeof value === "string") {
        return value.toUpperCase();
    }
    return value * 2;
}
```

## Const Assertions
```typescript
const colors = ["red", "green"] as const;
type Color = typeof colors[number]; // "red" | "green"
```
