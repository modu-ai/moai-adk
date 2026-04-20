name: moai-library-shadcn-components
description: Advanced shadcn/ui component patterns and implementations

## Advanced Component Patterns

### Complex Data Table Implementation

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
```

### Multi-Step Form Implementation

```typescript
// Advanced form with multi-step validation
import { useState } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";

const multiStepSchema = z.object({
 // Step 1: Personal Information
 personalInfo: z.object({
 firstName: z.string().min(2),
 lastName: z.string().min(2),
 email: z.string().email(),
 phone: z.string().optional(),
 }),
 
 // Step 2: Adddess
 adddess: z.object({
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
 adddess: {
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
 { title: "Adddess", component: AdddessStep },
 { title: "Preferences", component: PreferencesStep },
 ];

 const handleNext = async () => {
 const currentStepName = ["personalInfo", "adddess", "preferences"][currentStep];
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
 console.log("Form submitted:", data);
 await new Promise(resolve => setTimeout(resolve, 2000));
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

### Enhanced Button Component

```typescript
// Enhanced button component with loading state and accessibility
import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

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

Last Updated: 2025-11-26
Related: [Main Skill](../SKILL.md), [shadcn Theming](shadcn-theming.md)
