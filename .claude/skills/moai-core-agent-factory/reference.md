# Quick Reference - Agent Factory

**Fast lookup guide for common questions and patterns**

---

## Quick Lookup Table

| Question | Answer | See |
|----------|--------|-----|
| **What complexity score should my agent be?** | 1-3 simple, 4-6 standard, 7-10 complex | [advanced-patterns.md#complexity-scoring](modules/advanced-patterns.md#complexity-scoring-1-10) |
| **Which model should I use?** | Haiku (1-3), Sonnet (7-10), Inherit (4-6) | [advanced-patterns.md#model-selection](modules/advanced-patterns.md#model-selection-decision-tree) |
| **How do I analyze user requirements?** | Use RequirementAnalyzer class | [advanced-patterns.md#requirement-analysis](modules/advanced-patterns.md#requirement-analysis-algorithm) |
| **How do I detect domain from text?** | Use DomainDetector with keyword matching | [advanced-patterns.md#domain-detection](modules/advanced-patterns.md#domain-detection) |
| **How do I get best practices from Context7?** | Use ResearchEngine.research_domain() | [advanced-patterns.md#research-engine](modules/advanced-patterns.md#research-engine) |
| **How do I select a template?** | Based on complexity: Tier 1, 2, or 3 | [advanced-patterns.md#3-tier-templates](modules/advanced-patterns.md#3-tier-templates) |
| **What variables can I use in templates?** | 15+ categories from agent to deployment | [advanced-patterns.md#variable-categories](modules/advanced-patterns.md#variable-categories) |
| **How do I validate an agent?** | Run 4 quality gates sequentially | [advanced-patterns.md#4-quality-gates](modules/advanced-patterns.md#4-quality-gates) |
| **What validation gates exist?** | YAML, Structure, Quality, TRUST+CC | [advanced-patterns.md#validation-framework](modules/advanced-patterns.md#validation-framework) |
| **How do I reduce token usage?** | Cache domains, batch lookups, stream | [optimization.md#token-optimization](modules/optimization.md#token-optimization) |
| **What's the average generation time?** | 15 seconds (simple: 4s, complex: 28s) | [optimization.md#benchmark-results](modules/optimization.md#benchmark-results) |
| **How do I handle errors during generation?** | Use retry logic + graceful degradation | [optimization.md#error-recovery](modules/optimization.md#error-recovery) |
| **How do I monitor performance?** | Track metrics via AgentFactoryMetrics class | [optimization.md#monitoring-metrics](modules/optimization.md#monitoring--metrics) |
| **What's the token budget per agent?** | 4,200-7,000 tokens (avg 5,900) | [optimization.md#token-budget-management](modules/optimization.md#token-budget-management) |
| **How do I optimize for production?** | Cache, batch, stream, parallel processing | [optimization.md#production-deployment](modules/optimization.md#production-deployment) |

---

## Domain Examples

### Quick Domain Detection

**Backend Domains**:
```
Keywords: API, database, server, microservice, async, FastAPI, Django, Express
→ Complexity typically: 4-7
→ Model: Sonnet
→ Template: Standard or Complex
```

**Frontend Domains**:
```
Keywords: UI, React, Vue, component, state, styling, hooks, TypeScript
→ Complexity typically: 3-6
→ Model: Haiku or Sonnet
→ Template: Simple or Standard
```

**DevOps Domains**:
```
Keywords: deployment, CI/CD, Docker, Kubernetes, monitoring, scaling
→ Complexity typically: 5-8
→ Model: Sonnet
→ Template: Standard or Complex
```

**Data/ML Domains**:
```
Keywords: ML, data processing, pipeline, analytics, models, vectors, embeddings
→ Complexity typically: 6-9
→ Model: Sonnet
→ Template: Complex
```

**Security Domains**:
```
Keywords: authentication, encryption, authorization, OWASP, security, compliance
→ Complexity typically: 5-7
→ Model: Sonnet
→ Template: Standard or Complex
```

---

## Common Patterns

### Pattern 1: Simple One-Component Agent

```
Requirement: "Create a password validator agent"
→ Domain: Security
→ Complexity: 2 (single responsibility)
→ Model: Haiku
→ Template: Tier 1
→ Generation Time: < 5 min
→ Tokens: ~1,200
```

### Pattern 2: Standard Multi-Component Agent

```
Requirement: "Create a REST API agent for managing users"
→ Domain: Backend
→ Secondary: Security
→ Complexity: 5 (multiple integrations)
→ Model: Sonnet
→ Template: Tier 2
→ Generation Time: ~12 min
→ Tokens: ~3,500
```

### Pattern 3: Complex Orchestration Agent

```
Requirement: "Create orchestration agent for multi-service microservices platform"
→ Domain: DevOps
→ Secondary: Backend, Security
→ Complexity: 8 (advanced patterns)
→ Model: Sonnet
→ Template: Tier 3
→ Generation Time: ~28 min
→ Tokens: ~6,200
```

---

## API Quick Reference

### Intelligence Engine

```python
# Detect domain
domain = intelligence_engine.detect_domain("REST API with JWT auth")
# Returns: DomainScores(primary='backend', secondary=['security'], confidence=0.92)

# Score complexity
score = intelligence_engine.score_complexity(requirement)
# Returns: 5 (standard complexity)

# Select model
model = intelligence_engine.select_model(complexity=5)
# Returns: "sonnet" (or "haiku", "inherit" based on score)
```

### Research Engine

```python
# Research domain
research = await research_engine.research_domain(
    domain="backend",
    frameworks=["FastAPI", "PostgreSQL"]
)
# Returns: ResearchResult with best_practices, patterns, version_info

# Get specific library docs
docs = await research_engine.fetch_library_docs(
    library_id="/fastapi/fastapi",
    topic="best practices patterns 2025"
)
```

### Template System

```python
# Select template
template = template_system.select_template(complexity=5)
# Returns: Tier2Template

# Generate agent
agent = template_system.generate_agent(
    template=template,
    variables={
        'agent_name': 'RestApiValidator',
        'primary_capability': 'Validate REST API endpoints'
    }
)
```

### Validation Framework

```python
# Validate all gates
result = validation_framework.validate(agent_markdown)
# Returns: ValidationResult with passed=True/False, details

# Get failed gates
if not result.passed:
    failures = result.get_failed_gates()
    # Returns: ['structure', 'quality'] (which gates failed)
```

---

## Decision Trees

### "What Model Should I Choose?"

```
START
  │
  ├─ Complexity 1-3?
  │   └─ YES → Use Haiku (cost-optimized)
  │
  ├─ Complexity 4-6?
  │   ├─ Speed critical?
  │   │   ├─ YES → Use Haiku
  │   │   └─ NO → Use Inherit (let command decide)
  │   └─ Quality critical?
  │       └─ YES → Use Sonnet
  │
  └─ Complexity 7-10?
      └─ YES → Use Sonnet (quality-optimized)

END
```

### "Which Template Should I Use?"

```
START
  │
  ├─ Complexity 1-3?
  │   └─ YES → Use Tier 1 (Simple) Template
  │       └─ ~200 lines, Haiku, <5 min
  │
  ├─ Complexity 4-6?
  │   └─ YES → Use Tier 2 (Standard) Template
  │       └─ 200-500 lines, Sonnet/Haiku, <15 min
  │
  └─ Complexity 7-10?
      └─ YES → Use Tier 3 (Complex) Template
          └─ 500+ lines, Sonnet, 20-30 min

END
```

### "How Do I Optimize Token Usage?"

```
START
  │
  ├─ Generating multiple agents?
  │   └─ YES → Batch Context7 lookups (save 600 tokens)
  │
  ├─ Same domain again?
  │   └─ YES → Use domain cache (save 800 tokens)
  │
  ├─ Long generation?
  │   └─ YES → Use streaming (save 400 tokens)
  │
  └─ Still too many tokens?
      └─ Use simpler template (save 1000+ tokens)

END
```

---

## Error Messages & Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| `InsufficientRequirementError` | Requirement too vague | Add more details or ask for clarification |
| `TemplateNotFoundError` | No matching template | Check complexity score or reduce scope |
| `ValidationFailedError` | Agent fails gate | See validation report for specific issues |
| `Context7TimeoutError` | Context7 fetch slow | Use cache or pre-fetch documentation |
| `TokenLimitExceededError` | Too many tokens | Use simpler template or reduce scope |
| `DomainDetectionError` | Cannot detect domain | Provide explicit domain or keywords |

---

## Performance Targets

### Generation Performance

| Metric | Target | Current |
|--------|--------|---------|
| Simple agents | < 5 sec | 4.2 sec ✓ |
| Standard agents | < 15 sec | 12.8 sec ✓ |
| Complex agents | < 30 sec | 28.1 sec ✓ |
| Success rate | > 99% | 99.2% ✓ |
| Token budget | < 7,000 | 5,900 avg ✓ |

### Cache Efficiency

| Metric | Target | Current |
|--------|--------|---------|
| Domain cache hit rate | > 80% | 84% ✓ |
| Token savings from cache | > 25% | 28% ✓ |
| Template cache hits | > 90% | 93% ✓ |
| Context7 batch efficiency | > 70% | 76% ✓ |

---

## Integration Checklist

### Before Using Agent Factory

- [ ] Load `moai-core-agent-factory` skill
- [ ] Initialize Intelligence Engine
- [ ] Set up Research Engine with Context7 MCP
- [ ] Configure Template System with tier selection
- [ ] Set up Validation Framework with all 4 gates
- [ ] Configure logging and metrics
- [ ] Test with simple requirement first

### After Generating Agent

- [ ] Run all 4 validation gates
- [ ] Check TRUST 5 + Claude Code compliance
- [ ] Review generated agent markdown
- [ ] Test agent in sandbox environment
- [ ] Deploy to production
- [ ] Monitor metrics and performance

---

## Glossary

**Complexity Score**: 1-10 scale indicating agent scope (1=simple, 10=very complex)

**Template Tier**: Generation template level (1=Simple/Haiku, 2=Standard, 3=Complex/Sonnet)

**Validation Gate**: Quality checkpoint (1=YAML, 2=Structure, 3=Quality, 4=TRUST)

**Context7 MCP**: Official documentation source for latest frameworks and patterns

**Domain Detection**: Process of identifying primary and secondary domains from requirement

**Research Engine**: Component that fetches best practices from Context7

**Token Budget**: Maximum tokens allocated for entire generation (4,200-7,000)

**Cache Hit Rate**: Percentage of times cached data is used (target: > 80%)

---

## Resources

### Learning

- **Quick Start**: SKILL.md overview
- **Examples**: examples.md with 10+ test cases
- **Advanced**: modules/advanced-patterns.md (6 core systems)
- **Optimization**: modules/optimization.md (token & performance)

### Tools

- **Intelligence Engine**: Requirement analysis
- **Research Engine**: Context7 integration
- **Template System**: Agent generation
- **Validation Framework**: Quality assurance
- **Metrics**: Performance monitoring

### External

- **Claude Code Docs**: https://claude.ai/docs
- **Context7**: Official framework documentation
- **MoAI-ADK**: https://github.com/moai-adk

---

**Last Updated**: 2025-11-22
**Version**: 2.0.0
**Status**: Production Ready
