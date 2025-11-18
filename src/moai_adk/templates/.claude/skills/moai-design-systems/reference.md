# Design Systems Technical Reference

Comprehensive reference material for design system development, covering W3C DTCG 2025.10 specification, WCAG 2.2 accessibility standards, and tooling ecosystem.

---

## W3C Design Token Community Group (DTCG) 2025.10

### Specification Overview

**Official Spec**: https://tr.designtokens.org/format/  
**Status**: First stable version (released October 2025)  
**Purpose**: Vendor-agnostic JSON format for sharing design decisions across tools and platforms

### Token Format Structure

**Basic Token Anatomy**:

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "token-name": {
      "$value": "value",
      "$type": "type",
      "$description": "optional description"
    }
  }
}
```

**Key Properties**:
- `$value`: The token's value (string, number, array, or object)
- `$type`: Token type (color, dimension, fontFamily, fontWeight, duration, etc.)
- `$description`: Optional human-readable description
- `$extensions`: Optional vendor-specific extensions

### Supported Token Types

| Type | Example Value | Use Case |
|------|---------------|----------|
| `color` | `"#3b82f6"` or `"rgb(59, 130, 246)"` | Text, backgrounds, borders |
| `dimension` | `"1.5rem"` or `"24px"` | Spacing, sizing, typography |
| `fontFamily` | `["Inter", "sans-serif"]` | Font stacks |
| `fontWeight` | `700` or `"bold"` | Font weights |
| `duration` | `"200ms"` | Animation timing |
| `cubicBezier` | `[0.4, 0, 0.2, 1]` | Easing functions |
| `number` | `1.5` | Line heights, opacity |
| `strokeStyle` | `"solid"` or `"dashed"` | Border styles |
| `border` | `{ "color": "...", "width": "...", "style": "..." }` | Composite borders |
| `shadow` | `{ "offsetX": "...", "offsetY": "...", "blur": "...", "color": "..." }` | Box shadows |

### Token Aliasing (References)

**Syntax**: Use curly braces `{token.path}` to reference other tokens

```json
{
  "color": {
    "primary": {
      "500": { "$value": "#3b82f6", "$type": "color" }
    },
    "semantic": {
      "action": {
        "$value": "{color.primary.500}",
        "$type": "color",
        "$description": "Primary action color"
      }
    }
  }
}
```

**Benefits**:
- Single source of truth (change `primary.500`, all aliases update)
- Semantic naming (abstract away specific values)
- Multi-theme support (swap referenced tokens per theme)

### Theme Support

**Using `$extensions` for mode-specific values**:

```json
{
  "color": {
    "background": {
      "$type": "color",
      "$value": "#ffffff",
      "$extensions": {
        "mode": {
          "dark": "#1a1a1a"
        }
      }
    }
  }
}
```

---

## WCAG 2.2 Accessibility Guidelines

### Color Contrast Requirements

**Level AA** (Minimum):
- **Normal text**: 4.5:1 contrast ratio
- **Large text**: 3:1 contrast ratio (18pt+ or 14pt+ bold)
- **UI components**: 3:1 contrast ratio (buttons, form borders, icons)

**Level AAA** (Enhanced):
- **Normal text**: 7:1 contrast ratio
- **Large text**: 4.5:1 contrast ratio

**Text Size Definitions**:
- **Normal text**: < 18pt (24px) or < 14pt bold (18.67px)
- **Large text**: ≥ 18pt (24px) or ≥ 14pt bold (18.67px)

**Exceptions**:
- Inactive/disabled components
- Pure decoration (no functional purpose)
- Logotypes (brand names, logos)
- Text within images (where image contains significant other content)

### Contrast Calculation Formula

```
Contrast Ratio = (L1 + 0.05) / (L2 + 0.05)

Where:
- L1 = relative luminance of lighter color
- L2 = relative luminance of darker color
- Luminance = 0.2126 * R + 0.7152 * G + 0.0722 * B
  (R, G, B are sRGB values adjusted for gamma correction)
```

**Practical Examples**:

| Foreground | Background | Ratio | AA Pass | AAA Pass |
|------------|------------|-------|---------|----------|
| `#000000` (black) | `#ffffff` (white) | 21:1 | ✅ | ✅ |
| `#767676` (gray) | `#ffffff` (white) | 4.54:1 | ✅ | ❌ |
| `#ff0000` (red) | `#ffffff` (white) | 4:1 | ❌ | ❌ |
| `#0000ff` (blue) | `#ffffff` (white) | 8.6:1 | ✅ | ✅ |
| `#595959` (dark gray) | `#ffffff` (white) | 7.01:1 | ✅ | ✅ |

