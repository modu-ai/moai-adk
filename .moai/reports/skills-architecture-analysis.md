# MoAI-ADK Skills 아키텍처 심층 분석 보고서

> **작성일**: 2025-10-19
> **분석 범위**: 전체 46개 스킬 구조 적정성 평가
> **참조 문서**: Anthropic 공식 문서 3건

---

## 📋 Executive Summary

**핵심 질문**: "스킬이 이렇게 많은게 정상인가?"

**결론**: **부분적으로 적정, 일부 개선 필요**

- ✅ **Language Skills (24개)**: 완벽한 구조 - 변경 불필요
- ✅ **Domain Skills (9개)**: 적절한 구조 - 변경 불필요
- ✅ **Claude Code Skill (1개)**: 통합 완료 - 변경 불필요
- ⚠️ **Alfred Skills (12개)**: **개선 필요 - 3-4개로 통합 권장**

**권장 조치**: 46개 → **37-38개 스킬**로 최적화 (Alfred 스킬 통합)

---

## 🔍 Anthropic 공식 가이드라인 분석

### 핵심 원칙 1: Progressive Disclosure

> "Skills let Claude load information only as needed, similar to a well-organized manual that starts with a table of contents, then specific chapters, and finally a detailed appendix."

**의미**:
- Claude는 필요한 스킬만 선택적으로 로드
- 전체 스킬을 한 번에 로드하지 않음
- 따라서 **스킬 수가 많아도 컨텍스트 비용 증가 없음**

### 핵심 원칙 2: Unwieldy Threshold (비대화 기준점)

> "When the SKILL.md file becomes unwieldy, split its content into separate files and reference them."

**의미**:
- SKILL.md가 "비대해지면" 분리
- 기준: 명확한 숫자 없음, 주관적 판단
- 현재 스킬: 평균 **60-70줄** (비대하지 않음)

### 핵심 원칙 3: Mutual Exclusivity (상호 배타성)

> "If certain contexts are mutually exclusive or rarely used together, keeping the paths separate will reduce the token usage."

**의미**:
- **상호 배타적**인 컨텍스트는 분리 유지 → 토큰 효율 ↑
- **함께 사용**되는 컨텍스트는 통합 고려 → 토큰 효율 ↑

**역설명**: 자주 함께 사용되면 통합이 나을 수 있음

---

## 📊 현재 46개 스킬 분석

### 카테고리별 분포

| 카테고리            | 개수 | 평균 크기 | 상호 배타성            | Progressive Disclosure | 평가        |
| ------------------- | ---- | --------- | ---------------------- | ---------------------- | ----------- |
| **Language Skills** | 24개 | ~64줄     | ✅ 높음 (Python ≠ Java) | ✅ 우수                 | ✅ 최적      |
| **Domain Skills**   | 9개  | ~68줄     | ✅ 중간 (일부 중복)     | ✅ 양호                 | ✅ 적절      |
| **Alfred Skills**   | 12개 | ~68줄     | ❌ 낮음 (함께 사용)     | ⚠️ 보통                 | ⚠️ 개선 필요 |
| **Claude Code**     | 1개  | ~67줄     | N/A                    | N/A                    | ✅ 통합 완료 |

### Language Skills (24개) - ✅ 최적 구조

**구조**:
```
moai-lang-python       (pytest, mypy, ruff, black, uv)
moai-lang-typescript   (vitest, tsc, biome)
moai-lang-java         (junit, maven, checkstyle)
... (21개 더)
```

**분석**:
- ✅ 각 스킬 ~60-70줄 (비대하지 않음)
- ✅ 상호 배타적 (Python 프로젝트 ≠ Java 프로젝트)
- ✅ Progressive Disclosure 완벽 작동
- ✅ 토큰 효율 최대 (필요한 언어만 로드)

**예시**:
```
사용자: "/alfred:2-run AUTH-001"
Alfred: (Python 프로젝트 감지)
        → moai-lang-python만 로드
        → 다른 23개 언어 스킬은 로드하지 않음
        → 토큰 효율 ↑
```

**권장**: **변경 없음 - 현재 구조 유지**

### Domain Skills (9개) - ✅ 적절한 구조

**구조**:
```
moai-domain-backend       (Server architecture, API design)
moai-domain-frontend      (React/Vue, state management)
moai-domain-database      (Schema, indexing, migration)
moai-domain-mobile-app    (Flutter, React Native)
... (5개 더)
```

