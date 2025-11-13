---
name: "moai-lib-shadcn-ui"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise shadcn/ui Component Library with AI-powered design system architecture, Context7 integration, and intelligent component orchestration for modern React applications
keywords: ['shadcn-ui', 'component-library', 'design-system', 'react', 'typescript', 'tailwind', 'radix-ui', 'context7-integration', 'ai-orchestration', 'production-deployment']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise shadcn/ui Component Library Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lib-shadcn-ui |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Component Library Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when shadcn/ui keywords detected |

---

## What It Does

Enterprise shadcn/ui Component Library expert with AI-powered design system architecture, Context7 integration, and intelligent component orchestration for modern React applications.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Component Architecture** using Context7 MCP for latest design patterns
- 📊 **Intelligent Design System** with automated theme customization and optimization
- 🚀 **Advanced Component Orchestration** with AI-driven accessibility and performance
- 🔗 **Enterprise UI Framework** with zero-configuration design tokens and branding
- 📈 **Predictive Component Analytics** with usage forecasting and optimization insights

---

## When to Use

**Automatic triggers**:
- shadcn/ui component library and design system discussions
- React component architecture and UI framework planning
- Tailwind CSS integration and design token management
- Accessibility implementation and design customization

**Manual invocation**:
- Designing enterprise design systems with shadcn/ui
- Implementing comprehensive component libraries
- Planning theme customization and branding strategies
- Optimizing component performance and accessibility

---

# Quick Reference (Level 1)

## shadcn/ui Ecosystem (November 2025)

### Core Components
- **Form Components**: Input, Select, Checkbox, Radio, Textarea
- **Display Components**: Card, Dialog, Sheet, Drawer, Popover
- **Navigation Components**: Navigation Menu, Breadcrumb, Tabs, Pagination
- **Data Components**: Table, Calendar, DatePicker, Chart Components
- **Feedback Components**: Alert, Toast, Progress, Badge, Avatar

### Foundation Technologies
- **React 19**: Latest React with server components and concurrent features
- **TypeScript 5.5**: Full type safety and enhanced developer experience
- **Tailwind CSS 3.4**: Utility-first CSS framework with JIT compilation
- **Radix UI**: Unstyled, accessible component primitives
- **Framer Motion**: Production-ready motion library for animations
- **Lucide React**: Beautifully designed icon library

### Design System Features
- **Dark Mode**: Comprehensive dark mode support with system detection
- **Responsive Design**: Mobile-first responsive components
- **Accessibility**: WCAG 2.1 AA compliance with proper ARIA support
- **Theming**: CSS variables for comprehensive theme customization
- **Animation**: Smooth animations and transitions with Framer Motion

### Integration Benefits
- **Developer Experience**: Excellent DX with TypeScript and hot reloading
- **Customization**: Fully customizable with Tailwind CSS classes
- **Performance**: Optimized bundle size and runtime performance
- **Accessibility**: Built-in accessibility with Radix UI primitives
- **Consistency**: Design system ensures consistent UI across applications

---

# Core Implementation (Level 2)

## shadcn/ui Architecture Intelligence

```python
# AI-powered shadcn/ui architecture optimization with Context7
class ShadcnUIArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.component_analyzer = ComponentAnalyzer()
        self.theme_optimizer = ThemeOptimizer()
    
    async def design_optimal_shadcn_architecture(self, 
                                               requirements: DesignSystemRequirements) -> ShadcnUIArchitecture:
        """Design optimal shadcn/ui architecture using AI analysis."""
        
        # Get latest shadcn/ui and React documentation via Context7
        shadcn_docs = await self.context7_client.get_library_docs(
            context7_library_id='/shadcn-ui/docs',
            topic="component library design system theming accessibility 2025",
            tokens=3000
        )
        
        react_docs = await self.context7_client.get_library_docs(
            context7_library_id='/react/docs',
            topic="hooks server-components performance optimization 2025",
            tokens=2000
        )
        
        # Optimize component selection
        component_selection = self.component_analyzer.optimize_component_selection(
            requirements.ui_components,
            requirements.user_needs,
            shadcn_docs
        )
        
        # Optimize theme configuration
        theme_configuration = self.theme_optimizer.optimize_theme_system(
            requirements.brand_guidelines,
            requirements.accessibility_requirements,
            shadcn_docs
        )
        
        return ShadcnUIArchitecture(
            component_library=component_selection,
            theme_system=theme_configuration,
            accessibility_compliance=self._ensure_accessibility_compliance(
                requirements.accessibility_requirements
            ),
            performance_optimization=self._optimize_component_performance(
                component_selection
            ),
            integration_patterns=self._design_integration_patterns(
                requirements.framework_stack
            ),
            customization_strategy=self._plan_customization_strategy(
                requirements.customization_needs
            )
        )
```

