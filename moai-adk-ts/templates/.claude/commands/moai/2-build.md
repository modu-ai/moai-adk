---
name: moai:2-build
description: TDD 구현 (Red-Green-Refactor) + TAG 체인 연결 + Git 자동화
argument-hint: "SPEC-ID | all"
tools: Read, Write, Edit, MultiEdit, Bash, Task, WebFetch, Grep, Glob, TodoWrite
---

# MoAI-ADK 2단계: 4단계 TDD 구현 워크플로우

**TDD 구현 대상**: ${ARGUMENTS:-"모든 SPEC"}

## 🚀 4단계 최적화 워크플로우

프로젝트 문서를 분석하여 TDD 구현 계획을 수립하고, 승인된 계획으로 Red-Green-Refactor 사이클을 실행합니다.

### 핵심 기능

- **스마트 분석**: SPEC 문서 분석 및 구현 계획 자동 수립
- **TDD 구현**: 승인된 계획 기반 Red-Green-Refactor 사이클
- **TAG 연결**: @TEST TAG 자동 생성 및 Primary Chain 연결
- **Git 자동화**: TDD 단계별 구조화된 커밋

## 실행 순서

### Phase 1: 분석 및 계획 (명령어 레벨)

SPEC 문서 분석 및 구현 계획을 수립합니다:

```bash
# 1. SPEC 문서 로딩 및 분석
# - 요구사항 추출 및 복잡도 평가
# - 기술적 제약사항 확인
# - 의존성 및 영향 범위 분석

# 2. 구현 전략 결정
# - 프로젝트 언어 감지 및 최적화 전략
# - TDD 접근 방식 결정
# - 예상 작업 범위 및 시간 산정

# 3. 구현 계획 보고서 생성
# - 단계별 구현 계획 제시
# - 잠재적 위험 요소 식별
# - 품질 게이트 체크포인트 설정

# 4. 사용자 승인 요청
# "진행" 또는 "시작": TDD 구현 시작
# "수정 [내용]": 계획 수정 요청
# "중단": 구현 작업 중단
```

### Phase 2: TDD 구현 (code-builder 전담)

사용자 승인 후 code-builder 에이전트로 순수한 TDD 구현을 수행합니다:

@agent-code-builder "승인된 계획: ${ARGUMENTS:-"모든 SPEC"}의 TDD 구현을 Red-Green-Refactor 사이클로 수행해주세요"

- **언어별 최적화**: 프로젝트 언어 감지 후 최적 도구 선택
- **RED 단계**: 실패하는 테스트 작성 및 확인
- **GREEN 단계**: 최소 구현으로 테스트 통과
- **REFACTOR 단계**: 코드 품질 개선 및 TRUST 원칙 검증
- **품질 보장**: 커버리지 85% 이상, 린터/포매터 통과

### Phase 3: TAG 시스템 관리 (tag-agent 전담)

TDD 구현 후 tag-agent가 @TEST TAG 체인 생성 및 연결 작업을 수행합니다:

@agent-tag-agent "TDD 완료된 ${ARGUMENTS}의 @TEST TAG 체인 생성하고 @TASK와 연결, 인덱스 업데이트를 수행해주세요"

- **@TEST TAG 생성**: 생성된 테스트 파일 기반 TAG 생성
- **Primary Chain 연결**: @TASK:SPEC-XXX → @TEST:SPEC-XXX
- **체인 검증**: 체인 무결성 및 순환 참조 방지
- **인덱스 업데이트**: JSONL 기반 분산 인덱스 갱신

### Phase 4: Git 작업 자동화 (git-manager 전담)

마지막으로 TDD와 TAG 작업을 포함한 Git 자동화를 수행합니다:

@agent-git-manager "TDD 및 TAG 작업 완료된 ${ARGUMENTS}의 구조화된 커밋(RED-GREEN-REFACTOR)과 브랜치 동기화를 수행해주세요"

