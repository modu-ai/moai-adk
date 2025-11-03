# Design Systems: Practical Examples

Real-world code examples for implementing design systems with DTCG 2025.10 tokens, WCAG 2.2 accessibility, and Figma MCP workflows.

---

## Example 1: Complete Design Token Setup (DTCG 2025.10)

### Color Tokens

**File**: `tokens/color.json`

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "color": {
      "$type": "color",
      "gray": {
        "50": { "$value": "#f9fafb" },
        "100": { "$value": "#f3f4f6" },
        "200": { "$value": "#e5e7eb" },
        "300": { "$value": "#d1d5db" },
        "400": { "$value": "#9ca3af" },
        "500": { "$value": "#6b7280" },
        "600": { "$value": "#4b5563" },
        "700": { "$value": "#374151" },
        "800": { "$value": "#1f2937" },
        "900": { "$value": "#111827" }
      },
      "primary": {
        "50": { "$value": "#eff6ff" },
        "100": { "$value": "#dbeafe" },
        "200": { "$value": "#bfdbfe" },
        "300": { "$value": "#93c5fd" },
        "400": { "$value": "#60a5fa" },
        "500": { "$value": "#3b82f6" },
        "600": { "$value": "#2563eb" },
        "700": { "$value": "#1d4ed8" },
        "800": { "$value": "#1e40af" },
        "900": { "$value": "#1e3a8a" }
      },
      "semantic": {
        "text": {
          "primary": {
            "$value": "{color.gray.900}",
            "$description": "Main body text color",
            "$extensions": {
              "mode": {
                "dark": "{color.gray.50}"
              }
            }
          },
          "secondary": {
            "$value": "{color.gray.600}",
            "$description": "Supporting text color",
            "$extensions": {
              "mode": {
                "dark": "{color.gray.400}"
              }
            }
          },
          "disabled": {
            "$value": "{color.gray.400}",
            "$description": "Disabled text color",
            "$extensions": {
              "mode": {
                "dark": "{color.gray.600}"
              }
            }
          }
        },
        "background": {
          "default": {
            "$value": "#ffffff",
            "$description": "Default page background",
            "$extensions": {
              "mode": {
                "dark": "{color.gray.900}"
              }
            }
          },
          "elevated": {
            "$value": "{color.gray.50}",
            "$description": "Card and elevated surfaces",
            "$extensions": {
              "mode": {
                "dark": "{color.gray.800}"
              }
            }
          }
        },
        "action": {
          "primary": {
            "$value": "{color.primary.500}",
            "$description": "Primary action color (buttons, links)"
          },
          "primaryHover": {
            "$value": "{color.primary.600}",
            "$description": "Primary action hover state"
          }
        }
      }
    }
  }
}
```

### Typography Tokens

**File**: `tokens/typography.json`

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "font": {
      "family": {
        "$type": "fontFamily",
        "sans": {
          "$value": ["Inter", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "Roboto", "sans-serif"],
          "$description": "Primary sans-serif font stack"
        },
        "mono": {
          "$value": ["JetBrains Mono", "Menlo", "Monaco", "Courier New", "monospace"],
          "$description": "Monospace font stack for code"
        }
      },
      "size": {
        "$type": "dimension",
        "xs": { "$value": "0.75rem", "$description": "12px" },
        "sm": { "$value": "0.875rem", "$description": "14px" },
        "base": { "$value": "1rem", "$description": "16px" },
        "lg": { "$value": "1.125rem", "$description": "18px" },
        "xl": { "$value": "1.25rem", "$description": "20px" },
        "2xl": { "$value": "1.5rem", "$description": "24px" },
        "3xl": { "$value": "1.875rem", "$description": "30px" },
        "4xl": { "$value": "2.25rem", "$description": "36px" }
      },
      "weight": {
        "$type": "fontWeight",
        "normal": { "$value": "400" },
        "medium": { "$value": "500" },
        "semibold": { "$value": "600" },
        "bold": { "$value": "700" }
      },
      "lineHeight": {
        "$type": "number",
        "tight": { "$value": "1.25" },
        "normal": { "$value": "1.5" },
        "relaxed": { "$value": "1.75" }
      }
    }
  }
}
```

