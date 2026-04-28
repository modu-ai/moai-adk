# SPEC-CC297-001: Claude Code 2.1.97 Feature Adoption

---
id: SPEC-CC297-001
status: completed
---

## Overview

Claude Code 2.1.94~2.1.97 릴리즈에서 추가된 기능을 moai-adk-go에 채택한다.
핵심: UserPromptSubmit 훅에서 사용자 언어로 세션 제목 자동 설정.

## Requirements

### REQ-1: UserPromptSubmit Session Title (Priority: HIGH)

**EARS Format**: WHEN a user submits a prompt in a Claude Code session, the system SHALL set the session title automatically based on the current context (SPEC-ID, workflow phase, branch name) in the user's configured conversation_language.

**Acceptance Criteria**:
- AC-1.1: `HookSpecificOutput` struct에 `SessionTitle string` 필드 추가
- AC-1.2: UserPromptSubmit 핸들러가 첫 번째 프롬프트 제출 시 세션 제목 생성
- AC-1.3: SPEC 작업 중이면 제목 형식: `SPEC-XXX: <설명>` (설명은 conversation_language)
- AC-1.4: 비-SPEC 세션이면 제목 형식: `<ProjectName> / <BranchName>`
- AC-1.5: conversation_language 설정에 따른 다국어 제목 지원 (ko, en, ja, zh)
- AC-1.6: 세션당 1회만 제목 설정 (중복 설정 방지)
- AC-1.7: 테이블 드리븐 테스트로 85%+ 커버리지

