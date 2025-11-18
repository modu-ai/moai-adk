---
name: moai-alfred-agent-factory
version: 1.0.0
status: production
description: |
  Agent Factory master skill for intelligent Claude Code agent generation.
  Includes requirement analysis, domain detection, model selection, template system,
  validation framework, and advanced features. Contains algorithms, workflows, and patterns.
allowed-tools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep"]
tags: ["agent-factory", "agent-generation", "intelligence-engine", "domain-detection", "model-selection", "context7", "templates", "validation"]
---

# Agent Factory Intelligence Engine

> **Master Skill for Intelligent Agent Generation**
>
> Complete reference for agent-factory agent to automatically generate production-ready
> Claude Code sub-agents through requirement analysis, research, template generation,
> and validation.

**Version**: 1.0.0
**Status**: Production Ready
**Components**: 6 core systems + advanced features

---

## 🎯 Quick Start: The 6 Core Components

| Component | Location | Purpose |
|-----------|----------|---------|
| **Intelligence Engine** | [reference/intelligence-engine.md](reference/intelligence-engine.md) | Analyze requirements → domain/capabilities/complexity |
| **Research Engine** | [reference/research-engine.md](reference/research-engine.md) | Context7 MCP workflow → fetch docs → extract practices |
| **Template System** | [reference/template-system.md](reference/template-system.md) | 3-tier templates (Simple/Standard/Complex) + variables |
| **Validation Framework** | [reference/validation-framework.md](reference/validation-framework.md) | 4 quality gates + test cases |
| **Advanced Features** | [reference/advanced-features.md](reference/advanced-features.md) | Versioning, multi-domain, optimization |
| **Practical Examples** | [examples.md](examples.md) | 3 main test cases + edge cases |

---

## 🚀 Agent Generation Workflow

```
User Requirement
    ↓
[Stage 1-2] Intelligence Engine
  ├─ Parse requirement
  ├─ Detect domain (primary + secondary)
  ├─ Score complexity (1-10)
  ├─ Select model (Sonnet/Haiku/Inherit)
  └─ Calculate tools & skills
    ↓
[Stage 3] Research Engine (Context7 MCP)
  ├─ Resolve libraries
  ├─ Fetch documentation
  ├─ Extract best practices
  └─ Synthesize evidence
    ↓
[Stage 4-5] Template System
  ├─ Select template tier (1-3)
  ├─ Generate agent markdown
  └─ Apply variable substitution
    ↓
[Stage 6] Validation Framework
  ├─ Gate 1: YAML syntax
  ├─ Gate 2: Structure completeness
  ├─ Gate 3: Content quality
  └─ Gate 4: TRUST 5 + Claude Code v4.0
    ↓
Production-Ready Agent ✅
```

---

## 📚 Understanding Each Component

### Intelligence Engine
**Responsible for**: Requirement analysis and decision making

**Key Algorithms**:
- Requirement Analysis: Extract domain, capabilities, complexity
- Domain Detection: Primary + secondary domains with confidence
- Complexity Scoring: 1-10 scale for model selection
- Model Selection: Sonnet/Haiku/Inherit decision tree
- Tool Calculator: Minimum necessary permissions
- Skill Recommender: Match to 128+ MoAI-ADK skills

**See**: [reference/intelligence-engine.md](reference/intelligence-engine.md)

### Research Engine
**Responsible for**: Fetching official documentation and best practices

**Key Workflows**:
- Library Resolution: Convert domain → Context7 library IDs
- Documentation Fetch: Pull official docs from Context7 MCP
- Best Practice Extraction: Identify actionable patterns
- Pattern Identification: Architectural patterns for domain
- Quality Validation: Ensure research meets 70%+ threshold
- Evidence Synthesis: Consolidate into recommendations
- Fallback Strategy: WebFetch if Context7 unavailable

**See**: [reference/research-engine.md](reference/research-engine.md)

### Template System
**Responsible for**: Generating agent markdown with correct structure

**Key Components**:
- **Tier 1 (Simple)**: ~200 lines, Haiku, <5 min generation
- **Tier 2 (Standard)**: 200-500 lines, Inherit/Sonnet, <15 min
- **Tier 3 (Complex)**: 500+ lines, Sonnet, 20-30 min
- **Variables**: 15+ categories for dynamic substitution
- **YAML Generation**: Frontmatter with proper metadata

**Template Files** (in `templates/` subfolder):
- `simple-agent.template.md` – Tier 1 template for simple agents
- `standard-agent.template.md` – Tier 2 template for standard agents
- `complex-agent.template.md` – Tier 3 template for complex agents
- `VARIABLE_REFERENCE.md` – Complete variable substitution guide

**See**: [reference/template-system.md](reference/template-system.md)

### Validation Framework
**Responsible for**: Ensuring agent quality meets all standards

**4 Quality Gates**:
1. **Gate 1**: YAML syntax validation
2. **Gate 2**: Required sections verification
3. **Gate 3**: Content quality checks
4. **Gate 4**: TRUST 5 + Claude Code v4.0 compliance

**Includes**: 5 core test cases + 3 edge cases

**See**: [reference/validation-framework.md](reference/validation-framework.md)

### Advanced Features
**Responsible for**: Enterprise capabilities and optimization

**Features**:
- Semantic Versioning: Track agent versions
- Multi-Domain Support: 2-3 domain agents
- Performance Analyzer: Automatic optimization suggestions
- Enterprise Compliance: SOC2, GDPR, HIPAA support
- Audit Logging: Activity tracking and monitoring

