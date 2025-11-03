# Design Systems: Practical Examples

Real-world code examples for implementing design systems with DTCG 2025.10 tokens, WCAG 2.2 accessibility, and Figma MCP workflows.

---

## Example 1: DTCG 2025.10 Color Tokens

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "color": {
      "$type": "color",
      "primary": {
        "500": { "$value": "#3b82f6" },
        "600": { "$value": "#2563eb" }
      },
      "semantic": {
        "text": {
          "primary": { "$value": "{color.gray.900}" }
        }
      }
    }
  }
}
```

---

## Example 2: Atomic Design Folder Structure

```
src/design-system/
├── tokens/
│   ├── color.json
│   ├── typography.json
│   └── spacing.json
├── components/
│   ├── atoms/Button/
│   ├── molecules/FormField/
│   └── organisms/Header/
├── hooks/useKeyboardNavigation.ts
└── utils/a11y/contrast.ts
```

---

## Example 3: WCAG 2.2 Contrast Validation

```typescript
function getContrastRatio(fg: string, bg: string): number {
  const fgLum = getLuminance(hexToRgb(fg));
  const bgLum = getLuminance(hexToRgb(bg));
  const lighter = Math.max(fgLum, bgLum);
  const darker = Math.min(fgLum, bgLum);
  return (lighter + 0.05) / (darker + 0.05);
}

function meetsWCAG(fg: string, bg: string, level: 'AA' | 'AAA' = 'AA'): boolean {
  const ratio = getContrastRatio(fg, bg);
  return level === 'AAA' ? ratio >= 7 : ratio >= 4.5;
}
```

---

## Example 4: React Button Component

```typescript
import { forwardRef } from 'react';
import { cva } from 'class-variance-authority';

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-md font-medium',
  {
    variants: {
      variant: {
        primary: 'bg-primary-500 text-white hover:bg-primary-600',
        secondary: 'bg-gray-200 text-gray-900'
      }
    }
  }
);

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant, children, ...props }, ref) => (
    <button ref={ref} className={buttonVariants({ variant })} {...props}>
      {children}
    </button>
  )
);
```

---

## Example 5: Storybook Story

```typescript
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta: Meta<typeof Button> = {
  title: 'Design System/Atoms/Button',
  component: Button,
  tags: ['autodocs']
};

export default meta;

export const Primary: StoryObj = {
  args: {
    children: 'Primary Button',
    variant: 'primary'
  }
};
```

---

## Example 6: Figma MCP Workflow

**Step 1**: Create Figma variables for design tokens
**Step 2**: Extract tokens via MCP: `Extract design tokens as DTCG 2025.10`
**Step 3**: Transform with Style Dictionary
**Step 4**: Generate React components from Figma frames

---

**This skill provides**:
- ✅ DTCG 2025.10 token patterns
- ✅ WCAG 2.2 AA/AAA compliance
- ✅ Figma MCP integration
- ✅ Storybook documentation
- ✅ Production-ready components
