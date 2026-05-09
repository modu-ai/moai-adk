# UI/UX Implementation Examples

Practical code examples demonstrating core UI/UX patterns across React 19, Vue 3.5, design tokens, and accessibility.

---

## 1. Button Component (React 19)

### Basic Implementation

```typescript
import { forwardRef } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

const buttonVariants = cva(
 // Base styles
 'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
 {
 variants: {
 variant: {
 default: 'bg-primary text-primary-foreground hover:bg-primary/90',
 destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
 outline: 'border border-input bg-background hover:bg-accent',
 ghost: 'hover:bg-accent hover:text-accent-foreground',
 link: 'text-primary underline-offset-4 hover:underline',
 },
 size: {
 default: 'h-10 px-4 py-2',
 sm: 'h-9 rounded-md px-3',
 lg: 'h-11 rounded-md px-8',
 icon: 'h-10 w-10',
 },
 },
 defaultVariants: {
 variant: 'default',
 size: 'default',
 },
 }
)

export interface ButtonProps
 extends React.ButtonHTMLAttributes<HTMLButtonElement>,
 VariantProps<typeof buttonVariants> {
 asChild?: boolean
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
 ({ className, variant, size, ...props }, ref) => {
 return (
 <button
 className={buttonVariants({ variant, size, className })}
 ref={ref}
 {...props}
 />
 )
 }
)
Button.displayName = 'Button'

export { Button, buttonVariants }
```

### Usage Examples

```tsx
import { Button } from '@/components/ui/button'

<Button>Default Button</Button>
<Button variant="destructive">Delete</Button>
<Button variant="outline" size="sm">Small Outline</Button>
<Button variant="ghost" size="icon">
 <SearchIcon className="h-4 w-4" />
</Button>
```

---

## 2. Form with Validation (React Hook Form + Zod)

```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import {
 Form,
 FormControl,
 FormDescription,
 FormField,
 FormItem,
 FormLabel,
 FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'

// Schema definition
const formSchema = z.object({
 username: z.string().min(2, {
 message: 'Username must be at least 2 characters.',
 }),
 email: z.string().email({
 message: 'Please enter a valid email adddess.',
 }),
 password: z.string().min(8, {
 message: 'Password must be at least 8 characters.',
 }),
})

type FormValues = z.infer<typeof formSchema>

export function SignupForm() {
 const form = useForm<FormValues>({
 resolver: zodResolver(formSchema),
 defaultValues: {
 username: '',
 email: '',
 password: '',
 },
 })

 function onSubmit(values: FormValues) {
 console.log(values)
 }

 return (
 <Form {...form}>
 <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
 <FormField
 control={form.control}
 name="username"
 render={({ field }) => (
 <FormItem>
 <FormLabel>Username</FormLabel>
 <FormControl>
 <Input placeholder="johndoe" {...field} />
 </FormControl>
 <FormDescription>
 This is your public display name.
 </FormDescription>
 <FormMessage />
 </FormItem>
 )}
 />

 <FormField
 control={form.control}
 name="email"
 render={({ field }) => (
 <FormItem>
 <FormLabel>Email</FormLabel>
 <FormControl>
 <Input type="email" placeholder="john@example.com" {...field} />
 </FormControl>
 <FormMessage />
 </FormItem>
 )}
 />

 <FormField
 control={form.control}
 name="password"
 render={({ field }) => (
 <FormItem>
 <FormLabel>Password</FormLabel>
 <FormControl>
 <Input type="password" placeholder="••••••••" {...field} />
 </FormControl>
 <FormMessage />
 </FormItem>
 )}
 />

 <Button type="submit">Submit</Button>
 </form>
 </Form>
 )
}
```

---

## 3. Data Table (TanStack Table)