- **구조화된 커밋**: TDD 단계별 커밋 (🔴RED, 🟢GREEN, ♻️REFACTOR)
- **TAG 정보 포함**: 커밋 메시지에 TAG 체인 정보 자동 삽입
- **브랜치 동기화**: Personal/Team 모드별 Git 전략 적용
- **체크포인트 생성**: TDD 완료 상태 백업

## 품질 기준

- **TDD 사이클**: Red-Green-Refactor 완전 준수
- **에이전트 역할 분리**: 각 에이전트 고유 책임 영역 100% 준수
- **TRUST 원칙**: Test First, Readable, Unified, Secured, Trackable 검증
- **TAG 위임 완료**:
  - code-builder: TDD 구현 완성 (tag-agent 위임)
  - tag-agent: @TEST TAG 생성, 체인 관리, 인덱스 업데이트 독점 처리
  - git-manager: TDD 커밋 및 브랜치 동기화 전담
  - 중복 작업 0건 (각 에이전트 단일 책임)
  - 오케스트레이션 품질: 에이전트 간 데이터 전달 무결성

## 📋 TDD 도구 매핑 및 성능 목표

### 언어별 최적 라우팅 전략

| SPEC 타입 | 최적 언어 | 테스트 프레임워크 | 성능 목표 | 커버리지 목표 |
|-----------|-----------|-------------------|-----------|---------------|
| **CLI/시스템** | TypeScript | Vitest + tsx | < 50ms | 95%+ |
| **API/백엔드** | TypeScript/Go | Vitest/go test | < 150ms | 90%+ |
| **프론트엔드** | TypeScript | Vitest + Testing Library | < 100ms | 85%+ |
| **데이터 처리** | Python/TypeScript | pytest/Vitest | < 500ms | 85%+ |
| **범용** | 프로젝트 언어 감지 | 언어별 최적 도구 | 언어별 최적화 | 85%+ |

### 자동 언어 감지 및 도구 선택

```bash
# JavaScript/TypeScript 프로젝트
test=npm test, lint=eslint, format=prettier

# Python 프로젝트
test=pytest, lint=ruff, format=black

# Go 프로젝트
test=go test, lint=golint, format=gofmt

# 멀티 언어 프로젝트
자동 감지 후 주 언어 기준 도구 선택
```


## 📋 Phase 1: 분석 및 계획 실행 가이드

명령어 레벨에서 직접 수행하는 SPEC 분석 및 구현 계획 수립:

### 1. SPEC 문서 분석

```bash
# SPEC 문서 로딩 및 분석
# .moai/specs/${SPEC_ID}/ 디렉터리 또는 개별 SPEC 파일 확인
```

#### 분석 체크리스트

- [ ] **SPEC 문서 존재 확인**: spec.md, plan.md, acceptance.md
- [ ] **EARS 방법론 준수**: 5가지 구문 형식 완성도 확인
- [ ] **요구사항 명확성**: 기능/비기능 요구사항 구체성
- [ ] **의존성 분석**: 기존 SPEC과의 연관성 및 영향 범위
- [ ] **복잡도 평가**: 구현 난이도 (낮음/중간/높음)

### 2. 구현 전략 결정

#### 언어별 구현 기준

```typescript
// 프로젝트 언어 감지 로직
interface ImplementationStrategy {
  spec_id: string;
  complexity: 'low' | 'medium' | 'high';
  language: string;
  test_framework: string;
  approach: 'bottom-up' | 'top-down' | 'middle-out';
  estimated_time: string;
}
```

### 3. 구현 계획 보고서 생성

다음 형식으로 계획을 제시하고 사용자 승인을 받습니다:

