---
title: /moai harness 명령어
weight: 55
draft: false
---

Harness v4 Builder로 프로젝트 고유의 동적 전문가 팀을 생성하고 관리합니다.

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai:harness <자연어 요청>`을 입력하면 이 명령어를 바로 실행할 수 있습니다.
{{< /callout >}}

## 개요

`/moai:harness`는 MoAI-ADK의 **Harness v4 Builder**를 실행하여 프로젝트 요구사항에 맞춘 동적 전문가 팀을 자동 생성합니다.

### Harness v4 Builder란?

Harness v4 Builder는 Socratic 인터뷰 기반의 4-phase 워크플로우(ANALYZE → PLAN → GENERATE → ACTIVATE)로 팀을 구성합니다.

| 단계 | 설명 |
|------|------|
| ANALYZE | 프로젝트 구조, 사용 언어, 기존 에이전트 인벤토리 분석 |
| PLAN | 필요한 팀 규모(3~5명), 각 팀원의 역할, worktree 격리 여부 결정 |
| GENERATE | `.claude/agents/harness/` 에이전트 파일, `.moai/harness/manifest.json` 생성 |
| ACTIVATE | 팀 등록 및 `/harness:<name>` 커맨드 활성화 |

## 사용 방법

### 1단계: 자연어로 팀 생성 요청

```bash
> /moai:harness <자연어 요청>
```

**예시:**
```
우리 Go 백엔드 프로젝트에 맞는 전문가 팀을 만들어줘.
DB 마이그레이션, REST API 엔드포인트, 단위 테스트를 각각 담당할 팀이 필요해.
```

### 2단계: Builder의 자동 처리

Builder가 4-phase를 자동 실행합니다:

1. **ANALYZE**: Go, PostgreSQL, REST API 기술 스택 감지
2. **PLAN**: DB Engineer, API Developer, Test Engineer 3인 팀 구성 결정
3. **GENERATE**: 
   - `.claude/agents/harness/db-engineer.md`
   - `.claude/agents/harness/api-developer.md`
   - `.claude/agents/harness/test-engineer.md`
   - `.moai/harness/manifest.json` 생성
4. **ACTIVATE**: `/harness:backend-team` 커맨드 등록

### 3단계: 생성된 팀 활용

생성 후 모든 작업에서 팀을 자동 활용:

```bash
/moai run SPEC-BACKEND-001
/moai run --team SPEC-BACKEND-001    # 팀 모드 강제
```

MoAI가 SPEC 복잡도를 분석하여 manifest의 phase 순서대로 팀원을 자동 위임합니다.

## Harness 관리 커맨드

### harness list

생성된 모든 하네스 목록 조회:

```bash
/harness list
```

### harness:<name> status

특정 하네스의 상세 정보:

```bash
/harness:backend-team status
```

출력 정보:
- 팀원 목록 및 역할
- 사용 모델 (inherit, haiku, sonnet, opus)
- 선택적 worktree 격리 설정
- Manifest 버전 및 생성일

### harness:<name> edit

manifest.json과 에이전트 정의 편집:

```bash
/harness:backend-team edit
```

수정 가능한 항목:
- 팀원 추가/제거
- 스킬 사전 로드 목록
- Worktree 격리 정책
- 역할별 프롬프트

### harness:<name> remove

하네스 및 연관 파일 삭제:

```bash
/harness:backend-team remove
```

삭제되는 항목:
- `.claude/agents/harness/` 에이전트 정의
- `.moai/harness/manifest.json`
- 등록된 `/harness:<name>` 커맨드
- 워크트리 격리 정책

## Manifest 구조

Harness v4는 **manifest.json**으로 팀 구성을 정의합니다.

### manifest.json 예시

```json
{
  "spec_id": "HARNESS-BACKEND-001",
  "name": "Backend Development Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "worktree_isolation": "L1_optional",
  
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "architect",
          "role": "API 아키텍처 전문가",
          "model": "inherit",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "DB 설계 및 마이그레이션",
          "model": "inherit"
        },
        {
          "name": "api-developer",
          "role": "REST API 엔드포인트",
          "model": "inherit"
        },
        {
          "name": "test-engineer",
          "role": "단위 테스트",
          "model": "haiku"
        }
      ]
    }
  ]
}
```

### Phase 필드

| 필드 | 설명 |
|------|------|
| `name` | 단계 이름 (`plan`, `run`, `sync`) |
| `teammates` | 이 단계에 참여할 팀원 배열 |

### Teammate 필드

| 필드 | 기본값 | 설명 |
|------|--------|------|
| `name` | 필수 | 팀원 고유 식별자 |
| `role` | 필수 | 팀원의 역할 설명 |
| `model` | `inherit` | 모델 선택 (`inherit`, `haiku`, `sonnet`, `opus`) |
| `skills` | `[]` | 사전 로드할 스킬 목록 |

## Worktree 격리

Harness v4는 선택적 worktree 격리를 지원합니다.

### L1_optional (기본값)

```json
"worktree_isolation": "L1_optional"
```

Claude Code가 병렬 팀원 간 충돌 감지 시 자동으로 L1 워크트리를 생성합니다.

- **선택적**: 충돌 시에만 격리 적용
- **자동**: 런타임이 충돌 감지 후 자동 생성
- **비용**: 워크트리 격리 시 메모리 증가

### none

```json
"worktree_isolation": "none"
```

모든 팀원이 프로젝트 루트에서 작업합니다 (최소 메모리 사용).

## 팀 위임 워크플로우

Harness가 활성화되면 MoAI는 해당 팀을 자동으로 활용합니다.

### SPEC 실행 시 팀 위임

```bash
> /moai run SPEC-BACKEND-001
```

**MoAI의 자동 판단:**
1. SPEC 복잡도 추정 (파일 수, 코드 라인 수)
2. 적합한 하네스 선택
3. manifest phase 순서대로 팀원 순차/병렬 위임

### Phase 기반 위임 예시

```
PLAN Phase:
  → architect 팀원이 아키텍처 설계 담당

