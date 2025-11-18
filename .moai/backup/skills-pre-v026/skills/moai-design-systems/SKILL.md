---
name: moai-design-systems
description: Design system patterns, W3C DTCG 2025.10 token architecture, WCAG 2.2 accessibility standards, and Figma MCP workflows for consistent, accessible UI development
allowed-tools: [Read, Write, WebFetch, WebSearch]
tier: medium
version: "1.0.0"
updated: 2025-11-04
tags: [design-systems, design-tokens, accessibility, wcag, figma-mcp, dtcg]
status: stable
---

# Design Systems Development Skill

**Purpose**: Guide implementation of production-ready design systems using W3C DTCG 2025.10 token standards, WCAG 2.2 accessibility compliance, and Figma MCP automation workflows.

**When to use this Skill**:
- Setting up design token architecture for multi-platform projects
- Implementing accessible component libraries (WCAG 2.2 AA/AAA)
- Automating design-to-code workflows with Figma MCP
- Building maintainable design systems with Storybook
- Ensuring color contrast compliance and semantic token naming

**Latest Standards** (as of November 2025):
- **DTCG Specification**: 2025.10 (first stable version)
- **WCAG Guidelines**: 2.2 (AA: 4.5:1 text, AAA: 7:1 text)
- **Figma MCP**: Desktop + Remote server support
- **Style Dictionary**: 4.0 (DTCG-compatible)
- **Storybook**: 8.x (with Docs addon)

---

## Progressive Disclosure Structure

### Level 1: Quick Start Overview (Read This First)

**Design System Foundation** - Three Pillars:

1. **Design Tokens** (Single Source of Truth)
   - Color, typography, spacing, borders, shadows
   - Semantic naming: `color.primary.500`, `spacing.md`, `font.heading.lg`
   - Multi-theme support (light/dark modes)
   - Format: W3C DTCG 2025.10 JSON or Style Dictionary 4.0

2. **Component Library** (Atomic Design Pattern)
   - Atoms → Molecules → Organisms → Templates → Pages
   - Props API for reusability and composition
   - Variant states: default, hover, active, disabled, error, loading
   - Documentation: Storybook with auto-generated props/usage

3. **Accessibility Standards** (WCAG 2.2 Compliance)
   - Color contrast: 4.5:1 (AA), 7:1 (AAA) for text
   - Keyboard navigation: Tab order, focus management
   - Screen readers: ARIA roles, labels, live regions
   - Motion: `prefers-reduced-motion` support

**Tool Ecosystem Quick Reference**:

| Tool | Version | Purpose | Official Link |
|------|---------|---------|---------------|
| **W3C DTCG** | 2025.10 | Design token specification | https://designtokens.org |
| **Style Dictionary** | 4.0+ | Token transformation engine | https://styledictionary.com |
| **Figma MCP** | Latest | Design-to-code automation | https://help.figma.com/hc/en-us/articles/32132100833559 |
| **Storybook** | 8.x | Component documentation | https://storybook.js.org |
| **axe DevTools** | Latest | Accessibility testing | https://www.deque.com/axe/devtools/ |
| **Chromatic** | Latest | Visual regression testing | https://chromatic.com |

**Decision Points Checklist**:

- [ ] Choose token format: DTCG 2025.10 or Style Dictionary 4.0 (both compatible)
- [ ] Target WCAG level: AA (4.5:1) or AAA (7:1) contrast
- [ ] Component pattern: Atomic Design or alternative structure
- [ ] Documentation tool: Storybook, zeroheight, or custom
- [ ] Figma integration: MCP server (desktop vs remote)
- [ ] Testing strategy: Visual regression + accessibility + interaction

---

### Level 2: Implementation Patterns (How to Build)

#### Pattern 1: Design Token Architecture (DTCG 2025.10)

