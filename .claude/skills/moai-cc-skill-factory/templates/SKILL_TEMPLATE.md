---
name: "[Action Verb] [Domain]"
description: "[Capability 1], [Capability 2], or [Capability 3]. Use when [trigger 1], [trigger 2], or [trigger 3]."
allowed-tools: "Tool1, Tool2, Tool3"
---

# [Skill Name]

> **Quick Reference**: See [reference.md](reference.md) for detailed reference, [examples.md](examples.md) for real-world scenarios.

---

## What Is This Skill?

[One-sentence summary of what the Skill provides]

---

## Quick Start

[Fastest path to getting started. 5-10 lines maximum]

```
Code example showing immediate usage
```

---

## Overview

### Problem Domain

Describe the problem this Skill solves.

### Scope

What's included:
- Item 1
- Item 2
- Item 3

What's NOT included:
- Out of scope 1
- Out of scope 2

### Key Concepts

1. **Concept 1**: Brief explanation
2. **Concept 2**: Brief explanation
3. **Concept 3**: Brief explanation

---

## Framework: High Freedom (Principles)

> Use for conceptual guidance, decision frameworks, and principles

### Core Principles

1. **Principle 1**
   - Rationale
   - When to apply

2. **Principle 2**
   - Rationale
   - When to apply

3. **Principle 3**
   - Rationale
   - When to apply

### Decision Framework

```
Question 1: [Ask]
├─ If YES: [Recommendation 1]
└─ If NO: [Recommendation 2]

Question 2: [Ask]
├─ If YES: [Recommendation 1]
└─ If NO: [Recommendation 2]
```

### Trade-off Analysis

| Option | Pros | Cons | When to Use |
|--------|------|------|-------------|
| Approach 1 | Pro 1 | Con 1 | Scenario 1 |
| Approach 2 | Pro 2 | Con 2 | Scenario 2 |
| Approach 3 | Pro 3 | Con 3 | Scenario 3 |

---

## Framework: Medium Freedom (Patterns)

> Use for standard patterns, workflows, and pseudocode

### Pattern 1: [Pattern Name]

**When to use**: [Describe scenario]

```pseudocode
1. Step 1
   ├─ Substep 1
   └─ Substep 2

2. Step 2
   ├─ Substep 1
   └─ Substep 2

3. Step 3
```

**Example**:
```python
# Concrete implementation example
code_example_here()
```

### Pattern 2: [Pattern Name]

**When to use**: [Describe scenario]

```pseudocode
Flowchart or pseudocode here
```

**Example**:
```python
# Concrete implementation example
code_example_here()
```

---

## Framework: Low Freedom (Scripts)

> Use for deterministic, error-prone operations

### Script 1: [Script Purpose]

See `scripts/script-name.sh` for the complete implementation.

```bash
#!/bin/bash
set -euo pipefail

# This script does [what]

# Step 1: Validate inputs
if [[ -z "$VAR" ]]; then
  echo "ERROR: Variable required" >&2
  exit 1
fi

# Step 2: Execute operation
operation_here

# Step 3: Report success
echo "✓ Operation completed successfully"
exit 0
```

**Usage**:
```bash
./scripts/script-name.sh [args]
```

---

## Common Use Cases

### Use Case 1: [Scenario]

Description of the scenario.

**Recommended approach**: [Pattern/Principle reference]

**Example**: See [examples.md](examples.md) Example 1

### Use Case 2: [Scenario]

Description of the scenario.

**Recommended approach**: [Pattern/Principle reference]

**Example**: See [examples.md](examples.md) Example 2

---

## Best Practices

1. **Practice 1**: Recommendation with rationale
2. **Practice 2**: Recommendation with rationale
3. **Practice 3**: Recommendation with rationale

---

## Anti-Patterns to Avoid

❌ **Don't do this**: Explanation of what not to do and why

✅ **Do this instead**: Correct approach

---

## FAQ

### Q: [Frequently asked question 1]

A: [Answer with reference to relevant section]

### Q: [Frequently asked question 2]

A: [Answer with reference to relevant section]

### Q: [Frequently asked question 3]

A: [Answer with reference to relevant section]

---

## Related Resources

- [reference.md](reference.md) — Detailed reference documentation
- [examples.md](examples.md) — Real-world usage examples
- `scripts/` — Automation utilities
- `templates/` — Reusable templates

---

## Glossary

- **Term 1**: Definition
- **Term 2**: Definition
- **Term 3**: Definition

---

## Version & Status

**Version**: 0.1.0
**Status**: Draft
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Code Skills