## Component System Implementation

```typescript
// Advanced shadcn/ui component with TypeScript and modern patterns
import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

// Component variants using class-variance-authority
const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground shadow hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90",
        outline: "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
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
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
  loading?: boolean;
  loadingText?: string;
}

// Enhanced button component with loading state and accessibility
const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, loading, loadingText, children, disabled, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || loading}
        aria-disabled={disabled || loading}
        aria-describedby={loading ? "loading-description" : undefined}
        {...props}
      >
        {loading && (
          <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
        )}
        
        {loading && loadingText ? (
          <span id="loading-description" className="sr-only">
            {loadingText}
          </span>
        ) : null}
        
        {loading ? loadingText || children : children}
      </Comp>
    );
  }
);

Button.displayName = "Button";

export { Button, buttonVariants };

// Advanced form component with validation
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

const formSchema = z.object({
  username: z.string().min(2, {
    message: "Username must be at least 2 characters.",
  }),
  email: z.string().email({
    message: "Please enter a valid email address.",
  }),
  password: z.string().min(8, {
    message: "Password must be at least 8 characters.",
  }),
});

type FormValues = z.infer<typeof formSchema>;

export function UserForm() {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      email: "",
      password: "",
    },
  });

  const onSubmit = async (data: FormValues) => {
    try {
      // Handle form submission
      console.log("Form data:", data);
    } catch (error) {
      console.error("Submission error:", error);
    }
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="username">Username</Label>
        <Input
          id="username"
          type="text"
          placeholder="Enter your username"
          {...form.register("username")}
          aria-invalid={form.formState.errors.username ? "true" : "false"}
          aria-describedby={form.formState.errors.username ? "username-error" : undefined}
        />
        {form.formState.errors.username && (
          <p id="username-error" className="text-sm text-destructive">
            {form.formState.errors.username.message}
          </p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          type="email"
          placeholder="Enter your email"
          {...form.register("email")}
          aria-invalid={form.formState.errors.email ? "true" : "false"}
          aria-describedby={form.formState.errors.email ? "email-error" : undefined}
        />
        {form.formState.errors.email && (
          <p id="email-error" className="text-sm text-destructive">
            {form.formState.errors.email.message}
          </p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Password</Label>
        <Input
          id="password"
          type="password"
          placeholder="Enter your password"
          {...form.register("password")}
          aria-invalid={form.formState.errors.password ? "true" : "false"}
          aria-describedby={form.formState.errors.password ? "password-error" : undefined}
        />
        {form.formState.errors.password && (
          <p id="password-error" className="text-sm text-destructive">
            {form.formState.errors.password.message}
          </p>
        )}
      </div>

      <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
        {form.formState.isSubmitting ? "Creating account..." : "Create account"}
      </Button>
    </form>
  );
}
```

## Theme System Implementation

