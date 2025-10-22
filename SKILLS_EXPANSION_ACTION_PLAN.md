# Skills v3.0 Expansion Action Plan

**Goal**: Expand all 56 MoAI-ADK Skills from current average 163 LOC to target 1,200+ LOC per Skill
**Timeline**: 2-3 months (phased approach)
**Status**: 📋 Planning → 🚀 Ready to Execute

---

## 📊 Current State Assessment

### Size Distribution
- **< 200 LOC**: 50 Skills (89%)
- **200-500 LOC**: 4 Skills (7%)
- **500+ LOC**: 2 Skills (4%)

### Already Comprehensive (500+ LOC)
- ✅ `moai-essentials-debug`: 698 LOC
- ✅ `moai-alfred-tui-survey`: 635 LOC
- ✅ `moai-skill-factory`: 560 LOC
- ✅ `moai-lang-python`: 431 LOC
- ✅ `moai-foundation-trust`: 307 LOC
- ✅ `moai-domain-backend`: 290 LOC

### Need Expansion (< 300 LOC)
- 🎯 50 Skills need expansion to 1,200+ LOC

---

## 🎯 Expansion Strategy

### Standard Skill Template (1,200+ LOC)

```markdown
# moai-{tier}-{name}

**Model**: haiku/sonnet
**Tier**: Foundation/Essentials/Alfred/Domain/Language/Ops
**Purpose**: [Clear one-line purpose]

---

## ▶◀ Core Concept (100-200 LOC)
- Fundamental principles
- Key definitions
- When to use / When not to use

## 📚 Comprehensive Guide (300-500 LOC)
- Detailed methodology
- Step-by-step workflows
- Decision trees

## 💡 Practical Examples (300-500 LOC)
- 10+ Real-world examples
- Complete code snippets
- Before/After comparisons
- Use case scenarios

## 🚫 Anti-patterns (100-200 LOC)
- Common mistakes
- What NOT to do
- Red flags
- Recovery strategies

## 🔧 Troubleshooting (100-200 LOC)
- Common errors
- Diagnostic checklist
- Resolution steps
- Prevention tips

## 📖 Reference Materials (100-200 LOC)
- Official documentation links
- Best practice resources
- Community contributions
- Related Skills

## ✅ Checklists (100-200 LOC)
- Pre-flight checks
- Quality gates
- Verification steps
- Success criteria
```

---

## 📅 Phase 1: Foundation & Essentials (Week 1-2)

**Priority**: 🔥 **HIGHEST** — Core Skills that all agents depend on

### Foundation Tier (6 Skills)

#### 1. moai-foundation-trust (307 → 1,200+ LOC)
**Current**: Basic TRUST 5 principles
**Expansion**:
- [ ] Add detailed checklists for each principle (T.R.U.S.T)
- [ ] Add 20+ real-world examples (good vs bad code)
- [ ] Add testing framework comparison (Jest/Vitest/pytest/JUnit)
- [ ] Add coverage tool integration guides
- [ ] Add security vulnerability examples (OWASP Top 10)
- [ ] Add TAG traceability best practices
- [ ] Add TRUST violation recovery strategies

**Estimated Size**: 1,200+ LOC

#### 2. moai-foundation-tags (113 → 1,200+ LOC)
**Current**: Basic TAG naming rules
**Expansion**:
- [ ] Add TAG chain visualization techniques
- [ ] Add orphan TAG detection strategies
- [ ] Add TAG migration guide (old → new format)
- [ ] Add TAG versioning best practices
- [ ] Add 15+ TAG chain examples (SPEC → TEST → CODE → DOC)
- [ ] Add TAG conflict resolution procedures
- [ ] Add TAG integrity verification scripts

**Estimated Size**: 1,200+ LOC

#### 3. moai-foundation-specs (113 → 1,200+ LOC)
**Current**: Basic SPEC metadata structure
**Expansion**:
- [ ] Add SPEC lifecycle management (draft → active → completed → archived)
- [ ] Add SPEC versioning strategies (semantic versioning)
- [ ] Add SPEC migration guides (v1 → v2)
- [ ] Add 20+ SPEC template examples
- [ ] Add SPEC review checklist
- [ ] Add SPEC approval workflow
- [ ] Add SPEC traceability matrix examples

