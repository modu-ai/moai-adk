# moai-lib-shadcn-ui

**shadcn/ui: Accessible, Customizable React Components Built with Tailwind CSS & Radix UI**

> **Primary Agent**: frontend-expert
> **Secondary Agent**: ui-ux-expert
> **Version**: 1.0.0 (shadcn/ui v2.0+)
> **Keywords**: shadcn, shadcn/ui, react components, tailwind css, radix ui, accessibility, typescript, component library

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**shadcn/ui** provides a collection of beautifully designed, accessible, customizable React components built with TypeScript, Tailwind CSS, and Radix UI:

| Component Category | Examples | Use Case |
|------------------|----------|----------|
| **Input** | Button, Input, Textarea, Checkbox, Radio | Form inputs |
| **Select** | Select, ComboBox, Popover | Selection components |
| **Display** | Card, Badge, Alert, Toast | Content display |
| **Navigation** | Navigation Menu, Tabs, Breadcrumb | Site navigation |
| **Dialog** | Dialog, Alert Dialog, Drawer | Modal interactions |
| **Layout** | Sidebar, Tooltip, Dropdown Menu | Page structure |
| **Advanced** | Data Table, Calendar, Date Picker | Complex features |

**Key Features**:
- ✅ Copy-paste installation (no node_modules bloat)
- ✅ Built on Radix UI (accessible primitives)
- ✅ Styled with Tailwind CSS (fully customizable)
- ✅ TypeScript support (full type safety)
- ✅ Unstyled compound components (flexible composition)

**Installation**:
```bash
# Install shadcn/ui CLI
npx shadcn@latest init

# Add components as needed
npx shadcn@latest add button
npx shadcn@latest add card
npx shadcn@latest add dialog
```

**Configuration** (`components.json`):
```json
{
  "style": "default",
  "tailwind": {
    "config": "tailwind.config.js",
    "css": "src/globals.css",
    "baseColor": "slate"
  },
  "rsc": true,
  "tsx": true,
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils"
  }
}
```

---

### Level 2: Practical Implementation (Common Patterns)

#### Pattern 1: Button Component - All Variants

```tsx
import { Button } from "@/components/ui/button"

export function ButtonDemo() {
  return (
    <div className="flex flex-wrap gap-4">
      {/* Primary variant */}
      <Button>Click me</Button>

      {/* Secondary variant */}
      <Button variant="secondary">Secondary</Button>

      {/* Destructive (error) variant */}
      <Button variant="destructive">Delete</Button>

      {/* Outline variant */}
      <Button variant="outline">Outline</Button>

      {/* Ghost variant (transparent) */}
      <Button variant="ghost">Ghost</Button>

      {/* Link variant */}
      <Button variant="link">Link</Button>

      {/* Size variants */}
      <Button size="sm">Small</Button>
      <Button size="default">Default</Button>
      <Button size="lg">Large</Button>

      {/* Loading state */}
      <Button disabled>
        <span className="mr-2">⏳</span>
        Loading...
      </Button>

      {/* With Icon */}
      <Button>
        <svg className="mr-2 h-4 w-4" fill="none" stroke="currentColor">
          {/* Icon SVG */}
        </svg>
        With Icon
      </Button>
    </div>
  )
}
```

#### Pattern 2: Card Component - Layout Building Block

```tsx
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export function CardExample() {
  return (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Product Details</CardTitle>
        <CardDescription>
          View information about this product
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium">Product Name</label>
            <p className="text-gray-600">Widget Pro</p>
          </div>
          <div>
            <label className="text-sm font-medium">Price</label>
            <p className="text-2xl font-bold">$99.99</p>
          </div>
          <div>
            <label className="text-sm font-medium">Description</label>
            <p className="text-sm text-gray-600">
              High-quality product with excellent features
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
```

#### Pattern 3: Form with Validation

```tsx
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import React from "react"

export function FormExample() {
  const [errors, setErrors] = React.useState<Record<string, string>>({})

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)

    // Validate
    const newErrors: Record<string, string> = {}
    const email = formData.get("email") as string

    if (!email) {
      newErrors.email = "Email is required"
    } else if (!email.includes("@")) {
      newErrors.email = "Invalid email format"
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }

    // Submit form
    console.log("Form submitted:", Object.fromEntries(formData))
    setErrors({})
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Contact Form</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Name Input */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium mb-1">
              Full Name
            </label>
            <Input
              id="name"
              name="name"
              type="text"
              placeholder="John Doe"
              required
            />
          </div>

          {/* Email Input */}
          <div>
            <label htmlFor="email" className="block text-sm font-medium mb-1">
              Email
            </label>
            <Input
              id="email"
              name="email"
              type="email"
              placeholder="john@example.com"
              required
              className={errors.email ? "border-red-500" : ""}
            />
            {errors.email && (
              <p className="text-red-500 text-sm mt-1">{errors.email}</p>
            )}
          </div>

          {/* Message Textarea */}
          <div>
            <label htmlFor="message" className="block text-sm font-medium mb-1">
              Message
            </label>
            <textarea
              id="message"
              name="message"
              rows={4}
              placeholder="Your message..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>

          {/* Submit Button */}
          <Button type="submit" className="w-full">
            Send Message
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
```

