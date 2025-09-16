---
description: Constitution Check 및 기술 계획 수립 - 완성된 EARS 명세를 바탕으로 체계적 설계와 아키텍처 결정
argument-hint: <spec-id> [--phase=0|1|2] [--force-check]
allowed-tools: Read, Write, Edit, WebFetch, Task
---

# MoAI-ADK Constitution Check & 기술 계획

완성된 EARS 명세를 바탕으로 기술적 설계와 아키텍처 결정을 수행하는 체계적 계획 수립 단계입니다. MoAI-ADK의 Constitution Check를 통해 프로젝트의 복잡도와 적절성을 평가하고, 5가지 핵심 원칙에 따라 지속 가능성과 유지보수성을 미리 보장합니다.

## 🎯 실행 플로우

```mermaid
flowchart TD
    A[SPEC 문서 로드] --> B{[NEEDS CLARIFICATION] 검증}
    B -->|존재| C[⚠️ 명세 불완전]
    B -->|없음| D[🏛️ Constitution Check]
    
    D --> E[📋 Simplicity Check]
    E --> F[🏗️ Architecture Check]
    F --> G[🧪 Testing Check]
    G --> H[📊 Observability Check]
    H --> I[📦 Versioning Check]
    
    I --> J{Constitution 통과?}
    J -->|실패| K[🔴 위반 정당화 또는 수정]
    J -->|성공| L[📚 Phase 0: research.md]
    
    L --> M[📋 Phase 1: contracts/, data-model.md]
    M --> N[🛑 Phase 2 설명 (tasks.md 생성 안함)]
    N --> O[✅ /moai:4-tasks 명령 대기]
```

## 🤖 자연어 체이닝 오케스트레이션

🤖 **Constitution Check 및 기술 계획 수립을 전문 에이전트 체인으로 완전 자동화합니다.**

**Constitution 검증 단계**: Task tool을 사용하여 constitution-checker 에이전트를 호출하여 5가지 핵심 원칙(Simplicity, Architecture, Testing, Observability, Versioning)에 대한 체계적 검증을 수행합니다.

**기술 조사 단계**: Constitution Check 통과 시 Task tool을 사용하여 research-analyst 에이전트를 호출하여 WebFetch 기반 최신 기술 동향 조사와 아키텍처 패턴 분석을 수행합니다.

**설계 단계**: Task tool을 사용하여 architecture-designer 에이전트를 호출하여 contracts/ 디렉토리 생성, API 계약 정의, data-model.md 작성을 통한 체계적 아키텍처 설계를 수행합니다.

## 🏛️ Constitution Check - 5개 핵심 원칙

### 1. Simplicity (단순성) - 복잡도 제어
```markdown
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔍 Simplicity Check - 프로젝트 ≤ 3개
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

검증 항목:
✓ 독립적 모듈/프로젝트 수 분석
✓ 마이크로서비스 개수 확인
✓ 복잡도 점수 계산 (McCabe Complexity 기반)

결과:
- 현재 복잡도: [자동 계산]
- 임계값: 3개 프로젝트
- 상태: ✅ 통과 / ❌ 위반

위반 시 해결책:
1. 모듈 통합을 통한 복잡도 감소
2. 라이브러리 분리를 통한 재사용성 확보
3. 의존성 역전을 통한 결합도 감소
```

### 2. Architecture (아키텍처) - 라이브러리화 강제
```markdown
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🏗️ Architecture Check - 모든 기능 라이브러리화
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

검증 항목:
✓ 모든 기능이 독립적 라이브러리로 분리 가능한가?
✓ 순환 의존성이 없는가?
✓ 인터페이스 기반 설계가 되어 있는가?

결과:
- 라이브러리화 가능성: [자동 분석]
- 의존성 그래프: [시각화]
- 상태: ✅ 통과 / ❌ 위반

위반 시 해결책:
1. 의존성 역전 원칙 적용
2. 인터페이스 분리를 통한 모듈화
3. 공통 기능의 별도 라이브러리 분리
```