**Estimated Size**: 1,200+ LOC

#### 4. moai-foundation-ears (113 → 1,200+ LOC)
**Current**: Basic EARS 5 patterns
**Expansion**:
- [ ] Add 30+ EARS requirement examples
- [ ] Add EARS anti-patterns (what to avoid)
- [ ] Add EARS template library (10+ domains)
- [ ] Add EARS validation checklist
- [ ] Add EARS to user story conversion
- [ ] Add EARS to test case conversion
- [ ] Add EARS refactoring strategies

**Estimated Size**: 1,200+ LOC

#### 5. moai-foundation-git (122 → 1,200+ LOC)
**Current**: Basic GitFlow automation
**Expansion**:
- [ ] Add Git branching strategies (GitFlow, Trunk-based, GitHub Flow)
- [ ] Add commit message best practices (Conventional Commits)
- [ ] Add PR template library
- [ ] Add Git hook examples (pre-commit, pre-push)
- [ ] Add merge conflict resolution strategies
- [ ] Add Git rebase vs merge comparison
- [ ] Add Git submodule management

**Estimated Size**: 1,200+ LOC

#### 6. moai-foundation-langs (113 → 1,200+ LOC)
**Current**: Basic language detection
**Expansion**:
- [ ] Add multi-language project strategies
- [ ] Add language interoperability patterns
- [ ] Add language-specific toolchain comparison
- [ ] Add language migration guides (e.g., JS → TS)
- [ ] Add language benchmark comparison
- [ ] Add language ecosystem overview (23 languages)
- [ ] Add language selection decision tree

**Estimated Size**: 1,200+ LOC

### Essentials Tier (4 Skills)

#### 7. moai-essentials-debug (698 → 1,200+ LOC)
**Current**: Already comprehensive debugging guide
**Expansion**:
- [ ] Add debugger tool comparison (Chrome DevTools, VSCode, PyCharm)
- [ ] Add remote debugging strategies
- [ ] Add production debugging best practices
- [ ] Add 10+ real debugging case studies
- [ ] Add memory leak detection techniques
- [ ] Add performance profiling integration
- [ ] Add distributed tracing examples

**Estimated Size**: 1,200+ LOC

#### 8. moai-essentials-perf (113 → 1,200+ LOC)
**Current**: Basic performance concepts
**Expansion**:
- [ ] Add profiling tool comparison (Lighthouse, WebPageTest, py-spy)
- [ ] Add performance budget strategies
- [ ] Add 20+ optimization case studies
- [ ] Add database query optimization
- [ ] Add caching strategies (Redis, CDN, etc.)
- [ ] Add load testing guides (k6, Locust, JMeter)
- [ ] Add performance monitoring setup (APM, metrics)

**Estimated Size**: 1,200+ LOC

#### 9. moai-essentials-refactor (113 → 1,200+ LOC)
**Current**: Basic refactoring concepts
**Expansion**:
- [ ] Add 25+ refactoring patterns (Extract Method, Inline Variable, etc.)
- [ ] Add code smell detection guide (Long Method, Feature Envy, etc.)
- [ ] Add refactoring tool comparison (IDE features, static analysis)
- [ ] Add 15+ before/after refactoring examples
- [ ] Add refactoring safety checklist
- [ ] Add technical debt measurement
- [ ] Add refactoring cost-benefit analysis

**Estimated Size**: 1,200+ LOC

#### 10. moai-essentials-review (113 → 1,200+ LOC)
**Current**: Basic code review concepts
**Expansion**:
- [ ] Add code review checklist (security, performance, readability)
- [ ] Add PR review best practices
- [ ] Add 10+ review comment examples (good vs bad)
- [ ] Add review automation tools (SonarQube, CodeClimate)
- [ ] Add review metrics (time, quality, feedback)
- [ ] Add review culture building guide
- [ ] Add review template library

