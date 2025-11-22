# Advanced Patterns - Agent Factory

**Comprehensive guide to the 6 core systems of Agent Factory Intelligence Engine**

---

## Intelligence Engine

### Requirement Analysis Algorithm

```python
class RequirementAnalyzer:
    def analyze_requirement(self, user_input: str) -> RequirementAnalysis:
        """Extract domain, capabilities, and complexity from requirement."""

        # Step 1: Natural language parsing
        parsed = self.nlp_parser.parse(user_input)

        # Step 2: Extract key attributes
        attributes = {
            'domain': self._extract_domain(parsed),
            'capabilities': self._extract_capabilities(parsed),
            'constraints': self._extract_constraints(parsed),
            'scale': self._estimate_scale(parsed)
        }

        # Step 3: Validate and normalize
        normalized = self._normalize_attributes(attributes)

        return RequirementAnalysis(**normalized)
```

### Domain Detection

**Primary + Secondary Domain Scoring**:

```python
class DomainDetector:
    def detect_domains(self, requirement: str) -> DomainScores:
        """Detect primary and secondary domains with confidence."""

        domain_keywords = {
            'backend': ['API', 'database', 'server', 'microservice'],
            'frontend': ['UI', 'component', 'React', 'state', 'styling'],
            'devops': ['deployment', 'CI/CD', 'Docker', 'Kubernetes'],
            'security': ['authentication', 'encryption', 'authorization', 'OWASP'],
            'data': ['ML', 'data processing', 'analytics', 'pipeline'],
        }

        scores = {}
        for domain, keywords in domain_keywords.items():
            score = sum(1 for kw in keywords if kw.lower() in requirement.lower())
            scores[domain] = score / len(keywords)

        # Return sorted by confidence
        return sorted(scores.items(), key=lambda x: x[1], reverse=True)
```

### Complexity Scoring (1-10)

| Score | Characteristics | Model | Generation Time |
|-------|-----------------|-------|-----------------|
| 1-3   | Single component, < 200 lines, straightforward logic | Haiku | <5 min |
| 4-6   | Multiple components, 200-500 lines, moderate patterns | Sonnet/Haiku | <15 min |
| 7-10  | Complex system, 500+ lines, advanced patterns, multi-domain | Sonnet | 20-30 min |

**Scoring Formula**:
```python
def score_complexity(self, requirement: RequirementAnalysis) -> int:
    """Calculate complexity score (1-10)."""

    score = 0

    # Component count (0-3 points)
    score += min(3, len(requirement.capabilities) // 2)

    # Integration complexity (0-2 points)
    score += 2 if requirement.integrations > 3 else 1 if requirement.integrations else 0

    # Scale requirements (0-2 points)
    if requirement.scale == 'enterprise': score += 2
    elif requirement.scale == 'medium': score += 1

    # Special patterns (0-2 points)
    advanced_patterns = ['microservices', 'distributed', 'real-time', 'ML']
    score += min(2, sum(1 for p in advanced_patterns if p in requirement.text))

    # Multi-domain (0-1 point)
    score += 1 if len(requirement.secondary_domains) > 1 else 0

    return min(10, max(1, score))
```

### Model Selection Decision Tree

```
Complexity Score
  ├─ 1-3: Haiku (cost-optimized)
  ├─ 4-6: Inherit (use existing model in command)
  └─ 7-10: Sonnet (highest quality)

With Constraints:
  ├─ Speed critical: Always Haiku
  ├─ Quality critical: Always Sonnet
  └─ Default: Inherit (let command decide)
```

---

## Research Engine

### Context7 MCP Workflow

```python
class ResearchEngine:
    async def research_domain(self, domain: str, frameworks: List[str]) -> ResearchResult:
        """Research domain using Context7 MCP."""

        # Step 1: Resolve library IDs
        library_ids = await self._resolve_library_ids(domain, frameworks)

        # Step 2: Fetch documentation
        docs = await self.context7.get_library_docs(
            context7_library_id=library_ids[0],
            topic=f"{domain} best practices patterns 2025",
            tokens=5000
        )

        # Step 3: Extract best practices
        practices = self._extract_best_practices(docs)
        patterns = self._identify_patterns(docs)

        # Step 4: Synthesize evidence
        synthesis = self._synthesize_evidence(practices, patterns)

        return ResearchResult(
            best_practices=synthesis.practices,
            patterns=synthesis.patterns,
            version_info=synthesis.versions,
            references=docs.references
        )
```

