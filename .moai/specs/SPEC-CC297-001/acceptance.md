---
spec_id: SPEC-CC297-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  Post-implementation SDD artifact. PR #610에서 Claude Code 2.1.97 기능
  채택(sessionTitle, refreshInterval, git_worktree)이 출시됨.
  spec.md의 4개 REQ를 실제 구현(`internal/hook/user_prompt_submit.go`,
  `internal/statusline/types.go|builder.go|renderer.go`, 템플릿 yaml/tmpl)과
  대조하여 AC를 역도출. plan-auditor 2026-04-24 감사 시 acceptance.md 부재.
---

# Acceptance Criteria — SPEC-CC297-001

Claude Code 2.1.94~2.1.97 릴리스의 네 기능을 채택한 구현의 관찰 가능한 인수 기준.

## Traceability

| REQ ID | AC ID | Test / Evidence Reference |
|--------|-------|---------------------------|
| REQ-1 (UserPromptSubmit SessionTitle) | AC-001 ~ AC-004 | `internal/hook/user_prompt_submit.go`, `types.go:277` |
| REQ-2 (StatusLine refreshInterval) | AC-005, AC-006 | `internal/template/templates/.claude/settings.json.tmpl:406`, `statusline.yaml:13` |
| REQ-3 (StatusLine Worktree Segment) | AC-007, AC-008 | `internal/statusline/types.go:126,240`, `builder.go:205`, `renderer.go:313` |
| REQ-4 (Minimum Version Docs) | AC-009 | `worktree-integration.md` Minimum Version 표 |

## AC-001: HookSpecificOutput 구조체에 SessionTitle 필드가 존재한다

**Given** 훅 응답 타입 정의가 컴파일된 상태에서,
**When** `HookSpecificOutput` 구조체를 참조하면,
**Then** `SessionTitle string \`json:"sessionTitle,omitempty"\`` 필드가 존재해야 하며 UserPromptSubmit 훅에서 Claude Code UI의 세션 제목을 설정하는 경로로 사용되어야 한다.

**Verification**: `internal/hook/types.go:277` — `SessionTitle string \`json:"sessionTitle,omitempty"\``.

## AC-002: UserPromptSubmit 핸들러가 활성 SPEC 기반 제목을 생성한다

**Given** `cwd/.moai/specs/SPEC-XXX/spec.md`가 존재하는 프로젝트에서,
**When** 사용자가 프롬프트를 제출하면 `userPromptSubmitHandler.Handle`이 호출되고,
**Then** `detectActiveSpec(cwd)`가 가장 최근에 수정된 spec.md를 선택하여 `SPEC-ID: <제목>` 형식의 문자열을 `HookOutput.HookSpecificOutput.SessionTitle`에 설정해야 한다.

**Verification**: `internal/hook/user_prompt_submit.go:54-84` (Handle), `:107-158` (detectActiveSpec), `:86-102` (buildSessionTitle).

## AC-003: SPEC 미검출 시 projectName/branchName 형식으로 폴백한다

**Given** 활성 SPEC이 검출되지 않은 프로젝트에서,
**When** 프롬프트가 제출되면,
**Then** 제목은 `<projectName> / <branchName>` 형식(예: `moai-adk-go / feat/hook-enhancement`)이어야 하며, git 실패 시 `<projectName> / unknown`으로 graceful degrade 해야 한다.

**Verification**: `internal/hook/user_prompt_submit.go:180-207` — `buildProjectBranchTitle` + `getGitBranch`.

## AC-004: 제목이 비어 있으면 Claude Code는 기본 동작을 유지한다

**Given** `detectActiveSpec`과 `buildProjectBranchTitle` 모두 빈 문자열을 반환하는 경우에서,
**When** 핸들러가 응답을 생성하면,
**Then** `SessionTitle`은 omitempty 마샬링에 의해 JSON에서 생략되어 Claude Code UI가 이를 무시하고 기존 동작을 유지해야 한다.

**Verification**: `internal/hook/user_prompt_submit.go:73-75` — `if title == "" && additionalCtx == "" { return &HookOutput{}, nil }` + `types.go:277` omitempty 태그.

## AC-005: statusline.yaml이 refresh_interval 키를 10초 기본값으로 노출한다

**Given** 템플릿 `statusline.yaml`이 `moai init`으로 배포된 상태에서,
**When** 파일을 파싱하면,
**Then** `statusline.refresh_interval: 10` 키가 기본값으로 존재해야 하며 주석으로 `0 means no periodic refresh (only refresh on events)`가 명시되어야 한다.

