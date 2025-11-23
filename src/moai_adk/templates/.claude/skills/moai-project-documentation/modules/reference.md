# Advanced Patterns & Troubleshooting

Advanced documentation patterns, troubleshooting guides, and complete technical references.

## Advanced Pattern 1: Tiered Documentation Strategy

For large projects, consider separating documentation by audience:

### Tier 1: Executive Summary (product.md)
- **Audience**: Executives, product managers, non-technical stakeholders
- **Content**: Vision, success metrics, business impact
- **Length**: 1-2 pages

### Tier 2: Architecture (structure.md)
- **Audience**: Engineers, architects, technical leads
- **Content**: System design, module responsibilities, trade-offs
- **Length**: 3-5 pages

### Tier 3: Implementation (tech.md)
- **Audience**: Developers implementing features
- **Content**: Tech stack, quality gates, deployment procedures
- **Length**: 2-3 pages

### Tier 4: Detailed Guides (optional)
- **Audience**: New developers, contributors
- **Content**: Setup guide, coding standards, contribution guidelines
- **Length**: As needed

---

## Advanced Pattern 2: Documentation-Driven Development

Use documentation to drive architecture decisions:

### Step 1: Write product.md first
Define what you're building and for whom.

### Step 2: Write structure.md second
Design how you'll build it based on product requirements.

### Step 3: Write tech.md third
Choose technology that supports the architecture.

### Step 4: Implement code
Write code that matches the documented architecture.

This prevents "architecture drift" where code diverges from intended design.

---

## Advanced Pattern 3: Living Documentation

Keep documentation synchronized with code:

### Policy
- **Update cadence**: After every major feature or architecture change
- **Review process**: Code review includes documentation review
- **Ownership**: Architecture owner responsible for keeping docs fresh

### Automation
```bash
# Example: Validate documentation references actual code
grep -r "src/features/" structure.md | while read path; do
    if [ ! -d "$path" ]; then
        echo "ERROR: $path referenced in docs but doesn't exist"
    fi
done
```

### Version Bumping Rules
- **Major version** (v1.0.0 → v2.0.0): Architecture redesign
- **Minor version** (v1.0.0 → v1.1.0): New module or significant feature
- **Patch version** (v1.0.0 → v1.0.1): Documentation fix, small refinement

---

## Advanced Pattern 4: Multi-Project Documentation

If you have multiple related projects:

### Shared Elements
Create a shared documentation repository for:
- Team processes
- Shared infrastructure
- Organization standards
- Deployment procedures

### Project-Specific Elements
Each project maintains its own:
- product.md (specific to that project)
- structure.md (specific architecture)
- tech.md (specific tech stack)

### Linking
Reference shared docs:
```markdown
# Deployment Strategy

For our standard deployment pipeline, see [Organization Deployment Guide](../shared/deployment.md)

## Project-Specific Customizations

We override the standard pipeline with:
- Custom health check endpoint at /health/custom
- Additional smoke tests for financial transactions
```

---

## Troubleshooting Guide

### Problem 1: "Product.md is too vague"

**Symptoms**: Metrics are subjective, users are generic

**Solution**:
- ✅ Replace "developers" with specific personas
- ✅ Replace "popular" with measurable target (50K GitHub stars)
- ✅ Add timeline: "80% adoption within 2 weeks"

### Problem 2: "Structure.md doesn't match reality"

**Symptoms**: Code is organized differently than documented

**Solutions**:
1. Update docs to match code (if code is better)
2. Refactor code to match docs (if docs are better)
3. Split docs into "current" and "proposed" if mid-refactor

### Problem 3: "Tech stack is outdated"

**Symptoms**: HISTORY section shows last update 6 months ago

**Solution**:
- [ ] Schedule quarterly documentation review
- [ ] Update HISTORY with any changes
- [ ] Bump version number
- [ ] Compare recommended versions with team's actual versions

### Problem 4: "Modules in structure.md don't match git directories"

**Symptoms**: Documentation lists Module A but git has module-a/, ModuleA/, etc.

**Solution**:
- Use exact git paths in documentation
- Example: `src/features/auth/` not `Auth Module`
- Cross-reference: "Module: Authentication, Files: src/features/auth/"

### Problem 5: "Documentation is overwhelming (too many pages)"

**Symptoms**: Single SKILL.md >1000 lines, users don't know where to start

**Solution**:
- Create progressive disclosure with modules/ (this Skill pattern)
- Start with Quick Reference (5-minute read)
- Link to detailed guides only when needed
- Use table of contents with section references

### Problem 6: "Tech.md quality gates are impossible to meet"

**Symptoms**: 95% test coverage impossible to maintain, team always failing CI

**Solutions**:
1. Reduce coverage target to achievable level (85% is standard)
2. Allow exceptions with documentation (e.g., "Generated code excluded")
3. Use coverage thresholds per file type (integration tests may need <85%)

---

## Reference Material

### Common Quality Gate Standards