### Library Resolution

**Framework to Context7 Library Mapping**:

```python
FRAMEWORK_TO_LIBRARY = {
    # Backend
    'FastAPI': '/fastapi/fastapi',
    'Django': '/django/django',
    'Flask': '/pallets/flask',
    'Express': '/expressjs/express',
    'Spring Boot': '/spring-projects/spring-boot',

    # Frontend
    'React': '/facebook/react',
    'Vue': '/vuejs/vue',
    'Angular': '/angular/angular',
    'Next.js': '/vercel/next.js',
    'Svelte': '/sveltejs/svelte',

    # Database
    'PostgreSQL': '/postgresql/postgresql',
    'MongoDB': '/mongodb/docs',
    'Redis': '/redis/redis',

    # DevOps
    'Docker': '/moby/moby',
    'Kubernetes': '/kubernetes/kubernetes',
    'GitHub Actions': '/actions/starter-workflows',
}

def resolve_library_ids(self, domain: str, frameworks: List[str]) -> List[str]:
    """Resolve frameworks to Context7 library IDs."""
    library_ids = []
    for framework in frameworks:
        if framework in FRAMEWORK_TO_LIBRARY:
            library_ids.append(FRAMEWORK_TO_LIBRARY[framework])
    return library_ids
```

### Best Practice Extraction

```python
def extract_best_practices(self, docs: DocumentSet) -> List[BestPractice]:
    """Extract actionable best practices from documentation."""

    practices = []

    # Pattern 1: Code examples with explanations
    for example in docs.code_examples:
        practice = BestPractice(
            title=self._extract_title(example),
            description=self._extract_description(example),
            code=example.code,
            reasoning=self._extract_reasoning(example),
            applies_to=['agents', 'production']
        )
        practices.append(practice)

    # Pattern 2: Security recommendations
    for recommendation in docs.security_sections:
        practice = BestPractice(
            title=f"Security: {recommendation.title}",
            description=recommendation.content,
            category='security',
            applies_to=['all_agents']
        )
        practices.append(practice)

    return practices
```

---

## Template System

### 3-Tier Templates

**Tier 1: Simple Agent Template** (~200 lines)
```yaml
---
name: {{ agent_name }}
description: {{ agent_description }}
model: haiku
---

## Quick Reference

{{ quick_summary }}

## Implementation

### Component 1: {{ first_capability }}
{{ implementation_details }}

## Usage

{{ usage_examples }}
```

**Tier 2: Standard Agent Template** (200-500 lines)
```yaml
---
name: {{ agent_name }}
description: {{ agent_description }}
model: {{ model_selection }}
---

## Overview

{{ comprehensive_overview }}

## Architecture

### System Design
{{ architecture_diagram }}

### Components
{% for component in components %}
### {{ component.name }}
{{ component.details }}
{% endfor %}

## Implementation Patterns

{{ detailed_patterns }}

## Integration Points

{{ integration_guide }}

## Testing & Validation

{{ test_strategy }}
```

**Tier 3: Complex Agent Template** (500+ lines)
```yaml
---
name: {{ agent_name }}
description: {{ agent_description }}
model: sonnet
---

## Executive Summary
{{ strategic_overview }}

## Complete Architecture
{{ full_architecture }}

## Component Details
{% for system in major_systems %}
{{ system.full_documentation }}
{% endfor %}

## Advanced Patterns
{{ advanced_implementation_patterns }}

## Multi-Domain Integration
{{ orchestration_details }}

## Performance Considerations
{{ performance_optimization_guide }}

## Enterprise Features
{{ compliance_and_monitoring }}

## Deployment & Scaling
{{ production_deployment_guide }}
```

### Variable Categories

**15+ Variable Categories**:

1. **Agent Metadata** (5 vars)
   - `{{ agent_name }}` - Agent identifier
   - `{{ agent_description }}` - What agent does
   - `{{ version }}` - Semantic version
   - `{{ status }}` - Production/Beta/Experimental
   - `{{ author }}` - Skill/agent creator

