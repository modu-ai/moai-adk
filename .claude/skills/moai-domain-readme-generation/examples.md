# README Generation Examples

This document provides complete README generation scenarios across multiple languages and project types.

---

## Example 1: Python CLI Tool (MoAI-ADK Style)

### Input Context
```yaml
# pyproject.toml
[project]
name = "moai-adk"
version = "0.4.6"
description = "MoAI Agentic Development Kit - SPEC-First TDD with Alfred SuperAgent"
requires-python = ">=3.12"
license = {text = "MIT"}
dependencies = [
    "click>=8.1.7",
    "pyyaml>=6.0.1",
    "pytest>=8.0.0"
]

[project.urls]
Homepage = "https://github.com/modu-ai/moai-adk"
Repository = "https://github.com/modu-ai/moai-adk"
```

### Detection Phase Output
```
✅ Detected: Python CLI tool
✅ Package manager: uv (pyproject.toml found)
✅ Version: 0.4.6
✅ License: MIT
✅ Dependencies: 3 (click, pyyaml, pytest)
```

### Generated README Structure
```markdown
# MoAI-ADK (Agentic Development Kit)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.12+-blue)](https://www.python.org/)

> **MoAI-ADK delivers a seamless development workflow that naturally connects SPEC → TEST (TDD) → CODE → DOCUMENTATION with AI.**

---

## What is MoAI-ADK?

### The Problem: Trust Crisis in AI Development
[Generated from project description and context analysis]

### The Solution: SPEC-First TDD with Alfred SuperAgent
[Key value propositions extracted from metadata]

---

## Quick Start (5 minutes)

### Step 1: Install uv (about 30 seconds)
```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows (PowerShell)
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Verify installation
uv --version
```

### Step 2: Install MoAI-ADK (about 1 minute)
```bash
# Install in tool mode (recommended)
uv tool install moai-adk

# Verify installation
moai-adk --version
# Output: MoAI-ADK v0.4.6
```

### Step 3: Create Your First Project (about 1 minute)
```bash
# Initialize new project
moai-adk init my-project
cd my-project

# Start Claude Code
claude

# Run Alfred
/alfred:0-project
```

---

## Features

### Core Capabilities
- ✅ SPEC-First Development (EARS syntax)
- ✅ Test-Driven Development (TDD) automation
- ✅ Alfred SuperAgent orchestration
- ✅ 56 Claude Skills library
- ✅ Multi-language support (Python, TypeScript, Go, Rust, Java, etc.)
- ✅ Automatic documentation sync

### Advanced Features
- ✅ Claude Code Hooks integration
- ✅ GitFlow automation
- ✅ TAG system for traceability
- ✅ TRUST 5 principles enforcement
- ✅ Living Documentation generation

---

## Technology Stack

| Category | Technology | Version | Purpose |
| --- | --- | --- | --- |
| Language | Python | 3.12+ | Core implementation |
| CLI Framework | Click | 8.1+ | Command-line interface |
| Testing | pytest | 8.0+ | Unit & integration tests |
| Package Manager | uv | 0.x | Fast dependency management |

---

## Community & Support

| Channel | Link |
| --- | --- |
| GitHub Repository | https://github.com/modu-ai/moai-adk |
| Issue Tracker | https://github.com/modu-ai/moai-adk/issues |
| PyPI Package | https://pypi.org/project/moai-adk/ |
| Documentation | See `.moai/`, `.claude/` within project |

---

**MoAI-ADK v0.4.6** — SPEC-First TDD with AI SuperAgent
```

---

## Example 2: TypeScript Library (npm Package)

### Input Context
```json
// package.json
{
  "name": "@acme/data-validator",
  "version": "2.1.0",
  "description": "Type-safe runtime data validation for TypeScript with zero dependencies",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "test": "vitest",
    "lint": "biome check ."
  },
  "keywords": ["validation", "typescript", "type-safe", "runtime"],
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/acme/data-validator.git"
  },
  "engines": {
    "node": ">=18.0.0"
  },
  "devDependencies": {
    "typescript": "^5.3.0",
    "vitest": "^1.0.0",
    "@biomejs/biome": "^1.4.0"
  }
}
```

### Detection Phase Output
```
✅ Detected: TypeScript library
✅ Package manager: npm (package.json found)
✅ Version: 2.1.0
✅ License: MIT
✅ Node.js requirement: >=18.0.0
✅ Testing framework: Vitest
✅ Linter: Biome
```

### Generated README Structure
```markdown
# @acme/data-validator

[![npm version](https://img.shields.io/npm/v/@acme/data-validator)](https://www.npmjs.com/package/@acme/data-validator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.3+-blue)](https://www.typescriptlang.org/)

> **Type-safe runtime data validation for TypeScript with zero dependencies**

---

## What is @acme/data-validator?

### The Problem
TypeScript provides compile-time type safety, but runtime validation still requires manual checks or heavy validation libraries with complex schemas.

### The Solution
`@acme/data-validator` provides lightweight, type-safe runtime validation that integrates seamlessly with TypeScript's type system. No schemas, no learning curve—just TypeScript types.

---

## Quick Start

### Installation
```bash
# npm
npm install @acme/data-validator

