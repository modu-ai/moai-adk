# SPEC-SKILL-PORTFOLIO-OPT-001: 구현 계획 (Implementation Plan)

---
id: SPEC-SKILL-PORTFOLIO-OPT-001-PLAN
created_at: 2025-11-22
updated_at: 2025-11-22
version: 1.0.0
---

## 개요 (Overview)

본 구현 계획은 **MoAI-ADK Skills Portfolio 최적화 및 표준화**를 **4개 Phase**로 나누어 진행합니다.

**총 타임라인**: 3-5주 (Phase별 1-2주)
**총 작업량**: Story Points 13 (Fibonacci Scale)
**우선순위**: Critical

---

## Phase 1: 메타데이터 정규화 및 중복 제거 (1주)

### 목표
- 모든 127개 스킬에 필수 메타데이터 필드 추가
- 중복 스킬 3개 셋 병합하여 127개로 축소
- 비표준 명명 규칙 수정 (moai-domain-nano-banana → moai-google-nano-banana)

### 작업 항목

#### 1.1 메타데이터 표준화 (REQ-004)

**Task 1.1.1**: YAML Frontmatter 스키마 정의
- 필수 필드 7개 정의: `name`, `description`, `version`, `modularized`, `allowed-tools`, `last_updated`, `compliance_score`
- 선택 필드 9개 정의: `modules`, `dependencies`, `deprecated`, `successor`, `category_tier`, `auto_trigger_keywords`, `agent_coverage`
- 스키마 검증 스크립트 작성: `validate_skill_metadata.py`
- **예상 시간**: 1일

**Task 1.1.2**: 자동 메타데이터 추가 스크립트 구현
- 모든 127개 스킬의 SKILL.md 파일 읽기
- 기존 메타데이터와 병합 (중복 방지)
- 누락된 필드 자동 추가:
  - `version: 1.0.0` (기본값)
  - `last_updated: 2025-11-22` (스크립트 실행일)
  - `compliance_score: 0%` (초기값, 검증 후 업데이트)
- 자동 커밋: `git commit -m "feat: standardize skill metadata across 127 skills"`
- **예상 시간**: 1일

**Task 1.1.3**: 설명 품질 개선 (짧은 설명 확장)
- 100글자 미만 설명 39개 스킬 식별
- What + When 추가하여 100-200글자 범위로 확장
- 예시:
  - Before: "Session info"
  - After: "Provides session context and metadata for Claude Code. Use when debugging session issues or tracking agent activity across conversations."
- **예상 시간**: 2일

---

#### 1.2 중복 스킬 병합 (REQ-002)

**Task 1.2.1**: Docs 관련 중복 병합
- `moai-docs-generation` + `moai-docs-toolkit` → `moai-docs-toolkit` (통합)
  - 두 스킬의 모든 기능을 `moai-docs-toolkit/SKILL.md`에 병합
  - 모듈화: `modules/generation.md`, `modules/validation.md`
- `moai-docs-validation` + `moai-docs-linting` → `moai-docs-validation` (통합)
  - Linting 기능을 validation에 포함
- 원본 스킬 디렉토리를 `archived/` 이동
- **예상 시간**: 1일

**Task 1.2.2**: Testing/Security 중복 확인 및 병합
- 중복 테스트 스킬 식별 (예: moai-essentials-testing vs moai-testing-**)
- 중복 보안 스킬 식별 (예: moai-security-auth vs moai-auth-**)
- 병합 계획 수립 및 실행
- **예상 시간**: 1일

---

#### 1.3 비표준 명명 수정 (REQ-003)

**Task 1.3.1**: `moai-domain-nano-banana` 이름 변경
- 새 이름: `moai-google-nano-banana` (Google Nano 모델 명시)
- 파일 구조 이동:
  ```bash
  mv .claude/skills/moai-domain-nano-banana \
     .claude/skills/moai-google-nano-banana
  ```