```markdown
## 🔍 TDD 구현 계획 보고서: [SPEC-ID]

### 📊 분석 결과
- **복잡도**: [낮음/중간/높음] - [상세 근거]
- **예상 작업시간**: [N시간] - [산정 근거]
- **주요 기술 도전**: [구체적 어려움 3가지]

### 🎯 구현 전략
- **선택 언어**: [감지된 언어] - [선택 이유]
- **TDD 접근법**: [Bottom-up/Top-down/Middle-out] - [근거]
- **핵심 모듈**: [구현할 주요 모듈 목록]

### 🚨 위험 요소
- **기술적 위험**: [예상 문제점과 대응 방안]
- **의존성 위험**: [외부 라이브러리 이슈]
- **일정 위험**: [지연 가능성과 완화 방안]

### ✅ 품질 게이트
- **테스트 커버리지**: [목표 %] - [측정 방법]
- **성능 목표**: [구체적 지표] - [검증 방법]
- **보안 체크포인트**: [검증할 보안 항목]

---
**🔔 승인 요청**: 위 계획으로 TDD 구현을 진행하시겠습니까?

다음 중 하나를 선택해 주세요:
- **"진행"** 또는 **"시작"**: 계획대로 Phase 2 TDD 구현 시작
- **"수정 [구체적 변경사항]"**: 계획 수정 후 재검토
- **"중단"**: 구현 작업 중단
```

## 📋 에이전트 역할 분리 및 데이터 전달

### code-builder 전담 영역 (Phase 2)

- **TDD 구현**: Red-Green-Refactor 사이클 순수 실행
- **테스트 작성**: 언어별 최적 테스트 프레임워크 활용
- **품질 보장**: TRUST 5원칙 검증, 커버리지 85% 이상
- **코드 품질**: 린터, 포매터, 타입 체킹 통과

### tag-agent 전담 영역 (Phase 3)

- **@TEST TAG 생성**: 생성된 테스트 파일 기반 TAG 생성
- **Primary Chain 연결**: @TASK → @TEST 체인 무결성 보장
- **중복 방지**: 기존 TAG 검색 및 재사용 검토
- **인덱스 관리**: JSONL 기반 분산 인덱스 업데이트

### git-manager 전담 영역 (Phase 4)

- **구조화된 커밋**: TDD 단계별 커밋 (🔴RED, 🟢GREEN, ♻️REFACTOR)
- **TAG 정보 포함**: 커밋 메시지에 TAG 체인 정보 자동 삽입
- **브랜치 동기화**: Personal/Team 모드별 Git 전략 적용
- **체크포인트 생성**: TDD 완료 상태 백업 포인트

### 데이터 전달 인터페이스

```typescript
// Phase 1 → Phase 2: 승인된 구현 계획
interface ApprovedPlan {
  spec_id: string;
  complexity: 'low' | 'medium' | 'high';
  language: string;
  strategy: string;
  approved: boolean;
}

// Phase 2 → Phase 3: TDD 구현 결과
interface TDDResult {
  spec_id: string;
  test_files: string[];
  source_files: string[];
  tdd_phases: ['RED', 'GREEN', 'REFACTOR'];
  coverage_achieved: number;
}

// Phase 3 → Phase 4: TAG 체인 정보
interface TagChainResult {
  spec_id: string;
  created_test_tags: string[];
  chain_connection: string; // "@TASK:XXX → @TEST:XXX"
  related_tags: string[];
}
```

## 📊 품질 게이트 체크리스트

### Phase 2 완료 기준 (code-builder)
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 테스트 통과 (RED → GREEN → REFACTOR)
- [ ] 린터/포매터 통과
- [ ] TRUST 5원칙 검증 완료

### Phase 3 완료 기준 (tag-agent)
- [ ] @TEST TAG 생성 완료
- [ ] Primary Chain 연결 검증
- [ ] TAG 중복 없음
- [ ] JSONL 인덱스 업데이트 완료

### Phase 4 완료 기준 (git-manager)
- [ ] TDD 단계별 커밋 완료
- [ ] TAG 정보 포함 확인
- [ ] 브랜치 동기화 성공
- [ ] 체크포인트 생성 완료

## 🔄 다음 단계

- **TDD 및 TAG 연결 완료 후**: `/moai:3-sync`로 문서 동기화 진행
- **Living Document 업데이트**: TAG 체인 정보 문서 반영
- **에이전트 독립성 보장**: 각 Phase별 완전한 작업 격리
- **오류 복구**: Phase별 체크포인트로 정확한 상태 복구 가능