**Estimated Size**: 1,200+ LOC

---

## 📅 Phase 2: Alfred Tier (Week 3-4)

**Priority**: 🔥 **HIGH** — Alfred-specific workflow orchestration

### Alfred Tier (11 Skills)

#### 11-21. Alfred Skills (11 × 1,200+ LOC)

**Standard Expansion per Alfred Skill**:
- [ ] Add workflow diagrams (ASCII art or Mermaid)
- [ ] Add 10+ integration examples with other Skills
- [ ] Add error handling strategies
- [ ] Add performance optimization tips
- [ ] Add troubleshooting guide
- [ ] Add anti-patterns
- [ ] Add checklists

**Skills to Expand**:
1. `moai-alfred-code-reviewer` (113 → 1,200+ LOC)
2. `moai-alfred-debugger-pro` (113 → 1,200+ LOC)
3. `moai-alfred-ears-authoring` (113 → 1,200+ LOC)
4. `moai-alfred-git-workflow` (122 → 1,200+ LOC)
5. `moai-alfred-language-detection` (113 → 1,200+ LOC)
6. `moai-alfred-performance-optimizer` (113 → 1,200+ LOC)
7. `moai-alfred-refactoring-coach` (113 → 1,200+ LOC)
8. `moai-alfred-spec-metadata-validation` (113 → 1,200+ LOC)
9. `moai-alfred-tag-scanning` (113 → 1,200+ LOC)
10. `moai-alfred-trust-validation` (113 → 1,200+ LOC)
11. `moai-alfred-tui-survey` (635 → 1,200+ LOC) ← Already close

---

## 📅 Phase 3: Domain Tier (Week 5-6)

**Priority**: 🟡 **MEDIUM** — Domain-specific expertise

### Domain Tier (10 Skills)

**Standard Domain Expansion Template**:
- [ ] Add architecture patterns (3-5 patterns per domain)
- [ ] Add real-world case studies (5+ projects)
- [ ] Add technology stack recommendations
- [ ] Add best practices library
- [ ] Add anti-patterns
- [ ] Add migration guides
- [ ] Add troubleshooting guides

#### 22. moai-domain-backend (290 → 1,200+ LOC)
**Focus**: Microservices, Event-driven, DDD, CQRS, Saga patterns

#### 23. moai-domain-frontend (124 → 1,200+ LOC)
**Focus**: React/Vue/Angular architecture, State management, SSR/SSG

#### 24. moai-domain-web-api (123 → 1,200+ LOC)
**Focus**: REST, GraphQL, gRPC, API versioning, Rate limiting

#### 25. moai-domain-mobile-app (123 → 1,200+ LOC)
**Focus**: iOS/Android/Flutter, Offline-first, Push notifications

#### 26. moai-domain-security (123 → 1,200+ LOC)
**Focus**: OWASP Top 10, Authentication/Authorization, Secrets management

#### 27. moai-domain-devops (124 → 1,200+ LOC)
**Focus**: CI/CD, Infrastructure as Code, Container orchestration

#### 28. moai-domain-database (123 → 1,200+ LOC)
**Focus**: RDBMS, NoSQL, Schema design, Query optimization

#### 29. moai-domain-data-science (123 → 1,200+ LOC)
**Focus**: Data pipelines, ML workflows, Feature engineering

#### 30. moai-domain-ml (123 → 1,200+ LOC)
**Focus**: Model training, Deployment, Monitoring, MLOps

#### 31. moai-domain-cli-tool (123 → 1,200+ LOC)
**Focus**: Argument parsing, Configuration, Distribution

---

## 📅 Phase 4: Language Tier (Week 7-10)

**Priority**: 🟢 **LOWER** — Language-specific best practices

### Language Tier (23 Skills)

**Standard Language Expansion Template**:
- [ ] Add language-specific testing strategies
- [ ] Add language-specific linting/formatting
- [ ] Add language-specific performance tips
- [ ] Add language-specific security best practices
- [ ] Add language-specific toolchain guide
- [ ] Add 10+ code examples
- [ ] Add anti-patterns

