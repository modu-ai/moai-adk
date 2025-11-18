# Agent Factory - Quick Reference Guide

## Algorithmic Quick Reference

### 1. Requirement Analysis
```python
def analyze_requirement(user_input):
    domain = extract_domain(user_input)              # "backend"
    capabilities = extract_capabilities(user_input)   # ["create", "optimize"]
    complexity = assess_complexity(user_input, domain, capabilities)  # 5
    frameworks = extract_frameworks(user_input)       # ["FastAPI"]
    ambiguities = detect_ambiguities(domain, capabilities, frameworks)  # []
    return {domain, capabilities, complexity, frameworks, ambiguities}
```

### 2. Domain Detection
```python
detector = DomainDetector()
primary, confidence = detector.detect_primary_domain(user_input)
secondary = detector.detect_secondary_domains(primary, user_input)
# Result: primary="backend", confidence=0.95, secondary=["database"]
```

### 3. Complexity Scoring (1-10)
```
1-3:  Simple (formatting, linting, utilities)
4-6:  Standard (domain experts, integrators)
7-10: Complex (orchestrators, research-heavy)
```

### 4. Model Selection Decision Tree
```
Research needed?
├─ YES → SONNET (always for research)
└─ NO:
   Complexity >= 7?
   ├─ YES → SONNET (complex reasoning)
   └─ NO:
      Speed critical AND Simple?
      ├─ YES → HAIKU (fast execution)
      └─ NO → INHERIT (context-dependent)
```

### 5. Tool Permission Calculation
```
CORE_TOOLS = ["Read", "Grep", "Glob"]  # Always

Add by Domain:
- backend:      [Write, Edit, Bash, WebFetch]
- frontend:     [Write, Edit, MultiEdit]
- database:     [Bash, Write, Edit]
- security:     [Read, Grep, Bash]
- testing:      [Bash, Write]
- devops:       [Bash, Write, Edit]
- documentation:[Write, Edit, WebFetch]

Add by Capability:
- create:       [Write, Edit]
- analyze:      [Read, Grep]
- research:     [WebFetch, WebSearch, mcp_tools]
- integrate:    [Bash, WebFetch]
- optimize:     [Bash, Edit]
- validate:     [Bash]

Add by Complexity:
- >= 6:         [AskUserQuestion]
```

### 6. Skill Recommendation
```
AUTO_SKILLS:
  - moai-core-agent-guide
  - moai-core-workflow
  - moai-domain-{primary_domain}

CONDITIONAL_SKILLS:
  - moai-core-language-detection
  - moai-domain-{secondary_domains}
  - moai-lang-{language}
  - moai-essentials-{capabilities}
  - moai-foundation-specs (if complexity >= 7)
```

## Domain-to-Skills Mapping

```
backend         → moai-domain-backend
frontend        → moai-domain-frontend
database        → moai-domain-database
security        → moai-domain-security
testing         → moai-domain-testing
devops          → moai-domain-cloud, moai-domain-devops
performance     → moai-essentials-perf
documentation   → moai-docs-generation

create          → moai-essentials-refactor
analyze         → moai-essentials-debug, moai-core-code-reviewer
optimize        → moai-essentials-perf, moai-essentials-refactor
research        → moai-context7-lang-integration
integrate       → moai-cc-mcp-plugins
validate        → moai-foundation-trust
```

## Language-to-Skills Mapping

```
python          → moai-lang-python
typescript      → moai-lang-typescript
javascript      → moai-lang-javascript
go              → moai-lang-go
java            → moai-lang-java
rust            → moai-lang-rust
php             → moai-lang-php
c#              → moai-lang-csharp
```

## Template Selection Criteria

```
Complexity Score:
├─ 1-3:   Tier 1 (Simple, ~200 lines, <5 min, Haiku)
├─ 4-6:   Tier 2 (Standard, 200-500 lines, <15 min, Inherit/Sonnet)
└─ 7-10:  Tier 3 (Complex, 500+ lines, 20-30 min, Sonnet)

Characteristic Count:
├─ 1 domain + 1 capability:           Tier 1
├─ 1 domain + 2-3 capabilities:       Tier 2
├─ Multiple domains OR research:      Tier 3
└─ Multi-domain + research + MCP:     Tier 3 (enhanced)
```

