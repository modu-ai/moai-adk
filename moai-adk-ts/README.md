# MoAI-ADK

[![npm version](https://img.shields.io/npm/v/moai-adk)](https://www.npmjs.com/package/moai-adk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.2+-blue)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/node-18.0+-green)](https://nodejs.org/)

TypeScript-based SPEC-First TDD Development Kit with Universal Language Support

> **Note**: This is an early development version (v0.0.1). Features and APIs are subject to change.

## Features

- **SPEC-First TDD Workflow**: 3-stage development process (SPEC → TDD → Sync)
- **Universal Language Support**: Python, TypeScript, Java, Go, Rust, and more
- **Claude Code Integration**: 7 specialized agents for automated development
- **Complete Traceability**: @TAG system for full requirement-to-code tracking
- **Intelligent Diagnostics**: Automatic project language detection and environment optimization

## Installation

### Global Installation

```bash
# Using npm
npm install -g moai-adk

# Using Bun (recommended)
bun add -g moai-adk
```

### Local Installation

```bash
# Using npm
npm install moai-adk

# Using Bun
bun add moai-adk
```

## Requirements

- **Node.js**: 18.0 or higher
- **Git**: 2.30.0 or higher
- **npm**: 8.0.0 or higher (or Bun 1.2.0+)

## Quick Start

### 1. Initialize a New Project

```bash
moai init my-project
cd my-project
```

### 2. Check System Status

```bash
# Run system diagnostics
moai doctor

# Check project status
moai status
```

### 3. Development Workflow

```bash
# Stage 1: Write SPEC
/moai:1-spec "user authentication system"

# Stage 2: Implement with TDD
/moai:2-build SPEC-001

# Stage 3: Sync documentation
/moai:3-sync
```

## CLI Commands

### `moai init <project-name>`

Initialize a new MoAI-ADK project with the specified name.

```bash
moai init my-api --type web-api --language typescript
```

### `moai doctor`

Run comprehensive system diagnostics to verify environment setup.

```bash
moai doctor
moai doctor --list-backups
```

### `moai status`

Display current project status and configuration.

```bash
moai status --verbose
```

### `moai update`

Update MoAI-ADK templates to the latest version.

```bash
moai update --check
moai update --verbose
```

### `moai restore`

Restore project from backup.

```bash
moai restore <backup-path>
```

## Agent System

MoAI-ADK provides 7 specialized agents for different development tasks:

| Agent | Purpose | Usage |
|-------|---------|-------|
| **spec-builder** | EARS specification writing | `@agent-spec-builder "new feature spec"` |
| **code-builder** | TDD implementation | `@agent-code-builder "implement SPEC-001"` |
| **doc-syncer** | Documentation synchronization | `@agent-doc-syncer "update docs"` |
| **cc-manager** | Claude Code management | `@agent-cc-manager "optimize settings"` |
| **debug-helper** | Error diagnosis | `@agent-debug-helper "build failure analysis"` |
| **git-manager** | Git workflow automation | `@agent-git-manager "create feature branch"` |
| **trust-checker** | Quality verification | `@agent-trust-checker "code quality check"` |

## Language Support

| Language | Test Framework | Linter/Formatter | Build Tool |
|----------|----------------|------------------|------------|
| **TypeScript** | Vitest/Jest | Biome/ESLint | tsup/Vite |
| **Python** | pytest | ruff/black | uv/pip |
| **Java** | JUnit | checkstyle | Maven/Gradle |
| **Go** | go test | golint/gofmt | go mod |
| **Rust** | cargo test | clippy/rustfmt | cargo |

## @TAG System

The @TAG system provides complete traceability from requirements to implementation:

### Core Tags

- `@SPEC`: Requirements definition
- `@SPEC`: Architecture design
- `@CODE`: Implementation tasks
- `@TEST`: Test verification
- `@CODE`: Business features
- `@CODE`: Interface definitions
- `@CODE`: Security requirements
- `@DOC`: Documentation

### Usage Example

```typescript
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 -> @SPEC:AUTH-001 -> @CODE:AUTH-001 -> @TEST:AUTH-001
// Related: @CODE:AUTH-001, @CODE:AUTH-001:API

class AuthenticationService {
  /**
   * @CODE:AUTH-001:API: User authentication endpoint
   */
  async authenticate(email: string, password: string): Promise<boolean> {
    // @CODE:AUTH-001: Input validation
    if (!this.validateInput(email, password)) {
      return false;
    }

    return this.verifyCredentials(email, password);
  }
}
```

## TRUST Principles

All development follows the TRUST principles:

- **T**est First: Test-driven development (SPEC-First TDD)
- **R**eadable: Clear code (≤50 lines per function, clear naming)
- **U**nified: Single responsibility (≤300 lines per module, type safety)
- **S**ecured: Security by design (input validation, static analysis)
- **T**rackable: Complete traceability (@TAG system)

## API Usage

### Programmatic API

```typescript
import { MoAI } from 'moai-adk';

const moai = new MoAI();

// Initialize project
await moai.init('my-project', {
  type: 'web-api',
  language: 'typescript'
});

// Run diagnostics
const diagnostics = await moai.doctor();

// Get project status
const status = await moai.status();
```

### Configuration

```json
// .moai/config.json
{
  "project": {
    "name": "my-project",
    "type": "web-api",
    "language": "typescript"
  },
  "workflow": {
    "enableAutoSync": true,
    "gitIntegration": true
  }
}
```

## Development

### Setup Development Environment

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk/moai-adk-ts

# Install dependencies
bun install

# Run in development mode
bun run dev

# Build
bun run build

# Test
bun test
```

### Scripts

- `bun run build`: Build the project
- `bun run test`: Run tests
- `bun run test:coverage`: Run tests with coverage
- `bun run lint`: Lint code
- `bun run format`: Format code
- `bun run type-check`: Type checking

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Create a Pull Request

### Development Guidelines

- Follow TRUST principles
- Apply @TAG system
- Use TypeScript strict mode
- Maintain ≤50 lines per function
- Keep test coverage ≥85%

## License

This project is licensed under the [MIT License](LICENSE).

## Support

- **Issues**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)

---

**MoAI-ADK v0.0.1** - TypeScript-based SPEC-First TDD Development Framework