---
name: moai-project-documentation-checklists
description: Writing checklists, common mistakes, and project type examples
---

## Writing Checklists & Best Practices

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

**Mistake 1: Too Vague**

❌ **Wrong:**
- "Users are developers"
- "We'll measure success by growth"

✅ **Correct:**
- "Solo developers building web apps under time pressure, 3-7 person teams"
- "Measure success by 80% adoption within 2 weeks, 5 features/sprint velocity"

---

**Mistake 2: Over-Specified in product.md**

❌ **Wrong:**
- Function names, database schemas, API endpoints
- "We'll use Redis cache with 1-hour TTL"

✅ **Correct:**
- "Caching layer for performance"
- "Integration with external payment provider"

---

**Mistake 3: Inconsistent Across Documents**

❌ **Wrong:**
- product.md: "5 concurrent users"
- structure.md: "Designed for 10,000 daily users"

✅ **Correct:**
- All 3 documents agree on target scale, user types, quality standards

---

**Mistake 4: Outdated**

❌ **Wrong:**
- Last updated 6 months ago
- HISTORY section has no recent entries

✅ **Correct:**
- HISTORY updated every sprint
- version number incremented on changes

---

### Examples by Project Type

**Example 1: Web App (TaskFlow)**

Product.md excerpt:
```markdown
Quality Gates
- Test coverage: 85% minimum
- Type errors: Zero in strict mode
- Bundle size: <200KB gzipped
```

---

**Example 2: Mobile App (FitTracker)**

Product.md excerpt:
```markdown
Deployment
- App Store & Google Play via Fastlane
- TestFlight for beta testing
- Version every 2 weeks
```

---

**Example 3: CLI Tool (DataValidate)**

Product.md excerpt:
```markdown
Build
- Binary size: <100MB
- Startup time: <100ms
- Distribution: GitHub Releases + Homebrew
```

---

**Example 4: Library (TypeGuard)**

Product.md excerpt:
```markdown
Primary Language: TypeScript 5.2+
- Test coverage: 90% (libraries have higher bar)
- Type checking: Zero errors in strict mode
- Bundle: <50KB gzipped, tree-shakeable
```

---

**Example 5: Data Science (ML Pipeline)**

Product.md excerpt:
```markdown
Model Registry
- Track all model versions with MLflow
- A/B test in production before full deployment
- Monitor data drift and model decay
```

---

### Versioning & Updates

**When to update this Skill:**
- New programming languages added to MoAI
- New project type examples needed
- Quality gate standards change
- Package management tools evolve

**Current version:** 4.0.0 (2025-11-12)

---

**End of Module** | moai-project-documentation-checklists