```typescript
import { useState } from 'react'
import {
 useReactTable,
 getCoreRowModel,
 getSortedRowModel,
 getFilteredRowModel,
 getPaginationRowModel,
 flexRender,
 type ColumnDef,
 type SortingState,
 type ColumnFiltersState,
} from '@tanstack/react-table'

// Define data type
type Payment = {
 id: string
 amount: number
 status: 'pending' | 'processing' | 'success' | 'failed'
 email: string
}

// Define columns
export const columns: ColumnDef<Payment>[] = [
 {
 accessorKey: 'status',
 header: 'Status',
 },
 {
 accessorKey: 'email',
 header: 'Email',
 },
 {
 accessorKey: 'amount',
 header: () => <div className="text-right">Amount</div>,
 cell: ({ row }) => {
 const amount = parseFloat(row.getValue('amount'))
 const formatted = new Intl.NumberFormat('en-US', {
 style: 'currency',
 currency: 'USD',
 }).format(amount)
 return <div className="text-right font-medium">{formatted}</div>
 },
 },
]

export function DataTableDemo() {
 const [data] = useState<Payment[]>([
 {
 id: '728ed52f',
 amount: 100,
 status: 'pending',
 email: 'm@example.com',
 },
 // ... more data
 ])

 const [sorting, setSorting] = useState<SortingState>([])
 const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

 const table = useReactTable({
 data,
 columns,
 getCoreRowModel: getCoreRowModel(),
 getSortedRowModel: getSortedRowModel(),
 getFilteredRowModel: getFilteredRowModel(),
 getPaginationRowModel: getPaginationRowModel(),
 onSortingChange: setSorting,
 onColumnFiltersChange: setColumnFilters,
 state: {
 sorting,
 columnFilters,
 },
 })

 return (
 <div className="rounded-md border">
 <table className="min-w-full divide-y divide-gray-200">
 <thead>
 {table.getHeaderGroups().map((headerGroup) => (
 <tr key={headerGroup.id}>
 {headerGroup.headers.map((header) => (
 <th key={header.id} className="px-6 py-3 text-left">
 {header.isPlaceholder
 ? null
 : flexRender(
 header.column.columnDef.header,
 header.getContext()
 )}
 </th>
 ))}
 </tr>
 ))}
 </thead>
 <tbody>
 {table.getRowModel().rows.map((row) => (
 <tr key={row.id}>
 {row.getVisibleCells().map((cell) => (
 <td key={cell.id} className="px-6 py-4">
 {flexRender(cell.column.columnDef.cell, cell.getContext())}
 </td>
 ))}
 </tr>
 ))}
 </tbody>
 </table>
 </div>
 )
}
```

---

## 4. Theme Provider (Light/Dark Mode)

```typescript
import { createContext, useContext, useEffect, useState } from 'react'

type Theme = 'dark' | 'light' | 'system'

type ThemeProviderProps = {
 children: React.ReactNode
 defaultTheme?: Theme
 storageKey?: string
}

type ThemeProviderState = {
 theme: Theme
 setTheme: (theme: Theme) => void
}

const ThemeProviderContext = createContext<ThemeProviderState | undefined>(
 undefined
)

export function ThemeProvider({
 children,
 defaultTheme = 'system',
 storageKey = 'ui-theme',
 ...props
}: ThemeProviderProps) {
 const [theme, setTheme] = useState<Theme>(
 () => (localStorage.getItem(storageKey) as Theme) || defaultTheme
 )

 useEffect(() => {
 const root = window.document.documentElement
 root.classList.remove('light', 'dark')

 if (theme === 'system') {
 const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
 .matches
 ? 'dark'
 : 'light'
 root.classList.add(systemTheme)
 return
 }

 root.classList.add(theme)
 }, [theme])

 const value = {
 theme,
 setTheme: (theme: Theme) => {
 localStorage.setItem(storageKey, theme)
 setTheme(theme)
 },
 }

 return (
 <ThemeProviderContext.Provider {...props} value={value}>
 {children}
 </ThemeProviderContext.Provider>
 )
}

export const useTheme = () => {
 const context = useContext(ThemeProviderContext)
 if (context === undefined)
 throw new Error('useTheme must be used within a ThemeProvider')
 return context
}

// Usage
function ThemeToggle() {
 const { setTheme } = useTheme()

 return (
 <div>
 <button onClick={() => setTheme('light')}>Light</button>
 <button onClick={() => setTheme('dark')}>Dark</button>
 <button onClick={() => setTheme('system')}>System</button>
 </div>
 )
}
```

