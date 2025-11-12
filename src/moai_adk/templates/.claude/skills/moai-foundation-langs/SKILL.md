---
name: moai-foundation-langs
version: 4.0.0
created: 2025-10-22
updated: 2025-11-12
status: active
tier: Foundation
description: >-
  Complete programming language support matrix with 25+ languages (Python, JavaScript, 
  Go, Rust, PHP, Java, TypeScript, C++, C#, etc.) covering November 2025 stable versions. 
  Includes language detection, version management, best practices, and framework ecosystem 
  guidance for modern development.
keywords:
  - language
  - detection
  - framework
  - ecosystem
  - best-practices
  - version-management
  - polyglot
  - november-2025-stable
allowed-tools:
  - Read
  - Bash
  - Glob
---

# Foundation Langs Skill (Enterprise v4.0.0)

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-langs |
| **Version** | 4.0.0 (2025-11-12 Enterprise Release) |
| **Tier** | Foundation (Core Language Support) |
| **Allowed tools** | Read, Bash, Glob |
| **Auto-load** | Language-related discussions, version queries, framework decisions |
| **Language Coverage** | 25+ programming languages |
| **Latest Stable Stack** | November 2025 |

---

## What It Does

Provides comprehensive language and framework guidance for 25+ programming languages with November 2025 stable versions, automatic project language detection, version management strategies, ecosystem best practices, and migration paths.

**Key Capabilities**:
- Auto-detect project language from package.json, pyproject.toml, go.mod, Cargo.toml, etc.
- Version management for 25+ languages (latest stable as of November 2025)
- Framework selection guidance and ecosystem navigation
- Best practices enforcement for each language
- TRUST 5 principles integration (Language-specific patterns)
- TDD workflow support with language-specific test frameworks
- Migration and upgrade path guidance
- Security best practices per language

---

## When to Use

**Automatic Triggers**:
- Language selection for new projects
- Framework architecture decisions
- Version upgrade planning
- Dependency management strategies
- Test framework selection
- Code quality tooling setup

**Manual Invocation**:
- Language best practices clarification
- Framework ecosystem navigation
- Version compatibility questions
- Testing strategy selection
- Performance optimization patterns
- Security hardening guidance

---

## Supported Languages (25+)

### Tier 1: Enterprise Primary (November 2025 Stable)

| Language | Current Version | Release Date | LTS/Support | Recommended Use |
|----------|-----------------|--------------|-------------|-----------------|
| **Python** | 3.13.x | Oct 2024 | 2029-10-01 | AI/ML, Data Science, Web, Automation |
| **JavaScript** | Node.js 22.x | Apr 2024 | 2027-04-30 | Web (Frontend/Backend), Real-time Apps |
| **TypeScript** | 5.9.x | Aug 2025 | Rolling | Type-safe JavaScript Applications |
| **Go** | 1.25.x | Aug 2025 | N/A | Cloud, Microservices, DevOps |
| **Rust** | 1.91.x | Oct 2025 | Rolling | Systems, Security, Web Assembly |
| **Java** | 25 LTS | Sep 2025 | 2033-09-30 | Enterprise, Android, Backend |
| **PHP** | 8.4.x | Nov 2024 | 2026-12-31 | Web Development, CMS, Rapid Dev |
| **C++** | C++23 (ISO 2024) | Oct 2024 | N/A | Systems, Gaming, High Performance |
| **C#** | 13.0 | Nov 2024 | Rolling | .NET Applications, Enterprise |
| **Kotlin** | 2.1.x | Sep 2025 | Rolling | JVM, Android Native Development |

### Tier 2: Cloud Native & Modern Web

| Language | Current Version | Key Strengths |
|----------|-----------------|---------------|
| **Elixir** | 1.18.x | Distributed Systems, Fault Tolerance |
| **Scala** | 3.6.x | JVM, Functional Programming |
| **Clojure** | 1.12.x | LISP, Functional Paradigm |
| **Ruby** | 3.4.x | Web Development (Rails), Rapid Prototyping |
| **Swift** | 6.1.x | iOS/macOS, Apple Ecosystem |
| **Kotlin** | 2.1.x | Null Safety, Modern JVM Language |
| **F#** | 8.0.x | .NET, Functional Programming |

### Tier 3: Specialized & Emerging

| Language | Current Version | Niche |
|----------|-----------------|-------|
| **Zig** | 0.14.x | Low-level Systems Programming |
| **Julia** | 1.11.x | Numerical Computing, Science |
| **R** | 4.4.x | Statistical Computing, Data Analysis |
| **MATLAB** | R2025a | Engineering, Numerical Computing |
| **Lua** | 5.4.x | Embedded Scripting, Game Development |
| **Dart** | 3.6.x | Flutter Mobile Apps |
| **Haskell** | 9.14.x | Pure Functional Programming |

---

## Quick Language Selection Guide

### By Use Case

**Web Applications** (Full-stack):
- Tier 1: TypeScript/Node.js (fastest), Python+Django/FastAPI
- Alternative: Go (high performance), Ruby on Rails (rapid)

**Data Science & AI/ML**:
- Primary: Python 3.13 (TensorFlow, PyTorch, scikit-learn)
- Secondary: Julia (numerical computing)

**Systems & Performance**:
- Primary: Rust 1.91 (memory safety), Go 1.25 (concurrency)
- Alternative: C++ (C++23), Zig

**Cloud & DevOps**:
- Primary: Go 1.25 (container tooling, Kubernetes)
- Alternative: Rust (secure infrastructure), Python (automation)

**Mobile Applications**:
- iOS/macOS: Swift 6.1
- Android: Kotlin 2.1 (native), TypeScript/React Native (cross-platform)
- Cross-platform: Dart 3.6 (Flutter)

**Enterprise Systems**:
- Primary: Java 25 LTS, C# 13 (.NET 9)
- Alternative: Go (microservices)

---

## Language Detection Patterns

### Automatic Detection Strategy

```
Project Root Analysis
    ↓
1. Check package.json (Node.js/JavaScript/TypeScript)
2. Check pyproject.toml (Python 3.8+)
3. Check Cargo.toml (Rust)
4. Check go.mod (Go)
5. Check pom.xml / build.gradle (Java/Kotlin)
6. Check .csproj (C#/.NET)
7. Check composer.json (PHP)
8. Check Gemfile (Ruby)
9. Check pubspec.yaml (Dart/Flutter)
10. Check package.php (PHP Legacy)
    ↓
Detected Language + Version
```

### Key Configuration Files

| Language | Primary Config | Secondary | Version Marker |
|----------|----------------|-----------|----------------|
| JavaScript | package.json | .nvmrc | engines.node |
| Python | pyproject.toml | setup.py | python_requires |
| Go | go.mod | go.sum | go version |
| Rust | Cargo.toml | rust-toolchain.toml | edition |
| Java | pom.xml | build.gradle | <source> tag |
| PHP | composer.json | php.ini | "php" dependency |
| C# | .csproj | global.json | <TargetFramework> |
| Ruby | Gemfile | .ruby-version | ruby |
| Kotlin | build.gradle.kts | pom.xml | kotlin version |
| Dart | pubspec.yaml | .dart_tool | sdk constraint |

---

## Version Management Strategies

### Strategy 1: Lock File Management

**Best Practice**:
- Commit lock files (package-lock.json, Cargo.lock, go.sum, poetry.lock)
- Use deterministic builds across environments
- Implement regular vulnerability scanning

**Per Language**:
```
JavaScript: package-lock.json (npm) OR yarn.lock (yarn) OR pnpm-lock.yaml
Python: poetry.lock (recommended) OR requirements.txt
Go: go.sum (automatic)
Rust: Cargo.lock (if binary)
Java: maven lock (pom.xml) OR gradle.lock (experimental)
PHP: composer.lock (required)
```

### Strategy 2: Version Pinning Policies

**Development Dependencies**:
```
Allow: ^2.0.0 (compatible versions)
Lock patch: 2.1.3 for critical tools
```

**Production Dependencies**:
```
Pin major.minor: ^2.1.0 (safe)
Pin full version: 2.1.3 (maximum safety)
Quarterly review: Check for updates, security patches
```

**Critical Dependencies** (security-sensitive):
```
Approach: Full version pinning (2.1.3)
Review: Monthly security advisories
Testing: Full test suite before patch updates
```

### Strategy 3: Automated Dependency Updates

**Recommended Tooling by Language**:
- JavaScript: Dependabot (GitHub), Renovate (flexible)
- Python: Dependabot, pyup.io, pip-audit (security)
- Go: Dependabot, Renovate
- Rust: Dependabot, cargo-audit (security)
- Java: Dependabot, snyk (security)
- PHP: Composer advisories, snyk

**Configuration**:
```
Frequency: Weekly scans, review PRs daily
Security patches: Auto-merge low-risk, fast track
Major updates: Require testing, manual merge
```

---

## Framework Ecosystem Selection

### JavaScript/TypeScript Ecosystem (Node.js 22.x)

**Web Frameworks**:
- **Express.js** (v4.21.x): Minimal, battle-tested, largest ecosystem
- **Fastify** (v4.28.x): Performance-focused, modern
- **NestJS** (v10.4.x): Enterprise patterns, TypeScript-first

**React Ecosystem**:
- **React 19.x**: Component model, server components
- **Next.js 15.x**: Meta-framework (SSR, SSG, Edge)
- **TanStack Query 5.x**: Async state management

**API/Microservices**:
- **Fastify 4.x**: High-performance HTTP
- **tRPC**: Type-safe RPC without schemas
- **GraphQL with Apollo Server 4.x**

**Testing**:
- **Vitest 2.x**: Jest-compatible, Vite-native (recommended)
- **Jest 29.x**: Industry standard (legacy)
- **Playwright 1.48.x**: Browser automation

### Python Ecosystem (3.13.x)

**Web Frameworks**:
- **FastAPI 0.115.x**: Modern async-first (recommended for new projects)
- **Django 5.1.x**: Full-featured, batteries-included
- **Flask 3.1.x**: Lightweight, learning-friendly

**Data Science Stack**:
- **NumPy 2.1.x**: Numerical computing foundation
- **Pandas 2.2.x**: Data manipulation (migrating to Polars for scale)
- **Scikit-learn 1.5.x**: ML algorithms
- **PyTorch 2.5.x** (recommended) OR **TensorFlow 2.17.x**: Deep learning

**Testing**:
- **Pytest 8.3.x**: Standard test framework (industry standard)
- **Coverage.py 7.6.x**: Code coverage tracking

**Package Management**:
- **uv** (Rust-based, 0.4.x): Fastest, recommended for MoAI projects
- **Poetry 1.8.x**: Lock file management (traditional)
- **pip** (built-in): Simple, reliable

### Go Ecosystem (1.25.x)

**Web Frameworks**:
- **Gin 1.10.x**: High-performance router
- **Fiber 3.0.x**: Express.js-like for Go
- **Echo 4.12.x**: Minimal, flexible

**Testing**:
- **testing** (stdlib): Built-in, sufficient for most cases
- **Testify 1.9.x**: Assertions, mocking
- **GoConvey**: BDD-style testing

**ORM/Database**:
- **GORM 1.25.x**: Most popular ORM
- **Sqlc 1.27.x**: Type-safe SQL

**Cloud/Deployment**:
- **Kubernetes**: Native support (go-client)
- **Docker**: Official support in stdlib

### Rust Ecosystem (1.91.x)

**Web Frameworks**:
- **Actix-web 4.9.x**: High-performance, actor-based
- **Axum 0.8.x**: Modern async (Tokio-based, recommended)
- **Rocket 0.5.x**: Developer-friendly ergonomics

**Async Runtime**:
- **Tokio 1.41.x**: De facto standard async runtime
- **async-std 1.13.x**: Alternative stdlib-like

**Testing**:
- **Criterion 0.5.x**: Benchmarking
- **Proptest 1.4.x**: Property-based testing

**Security**:
- **Rustls 0.23.x**: TLS without OpenSSL bindings
- **Ring 0.17.x**: Cryptography primitives

---

## Best Practices by Language

### Python 3.13 Best Practices

**Code Quality**:
```
Linting: Ruff (Rust-based, 0.8.x) [recommended: replaces black+isort+flake8]
Type checking: Pyright 1.1.x (VS Code native)
Formatting: Black 24.10.x OR Ruff format
```

**Testing**:
```
Framework: Pytest 8.3.x (standard)
Coverage: Coverage.py 7.6.x (≥85% target)
Async testing: pytest-asyncio 0.24.x
```

**Package Structure**:
```
src/mypackage/
  ├─ __init__.py
  ├─ core.py
  └─ __main__.py
tests/
  ├─ test_core.py
  └─ conftest.py
pyproject.toml (modern: replaces setup.py)
```

### JavaScript/TypeScript Best Practices (Node.js 22.x)

**Code Quality**:
```
Linting: ESLint 9.x with TypeScript parser
Formatting: Prettier 3.3.x (non-negotiable for team)
Type checking: TypeScript 5.9.x compiler
```

**Testing Framework Selection**:
```
New projects: Vitest 2.x (recommended: faster than Jest)
Existing Jest: Continue (large ecosystem)
E2E: Playwright 1.48.x (Selenium replacement)
```

**Package Structure**:
```
src/
  ├─ index.ts
  ├─ services/
  └─ types/
tests/
  ├─ unit/
  ├─ integration/
  └─ e2e/
package.json
tsconfig.json
vitest.config.ts
```

### Go Best Practices (1.25.x)

**Code Quality**:
```
Linting: golangci-lint 1.61.x
Formatting: gofmt (stdlib) + goimports
Vet: go vet (stdlib)
```

**Testing**:
```
Framework: testing (stdlib, sufficient)
Assertions: Testify 1.9.x (optional)
Table-driven tests: Standard Go pattern
```

**Module Structure**:
```
project/
  ├─ main.go
  ├─ internal/
  │  ├─ handler/
  │  ├─ service/
  │  └─ repository/
  ├─ pkg/
  │  └─ domain/
  ├─ go.mod
  └─ go.sum
```

### Rust Best Practices (1.91.x)

**Code Quality**:
```
Linting: Clippy (built-in, cargo clippy)
Formatting: Rustfmt (built-in, cargo fmt)
Security: cargo-audit 0.20.x
```

**Testing**:
```
Unit tests: #[cfg(test)] modules
Integration: tests/ directory (separate binary)
Benchmarks: Criterion 0.5.x
Property testing: Proptest 1.4.x
```

**Cargo.toml Structure**:
```
[package]
[dependencies]
[dev-dependencies]
[build-dependencies]

[[bin]]    # Multiple binaries
[[example]] # Examples
```

---

## Security Best Practices (Language-Agnostic)

### Dependency Security Audit

**JavaScript/TypeScript**:
```bash
npm audit
npm audit fix
yarn audit
pnpm audit
# Recommended: Snyk (snyk test)
```

**Python**:
```bash
pip-audit
poetry check --lock
safety check
# Recommended: Bandit (bandit -r src/)
```

**Go**:
```bash
go list -json all | nancy sleuth
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

**Rust**:
```bash
cargo audit
cargo audit --deny warnings
# Recommended: RUSTSEC database
```

### Version Constraints for Security

| Language | Strategy |
|----------|----------|
| **Python** | Major.Minor pinning: requests==2.31.* |
| **JavaScript** | Exact pinning for security tools: eslint==9.0.0 |
| **Go** | Module required, semantic versioning v1.2.3 |
| **Rust** | Cargo.lock for binaries, flexible for libraries |

---

## Migration & Upgrade Paths

### Python 2.7 → 3.13 Migration

```
Status: Python 2.7 EOL 2020-01-01 (5+ years outdated)
Recommended: Migrate to Python 3.12+ immediately

Path:
  1. 2.7 (EOL, danger)
  2. 3.8 (EOL 2024-10, minimal target)
  3. 3.11 (Security fixes 2027-10)
  4. 3.13 (Current, 2029-10 support)
```

### Node.js 16 → Node.js 22 Migration

```
Status: Node.js 16 EOL 2023-09-11
Recommended: Upgrade to Node.js 22 LTS (support until 2027-04)

Path:
  1. 16.x (EOL, unmaintained) → 18.x (EOL 2025-04) → 22.x (LTS 2027)
  2. Test thoroughly: Verify breaking changes per version
  3. Update lockfiles: Fresh npm install after node update
```

### Go 1.19 → 1.25 Migration

```
Status: Go 1.19 EOL 2023-12-01
Recommended: Upgrade to Go 1.25 (released Aug 2025)

Path:
  1. Update go.mod: go 1.25 (top of file)
  2. Run: go mod tidy (resolve dependencies)
  3. Test: go test ./...
  4. Verify: go vet ./...
```

---

## Troubleshooting Language Issues

### Issue 1: Version Mismatch Between CI/CD and Local

**Symptoms**: Tests pass locally, fail in CI

**Root Cause**: .nvmrc, .python-version, or go.mod mismatch

**Solution**:
```bash
# Verify version file exists
ls -la .nvmrc .python-version go.mod

# Pin versions explicitly
echo "3.13.0" > .python-version
echo "22.11.0" > .nvmrc
```

### Issue 2: Dependency Conflict (Circular, Version Incompatibility)

**Symptoms**: Install fails with conflict messages

**Solution by Language**:
```bash
# JavaScript
npm install --legacy-peer-deps (temporary workaround)
npm ls (show dependency tree)

# Python
pip install --use-deprecated=legacy-resolver (old behavior)
poetry show --tree

# Go
go get -u (upgrade all)
go mod why -m <module> (debug why included)

# Rust
cargo tree -i (inverse dependency tree)
```

### Issue 3: Security Vulnerability in Dependency

**Immediate Actions**:
```bash
# 1. Identify severity
npm audit | grep high/critical

# 2. Update (patch first)
npm update <package>

# 3. Lock new version
npm install --save-exact <package>@<safe-version>

# 4. Test thoroughly
npm test
```

---

## Official Documentation Links (November 2025)

### Tier 1: Enterprise Languages

- Python 3.13 Docs: https://docs.python.org/3/
- Node.js 22 Docs: https://nodejs.org/docs/latest-v22.x/api/
- TypeScript 5.9 Docs: https://www.typescriptlang.org/docs/
- Go 1.25 Docs: https://go.dev/doc/
- Rust 1.91 Docs: https://doc.rust-lang.org/stable/
- Java 25 Docs: https://docs.oracle.com/en/java/javase/25/
- PHP 8.4 Docs: https://www.php.net/docs.php
- C++23 Standard: https://isocpp.org/std/the-standard
- C# 13 Docs: https://learn.microsoft.com/en-us/dotnet/csharp/
- Kotlin 2.1 Docs: https://kotlinlang.org/docs/home.html

### Framework & Ecosystem Links

**JavaScript**:
- Express.js: https://expressjs.com/
- Next.js 15: https://nextjs.org/docs
- NestJS: https://docs.nestjs.com/

**Python**:
- FastAPI: https://fastapi.tiangolo.com/
- Django 5.1: https://docs.djangoproject.com/en/5.1/
- Pytest: https://docs.pytest.org/

**Go**:
- Go Standard Library: https://pkg.go.dev/std
- Gin Web Framework: https://gin-gonic.com/docs/
- GORM: https://gorm.io/docs/

**Rust**:
- Rust Book: https://doc.rust-lang.org/book/
- Tokio Async: https://tokio.rs/
- Axum Web Framework: https://github.com/tokio-rs/axum

---

## Progressive Disclosure Levels

### Level 1: Quick Selection (Beginner)
- Which language should I use?
- What's the current stable version?
- How do I install?

### Level 2: Framework & Tooling (Intermediate)
- Framework selection guidance
- Testing strategy per language
- Version management best practices

### Level 3: Advanced Topics (Expert)
- Performance optimization per language
- Security hardening strategies
- Large-scale architecture patterns
- Migration pathways

---

## Related Skills

- `moai-alfred-best-practices`: TRUST 5 principles per language
- `moai-domain-web-api`: JavaScript/TypeScript web development
- `moai-domain-python-expert`: Python-specific guidance
- `moai-domain-devops`: Go/Rust for infrastructure
- `moai-foundation-tags`: TAG system across all languages

---

## Summary

**moai-foundation-langs** (Enterprise v4.0.0) provides:
- ✓ 25+ programming languages with November 2025 stable versions
- ✓ Automatic project language detection
- ✓ Framework ecosystem guidance
- ✓ Version management strategies (lock files, pinning, updates)
- ✓ Language-specific best practices
- ✓ Security audit procedures
- ✓ Migration and upgrade paths
- ✓ Official documentation links
- ✓ TRUST 5 principles integration
- ✓ TDD workflow support

**Use when**: Starting new projects, selecting frameworks, upgrading versions, or clarifying language best practices.
