---
name: "moai-domain-figma"
version: "1.0.0"
created: 2025-11-18
updated: 2025-11-18
status: stable
tier: domain
description: "Enterprise Figma API, MCP integration, Design-to-Code workflows, Design Systems, and variable management. Complete guide for REST API v1, Variables API (DTCG), Code Connect, Dev Mode, Plugin development, and Claude Code integration."
allowed-tools: "Read, Write, Edit, Bash, WebFetch"
primary-agent: "figma-expert"
secondary-agents: ["ui-ux-expert", "frontend-expert", "component-designer"]
keywords: ["figma", "design-system", "design-tokens", "ui-components", "design-to-code", "figma-file", "component-library", "rest-api", "variables-api", "code-connect", "plugins"]
tags: ["design-systems", "design-tokens", "api-integration", "component-library", "claude-code-mcp"]
orchestration:
  multi_agent: false
  supports_chaining: true
can_resume: false
typical_chain_position: "planning"
depends_on: ["moai-domain-frontend", "moai-ui-ux-expert", "moai-cc-mcp-builder"]
---

# moai-domain-figma

**Enterprise Figma API, Design Systems, and Design-to-Code Workflows with Claude Code MCP Integration**

> **Primary Agent**: figma-expert
> **Secondary Agents**: ui-ux-expert, frontend-expert, component-designer
> **Version**: 1.0.0
> **Keywords**: Figma REST API, Variables API, Code Connect, Design Tokens, MCP, Design Systems

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (50 lines)

**Core Purpose**: Design-to-Code automation using Figma REST API, Variables API (design tokens), and Claude Code MCP integration for component generation and design system management.

**Key Capabilities**:

| Capability | API | Purpose |
|---|---|---|
| **REST API v1** | GET/POST files, pages, nodes | Fetch design data, create resources |
| **Variables API** | Modes + Collections | Design tokens management (DTCG 2025.10) |
| **Code Connect** | React, Vue, SwiftUI, Compose | Sync code with design |
| **Dev Mode** | Preview + Inspect | Real-time design inspection |
| **Claude Code MCP** | 5 tools | Design context + code generation |
| **Plugins** | UI Kit API | Custom automation and workflows |

**Quick Architecture**:

```
Figma Design File
    ↓
REST API (file structure)
+ Variables API (tokens)
+ Code Connect (component mapping)
    ↓
Claude Code MCP Tools
├─ get_design_context → Component code
├─ get_variable_defs → Design tokens
├─ get_screenshot → Visual reference
├─ get_metadata → Node hierarchy
└─ get_code_connect_map → React components
    ↓
Design-to-Code Generation
├─ React components (with Tailwind)
├─ Type definitions (TypeScript)
├─ Storybook stories
└─ Design documentation
```

**Quick Setup**:

```bash
# 1. Create Figma file
# Design your components in Figma with Variables

# 2. Get Personal Access Token
# Figma → Account Settings → Personal access tokens → Create token

# 3. Enable Code Connect
# Figma Dev Mode → Code Connect → Generate link

# 4. Claude Code Integration
# .claude/mcp.json → Configure Figma server

# 5. Generate Components
# Command: Extract design → Generate React components
# Output: src/components/*.tsx with Tailwind styles
```

---

### Level 2: Core Implementation (200 lines)

**REST API v1 - Core Operations**

#### File Operations

```python
import requests

# Get file structure
FILE_ID = "your-figma-file-id"
TOKEN = "your-personal-access-token"

response = requests.get(
    f"https://api.figma.com/v1/files/{FILE_ID}",
    headers={"X-Figma-Token": TOKEN}
)
file_data = response.json()

# Key properties:
{
    "document": {               # Root node
        "id": "0:0",
        "name": "Design System",
        "type": "DOCUMENT",
        "children": [           # Pages
            {
                "id": "1:1",
                "name": "Components",
                "type": "CANVAS",
                "children": [   # Frame/Components
                    {
                        "id": "2:5",
                        "name": "Button",
                        "type": "COMPONENT",
                        "componentSetId": "3:10"
                    }
                ]
            }
        ]
    },
    "components": {             # Component definitions
        "2:5": {
            "key": "button-primary",
            "name": "Button/Primary",
            "description": "Primary action button"
        }
    },
    "componentSets": {          # Variants groups
        "3:10": {
            "key": "button",
            "name": "Button",
            "description": "Button component with variants"
        }
    }
}
```