- SKILL.md 내 `name` 필드 업데이트
- **예상 시간**: 0.5일

**Task 1.3.2**: 하위 호환성 보장
- Alias 생성: `moai-domain-nano-banana` → `moai-google-nano-banana`
- Migration guide 작성: `.moai/docs/skill-migration-guide.md`
- 35개 에이전트 파일에서 참조 확인 및 업데이트
- **예상 시간**: 0.5일

---

### Phase 1 출력물 (Deliverables)
- [ ] 127개 스킬 메타데이터 표준화 완료
- [ ] 중복 스킬 병합 완료 (134 → 127개)
- [ ] 비표준 명명 수정 완료 (moai-google-nano-banana)
- [ ] 메타데이터 검증 스크립트: `validate_skill_metadata.py`
- [ ] Migration guide: `.moai/docs/skill-migration-guide.md`

### Phase 1 품질 게이트
- ✅ 모든 127개 스킬에 필수 필드 7개 존재
- ✅ 중복 스킬 0개
- ✅ 명명 규칙 준수율 100%
- ✅ 설명 길이 100-200글자 비율 ≥60%
- ✅ YAML 파싱 오류 0개

---

## Phase 2: 명명 규칙 및 카테고리 통합 (1-2주)

### 목표
- 32개 도메인을 10개 티어로 통합
- 모든 스킬에 `category_tier` 필드 추가
- 티어별 스킬 목록 자동 생성

### 작업 항목

#### 2.1 카테고리 통합 (REQ-001)

**Task 2.1.1**: 10개 티어 정의 및 스킬 할당
- 티어 1 (moai-lang-*): 13개 스킬
- 티어 2 (moai-domain-*): 13개 스킬
- 티어 3 (moai-security-*): 8개 스킬
- 티어 4 (moai-core-*): 8개 스킬
- 티어 5 (moai-foundation-*): 5개 스킬
- 티어 6 (moai-cc-*): 7개 스킬
- 티어 7 (moai-baas-*): 10개 스킬
- 티어 8 (moai-essentials-*): 4개 스킬
- 티어 9 (moai-project-*): 4개 스킬
- 티어 10 (moai-lib-*): 1개 스킬
- 특수 스킬: 20개 (docs, design, mcp, etc.)
- **예상 시간**: 1일

**Task 2.1.2**: `category_tier` 필드 자동 추가
- 스킬 이름 패턴 분석하여 티어 자동 할당
- SKILL.md에 `category_tier: 1-10` 추가
- 특수 스킬은 `category_tier: special` 태그
- **예상 시간**: 0.5일

**Task 2.1.3**: 티어별 스킬 목록 생성
- `.moai/reports/skill-tiers-2025-11-22.md` 생성
- 10개 티어별로 스킬 목록 정리 (표 형식)
- 각 티어의 총 스킬 수 및 비율 표시
- **예상 시간**: 0.5일

---

#### 2.2 Gerund Form 검토 (선택사항)

**Task 2.2.1**: Gerund Form 권장 사항 검토
- Claude Code 공식 표준: "verb+ing" 형태 권장 (예: processing-pdfs)
- 현재 MoAI-ADK: 대부분 명사형 (예: moai-domain-cli-tool)
- 결정: **기존 유지** (명확성이 더 중요)
- 새 스킬 생성 시에만 Gerund Form 고려
- **예상 시간**: 0일 (검토만)

---

### Phase 2 출력물 (Deliverables)
- [ ] 10개 티어 정의 완료
- [ ] 127개 스킬에 `category_tier` 필드 추가
- [ ] 티어별 스킬 목록: `.moai/reports/skill-tiers-2025-11-22.md`
- [ ] 미분류 스킬 0개 검증

### Phase 2 품질 게이트
- ✅ 모든 스킬이 10개 티어 중 하나에 할당됨
- ✅ 티어별 스킬 수 균형 확인 (1-15개 범위)
- ✅ 특수 스킬 명확히 식별됨

