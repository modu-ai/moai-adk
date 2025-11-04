# Skills Description 개선 정책

> MoAI-ADK의 모든 Skills에 적용되는 description 작성 표준

---

## 🎯 설명 (Description) 작성 원칙

### 1. 기본 구조

모든 Skill의 description은 다음 3가지를 포함해야 합니다:

```
[What it does]: 간단한 기능 설명 (8-12 단어)
[Key capabilities]: 핵심 기능 목록 (2-3개)
[When to use]: 사용 시점 (3-5개 트리거 키워드, "Use when" 사용)
```

### 2. 작성 템플릿

**기본 템플릿**:
```
description: [기능 설명]. [핵심 기능]. Use when [트리거1], [트리거2], or [트리거3].
```

**확장 템플릿** (필요 시):
```
description: [기능 설명]. [핵심 기능1], [핵심 기능2]. Use when [트리거1], [트리거2], [트리거3], or [트리거4]. Automatically activates [연관 Skills] for [목적].
```

### 3. Skills 카테고리별 템플릿

#### Foundation Skills (moai-foundation-*)
```
description: [기능]. [핵심 기능 (예: validation, authoring)]. Use when [작업1], [작업2], [작업3], or [작업4].
```
**예시**:
- ✅ "Validates SPEC YAML frontmatter (7 required fields) and HISTORY section. Use when creating SPEC documents, validating SPEC metadata, checking SPEC structure, or authoring specifications."
- ❌ "SPEC metadata validation" (너무 짧음)

#### Alfred Skills (moai-alfred-*)
```
description: [기능]. [핵심 능력]. Use when [검증/분석/관리 대상], [조건], or [상황]. Automatically activates [연관 Skills] for [목적].
```
**예시**:
- ✅ "Validates TRUST 5-principles (Test 85%+, Code constraints, Architecture unity, Security, TAG trackability). Use when validating code quality, checking TRUST compliance, verifying test coverage, or analyzing security patterns. Automatically activates moai-foundation-trust and language-specific skills for comprehensive validation."
- ❌ "TRUST validation" (너무 짧고 트리거 키워드 없음)

#### Language Skills (moai-lang-*)
```
description: [언어] best practices with [주요 도구]. Use when [개발 활동], [패턴], or [특수 케이스].
```
**예시**:
- ✅ "Python best practices with pytest, mypy, ruff, black. Use when writing Python code, implementing tests, type-checking, formatting code, or following PEP standards."
- ❌ "Python best practices" (너무 일반적)

#### Domain Skills (moai-domain-*)
```
description: [도메인] development with [주요 기술/패턴]. Use when [도메인 활동1], [도메인 활동2], or [특수 상황].
```
**예시**:
- ✅ "Backend API development with REST patterns, authentication, error handling. Use when designing REST APIs, implementing authentication, or building backend services."
- ❌ "Backend development" (구체성 부족)

### 4. 트리거 키워드 가이드

#### 작업 중심 키워드
- "creating [artifact]", "writing [code/docs]", "implementing [feature]"
- "validating [aspect]", "checking [quality]", "verifying [compliance]"
- "analyzing [code/data]", "debugging [issue]", "diagnosing [problem]"

#### 조건 중심 키워드
- "when working with [file type/framework]"
- "when developing [feature/component]"
- "when applying [pattern/practice]"

#### 상황 중심 키워드
- "for [workflow/process]", "during [phase/stage]"
- "to [achieve goal/outcome]"

### 5. 금지 사항 (Anti-patterns)

❌ **너무 짧음**:
```
description: Helps with documents
```

❌ **트리거 키워드 없음**:
```
description: PDF processing tool
```

❌ **"I can", "You can" 사용** (첫 인칭 회피):
```
description: I help you process Excel files
```

❌ **기술적 세부사항만**:
```
description: Uses pdfplumber library
```

---

## 📊 Skills 목록별 개선 체크리스트

### Priority 1: Foundation Skills (7개)
- [ ] moai-foundation-specs ✅ 완료
- [ ] moai-foundation-ears ✅ 완료
- [ ] moai-foundation-tags
- [ ] moai-foundation-trust
- [ ] moai-foundation-langs
- [ ] moai-claude-code
- [ ] moai-foundation-git

### Priority 2: Alfred Skills (10개)
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

### Priority 3: Language Skills (20개)
- moai-lang-typescript
- moai-lang-python
- moai-lang-go
- moai-lang-rust
- moai-lang-java
- ... (14개 더)

### Priority 4: Domain Skills (12개)
- moai-domain-backend
- moai-domain-frontend
- moai-domain-web-api
- moai-domain-database
- ... (8개 더)

### Priority 5: Essentials Skills (4개)
- moai-essentials-debug
- moai-essentials-review
- moai-essentials-refactor
- moai-essentials-perf

---

## 🔧 개선 방법

### 방법 1: 개별 Edit (고품질)
각 Skill의 `description` 필드를 수작업 Edit
- 이점: 각 Skill에 최적화된 description
- 단점: 시간이 오래 걸림 (60+ Skills)

### 방법 2: 정책 기반 일괄 개선 (효율성)
템플릿을 기반으로 Pattern matching으로 우선순위별 개선
1. Priority 1-2: 개별 Edit (가장 중요)
2. Priority 3-5: 템플릿 기반 일괄 적용

### 현재 진행 상태
- ✅ moai-foundation-specs (Priority 1)
- ✅ moai-foundation-ears (Priority 1)
- ⏳ 나머지 50+ Skills (Priority 2-5)

---

## 📝 작성 예시

### 좋은 예시

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

### 수정 전 vs 수정 후

| Skill | Before | After |
|-------|--------|-------|
| moai-foundation-specs | "Validates SPEC YAML frontmatter (7 required fields) and HISTORY section" | "Validates SPEC YAML frontmatter (7 required fields id, version, status, created, updated, author, priority) and HISTORY section. Use when creating SPEC documents, validating SPEC metadata, checking SPEC structure, or authoring specifications." |
| moai-foundation-ears | "EARS requirement authoring guide (Ubiquitous/Event/State/Optional/Constraints)" | "EARS requirement authoring guide covering Ubiquitous/Event/State/Optional/Constraints syntax patterns. Use when writing requirements, authoring specifications, defining system behavior, or creating functional requirements." |

---

## ✅ 최종 검증 체크리스트

각 Skill의 description이 다음을 만족하는지 확인:

- [ ] **"What it does"**: 명확한 기능 설명 (명사 + 동사)
- [ ] **"Use when"**: 3~5개 구체적 트리거 키워드 포함
- [ ] **길이**: 한 줄 (150-200 자 권장)
- [ ] **관계**: 연관된 Skills 명시 (선택)
- [ ] **발견성**: 검색 키워드 포함 (sub-agent가 발견 가능)
- [ ] **첫 인칭 회피**: "I", "You", "Our" 사용 안 함
- [ ] **기술 중심성 회피**: 라이브러리/도구명보다 기능/목적 우선

---

**작성일**: 2025-10-20
**최종 업데이트**: Phase 2 진행 중
