# Design System Component Strategy

_Last updated: 2025-11-22_

## Component Hierarchy

### Main Components vs. Instances
- Main components stored in dedicated library file
- Instances created in working files
- Version control through Figma Teams
- Regular audits to detect orphaned instances

### Naming Convention
```
Category / Component Name / Property

Button / Primary / Default
Button / Primary / Hovered
Card / Product / Large
Input / Text / Filled
```

## Variant Management

### Creating Variants
- Group related states into single component
- Use boolean and enum properties
- Name variants descriptively
- Document each variant's purpose

### State Coverage
- Default state
- Hover/Focus states
- Disabled state
- Loading state
- Error state
- Success state

## Design Tokens

### Token Categories
- Colors: Primary, secondary, neutral, semantic
- Typography: Font family, size, weight, line height
- Spacing: Units (8px scale recommended)
- Sizing: Component sizes
- Shadows: Elevation system
- Radius: Border radius scale

### Token Naming
```
{category}-{subcategory}-{state}

color-primary-default
color-primary-hover
spacing-4
typography-body-regular
shadow-elevation-1
```

## Component Documentation

### Specs Document
- Component description
- Use cases
- Dos and don'ts
- Examples in context
- Accessibility notes
- Developer integration

### Asset Library
- Icons
- Illustrations
- Colors (swatches)
- Typography samples
- Spacing system

---

**Last Updated**: 2025-11-22