**See**: [reference/advanced-features.md](reference/advanced-features.md)

---

## 🔗 Integration Points

### With agent-factory Agent
```yaml
---
name: agent-factory
model: sonnet
---

## Required Skills
Skill("moai-alfred-agent-factory")  # This master skill

# In execution:
1. Load this skill
2. Use Intelligence Engine (Stage 1-2)
3. Use Research Engine (Stage 3)
4. Use Template System (Stage 4-5)
5. Use Validation Framework (Stage 6)
```

### With @agent-cc-manager
After generation, delegate to @agent-cc-manager for:
- Claude Code v4.0 compliance verification
- `.claude/settings.json` validation
- MCP server configuration check
- Hook registration validation

### With @agent-mcp-context7-integrator
Research Engine delegates to for:
- Library ID resolution
- Official documentation fetching
- Best practice identification
- Latest API version discovery

---

## 📊 Performance Expectations

| Agent Type | Complexity | Generation Time | Result |
|-----------|-----------|-----------------|--------|
| Simple | 1-3 | <5 min | Tier 1 template |
| Standard | 4-6 | <15 min | Tier 2 template |
| Complex | 7-10 | 20-30 min | Tier 3 template + orchestration |

---

## 🧠 How to Use This Skill

### For Agent-Factory Agent

```python
# 1. Load this skill
skill = Skill("moai-alfred-agent-factory")

# 2. Use Intelligence Engine for analysis
domain = skill.intelligence_engine.detect_domain(user_input)
complexity = skill.intelligence_engine.complexity_score(domain, capabilities)
model = skill.intelligence_engine.select_model(complexity)

# 3. Use Research Engine for documentation
research = skill.research_engine.research_domain(domain, frameworks)
practices = research.best_practices
patterns = research.patterns

# 4. Use Template System for generation
template = skill.template_system.select_template(complexity)
agent_markdown = skill.template_system.generate_agent(
    template=template,
    variables={...}
)

# 5. Use Validation Framework for quality
validation = skill.validation_framework.validate(agent_markdown)
if validation.passed:
    return agent_markdown
else:
    return validation.issues
```

### For Reference Lookup

Need specific information? Use the reference files:
- **"What model should I pick?"** → See [intelligence-engine.md](reference/intelligence-engine.md#model-selection)
- **"How do I get best practices?"** → See [research-engine.md](reference/research-engine.md#best-practice-extraction)
- **"What variables exist?"** → See [template-system.md](reference/template-system.md#variables)
- **"How are agents validated?"** → See [validation-framework.md](reference/validation-framework.md#quality-gates)

---

## 📖 Reference Files Organization

```
moai-alfred-agent-factory/
├── SKILL.md                        (this file - navigation hub)
├── reference/                      (detailed documentation)
│   ├── intelligence-engine.md      (280 lines: algorithms)
│   ├── research-engine.md          (320 lines: workflows)
│   ├── template-system.md          (280 lines: templates)
│   ├── validation-framework.md     (220 lines: testing)
│   └── advanced-features.md        (200 lines: enterprise)
├── templates/                      (reusable agent templates)
│   ├── simple-agent.template.md    (Tier 1: <200 lines, Haiku)
│   ├── standard-agent.template.md  (Tier 2: 200-500 lines)
│   ├── complex-agent.template.md   (Tier 3: 500+ lines)
│   └── VARIABLE_REFERENCE.md       (15+ variable categories)
├── examples.md                     (719 lines: use cases)
└── reference.md                    (400 lines: quick lookup)
```

---

## ✨ Key Highlights

✅ **Comprehensive**: 6 core systems with complete documentation
✅ **Modular**: Each system independently referenceable
✅ **Practical**: Includes algorithms with code examples
✅ **Tested**: 5 core + 3 edge case test scenarios
✅ **Enterprise**: Versioning, compliance, optimization
✅ **Official**: Follows Claude Code v4.0 standards

---

## 🎓 Learning Path

**New to agent-factory?**
1. Read this overview
2. Review [examples.md](examples.md) for practical cases
3. Dive into specific reference files as needed

**Integrating with your agent?**
1. Load this skill with `Skill("moai-alfred-agent-factory")`
2. Reference specific algorithms from this overview
3. Consult detailed docs in reference/ folder

**Need specific feature?**
Use the component table above and jump to the corresponding reference file.

---

## 📞 Quick Reference

| Question | Link |
|----------|------|
| How do I analyze user requirements? | [intelligence-engine.md → Requirement Analysis](reference/intelligence-engine.md) |
| How do I detect the domain? | [intelligence-engine.md → Domain Detection](reference/intelligence-engine.md) |
| How do I select the right model? | [intelligence-engine.md → Model Selection](reference/intelligence-engine.md) |
| How do I research best practices? | [research-engine.md → Research Workflow](reference/research-engine.md) |
| What templates are available? | [template-system.md → Templates](reference/template-system.md) |
| How do I validate agents? | [validation-framework.md → Quality Gates](reference/validation-framework.md) |
| Can I see examples? | [examples.md](examples.md) |
| Need a quick lookup? | [reference.md](reference.md) |

---

**Created**: 2025-11-15
**Version**: 1.0.0
**Status**: Production Ready
**Total Content**: 2,800+ lines across organized reference files
