# Complete Writing Guides

Detailed guides for writing product.md, structure.md, and tech.md for each project type.

## product.md Writing Guide

### Document Structure (Template)

```markdown
# Mission & Strategy

## Problem We Solve
[1-2 sentences describing the core problem]

## Target Users
[Specific personas, not "developers"]

## Value Proposition
[Why should users care? What's unique?]

---

# Success Metrics

## Key Performance Indicators (KPIs)
- Metric 1: Target value
- Metric 2: Target value
- Metric 3: Target value

## Measurement Frequency
[Daily/weekly/monthly]

---

# Next Features (SPEC Backlog)

## Backlog
1. **SPEC-001**: [Feature name] - Priority: High
2. **SPEC-002**: [Feature name] - Priority: Medium
3. **SPEC-003**: [Feature name] - Priority: Medium
4. **SPEC-004**: [Feature name] - Priority: Low

---

# HISTORY

**v1.0.0** (2025-11-22)
- 🎯 Initial product definition
- ✨ Target users identified
- ✨ Success metrics defined
```

### Writing by Project Type

**Web Application Product.md Focus:**
- User personas: team lead, individual contributor, customer
- Adoption targets: 80% within 2 weeks
- Integration capabilities: Slack, GitHub, Jira
- Real-time collaboration features
- Performance targets: <2s page load
- Scaling strategy: 1M concurrent users

**Mobile Application Product.md Focus:**
- User personas: iOS users, Android users, power users
- Retention metrics: DAU, MAU, churn rate
- App store presence: rating target (4.5+), download goal (100K+)
- Offline capability requirements
- Push notification strategy: frequency, personalization
- Platform-specific features: GPS, camera, contacts

**CLI Tool Product.md Focus:**
- Target workflow: validate → deploy → monitor
- Performance benchmarks: 1M records in <5s
- Multi-format support: JSON, CSV, Avro
- Ecosystem adoption: GitHub stars (500+), npm downloads (10K/month)
- Integration with existing tools: pipe compatibility

**Library Product.md Focus:**
- API design philosophy: composable, type-safe
- Developer experience: time-to-first-validation <5 min
- Performance characteristics: zero-cost abstractions
- Community engagement: issue response time <24h, contributions
- Adoption targets: 10K npm downloads/month

**Data Science Product.md Focus:**
- Model metrics: accuracy (95%+), precision, recall, F1
- Data quality requirements: <1% missing values, outlier handling
- Scalability targets: 1B+ records processed daily
- Integration with ML platforms: MLflow, Weights & Biases
- Inference latency: <100ms per prediction

---

## structure.md Writing Guide

### Document Structure (Template)

```markdown
# System Architecture

## Design Pattern
[Describe overall approach: microservices, monolith, serverless, etc.]

## Layers/Tiers
[Describe how layers communicate]

## Architecture Diagram
[ASCII or visual representation]

---

# Core Modules

## Module 1: [Name]
- **Responsibility**: [What does it do?]
- **Files**: [Location in repo]
- **Dependencies**: [What does it depend on?]

## Module 2: [Name]
[Same structure as Module 1]

---

# External Integrations

## Integration 1: [System Name]
- **Purpose**: [What do we use it for?]
- **Authentication**: [OAuth, API key, etc.]
- **Failure Mode**: [What happens if it fails?]

---

# Traceability

## SPEC to Code Mapping
- SPEC-001 → src/features/feature-name/
- SPEC-002 → src/features/another-feature/

---

# Trade-offs

## Decision 1: [Decision Name]
- **Choice**: [We chose X]
- **Why**: [Why did we choose it?]
- **Trade-off**: [What did we sacrifice?]

---

# HISTORY

**v1.0.0** (2025-11-22)
- 🏗️ Initial architecture defined
- 📋 Core modules identified
```

### Architecture Patterns by Project Type

**Web Application Architecture:**
```
Frontend (React/Vue)
    ↓
API Gateway (Rate limiting, auth)
    ↓
API Layer (FastAPI/Node)
    ↓
Business Logic Layer
    ↓
Database Layer (PostgreSQL)
    ↓
WebSocket Server (Real-time) + Message Queue + Background Workers
```

**Mobile Application Architecture:**
```
UI Presentation Layer (Widgets/Screens)
    ↓
State Management (Redux/Bloc/Riverpod)
    ↓
Repository Pattern (Data access abstraction)
    ↓
Local Database (SQLite/Realm) + Remote API (REST/GraphQL)
    ↓
Authentication Module + Native Modules (Camera, GPS)
    ↓
Offline Sync Engine
```

**CLI Tool Architecture:**
```
Command Parser (Parse arguments)
    ↓
Command Router (Route to handler)
    ↓
Input Validation (Type check, format)
    ↓
Core Logic (Business logic)
    ↓
Caching Layer (Optional: cache results)
    ↓
Output Formatter (JSON, CSV, table)
```