## Research Quality Thresholds

```
documentation_coverage:  >= 70%
best_practice_count:     >= 5
code_example_count:      >= 3
pattern_diversity:       >= 3 different patterns
source_reliability:      >= 0.80
currency_score:          >= 0.75
overall_quality:         >= 0.70 (70%)
```

## Validation Gates Checklist

### Gate 1: Syntax Validation
- [ ] Valid YAML frontmatter
- [ ] No markdown syntax errors
- [ ] Proper heading hierarchy (h1 → h2 → h3)
- [ ] Code blocks have language specifiers

### Gate 2: Structure Validation
- [ ] All required sections present
- [ ] Required YAML fields: name, description, tools, model
- [ ] Tool list properly formatted as array
- [ ] Skill integration documented

### Gate 3: Content Validation
- [ ] Tools match capabilities
- [ ] Model selection justified
- [ ] Skills relevant to domain
- [ ] Language handling explicit
- [ ] Delegation rules clear
- [ ] DO/DO NOT sections specific

### Gate 4: Quality Integration
- [ ] TRUST 5 compliance (Test/Readable/Unified/Secured/Trackable)
- [ ] Claude Code   standards
- [ ] Enterprise requirements (if applicable)
- [ ] Pass quality-gate agent

## Common Domain Keywords

```
BACKEND:
  api, server, rest, graphql, endpoint, route, 
  database, authentication, microservice, fastapi, 
  express, django, spring

FRONTEND:
  ui, component, react, vue, angular, jsx, tsx,
  state, redux, styling, css, responsive, viewport

DATABASE:
  schema, migration, query, orm, sql, nosql,
  index, performance, normalization, relationship

SECURITY:
  auth, encryption, vulnerability, owasp, audit,
  token, jwt, oauth, certificate, ssl, threat

TESTING:
  test, coverage, integration, e2e, unit, mock,
  fixture, assertion, pytest, jest, vitest

DEVOPS:
  deploy, ci/cd, docker, kubernetes, terraform,
  infrastructure, container, pipeline, automation

PERFORMANCE:
  optimize, benchmark, profiling, cache, load,
  memory, speed, throughput, latency
```

## Common Capability Keywords

```
CREATE:
  generate, create, build, design, architect, write,
  implement, develop, construct

ANALYZE:
  analyze, review, audit, inspect, evaluate, assess,
  examine, diagnose, investigate

OPTIMIZE:
  optimize, improve, refactor, enhance, performance,
  speed, efficiency, streamline

INTEGRATE:
  integrate, connect, sync, combine, coordinate,
  merge, bridge, link

RESEARCH:
  research, investigate, explore, study, analyze,
  survey, examine patterns

VALIDATE:
  validate, verify, check, test, confirm, ensure,
  assert, validate

MONITOR:
  monitor, track, observe, measure, collect,
  gather metrics, surveillance
```

## MCP Integration Quick Start

### Context7 Resolution
```python
# Resolve library name to ID
library = "FastAPI"
library_id = await mcp__context7__resolve_library_id(library)
# → Result: "/tiangolo/fastapi"
```

### Fetch Documentation
```python
# Get official docs
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/tiangolo/fastapi",
    topic="Best practices and patterns",
    tokens=5000
)
```

### Research Workflow Steps
1. **Resolve**: Library name → Context7 ID
2. **Fetch**: Get official documentation (5000 tokens)
3. **Extract**: Parse practices, patterns, examples
4. **Validate**: Check quality (≥70% threshold)
5. **Synthesize**: Consolidate findings
6. **Fallback**: Use established patterns if MCP fails

## Quick Decision Matrix

| Requirement | Decision | Rationale |
|---|---|---|
| Simple formatting task | Tier 1, Haiku, <200 lines | Fast, minimal reasoning |
| Backend API design | Tier 2, Inherit, 200-500 lines | Medium complexity |
| OWASP audit + research | Tier 3, Sonnet, 500+ lines | Complex + research |
| Multi-domain (2+) | Tier 2/3, Sonnet | Increased coordination |
| Fast execution needed | Haiku if complexity ≤4 | Cost-optimized |
| Complex reasoning | Sonnet always | Quality over speed |
| Research required | Sonnet + Context7 | Evidence synthesis |

## Fallback Strategy

