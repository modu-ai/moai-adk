# Configuration Guide

## Overview

Comprehensive configuration options and best practices.

## Basic Configuration

### YAML Configuration

```yaml
# config.yaml
skill:
 name: "skill-name"
 version: "1.0.0"
 options:
 mode: "production"
 timeout: 60
 retries: 3
```

### Environment Variables

```bash
export SKILL_MODE=production
export SKILL_TIMEOUT=60
export SKILL_RETRIES=3
```

## Advanced Configuration

### Custom Settings

```python
from moai_adk import Config

config = Config.from_file("config.yaml")
config.set("custom_option", "value")
config.save()
```

### Profile-Based Configuration

```yaml
# config.yaml
profiles:
 development:
 mode: "debug"
 timeout: 300
 production:
 mode: "strict"
 timeout: 30
```

---
Last Updated: 2025-11-23
Status: Production Ready
