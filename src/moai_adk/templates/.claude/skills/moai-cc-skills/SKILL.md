---
name: "Creating and Optimizing Claude Code Skills"
description: "Design reusable knowledge capsules with progressive disclosure (metadata → content → resources). Apply freedom levels (high/medium/low), create examples, validate YAML. Use when building domain-specific guidance or automating recurring patterns."
allowed-tools: "Read, Write, Edit, Glob, Bash"
---

# Creating and Optimizing Claude Code Skills

Skills are modular, reusable knowledge capsules that load via Progressive Disclosure. They combine templates, patterns, and best practices without blocking user workflow.

## Skill File Structure

**Location**: `.claude/skills/skill-name/`

```
skill-name/
├── SKILL.md              # Main instructions (~400 lines)
├── reference.md          # Detailed reference (optional)
├── examples.md           # Real-world examples (optional)
├── CHECKLIST.md          # Validation guide (optional)
├── scripts/              # Utility scripts
│   ├── validator.sh
│   └── formatter.py
└── templates/            # Reusable templates
    ├── config.json
    └── main.md
```

## YAML Frontmatter

```yaml
---
name: "Skill Name (Gerund + Domain)"
description: "[Capability]. Use when [trigger 1], [trigger 2], [trigger 3]."
allowed-tools: "Read, Write, Bash(python:*)"
---
```

### Field Rules

| Field | Max | Format | Example |
|-------|-----|--------|---------|
| `name` | 64 chars | Gerund + Domain | "Processing CSV Files with Python" |
| `description` | 1024 chars | 3rd person, 3+ triggers | "Extract data, validate schema, merge datasets. Use when processing CSV, Excel files, or data cleaning." |
| `allowed-tools` | — | Comma-separated minimal list | "Read, Bash(python:*), Bash(grep:*)" |

## Freedom Level Framework

| Freedom | % | Use Case | Content Style |
|---------|---|----------|----------------|
| **High** | 20-30% | Flexible, creative | Principles, trade-offs, decision trees |
| **Medium** | 50-60% | Standard patterns | Pseudocode, flowcharts, templates |
| **Low** | 10-20% | Deterministic, error-prone | Scripts, error handling, validation |

### High-Freedom: Architecture Design

```markdown
## Architecture Trade-offs

| Pattern | Pros | Cons | When |
|---------|------|------|------|
| Monolith | Simple, unified | Scales poorly | MVP, <10 people |
| Microservices | Scalable, independent | Complex, distributed tracing | 10+ teams |
| Serverless | Zero ops, elastic | Cold starts, vendor lock-in | Event-driven |

Choose based on:
1. **Team size**: Monolith if <5 devs
2. **Scale**: Microservices if >10M requests/month
3. **Budget**: Serverless if unpredictable workload
```

### Medium-Freedom: Data Validation Pattern

```pseudocode
## Validation Workflow

1. Load data source (CSV, JSON, API)
2. For each record:
   a. Check required fields exist
   b. Validate data types
   c. Verify constraints (range, format)
   d. Check references (foreign keys)
3. Collect errors with row numbers
4. Report summary + fix suggestions
5. Proceed only if all errors fixed

## Example: CSV Validation
   - Input: data.csv
   - Schema: validators/schema.json
   - Output: validation_report.json
```

### Low-Freedom: Security Validator Script

```python
#!/usr/bin/env python3
"""Validate code for OWASP Top 10 risks"""
import re
import sys

PATTERNS = [
    (r"eval\(", "OWASP-A1: Code injection risk"),
    (r"SELECT.*\$", "OWASP-A3: SQL injection risk"),
    (r"os\.system\(", "OWASP-A6: Command injection risk"),
]

def validate_file(filepath):
    with open(filepath) as f:
        content = f.read()

    issues = []
    for pattern, risk in PATTERNS:
        if re.search(pattern, content):
            issues.append(risk)

    return issues

if __name__ == "__main__":
    for filepath in sys.argv[1:]:
        issues = validate_file(filepath)
        if issues:
            print(f"⚠️  {filepath}:")
            for issue in issues:
                print(f"  • {issue}")
```

