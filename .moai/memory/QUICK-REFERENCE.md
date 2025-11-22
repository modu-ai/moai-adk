# MoAI-ADK Agent-Skill Quick Reference

**Version**: 2.0
**Updated**: 2025-11-22
**For**: Quick lookup of agent skills and delegation patterns

---

## 31 Agents with Primary Skills

### Planning & Design (4)

1. **spec-builder** (6 skills)
   - moai-foundation-ears, moai-foundation-specs, moai-core-spec-authoring
   - moai-core-ask-user-questions, moai-spec-intelligent-workflow, moai-lang-python

2. **api-designer** (5 skills)
   - moai-domain-backend, moai-domain-web-api, moai-security-api
   - moai-lang-typescript, moai-context7-lang-integration

3. **implementation-planner** (6 skills)
   - moai-foundation-specs, moai-core-workflow, moai-core-todowrite-pattern
   - moai-foundation-trust, moai-core-dev-guide, moai-lang-python

4. **component-designer** (6 skills)
   - moai-domain-frontend, moai-design-systems, moai-component-designer
   - moai-lib-shadcn-ui, moai-lang-typescript, moai-icons-vector

### Implementation (8)

5. **backend-expert** (8 skills)
   - moai-domain-backend, moai-domain-database, moai-security-api, moai-security-auth
   - moai-lang-python, moai-lang-go, moai-context7-lang-integration, moai-essentials-perf

6. **frontend-expert** (9 skills)
   - moai-domain-frontend, moai-design-systems, moai-lib-shadcn-ui, moai-streaming-ui
   - moai-lang-typescript, moai-lang-javascript, moai-lang-tailwind-css
   - moai-icons-vector, moai-essentials-perf

7. **tdd-implementer** (10 skills)
   - moai-foundation-trust, moai-core-dev-guide, moai-domain-testing, moai-essentials-review
   - moai-essentials-refactor, moai-essentials-debug
   - moai-lang-python, moai-lang-typescript, moai-domain-backend, moai-domain-frontend

8. **database-expert** (8 skills)
   - moai-domain-database, moai-lang-sql, moai-essentials-perf, moai-lang-python
   - moai-domain-cloud, moai-baas-supabase-ext, moai-baas-neon-ext, moai-security-encryption

9. **migration-expert** (5 skills)
   - moai-domain-database, moai-lang-sql, moai-lang-python
   - moai-essentials-perf, moai-domain-backend

10. **devops-expert** (7 skills)
    - moai-domain-devops, moai-domain-cloud, moai-cloud-aws-advanced, moai-cloud-gcp-advanced
    - moai-domain-monitoring, moai-security-secrets, moai-baas-vercel-ext

11. **accessibility-expert** (4 skills)
    - moai-domain-frontend, moai-design-systems, moai-domain-testing, moai-playwright-webapp-testing

12. **ui-ux-expert** (7 skills)
    - moai-domain-frontend, moai-design-systems, moai-lib-shadcn-ui, moai-component-designer
    - moai-lang-tailwind-css, moai-icons-vector, moai-streaming-ui

### Quality (3)

13. **quality-gate** (8 skills)
    - moai-foundation-trust, moai-essentials-review, moai-core-code-reviewer, moai-domain-testing
    - moai-essentials-debug, moai-essentials-perf, moai-essentials-refactor, moai-domain-security

14. **security-expert** (9 skills)
    - moai-domain-security, moai-security-owasp, moai-security-identity, moai-security-threat
    - moai-security-api, moai-security-auth, moai-security-encryption
    - moai-security-compliance, moai-security-zero-trust

15. **trust-checker** (5 skills)
    - moai-foundation-trust, moai-essentials-review, moai-core-code-reviewer
    - moai-domain-testing, moai-essentials-debug

### Documentation (3)

16. **doc-syncer** (6 skills)
    - moai-docs-generation, moai-docs-validation, moai-docs-toolkit, moai-docs-unified
    - moai-foundation-specs, moai-mermaid-diagram-expert

17. **docs-manager** (8 skills)
    - moai-docs-generation, moai-docs-validation, moai-docs-toolkit, moai-docs-unified
    - moai-project-documentation, moai-readme-expert
    - moai-mermaid-diagram-expert, moai-cc-claude-md

18. **sync-manager** (5 skills)
    - moai-docs-generation, moai-docs-validation, moai-docs-unified
    - moai-foundation-git, moai-change-logger

### DevOps (2)