**Verification**: `internal/template/templates/.moai/config/sections/statusline.yaml:11-13`.

## AC-006: settings.json.tmpl이 statusLine.refreshInterval 필드를 렌더링한다

**Given** `moai init` 시 settings.json 렌더링에서,
**When** `statusline` 섹션을 검사하면,
**Then** `"refreshInterval": 10` 필드가 존재해야 하며, 0이면 Claude Code가 주기 갱신을 비활성화해야 한다.

**Verification**: `internal/template/templates/.claude/settings.json.tmpl:406`.

## AC-007: WorkspaceInfo에 GitWorktree 필드, 세그먼트 상수 SegmentWorktree가 존재한다

**Given** statusline 타입 정의 모듈에서,
**When** 타입을 참조하면,
**Then** (a) `WorkspaceInfo.GitWorktree string \`json:"git_worktree"\`` 필드가 존재하고, (b) 패키지 상수 `SegmentWorktree = "worktree"`가 정의되어 있어야 한다.

**Verification**: `internal/statusline/types.go:126` (GitWorktree), `:240` (`SegmentWorktree = "worktree" // Active worktree indicator [WT]`).

## AC-008: Claude Code가 git_worktree를 제공하면 렌더러가 worktree 인디케이터를 표시한다

**Given** Claude Code 2.1.97+가 statusline JSON에 `workspace.git_worktree: "<path>"`를 제공하는 상황에서,
**When** renderer가 브랜치 라인을 렌더링하면,
**Then** `SegmentWorktree`가 활성화된 경우 브랜치 표시 앞에 worktree 인디케이터(예: `[WT]` prefix)가 표시되어야 하고, `GitWorktree`가 빈 문자열이면 표시하지 않아야 한다.

**Verification**: `internal/statusline/builder.go:205` (`data.Worktree = input.Workspace.GitWorktree`), `renderer.go:313` (`if r.isSegmentEnabled(SegmentWorktree) && data.Worktree != ""`).

## AC-009: 문서가 Claude Code 2.1.97을 worktree 최소 권장 버전으로 명시한다

**Given** `.claude/rules/moai/workflow/worktree-integration.md` 파일에서,
**When** Minimum Version Requirements 섹션을 조회하면,
**Then** 표는 `Worktree CWD isolation fix` 및 `Stop/SubagentStop hook stability`에 대해 **2.1.97**을 최소 버전으로 명시해야 한다.

**Verification**: `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/workflow/worktree-integration.md` "Minimum Version Requirements" 섹션의 `Worktree CWD isolation fix | 2.1.97 | ...` 및 `Stop/SubagentStop hook stability | 2.1.97`.

## Partial / Deferred

- **다국어 타이틀 로직 (AC-1.5 in spec.md)**: 현재 구현(`user_prompt_submit.go`)은 SPEC 제목에서 markdown heading을 그대로 사용하며 별도 `conversation_language` 조회 분기를 두지 않음. SPEC의 heading 자체가 `code_comments`/프로젝트 언어로 작성되므로 다국어가 자연히 반영되나, REQ-1의 "conversation_language 설정에 따른 명시적 분기"는 구현 생략됨. **PARTIAL: heading pass-through로 다국어 달성, ConfigProvider 기반 분기 미구현**.

## Edge Cases

- **EC-01**: `.moai/specs/` 디렉터리 없음 → `filepath.Glob` 에러 무시 후 projectName/branch 폴백.
- **EC-02**: spec.md 첫 번째 `# ` heading 없음 → SPEC-ID만 반환.
- **EC-03**: heading에 `SPEC-XXX: ` 접두사 중복 → `strings.TrimPrefix`로 제거 후 재조립.
- **EC-04**: `refreshInterval: 0` → Claude Code가 주기 갱신을 비활성화(기존 동작 유지).
- **EC-05**: `GitWorktree == ""` → 워크트리 세그먼트 생략(기존 동작 유지).

## Definition of Done

- [x] `HookSpecificOutput.SessionTitle` 필드 추가 (AC-001)
- [x] SPEC 감지 기반 제목 생성 (AC-002)
- [x] projectName/branch 폴백 (AC-003)
- [x] `refreshInterval` 설정/템플릿 렌더링 (AC-005, AC-006)
- [x] `GitWorktree` 필드 + `SegmentWorktree` 상수 + 렌더링 (AC-007, AC-008)
- [x] 문서 최소 버전 2.1.97 명시 (AC-009)
- [x] PR #610 merged to main
- [ ] **PARTIAL**: 명시적 다국어 분기(ConfigProvider language.yaml 조회) 미구현 — heading pass-through로 우회
