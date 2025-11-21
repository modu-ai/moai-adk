---
name: moai-project-documentation-structure-tech
description: Structure.md and Tech.md writing guides with architecture patterns
---

## Structure.md & Tech.md Writing Guides

### Part 3: Structure.md Writing Guide

#### Document Structure

```markdown
# System Architecture
- What's the overall design pattern?
- What layers/tiers exist?
- How do components interact?

# Core Modules
- What are the main building blocks?
- What's each module responsible for?
- How do they communicate?

# External Integrations
- What external systems do we depend on?
- How do we authenticate?
- What's our fallback strategy?

# Traceability
- How do SPECs map to code?
- How do we trace changes?
```

#### Architecture Patterns by Project Type

**Web Application Architecture:**
```
Frontend (React/Vue) ↔ API Layer (FastAPI/Node) ↔ Database (PostgreSQL)
    ↓
WebSocket Server (Real-time features)
    ↓
Message Queue (Async jobs)
    ↓
Background Workers
```

**Mobile Application Architecture:**
```
UI Layer (Screens, Widgets)
    ↓
State Management (Bloc, Redux, Riverpod)
    ↓
Data Layer (Local DB: SQLite/Realm, Remote: REST/GraphQL)
    ↓
Authentication (OAuth, JWT)
    ↓
Native Modules (Camera, GPS, Contacts)
    ↓
Offline Sync Engine
```

**CLI Tool Architecture:**
```
Input Parsing → Command Router → Core Logic → Output Formatter
                                    ↓
                           Validation Layer
                                    ↓
                            Caching Layer
```

**Library Architecture:**
```
Public API Surface
    ↓
Type Guards / Validation
    ↓
Core Logic
    ↓
Platform Adapters (Node.js, Browser, Deno)
```

**Data Science Architecture:**
```
Data Ingestion → Feature Engineering → Model Training → Inference
    ↓
Feature Store
    ↓
Model Registry
    ↓
Monitoring & Alerting
```

---

### Part 4: Tech.md Writing Guide

#### Document Structure

```markdown
# Technology Stack
- What language(s)?
- What version ranges?
- Why these choices?

# Quality Gates
- What's required to merge?
- How do we measure quality?
- What tools enforce standards?

# Security Policy
- How do we manage secrets?
- How do we handle vulnerabilities?
- What's our incident response?

# Deployment Strategy
- Where do we deploy?
- How do we release?
- How do we rollback?
```

#### Tech Stack Examples by Type

**Web Application:**
```
Frontend: TypeScript, React 18, Vitest, TailwindCSS
Backend: Python 3.13, FastAPI, pytest (85% coverage)
Database: PostgreSQL 15, Alembic migrations
DevOps: Docker, Kubernetes, GitHub Actions
Quality: TypeScript strict mode, mypy, ruff, GitHub code scanning
```

**Mobile Application:**
```
Framework: Flutter 3.13 or React Native 0.72
Language: Dart or TypeScript
Testing: flutter test or Jest, 80%+ coverage
State Management: Riverpod, Bloc, or Redux
Local Database: SQLite, Hive, or Realm
HTTP Client: Dio or Axios wrapper
UI: Material Design or Cupertino
DevOps: Fastlane, GitHub Actions for app store deployment
Quality: flutter analyze, dart format, excellent test coverage
Performance: App size <50MB (iOS), startup <2s
```

**CLI Tool:**
```
Language: Go 1.21 or Python 3.13
Testing: Go's built-in testing or pytest
Packaging: Single binary (Go) or PyPI (Python)
Quality: golangci-lint or ruff, <100MB binary
Performance: <100ms startup time
```

**Library:**
```
Language: TypeScript 5.2 or Python 3.13
Testing: Vitest or pytest, 90%+ coverage (libraries = higher bar)
Package Manager: npm/pnpm or uv
Documentation: TSDoc/JSDoc or Google-style docstrings
Type Safety: TypeScript strict or mypy strict
```

**Data Science:**
```
Language: Python 3.13, Jupyter notebooks
ML Framework: scikit-learn, PyTorch, or TensorFlow
Data: pandas, Polars, DuckDB
Testing: pytest, nbval (notebook validation)
Experiment Tracking: MLflow, Weights & Biases
Quality: 80% code coverage, data validation tests
```

---

**End of Module** | moai-project-documentation-structure-tech
