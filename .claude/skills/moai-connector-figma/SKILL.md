---
name: moai-connector-figma
description: Design System & UI Kit Development with Figma 2025, Design Tokens, and Component Documentation
version: 1.0.0
modularized: true
tags:
  - architecture
  - figma
  - enterprise
  - patterns
updated: 2025-11-26
status: active
---

## Quick Reference (30 seconds)

# Enterprise Figma & Design Systems

**Primary Focus**: Design systems, component libraries, design tokens, documentation
**Best For**: UI/UX design, design system creation, component documentation, design handoff
**Key Tools**: Figma 2025, FigJam, design tokens, component variants
**Auto-triggers**: Figma files, design system discussions, component documentation

| Tool | Version | Features |
|------|---------|----------|
| Figma | 2025 | Real-time collaboration, AI improvements |
| FigJam | Latest | Whiteboarding, collaborative design |
| Design Tokens | 2.0 | Token standardization |

---


## Implementation Guide (5 minutes)

### Features

- Design system architecture with W3C DTCG 2.0 token standards
- Component library management with variants and states
- Design-to-code workflow automation via Figma MCP
- Accessibility compliance auditing (WCAG 2.2)
- Real-time collaboration and version control
- Asset export and developer handoff

### When to Use

- Creating or refactoring design systems for multi-platform projects
- Building accessible component libraries with proper documentation
- Automating design token synchronization between design and code
- Setting up design-to-development workflows with version control
- Implementing design system governance and maintenance processes

### Core Patterns

**Pattern 1: Design Token Architecture**
```javascript
// Design tokens following DTCG 2.0
{
  "color": {
    "brand": {
      "primary": { "$value": "#0066CC", "$type": "color" },
      "secondary": { "$value": "#6C757D", "$type": "color" }
    }
  },
  "spacing": {
    "base": { "$value": "8px", "$type": "dimension" }
  }
}
```

**Pattern 2: Component Variant System**
- Create main components with logical variant properties (size, state, theme)
- Use auto-layout for responsive behavior
- Document usage guidelines in component descriptions
- Maintain consistent naming: Component/Variant/State

**Pattern 3: Design-to-Code Workflow**
1. Design components in Figma with proper naming
2. Export design tokens via Figma MCP plugin
3. Sync tokens to code repository (JSON → CSS/SCSS/JS)
4. Generate component boilerplate from Figma specs
5. Validate design compliance with automated tests

## What It Does

Enterprise-grade design system and UI kit development with Figma. Component documentation, design tokens, accessibility, and seamless developer handoff.

**Key capabilities**:
- ✅ Design system architecture and governance
- ✅ Component libraries with variants
- ✅ Design tokens and design-to-dev workflow
- ✅ Accessibility auditing in Figma
- ✅ Documentation and design specs
- ✅ Asset management and versioning
- ✅ Developer handoff and code generation

---

## When to Use

**Automatic triggers**:
- Design system creation
- Component library management
- UI kit development
- Design documentation

**Manual invocation**:
- Design system audit
- Component strategy review
- Token management
- Design-to-dev workflow optimization

---

## Three-Level Learning Path

### Level 1: Fundamentals (See examples.md)

Core design system concepts:
- **Figma Basics**: Pages, frames, components, variants
- **Component System**: Primary vs secondary components
- **Design Tokens**: Colors, typography, spacing
- **Documentation**: Specs, guidelines, patterns
- **Accessibility**: Color contrast, labels, states

### Level 2: Advanced Patterns (See modules/component-strategy.md)

Production design systems:
- **Variant Management**: States, sizes, variations
- **Token Architecture**: Design tokens for dev/design
- **Component Governance**: Naming, updates, versioning
- **Design Documentation**: Specifications, usage
- **Figma Plugins**: Automation, token sync

### Level 3: Developer Handoff (See modules/dev-workflow.md)

Design-to-development workflow:
- **Code Generation**: Components from Figma
- **Specs & Assets**: Automated export
- **Design Tokens**: Sync to code repositories
- **CI/CD Integration**: Design system versioning
- **Quality Assurance**: Design compliance testing

---

## Best Practices

✅ **DO**:
- Use main components for reusability
- Maintain consistent naming conventions
- Document all design tokens
- Version design system regularly
- Conduct accessibility audits
- Review component variants
- Keep documentation updated

❌ **DON'T**:
- Create duplicate components
- Skip accessibility checks
- Ignore design token standardization
- Over-complicate component structure
- Use inconsistent naming
- Forget to document changes
- Ignore developer feedback

---

## Tool Versions (2025-11-22)

| Tool | Version | Purpose |
|------|---------|---------|
| **Figma** | 2025 | Design tool |
| **Design Tokens** | 2.0 | Token standard |
| **FigJam** | Latest | Collaboration |
| **Penpot** | Latest | Open source alternative |

---

## Works Well With

- `moai-domain-frontend` (React component mapping)
- `moai-lang-html-css` (HTML/CSS semantic markup)
- `moai-system-universal` (UX/UI design principles)

---

## Learn More

- **Examples**: See `examples.md` for design system patterns
- **Component Strategy**: See `modules/component-strategy.md` for component systems
- **Dev Workflow**: See `modules/dev-workflow.md` for design-to-dev handoff
- **Figma Docs**: https://help.figma.com/
- **Design Tokens**: https://designtokens.org/

---

## Changelog

- **v4.0.0** (2025-11-22): Modularized with strategy and workflow modules
- **v3.0.0** (2025-11-13): Figma 2025 features, design tokens 2.0
- **v2.0.0** (2025-10-01): Component variants, design systems
- **v1.0.0** (2025-03-01): Initial release

---

**Skills**: Skill("moai-lang-unified"), Skill("moai-lang-html-css"), Skill("moai-system-universal")
**Auto-loads**: Design system files, Figma projects