19. **devops-expert** (see #10 above)

20. **monitoring-expert** (5 skills)
    - moai-domain-monitoring, moai-domain-cloud, moai-domain-devops
    - moai-essentials-perf, moai-essentials-debug

### Optimization (2)

21. **performance-engineer** (7 skills)
    - moai-essentials-perf, moai-domain-monitoring, moai-essentials-debug
    - moai-domain-backend, moai-domain-frontend, moai-domain-database

22. **format-expert** (5 skills)
    - moai-essentials-refactor, moai-lang-python, moai-lang-typescript
    - moai-core-practices, moai-foundation-trust

### Integration (5)

23. **mcp-context7-integrator** (4 skills)
    - moai-context7-lang-integration, moai-context7-integration
    - moai-mcp-integration, moai-jit-docs-enhanced

24. **mcp-figma-integrator** (7 skills)
    - moai-domain-figma, moai-design-systems, moai-lang-typescript, moai-domain-frontend
    - moai-mcp-integration, moai-component-designer, moai-lib-shadcn-ui

25. **mcp-notion-integrator** (4 skills)
    - moai-domain-notion, moai-mcp-integration
    - moai-docs-generation, moai-document-processing

26. **mcp-playwright-integrator** (4 skills)
    - moai-playwright-webapp-testing, moai-mcp-integration
    - moai-domain-testing, moai-domain-frontend

27. **debug-helper** (6 skills)
    - moai-essentials-debug, moai-essentials-perf
    - moai-domain-backend, moai-domain-frontend
    - moai-lang-python, moai-lang-typescript

### Management (4)

28. **project-manager** (6 skills)
    - moai-cc-configuration, moai-project-config-manager, moai-project-documentation
    - moai-project-language-initializer, moai-project-batch-questions, moai-core-ask-user-questions

29. **git-manager** (3 skills)
    - moai-foundation-git, moai-change-logger, moai-core-session-state

30. **agent-factory** (7 skills)
    - moai-core-agent-factory, moai-foundation-ears, moai-foundation-specs
    - moai-core-language-detection, moai-core-agent-guide
    - moai-context7-lang-integration, moai-core-ask-user-questions

31. **cc-manager** (7 skills)
    - moai-cc-configuration, moai-cc-hooks, moai-cc-settings, moai-cc-agents
    - moai-cc-skills, moai-cc-commands, moai-cc-memory

32. **skill-factory** (6 skills)
    - moai-cc-skill-factory, moai-cc-skills, moai-core-ask-user-questions
    - moai-foundation-trust, moai-docs-generation, moai-context7-lang-integration

---

## Quick Delegation Patterns

### Pattern 1: Basic Delegation with Skills
```python
Task(
  subagent_type="backend-expert",
  prompt="""
  Implement authentication API.

  Load skills:
  - moai-domain-backend
  - moai-security-api
  - moai-lang-python
  """
)
```

### Pattern 2: Multi-Domain Task
```python
Task(
  subagent_type="backend-expert",
  prompt="""
  Implement with cross-domain requirements.

  Load skills:
  - moai-domain-backend (core)
  - moai-security-api (security)
  - moai-essentials-perf (optimization)
  - moai-domain-database (database)
  """
)
```

### Pattern 3: Quality-First (TRUST 5)
```python
Task(
  subagent_type="tdd-implementer",
  prompt="""
  Implement with TRUST 5 compliance.

  Load TRUST skills:
  - moai-foundation-trust
  - moai-essentials-review
  - moai-domain-testing
  - moai-core-dev-guide
  """
)
```

---

## Skill Categories Quick Reference

| Category | Pattern | Examples |
|----------|---------|----------|
| Foundation | moai-foundation-* | trust, ears, specs, git |
| Core | moai-core-* | agent-factory, workflow, dev-guide |
| Domain | moai-domain-* | backend, frontend, database, security |
| Language | moai-lang-* | python, typescript, sql |
| Essential | moai-essentials-* | debug, perf, review, refactor |
| Security | moai-security-* | owasp, api, auth, encryption |
| Docs | moai-docs-* | generation, validation, unified |
| Claude Code | moai-cc-* | configuration, skills, agents |

---

## Agent Selection Quick Guide

**Need to...** → **Use Agent** → **With Skills**

- Plan feature → spec-builder → ears, specs, ask-questions
- Design API → api-designer → web-api, security-api, typescript
- Implement backend → backend-expert → backend, security-api, python
- Implement frontend → frontend-expert → frontend, design-systems, typescript
- TDD implementation → tdd-implementer → trust, dev-guide, testing
- Database work → database-expert → database, sql, performance
- Validate quality → quality-gate → trust, review, testing
- Security audit → security-expert → security, owasp, auth
- Generate docs → doc-syncer → docs-generation, specs, mermaid
- Deploy → devops-expert → devops, cloud, monitoring
- Debug issue → debug-helper → debug, performance, backend/frontend

---

## Skill Loading Order

**Recommended loading sequence**:
1. Foundation (moai-foundation-trust, moai-foundation-specs)
2. Domain (moai-domain-backend, moai-domain-frontend)
3. Security (moai-security-api, moai-security-auth)
4. Language (moai-lang-python, moai-lang-typescript)
5. Essential (moai-essentials-perf, moai-essentials-debug)
6. Specialized (moai-lib-shadcn-ui, moai-mermaid-diagram-expert)

---

## Quick Stats

- **Total Agents**: 31
- **Total Skills**: 138 available
- **Skill Assignments**: 287 recommended
- **Average Skills/Agent**: 9.3
- **Agents with 6+ Skills**: 84%
- **Critical Gaps Resolved**: 10/10 (100%)

---

**For Complete Details**: See `.moai/reports/agents-complete-analysis.md`
**For Patterns**: See `.moai/memory/delegation-patterns.md`
**For Discovery**: See `.moai/memory/skills.md`
**For Rules**: See `.moai/memory/execution-rules.md`
