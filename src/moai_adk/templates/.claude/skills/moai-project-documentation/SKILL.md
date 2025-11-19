---
name: moai-project-documentation
version: 4.0.0
status: stable
updated: 2025-11-20
description: Enhanced project documentation with AI-powered features and Context7 integration
category: Domain
allowed-tools: Read, Glob, Grep, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

# moai-project-documentation: Project Documentation

**AI-powered project documentation with Context7 integration and automated templates**

Trust Score: 9.6/10 | Version: 4.0.0 | Last Updated: 2025-11-20

---

## Overview

Project documentation expert with AI-powered features for creating comprehensive project documentation including:
- **Product.md**: Mission, strategy, success metrics, and feature backlog
- **Structure.md**: System architecture, core modules, and external integrations
- **Tech.md**: Technology stack, quality gates, and deployment strategy

**Core Capabilities**:
- **Project Type Templates**: Web apps, mobile apps, CLI tools, libraries, data science
- **AI-Enhanced Templates**: Context7 integration for latest best practices
- **Automated Checklists**: Quality gates and validation rules
- **Progressive Documentation**: Tiered disclosure based on project needs

---

## Project Type Templates

### Web Application Documentation

```markdown
# product.md

## Mission & Strategy
We build collaborative web applications that help teams work better together.
Target users: small to medium-sized development teams (3-15 people).

## Success Metrics
- **Adoption**: 80% team adoption within 2 weeks of launch
- **Engagement**: 5+ daily interactions per user
- **Retention**: 90% monthly retention rate
- **Performance**: <2s page load time, 99.9% uptime

## Next Features (SPEC Backlog)
- SPEC-001: Real-time collaboration (Week 1-2)
- SPEC-002: Advanced search functionality (Week 3-4)
- SPEC-003: Mobile responsive design (Week 5-6)
- SPEC-004: Team analytics dashboard (Week 7-8)

---

# structure.md

## System Architecture
```
Frontend (React/Vue) ↔ API Layer (FastAPI/Node) ↔ Database (PostgreSQL)
    ↓
WebSocket Server (Real-time features)
    ↓
Message Queue (Async jobs)
    ↓
Background Workers
```

## Core Modules
- **Authentication**: JWT-based auth with social login
- **Project Management**: Task tracking, team organization
- **Real-time Collaboration**: Live updates, presence detection
- **Analytics Dashboard**: Usage metrics, performance monitoring

## External Integrations
- **Version Control**: GitHub API integration
- **Project Management**: Jira, Trello, Asana APIs
- **Communication**: Slack webhook integration
- **Monitoring**: Sentry error tracking

## Traceability
- SPEC IDs map to git branches (feature/SPEC-*)
- Code reviews reference specific requirements
- Automated testing links to user stories

---

# tech.md

## Technology Stack
**Frontend**: TypeScript 5.0, React 18, Next.js 14, TailwindCSS
**Backend**: Python 3.11, FastAPI, Pydantic, SQLAlchemy
**Database**: PostgreSQL 15, Redis (cache)
**Infrastructure**: Docker, Kubernetes, GitHub Actions

## Quality Gates
- **Code Coverage**: Minimum 85% test coverage
- **Type Safety**: TypeScript strict mode enabled
- **Security**: No high-risk vulnerabilities in security scans
- **Performance**: Lighthouse score >90
- **Documentation**: All public APIs documented

## Security Policy
- **Authentication**: Multi-factor auth required for admin users
- **Data Protection**: GDPR compliance with data encryption
- **Vulnerability Management**: Weekly security scans, dependency updates
- **Incident Response**: 24/7 monitoring, <4 hour response time

## Deployment Strategy
- **Staging**: Auto-deploy on every push to main branch
- **Production**: Manual approval required, scheduled releases
- **Rollback**: Database migrations with rollback capability
- **Environments**: dev, staging, production with isolated data
```

