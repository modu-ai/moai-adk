# Contributing to DevOps Plugin

Welcome to the MoAI DevOps Plugin! This guide will help you contribute effectively.

## Getting Started

### Prerequisites
- Understanding of CI/CD, Docker, Kubernetes
- Knowledge of cloud platforms (Vercel, Render, Supabase)
- Experience with infrastructure as code
- Bash/Shell scripting knowledge

### Development Setup

```bash
# Clone repository
git clone https://github.com/moai-adk/moai-marketplace.git
cd moai-marketplace/plugins/moai-plugin-devops

# Install dependencies
npm install

# Verify setup
npm run test
```

### Running Tests

```bash
# Test all workflows
npm test

# Test specific area
npm test -- --testPathPattern=vercel

# Run with coverage
npm test -- --coverage
```

## Contribution Areas

### MCP Server Skills
When adding/updating MCP server integrations:
- Document API authentication
- Include error handling examples
- Test with real credentials (safely)
- Update version compatibility

### Deployment Patterns
For deployment scripts:
- Include health checks
- Document rollback procedures
- Add monitoring examples
- Test in staging first

### CI/CD Workflows
For GitHub Actions:
- Test all branches
- Include notification steps
- Add status reporting
- Document triggers

## Skill Structure

Skills should cover:
- Platform setup and authentication
- Common deployment patterns
- Error handling and debugging
- Monitoring and alerts
- Cost optimization tips

## Branch Naming
- Feature: `feature/platform-integration`
- Fix: `fix/deployment-issue`
- Docs: `docs/setup-guide`

## Commit Messages
```
feat(Render): add background job scheduling
fix(CI): resolve GitHub Actions caching issue
docs(Vercel): update edge function guide
```

## PR Guidelines

When submitting PRs:
- Include deployment examples
- Test in staging environment
- Document breaking changes
- Update troubleshooting guide

## Security
- Never commit credentials
- Use environment variables
- Document security best practices
- Include OWASP considerations

## Questions?

- Check troubleshooting guide
- Open an issue
- Contact maintainers

Thank you for contributing! 🚀
