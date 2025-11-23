# Execution Guide

## Overview

Step-by-step guide for executing documentation validation workflows.

## Quick Start

### Prerequisites

```bash
pip install -r requirements.txt
python --version  # 3.11+
```

### Basic Execution

```bash
# Run all validations
python validate_docs.py --all

# Run specific checks
python validate_docs.py --check links
python validate_docs.py --check structure
```

## Advanced Usage

### Custom Configuration

```yaml
# validation-config.yaml
checks:
  links:
    enabled: true
    external: true
  structure:
    enabled: true
    strict: false
```

### CI/CD Integration

```yaml
# .github/workflows/docs-validation.yml
name: Docs Validation
on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run validation
        run: python validate_docs.py --all
```

---
**Last Updated**: 2025-11-23
**Status**: Production Ready
