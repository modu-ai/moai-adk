# Complete Examples & Checklists

Full documentation examples for each project type and detailed checklists.

## Writing Checklists

### product.md Checklist

- [ ] **Mission Statement**: 1-2 sentences, clear and specific
  - ❌ Bad: "Build great software"
  - ✅ Good: "Enable solo developers to launch web apps in <2 weeks"

- [ ] **Users**: Specific personas, NOT generic "developers"
  - ❌ Bad: "Developers"
  - ✅ Good: "Solo developers (1-3 years experience), 3-7 person startups"

- [ ] **Problems**: Ranked by priority
  - ✅ Priority 1: Most critical user problem
  - ✅ Priority 2: Secondary problem
  - ✅ Priority 3: Nice-to-have

- [ ] **Success Metrics**: Measurable with specific targets
  - ❌ Bad: "Become popular"
  - ✅ Good: "80% team adoption within 2 weeks, 5-feature velocity"

- [ ] **Feature Backlog**: 3-5 next SPEC items with priority
  - SPEC-001 (High): Feature name
  - SPEC-002 (Medium): Feature name
  - SPEC-003 (Low): Feature name

- [ ] **HISTORY Section**: Includes v0.1.0 with date
  - ✅ **v1.0.0** (2025-11-22) - Initial definition

- [ ] **Updated Recently**: Not older than last sprint
  - ❌ Last updated: 6 months ago
  - ✅ Last updated: 2 weeks ago

### structure.md Checklist

- [ ] **Architecture**: Visualized or clearly described
  - ✅ ASCII diagram or text description of how layers communicate

- [ ] **Modules Map to Git**: Each module maps to actual directory
  - ✅ Module: Core Logic
  - ✅ Files: src/logic/core/, src/logic/validation/

- [ ] **External Integrations**: List with auth and failure modes
  - ❌ Missing: "Uses PostgreSQL"
  - ✅ Complete: "PostgreSQL (connection pooling, circuit breaker, retry 3x)"

- [ ] **Traceability Explained**: How SPECs map to code
  - ✅ SPEC-001 → src/features/auth/login.py

- [ ] **Trade-offs Documented**: Why this design?
  - ✅ "We chose REST over GraphQL because..."
  - ✅ "Trade-off: Simpler API but fewer query optimizations"

- [ ] **HISTORY Section**: Includes v0.1.0

- [ ] **Data Flow**: Clear description or diagram
  - ✅ "User input → validation → API → database → response"

### tech.md Checklist

- [ ] **Primary Language**: Version range specified
  - ❌ Bad: "Python"
  - ✅ Good: "Python 3.13+"

- [ ] **Quality Gates**: Define failure criteria
  - ✅ "85% test coverage minimum"
  - ✅ "Zero type errors in strict mode"
  - ✅ "All linting errors must be fixed"

- [ ] **Security Policy**: Covers secrets, audits, incidents
  - ✅ "Secrets: GitHub Secrets only, never hardcoded"
  - ✅ "Vulnerabilities: Dependabot, monthly audits"
  - ✅ "Incident: Report to [email], response <24h"

- [ ] **Deployment**: Full release flow documented
  - ✅ Step 1: Tag release in Git
  - ✅ Step 2: GitHub Actions builds and tests
  - ✅ Step 3: Deploy to staging, run smoke tests
  - ✅ Step 4: Deploy to production
  - ✅ Step 5: Monitor error rates for 1 hour

- [ ] **Environment Profiles**: dev/test/prod described
  - ✅ Development: Local machine, SQLite
  - ✅ Staging: AWS EC2, PostgreSQL, GitHub main branch
  - ✅ Production: AWS ECS, PostgreSQL, Git tag

- [ ] **Rollback Procedure**: How to undo if broken
  - ✅ "Revert last commit, re-tag, re-deploy"

- [ ] **HISTORY Section**: Includes v0.1.0

---

## Example 1: Web Application (TaskFlow)

### Example product.md

```markdown
# Mission & Strategy

## Problem We Solve
TaskFlow helps busy professionals organize and prioritize their work without getting lost in complex software.

## Target Users
- **Busy professionals**: Software engineers, product managers, executives (2-10 years experience)
- **Team size**: 3-20 person teams
- **Environment**: Remote-first, asynchronous collaboration

## Value Proposition
- **Faster than Jira** (70% less clicking to create a task)
- **No training curve** (familiar to Trello users)
- **Built for async teams** (see updates without meetings)

---

# Success Metrics

## Key Performance Indicators
- **Adoption**: 80% of team members using TaskFlow within 2 weeks of launch
- **Engagement**: 70% daily active users (DAU) / monthly active users (MAU)
- **Retention**: <5% monthly churn rate
- **Velocity**: 5-7 features per sprint

## Measurement Frequency
- Daily: DAU/MAU, error rates
- Weekly: Adoption metrics, feature usage
- Monthly: Churn analysis, customer feedback

---

# Next Features (SPEC Backlog)

## High Priority
- **SPEC-001**: OAuth2 GitHub integration (estimated: 2 days)
- **SPEC-002**: Real-time collaboration (estimated: 5 days)

## Medium Priority
- **SPEC-003**: Mobile app (estimated: 3 weeks)

## Low Priority
- **SPEC-004**: API for third-party integrations (estimated: 2 weeks)

---

# HISTORY

**v1.0.0** (2025-11-22)
- 🎯 Initial product definition
- ✨ Target users: Busy professionals, 3-20 person teams
- ✨ Success metrics: 80% adoption, <5% churn
```

---

## Example 2: Mobile Application (FitTracker)

### Example product.md

