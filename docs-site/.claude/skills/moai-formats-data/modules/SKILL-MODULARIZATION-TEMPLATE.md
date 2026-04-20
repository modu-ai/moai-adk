# Skill Modularization Template

> Purpose: Template for restructuring oversized skills (>500 lines) into compliant Claude Code format
> Target: Progressive disclosure with modules/ directory structure
> Compliance: <500 lines main SKILL.md, detailed modules for implementation

## When to Use This Template

Apply to skills that exceed 500 lines:
- Current skill has >500 lines (violates Claude Code standards)
- Contains multiple distinct technical areas
- Has extensive implementation examples
- Requires progressive disclosure for optimal user experience

Typical candidates:
- Domain skills with multiple subdomains
- Implementation-heavy skills with extensive code examples
- Skills covering multiple technologies or approaches

## Modularization Process

### 1. Analysis Phase

Identify distinct technical areas:
```python
# Example for moai-formats-data:
technical_areas = [
 "TOON encoding implementation",
 "JSON/YAML optimization",
 "Data validation system",
 "Caching and performance"
]
```

Map content structure:
- Quick Reference content (keep in main SKILL.md)
- Implementation basics (keep in main SKILL.md)
- Advanced features (move to modules)
- Extended code examples (move to modules)

### 2. Main SKILL.md Structure

YAML Header (required):
```yaml
---
name: skill-name
description: Clear, concise description (under 80 chars)
version: 1.0.0
category: library|domain|integration|workflow
tags: [3-7 relevant tags]
updated: YYYY-MM-DD
status: active|development|deprecated
author: MoAI-ADK Team
---
```

Progressive Disclosure Sections:

1. Quick Reference (30 seconds) - Essential overview
 - Core capabilities (3-6 bullet points with emojis)
 - When to use (3-5 bullet points)
 - Quick start code snippet (5-10 lines max)

2. Implementation Guide (5 minutes) - Practical basics
 - Core concepts (brief explanations)
 - Basic implementation examples
 - Common use cases with short examples

3. Advanced Features (10+ minutes) - Extended usage
 - Advanced patterns and techniques
 - Integration examples
 - Performance considerations

Cross-references to modules:
```markdown
## Module References

Core Implementation Modules:
- [`modules/module-name.md`](./modules/module-name.md) - Brief description
```

### 3. Module Structure

Each module follows this template:

```markdown
# Module Title

> Module: Core area description
> Complexity: Basic|Intermediate|Advanced
> Time: X+ minutes
> Dependencies: List of required libraries

## Core Implementation

Complete code implementation with:
- Full class/function definitions
- Comprehensive examples
- Error handling
- Performance characteristics

## Advanced Features

Extended functionality:
- Custom extensions
- Integration patterns
- Performance optimization
- Edge case handling

## Best Practices

Guidelines for production use:
- Performance tips
- Security considerations
- Maintenance recommendations

---

Module: `modules/module-name.md`
Related: [Other Module](./other-module.md) | [Related Module](./related-module.md)
```

### 4. Content Distribution Rules

Keep in main SKILL.md:
- Quick Reference section (30 seconds)
- Basic Implementation (5 minutes)
- Essential code examples (under 20 lines each)
- Overview and integration patterns
- Module cross-references

Move to modules:
- Complete implementation classes (>50 lines)
- Extended examples and use cases
- Advanced features and patterns
- Performance optimization details
- Complex integration examples

### 5. Validation Checklist

Main SKILL.md compliance:
- [ ] Under 500 lines total
- [ ] Complete YAML metadata
- [ ] Progressive disclosure sections
- [ ] Quick Reference (30s) present and concise
- [ ] Implementation Guide (5min) covers basics
- [ ] Module cross-references use forward slashes
- [ ] No duplicate content with modules

Module structure compliance:
- [ ] Clear module headers with metadata
- [ ] Focused on single technical area
- [ ] Complete implementation examples
- [ ] Cross-references to related modules
- [ ] Consistent formatting across modules

Content quality:
- [ ] No information lost during modularization
- [ ] Clear navigation between main and modules
- [ ] Proper forward slashes in all paths
- [ ] Consistent code formatting
- [ ] Comprehensive examples in modules

## Template Application Example

### Before (832 lines - Non-compliant)
```markdown
# Skill Name
# ... 832 lines of mixed content
# Basic overview mixed with advanced implementation
# No clear progression from simple to complex
```

### After (Compliant Structure)

Main SKILL.md (490 lines):
```markdown
---
name: moai-formats-data
# ... YAML metadata
---

# Data Format Specialist

## Quick Reference (30 seconds)
Quick overview with 3-6 key capabilities
When to use bullets
Quick start example (5-10 lines)

## Implementation Guide (5 minutes)
Core concepts
Basic implementation examples
Common use cases

## Advanced Features (10+ minutes)
Advanced patterns
Integration examples

## Module References
- [`modules/toon-encoding.md`](./modules/toon-encoding.md) - TOON implementation
- [`modules/json-optimization.md`](./modules/json-optimization.md) - JSON optimization
# ... other module references
```

modules/TOON-encoding.md (200+ lines):
```markdown
# TOON Encoding Implementation

> Module: Core TOON implementation
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: Python 3.8+, typing, datetime

## Core Implementation
[Complete TOON implementation code]

## Advanced Features
[Custom type handlers, streaming, etc.]

---

Module: `modules/toon-encoding.md`
Related: [JSON Optimization](./json-optimization.md)
```

## Benefits of Modularization

For Users:
- Progressive disclosure respects their time investment
- Clear path from basic to advanced usage
- Focused modules for specific learning goals
- Better navigation and information architecture

For Maintainers:
- Easier to update specific areas
- Reduced cognitive load when reviewing changes
- Better code organization and reuse
- Simplified testing and validation

For Compliance:
- Meets Claude Code <500 line requirement
- Follows progressive disclosure best practices
- Maintains complete functionality
- Improves token efficiency and loading performance

## Automated Validation

Use this script to validate compliance:
```python
def validate_skill_modularization(skill_path: str) -> dict:
 """Validate skill meets Claude Code modularization standards."""
 main_file = f"{skill_path}/SKILL.md"
 modules_dir = f"{skill_path}/modules"

 # Check main file length
 with open(main_file, 'r') as f:
 main_lines = len(f.readlines())

 # Check for required sections
 with open(main_file, 'r') as f:
 content = f.read()
 has_quick_ref = "Quick Reference" in content
 has_implementation = "Implementation Guide" in content
 has_modules = "Module References" in content

 return {
 "main_file_lines": main_lines,
 "under_500_lines": main_lines < 500,
 "has_required_sections": has_quick_ref and has_modules,
 "compliant": main_lines < 500 and has_quick_ref and has_modules
 }
```

---

Template Version: 1.0.0
Last Updated: 2025-11-30
Purpose: Claude Code skill compliance and progressive disclosure
