# Advanced Component Design Patterns

## Component Architecture

### Atomic Design Methodology

Atomic design breaks down UI into hierarchical components:

```typescript
// Atoms: Basic building blocks
export function Label({ htmlFor, children }: { htmlFor: string; children: React.ReactNode }) {
  return <label htmlFor={htmlFor} className="text-sm font-medium">{children}</label>
}

export function Input({ id, ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input id={id} className="border rounded px-2 py-1" {...props} />
}

// Molecules: Combinations of atoms
export function FormField({
  label,
  id,
  ...inputProps
}: {
  label: string
  id: string
} & React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <div className="mb-4">
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} {...inputProps} />
    </div>
  )
}

// Organisms: Complex compositions
export function LoginForm() {
  return (
    <form className="space-y-4">
      <FormField label="Email" id="email" type="email" placeholder="Enter email" />
      <FormField label="Password" id="password" type="password" placeholder="Enter password" />
      <button className="bg-blue-500 text-white px-4 py-2 rounded">Sign In</button>
    </form>
  )
}
```

### Compound Component Pattern

```typescript
// Parent component with children
export function Tabs({ children }: { children: React.ReactNode }) {
  const [activeTab, setActiveTab] = React.useState(0)

  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      {children}
    </TabsContext.Provider>
  )
}

export function TabsList({ children }: { children: React.ReactNode }) {
  return <div className="flex border-b">{children}</div>
}

export function TabsTrigger({ value, children }: { value: number; children: React.ReactNode }) {
  const { activeTab, setActiveTab } = React.useContext(TabsContext)
  return (
    <button
      onClick={() => setActiveTab(value)}
      className={activeTab === value ? "border-b-2 border-blue-500" : ""}
    >
      {children}
    </button>
  )
}

export function TabsContent({ value, children }: { value: number; children: React.ReactNode }) {
  const { activeTab } = React.useContext(TabsContext)
  return activeTab === value ? <div>{children}</div> : null
}

// Usage
<Tabs>
  <TabsList>
    <TabsTrigger value={0}>Tab 1</TabsTrigger>
    <TabsTrigger value={1}>Tab 2</TabsTrigger>
  </TabsList>
  <TabsContent value={0}>Content 1</TabsContent>
  <TabsContent value={1}>Content 2</TabsContent>
</Tabs>
```

## Reusable Component Patterns

### Wrapper Component Pattern

```typescript
interface WrapperProps {
  children: React.ReactNode
  className?: string
  variant?: "default" | "elevated" | "outlined"
}

export function ComponentWrapper({ children, className = "", variant = "default" }: WrapperProps) {
  const variantClasses = {
    default: "p-4 rounded",
    elevated: "p-4 rounded shadow-lg",
    outlined: "p-4 rounded border",
  }

  return (
    <div className={`${variantClasses[variant]} ${className}`}>
      {children}
    </div>
  )
}
```

### Render Props Pattern

```typescript
interface ToggleProps {
  children: (state: { isOpen: boolean; toggle: () => void }) => React.ReactNode
}

export function Toggle({ children }: ToggleProps) {
  const [isOpen, setIsOpen] = React.useState(false)

  return children({
    isOpen,
    toggle: () => setIsOpen(!isOpen),
  })
}

// Usage
<Toggle>
  {({ isOpen, toggle }) => (
    <>
      <button onClick={toggle}>Toggle</button>
      {isOpen && <p>Content is visible</p>}
    </>
  )}
</Toggle>
```

### Higher-Order Component (HOC)

```typescript
function withTheme<P extends object>(Component: React.ComponentType<P & { theme: string }>) {
  return function ThemedComponent(props: P) {
    const [theme, setTheme] = React.useState("light")

    return <Component {...props} theme={theme} />
  }
}

// Usage
const ThemedButton = withTheme(Button)
```

## Component Composition Strategies

### Composition with Slots

