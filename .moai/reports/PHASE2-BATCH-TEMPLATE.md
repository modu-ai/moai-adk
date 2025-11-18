# PHASE 2: Language Skills Batch Processing Template

**Batch Update Guide for remaining 17 Language Skills (moai-lang-*)**

Updated: 2025-11-19 | Target: All 21 Language Skills to v4.0.0

---

## Overview

This template provides the pattern for updating all remaining 17 Language Skills to enterprise v4.0.0 standard. The 4 representative Skills are already updated:

✅ **Updated (Representative Sample)**:
- moai-lang-python (3.13.9, FastAPI 0.121.0)
- moai-lang-typescript (5.9.3, Node.js 24.11.0)
- moai-lang-go (1.25.x, Fiber v3)
- moai-lang-shell (5.2.37, ShellCheck 0.10.0)

⏳ **Remaining (17 Skills)**:
- moai-lang-c, moai-lang-cpp, moai-lang-csharp
- moai-lang-dart, moai-lang-html-css, moai-lang-java
- moai-lang-javascript, moai-lang-kotlin, moai-lang-php
- moai-lang-r, moai-lang-ruby, moai-lang-rust
- moai-lang-scala, moai-lang-sql, moai-lang-swift
- moai-lang-tailwind-css, moai-lang-template

---

## Version Matrix: November 2025 Stable

| Language | Version | Framework | Package Manager | Release Date | Support |
|----------|---------|-----------|-----------------|--------------|---------|
| **Python** | 3.13.9 | FastAPI 0.121.0, Django 5.2 | pip/uv | Oct 2025 | Oct 2029 |
| **TypeScript** | 5.9.3 | Next.js 16, React 19.2 | npm/pnpm | Aug 2025 | Active |
| **JavaScript** | ES2024 | Node.js 24.11.0 | npm/pnpm | Oct 2024 | Apr 2027 |
| **Go** | 1.25.4 | Fiber v3, gRPC | go mod | Nov 2025 | Active |
| **Java** | 21 LTS | Spring Boot 3.4 | Maven/Gradle | Sep 2023 | Sep 2031 |
| **Kotlin** | 2.1.0 | Spring Boot 3.4 | Gradle | Oct 2024 | Active |
| **Rust** | 1.83.0 | Actix-web 4.9 | Cargo | Nov 2024 | Active |
| **C** | C23 | CMake 3.31 | gcc/clang | 2023 | Active |
| **C++** | C++23 | CMake 3.31 | gcc/clang | 2023 | Active |
| **C#** | 13.0 | .NET 9.0 | NuGet | Nov 2024 | Active |
| **Ruby** | 3.4.2 | Rails 8.0 | Bundler | Dec 2024 | Active |
| **PHP** | 8.4 | Laravel 11 | Composer | Nov 2024 | Active |
| **SQL** | ANSI SQL 2023 | PostgreSQL 17 | N/A | Oct 2024 | Active |
| **R** | 4.5.1 | tidyverse 2.0.0 | CRAN | Nov 2024 | Active |
| **Dart** | 3.6.0 | Flutter 3.27 | Pub | Nov 2024 | Active |
| **Swift** | 6.0 | SwiftUI | Swift Package Manager | Mar 2025 | Active |
| **Scala** | 3.5 | Play Framework 3.2 | SBT | Oct 2024 | Active |
| **Shell** | Bash 5.2 | ShellCheck 0.10 | Native | Jan 2025 | Active |

---

## Standard Skill.md Structure (v4.0.0)

All Language Skills follow this Progressive Disclosure format:

