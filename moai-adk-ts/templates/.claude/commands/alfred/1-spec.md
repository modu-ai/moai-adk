---
name: alfred:1-spec
description: EARS 명세 작성 + 브랜치/PR 생성
argument-hint: "제목1 제목2 ... | SPEC-ID 수정내용"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash
---

# 🏗️ MoAI-ADK 1단계: EARS 명세 작성 + 브랜치/PR 생성

**우선순위:** 커맨드 지침은 에이전트 지침보다 **상위**이다.

## 🎯 커맨드 목적

프로젝트 문서를 분석하여 EARS 구문의 명세서를 작성하고, Personal/Team 모드에 따라 Git 브랜치 및 PR을 생성합니다.

**SPEC 자동 제안/생성 대상**: $ARGUMENTS

## 📋 실행 흐름

1. **프로젝트 분석**: product/structure/tech.md 심층 분석
2. **SPEC 후보 발굴**: 비즈니스 요구사항 기반 우선순위 결정
3. **사용자 확인**: 작성 계획 검토 및 승인
4. **SPEC 작성**: EARS 구조의 명세서 생성 (spec.md, plan.md, acceptance.md)
5. **Git 작업**: git-manager를 통한 브랜치/PR 생성

## 🔗 연관 에이전트

- **Primary**: spec-builder (🏗️ 설계자) - SPEC 문서 작성 전담
- **Secondary**: git-manager (🌿 정원사) - Git 브랜치/PR 생성 전담

## 💡 사용 예시

```bash
/alfred:1-spec                      # 프로젝트 문서 기반 자동 제안
/alfred:1-spec "JWT 인증 시스템"       # 단일 SPEC 수동 생성
/alfred:1-spec SPEC-001 "보안 보강"   # 기존 SPEC 보완
```

## 🔍 STEP 1: SPEC 분석 및 구현 계획 수립

프로젝트 문서를 분석하여 SPEC 후보를 제안하고 구현 전략을 수립한 후 사용자 확인을 받습니다.

### SPEC 분석 진행

1. **프로젝트 문서 분석**
   - product/structure/tech.md 심층 분석
   - 기존 SPEC 목록 및 우선순위 검토
   - 구현 가능성 및 복잡도 평가

2. **SPEC 후보 발굴**
   - 핵심 비즈니스 요구사항 추출
   - 기술적 제약사항 반영
   - 우선순위별 SPEC 후보 리스트 생성

3. **구현 계획 보고**
   - 단계별 SPEC 작성 계획 제시
   - 예상 작업 범위 및 의존성 분석
   - EARS 구조 및 Acceptance Criteria 설계

### 사용자 확인 단계

구현 계획 검토 후 다음 중 선택하세요:
- **"진행"** 또는 **"시작"**: 계획대로 SPEC 작성 시작
- **"수정 [내용]"**: 계획 수정 요청
- **"중단"**: SPEC 작성 중단

---

## 🚀 STEP 2: SPEC 문서 작성 실행 (사용자 승인 후)

사용자 승인 후 spec-builder 에이전트가 **EARS 방식의 구조화된 명세서 작성**과 **모드별 브랜치/PR 생성**을 수행합니다.

## 기능

- **ULTRATHINK**: `.moai/project/{product,structure,tech}.md`를 분석해 구현 후보를 제안하고 사용자 승인 후 SPEC을 생성합니다.
- **Personal 모드**: `.moai/specs/SPEC-XXX/` 디렉터리와 템플릿 문서를 만듭니다.
- **Team 모드**: GitHub Issue(또는 Discussion)를 생성하고 브랜치 템플릿과 연결합니다.

## 사용법

```bash
/alfred:1-spec                      # 프로젝트 문서 기반 자동 제안 (권장)
/alfred:1-spec "JWT 인증 시스템"       # 단일 SPEC 수동 생성
/alfred:1-spec SPEC-001 "보안 보강"   # 기존 SPEC 보완
```