```typescript
interface CardProps {
  header?: React.ReactNode
  footer?: React.ReactNode
  children: React.ReactNode
}

export function Card({ header, footer, children }: CardProps) {
  return (
    <div className="border rounded p-4">
      {header && <div className="mb-4 pb-4 border-b">{header}</div>}
      <div>{children}</div>
      {footer && <div className="mt-4 pt-4 border-t">{footer}</div>}
    </div>
  )
}

// Usage
<Card
  header={<h2>Title</h2>}
  footer={<button>Close</button>}
>
  Content
</Card>
```

## Component API Design

### Flexible Props API

```typescript
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "danger"
  size?: "sm" | "md" | "lg"
  isLoading?: boolean
  leftIcon?: React.ReactNode
  rightIcon?: React.ReactNode
}

export function Button({
  variant = "primary",
  size = "md",
  isLoading,
  leftIcon,
  rightIcon,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      className={`${getVariantClass(variant)} ${getSizeClass(size)}`}
      disabled={isLoading || props.disabled}
    >
      {isLoading && <Spinner />}
      {leftIcon}
      {children}
      {rightIcon}
    </button>
  )
}
```

## Testing Patterns

### Component Testing Strategy

```typescript
import { render, screen } from "@testing-library/react"
import { Button } from "./Button"

describe("Button", () => {
  it("renders with label", () => {
    render(<Button>Click me</Button>)
    expect(screen.getByText("Click me")).toBeInTheDocument()
  })

  it("handles click events", async () => {
    const handleClick = vi.fn()
    render(<Button onClick={handleClick}>Click me</Button>)

    await userEvent.click(screen.getByText("Click me"))
    expect(handleClick).toHaveBeenCalledOnce()
  })

  it("supports different variants", () => {
    const { container } = render(<Button variant="danger">Delete</Button>)
    expect(container.querySelector(".btn-danger")).toBeInTheDocument()
  })
})
```

## Accessibility Standards

### ARIA Implementation

```typescript
interface DialogProps {
  isOpen: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
}

export function Dialog({ isOpen, onClose, title, children }: DialogProps) {
  if (!isOpen) return null

  return (
    <div
      role="dialog"
      aria-labelledby="dialog-title"
      aria-modal="true"
      className="fixed inset-0 flex items-center justify-center"
    >
      <div className="bg-white rounded shadow-lg">
        <h2 id="dialog-title" className="text-lg font-bold p-4 border-b">
          {title}
        </h2>
        <div className="p-4">{children}</div>
        <button onClick={onClose} aria-label="Close dialog">
          Close
        </button>
      </div>
    </div>
  )
}
```

### Semantic HTML

```typescript
export function Article() {
  return (
    <article>
      <header>
        <h1>Article Title</h1>
        <time dateTime="2025-11-22">November 22, 2025</time>
      </header>
      <section>
        <p>Article content...</p>
      </section>
      <footer>
        <p>Article metadata</p>
      </footer>
    </article>
  )
}
```

## State Management Patterns

### Local State with useReducer

```typescript
interface TodoState {
  todos: Array<{ id: string; text: string; done: boolean }>
}

type TodoAction =
  | { type: "ADD"; payload: string }
  | { type: "TOGGLE"; payload: string }
  | { type: "REMOVE"; payload: string }

export function TodoList() {
  const [state, dispatch] = React.useReducer(
    (state: TodoState, action: TodoAction) => {
      switch (action.type) {
        case "ADD":
          return {
            todos: [...state.todos, { id: Date.now().toString(), text: action.payload, done: false }],
          }
        case "TOGGLE":
          return {
            todos: state.todos.map((t) =>
              t.id === action.payload ? { ...t, done: !t.done } : t
            ),
          }
        case "REMOVE":
          return {
            todos: state.todos.filter((t) => t.id !== action.payload),
          }
        default:
          return state
      }
    },
    { todos: [] }
  )

  return (
    <div>
      {state.todos.map((todo) => (
        <div key={todo.id}>
          <input
            type="checkbox"
            checked={todo.done}
            onChange={() => dispatch({ type: "TOGGLE", payload: todo.id })}
          />
          {todo.text}
        </div>
      ))}
    </div>
  )
}
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
