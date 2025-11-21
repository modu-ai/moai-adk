---
name: moai-project-documentation-types
description: Project type selection and product documentation writing guide
---

## Project Type Selection & Product.md Guide

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

**End of Module** | moai-project-documentation-types
