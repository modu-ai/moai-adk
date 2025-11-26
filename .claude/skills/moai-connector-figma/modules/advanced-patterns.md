# Advanced Design System Patterns

## Design Token Architecture for Scale

### Problem: Managing Complex Token Systems

Large design systems need token hierarchies: primitive tokens → semantic tokens → component tokens.

### Solution: Token Inheritance Hierarchy

**Implementation Pattern**:
```json
{
  "color-primitive": {
    "blue-50": "#F0F9FF",
    "blue-100": "#E0F2FE",
    "blue-500": "#0EA5E9",
    "blue-900": "#0C2340"
  },
  "color-semantic": {
    "primary": "{color-primitive.blue-500}",
    "primary-hover": "{color-primitive.blue-600}",
    "primary-disabled": "{color-primitive.gray-300}",
    "success": "{color-primitive.green-500}",
    "error": "{color-primitive.red-500}",
    "warning": "{color-primitive.yellow-500}"
  },
  "color-component": {
    "button-primary": "{color-semantic.primary}",
    "button-primary-hover": "{color-semantic.primary-hover}",
    "button-primary-disabled": "{color-semantic.primary-disabled}",
    "card-background": "{color-semantic.white}",
    "card-border": "{color-semantic.gray-200}"
  }
}
```

**Benefits**:
- Single source of truth for colors
- Easy theme switching (light/dark mode)
- Consistent semantic naming
- Clear hierarchy reduces complexity
- 80% fewer color definitions needed

---

## Component Variant Matrix System

### Problem: Managing Component Variants Across States and Sizes

Components have multiple dimensions: size, variant, state, disabled state.

### Solution: 2D Variant Matrix

**Structure**:
```
Button Component Variants:
├── Sizes: sm, md, lg, xl (4)
├── Variants: primary, secondary, ghost, danger (4)
├── States: default, hover, active, focus, disabled (5)
└── Total: 4 × 4 × 5 = 80 variants

Organized in Figma:
button
  ├── sm
  │   ├── primary
  │   │   ├── default
  │   │   ├── hover
  │   │   ├── active
  │   │   ├── focus
  │   │   └── disabled
  │   ├── secondary
  │   │   └── ... (same states)
  │   └── ... (other variants)
  ├── md
  │   └── ... (same structure)
  └── lg, xl
      └── ... (same structure)
```

**Naming Convention**:
```
ComponentName/Size/Variant/State

Examples:
- Button/md/primary/default
- Button/lg/secondary/hover
- Input/md/default/focus
- Card/default/default/default (single variant)
```

---

## Component Composition Patterns

### Atomic Design Hierarchy

**Atoms** (single, reusable elements):
```
Colors, Typography, Icons, Buttons, Inputs, Labels
```

**Molecules** (2-3 atoms combined):
```
FormField = Label + Input + Error
SearchBox = Input + Icon
Card = Container + Typography
```

**Organisms** (complex components):
```
Form = Multiple FormFields + Button
Modal = Header + Body + Footer
Navigation = Logo + Menu Items + Avatar
```

**Implementation in Figma**:
- Create main components for each level
- Use nested components (atoms within molecules)
- Document component hierarchy
- Define variant strategy per level

---

## Design System Versioning Strategy

### Semantic Versioning for Design Systems

```
v{MAJOR}.{MINOR}.{PATCH}

MAJOR: Breaking changes (component API changes)
MINOR: New components or features (backward compatible)
PATCH: Bug fixes, token adjustments (no API changes)

Examples:
v2.0.0 → Button component API changes (breaking)
v2.1.0 → New Badge component added (feature)
v2.1.1 → Button hover color adjusted (patch)
```

### Version Management Workflow

1. **Feature Branch**: `design-system/button-api-redesign`
2. **Design Review**: Team approval
3. **Token Export**: Generate token files
4. **Version Tag**: `git tag v3.0.0`
5. **Release Notes**: Document changes
6. **Code Generation**: Update React components
7. **Storybook Update**: Publish new version

---

## Dark Mode Implementation

### Token Strategy for Dark Mode

**CSS Variable Generation**:
```css
/* Light theme (default) */
:root {
  --color-primary: #0EA5E9;
  --color-background: #FFFFFF;
  --color-text: #1F2937;
  --color-border: #E5E7EB;
}

/* Dark theme */
[data-theme="dark"] {
  --color-primary: #06B6D4;
  --color-background: #0F172A;
  --color-text: #F1F5F9;
  --color-border: #334155;
}
```