### 3. Testing (테스팅) - TDD 강제
```markdown
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🧪 Testing Check - RED-GREEN-Refactor 강제
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

검증 항목:
✓ 모든 기능에 테스트 우선 작성 가능한가?
✓ 테스트 커버리지 목표 설정 (기본 80%)
✓ 테스트 자동화 파이프라인 구성

결과:
- TDD 적용 가능성: [분석]
- 예상 커버리지: [계산]
- 상태: ✅ 통과 / ❌ 위반

강제 적용:
1. 모든 /moai:5-dev 명령에서 TDD 사이클 강제
2. Hook 시스템을 통한 테스트 없는 구현 차단
3. 커버리지 임계값 미달 시 자동 차단
```

### 4. Observability (관찰 가능성) - 구조화된 로깅
```markdown
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Observability Check - 구조화된 로깅 강제
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

검증 항목:
✓ 모든 주요 기능에 로깅 포인트 설계
✓ 구조화된 로그 형식 (JSON, OpenTelemetry)
✓ 메트릭과 추적 시스템 통합

결과:
- 로깅 설계 완성도: [분석]
- 관측점 커버리지: [계산]
- 상태: ✅ 통과 / ❌ 위반

자동 적용:
1. 모든 API 엔드포인트에 자동 로깅
2. 에러 추적 및 성능 모니터링 활성화
3. 비즈니스 메트릭 자동 수집
```

### 5. Versioning (버전 관리) - MAJOR.MINOR.BUILD
```markdown
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📦 Versioning Check - MAJOR.MINOR.BUILD 강제
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

검증 항목:
✓ 시맨틱 버저닝 규칙 준수
✓ 자동 버전 관리 시스템 구성
✓ 하위 호환성 보장 체계

결과:
- 버전 관리 전략: [자동 설정]
- 릴리스 자동화: [구성]
- 상태: ✅ 통과 / ❌ 위반

자동 적용:
1. Git tag 기반 자동 버전 관리
2. 브레이킹 체인지 감지 및 MAJOR 버전 자동 증가
3. 자동 CHANGELOG 생성
```

## 📚 Phase 0: Research 생성

Constitution Check 통과 후 기술 조사를 수행합니다:

```markdown
# research.md 자동 생성

## WebFetch 기반 기술 조사

### 선택된 기술 스택 최신 동향
- [기술명]: 최신 버전, 보안 이슈, 성능 개선사항
- [프레임워크]: 최신 패턴, 모범 사례, 마이그레이션 가이드

### 아키텍처 패턴 분석
- [선택한 패턴]: 구현 방법, 장단점, 팀 숙련도 고려사항
- 대안 패턴: 비교 분석 및 선택 근거

### 보안 및 성능 고려사항
- 보안 위협 모델링
- 성능 최적화 전략
- 운영 환경 요구사항
```

## 📋 Phase 1: Contracts & Data Model

```markdown
# contracts/ 디렉토리 자동 생성

## API 계약 정의
- /api/users/: OpenAPI 3.0 스펙
- /api/auth/: 인증/권한 계약
- /api/[domain]/: 도메인별 API 계약

## data-model.md 생성

### 엔티티 설계
- User: 사용자 모델
- [Domain]: 비즈니스 도메인 모델
- Relationship: 관계 정의

### 데이터 흐름 설계
- 입력 검증 규칙
- 비즈니스 로직 흐름
- 출력 변환 규칙
```

## 📋 Phase 2: 작업 분해 자동 진행

Constitution Check 통과 후 자동으로 작업 분해를 진행합니다:

```markdown
🤖 Phase 2: 작업 분해 자동 실행

Task tool을 사용하여 task-decomposer 에이전트를 호출하여:
- 완성된 계획을 실제 구현 가능한 작업 단위로 분해
- TDD 순서에 맞는 Red-Green-Refactor 작업 순서 결정
- 병렬 실행 가능한 작업을 [P] 마커로 식별
- 의존성 그래프를 최적화하여 개발 속도 극대화

자동 생성:
- tasks.md: 상세 작업 분해 문서
- dependency-graph.md: 작업 간 의존성 시각화
- parallel-waves.md: 병렬 실행 계획

이렇게 계획과 작업 분해를 연속적으로 처리하여:
- 사용자의 추가 명령어 입력 불필요
- 컨텍스트 연속성 보장으로 더 일관된 결과
- 전체 워크플로우 시간 단축
```