```typescript
// Advanced theme system with CSS variables and dark mode
import { createContext, useContext, useEffect, useState } from "react";

type Theme = "dark" | "light" | "system";

interface ThemeProviderProps {
  children: React.ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
  attribute?: string;
  enableSystem?: boolean;
}

interface ThemeProviderState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

const ThemeProviderContext = createContext<ThemeProviderState | undefined>(undefined);

export function ThemeProvider({
  children,
  defaultTheme = "system",
  storageKey = "ui-theme",
  attribute = "class",
  enableSystem = true,
  ...props
}: ThemeProviderProps) {
  const [theme, setTheme] = useState<Theme>(() => {
    if (typeof window !== "undefined") {
      return (localStorage.getItem(storageKey) as Theme) || defaultTheme;
    }
    return defaultTheme;
  });

  useEffect(() => {
    const root = window.document.documentElement;

    root.classList.remove("light", "dark");

    if (theme === "system" && enableSystem) {
      const systemTheme = window.matchMedia("(prefers-color-scheme: dark)")
        .matches
        ? "dark"
        : "light";

      root.classList.add(systemTheme);
      return;
    }

    root.classList.add(theme);
  }, [theme, enableSystem, attribute]);

  const value = {
    theme,
    setTheme: (theme: Theme) => {
      localStorage.setItem(storageKey, theme);
      setTheme(theme);
    },
  };

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext);

  if (context === undefined)
    throw new Error("useTheme must be used within a ThemeProvider");

  return context;
};

// Theme configuration with design tokens
export const theme = {
  light: {
    background: "hsl(0 0% 100%)",
    foreground: "hsl(240 10% 3.9%)",
    card: "hsl(0 0% 100%)",
    cardForeground: "hsl(240 10% 3.9%)",
    popover: "hsl(0 0% 100%)",
    popoverForeground: "hsl(240 10% 3.9%)",
    primary: "hsl(240 9% 10%)",
    primaryForeground: "hsl(0 0% 98%)",
    secondary: "hsl(240 4.8% 95.9%)",
    secondaryForeground: "hsl(240 3.8% 46.1%)",
    muted: "hsl(240 4.8% 95.9%)",
    mutedForeground: "hsl(240 3.8% 46.1%)",
    accent: "hsl(240 4.8% 95.9%)",
    accentForeground: "hsl(240 5.9% 10%)",
    destructive: "hsl(0 72.22% 50.59%)",
    destructiveForeground: "hsl(0 0% 98%)",
    border: "hsl(240 5.9% 90%)",
    input: "hsl(240 5.9% 90%)",
    ring: "hsl(240 5.9% 10%)",
  },
  dark: {
    background: "hsl(240 10% 3.9%)",
    foreground: "hsl(0 0% 98%)",
    card: "hsl(240 10% 3.9%)",
    cardForeground: "hsl(0 0% 98%)",
    popover: "hsl(240 10% 3.9%)",
    popoverForeground: "hsl(0 0% 98%)",
    primary: "hsl(0 0% 98%)",
    primaryForeground: "hsl(240 9% 10%)",
    secondary: "hsl(240 3.7% 15.9%)",
    secondaryForeground: "hsl(0 0% 98%)",
    muted: "hsl(240 3.7% 15.9%)",
    mutedForeground: "hsl(240 5% 64.9%)",
    accent: "hsl(240 3.7% 15.9%)",
    accentForeground: "hsl(0 0% 98%)",
    destructive: "hsl(0 62.8% 30.6%)",
    destructiveForeground: "hsl(0 0% 98%)",
    border: "hsl(240 3.7% 15.9%)",
    input: "hsl(240 3.7% 15.9%)",
    ring: "hsl(240 4.9% 83.9%)",
  },
};

// Apply theme CSS variables
export function applyThemeCSS() {
  const root = document.documentElement;
  
  Object.entries(theme.light).forEach(([key, value]) => {
    root.style.setProperty(`--${key}`, value);
  });
}
```

---

# Advanced Implementation (Level 3)

## Advanced Component Patterns