**Token Structure** - Semantic Naming Convention:

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "color": {
      "$type": "color",
      "primary": {
        "50": { "$value": "#eff6ff" },
        "100": { "$value": "#dbeafe" },
        "500": { "$value": "#3b82f6" },
        "900": { "$value": "#1e3a8a" }
      },
      "semantic": {
        "text": {
          "primary": { "$value": "{color.gray.900}" },
          "secondary": { "$value": "{color.gray.600}" },
          "disabled": { "$value": "{color.gray.400}" }
        },
        "background": {
          "default": { "$value": "{color.white}" },
          "elevated": { "$value": "{color.gray.50}" }
        }
      }
    },
    "spacing": {
      "$type": "dimension",
      "xs": { "$value": "0.25rem" },
      "sm": { "$value": "0.5rem" },
      "md": { "$value": "1rem" },
      "lg": { "$value": "1.5rem" },
      "xl": { "$value": "2rem" }
    },
    "typography": {
      "$type": "fontFamily",
      "sans": { "$value": ["Inter", "system-ui", "sans-serif"] },
      "mono": { "$value": ["JetBrains Mono", "monospace"] }
    },
    "fontSize": {
      "$type": "dimension",
      "sm": { "$value": "0.875rem" },
      "base": { "$value": "1rem" },
      "lg": { "$value": "1.125rem" },
      "xl": { "$value": "1.25rem" }
    }
  }
}
```

**Multi-Theme Support** (Light/Dark Mode):

```json
{
  "color": {
    "semantic": {
      "background": {
        "$type": "color",
        "default": {
          "$value": "{color.white}",
          "$extensions": {
            "mode": {
              "dark": "{color.gray.900}"
            }
          }
        }
      }
    }
  }
}
```

**Style Dictionary Configuration** (v4.0+):

```javascript
// style-dictionary.config.js
export default {
  source: ['tokens/**/*.json'],
  platforms: {
    css: {
      transformGroup: 'css',
      buildPath: 'build/css/',
      files: [{
        destination: 'variables.css',
        format: 'css/variables'
      }]
    },
    js: {
      transformGroup: 'js',
      buildPath: 'build/js/',
      files: [{
        destination: 'tokens.js',
        format: 'javascript/es6'
      }]
    }
  }
};
```

#### Pattern 2: Atomic Design Component Structure

**Folder Hierarchy**:

```
src/design-system/
├── tokens/                    # Design tokens (DTCG format)
│   ├── color.json
│   ├── typography.json
│   └── spacing.json
├── components/
│   ├── atoms/                 # Basic building blocks
│   │   ├── Button/
│   │   │   ├── Button.tsx
│   │   │   ├── Button.stories.tsx
│   │   │   ├── Button.test.tsx
│   │   │   └── index.ts
│   │   ├── Input/
│   │   └── Icon/
│   ├── molecules/             # Simple combinations
│   │   ├── FormField/
│   │   ├── SearchBar/
│   │   └── Card/
│   ├── organisms/             # Complex sections
│   │   ├── Header/
│   │   ├── Footer/
│   │   └── DataTable/
│   └── templates/             # Page layouts
│       ├── DashboardLayout/
│       └── AuthLayout/
└── styles/
    ├── global.css
    └── theme.css
```

**Component Props API** (Reusability Pattern):

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

#### Pattern 3: WCAG 2.2 Accessibility Implementation

**Color Contrast Validation** (Automated Check):

```typescript
// utils/a11y/contrast.ts
/**
 * Calculate relative luminance for WCAG compliance
 * @see https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html
 */