입력하지 않으면 Q&A 결과를 기반으로 우선순위 3~5건을 제안하며, 승인한 항목만 실제 SPEC으로 확정됩니다.

## 모드별 처리 요약

| 모드     | 산출물                                                               | 추가 작업                                       |
| -------- | -------------------------------------------------------------------- | ----------------------------------------------- |
| Personal | `.moai/specs/SPEC-XXX/spec.md`, `plan.md`, `acceptance.md` 등 템플릿 | git-manager 에이전트가 자동으로 체크포인트 생성 |
| Team     | GitHub Issue(`[SPEC-XXX] 제목`), Draft PR(옵션)                      | `gh` CLI 로그인 유지, 라벨/담당자 지정 안내     |

## 입력 옵션

- **자동 제안**: `/alfred:1-spec` → 프로젝트 문서 핵심 bullet을 기반으로 후보 리스트 작성
- **수동 생성**: 제목을 인수로 전달 → 1건만 생성, Acceptance 템플릿은 회신 후 보완
- **보완 모드**: `SPEC-ID "메모"` 형식으로 전달 → 기존 SPEC 문서/Issue를 업데이트

## 📋 STEP 1 실행 가이드: SPEC 분석 및 계획 수립

### 1. 프로젝트 문서 분석

다음을 우선적으로 실행하여 SPEC 후보를 분석합니다:

```bash
# 프로젝트 문서 기반 SPEC 분석
@agent-spec-builder "$ARGUMENTS 분석 및 SPEC 계획 수립"
```

#### 분석 체크리스트

- [ ] **요구사항 추출**: product.md의 핵심 비즈니스 요구사항 파악
- [ ] **아키텍처 제약**: structure.md의 시스템 설계 제약사항 확인
- [ ] **기술적 제약**: tech.md의 기술 스택 및 품질 정책
- [ ] **기존 SPEC**: 현재 SPEC 목록 및 우선순위 검토

### 2. SPEC 후보 발굴 전략

#### 우선순위 결정 기준

| 우선순위 | 기준 | SPEC 후보 유형 |
|---------|------|----------------|
| **높음** | 핵심 비즈니스 가치 | 사용자 핵심 기능, API 설계 |
| **중간** | 시스템 안정성 | 인증/보안, 데이터 관리 |
| **낮음** | 개선 및 확장 | UI/UX 개선, 성능 최적화 |

#### SPEC 타입별 접근법

- **API/백엔드**: 엔드포인트 설계, 데이터 모델, 인증
- **프론트엔드**: 사용자 인터페이스, 상태 관리, 라우팅
- **인프라**: 배포, 모니터링, 보안 정책
- **품질**: 테스트 전략, 성능 기준, 문서화

### 3. SPEC 작성 계획 보고서 생성

다음 형식으로 계획을 제시합니다:

```
## SPEC 작성 계획 보고서: [TARGET]

### 📊 분석 결과
- **발굴된 SPEC 후보**: [개수 및 카테고리]
- **우선순위 높음**: [핵심 SPEC 목록]
- **예상 작업시간**: [시간 산정]

### 🎯 작성 전략
- **선택된 SPEC**: [작성할 SPEC ID 및 제목]
- **EARS 구조**: [Event-Action-Response-State 설계]
- **Acceptance Criteria**: [Given-When-Then 시나리오]

### 🚨 주의사항
- **기술적 제약**: [고려해야 할 제약사항]
- **의존성**: [다른 SPEC과의 연관성]
- **브랜치 전략**: [Personal/Team 모드별 처리]

### ✅ 예상 산출물
- **spec.md**: [EARS 구조의 핵심 명세]
- **plan.md**: [구현 계획서]
- **acceptance.md**: [인수 기준]
- **브랜치/PR**: [모드별 Git 작업]

---
**승인 요청**: 위 계획으로 SPEC 작성을 진행하시겠습니까?
("진행", "수정 [내용]", "중단" 중 선택)
```

---

