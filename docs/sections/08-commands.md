# MoAI-ADK 명령어 시스템

## 🎯 6개 슬래시 명령어 개요

MoAI-ADK는 4단계 파이프라인을 지원하는 6개의 연번순 슬래시 명령어를 제공합니다.

### 명령어 체계

| 순서  | 명령어            | 담당 에이전트                   | 기능                | 단계      |
| ----- | ----------------- | ------------------------------- | ------------------- | --------- |
| **1** | `/moai:1-project` | steering-architect              | 프로젝트 설정       | 초기화    |
| **2** | `/moai:2-spec`    | spec-manager                    | EARS 형식 명세 작성 | SPECIFY   |
| **3** | `/moai:3-plan`    | plan-architect                  | Constitution Check  | PLAN      |
| **4** | `/moai:4-tasks`   | task-decomposer                 | TDD 작업 분해       | TASKS     |
| **5** | `/moai:5-dev`     | code-generator + test-automator | 자동 구현           | IMPLEMENT |
| **6** | `/moai:6-sync`    | doc-syncer + tag-indexer        | 문서 동기화         | 동기화    |

## 모델 사용 가이드

| 명령어            | 권장 모델              | 비고                                                 |
| ----------------- | ---------------------- | ---------------------------------------------------- |
| `/moai:1-project` | `sonnet`               | 프로젝트 설정 및 기본 문서화                         |
| `/moai:2-spec`    | `sonnet`               | 명세 작성/정제 (`all` 옵션: 최대 10개 병렬)          |
| `/moai:3-plan`    | `opusplan` (plan 모드) | 복잡한 설계·검증 전용 (`all` 옵션: 최대 10개 병렬)   |
| `/moai:4-tasks`   | `sonnet`               | TDD 작업 분해 (`all` 옵션: 최대 10개 병렬)           |
| `/moai:5-dev`     | `sonnet`               | Red-Green-Refactor 구현 (`all` 옵션: 최대 10개 병렬) |
| `/moai:6-sync`    | `haiku`                | 문서/인덱싱 동기화 속도 최적                         |

> `CLAUDE.md`의 “모델 사용 가이드(opusplan)” 섹션에서 세부 운영 수칙을 확인하세요.

## 명령어 상세

### /moai:1-project - 프로젝트 설정

**기능**: 대화형 마법사를 통한 프로젝트 초기 설정

```bash
# 현재 디렉토리명으로 프로젝트 초기화
/moai:1-project

# 프로젝트명을 지정한 초기화
/moai:1-project my-awesome-app
```

**인자**: `[프로젝트이름]` (선택사항, 기본값: 현재 디렉토리명)

**생성 결과**

- **Steering 문서**: product.md, structure.md, tech.md
- **Top-3 SPEC 디렉터리** (`SPEC-00X/`):
  - 기본 필수: `spec.md` (EARS 형식), `acceptance.md` (수락 기준)
  - 조건부 선택: `design.md`, `data-model.md`, `contracts/`, `research.md`
- **백로그 STUB**: `.moai/specs/backlog/` ([NEEDS CLARIFICATION] 마커 포함)
- **메모리 시스템**: `.moai/memory/common.md` + 선택 기술 스택별 메모
- **16-Core TAG 시스템**: 초기화 및 추적성 매트릭스 구축

**SPEC 파일 생성 규칙**
각 SPEC-XXX 디렉터리에는 **내용에 따라 필요한 파일만** 생성:

- **기본 필수**: `spec.md` (EARS 형식), `acceptance.md` (수락 기준)
- **조건부 선택**: `design.md` (복잡한 설계), `data-model.md` (데이터 구조), `contracts/` (API), `research.md` (기술 조사)
- **생성 기준**: UX/UI 개선→기본 파일만, API 개발→contracts/ 추가, 데이터 처리→data-model.md 추가

**확정 절차**:

1. 10단계 대화형 질문 완료 및 최종 요약 확인
2. `모델 opusplan`으로 Plan 모드에서 세부 계획을 검토
3. 필요한 추론을 마친 후 실행 모드(`모델 sonnet`)로 복귀
4. `확정` 응답 시 마법사가 Steering 문서 + Top-3 SPEC + Constitution Check 자동 실행

### /moai:2-spec - EARS 형식 명세 작성

**기능**: 요구사항을 EARS 형식 명세로 변환

```bash
# 새 기능 SPEC 생성
/moai:2-spec "JWT 기반 사용자 인증 시스템"
/moai:2-spec "실시간 채팅 시스템 - WebSocket 기반, 파일 첨부 지원, 읽음 표시 기능"

# 기존 SPEC 수정/보완
/moai:2-spec SPEC-001 "추가 보안 요구사항 반영"
/moai:2-spec SPEC-003  # 특정 SPEC 재생성

# 전체 프로젝트 SPEC 병렬 생성
/moai:2-spec all                    # Steering 문서 기반 모든 SPEC 자동 생성
/moai:2-spec all "P0,P1만 생성"     # 특정 우선순위만 생성
/moai:2-spec all "auth,payment 도메인만"  # 특정 도메인만 생성
```

