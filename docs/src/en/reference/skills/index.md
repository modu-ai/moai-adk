# Claude Skills System Reference

Complete guide to MoAI-ADK's 93 Claude Skills.

## Overview

**Claude Skills** are **reusable knowledge capsules** that Alfred utilizes. Each Skill contains prompts, examples, and best practices optimized for a specific domain or technology.

### Skill Features

- **Progressive Disclosure**: Loaded on-demand only when needed
- **Modular**: Independently maintainable
- **Reusable**: Shareable across multiple agents
- **Version Control**: Track versions of each Skill
- **Documented**: Each Skill includes its own documentation

## Skill Classification

### 6 Layers

```
┌────────────────────────────────────────────┐
│  Foundation (Foundation)                   │
│  TRUST, TAG, SPEC writing, Git workflow    │
└────────────────────────────────────────────┘
                    ↓
┌────────────────────────────────────────────┐
│  Essentials (Essential)                    │
│  Debugging, performance, refactoring, testing │
└────────────────────────────────────────────┘
                    ↓
┌────────────────────────────────────────────┐
│  Alfred (Alfred-specific)                  │
│  Agent guides, workflow, decision-making   │
└────────────────────────────────────────────┘
                    ↓
┌────────────────────────────────────────────┐
│  Domain (Domain)                           │
│  Backend, frontend, security, DB, ML       │
└────────────────────────────────────────────┘
                    ↓
┌────────────────────────────────────────────┐
│  Languages (Languages)                     │
│  Python, TypeScript, Go, Rust, etc. 20     │
└────────────────────────────────────────────┘
                    ↓
┌────────────────────────────────────────────┐
│  CC (Claude Code)                          │
│  Configuration, permissions, MCP, Hooks management │
└────────────────────────────────────────────┘
```

## <span class="material-icons">library_books</span> Skills List (93)

### 1️⃣ Foundation (Foundation Skills)

**Essential basic skills for all projects**

| Skill                      | Description                                                       | Documentation                                 |
| -------------------------- | ----------------------------------------------------------------- | ---------------------------------------------- |
| **moai-foundation-trust** | TRUST 5 principles (Test, Readable, Unified, Secured, Trackable) | Foundation |
| **moai-foundation-tags**  | TAG system and traceability (@SPEC, @TEST, @CODE, @DOC)           | Foundation |
| **moai-alfred-workflow**  | Alfred 4-step workflow                                            | Alfred    |

### 2️⃣ Essentials (Essential Skills)

**Skills frequently used during development**

| Skill                         | Description                            | When to Use         |
| ----------------------------- | -------------------------------------- | ------------------- |
| **moai-essentials-debug**    | Advanced debugging, stack trace analysis | When errors occur      |
| **moai-essentials-perf**     | Performance optimization, bottleneck analysis | When performance improvement needed |
| **moai-essentials-refactor** | Refactoring guide, design patterns      | When code improvement needed |
| **moai-essentials-review**   | Automated code review                  | Pre-commit verification      |

### 3️⃣ Alfred (Alfred-specific)

**Skills for Alfred and sub-agents**

| Skill                               | Description                        | Target                |
| ----------------------------------- | ---------------------------------- | --------------------- |
| **moai-alfred-agent-guide**        | 19-member team structure, selection algorithm | Agent team management    |
| **moai-alfred-ask-user-questions** | Optimal AskUserQuestion usage      | User interaction     |
| **moai-alfred-personas**           | Alfred adaptive persona            | Communication style |
| **moai-alfred-best-practices**     | TRUST, TAG, Skill invocation rules | Quality assurance           |
| **moai-alfred-context-budget**     | Context window optimization        | Memory management         |

### 4️⃣ Domain (Domain Skills)

**Domain expert knowledge**

#### Backend

- **moai-domain-backend**: API, server, microservices
- **moai-domain-web-api**: REST API, GraphQL design

#### Frontend

- **moai-domain-frontend**: React, Vue, Angular
- **moai-design-systems**: Design systems, accessibility

#### Data & Performance

- **moai-domain-database**: DB design, optimization, migration
- **moai-domain-ml**: Machine learning, model training, deployment

#### Infrastructure & Security

- **moai-domain-security**: OWASP, security vulnerabilities, compliance
- **devops-expert**: Docker, Kubernetes, CI/CD

#### Mobile

