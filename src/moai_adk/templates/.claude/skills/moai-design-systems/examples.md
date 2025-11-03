# Design Systems - Practical Examples

## Example 1: Complete Design Token Setup

### Step 1: Define Core Tokens (core/color.json)

```json
{
  "color": {
    "primary": {
      "50": { "$value": "#EFF6FF", "$type": "color", "$description": "Lightest primary" },
      "100": { "$value": "#DBEAFE", "$type": "color" },
      "200": { "$value": "#BFDBFE", "$type": "color" },
      "300": { "$value": "#93C5FD", "$type": "color" },
      "400": { "$value": "#60A5FA", "$type": "color" },
      "500": { "$value": "#3B82F6", "$type": "color", "$description": "Base primary" },
      "600": { "$value": "#2563EB", "$type": "color" },
      "700": { "$value": "#1D4ED8", "$type": "color" },
      "800": { "$value": "#1E40AF", "$type": "color" },
      "900": { "$value": "#1E3A8A", "$type": "color", "$description": "Darkest primary" }
    },
    "neutral": {
      "50": { "$value": "#F9FAFB", "$type": "color" },
      "100": { "$value": "#F3F4F6", "$type": "color" },
      "200": { "$value": "#E5E7EB", "$type": "color" },
      "300": { "$value": "#D1D5DB", "$type": "color" },
      "400": { "$value": "#9CA3AF", "$type": "color" },
      "500": { "$value": "#6B7280", "$type": "color" },
      "600": { "$value": "#4B5563", "$type": "color" },
      "700": { "$value": "#374151", "$type": "color" },
      "800": { "$value": "#1F2937", "$type": "color" },
      "900": { "$value": "#111827", "$type": "color" }
    },
    "success": {
      "500": { "$value": "#10B981", "$type": "color", "$description": "Success state" }
    },
    "error": {
      "500": { "$value": "#EF4444", "$type": "color", "$description": "Error state" }
    },
    "warning": {
      "500": { "$value": "#F59E0B", "$type": "color", "$description": "Warning state" }
    }
  }
}
```

### Step 2: Define Semantic Tokens (semantic/buttons.json)

```json
{
  "button": {
    "primary": {
      "background": {
        "$value": "{color.primary.500}",
        "$type": "color",
        "$description": "Primary button background"
      },
      "background-hover": {
        "$value": "{color.primary.600}",
        "$type": "color"
      },
      "text": {
        "$value": "#FFFFFF",
        "$type": "color"
      }
    },
    "secondary": {
      "background": {
        "$value": "{color.neutral.100}",
        "$type": "color"
      },
      "background-hover": {
        "$value": "{color.neutral.200}",
        "$type": "color"
      },
      "text": {
        "$value": "{color.neutral.900}",
        "$type": "color"
      }
    }
  }
}
```

### Step 3: Define Typography Tokens (core/typography.json)

```json
{
  "typography": {
    "font-family": {
      "base": { "$value": "Inter, system-ui, sans-serif", "$type": "fontFamily" },
      "heading": { "$value": "Inter, system-ui, sans-serif", "$type": "fontFamily" },
      "mono": { "$value": "JetBrains Mono, monospace", "$type": "fontFamily" }
    },
    "font-size": {
      "xs": { "$value": "12px", "$type": "dimension" },
      "sm": { "$value": "14px", "$type": "dimension" },
      "base": { "$value": "16px", "$type": "dimension" },
      "lg": { "$value": "18px", "$type": "dimension" },
      "xl": { "$value": "20px", "$type": "dimension" },
      "2xl": { "$value": "24px", "$type": "dimension" },
      "3xl": { "$value": "30px", "$type": "dimension" },
      "4xl": { "$value": "36px", "$type": "dimension" }
    },
    "font-weight": {
      "normal": { "$value": "400", "$type": "fontWeight" },
      "medium": { "$value": "500", "$type": "fontWeight" },
      "semibold": { "$value": "600", "$type": "fontWeight" },
      "bold": { "$value": "700", "$type": "fontWeight" }
    },
    "line-height": {
      "tight": { "$value": "1.25", "$type": "number" },
      "normal": { "$value": "1.5", "$type": "number" },
      "relaxed": { "$value": "1.75", "$type": "number" }
    }
  }
}
```

### Step 4: Define Spacing Tokens (core/spacing.json)