**인자**: `<SPEC-ID|"기능설명"|"all"> [추가세부사항...]`

**생성 결과**

- **`SPEC-00X/` 디렉터리**: 내용별 맞춤 파일 생성
  - 기본: `spec.md`, `acceptance.md`
  - 선택적: `design.md`, `data-model.md`, `contracts/`, `research.md`
- **[NEEDS CLARIFICATION] 마커**: 불완전한 요구사항 자동 표시
- **백로그 관리**: `.moai/specs/backlog/` STUB 보관
- **승격 지원**: `/moai:2-spec all "백로그 승격"` 명령으로 SPEC-00X 승격

**사용 흐름**:

1. 생성 전 `SPEC 미리보기`를 확인하고 Plan 모드(`모델 opusplan`)에서 보완할 내용을 검토합니다.
2. 사용자 확정(“확정”, “좋습니다”) 이후에만 문서가 생성됩니다.

**`all` 옵션 - 병렬 SPEC 생성** :

- **전체 프로젝트 SPEC 상태 스캔**: 기존 SPEC 디렉터리 자동 감지
- **Steering 문서 기반 분석**: product.md/structure.md/tech.md 분석으로 누락 기능 파악
- **우선순위별 일괄 처리**: P0/P1/P2 순으로 SPEC 생성/갱신
- **도메인별 필터링**: "auth,payment 도메인만" 등 선택적 생성
- **[NEEDS CLARIFICATION] 해소**: 기존 SPEC의 불완전 마커 자동 해결
- **백로그 승격**: `.moai/specs/backlog/` → `SPEC-00X/` 승격 지원
- **병렬 처리 최적화**: 의존성 없는 작업 **최대 10개 동시 실행**, 70% 시간 단축
- **전문 에이전트 분산**: spec-manager 에이전트가 독립 SPEC별로 병렬 처리

### /moai:3-plan - Constitution Check

**기능**: 5원칙 준수 검증 및 계획 수립

```bash
# 단일 SPEC 계획 수립
/moai:3-plan SPEC-001

# 모든 SPEC 병렬 계획 생성
/moai:3-plan all

# 품질 게이트 검증
/moai:3-plan SPEC-001 --strict

# 특정 조건의 SPEC들만 계획 생성
/moai:3-plan all "P0,P1 우선순위만"
/moai:3-plan all "완성된 SPEC만"
```

**인자**: `<SPEC-ID|"all"> [필터조건...]`

**💡 Pro Tip - Opusplan 모델 사용 권장**:
계획 수립은 복잡한 사고가 필요하므로 `opusplan` 모델을 강력히 권장합니다:

```bash
# 1단계: Opusplan 모델로 전환
claude --model opusplan

# 2단계: Plan 모드에서 계획 수립
⏸ plan mode on (shift+tab to cycle)
> /moai:3-plan SPEC-001

# 3단계: 계획 완료 후 자동으로 실행 모드 전환
⏵⏵ accept edits on (shift+tab to cycle)
```

**생성 결과**:

- plan.md (구현 계획)
- ADR 문서 (아키텍처 결정)
- 품질 게이트 통과 인증

**`all` 옵션 - 병렬 PLAN 생성** :

- **SPEC 디렉터리 전체 스캔**: 생성된 모든 SPEC 자동 감지
- **Constitution 5원칙 검증**: 각 SPEC별 독립 검증 수행
- **병렬 처리 최적화**: 의존성 없는 계획 수립 **최대 10개 동시 실행**
- **전문 에이전트 분산**: plan-architect 에이전트가 독립 SPEC별로 병렬 처리
- **의존성 관계 분석**: 자동 의존성 체인 분석 및 실행 순서 최적화
- **필터 조건 지원**: 우선순위, 완성도, 도메인별 선택적 생성

### /moai:4-tasks - TDD 작업 분해

**기능**: 명세를 테스트 우선 작업으로 분해

```bash
# 단일 SPEC 작업 분해
/moai:4-tasks SPEC-001

# 모든 SPEC 병렬 작업 분해
/moai:4-tasks all

# Sprint 기반 분해
/moai:4-tasks SPEC-001 --sprint 5days
/moai:4-tasks all --sprint 2weeks
```

**인자**: `<SPEC-ID|"all"> [옵션...]`

**생성 결과**:

- tasks.md (작업 목록)
- 테스트 케이스 정의
- Red-Green-Refactor 계획

**`all` 옵션 - 병렬 작업 분해**:

- **계획된 모든 SPEC 스캔**: plan.md가 있는 SPEC만 자동 선택
- **병렬 처리 최적화**: 의존성 없는 작업 분해 **최대 10개 동시 실행**
- **전문 에이전트 분산**: task-decomposer 에이전트가 독립 SPEC별로 병렬 처리
- **Sprint 단위 최적화**: 전체 프로젝트를 Sprint 기간별로 균등 분배

### /moai:5-dev - Red-Green-Refactor 구현

**기능**: TDD 사이클 기반 자동 구현

