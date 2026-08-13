---
title: Contributing
weight: 110
draft: false
---

MoAI-ADK is an open-source project and contributions are welcome! MoAI-ADK
itself is developed with the SPEC-based 3-phase workflow and the TRUST 5
quality gates — the quality bar for contributions (coverage, lint,
conventional commits) follows those same standards.


## Contribution flow at a glance

The contribution flow has five steps. Naming why each step exists makes the procedure below fall into place naturally.

1. **Explore** — Look over the repository structure and existing issues.
2. **Branch** — Create a working branch to isolate your changes from the main code.
3. **Implement** — Make the change and verify it with tests.
4. **Submit** — Open a PR following the Conventional Commits message format.
5. **Review** — Reflect feedback and iterate until the change is merged into main.

{{< icon bulb muted >}} First-time contributors may want to start with issues labeled
"good first issue". Their scope is small and bounded, so the entry is gentle.


## Quick start

1. **Fork** the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Write tests (TDD for new code, characterization tests for existing code)
4. Verify all tests pass: `make test`
5. Verify linting passes: `make lint`
6. Format the code: `make fmt`
7. Commit with a Conventional Commit message
8. Open a Pull Request

## Code quality requirements

The **T**ested / **T**rackable criteria of the TRUST 5 framework apply as-is:

| Item | Standard |
|------|------|
| Test coverage | **85%** or higher |
| Lint errors | **0** |
| Type errors | **0** |
| Commit messages | Conventional Commits format |

## Commit message format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation change |
| `style` | Code formatting (no functional change) |
| `refactor` | Refactoring (no functional change) |
| `perf` | Performance improvement |
| `test` | Adding/updating tests |
| `chore` | Build/tooling change |
| `revert` | Revert a previous commit |

### Examples

```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
docs(readme): update agent count and statistics
```

## Development environment setup

### Required tools

- **Go 1.26+** — the core development language
- **Git** — version control
- **make** — build commands

### Key commands

```bash
make build        # Build the project
make test         # Run tests
make test-race    # Run tests with race detection
make lint         # Run the linter
make fmt          # Format the code
make install      # Install locally
make clean        # Clean build artifacts
```

## Pull Request guide

### When writing a PR

- A clear, concise title (70 characters or less)
- A summary of the changes (Summary section)
- A test plan (Test Plan section)
- Related issue references (e.g., `Fixes #123`)

### PR checklist

- [ ] Tests added/updated
- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Commit messages follow Conventional Commits
- [ ] Documentation updated (if needed)


## Create an issue right from the session

When you find a bug or want to propose a feature while using MoAI-ADK, you do not have to leave for the GitHub web. Type `/moai feedback` inside the session, and an issue draft that reflects the current conversation context is composed and submitted to the repository's issue tracker automatically. Bug-reproduction steps and environment info are attached, so the back-and-forth is reduced.


## Community

- **Issue tracker**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — bug reports and feature requests.
- **Discord**: [Discord Community](https://discord.gg/Z7E7Mdc5aN) — real-time chat, tips
- **Official docs**: [adk.mo.ai.kr](https://adk.mo.ai.kr)

## License

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) — free to use, modify, and distribute.