**Test Coverage by Project Type**:
- Web Application: 80% (focus on API logic, UI harder to test)
- Library: 90%+ (libraries have higher bar)
- Data Science: 75-80% (notebook-based, harder to test)
- CLI Tool: 85% (include integration tests)

**Performance Targets by Type**:
- Web App: <200KB bundle (gzipped), <2s page load
- Mobile: <50MB app size, <2s startup
- CLI Tool: <100ms startup
- Library: Zero overhead when possible
- Data Science: <100ms per prediction (p99)

**Security Standards**:
- All projects: Zero hardcoded secrets
- All projects: Dependency scanning (Dependabot)
- Web/API projects: OWASP Top 10 validation
- Data projects: PII handling documentation

---

## Documentation Template Checklist

Use this checklist when starting new documentation:

### Pre-Writing Checklist
- [ ] Identified project type (Web, Mobile, CLI, Library, Data Science)
- [ ] Identified target audience (who reads this?)
- [ ] Identified success metrics (how will we measure?)
- [ ] Identified constraints (time, budget, scale)

### Writing Checklist (All Three Docs)
- [ ] Specific, not vague (names, numbers, dates)
- [ ] Consistent across all three documents (scales match, users match)
- [ ] Up-to-date (HISTORY section recent)
- [ ] Actionable (readers know what to do)
- [ ] Complete (all major decisions documented)

### Post-Writing Checklist
- [ ] Reviewed by architect/tech lead
- [ ] Reviewed by team members (readability)
- [ ] Updated git links (match actual repo structure)
- [ ] Added HISTORY section
- [ ] Set version number (v0.1.0 minimum)
- [ ] Added to team wiki/documentation site

---

## Related Documentation

### External Resources
- [Divio Documentation Framework](https://documentation.divio.com/) - Tutorial, How-to, Explanation, Reference framework
- [12Factor App](https://12factor.net/) - Configuration and environment best practices
- [Architecture Decision Records (ADR)](https://adr.github.io/) - Document important decisions

### MoAI-ADK Integration
- Use Skill("moai-core-spec-authoring") after defining product.md to create SPEC-001
- Use Skill("moai-project-config-manager") to manage .moai/config/config.json
- Use Skill("moai-docs-toolkit") to generate API documentation from code

### Example Projects
- [React Documentation](https://react.dev/) - Well-organized, multi-tier documentation
- [Python Requests](https://docs.python-requests.org/) - Clear, practical guide
- [Kubernetes Docs](https://kubernetes.io/docs/) - Comprehensive, task-based organization

---

## Versioning Strategy for Documentation

### When to Bump Versions

**Patch (v1.0.0 → v1.0.1)**:
- [ ] Typo or grammar fix
- [ ] Clarification without changing meaning
- [ ] Updated timestamp or reference link
- [ ] Example code improvement

**Minor (v1.0.0 → v1.1.0)**:
- [ ] New module or feature documented
- [ ] New tech stack component added
- [ ] New quality gate standard
- [ ] Architecture refinement (same overall pattern)

**Major (v1.0.0 → v2.0.0)**:
- [ ] Architecture redesign (different pattern)
- [ ] Technology complete rewrite
- [ ] Target audience change (e.g., from startup to enterprise)
- [ ] Business model change

### HISTORY Section Example

```markdown
# HISTORY

## v1.2.1 (2025-11-22)
- 🐛 Fixed broken link to deployment guide
- ✨ Clarified API response format in structure.md

## v1.2.0 (2025-11-15)
- 🎉 Added mobile architecture documentation
- ✨ Documented new WebSocket module

## v1.1.0 (2025-11-08)
- 📚 Expanded tech.md with pre-commit hook configuration
- ✨ Added monitoring section to structure.md

## v1.0.0 (2025-11-01)
- 🎯 Initial documentation set
- ✨ Defined product.md, structure.md, tech.md
```

---

## FAQ

**Q: How often should we update documentation?**

A: At minimum, after every major feature or architecture change. Ideally, every sprint. Set a calendar reminder for quarterly review.

---

**Q: Can documentation be too detailed?**

A: Yes. Use progressive disclosure:
- SKILL.md: Core concepts (this Skill)
- modules/: Detailed guides
- Separate guide: Implementation details

Readers choose their depth level.

---

**Q: What if documentation and code disagree?**

A: Fix the mismatch immediately:
1. Is code correct? → Update docs
2. Is docs correct? → Update code
3. Both need revision? → Discuss with team

Never let them diverge.

---

**Q: How do we keep documentation synchronized?**

A: Automation:
```bash
# Pre-commit hook to check docs updated
git diff HEAD~1 --name-only | grep -E '\.py|\.ts|\.go' && \
  git diff HEAD~1 --name-only | grep -E 'product.md|structure.md|tech.md' || \
  echo "WARNING: Code changed but docs not updated"
```

---

**Q: Should we keep docs in code repo or separate wiki?**

A: **Recommended**: In code repo
- ✅ Versioned with code
- ✅ Part of code review process
- ✅ Single source of truth
- ✅ Works offline

Optional: Mirror important docs to wiki for visibility.

