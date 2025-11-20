# Design Systems Examples

Practical implementations for design tokens, components, and accessibility.

---

## DTCG Tokens

### Basic Token Structure

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "color": {
      "$type": "color",
      "primary": {
        "50": { "$value": "#eff6ff" },
        "500": { "$value": "#3b82f6" },
        "900": { "$value": "#1e3a8a" }
      },
      "semantic": {
        "text": {
          "primary": { "$value": "{color.gray.900}" },
          "secondary": { "$value": "{color.gray.600}" }
        }
      }
    },
    "spacing": {
      "$type": "dimension",
      "xs": { "$value": "0.25rem" },
      "sm": { "$value": "0.5rem" },
      "md": { "$value": "1rem" },
      "lg": { "$value": "1.5rem" }
    },
    "typography": {
      "$type": "fontFamily",
      "sans": { "$value": ["Inter", "system-ui", "sans-serif"] }
    }
  }
}
```

### Style Dictionary Config

```javascript
// style-dictionary.config.js
export default {
  source: ["tokens/**/*.json"],
  platforms: {
    css: {
      transformGroup: "css",
      buildPath: "build/css/",
      files: [
        {
          destination: "variables.css",
          format: "css/variables",
        },
      ],
    },
    js: {
      transformGroup: "js",
      buildPath: "build/js/",
      files: [
        {
          destination: "tokens.js",
          format: "javascript/es6",
        },
      ],
    },
  },
};
```

**Build**:

```bash
npm install style-dictionary
npx style-dictionary build
```

**Output (CSS)**:

```css
:root {
  --color-primary-500: #3b82f6;
  --spacing-md: 1rem;
  --typography-sans: Inter, system-ui, sans-serif;
}
```

---

## Component Library

### Button with CVA (React)

```typescript
import { cva, type VariantProps } from "class-variance-authority";
import { forwardRef } from "react";

const button = cva("rounded font-medium transition focus:outline-none", {
  variants: {
    intent: {
      primary: "bg-primary-500 text-white hover:bg-primary-600",
      secondary: "bg-gray-200 text-gray-900 hover:bg-gray-300",
      outline: "border border-gray-300 hover:bg-gray-100",
    },
    size: {
      sm: "h-8 px-3 text-sm",
      md: "h-10 px-4 text-base",
      lg: "h-12 px-6 text-lg",
    },
  },
  defaultVariants: {
    intent: "primary",
    size: "md",
  },
});

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof button> {
  isLoading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { className, intent, size, isLoading, children, disabled, ...props },
    ref
  ) => {
    return (
      <button
        ref={ref}
        className={button({ intent, size, className })}
        disabled={disabled || isLoading}
        aria-busy={isLoading}
        {...props}
      >
        {isLoading ? <Spinner /> : children}
      </button>
    );
  }
);
```

### Storybook Story

```typescript
// Button.stories.tsx
import type { Meta, StoryObj } from "@storybook/react";
import { Button } from "./Button";

const meta: Meta<typeof Button> = {
  title: "Atoms/Button",
  component: Button,
  tags: ["autodocs"],
  argTypes: {
    intent: {
      control: "select",
      options: ["primary", "secondary", "outline"],
    },
    size: {
      control: "select",
      options: ["sm", "md", "lg"],
    },
  },
};

export default meta;
type Story = StoryObj<typeof Button>;

export const Primary: Story = {
  args: {
    children: "Click me",
    intent: "primary",
  },
};

export const AllVariants: Story = {
  render: () => (
    <div className="flex gap-4">
      <Button intent="primary">Primary</Button>
      <Button intent="secondary">Secondary</Button>
      <Button intent="outline">Outline</Button>
    </div>
  ),
};
```

---

## Accessibility

### Color Contrast Check

```typescript
// utils/contrast.ts
function getLuminance(rgb: [number, number, number]): number {
  const [r, g, b] = rgb.map((val) => {
    const sRGB = val / 255;
    return sRGB <= 0.03928
      ? sRGB / 12.92
      : Math.pow((sRGB + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

export function getContrastRatio(fg: string, bg: string): number {
  const fgLum = getLuminance(hexToRgb(fg));
  const bgLum = getLuminance(hexToRgb(bg));
  const lighter = Math.max(fgLum, bgLum);
  const darker = Math.min(fgLum, bgLum);
  return (lighter + 0.05) / (darker + 0.05);
}

export function meetsWCAG(
  fg: string,
  bg: string,
  level: "AA" | "AAA" = "AA",
  isLargeText = false
): boolean {
  const ratio = getContrastRatio(fg, bg);

  if (level === "AAA") {
    return isLargeText ? ratio >= 4.5 : ratio >= 7;
  }
  return isLargeText ? ratio >= 3 : ratio >= 4.5;
}

// Usage
const passes = meetsWCAG("#3b82f6", "#ffffff"); // Check blue on white
console.log(`Contrast ratio passes AA: ${passes}`);
```

### Keyboard Navigation

```typescript
// hooks/useKeyboardNav.ts
import { useEffect, useRef } from "react";

export function useKeyboardNav<T extends HTMLElement>(
  options: {
    onEscape?: () => void;
    onEnter?: () => void;
    trapFocus?: boolean;
  } = {}
) {
  const ref = useRef<T>(null);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        options.onEscape?.();
      } else if (e.key === "Enter") {
        options.onEnter?.();
      } else if (e.key === "Tab" && options.trapFocus) {
        const focusable = element.querySelectorAll<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (e.shiftKey && document.activeElement === first) {
          last.focus();
          e.preventDefault();
        } else if (!e.shiftKey && document.activeElement === last) {
          first.focus();
          e.preventDefault();
        }
      }
    };

    element.addEventListener("keydown", handleKeyDown);
    return () => element.removeEventListener("keydown", handleKeyDown);
  }, [options]);

  return ref;
}
```

### Reduced Motion

```css
/* styles/motion.css */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

---

## Figma Integration

### MCP Configuration

```json
// claude_desktop_config.json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "figma-dev-mode-mcp-server"],
      "env": {
        "FIGMA_ACCESS_TOKEN": "your-token"
      }
    }
  }
}
```

### Extract Tokens Workflow

1. **Create Figma Variables** (Color, Spacing, Typography)
2. **Use MCP** to extract:
   ```
   Prompt: "Extract all design tokens from this Figma file"
   ```
3. **Transform** with Style Dictionary:
   ```bash
   npx style-dictionary build
   ```

---

## Accessibility Testing

### jest-axe Setup

```typescript
// tests/setup.ts
import "@testing-library/jest-dom";
import { toHaveNoViolations } from "jest-axe";

expect.extend(toHaveNoViolations);
```

### Component Test

```typescript
// Button.test.tsx
import { render } from "@testing-library/react";
import { axe } from "jest-axe";
import { Button } from "./Button";

describe("Button Accessibility", () => {
  it("should have no a11y violations", async () => {
    const { container } = render(<Button>Click me</Button>);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });

  it("should indicate loading state", () => {
    const { getByRole } = render(<Button isLoading>Loading</Button>);
    const button = getByRole("button");
    expect(button).toHaveAttribute("aria-busy", "true");
  });
});
```

---

**See also**: [reference.md](./reference.md) for WCAG guidelines and Figma MCP API details.
