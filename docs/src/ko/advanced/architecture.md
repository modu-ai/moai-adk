# 핵심 아키텍처: Alfred 하이브리드 시스템

MoAI-ADK의 심장인 Alfred 슈퍼에이전트의 아키텍처를 깊이 있게 이해하세요.

## 전체 구조

<span class="material-icons">architecture</span> **시스템 아키텍처**

### 4계층 스택

```
┌─────────────────────────────────────────────┐
│           Commands (/alfred:*)              │
│   (Workflow Orchestration & User Entry)     │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│        Sub-agents (19명 팀)                  │
│   (Deep reasoning & Decision making)        │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│       Claude Skills (93개)                   │
│   (Reusable knowledge capsules)             │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│           Hooks (<100ms)                    │
│   (Guardrails & Context)                    │
└─────────────────────────────────────────────┘
```

______________________________________________________________________

## Alfred SuperAgent (슈퍼에이전트)

<span class="material-icons">psychology</span> **중앙 조율 시스템**

Alfred는 **SPEC → TDD → Sync** 워크플로우를 오케스트레이션하는 중앙 조율자입니다.

### 핵심 특성

| 특성          | 설명                                       |
| ------------- | ------------------------------------------ |
| **자율성**    | 사용자 의도를 파악하고 독립적으로 의사결정 |
| **추론 능력** | 복잡한 작업을 단계별로 분해                |
| **팀 조율**   | 19명의 전문가를 최적으로 배치              |
| **학습**      | 세션 로그로부터 지속적 개선                |
| **투명성**    | 모든 결정을 추적 가능하게 기록             |

______________________________________________________________________

## 19명의 팀 구조

<span class="material-icons">groups</span> **전문가 팀 조직**

### 10명의 핵심 Sub-agents

| Agent                      | 역할                     | 활성화 조건          |
| -------------------------- | ------------------------ | -------------------- |
| **project-manager**        | 프로젝트 초기화 및 설정  | `/alfred:0-project`  |
| **spec-builder**           | SPEC 작성 (EARS 문법)    | `/alfred:1-plan`     |
| **implementation-planner** | 아키텍처 및 구현 계획    | `/alfred:2-run` 시작 |
| **tdd-implementer**        | RED→GREEN→REFACTOR 실행  | `/alfred:2-run` 중   |
| **doc-syncer**             | 문서 자동 생성 및 동기화 | `/alfred:3-sync`     |
| **tag-agent**              | TAG 검증 및 추적성 관리  | `/alfred:3-sync`     |
| **git-manager**            | Git 워크플로우 자동화    | 모든 단계            |
| **trust-checker**          | TRUST 5 원칙 검증        | `/alfred:2-run` 완료 |
| **quality-gate**           | 릴리즈 준비 상태 확인    | `/alfred:3-sync`     |
| **debug-helper**           | 오류 분석 및 해결        | 필요시 자동 활성화   |

### 6명의 전문가 Agents

| Expert              | 도메인                        | 활성화 조건              |
| ------------------- | ----------------------------- | ------------------------ |
| **backend-expert**  | API, 서버, DB 아키텍처        | SPEC에 서버/API 키워드   |
| **frontend-expert** | UI, 상태관리, 성능            | SPEC에 프론트엔드 키워드 |
| **devops-expert**   | 배포, CI/CD, 인프라           | SPEC에 배포 키워드       |
| **ui-ux-expert**    | 디자인 시스템, 접근성         | SPEC에 디자인 키워드     |
| **security-expert** | 보안 분석, 취약점 진단        | SPEC에 보안 키워드       |
| **database-expert** | DB 설계, 최적화, 마이그레이션 | SPEC에 DB 키워드         |

### 2명의 빌트인 Agents (Claude)

- **Claude Opus/Sonnet**: 복잡한 추론 필요시
- **Claude Haiku**: 경량 작업용

______________________________________________________________________

## 하이브리드 패턴

<span class="material-icons">hub</span> **협업 패턴**

### Lead-Specialist 패턴

특화된 도메인 전문가가 리드 에이전트를 지원합니다.

```
사용자 요청
    ↓
Alfred (Lead)
    ├─→ 프론트엔드 키워드 감지
    │   └─→ frontend-expert 활성화 (Specialist)
    ├─→ 데이터베이스 키워드 감지
    │   └─→ database-expert 활성화 (Specialist)
    └─→ 보안 키워드 감지
        └─→ security-expert 활성화 (Specialist)
```

**사용 사례**:

- UI 컴포넌트 설계 필요 → UI/UX Expert
- DB 성능 최적화 → Database Expert
- 보안 검토 → Security Expert

### Master-Clone 패턴

대규모 작업은 Alfred 복제본들이 병렬로 처리합니다.

```
대규모 작업 (100+ 파일, 5+ 단계)
    ↓
Master Alfred (조율)
    ├─→ Clone-1: 모듈 A 리팩토링
    ├─→ Clone-2: 모듈 B 리팩토링
    └─→ Clone-3: 모듈 C 리팩토링
    ↓
결과 병합 및 통합
```

**사용 사례**:

- 대규모 마이그레이션 (v1.0 → v2.0)
- 전체 아키텍처 리팩토링
- 다중 도메인 동시 작업

______________________________________________________________________

## :bullseye: 4단계 워크플로우

### Phase 1: 의도 파악 (Intent Understanding)