## ⚠️ 에러 처리

### SPEC 문서 없음
```markdown
❌ ERROR: SPEC 문서를 찾을 수 없습니다.

먼저 다음 명령으로 SPEC을 작성해주세요:
> /moai:2-spec [feature-name] "기능 설명"

SPEC 문서 경로: .moai/specs/SPEC-XXX/spec.md
```

### [NEEDS CLARIFICATION] 마커 감지
```markdown
⚠️ WARNING: SPEC에 불명확한 부분이 있습니다.

다음 항목들을 먼저 명확히 해주세요:
- [NEEDS CLARIFICATION: 사용자 권한 체계]
- [NEEDS CLARIFICATION: 데이터 보관 정책]

SPEC을 완성한 후 다시 실행해주세요.
```

### Constitution 위반
```markdown
🔴 CONSTITUTION 위반 감지

위반 내용:
- Simplicity: 4개 독립 모듈 (허용: 3개)
- Testing: TDD 적용 불가능한 레거시 코드 존재

해결 방안:
1. 위반을 정당화하고 계속 진행 [y/N]
2. 설계를 수정하여 Constitution 준수 [권장]
3. Constitution 규칙을 프로젝트에 맞게 조정

선택: _
```

## 🎯 사용 예시

### 기본 사용법
```bash
# SPEC-001에 대한 계획 수립
> /moai:3-plan SPEC-001

# 특정 Phase만 실행
> /moai:3-plan SPEC-001 --phase=0  # research만
> /moai:3-plan SPEC-001 --phase=1  # contracts만

# Constitution Check 강제 실행
> /moai:3-plan SPEC-001 --force-check
```

### 고급 사용법
```bash
# 여러 SPEC에 대한 통합 계획
> /moai:3-plan SPEC-001,SPEC-002,SPEC-003

# 특정 Constitution 원칙만 체크
> /moai:3-plan SPEC-001 --check=simplicity,testing
```

## ✅ 완료 시 산출물

```markdown
✅ Constitution Check & 계획 수립이 완료되었습니다!

📊 Constitution Check 결과:
  ✅ Simplicity: 2개 프로젝트 (3개 이하)
  ✅ Architecture: 모든 기능 라이브러리화 가능
  ✅ Testing: TDD 적용 100% 가능
  ✅ Observability: 구조화된 로깅 설계 완료
  ✅ Versioning: MAJOR.MINOR.BUILD 시스템 구성

📁 생성된 파일:
  ├── .moai/specs/SPEC-001/
  │   ├── plan.md           # 구현 계획서
  │   ├── research.md       # 기술 조사 결과
  │   ├── data-model.md     # 데이터 모델 설계
  │   └── contracts/        # API 계약 정의
  │       ├── users.yaml    # 사용자 API
  │       └── auth.yaml     # 인증 API
  └── .moai/memory/decisions/
      └── ADR-001-architecture.md  # 아키텍처 결정 기록

🚀 다음 단계:
  > /moai:4-tasks SPEC-001   # TDD 작업 분해
  > /moai:5-dev T001         # 첫 번째 태스크 구현

💡 Pro Tip: Constitution 원칙은 Hook 시스템이 자동으로 강제합니다.
```

## 🔍 참고 문서

이 명령어는 다음 원칙을 구현합니다:
- **MoAI Constitution**: 5가지 핵심 원칙 자동 검증
- **Claude Code 표준**: 공식 문서 기반 설계 방법론
- **Spec-First**: 명세 기반 체계적 계획 수립
- **TDD-Ready**: 테스트 우선 개발을 위한 사전 준비