## 🚀 STEP 2 실행 가이드: SPEC 작성 (승인 후)

사용자가 **"진행"** 또는 **"시작"**을 선택한 경우에만 다음을 실행합니다:

```bash
# SPEC 문서 작성 시작
@agent-spec-builder "$ARGUMENTS SPEC 문서 작성 시작 (사용자 승인 완료)"
```

### EARS 명세 작성 가이드

1. **Event**: 시스템에 발생하는 트리거 이벤트 정의
2. **Action**: 이벤트에 대한 시스템의 행동 명세
3. **Response**: 행동의 결과로 나타나는 응답 정의
4. **State**: 시스템 상태 변화 및 부작용 명시

### 📄 SPEC 문서 템플릿

#### YAML Front Matter 스키마

**spec.md 파일 상단에 반드시 포함**해야 하는 메타데이터:

```yaml
---
id: AUTH-001                    # 필수: <도메인>-<3자리> 형식 (영구 불변)
version: 1.0.0                  # 필수: SemVer 형식 (major.minor.patch)
status: draft                   # 필수: draft|active|deprecated
created: 2025-09-15            # 필수: 생성일 (YYYY-MM-DD)
updated: 2025-10-01            # 필수: 최종 수정일 (YYYY-MM-DD)
author: @username              # 선택: 작성자 GitHub ID
reviewers:                     # 선택: 검토자 목록
  - @reviewer1
  - @reviewer2
labels:                        # 선택: 분류 태그
  - authentication
  - security
  - api
priority: high                 # 선택: high|medium|low
---
```

**필드 설명**:
- **id**: TAG ID와 동일 (`<도메인>-<3자리>`) - 생성 후 절대 변경 금지
- **version**: 명세 버전 (HISTORY 섹션과 동기화 필수)
- **status**:
  - `draft` - 작성 중
  - `active` - 승인됨, 구현 진행/완료
  - `deprecated` - 더 이상 사용 안 함
- **created/updated**: ISO 8601 날짜 형식
- **author**: 최초 작성자 (GitHub ID 권장)
- **reviewers**: 검토자 목록 (Team 모드에서 활용)
- **labels**: 검색 및 분류용 태그
- **priority**: 구현 우선순위 (시간 예측 금지)

#### HISTORY 섹션 (필수)

**YAML Front Matter 직후**에 반드시 HISTORY 섹션을 포함해야 합니다:

```markdown
# @SPEC:AUTH-001: JWT 기반 인증 시스템

## HISTORY

### v1.0.0 (2025-09-15)
- **INITIAL**: 기본 JWT 인증 명세 작성
- **AUTHOR**: @goos
- **SCOPE**:
  - 로그인/로그아웃 기본 플로우
  - 액세스 토큰 발급/검증
  - 리프레시 토큰 메커니즘

### v1.1.0 (2025-09-20)
- **ADDED**: 소셜 로그인 요구사항 추가
- **AUTHOR**: @goos
- **REVIEW**: @security-team (승인)
- **CHANGES**:
  - OAuth2 통합 요구사항
  - Google/GitHub 로그인 지원

### v2.0.0 (2025-10-01)
- **BREAKING**: 토큰 만료 시간 정책 변경
- **CHANGED**: 액세스 토큰 15분 → 30분
- **ADDED**: 자동 토큰 갱신 요구사항
- **AUTHOR**: @goos
- **REVIEW**: @product-team, @security-team (승인)
- **MIGRATION**: 기존 클라이언트 토큰 갱신 로직 업데이트 필요
```

**HISTORY 작성 규칙**:
- **버전 순서**: 최신 버전이 위로 (역순)
- **변경 타입** (대문자):
  - `INITIAL` - 최초 작성
  - `ADDED` - 새 요구사항 추가
  - `CHANGED` - 기존 요구사항 수정
  - `BREAKING` - 하위 호환성 깨는 변경
  - `DEPRECATED` - 폐기 예정 표시
  - `REMOVED` - 요구사항 제거
  - `FIXED` - 명세 오류 수정