---

## Phase 3: 문서 및 파일 품질 개선 (1-2주)

### 목표
- 파일 크기 500줄 초과 스킬 모듈화
- Progressive Disclosure 일관성 적용
- 모듈화 완성도 4% → 30% 향상

### 작업 항목

#### 3.1 파일 크기 초과 스킬 모듈화

**Task 3.1.1**: moai-essentials-debug 모듈화
- 현재 크기: 410+줄 (초과)
- SKILL.md: 핵심만 (Quick Ref + What + When) 250줄로 축약
- 모듈 생성:
  - `modules/ai-pattern-recognition.md`: AI 패턴 분석 로직
  - `modules/multi-process-debugging.md`: 다중 프로세스 디버깅
  - `modules/context7-integration.md`: Context7 통합
- **예상 시간**: 0.5일

**Task 3.1.2**: moai-lang-python 모듈화
- 현재 크기: 412줄 (경계선)
- SKILL.md: 350줄로 축약
- 모듈 생성:
  - `modules/fastapi-patterns.md`: FastAPI 전문 패턴
  - `modules/async-patterns.md`: Async/await 심화
  - `modules/testing-patterns.md`: Python 테스트 패턴
- **예상 시간**: 0.5일

**Task 3.1.3**: moai-domain-backend 모듈화
- 현재 크기: 476줄 (경계선)
- SKILL.md: 공통 백엔드 패턴만 300줄로 축약
- 모듈 생성:
  - `modules/fastapi-advanced.md`: FastAPI 심화
  - `modules/django-optimization.md`: Django 최적화
  - `modules/express-patterns.md`: Express.js 패턴
- **예상 시간**: 0.5일

**Task 3.1.4**: 기타 초과 스킬 모듈화 (2-3개 추가)
- 파일 크기 분석하여 >500줄 스킬 식별
- 각 스킬별 모듈화 계획 수립 및 실행
- **예상 시간**: 1일

---

#### 3.2 Progressive Disclosure 일관성

**Task 3.2.1**: `modularized` 필드 표준화
- 모듈화된 스킬: `modularized: true` + `modules: [list]`
- 모듈화 안된 스킬: `modularized: false`
- 자동 검증 스크립트: `validate_modularization.py`
- **예상 시간**: 0.5일

**Task 3.2.2**: 3-Level 문서 구조 검증
- Level 1: Metadata (항상 로드)
- Level 2: SKILL.md (필요시 로드)
- Level 3: modules/, examples.md, reference.md (온디맨드)
- 파일 참조 규칙 검증: 단일 레벨만 (modules/file.md ✓, reference/subdir/file.md ❌)
- **예상 시간**: 0.5일

---

### Phase 3 출력물 (Deliverables)
- [ ] 파일 크기 초과 스킬 0개 (모두 <500줄)
- [ ] 모듈화 완성도 30% 이상 (38개+ 스킬)
- [ ] `modularized` 필드 표준화 완료
- [ ] Progressive Disclosure 일관성 검증 완료

### Phase 3 품질 게이트
- ✅ 모든 스킬 파일 크기 <500줄
- ✅ 모듈화된 스킬 비율 ≥30%
- ✅ 파일 참조 규칙 준수율 100%

---

## Phase 4: 신규 필수 스킬 5개 추가 (1주)

### 목표
- 현대적 개발 워크플로우 지원을 위한 5개 스킬 생성
- Auto-Trigger 로직 CLAUDE.md 통합
- Agent-Skill 커버리지 85% 달성

### 작업 항목

#### 4.1 신규 스킬 5개 생성 (REQ-005)

**Task 4.1.1**: moai-core-code-templates 생성
- 목적: 재사용 가능한 코드 템플릿 생성 및 관리
- 주요 기능:
  - FastAPI boilerplate 템플릿
  - React component 템플릿
  - Vue 3 composition API 템플릿
  - Next.js 14+ App Router 템플릿
