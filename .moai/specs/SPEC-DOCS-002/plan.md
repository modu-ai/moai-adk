# @SPEC:DOCS-002 구현 계획서

## 개요

Phase 2는 VitePress 문서 사이트에 **핵심 개념 3개 페이지**를 추가하여 MoAI-ADK의 차별화된 가치(Alfred, TAG, TRUST)를 명확히 전달합니다.

---

## 작업 범위

### 1. Alfred 10개 AI 에이전트 팀 (~300 LOC)

**파일**: `docs/concepts/alfred-agents.md`

**콘텐츠 소스**: README.md 608~932줄

**구성**:

1. **Alfred SuperAgent 소개** (50 LOC)

   - 중앙 오케스트레이터 역할
   - 배트맨 집사 Alfred에서 영감
   - 9개 전문 에이전트 조율

2. **9개 전문 에이전트 표** (100 LOC)
   | 에이전트 | 페르소나 | 전문 영역 | 커맨드 | 위임 시점 |
   | --------------- | ----------------- | ---------------- | ---------------------- | -------------- |
   | spec-builder | 시스템 아키텍트 | SPEC 작성 | `/alfred:1-plan` | 명세 필요 시 |
   | code-builder | 수석 개발자 | TDD 구현 | `/alfred:2-run` | 구현 단계 |
   | doc-syncer | 테크니컬 라이터 | 문서 동기화 | `/alfred:3-sync` | 동기화 필요 시 |
   | tag-agent | 지식 관리자 | TAG 시스템 | `@agent-tag-agent` | TAG 작업 시 |
   | git-manager | 릴리스 엔지니어 | Git 워크플로우 | `@agent-git-manager` | Git 조작 시 |
   | debug-helper | 트러블슈팅 전문가 | 오류 진단 | `@agent-debug-helper` | 에러 발생 시 |
   | trust-checker | 품질 보증 리드 | TRUST 검증 | `@agent-trust-checker` | 검증 요청 시 |
   | cc-manager | 데브옵스 엔지니어 | Claude Code 설정 | `@agent-cc-manager` | 설정 필요 시 |
   | project-manager | 프로젝트 매니저 | 프로젝트 초기화 | `/alfred:0-project` | 프로젝트 시작 |

3. **Mermaid 오케스트레이션 다이어그램** (50 LOC)

   ```mermaid
   graph TD
       A[사용자 요청] --> B[Alfred 분석]
       B --> C{작업 분해}
       C -->|SPEC 작성| D[spec-builder]
       C -->|TDD 구현| E[code-builder]
       C -->|문서 동기화| F[doc-syncer]
       D --> G[품질 게이트]
       E --> G
       F --> G
       G --> H[Alfred 통합 보고]
   ```

4. **에이전트 협업 원칙** (100 LOC)
   - 단일 책임 원칙
   - 중앙 조율 (Alfred만 에이전트 간 작업 조율)
   - 품질 게이트 (TRUST, @TAG 검증)

### 2. @TAG 추적성 시스템 (~250 LOC)

**파일**: `docs/concepts/tag-system.md`

**콘텐츠 소스**: README.md 813~931줄

**구성**:

1. **CODE-FIRST 원칙** (50 LOC)

   - TAG의 진실은 코드 자체에만 존재
   - 중간 캐시 없음 (코드 직접 스캔)
   - `rg '@TAG' -n` 방식

2. **TAG 체인** (80 LOC)

   ```mermaid
   graph LR
       A[@SPEC:ID] --> B[@TEST:ID]
       B --> C[@CODE:ID]
       C --> D[@DOC:ID]
   ```

   - @SPEC → @TEST → @CODE → @DOC 4단계
   - TAG ID 규칙: `<도메인>-<3자리>`
   - 버전 관리: v0.1.0 (INITIAL) → v1.0.0 (안정화)

3. **언어별 TAG 사용 예시** (100 LOC)

   - TypeScript: `// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md`
   - Python: `# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md`
   - Dart: `// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md`

4. **TAG 검증 방법** (20 LOC)

   ```bash
   # 전체 TAG 스캔
   rg '@(SPEC|TEST|CODE|DOC):' -n

   # 고아 TAG 탐지
   rg 'AUTH-001' -n
   ```

### 3. TRUST 5원칙 (~200 LOC)

**파일**: `docs/concepts/trust-principles.md`

**콘텐츠 소스**: development-guide.md + README 분산 섹션

**구성**:

1. **T (Test First)** (40 LOC)

   - 테스트 커버리지 ≥85%
   - TDD Red-Green-Refactor 사이클
   - 언어별 테스트 프레임워크 (Jest, pytest, go test)

