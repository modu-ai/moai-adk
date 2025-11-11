---
title: Code Style Guide
description: MoAI-ADK Python, Markdown, YAML code style standards
status: stable
---

# Code Style Guide

This guide explains the code style standards for MoAI-ADK. All contributors must follow this guide.

## Python Code Style

### Standard Compliance

- **Standard**: PEP 8 + Black formatting
- **Linter**: Ruff + mypy (type checking)
- **Formatter**: Black (auto-formatting)

### File Structure

```python
"""
Module description.

This module... detailed description
"""

# Standard library
import os
import sys
from pathlib import Path
from typing import Optional

# Third-party libraries
import pytest
from pydantic import BaseModel

# Local libraries
from moai_adk.core import Agent
from moai_adk.utils import logger


class MyClass:
    """Class description."""

    def method(self) -> None:
        """Method description."""
        pass
```

### Naming Rules

| Item | Rule | Example |
|------|------|------|
| **Class** | PascalCase | `class MyAgent:` |
| **Function/Method** | snake_case | `def get_config():` |
| **Constant** | UPPER_SNAKE_CASE | `DEFAULT_TIMEOUT = 30` |
| **Private** | _leading_underscore | `def _internal_method():` |
| **Module** | snake_case | `my_module.py` |

### Type Hints

```python
from typing import Optional, List, Dict, Union

def process_data(
    items: List[str],
    config: Optional[Dict[str, int]] = None,
) -> bool:
    """
    Data processing function.

    Args:
        items: List of items to process
        config: Optional configuration dictionary

    Returns:
        Processing success status

    Raises:
        ValueError: Invalid input
    """
    if not items:
        raise ValueError("items cannot be empty")
    return True
```

### Comments and Docstrings

```python
def calculate_score(value: int) -> float:
    """
    Score calculation.

    This function calculates a normalized score based on input value.
    Range is between 0.0 and 1.0.

    Args:
        value: Input value to calculate (0-100)

    Returns:
        Normalized score (0.0-1.0)

    Examples:
        >>> calculate_score(50)
        0.5
    """
    # Range validation
    if not 0 <= value <= 100:
        raise ValueError(f"Value must be 0-100, got {value}")

    # Score calculation
    return value / 100.0
```

### Line Length and Formatting

```python
# Black default: 88 characters
# Automatically wraps if too long

def long_function_name(
    param1: str,
    param2: int,
    param3: Optional[Dict[str, Any]] = None,
) -> Tuple[str, int]:
    """Long function definition example."""
    pass
```

## Markdown Style

### File Structure

```markdown
---
title: Page Title
description: Page Description
status: stable
---

# H1 Title

All markdown files follow this structure.

## H2 Section

### H3 Subsection

Avoid deeper headings (don't use H4+).

### List Format

**Bullet points**:
- First item
- Second item
- Third item

**Numbered list**:
1. First step
2. Second step
3. Third step

### Emphasis

- **Bold** (important emphasis)
- *Italic* (term emphasis)
- ` ` (inline code)
```

### Code Blocks

````markdown
```python
# Python code
def hello():
    print("Hello, World!")
```

```bash
# Bash command
uv run pytest
```

```yaml
# YAML configuration
key: value
nested:
  item: value
```
````

### Tables

```markdown
| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Content A | Content B | Content C |
| Content D | Content E | Content F |
```

## YAML Style

### Configuration Files

```yaml
# Comments have space after #
key: value

# Nested structure uses 2-space indentation
parent:
  child: value
  list_item:
    - item1
    - item2

# Complex values expressed in multiple lines
description: |
  Multiple lines
  of text are
  expressed with pipe.
```

## Automated Style Checks

### Ruff (Linting)

```bash
# Style check
uv run ruff check src/

# Auto-fix
uv run ruff check --fix src/
```

### Black (Formatting)

```bash
# Check formatting
uv run black --check src/

# Auto-format
uv run black src/
```

### mypy (Type Checking)

```bash
# Type check
uv run mypy src/moai_adk
```

### Integrated Checks

```bash
# Run all checks
uv run ruff check src/
uv run black --check src/
uv run mypy src/moai_adk
```

## Pre-commit Configuration

`.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix]

  - repo: https://github.com/psf/black
    rev: 23.10.0
    hooks:
      - id: black

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.6.0
    hooks:
      - id: mypy
```

## Checklist

Check before PR submission:

- [ ] Python code formatted with Black
- [ ] Ruff linting passes (no errors)
- [ ] mypy type checking passes
- [ ] Markdown files have correct structure
- [ ] Comments and docstrings added to code
- [ ] Test code included (test coverage 87%+)

## Reference Materials

- [PEP 8](https://www.python.org/dev/peps/pep-0008/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [Black Code Style](https://black.readthedocs.io/)
- [CommonMark Spec](https://spec.commonmark.org/)

---

**Questions?** Ask style-related questions on GitHub Discussions!




