---
name: moai-design-systems
version: 1.0.0
created: 2025-11-04
updated: 2025-11-04
status: active
description: Comprehensive guide for creating accessible, production-grade design systems with design tokens, component libraries, WCAG 2.1 compliance, and Figma MCP integration.
keywords: ['design-systems', 'design-tokens', 'accessibility', 'wcag', 'figma', 'mcp', 'atomic-design', 'component-library', 'storybook']
allowed-tools:
  - Read
  - Bash
---

# Design Systems Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-design-systems |
| **Version** | 1.0.0 (2025-11-04) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Domain |
| **Freedom Level** | Medium (can be specialized for frameworks) |

---

## What It Does

Comprehensive guide for creating accessible, production-grade design systems with design tokens, component libraries, WCAG 2.1 compliance, and Figma MCP integration.

**Key capabilities**:
- ✅ Design token architecture (W3C DTCG spec 2025.10)
- ✅ Component library structure (Atomic Design methodology)
- ✅ Accessibility standards (WCAG 2.1 AA/AAA)
- ✅ Figma MCP workflow (design-to-code automation)
- ✅ Testing & validation strategies
- ✅ Latest tool versions (2025-11-04)

---

## When to Use

**Automatic triggers**:
- Design system creation or refactoring
- Component library architecture
- Accessibility compliance reviews
- Design token implementation
- Figma-to-code workflow setup

**Manual invocation**:
- Design new design system from scratch
- Audit existing design system for WCAG compliance
- Migrate to design token-based architecture
- Set up Figma MCP automation

---

## Progressive Disclosure

### Level 0: Quick Summary (2-3 sentences)

Design systems require four pillars: design tokens (W3C DTCG spec), component libraries (Atomic Design), accessibility compliance (WCAG 2.1 AA minimum), and automation (Figma MCP). Use this Skill when creating production-grade design systems with comprehensive documentation, testing, and version control.

### Level 1: Key Principles & Quick Reference

**Core Principles**:
1. **Token-First Architecture**: All visual decisions stored as design tokens (colors, typography, spacing)
2. **Atomic Component Structure**: Build from atoms → molecules → organisms → templates → pages
3. **Accessibility Baseline**: WCAG 2.1 AA minimum (4.5:1 contrast, keyboard nav, ARIA patterns)
4. **Automated Workflow**: Figma MCP for design token extraction and component sync
5. **Living Documentation**: Storybook with auto-generated docs and interactive examples

**Quick Decision Tree**:
```
Starting a design system?
├─ Define design tokens (colors, typography, spacing)
├─ Structure components (atomic design)
├─ Ensure accessibility (WCAG 2.1 AA)
├─ Set up automation (Figma MCP)
└─ Document everything (Storybook)
```

### Level 2: Detailed Guidance

#### 1. Design Token Architecture (150 words)

**W3C Design Tokens Community Group Format (2025.10 stable)**

Design tokens are the atomic building blocks of your design system—colors, typography, spacing, shadows, and more. Use the W3C DTCG specification (stable as of 2025-10-28) with JSON format:

```json
{
  "color": {
    "primary": {
      "$value": "#0066CC",
      "$type": "color",
      "$description": "Primary brand color"
    },
    "neutral": {
      "50": { "$value": "#F9FAFB", "$type": "color" },
      "900": { "$value": "#111827", "$type": "color" }
    }
  },
  "typography": {
    "heading": {
      "font-family": { "$value": "Inter", "$type": "fontFamily" },
      "font-size": { "$value": "32px", "$type": "dimension" },
      "line-height": { "$value": "1.2", "$type": "number" }
    }
  }
}
```

**Semantic Naming Strategy**:
- Use intent-based names: `primary`, `success`, `error`, `neutral` (not `blue`, `green`, `red`)
- Support theming: `light` and `dark` mode variants
- Export formats: JSON, CSS Custom Properties, SCSS variables, Tailwind config

**Tool Ecosystem**:
- **Style Dictionary v4**: Transform tokens to any platform (iOS, Android, web)
- **Tokens Studio**: Figma plugin for token management
- **Terrazzo**: Design token build tool

#### 2. Component Library Structure (150 words)

**Atomic Design Methodology**

Organize components in five hierarchical levels (Brad Frost, atomicdesign.bradfrost.com):

**Atoms** (basic building blocks):
- Buttons, inputs, labels, icons, typography styles, color swatches
- Example: `<Button variant="primary" size="medium">Click me</Button>`