**Tools for Testing**:
- WebAIM Contrast Checker: https://webaim.org/resources/contrastchecker/
- Chrome DevTools (Color Picker shows contrast ratio)
- axe DevTools browser extension

### Keyboard Navigation Standards

**Essential Requirements**:
1. **Tab Order**: All interactive elements accessible via Tab key
2. **Focus Indicators**: Visible focus state (at least 3:1 contrast)
3. **Escape Key**: Dismisses modals/dropdowns
4. **Arrow Keys**: Navigate within composite widgets (menus, tabs)
5. **Enter/Space**: Activate buttons and links

**Focus Trap Pattern** (for modals):

```
When modal opens:
1. Focus moves to first interactive element
2. Tab cycles through modal elements only
3. Shift+Tab reverses cycle
4. Escape closes modal and returns focus
```

### ARIA Patterns Reference

**Common Roles**:

| Role | Purpose | Required Attributes |
|------|---------|---------------------|
| `button` | Interactive control | `aria-label` or visible text |
| `dialog` | Modal window | `aria-labelledby` or `aria-label` |
| `alert` | Important message | None (implicitly polite) |
| `alertdialog` | Alert requiring response | `aria-labelledby` or `aria-label` |
| `menu` | Menu widget | `aria-orientation` |
| `menuitem` | Menu option | None |
| `tablist` | Tab navigation | `aria-orientation` |
| `tab` | Individual tab | `aria-selected`, `aria-controls` |
| `tabpanel` | Tab content | `aria-labelledby` |

**State Attributes**:

| Attribute | Values | Use Case |
|-----------|--------|----------|
| `aria-expanded` | `true`/`false` | Collapsible sections, dropdowns |
| `aria-selected` | `true`/`false` | Selected tabs, list items |
| `aria-checked` | `true`/`false`/`mixed` | Checkboxes, radio buttons |
| `aria-disabled` | `true`/`false` | Disabled controls |
| `aria-invalid` | `true`/`false` | Form validation errors |
| `aria-hidden` | `true`/`false` | Hide from screen readers |
| `aria-live` | `polite`/`assertive`/`off` | Dynamic content updates |

**Live Regions** (for dynamic updates):

```html
<!-- Polite: waits for user pause -->
<div aria-live="polite" aria-atomic="true">
  Loading complete: 10 items found
</div>

<!-- Assertive: interrupts immediately -->
<div aria-live="assertive" role="alert">
  Error: Form submission failed
</div>
```

### Motion & Animation Accessibility

**`prefers-reduced-motion` media query**:

```css
/* Default: full animations */
.fade-in {
  animation: fadeIn 300ms ease-in;
}

/* Reduced motion: instant transition */
@media (prefers-reduced-motion: reduce) {
  .fade-in {
    animation: none;
    opacity: 1;
  }
}
```

**Safe Animation Principles**:
- Avoid large/rapid movements (triggers vestibular disorders)
- Provide instant alternatives for reduced motion users
- Never rely solely on animation to convey information
- Test with macOS "Reduce motion" or Windows "Show animations"

---

## Figma MCP Server Reference

### Official Documentation

**Guide**: https://help.figma.com/hc/en-us/articles/32132100833559  
**GitHub**: https://github.com/figma/mcp-server-guide  
**Blog Announcement**: https://www.figma.com/blog/introducing-figma-mcp-server/

### Capabilities

**Core Features**:
1. **Code Generation**: Convert frames to React/Vue/TypeScript code
2. **Design Token Extraction**: Pull variables (colors, spacing, typography)
3. **Component Metadata**: Extract layer names, properties, constraints
4. **Layout Information**: Flex/Grid layouts, auto-layout properties
5. **Multi-File Support**: Access Figma, FigJam, and Make files

### Server Types

**Desktop Server** (Local):
- Endpoint: `http://127.0.0.1:3845/mcp`
- Requirements: Figma desktop app + Dev/Full seat on paid plan
- Benefits: Selection-based workflow, no rate limits

**Remote Server** (Hosted):
- Endpoint: `https://mcp.figma.com/mcp`
- Requirements: Any Figma plan
- Rate Limits: 6 calls/month (Starter/View), per-minute limits (Dev/Full)

### Integration Setup

