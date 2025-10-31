# moai-design-shadcn-ui

Master shadcn/ui component customization, Radix UI primitives, accessibility, and dark mode.

## Quick Start

shadcn/ui is a collection of copy-paste React components built on Radix UI and Tailwind CSS. Use this skill when building component libraries, customizing component appearance, implementing accessible UI patterns, or creating consistent design systems.

## Core Patterns

### Pattern 1: Installing & Customizing Components

**Pattern**: Install and customize shadcn/ui components to match your design system.

```bash
# Install shadcn/ui CLI
npm install -D shadcn-ui

# Initialize shadcn/ui in your project
npx shadcn-ui@latest init

# Install specific components
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add dialog
```

```typescript
// components/ui/button.tsx - Customized button component
import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
        icon: "h-10 w-10",
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

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
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

export { Button, buttonVariants }
```

**When to use**:
- Building consistent UI systems
- Implementing accessible components
- Creating design tokens that scale
- Customizing existing components for brand

**Key benefits**:
- Built on accessible primitives (Radix UI)
- Full TypeScript support
- Utility-first styling (Tailwind)
- Copy-paste allows full customization

### Pattern 2: Building Complex Components

**Pattern**: Compose shadcn/ui components to create complex UI patterns.

```typescript
// components/UserForm.tsx - Complex form composition
'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';

import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';

const formSchema = z.object({
  username: z.string().min(2).max(50),
  email: z.string().email(),
  role: z.enum(['admin', 'user', 'viewer']),
  bio: z.string().max(500).optional(),
});

type FormValues = z.infer<typeof formSchema>;

export function UserForm() {
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      email: '',
      role: 'user',
      bio: '',
    },
  });

  async function onSubmit(values: FormValues) {
    setIsLoading(true);
    try {
      const response = await fetch('/api/users', {
        method: 'POST',
        body: JSON.stringify(values),
      });
      // Handle success
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Card className="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>User Profile</CardTitle>
        <CardDescription>Update your profile information</CardDescription>
      </CardHeader>
      <CardContent>
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
              name="role"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Role</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="admin">Administrator</SelectItem>
                      <SelectItem value="user">User</SelectItem>
                      <SelectItem value="viewer">Viewer</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="bio"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Bio</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Tell us about yourself"
                      className="resize-none"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Maximum 500 characters
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Saving...' : 'Save Changes'}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
```

**When to use**:
- Building form interfaces with validation
- Creating modals, dialogs, and overlays
- Building data tables with sorting/filtering
- Creating navigation patterns

**Key benefits**:
- Composable component architecture
- Built-in accessibility patterns
- State management integrated
- Type-safe through TypeScript

### Pattern 3: Dark Mode & Theme Variants

**Pattern**: Implement dark mode and theme variants using CSS custom properties.

```typescript
// app/layout.tsx - Theme provider setup
import { ThemeProvider } from '@/components/ThemeProvider';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}

// components/ThemeToggle.tsx - Toggle component
'use client';

import { useTheme } from 'next-themes';
import { Button } from '@/components/ui/button';
import { Moon, Sun } from 'lucide-react';

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
    >
      <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  );
}

// tailwind.config.ts - Theme configuration
import type { Config } from "tailwindcss"

const config = {
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
} satisfies Config

export default config
```

**When to use**:
- Implementing user preference for light/dark mode
- Supporting system theme preference
- Creating multiple color schemes
- Brand customization and white-labeling

**Key benefits**:
- CSS variables for dynamic theming
- System preference detection
- Smooth transitions between themes
- Easy customization without code changes

## Progressive Disclosure

### Level 1: Basic Components
- Install and use pre-built components (button, card, input)
- Basic customization via props and className
- Dark mode support with minimal setup

### Level 2: Advanced Patterns
- Component composition for complex UI
- Custom variants using CVA (Class Variance Authority)
- Form integration with validation
- Dialog and modal patterns

### Level 3: Expert Design Systems
- Creating custom component libraries
- Design token management
- Building accessible form patterns
- Theme extension and customization

## Works Well With

- **Tailwind CSS 4**: Utility-first styling foundation
- **Radix UI**: Headless component primitives
- **React Hook Form**: Form state and validation
- **Zod**: TypeScript-first schema validation
- **Next.js 16**: Server components and streaming
- **TypeScript**: Full type safety for components

## References

- **Official Website**: https://ui.shadcn.com
- **Components**: https://ui.shadcn.com/docs/components
- **Installation**: https://ui.shadcn.com/docs/installation
- **Radix UI**: https://www.radix-ui.com
- **CVA (Class Variance Authority)**: https://cva.style