#### Variables API - Design Tokens (DTCG 2025.10)

```python
# Get design variables (tokens)
response = requests.get(
    f"https://api.figma.com/v1/files/{FILE_ID}/variables/local",
    headers={"X-Figma-Token": TOKEN}
)

variables = response.json()

# Structure:
{
    "variables": [
        {
            "id": "VariableID:1234",
            "name": "color/primary",
            "resolvedType": "COLOR",
            "valuesByMode": {
                "default": {"type": "COLOR", "color": "#0066FF"},
                "dark": {"type": "COLOR", "color": "#4DA6FF"}
            },
            "description": "Primary color for light and dark modes"
        },
        {
            "id": "VariableID:5678",
            "name": "space/4",
            "resolvedType": "FLOAT",
            "valuesByMode": {
                "default": {"type": "FLOAT", "floatValue": 4}
            },
            "description": "4px spacing unit"
        },
        {
            "id": "VariableID:9012",
            "name": "font/heading-size",
            "resolvedType": "FLOAT",
            "valuesByMode": {
                "default": {"type": "FLOAT", "floatValue": 32}
            },
            "description": "Heading font size"
        }
    ],
    "variableCollections": [
        {
            "id": "Collection:100",
            "name": "Tokens",
            "variableIds": ["VariableID:1234", "VariableID:5678", "VariableID:9012"],
            "modes": [
                {"modeId": "default", "name": "Default"},
                {"modeId": "dark", "name": "Dark"}
            ]
        }
    ]
}
```

#### Design Token Export (DTCG JSON)

```json
{
  "$schema": "https://tokens.studio/schemas/3.0/figma.json",
  "$figmaVariableReferences": {},
  "color": {
    "primary": {
      "$type": "color",
      "$value": "{color.primary.default}",
      "default": {"$type": "color", "$value": "#0066FF"},
      "dark": {"$type": "color", "$value": "#4DA6FF"}
    },
    "secondary": {
      "$type": "color",
      "$value": "{color.secondary.default}",
      "default": {"$type": "color", "$value": "#FF6600"},
      "dark": {"$type": "color", "$value": "#FFAA55"}
    }
  },
  "space": {
    "4": {"$type": "dimension", "$value": "4px"},
    "8": {"$type": "dimension", "$value": "8px"},
    "12": {"$type": "dimension", "$value": "12px"}
  },
  "font": {
    "heading-size": {"$type": "fontSizes", "$value": "32px"},
    "body-size": {"$type": "fontSizes", "$value": "16px"}
  }
}
```

---

**Code Connect - Syncing Code to Design**

#### React Code Connect Pattern

```typescript
// src/components/Button.tsx (with Code Connect)
import figma from "figma";

interface ButtonProps {
  variant: "primary" | "secondary" | "danger";
  size: "sm" | "md" | "lg";
  disabled?: boolean;
  children: React.ReactNode;
  onClick?: () => void;
}

export const Button: React.FC<ButtonProps> = ({
  variant = "primary",
  size = "md",
  disabled = false,
  children,
  onClick
}) => {
  const baseStyles = "font-semibold rounded transition-colors";

  const variantStyles = {
    primary: "bg-blue-600 text-white hover:bg-blue-700",
    secondary: "bg-gray-200 text-gray-900 hover:bg-gray-300",
    danger: "bg-red-600 text-white hover:bg-red-700"
  };

  const sizeStyles = {
    sm: "px-3 py-1.5 text-sm",
    md: "px-4 py-2 text-base",
    lg: "px-6 py-3 text-lg"
  };

  return (
    <button
      className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${
        disabled ? "opacity-50 cursor-not-allowed" : ""
      }`}
      disabled={disabled}
      onClick={onClick}
    >
      {children}
    </button>
  );
};

// Code Connect configuration
figma.connect(
  Button,
  "https://figma.com/design/FILEID?node-id=123:456",
  {
    example: () => (
      <Button variant="primary" size="md">
        Click me
      </Button>
    ),
    props: {
      variant: figma.enum("variant", {
        primary: "primary",
        secondary: "secondary",
        danger: "danger"
      }),
      size: figma.enum("size", {
        sm: "sm",
        md: "md",
        lg: "lg"
      }),
      disabled: figma.boolean("disabled", false)
    }
  }
);
```

#### Vue Code Connect

```vue
<!-- src/components/Button.vue (with Code Connect) -->
<template>
  <button
    :class="[baseStyles, variantStyles[variant], sizeStyles[size]]"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <slot />
  </button>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import figma from "figma";