```markdown
---
name: "moai-lang-{language}"
version: "4.0.0"
tier: Language
description: |
  Enterprise {Language} with {key_feature_1}, {key_feature_2}, {key_framework}.
  Activates for {use_case_1}, {use_case_2}, {use_case_3}.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# Modern {Language} Development — Enterprise v4.0

## Quick Summary

**Primary Focus**: {Language} X.X with {Framework}, {Tool}, and production patterns
**Best For**: {use_case_list}
**Key Libraries**: {library1}, {library2}, {library3}
**Auto-triggers**: {keyword1}, {keyword2}, {keyword3}

| Version | Release | Support |
|---------|---------|---------|
| {Lang} X.X.X | {Month} {Year} | {Support Date} |
| {Framework} X.X | {Month} {Year} | {Support Date} |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core {Language} concepts:
- {Concept 1}: Brief description
- {Concept 2}: Brief description
- {Concept 3}: Brief description
- **Examples**: See `examples.md` for full code samples

### Level 2: Advanced Patterns (See reference.md)

Production-ready patterns:
- {Pattern 1}: Description with benefits
- {Pattern 2}: Description with benefits
- {Pattern 3}: Description with benefits
- **Reference**: See `reference.md` for API details

### Level 3: Production Deployment (Consult specialized Skills)

Enterprise deployment:
- {Deployment 1}: Optimization details
- {Deployment 2}: Monitoring patterns
- {Deployment 3}: Scaling strategies
- **Details**: Skill("moai-essentials-perf"), Skill("moai-domain-devops")

---

## Technology Stack (November 2025 Stable)

### Runtime & Core
- **{Language} X.X.X** (Latest, {Month} {Year})
  - {Key feature 1}
  - {Key feature 2}
- {Runtime 2}: Standard library / built-in support
- {Runtime 3}: Native or common pattern

### Frameworks & Libraries
- **{Framework 1}** (X.X, production-ready)
  - {Feature 1}
  - {Feature 2}
- **{Framework 2}** (X.X, alternative option)
- **{Library 1}** (X.X, standard choice)

### Build & Package Management
- **{Package Manager}** (X.X)
  - {Feature 1}
  - {Feature 2}

### Testing & Quality
- **{Testing Framework}** (X.X)
- **{Linter/Formatter}** (X.X)

---

## Quick Reference: Hello World & Core Patterns

### Minimal Example
\`\`\`{language}
// Core language syntax
// Production-ready minimal example
\`\`\`

### Framework Example
\`\`\`{language}
// Using primary framework
// Show key feature and benefits
\`\`\`

### Production Pattern
\`\`\`{language}
// Real-world production pattern
// Show error handling and best practices
\`\`\`

---

## Production Best Practices

1. **Always use {best practice 1}** for {benefit}
2. **Prefer {pattern 1}** over {anti-pattern} for {benefit}
3. **Use {tool 1}** for {use case}
4. **Implement {pattern 2}** for {security/performance}
5. **Test {aspect}** with {testing framework}

---

## Learn More

- **Examples**: See `examples.md` for {framework}, {library}, and {pattern} examples
- **Reference**: See `reference.md` for API details, configuration, and troubleshooting
- **Official Docs**: {link_to_official_docs}
- **Community**: {community_resource_link}

---

**Skills**: Skill("moai-essentials-debug"), Skill("moai-essentials-perf"), Skill("moai-domain-{domain}")
**Auto-loads**: {Language} projects mentioning {keyword1}, {keyword2}, {keyword3}
```

---

## Checklist: Per-Language Skill Updates

Use this checklist for EACH language skill update:

### Frontmatter (YAML)
- [ ] `name`: `moai-lang-{language}` (kebab-case)
- [ ] `version`: `4.0.0`
- [ ] `tier`: `Language`
- [ ] `description`: ~150 chars, includes framework + auto-trigger keywords
- [ ] `allowed-tools`: Read, Bash, WebSearch, WebFetch (all 4)
- [ ] `status`: `stable`

### Content Structure
- [ ] **Quick Summary**: Version table with current framework versions
- [ ] **Three-Level Path**: Fundamentals → Advanced → Production
- [ ] **Tech Stack**: Runtime, Frameworks, Package Manager, Testing, Quality
- [ ] **3-5 Code Examples**: Hello World, Framework, Production Pattern
- [ ] **10 Best Practices**: Language-specific do's and don'ts
- [ ] **Learn More Links**: Official docs + community resources
- [ ] **Skills References**: Cross-link to moai-essentials-* and moai-domain-*