### Spacing Tokens

**File**: `tokens/spacing.json`

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "spacing": {
      "$type": "dimension",
      "0": { "$value": "0" },
      "1": { "$value": "0.25rem", "$description": "4px" },
      "2": { "$value": "0.5rem", "$description": "8px" },
      "3": { "$value": "0.75rem", "$description": "12px" },
      "4": { "$value": "1rem", "$description": "16px" },
      "5": { "$value": "1.25rem", "$description": "20px" },
      "6": { "$value": "1.5rem", "$description": "24px" },
      "8": { "$value": "2rem", "$description": "32px" },
      "10": { "$value": "2.5rem", "$description": "40px" },
      "12": { "$value": "3rem", "$description": "48px" },
      "16": { "$value": "4rem", "$description": "64px" }
    }
  }
}
```

### Style Dictionary Build Configuration

**File**: `style-dictionary.config.js`

```javascript
export default {
  source: ['tokens/**/*.json'],
  
  platforms: {
    // CSS Variables
    css: {
      transformGroup: 'css',
      buildPath: 'build/css/',
      files: [{
        destination: 'variables.css',
        format: 'css/variables',
        options: {
          outputReferences: true
        }
      }]
    },
    
    // JavaScript/TypeScript
    js: {
      transformGroup: 'js',
      buildPath: 'build/js/',
      files: [{
        destination: 'tokens.js',
        format: 'javascript/es6'
      }, {
        destination: 'tokens.d.ts',
        format: 'typescript/es6-declarations'
      }]
    },
    
    // Tailwind CSS
    tailwind: {
      transformGroup: 'js',
      buildPath: 'build/tailwind/',
      files: [{
        destination: 'tokens.js',
        format: 'javascript/module-flat'
      }]
    }
  }
};
```

**Generated Output** (`build/css/variables.css`):

```css
:root {
  /* Colors */
  --color-gray-50: #f9fafb;
  --color-gray-900: #111827;
  --color-primary-500: #3b82f6;
  --color-semantic-text-primary: var(--color-gray-900);
  --color-semantic-action-primary: var(--color-primary-500);
  
  /* Typography */
  --font-family-sans: Inter, -apple-system, BlinkMacSystemFont, sans-serif;
  --font-size-base: 1rem;
  --font-weight-bold: 700;
  --font-line-height-normal: 1.5;
  
  /* Spacing */
  --spacing-4: 1rem;
  --spacing-8: 2rem;
}

[data-theme="dark"] {
  --color-semantic-text-primary: var(--color-gray-50);
  --color-semantic-background-default: var(--color-gray-900);
}
```

---

## Example 2: Atomic Design Folder Structure

```
src/design-system/
├── tokens/                           # DTCG 2025.10 JSON tokens
│   ├── color.json
│   ├── typography.json
│   ├── spacing.json
│   ├── border.json
│   └── shadow.json
│
├── components/
│   ├── atoms/                        # Basic building blocks
│   │   ├── Button/
│   │   │   ├── Button.tsx
│   │   │   ├── Button.stories.tsx
│   │   │   ├── Button.test.tsx
│   │   │   ├── Button.module.css
│   │   │   └── index.ts
│   │   ├── Input/
│   │   ├── Label/
│   │   ├── Icon/
│   │   ├── Spinner/
│   │   └── Badge/
│   │
│   ├── molecules/                    # Simple component combinations
│   │   ├── FormField/
│   │   │   ├── FormField.tsx
│   │   │   ├── FormField.stories.tsx
│   │   │   ├── FormField.test.tsx
│   │   │   └── index.ts
│   │   ├── SearchBar/
│   │   ├── Card/
│   │   ├── Alert/
│   │   └── Dropdown/
│   │
│   ├── organisms/                    # Complex component sections
│   │   ├── Header/
│   │   ├── Footer/
│   │   ├── DataTable/
│   │   ├── Modal/
│   │   └── NavigationMenu/
│   │
│   └── templates/                    # Page-level layouts
│       ├── DashboardLayout/
│       ├── AuthLayout/
│       └── LandingPageLayout/
│
├── hooks/                            # Shared React hooks
│   ├── useKeyboardNavigation.ts
│   ├── useFocusTrap.ts
│   └── useReducedMotion.ts
│
├── utils/                            # Utility functions
│   ├── a11y/
│   │   ├── contrast.ts
│   │   └── ariaUtils.ts
│   └── tokens/
│       └── tokenHelpers.ts
│
├── styles/                           # Global styles
│   ├── global.css
│   ├── reset.css
│   └── theme.css
│
└── index.ts                          # Public API exports
```

**Public API** (`src/design-system/index.ts`):

```typescript
// Atoms
export { Button } from './components/atoms/Button';
export { Input } from './components/atoms/Input';
export { Icon } from './components/atoms/Icon';