```
사용자 요청 → 명확성 평가
├─ 명확: Phase 2로 진행
└─ 불명확: AskUserQuestion → 사용자 응답 → Phase 2로 진행
```

**Alfred의 역할**:

- 요청 분석 및 분류
- 필요시 추가 정보 수집
- 작업 범위 확정

### Phase 2: 계획 수립 (Plan Creation)

```
Plan Agent 호출
    ↓
├─ 작업 분해 (Decomposition)
├─ 의존성 분석 (Dependency Analysis)
├─ 병렬화 기회 식별 (Parallelization)
├─ 파일 목록 명시 (File List)
└─ 시간 추정 (Time Estimation)
    ↓
사용자 승인 (AskUserQuestion)
    ↓
TodoWrite 초기화
```

### Phase 3: 작업 실행 (Execution)

```
RED Phase
├─ 테스트 작성
└─ 모두 실패 확인

GREEN Phase
├─ 최소 구현
└─ 모두 통과 확인

REFACTOR Phase
├─ 코드 개선
└─ 테스트 유지
```

**TDD 엄격함**:

- RED: 구현 코드 금지
- GREEN: 최소한만 추가
- REFACTOR: 테스트 유지

### Phase 4: 보고 및 커밋 (Report & Commit)

```
작업 완료
    ↓
├─ 문서 생성 (생성 설정에 따라)
├─ Git 커밋 (자동)
├─ PR 생성 (팀 모드)
└─ 정리
```

______________________________________________________________________

## :link: TAG 시스템 (추적성)

### TAG 체인

```
SPEC-001 (요구사항)
    ↓
@TEST:APP-001:* (테스트)
    ↓
@CODE:APP-001:* (구현)
    ↓
@DOC:APP-001:* (문서)
    ↓
상호 참조 (완전한 추적성)
```

### 추적성 보장

| 아티팩트 | TAG                | 용도          |
| -------- | ------------------ | ------------- |
| SPEC     | `SPEC-001`         | 요구사항 정의 |
| 테스트   | `@TEST:SPEC-001`   | 요구사항 검증 |
| 코드     | `@CODE:SPEC-001:*` | 구현 추적     |
| 문서     | `@DOC:SPEC-001`    | 문서 동기화   |

______________________________________________________________________

## 💡 핵심 스킬 시스템

### 93개 Claude Skills

Skills는 **Progressive Disclosure** 원칙으로 필요할 때만 로드됩니다.

#### 기초 스킬 (Foundation)

- TRUST 5 원칙
- TAG 시스템
- SPEC 작성법
- Git 워크플로우

#### 필수 스킬 (Essentials)

- 디버깅
- 성능 최적화
- 리팩토링
- 테스트 작성

#### Alfred 스킬 (Alfred)

- 에이전트 가이드
- 워크플로우
- 의사결정 원칙

#### 도메인 스킬 (Domain)

- 데이터베이스
- 백엔드 API
- 프론트엔드 UI
- 보안

#### 언어 스킬 (Language)

- Python 3.13+
- TypeScript 5.7+
- Go 1.24+
- 기타 20개 언어

______________________________________________________________________

## 🔒 안전 메커니즘 (Hooks)

### SessionStart Hook

- 프로젝트 상태 확인
- 세션 로그 분석
- 설정 검증

### PreToolUse Hook

- 위험한 명령 차단
- 권한 검증
- 컨텍스트 전달

### PostToolUse Hook

- 결과 분석
- 오류 감지
- 자동 수정 제안

______________________________________________________________________

## 📊 성능 지표

### 예상 실행 시간

| 단계          | 평균 시간 | 범위       |
| ------------- | --------- | ---------- |
| 의도 파악     | 1분       | 1-5분      |
| 계획 수립     | 2분       | 1-5분      |
| RED 단계      | 3분       | 1-10분     |
| GREEN 단계    | 5분       | 2-15분     |
| REFACTOR 단계 | 5분       | 2-15분     |
| 동기화        | 2분       | 1-5분      |
| **전체**      | **18분**  | **8-55분** |

### 생산성 개선

기존 개발 vs MoAI-ADK:

| 지표            | 기존 | MoAI-ADK | 개선율    |
| --------------- | ---- | -------- | --------- |
| 개발 속도       | 100% | 250%     | **+150%** |
| 버그 감소       | 100% | 20%      | **-80%**  |
| 문서화 시간     | 100% | 10%      | **-90%**  |
| 테스트 커버리지 | 60%  | 95%      | **+35%**  |

______________________________________________________________________

## 🚀 확장성

### 수평 확장 (Horizontal)

Master-Clone 패턴으로 무제한 병렬화:

- 10개 모듈 동시 개발
- 100+ 파일 동시 변경
- 재쓰기 및 마이그레이션

### 수직 확장 (Vertical)

추가 전문가 에이전트 통합:

- 새로운 도메인 전문가 추가
- 커스텀 Skills 확장
- 팀 크기 무제한

______________________________________________________________________

## <span class="material-icons">library_books</span> 다음 학습

- [Alfred 워크플로우](../guides/alfred/index.md) - 4단계 상세 가이드
- [SPEC 작성](../guides/specs/basics.md) - EARS 문법 마스터
- [TDD 구현](../guides/tdd/index.md) - RED-GREEN-REFACTOR
- [TAG 시스템](../guides/specs/tags.md) - 완벽한 추적성

______________________________________________________________________

**궁금한 점?** [온라인 문서 포털](https://adk.mo.ai.kr)을 방문하세요.