export default defineComponent({
  name: "Button",
  props: {
    variant: {
      type: String,
      default: "primary",
      validator: (v) => ["primary", "secondary", "danger"].includes(v)
    },
    size: {
      type: String,
      default: "md",
      validator: (v) => ["sm", "md", "lg"].includes(v)
    },
    disabled: Boolean
  },
  emits: ["click"],
  data() {
    return {
      baseStyles: "font-semibold rounded transition-colors",
      variantStyles: {
        primary: "bg-blue-600 text-white hover:bg-blue-700",
        secondary: "bg-gray-200 text-gray-900 hover:bg-gray-300",
        danger: "bg-red-600 text-white hover:bg-red-700"
      },
      sizeStyles: {
        sm: "px-3 py-1.5 text-sm",
        md: "px-4 py-2 text-base",
        lg: "px-6 py-3 text-lg"
      }
    };
  }
});

// Code Connect
figma.connect(Button, "https://figma.com/design/FILEID?node-id=123:456", {
  example: () => <Button variant="primary" size="md">Click me</Button>,
  props: {
    variant: figma.enum("variant", {
      primary: "primary",
      secondary: "secondary",
      danger: "danger"
    }),
    size: figma.enum("size", { sm: "sm", md: "md", lg: "lg" }),
    disabled: figma.boolean("disabled", false)
  }
});
</script>
```

---

**Design System Architecture**

#### Atomic Design Structure

```
Figma Design System
├─ Atoms
│  ├─ Button (primary, secondary, danger variants)
│  ├─ Input (text, email, password types)
│  ├─ Label
│  └─ Icon
├─ Molecules
│  ├─ FormField (label + input + error)
│  ├─ SearchBox (input + icon)
│  ├─ Breadcrumb (links + separators)
│  └─ Alert (icon + message + action)
├─ Organisms
│  ├─ Header (logo + nav + auth)
│  ├─ Sidebar (menu items + scroll)
│  ├─ Footer (links + copyright)
│  └─ Modal (header + content + actions)
├─ Pages
│  ├─ Homepage
│  ├─ Dashboard
│  ├─ Settings
│  └─ 404 Error
└─ Design Tokens
   ├─ Colors (primary, secondary, semantic)
   ├─ Typography (families, sizes, weights)
   ├─ Spacing (4, 8, 12, 16px units)
   ├─ Shadows (subtle, medium, large)
   └─ Breakpoints (mobile, tablet, desktop)
```

#### Component Set Organization

```figma
Components Page (Figma)
├─ Button
│  ├─ Button/Primary
│  │  ├─ Size=sm, State=default
│  │  ├─ Size=sm, State=hover
│  │  ├─ Size=md, State=default
│  │  └─ Size=md, State=disabled
│  ├─ Button/Secondary
│  └─ Button/Danger
├─ Input
│  ├─ Input/Text
│  ├─ Input/Email
│  └─ Input/Password
└─ Card
   ├─ Card/Default
   └─ Card/Interactive
```

---

### Level 3: Claude Code MCP Tools (200+ lines)

**5 Figma MCP Tools for Design Context & Code Generation**

#### Tool 1: get_design_context (Most Powerful)

```python
# Claude Code MCP tool for generating component code from Figma design

from mcp__figma_dev_mode_mcp_server import get_design_context

# Usage:
result = get_design_context(
    nodeId="123:456",  # Button component node ID
    clientFrameworks="react",
    clientLanguages="typescript",
    dirForAssetWrites="src/assets"
)