// Molecules
export { FormField } from './components/molecules/FormField';
export { Card } from './components/molecules/Card';

// Organisms
export { Header } from './components/organisms/Header';
export { DataTable } from './components/organisms/DataTable';

// Hooks
export { useKeyboardNavigation } from './hooks/useKeyboardNavigation';
export { useFocusTrap } from './hooks/useFocusTrap';

// Utils
export { getContrastRatio, meetsWCAG } from './utils/a11y/contrast';
```

---

## Example 3: WCAG 2.2 AA Compliance Checklist

### Color Contrast Validation

**Automated Test** (`utils/a11y/contrast.test.ts`):

```typescript
import { getContrastRatio, meetsWCAG } from './contrast';

describe('Color Contrast WCAG 2.2 Compliance', () => {
  describe('Level AA Requirements', () => {
    it('should pass AA for normal text (4.5:1)', () => {
      // Gray 600 on white background
      expect(meetsWCAG('#4b5563', '#ffffff', 'AA', false)).toBe(true);
      expect(getContrastRatio('#4b5563', '#ffffff')).toBeGreaterThanOrEqual(4.5);
    });

    it('should pass AA for large text (3:1)', () => {
      // Gray 400 on white background (18pt+)
      expect(meetsWCAG('#9ca3af', '#ffffff', 'AA', true)).toBe(true);
      expect(getContrastRatio('#9ca3af', '#ffffff')).toBeGreaterThanOrEqual(3);
    });

    it('should fail AA for insufficient contrast', () => {
      // Light gray on white (fails)
      expect(meetsWCAG('#e5e7eb', '#ffffff', 'AA', false)).toBe(false);
    });

    it('should pass AA for UI components (3:1)', () => {
      // Gray 300 border on white background
      expect(getContrastRatio('#d1d5db', '#ffffff')).toBeGreaterThanOrEqual(3);
    });
  });

  describe('Level AAA Requirements', () => {
    it('should pass AAA for normal text (7:1)', () => {
      // Gray 900 on white background
      expect(meetsWCAG('#111827', '#ffffff', 'AAA', false)).toBe(true);
      expect(getContrastRatio('#111827', '#ffffff')).toBeGreaterThanOrEqual(7);
    });

    it('should pass AAA for large text (4.5:1)', () => {
      // Gray 600 on white background (18pt+)
      expect(meetsWCAG('#4b5563', '#ffffff', 'AAA', true)).toBe(true);
      expect(getContrastRatio('#4b5563', '#ffffff')).toBeGreaterThanOrEqual(4.5);
    });
  });
});
```

**Token Validation Script** (`scripts/validate-contrast.ts`):

```typescript
import { readFileSync } from 'fs';
import { getContrastRatio } from '../utils/a11y/contrast';

interface ColorToken {
  $value: string;
}