2. **R (Readable)** (40 LOC)

   - 파일 ≤300 LOC
   - 함수 ≤50 LOC
   - 복잡도 ≤10
   - 의도 드러내는 네이밍

3. **U (Unified)** (30 LOC)

   - 아키텍처 통합성
   - 일관된 코딩 스타일
   - 언어별 표준 도구 (ESLint, ruff, golint)

4. **S (Secured)** (50 LOC)

   - SQL Injection 방어
   - XSS 방어
   - CSRF 방어
   - 입력 검증, 환경 변수 사용

5. **T (Trackable)** (40 LOC)
   - @TAG 체인 무결성
   - 고아 TAG 자동 탐지
   - TAG ID 중복 방지

---

## 콘텐츠 소스 비율

**Phase 2 목표**: README 60% + dev-guide 30% + 신규 10%

| 페이지              | README        | dev-guide     | 신규          | 합계    |
| ------------------- | ------------- | ------------- | ------------- | ------- |
| alfred-agents.md    | 270 LOC (90%) | 0             | 30 LOC (10%)  | 300 LOC |
| tag-system.md       | 180 LOC (72%) | 20 LOC (8%)   | 50 LOC (20%)  | 250 LOC |
| trust-principles.md | 50 LOC (25%)  | 120 LOC (60%) | 30 LOC (15%)  | 200 LOC |
| **합계**            | 500 LOC (67%) | 140 LOC (19%) | 110 LOC (14%) | 750 LOC |

**결과**: README 67% (목표 60% 초과 달성) ✅

---

## 기술 스택

- **VitePress**: v1.6.4 (Phase 1에서 설정 완료)
- **Markdown**: GitHub Flavored Markdown
- **Mermaid**: 다이어그램 렌더링 (VitePress 기본 지원)
- **테스트**: Vitest (기존 설정 활용)

---

## 구현 전략

### Top-down 접근법

1. **VitePress Sidebar 확장** (우선):

   - `docs/.vitepress/config.mts` 수정
   - `concepts` 섹션에 3개 항목 추가

2. **페이지 작성** (순차):

   1. alfred-agents.md (가장 중요)
   2. tag-system.md
   3. trust-principles.md

3. **테스트 작성** (병렬):
   - 3개 페이지 존재 확인
   - VitePress 빌드 성공 검증
   - 링크 유효성 검증

### TDD 사이클 적용

- **RED**: 3개 페이지 존재 확인 테스트 작성 (실패)
- **GREEN**: 3개 페이지 작성 (최소 구현)
- **REFACTOR**: Mermaid 다이어그램 추가, 가독성 개선

---

## 예상 일정

| 작업                     | 예상 시간 | 담당 에이전트 |
| ------------------------ | --------- | ------------- |
| Sidebar 설정 확장        | 15분      | code-builder  |
| alfred-agents.md 작성    | 30분      | code-builder  |
| tag-system.md 작성       | 25분      | code-builder  |
| trust-principles.md 작성 | 20분      | code-builder  |
| 테스트 작성 및 검증      | 20분      | code-builder  |
| 빌드 확인 및 품질 검증   | 10분      | trust-checker |
| **합계**                 | **2시간** |               |

---

## 리스크 및 대응

### 기술적 리스크

1. **Mermaid 렌더링 실패**

   - **확률**: 낮음 (VitePress 기본 지원)
   - **대응**: 다이어그램 문법 검증, 사전 테스트

2. **README 콘텐츠 부족**

   - **확률**: 낮음 (608~932줄 충분)
   - **대응**: dev-guide.md 보완 활용

3. **빌드 시간 증가**
   - **확률**: 낮음 (페이지 3개 추가)
   - **대응**: 빌드 시간 모니터링 (< 3초 목표)

### 콘텐츠 리스크

1. **중복 내용**

   - **확률**: 중간 (Phase 1과 중복 가능)
   - **대응**: Phase 1 페이지 검토 후 차별화

2. **깊이 부족**
   - **확률**: 낮음 (README 상세 내용 충분)
   - **대응**: dev-guide.md 심화 내용 추가

---

## 다음 단계 (Phase 3-4)

**Phase 3** (Priority 2):

- Output Styles (4가지 스타일)
- CLI Reference
- 문제 해결

**Phase 4** (Priority 3):

- 언어별 가이드 (TypeScript, Python, Flutter, Go, Rust)
- 고급 주제
- 기여 가이드

---

**작성자**: spec-builder 에이전트
**검토자**: trust-checker 에이전트
**승인**: @Goos
