# Contributing to UI/UX Plugin

Welcome to the MoAI UI/UX Plugin! This guide helps you contribute to design systems and Figma integration.

## Getting Started

### Prerequisites
- Figma API knowledge
- React and design system familiarity
- TypeScript and Tailwind CSS experience
- Design principles understanding

### Development Setup

```bash
# Clone repository
git clone https://github.com/moai-adk/moai-marketplace.git
cd moai-marketplace/plugins/moai-plugin-uiux

# Install dependencies
npm install

# Setup environment
cp .env.example .env.local
# Add Figma API token
```

### Testing Skills

```bash
# Validate skill structure
npm run validate-skills

# Test Figma integration
npm test -- --testPathPattern=figma

# Test design-to-code conversion
npm test -- --testPathPattern=design-to-code
```

## Contribution Areas

### Design System Skills
When contributing design system content:
- Document design tokens and variables
- Include theming examples
- Show dark mode implementation
- Add accessibility guidelines

### Figma Integration Skills
For Figma MCP skills:
- Include API setup examples
- Document webhook configuration
- Show design sync workflows
- Add error handling

### Component Skills
For component documentation:
- Include design specifications
- Show customization patterns
- Demonstrate variants
- Add accessibility considerations

## Branch Naming
- Feature: `feature/design-pattern-name`
- Fix: `fix/issue-description`
- Docs: `docs/skill-name`

## Commit Messages
```
feat(Figma): add component sync skill
fix(Design): resolve theme switching bug
docs(shadcn): update customization guide
```

## Skill Guidelines

### Design Token Skills
Should cover:
- Color system definition
- Typography scale
- Spacing system
- Shadow system

### Component Skills
Should include:
- Component props and API
- Usage examples
- Customization patterns
- Dark mode support
- Accessibility features

## PR Checklist
- [ ] Skill structure is valid
- [ ] Examples are runnable
- [ ] Design tokens are consistent
- [ ] Accessibility guidelines included
- [ ] Dark mode examples provided

## Design Review

For visual changes:
- Include before/after screenshots
- Test on multiple screen sizes
- Verify dark mode
- Check accessibility (a11y)

## Questions?

- Check Figma API docs
- Review design system docs
- Join design community
- Contact maintainers

Thank you for contributing! 🎨