2. **Capabilities** (5 vars)
   - `{{ primary_capability }}` - Main function
   - `{{ secondary_capabilities }}` - Additional features
   - `{{ integrations }}` - External systems
   - `{{ constraints }}` - Limitations
   - `{{ performance_targets }}` - SLA/benchmarks

3. **Architecture** (3 vars)
   - `{{ architecture_type }}` - Monolithic/Microservice/Orchestrated
   - `{{ dependency_diagram }}` - Component relationships
   - `{{ data_flow }}` - Information movement

4. **Implementation** (2 vars)
   - `{{ code_examples }}` - Practical usage
   - `{{ error_handling }}` - Exception strategies

---

## Validation Framework

### 4 Quality Gates

**Gate 1: YAML Syntax Validation**
```python
def validate_yaml_syntax(self, agent_markdown: str) -> ValidationResult:
    """Validate YAML frontmatter syntax."""
    try:
        # Extract frontmatter
        frontmatter = extract_frontmatter(agent_markdown)
        # Parse YAML
        yaml.safe_load(frontmatter)
        return ValidationResult(passed=True)
    except yaml.YAMLError as e:
        return ValidationResult(passed=False, errors=[str(e)])
```

**Gate 2: Required Sections**
```python
REQUIRED_SECTIONS = [
    'Overview',
    'What It Does',
    'When to Use',
    'Best Practices',
    'Three-Level Learning Path',
    'Integration Points'
]

def validate_structure(self, markdown: str) -> ValidationResult:
    """Validate required sections exist."""
    missing = []
    for section in REQUIRED_SECTIONS:
        if f"## {section}" not in markdown:
            missing.append(section)

    if missing:
        return ValidationResult(
            passed=False,
            errors=[f"Missing sections: {', '.join(missing)}"]
        )
    return ValidationResult(passed=True)
```

**Gate 3: Content Quality**
```python
def validate_content_quality(self, markdown: str) -> ValidationResult:
    """Check content quality metrics."""

    checks = {
        'code_examples': self._count_code_examples(markdown) >= 5,
        'documentation_complete': self._check_documentation_depth(markdown),
        'readability': self._check_readability_score(markdown) >= 7,
        'no_broken_links': self._check_internal_links(markdown),
    }

    if all(checks.values()):
        return ValidationResult(passed=True)
    else:
        failed = [k for k, v in checks.items() if not v]
        return ValidationResult(passed=False, errors=failed)
```

**Gate 4: TRUST 5 + Claude Code Compliance**
```python
def validate_trust_compliance(self, markdown: str) -> ValidationResult:
    """Validate TRUST 5 + Claude Code principles."""

    trust_checks = {
        'test_driven': self._has_test_examples(markdown),
        'readable': self._check_code_clarity(markdown),
        'unified': self._check_consistent_patterns(markdown),
        'secured': self._check_security_practices(markdown),
        'evaluated': self._check_metrics_included(markdown),
    }

    cc_checks = {
        'skill_references': self._has_skill_references(markdown),
        'tool_usage': self._validates_tool_usage(markdown),
        'delegation_pattern': self._follows_delegation_pattern(markdown),
    }

    all_passed = all(trust_checks.values()) and all(cc_checks.values())

    return ValidationResult(
        passed=all_passed,
        trust_scores=trust_checks,
        claude_code_scores=cc_checks
    )
```

### Test Cases

**5 Core Test Scenarios**:
1. Simple authentication agent generation
2. Complex microservice orchestrator generation
3. Multi-domain agent creation (backend + security)
4. Edge case: very simple vs. very complex requirements
5. End-to-end generation with validation

**3 Edge Cases**:
1. Missing or ambiguous requirements
2. Conflicting capability requirements
3. Unsupported framework combination

---

## Advanced Features

### Semantic Versioning

```python
class AgentVersion:
    major: int  # Breaking changes
    minor: int  # New features (backward compatible)
    patch: int  # Bug fixes

    def __str__(self):
        return f"{self.major}.{self.minor}.{self.patch}"

    def is_compatible_with(self, other: 'AgentVersion') -> bool:
        """Check compatibility with another version."""
        return self.major == other.major
```

