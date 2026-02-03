name: moai-foundation-uiux-components
description: Atomic Design component patterns and implementation

## Atomic Design Component Structure

### Folder Hierarchy

```
src/design-system/
 tokens/ # Design tokens (DTCG format)
 color.json
 typography.json
 spacing.json
 components/
 atoms/ # Basic building blocks
 Button/
 Button.tsx
 Button.stories.tsx
 Button.test.tsx
 index.ts
 Input/
 Icon/
 molecules/ # Simple combinations
 FormField/
 SearchBar/
 Card/
 organisms/ # Complex sections
 Header/
 Footer/
 DataTable/
 templates/ # Page layouts
 DashboardLayout/
 AuthLayout/
 styles/
 global.css
 theme.css
```

### Component Props API Pattern

```typescript
// atoms/Button/Button.tsx
import { forwardRef } from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

const buttonVariants = cva(
 'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50',
 {
 variants: {
 variant: {
 primary: 'bg-primary-500 text-white hover:bg-primary-600',
 secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300',
 outline: 'border border-gray-300 bg-transparent hover:bg-gray-100',
 ghost: 'hover:bg-gray-100',
 danger: 'bg-red-500 text-white hover:bg-red-600'
 },
 size: {
 sm: 'h-8 px-3 text-sm',
 md: 'h-10 px-4 text-base',
 lg: 'h-12 px-6 text-lg'
 }
 },
 defaultVariants: {
 variant: 'primary',
 size: 'md'
 }
 }
);

export interface ButtonProps
 extends React.ButtonHTMLAttributes<HTMLButtonElement>,
 VariantProps<typeof buttonVariants> {
 isLoading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
 ({ className, variant, size, isLoading, children, disabled, ...props }, ref) => {
 return (
 <button
 ref={ref}
 className={buttonVariants({ variant, size, className })}
 disabled={disabled || isLoading}
 aria-busy={isLoading}
 {...props}
 >
 {isLoading ? <Spinner /> : children}
 </button>
 );
 }
);

Button.displayName = 'Button';
```

### Storybook Documentation Setup

Installation & Configuration:

```bash
npx storybook@latest init
```

Storybook Configuration (v8.x):

```typescript
// .storybook/main.ts
import type { StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
 stories: ['../src//*.stories.@(js|jsx|ts|tsx|mdx)'],
 addons: [
 '@storybook/addon-links',
 '@storybook/addon-essentials',
 '@storybook/addon-interactions',
 '@storybook/addon-a11y', // Accessibility testing
 ],
 framework: {
 name: '@storybook/react-vite',
 options: {},
 },
 docs: {
 autodocs: 'tag', // Auto-generate docs
 },
};

export default config;
```

### Component Story with Accessibility Tests

```typescript
// components/atoms/Button/Button.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta: Meta<typeof Button> = {
 title: 'Atoms/Button',
 component: Button,
 tags: ['autodocs'],
 argTypes: {
 variant: {
 control: 'select',
 options: ['primary', 'secondary', 'outline', 'ghost', 'danger'],
 },
 size: {
 control: 'select',
 options: ['sm', 'md', 'lg'],
 },
 },
};

export default meta;
type Story = StoryObj<typeof Button>;

export const Primary: Story = {
 args: {
 children: 'Primary Button',
 variant: 'primary',
 },
};

export const AllVariants: Story = {
 render: () => (
 <div className="flex gap-4">
 <Button variant="primary">Primary</Button>
 <Button variant="secondary">Secondary</Button>
 <Button variant="outline">Outline</Button>
 <Button variant="ghost">Ghost</Button>
 <Button variant="danger">Danger</Button>
 </div>
 ),
};

export const Disabled: Story = {
 args: {
 children: 'Disabled Button',
 disabled: true,
 },
};

export const Loading: Story = {
 args: {
 children: 'Loading Button',
 isLoading: true,
 },
};
```

### Input Component with Accessibility

```typescript
// components/atoms/Input/Input.tsx
export const Input = forwardRef<HTMLInputElement, InputProps>(
 ({ label, error, required, ...props }, ref) => {
 const inputId = useId();
 const errorId = `${inputId}-error`;
 
 return (
 <div className="form-field">
 <label htmlFor={inputId} className="form-label">
 {label}
 {required && <span aria-label="required">*</span>}
 </label>
 
 <input
 ref={ref}
 id={inputId}
 aria-invalid={!!error}
 aria-describedby={error ? errorId : undefined}
 aria-required={required}
 {...props}
 />
 
 {error && (
 <span id={errorId} role="alert" className="error-message">
 {error}
 </span>
 )}
 </div>
 );
 }
);
```

---

Last Updated: 2025-11-26
Related: [Main Skill](../SKILL.md), [Accessibility WCAG](accessibility-wcag.md)