**Molecules** (simple combinations):
- Form fields (label + input), search box (input + icon), card headers
- Example: `<FormField label="Email" input={<Input type="email" />} />`

**Organisms** (complex components):
- Navigation bars, footers, forms, data tables, modals
- Example: `<Header logo={<Logo />} nav={<Nav />} search={<SearchBox />} />`

**Templates** (page layouts):
- Grid systems, content areas, sidebar layouts
- Focus on structure, not content

**Pages** (final UI):
- Real content applied to templates
- Demonstrates design system resilience

**Component Props API Design**:
```typescript
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'tertiary';
  size: 'small' | 'medium' | 'large';
  disabled?: boolean;
  loading?: boolean;
  icon?: ReactNode;
  onClick: () => void;
}
```

**Governance Model** (from 2025 best practices):
- **Design System Manager**: Overall strategy and roadmap
- **Component Library Curator**: Updates, deprecations, versioning
- **Documentation Specialist**: Guidelines, examples, API references

#### 3. Accessibility Standards (WCAG 2.1) (150 words)

**WCAG 2.1 AA Baseline** (industry standard, legal requirement):

**Contrast Requirements**:
- Normal text (< 18pt): **4.5:1** contrast ratio minimum
- Large text (≥ 18pt or 14pt bold): **3:1** contrast ratio
- UI components (buttons, form borders): **3:1** contrast ratio
- Use tools: Contrast Checker, Figma plugins (Stark, A11y)

**Keyboard Navigation**:
- All interactive elements must be keyboard accessible (Tab, Enter, Space, Esc)
- Focus indicators must be visible (outline, border, background change)
- No keyboard traps (focus can always move away)
- Logical tab order (matches visual order)

**ARIA Patterns** (W3C WAI-ARIA):
- Use semantic HTML first: `<button>`, `<nav>`, `<main>`, `<article>`
- Add ARIA when semantic HTML insufficient: `role="dialog"`, `aria-label`, `aria-describedby`
- Common patterns: modals, tabs, accordions, dropdowns, tooltips
- Test with screen readers: NVDA (Windows), VoiceOver (macOS/iOS), TalkBack (Android)

**WCAG 2.1 AAA Enhancements** (optional, specialized use):
- Enhanced contrast: **7:1** (normal text), **4.5:1** (large text)
- Target size: **44 CSS pixels** minimum (touch targets)
- Sign language interpretation, extended audio descriptions

**Testing Strategy**:
- **Automated**: axe-core, Lighthouse, WAVE (catches ~30% of issues)
- **Manual**: Keyboard-only testing, screen reader testing (catches remaining 70%)
- **CI/CD Integration**: Run accessibility tests in pipeline (fail build on violations)

#### 4. Figma MCP Workflow (150 words)

**Figma MCP for Design-to-Code Automation** (2025 latest)

Figma MCP (Model Context Protocol) enables AI-powered design token extraction and component generation:

**Setup Requirements**:
- Figma account + Personal Access Token (Settings → Account → Personal access tokens)
- Claude Code, Cursor, VS Code, or Windsurf with MCP support
- MCP server configuration (`.claude/mcp.json`)

**Workflow**:
1. **Design Token Extraction**: MCP server scans Figma file for colors, typography, spacing, shadows, gradients
2. **Export Formats**: Generate tokens in CSS, SCSS, TypeScript, JSON
3. **Component Spec Generation**: Extract component structure, variants, props from Figma frames
4. **Code Connect**: Map Figma components to code components (live sync)
5. **Natural Language Commands**: "Extract design tokens from this frame" → AI generates token JSON

**AI-Powered Features** (2025):
- **Contextual suggestions**: "This button should use `color.primary` token"
- **Standards compliance**: Validate token names against defined conventions
- **Codebase scanning**: Output structured rules file (token definitions, component libraries, style hierarchies)

**Example MCP Configuration**:
```json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "@figma/mcp-figma"],
      "env": {
        "FIGMA_PERSONAL_ACCESS_TOKEN": "your-token-here"
      }
    }
  }
}
```

**Best Practices**:
- Maintain single source of truth (Figma or code, not both)
- Use Figma as design source → Export to code (one-way sync)
- Version control tokens (Git) for change tracking
- Automate token updates in CI/CD pipeline

#### 5. Testing & Validation (120 words)