RUN Phase:
  → db-engineer, api-developer 병렬 위임
  → test-engineer 순차 위임 (테스트)

SYNC Phase:
  → 문서 생성 및 PR 작성 (기본 manager-docs)
```

## 자연어 요청의 힘

Harness v4 Builder는 Socratic 인터뷰 방식으로 요구사항을 파악합니다.

### 효과적인 요청 예시

```
우리 팀은 Python FastAPI 백엔드를 개발 중입니다.
API 엔드포인트, 데이터 검증, 에러 핸들링을 잘하는 팀이 필요합니다.
```

Builder가 자동으로:
- Python, FastAPI, asyncio 기술 스택 감지
- 3~5명 팀 규모 결정
- 각 팀원의 특화 영역 설정
- 필요한 스킬 사전 로드

### 불명확한 요청은 Builder가 물어봅니다

```
팀이 필요합니다.

→ Builder: 프로젝트의 주요 기술은? (언어, 프레임워크)
→ Builder: 팀이 집중할 영역은? (백엔드, 프론트엔드, 전체)
→ Builder: 특별히 필요한 전문성은?
```

## 관련 문서

- [Harness v4 Builder 가이드](/advanced/builder-agents) - Builder 4-phase 상세
- [에이전트 가이드](/advanced/agent-guide) - 8개 핵심 에이전트 이해
- [SPEC 기반 개발](/workflow-commands/moai-plan) - SPEC 워크플로우 개요

{{< callout type="info" >}}
**팁**: Harness를 한 번 생성하면, 모든 후속 작업에서 그 팀이 자동으로 활용됩니다. `/harness:team-name` 커맨드로 언제든 재사용할 수 있습니다.
{{< /callout >}}
