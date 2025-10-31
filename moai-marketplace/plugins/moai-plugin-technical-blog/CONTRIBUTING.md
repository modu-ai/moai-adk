# Contributing to Technical Blog Plugin

Welcome to the MoAI Technical Blog Plugin! Help us build the best technical writing tools.

## Getting Started

### Prerequisites
- Technical writing experience
- Markdown proficiency
- Understanding of SEO
- Content management familiarity

### Development Setup

```bash
# Clone repository
git clone https://github.com/moai-adk/moai-marketplace.git
cd moai-marketplace/plugins/moai-plugin-technical-blog

# Install dependencies
npm install

# Start development
npm run dev
```

### Testing

```bash
# Test content skills
npm test

# Test markdown parsing
npm test -- --testPathPattern=markdown

# Test SEO optimization
npm test -- --testPathPattern=seo

# Check coverage
npm test -- --coverage
```

## Contribution Areas

### Content Skills
When adding content skills:
- Provide real-world examples
- Include templates
- Document best practices
- Add workflow guides

### Publishing Integration
For multi-platform skills:
- Test platform-specific formatting
- Include API examples
- Document authentication
- Add error handling

### SEO & Analytics Skills
For optimization content:
- Include checklists
- Provide real examples
- Document tools and resources
- Add measurement approaches

## Branch Naming
- Feature: `feature/skill-name`
- Fix: `fix/issue-description`
- Docs: `docs/guide-name`

## Commit Messages
```
feat(Content): add image generation skill
fix(SEO): resolve meta tag formatting
docs(Strategy): update editorial calendar
```

## Skill Guidelines

### Blog Strategy Skills
Should cover:
- Content planning
- Audience targeting
- Performance metrics
- Growth strategies

### Publishing Skills
Should include:
- Platform setup
- Content formatting
- Publishing workflows
- Automation examples

### SEO Skills
Should demonstrate:
- Meta optimization
- Structured data
- Keyword research
- Content strategy

## Content Quality

### Writing Standards
- Clear, concise language
- Real-world examples
- Links to resources
- Template downloads

### Technical Accuracy
- Verify all code examples
- Test publishing workflows
- Check platform APIs
- Validate tools/resources

## PR Checklist
- [ ] Content is clear and accurate
- [ ] Examples are tested
- [ ] Links are valid
- [ ] Formatting is consistent
- [ ] No typos or grammar errors

## Review Process

Before publishing:
1. Self-review for clarity
2. Check all links
3. Test all examples
4. Peer review from team
5. Final approval

## Questions?

- Check existing docs
- Review style guide
- Ask in discussions
- Contact maintainers

Thank you for contributing! 📝
