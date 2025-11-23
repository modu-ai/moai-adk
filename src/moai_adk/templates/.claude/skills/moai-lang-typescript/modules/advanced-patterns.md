# TypeScript Advanced Patterns

## Mapped Types
```typescript
type Readonly<T> = {
    readonly [P in keyof T]: T[P];
};

type Optional<T> = {
    [P in keyof T]?: T[P];
};
```

## Conditional Types
```typescript
type NonNullable<T> = T extends null | undefined ? never : T;
```

## Template Literal Types
```typescript
type EventName = "click" | "focus";
type Handler = `on${Capitalize<EventName>}`; // "onClick" | "onFocus"
```

## Discriminated Unions
```typescript
type Shape =
    | { kind: "circle"; radius: number }
    | { kind: "square"; size: number };
```
