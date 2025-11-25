# Examples

## Overview

Real-world examples and use cases.

## Basic Examples

### Example 1: Simple Usage

```python
# Initialize
from moai_adk import Skill

skill = Skill("skill-name")

# Execute
result = skill.run({
    "input": "test data"
})

print(result.output)
```

### Example 2: Advanced Usage

```python
# With configuration
result = skill.run(
    data={"complex": "data"},
    config={
        "mode": "advanced",
        "optimization": true
    }
)

# Process results
for item in result.items:
    process(item)
```

## Production Examples

### Example 3: CI/CD Integration

```yaml
# .github/workflows/skill-execution.yml
name: Run Skill
on: [push]

jobs:
  execute:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run skill
        run: |
          python -m moai_adk.skills.run             --skill skill-name             --input data.json
```

---
**Last Updated**: 2025-11-23
**Status**: Production Ready