```typescript
// Complex data table with shadcn/ui components
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useReactTable,
} from "@tanstack/react-table";

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  searchKey?: string;
  filterableColumns?: string[];
}

export function DataTable<TData, TValue>({
  columns,
  data,
  searchKey,
  filterableColumns = [],
}: DataTableProps<TData, TValue>) {
  const [globalFilter, setGlobalFilter] = useState("");
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    onColumnFiltersChange: setColumnFilters,
    state: {
      globalFilter,
      columnFilters,
    },
  });

  return (
    <div className="space-y-4">
      {/* Search and Filters */}
      <div className="flex items-center gap-4">
        <Input
          placeholder={`Filter ${searchKey}...`}
          value={(globalFilter ?? "") as string}
          onChange={(event) => setGlobalFilter(String(event.target.value))}
          className="max-w-sm"
        />
        
        {/* Column Filters */}
        {filterableColumns.map((column) => (
          <DataTableColumnFilter
            key={column}
            column={table.getColumn(column)}
            title={column}
          />
        ))}
      </div>

      {/* Data Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-end space-x-2 py-4">
        <Button
          variant="outline"
          size="sm"
          onClick={() => table.previousPage()}
          disabled={!table.getCanPreviousPage()}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => table.nextPage()}
          disabled={!table.getCanNextPage()}
        >
          Next
        </Button>
      </div>
    </div>
  );
}

// Advanced form with multi-step validation
import { useState } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const multiStepSchema = z.object({
  // Step 1: Personal Information
  personalInfo: z.object({
    firstName: z.string().min(2),
    lastName: z.string().min(2),
    email: z.string().email(),
    phone: z.string().optional(),
  }),
  
  // Step 2: Address
  address: z.object({
    street: z.string().min(5),
    city: z.string().min(2),
    state: z.string().min(2),
    zipCode: z.string().regex(/^\d{5}(-\d{4})?$/),
    country: z.string().min(2),
  }),
  
  // Step 3: Preferences
  preferences: z.object({
    newsletter: z.boolean(),
    notifications: z.boolean(),
    theme: z.enum(["light", "dark", "system"]),
  }),
});

type MultiStepFormValues = z.infer<typeof multiStepSchema>;

export function MultiStepForm() {
  const [currentStep, setCurrentStep] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const methods = useForm<MultiStepFormValues>({
    resolver: zodResolver(multiStepSchema),
    defaultValues: {
      personalInfo: {
        firstName: "",
        lastName: "",
        email: "",
        phone: "",
      },
      address: {
        street: "",
        city: "",
        state: "",
        zipCode: "",
        country: "",
      },
      preferences: {
        newsletter: false,
        notifications: true,
        theme: "system",
      },
    },
  });

  const steps = [
    { title: "Personal Info", component: PersonalInfoStep },
    { title: "Address", component: AddressStep },
    { title: "Preferences", component: PreferencesStep },
  ];

  const handleNext = async () => {
    const currentStepName = ["personalInfo", "address", "preferences"][currentStep];
    const isValid = await methods.trigger(currentStepName as any);
    
    if (isValid && currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handlePrevious = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const onSubmit = async (data: MultiStepFormValues) => {
    setIsSubmitting(true);
    try {
      // Handle form submission
      console.log("Form submitted:", data);
      await new Promise(resolve => setTimeout(resolve, 2000)); // Simulate API call
    } catch (error) {
      console.error("Submission error:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const progress = ((currentStep + 1) / steps.length) * 100;
  const CurrentStepComponent = steps[currentStep].component;

  return (
    <div className="w-full max-w-2xl mx-auto p-6">
      <div className="mb-8">
        <h1 className="text-2xl font-bold mb-4">Multi-Step Form</h1>
        <Progress value={progress} className="mb-2" />
        <p className="text-sm text-muted-foreground">
          Step {currentStep + 1} of {steps.length}: {steps[currentStep].title}
        </p>
      </div>

      <FormProvider {...methods}>
        <form onSubmit={methods.handleSubmit(onSubmit)}>
          <div className="mb-8">
            <CurrentStepComponent />
          </div>

          <div className="flex justify-between">
            <Button
              type="button"
              variant="outline"
              onClick={handlePrevious}
              disabled={currentStep === 0}
            >
              Previous
            </Button>

            {currentStep === steps.length - 1 ? (
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? "Submitting..." : "Submit"}
              </Button>
            ) : (
              <Button type="button" onClick={handleNext}>
                Next
              </Button>
            )}
          </div>
        </form>
      </FormProvider>
    </div>
  );
}
```

### Performance Optimization

