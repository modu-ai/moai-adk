# Advanced shadcn/ui Component Patterns

## Architecture Patterns

### Complex Component Composition

shadcn/ui excels at composing complex UI from simple, reusable components. The architecture follows a composition-over-inheritance approach.

Pattern: Compound Components

```typescript
// Compound component pattern with shadcn/ui
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import React, { createContext, useContext } from "react"

// Create a context for form state
interface FormContextType {
 data: Record<string, any>
 setData: (key: string, value: any) => void
 errors: Record<string, string>
}

const FormContext = createContext<FormContextType | undefined>(undefined)

export function Form({ children, onSubmit }: { children: React.ReactNode; onSubmit: (data: any) => void }) {
 const [data, setData] = React.useState({})
 const [errors, setErrors] = React.useState({})

 return (
 <FormContext.Provider value={{ data, setData: (k, v) => setData(prev => ({ ...prev, [k]: v })), errors }}>
 <form
 onSubmit={(e) => {
 e.preventDefault()
 onSubmit(data)
 }}
 >
 {children}
 </form>
 </FormContext.Provider>
 )
}

export function FormField({ label, name, type = "text" }: { label: string; name: string; type?: string }) {
 const context = useContext(FormContext)
 if (!context) throw new Error("FormField must be used inside Form")

 return (
 <div className="mb-4">
 <label htmlFor={name} className="block text-sm font-medium mb-1">
 {label}
 </label>
 <Input
 id={name}
 type={type}
 value={context.data[name] || ""}
 onChange={(e) => context.setData(name, e.target.value)}
 />
 </div>
 )
}

// Usage
export function MyForm() {
 return (
 <Form onSubmit={(data) => console.log(data)}>
 <Card>
 <CardHeader>
 <CardTitle>Contact Form</CardTitle>
 <CardDescription>Enter your details</CardDescription>
 </CardHeader>
 <CardContent>
 <FormField label="Name" name="name" />
 <FormField label="Email" name="email" type="email" />
 <Button type="submit">Submit</Button>
 </CardContent>
 </Card>
 </Form>
 )
}
```

### Design Token Integration

shadcn/ui uses CSS variables for theming. Components automatically adapt to theme changes.

Pattern: Theme-Aware Components

```typescript
// Using design tokens in custom components
export function StatusBadge({ status }: { status: "success" | "error" | "warning" }) {
 const statusConfig = {
 success: "bg-green-500/20 text-green-700 dark:text-green-400",
 error: "bg-red-500/20 text-red-700 dark:text-red-400",
 warning: "bg-yellow-500/20 text-yellow-700 dark:text-yellow-400",
 }

 return (
 <span className={`px-3 py-1 rounded-md text-sm font-medium ${statusConfig[status]}`}>
 {status.charAt(0).toUpperCase() + status.slice(1)}
 </span>
 )
}

// Theme switching with context
import { useEffect, useState } from "react"

export function ThemeSwitcher() {
 const [theme, setTheme] = useState("light")

 useEffect(() => {
 const html = document.documentElement
 if (theme === "dark") {
 html.classList.add("dark")
 } else {
 html.classList.remove("dark")
 }
 }, [theme])

 return (
 <Button
 variant="outline"
 onClick={() => setTheme(theme === "light" ? "dark" : "light")}
 >
 {theme === "light" ? "" : ""}
 </Button>
 )
}
```

## Advanced Component Patterns

### Form Composition with Validation

shadcn/ui Button, Input, and Form components integrate well with validation libraries like react-hook-form.

```typescript
import { useForm } from "react-hook-form"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

interface FormData {
 username: string
 email: string
}

export function AdvancedForm() {
 const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
 defaultValues: { username: "", email: "" }
 })

 const onSubmit = (data: FormData) => {
 console.log("Form submitted:", data)
 }

 return (
 <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
 <div>
 <Input
 {...register("username", { required: "Username is required" })}
 placeholder="Enter username"
 />
 {errors.username && <p className="text-red-500 text-sm">{errors.username.message}</p>}
 </div>

 <div>
 <Input
 {...register("email", { required: "Email is required", pattern: { value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i, message: "Invalid email" } })}
 placeholder="Enter email"
 type="email"
 />
 {errors.email && <p className="text-red-500 text-sm">{errors.email.message}</p>}
 </div>

 <Button type="submit">Submit</Button>
 </form>
 )
}
```

### Complex Data Table with Sorting and Filtering