**Claude Desktop Configuration** (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "@figma/mcp-server"],
      "env": {
        "FIGMA_ACCESS_TOKEN": "your-personal-access-token"
      }
    }
  }
}
```

**VS Code / Cursor Setup**:

```json
{
  "mcp.servers": {
    "figma": {
      "url": "http://127.0.0.1:3845/mcp",
      "auth": {
        "type": "bearer",
        "token": "your-figma-token"
      }
    }
  }
}
```

### Workflow Patterns

**Selection-Based** (Desktop only):
1. Select frame/component in Figma
2. Prompt AI: "Generate React component from this design"
3. MCP extracts selection context
4. Returns code with design tokens applied

**Link-Based** (Desktop + Remote):
1. Copy Figma frame URL
2. Prompt AI: "Extract design tokens from https://figma.com/file/..."
3. MCP fetches node data via API
4. Returns DTCG-compatible JSON

**Code Connect Integration**:
- Link Figma components to actual codebase components
- AI references existing code patterns
- Ensures consistency with established architecture

### Best Practices

1. **Structure Figma Files**: Use semantic layer names (`PrimaryButton`, not `Rectangle 1`)
2. **Use Variables**: Define colors/spacing as Figma variables (easier extraction)
3. **Enable Code Connect**: Link components to code for better AI context
4. **Document Conventions**: Add descriptions to components for AI guidance
5. **Version Control**: Sync Figma updates with codebase regularly

---

## Style Dictionary 4.0 Reference

### Official Documentation

**Website**: https://styledictionary.com  
**GitHub**: https://github.com/style-dictionary/style-dictionary  
**Migration Guide**: https://styledictionary.com/version-4/migration/

### Key Features ( +)

**DTCG Compatibility**:
- Supports both DTCG 2025.10 format (`$value`, `$type`) and original format (`value`, `type`)
- Cannot mix formats in single instance
- Forward-compatible with DTCG spec

**Type Safety**:
- Improved TypeScript types
- Standalone type interfaces via `style-dictionary/types`
- Better IDE autocomplete

**ES Modules**:
- Native ESM support (no CommonJS conversion)
- Async configuration
- Dynamic imports

### Configuration Structure

**Basic Config** (`style-dictionary.config.js`):

```javascript
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
    }
  }
};
```

**Advanced Config** (Multi-platform):

```javascript
export default {
  source: ['tokens/**/*.json'],
  platforms: {
    css: {
      transformGroup: 'css',
      buildPath: 'dist/css/',
      files: [{
        destination: 'variables.css',
        format: 'css/variables',
        options: {
          outputReferences: true // Use CSS var references
        }
      }]
    },
    js: {
      transformGroup: 'js',
      buildPath: 'dist/js/',
      files: [{
        destination: 'tokens.js',
        format: 'javascript/es6'
      }]
    },
    android: {
      transformGroup: 'android',
      buildPath: 'dist/android/',
      files: [{
        destination: 'colors.xml',
        format: 'android/resources'
      }]
    },
    ios: {
      transformGroup: 'ios',
      buildPath: 'dist/ios/',
      files: [{
        destination: 'StyleDictionary.swift',
        format: 'ios-swift/class.swift'
      }]
    }
  }
};
```

### Transform Groups

**Built-in Transform Groups**:

| Group | Transforms | Output Example |
|-------|-----------|----------------|
| `css` | `name/cti/kebab`, `size/px`, `color/css` | `--color-primary-500: #3b82f6;` |
| `js` | `name/cti/camel`, `size/rem`, `color/hex` | `colorPrimary500: '#3b82f6'` |
| `android` | `name/cti/snake`, `size/dp`, `color/hex8` | `<color name="color_primary_500">#FF3B82F6</color>` |
| `ios` | `name/ti/camel`, `size/pt`, `color/UIColor` | `UIColor(red: 0.23, green: 0.51, blue: 0.96, alpha: 1)` |

### Custom Transforms

**Example: PX to REM conversion**:

```javascript
export default {
  transform: {
    'size/pxToRem': {
      type: 'value',
      filter: (token) => token.type === 'dimension' && token.value.endsWith('px'),
      transform: (token) => {
        const px = parseFloat(token.value);
        return `${px / 16}rem`;
      }
    }
  },
  platforms: {
    css: {
      transforms: ['size/pxToRem', 'name/cti/kebab'],
      // ...
    }
  }
};
```

---

## Storybook 8 Reference

### Official Documentation

**Website**: https://storybook.js.org  
**Docs**: https://storybook.js.org/docs/react/get-started/introduction  
**Design Systems**: https://storybook.js.org/blog/4-ways-to-document-your-design-system-with-storybook/

### Key Features (v8.x)

**Docs Addon** (Automatic Documentation):
- Auto-generates component docs from TypeScript types
- Props table extraction
- Usage examples
- Code snippets

**Component Story Format (CSF3)**:
- Simplified story syntax
- TypeScript-first
- Better IDE support

**Accessibility Testing** (a11y addon):
- Built-in axe-core integration
- Real-time violation detection
- WCAG compliance checks