**분석**:
- ✅ 각 스킬 ~60-70줄 (비대하지 않음)
- ✅ 대부분 상호 배타적 (Backend ≠ Mobile App)
- ⚠️ 일부 중복 (Backend + Web API, Frontend + Mobile App)
- ✅ Progressive Disclosure 양호

**권장**: **변경 없음 - 현재 구조 유지**

(중복이 있지만, 통합 시 비대해질 위험)

### Alfred Skills (12개) - ⚠️ 개선 필요

**현재 구조**:
```
1. moai-alfred-code-reviewer
2. moai-alfred-debugger-pro
3. moai-alfred-ears-authoring
4. moai-alfred-feature-selector
5. moai-alfred-git-workflow
6. moai-alfred-language-detection
7. moai-alfred-performance-optimizer
8. moai-alfred-refactoring-coach
9. moai-alfred-spec-metadata-validation
10. moai-alfred-tag-scanning
11. moai-alfred-template-generator
12. moai-alfred-trust-validation
```

**문제점**:
- ❌ **상호 배타적이지 않음** (함께 사용됨)
- ❌ `/alfred:3-sync` 실행 시 3-4개 스킬 동시 로드
  - trust-validation + tag-scanning + spec-metadata-validation
- ❌ Progressive Disclosure 비효율 (여러 스킬을 동시에 로드해야 함)

**Anthropic 원칙 위반**:
> "If certain contexts are mutually exclusive or rarely used together, keeping the paths separate will reduce the token usage."

→ Alfred 스킬들은 **자주 함께 사용**되므로 **분리 유지가 토큰 낭비**

**권장**: **통합 (12개 → 3-4개)**

---

## 💡 재설계 권장안

### 제안 1: Alfred Skills 워크플로우 기반 통합 (12개 → 3개)

#### 1. moai-alfred-planning (기획 단계)
**통합 대상 (4개 → 1개)**:
- ~~moai-alfred-feature-selector~~ → planning
- ~~moai-alfred-language-detection~~ → planning
- ~~moai-alfred-template-generator~~ → planning
- ~~moai-alfred-ears-authoring~~ → planning

**크기 예상**: ~270줄 (4 × 68줄)
**사용 시점**: `/alfred:1-plan` 실행 시
**장점**: 기획 관련 모든 기능을 한 번에 로드

#### 2. moai-alfred-validation (검증 단계)
**통합 대상 (3개 → 1개)**:
- ~~moai-alfred-trust-validation~~ → validation
- ~~moai-alfred-tag-scanning~~ → validation
- ~~moai-alfred-spec-metadata-validation~~ → validation

**크기 예상**: ~200줄 (3 × 68줄)
**사용 시점**: `/alfred:3-sync` 실행 시
**장점**: 검증 관련 모든 기능을 한 번에 로드 (현재도 동시 로드됨)

#### 3. moai-alfred-development (개발 단계)
**통합 대상 (5개 → 1개)**:
- ~~moai-alfred-code-reviewer~~ → development
- ~~moai-alfred-debugger-pro~~ → development
- ~~moai-alfred-performance-optimizer~~ → development
- ~~moai-alfred-refactoring-coach~~ → development
- ~~moai-alfred-git-workflow~~ → development

**크기 예상**: ~340줄 (5 × 68줄)
**사용 시점**: `/alfred:2-run` 및 코드 작성 중
**장점**: 개발 관련 모든 기능을 한 번에 로드

**결과**:
- 전체 스킬: 46개 → **37개** (9개 감소)
- Alfred 스킬: 12개 → **3개** (9개 감소)
- 평균 통합 스킬 크기: ~270줄 (여전히 비대하지 않음)

### 제안 2: Alfred Skills 기능별 통합 (12개 → 4개)

#### 1. moai-alfred-quality (품질 관리)
**통합 대상 (5개 → 1개)**:
- code-reviewer + performance-optimizer + refactoring-coach + trust-validation + tag-scanning

#### 2. moai-alfred-structure (구조 관리)
**통합 대상 (4개 → 1개)**:
- spec-metadata-validation + ears-authoring + template-generator + feature-selector

#### 3. moai-alfred-workflow (워크플로우)
**통합 대상 (2개 → 1개)**:
- git-workflow + language-detection

#### 4. moai-alfred-debugging (디버깅)
**통합 대상 (1개)**:
- debugger-pro (변경 없음)

**결과**:
- 전체 스킬: 46개 → **38개** (8개 감소)
- Alfred 스킬: 12개 → **4개** (8개 감소)

---

## 📈 통합 전후 비교

### 현재 구조 (46개)

**장점**:
- ✅ 단일 책임 원칙 준수 (각 스킬이 하나의 기능만 담당)
- ✅ 스킬 발견이 쉬움 (명확한 네이밍)

