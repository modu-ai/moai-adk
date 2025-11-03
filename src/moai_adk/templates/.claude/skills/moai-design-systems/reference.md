# Design Systems - Technical Reference

## Official Specifications

### W3C Design Tokens Community Group (DTCG)

**Specification**: [Design Tokens Format Module 2025.10](https://www.designtokens.org/tr/drafts/format/)

**Status**: Stable (released 2025-10-28)

**Key Features**:
- Vendor-neutral JSON format
- Multi-file support
- Theming capabilities
- Advanced color support
- Token references (`$ref`, `{token}`)

**Required Properties**:
```json
{
  "tokenName": {
    "$value": "...",    // Required: token value
    "$type": "...",     // Required: token type
    "$description": ""  // Optional: human-readable description
  }
}
```

**Supported Token Types**:
- `color` - Hex, RGB, HSL
- `dimension` - Pixels, rem, em
- `fontFamily` - Font stack
- `fontWeight` - 100-900 or keywords
- `duration` - Milliseconds, seconds
- `cubicBezier` - Easing functions
- `number` - Unitless values
- `strokeStyle` - Solid, dashed, dotted
- `border` - Composite token
- `transition` - Composite token
- `shadow` - Composite token
- `gradient` - Composite token
- `typography` - Composite token

### WCAG 2.1 Conformance Levels

**Specification**: [Web Content Accessibility Guidelines 2.1](https://www.w3.org/TR/WCAG21/)

**Level A** (Foundational):
- 1.1.1 Non-text Content: Provide text alternatives
- 2.1.1 Keyboard: All functionality keyboard accessible
- 2.1.2 No Keyboard Trap: Focus can move away
- 2.4.1 Bypass Blocks: Skip navigation links
- 3.1.1 Language of Page: Specify default language
- 4.1.1 Parsing: No major HTML errors
- 4.1.2 Name, Role, Value: UI components have accessible names

**Level AA** (Industry Standard):
- 1.4.3 Contrast (Minimum): 4.5:1 for text, 3:1 for UI components
- 1.4.5 Images of Text: Use real text, not images
- 1.4.10 Reflow: No 2D scroll at 320px width
- 1.4.11 Non-text Contrast: 3:1 for UI components
- 1.4.12 Text Spacing: Support custom line-height, letter-spacing
- 1.4.13 Content on Hover: Hoverable, dismissable, persistent
- 2.4.7 Focus Visible: Keyboard focus indicator visible
- 3.2.4 Consistent Identification: Same function = same label
- 4.1.3 Status Messages: Programmatic announcement

**Level AAA** (Enhanced):
- 1.4.6 Contrast (Enhanced): 7:1 for text, 4.5:1 for large text
- 1.4.8 Visual Presentation: Line-height 1.5+, paragraph spacing 2x
- 2.4.8 Location: User knows where they are (breadcrumbs)
- 2.5.5 Target Size: 44x44 CSS pixels minimum
- 3.1.3 Unusual Words: Definitions provided
- 3.1.4 Abbreviations: Expanded form or definition

---

## Tool Ecosystem

### Style Dictionary v4

**Documentation**: [https://styledictionary.com/](https://styledictionary.com/)

**Configuration** (`style-dictionary.config.json`):
```json
{
  "source": ["tokens/**/*.json"],
  "platforms": {
    "css": {
      "transformGroup": "css",
      "buildPath": "dist/css/",
      "files": [{
        "destination": "variables.css",
        "format": "css/variables"
      }]
    },
    "js": {
      "transformGroup": "js",
      "buildPath": "dist/js/",
      "files": [{
        "destination": "tokens.js",
        "format": "javascript/es6"
      }]
    }
  }
}
```

**Transformation Pipeline**:
1. **Source**: Read JSON token files
2. **Transform**: Apply platform-specific transformations (px → rem, color formats)
3. **Format**: Output in platform format (CSS, SCSS, JS, iOS, Android)
4. **Build**: Write to destination files

**Custom Transforms** (example: px to rem):
```javascript
StyleDictionary.registerTransform({
  name: 'size/pxToRem',
  type: 'value',
  matcher: (token) => token.type === 'dimension',
  transformer: (token) => {
    const value = parseFloat(token.value);
    return `${value / 16}rem`;
  }
});
```

### Figma MCP Integration

**Setup** (`.claude/mcp.json`):
```json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "@figma/mcp-figma"],
      "env": {
        "FIGMA_PERSONAL_ACCESS_TOKEN": "figd_xxxxx"
      }
    }
  }
}
```

**Personal Access Token Creation**:
1. Go to Figma Settings → Account
2. Scroll to "Personal access tokens"
3. Click "Generate new token"
4. Name: "MCP Server Access"
5. Scopes: File content (read)
6. Copy token (starts with `figd_`)

**MCP Capabilities**:
- Read Figma file structure (frames, components, variants)
- Extract design tokens (colors, typography, spacing)
- Query component properties (width, height, fills, strokes)
- Generate code specifications from designs
- Validate token usage across designs

**Example AI Prompt** (with MCP active):
```
"Extract all color tokens from the 'Design System' page in this Figma file
and generate a JSON file following W3C DTCG spec"
```

### Storybook 8.0

**Documentation**: [https://storybook.js.org/](https://storybook.js.org/)

**Configuration** (`.storybook/main.ts`):
```typescript
import type { StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
  stories: ['../src/**/*.mdx', '../src/**/*.stories.@(js|jsx|ts|tsx)'],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions',
    '@storybook/addon-a11y', // Accessibility testing addon
  ],
  framework: {
    name: '@storybook/react-vite',
    options: {},
  },
  docs: {
    autodocs: 'tag', // Auto-generate docs for components with 'autodocs' tag
  },
};

export default config;
```

**Accessibility Addon**:
```typescript
// .storybook/preview.ts
import { withA11y } from '@storybook/addon-a11y';

export const decorators = [withA11y];

export const parameters = {
  a11y: {
    config: {
      rules: [
        {
          id: 'color-contrast',
          enabled: true,
        },
        {
          id: 'label',
          enabled: true,
        },
      ],
    },
  },
};
```

---

## Atomic Design Hierarchy

### Atoms

**Characteristics**:
- Smallest indivisible components
- No internal state (usually)
- Highly reusable
- Represent design tokens visually

**Examples**:
- `<Button />` - Single button element
- `<Input />` - Text input field
- `<Label />` - Form label
- `<Icon />` - SVG icon wrapper
- `<Typography />` - Text renderer with variants (h1, h2, body, caption)
- `<Spacer />` - Spacing utility component

**Props Pattern**:
```typescript
// Minimal, focused API
interface ButtonProps {
  variant: 'primary' | 'secondary';
  size: 'small' | 'medium' | 'large';
  disabled?: boolean;
  children: ReactNode;
  onClick: () => void;
}
```

### Molecules

**Characteristics**:
- Combination of 2-5 atoms
- Represent simple UI patterns
- Limited internal logic
- Single responsibility

**Examples**:
- `<FormField />` - Label + Input + Error message
- `<SearchBox />` - Input + Search icon + Clear button
- `<Badge />` - Icon + Text
- `<Avatar />` - Image + Fallback text
- `<Pagination />` - Previous/Next buttons + Page numbers

**Props Pattern**:
```typescript
interface FormFieldProps {
  label: string;
  input: ReactElement; // Atom: <Input />
  error?: string;
  helperText?: string;
  required?: boolean;
}
```

### Organisms

**Characteristics**:
- Combination of molecules and atoms
- Represent complete UI sections
- May contain business logic
- Context-aware (forms, navigation)

**Examples**:
- `<Header />` - Logo + Navigation + Search + User menu
- `<Footer />` - Links + Copyright + Social icons
- `<DataTable />` - Headers + Rows + Pagination + Sorting
- `<Modal />` - Header + Body + Footer + Close button
- `<Form />` - Multiple FormFields + Submit button + Validation

**Props Pattern**:
```typescript
interface HeaderProps {
  logo: ReactElement;
  navigation: NavigationItem[];
  searchBox?: ReactElement;
  userMenu?: ReactElement;
  onNavigate: (path: string) => void;
}
```

### Templates

**Characteristics**:
- Page-level layouts
- No real content (placeholders)
- Focus on structure and grid
- Responsive breakpoints defined

**Examples**:
- `<DashboardLayout />` - Sidebar + Header + Main content area
- `<BlogPostLayout />` - Header + Hero image + Article body + Comments
- `<SettingsLayout />` - Tabs navigation + Content panel

**Props Pattern**:
```typescript
interface DashboardLayoutProps {
  sidebar: ReactElement;
  header: ReactElement;
  children: ReactNode; // Main content
  footer?: ReactElement;
}
```

### Pages

**Characteristics**:
- Templates with real content
- Demonstrate actual use cases
- Test design system resilience
- Reflect production scenarios

**Examples**:
- `<HomePage />` - Uses DashboardLayout with real data
- `<ProductPage />` - Uses BlogPostLayout with product info
- `<SettingsPage />` - Uses SettingsLayout with user settings

---

## ARIA Patterns & Examples

### Modal Dialog

**HTML Structure**:
```html
<div role="dialog" aria-modal="true" aria-labelledby="dialog-title">
  <h2 id="dialog-title">Confirm Action</h2>
  <p>Are you sure you want to delete this item?</p>
  <button>Cancel</button>
  <button>Delete</button>
</div>
```

**Keyboard Interactions**:
- `Tab` - Move focus within modal
- `Shift + Tab` - Move focus backward
- `Esc` - Close modal
- Focus trap: Focus cycles within modal, cannot leave

**React Implementation**:
```typescript
import { useEffect, useRef } from 'react';
import FocusTrap from 'focus-trap-react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export function Modal({ isOpen, onClose, title, children }: ModalProps) {
  const titleId = useRef(`modal-title-${Math.random()}`);
  
  useEffect(() => {
    if (!isOpen) return;
    
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);
  
  if (!isOpen) return null;
  
  return (
    <FocusTrap>
      <div role="dialog" aria-modal="true" aria-labelledby={titleId.current}>
        <h2 id={titleId.current}>{title}</h2>
        {children}
      </div>
    </FocusTrap>
  );
}
```

### Tabs

**HTML Structure**:
```html
<div role="tablist" aria-label="Settings tabs">
  <button role="tab" aria-selected="true" aria-controls="panel-1">Account</button>
  <button role="tab" aria-selected="false" aria-controls="panel-2">Privacy</button>
  <button role="tab" aria-selected="false" aria-controls="panel-3">Notifications</button>
</div>

<div role="tabpanel" id="panel-1" aria-labelledby="tab-1">
  Account settings content...
</div>
```

**Keyboard Interactions**:
- `Arrow Left/Right` - Navigate between tabs
- `Home` - First tab
- `End` - Last tab
- `Tab` - Move focus to panel content

### Dropdown Menu

**HTML Structure**:
```html
<button aria-haspopup="true" aria-expanded="false" aria-controls="menu-1">
  Options
</button>

<ul role="menu" id="menu-1">
  <li role="menuitem">Edit</li>
  <li role="menuitem">Delete</li>
  <li role="separator"></li>
  <li role="menuitem">Share</li>
</ul>
```

**Keyboard Interactions**:
- `Enter/Space` - Open menu
- `Arrow Up/Down` - Navigate items
- `Esc` - Close menu
- `Home/End` - First/last item

---

## Testing Strategies

### Unit Testing (Vitest + Testing Library)

**Test File Structure**:
```typescript
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from './Button';

describe('Button', () => {
  it('renders with correct text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument();
  });
  
  it('calls onClick when clicked', async () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    await userEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });
  
  it('is disabled when disabled prop is true', () => {
    render(<Button disabled>Click me</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });
});
```

### Accessibility Testing (jest-axe)

**Test File**:
```typescript
import { render } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import { Form } from './Form';

expect.extend(toHaveNoViolations);

describe('Form Accessibility', () => {
  it('should not have accessibility violations', async () => {
    const { container } = render(
      <Form>
        <FormField label="Email" input={<Input type="email" />} />
        <Button type="submit">Submit</Button>
      </Form>
    );
    
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
});
```

### Visual Regression Testing (Chromatic)

**Storybook + Chromatic Integration**:
```bash
# Install Chromatic
npm install --save-dev chromatic

# Run visual tests
npx chromatic --project-token=<your-token>
```

**CI/CD Integration** (GitHub Actions):
```yaml
name: Visual Regression
on: [push]

jobs:
  chromatic:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      - run: npx chromatic --project-token=${{ secrets.CHROMATIC_TOKEN }}
```

### E2E Testing (Playwright)

**Test File**:
```typescript
import { test, expect } from '@playwright/test';

test('complete registration flow', async ({ page }) => {
  await page.goto('/register');
  
  // Fill form
  await page.fill('[aria-label="Email"]', 'user@example.com');
  await page.fill('[aria-label="Password"]', 'SecurePass123!');
  await page.click('button:has-text("Sign up")');
  
  // Verify success
  await expect(page.locator('text=Welcome!')).toBeVisible();
});

test('keyboard navigation', async ({ page }) => {
  await page.goto('/');
  
  // Tab through interactive elements
  await page.keyboard.press('Tab');
  await expect(page.locator(':focus')).toHaveAttribute('role', 'button');
  
  await page.keyboard.press('Tab');
  await expect(page.locator(':focus')).toHaveAttribute('role', 'link');
});
```

---

## Version Control & Release Strategy

### Token Versioning (Git)

**Directory Structure**:
```
tokens/
├─ core/
│  ├─ color.json
│  ├─ typography.json
│  ├─ spacing.json
├─ semantic/
│  ├─ buttons.json
│  ├─ forms.json
├─ themes/
│  ├─ light.json
│  ├─ dark.json
└─ package.json  # Version number
```

**Semantic Versioning**:
- **MAJOR**: Breaking changes (renamed tokens, removed tokens)
- **MINOR**: New tokens added (backward compatible)
- **PATCH**: Bug fixes (corrected values)

### Component Versioning (npm)

**package.json**:
```json
{
  "name": "@company/design-system",
  "version": "2.1.0",
  "exports": {
    "./tokens": "./dist/tokens/index.js",
    "./components": "./dist/components/index.js"
  },
  "peerDependencies": {
    "react": "^19.0.0"
  }
}
```

**Release Process**:
1. Run tests: `npm test`
2. Build: `npm run build`
3. Version bump: `npm version minor`
4. Publish: `npm publish`
5. Tag in Git: `git tag v2.1.0 && git push --tags`

### Migration Guides

**Example: v1 → v2 Breaking Changes**:
```markdown
# Migration Guide: v1 → v2

## Breaking Changes

### Renamed Tokens
- `color-brand-primary` → `color-primary`
- `spacing-unit-base` → `spacing-base`

**Migration**:
```bash
# Automated find-replace
find src/ -type f -exec sed -i 's/color-brand-primary/color-primary/g' {} +
```

### Removed Components
- `<LegacyButton />` - Use `<Button variant="legacy" />` instead

**Migration**:
```tsx
// Before
<LegacyButton>Click</LegacyButton>

// After
<Button variant="legacy">Click</Button>
```
```

---

## Performance Optimization

### Token Delivery

**CSS Custom Properties** (runtime theming):
```css
/* Pros: Runtime theme switching, easy debugging */
/* Cons: ~5-10ms render overhead */
:root {
  --color-primary: #0066CC;
}

button {
  background: var(--color-primary);
}
```

**CSS-in-JS** (compile-time optimization):
```typescript
// Pros: Type safety, optimized bundles
// Cons: No runtime theming
import { styled } from '@stitches/react';
import { tokens } from './tokens';

const Button = styled('button', {
  background: tokens.color.primary,
});
```

### Component Lazy Loading

```typescript
import { lazy, Suspense } from 'react';

const HeavyChart = lazy(() => import('./HeavyChart'));

function Dashboard() {
  return (
    <Suspense fallback={<Spinner />}>
      <HeavyChart />
    </Suspense>
  );
}
```

---

**Reference Version**: 1.0.0 (2025-11-04)