### Multi-Domain Support

```python
def generate_multi_domain_agent(
    self,
    primary_domain: str,
    secondary_domains: List[str],
    requirement: RequirementAnalysis
) -> AgentMarkdown:
    """Generate agent for 2-3 domains."""

    # Research each domain
    all_research = {}
    for domain in [primary_domain] + secondary_domains:
        all_research[domain] = self.research_engine.research_domain(domain)

    # Generate orchestrator template
    template = self.template_system.select_template(complexity=8)

    # Integrate domain-specific practices
    variables = {
        'primary_domain': primary_domain,
        'domain_practices': all_research,
        'orchestration_pattern': self._select_orchestration_pattern(
            secondary_domains
        ),
    }

    return self.template_system.generate_agent(template, variables)
```

### Performance Optimizer

```python
def optimize_agent_performance(self, agent: AgentMarkdown) -> OptimizationPlan:
    """Analyze agent for performance improvements."""

    analysis = {
        'token_efficiency': self._analyze_token_usage(agent),
        'latency_optimization': self._find_latency_improvements(agent),
        'resource_efficiency': self._check_resource_usage(agent),
        'cache_opportunities': self._identify_caching_patterns(agent),
    }

    recommendations = []
    for aspect, findings in analysis.items():
        if findings.issues:
            recommendations.extend(findings.recommendations)

    return OptimizationPlan(
        current_score=self._calculate_score(analysis),
        recommendations=recommendations,
        estimated_improvement='15-30%'
    )
```

### Enterprise Compliance

```python
class ComplianceChecker:
    STANDARDS = {
        'SOC2': [...],  # SOC 2 Type II requirements
        'GDPR': [...],  # GDPR compliance checklist
        'HIPAA': [...], # HIPAA security requirements
        'OWASP': [...], # OWASP Top 10 compliance
    }

    def check_compliance(self, agent: AgentMarkdown, standards: List[str]) -> ComplianceReport:
        """Check agent against specified standards."""
        report = ComplianceReport()

        for standard in standards:
            requirements = self.STANDARDS[standard]
            checks = self._check_requirements(agent, requirements)
            report.add_standard_result(standard, checks)

        return report
```

---

## Integration Patterns

### With agent-factory Agent

```yaml
---
name: agent-factory
model: sonnet
skills:
  - moai-core-agent-factory
  - moai-context7-mcp-integration
---

## Execution Flow

1. **Parse**: Load this skill (Intelligence Engine, Research Engine, Template System, Validation)
2. **Analyze**: Use Intelligence Engine for requirement analysis
3. **Research**: Use Research Engine with Context7 for best practices
4. **Generate**: Use Template System to create agent markdown
5. **Validate**: Use Validation Framework for quality assurance
6. **Return**: Production-ready agent markdown
```

### Delegation Protocol

```
User Requirement
    ↓
agent-factory loads moai-core-agent-factory
    ├─ Intelligence Engine: Analyze + score
    ├─ Research Engine: Context7 lookup
    ├─ Template System: Generate markdown
    ├─ Validation Framework: Quality check
    └─ Optimization: Performance suggestions
    ↓
Returns: AgentMarkdown + ValidationReport
```

---

## Error Handling

```python
class AgentGenerationError(Exception):
    """Base exception for agent generation errors."""
    pass

class InsufficientRequirementError(AgentGenerationError):
    """Raised when requirement is too vague."""
    pass

class TemplateNotFoundError(AgentGenerationError):
    """Raised when no matching template exists."""
    pass

class ValidationFailedError(AgentGenerationError):
    """Raised when validation gates fail."""
    pass

def generate_with_fallback(self, requirement: str) -> AgentMarkdown:
    """Generate agent with error recovery."""
    try:
        return self.generate_agent(requirement)
    except InsufficientRequirementError:
        # Return simpler template
        return self.generate_simple_agent(requirement)
    except ValidationFailedError as e:
        # Log issues and return with warnings
        return self.generate_with_warnings(requirement, e.failures)
```

---

**Last Updated**: 2025-11-22
**Lines**: ~600 (advanced patterns)
**Status**: Production Ready
