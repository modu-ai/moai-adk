# UI/UX Plugin

**Design automation with Figma MCP** — 7 agents, design-to-code conversion, shadcn/ui components, Tailwind CSS theming, accessibility auditing.

## 🎯 What It Does

Automate design-to-code workflow and design system management:

```bash
/plugin install moai-plugin-uiux
```

**Automatically provides**:
- 🎨 Figma design file access and API integration
- 💻 Design-to-code conversion and CSS generation
- 🧩 shadcn/ui component library with customization
- 🎭 Tailwind CSS design tokens and theming
- ♿ WCAG accessibility auditing and compliance
- 📚 Design documentation and component library generation

## 🏗️ Architecture

### 7 Specialist Agents

| Agent | Role | When to Use |
|-------|------|------------|
| **Figma Designer** | Design file access, collaboration | Managing Figma files |
| **Design Strategist** | Design system strategy | Planning design systems |
| **Design System Architect** | System establishment and governance | Building design systems |
| **Component Builder** | Component development and customization | Building components |
| **CSS/HTML Generator** | Code generation from designs | Converting designs to code |
| **Accessibility Specialist** | WCAG compliance auditing | Accessibility testing |
| **Design Documentation Writer** | Component library documentation | Creating docs |

### 6 Skills

1. **moai-design-figma-mcp** — Figma API, design collaboration, webhooks
2. **moai-design-figma-to-code** — Design-to-code pipelines, asset extraction
3. **moai-design-shadcn-ui** — Component customization, Radix UI integration
4. **moai-design-tailwind-v4** — Utility-first CSS, design tokens, theming
5. **moai-lang-tailwind-shadcn** — Integration patterns, component theming
6. **moai-domain-frontend** — Frontend patterns, accessibility, performance

## ⚡ Quick Start

### Installation

```bash
/plugin install moai-plugin-uiux
```

### MCP Server Configuration

This plugin requires Figma API access:

**Required Environment Variables**:
```bash
FIGMA_API_TOKEN=<your-figma-token>
```

Get your token at: https://www.figma.com/developers/api#access-tokens

### Use with MoAI-ADK

The UI/UX plugin provides agents for:

1. **Design system** - Design Strategist plans structure
2. **Components** - Component Builder creates components
3. **Code generation** - CSS/HTML Generator converts designs
4. **Accessibility** - Accessibility Specialist audits compliance
5. **Documentation** - Documentation Writer creates guides

## 📊 Design-to-Code Workflow

```
Figma Design File
    ↓
[Figma Designer]
├─ Access design tokens
├─ Extract components
└─ Get design specs
    ↓
[Design System Architect]
├─ Analyze design system
├─ Create token mapping
└─ Plan component architecture
    ↓
[CSS/HTML Generator]
├─ Generate HTML structure
├─ Create Tailwind classes
└─ Export to React/Next.js
    ↓
[Component Builder]
├─ Create shadcn/ui components
├─ Customize variants
└─ Add theme support
    ↓
[Accessibility Specialist]
├─ Audit for WCAG compliance
├─ Test keyboard navigation
└─ Verify color contrast
    ↓
[Design Documentation Writer]
├─ Create component docs
├─ Write usage guidelines
└─ Generate Storybook
```

## 🎨 Features

### Figma Integration
- **Design File Access** - Read Figma files via API
- **Component Sync** - Keep designs and code in sync
- **Design Tokens** - Extract colors, typography, spacing
- **Webhooks** - Real-time design updates
- **Collaboration** - Share feedback in Figma

### Design-to-Code
- **Automatic CSS** - Generate Tailwind classes
- **Component Export** - Convert designs to React
- **Asset Extraction** - Export images and icons
- **Code Formatting** - Consistent code output
- **Type Generation** - TypeScript interfaces

### Component Library
- **shadcn/ui Base** - Built on accessible primitives
- **Customizable** - Extend components for your needs
- **Tailwind Integration** - Design token support
- **Dark Mode** - Automatic theme variants
- **Responsive** - Mobile-first by default

### Design System
- **Token Definition** - Colors, typography, spacing
- **Variant Management** - Size, state, theme variants
- **Documentation** - Auto-generated guides
- **Storybook** - Interactive component explorer
- **Version Control** - Git-based design history