---

## 5. Icon Usage Patterns

```typescript
import { Heart, Search, Settings } from 'lucide-react'
import { FC, SVGProps } from 'react'

// Type-safe icon button
type IconType = FC<SVGProps<SVGSVGElement>>

interface IconButtonProps {
 icon: IconType
 label: string
 onClick?: () => void
}

function IconButton({ icon: Icon, label, onClick }: IconButtonProps) {
 return (
 <button
 onClick={onClick}
 aria-label={label}
 className="p-2 rounded-lg hover:bg-gray-100"
 >
 <Icon className="w-5 h-5" />
 </button>
 )
}

// Usage examples
export function IconExamples() {
 return (
 <div className="flex gap-2">
 {/* Basic icons */}
 <Heart size={24} color="#ef4444" />
 <Search className="w-6 h-6 text-gray-500" />
 <Settings size={20} />

 {/* Icon buttons */}
 <IconButton icon={Heart} label="Like" onClick={() => console.log('Liked')} />
 <IconButton icon={Search} label="Search" />
 <IconButton icon={Settings} label="Settings" />
 </div>
 )
}
```

---

## 6. Accessible Modal Dialog

```typescript
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

export function AccessibleDialog() {
 return (
 <Dialog>
 <DialogTrigger asChild>
 <Button>Open Dialog</Button>
 </DialogTrigger>
 <DialogContent>
 <DialogHeader>
 <DialogTitle>Delete Account</DialogTitle>
 </DialogHeader>
 <p className="text-sm text-muted-foreground">
 Are you sure you want to delete your account? This action cannot be undone.
 </p>
 <div className="flex justify-end gap-2">
 <Button variant="outline">Cancel</Button>
 <Button variant="destructive">Delete</Button>
 </div>
 </DialogContent>
 </Dialog>
 )
}
```

---

## 7. Vue 3.5 Composition API Example

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'

const schema = toTypedSchema(
 z.object({
 email: z.string().email('Invalid email adddess'),
 password: z.string().min(8, 'Password must be at least 8 characters'),
 })
)

const { defineComponentBinds, handleSubmit, errors } = useForm({
 validationSchema: schema,
})

const email = defineComponentBinds('email')
const password = defineComponentBinds('password')

const onSubmit = handleSubmit((values) => {
 console.log('Form submitted:', values)
})
</script>

<template>
 <form @submit="onSubmit" class="space-y-4">
 <div>
 <label for="email">Email</label>
 <input
 id="email"
 v-bind="email"
 type="email"
 class="border rounded px-3 py-2 w-full"
 />
 <span v-if="errors.email" class="text-red-500 text-sm">
 {{ errors.email }}
 </span>
 </div>

 <div>
 <label for="password">Password</label>
 <input
 id="password"
 v-bind="password"
 type="password"
 class="border rounded px-3 py-2 w-full"
 />
 <span v-if="errors.password" class="text-red-500 text-sm">
 {{ errors.password }}
 </span>
 </div>

 <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded">
 Submit
 </button>
 </form>
</template>
```

---

## 8. Design Token Implementation

```typescript
// tokens/colors.ts
export const colors = {
 primary: {
 50: '#eff6ff',
 100: '#dbeafe',
 500: '#3b82f6',
 900: '#1e3a8a',
 },
 semantic: {
 text: {
 primary: 'var(--color-text-primary)',
 secondary: 'var(--color-text-secondary)',
 },
 background: {
 default: 'var(--color-bg-default)',
 elevated: 'var(--color-bg-elevated)',
 },
 },
} as const

// Usage in components
import { colors } from '@/tokens/colors'

<div style={{ color: colors.semantic.text.primary }}>
 Text with semantic token
</div>
```

---

Last Updated: 2025-11-26
Status: Production Ready