const colorTokens = JSON.parse(readFileSync('tokens/color.json', 'utf-8'));
const semantic = colorTokens.$tokens.color.semantic;

console.log('WCAG 2.2 AA Contrast Validation\n');

// Validate text colors
const textPrimary = semantic.text.primary.$value;
const bgDefault = semantic.background.default.$value;
const ratio = getContrastRatio(textPrimary, bgDefault);

console.log(`Text Primary on Default Background: ${ratio.toFixed(2)}:1`);
console.log(`✅ WCAG AA (4.5:1): ${ratio >= 4.5 ? 'PASS' : 'FAIL'}`);
console.log(`✅ WCAG AAA (7:1): ${ratio >= 7 ? 'PASS' : 'FAIL'}\n`);

// Validate all semantic text colors
const textColors = ['primary', 'secondary', 'disabled'];
const backgrounds = ['default', 'elevated'];

textColors.forEach(textType => {
  backgrounds.forEach(bgType => {
    const textColor = semantic.text[textType].$value;
    const bgColor = semantic.background[bgType].$value;
    const ratio = getContrastRatio(textColor, bgColor);
    
    const status = ratio >= 4.5 ? '✅' : '❌';
    console.log(`${status} ${textType} on ${bgType}: ${ratio.toFixed(2)}:1`);
  });
});
```

### Keyboard Navigation Implementation

**Hook** (`hooks/useKeyboardNavigation.ts`):

```typescript
import { useEffect, useRef } from 'react';

interface KeyboardNavigationOptions {
  onEscape?: () => void;
  onEnter?: () => void;
  trapFocus?: boolean;
  enableArrowKeys?: boolean;
}

export function useKeyboardNavigation<T extends HTMLElement>(
  options: KeyboardNavigationOptions = {}
) {
  const elementRef = useRef<T>(null);

  useEffect(() => {
    const element = elementRef.current;
    if (!element) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Escape key handling
      if (e.key === 'Escape' && options.onEscape) {
        options.onEscape();
        return;
      }

      // Enter key handling
      if (e.key === 'Enter' && options.onEnter) {
        options.onEnter();
        return;
      }

      // Focus trap (Tab key)
      if (e.key === 'Tab' && options.trapFocus) {
        const focusableElements = element.querySelectorAll<HTMLElement>(
          'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
        );

        if (focusableElements.length === 0) return;

        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];

        if (e.shiftKey) {
          // Shift+Tab: going backwards
          if (document.activeElement === firstElement) {
            lastElement.focus();
            e.preventDefault();
          }
        } else {
          // Tab: going forwards
          if (document.activeElement === lastElement) {
            firstElement.focus();
            e.preventDefault();
          }
        }
      }

      // Arrow key navigation
      if (options.enableArrowKeys && ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(e.key)) {
        const focusableElements = Array.from(
          element.querySelectorAll<HTMLElement>(
            '[role="menuitem"], [role="option"], [role="tab"]'
          )
        );

        if (focusableElements.length === 0) return;

        const currentIndex = focusableElements.indexOf(document.activeElement as HTMLElement);
        let nextIndex = currentIndex;

        if (e.key === 'ArrowDown' || e.key === 'ArrowRight') {
          nextIndex = (currentIndex + 1) % focusableElements.length;
        } else if (e.key === 'ArrowUp' || e.key === 'ArrowLeft') {
          nextIndex = (currentIndex - 1 + focusableElements.length) % focusableElements.length;
        }

        focusableElements[nextIndex].focus();
        e.preventDefault();
      }
    };

    element.addEventListener('keydown', handleKeyDown);
    return () => element.removeEventListener('keydown', handleKeyDown);
  }, [options]);

  return elementRef;
}
```

**Usage Example** (Modal component):

```typescript
import { useKeyboardNavigation } from '../../hooks/useKeyboardNavigation';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
}

