---
title: Hooks 이벤트 레퍼런스
weight: 60
draft: false
---

MoAI-ADK v2.10.1 기준, Claude Code의 훅 시스템은 **27개 이벤트 타입**, **4가지 훅 타입**, **이벤트별 매처**, **스마트 동작**을 지원합니다.

> 훅의 기본 개념과 설정 방법은 [Hooks 가이드](/ko/advanced/hooks-guide)를 참조하세요. 이 페이지는 이벤트 전체 레퍼런스입니다.

## 훅 타입

| 타입 | 설명 | 예시 |
|------|------|------|
| **command** | 셸 스크립트 실행 | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM 평가 | 프롬프트 텍스트를 LLM이 실행하여 결과 반환 |
| **agent** | 서브에이전트 검증 | 에이전트가 작업을 검증하고 결과 반환 |
| **http** | 웹훅 엔드포인트 | HTTP POST 요청으로 이벤트 전달 |

## 이벤트 전체 레퍼런스 (27개)

### 라이프사이클 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `SessionStart` | 세션 시작 | — |
| `SessionEnd` | 세션 종료 | — |
| `Stop` | 에이전트 정지 | — |
| `SubagentStop` | 서브에이전트 정지 | — |
| `SubagentStart` | 서브에이전트 시작 | — |
| `StopFailure` | 정지 실패 | `errorType` |
| `Setup` | 초기 설정 | — |

### 도구 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `PreToolUse` | 도구 실행 전 | `toolName` |
| `PostToolUse` | 도구 실행 후 | `toolName` |
| `PostToolUseFailure` | 도구 실행 실패 | `toolName`, `errorType` |

### 컨텍스트 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `PreCompact` | 컨텍스트 압축 전 | — |
| `PostCompact` | 컨텍스트 압축 후 | — |
| `InstructionsLoaded` | 인스트럭션 로드 완료 | — |

### 입력 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `UserPromptSubmit` | 사용자 프롬프트 제출 | — |
| `Elicitation` | Elicitation 시작 | — |
| `ElicitationResult` | Elicitation 완료 | — |

### 보안 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `PermissionRequest` | 권한 요청 | `toolName` |
| `PermissionDenied` | 권한 거부 | `toolName` |

### 팀 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `TeammateIdle` | 팀원 유휴 상태 전환 | — |
| `TaskCompleted` | 태스크 완료 표시 | — |
| `TaskCreated` | 태스크 생성 | — |

### 워크트리 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `WorktreeCreate` | 워크트리 생성 | — |
| `WorktreeRemove` | 워크트리 삭제 | — |

### 환경 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `ConfigChange` | 설정 변경 | `configSource` |
| `CwdChanged` | 작업 디렉터리 변경 | — |
| `FileChanged` | 파일 변경 | — |

### UI 이벤트

| 이벤트 | 설명 | 매처 |
|--------|------|------|
| `Notification` | 사용자 알림 | — |

## 스마트 동작 (Smart Behaviors)

MoAI-ADK 훅은 단순 이벤트 처리를 넘어 지능적인 동작을 수행합니다:

### PermissionDenied 자동 재시도

읽기 전용 도구(Read, Grep, Glob)의 권한이 거부되면, 훅이 자동으로 재시도를 트리거합니다. 이는 백그라운드 에이전트에서 권한 프롬프트가 표시되지 않는 문제를 완화합니다.

### StopFailure 에러 타입 응답

에이전트 정지 실패 시 에러 타입에 따라 차별화된 응답을 제공합니다. 장시간 실행 세션에서의 안정성을 보장합니다.

### PostCompact 세션 메모 복원

컨텍스트 압축 후 중요한 세션 메모(진행 상태, SPEC 참조)를 자동으로 복원합니다. 이를 통해 컨텍스트 압축 시 핵심 정보 유실을 방지합니다.

### SubagentStart 컨텍스트 주입

서브에이전트 시작 시 필요한 컨텍스트(프로젝트 규칙, MX 태그, 진행 상태)를 자동 주입합니다.

## 매처 (Matchers)

매처를 사용하면 특정 조건에서만 훅이 실행되도록 필터링할 수 있습니다:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": { "toolName": "Bash" },
      "hooks": [{
        "type": "command",
        "command": "echo 'Bash tool detected'",
        "timeout": 5
      }]
    }]
  }
}
```

### 사용 가능한 매처 필드

| 매처 필드 | 적용 이벤트 | 설명 |
|----------|-----------|------|
| `toolName` | PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, PermissionDenied | 도구 이름으로 필터 |
| `errorType` | StopFailure, PostToolUseFailure | 에러 유형으로 필터 |
| `configSource` | ConfigChange | 설정 소스로 필터 |

## CLAUDE_ENV_FILE

`CwdChanged`와 `FileChanged` 훅을 통해 환경 변수를 지속적으로 관리할 수 있습니다:

```bash
# .claude/hooks/moai/handle-cwd-changed.sh
# CLAUDE_ENV_FILE을 통해 환경 변수 영속화
echo "MOAI_PROJECT_DIR=$(pwd)" >> "$CLAUDE_ENV_FILE"
```

이를 통해 세션 간 환경 변수를 유지하고, 디렉터리 변경 시 자동으로 환경을 재설정할 수 있습니다.

## MoAI-ADK가 사용하는 주요 훅

| 이벤트 | MoAI 핸들러 | 역할 |
|--------|-----------|------|
| `SessionStart` | `handle-session-start.sh` | Statusline 초기화, 메트릭 세션 시작 |
| `PostToolUse` | `handle-post-tool.sh` | Task 메트릭 로깅 |
| `TeammateIdle` | `handle-teammate-idle.sh` | LSP 품질 게이트 검증 |
| `TaskCompleted` | `handle-task-completed.sh` | SPEC 문서 존재 확인 |
| `WorktreeCreate` | `handle-worktree-create.sh` | 워크트리 생성 로깅 |
| `WorktreeRemove` | `handle-worktree-remove.sh` | 워크트리 삭제 로깅 |
| `UserPromptSubmit` | `handle-user-prompt.sh` | 품질 게이트 자동 실행 |

## 다음 단계

- [Hooks 가이드](/ko/advanced/hooks-guide) — 훅 기본 개념과 설정 방법
- [settings.json 가이드](/ko/advanced/settings-json) — settings.json 전체 레퍼런스
- [CLI 레퍼런스](/ko/getting-started/cli) — `moai hook` 명령어 상세
