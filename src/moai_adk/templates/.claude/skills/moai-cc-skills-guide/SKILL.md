---
name: moai-cc-skills-guide
description: Complete guide to Claude Code Skills creation, discovery mechanisms, best practices, and troubleshooting. Use when building new Skills or optimizing existing ones with official Claude Code standards.
version: 1.0.0
modularized: false
tags:
  - enterprise
  - skills
  - configuration
  - claude-code
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: cc, moai, guide, skills  


# Claude Code Skills Complete Guide

Comprehensive guide to creating, discovering, and deploying Agent Skills following official Claude Code standards.

## Quick Reference

**Skills** are directories containing `SKILL.md` + optional supporting files. They are **model-invoked** (auto-discovered) and provide focused capabilities.

```
skill-name/
├── SKILL.md (required)
├── reference.md (optional)
├── scripts/ (optional)
└── templates/ (optional)
```

## Implementation Guide

### Storage Locations (3 Sources)

**1. Personal Skills** (`~/.claude/skills/`)
- Individual workflows
- Machine-specific
- Not version-controlled
- Highest discovery priority in personal projects

**2. Project Skills** (`.claude/skills/`)
- Team-shared
- Version-controlled in git
- Available to all project members
- Recommended for teams

**3. Plugin Skills** (Bundled with plugins)
- Distributed via Claude Code plugins
- Broadest reach (entire organizations)
- Professionally packaged
- Recommended for public distribution

**Discovery Priority**: Project > Personal > Plugin (first match wins)

### Required Fields

**name** (mandatory)
- Lowercase, numbers, hyphens only
- Max 64 characters
- Kebab-case format
- Example: `safe-file-reader`, `python-testing-patterns`

**description** (mandatory)
- Max 1024 characters
- Include triggering scenarios
- Be specific (not vague)
- Example: "Read files safely without making changes. Use when you need read-only file access with Grep searches and Glob pattern matching."

### Optional Fields

**allowed-tools** (recommended)
- Comma-separated list
- Principle of least privilege
- Restricts Claude's tool access
- Example: `Read, Grep, Glob`

## Advanced Patterns

### Model-Invoked Discovery

Claude **automatically activates** relevant Skills based on:
- Task description matching skill purpose
- Specific terminology in the request
- Context and required capabilities

**Best Practice**: Write descriptions that include specific triggers
```
❌ Bad: "Helps with files"
✅ Good: "Read and search files without modifications. Use when you need read-only access with pattern matching."
```

### Progressive Disclosure

Supporting files load **only when needed**:
- Level 1: SKILL.md (always loaded)
- Level 2: reference.md (loaded if referenced)
- Level 3: scripts/, templates/ (loaded on demand)

This minimizes context window usage.

### Tool Restrictions (allowed-tools)

The `allowed-tools` field creates safe, focused skills:
```yaml
# Read-only skill
allowed-tools: Read, Grep, Glob

# Code modification skill
allowed-tools: Read, Write, Edit, Bash

# Analysis skill
allowed-tools: Read, Grep, WebFetch
```

**Benefits**:
- Security (prevent dangerous operations)
- Focus (limit to relevant tools)
- User trust (transparent capabilities)

### Best Practices (8 Core Principles)

**1. Single Responsibility**
- One clear capability per Skill
- Avoid kitchen-sink utilities
- Focus depth over breadth

**2. Specific Descriptions**
- Include triggering scenarios
- Use domain terminology
- Give concrete examples
- Avoid generic wording

**3. Distinct Terminology**
- Prevent confusion with similar Skills
- Use unique keywords
- Define scope clearly

**4. Team Testing & Versioning**
- Test with teammates before deployment
- Use semantic versioning (1.0.0, 1.1.0, 2.0.0)
- Document changes in version history
- Commit to git for tracking

**5. Progressive Disclosure Structure**
- Quick Reference (30 seconds)
- Implementation Guide (step-by-step)
- Advanced Patterns (expertise level)

**6. Practical Examples**
- Real-world code examples
- Working patterns
- Copy-paste ready
- Test before including

**7. Clear Tool Restrictions**
- Use `allowed-tools` when appropriate
- Document why tools are restricted
- Enable secure, focused workflows

**8. Plugin Distribution**
- Package for broad sharing
- Create plugin with Skill bundled
- Document setup requirements
- Provide version management

### Common Issues & Solutions

**Issue: Vague Descriptions**
- Problem: Skill not discovered when relevant
- Solution: Add specific triggering scenarios
```yaml
❌ description: "Helps with Python"
✅ description: "Analyze Python code for type errors, performance issues, and security vulnerabilities using static analysis patterns"
```

**Issue: Invalid YAML Syntax**
- Problem: Skill fails to load silently
- Solution: Validate YAML frontmatter
```yaml
# ✅ Correct
---
name: my-skill
description: Description
allowed-tools: Read, Write
---

# ❌ Wrong (tabs instead of spaces, missing quotes)
---
name: my-skill
	description: Description  # Tab indentation!
---
```

**Issue: Incorrect File Paths**
- Problem: Reference files not found
- Solution: Verify paths with `ls` command
```bash
# Verify structure
ls -la .claude/skills/my-skill/
# Should show: SKILL.md, reference.md, etc.
```

**Issue: Loading Errors**
- Problem: Unknown why Skill isn't loading
- Solution: Use Claude Code debug mode
```bash
claude --debug  # Shows loading errors for all Skills
```

**Issue: Skill Not Discovered**
- Problem: Claude doesn't activate the Skill
- Solution: Review description specificity
- Check for conflicting Skills with similar names
- Ensure unique, clear terminology

## Architecture Considerations

### When to Create a Skill

Create a Skill when:
- Capability reused across projects
- Multiple users need the same guidance
- Specialized domain expertise needed
- Clear, focused responsibility

### When NOT to Create a Skill

Don't create a Skill for:
- One-time temporary guidance
- Project-specific patterns
- Generic instructions (use docs instead)
- Duplicate functionality

### Skill Maturity Levels

**Level 1: Experimental** (v0.x.x)
- Early development
- Testing with small group
- Frequent changes expected
- Not for production

**Level 2: Stable** (v1.x.x)
- Tested extensively
- Ready for team use
- Version-controlled
- Backward compatible

**Level 3: Distributed** (v2.0+)
- Published via plugin
- Professional documentation
- Long-term support
- Multiple versions supported

---

**Version**: 1.0.0
**Updated**: 2025-11-22
**Reference**: https://code.claude.com/docs/en/skills