- **필수 항목**: 버전, 날짜, AUTHOR, 변경 내용
- **선택 항목**: REVIEW (검토자), SCOPE (범위), MIGRATION (마이그레이션 가이드)

#### SPEC 문서 전체 구조

```markdown
---
id: AUTH-001
version: 1.0.0
status: draft
created: 2025-09-15
updated: 2025-09-15
author: @username
---

# @SPEC:AUTH-001: [SPEC 제목]

## HISTORY
[버전별 변경 이력 - 위 예시 참조]

## Environment (환경)
[시스템 환경 및 전제 조건]

## Assumptions (가정)
[설계 가정 사항]

## Requirements (요구사항)
### Ubiquitous (필수 기능)
- 시스템은 [기능]을 제공해야 한다

### Event-driven (이벤트 기반)
- WHEN [조건]이면, 시스템은 [동작]해야 한다

### State-driven (상태 기반)
- WHILE [상태]일 때, 시스템은 [동작]해야 한다

### Optional (선택 기능)
- WHERE [조건]이면, 시스템은 [동작]할 수 있다

### Constraints (제약사항)
- IF [조건]이면, 시스템은 [제약]해야 한다

## Traceability (@TAG)
- **SPEC**: @SPEC:AUTH-001
- **TEST**: tests/auth/test_service.py
- **CODE**: src/auth/service.py
- **DOC**: docs/api/authentication.md
```

### 에이전트 협업 구조

- **1단계**: `spec-builder` 에이전트가 프로젝트 문서 분석 및 SPEC 문서 작성을 전담합니다.
- **2단계**: `git-manager` 에이전트가 브랜치 생성, GitHub Issue/PR 생성을 전담합니다.
- **단일 책임 원칙**: spec-builder는 SPEC 작성만, git-manager는 Git/GitHub 작업만 수행합니다.
- **순차 실행**: spec-builder → git-manager 순서로 실행하여 명확한 의존성을 유지합니다.
- **에이전트 간 호출 금지**: 각 에이전트는 다른 에이전트를 직접 호출하지 않고, 커맨드 레벨에서만 순차 실행합니다.

## 🚀 최적화된 워크플로우 실행 순서

### Phase 1: 병렬 프로젝트 분석 (성능 최적화)

**동시에 수행**:

```
Task 1 (haiku): 프로젝트 구조 스캔
├── 언어/프레임워크 감지
├── 기존 SPEC 목록 수집
└── 우선순위 백로그 초안

Task 2 (sonnet): 심화 문서 분석
├── product.md 요구사항 추출
├── structure.md 아키텍처 분석
└── tech.md 기술적 제약사항
```

**성능 향상**: 기본 스캔과 심화 분석을 병렬 처리하여 대기 시간 최소화

### Phase 2: SPEC 문서 통합 작성

`spec-builder` 에이전트(sonnet)가 병렬 분석 결과를 통합하여:

- 프로젝트 문서 기반 기능 후보 제안
- 사용자 승인 후 SPEC 문서 작성 (MultiEdit 활용)
- 3개 파일 동시 생성 (spec.md, plan.md, acceptance.md)

### Phase 3: Git 작업 처리

`git-manager` 에이전트(haiku)가 최종 처리:

- **브랜치 생성**: 모드별 전략(Personal/Team) 적용
- **GitHub Issue 생성**: Team 모드에서 SPEC Issue 생성
- **초기 커밋**: SPEC 문서 커밋 및 태그 생성

**중요**: 각 에이전트는 독립적으로 실행되며, 에이전트 간 직접 호출은 금지됩니다.

## 에이전트 역할 분리

### spec-builder 전담 영역

- 프로젝트 문서 분석 및 SPEC 후보 발굴
- EARS 구조의 명세서 작성
- Acceptance Criteria 작성 (Given-When-Then)
- SPEC 문서 품질 검증
- @TAG 시스템 적용

### git-manager 전담 영역

