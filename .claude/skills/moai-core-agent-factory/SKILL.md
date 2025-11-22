---
name: moai-core-agent-factory
description: Intelligent Claude Code agent generation with requirement analysis, domain detection, and template system
version: 2.0.0
last_updated: 2025-11-22
---

# Agent Factory Intelligence Engine

**Enterprise-grade AI-powered agent generation for Claude Code platform**

> Master skill for automatic production-ready sub-agent creation through intelligent requirement analysis, research, template generation, and validation

**Version**: 2.0.0
**Status**: Production Ready
**Components**: 6 core systems + enterprise features

---

## Quick Reference (30 seconds)

### The 6 Core Components

| Component | Purpose | Learn More |
|-----------|---------|-----------|
| **Intelligence Engine** | Analyze requirements → domain/capabilities/complexity | [advanced-patterns.md](modules/advanced-patterns.md#intelligence-engine) |
| **Research Engine** | Context7 MCP workflow → fetch docs → extract practices | [advanced-patterns.md](modules/advanced-patterns.md#research-engine) |
| **Template System** | 3-tier templates (Simple/Standard/Complex) + variables | [advanced-patterns.md](modules/advanced-patterns.md#template-system) |
| **Validation Framework** | 4 quality gates + test cases | [advanced-patterns.md](modules/advanced-patterns.md#validation-framework) |
| **Advanced Features** | Versioning, multi-domain, optimization | [advanced-patterns.md](modules/advanced-patterns.md#advanced-features) |
| **Practical Examples** | 3 main test cases + edge cases | [examples.md](examples.md) |

---

## What It Does

Agent Factory generates production-ready Claude Code sub-agents through intelligent workflow:

```
User Requirement
    ↓
[Intelligence Engine]
  ├─ Parse requirement
  ├─ Detect domain (primary + secondary)
  ├─ Score complexity (1-10)
  └─ Select model (Sonnet/Haiku/Inherit)
    ↓
[Research Engine] - Context7 MCP
  ├─ Resolve libraries
  ├─ Fetch documentation
  ├─ Extract best practices
    ↓
[Template System]
  ├─ Select template tier (1-3)
  ├─ Generate agent markdown
    ↓
[Validation Framework]
  ├─ 4 quality gates
  ├─ TRUST 5 + Claude Code compliance
    ↓
Production-Ready Agent ✅
```

---

## Core Capabilities

### 1. Intelligent Requirement Analysis
- Extract domain, capabilities, complexity from natural language
- Auto-detect primary + secondary domains with confidence scores
- Complexity scoring (1-10 scale) for model selection
- Tool permission calculator

### 2. Research Engine with Context7 MCP
- Library ID resolution for 100+ frameworks
- Official documentation fetching via Context7
- Best practice extraction from latest sources
- Pattern identification and validation

### 3. 3-Tier Template System
- **Tier 1 (Simple)**: ~200 lines, Haiku, <5 min
- **Tier 2 (Standard)**: 200-500 lines, Inherit/Sonnet, <15 min
- **Tier 3 (Complex)**: 500+ lines, Sonnet, 20-30 min
- 15+ variable categories for customization

### 4. Validation Framework
- Gate 1: YAML syntax validation
- Gate 2: Required sections verification
- Gate 3: Content quality checks
- Gate 4: TRUST 5 + Claude Code compliance
- 5 core + 3 edge case test scenarios

### 5. Advanced Features
- Semantic versioning for agent tracking
- Multi-domain agent support (2-3 domains)
- Performance optimization suggestions
- Enterprise compliance (SOC2, GDPR, HIPAA)
- Audit logging and monitoring

---

## How to Use This Skill

### For Agent-Factory Agent

```python
# 1. Load this skill
skill = Skill("moai-core-agent-factory")

# 2. Intelligence Engine: Analyze requirement
domain = skill.intelligence_engine.detect_domain(user_input)
complexity = skill.intelligence_engine.score_complexity(domain)
model = skill.intelligence_engine.select_model(complexity)

# 3. Research Engine: Get best practices
research = skill.research_engine.research_domain(domain)
practices = research.best_practices
patterns = research.patterns

# 4. Template System: Generate agent
template = skill.template_system.select_template(complexity)
agent_markdown = skill.template_system.generate_agent(
    template=template,
    variables={...}
)

# 5. Validation: Quality check
validation = skill.validation_framework.validate(agent_markdown)
if validation.passed:
    return agent_markdown
```

### For Reference Lookup

- **"What model should I pick?"** → See [advanced-patterns.md#model-selection](modules/advanced-patterns.md#model-selection)
- **"How do I get best practices?"** → See [advanced-patterns.md#research-engine](modules/advanced-patterns.md#research-engine)
- **"What variables exist?"** → See [advanced-patterns.md#template-system](modules/advanced-patterns.md#template-system)
- **"How are agents validated?"** → See [advanced-patterns.md#validation-framework](modules/advanced-patterns.md#validation-framework)

---

## Performance Expectations

| Agent Type | Complexity | Generation Time | Result |
|-----------|-----------|-----------------|--------|
| Simple | 1-3 | <5 min | Tier 1 template |
| Standard | 4-6 | <15 min | Tier 2 template |
| Complex | 7-10 | 20-30 min | Tier 3 + orchestration |

---

## Three-Level Learning Path

### Level 1: Fundamentals (This File + examples.md)
- Quick reference overview
- Practical use cases
- Core workflow understanding

### Level 2: Advanced Patterns
- See [modules/advanced-patterns.md](modules/advanced-patterns.md)
- 6 core systems detailed
- Integration patterns
- Enterprise features

### Level 3: Optimization & Best Practices
- See [modules/optimization.md](modules/optimization.md)
- Performance tuning
- Token optimization
- Production deployment

---

## Best Practices

### DO
- ✅ Use Intelligence Engine for accurate domain detection
- ✅ Leverage Research Engine with Context7 for latest patterns
- ✅ Select appropriate template tier based on complexity
- ✅ Run all 4 validation gates before deployment
- ✅ Monitor agent performance and usage patterns
- ✅ Update agents when research reveals new patterns

### DON'T
- ❌ Skip Research Engine for domain knowledge
- ❌ Use wrong template tier (too simple = incomplete, too complex = token waste)
- ❌ Ignore validation framework warnings
- ❌ Hardcode domain assumptions without Intelligence Engine
- ❌ Deploy without TRUST 5 + Claude Code compliance check

---

## Integration Points

### With agent-factory Agent
```yaml
---
name: agent-factory
model: sonnet
---

## Required Skills
Skill("moai-core-agent-factory")

# Execution:
1. Load this skill
2. Use Intelligence Engine (requirement analysis)
3. Use Research Engine (best practices)
4. Use Template System (generation)
5. Use Validation Framework (quality check)
```

### With Predecessor/Successor Agents
- **Receives from**: plan-agent (requirement specification)
- **Delegates to**: cc-manager (Claude Code compliance check)
- **Integrates with**: mcp-context7-integrator (library research)

---

## File Organization

```
moai-core-agent-factory/
├── SKILL.md                    (this file - overview)
├── examples.md                 (10+ practical examples)
├── modules/
│   ├── advanced-patterns.md    (6 core systems detailed)
│   └── optimization.md         (performance & token optimization)
├── reference.md                (quick lookup reference)
└── templates/                  (agent templates)
    ├── simple-agent.template.md
    ├── standard-agent.template.md
    ├── complex-agent.template.md
    └── VARIABLE_REFERENCE.md
```

---

## Key Highlights

✅ **Comprehensive**: 6 core systems with complete documentation
✅ **Modular**: Each system independently referenceable
✅ **Practical**: Algorithms with code examples
✅ **Tested**: 5 core + 3 edge case scenarios
✅ **Enterprise**: Versioning, compliance, optimization
✅ **Official**: Follows Claude Code standards

---

## Related Skills

- `moai-core-workflow` - Agent workflow orchestration
- `moai-core-config-schema` - Configuration management
- `moai-domain-testing` - Test strategies for agents
- `moai-essentials-debug` - Debugging agent issues

---

**Version**: 2.0.0
**Status**: Production Ready (2025-11-22)
**Maintained By**: Agent Factory (MoAI-ADK)
