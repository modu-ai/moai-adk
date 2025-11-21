---
name: moai-cc-skill-factory
description: Enterprise AI-powered skill creation factory with intelligent generation, optimization, and Context7 integration
allowed-tools: [Read, Write, Edit, Bash, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
---

## Quick Reference

**Enterprise Skill Creation Factory**

**What it does**: AI-powered orchestrator for creating, optimizing, and managing MoAI skills at enterprise scale with intelligent generation, Context7 integration, and automated quality assurance.

**Core Capabilities**:
- ✅ AI-driven skill discovery from user requirements
- ✅ Automated skill architecture design (5-section SKILL.md pattern)
- ✅ Context7-powered content generation (50+ languages, 200+ frameworks)
- ✅ Real-time skill optimization and performance tuning
- ✅ Batch skill generation for team standardization
- ✅ Quality gates: test coverage ≥90%, security compliance, OWASP validation
- ✅ Modular structure: SKILL.md + modules/ + examples.md + reference.md

**When to Use**:
- Creating new enterprise skills from scratch
- Batch-updating skill sets for consistency
- Optimizing existing skills for performance
- Generating skill templates for teams
- Implementing skill standardization across projects
- Scaling skill creation with AI automation

**Quick Start**:
```
1. Define skill requirements (domain, purpose, scope)
2. Initialize skill factory with Context7 patterns
3. Generate modular skill structure
4. Validate with quality gates
5. Deploy to production
```

---

## Implementation Guide

### Skill Generation Workflow (5 Phases)

**Phase 1: Discovery**
- Analyze requirement specifications
- Identify domain and use cases
- Extract key features and capabilities
- Context7 pattern matching

**Phase 2: Architecture**
- Design SKILL.md structure (7-section)
- Plan modules for deep content
- Determine example coverage (10+ examples)
- Create reference documentation

**Phase 3: Content Generation**
- Generate Quick Reference (1 line)
- Create What It Does section (3-5 lines)
- Document When to Use (3-5 scenarios)
- Build Key Features list (5-8 items)
- Add Works Well With (3-5 related skills)
- Define Core Concepts (3-5 concepts)
- Structure Best Practices (5 DO + 5 DON'T)

**Phase 4: Modularization**
- Extract advanced patterns → modules/
- Move detailed examples → examples.md
- Build comprehensive reference → reference.md
- Keep SKILL.md ≤500 lines

**Phase 5: Quality Assurance**
- Verify 7-section structure
- Check SKILL.md ≤500 lines
- Validate 10+ examples
- Ensure consistency with standards
- Test skill activation

### Skill Creation Template Pattern

```python
from dataclasses import dataclass
from typing import Optional

@dataclass
class SkillRequirements:
    """Input for skill creation factory."""
    name: str  # e.g., "moai-mcp-integration"
    description: str  # 1-2 sentences
    domain: str  # e.g., "backend", "frontend", "infrastructure"
    purpose: str  # What the skill enables
    key_features: list[str]  # 5-8 core capabilities
    related_skills: list[str]  # 3-5 complementary skills
    examples_count: int = 10  # Minimum examples needed

class SkillFactory:
    """Generate production-ready skills."""

    async def create_skill(self, requirements: SkillRequirements) -> dict:
        """Create complete modular skill."""
        # 1. Get Context7 patterns for domain
        patterns = await self.get_context7_patterns(requirements.domain)

        # 2. Generate SKILL.md (7 sections, ≤500 lines)
        skill_md = self.generate_skill_core(requirements, patterns)

        # 3. Create supporting files
        examples = self.generate_examples(requirements, patterns)
        reference = self.generate_reference(requirements, patterns)
        modules = self.generate_modules(requirements, patterns)

        # 4. Validate structure
        self.validate_skill_structure({
            'SKILL.md': skill_md,
            'examples.md': examples,
            'reference.md': reference,
            'modules': modules
        })

        return {
            'SKILL.md': skill_md,
            'examples.md': examples,
            'reference.md': reference,
            'modules': modules
        }
```

### Modular Structure Best Practices

**SKILL.md** (≤500 lines): Quick Reference + Core Information
- Section 1: Quick Reference (20 lines)
- Section 2: What It Does (10 lines)
- Section 3: When to Use (15 lines)
- Section 4: Key Features (20 lines)
- Section 5: Works Well With (15 lines)
- Section 6: Core Concepts (20 lines)
- Section 7: Best Practices (30 lines)

**modules/** (3-5 files): Deep Dives by Topic
- advanced-patterns.md (400+ lines)
- optimization.md (300+ lines)
- architecture.md (350+ lines)
- deployment.md (300+ lines)

**examples.md** (400+ lines): 10+ Real-World Patterns
- Complete, runnable examples
- Copy-paste ready code
- Multiple use cases per example

**reference.md** (300+ lines): Specifications & API Docs
- Protocol specifications
- API reference
- Configuration options
- Troubleshooting guide

---

## Best Practices

✅ **DO**:
- Keep SKILL.md focused on quick reference + essentials
- Move detailed content to modules/
- Include 10+ real-world examples with context
- Validate against quality gates (coverage ≥90%)
- Apply Context7 patterns for latest standards
- Test skill activation in Claude Code
- Maintain modular structure for maintainability

❌ **DON'T**:
- Exceed 500 lines in SKILL.md
- Repeat content across sections
- Include pseudo-code or non-functional examples
- Skip example explanations
- Ignore quality gates or testing
- Mix multiple topics in single module
- Leave sections incomplete or vague

---

## Works Well With

- `moai-context7-integration` - Real-time documentation patterns (50+ languages)
- `moai-cc-configuration` - Skill configuration management
- `moai-essentials-debug` - Skill debugging and troubleshooting
- `moai-foundation-trust` - Security validation and compliance
- `moai-core-code-reviewer` - Skill code quality review

---

## Core Concepts

1. **Modular Architecture**: SKILL.md + modules split reduces cognitive load
2. **7-Section Structure**: Standardized format enables rapid skill consumption
3. **Context7 Integration**: Access 200+ framework patterns for current best practices
4. **Quality Gates**: Enforce consistency (500-line limit, 90% test coverage, OWASP)
5. **Batch Generation**: Create multiple skills with consistent structure
6. **Progressive Disclosure**: Quick ref first, deep dives in modules for interested readers

---

## Changelog

- **v3.0.0** (2025-11-22): Complete redesign - modular architecture, 7-section structure, Context7 integration
- **v2.0.0** (2025-11-20): AI-powered generation framework with automation
- **v1.0.0** (2025-11-10): Initial skill factory

---

**End of Core Skill** | See `modules/`, `examples.md`, and `reference.md` for detailed patterns | Status: Production Ready