- 파일 구조:
  ```
  moai-core-code-templates/
  ├── SKILL.md (핵심 설명)
  ├── modules/
  │   ├── fastapi-templates.md
  │   ├── react-templates.md
  │   ├── vue-templates.md
  │   └── nextjs-templates.md
  └── examples.md (10+ 템플릿 예제)
  ```
- **예상 시간**: 1일

**Task 4.1.2**: moai-security-api-versioning 생성
- 목적: API 버전 관리 및 하위 호환성 전략
- 주요 기능:
  - Semantic API versioning (v1, v2, ...)
  - Deprecation 관리 및 마이그레이션 가이드
  - REST, GraphQL, gRPC 버전 관리 패턴
- 파일 구조:
  ```
  moai-security-api-versioning/
  ├── SKILL.md
  ├── modules/
  │   ├── rest-versioning.md
  │   ├── graphql-versioning.md
  │   └── grpc-versioning.md
  └── examples.md
  ```
- **예상 시간**: 1일

**Task 4.1.3**: moai-essentials-testing-integration 생성
- 목적: 통합 테스트 전략 및 E2E 테스트 자동화
- 주요 기능:
  - Integration test 패턴 (API contract testing)
  - E2E 시나리오 자동화 (Playwright, Cypress)
  - Test fixture 관리
- 파일 구조:
  ```
  moai-essentials-testing-integration/
  ├── SKILL.md
  ├── modules/
  │   ├── api-contract-testing.md
  │   ├── e2e-automation.md
  │   └── test-fixtures.md
  └── examples.md
  ```
- **예상 시간**: 1일

**Task 4.1.4**: moai-essentials-performance-profiling 생성
- 목적: 성능 프로파일링 및 최적화 전략
- 주요 기능:
  - CPU/Memory 프로파일링 (Python cProfile, Node.js profiler)
  - 병목 현상 자동 분석
  - 최적화 제안 (알고리즘, 데이터 구조)
- 파일 구조:
  ```
  moai-essentials-performance-profiling/
  ├── SKILL.md
  ├── modules/
  │   ├── cpu-profiling.md
  │   ├── memory-profiling.md
  │   └── optimization-strategies.md
  └── examples.md
  ```
- **예상 시간**: 1일

**Task 4.1.5**: moai-security-accessibility-wcag3 생성
- 목적: WCAG 3.0 접근성 표준 준수 검증
- 주요 기능:
  - A11y 자동 검증 (axe-core, Pa11y)
  - ARIA 속성 자동 검사
  - 키보드 네비게이션 테스트
  - Lighthouse 통합
- 파일 구조:
  ```
  moai-security-accessibility-wcag3/
  ├── SKILL.md
  ├── modules/
  │   ├── wcag3-guidelines.md
  │   ├── automated-testing.md
  │   └── aria-best-practices.md
  └── examples.md
  ```
- **예상 시간**: 1일

---

#### 4.2 Auto-Trigger 로직 구현 (REQ-006)

**Task 4.2.1**: CLAUDE.md에 Auto-Trigger 섹션 추가
- 위치: `CLAUDE.md` → **Rule 8: Config 기반 자동 동작** 섹션
- 키워드 매핑 테이블 추가:
  ```markdown
  ### Auto-Trigger 로직

  Alfred는 사용자 요청의 키워드를 분석하여 적절한 스킬을 자동 선택합니다.

  | 키워드 | 자동 호출 스킬 | 에이전트 |
  |--------|---------------|---------|
  | "authentication", "jwt", "oauth" | moai-security-auth | backend-expert |
  | "optimize", "performance" | moai-essentials-perf | performance-engineer |
  | "SPEC", "specification" | moai-foundation-specs | spec-builder |
  | "test", "testing", "integration" | moai-essentials-testing-integration | test-engineer |
  | ... | ... | ... |
  ```