- **moai-domain-mobile-app**: Flutter, React Native

### 5️⃣ Languages (Language Skills)

**Best practices by programming language**

#### Popular Languages (8)

- **moai-lang-python**: Python 3.13+ (pytest, mypy, ruff, uv)
- **moai-lang-typescript**: TypeScript 5.7+ (Vitest, Biome)
- **moai-lang-javascript**: JavaScript ES2024+ (Jest, ESLint, Prettier)
- **moai-lang-go**: Go 1.24+ (go test, golangci-lint)
- **moai-lang-rust**: Rust 1.84+ (cargo, clippy)
- **moai-lang-kotlin**: Kotlin 2.1+ (KMP, coroutines)
- **moai-lang-java**: Java 21+ (Maven, Gradle, JUnit)
- **moai-lang-csharp**: C# 13+ (.NET 8, xUnit)

#### Other Languages (12)

- **moai-lang-php**: PHP 8.4+ (Laravel, Symfony)
- **moai-lang-ruby**: Ruby 3.4+ (Rails, RSpec)
- **moai-lang-sql**: SQL (pgTAP, sqlfluff)
- **moai-lang-shell**: Shell scripting (bats-core, shellcheck)
- **moai-lang-r**: R 4.4+ (testthat, lintr)
- **moai-lang-cpp**: C++ 20+ (Catch2, CMake)
- **moai-lang-c**: C17/C23 (Unity, cppcheck)
- **moai-lang-dart**: Dart 3.x (Flutter, null safety)
- **moai-lang-scala**: Scala 3+ (ScalaTest, SBT)
- **moai-lang-swift**: Swift 5.9+ (XCTest, SPM)
- **moai-lang-haskell**: Haskell (HUnit, Cabal)
- **moai-lang-template**: Templates for 13 other languages

### 6️⃣ Claude Code (CC) Configuration

**Claude Code configuration and integration**

| Skill                      | Description                               |
| -------------------------- | ------------------------------------------ |
| **moai-cc-configuration** | settings.json, permissions, hooks          |
| **moai-cc-memory**        | Session memory, Context window optimization |
| **moai-cc-skill-factory** | Skill creation and maintenance             |
| **moai-cc-claude-md**     | CLAUDE.md project guideline writing        |

## Skill Selection Guide

### Skills by Situation

```
Error occurs
    └─→ moai-essentials-debug

Performance issues
    └─→ moai-essentials-perf

Code improvement needed
    └─→ moai-essentials-refactor

New feature development (API)
    ├─→ moai-domain-backend
    ├─→ moai-domain-web-api
    └─→ moai-lang-python

New feature development (UI)
    ├─→ moai-domain-frontend
    ├─→ moai-design-systems
    └─→ moai-lang-typescript

Deployment/CI/CD
    └─→ devops-expert (or moai-domain-devops)

Database
    └─→ moai-domain-database

Security review
    └─→ moai-domain-security
```

## Skill Usage Patterns

### Skill Invocation

```python
# Within Alfred sub-agent
Skill("moai-lang-python")  # Load Python best practices

# Or auto-activation
# - Detect "Python" keyword in SPEC
# - Automatically load appropriate language Skill
```

### Progressive Disclosure

```
Request
    ↓
Alfred (intent analysis)
    ├─ "Python API" detected
    ├─ Skill("moai-lang-python") loaded
    ├─ Skill("moai-domain-backend") loaded
    └─ Skill("moai-domain-web-api") loaded
    ↓
Only necessary Skills loaded into memory
```

## Skills Statistics

- **Total Skills**: 93
- **Foundation**: 3
- **Essentials**: 4
- **Alfred**: 5
- **Domain**: 8
- **Languages**: 20
- **CC**: 4

## Detailed References

- **[Foundation Skills](foundation.md)** - Foundation skills details
- **[Language Skills](languages.md)** - Language-specific skills details
- **[Alfred Skills](alfred.md)** - Alfred-specific skills details

## 🆘 Skill FAQ

### "Required Skill not available"

→ You can request or suggest new Skills in GitHub Issues

### "I want to manually load a Skill"

→ All Skills can be explicitly invoked with `Skill("skill-name")`

### "Can I use multiple Skills simultaneously?"

→ Yes, you can combine multiple Skills by loading them sequentially

______________________________________________________________________

**Next**: [Foundation Skills](foundation.md) or [Language Skills](languages.md)
