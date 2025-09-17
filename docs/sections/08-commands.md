# MoAI-ADK 명령어 시스템

## 🎯 6개 슬래시 명령어 개요

MoAI-ADK는 4단계 파이프라인을 지원하는 6개의 연번순 슬래시 명령어를 제공합니다.

### 명령어 체계

| 순서 | 명령어 | 담당 에이전트 | 기능 | 단계 |
|------|--------|---------------|------|------|
| **1** | `/moai:1-project` | steering-architect | 프로젝트 설정 | 초기화 |
| **2** | `/moai:2-spec` | spec-manager | EARS 형식 명세 작성 | SPECIFY |
| **3** | `/moai:3-plan` | plan-architect | Constitution Check | PLAN |
| **4** | `/moai:4-tasks` | task-decomposer | TDD 작업 분해 | TASKS |
| **5** | `/moai:5-dev` | code-generator + test-automator | 자동 구현 | IMPLEMENT |
| **6** | `/moai:6-sync` | doc-syncer + tag-indexer | 문서 동기화 | 동기화 |

## 모델 사용 가이드

| 명령어 | 권장 모델 | 비고 |
|--------|-----------|------|
| `/moai:1-project` | `sonnet` | 프로젝트 설정 및 기본 문서화 |
| `/moai:2-spec` | `sonnet` | 명세 작성/정제 |
| `/moai:3-plan` | `opusplan` (plan 모드) | 복잡한 설계·검증 전용 (계획 후 실행은 Sonnet) |
| `/moai:4-tasks` | `sonnet` | TDD 작업 분해 |
| `/moai:5-dev` | `sonnet` | Red-Green-Refactor 구현 |
| `/moai:6-sync` | `haiku` | 문서/인덱싱 동기화 속도 최적 |

> `CLAUDE.md`의 “모델 사용 가이드(opusplan)” 섹션에서 세부 운영 수칙을 확인하세요.

## 명령어 상세

### /moai:1-project - 프로젝트 설정
**기능**: 대화형 마법사를 통한 프로젝트 초기 설정

```bash
# 프로젝트 초기화
/moai:1-project

# 설정 변경
※ 설정/수정은 별도 서브커맨드 없이 `/moai:1-project`에서 대화형으로 처리합니다.
```

**생성 결과**:
- Steering 문서 (product.md, structure.md, tech.md)
- 프로젝트별 디렉토리 구조
- 언어/프레임워크별 맞춤 설정

### /moai:2-spec - EARS 형식 명세 작성
**기능**: 요구사항을 EARS 형식 명세로 변환

```bash
# 새 명세 작성
# 1) 슬러그 + 설명 (기존)
/moai:2-spec user-auth "JWT 기반 사용자 인증 시스템"

# 2) 설명만 입력 (신규) → 슬러그 자동 생성
/moai:2-spec "실시간 알림 시스템"   # ⇒ Slug: user-notification

# 기존 명세 수정
/moai:2-spec SPEC-001 --update
```

**생성 결과**:
- SPEC-XXX 디렉토리
- spec.md (EARS 형식 명세)
- [NEEDS CLARIFICATION] 마커

**자동 슬러그 생성 규칙**
- 설명만 입력 시 영어 케밥케이스 슬러그 생성(2~4단어, 소문자-하이픈).
- 예: "실시간 알림 시스템" → `user-notification` (문맥상 `realtime-notification`도 허용).
- 충돌 시 `-2`, `-3` 접미사로 회피. 생성된 슬러그는 출력 상단에 보고됩니다.

### /moai:3-plan - Constitution Check
**기능**: 5원칙 준수 검증 및 계획 수립

```bash
# Constitution Check 실행
/moai:3-plan SPEC-001

# 품질 게이트 검증
/moai:3-plan SPEC-001 --strict
```

**생성 결과**:
- plan.md (구현 계획)
- ADR 문서 (아키텍처 결정)
- 품질 게이트 통과 인증

### /moai:4-tasks - TDD 작업 분해
**기능**: 명세를 테스트 우선 작업으로 분해

```bash
# 작업 분해 실행
/moai:4-tasks SPEC-001

# Sprint 기반 분해
/moai:4-tasks SPEC-001 --sprint 5days
```

**생성 결과**:
- tasks.md (작업 목록)
- 테스트 케이스 정의
- Red-Green-Refactor 계획

### /moai:5-dev - Red-Green-Refactor 구현
**기능**: TDD 사이클 기반 자동 구현

```bash
# 단일 태스크 구현
/moai:5-dev T001

# 병렬 구현
/moai:5-dev T001 T002 T003

# 전체 SPEC 구현
/moai:5-dev SPEC-001
```

**실행 과정**:
1. 테스트 작성 (RED)
2. 최소 구현 (GREEN)
3. 리팩토링 (REFACTOR)
4. 커버리지 검증

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
- 업데이트 실행 시 자동 백업(.moai_backup_*)을 생성하므로 필요하면 복원하거나 변경 사항을 비교할 수 있습니다.