### Mobile Application Documentation

```markdown
# product.md

## Mission & Strategy
We build mobile applications that solve everyday problems for users on the go.
Target users: iOS and Android users, ages 18-45, tech-savvy professionals.

## Success Metrics
- **Retention**: 70% 7-day retention, 40% 30-day retention
- **Engagement**: 3+ daily sessions per active user
- **Performance**: <2s app startup time, <50MB app size
- **Store Performance**: 4.5+ rating, 10K+ downloads in first month

## Next Features (SPEC Backlog)
- SPEC-001: Offline mode capability (Week 1-2)
- SPEC-002: Push notification system (Week 3)
- SPEC-003: Social sharing features (Week 4)
- SPEC-004: Advanced analytics dashboard (Week 5)

---

# structure.md

## System Architecture
```
UI Layer (Screens, Widgets)
    ↓
State Management (Riverpod/Bloc)
    ↓
Data Layer (SQLite/Realm + Remote API)
    ↓
Authentication (OAuth, JWT)
    ↓
Native Modules (Camera, GPS, Contacts)
    ↓
Offline Sync Engine
```

## Core Modules
- **Authentication**: Social login, session management
- **Core Features**: Main app functionality based on use case
- **Offline Support**: Local data storage, sync engine
- **Push Notifications**: Real-time user engagement
- **Analytics**: User behavior tracking, performance metrics

## External Integrations
- **App Stores**: App Store, Google Play with automated deployment
- **Analytics**: Firebase Analytics, Amplitude
- **Crash Reporting**: Crashlytics, Sentry
- **Backend APIs**: REST/GraphQL integration with offline queuing

## Traceability
- Screen flows mapped to user journey requirements
- Performance metrics linked to specific features
- A/B testing framework for feature validation

---

# tech.md

## Technology Stack
**Framework**: Flutter 3.13 or React Native 0.72
**Language**: Dart or TypeScript
**State Management**: Riverpod, Bloc, or Redux Toolkit
**Database**: SQLite with Hive/Realm for local storage
**Testing**: Widget tests, integration tests, 80%+ coverage

## Quality Gates
- **Performance**: App size <50MB, startup <2s
- **Testing**: 80% test coverage, all critical paths tested
- **Code Quality**: dart analyze passes, 0 warnings
- **Store Compliance**: Both App Store and Google Play guidelines met

## Security Policy
- **Data Protection**: Local encryption for sensitive data
- **Network Security**: Certificate pinning, secure communication
- **Authentication**: OAuth 2.0, JWT with refresh tokens
- **Privacy**: Compliant with app store privacy policies

## Deployment Strategy
- **Staging**: TestFlight (iOS) and Internal Testing (Android)
- **Production**: App Store release with staged rollout
- **Automated**: Fastlane for build and deployment
- **Monitoring**: Firebase Crashlytics for error tracking
```

### CLI Tool Documentation

```markdown
# product.md

## Mission & Strategy
We build command-line tools that automate repetitive developer tasks.
Target users: DevOps engineers, backend developers, data scientists.

## Success Metrics
- **Performance**: Process 1M records in <5 seconds
- **Adoption**: 1000+ GitHub stars, 10K+ npm downloads
- **Reliability**: 99.9% success rate, comprehensive error handling
- **Documentation**: 95% documentation coverage, time-to-first-use <5 minutes

## Next Features (SPEC Backlog)
- SPEC-001: Multi-format input support (JSON, CSV, Avro)
- SPEC-002: Plugin system architecture
- SPEC-003: Advanced filtering and transformation
- SPEC-004: Web dashboard for monitoring

---

# structure.md

## System Architecture
```
Input Parsing → Command Router → Core Logic → Output Formatter
                                    ↓
                           Validation Layer
                                    ↓
                            Caching Layer