```json
{
  "spacing": {
    "0": { "$value": "0", "$type": "dimension" },
    "1": { "$value": "4px", "$type": "dimension" },
    "2": { "$value": "8px", "$type": "dimension" },
    "3": { "$value": "12px", "$type": "dimension" },
    "4": { "$value": "16px", "$type": "dimension" },
    "5": { "$value": "20px", "$type": "dimension" },
    "6": { "$value": "24px", "$type": "dimension" },
    "8": { "$value": "32px", "$type": "dimension" },
    "10": { "$value": "40px", "$type": "dimension" },
    "12": { "$value": "48px", "$type": "dimension" },
    "16": { "$value": "64px", "$type": "dimension" }
  }
}
```

### Step 5: Transform Tokens with Style Dictionary

**style-dictionary.config.json**:
```json
{
  "source": ["tokens/**/*.json"],
  "platforms": {
    "css": {
      "transformGroup": "css",
      "buildPath": "dist/css/",
      "files": [
        {
          "destination": "variables.css",
          "format": "css/variables"
        }
      ]
    },
    "js": {
      "transformGroup": "js",
      "buildPath": "dist/js/",
      "files": [
        {
          "destination": "tokens.js",
          "format": "javascript/es6"
        }
      ]
    },
    "typescript": {
      "transformGroup": "js",
      "buildPath": "dist/ts/",
      "files": [
        {
          "destination": "tokens.ts",
          "format": "typescript/es6-declarations"
        }
      ]
    }
  }
}
```

**Generated CSS (dist/css/variables.css)**:
```css
:root {
  --color-primary-50: #EFF6FF;
  --color-primary-500: #3B82F6;
  --color-primary-900: #1E3A8A;
  --color-neutral-50: #F9FAFB;
  --color-neutral-900: #111827;
  
  --button-primary-background: var(--color-primary-500);
  --button-primary-background-hover: var(--color-primary-600);
  --button-primary-text: #FFFFFF;
  
  --typography-font-family-base: Inter, system-ui, sans-serif;
  --typography-font-size-base: 16px;
  --typography-font-weight-normal: 400;
  
  --spacing-4: 16px;
  --spacing-8: 32px;
}
```

---

## Example 2: Atomic Design Component Library

### Atom: Button Component

**Button.tsx**:
```typescript
import { ButtonHTMLAttributes, ReactNode } from 'react';
import styles from './Button.module.css';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  /**
   * Visual variant
   * @default 'primary'
   */
  variant?: 'primary' | 'secondary' | 'tertiary' | 'ghost';
  
  /**
   * Size of button
   * @default 'medium'
   */
  size?: 'small' | 'medium' | 'large';
  
  /**
   * Loading state with spinner
   * @default false
   */
  loading?: boolean;
  
  /**
   * Icon component
   */
  icon?: ReactNode;
  
  /**
   * Button content
   */
  children: ReactNode;
}

export function Button({
  variant = 'primary',
  size = 'medium',
  loading = false,
  icon,
  disabled,
  children,
  className,
  ...props
}: ButtonProps) {
  return (
    <button
      className={`${styles.button} ${styles[variant]} ${styles[size]} ${className || ''}`}
      disabled={disabled || loading}
      {...props}
    >
      {loading && <span className={styles.spinner} aria-hidden="true" />}
      {icon && <span className={styles.icon}>{icon}</span>}
      <span>{children}</span>
    </button>
  );
}
```

**Button.module.css**:
```css
.button {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  font-family: var(--typography-font-family-base);
  font-weight: var(--typography-font-weight-medium);
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.button:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

/* Variants */
.primary {
  background: var(--button-primary-background);
  color: var(--button-primary-text);
}

.primary:hover:not(:disabled) {
  background: var(--button-primary-background-hover);
}

.secondary {
  background: var(--button-secondary-background);
  color: var(--button-secondary-text);
}

/* Sizes */
.small {
  padding: var(--spacing-1) var(--spacing-3);
  font-size: var(--typography-font-size-sm);
}

.medium {
  padding: var(--spacing-2) var(--spacing-4);
  font-size: var(--typography-font-size-base);
}

.large {
  padding: var(--spacing-3) var(--spacing-6);
  font-size: var(--typography-font-size-lg);
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
```

### Molecule: FormField Component