```typescript
import { DataTable } from "@/components/ui/data-table"
import { Button } from "@/components/ui/button"
import { useReactTable, getCoreRowModel, getSortedRowModel, getFilteredRowModel } from "@tanstack/react-table"

export function UserDataTable({ users }: { users: User[] }) {
 const [sorting, setSorting] = React.useState([])
 const [columnFilters, setColumnFilters] = React.useState([])

 const table = useReactTable({
 data: users,
 columns: [
 {
 accessorKey: "name",
 header: "Name",
 },
 {
 accessorKey: "email",
 header: "Email",
 },
 ],
 getCoreRowModel: getCoreRowModel(),
 getSortedRowModel: getSortedRowModel(),
 getFilteredRowModel: getFilteredRowModel(),
 state: {
 sorting,
 columnFilters,
 },
 onSortingChange: setSorting,
 onColumnFiltersChange: setColumnFilters,
 })

 return (
 <div>
 <Button onClick={() => setColumnFilters([{ id: "name", value: "john" }])}>
 Filter by John
 </Button>
 <DataTable table={table} />
 </div>
 )
}
```

## Component Composition Strategies

### Building Component Trees

shadcn/ui components can be nested and composed to create complex layouts.

Strategy: Hierarchical Component Structure

```typescript
// Parent wrapper component
export function DashboardLayout({ children }: { children: React.ReactNode }) {
 return (
 <div className="grid grid-cols-4 gap-4 p-6">
 <aside className="col-span-1">
 <Navigation />
 </aside>
 <main className="col-span-3">
 {children}
 </main>
 </div>
 )
}

// Navigation subcomponent
function Navigation() {
 return (
 <nav className="space-y-2">
 <Button variant="ghost" className="w-full justify-start">Dashboard</Button>
 <Button variant="ghost" className="w-full justify-start">Settings</Button>
 <Button variant="ghost" className="w-full justify-start">Profile</Button>
 </nav>
 )
}

// Usage
export function Dashboard() {
 return (
 <DashboardLayout>
 <Card>
 <CardHeader>
 <CardTitle>Welcome</CardTitle>
 </CardHeader>
 <CardContent>
 Dashboard content here
 </CardContent>
 </Card>
 </DashboardLayout>
 )
}
```

## Design System Scaling

### Managing Component Variants

shadcn/ui components support variants for different use cases.

Pattern: Variant Management

```typescript
// Using Button variants for consistent interface
export function ActionButtons() {
 return (
 <div className="flex gap-2">
 <Button variant="default">Primary Action</Button>
 <Button variant="secondary">Secondary Action</Button>
 <Button variant="outline">Outline Action</Button>
 <Button variant="destructive">Delete</Button>
 <Button variant="ghost">Ghost Action</Button>
 </div>
 )
}

// Creating custom size variants
export function SizedButtons() {
 return (
 <div className="flex gap-2 items-center">
 <Button size="sm">Small</Button>
 <Button size="default">Default</Button>
 <Button size="lg">Large</Button>
 </div>
 )
}
```

## Responsive Design Patterns

### Mobile-First Approach with Tailwind

shadcn/ui uses Tailwind CSS which supports responsive breakpoints.

```typescript
export function ResponsiveCard() {
 return (
 <Card className="w-full sm:max-w-sm md:max-w-md lg:max-w-lg">
 <CardHeader className="p-4 sm:p-6">
 <CardTitle className="text-lg sm:text-xl md:text-2xl">
 Responsive Card
 </CardTitle>
 </CardHeader>
 <CardContent className="p-4 sm:p-6">
 <p className="text-sm sm:text-base md:text-lg">
 Content adapts to screen size
 </p>
 </CardContent>
 </Card>
 )
}
```

## Performance Patterns

### Memoization and Lazy Loading

shadcn/ui components benefit from React.memo for preventing unnecessary re-renders.

```typescript
import React from "react"
import { Card, CardContent } from "@/components/ui/card"

interface UserCardProps {
 user: User
}

export const UserCard = React.memo(function UserCard({ user }: UserCardProps) {
 return (
 <Card>
 <CardContent>
 <p>{user.name}</p>
 <p className="text-gray-500">{user.email}</p>
 </CardContent>
 </Card>
 )
})
```

## Accessibility Standards

### ARIA Integration

shadcn/ui components come with built-in accessibility features.

```typescript
export function AccessibleDialog() {
 const [open, setOpen] = React.useState(false)

 return (
 <>
 <Button onClick={() => setOpen(true)} aria-label="Open dialog">
 Open
 </Button>
 {open && (
 <div role="dialog" aria-labelledby="dialog-title" aria-modal="true">
 <h2 id="dialog-title">Dialog Title</h2>
 <p>Dialog content here</p>
 <Button onClick={() => setOpen(false)}>Close</Button>
 </div>
 )}
 </>
 )
}
```

---

Version: 4.0.0
Last Updated: 2025-11-22
Status: Production Ready
