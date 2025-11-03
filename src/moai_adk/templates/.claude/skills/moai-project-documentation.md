# moai-project-documentation

## Metadata

- **Name**: moai-project-documentation
- **Domain**: Project Documentation & Planning
- **Freedom Level**: high
- **Target Users**: Project owners, architects, tech leads
- **Invocation**: Skill("moai-project-documentation")
- **Progressive Disclosure**: Metadata → Content (full guide) → Resources (examples)

---

## Purpose

Guide interactive creation of three core project documentation files (product.md, structure.md, tech.md) based on project type and user input. Provides templates, examples, checklists, and best practices for each project type (Web App, CLI Tool, Library, Data Science).

---

## Quick Start

**When to use this Skill:**
- During `/alfred:0-project` to generate project documentation
- To refine existing product.md, structure.md, or tech.md
- To understand best practices for project documentation

**Basic Usage:**
```
User: /alfred:0-project
Alfred: Invokes Skill("moai-project-documentation")
Skill: Guides user through project type selection → documentation generation
```

---

## Core Content

### Part 1: Project Type Selection

Ask user to identify their project type:

1. **Web Application**
   - Examples: SaaS, web dashboard, REST API backend
   - Focus: User personas, adoption metrics, real-time features
   - Template style: Business-focused with UX emphasis

2. **Mobile Application**
   - Examples: iOS/Android app, cross-platform app (Flutter, React Native)
   - Focus: User retention, app store metrics, offline capability
   - Template style: UX-driven with platform-specific performance
   - Frameworks: Flutter, React Native, Swift, Kotlin

3. **CLI Tool / Utility**
   - Examples: Data validator, deployment tool, package manager
   - Focus: Performance, integration, ecosystem adoption
   - Template style: Technical with use case emphasis

4. **Shared Library / SDK**
   - Examples: Type validator, data parser, API client
   - Focus: Developer experience, ecosystem adoption, performance
   - Template style: API-first with integration focus

5. **Data Science / ML Project**
   - Examples: Recommendation system, ML pipeline, analytics
   - Focus: Data quality, model metrics, scalability
   - Template style: Metrics-driven with data emphasis

---

### Part 2: Product.md Writing Guide

#### Document Structure

```markdown
# Mission & Strategy
- What problem do we solve?
- Who are the users?
- What's our value proposition?

# Success Metrics
- How do we measure impact?
- What are KPIs?
- How often do we measure?

# Next Features (SPEC Backlog)
- What features are coming?
- How are they prioritized?
```

#### Writing by Project Type

**Web Application Product.md Focus:**
- User personas (team lead, individual contributor, customer)
- Adoption targets (80% within 2 weeks)
- Integration capabilities (Slack, GitHub, Jira)
- Real-time collaboration features

**Mobile Application Product.md Focus:**
- User personas (iOS users, Android users, power users)
- Retention metrics (DAU, MAU, churn rate)
- App store presence (rating target, download goal)
- Offline capability requirements
- Push notification strategy
- Platform-specific features (GPS, camera, contacts)

**CLI Tool Product.md Focus:**
- Target workflow (validate → deploy → monitor)
- Performance benchmarks (1M records in <5s)
- Multi-format support (JSON, CSV, Avro)
- Ecosystem adoption (GitHub stars, npm downloads)

**Library Product.md Focus:**
- API design philosophy (composable, type-safe)
- Developer experience (time-to-first-validation <5 min)
- Performance characteristics (zero-cost abstractions)
- Community engagement (issue response time, contributions)

**Data Science Product.md Focus:**
- Model metrics (accuracy, precision, recall)
- Data quality requirements
- Scalability targets (1B+ records)
- Integration with ML platforms (MLflow, W&B)

---

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

### Part 5: Writing Checklists

#### Product.md Checklist
- [ ] Mission statement is 1-2 sentences
- [ ] Users are specific (not "developers")
- [ ] Problems are ranked by priority
- [ ] Success metrics are measurable
- [ ] Feature backlog has 3-5 next SPECs
- [ ] HISTORY section has v0.1.0

#### Structure.md Checklist
- [ ] Architecture is visualized or clearly described
- [ ] Modules map to git directories
- [ ] External integrations list auth and failure modes
- [ ] Traceability explains TAG system
- [ ] Trade-offs are documented (why this design?)
- [ ] HISTORY section has v0.1.0

