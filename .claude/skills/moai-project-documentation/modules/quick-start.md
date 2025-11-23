# Quick Start: Project Type Selection

This guide helps you identify your project type and understand its documentation focus.

## Five Project Types

### 1. Web Application

**Examples**: SaaS product, web dashboard, REST API backend, web framework app

**Characteristics**:
- User-facing interface (UI/frontend)
- Server-side processing
- Database for persistence
- Real-time collaboration (optional)

**Documentation Focus in product.md**:
- User personas (team lead, individual contributor, customer)
- Adoption targets (e.g., 80% within 2 weeks)
- Integration capabilities (Slack, GitHub, Jira)
- Real-time collaboration features
- Feature adoption timeline

**Architecture Example**:
```
Frontend (React/Vue) ↔ API Layer (FastAPI/Node) ↔ Database (PostgreSQL)
    ↓
WebSocket Server (Real-time)
    ↓
Message Queue (Async jobs)
    ↓
Background Workers
```

**Tech Stack Example**:
```
Frontend: TypeScript, React 18, Vitest, TailwindCSS
Backend: Python 3.13, FastAPI, pytest (85% coverage)
Database: PostgreSQL 15, Alembic migrations
DevOps: Docker, Kubernetes, GitHub Actions
Quality: TypeScript strict, mypy, ruff, code scanning
```

---

### 2. Mobile Application

**Examples**: iOS/Android app, cross-platform app (Flutter/React Native), mobile game, native app

**Characteristics**:
- Mobile OS specific (iOS, Android, or cross-platform)
- Limited resources (battery, network, storage)
- Platform-specific APIs (camera, GPS, contacts)
- App store distribution

**Documentation Focus in product.md**:
- User personas (iOS users, Android users, power users)
- Retention metrics (DAU, MAU, churn rate)
- App store presence (rating target, download goal)
- Offline capability requirements
- Push notification strategy
- Platform-specific features (GPS, camera, contacts)

**Architecture Example**:
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

**Tech Stack Example**:
```
Framework: Flutter 3.13 or React Native 0.72
Language: Dart or TypeScript
Testing: flutter test or Jest, 80%+ coverage
State: Riverpod, Bloc, or Redux
Database: SQLite, Hive, or Realm
HTTP: Dio or Axios wrapper
UI: Material Design or Cupertino
DevOps: Fastlane, GitHub Actions for app store
Quality: flutter analyze, dart format, excellent coverage
Performance: App size <50MB (iOS), startup <2s
```

---

### 3. CLI Tool / Utility

**Examples**: Data validator, deployment tool, package manager, CLI utility, command-line app

**Characteristics**:
- Command-line interface
- Single executable binary (often)
- High performance requirement
- Integration with other tools (pipes, redirects)

**Documentation Focus in product.md**:
- Target workflow (validate → deploy → monitor)
- Performance benchmarks (1M records in <5s)
- Multi-format support (JSON, CSV, Avro)
- Ecosystem adoption (GitHub stars, npm downloads, package manager)
- Integration patterns (Unix pipes, environment variables)

**Architecture Example**:
```
Input Parsing → Command Router → Core Logic → Output Formatter
                                    ↓
                           Validation Layer
                                    ↓
                            Caching Layer
```

**Tech Stack Example**:
```
Language: Go 1.21 or Python 3.13
Testing: Go's built-in or pytest
Packaging: Single binary (Go) or PyPI (Python)
Quality: golangci-lint or ruff, <100MB binary
Performance: <100ms startup time
Distribution: GitHub Releases + Homebrew (Go)
```

---

### 4. Shared Library / SDK

**Examples**: Type validator library, API client, data parser, HTTP client, utility library

**Characteristics**:
- Reusable by other projects
- Well-defined public API
- High test coverage (more critical than apps)
- Type safety important (if applicable)

**Documentation Focus in product.md**:
- API design philosophy (composable, type-safe)
- Developer experience (time-to-first-validation <5 min)
- Performance characteristics (zero-cost abstractions)
- Community engagement (issue response time, contributions)
- Ecosystem adoption (npm downloads, GitHub stars)

**Architecture Example**:
```
Public API Surface
    ↓
Type Guards / Validation
    ↓
Core Logic
    ↓
Platform Adapters (Node.js, Browser, Deno)
```

**Tech Stack Example**:
```
Language: TypeScript 5.2 or Python 3.13
Testing: Vitest or pytest, 90%+ coverage (higher bar for libs)
Package Manager: npm/pnpm or uv
Documentation: TSDoc/JSDoc or Google-style docstrings
Type Safety: TypeScript strict or mypy strict
Bundle: <50KB gzipped, tree-shakeable
```

---

### 5. Data Science / ML Project

**Examples**: Recommendation system, ML pipeline, analytics platform, data science experiment

**Characteristics**:
- Data-driven
- Experimental/iterative
- Model training and inference
- Data quality critical
- Monitoring and versioning important

**Documentation Focus in product.md**:
- Model metrics (accuracy, precision, recall, F1)
- Data quality requirements
- Scalability targets (1B+ records)
- Integration with ML platforms (MLflow, Weights & Biases)
- Feature engineering approach
- Model versioning strategy

**Architecture Example**:
```
Data Ingestion → Feature Engineering → Model Training → Inference
    ↓
Feature Store
    ↓
Model Registry
    ↓
Monitoring & Alerting
```

**Tech Stack Example**:
```
Language: Python 3.13, Jupyter notebooks
ML Framework: scikit-learn, PyTorch, or TensorFlow
Data: pandas, Polars, DuckDB
Testing: pytest, nbval (notebook validation)
Experiment Tracking: MLflow, Weights & Biases
Quality: 80% code coverage, data validation tests
Deployment: FastAPI wrapper or batch pipeline
```

---

## Decision Tree

```
What kind of project is this?
│
├─ Has UI for end users? YES
│  ├─ Mobile (iOS/Android)?
│  │  └─ → Mobile Application ✓
│  │
│  └─ Web-based (browser)?
│     └─ → Web Application ✓
│
├─ Is it a command-line tool?
│  └─ → CLI Tool / Utility ✓
│
├─ Is it for other developers to use?
│  └─ → Shared Library / SDK ✓
│
└─ Is it data/ML focused?
   └─ → Data Science / ML Project ✓
```

---

## Next Steps

After determining your project type:

1. ✅ You are here - understand your project type
2. Use **modules/guides.md** to write product.md, structure.md, tech.md
3. Follow **modules/checklists-examples.md** to see full examples
4. Use **modules/reference.md** for detailed technical patterns
