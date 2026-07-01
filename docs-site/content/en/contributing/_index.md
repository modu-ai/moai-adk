---
title: Contributing
weight: 110
draft: false
---

MoAI-ADK is an open-source project and we welcome contributions! This guide explains how to contribute to the project.

## Quick Start

1. **Fork** the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Write tests (new code uses TDD, existing code uses characterization tests)
4. Verify all tests pass: `make test`
5. Verify linting passes: `make lint`
6. Format code: `make fmt`
7. Commit with Conventional Commit message
8. Create a Pull Request

## Code Quality Requirements

| Item | Criteria |
|------|----------|
| Test Coverage | **85%** or higher |
| Lint Errors | **0** |
| Type Errors | **0** |
| Commit Message | Conventional Commits format |

## Commit Message Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation change |
| `style` | Code formatting (no behavior change) |
| `refactor` | Refactoring (no behavior change) |
| `perf` | Performance improvement |
| `test` | Add/modify tests |
| `chore` | Build/tool changes |
| `revert` | Revert previous commit |

### Examples

```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
docs(readme): update agent count and statistics
```

## Development Environment Setup

### Required Tools

- **Go 1.26+** — Core development language
- **Git** — Version control
- **make** — Build commands

### Key Commands

```bash
make build        # Build the project
make test         # Run tests
make test-race    # Race condition detection tests
make lint         # Run linter
make fmt          # Format code
make install      # Install locally
make clean        # Clean build artifacts
```

## Pull Request Guide

### When Writing a PR

- Clear and concise title (70 characters or less)
- Summary of changes (Summary section)
- Test plan (Test Plan section)
- Reference related issues (e.g., `Fixes #123`)

### PR Checklist

- [ ] Tests added/updated
- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Commit message follows Conventional Commits format
- [ ] Documentation updated (if needed)

## Community

- **Issue Tracker**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — Bug reports, feature requests
- **Official Documentation**: [adk.mo.ai.kr](https://adk.mo.ai.kr)

## License

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) — Free to use, modify, and distribute.