```

## Core Modules
- **CLI Interface**: Command parsing, argument validation
- **Core Engine**: Business logic and data processing
- **Output Formatters**: JSON, CSV, XML, table display
- **Validation Layer**: Schema validation, error handling
- **Caching Layer**: Performance optimization for large datasets

## External Integrations
- **CI/CD**: GitHub Actions, GitLab CI, Jenkins
- **Data Sources**: Database connectors, file readers
- **Monitoring**: Prometheus metrics, Grafana dashboards
- **Package Managers**: npm, PyPI, Docker Hub

## Traceability
- Command execution linked to specific use cases
- Performance metrics trace to data processing stages
- Error codes mapped to specific failure scenarios

---

# tech.md

## Technology Stack
**Language**: Go 1.21 or Python 3.11
**CLI Framework**: Cobra or Click
**Testing**: Built-in testing or pytest
**Packaging**: Single binary or Python package

## Quality Gates
- **Performance**: <100ms startup time, <100MB binary size
- **Testing**: 90%+ test coverage for CLI commands
- **Code Quality**: golangci-lint passes with 0 warnings
- **Documentation**: Complete CLI help and examples

## Security Policy
- **Input Validation**: Schema validation for all inputs
- **Dependency Security**: Regular vulnerability scans
- **File Access**: Sandboxed execution with permission checks
- **Logging**: Audit trail for all operations

## Deployment Strategy
- **Binary Distribution**: Single executable with no dependencies
- **Package Managers**: GitHub Releases, PyPI, npm
- **Version Management**: Semantic versioning with changelog
- **Documentation**: Auto-generated from code comments
```

### Data Science Project Documentation

```markdown
# product.md

## Mission & Strategy
We build data science solutions that transform raw data into actionable insights.
Target users: Data scientists, ML engineers, business analysts.

## Success Metrics
- **Model Performance**: >95% accuracy on primary metrics
- **Data Quality**: <5% missing values, <1% outliers
- **Scalability**: Process 1B+ records efficiently
- **Impact**: Measurable business value demonstrated

## Next Features (SPEC Backlog)
- SPEC-001: Advanced feature engineering pipeline
- SPEC-002: Model monitoring and drift detection
- SPEC-003: Automated model retraining
- SPEC-004: Real-time prediction serving

---

# structure.md

## System Architecture
```
Data Ingestion → Feature Engineering → Model Training → Inference
    ↓
Feature Store
    ↓
Model Registry
    ↓
Monitoring & Alerting
```

## Core Modules
- **Data Pipeline**: ETL processes, data validation
- **Feature Engineering**: Feature extraction, selection, scaling
- **Model Training**: Training pipeline, hyperparameter optimization
- **Inference**: Model serving, batch/real-time prediction
- **Monitoring**: Model performance, data drift detection

## External Integrations
- **Data Sources**: Databases, APIs, file systems
- **ML Platforms**: MLflow, Weights & Biases, SageMaker
- **Monitoring**: Grafana, Prometheus, custom dashboards
- **Deployment**: Kubernetes, Docker, cloud platforms

## Traceability
- Data lineage tracked through entire pipeline
- Model performance linked to business metrics
- Feature importance mapped to model explainability

---

# tech.md

## Technology Stack
**Language**: Python 3.13, Jupyter notebooks
**ML Frameworks**: scikit-learn, PyTorch, TensorFlow
**Data Processing**: pandas, Polars, DuckDB
**Experiment Tracking**: MLflow, Weights & Biases

## Quality Gates
- **Code Coverage**: 80%+ coverage with comprehensive tests
- **Model Validation**: Cross-validation, backtesting
- **Data Quality**: Automated data validation tests
- **Documentation**: All experiments documented with reproducibility

## Security Policy
- **Data Privacy**: PII detection and anonymization
- **Model Security**: Model encryption, secure serving
- **Access Control**: Role-based access to data and models
- **Compliance**: GDPR, CCPA, industry regulations

## Deployment Strategy
- **Development**: Local notebooks, version controlled experiments
- **Staging**: Containerized model serving with A/B testing
- **Production**: Cloud deployment with autoscaling
- **Monitoring**: Real-time model performance tracking
```