**Target Sizes**:
- **Major Languages** (Python, TypeScript, JavaScript, Java, Go, Rust): 1,200+ LOC
- **Other Languages**: 300+ LOC

#### High Priority Languages (1,200+ LOC)

32. `moai-lang-python` (431 → 1,200+ LOC)
33. `moai-lang-typescript` (127 → 1,200+ LOC)
34. `moai-lang-javascript` (125 → 1,200+ LOC)
35. `moai-lang-java` (124 → 1,200+ LOC)
36. `moai-lang-go` (124 → 1,200+ LOC)
37. `moai-lang-rust` (124 → 1,200+ LOC)

#### Standard Languages (300+ LOC)

38-54. Remaining 17 languages (120-125 → 300+ LOC each)

---

## 📅 Phase 5: Ops & Meta (Week 11-12)

**Priority**: 🟢 **LOWER** — Claude Code operations and meta-Skills

#### 55. moai-claude-code (121 → 1,200+ LOC)
**Focus**: Claude Code session management, output formatting, Skill deployment

#### 56. moai-skill-factory (560 → 1,200+ LOC)
**Focus**: Skill creation orchestration, quality validation, template application

---

## 🛠️ Expansion Workflow

### Step-by-Step Process per Skill

1. **Audit Current Content** (15 min)
   - Read existing SKILL.md, examples.md, reference.md
   - Identify gaps and expansion opportunities

2. **Research & Gather Materials** (30 min)
   - Review official documentation
   - Collect best practices from community
   - Identify relevant code examples

3. **Draft Expansion** (2-3 hours)
   - Write Core Concept section (100-200 LOC)
   - Write Comprehensive Guide (300-500 LOC)
   - Write Practical Examples (300-500 LOC)
   - Write Anti-patterns (100-200 LOC)
   - Write Troubleshooting (100-200 LOC)
   - Write Reference Materials (100-200 LOC)
   - Write Checklists (100-200 LOC)

4. **Review & Validate** (30 min)
   - Check for completeness (all sections present)
   - Verify examples are runnable
   - Ensure consistency with MoAI-ADK standards
   - Validate line count (≥ 1,200 LOC)

5. **Integration Test** (15 min)
   - Test Skill activation in relevant agent
   - Verify Progressive Disclosure compliance
   - Check JIT loading behavior

6. **Document & Commit** (15 min)
   - Update SKILLS_EXPANSION_TRACKER.md
   - Commit with message: `feat(skills): expand [skill-name] to 1,200+ LOC`

**Total Time per Skill**: ~4-5 hours

---

## 📊 Progress Tracking

### Completion Metrics

| Phase | Skills | Target LOC | Current LOC | Completion % |
|---|---|---|---|---|
| Phase 1 (Foundation/Essentials) | 10 | 12,000+ | 2,930 | 24% |
| Phase 2 (Alfred) | 11 | 13,200+ | 1,925 | 15% |
| Phase 3 (Domain) | 10 | 12,000+ | 1,473 | 12% |
| Phase 4 (Language) | 23 | 13,800+ | 3,129 | 23% |
| Phase 5 (Ops/Meta) | 2 | 2,400+ | 681 | 28% |
| **Total** | **56** | **53,400+** | **10,138** | **19%** |

### Weekly Milestones

- **Week 1**: Complete Foundation Tier (6 Skills)
- **Week 2**: Complete Essentials Tier (4 Skills)
- **Week 3**: Complete Alfred Tier Part 1 (6 Skills)
- **Week 4**: Complete Alfred Tier Part 2 (5 Skills)
- **Week 5**: Complete Domain Tier Part 1 (5 Skills)
- **Week 6**: Complete Domain Tier Part 2 (5 Skills)
- **Week 7-8**: Complete Major Languages (6 Skills)
- **Week 9-10**: Complete Remaining Languages (17 Skills)
- **Week 11-12**: Complete Ops & Meta (2 Skills)

---

## 🎯 Success Criteria

