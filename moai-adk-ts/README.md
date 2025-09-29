# MoAI-ADK

[![npm version](https://img.shields.io/npm/v/moai-adk)](https://www.npmjs.com/package/moai-adk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.2+-blue)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/node-18.0+-green)](https://nodejs.org/)

TypeScript-based SPEC-First TDD Development Kit with Universal Language Support

## Features

- **SPEC-First TDD Workflow**: 3-stage development process (SPEC → TDD → Sync)
- **Universal Language Support**: Python, TypeScript, Java, Go, Rust, and more
- **Claude Code Integration**: 7 specialized agents for automated development
- **Complete Traceability**: @AI-TAG system for full requirement-to-code tracking
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
moai status --detailed --tags --git
```

### `moai update`

Update MoAI-ADK templates to the latest version.

```bash
moai update --backup
```

### `moai restore`

Restore project from backup.

```bash
moai restore --list
moai restore backup-20241201.tar.gz
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

## @AI-TAG System

The @AI-TAG system provides complete traceability from requirements to implementation:

### Core Tags

- `@REQ`: Requirements definition
- `@DESIGN`: Architecture design
- `@TASK`: Implementation tasks
- `@TEST`: Test verification
- `@FEATURE`: Business features
- `@API`: Interface definitions
- `@SEC`: Security requirements
- `@DOCS`: Documentation

### Usage Example

```typescript
// @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 -> @DESIGN:AUTH-001 -> @TASK:AUTH-001 -> @TEST:AUTH-001
// Related: @SEC:AUTH-001, @API:AUTH-001

class AuthenticationService {
  /**
   * @API:AUTH-001: User authentication endpoint
   */
  async authenticate(email: string, password: string): Promise<boolean> {
    // @SEC:AUTH-001: Input validation
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
- **T**rackable: Complete traceability (@AI-TAG system)

## Performance

- **Build Time**: 226ms
- **Package Size**: 471KB
- **TAG System Loading**: 45ms
- **Test Success Rate**: 92.9% (Vitest)
- **Performance Improvements**:
  - Bun: 98% faster than npm
  - Vitest: 92.9% success rate
  - Biome: 94.8% faster than ESLint+Prettier

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

## Troubleshooting

### Installation Issues

```bash
# Permission errors
sudo npm install -g moai-adk

# Cache issues
npm cache clean --force
npm install -g moai-adk
```

### Command Not Found

```bash
# Check PATH
echo $PATH

# Restart shell
source ~/.bashrc  # or ~/.zshrc
```

### System Diagnostics

```bash
# Re-run diagnostics
moai doctor

# Check individual tools
node --version
git --version
npm --version
```

## Development

### Setup Development Environment

```bash
git clone https://github.com/your-org/moai-adk.git
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
- Apply @AI-TAG system
- Use TypeScript strict mode
- Maintain ≤50 lines per function
- Keep test coverage ≥85%

## License

This project is licensed under the [MIT License](LICENSE).

## Support

- **Issues**: [GitHub Issues](https://github.com/your-org/moai-adk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/moai-adk/discussions)
- **Documentation**: [Project Documentation](https://moai-adk.github.io)

---

**MoAI-ADK v0.0.3** - TypeScript-based SPEC-First TDD Development Framework