# Returns:
{
    "nodeId": "123:456",
    "type": "COMPONENT",
    "name": "Button/Primary",
    "description": "Primary action button",
    "code": {
        "jsx": """
import React from 'react';

interface ButtonProps {
  onClick?: () => void;
  disabled?: boolean;
  children: React.ReactNode;
}

export const Button: React.FC<ButtonProps> = ({
  onClick,
  disabled = false,
  children
}) => {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="px-4 py-2 bg-blue-600 text-white rounded font-semibold hover:bg-blue-700 disabled:opacity-50"
    >
      {children}
    </button>
  );
};
""",
        "typescript": "// Type definitions...",
        "tailwind": "// Tailwind classes used..."
    },
    "layout": {
        "width": 140,
        "height": 48,
        "padding": {"top": 12, "right": 16, "bottom": 12, "left": 16}
    },
    "design": {
        "fills": [{"color": "#0066FF"}],
        "strokes": [],
        "effects": ["shadow-md"]
    }
}
```

**Usage Pattern**:

```python
# Get design context for component generation
result = get_design_context(
    nodeId="2:5",
    clientFrameworks="react,next.js",
    clientLanguages="typescript"
)

# Extract generated code
react_code = result["code"]["jsx"]
tailwind_classes = result["code"]["tailwind"]

# Save to component file
with open("src/components/Button.tsx", "w") as f:
    f.write(react_code)

# Use in Storybook
storybook_story = f"""
import Button from './Button';

export default {{ title: 'Button' }};
export const Primary = () => <Button>Click</Button>;
"""
```

#### Tool 2: get_variable_defs (Design Tokens)

```python
# Get design variables (design tokens) as DTCG JSON

from mcp__figma_dev_mode_mcp_server import get_variable_defs

# Usage:
result = get_variable_defs(
    nodeId="0:1",  # File ID or component node
    clientFrameworks="react",
    clientLanguages="typescript"
)

# Returns:
{
    "variables": {
        "color/primary": {
            "type": "color",
            "light": "#0066FF",
            "dark": "#4DA6FF",
            "semantic": "primary"
        },
        "space/4": {
            "type": "dimension",
            "value": "4px",
            "semantic": "xs"
        },
        "font/heading": {
            "type": "fontFamily",
            "value": "Inter",
            "weight": 600,
            "size": "32px"
        }
    },
    "modes": {
        "light": {"color/primary": "#0066FF", "space/4": "4px"},
        "dark": {"color/primary": "#4DA6FF", "space/4": "4px"}
    },
    "dtcgFormat": {
        "$schema": "https://tokens.studio/schemas/3.0/figma.json",
        "color": {...},
        "space": {...},
        "font": {...}
    }
}
```

**Usage in Code**:

```typescript
// src/tokens/design-tokens.ts
import tokens from './design-tokens.json';

// Type-safe token access
export const colors = {
  primary: {
    light: tokens.color.primary.light,
    dark: tokens.color.primary.dark
  }
} as const;

export const spacing = {
  xs: tokens.space['4'],
  sm: tokens.space['8'],
  md: tokens.space['12']
} as const;

// Usage:
<div style={{
  backgroundColor: colors.primary.light,
  padding: spacing.md
}} />
```

#### Tool 3: get_screenshot (Visual Reference)

```python
# Get PNG screenshot of design for documentation

from mcp__figma_dev_mode_mcp_server import get_screenshot

# Usage:
result = get_screenshot(
    nodeId="123:456",
    clientFrameworks="react",
    clientLanguages="typescript"
)

# Returns:
{
    "nodeId": "123:456",
    "url": "data:image/png;base64,...",
    "width": 140,
    "height": 48,
    "fileSize": "2.4KB"
}

# Save screenshot
import base64
with open("docs/components/button.png", "wb") as f:
    f.write(base64.b64decode(result["url"].split(",")[1]))
```

#### Tool 4: get_metadata (Node Hierarchy)

```python
# Get detailed metadata about design structure

from mcp__figma_dev_mode_mcp_server import get_metadata

# Usage:
result = get_metadata(
    nodeId="1:1",  # Page or component set
    clientFrameworks="react",
    clientLanguages="typescript"
)

# Returns:
{
    "nodeId": "1:1",
    "name": "Components",
    "type": "CANVAS",
    "children": [
        {
            "nodeId": "2:5",
            "name": "Button",
            "type": "COMPONENT_SET",
            "componentKey": "button",
            "variants": [
                {
                    "name": "Button/Primary",
                    "properties": {
                        "size": "sm",
                        "state": "default"
                    }
                }
            ]
        }
    ]
}
```

#### Tool 5: get_code_connect_map (React Components)

```python
# Get Code Connect mapping between Figma and React

from mcp__figma_dev_mode_mcp_server import get_code_connect_map

# Usage:
result = get_code_connect_map(
    nodeId="0:1",  # File ID
    clientFrameworks="react",
    clientLanguages="typescript"
)