# pnpm
pnpm add @acme/data-validator

# yarn
yarn add @acme/data-validator
```

### Basic Usage
```typescript
import { validate, string, number } from '@acme/data-validator';

// Define validation schema
const userSchema = {
  name: string(),
  age: number(),
  email: string().email()
};

// Validate data
const result = validate(userSchema, inputData);

if (result.success) {
  console.log('Valid user:', result.data);
  // result.data is fully typed!
} else {
  console.error('Validation errors:', result.errors);
}
```

---

## Features

- ✅ **Zero dependencies**: No external dependencies
- ✅ **Type inference**: Automatic TypeScript type inference
- ✅ **Composable validators**: Build complex validations from primitives
- ✅ **Clear error messages**: Detailed validation error reporting
- ✅ **Lightweight**: < 5KB minified + gzipped

---

## Technology Stack

| Category | Technology | Version | Purpose |
| --- | --- | --- | --- |
| Language | TypeScript | 5.3+ | Type-safe implementation |
| Testing | Vitest | 1.0+ | Unit tests |
| Linter | Biome | 1.4+ | Code quality |
| Package Manager | npm/pnpm | - | Distribution |

---

## Documentation

| Resource | Link |
| --- | --- |
| API Reference | [API.md](docs/API.md) |
| Examples | [examples/](examples/) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

---

## Community & Support

| Channel | Link |
| --- | --- |
| GitHub Repository | https://github.com/acme/data-validator |
| npm Package | https://www.npmjs.com/package/@acme/data-validator |
| Issue Tracker | https://github.com/acme/data-validator/issues |

---

**@acme/data-validator v2.1.0** — Type-safe runtime validation for TypeScript
```

---

## Example 3: Go Microservice (Docker Deployment)

### Input Context
```go
// go.mod
module github.com/acme/order-service

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.8.4
)
```

```dockerfile
# Dockerfile present
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o order-service ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/order-service /usr/local/bin/
EXPOSE 8080
CMD ["order-service"]
```

### Detection Phase Output
```
✅ Detected: Go microservice (main.go + Dockerfile found)
✅ Go version: 1.22
✅ Framework: Gin (HTTP router)
✅ Database: PostgreSQL (lib/pq driver)
✅ Deployment: Docker (Dockerfile present)
```

### Generated README Structure
```markdown
# Order Service

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-27.0+-blue)](https://www.docker.com/)

> **Scalable order management microservice built with Go and PostgreSQL**

---

## What is Order Service?

A high-performance RESTful API for order management with:
- Fast order processing (< 100ms response time)
- PostgreSQL persistence with transaction support
- Docker containerization for easy deployment
- Comprehensive test coverage (85%+)

---

## Quick Start (5 minutes)

### Prerequisites
- Go 1.22+
- PostgreSQL 16+
- Docker 27.0+ (optional)

### Method 1: Run with Go
```bash
# Clone repository
git clone https://github.com/acme/order-service.git
cd order-service

# Install dependencies
go mod download

# Run database migrations
go run cmd/migrate/main.go

# Start server
go run cmd/server/main.go
# Server running on :8080
```

### Method 2: Run with Docker
```bash
# Build image
docker build -t order-service:latest .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/orders" \
  order-service:latest
```

### Verify Installation
```bash
# Health check
curl http://localhost:8080/health
# {"status": "healthy"}

# Create test order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"product_id": "abc123", "quantity": 5}'
```

---

## Features

### Core Capabilities
- ✅ RESTful API with Gin framework
- ✅ PostgreSQL database with connection pooling
- ✅ Transaction support for atomic operations
- ✅ Request validation and error handling
- ✅ Structured logging with context

### Advanced Features
- ✅ Docker multi-stage builds (optimized images)
- ✅ Health check endpoints
- ✅ Graceful shutdown handling
- ✅ Prometheus metrics (planned)

---

## Technology Stack

| Category | Technology | Version | Purpose |
| --- | --- | --- | --- |
| Language | Go | 1.22+ | Core implementation |
| Framework | Gin | 1.10+ | HTTP router |
| Database | PostgreSQL | 16+ | Data persistence |
| Driver | lib/pq | 1.10+ | PostgreSQL driver |
| Testing | testify | 1.8+ | Unit & integration tests |
| Deployment | Docker | 27.0+ | Containerization |

---

## API Documentation

See [API.md](docs/API.md) for full endpoint documentation.

### Quick Reference

| Endpoint | Method | Description |
| --- | --- | --- |
| `/health` | GET | Health check |
| `/api/v1/orders` | GET | List orders |
| `/api/v1/orders` | POST | Create order |
| `/api/v1/orders/:id` | GET | Get order by ID |
| `/api/v1/orders/:id` | PUT | Update order |
| `/api/v1/orders/:id` | DELETE | Delete order |

---

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/handlers -run TestCreateOrder
```

---

## Community & Support

| Channel | Link |
| --- | --- |
| GitHub Repository | https://github.com/acme/order-service |
| Issue Tracker | https://github.com/acme/order-service/issues |
| Docker Hub | https://hub.docker.com/r/acme/order-service |

