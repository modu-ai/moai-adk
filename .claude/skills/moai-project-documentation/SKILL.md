---
name: moai-project-documentation
description: Enhanced project documentation with AI-powered features. Enhanced with Context7 (project)
version: 1.0.1
modularized: true
---

## 📊 Skill Metadata

**Name**: moai-project-documentation
**Domain**: Project Documentation & Planning
**Freedom Level**: high
**Target Users**: Project owners, architects, tech leads
**Invocation**: Skill("moai-project-documentation")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed guides)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Guide interactive creation of three core project documentation files:
- **product.md** - Mission, users, success metrics, feature backlog
- **structure.md** - Architecture, modules, external integrations, traceability
- **tech.md** - Technology stack, quality gates, security policy, deployment

**Key Pattern**: Document structure depends on **project type**:

1. **Web Application** (SaaS, REST API, web dashboard)
   - Focus: User personas, adoption metrics, real-time features
   - Tech: TypeScript/React, Python/FastAPI, PostgreSQL

2. **Mobile Application** (iOS/Android, Flutter, React Native)
   - Focus: User retention, app store metrics, offline capability
   - Tech: Flutter/React Native, SQLite, OAuth/JWT

3. **CLI Tool / Utility** (deployment tool, package manager)
   - Focus: Performance, integration, ecosystem adoption
   - Tech: Go or Python, single binary, <100ms startup

4. **Shared Library / SDK** (API client, validator, parser)
   - Focus: Developer experience, ecosystem adoption
   - Tech: TypeScript/Python, 90%+ test coverage

5. **Data Science / ML Project** (recommendation system, pipeline)
   - Focus: Model metrics, data quality, scalability
   - Tech: Python, scikit-learn/PyTorch, MLflow

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: product.md Structure

```markdown
# Mission & Strategy
- What problem do we solve?
- Who are the users?
- Value proposition

# Success Metrics
- KPIs (measurable targets)
- Measurement frequency
- Examples: "80% adoption within 2 weeks"

# Next Features (SPEC Backlog)
- 3-5 prioritized features
```

### Pattern 2: structure.md Structure

```markdown
# System Architecture
- Overall design pattern
- Layers/tiers and interactions
- Visual diagram or flow

# Core Modules
- Main building blocks
- Responsibilities
- Communication patterns

# External Integrations
- Dependencies with auth/failure modes
- Examples: Payment provider, message queue

# Traceability
- SPEC ID → code mapping
- Change tracking
```

### Pattern 3: tech.md Structure

```markdown
# Technology Stack
- Language(s) with version ranges
- Framework choices and rationale

# Quality Gates
- Required for merge (e.g., 85% test coverage)
- Enforcement tools

# Security Policy
- Secret management
- Vulnerability handling
- Incident response

# Deployment Strategy
- Target environments
- Release process
- Rollback procedure
```

### Pattern 4: Common Mistakes to Avoid

❌ **Too Vague**: "Users are developers", "Success is growth"
✅ **Specific**: "Solo developers, 3-7 person teams", "80% adoption in 2 weeks"

❌ **Over-Specified**: Function names, DB schemas, API endpoints
✅ **Architecture-Level**: "Caching layer", "External payment provider"

❌ **Inconsistent**: Different user scales across documents
✅ **Aligned**: All documents agree on target scale and quality standards

❌ **Outdated**: Last updated 6 months ago
✅ **Fresh**: HISTORY updated each sprint, version incremented

### Pattern 5: Writing Checklists

**product.md Checklist:**
- [ ] Mission: 1-2 sentences
- [ ] Users: Specific (not "developers")
- [ ] Problems: Ranked by priority
- [ ] Metrics: Measurable with targets
- [ ] Feature backlog: 3-5 next SPECs
- [ ] HISTORY section: version included

**structure.md Checklist:**
- [ ] Architecture: Visualized or clearly described
- [ ] Modules: Map to git directories
- [ ] Integrations: List auth, failure modes
- [ ] Traceability: TAG system explained
- [ ] Trade-offs: Why this design?
- [ ] HISTORY section: version included

**tech.md Checklist:**
- [ ] Language: Version range specified
- [ ] Quality gates: Failure criteria defined
- [ ] Security: Covers secrets, audits, incidents
- [ ] Deployment: Full release flow documented
- [ ] Environments: dev/test/prod profiles
- [ ] HISTORY section: version included

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed guides and examples, see:

- **[modules/quick-start.md](modules/quick-start.md)** - Project type selection guide with detailed examples
- **[modules/guides.md](modules/guides.md)** - Complete product.md, structure.md, tech.md writing guides per project type
- **[modules/checklists-examples.md](modules/checklists-examples.md)** - Full examples by project type (Web App, Mobile, CLI, Library, Data Science)
- **[modules/reference.md](modules/reference.md)** - Advanced patterns, troubleshooting, complete technical references

---

## 🔗 Integration with Other Skills

**Complementary Skills:**
- Skill("moai-project-config-manager") - Manage project configuration files
- Skill("moai-core-spec-authoring") - Create SPEC documents for features
- Skill("moai-docs-toolkit") - Generate API documentation

**Next Steps:**
After creating product.md, structure.md, tech.md:
1. Use Skill("moai-core-spec-authoring") to create SPEC-001 for first feature
2. Use Skill("moai-docs-toolkit") to generate API documentation
3. Use Skill("moai-project-config-manager") to update .moai/config/config.json

---

## 📈 Version History

**1.0.1** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure pattern
- 📚 Content moved to modules/ for better organization
- ✨ Core patterns highlighted in SKILL.md
- ✨ Added integration with other Skills

**1.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 5 project type templates
- ✨ Writing checklists and examples

---

**Maintained by**: alfred
**Domain**: Project Documentation & Planning
**Generated with**: MoAI-ADK Skill Factory