#### Tech.md Checklist
- [ ] Primary language with version range specified
- [ ] Quality gates define failure criteria (what blocks merges?)
- [ ] Security policy covers secrets, audits, incidents
- [ ] Deployment describes full release flow
- [ ] Environment profiles (dev/test/prod) included
- [ ] HISTORY section has v0.1.0

---

### Part 6: Common Mistakes to Avoid

❌ **Too Vague**
- "Users are developers"
- "We'll measure success by growth"

✅ **Specific**
- "Solo developers building web apps under time pressure, 3-7 person teams"
- "Measure success by 80% adoption within 2 weeks, 5 features/sprint velocity"

---

❌ **Over-Specified in product.md**
- Function names, database schemas, API endpoints
- "We'll use Redis cache with 1-hour TTL"

✅ **Architecture-Level**
- "Caching layer for performance"
- "Integration with external payment provider"

---

❌ **Inconsistent Across Documents**
- product.md: "5 concurrent users"
- structure.md: "Designed for 10,000 daily users"

✅ **Aligned**
- All 3 documents agree on target scale, user types, quality standards

---

❌ **Outdated**
- Last updated 6 months ago
- HISTORY section has no recent entries

✅ **Fresh**
- HISTORY updated every sprint
- version number incremented on changes

---

## Examples by Project Type

### Example 1: Web App (TaskFlow)

**Product.md excerpt:**
```markdown
## Mission
"Simplify team collaboration through real-time task management."

## Users: Engineering Team Lead
- Manages 5-7 person team
- Needs: Centralized visibility, blocker tracking, code review integration
- Success: Reduce sprint planning from 2h to 45 min
```

**Structure.md excerpt:**
```markdown
## Architecture: Layered
- Frontend: React dashboard with real-time updates (WebSocket)
- API: GraphQL backend with permission checks
- Database: PostgreSQL with event log table
- Cache: Redis for session and frequently accessed data
```

**Tech.md excerpt:**
```markdown
## Primary Language: TypeScript 5.2+
- Rationale: Type safety, strong React tooling, compiled performance

## Quality Gates
- Test coverage: 85% minimum
- Type errors: Zero in strict mode
- Bundle size: <200KB gzipped
```

---

### Example 2: Mobile App (FitTracker)

**Product.md excerpt:**
```markdown
## Mission
"Make fitness tracking simple and engaging with offline-first capabilities."

## Users: Individual Fitness Enthusiast
- Tracks workouts daily, wants quick data entry
- Needs: Offline support (no WiFi at gym), social sharing, integration with Apple Health/Google Fit
- Success: 100K downloads in 3 months, 4.5+ star rating

## Metrics
- DAU (Daily Active Users): 60% of installs
- Session length: Average 5-10 minutes
- Retention: 30-day retention >40%
```

**Structure.md excerpt:**
```markdown
## Architecture: Layered with Offline-First
UI Layer (Screens) → State Management (Riverpod) → Data Layer (SQLite) → Cloud Sync

- UI: Bottom navigation, workout detail screens, statistics charts
- State: Provider-based state (Riverpod) for reactive updates
- Local DB: SQLite with drift ORM for type-safe queries
- Sync: Background sync engine that resolves conflicts
- Native: Permission handling for HealthKit, sensors
```

**Tech.md excerpt:**
```markdown
## Primary Language: Dart 3.0+ (Flutter)
- Rationale: Single codebase for iOS/Android, hot reload, excellent performance

## Quality Gates
- Test coverage: 80% (widget + unit tests)
- App size: <50MB iOS, <60MB Android
- Startup time: <2 seconds
- Frame rate: 60fps (no jank)

## Deployment
- App Store & Google Play via Fastlane
- TestFlight for beta testing
- Version every 2 weeks
```
---

### Example 3: CLI Tool (DataValidate)

**Product.md excerpt:**
```markdown
## Mission
"Validate data integrity before it reaches your database."

## Success Metrics
- Adoption: 10,000+ monthly downloads by month 3
- Performance: Validate 1M records in <5 seconds
- Satisfaction: 4.5+ rating on npm
```