**Library Architecture:**
```
Public API Surface (Exported functions/classes)
    ↓
Type Guards / Validation (Ensure type safety)
    ↓
Core Logic (Implementation)
    ↓
Platform Adapters (Node.js, Browser, Deno)
    ↓
Error Handling (Consistent error types)
```

**Data Science Architecture:**
```
Data Ingestion Layer
    ↓
Data Validation & Cleaning
    ↓
Feature Engineering Layer
    ↓
Model Training Pipeline
    ↓
Model Evaluation & Validation
    ↓
Model Registry (Version control)
    ↓
Inference Engine
    ↓
Monitoring & Alerting
```

---

## tech.md Writing Guide

### Document Structure (Template)

```markdown
# Technology Stack

## Primary Language
- **Language**: [Name and version]
- **Version Range**: [e.g., 3.13+]
- **Why This Choice**: [Rationale]

## Framework / Runtime
- **Framework**: [Name and version]
- **Alternative**: [What else we considered]

---

# Quality Gates

## Test Coverage
- **Required**: 85% minimum
- **Tool**: pytest / Jest / vitest
- **Enforcement**: Pre-commit hook, CI/CD gate

## Type Safety
- **Standard**: strict mode (TypeScript) / mypy strict (Python)
- **No Error Tolerance**: Zero type errors before merge

## Linting & Formatting
- **Tools**: ruff, black, prettier, ESLint
- **Enforcement**: Pre-commit hook

---

# Security Policy

## Secret Management
- **Tool**: dotenv (development), GitHub Secrets (CI/CD)
- **Pattern**: Environment variables, never hardcoded

## Vulnerability Handling
- **Scanning**: Dependabot, GitHub security scanning
- **SLA**: Critical = fix within 24 hours

## Incident Response
- **Reporting**: [How to report security issues]
- **Response Time**: [SLA for security patches]

---

# Deployment Strategy

## Environments
- **Development**: Local machine
- **Staging**: Pre-production (optional)
- **Production**: User-facing

## Release Process
1. [Step 1]
2. [Step 2]
3. [Step 3]

## Rollback Procedure
- [How to rollback if something breaks]

---

# HISTORY

**v1.0.0** (2025-11-22)
- 🔧 Initial tech stack defined
- ✅ Quality gates established
```

### Tech Stack Examples by Type

**Web Application Stack:**
```
Frontend: TypeScript 5.2+, React 18+, Vitest, TailwindCSS
Backend: Python 3.13+, FastAPI, pytest (85%+ coverage)
Database: PostgreSQL 15+, Alembic migrations
DevOps: Docker, Kubernetes, GitHub Actions
Quality: TypeScript strict, mypy strict, ruff, GitHub code scanning
Performance: <200KB gzipped bundle, <2s page load
```

**Mobile Application Stack:**
```
Framework: Flutter 3.13+ or React Native 0.72+
Language: Dart or TypeScript
Testing: flutter test or Jest, 80%+ coverage
State: Riverpod, Bloc, or Redux
Database: SQLite 3, Hive, or Realm
HTTP: Dio 5.0+ or Axios
UI: Material Design 3 or Cupertino
DevOps: Fastlane 2.200+, GitHub Actions
Quality: flutter analyze, dart format excellent coverage
Performance: App size <50MB, startup <2s
```

**CLI Tool Stack:**
```
Language: Go 1.21+ or Python 3.13+
Testing: Go built-in testing or pytest, 85%+ coverage
Package: Single binary (Go) or PyPI (Python)
Quality: golangci-lint or ruff, <100MB binary
Performance: <100ms startup time
Distribution: GitHub Releases + Homebrew (Go)
CI/CD: GitHub Actions for cross-platform builds
```

**Library Stack:**
```
Language: TypeScript 5.2+ or Python 3.13+
Testing: Vitest or pytest, 90%+ coverage (libraries have higher bar)
Package: npm/pnpm or PyPI + uv
Documentation: TSDoc/JSDoc or Google-style docstrings
Type: TypeScript strict or mypy strict
Bundle: <50KB gzipped, tree-shakeable
CI/CD: GitHub Actions for testing, linting, publishing
```

**Data Science Stack:**
```
Language: Python 3.13+, Jupyter 7.0+
ML Framework: scikit-learn 1.3+, PyTorch 2.0+, or TensorFlow 2.13+
Data: pandas 2.0+, Polars 0.19+, DuckDB 0.8+
Testing: pytest 7.0+, nbval for notebooks
Tracking: MLflow 2.0+ or Weights & Biases
Quality: 80%+ code coverage, data validation tests
Deployment: FastAPI wrapper or batch pipeline
```

---

## Common Patterns Across All Types

### Pre-commit Hooks Example

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/psf/black
    rev: 23.0.0
    hooks:
      - id: black

  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix]

  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v3.0.0
    hooks:
      - id: prettier
```

### GitHub Actions CI/CD Pattern

```yaml
# .github/workflows/quality.yml
name: Quality Gates

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: pytest --cov=src/ --cov-fail-under=85
      - name: Lint
        run: ruff check src/
      - name: Type check
        run: mypy src/
```