```bash
# 단일 태스크 구현
/moai:5-dev T001

# 태스크 병렬 구현
/moai:5-dev T001 T002 T003

# 단일 SPEC 전체 구현
/moai:5-dev SPEC-001

# 모든 SPEC 병렬 구현
/moai:5-dev all

# 특정 조건의 태스크만 구현
/moai:5-dev all --priority P0,P1
/moai:5-dev all --ready-tasks-only
```

**인자**: `<TASK-ID|SPEC-ID|"all"> [옵션...]`

**실행 과정**:

1. 테스트 작성 (RED)
2. 최소 구현 (GREEN)
3. 리팩토링 (REFACTOR)
4. 커버리지 검증

**`all` 옵션 - 병렬 구현**:

- **작업 분해된 모든 SPEC 스캔**: tasks.md가 있는 SPEC만 자동 선택
- **병렬 처리 최적화**: 의존성 없는 구현 작업 **최대 10개 동시 실행**
- **전문 에이전트 분산**: code-generator + test-automator 에이전트가 독립 태스크별로 병렬 처리
- **TDD 사이클 보장**: 각 태스크마다 Red-Green-Refactor 완전 수행
- **커버리지 통합**: 병렬 실행 후 전체 커버리지 통합 보고서 생성

### /moai:6-sync - 문서 동기화

**기능**: 코드와 문서의 실시간 동기화

```bash
# 전체 동기화
/moai:6-sync

# 특정 파일 동기화
/moai:6-sync src/auth.py

# TAG 인덱스 갱신
/moai:6-sync --tags-only
```

**동기화 범위**:

- @TAG 인덱스 업데이트
- 추적성 매트릭스 갱신
- API 문서 자동 생성

## 명령어 실행 플로우

### 표준 개발 플로우

```bash
# 1. 프로젝트 설정 (최초 1회)
/moai:1-project

# 2-6. 기능 개발 사이클 (반복)
/moai:2-spec payment "Stripe 결제 시스템"
/moai:3-plan SPEC-002
/moai:4-tasks SPEC-002
/moai:5-dev T001 T002 T003
/moai:6-sync
```

### 배치 개발 플로우 (병렬 처리) - **최대 10개 동시 실행**

```bash
# 1. 프로젝트 설정
/moai:1-project

# 2. 모든 기능 SPEC 병렬 생성 (최대 10개 동시)
/moai:2-spec all "P0,P1 우선순위"

# 3. 모든 SPEC 계획 병렬 생성 (Opusplan 권장, 최대 10개 동시)
claude --model opusplan
/moai:3-plan all

# 4. 모든 SPEC 작업 분해 병렬 실행 (최대 10개 동시)
/moai:4-tasks all

# 5. 모든 태스크 병렬 구현 (최대 10개 동시)
/moai:5-dev all

# 6. 전체 문서 동기화
/moai:6-sync
```

**⚡ 병렬 처리 효과**:

- **시간 단축**: 70-80% 시간 절약 (기존 순차 처리 대비)
- **에이전트 활용**: 각 전문 에이전트가 독립적으로 최대 10개 작업 동시 처리
- **의존성 관리**: 자동 의존성 분석으로 안전한 병렬 실행 보장

### 빠른 개발 플로우

```bash
# 전체 파이프라인 자동 실행
/moai:2-spec quick-feature "간단한 기능" --auto-pipeline
```

## Hook 통합

각 명령어는 Hook 시스템과 완전히 통합됩니다:

### PreToolUse Hook

- Constitution 규칙 검증
- @TAG 형식 검증
- 정책 준수 확인

### PostToolUse Hook

- 자동 문서 동기화
- 인덱스 업데이트
- 다음 단계 안내

## 에러 처리

### 자동 복구

```bash
# Constitution Check 실패 시
/moai:3-plan SPEC-001 --fix-issues

# 모든 SPEC 계획 재생성
/moai:3-plan all --retry-failed

# TAG 불일치 시
/moai:6-sync --repair-tags

# 테스트 실패 시
/moai:5-dev T001 --debug-tests
```

### 상태 확인

```bash
# 현재 프로젝트 상태(요약)
moai status

# 상세 상태(버전/파일 카운트 등)
moai status -v

# 특정 경로의 프로젝트 상태
moai status -p /path/to/project
```

명령어 시스템은 **직관적인 워크플로우**와 **자동화된 품질 보장**을 통해 개발 생산성을 극대화합니다.

## moai update - 리소스/패키지 업데이트

MoAI-ADK CLI는 프로젝트 리소스 버전을 추적하고 자동 업데이트를 지원합니다.

```bash
# 업데이트 필요 여부만 확인
moai update --check

# 템플릿/Hook/UI 리소스 갱신 (기본값: 백업 생성 후 덮어쓰기)
moai update --resources-only

# 패키지 업그레이드 안내 포함 (pip 수동 설치 필요)
moai update
```

- `.moai/version.json`에 현재 템플릿 버전이 기록되며 `moai status`에서 확인할 수 있습니다.
- `--check`는 설치된 템플릿 버전과 패키지 제공 버전을 비교해 업데이트 필요 여부를 알려줍니다.
- 업데이트 실행 시 자동 백업(.moai*backup*\*)을 생성하므로 필요하면 복원하거나 변경 사항을 비교할 수 있습니다.