**Technical Approach**:
- `internal/hook/types.go`: HookSpecificOutput에 SessionTitle 추가
- `internal/hook/user_prompt_submit.go`: 핸들러 확장
  - ConfigProvider로 language.yaml 읽기
  - 프로젝트 디렉토리에서 활성 SPEC 감지 (.moai/specs/*/spec.md)
  - git branch명 추출
  - 다국어 제목 생성 로직
- 세션 상태 추적: 첫 프롬프트 감지용 플래그 (세션당 1회)

**다국어 제목 예시**:
| Context | ko | en | ja |
|---------|----|----|-----|
| SPEC 작업 | `SPEC-CC297-001: CC 2.97 기능 채택` | `SPEC-CC297-001: CC 2.97 Feature Adoption` | `SPEC-CC297-001: CC 2.97 機能採用` |
| 일반 작업 | `moai-adk-go / feat/hook-enhancement` | `moai-adk-go / feat/hook-enhancement` | `moai-adk-go / feat/hook-enhancement` |

### REQ-2: StatusLine refreshInterval (Priority: HIGH)

**EARS Format**: WHEN moai-adk-go generates settings.json, the system SHALL include a configurable `refreshInterval` field in the statusLine section.

**Acceptance Criteria**:
- AC-2.1: `statusline.yaml`에 `refresh_interval` 설정 추가 (기본값: 10초 = 10000ms)
- AC-2.2: `settings.json.tmpl`의 statusLine 섹션에 `refreshInterval` 필드 추가
- AC-2.3: settings.json 생성 시 statusline.yaml의 값 반영
- AC-2.4: 0이면 비활성 (일회성 실행, 현재 동작 유지)

**Technical Approach**:
- `internal/template/templates/.moai/config/sections/statusline.yaml`: refresh_interval 추가
- `internal/template/templates/.claude/settings.json.tmpl`: statusLine에 refreshInterval 추가
- settings 렌더링 시 config 값 주입

### REQ-3: StatusLine Worktree Segment (Priority: HIGH)

**EARS Format**: WHEN Claude Code provides `workspace.git_worktree` in status line JSON input AND the session is running inside a git worktree, the system SHALL display a worktree indicator in the status line.

**Acceptance Criteria**:
- AC-3.1: `WorkspaceInfo` struct에 `GitWorktree string` 필드 추가
- AC-3.2: `SegmentWorktree` 세그먼트 상수 추가
- AC-3.3: worktree 감지 시 브랜치 라인에 worktree 표시 (예: `[WT]` prefix)
- AC-3.4: statusline.yaml에 `worktree` 세그먼트 토글 추가
- AC-3.5: worktree 미사용 시 표시 없음 (기본 동작 유지)

**Technical Approach**:
- `internal/statusline/types.go`: WorkspaceInfo에 GitWorktree, SegmentWorktree 추가
- `internal/statusline/renderer.go`: renderDirGitLine에 worktree 표시 로직
- `internal/statusline/builder.go`: collectAll에 worktree 데이터 수집

### REQ-4: Minimum Version Documentation (Priority: MEDIUM)

**EARS Format**: WHEN moai-adk-go documentation references worktree isolation or Stop/SubagentStop hooks, the system SHALL specify Claude Code 2.1.97 as the minimum recommended version.

**Acceptance Criteria**:
- AC-4.1: worktree-integration.md에 최소 버전 요구사항 추가 (CWD leak fix)
- AC-4.2: CLAUDE.md 또는 관련 문서에 Stop/SubagentStop 안정성 참고 추가
- AC-4.3: settings.json.tmpl 주석에 최소 권장 버전 명시

**Technical Approach**:
- Template 문서 파일 업데이트 (make build 후 로컬 동기화)

## Scope

### In Scope
- HookSpecificOutput.SessionTitle 필드 추가 및 구현
- 다국어 세션 제목 생성 로직
- StatusLine refreshInterval 설정
- StatusLine worktree 세그먼트
- 최소 버전 문서 업데이트

### Out of Scope
- Output style keep-coding-instructions (이미 적용됨)
- Claude Code 내부 버그 수정 (2.1.97에서 해결됨)
- Bedrock/Mantle 지원 (moai-adk-go 범위 외)

## Dependencies

- Claude Code >= 2.1.94 (sessionTitle 지원)
- Claude Code >= 2.1.97 (refreshInterval, git_worktree, CWD leak fix)
- 기존 SPEC: SPEC-HOOK-001~009, SPEC-STATUSLINE-001~002

## Files to Modify

### Go Source (implementation)
1. `internal/hook/types.go` — HookSpecificOutput에 SessionTitle 추가
2. `internal/hook/user_prompt_submit.go` — 세션 제목 생성 로직
3. `internal/hook/user_prompt_submit_test.go` — 테스트 확장
4. `internal/statusline/types.go` — WorkspaceInfo.GitWorktree, SegmentWorktree
5. `internal/statusline/types_test.go` — 타입 테스트
6. `internal/statusline/renderer.go` — worktree 세그먼트 렌더링
7. `internal/statusline/renderer_test.go` — 렌더러 테스트
8. `internal/statusline/builder.go` — worktree 데이터 수집

### Templates (make build 필요)
9. `internal/template/templates/.moai/config/sections/statusline.yaml` — refresh_interval
10. `internal/template/templates/.claude/settings.json.tmpl` — refreshInterval

### Documentation (templates)
11. `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` — 최소 버전

### Local Sync (make build 후)
12. `.moai/config/sections/statusline.yaml`
13. `.claude/settings.json`
14. `.claude/rules/moai/workflow/worktree-integration.md`

## Risk Assessment

- **Low Risk**: 모든 변경이 기존 동작에 additive (기존 기능 깨뜨리지 않음)
- **SessionTitle**: 첫 프롬프트에만 동작, 빈 문자열이면 Claude Code가 무시
- **refreshInterval**: 0이면 현재 동작과 동일 (비활성)
- **GitWorktree**: 빈 문자열이면 세그먼트 미표시

## Estimation

- REQ-1 (SessionTitle): 핵심 기능 — Go 핸들러 + 다국어 + 테스트
- REQ-2 (refreshInterval): 설정 추가 — YAML + 템플릿
- REQ-3 (Worktree Segment): 중간 — struct + renderer + 테스트
- REQ-4 (Docs): 경량 — 문서 업데이트

---

Version: 1.0.0
Created: 2026-04-09
Author: MoAI (manager-spec)
Branch: feat/hook-infra-enhancement