**단점**:
- ❌ Alfred 스킬들이 함께 로드됨 (토큰 비효율)
- ❌ 유지보수 부담 (12개 Alfred 스킬 개별 업데이트)

### 제안 구조 (37개 - 제안 1 기준)

**장점**:
- ✅ 토큰 효율 증가 (워크플로우별 한 번에 로드)
- ✅ 유지보수 간소화 (3개 Alfred 스킬만 관리)
- ✅ Progressive Disclosure 개선
- ✅ Anthropic 원칙 준수 (함께 사용되는 컨텍스트 통합)

**단점**:
- ⚠️ 스킬 크기 증가 (~270줄, 여전히 비대하지 않음)
- ⚠️ 일부 단일 책임 원칙 완화

---

## 🎯 최종 권장사항

### Language Skills (24개)
**권장**: ✅ **현재 구조 유지 (변경 없음)**

**이유**:
- 상호 배타적 (Python ≠ Java)
- Progressive Disclosure 완벽
- 토큰 효율 최대

### Domain Skills (9개)
**권장**: ✅ **현재 구조 유지 (변경 없음)**

**이유**:
- 대부분 상호 배타적
- 통합 시 비대해질 위험
- 현재 구조가 적절

### Claude Code Skill (1개)
**권장**: ✅ **현재 구조 유지 (변경 없음)**

**이유**:
- 이미 5개 템플릿 통합 완료
- 적절한 크기

### Alfred Skills (12개)
**권장**: ⚠️ **통합 권장 (12개 → 3개)**

**제안**: **제안 1 (워크플로우 기반)** 채택
- moai-alfred-planning (4개 통합)
- moai-alfred-validation (3개 통합)
- moai-alfred-development (5개 통합)

**이유**:
- 현재: 함께 사용되는 스킬들이 분리되어 있음 (토큰 비효율)
- 개선: 워크플로우별 통합 → 토큰 효율 증가
- Anthropic 원칙 준수: "자주 함께 사용되면 통합"

---

## 📝 구현 계획

### Phase 1: SPEC 작성
- SPEC-SKILL-CONSOLIDATE-001 생성
- Alfred 스킬 12개 → 3개 통합 계획 수립

### Phase 2: 스킬 통합
- moai-alfred-planning 생성 (4개 통합)
- moai-alfred-validation 생성 (3개 통합)
- moai-alfred-development 생성 (5개 통합)

### Phase 3: 기존 스킬 삭제
- 12개 개별 Alfred 스킬 삭제
- Templates 디렉토리 동기화

### Phase 4: 검증
- Progressive Disclosure 테스트
- 토큰 효율 측정
- 기능 회귀 테스트

---

## 🔬 근거 자료

### Anthropic 공식 문서 인용

1. **Progressive Disclosure**:
   > "Skills let Claude load information only as needed, similar to a well-organized manual that starts with a table of contents, then specific chapters, and finally a detailed appendix."

2. **Mutual Exclusivity**:
   > "If certain contexts are mutually exclusive or rarely used together, keeping the paths separate will reduce the token usage."

3. **Unwieldy Threshold**:
   > "When the SKILL.md file becomes unwieldy, split its content into separate files and reference them."

### 토큰 효율 분석

**현재 (Alfred 12개 분리)**:
```
/alfred:3-sync 실행 시:
- trust-validation (68줄)
- tag-scanning (68줄)
- spec-metadata-validation (68줄)
= 총 204줄 로드
```

**제안 (Alfred 3개 통합)**:
```
/alfred:3-sync 실행 시:
- validation (200줄)
= 총 200줄 로드 (4줄 감소, YAML 중복 제거)
```

---

## ✅ 결론

**Q: 스킬이 이렇게 많은게 정상인가?**

**A: 부분적으로 정상, 일부 개선 필요**

- ✅ **Language/Domain 스킬 (33개)**: 완벽히 정상 - Anthropic 원칙 준수
- ⚠️ **Alfred 스킬 (12개)**: 개선 필요 - 상호 배타성 원칙 위반

**최종 권장**:
- **46개 → 37개 스킬**로 최적화
- Alfred 스킬만 통합 (12개 → 3개)
- Language/Domain 스킬은 현재 구조 유지

**다음 단계**:
1. 사용자 승인 대기
2. SPEC-SKILL-CONSOLIDATE-001 작성
3. Alfred 스킬 통합 실행

---

**작성**: Alfred (MoAI SuperAgent)
**검토 필요**: @Goos
**최종 업데이트**: 2025-10-19
