# MoAI Foundation Core - Modules Directory

Purpose: Extended documentation modules for moai-foundation-core skill, providing deep dives into each foundational principle.

Version: 1.0.0
Last Updated: 2025-11-25

---

## Module Overview

This directory contains 9 comprehensive modules covering MoAI-ADK's foundational principles and execution rules:

### 1. trust-5-framework.md (982 lines)
TRUST 5 Quality Framework - Automated quality gates ensuring code quality, security, maintainability, and traceability.

Quick Access:
- Test-first (≥85% coverage)
- Readable (≤10 complexity)
- Unified (consistent patterns)
- Secured (OWASP compliance)
- Trackable (clear commits)

Use Cases:
- Quality gate configuration
- CI/CD pipeline integration
- Pre-commit hook setup
- TRUST 5 validation framework

---

### 2. spec-first-ddd.md (866 lines)
Specification-Driven Development - EARS format requirements with ANALYZE-PRESERVE-IMPROVE DDD cycles.

Quick Access:
- SPEC generation (/moai:1-plan)
- EARS format patterns
- DDD implementation (/moai:2-run)
- Documentation sync (/moai:3-sync)

Use Cases:
- New feature development
- Requirement specification
- Test-driven implementation
- Documentation automation

---

### 3. delegation-patterns.md (757 lines)
Agent Orchestration - Task delegation strategies for specialized agents without direct execution.

Quick Access:
- Sequential delegation (dependencies)
- Parallel delegation (independent tasks)
- Conditional delegation (analysis-based routing)
- Context passing optimization

Use Cases:
- Complex workflow orchestration
- Multi-agent coordination
- Error handling and recovery
- Performance optimization

---

### 4. token-optimization.md (656 lines)
Budget Management - Efficient 200K token budget through strategic context management.

Quick Access:
- Phase-based allocation (SPEC 30K | DDD 180K | Docs 40K)
- /clear execution rules
- Selective file loading
- Model selection strategy

Use Cases:
- Token budget planning
- Context optimization
- Cost reduction (60-70% savings)
- Performance tuning

---

### 5. progressive-disclosure.md (576 lines)
Content Architecture - Three-tier knowledge delivery balancing value with depth.

Quick Access:
- Level 1: Quick Reference (30s, 1K tokens)
- Level 2: Implementation Guide (5min, 3K tokens)
- Level 3: Advanced Patterns (10+min, 5K tokens)
- 500-line SKILL.md limit enforcement

Use Cases:
- Skill content structuring
- Documentation architecture
- File splitting strategy
- Progressive loading

---

### 6. modular-system.md (588 lines)
File Organization - Scalable file structure for unlimited content depth.

Quick Access:
- Standard structure (SKILL.md + modules/ + examples.md + reference.md)
- File splitting strategy
- Cross-reference patterns
- Module discovery

Use Cases:
- Skill organization
- File structure validation
- Automated splitting
- Navigation generation

---

### 7. agents-reference.md (NEW)
Agent Catalog - Complete reference of MoAI-ADK's 26 specialized agents with 7-tier hierarchy.

Quick Access:
- `{domain}-{role}` naming convention
- 7 tiers: workflow, core, domain, mcp, factory, support, ai
- Agent selection criteria
- MCP Resume pattern (40-60% token savings)

Use Cases:
- Agent selection for Task() delegation
- Understanding agent hierarchy
- MCP integrator usage
- Historical agent merges

---

### 8. commands-reference.md (NEW)
Command Catalog - Complete reference for MoAI-ADK's 6 core commands in SPEC-First DDD workflow.

Quick Access:
- /moai:0-project (Project init)
- /moai:1-plan (SPEC generation)
- /moai:2-run (DDD implementation)
- /moai:3-sync (Documentation)
- /moai:9-feedback (Improvement)
- /moai:99-release (Deployment)

Use Cases:
- Command workflow execution
- /clear execution rules
- Token budget by command
- Git integration patterns

---

### 9. execution-rules.md (NEW)
Security & Constraints - Security policies, execution constraints, and Git workflow strategies.