**Figma Implementation**:
1. Create token groups: `light/` and `dark/`
2. Map same token names to different values
3. Export as CSS variables
4. Use in components: `background: var(--color-background)`

---

## Accessibility-First Component Design

### Component Accessibility Checklist

**Color Contrast**:
- WCAG AA: 4.5:1 for text, 3:1 for UI
- WCAG AAA: 7:1 for text, 4.5:1 for UI

**Interactive States**:
- Visible focus indicator (≥3px)
- Touch targets: ≥44x44px
- Clear hover states
- Disabled state visually distinct

**Keyboard Navigation**:
- Tab order logical
- All controls keyboard accessible
- Escape key to close modals
- Arrow keys for menus

**Implementation Example**:
```jsx
// Accessible Button Component
<button
  className="button"
  aria-pressed={isActive}
  aria-label="Toggle sidebar"
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      handleClick();
    }
  }}
>
  Menu
</button>

/* CSS with accessibility focus */
.button {
  min-height: 44px;      /* Touch target */
  min-width: 44px;
  color: var(--color-text);
  background: var(--color-button-bg);
}

.button:focus-visible {
  outline: 3px solid var(--color-focus);  /* Visible focus */
  outline-offset: 2px;
}

.button:disabled {
  opacity: 0.5;          /* Visually distinct */
  cursor: not-allowed;
}
```

---

## Figma Plugin Automation

### Custom Plugin for Token Synchronization

```javascript
// Plugin: Auto-sync design tokens to code
figma.showUI(__html__);

figma.ui.onmessage = (msg) => {
  if (msg.type === 'EXPORT_TOKENS') {
    // Get all color variables
    const colors = figma.variables.getLocalVariables()
      .filter(v => v.group === 'Color');

    // Generate token JSON
    const tokenJSON = {
      color: {}
    };

    colors.forEach(token => {
      // Extract RGB values
      const color = token.valuesByMode.mode1;
      tokenJSON.color[token.name] = rgbToHex(color);
    });

    // Send to code generation service
    figma.ui.postMessage({
      type: 'TOKENS_READY',
      data: tokenJSON
    });
  }
};
```

### Plugin Benefits
- Auto-export token changes
- Sync with code repositories
- Validate token usage
- Check accessibility compliance
- Generate component documentation

---

## Design-to-Dev Handoff Process

### Code Connect Integration

**Component Metadata**:
```javascript
// figma-metadata.json
{
  "Button": {
    "component": "src/components/Button.tsx",
    "props": [
      {
        "name": "variant",
        "figmaProperty": "Variant",
        "type": "enum",
        "values": ["primary", "secondary", "ghost"]
      },
      {
        "name": "size",
        "figmaProperty": "Size",
        "type": "enum",
        "values": ["sm", "md", "lg"]
      },
      {
        "name": "disabled",
        "figmaProperty": "State/disabled",
        "type": "boolean"
      }
    ]
  }
}
```

**Automated Storybook Generation**:
```jsx
// Auto-generated from Figma metadata
export default {
  title: 'Components/Button',
  component: Button,
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'ghost']
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg']
    },
    disabled: { control: 'boolean' }
  }
};

export const Primary = {
  args: { variant: 'primary', size: 'md' }
};
export const Secondary = {
  args: { variant: 'secondary', size: 'md' }
};
```

---

## Performance Optimization for Large Design Systems

### File Organization Strategy

**Problem**: Large Figma files become slow with 1000+ components.

**Solution**: File-per-feature organization
```
Design System
├── 01-tokens.figma (Design tokens only)
├── 02-atoms.figma (Button, Input, Label)
├── 03-molecules.figma (FormField, Card, SearchBox)
├── 04-organisms.figma (Form, Modal, Navigation)
└── 05-templates.figma (Page layouts)
```

**Benefits**:
- Faster file loading
- Easier collaboration (less conflict)
- Clearer organization
- Better team ownership

### Asset Management

```
Assets
├── Icons (1000+ SVGs)
├── Illustrations (100+ assets)
├── Photography (brand guidelines)
└── Patterns (backgrounds, textures)
```

Use Figma Assets panel or external asset library (Figma integrations).

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