**Structure.md excerpt:**
```markdown
## Architecture: Pipeline
Input Parsing → Schema Loading → Validation Engine → Error Formatting → Output

- Input Layer: Supports JSON, CSV, Avro formats
- Schema Layer: Caches compiled schemas
- Validation: Single-threaded streaming, memory-efficient
- Output: Structured errors with line numbers and suggestions
```

**Tech.md excerpt:**
```markdown
## Primary Language: Go 1.21+
- Rationale: Fast compilation, single binary, native concurrency

## Build
- Binary size: <100MB
- Startup time: <100ms
- Distribution: GitHub Releases + Homebrew
```

---

### Example 4: Library (TypeGuard)

**Product.md excerpt:**
```markdown
## Mission
"Bring TypeScript type safety to runtime with zero overhead."

## Value Prop
- Zero boilerplate (auto-generate guards from interfaces)
- Fast validation (<100μs per object)
- Clear error messages (field-by-field trace)
```

**Structure.md excerpt:**
```markdown
## Modules
- Type Guards: Generate runtime guards from TypeScript types
- Validators: Custom validation functions (email, URL, regex)
- Error Formatter: Detailed error messages with paths
- Platform Adapters: Node.js, Browser, Deno support
```

**Tech.md excerpt:**
```markdown
## Primary Language: TypeScript 5.2+
- Test coverage: 90% (libraries have higher bar)
- Type checking: Zero errors in strict mode
- Bundle: <50KB gzipped, tree-shakeable
```

---

### Example 5: Data Science (ML Pipeline)

**Product.md excerpt:**
```markdown
## Mission
"Predict customer churn with 90%+ accuracy in production."

## Success Metrics
- Model accuracy: 90%+ on test set
- Inference latency: <100ms p95
- Data quality: 99%+ complete features
```

**Structure.md excerpt:**
```markdown
## Pipeline
Data Ingestion → Feature Engineering → Model Training → Validation → Inference

- Data Layer: PostgreSQL + Snowflake for warehousing
- Feature Store: Feast for feature versioning
- Model Registry: MLflow for experiment tracking
- Monitoring: DataDog alerts on model drift
```

**Tech.md excerpt:**
```markdown
## Primary Language: Python 3.13+
- ML Framework: scikit-learn + XGBoost
- Data: pandas + Polars for processing
- Notebooks: Jupyter with nbval validation
- Testing: pytest for pipeline tests (80% coverage)
```

---

## Resources & Templates

### Quick Generator (for project-manager agent)

When invoking from `/alfred:0-project`:

1. **Ask: What's your project type?**
   - Options: Web App, Mobile App, CLI Tool, Library, Data Science

2. **Ask: Who are your primary users?**
   - Text input: Free-form description

3. **Ask: What problem does it solve?**
   - Text input: 1-2 sentence problem statement

4. **Ask: How will you measure success?**
   - Suggestions based on project type:
     - Web App: Adoption rate, feature velocity, user engagement
     - Mobile App: DAU/MAU, app store rating, retention rate, download count
     - CLI: Downloads, performance benchmarks, ecosystem adoption
     - Library: npm downloads, community contributions, GitHub stars
     - Data Science: Model accuracy, data quality, inference latency

5. **Generate**:
   - Create product.md with user inputs
   - Generate structure.md template for their architecture
   - Create tech.md with recommended stack for their type

6. **Offer next step**: `/alfred:1-plan` to create first SPEC

---

## Integration with Alfred Workflow

**When invoked:**
```python
# In project-manager agent during /alfred:0-project
Skill("moai-project-documentation")
  → Ask project type
  → Gather user inputs
  → Generate .moai/project/product.md
  → Generate .moai/project/structure.md
  → Generate .moai/project/tech.md
```

**Output files:**
```
.moai/project/
├── product.md (generated from user input + project type)
├── structure.md (template + user architecture choices)
└── tech.md (recommended stack + quality gates)
```

---

## Versioning & Updates

**When to update this Skill:**
- New programming languages added to MoAI
- New project type examples needed
- Quality gate standards change
- Package management tools evolve

**Current version:** 0.1.0 (2025-11-04)

---

## Further Reading

- EARS requirement authoring: moai-foundation-ears
- SPEC creation guide: moai-spec-authoring
- Project initialization: moai-project-init-flow