**FormField.tsx**:
```typescript
import { ReactElement } from 'react';
import styles from './FormField.module.css';

export interface FormFieldProps {
  /**
   * Field label
   */
  label: string;
  
  /**
   * Input component (atom)
   */
  input: ReactElement;
  
  /**
   * Error message
   */
  error?: string;
  
  /**
   * Helper text
   */
  helperText?: string;
  
  /**
   * Required field
   * @default false
   */
  required?: boolean;
  
  /**
   * Field ID (for label association)
   */
  id: string;
}

export function FormField({
  label,
  input,
  error,
  helperText,
  required,
  id,
}: FormFieldProps) {
  const describedBy = [];
  if (error) describedBy.push(`${id}-error`);
  if (helperText) describedBy.push(`${id}-helper`);
  
  return (
    <div className={styles.formField}>
      <label htmlFor={id} className={styles.label}>
        {label}
        {required && <span className={styles.required} aria-label="required">*</span>}
      </label>
      
      {/* Clone input with accessibility props */}
      {React.cloneElement(input, {
        id,
        'aria-invalid': !!error,
        'aria-describedby': describedBy.length > 0 ? describedBy.join(' ') : undefined,
      })}
      
      {helperText && (
        <p id={`${id}-helper`} className={styles.helperText}>
          {helperText}
        </p>
      )}
      
      {error && (
        <p id={`${id}-error`} className={styles.error} role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
```

### Organism: Header Component

**Header.tsx**:
```typescript
import { ReactNode } from 'react';
import styles from './Header.module.css';

export interface NavigationItem {
  label: string;
  href: string;
  current?: boolean;
}

export interface HeaderProps {
  /**
   * Logo component
   */
  logo: ReactNode;
  
  /**
   * Navigation items
   */
  navigation: NavigationItem[];
  
  /**
   * Search box component (optional)
   */
  searchBox?: ReactNode;
  
  /**
   * User menu component (optional)
   */
  userMenu?: ReactNode;
  
  /**
   * Navigation handler
   */
  onNavigate: (path: string) => void;
}

export function Header({
  logo,
  navigation,
  searchBox,
  userMenu,
  onNavigate,
}: HeaderProps) {
  return (
    <header className={styles.header}>
      <div className={styles.container}>
        {/* Logo */}
        <div className={styles.logo}>{logo}</div>
        
        {/* Navigation */}
        <nav className={styles.nav} aria-label="Main navigation">
          <ul className={styles.navList}>
            {navigation.map((item) => (
              <li key={item.href}>
                <a
                  href={item.href}
                  onClick={(e) => {
                    e.preventDefault();
                    onNavigate(item.href);
                  }}
                  aria-current={item.current ? 'page' : undefined}
                  className={item.current ? styles.navLinkActive : styles.navLink}
                >
                  {item.label}
                </a>
              </li>
            ))}
          </ul>
        </nav>
        
        {/* Search */}
        {searchBox && <div className={styles.search}>{searchBox}</div>}
        
        {/* User Menu */}
        {userMenu && <div className={styles.userMenu}>{userMenu}</div>}
      </div>
    </header>
  );
}
```

---

## Example 3: WCAG 2.1 AA Compliance

### Accessible Modal Dialog

**Modal.tsx**:
```typescript
import { ReactNode, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import FocusTrap from 'focus-trap-react';
import styles from './Modal.module.css';

export interface ModalProps {
  /**
   * Modal open state
   */
  isOpen: boolean;
  
  /**
   * Close handler
   */
  onClose: () => void;
  
  /**
   * Modal title
   */
  title: string;
  
  /**
   * Modal content
   */
  children: ReactNode;
  
  /**
   * Footer actions (buttons)
   */
  footer?: ReactNode;
}

export function Modal({ isOpen, onClose, title, children, footer }: ModalProps) {
  const titleId = useRef(`modal-title-${Math.random().toString(36).slice(2)}`);
  const previousFocus = useRef<HTMLElement | null>(null);
  
  useEffect(() => {
    if (!isOpen) return;
    
    // Store currently focused element
    previousFocus.current = document.activeElement as HTMLElement;
    
    // Prevent body scroll
    document.body.style.overflow = 'hidden';
    
    // Handle Escape key
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    
    document.addEventListener('keydown', handleEscape);
    
    return () => {
      // Restore body scroll
      document.body.style.overflow = '';
      
      // Restore focus to previous element
      previousFocus.current?.focus();
      
      document.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen, onClose]);
  
  if (!isOpen) return null;
  
  return createPortal(
    <div className={styles.overlay} onClick={onClose} aria-hidden="true">
      <FocusTrap>
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby={titleId.current}
          className={styles.modal}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className={styles.header}>
            <h2 id={titleId.current} className={styles.title}>
              {title}
            </h2>
            <button
              onClick={onClose}
              aria-label="Close modal"
              className={styles.closeButton}
            >
              <CloseIcon />
            </button>
          </div>
          
          {/* Body */}
          <div className={styles.body}>{children}</div>
          
          {/* Footer */}
          {footer && <div className={styles.footer}>{footer}</div>}
        </div>
      </FocusTrap>
    </div>,
    document.body
  );
}
```

