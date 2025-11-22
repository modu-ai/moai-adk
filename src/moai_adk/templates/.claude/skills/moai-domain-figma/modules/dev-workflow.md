# Figma Developer Workflow & Code Generation

_Last updated: 2025-11-22_

## Design-to-Code Workflow

### Export Strategy
1. Prepare Figma components with constraints
2. Export assets (icons, illustrations) as SVG/PNG
3. Generate code from components using plugins
4. Map design tokens to CSS variables
5. Version design system in repository

### Figma Plugins for Development
- **Figma Tokens**: Sync design tokens to code
- **Component Extractor**: Export component specs
- **Export API**: Automated asset generation
- **Storybook**: Generate Storybook stories
- **Design System Tools**: Custom automation

## Design Tokens Sync

### Token Categories
```json
{
  "color": {
    "primary": { "value": "#007BFF" },
    "secondary": { "value": "#6C757D" }
  },
  "spacing": {
    "xs": { "value": "4px" },
    "sm": { "value": "8px" },
    "md": { "value": "16px" }
  }
}
```

### Integration with Code
- Sync tokens to CSS variables
- Generate TypeScript types
- Update design token library
- Automate version bumps

## Asset Management

### Icon System
- Export icons as SVG components
- Maintain icon naming convention
- Size variants (16px, 24px, 32px)
- Color accessibility checks

### Component Assets
- Illustrations as SVG
- Photography optimized
- Dark mode variants
- Responsive images

## Specification & Documentation

### Component Specs
- Use Figma Specs feature
- Document all interactions
- Show responsive behavior
- Include accessibility notes
- Link to development code

### Handoff Checklist
- [ ] All components exported
- [ ] Design tokens documented
- [ ] Assets optimized
- [ ] Specs reviewed
- [ ] Accessibility verified
- [ ] Developer notes added
- [ ] Versioning documented

---

**Last Updated**: 2025-11-22

