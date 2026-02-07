# Contributing Guide

Thank you for your interest in contributing to MoAI-ADK Go Edition! This guide will help you contribute effectively to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Writing Tests](#writing-tests)
- [Commit Rules](#commit-rules)
- [Pull Requests](#pull-requests)

---

## Getting Started

### 1. Fork and Clone Repository

```bash
# Fork the repository on GitHub
# Clone your fork locally
git clone https://github.com/YOUR_USERNAME/moai-adk-go.git
cd moai-adk-go

# Add upstream remote
git remote add upstream https://github.com/modu-ai/moai-adk-go.git
```

### 2. Create Development Branch

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test additions or modifications

---

## Development Setup

### Required Tools

- **Go 1.25+**: Install from [golang.org](https://go.dev/dl/)
- **golangci-lint**: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **gofumpt**: `go install mvdan.cc/gofumpt@latest`

### Development Workflow

```bash
# 1. Install dependencies
go mod download

# 2. Run tests
make test

# 3. Check linting
make lint

# 4. Format code
make fmt

# 5. Build
make build
```

---

## Code Style

### Go Conventions

All code must follow [Effective Go](https://go.dev/doc/effective-go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

### Project-Specific Rules

1. **Package names**: lowercase, single word (e.g., `template`, `cli`)
2. **Exported**: PascalCase (e.g., `TemplateDeployer`, `NewManager`)
3. **Unexported**: camelCase (e.g., `deployTemplates`, `localConfig`)
4. **Constants**: PascalCase or UPPER_SNAKE_CASE
5. **Interfaces**: Names ending in `-er` (e.g., `Reader`, `Writer`)

### Comments

```go
// Package template provides template deployment and rendering...
package template

// Deployer extracts and deploys templates from an embedded filesystem...
type Deployer interface { ... }

// NewDeployer creates a new Deployer with the given options...
func NewDeployer(opts ...Option) (Deployer, error) { ... }
```

### Error Handling

```go
// Error wrapping - always include context
if err != nil {
    return fmt.Errorf("deploy templates: %w", err)
}

// Error wrapping - include path information
return fmt.Errorf("read %q: %w", path, err)  // GOOD
return fmt.Errorf("read " + path + ": " + err.Error())  // BAD
```

---

## Writing Tests

### Test Coverage Requirements

- All packages must maintain **minimum 85%** coverage
- New code follows **TDD approach** (RED-GREEN-REFACTOR)
- Existing code changes follow **DDD approach** (ANALYZE-PRESERVE-IMPROVE)

### Table-Driven Tests (Recommended)

```go
func TestBuildRequiredPATH(t *testing.T) {
    tests := []struct {
        name    string
        goBin   string
        goPath  string
        want    string
    }{
        {"default", "", "", wantDefault},
        {"custom bin", "/custom/bin", "", wantCustom},
        {"custom path", "", "/custom/path", wantPath},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Race Detection

```bash
# Always run tests with race detection
go test -race ./...

# Using Makefile
make test
```

---

## Commit Rules

### Commit Message Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Type Categories

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only changes
- `style`: Code formatting, semicolons, etc (no code behavior change)
- `refactor`: Code refactoring (neither new feature nor bug fix)
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Build process, tool updates, etc.

### Examples

```
feat(cli): add interactive wizard for project initialization

The new wizard guides users through:
- Development mode selection (DDD/TDD/Hybrid)
- Language preference configuration
- Quality gates customization

Closes #123
```

```
fix(update): correct version comparison for go-v prefix

The update checker was failing to compare versions starting with
"go-v" prefix. This fix normalizes version strings before comparison.

Fixes #456
```

---

## Pull Requests

### Pre-Submission Checklist

- [ ] All tests pass (`make test`)
- [ ] Zero lint errors (`make lint`)
- [ ] Code formatted (`make fmt`)
- [ ] 85%+ test coverage
- [ ] Race detection passes
- [ ] Commit messages follow conventional commit format
- [ ] PR description includes summary of changes

### PR Template

```markdown
## Description
This PR implements [feature/bug fix].

## Changes
- Change 1
- Change 2

## Testing
- Test method 1
- Test method 2

## Checklist
- [x] All tests pass
- [x] Linting passes
- [x] Documentation updated
```

---

## Code of Conduct

- Maintain respect and inclusivity
- Provide constructive feedback
- Help and support other contributors
- Report violations to [report@modu-ai](mailto:report@modu-ai)

---

## License

Your contributions will be licensed under the project's [Copyleft 3.0 License](LICENSE).

---

For questions, please create a [GitHub Issue](https://github.com/modu-ai/moai-adk-go/issues).