**Accessibility Testing**:
- **Automated tools**: WAVE (browser extension), Lighthouse (DevTools), axe-core (Jest integration)
- **Manual testing**: Keyboard-only navigation, screen reader testing (NVDA, VoiceOver)
- **CI/CD integration**: Fail builds on accessibility violations

**Design Consistency Audit**:
- Token usage coverage: Are all components using design tokens?
- Component variants: Do all states (hover, focus, disabled, loading) exist?
- Documentation coverage: Are all components documented in Storybook?

**Component Coverage Checklist**:
- [ ] All atomic components defined (buttons, inputs, typography)
- [ ] All molecules tested in isolation
- [ ] All organisms have accessibility tests
- [ ] All templates responsive (mobile, tablet, desktop)
- [ ] All pages pass WCAG 2.1 AA validation

**Testing Stack** (2025 best practices):
- **Unit tests**: Vitest, Jest (component logic)
- **Accessibility tests**: @axe-core/react, jest-axe
- **Visual regression**: Chromatic, Percy (detect unintended changes)
- **E2E tests**: Playwright, Cypress (user flows)

### Level 3: Advanced Patterns & Code Examples

#### Advanced Design Token Theming

**Multi-Theme Support** (light/dark/high-contrast):
```json
{
  "color": {
    "background": {
      "primary": {
        "$value": "{color.neutral.50}",
        "$type": "color",
        "$extensions": {
          "mode": {
            "light": "{color.neutral.50}",
            "dark": "{color.neutral.900}",
            "high-contrast": "#FFFFFF"
          }
        }
      }
    }
  }
}
```

**CSS Custom Properties Export** (Style Dictionary):
```css
:root {
  --color-primary: #0066CC;
  --color-neutral-50: #F9FAFB;
  --typography-heading-font-family: Inter;
  --spacing-xs: 4px;
  --spacing-sm: 8px;
}

[data-theme="dark"] {
  --color-primary: #3B82F6;
  --color-neutral-50: #111827;
}
```

#### Component API Documentation (TypeScript)

**Button Component Spec**:
```typescript
/**
 * Primary UI component for user interaction
 * @see https://design-system.example.com/button
 */
export interface ButtonProps {
  /**
   * Visual variant
   * @default 'primary'
   */
  variant: 'primary' | 'secondary' | 'tertiary' | 'ghost';
  
  /**
   * Size of button
   * @default 'medium'
   */
  size: 'small' | 'medium' | 'large';
  
  /**
   * Disabled state
   * @default false
   */
  disabled?: boolean;
  
  /**
   * Loading state with spinner
   * @default false
   */
  loading?: boolean;
  
  /**
   * Icon component (appears before text)
   */
  icon?: ReactNode;
  
  /**
   * Click handler
   */
  onClick: () => void;
  
  /**
   * Accessible label (for icon-only buttons)
   */
  'aria-label'?: string;
}
```

#### Storybook Integration

**Button.stories.tsx** (Storybook 8.0):
```typescript
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta: Meta<typeof Button> = {
  title: 'Components/Button',
  component: Button,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Primary UI component for user interaction. Supports variants, sizes, disabled, and loading states.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'tertiary', 'ghost'],
    },
    size: {
      control: 'select',
      options: ['small', 'medium', 'large'],
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

export const Disabled: Story = {
  args: {
    variant: 'primary',
    disabled: true,
    children: 'Disabled Button',
  },
};

export const Loading: Story = {
  args: {
    variant: 'primary',
    loading: true,
    children: 'Loading...',
  },
};
```

#### Accessibility Testing (Jest + axe-core)

**Button.test.tsx**:
```typescript
import { render, screen } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
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
  
  it('should have accessible name for icon-only buttons', async () => {
    render(
      <Button 
        variant="ghost" 
        icon={<CloseIcon />} 
        aria-label="Close modal"
        onClick={() => {}}
      />
    );
    
    const button = screen.getByRole('button', { name: 'Close modal' });
    expect(button).toBeInTheDocument();
  });
});
```

---

## Tool Version Matrix (2025-11-04)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **W3C DTCG Spec** | 2025.10 | Design tokens standard | ✅ Stable |
| **Style Dictionary** | 4.x | Token transformation | ✅ Current |
| **Figma MCP** | Latest | Design-to-code automation | ✅ Current |
| **Storybook** | 8.0+ | Component documentation | ✅ Current |
| **axe-core** | 4.10+ | Accessibility testing | ✅ Current |
| **Vitest** | 2.1+ | Unit testing | ✅ Current |
| **Playwright** | 1.48+ | E2E testing | ✅ Current |

