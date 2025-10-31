# Contributing to Frontend Plugin

Thank you for contributing to the MoAI Frontend Plugin! This document provides guidelines for contributing.

## Getting Started

### Prerequisites
- Node.js 18+
- TypeScript knowledge
- React and Next.js familiarity

### Development Setup

```bash
# Clone and setup
git clone https://github.com/moai-adk/moai-marketplace.git
cd moai-marketplace/plugins/moai-plugin-frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

### Running Tests

```bash
# Run all tests
npm test

# Run with coverage
npm test -- --coverage

# Run specific test
npm test components/Button.test.tsx
```

### Code Quality

```bash
# Type checking
npm run type-check

# Linting
npm run lint

# Format code
npm run format
```

## Making Changes

### Branch Naming
- Feature: `feature/component-name`
- Bug: `fix/short-description`
- Docs: `docs/short-description`

### Commit Messages
```
feat(Button): add size variants
fix(Layout): resolve responsive issue
docs(Components): update usage guide
```

## Skill Guidelines

### Component Skills
Skills should demonstrate:
- Component structure and props
- Common patterns and use cases
- Customization and theming
- Accessibility considerations
- Performance optimization

### Code Examples
```typescript
// Clear, runnable examples
// Show real-world usage
// Include TypeScript types
// Demonstrate best practices
```

## Testing Requirements
- Unit tests for components
- Integration tests for features
- Accessibility testing with jest-axe
- Minimum 80% coverage

## Documentation

### Update README when:
- Adding new components
- Changing component API
- Updating dependencies
- Adding new features

### Code Comments
- Explain complex logic
- Document edge cases
- Reference related docs
- Include TypeScript types

## PR Checklist
- [ ] Code follows style guide
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No TypeScript errors
- [ ] Components tested on mobile

## Questions?

- Create an issue
- Join our Discord
- Check existing discussions

Thank you for contributing! 🚀
