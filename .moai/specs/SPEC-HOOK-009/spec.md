---
id: SPEC-HOOK-009
version: "1.0.0"
status: draft
created: "2026-03-20"
updated: "2026-03-20"
author: GOOS
priority: high
issue_number: 0
---

## HISTORY

| 버전 | 날짜 | 변경 내용 |
|------|------|----------|
| 1.0.0 | 2026-03-20 | 초안 작성 |

---

# SPEC-HOOK-009: 신규 훅 이벤트 추가 (PostCompact, InstructionsLoaded, StopFailure)

## 1. 개요

Claude Code 2.1.69~2.1.80에서 추가된 신규 훅 이벤트 3종을 MoAI-ADK-Go 훅 시스템에 등록한다.

## 2. 요구사항 (EARS 포맷)

### REQ-HOOK-040: PostCompact 이벤트 (Event-Driven)

Claude Code가 컨텍스트 컴팩션을 완료한 후 PostCompact 이벤트가 발생하면, MoAI는 세션 상태 복원과 메모리 정리를 수행해야 한다.

### REQ-HOOK-041: InstructionsLoaded 이벤트 (Event-Driven)

Claude Code가 CLAUDE.md 또는 `.claude/rules/*.md` 파일을 컨텍스트에 로딩하면, MoAI는 프로젝트 설정 유효성을 검증해야 한다.

### REQ-HOOK-042: StopFailure 이벤트 (Event-Driven)

Claude Code 턴이 API 오류(rate limit, auth 실패 등)로 종료되면, MoAI는 오류를 로깅하고 복구 가능한 조치를 안내해야 한다.

## 3. 범위

### 포함
- `internal/hook/types.go`: 3개 이벤트 상수 + ValidEventTypes 업데이트
- `internal/hook/post_compact.go`: PostCompact 핸들러 (신규)
- `internal/hook/instructions_loaded.go`: InstructionsLoaded 핸들러 (신규)
- `internal/hook/stop_failure.go`: StopFailure 핸들러 (신규)
- `internal/cli/deps.go`: InitDependencies()에 핸들러 등록
- `internal/cli/update.go`: settings.json 훅 설정 생성기에 이벤트 추가 (필요 시)
- 테스트 파일 3개

### 제외
- settings.json 템플릿 자체 수정 (hook.sh 래퍼가 이미 모든 이벤트를 라우팅)
- 기존 핸들러 수정

## 4. 기술 제약사항

- 기존 Handler 인터페이스 준수 (EventType() + Handle())
- 모든 핸들러는 non-blocking (오류 시 empty HookOutput 반환)
- go build/test/vet 통과
