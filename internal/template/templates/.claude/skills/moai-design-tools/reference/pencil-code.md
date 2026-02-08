# Pencil-to-Code Export Guide

Export .pen designs to production-ready React components with Tailwind CSS styling.

## Overview

Pencil-to-code export converts finalized .pen designs into React components with Tailwind CSS classes, maintaining design fidelity while providing customizable implementation options.

## Export Configuration

### Setup

```typescript
// pencil.config.js
module.exports = {
  framework: 'react',
  styling: 'tailwind',
  output: './src/components/generated',
  options: {
    typescript: true,
    responsive: true,
    accessibility: true,
    testing: true
  }
};
```

### Export Options

```typescript
const exportOptions = {
  // Component format
  format: 'react',           // react, vue, svelte
  language: 'typescript',    // typescript, javascript

  // Styling
  styling: 'tailwind',       // tailwind, css-in-js, css-modules
  designTokens: true,        // Use design tokens instead of hardcoded values

  // Structure
  components: true,          // Generate separate component files
  stories: true,             // Generate Storybook stories
  tests: true,               // Generate test files

  // Optimization
  minify: false,             // Minify output
  treeShaking: true,         // Enable tree-shaking

  // Documentation
  props: true,               // Generate props interface
  comments: true,            // Add JSDoc comments
  examples: true             // Generate usage examples
};
```

## Component Generation

### Basic Component

```typescript
// Input: .pen design
// Output: Button.tsx

import { ButtonHTMLAttributes, forwardRef } from 'react';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'tertiary';
  size?: 'small' | 'medium' | 'large';
  isLoading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'medium', isLoading, children, ...props }, ref) => {
    const baseStyles = 'inline-flex items-center justify-center font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2';

    const variantStyles = {
      primary: 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500',
      secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 focus:ring-gray-500',
      tertiary: 'bg-transparent text-gray-700 hover:bg-gray-100 focus:ring-gray-500'
    };

    const sizeStyles = {
      small: 'px-3 py-1.5 text-sm',
      medium: 'px-4 py-2 text-base',
      large: 'px-6 py-3 text-lg'
    };

    return (
      <button
        ref={ref}
        className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${isLoading ? 'opacity-75 cursor-not-allowed' : ''}`}
        disabled={isLoading}
        {...props}
      >
        {isLoading ? (
          <svg className="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        ) : null}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
```

### Layout Component

```typescript
// Input: .pen layout design
// Output: LoginForm.tsx

import { useState } from 'react';
import { Button } from './Button';
import { Input } from './Input';

interface LoginFormProps {
  onSubmit: (email: string, password: string) => void;
}

export const LoginForm = ({ onSubmit }: LoginFormProps) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(email, password);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
      <Input
        type="email"
        label="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="you@example.com"
        required
      />

      <Input
        type="password"
        label="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="••••••••"
        required
      />

      <Button type="submit" variant="primary" size="medium">
        Sign In
      </Button>
    </form>
  );
};
```

## Tailwind CSS Integration

### Design Token Mapping

```css
/* tailwind.config.js */
module.exports = {
  theme: {
    extend: {
      colors: {
        // Map design tokens to Tailwind
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',  /* From .pen design tokens */
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        }
      },
      spacing: {
        /* Custom spacing from design */
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
      }
    }
  }
};
```

### Responsive Classes

```typescript
// Responsive design from .pen
<div className="
  grid
  grid-cols-1
  md:grid-cols-2
  lg:grid-cols-3
  gap-4
  md:gap-6
  lg:gap-8
">
  {/* Cards */}
</div>
```

## Props API Design

### Component Props

```typescript
// Generated props interface
export interface CardProps {
  // Content
  children: React.ReactNode;
  title?: string;
  description?: string;

  // Styling
  variant?: 'default' | 'bordered' | 'elevated';
  padding?: 'none' | 'small' | 'medium' | 'large';

  // State
  isLoading?: boolean;
  isDisabled?: boolean;

  // Events
  onClick?: () => void;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;

  // Accessibility
  ariaLabel?: string;
  role?: string;
}
```

### State Management

```typescript
// Stateful component with hooks
export const ToggleButton = () => {
  const [isToggled, setIsToggled] = useState(false);

  return (
    <button
      onClick={() => setIsToggled(!isToggled)}
      className={`
        relative inline-flex h-6 w-11 items-center rounded-full
        transition-colors
        ${isToggled ? 'bg-blue-600' : 'bg-gray-200'}
      `}
    >
      <span
        className={`
          inline-block h-4 w-4 transform rounded-full bg-white
          transition-transform
          ${isToggled ? 'translate-x-6' : 'translate-x-1'}
        `}
      />
    </button>
  );
};
```

## Testing Generated Components

### Unit Tests

```typescript
// Button.test.tsx
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders children correctly', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('applies variant styles', () => {
    const { rerender } = render(<Button variant="primary">Test</Button>);
    expect(screen.getByRole('button')).toHaveClass('bg-blue-600');

    rerender(<Button variant="secondary">Test</Button>);
    expect(screen.getByRole('button')).toHaveClass('bg-gray-200');
  });

  it('applies size styles', () => {
    render(<Button size="large">Test</Button>);
    expect(screen.getByRole('button')).toHaveClass('px-6', 'py-3');
  });

  it('shows loading state', () => {
    render(<Button isLoading>Loading</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
    expect(screen.getByRole('button')).toHaveClass('opacity-75');
  });
});
```

### Accessibility Tests

```typescript
// Accessibility tests
import { axe } from 'jest-axe';