### Accessibility
- **WCAG 2.1** - AA and AAA compliance
- **Keyboard Navigation** - Full keyboard support
- **Color Contrast** - Automatic contrast checking
- **ARIA Attributes** - Semantic markup
- **Screen Reader Testing** - Automated testing

## 🚀 Typical Design System Workflow

### Phase 1: Design Planning

```
Create Figma file
    ↓
[Design Strategist]
├─ Define design system principles
├─ Plan component hierarchy
└─ Establish naming conventions
    ↓
[Design System Architect]
├─ Create token structure
├─ Define color palette
└─ Setup typography scale
```

### Phase 2: Component Design

```
Design components in Figma
    ├─ Button (with variants)
    ├─ Input (with states)
    ├─ Card (with layouts)
    └─ Modal (with animations)
    ↓
[Figma Designer]
├─ Organize component library
├─ Create component variations
└─ Document design specs
```

### Phase 3: Code Generation

```
Export designs
    ↓
[CSS/HTML Generator]
├─ Generate HTML templates
├─ Create Tailwind classes
└─ Export assets
    ↓
[Component Builder]
├─ Create React components
├─ Add TypeScript types
└─ Setup Storybook stories
```

### Phase 4: Accessibility & Docs

```
Review components
    ↓
[Accessibility Specialist]
├─ Test keyboard navigation
├─ Check color contrast
└─ Verify ARIA labels
    ↓
[Documentation Writer]
├─ Create usage guidelines
├─ Write code examples
└─ Generate component catalog
```

## 📚 Skills Explained

### moai-design-figma-mcp
Figma platform automation:
- **API Access** - Read Figma files programmatically
- **Design Tokens** - Extract colors, typography, spacing
- **Components** - Access Figma component library
- **Webhooks** - Real-time design change notifications
- **Collaboration** - Comment and feedback integration

### moai-design-figma-to-code
Design-to-code conversion:
- **HTML Generation** - Create semantic HTML
- **CSS Generation** - Tailwind class generation
- **Asset Export** - Image and icon export
- **Component Mapping** - Design to React mapping
- **Code Formatting** - Consistent code output

### moai-design-shadcn-ui
shadcn/ui integration:
- **Component Customization** - Tailored variants
- **Radix UI Primitives** - Accessible base components
- **Dark Mode** - Theme switching support
- **Composition** - Complex component patterns
- **Accessibility** - Built-in WCAG compliance

### moai-design-tailwind-v4
Tailwind CSS theming:
- **Design Tokens** - CSS custom properties
- **Color System** - Comprehensive color palettes
- **Typography Scale** - Font sizing hierarchy
- **Spacing System** - Consistent spacing
- **Theme Customization** - Brand customization

### moai-lang-tailwind-shadcn
Integration patterns:
- **Theme Configuration** - Tailwind + shadcn
- **Component Theming** - Dark mode variants
- **Custom Components** - Extend shadcn/ui
- **Token Mapping** - Design tokens to CSS
- **Responsive Design** - Mobile-first approach

## 🔗 Integration with Other Plugins

- **Frontend Plugin**: Build React components with designs
- **Backend Plugin**: API integration for design content
- **DevOps Plugin**: Deploy design systems to production

## 🔧 Common Design Tasks

### Create Design System
1. Plan system structure in Figma
2. Define design tokens
3. Create component library
4. Generate documentation

### Build Components
1. Design in Figma
2. Extract design specifications
3. Create React components
4. Audit accessibility

### Sync Designs
1. Update Figma file
2. Detect changes via webhooks
3. Generate updated code
4. Commit changes to repository

### Audit Accessibility
1. Review component markup
2. Test keyboard navigation
3. Check color contrast
4. Verify screen reader support

## 📖 Documentation

- Figma API: https://www.figma.com/developers/api
- shadcn/ui: https://ui.shadcn.com
- Tailwind CSS: https://tailwindcss.com
- WCAG: https://www.w3.org/WAI/WCAG21/quickref/
- Accessibility: https://www.a11y-101.com

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - See LICENSE file for details

---

**Created by**: GOOS
**Version**: 1.0.0-dev
**Status**: Development
**Updated**: 2025-10-31