## Progressive Disclosure Pattern

**Level 1**: Metadata (always loaded)
```yaml
name: "Database Query Optimization"
description: "Profile slow queries, analyze execution plans, suggest indexes..."
allowed-tools: "Read, Bash(sqlite3:*)"
```

**Level 2**: Main content (on demand)
- Overview + quick start
- 3-4 core patterns
- Decision tree

**Level 3**: Resources (when referenced)
- `reference.md` — Detailed specs
- `examples.md` — Real-world scenarios
- `scripts/` — Utility tools

## Skill Discovery Keywords

Embed in description for auto-activation:

✅ **DO** include:
- Problem domain: "CSV", "SQL", "React", "API"
- Operation: "parsing", "validation", "optimization", "debugging"
- Tech stack: "Python", "TypeScript", "Go"

Example description:
```
"Parse CSV files, validate data schemas, clean datasets.
Use when processing CSV, Excel, JSON data or when user mentions
data cleaning, validation, or transformation."
```

## Medium-Freedom: Example Skill Template

```yaml
---
name: "Testing React Components with Vitest"
description: "Write unit tests for React components, mock hooks, test async behavior. Use when testing React components, setting up Vitest, or debugging test failures."
allowed-tools: "Read, Write, Bash(npm:*), Bash(npm run:*)"
---

# Testing React Components with Vitest

## Quick Start

1. Install: `npm install -D vitest @testing-library/react`
2. Create test: `src/Component.test.tsx`
3. Run: `npm run test`

## Core Patterns

### Pattern 1: Basic Component Test
```typescript
import { render, screen } from '@testing-library/react'
import { Button } from './Button'

it('renders button with text', () => {
  render(<Button>Click me</Button>)
  expect(screen.getByText('Click me')).toBeInTheDocument()
})
```

### Pattern 2: Mock Hooks
```typescript
import { useUserStore } from './store'

vi.mock('./store', () => ({
  useUserStore: vi.fn(() => ({ user: null }))
}))
```

## Examples

See [examples.md](examples.md) for integration with Redux, React Query, form testing.
```

## Skill Validation Checklist

- [ ] `name` is specific (not "React Helper")
- [ ] `description` includes 3+ discovery keywords
- [ ] Main content ≤ 500 lines
- [ ] At least 2 concrete examples
- [ ] Progressive Disclosure applied (metadata → content → resources)
- [ ] Freedom levels mixed appropriately
- [ ] No hardcoded paths or time-sensitive data
- [ ] YAML frontmatter is valid
- [ ] Paths use forward slashes (`/` not `\`)

## Skill Optimization Patterns

### Anti-pattern 1: Too Generic
```yaml
❌ name: "Python Helper"
✅ name: "Testing Python Code with pytest"
```

### Anti-pattern 2: No Examples
```yaml
❌ Only principles, no code samples
✅ Each concept has working example
```

### Anti-pattern 3: Nested Files
```yaml
❌ docs/api/v2/reference/security.md
✅ reference.md (one level)
```

### Anti-pattern 4: Absolute Paths
```yaml
❌ /Users/name/project/config.json
✅ ./config.json or ${PROJECT_DIR}/config.json
```

## Best Practices

✅ **DO**:
- Keep < 500 lines per SKILL.md
- Use relative paths with forward slashes
- Provide working code examples
- Design for reusability across projects
- Include both success and error paths

❌ **DON'T**:
- Mix too many domains (split into separate Skills)
- Explain fundamentals (trust Claude already knows)
- Create nested directory structures
- Include time-sensitive info (dates, versions)
- Use Windows-style paths

## Skill Lifecycle

1. **Create**: Design persona, scope, freedom levels
2. **Implement**: Write SKILL.md + examples
3. **Test**: Verify with Haiku, Sonnet, Opus
4. **Deploy**: Add to `.claude/skills/`
5. **Iterate**: Get feedback, update as needed

---

**Reference**: Claude Code Skills official documentation
**Version**: 1.0.0