```markdown
# Mission & Strategy

## Problem We Solve
FitTracker makes fitness tracking effortless by automatically logging activities from your phone.

## Target Users
- **Active people**: Ages 18-45, 3-5 workouts/week
- **Tech-savvy**: Comfortable with smartphone apps
- **Goal-oriented**: Want to track progress over months/years

---

# Success Metrics

## Key Performance Indicators
- **Downloads**: 100K in first 6 months
- **Rating**: 4.5+ stars (App Store + Google Play)
- **Retention**: 40% MAU (monthly active users)
- **DAU**: 50% of installed users
- **Churn**: <2% weekly churn after first month

---

# Next Features

## High Priority
- **SPEC-001**: Wearable integration (Apple Watch, Fitbit)
- **SPEC-002**: Social challenges (compete with friends)

---

# HISTORY

**v1.0.0** (2025-11-22)
- 📱 iOS and Android apps launched
- ✨ Retention: 40% MAU target
```

---

## Example 3: CLI Tool (DataValidate)

### Example product.md

```markdown
# Mission & Strategy

## Problem We Solve
DataValidate validates datasets in milliseconds, catching quality issues before they affect ML models.

## Target Users
- **Data engineers**: Building data pipelines
- **ML practitioners**: Preparing training data
- **Team size**: Individual contributors to teams of 50+

---

# Success Metrics

## Key Performance Indicators
- **Performance**: <5s for 1M records validation
- **Adoption**: 500+ GitHub stars
- **Distribution**: 10K npm downloads/month
- **Ecosystem**: Native support for JSON, CSV, Avro, Parquet

---

# Next Features

- **SPEC-001**: Kafka integration
- **SPEC-002**: Schema evolution support

---

# HISTORY

**v1.0.0** (2025-11-22)
- 🚀 CLI tool released
- ✨ Performance: <5s for 1M records
```

---

## Example 4: Library (TypeGuard)

### Example product.md

```markdown
# Mission & Strategy

## Problem We Solve
TypeGuard provides runtime type validation for TypeScript with zero overhead.

## Target Users
- **TypeScript developers**: Building type-safe applications
- **Library authors**: Need runtime validation
- **Scale**: From solo projects to enterprise systems

---

# Success Metrics

## Key Performance Indicators
- **Adoption**: 50K npm downloads/month
- **Quality**: 90%+ test coverage (libraries = higher bar)
- **Performance**: <1ms validation per object
- **Community**: <24h issue response time

---

# Next Features

- **SPEC-001**: Async validation support
- **SPEC-002**: Custom validator plugins

---

# HISTORY

**v1.0.0** (2025-11-22)
- 📦 NPM package released
- ✨ Runtime type validation with zero overhead
```

---

## Example 5: Data Science Project (ML Pipeline)

### Example product.md

```markdown
# Mission & Strategy

## Problem We Solve
Our recommendation engine increases platform engagement by 30% through personalized content suggestions.

## Target Users
- **Content platforms**: E-commerce, streaming, social media
- **Scale**: 1M+ users, 1B+ recommendations/day

---

# Success Metrics

## Key Performance Indicators
- **Model Accuracy**: 92% precision@10
- **Latency**: <100ms per prediction (p99)
- **Impact**: 25-30% engagement increase
- **Reliability**: 99.9% uptime

---

# Next Features

- **SPEC-001**: Real-time feature updates
- **SPEC-002**: Multi-armed bandit exploration

---

# HISTORY

**v1.0.0** (2025-11-22)
- 🤖 ML pipeline deployed
- ✨ 92% precision, 25% engagement lift
```

---

## Mistake Examples & Corrections

### Mistake 1: Too Vague

❌ **WRONG**:
```
Users: Developers
Success Metric: Become popular
Problem: Make development easier
```

✅ **CORRECT**:
```
Users: Solo developers (2-5 years experience), 3-7 person startups
Success Metric: 80% adoption within 2 weeks, 5-7 feature velocity
Problem: Complex frameworks overwhelm developers; we simplify with conventions
```

---

### Mistake 2: Inconsistent Across Documents

❌ **WRONG**:
```
product.md: "Target 5 concurrent users"
structure.md: "Designed for 10,000 concurrent users"
tech.md: "Database supports 100,000 concurrent users"
```

✅ **CORRECT**:
```
All documents agree: "Target 10,000 concurrent users"
- product.md: "Success metric: support 10K concurrent users"
- structure.md: "Designed for 10K concurrent users with auto-scaling"
- tech.md: "PostgreSQL connection pool sized for 10K concurrent users"
```

---

### Mistake 3: Over-Specified in product.md

❌ **WRONG**:
```
product.md should include:
- Database schema (users table, posts table, likes table)
- API endpoint details (GET /api/v1/users/:id)
- Redis cache with 1-hour TTL
```

✅ **CORRECT**:
```
product.md should be architecture-level:
- "Caching layer for performance"
- "External database for persistence"
- "API for client-server communication"
```

---

### Mistake 4: Outdated Documentation

❌ **WRONG**:
```
product.md
Last modified: 2024-05-15 (6 months ago)
No HISTORY section
Version: ???
```

✅ **CORRECT**:
```
product.md
Last modified: 2025-11-22 (this week)
HISTORY section:
**v1.0.2** (2025-11-22)
- Updated success metrics based on Q4 performance
**v1.0.1** (2025-11-15)
- Added mobile roadmap
**v1.0.0** (2025-11-01)
- Initial definition
```

---

## Next Steps

1. ✅ Choose your project type (from quick-start.md)
2. Use guides.md to write product.md, structure.md, tech.md
3. ✅ You are here - compare your drafts against these examples
4. Use reference.md for advanced patterns and troubleshooting
