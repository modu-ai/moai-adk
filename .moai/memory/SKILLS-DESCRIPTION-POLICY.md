# Skills Description Improvement Policy

> Description authoring standards applied to all Skills in MoAI-ADK

---

## 🎯 Description Authoring Principles

### 1. Basic Structure

All Skill descriptions must include the following 3 elements:

```
[What it does]: Brief functional description (8-12 words)
[Key capabilities]: List of core functions (2-3 items)
[When to use]: Usage timing (3-5 trigger keywords, use "Use when")
```

### 2. Authoring Templates

**Basic Template**:
```
description: [Feature description]. [Core capabilities]. Use when [trigger1], [trigger2], or [trigger3].
```

**Extended Template** (when needed):
```
description: [Feature description]. [Core capability1], [Core capability2]. Use when [trigger1], [trigger2], [trigger3], or [trigger4]. Automatically activates [related Skills] for [purpose].
```

### 3. Skills Category-specific Templates

#### Foundation Skills (moai-foundation-*)
```
description: [Function]. [Core capabilities (e.g., validation, authoring)]. Use when [task1], [task2], [task3], or [task4].
```
**Example**:
- ✅ "Validates SPEC YAML frontmatter (7 required fields) and HISTORY section. Use when creating SPEC documents, validating SPEC metadata, checking SPEC structure, or authoring specifications."
- ❌ "SPEC metadata validation" (too short)

#### Alfred Skills (moai-alfred-*)
```
description: [Function]. [Core capabilities]. Use when [validation/analysis/management target], [condition], or [situation]. Automatically activates [related Skills] for [purpose].
```
**Example**:
- ✅ "Validates TRUST 5 principles (Test 85%+, Code constraints, Architecture unity, Security, TAG trackability). Use when validating code quality, checking TRUST compliance, verifying test coverage, or analyzing security patterns. Automatically activates moai-foundation-trust and language-specific skills for comprehensive validation."
- ❌ "TRUST validation" (too short and no trigger keywords)

#### Language Skills (moai-lang-*)
```
description: [Language] best practices with [main tools]. Use when [development activity], [pattern], or [special case].
```
**Example**:
- ✅ "Python best practices with pytest, mypy, ruff, black. Use when writing Python code, implementing tests, type-checking, formatting code, or following PEP standards."
- ❌ "Python best practices" (too general)

#### Domain Skills (moai-domain-*)
```
description: [Domain] development with [main technologies/patterns]. Use when [domain activity1], [domain activity2], or [special situation].
```
**Example**:
- ✅ "Backend API development with REST patterns, authentication, error handling. Use when designing REST APIs, implementing authentication, or building backend services."
- ❌ "Backend development" (lack of specificity)

### 4. Trigger Keyword Guide

#### Task-focused Keywords
- "creating [artifact]", "writing [code/docs]", "implementing [feature]"
- "validating [aspect]", "checking [quality]", "verifying [compliance]"
- "analyzing [code/data]", "debugging [issue]", "diagnosing [problem]"

#### Condition-focused Keywords
- "when working with [file type/framework]"
- "when developing [feature/component]"
- "when applying [pattern/practice]"

#### Situation-focused Keywords
- "for [workflow/process]", "during [phase/stage]"
- "to [achieve goal/outcome]"

### 5. Prohibited Patterns (Anti-patterns)

❌ **Too Short**:
```
description: Helps with documents
```

❌ **No Trigger Keywords**:
```
description: PDF processing tool
```

❌ **Using "I can", "You can"** (avoid first person):
```
description: I help you process Excel files
```

❌ **Technical Details Only**:
```
description: Uses pdfplumber library
```

---

## 📊 Skills List Improvement Checklist

### Priority 1: Foundation Skills (7 items)
- [ ] moai-foundation-specs ✅ Complete
- [ ] moai-foundation-ears ✅ Complete
- [ ] moai-foundation-tags
- [ ] moai-foundation-trust
- [ ] moai-foundation-langs
- [ ] moai-claude-code
- [ ] moai-foundation-git