Quick Access:
- Agent-First Mandate (Task() only)
- Security Sandbox (protected paths, forbidden commands)
- Permission System (RBAC 4 levels)
- Git Strategy 3-Mode System (Manual/Personal/Team)
- TRUST 5 Quality Gates
- Compliance (GDPR, CCPA, OWASP, SOC 2, ISO 27001)

Use Cases:
- Security constraint enforcement
- Git workflow configuration
- Compliance validation
- Error handling protocols

---

## Usage Patterns

### Loading Individual Modules

```python
# Load specific module
from pathlib import Path

skill_path = Path(".claude/skills/moai-foundation-core")
module_path = skill_path / "modules" / "trust-5-framework.md"

with open(module_path) as f:
 content = f.read()
```

### Progressive Loading

```python
# Load progressively based on user needs
class ModuleLoader:
 def load_quick_reference(self, module_name: str):
 """Load Quick Reference section only (~1K tokens)."""
 content = self.load_module(module_name)
 return self.extract_section(content, "Quick Reference")
 
 def load_implementation(self, module_name: str):
 """Load Implementation Guide (~3K tokens)."""
 content = self.load_module(module_name)
 return self.extract_section(content, "Implementation Guide")
 
 def load_advanced(self, module_name: str):
 """Load Advanced Patterns (~5K tokens)."""
 content = self.load_module(module_name)
 return self.extract_section(content, "Advanced Implementation")
```

### Searching Across Modules

```python
class ModuleSearch:
 def search_topic(self, query: str) -> list:
 """Search for topic across all modules."""
 modules_dir = Path(".claude/skills/moai-foundation-core/modules")
 results = []
 
 for module_file in modules_dir.glob("*.md"):
 with open(module_file) as f:
 content = f.read()
 
 if query.lower() in content.lower():
 results.append({
 "module": module_file.stem,
 "matches": content.lower().count(query.lower())
 })
 
 return sorted(results, key=lambda x: x["matches"], reverse=True)

# Usage
searcher = ModuleSearch()
results = searcher.search_topic("security")
# Results: trust-5-framework (high), delegation-patterns (medium), etc.
```

---

## Integration with SKILL.md

The main SKILL.md file (409 lines, within 500-line limit) provides:
- Quick overview of all 6 principles
- Entry points to each module
- Cross-references for deep dives
- Works Well With integration

Cross-Reference Pattern:
```markdown
<!-- In SKILL.md -->
### 1. TRUST 5 Framework - Quality Assurance System

Quick overview...

Detailed Reference: [TRUST 5 Framework Module](modules/trust-5-framework.md)
```

---

## Module Statistics

| Module | Lines | Topics Covered | Use Cases |
|--------|-------|----------------|-----------|
| trust-5-framework | 982 | Quality gates, CI/CD, validation | 4 |
| spec-first-ddd | 866 | SPEC, EARS, DDD, docs | 4 |
| delegation-patterns | 757 | Sequential, parallel, conditional | 4 |
| token-optimization | 656 | Budget, /clear, loading, models | 4 |
| progressive-disclosure | 576 | 3 levels, 500-line limit, splitting | 4 |
| modular-system | 588 | File structure, organization, discovery | 4 |
| agents-reference | ~400 | 26 agents, 7 tiers, MCP Resume | 4 |
| commands-reference | ~300 | 6 commands, workflow, /clear rules | 4 |
| execution-rules | ~700 | Security, Git, compliance, RBAC | 4 |
| Total | ~6,825 | 36 major topics | 36 use cases |

---

## Works Well With

Skills:
- moai-foundation-core (parent skill)
- moai-cc-skill-factory (skill creation)
- moai-core-agent-factory (agent creation)

Agents:
- skill-factory (module generation)
- docs-manager (documentation)
- quality-gate (validation)

Commands:
- /moai:1-plan (SPEC-First DDD)
- /moai:2-run (DDD implementation)
- /moai:3-sync (Documentation)
- /clear (Token optimization)

---

Maintained by: MoAI-ADK Team
Status: Production Ready
Next Review: As needed when foundation principles evolve