### Auto-triggers
- [ ] Keywords in description (3-5 common search terms)
- [ ] Bottom of file lists auto-load conditions
- [ ] Examples match auto-trigger keywords

### Version Accuracy
- [ ] Check official documentation for current version
- [ ] Verify release date and support timeline
- [ ] Confirm deprecated features are NOT recommended
- [ ] Update any security advisories

### Cross-references
- [ ] All internal Skill references use `Skill("name")` format
- [ ] External links are HTTPS and valid
- [ ] References to examples.md and reference.md present

---

## Template Variables Reference

| Variable | Meaning | Examples |
|----------|---------|----------|
| `{Language}` | Full language name | Python, TypeScript, Go |
| `{language}` | Lowercase for code | python, typescript, go |
| `{key_feature_1}` | Primary differentiation | async/await, strict typing, systems programming |
| `{Framework}` | Main web/app framework | FastAPI, Next.js, Fiber |
| `{use_case_1}` | Primary use case | REST APIs, Web Applications, CLI tools |
| `{Version}` | Current stable version | 3.13.9, 5.9.3, 1.25.4 |
| `{Month} {Year}` | Release date | November 2025, August 2025 |
| `{Support Date}` | End of support | Oct 2029, Apr 2027, Active |
| `{keyword1}` | Auto-trigger keyword | fastapi, typescript, goroutine |

---

## Language-Specific Notes

### Python (moai-lang-python) ✅ UPDATED
- Latest: Python 3.13.9 (Oct 2025)
- Framework: FastAPI 0.121.0, Django 5.2 LTS
- Highlights: JIT compiler (PEP 744), free-threaded mode
- Package Manager: uv, pip (modern approach)

### TypeScript (moai-lang-typescript) ✅ UPDATED
- Latest: TypeScript 5.9.3 (Aug 2025)
- Framework: Next.js 16, React 19.2
- Highlights: Strict mode, type inference, Turbopack
- Package Manager: npm, pnpm, yarn

### Go (moai-lang-go) ✅ UPDATED
- Latest: Go 1.25.4 (Nov 2025)
- Framework: Fiber v3, gRPC
- Highlights: Goroutines, static typing, fast compilation
- Package Manager: go mod

### Shell (moai-lang-shell) ✅ UPDATED
- Latest: Bash 5.2 (Jan 2025)
- Framework: ShellCheck, bats-core
- Highlights: POSIX compliance, defensive scripting
- Package Manager: Native (apt, brew, etc.)

### JavaScript (moai-lang-javascript)
- Latest: ES2024 standard
- Framework: Node.js 24.11.0 (Oct 2024), Express, Fastify
- Highlights: Native async/await, modules (ESM)
- Package Manager: npm, pnpm

### Java (moai-lang-java)
- Latest: Java 21 LTS (Sept 2023, support until Sept 2031)
- Framework: Spring Boot 3.4, Quarkus
- Highlights: JVM optimization, native compilation (GraalVM)
- Package Manager: Maven, Gradle

### Rust (moai-lang-rust)
- Latest: Rust 1.83.0 (Nov 2024)
- Framework: Actix-web 4.9, Tokio async
- Highlights: Memory safety, zero-cost abstractions
- Package Manager: Cargo

### C# (moai-lang-csharp)
- Latest: C# 13.0 (.NET 9.0, Nov 2024)
- Framework: ASP.NET Core, Unity
- Highlights: LINQ, async/await, nullable reference types
- Package Manager: NuGet

### Kotlin (moai-lang-kotlin)
- Latest: Kotlin 2.1.0 (Oct 2024)
- Framework: Spring Boot 3.4, Android
- Highlights: Null safety, coroutines, extension functions
- Package Manager: Gradle, Maven