#### Pattern 4: Dialog (Modal) Component

```tsx
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"

export function DialogExample() {
  return (
    <Dialog>
      {/* Trigger Button */}
      <DialogTrigger asChild>
        <Button>Open Dialog</Button>
      </DialogTrigger>

      {/* Dialog Content */}
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Confirm Action</DialogTitle>
          <DialogDescription>
            Are you sure you want to proceed with this action?
          </DialogDescription>
        </DialogHeader>

        {/* Dialog Body */}
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            This action cannot be undone. Please confirm that you want to continue.
          </p>

          {/* Dialog Footer with Actions */}
          <div className="flex justify-end gap-2">
            {/* Close Button */}
            <DialogTrigger asChild>
              <Button variant="outline">Cancel</Button>
            </DialogTrigger>

            {/* Confirm Button */}
            <Button variant="destructive">Confirm</Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
```

#### Pattern 5: Tabs Component for Navigation

```tsx
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Card, CardContent } from "@/components/ui/card"

export function TabsExample() {
  return (
    <Tabs defaultValue="overview" className="w-full">
      {/* Tab Triggers */}
      <TabsList className="grid w-full grid-cols-3">
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="details">Details</TabsTrigger>
        <TabsTrigger value="reviews">Reviews</TabsTrigger>
      </TabsList>

      {/* Tab Content - Overview */}
      <TabsContent value="overview">
        <Card>
          <CardContent className="pt-6">
            <h3 className="font-bold mb-2">Product Overview</h3>
            <p className="text-gray-600">
              This is the overview tab content. Provide a general summary of the product here.
            </p>
          </CardContent>
        </Card>
      </TabsContent>

      {/* Tab Content - Details */}
      <TabsContent value="details">
        <Card>
          <CardContent className="pt-6">
            <h3 className="font-bold mb-2">Product Details</h3>
            <ul className="list-disc list-inside space-y-1 text-gray-600">
              <li>Feature 1</li>
              <li>Feature 2</li>
              <li>Feature 3</li>
            </ul>
          </CardContent>
        </Card>
      </TabsContent>

      {/* Tab Content - Reviews */}
      <TabsContent value="reviews">
        <Card>
          <CardContent className="pt-6">
            <h3 className="font-bold mb-2">Customer Reviews</h3>
            <div className="space-y-3">
              <div className="border-l-4 border-yellow-400 pl-4">
                <p className="font-medium">⭐⭐⭐⭐⭐ Great product!</p>
                <p className="text-sm text-gray-600">Very satisfied with my purchase.</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}
```

#### Pattern 6: Data Table with Sorting & Filtering

```tsx
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import React from "react"

interface User {
  id: string
  name: string
  email: string
  status: "active" | "inactive"
}

const users: User[] = [
  { id: "1", name: "John Doe", email: "john@example.com", status: "active" },
  { id: "2", name: "Jane Smith", email: "jane@example.com", status: "active" },
  { id: "3", name: "Bob Wilson", email: "bob@example.com", status: "inactive" },
]

export function DataTableExample() {
  const [sortField, setSortField] = React.useState<keyof User>("name")
  const [sortOrder, setSortOrder] = React.useState<"asc" | "desc">("asc")

  const handleSort = (field: keyof User) => {
    if (sortField === field) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc")
    } else {
      setSortField(field)
      setSortOrder("asc")
    }
  }

  const sortedUsers = [...users].sort((a, b) => {
    const aVal = a[sortField]
    const bVal = b[sortField]
    const comparison = aVal < bVal ? -1 : aVal > bVal ? 1 : 0
    return sortOrder === "asc" ? comparison : -comparison
  })

  return (
    <div className="border rounded-lg">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead
              onClick={() => handleSort("name")}
              className="cursor-pointer hover:bg-gray-100"
            >
              Name {sortField === "name" && (sortOrder === "asc" ? "↑" : "↓")}
            </TableHead>
            <TableHead
              onClick={() => handleSort("email")}
              className="cursor-pointer hover:bg-gray-100"
            >
              Email {sortField === "email" && (sortOrder === "asc" ? "↑" : "↓")}
            </TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sortedUsers.map((user) => (
            <TableRow key={user.id}>
              <TableCell className="font-medium">{user.name}</TableCell>
              <TableCell>{user.email}</TableCell>
              <TableCell>
                <span
                  className={`inline-block px-2 py-1 rounded text-sm ${
                    user.status === "active"
                      ? "bg-green-100 text-green-800"
                      : "bg-gray-100 text-gray-800"
                  }`}
                >
                  {user.status}
                </span>
              </TableCell>
              <TableCell>
                <Button size="sm" variant="ghost">
                  Edit
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
```