- 모든 Git 브랜치 생성 및 관리
- 모드별 브랜치 전략 적용
- GitHub Issue/PR 생성
- 초기 커밋 및 태그 생성
- 원격 동기화 처리

## 2단계 워크플로우 실행 순서

### Phase 1: 분석 및 계획 단계

**SPEC 분석기**가 다음을 수행:

1. **프로젝트 문서 로딩**: product/structure/tech.md 심층 분석
2. **SPEC 후보 발굴**: 비즈니스 요구사항 기반 우선순위 결정
3. **구현 전략 수립**: EARS 구조 및 Acceptance 설계
4. **작성 계획 생성**: 단계별 SPEC 작성 접근 방식 제시
5. **사용자 승인 대기**: 계획 검토 및 피드백 수집

### Phase 2: SPEC 작성 단계 (승인 후)

`spec-builder` 에이전트가 사용자 승인 후 **연속적으로** 수행:

1. **EARS 명세 작성**: Event-Action-Response-State 구조화
2. **Acceptance Criteria**: Given-When-Then 시나리오 작성
3. **문서 품질 검증**: TRUST 원칙 및 @TAG 적용
4. **템플릿 생성**: spec.md, plan.md, acceptance.md 동시 생성

### Phase 3: Git 작업 (git-manager)

`git-manager` 에이전트가 SPEC 완료 후 **한 번에** 수행:

1. **브랜치 생성**: 모드별 브랜치 전략 적용
2. **GitHub Issue**: Team 모드에서 SPEC Issue 생성
3. **초기 커밋**: SPEC 문서 커밋 및 태그 생성
4. **원격 동기화**: 모드별 동기화 전략 적용

## 작성 팁

- product/structure/tech 문서에 없는 정보는 새로 질문해 보완합니다.
- Acceptance Criteria는 Given/When/Then 3단으로 최소 2개 이상 작성하도록 유도합니다.
- TRUST 원칙 중 Readable(읽기 쉬움) 기준 완화로 인해 모듈 수가 권장치(기본 5)를 초과하는 경우, 근거를 SPEC `context` 섹션에 함께 기록하세요.

---

## 🧠 Context Management (컨텍스트 관리)

> 본 커맨드는 **컨텍스트 엔지니어링** 원칙을 따릅니다.
> **컨텍스트 예산/토큰 예산은 다루지 않습니다**.

### JIT Retrieval (필요 시 로딩)

**우선 로드** (SPEC 작성 시작 시):
- `.moai/project/product.md` - 비즈니스 요구사항 및 사용자 스토리

**필요 시 로드** (SPEC 후보 발굴 시):
- `.moai/project/structure.md` - 시스템 아키텍처 및 모듈 설계
- `.moai/project/tech.md` - 기술 스택 및 품질 게이트

**지연 로드** (기존 SPEC 확인 필요 시):
- `.moai/specs/` - 기존 SPEC 목록 및 의존성 분석

### Compaction 권장 시점

**트리거 조건**:
- SPEC 작성 완료 후 다음 단계(2-build) 진행 전
- 토큰 사용량 > 70% (140,000 / 200,000)

**권장 메시지**:
```markdown
**권장사항**: SPEC 작성이 완료되었습니다. 다음 단계(`/alfred:2-build`) 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.
```

### Structured Memory 활용

**중요한 SPEC 결정 기록**:
```bash
# 의사결정 로그
.moai/memory/decisions/2025-10-02-spec-001-architecture-choice.md
```

**제약사항 문서화**:
```bash
# 기술적 제약사항
.moai/memory/constraints/spec-001-performance-constraints.md
```

**리스크 관리**:
```bash
# 식별된 리스크
.moai/memory/risks/spec-001-integration-risk.md
```

---

## 다음 단계

**권장사항**: 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.

- `/alfred:2-build SPEC-XXX`로 TDD 구현 시작
- 팀 모드: Issue 생성 후 git-manager 에이전트가 자동으로 브랜치 생성