### Accessible Form with Validation

**RegistrationForm.tsx**:
```typescript
import { useState } from 'react';
import { FormField } from './FormField';
import { Input } from './Input';
import { Button } from './Button';

export function RegistrationForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  
  const validate = () => {
    const newErrors: Record<string, string> = {};
    
    if (!email) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Email is invalid';
    }
    
    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validate()) {
      // Focus first error field
      const firstErrorField = document.querySelector('[aria-invalid="true"]') as HTMLElement;
      firstErrorField?.focus();
      return;
    }
    
    // Submit form
    console.log('Form submitted:', { email, password });
  };
  
  return (
    <form onSubmit={handleSubmit} noValidate>
      <FormField
        id="email"
        label="Email"
        required
        input={
          <Input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
          />
        }
        error={errors.email}
        helperText="We'll never share your email"
      />
      
      <FormField
        id="password"
        label="Password"
        required
        input={
          <Input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="new-password"
          />
        }
        error={errors.password}
        helperText="Minimum 8 characters"
      />
      
      <Button type="submit" variant="primary">
        Create Account
      </Button>
    </form>
  );
}
```

---

## Example 4: Figma MCP Integration Workflow

### Step 1: Extract Design Tokens from Figma

**Prompt to Claude Code (with Figma MCP active)**:
```
"Extract all color tokens from the 'Design System / Colors' page
in file 'abc123' and generate a JSON file following W3C DTCG spec.
Include semantic naming (primary, neutral, success, error, warning)."
```

**MCP Response** (generated tokens.json):
```json
{
  "color": {
    "primary": {
      "500": {
        "$value": "#3B82F6",
        "$type": "color",
        "$description": "Primary brand color"
      }
    },
    "neutral": {
      "50": { "$value": "#F9FAFB", "$type": "color" },
      "900": { "$value": "#111827", "$type": "color" }
    },
    "success": {
      "500": { "$value": "#10B981", "$type": "color" }
    }
  }
}
```

### Step 2: Generate Component Spec from Figma Frame

**Prompt**:
```
"Generate TypeScript interface for the Button component
in frame 'Components/Button' with all variants (primary, secondary)
and sizes (small, medium, large)."
```

**MCP Response** (ButtonProps.ts):
```typescript
export interface ButtonProps {
  variant: 'primary' | 'secondary' | 'tertiary';
  size: 'small' | 'medium' | 'large';
  disabled?: boolean;
  children: React.ReactNode;
}

// Extracted dimensions from Figma
export const ButtonSizes = {
  small: { height: 32, paddingX: 12, fontSize: 14 },
  medium: { height: 40, paddingX: 16, fontSize: 16 },
  large: { height: 48, paddingX: 24, fontSize: 18 },
};
```

### Step 3: Automate Token Sync in CI/CD

**GitHub Actions workflow** (.github/workflows/sync-tokens.yml):
```yaml
name: Sync Design Tokens

on:
  workflow_dispatch:  # Manual trigger
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      
      - name: Install dependencies
        run: npm ci
      
      - name: Extract tokens from Figma
        env:
          FIGMA_ACCESS_TOKEN: ${{ secrets.FIGMA_ACCESS_TOKEN }}
        run: |
          npx figma-tokens-cli extract \
            --file-key abc123 \
            --output tokens/figma.json
      
      - name: Transform tokens with Style Dictionary
        run: npx style-dictionary build
      
      - name: Commit changes
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add tokens/ dist/
          git commit -m "chore: sync design tokens from Figma" || exit 0
          git push
```

---

## Example 5: Storybook Documentation Setup

### Component Story with Controls

**Button.stories.tsx**:
```typescript
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';
import { PlusIcon } from './icons';

const meta: Meta<typeof Button> = {
  title: 'Components/Button',
  component: Button,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
Primary UI component for user interaction. Supports multiple variants,
sizes, loading states, and icons. All buttons meet WCAG 2.1 AA contrast
requirements (4.5:1 minimum) and are fully keyboard accessible.
        `.trim(),
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'tertiary', 'ghost'],
      description: 'Visual variant of the button',
      table: {
        defaultValue: { summary: 'primary' },
      },
    },
    size: {
      control: 'select',
      options: ['small', 'medium', 'large'],
    },
    disabled: {
      control: 'boolean',
    },
    loading: {
      control: 'boolean',
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    variant: 'primary',
    children: 'Button',
  },
};

export const Secondary: Story = {
  args: {
    variant: 'secondary',
    children: 'Button',
  },
};

export const WithIcon: Story = {
  args: {
    variant: 'primary',
    icon: <PlusIcon />,
    children: 'Add Item',
  },
};