```typescript
// Performance optimization for shadcn/ui components
export class ComponentPerformanceOptimizer {
  // Lazy loading components with React.lazy
  static createLazyComponent(importFn: () => Promise<{ default: React.ComponentType<any> }>) {
    return React.lazy(importFn);
  }

  // Component memoization with React.memo
  static createMemoizedComponent<P extends object>(
    Component: React.ComponentType<P>,
    areEqual?: (prevProps: P, nextProps: P) => boolean
  ) {
    return React.memo(Component, areEqual);
  }

  // Hook for performance monitoring
  static usePerformanceMonitor(componentName: string) {
    const [renderCount, setRenderCount] = useState(0);
    const [renderTime, setRenderTime] = useState(0);

    useEffect(() => {
      const startTime = performance.now();
      
      setRenderCount(prev => prev + 1);
      
      return () => {
        const endTime = performance.now();
        setRenderTime(endTime - startTime);
        
        if (process.env.NODE_ENV === 'development') {
          console.log(
            `${componentName} render #${renderCount}: ${renderTime.toFixed(2)}ms`
          );
        }
      };
    });

    return { renderCount, renderTime };
  }

  // Bundle size optimization
  static optimizeBundleSize() {
    return {
      treeShaking: {
        enabled: true,
        description: "Remove unused components and utilities",
      },
      codeSplitting: {
        enabled: true,
        description: "Split components into separate chunks",
      },
      compression: {
        enabled: true,
        description: "Compress bundle with gzip/brotli",
      },
    };
  }

  // Runtime performance optimization
  static optimizeRuntimePerformance() {
    return {
      virtualScrolling: {
        enabled: true,
        description: "Virtual scrolling for large data sets",
      },
      memoization: {
        enabled: true,
        description: "Memoize expensive computations",
      },
      debouncing: {
        enabled: true,
        description: "Debounce user interactions",
      },
    };
  }
}
```

---

# Reference & Integration (Level 4)

## API Reference

### Core shadcn/ui Operations
- `create_component(variant_config, props)` - Create new component
- `customize_theme(theme_config, colors)` - Customize theme
- `install_component(component_name)` - Install new component
- `setup_design_system(tokens, branding)` - Set up design system
- `optimize_components(performance_config)` - Optimize component performance

### Context7 Integration
- `get_latest_shadcn_docs()` - shadcn/ui docs via Context7
- `analyze_component_patterns()` - Component patterns via Context7
- `optimize_theme_system()` - Theme optimization via Context7

## Best Practices (November 2025)

### DO
- Use TypeScript for type safety and better developer experience
- Implement proper accessibility with ARIA attributes
- Customize themes using CSS variables for consistency
- Optimize component performance with memoization and lazy loading
- Use form validation with proper error handling
- Implement responsive design with Tailwind CSS
- Use Radix UI primitives for accessible components
- Test components across different browsers and devices

### DON'T
- Skip TypeScript type definitions and validations
- Ignore accessibility requirements and WCAG compliance
- Override component styles without understanding the CSS structure
- Forget to implement proper error handling and validation
- Skip performance optimization for large datasets
- Use inline styles instead of Tailwind CSS classes
- Forget to implement proper keyboard navigation
- Skip component testing and user experience validation

## Works Well With

- `moai-domain-frontend` (Frontend development patterns)
- `moai-baas-foundation` (Enterprise UI architecture)
- `moai-security-api` (UI security implementation)
- `moai-essentials-perf` (Performance optimization)
- `moai-foundation-trust` (Accessibility and compliance)
- `moai-domain-backend` (Backend integration)
- `moai-baas-vercel-ext` (Frontend deployment)
- `moai-domain-testing` (Component testing strategies)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 shadcn/ui ecosystem updates, and advanced component patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, component patterns, theme system
- **v1.0.0** (2025-11-11): Initial shadcn/ui component library

---

**End of Skill** | Updated 2025-11-13

## Design System Integration

### Enterprise Features
- Comprehensive component library with enterprise-grade features
- Advanced theming system with brand customization
- Performance optimization for large-scale applications
- Accessibility compliance with WCAG 2.1 AA standards

### Modern Development
- TypeScript-first development with full type safety
- Modern React patterns with hooks and server components
- Tailwind CSS integration for consistent styling
- Framer Motion for smooth animations and transitions

---

**End of Enterprise shadcn/ui Component Library Expert v4.0.0**