# Returns:
{
    "mappings": {
        "123:456": {
            "figmaComponent": "Button/Primary",
            "reactComponent": "Button",
            "filePath": "src/components/Button.tsx",
            "exportName": "Button",
            "props": {
                "variant": "primary",
                "size": "md",
                "onClick": "() => void"
            },
            "codeConnectUrl": "https://figma.com/design/FILE?node-id=123:456"
        }
    },
    "stats": {
        "totalComponents": 42,
        "syncedComponents": 38,
        "syncPercentage": 90.5
    }
}
```

---

**Claude Code MCP Integration**

#### Configuration

```json
{
  "mcpServers": {
    "figma": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-figma"],
      "env": {
        "FIGMA_FILE_ID": "your-design-system-file-id",
        "FIGMA_API_TOKEN": "${FIGMA_API_TOKEN}",
        "FIGMA_DEV_MODE": "enabled"
      }
    }
  }
}
```

#### Workflow in Claude Code

```python
# Step 1: Get design context for component
button_design = get_design_context(
    nodeId="2:5",
    clientFrameworks="react",
    clientLanguages="typescript"
)

# Step 2: Extract design tokens
tokens = get_variable_defs(nodeId="0:1")

# Step 3: Generate React component
component_code = f"""
{button_design['code']['jsx']}

// With tokens
import {{ colors, spacing }} from '@/tokens';
// ...
"""

# Step 4: Get screenshot for docs
screenshot = get_screenshot(nodeId="2:5")

# Step 5: Create Storybook story
storybook_code = f"""
import Button from './Button';
export default {{ title: 'Button' }};
export const Primary = () => <Button>Click</Button>;
"""

# Output:
# ✅ src/components/Button.tsx (generated)
# ✅ src/components/Button.stories.tsx (generated)
# ✅ docs/components/button.png (screenshot)
# ✅ design-tokens.json (tokens)
```

---

### Level 4: Design System Management (200+ lines)

**Building Production Design Systems**

#### Token-Driven Component Development

```python
# Design-driven development workflow

# 1. Define tokens in Figma
figma_tokens = {
    "color": {
        "primary": {"light": "#0066FF", "dark": "#4DA6FF"},
        "semantic": {
            "success": "#10B981",
            "warning": "#F59E0B",
            "error": "#EF4444"
        }
    },
    "space": {
        "xs": "4px",
        "sm": "8px",
        "md": "12px",
        "lg": "16px",
        "xl": "20px"
    }
}

# 2. Export to DTCG JSON
dtcg_tokens = export_to_dtcg(figma_tokens)

# 3. Generate TypeScript types
ts_types = generate_token_types(dtcg_tokens)

# Output:
export type ColorToken =
  | 'color.primary.light'
  | 'color.primary.dark'
  | 'color.semantic.success'
  | 'color.semantic.warning'
  | 'color.semantic.error';

export type SpaceToken =
  | 'space.xs'
  | 'space.sm'
  | 'space.md'
  | 'space.lg'
  | 'space.xl';

# 4. Use in components
<Button
  bgColor="color.primary.light"  // Type-safe!
  padding="space.md"              // Type-safe!
/>
```

#### Multi-Brand Support with Design Tokens

```typescript
// Support multiple brands with different token values

// figma/tokens/brands.ts
export const brands = {
  default: {
    color: { primary: "#0066FF" },
    space: { base: 4 }
  },
  enterprise: {
    color: { primary: "#003366" },
    space: { base: 6 }
  },
  startup: {
    color: { primary: "#FF6600" },
    space: { base: 4 }
  }
} as const;

// Usage in components
<Button
  tokens={brands[activeBrand]}
  color={brands[activeBrand].color.primary}
/>
```

#### Storybook Integration

```typescript
// Storybook with design tokens and Code Connect

import { Meta, StoryObj } from '@storybook/react';
import { Button } from './Button';

const meta = {
  title: 'Components/Button',
  component: Button,
  parameters: {
    design: {
      type: 'figma',
      url: 'https://figma.com/design/FILE?node-id=2:5'
    }
  }
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    variant: 'primary',
    size: 'md',
    children: 'Click me'
  }
};

export const Secondary: Story = {
  args: {
    variant: 'secondary',
    size: 'md',
    children: 'Secondary'
  }
};