### Ruby (moai-lang-ruby)
- Latest: Ruby 3.4.2 (Dec 2024)
- Framework: Rails 8.0, Sinatra
- Highlights: Convention over configuration, REPL
- Package Manager: Bundler

### PHP (moai-lang-php)
- Latest: PHP 8.4 (Nov 2024)
- Framework: Laravel 11, Symfony
- Highlights: Type hints, attributes, fibers (async)
- Package Manager: Composer

### SQL (moai-lang-sql)
- Latest: ANSI SQL 2023
- Database: PostgreSQL 17, MySQL 8.4
- Highlights: CTEs, window functions, JSON support
- Package Manager: N/A (dialect-specific)

---

## Time Estimates

| Category | Count | Per-Skill | Total |
|----------|-------|-----------|-------|
| **Language Skills** | 17 | 15-20 min | 4-5.5 hours |
| **With validation** | 17 | 20-25 min | 5.5-7 hours |
| **Parallel (3 devs)** | 17 | 15-20 min | 1.5-2 hours |

---

## Automation Approach

For faster batch processing, consider:

1. **Template Substitution**: Use script to populate version matrix
2. **Auto-version Detection**: Query official docs (npm, PyPI, etc.)
3. **Link Validation**: Check all external links for 200 status
4. **Spell Check**: Run against common typos
5. **Structure Validation**: Verify YAML frontmatter and markdown headers

### Bash Script Template
```bash
#!/bin/bash
# Batch update Language Skills

LANGUAGES=(javascript java kotlin php ruby rust scala csharp cpp c r dart swift html-css tailwind-css template)

for lang in "${LANGUAGES[@]}"; do
  echo "Processing moai-lang-$lang..."
  
  # 1. Read current version
  current=$(grep "version:" "/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills/moai-lang-$lang/SKILL.md" | head -1)
  
  # 2. Check if needs update (not 4.0.0)
  if [[ ! "$current" =~ "4.0.0" ]]; then
    echo "  → Needs update from $(echo $current | cut -d'"' -f2)"
    
    # 3. Generate updated SKILL.md using template
    # (Implement template rendering logic here)
  else
    echo "  → Already v4.0.0"
  fi
done
```

---

## Quality Gates

Each updated Skill must pass:

### ✅ MUST PASS
- [ ] YAML frontmatter valid
- [ ] Version is exactly "4.0.0"
- [ ] Status is "stable"
- [ ] All 4 allowed-tools present
- [ ] Markdown headers valid and complete
- [ ] No duplicate content

### ✅ SHOULD PASS
- [ ] Tech stack versions current (within 3 months)
- [ ] All external links return 200 status
- [ ] No spelling errors
- [ ] Cross-references use `Skill()` format
- [ ] 3-5 code examples with proper syntax highlighting
- [ ] 10+ best practices listed

### ✅ NICE TO HAVE
- [ ] Personality/voice matches other Skills
- [ ] Consistent terminology across all Skills
- [ ] Examples show real-world use cases
- [ ] Performance tips included where relevant

---

## Next Steps

1. **Copy this template** for each of 17 remaining languages
2. **Populate version matrix** from official sources (November 2025 data)
3. **Generate 3-5 examples** per language using latest best practices
4. **Validate links** to official documentation
5. **Cross-reference** with moai-domain-* and moai-essentials-* Skills
6. **Run TRUST 5 validation** (see PHASE4 guide)
7. **Update timestamps** to 2025-11-19 or current date

---

## References

- **Python**: https://www.python.org/downloads/
- **TypeScript**: https://www.typescriptlang.org/
- **JavaScript**: https://nodejs.org/
- **Go**: https://go.dev/
- **Java**: https://www.oracle.com/java/
- **Rust**: https://www.rust-lang.org/
- **All Others**: Visit official language websites

---

**Scope**: Language Skills batch update strategy
**Version**: 1.0 (2025-11-19)
**Purpose**: Standardize all 21 Language Skills to v4.0.0 Enterprise standard
**Estimated Completion**: 5-7 hours (sequential) or 1.5-2 hours (parallel)