- **예상 시간**: 0.5일

**Task 4.2.2**: `auto_trigger_keywords` 필드 추가
- 모든 127개 스킬에 `auto_trigger_keywords` 배열 추가
- 예시:
  ```yaml
  auto_trigger_keywords:
    - authentication
    - jwt
    - oauth
    - token
  ```
- **예상 시간**: 0.5일

---

#### 4.3 Agent-Skill 커버리지 85% 달성 (REQ-007)

**Task 4.3.1**: 35개 에이전트 스킬 참조 분석
- `.moai/memory/agents.md` 파일 분석
- 각 에이전트의 스킬 참조 여부 확인
- 현재 커버리지 계산: `(참조 에이전트 수 / 35) × 100`
- **예상 시간**: 0.5일

**Task 4.3.2**: 미참조 에이전트에 스킬 추가
- 커버리지 85% 미달 시 (30개 미만 참조):
  - 각 에이전트의 역할 분석
  - 적절한 스킬 매핑
  - 에이전트 파일에 스킬 참조 추가
- 예시:
  ```yaml
  # .claude/agents/backend-expert.yaml
  skills:
    - moai-domain-backend
    - moai-lang-python
    - moai-security-auth
  ```
- **예상 시간**: 1일

**Task 4.3.3**: `agent_coverage` 필드 추가
- 모든 127개 스킬에 `agent_coverage` 배열 추가
- 해당 스킬을 참조하는 에이전트 목록 명시
- 예시:
  ```yaml
  agent_coverage:
    - spec-builder
    - tdd-implementer
    - backend-expert
  ```
- **예상 시간**: 0.5일

---

### Phase 4 출력물 (Deliverables)
- [ ] 신규 스킬 5개 생성 완료
- [ ] CLAUDE.md Auto-Trigger 로직 통합 완료
- [ ] Agent-Skill 커버리지 ≥85% 달성
- [ ] `auto_trigger_keywords` 필드 127개 스킬에 추가
- [ ] `agent_coverage` 필드 127개 스킬에 추가

### Phase 4 품질 게이트
- ✅ 신규 스킬 5개 모두 메타데이터 완성
- ✅ Auto-Trigger 키워드 정확도 ≥95% (테스트)
- ✅ Agent 커버리지 ≥85% (30개 이상 에이전트)
- ✅ 모든 스킬에 `auto_trigger_keywords` 존재

---

## 마일스톤 (Milestones)

### Milestone 1: 메타데이터 정규화 완료 (1주차 종료)
- **목표**: 모든 스킬 메타데이터 표준화, 중복 제거, 명명 수정
- **검증**: 필수 필드 100%, 중복 0개, 명명 규칙 100%

### Milestone 2: 카테고리 통합 완료 (2주차 종료)
- **목표**: 10개 티어 정의, 스킬 할당, 티어별 목록 생성
- **검증**: 미분류 스킬 0개, 티어별 균형 확인

### Milestone 3: 문서 품질 개선 완료 (3주차 종료)
- **목표**: 파일 크기 초과 0개, 모듈화 30%, Progressive Disclosure 일관성
- **검증**: 모든 파일 <500줄, 모듈화 비율 ≥30%

### Milestone 4: 신규 스킬 및 Auto-Trigger 완료 (4주차 종료)
- **목표**: 5개 신규 스킬 생성, Auto-Trigger 로직 통합, Agent 커버리지 85%
- **검증**: 신규 스킬 메타데이터 완성, 키워드 정확도 ≥95%, 커버리지 ≥85%

---

## 기술 접근 방식 (Technical Approach)

### 자동화 스크립트
1. **validate_skill_metadata.py**: YAML 메타데이터 검증
2. **standardize_metadata.py**: 메타데이터 자동 추가
3. **validate_modularization.py**: 모듈화 일관성 검증
4. **analyze_agent_coverage.py**: Agent-Skill 커버리지 분석
5. **auto_trigger_keyword_matcher.py**: 키워드 매칭 정확도 테스트