---

## Best Practices

✅ **DO**:
- Use semantic token names (`primary`, `success`, `error` not `blue`, `green`, `red`)
- Start with atomic components (buttons, inputs) before complex organisms
- Test accessibility with both automated tools (30%) and manual testing (70%)
- Document all components in Storybook with live examples
- Version control design tokens in Git
- Maintain WCAG 2.1 AA baseline (4.5:1 contrast, keyboard nav)
- Use Figma MCP for automated token extraction
- Implement strong governance (manager, curator, documentation specialist)

❌ **DON'T**:
- Hard-code colors/spacing in components (use tokens)
- Skip keyboard navigation testing
- Rely only on automated accessibility testing
- Create duplicate components without governance
- Mix token formats (pick DTCG v4 or legacy, not both)
- Ignore ARIA patterns for complex components (modals, tabs, dropdowns)
- Skip documentation updates when changing components
- Use WCAG 2.0 (upgrade to 2.1)

---

## Anti-Patterns

🚫 **Token Chaos**: Hard-coded colors scattered across components
- **Solution**: Centralize all visual decisions in design tokens

🚫 **Inaccessible Components**: No keyboard support, missing ARIA labels
- **Solution**: Test with keyboard-only and screen readers

🚫 **Documentation Debt**: Outdated examples, missing props
- **Solution**: Auto-generate docs with Storybook, enforce updates in PR reviews

🚫 **Manual Token Sync**: Copy-paste from Figma to code
- **Solution**: Automate with Figma MCP + CI/CD pipeline

🚫 **Monolithic Components**: Single component with 50+ props
- **Solution**: Apply Atomic Design (split into atoms, molecules, organisms)

---

## References (Official Documentation)

### Design Tokens
- [W3C Design Tokens Spec (2025.10)](https://www.designtokens.org/tr/drafts/format/)
- [Style Dictionary v4 Documentation](https://styledictionary.com/)
- [Tokens Studio for Figma](https://tokens.studio/)

### Accessibility
- [WCAG 2.1 Guidelines](https://www.w3.org/TR/WCAG21/)
- [WAI-ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [axe-core Accessibility Testing](https://github.com/dequelabs/axe-core)

### Figma Integration
- [Figma MCP Documentation](https://www.figma.com/blog/design-systems-ai-mcp/)
- [Figma Design Tokens Plugin](https://www.figma.com/community/plugin/888356646278934516)

### Component Libraries
- [Atomic Design Methodology](https://atomicdesign.bradfrost.com/)
- [Storybook Documentation](https://storybook.js.org/)

### Tools & Testing
- [WAVE Accessibility Tool](https://wave.webaim.org/)
- [Lighthouse Accessibility Audits](https://developer.chrome.com/docs/lighthouse/)
- [Playwright E2E Testing](https://playwright.dev/)

---

## Failure Modes

### When required tools are not installed
- **Symptom**: Token transformation fails, tests don't run
- **Solution**: Install Style Dictionary, axe-core, Storybook via package manager

### When WCAG violations exist
- **Symptom**: Low contrast, missing keyboard navigation, no ARIA labels
- **Solution**: Run automated tests (Lighthouse, axe-core), manual keyboard/screen reader testing

### When design tokens are inconsistent
- **Symptom**: Hard-coded colors, duplicate spacing values, no theming support
- **Solution**: Migrate to centralized token JSON, use Style Dictionary for exports

### When components lack documentation
- **Symptom**: Developers don't know how to use components, duplicate implementations
- **Solution**: Set up Storybook, auto-generate docs, enforce documentation in PR reviews

---

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-domain-frontend` for framework-specific patterns
- Integration with `moai-foundation-trust` for quality gates (85% test coverage)
- Figma account + Personal Access Token for MCP integration

---

## Works Well With

- `moai-domain-frontend` (React/Vue/Angular component implementation)
- `moai-lang-typescript` (TypeScript component props)
- `moai-foundation-trust` (quality gates, test coverage)
- `moai-alfred-code-reviewer` (accessibility review in PRs)

---

## Changelog

- **v1.0.0** (2025-11-04): Initial release with W3C DTCG spec 2025.10, WCAG 2.1, Figma MCP integration, Atomic Design methodology, Storybook best practices

---

**End of Skill** | Created 2025-11-04
