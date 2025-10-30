---
title: shadcn/ui Component Patterns
description: Master shadcn/ui component customization, accessibility, and Radix UI integration patterns
freedom_level: high
tier: design
updated: 2025-10-31
---

# shadcn/ui Component Patterns

## Overview

shadcn/ui is a collection of reusable React components built on Radix UI primitives and styled with Tailwind CSS. Unlike traditional component libraries, shadcn/ui components are copied directly into your project, giving you full ownership and customization control. This skill covers component patterns, theming, accessibility, and best practices for 2025.

## Key Patterns

### 1. Component Installation & Customization

**Pattern**: Install components via CLI and customize directly in your codebase.

```bash
# Install shadcn/ui CLI
npx shadcn@latest init

# Add individual components
npx shadcn@latest add button
npx shadcn@latest add dialog
npx shadcn@latest add form

# Components are copied to components/ui/
# You now own the code - customize freely!
```

```typescript
// components/ui/button.tsx (auto-generated, now yours to modify)
import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground shadow hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90",
        outline: "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
        // 🎨 Add custom variant
        success: "bg-green-600 text-white shadow hover:bg-green-700"
      },
      size: {
        default: "h-9 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-10 rounded-md px-8",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"
```

**Benefits**: No version conflicts, full code control, easy customization, no bundle bloat from unused components.

### 2. Theming with CSS Variables

**Pattern**: shadcn/ui uses CSS variables for consistent theming across all components.

```css
/* app/globals.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    /* Light mode */
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
    --radius: 0.5rem;
  }

  .dark {
    /* Dark mode */
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    /* ... other dark mode colors */
  }
  
  /* 🎨 Custom theme: Brand colors */
  .theme-ocean {
    --primary: 199 89% 48%;
    --primary-foreground: 0 0% 100%;
    --secondary: 199 89% 85%;
    --accent: 173 58% 39%;
  }
}
```

```typescript
// Apply theme dynamically
export function ThemeToggle() {
  const [theme, setTheme] = useState('light')
  
  return (
    <div className={theme}>
      <Button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
        Toggle Theme
      </Button>
    </div>
  )
}
```

### 3. Accessible Forms with React Hook Form

**Pattern**: Combine shadcn/ui Form components with react-hook-form for type-safe validation.

```typescript
'use client'

import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"

const formSchema = z.object({
  username: z.string().min(3, "Username must be at least 3 characters"),
  email: z.string().email("Invalid email address"),
  age: z.coerce.number().min(18, "Must be 18 or older"),
})

export function ProfileForm() {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      email: "",
    },
  })

  async function onSubmit(values: z.infer<typeof formSchema>) {
    console.log(values)
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input placeholder="johndoe" {...field} />
              </FormControl>
              <FormDescription>Your public display name</FormDescription>
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
        
        <Button type="submit">Submit</Button>
      </form>
    </Form>
  )
}
```

**Benefits**: ARIA labels auto-generated, keyboard navigation, screen reader support, type-safe validation.

### 4. Dialog/Modal Patterns

**Pattern**: Use Radix Dialog primitives for accessible modals.

```typescript
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"

export function DeleteConfirmDialog() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="destructive">Delete Item</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Are you absolutely sure?</DialogTitle>
          <DialogDescription>
            This action cannot be undone. This will permanently delete your
            item from our servers.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline">Cancel</Button>
          <Button variant="destructive" onClick={handleDelete}>
            Confirm Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

**Accessibility Features**: Focus trap, Esc key closes, focus returns to trigger, ARIA attributes, backdrop click closes.

### 5. Composing Complex Components

**Pattern**: Combine primitives to build custom components.

```typescript
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

export function UserCard({ user }: { user: User }) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-4">
          <Avatar>
            <AvatarImage src={user.avatar} />
            <AvatarFallback>{user.initials}</AvatarFallback>
          </Avatar>
          <div>
            <CardTitle>{user.name}</CardTitle>
            <CardDescription>{user.email}</CardDescription>
          </div>
          {user.isPremium && <Badge variant="default">Premium</Badge>}
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">{user.bio}</p>
      </CardContent>
      <CardFooter className="flex justify-between">
        <Button variant="outline">View Profile</Button>
        <Button>Send Message</Button>
      </CardFooter>
    </Card>
  )
}
```

### 6. Data Tables with Sorting & Filtering

**Pattern**: Use shadcn/ui Table components with TanStack Table.

```typescript
import { ColumnDef } from "@tanstack/react-table"
import { DataTable } from "@/components/ui/data-table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ArrowUpDown } from "lucide-react"

type Payment = {
  id: string
  amount: number
  status: "pending" | "processing" | "success" | "failed"
  email: string
}

const columns: ColumnDef<Payment>[] = [
  {
    accessorKey: "email",
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Email
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
  },
  {
    accessorKey: "amount",
    header: "Amount",
    cell: ({ row }) => {
      const amount = parseFloat(row.getValue("amount"))
      const formatted = new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: "USD",
      }).format(amount)
      return <div className="font-medium">{formatted}</div>
    },
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const status = row.getValue("status") as string
      return (
        <Badge variant={status === "success" ? "default" : "secondary"}>
          {status}
        </Badge>
      )
    },
  },
]

export function PaymentsTable({ data }: { data: Payment[] }) {
  return <DataTable columns={columns} data={data} />
}
```

### 7. Responsive Design Patterns

**Pattern**: Use Tailwind breakpoints with shadcn/ui components.

```typescript
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { Menu } from "lucide-react"

export function ResponsiveNav() {
  return (
    <nav>
      {/* Desktop: Horizontal nav */}
      <div className="hidden md:flex gap-4">
        <Button variant="ghost">Home</Button>
        <Button variant="ghost">Products</Button>
        <Button variant="ghost">About</Button>
      </div>
      
      {/* Mobile: Sheet drawer */}
      <Sheet>
        <SheetTrigger asChild className="md:hidden">
          <Button variant="ghost" size="icon">
            <Menu />
          </Button>
        </SheetTrigger>
        <SheetContent side="left">
          <div className="flex flex-col gap-4 mt-8">
            <Button variant="ghost" className="justify-start">Home</Button>
            <Button variant="ghost" className="justify-start">Products</Button>
            <Button variant="ghost" className="justify-start">About</Button>
          </div>
        </SheetContent>
      </Sheet>
    </nav>
  )
}
```

## Checklist

- [ ] Initialize shadcn/ui: `npx shadcn@latest init`
- [ ] Configure CSS variables for theme consistency (light/dark mode)
- [ ] Use Form components with react-hook-form + zod for validation
- [ ] Verify ARIA attributes on interactive components (use axe DevTools)
- [ ] Test keyboard navigation (Tab, Enter, Esc, Arrow keys)
- [ ] Add custom variants to button/badge/card components as needed
- [ ] Use Radix primitives directly for advanced customization
- [ ] Test responsive patterns on mobile/tablet/desktop breakpoints
- [ ] Implement dark mode toggle with CSS variable switching

## Resources

- **Official Documentation**: https://www.shadcn.io/ui
- **Component Source Code**: https://github.com/shadcn-ui/ui
- **Radix UI Primitives**: https://www.radix-ui.com/primitives
- **Tailwind CSS**: https://tailwindcss.com
- **Installation Guide (2025)**: https://markaicode.com/shadcn-ui-installation-customization-guide-2025/
- **Accessibility Best Practices**: https://www.radix-ui.com/primitives/docs/overview/accessibility

---

**Version**: 1.0.0  
**Last Updated**: 2025-10-31  
**Model Recommendation**: Sonnet (component composition and accessibility reasoning)