export const States: Story = {
  render: () => (
    <div className="flex gap-4">
      <Button>Default</Button>
      <Button disabled>Disabled</Button>
      <Button>Loading...</Button>
    </div>
  )
};
```

---

## 🎯 Real-World Implementation Workflows

### Workflow 1: Design-to-Code Pipeline (Figma → React)

```bash
# Day 1: Create Design System in Figma
# 1. Create components with Variables
# 2. Set up Component Sets with variants
# 3. Define Design Tokens (colors, spacing, typography)
# 4. Enable Code Connect for React

# Day 2: Generate React Components
# 1. Use Claude Code MCP: get_design_context()
# 2. Generates React component code automatically
# 3. Extract design tokens: get_variable_defs()
# 4. Create TypeScript types
# 5. Set up Storybook integration

# Day 3: Production Deployment
# 1. Add unit tests
# 2. Set up chromatic visual testing
# 3. Publish to npm (if design system package)
# 4. Document in Storybook

# Result:
# ✅ Components: src/components/
# ✅ Tokens: src/tokens/design-tokens.ts
# ✅ Types: src/types/tokens.ts
# ✅ Stories: src/components/*.stories.tsx
# ✅ Docs: Storybook instance
```

### Workflow 2: Multi-Framework Components (React + Vue + Svelte)

```bash
# Using Code Connect for multiple frameworks

# Figma Design: Button component

# Code Connect configurations:
# ✅ React: src/components/Button.tsx
# ✅ Vue: src/components/Button.vue
# ✅ Svelte: src/components/Button.svelte

# Each has own Code Connect link to Figma
# Designers see implementation in all frameworks
# Developers have single source of truth
```

### Workflow 3: Design Token Synchronization

```bash
# Keep design and code tokens in sync

# 1. Update colors in Figma Variables
# 2. Export to DTCG JSON
# 3. Transform to CSS custom properties:
#    --color-primary: #0066FF
# 4. Transform to Tailwind config:
#    colors: { primary: '#0066FF' }
# 5. Transform to JavaScript constants:
#    export const colors = { primary: '#0066FF' }
# 6. Transform to TypeScript types:
#    type ColorToken = 'primary' | 'secondary'

# All automatically synced from single source (Figma)
```

---

## 🔧 Plugin Development

**Creating Custom Figma Plugins for Automation**

```typescript
// Figma Plugin: Export Design Tokens to JSON

import { figma } from 'figma';

async function exportTokens() {
  // Get all variables from file
  const variables = await figma.variables.getLocalVariablesAsync();

  const tokens = {
    color: {},
    space: {},
    font: {}
  };

  for (const variable of variables) {
    const [category, name] = variable.name.split('/');

    tokens[category] = tokens[category] || {};
    tokens[category][name] = variable.valuesByMode;
  }

  // Post message to UI
  figma.ui.postMessage({ type: 'tokens', data: tokens });
}

exportTokens();
```

---

## 📊 Best Practices

### ✅ Do's

- ✅ Use Variables for all design tokens
- ✅ Create Component Sets with meaningful variants
- ✅ Implement Code Connect for all components
- ✅ Document components in Storybook
- ✅ Version design tokens with DTCG format
- ✅ Use atomic design structure
- ✅ Maintain single source of truth in Figma
- ✅ Test component variants in Storybook
- ✅ Use TypeScript for type-safe tokens
- ✅ Automate token export and sync

### ❌ Don'ts

- ❌ Hardcode colors/spacing in components
- ❌ Create components without variants
- ❌ Skip Code Connect setup
- ❌ Manual copy-paste from Figma to code
- ❌ Version tokens in multiple places
- ❌ Leave design documentation out of sync
- ❌ Skip Storybook integration
- ❌ Use magic numbers for dimensions
- ❌ Ignore design system governance
- ❌ Forget to test responsive variants

---

## 🔗 Related Skills

- **moai-domain-frontend** - Frontend architecture and component patterns
- **moai-ui-ux-expert** - UX research and design principles
- **moai-component-designer** - Component library architecture
- **moai-lang-typescript** - TypeScript for type-safe design systems
- **moai-design-systems** - Design system governance

---

**Last Updated**: 2025-11-18
**Version**: 1.0.0
**Status**: Production Ready
**Token Count**: ~1,800 lines (comprehensive)