describe('Button accessibility', () => {
  it('has no accessibility violations', async () => {
    const { container } = render(<Button>Accessible Button</Button>);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });

  it('supports keyboard navigation', () => {
    render(<Button>Keyboard Button</Button>);
    const button = screen.getByRole('button');

    button.focus();
    expect(button).toHaveFocus();

    fireEvent.keyDown(button, { key: 'Enter' });
    // Test enter key behavior
  });
});
```

## Optimization Strategies

### Code Splitting

```typescript
// Lazy load generated components
const HeavyComponent = lazy(() => import('./generated/HeavyComponent'));

function App() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <HeavyComponent />
    </Suspense>
  );
}
```

### Tree Shaking

```typescript
// Export individual components
export { Button } from './Button';
export { Input } from './Input';
export { Card } from './Card';

// Avoid barrel exports for unused components
```

### Bundle Size

```typescript
// Use dynamic imports for large components
const Chart = dynamic(() => import('./generated/Chart'), {
  loading: () => <div>Loading chart...</div>,
  ssr: false
});
```

## Best Practices

### Component Organization

```
src/
  components/
    generated/           # Auto-generated from .pen
      Button.tsx
      Input.tsx
      Card.tsx
    ui/                  # Custom components
      EnhancedButton.tsx # Wraps generated Button
    index.ts             # Public API
```

### Custom Wrappers

```typescript
// Wrap generated components with custom logic
import { Button as GeneratedButton } from './generated/Button';

export const Button = ({ onClick, ...props }) => {
  const handleClick = () => {
    // Custom analytics
    trackEvent('button_clicked');
    onClick?.();
  };

  return <GeneratedButton onClick={handleClick} {...props} />;
};
```

### Documentation

```typescript
/**
 * Primary Button Component
 *
 * @example
 * ```tsx
 * <Button variant="primary" size="medium" onClick={handleClick}>
 *   Click me
 * </Button>
 * ```
 *
 * @param variant - Visual style variant
 * @param size - Size variant
 * @param isLoading - Show loading state
 * @param children - Button content
 */
export const Button = ({ variant, size, isLoading, children }: ButtonProps) => {
  // ...
};
```

## Common Patterns

### Pattern 1: Form Components

```typescript
// Generate form components with validation
export const FormField = ({ label, error, ...props }) => (
  <div className="space-y-1">
    <label className="block text-sm font-medium text-gray-700">
      {label}
    </label>
    <Input
      className={error ? 'border-red-500' : ''}
      aria-invalid={!!error}
      aria-describedby={error ? `${props.id}-error` : undefined}
      {...props}
    />
    {error && (
      <p id={`${props.id}-error`} className="text-sm text-red-600">
        {error}
      </p>
    )}
  </div>
);
```

### Pattern 2: Data Display

```typescript
// Generate data table components
export const DataTable = ({ columns, data }) => (
  <div className="overflow-x-auto">
    <table className="min-w-full divide-y divide-gray-200">
      <thead className="bg-gray-50">
        <tr>
          {columns.map((col) => (
            <th key={col.key} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              {col.label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody className="bg-white divide-y divide-gray-200">
        {data.map((row, i) => (
          <tr key={i}>
            {columns.map((col) => (
              <td key={col.key} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {row[col.key]}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);
```

### Pattern 3: Responsive Layouts

```typescript
// Generate responsive grid layouts
export const Grid = ({ children, cols = 1 }) => (
  <div className={`
    grid
    gap-6
    grid-cols-1
    md:grid-cols-2
    lg:grid-cols-${cols}
  `}>
    {children}
  </div>
);
```

## Error Handling

### Export Failures

```typescript
// Handle export errors gracefully
try {
  const components = await pencil.export_to_react(penFrame, options);
} catch (error) {
  if (error.code === 'INVALID_DESIGN') {
    console.error('Design file is invalid:', error.details);
  } else if (error.code === 'MISSING_TOKENS') {
    console.error('Design tokens not found:', error.missing);
  }
}
```

### Validation Errors

```typescript
// Validate exported components
const validateComponent = (component) => {
  if (!component.props) {
    throw new Error('Missing props definition');
  }
  if (!component.styles) {
    throw new Error('Missing styles');
  }
  return true;
};
```

## Resources

- React Documentation: https://react.dev
- Tailwind CSS: https://tailwindcss.com
- Pencil Export API: https://docs.pencil.dev/export
- Component Patterns: https://reactpatterns.com

---

Last Updated: 2026-02-09
Tool Version: Pencil Export 1.0.0