---

**Order Service v1.0.0** — Scalable order management microservice
```

---

## Example 4: Rust CLI Tool (cargo install)

### Input Context
```toml
# Cargo.toml
[package]
name = "json-query"
version = "0.3.2"
edition = "2021"
description = "Fast JSON querying and filtering from the command line"
license = "MIT OR Apache-2.0"
repository = "https://github.com/acme/json-query"

[dependencies]
clap = { version = "4.4", features = ["derive"] }
serde_json = "1.0"
anyhow = "1.0"

[dev-dependencies]
assert_cmd = "2.0"
predicates = "3.0"
```

### Detection Phase Output
```
✅ Detected: Rust CLI tool (Cargo.toml + src/main.rs found)
✅ Rust edition: 2021
✅ Version: 0.3.2
✅ License: MIT OR Apache-2.0
✅ CLI framework: Clap 4.4
✅ JSON parsing: serde_json
```

### Generated README Structure
```markdown
# json-query

[![Crates.io](https://img.shields.io/crates/v/json-query)](https://crates.io/crates/json-query)
[![License](https://img.shields.io/badge/License-MIT%20OR%20Apache--2.0-blue)](https://github.com/acme/json-query)
[![Rust](https://img.shields.io/badge/Rust-2021-orange)](https://www.rust-lang.org/)

> **Fast JSON querying and filtering from the command line**

---

## What is json-query?

### The Problem
Parsing and filtering JSON in the terminal typically requires complex `jq` expressions or writing custom scripts. For quick queries, this overhead slows down exploration.

### The Solution
`json-query` provides a simple, intuitive CLI for common JSON operations with:
- **Fast execution**: Rust's performance for large JSON files
- **Simple syntax**: Plain English filters like `where status == "active"`
- **Pretty output**: Automatic colorized formatting

---

## Quick Start

### Installation
```bash
# Install via cargo
cargo install json-query

# Verify installation
json-query --version
# json-query 0.3.2
```

### Basic Usage
```bash
# Query JSON file
json-query data.json "where status == 'active'"

# Filter array
echo '[{"name":"Alice","age":30},{"name":"Bob","age":25}]' | json-query "where age > 26"
# Output:
# [
#   {"name": "Alice", "age": 30}
# ]

# Select specific fields
json-query data.json "select name, email where role == 'admin'"
```

---

## Features

### Core Operations
- ✅ **where**: Filter objects by condition
- ✅ **select**: Choose specific fields
- ✅ **sort**: Order results by field
- ✅ **limit**: Restrict output count
- ✅ **count**: Count matching items

### Advanced Features
- ✅ Pretty-printed JSON output with colors
- ✅ Streaming mode for large files (low memory usage)
- ✅ Nested field access (`user.profile.email`)
- ✅ Multiple condition operators (`==`, `!=`, `>`, `<`, `contains`)

---

## Examples

### Filter by Multiple Conditions
```bash
json-query users.json "where age > 25 AND role == 'developer'"
```

### Select and Sort
```bash
json-query products.json "select name, price where category == 'electronics' sort price desc limit 10"
```

### Count Items
```bash
json-query orders.json "count where status == 'completed'"
# 142
```

---

## Technology Stack

| Category | Technology | Version | Purpose |
| --- | --- | --- | --- |
| Language | Rust | 2021 | Core implementation |
| CLI Framework | Clap | 4.4+ | Argument parsing |
| JSON Parsing | serde_json | 1.0+ | JSON serialization |
| Error Handling | anyhow | 1.0+ | Error management |
| Testing | assert_cmd | 2.0+ | CLI integration tests |

---

## Performance

Benchmarks on a 100MB JSON file (MacBook Pro M2):
- Filter operation: ~250ms
- Select + sort: ~320ms
- Count: ~180ms

Memory usage stays under 50MB even for multi-GB files (streaming mode).

---

## Community & Support

| Channel | Link |
| --- | --- |
| GitHub Repository | https://github.com/acme/json-query |
| Crates.io | https://crates.io/crates/json-query |
| Issue Tracker | https://github.com/acme/json-query/issues |

---

**json-query v0.3.2** — Fast JSON querying from the command line
```

---

## Key Takeaways

### Pattern Recognition
1. **Header badges**: Reflect package registry, language, license
2. **Problem-Solution structure**: Always explain "why this exists"
3. **Quick Start**: 3-step pattern with time estimates
4. **Features**: Group by core vs advanced, use visual indicators
5. **Tech Stack table**: Category, technology, version, purpose
6. **Community section**: Repository, package registry, issues

### Language-Specific Adaptations
- **Python**: Emphasize `uv` over `pip` when applicable
- **TypeScript**: Show npm/pnpm/yarn alternatives
- **Go**: Include Docker deployment if Dockerfile present
- **Rust**: Highlight performance benchmarks and memory efficiency

### Update Strategy
- **New project**: Generate all sections
- **Existing project**: Preserve custom sections, update standard ones
- **Version bump**: Update badges, installation commands, changelog references