---

### Level 3: Advanced Patterns (Expert Reference)

#### Advanced Pattern 1: Custom Component Composition with asChild

```tsx
import { Button } from "@/components/ui/button"
import Link from "next/link"

// Using asChild to render as a different element
export function NavigationLink() {
  return (
    <Button asChild>
      <Link href="/dashboard">Go to Dashboard</Link>
    </Button>
  )
}

// Custom link component with Button styling
export function StyledLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Button variant="link" asChild>
      <a href={href}>{children}</a>
    </Button>
  )
}
```

#### Advanced Pattern 2: Custom Theme with CSS Variables

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
      },
    },
  },
}
```

```css
/* globals.css */
@import "tailwindcss";

@theme {
  --border: 214.3 31.8% 91.4%;
  --input: 214.3 31.8% 91.4%;
  --ring: 222.2 84% 4.9%;
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  --secondary: 210 40% 96%;
  --secondary-foreground: 222.2 47.4% 11.2%;
}

.dark {
  --border: 217.2 32.6% 17.5%;
  --input: 217.2 32.6% 17.5%;
  --ring: 212.7 26.8% 83.9%;
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  --primary: 210 40% 98%;
  --primary-foreground: 222.2 47.4% 11.2%;
  --secondary: 217.2 32.6% 17.5%;
  --secondary-foreground: 210 40% 98%;
}
```

#### Advanced Pattern 3: Accessible Modal with Focus Management

```tsx
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import React from "react"

export function AccessibleModal() {
  const triggerRef = React.useRef<HTMLButtonElement>(null)
  const [open, setOpen] = React.useState(false)

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen)
    if (!newOpen) {
      // Restore focus to trigger button
      triggerRef.current?.focus()
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button ref={triggerRef}>Open Modal</Button>
      </DialogTrigger>
      <DialogContent
        aria-describedby="modal-description"
        onEscapeKeyDown={() => handleOpenChange(false)}
      >
        <DialogHeader>
          <DialogTitle>Important Notice</DialogTitle>
        </DialogHeader>
        <div id="modal-description" className="space-y-4">
          <p>
            This is an accessible modal dialog. Press ESC to close or click outside.
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => handleOpenChange(false)}
            >
              Cancel
            </Button>
            <Button onClick={() => handleOpenChange(false)}>
              Confirm
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
```

---

## 🎯 Best Practices

### 1. Component Organization
```
src/components/
├── ui/              # shadcn/ui components (copy-pasted)
│   ├── button.tsx
│   ├── card.tsx
│   └── dialog.tsx
├── common/          # Reusable custom components
│   ├── Header.tsx
│   ├── Footer.tsx
│   └── Navigation.tsx
└── features/        # Feature-specific components
    ├── Dashboard/
    └── Profile/
```

### 2. Accessibility Checklist
- ✅ Use semantic HTML (components handle this)
- ✅ Proper ARIA labels for complex interactions
- ✅ Keyboard navigation support (built-in via Radix UI)
- ✅ Focus management (Dialog, Modal handle this)
- ✅ Color contrast (follow Tailwind color scales)

### 3. Customization Strategy
```tsx
// Extend shadcn/ui component with custom styling
import { Button } from "@/components/ui/button"

export function CustomButton(props) {
  return (
    <Button
      {...props}
      className="rounded-full font-bold uppercase" // Add custom classes
    />
  )
}
```

### 4. Performance Optimization
```typescript
// Use React.memo for expensive components
import { memo } from "react"
import { Card } from "@/components/ui/card"

export const MemoizedCard = memo(function Card() {
  return <Card>Content</Card>
})
```

---

## 📚 Official References

- **shadcn/ui Docs**: https://ui.shadcn.com/
- **shadcn/ui GitHub**: https://github.com/shadcn-ui/ui
- **Radix UI Docs**: https://www.radix-ui.com/
- **shadcn/ui Components**: https://ui.shadcn.com/docs/components/button
- **Tailwind CSS**: https://tailwindcss.com/

---

## 🔗 Related Skills

- `Skill("moai-lang-tailwind-css")` – Tailwind CSS framework
- `Skill("moai-lang-html-css")` – HTML/CSS foundations
- `Skill("moai-domain-frontend")` – Full React/frontend architecture
