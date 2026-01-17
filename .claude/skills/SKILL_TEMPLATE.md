---
# Level 1: Core Metadata (Always Loaded)
# This section (~100 tokens) is loaded during agent initialization
name: "skill-name"
description: "Brief one-line description of what this skill does"
version: "1.0.0"
category: "domain"  # foundation, lang, platform, library, workflow, domain
modularized: true
user-invocable: true  # Can users invoke this skill directly?

# Progressive Disclosure Configuration
progressive_disclosure:
  enabled: true  # Enable 3-level loading for this skill
  level1_tokens: ~100  # Estimated tokens for metadata only
  level2_tokens: ~5000  # Estimated tokens for full body

# Trigger Conditions for Level 2 Loading
triggers:
  keywords: ["keyword1", "keyword2", "keyword3"]  # Load full skill when these appear
  phases: ["plan", "run", "sync"]  # Load during specific phases
  agents: ["agent-name", "another-agent"]  # Load for specific agents
  languages: ["python", "typescript"]  # Load for specific languages

# Dependencies (loaded at Level 1 if declared)
requires: []  # Skills that must be loaded before this one
optional_requires: []  # Skills that can enhance this skill

# Allowed Tools
allowed-tools:
  - Read
  - Write
  - Edit

# Skill Metadata
tags: ["tag1", "tag2"]
updated: 2026-01-16
status: "active"  # active, experimental, deprecated
---

# Level 2: Skill Body (Conditional Load)

<!-- This section (~5K tokens) is loaded ONLY when triggers match -->

## Quick Reference

**What is [skill-name]?**

One-sentence answer explaining the core purpose.

**Key Benefits:**
- Benefit 1
- Benefit 2
- Benefit 3

**When to Use:**
- Use case 1
- Use case 2

**Quick Links:**
- Implementation: #implementation-guide
- Examples: examples.md
- Reference: reference.md

---

## Implementation Guide

### Core Concepts

Brief explanation of the main concepts (1-2 paragraphs).

### Usage Pattern

Basic usage example with code snippet.

### Integration Points

How this skill integrates with other skills and agents.

---

## Advanced Topics

### Performance Considerations

Tips for optimal performance.

### Edge Cases

Handling special cases and error scenarios.

### Best Practices

Recommended patterns and anti-patterns.

---

## Works Well With

**Related Skills:**
- skill-one: When to use together
- skill-two: Complementary use cases

**Agents:**
- agent-name: How this agent uses this skill

**Commands:**
- /command: Command that invokes this skill

---

# Level 3: Bundled Files (On-Demand Load)

<!-- These files are loaded by Claude when needed, unlimited size -->

## Module References

Extended documentation in modular files:

- **modules/patterns.md**: Detailed design patterns
- **modules/examples.md**: Comprehensive examples
- **modules/reference.md**: API reference
- **modules/troubleshooting.md**: Common issues and solutions

## Examples

Working code samples in **examples.md**:
- Example 1: Basic usage
- Example 2: Advanced usage
- Example 3: Integration pattern

## Reference

External resources in **reference.md**:
- Official documentation
- Community resources
- Related tools

## Scripts

Utility scripts in **scripts/** directory:
- script-name.py: Script description
- script-name.sh: Script description

---

# Progressive Disclosure Levels Summary

| Level | What | When | Token Cost |
|-------|------|------|------------|
| 1 | YAML Metadata only | Agent initialization | ~100 tokens |
| 2 | SKILL.md Body | Trigger keywords match | ~5K tokens |
| 3+ | Bundled files | Claude decides | Unlimited |

---

# Backward Compatibility Note

For agents that don't support Progressive Disclosure yet, the entire SKILL.md (Levels 1 + 2) will be loaded at initialization. This ensures compatibility while enabling optimization for Progressive Disclosure-aware agents.
