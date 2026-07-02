---
title: Catalog System
weight: 80
draft: false
---

Optimizes project initialization with a 3-tier catalog manifest and slim init.

## Overview

The catalog system in MoAI-ADK v2.15+ manages every agent, skill, plugin, and rule
through a **3-tier manifest**. `moai init --slim` deploys only the minimal templates
a project actually needs, shortening initialization time.

## 3-Tier Manifest

| Tier | Description | Deployment criteria |
|------|------|----------|
| **Tier 1 (Core)** | Core infrastructure — orchestrator, quality gates, base skills | Always deployed |
| **Tier 2 (Standard)** | Standard extensions — language-specific rules, framework skills | Deployed when the project's language/framework is detected |
| **Tier 3 (Optional)** | Optional — domain skills, platform-specific settings | Deployed only on explicit request or project configuration |

## Catalog File

The catalog manifest is defined in YAML format:

```yaml
# Example catalog entry
- id: moai-workflow-tdd
  tier: 1                    # 1=Core, 2=Standard, 3=Optional
  type: skill
  path: .claude/skills/moai/workflows/tdd.md
  languages: []              # empty array = all languages
  frameworks: []
  hash: abc123...             # content hash (integrity verification)
```

## SlimFS Filter

`moai init --slim` restricts the deployed files through the SlimFS filter:

```bash
# Full installation (all tiers)
moai init my-project

# Slim installation (Tier 1 + detected Tier 2 only)
moai init --slim my-project
```

### Filter Logic

1. Tier 1 is always included
2. Detects the project language (Go, Python, TypeScript, etc.)
3. Includes only the Tier 2 entries matching the detected language
4. Excludes Tier 3

## Typed Loader

The `LoadCatalog()` function loads the manifest in a type-safe way:

- Validates the 3-tier classification
- Checks hash integrity (Hash Sentinel)
- Detects missing fields
- 100% test coverage

## Using the Catalog

### Project Initialization

```bash
# Standard initialization — deploys all templates
moai init my-project

# Slim initialization — deploys only the minimal templates
moai init --slim my-project
```

### Updates

```bash
# Catalog-based update
moai update                  # updates all tiers
moai update --slim           # updates in slim mode
```

## Related Documentation

- [Installation](/getting-started/installation) — Installation guide
- [Initial Setup](/getting-started/init-wizard) — init wizard
- [Update](/getting-started/update) — Update guide
- [Skill Guide](/advanced/skill-guide) — Skill authoring guide
