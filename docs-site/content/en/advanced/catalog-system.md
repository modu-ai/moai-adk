---
title: Catalog System
weight: 80
draft: false
---

Tokenomics is not a principle that applies only to tokens. Every template file deployed into a project is ultimately a candidate for the context a session will load. The catalog system reduces this cost from the initialization stage on, following the principle of "deploy only what is needed."

## Overview

The catalog system in MoAI-ADK v2.15+ manages every agent, skill, plugin, and rule through a **3-tier manifest**. With `moai init --slim`, only the minimum templates a project needs are selected and deployed, so initialization is faster and the files left in the project stay lightweight.

## The 3-Tier Manifest

Every deployable item belongs to one of three tiers.

| Tier | Description | Deployment criterion |
|------|------|----------|
| **Tier 1 (Core)** | Core infrastructure — orchestrator, quality gates, base skills | Always deployed |
| **Tier 2 (Standard)** | Standard extensions — per-language rules, framework skills | When the project language/framework is detected |
| **Tier 3 (Optional)** | Optional — domain skills, platform-specific settings | On explicit request or project configuration |

## The Catalog File

The catalog manifest is defined in YAML format.

```yaml
# 카탈로그 엔트리 예시
- id: moai-workflow-tdd
  tier: 1                    # 1=Core, 2=Standard, 3=Optional
  type: skill
  path: .claude/skills/moai/workflows/tdd.md
  languages: []              # 빈 배열 = 모든 언어
  frameworks: []
  hash: abc123...             # 콘텐츠 해시 (무결성 검증)
```

The `hash` field carries a content hash, so the loader can verify whether a deployed file has been corrupted or arbitrarily modified.

## The SlimFS Filter

`moai init --slim` restricts deployed files through the SlimFS filter.

```bash
# 전체 설치 (모든 계층)
moai init my-project

# Slim 설치 (Tier 1 + 감지된 Tier 2만)
moai init --slim my-project
```

### Filter Logic

The filter operates in four steps.

1. Tier 1 is always included
2. Project language detection (Go, Python, TypeScript, etc.)
3. Only Tier 2 items matching the detected languages are included
4. Tier 3 is excluded

## Typed Loader

The `LoadCatalog()` function loads the manifest in a type-safe way. Because it validates struct by struct rather than relying on string parsing, manifest errors are caught before deployment.

- 3-tier classification validation
- Hash integrity checks (Hash Sentinel)
- Missing-field detection
- 100% test coverage

## Using the Catalog

### Project Initialization

```bash
# 일반 초기화 — 모든 템플릿 배포
moai init my-project

# Slim 초기화 — 최소 템플릿만 배포
moai init --slim my-project
```

### Updates

Updates operate against the same catalog, so a project initialized with slim should also be updated with slim.

```bash
# 카탈로그 기반 업데이트
moai update                  # 모든 계층 업데이트
moai update --slim           # slim 모드로 업데이트
```

## Related Documents

- [Installation](/en/getting-started/installation) — installation guide
- [Initial Setup](/en/getting-started/init-wizard) — the init wizard
- [Updating](/en/getting-started/update) — update guide
- [Skill Guide](/en/advanced/skill-guide) — skill authoring guide