### Per-Skill Criteria
- ✅ Total content ≥ 1,200 LOC (or 300+ LOC for secondary languages)
- ✅ All 7 sections present (Core, Guide, Examples, Anti-patterns, Troubleshooting, Reference, Checklists)
- ✅ At least 10 practical examples
- ✅ At least 5 anti-patterns documented
- ✅ At least 3 troubleshooting scenarios
- ✅ Links to 5+ external resources
- ✅ Integration tested with relevant agent

### Project-Level Criteria
- ✅ All 56 Skills meet size targets
- ✅ Average Skill size ≥ 950 LOC (53,400 / 56)
- ✅ Zero Skill < 100 LOC
- ✅ Progressive Disclosure compliance maintained
- ✅ JIT loading performance not degraded

---

## 🚀 Getting Started

### Immediate Next Steps

1. **Set Up Tracking** (Today)
   - [ ] Create `SKILLS_EXPANSION_TRACKER.md`
   - [ ] Set up GitHub Project board (optional)
   - [ ] Create expansion branch: `feat/skills-v3-expansion`

2. **Pilot Expansion** (Week 1, Day 1-2)
   - [ ] Select 1 Foundation Skill as pilot (e.g., `moai-foundation-trust`)
   - [ ] Follow expansion workflow
   - [ ] Measure actual time spent
   - [ ] Refine process based on learnings

3. **Begin Phase 1** (Week 1, Day 3-5)
   - [ ] Expand remaining 5 Foundation Skills
   - [ ] Expand 4 Essentials Skills

4. **Weekly Review** (Every Friday)
   - [ ] Update progress tracker
   - [ ] Review quality of expanded Skills
   - [ ] Adjust timeline if needed

---

## 📝 Templates

### Expansion Commit Message Template

```
feat(skills): expand [skill-name] to 1,200+ LOC

- Add comprehensive guide (300+ LOC)
- Add 10+ practical examples (300+ LOC)
- Add anti-patterns section (100+ LOC)
- Add troubleshooting guide (100+ LOC)
- Add reference materials (100+ LOC)
- Add checklists (100+ LOC)

Total: [actual LOC] lines
Phase: [Phase number]
Skill Tier: [Foundation/Essentials/Alfred/Domain/Language/Ops/Meta]
```

### Section Template (for copy-paste)

```markdown
## 💡 Practical Examples

### Example 1: [Scenario Name]

**Context**: [When to use this pattern]

**Before** (❌ Anti-pattern):
```[language]
[bad code example]
```

**After** (✅ Best practice):
```[language]
[good code example]
```

**Key Takeaways**:
- [Lesson 1]
- [Lesson 2]
- [Lesson 3]

**Related Skills**: [Link to related Skills]

---

### Example 2: [Next Scenario]
...
```

---

## 🏆 Expected Outcomes

### By End of Phase 1 (Week 2)
- ✅ 10 Skills expanded (Foundation + Essentials)
- ✅ Total LOC: 10,138 → 22,138 (+12,000)
- ✅ Completion: 19% → 41%

### By End of Phase 2 (Week 4)
- ✅ 21 Skills expanded
- ✅ Total LOC: 22,138 → 35,338 (+13,200)
- ✅ Completion: 41% → 66%

### By End of Phase 3 (Week 6)
- ✅ 31 Skills expanded
- ✅ Total LOC: 35,338 → 47,338 (+12,000)
- ✅ Completion: 66% → 89%

### By End of Phase 5 (Week 12)
- ✅ 56 Skills expanded (**100%**)
- ✅ Total LOC: 10,138 → 53,400+ (+43,262)
- ✅ Completion: 19% → **100%**

---

**Action Plan Status**: ✅ **APPROVED & READY**
**Start Date**: [To be determined]
**Expected Completion**: 12 weeks from start
**Next Step**: Create `SKILLS_EXPANSION_TRACKER.md` and begin pilot expansion

---

**Document Owner**: cc-manager (MoAI-ADK Control Tower)
**Last Updated**: 2025-10-22
**Version**: 1.0
