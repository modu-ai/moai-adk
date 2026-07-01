---
title: 빌더 에이전트와 하네스 v4
weight: 40
draft: false
---

MoAI-ADK 확장을 위한 Harness v4 Builder를 상세히 안내합니다.

{{< callout type="info" >}}
  **한 줄 요약**: Harness v4 Builder는 자연어 요청으로 프로젝트 고유의 전문가 팀을 동적으로 생성합니다. 4단계 워크플로우(ANALYZE → PLAN → GENERATE → ACTIVATE)와 manifest 기반 Runner로 구성됩니다.
{{< /callout >}}

## Harness v4 Builder란?

Harness v4 Builder는 `/moai:harness <자연어 요청>`을 통해 **프로젝트 고유의 전문가 팀을 동적으로 생성**합니다.

### 이전 버전과의 차이

| 구분 | 이전 (v3/정적 모델) | 현재 (v4 Builder) |
|------|-----|-----------|
| 생성 방식 | 3가지 빌더 에이전트 (빌더-스킬, 빌더-에이전트, 빌더-플러그인) | 단일 Harness v4 Builder (동적 생성) |
| 워크플로우 | 사용자 정의 구조 | 4-phase ANALYZE → PLAN → GENERATE → ACTIVATE |
| 실행 방식 | 각각 독립적 | Manifest 기반 Runner (선택적 worktree 격리) |
| 확장성 | 제한적 | 프로젝트 컨텍스트 자동 감지 |

## Harness v4 Builder 4-Phase Workflow

### 1. ANALYZE (분석 단계)

현재 프로젝트를 분석하고 필요한 전문성을 파악합니다.

- 소스 코드 구조 분석
- 사용 언어 및 프레임워크 감지
- 기존 에이전트/스킬 인벤토리 조사
- 프로젝트 규모 추정

### 2. PLAN (계획 단계)

필요한 전문가 팀의 구성과 역할을 정의합니다.

- 팀 규모 결정 (3~5 팀원)
- 각 팀원의 역할 프로필 정의
- worktree 격리 필요성 판단
- Manifest 스키마 설계

### 3. GENERATE (생성 단계)

실제 에이전트 정의와 설정을 생성합니다.

- `.claude/agents/harness/` 아래 에이전트 파일 생성
- `.moai/harness/manifest.json` 생성 (Runner 설정)
- 역할별 시스템 프롬프트 작성
- 스킬 사전 로드 목록 정의

### 4. ACTIVATE (활성화 단계)

생성된 하네스를 즉시 사용 가능하도록 활성화합니다.

- 에이전트 등록 및 검증
- Manifest Runner 초기화
- 선택적 worktree 생성 및 격리 설정
- 팀원 자동 위임 규칙 활성화

## Manifest 기반 Runner

Harness v4는 **Manifest 기반 Runner**를 사용하여 생성된 팀을 운영합니다.

### manifest.json 구조

```json
{
  "spec_id": "HARNESS-PROJECT-001",
  "name": "My Project Custom Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "researcher",
          "model": "haiku",
          "mode": "plan",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "implementer",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        }
      ]
    }
  ],
  "worktree_isolation": "L1_optional"
}
```

### Runner 동작

1. **Phase 진입**: manifest의 phase 시퀀스를 따라 진행
2. **Teammate Spawn**: 각 phase의 teammates를 동적으로 생성
3. **Isolation 적용**: 조건부 worktree 격리 적용
4. **Result Aggregation**: 각 teammate의 결과를 통합

## Harness Lifecycle Commands

Harness v4 Builder로 생성된 하네스는 `/harness:<name>` 명령어로 관리됩니다.

### 사용 가능한 명령어

```bash
# 생성된 하네스 목록 조회
/harness list

# 특정 하네스 상태 확인
/harness:my-project-team status

# 하네스 설정 편집
/harness:my-project-team edit

# 하네스 삭제
/harness:my-project-team remove

# Harness v4 Builder로 새 하네스 생성
/moai:harness <자연어 요청>
```

## 자연어 요청으로 하네스 생성

### 기본 사용법

```bash
> 우리 백엔드 프로젝트에 맞는 전문가 팀을 만들어줘.
> API 설계, DB 스키마, 테스트를 담당할 팀이 필요해.
```

### Builder의 동작 흐름

1. ANALYZE: 프로젝트 구조(Go, PostgreSQL, REST API)를 분석
2. PLAN: 3명 팀 (API Designer, DB Specialist, Test Engineer) 결정
3. GENERATE: 각 에이전트 정의와 manifest.json 생성
4. ACTIVATE: 팀 활성화 및 `/harness:backend-team` 커맨드 등록

### 생성 결과 위치

- 에이전트 정의: `.claude/agents/harness/api-designer.md`, `db-specialist.md`, ...
- Manifest: `.moai/harness/manifest.json`
- 선택적 워크트리: `~/.moai/worktrees/<project>/` (사용자 opt-in 시)

## Worktree 격리 (선택적)

Harness v4는 조건부 worktree 격리를 지원합니다.

### L1 격리 (Optional)

Claude Code 런타임이 에이전트당 L1 워크트리를 생성합니다.

- **사용 시점**: 병렬 팀원이 같은 파일을 편집할 때
- **격리 범위**: 각 팀원의 파일 쓰기가 독립적인 워크트리에서 발생
- **비용**: 추가 메모리 + 병렬 이점 상쇄

### 비활성화

manifest의 `"worktree_isolation": "none"`으로 설정하면 L1 격리 생략.

## 관련 문서

- [Harness v4 Builder 심화 가이드](/advanced/harness-v4-builder) - Builder 4-phase 상세 및 manifest 스키마
- [에이전트 가이드](/advanced/agent-guide) - 8개 핵심 에이전트 카탈로그
- [동적 워크플로우](/advanced/ultracode-workflows) - `/effort ultracode` 병렬 실행

{{< callout type="info" >}}
**팁**: Harness v4 Builder는 프로젝트마다 **커스텀 팀을 한 번만 생성**하면, 이후 모든 작업에서 자동으로 해당 팀이 위임됩니다. 처음 생성 후엔 `/harness:team-name`으로 언제든 재활용할 수 있습니다.
{{< /callout >}}