**When Context7 MCP unavailable**:
1. Use established patterns from 30+ existing agents
2. Apply WebFetch for framework documentation
3. Leverage Skill knowledge bases
4. Document fallback in agent comments
5. Note limitation in agent description
6. Quality score: ~0.65 (vs 0.75+ with MCP)

## Performance Benchmarks

| Operation | Duration |
|---|---|
| Parse requirements | ~200ms |
| Detect domain | ~100ms |
| Score complexity | ~50ms |
| Select model | <10ms |
| Calculate tools | ~100ms |
| Recommend skills | ~200ms |
| Research (Context7) | 5-10 sec |
| Generate template | 2-3 sec |
| Validate gates | 1-2 sec |
| **Total (simple)** | **<5 min** |
| **Total (standard)** | **<15 min** |
| **Total (complex)** | **20-30 min** |

## Success Metrics

```
Model Selection Accuracy:           >= 90%
Skill Recommendation Accuracy:      >= 85%
Tool Permission Appropriateness:    >= 95%
YAML Validity:                      100%
Content Completeness:               100%
Validation Gates Pass Rate:         >= 95%
Multi-model Testing:                Haiku & Sonnet
Enterprise Compliance:              SOC2/GDPR/HIPAA ready
```

## Variable Reference (15+ Categories)

### Identity Variables
- `{{AGENT_NAME}}` - kebab-case identifier
- `{{AGENT_ID}}` - Unique ID
- `{{VERSION}}` - Semantic version

### Description Variables
- `{{AGENT_DESCRIPTION}}` - Brief summary
- `{{PROACTIVE_TRIGGERS}}` - When to use

### Domain Variables
- `{{PRIMARY_DOMAIN}}` - Main domain
- `{{SECONDARY_DOMAINS}}` - Array of secondary

### Capability Variables
- `{{DOMAIN_CAPABILITIES}}` - Detected capabilities
- `{{SPECIALIZED_CAPABILITIES}}` - Custom capabilities

### Model Variables
- `{{MODEL_SELECTION}}` - sonnet/haiku/inherit
- `{{MODEL_JUSTIFICATION}}` - Why this model

### Tool Variables
- `{{TOOL_PERMISSIONS}}` - Full tool list
- `{{MCP_TOOLS}}` - MCP-specific tools
- `{{RESEARCH_TOOLS}}` - Context7 tools

### Skill Variables
- `{{AUTO_LOAD_SKILLS}}` - Always loaded
- `{{CONDITIONAL_SKILLS}}` - Load on-demand
- `{{SKILL_LOADING_PATTERN}}` - How to load

### Responsibility Variables
- `{{DO_RESPONSIBILITIES}}` - What agent does
- `{{DONT_RESPONSIBILITIES}}` - What it doesn't

### Workflow Variables
- `{{WORKFLOW_STEPS}}` - Sequential steps
- `{{DECISION_POINTS}}` - Branch logic
- `{{DELEGATION_TARGETS}}` - Other agents

### Advanced Variables
- `{{LANGUAGE_HANDLING}}` - Multi-language rules
- `{{MCP_INTEGRATION}}` - MCP server usage
- `{{ORCHESTRATION_METADATA}}` - Complex info
- `{{COMPLIANCE_REQUIREMENTS}}` - Enterprise needs

## Integration Points Checklist

### With agent-factory.md Workflow
- [ ] Stage 1: Requirements analysis executed
- [ ] Stage 2: Complexity assessment completed
- [ ] Stage 3: Research phase (if needed) delegated
- [ ] Stage 4: Template/Skills selected
- [ ] Stage 5: Agent specification generated
- [ ] Stage 6: All validation gates passed

### With @agent-cc-manager
- [ ] YAML syntax validated
- [ ] Tool permissions approved
- [ ] MCP configurations checked
- [ ] Claude Code   compliance verified

### With @agent-quality-gate
- [ ] TRUST 5 compliance checked
- [ ] Content accuracy verified
- [ ] Security audit passed
- [ ] Performance optimized

### With Context7 MCP
- [ ] Libraries resolved
- [ ] Documentation fetched
- [ ] Practices extracted
- [ ] Patterns identified
- [ ] Quality validated

---

**Version**: 1.0.0  
**Updated**: 2025-11-15  
**Status**: Production Ready