export function Modal({ isOpen, onClose, children }: ModalProps) {
  const modalRef = useKeyboardNavigation<HTMLDivElement>({
    onEscape: onClose,
    trapFocus: true
  });

  useEffect(() => {
    if (isOpen && modalRef.current) {
      const firstFocusable = modalRef.current.querySelector<HTMLElement>(
        'button, [href], input'
      );
      firstFocusable?.focus();
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      ref={modalRef}
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      className="modal-overlay"
    >
      <div className="modal-content">
        {children}
      </div>
    </div>
  );
}
```

### ARIA Implementation Examples

**FormField with Complete ARIA** (`components/molecules/FormField/FormField.tsx`):

```typescript
import { useId, forwardRef } from 'react';
import { Input } from '../../atoms/Input';
import { Label } from '../../atoms/Label';

interface FormFieldProps {
  label: string;
  error?: string;
  hint?: string;
  required?: boolean;
  disabled?: boolean;
  type?: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

export const FormField = forwardRef<HTMLInputElement, FormFieldProps>(
  ({ label, error, hint, required, disabled, ...inputProps }, ref) => {
    const inputId = useId();
    const errorId = `${inputId}-error`;
    const hintId = `${inputId}-hint`;

    return (
      <div className="form-field">
        <Label htmlFor={inputId}>
          {label}
          {required && (
            <span className="required-indicator" aria-label="required field">
              *
            </span>
          )}
        </Label>

        {hint && !error && (
          <p id={hintId} className="hint-text">
            {hint}
          </p>
        )}

        <Input
          ref={ref}
          id={inputId}
          aria-invalid={!!error}
          aria-describedby={
            error ? errorId : hint ? hintId : undefined
          }
          aria-required={required}
          disabled={disabled}
          {...inputProps}
        />

        {error && (
          <p id={errorId} role="alert" className="error-text">
            {error}
          </p>
        )}
      </div>
    );
  }
);

FormField.displayName = 'FormField';
```

---

## Example 4: Figma → React Component Export Workflow

### Step 1: Figma File Structure

**Best Practices**:
- Name layers semantically: `PrimaryButton`, `HeaderNavigation`, `ProductCard`
- Use Figma Variables for colors, spacing, typography
- Apply Auto Layout for responsive components
- Add component descriptions for AI context

**Figma Variables Example**:

```
Colors/
  Primary/
    primary-50: #eff6ff
    primary-500: #3b82f6
    primary-600: #2563eb
  
  Semantic/
    text-primary: {gray-900}
    action-primary: {primary-500}

Spacing/
  spacing-2: 8px
  spacing-4: 16px
  spacing-6: 24px

Typography/
  font-size-base: 16px
  font-weight-bold: 700
```

### Step 2: Extract Design Tokens via MCP

**Prompt to Claude**:

```
Extract all design tokens from this Figma file:
https://www.figma.com/file/XYZ123/DesignSystem

Output as DTCG 2025.10 JSON format.
```

**MCP Response** (Generated JSON):

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "color": {
      "primary": {
        "$type": "color",
        "500": { "$value": "#3b82f6" }
      }
    },
    "spacing": {
      "$type": "dimension",
      "4": { "$value": "1rem" }
    }
  }
}
```

### Step 3: Generate Component Code

**Prompt to Claude**:

```
Generate a React TypeScript component from this Figma frame:
https://www.figma.com/file/XYZ123/DesignSystem?node-id=123:456

Requirements:
- Use design tokens from tokens/color.json
- WCAG 2.2 AA compliant
- Include Storybook story
- Add accessibility tests
```

**Generated Component** (`components/atoms/Button/Button.tsx`):

```typescript
import { forwardRef } from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import styles from './Button.module.css';

const buttonVariants = cva(styles.button, {
  variants: {
    variant: {
      primary: styles.primary,
      secondary: styles.secondary,
      outline: styles.outline
    },
    size: {
      sm: styles.sm,
      md: styles.md,
      lg: styles.lg
    }
  },
  defaultVariants: {
    variant: 'primary',
    size: 'md'
  }
});

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
        {isLoading ? (
          <>
            <span className={styles.spinner} aria-hidden="true" />
            <span className="sr-only">Loading...</span>
          </>
        ) : (
          children
        )}
      </button>
    );
  }
);

Button.displayName = 'Button';
```

**Generated Styles** (`Button.module.css`):

```css
.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  font-weight: 500;
  transition: background-color 0.2s, color 0.2s;
  cursor: pointer;
  border: none;
}

.button:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

.button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.primary {
  background-color: var(--color-primary-500);
  color: #ffffff;
}

.primary:hover:not(:disabled) {
  background-color: var(--color-primary-600);
}

.secondary {
  background-color: var(--color-gray-200);
  color: var(--color-gray-900);
}

.secondary:hover:not(:disabled) {
  background-color: var(--color-gray-300);
}

.sm {
  height: 2rem;
  padding: 0 0.75rem;
  font-size: 0.875rem;
}

.md {
  height: 2.5rem;
  padding: 0 1rem;
  font-size: 1rem;
}

.lg {
  height: 3rem;
  padding: 0 1.5rem;
  font-size: 1.125rem;
}

@media (prefers-reduced-motion: reduce) {
  .button {
    transition: none;
  }
}
```

---

## Example 5: Storybook Configuration & Stories

### Storybook Main Configuration

**File**: `.storybook/main.ts`

```typescript
import type { StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
  stories: [
    '../src/**/*.mdx',
    '../src/**/*.stories.@(js|jsx|mjs|ts|tsx)'
  ],
  
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions',
    '@storybook/addon-a11y', // Accessibility testing
    '@chromatic-com/storybook' // Visual regression
  ],
  
  framework: {
    name: '@storybook/react-vite',
    options: {}
  },
  
  docs: {
    autodocs: 'tag'
  },
  
  viteFinal: async (config) => {
    // Add custom Vite config here
    return config;
  }
};

export default config;
```

### Theme Configuration

**File**: `.storybook/preview.ts`

```typescript
import type { Preview } from '@storybook/react';
import '../src/design-system/styles/global.css';

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: '^on[A-Z].*' },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i
      }
    },
    // Accessibility addon configuration
    a11y: {
      config: {
        rules: [
          {
            id: 'color-contrast',
            enabled: true
          },
          {
            id: 'label',
            enabled: true
          }
        ]
      }
    }
  },
  
  // Global decorators
  decorators: [
    (Story) => (
      <div style={{ padding: '2rem' }}>
        <Story />
      </div>
    )
  ],
  
  // Theme switching
  globalTypes: {
    theme: {
      description: 'Global theme for components',
      defaultValue: 'light',
      toolbar: {
        title: 'Theme',
        icon: 'circlehollow',
        items: ['light', 'dark'],
        dynamicTitle: true
      }
    }
  }
};

export default preview;
```

### Complete Component Story

**File**: `components/atoms/Button/Button.stories.tsx`

```typescript
import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { Button } from './Button';

const meta: Meta<typeof Button> = {
  title: 'Design System/Atoms/Button',
  component: Button,
  tags: ['autodocs'],
  
  parameters: {
    layout: 'centered',
    a11y: {
      // Component-specific a11y config
      config: {
        rules: [{ id: 'button-name', enabled: true }]
      }
    }
  },
  
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'outline'],
      description: 'Visual style variant',
      table: {
        type: { summary: 'string' },
        defaultValue: { summary: 'primary' }
      }
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Button size',
      table: {
        type: { summary: 'string' },
        defaultValue: { summary: 'md' }
      }
    },
    isLoading: {
      control: 'boolean',
      description: 'Loading state indicator'
    },
    disabled: {
      control: 'boolean',
      description: 'Disabled state'
    }
  },
  
  args: {
    onClick: fn()
  }
};

export default meta;
type Story = StoryObj<typeof Button>;

// Default story
export const Primary: Story = {
  args: {
    children: 'Primary Button',
    variant: 'primary'
  }
};

export const Secondary: Story = {
  args: {
    children: 'Secondary Button',
    variant: 'secondary'
  }
};

export const Outline: Story = {
  args: {
    children: 'Outline Button',
    variant: 'outline'
  }
};

// Size variations
export const Small: Story = {
  args: {
    children: 'Small Button',
    size: 'sm'
  }
};

export const Large: Story = {
  args: {
    children: 'Large Button',
    size: 'lg'
  }
};

// State variations
export const Disabled: Story = {
  args: {
    children: 'Disabled Button',
    disabled: true
  }
};

export const Loading: Story = {
  args: {
    children: 'Loading Button',
    isLoading: true
  }
};

// All variants showcase
export const AllVariants: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
      <Button variant="primary">Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="primary" disabled>Disabled</Button>
      <Button variant="primary" isLoading>Loading</Button>
    </div>
  )
};

// All sizes showcase
export const AllSizes: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">Large</Button>
    </div>
  )
};

// Interaction test
export const WithInteraction: Story = {
  args: {
    children: 'Click to test'
  },
  play: async ({ canvasElement, step }) => {
    const { userEvent, within, expect } = await import('@storybook/test');
    const canvas = within(canvasElement);
    
    await step('Click button', async () => {
      const button = canvas.getByRole('button');
      await userEvent.click(button);
    });
    
    await step('Verify focus', async () => {
      const button = canvas.getByRole('button');
      await expect(button).toHaveFocus();
    });
  }
};
```

### Accessibility Test Story

**Dedicated a11y story**:

```typescript
export const AccessibilityTest: Story = {
  args: {
    children: 'Accessibility Test Button'
  },
  parameters: {
    a11y: {
      element: '#storybook-root',
      config: {
        rules: [
          { id: 'color-contrast', enabled: true },
          { id: 'button-name', enabled: true },
          { id: 'focus-order', enabled: true }
        ]
      }
    }
  }
};
```

---

## Example 6: Visual Regression Testing (Chromatic)

### GitHub Actions Workflow

**File**: `.github/workflows/chromatic.yml`

```yaml
name: Chromatic

on:
  push:
    branches:
      - main
      - develop
  pull_request:
    branches:
      - main

jobs:
  chromatic:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0 # Required for Chromatic
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Build Storybook
        run: npm run build-storybook
      
      - name: Publish to Chromatic
        uses: chromaui/action@v1
        with:
          projectToken: ${{ secrets.CHROMATIC_PROJECT_TOKEN }}
          exitZeroOnChanges: true
          onlyChanged: true # Only test changed components
          buildScriptName: 'build-storybook'
```

### Package.json Scripts

```json
{
  "scripts": {
    "storybook": "storybook dev -p 6006",
    "build-storybook": "storybook build",
    "chromatic": "chromatic --project-token=$CHROMATIC_PROJECT_TOKEN",
    "test:visual": "npm run build-storybook && npm run chromatic"
  },
  "devDependencies": {
    "@chromatic-com/storybook": "^1.0.0",
    "@storybook/addon-a11y": "^8.0.0",
    "@storybook/react-vite": "^8.0.0",
    "chromatic": "^11.0.0"
  }
}
```

---

## Complete Example: Production-Ready Button Component

**All files combined for a production-ready atomic component:**

### 1. Component (`Button.tsx`)
### 2. Styles (`Button.module.css`)
### 3. Tests (`Button.test.tsx`)
### 4. Stories (`Button.stories.tsx`)
### 5. Index (`index.ts`)

**See SKILL.md Level 2 for complete implementations.**

---

**This examples file demonstrates:**
- ✅ DTCG 2025.10 token structure
- ✅ WCAG 2.2 AA/AAA compliance
- ✅ Figma MCP integration workflow
- ✅ Storybook 8 documentation patterns
- ✅ Visual regression testing setup
- ✅ Accessibility testing automation
- ✅ Production-ready component architecture