### Story Structure

**Meta Configuration**:

```typescript
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta: Meta<typeof Button> = {
  title: 'Components/Button', // Sidebar hierarchy
  component: Button,
  tags: ['autodocs'], // Auto-generate docs page
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'outline'],
      description: 'Visual style variant',
      table: {
        defaultValue: { summary: 'primary' }
      }
    }
  }
};

export default meta;
type Story = StoryObj<typeof Button>;
```

**Individual Stories**:

```typescript
export const Primary: Story = {
  args: {
    children: 'Click me',
    variant: 'primary'
  }
};

export const WithIcon: Story = {
  args: {
    children: 'Save',
    variant: 'primary',
    icon: <SaveIcon />
  }
};

// Play function for interaction testing
export const Interactive: Story = {
  args: {
    children: 'Toggle',
    variant: 'primary'
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const button = canvas.getByRole('button');
    await userEvent.click(button);
    await expect(button).toHaveFocus();
  }
};
```

### Addons Reference

**Essential Addons**:

| Addon | Purpose | Installation |
|-------|---------|--------------|
| `@storybook/addon-essentials` | Core features (controls, docs, actions) | Included by default |
| `@storybook/addon-a11y` | Accessibility testing | `npm i -D @storybook/addon-a11y` |
| `@storybook/addon-interactions` | User interaction testing | `npm i -D @storybook/addon-interactions` |
| `@storybook/addon-links` | Story navigation | Included in essentials |
| `@chromatic-com/storybook` | Visual regression testing | `npm i -D @chromatic-com/storybook` |

---

## Testing Tools Reference

### axe-core / jest-axe

**Installation**:

```bash
npm install --save-dev jest-axe @axe-core/react
```

**Setup** (`tests/setup.ts`):

```typescript
import '@testing-library/jest-dom';
import { toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);
```

**Usage in Tests**:

```typescript
import { axe } from 'jest-axe';

it('should have no accessibility violations', async () => {
  const { container } = render(<Button>Click me</Button>);
  const results = await axe(container);
  expect(results).toHaveNoViolations();
});
```

**axe DevTools Configuration**:

```typescript
import { configureAxe } from 'jest-axe';

const axe = configureAxe({
  rules: {
    'color-contrast': { enabled: true },
    'label': { enabled: true },
    'button-name': { enabled: true }
  }
});
```

### Chromatic (Visual Regression)

**Official Site**: https://chromatic.com  
**Pricing**: Free for open source, paid for private repos

**Setup**:

```bash
npm install --save-dev chromatic
npx chromatic --project-token=<your-token>
```

**GitHub Actions Integration**:

```yaml
- name: Run Chromatic
  uses: chromaui/action@v1
  with:
    projectToken: ${{ secrets.CHROMATIC_PROJECT_TOKEN }}
    exitZeroOnChanges: true
```

**Review Process**:
1. Push changes to branch
2. Chromatic captures screenshots
3. Compare with baseline
4. Approve/reject visual changes
5. Baseline updates on approval

---

## Version Compatibility Matrix

| Tool | Version | DTCG Support | Release Date |
|------|---------|--------------|--------------|
| DTCG Spec | 2025.10 | N/A | October 2025 |
| Style Dictionary | 4.0+ | ✅ Full | 2024 |
| Figma MCP | Latest | ✅ Variables | 2025 |
| Storybook | 8.x | N/A | 2024 |
| axe-core | 4.8+ | N/A | 2024 |
| React | 18+ | N/A | 2022 |
| TypeScript | 5.0+ | N/A | 2023 |

---

## External Resources

**Specifications**:
- W3C DTCG: https://tr.designtokens.org/format/
- WCAG 2.2: https://www.w3.org/WAI/WCAG22/quickref/
- ARIA Authoring Practices: https://www.w3.org/WAI/ARIA/apg/

**Tools**:
- Style Dictionary: https://styledictionary.com
- Figma MCP: https://help.figma.com/hc/en-us/articles/32132100833559
- Storybook: https://storybook.js.org
- axe DevTools: https://www.deque.com/axe/devtools/
- Chromatic: https://chromatic.com

**Design Systems Examples**:
- Shopify Polaris: https://polaris.shopify.com
- IBM Carbon: https://carbondesignsystem.com
- GitHub Primer: https://primer.style
- Salesforce Lightning: https://lightningdesignsystem.com

**Color Contrast Tools**:
- WebAIM Contrast Checker: https://webaim.org/resources/contrastchecker/
- Coolors Contrast Checker: https://coolors.co/contrast-checker
- Who Can Use: https://whocanuse.com