function getLuminance(rgb: [number, number, number]): number {
  const [r, g, b] = rgb.map(val => {
    const sRGB = val / 255;
    return sRGB <= 0.03928
      ? sRGB / 12.92
      : Math.pow((sRGB + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/**
 * Calculate contrast ratio between two colors
 * WCAG AA: 4.5:1 (normal text), 3:1 (large text)
 * WCAG AAA: 7:1 (normal text), 4.5:1 (large text)
 */
export function getContrastRatio(
  foreground: string,
  background: string
): number {
  const fgLum = getLuminance(hexToRgb(foreground));
  const bgLum = getLuminance(hexToRgb(background));
  const lighter = Math.max(fgLum, bgLum);
  const darker = Math.min(fgLum, bgLum);
  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Check if color pair meets WCAG AA/AAA requirements
 */
export function meetsWCAG(
  foreground: string,
  background: string,
  level: 'AA' | 'AAA' = 'AA',
  isLargeText: boolean = false
): boolean {
  const ratio = getContrastRatio(foreground, background);
  
  if (level === 'AAA') {
    return isLargeText ? ratio >= 4.5 : ratio >= 7;
  }
  
  // AA level
  return isLargeText ? ratio >= 3 : ratio >= 4.5;
}
```

**Keyboard Navigation** (Focus Management):

```typescript
// hooks/useKeyboardNavigation.ts
import { useEffect, useRef } from 'react';

export function useKeyboardNavigation<T extends HTMLElement>(
  options: {
    onEscape?: () => void;
    onEnter?: () => void;
    trapFocus?: boolean;
  } = {}
) {
  const elementRef = useRef<T>(null);

  useEffect(() => {
    const element = elementRef.current;
    if (!element) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        options.onEscape?.();
      } else if (e.key === 'Enter') {
        options.onEnter?.();
      } else if (e.key === 'Tab' && options.trapFocus) {
        // Focus trap implementation
        const focusableElements = element.querySelectorAll<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];

        if (e.shiftKey && document.activeElement === firstElement) {
          lastElement.focus();
          e.preventDefault();
        } else if (!e.shiftKey && document.activeElement === lastElement) {
          firstElement.focus();
          e.preventDefault();
        }
      }
    };

    element.addEventListener('keydown', handleKeyDown);
    return () => element.removeEventListener('keydown', handleKeyDown);
  }, [options]);

  return elementRef;
}
```

**ARIA Labels & Roles** (Screen Reader Support):

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

**Motion Accessibility** (Reduced Motion Support):

```css
/* styles/motion.css */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}

/* Safe animations for reduced motion users */
.fade-enter {
  opacity: 0;
}

.fade-enter-active {
  opacity: 1;
  transition: opacity 200ms ease-in;
}

@media (prefers-reduced-motion: reduce) {
  .fade-enter-active {
    transition: none;
    opacity: 1;
  }
}
```

#### Pattern 4: Figma MCP Integration Workflow

**Setup** (Desktop Server):

```bash
# Install Figma desktop app
# Enable MCP server in settings
# Desktop server runs at http://127.0.0.1:3845/mcp
```

**MCP Configuration** (Claude Desktop):

```json
{
  "mcpServers": {
    "figma": {
      "command": "figma-mcp",
      "args": [],
      "env": {
        "FIGMA_ACCESS_TOKEN": "your-figma-token"
      }
    }
  }
}
```

**Extracting Design Tokens from Figma**:

1. **Create Figma Variables** (Color, Typography, Spacing)
2. **Use MCP Server** to extract variables:
   - Select frame in Figma
   - Prompt: "Extract all design tokens from this frame"
   - MCP returns DTCG-compatible JSON
3. **Transform to Code** using Style Dictionary

**Component Code Generation**:

```
User Workflow:
1. Select component frame in Figma
2. Prompt: "Generate React component from this design"
3. MCP extracts:
   - Component structure
   - Applied design tokens
   - Layout properties (flex, grid)
   - Typography and spacing
4. Output: TypeScript React component with props
```

**Automation Pattern** (Link-based workflow):

```typescript
// scripts/sync-figma-tokens.ts
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

async function syncFigmaTokens(figmaFileUrl: string) {
  // Use Figma MCP to extract tokens
  const { stdout } = await execAsync(
    `figma-mcp extract-tokens --url="${figmaFileUrl}"`
  );
  
  const tokens = JSON.parse(stdout);
  
  // Write to tokens directory
  await writeFile('tokens/color.json', JSON.stringify(tokens.color, null, 2));
  await writeFile('tokens/spacing.json', JSON.stringify(tokens.spacing, null, 2));
  
  // Run Style Dictionary build
  await execAsync('npm run tokens:build');
  
  console.log('✅ Design tokens synchronized from Figma');
}
```

#### Pattern 5: Storybook Documentation Setup

**Installation & Configuration**:

```bash
npx storybook@latest init
```

**Storybook Configuration** (v8.x):

```typescript
// .storybook/main.ts
import type { StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(js|jsx|ts|tsx|mdx)'],
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

**Component Story** (with accessibility tests):

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

---

### Level 3: Advanced Topics (Deep Expertise)

#### Advanced 1: Type-Safe Design Tokens (TypeScript)

**Generate TypeScript Types from DTCG Tokens**:

```typescript
// scripts/generate-token-types.ts
import { readFileSync, writeFileSync } from 'fs';

interface DTCGToken {
  $value: string | number | string[];
  $type?: string;
  [key: string]: any;
}

function generateTypes(tokens: Record<string, any>, prefix = ''): string {
  let types = '';
  
  for (const [key, value] of Object.entries(tokens)) {
    if (value.$value !== undefined) {
      const tokenPath = `${prefix}${key}`.replace(/\./g, '-');
      types += `export const ${tokenPath} = '${value.$value}';\n`;
    } else {
      types += generateTypes(value, `${prefix}${key}.`);
    }
  }
  
  return types;
}

const colorTokens = JSON.parse(readFileSync('tokens/color.json', 'utf-8'));
const types = generateTypes(colorTokens.$tokens);
writeFileSync('src/tokens/colors.ts', types);
```

#### Advanced 2: Visual Regression Testing (Chromatic)

```bash
# Install Chromatic
npm install --save-dev chromatic

# Configure in package.json
{
  "scripts": {
    "chromatic": "chromatic --project-token=<your-token>"
  }
}
```

**CI/CD Integration**:

```yaml
# .github/workflows/chromatic.yml
name: Chromatic

on: [push]

jobs:
  chromatic:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run Chromatic
        uses: chromaui/action@v1
        with:
          projectToken: ${{ secrets.CHROMATIC_PROJECT_TOKEN }}
```

#### Advanced 3: Accessibility Testing Automation

**Jest + jest-axe Configuration**:

```typescript
// tests/setup.ts
import '@testing-library/jest-dom';
import { toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);
```

**Component Accessibility Tests**:

```typescript
// components/atoms/Button/Button.test.tsx
import { render } from '@testing-library/react';
import { axe } from 'jest-axe';
import { Button } from './Button';

describe('Button Accessibility', () => {
  it('should have no accessibility violations', async () => {
    const { container } = render(<Button>Click me</Button>);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });

  it('should have correct ARIA attributes when disabled', () => {
    const { getByRole } = render(<Button disabled>Disabled</Button>);
    const button = getByRole('button');
    expect(button).toHaveAttribute('aria-disabled', 'true');
  });

  it('should indicate loading state to screen readers', () => {
    const { getByRole } = render(<Button isLoading>Loading</Button>);
    const button = getByRole('button');
    expect(button).toHaveAttribute('aria-busy', 'true');
  });
});
```

---

## Best Practices Checklist

Design Token Architecture:
- [ ] Use semantic naming (`color.primary.500` not `color.blue`)
- [ ] Implement aliasing for themes (`{color.white}` references)
- [ ] Validate DTCG 2025.10 spec compliance
- [ ] Version tokens with semantic versioning
- [ ] Document token usage in Storybook

Component Development:
- [ ] Follow Atomic Design hierarchy (Atoms → Molecules → Organisms)
- [ ] Create variant-based props APIs (not separate components)
- [ ] Document all props with TypeScript types
- [ ] Write Storybook stories for all variants
- [ ] Test component accessibility with jest-axe

Accessibility:
- [ ] Verify 4.5:1 contrast for all text (WCAG AA)
- [ ] Implement keyboard navigation for all interactive elements
- [ ] Add ARIA labels to form fields and buttons
- [ ] Test with screen readers (NVDA, JAWS, VoiceOver)
- [ ] Support `prefers-reduced-motion`

Testing:
- [ ] Visual regression tests for all components (Chromatic)
- [ ] Accessibility tests with axe-core
- [ ] Interaction tests with Testing Library
- [ ] Cross-browser compatibility checks

Figma Integration:
- [ ] Set up Figma MCP server (desktop or remote)
- [ ] Extract design tokens from Figma variables
- [ ] Automate component code generation
- [ ] Sync design changes with codebase

---

## When NOT to Use This Skill

- **Simple static sites**: Overkill for projects without complex UI requirements
- **Rapid prototyping**: Design systems add overhead during early exploration
- **Single-use projects**: Token architecture benefits long-term maintenance
- **Non-web platforms**: This Skill focuses on web (React/Vue/TypeScript)

For these cases, consider:
- Plain CSS/Tailwind for static sites
- Component libraries (Material-UI, shadcn/ui) for rapid development
- Platform-specific design systems (iOS HIG, Material Design for Android)

---

**Related Skills**:
- `moai-domain-frontend` - Frontend architecture patterns
- `moai-lang-typescript` - TypeScript best practices
- `moai-foundation-testing` - Testing strategies

**Official Resources**:
- W3C DTCG: https://designtokens.org
- WCAG 2.2: https://www.w3.org/WAI/WCAG22/quickref/
- Figma MCP: https://help.figma.com/hc/en-us/articles/32132100833559
- Style Dictionary: https://styledictionary.com
- Storybook: https://storybook.js.org