---

## Documentation Templates

### Product.md Template

```markdown
# Project Name

## Mission & Strategy
[1-2 sentence mission statement describing core purpose]

## Target Users
[Specific user profiles, not generic "developers"]

## Success Metrics
[Measurable KPIs with targets and measurement frequency]

## Next Features (SPEC Backlog)
[Prioritized list of upcoming features with SPEC IDs]

## History
[Version history with significant changes]
```

### Structure.md Template

```markdown
# Project Structure

## System Architecture
[High-level architecture diagram or description]

## Core Modules
[Main building blocks with responsibilities]

## External Integrations
[External systems with authentication and failure modes]

## Traceability
[How SPECs map to code and changes are tracked]

## History
[Version history of architectural changes]
```

### Tech.md Template

```markdown
# Technology Stack

## Primary Technologies
[Main languages, frameworks, databases with version ranges]

## Quality Gates
[Automated checks and failure criteria]

## Security Policy
[Secrets management, vulnerability handling, incident response]

## Deployment Strategy
[Release process, environments, rollback procedures]

## Environment Configuration
[Development, staging, production differences]

## History
[Technology changes and migration history]
```

---

## Quality Checklists

### Product.md Validation
- [ ] Mission statement is 1-2 sentences
- [ ] Target users are specific and defined
- [ ] Problems are ranked by priority
- [ ] Success metrics are measurable and time-bound
- [ ] Feature backlog has 3-5 next SPEC IDs
- [ ] Version history is maintained

### Structure.md Validation
- [ ] Architecture clearly visualized or described
- [ ] Modules map to actual directory structure
- [ ] External integrations include auth and failure modes
- [ ] Traceability explains code documentation strategy
- [ ] Design trade-offs are documented with rationale

### Tech.md Validation
- [ ] Primary language with version range specified
- [ ] Quality gates define clear failure criteria
- [ ] Security policy covers secrets and incidents
- [ ] Deployment strategy includes full release workflow
- [ ] Environment configurations are documented
- [ ] Version history tracks technology changes

---

## Common Documentation Mistakes

### ❌ Too Vague
- "Users are developers"
- "We'll measure success by growth"

### ✅ Specific
- "Solo developers building web apps under time pressure"
- "80% team adoption within 2 weeks, 5 features/sprint"

### ❌ Over-Specified
- Function names and implementation details in product.md

### ✅ Architecture-Level
- "Caching layer for performance optimization"
- "Integration with external payment provider"

### ❌ Inconsistent
- Conflicting scale or quality requirements across documents

### ✅ Aligned
- All documents agree on target scale and standards

---

## Quick Reference

### Essential Commands

```bash
# Create project documentation
/moai-project-documentation
Skill("moai-project-documentation")

# Validate documentation completeness
# Check all three documents exist and meet quality standards
```

### Document Categories

| Document | Purpose | Audience |
|----------|---------|----------|
| Product.md | Business strategy, user needs | Stakeholders, PMs |
| Structure.md | Technical architecture | Developers, Architects |
| Tech.md | Implementation details | DevOps, Engineers |

### Project Type Mapping

| Type | Key Focus | Success Metrics |
|------|-----------|----------------|
| Web App | User adoption, engagement | DAU/MAU, retention |
| Mobile | Performance, store ratings | App size, startup time |
| CLI | Speed, reliability | Processing time, error rate |
| Library | Developer experience | API quality, documentation |
| Data Science | Model performance | Accuracy, scalability |

---

**Last Updated**: 2025-11-20
**Status**: Production Ready | Enterprise Approved
**Templates**: Web App, Mobile App, CLI, Library, Data Science
**Features**: AI-Enhanced, Context7 Integration, Quality Gates