export const Loading: Story = {
  args: {
    variant: 'primary',
    loading: true,
    children: 'Loading...',
  },
};

export const Disabled: Story = {
  args: {
    variant: 'primary',
    disabled: true,
    children: 'Disabled',
  },
};

export const AllSizes: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
      <Button size="small">Small</Button>
      <Button size="medium">Medium</Button>
      <Button size="large">Large</Button>
    </div>
  ),
};
```

### MDX Documentation with Live Examples

**Button.mdx**:
```mdx
import { Canvas, Meta, Story } from '@storybook/blocks';
import * as ButtonStories from './Button.stories';

<Meta of={ButtonStories} />

# Button

Primary UI component for user interaction.

## Usage

Import the Button component:

```tsx
import { Button } from '@company/design-system';

function App() {
  return (
    <Button variant="primary" onClick={() => console.log('clicked')}>
      Click me
    </Button>
  );
}
```

## Variants

The Button component supports four visual variants:

<Canvas of={ButtonStories.Primary} />
<Canvas of={ButtonStories.Secondary} />

## Sizes

Three sizes are available: small, medium (default), and large.

<Canvas of={ButtonStories.AllSizes} />

## Loading State

Use the `loading` prop to display a spinner:

<Canvas of={ButtonStories.Loading} />

## Accessibility

- **Keyboard Navigation**: All buttons are focusable and activatable via Enter/Space keys
- **Contrast Ratio**: 4.5:1 minimum (WCAG 2.1 AA)
- **Focus Indicator**: 2px outline with 2px offset
- **ARIA**: Use `aria-label` for icon-only buttons

### Example: Icon-only Button

```tsx
<Button variant="ghost" icon={<CloseIcon />} aria-label="Close modal">
  {/* No visible text */}
</Button>
```

## API

See the interactive controls below to explore all props.
```

---

## Example 6: Testing Strategy

### Accessibility Test Suite

**Button.a11y.test.tsx**:
```typescript
import { render, screen } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import userEvent from '@testing-library/user-event';
import { Button } from './Button';

expect.extend(toHaveNoViolations);

describe('Button Accessibility', () => {
  it('should not have accessibility violations', async () => {
    const { container } = render(
      <Button variant="primary" onClick={() => {}}>
        Click me
      </Button>
    );
    
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
  
  it('should be keyboard accessible', async () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    const button = screen.getByRole('button');
    button.focus();
    
    expect(button).toHaveFocus();
    
    // Activate with Enter
    await userEvent.keyboard('{Enter}');
    expect(handleClick).toHaveBeenCalledTimes(1);
    
    // Activate with Space
    await userEvent.keyboard(' ');
    expect(handleClick).toHaveBeenCalledTimes(2);
  });
  
  it('should have sufficient color contrast', () => {
    const { container } = render(<Button variant="primary">Text</Button>);
    
    const button = container.querySelector('button');
    const styles = window.getComputedStyle(button!);
    
    // Get computed colors
    const bgColor = styles.backgroundColor;
    const textColor = styles.color;
    
    // Calculate contrast ratio (using external library or manual calculation)
    const contrastRatio = calculateContrast(bgColor, textColor);
    
    // WCAG 2.1 AA requires 4.5:1 for normal text
    expect(contrastRatio).toBeGreaterThanOrEqual(4.5);
  });
  
  it('should have accessible name for icon-only buttons', () => {
    render(
      <Button variant="ghost" icon={<CloseIcon />} aria-label="Close">
        {/* No visible text */}
      </Button>
    );
    
    const button = screen.getByRole('button', { name: 'Close' });
    expect(button).toBeInTheDocument();
  });
});
```

### Visual Regression Test

**Button.visual.test.ts** (Playwright):
```typescript
import { test, expect } from '@playwright/test';

test.describe('Button Visual Regression', () => {
  test('should match screenshot for all variants', async ({ page }) => {
    await page.goto('http://localhost:6006/iframe.html?id=components-button--all-variants');
    
    // Wait for fonts to load
    await page.waitForLoadState('networkidle');
    
    // Take screenshot
    await expect(page).toHaveScreenshot('button-variants.png', {
      maxDiffPixels: 100,
    });
  });
  
  test('should match screenshot for all sizes', async ({ page }) => {
    await page.goto('http://localhost:6006/iframe.html?id=components-button--all-sizes');
    await page.waitForLoadState('networkidle');
    
    await expect(page).toHaveScreenshot('button-sizes.png');
  });
});
```

---

**Examples Version**: 1.0.0 (2025-11-04)