### Priority 2: Alfred Skills (10 items)
- [ ] moai-alfred-tag-scanning
- [ ] moai-alfred-trust-validation
- [ ] moai-alfred-spec-metadata-validation
- [ ] moai-alfred-code-reviewer
- [ ] moai-alfred-git-workflow
- [ ] moai-alfred-ears-authoring
- [ ] moai-alfred-debugger-pro
- [ ] moai-alfred-language-detection
- [ ] moai-alfred-performance-optimizer
- [ ] moai-alfred-refactoring-coach

### Priority 3: Language Skills (20 items)
- moai-lang-typescript
- moai-lang-python
- moai-lang-go
- moai-lang-rust
- moai-lang-java
- ... (14 more)

### Priority 4: Domain Skills (12 items)
- moai-domain-backend
- moai-domain-frontend
- moai-domain-web-api
- moai-domain-database
- ... (8 more)

### Priority 5: Essentials Skills (4 items)
- moai-essentials-debug
- moai-essentials-review
- moai-essentials-refactor
- moai-essentials-perf

---

## 🔧 Improvement Methods

### Method 1: Individual Edit (High quality)
Manually edit each Skill's `description` field
- Advantage: Optimized description for each Skill
- Disadvantage: Time-consuming (60+ Skills)

### Method 2: Policy-based Batch Improvement (Efficiency)
Pattern matching-based improvement by priority using templates
1. Priority 1-2: Individual Edit (most important)
2. Priority 3-5: Batch application using templates

### Current Progress Status
- ✅ moai-foundation-specs (Priority 1)
- ✅ moai-foundation-ears (Priority 1)
- ⏳ Remaining 50+ Skills (Priority 2-5)

---

## 📝 Authoring Examples

### Good Examples

#### Foundation Skill
```yaml
description: Validates SPEC YAML frontmatter (7 required fields id, version, status, created, updated, author, priority) and HISTORY section. Use when creating SPEC documents, validating SPEC metadata, checking SPEC structure, or authoring specifications.
```

#### Alfred Skill
```yaml
description: Generates descriptive commit messages by analyzing git diffs. Use when writing commit messages, reviewing staged changes, or summarizing code modifications. Automatically activates git-workflow skill for advanced patterns.
```

#### Language Skill
```yaml
description: TypeScript best practices with Vitest, Biome, strict typing. Use when implementing TypeScript code, writing tests, checking code quality, or applying type safety patterns.
```

#### Domain Skill
```yaml
description: REST API design patterns with authentication, versioning, error handling. Use when designing REST APIs, implementing authentication, building backend services, or managing API versions.
```

### Before vs After

| Skill | Before | After |
|-------|--------|-------|
| moai-foundation-specs | "Validates SPEC YAML frontmatter (7 required fields) and HISTORY section" | "Validates SPEC YAML frontmatter (7 required fields id, version, status, created, updated, author, priority) and HISTORY section. Use when creating SPEC documents, validating SPEC metadata, checking SPEC structure, or authoring specifications." |
| moai-foundation-ears | "EARS requirement authoring guide (Ubiquitous/Event/State/Optional/Constraints)" | "EARS requirement authoring guide covering Ubiquitous/Event/State/Optional/Constraints syntax patterns. Use when writing requirements, authoring specifications, defining system behavior, or creating functional requirements." |

---

## ✅ Final Validation Checklist

Verify that each Skill's description satisfies the following:

- [ ] **"What it does"**: Clear functional description (noun + verb)
- [ ] **"Use when"**: Includes 3~5 specific trigger keywords
- [ ] **Length**: Single line (recommended 150-200 characters)
- [ ] **Relationships**: Specify related Skills (optional)
- [ ] **Discoverability**: Includes search keywords (sub-agents can discover)
- [ ] **Avoid First Person**: Do not use "I", "You", "Our"
- [ ] **Avoid Technical Centricity**: Prioritize function/purpose over library/tool names

---

**Created**: 2025-10-20
**Last Updated**: Phase 2 in Progress
**Document Version**: v1.0.0