### Git 워크플로우
- **모드**: `develop_direct` (브랜치 생성 안함)
- **커밋 전략**: Phase별 atomic commit
  - Phase 1 완료: `feat: standardize metadata and merge duplicates`
  - Phase 2 완료: `feat: consolidate 32 domains into 10 tiers`
  - Phase 3 완료: `feat: improve file quality with modularization`
  - Phase 4 완료: `feat: add 5 new essential skills and auto-trigger logic`

### 테스트 전략
- **유닛 테스트**: 각 스크립트별 pytest 테스트 케이스
- **통합 테스트**: Agent-Skill 통합 검증
- **E2E 테스트**: Auto-Trigger 로직 키워드 매칭 정확도

---

## 리스크 및 대응 계획 (Risks & Mitigation)

### 리스크 1: 하위 호환성 파괴
**완화 전략**:
- Alias 생성 (moai-domain-nano-banana → moai-google-nano-banana)
- Migration guide 작성 (`.moai/docs/skill-migration-guide.md`)
- 35개 에이전트 파일 참조 자동 검증 스크립트

### 리스크 2: 병합 후 기능 손실
**완화 전략**:
- 병합 전 기능 목록 완전 비교
- 테스트 커버리지 100% 유지
- 원본 스킬을 `archived/`로 이동 (복구 가능)

### 리스크 3: Auto-Trigger 오작동
**완화 전략**:
- Fallback 메커니즘 (키워드 매칭 실패 시 사용자 선택 프롬프트)
- 키워드 매칭 정확도 테스트 (≥95%)
- 사용자 피드백 루프 (오작동 보고 및 개선)

### 리스크 4: YAML 파싱 오류
**완화 전략**:
- YAML 검증 스크립트 자동 실행 (CI/CD 통합)
- 파싱 오류 시 자동 롤백
- 모든 변경사항 Git 추적 (복구 용이)

---

## 아키텍처 디자인 방향

### 메타데이터 구조
```yaml
# Required Fields (7)
name: moai-[category]-[feature]
description: "What + When + How (100-200글자)"
version: X.Y.Z
modularized: true | false
allowed-tools: [Read, Bash, WebFetch]
last_updated: YYYY-MM-DD
compliance_score: XX%

# Optional Fields (9)
modules: [list]
dependencies: [list]
deprecated: false
successor: null
category_tier: 1-10
auto_trigger_keywords: [list]
agent_coverage: [list]
```

### Auto-Trigger 로직 흐름도

```
사용자 요청
  ↓
키워드 추출 (NLP or 단순 매칭)
  ↓
auto_trigger_keywords 매칭
  ↓
매칭된 스킬 우선순위 정렬
  ↓
최우선 스킬 자동 선택
  ↓
에이전트 호출 (agent_coverage 참조)
  ↓
작업 실행
```

### Agent-Skill 커버리지 매핑

```
35 Agents
  ├── spec-builder → moai-foundation-specs, moai-core-spec-authoring
  ├── tdd-implementer → moai-foundation-trust, moai-essentials-testing-integration
  ├── backend-expert → moai-domain-backend, moai-lang-python, moai-security-auth
  ├── frontend-expert → moai-domain-frontend, moai-lang-javascript, moai-lang-typescript
  ├── ...
  └── (30개 이상 에이전트에 스킬 매핑 필요)
```

---

## 다음 단계 (Next Steps)

1. **Phase 1 시작**: 메타데이터 표준화 스크립트 구현
2. **테스트 케이스 작성**: 각 REQ별 테스트 시나리오 정의
3. **CI/CD 통합**: YAML 검증 스크립트 자동 실행
4. **사용자 피드백**: Auto-Trigger 로직 정확도 모니터링

---

**Last Updated**: 2025-11-22
**Version**: 1.0.0
