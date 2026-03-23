---
id: SPEC-STATUSLINE-002
version: "1.0.0"
status: completed
created: "2026-03-20"
updated: "2026-03-20"
author: GOOS
priority: high
issue_number: 0
---

# SPEC-STATUSLINE-002: rate_limits statusline 지원

## 1. 개요

Claude Code v2.1.80에서 statusline JSON에 추가된 `rate_limits` 필드를 파싱하여 5시간/7일 rate limit 사용량을 표시한다.

## 2. 요구사항

### REQ-SL-010: rate_limits 필드 파싱

StdinData에 `rate_limits` 필드를 추가하고, `RateLimitInfo` / `RateLimitWindow` 타입을 정의한다.

### REQ-SL-011: 렌더러에서 rate_limits 우선 사용

기존 MoAI API 호출(`Usage`)보다 Claude Code에서 직접 제공하는 `RateLimits`를 우선 사용한다.

## 3. 구현 요약

- `internal/statusline/types.go`: `RateLimitInfo`, `RateLimitWindow` 타입 추가, `StdinData`/`StatusData`에 필드 추가
- `internal/statusline/builder.go`: `collectAll()`에서 RateLimits 전달
- `internal/statusline/renderer.go`: `renderFullV3()`에서 RateLimits 우선 사용 로직
